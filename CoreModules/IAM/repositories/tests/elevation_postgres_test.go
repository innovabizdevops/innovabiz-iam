// Package tests implementa testes para os repositórios do sistema IAM
// da plataforma INNOVABIZ.
package tests

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/innovabiz/iam/services/elevation"
	"github.com/innovabiz/iam/repositories"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

// ElevationRepositoryTestSuite define a suite de testes para o repositório
// de tokens de elevação PostgreSQL.
type ElevationRepositoryTestSuite struct {
	suite.Suite
	db         *sqlx.DB
	repository *repositories.PostgresElevationRepository
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	ctx        context.Context
	markets    []string
	tenants    []string
}

// SetupSuite configura o ambiente para a suite de testes
func (s *ElevationRepositoryTestSuite) SetupSuite() {
	var err error
	s.ctx = context.Background()

	// Configuração dos mercados e tenants para testes multi-mercado e multi-tenant
	s.markets = []string{"angola", "brasil", "mozambique"}
	s.tenants = []string{"tenant1", "tenant2", "tenant3"}

	// Cria pool de Docker
	s.pool, err = dockertest.NewPool("")
	require.NoError(s.T(), err, "Failed to connect to Docker")

	// Configura recursos do Docker
	s.resource, err = s.pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14",
		Env: []string{
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=testdb",
			"listen_addresses='*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	require.NoError(s.T(), err, "Failed to start PostgreSQL container")

	// Configura mapeamento de portas
	hostPort := s.resource.GetPort("5432/tcp")
	dsn := fmt.Sprintf("postgres://postgres:postgres@localhost:%s/testdb?sslmode=disable", hostPort)

	// Tenta conectar ao banco de dados
	s.T().Logf("Connecting to PostgreSQL on localhost:%s", hostPort)

	// Retry para aguardar o banco ficar disponível
	err = s.pool.Retry(func() error {
		var err error
		s.db, err = sqlx.Open("postgres", dsn)
		if err != nil {
			return err
		}
		return s.db.Ping()
	})
	require.NoError(s.T(), err, "Failed to connect to PostgreSQL container")

	// Lê e executa scripts SQL para criar tabelas
	migrationPath := "../../migrations/000001_create_elevation_tokens_table.up.sql"
	migrationSQL, err := os.ReadFile(migrationPath)
	require.NoError(s.T(), err, "Failed to read migration SQL file")

	_, err = s.db.Exec(string(migrationSQL))
	require.NoError(s.T(), err, "Failed to execute migration SQL")

	// Cria repositório
	s.repository = repositories.NewPostgresElevationRepository(s.db)
	require.NotNil(s.T(), s.repository, "Failed to create repository instance")
}

// TearDownSuite finaliza o ambiente após os testes
func (s *ElevationRepositoryTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}

	if s.resource != nil {
		s.pool.Purge(s.resource)
	}
}

// createTestToken cria um token de teste
func (s *ElevationRepositoryTestSuite) createTestToken(
	userID string,
	tenantID string,
	market string,
	status elevation.Status,
	emergency bool,
) *elevation.Token {
	tokenID := uuid.New().String()
	now := time.Now()

	token := &elevation.Token{
		ID:            tokenID,
		UserID:        userID,
		TenantID:      tenantID,
		Market:        market,
		Scopes:        []string{"read:files", "write:files"},
		Status:        status,
		Justification: "Necessário para teste",
		CreatedAt:     now,
		ExpiresAt:     now.Add(1 * time.Hour),
		Emergency:     emergency,
	}

	// Configura campos específicos baseados no status
	switch status {
	case elevation.StatusActive:
		token.ApprovedBy = "approver1"
		token.ApprovedAt = now.Add(-10 * time.Minute)
	case elevation.StatusDenied:
		token.DeniedBy = "denier1"
		token.DeniedAt = now.Add(-10 * time.Minute)
		token.DenyReason = "Solicitação negada por motivos de segurança"
	case elevation.StatusRevoked:
		token.ApprovedBy = "approver1"
		token.ApprovedAt = now.Add(-20 * time.Minute)
		token.RevokedBy = "revoker1"
		token.RevokedAt = now.Add(-5 * time.Minute)
		token.RevokeReason = "Solicitação revogada por não ser mais necessária"
	case elevation.StatusExpired:
		token.ApprovedBy = "approver1"
		token.ApprovedAt = now.Add(-2 * time.Hour)
		token.ExpiresAt = now.Add(-10 * time.Minute)
	}

	return token
}

