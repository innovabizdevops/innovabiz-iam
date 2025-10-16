// Package integration_tests implementa testes de integração para a camada 
// de observabilidade dos hooks MCP-IAM da plataforma INNOVABIZ.
//
// Estes testes validam a integração entre métricas, tracing e logging
// em diversos cenários multi-mercado, multi-tenant e multi-contexto,
// verificando conformidade com normas e frameworks internacionais.
//
// Conformidades: ISO/IEC 29119, ISO 9001, ISO 27001, COBIT 2019, TOGAF 10.0
// Frameworks de Teste: TDD, BDD, AAA (Arrange-Act-Assert), FIRST
package integration_tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/innovabiz/iam/constants"
	"github.com/innovabiz/iam/observability/adapter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

// marketConfig contém configurações de teste para um mercado específico
type marketConfig struct {
	market            string
	tenantType        string
	hookType          string
	complianceLevel   string
	regulations       []string
	requiresMFA       bool
	mfaLevel          string
	requiresApproval  bool
	retentionYears    int
}

// TestConfigurations mapeia configurações de teste para diferentes mercados
var TestConfigurations = map[string]marketConfig{
	constants.MarketAngola: {
		market:            constants.MarketAngola,
		tenantType:        constants.TenantFinancial,
		hookType:          constants.HookTypePrivilegeElevation,
		complianceLevel:   constants.ComplianceStrict,
		regulations:       []string{"BNA", "ISO27001"},
		requiresMFA:       true,
		mfaLevel:          constants.MFALevelHigh,
		requiresApproval:  true,
		retentionYears:    7,
	},
	constants.MarketBrazil: {
		market:            constants.MarketBrazil,
		tenantType:        constants.TenantHealthcare,
		hookType:          constants.HookTypeTokenValidation,
		complianceLevel:   constants.ComplianceStrict,
		regulations:       []string{"LGPD", "BACEN"},
		requiresMFA:       true,
		mfaLevel:          constants.MFALevelHigh,
		requiresApproval:  true,
		retentionYears:    5,
	},
	constants.MarketEU: {
		market:            constants.MarketEU,
		tenantType:        constants.TenantTelecom,
		hookType:          constants.HookTypeAudit,
		complianceLevel:   constants.ComplianceEnhanced,
		regulations:       []string{"GDPR", "PSD2"},
		requiresMFA:       true,
		mfaLevel:          constants.MFALevelMedium,
		requiresApproval:  true,
		retentionYears:    7,
	},
	constants.MarketUSA: {
		market:            constants.MarketUSA,
		tenantType:        constants.TenantRetail,
		hookType:          constants.HookTypeElevationApproval,
		complianceLevel:   constants.ComplianceStandard,
		regulations:       []string{"SOX", "PCI-DSS"},
		requiresMFA:       true,
		mfaLevel:          constants.MFALevelMedium,
		requiresApproval:  false,
		retentionYears:    5,
	},
}

// TestIntegrationObservabilitySetup valida a configuração do adaptador para múltiplos mercados
func TestIntegrationObservabilitySetup(t *testing.T) {
	for market, config := range TestConfigurations {
		t.Run(fmt.Sprintf("Setup_%s", market), func(t *testing.T) {
			// Arrange
			tempDir := t.TempDir()
			logsDir := filepath.Join(tempDir, "logs", market)
			
			// Inicializar configuração de adaptador
			obsConfig := adapter.Config{
				Environment:           constants.EnvTest,
				ServiceName:           fmt.Sprintf("mcp-iam-hook-test-%s", market),
				OTLPEndpoint:          "",
				MetricsPort:           0,
				ComplianceLogsPath:    logsDir,
				EnableComplianceAudit: true,
				StructuredLogging:     true,
				LogLevel:              "debug",
			}
			
			// Act
			obs, err := adapter.NewHookObservability(obsConfig)
			
			// Assert
			require.NoError(t, err)
			require.NotNil(t, obs)
			
			// Registrar metadados de compliance
			obs.RegisterComplianceMetadata(
				config.market,
				config.regulations[0],
				config.requiresApproval,
				config.mfaLevel,
				config.retentionYears,
			)
			
			// Criar contexto de mercado
			marketCtx := adapter.NewMarketContext(config.market, config.tenantType, config.hookType)
			
			// Verificar se o contexto de mercado foi criado corretamente
			assert.Equal(t, config.market, marketCtx.Market)
			assert.Equal(t, config.tenantType, marketCtx.TenantType)
			assert.Equal(t, config.hookType, marketCtx.HookType)
			assert.NotEmpty(t, marketCtx.ApplicableRegulations)
		})
	}
}

