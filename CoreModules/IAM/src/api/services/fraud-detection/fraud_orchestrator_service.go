/**
 * Serviço de Orquestração de Agentes de Detecção de Fraude
 * 
 * Este serviço central coordena múltiplos agentes de detecção de fraude especializados
 * por região e domínio, fornecendo uma interface unificada para o sistema de segurança.
 * 
 * Autor: Eduardo Jeremias
 * Projeto: INNOVABIZ IAM/TrustGuard
 * Data: 20/08/2025
 */

package frauddetection

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/innovabiz/iam/pkg/logging"
	"github.com/innovabiz/iam/pkg/metrics"
	"github.com/innovabiz/iam/pkg/tracing"
	"github.com/innovabiz/iam/src/api/services/trust"
	"github.com/innovabiz/iam/src/api/models"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/trace"
)

// Interface para agentes de detecção de fraude
type FraudDetectionAgent interface {
	// Identifica o agente
	GetAgentInfo() AgentInfo
	
	// Avalia uma transação ou evento em busca de sinais de fraude
	EvaluateTransaction(ctx context.Context, transaction *FraudDetectionRequest) (*FraudDetectionResult, error)
	
	// Verifica documentos de identidade
	VerifyDocument(ctx context.Context, document *DocumentVerificationRequest) (*DocumentVerificationResult, error)
	
	// Analisa comportamento de utilizador
	AnalyzeUserBehavior(ctx context.Context, behavior *UserBehaviorData) (*UserBehaviorResult, error)
	
	// Atualiza o modelo com feedback
	UpdateWithFeedback(ctx context.Context, feedback *FraudFeedback) error
	
	// Verifica se o agente está saudável
	HealthCheck(ctx context.Context) (bool, error)
}

// Informações sobre o agente
type AgentInfo struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Version          string   `json:"version"`
	RegionCodes      []string `json:"regionCodes"`      // Códigos de regiões suportadas (ex: "AO", "BR", "PT")
	SupportedDomains []string `json:"supportedDomains"` // Domínios de especialidade (ex: "transaction", "document", "behavior")
	Capabilities     []string `json:"capabilities"`     // Recursos específicos (ex: "biometrics", "device_fingerprint")
	Priority         int      `json:"priority"`         // Prioridade para resolução de conflitos (maior = mais importante)
}

// Requisição para detecção de fraude
type FraudDetectionRequest struct {
	TransactionID       string                 `json:"transactionId"`
	UserID              string                 `json:"userId"`
	TenantID            string                 `json:"tenantId"`
	RegionCode          string                 `json:"regionCode"`
	ContextID           string                 `json:"contextId"`
	TransactionType     string                 `json:"transactionType"`
	Amount              float64                `json:"amount"`
	Currency            string                 `json:"currency"`
	Timestamp           time.Time              `json:"timestamp"`
	DeviceInfo          *DeviceInfo            `json:"deviceInfo"`
	LocationInfo        *LocationInfo          `json:"locationInfo"`
	AdditionalData      map[string]interface{} `json:"additionalData"`
	TrustScore          float64                `json:"trustScore"`
	HistoricalBehavior  []UserBehaviorData     `json:"historicalBehavior,omitempty"`
	DocumentReferences  []string               `json:"documentReferences,omitempty"`
	UserAttributes      map[string]interface{} `json:"userAttributes,omitempty"`
}

// Informações sobre o dispositivo
type DeviceInfo struct {
	DeviceID           string            `json:"deviceId"`
	DeviceType         string            `json:"deviceType"`
	DeviceModel        string            `json:"deviceModel"`
	DeviceOS           string            `json:"deviceOS"`
	DeviceIPAddress    string            `json:"deviceIPAddress"`
	DeviceFingerprint  string            `json:"deviceFingerprint"`
	Browser            string            `json:"browser,omitempty"`
	UserAgent          string            `json:"userAgent,omitempty"`
	AdditionalMetadata map[string]string `json:"additionalMetadata,omitempty"`
}

