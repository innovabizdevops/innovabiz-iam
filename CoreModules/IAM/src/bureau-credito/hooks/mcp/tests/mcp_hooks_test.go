/**
 * @file mcp_hooks_test.go
 * @description Testes para hooks MCP do Bureau de Crédito
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	"innovabiz/iam/src/bureau-credito/hooks"
	"innovabiz/iam/src/bureau-credito/hooks/mcp"
)

// TestMCPAdapterIntegration testa a integração do adaptador MCP com hooks
func TestMCPAdapterIntegration(t *testing.T) {
	// Configurar gerenciador de hooks
	hookManager := hooks.NewHookManager(false)

	// Configurar hooks de teste (mock)
	mockHook := &MockHook{
		ExecutionCount: 0,
		LastHookType:   "",
		LastOperation:  "",
	}
	hookManager.RegisterHook(mockHook)

	// Configurar adaptador MCP
	mcpAdapter := mcp.NewMCPAdapter(hookManager, &mcp.MCPAdapterConfig{
		ServiceName:    "bureau-credito-test",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Region:         "local",
	})

	// Criar evento MCP de teste
	event := mcp.NewMCPEvent("before", "/assessment")
	event.RequestID = "test-request-123"
	event.CorrelationID = "test-correlation-456"
	event.ProviderID = "TEST_PROVIDER"
	event.TenantID = "test-tenant"
	event.UserID = "test-user"
	event.Data = map[string]interface{}{
		"testKey": "testValue",
	}
	event.Metadata = map[string]interface{}{
		"operationType": "assessment",
		"priority":      "high",
	}

	// Processar evento
	err := mcpAdapter.ProcessMCPEvent(context.Background(), event)

	// Verificar resultados
	assert.NoError(t, err)
	assert.Equal(t, 1, mockHook.ExecutionCount)
	assert.Equal(t, string(hooks.HookBefore), mockHook.LastHookType)
	assert.Equal(t, string(hooks.OpAssessment), mockHook.LastOperation)
	assert.Equal(t, "test-request-123", mockHook.LastMetadata.RequestID)
}

// TestFullObservabilityPipeline testa o pipeline completo de observabilidade
func TestFullObservabilityPipeline(t *testing.T) {
	// Criar registro Prometheus
	registry := prometheus.NewRegistry()

	// Configurar gerenciador de hooks
	hookManager := hooks.NewHookManager(false)

	// Registrar hooks reais
	metricsHook := hooks.NewMetricsHook(registry)
	loggingHook := hooks.NewLoggingHook(nil)

	// Registrar hooks no gerenciador
	hookManager.RegisterHook(metricsHook)
	hookManager.RegisterHook(loggingHook)

	// Configurar adaptador MCP
	mcpAdapter := mcp.NewMCPAdapter(hookManager, &mcp.MCPAdapterConfig{
		ServiceName:    "bureau-credito-test",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Region:         "local",
	})

	// Criar evento de início
	beforeEvent := mcp.NewMCPEvent("before", "/assessment")
	beforeEvent.RequestID = "req-123"
	beforeEvent.CorrelationID = "corr-456"
	beforeEvent.ProviderID = "SERASA"
	beforeEvent.TenantID = "tenant-789"
	beforeEvent.UserID = "user-abc"
	beforeEvent.Data = map[string]interface{}{
		"identityInputs": map[string]string{
			"documentNumber": "12345678900",
			"name":          "João da Silva",
		},
	}
	beforeEvent.Metadata = map[string]interface{}{
		"operationType": "assessment",
		"priority":      "high",
	}

	// Processar evento de início
	err := mcpAdapter.ProcessMCPEvent(context.Background(), beforeEvent)
	assert.NoError(t, err)

	// Simular processamento
	time.Sleep(100 * time.Millisecond)

	// Criar evento de conclusão
	afterEvent := mcp.NewMCPEvent("after", "/assessment")
	afterEvent.RequestID = "req-123"
	afterEvent.CorrelationID = "corr-456"
	afterEvent.ProviderID = "SERASA"
	afterEvent.TenantID = "tenant-789"
	afterEvent.UserID = "user-abc"
	afterEvent.Duration = 100 // 100ms
	afterEvent.Data = map[string]interface{}{
		"score": 750,
		"risk":  "low",
		"recommendations": []string{
			"approve",
			"monitor",
		},
	}
	afterEvent.Metadata = map[string]interface{}{
		"operationType": "assessment",
	}

	// Processar evento de conclusão
	err = mcpAdapter.ProcessMCPEvent(context.Background(), afterEvent)
	assert.NoError(t, err)

	// As verificações aqui seriam idealmente mais detalhadas
	// em um ambiente de teste real, verificaríamos as métricas registradas
}

// MockHook é um hook mock para testes
type MockHook struct {
	hooks.BaseHook
	ExecutionCount int
	LastHookType   string
	LastOperation  string
	LastMetadata   hooks.HookMetadata
}

// Execute implementa a interface Hook para testes
func (h *MockHook) Execute(ctx context.Context, hookType hooks.HookType, metadata hooks.HookMetadata, payload interface{}) error {
	h.ExecutionCount++
	h.LastHookType = string(hookType)
	h.LastOperation = string(metadata.OperationType)
	h.LastMetadata = metadata
	return nil
}

// ShouldExecute implementa a interface Hook
func (h *MockHook) ShouldExecute(ctx context.Context, hookType hooks.HookType, metadata hooks.HookMetadata) bool {
	return true
}