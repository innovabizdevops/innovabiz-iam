// Package main fornece uma interface de linha de comando para gerenciar
// a observabilidade dos hooks MCP-IAM da plataforma INNOVABIZ.
//
// Esta CLI permite configurar, testar e monitorar a instrumentação
// de hooks MCP-IAM em diferentes mercados, com suporte a métricas Prometheus,
// tracing OpenTelemetry e logging estruturado via Zap.
//
// Conformidades: ISO/IEC 27001, ISO 20000, COBIT 2019, TOGAF 10.0, DMBOK 2.0
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/innovabiz/iam/constants"
	"github.com/innovabiz/iam/observability/adapter"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

var (
	// Configurações globais
	cfgEnvironment      string
	cfgServiceName      string
	cfgOTLPEndpoint     string
	cfgMetricsPort      int
	cfgComplianceLogsPath string
	cfgLogLevel         string
	cfgStructuredLogging bool
	cfgMarket           string
	cfgTenantType       string
	cfgHookType         string

	// Flags para simulações
	simulateError       bool
	simulationCount     int
	simulationDelay     int
)

// rootCmd representa o comando base da aplicação
var rootCmd = &cobra.Command{
	Use:   "observability-cli",
	Short: "CLI para gerenciar a observabilidade de hooks MCP-IAM",
	Long: `Ferramenta de linha de comando para configurar, testar e monitorar
a observabilidade de hooks MCP-IAM da plataforma INNOVABIZ.

Suporta configurações específicas por mercado, tenant e tipo de hook,
com integração a métricas Prometheus, tracing OpenTelemetry e logging 
estruturado via Zap.`,
}

// configCmd representa o comando para gerenciar configurações
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Gerenciar configurações de observabilidade",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// configShowCmd exibe a configuração atual
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Exibir configuração atual",
	Run: func(cmd *cobra.Command, args []string) {
		config := buildConfig()
		
		// Converter para JSON formatado
		jsonBytes, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			fmt.Printf("Erro ao formatar configuração: %v\n", err)
			return
		}
		
		color.Cyan("Configuração atual:")
		fmt.Println(string(jsonBytes))
	},
}

// configValidateCmd valida a configuração atual
var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validar configuração atual",
	Run: func(cmd *cobra.Command, args []string) {
		config := buildConfig()
		
		if err := config.Validate(); err != nil {
			color.Red("Configuração inválida: %v", err)
			os.Exit(1)
		}
		
		color.Green("✓ Configuração válida")
		
		if err := config.EnsureDirectories(); err != nil {
			color.Yellow("⚠ Aviso na criação de diretórios: %v", err)
		} else {
			color.Green("✓ Diretórios verificados")
		}
	},
}

// testCmd representa o comando para testar a observabilidade
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Testar funcionalidades de observabilidade",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// testHookOperationsCmd simula operações de hook para testar observabilidade
var testHookOperationsCmd = &cobra.Command{
	Use:   "hook-operations",
	Short: "Simular operações de hook para testar observabilidade",
	Run: func(cmd *cobra.Command, args []string) {
		config := buildConfig()
		
		if err := config.Validate(); err != nil {
			color.Red("Configuração inválida: %v", err)
			os.Exit(1)
		}
		
		color.Cyan("Inicializando adaptador de observabilidade...")
		obs, err := adapter.NewHookObservability(config)
		if err != nil {
			color.Red("Erro ao inicializar adaptador: %v", err)
			os.Exit(1)
		}
		
		// Criar contexto de mercado
		marketCtx := adapter.NewMarketContext(cfgMarket, cfgTenantType, cfgHookType)
		
		// Registrar metadados de compliance conforme mercado
		registerMarketComplianceMetadata(obs, cfgMarket)
		
		color.Cyan("Executando %d simulações de operações com delay de %dms...", simulationCount, simulationDelay)
		
		userId := fmt.Sprintf("test-user-%s", time.Now().Format("20060102150405"))
		ctx := context.Background()
		
		// Executar simulações
		for i := 1; i <= simulationCount; i++ {
			color.Yellow("Simulação %d/%d", i, simulationCount)
			
			// Simular validação de escopo
			testValidateScope(ctx, obs, marketCtx, userId, i)
			time.Sleep(time.Duration(simulationDelay) * time.Millisecond)
			
			// Simular validação MFA
			testValidateMFA(ctx, obs, marketCtx, userId, i)
			time.Sleep(time.Duration(simulationDelay) * time.Millisecond)
			
			// Simular evento de auditoria
			testAuditEvent(ctx, obs, marketCtx, userId, i)
			time.Sleep(time.Duration(simulationDelay) * time.Millisecond)
			
			// Simular evento de segurança
			testSecurityEvent(ctx, obs, marketCtx, userId, i)
			time.Sleep(time.Duration(simulationDelay) * time.Millisecond)
		}
		
		color.Green("✓ %d simulações concluídas com sucesso", simulationCount)
		
		// Se estiver usando métricas, exibir instruções
		if config.MetricsPort > 0 {
			color.Cyan("\nMétricas Prometheus disponíveis em: http://localhost:%d/metrics", config.MetricsPort)
		}
		
		// Exibir caminho dos logs
		if config.ComplianceLogsPath != "" {
			color.Cyan("Logs de compliance disponíveis em: %s", config.ComplianceLogsPath)
		}
	},
}

