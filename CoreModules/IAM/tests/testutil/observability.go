// Package testutil fornece utilitários para testes do MCP-IAM
package testutil

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TestObservability gerencia os componentes de observabilidade para testes
type TestObservability struct {
	logger        *zap.SugaredLogger
	tracer        opentracing.Tracer
	tracerCloser  func()
	testName      string
	metricsPrefix string
}

// NewTestObservability cria uma nova instância de TestObservability
func NewTestObservability(testName string) (*TestObservability, error) {
	// Configurar logger
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	loggerConfig.EncoderConfig.TimeKey = "timestamp"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Determinar se deve logar para arquivo ou console
	outputPaths := []string{"stdout"}
	if os.Getenv("TEST_LOG_FILE") != "" {
		outputPaths = append(outputPaths, os.Getenv("TEST_LOG_FILE"))
	}
	loggerConfig.OutputPaths = outputPaths

	// Criar logger
	baseLogger, err := loggerConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("falha ao criar logger: %w", err)
	}

	// Configurar tracer
	tracerConfig := jaegercfg.Configuration{
		ServiceName: fmt.Sprintf("mcp-iam-test-%s", testName),
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1.0, // Sempre rastrear durante testes
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
		},
	}

	// Verificar se deve usar agente Jaeger real ou coletor noop
	var tracer opentracing.Tracer
	var tracerCloser func()

	if os.Getenv("JAEGER_AGENT_HOST") != "" {
		jLogger := log.StdLogger
		tr, closer, err := tracerConfig.NewTracer(
			jaegercfg.Logger(jLogger),
		)
		if err != nil {
			return nil, fmt.Errorf("falha ao criar tracer: %w", err)
		}
		tracer = tr
		tracerCloser = func() { closer.Close() }
	} else {
		// Usar tracer noop para testes que não precisam de tracing distribuído
		tracer = opentracing.NoopTracer{}
		tracerCloser = func() {}
	}

	return &TestObservability{
		logger:        baseLogger.Sugar().With("test", testName),
		tracer:        tracer,
		tracerCloser:  tracerCloser,
		testName:      testName,
		metricsPrefix: fmt.Sprintf("mcp_iam_test_%s", testName),
	}, nil
}

// Logger retorna o logger configurado
func (o *TestObservability) Logger() *zap.SugaredLogger {
	return o.logger
}

// Tracer retorna o tracer configurado
func (o *TestObservability) Tracer() opentracing.Tracer {
	return o.tracer
}

// RecordTestStart registra o início de um teste com um span
func (o *TestObservability) RecordTestStart(ctx context.Context, testID string) context.Context {
	o.logger.Infow("Iniciando teste",
		"test_id", testID,
		"timestamp", time.Now().Format(time.RFC3339),
	)

	span := o.tracer.StartSpan(fmt.Sprintf("test.%s", testID))
	ctx = opentracing.ContextWithSpan(ctx, span)

	return ctx
}

// RecordTestEnd registra o fim de um teste
func (o *TestObservability) RecordTestEnd(ctx context.Context, testID string, success bool, duration time.Duration) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.SetTag("test.success", success)
		span.SetTag("test.duration_ms", duration.Milliseconds())
		span.Finish()
	}

	o.logger.Infow("Teste concluído",
		"test_id", testID,
		"success", success,
		"duration_ms", duration.Milliseconds(),
		"timestamp", time.Now().Format(time.RFC3339),
	)
}

// RecordError registra um erro durante o teste
func (o *TestObservability) RecordError(ctx context.Context, err error, message string, keysAndValues ...interface{}) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.SetTag("error", true)
		span.SetTag("error.message", err.Error())
	}

	fields := append([]interface{}{"error", err}, keysAndValues...)
	o.logger.Errorw(message, fields...)
}

// RecordAction registra uma ação durante o teste
func (o *TestObservability) RecordAction(ctx context.Context, action string, keysAndValues ...interface{}) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.SetTag("action", action)
		for i := 0; i < len(keysAndValues); i += 2 {
			if i+1 < len(keysAndValues) {
				key, ok := keysAndValues[i].(string)
				if ok {
					span.SetTag(key, keysAndValues[i+1])
				}
			}
		}
	}

	o.logger.Infow(action, keysAndValues...)
}

// StartOperation inicia um span para uma operação específica
func (o *TestObservability) StartOperation(ctx context.Context, operationName string) (context.Context, opentracing.Span) {
	var span opentracing.Span
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		span = o.tracer.StartSpan(operationName, opentracing.ChildOf(parentSpan.Context()))
	} else {
		span = o.tracer.StartSpan(operationName)
	}

	return opentracing.ContextWithSpan(ctx, span), span
}

// RecordMetric registra uma métrica durante o teste
func (o *TestObservability) RecordMetric(name string, value float64, tags map[string]string) {
	// Em uma implementação real, conectaria com Prometheus, StatsD, etc.
	// Para testes, apenas logamos a métrica
	tagFields := make([]interface{}, 0, len(tags)*2)
	for k, v := range tags {
		tagFields = append(tagFields, k, v)
	}

	o.logger.Infow(fmt.Sprintf("Metric: %s", name),
		append([]interface{}{"metric", name, "value", value}, tagFields...)...,
	)
}

