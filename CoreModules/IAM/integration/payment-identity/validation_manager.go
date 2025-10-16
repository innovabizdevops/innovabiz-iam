package paymentidentity

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/innovabizdevops/innovabiz-iam/observability/adapter"
	"github.com/innovabizdevops/innovabiz-iam/observability/logging"
	"github.com/innovabizdevops/innovabiz-iam/observability/metrics"
	"github.com/innovabizdevops/innovabiz-iam/observability/tracing"
)

// ValidationManagerConfig contém configurações para o ValidationManager
type ValidationManagerConfig struct {
	RequestTimeout         time.Duration            `json:"request_timeout"`
	ChallengeExpiration    time.Duration            `json:"challenge_expiration"`
	ValidationExpiration   time.Duration            `json:"validation_expiration"`
	MaxRetries             int                      `json:"max_retries"`
	RegionalSettings       map[string]RegionalConfig `json:"regional_settings"`
	RiskThresholds         RiskThresholdConfig      `json:"risk_thresholds"`
	ObservabilityEnabled   bool                     `json:"observability_enabled"`
}

// RegionalConfig contém configurações específicas para cada região
type RegionalConfig struct {
	ValidationLevelMapping map[string]string        `json:"validation_level_mapping"`
	ComplianceRules        map[string]interface{}   `json:"compliance_rules"`
	RiskFactorWeights      map[string]float64       `json:"risk_factor_weights"`
	RequiredChallenges     map[string][]string      `json:"required_challenges"`
	PaymentMethodSettings  map[string]interface{}   `json:"payment_method_settings"`
}

// RiskThresholdConfig define limites para categorização de risco
type RiskThresholdConfig struct {
	LowRiskThreshold      int `json:"low_risk_threshold"`
	MediumRiskThreshold   int `json:"medium_risk_threshold"`
	HighValueTransaction  float64 `json:"high_value_transaction"`
	TrustedUserThreshold  int `json:"trusted_user_threshold"`
}

// ValidationEngine é uma interface para motor de validação de identidade
type ValidationEngine interface {
	ValidateIdentity(ctx context.Context, req PaymentIdentityRequest) (*ValidationResult, error)
	ValidateChallenge(ctx context.Context, challengeID string, response map[string]interface{}) (bool, error)
	GetValidationLevel(ctx context.Context, req PaymentIdentityRequest) string
}

// ValidationResult contém o resultado da validação de identidade
type ValidationResult struct {
	TrustScore         int                    `json:"trust_score"`
	RiskFactors        []string               `json:"risk_factors"`
	ValidationLevel    string                 `json:"validation_level"`
	ComplianceResults  map[string]bool        `json:"compliance_results"`
	RequiresChallenge  bool                   `json:"requires_challenge"`
	ChallengeType      string                 `json:"challenge_type,omitempty"`
	ChallengeDetails   *ChallengeDetails      `json:"challenge_details,omitempty"`
	AdditionalData     map[string]interface{} `json:"additional_data,omitempty"`
}

// ValidationManager gerencia o fluxo de validação de identidade para pagamentos
type ValidationManager struct {
	config           ValidationManagerConfig
	validationEngine ValidationEngine
	logger           logging.Logger
	tracer           tracing.Tracer
	metricsRecorder  metrics.Recorder
	activeSessions   sync.Map
}

// ValidationSession representa uma sessão de validação de identidade
type ValidationSession struct {
	RequestID       string
	Request         PaymentIdentityRequest
	ValidationLevel string
	Result          *ValidationResult
	Status          string
	Created         time.Time
	Expires         time.Time
	ChallengeID     string
	Attempts        int
}