// testTraceExportCmd testa a exportação de traces
var testTraceExportCmd = &cobra.Command{
	Use:   "trace-export",
	Short: "Testar exportação de traces OpenTelemetry",
	Run: func(cmd *cobra.Command, args []string) {
		config := buildConfig()
		
		if err := config.Validate(); err != nil {
			color.Red("Configuração inválida: %v", err)
			os.Exit(1)
		}
		
		if config.OTLPEndpoint == "" {
			color.Red("Endpoint OTLP não configurado")
			os.Exit(1)
		}
		
		color.Cyan("Inicializando adaptador de observabilidade...")
		obs, err := adapter.NewHookObservability(config)
		if err != nil {
			color.Red("Erro ao inicializar adaptador: %v", err)
			os.Exit(1)
		}
		
		// Criar contexto de mercado
		marketCtx := adapter.NewMarketContext(cfgMarket, cfgTenantType, cfgHookType)
		
		color.Cyan("Enviando traces para %s...", config.OTLPEndpoint)
		ctx := context.Background()
		userId := fmt.Sprintf("test-user-%s", time.Now().Format("20060102150405"))
		
		// Criar um trace principal
		err = obs.ObserveHookOperation(
			ctx,
			marketCtx,
			constants.OperationValidateScope,
			userId,
			"Validação de escopo para teste de exportação",
			[]attribute.KeyValue{
				attribute.String("test_type", "export"),
				attribute.String("market", marketCtx.Market),
				attribute.String("tenant_type", marketCtx.TenantType),
			},
			func(ctx context.Context) error {
				// Simular trabalho com sub-spans
				for i := 1; i <= 5; i++ {
					attrName := fmt.Sprintf("sub-operation-%d", i)
					err := obs.ObserveHookOperation(
						ctx,
						marketCtx,
						attrName,
						userId,
						fmt.Sprintf("Sub-operação %d", i),
						[]attribute.KeyValue{
							attribute.Int("sub_op_id", i),
						},
						func(ctx context.Context) error {
							// Simular trabalho
							time.Sleep(100 * time.Millisecond)
							return nil
						},
					)
					if err != nil {
						return err
					}
				}
				return nil
			},
		)
		
		if err != nil {
			color.Red("Erro ao enviar traces: %v", err)
			os.Exit(1)
		}
		
		color.Green("✓ Traces enviados com sucesso para %s", config.OTLPEndpoint)
		color.Cyan("Verifique seu coletor OpenTelemetry para visualizar os traces")
	},
}

// metricsCmd representa o comando para gerenciar métricas
var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Gerenciar métricas de observabilidade",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// metricsExposeCmd expõe métricas em um servidor HTTP
var metricsExposeCmd = &cobra.Command{
	Use:   "expose",
	Short: "Expor métricas em servidor HTTP",
	Run: func(cmd *cobra.Command, args []string) {
		config := buildConfig()
		
		if config.MetricsPort <= 0 {
			color.Red("Porta de métricas inválida")
			os.Exit(1)
		}
		
		color.Cyan("Inicializando adaptador de observabilidade...")
		obs, err := adapter.NewHookObservability(config)
		if err != nil {
			color.Red("Erro ao inicializar adaptador: %v", err)
			os.Exit(1)
		}
		
		// Simular algumas métricas
		marketCtx := adapter.NewMarketContext(cfgMarket, cfgTenantType, cfgHookType)
		obs.UpdateActiveElevations(marketCtx, 5)
		obs.RecordTestCoverage(cfgHookType, 95.5)
		
		color.Green("✓ Servidor de métricas iniciado na porta %d", config.MetricsPort)
		color.Cyan("Acesse http://localhost:%d/metrics no navegador", config.MetricsPort)
		color.Cyan("Pressione Ctrl+C para encerrar")
		
		// Aguardar indefinidamente (servidor HTTP roda em goroutine separada)
		select {}
	},
}

// Funções auxiliares

