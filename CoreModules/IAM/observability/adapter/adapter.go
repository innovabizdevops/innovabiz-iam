// Package adapter fornece um adaptador de observabilidade para instrumentação de hooks MCP-IAM,
// integrando métricas Prometheus, tracing OpenTelemetry e logging estruturado via Zap.
//
// Suporta múltiplos mercados e tenants com configurações de compliance específicas,
// permitindo observabilidade completa em ambientes regulados.
//
// Conformidades: ISO/IEC 27001, ISO 20000, COBIT 2019, TOGAF 10.0, DMBOK 2.0
package adapter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/innovabiz/iam/constants"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
)

// HookObservability é o adaptador principal de observabilidade para hooks MCP-IAM
type HookObservability struct {
	config            Config
	logger            *zap.Logger
	tracer            trace.Tracer
	tracerProvider    *sdktrace.TracerProvider
	metricsRegistry   *prometheus.Registry
	metricsServer     *http.Server
	complianceMetadata map[string]ComplianceMetadata
	mutex             sync.RWMutex

	// Métricas Prometheus
	hookCallsTotal          *prometheus.CounterVec
	hookErrorsTotal         *prometheus.CounterVec
	hookDurationSeconds     *prometheus.HistogramVec
	hookActiveElevations    *prometheus.GaugeVec
	mfaValidationTotal      *prometheus.CounterVec
	scopeValidationTotal    *prometheus.CounterVec
	testCoveragePct         *prometheus.GaugeVec
	complianceEventsTotal   *prometheus.CounterVec
	securityEventsTotal     *prometheus.CounterVec
}

// NewHookObservability cria uma nova instância do adaptador de observabilidade
func NewHookObservability(config Config) (*HookObservability, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuração inválida: %w", err)
	}

	if err := config.EnsureDirectories(); err != nil {
		return nil, fmt.Errorf("falha ao criar diretórios: %w", err)
	}

	// Inicializar adaptador
	h := &HookObservability{
		config:             config,
		complianceMetadata: make(map[string]ComplianceMetadata),
	}

	// Configurar componentes
	if err := h.setupLogger(); err != nil {
		return nil, fmt.Errorf("falha ao configurar logger: %w", err)
	}

	if err := h.setupTracer(); err != nil {
		h.logger.Warn("Falha ao configurar tracer, continuando sem tracing", zap.Error(err))
	}

	if err := h.setupMetrics(); err != nil {
		h.logger.Warn("Falha ao configurar métricas, continuando sem métricas", zap.Error(err))
	}

	h.logger.Info("Adaptador de observabilidade inicializado com sucesso",
		zap.String("service", config.ServiceName),
		zap.String("environment", config.Environment),
		zap.Bool("metrics_enabled", config.MetricsPort > 0),
		zap.Bool("tracing_enabled", config.OTLPEndpoint != ""),
	)

	return h, nil
}

