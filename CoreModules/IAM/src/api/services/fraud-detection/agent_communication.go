/**
 * Mecanismo de Comunicação entre Agentes e Orquestrador
 * 
 * Implementa os protocolos de comunicação e troca de mensagens entre o
 * orquestrador central e os agentes IA distribuídos de detecção de fraude.
 * 
 * Autor: Eduardo Jeremias
 * Projeto: INNOVABIZ IAM/TrustGuard
 * Data: 20/08/2025
 */

package frauddetection

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/innovabiz/iam/pkg/logging"
	"github.com/innovabiz/iam/pkg/tracing"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// Tipos de mensagens para comunicação entre agentes e orquestrador
const (
	MessageTypeHeartbeat         = "heartbeat"
	MessageTypeRegistration      = "registration"
	MessageTypeTransactionEval   = "transaction_eval"
	MessageTypeDocumentVerify    = "document_verify"
	MessageTypeBehaviorAnalysis  = "behavior_analysis"
	MessageTypeFeedback          = "feedback"
	MessageTypeModelUpdate       = "model_update"
	MessageTypeAlert             = "alert"
	MessageTypeConfigUpdate      = "config_update"
)

// Canais de comunicação disponíveis
const (
	ChannelTypeGRPC  = "grpc"
	ChannelTypeKafka = "kafka"
	ChannelTypeHTTP  = "http"
)

