// Package elevation contém testes de integração para o sistema de elevação de privilégios MCP-IAM.
// Este arquivo configura o ambiente geral de testes e orquestra a execução integrada.
package elevation

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/innovabiz/iam/auth"
	"github.com/innovabiz/iam/config"
	"github.com/innovabiz/iam/constants"
	"github.com/innovabiz/iam/logging"
	"github.com/innovabiz/iam/metrics"
	"github.com/innovabiz/iam/tenant"
	"github.com/innovabiz/iam/tracing"
	"github.com/innovabiz/tests/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// testMarkets define os mercados para teste multi-dimensional
var testMarkets = []string{
	"angola",   // Mercado com LGPDAO/BNA
	"brasil",   // Mercado com LGPD
	"mocambique", // Mercado com regulações SADC
}

// TestMain configura o ambiente de testes
func TestMain(m *testing.M) {
	// Configura logger para testes
	logger := logging.InitTestLogger()
	
	// Inicia trace global
	tracer, closer := tracing.InitTestTracer("iam-elevation-integration-tests")
	defer closer.Close()
	
	// Configura métricas para testes
	metrics.InitTestMetrics()
	
	// Define variáveis de ambiente para testes
	os.Setenv("IAM_TEST_MODE", "true")
	os.Setenv("IAM_DEFAULT_MARKET", "angola")
	os.Setenv("IAM_MULTI_MARKET_ENABLED", "true")
	os.Setenv("IAM_ELEVATION_AUDIT_LEVEL", "full") // Garante auditoria completa
	
	// Configura containers de teste para PostgreSQL e Redis, se necessário
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	// Inicializa containers de teste
	dbContainer, redisContainer, err := testutil.StartTestContainers(ctx)
	if err != nil {
		logger.Fatal("Falha ao iniciar containers de teste", zap.Error(err))
	}
	defer func() {
		if err := testutil.StopTestContainers(ctx, dbContainer, redisContainer); err != nil {
			logger.Error("Erro ao parar containers de teste", zap.Error(err))
		}
	}()
	
	// Configura banco de dados de teste para o ambiente
	if err := testutil.InitTestDatabase(ctx, dbContainer); err != nil {
		logger.Fatal("Falha ao inicializar banco de dados de teste", zap.Error(err))
	}
	
	// Configura Redis para o ambiente de teste
	if err := testutil.InitTestRedis(ctx, redisContainer); err != nil {
		logger.Fatal("Falha ao inicializar Redis de teste", zap.Error(err))
	}
	
	// Configura e executa testes
	code := m.Run()
	
	// Limpa ambiente de teste
	os.Exit(code)
}