// NewValidationManager cria uma nova instância do ValidationManager
func NewValidationManager(config ValidationManagerConfig, engine ValidationEngine) (*ValidationManager, error) {
	// Inicializa observabilidade
	obsAdapter, err := adapter.NewAdapter(adapter.Config{
		ServiceName: "payment-identity-validation",
	})
	
	if err != nil {
		return nil, fmt.Errorf("falha ao inicializar observabilidade: %w", err)
	}

	manager := &ValidationManager{
		config:           config,
		validationEngine: engine,
		logger:           obsAdapter.Logger(),
		tracer:           obsAdapter.Tracer(),
		metricsRecorder:  obsAdapter.Metrics(),
	}

	// Inicializar limpeza periódica de sessões expiradas
	go manager.startSessionCleanup()

	manager.logger.Info("ValidationManager inicializado com sucesso")
	
	return manager, nil
}

// Inicia o processo de validação de identidade para um pagamento
func (m *ValidationManager) ValidatePaymentIdentity(ctx context.Context, req PaymentIdentityRequest) (*PaymentIdentityResponse, error) {
	ctx, span := m.tracer.StartSpan(ctx, "ValidationManager.ValidatePaymentIdentity")
	defer span.End()
	
	start := time.Now()
	
	// Adiciona timestamp se não estiver presente
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}
	
	// Registra início do processo
	m.logger.InfoWithContext(ctx, "Iniciando validação de identidade para pagamento",
		"request_id", req.RequestID,
		"transaction_id", req.TransactionID,
		"user_id", req.UserID,
		"tenant_id", req.TenantID,
		"region", req.RegionCode,
		"payment_method", req.PaymentMethod,
	)
	
	// Registra métricas de solicitação
	m.metricsRecorder.CounterInc("payment_identity_validation_requests_total", map[string]string{
		"tenant_id":      req.TenantID,
		"region":         req.RegionCode,
		"payment_method": req.PaymentMethod,
	})
	
	// Determina o nível de validação a ser aplicado
	validationLevel := m.validationEngine.GetValidationLevel(ctx, req)
	
	// Cria e armazena sessão de validação
	session := &ValidationSession{
		RequestID:       req.RequestID,
		Request:         req,
		ValidationLevel: validationLevel,
		Status:          StatusInProgress,
		Created:         time.Now(),
		Expires:         time.Now().Add(m.config.ValidationExpiration),
		Attempts:        0,
	}
	
	m.activeSessions.Store(req.RequestID, session)
	
	// Executa validação de identidade
	result, err := m.validationEngine.ValidateIdentity(ctx, req)
	if err != nil {
		m.logger.ErrorWithContext(ctx, "Erro ao validar identidade", 
			"error", err.Error(),
			"request_id", req.RequestID)
		
		m.metricsRecorder.CounterInc("payment_identity_validation_errors_total", map[string]string{
			"error_type": "validation_failure",
			"region": req.RegionCode,
		})
		
		session.Status = StatusFailed
		m.activeSessions.Store(req.RequestID, session)
		
		return nil, err
	}
	
	// Atualiza sessão com resultado
	session.Result = result
	
	// Determina status final da validação
	var finalStatus string
	if result.RequiresChallenge {
		finalStatus = StatusPending
		session.ChallengeID = result.ChallengeDetails.ChallengeID
	} else if len(result.RiskFactors) > 0 && result.TrustScore < m.config.RiskThresholds.TrustedUserThreshold {
		finalStatus = StatusRejected
	} else {
		finalStatus = StatusCompleted
	}
	
	session.Status = finalStatus
	m.activeSessions.Store(req.RequestID, session)
	
	// Prepara resposta
	response := &PaymentIdentityResponse{
		RequestID:              req.RequestID,
		TransactionID:          req.TransactionID,
		Status:                 finalStatus,
		AppliedValidationLevel: validationLevel,
		RiskCategory:           getRiskCategory(result.TrustScore, m.config.RiskThresholds),
		TrustScore:             result.TrustScore,
		Verified:               finalStatus == StatusCompleted,
		ChallengeRequired:      result.RequiresChallenge,
		RiskFactors:            result.RiskFactors,
		ProcessingTimeMs:       time.Since(start).Milliseconds(),
		ComplianceResults:      result.ComplianceResults,
		ValidationExpiry:       session.Expires,
	}
	
	// Adiciona detalhes do desafio se necessário
	if result.RequiresChallenge && result.ChallengeDetails != nil {
		response.ChallengeType = result.ChallengeDetails.ChallengeType
		response.ChallengeDetails = result.ChallengeDetails
	}
	
	// Determina ação recomendada
	response.RecommendedAction = determineRecommendedAction(result, finalStatus)
	
	// Adiciona informações de auditoria
	response.AuditInfo = map[string]interface{}{
		"validation_timestamp": time.Now().Format(time.RFC3339),
		"validation_version":   "1.0",
		"validation_level":     validationLevel,
	}
	
	// Registra métricas de resultado
	m.metricsRecorder.HistogramObserve("payment_identity_validation_duration_ms", float64(response.ProcessingTimeMs), map[string]string{
		"region":      req.RegionCode,
		"status":      finalStatus,
		"challenged":  fmt.Sprintf("%t", result.RequiresChallenge),
	})
	
	m.logger.InfoWithContext(ctx, "Validação de identidade para pagamento concluída",
		"request_id", req.RequestID,
		"status", finalStatus,
		"trust_score", result.TrustScore,
		"challenge_required", result.RequiresChallenge,
		"processing_time_ms", response.ProcessingTimeMs,
	)
	
	return response, nil
}

