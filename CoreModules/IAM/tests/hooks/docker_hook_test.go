package hooks

import (
	"context"
	"testing"
	"time"

	"github.com/innovabiz/iam/models"
	"github.com/innovabiz/iam/services/elevation/hooks"
	"github.com/innovabiz/iam/utils/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// MockMetadataProvider é um mock para o provedor de metadados de conformidade
type MockMetadataProvider struct {
	mock.Mock
}

func (m *MockMetadataProvider) GetMarketComplianceMetadata(ctx context.Context, market string, operationType string) (map[string]interface{}, error) {
	args := m.Called(ctx, market, operationType)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockMetadataProvider) GetTenantComplianceRules(ctx context.Context, tenantID string, market string) (map[string]interface{}, error) {
	args := m.Called(ctx, tenantID, market)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// TestDockerHook_ValidateScope testa a validação de escopo do Docker Hook
func TestDockerHook_ValidateScope(t *testing.T) {
	// Configuração de teste com mercados suportados
	logger := zap.NewNop()
	tp := trace.NewNoopTracerProvider()
	tracer := tp.Tracer("test")
	
	// Casos de teste para diferentes mercados e tenants
	tests := []struct {
		name         string
		market       string
		tenantID     string
		scope        string
		expectErr    bool
		errorMessage string
	}{
		{
			name:      "Angola - Escopo docker:run válido",
			market:    "angola",
			tenantID:  "tenant_angola_001",
			scope:     "docker:run",
			expectErr: false,
		},
		{
			name:         "Angola - Escopo inválido",
			market:       "angola",
			tenantID:     "tenant_angola_001",
			scope:        "docker:invalid",
			expectErr:    true,
			errorMessage: "escopo inválido",
		},
		{
			name:      "Brasil - Escopo docker:build válido",
			market:    "brasil",
			tenantID:  "tenant_brasil_001",
			scope:     "docker:build",
			expectErr: false,
		},
		{
			name:      "UE - Escopo docker:exec válido",
			market:    "eu",
			tenantID:  "tenant_eu_001",
			scope:     "docker:exec",
			expectErr: false,
		},
		{
			name:      "Moçambique - Escopo docker:run válido",
			market:    "mocambique",
			tenantID:  "tenant_mocambique_001",
			scope:     "docker:run",
			expectErr: false,
		},
		{
			name:         "China - Escopo restrito para o mercado",
			market:       "china",
			tenantID:     "tenant_china_001",
			scope:        "docker:system",
			expectErr:    true,
			errorMessage: "escopo restrito para o mercado",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configuração do mock
			mockProvider := new(MockMetadataProvider)
			
			// Configurar comportamento esperado do mock para diferentes mercados
			mockMetadata := make(map[string]interface{})
			mockRules := make(map[string]interface{})
			
			// Configurações específicas para cada mercado
			switch tt.market {
			case "angola":
				mockMetadata["requires_dual_approval"] = true
				mockMetadata["audit_retention_years"] = 7
				mockRules["allow_run"] = true
				mockRules["allow_build"] = true
				mockRules["allow_exec"] = false
			case "brasil":
				mockMetadata["requires_lgpd_compliance"] = true
				mockRules["allow_run"] = true
				mockRules["allow_build"] = true
				mockRules["allow_exec"] = true
			case "eu":
				mockMetadata["requires_gdpr_compliance"] = true
				mockRules["allow_run"] = true
				mockRules["allow_build"] = true
				mockRules["allow_exec"] = true
			case "china":
				mockMetadata["requires_local_storage"] = true
				mockRules["allow_run"] = true
				mockRules["allow_build"] = true
				mockRules["allow_system"] = false
			default:
				mockRules["allow_run"] = true
				mockRules["allow_build"] = true
			}
			
			mockProvider.On("GetMarketComplianceMetadata", mock.Anything, tt.market, "docker").Return(mockMetadata, nil)
			mockProvider.On("GetTenantComplianceRules", mock.Anything, tt.tenantID, tt.market).Return(mockRules, nil)

			// Criar hook com dependências mockadas
			observabilityService := observability.NewObservabilityService(logger, tracer)
			hook := hooks.NewDockerHook(observabilityService, mockProvider)
			
			// Executar o teste
			ctx := context.Background()
			result, err := hook.ValidateScope(ctx, tt.scope, tt.tenantID, tt.market)
			
			// Verificar resultados
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// TestDockerHook_RequiresMFA testa a validação de requisitos MFA
func TestDockerHook_RequiresMFA(t *testing.T) {
	// Configuração de teste
	logger := zap.NewNop()
	tp := trace.NewNoopTracerProvider()
	tracer := tp.Tracer("test")

	tests := []struct {
		name      string
		market    string
		tenantID  string
		scope     string
		expectMFA bool
		mfaLevel  string
	}{
		{
			name:      "Angola - MFA forte requerido para docker:run",
			market:    "angola",
			tenantID:  "tenant_angola_001",
			scope:     "docker:run",
			expectMFA: true,
			mfaLevel:  "forte",
		},
		{
			name:      "Brasil - MFA forte requerido para docker:system",
			market:    "brasil",
			tenantID:  "tenant_brasil_001",
			scope:     "docker:system",
			expectMFA: true,
			mfaLevel:  "forte",
		},
		{
			name:      "UE - MFA forte requerido para docker:exec",
			market:    "eu",
			tenantID:  "tenant_eu_001",
			scope:     "docker:exec",
			expectMFA: true,
			mfaLevel:  "forte",
		},
		{
			name:      "Moçambique - MFA médio para docker:build",
			market:    "mocambique",
			tenantID:  "tenant_mocambique_001",
			scope:     "docker:build",
			expectMFA: true,
			mfaLevel:  "medio",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configuração do mock
			mockProvider := new(MockMetadataProvider)
			
			// Configurar comportamento esperado do mock
			mockMetadata := make(map[string]interface{})
			mockRules := make(map[string]interface{})
			
			// Configurações específicas para cada mercado
			switch tt.market {
			case "angola":
				mockMetadata["mfa_level"] = "forte"
			case "brasil":
				mockMetadata["mfa_level"] = "forte"
			case "eu":
				mockMetadata["mfa_level"] = "forte"
			case "mocambique":
				mockMetadata["mfa_level"] = "medio"
			default:
				mockMetadata["mfa_level"] = "padrao"
			}
			
			mockProvider.On("GetMarketComplianceMetadata", mock.Anything, tt.market, "docker").Return(mockMetadata, nil)
			mockProvider.On("GetTenantComplianceRules", mock.Anything, tt.tenantID, tt.market).Return(mockRules, nil)

			// Criar hook com dependências mockadas
			observabilityService := observability.NewObservabilityService(logger, tracer)
			hook := hooks.NewDockerHook(observabilityService, mockProvider)
			
			// Executar o teste
			ctx := context.Background()
			requiresMFA, mfaLevel, err := hook.RequiresMFA(ctx, tt.scope, tt.tenantID, tt.market)
			
			// Verificar resultados
			assert.NoError(t, err)
			assert.Equal(t, tt.expectMFA, requiresMFA)
			if tt.expectMFA {
				assert.Equal(t, tt.mfaLevel, mfaLevel)
			}
		})
	}
}

// TestDockerHook_GetApprovers testa a obtenção de aprovadores
func TestDockerHook_GetApprovers(t *testing.T) {
	// Configuração de teste
	logger := zap.NewNop()
	tp := trace.NewNoopTracerProvider()
	tracer := tp.Tracer("test")

	tests := []struct {
		name           string
		market         string
		tenantID       string
		scope          string
		userID         string
		expectedLevels int
	}{
		{
			name:           "Angola - Aprovação dupla para docker:system",
			market:         "angola",
			tenantID:       "tenant_angola_001",
			scope:          "docker:system",
			userID:         "user_001",
			expectedLevels: 2,
		},
		{
			name:           "Brasil - Aprovação simples para docker:run",
			market:         "brasil",
			tenantID:       "tenant_brasil_001",
			scope:          "docker:run",
			userID:         "user_002",
			expectedLevels: 1,
		},
		{
			name:           "UE - Aprovação dupla para docker:exec",
			market:         "eu",
			tenantID:       "tenant_eu_001",
			scope:          "docker:exec",
			userID:         "user_003",
			expectedLevels: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configuração do mock
			mockProvider := new(MockMetadataProvider)
			
			// Configurar comportamento esperado do mock
			mockMetadata := make(map[string]interface{})
			mockRules := make(map[string]interface{})
			
			// Configurações específicas para cada mercado
			switch tt.market {
			case "angola":
				mockMetadata["approval_levels"] = 2
			case "brasil":
				mockMetadata["approval_levels"] = 1
			case "eu":
				mockMetadata["approval_levels"] = 2
			default:
				mockMetadata["approval_levels"] = 1
			}
			
			mockProvider.On("GetMarketComplianceMetadata", mock.Anything, tt.market, "docker").Return(mockMetadata, nil)
			mockProvider.On("GetTenantComplianceRules", mock.Anything, tt.tenantID, tt.market).Return(mockRules, nil)

			// Criar hook com dependências mockadas
			observabilityService := observability.NewObservabilityService(logger, tracer)
			hook := hooks.NewDockerHook(observabilityService, mockProvider)
			
			// Executar o teste
			ctx := context.Background()
			approvers, err := hook.GetApprovers(ctx, tt.scope, tt.userID, tt.tenantID, tt.market)
			
			// Verificar resultados
			assert.NoError(t, err)
			assert.NotNil(t, approvers)
			assert.Equal(t, tt.expectedLevels, len(approvers))
		})
	}
}

// TestDockerHook_ValidateTokenUsage testa a validação de uso de token
func TestDockerHook_ValidateTokenUsage(t *testing.T) {
	// Configuração de teste
	logger := zap.NewNop()
	tp := trace.NewNoopTracerProvider()
	tracer := tp.Tracer("test")

	// Mock do token e request
	mockToken := &models.ElevationToken{
		ID:        "token_001",
		Scope:     "docker:run",
		UserID:    "user_001",
		TenantID:  "tenant_angola_001",
		Market:    "angola",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	
	// Casos de teste para diferentes cenários
	tests := []struct {
		name      string
		token     *models.ElevationToken
		request   map[string]interface{}
		expectErr bool
	}{
		{
			name:  "Angola - Uso válido de docker:run",
			token: mockToken,
			request: map[string]interface{}{
				"command": "docker run --rm ubuntu:latest echo hello",
				"user_id": "user_001",
			},
			expectErr: false,
		},
		{
			name:  "Angola - Comando proibido",
			token: mockToken,
			request: map[string]interface{}{
				"command": "docker run --privileged --rm ubuntu:latest echo hello",
				"user_id": "user_001",
			},
			expectErr: true,
		},
		{
			name:  "Angola - Usuário diferente",
			token: mockToken,
			request: map[string]interface{}{
				"command": "docker run --rm ubuntu:latest echo hello",
				"user_id": "user_002", // Usuário diferente do token
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configuração do mock
			mockProvider := new(MockMetadataProvider)
			
			// Configurar comportamento esperado do mock
			mockMetadata := make(map[string]interface{})
			mockRules := make(map[string]interface{})
			
			// Adicionar regras específicas para Angola
			mockMetadata["blocked_flags"] = []string{"--privileged", "--cap-add=SYS_ADMIN"}
			
			mockProvider.On("GetMarketComplianceMetadata", mock.Anything, tt.token.Market, "docker").Return(mockMetadata, nil)
			mockProvider.On("GetTenantComplianceRules", mock.Anything, tt.token.TenantID, tt.token.Market).Return(mockRules, nil)

			// Criar hook com dependências mockadas
			observabilityService := observability.NewObservabilityService(logger, tracer)
			hook := hooks.NewDockerHook(observabilityService, mockProvider)
			
			// Executar o teste
			ctx := context.Background()
			valid, auditMetadata, err := hook.ValidateTokenUsage(ctx, tt.token, tt.request)
			
			// Verificar resultados
			if tt.expectErr {
				assert.Error(t, err)
				assert.False(t, valid)
			} else {
				assert.NoError(t, err)
				assert.True(t, valid)
				assert.NotNil(t, auditMetadata)
				
				// Verificar metadados específicos de auditoria para Angola
				if tt.token.Market == "angola" {
					assert.Contains(t, auditMetadata, "audit_retention_years")
					assert.Equal(t, float64(7), auditMetadata["audit_retention_years"])
				}
			}
		})
	}
}

// TestDockerHook_GenerateAuditMetadata testa a geração de metadados de auditoria
func TestDockerHook_GenerateAuditMetadata(t *testing.T) {
	// Configuração de teste
	logger := zap.NewNop()
	tp := trace.NewNoopTracerProvider()
	tracer := tp.Tracer("test")

	// Mock do token e request
	mockToken := &models.ElevationToken{
		ID:        "token_001",
		Scope:     "docker:run",
		UserID:    "user_001",
		TenantID:  "tenant_angola_001",
		Market:    "angola",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	
	// Configuração do request para diferentes mercados
	tests := []struct {
		name            string
		token           *models.ElevationToken
		request         map[string]interface{}
		market          string
		expectedMetadata []string
	}{
		{
			name:  "Angola - Metadados específicos BNA",
			token: mockToken,
			request: map[string]interface{}{
				"command": "docker run --rm ubuntu:latest echo hello",
				"user_id": "user_001",
			},
			market: "angola",
			expectedMetadata: []string{
				"audit_retention_years",
				"requires_dual_approval",
				"operation_timestamp",
				"command_hash",
			},
		},
		{
			name:  "Brasil - Metadados LGPD",
			token: &models.ElevationToken{
				ID:        "token_002",
				Scope:     "docker:run",
				UserID:    "user_002",
				TenantID:  "tenant_brasil_001",
				Market:    "brasil",
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(1 * time.Hour),
			},
			request: map[string]interface{}{
				"command": "docker run --rm ubuntu:latest echo hello",
				"user_id": "user_002",
			},
			market: "brasil",
			expectedMetadata: []string{
				"lgpd_justification",
				"data_purpose",
				"operation_timestamp",
				"command_hash",
			},
		},
		{
			name:  "UE - Metadados GDPR",
			token: &models.ElevationToken{
				ID:        "token_003",
				Scope:     "docker:run",
				UserID:    "user_003",
				TenantID:  "tenant_eu_001",
				Market:    "eu",
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(1 * time.Hour),
			},
			request: map[string]interface{}{
				"command": "docker run --rm ubuntu:latest echo hello",
				"user_id": "user_003",
			},
			market: "eu",
			expectedMetadata: []string{
				"gdpr_legal_basis",
				"data_minimization",
				"operation_timestamp",
				"command_hash",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configuração do mock
			mockProvider := new(MockMetadataProvider)
			
			// Configurar comportamento esperado do mock para cada mercado
			mockMetadata := make(map[string]interface{})
			mockRules := make(map[string]interface{})
			
			// Configurações específicas para cada mercado
			switch tt.market {
			case "angola":
				mockMetadata["audit_retention_years"] = 7
				mockMetadata["requires_dual_approval"] = true
			case "brasil":
				mockMetadata["lgpd_justification"] = "Execução de contrato"
				mockMetadata["data_purpose"] = "Suporte técnico"
			case "eu":
				mockMetadata["gdpr_legal_basis"] = "Legítimo interesse"
				mockMetadata["data_minimization"] = true
			}
			
			mockProvider.On("GetMarketComplianceMetadata", mock.Anything, tt.token.Market, "docker").Return(mockMetadata, nil)
			mockProvider.On("GetTenantComplianceRules", mock.Anything, tt.token.TenantID, tt.token.Market).Return(mockRules, nil)

			// Criar hook com dependências mockadas
			observabilityService := observability.NewObservabilityService(logger, tracer)
			hook := hooks.NewDockerHook(observabilityService, mockProvider)
			
			// Executar o teste
			ctx := context.Background()
			auditMetadata, err := hook.GenerateAuditMetadata(ctx, tt.token, tt.request)
			
			// Verificar resultados
			assert.NoError(t, err)
			assert.NotNil(t, auditMetadata)
			
			// Verificar campos específicos para cada mercado
			for _, field := range tt.expectedMetadata {
				assert.Contains(t, auditMetadata, field)
			}
		})
	}
}