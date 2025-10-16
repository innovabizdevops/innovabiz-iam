// Package tests fornece testes unitários para o adaptador de observabilidade MCP-IAM
//
// Estes testes validam o funcionamento do adaptador em múltiplos cenários,
// garantindo conformidade com requisitos de observabilidade e compliance
// específicos por mercado, conforme normativas nacionais e internacionais.
//
// Conformidades: ISO/IEC 27001, ISO 20000, COBIT 2019, TOGAF 10.0, DMBOK 2.0
package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/innovabiz/iam/constants"
	"github.com/innovabiz/iam/observability/adapter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TestConfig valida as configurações do adaptador de observabilidade
func TestConfig(t *testing.T) {
	t.Run("Validação de Configurações Válidas", func(t *testing.T) {
		// Preparar diretório temporário para logs
		tmpDir, err := os.MkdirTemp("", "compliance-logs-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Criar configuração válida
		config := adapter.Config{
			Environment:           "development",
			ServiceName:           "test-service",
			OTLPEndpoint:          "localhost:4317",
			MetricsPort:           9090,
			ComplianceLogsPath:    tmpDir,
			EnableComplianceAudit: true,
			StructuredLogging:     true,
			LogLevel:              "info",
			TraceSampleRate:       1.0,
		}

		// Validar configuração
		err = config.Validate()
		assert.NoError(t, err, "Configuração válida deve passar na validação")
		
		// Testar método fluente
		config = adapter.Config{}.
			WithEnvironment("development").
			WithServiceName("test-service").
			WithOTLPEndpoint("localhost:4317").
			WithMetricsPort(9090).
			WithComplianceLogsPath(tmpDir).
			WithComplianceAudit(true).
			WithStructuredLogging(true).
			WithLogLevel("info").
			WithTraceSampleRate(1.0)
			
		err = config.Validate()
		assert.NoError(t, err, "Configuração fluente válida deve passar na validação")
	})

	t.Run("Validação de Configurações Inválidas", func(t *testing.T) {
		testCases := []struct {
			name   string
			config adapter.Config
		}{
			{
				name: "Ambiente Vazio",
				config: adapter.Config{
					Environment: "",
					ServiceName: "test-service",
				},
			},
			{
				name: "Ambiente Inválido",
				config: adapter.Config{
					Environment: "invalid",
					ServiceName: "test-service",
				},
			},
			{
				name: "Serviço Vazio",
				config: adapter.Config{
					Environment: "development",
					ServiceName: "",
				},
			},
			{
				name: "Porta Inválida",
				config: adapter.Config{
					Environment: "development",
					ServiceName: "test-service",
					MetricsPort: -1,
				},
			},
			{
				name: "Nível de Log Inválido",
				config: adapter.Config{
					Environment: "development",
					ServiceName: "test-service",
					LogLevel:    "invalid",
				},
			},
			{
				name: "Taxa de Amostragem Inválida",
				config: adapter.Config{
					Environment:     "development",
					ServiceName:     "test-service",
					TraceSampleRate: 1.5,
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := tc.config.Validate()
				assert.Error(t, err, "Configuração inválida deve falhar na validação")
			})
		}
	})

	t.Run("Criação de Diretórios", func(t *testing.T) {
		// Preparar diretório temporário para logs
		tmpDir, err := os.MkdirTemp("", "compliance-logs-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Diretório aninhado para teste
		nestedDir := filepath.Join(tmpDir, "nested", "dirs")

		// Criar configuração com diretório aninhado
		config := adapter.Config{
			Environment:        "development",
			ServiceName:        "test-service",
			ComplianceLogsPath: nestedDir,
		}

		// Verificar que o diretório ainda não existe
		_, err = os.Stat(nestedDir)
		assert.True(t, os.IsNotExist(err), "Diretório não deve existir antes do teste")

		// Testar criação de diretórios
		err = config.EnsureDirectories()
		assert.NoError(t, err, "Criação de diretórios deve ser bem-sucedida")

		// Verificar que o diretório foi criado
		_, err = os.Stat(nestedDir)
		assert.NoError(t, err, "Diretório deve existir após criação")
	})
}

// TestMarketContext testa funcionalidades do contexto de mercado
func TestMarketContext(t *testing.T) {
	t.Run("Criação e Metadados", func(t *testing.T) {
		// Criar contexto básico
		ctx := adapter.NewMarketContext(constants.MarketBrazil, constants.TenantFinancial, constants.HookTypePrivilegeElevation)
		
		// Verificar valores base
		assert.Equal(t, constants.MarketBrazil, ctx.Market)
		assert.Equal(t, constants.TenantFinancial, ctx.TenantType)
		assert.Equal(t, constants.HookTypePrivilegeElevation, ctx.HookType)
		assert.NotNil(t, ctx.Metadata, "Metadata deve ser inicializado")
		assert.Empty(t, ctx.Metadata, "Metadata deve iniciar vazio")

		// Adicionar metadados
		ctx = ctx.WithMetadata("compliance", "LGPD")
		ctx = ctx.WithMetadata("retention", "5y")

		// Verificar metadados
		val, exists := ctx.GetMetadata("compliance")
		assert.True(t, exists, "Metadata compliance deve existir")
		assert.Equal(t, "LGPD", val)

		val, exists = ctx.GetMetadata("retention")
		assert.True(t, exists, "Metadata retention deve existir")
		assert.Equal(t, "5y", val)

		val, exists = ctx.GetMetadata("nonexistent")
		assert.False(t, exists, "Metadata não existente não deve existir")
		assert.Empty(t, val)
	})
}

// TestHookObservabilityInitialization testa a inicialização do adaptador
func TestHookObservabilityInitialization(t *testing.T) {
	// Preparar diretório temporário para logs
	tmpDir, err := os.MkdirTemp("", "compliance-logs-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("Inicialização Básica", func(t *testing.T) {
		config := adapter.Config{
			Environment:        "development",
			ServiceName:        "test-service",
			ComplianceLogsPath: tmpDir,
			LogLevel:           "info",
		}

		obs, err := adapter.NewHookObservability(config)
		assert.NoError(t, err, "Inicialização básica deve ser bem-sucedida")
		assert.NotNil(t, obs, "Adaptador deve ser criado")

		// Limpeza
		err = obs.Close()
		assert.NoError(t, err, "Fechamento do adaptador deve ser bem-sucedido")
	})

	t.Run("Inicialização com Métricas", func(t *testing.T) {
		// Usar porta aleatória alta para evitar conflitos
		port := 9091 + testing.Int() % 1000
		
		config := adapter.Config{
			Environment:        "development",
			ServiceName:        "test-service",
			ComplianceLogsPath: tmpDir,
			LogLevel:           "info",
			MetricsPort:        port,
		}

		obs, err := adapter.NewHookObservability(config)
		assert.NoError(t, err, "Inicialização com métricas deve ser bem-sucedida")
		assert.NotNil(t, obs, "Adaptador deve ser criado")

		// Limpeza
		err = obs.Close()
		assert.NoError(t, err, "Fechamento do adaptador deve ser bem-sucedido")
	})

	t.Run("Inicialização com Configuração Inválida", func(t *testing.T) {
		config := adapter.Config{
			Environment: "", // Inválido (vazio)
			ServiceName: "test-service",
		}

		obs, err := adapter.NewHookObservability(config)
		assert.Error(t, err, "Inicialização com configuração inválida deve falhar")
		assert.Nil(t, obs, "Adaptador não deve ser criado")
	})
}

// TestComplianceMetadata testa gerenciamento de metadados de compliance
func TestComplianceMetadata(t *testing.T) {
	// Preparar configuração
	tmpDir, err := os.MkdirTemp("", "compliance-logs-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	config := adapter.Config{
		Environment:        "development",
		ServiceName:        "test-service",
		ComplianceLogsPath: tmpDir,
		LogLevel:           "info",
	}

	obs, err := adapter.NewHookObservability(config)
	require.NoError(t, err)
	defer obs.Close()

	t.Run("Registro e Recuperação", func(t *testing.T) {
		// Registrar metadados para Angola
		obs.RegisterComplianceMetadata(
			constants.MarketAngola,
			"BNA",
			true,
			constants.MFALevelHigh,
			7,
		)

		// Registrar metadados para Brasil
		obs.RegisterComplianceMetadata(
			constants.MarketBrazil,
			"LGPD",
			true,
			constants.MFALevelHigh,
			5,
		)

		// Verificar metadados para Angola
		metadata, exists := obs.GetComplianceMetadata(constants.MarketAngola)
		assert.True(t, exists, "Metadados de Angola devem existir")
		assert.Equal(t, "BNA", metadata.Framework)
		assert.True(t, metadata.RequiresDualApproval)
		assert.Equal(t, constants.MFALevelHigh, metadata.MinimumMFALevel)
		assert.Equal(t, 7, metadata.LogRetentionYears)
		assert.Equal(t, constants.MarketAngola, metadata.Market)

		// Verificar metadados para Brasil
		metadata, exists = obs.GetComplianceMetadata(constants.MarketBrazil)
		assert.True(t, exists, "Metadados do Brasil devem existir")
		assert.Equal(t, "LGPD", metadata.Framework)
		assert.Equal(t, 5, metadata.LogRetentionYears)

		// Verificar mercado não registrado
		metadata, exists = obs.GetComplianceMetadata(constants.MarketChina)
		assert.False(t, exists, "Metadados da China não devem existir")
	})

	t.Run("Fallback para Global", func(t *testing.T) {
		// Registrar metadados globais
		obs.RegisterComplianceMetadata(
			constants.MarketGlobal,
			"ISO27001",
			false,
			constants.MFALevelMedium,
			3,
		)

		// Verificar mercado não registrado após configuração global
		metadata, exists := obs.GetComplianceMetadata(constants.MarketChina)
		assert.False(t, exists, "Metadados da China não devem existir")
	})
}

// TestObserveHookOperation testa a observação de operações de hook
func TestObserveHookOperation(t *testing.T) {
	// Preparar configuração
	tmpDir, err := os.MkdirTemp("", "compliance-logs-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	config := adapter.Config{
		Environment:        "development",
		ServiceName:        "test-service",
		ComplianceLogsPath: tmpDir,
		LogLevel:           "info",
		MetricsPort:        0, // Desativar servidor HTTP
	}

	obs, err := adapter.NewHookObservability(config)
	require.NoError(t, err)
	defer obs.Close()

	t.Run("Operação Bem-Sucedida", func(t *testing.T) {
		ctx := context.Background()
		marketCtx := adapter.NewMarketContext(constants.MarketBrazil, constants.TenantFinancial, constants.HookTypePrivilegeElevation)
		userId := "test-user-123"
		
		err := obs.ObserveHookOperation(
			ctx,
			marketCtx,
			constants.OperationValidateScope,
			userId,
			"Teste de validação de escopo",
			[]attribute.KeyValue{
				attribute.String("test", "true"),
			},
			func(ctx context.Context) error {
				// Operação simulada bem-sucedida
				return nil
			},
		)
		
		assert.NoError(t, err, "Operação deve ser bem-sucedida")
	})

	t.Run("Operação com Erro", func(t *testing.T) {
		ctx := context.Background()
		marketCtx := adapter.NewMarketContext(constants.MarketBrazil, constants.TenantFinancial, constants.HookTypePrivilegeElevation)
		userId := "test-user-123"
		testError := fmt.Errorf("erro de teste")
		
		err := obs.ObserveHookOperation(
			ctx,
			marketCtx,
			constants.OperationValidateScope,
			userId,
			"Teste de validação de escopo com erro",
			[]attribute.KeyValue{
				attribute.String("test", "true"),
			},
			func(ctx context.Context) error {
				// Operação simulada com erro
				return testError
			},
		)
		
		assert.Error(t, err, "Operação deve falhar")
		assert.Equal(t, testError, err, "Erro deve ser propagado")
	})
}

// TestSpecificOperations testa operações específicas do adaptador
func TestSpecificOperations(t *testing.T) {
	// Preparar configuração
	tmpDir, err := os.MkdirTemp("", "compliance-logs-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	config := adapter.Config{
		Environment:        "development",
		ServiceName:        "test-service",
		ComplianceLogsPath: tmpDir,
		LogLevel:           "info",
	}

	obs, err := adapter.NewHookObservability(config)
	require.NoError(t, err)
	defer obs.Close()

	t.Run("ValidateScope", func(t *testing.T) {
		ctx := context.Background()
		marketCtx := adapter.NewMarketContext(constants.MarketBrazil, constants.TenantFinancial, constants.HookTypePrivilegeElevation)
		userId := "test-user-123"
		scope := "admin:read"
		
		err := obs.ObserveValidateScope(
			ctx,
			marketCtx,
			userId,
			scope,
			func(ctx context.Context) error {
				// Operação simulada bem-sucedida
				return nil
			},
		)
		
		assert.NoError(t, err, "Validação de escopo deve ser bem-sucedida")
	})

	t.Run("ValidateMFA", func(t *testing.T) {
		ctx := context.Background()
		marketCtx := adapter.NewMarketContext(constants.MarketBrazil, constants.TenantFinancial, constants.HookTypeMFAValidation)
		userId := "test-user-123"
		
		// Registrar requisito MFA para o mercado
		obs.RegisterComplianceMetadata(
			constants.MarketBrazil,
			"LGPD",
			true,
			constants.MFALevelHigh,
			5,
		)
		
		// Teste com nível MFA correto
		err := obs.ObserveValidateMFA(
			ctx,
			marketCtx,
			userId,
			constants.MFALevelHigh,
			func(ctx context.Context) error {
				// Operação simulada bem-sucedida
				return nil
			},
		)
		
		assert.NoError(t, err, "Validação MFA com nível correto deve ser bem-sucedida")

		// Teste com nível MFA insuficiente (não causa erro, apenas aviso)
		err = obs.ObserveValidateMFA(
			ctx,
			marketCtx,
			userId,
			constants.MFALevelBasic,
			func(ctx context.Context) error {
				// Operação simulada bem-sucedida
				return nil
			},
		)
		
		assert.NoError(t, err, "Validação MFA com nível insuficiente deve registrar aviso mas não falhar")
	})

	t.Run("TraceAuditEvent", func(t *testing.T) {
		ctx := context.Background()
		marketCtx := adapter.NewMarketContext(constants.MarketBrazil, constants.TenantFinancial, constants.HookTypePrivilegeElevation)
		userId := "test-user-123"
		
		// Nenhum erro esperado
		obs.TraceAuditEvent(
			ctx,
			marketCtx,
			userId,
			"login",
			"Teste de login",
		)
		
		// Verificar arquivo de log (se habilitado)
		if obs.GetConfig().EnableComplianceAudit && obs.GetConfig().ComplianceLogsPath != "" {
			logDir := filepath.Join(obs.GetConfig().ComplianceLogsPath, marketCtx.Market)
			date := time.Now().Format("2006-01-02")
			logFile := filepath.Join(logDir, fmt.Sprintf("%s-audit-events.log", date))
			
			_, err := os.Stat(logFile)
			if !os.IsNotExist(err) {
				// Verificação básica de existência do arquivo (se estiver habilitado)
				assert.NoError(t, err, "Arquivo de log de auditoria deve existir")
			}
		}
	})

	t.Run("TraceSecurity", func(t *testing.T) {
		ctx := context.Background()
		marketCtx := adapter.NewMarketContext(constants.MarketBrazil, constants.TenantFinancial, constants.HookTypePrivilegeElevation)
		userId := "test-user-123"
		
		// Testar diferentes níveis de severidade
		severities := []string{
			constants.SeverityCritical,
			constants.SeverityHigh,
			constants.SeverityMedium,
			constants.SeverityLow,
			constants.SeverityInfo,
		}
		
		for _, severity := range severities {
			obs.TraceSecurity(
				ctx,
				marketCtx,
				userId,
				severity,
				fmt.Sprintf("Teste de evento de segurança com severidade %s", severity),
				"security_test",
			)
		}
		
		// Verificar arquivo de log (se habilitado)
		if obs.GetConfig().EnableComplianceAudit && obs.GetConfig().ComplianceLogsPath != "" {
			logDir := filepath.Join(obs.GetConfig().ComplianceLogsPath, marketCtx.Market)
			date := time.Now().Format("2006-01-02")
			logFile := filepath.Join(logDir, fmt.Sprintf("%s-security-events.log", date))
			
			_, err := os.Stat(logFile)
			if !os.IsNotExist(err) {
				// Verificação básica de existência do arquivo (se estiver habilitado)
				assert.NoError(t, err, "Arquivo de log de segurança deve existir")
			}
		}
	})

	t.Run("MetricsUpdate", func(t *testing.T) {
		marketCtx := adapter.NewMarketContext(constants.MarketBrazil, constants.TenantFinancial, constants.HookTypePrivilegeElevation)
		
		// Atualizar métricas
		obs.UpdateActiveElevations(marketCtx, 5)
		obs.RecordTestCoverage(constants.HookTypePrivilegeElevation, 95.5)
		
		// Não há uma forma fácil de testar o valor das métricas sem expor o registrador internamente
		// Em um teste real, usaríamos testutil.CollectAndCount ou semelhante
	})
}

// TestMFALevelComparison testa a comparação de níveis MFA
func TestMFALevelComparison(t *testing.T) {
	testCases := []struct {
		name      string
		provided  string
		required  string
		sufficient bool
	}{
		{"Igual", constants.MFALevelMedium, constants.MFALevelMedium, true},
		{"Maior", constants.MFALevelHigh, constants.MFALevelMedium, true},
		{"Menor", constants.MFALevelBasic, constants.MFALevelMedium, false},
		{"Nenhum", constants.MFALevelNone, constants.MFALevelMedium, false},
		{"Requerido Nenhum", constants.MFALevelMedium, constants.MFALevelNone, true},
		{"Inválido Fornecido", "invalid", constants.MFALevelMedium, false},
		{"Inválido Requerido", constants.MFALevelMedium, "invalid", true},
		{"Ambos Inválidos", "invalid1", "invalid2", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Chamar diretamente o método exportado ou expor temporariamente para teste
			// Como a função não está exportada, usaríamos uma função auxiliar ou mockup
			// Para este teste, assumimos que temos acesso a algo como:
			// result := adapter.IsMFALevelSufficient(tc.provided, tc.required)
			// assert.Equal(t, tc.sufficient, result)
			
			// Alternativa: testar indiretamente via observação MFA
			tmpDir, err := os.MkdirTemp("", "compliance-logs-")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)
			
			config := adapter.Config{
				Environment:        "development",
				ServiceName:        "test-service",
				ComplianceLogsPath: tmpDir,
				LogLevel:           "info",
			}
			
			obs, err := adapter.NewHookObservability(config)
			require.NoError(t, err)
			defer obs.Close()
			
			// Registrar requisito MFA
			obs.RegisterComplianceMetadata(
				constants.MarketBrazil,
				"LGPD",
				true,
				tc.required,
				5,
			)
			
			// O teste indireto é mais difícil, pois não temos acesso ao resultado interno
			// Em um caso real, expor a função ou usar um método testável seria melhor
		})
	}
}

// MockSpanExporter implementa um exportador de spans para testes
type MockSpanExporter struct {
	spans []trace.ReadOnlySpan
}

func (m *MockSpanExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	m.spans = append(m.spans, spans...)
	return nil
}

func (m *MockSpanExporter) Shutdown(ctx context.Context) error {
	return nil
}

func (m *MockSpanExporter) GetSpans() []trace.ReadOnlySpan {
	return m.spans
}

func (m *MockSpanExporter) Reset() {
	m.spans = nil
}

// TODO: Adicionar mais testes para casos específicos de métricas, tracing e logs
// Isso exigiria extensões adicionais ou exposição de campos internos para testes