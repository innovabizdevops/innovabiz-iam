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

// TestFigmaHook_ValidateScope testa a validação de escopo do Figma Hook
func TestFigmaHook_ValidateScope(t *testing.T) {
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
			name:      "Angola - Escopo figma:view válido",
			market:    "angola",
			tenantID:  "tenant_angola_001",
			scope:     "figma:view",
			expectErr: false,
		},
		{
			name:         "Angola - Escopo inválido",
			market:       "angola",
			tenantID:     "tenant_angola_001",
			scope:        "figma:invalid",
			expectErr:    true,
			errorMessage: "escopo inválido",
		},
		{
			name:      "Brasil - Escopo figma:comment válido",
			market:    "brasil",
			tenantID:  "tenant_brasil_001",
			scope:     "figma:comment",
			expectErr: false,
		},
		{
			name:      "UE - Escopo figma:edit válido",
			market:    "eu",
			tenantID:  "tenant_eu_001",
			scope:     "figma:edit",
			expectErr: false,
		},
		{
			name:         "China - Escopo restrito para o mercado",
			market:       "china",
			tenantID:     "tenant_china_001",
			scope:        "figma:admin",
			expectErr:    true,
			errorMessage: "escopo restrito para o mercado",
		},
		{
			name:      "Moçambique - Escopo figma:comment válido",
			market:    "mocambique",
			tenantID:  "tenant_mocambique_001",
			scope:     "figma:comment",
			expectErr: false,
		},
	]

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
				mockRules["allow_view"] = true
				mockRules["allow_comment"] = true
				mockRules["allow_edit"] = true
				mockRules["allow_admin"] = false
			case "brasil":
				mockMetadata["requires_lgpd_compliance"] = true
				mockRules["allow_view"] = true
				mockRules["allow_comment"] = true
				mockRules["allow_edit"] = true
				mockRules["allow_admin"] = true
			case "eu":
				mockMetadata["requires_gdpr_compliance"] = true
				mockRules["allow_view"] = true
				mockRules["allow_comment"] = true
				mockRules["allow_edit"] = true
				mockRules["allow_admin"] = true
			case "china":
				mockMetadata["requires_local_storage"] = true
				mockRules["allow_view"] = true
				mockRules["allow_comment"] = true
				mockRules["allow_edit"] = false
				mockRules["allow_admin"] = false // Restrição específica para China
			case "mocambique":
				mockRules["allow_view"] = true
				mockRules["allow_comment"] = true
				mockRules["allow_edit"] = false
				mockRules["allow_admin"] = false
			default:
				mockRules["allow_view"] = true
				mockRules["allow_comment"] = true
			}
			
			mockProvider.On("GetMarketComplianceMetadata", mock.Anything, tt.market, "figma").Return(mockMetadata, nil)
			mockProvider.On("GetTenantComplianceRules", mock.Anything, tt.tenantID, tt.market).Return(mockRules, nil)

			// Criar hook com dependências mockadas
			observabilityService := observability.NewObservabilityService(logger, tracer)
			hook := hooks.NewFigmaHook(observabilityService, mockProvider)
			
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

// TestFigmaHook_RequiresMFA testa a validação de requisitos MFA
func TestFigmaHook_RequiresMFA(t *testing.T) {
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
			name:      "Angola - MFA forte requerido para figma:admin",
			market:    "angola",
			tenantID:  "tenant_angola_001",
			scope:     "figma:admin",
			expectMFA: true,
			mfaLevel:  "forte",
		},
		{
			name:      "Brasil - MFA forte requerido para figma:admin",
			market:    "brasil",
			tenantID:  "tenant_brasil_001",
			scope:     "figma:admin",
			expectMFA: true,
			mfaLevel:  "forte",
		},
		{
			name:      "UE - MFA médio para figma:edit",
			market:    "eu",
			tenantID:  "tenant_eu_001",
			scope:     "figma:edit",
			expectMFA: true,
			mfaLevel:  "medio",
		},
		{
			name:      "Brasil - MFA médio para figma:edit",
			market:    "brasil",
			tenantID:  "tenant_brasil_001",
			scope:     "figma:edit",
			expectMFA: true,
			mfaLevel:  "medio",
		},
		{
			name:      "Angola - Sem MFA para figma:view",
			market:    "angola",
			tenantID:  "tenant_angola_001",
			scope:     "figma:view",
			expectMFA: false,
			mfaLevel:  "",
		},
		{
			name:      "China - MFA básico para figma:comment",
			market:    "china",
			tenantID:  "tenant_china_001",
			scope:     "figma:comment",
			expectMFA: true,
			mfaLevel:  "basico",
		},
	]

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configuração do mock
			mockProvider := new(MockMetadataProvider)
			
			// Configurar comportamento esperado do mock
			mockMetadata := make(map[string]interface{})
			mockRules := make(map[string]interface{})
			
			// Configurações específicas para cada escopo e mercado
			if tt.scope == "figma:admin" {
				mockMetadata["mfa_level"] = "forte"
				mockMetadata["require_mfa"] = true
			} else if tt.scope == "figma:edit" {
				mockMetadata["mfa_level"] = "medio"
				mockMetadata["require_mfa"] = true
			} else if tt.scope == "figma:comment" && tt.market == "china" {
				mockMetadata["mfa_level"] = "basico"
				mockMetadata["require_mfa"] = true
			} else if tt.scope == "figma:view" {
				mockMetadata["require_mfa"] = false
			}
			
			mockProvider.On("GetMarketComplianceMetadata", mock.Anything, tt.market, "figma").Return(mockMetadata, nil)
			mockProvider.On("GetTenantComplianceRules", mock.Anything, tt.tenantID, tt.market).Return(mockRules, nil)

			// Criar hook com dependências mockadas
			observabilityService := observability.NewObservabilityService(logger, tracer)
			hook := hooks.NewFigmaHook(observabilityService, mockProvider)
			
			// Executar o teste
			ctx := context.Background()
			requiresMFA, mfaLevel, err := hook.RequiresMFA(ctx, tt.scope, tt.tenantID, tt.market)
			
			// Verificar resultados
			assert.NoError(t, err)
			assert.Equal(t, tt.expectMFA, requiresMFA)
			assert.Equal(t, tt.mfaLevel, mfaLevel)
		})
	}
}

