/**
 * @file routes.go
 * @description Configuração de rotas REST para o serviço de identidade multi-contexto
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package rest

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"innovabiz/iam/src/multi-context/api/rest/middleware"
)

// SetupRoutes configura todas as rotas do serviço de identidade multi-contexto
func SetupRoutes(controller *MultiContextController) http.Handler {
	r := mux.NewRouter()
	
	// Middleware para capturar métricas e logs
	r.Use(middleware.RequestLogger)
	r.Use(middleware.MetricsCollector)
	
	// API versioning
	api := r.PathPrefix("/api/v1").Subrouter()
	
	// Middleware de autenticação
	api.Use(middleware.TokenValidator)
	
	// Rotas para contextos de identidade
	contexts := api.PathPrefix("/contexts").Subrouter()
	contexts.HandleFunc("", controller.ListContexts).Methods("GET")
	contexts.HandleFunc("", controller.CreateContext).Methods("POST") // To be implemented
	contexts.HandleFunc("/{id}", controller.GetContext).Methods("GET")
	contexts.HandleFunc("/{id}", controller.UpdateContext).Methods("PUT") // To be implemented
	contexts.HandleFunc("/{id}/verification-history", controller.GetContextVerificationHistory).Methods("GET")
	contexts.HandleFunc("/{id}/verification-level", controller.UpdateContextVerificationLevel).Methods("PATCH")
	contexts.HandleFunc("/{id}/trust-score", controller.UpdateContextTrustScore).Methods("PATCH")
	contexts.HandleFunc("/{id}/attributes", controller.GetContextAttributes).Methods("GET") // To be implemented
	
	// Rotas para atributos
	attributes := api.PathPrefix("/attributes").Subrouter()
	attributes.HandleFunc("", controller.ListAttributes).Methods("GET")
	attributes.HandleFunc("", controller.CreateAttribute).Methods("POST")
	attributes.HandleFunc("/search", controller.SearchAttributes).Methods("POST")
	attributes.HandleFunc("/{id}", controller.GetAttribute).Methods("GET")
	attributes.HandleFunc("/{id}", controller.UpdateAttribute).Methods("PUT")
	attributes.HandleFunc("/{id}", controller.DeleteAttribute).Methods("DELETE")
	attributes.HandleFunc("/{id}/verification", controller.VerifyAttribute).Methods("PATCH")
	attributes.HandleFunc("/{id}/verification-history", controller.GetAttributeVerificationHistory).Methods("GET")
	
	// Health check e monitoramento
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")
	
	// Configuração CORS
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-User-ID", "X-Tenant-ID", "X-User-Role", "X-User-Roles"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	
	return corsMiddleware.Handler(r)
}