/**
 * @file server.go
 * @description Servidor API REST/GraphQL para o Bureau de Crédito
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"innovabiz/iam/src/bureau-credito/api/graphql"
	"innovabiz/iam/src/bureau-credito/orchestration"
)

// APIServer representa o servidor API do Bureau de Crédito
type APIServer struct {
	httpServer   *http.Server
	router       *mux.Router
	orchestrator *orchestration.BureauOrchestrator
	config       *Config
}

// Config representa a configuração do servidor API
type Config struct {
	Port            int    `json:"port"`
	ReadTimeout     int    `json:"readTimeoutSecs"`
	WriteTimeout    int    `json:"writeTimeoutSecs"`
	ShutdownTimeout int    `json:"shutdownTimeoutSecs"`
	LogLevel        string `json:"logLevel"`
	EnableGraphQL   bool   `json:"enableGraphQL"`
	EnableCORS      bool   `json:"enableCORS"`
	EnableMetrics   bool   `json:"enableMetrics"`
}

// DefaultConfig retorna uma configuração padrão
func DefaultConfig() *Config {
	return &Config{
		Port:            8080,
		ReadTimeout:     30,
		WriteTimeout:    30,
		ShutdownTimeout: 15,
		LogLevel:        "info",
		EnableGraphQL:   true,
		EnableCORS:      true,
		EnableMetrics:   true,
	}
}

// NewAPIServer cria um novo servidor API
func NewAPIServer(orchestrator *orchestration.BureauOrchestrator, config *Config) *APIServer {
	// Usar configuração padrão se não fornecida
	if config == nil {
		config = DefaultConfig()
	}

	// Configurar logger
	setLogLevel(config.LogLevel)

	// Criar roteador
	router := mux.NewRouter()

	// Configurar servidor HTTP
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		ReadTimeout:  time.Duration(config.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.WriteTimeout) * time.Second,
		Handler:      router,
	}

	return &APIServer{
		httpServer:   httpServer,
		router:       router,
		orchestrator: orchestrator,
		config:       config,
	}
}

// SetupRoutes configura as rotas do servidor
func (s *APIServer) SetupRoutes() {
	// Criar controlador REST
	controller := NewBureauController(s.orchestrator)

	// Configurar middlewares
	s.router.Use(LoggingMiddleware)
	s.router.Use(JSONContentTypeMiddleware)
	
	if s.config.EnableMetrics {
		s.router.Use(MetricsMiddleware)
	}

	// Configurar CORS se habilitado
	if s.config.EnableCORS {
		corsMiddleware := cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Tenant-ID", "X-Correlation-ID"},
			AllowCredentials: true,
			MaxAge:           86400, // 24 horas
		})
		s.router.Use(corsMiddleware.Handler)
	}

	// Registrar rotas REST
	apiRouter := s.router.PathPrefix("/api/v1").Subrouter()
	controller.RegisterRoutes(apiRouter)

	// Configurar GraphQL se habilitado
	if s.config.EnableGraphQL {
		s.setupGraphQL()
	}

	// Rota de health check raiz
	s.router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"status":    "UP",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}).Methods("GET")

	log.Info().Int("port", s.config.Port).Msg("Rotas configuradas com sucesso")
}

// setupGraphQL configura o servidor GraphQL
func (s *APIServer) setupGraphQL() {
	// Criar resolvers GraphQL
	resolvers := graphql.NewResolvers(s.orchestrator)

	// Configurar schema GraphQL (simplificado)
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"assessment": &graphql.Field{
				Type: AssessmentResponseType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, ok := p.Args["id"].(string)
					if !ok {
						return nil, fmt.Errorf("ID inválido")
					}
					return resolvers.GetAssessment(p.Context, id)
				},
			},
			"assessmentStatus": &graphql.Field{
				Type: AssessmentStatusType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, ok := p.Args["id"].(string)
					if !ok {
						return nil, fmt.Errorf("ID inválido")
					}
					return resolvers.GetAssessmentStatus(p.Context, id)
				},
			},
			"health": &graphql.Field{
				Type: HealthStatusType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return resolvers.GetHealth(p.Context)
				},
			},
			"serviceInfo": &graphql.Field{
				Type: ServiceInfoType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return resolvers.GetServiceInfo(p.Context)
				},
			},
		},
	})

	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootMutation",
		Fields: graphql.Fields{
			"requestAssessment": &graphql.Field{
				Type: AssessmentResponseType,
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(AssessmentRequestInputType),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					input, ok := p.Args["input"].(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("Input inválido")
					}
					return resolvers.RequestAssessment(p.Context, input)
				},
			},
			"requestBatchAssessment": &graphql.Field{
				Type: BatchAssessmentResponseType,
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(BatchAssessmentRequestInputType),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					input, ok := p.Args["input"].(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("Input inválido")
					}
					return resolvers.RequestBatchAssessment(p.Context, input)
				},
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Erro ao criar schema GraphQL")
	}

	// Configurar handler GraphQL
	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	// Registrar endpoint GraphQL
	s.router.Handle("/graphql", h)
	log.Info().Msg("GraphQL configurado com sucesso em /graphql")
}

// Start inicia o servidor API
func (s *APIServer) Start() error {
	// Configurar manipulação de sinais para graceful shutdown
	idleConnsClosed := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		<-sigint

		log.Info().Msg("Recebido sinal para desligar servidor")

		// Definir timeout para shutdown
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			time.Duration(s.config.ShutdownTimeout)*time.Second,
		)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("Erro ao realizar shutdown do servidor HTTP")
		}

		close(idleConnsClosed)
	}()

	// Iniciar servidor
	log.Info().Int("port", s.config.Port).Msg("Iniciando servidor API do Bureau de Crédito")
	if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("erro ao iniciar servidor HTTP: %w", err)
	}

	<-idleConnsClosed
	log.Info().Msg("Servidor desligado com sucesso")
	return nil
}

// LoggingMiddleware registra informações das requisições
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Extrair informações adicionais dos cabeçalhos
		tenantID := r.Header.Get("X-Tenant-ID")
		correlationID := r.Header.Get("X-Correlation-ID")

		// Criar logger com contexto
		reqLogger := log.With().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote", r.RemoteAddr).
			Str("tenantId", tenantID).
			Str("correlationId", correlationID).
			Logger()

		reqLogger.Info().Msg("Requisição recebida")

		// Criar wrapper para o ResponseWriter para capturar o código de status
		ww := &responseWriterWrapper{w: w, status: http.StatusOK}

		// Chamar próximo handler
		next.ServeHTTP(ww, r)

		// Registrar informações da resposta
		duration := time.Since(start)
		reqLogger.Info().
			Int("status", ww.status).
			Dur("duration", duration).
			Msg("Requisição processada")
	})
}

// JSONContentTypeMiddleware define o cabeçalho Content-Type como application/json
func JSONContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Apenas para requisições não OPTIONS
		if r.Method != "OPTIONS" {
			w.Header().Set("Content-Type", "application/json")
		}
		next.ServeHTTP(w, r)
	})
}

// MetricsMiddleware coleta métricas das requisições
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Aqui seria implementada a coleta de métricas
		// Por exemplo, incrementando contadores de requisições, etc.
		next.ServeHTTP(w, r)
	})
}

// responseWriterWrapper é um wrapper para http.ResponseWriter que captura o código de status
type responseWriterWrapper struct {
	w      http.ResponseWriter
	status int
}

func (ww *responseWriterWrapper) Header() http.Header {
	return ww.w.Header()
}

func (ww *responseWriterWrapper) Write(data []byte) (int, error) {
	return ww.w.Write(data)
}

func (ww *responseWriterWrapper) WriteHeader(statusCode int) {
	ww.status = statusCode
	ww.w.WriteHeader(statusCode)
}

// setLogLevel configura o nível de log
func setLogLevel(level string) {
	// Configurar formato de logs para console
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	log.Logger = zerolog.New(output).With().Timestamp().Caller().Logger()

	// Configurar nível de log
	var logLevel zerolog.Level
	switch level {
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	default:
		logLevel = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(logLevel)
}

// Tipos GraphQL (simplificados - em uma implementação real, esses tipos estariam no pacote graphql)

// Declaração simplificada de tipos GraphQL
var (
	DateTimeScalar = graphql.NewScalar(graphql.ScalarConfig{
		Name:        "DateTime",
		Description: "DateTime scalar type",
		Serialize: func(value interface{}) interface{} {
			switch v := value.(type) {
			case time.Time:
				return v.Format(time.RFC3339)
			default:
				return nil
			}
		},
	})

	JSONScalar = graphql.NewScalar(graphql.ScalarConfig{
		Name:        "JSON",
		Description: "JSON scalar type",
		Serialize: func(value interface{}) interface{} {
			return value
		},
	})

	ErrorDetailsType = graphql.NewObject(graphql.ObjectConfig{
		Name: "ErrorDetails",
		Fields: graphql.Fields{
			"errorCode":      &graphql.Field{Type: graphql.String},
			"errorMessage":   &graphql.Field{Type: graphql.String},
			"failedServices": &graphql.Field{Type: graphql.NewList(graphql.String)},
			"partialResults": &graphql.Field{Type: graphql.Boolean},
			"errorSource":    &graphql.Field{Type: graphql.String},
		},
	})

	AssessmentResponseType = graphql.NewObject(graphql.ObjectConfig{
		Name: "AssessmentResponse",
		Fields: graphql.Fields{
			"responseId":      &graphql.Field{Type: graphql.ID},
			"requestId":       &graphql.Field{Type: graphql.ID},
			"correlationId":   &graphql.Field{Type: graphql.String},
			"userId":          &graphql.Field{Type: graphql.ID},
			"tenantId":        &graphql.Field{Type: graphql.ID},
			"status":          &graphql.Field{Type: graphql.String},
			"completedAt":     &graphql.Field{Type: DateTimeScalar},
			"processingTimeMs": &graphql.Field{Type: graphql.Int},
			"trustScore":      &graphql.Field{Type: graphql.Int},
			"riskLevel":       &graphql.Field{Type: graphql.String},
			"decision":        &graphql.Field{Type: graphql.String},
			"confidence":      &graphql.Field{Type: graphql.Float},
			"requiredActions": &graphql.Field{Type: graphql.NewList(graphql.String)},
			"suggestedActions": &graphql.Field{Type: graphql.NewList(graphql.String)},
			"warnings":        &graphql.Field{Type: graphql.NewList(graphql.String)},
			"errorDetails":    &graphql.Field{Type: ErrorDetailsType},
			"dataSources":     &graphql.Field{Type: graphql.NewList(graphql.String)},
			// Resultados detalhados (simplificados)
			"identityResults":  &graphql.Field{Type: JSONScalar},
			"creditResults":    &graphql.Field{Type: JSONScalar},
			"fraudResults":     &graphql.Field{Type: JSONScalar},
			"complianceResults": &graphql.Field{Type: JSONScalar},
			"riskResults":      &graphql.Field{Type: JSONScalar},
		},
	})

	BatchAssessmentResponseType = graphql.NewObject(graphql.ObjectConfig{
		Name: "BatchAssessmentResponse",
		Fields: graphql.Fields{
			"responses": &graphql.Field{Type: graphql.NewList(AssessmentResponseType)},
			"success":   &graphql.Field{Type: graphql.Int},
			"failed":    &graphql.Field{Type: graphql.Int},
			"total":     &graphql.Field{Type: graphql.Int},
		},
	})

	AssessmentStatusType = graphql.NewObject(graphql.ObjectConfig{
		Name: "AssessmentStatus",
		Fields: graphql.Fields{
			"requestId": &graphql.Field{Type: graphql.ID},
			"status":    &graphql.Field{Type: graphql.String},
		},
	})

	HealthStatusType = graphql.NewObject(graphql.ObjectConfig{
		Name: "HealthStatus",
		Fields: graphql.Fields{
			"status":    &graphql.Field{Type: graphql.String},
			"timestamp": &graphql.Field{Type: DateTimeScalar},
		},
	})

	ServiceInfoType = graphql.NewObject(graphql.ObjectConfig{
		Name: "ServiceInfo",
		Fields: graphql.Fields{
			"serviceName":    &graphql.Field{Type: graphql.String},
			"version":        &graphql.Field{Type: graphql.String},
			"buildTimestamp": &graphql.Field{Type: graphql.String},
			"features":       &graphql.Field{Type: graphql.NewList(graphql.String)},
		},
	})

	// Input types
	IdentityDataInputType = graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "IdentityDataInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"documentNumber":    &graphql.InputObjectFieldConfig{Type: graphql.String},
			"documentType":      &graphql.InputObjectFieldConfig{Type: graphql.String},
			"name":              &graphql.InputObjectFieldConfig{Type: graphql.String},
			"dateOfBirth":       &graphql.InputObjectFieldConfig{Type: graphql.String},
			"email":             &graphql.InputObjectFieldConfig{Type: graphql.String},
			"phoneNumber":       &graphql.InputObjectFieldConfig{Type: graphql.String},
			"address":           &graphql.InputObjectFieldConfig{Type: graphql.String},
			"nationality":       &graphql.InputObjectFieldConfig{Type: graphql.String},
			"biometricData":     &graphql.InputObjectFieldConfig{Type: JSONScalar},
			"verificationLevel": &graphql.InputObjectFieldConfig{Type: graphql.Int},
		},
	})

	AssessmentTypeEnum = graphql.NewEnum(graphql.EnumConfig{
		Name: "AssessmentType",
		Values: graphql.EnumValueConfigMap{
			"IDENTITY":      &graphql.EnumValueConfig{Value: "IDENTITY"},
			"CREDIT":        &graphql.EnumValueConfig{Value: "CREDIT"},
			"FRAUD":         &graphql.EnumValueConfig{Value: "FRAUD"},
			"COMPLIANCE":    &graphql.EnumValueConfig{Value: "COMPLIANCE"},
			"RISK":          &graphql.EnumValueConfig{Value: "RISK"},
			"COMPREHENSIVE": &graphql.EnumValueConfig{Value: "COMPREHENSIVE"},
		},
	})

	AssessmentRequestInputType = graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "AssessmentRequestInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"userId":            &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.ID)},
			"tenantId":          &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.ID)},
			"correlationId":     &graphql.InputObjectFieldConfig{Type: graphql.String},
			"assessmentTypes":   &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.NewList(AssessmentTypeEnum))},
			"creditProviders":   &graphql.InputObjectFieldConfig{Type: graphql.NewList(graphql.String)},
			"identityProviders": &graphql.InputObjectFieldConfig{Type: graphql.NewList(graphql.String)},
			"complianceRules":   &graphql.InputObjectFieldConfig{Type: graphql.NewList(graphql.String)},
			"identityData":      &graphql.InputObjectFieldConfig{Type: IdentityDataInputType},
			"timeoutMs":         &graphql.InputObjectFieldConfig{Type: graphql.Int},
			"forceRefresh":      &graphql.InputObjectFieldConfig{Type: graphql.Boolean},
			"requireAllResults": &graphql.InputObjectFieldConfig{Type: graphql.Boolean},
			"failFast":          &graphql.InputObjectFieldConfig{Type: graphql.Boolean},
			"customAttributes":  &graphql.InputObjectFieldConfig{Type: JSONScalar},
		},
	})

	BatchAssessmentRequestInputType = graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "BatchAssessmentRequestInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"requests": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.NewList(AssessmentRequestInputType)),
			},
		},
	})
)