// VerifyChallengeResponse verifica a resposta a um desafio
func (m *ValidationManager) VerifyChallengeResponse(ctx context.Context, requestID, challengeID string, response map[string]interface{}) (*PaymentIdentityResponse, error) {
	ctx, span := m.tracer.StartSpan(ctx, "ValidationManager.VerifyChallengeResponse")
	defer span.End()
	
	// Recupera a sessão
	sessionObj, found := m.activeSessions.Load(requestID)
	if !found {
		return nil, errors.New("sessão de validação não encontrada ou expirada")
	}
	
	session, ok := sessionObj.(*ValidationSession)
	if !ok {
		return nil, errors.New("erro ao recuperar sessão de validação")
	}
	
	// Verifica se o desafio corresponde
	if session.ChallengeID != challengeID {
		return nil, errors.New("ID de desafio inválido")
	}
	
	// Verifica se a sessão expirou
	if time.Now().After(session.Expires) {
		session.Status = StatusExpired
		m.activeSessions.Store(requestID, session)
		return nil, errors.New("sessão de validação expirada")
	}
	
	// Incrementa contagem de tentativas
	session.Attempts++
	
	// Verifica resposta ao desafio
	verified, err := m.validationEngine.ValidateChallenge(ctx, challengeID, response)
	if err != nil {
		m.logger.ErrorWithContext(ctx, "Erro ao validar resposta ao desafio", 
			"error", err.Error(),
			"request_id", requestID,
			"challenge_id", challengeID)
			
		if session.Attempts >= m.config.MaxRetries {
			session.Status = StatusRejected
			m.activeSessions.Store(requestID, session)
		}
			
		return nil, err
	}
	
	// Atualiza status da sessão
	if verified {
		session.Status = StatusCompleted
	} else {
		if session.Attempts >= m.config.MaxRetries {
			session.Status = StatusRejected
		}
	}
	
	m.activeSessions.Store(requestID, session)
	
	// Prepara resposta
	result := &PaymentIdentityResponse{
		RequestID:              session.Request.RequestID,
		TransactionID:          session.Request.TransactionID,
		Status:                 session.Status,
		AppliedValidationLevel: session.ValidationLevel,
		RiskCategory:           getRiskCategory(session.Result.TrustScore, m.config.RiskThresholds),
		TrustScore:             session.Result.TrustScore,
		Verified:               verified,
		ChallengeRequired:      !verified && session.Attempts < m.config.MaxRetries,
		RiskFactors:            session.Result.RiskFactors,
		ComplianceResults:      session.Result.ComplianceResults,
		ValidationExpiry:       session.Expires,
		AuditInfo: map[string]interface{}{
			"validation_timestamp": time.Now().Format(time.RFC3339),
			"challenge_verified":   verified,
			"attempts":             session.Attempts,
		},
	}
	
	// Adiciona detalhes do desafio se ainda for necessário
	if result.ChallengeRequired {
		result.ChallengeType = session.Result.ChallengeType
		result.ChallengeDetails = session.Result.ChallengeDetails
	}
	
	// Determina ação recomendada
	result.RecommendedAction = determineRecommendedAction(session.Result, session.Status)
	
	m.logger.InfoWithContext(ctx, "Verificação de resposta ao desafio concluída",
		"request_id", requestID,
		"challenge_id", challengeID,
		"verified", verified,
		"status", session.Status,
		"attempts", session.Attempts,
	)
	
	return result, nil
}