// Close finaliza recursos do adaptador de observabilidade
func (h *HookObservability) Close() error {
	var errs []error

	// Fechar servidor de métricas se estiver ativo
	if h.metricsServer != nil {
		if err := h.metricsServer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("erro ao fechar servidor de métricas: %w", err))
		}
	}

	// Fechar provedor de traces se estiver ativo
	if h.tracerProvider != nil {
		if err := h.tracerProvider.Shutdown(context.Background()); err != nil {
			errs = append(errs, fmt.Errorf("erro ao fechar provedor de traces: %w", err))
		}
	}

	// Sincronizar logger para garantir que todos os logs foram escritos
	if h.logger != nil {
		if err := h.logger.Sync(); err != nil {
			// Ignorar erro específico do Zap que ocorre em alguns sistemas
			if err.Error() != "sync /dev/stderr: inappropriate ioctl for device" {
				errs = append(errs, fmt.Errorf("erro ao sincronizar logger: %w", err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("erros ao fechar adaptador de observabilidade: %v", errs)
	}

	return nil
}

// setupLogger configura o logger Zap
func (h *HookObservability) setupLogger() error {
	var cfg zap.Config

	// Definir nível de log
	level := zap.InfoLevel
	switch h.config.LogLevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	}

	// Configurar encoding conforme preferência
	if h.config.StructuredLogging {
		// JSON estruturado
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		// Logs no formato console
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Configurações comuns
	cfg.Level = zap.NewAtomicLevelAt(level)
	cfg.DisableCaller = false
	cfg.DisableStacktrace = false
	cfg.InitialFields = map[string]interface{}{
		"service":     h.config.ServiceName,
		"environment": h.config.Environment,
	}

	// Criar logger
	logger, err := cfg.Build()
	if err != nil {
		return fmt.Errorf("erro ao criar logger: %w", err)
	}

	h.logger = logger
	return nil
}

// setupTracer configura o tracer OpenTelemetry
func (h *HookObservability) setupTracer() error {
	// Se não houver endpoint OTLP configurado, desabilitar tracer
	if h.config.OTLPEndpoint == "" {
		h.logger.Info("Endpoint OTLP não configurado, tracing desativado")
		h.tracer = trace.NewNoopTracerProvider().Tracer("noop")
		return nil
	}

	// Criar recurso com informações do serviço
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(h.config.ServiceName),
			semconv.ServiceVersion("1.0.0"),
			semconv.DeploymentEnvironment(h.config.Environment),
		),
	)
	if err != nil {
		return fmt.Errorf("falha ao criar recurso de tracing: %w", err)
	}

	// Configurar exporter OTLP
	exporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithEndpoint(h.config.OTLPEndpoint),
		otlptracegrpc.WithInsecure(), // Remover em produção e usar TLS
	)
	if err != nil {
		return fmt.Errorf("falha ao criar exportador OTLP: %w", err)
	}

	// Criar provedor de trace com amostragem configurável
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(h.config.TraceSampleRate)),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Registrar o provedor de trace globalmente
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	h.tracerProvider = tp
	h.tracer = tp.Tracer("innovabiz.iam.hooks")

	return nil
}// setupMetrics configura métricas Prometheus e inicia servidor HTTP
func (h *HookObservability) setupMetrics() error {
	// Se a porta de métricas não estiver configurada, desabilitar métricas
	if h.config.MetricsPort <= 0 {
		h.logger.Info("Porta de métricas não configurada, métricas desativadas")
		return nil
	}

	// Criar registro Prometheus
	registry := prometheus.NewRegistry()
	h.metricsRegistry = registry

	// Registrar métricas padrão
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	registry.MustRegister(prometheus.NewGoCollector())

	// Contador de chamadas de hook
	h.hookCallsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "innovabiz_iam_hook_calls_total",
			Help: "Total de chamadas de hook MCP-IAM",
		},
		[]string{"market", "tenant_type", "hook_type", "operation"},
	)
	registry.MustRegister(h.hookCallsTotal)

	// Contador de erros de hook
	h.hookErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "innovabiz_iam_hook_errors_total",
			Help: "Total de erros em chamadas de hook MCP-IAM",
		},
		[]string{"market", "tenant_type", "hook_type", "operation"},
	)
	registry.MustRegister(h.hookErrorsTotal)

	// Histograma de duração de execução
	h.hookDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "innovabiz_iam_hook_duration_seconds",
			Help:    "Tempo de execução de hooks MCP-IAM em segundos",
			Buckets: []float64{0.001, 0.01, 0.05, 0.1, 0.5, 1, 2, 5},
		},
		[]string{"market", "tenant_type", "hook_type", "operation"},
	)
	registry.MustRegister(h.hookDurationSeconds)

	// Gauge para elevações de privilégio ativas
	h.hookActiveElevations = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "innovabiz_iam_active_elevations",
			Help: "Número de elevações de privilégio ativas por mercado e tipo de tenant",
		},
		[]string{"market", "tenant_type"},
	)
	registry.MustRegister(h.hookActiveElevations)

	// Contador de validações MFA
	h.mfaValidationTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "innovabiz_iam_mfa_validations_total",
			Help: "Total de validações MFA por mercado, tenant e resultado",
		},
		[]string{"market", "tenant_type", "level", "result"},
	)
	registry.MustRegister(h.mfaValidationTotal)

	// Contador de validações de escopo
	h.scopeValidationTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "innovabiz_iam_scope_validations_total",
			Help: "Total de validações de escopo por mercado, tenant e resultado",
		},
		[]string{"market", "tenant_type", "scope", "result"},
	)
	registry.MustRegister(h.scopeValidationTotal)

	// Gauge para cobertura de testes
	h.testCoveragePct = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "innovabiz_iam_test_coverage_percent",
			Help: "Percentual de cobertura de testes por tipo de hook",
		},
		[]string{"hook_type"},
	)
	registry.MustRegister(h.testCoveragePct)

	// Contador de eventos de compliance
	h.complianceEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "innovabiz_iam_compliance_events_total",
			Help: "Total de eventos de compliance por mercado e framework",
		},
		[]string{"market", "framework"},
	)
	registry.MustRegister(h.complianceEventsTotal)

	// Contador de eventos de segurança
	h.securityEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "innovabiz_iam_security_events_total",
			Help: "Total de eventos de segurança por mercado e severidade",
		},
		[]string{"market", "severity", "event_type"},
	)
	registry.MustRegister(h.securityEventsTotal)

	// Iniciar servidor HTTP para expor métricas
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", h.config.MetricsPort),
		Handler: http.DefaultServeMux,
	}

	h.metricsServer = server
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.logger.Error("Erro ao iniciar servidor de métricas",
				zap.Error(err),
				zap.Int("port", h.config.MetricsPort),
			)
		}
	}()

	h.logger.Info("Servidor de métricas Prometheus iniciado",
		zap.Int("port", h.config.MetricsPort),
	)

	return nil
}

