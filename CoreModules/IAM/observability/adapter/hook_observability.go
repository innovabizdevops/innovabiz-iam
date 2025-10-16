// Package adapter fornece uma camada de adaptação para integração dos componentes
// de observabilidade dos hooks MCP-IAM da plataforma INNOVABIZ.
//
// Este adaptador unifica métricas, tracing e logging em uma interface coerente
// para monitoramento holístico das operações de hook, garantindo compliance com
// normas e frameworks internacionais, com suporte às dimensões multi-mercado,
// multi-tenant e multi-contexto.
//
// Conformidades: ISO/IEC 27001, ISO 20000, COBIT 2019, TOGAF 10.0, DMBOK 2.0
// Frameworks: OpenTelemetry, Prometheus, RED Method, USE Method
package adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/innovabiz/iam/constants"
	"github.com/innovabiz/iam/observability/logging"
	"github.com/innovabiz/iam/observability/metrics"
	"github.com/innovabiz/iam/observability/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// HookObservability é um adaptador que integra métricas, tracing e logging
// para monitoramento completo das operações de hooks MCP-IAM.
type HookObservability struct {
	metrics    *metrics.HookMetrics
	tracer     *tracing.HookTracer
	logger     *logging.HookLogger
	env        string
	serviceName string
}

// Config contém configurações para o adaptador de observabilidade
type Config struct {
	Environment           string
	ServiceName           string
	OTLPEndpoint          string
	MetricsPort           int
	ComplianceLogsPath    string
	EnableComplianceAudit bool
	StructuredLogging     bool
	LogLevel              string
}

// MarketContext contém informações de contexto específicas de mercado para operações de hook
type MarketContext struct {
	Market             string
	TenantType         string
	HookType           string
	ComplianceLevel    string
	ApplicableRegulations []string
}

// NewHookObservability cria um novo adaptador de observabilidade
func NewHookObservability(config Config) (*HookObservability, error) {
	var err error

	// Inicializar logger
	var hookLogger *logging.HookLogger
	if config.EnableComplianceAudit && config.ComplianceLogsPath != "" {
		hookLogger, err = logging.NewComplianceLogger(config.Environment, config.ComplianceLogsPath)
		if err != nil {
			return nil, fmt.Errorf("falha ao inicializar logger de compliance: %w", err)
		}
	} else if config.Environment == constants.EnvDevelopment {
		hookLogger = logging.NewDevelopmentLogger()
	} else {
		hookLogger = logging.NewHookLogger(config.Environment)
	}
	
	// Inicializar logger zap para tracing
	logger := hookLogger.WithContext(context.Background())
	
	// Inicializar tracer
	hookTracer := tracing.NewHookTracer(config.ServiceName, logger)
	
	// Inicializar métricas
	tracer := trace.Tracer(config.ServiceName)
	hookMetrics := metrics.NewHookMetrics(config.Environment, tracer)
	
	return &HookObservability{
		metrics:     hookMetrics,
		tracer:      hookTracer,
		logger:      hookLogger,
		env:         config.Environment,
		serviceName: config.ServiceName,
	}, nil
}