// Retorna o status de uma validação em andamento
func (m *ValidationManager) GetValidationStatus(ctx context.Context, requestID string) (*PaymentIdentityResponse, error) {
	ctx, span := m.tracer.StartSpan(ctx, "ValidationManager.GetValidationStatus")
	defer span.End()
	
	// Recupera a sessão
	sessionObj, found := m.activeSessions.Load(requestID)
	if !found {
		return nil, errors.New("sessão de validação não encontrada")
	}
	
	session, ok := sessionObj.(*ValidationSession)
	if !ok {
		return nil, errors.New("erro ao recuperar sessão de validação")
	}
	
	// Verifica se a sessão expirou
	if time.Now().After(session.Expires) && session.Status != StatusCompleted {
		session.Status = StatusExpired
		m.activeSessions.Store(requestID, session)
	}
	
	// Prepara resposta
	result := &PaymentIdentityResponse{
		RequestID:              session.Request.RequestID,
		TransactionID:          session.Request.TransactionID,
		Status:                 session.Status,
		AppliedValidationLevel: session.ValidationLevel,
		ValidationExpiry:       session.Expires,
		AuditInfo: map[string]interface{}{
			"created": session.Created.Format(time.RFC3339),
		},
	}
	
	// Adiciona detalhes completos se o resultado estiver disponível
	if session.Result != nil {
		result.RiskCategory = getRiskCategory(session.Result.TrustScore, m.config.RiskThresholds)
		result.TrustScore = session.Result.TrustScore
		result.Verified = session.Status == StatusCompleted
		result.ChallengeRequired = session.Status == StatusPending
		result.RiskFactors = session.Result.RiskFactors
		result.ComplianceResults = session.Result.ComplianceResults
		
		// Adiciona detalhes do desafio se ainda for necessário
		if result.ChallengeRequired {
			result.ChallengeType = session.Result.ChallengeType
			result.ChallengeDetails = session.Result.ChallengeDetails
		}
		
		// Determina ação recomendada
		result.RecommendedAction = determineRecommendedAction(session.Result, session.Status)
	}
	
	m.logger.InfoWithContext(ctx, "Status de validação consultado",
		"request_id", requestID,
		"status", session.Status,
	)
	
	return result, nil
}

// Remove sessões expiradas
func (m *ValidationManager) startSessionCleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		count := 0
		m.activeSessions.Range(func(key, value interface{}) bool {
			session, ok := value.(*ValidationSession)
			if !ok {
				m.activeSessions.Delete(key)
				count++
				return true
			}
			
			if time.Now().After(session.Expires) {
				m.activeSessions.Delete(key)
				count++
			}
			
			return true
		})
		
		if count > 0 {
			m.logger.Info("Limpeza de sessões expiradas concluída", "removed_sessions", count)
		}
	}
}

// Determina categoria de risco com base na pontuação de confiança
func getRiskCategory(trustScore int, thresholds RiskThresholdConfig) string {
	if trustScore >= thresholds.LowRiskThreshold {
		return RiskCategoryLow
	} else if trustScore >= thresholds.MediumRiskThreshold {
		return RiskCategoryMedium
	}
	return RiskCategoryHigh
}

// Determina ação recomendada com base no resultado da validação e status
func determineRecommendedAction(result *ValidationResult, status string) string {
	if status == StatusCompleted {
		return "approve"
	} else if status == StatusRejected || status == StatusExpired {
		return "reject"
	} else if status == StatusFailed {
		return "retry"
	} else if status == StatusPending {
		return "challenge"
	}
	return "review"
}