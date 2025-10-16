package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// RecoveryMiddleware captura pânicos na aplicação e registra-os adequadamente,
// retornando uma resposta de erro 500 em vez de encerrar o servidor.
// Implementado de acordo com as melhores práticas de segurança ISO/IEC 27001 e OWASP.
func RecoveryMiddleware(logger zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Obter o rastreamento de pilha
					stack := debug.Stack()
					
					// Configurar o contexto de logging
					errorLogger := logger.With().
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Str("remote_addr", r.RemoteAddr).
						Str("panic", fmt.Sprintf("%v", err)).
						Logger()

					// Registrar o erro com o rastreamento de pilha
					errorLogger.Error().
						RawJSON("stack_trace", []byte(fmt.Sprintf("%q", stack))).
						Msg("Recuperado de pânico durante o processamento da requisição")

					// Registrar o evento no sistema de tracing
					ctx := r.Context()
					span := otel.GetTracerProvider().Tracer("innovabiz.iam.middleware").
						Start(ctx, "panic.recovery")
					span.SetStatus(codes.Error, fmt.Sprintf("%v", err))
					span.SetAttributes(
						attribute.String("error.type", "panic"),
						attribute.String("error.message", fmt.Sprintf("%v", err)),
						attribute.String("error.stack", string(stack)),
					)
					span.End()

					// Retornar resposta de erro 500 ao cliente
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					
					type errorResponse struct {
						Status  int    `json:"status"`
						Code    string `json:"code"`
						Message string `json:"message"`
					}
					
					json.NewEncoder(w).Encode(errorResponse{
						Status:  http.StatusInternalServerError,
						Code:    "internal_server_error",
						Message: "Ocorreu um erro interno no servidor. Por favor, tente novamente mais tarde.",
					})
				}
			}()
			
			// Prosseguir com a requisição
			next.ServeHTTP(w, r)
		})
	}
}