// Package elevation contém testes de integração para o sistema de elevação de privilégios MCP-IAM.
// Este arquivo implementa a validação de eventos de auditoria gerados durante os testes de elevação.
package elevation

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/innovabiz/iam/audit"
	"github.com/innovabiz/iam/elevation"
	"github.com/innovabiz/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestAuditValidation valida os eventos de auditoria gerados pelos testes de elevação
func TestAuditValidation(t *testing.T) {
	span, ctx := testutil.StartTestSpan(context.Background(), "TestAuditValidation")
	defer span.Finish()
	
	logger := testutil.GetTestLogger()
	logger.Info("Iniciando validação de eventos de auditoria para testes de elevação",
		zap.String("test", "audit_validation"))

	// Configura serviço de elevação e auditoria para testes
	elevationService, err := setupElevationService()
	require.NoError(t, err, "Falha ao configurar serviço de elevação")
	
	// Obtém referência ao mock do serviço de auditoria
	auditService := elevationService.GetAuditService().(*audit.MockAuditService)
	require.NotNil(t, auditService, "Serviço de auditoria mock não configurado corretamente")
	
	// Testes de validação de eventos por categoria
	t.Run("ValidaEventosElevação", func(t *testing.T) {
		validateElevationEvents(t, ctx, auditService)
	})
	
	t.Run("ValidaEventosDockerMCP", func(t *testing.T) {
		validateDockerEvents(t, ctx, auditService)
	})
	
	t.Run("ValidaEventosDesktopCommanderMCP", func(t *testing.T) {
		validateDesktopCommanderEvents(t, ctx, auditService)
	})
	
	t.Run("ValidaEventosGitHubMCP", func(t *testing.T) {
		validateGitHubEvents(t, ctx, auditService)
	})
	
	t.Run("ValidaEventosFigmaMCP", func(t *testing.T) {
		validateFigmaEvents(t, ctx, auditService)
	})
	
	t.Run("ValidaEventosMultiMercado", func(t *testing.T) {
		validateMultiMarketEvents(t, ctx, auditService)
	})
	
	t.Run("ValidaEventosAdministrativos", func(t *testing.T) {
		validateAdministrativeEvents(t, ctx, auditService)
	})
}

// validateElevationEvents valida os eventos de auditoria relacionados ao ciclo de vida da elevação
func validateElevationEvents(t *testing.T, ctx context.Context, auditService *audit.MockAuditService) {
	// Obtém eventos de auditoria relacionados ao ciclo de vida de elevação
	events := auditService.GetEventsByType([]string{
		"elevation_requested",
		"elevation_approved",
		"elevation_denied",
		"elevation_expired",
		"elevation_revoked",
		"elevation_used",
	})
	
	// Verifica se existem eventos básicos registrados
	require.NotEmpty(t, events, "Devem existir eventos de ciclo de vida de elevação registrados")
	
	// Validação por tipo de evento
	eventCounts := map[string]int{}
	for _, event := range events {
		eventCounts[event.Type]++
	}
	
	// Deve haver pelo menos uma solicitação de elevação
	assert.GreaterOrEqual(t, eventCounts["elevation_requested"], 1, 
		"Deve haver pelo menos uma solicitação de elevação")
	
	// Valida eventos de solicitação de elevação
	for _, event := range filterEventsByType(events, "elevation_requested") {
		// Campos obrigatórios
		assert.NotEmpty(t, event.UserID, "UserID deve estar presente em eventos de solicitação")
		assert.NotEmpty(t, event.TenantID, "TenantID deve estar presente em eventos de solicitação")
		assert.NotEmpty(t, event.Market, "Market deve estar presente em eventos de solicitação")
		
		// Campos específicos
		assert.NotEmpty(t, event.Metadata["elevation_id"], "ID de elevação deve estar presente")
		assert.NotEmpty(t, event.Metadata["scopes"], "Escopos devem estar presentes")
		assert.NotEmpty(t, event.Metadata["justification"], "Justificativa deve estar presente")
	}
	
	// Valida eventos de aprovação de elevação
	for _, event := range filterEventsByType(events, "elevation_approved") {
		// Campos básicos
		assert.NotEmpty(t, event.UserID, "UserID deve estar presente em eventos de aprovação")
		assert.NotEmpty(t, event.TenantID, "TenantID deve estar presente em eventos de aprovação")
		
		// Campos específicos
		assert.NotEmpty(t, event.Metadata["elevation_id"], "ID de elevação deve estar presente")
		assert.NotEmpty(t, event.Metadata["approver_id"], "ID do aprovador deve estar presente")
	}
	
	// Valida eventos de uso de elevação
	for _, event := range filterEventsByType(events, "elevation_used") {
		// Campos básicos
		assert.NotEmpty(t, event.UserID, "UserID deve estar presente em eventos de uso")
		assert.NotEmpty(t, event.TenantID, "TenantID deve estar presente em eventos de uso")
		
		// Campos específicos
		assert.NotEmpty(t, event.Metadata["elevation_id"], "ID de elevação deve estar presente")
		assert.NotEmpty(t, event.Metadata["operation"], "Operação deve estar presente")
		assert.NotEmpty(t, event.Metadata["scope_used"], "Escopo usado deve estar presente")
	}
	
	// Verifica fluxos completos (solicitação → aprovação → uso)
	validateElevationFlows(t, events)
}

