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

// MockAuthorizer é um mock do autorizador base para testes unitários
type MockAuthorizer struct {
	mock.Mock
}

func (m *MockAuthorizer) Authorize(ctx context.Context, request *authorization.AuthorizationRequest) (bool, string, error) {
	args := m.Called(ctx, request)
	return args.Bool(0), args.String(1), args.Error(2)
}

// MockMFAProvider é um mock do provedor de MFA para testes unitários
type MockMFAProvider struct {
	mock.Mock
}

func (m *MockMFAProvider) RequireMFA(ctx context.Context, userID string, resourceID string, riskScore float64) (bool, error) {
	args := m.Called(ctx, userID, resourceID, riskScore)
	return args.Bool(0), args.Error(1)
}

func (m *MockMFAProvider) VerifyMFA(ctx context.Context, userID string, mfaToken string) (bool, error) {
	args := m.Called(ctx, userID, mfaToken)
	return args.Bool(0), args.Error(1)
}

// MockAuditLogger é um mock do logger de auditoria para testes unitários
type MockAuditLogger struct {
	mock.Mock
}

func (m *MockAuditLogger) LogAuthorizationDecision(ctx context.Context, request *authorization.AuthorizationRequest, allowed bool, reason string) error {
	args := m.Called(ctx, request, allowed, reason)
	return args.Error(0)
}

func (m *MockAuditLogger) LogSecurityEvent(ctx context.Context, eventType string, details map[string]interface{}) error {
	args := m.Called(ctx, eventType, details)
	return args.Error(0)
}

