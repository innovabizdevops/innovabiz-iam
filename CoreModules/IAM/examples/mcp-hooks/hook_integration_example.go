// Package main fornece um exemplo prático de integração entre os hooks MCP-IAM 
// e o adaptador de observabilidade da plataforma INNOVABIZ.
//
// Este exemplo demonstra a instrumentação completa de hooks em múltiplos mercados e tenants,
// com suporte a normas internacionais e requisitos específicos de compliance.
//
// Conformidades: ISO/IEC 27001, COBIT 2019, TOGAF 10.0, DMBOK 2.0, BNA, GDPR, LGPD
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/innovabiz/iam/constants"
	"github.com/innovabiz/iam/observability/adapter"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// Hook representa um hook MCP-IAM genérico
type Hook struct {
	name       string
	hookType   string
	obs        *adapter.HookObservability
	marketCtxs map[string]adapter.MarketContext
}

func main() {
	// Configurar caminho para logs de compliance
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao obter diretório de configuração: %v\n", err)
		os.Exit(1)
	}
	logsPath := filepath.Join(userConfigDir, "innovabiz", "compliance")

	// Configurar adaptador de observabilidade
	fmt.Println("🚀 Inicializando adaptador de observabilidade para hooks MCP-IAM...")
	config := adapter.Config{
		Environment:           "development",
		ServiceName:           "mcp-iam-hooks-example",
		OTLPEndpoint:          "localhost:4317", // Alterar conforme seu ambiente
		MetricsPort:           9090,             // Porta para métricas Prometheus
		ComplianceLogsPath:    logsPath,
		EnableComplianceAudit: true,
		StructuredLogging:     true,
		LogLevel:              "debug",
		TraceSampleRate:       1.0, // Capturar todos os traces em desenvolvimento
	}

	// Criar adaptador de observabilidade
	obs, err := adapter.NewHookObservability(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Erro ao inicializar adaptador: %v\n", err)
		os.Exit(1)
	}
	defer obs.Close()

	// Criar hook de exemplo
	hook := createExampleHook(obs)

	// Registrar metadados de compliance para diferentes mercados
	fmt.Println("📝 Registrando metadados de compliance por mercado...")
	registerComplianceMetadata(obs)

	// Simular operações do hook em diferentes mercados
	fmt.Println("🔄 Simulando operações do hook em múltiplos mercados...")
	simulateHookOperations(hook)

	// Aguardar sinais para encerramento gracioso
	fmt.Println("✅ Hook em execução. Métricas disponíveis em http://localhost:9090/metrics")
	fmt.Println("📊 Pressione Ctrl+C para encerrar")
	
	// Configurar captura de sinais para encerramento gracioso
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	fmt.Println("\n👋 Encerrando exemplo de hooks MCP-IAM...")
}

// createExampleHook cria um hook MCP-IAM de exemplo com suporte a múltiplos mercados
func createExampleHook(obs *adapter.HookObservability) *Hook {
	hook := &Hook{
		name:       "privilege-elevation-hook",
		hookType:   constants.HookTypePrivilegeElevation,
		obs:        obs,
		marketCtxs: make(map[string]adapter.MarketContext),
	}

	// Criar contextos para diferentes mercados
	markets := []string{
		constants.MarketAngola,
		constants.MarketBrazil,
		constants.MarketEU,
		constants.MarketChina,
		constants.MarketUSA,
		constants.MarketGlobal,
	}

	tenantTypes := []string{
		constants.TenantFinancial,
		constants.TenantRetail,
		constants.TenantHealthcare,
		constants.TenantGovernment,
	}

	// Registrar contextos de mercado/tenant
	for _, market := range markets {
		for _, tenantType := range tenantTypes {
			ctxKey := fmt.Sprintf("%s-%s", market, tenantType)
			marketCtx := adapter.NewMarketContext(market, tenantType, hook.hookType).
				WithMetadata("integration", "example").
				WithMetadata("version", "1.0.0")
			
			hook.marketCtxs[ctxKey] = marketCtx
		}
	}

	fmt.Printf("✓ Hook '%s' criado com suporte a %d contextos de mercado/tenant\n", 
		hook.name, len(hook.marketCtxs))

	return hook
}

