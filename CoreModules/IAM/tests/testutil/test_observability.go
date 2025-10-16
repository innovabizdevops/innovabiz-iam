// Package testutil fornece utilitários compartilhados para testes do módulo IAM
package testutil

import (
	"context"
	"time"
	
	"go.uber.org/zap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

// TestObservability encapsula funcionalidades de observabilidade para testes
// Implementa logging estruturado, tracing distribuído e coleta de métricas
// Compatível com padrões OpenTelemetry para integração com sistemas de monitoramento
type TestObservability struct {
	logger        *zap.Logger
	tracer        trace.Tracer
	traceProvider *trace.TracerProvider
	metricClient  MetricClient
}

// MetricClient interface para registro de métricas de testes
type MetricClient interface {
	CounterInc(name string, tags map[string]string)
	GaugeSet(name string, value float64, tags map[string]string)
	HistogramRecord(name string, value float64, tags map[string]string)
}

// NewTestObservability cria uma nova instância de observabilidade para testes
// Configura logging, tracing e métricas integradas para monitoramento detalhado
// dos testes automatizados, facilitando diagnóstico e análise de problemas
func NewTestObservability(testID string) (*TestObservability, error) {
	// Configurar logger com campos estruturados
	logger, err := zap.NewDevelopment(
		zap.Fields(
			zap.String("test_id", testID),
			zap.String("module", "iam"),
			zap.String("component", "authorization"),
		),
	)
	if err != nil {
		return nil, err
	}
	
	// Configurar tracing com Jaeger
	// Compatível com OpenTelemetry para análise distribuída
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(
		jaeger.WithEndpoint("http://jaeger:14268/api/traces"),
	))
	if err != nil {
		return nil, err
	}
	
	// Configurar provedor de traces com recursos e atributos
	// relevantes para identificação e categorização
	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("innovabiz-iam-tests"),
			attribute.String("test.id", testID),
			attribute.String("environment", "test"),
		)),
	)
	
	// Configurar OpenTelemetry com o provider criado
	otel.SetTracerProvider(traceProvider)
	tracer := traceProvider.Tracer("innovabiz-iam-test-tracer")
	
	// Cliente de métricas mock para uso em testes
	// Será substituído por implementação real em produção
	metricClient := NewMockMetricClient()
	
	return &TestObservability{
		logger:        logger,
		tracer:        tracer,
		traceProvider: traceProvider,
		metricClient:  metricClient,
	}, nil
}

// RecordTestStart registra o início de um teste com spans e logs
// Permite rastreamento completo da execução do teste e seus resultados
func (o *TestObservability) RecordTestStart(ctx context.Context, testName string) context.Context {
	ctx, span := o.tracer.Start(ctx, "test:"+testName)
	
	o.logger.Info("Teste iniciado",
		zap.String("test_name", testName),
		zap.Time("start_time", time.Now()),
	)
	
	o.metricClient.CounterInc("test_executions_total", map[string]string{
		"test_name": testName,
		"status":    "started",
	})
	
	return ctx
}

// RecordTestEnd registra o fim de um teste com resultado e duração
// Alimenta sistema de métricas e logs para análise de desempenho
func (o *TestObservability) RecordTestEnd(ctx context.Context, testName string, success bool, duration time.Duration) {
	span := trace.SpanFromContext(ctx)
	defer span.End()
	
	// Adicionar atributos ao span para análise posterior
	span.SetAttributes(
		attribute.Bool("test.success", success),
		attribute.Int64("test.duration_ms", duration.Milliseconds()),
	)
	
	// Registrar conclusão em log estruturado
	o.logger.Info("Teste concluído",
		zap.String("test_name", testName),
		zap.Bool("success", success),
		zap.Duration("duration", duration),
		zap.Time("end_time", time.Now()),
	)
	
	// Incrementar contadores de execução por resultado
	o.metricClient.CounterInc("test_executions_total", map[string]string{
		"test_name": testName,
		"status":    success ? "succeeded" : "failed",
	})
	
	// Registrar duração como métrica para análise de tendências
	o.metricClient.GaugeSet("test_duration_ms", float64(duration.Milliseconds()), map[string]string{
		"test_name": testName,
	})
}

// RecordTestEvent registra evento específico durante a execução do teste
// Útil para marcar pontos de interesse, como chamadas a serviços externos
func (o *TestObservability) RecordTestEvent(ctx context.Context, eventName string, attributes map[string]string) {
	span := trace.SpanFromContext(ctx)
	
	// Converter atributos de string para attribute.KeyValue
	attrs := make([]attribute.KeyValue, 0, len(attributes))
	for k, v := range attributes {
		attrs = append(attrs, attribute.String(k, v))
	}
	
	// Adicionar evento ao span atual
	span.AddEvent(eventName, trace.WithAttributes(attrs...))
	
	// Registrar evento também nos logs
	fields := []zap.Field{zap.String("event", eventName)}
	for k, v := range attributes {
		fields = append(fields, zap.String(k, v))
	}
	
	o.logger.Info("Evento de teste", fields...)
}

// Shutdown realiza limpeza de recursos de observabilidade
// Deve ser chamado no final dos testes para liberação adequada
func (o *TestObservability) Shutdown(ctx context.Context) error {
	if err := o.traceProvider.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}

// NewMockMetricClient cria um cliente de métricas simulado para testes
// Em produção, seria substituído por cliente real (Prometheus, etc.)
func NewMockMetricClient() MetricClient {
	return &mockMetricClient{}
}

// Implementação mock do cliente de métricas para testes
type mockMetricClient struct {
	counters   map[string]int
	gauges     map[string]float64
	histograms map[string][]float64
}

func (m *mockMetricClient) CounterInc(name string, tags map[string]string) {
	// Em implementação real, enviaria para Prometheus/StatsD/etc.
}

func (m *mockMetricClient) GaugeSet(name string, value float64, tags map[string]string) {
	// Em implementação real, enviaria para Prometheus/StatsD/etc.
}

func (m *mockMetricClient) HistogramRecord(name string, value float64, tags map[string]string) {
	// Em implementação real, enviaria para Prometheus/StatsD/etc.
}