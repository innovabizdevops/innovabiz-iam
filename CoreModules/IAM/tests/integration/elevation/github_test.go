// Package elevation contém testes de integração para o sistema de elevação de privilégios MCP-IAM.
// Este arquivo testa o fluxo de elevação de privilégios com o hook GitHub.
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
	"go.uber.org/zap"
)

// TestGitHubElevation testa o fluxo completo de elevação para o hook GitHub
func TestGitHubElevation(t *testing.T) {
	span, ctx := testutil.StartTestSpan(context.Background(), "TestGitHubElevation")
	defer span.Finish()
	
	logger := testutil.GetTestLogger()
	logger.Info("Iniciando testes de elevação para hook GitHub",
		zap.String("test", "github_elevation"))

	// Configura serviço de elevação para testes
	elevationService, err := setupElevationService()
	require.NoError(t, err, "Falha ao configurar serviço de elevação")

	// Configura o hook GitHub para testes
	githubHook := hooks.NewGitHubMCPHook(elevationService, logger)
	
	// Define cenários de teste por mercado
	testScenarios := []struct {
		name          string
		marketCode    string
		tenantID      string
		userID        string
		operations    []hooks.GitHubOperation
		description   string
		expectElevation bool
	}{
		{
			name:       "Angola - Proteção de Branch Principal",
			marketCode: "angola",
			tenantID:   "tenant-angola-123",
			userID:     "user-admin-456",
			operations: []hooks.GitHubOperation{
				{
					Type:       "merge_pull_request",
					RepoOwner:  "innovabiz",
					RepoName:   "payment-gateway",
					BaseBranch: "main",
					HeadBranch: "feature/nova-integracao",
					PullNumber: 123,
				},
				{
					Type:       "push_force",
					RepoOwner:  "innovabiz",
					RepoName:   "payment-gateway",
					Branch:     "main",
				},
			},
			description:    "Operações sensíveis em branches protegidos para Angola",
			expectElevation: true,
		},
		{
			name:       "Brasil - Proteção de Repositório Crítico",
			marketCode: "brasil",
			tenantID:   "tenant-brasil-456",
			userID:     "user-admin-456",
			operations: []hooks.GitHubOperation{
				{
					Type:       "delete_branch",
					RepoOwner:  "innovabiz",
					RepoName:   "payment-gateway",
					Branch:     "main",
				},
				{
					Type:       "push_force",
					RepoOwner:  "innovabiz",
					RepoName:   "payment-gateway",
					Branch:     "release/v1.0",
				},
			},
			description:    "Operações críticas em repositórios de pagamento para Brasil",
			expectElevation: true,
		},
		{
			name:       "Moçambique - Operações Normais",
			marketCode: "mocambique",
			tenantID:   "tenant-mocambique-789",
			userID:     "user-admin-456",
			operations: []hooks.GitHubOperation{
				{
					Type:       "merge_pull_request",
					RepoOwner:  "innovabiz",
					RepoName:   "docs",
					BaseBranch: "main",
					HeadBranch: "feature/nova-documentacao",
					PullNumber: 45,
				},
				{
					Type:       "create_repository",
					RepoOwner:  "innovabiz",
					RepoName:   "novo-repo-teste",
				},
			},
			description:    "Operações em repositórios não críticos para Moçambique",
			expectElevation: false,
		},
	}
	
	// Executa testes para cada cenário
	for _, scenario := range testScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			testCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()
			
			// Gera token de autenticação para o usuário
			authToken := createTestAuthToken(t, scenario.userID, scenario.tenantID, scenario.marketCode)
			testCtx = auth.ContextWithToken(testCtx, authToken)
			
			// Testa cada operação no cenário
			for _, operation := range scenario.operations {
				testGitHubOperation(t, testCtx, elevationService, githubHook, operation, 
					scenario.marketCode, scenario.tenantID, scenario.userID, scenario.expectElevation)
			}
		})
	}
	
	// Testes específicos para branch protection
	t.Run("ProteçãoDeBranchPersonalizadaPorMercado", func(t *testing.T) {
		// Testa regras específicas por mercado para proteção de branch
		branchProtectionTests := []struct {
			market     string
			tenantID   string
			operation  hooks.GitHubOperation
			expectElevation bool
			description string
		}{
			{
				market:     "angola",
				tenantID:   "tenant-angola-123",
				operation: hooks.GitHubOperation{
					Type:       "push_force",
					RepoOwner:  "innovabiz",
					RepoName:   "core-banking",
					Branch:     "production",
				},
				expectElevation: true,
				description:    "Angola requer elevação para force push em branch de produção",
			},
			{
				market:     "brasil",
				tenantID:   "tenant-brasil-456",
				operation: hooks.GitHubOperation{
					Type:       "push_force",
					RepoOwner:  "innovabiz",
					RepoName:   "core-banking",
					Branch:     "develop",
				},
				expectElevation: false,
				description:    "Brasil não requer elevação para force push em branch de desenvolvimento",
			},
			{
				market:     "mocambique",
				tenantID:   "tenant-mocambique-789",
				operation: hooks.GitHubOperation{
					Type:       "push_force",
					RepoOwner:  "innovabiz",
					RepoName:   "mobile-money",
					Branch:     "main",
				},
				expectElevation: true,
				description:    "Moçambique requer elevação para force push em branch principal de mobile-money",
			},
		}
		
		for _, test := range branchProtectionTests {
			t.Run(test.description, func(t *testing.T) {
				testCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				
				// Gera token de autenticação para o usuário
				authToken := createTestAuthToken(t, "user-admin-456", test.tenantID, test.market)
				testCtx = auth.ContextWithToken(testCtx, authToken)
				
				testGitHubOperation(t, testCtx, elevationService, githubHook, test.operation, 
					test.market, test.tenantID, "user-admin-456", test.expectElevation)
			})
		}
	})
	
	// Teste para modo de emergência
	t.Run("ElevaçãoEmergência", func(t *testing.T) {
		testCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		
		market := "angola"
		tenantID := "tenant-angola-123"
		userID := "user-admin-456"
		
		// Operação crítica que normalmente requer aprovação
		operation := hooks.GitHubOperation{
			Type:       "delete_repository",
			RepoOwner:  "innovabiz",
			RepoName:   "payment-gateway",
		}
		
		// Gera token de autenticação
		authToken := createTestAuthToken(t, userID, tenantID, market)
		testCtx = auth.ContextWithToken(testCtx, authToken)
		
		// Tenta operação sem elevação (deve ser negada)
		result, err := githubHook.ProcessOperation(testCtx, operation)
		require.Error(t, err, "Operação crítica deve ser negada sem elevação")
		assert.Nil(t, result, "Resultado deve ser nil quando operação é negada")
		
		// Solicita elevação de emergência (sem aprovação)
		elevationRequest := &elevation.ElevationRequest{
			UserID:        userID,
			TenantID:      tenantID,
			Market:        market,
			Justification: "EMERGÊNCIA: Remoção de repositório com credenciais expostas",
			Scopes:        []string{"github:delete:repository"},
			Duration:      60, // 60 minutos
			Emergency:     true, // Modo de emergência
		}
		
		// Solicita elevação de emergência
		elevToken, err := elevationService.RequestElevation(testCtx, elevationRequest)
		require.NoError(t, err, "Falha ao solicitar elevação de emergência")
		
		// Verifica se o token foi ativado automaticamente por ser emergência
		assert.Equal(t, elevation.StatusActive, elevToken.Status, "Tokens de emergência devem ser ativados automaticamente")
		
		// Adiciona token de elevação ao contexto
		elevCtx := elevation.ContextWithElevationToken(testCtx, elevToken)
		elevCtx = auth.ContextWithToken(elevCtx, authToken)
		
		// Tenta novamente a operação com elevação de emergência
		result, err = githubHook.ProcessOperation(elevCtx, operation)
		require.NoError(t, err, "Operação crítica deve ser permitida com elevação de emergência")
		assert.NotNil(t, result, "Resultado não deve ser nil quando operação é permitida")
		assert.True(t, result.Allowed, "Operação deve ser permitida com elevação de emergência")
		assert.Contains(t, result.Message, "emergency", "Mensagem deve indicar que foi aprovação de emergência")
		
		// Registra evento de auditoria especial para elevação de emergência
		testutil.LogAuditEvent(testCtx, "emergency_elevation_used", map[string]interface{}{
			"tenant_id":      tenantID,
			"user_id":        userID,
			"market":         market,
			"operation":      operation.Type,
			"repo":           operation.RepoName,
			"elevation_id":   elevToken.ID,
			"justification": elevToken.Justification,
		})
	})
}

