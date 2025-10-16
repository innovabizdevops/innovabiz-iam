package hooks

import (
	"context"
	"testing"
	"time"

	"github.com/innovabiz/iam/models"
	"github.com/innovabiz/iam/services/elevation"
	"github.com/innovabiz/iam/services/elevation/hooks"
	"github.com/innovabiz/iam/utils/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// MockElevationStore é um mock para o repositório de tokens de elevação
type MockElevationStore struct {
	mock.Mock
}

func (m *MockElevationStore) CreateElevationRequest(ctx context.Context, req *models.ElevationRequest) (string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Error(1)
}

func (m *MockElevationStore) GetElevationRequest(ctx context.Context, id string) (*models.ElevationRequest, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.ElevationRequest), args.Error(1)
}

func (m *MockElevationStore) UpdateElevationRequest(ctx context.Context, req *models.ElevationRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockElevationStore) CreateElevationToken(ctx context.Context, token *models.ElevationToken) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}

func (m *MockElevationStore) GetElevationToken(ctx context.Context, id string) (*models.ElevationToken, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.ElevationToken), args.Error(1)
}

func (m *MockElevationStore) UpdateElevationToken(ctx context.Context, token *models.ElevationToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockElevationStore) GetElevationTokenByUserScopeMarket(ctx context.Context, userID, scope, market string) (*models.ElevationToken, error) {
	args := m.Called(ctx, userID, scope, market)
	return args.Get(0).(*models.ElevationToken), args.Error(1)
}

// MockUserService é um mock para o serviço de usuários
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) VerifyMFA(ctx context.Context, userID string, mfaCode string, mfaLevel string) (bool, error) {
	args := m.Called(ctx, userID, mfaCode, mfaLevel)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserService) GetApprovers(ctx context.Context, tenantID string, market string, level int) ([]string, error) {
	args := m.Called(ctx, tenantID, market, level)
	return args.Get(0).([]string), args.Error(1)
}

// MockAuditService é um mock para o serviço de auditoria
type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) LogElevationRequest(ctx context.Context, req *models.ElevationRequest, metadata map[string]interface{}) error {
	args := m.Called(ctx, req, metadata)
	return args.Error(0)
}

func (m *MockAuditService) LogElevationApproval(ctx context.Context, req *models.ElevationRequest, approverID string, metadata map[string]interface{}) error {
	args := m.Called(ctx, req, approverID, metadata)
	return args.Error(0)
}

func (m *MockAuditService) LogElevationUsage(ctx context.Context, token *models.ElevationToken, metadata map[string]interface{}) error {
	args := m.Called(ctx, token, metadata)
	return args.Error(0)
}

