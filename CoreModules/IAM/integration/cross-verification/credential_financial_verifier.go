package crossverification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"sync"

	"github.com/innovabizdevops/innovabiz-iam/constants"
	"github.com/innovabizdevops/innovabiz-iam/observability/adapter"
	"github.com/innovabizdevops/innovabiz-iam/observability/logging"
	"github.com/innovabizdevops/innovabiz-iam/observability/metrics"
	"github.com/innovabizdevops/innovabiz-iam/observability/tracing"
)

// Constantes para o verificador cruzado
const (
	// Níveis de confiança
	TrustLevelVeryHigh = "very_high" // 90-100
	TrustLevelHigh     = "high"      // 70-89
	TrustLevelMedium   = "medium"    // 50-69
	TrustLevelLow      = "low"       // 30-49
	TrustLevelVeryLow  = "very_low"  // 0-29

	// Status da verificação
	VerificationStatusPassed  = "passed"
	VerificationStatusFailed  = "failed"
	VerificationStatusPartial = "partial"
	VerificationStatusError   = "error"
	VerificationStatusPending = "pending"

	// Tipos de anomalias
	AnomalyIdentityMismatch   = "identity_mismatch"
	AnomalyFinancialMismatch  = "financial_mismatch"
	AnomalyBehavioralAnomaly  = "behavioral_anomaly"
	AnomalyGeographicAnomaly  = "geographic_anomaly"
	AnomalyDeviceAnomaly      = "device_anomaly"
	AnomalyTransactionPattern = "transaction_pattern"
	AnomalyDocumentAnomaly    = "document_anomaly"
)

// CredentialFinancialVerificationRequest representa uma requisição de verificação cruzada
type CredentialFinancialVerificationRequest struct {
	RequestID         string                 `json:"request_id"`
	TenantID          string                 `json:"tenant_id"`
	UserID            string                 `json:"user_id"`
	IdentityID        string                 `json:"identity_id"`
	TransactionID     string                 `json:"transaction_id,omitempty"`
	AccountID         string                 `json:"account_id"`
	FinancialProducts []string               `json:"financial_products"`
	VerificationLevel string                 `json:"verification_level"`
	IdentityData      IdentityData           `json:"identity_data"`
	FinancialData     FinancialData          `json:"financial_data"`
	DeviceData        DeviceData             `json:"device_data"`
	ContextData       map[string]interface{} `json:"context_data,omitempty"`
	CountryCode       string                 `json:"country_code"`
	RegionCode        string                 `json:"region_code"`
	Timestamp         time.Time              `json:"timestamp"`
}

// CredentialFinancialVerificationResponse representa uma resposta de verificação cruzada
type CredentialFinancialVerificationResponse struct {
	RequestID           string                     `json:"request_id"`
	VerificationID      string                     `json:"verification_id"`
	Status              string                     `json:"status"`
	TrustScore          int                        `json:"trust_score"`
	TrustLevel          string                     `json:"trust_level"`
	VerificationResults map[string]VerificationResult `json:"verification_results"`
	DetectedAnomalies   []Anomaly                  `json:"detected_anomalies,omitempty"`
	ProcessingTimeMs    int64                      `json:"processing_time_ms"`
	ComplianceResults   map[string]bool            `json:"compliance_results"`
	RecommendedAction   string                     `json:"recommended_action"`
	AuditInfo           map[string]interface{}     `json:"audit_info"`
	Timestamp           time.Time                  `json:"timestamp"`
}

// IdentityData contém informações de identidade do usuário
type IdentityData struct {
	DocumentType        string   `json:"document_type"`
	DocumentNumber      string   `json:"document_number"`
	DocumentExpiry      string   `json:"document_expiry,omitempty"`
	FullName            string   `json:"full_name"`
	DateOfBirth         string   `json:"date_of_birth"`
	Nationality         string   `json:"nationality"`
	Address             Address  `json:"address"`
	ContactInfo         ContactInfo `json:"contact_info"`
	VerificationStatus  string   `json:"verification_status"`
	VerificationMethods []string `json:"verification_methods"`
	LastVerified        string   `json:"last_verified"`
	BiometricStatus     string   `json:"biometric_status,omitempty"`
}

