/**
 * @file tracing_hook.go
 * @description Hook para distributed tracing do Bureau de Crédito
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package hooks

import (
	"context"
	
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TracingHook implementa distributed tracing para operações do Bureau de Crédito
type TracingHook struct {
	BaseHook
	tracer         trace.Tracer
	spanAttributes []attribute.KeyValue
}

// TracingConfig contém configurações para o hook de tracing
type TracingConfig struct {
	TracerName     string
	SpanAttributes []attribute.KeyValue
}

// NewTracingHook cria uma nova instância do hook de tracing
func NewTracingHook(config *TracingConfig) *TracingHook {
	if config == nil {
		config = &TracingConfig{
			TracerName: "innovabiz.iam.bureau-credito",
			SpanAttributes: []attribute.KeyValue{
				attribute.String("service", "bureau-credito"),
				attribute.String("component", "assessment"),
			},
		}
	}
	
	return &TracingHook{
		BaseHook: BaseHook{
			Name:     "tracing_hook",
			Priority: 5,
		},
		tracer:         otel.Tracer(config.TracerName),
		spanAttributes: config.SpanAttributes,
	}
}

// Execute implementa a interface Hook para tracing
func (h *TracingHook) Execute(ctx context.Context, hookType HookType, metadata HookMetadata, payload interface{}) error {
	switch hookType {
	case HookBefore:
		return h.handleBefore(ctx, metadata, payload)
	case HookAfter:
		return h.handleAfter(ctx, metadata, payload)
	case HookError:
		return h.handleError(ctx, metadata, payload)
	default:
		return nil
	}
}

// handleBefore processa o hook antes da operação principal
func (h *TracingHook) handleBefore(ctx context.Context, metadata HookMetadata, payload interface{}) error {
	// Criar um novo span ou obter o span atual do contexto
	spanName := string(metadata.OperationType)
	
	// Coletar atributos padrão para o span
	attrs := append(h.spanAttributes,
		attribute.String("request.id", metadata.RequestID),
		attribute.String("correlation.id", metadata.CorrelationID),
		attribute.String("provider.id", metadata.ProviderID),
		attribute.String("tenant.id", metadata.TenantID),
		attribute.String("user.id", metadata.UserID),
		attribute.String("environment", metadata.Environment),
		attribute.String("region", metadata.Region),
	)
	
	// Adicionar atributos dos labels
	for k, v := range metadata.Labels {
		attrs = append(attrs, attribute.String("label."+k, v))
	}
	
	// Criar span com opções
	ctx, span := h.tracer.Start(ctx,
		spanName,
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindServer),
	)
	
	// Armazenar o span no contexto para uso posterior
	return storeSpanInContext(ctx, span)
}

// handleAfter processa o hook após a operação principal
func (h *TracingHook) handleAfter(ctx context.Context, metadata HookMetadata, payload interface{}) error {
	// Recuperar span do contexto
	span, ok := getSpanFromContext(ctx)
	if !ok {
		return nil
	}
	
	// Adicionar atributos adicionais com base no resultado
	span.SetAttributes(attribute.Int64("duration_ms", metadata.Duration.Milliseconds()))
	
	// Adicionar eventos específicos por tipo de operação
	h.addOperationSpecificEvents(span, metadata, payload)
	
	// Finalizar span
	span.End()
	
	return nil
}

// handleError processa o hook em caso de erro
func (h *TracingHook) handleError(ctx context.Context, metadata HookMetadata, payload interface{}) error {
	// Recuperar span do contexto
	span, ok := getSpanFromContext(ctx)
	if !ok {
		return nil
	}
	
	// Marcar span como erro
	if err, ok := payload.(error); ok {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Error, "unknown error")
	}
	
	// Adicionar atributos adicionais
	span.SetAttributes(attribute.String("error.type", classifyError(payload)))
	
	// Não finalizar o span aqui, pois pode haver mais informações a adicionar
	// O span será finalizado no método After
	
	return nil
}

// addOperationSpecificEvents adiciona eventos específicos por tipo de operação
func (h *TracingHook) addOperationSpecificEvents(span trace.Span, metadata HookMetadata, payload interface{}) {
	// Implementar eventos específicos por tipo de operação
	switch metadata.OperationType {
	case OpAssessment:
		// Adicionar eventos específicos de avaliação de crédito
		span.AddEvent("assessment.complete", trace.WithAttributes(
			attribute.String("assessment.id", extractAssessmentID(payload)),
		))
		
	case OpBatchAssessment:
		// Adicionar eventos específicos de avaliação em lote
		span.AddEvent("batch_assessment.complete", trace.WithAttributes(
			attribute.Int("batch.size", extractBatchSize(payload)),
			attribute.Int("batch.success_count", extractBatchSuccessCount(payload)),
		))
		
	case OpFraudDetection:
		// Adicionar eventos específicos de detecção de fraude
		if fraudResult, ok := extractFraudDetectionResult(payload); ok {
			span.AddEvent("fraud_detection.result", trace.WithAttributes(
				attribute.String("fraud.type", fraudResult.FraudType),
				attribute.String("fraud.severity", fraudResult.Severity),
				attribute.Float64("fraud.score", fraudResult.Score),
			))
		}
	}
}

// Funções auxiliares para extração de dados das payloads

func extractAssessmentID(payload interface{}) string {
	// Implementação real depende da estrutura do payload
	// Este é apenas um stub
	return "assessment-123"
}

func extractBatchSize(payload interface{}) int {
	// Implementação real depende da estrutura do payload
	// Este é apenas um stub
	return 10
}

func extractBatchSuccessCount(payload interface{}) int {
	// Implementação real depende da estrutura do payload
	// Este é apenas um stub
	return 8
}

// classifyError classifica um erro para tracing
func classifyError(payload interface{}) string {
	// Implementação real depende da estrutura do payload
	// Este é apenas um stub
	if err, ok := payload.(error); ok {
		// Implementar classificação de erros
		return err.Error()
	}
	return "unknown_error"
}

// Chave de contexto para armazenar o span
type spanKey struct{}

// storeSpanInContext armazena um span no contexto
func storeSpanInContext(ctx context.Context, span trace.Span) error {
	return nil // Esta implementação é simplificada
}

// getSpanFromContext recupera um span do contexto
func getSpanFromContext(ctx context.Context) (trace.Span, bool) {
	// Implementação real recuperaria o span do contexto
	// Este é apenas um stub que simula falha
	return nil, false
}