// populateDatabase insere dados de teste no banco
func (s *ElevationRepositoryTestSuite) populateDatabase() {
	// Cria tokens para cada combinação de mercado e tenant
	statuses := []elevation.Status{
		elevation.StatusPendingApproval,
		elevation.StatusActive,
		elevation.StatusDenied,
		elevation.StatusRevoked,
		elevation.StatusExpired,
	}

	userIDs := []string{"user1", "user2", "user3"}

	for _, market := range s.markets {
		for _, tenantID := range s.tenants {
			for _, userID := range userIDs {
				for _, status := range statuses {
					// Tokens regulares
					token := s.createTestToken(userID, tenantID, market, status, false)
					err := s.repository.SaveToken(s.ctx, token)
					require.NoError(s.T(), err, "Failed to save test token")

					// Alguns tokens de emergência (apenas para status active e expired)
					if status == elevation.StatusActive || status == elevation.StatusExpired {
						emergencyToken := s.createTestToken(userID, tenantID, market, status, true)
						err := s.repository.SaveToken(s.ctx, emergencyToken)
						require.NoError(s.T(), err, "Failed to save emergency test token")
					}
				}
			}
		}
	}
}

// TestCountActiveTokensByUser testa a contagem de tokens ativos por usuário
func (s *ElevationRepositoryTestSuite) TestCountActiveTokensByUser() {
	userID := "count_user"
	tenantID := "count_tenant"
	market := "brasil"
	
	// Cria tokens ativos
	for i := 0; i < 3; i++ {
		token := s.createTestToken(userID, tenantID, market, elevation.StatusActive, false)
		err := s.repository.SaveToken(s.ctx, token)
		require.NoError(s.T(), err)
	}
	
	// Cria um token com status diferente
	inactiveToken := s.createTestToken(userID, tenantID, market, elevation.StatusRevoked, false)
	err := s.repository.SaveToken(s.ctx, inactiveToken)
	require.NoError(s.T(), err)
	
	// Testa contagem
	count, err := s.repository.CountActiveTokensByUser(s.ctx, userID, tenantID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 3, count, "Should count 3 active tokens")
	
	// Testa contagem para usuário sem tokens
	count, err = s.repository.CountActiveTokensByUser(s.ctx, "nonexistent", tenantID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 0, count, "Should count 0 tokens for nonexistent user")
}

// TestListPendingTokens testa listagem de tokens pendentes
func (s *ElevationRepositoryTestSuite) TestListPendingTokens() {
	// Limpa dados anteriores e popula com tokens de teste
	s.populateDatabase()
	
	// Para cada tenant e mercado, deve haver 3 tokens pendentes (um para cada usuário)
	for _, market := range s.markets {
		for _, tenant := range s.tenants {
			tokens, err := s.repository.ListPendingTokens(s.ctx, tenant, market)
			require.NoError(s.T(), err)
			assert.Equal(s.T(), 3, len(tokens), 
				fmt.Sprintf("Should find 3 pending tokens for tenant %s, market %s", tenant, market))
			
			for _, token := range tokens {
				assert.Equal(s.T(), elevation.StatusPendingApproval, token.Status)
				assert.Equal(s.T(), tenant, token.TenantID)
				assert.Equal(s.T(), market, token.Market)
			}
		}
	}
}