// RegisterComplianceMetadata registra metadados de compliance para um mercado específico
func (h *HookObservability) RegisterComplianceMetadata(market, framework string, requiresDualApproval bool, mfaLevel string, retentionYears int) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	metadata := ComplianceMetadata{
		Framework:          framework,
		RequiresDualApproval: requiresDualApproval,
		MinimumMFALevel:    mfaLevel,
		LogRetentionYears:  retentionYears,
		Market:             market,
	}

	h.complianceMetadata[market] = metadata

	h.logger.Info("Metadados de compliance registrados",
		zap.String("market", market),
		zap.String("framework", framework),
		zap.Bool("requires_dual_approval", requiresDualApproval),
		zap.String("mfa_level", mfaLevel),
		zap.Int("retention_years", retentionYears),
	)
}

// GetComplianceMetadata obtém metadados de compliance para um mercado específico
func (h *HookObservability) GetComplianceMetadata(market string) (ComplianceMetadata, bool) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	metadata, exists := h.complianceMetadata[market]
	if !exists {
		// Tentar fallback para configuração global
		metadata, exists = h.complianceMetadata[constants.MarketGlobal]
	}
	
	return metadata, exists
}

// ObserveHookOperation observa uma operação genérica de hook com tracing
func (h *HookObservability) ObserveHookOperation(
	ctx context.Context,
	marketCtx MarketContext,
	operation string,
	userId string,
	description string,
	attrs []attribute.KeyValue,
	fn func(context.Context) error,
) error {
	startTime := time.Now()

	// Incrementar contador de chamadas
	h.hookCallsTotal.WithLabelValues(
		marketCtx.Market,
		marketCtx.TenantType,
		marketCtx.HookType,
		operation,
	).Inc()

	// Criar span para a operação
	ctx, span := h.tracer.Start(ctx, fmt.Sprintf("hook.%s", operation),
		trace.WithAttributes(
			attribute.String("market", marketCtx.Market),
			attribute.String("tenant_type", marketCtx.TenantType),
			attribute.String("hook_type", marketCtx.HookType),
			attribute.String("operation", operation),
			attribute.String("user_id", userId),
			attribute.String("description", description),
		),
	)
	// Adicionar atributos extras ao span
	span.SetAttributes(attrs...)
	defer span.End()

	// Logger contextualizado
	logger := h.logger.With(
		zap.String("market", marketCtx.Market),
		zap.String("tenant_type", marketCtx.TenantType),
		zap.String("hook_type", marketCtx.HookType),
		zap.String("operation", operation),
		zap.String("user_id", userId),
	)

	// Registrar início da operação
	logger.Debug("Iniciando operação de hook",
		zap.String("description", description),
	)

	// Executar função de operação
	err := fn(ctx)

	// Registrar tempo de execução
	duration := time.Since(startTime).Seconds()
	h.hookDurationSeconds.WithLabelValues(
		marketCtx.Market,
		marketCtx.TenantType,
		marketCtx.HookType,
		operation,
	).Observe(duration)

	// Registrar resultado
	if err != nil {
		// Incrementar contador de erros
		h.hookErrorsTotal.WithLabelValues(
			marketCtx.Market,
			marketCtx.TenantType,
			marketCtx.HookType,
			operation,
		).Inc()

		// Registrar erro no span
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		// Log de erro
		logger.Error("Falha na operação de hook",
			zap.Error(err),
			zap.Float64("duration_seconds", duration),
		)
		return err
	}

	// Marcar span como bem-sucedido
	span.SetStatus(codes.Ok, "")

	// Log de sucesso
	logger.Debug("Operação de hook concluída com sucesso",
		zap.Float64("duration_seconds", duration),
	)
	return nil
}

