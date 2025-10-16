// Package tracing fornece instrumentação para observabilidade dos hooks MCP-IAM
//
// Este pacote implementa tracing distribuído para os hooks MCP-IAM da plataforma INNOVABIZ,
// usando OpenTelemetry como framework de observabilidade. Suporta dimensões
// multi-mercado, multi-tenant e multi-contexto conforme requisitos da plataforma.
//
// Conformidades: ISO/IEC 27001, ISO 20000, COBIT 2019, NIST, TOGAF 10.0, DMBOK 2.0
// Frameworks: OpenTelemetry, W3C Trace Context, B3 Propagation
package tracing

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/innovabiz/iam/constants"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// HookTracer encapsula funcionalidades de tracing para hooks MCP-IAM
type HookTracer struct {
	tracer trace.Tracer
	logger *zap.Logger
}

// Constantes para tipos de operações de hooks
const (
	OperationValidateScope       = "validate_scope"
	OperationGetApprovers        = "get_approvers"
	OperationValidateMFA         = "validate_mfa"
	OperationValidateToken       = "validate_token"
	OperationGenerateAuditData   = "generate_audit_data"
	OperationCompleteElevation   = "complete_elevation"
)

// Constantes para atributos de span
const (
	AttributeMarket          = "market"
	AttributeTenantType      = "tenant_type"
	AttributeHookType        = "hook_type"
	AttributeScope           = "scope"
	AttributeMFALevel        = "mfa_level"
	AttributeTokenId         = "token_id"
	AttributeUserId          = "user_id"
	AttributeOperation       = "operation"
	AttributeStatus          = "status"
	AttributeComplianceLevel = "compliance_level"
	AttributeRegulation      = "regulation"
	AttributeDuration        = "duration_ms"
	AttributeError           = "error"
)

// NewHookTracer cria uma nova instância de HookTracer
func NewHookTracer(serviceName string, logger *zap.Logger) *HookTracer {
	tracer := otel.Tracer(serviceName)
	
	return &HookTracer{
		tracer: tracer,
		logger: logger,
	}
}

// InitTracer inicializa o provedor de tracing global
func InitTracer(ctx context.Context, serviceName, environment string, endpoint string) (func(context.Context) error, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(constants.Version),
			semconv.DeploymentEnvironmentKey.String(environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar recurso: %w", err)
	}

	// Configurar exportador OTLP
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar exportador: %w", err)
	}

	// Configurar processador de span em batch
	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	
	// Configurar provedor de tracer com exportador e recurso
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	
	// Configurar propagador global para compatibilidade W3C
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	
	// Definir provedor de tracer global
	otel.SetTracerProvider(tracerProvider)
	
	return tracerProvider.Shutdown, nil
}

// TraceHookOperation executa uma operação de hook com tracing
func (ht *HookTracer) TraceHookOperation(
	ctx context.Context,
	hookType, 
	market, 
	tenantType, 
	operation string, 
	attributes []attribute.KeyValue,
	operation_func func(context.Context) error,
) error {
	// Criar span para a operação
	ctx, span := ht.tracer.Start(
		ctx, 
		fmt.Sprintf("%s.%s", hookType, operation),
		trace.WithAttributes(
			attribute.String(AttributeHookType, hookType),
			attribute.String(AttributeMarket, market),
			attribute.String(AttributeTenantType, tenantType),
			attribute.String(AttributeOperation, operation),
		),
	)
	defer span.End()
	
	// Adicionar atributos customizados ao span
	span.SetAttributes(attributes...)

	startTime := time.Now()
	err := operation_func(ctx)
	duration := time.Since(startTime).Milliseconds()
	
	// Registrar duração da operação
	span.SetAttributes(attribute.Int64(AttributeDuration, duration))
	
	// Registrar resultado da operação
	if err != nil {
		span.SetAttributes(
			attribute.String(AttributeStatus, "error"),
			attribute.String(AttributeError, err.Error()),
		)
		span.RecordError(err)
		
		// Logar erro detalhado
		ht.logger.Error("Falha na operação de hook",
			zap.String("hook_type", hookType),
			zap.String("market", market),
			zap.String("tenant_type", tenantType),
			zap.String("operation", operation),
			zap.Int64("duration_ms", duration),
			zap.Error(err),
		)
	} else {
		span.SetAttributes(attribute.String(AttributeStatus, "success"))
		
		// Logar sucesso com informações detalhadas para auditoria
		if ht.logger.Core().Enabled(zap.DebugLevel) {
			ht.logger.Debug("Operação de hook bem-sucedida",
				zap.String("hook_type", hookType),
				zap.String("market", market),
				zap.String("tenant_type", tenantType),
				zap.String("operation", operation),
				zap.Int64("duration_ms", duration),
			)
		}
	}
	
	return err
}