// Informações de localização
type LocationInfo struct {
	Country     string  `json:"country"`
	Region      string  `json:"region,omitempty"`
	City        string  `json:"city,omitempty"`
	Coordinates *struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"coordinates,omitempty"`
	IPLocation *struct {
		Country string `json:"country"`
		Region  string `json:"region"`
		City    string `json:"city"`
	} `json:"ipLocation,omitempty"`
}

// Resultado da detecção de fraude
type FraudDetectionResult struct {
	TransactionID      string                  `json:"transactionId"`
	IsHighRisk         bool                    `json:"isHighRisk"`
	IsFraudSuspected   bool                    `json:"isFraudSuspected"`
	FraudProbability   float64                 `json:"fraudProbability"`
	RiskScore          float64                 `json:"riskScore"`
	AlertLevel         int                     `json:"alertLevel"` // 0-5 (0=baixo, 5=crítico)
	Confidence         float64                 `json:"confidence"`
	ReasonCodes        []string                `json:"reasonCodes"`
	Explanation        string                  `json:"explanation"`
	SuggestedAction    string                  `json:"suggestedAction"`
	AgentID            string                  `json:"agentId"`
	RegionCode         string                  `json:"regionCode"`
	ProcessingTime     time.Duration           `json:"processingTime"`
	DetectionPatterns  []string                `json:"detectionPatterns,omitempty"`
	AdditionalFindings map[string]interface{}  `json:"additionalFindings,omitempty"`
	RecommendedActions []RecommendedAction     `json:"recommendedActions,omitempty"`
	Timestamp          time.Time               `json:"timestamp"`
}

// Ação recomendada
type RecommendedAction struct {
	ActionType     string                 `json:"actionType"` // ex: "block", "additional_verification", "notify"
	Severity       int                    `json:"severity"`   // 1-5
	Message        string                 `json:"message"`
	SecurityAdjust *SecurityAdjustment    `json:"securityAdjust,omitempty"`
	Parameters     map[string]interface{} `json:"parameters,omitempty"`
}

// Ajuste de segurança recomendado
type SecurityAdjustment struct {
	Dimension     string  `json:"dimension"`
	Direction     string  `json:"direction"`
	Magnitude     int     `json:"magnitude"`
	DurationHours int     `json:"durationHours"`
	Reason        string  `json:"reason"`
}

// Dados para verificação de documento
type DocumentVerificationRequest struct {
	DocumentID     string                 `json:"documentId"`
	DocumentType   string                 `json:"documentType"` // ex: "passport", "id_card", "drivers_license"
	UserID         string                 `json:"userId"`
	TenantID       string                 `json:"tenantId"`
	RegionCode     string                 `json:"regionCode"`
	IssuingCountry string                 `json:"issuingCountry"`
	DocumentData   map[string]interface{} `json:"documentData"`
	DocumentImages []DocumentImage        `json:"documentImages,omitempty"`
}

// Imagem de documento
type DocumentImage struct {
	Type      string `json:"type"` // ex: "front", "back", "selfie"
	ImageURI  string `json:"imageUri"`
	Format    string `json:"format"`
	Timestamp time.Time `json:"timestamp"`
}

// Resultado da verificação de documento
type DocumentVerificationResult struct {
	DocumentID        string                 `json:"documentId"`
	IsVerified        bool                   `json:"isVerified"`
	Confidence        float64                `json:"confidence"`
	RiskLevel         int                    `json:"riskLevel"` // 0-5 (0=baixo, 5=crítico)
	VerificationFlags map[string]bool        `json:"verificationFlags"`
	Findings          map[string]interface{} `json:"findings,omitempty"`
	ReasonCodes       []string               `json:"reasonCodes"`
	AgentID           string                 `json:"agentId"`
	ProcessingTime    time.Duration          `json:"processingTime"`
	RecommendedActions []RecommendedAction    `json:"recommendedActions,omitempty"`
	Timestamp         time.Time              `json:"timestamp"`
}

// Dados de comportamento do utilizador
type UserBehaviorData struct {
	UserID             string                 `json:"userId"`
	TenantID           string                 `json:"tenantId"`
	SessionID          string                 `json:"sessionId,omitempty"`
	BehaviorType       string                 `json:"behaviorType"` // ex: "login", "browsing", "transaction"
	ActivityTimestamp  time.Time              `json:"activityTimestamp"`
	DeviceInfo         *DeviceInfo            `json:"deviceInfo"`
	LocationInfo       *LocationInfo          `json:"locationInfo"`
	BehaviorMetrics    map[string]float64     `json:"behaviorMetrics,omitempty"`
	ActivitySequence   []UserActivity         `json:"activitySequence,omitempty"`
	AdditionalContext  map[string]interface{} `json:"additionalContext,omitempty"`
}

// Atividade do utilizador
type UserActivity struct {
	ActivityType string    `json:"activityType"`
	Timestamp    time.Time `json:"timestamp"`
	Duration     int       `json:"duration,omitempty"` // em segundos
	Details      map[string]interface{} `json:"details,omitempty"`
}

// Resultado da análise de comportamento
type UserBehaviorResult struct {
	UserID           string                 `json:"userId"`
	AnomalyDetected  bool                   `json:"anomalyDetected"`
	AnomalyScore     float64                `json:"anomalyScore"` // 0-1
	Confidence       float64                `json:"confidence"`
	BehaviorPatterns []string               `json:"behaviorPatterns"`
	RiskLevel        int                    `json:"riskLevel"` // 0-5
	Explanation      string                 `json:"explanation"`
	AgentID          string                 `json:"agentId"`
	RegionRelevance  map[string]float64     `json:"regionRelevance,omitempty"` // relevância por região
	AdditionalInsights map[string]interface{} `json:"additionalInsights,omitempty"`
	RecommendedActions []RecommendedAction   `json:"recommendedActions,omitempty"`
	Timestamp        time.Time              `json:"timestamp"`
}

// Feedback sobre detecção de fraude
type FraudFeedback struct {
	TransactionID     string    `json:"transactionId"`
	DocumentID        string    `json:"documentId,omitempty"`
	UserID            string    `json:"userId,omitempty"`
	ActuallyFraud     bool      `json:"actuallyFraud"`
	FeedbackSource    string    `json:"feedbackSource"` // ex: "manual_review", "chargeback", "user_report"
	FeedbackNotes     string    `json:"feedbackNotes,omitempty"`
	AdditionalData    map[string]interface{} `json:"additionalData,omitempty"`
	FeedbackTimestamp time.Time `json:"feedbackTimestamp"`
}

// Configuração do orquestrador
type FraudOrchestratorConfig struct {
	// Configurações gerais
	EnabledRegions          []string `json:"enabledRegions"`
	DefaultTimeoutSeconds   int      `json:"defaultTimeoutSeconds"`
	EnableParallelProcessing bool    `json:"enableParallelProcessing"`
	MaxConcurrentRequests    int     `json:"maxConcurrentRequests"`
	
	// Kafka
	KafkaBootstrapServers   []string `json:"kafkaBootstrapServers"`
	KafkaConsumerGroup      string   `json:"kafkaConsumerGroup"`
	KafkaTransactionTopic   string   `json:"kafkaTransactionTopic"`
	KafkaDocumentTopic      string   `json:"kafkaDocumentTopic"`
	KafkaBehaviorTopic      string   `json:"kafkaBehaviorTopic"`
	KafkaFeedbackTopic      string   `json:"kafkaFeedbackTopic"`
	KafkaAlertTopic         string   `json:"kafkaAlertTopic"`
	
	// TrustGuard
	TrustScoreThreshold     float64  `json:"trustScoreThreshold"`
	EnableTrustScoreAdjust  bool     `json:"enableTrustScoreAdjust"`
	
	// Agentes
	AgentHeartbeatIntervalSeconds int `json:"agentHeartbeatIntervalSeconds"`
	AgentRegistryRefreshInterval  int `json:"agentRegistryRefreshInterval"`
	
	// Cache
	EnableResultCaching    bool `json:"enableResultCaching"`
	CacheExpirationMinutes int  `json:"cacheExpirationMinutes"`
}

// FraudOrchestratorService coordena os agentes de detecção de fraude
type FraudOrchestratorService struct {
	config        FraudOrchestratorConfig
	agents        map[string]FraudDetectionAgent // mapa de ID para agente
	agentsByRegion map[string][]FraudDetectionAgent // mapa de código de região para lista de agentes
	mu            sync.RWMutex
	trustService  *trust.TrustScoreService
	logger        *logging.Logger
	tracer        trace.Tracer
	metrics       *metrics.MetricsCollector
	kafkaWriter   *kafka.Writer
	kafkaReaders  map[string]*kafka.Reader
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewFraudOrchestratorService cria uma nova instância do serviço de orquestração
func NewFraudOrchestratorService(
	config FraudOrchestratorConfig,
	trustService *trust.TrustScoreService,
	logger *logging.Logger,
) (*FraudOrchestratorService, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Configurar tracer para OpenTelemetry
	t := tracing.GetTracer("fraud-orchestrator-service")
	
	// Configurar coletor de métricas
	m := metrics.NewMetricsCollector("fraud_orchestrator")
	
	// Configurar Kafka writer para alertas
	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP(config.KafkaBootstrapServers...),
		Topic:    config.KafkaAlertTopic,
		Balancer: &kafka.LeastBytes{},
	}
	
	// Inicializar mapas de agentes
	agentsMap := make(map[string]FraudDetectionAgent)
	agentsByRegion := make(map[string][]FraudDetectionAgent)
	
	// Inicializar leitores Kafka
	kafkaReaders := make(map[string]*kafka.Reader)
	
	service := &FraudOrchestratorService{
		config:        config,
		agents:        agentsMap,
		agentsByRegion: agentsByRegion,
		trustService:  trustService,
		logger:        logger,
		tracer:        t,
		metrics:       m,
		kafkaWriter:   kafkaWriter,
		kafkaReaders:  kafkaReaders,
		ctx:           ctx,
		cancel:        cancel,
	}
	
	return service, nil
}

// Start inicia o serviço de orquestração
func (s *FraudOrchestratorService) Start() error {
	ctx, span := s.tracer.Start(s.ctx, "FraudOrchestratorService.Start")
	defer span.End()
	
	s.logger.Info(ctx, "Iniciando serviço de orquestração de detecção de fraudes")
	
	// Iniciar o registro de agentes
	if err := s.initializeAgentRegistry(ctx); err != nil {
		s.logger.Error(ctx, "Erro ao inicializar registro de agentes: %v", err)
		return fmt.Errorf("falha ao inicializar registro de agentes: %w", err)
	}
	
	// Configurar consumidores Kafka
	if err := s.setupKafkaConsumers(ctx); err != nil {
		s.logger.Error(ctx, "Erro ao configurar consumidores Kafka: %v", err)
		return fmt.Errorf("falha ao configurar consumidores Kafka: %w", err)
	}
	
	// Iniciar verificações de saúde periódicas dos agentes
	go s.startAgentHealthChecks(ctx)
	
	s.logger.Info(ctx, "Serviço de orquestração de detecção de fraudes iniciado com sucesso")
	return nil
}

// Stop interrompe o serviço de orquestração
func (s *FraudOrchestratorService) Stop() error {
	s.logger.Info(s.ctx, "Parando serviço de orquestração de detecção de fraudes")
	
	// Cancelar contexto principal
	s.cancel()
	
	// Fechar writer do Kafka
	if err := s.kafkaWriter.Close(); err != nil {
		s.logger.Error(s.ctx, "Erro ao fechar Kafka writer: %v", err)
	}
	
	// Fechar readers do Kafka
	for topic, reader := range s.kafkaReaders {
		if err := reader.Close(); err != nil {
			s.logger.Error(s.ctx, "Erro ao fechar Kafka reader para tópico %s: %v", topic, err)
		}
	}
	
	s.logger.Info(s.ctx, "Serviço de orquestração de detecção de fraudes parado com sucesso")
	return nil
}

// RegisterAgent registra um novo agente no orquestrador
func (s *FraudOrchestratorService) RegisterAgent(ctx context.Context, agent FraudDetectionAgent) error {
	ctx, span := s.tracer.Start(ctx, "FraudOrchestratorService.RegisterAgent")
	defer span.End()
	
	info := agent.GetAgentInfo()
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Verificar se o agente já está registrado
	if _, exists := s.agents[info.ID]; exists {
		s.logger.Warn(ctx, "Agente com ID %s já registrado, substituindo", info.ID)
	}
	
	// Registrar o agente
	s.agents[info.ID] = agent
	
	// Registrar por região
	for _, regionCode := range info.RegionCodes {
		if _, exists := s.agentsByRegion[regionCode]; !exists {
			s.agentsByRegion[regionCode] = []FraudDetectionAgent{}
		}
		s.agentsByRegion[regionCode] = append(s.agentsByRegion[regionCode], agent)
	}
	
	s.logger.Info(ctx, "Agente registrado com sucesso: %s (regiões: %v)", info.Name, info.RegionCodes)
	
	// Incrementar métrica de agentes registrados
	s.metrics.IncrementCounter("agents_registered_total", 1, map[string]string{
		"agent_id":   info.ID,
		"agent_name": info.Name,
	})
	
	return nil
}

// UnregisterAgent remove um agente do orquestrador
func (s *FraudOrchestratorService) UnregisterAgent(ctx context.Context, agentID string) error {
	ctx, span := s.tracer.Start(ctx, "FraudOrchestratorService.UnregisterAgent")
	defer span.End()
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Verificar se o agente existe
	agent, exists := s.agents[agentID]
	if !exists {
		return fmt.Errorf("agente com ID %s não encontrado", agentID)
	}
	
	info := agent.GetAgentInfo()
	
	// Remover o agente do mapa principal
	delete(s.agents, agentID)
	
	// Remover o agente dos mapas de região
	for _, regionCode := range info.RegionCodes {
		if agents, exists := s.agentsByRegion[regionCode]; exists {
			updatedAgents := []FraudDetectionAgent{}
			for _, a := range agents {
				if a.GetAgentInfo().ID != agentID {
					updatedAgents = append(updatedAgents, a)
				}
			}
			s.agentsByRegion[regionCode] = updatedAgents
		}
	}
	
	s.logger.Info(ctx, "Agente removido com sucesso: %s", agentID)
	return nil
}

// ProcessTransactionFraudDetection processa uma requisição de detecção de fraude de transação
func (s *FraudOrchestratorService) ProcessTransactionFraudDetection(
	ctx context.Context,
	request *FraudDetectionRequest,
) (*FraudDetectionResult, error) {
	ctx, span := s.tracer.Start(ctx, "FraudOrchestratorService.ProcessTransactionFraudDetection")
	defer span.End()
	
	startTime := time.Now()
	
	// Registrar métricas de entrada
	s.metrics.IncrementCounter("fraud_detection_requests_total", 1, map[string]string{
		"transaction_type": request.TransactionType,
		"region_code":      request.RegionCode,
	})
	
	// Obter agentes relevantes para a região
	relevantAgents := s.getRelevantAgentsForRegion(ctx, request.RegionCode)
	if len(relevantAgents) == 0 {
		s.logger.Warn(ctx, "Nenhum agente disponível para a região %s", request.RegionCode)
		return &FraudDetectionResult{
			TransactionID:    request.TransactionID,
			IsHighRisk:       false,
			IsFraudSuspected: false,
			FraudProbability: 0,
			RiskScore:        0,
			AlertLevel:       0,
			Confidence:       0,
			ReasonCodes:      []string{"NO_AGENT_AVAILABLE"},
			Explanation:      "Nenhum agente disponível para processar esta transação",
			SuggestedAction:  "manual_review",
			RegionCode:       request.RegionCode,
			ProcessingTime:   time.Since(startTime),
			Timestamp:        time.Now(),
		}, nil
	}
	
	// Processamento em paralelo ou sequencial dependendo da configuração
	var results []*FraudDetectionResult
	if s.config.EnableParallelProcessing {
		results = s.processTransactionInParallel(ctx, request, relevantAgents)
	} else {
		results = s.processTransactionSequential(ctx, request, relevantAgents)
	}
	
	// Combinar resultados usando estratégia de consenso
	finalResult := s.combineDetectionResults(ctx, results, request)
	finalResult.ProcessingTime = time.Since(startTime)
	
	// Atualizar métricas
	s.updateFraudDetectionMetrics(finalResult)
	
	// Se for detectada fraude de alto risco, publicar alerta
	if finalResult.IsHighRisk && finalResult.IsFraudSuspected {
		go s.publishFraudAlert(ctx, finalResult)
	}
	
	return finalResult, nil
}

// Auxiliares e métodos internos restantes serão implementados posteriormente