// testGitHubOperation testa uma operação específica do GitHub
func testGitHubOperation(t *testing.T, ctx context.Context, elevationService *elevation.Service, 
	githubHook *hooks.GitHubMCPHook, operation hooks.GitHubOperation, 
	market, tenantID, userID string, expectElevation bool) {
	
	// Teste sem elevação
	t.Run("SemElevação", func(t *testing.T) {
		// Tenta operação sem elevação
		result, err := githubHook.ProcessOperation(ctx, operation)
		
		if expectElevation {
			// Se espera-se que precise de elevação
			require.Error(t, err, "Operação sensível deve ser negada sem elevação")
			assert.Contains(t, err.Error(), "elevation required", "Erro deve indicar que elevação é necessária")
			assert.Nil(t, result, "Resultado deve ser nil quando elevação é necessária")
		} else {
			// Se não precisa de elevação
			require.NoError(t, err, "Operação não sensível não deve requerer elevação")
			assert.NotNil(t, result, "Resultado não deve ser nil para operação permitida")
			assert.True(t, result.Allowed, "Operação deve ser permitida sem elevação")
		}
	})
	
	// Se a operação requer elevação, teste o fluxo completo
	if expectElevation {
		t.Run("ComElevação", func(t *testing.T) {
			// Determina escopo de elevação necessário
			scope := getGitHubElevationScope(operation)
			
			// Solicita elevação de privilégios
			elevationRequest := &elevation.ElevationRequest{
				UserID:        userID,
				TenantID:      tenantID,
				Market:        market,
				Justification: "Teste de operação GitHub com elevação",
				Scopes:        []string{scope},
				Duration:      30, // 30 minutos
				Emergency:     false,
			}
			
			// Submete solicitação de elevação
			elevToken, err := elevationService.RequestElevation(ctx, elevationRequest)
			require.NoError(t, err, "Falha ao solicitar elevação")
			
			// Verifica se requer aprovação com base no mercado/política
			if requiresGitHubApproval(market, scope) {
				assert.Equal(t, elevation.StatusPendingApproval, elevToken.Status, "Status deve ser pendente aprovação")
				
				// Simula aprovação por supervisor
				supervisorID := "user-supervisor-789"
				supervisorCtx := createSupervisorContext(ctx, supervisorID, tenantID, market)
				
				approvedToken, err := elevationService.ApproveElevation(supervisorCtx, elevToken.ID)
				require.NoError(t, err, "Falha ao aprovar elevação")
				
				// Atualiza token após aprovação
				elevToken = approvedToken
			}
			
			// Verifica status ativo
			assert.Equal(t, elevation.StatusActive, elevToken.Status, "Status deve ser ativo após aprovação")
			
			// Adiciona token de elevação ao contexto
			elevCtx := elevation.ContextWithElevationToken(ctx, elevToken)
			elevCtx = auth.ContextWithToken(elevCtx, createTestAuthToken(t, userID, tenantID, market))
			
			// Tenta novamente a operação com elevação
			result, err := githubHook.ProcessOperation(elevCtx, operation)
			require.NoError(t, err, "Operação deve ser permitida com elevação")
			assert.NotNil(t, result, "Resultado não deve ser nil quando operação é permitida")
			assert.True(t, result.Allowed, "Operação deve ser permitida com elevação")
			
			// Registra evento de auditoria para análise posterior
			testutil.LogAuditEvent(ctx, "github_operation_execution", map[string]interface{}{
				"tenant_id":   tenantID,
				"user_id":     userID,
				"market":      market,
				"operation":   operation.Type,
				"repo":        operation.RepoName,
				"elevation_id": elevToken.ID,
				"approved":    true,
			})
		})
	}
}

