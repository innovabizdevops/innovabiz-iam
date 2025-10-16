package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("innovabiz.iam.infrastructure.persistence.postgres")

// Config armazena a configuração de conexão com o PostgreSQL
type Config struct {
	Host     string // Host do servidor PostgreSQL
	Port     int    // Porta do servidor PostgreSQL
	User     string // Usuário para autenticação
	Password string // Senha para autenticação
	Database string // Nome do banco de dados
	SSLMode  string // Modo SSL (disable, require, verify-ca, verify-full)

	// Configurações avançadas de pool de conexões
	MaxConns           int           // Máximo de conexões no pool (0 = sem limite)
	MinConns           int           // Mínimo de conexões no pool (0 = sem limite)
	MaxConnLifetime    time.Duration // Tempo máximo de vida de uma conexão (0 = sem limite)
	MaxConnIdleTime    time.Duration // Tempo máximo que uma conexão pode ficar ociosa (0 = sem limite)
	HealthCheckPeriod  time.Duration // Período para verificação de saúde das conexões
	StatementCacheMode string        // Modo de cache de statements (prepare, describe, disabled)
}

// DefaultConfig retorna uma configuração padrão otimizada para ambiente de desenvolvimento
func DefaultConfig() Config {
	return Config{
		Host:              "localhost",
		Port:              5432,
		User:              "postgres",
		Password:          "postgres",
		Database:          "innovabiz_iam",
		SSLMode:           "disable",
		MaxConns:          20,
		MinConns:          5,
		MaxConnLifetime:   1 * time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 5 * time.Minute,
	}
}

// ProductionConfig retorna uma configuração otimizada para ambiente de produção
func ProductionConfig() Config {
	return Config{
		Host:              "localhost",
		Port:              5432,
		User:              "postgres",
		Password:          "postgres", // Em produção, deve ser obtido de um gestor de segredos
		Database:          "innovabiz_iam",
		SSLMode:           "verify-full",
		MaxConns:          100,
		MinConns:          10,
		MaxConnLifetime:   30 * time.Minute,
		MaxConnIdleTime:   10 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
		StatementCacheMode: "prepare",
	}
}

// ConnString retorna a string de conexão para o PostgreSQL
func (c Config) ConnString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, c.SSLMode,
	)
}

// PoolConfig retorna a configuração do pool de conexões
func (c Config) PoolConfig() (*pgxpool.Config, error) {
	// Parse da string de conexão para obter a configuração base
	poolConfig, err := pgxpool.ParseConfig(c.ConnString())
	if err != nil {
		return nil, fmt.Errorf("erro ao analisar configuração de pool: %w", err)
	}

	// Configurações avançadas do pool
	if c.MaxConns > 0 {
		poolConfig.MaxConns = int32(c.MaxConns)
	}
	if c.MinConns > 0 {
		poolConfig.MinConns = int32(c.MinConns)
	}
	if c.MaxConnLifetime > 0 {
		poolConfig.MaxConnLifetime = c.MaxConnLifetime
	}
	if c.MaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime = c.MaxConnIdleTime
	}
	if c.HealthCheckPeriod > 0 {
		poolConfig.HealthCheckPeriod = c.HealthCheckPeriod
	}

	// Interceptor para adicionar telemetria e logging
	poolConfig.ConnConfig.Tracer = &pgxTracer{}

	// Adicionar middleware para logging e rastreamento
	poolConfig.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		log.Ctx(ctx).Debug().Msg("Adquirindo conexão do pool")
		return true
	}

	poolConfig.AfterRelease = func(conn *pgx.Conn) bool {
		log.Debug().Msg("Liberando conexão para o pool")
		return true
	}

	return poolConfig, nil
}

// pgxTracer implementa pgx.QueryTracer para instrumentação com OpenTelemetry
type pgxTracer struct{}

// TraceQueryStart é chamado quando uma consulta começa
func (p *pgxTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	if span := otel.GetTextMapPropagator().Extract(ctx, nil); span != nil {
		ctx, span := tracer.Start(ctx, "PGX:"+data.SQL)
		span.SetAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.statement", data.SQL),
		)
		return ctx
	}
	return ctx
}

// TraceQueryEnd é chamado quando uma consulta termina
func (p *pgxTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	if span := otel.GetTextMapPropagator().Extract(ctx, nil); span != nil {
		span := otel.Tracer("").SpanFromContext(ctx)
		if data.Err != nil {
			span.SetStatus(codes.Error, data.Err.Error())
			span.RecordError(data.Err)
		}
		span.End()
	}
}

// DB encapsula a conexão pool com PostgreSQL e métodos auxiliares
type DB struct {
	pool *pgxpool.Pool
}

// Connect inicializa uma nova conexão pool com o PostgreSQL
func Connect(config Config) (*DB, error) {
	ctx, span := tracer.Start(context.Background(), "PostgreSQL.Connect")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.host", config.Host),
		attribute.Int("db.port", config.Port),
		attribute.String("db.name", config.Database),
		attribute.String("db.user", config.User),
	)

	// Obter configuração do pool
	poolConfig, err := config.PoolConfig()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, fmt.Errorf("erro na configuração do pool: %w", err)
	}

	// Inicializar o pool de conexões
	pool, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, fmt.Errorf("erro ao conectar ao PostgreSQL: %w", err)
	}

	// Testar a conexão
	if err := pool.Ping(ctx); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		pool.Close()
		return nil, fmt.Errorf("erro ao verificar conexão com PostgreSQL: %w", err)
	}

	log.Info().
		Str("host", config.Host).
		Int("port", config.Port).
		Str("database", config.Database).
		Msg("Conexão com PostgreSQL estabelecida com sucesso")

	return &DB{pool: pool}, nil
}

// Close fecha o pool de conexões com o PostgreSQL
func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
		log.Info().Msg("Pool de conexões PostgreSQL fechado")
	}
}

// Pool retorna o pool de conexões para uso direto quando necessário
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// InTransaction executa operações dentro de uma transação com rastreamento OpenTelemetry
func (db *DB) InTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	ctx, span := tracer.Start(ctx, "PostgreSQL.InTransaction")
	defer span.End()

	// Iniciar a transação
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}

	// Executar a função dentro da transação
	err = fn(ctx, tx)
	
	// Determinar se é necessário commit ou rollback
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		
		// Tentar rollback
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Error().Err(rbErr).Msg("Erro ao fazer rollback de transação")
			span.RecordError(rbErr)
			// Retornar o erro original, pois é o mais importante
		}
		
		return err
	}

	// Tentar commit
	if err := tx.Commit(ctx); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return fmt.Errorf("erro ao fazer commit de transação: %w", err)
	}

	return nil
}