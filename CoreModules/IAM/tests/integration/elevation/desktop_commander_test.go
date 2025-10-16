// Package elevation contém testes de integração para o sistema de elevação de privilégios MCP-IAM.
// Este arquivo testa o fluxo de elevação de privilégios com o hook Desktop Commander.
package elevation

import (
	"context"
	"testing"
	"time"

	"github.com/innovabiz/iam/auth"
	"github.com/innovabiz/iam/elevation"
	"github.com/innovabiz/iam/mcp/hooks"
	"github.com/innovabiz/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDesktopCommanderElevation testa o fluxo completo de elevação para o hook Desktop Commander
func TestDesktopCommanderElevation(t *testing.T) {
	span, ctx := testutil.StartTestSpan(context.Background(), "TestDesktopCommanderElevation")
	defer span.Finish()
	
	logger := testutil.GetTestLogger()

	// Configura serviço de elevação para testes
	elevationService, err := setupElevationService()
	require.NoError(t, err, "Falha ao configurar serviço de elevação")

	// Configura o hook Desktop Commander para testes
	dcHook := hooks.NewDesktopCommanderMCPHook(elevationService, logger)
	
	// Testes para diferentes mercados e cenários
	testMarkets := []struct {
		name      string
		marketCode string
		tenantID  string
		commands  []hooks.DCCommandRequest
	}{
		{
			name:      "Angola",
			marketCode: "angola",
			tenantID:  "tenant-angola-123",
			commands: []hooks.DCCommandRequest{
				{
					Command:  "execute_command",
					Args:     map[string]interface{}{"command": "sudo apt update", "cwd": "/home/user"},
					CommandType: "sudo",
					UserID:   "user-admin-456",
					TenantID: "tenant-angola-123",
				},
				{
					Command:  "edit_config",
					Args:     map[string]interface{}{"key": "allowedDirectories", "value": []string{}},
					CommandType: "system_config",
					UserID:   "user-admin-456",
					TenantID: "tenant-angola-123",
				},
			},
		},
		{
			name:      "Brasil",
			marketCode: "brasil",
			tenantID:  "tenant-brasil-456",
			commands: []hooks.DCCommandRequest{
				{
					Command:  "execute_command",
					Args:     map[string]interface{}{"command": "sudo systemctl restart postgresql", "cwd": "/etc/postgresql"},
					CommandType: "sudo",
					UserID:   "user-admin-456",
					TenantID: "tenant-brasil-456",
				},
			},
		},
		{
			name:      "Moçambique",
			marketCode: "mocambique",
			tenantID:  "tenant-mocambique-789",
			commands: []hooks.DCCommandRequest{
				{
					Command:  "execute_command",
					Args:     map[string]interface{}{"command": "sudo rm -rf /var/log/old/*", "cwd": "/var/log"},
					CommandType: "sudo",
					UserID:   "user-admin-456",
					TenantID: "tenant-mocambique-789",
				},
			},
		},
	}
	
	for _, market := range testMarkets {
		t.Run(market.name, func(t *testing.T) {
			for _, cmd := range market.commands {
				testCommandElevation(t, ctx, elevationService, dcHook, cmd, market.marketCode, market.tenantID)
			}
		})
	}
	
	// Teste para comandos que não requerem elevação
	t.Run("ComandosSemElevacao", func(t *testing.T) {
		safeCommands := []hooks.DCCommandRequest{
			{
				Command:  "list_directory",
				Args:     map[string]interface{}{"path": "/home/user"},
				CommandType: "file",
				UserID:   "user-admin-456",
				TenantID: "tenant-angola-123",
			},
			{
				Command:  "read_file",
				Args:     map[string]interface{}{"path": "/home/user/document.txt"},
				CommandType: "file",
				UserID:   "user-admin-456",
				TenantID: "tenant-angola-123",
			},
		}
		
		for _, cmd := range safeCommands {
			testCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			
			// Cria token de autenticação para teste
			authToken := createTestAuthToken(t, cmd.UserID, cmd.TenantID, "angola")
			testCtx = auth.ContextWithToken(testCtx, authToken)
			
			// Executa comando que não requer elevação
			response, err := dcHook.ProcessCommand(testCtx, cmd)
			
			// Não deve exigir elevação
			assert.NoError(t, err, "Comando seguro não deveria exigir elevação")
			assert.NotNil(t, response, "Resposta não deveria ser nil para comando seguro")
			assert.True(t, response.Allowed, "Comando seguro deveria ser permitido")
		}
	})
}

// testCommandElevation testa um comando específico com fluxo de elevação
func testCommandElevation(t *testing.T, ctx context.Context, elevationService *elevation.Service, 
	dcHook *hooks.DesktopCommanderMCPHook, command hooks.DCCommandRequest, market, tenantID string) {
	
	userID := command.UserID
	
	// Cenário 1: Tentativa sem elevação (deve ser negado)
	t.Run("DeveSolicitarElevacaoParaComandoSensivel", func(t *testing.T) {
		testCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		
		// Cria token de autenticação para teste
		authToken := createTestAuthToken(t, userID, tenantID, market)
		testCtx = auth.ContextWithToken(testCtx, authToken)
		
		// Tenta executar comando sem elevação
		response, err := dcHook.ProcessCommand(testCtx, command)
		
		// Deve retornar erro de elevação necessária para comandos sensíveis
		if isCommandSensitive(command) {
			require.Error(t, err, "Deveria exigir elevação para comando sensível")
			assert.Contains(t, err.Error(), "elevation required", "Erro deveria indicar necessidade de elevação")
			assert.Nil(t, response, "Resposta deve ser nil quando elevação é necessária")
		} else {
			assert.NoError(t, err, "Comando não sensível não deveria exigir elevação")
			assert.NotNil(t, response, "Resposta não deveria ser nil para comando não sensível")
			assert.True(t, response.Allowed, "Comando não sensível deveria ser permitido")
		}
	})
	
	// Se o comando for sensível, teste o fluxo completo de elevação
	if isCommandSensitive(command) {
		// Cenário 2: Solicita elevação e aprovação
		t.Run("DevePermitirComandoAposElevacao", func(t *testing.T) {
			testCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			
			// Cria token de autenticação para teste
			authToken := createTestAuthToken(t, userID, tenantID, market)
			
			// Obtém escopo de elevação necessário para o comando
			scope := getElevationScopeForCommand(command)
			
			// Solicita elevação de privilégios
			elevationRequest := &elevation.ElevationRequest{
				UserID:        userID,
				TenantID:      tenantID,
				Justification: "Manutenção do sistema em ambiente de testes",
				Scopes:        []string{scope},
				Duration:      30, // 30 minutos
				Emergency:     false,
				Market:        market,
			}
			
			// Contexto com token de autenticação
			requestCtx := auth.ContextWithToken(testCtx, authToken)
			
			// Submete solicitação de elevação
			elevationToken, err := elevationService.RequestElevation(requestCtx, elevationRequest)
			require.NoError(t, err, "Falha ao solicitar elevação")
			assert.NotNil(t, elevationToken, "Token de elevação não deveria ser nil")
			
			// Verifica se requer aprovação com base no mercado/política
			if requiresApproval(market, scope) {
				assert.Equal(t, elevation.StatusPendingApproval, elevationToken.Status, "Status deve ser pendente aprovação")
				
				// Simula aprovação por supervisor
				supervisorID := "user-supervisor-789"
				supervisorCtx := createSupervisorContext(testCtx, supervisorID, tenantID, market)
				
				approvedToken, err := elevationService.ApproveElevation(supervisorCtx, elevationToken.ID)
				require.NoError(t, err, "Falha ao aprovar elevação")
				
				// Atualiza token após aprovação
				elevationToken = approvedToken
			}
			
			// Verifica status ativo
			assert.Equal(t, elevation.StatusActive, elevationToken.Status, "Status deve ser ativo após aprovação")
			
			// Adiciona token de elevação ao contexto
			elevCtx := elevation.ContextWithElevationToken(testCtx, elevationToken)
			elevCtx = auth.ContextWithToken(elevCtx, authToken)
			
			// Tenta novamente o comando com contexto de elevação
			response, err := dcHook.ProcessCommand(elevCtx, command)
			
			// Deve ser permitido com elevação
			require.NoError(t, err, "Comando deveria ser permitido com elevação")
			assert.NotNil(t, response, "Resposta não deveria ser nil quando comando é permitido")
			assert.True(t, response.Allowed, "Comando deveria ser permitido")
			
			// Registra evento de auditoria para análise posterior
			testutil.LogAuditEvent(testCtx, "desktop_command_execution", map[string]interface{}{
				"tenant_id":   tenantID,
				"user_id":     userID,
				"market":      market,
				"command":     command.Command,
				"command_type": command.CommandType,
				"elevation_id": elevationToken.ID,
				"approved":    true,
			})
		})
	}
}

// isCommandSensitive verifica se um comando requer elevação de privilégios
func isCommandSensitive(command hooks.DCCommandRequest) bool {
	// Comandos que sempre requerem elevação
	sensitiveCommands := map[string]bool{
		"execute_command": true,
		"edit_config":     true,
		"force_terminate": true,
	}
	
	// Comandos que dependem do tipo para requerer elevação
	if command.Command == "execute_command" {
		cmdType := command.CommandType
		return cmdType == "sudo" || cmdType == "system"
	}
	
	// Verifica comandos sensíveis por padrão
	return sensitiveCommands[command.Command]
}

// getElevationScopeForCommand determina o escopo de elevação necessário para um comando
func getElevationScopeForCommand(command hooks.DCCommandRequest) string {
	// Mapeamento de comandos para escopos de elevação
	commandScopes := map[string]string{
		"execute_command": "desktop:execute_command:sudo",
		"edit_config":     "desktop:edit_config",
		"force_terminate": "desktop:force_terminate",
	}
	
	// Se houver um mapeamento específico para o comando, use-o
	if scope, exists := commandScopes[command.Command]; exists {
		return scope
	}
	
	// Escopo padrão para comandos não mapeados explicitamente
	return "desktop:command:" + command.Command
}

// requiresApproval verifica se a elevação para o mercado/escopo requer aprovação
func requiresApproval(market, scope string) bool {
	// Em um cenário real, isso consultaria o banco de dados ou cache
	// Para simplificar, usamos regras hard-coded para testes
	
	// Angola e Brasil sempre requerem aprovação
	if market == "angola" || market == "brasil" {
		return true
	}
	
	// Moçambique só requer para comandos críticos
	if market == "mocambique" {
		criticalScopes := map[string]bool{
			"desktop:execute_command:sudo": true,
		}
		return criticalScopes[scope]
	}
	
	// Por padrão, requer aprovação para garantir segurança
	return true
}