// ObserveHookOperation executa uma operação de hook com observabilidade completa (métricas, tracing e logging)
func (ho *HookObservability) ObserveHookOperation(
	ctx context.Context, 
	marketCtx MarketContext, 
	operation string,
	userId string, 
	description string, 
	attributes []attribute.KeyValue,
	operationFunc func(context.Context) error,
) error {
	// Adicionar informações de contexto ao logger
	logCtx := ho.logger.WithContext(ctx)
	
	// Registrar início da operação
	logCtx.Info(fmt.Sprintf("Iniciando operação de hook: %s", description),
		zap.String("hook_type", marketCtx.HookType),
		zap.String("market", marketCtx.Market),
		zap.String("tenant_type", marketCtx.TenantType),
		zap.String("operation", operation),
		zap.String("user_id", userId),
	)
	
	// Registrar solicitação de elevação em métricas
	ho.metrics.RecordElevationRequest(ctx, marketCtx.Market, marketCtx.TenantType, marketCtx.HookType)
	
	// Registrar verificação de compliance
	if marketCtx.ComplianceLevel != "" {
		ho.metrics.RecordComplianceCheck(
			ctx, 
			marketCtx.Market, 
			marketCtx.ComplianceLevel, 
			marketCtx.HookType,
		)
	}
	
	// Executar operação com instrumentação
	startTime := time.Now()
	err := ho.tracer.TraceHookOperation(
		ctx,
		marketCtx.HookType,
		marketCtx.Market,
		marketCtx.TenantType,
		operation,
		attributes,
		operationFunc,
	)
	duration := time.Since(startTime)
	
	// Registrar duração da operação em métricas
	ho.metrics.ValidationDuration.WithLabelValues(
		ho.env,
		marketCtx.Market,
		marketCtx.TenantType,
		marketCtx.HookType,
		operation,
	).Observe(float64(duration.Milliseconds()))
	
	// Registrar resultado da operação
	if err != nil {
		// Incrementar contador de rejeições
		ho.metrics.ElevationRequestsRejectedTotal.WithLabelValues(
			ho.env,
			marketCtx.Market,
			marketCtx.TenantType,
			marketCtx.HookType,
			err.Error(),
		).Inc()
		
		// Registrar erro no log
		ho.logger.LogHookError(
			ctx,
			marketCtx.Market,
			marketCtx.TenantType,
			marketCtx.HookType,
			operation,
			userId,
			err,
			zap.Duration("duration", duration),
		)
		
		// Registrar evento de auditoria para rejeição
		ho.TraceAuditEvent(
			ctx,
			marketCtx,
			userId,
			"elevation_rejected",
			fmt.Sprintf("Solicitação rejeitada: %s", err.Error()),
		)
	} else {
		// Registrar sucesso no log
		logCtx.Info(fmt.Sprintf("Operação de hook concluída com sucesso: %s", description),
			zap.String("hook_type", marketCtx.HookType),
			zap.String("market", marketCtx.Market),
			zap.String("tenant_type", marketCtx.TenantType),
			zap.String("operation", operation),
			zap.String("user_id", userId),
			zap.Duration("duration", duration),
		)
		
		// Registrar evento de auditoria para sucesso
		ho.TraceAuditEvent(
			ctx,
			marketCtx,
			userId,
			"elevation_approved",
			fmt.Sprintf("Solicitação aprovada para operação: %s", operation),
		)
	}
	
	return err
}

// ObserveValidateScope instrumenta a validação de escopo de um hook
func (ho *HookObservability) ObserveValidateScope(
	ctx context.Context,
	marketCtx MarketContext,
	userId string,
	scope string,
	validateFunc func(context.Context) error,
) error {
	// Atributos específicos para validação de escopo
	attributes := []attribute.KeyValue{
		attribute.String("scope", scope),
		attribute.String("user_id", userId),
	}
	
	description := fmt.Sprintf("Validação de escopo '%s'", scope)
	
	// Adicionar verificações específicas de compliance para escopo
	for _, regulation := range marketCtx.ApplicableRegulations {
		ho.logger.LogComplianceEvent(
			ctx,
			marketCtx.Market,
			marketCtx.TenantType,
			marketCtx.HookType,
			constants.OperationValidateScope,
			userId,
			regulation,
			"verificando",
			fmt.Sprintf("Validação de escopo '%s' conforme regulação '%s'", scope, regulation),
		)
	}
	
	// Executar operação com observabilidade completa
	return ho.ObserveHookOperation(
		ctx,
		marketCtx,
		constants.OperationValidateScope,
		userId,
		description,
		attributes,
		validateFunc,
	)
}