// FinancialData contém informações financeiras do usuário
type FinancialData struct {
	AccountDetails        []AccountDetail  `json:"account_details"`
	CreditScore           int              `json:"credit_score"`
	CreditHistory         []CreditEvent    `json:"credit_history,omitempty"`
	FinancialProfile      FinancialProfile `json:"financial_profile"`
	PaymentHistory        []PaymentEvent   `json:"payment_history,omitempty"`
	OutstandingLiabilities []Liability     `json:"outstanding_liabilities,omitempty"`
	IncomeVerification    IncomeVerification `json:"income_verification,omitempty"`
}

// DeviceData contém informações do dispositivo usado na solicitação
type DeviceData struct {
	DeviceID             string `json:"device_id"`
	DeviceType           string `json:"device_type"`
	IPAddress            string `json:"ip_address"`
	UserAgent            string `json:"user_agent"`
	DeviceFingerprint    string `json:"device_fingerprint"`
	GeoLocation          string `json:"geo_location,omitempty"`
	TrustedDevice        bool   `json:"trusted_device"`
	LastAuthLocation     string `json:"last_auth_location,omitempty"`
	LastAuthTime         string `json:"last_auth_time,omitempty"`
	AnomalyScore         int    `json:"anomaly_score"`
	BehavioralScore      int    `json:"behavioral_score,omitempty"`
}

// Address representa um endereço
type Address struct {
	StreetAddress  string `json:"street_address"`
	City           string `json:"city"`
	State          string `json:"state"`
	PostalCode     string `json:"postal_code"`
	Country        string `json:"country"`
	AddressType    string `json:"address_type"`
	VerifiedStatus bool   `json:"verified_status"`
}

// ContactInfo contém dados de contato do usuário
type ContactInfo struct {
	Email           string `json:"email"`
	PhoneNumber     string `json:"phone_number"`
	EmailVerified   bool   `json:"email_verified"`
	PhoneVerified   bool   `json:"phone_verified"`
	PreferredMethod string `json:"preferred_method"`
}

// AccountDetail contém detalhes de uma conta financeira
type AccountDetail struct {
	AccountID         string  `json:"account_id"`
	AccountType       string  `json:"account_type"`
	AccountStatus     string  `json:"account_status"`
	CurrentBalance    float64 `json:"current_balance"`
	AvailableBalance  float64 `json:"available_balance"`
	Currency          string  `json:"currency"`
	LastActivity      string  `json:"last_activity"`
	AccountHolder     string  `json:"account_holder"`
	InstitutionName   string  `json:"institution_name"`
	AccountAgeDays    int     `json:"account_age_days"`
	PrimarySavingsAcc bool    `json:"primary_savings_acc,omitempty"`
	PrimaryCheckingAcc bool   `json:"primary_checking_acc,omitempty"`
}

// CreditEvent representa um evento de crédito
type CreditEvent struct {
	EventType    string  `json:"event_type"`
	EventDate    string  `json:"event_date"`
	Description  string  `json:"description"`
	Amount       float64 `json:"amount,omitempty"`
	Currency     string  `json:"currency,omitempty"`
	Institution  string  `json:"institution,omitempty"`
	ReferenceID  string  `json:"reference_id,omitempty"`
	Severity     string  `json:"severity,omitempty"`
	ImpactScore  int     `json:"impact_score,omitempty"`
	ReportedBy   string  `json:"reported_by,omitempty"`
}