// TestDockerAuthzHooks_PreExecuteHook verifica o comportamento do hook de pré-execução
// para comandos Docker e Kubernetes com diferentes níveis de sensibilidade
func TestDockerAuthzHooks_PreExecuteHook(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("docker_hooks_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Casos de teste para diferentes operações Docker/K8s
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
			name: "Lista de pods - Operação segura",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp0_kubectl_get",
				Args: map[string]interface{}{
					"resourceType": "pods",
					"namespace":    "default",
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:operator:123",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"operator"},
				RiskScore:    0.2,
				RequestTime:  time.Now(),
			},
			baseAllow:      true,
			baseReason:     "Política padrão permite acesso de leitura",
			requireMFA:     false,
			expectAllow:    true,
			expectReason:   "Política padrão permite acesso de leitura",
			expectMFACheck: false,
			expectAuditLog: true,
		},
		{
			name: "Exclusão de deployment - Operação sensível com MFA",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp0_kubectl_delete",
				Args: map[string]interface{}{
					"resourceType": "deployment",
					"name":         "payment-service",
					"namespace":    "production",
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:admin:456",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"admin", "devops"},
				RiskScore:    0.6,
				RequestTime:  time.Now(),
				MFAToken:     "valid-mfa-token-123",
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
			name: "Exclusão de pod de produção - Alto risco sem MFA",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp0_kubectl_delete",
				Args: map[string]interface{}{
					"resourceType": "pod",
					"name":         "database-primary",
					"namespace":    "production",
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:developer:789",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"developer"},
				RiskScore:    0.8,
				RequestTime:  time.Now(),
				// Sem MFA Token
			},
			baseAllow:      true,
			baseReason:     "Política permite acesso com MFA",
			requireMFA:     true,
			mfaValid:       false, // MFA não fornecido
			expectAllow:    false,
			expectReason:   "MFA obrigatório para esta operação sensível",
			expectMFACheck: true,
			expectAuditLog: true,
		},
		{
			name: "Execução de comando em pod - Verificação de acesso específica",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp0_exec_in_pod",
				Args: map[string]interface{}{
					"name":      "database-primary",
					"namespace": "production",
					"command":   "psql -U postgres",
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:dba:101",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"dba", "operator"},
				RiskScore:    0.5,
				RequestTime:  time.Now(),
				MFAToken:     "valid-mfa-token-101",
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
			}			// Criar o hook de autorização Docker com os mocks
			dockerHook := hooks.NewDockerAuthzHooks(mockAuthorizer, mockMFAProvider, mockAuditLogger)
			
			// Executar o hook de pré-execução
			allowed, reason, err := dockerHook.PreExecuteHook(testCtx, tc.securityCtx, tc.toolRequest)
			
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

// TestDockerAuthzHooks_PostExecuteHook verifica o comportamento do hook de pós-execução
// que registra eventos de auditoria após a execução dos comandos
func TestDockerAuthzHooks_PostExecuteHook(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("docker_hooks_post_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Casos de teste para diferentes resultados de execução
	testCases := []struct {
		name           string
		toolRequest    *hooks.ToolRequest
		securityCtx    *authorization.SecurityContext
		executionResult *hooks.ExecutionResult
		expectAuditLog bool
	}{
		{
			name: "Execução bem-sucedida de comando de listagem",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp0_kubectl_get",
				Args: map[string]interface{}{
					"resourceType": "pods",
					"namespace":    "default",
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:operator:123",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"operator"},
				RiskScore:    0.2,
				RequestTime:  time.Now(),
			},
			executionResult: &hooks.ExecutionResult{
				Success:     true,
				StatusCode:  200,
				Message:     "Operação concluída com sucesso",
				ExecutionID: "exec-123",
				Duration:    time.Millisecond * 150,
			},
			expectAuditLog: true,
		},
		{
			name: "Falha na execução de comando de exclusão",
			toolRequest: &hooks.ToolRequest{
				ToolName: "mcp0_kubectl_delete",
				Args: map[string]interface{}{
					"resourceType": "deployment",
					"name":         "payment-service",
					"namespace":    "production",
				},
			},
			securityCtx: &authorization.SecurityContext{
				UserID:       "user:developer:789",
				TenantID:     "tenant_angola_1",
				Roles:        []string{"developer"},
				RiskScore:    0.6,
				RequestTime:  time.Now(),
				MFAToken:     "valid-mfa-token-789",
			},
			executionResult: &hooks.ExecutionResult{
				Success:     false,
				StatusCode:  403,
				Message:     "Permissão negada: insufficient permissions",
				ExecutionID: "exec-456",
				Duration:    time.Millisecond * 75,
				Error:       hooks.ErrPermissionDenied,
			},
			expectAuditLog: true,
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
			
			if tc.expectAuditLog {
				mockAuditLogger.On("LogSecurityEvent", mock.Anything, "docker_command_execution",
					mock.MatchedBy(func(details map[string]interface{}) bool {
						// Verificar se os detalhes incluem as informações esperadas
						_, hasToolName := details["tool_name"]
						_, hasUserID := details["user_id"]
						_, hasSuccess := details["success"]
						_, hasDuration := details["duration_ms"]
						
						return hasToolName && hasUserID && hasSuccess && hasDuration
					})).Return(nil)
			}
			
			// Criar o hook de autorização Docker com os mocks
			dockerHook := hooks.NewDockerAuthzHooks(mockAuthorizer, mockMFAProvider, mockAuditLogger)
			
			// Executar o hook de pós-execução
			err = dockerHook.PostExecuteHook(testCtx, tc.securityCtx, tc.toolRequest, tc.executionResult)
			
			// Verificar resultados
			require.NoError(t, err, "Execução do hook pós-execução não deveria falhar")
			
			// Verificar chamadas aos mocks
			mockAuditLogger.AssertExpectations(t)
			
			// Registrar conclusão do teste
			obs.RecordTestEnd(testCtx, tc.name, err == nil, time.Since(time.Now()))
		})
	}
}

// TestDockerAuthzHooks_ElevatedPrivileges testa cenários com elevação temporária de privilégios
// Verifica se o hook permite acesso temporário a operações sensíveis em situações de emergência
func TestDockerAuthzHooks_ElevatedPrivileges(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("docker_hooks_elevation_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Cenário de emergência com elevação de privilégios
	securityCtxWithElevation := &authorization.SecurityContext{
		UserID:      "user:operator:123",
		TenantID:    "tenant_angola_1",
		Roles:       []string{"operator"},
		RiskScore:   0.7, // Risco elevado
		RequestTime: time.Now(),
		MFAToken:    "valid-mfa-token-123",
		ElevationContext: &authorization.ElevationContext{
			ElevationID:      "elev-2025-001",
			ApprovedBy:       "manager:ops:456",
			Justification:    "Incidente de produção #INC-2025-42",
			ExpirationTime:   time.Now().Add(30 * time.Minute),
			ElevatedRoles:    []string{"admin"},
			ElevatedScopes:   []string{"k8s:production:pods:delete"},
			ApprovalTime:     time.Now().Add(-5 * time.Minute),
			ApprovalEvidence: "ticket:INC-2025-42",
		},
	}
	
	// Comando sensível que normalmente seria negado
	sensitiveRequest := &hooks.ToolRequest{
		ToolName: "mcp0_kubectl_delete",
		Args: map[string]interface{}{
			"resourceType":  "pod",
			"name":          "critical-service-pod",
			"namespace":     "production",
			"force":         true,
			"gracePeriodSeconds": 0,
		},
	}
	
	// Configurar mocks
	mockAuthorizer := new(MockAuthorizer)
	mockAuthorizer.On("Authorize", mock.Anything, mock.Anything).
		Return(false, "Acesso padrão negado para este nível de privilégio", nil)
		
	mockMFAProvider := new(MockMFAProvider)
	mockMFAProvider.On("RequireMFA", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(true, nil)
	mockMFAProvider.On("VerifyMFA", mock.Anything, securityCtxWithElevation.UserID, securityCtxWithElevation.MFAToken).
		Return(true, nil)
		
	mockAuditLogger := new(MockAuditLogger)
	mockAuditLogger.On("LogAuthorizationDecision", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	mockAuditLogger.On("LogSecurityEvent", mock.Anything, "privilege_elevation_used", mock.Anything).
		Return(nil)
		
	// Criar o hook de autorização Docker com os mocks
	dockerHook := hooks.NewDockerAuthzHooks(mockAuthorizer, mockMFAProvider, mockAuditLogger)
	
	// Teste principal: Verificar se a elevação de privilégios permite acesso
	t.Run("Elevação de privilégios permite acesso a operação sensível", func(t *testing.T) {
		ctx := context.Background()
		testCtx := obs.RecordTestStart(ctx, "elevated_privileges")
		
		// Executar o hook de pré-execução
		allowed, reason, err := dockerHook.PreExecuteHook(testCtx, securityCtxWithElevation, sensitiveRequest)
		
		// Verificar resultados
		require.NoError(t, err, "Execução do hook não deveria falhar")
		assert.True(t, allowed, "Elevação de privilégios deveria permitir acesso")
		assert.Contains(t, reason, "elevação temporária de privilégios", "Razão deveria mencionar elevação de privilégios")
		
		// Verificar chamadas aos mocks
		mockAuthorizer.AssertCalled(t, "Authorize", mock.Anything, mock.Anything)
		mockMFAProvider.AssertCalled(t, "RequireMFA", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		mockMFAProvider.AssertCalled(t, "VerifyMFA", mock.Anything, securityCtxWithElevation.UserID, securityCtxWithElevation.MFAToken)
		mockAuditLogger.AssertCalled(t, "LogSecurityEvent", mock.Anything, "privilege_elevation_used", mock.Anything)
		
		// Registrar conclusão do teste
		obs.RecordTestEnd(testCtx, "elevated_privileges", err == nil && allowed, time.Since(time.Now()))
	})
	
	// Teste adicional: Verificar elevação expirada
	t.Run("Elevação de privilégios expirada não permite acesso", func(t *testing.T) {
		ctx := context.Background()
		testCtx := obs.RecordTestStart(ctx, "expired_elevation")
		
		// Criar contexto com elevação expirada
		expiredElevation := &authorization.SecurityContext{
			UserID:      "user:operator:123",
			TenantID:    "tenant_angola_1",
			Roles:       []string{"operator"},
			RiskScore:   0.7,
			RequestTime: time.Now(),
			MFAToken:    "valid-mfa-token-123",
			ElevationContext: &authorization.ElevationContext{
				ElevationID:      "elev-2025-002",
				ApprovedBy:       "manager:ops:456",
				Justification:    "Incidente de produção #INC-2025-43",
				ExpirationTime:   time.Now().Add(-10 * time.Minute), // Expirado há 10 minutos
				ElevatedRoles:    []string{"admin"},
				ElevatedScopes:   []string{"k8s:production:pods:delete"},
				ApprovalTime:     time.Now().Add(-40 * time.Minute),
				ApprovalEvidence: "ticket:INC-2025-43",
			},
		}
		
		// Configurar mock do logger para este caso específico
		mockAuditLogger.On("LogSecurityEvent", mock.Anything, "elevation_expired", mock.Anything).
			Return(nil)
		
		// Executar o hook de pré-execução
		allowed, reason, err := dockerHook.PreExecuteHook(testCtx, expiredElevation, sensitiveRequest)
		
		// Verificar resultados
		require.NoError(t, err, "Execução do hook não deveria falhar")
		assert.False(t, allowed, "Elevação expirada não deveria permitir acesso")
		assert.Contains(t, reason, "expirada", "Razão deveria mencionar elevação expirada")
		
		// Registrar conclusão do teste
		obs.RecordTestEnd(testCtx, "expired_elevation", err == nil && !allowed, time.Since(time.Now()))
	})
}