// validateElevationFlows verifica se há fluxos completos de elevação nos eventos
func validateElevationFlows(t *testing.T, events []audit.Event) {
	// Mapeia eventos por ID de elevação
	elevationEvents := map[string]map[string]bool{}
	
	for _, event := range events {
		elevationID, ok := event.Metadata["elevation_id"].(string)
		if !ok || elevationID == "" {
			continue
		}
		
		if elevationEvents[elevationID] == nil {
			elevationEvents[elevationID] = make(map[string]bool)
		}
		elevationEvents[elevationID][event.Type] = true
	}
	
	// Conta fluxos completos
	completeFlows := 0
	for elevationID, eventTypes := range elevationEvents {
		if eventTypes["elevation_requested"] && 
			(eventTypes["elevation_approved"] || eventTypes["elevation_emergency_auto_approved"]) && 
			eventTypes["elevation_used"] {
			completeFlows++
			t.Logf("Fluxo completo detectado para elevação ID: %s", elevationID)
		}
	}
	
	// Deve haver pelo menos um fluxo completo
	assert.GreaterOrEqual(t, completeFlows, 1, 
		"Deve haver pelo menos um fluxo completo de elevação (solicitação → aprovação → uso)")
}

// validateDockerEvents valida os eventos de auditoria relacionados ao hook Docker
func validateDockerEvents(t *testing.T, ctx context.Context, auditService *audit.MockAuditService) {
	// Obtém eventos específicos do Docker
	events := auditService.GetEventsByType([]string{
		"docker_command_requested",
		"docker_command_elevated",
		"docker_command_denied",
		"docker_command_executed",
	})
	
	// Verifica se existem eventos Docker registrados
	require.NotEmpty(t, events, "Devem existir eventos Docker registrados")
	
	// Valida eventos de comandos Docker
	for _, event := range events {
		// Campos básicos em todos eventos Docker
		assert.NotEmpty(t, event.UserID, "UserID deve estar presente em eventos Docker")
		assert.NotEmpty(t, event.TenantID, "TenantID deve estar presente em eventos Docker")
		assert.NotEmpty(t, event.Market, "Market deve estar presente em eventos Docker")
		
		// Campos específicos para comandos
		assert.NotEmpty(t, event.Metadata["command"], "Comando deve estar presente")
		
		if event.Type == "docker_command_elevated" {
			assert.NotEmpty(t, event.Metadata["elevation_id"], "ID de elevação deve estar presente")
		}
	}
	
	// Verifica comandos sensíveis específicos
	sensitiveCommands := []string{"exec", "run", "system", "prune"}
	foundSensitiveCommand := false
	
	for _, event := range events {
		cmd, ok := event.Metadata["command"].(string)
		if !ok {
			continue
		}
		
		for _, sensitive := range sensitiveCommands {
			if stringContains(cmd, sensitive) {
				foundSensitiveCommand = true
				
				// Comandos sensíveis devem ter sido elevados ou negados
				if event.Type == "docker_command_executed" {
					elevationID, hasElevation := event.Metadata["elevation_id"].(string)
					assert.True(t, hasElevation && elevationID != "", 
						"Comando sensível deve ter ID de elevação quando executado: %s", cmd)
				}
				
				break
			}
		}
	}
	
	assert.True(t, foundSensitiveCommand, 
		"Deve existir pelo menos um comando sensível Docker nos eventos de auditoria")
}

