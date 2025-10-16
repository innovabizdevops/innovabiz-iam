// Package hooks_test contém testes unitários para os hooks de autorização dos servidores MCP
package hooks_test

import (
	"context"
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	
	"github.com/innovabizdevops/innovabiz-iam/authorization"
	"github.com/innovabizdevops/innovabiz-iam/authorization/hooks"
	"github.com/innovabizdevops/innovabiz-iam/authorization/policy"
	"github.com/innovabizdevops/innovabiz-iam/tests/testutil"
)

// TestGitHubAuthzHooks_PreExecuteHook verifica o comportamento do hook de pré-execução
// para operações do GitHub com diferentes níveis de sensibilidade
func TestGitHubAuthzHooks_PreExecuteHook(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("github_hooks_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Casos de teste para diferentes operações do GitHub
	testCases := []struct {
		name           string
		toolRequest    *hooks.ToolRequest
		securityCtx    *authorization.SecurityContext
		baseAllow      bool
		baseReason     string
		requireMFA     bool
		mfaValid       bool
		expectAllow    bool
		expectReason   string
		expectMFACheck bool
		expectAuditLog bool
	}{
		{
			name: "Leitura de conteúdo de arquivo público - Operação segura",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp7_get_file_contents",
				Args: map[string]interface{}{
					"owner": "innovabizdevops",
					"repo":  "innovabiz-docs-public",
					"path":  "README.md",
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:developer:123",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"developer"},
				RiskScore:    0.1,
				RequestTime:  time.Now(),
				BusinessUnit: "technology",
				Market:       "angola",
			},
			baseAllow:      true,
			baseReason:     "Política padrão permite acesso de leitura a repositório público",
			requireMFA:     false,
			expectAllow:    true,
			expectReason:   "Política padrão permite acesso de leitura a repositório público",
			expectMFACheck: false,
			expectAuditLog: true,
		},
		{
			name: "Criação de pull request em repositório de produção - Operação sensível com MFA",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp7_create_pull_request",
				Args: map[string]interface{}{
					"owner": "innovabizdevops",
					"repo":  "innovabiz-iam",
					"title": "Feature: Implementação de autenticação adaptativa",
					"head":  "feature/auth-adaptativa",
					"base":  "main",
					"body":  "Implementação do módulo de autenticação adaptativa conforme especificação.",
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:senior_dev:456",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"senior_developer", "tech_lead"},
				RiskScore:    0.4,
				RequestTime:  time.Now(),
				MFAToken:     "valid-mfa-token-456",
				BusinessUnit: "technology",
				Market:       "angola",
				AccessScopes: []string{"github:repo:write"},
			},
			baseAllow:      true,
			baseReason:     "Política permite acesso com MFA",
			requireMFA:     true,
			mfaValid:       true,
			expectAllow:    true,
			expectReason:   "Política permite acesso com MFA",
			expectMFACheck: true,
			expectAuditLog: true,
		},
		{
			name: "Merge de pull request em branch principal sem MFA",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp7_merge_pull_request",
				Args: map[string]interface{}{
					"owner":        "innovabizdevops",
					"repo":         "innovabiz-payment-gateway",
					"pull_number":  42,
					"merge_method": "squash",
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:developer:789",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"developer"},
				RiskScore:    0.6,
				RequestTime:  time.Now(),
				// Sem MFA Token
				BusinessUnit: "technology",
				Market:       "angola",
			},
			baseAllow:      false,
			baseReason:     "Operação de merge em repositório crítico requer MFA",
			requireMFA:     true,
			mfaValid:       false, // MFA não fornecido
			expectAllow:    false,
			expectReason:   "Operação de merge em repositório crítico requer MFA",
			expectMFACheck: true,
			expectAuditLog: true,
		},
		{
			name: "Acesso a repositório de outro mercado",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp7_get_file_contents",
				Args: map[string]interface{}{
					"owner": "innovabizdevops",
					"repo":  "innovabiz-iam-brasil",
					"path":  "src/config/market_specific_rules.go",
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:developer:123",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"developer"},
				RiskScore:    0.3,
				RequestTime:  time.Now(),
				BusinessUnit: "technology",
				Market:       "angola", // Usuário do mercado de Angola tentando acessar código específico do Brasil
			},
			baseAllow:      false,
			baseReason:     "Acesso cross-market negado por política de isolamento",
			requireMFA:     false,
			expectAllow:    false,
			expectReason:   "Acesso cross-market negado por política de isolamento",
			expectMFACheck: false,
			expectAuditLog: true,
		},
		{
			name: "Push para repositório sensível com elevação temporária",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp7_push_files",
				Args: map[string]interface{}{
					"owner":   "innovabizdevops",
					"repo":    "innovabiz-security",
					"branch":  "main",
					"message": "Atualização urgente de regras de segurança",
					"files": []map[string]interface{}{
						{
							"path":    "security/rules.json",
							"content": "{\"rules\": []}",
						},
					},
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:security_dev:555",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"security_developer"},
				RiskScore:    0.7,
				RequestTime:  time.Now(),
				MFAToken:     "valid-mfa-token-555",
				BusinessUnit: "security",
				Market:       "angola",
				ElevationContext: &authorization.ElevationContext{
					ElevationID:      "elev-2025-010",
					ApprovedBy:       "director:security:777",
					Justification:    "Correção urgente de vulnerabilidade #SEC-2025-55",
					ExpirationTime:   time.Now().Add(60 * time.Minute),
					ElevatedRoles:    []string{"security_admin"},
					ElevatedScopes:   []string{"github:security:write"},
					ApprovalTime:     time.Now().Add(-10 * time.Minute),
					ApprovalEvidence: "ticket:SEC-2025-55",
				},
			},
			baseAllow:      false, // Sem elevação seria negado
			baseReason:     "Push para repositório de segurança requer privilégios elevados",
			requireMFA:     true,
			mfaValid:       true,
			expectAllow:    true, // Com elevação é permitido
			expectReason:   "Acesso permitido via elevação temporária de privilégios",
			expectMFACheck: true,
			expectAuditLog: true,
		},
	}	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, tc.name)
			
			// Configurar mocks
			mockAuthorizer := new(MockAuthorizer)
			mockAuthorizer.On("Authorize", mock.Anything, mock.Anything).
				Return(tc.baseAllow, tc.baseReason, nil)
				
			mockMFAProvider := new(MockMFAProvider)
			if tc.expectMFACheck {
				mockMFAProvider.On("RequireMFA", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(tc.requireMFA, nil)
					
				if tc.requireMFA && tc.securityCtx.MFAToken != "" {
					mockMFAProvider.On("VerifyMFA", mock.Anything, tc.securityCtx.UserID, tc.securityCtx.MFAToken).
						Return(tc.mfaValid, nil)
				}
			}
			
			mockAuditLogger := new(MockAuditLogger)
			if tc.expectAuditLog {
				mockAuditLogger.On("LogAuthorizationDecision", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
			}
			
			// Criar o hook de autorização GitHub com os mocks
			githubHook := hooks.NewGitHubAuthzHooks(mockAuthorizer, mockMFAProvider, mockAuditLogger)
			
			// Executar o hook de pré-execução
			allowed, reason, err := githubHook.PreExecuteHook(testCtx, tc.securityCtx, tc.toolRequest)
			
			// Verificar resultados
			require.NoError(t, err, "Execução do hook não deveria falhar")
			assert.Equal(t, tc.expectAllow, allowed, "Resultado de autorização incorreto")
			assert.Contains(t, reason, tc.expectReason, "Razão de autorização incorreta")
			
			// Verificar chamadas aos mocks
			mockAuthorizer.AssertExpectations(t)
			mockMFAProvider.AssertExpectations(t)
			mockAuditLogger.AssertExpectations(t)
			
			// Registrar conclusão do teste
			obs.RecordTestEnd(testCtx, tc.name, err == nil && allowed == tc.expectAllow, time.Since(time.Now()))
		})
	}
}

