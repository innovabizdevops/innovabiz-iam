package crossverification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/innovabizdevops/innovabiz-iam/constants"
	"github.com/innovabizdevops/innovabiz-iam/observability/adapter"
	"github.com/innovabizdevops/innovabiz-iam/observability/logging"
	"github.com/innovabizdevops/innovabiz-iam/observability/metrics"
	"github.com/innovabizdevops/innovabiz-iam/observability/tracing"
)

// OrchestrationConfig contém configurações para o orquestrador de verificações cruzadas
type OrchestrationConfig struct {
	DefaultTimeout      time.Duration            `json:"default_timeout"`
	VerifierTimeouts    map[string]time.Duration `json:"verifier_timeouts"`
	MinRequiredScore    int                      `json:"min_required_score"`
	EnableCaching       bool                     `json:"enable_caching"`
	CacheTTL            time.Duration            `json:"cache_ttl"`
	ParallelVerifiers   bool                     `json:"parallel_verifiers"`
	RegionalSettings    map[string]RegionalOrchestrationConfig `json:"regional_settings"`
	EnableFallbacks     bool                     `json:"enable_fallbacks"`
	EventBrokerURL      string                   `json:"event_broker_url"`
	AuditLogEnabled     bool                     `json:"audit_log_enabled"`
}

// RegionalOrchestrationConfig contém configurações regionais para orquestração
type RegionalOrchestrationConfig struct {
	MinRequiredScore    int                      `json:"min_required_score"`
	RequiredVerifiers   []string                 `json:"required_verifiers"`
	OptionalVerifiers   []string                 `json:"optional_verifiers"`
	VerifierWeights     map[string]int           `json:"verifier_weights"`
}

// CrossVerificationOrchestrator gerencia o fluxo de verificações cruzadas entre IAM e outros módulos
type CrossVerificationOrchestrator struct {
	config          OrchestrationConfig
	logger          logging.Logger
	tracer          tracing.Tracer
	metricsRecorder metrics.Metrics
	verifiersMu     sync.RWMutex
	verifiers       map[string]CategoryVerifier
	resultCache     sync.Map
}

// VerificationHistoryEntry representa um registro de histórico de verificação
type VerificationHistoryEntry struct {
	VerificationID       string                           `json:"verification_id"`
	RequestID            string                           `json:"request_id"`
	TenantID             string                           `json:"tenant_id"`
	UserID               string                           `json:"user_id"`
	Status               string                           `json:"status"`
	TrustScore           int                              `json:"trust_score"`
	VerifierResults      map[string]VerificationResult    `json:"verifier_results"`
	Timestamp            time.Time                        `json:"timestamp"`
	ProcessingTime       int64                            `json:"processing_time_ms"`
	DecisionContext      map[string]interface{}           `json:"decision_context,omitempty"`
}

// NewCrossVerificationOrchestrator cria uma nova instância do orquestrador
func NewCrossVerificationOrchestrator(config OrchestrationConfig) (*CrossVerificationOrchestrator, error) {
	// Inicializa observabilidade
	obsAdapter, err := adapter.NewAdapter(adapter.Config{
		ServiceName: "cross-verification-orchestrator",
		Environment: constants.Environment,
	})
	
	if err != nil {
		return nil, fmt.Errorf("falha ao inicializar observabilidade: %w", err)
	}

	orchestrator := &CrossVerificationOrchestrator{
		config:          config,
		logger:          obsAdapter.Logger(),
		tracer:          obsAdapter.Tracer(),
		metricsRecorder: obsAdapter.Metrics(),
		verifiers:       make(map[string]CategoryVerifier),
	}

	// Iniciar limpeza periódica de cache se habilitado
	if config.EnableCaching && config.CacheTTL > 0 {
		go orchestrator.startCacheCleanup()
	}

	orchestrator.logger.Info("CrossVerificationOrchestrator inicializado com sucesso")
	
	return orchestrator, nil
}

// RegisterVerifier registra um verificador de categoria
func (o *CrossVerificationOrchestrator) RegisterVerifier(verifier CategoryVerifier) {
	o.verifiersMu.Lock()
	defer o.verifiersMu.Unlock()
	o.verifiers[verifier.GetCategory()] = verifier
	o.logger.Info("Verificador registrado", "category", verifier.GetCategory())
}

