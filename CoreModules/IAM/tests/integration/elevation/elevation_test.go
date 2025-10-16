// Package elevation contém testes de integração para o sistema de elevação de privilégios MCP-IAM.
// Este arquivo testa o fluxo completo de elevação de privilégios com os diferentes hooks MCP.
package elevation

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/innovabiz/iam/auth"
	"github.com/innovabiz/iam/core/tenant"
	"github.com/innovabiz/iam/elevation"
	"github.com/innovabiz/iam/mcp/hooks"
	"github.com/innovabiz/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestElevationSetup configura o ambiente para testes de elevação
func TestElevationSetup(t *testing.T) {
	// Inicializa o logger de testes
	logger, err := testutil.NewTestLogger("elevation_test")
	require.NoError(t, err, "Falha ao inicializar logger de teste")
	defer logger.Sync()

	// Inicializa o tracer de observabilidade
	tracer, closer, err := testutil.NewTestTracer("elevation_test")
	require.NoError(t, err, "Falha ao inicializar tracer de teste")
	defer closer.Close()

	// Inicializa o coletor de métricas
	metrics := testutil.NewTestMetrics("elevation_test")

	// Configura observabilidade para os testes
	obs := testutil.NewObservability(logger, tracer, metrics)
	
	// Configura banco de dados PostgreSQL para testes via testcontainers
	dbContainer, dbConn, err := testutil.SetupTestPostgreSQL()
	require.NoError(t, err, "Falha ao inicializar container PostgreSQL")
	defer dbContainer.Terminate(context.Background())

	// Configura Redis para cache via testcontainers
	redisContainer, redisClient, err := testutil.SetupTestRedis()
	require.NoError(t, err, "Falha ao inicializar container Redis")
	defer redisContainer.Terminate(context.Background())

	// Aplica migrações de banco de dados
	err = testutil.ApplyMigrations(dbConn, "../../../migrations")
	require.NoError(t, err, "Falha ao aplicar migrações")

	// Seed de dados iniciais para testes
	err = seedTestData(dbConn)
	require.NoError(t, err, "Falha ao inserir dados de teste")

	// Testa a conexão com o banco de dados
	err = dbConn.Ping()
	require.NoError(t, err, "Falha na conexão com banco de dados")

	// Testa a conexão com o Redis
	_, err = redisClient.Ping(context.Background()).Result()
	require.NoError(t, err, "Falha na conexão com Redis")

	// Configura variáveis de ambiente para testes
	os.Setenv("IAM_TEST_MODE", "true")
	os.Setenv("IAM_DEFAULT_MARKET", "angola")
	os.Setenv("IAM_ENABLE_AUDIT", "true")
	
	// Log de setup concluído
	logger.Info("Ambiente de teste configurado com sucesso", 
		zap.String("test", "elevation"),
		zap.Bool("postgres_connected", true),
		zap.Bool("redis_connected", true),
	)
}

