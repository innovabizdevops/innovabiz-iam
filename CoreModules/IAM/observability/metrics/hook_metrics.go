// Package metrics fornece instrumentação para métricas dos hooks MCP-IAM
//
// Este pacote implementa métricas Prometheus para os hooks MCP-IAM da plataforma INNOVABIZ,
// seguindo os padrões de observabilidade e governança internacional. Suporta dimensões
// multi-mercado, multi-tenant e multi-contexto conforme requisitos da plataforma.
//
// Conformidades: ISO/IEC 27001, ISO 20000, COBIT 2019, NIST, TOGAF 10.0, DMBOK 2.0
// Frameworks: OpenTelemetry, Prometheus, RED Method, USE Method
package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Constantes para dimensões de mercado e tipos de tenant
const (
	// Mercados suportados
	MarketAngola      = "angola"
	MarketBrasil      = "brasil"
	MarketUE          = "eu"
	MarketChina       = "china"
	MarketMocambique  = "mocambique"
	MarketBRICS       = "brics"
	MarketEUA         = "eua"
	MarketSADC        = "sadc"
	MarketPALOP       = "palop"
	MarketGlobal      = "global"

	// Tipos de tenant
	TenantFinancial      = "financial"
	TenantGovernment     = "government"
	TenantHealthcare     = "healthcare"
	TenantRetail         = "retail"
	TenantTelecom        = "telecom"
	TenantEducation      = "education"
	TenantEnergy         = "energy"
	TenantManufacturing  = "manufacturing"

	// Tipos de hook
	HookTypeFigma  = "figma"
	HookTypeDocker = "docker"
	HookTypeGitHub = "github"

	// Níveis de conformidade
	ComplianceStandard  = "standard"
	ComplianceEnhanced  = "enhanced"
	ComplianceStrict    = "strict"
)

var (
	// ElevationRequestsTotal conta o número total de solicitações de elevação
	ElevationRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "innovabiz",
			Subsystem: "iam",
			Name:      "hook_elevation_requests_total",
			Help:      "Número total de solicitações de elevação processadas pelos hooks IAM",
		},
		[]string{"environment", "market", "tenant_type", "hook_type"},
	)

	// ElevationRequestsRejectedTotal conta o número de solicitações rejeitadas
	ElevationRequestsRejectedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "innovabiz",
			Subsystem: "iam",
			Name:      "hook_elevation_requests_rejected_total",
			Help:      "Número total de solicitações de elevação rejeitadas",
		},
		[]string{"environment", "market", "tenant_type", "hook_type", "reason"},
	)

	// ValidationDuration mede o tempo de validação de hooks
	ValidationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "innovabiz",
			Subsystem: "iam",
			Name:      "hook_validation_duration_milliseconds",
			Help:      "Duração das operações de validação em milissegundos",
			Buckets:   prometheus.ExponentialBuckets(0.1, 2, 10), // 0.1ms a ~100ms
		},
		[]string{"environment", "market", "tenant_type", "hook_type", "operation"},
	)

	// ComplianceChecksTotal conta o número de verificações de compliance
	ComplianceChecksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "innovabiz",
			Subsystem: "iam",
			Name:      "hook_compliance_checks_total",
			Help:      "Número total de verificações de conformidade realizadas",
		},
		[]string{"environment", "market", "compliance_type", "hook_type"},
	)

	// MFAChecksTotal conta o número de verificações MFA
	MFAChecksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "innovabiz",
			Subsystem: "iam",
			Name:      "hook_mfa_checks_total",
			Help:      "Número total de verificações MFA realizadas",
		},
		[]string{"environment", "market", "mfa_level", "success", "hook_type"},
	)

	// TokenOperationsTotal conta operações relacionadas a tokens
	TokenOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "innovabiz",
			Subsystem: "iam",
			Name:      "hook_token_operations_total",
			Help:      "Número total de operações de token realizadas",
		},
		[]string{"environment", "market", "operation", "hook_type"},
	)

	// ActiveElevations monitora o número de elevações ativas
	ActiveElevations = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "innovabiz",
			Subsystem: "iam",
			Name:      "hook_active_elevations",
			Help:      "Número atual de elevações de privilégio ativas",
		},
		[]string{"environment", "market", "tenant_type", "hook_type"},
	)

	// TestCoverage monitora a cobertura de testes por hook
	TestCoverage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "innovabiz",
			Subsystem: "iam",
			Name:      "hook_test_coverage",
			Help:      "Porcentagem de cobertura de testes por hook",
		},
		[]string{"environment", "hook_type"},
	)

	// ComplianceMetadata expõe metadados de compliance por mercado
	ComplianceMetadata = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "innovabiz",
			Subsystem: "iam",
			Name:      "hook_compliance_metadata",
			Help:      "Metadados de compliance por mercado",
		},
		[]string{
			"environment", 
			"market", 
			"compliance_framework", 
			"requires_dual_approval", 
			"requires_mfa", 
			"retention_years",
		},
	)
)