// TestFluxoCompleto_ElevacaoAngola testa o fluxo completo de elevação para o mercado de Angola
func TestFluxoCompleto_ElevacaoAngola(t *testing.T) {
	// Configuração inicial
	logger := zap.NewNop()
	tp := trace.NewNoopTracerProvider()
	tracer := tp.Tracer("test")
	observabilityService := observability.NewObservabilityService(logger, tracer)
	
	// Configuração de mocks
	mockMetadataProvider := new(MockMetadataProvider)
	mockElevationStore := new(MockElevationStore)
	mockUserService := new(MockUserService)
	mockAuditService := new(MockAuditService)
	
	// Criar hooks mockados
	dockerHook := hooks.NewDockerHook(observabilityService, mockMetadataProvider)
	githubHook := hooks.NewGitHubHook(observabilityService, mockMetadataProvider)
	desktopHook := hooks.NewDesktopCommanderHook(observabilityService, mockMetadataProvider)
	figmaHook := hooks.NewFigmaHook(observabilityService, mockMetadataProvider)
	
	// Registrar hooks em um mapa
	hookRegistry := map[string]elevation.ElevationHook{
		"docker":  dockerHook,
		"github":  githubHook,
		"desktop": desktopHook,
		"figma":   figmaHook,
	}
	
	// Criar serviço de elevação
	elevationService := elevation.NewElevationService(
		observabilityService,
		mockElevationStore,
		mockUserService,
		mockAuditService,
		hookRegistry,
	)
	
	// Parâmetros do teste
	userID := "user_angola_001"
	tenantID := "tenant_angola_001"
	market := "angola"
	scope := "docker:run"
	requestID := "request_001"
	approverID1 := "approver_001"
	approverID2 := "approver_002"
	tokenID := "token_001"
	
	// Mock de retorno para metadados de conformidade do mercado de Angola
	mockMetadata := map[string]interface{}{
		"requires_dual_approval":  true,
		"audit_retention_years":   7,
		"mfa_level":               "forte",
		"approval_levels":         2,
		"blocked_flags":           []string{"--privileged", "--cap-add=SYS_ADMIN"},
		"protected_repositories":  []string{"/var/lib/system"},
		"protected_directories":   []string{"/etc/secrets"},
		"blocked_commands":        []string{"delete_directory"},
	}
	
	// Mock de retorno para regras específicas do tenant
	mockRules := map[string]interface{}{
		"allow_run":   true,
		"allow_build": true,
		"allow_exec":  false,
	}
	
	// Configurar comportamento dos mocks
	mockMetadataProvider.On("GetMarketComplianceMetadata", mock.Anything, market, "docker").Return(mockMetadata, nil)
	mockMetadataProvider.On("GetTenantComplianceRules", mock.Anything, tenantID, market).Return(mockRules, nil)
	
	// Mock de usuário
	mockUser := &models.User{
		ID:       userID,
		TenantID: tenantID,
		Market:   market,
		Email:    "usuario@angola.com",
		Name:     "Usuário de Angola",
	}
	mockUserService.On("GetUser", mock.Anything, userID).Return(mockUser, nil)
	
	// Mock para verificação de MFA
	mockUserService.On("VerifyMFA", mock.Anything, userID, "123456", "forte").Return(true, nil)
	
	// Mock para obtenção de aprovadores
	mockApprovers := []string{approverID1, approverID2}
	mockUserService.On("GetApprovers", mock.Anything, tenantID, market, 1).Return(mockApprovers, nil)
	
	// Mock para criação de request de elevação
	mockElevationStore.On("CreateElevationRequest", mock.Anything, mock.AnythingOfType("*models.ElevationRequest")).
		Return(requestID, nil)
	
	// Mock para atualização de request de elevação
	mockElevationStore.On("UpdateElevationRequest", mock.Anything, mock.AnythingOfType("*models.ElevationRequest")).
		Return(nil)
	
	// Mock para obtenção de request de elevação
	mockRequest := &models.ElevationRequest{
		ID:          requestID,
		UserID:      userID,
		Scope:       scope,
		TenantID:    tenantID,
		Market:      market,
		Status:      models.RequestStatusPending,
		CreatedAt:   time.Now(),
		ApprovalIDs: mockApprovers,
	}
	mockElevationStore.On("GetElevationRequest", mock.Anything, requestID).Return(mockRequest, nil)
	
	// Mock para criação de token de elevação
	mockElevationStore.On("CreateElevationToken", mock.Anything, mock.AnythingOfType("*models.ElevationToken")).
		Return(tokenID, nil)
	
	// Mock para obtenção de token de elevação
	mockToken := &models.ElevationToken{
		ID:        tokenID,
		RequestID: requestID,
		UserID:    userID,
		Scope:     scope,
		TenantID:  tenantID,
		Market:    market,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Status:    models.TokenStatusActive,
	}
	mockElevationStore.On("GetElevationToken", mock.Anything, tokenID).Return(mockToken, nil)
	mockElevationStore.On("GetElevationTokenByUserScopeMarket", mock.Anything, userID, scope, market).
		Return((*models.ElevationToken)(nil), models.ErrNotFound)
	
	// Mock para logging de auditoria
	mockAuditService.On("LogElevationRequest", mock.Anything, mock.AnythingOfType("*models.ElevationRequest"), mock.Anything).
		Return(nil)
	mockAuditService.On("LogElevationApproval", mock.Anything, mock.AnythingOfType("*models.ElevationRequest"), mock.Anything, mock.Anything).
		Return(nil)
	mockAuditService.On("LogElevationUsage", mock.Anything, mock.AnythingOfType("*models.ElevationToken"), mock.Anything).
		Return(nil)
	
	// Contexto do teste
	ctx := context.Background()
	
	// ETAPA 1: Criação do request de elevação
	t.Run("Criar request de elevação", func(t *testing.T) {
		reqID, err := elevationService.RequestElevation(ctx, userID, scope, tenantID, market)
		require.NoError(t, err)
		assert.Equal(t, requestID, reqID)
	})
	
	// ETAPA 2: Verificação de MFA
	t.Run("Verificar MFA", func(t *testing.T) {
		err := elevationService.VerifyMFA(ctx, requestID, "123456")
		require.NoError(t, err)
	})
	
	// ETAPA 3: Aprovação do request (primeira aprovação)
	t.Run("Primeira aprovação", func(t *testing.T) {
		err := elevationService.ApproveRequest(ctx, requestID, approverID1)
		require.NoError(t, err)
	})
	
	// ETAPA 4: Aprovação final (segunda aprovação)
	t.Run("Segunda aprovação", func(t *testing.T) {
		// Atualizar o mock do request para refletir a primeira aprovação
		updatedRequest := *mockRequest
		updatedRequest.Approvals = []string{approverID1}
		mockElevationStore.On("GetElevationRequest", mock.Anything, requestID).Return(&updatedRequest, nil)
		
		err := elevationService.ApproveRequest(ctx, requestID, approverID2)
		require.NoError(t, err)
	})
	
	// ETAPA 5: Uso do token
	t.Run("Uso do token", func(t *testing.T) {
		request := map[string]interface{}{
			"command": "docker run --rm ubuntu:latest echo hello",
			"user_id": userID,
		}
		valid, metadata, err := elevationService.ValidateTokenUsage(ctx, tokenID, request)
		require.NoError(t, err)
		assert.True(t, valid)
		assert.NotNil(t, metadata)
		assert.Contains(t, metadata, "audit_retention_years")
	})
}

