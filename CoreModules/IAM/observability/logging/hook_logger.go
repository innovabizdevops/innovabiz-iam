// Package logging fornece instrumentação para logging estruturado dos hooks MCP-IAM
//
// Este pacote implementa logging estruturado para os hooks MCP-IAM da plataforma INNOVABIZ,
// usando Zap como framework de logging. Suporta dimensões multi-mercado, multi-tenant e 
// multi-contexto conforme requisitos da plataforma, além de integração com tracing
// distribuído para correlação de eventos.
//
// Conformidades: ISO/IEC 27001, ISO 20000, COBIT 2019, NIST, TOGAF 10.0, DMBOK 2.0
// Frameworks: OpenTelemetry, OWASP, ISO 27018, LGPD, GDPR
package logging

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// HookLogger encapsula funcionalidades de logging para hooks MCP-IAM
type HookLogger struct {
	logger *zap.Logger
	env    string
}

// Níveis de log específicos para compliance
const (
	LogLevelAudit    = "AUDIT"
	LogLevelSecurity = "SECURITY"
	LogLevelCompliance = "COMPLIANCE"
)

// NewHookLogger cria uma nova instância do HookLogger
func NewHookLogger(env string) *HookLogger {
	// Configuração base
	config := zap.NewProductionConfig()
	
	// Adicionar campos default
	config.InitialFields = map[string]interface{}{
		"service":     "innovabiz-iam",
		"component":   "hooks",
		"environment": env,
	}
	
	// Configurar encoding para incluir campos necessários para compliance
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	config.EncoderConfig.StacktraceKey = "stacktrace"
	
	// Criar logger
	logger, _ := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	
	return &HookLogger{
		logger: logger,
		env:    env,
	}
}

// NewDevelopmentLogger cria um logger específico para ambiente de desenvolvimento
func NewDevelopmentLogger() *HookLogger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := config.Build()
	
	return &HookLogger{
		logger: logger,
		env:    "development",
	}
}

// NewComplianceLogger cria um logger específico para compliance
func NewComplianceLogger(env, logPath string) (*HookLogger, error) {
	// Criar encoder para logs de compliance
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	
	// Configurar saída para arquivo e console
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	
	// Abrir arquivo de log para compliance (append only)
	logFile, err := os.OpenFile(
		logPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return nil, err
	}
	
	// Configurar core para múltiplas saídas
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), zap.InfoLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zap.InfoLevel),
	)
	
	// Criar logger com fields padrão
	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(
			zap.String("service", "innovabiz-iam"),
			zap.String("component", "hooks"),
			zap.String("environment", env),
		),
	)
	
	return &HookLogger{
		logger: logger,
		env:    env,
	}, nil
}

// WithContext adiciona informações de contexto ao logger
func (hl *HookLogger) WithContext(ctx context.Context) *zap.Logger {
	// Extrair informações de trace do contexto
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return hl.logger.With(
			zap.String("trace_id", span.SpanContext().TraceID().String()),
			zap.String("span_id", span.SpanContext().SpanID().String()),
		)
	}
	return hl.logger
}

// LogHookOperation registra operação de hook com detalhes
func (hl *HookLogger) LogHookOperation(
	ctx context.Context,
	level zapcore.Level,
	market string,
	tenantType string,
	hookType string,
	operation string,
	userId string,
	message string,
	fields ...zap.Field,
) {
	// Preparar campos padrão
	baseFields := []zap.Field{
		zap.String("market", market),
		zap.String("tenant_type", tenantType),
		zap.String("hook_type", hookType),
		zap.String("operation", operation),
		zap.String("user_id", userId),
		zap.String("timestamp", time.Now().Format(time.RFC3339)),
	}
	
	// Adicionar campos extras
	allFields := append(baseFields, fields...)
	
	// Obter logger com contexto
	logger := hl.WithContext(ctx)
	
	// Registrar log com nível apropriado
	switch level {
	case zapcore.DebugLevel:
		logger.Debug(message, allFields...)
	case zapcore.InfoLevel:
		logger.Info(message, allFields...)
	case zapcore.WarnLevel:
		logger.Warn(message, allFields...)
	case zapcore.ErrorLevel:
		logger.Error(message, allFields...)
	case zapcore.DPanicLevel:
		logger.DPanic(message, allFields...)
	case zapcore.PanicLevel:
		logger.Panic(message, allFields...)
	case zapcore.FatalLevel:
		logger.Fatal(message, allFields...)
	}
}

// LogAuditEvent registra evento de auditoria específico para compliance
func (hl *HookLogger) LogAuditEvent(
	ctx context.Context,
	market string,
	tenantType string,
	hookType string,
	operation string,
	userId string,
	eventType string,
	eventDetails string,
	fields ...zap.Field,
) {
	// Adicionar campos específicos de auditoria
	auditFields := []zap.Field{
		zap.String("log_type", LogLevelAudit),
		zap.String("event_type", eventType),
		zap.String("event_details", eventDetails),
		zap.String("audit_timestamp", time.Now().Format(time.RFC3339)),
	}
	
	// Combinar campos
	allFields := append(auditFields, fields...)
	
	// Registrar evento de auditoria
	hl.LogHookOperation(
		ctx,
		zapcore.InfoLevel,
		market,
		tenantType,
		hookType,
		operation,
		userId,
		"Evento de auditoria",
		allFields...,
	)
}