// HookMetrics encapsula métricas e tracing para hooks
type HookMetrics struct {
	environment string
	tracer      trace.Tracer
}

// NewHookMetrics cria uma nova instância de HookMetrics
func NewHookMetrics(environment string, tracer trace.Tracer) *HookMetrics {
	return &HookMetrics{
		environment: environment,
		tracer:      tracer,
	}
}

// ObserveValidation mede o tempo de execução de uma operação de validação
func (hm *HookMetrics) ObserveValidation(
	ctx context.Context, 
	market, 
	tenantType, 
	hookType, 
	operation string, 
	f func(context.Context) error,
) error {
	// Iniciar tracing
	ctx, span := hm.tracer.Start(
		ctx, 
		"hook.validation."+operation,
		trace.WithAttributes(
			attribute.String("market", market),
			attribute.String("tenant_type", tenantType),
			attribute.String("hook_type", hookType),
			attribute.String("operation", operation),
		),
	)
	defer span.End()

	// Métricas de duração
	start := time.Now()
	err := f(ctx)
	duration := time.Since(start).Milliseconds()

	// Registrar métricas
	ValidationDuration.WithLabelValues(
		hm.environment, market, tenantType, hookType, operation,
	).Observe(float64(duration))

	// Registrar erro no span e métricas, se houver
	if err != nil {
		span.RecordError(err)
		span.SetStatus(trace.StatusCodeError, err.Error())
		
		ElevationRequestsRejectedTotal.WithLabelValues(
			hm.environment, market, tenantType, hookType, err.Error(),
		).Inc()
	}

	return err
}

// RecordElevationRequest registra uma solicitação de elevação
func (hm *HookMetrics) RecordElevationRequest(
	ctx context.Context, 
	market, 
	tenantType, 
	hookType string,
) {
	ElevationRequestsTotal.WithLabelValues(
		hm.environment, market, tenantType, hookType,
	).Inc()

	span := trace.SpanFromContext(ctx)
	span.AddEvent("elevation.request", trace.WithAttributes(
		attribute.String("market", market),
		attribute.String("tenant_type", tenantType),
		attribute.String("hook_type", hookType),
	))
}

// RecordComplianceCheck registra uma verificação de compliance
func (hm *HookMetrics) RecordComplianceCheck(
	ctx context.Context, 
	market, 
	complianceType, 
	hookType string,
) {
	ComplianceChecksTotal.WithLabelValues(
		hm.environment, market, complianceType, hookType,
	).Inc()

	span := trace.SpanFromContext(ctx)
	span.AddEvent("compliance.check", trace.WithAttributes(
		attribute.String("market", market),
		attribute.String("compliance_type", complianceType),
		attribute.String("hook_type", hookType),
	))
}

// RecordMFACheck registra uma verificação MFA
func (hm *HookMetrics) RecordMFACheck(
	ctx context.Context, 
	market, 
	mfaLevel string, 
	success bool, 
	hookType string,
) {
	successStr := "false"
	if success {
		successStr = "true"
	}

	MFAChecksTotal.WithLabelValues(
		hm.environment, market, mfaLevel, successStr, hookType,
	).Inc()

	span := trace.SpanFromContext(ctx)
	span.AddEvent("mfa.check", trace.WithAttributes(
		attribute.String("market", market),
		attribute.String("mfa_level", mfaLevel),
		attribute.Bool("success", success),
		attribute.String("hook_type", hookType),
	))
}

// RecordTokenOperation registra uma operação de token
func (hm *HookMetrics) RecordTokenOperation(
	ctx context.Context, 
	market, 
	operation, 
	hookType string,
) {
	TokenOperationsTotal.WithLabelValues(
		hm.environment, market, operation, hookType,
	).Inc()

	span := trace.SpanFromContext(ctx)
	span.AddEvent("token.operation", trace.WithAttributes(
		attribute.String("market", market),
		attribute.String("operation", operation),
		attribute.String("hook_type", hookType),
	))
}

// UpdateActiveElevations atualiza o contador de elevações ativas
func (hm *HookMetrics) UpdateActiveElevations(
	market, 
	tenantType, 
	hookType string, 
	count int,
) {
	ActiveElevations.WithLabelValues(
		hm.environment, market, tenantType, hookType,
	).Set(float64(count))
}

// RecordTestCoverage registra a cobertura de testes
func (hm *HookMetrics) RecordTestCoverage(hookType string, coverage float64) {
	TestCoverage.WithLabelValues(hm.environment, hookType).Set(coverage)
}

// RegisterComplianceMetadata registra metadados de compliance
func (hm *HookMetrics) RegisterComplianceMetadata(
	market, 
	framework string, 
	requiresDualApproval bool, 
	mfaLevel string,
	retentionYears int,
) {
	dualApproval := "false"
	if requiresDualApproval {
		dualApproval = "true"
	}

	retentionYearsStr := fmt.Sprintf("%d", retentionYears)

	ComplianceMetadata.WithLabelValues(
		hm.environment,
		market,
		framework,
		dualApproval,
		mfaLevel,
		retentionYearsStr,
	).Set(1)
}