// TestDockerElevation testa o fluxo completo de elevação de privilégios para o hook Docker
func TestDockerElevation(t *testing.T) {
	span, ctx := testutil.StartTestSpan(context.Background(), "TestDockerElevation")
	defer span.Finish()
	
	logger := testutil.GetTestLogger()

	// Configura serviço de elevação para testes
	elevationService, err := setupElevationService()
	require.NoError(t, err, "Falha ao configurar serviço de elevação")

	// Configura o hook Docker para testes
	dockerHook := hooks.NewDockerMCPHook(elevationService, logger)
	
	// Configura os dados de teste
	tenantID := "tenant-angola-123"
	userID := "user-admin-456"
	market := "angola"
	
	// Cria um token de autenticação simulado para o usuário de teste
	authToken := createTestAuthToken(t, userID, tenantID, market)
	
	// Define o comando Docker sensível para teste
	dockerCommand := &hooks.DockerCommandRequest{
		Command: "container",
		Subcommand: "exec",
		Args: []string{"-it", "production-database", "/bin/bash"},
		UserID: userID,
		TenantID: tenantID,
	}
	
	// Cenário 1: Tentativa sem elevação (deve ser negado)
	t.Run("DeveSolicitarElevacaoParaComandoSensivel", func(t *testing.T) {
		testCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		
		testCtx = auth.ContextWithToken(testCtx, authToken)
		
		// Tenta executar comando sem elevação
		response, err := dockerHook.ProcessCommand(testCtx, dockerCommand)
		
		// Deve retornar erro de elevação necessária
		require.Error(t, err, "Deveria exigir elevação para comando sensível")
		assert.Contains(t, err.Error(), "elevation required", "Erro deveria indicar necessidade de elevação")
		assert.Nil(t, response, "Resposta deve ser nil quando elevação é necessária")
		
		// Verifica se evento de auditoria foi gerado
		// TODO: Validar registros de auditoria
	})
	
	// Cenário 2: Solicita elevação e aprovação
	t.Run("DeveSolicitarAprovacaoElevacao", func(t *testing.T) {
		testCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		
		testCtx = auth.ContextWithToken(testCtx, authToken)
		
		// Solicita elevação de privilégios
		elevationRequest := &elevation.ElevationRequest{
			UserID: userID,
			TenantID: tenantID,
			Justification: "Manutenção emergencial no banco de dados de produção",
			Scopes: []string{"docker:container:exec"},
			Duration: 30, // 30 minutos
			Emergency: false,
		}
		
		// Submete solicitação de elevação
		elevationToken, err := elevationService.RequestElevation(testCtx, elevationRequest)
		require.NoError(t, err, "Falha ao solicitar elevação")
		assert.NotNil(t, elevationToken, "Token de elevação não deveria ser nil")
		assert.Equal(t, elevation.StatusPendingApproval, elevationToken.Status, "Status deve ser pendente aprovação")
		
		// Guarda ID para aprovação
		elevationID := elevationToken.ID
		
		// Simula aprovação por outro usuário (supervisor)
		supervisorID := "user-supervisor-789"
		supervisorCtx := createSupervisorContext(testCtx, supervisorID, tenantID, market)
		
		approvedToken, err := elevationService.ApproveElevation(supervisorCtx, elevationID)
		require.NoError(t, err, "Falha ao aprovar elevação")
		assert.Equal(t, elevation.StatusActive, approvedToken.Status, "Status deve ser ativo após aprovação")
		
		// Verifica se eventos de auditoria foram gerados
		// TODO: Validar registros de auditoria
	})
	
	// Cenário 3: Usa token de elevação para comando sensível
	t.Run("DevePermitirComandoComElevacao", func(t *testing.T) {
		testCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		
		// Obtém um token de elevação válido para o teste
		elevationToken, err := getOrCreateTestElevationToken(testCtx, elevationService, userID, tenantID, market)
		require.NoError(t, err, "Falha ao obter token de elevação para teste")
		
		// Adiciona token de elevação ao contexto
		elevCtx := elevation.ContextWithElevationToken(testCtx, elevationToken)
		elevCtx = auth.ContextWithToken(elevCtx, authToken)
		
		// Tenta novamente o comando com contexto de elevação
		response, err := dockerHook.ProcessCommand(elevCtx, dockerCommand)
		
		// Deve ser permitido com elevação
		require.NoError(t, err, "Comando deveria ser permitido com elevação")
		assert.NotNil(t, response, "Resposta não deveria ser nil quando comando é permitido")
		assert.Equal(t, "container exec operation allowed", response.Message, "Mensagem de resposta incorreta")
		assert.True(t, response.Allowed, "Comando deveria ser permitido")
		
		// Verifica se evento de auditoria foi gerado
		// TODO: Validar registros de auditoria
	})
	
	// Cenário 4: Tenta usar token de elevação expirado
	t.Run("DeveRejeitarComandoComElevacaoExpirada", func(t *testing.T) {
		testCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		
		// Obtém um token de elevação expirado para o teste
		expiredToken := getTestExpiredElevationToken(t, userID, tenantID, market)
		
		// Adiciona token de elevação expirado ao contexto
		elevCtx := elevation.ContextWithElevationToken(testCtx, expiredToken)
		elevCtx = auth.ContextWithToken(elevCtx, authToken)
		
		// Tenta o comando com contexto de elevação expirada
		response, err := dockerHook.ProcessCommand(elevCtx, dockerCommand)
		
		// Deve ser rejeitado por ter elevação expirada
		require.Error(t, err, "Comando deveria ser rejeitado com elevação expirada")
		assert.Contains(t, err.Error(), "elevation token expired", "Erro deveria indicar expiração do token")
		assert.Nil(t, response, "Resposta deve ser nil quando elevação é expirada")
		
		// Verifica se evento de auditoria foi gerado
		// TODO: Validar registros de auditoria
	})
}