// ObserveValidateMFA instrumenta a validação MFA de um hook
func (ho *HookObservability) ObserveValidateMFA(
	ctx context.Context,
	marketCtx MarketContext,
	userId string,
	mfaLevel string,
	validateFunc func(context.Context) error,
) error {
	// Atributos específicos para validação MFA
	attributes := []attribute.KeyValue{
		attribute.String("mfa_level", mfaLevel),
		attribute.String("user_id", userId),
	}
	
	description := fmt.Sprintf("Validação MFA nível '%s'", mfaLevel)
	
	// Executar operação com observabilidade completa
	err := ho.ObserveHookOperation(
		ctx,
		marketCtx,
		constants.OperationValidateMFA,
		userId,
		description,
		attributes,
		validateFunc,
	)
	
	// Registrar resultado específico de MFA
	if err != nil {
		ho.metrics.RecordMFACheck(ctx, marketCtx.Market, mfaLevel, false, marketCtx.HookType)
	} else {
		ho.metrics.RecordMFACheck(ctx, marketCtx.Market, mfaLevel, true, marketCtx.HookType)
	}
	
	return err
}

// ObserveValidateToken instrumenta a validação de token de um hook
func (ho *HookObservability) ObserveValidateToken(
	ctx context.Context,
	marketCtx MarketContext,
	userId string,
	tokenId string,
	scope string,
	validateFunc func(context.Context) error,
) error {
	// Atributos específicos para validação de token
	attributes := []attribute.KeyValue{
		attribute.String("token_id", tokenId),
		attribute.String("scope", scope),
		attribute.String("user_id", userId),
	}
	
	description := fmt.Sprintf("Validação de token '%s' para escopo '%s'", tokenId, scope)
	
	// Registrar operação de token
	ho.metrics.RecordTokenOperation(ctx, marketCtx.Market, "validate", marketCtx.HookType)
	
	// Executar operação com observabilidade completa
	return ho.ObserveHookOperation(
		ctx,
		marketCtx,
		constants.OperationValidateToken,
		userId,
		description,
		attributes,
		validateFunc,
	)
}

// ObserveGetApprovers instrumenta a obtenção de aprovadores para um hook
func (ho *HookObservability) ObserveGetApprovers(
	ctx context.Context,
	marketCtx MarketContext,
	userId string,
	scope string,
	getApproversFunc func(context.Context) error,
) error {
	// Atributos específicos para obtenção de aprovadores
	attributes := []attribute.KeyValue{
		attribute.String("scope", scope),
		attribute.String("user_id", userId),
	}
	
	description := fmt.Sprintf("Obtenção de aprovadores para escopo '%s'", scope)
	
	// Executar operação com observabilidade completa
	return ho.ObserveHookOperation(
		ctx,
		marketCtx,
		constants.OperationGetApprovers,
		userId,
		description,
		attributes,
		getApproversFunc,
	)
}

// ObserveGenerateAuditData instrumenta a geração de dados de auditoria
func (ho *HookObservability) ObserveGenerateAuditData(
	ctx context.Context,
	marketCtx MarketContext,
	userId string,
	scope string,
	generateAuditFunc func(context.Context) error,
) error {
	// Atributos específicos para geração de dados de auditoria
	attributes := []attribute.KeyValue{
		attribute.String("scope", scope),
		attribute.String("user_id", userId),
	}
	
	description := fmt.Sprintf("Geração de dados de auditoria para escopo '%s'", scope)
	
	// Executar operação com observabilidade completa
	return ho.ObserveHookOperation(
		ctx,
		marketCtx,
		constants.OperationGenerateAuditData,
		userId,
		description,
		attributes,
		generateAuditFunc,
	)
}

// ObserveCompleteElevation instrumenta a conclusão de uma elevação de privilégio
func (ho *HookObservability) ObserveCompleteElevation(
	ctx context.Context,
	marketCtx MarketContext,
	userId string,
	tokenId string,
	scope string,
	completeFunc func(context.Context) error,
) error {
	// Atributos específicos para conclusão de elevação
	attributes := []attribute.KeyValue{
		attribute.String("token_id", tokenId),
		attribute.String("scope", scope),
		attribute.String("user_id", userId),
	}
	
	description := fmt.Sprintf("Conclusão de elevação para token '%s' e escopo '%s'", tokenId, scope)
	
	// Registrar operação de token
	ho.metrics.RecordTokenOperation(ctx, marketCtx.Market, "complete_elevation", marketCtx.HookType)
	
	// Executar operação com observabilidade completa
	return ho.ObserveHookOperation(
		ctx,
		marketCtx,
		constants.OperationCompleteElevation,
		userId,
		description,
		attributes,
		completeFunc,
	)
}

