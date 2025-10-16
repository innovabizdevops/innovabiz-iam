/**
 * @file metrics.go
 * @description Middleware para coleta de métricas de APIs REST
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package middleware

import (
	"net/http"
	"strconv"
	"time"
	"sync"
)

// Estruturas para armazenamento de métricas
type apiMetrics struct {
	sync.RWMutex
	requestCount      map[string]map[int]int // método+path -> statusCode -> contagem
	responseTime      map[string][]float64   // método+path -> lista de tempos de resposta (ms)
	requestSizeSum    map[string]int64       // método+path -> soma dos tamanhos das requisições
	responseSizeSum   map[string]int64       // método+path -> soma dos tamanhos das respostas
	errorCount        map[string]int         // método+path -> contagem de erros
	lastRequestTimes  map[string]time.Time   // método+path -> timestamp da última requisição
}

// Singleton para métricas
var (
	metrics *apiMetrics
	once    sync.Once
)

// getMetricsInstance retorna a instância singleton das métricas
func getMetricsInstance() *apiMetrics {
	once.Do(func() {
		metrics = &apiMetrics{
			requestCount:     make(map[string]map[int]int),
			responseTime:     make(map[string][]float64),
			requestSizeSum:   make(map[string]int64),
			responseSizeSum:  make(map[string]int64),
			errorCount:       make(map[string]int),
			lastRequestTimes: make(map[string]time.Time),
		}
	})
	return metrics
}

// MetricsCollector é um middleware para coletar métricas de requisições HTTP
func MetricsCollector(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Início da medição
		startTime := time.Now()
		
		// Capturar tamanho da requisição
		requestSize := r.ContentLength
		
		// Preparar recorder para capturar status e tamanho da resposta
		recorder := &ResponseRecorder{
			ResponseWriter: w,
			StatusCode:     http.StatusOK, // Valor padrão
			Body:           []byte{},
		}
		
		// Processar requisição
		next.ServeHTTP(recorder, r)
		
		// Fim da medição e coleta de métricas
		duration := time.Since(startTime)
		durationMs := float64(duration) / float64(time.Millisecond)
		
		// Tamanho da resposta
		responseSize := int64(len(recorder.Body))
		
		// Chave para indexação das métricas
		key := r.Method + ":" + r.URL.Path
		
		// Obter instância de métricas
		metrics := getMetricsInstance()
		
		// Lock para atualização segura
		metrics.Lock()
		defer metrics.Unlock()
		
		// Atualizar métricas
		
		// Contagem por status code
		if _, exists := metrics.requestCount[key]; !exists {
			metrics.requestCount[key] = make(map[int]int)
		}
		metrics.requestCount[key][recorder.StatusCode]++
		
		// Tempos de resposta
		metrics.responseTime[key] = append(metrics.responseTime[key], durationMs)
		
		// Tamanhos de requisição e resposta
		metrics.requestSizeSum[key] += requestSize
		metrics.responseSizeSum[key] += responseSize
		
		// Contagem de erros (status code >= 400)
		if recorder.StatusCode >= 400 {
			metrics.errorCount[key]++
		}
		
		// Atualizar timestamp da última requisição
		metrics.lastRequestTimes[key] = time.Now()
	})
}

// GetMetricsSummary retorna um resumo das métricas coletadas
// Esta função seria chamada por endpoints de monitoramento ou health check
func GetMetricsSummary() map[string]interface{} {
	metrics := getMetricsInstance()
	metrics.RLock()
	defer metrics.RUnlock()
	
	// Construir resumo de métricas
	summary := make(map[string]interface{})
	endpointMetrics := make(map[string]interface{})
	
	for key, statusCodes := range metrics.requestCount {
		// Calcular total de requisições para este endpoint
		totalRequests := 0
		for _, count := range statusCodes {
			totalRequests += count
		}
		
		// Calcular tempo médio de resposta
		avgResponseTime := 0.0
		if responseTimes, exists := metrics.responseTime[key]; exists && len(responseTimes) > 0 {
			sum := 0.0
			for _, t := range responseTimes {
				sum += t
			}
			avgResponseTime = sum / float64(len(responseTimes))
		}
		
		// Preparar métricas por status code
		statusCodeMetrics := make(map[string]int)
		for code, count := range statusCodes {
			statusCodeMetrics[strconv.Itoa(code)] = count
		}
		
		// Compilar métricas para este endpoint
		endpointMetrics[key] = map[string]interface{}{
			"totalRequests":    totalRequests,
			"avgResponseTime":  avgResponseTime,
			"errorCount":       metrics.errorCount[key],
			"byStatusCode":     statusCodeMetrics,
			"totalRequestSize": metrics.requestSizeSum[key],
			"totalResponseSize": metrics.responseSizeSum[key],
			"lastRequest":       metrics.lastRequestTimes[key],
		}
	}
	
	summary["endpoints"] = endpointMetrics
	summary["collectedSince"] = time.Now().Add(-24 * time.Hour) // Apenas um exemplo
	
	return summary
}