// Package adapter fornece configurações para o adaptador de observabilidade MCP-IAM
//
// Este arquivo define as estruturas de configuração e contexto utilizadas pelo adaptador de
// observabilidade para hooks MCP-IAM, suportando múltiplos mercados, tenants e tipos de hooks
// com requisitos específicos de compliance.
//
// Conformidades: ISO/IEC 27001, ISO 20000, COBIT 2019, TOGAF 10.0, DMBOK 2.0
package adapter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config define as configurações do adaptador de observabilidade
type Config struct {
	// Ambiente de execução (development, test, staging, production)
	Environment string

	// Nome do serviço para identificação em logs e traces
	ServiceName string

	// Endpoint para exportação OpenTelemetry (ex: localhost:4317)
	OTLPEndpoint string

	// Porta para expor métricas Prometheus (0 para desativar)
	MetricsPort int

	// Caminho para logs de compliance
	ComplianceLogsPath string

	// Ativar auditoria de compliance
	EnableComplianceAudit bool

	// Usar logging estruturado (formato JSON)
	StructuredLogging bool

	// Nível de log (debug, info, warn, error)
	LogLevel string

	// Taxa de amostragem para traces (0.0-1.0)
	TraceSampleRate float64
}

// MarketContext encapsula informações de contexto de mercado para observabilidade
type MarketContext struct {
	// Mercado (ex: Angola, Brazil, EU, USA, Global)
	Market string

	// Tipo de tenant (ex: Financial, Retail, Healthcare)
	TenantType string

	// Tipo de hook (ex: PrivilegeElevation, MFAValidation)
	HookType string

	// Metadados adicionais específicos do contexto
	Metadata map[string]string
}

// ComplianceMetadata define metadados de compliance específicos por mercado
type ComplianceMetadata struct {
	// Framework de compliance aplicável (ex: GDPR, LGPD, PCI-DSS)
	Framework string

	// Se requer aprovação dual para operações sensíveis
	RequiresDualApproval bool

	// Nível mínimo de MFA requerido
	MinimumMFALevel string

	// Anos de retenção de logs de auditoria
	LogRetentionYears int

	// Mercado associado
	Market string
}

// Validate valida a configuração do adaptador
func (c *Config) Validate() error {
	// Validar ambiente
	if c.Environment == "" {
		return fmt.Errorf("ambiente não configurado")
	}
	validEnvironments := []string{"development", "test", "staging", "production"}
	validEnv := false
	for _, env := range validEnvironments {
		if strings.EqualFold(c.Environment, env) {
			validEnv = true
			break
		}
	}
	if !validEnv {
		return fmt.Errorf("ambiente inválido: %s", c.Environment)
	}

	// Validar nome do serviço
	if c.ServiceName == "" {
		return fmt.Errorf("nome do serviço não configurado")
	}

	// Validar porta de métricas se configurada
	if c.MetricsPort < 0 || c.MetricsPort > 65535 {
		return fmt.Errorf("porta de métricas inválida: %d", c.MetricsPort)
	}

	// Validar nível de log
	if c.LogLevel != "" {
		validLogLevels := []string{"debug", "info", "warn", "error"}
		validLevel := false
		for _, level := range validLogLevels {
			if strings.EqualFold(c.LogLevel, level) {
				validLevel = true
				break
			}
		}
		if !validLevel {
			return fmt.Errorf("nível de log inválido: %s", c.LogLevel)
		}
	} else {
		// Definir valor padrão
		c.LogLevel = "info"
	}

	// Validar taxa de amostragem
	if c.TraceSampleRate < 0 || c.TraceSampleRate > 1.0 {
		return fmt.Errorf("taxa de amostragem inválida: %f, deve estar entre 0.0 e 1.0", c.TraceSampleRate)
	}
	if c.TraceSampleRate == 0 {
		// Definir valor padrão
		c.TraceSampleRate = 1.0
	}

	return nil
}

// EnsureDirectories garante que os diretórios necessários existam
func (c *Config) EnsureDirectories() error {
	// Verificar se o caminho de logs de compliance está configurado
	if c.ComplianceLogsPath == "" {
		return nil
	}

	// Criar diretório de logs de compliance se não existir
	if err := os.MkdirAll(c.ComplianceLogsPath, 0755); err != nil {
		return fmt.Errorf("falha ao criar diretório de logs de compliance: %w", err)
	}

	return nil
}

// WithEnvironment define o ambiente de execução
func (c *Config) WithEnvironment(env string) *Config {
	c.Environment = env
	return c
}

// WithServiceName define o nome do serviço
func (c *Config) WithServiceName(name string) *Config {
	c.ServiceName = name
	return c
}

// WithOTLPEndpoint define o endpoint para exportação OpenTelemetry
func (c *Config) WithOTLPEndpoint(endpoint string) *Config {
	c.OTLPEndpoint = endpoint
	return c
}

// WithMetricsPort define a porta para métricas Prometheus
func (c *Config) WithMetricsPort(port int) *Config {
	c.MetricsPort = port
	return c
}

// WithComplianceLogsPath define o caminho para logs de compliance
func (c *Config) WithComplianceLogsPath(path string) *Config {
	c.ComplianceLogsPath = path
	return c
}

// WithComplianceAudit ativa ou desativa auditoria de compliance
func (c *Config) WithComplianceAudit(enable bool) *Config {
	c.EnableComplianceAudit = enable
	return c
}

// WithStructuredLogging ativa ou desativa logging estruturado
func (c *Config) WithStructuredLogging(enable bool) *Config {
	c.StructuredLogging = enable
	return c
}

// WithLogLevel define o nível de log
func (c *Config) WithLogLevel(level string) *Config {
	c.LogLevel = level
	return c
}

// WithTraceSampleRate define a taxa de amostragem para traces
func (c *Config) WithTraceSampleRate(rate float64) *Config {
	c.TraceSampleRate = rate
	return c
}

// NewMarketContext cria um novo contexto de mercado para observabilidade
func NewMarketContext(market, tenantType, hookType string) MarketContext {
	return MarketContext{
		Market:     market,
		TenantType: tenantType,
		HookType:   hookType,
		Metadata:   make(map[string]string),
	}
}

// WithMetadata adiciona metadados ao contexto de mercado
func (mc MarketContext) WithMetadata(key, value string) MarketContext {
	mc.Metadata[key] = value
	return mc
}

// GetMetadata obtém um valor de metadado do contexto de mercado
func (mc MarketContext) GetMetadata(key string) (string, bool) {
	value, exists := mc.Metadata[key]
	return value, exists
}