// TraceScopeValidation executa validação de escopo com tracing
func (ht *HookTracer) TraceScopeValidation(
	ctx context.Context,
	hookType, 
	market, 
	tenantType,
	scope string,
	validate_func func(context.Context) error,
) error {
	attributes := []attribute.KeyValue{
		attribute.String(AttributeScope, scope),
	}
	
	return ht.TraceHookOperation(
		ctx,
		hookType,
		market,
		tenantType,
		OperationValidateScope,
		attributes,
		validate_func,
	)
}

// TraceMFAValidation executa validação MFA com tracing
func (ht *HookTracer) TraceMFAValidation(
	ctx context.Context,
	hookType, 
	market, 
	tenantType,
	mfaLevel,
	userId string,
	validate_func func(context.Context) error,
) error {
	attributes := []attribute.KeyValue{
		attribute.String(AttributeMFALevel, mfaLevel),
		attribute.String(AttributeUserId, userId),
	}
	
	return ht.TraceHookOperation(
		ctx,
		hookType,
		market,
		tenantType,
		OperationValidateMFA,
		attributes,
		validate_func,
	)
}

// TraceTokenValidation executa validação de token com tracing
func (ht *HookTracer) TraceTokenValidation(
	ctx context.Context,
	hookType, 
	market, 
	tenantType,
	tokenId,
	scope string,
	validate_func func(context.Context) error,
) error {
	attributes := []attribute.KeyValue{
		attribute.String(AttributeTokenId, tokenId),
		attribute.String(AttributeScope, scope),
	}
	
	return ht.TraceHookOperation(
		ctx,
		hookType,
		market,
		tenantType,
		OperationValidateToken,
		attributes,
		validate_func,
	)
}

// TraceComplianceCheck executa verificação de compliance com tracing
func (ht *HookTracer) TraceComplianceCheck(
	ctx context.Context,
	hookType, 
	market, 
	tenantType,
	complianceLevel,
	regulation string,
	check_func func(context.Context) error,
) error {
	attributes := []attribute.KeyValue{
		attribute.String(AttributeComplianceLevel, complianceLevel),
		attribute.String(AttributeRegulation, regulation),
	}
	
	ctx, span := ht.tracer.Start(
		ctx, 
		fmt.Sprintf("%s.compliance_check", hookType),
		trace.WithAttributes(attributes...),
	)
	defer span.End()
	
	err := check_func(ctx)
	
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("compliance_result", "failed"))
	} else {
		span.SetAttributes(attribute.String("compliance_result", "passed"))
	}
	
	return err
}

// TraceAuditEvent registra um evento de auditoria com tracing
func (ht *HookTracer) TraceAuditEvent(
	ctx context.Context,
	hookType, 
	market, 
	tenantType,
	userId,
	eventType,
	eventDetails string,
) {
	ctx, span := ht.tracer.Start(
		ctx, 
		fmt.Sprintf("%s.audit_event", hookType),
		trace.WithAttributes(
			attribute.String(AttributeHookType, hookType),
			attribute.String(AttributeMarket, market),
			attribute.String(AttributeTenantType, tenantType),
			attribute.String(AttributeUserId, userId),
			attribute.String("event_type", eventType),
			attribute.String("event_details", eventDetails),
		),
	)
	defer span.End()
	
	// Registrar evento de auditoria nos logs
	ht.logger.Info("Evento de auditoria",
		zap.String("hook_type", hookType),
		zap.String("market", market),
		zap.String("tenant_type", tenantType),
		zap.String("user_id", userId),
		zap.String("event_type", eventType),
		zap.String("event_details", eventDetails),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
		zap.String("span_id", span.SpanContext().SpanID().String()),
	)
}

// ExtractTraceInfo extrai informações de trace do contexto
func (ht *HookTracer) ExtractTraceInfo(ctx context.Context) map[string]string {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return map[string]string{}
	}
	
	return map[string]string{
		"trace_id": span.SpanContext().TraceID().String(),
		"span_id":  span.SpanContext().SpanID().String(),
	}
}