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
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TestGitHubHook_ValidateScope testa a validação de escopo do GitHub Hook
func TestGitHubHook_ValidateScope(t *testing.T) {
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
			name:      "Angola - Escopo github:read válido",
			market:    "angola",
			tenantID:  "tenant_angola_001",
			scope:     "github:read",
			expectErr: false,
		},
		{
			name:         "Angola - Escopo inválido",
			market:       "angola",
			tenantID:     "tenant_angola_001",
			scope:        "github:invalid",
			expectErr:    true,
			errorMessage: "escopo inválido",
		},
		{
			name:      "Brasil - Escopo github:write válido",
			market:    "brasil",
			tenantID:  "tenant_brasil_001",
			scope:     "github:write",
			expectErr: false,
		},
		{
			name:      "UE - Escopo github:admin válido",
			market:    "eu",
			tenantID:  "tenant_eu_001",
			scope:     "github:admin",
			expectErr: false,
		},
		{
			name:         "China - Escopo restrito para o mercado",
			market:       "china",
			tenantID:     "tenant_china_001",
			scope:        "github:admin",
			expectErr:    true,
			errorMessage: "escopo restrito para o mercado",
		},
		{
			name:      "SADC - Escopo github:write válido",
			market:    "sadc",
			tenantID:  "tenant_sadc_001",
			scope:     "github:write",
			expectErr: false,
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
				mockRules["allow_read"] = true
				mockRules["allow_write"] = true
				mockRules["allow_admin"] = true
			case "brasil":
				mockMetadata["requires_lgpd_compliance"] = true
				mockRules["allow_read"] = true
				mockRules["allow_write"] = true
				mockRules["allow_admin"] = true
			case "eu":
				mockMetadata["requires_gdpr_compliance"] = true
				mockRules["allow_read"] = true
				mockRules["allow_write"] = true
				mockRules["allow_admin"] = true
			case "china":
				mockMetadata["requires_local_storage"] = true
				mockRules["allow_read"] = true
				mockRules["allow_write"] = true
				mockRules["allow_admin"] = false // Restrição específica para China
			case "sadc":
				mockRules["allow_read"] = true
				mockRules["allow_write"] = true
				mockRules["allow_admin"] = false
			default:
				mockRules["allow_read"] = true
				mockRules["allow_write"] = true
			}
			
			mockProvider.On("GetMarketComplianceMetadata", mock.Anything, tt.market, "github").Return(mockMetadata, nil)
			mockProvider.On("GetTenantComplianceRules", mock.Anything, tt.tenantID, tt.market).Return(mockRules, nil)

			// Criar hook com dependências mockadas
			observabilityService := observability.NewObservabilityService(logger, tracer)
			hook := hooks.NewGitHubHook(observabilityService, mockProvider)
			
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

