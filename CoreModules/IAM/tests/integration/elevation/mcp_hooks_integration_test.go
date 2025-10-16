// Package integration_test contém testes de integração para validar o funcionamento
// completo do sistema de elevação de privilégios com hooks MCP
package integration_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/innovabizdevops/innovabiz-iam/authorization/elevation"
	"github.com/innovabizdevops/innovabiz-iam/hooks/mcp"
	"github.com/innovabizdevops/innovabiz-iam/tests/testutil"
)

// TestMCPHooksIntegration realiza testes de integração completos para validar
// o fluxo de elevação de privilégios integrado com hooks MCP
func TestMCPHooksIntegration(t *testing.T) {
	// Verificar se os testes de integração estão habilitados
	if testing.Short() {
		t.Skip("Pulando testes de integração em modo curto")
	}

	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("mcp_hooks_integration_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())

	// Iniciar contexto de teste com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Registrar início do teste
	testCtx := obs.RecordTestStart(ctx, "MCPHooksIntegration")

	// Configurar e iniciar containers para o teste
	dbContainer, err := setupDatabaseContainer(testCtx)
	require.NoError(t, err, "Falha ao configurar container do banco de dados")
	defer dbContainer.Terminate(testCtx)

	redisContainer, err := setupRedisContainer(testCtx)
	require.NoError(t, err, "Falha ao configurar container do Redis")
	defer redisContainer.Terminate(testCtx)

	// Inicializar conexões e componentes
	dbConn, err := connectToDatabase(testCtx, dbContainer)
	require.NoError(t, err, "Falha ao conectar ao banco de dados")
	defer dbConn.Close()

	redisClient, err := connectToRedis(testCtx, redisContainer)
	require.NoError(t, err, "Falha ao conectar ao Redis")
	defer redisClient.Close()

	// Configurar serviço de elevação de privilégios real
	elevationService, err := setupElevationService(testCtx, dbConn, redisClient, obs)
	require.NoError(t, err, "Falha ao configurar serviço de elevação")

	// Configurar hooks MCP
	dockerHook := mcp.NewDockerHook(elevationService)
	desktopCommanderHook := mcp.NewDesktopCommanderHook(elevationService)
	githubHook := mcp.NewGitHubHook(elevationService)
	figmaHook := mcp.NewFigmaHook(elevationService)

	// Configurar observabilidade para os hooks
	dockerHook.ConfigureObservability(obs.Logger(), obs.Tracer())
	desktopCommanderHook.ConfigureObservability(obs.Logger(), obs.Tracer())
	githubHook.ConfigureObservability(obs.Logger(), obs.Tracer())
	figmaHook.ConfigureObservability(obs.Logger(), obs.Tracer())

	// Configurar mapeamentos de escopos para os hooks
	configureScopeMappings(dockerHook, desktopCommanderHook, githubHook, figmaHook)

	// Configurar proteções específicas para os hooks
	configureProtections(dockerHook, desktopCommanderHook, githubHook, figmaHook)

	// Executar subtestes
	t.Run("FluxoCompletoDockerElevation", testDockerElevationFlow(testCtx, dockerHook, elevationService))
	t.Run("IsolamentoMultiTenantDesktopCommander", testDesktopCommanderIsolation(testCtx, desktopCommanderHook, elevationService))
	t.Run("IntegracaoMFAGitHub", testGitHubMFAIntegration(testCtx, githubHook, elevationService))
	t.Run("AuditoriaBranchesProtegidos", testProtectedBranchAuditing(testCtx, githubHook, elevationService))
	t.Run("RastreabilidadeFigma", testFigmaTraceability(testCtx, figmaHook, elevationService))

	// Registrar conclusão do teste
	obs.RecordTestEnd(testCtx, "MCPHooksIntegration", err == nil, time.Since(time.Now()))
}

// Funções de teste para os diferentes fluxos
func testDockerElevationFlow(ctx context.Context, dockerHook *mcp.DockerHook, elevationService *elevation.PrivilegeElevationService) func(t *testing.T) {
	return func(t *testing.T) {
		// Definir IDs para o teste
		userID := "user:admin:e2e-test-123"
		tenantID := "tenant_angola_1"
		market := "angola"

		// Criar contexto com tenant
		hookCtx := mcp.WithTenantID(ctx, tenantID)

		// 1. Solicitar elevação através do hook Docker
		elevationRequest := &mcp.ElevationRequestPayload{
			UserID:          userID,
			RequestedScopes: []string{"docker:container:run", "docker:container:stop"},
			Justification:   "Manutenção emergencial de containers",
			Duration:        "15m",
			EmergencyAccess: true,
			Context: map[string]interface{}{
				"tenant_id":     tenantID,
				"market":        market,
				"business_unit": "operations",
				"client_ip":     "192.168.1.100",
				"user_agent":    "MCP Docker Client/1.0",
			},
		}

		// Serializar requisição
		requestJSON, err := json.Marshal(elevationRequest)
		require.NoError(t, err, "Erro ao serializar requisição de elevação")

		// Solicitar elevação
		resultJSON, err := dockerHook.RequestDockerElevation(hookCtx, requestJSON)
		require.NoError(t, err, "Solicitação de elevação não deveria falhar")

		// Deserializar resposta
		var elevationResponse mcp.ElevationResponse
		err = json.Unmarshal(resultJSON, &elevationResponse)
		require.NoError(t, err, "Erro ao deserializar resposta de elevação")

		// Verificar resposta
		assert.NotEmpty(t, elevationResponse.ElevationID, "ID de elevação não deveria estar vazio")
		assert.NotEmpty(t, elevationResponse.ElevationToken, "Token de elevação não deveria estar vazio")
		assert.Contains(t, elevationResponse.ElevatedScopes, "docker:container:run", "Escopo docker:container:run deveria estar elevado")
		assert.Contains(t, elevationResponse.ElevatedScopes, "docker:container:stop", "Escopo docker:container:stop deveria estar elevado")

		// 2. Usar a elevação para autorizar um comando Docker
		// Preparar payload para comando Docker
		dockerCommandPayload := map[string]interface{}{
			"args": []string{"run", "--rm", "nginx:latest"},
		}

		commandJSON, err := json.Marshal(dockerCommandPayload)
		require.NoError(t, err, "Erro ao serializar comando Docker")

		// Adicionar token de elevação ao contexto
		elevatedCtx := mcp.WithElevationToken(hookCtx, elevationResponse.ElevationToken)

		// Autorizar comando
		allowed, reason, err := dockerHook.AuthorizeDockerCommand(elevatedCtx, "docker", commandJSON)
		require.NoError(t, err, "A autorização não deveria falhar")

		// Verificar resultados
		assert.True(t, allowed, "O comando deveria ser autorizado")
		assert.Contains(t, reason, "Elevação verificada com sucesso", "Razão de autorização incorreta")

		// 3. Verificar registro de uso
		// Consultar registros de auditoria para confirmar o registro de uso
		auditEvents, err := elevationService.QueryAuditEvents(ctx, map[string]interface{}{
			"elevation_id": elevationResponse.ElevationID,
			"event_type":   "elevation_usage",
		})

		require.NoError(t, err, "Consulta de auditoria não deveria falhar")
		assert.GreaterOrEqual(t, len(auditEvents), 1, "Deveria haver pelo menos um evento de uso registrado")

		// 4. Revogar elevação
		err = elevationService.RevokeElevation(ctx, elevationResponse.ElevationID, "Teste de revogação")
		require.NoError(t, err, "Revogação não deveria falhar")

		// 5. Tentar usar elevação revogada
		allowed, reason, err = dockerHook.AuthorizeDockerCommand(elevatedCtx, "docker", commandJSON)
		require.NoError(t, err, "A autorização não deveria falhar tecnicamente")
		
		// Verificar que a elevação foi revogada
		assert.False(t, allowed, "O comando não deveria ser autorizado após revogação")
		assert.Contains(t, reason, "Token inválido ou revogado", "Razão de negação incorreta")
	}
}

// setupDatabaseContainer inicializa um container PostgreSQL para os testes
func setupDatabaseContainer(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:14-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "innovabiz_iam_test",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	return container, err
}

// setupRedisContainer inicializa um container Redis para os testes
func setupRedisContainer(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	return container, err
}