// TraceAuditEvent registra um evento de auditoria com correlação de tracing
func (ho *HookObservability) TraceAuditEvent(
	ctx context.Context,
	marketCtx MarketContext,
	userId string,
	eventType string,
	eventDetails string,
) {
	// Registrar evento de auditoria no tracing
	ho.tracer.TraceAuditEvent(
		ctx,
		marketCtx.HookType,
		marketCtx.Market,
		marketCtx.TenantType,
		userId,
		eventType,
		eventDetails,
	)
	
	// Registrar evento de auditoria no log
	ho.logger.LogAuditEvent(
		ctx,
		marketCtx.Market,
		marketCtx.TenantType,
		marketCtx.HookType,
		"audit",
		userId,
		eventType,
		eventDetails,
	)
}

// TraceSecurity registra um evento de segurança com correlação de tracing
func (ho *HookObservability) TraceSecurity(
	ctx context.Context,
	marketCtx MarketContext,
	userId string,
	severity string,
	eventDetails string,
	operation string,
) {
	// Registrar evento de segurança no log
	ho.logger.LogSecurityEvent(
		ctx,
		marketCtx.Market,
		marketCtx.TenantType,
		marketCtx.HookType,
		operation,
		userId,
		severity,
		eventDetails,
	)
	
	// Adicionar evento ao span atual, se houver
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent("security.event", trace.WithAttributes(
			attribute.String("market", marketCtx.Market),
			attribute.String("tenant_type", marketCtx.TenantType),
			attribute.String("hook_type", marketCtx.HookType),
			attribute.String("severity", severity),
			attribute.String("event_details", eventDetails),
			attribute.String("user_id", userId),
		))
	}
}

// UpdateActiveElevations atualiza o contador de elevações ativas
func (ho *HookObservability) UpdateActiveElevations(
	marketCtx MarketContext,
	count int,
) {
	ho.metrics.UpdateActiveElevations(
		marketCtx.Market,
		marketCtx.TenantType,
		marketCtx.HookType,
		count,
	)
}

// RecordTestCoverage registra a cobertura de testes
func (ho *HookObservability) RecordTestCoverage(
	hookType string,
	coverage float64,
) {
	ho.metrics.RecordTestCoverage(hookType, coverage)
}

// RegisterComplianceMetadata registra metadados de compliance para um mercado
func (ho *HookObservability) RegisterComplianceMetadata(
	market string,
	framework string,
	requiresDualApproval bool,
	mfaLevel string,
	retentionYears int,
) {
	ho.metrics.RegisterComplianceMetadata(
		market,
		framework,
		requiresDualApproval,
		mfaLevel,
		retentionYears,
	)
}

// NewMarketContext cria um contexto de mercado com base nas constantes e requisitos
func NewMarketContext(market, tenantType, hookType string) MarketContext {
	// Obter requisitos de compliance para o mercado
	complianceReqs, exists := constants.ComplianceRequirements[market]
	if !exists {
		complianceReqs = constants.ComplianceRequirements[constants.MarketGlobal]
	}
	
	// Determinar nível de compliance com base no tipo de tenant
	complianceLevel := constants.ComplianceStandard
	switch tenantType {
	case constants.TenantFinancial, constants.TenantGovernment, constants.TenantHealthcare:
		complianceLevel = constants.ComplianceStrict
	case constants.TenantTelecom, constants.TenantEnergy:
		complianceLevel = constants.ComplianceEnhanced
	}
	
	// Obter regulações aplicáveis
	regulations := []string{}
	if frameworks, ok := complianceReqs["framework"].([]string); ok {
		regulations = frameworks
	}
	
	return MarketContext{
		Market:                market,
		TenantType:            tenantType,
		HookType:              hookType,
		ComplianceLevel:       complianceLevel,
		ApplicableRegulations: regulations,
	}
}