// TestIntegrationObserveHookOperation valida a observabilidade de operações de hook em diferentes mercados
func TestIntegrationObserveHookOperation(t *testing.T) {
	// Inicializar tracer para teste
	exporter, _ := stdouttrace.New(stdouttrace.WithPrettyPrint())
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithBatcher(exporter),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("mcp-iam-hook-test"),
			semconv.ServiceVersionKey.String("v1.0.0"),
			attribute.String("environment", "test"),
		)),
	)
	defer tp.Shutdown(context.Background())
	otel.SetTracerProvider(tp)
	
	for market, config := range TestConfigurations {
		t.Run(fmt.Sprintf("HookOperation_%s", market), func(t *testing.T) {
			// Arrange
			tempDir := t.TempDir()
			logsDir := filepath.Join(tempDir, "logs", market)
			
			// Criar diretório para logs
			err := os.MkdirAll(logsDir, 0755)
			require.NoError(t, err)
			
			// Inicializar configuração de adaptador
			obsConfig := adapter.Config{
				Environment:           constants.EnvTest,
				ServiceName:           fmt.Sprintf("mcp-iam-hook-test-%s", market),
				OTLPEndpoint:          "",
				MetricsPort:           0,
				ComplianceLogsPath:    logsDir,
				EnableComplianceAudit: true,
				StructuredLogging:     true,
				LogLevel:              "debug",
			}
			
			// Criar adaptador de observabilidade
			obs, err := adapter.NewHookObservability(obsConfig)
			require.NoError(t, err)
			
			// Criar contexto de mercado
			marketCtx := adapter.NewMarketContext(config.market, config.tenantType, config.hookType)
			
			// Contexto com cancelamento para testes
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			
			// Contador para verificar se a função foi chamada
			operationCalled := false
			userId := "test-user-123"
			operation := constants.OperationValidateScope
			description := "Validação de escopo para teste"
			
			// Act - Executar operação de hook
			err = obs.ObserveHookOperation(
				ctx,
				marketCtx,
				operation,
				userId,
				description,
				[]attribute.KeyValue{
					attribute.String("scope", "admin:read"),
					attribute.String("test_market", marketCtx.Market),
				},
				func(ctx context.Context) error {
					operationCalled = true
					// Simular trabalho
					time.Sleep(10 * time.Millisecond)
					return nil
				},
			)
			
			// Assert
			assert.NoError(t, err)
			assert.True(t, operationCalled, "Operação de hook deveria ter sido executada")
			
			// Verificar se logs foram criados (abordagem simplificada para teste)
			files, err := os.ReadDir(logsDir)
			require.NoError(t, err)
			assert.True(t, len(files) > 0, "Arquivos de log deveriam ter sido criados")
		})
	}
}

