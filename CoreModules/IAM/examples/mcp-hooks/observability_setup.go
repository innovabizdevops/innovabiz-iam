// Package mcp_hooks fornece exemplos de configuração e inicialização
// da camada de observabilidade para hooks MCP-IAM da plataforma INNOVABIZ.
//
// Este exemplo demonstra como configurar e inicializar o adaptador de observabilidade
// para diferentes mercados, com configurações específicas que respeitam as regulações
// e requisitos locais.
//
// Conformidades: ISO/IEC 27001, ISO 20000, COBIT 2019, TOGAF 10.0, DMBOK 2.0,
// NIST SP 800-53, GDPR, LGPD, BNA, PIPL
package mcp_hooks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/innovabiz/iam/constants"
	"github.com/innovabiz/iam/observability/adapter"
)

// ObservabilityConfigurator configura a observabilidade para hooks MCP-IAM
type ObservabilityConfigurator struct {
	baseConfig adapter.Config
	market     string
	env        string
}

// NewObservabilityConfigurator cria um novo configurador de observabilidade
func NewObservabilityConfigurator(market, env string) *ObservabilityConfigurator {
	// Configuração base com valores padrão
	baseConfig := adapter.Config{
		Environment:           env,
		ServiceName:           "mcp-iam-hooks",
		OTLPEndpoint:          "localhost:4317", // Endpoint padrão para OpenTelemetry
		MetricsPort:           9090,             // Porta padrão para métricas Prometheus
		EnableComplianceAudit: true,
		StructuredLogging:     true,
		LogLevel:              "info",
	}

	return &ObservabilityConfigurator{
		baseConfig: baseConfig,
		market:     market,
		env:        env,
	}
}

// ConfigureForMarket configura a observabilidade específica para um mercado
func (oc *ObservabilityConfigurator) ConfigureForMarket() (*adapter.HookObservability, error) {
	// Clonar configuração base
	config := oc.baseConfig

	// Customizar configuração por mercado
	switch oc.market {
	case constants.MarketAngola:
		return oc.configureForAngola(config)
	case constants.MarketBrazil:
		return oc.configureForBrazil(config)
	case constants.MarketEU:
		return oc.configureForEU(config)
	case constants.MarketChina:
		return oc.configureForChina(config)
	case constants.MarketMozambique:
		return oc.configureForMozambique(config)
	case constants.MarketSADC:
		return oc.configureForSADC(config)
	case constants.MarketPALOP:
		return oc.configureForPALOP(config)
	case constants.MarketBRICS:
		return oc.configureForBRICS(config)
	case constants.MarketUSA:
		return oc.configureForUSA(config)
	default:
		return oc.configureForGlobal(config)
	}
}

// configureForAngola configura observabilidade específica para Angola
func (oc *ObservabilityConfigurator) configureForAngola(config adapter.Config) (*adapter.HookObservability, error) {
	// Customizar nome do serviço para incluir mercado
	config.ServiceName = fmt.Sprintf("%s-angola", config.ServiceName)
	
	// Configurar caminho para logs de compliance conforme regulação BNA
	if oc.env == constants.EnvProduction {
		config.ComplianceLogsPath = "/compliance/angola/bna"
		// BNA exige retenção mais longa de logs de compliance
		config.LogLevel = "info"
	} else {
		config.ComplianceLogsPath = filepath.Join(os.TempDir(), "compliance", "angola", "bna")
	}
	
	// Inicializar observabilidade
	obs, err := adapter.NewHookObservability(config)
	if err != nil {
		return nil, fmt.Errorf("falha ao configurar observabilidade para Angola: %w", err)
	}
	
	// Registrar metadados de compliance para Angola
	oc.registerAngolaComplianceMetadata(obs)
	
	return obs, nil
}

