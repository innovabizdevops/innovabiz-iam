package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"innovabiz/iam/identity-service/internal/application/impl"
	"innovabiz/iam/identity-service/internal/domain/events"
	"innovabiz/iam/identity-service/internal/infrastructure/persistence/postgres"
	"innovabiz/iam/identity-service/internal/interface/api/server"
)

const (
	serviceName    = "innovabiz.iam.identity-service"
	serviceVersion = "0.1.0"
)

func main() {
	// Carregar variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Aviso: Arquivo .env não encontrado ou não pode ser carregado: %v\n", err)
	}

	// Configurar logger
	setupLogger()
	log.Info().Msg("Iniciando serviço de identidade do INNOVABIZ IAM")

	// Configurar OpenTelemetry
	tp, err := setupTelemetry()
	if err != nil {
		log.Fatal().Err(err).Msg("Falha ao configurar telemetria")
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Error().Err(err).Msg("Falha ao encerrar o provedor de telemetria")
		}
	}()

	// Configurar banco de dados
	dbConfig := postgres.DefaultConfig()
	dbConfig.DSN = getEnv("DATABASE_URL", dbConfig.DSN)
	dbConfig.MaxOpenConns = getEnvInt("DB_MAX_OPEN_CONNS", dbConfig.MaxOpenConns)
	dbConfig.MaxIdleConns = getEnvInt("DB_MAX_IDLE_CONNS", dbConfig.MaxIdleConns)
	dbConfig.ConnMaxLifetime = getEnvDuration("DB_CONN_MAX_LIFETIME", dbConfig.ConnMaxLifetime)
	dbConfig.ConnMaxIdleTime = getEnvDuration("DB_CONN_MAX_IDLE_TIME", dbConfig.ConnMaxIdleTime)
	
	log.Info().Msg("Conectando ao banco de dados PostgreSQL")
	db, err := postgres.NewPostgresDB(dbConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Falha ao conectar ao banco de dados")
	}
	defer db.Close()

	// Configurar repositórios
	log.Info().Msg("Inicializando repositórios")
	roleRepo := postgres.NewRoleRepository(db, log.With().Str("component", "RoleRepository").Logger())
	// Configurar outros repositórios conforme necessário
	// userRepo := postgres.NewUserRepository(db, log.With().Str("component", "UserRepository").Logger())
	// permissionRepo := postgres.NewPermissionRepository(db, log.With().Str("component", "PermissionRepository").Logger())

	// Configurar barramento de eventos
	log.Info().Msg("Inicializando barramento de eventos")
	eventBus := events.NewInMemoryEventBus(log.With().Str("component", "EventBus").Logger())
	// Registrar listeners e publishers conforme necessário
	
	// Configurar serviços
	log.Info().Msg("Inicializando serviços de aplicação")
	serviceFactory := impl.NewServiceFactory(
		roleRepo,
		// userRepo,
		// permissionRepo,
		eventBus,
		log.With().Str("component", "ServiceFactory").Logger(),
	)
	
	roleService := serviceFactory.NewRoleService()

	// Configurar servidor HTTP
	log.Info().Msg("Inicializando servidor HTTP")
	serverConfig := server.DefaultConfig()
	serverConfig.Port = getEnv("HTTP_PORT", serverConfig.Port)
	serverConfig.ReadTimeout = getEnvDuration("HTTP_READ_TIMEOUT", serverConfig.ReadTimeout)
	serverConfig.WriteTimeout = getEnvDuration("HTTP_WRITE_TIMEOUT", serverConfig.WriteTimeout)
	serverConfig.ShutdownTimeout = getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", serverConfig.ShutdownTimeout)
	
	// Configurar CORS
	if corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", ""); corsOrigins != "" {
		serverConfig.AllowedOrigins = []string{corsOrigins}
	}
	
	httpServer := server.New(serverConfig, roleService, log.With().Str("component", "Server").Logger())

	// Iniciar servidor HTTP em uma goroutine
	go func() {
		log.Info().Msgf("Servidor HTTP iniciado na porta %s", serverConfig.Port)
		if err := httpServer.Start(); err != nil {
			log.Fatal().Err(err).Msg("Falha ao iniciar o servidor HTTP")
		}
	}()

	// Configurar canal para capturar sinais de interrupção
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	// Esperar por sinal para encerramento
	sig := <-quit
	log.Info().Msgf("Recebido sinal %s, iniciando encerramento gracioso", sig)

	// Encerrar servidor HTTP
	ctx, cancel := context.WithTimeout(context.Background(), serverConfig.ShutdownTimeout)
	defer cancel()

	log.Info().Msg("Encerrando servidor HTTP")
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Falha ao encerrar o servidor HTTP")
	}

	log.Info().Msg("Serviço de identidade do INNOVABIZ IAM encerrado com sucesso")
}

// setupLogger configura o logger global
func setupLogger() {
	logLevel := getEnv("LOG_LEVEL", "info")
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	
	zerolog.SetGlobalLevel(level)
	
	if getEnv("LOG_FORMAT", "json") == "console" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}
	
	// Configurar metadados globais
	log.Logger = log.With().
		Str("service", serviceName).
		Str("version", serviceVersion).
		Logger()
}

// setupTelemetry configura o OpenTelemetry
func setupTelemetry() (*trace.TracerProvider, error) {
	ctx := context.Background()
	
	// Verificar se a telemetria está habilitada
	if getEnv("OTEL_ENABLED", "false") != "true" {
		// Configurar um provedor sem operações se desabilitado
		tp := trace.NewTracerProvider()
		otel.SetTracerProvider(tp)
		return tp, nil
	}
	
	// Configurar conexão com o coletor OTLP
	conn, err := grpc.DialContext(ctx,
		getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao coletor OTLP: %w", err)
	}
	
	// Criar o exportador
	traceExporter, err := otlptrace.New(ctx, otlptracegrpc.NewClient(
		otlptracegrpc.WithGRPCConn(conn),
	))
	if err != nil {
		return nil, fmt.Errorf("falha ao criar exportador OTLP: %w", err)
	}
	
	// Criar provedor de tracer
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			attribute.String("environment", getEnv("ENVIRONMENT", "development")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar recurso: %w", err)
	}
	
	tp := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()),
	)
	
	otel.SetTracerProvider(tp)
	return tp, nil
}

// Helper functions for environment variables

// getEnv retorna o valor da variável de ambiente ou o valor padrão
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvInt retorna o valor da variável de ambiente como inteiro ou o valor padrão
func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvDuration retorna o valor da variável de ambiente como duração ou o valor padrão
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}