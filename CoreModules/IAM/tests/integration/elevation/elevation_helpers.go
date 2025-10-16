// Package elevation contém testes de integração para o sistema de elevação de privilégios MCP-IAM.
// Este arquivo contém funções auxiliares para os testes de elevação.
package elevation

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/innovabiz/iam/auth"
	"github.com/innovabiz/iam/core/tenant"
	"github.com/innovabiz/iam/elevation"
	"github.com/innovabiz/tests/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// seedTestData insere dados iniciais necessários para os testes de elevação
func seedTestData(db *sql.DB) error {
	// Mercados para teste
	markets := []string{"angola", "brasil", "mocambique"}
	
	// Inserir mercados de teste
	for _, market := range markets {
		_, err := db.Exec(
			"INSERT INTO markets (code, name, status, created_at) VALUES ($1, $2, 'active', NOW()) ON CONFLICT DO NOTHING",
			market, fmt.Sprintf("Mercado %s", market),
		)
		if err != nil {
			return fmt.Errorf("falha ao inserir mercado %s: %w", market, err)
		}
	}
	
	// Inserir tenants de teste
	tenants := []struct {
		ID           string
		Name         string
		Code         string
		PrimaryMarket string
		Status       string
	}{
		{"tenant-angola-123", "Banco Angola", "banco-angola", "angola", "active"},
		{"tenant-brasil-456", "Fintech Brasil", "fintech-brasil", "brasil", "active"},
		{"tenant-mocambique-789", "Pagamentos Moçambique", "pagamentos-mz", "mocambique", "active"},
	}
	
	for _, t := range tenants {
		_, err := db.Exec(
			`INSERT INTO tenants 
			(id, name, code, primary_market, status, created_at) 
			VALUES ($1, $2, $3, $4, $5, NOW()) ON CONFLICT DO NOTHING`,
			t.ID, t.Name, t.Code, t.PrimaryMarket, t.Status,
		)
		if err != nil {
			return fmt.Errorf("falha ao inserir tenant %s: %w", t.ID, err)
		}
	}
	
	// Inserir usuários de teste
	users := []struct {
		ID       string
		Username string
		Email    string
		Name     string
		TenantID string
		Roles    []string
	}{
		{
			ID:       "user-admin-456",
			Username: "admin",
			Email:    "admin@innovabiz.ao",
			Name:     "Administrador",
			TenantID: "tenant-angola-123",
			Roles:    []string{"admin", "user"},
		},
		{
			ID:       "user-supervisor-789",
			Username: "supervisor",
			Email:    "supervisor@innovabiz.ao",
			Name:     "Supervisor",
			TenantID: "tenant-angola-123",
			Roles:    []string{"supervisor", "security_admin"},
		},
		{
			ID:       "user-regular-123",
			Username: "usuario",
			Email:    "usuario@innovabiz.ao",
			Name:     "Usuário Regular",
			TenantID: "tenant-angola-123",
			Roles:    []string{"user"},
		},
	}
	
	for _, u := range users {
		// Inserir usuário
		_, err := db.Exec(
			`INSERT INTO users 
			(id, username, email, display_name, status, password_hash, created_at) 
			VALUES ($1, $2, $3, $4, 'active', '$2a$10$hashhashhashhashhashhash', NOW()) ON CONFLICT DO NOTHING`,
			u.ID, u.Username, u.Email, u.Name,
		)
		if err != nil {
			return fmt.Errorf("falha ao inserir usuário %s: %w", u.ID, err)
		}
		
		// Inserir associação tenant-usuário
		_, err = db.Exec(
			`INSERT INTO tenant_users 
			(tenant_id, user_id, status, created_at) 
			VALUES ($1, $2, 'active', NOW()) ON CONFLICT DO NOTHING`,
			u.TenantID, u.ID,
		)
		if err != nil {
			return fmt.Errorf("falha ao inserir associação tenant-usuário %s-%s: %w", u.TenantID, u.ID, err)
		}
		
		// Inserir roles para o usuário
		for _, role := range u.Roles {
			_, err = db.Exec(
				`INSERT INTO user_roles 
				(user_id, tenant_id, role_name, created_at) 
				VALUES ($1, $2, $3, NOW()) ON CONFLICT DO NOTHING`,
				u.ID, u.TenantID, role,
			)
			if err != nil {
				return fmt.Errorf("falha ao inserir role %s para usuário %s: %w", role, u.ID, err)
			}
		}
	}
	
	// Inserir políticas de elevação para comandos sensíveis
	elevationPolicies := []struct {
		Scope        string
		Market       string
		RequiresApproval bool
		RequiresMFA  bool
		MaxDuration  int
	}{
		// Políticas para Docker
		{"docker:container:exec", "angola", true, true, 60},
		{"docker:container:exec", "brasil", true, true, 45},
		{"docker:container:exec", "mocambique", true, false, 90},
		{"docker:volume:mount", "angola", true, true, 60},
		{"docker:volume:mount", "brasil", true, true, 45},
		{"docker:volume:mount", "mocambique", true, false, 90},
		
		// Políticas para Desktop Commander
		{"desktop:execute_command:sudo", "angola", true, true, 30},
		{"desktop:execute_command:sudo", "brasil", true, true, 30},
		{"desktop:execute_command:sudo", "mocambique", true, false, 60},
		{"desktop:edit_config", "angola", true, true, 45},
		{"desktop:edit_config", "brasil", true, true, 45},
		{"desktop:edit_config", "mocambique", true, false, 60},
		
		// Políticas para GitHub
		{"github:merge:main", "angola", true, false, 120},
		{"github:merge:main", "brasil", true, false, 120},
		{"github:merge:main", "mocambique", true, false, 120},
		{"github:push:force", "angola", true, true, 60},
		{"github:push:force", "brasil", true, true, 45},
		{"github:push:force", "mocambique", true, false, 90},
		
		// Políticas para Figma
		{"figma:delete:file", "angola", true, false, 120},
		{"figma:delete:file", "brasil", true, false, 120},
		{"figma:delete:file", "mocambique", false, false, 120},
	}
	
	for _, p := range elevationPolicies {
		_, err := db.Exec(
			`INSERT INTO elevation_policies 
			(scope, market, requires_approval, requires_mfa, max_duration_minutes, created_at) 
			VALUES ($1, $2, $3, $4, $5, NOW()) ON CONFLICT DO NOTHING`,
			p.Scope, p.Market, p.RequiresApproval, p.RequiresMFA, p.MaxDuration,
		)
		if err != nil {
			return fmt.Errorf("falha ao inserir política de elevação %s-%s: %w", p.Scope, p.Market, err)
		}
	}
	
	return nil
}