// TestFigmaHook_GetApprovers testa a obtenção de aprovadores
func TestFigmaHook_GetApprovers(t *testing.T) {
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
			name:           "Angola - Aprovação dupla para figma:admin",
			market:         "angola",
			tenantID:       "tenant_angola_001",
			scope:          "figma:admin",
			userID:         "user_001",
			expectedLevels: 2,
		},
		{
			name:           "Brasil - Aprovação simples para figma:edit",
			market:         "brasil",
			tenantID:       "tenant_brasil_001",
			scope:          "figma:edit",
			userID:         "user_002",
			expectedLevels: 1,
		},
		{
			name:           "UE - Aprovação simples para figma:comment",
			market:         "eu",
			tenantID:       "tenant_eu_001",
			scope:          "figma:comment",
			userID:         "user_003",
			expectedLevels: 1,
		},
		{
			name:           "Moçambique - Sem aprovação para figma:view",
			market:         "mocambique",
			tenantID:       "tenant_mocambique_001",
			scope:          "figma:view",
			userID:         "user_004",
			expectedLevels: 0,
		},
		{
			name:           "BRICS - Aprovação dupla para figma:admin",
			market:         "brics",
			tenantID:       "tenant_brics_001",
			scope:          "figma:admin",
			userID:         "user_005",
			expectedLevels: 2,
		},
	]

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configuração do mock
			mockProvider := new(MockMetadataProvider)
			
			// Configurar comportamento esperado do mock
			mockMetadata := make(map[string]interface{})
			mockRules := make(map[string]interface{})
			
			// Configurações específicas para cada mercado e escopo
			if tt.scope == "figma:admin" {
				mockMetadata["approval_levels"] = 2
			} else if tt.scope == "figma:edit" || tt.scope == "figma:comment" {
				mockMetadata["approval_levels"] = 1
			} else {
				mockMetadata["approval_levels"] = 0
			}
			
			mockProvider.On("GetMarketComplianceMetadata", mock.Anything, tt.market, "figma").Return(mockMetadata, nil)
			mockProvider.On("GetTenantComplianceRules", mock.Anything, tt.tenantID, tt.market).Return(mockRules, nil)

			// Criar hook com dependências mockadas
			observabilityService := observability.NewObservabilityService(logger, tracer)
			hook := hooks.NewFigmaHook(observabilityService, mockProvider)
			
			// Executar o teste
			ctx := context.Background()
			approvers, err := hook.GetApprovers(ctx, tt.scope, tt.userID, tt.tenantID, tt.market)
			
			// Verificar resultados
			assert.NoError(t, err)
			
			if tt.expectedLevels == 0 {
				assert.Empty(t, approvers)
			} else {
				assert.NotNil(t, approvers)
				assert.Equal(t, tt.expectedLevels, len(approvers))
			}
		})
	}
}