// LogSecurityEvent registra evento de segurança
func (hl *HookLogger) LogSecurityEvent(
	ctx context.Context,
	market string,
	tenantType string,
	hookType string,
	operation string,
	userId string,
	severity string,
	eventDetails string,
	fields ...zap.Field,
) {
	// Adicionar campos específicos de segurança
	securityFields := []zap.Field{
		zap.String("log_type", LogLevelSecurity),
		zap.String("severity", severity),
		zap.String("event_details", eventDetails),
	}
	
	// Combinar campos
	allFields := append(securityFields, fields...)
	
	// Determinar nível de log baseado na severidade
	var level zapcore.Level
	switch severity {
	case "critical", "high":
		level = zapcore.ErrorLevel
	case "medium":
		level = zapcore.WarnLevel
	default:
		level = zapcore.InfoLevel
	}
	
	// Registrar evento de segurança
	hl.LogHookOperation(
		ctx,
		level,
		market,
		tenantType,
		hookType,
		operation,
		userId,
		"Evento de segurança",
		allFields...,
	)
}

// LogComplianceEvent registra evento relacionado a conformidade regulatória
func (hl *HookLogger) LogComplianceEvent(
	ctx context.Context,
	market string,
	tenantType string,
	hookType string,
	operation string,
	userId string,
	regulation string,
	status string,
	details string,
	fields ...zap.Field,
) {
	// Adicionar campos específicos de compliance
	complianceFields := []zap.Field{
		zap.String("log_type", LogLevelCompliance),
		zap.String("regulation", regulation),
		zap.String("compliance_status", status),
		zap.String("details", details),
	}
	
	// Combinar campos
	allFields := append(complianceFields, fields...)
	
	// Registrar evento de compliance
	hl.LogHookOperation(
		ctx,
		zapcore.InfoLevel,
		market,
		tenantType,
		hookType,
		operation,
		userId,
		"Verificação de compliance",
		allFields...,
	)
}

// LogHookError registra erro em operação de hook com detalhes específicos
func (hl *HookLogger) LogHookError(
	ctx context.Context,
	market string,
	tenantType string,
	hookType string,
	operation string,
	userId string,
	err error,
	fields ...zap.Field,
) {
	// Adicionar campos específicos de erro
	errorFields := []zap.Field{
		zap.Error(err),
		zap.String("error_time", time.Now().Format(time.RFC3339)),
	}
	
	// Combinar campos
	allFields := append(errorFields, fields...)
	
	// Registrar erro
	hl.LogHookOperation(
		ctx,
		zapcore.ErrorLevel,
		market,
		tenantType,
		hookType,
		operation,
		userId,
		"Erro na operação de hook",
		allFields...,
	)
}

// Sync garante que todos os logs sejam persistidos
func (hl *HookLogger) Sync() error {
	return hl.logger.Sync()
}

// MarsketSpecificLogger retorna um logger com configurações específicas para o mercado
func (hl *HookLogger) MarketSpecificLogger(market string) *zap.Logger {
	// Configurar regras específicas de log por mercado
	var retentionDays int
	var complianceFields []zap.Field
	
	switch market {
	case "angola":
		retentionDays = 7 * 365 // 7 anos (BNA)
		complianceFields = []zap.Field{
			zap.String("compliance_framework", "BNA,ARSSI"),
			zap.Bool("dual_approval_required", true),
		}
	case "brasil":
		retentionDays = 5 * 365 // 5 anos (LGPD)
		complianceFields = []zap.Field{
			zap.String("compliance_framework", "LGPD,BACEN"),
			zap.Bool("data_subject_consent_required", true),
		}
	case "eu":
		retentionDays = 2 * 365 // 2 anos (GDPR)
		complianceFields = []zap.Field{
			zap.String("compliance_framework", "GDPR,EBA"),
			zap.Bool("data_minimization_applied", true),
		}
	case "china":
		retentionDays = 3 * 365 // 3 anos (PIPL)
		complianceFields = []zap.Field{
			zap.String("compliance_framework", "PIPL,CSL"),
			zap.Bool("data_localization_required", true),
		}
	case "mocambique":
		retentionDays = 5 * 365 // 5 anos (LPDP)
		complianceFields = []zap.Field{
			zap.String("compliance_framework", "LPDP,BM"),
			zap.Bool("sadc_compliance", true),
		}
	default:
		retentionDays = 7 * 365 // Padrão global conservador
		complianceFields = []zap.Field{
			zap.String("compliance_framework", "Global"),
		}
	}
	
	// Retornar logger com configurações específicas do mercado
	return hl.logger.With(
		append(
			complianceFields,
			zap.Int("retention_days", retentionDays),
			zap.String("market", market),
		)...,
	)
}