// ObserveValidateScope observa uma operação de validação de escopo
func (h *HookObservability) ObserveValidateScope(
	ctx context.Context,
	marketCtx MarketContext,
	userId string,
	scope string,
	fn func(context.Context) error,
) error {
	// Attributes específicos para validação de escopo
	attrs := []attribute.KeyValue{
		attribute.String("scope", scope),
	}

	// Executar observação genérica
	err := h.ObserveHookOperation(
		ctx,
		marketCtx,
		constants.OperationValidateScope,
		userId,
		fmt.Sprintf("Validação de escopo '%s'", scope),
		attrs,
		fn,
	)

	// Incrementar contador específico de validação de escopo
	result := "success"
	if err != nil {
		result = "failure"
	}

	h.scopeValidationTotal.WithLabelValues(
		marketCtx.Market,
		marketCtx.TenantType,
		scope,
		result,
	).Inc()

	return err
}

// ObserveValidateMFA observa uma operação de validação MFA
func (h *HookObservability) ObserveValidateMFA(
	ctx context.Context,
	marketCtx MarketContext,
	userId string,
	mfaLevel string,
	fn func(context.Context) error,
) error {
	// Verificar conformidade com requisitos de MFA do mercado
	if metadata, exists := h.GetComplianceMetadata(marketCtx.Market); exists {
		if metadata.MinimumMFALevel != "" && mfaLevel != "" {
			if !isMFALevelSufficient(mfaLevel, metadata.MinimumMFALevel) {
				h.logger.Warn("Nível MFA abaixo do exigido para compliance",
					zap.String("market", marketCtx.Market),
					zap.String("framework", metadata.Framework),
					zap.String("required_level", metadata.MinimumMFALevel),
					zap.String("provided_level", mfaLevel),
				)
			}
		}
	}

	// Attributes específicos para validação MFA
	attrs := []attribute.KeyValue{
		attribute.String("mfa_level", mfaLevel),
	}

	// Executar observação genérica
	err := h.ObserveHookOperation(
		ctx,
		marketCtx,
		constants.OperationValidateMFA,
		userId,
		fmt.Sprintf("Validação MFA nível '%s'", mfaLevel),
		attrs,
		fn,
	)

	// Incrementar contador específico de validação MFA
	result := "success"
	if err != nil {
		result = "failure"
	}

	h.mfaValidationTotal.WithLabelValues(
		marketCtx.Market,
		marketCtx.TenantType,
		mfaLevel,
		result,
	).Inc()

	return err
}

// TraceAuditEvent registra um evento de auditoria
func (h *HookObservability) TraceAuditEvent(
	ctx context.Context,
	marketCtx MarketContext,
	userId string,
	eventType string,
	details string,
) {
	// Criar span para evento de auditoria
	_, span := h.tracer.Start(ctx, "hook.audit_event",
		trace.WithAttributes(
			attribute.String("market", marketCtx.Market),
			attribute.String("tenant_type", marketCtx.TenantType),
			attribute.String("hook_type", marketCtx.HookType),
			attribute.String("user_id", userId),
			attribute.String("event_type", eventType),
			attribute.String("details", details),
			attribute.String("event_category", "audit"),
		),
	)
	defer span.End()

	// Registrar no log
	h.logger.Info("Evento de auditoria",
		zap.String("market", marketCtx.Market),
		zap.String("tenant_type", marketCtx.TenantType),
		zap.String("hook_type", marketCtx.HookType),
		zap.String("user_id", userId),
		zap.String("event_type", eventType),
		zap.String("details", details),
	)

	// Registrar log de compliance se habilitado
	if h.config.EnableComplianceAudit && h.config.ComplianceLogsPath != "" {
		h.logComplianceEvent(marketCtx.Market, "audit", userId, eventType, details)
	}

	// Se houver metadados de compliance para o mercado, incrementar contador específico
	if metadata, exists := h.GetComplianceMetadata(marketCtx.Market); exists {
		h.complianceEventsTotal.WithLabelValues(
			marketCtx.Market,
			metadata.Framework,
		).Inc()
	}
}// TraceSecurity registra um evento de segurança
func (h *HookObservability) TraceSecurity(
	ctx context.Context,
	marketCtx MarketContext,
	userId string,
	severity string,
	details string,
	eventType string,
) {
	// Criar span para evento de segurança
	_, span := h.tracer.Start(ctx, "hook.security_event",
		trace.WithAttributes(
			attribute.String("market", marketCtx.Market),
			attribute.String("tenant_type", marketCtx.TenantType),
			attribute.String("hook_type", marketCtx.HookType),
			attribute.String("user_id", userId),
			attribute.String("severity", severity),
			attribute.String("details", details),
			attribute.String("event_type", eventType),
			attribute.String("event_category", "security"),
		),
	)
	defer span.End()

	// Determinar nível de log com base na severidade
	var logFn func(string, ...zap.Field)
	switch severity {
	case constants.SeverityCritical:
		logFn = h.logger.Error
	case constants.SeverityHigh:
		logFn = h.logger.Error
	case constants.SeverityMedium:
		logFn = h.logger.Warn
	case constants.SeverityLow:
		logFn = h.logger.Info
	case constants.SeverityInfo:
		logFn = h.logger.Info
	default:
		logFn = h.logger.Info
	}

	// Registrar no log
	logFn("Evento de segurança",
		zap.String("market", marketCtx.Market),
		zap.String("tenant_type", marketCtx.TenantType),
		zap.String("hook_type", marketCtx.HookType),
		zap.String("user_id", userId),
		zap.String("severity", severity),
		zap.String("event_type", eventType),
		zap.String("details", details),
	)

	// Registrar log de compliance se habilitado
	if h.config.EnableComplianceAudit && h.config.ComplianceLogsPath != "" {
		h.logComplianceEvent(marketCtx.Market, "security", userId, eventType, details)
	}

	// Incrementar contador de eventos de segurança
	h.securityEventsTotal.WithLabelValues(
		marketCtx.Market,
		severity,
		eventType,
	).Inc()
}