// TestIntegrationSuite executa uma verificação completa de integração multi-hook
func TestIntegrationSuite(t *testing.T) {
	span, ctx := testutil.StartTestSpan(context.Background(), "TestIntegrationSuite")
	defer span.Finish()
	
	logger := testutil.GetTestLogger()
	logger.Info("Iniciando suite completa de testes de integração MCP-IAM",
		zap.String("test", "integration_suite"))

	// Configura serviço de elevação para testes
	elevationService, err := setupElevationService()
	require.NoError(t, err, "Falha ao configurar serviço de elevação")

	// Inicializa hooks MCP para testes
	dockerHook := initDockerHook(t, elevationService)
	desktopCommanderHook := initDesktopCommanderHook(t, elevationService)
	githubHook := initGitHubHook(t, elevationService)
	figmaHook := initFigmaHook(t, elevationService)
	
	// Teste integrado multi-hook, multi-tenant, multi-mercado
	for _, market := range testMarkets {
		t.Run(market+"_IntegracaoMultiHook", func(t *testing.T) {
			testCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			
			tenantID := "tenant-" + market + "-123"
			userID := "user-admin-456"
			
			// Configura contexto de autenticação
			authToken := createTestAuthToken(t, userID, tenantID, market)
			testCtx = auth.ContextWithToken(testCtx, authToken)
			
			// 1. Tenta operação Docker sensível sem elevação (deve falhar)
			dockerCmd := []string{"system", "prune", "--all", "--force"}
			_, dockerErr := dockerHook.ProcessCommand(testCtx, dockerCmd)
			require.Error(t, dockerErr, "Docker: Comando sensível deve ser negado sem elevação")
			
			// 2. Solicita elevação para Docker
			elevationRequest := &elevation.ElevationRequest{
				UserID:        userID,
				TenantID:      tenantID,
				Market:        market,
				Justification: "Teste de integração multi-hook",
				Scopes:        []string{"docker:system", "github:push:force", "figma:delete:file", "dc:fs:write"},
				Duration:      60, // 60 minutos
				Emergency:     false,
			}
			
			elevToken, err := elevationService.RequestElevation(testCtx, elevationRequest)
			require.NoError(t, err, "Falha ao solicitar elevação")
			
			// 3. Aprova elevação se necessário
			if elevToken.Status == elevation.StatusPendingApproval {
				supervisorID := "user-supervisor-789"
				supervisorCtx := createSupervisorContext(testCtx, supervisorID, tenantID, market)
				elevToken, err = elevationService.ApproveElevation(supervisorCtx, elevToken.ID)
				require.NoError(t, err, "Falha ao aprovar elevação")
				require.Equal(t, elevation.StatusActive, elevToken.Status, "Status deve ser ativo após aprovação")
			}
			
			// 4. Adiciona token de elevação ao contexto
			elevCtx := elevation.ContextWithElevationToken(testCtx, elevToken)
			elevCtx = auth.ContextWithToken(elevCtx, authToken)
			
			// 5. Tenta operação Docker com elevação (deve suceder)
			dockerResult, err := dockerHook.ProcessCommand(elevCtx, dockerCmd)
			require.NoError(t, err, "Docker: Comando deve ser permitido com elevação")
			require.True(t, dockerResult.Allowed, "Docker: Operação deve ser permitida com elevação")
			
			// 6. Tenta operação GitHub com a mesma elevação
			githubOp := hooks.GitHubOperation{
				Type:       "push_force",
				RepoOwner:  "innovabiz",
				RepoName:   "core-banking",
				Branch:     "main",
			}
			
			githubResult, err := githubHook.ProcessOperation(elevCtx, githubOp)
			require.NoError(t, err, "GitHub: Operação deve ser permitida com mesma elevação")
			require.True(t, githubResult.Allowed, "GitHub: Operação deve ser permitida com elevação")
			
			// 7. Tenta operação Figma com a mesma elevação
			figmaOp := hooks.FigmaOperation{
				Type:      "delete_file",
				FileKey:   "file_key_123456",
				ProjectID: "project_123",
				TeamID:    "team_innovabiz",
			}
			
			figmaResult, err := figmaHook.ProcessOperation(elevCtx, figmaOp)
			require.NoError(t, err, "Figma: Operação deve ser permitida com mesma elevação")
			require.True(t, figmaResult.Allowed, "Figma: Operação deve ser permitida com elevação")
			
			// 8. Tenta operação Desktop Commander com a mesma elevação
			dcCmd := "rm -rf /sensitive/data"
			dcResult, err := desktopCommanderHook.ProcessCommand(elevCtx, dcCmd)
			require.NoError(t, err, "Desktop Commander: Comando deve ser permitido com mesma elevação")
			require.True(t, dcResult.Allowed, "Desktop Commander: Operação deve ser permitida com elevação")
			
			// Verifica que a mesma elevação funciona para múltiplos hooks (integração completa)
			logger.Info("Teste de integração multi-hook concluído com sucesso",
				zap.String("market", market),
				zap.String("tenant", tenantID),
				zap.String("elevation_id", elevToken.ID))
			
			// Registra evento de auditoria para a integração completa
			testutil.LogAuditEvent(testCtx, "multi_hook_integration_test", map[string]interface{}{
				"tenant_id":    tenantID,
				"user_id":      userID,
				"market":       market,
				"elevation_id": elevToken.ID,
				"hooks_tested": []string{"docker", "github", "figma", "desktop_commander"},
				"success":      true,
			})
		})
	}
	
	// Teste de revogação de token de elevação
	t.Run("RevogacaoToken", func(t *testing.T) {
		testCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		
		market := "angola"
		tenantID := "tenant-angola-123"
		userID := "user-admin-456"
		
		// Configura contexto de autenticação
		authToken := createTestAuthToken(t, userID, tenantID, market)
		testCtx = auth.ContextWithToken(testCtx, authToken)
		
		// Solicita elevação
		elevationRequest := &elevation.ElevationRequest{
			UserID:        userID,
			TenantID:      tenantID,
			Market:        market,
			Justification: "Teste de revogação",
			Scopes:        []string{"docker:system"},
			Duration:      60,
			Emergency:     false,
		}
		
		elevToken, err := elevationService.RequestElevation(testCtx, elevationRequest)
		require.NoError(t, err, "Falha ao solicitar elevação")
		
		// Aprova elevação se necessário
		if elevToken.Status == elevation.StatusPendingApproval {
			supervisorID := "user-supervisor-789"
			supervisorCtx := createSupervisorContext(testCtx, supervisorID, tenantID, market)
			elevToken, err = elevationService.ApproveElevation(supervisorCtx, elevToken.ID)
			require.NoError(t, err, "Falha ao aprovar elevação")
		}
		
		// Contexto com elevação
		elevCtx := elevation.ContextWithElevationToken(testCtx, elevToken)
		elevCtx = auth.ContextWithToken(elevCtx, authToken)
		
		// Tenta operação com elevação (deve suceder)
		dockerCmd := []string{"system", "prune", "--all"}
		result, err := dockerHook.ProcessCommand(elevCtx, dockerCmd)
		require.NoError(t, err, "Operação deve ser permitida com elevação")
		require.True(t, result.Allowed, "Operação deve ser permitida com elevação")
		
		// Revoga token de elevação
		supervisorID := "user-supervisor-123"
		supervisorCtx := createSupervisorContext(testCtx, supervisorID, tenantID, market)
		err = elevationService.RevokeElevation(supervisorCtx, elevToken.ID, "Teste de revogação de acesso")
		require.NoError(t, err, "Falha ao revogar elevação")
		
		// Tenta operação após revogação (deve falhar)
		// Atualiza contexto com status mais recente
		revokedToken, _ := elevationService.GetElevationToken(testCtx, elevToken.ID)
		elevCtxRevoked := elevation.ContextWithElevationToken(testCtx, revokedToken)
		elevCtxRevoked = auth.ContextWithToken(elevCtxRevoked, authToken)
		
		_, err = dockerHook.ProcessCommand(elevCtxRevoked, dockerCmd)
		require.Error(t, err, "Operação deve ser negada após revogação")
		
		// Verifica status do token
		updatedToken, err := elevationService.GetElevationToken(testCtx, elevToken.ID)
		require.NoError(t, err, "Falha ao recuperar token")
		require.Equal(t, elevation.StatusRevoked, updatedToken.Status, "Status deve ser revogado")
		
		// Registra evento de auditoria
		testutil.LogAuditEvent(testCtx, "elevation_revocation_test", map[string]interface{}{
			"tenant_id":    tenantID,
			"user_id":      userID,
			"market":       market,
			"elevation_id": elevToken.ID,
			"revoked_by":   supervisorID,
			"success":      true,
		})
	})
	
	// Teste de políticas específicas por mercado
	t.Run("PoliticasEspecificasPorMercado", func(t *testing.T) {
		testMarketPolicies := map[string]struct {
			market        string
			tenantID      string
			operation     string
			requiresMFA   bool
			requiresApproval bool
		}{
			"Angola_RestricaoForte": {
				market:        "angola",
				tenantID:      "tenant-angola-123",
				operation:     "docker:system",
				requiresMFA:   true,
				requiresApproval: true,
			},
			"Brasil_RestricaoMedia": {
				market:        "brasil",
				tenantID:      "tenant-brasil-456",
				operation:     "github:push:force",
				requiresMFA:   true,
				requiresApproval: false,
			},
			"Mocambique_RestricaoMinima": {
				market:        "mocambique",
				tenantID:      "tenant-mocambique-789",
				operation:     "figma:delete:file",
				requiresMFA:   false,
				requiresApproval: false,
			},
		}
		
		for testName, testCase := range testMarketPolicies {
			t.Run(testName, func(t *testing.T) {
				testCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
				defer cancel()
				
				userID := "user-admin-456"
				
				// Configura contexto de autenticação com/sem MFA conforme teste
				authToken := createTestAuthToken(t, userID, testCase.tenantID, testCase.market)
				authToken.MFACompleted = !testCase.requiresMFA // Oposto para testar a validação
				testCtx = auth.ContextWithToken(testCtx, authToken)
				
				// Solicita elevação
				elevationRequest := &elevation.ElevationRequest{
					UserID:        userID,
					TenantID:      testCase.tenantID,
					Market:        testCase.market,
					Justification: "Teste de política específica por mercado",
					Scopes:        []string{testCase.operation},
					Duration:      30,
					Emergency:     false,
				}
				
				// Tenta solicitar elevação
				elevToken, err := elevationService.RequestElevation(testCtx, elevationRequest)
				
				if testCase.requiresMFA && !authToken.MFACompleted {
					// Deve falhar se precisar de MFA e não tiver
					require.Error(t, err, "Deve exigir MFA para solicitar elevação")
					require.Nil(t, elevToken, "Não deve retornar token quando MFA é necessário")
				} else {
					// Caso contrário deve permitir
					require.NoError(t, err, "Deve permitir solicitar elevação")
					require.NotNil(t, elevToken, "Deve retornar token de elevação")
					
					// Verifica se precisa de aprovação conforme política do mercado
					if testCase.requiresApproval {
						require.Equal(t, elevation.StatusPendingApproval, elevToken.Status, 
							"Status deve ser pendente aprovação")
						
						// Simula aprovação
						supervisorID := "user-supervisor-789"
						supervisorCtx := createSupervisorContext(testCtx, supervisorID, testCase.tenantID, testCase.market)
						approvedToken, err := elevationService.ApproveElevation(supervisorCtx, elevToken.ID)
						require.NoError(t, err, "Falha ao aprovar elevação")
						require.Equal(t, elevation.StatusActive, approvedToken.Status, "Status deve ser ativo após aprovação")
					} else {
						// Se não requer aprovação, deve já estar ativo
						require.Equal(t, elevation.StatusActive, elevToken.Status, 
							"Status deve ser ativo imediatamente (sem aprovação)")
					}
				}
				
				// Registra política específica do mercado
				testutil.LogAuditEvent(testCtx, "market_specific_policy_test", map[string]interface{}{
					"tenant_id":     testCase.tenantID,
					"user_id":       userID,
					"market":        testCase.market,
					"operation":     testCase.operation,
					"requires_mfa":  testCase.requiresMFA,
					"needs_approval": testCase.requiresApproval,
					"test_result":   err == nil,
				})
			})
		}
	})
	
	// Teste de isolation completa entre tenants
	t.Run("IsolationEntreTenants", func(t *testing.T) {
		testCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		
		market := "angola"
		tenant1ID := "tenant-angola-123"
		tenant2ID := "tenant-angola-456"
		user1ID := "user-admin-111"
		user2ID := "user-admin-222"
		
		// Cria token de elevação para tenant 1
		authToken1 := createTestAuthToken(t, user1ID, tenant1ID, market)
		ctx1 := auth.ContextWithToken(testCtx, authToken1)
		
		elevRequest1 := &elevation.ElevationRequest{
			UserID:        user1ID,
			TenantID:      tenant1ID,
			Market:        market,
			Justification: "Teste de isolation entre tenants",
			Scopes:        []string{"docker:system"},
			Duration:      30,
			Emergency:     false,
		}
		
		elevToken1, err := elevationService.RequestElevation(ctx1, elevRequest1)
		require.NoError(t, err, "Falha ao solicitar elevação para tenant 1")
		
		// Aprova token para tenant 1
		supervisorID := "user-supervisor-789"
		supervisorCtx := createSupervisorContext(ctx1, supervisorID, tenant1ID, market)
		elevToken1, err = elevationService.ApproveElevation(supervisorCtx, elevToken1.ID)
		require.NoError(t, err, "Falha ao aprovar elevação para tenant 1")
		
		// Tenta usar token do tenant 1 no tenant 2
		authToken2 := createTestAuthToken(t, user2ID, tenant2ID, market)
		ctx2 := auth.ContextWithToken(testCtx, authToken2)
		ctx2WithWrongElev := elevation.ContextWithElevationToken(ctx2, elevToken1)
		
		dockerCmd := []string{"system", "prune", "--all"}
		_, err = dockerHook.ProcessCommand(ctx2WithWrongElev, dockerCmd)
		require.Error(t, err, "Deve negar uso de token de outro tenant")
		require.Contains(t, err.Error(), "tenant mismatch", 
			"Erro deve mencionar incompatibilidade de tenant")
		
		// Verifica que tenant pode usar apenas seus próprios tokens
		elevRequest2 := &elevation.ElevationRequest{
			UserID:        user2ID,
			TenantID:      tenant2ID,
			Market:        market,
			Justification: "Teste de isolation para tenant 2",
			Scopes:        []string{"docker:system"},
			Duration:      30,
			Emergency:     true, // Usa emergência para evitar aprovação
		}
		
		elevToken2, err := elevationService.RequestElevation(ctx2, elevRequest2)
		require.NoError(t, err, "Falha ao solicitar elevação para tenant 2")
		
		ctx2WithRightElev := elevation.ContextWithElevationToken(ctx2, elevToken2)
		result, err := dockerHook.ProcessCommand(ctx2WithRightElev, dockerCmd)
		require.NoError(t, err, "Deve permitir uso do próprio token de tenant")
		require.True(t, result.Allowed, "Operação deve ser permitida com token do próprio tenant")
		
		// Registra teste de isolation
		testutil.LogAuditEvent(testCtx, "tenant_isolation_test", map[string]interface{}{
			"tenant1_id":   tenant1ID,
			"tenant2_id":   tenant2ID,
			"user1_id":     user1ID,
			"user2_id":     user2ID,
			"market":       market,
			"elevation1_id": elevToken1.ID,
			"elevation2_id": elevToken2.ID,
			"isolation_preserved": true,
		})
	})
}