// registerComplianceMetadata registra metadados de compliance específicos por mercado
func registerComplianceMetadata(obs *adapter.HookObservability) {
	// Angola - Banco Nacional de Angola
	obs.RegisterComplianceMetadata(
		constants.MarketAngola,
		"BNA",
		true,  // Requer aprovação dual
		constants.MFALevelHigh,
		7,     // 7 anos de retenção
	)

	// Brasil - LGPD e BACEN
	obs.RegisterComplianceMetadata(
		constants.MarketBrazil,
		"LGPD",
		true,  // Requer aprovação dual
		constants.MFALevelHigh,
		5,     // 5 anos de retenção
	)
	obs.RegisterComplianceMetadata(
		constants.MarketBrazil,
		"BACEN",
		true,  // Requer aprovação dual
		constants.MFALevelHigh,
		10,    // 10 anos de retenção
	)

	// União Europeia - GDPR
	obs.RegisterComplianceMetadata(
		constants.MarketEU,
		"GDPR",
		true,  // Requer aprovação dual
		constants.MFALevelHigh,
		7,     // 7 anos de retenção
	)

	// China - Cybersecurity Law
	obs.RegisterComplianceMetadata(
		constants.MarketChina,
		"CSL",
		true,  // Requer aprovação dual
		constants.MFALevelHigh,
		5,     // 5 anos de retenção
	)

	// Estados Unidos - SOX
	obs.RegisterComplianceMetadata(
		constants.MarketUSA,
		"SOX",
		true,  // Requer aprovação dual
		constants.MFALevelMedium,
		7,     // 7 anos de retenção
	)

	// Configuração global padrão
	obs.RegisterComplianceMetadata(
		constants.MarketGlobal,
		"ISO27001",
		false, // Não requer aprovação dual por padrão
		constants.MFALevelMedium,
		3,     // 3 anos de retenção
	)
}