// Verify executa o processo de verificação cruzada orquestrado
func (o *CrossVerificationOrchestrator) Verify(ctx context.Context, req *CredentialFinancialVerificationRequest) (*CredentialFinancialVerificationResponse, error) {
	ctx, span := o.tracer.StartSpan(ctx, "CrossVerificationOrchestrator.Verify")
	defer span.End()
	
	start := time.Now()
	
	// Adiciona timestamp se não estiver presente
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}
	
	// Registra início da verificação
	o.logger.InfoWithContext(ctx, "Iniciando orquestração de verificação cruzada",
		"request_id", req.RequestID,
		"user_id", req.UserID,
		"tenant_id", req.TenantID,
		"region", req.RegionCode,
	)
	
	// Registra métricas de solicitação
	o.metricsRecorder.CounterInc("cross_verification_requests_total", map[string]string{
		"tenant_id": req.TenantID,
		"region":    req.RegionCode,
	})
	
	// Verifica cache se habilitado
	if o.config.EnableCaching {
		if cachedResp, found := o.resultCache.Load(req.RequestID); found {
			o.logger.InfoWithContext(ctx, "Resposta recuperada do cache", "request_id", req.RequestID)
			return cachedResp.(*CredentialFinancialVerificationResponse), nil
		}
	}
	
	// Gera ID de verificação
	verificationID := fmt.Sprintf("orch-%s", uuid.New().String())
	
	// Recupera configurações regionais
	regionalConfig, ok := o.config.RegionalSettings[req.RegionCode]
	if !ok {
		o.logger.WarnWithContext(ctx, "Configurações regionais não encontradas, usando padrões",
			"region", req.RegionCode)
		
		// Configuração padrão
		regionalConfig = RegionalOrchestrationConfig{
			MinRequiredScore: o.config.MinRequiredScore,
		}
	}
	
	// Determina quais verificadores usar
	var verifiersToUse []CategoryVerifier
	verifierResults := make(map[string]VerificationResult)
	
	o.verifiersMu.RLock()
	
	// Primeiro adiciona verificadores obrigatórios
	for _, category := range regionalConfig.RequiredVerifiers {
		if verifier, exists := o.verifiers[category]; exists {
			verifiersToUse = append(verifiersToUse, verifier)
		} else {
			o.logger.WarnWithContext(ctx, "Verificador obrigatório não encontrado",
				"category", category, "request_id", req.RequestID)
		}
	}
	
	// Adiciona verificadores opcionais
	for _, category := range regionalConfig.OptionalVerifiers {
		if verifier, exists := o.verifiers[category]; exists {
			verifiersToUse = append(verifiersToUse, verifier)
		}
	}
	
	// Se nenhum verificador específico foi configurado, use todos disponíveis
	if len(verifiersToUse) == 0 {
		for _, verifier := range o.verifiers {
			verifiersToUse = append(verifiersToUse, verifier)
		}
	}
	
	o.verifiersMu.RUnlock()
	
	// Executa verificadores
	if o.config.ParallelVerifiers {
		verifierResults = o.executeVerifiersParallel(ctx, req, verifiersToUse)
	} else {
		verifierResults = o.executeVerifiersSequential(ctx, req, verifiersToUse)
	}
	
	// Calcula pontuação final
	trustScore, anomalies := o.calculateFinalScore(verifierResults, regionalConfig)
	
	// Determina status com base na pontuação
	status := getVerificationStatus(trustScore, regionalConfig.MinRequiredScore)
	
	// Prepara resposta
	response := &CredentialFinancialVerificationResponse{
		RequestID:           req.RequestID,
		VerificationID:      verificationID,
		Status:              status,
		TrustScore:          trustScore,
		TrustLevel:          getTrustLevel(trustScore),
		VerificationResults: verifierResults,
		DetectedAnomalies:   anomalies,
		ProcessingTimeMs:    time.Since(start).Milliseconds(),
		RecommendedAction:   determineRecommendedAction(trustScore, len(anomalies)),
		Timestamp:           time.Now(),
		AuditInfo: map[string]interface{}{
			"verification_timestamp": time.Now().Format(time.RFC3339),
			"orchestration_version": "1.0",
			"verifiers_executed":    len(verifierResults),
			"anomaly_count":         len(anomalies),
		},
	}
	
	// Registra histórico de verificação
	o.logVerificationHistory(ctx, req, response)
	
	// Adiciona ao cache se habilitado
	if o.config.EnableCaching {
		o.resultCache.Store(req.RequestID, response)
	}
	
	// Registra métricas de resultado
	o.metricsRecorder.HistogramObserve("cross_verification_duration_ms", float64(response.ProcessingTimeMs), map[string]string{
		"region": req.RegionCode,
		"status": status,
	})
	
	o.metricsRecorder.HistogramObserve("cross_verification_trust_score", float64(trustScore), map[string]string{
		"region": req.RegionCode,
	})
	
	o.logger.InfoWithContext(ctx, "Orquestração de verificação cruzada concluída",
		"request_id", req.RequestID,
		"verification_id", verificationID,
		"status", status,
		"trust_score", trustScore,
		"trust_level", getTrustLevel(trustScore),
		"anomaly_count", len(anomalies),
		"processing_time_ms", response.ProcessingTimeMs,
	)
	
	return response, nil
}