// initDockerHook inicializa o hook Docker para testes
func initDockerHook(t *testing.T, elevationService *elevation.Service) *hooks.DockerMCPHook {
	logger := testutil.GetTestLogger()
	dockerHook := hooks.NewDockerMCPHook(elevationService, logger)
	require.NotNil(t, dockerHook, "Hook Docker não deve ser nil")
	return dockerHook
}

// initDesktopCommanderHook inicializa o hook Desktop Commander para testes
func initDesktopCommanderHook(t *testing.T, elevationService *elevation.Service) *hooks.DesktopCommanderMCPHook {
	logger := testutil.GetTestLogger()
	dcHook := hooks.NewDesktopCommanderMCPHook(elevationService, logger)
	require.NotNil(t, dcHook, "Hook Desktop Commander não deve ser nil")
	return dcHook
}

// initGitHubHook inicializa o hook GitHub para testes
func initGitHubHook(t *testing.T, elevationService *elevation.Service) *hooks.GitHubMCPHook {
	logger := testutil.GetTestLogger()
	githubHook := hooks.NewGitHubMCPHook(elevationService, logger)
	require.NotNil(t, githubHook, "Hook GitHub não deve ser nil")
	return githubHook
}

// initFigmaHook inicializa o hook Figma para testes
func initFigmaHook(t *testing.T, elevationService *elevation.Service) *hooks.FigmaMCPHook {
	logger := testutil.GetTestLogger()
	figmaHook := hooks.NewFigmaMCPHook(elevationService, logger)
	require.NotNil(t, figmaHook, "Hook Figma não deve ser nil")
	return figmaHook
}