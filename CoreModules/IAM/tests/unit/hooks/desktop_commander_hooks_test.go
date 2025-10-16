// Package hooks_test contém testes unitários para os hooks de autorização dos servidores MCP
package hooks_test

import (
	"context"
	"testing"
	"time"
	"path/filepath"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	
	"github.com/innovabizdevops/innovabiz-iam/authorization"
	"github.com/innovabizdevops/innovabiz-iam/authorization/hooks"
	"github.com/innovabizdevops/innovabiz-iam/authorization/policy"
	"github.com/innovabizdevops/innovabiz-iam/tests/testutil"
)

// TestDesktopCommanderAuthzHooks_PreExecuteHook verifica o comportamento do hook de pré-execução
// para comandos do Desktop Commander com diferentes níveis de sensibilidade
func TestDesktopCommanderAuthzHooks_PreExecuteHook(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("desktop_commander_hooks_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Base paths para cenários de testes
	basePaths := map[string]string{
		"angola":  "/dados/clientes/angola",
		"brasil":  "/dados/clientes/brasil",
		"eua":     "/dados/clientes/eua",
		"global":  "/dados/global",
		"publico": "/dados/publico",
	}
	
	// Casos de teste para diferentes operações do Desktop Commander
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
			name: "Leitura de arquivo público - Operação segura",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp1_read_file",
				Args: map[string]interface{}{
					"path": filepath.Join(basePaths["publico"], "relatorio_anual.pdf"),
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:analista:123",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"analista"},
				RiskScore:    0.1,
				RequestTime:  time.Now(),
				BusinessUnit: "analytics",
				Market:       "angola",
			},
			baseAllow:      true,
			baseReason:     "Política padrão permite acesso de leitura a arquivos públicos",
			requireMFA:     false,
			expectAllow:    true,
			expectReason:   "Política padrão permite acesso de leitura a arquivos públicos",
			expectMFACheck: false,
			expectAuditLog: true,
		},
		{
			name: "Escrita em diretório sensível - Operação sensível com MFA",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp1_write_file",
				Args: map[string]interface{}{
					"path":    filepath.Join(basePaths["angola"], "financeiro", "contratos", "contrato_importante.pdf"),
					"content": "CONTEÚDO DO CONTRATO",
					"mode":    "rewrite",
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:gerente:456",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"gerente_financeiro"},
				RiskScore:    0.5,
				RequestTime:  time.Now(),
				MFAToken:     "valid-mfa-token-456",
				BusinessUnit: "finance",
				Market:       "angola",
				AccessScopes: []string{"finance:contracts:write"},
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
			name: "Execução de processo potencialmente perigoso sem MFA",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp1_start_process",
				Args: map[string]interface{}{
					"command":    "rm -rf /dados/clientes/",
					"shell":      "bash",
					"timeout_ms": 5000,
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:developer:789",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"developer"},
				RiskScore:    0.8,
				RequestTime:  time.Now(),
				// Sem MFA Token
				BusinessUnit: "technology",
				Market:       "angola",
			},
			baseAllow:      false,
			baseReason:     "Comando potencialmente destrutivo requer aprovação especial e MFA",
			requireMFA:     true,
			mfaValid:       false, // MFA não fornecido
			expectAllow:    false,
			expectReason:   "Comando potencialmente destrutivo requer aprovação especial e MFA",
			expectMFACheck: true,
			expectAuditLog: true,
		},
		{
			name: "Pesquisa de código em mercado não autorizado",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp1_search_code",
				Args: map[string]interface{}{
					"path":    filepath.Join(basePaths["brasil"]),
					"pattern": "API_KEY",
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:analista:123",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"analista"},
				RiskScore:    0.3,
				RequestTime:  time.Now(),
				BusinessUnit: "analytics",
				Market:       "angola", // Usuário do mercado de Angola tentando acessar dados do Brasil
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
			name: "Configuração de segurança com elevação temporária",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp1_set_config_value",
				Args: map[string]interface{}{
					"key":   "allowedDirectories",
					"value": []string{"/dados/clientes/angola", "/dados/publico"},
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:seguranca:555",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"security_analyst"},
				RiskScore:    0.4,
				RequestTime:  time.Now(),
				MFAToken:     "valid-mfa-token-555",
				BusinessUnit: "security",
				Market:       "angola",
				ElevationContext: &authorization.ElevationContext{
					ElevationID:      "elev-2025-005",
					ApprovedBy:       "director:security:777",
					Justification:    "Correção de configuração de segurança #SEC-2025-42",
					ExpirationTime:   time.Now().Add(60 * time.Minute),
					ElevatedRoles:    []string{"security_admin"},
					ElevatedScopes:   []string{"system:config:write"},
					ApprovalTime:     time.Now().Add(-10 * time.Minute),
					ApprovalEvidence: "ticket:SEC-2025-42",
				},
			},
			baseAllow:      false, // Sem elevação seria negado
			baseReason:     "Configuração de sistema requer privilégios elevados",
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
			
			// Criar o hook de autorização Desktop Commander com os mocks
			dcHook := hooks.NewDesktopCommanderAuthzHooks(mockAuthorizer, mockMFAProvider, mockAuditLogger)
			
			// Executar o hook de pré-execução
			allowed, reason, err := dcHook.PreExecuteHook(testCtx, tc.securityCtx, tc.toolRequest)
			
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

// TestDesktopCommanderAuthzHooks_IsolamentoDeTenant testa o isolamento multi-tenant
// para operações de acesso a arquivos e sistemas
func TestDesktopCommanderAuthzHooks_IsolamentoDeTenant(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("desktop_commander_tenant_isolation_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Paths para teste de isolamento
	tenantPaths := map[string]string{
		"tenant_angola_1": "/dados/tenants/tenant_angola_1",
		"tenant_angola_2": "/dados/tenants/tenant_angola_2",
		"tenant_brasil_1": "/dados/tenants/tenant_brasil_1",
		"tenant_global":   "/dados/tenants/global",
	}
	
	// Casos de teste para isolamento de tenant
	testCases := []struct {
		name           string
		toolRequest    *hooks.ToolRequest
		securityCtx    *authorization.SecurityContext
		expectAllow    bool
		expectReason   string
	}{
		{
			name: "Acesso ao próprio tenant - Permitido",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp1_read_file",
				Args: map[string]interface{}{
					"path": filepath.Join(tenantPaths["tenant_angola_1"], "config.json"),
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:admin:123",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"admin"},
				RiskScore:    0.2,
				RequestTime:  time.Now(),
				BusinessUnit: "technology",
				Market:       "angola",
			},
			expectAllow:    true,
			expectReason:   "Acesso permitido ao próprio tenant",
		},
		{
			name: "Acesso cross-tenant - Negado",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp1_read_file",
				Args: map[string]interface{}{
					"path": filepath.Join(tenantPaths["tenant_angola_2"], "users.db"),
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:admin:123",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"admin"},
				RiskScore:    0.2,
				RequestTime:  time.Now(),
				BusinessUnit: "technology",
				Market:       "angola",
			},
			expectAllow:    false,
			expectReason:   "Isolamento multi-tenant: acesso cross-tenant negado",
		},
		{
			name: "Acesso a tenant global com função global - Permitido",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp1_read_file",
				Args: map[string]interface{}{
					"path": filepath.Join(tenantPaths["tenant_global"], "global_config.json"),
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:global_admin:789",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"admin", "global_admin"},
				RiskScore:    0.3,
				RequestTime:  time.Now(),
				BusinessUnit: "executive",
				Market:       "global",
				AccessScopes: []string{"global:config:read"},
			},
			expectAllow:    true,
			expectReason:   "Acesso global permitido",
		},
		{
			name: "Acesso cross-market - Negado",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp1_read_file",
				Args: map[string]interface{}{
					"path": filepath.Join(tenantPaths["tenant_brasil_1"], "market_data.json"),
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:admin:123",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"admin"},
				RiskScore:    0.4,
				RequestTime:  time.Now(),
				BusinessUnit: "technology",
				Market:       "angola",
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
			
			// Configurar comportamento específico para isolamento
			mockAuthorizer.On("Authorize", mock.Anything, mock.MatchedBy(func(req *authorization.AuthorizationRequest) bool {
				// Verificar se contém a policy de tenant isolation
				return req.Policies != nil && len(req.Policies) > 0
			})).Return(tc.expectAllow, tc.expectReason, nil)
			
			mockAuditLogger.On("LogAuthorizationDecision", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(nil)
				
			// Criar o hook de autorização Desktop Commander com os mocks
			dcHook := hooks.NewDesktopCommanderAuthzHooks(mockAuthorizer, mockMFAProvider, mockAuditLogger)
			
			// Adicionar políticas de isolamento de tenant
			dcHook.AddTenantIsolationPolicy(&policy.TenantIsolationPolicy{
				TenantPaths:   tenantPaths,
				GlobalRoles:   []string{"global_admin", "super_admin"},
				GlobalScopes:  []string{"global:config:read", "global:config:write"},
				CrossMarketAllowed: false,
			})
			
			// Executar o hook de pré-execução
			allowed, reason, err := dcHook.PreExecuteHook(testCtx, tc.securityCtx, tc.toolRequest)
			
			// Verificar resultados
			require.NoError(t, err, "Execução do hook não deveria falhar")
			assert.Equal(t, tc.expectAllow, allowed, "Resultado de autorização incorreto")
			assert.Contains(t, reason, tc.expectReason, "Razão de autorização incorreta")
			
			// Verificar chamadas aos mocks
			mockAuthorizer.AssertExpectations(t)
			mockAuditLogger.AssertExpectations(t)
			
			// Registrar conclusão do teste
			obs.RecordTestEnd(testCtx, tc.name, err == nil && allowed == tc.expectAllow, time.Since(time.Now()))
		})
	}
}

// TestDesktopCommanderAuthzHooks_SegurancaArquivos testa a segurança específica para
// operações de arquivos sensíveis
func TestDesktopCommanderAuthzHooks_SegurancaArquivos(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("desktop_commander_file_security_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Casos de teste para segurança de arquivos
	testCases := []struct {
		name           string
		toolRequest    *hooks.ToolRequest
		securityCtx    *authorization.SecurityContext
		expectAllow    bool
		expectReason   string
		expectSLALevel string
	}{
		{
			name: "Acesso a arquivo de configuração - Alta sensibilidade",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp1_write_file",
				Args: map[string]interface{}{
					"path": "/etc/innovabiz/security/encryption_keys.conf",
					"content": "nova_chave=abc123",
					"mode": "append",
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:security_admin:101",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"security_admin"},
				RiskScore:    0.7,
				RequestTime:  time.Now(),
				MFAToken:     "valid-mfa-token-101",
				BusinessUnit: "security",
				Market:       "angola",
				AccessScopes: []string{"security:encryption:manage"},
			},
			expectAllow:    false, // Mesmo com MFA, é negado por alto risco
			expectReason:   "Arquivo de alta sensibilidade requer aprovação especial",
			expectSLALevel: "critical",
		},
		{
			name: "Acesso a arquivo de logs - Média sensibilidade",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp1_read_file",
				Args: map[string]interface{}{
					"path": "/var/log/innovabiz/security/auth.log",
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:security_analyst:202",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"security_analyst"},
				RiskScore:    0.4,
				RequestTime:  time.Now(),
				MFAToken:     "valid-mfa-token-202",
				BusinessUnit: "security",
				Market:       "angola",
				AccessScopes: []string{"security:logs:read"},
			},
			expectAllow:    true,
			expectReason:   "Acesso permitido a logs de segurança",
			expectSLALevel: "high",
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
			
			// Configurar comportamento específico para segurança de arquivos
			mockAuthorizer.On("Authorize", mock.Anything, mock.MatchedBy(func(req *authorization.AuthorizationRequest) bool {
				// Verificar se é uma operação de arquivo
				toolName, ok := req.Resource.Attributes["tool_name"].(string)
				return ok && (toolName == "mcp1_read_file" || toolName == "mcp1_write_file")
			})).Return(tc.expectAllow, tc.expectReason, nil)
			
			mockMFAProvider.On("RequireMFA", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(true, nil)
			mockMFAProvider.On("VerifyMFA", mock.Anything, mock.Anything, mock.Anything).
				Return(true, nil)
				
			mockAuditLogger.On("LogAuthorizationDecision", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(nil)
			mockAuditLogger.On("LogSecurityEvent", mock.Anything, "file_access_attempt", mock.MatchedBy(func(details map[string]interface{}) bool {
				// Verificar se os detalhes de auditoria incluem SLA level
				slaLevel, ok := details["sla_level"].(string)
				return ok && slaLevel == tc.expectSLALevel
			})).Return(nil)
				
			// Criar o hook de autorização Desktop Commander com os mocks
			dcHook := hooks.NewDesktopCommanderAuthzHooks(mockAuthorizer, mockMFAProvider, mockAuditLogger)
			
			// Configurar sensibilidade de arquivos
			dcHook.ConfigureFileSensitivity(map[string]hooks.FileSensitivityConfig{
				"/etc/innovabiz/security/": {
					Level:         hooks.SensitivityCritical,
					RequiresMFA:   true,
					RequiresApproval: true,
					SLALevel:      "critical",
				},
				"/var/log/innovabiz/security/": {
					Level:         hooks.SensitivityHigh,
					RequiresMFA:   true,
					RequiresApproval: false,
					SLALevel:      "high",
				},
			})
			
			// Executar o hook de pré-execução
			allowed, reason, err := dcHook.PreExecuteHook(testCtx, tc.securityCtx, tc.toolRequest)
			
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