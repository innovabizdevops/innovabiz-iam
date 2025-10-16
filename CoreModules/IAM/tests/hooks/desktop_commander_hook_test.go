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

// TestDesktopCommanderHook_ValidateScope testa a validação de escopo do Desktop Commander Hook
func TestDesktopCommanderHook_ValidateScope(t *testing.T) {
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
			name:      "Angola - Escopo desktop:read válido",
			market:    "angola",
			tenantID:  "tenant_angola_001",
			scope:     "desktop:read",
			expectErr: false,
		},
		{
			name:         "Angola - Escopo inválido",
			market:       "angola",
			tenantID:     "tenant_angola_001",
			scope:        "desktop:invalid",
			expectErr:    true,
			errorMessage: "escopo inválido",
		},
		{
			name:      "Brasil - Escopo desktop:write válido",
			market:    "brasil",
			tenantID:  "tenant_brasil_001",
			scope:     "desktop:write",
			expectErr: false,
		},
		{
			name:      "UE - Escopo desktop:admin válido",
			market:    "eu",
			tenantID:  "tenant_eu_001",
			scope:     "desktop:admin",
			expectErr: false,
		},
		{
			name:         "China - Escopo restrito para o mercado",
			market:       "china",
			tenantID:     "tenant_china_001",
			scope:        "desktop:system",
			expectErr:    true,
			errorMessage: "escopo restrito para o mercado",
		},
		{
			name:      "Moçambique - Escopo desktop:read válido",
			market:    "mocambique",
			tenantID:  "tenant_mocambique_001",
			scope:     "desktop:read",
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
				mockRules["allow_read"] = true
				mockRules["allow_write"] = true
				mockRules["allow_admin"] = true
				mockRules["allow_system"] = false
			case "brasil":
				mockMetadata["requires_lgpd_compliance"] = true
				mockRules["allow_read"] = true
				mockRules["allow_write"] = true
				mockRules["allow_admin"] = true
				mockRules["allow_system"] = false
			case "eu":
				mockMetadata["requires_gdpr_compliance"] = true
				mockRules["allow_read"] = true
				mockRules["allow_write"] = true
				mockRules["allow_admin"] = true
				mockRules["allow_system"] = false
			case "china":
				mockMetadata["requires_local_storage"] = true
				mockRules["allow_read"] = true
				mockRules["allow_write"] = true
				mockRules["allow_admin"] = false
				mockRules["allow_system"] = false // Restrição específica para China
			case "mocambique":
				mockRules["allow_read"] = true
				mockRules["allow_write"] = true
				mockRules["allow_admin"] = false
				mockRules["allow_system"] = false
			default:
				mockRules["allow_read"] = true
				mockRules["allow_write"] = true
			}
			
			mockProvider.On("GetMarketComplianceMetadata", mock.Anything, tt.market, "desktop").Return(mockMetadata, nil)
			mockProvider.On("GetTenantComplianceRules", mock.Anything, tt.tenantID, tt.market).Return(mockRules, nil)

			// Criar hook com dependências mockadas
			observabilityService := observability.NewObservabilityService(logger, tracer)
			hook := hooks.NewDesktopCommanderHook(observabilityService, mockProvider)
			
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