// configureForBrazil configura observabilidade específica para Brasil
func (oc *ObservabilityConfigurator) configureForBrazil(config adapter.Config) (*adapter.HookObservability, error) {
	// Customizar nome do serviço para incluir mercado
	config.ServiceName = fmt.Sprintf("%s-brazil", config.ServiceName)
	
	// Configurar caminho para logs de compliance conforme LGPD
	if oc.env == constants.EnvProduction {
		config.ComplianceLogsPath = "/compliance/brazil/lgpd"
	} else {
		config.ComplianceLogsPath = filepath.Join(os.TempDir(), "compliance", "brazil", "lgpd")
	}
	
	// LGPD exige logs detalhados de auditoria para operações sensíveis
	config.LogLevel = "debug"
	
	// Inicializar observabilidade
	obs, err := adapter.NewHookObservability(config)
	if err != nil {
		return nil, fmt.Errorf("falha ao configurar observabilidade para Brasil: %w", err)
	}
	
	// Registrar metadados de compliance para Brasil
	oc.registerBrazilComplianceMetadata(obs)
	
	return obs, nil
}

// configureForEU configura observabilidade específica para União Europeia
func (oc *ObservabilityConfigurator) configureForEU(config adapter.Config) (*adapter.HookObservability, error) {
	// Customizar nome do serviço para incluir mercado
	config.ServiceName = fmt.Sprintf("%s-eu", config.ServiceName)
	
	// Configurar caminho para logs de compliance conforme GDPR
	if oc.env == constants.EnvProduction {
		config.ComplianceLogsPath = "/compliance/eu/gdpr"
	} else {
		config.ComplianceLogsPath = filepath.Join(os.TempDir(), "compliance", "eu", "gdpr")
	}
	
	// GDPR exige logs detalhados de auditoria para operações de dados pessoais
	config.LogLevel = "debug"
	
	// Inicializar observabilidade
	obs, err := adapter.NewHookObservability(config)
	if err != nil {
		return nil, fmt.Errorf("falha ao configurar observabilidade para EU: %w", err)
	}
	
	// Registrar metadados de compliance para EU
	oc.registerEUComplianceMetadata(obs)
	
	return obs, nil
}

// configureForChina configura observabilidade específica para China
func (oc *ObservabilityConfigurator) configureForChina(config adapter.Config) (*adapter.HookObservability, error) {
	// Customizar nome do serviço para incluir mercado
	config.ServiceName = fmt.Sprintf("%s-china", config.ServiceName)
	
	// Configurar caminho para logs de compliance conforme PIPL
	if oc.env == constants.EnvProduction {
		config.ComplianceLogsPath = "/compliance/china/pipl"
	} else {
		config.ComplianceLogsPath = filepath.Join(os.TempDir(), "compliance", "china", "pipl")
	}
	
	// PIPL exige logs detalhados de auditoria para operações com dados
	config.LogLevel = "debug"
	
	// Inicializar observabilidade
	obs, err := adapter.NewHookObservability(config)
	if err != nil {
		return nil, fmt.Errorf("falha ao configurar observabilidade para China: %w", err)
	}
	
	// Registrar metadados de compliance para China
	oc.registerChinaComplianceMetadata(obs)
	
	return obs, nil
}

// configureForMozambique configura observabilidade específica para Moçambique
func (oc *ObservabilityConfigurator) configureForMozambique(config adapter.Config) (*adapter.HookObservability, error) {
	// Customizar nome do serviço para incluir mercado
	config.ServiceName = fmt.Sprintf("%s-mozambique", config.ServiceName)
	
	if oc.env == constants.EnvProduction {
		config.ComplianceLogsPath = "/compliance/mozambique"
	} else {
		config.ComplianceLogsPath = filepath.Join(os.TempDir(), "compliance", "mozambique")
	}
	
	// Inicializar observabilidade
	obs, err := adapter.NewHookObservability(config)
	if err != nil {
		return nil, fmt.Errorf("falha ao configurar observabilidade para Moçambique: %w", err)
	}
	
	// Registrar metadados de compliance para Moçambique
	oc.registerMozambiqueComplianceMetadata(obs)
	
	return obs, nil
}