// FinancialProfile contém o perfil financeiro do usuário
type FinancialProfile struct {
	MonthlyIncome           float64            `json:"monthly_income,omitempty"`
	MonthlyExpenses         float64            `json:"monthly_expenses,omitempty"`
	DisposableIncome        float64            `json:"disposable_income,omitempty"`
	DebtToIncomeRatio       float64            `json:"debt_to_income_ratio,omitempty"`
	AvgMonthlyTransactions  int                `json:"avg_monthly_transactions,omitempty"`
	AvgTransactionValue     float64            `json:"avg_transaction_value,omitempty"`
	SpendingCategories      map[string]float64 `json:"spending_categories,omitempty"`
	FinancialRiskScore      int                `json:"financial_risk_score,omitempty"`
	FinancialStabilityScore int                `json:"financial_stability_score,omitempty"`
	WealthTier              string             `json:"wealth_tier,omitempty"`
}

// PaymentEvent representa um evento de pagamento
type PaymentEvent struct {
	PaymentID        string  `json:"payment_id"`
	PaymentType      string  `json:"payment_type"`
	PaymentDate      string  `json:"payment_date"`
	Amount           float64 `json:"amount"`
	Currency         string  `json:"currency"`
	Status           string  `json:"status"`
	Recipient        string  `json:"recipient,omitempty"`
	RecipientAccount string  `json:"recipient_account,omitempty"`
	Description      string  `json:"description,omitempty"`
	Category         string  `json:"category,omitempty"`
	Channel          string  `json:"channel,omitempty"`
}

// Liability representa uma responsabilidade financeira
type Liability struct {
	LiabilityID       string  `json:"liability_id"`
	LiabilityType     string  `json:"liability_type"`
	OriginalAmount    float64 `json:"original_amount"`
	OutstandingAmount float64 `json:"outstanding_amount"`
	Currency          string  `json:"currency"`
	InterestRate      float64 `json:"interest_rate"`
	MonthlyPayment    float64 `json:"monthly_payment"`
	StartDate         string  `json:"start_date"`
	EndDate           string  `json:"end_date"`
	InstitutionName   string  `json:"institution_name"`
	Status            string  `json:"status"`
}

// IncomeVerification contém informações de verificação de renda
type IncomeVerification struct {
	Verified          bool    `json:"verified"`
	VerificationDate  string  `json:"verification_date,omitempty"`
	VerificationMethod string  `json:"verification_method,omitempty"`
	DeclaredIncome    float64 `json:"declared_income,omitempty"`
	VerifiedIncome    float64 `json:"verified_income,omitempty"`
	Currency          string  `json:"currency,omitempty"`
	Confidence        float64 `json:"confidence,omitempty"`
	IncomeSource      string  `json:"income_source,omitempty"`
	FrequencyType     string  `json:"frequency_type,omitempty"`
}

// VerificationResult representa o resultado de uma verificação específica
type VerificationResult struct {
	Category       string                 `json:"category"`
	Status         string                 `json:"status"`
	Score          int                    `json:"score"`
	Description    string                 `json:"description"`
	VerifiedFields []string               `json:"verified_fields,omitempty"`
	FailedFields   []string               `json:"failed_fields,omitempty"`
	Details        map[string]interface{} `json:"details,omitempty"`
}

// Anomaly representa uma anomalia detectada
type Anomaly struct {
	AnomalyType     string                 `json:"anomaly_type"`
	Severity        string                 `json:"severity"`
	Description     string                 `json:"description"`
	DetectionMethod string                 `json:"detection_method"`
	AffectedFields  []string               `json:"affected_fields,omitempty"`
	ConfidenceScore float64                `json:"confidence_score"`
	DetailedInfo    map[string]interface{} `json:"detailed_info,omitempty"`
}

// VerifierConfig contém as configurações do verificador cruzado
type VerifierConfig struct {
	RegionalSettings     map[string]RegionalSettings `json:"regional_settings"`
	VerificationTimeout  time.Duration              `json:"verification_timeout"`
	MinRequiredScore     int                        `json:"min_required_score"`
	EnableCaching        bool                       `json:"enable_caching"`
	CacheTTL             time.Duration              `json:"cache_ttl"`
	ObservabilityEnabled bool                       `json:"observability_enabled"`
	CategoryWeights      map[string]int             `json:"category_weights"`
}

