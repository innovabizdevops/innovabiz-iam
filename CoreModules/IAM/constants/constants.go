// Package constants fornece as constantes padronizadas para a plataforma MCP-IAM
// do sistema INNOVABIZ, incluindo definições para hooks, mercados, tenants,
// operações, níveis MFA, severidades e outros valores utilizados em todo o sistema.
//
// As constantes seguem padrões e recomendações internacionais de segurança,
// incluindo ISO/IEC 27001, NIST, GDPR, LGPD e outras regulações relevantes.
package constants

// Ambientes
const (
	EnvironmentDevelopment = "development"
	EnvironmentStaging     = "staging"
	EnvironmentProduction  = "production"
	EnvironmentTest        = "test"
	EnvironmentSandbox     = "sandbox"
)

// Mercados suportados pela plataforma (regras de compliance específicas)
const (
	MarketAngola  = "Angola"
	MarketBrazil  = "Brazil"
	MarketEU      = "EU"
	MarketUSA     = "USA"
	MarketChina   = "China"
	MarketSADC    = "SADC"
	MarketGlobal  = "Global" // Configuração padrão quando não houver específica
)

// Tipos de tenant suportados pela plataforma
const (
	TenantFinancial    = "Financial"
	TenantRetail       = "Retail"
	TenantGovernment   = "Government"
	TenantHealthcare   = "Healthcare"
	TenantTelecom      = "Telecom"
	TenantEnergy       = "Energy"
	TenantEducation    = "Education"
	TenantManufacturing = "Manufacturing"
)

// Tipos de hooks MCP-IAM
const (
	HookTypePrivilegeElevation   = "PrivilegeElevation"
	HookTypeMFAValidation        = "MFAValidation"
	HookTypeScopeValidation      = "ScopeValidation"
	HookTypeTokenIntrospection   = "TokenIntrospection"
	HookTypeUserAuthentication   = "UserAuthentication"
	HookTypeAuditLogging         = "AuditLogging"
	HookTypeRateLimiting         = "RateLimiting"
	HookTypeSecurityAlert        = "SecurityAlert"
)

// Operações comuns de hooks
const (
	OperationValidateScope      = "validate_scope"
	OperationValidateMFA        = "validate_mfa"
	OperationValidateToken      = "validate_token"
	OperationElevatePrivilege   = "elevate_privilege"
	OperationRevokePrivilege    = "revoke_privilege"
	OperationAuthorizeAccess    = "authorize_access"
	OperationTraceActivity      = "trace_activity"
	OperationEnrollMFA          = "enroll_mfa"
	OperationResetMFA           = "reset_mfa"
	OperationLimitRate          = "limit_rate"
)

// Níveis MFA em ordem crescente de segurança
const (
	MFALevelNone    = "none"     // Sem autenticação multi-fator
	MFALevelBasic   = "basic"    // SMS, Email
	MFALevelMedium  = "medium"   // TOTP (Google Authenticator)
	MFALevelHigh    = "high"     // Biometria, tokens físicos
	MFALevelAdvanced = "advanced" // Múltiplos fatores combinados
)

// Níveis de severidade para eventos de segurança (conforme ISO 27001 e NIST)
const (
	SeverityInfo     = "info"     // Informativo, sem impacto de segurança
	SeverityLow      = "low"      // Baixo impacto, sem urgência
	SeverityMedium   = "medium"   // Impacto moderado, requer atenção
	SeverityHigh     = "high"     // Alto impacto, requer ação imediata
	SeverityCritical = "critical" // Impacto crítico, emergência de segurança
)

// Níveis de log
const (
	LogLevelDebug    = "debug"
	LogLevelInfo     = "info"
	LogLevelWarning  = "warning"
	LogLevelError    = "error"
	LogLevelCritical = "critical"
)

// Frameworks de compliance suportados por mercado
const (
	// Frameworks globais
	FrameworkISO27001 = "ISO27001"
	FrameworkSOX      = "SOX"
	FrameworkCOBIT    = "COBIT"
	FrameworkPCIDSS   = "PCIDSS"
	FrameworkHIPAA    = "HIPAA"
	
	// Frameworks regionais
	FrameworkGDPR     = "GDPR"    // União Europeia
	FrameworkLGPD     = "LGPD"    // Brasil
	FrameworkCCSPA    = "CCSPA"   // China
	FrameworkCCPA     = "CCPA"    // Califórnia, EUA
	FrameworkBNA      = "BNA"     // Banco Nacional de Angola
	FrameworkBACEN    = "BACEN"   // Banco Central do Brasil
)

// Tipos de eventos de auditoria (conforme categorização NIST SP 800-92)
const (
	AuditEventAuthentication = "authentication" // Login, logout, alteração de credenciais
	AuditEventAuthorization  = "authorization"  // Acesso a recursos, validação de escopo
	AuditEventDataAccess     = "data_access"    // Leitura, modificação de dados
	AuditEventSystemChange   = "system_change"  // Alterações de configuração, políticas
	AuditEventPrivileged     = "privileged"     // Ações de administradores, elevação de privilégio
	AuditEventCompliance     = "compliance"     // Eventos relacionados a conformidade
	AuditEventSecurity       = "security"       // Alertas, violações, tentativas de invasão
)

// Flags da CLI para facilitar uso
const (
	FlagEnvironment          = "environment"
	FlagServiceName          = "service-name"
	FlagMarket               = "market"
	FlagTenantType           = "tenant-type"
	FlagHookType             = "hook-type"
	FlagOTLPEndpoint         = "otlp-endpoint"
	FlagMetricsPort          = "metrics-port"
	FlagLogsPath             = "logs-path"
	FlagStructuredLogging    = "structured-logging"
	FlagLogLevel             = "log-level"
	FlagSampleRate           = "sample-rate"
	FlagCount                = "count"
	FlagDelay                = "delay"
	FlagSimulateErrors       = "simulate-errors"
)

// Padrões de configuração para facilitar uso
const (
	DefaultServiceName   = "mcp-iam-hooks"
	DefaultEnvironment   = EnvironmentDevelopment
	DefaultMarket        = MarketGlobal
	DefaultTenantType    = TenantFinancial
	DefaultHookType      = HookTypePrivilegeElevation
	DefaultOTLPEndpoint  = "localhost:4317"
	DefaultMetricsPort   = 9090
	DefaultLogLevel      = LogLevelInfo
	DefaultSampleRate    = 1.0  // 100%
	DefaultSimulationCount = 10
	DefaultDelay         = 100  // ms
)