// configureForSADC configura observabilidade específica para SADC
func (oc *ObservabilityConfigurator) configureForSADC(config adapter.Config) (*adapter.HookObservability, error) {
	// Customizar nome do serviço para incluir mercado
	config.ServiceName = fmt.Sprintf("%s-sadc", config.ServiceName)
	
	if oc.env == constants.EnvProduction {
		config.ComplianceLogsPath = "/compliance/sadc"
	} else {
		config.ComplianceLogsPath = filepath.Join(os.TempDir(), "compliance", "sadc")
	}
	
	// Inicializar observabilidade
	obs, err := adapter.NewHookObservability(config)
	if err != nil {
		return nil, fmt.Errorf("falha ao configurar observabilidade para SADC: %w", err)
	}
	
	// Registrar metadados de compliance para SADC
	oc.registerSADCComplianceMetadata(obs)
	
	return obs, nil
}

// configureForPALOP configura observabilidade específica para PALOP
func (oc *ObservabilityConfigurator) configureForPALOP(config adapter.Config) (*adapter.HookObservability, error) {
	// Customizar nome do serviço para incluir mercado
	config.ServiceName = fmt.Sprintf("%s-palop", config.ServiceName)
	
	if oc.env == constants.EnvProduction {
		config.ComplianceLogsPath = "/compliance/palop"
	} else {
		config.ComplianceLogsPath = filepath.Join(os.TempDir(), "compliance", "palop")
	}
	
	// Inicializar observabilidade
	obs, err := adapter.NewHookObservability(config)
	if err != nil {
		return nil, fmt.Errorf("falha ao configurar observabilidade para PALOP: %w", err)
	}
	
	// Registrar metadados de compliance para PALOP
	oc.registerPALOPComplianceMetadata(obs)
	
	return obs, nil
}

// configureForBRICS configura observabilidade específica para BRICS
func (oc *ObservabilityConfigurator) configureForBRICS(config adapter.Config) (*adapter.HookObservability, error) {
	// Customizar nome do serviço para incluir mercado
	config.ServiceName = fmt.Sprintf("%s-brics", config.ServiceName)
	
	if oc.env == constants.EnvProduction {
		config.ComplianceLogsPath = "/compliance/brics"
	} else {
		config.ComplianceLogsPath = filepath.Join(os.TempDir(), "compliance", "brics")
	}
	
	// Inicializar observabilidade
	obs, err := adapter.NewHookObservability(config)
	if err != nil {
		return nil, fmt.Errorf("falha ao configurar observabilidade para BRICS: %w", err)
	}
	
	// Registrar metadados de compliance para BRICS
	oc.registerBRICSComplianceMetadata(obs)
	
	return obs, nil
}

// configureForUSA configura observabilidade específica para EUA
func (oc *ObservabilityConfigurator) configureForUSA(config adapter.Config) (*adapter.HookObservability, error) {
	// Customizar nome do serviço para incluir mercado
	config.ServiceName = fmt.Sprintf("%s-usa", config.ServiceName)
	
	// Configurar caminho para logs de compliance conforme regulações americanas
	if oc.env == constants.EnvProduction {
		config.ComplianceLogsPath = "/compliance/usa/sox"
	} else {
		config.ComplianceLogsPath = filepath.Join(os.TempDir(), "compliance", "usa", "sox")
	}
	
	// SOX e outras regulações exigem logs detalhados
	config.LogLevel = "info"
	
	// Inicializar observabilidade
	obs, err := adapter.NewHookObservability(config)
	if err != nil {
		return nil, fmt.Errorf("falha ao configurar observabilidade para EUA: %w", err)
	}
	
	// Registrar metadados de compliance para EUA
	oc.registerUSAComplianceMetadata(obs)
	
	return obs, nil
}

// configureForGlobal configura observabilidade global (padrão)
func (oc *ObservabilityConfigurator) configureForGlobal(config adapter.Config) (*adapter.HookObservability, error) {
	// Customizar nome do serviço para incluir mercado
	config.ServiceName = fmt.Sprintf("%s-global", config.ServiceName)
	
	if oc.env == constants.EnvProduction {
		config.ComplianceLogsPath = "/compliance/global"
	} else {
		config.ComplianceLogsPath = filepath.Join(os.TempDir(), "compliance", "global")
	}
	
	// Inicializar observabilidade
	obs, err := adapter.NewHookObservability(config)
	if err != nil {
		return nil, fmt.Errorf("falha ao configurar observabilidade global: %w", err)
	}
	
	// Registrar metadados de compliance global
	oc.registerGlobalComplianceMetadata(obs)
	
	return obs, nil
}