// RegionalSettings contém configurações específicas para cada região
type RegionalSettings struct {
	RequiredFields       map[string][]string       `json:"required_fields"`
	MinimumThresholds    map[string]int            `json:"minimum_thresholds"`
	ComplianceRules      map[string]interface{}    `json:"compliance_rules"`
	ValidationPriorities map[string]int            `json:"validation_priorities"`
}

// CredentialFinancialVerifier gerencia as verificações cruzadas entre credenciais IAM e dados financeiros
type CredentialFinancialVerifier struct {
	config          VerifierConfig
	logger          logging.Logger
	tracer          tracing.Tracer
	metricsRecorder metrics.Recorder
	verifiersMu     sync.RWMutex
	verifiers       map[string]CategoryVerifier
	cacheMu         sync.RWMutex
	cache           map[string]*CredentialFinancialVerificationResponse
}

// CategoryVerifier representa a interface para verificadores de categorias específicas
type CategoryVerifier interface {
	Verify(ctx context.Context, request *CredentialFinancialVerificationRequest) (*VerificationResult, error)
	GetCategory() string
	GetWeight() int
}

// NewCredentialFinancialVerifier cria uma nova instância do verificador cruzado
func NewCredentialFinancialVerifier(config VerifierConfig) (*CredentialFinancialVerifier, error) {
	// Inicializa observabilidade
	obsAdapter, err := adapter.NewAdapter(adapter.Config{
		ServiceName: "credential-financial-verifier",
		Environment: constants.Environment,
	})
	
	if err != nil {
		return nil, fmt.Errorf("falha ao inicializar observabilidade: %w", err)
	}

	verifier := &CredentialFinancialVerifier{
		config:          config,
		logger:          obsAdapter.Logger(),
		tracer:          obsAdapter.Tracer(),
		metricsRecorder: obsAdapter.Metrics(),
		verifiers:       make(map[string]CategoryVerifier),
		cache:           make(map[string]*CredentialFinancialVerificationResponse),
	}

	// Registrar verificadores de categorias
	if err := verifier.registerDefaultVerifiers(); err != nil {
		return nil, fmt.Errorf("falha ao registrar verificadores: %w", err)
	}

	verifier.logger.Info("CredentialFinancialVerifier inicializado com sucesso")
	
	return verifier, nil
}

// Registra os verificadores de categoria padrão
func (v *CredentialFinancialVerifier) registerDefaultVerifiers() error {
	// Inicialize os verificadores de categorias específicas aqui
	// Exemplo: v.RegisterVerifier(NewIdentityVerifier(...))
	return nil
}

// RegisterVerifier registra um verificador de categoria
func (v *CredentialFinancialVerifier) RegisterVerifier(verifier CategoryVerifier) {
	v.verifiersMu.Lock()
	defer v.verifiersMu.Unlock()
	v.verifiers[verifier.GetCategory()] = verifier
}