// TestGitHubHook_RequiresMFA testa a validação de requisitos MFA
func TestGitHubHook_RequiresMFA(t *testing.T) {
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
			name:      "Angola - MFA forte requerido para github:admin",
			market:    "angola",
			tenantID:  "tenant_angola_001",
			scope:     "github:admin",
			expectMFA: true,
			mfaLevel:  "forte",
		},
		{
			name:      "Brasil - MFA forte requerido para github:admin",
			market:    "brasil",
			tenantID:  "tenant_brasil_001",
			scope:     "github:admin",
			expectMFA: true,
			mfaLevel:  "forte",
		},
		{
			name:      "UE - MFA forte requerido para github:admin",
			market:    "eu",
			tenantID:  "tenant_eu_001",
			scope:     "github:admin",
			expectMFA: true,
			mfaLevel:  "forte",
		},
		{
			name:      "Brasil - MFA médio para github:write",
			market:    "brasil",
			tenantID:  "tenant_brasil_001",
			scope:     "github:write",
			expectMFA: true,
			mfaLevel:  "medio",
		},
		{
			name:      "Angola - MFA padrão para github:read",
			market:    "angola",
			tenantID:  "tenant_angola_001",
			scope:     "github:read",
			expectMFA: true,
			mfaLevel:  "padrao",
		},
		{
			name:      "Moçambique - MFA médio para github:write",
			market:    "mocambique",
			tenantID:  "tenant_mocambique_001",
			scope:     "github:write",
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
			
			// Configurações específicas para cada escopo e mercado
			switch tt.scope {
			case "github:admin":
				mockMetadata["mfa_level"] = "forte"
			case "github:write":
				mockMetadata["mfa_level"] = "medio"
			default:
				mockMetadata["mfa_level"] = "padrao"
			}
			
			mockProvider.On("GetMarketComplianceMetadata", mock.Anything, tt.market, "github").Return(mockMetadata, nil)
			mockProvider.On("GetTenantComplianceRules", mock.Anything, tt.tenantID, tt.market).Return(mockRules, nil)

			// Criar hook com dependências mockadas
			observabilityService := observability.NewObservabilityService(logger, tracer)
			hook := hooks.NewGitHubHook(observabilityService, mockProvider)
			
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

// TestGitHubHook_GetApprovers testa a obtenção de aprovadores
func TestGitHubHook_GetApprovers(t *testing.T) {
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
			name:           "Angola - Aprovação dupla para github:admin",
			market:         "angola",
			tenantID:       "tenant_angola_001",
			scope:          "github:admin",
			userID:         "user_001",
			expectedLevels: 2,
		},
		{
			name:           "Brasil - Aprovação simples para github:write",
			market:         "brasil",
			tenantID:       "tenant_brasil_001",
			scope:          "github:write",
			userID:         "user_002",
			expectedLevels: 1,
		},
		{
			name:           "UE - Aprovação dupla para github:admin",
			market:         "eu",
			tenantID:       "tenant_eu_001",
			scope:          "github:admin",
			userID:         "user_003",
			expectedLevels: 2,
		},
		{
			name:           "SADC - Aprovação simples para github:read",
			market:         "sadc",
			tenantID:       "tenant_sadc_001",
			scope:          "github:read",
			userID:         "user_004",
			expectedLevels: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configuração do mock
			mockProvider := new(MockMetadataProvider)
			
			// Configurar comportamento esperado do mock
			mockMetadata := make(map[string]interface{})
			mockRules := make(map[string]interface{})
			
			// Configurações específicas para cada mercado e escopo
			switch {
			case tt.market == "angola" && tt.scope == "github:admin":
				mockMetadata["approval_levels"] = 2
			case tt.market == "eu" && tt.scope == "github:admin":
				mockMetadata["approval_levels"] = 2
			default:
				mockMetadata["approval_levels"] = 1
			}
			
			mockProvider.On("GetMarketComplianceMetadata", mock.Anything, tt.market, "github").Return(mockMetadata, nil)
			mockProvider.On("GetTenantComplianceRules", mock.Anything, tt.tenantID, tt.market).Return(mockRules, nil)

			// Criar hook com dependências mockadas
			observabilityService := observability.NewObservabilityService(logger, tracer)
			hook := hooks.NewGitHubHook(observabilityService, mockProvider)
			
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

// TestGitHubHook_ValidateTokenUsage testa a validação de uso de token
func TestGitHubHook_ValidateTokenUsage(t *testing.T) {
	// Configuração de teste
	logger := zap.NewNop()
	tp := trace.NewNoopTracerProvider()
	tracer := tp.Tracer("test")

	// Mock do token e request
	mockToken := &models.ElevationToken{
		ID:        "token_001",
		Scope:     "github:write",
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
			name:  "Angola - Uso válido de github:write",
			token: mockToken,
			request: map[string]interface{}{
				"operation": "push",
				"repository": "repo-teste",
				"user_id": "user_001",
			},
			expectErr: false,
		},
		{
			name:  "Angola - Operação proibida",
			token: mockToken,
			request: map[string]interface{}{
				"operation": "delete_repo",
				"repository": "repo-principal",
				"user_id": "user_001",
			},
			expectErr: true,
		},
		{
			name:  "Angola - Usuário diferente",
			token: mockToken,
			request: map[string]interface{}{
				"operation": "push",
				"repository": "repo-teste",
				"user_id": "user_002", // Usuário diferente do token
			},
			expectErr: true,
		},
		{
			name:  "Angola - Repositório protegido",
			token: mockToken,
			request: map[string]interface{}{
				"operation": "push",
				"repository": "repo-financeiro", // Repositório com proteção especial
				"user_id": "user_001",
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
			mockMetadata["protected_repositories"] = []string{"repo-financeiro", "repo-principal"}
			mockMetadata["blocked_operations"] = []string{"delete_repo", "delete_branch"}
			
			mockProvider.On("GetMarketComplianceMetadata", mock.Anything, tt.token.Market, "github").Return(mockMetadata, nil)
			mockProvider.On("GetTenantComplianceRules", mock.Anything, tt.token.TenantID, tt.token.Market).Return(mockRules, nil)

			// Criar hook com dependências mockadas
			observabilityService := observability.NewObservabilityService(logger, tracer)
			hook := hooks.NewGitHubHook(observabilityService, mockProvider)
			
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

// TestGitHubHook_GenerateAuditMetadata testa a geração de metadados de auditoria
func TestGitHubHook_GenerateAuditMetadata(t *testing.T) {
	// Configuração de teste
	logger := zap.NewNop()
	tp := trace.NewNoopTracerProvider()
	tracer := tp.Tracer("test")

	// Configuração do request para diferentes mercados
	tests := []struct {
		name            string
		market          string
		tenantID        string
		scope           string
		operation       string
		expectedMetadata []string
	}{
		{
			name:      "Angola - Metadados específicos BNA",
			market:    "angola",
			tenantID:  "tenant_angola_001",
			scope:     "github:write",
			operation: "push",
			expectedMetadata: []string{
				"audit_retention_years",
				"requires_dual_approval",
				"operation_timestamp",
				"git_hash",
			},
		},
		{
			name:      "Brasil - Metadados LGPD",
			market:    "brasil",
			tenantID:  "tenant_brasil_001",
			scope:     "github:write",
			operation: "push",
			expectedMetadata: []string{
				"lgpd_justification",
				"data_purpose",
				"operation_timestamp",
				"git_hash",
			},
		},
		{
			name:      "UE - Metadados GDPR",
			market:    "eu",
			tenantID:  "tenant_eu_001",
			scope:     "github:admin",
			operation: "merge",
			expectedMetadata: []string{
				"gdpr_legal_basis",
				"data_minimization",
				"operation_timestamp",
				"git_hash",
			},
		},
		{
			name:      "Moçambique - Metadados Banco de Moçambique",
			market:    "mocambique",
			tenantID:  "tenant_mocambique_001",
			scope:     "github:admin",
			operation: "merge",
			expectedMetadata: []string{
				"banco_mocambique_compliance",
				"sadc_compliance",
				"operation_timestamp",
				"git_hash",
			},
		},
		{
			name:      "China - Metadados regulatórios chineses",
			market:    "china",
			tenantID:  "tenant_china_001",
			scope:     "github:read",
			operation: "clone",
			expectedMetadata: []string{
				"china_cybersecurity_compliance",
				"data_localization",
				"operation_timestamp",
				"git_hash",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock do token
			token := &models.ElevationToken{
				ID:        "token_test",
				Scope:     tt.scope,
				UserID:    "user_test",
				TenantID:  tt.tenantID,
				Market:    tt.market,
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(1 * time.Hour),
			}
			
			// Mock do request
			request := map[string]interface{}{
				"operation":  tt.operation,
				"repository": "repo-teste",
				"user_id":    "user_test",
			}
			
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
				mockMetadata["data_purpose"] = "Desenvolvimento de software"
			case "eu":
				mockMetadata["gdpr_legal_basis"] = "Legítimo interesse"
				mockMetadata["data_minimization"] = true
			case "mocambique":
				mockMetadata["banco_mocambique_compliance"] = true
				mockMetadata["sadc_compliance"] = true
			case "china":
				mockMetadata["china_cybersecurity_compliance"] = true
				mockMetadata["data_localization"] = true
			}
			
			mockProvider.On("GetMarketComplianceMetadata", mock.Anything, tt.market, "github").Return(mockMetadata, nil)
			mockProvider.On("GetTenantComplianceRules", mock.Anything, tt.tenantID, tt.market).Return(mockRules, nil)

			// Criar hook com dependências mockadas
			observabilityService := observability.NewObservabilityService(logger, tracer)
			hook := hooks.NewGitHubHook(observabilityService, mockProvider)
			
			// Executar o teste
			ctx := context.Background()
			auditMetadata, err := hook.GenerateAuditMetadata(ctx, token, request)
			
			// Verificar resultados
			assert.NoError(t, err)
			assert.NotNil(t, auditMetadata)
			
			// Verificar campos específicos para cada mercado
			for _, field := range tt.expectedMetadata {
				assert.Contains(t, auditMetadata, field, "Campo %s ausente para mercado %s", field, tt.market)
			}
		})
	}
}