// TestGitHubAuthzHooks_RepoSensitivityControl testa o controle de acesso baseado na sensibilidade
// de diferentes repositórios do GitHub
func TestGitHubAuthzHooks_RepoSensitivityControl(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("github_repo_sensitivity_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Mapeamento de sensibilidade de repositórios para teste
	repoSensitivity := map[string]hooks.RepoSensitivity{
		"innovabizdevops/innovabiz-iam":            hooks.RepoSensitivityCritical,
		"innovabizdevops/innovabiz-payment-gateway": hooks.RepoSensitivityCritical,
		"innovabizdevops/innovabiz-security":        hooks.RepoSensitivityCritical,
		"innovabizdevops/innovabiz-mobile-money":    hooks.RepoSensitivityHigh,
		"innovabizdevops/innovabiz-marketplace":     hooks.RepoSensitivityHigh,
		"innovabizdevops/innovabiz-ui-components":   hooks.RepoSensitivityMedium,
		"innovabizdevops/innovabiz-docs-internal":   hooks.RepoSensitivityMedium,
		"innovabizdevops/innovabiz-docs-public":     hooks.RepoSensitivityLow,
	}
	
	// Mapeamento de mercados a repositórios
	repoMarkets := map[string][]string{
		"innovabizdevops/innovabiz-iam":             {"global"}, // Multi-mercado/global
		"innovabizdevops/innovabiz-payment-gateway":  {"global"}, // Multi-mercado/global
		"innovabizdevops/innovabiz-iam-angola":       {"angola"},
		"innovabizdevops/innovabiz-iam-brasil":       {"brasil"},
		"innovabizdevops/innovabiz-marketplace-angola": {"angola"},
	}
	
	// Casos de teste para controle de sensibilidade
	testCases := []struct {
		name           string
		toolRequest    *hooks.ToolRequest
		securityCtx    *authorization.SecurityContext
		expectAllow    bool
		expectReason   string
	}{
		{
			name: "Acesso de leitura a repositório crítico - Developer com MFA",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp7_get_file_contents",
				Args: map[string]interface{}{
					"owner": "innovabizdevops",
					"repo":  "innovabiz-iam",
					"path":  "src/core/authorization.go",
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:developer:123",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"developer"},
				RiskScore:    0.3,
				RequestTime:  time.Now(),
				MFAToken:     "valid-mfa-token-123",
				BusinessUnit: "technology",
				Market:       "angola",
				AccessScopes: []string{"github:repo:read"},
			},
			expectAllow:    true,
			expectReason:   "Acesso de leitura permitido após validação de MFA",
		},
		{
			name: "Push para repositório crítico - Developer sem permissão específica",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp7_push_files",
				Args: map[string]interface{}{
					"owner":   "innovabizdevops",
					"repo":    "innovabiz-iam",
					"branch":  "feature/nova-funcionalidade",
					"message": "Implementação de nova funcionalidade",
					"files": []map[string]interface{}{
						{
							"path":    "src/core/authorization.go",
							"content": "conteúdo atualizado",
						},
					},
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:developer:123",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"developer"},
				RiskScore:    0.5,
				RequestTime:  time.Now(),
				MFAToken:     "valid-mfa-token-123",
				BusinessUnit: "technology",
				Market:       "angola",
				AccessScopes: []string{"github:repo:read"}, // Falta write
			},
			expectAllow:    false,
			expectReason:   "Escopo de acesso insuficiente para escrita em repositório crítico",
		},
		{
			name: "Merge em branch principal de repositório crítico - Tech Lead",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp7_merge_pull_request",
				Args: map[string]interface{}{
					"owner":        "innovabizdevops",
					"repo":         "innovabiz-payment-gateway",
					"pull_number":  42,
					"merge_method": "squash",
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:tech_lead:456",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"tech_lead", "senior_developer"},
				RiskScore:    0.4,
				RequestTime:  time.Now(),
				MFAToken:     "valid-mfa-token-456",
				BusinessUnit: "technology",
				Market:       "angola",
				AccessScopes: []string{"github:repo:write", "github:repo:admin"},
			},
			expectAllow:    true,
			expectReason:   "Permissão de merge concedida para tech lead após validação de MFA",
		},
		{
			name: "Acesso a repositório específico do mercado - Erro de mercado",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp7_get_file_contents",
				Args: map[string]interface{}{
					"owner": "innovabizdevops",
					"repo":  "innovabiz-iam-brasil", // Repositório específico do Brasil
					"path":  "src/config/market_specific.go",
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:developer:123",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"developer"},
				RiskScore:    0.3,
				RequestTime:  time.Now(),
				MFAToken:     "valid-mfa-token-123",
				BusinessUnit: "technology",
				Market:       "angola", // Usuário do mercado Angola tentando acessar repositório do Brasil
			},
			expectAllow:    false,
			expectReason:   "Isolamento multi-market: acesso cross-market negado",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, tc.name)
			
			// Configurar mocks
			mockAuthorizer := new(MockAuthorizer)
			mockMFAProvider := new(MockMFAProvider)
			mockAuditLogger := new(MockAuditLogger)
			
			// Configuração de resposta padrão para o Authorize
			mockAuthorizer.On("Authorize", mock.Anything, mock.Anything).
				Return(tc.expectAllow, tc.expectReason, nil)
				
			// MFA sempre requerido para repositórios críticos/altos
			mockMFAProvider.On("RequireMFA", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(true, nil)
			mockMFAProvider.On("VerifyMFA", mock.Anything, mock.Anything, mock.Anything).
				Return(true, nil)
				
			mockAuditLogger.On("LogAuthorizationDecision", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(nil)
			mockAuditLogger.On("LogSecurityEvent", mock.Anything, mock.Anything, mock.Anything).
				Return(nil)
				
			// Criar o hook de autorização GitHub com os mocks
			githubHook := hooks.NewGitHubAuthzHooks(mockAuthorizer, mockMFAProvider, mockAuditLogger)
			
			// Configurar sensibilidade dos repositórios e mapeamento de mercado
			githubHook.ConfigureRepoSensitivity(repoSensitivity)
			githubHook.ConfigureMarketRepoMapping(repoMarkets)
			
			// Executar o hook de pré-execução
			allowed, reason, err := githubHook.PreExecuteHook(testCtx, tc.securityCtx, tc.toolRequest)
			
			// Verificar resultados
			require.NoError(t, err, "Execução do hook não deveria falhar")
			assert.Equal(t, tc.expectAllow, allowed, "Resultado de autorização incorreto")
			assert.Contains(t, reason, tc.expectReason, "Razão de autorização incorreta")
			
			// Verificar chamadas aos mocks
			mockAuthorizer.AssertExpectations(t)
			mockMFAProvider.AssertExpectations(t)
			mockAuditLogger.AssertExpectations(t)
			
			// Registrar conclusão do teste
			obs.RecordTestEnd(testCtx, tc.name, err == nil && allowed == tc.expectAllow, time.Since(time.Now()))
		})
	}
}

// TestGitHubAuthzHooks_ComplianceChecks testa verificações adicionais de compliance
// relacionadas a operações específicas do GitHub
func TestGitHubAuthzHooks_ComplianceChecks(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("github_compliance_checks_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Casos de teste para verificações de compliance
	testCases := []struct {
		name           string
		toolRequest    *hooks.ToolRequest
		securityCtx    *authorization.SecurityContext
		expectAllow    bool
		expectReason   string
		expectComplianceRecord bool
		complianceType string
	}{
		{
			name: "Criação de repositório público - Verificação de compliance",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp7_create_repository",
				Args: map[string]interface{}{
					"name":        "novo-componente-open-source",
					"description": "Componente open source para integração com plataformas de pagamento",
					"private":     false,
					"autoInit":    true,
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:tech_lead:456",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"tech_lead"},
				RiskScore:    0.3,
				RequestTime:  time.Now(),
				MFAToken:     "valid-mfa-token-456",
				BusinessUnit: "technology",
				Market:       "angola",
			},
			expectAllow:    true,
			expectReason:   "Operação aprovada com registro de compliance para auditoria",
			expectComplianceRecord: true,
			complianceType: "open_source_release",
		},
		{
			name: "Fork de repositório de terceiros - Verificação de licença",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp7_fork_repository",
				Args: map[string]interface{}{
					"owner": "external-org",
					"repo":  "payment-sdk",
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:developer:123",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"developer"},
				RiskScore:    0.2,
				RequestTime:  time.Now(),
				MFAToken:     "valid-mfa-token-123",
				BusinessUnit: "technology",
				Market:       "angola",
			},
			expectAllow:    true,
			expectReason:   "Operação aprovada com verificação de licença",
			expectComplianceRecord: true,
			complianceType: "third_party_code_usage",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, tc.name)
			
			// Configurar mocks
			mockAuthorizer := new(MockAuthorizer)
			mockMFAProvider := new(MockMFAProvider)
			mockAuditLogger := new(MockAuditLogger)
			
			// Configuração de resposta padrão para o Authorize
			mockAuthorizer.On("Authorize", mock.Anything, mock.Anything).
				Return(true, "Autorização base aprovada", nil)
				
			// MFA sempre requerido para operações de compliance
			mockMFAProvider.On("RequireMFA", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(true, nil)
			mockMFAProvider.On("VerifyMFA", mock.Anything, mock.Anything, mock.Anything).
				Return(true, nil)
				
			mockAuditLogger.On("LogAuthorizationDecision", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(nil)
				
			// Verificação de registro de compliance
			if tc.expectComplianceRecord {
				mockAuditLogger.On("LogSecurityEvent", mock.Anything, tc.complianceType, mock.MatchedBy(func(details map[string]interface{}) bool {
					// Verificar se os detalhes de compliance contêm campos obrigatórios
					_, hasUserID := details["user_id"]
					_, hasTimestamp := details["timestamp"]
					_, hasMarket := details["market"]
					_, hasOperation := details["operation"]
					
					return hasUserID && hasTimestamp && hasMarket && hasOperation
				})).Return(nil)
			}
				
			// Criar o hook de autorização GitHub com os mocks
			githubHook := hooks.NewGitHubAuthzHooks(mockAuthorizer, mockMFAProvider, mockAuditLogger)
			
			// Adicionar verificador de compliance
			githubHook.RegisterComplianceChecker(&hooks.GitHubComplianceChecker{
				CheckPublicRepoCreation: true,
				CheckThirdPartyCodeUsage: true,
				RequireApprovalForPublicRepos: false, // Para o teste, não requer aprovação adicional
			})
			
			// Executar o hook de pré-execução
			allowed, reason, err := githubHook.PreExecuteHook(testCtx, tc.securityCtx, tc.toolRequest)
			
			// Verificar resultados
			require.NoError(t, err, "Execução do hook não deveria falhar")
			assert.Equal(t, tc.expectAllow, allowed, "Resultado de autorização incorreto")
			assert.Contains(t, reason, tc.expectReason, "Razão de autorização incorreta")
			
			// Verificar chamadas aos mocks
			mockAuthorizer.AssertExpectations(t)
			mockMFAProvider.AssertExpectations(t)
			mockAuditLogger.AssertExpectations(t)
			
			// Registrar conclusão do teste
			obs.RecordTestEnd(testCtx, tc.name, err == nil && allowed == tc.expectAllow, time.Since(time.Now()))
		})
	}
}