// TestFluxoCompleto_ElevacaoUE testa o fluxo completo de elevação para o mercado da União Europeia
func TestFluxoCompleto_ElevacaoUE(t *testing.T) {
	// Configuração inicial
	logger := zap.NewNop()
	tp := trace.NewNoopTracerProvider()
	tracer := tp.Tracer("test")
	observabilityService := observability.NewObservabilityService(logger, tracer)
	
	// Configuração de mocks
	mockMetadataProvider := new(MockMetadataProvider)
	mockElevationStore := new(MockElevationStore)
	mockUserService := new(MockUserService)
	mockAuditService := new(MockAuditService)
	
	// Criar hook do GitHub
	githubHook := hooks.NewGitHubHook(observabilityService, mockMetadataProvider)
	
	// Registrar hooks em um mapa
	hookRegistry := map[string]elevation.ElevationHook{
		"github": githubHook,
	}
	
	// Criar serviço de elevação
	elevationService := elevation.NewElevationService(
		observabilityService,
		mockElevationStore,
		mockUserService,
		mockAuditService,
		hookRegistry,
	)
	
	// Parâmetros do teste
	userID := "user_eu_001"
	tenantID := "tenant_eu_001"
	market := "eu"
	scope := "github:admin"
	requestID := "request_eu_001"
	approverID1 := "approver_eu_001"
	approverID2 := "approver_eu_002"
	tokenID := "token_eu_001"
	
	// Mock de retorno para metadados de conformidade da UE
	mockMetadata := map[string]interface{}{
		"requires_dual_approval": true,
		"gdpr_legal_basis":       "Legítimo interesse",
		"data_minimization":      true,
		"mfa_level":              "forte",
		"approval_levels":        2,
		"protected_repositories": []string{"repo-financeiro", "repo-gdpr"},
		"blocked_operations":     []string{"delete_repo"},
	}
	
	// Mock de retorno para regras específicas do tenant
	mockRules := map[string]interface{}{
		"allow_read":  true,
		"allow_write": true,
		"allow_admin": true,
	}
	
	// Configurar comportamento dos mocks
	mockMetadataProvider.On("GetMarketComplianceMetadata", mock.Anything, market, "github").Return(mockMetadata, nil)
	mockMetadataProvider.On("GetTenantComplianceRules", mock.Anything, tenantID, market).Return(mockRules, nil)
	
	// Mock de usuário
	mockUser := &models.User{
		ID:       userID,
		TenantID: tenantID,
		Market:   market,
		Email:    "usuario@eu.europa",
		Name:     "Usuário da UE",
	}
	mockUserService.On("GetUser", mock.Anything, userID).Return(mockUser, nil)
	
	// Mock para verificação de MFA
	mockUserService.On("VerifyMFA", mock.Anything, userID, "654321", "forte").Return(true, nil)
	
	// Mock para obtenção de aprovadores
	mockApprovers := []string{approverID1, approverID2}
	mockUserService.On("GetApprovers", mock.Anything, tenantID, market, 1).Return(mockApprovers, nil)
	
	// Mock para criação de request de elevação
	mockElevationStore.On("CreateElevationRequest", mock.Anything, mock.AnythingOfType("*models.ElevationRequest")).
		Return(requestID, nil)
	
	// Mock para atualização de request de elevação
	mockElevationStore.On("UpdateElevationRequest", mock.Anything, mock.AnythingOfType("*models.ElevationRequest")).
		Return(nil)
	
	// Mock para obtenção de request de elevação
	mockRequest := &models.ElevationRequest{
		ID:          requestID,
		UserID:      userID,
		Scope:       scope,
		TenantID:    tenantID,
		Market:      market,
		Status:      models.RequestStatusPending,
		CreatedAt:   time.Now(),
		ApprovalIDs: mockApprovers,
	}
	mockElevationStore.On("GetElevationRequest", mock.Anything, requestID).Return(mockRequest, nil)
	
	// Mock para criação de token de elevação
	mockElevationStore.On("CreateElevationToken", mock.Anything, mock.AnythingOfType("*models.ElevationToken")).
		Return(tokenID, nil)
	
	// Mock para obtenção de token de elevação
	mockToken := &models.ElevationToken{
		ID:        tokenID,
		RequestID: requestID,
		UserID:    userID,
		Scope:     scope,
		TenantID:  tenantID,
		Market:    market,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Status:    models.TokenStatusActive,
	}
	mockElevationStore.On("GetElevationToken", mock.Anything, tokenID).Return(mockToken, nil)
	mockElevationStore.On("GetElevationTokenByUserScopeMarket", mock.Anything, userID, scope, market).
		Return((*models.ElevationToken)(nil), models.ErrNotFound)
	
	// Mock para logging de auditoria
	mockAuditService.On("LogElevationRequest", mock.Anything, mock.AnythingOfType("*models.ElevationRequest"), mock.Anything).
		Return(nil)
	mockAuditService.On("LogElevationApproval", mock.Anything, mock.AnythingOfType("*models.ElevationRequest"), mock.Anything, mock.Anything).
		Return(nil)
	mockAuditService.On("LogElevationUsage", mock.Anything, mock.AnythingOfType("*models.ElevationToken"), mock.Anything).
		Return(nil)
	
	// Contexto do teste
	ctx := context.Background()
	
	// ETAPA 1: Criação do request de elevação com GDPR justification
	t.Run("Criar request de elevação GDPR", func(t *testing.T) {
		reqID, err := elevationService.RequestElevation(ctx, userID, scope, tenantID, market)
		require.NoError(t, err)
		assert.Equal(t, requestID, reqID)
	})
	
	// ETAPA 2: Verificação de MFA
	t.Run("Verificar MFA", func(t *testing.T) {
		err := elevationService.VerifyMFA(ctx, requestID, "654321")
		require.NoError(t, err)
	})
	
	// ETAPA 3: Aprovação do request (primeira aprovação)
	t.Run("Primeira aprovação", func(t *testing.T) {
		err := elevationService.ApproveRequest(ctx, requestID, approverID1)
		require.NoError(t, err)
	})
	
	// ETAPA 4: Aprovação final (segunda aprovação)
	t.Run("Segunda aprovação", func(t *testing.T) {
		// Atualizar o mock do request para refletir a primeira aprovação
		updatedRequest := *mockRequest
		updatedRequest.Approvals = []string{approverID1}
		mockElevationStore.On("GetElevationRequest", mock.Anything, requestID).Return(&updatedRequest, nil)
		
		err := elevationService.ApproveRequest(ctx, requestID, approverID2)
		require.NoError(t, err)
	})
	
	// ETAPA 5: Uso do token com validação GDPR
	t.Run("Uso do token com validação GDPR", func(t *testing.T) {
		request := map[string]interface{}{
			"operation":  "merge",
			"repository": "repo-produto",
			"user_id":    userID,
			"gdpr_data":  false,
		}
		valid, metadata, err := elevationService.ValidateTokenUsage(ctx, tokenID, request)
		require.NoError(t, err)
		assert.True(t, valid)
		assert.NotNil(t, metadata)
		assert.Contains(t, metadata, "gdpr_legal_basis")
		assert.Contains(t, metadata, "data_minimization")
	})
}

