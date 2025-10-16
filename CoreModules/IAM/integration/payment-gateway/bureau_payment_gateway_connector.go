package paymentgateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	cv "github.com/innovabizdevops/innovabiz-iam/integration/cross-verification"
	"github.com/innovabizdevops/innovabiz-iam/observability/adapter"
	"github.com/innovabizdevops/innovabiz-iam/observability/logging"
	"github.com/innovabizdevops/innovabiz-iam/observability/metrics"
	"github.com/innovabizdevops/innovabiz-iam/observability/tracing"
)

// Definição de constantes
const (
	// Status de transação
	TransactionStatusPending   = "pending"
	TransactionStatusApproved  = "approved"
	TransactionStatusDenied    = "denied"
	TransactionStatusChallenged = "challenged"
	TransactionStatusError     = "error"
	
	// Níveis de verificação
	VerificationLevelBasic     = "basic"
	VerificationLevelStandard  = "standard"
	VerificationLevelAdvanced  = "advanced"
	VerificationLevelPremium   = "premium"
	
	// Cache TTL padrão
	DefaultCacheTTL            = 5 * time.Minute
	
	// Limites de transação padrão
	DefaultTransactionTimeoutSec = 30
)

// BureauPaymentGatewayConfig contém configurações para o conector
type BureauPaymentGatewayConfig struct {
	VerifierURL                 string                          `json:"verifier_url"`
	BureauServiceURL            string                          `json:"bureau_service_url"`
	PaymentGatewayURL           string                          `json:"payment_gateway_url"`
	EnableTrustScoring          bool                            `json:"enable_trust_scoring"`
	EnableCaching               bool                            `json:"enable_caching"`
	CacheTTL                    time.Duration                   `json:"cache_ttl"`
	TransactionTimeoutSec       int                             `json:"transaction_timeout_sec"`
	EnableConcurrentProcessing  bool                            `json:"enable_concurrent_processing"`
	MaxConcurrentVerifications  int                             `json:"max_concurrent_verifications"`
	TransactionLimits           map[string]TransactionLimit     `json:"transaction_limits"`
	RegionalSettings            map[string]RegionalConnectorSettings `json:"regional_settings"`
	VerificationRules           VerificationRules               `json:"verification_rules"`
}

// TransactionLimit define limites para transações
type TransactionLimit struct {
	SingleTransactionMax        float64 `json:"single_transaction_max"`
	DailyTransactionMax         float64 `json:"daily_transaction_max"`
	MonthlyTransactionMax       float64 `json:"monthly_transaction_max"`
	RequiredTrustScore          int     `json:"required_trust_score"`
	RequiresEnhancedVerification bool    `json:"requires_enhanced_verification"`
}

// RegionalConnectorSettings define configurações específicas por região
type RegionalConnectorSettings struct {
	MinTrustScore               int      `json:"min_trust_score"`
	RequiredVerificationLevel   string   `json:"required_verification_level"`
	MandatoryFields             []string `json:"mandatory_fields"`
	ComplianceRules             []string `json:"compliance_rules"`
	AllowedPaymentMethods       []string `json:"allowed_payment_methods"`
}

// VerificationRules define regras de verificação para diferentes cenários
type VerificationRules struct {
	HighValueTransactions       RuleSet  `json:"high_value_transactions"`
	NewAccounts                 RuleSet  `json:"new_accounts"`
	InternationalTransactions   RuleSet  `json:"international_transactions"`
	RecurringTransactions       RuleSet  `json:"recurring_transactions"`
	HighRiskCategories          RuleSet  `json:"high_risk_categories"`
}

// RuleSet define um conjunto de regras para verificação
type RuleSet struct {
	MinTrustScore               int      `json:"min_trust_score"`
	RequiredVerificationLevel   string   `json:"required_verification_level"`
	RequiredChallenges          []string `json:"required_challenges"`
	ExtraVerifications          []string `json:"extra_verifications"`
	ApprovalThreshold           int      `json:"approval_threshold"`
}

// BureauPaymentGatewayConnector implementa a integração entre Bureau de Crédito e PaymentGateway
type BureauPaymentGatewayConnector struct {
	config            BureauPaymentGatewayConfig
	logger            logging.Logger
	tracer            tracing.Tracer
	metricsRecorder   metrics.Metrics
	verifier          *cv.CrossVerificationOrchestrator
	transactionCache  sync.Map
	challengeManager  *ChallengeManager
}

// NewBureauPaymentGatewayConnector cria uma nova instância do conector
func NewBureauPaymentGatewayConnector(config BureauPaymentGatewayConfig) (*BureauPaymentGatewayConnector, error) {
	// Inicializar observabilidade
	obsAdapter, err := adapter.NewAdapter(adapter.Config{
		ServiceName: "bureau-payment-gateway-connector",
	})
	
	if err != nil {
		return nil, fmt.Errorf("falha ao inicializar observabilidade: %w", err)
	}
	
	// Configurar valores padrão
	if config.CacheTTL == 0 {
		config.CacheTTL = DefaultCacheTTL
	}
	
	if config.TransactionTimeoutSec == 0 {
		config.TransactionTimeoutSec = DefaultTransactionTimeoutSec
	}
	
	connector := &BureauPaymentGatewayConnector{
		config:          config,
		logger:          obsAdapter.Logger(),
		tracer:          obsAdapter.Tracer(),
		metricsRecorder: obsAdapter.Metrics(),
	}
	
	// Inicializar gerenciador de desafios
	connector.challengeManager = NewChallengeManager(ChallengeManagerConfig{})
	
	// Iniciar limpeza periódica de cache
	if config.EnableCaching {
		go connector.startCacheCleanup()
	}
	
	connector.logger.Info("BureauPaymentGatewayConnector inicializado com sucesso")
	
	return connector, nil
}

