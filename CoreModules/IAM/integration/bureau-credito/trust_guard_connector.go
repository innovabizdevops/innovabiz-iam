package bureaucredito

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/innovabizdevops/innovabiz-iam/constants"
	"github.com/innovabizdevops/innovabiz-iam/observability/logging"
	"github.com/innovabizdevops/innovabiz-iam/observability/metrics"
	"github.com/innovabizdevops/innovabiz-iam/observability/tracing"
)

// TrustGuardConnector é o adaptador para integração com o serviço TrustGuard
// para verificação de identidade em Bureau de Créditos
type TrustGuardConnector struct {
	apiBaseURL      string
	apiKey          string
	httpClient      *http.Client
	logger          logging.Logger
	metrics         metrics.Metrics
	tracer          tracing.Tracer
	defaultTimeout  time.Duration
	retryAttempts   int
	retryDelay      time.Duration
}

// TrustGuardConfig é a configuração do conector TrustGuard
type TrustGuardConfig struct {
	APIURL         string
	APIKey         string
	Timeout        time.Duration
	RetryAttempts  int
	RetryDelay     time.Duration
	Logger         logging.Logger
	MetricsClient  metrics.Metrics
	TracingClient  tracing.Tracer
}

// VerificationRequest representa uma solicitação de verificação de identidade
type VerificationRequest struct {
	UserID            string                 `json:"userId"`
	VerificationType  string                 `json:"verificationType"` // DOCUMENT, BIOMETRIC, LIVENESS, etc.
	DocumentType      string                 `json:"documentType,omitempty"`
	DocumentData      map[string]interface{} `json:"documentData,omitempty"`
	BiometricData     map[string]interface{} `json:"biometricData,omitempty"`
	AdditionalContext map[string]interface{} `json:"additionalContext,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// VerificationResponse é a resposta de uma verificação de identidade
type VerificationResponse struct {
	VerificationID    string                 `json:"verificationId"`
	Status            string                 `json:"status"`
	Score             float64                `json:"score"`
	Confidence        float64                `json:"confidence"`
	Timestamp         time.Time              `json:"timestamp"`
	ExpiresAt         time.Time              `json:"expiresAt"`
	VerificationType  string                 `json:"verificationType"`
	Warnings          []string               `json:"warnings"`
	Details           map[string]interface{} `json:"details"`
	ComplianceStatus  map[string]interface{} `json:"complianceStatus"`
	RecommendedAction string                 `json:"recommendedAction"`
}

// NewTrustGuardConnector cria uma nova instância do conector TrustGuard
func NewTrustGuardConnector(cfg TrustGuardConfig) (*TrustGuardConnector, error) {
	if cfg.APIURL == "" {
		return nil, fmt.Errorf("TrustGuard API URL é obrigatória")
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("TrustGuard API Key é obrigatória")
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	retryAttempts := cfg.RetryAttempts
	if retryAttempts <= 0 {
		retryAttempts = 3
	}

	retryDelay := cfg.RetryDelay
	if retryDelay <= 0 {
		retryDelay = 500 * time.Millisecond
	}

	// Logger padrão se não for fornecido
	logger := cfg.Logger
	if logger == nil {
		// Utilizar logger padrão
		logger = logging.NewDefaultLogger()
	}

	// Metrics padrão se não for fornecido
	metricsClient := cfg.MetricsClient
	if metricsClient == nil {
		// Utilizar cliente de métricas padrão
		metricsClient = metrics.NewNoOpMetrics()
	}

	// Tracer padrão se não for fornecido
	tracingClient := cfg.TracingClient
	if tracingClient == nil {
		// Utilizar cliente de tracing padrão
		tracingClient = tracing.NewNoOpTracer()
	}

	return &TrustGuardConnector{
		apiBaseURL:     cfg.APIURL,
		apiKey:         cfg.APIKey,
		httpClient:     &http.Client{Timeout: timeout},
		logger:         logger,
		metrics:        metricsClient,
		tracer:         tracingClient,
		defaultTimeout: timeout,
		retryAttempts:  retryAttempts,
		retryDelay:     retryDelay,
	}, nil
}

// VerifyIdentity realiza uma verificação de identidade através do TrustGuard
func (c *TrustGuardConnector) VerifyIdentity(ctx context.Context, req *VerificationRequest) (*VerificationResponse, error) {
	// Criar contexto com tracing
	ctx, span := c.tracer.Start(ctx, "TrustGuardConnector.VerifyIdentity")
	defer span.End()

	// Adicionar atributos ao span
	span.SetAttributes(
		tracing.StringAttribute("verificationType", req.VerificationType),
		tracing.StringAttribute("userId", req.UserID),
	)

	// Registrar início da operação
	c.logger.Info(
		"Iniciando verificação de identidade com TrustGuard",
		logging.String("userId", req.UserID),
		logging.String("verificationType", req.VerificationType),
	)

	// Iniciar timer para métricas
	timer := c.metrics.NewTimer()
	defer func() {
		// Registrar duração da operação
		c.metrics.RecordTimer(
			"iam.trustguard.verification_duration",
			timer.Elapsed(),
			metrics.Tag{Key: "verification_type", Value: req.VerificationType},
		)
	}()

	// Preparar payload
	payload, err := json.Marshal(req)
	if err != nil {
		c.logger.Error(
			"Erro ao serializar requisição para TrustGuard",
			logging.String("error", err.Error()),
			logging.String("userId", req.UserID),
		)
		c.metrics.IncrementCounter("iam.trustguard.error", metrics.Tag{Key: "error_type", Value: "serialization_error"})
		return nil, fmt.Errorf("falha ao serializar requisição: %w", err)
	}

	// Executar requisição com retry
	var response *VerificationResponse
	var lastError error

	for attempt := 0; attempt <= c.retryAttempts; attempt++ {
		// Se não for a primeira tentativa, aguardar antes de retry
		if attempt > 0 {
			select {
			case <-time.After(c.retryDelay * time.Duration(attempt)):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		// Executar a verificação
		response, lastError = c.executeVerification(ctx, payload)
		if lastError == nil {
			// Sucesso, retornar resposta
			break
		}

		c.logger.Warn(
			"Tentativa de verificação falhou, executando retry",
			logging.String("error", lastError.Error()),
			logging.Int("attempt", attempt),
			logging.Int("maxAttempts", c.retryAttempts),
			logging.String("userId", req.UserID),
		)
	}

	// Se ainda tiver erro após todas as tentativas
	if lastError != nil {
		c.logger.Error(
			"Todas as tentativas de verificação falharam",
			logging.String("error", lastError.Error()),
			logging.String("userId", req.UserID),
			logging.String("verificationType", req.VerificationType),
		)
		c.metrics.IncrementCounter("iam.trustguard.error", metrics.Tag{Key: "error_type", Value: "max_retries_exceeded"})
		return nil, fmt.Errorf("verificação de identidade falhou após %d tentativas: %w", c.retryAttempts, lastError)
	}

	// Verificação bem-sucedida
	c.logger.Info(
		"Verificação de identidade concluída com sucesso",
		logging.String("userId", req.UserID),
		logging.String("verificationId", response.VerificationID),
		logging.String("status", response.Status),
		logging.Float64("score", response.Score),
	)

	// Registrar métrica de sucesso
	c.metrics.IncrementCounter("iam.trustguard.verification", 
		metrics.Tag{Key: "verification_type", Value: req.VerificationType},
		metrics.Tag{Key: "status", Value: response.Status},
	)

	return response, nil
}

// executeVerification executa uma requisição de verificação ao serviço TrustGuard
func (c *TrustGuardConnector) executeVerification(ctx context.Context, payload []byte) (*VerificationResponse, error) {
	// Criar URL para requisição
	url := fmt.Sprintf("%s/api/v1/verify", c.apiBaseURL)

	// Criar request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	// Configurar headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("X-Request-ID", constants.GetRequestIDFromContext(ctx))
	req.Header.Set("X-Client-ID", "INNOVABIZ-IAM")

	// Executar requisição
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro de comunicação com TrustGuard: %w", err)
	}
	defer resp.Body.Close()

	// Verificar status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TrustGuard retornou status code inválido: %d", resp.StatusCode)
	}

	// Processar resposta
	var response VerificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("erro ao processar resposta: %w", err)
	}

	return &response, nil
}

// ValidateDocumentIdentity valida a identidade com base em documentos
func (c *TrustGuardConnector) ValidateDocumentIdentity(ctx context.Context, userID string, documentType string, documentData map[string]interface{}) (*VerificationResponse, error) {
	req := &VerificationRequest{
		UserID:           userID,
		VerificationType: "DOCUMENT",
		DocumentType:     documentType,
		DocumentData:     documentData,
		Metadata: map[string]interface{}{
			"source":    "IAM",
			"service":   "BureauCredito",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	return c.VerifyIdentity(ctx, req)
}

// ValidateBiometricIdentity valida a identidade com base em biometria
func (c *TrustGuardConnector) ValidateBiometricIdentity(ctx context.Context, userID string, biometricData map[string]interface{}) (*VerificationResponse, error) {
	req := &VerificationRequest{
		UserID:           userID,
		VerificationType: "BIOMETRIC",
		BiometricData:    biometricData,
		Metadata: map[string]interface{}{
			"source":    "IAM",
			"service":   "BureauCredito",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	return c.VerifyIdentity(ctx, req)
}

// GetVerificationStatus obtém o status de uma verificação
func (c *TrustGuardConnector) GetVerificationStatus(ctx context.Context, verificationID string) (*VerificationResponse, error) {
	// Criar contexto com tracing
	ctx, span := c.tracer.Start(ctx, "TrustGuardConnector.GetVerificationStatus")
	defer span.End()

	// Adicionar atributos ao span
	span.SetAttributes(
		tracing.StringAttribute("verificationId", verificationID),
	)

	// Construir URL
	url := fmt.Sprintf("%s/api/v1/verification/%s", c.apiBaseURL, verificationID)

	// Criar request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	// Configurar headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("X-Request-ID", constants.GetRequestIDFromContext(ctx))

	// Executar requisição
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.metrics.IncrementCounter("iam.trustguard.error", metrics.Tag{Key: "error_type", Value: "communication_error"})
		return nil, fmt.Errorf("erro de comunicação com TrustGuard: %w", err)
	}
	defer resp.Body.Close()

	// Verificar status code
	if resp.StatusCode != http.StatusOK {
		c.metrics.IncrementCounter("iam.trustguard.error", metrics.Tag{Key: "error_type", Value: "invalid_status_code"})
		return nil, fmt.Errorf("TrustGuard retornou status code inválido: %d", resp.StatusCode)
	}

	// Processar resposta
	var response VerificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		c.metrics.IncrementCounter("iam.trustguard.error", metrics.Tag{Key: "error_type", Value: "response_decode_error"})
		return nil, fmt.Errorf("erro ao processar resposta: %w", err)
	}

	return &response, nil
}