// buildConfig cria uma configuração a partir das flags
func buildConfig() adapter.Config {
	config := adapter.Config{
		Environment:           cfgEnvironment,
		ServiceName:           cfgServiceName,
		OTLPEndpoint:          cfgOTLPEndpoint,
		MetricsPort:           cfgMetricsPort,
		ComplianceLogsPath:    cfgComplianceLogsPath,
		EnableComplianceAudit: true,
		StructuredLogging:     cfgStructuredLogging,
		LogLevel:              cfgLogLevel,
	}
	return config
}

// testValidateScope testa a validação de escopo
func testValidateScope(ctx context.Context, obs *adapter.HookObservability, marketCtx adapter.MarketContext, userId string, iteration int) {
	scope := "admin:read"
	if iteration%3 == 0 { // A cada 3 iterações, usar um escopo diferente
		scope = "system:admin"
	}
	
	color.Cyan("Validando escopo '%s' para usuário '%s'...", scope, userId)
	
	err := obs.ObserveValidateScope(
		ctx,
		marketCtx,
		userId,
		scope,
		func(ctx context.Context) error {
			// Simular trabalho
			time.Sleep(50 * time.Millisecond)
			
			// Simular erro conforme configuração
			if simulateError && iteration%3 == 0 {
				return fmt.Errorf("erro simulado: escopo '%s' não autorizado", scope)
			}
			return nil
		},
	)
	
	if err != nil {
		color.Red("✗ Validação de escopo falhou: %v", err)
	} else {
		color.Green("✓ Validação de escopo concluída com sucesso")
	}
}

// testValidateMFA testa a validação MFA
func testValidateMFA(ctx context.Context, obs *adapter.HookObservability, marketCtx adapter.MarketContext, userId string, iteration int) {
	// Determinar nível MFA com base na iteração
	var mfaLevel string
	switch iteration % 3 {
	case 0:
		mfaLevel = constants.MFALevelBasic
	case 1:
		mfaLevel = constants.MFALevelMedium
	case 2:
		mfaLevel = constants.MFALevelHigh
	}
	
	color.Cyan("Validando MFA nível '%s' para usuário '%s'...", mfaLevel, userId)
	
	err := obs.ObserveValidateMFA(
		ctx,
		marketCtx,
		userId,
		mfaLevel,
		func(ctx context.Context) error {
			// Simular trabalho
			time.Sleep(100 * time.Millisecond)
			
			// Simular erro conforme configuração
			if simulateError && iteration%4 == 0 {
				return fmt.Errorf("erro simulado: MFA nível '%s' falhou", mfaLevel)
			}
			return nil
		},
	)
	
	if err != nil {
		color.Red("✗ Validação MFA falhou: %v", err)
	} else {
		color.Green("✓ Validação MFA concluída com sucesso")
	}
}

// testAuditEvent testa o registro de eventos de auditoria
func testAuditEvent(ctx context.Context, obs *adapter.HookObservability, marketCtx adapter.MarketContext, userId string, iteration int) {
	// Determinar tipo de evento com base na iteração
	eventTypes := []string{"login", "privilege_elevation", "role_change", "permission_grant"}
	eventType := eventTypes[iteration%len(eventTypes)]
	
	eventDetails := fmt.Sprintf("Evento de auditoria '%s' para usuário '%s' (iteração %d)", 
		eventType, userId, iteration)
	
	color.Cyan("Registrando evento de auditoria: %s", eventType)
	
	obs.TraceAuditEvent(
		ctx,
		marketCtx,
		userId,
		eventType,
		eventDetails,
	)
	
	color.Green("✓ Evento de auditoria registrado")
}

// testSecurityEvent testa o registro de eventos de segurança
func testSecurityEvent(ctx context.Context, obs *adapter.HookObservability, marketCtx adapter.MarketContext, userId string, iteration int) {
	// Determinar severidade com base na iteração
	var severity string
	switch iteration % 5 {
	case 0:
		severity = constants.SeverityCritical
	case 1:
		severity = constants.SeverityHigh
	case 2:
		severity = constants.SeverityMedium
	case 3:
		severity = constants.SeverityLow
	case 4:
		severity = constants.SeverityInfo
	}
	
	eventType := "security_check"
	details := fmt.Sprintf("Evento de segurança com severidade '%s' para usuário '%s'", 
		severity, userId)
	
	color.Cyan("Registrando evento de segurança: %s (severidade %s)", eventType, severity)
	
	obs.TraceSecurity(
		ctx,
		marketCtx,
		userId,
		severity,
		details,
		eventType,
	)
	
	color.Green("✓ Evento de segurança registrado")
}