// RecordEvent registra um evento durante o teste
func (o *TestObservability) RecordEvent(ctx context.Context, eventType string, metadata map[string]interface{}) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.SetTag("event.type", eventType)
		for k, v := range metadata {
			span.SetTag(fmt.Sprintf("event.%s", k), v)
		}
	}

	fields := make([]interface{}, 0, len(metadata)*2+2)
	fields = append(fields, "event_type", eventType)
	for k, v := range metadata {
		fields = append(fields, k, v)
	}

	o.logger.Infow("Evento registrado", fields...)
}

// Shutdown finaliza a observabilidade, fechando tracers
func (o *TestObservability) Shutdown(ctx context.Context) {
	if o.tracerCloser != nil {
		o.tracerCloser()
	}

	// Garantir que todos os logs sejam descarregados
	_ = o.logger.Sync()
}

// WithCorrelationID adiciona um ID de correlação ao contexto
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.SetTag("correlation_id", correlationID)
	}
	return ctx
}

// WithUserID adiciona um ID de usuário ao contexto
func WithUserID(ctx context.Context, userID string) context.Context {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.SetTag("user_id", userID)
	}
	return ctx
}

// WithTenantID adiciona um ID de tenant ao contexto
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.SetTag("tenant_id", tenantID)
	}
	return ctx
}

// WithMarket adiciona um mercado ao contexto
func WithMarket(ctx context.Context, market string) context.Context {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.SetTag("market", market)
	}
	return ctx
}

// RecordAPICall registra uma chamada de API durante o teste
func (o *TestObservability) RecordAPICall(ctx context.Context, method, path string, statusCode int, durationMs int64) {
	o.RecordAction(ctx, "API Call",
		"method", method,
		"path", path,
		"status_code", statusCode,
		"duration_ms", durationMs,
	)

	o.RecordMetric("api_call_duration_ms", float64(durationMs), map[string]string{
		"method":      method,
		"path":        path,
		"status_code": fmt.Sprintf("%d", statusCode),
	})
}

// RecordDBQuery registra uma consulta de banco de dados durante o teste
func (o *TestObservability) RecordDBQuery(ctx context.Context, operation string, durationMs int64) {
	o.RecordAction(ctx, "Database Query",
		"operation", operation,
		"duration_ms", durationMs,
	)

	o.RecordMetric("db_query_duration_ms", float64(durationMs), map[string]string{
		"operation": operation,
	})
}

// RecordCacheOperation registra uma operação de cache durante o teste
func (o *TestObservability) RecordCacheOperation(ctx context.Context, operation string, key string, hit bool, durationMs int64) {
	o.RecordAction(ctx, "Cache Operation",
		"operation", operation,
		"key", key,
		"hit", hit,
		"duration_ms", durationMs,
	)

	o.RecordMetric("cache_operation_duration_ms", float64(durationMs), map[string]string{
		"operation": operation,
		"hit":       fmt.Sprintf("%t", hit),
	})
}

// RecordMFAOperation registra uma operação MFA durante o teste
func (o *TestObservability) RecordMFAOperation(ctx context.Context, operation string, method string, success bool) {
	o.RecordAction(ctx, "MFA Operation",
		"operation", operation,
		"method", method,
		"success", success,
	)

	o.RecordMetric("mfa_operation", 1.0, map[string]string{
		"operation": operation,
		"method":    method,
		"success":   fmt.Sprintf("%t", success),
	})
}

// RecordElevationOperation registra uma operação de elevação durante o teste
func (o *TestObservability) RecordElevationOperation(ctx context.Context, operation string, elevationID string, success bool) {
	o.RecordAction(ctx, "Elevation Operation",
		"operation", operation,
		"elevation_id", elevationID,
		"success", success,
	)

	o.RecordMetric("elevation_operation", 1.0, map[string]string{
		"operation":    operation,
		"elevation_id": elevationID,
		"success":      fmt.Sprintf("%t", success),
	})
}

// RecordRegulatoryCheck registra uma verificação regulatória durante o teste
func (o *TestObservability) RecordRegulatoryCheck(ctx context.Context, regulation string, market string, passed bool) {
	o.RecordAction(ctx, "Regulatory Check",
		"regulation", regulation,
		"market", market,
		"passed", passed,
	)

	o.RecordMetric("regulatory_check", 1.0, map[string]string{
		"regulation": regulation,
		"market":     market,
		"passed":     fmt.Sprintf("%t", passed),
	})
}

// RecordIsolationCheck registra uma verificação de isolamento multi-tenant durante o teste
func (o *TestObservability) RecordIsolationCheck(ctx context.Context, accessType string, targetTenant string, sourceTenant string, allowed bool) {
	o.RecordAction(ctx, "Multi-Tenant Isolation Check",
		"access_type", accessType,
		"target_tenant", targetTenant,
		"source_tenant", sourceTenant,
		"allowed", allowed,
	)

	o.RecordMetric("isolation_check", 1.0, map[string]string{
		"access_type":   accessType,
		"target_tenant": targetTenant,
		"source_tenant": sourceTenant,
		"allowed":       fmt.Sprintf("%t", allowed),
	})
}