// TestFluxoRejeitado_Brasil testa um cenário de elevação rejeitada para o mercado Brasil
func TestFluxoRejeitado_Brasil(t *testing.T) {
	// Configuração inicial
	logger := zap.NewNop()
	tp := trace.NewNoopTracerProvider()
	tracer := tp.Tracer("test")
	observabilityService := observability.NewObservabilityService(logger, tracer)
	
	// Configuração de mocks
	mockMetadataProvider := new(MockMetadataProvider)
	mockElevationStore := new(MockElevationStore)
	mockUserService := new(MockUserService)
	mockAuditService := new(MockAuditService)
	
	// Criar hook do Desktop Commander
	desktopHook := hooks.NewDesktopCommanderHook(observabilityService, mockMetadataProvider)
	
	// Registrar hooks em um mapa
	hookRegistry := map[string]elevation.ElevationHook{
		"desktop": desktopHook,
	}
	
	// Criar serviço de elevação
	elevationService := elevation.NewElevationService(
		observabilityService,
		mockElevationStore,
		mockUserService,
		mockAuditService,
		hookRegistry,
	)
	
	// Parâmetros do teste
	userID := "user_brasil_001"
	tenantID := "tenant_brasil_001"
	market := "brasil"
	scope := "desktop:system"  // Escopo restrito
	requestID := "request_br_001"
	
	// Mock de retorno para metadados de conformidade do Brasil
	mockMetadata := map[string]interface{}{
		"lgpd_justification": "Execução de contrato",
		"data_purpose":       "Operação técnica",
		"mfa_level":          "forte",
		"approval_levels":    1,
	}
	
	// Mock de retorno para regras específicas do tenant - não permite desktop:system
	mockRules := map[string]interface{}{
		"allow_read":   true,
		"allow_write":  true,
		"allow_admin":  true,
		"allow_system": false,  // Escopo restrito
	}
	
	// Configurar comportamento dos mocks
	mockMetadataProvider.On("GetMarketComplianceMetadata", mock.Anything, market, "desktop").Return(mockMetadata, nil)
	mockMetadataProvider.On("GetTenantComplianceRules", mock.Anything, tenantID, market).Return(mockRules, nil)
	
	// Mock de usuário
	mockUser := &models.User{
		ID:       userID,
		TenantID: tenantID,
		Market:   market,
		Email:    "usuario@brasil.com",
		Name:     "Usuário do Brasil",
	}
	mockUserService.On("GetUser", mock.Anything, userID).Return(mockUser, nil)
	
	// Contexto do teste
	ctx := context.Background()
	
	// ETAPA 1: Tentativa de criar request com escopo não permitido
	t.Run("Tentativa de criar request com escopo não permitido", func(t *testing.T) {
		_, err := elevationService.RequestElevation(ctx, userID, scope, tenantID, market)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "escopo restrito para o mercado")
	})
}