// setupElevationService configura o serviço de elevação para testes
func setupElevationService() (*elevation.Service, error) {
	logger, err := testutil.NewTestLogger("elevation_service")
	if err != nil {
		return nil, fmt.Errorf("falha ao criar logger para serviço de elevação: %w", err)
	}
	
	tracer, closer, err := testutil.NewTestTracer("elevation_service")
	if err != nil {
		return nil, fmt.Errorf("falha ao criar tracer para serviço de elevação: %w", err)
	}
	defer closer.Close()
	
	metrics := testutil.NewTestMetrics("elevation_service")
	
	obs := testutil.NewObservability(logger, tracer, metrics)
	
	// Configura repositório de elevação (mock para testes)
	repo := elevation.NewMockRepository()
	
	// Configura cliente de notificações (mock para testes)
	notifier := elevation.NewMockNotifier()
	
	// Configura serviço de auditoria (mock para testes)
	auditService := testutil.NewMockAuditService()
	
	// Configura serviço de tenant
	tenantService := tenant.NewMockService()
	
	// Configura serviço de elevação
	service := elevation.NewService(
		elevation.WithRepository(repo),
		elevation.WithNotifier(notifier),
		elevation.WithAuditService(auditService),
		elevation.WithTenantService(tenantService),
		elevation.WithObservability(obs),
	)
	
	return service, nil
}

// createTestAuthToken cria um token de autenticação simulado para testes
func createTestAuthToken(t *testing.T, userID, tenantID, market string) *auth.Token {
	token := &auth.Token{
		UserID:       userID,
		TenantID:     tenantID,
		Market:       market,
		Roles:        []string{"admin", "user"},
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		IssuedAt:     time.Now().Unix(),
		Issuer:       "innovabiz-iam-test",
		TokenType:    auth.TokenTypeAccess,
		SessionID:    uuid.New().String(),
		MFACompleted: true,
	}
	
	return token
}

// createSupervisorContext cria um contexto simulando um supervisor para aprovações
func createSupervisorContext(ctx context.Context, supervisorID, tenantID, market string) context.Context {
	supervisorToken := &auth.Token{
		UserID:       supervisorID,
		TenantID:     tenantID,
		Market:       market,
		Roles:        []string{"supervisor", "security_admin"},
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		IssuedAt:     time.Now().Unix(),
		Issuer:       "innovabiz-iam-test",
		TokenType:    auth.TokenTypeAccess,
		SessionID:    uuid.New().String(),
		MFACompleted: true,
	}
	
	return auth.ContextWithToken(ctx, supervisorToken)
}

// getOrCreateTestElevationToken obtém ou cria um token de elevação válido para testes
func getOrCreateTestElevationToken(ctx context.Context, service *elevation.Service, userID, tenantID, market string) (*elevation.Token, error) {
	// Verifica se já existe uma elevação ativa para o usuário
	tokens, err := service.ListActiveElevationsForUser(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}
	
	// Se encontrou, retorna o primeiro token ativo
	if len(tokens) > 0 {
		return tokens[0], nil
	}
	
	// Se não encontrou, cria um novo token
	// Solicita elevação
	elevationRequest := &elevation.ElevationRequest{
		UserID:       userID,
		TenantID:     tenantID,
		Justification: "Teste automatizado de elevação",
		Scopes:       []string{"docker:container:exec", "desktop:execute_command:sudo"},
		Duration:     30, // 30 minutos
		Emergency:    false,
	}
	
	pendingToken, err := service.RequestElevation(ctx, elevationRequest)
	if err != nil {
		return nil, err
	}
	
	// Simula aprovação por supervisor
	supervisorID := "user-supervisor-789"
	supervisorCtx := createSupervisorContext(ctx, supervisorID, tenantID, market)
	
	approvedToken, err := service.ApproveElevation(supervisorCtx, pendingToken.ID)
	if err != nil {
		return nil, err
	}
	
	return approvedToken, nil
}

// getTestExpiredElevationToken cria um token de elevação expirado para testes
func getTestExpiredElevationToken(t *testing.T, userID, tenantID, market string) *elevation.Token {
	now := time.Now()
	return &elevation.Token{
		ID:           uuid.New().String(),
		UserID:       userID,
		TenantID:     tenantID,
		Market:       market,
		Scopes:       []string{"docker:container:exec"},
		Status:       elevation.StatusActive,
		CreatedAt:    now.Add(-2 * time.Hour),
		ExpiresAt:    now.Add(-1 * time.Hour), // Expirado há 1 hora
		ApprovedBy:   "user-supervisor-789",
		Justification: "Token expirado para teste",
		Emergency:    false,
	}
}