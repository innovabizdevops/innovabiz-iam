/**
 * @file logging.go
 * @description Middleware para logging de requisições HTTP
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package middleware

import (
	"log"
	"net/http"
	"time"
)

// ResponseRecorder é um wrapper para http.ResponseWriter que captura o status code
type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

// WriteHeader sobrescreve o método original para capturar o status code
func (rr *ResponseRecorder) WriteHeader(statusCode int) {
	rr.StatusCode = statusCode
	rr.ResponseWriter.WriteHeader(statusCode)
}

// Write sobrescreve o método original para capturar o corpo da resposta
func (rr *ResponseRecorder) Write(b []byte) (int, error) {
	rr.Body = append(rr.Body, b...)
	return rr.ResponseWriter.Write(b)
}

// RequestLogger é um middleware para logging de requisições HTTP
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Registrar início da requisição
		startTime := time.Now()

		// Preparar recorder para capturar status code e corpo da resposta
		recorder := &ResponseRecorder{
			ResponseWriter: w,
			StatusCode:     http.StatusOK, // Valor padrão
		}

		// Processar a requisição
		next.ServeHTTP(recorder, r)

		// Calcular duração da requisição
		duration := time.Since(startTime)

		// Verificar se é uma requisição bem-sucedida (status code 2xx)
		success := recorder.StatusCode >= 200 && recorder.StatusCode < 300

		// Extrair informações relevantes
		method := r.Method
		path := r.URL.Path
		statusCode := recorder.StatusCode
		durationMs := float64(duration) / float64(time.Millisecond)
		userID := r.Header.Get("X-User-ID")
		tenantID := r.Header.Get("X-Tenant-ID")

		// Limitar tamanho do corpo para logging
		bodySize := len(recorder.Body)
		
		// Log completo com todas as informações relevantes
		log.Printf(
			"[REQUEST] Method: %s | Path: %s | Status: %d | Duration: %.2fms | Success: %v | UserID: %s | TenantID: %s | BodySize: %d bytes",
			method, path, statusCode, durationMs, success, userID, tenantID, bodySize,
		)

		// Enviar dados para sistema de logging centralizado
		// Em um sistema real, isso enviaria para um sistema como o ELK Stack ou similar
		// Por simplicidade, apenas mostramos o log no stdout
	})
}