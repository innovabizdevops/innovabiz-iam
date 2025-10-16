// Implementação do Logger usando Zap - INNOVABIZ Platform
// Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
// PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III, PSD2, AML/KYC
package logging

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Constantes para chaves de contexto
const (
	RequestIDKey     = "request_id"
	CorrelationIDKey = "correlation_id"
	SessionIDKey     = "session_id"
	TenantIDKey      = "tenant_id"
	UserIDKey        = "user_id"
)

// ZapLogger implementa a interface Logger usando a biblioteca Zap
type ZapLogger struct {
	logger     *zap.Logger
	auditLogger *zap.Logger
	metadata   LogMetadata
	config     LoggerConfig
}

// NewZapLogger cria uma nova instância do ZapLogger
func NewZapLogger(config LoggerConfig) (Logger, error) {
	// Configuração do logger principal
	zapConfig := zap.NewProductionConfig()
	
	// Definir o nível de log
	switch config.Level {
	case DebugLevel:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case InfoLevel:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case WarnLevel:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case ErrorLevel:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case FatalLevel:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}
	
	// Configurar formato de saída
	switch config.Format {
	case JSONFormat:
		zapConfig.Encoding = "json"
	case TextFormat, PrettyFormat:
		zapConfig.Encoding = "console"
	default:
		zapConfig.Encoding = "json"
	}
	
	// Configurar campos comuns
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	
	// Criar cores de saída
	var cores []zapcore.Core
	
	// Adicionar saída para console se habilitado
	if config.EnableConsole {
		consoleEncoder := zapcore.NewConsoleEncoder(zapConfig.EncoderConfig)
		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			zapConfig.Level,
		)
		cores = append(cores, consoleCore)
	}
	
	// Adicionar saída para arquivo se habilitado
	if config.EnableFile && config.FilePath != "" {
		// Criar diretório se não existir
		logDir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, err
		}
		
		// Configurar rotação de logs
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.FilePath,
			MaxSize:    config.MaxSize,    // megabytes
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,     // dias
			Compress:   config.Compress,
		})
		
		fileEncoder := zapcore.NewJSONEncoder(zapConfig.EncoderConfig)
		fileCore := zapcore.NewCore(
			fileEncoder,
			fileWriter,
			zapConfig.Level,
		)
		cores = append(cores, fileCore)
	}
	
	// Criar o logger principal
	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	
	// Adicionar metadados comuns
	logger = logger.With(
		zap.String("service", config.Metadata.ServiceName),
		zap.String("version", config.Metadata.ServiceVersion),
		zap.String("environment", config.Metadata.Environment),
		zap.String("host", config.Metadata.HostName),
	)
	
	if config.Metadata.PodName != "" {
		logger = logger.With(zap.String("pod", config.Metadata.PodName))
	}
	
	if config.Metadata.NodeName != "" {
		logger = logger.With(zap.String("node", config.Metadata.NodeName))
	}
	
	if config.Metadata.ClusterName != "" {
		logger = logger.With(zap.String("cluster", config.Metadata.ClusterName))
	}
	
	// Configurar logger de auditoria
	var auditLogger *zap.Logger
	if config.EnableAuditLogging {
		auditLogPath := config.AuditLogPath
		if auditLogPath == "" {
			auditLogPath = filepath.Join(filepath.Dir(config.FilePath), "audit.log")
		}
		
		// Criar diretório se não existir
		auditLogDir := filepath.Dir(auditLogPath)
		if err := os.MkdirAll(auditLogDir, 0755); err != nil {
			return nil, err
		}
		
		// Configurar rotação de logs de auditoria
		auditWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   auditLogPath,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		})
		
		auditEncoder := zapcore.NewJSONEncoder(zapConfig.EncoderConfig)
		auditCore := zapcore.NewCore(
			auditEncoder,
			auditWriter,
			zap.NewAtomicLevelAt(zapcore.InfoLevel), // Log de auditoria sempre no nível INFO ou superior
		)
		
		auditLogger = zap.New(auditCore)
	} else {
		auditLogger = logger
	}
	
	return &ZapLogger{
		logger:      logger,
		auditLogger: auditLogger,
		metadata:    config.Metadata,
		config:      config,
	}, nil
}

// Debug implementa o método Debug da interface Logger
func (l *ZapLogger) Debug(msg string, keyvals ...interface{}) {
	fields := parseKeyvals(keyvals...)
	l.logger.Debug(msg, fields...)
}

// Info implementa o método Info da interface Logger
func (l *ZapLogger) Info(msg string, keyvals ...interface{}) {
	fields := parseKeyvals(keyvals...)
	l.logger.Info(msg, fields...)
}

// Warn implementa o método Warn da interface Logger
func (l *ZapLogger) Warn(msg string, keyvals ...interface{}) {
	fields := parseKeyvals(keyvals...)
	l.logger.Warn(msg, fields...)
}

// Error implementa o método Error da interface Logger
func (l *ZapLogger) Error(msg string, keyvals ...interface{}) {
	fields := parseKeyvals(keyvals...)
	l.logger.Error(msg, fields...)
}