// TestFigmaHook_ValidateTokenUsage testa a validação de uso de token
func TestFigmaHook_ValidateTokenUsage(t *testing.T) {
	// Configuração de teste
	logger := zap.NewNop()
	tp := trace.NewNoopTracerProvider()
	tracer := tp.Tracer("test")

	// Mock do token e request
	mockToken := &models.ElevationToken{
		ID:        "token_001",
		Scope:     "figma:comment",
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
			name:  "Angola - Uso válido de figma:comment",
			token: mockToken,
			request: map[string]interface{}{
				"action":      "comment",
				"file_id":     "figma_file_123",
				"comment":     "Ajuste o alinhamento do botão",
				"user_id":     "user_001",
				"project_id":  "project_abc",
			},
			expectErr: false,
		},
		{
			name:  "Angola - Ação proibida",
			token: mockToken,
			request: map[string]interface{}{
				"action":      "delete",
				"file_id":     "figma_file_123",
				"user_id":     "user_001",
				"project_id":  "project_abc",
			},
			expectErr: true,
		},
		{
			name:  "Angola - Usuário diferente",
			token: mockToken,
			request: map[string]interface{}{
				"action":      "comment",
				"file_id":     "figma_file_123",
				"comment":     "Ajuste o alinhamento do botão",
				"user_id":     "user_002", // Usuário diferente do token
				"project_id":  "project_abc",
			},
			expectErr: true,
		},
		{
			name:  "Angola - Projeto protegido",
			token: mockToken,
			request: map[string]interface{}{
				"action":      "comment",
				"file_id":     "figma_file_456",
				"comment":     "Ajuste o alinhamento do botão",
				"user_id":     "user_001",
				"project_id":  "project_restricted", // Projeto com proteção especial
			},
			expectErr: true,
		},
	]

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configuração do mock
			mockProvider := new(MockMetadataProvider)
			
			// Configurar comportamento esperado do mock
			mockMetadata := make(map[string]interface{})
			mockRules := make(map[string]interface{})
			
			// Adicionar regras específicas para Angola
			mockMetadata["protected_projects"] = []string{"project_restricted", "project_financial"}
			mockMetadata["blocked_actions"] = []string{"delete", "share_externally"}
			mockMetadata["audit_retention_years"] = 7
			
			mockProvider.On("GetMarketComplianceMetadata", mock.Anything, tt.token.Market, "figma").Return(mockMetadata, nil)
			mockProvider.On("GetTenantComplianceRules", mock.Anything, tt.token.TenantID, tt.token.Market).Return(mockRules, nil)

			// Criar hook com dependências mockadas
			observabilityService := observability.NewObservabilityService(logger, tracer)
			hook := hooks.NewFigmaHook(observabilityService, mockProvider)
			
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

// TestFigmaHook_GenerateAuditMetadata testa a geração de metadados de auditoria
func TestFigmaHook_GenerateAuditMetadata(t *testing.T) {
	// Configuração de teste
	logger := zap.NewNop()
	tp := trace.NewNoopTracerProvider()
	tracer := tp.Tracer("test")

	// Configuração do request para diferentes mercados
	tests := []struct {
		name             string
		market           string
		tenantID         string
		scope            string
		action           string
		expectedMetadata []string
	}{
		{
			name:     "Angola - Metadados específicos BNA",
			market:   "angola",
			tenantID: "tenant_angola_001",
			scope:    "figma:edit",
			action:   "edit",
			expectedMetadata: []string{
				"audit_retention_years",
				"requires_dual_approval",
				"operation_timestamp",
				"file_version",
				"data_classification",
			},
		},
		{
			name:     "Brasil - Metadados LGPD",
			market:   "brasil",
			tenantID: "tenant_brasil_001",
			scope:    "figma:admin",
			action:   "permissions",
			expectedMetadata: []string{
				"lgpd_justification",
				"data_purpose",
				"operation_timestamp",
				"file_version",
				"data_classification",
			},
		},
		{
			name:     "UE - Metadados GDPR",
			market:   "eu",
			tenantID: "tenant_eu_001",
			scope:    "figma:comment",
			action:   "comment",
			expectedMetadata: []string{
				"gdpr_legal_basis",
				"data_minimization",
				"operation_timestamp",
				"file_version",
			},
		},
		{
			name:     "BRICS - Metadados específicos",
			market:   "brics",
			tenantID: "tenant_brics_001",
			scope:    "figma:view",
			action:   "view",
			expectedMetadata: []string{
				"brics_compliance",
				"operation_timestamp",
				"access_purpose",
			},
		},
	]

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
				"action":     tt.action,
				"file_id":    "figma_file_test",
				"user_id":    "user_test",
				"project_id": "project_test",
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
				mockMetadata["data_classification"] = "confidencial"
			case "brasil":
				mockMetadata["lgpd_justification"] = "Execução de contrato"
				mockMetadata["data_purpose"] = "Design de produto"
				mockMetadata["data_classification"] = "interna"
			case "eu":
				mockMetadata["gdpr_legal_basis"] = "Legítimo interesse"
				mockMetadata["data_minimization"] = true
			case "brics":
				mockMetadata["brics_compliance"] = true
				mockMetadata["access_purpose"] = "Revisão de design"
			}
			
			mockProvider.On("GetMarketComplianceMetadata", mock.Anything, tt.market, "figma").Return(mockMetadata, nil)
			mockProvider.On("GetTenantComplianceRules", mock.Anything, tt.tenantID, tt.market).Return(mockRules, nil)

			// Criar hook com dependências mockadas
			observabilityService := observability.NewObservabilityService(logger, tracer)
			hook := hooks.NewFigmaHook(observabilityService, mockProvider)
			
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
			
			// Todos os metadados devem conter marcação de timestamp
			assert.Contains(t, auditMetadata, "operation_timestamp")
			
			// Adicionar versão do arquivo para rastreabilidade (exceto para view)
			if tt.action != "view" {
				assert.Contains(t, auditMetadata, "file_version")
			}
		})
	}
}