// TestListTokensByStatus testa listagem de tokens por status
func (s *ElevationRepositoryTestSuite) TestListTokensByStatus() {
	// Limpa dados anteriores e popula com tokens de teste
	s.populateDatabase()
	
	testCases := []struct {
		status          elevation.Status
		expectedPerTenant int
	}{
		{elevation.StatusPendingApproval, 3}, // 3 usuários
		{elevation.StatusActive, 6},          // 3 usuários + 3 emergências
		{elevation.StatusDenied, 3},          // 3 usuários
		{elevation.StatusRevoked, 3},         // 3 usuários
		{elevation.StatusExpired, 6},         // 3 usuários + 3 emergências
	}
	
	for _, tc := range testCases {
		for _, market := range s.markets {
			for _, tenant := range s.tenants {
				tokens, err := s.repository.ListTokensByStatus(s.ctx, tenant, market, tc.status)
				require.NoError(s.T(), err)
				assert.Equal(s.T(), tc.expectedPerTenant, len(tokens), 
					fmt.Sprintf("Should find %d tokens with status %s for tenant %s, market %s", 
						tc.expectedPerTenant, tc.status, tenant, market))
				
				for _, token := range tokens {
					assert.Equal(s.T(), tc.status, token.Status)
					assert.Equal(s.T(), tenant, token.TenantID)
					assert.Equal(s.T(), market, token.Market)
				}
			}
		}
	}
}

// TestListExpiredTokens testa listagem de tokens expirados
func (s *ElevationRepositoryTestSuite) TestListExpiredTokens() {
	// Cria tokens especificamente para este teste
	userID := "expired_user"
	tenantID := "expired_tenant"
	market := "angola"
	
	// Tokens expirados mas ainda marcados como ativos
	for i := 0; i < 3; i++ {
		token := s.createTestToken(userID, tenantID, market, elevation.StatusActive, false)
		token.ExpiresAt = time.Now().Add(-1 * time.Hour) // Expirou há uma hora
		err := s.repository.SaveToken(s.ctx, token)
		require.NoError(s.T(), err)
	}
	
	// Token ativo e não expirado para controle
	activeToken := s.createTestToken(userID, tenantID, market, elevation.StatusActive, false)
	activeToken.ExpiresAt = time.Now().Add(1 * time.Hour) // Expira em uma hora
	err := s.repository.SaveToken(s.ctx, activeToken)
	require.NoError(s.T(), err)
	
	// Tokens já marcados como expirados (não devem aparecer no resultado)
	expiredToken := s.createTestToken(userID, tenantID, market, elevation.StatusExpired, false)
	expiredToken.ExpiresAt = time.Now().Add(-2 * time.Hour) // Expirou há duas horas
	err = s.repository.SaveToken(s.ctx, expiredToken)
	require.NoError(s.T(), err)
	
	// Busca tokens expirados
	tokens, err := s.repository.ListExpiredTokens(s.ctx)
	require.NoError(s.T(), err)
	
	// Deve encontrar apenas os 3 tokens que estão marcados como ativos mas expirados
	// Contagem pode ser maior se houver outros testes criando tokens expirados
	foundOurTokens := 0
	for _, token := range tokens {
		if token.UserID == userID && token.TenantID == tenantID && token.Market == market {
			foundOurTokens++
			assert.Equal(s.T(), elevation.StatusActive, token.Status)
			assert.True(s.T(), token.ExpiresAt.Before(time.Now()))
		}
	}
	
	assert.Equal(s.T(), 3, foundOurTokens, "Should find all 3 expired tokens")
}

