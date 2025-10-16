/**
 * @file metrics_hook.go
 * @description Hook para coleta de métricas do Bureau de Crédito
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package hooks

import (
	"context"
	"time"
	
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

// MetricsHook implementa a coleta de métricas para operações do Bureau de Crédito
type MetricsHook struct {
	BaseHook
	registry              *prometheus.Registry
	requestCounter        *prometheus.CounterVec
	requestDurationHist   *prometheus.HistogramVec
	requestErrorCounter   *prometheus.CounterVec
	providerRequestHist   *prometheus.HistogramVec
	fraudDetectionCounter *prometheus.CounterVec
	scoreBuckets          *prometheus.HistogramVec
}

// NewMetricsHook cria uma nova instância do hook de métricas
func NewMetricsHook(registry *prometheus.Registry) *MetricsHook {
	hook := &MetricsHook{
		BaseHook: BaseHook{
			Name:     "metrics_hook",
			Priority: 10,
		},
		registry: registry,
	}
	
	// Inicializar métricas
	hook.initMetrics()
	
	return hook
}

// initMetrics inicializa as métricas Prometheus
func (h *MetricsHook) initMetrics() {
	// Contador de requisições
	h.requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bureau_credito_requests_total",
			Help: "Número total de requisições para o Bureau de Crédito",
		},
		[]string{"operation", "provider", "tenant", "status"},
	)
	h.registry.MustRegister(h.requestCounter)
	
	// Histograma de duração das requisições
	h.requestDurationHist = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bureau_credito_request_duration_seconds",
			Help:    "Duração das requisições para o Bureau de Crédito",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 10), // 10ms a ~10s
		},
		[]string{"operation", "provider", "tenant"},
	)
	h.registry.MustRegister(h.requestDurationHist)
	
	// Contador de erros
	h.requestErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bureau_credito_errors_total",
			Help: "Número total de erros em requisições para o Bureau de Crédito",
		},
		[]string{"operation", "provider", "tenant", "error_type"},
	)
	h.registry.MustRegister(h.requestErrorCounter)
	
	// Histograma de requisições por provedor
	h.providerRequestHist = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bureau_credito_provider_request_duration_seconds",
			Help:    "Duração das requisições para provedores externos de crédito",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 8), // 100ms a ~25s
		},
		[]string{"provider", "operation", "tenant"},
	)
	h.registry.MustRegister(h.providerRequestHist)
	
	// Contador de detecção de fraudes
	h.fraudDetectionCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bureau_credito_fraud_detection_total",
			Help: "Número total de detecções de fraude",
		},
		[]string{"provider", "tenant", "fraud_type", "severity"},
	)
	h.registry.MustRegister(h.fraudDetectionCounter)
	
	// Histograma para scores de crédito
	h.scoreBuckets = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bureau_credito_score_distribution",
			Help:    "Distribuição de scores de crédito",
			Buckets: prometheus.LinearBuckets(0, 100, 10), // 0-1000 em intervalos de 100
		},
		[]string{"provider", "tenant", "score_type"},
	)
	h.registry.MustRegister(h.scoreBuckets)
}

// Execute implementa a interface Hook para coleta de métricas
func (h *MetricsHook) Execute(ctx context.Context, hookType HookType, metadata HookMetadata, payload interface{}) error {
	// Processar métricas com base no tipo de hook e operação
	switch hookType {
	case HookBefore:
		// Incrementar contador de requisições
		h.requestCounter.WithLabelValues(
			string(metadata.OperationType),
			metadata.ProviderID,
			metadata.TenantID,
			"started",
		).Inc()
		
	case HookAfter:
		// Registrar duração da operação
		h.requestDurationHist.WithLabelValues(
			string(metadata.OperationType),
			metadata.ProviderID,
			metadata.TenantID,
		).Observe(metadata.Duration.Seconds())
		
		// Incrementar contador de requisições concluídas
		h.requestCounter.WithLabelValues(
			string(metadata.OperationType),
			metadata.ProviderID,
			metadata.TenantID,
			"completed",
		).Inc()
		
		// Processar métricas específicas por tipo de operação
		h.processOperationSpecificMetrics(metadata, payload)
		
	case HookError:
		// Incrementar contador de erros
		errorType := "unknown"
		if err, ok := payload.(error); ok {
			errorType = classifyError(err)
		}
		
		h.requestErrorCounter.WithLabelValues(
			string(metadata.OperationType),
			metadata.ProviderID,
			metadata.TenantID,
			errorType,
		).Inc()
	}
	
	return nil
}

// classifyError classifica um erro para métricas
func classifyError(err error) string {
	// Implementar classificação de erros
	// Por enquanto, apenas um stub simples
	return "generic_error"
}

// processOperationSpecificMetrics processa métricas específicas por tipo de operação
func (h *MetricsHook) processOperationSpecificMetrics(metadata HookMetadata, payload interface{}) {
	// Implementação específica por tipo de operação
	switch metadata.OperationType {
	case OpFraudDetection:
		// Processar métricas específicas de detecção de fraude
		if fraudResult, ok := extractFraudDetectionResult(payload); ok {
			h.fraudDetectionCounter.WithLabelValues(
				metadata.ProviderID,
				metadata.TenantID,
				fraudResult.FraudType,
				fraudResult.Severity,
			).Inc()
		}
		
	case OpCreditScore:
		// Processar métricas de score de crédito
		if scoreResult, ok := extractCreditScoreResult(payload); ok {
			h.scoreBuckets.WithLabelValues(
				metadata.ProviderID,
				metadata.TenantID,
				scoreResult.ScoreType,
			).Observe(float64(scoreResult.Score))
		}
	}
}

// Estruturas auxiliares para extração de dados das payloads

type fraudDetectionResult struct {
	FraudType string
	Severity  string
	Score     float64
}

type creditScoreResult struct {
	Score     int
	ScoreType string
}

// extractFraudDetectionResult extrai dados de detecção de fraude do payload
func extractFraudDetectionResult(payload interface{}) (fraudDetectionResult, bool) {
	// Implementação real depende da estrutura do payload
	// Este é apenas um stub
	
	// Exemplo de log para debug
	log.Debug().Interface("payload", payload).Msg("Extraindo resultado de detecção de fraude")
	
	// Retornar resultado fictício para demonstração
	return fraudDetectionResult{
		FraudType: "identity_theft",
		Severity:  "high",
		Score:     0.95,
	}, true
}

// extractCreditScoreResult extrai dados de score de crédito do payload
func extractCreditScoreResult(payload interface{}) (creditScoreResult, bool) {
	// Implementação real depende da estrutura do payload
	// Este é apenas um stub
	
	// Exemplo de log para debug
	log.Debug().Interface("payload", payload).Msg("Extraindo resultado de score de crédito")
	
	// Retornar resultado fictício para demonstração
	return creditScoreResult{
		Score:     750,
		ScoreType: "fico",
	}, true
}