// Fatal implementa o método Fatal da interface Logger
func (l *ZapLogger) Fatal(msg string, keyvals ...interface{}) {
	fields := parseKeyvals(keyvals...)
	l.logger.Fatal(msg, fields...)
}

// DebugContext implementa o método DebugContext da interface Logger
func (l *ZapLogger) DebugContext(ctx context.Context, msg string, keyvals ...interface{}) {
	fields := parseKeyvals(keyvals...)
	fields = append(fields, extractContextFields(ctx)...)
	l.logger.Debug(msg, fields...)
}

// InfoContext implementa o método InfoContext da interface Logger
func (l *ZapLogger) InfoContext(ctx context.Context, msg string, keyvals ...interface{}) {
	fields := parseKeyvals(keyvals...)
	fields = append(fields, extractContextFields(ctx)...)
	l.logger.Info(msg, fields...)
}

// WarnContext implementa o método WarnContext da interface Logger
func (l *ZapLogger) WarnContext(ctx context.Context, msg string, keyvals ...interface{}) {
	fields := parseKeyvals(keyvals...)
	fields = append(fields, extractContextFields(ctx)...)
	l.logger.Warn(msg, fields...)
}

// ErrorContext implementa o método ErrorContext da interface Logger
func (l *ZapLogger) ErrorContext(ctx context.Context, msg string, keyvals ...interface{}) {
	fields := parseKeyvals(keyvals...)
	fields = append(fields, extractContextFields(ctx)...)
	l.logger.Error(msg, fields...)
}

// FatalContext implementa o método FatalContext da interface Logger
func (l *ZapLogger) FatalContext(ctx context.Context, msg string, keyvals ...interface{}) {
	fields := parseKeyvals(keyvals...)
	fields = append(fields, extractContextFields(ctx)...)
	l.logger.Fatal(msg, fields...)
}

// AuditLog implementa o método AuditLog da interface Logger
func (l *ZapLogger) AuditLog(event AuditEvent) {
	// Converter o evento para JSON para logging
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		l.Error("failed to marshal audit event", "error", err)
		return
	}
	
	// Converter JSON para mapa para facilitar o log
	var eventMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &eventMap); err != nil {
		l.Error("failed to unmarshal audit event", "error", err)
		return
	}
	
	// Converter mapa para campos Zap
	fields := []zap.Field{
		zap.String("log_type", "audit"),
	}
	
	for k, v := range eventMap {
		fields = append(fields, zap.Any(k, v))
	}
	
	// Registrar evento de auditoria
	l.auditLogger.Info("AUDIT", fields...)
}

// SecurityLog implementa o método SecurityLog da interface Logger
func (l *ZapLogger) SecurityLog(event SecurityEvent) {
	// Converter o evento para JSON para logging
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		l.Error("failed to marshal security event", "error", err)
		return
	}
	
	// Converter JSON para mapa para facilitar o log
	var eventMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &eventMap); err != nil {
		l.Error("failed to unmarshal security event", "error", err)
		return
	}
	
	// Converter mapa para campos Zap
	fields := []zap.Field{
		zap.String("log_type", "security"),
	}
	
	for k, v := range eventMap {
		fields = append(fields, zap.Any(k, v))
	}
	
	// Registrar evento de segurança com nível apropriado
	switch event.Severity {
	case "CRITICAL":
		l.auditLogger.Error("SECURITY", fields...)
	case "HIGH":
		l.auditLogger.Error("SECURITY", fields...)
	case "MEDIUM":
		l.auditLogger.Warn("SECURITY", fields...)
	case "LOW":
		l.auditLogger.Info("SECURITY", fields...)
	default:
		l.auditLogger.Info("SECURITY", fields...)
	}
}

// Função auxiliar para extrair campos do contexto
func extractContextFields(ctx context.Context) []zap.Field {
	var fields []zap.Field
	
	// Extrair campos comuns do contexto
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}
	
	if correlationID, ok := ctx.Value(CorrelationIDKey).(string); ok && correlationID != "" {
		fields = append(fields, zap.String("correlation_id", correlationID))
	}
	
	if sessionID, ok := ctx.Value(SessionIDKey).(string); ok && sessionID != "" {
		fields = append(fields, zap.String("session_id", sessionID))
	}
	
	if tenantID, ok := ctx.Value(TenantIDKey).(string); ok && tenantID != "" {
		fields = append(fields, zap.String("tenant_id", tenantID))
	}
	
	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}
	
	return fields
}

// Função auxiliar para converter pares chave-valor em campos Zap
func parseKeyvals(keyvals ...interface{}) []zap.Field {
	if len(keyvals)%2 != 0 {
		return []zap.Field{zap.Any("UNPAIRED_KEY", keyvals[len(keyvals)-1])}
	}
	
	fields := make([]zap.Field, 0, len(keyvals)/2)
	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			key = "INVALID_KEY"
		}
		
		fields = append(fields, zap.Any(key, keyvals[i+1]))
	}
	
	return fields
}