// TestGetTokenHistory testa recuperação do histórico de tokens
func (s *ElevationRepositoryTestSuite) TestGetTokenHistory() {
	// Cria um token e executa várias operações para gerar histórico
	userID := "history_user"
	tenantID := "history_tenant"
	market := "brasil"
	
	// Cria token pendente
	token := s.createTestToken(userID, tenantID, market, elevation.StatusPendingApproval, false)
	err := s.repository.SaveToken(s.ctx, token)
	require.NoError(s.T(), err)
	
	// Aprova o token
	token.Status = elevation.StatusActive
	token.ApprovedBy = "approver_user"
	token.ApprovedAt = time.Now()
	err = s.repository.UpdateToken(s.ctx, token)
	require.NoError(s.T(), err)
	
	// Revoga o token
	token.Status = elevation.StatusRevoked
	token.RevokedBy = "revoker_user"
	token.RevokedAt = time.Now()
	token.RevokeReason = "Não é mais necessário"
	err = s.repository.UpdateToken(s.ctx, token)
	require.NoError(s.T(), err)
	
	// Recupera histórico
	history, err := s.repository.GetTokenHistory(s.ctx, token.ID)
	require.NoError(s.T(), err)
	
	// Deve ter 3 entradas: criação, aprovação e revogação
	assert.Equal(s.T(), 3, len(history), "Should have 3 history entries")
	
	// Verifica sequência de eventos
	assert.Equal(s.T(), "created", history[0].Action)
	assert.Equal(s.T(), elevation.Status(""), history[0].PrevStatus) // Status vazio para criação
	assert.Equal(s.T(), elevation.StatusPendingApproval, history[0].NewStatus)
	
	assert.Equal(s.T(), "approved", history[1].Action)
	assert.Equal(s.T(), elevation.StatusPendingApproval, history[1].PrevStatus)
	assert.Equal(s.T(), elevation.StatusActive, history[1].NewStatus)
	
	assert.Equal(s.T(), "revoked", history[2].Action)
	assert.Equal(s.T(), elevation.StatusActive, history[2].PrevStatus)
	assert.Equal(s.T(), elevation.StatusRevoked, history[2].NewStatus)
	assert.Equal(s.T(), "Não é mais necessário", history[2].Reason)
}

// TestGetTokenStats testa recuperação de estatísticas de tokens
func (s *ElevationRepositoryTestSuite) TestGetTokenStats() {
	// Limpa dados anteriores e popula com tokens de teste
	s.populateDatabase()
	
	// Teste para cada mercado e tenant
	for _, market := range s.markets {
		for _, tenant := range s.tenants {
			stats, err := s.repository.GetTokenStats(s.ctx, tenant, market, time.Now().Add(-24*time.Hour))
			require.NoError(s.T(), err)
			
			// Verificações baseadas nos dados inseridos por populateDatabase
			assert.Equal(s.T(), 15, stats.TotalRequested, "Total tokens should be 15 (3 users * 5 statuses)")
			assert.Equal(s.T(), 12, stats.TotalApproved, "Approved tokens should be 12 (3 users * 4 statuses)")
			assert.Equal(s.T(), 3, stats.TotalDenied, "Denied tokens should be 3 (3 users)")
			assert.Equal(s.T(), 3, stats.TotalRevoked, "Revoked tokens should be 3 (3 users)")
			assert.Equal(s.T(), 6, stats.TotalExpired, "Expired tokens should be 6 (3 users + 3 emergencies)")
			assert.Equal(s.T(), 6, stats.TotalEmergency, "Emergency tokens should be 6 (3 active + 3 expired)")
			assert.Equal(s.T(), 6, stats.CurrentActiveTokens, "Active tokens should be 6 (3 users + 3 emergencies)")
			assert.Equal(s.T(), 3, stats.CurrentPendingTokens, "Pending tokens should be 3 (3 users)")
		}
	}
}

// TestDeleteToken testa exclusão de tokens
func (s *ElevationRepositoryTestSuite) TestDeleteToken() {
	// Cria token para teste de exclusão
	token := s.createTestToken("delete_user", "delete_tenant", "mozambique", elevation.StatusActive, false)
	err := s.repository.SaveToken(s.ctx, token)
	require.NoError(s.T(), err)
	
	// Verifica que o token existe
	_, err = s.repository.GetToken(s.ctx, token.ID)
	require.NoError(s.T(), err, "Token should exist before deletion")
	
	// Exclui o token
	err = s.repository.DeleteToken(s.ctx, token.ID)
	require.NoError(s.T(), err, "Failed to delete token")
	
	// Verifica que o token não existe mais
	_, err = s.repository.GetToken(s.ctx, token.ID)
	assert.Error(s.T(), err, "Token should not exist after deletion")
}

// RunTests executa os testes da suite
func TestElevationRepositoryTestSuite(t *testing.T) {
	// Pula testes se a variável de ambiente SKIP_INTEGRATION_TESTS estiver definida
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration tests")
	}

	suite.Run(t, new(ElevationRepositoryTestSuite))
}