// registerMarketComplianceMetadata registra metadados de compliance específicos do mercado
func registerMarketComplianceMetadata(obs *adapter.HookObservability, market string) {
	switch market {
	case constants.MarketAngola:
		obs.RegisterComplianceMetadata(
			constants.MarketAngola,
			"BNA",
			true, // Requer aprovação dual
			constants.MFALevelHigh,
			7, // 7 anos de retenção
		)
	case constants.MarketBrazil:
		obs.RegisterComplianceMetadata(
			constants.MarketBrazil,
			"LGPD",
			true, // Requer aprovação dual
			constants.MFALevelHigh,
			5, // 5 anos de retenção
		)
		obs.RegisterComplianceMetadata(
			constants.MarketBrazil,
			"BACEN",
			true, // Requer aprovação dual
			constants.MFALevelHigh,
			10, // 10 anos de retenção
		)
	case constants.MarketEU:
		obs.RegisterComplianceMetadata(
			constants.MarketEU,
			"GDPR",
			true, // Requer aprovação dual
			constants.MFALevelHigh,
			7, // 7 anos de retenção
		)
	case constants.MarketUSA:
		obs.RegisterComplianceMetadata(
			constants.MarketUSA,
			"SOX",
			true, // Requer aprovação dual
			constants.MFALevelMedium,
			7, // 7 anos de retenção
		)
	default:
		obs.RegisterComplianceMetadata(
			constants.MarketGlobal,
			"ISO27001",
			false, // Não requer aprovação dual por padrão
			constants.MFALevelMedium,
			3, // 3 anos de retenção
		)
	}
}

func init() {
	// Configuração de diretório de logs padrão baseado no diretório do usuário
	userConfigDir, err := os.UserConfigDir()
	var defaultLogsPath string
	if err != nil {
		defaultLogsPath = filepath.Join(os.TempDir(), "innovabiz", "compliance")
	} else {
		defaultLogsPath = filepath.Join(userConfigDir, "innovabiz", "compliance")
	}

	// Flags globais da aplicação
	rootCmd.PersistentFlags().StringVar(&cfgEnvironment, "environment", "development", "Ambiente (development, test, staging, production)")
	rootCmd.PersistentFlags().StringVar(&cfgServiceName, "service-name", "mcp-iam-hooks-cli", "Nome do serviço")
	rootCmd.PersistentFlags().StringVar(&cfgOTLPEndpoint, "otlp-endpoint", "", "Endpoint para exportação OpenTelemetry (ex: localhost:4317)")
	rootCmd.PersistentFlags().IntVar(&cfgMetricsPort, "metrics-port", 9090, "Porta para métricas Prometheus")
	rootCmd.PersistentFlags().StringVar(&cfgComplianceLogsPath, "logs-path", defaultLogsPath, "Caminho para logs de compliance")
	rootCmd.PersistentFlags().StringVar(&cfgLogLevel, "log-level", "info", "Nível de log (debug, info, warn, error)")
	rootCmd.PersistentFlags().BoolVar(&cfgStructuredLogging, "structured-logging", true, "Usar logs estruturados (formato JSON)")
	rootCmd.PersistentFlags().StringVar(&cfgMarket, "market", constants.MarketGlobal, fmt.Sprintf("Mercado (%s, %s, %s, etc)", constants.MarketAngola, constants.MarketBrazil, constants.MarketEU))
	rootCmd.PersistentFlags().StringVar(&cfgTenantType, "tenant-type", constants.TenantFinancial, fmt.Sprintf("Tipo de tenant (%s, %s, %s, etc)", constants.TenantFinancial, constants.TenantRetail, constants.TenantHealthcare))
	rootCmd.PersistentFlags().StringVar(&cfgHookType, "hook-type", constants.HookTypePrivilegeElevation, fmt.Sprintf("Tipo de hook (%s, %s, etc)", constants.HookTypePrivilegeElevation, constants.HookTypeMFAValidation))

	// Flags específicas dos comandos de teste
	testHookOperationsCmd.Flags().BoolVar(&simulateError, "simulate-error", false, "Simular erros nas operações")
	testHookOperationsCmd.Flags().IntVar(&simulationCount, "count", 5, "Número de simulações a executar")
	testHookOperationsCmd.Flags().IntVar(&simulationDelay, "delay", 200, "Delay entre simulações (ms)")

	// Estrutura de comandos
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configValidateCmd)

	rootCmd.AddCommand(testCmd)
	testCmd.AddCommand(testHookOperationsCmd)
	testCmd.AddCommand(testTraceExportCmd)

	rootCmd.AddCommand(metricsCmd)
	metricsCmd.AddCommand(metricsExposeCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}