// Funções de registro de metadados de compliance para cada mercado

// registerAngolaComplianceMetadata registra metadados de compliance específicos de Angola
func (oc *ObservabilityConfigurator) registerAngolaComplianceMetadata(obs *adapter.HookObservability) {
	// Registrar metadados de compliance para instituições financeiras em Angola
	obs.RegisterComplianceMetadata(
		constants.MarketAngola,
		"BNA",
		true, // Requer aprovação dual
		constants.MFALevelHigh,
		7, // 7 anos de retenção
	)
}

// registerBrazilComplianceMetadata registra metadados de compliance específicos do Brasil
func (oc *ObservabilityConfigurator) registerBrazilComplianceMetadata(obs *adapter.HookObservability) {
	// LGPD
	obs.RegisterComplianceMetadata(
		constants.MarketBrazil,
		"LGPD",
		true, // Requer aprovação dual
		constants.MFALevelHigh,
		5, // 5 anos de retenção
	)
	
	// BACEN para instituições financeiras
	obs.RegisterComplianceMetadata(
		constants.MarketBrazil,
		"BACEN",
		true, // Requer aprovação dual
		constants.MFALevelHigh,
		10, // 10 anos de retenção
	)
}

// registerEUComplianceMetadata registra metadados de compliance específicos da UE
func (oc *ObservabilityConfigurator) registerEUComplianceMetadata(obs *adapter.HookObservability) {
	// GDPR
	obs.RegisterComplianceMetadata(
		constants.MarketEU,
		"GDPR",
		true, // Requer aprovação dual
		constants.MFALevelHigh,
		7, // 7 anos de retenção
	)
	
	// PSD2 para instituições financeiras
	obs.RegisterComplianceMetadata(
		constants.MarketEU,
		"PSD2",
		true, // Requer aprovação dual
		constants.MFALevelHigh,
		10, // 10 anos de retenção
	)
}

// registerChinaComplianceMetadata registra metadados de compliance específicos da China
func (oc *ObservabilityConfigurator) registerChinaComplianceMetadata(obs *adapter.HookObservability) {
	// PIPL
	obs.RegisterComplianceMetadata(
		constants.MarketChina,
		"PIPL",
		true, // Requer aprovação dual
		constants.MFALevelHigh,
		5, // 5 anos de retenção
	)
}

// registerMozambiqueComplianceMetadata registra metadados de compliance específicos de Moçambique
func (oc *ObservabilityConfigurator) registerMozambiqueComplianceMetadata(obs *adapter.HookObservability) {
	obs.RegisterComplianceMetadata(
		constants.MarketMozambique,
		"Banco_Mocambique",
		true, // Requer aprovação dual
		constants.MFALevelMedium,
		5, // 5 anos de retenção
	)
}

// registerSADCComplianceMetadata registra metadados de compliance específicos da SADC
func (oc *ObservabilityConfigurator) registerSADCComplianceMetadata(obs *adapter.HookObservability) {
	obs.RegisterComplianceMetadata(
		constants.MarketSADC,
		"SADC_Framework",
		false, // Não requer aprovação dual por padrão
		constants.MFALevelBasic,
		3, // 3 anos de retenção
	)
}

// registerPALOPComplianceMetadata registra metadados de compliance específicos dos PALOP
func (oc *ObservabilityConfigurator) registerPALOPComplianceMetadata(obs *adapter.HookObservability) {
	obs.RegisterComplianceMetadata(
		constants.MarketPALOP,
		"PALOP_Framework",
		false, // Não requer aprovação dual por padrão
		constants.MFALevelBasic,
		3, // 3 anos de retenção
	)
}

