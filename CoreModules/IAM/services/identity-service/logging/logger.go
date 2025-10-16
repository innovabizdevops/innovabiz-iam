// Interface de logging para o serviço de identidade - INNOVABIZ Platform
// Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
// PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III, PSD2, AML/KYC
package logging

import (
	"context"
	"time"
)

// Logger define a interface para logging utilizada pelo serviço de identidade
type Logger interface {
	// Métodos principais de logging
	Debug(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Warn(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})
	Fatal(msg string, keyvals ...interface{})

	// Métodos com contexto para rastreabilidade
	DebugContext(ctx context.Context, msg string, keyvals ...interface{})
	InfoContext(ctx context.Context, msg string, keyvals ...interface{})
	WarnContext(ctx context.Context, msg string, keyvals ...interface{})
	ErrorContext(ctx context.Context, msg string, keyvals ...interface{})
	FatalContext(ctx context.Context, msg string, keyvals ...interface{})

	// Logs de auditoria específicos
	AuditLog(event AuditEvent)
	SecurityLog(event SecurityEvent)
}

// LogLevel define os níveis de log suportados
type LogLevel int

// Constantes para níveis de log
const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// AuditEvent representa um evento de auditoria
type AuditEvent struct {
	EventID        string                 `json:"event_id"`
	EventType      string                 `json:"event_type"`
	EventCategory  string                 `json:"event_category"`
	TenantID       string                 `json:"tenant_id"`
	UserID         string                 `json:"user_id,omitempty"`
	Username       string                 `json:"username,omitempty"`
	ResourceType   string                 `json:"resource_type"`
	ResourceID     string                 `json:"resource_id,omitempty"`
	Action         string                 `json:"action"`
	Status         string                 `json:"status"`
	StatusCode     int                    `json:"status_code,omitempty"`
	Description    string                 `json:"description"`
	Timestamp      time.Time              `json:"timestamp"`
	ClientIP       string                 `json:"client_ip,omitempty"`
	UserAgent      string                 `json:"user_agent,omitempty"`
	RequestID      string                 `json:"request_id,omitempty"`
	SessionID      string                 `json:"session_id,omitempty"`
	CorrelationID  string                 `json:"correlation_id,omitempty"`
	Source         string                 `json:"source,omitempty"`
	Target         string                 `json:"target,omitempty"`
	Details        map[string]interface{} `json:"details,omitempty"`
	OldValue       string                 `json:"old_value,omitempty"`
	NewValue       string                 `json:"new_value,omitempty"`
	ComplianceRefs []string               `json:"compliance_refs,omitempty"`
}

// SecurityEvent representa um evento de segurança
type SecurityEvent struct {
	EventID        string                 `json:"event_id"`
	EventType      string                 `json:"event_type"` // AUTH_FAILURE, ACCESS_DENIED, POLICY_VIOLATION, SUSPICIOUS_ACTIVITY, etc.
	Severity       string                 `json:"severity"`   // LOW, MEDIUM, HIGH, CRITICAL
	TenantID       string                 `json:"tenant_id"`
	UserID         string                 `json:"user_id,omitempty"`
	Username       string                 `json:"username,omitempty"`
	ResourceType   string                 `json:"resource_type,omitempty"`
	ResourceID     string                 `json:"resource_id,omitempty"`
	Action         string                 `json:"action,omitempty"`
	Status         string                 `json:"status"`
	StatusCode     int                    `json:"status_code,omitempty"`
	Description    string                 `json:"description"`
	Timestamp      time.Time              `json:"timestamp"`
	ClientIP       string                 `json:"client_ip,omitempty"`
	UserAgent      string                 `json:"user_agent,omitempty"`
	RequestID      string                 `json:"request_id,omitempty"`
	SessionID      string                 `json:"session_id,omitempty"`
	CorrelationID  string                 `json:"correlation_id,omitempty"`
	Source         string                 `json:"source,omitempty"`
	Details        map[string]interface{} `json:"details,omitempty"`
	MitigationSteps []string              `json:"mitigation_steps,omitempty"`
	RiskScore      float64                `json:"risk_score,omitempty"`
	ComplianceRefs []string               `json:"compliance_refs,omitempty"`
}

// LogMetadata contém metadados comuns para todos os logs
type LogMetadata struct {
	ServiceName    string    `json:"service_name"`
	ServiceVersion string    `json:"service_version"`
	Environment    string    `json:"environment"`
	HostName       string    `json:"host_name"`
	PodName        string    `json:"pod_name,omitempty"`
	NodeName       string    `json:"node_name,omitempty"`
	ClusterName    string    `json:"cluster_name,omitempty"`
	Region         string    `json:"region,omitempty"`
	DataCenter     string    `json:"data_center,omitempty"`
}

// LogFormat define o formato de saída para logs
type LogFormat string

// Constantes para formato de logs
const (
	JSONFormat   LogFormat = "json"
	TextFormat   LogFormat = "text"
	PrettyFormat LogFormat = "pretty"
)

// LoggerConfig contém a configuração para o logger
type LoggerConfig struct {
	Level              LogLevel    `json:"level"`
	Format             LogFormat   `json:"format"`
	OutputPaths        []string    `json:"output_paths"`
	ErrorOutputPaths   []string    `json:"error_output_paths"`
	Metadata           LogMetadata `json:"metadata"`
	EnableConsole      bool        `json:"enable_console"`
	EnableFile         bool        `json:"enable_file"`
	FilePath           string      `json:"file_path,omitempty"`
	MaxSize            int         `json:"max_size,omitempty"`    // megabytes
	MaxBackups         int         `json:"max_backups,omitempty"` // número de arquivos de backup
	MaxAge             int         `json:"max_age,omitempty"`     // dias
	Compress           bool        `json:"compress,omitempty"`    // comprimir logs antigos
	EnableAuditLogging bool        `json:"enable_audit_logging"`
	AuditLogPath       string      `json:"audit_log_path,omitempty"`
	AuditLogSeparate   bool        `json:"audit_log_separate"`
	EnableMetrics      bool        `json:"enable_metrics"`
	EnableTracing      bool        `json:"enable_tracing"`
}

// NewLogger cria uma nova instância do logger baseado na configuração
func NewLogger(config LoggerConfig) (Logger, error) {
	// Implementação a ser fornecida pela implementação concreta
	// (Zap, Logrus, etc.)
	return nil, nil
}