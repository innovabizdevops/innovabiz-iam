/**
 * Mecanismo de Priorização e Escalonamento de Alertas de Fraude
 * 
 * Este componente gerencia a priorização, o escalonamento e o encaminhamento
 * de alertas de fraude identificados pelos agentes IA, com suporte a múltiplos
 * níveis de severidade e regras de escalonamento específicas por região.
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
	"sort"
	"sync"
	"time"

	"github.com/innovabiz/iam/pkg/logging"
	"github.com/innovabiz/iam/pkg/metrics"
	"github.com/innovabiz/iam/pkg/notification"
	"github.com/innovabiz/iam/pkg/tracing"
	"github.com/innovabiz/iam/src/api/models"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// FraudAlertLevel define os níveis de alerta de fraude
type FraudAlertLevel int

const (
	AlertLevelInfo     FraudAlertLevel = 1
	AlertLevelLow      FraudAlertLevel = 2
	AlertLevelMedium   FraudAlertLevel = 3
	AlertLevelHigh     FraudAlertLevel = 4
	AlertLevelCritical FraudAlertLevel = 5
)

// FraudAlertStatus define o status de um alerta de fraude
type FraudAlertStatus string

const (
	AlertStatusNew        FraudAlertStatus = "new"
	AlertStatusAssigned   FraudAlertStatus = "assigned"
	AlertStatusInProgress FraudAlertStatus = "in_progress"
	AlertStatusEscalated  FraudAlertStatus = "escalated"
	AlertStatusResolved   FraudAlertStatus = "resolved"
	AlertStatusClosed     FraudAlertStatus = "closed"
	AlertStatusFalseAlarm FraudAlertStatus = "false_alarm"
)

// FraudAlertPriority define a prioridade de um alerta de fraude
type FraudAlertPriority int

const (
	PriorityLow       FraudAlertPriority = 1
	PriorityNormal    FraudAlertPriority = 2
	PriorityHigh      FraudAlertPriority = 3
	PriorityUrgent    FraudAlertPriority = 4
	PriorityCritical  FraudAlertPriority = 5
)

// FraudAlertAction define as ações possíveis para um alerta
type FraudAlertAction string

const (
	ActionMonitor          FraudAlertAction = "monitor"
	ActionNotify           FraudAlertAction = "notify"
	ActionBlock            FraudAlertAction = "block"
	ActionVerify           FraudAlertAction = "verify"
	ActionEscalate         FraudAlertAction = "escalate"
	ActionSecurityAdjust   FraudAlertAction = "security_adjust"
	ActionManualReview     FraudAlertAction = "manual_review"
	ActionReport           FraudAlertAction = "report"
)

// FraudAlert representa um alerta de fraude no sistema
type FraudAlert struct {
	AlertID           string                 `json:"alertId"`
	SourceAgentID     string                 `json:"sourceAgentId"`
	TransactionID     string                 `json:"transactionId,omitempty"`
	DocumentID        string                 `json:"documentId,omitempty"`
	UserID            string                 `json:"userId"`
	TenantID          string                 `json:"tenantId"`
	RegionCode        string                 `json:"regionCode"`
	ContextID         string                 `json:"contextId"`
	Category          string                 `json:"category"` // "transaction", "document", "behavior", "pattern"
	AlertLevel        FraudAlertLevel        `json:"alertLevel"`
	Priority          FraudAlertPriority     `json:"priority"`
	Status            FraudAlertStatus       `json:"status"`
	Title             string                 `json:"title"`
	Description       string                 `json:"description"`
	DetectionPatterns []string               `json:"detectionPatterns,omitempty"`
	RiskScore         float64                `json:"riskScore"`
	Confidence        float64                `json:"confidence"`
	ImpactLevel       int                    `json:"impactLevel"` // 1-5
	RequiredActions   []FraudAlertAction     `json:"requiredActions"`
	SecurityAdjusts   []*SecurityAdjustment  `json:"securityAdjusts,omitempty"`
	TrustScoreImpact  float64                `json:"trustScoreImpact"`
	AssignedToID      string                 `json:"assignedToId,omitempty"`
	AssignedToTeam    string                 `json:"assignedToTeam,omitempty"`
	EscalationLevel   int                    `json:"escalationLevel"`
	EscalationHistory []*EscalationEvent     `json:"escalationHistory,omitempty"`
	RelatedAlerts     []string               `json:"relatedAlerts,omitempty"`
	Evidence          map[string]interface{} `json:"evidence,omitempty"`
	TimeToLiveMinutes int                    `json:"timeToLiveMinutes"` // tempo até auto-resolução
	CreatedAt         time.Time              `json:"createdAt"`
	UpdatedAt         time.Time              `json:"updatedAt"`
	EscalateAfter     time.Time              `json:"escalateAfter,omitempty"`
	ResolvedAt        time.Time              `json:"resolvedAt,omitempty"`
	NotificationSent  bool                   `json:"notificationSent"`
	NotifiedTo        []string               `json:"notifiedTo,omitempty"`
	IsTest            bool                   `json:"isTest"`
	TagsMeta          []string               `json:"tagsMeta,omitempty"`
}

// EscalationEvent representa um evento de escalonamento de alerta
type EscalationEvent struct {
	EscalationID        string    `json:"escalationId"`
	AlertID             string    `json:"alertId"`
	EscalationLevel     int       `json:"escalationLevel"`
	EscalationReason    string    `json:"escalationReason"`
	PreviousAssignedTo  string    `json:"previousAssignedTo,omitempty"`
	NewAssignedTo       string    `json:"newAssignedTo,omitempty"`
	PreviousPriority    FraudAlertPriority `json:"previousPriority"`
	NewPriority         FraudAlertPriority `json:"newPriority"`
	EscalatedByUserID   string    `json:"escalatedByUserId,omitempty"`
	EscalatedBySystem   bool      `json:"escalatedBySystem"`
	Notes               string    `json:"notes,omitempty"`
	Timestamp           time.Time `json:"timestamp"`
}

// AlertNotification representa uma notificação relacionada a um alerta de fraude
type AlertNotification struct {
	NotificationID  string                 `json:"notificationId"`
	AlertID         string                 `json:"alertId"`
	Type            string                 `json:"type"` // "email", "sms", "push", "webhook", "api"
	Recipient       string                 `json:"recipient"`
	RecipientType   string                 `json:"recipientType"` // "user", "team", "system"
	Title           string                 `json:"title"`
	Message         string                 `json:"message"`
	Priority        int                    `json:"priority"`
	DeliveryStatus  string                 `json:"deliveryStatus,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
	RequiresAck     bool                   `json:"requiresAck"`
	AcknowledgedAt  time.Time              `json:"acknowledgedAt,omitempty"`
	AcknowledgedBy  string                 `json:"acknowledgedBy,omitempty"`
	Data            map[string]interface{} `json:"data,omitempty"`
	RetryCount      int                    `json:"retryCount"`
	ScheduledFor    time.Time              `json:"scheduledFor,omitempty"`
}

// EscalationRule define uma regra para escalonamento automático
type EscalationRule struct {
	RuleID              string                 `json:"ruleId"`
	TenantID            string                 `json:"tenantId"`
	Name                string                 `json:"name"`
	Description         string                 `json:"description"`
	RegionCodes         []string               `json:"regionCodes"`
	ContextIDs          []string               `json:"contextIds,omitempty"`
	AlertCategories     []string               `json:"alertCategories,omitempty"` // se vazio, aplica-se a todas
	AlertLevels         []FraudAlertLevel      `json:"alertLevels,omitempty"`     // se vazio, aplica-se a todas
	MinPriority         FraudAlertPriority     `json:"minPriority"`
	MinRiskScore        float64                `json:"minRiskScore"`
	TimeThresholdMinutes int                   `json:"timeThresholdMinutes"`
	EscalationTargets   []EscalationTarget     `json:"escalationTargets"`
	Actions             []FraudAlertAction     `json:"actions,omitempty"`
	AdditionalCriteria  map[string]interface{} `json:"additionalCriteria,omitempty"`
	IsEnabled           bool                   `json:"isEnabled"`
	CreatedAt           time.Time              `json:"createdAt"`
	UpdatedAt           time.Time              `json:"updatedAt"`
}

// EscalationTarget define um alvo para escalonamento de alertas
type EscalationTarget struct {
	TargetID         string `json:"targetId"`
	TargetType       string `json:"targetType"` // "user", "team", "role", "system"
	TargetIdentifier string `json:"targetIdentifier"`
	EscalationLevel  int    `json:"escalationLevel"`
	NotificationMethods []string `json:"notificationMethods,omitempty"` // "email", "sms", etc.
	IsActive         bool   `json:"isActive"`
}

// RegionSpecificConfig contém configurações específicas por região
type RegionSpecificConfig struct {
	RegionCode            string                   `json:"regionCode"`
	AlertThresholds       map[string]float64       `json:"alertThresholds"`
	RequiredActions       map[FraudAlertLevel][]FraudAlertAction `json:"requiredActions"`
	DefaultEscalationTeam string                   `json:"defaultEscalationTeam"`
	EscalationTimeouts    map[FraudAlertPriority]int `json:"escalationTimeouts"` // minutos
	AutoResolveTimeouts   map[FraudAlertLevel]int    `json:"autoResolveTimeouts"` // minutos
	NotificationTemplates map[string]string        `json:"notificationTemplates"`
	RegionalContacts      map[string]string        `json:"regionalContacts"`
}

// AlertEscalationConfig contém configurações para o sistema de escalonamento
type AlertEscalationConfig struct {
	DefaultEscalationTimeoutMinutes  int                      `json:"defaultEscalationTimeoutMinutes"`
	DefaultAutoResolveTimeoutMinutes int                      `json:"defaultAutoResolveTimeoutMinutes"`
	DefaultNotificationTemplates     map[string]string        `json:"defaultNotificationTemplates"`
	PriorityCalculationWeights       map[string]float64       `json:"priorityCalculationWeights"`
	RegionalConfigs                  []*RegionSpecificConfig  `json:"regionalConfigs"`
	GlobalEscalationRules            []*EscalationRule        `json:"globalEscalationRules"`
	NotificationChannels             map[string]bool          `json:"notificationChannels"`
	AlertExpirationDays              int                      `json:"alertExpirationDays"`
	MaxEscalationLevel               int                      `json:"maxEscalationLevel"`
	EnableBatchNotifications         bool                     `json:"enableBatchNotifications"`
	BatchIntervalMinutes             int                      `json:"batchIntervalMinutes"`
}

// AlertEscalationService gerencia o escalonamento de alertas de fraude
type AlertEscalationService struct {
	config            AlertEscalationConfig
	logger            *logging.Logger
	tracer            trace.Tracer
	metrics           *metrics.MetricsCollector
	notificationSvc   *notification.NotificationService
	escalationRules   map[string][]*EscalationRule // mapeado por regionCode
	regionalConfigs   map[string]*RegionSpecificConfig // mapeado por regionCode
	activeAlerts      sync.Map // map[string]*FraudAlert (alertID -> alerta)
	kafkaWriter       *kafka.Writer
	ctx               context.Context
	cancel            context.CancelFunc
	mu                sync.RWMutex
}

// NewAlertEscalationService cria uma nova instância do serviço de escalonamento
func NewAlertEscalationService(
	config AlertEscalationConfig,
	logger *logging.Logger,
	notificationSvc *notification.NotificationService,
) (*AlertEscalationService, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Configurar tracer para OpenTelemetry
	t := tracing.GetTracer("alert-escalation-service")
	
	// Configurar coletor de métricas
	m := metrics.NewMetricsCollector("fraud_alert_escalation")
	
	// Organizar regras e configurações por região
	escalationRules := make(map[string][]*EscalationRule)
	regionalConfigs := make(map[string]*RegionSpecificConfig)
	
	// Mapear configurações regionais
	for _, regionalConfig := range config.RegionalConfigs {
		regionalConfigs[regionalConfig.RegionCode] = regionalConfig
	}
	
	// Mapear regras de escalonamento global para cada região
	for _, rule := range config.GlobalEscalationRules {
		for _, regionCode := range rule.RegionCodes {
			if _, exists := escalationRules[regionCode]; !exists {
				escalationRules[regionCode] = []*EscalationRule{}
			}
			escalationRules[regionCode] = append(escalationRules[regionCode], rule)
		}
	}
	
	service := &AlertEscalationService{
		config:          config,
		logger:          logger,
		tracer:          t,
		metrics:         m,
		notificationSvc: notificationSvc,
		escalationRules: escalationRules,
		regionalConfigs: regionalConfigs,
		ctx:             ctx,
		cancel:          cancel,
	}
	
	return service, nil
}

// Start inicia o serviço de escalonamento de alertas
func (s *AlertEscalationService) Start() error {
	s.logger.Info(s.ctx, "Iniciando serviço de escalonamento de alertas de fraude")
	
	// Iniciar goroutine para processar escalonamentos pendentes
	go s.startEscalationProcessor()
	
	// Iniciar goroutine para processar alertas expirados
	go s.startAlertExpirationProcessor()
	
	// Iniciar goroutine para processamento em lote de notificações, se habilitado
	if s.config.EnableBatchNotifications {
		go s.startBatchNotificationProcessor()
	}
	
	s.logger.Info(s.ctx, "Serviço de escalonamento de alertas iniciado com sucesso")
	return nil
}

// Stop interrompe o serviço de escalonamento de alertas
func (s *AlertEscalationService) Stop() error {
	s.logger.Info(s.ctx, "Parando serviço de escalonamento de alertas")
	
	// Cancelar contexto para interromper todas as goroutines
	s.cancel()
	
	// Fechar o writer do Kafka, se existir
	if s.kafkaWriter != nil {
		if err := s.kafkaWriter.Close(); err != nil {
			s.logger.Error(s.ctx, "Erro ao fechar Kafka writer: %v", err)
		}
	}
	
	s.logger.Info(s.ctx, "Serviço de escalonamento de alertas parado com sucesso")
	return nil
}

// ProcessAlert processa um novo alerta e aplica a lógica de priorização e escalonamento
func (s *AlertEscalationService) ProcessAlert(
	ctx context.Context,
	detectionResult *FraudDetectionResult,
	request *FraudDetectionRequest,
) (*FraudAlert, error) {
	ctx, span := s.tracer.Start(ctx, "AlertEscalationService.ProcessAlert")
	defer span.End()
	
	// Criar um novo alerta a partir do resultado de detecção
	alert := s.createAlertFromDetectionResult(detectionResult, request)
	
	// Calcular a prioridade do alerta
	alert.Priority = s.calculateAlertPriority(ctx, alert, detectionResult)
	
	// Aplicar regras específicas da região
	s.applyRegionalRules(ctx, alert)
	
	// Determinar ações necessárias
	s.determineRequiredActions(ctx, alert)
	
	// Definir tempo de escalonamento automático
	s.setEscalationTime(ctx, alert)
	
	// Salvar o alerta como ativo
	s.activeAlerts.Store(alert.AlertID, alert)
	
	// Incrementar métrica de alertas criados
	s.metrics.IncrementCounter("alerts_created_total", 1, map[string]string{
		"region_code": alert.RegionCode,
		"level":       fmt.Sprintf("%d", alert.AlertLevel),
		"category":    alert.Category,
		"priority":    fmt.Sprintf("%d", alert.Priority),
	})
	
	// Se não for um teste, processar notificações e ações
	if !alert.IsTest {
		// Enviar notificações imediatas, se necessário
		if s.shouldNotifyImmediately(alert) {
			s.sendAlertNotifications(ctx, alert)
		}
		
		// Executar ações automáticas imediatas
		s.executeRequiredActions(ctx, alert)
	}
	
	s.logger.Info(ctx, "Alerta de fraude processado: %s (nível: %d, prioridade: %d)", 
		alert.AlertID, alert.AlertLevel, alert.Priority)
	
	return alert, nil
}

// UpdateAlertStatus atualiza o status de um alerta
func (s *AlertEscalationService) UpdateAlertStatus(
	ctx context.Context,
	alertID string,
	newStatus FraudAlertStatus,
	updatedByUserID string,
	notes string,
) error {
	ctx, span := s.tracer.Start(ctx, "AlertEscalationService.UpdateAlertStatus")
	defer span.End()
	
	// Obter o alerta ativo
	alertValue, exists := s.activeAlerts.Load(alertID)
	if !exists {
		return fmt.Errorf("alerta com ID %s não encontrado", alertID)
	}
	
	alert := alertValue.(*FraudAlert)
	oldStatus := alert.Status
	
	// Atualizar o status
	alert.Status = newStatus
	alert.UpdatedAt = time.Now()
	
	// Se for resolvido ou fechado, definir ResolvedAt
	if newStatus == AlertStatusResolved || newStatus == AlertStatusClosed {
		alert.ResolvedAt = time.Now()
	}
	
	// Registrar atualização
	s.logger.Info(ctx, "Status do alerta %s atualizado: %s -> %s", alertID, oldStatus, newStatus)
	
	// Incrementar métrica de atualizações de status
	s.metrics.IncrementCounter("alert_status_updates_total", 1, map[string]string{
		"old_status": string(oldStatus),
		"new_status": string(newStatus),
		"region":     alert.RegionCode,
	})
	
	// Se o alerta foi fechado ou marcado como falso alarme, remover da lista de ativos
	if newStatus == AlertStatusClosed || newStatus == AlertStatusFalseAlarm {
		s.activeAlerts.Delete(alertID)
	} else {
		// Atualizar o alerta na lista de ativos
		s.activeAlerts.Store(alertID, alert)
	}
	
	// Se necessário, enviar notificação sobre mudança de status
	if s.shouldNotifyStatusChange(oldStatus, newStatus) {
		s.sendStatusChangeNotification(ctx, alert, oldStatus, updatedByUserID, notes)
	}
	
	return nil
}

// EscalateAlert realiza o escalonamento manual de um alerta
func (s *AlertEscalationService) EscalateAlert(
	ctx context.Context,
	alertID string,
	targetTeam string,
	escalationLevel int,
	reason string,
	escalatedByUserID string,
) error {
	ctx, span := s.tracer.Start(ctx, "AlertEscalationService.EscalateAlert")
	defer span.End()
	
	// Obter o alerta
	alertValue, exists := s.activeAlerts.Load(alertID)
	if !exists {
		return fmt.Errorf("alerta com ID %s não encontrado", alertID)
	}
	
	alert := alertValue.(*FraudAlert)
	
	// Criar evento de escalonamento
	escalationEvent := &EscalationEvent{
		EscalationID:       fmt.Sprintf("esc-%s-%d", alertID, time.Now().UnixNano()),
		AlertID:            alertID,
		EscalationLevel:    escalationLevel,
		EscalationReason:   reason,
		PreviousAssignedTo: alert.AssignedToTeam,
		NewAssignedTo:      targetTeam,
		PreviousPriority:   alert.Priority,
		NewPriority:        FraudAlertPriority(min(int(alert.Priority)+1, int(PriorityCritical))), // Aumentar prioridade
		EscalatedByUserID:  escalatedByUserID,
		EscalatedBySystem:  false,
		Notes:              reason,
		Timestamp:          time.Now(),
	}
	
	// Atualizar o alerta
	alert.EscalationLevel = escalationLevel
	alert.AssignedToTeam = targetTeam
	alert.Priority = escalationEvent.NewPriority
	alert.Status = AlertStatusEscalated
	alert.UpdatedAt = time.Now()
	
	// Adicionar ao histórico de escalonamento
	if alert.EscalationHistory == nil {
		alert.EscalationHistory = []*EscalationEvent{}
	}
	alert.EscalationHistory = append(alert.EscalationHistory, escalationEvent)
	
	// Atualizar o alerta na lista de ativos
	s.activeAlerts.Store(alertID, alert)
	
	// Registrar o escalonamento
	s.logger.Info(ctx, "Alerta %s escalonado para nível %d (equipe: %s)", 
		alertID, escalationLevel, targetTeam)
	
	// Incrementar métrica de escalonamento
	s.metrics.IncrementCounter("alert_escalations_total", 1, map[string]string{
		"region":           alert.RegionCode,
		"escalation_level": fmt.Sprintf("%d", escalationLevel),
		"manual":           "true",
	})
	
	// Enviar notificação de escalonamento
	s.sendEscalationNotification(ctx, alert, escalationEvent)
	
	return nil
}

// GetActiveAlerts retorna todos os alertas ativos que correspondem aos critérios
func (s *AlertEscalationService) GetActiveAlerts(
	ctx context.Context,
	tenantID string,
	regionCode string,
	minLevel FraudAlertLevel,
	statuses []FraudAlertStatus,
) []*FraudAlert {
	ctx, span := s.tracer.Start(ctx, "AlertEscalationService.GetActiveAlerts")
	defer span.End()
	
	var result []*FraudAlert
	
	// Converter os status para um mapa para busca mais eficiente
	statusMap := make(map[FraudAlertStatus]bool)
	for _, status := range statuses {
		statusMap[status] = true
	}
	
	// Percorrer todos os alertas ativos
	s.activeAlerts.Range(func(_, value interface{}) bool {
		alert := value.(*FraudAlert)
		
		// Aplicar filtros
		matchesTenant := tenantID == "" || alert.TenantID == tenantID
		matchesRegion := regionCode == "" || alert.RegionCode == regionCode
		matchesLevel := alert.AlertLevel >= minLevel
		matchesStatus := len(statusMap) == 0 || statusMap[alert.Status]
		
		if matchesTenant && matchesRegion && matchesLevel && matchesStatus {
			result = append(result, alert)
		}
		
		return true
	})
	
	// Ordenar por prioridade e nível
	sort.Slice(result, func(i, j int) bool {
		if result[i].Priority == result[j].Priority {
			return result[i].AlertLevel > result[j].AlertLevel // maior nível de alerta primeiro
		}
		return result[i].Priority > result[j].Priority // maior prioridade primeiro
	})
	
	return result
}

// Métodos auxiliares e processadores em segundo plano a serem implementados posteriormente