// validateDesktopCommanderEvents valida os eventos de auditoria relacionados ao hook Desktop Commander
func validateDesktopCommanderEvents(t *testing.T, ctx context.Context, auditService *audit.MockAuditService) {
	// Obtém eventos específicos do Desktop Commander
	events := auditService.GetEventsByType([]string{
		"dc_command_requested",
		"dc_command_elevated",
		"dc_command_denied",
		"dc_command_executed",
		"dc_file_operation",
	})
	
	// Verifica se existem eventos Desktop Commander registrados
	require.NotEmpty(t, events, "Devem existir eventos Desktop Commander registrados")
	
	// Valida eventos de operações no sistema de arquivos
	fileEvents := filterEventsByType(events, "dc_file_operation")
	assert.NotEmpty(t, fileEvents, "Devem existir eventos de operações em arquivos")
	
	for _, event := range fileEvents {
		assert.NotEmpty(t, event.Metadata["operation"], "Tipo de operação deve estar presente")
		assert.NotEmpty(t, event.Metadata["path"], "Caminho do arquivo deve estar presente")
	}
	
	// Valida eventos de comandos executados
	cmdEvents := filterEventsByType(events, "dc_command_executed")
	assert.NotEmpty(t, cmdEvents, "Devem existir eventos de comandos executados")
	
	for _, event := range cmdEvents {
		assert.NotEmpty(t, event.Metadata["command"], "Comando deve estar presente")
		assert.NotEmpty(t, event.UserID, "UserID deve estar presente")
		assert.NotEmpty(t, event.TenantID, "TenantID deve estar presente")
		assert.NotEmpty(t, event.Market, "Market deve estar presente")
	}
	
	// Verifica operações sensíveis específicas
	sensitiveOps := []string{"write_file", "edit_block", "execute_command", "start_process"}
	foundSensitiveOp := false
	
	for _, event := range events {
		op, ok := event.Metadata["operation"].(string)
		if !ok {
			cmd, ok := event.Metadata["command"].(string)
			if ok {
				op = cmd
			} else {
				continue
			}
		}
		
		for _, sensitive := range sensitiveOps {
			if stringContains(op, sensitive) {
				foundSensitiveOp = true
				break
			}
		}
	}
	
	assert.True(t, foundSensitiveOp, 
		"Deve existir pelo menos uma operação sensível Desktop Commander nos eventos de auditoria")
}

// validateGitHubEvents valida os eventos de auditoria relacionados ao hook GitHub
func validateGitHubEvents(t *testing.T, ctx context.Context, auditService *audit.MockAuditService) {
	// Obtém eventos específicos do GitHub
	events := auditService.GetEventsByType([]string{
		"github_operation_requested",
		"github_operation_elevated",
		"github_operation_denied",
		"github_operation_execution",
		"emergency_elevation_used",
	})
	
	// Verifica se existem eventos GitHub registrados
	require.NotEmpty(t, events, "Devem existir eventos GitHub registrados")
	
	// Valida operações GitHub executadas
	for _, event := range filterEventsByType(events, "github_operation_execution") {
		assert.NotEmpty(t, event.Metadata["operation"], "Tipo de operação deve estar presente")
		assert.NotEmpty(t, event.Metadata["elevation_id"], "ID de elevação deve estar presente")
		assert.NotEmpty(t, event.UserID, "UserID deve estar presente")
		assert.NotEmpty(t, event.TenantID, "TenantID deve estar presente")
	}
	
	// Verifica eventos de operações de emergência
	emergencyEvents := filterEventsByType(events, "emergency_elevation_used")
	for _, event := range emergencyEvents {
		assert.NotEmpty(t, event.Metadata["justification"], "Justificativa de emergência deve estar presente")
		assert.NotEmpty(t, event.Metadata["elevation_id"], "ID de elevação deve estar presente")
		
		// Verificações adicionais de segurança para eventos de emergência
		assert.NotEmpty(t, event.ClientIP, "IP do cliente deve estar presente em elevações de emergência")
		assert.NotEmpty(t, event.Metadata["user_agent"], "User-Agent deve estar presente em elevações de emergência")
	}
}