// TestDesktopCommanderHook_RequiresMFA testa a validação de requisitos MFA
func TestDesktopCommanderHook_RequiresMFA(t *testing.T) {
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
			name:      "Angola - MFA forte requerido para desktop:admin",
			market:    "angola",
			tenantID:  "tenant_angola_001",
			scope:     "desktop:admin",
			expectMFA: true,
			mfaLevel:  "forte",
		},
		{
			name:      "Brasil - MFA forte requerido para desktop:admin",
			market:    "brasil",
			tenantID:  "tenant_brasil_001",
			scope:     "desktop:admin",
			expectMFA: true,
			mfaLevel:  "forte",
		},
		{
			name:      "UE - MFA forte requerido para desktop:system",
			market:    "eu",
			tenantID:  "tenant_eu_001",
			scope:     "desktop:system",
			expectMFA: true,
			mfaLevel:  "forte",
		},
		{
			name:      "Brasil - MFA médio para desktop:write",
			market:    "brasil",
			tenantID:  "tenant_brasil_001",
			scope:     "desktop:write",
			expectMFA: true,
			mfaLevel:  "medio",
		},
		{
			name:      "Angola - Sem MFA para desktop:read",
			market:    "angola",
			tenantID:  "tenant_angola_001",
			scope:     "desktop:read",
			expectMFA: false,
			mfaLevel:  "",
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
			if tt.scope == "desktop:admin" || tt.scope == "desktop:system" {
				mockMetadata["mfa_level"] = "forte"
				mockMetadata["require_mfa"] = true
			} else if tt.scope == "desktop:write" {
				mockMetadata["mfa_level"] = "medio"
				mockMetadata["require_mfa"] = true
			} else {
				// Para desktop:read
				mockMetadata["require_mfa"] = false
			}
			
			mockProvider.On("GetMarketComplianceMetadata", mock.Anything, tt.market, "desktop").Return(mockMetadata, nil)
			mockProvider.On("GetTenantComplianceRules", mock.Anything, tt.tenantID, tt.market).Return(mockRules, nil)

			// Criar hook com dependências mockadas
			observabilityService := observability.NewObservabilityService(logger, tracer)
			hook := hooks.NewDesktopCommanderHook(observabilityService, mockProvider)
			
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

// TestDesktopCommanderHook_GetApprovers testa a obtenção de aprovadores
func TestDesktopCommanderHook_GetApprovers(t *testing.T) {
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
			name:           "Angola - Aprovação dupla para desktop:system",
			market:         "angola",
			tenantID:       "tenant_angola_001",
			scope:          "desktop:system",
			userID:         "user_001",
			expectedLevels: 2,
		},
		{
			name:           "Brasil - Aprovação dupla para desktop:admin",
			market:         "brasil",
			tenantID:       "tenant_brasil_001",
			scope:          "desktop:admin",
			userID:         "user_002",
			expectedLevels: 2,
		},
		{
			name:           "UE - Aprovação simples para desktop:write",
			market:         "eu",
			tenantID:       "tenant_eu_001",
			scope:          "desktop:write",
			userID:         "user_003",
			expectedLevels: 1,
		},
		{
			name:           "Moçambique - Sem aprovação para desktop:read",
			market:         "mocambique",
			tenantID:       "tenant_mocambique_001",
			scope:          "desktop:read",
			userID:         "user_004",
			expectedLevels: 0,
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
			if tt.scope == "desktop:system" || (tt.scope == "desktop:admin" && (tt.market == "angola" || tt.market == "brasil")) {
				mockMetadata["approval_levels"] = 2
			} else if tt.scope == "desktop:write" || (tt.scope == "desktop:admin" && tt.market != "angola" && tt.market != "brasil") {
				mockMetadata["approval_levels"] = 1
			} else {
				mockMetadata["approval_levels"] = 0
			}
			
			mockProvider.On("GetMarketComplianceMetadata", mock.Anything, tt.market, "desktop").Return(mockMetadata, nil)
			mockProvider.On("GetTenantComplianceRules", mock.Anything, tt.tenantID, tt.market).Return(mockRules, nil)

			// Criar hook com dependências mockadas
			observabilityService := observability.NewObservabilityService(logger, tracer)
			hook := hooks.NewDesktopCommanderHook(observabilityService, mockProvider)
			
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

// TestDesktopCommanderHook_ValidateTokenUsage testa a validação de uso de token
func TestDesktopCommanderHook_ValidateTokenUsage(t *testing.T) {
	// Configuração de teste
	logger := zap.NewNop()
	tp := trace.NewNoopTracerProvider()
	tracer := tp.Tracer("test")

	// Mock do token e request
	mockToken := &models.ElevationToken{
		ID:        "token_001",
		Scope:     "desktop:write",
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
			name:  "Angola - Uso válido de desktop:write",
			token: mockToken,
			request: map[string]interface{}{
				"command": "write_file",
				"path":    "/tmp/file.txt",
				"user_id": "user_001",
			},
			expectErr: false,
		},
		{
			name:  "Angola - Comando proibido",
			token: mockToken,
			request: map[string]interface{}{
				"command": "delete_directory",
				"path":    "/var/lib/system",
				"user_id": "user_001",
			},
			expectErr: true,
		},
		{
			name:  "Angola - Usuário diferente",
			token: mockToken,
			request: map[string]interface{}{
				"command": "write_file",
				"path":    "/tmp/file.txt",
				"user_id": "user_002", // Usuário diferente do token
			},
			expectErr: true,
		},
		{
			name:  "Angola - Diretório protegido",
			token: mockToken,
			request: map[string]interface{}{
				"command": "read_file",
				"path":    "/etc/secrets/api_keys.conf", // Diretório com proteção especial
				"user_id": "user_001",
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
			mockMetadata["protected_directories"] = []string{"/etc/secrets", "/var/lib/system", "/usr/local/security"}
			mockMetadata["blocked_commands"] = []string{"delete_directory", "format_disk", "change_permissions"}
			mockMetadata["audit_retention_years"] = 7
			
			mockProvider.On("GetMarketComplianceMetadata", mock.Anything, tt.token.Market, "desktop").Return(mockMetadata, nil)
			mockProvider.On("GetTenantComplianceRules", mock.Anything, tt.token.TenantID, tt.token.Market).Return(mockRules, nil)

			// Criar hook com dependências mockadas
			observabilityService := observability.NewObservabilityService(logger, tracer)
			hook := hooks.NewDesktopCommanderHook(observabilityService, mockProvider)
			
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

// TestDesktopCommanderHook_GenerateAuditMetadata testa a geração de metadados de auditoria
func TestDesktopCommanderHook_GenerateAuditMetadata(t *testing.T) {
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
		command          string
		expectedMetadata []string
	}{
		{
			name:     "Angola - Metadados específicos BNA",
			market:   "angola",
			tenantID: "tenant_angola_001",
			scope:    "desktop:write",
			command:  "write_file",
			expectedMetadata: []string{
				"audit_retention_years",
				"requires_dual_approval",
				"operation_timestamp",
				"command_hash",
				"data_classification",
			},
		},
		{
			name:     "Brasil - Metadados LGPD",
			market:   "brasil",
			tenantID: "tenant_brasil_001",
			scope:    "desktop:admin",
			command:  "create_directory",
			expectedMetadata: []string{
				"lgpd_justification",
				"data_purpose",
				"operation_timestamp",
				"command_hash",
				"data_classification",
			},
		},
		{
			name:     "UE - Metadados GDPR",
			market:   "eu",
			tenantID: "tenant_eu_001",
			scope:    "desktop:system",
			command:  "edit_file",
			expectedMetadata: []string{
				"gdpr_legal_basis",
				"data_minimization",
				"operation_timestamp",
				"command_hash",
				"data_classification",
			},
		},
		{
			name:     "China - Metadados regulatórios chineses",
			market:   "china",
			tenantID: "tenant_china_001",
			scope:    "desktop:read",
			command:  "read_file",
			expectedMetadata: []string{
				"china_cybersecurity_compliance",
				"data_localization",
				"operation_timestamp",
				"command_hash",
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
				"command": tt.command,
				"path":    "/tmp/test.txt",
				"user_id": "user_test",
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
				mockMetadata["data_purpose"] = "Operação técnica"
				mockMetadata["data_classification"] = "sensível"
			case "eu":
				mockMetadata["gdpr_legal_basis"] = "Legítimo interesse"
				mockMetadata["data_minimization"] = true
				mockMetadata["data_classification"] = "pessoal"
			case "china":
				mockMetadata["china_cybersecurity_compliance"] = true
				mockMetadata["data_localization"] = true
			}
			
			mockProvider.On("GetMarketComplianceMetadata", mock.Anything, tt.market, "desktop").Return(mockMetadata, nil)
			mockProvider.On("GetTenantComplianceRules", mock.Anything, tt.tenantID, tt.market).Return(mockRules, nil)

			// Criar hook com dependências mockadas
			observabilityService := observability.NewObservabilityService(logger, tracer)
			hook := hooks.NewDesktopCommanderHook(observabilityService, mockProvider)
			
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
		})
	}
}