// Verify executa a verificação cruzada entre credenciais IAM e dados financeiros
func (v *CredentialFinancialVerifier) Verify(ctx context.Context, req *CredentialFinancialVerificationRequest) (*CredentialFinancialVerificationResponse, error) {
	ctx, span := v.tracer.StartSpan(ctx, "CredentialFinancialVerifier.Verify")
	defer span.End()
	
	start := time.Now()
	
	// Adiciona timestamp se não estiver presente
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}

	// Registra início da verificação
	v.logger.InfoWithContext(ctx, "Iniciando verificação cruzada",
		"request_id", req.RequestID,
		"user_id", req.UserID,
		"tenant_id", req.TenantID,
		"region", req.RegionCode,
	)

	// Registra métricas de solicitação
	v.metricsRecorder.CounterInc("credential_financial_verification_requests_total", map[string]string{
		"tenant_id": req.TenantID,
		"region":    req.RegionCode,
	})

	// Verificando cache
	if v.config.EnableCaching {
		if cachedResp := v.getFromCache(req.RequestID); cachedResp != nil {
			v.logger.InfoWithContext(ctx, "Resposta recuperada do cache", "request_id", req.RequestID)
			return cachedResp, nil
		}
	}

	// Prepara ID da verificação
	verificationID := fmt.Sprintf("cfv-%d", time.Now().UnixNano())

	// Configura resposta inicial
	response := &CredentialFinancialVerificationResponse{
		RequestID:         req.RequestID,
		VerificationID:    verificationID,
		Status:            VerificationStatusPending,
		VerificationResults: make(map[string]VerificationResult),
		ComplianceResults: make(map[string]bool),
		Timestamp:         time.Now(),
	}

	// Recupera configurações regionais
	regionalSettings, ok := v.config.RegionalSettings[req.RegionCode]
	if !ok {
		v.logger.WarnWithContext(ctx, "Configurações regionais não encontradas, usando padrões",
			"region", req.RegionCode)
	}

	// Executar verificadores de categoria em paralelo
	var wg sync.WaitGroup
	var mu sync.Mutex
	anomalies := make([]Anomaly, 0)
	totalScore := 0
	totalWeight := 0

	v.verifiersMu.RLock()
	for _, verifier := range v.verifiers {
		wg.Add(1)
		go func(ver CategoryVerifier) {
			defer wg.Done()
			
			subCtx, cancel := context.WithTimeout(ctx, v.config.VerificationTimeout)
			defer cancel()
			
			result, err := ver.Verify(subCtx, req)
			if err != nil {
				v.logger.ErrorWithContext(ctx, "Erro durante verificação de categoria",
					"error", err.Error(),
					"category", ver.GetCategory(),
					"request_id", req.RequestID)
				
				mu.Lock()
				response.VerificationResults[ver.GetCategory()] = VerificationResult{
					Category:    ver.GetCategory(),
					Status:      VerificationStatusError,
					Score:       0,
					Description: fmt.Sprintf("Erro durante verificação: %s", err.Error()),
				}
				mu.Unlock()
				return
			}
			
			mu.Lock()
			response.VerificationResults[ver.GetCategory()] = *result
			
			// Acumular pontuação ponderada
			weight := ver.GetWeight()
			totalScore += result.Score * weight
			totalWeight += weight
			
			// Coletar anomalias se houver falhas
			if result.Status == VerificationStatusFailed && len(result.FailedFields) > 0 {
				anomalies = append(anomalies, Anomaly{
					AnomalyType:     fmt.Sprintf("%s_mismatch", ver.GetCategory()),
					Severity:        getSeverityByScore(result.Score),
					Description:     fmt.Sprintf("Inconsistências detectadas na categoria %s", ver.GetCategory()),
					DetectionMethod: "cross_verification",
					AffectedFields:  result.FailedFields,
					ConfidenceScore: float64(100 - result.Score) / 100.0,
				})
			}
			mu.Unlock()
		}(verifier)
	}
	v.verifiersMu.RUnlock()
	
	wg.Wait()
	
	// Calcular pontuação final
	trustScore := 0
	if totalWeight > 0 {
		trustScore = totalScore / totalWeight
	}
	
	// Determinar nível de confiança e status
	trustLevel := getTrustLevel(trustScore)
	verificationStatus := getVerificationStatus(trustScore, v.config.MinRequiredScore)
	
	// Atualiza resposta com resultados finais
	response.Status = verificationStatus
	response.TrustScore = trustScore
	response.TrustLevel = trustLevel
	response.DetectedAnomalies = anomalies
	response.ProcessingTimeMs = time.Since(start).Milliseconds()
	
	// Determina ação recomendada
	response.RecommendedAction = determineRecommendedAction(trustScore, len(anomalies))
	
	// Prepara informações de auditoria
	response.AuditInfo = map[string]interface{}{
		"verification_timestamp": time.Now().Format(time.RFC3339),
		"verification_version":   "1.0",
		"total_categories":       len(v.verifiers),
		"anomaly_count":          len(anomalies),
		"processing_time_ms":     response.ProcessingTimeMs,
	}
	
	// Adicionar ao cache se habilitado
	if v.config.EnableCaching {
		v.addToCache(req.RequestID, response)
	}
	
	// Registra métricas de resultado
	v.metricsRecorder.HistogramObserve("credential_financial_verification_duration_ms", float64(response.ProcessingTimeMs), map[string]string{
		"region": req.RegionCode,
		"status": verificationStatus,
	})
	
	v.metricsRecorder.HistogramObserve("credential_financial_verification_trust_score", float64(trustScore), map[string]string{
		"region": req.RegionCode,
	})
	
	v.logger.InfoWithContext(ctx, "Verificação cruzada concluída",
		"request_id", req.RequestID,
		"verification_id", verificationID,
		"status", verificationStatus,
		"trust_score", trustScore,
		"trust_level", trustLevel,
		"anomaly_count", len(anomalies),
		"processing_time_ms", response.ProcessingTimeMs,
	)
	
	return response, nil
}