// TestIntegrationValidateScope valida a observabilidade de validação de escopo em diferentes mercados
func TestIntegrationValidateScope(t *testing.T) {
	for market, config := range TestConfigurations {
		t.Run(fmt.Sprintf("ValidateScope_%s", market), func(t *testing.T) {
			// Arrange
			tempDir := t.TempDir()
			logsDir := filepath.Join(tempDir, "logs", market)
			
			// Criar diretório para logs
			err := os.MkdirAll(logsDir, 0755)
			require.NoError(t, err)
			
			// Inicializar configuração de adaptador
			obsConfig := adapter.Config{
				Environment:           constants.EnvTest,
				ServiceName:           fmt.Sprintf("mcp-iam-hook-test-%s", market),
				OTLPEndpoint:          "",
				MetricsPort:           0,
				ComplianceLogsPath:    logsDir,
				EnableComplianceAudit: true,
				StructuredLogging:     true,
				LogLevel:              "debug",
			}
			
			// Criar adaptador de observabilidade
			obs, err := adapter.NewHookObservability(obsConfig)
			require.NoError(t, err)
			
			// Criar contexto de mercado
			marketCtx := adapter.NewMarketContext(config.market, config.tenantType, config.hookType)
			
			// Contexto com cancelamento para testes
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			
			// Contador para verificar se a função foi chamada
			validateCalled := false
			userId := "test-user-123"
			scope := "admin:read"
			
			// Act - Validar escopo
			err = obs.ObserveValidateScope(
				ctx,
				marketCtx,
				userId,
				scope,
				func(ctx context.Context) error {
					validateCalled = true
					// Simular trabalho
					time.Sleep(10 * time.Millisecond)
					return nil
				},
			)
			
			// Assert
			assert.NoError(t, err)
			assert.True(t, validateCalled, "Validação de escopo deveria ter sido executada")
			
			// Verificar se logs foram criados
			files, err := os.ReadDir(logsDir)
			require.NoError(t, err)
			assert.True(t, len(files) > 0, "Arquivos de log deveriam ter sido criados")
			
			// Verificar conteúdo dos logs (abordagem simplificada)
			logFound := false
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".log") {
					logContent, err := os.ReadFile(filepath.Join(logsDir, file.Name()))
					require.NoError(t, err)
					
					logText := string(logContent)
					if strings.Contains(logText, userId) && strings.Contains(logText, scope) {
						logFound = true
						break
					}
				}
			}
			assert.True(t, logFound, "Log deveria conter informações de usuário e escopo")
		})
	}
}

// TestIntegrationValidateMFA valida a observabilidade de validação MFA em diferentes mercados
func TestIntegrationValidateMFA(t *testing.T) {
	for market, config := range TestConfigurations {
		t.Run(fmt.Sprintf("ValidateMFA_%s", market), func(t *testing.T) {
			// Arrange
			tempDir := t.TempDir()
			logsDir := filepath.Join(tempDir, "logs", market)
			
			// Criar diretório para logs
			err := os.MkdirAll(logsDir, 0755)
			require.NoError(t, err)
			
			// Inicializar configuração de adaptador
			obsConfig := adapter.Config{
				Environment:           constants.EnvTest,
				ServiceName:           fmt.Sprintf("mcp-iam-hook-test-%s", market),
				OTLPEndpoint:          "",
				MetricsPort:           0,
				ComplianceLogsPath:    logsDir,
				EnableComplianceAudit: true,
				StructuredLogging:     true,
				LogLevel:              "debug",
			}
			
			// Criar adaptador de observabilidade com registro em Prometheus
			registry := prometheus.NewRegistry()
			obs, err := adapter.NewHookObservability(obsConfig)
			require.NoError(t, err)
			
			// Criar contexto de mercado
			marketCtx := adapter.NewMarketContext(config.market, config.tenantType, config.hookType)
			
			// Contexto com cancelamento para testes
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			
			// Contador para verificar se a função foi chamada
			validateCalled := false
			userId := "test-user-123"
			
			// Act - Validar MFA
			err = obs.ObserveValidateMFA(
				ctx,
				marketCtx,
				userId,
				config.mfaLevel,
				func(ctx context.Context) error {
					validateCalled = true
					// Simular trabalho
					time.Sleep(10 * time.Millisecond)
					return nil
				},
			)
			
			// Assert
			assert.NoError(t, err)
			assert.True(t, validateCalled, "Validação MFA deveria ter sido executada")
		})
	}
}