// validateFigmaEvents valida os eventos de auditoria relacionados ao hook Figma
func validateFigmaEvents(t *testing.T, ctx context.Context, auditService *audit.MockAuditService) {
	// Obtém eventos específicos do Figma
	events := auditService.GetEventsByType([]string{
		"figma_operation_requested",
		"figma_operation_elevated",
		"figma_operation_denied",
		"figma_operation_execution",
		"figma_batch_operation",
	})
	
	// Verifica se existem eventos Figma registrados
	require.NotEmpty(t, events, "Devem existir eventos Figma registrados")
	
	// Valida operações Figma executadas
	for _, event := range filterEventsByType(events, "figma_operation_execution") {
		assert.NotEmpty(t, event.Metadata["operation"], "Tipo de operação deve estar presente")
		assert.NotEmpty(t, event.Metadata["file_key"], "Chave do arquivo deve estar presente")
		assert.NotEmpty(t, event.UserID, "UserID deve estar presente")
		assert.NotEmpty(t, event.TenantID, "TenantID deve estar presente")
	}
	
	// Verifica operações em lote
	batchEvents := filterEventsByType(events, "figma_batch_operation")
	if len(batchEvents) > 0 {
		// Verificações para operações em lote
		for _, event := range batchEvents {
			assert.NotEmpty(t, event.Metadata["batch_index"], "Índice da operação em lote deve estar presente")
			assert.NotEmpty(t, event.Metadata["elevation_id"], "ID de elevação deve estar presente")
		}
		
		// Verifica sequência de operações em lote
		validateBatchSequence(t, batchEvents)
	}
}

// validateBatchSequence verifica se uma sequência de operações em lote está correta
func validateBatchSequence(t *testing.T, events []audit.Event) {
	batchIndices := map[string]bool{}
	
	// Extrai índices de lote
	for _, event := range events {
		batchIndex, ok := event.Metadata["batch_index"].(float64)
		if !ok {
			continue
		}
		
		indexStr := fmt.Sprintf("%d", int(batchIndex))
		batchIndices[indexStr] = true
	}
	
	// Deve haver pelo menos uma operação em lote
	assert.NotEmpty(t, batchIndices, "Deve haver pelo menos uma operação em lote")
	
	// Para operações em lote, deve haver múltiplos eventos
	if len(batchIndices) > 0 {
		assert.GreaterOrEqual(t, len(batchIndices), 2, 
			"Deve haver pelo menos duas operações em uma sequência em lote")
	}
}

// validateMultiMarketEvents valida os eventos de auditoria por mercado específico
func validateMultiMarketEvents(t *testing.T, ctx context.Context, auditService *audit.MockAuditService) {
	// Obtém todos os eventos de auditoria
	events := auditService.GetAllEvents()
	
	// Agrupa eventos por mercado
	marketEvents := make(map[string][]audit.Event)
	for _, event := range events {
		if event.Market != "" {
			marketEvents[event.Market] = append(marketEvents[event.Market], event)
		}
	}
	
	// Verifica cobertura dos mercados principais
	requiredMarkets := []string{"angola", "brasil", "mocambique"}
	for _, market := range requiredMarkets {
		marketEvts, exists := marketEvents[market]
		assert.True(t, exists, "Devem existir eventos para o mercado %s", market)
		assert.NotEmpty(t, marketEvts, "Eventos do mercado %s não podem estar vazios", market)
		
		t.Logf("Mercado %s: %d eventos registrados", market, len(marketEvts))
	}
	
	// Verifica regras específicas por mercado
	verifyMarketSpecificRules(t, marketEvents)
}