// getGitHubElevationScope determina o escopo de elevação necessário para uma operação
func getGitHubElevationScope(operation hooks.GitHubOperation) string {
	// Mapeamento de operações para escopos de elevação
	operationScopes := map[string]string{
		"merge_pull_request": "github:merge:main",
		"push_force":         "github:push:force",
		"delete_branch":      "github:delete:branch",
		"delete_repository":  "github:delete:repository",
	}
	
	// Se houver um mapeamento específico para a operação, use-o
	if scope, exists := operationScopes[operation.Type]; exists {
		return scope
	}
	
	// Escopo padrão para operações não mapeadas explicitamente
	return "github:operation:" + operation.Type
}

// requiresGitHubApproval verifica se a elevação para GitHub requer aprovação por mercado/escopo
func requiresGitHubApproval(market, scope string) bool {
	// Em um cenário real, isso consultaria o banco de dados ou cache
	// Para simplificar, usamos regras hard-coded para testes
	
	// Operações críticas sempre requerem aprovação independente do mercado
	criticalScopes := map[string]bool{
		"github:delete:repository": true,
		"github:push:force":        true,
	}
	
	if criticalScopes[scope] {
		return true
	}
	
	// Angola sempre requer aprovação para qualquer operação de GitHub
	if market == "angola" {
		return true
	}
	
	// Brasil requer aprovação apenas para operações em branches principais
	if market == "brasil" && (scope == "github:merge:main" || scope == "github:delete:branch") {
		return true
	}
	
	// Moçambique tem requisitos mais flexíveis
	if market == "mocambique" {
		return scope == "github:delete:repository" // Apenas deleção de repositório requer aprovação
	}
	
	// Por padrão, requer aprovação para garantir segurança
	return true
}