// Executa verificadores em paralelo
func (o *CrossVerificationOrchestrator) executeVerifiersParallel(ctx context.Context, req *CredentialFinancialVerificationRequest, verifiers []CategoryVerifier) map[string]VerificationResult {
	results := make(map[string]VerificationResult)
	resultsMu := sync.Mutex{}
	var wg sync.WaitGroup
	
	for _, verifier := range verifiers {
		wg.Add(1)
		go func(v CategoryVerifier) {
			defer wg.Done()
			
			category := v.GetCategory()
			timeout := o.getVerifierTimeout(category)
			
			verifierCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			
			start := time.Now()
			
			o.logger.DebugWithContext(ctx, "Executando verificador", 
				"category", category, 
				"timeout", timeout.String())
			
			result, err := v.Verify(verifierCtx, req)
			processingTime := time.Since(start).Milliseconds()
			
			if err != nil {
				o.logger.ErrorWithContext(ctx, "Erro ao executar verificador", 
					"category", category,
					"error", err.Error(),
					"processing_time_ms", processingTime)
				
				resultsMu.Lock()
				results[category] = VerificationResult{
					Category:    category,
					Status:      VerificationStatusError,
					Score:       0,
					Description: fmt.Sprintf("Erro durante verificação: %s", err.Error()),
				}
				resultsMu.Unlock()
				
				o.metricsRecorder.CounterInc("cross_verification_verifier_errors", map[string]string{
					"category": category,
				})
				
				return
			}
			
			o.metricsRecorder.HistogramObserve("cross_verification_verifier_duration_ms", 
				float64(processingTime), 
				map[string]string{
					"category": category,
					"status":   result.Status,
				})
			
			resultsMu.Lock()
			results[category] = *result
			resultsMu.Unlock()
		}(verifier)
	}
	
	wg.Wait()
	return results
}

// Executa verificadores sequencialmente
func (o *CrossVerificationOrchestrator) executeVerifiersSequential(ctx context.Context, req *CredentialFinancialVerificationRequest, verifiers []CategoryVerifier) map[string]VerificationResult {
	results := make(map[string]VerificationResult)
	
	for _, verifier := range verifiers {
		category := verifier.GetCategory()
		timeout := o.getVerifierTimeout(category)
		
		verifierCtx, cancel := context.WithTimeout(ctx, timeout)
		
		start := time.Now()
		
		o.logger.DebugWithContext(ctx, "Executando verificador", 
			"category", category, 
			"timeout", timeout.String())
		
		result, err := verifier.Verify(verifierCtx, req)
		processingTime := time.Since(start).Milliseconds()
		
		cancel()
		
		if err != nil {
			o.logger.ErrorWithContext(ctx, "Erro ao executar verificador", 
				"category", category,
				"error", err.Error(),
				"processing_time_ms", processingTime)
			
			results[category] = VerificationResult{
				Category:    category,
				Status:      VerificationStatusError,
				Score:       0,
				Description: fmt.Sprintf("Erro durante verificação: %s", err.Error()),
			}
			
			o.metricsRecorder.CounterInc("cross_verification_verifier_errors", map[string]string{
				"category": category,
			})
			
			continue
		}
		
		o.metricsRecorder.HistogramObserve("cross_verification_verifier_duration_ms", 
			float64(processingTime), 
			map[string]string{
				"category": category,
				"status":   result.Status,
			})
		
		results[category] = *result
	}
	
	return results
}

// Retorna o timeout configurado para um verificador específico
func (o *CrossVerificationOrchestrator) getVerifierTimeout(category string) time.Duration {
	if timeout, ok := o.config.VerifierTimeouts[category]; ok {
		return timeout
	}
	return o.config.DefaultTimeout
}