// verifyMarketSpecificRules valida regras de auditoria específicas por mercado
func verifyMarketSpecificRules(t *testing.T, marketEvents map[string][]audit.Event) {
	// Angola - Verificações específicas conforme BNA
	angolaEvents := marketEvents["angola"]
	if len(angolaEvents) > 0 {
		// Em Angola, todas operações sensíveis devem ter elevação aprovada
		for _, event := range angolaEvents {
			if eventIsSensitiveOperation(event) {
				elevationID, hasElevation := event.Metadata["elevation_id"].(string)
				assert.True(t, hasElevation && elevationID != "", 
					"Angola: Operação sensível deve ter elevação: %s", event.Type)
			}
		}
	}
	
	// Brasil - Verificações específicas conforme LGPD
	brasilEvents := marketEvents["brasil"]
	if len(brasilEvents) > 0 {
		// Brasil exige registro detalhado de processamento de dados
		for _, event := range brasilEvents {
			// Para operações que podem envolver dados pessoais
			if eventInvolvesPII(event) {
				assert.NotEmpty(t, event.Metadata["data_purpose"], 
					"Brasil (LGPD): Operações com dados pessoais devem registrar propósito")
			}
		}
	}
}

// validateAdministrativeEvents valida os eventos administrativos do sistema IAM
func validateAdministrativeEvents(t *testing.T, ctx context.Context, auditService *audit.MockAuditService) {
	// Obtém eventos administrativos
	events := auditService.GetEventsByType([]string{
		"policy_change",
		"role_change",
		"user_access_change",
		"elevation_policy_change",
		"system_startup",
		"system_shutdown",
	})
	
	// Verifica se existem eventos administrativos
	if len(events) > 0 {
		for _, event := range events {
			// Eventos administrativos devem ter usuário administrativo
			assert.NotEmpty(t, event.UserID, "UserID deve estar presente em eventos administrativos")
			
			// Eventos de alteração de política devem ter detalhes
			if event.Type == "elevation_policy_change" || event.Type == "policy_change" {
				assert.NotEmpty(t, event.Metadata["change_details"], 
					"Detalhes da alteração devem estar presentes em mudanças de política")
			}
		}
	}
}

// filterEventsByType filtra eventos por tipo
func filterEventsByType(events []audit.Event, eventType string) []audit.Event {
	var filtered []audit.Event
	for _, event := range events {
		if event.Type == eventType {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// stringContains verifica se uma string contém outra substring
func stringContains(s string, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// eventIsSensitiveOperation determina se um evento representa uma operação sensível
func eventIsSensitiveOperation(event audit.Event) bool {
	sensitiveTypes := map[string]bool{
		"docker_command_executed": true,
		"dc_command_executed": true,
		"github_operation_execution": true,
		"figma_operation_execution": true,
		"elevation_used": true,
	}
	
	if sensitiveTypes[event.Type] {
		return true
	}
	
	// Verificações adicionais para comandos específicos
	if cmd, ok := event.Metadata["command"].(string); ok {
		sensitiveCommands := []string{"rm", "exec", "prune", "delete", "write", "create"}
		for _, sensitive := range sensitiveCommands {
			if stringContains(cmd, sensitive) {
				return true
			}
		}
	}
	
	// Verificação para operações
	if op, ok := event.Metadata["operation"].(string); ok {
		sensitiveOps := []string{"delete", "create", "transfer", "exec"}
		for _, sensitive := range sensitiveOps {
			if stringContains(op, sensitive) {
				return true
			}
		}
	}
	
	return false
}

// eventInvolvesPII determina se um evento pode envolver dados pessoais
func eventInvolvesPII(event audit.Event) bool {
	// Tipos de eventos que podem envolver PII
	piiEventTypes := map[string]bool{
		"docker_command_executed": true, // Pode acessar containers com dados sensíveis
		"dc_file_operation": true,       // Pode manipular arquivos com dados pessoais
		"github_operation_execution": true, // Pode expor credenciais ou dados
	}
	
	if piiEventTypes[event.Type] {
		return true
	}
	
	// Verificações adicionais para comandos específicos
	if cmd, ok := event.Metadata["command"].(string); ok {
		piiCommands := []string{"select", "database", "user", "credential", "secret", "password", "token"}
		for _, piiCmd := range piiCommands {
			if stringContains(cmd, piiCmd) {
				return true
			}
		}
	}
	
	return false
}