// SetVerifier configura o orquestrador de verificação cruzada
func (c *BureauPaymentGatewayConnector) SetVerifier(verifier *cv.CrossVerificationOrchestrator) {
	c.verifier = verifier
	c.logger.Info("Orquestrador de verificação cruzada configurado")
}

// ProcessPayment processa um pagamento com verificação de identidade e dados financeiros
func (c *BureauPaymentGatewayConnector) ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	ctx, span := c.tracer.StartSpan(ctx, "BureauPaymentGatewayConnector.ProcessPayment")
	defer span.End()
	
	start := time.Now()
	
	c.logger.InfoWithContext(ctx, "Iniciando processamento de pagamento com verificação",
		"request_id", req.RequestID,
		"user_id", req.UserID,
		"tenant_id", req.TenantID,
		"amount", req.Amount,
		"currency", req.Currency)
	
	// Registra métricas de início
	c.metricsRecorder.CounterInc("payment_gateway_requests_total", map[string]string{
		"tenant_id": req.TenantID,
		"region": req.RegionCode,
		"payment_method": req.PaymentMethod,
	})
	
	// Verificar cache para transações repetidas
	if c.config.EnableCaching {
		if cached, found := c.transactionCache.Load(req.RequestID); found {
			c.logger.InfoWithContext(ctx, "Resposta recuperada do cache", "request_id", req.RequestID)
			return cached.(*PaymentResponse), nil
		}
	}
	
	// Gerar ID de transação se não fornecido
	if req.TransactionID == "" {
		req.TransactionID = fmt.Sprintf("tx-%s", uuid.New().String())
	}
	
	// Verificar limites de transação
	limitResult, err := c.checkTransactionLimits(ctx, req)
	if err != nil {
		return c.createErrorResponse(req, "limite_transacao_erro", err.Error()), nil
	}
	
	if !limitResult.Allowed {
		return c.createErrorResponse(req, "limite_excedido", limitResult.Reason), nil
	}
	
	// Determinar o nível de verificação necessário
	verificationLevel, extraChecks := c.determineVerificationLevel(ctx, req)
	
	c.logger.InfoWithContext(ctx, "Nível de verificação determinado",
		"transaction_id", req.TransactionID,
		"verification_level", verificationLevel,
		"extra_checks", extraChecks)
	
	// Construir requisição para verificação cruzada
	verificationReq := &cv.CredentialFinancialVerificationRequest{
		RequestID:       req.RequestID,
		TransactionID:   req.TransactionID,
		UserID:          req.UserID,
		TenantID:        req.TenantID,
		RegionCode:      req.RegionCode,
		VerificationLevel: verificationLevel,
		ContextData:     map[string]interface{}{
			"transaction_amount": req.Amount,
			"transaction_currency": req.Currency,
			"payment_method": req.PaymentMethod,
			"merchant_id": req.MerchantID,
			"merchant_category": req.MerchantCategory,
			"extra_checks": extraChecks,
		},
		FinancialProducts: req.FinancialProducts,
		IdentityData:      mapToIdentityData(req.UserData),
		FinancialData:     mapToFinancialData(req.FinancialData),
		DeviceData:        mapToDeviceData(req.DeviceInfo),
		Timestamp:         time.Now(),
	}
	
	// Executar verificação cruzada
	verificationResp, err := c.performCrossVerification(ctx, req, verificationReq)
	if err != nil {
		c.logger.ErrorWithContext(ctx, "Erro ao executar verificação cruzada",
			"request_id", req.RequestID,
			"transaction_id", req.TransactionID,
			"error", err.Error())
		
		// Registrar falha na métrica
		c.metricsRecorder.CounterInc("payment_verification_errors", map[string]string{
			"region": req.RegionCode,
			"error_type": "verification_failure",
		})
		
		return c.createErrorResponse(req, "verificacao_erro", fmt.Sprintf("Erro ao processar verificação: %s", err.Error())), nil
	}
	
	// Processar resultado da verificação
	response, err := c.processVerificationResult(ctx, req, verificationResp)
	if err != nil {
		c.logger.ErrorWithContext(ctx, "Erro ao processar resultado da verificação",
			"request_id", req.RequestID,
			"transaction_id", req.TransactionID,
			"error", err.Error())
		
		return c.createErrorResponse(req, "processamento_erro", fmt.Sprintf("Erro ao processar resultado: %s", err.Error())), nil
	}
	
	// Calcular tempo total de processamento
	totalProcessingTime := time.Since(start).Milliseconds()
	response.ProcessingTimeMs = totalProcessingTime
	
	// Registrar métricas de desempenho
	c.metricsRecorder.HistogramObserve("payment_processing_time_ms", float64(totalProcessingTime), map[string]string{
		"region": req.RegionCode,
		"status": response.Status,
	})
	
	c.logger.InfoWithContext(ctx, "Processamento de pagamento finalizado",
		"request_id", req.RequestID,
		"transaction_id", req.TransactionID,
		"status", response.Status,
		"processing_time_ms", totalProcessingTime)
	
	return response, nil