// UpdateActiveElevations atualiza o número de elevações de privilégio ativas
func (h *HookObservability) UpdateActiveElevations(marketCtx MarketContext, count float64) {
	h.hookActiveElevations.WithLabelValues(
		marketCtx.Market,
		marketCtx.TenantType,
	).Set(count)

	h.logger.Debug("Elevações de privilégio ativas atualizadas",
		zap.String("market", marketCtx.Market),
		zap.String("tenant_type", marketCtx.TenantType),
		zap.Float64("count", count),
	)
}

// RecordTestCoverage registra a cobertura de testes para um tipo de hook
func (h *HookObservability) RecordTestCoverage(hookType string, coveragePct float64) {
	h.testCoveragePct.WithLabelValues(hookType).Set(coveragePct)

	h.logger.Info("Cobertura de testes registrada",
		zap.String("hook_type", hookType),
		zap.Float64("coverage_pct", coveragePct),
	)
}

// logComplianceEvent registra um evento de compliance em arquivo
func (h *HookObservability) logComplianceEvent(market, eventCategory, userId, eventType, details string) {
	// Criar diretório específico para o mercado se não existir
	marketDir := filepath.Join(h.config.ComplianceLogsPath, market)
	if err := os.MkdirAll(marketDir, 0755); err != nil {
		h.logger.Error("Falha ao criar diretório de logs de compliance",
			zap.String("market", market),
			zap.String("dir", marketDir),
			zap.Error(err),
		)
		return
	}

	// Criar nome do arquivo baseado na data
	date := time.Now().Format("2006-01-02")
	fileName := fmt.Sprintf("%s-%s-events.log", date, eventCategory)
	filePath := filepath.Join(marketDir, fileName)

	// Formatar evento
	timestamp := time.Now().Format(time.RFC3339)
	logLine := fmt.Sprintf("[%s] [%s] [%s] [%s] [%s]: %s\n",
		timestamp, market, eventCategory, userId, eventType, details)

	// Abrir arquivo em modo append
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		h.logger.Error("Falha ao abrir arquivo de log de compliance",
			zap.String("file", filePath),
			zap.Error(err),
		)
		return
	}
	defer f.Close()

	// Escrever evento
	if _, err := f.WriteString(logLine); err != nil {
		h.logger.Error("Falha ao escrever evento de compliance",
			zap.String("file", filePath),
			zap.Error(err),
		)
	}
}

// isMFALevelSufficient verifica se o nível MFA fornecido atende ao mínimo requerido
func isMFALevelSufficient(provided, required string) bool {
	// Mapear níveis MFA para valores numéricos
	levels := map[string]int{
		constants.MFALevelNone:  0,
		constants.MFALevelBasic: 1,
		constants.MFALevelMedium: 2,
		constants.MFALevelHigh:  3,
	}

	// Se o nível não estiver mapeado, assumir valor 0
	providedLevel, ok := levels[provided]
	if !ok {
		providedLevel = 0
	}

	requiredLevel, ok := levels[required]
	if !ok {
		requiredLevel = 0
	}

	// Verificar se o nível fornecido é pelo menos igual ao requerido
	return providedLevel >= requiredLevel
}