// Recupera uma resposta do cache
func (v *CredentialFinancialVerifier) getFromCache(requestID string) *CredentialFinancialVerificationResponse {
	v.cacheMu.RLock()
	defer v.cacheMu.RUnlock()
	
	if resp, ok := v.cache[requestID]; ok {
		return resp
	}
	return nil
}

// Adiciona uma resposta ao cache
func (v *CredentialFinancialVerifier) addToCache(requestID string, response *CredentialFinancialVerificationResponse) {
	v.cacheMu.Lock()
	defer v.cacheMu.Unlock()
	
	v.cache[requestID] = response
	
	// Agendar remoção após expirar TTL
	go func() {
		time.Sleep(v.config.CacheTTL)
		v.cacheMu.Lock()
		delete(v.cache, requestID)
		v.cacheMu.Unlock()
	}()
}

// Determina o nível de confiança com base na pontuação
func getTrustLevel(score int) string {
	switch {
	case score >= 90:
		return TrustLevelVeryHigh
	case score >= 70:
		return TrustLevelHigh
	case score >= 50:
		return TrustLevelMedium
	case score >= 30:
		return TrustLevelLow
	default:
		return TrustLevelVeryLow
	}
}

// Determina o status da verificação com base na pontuação
func getVerificationStatus(score, minRequiredScore int) string {
	if score >= minRequiredScore {
		return VerificationStatusPassed
	}
	if score >= minRequiredScore/2 {
		return VerificationStatusPartial
	}
	return VerificationStatusFailed
}

// Determina a severidade com base na pontuação
func getSeverityByScore(score int) string {
	switch {
	case score < 30:
		return "high"
	case score < 60:
		return "medium"
	default:
		return "low"
	}
}

// Determina a ação recomendada com base na pontuação e anomalias
func determineRecommendedAction(trustScore int, anomalyCount int) string {
	if trustScore >= 80 && anomalyCount == 0 {
		return "approve"
	} else if trustScore >= 60 && anomalyCount <= 2 {
		return "review"
	} else if trustScore >= 40 {
		return "additional_verification"
	} else {
		return "reject"
	}
}

// GetVerificationStatus obtém o status de uma verificação pelo ID
func (v *CredentialFinancialVerifier) GetVerificationStatus(ctx context.Context, requestID string) (*CredentialFinancialVerificationResponse, error) {
	// Verificando cache
	if v.config.EnableCaching {
		if cachedResp := v.getFromCache(requestID); cachedResp != nil {
			return cachedResp, nil
		}
	}
	
	return nil, errors.New("verificação não encontrada")
}

// Close libera recursos utilizados pelo verificador
func (v *CredentialFinancialVerifier) Close() error {
	v.logger.Info("Finalizando verificador credencial-financeiro")
	return nil
}