// TestTokenExpirado_Mocambique testa um cenário de token expirado para o mercado de Moçambique
func TestTokenExpirado_Mocambique(t *testing.T) {
	// Configuração inicial
	logger := zap.NewNop()
	tp := trace.NewNoopTracerProvider()
	tracer := tp.Tracer("test")
	observabilityService := observability.NewObservabilityService(logger, tracer)
	
	// Configuração de mocks
	mockMetadataProvider := new(MockMetadataProvider)
	mockElevationStore := new(MockElevationStore)
	mockUserService := new(MockUserService)
	mockAuditService := new(MockAuditService)
	
	// Criar hook do Figma
	figmaHook := hooks.NewFigmaHook(observabilityService, mockMetadataProvider)
	
	// Registrar hooks em um mapa
	hookRegistry := map[string]elevation.ElevationHook{
		"figma": figmaHook,
	}
	
	// Criar serviço de elevação
	elevationService := elevation.NewElevationService(
		observabilityService,
		mockElevationStore,
		mockUserService,
		mockAuditService,
		hookRegistry,
	)
	
	// Parâmetros do teste
	userID := "user_mocambique_001"
	tenantID := "tenant_mocambique_001"
	market := "mocambique"
	scope := "figma:comment"
	tokenID := "token_mz_001"
	
	// Mock de token expirado
	expiredToken := &models.ElevationToken{
		ID:        tokenID,
		UserID:    userID,
		Scope:     scope,
		TenantID:  tenantID,
		Market:    market,
		CreatedAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt: time.Now().Add(-1 * time.Hour),  // Expirado
		Status:    models.TokenStatusActive,
	}
	
	// Mock para obtenção de token de elevação
	mockElevationStore.On("GetElevationToken", mock.Anything, tokenID).Return(expiredToken, nil)
	
	// Contexto do teste
	ctx := context.Background()
	
	// ETAPA 1: Tentativa de usar token expirado
	t.Run("Tentativa de usar token expirado", func(t *testing.T) {
		request := map[string]interface{}{
			"action":  "comment",
			"file_id": "figma_file_123",
			"user_id": userID,
		}
		valid, _, err := elevationService.ValidateTokenUsage(ctx, tokenID, request)
		require.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), "token expirado")
	})
}