package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"innovabiz/iam/identity-service/internal/application"
	"innovabiz/iam/identity-service/internal/interface/api/handler"
	"innovabiz/iam/identity-service/internal/interface/middleware"
)

// Server representa o servidor HTTP da aplicação
type Server struct {
	router      *mux.Router
	httpServer  *http.Server
	logger      zerolog.Logger
	tracer      trace.Tracer
	roleService application.RoleService
	// Adicionar outros serviços conforme necessário
}

// Config representa a configuração do servidor HTTP
type Config struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	AllowedOrigins  []string
	AllowedMethods  []string
	AllowedHeaders  []string
}

// DefaultConfig retorna uma configuração padrão para o servidor
func DefaultConfig() Config {
	return Config{
		Port:            "8080",
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		ShutdownTimeout: 15 * time.Second,
		AllowedOrigins:  []string{"*"},
		AllowedMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:  []string{"Content-Type", "Authorization", "X-Tenant-ID", "X-User-ID"},
	}
}

// New cria uma nova instância do servidor HTTP
func New(config Config, roleService application.RoleService, logger zerolog.Logger) *Server {
	router := mux.NewRouter()
	tracer := otel.Tracer("innovabiz.iam.identity-service.http")

	// Configurar o middleware OpenTelemetry
	router.Use(otelmux.Middleware("innovabiz.iam.identity-service"))

	// Configurar middleware de logging
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			logger := logger.With().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Logger()

			logger.Info().Msg("Request iniciada")

			sw := &statusWriter{ResponseWriter: w}
			next.ServeHTTP(sw, r)

			latency := time.Since(start)
			logger.Info().
				Int("status", sw.status).
				Str("latency", latency.String()).
				Msg("Request finalizada")
		})
	})

	// Configurar CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   config.AllowedOrigins,
		AllowedMethods:   config.AllowedMethods,
		AllowedHeaders:   config.AllowedHeaders,
		AllowCredentials: true,
	})
	router.Use(handlers.CompressHandler)
	router.Use(corsHandler.Handler)

	// Configurar recuperação de pânico
	router.Use(middleware.RecoveryMiddleware(logger))

	// Configurar servidor HTTP
	httpServer := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		ErrorLog:     nil, // Usar zerolog ao invés do log padrão
	}

	return &Server{
		router:      router,
		httpServer:  httpServer,
		logger:      logger.With().Str("component", "Server").Logger(),
		tracer:      tracer,
		roleService: roleService,
	}
}

// Start inicia o servidor HTTP
func (s *Server) Start() error {
	s.registerRoutes()

	s.logger.Info().Msgf("Servidor iniciado na porta %s", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Error().Err(err).Msg("Erro ao iniciar o servidor")
		return err
	}

	return nil
}

// Shutdown encerra o servidor HTTP graciosamente
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info().Msg("Encerrando o servidor...")
	return s.httpServer.Shutdown(ctx)
}

// registerRoutes registra todas as rotas da API
func (s *Server) registerRoutes() {
	s.registerAPIRoutes()
	s.registerHealthCheckRoutes()
	s.registerDocsRoutes()
}

// registerAPIRoutes registra as rotas da API principal
func (s *Server) registerAPIRoutes() {
	api := s.router.PathPrefix("/api/v1").Subrouter()
	
	// Registrar middleware de autenticação (JWT) para as rotas da API
	// Em um ambiente de produção, descomente esta linha e implemente o middleware
	// api.Use(middleware.AuthenticationMiddleware())
	
	// Registrar handlers
	s.registerRoleHandler(api)
	
	// Registrar outros handlers conforme necessário
	// s.registerUserHandler(api)
	// s.registerPermissionHandler(api)
	// s.registerTenantHandler(api)
}

// registerRoleHandler registra as rotas do RoleHandler
func (s *Server) registerRoleHandler(router *mux.Router) {
	roleHandler := handler.NewRoleHandler(s.roleService, s.logger, s.tracer)
	roleHandler.RegisterRoutes(router)
}

// registerHealthCheckRoutes registra as rotas de health check
func (s *Server) registerHealthCheckRoutes() {
	s.router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"UP"}`)
	}).Methods(http.MethodGet)

	s.router.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		// Implementar verificações de prontidão, como conexão com banco de dados
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"READY"}`)
	}).Methods(http.MethodGet)
}

// registerDocsRoutes registra as rotas de documentação da API
func (s *Server) registerDocsRoutes() {
	// TODO: Implementar quando tivermos documentação OpenAPI
	// s.router.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs"))))
}

// statusWriter é um wrapper para http.ResponseWriter para capturar o código de status da resposta
type statusWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader implementa a interface http.ResponseWriter
func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// Write implementa a interface http.ResponseWriter
func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}