// AgentMessage representa uma mensagem entre agente e orquestrador
type AgentMessage struct {
	MessageID      string                 `json:"messageId"`
	MessageType    string                 `json:"messageType"`
	SenderID       string                 `json:"senderId"`
	RecipientID    string                 `json:"recipientId,omitempty"` // Vazio para broadcast
	CorrelationID  string                 `json:"correlationId,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
	Priority       int                    `json:"priority"`
	TTLSeconds     int                    `json:"ttlSeconds"`
	Channel        string                 `json:"channel"`
	Payload        map[string]interface{} `json:"payload"`
	RegionCodes    []string               `json:"regionCodes,omitempty"`
	RequiresAck    bool                   `json:"requiresAck"`
	DeliveryStatus string                 `json:"deliveryStatus,omitempty"`
	ErrorCode      int                    `json:"errorCode,omitempty"`
	ErrorMessage   string                 `json:"errorMessage,omitempty"`
}

// Configuração de comunicação com agentes
type AgentCommunicationConfig struct {
	// Configuração GRPC
	GRPCEnabled           bool     `json:"grpcEnabled"`
	GRPCPort              int      `json:"grpcPort"`
	GRPCMaxMessageSizeMB  int      `json:"grpcMaxMessageSizeMB"`
	GRPCKeepAliveSeconds  int      `json:"grpcKeepAliveSeconds"`
	
	// Configuração Kafka
	KafkaEnabled          bool     `json:"kafkaEnabled"`
	KafkaBootstrapServers []string `json:"kafkaBootstrapServers"`
	KafkaTopic            string   `json:"kafkaTopic"`
	KafkaConsumerGroup    string   `json:"kafkaConsumerGroup"`
	
	// Configuração HTTP
	HTTPEnabled           bool     `json:"httpEnabled"`
	HTTPPort              int      `json:"httpPort"`
	HTTPBasePath          string   `json:"httpBasePath"`
	
	// Configuração geral
	MessageTimeoutSeconds int      `json:"messageTimeoutSeconds"`
	HeartbeatIntervalSec  int      `json:"heartbeatIntervalSec"`
	RetryMaxAttempts      int      `json:"retryMaxAttempts"`
	RetryIntervalMs       int      `json:"retryIntervalMs"`
}

// Interface para manipuladores de mensagens
type MessageHandler interface {
	HandleMessage(ctx context.Context, message *AgentMessage) (*AgentMessage, error)
}

// Interface para canais de comunicação
type CommunicationChannel interface {
	Initialize(ctx context.Context) error
	SendMessage(ctx context.Context, message *AgentMessage) error
	ReceiveMessages(ctx context.Context) (<-chan *AgentMessage, error)
	Close(ctx context.Context) error
}

// AgentCommunicator gerencia a comunicação entre agentes e orquestrador
type AgentCommunicator struct {
	config          AgentCommunicationConfig
	logger          *logging.Logger
	tracer          trace.Tracer
	messageHandlers map[string]MessageHandler
	channels        map[string]CommunicationChannel
	pendingMessages sync.Map // map[string]*AgentMessage - messageID para mensagem pendente
	waitingReplies  sync.Map // map[string]chan *AgentMessage - correlationID para canal de resposta
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
}

// Implementação do canal GRPC
type GRPCChannel struct {
	server     *grpc.Server
	clients    map[string]*grpc.ClientConn
	port       int
	logger     *logging.Logger
	tracer     trace.Tracer
	messageCh  chan *AgentMessage
	clientsMu  sync.RWMutex
}

// Implementação do canal Kafka
type KafkaChannel struct {
	writer       *kafka.Writer
	reader       *kafka.Reader
	topic        string
	bootstrapServers []string
	consumerGroup string
	logger       *logging.Logger
	tracer       trace.Tracer
	messageCh    chan *AgentMessage
}

// NewAgentCommunicator cria uma nova instância do comunicador de agentes
func NewAgentCommunicator(
	config AgentCommunicationConfig,
	logger *logging.Logger,
) (*AgentCommunicator, error) {
	ctx, cancel := context.WithCancel(context.Background())
	tracer := tracing.GetTracer("agent-communicator")
	
	communicator := &AgentCommunicator{
		config:          config,
		logger:          logger,
		tracer:          tracer,
		messageHandlers: make(map[string]MessageHandler),
		channels:        make(map[string]CommunicationChannel),
		ctx:             ctx,
		cancel:          cancel,
	}
	
	return communicator, nil
}

// Initialize inicializa o comunicador de agentes
func (c *AgentCommunicator) Initialize(ctx context.Context) error {
	ctx, span := c.tracer.Start(ctx, "AgentCommunicator.Initialize")
	defer span.End()
	
	c.logger.Info(ctx, "Inicializando comunicador de agentes")
	
	// Inicializar canais de comunicação
	if c.config.GRPCEnabled {
		grpcChannel := &GRPCChannel{
			port:      c.config.GRPCPort,
			logger:    c.logger,
			tracer:    c.tracer,
			messageCh: make(chan *AgentMessage, 100),
			clients:   make(map[string]*grpc.ClientConn),
		}
		
		if err := grpcChannel.Initialize(ctx); err != nil {
			return fmt.Errorf("falha ao inicializar canal GRPC: %w", err)
		}
		
		c.channels[ChannelTypeGRPC] = grpcChannel
		c.logger.Info(ctx, "Canal GRPC inicializado na porta %d", c.config.GRPCPort)
	}
	
	if c.config.KafkaEnabled {
		kafkaChannel := &KafkaChannel{
			topic:            c.config.KafkaTopic,
			bootstrapServers: c.config.KafkaBootstrapServers,
			consumerGroup:    c.config.KafkaConsumerGroup,
			logger:           c.logger,
			tracer:           c.tracer,
			messageCh:        make(chan *AgentMessage, 100),
		}
		
		if err := kafkaChannel.Initialize(ctx); err != nil {
			return fmt.Errorf("falha ao inicializar canal Kafka: %w", err)
		}
		
		c.channels[ChannelTypeKafka] = kafkaChannel
		c.logger.Info(ctx, "Canal Kafka inicializado com tópico %s", c.config.KafkaTopic)
	}
	
	// Iniciar goroutines para processar mensagens de cada canal
	for channelType, channel := range c.channels {
		go c.processChannelMessages(ctx, channelType, channel)
	}
	
	// Iniciar goroutine para limpeza de mensagens expiradas
	go c.startExpirationCleaner(ctx)
	
	c.logger.Info(ctx, "Comunicador de agentes inicializado com sucesso")
	return nil
}

// RegisterMessageHandler registra um manipulador para um tipo de mensagem
func (c *AgentCommunicator) RegisterMessageHandler(messageType string, handler MessageHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.messageHandlers[messageType] = handler
	c.logger.Info(c.ctx, "Manipulador registrado para mensagens do tipo %s", messageType)
}

// SendMessage envia uma mensagem através do canal especificado
func (c *AgentCommunicator) SendMessage(
	ctx context.Context, 
	message *AgentMessage,
) error {
	ctx, span := c.tracer.Start(
		ctx, 
		"AgentCommunicator.SendMessage",
		trace.WithAttributes(
			attribute.String("message.id", message.MessageID),
			attribute.String("message.type", message.MessageType),
		),
	)
	defer span.End()
	
	// Validar a mensagem
	if message.MessageID == "" || message.MessageType == "" || message.SenderID == "" {
		return errors.New("mensagem inválida: MessageID, MessageType e SenderID são obrigatórios")
	}
	
	// Definir timestamp se não estiver definido
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}
	
	// Se o canal não for especificado, usar o padrão disponível
	if message.Channel == "" {
		if c.config.GRPCEnabled {
			message.Channel = ChannelTypeGRPC
		} else if c.config.KafkaEnabled {
			message.Channel = ChannelTypeKafka
		} else if c.config.HTTPEnabled {
			message.Channel = ChannelTypeHTTP
		} else {
			return errors.New("nenhum canal de comunicação disponível")
		}
	}
	
	// Verificar se o canal especificado está disponível
	channel, ok := c.channels[message.Channel]
	if !ok {
		return fmt.Errorf("canal %s não está disponível", message.Channel)
	}
	
	// Se a mensagem requer confirmação, registrá-la como pendente
	if message.RequiresAck {
		c.pendingMessages.Store(message.MessageID, message)
	}
	
	// Enviar a mensagem
	err := channel.SendMessage(ctx, message)
	if err != nil {
		return fmt.Errorf("falha ao enviar mensagem: %w", err)
	}
	
	c.logger.Debug(ctx, "Mensagem %s enviada com sucesso pelo canal %s", message.MessageID, message.Channel)
	return nil
}

// SendAndWaitReply envia uma mensagem e aguarda por uma resposta
func (c *AgentCommunicator) SendAndWaitReply(
	ctx context.Context, 
	message *AgentMessage,
	timeoutSeconds int,
) (*AgentMessage, error) {
	ctx, span := c.tracer.Start(ctx, "AgentCommunicator.SendAndWaitReply")
	defer span.End()
	
	// Garantir que a mensagem tem um correlationID
	if message.CorrelationID == "" {
		message.CorrelationID = message.MessageID
	}
	
	// Configurar para receber resposta
	message.RequiresAck = true
	
	// Criar canal para esperar a resposta
	replyChan := make(chan *AgentMessage, 1)
	c.waitingReplies.Store(message.CorrelationID, replyChan)
	
	// Garantir que removemos o canal quando sairmos
	defer c.waitingReplies.Delete(message.CorrelationID)
	
	// Enviar a mensagem
	if err := c.SendMessage(ctx, message); err != nil {
		return nil, err
	}
	
	// Configurar timeout
	timeout := time.Duration(timeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = time.Duration(c.config.MessageTimeoutSeconds) * time.Second
	}
	
	// Esperar pela resposta ou timeout
	select {
	case reply := <-replyChan:
		return reply, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout aguardando resposta para mensagem %s", message.MessageID)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Close fecha o comunicador e seus canais
func (c *AgentCommunicator) Close() error {
	c.logger.Info(c.ctx, "Fechando comunicador de agentes")
	
	// Cancelar o contexto para interromper goroutines
	c.cancel()
	
	// Fechar todos os canais
	for channelType, channel := range c.channels {
		if err := channel.Close(c.ctx); err != nil {
			c.logger.Error(c.ctx, "Erro ao fechar canal %s: %v", channelType, err)
		}
	}
	
	c.logger.Info(c.ctx, "Comunicador de agentes fechado com sucesso")
	return nil
}

// processChannelMessages processa mensagens recebidas de um canal
func (c *AgentCommunicator) processChannelMessages(
	ctx context.Context,
	channelType string,
	channel CommunicationChannel,
) {
	ctx, span := c.tracer.Start(ctx, "AgentCommunicator.processChannelMessages")
	defer span.End()
	
	c.logger.Info(ctx, "Iniciando processamento de mensagens do canal %s", channelType)
	
	// Obter canal de mensagens
	messageCh, err := channel.ReceiveMessages(ctx)
	if err != nil {
		c.logger.Error(ctx, "Erro ao receber mensagens do canal %s: %v", channelType, err)
		return
	}
	
	for {
		select {
		case <-ctx.Done():
			c.logger.Info(ctx, "Processamento de mensagens do canal %s interrompido", channelType)
			return
		case message, ok := <-messageCh:
			if !ok {
				c.logger.Warn(ctx, "Canal de mensagens %s fechado", channelType)
				return
			}
			
			// Processar a mensagem
			go c.handleIncomingMessage(ctx, message)
		}
	}
}

// handleIncomingMessage processa uma mensagem recebida
func (c *AgentCommunicator) handleIncomingMessage(ctx context.Context, message *AgentMessage) {
	ctx, span := c.tracer.Start(
		ctx, 
		"AgentCommunicator.handleIncomingMessage",
		trace.WithAttributes(
			attribute.String("message.id", message.MessageID),
			attribute.String("message.type", message.MessageType),
			attribute.String("message.sender", message.SenderID),
		),
	)
	defer span.End()
	
	// Se for uma resposta a uma mensagem anterior
	if message.CorrelationID != "" && message.CorrelationID != message.MessageID {
		// Verificar se há alguém esperando por essa resposta
		if ch, ok := c.waitingReplies.Load(message.CorrelationID); ok {
			replyChan := ch.(chan *AgentMessage)
			
			// Enviar resposta ao canal
			select {
			case replyChan <- message:
				c.logger.Debug(ctx, "Resposta enviada para canal de correlação %s", message.CorrelationID)
			default:
				c.logger.Warn(ctx, "Canal de resposta para %s está cheio ou fechado", message.CorrelationID)
			}
			return
		}
	}
	
	// Verificar se há um manipulador para este tipo de mensagem
	c.mu.RLock()
	handler, exists := c.messageHandlers[message.MessageType]
	c.mu.RUnlock()
	
	if !exists {
		c.logger.Warn(ctx, "Nenhum manipulador registrado para mensagem do tipo %s", message.MessageType)
		return
	}
	
	// Processar mensagem com o manipulador registrado
	reply, err := handler.HandleMessage(ctx, message)
	if err != nil {
		c.logger.Error(ctx, "Erro ao processar mensagem %s: %v", message.MessageID, err)
		
		// Se a mensagem requer confirmação, enviar erro
		if message.RequiresAck {
			errorReply := &AgentMessage{
				MessageID:      fmt.Sprintf("error-%s", message.MessageID),
				MessageType:    fmt.Sprintf("error-%s", message.MessageType),
				SenderID:       message.RecipientID,
				RecipientID:    message.SenderID,
				CorrelationID:  message.MessageID,
				Timestamp:      time.Now(),
				Channel:        message.Channel,
				ErrorCode:      500,
				ErrorMessage:   err.Error(),
			}
			
			if err := c.SendMessage(ctx, errorReply); err != nil {
				c.logger.Error(ctx, "Erro ao enviar mensagem de erro: %v", err)
			}
		}
		return
	}
	
	// Se houver uma resposta e a mensagem original requer confirmação, enviar a resposta
	if reply != nil && message.RequiresAck {
		// Configurar metadados de resposta
		if reply.MessageID == "" {
			reply.MessageID = fmt.Sprintf("reply-%s", message.MessageID)
		}
		reply.CorrelationID = message.MessageID
		reply.RecipientID = message.SenderID
		reply.Channel = message.Channel
		
		if err := c.SendMessage(ctx, reply); err != nil {
			c.logger.Error(ctx, "Erro ao enviar resposta: %v", err)
		}
	}
}

// startExpirationCleaner inicia um limpador para remover mensagens expiradas
func (c *AgentCommunicator) startExpirationCleaner(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.cleanExpiredMessages()
		}
	}
}

// cleanExpiredMessages remove mensagens pendentes expiradas
func (c *AgentCommunicator) cleanExpiredMessages() {
	now := time.Now()
	
	// Verificar mensagens pendentes
	c.pendingMessages.Range(func(key, value interface{}) bool {
		message := value.(*AgentMessage)
		
		// Calcular tempo de expiração
		expireTime := message.Timestamp.Add(time.Duration(message.TTLSeconds) * time.Second)
		
		// Se o tempo expirou, remover a mensagem
		if now.After(expireTime) {
			c.pendingMessages.Delete(key)
			c.logger.Debug(c.ctx, "Mensagem pendente expirada removida: %s", message.MessageID)
		}
		
		return true
	})
}

// GetHeartbeatMessage cria uma mensagem de heartbeat
func (c *AgentCommunicator) GetHeartbeatMessage(senderID string) *AgentMessage {
	return &AgentMessage{
		MessageID:   fmt.Sprintf("heartbeat-%s-%d", senderID, time.Now().UnixNano()),
		MessageType: MessageTypeHeartbeat,
		SenderID:    senderID,
		Timestamp:   time.Now(),
		Priority:    1,
		TTLSeconds:  60,
		Channel:     ChannelTypeGRPC, // Preferencial para heartbeats
		Payload: map[string]interface{}{
			"status": "active",
			"uptime": time.Now().Unix(),
		},
		RequiresAck: true,
	}
}

// GetRegistrationMessage cria uma mensagem de registro de agente
func (c *AgentCommunicator) GetRegistrationMessage(agentInfo AgentInfo) *AgentMessage {
	payload, _ := json.Marshal(agentInfo)
	payloadMap := make(map[string]interface{})
	json.Unmarshal(payload, &payloadMap)
	
	return &AgentMessage{
		MessageID:   fmt.Sprintf("registration-%s-%d", agentInfo.ID, time.Now().UnixNano()),
		MessageType: MessageTypeRegistration,
		SenderID:    agentInfo.ID,
		Timestamp:   time.Now(),
		Priority:    2,
		TTLSeconds:  300,
		Channel:     ChannelTypeGRPC, // Preferencial para registros
		Payload:     payloadMap,
		RegionCodes: agentInfo.RegionCodes,
		RequiresAck: true,
	}
}