// TestIntegrationErrorHandling valida o tratamento de erros na observabilidade
func TestIntegrationErrorHandling(t *testing.T) {
	for market, config := range TestConfigurations {
		t.Run(fmt.Sprintf("ErrorHandling_%s", market), func(t *testing.T) {
			// Arrange
			tempDir := t.TempDir()
			logsDir := filepath.Join(tempDir, "logs", market)
			
			// Criar diretório para logs
			err := os.MkdirAll(logsDir, 0755)
			require.NoError(t, err)
			
			// Inicializar configuração de adaptador
			obsConfig := adapter.Config{
				Environment:           constants.EnvTest,
				ServiceName:           fmt.Sprintf("mcp-iam-hook-test-%s", market),
				OTLPEndpoint:          "",
				MetricsPort:           0,
				ComplianceLogsPath:    logsDir,
				EnableComplianceAudit: true,
				StructuredLogging:     true,
				LogLevel:              "debug",
			}
			
			// Criar adaptador de observabilidade
			obs, err := adapter.NewHookObservability(obsConfig)
			require.NoError(t, err)
			
			// Criar contexto de mercado
			marketCtx := adapter.NewMarketContext(config.market, config.tenantType, config.hookType)
			
			// Contexto com cancelamento para testes
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			
			// Contador para verificar se a função foi chamada
			operationCalled := false
			userId := "test-user-123"
			operation := constants.OperationValidateScope
			description := "Validação de escopo para teste"
			expectedError := fmt.Errorf("erro de teste: acesso negado")
			
			// Act - Executar operação de hook com erro
			err = obs.ObserveHookOperation(
				ctx,
				marketCtx,
				operation,
				userId,
				description,
				[]attribute.KeyValue{
					attribute.String("scope", "admin:write"),
					attribute.String("test_market", marketCtx.Market),
				},
				func(ctx context.Context) error {
					operationCalled = true
					return expectedError
				},
			)
			
			// Assert
			assert.Error(t, err)
			assert.Equal(t, expectedError, err)
			assert.True(t, operationCalled, "Operação de hook deveria ter sido executada mesmo com erro")
			
			// Verificar se logs de erro foram criados
			files, err := os.ReadDir(logsDir)
			require.NoError(t, err)
			assert.True(t, len(files) > 0, "Arquivos de log deveriam ter sido criados")
			
			// Verificar conteúdo dos logs de erro (abordagem simplificada)
			errorLogFound := false
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".log") {
					logContent, err := os.ReadFile(filepath.Join(logsDir, file.Name()))
					require.NoError(t, err)
					
					logText := string(logContent)
					if strings.Contains(logText, "erro de teste") {
						errorLogFound = true
						break
					}
				}
			}
			assert.True(t, errorLogFound, "Log deveria conter informações do erro")
		})
	}
}

// TestIntegrationAuditEvents valida o registro de eventos de auditoria
func TestIntegrationAuditEvents(t *testing.T) {
	for market, config := range TestConfigurations {
		t.Run(fmt.Sprintf("AuditEvents_%s", market), func(t *testing.T) {
			// Arrange
			tempDir := t.TempDir()
			logsDir := filepath.Join(tempDir, "logs", market)
			
			// Criar diretório para logs
			err := os.MkdirAll(logsDir, 0755)
			require.NoError(t, err)
			
			// Inicializar configuração de adaptador
			obsConfig := adapter.Config{
				Environment:           constants.EnvTest,
				ServiceName:           fmt.Sprintf("mcp-iam-hook-test-%s", market),
				OTLPEndpoint:          "",
				MetricsPort:           0,
				ComplianceLogsPath:    logsDir,
				EnableComplianceAudit: true,
				StructuredLogging:     true,
				LogLevel:              "debug",
			}
			
			// Criar adaptador de observabilidade
			obs, err := adapter.NewHookObservability(obsConfig)
			require.NoError(t, err)
			
			// Criar contexto de mercado
			marketCtx := adapter.NewMarketContext(config.market, config.tenantType, config.hookType)
			
			// Contexto com cancelamento para testes
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			
			// Dados para o evento de auditoria
			userId := "test-user-123"
			eventType := "privilege_elevation"
			eventDetails := fmt.Sprintf("Elevação de privilégio para usuário teste em %s", config.market)
			
			// Act - Registrar evento de auditoria
			obs.TraceAuditEvent(
				ctx,
				marketCtx,
				userId,
				eventType,
				eventDetails,
			)
			
			// Registrar evento de segurança
			obs.TraceSecurity(
				ctx,
				marketCtx,
				userId,
				"medium",
				"Evento de segurança para teste",
				"security_test",
			)
			
			// Assert - Verificar se logs foram criados
			files, err := os.ReadDir(logsDir)
			require.NoError(t, err)
			assert.True(t, len(files) > 0, "Arquivos de log deveriam ter sido criados")
			
			// Verificar conteúdo dos logs (abordagem simplificada)
			auditLogFound := false
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".log") {
					logContent, err := os.ReadFile(filepath.Join(logsDir, file.Name()))
					require.NoError(t, err)
					
					logText := string(logContent)
					if strings.Contains(logText, eventType) && strings.Contains(logText, userId) {
						auditLogFound = true
						break
					}
				}
			}
			assert.True(t, auditLogFound, "Log deveria conter informações do evento de auditoria")
		})
	}
}