// registerBRICSComplianceMetadata registra metadados de compliance específicos dos BRICS
func (oc *ObservabilityConfigurator) registerBRICSComplianceMetadata(obs *adapter.HookObservability) {
	obs.RegisterComplianceMetadata(
		constants.MarketBRICS,
		"BRICS_Framework",
		false, // Não requer aprovação dual por padrão
		constants.MFALevelMedium,
		5, // 5 anos de retenção
	)
}

// registerUSAComplianceMetadata registra metadados de compliance específicos dos EUA
func (oc *ObservabilityConfigurator) registerUSAComplianceMetadata(obs *adapter.HookObservability) {
	// SOX para empresas de capital aberto
	obs.RegisterComplianceMetadata(
		constants.MarketUSA,
		"SOX",
		true, // Requer aprovação dual
		constants.MFALevelHigh,
		7, // 7 anos de retenção
	)
	
	// HIPAA para instituições de saúde
	obs.RegisterComplianceMetadata(
		constants.MarketUSA,
		"HIPAA",
		true, // Requer aprovação dual
		constants.MFALevelHigh,
		6, // 6 anos de retenção
	)
}

// registerGlobalComplianceMetadata registra metadados de compliance globais
func (oc *ObservabilityConfigurator) registerGlobalComplianceMetadata(obs *adapter.HookObservability) {
	obs.RegisterComplianceMetadata(
		constants.MarketGlobal,
		"ISO27001",
		false, // Não requer aprovação dual por padrão
		constants.MFALevelMedium,
		3, // 3 anos de retenção
	)
}

// SetupObservabilityExample demonstra como configurar e usar observabilidade para hooks MCP-IAM
func SetupObservabilityExample() {
	// Configurar observabilidade para Angola em ambiente de produção
	angolaConfigurator := NewObservabilityConfigurator(constants.MarketAngola, constants.EnvProduction)
	angolaObs, err := angolaConfigurator.ConfigureForMarket()
	if err != nil {
		fmt.Printf("Erro ao configurar observabilidade para Angola: %v\n", err)
		return
	}
	
	// Configurar observabilidade para Brasil em ambiente de desenvolvimento
	brazilConfigurator := NewObservabilityConfigurator(constants.MarketBrazil, constants.EnvDevelopment)
	brazilObs, err := brazilConfigurator.ConfigureForMarket()
	if err != nil {
		fmt.Printf("Erro ao configurar observabilidade para Brasil: %v\n", err)
		return
	}
	
	// Criar contexto
	ctx := context.Background()
	
	// Criar adaptadores de mercado para Angola (tenant financeiro)
	angolaMarketCtx := adapter.NewMarketContext(constants.MarketAngola, constants.TenantFinancial, constants.HookTypePrivilegeElevation)
	
	// Registrar evento de auditoria para Angola
	angolaObs.TraceAuditEvent(
		ctx,
		angolaMarketCtx,
		"user123",
		"system_startup",
		"Sistema inicializado com configurações específicas de Angola",
	)
	
	// Criar adaptadores de mercado para Brasil (tenant de saúde)
	brazilMarketCtx := adapter.NewMarketContext(constants.MarketBrazil, constants.TenantHealthcare, constants.HookTypePrivilegeElevation)
	
	// Registrar evento de auditoria para Brasil
	brazilObs.TraceAuditEvent(
		ctx,
		brazilMarketCtx,
		"user456",
		"system_startup",
		"Sistema inicializado com configurações específicas do Brasil",
	)
	
	// Exemplo de uso em validação de escopo para Angola
	_ = angolaObs.ObserveValidateScope(
		ctx,
		angolaMarketCtx,
		"user123",
		"admin:read",
		func(ctx context.Context) error {
			// Simulação de validação de escopo
			fmt.Println("Validando escopo para Angola...")
			return nil
		},
	)
	
	// Exemplo de uso em validação MFA para Brasil
	_ = brazilObs.ObserveValidateMFA(
		ctx,
		brazilMarketCtx,
		"user456",
		constants.MFALevelHigh,
		func(ctx context.Context) error {
			// Simulação de validação MFA
			fmt.Println("Validando MFA para Brasil...")
			return nil
		},
	)
}