// simulateHookOperations simula operações em diferentes mercados e tenants
func simulateHookOperations(hook *Hook) {
	// Contexto base
	ctx := context.Background()

	// Identificadores de usuário para teste
	users := map[string]string{
		constants.MarketAngola:  "user-123-angola",
		constants.MarketBrazil:  "user-456-brazil",
		constants.MarketEU:      "user-789-eu",
		constants.MarketChina:   "user-101-china",
		constants.MarketUSA:     "user-202-usa",
		constants.MarketGlobal:  "user-303-global",
	}

	// Simular elevações de privilégio em diferentes mercados
	for market, userId := range users {
		// Para demonstração, usar apenas tenant financeiro
		ctxKey := fmt.Sprintf("%s-%s", market, constants.TenantFinancial)
		marketCtx, exists := hook.marketCtxs[ctxKey]
		
		if !exists {
			fmt.Printf("⚠️ Contexto não encontrado para %s\n", ctxKey)
			continue
		}

		fmt.Printf("🔍 Simulando operações para mercado: %s, usuário: %s\n", market, userId)

		// 1. Validar MFA
		mfaLevel := getMFALevelForMarket(market)
		err := hook.obs.ObserveValidateMFA(
			ctx,
			marketCtx,
			userId,
			mfaLevel,
			func(ctx context.Context) error {
				// Simular operação de validação MFA
				time.Sleep(50 * time.Millisecond)
				
				// Simulamos sucesso para todos os mercados
				return nil
			},
		)
		if err != nil {
			fmt.Printf("❌ Erro na validação MFA para %s: %v\n", market, err)
		}

		// 2. Validar escopo
		scope := getScopeForMarket(market)
		err = hook.obs.ObserveValidateScope(
			ctx,
			marketCtx,
			userId,
			scope,
			func(ctx context.Context) error {
				// Simular operação de validação de escopo
				time.Sleep(30 * time.Millisecond)
				
				// Simulamos sucesso para todos os mercados
				return nil
			},
		)
		if err != nil {
			fmt.Printf("❌ Erro na validação de escopo para %s: %v\n", market, err)
		}

		// 3. Registrar evento de auditoria para elevação de privilégio
		hook.obs.TraceAuditEvent(
			ctx,
			marketCtx,
			userId,
			"privilege_elevation",
			fmt.Sprintf("Elevação de privilégio concedida para usuário %s no mercado %s", userId, market),
		)

		// 4. Registrar evento de segurança para verificação de conformidade
		hook.obs.TraceSecurity(
			ctx,
			marketCtx,
			userId,
			constants.SeverityInfo,
			fmt.Sprintf("Verificação de conformidade para elevação de privilégio no mercado %s", market),
			"compliance_check",
		)

		// 5. Atualizar métricas de elevação de privilégio ativa
		hook.obs.UpdateActiveElevations(marketCtx, 1)

		// Pause entre operações de mercados diferentes para demonstração
		time.Sleep(200 * time.Millisecond)
	}

	// Atualizar métricas de cobertura de testes
	hook.obs.RecordTestCoverage(constants.HookTypePrivilegeElevation, 92.5)
	hook.obs.RecordTestCoverage(constants.HookTypeMFAValidation, 94.3)
	hook.obs.RecordTestCoverage(constants.HookTypeTokenIntrospection, 90.8)

	// Exemplo de trace com span personalizado para processo complexo
	marketCtx := hook.marketCtxs[fmt.Sprintf("%s-%s", constants.MarketBrazil, constants.TenantFinancial)]
	
	// Criar trace com múltiplos spans aninhados
	hook.obs.ObserveHookOperation(
		ctx,
		marketCtx,
		"complex_privilege_operation",
		users[constants.MarketBrazil],
		"Operação complexa de elevação de privilégio com aprovação multi-nível",
		[]attribute.KeyValue{
			attribute.String("operation_id", "op-12345"),
			attribute.String("workflow", "multi-approval"),
			attribute.String("compliance_level", "high"),
		},
		func(ctx context.Context) error {
			// Executar sub-operações no mesmo trace
			for i := 1; i <= 3; i++ {
				subOpName := fmt.Sprintf("approval_step_%d", i)
				
				// Criar sub-span para cada etapa
				err := hook.obs.ObserveHookOperation(
					ctx,
					marketCtx,
					subOpName,
					users[constants.MarketBrazil],
					fmt.Sprintf("Etapa de aprovação %d/3", i),
					[]attribute.KeyValue{
						attribute.Int("step", i),
						attribute.Int("total_steps", 3),
					},
					func(ctx context.Context) error {
						// Simular processamento desta etapa
						time.Sleep(100 * time.Millisecond)
						return nil
					},
				)
				
				if err != nil {
					return fmt.Errorf("erro na etapa %d: %w", i, err)
				}
			}
			
			return nil
		},
	)
}

// getMFALevelForMarket retorna o nível MFA apropriado para cada mercado
func getMFALevelForMarket(market string) string {
	switch market {
	case constants.MarketAngola, constants.MarketBrazil, constants.MarketEU, constants.MarketChina:
		return constants.MFALevelHigh
	case constants.MarketUSA:
		return constants.MFALevelMedium
	default:
		return constants.MFALevelMedium
	}
}

// getScopeForMarket retorna o escopo apropriado para cada mercado
func getScopeForMarket(market string) string {
	switch market {
	case constants.MarketAngola:
		return "admin:angola:finance"
	case constants.MarketBrazil:
		return "admin:brazil:finance"
	case constants.MarketEU:
		return "admin:eu:finance"
	case constants.MarketChina:
		return "admin:china:finance"
	case constants.MarketUSA:
		return "admin:usa:finance"
	default:
		return "admin:global:finance"
	}
}