// TestIntegrationMultiMarketTracking valida o rastreamento simultâneo de operações em múltiplos mercados
func TestIntegrationMultiMarketTracking(t *testing.T) {
	// Arrange
	// Selecionar dois mercados para teste simultâneo
	markets := []string{constants.MarketAngola, constants.MarketBrazil}
	
	tempDir := t.TempDir()
	
	// Inicializar configuração para cada mercado
	obsAdapters := make(map[string]*adapter.HookObservability)
	marketContexts := make(map[string]adapter.MarketContext)
	
	for _, market := range markets {
		config := TestConfigurations[market]
		logsDir := filepath.Join(tempDir, "logs", market)
		
		// Criar diretório para logs
		err := os.MkdirAll(logsDir, 0755)
		require.NoError(t, err)
		
		// Inicializar adaptador
		obsConfig := adapter.Config{
			Environment:           constants.EnvTest,
			ServiceName:           fmt.Sprintf("mcp-iam-hook-test-%s", market),
			OTLPEndpoint:          "",
			MetricsPort:           0,
			ComplianceLogsPath:    logsDir,
			EnableComplianceAudit: true,
			StructuredLogging:     true,
			LogLevel:              "debug",
		}
		
		obs, err := adapter.NewHookObservability(obsConfig)
		require.NoError(t, err)
		obsAdapters[market] = obs
		
		// Criar contexto de mercado
		marketCtx := adapter.NewMarketContext(config.market, config.tenantType, config.hookType)
		marketContexts[market] = marketCtx
	}
	
	// Contexto com cancelamento para testes
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Simultaneamente validar escopo em diferentes mercados
	for _, market := range markets {
		market := market // Capturar variável para goroutine
		t.Run(fmt.Sprintf("MultiMarket_%s", market), func(t *testing.T) {
			obs := obsAdapters[market]
			marketCtx := marketContexts[market]
			userId := fmt.Sprintf("user-%s", market)
			scope := "admin:read"
			
			// Act - Validar escopo
			err := obs.ObserveValidateScope(
				ctx,
				marketCtx,
				userId,
				scope,
				func(ctx context.Context) error {
					// Simular trabalho
					time.Sleep(10 * time.Millisecond)
					return nil
				},
			)
			
			// Assert
			assert.NoError(t, err)
		})
	}
}

// TestIntegrationComplianceAudit valida os logs de auditoria de compliance
func TestIntegrationComplianceAudit(t *testing.T) {
	for market, config := range TestConfigurations {
		t.Run(fmt.Sprintf("ComplianceAudit_%s", market), func(t *testing.T) {
			// Arrange
			tempDir := t.TempDir()
			logsDir := filepath.Join(tempDir, "logs", market)
			
			// Criar diretório para logs
			err := os.MkdirAll(logsDir, 0755)
			require.NoError(t, err)
			
			// Inicializar configuração de adaptador
			obsConfig := adapter.Config{
				Environment:           constants.EnvTest,
				ServiceName:           fmt.Sprintf("mcp-iam-hook-test-%s", market),
				OTLPEndpoint:          "",
				MetricsPort:           0,
				ComplianceLogsPath:    logsDir,
				EnableComplianceAudit: true,
				StructuredLogging:     true,
				LogLevel:              "debug",
			}
			
			// Criar adaptador de observabilidade
			obs, err := adapter.NewHookObservability(obsConfig)
			require.NoError(t, err)
			
			// Registrar metadados de compliance
			for _, regulation := range config.regulations {
				obs.RegisterComplianceMetadata(
					config.market,
					regulation,
					config.requiresApproval,
					config.mfaLevel,
					config.retentionYears,
				)
			}
			
			// Criar contexto de mercado
			marketCtx := adapter.NewMarketContext(config.market, config.tenantType, config.hookType)
			
			// Act - Atualizar contadores de elevação ativa
			obs.UpdateActiveElevations(marketCtx, 5)
			
			// Registrar cobertura de testes
			obs.RecordTestCoverage(config.hookType, 95.5)
			
			// Assert - Verificar que as operações não causaram erros
			// (Neste teste, estamos validando que as funções não causam exceções)
		})
	}
}