// Calcula pontuação final e deteta anomalias
func (o *CrossVerificationOrchestrator) calculateFinalScore(results map[string]VerificationResult, config RegionalOrchestrationConfig) (int, []Anomaly) {
	if len(results) == 0 {
		return 0, nil
	}
	
	totalWeight := 0
	weightedScore := 0
	anomalies := []Anomaly{}
	
	for category, result := range results {
		// Determina peso do verificador
		weight := 1
		if w, ok := config.VerifierWeights[category]; ok {
			weight = w
		} else if v, found := o.verifiers[category]; found {
			weight = v.GetWeight()
		}
		
		// Adiciona à pontuação ponderada
		weightedScore += result.Score * weight
		totalWeight += weight
		
		// Coleta anomalias
		if result.Status == VerificationStatusFailed || result.Status == VerificationStatusPartial {
			for _, field := range result.FailedFields {
				anomalies = append(anomalies, Anomaly{
					AnomalyType:     fmt.Sprintf("%s_anomaly", category),
					Severity:        getSeverityByScore(result.Score),
					Description:     fmt.Sprintf("Anomalia detectada em %s: %s", category, field),
					DetectionMethod: "cross_verification_orchestrator",
					AffectedFields:  []string{field},
					ConfidenceScore: float64(100 - result.Score) / 100.0,
				})
			}
		}
	}
	
	// Calcular pontuação final
	finalScore := 0
	if totalWeight > 0 {
		finalScore = weightedScore / totalWeight
	}
	
	return finalScore, anomalies
}

// Registra histórico de verificação
func (o *CrossVerificationOrchestrator) logVerificationHistory(ctx context.Context, req *CredentialFinancialVerificationRequest, resp *CredentialFinancialVerificationResponse) {
	if !o.config.AuditLogEnabled {
		return
	}
	
	history := VerificationHistoryEntry{
		VerificationID:  resp.VerificationID,
		RequestID:       req.RequestID,
		TenantID:        req.TenantID,
		UserID:          req.UserID,
		Status:          resp.Status,
		TrustScore:      resp.TrustScore,
		VerifierResults: resp.VerificationResults,
		Timestamp:       time.Now(),
		ProcessingTime:  resp.ProcessingTimeMs,
		DecisionContext: map[string]interface{}{
			"region_code":         req.RegionCode,
			"financial_products":  req.FinancialProducts,
			"verification_level":  req.VerificationLevel,
			"transaction_id":      req.TransactionID,
			"detected_anomalies":  len(resp.DetectedAnomalies),
			"recommended_action":  resp.RecommendedAction,
		},
	}
	
	// Aqui seria implementado o armazenamento do histórico em um sistema de audit log
	// Por exemplo, enviando para um sistema de eventos ou gravando em um banco de dados
	
	logEntry, _ := json.Marshal(history)
	o.logger.InfoWithContext(ctx, "Registro de verificação adicionado ao histórico",
		"verification_id", history.VerificationID,
		"user_id", history.UserID,
		"status", history.Status,
		"trust_score", history.TrustScore,
		"log_entry", string(logEntry),
	)
}

// Inicia limpeza periódica do cache
func (o *CrossVerificationOrchestrator) startCacheCleanup() {
	ticker := time.NewTicker(o.config.CacheTTL / 2)
	defer ticker.Stop()
	
	for range ticker.C {
		now := time.Now()
		expiredCount := 0
		
		o.resultCache.Range(func(key, value interface{}) bool {
			resp, ok := value.(*CredentialFinancialVerificationResponse)
			if !ok {
				o.resultCache.Delete(key)
				expiredCount++
				return true
			}
			
			// Verifica se o resultado está no cache há mais tempo que o TTL
			if now.Sub(resp.Timestamp) > o.config.CacheTTL {
				o.resultCache.Delete(key)
				expiredCount++
			}
			
			return true
		})
		
		if expiredCount > 0 {
			o.logger.Debug("Limpeza de cache concluída", "expired_items", expiredCount)
		}
	}
}

// GetVerificationHistory retorna o histórico de verificações para um usuário
func (o *CrossVerificationOrchestrator) GetVerificationHistory(ctx context.Context, userID, tenantID string, limit int) ([]VerificationHistoryEntry, error) {
	// Implementação real buscaria do armazenamento de histórico
	// Este é um placeholder
	o.logger.InfoWithContext(ctx, "Consultando histórico de verificações",
		"user_id", userID,
		"tenant_id", tenantID,
		"limit", limit)
		
	return nil, errors.New("método não implementado")
}