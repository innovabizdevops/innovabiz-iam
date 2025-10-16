// Package authorization_test contém testes unitários para o componente de autorização MCP-IAM
package authorization_test

import (
	"context"
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	
	"github.com/innovabizdevops/innovabiz-iam/authorization"
	"github.com/innovabizdevops/innovabiz-iam/authorization/policy"
	"github.com/innovabizdevops/innovabiz-iam/tests/testutil"
)

// MockPolicyEngine é um mock da engine de políticas para testes unitários
type MockPolicyEngine struct {
	mock.Mock
}

func (m *MockPolicyEngine) EvaluatePolicy(ctx context.Context, request *authorization.AuthorizationRequest) (*policy.Decision, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*policy.Decision), args.Error(1)
}

func (m *MockPolicyEngine) GetPolicyVersion() string {
	args := m.Called()
	return args.String(0)
}

// TestCoreAuthorizer_Authorize verifica se o autorizador principal funciona corretamente
// Implementa testes para diferentes cenários de autorização, incluindo casos de permitir,
// negar, elevar privilégios e delegação de acesso
func TestCoreAuthorizer_Authorize(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("core_authorizer_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Gerador de dados de teste para múltiplos cenários
	dataGen := testutil.NewAuthorizationTestDataGenerator()
	
	// Casos de teste
	testCases := []struct {
		name           string
		request        *authorization.AuthorizationRequest
		policyDecision *policy.Decision
		policyError    error
		expectAllow    bool
		expectReason   string
		expectError    bool
	}{
		{
			name: "Acesso permitido simples",
			request: &authorization.AuthorizationRequest{
				Subject:  "user:admin:test",
				Resource: "desktop_commander:read_file",
				Action:   "read",
				Context:  "normal",
				Tenant:   "tenant_angola_1",
				Environment: &authorization.RequestEnvironment{
					RiskScore: 0.2, // Baixo risco
				},
			},
			policyDecision: &policy.Decision{
				Allow:     true,
				Reason:    "Política base permite acesso",
				PolicyID:  "base_read_policy",
				Timestamp: time.Now(),
			},
			policyError:  nil,
			expectAllow:  true,
			expectReason: "Política base permite acesso",
			expectError:  false,
		},
		{
			name: "Acesso negado por política",
			request: &authorization.AuthorizationRequest{
				Subject:  "user:guest:test",
				Resource: "desktop_commander:write_file",
				Action:   "write",
				Context:  "normal",
				Tenant:   "tenant_angola_1",
				Environment: &authorization.RequestEnvironment{
					RiskScore: 0.3, // Risco médio-baixo
				},
			},
			policyDecision: &policy.Decision{
				Allow:     false,
				Reason:    "Usuário não tem permissão para escrita",
				PolicyID:  "base_write_restriction",
				Timestamp: time.Now(),
			},
			policyError:  nil,
			expectAllow:  false,
			expectReason: "Usuário não tem permissão para escrita",
			expectError:  false,
		},
		{
			name: "Acesso elevado por exceção",
			request: &authorization.AuthorizationRequest{
				Subject:  "user:operator:test",
				Resource: "mcp_docker:kubectl:pod",
				Action:   "delete",
				Context:  "emergency",
				Tenant:   "tenant_angola_1",
				Environment: &authorization.RequestEnvironment{
					RiskScore:  0.6, // Risco médio-alto
					EmergencyID: "INC-2025-001",
				},
				Attributes: map[string]interface{}{
					"justification": "Correção de incidente de produção",
					"approved_by":   "manager:operations",
					"ticket_id":     "TICKET-2025-567",
				},
			},
			policyDecision: &policy.Decision{
				Allow:     true,
				Reason:    "Acesso de emergência aprovado com elevação temporária",
				PolicyID:  "emergency_elevation_policy",
				Timestamp: time.Now(),
				Conditions: []*policy.Condition{
					{
						Type:       policy.ConditionTypeTimeLimit,
						Parameters: map[string]interface{}{"duration_minutes": 60},
					},
					{
						Type:       policy.ConditionTypeMFA,
						Parameters: map[string]interface{}{"required": true},
					},
					{
						Type:       policy.ConditionTypeAudit,
						Parameters: map[string]interface{}{"level": "detailed"},
					},
				},
			},
			policyError:  nil,
			expectAllow:  true,
			expectReason: "Acesso de emergência aprovado com elevação temporária",
			expectError:  false,
		},
		{
			name: "Erro na avaliação de política",
			request: &authorization.AuthorizationRequest{
				Subject:  "user:operator:test",
				Resource: "risk_engine:rules",
				Action:   "write",
				Context:  "normal",
				Tenant:   "tenant_angola_1",
				Environment: &authorization.RequestEnvironment{
					RiskScore: 0.5, // Risco médio
				},
			},
			policyDecision: nil,
			policyError:    authorization.ErrPolicyEvaluationFailed,
			expectAllow:    false,
			expectReason:   "Falha na avaliação de política",
			expectError:    true,
		},
		{
			name: "Acesso com delegação",
			request: &authorization.AuthorizationRequest{
				Subject:  "user:analyst:test",
				Resource: "payment_gateway:transaction:approve",
				Action:   "approve",
				Context:  "normal",
				Tenant:   "tenant_angola_1",
				Environment: &authorization.RequestEnvironment{
					RiskScore: 0.4, // Risco médio
				},
				Attributes: map[string]interface{}{
					"delegated_by":        "manager:finance",
					"delegation_id":       "DEL-2025-042",
					"delegation_expiry":   time.Now().Add(24 * time.Hour),
					"delegation_approved": true,
				},
			},
			policyDecision: &policy.Decision{
				Allow:     true,
				Reason:    "Acesso permitido por delegação aprovada",
				PolicyID:  "delegation_policy",
				Timestamp: time.Now(),
				Conditions: []*policy.Condition{
					{
						Type:       policy.ConditionTypeAudit,
						Parameters: map[string]interface{}{"level": "detailed"},
					},
					{
						Type:       policy.ConditionTypeNotify,
						Parameters: map[string]interface{}{"notify": "manager:finance"},
					},
				},
			},
			policyError:  nil,
			expectAllow:  true,
			expectReason: "Acesso permitido por delegação aprovada",
			expectError:  false,
		},
		{
			name: "Acesso multi-tenant isolado",
			request: &authorization.AuthorizationRequest{
				Subject:  "user:admin:tenant_brasil_1",
				Resource: "crm:customer:tenant_angola_1",
				Action:   "read",
				Context:  "normal",
				Tenant:   "tenant_brasil_1",
				Environment: &authorization.RequestEnvironment{
					RiskScore: 0.3,
				},
			},
			policyDecision: &policy.Decision{
				Allow:     false,
				Reason:    "Isolamento multi-tenant impede acesso entre tenants",
				PolicyID:  "tenant_isolation_policy",
				Timestamp: time.Now(),
			},
			policyError:  nil,
			expectAllow:  false,
			expectReason: "Isolamento multi-tenant impede acesso entre tenants",
			expectError:  false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, tc.name)
			startTime := time.Now()
			
			// Configurar mock da policy engine
			mockPolicyEngine := new(MockPolicyEngine)
			mockPolicyEngine.On("EvaluatePolicy", mock.Anything, tc.request).
				Return(tc.policyDecision, tc.policyError)
			
			// Criar o autorizador com a policy engine mockada
			authorizer := authorization.NewCoreAuthorizer(mockPolicyEngine)
			
			// Executar a função sendo testada
			allow, reason, err := authorizer.Authorize(testCtx, tc.request)
			
			// Verificar resultados
			if tc.expectError {
				assert.Error(t, err, "Deveria retornar um erro")
			} else {
				assert.NoError(t, err, "Não deveria retornar erro")
			}
			
			assert.Equal(t, tc.expectAllow, allow, "Resultado de autorização incorreto")
			assert.Contains(t, reason, tc.expectReason, "Razão de autorização incorreta")
			
			// Verificar se o mock foi chamado como esperado
			mockPolicyEngine.AssertExpectations(t)
			
			// Registrar conclusão do teste
			duration := time.Since(startTime)
			obs.RecordTestEnd(testCtx, tc.name, err == nil && allow == tc.expectAllow, duration)
		})
	}
}

// TestCoreAuthorizer_WithRiskEvaluation testa o componente de avaliação de risco
// Verifica se o autorizador ajusta decisões com base em fatores de risco
// e implementa controles adaptativos
func TestCoreAuthorizer_WithRiskEvaluation(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("core_authorizer_risk_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Criar gerador de dados de teste
	dataGen := testutil.NewAuthorizationTestDataGenerator()
	
	// Gerar conjunto de requisições com diferentes níveis de risco
	testRequests := []*authorization.AuthorizationRequest{
		// Requisição de baixo risco
		{
			Subject:  "user:operator:test",
			Resource: "mobile_money:account:view",
			Action:   "read",
			Context:  "normal",
			Tenant:   "tenant_angola_1",
			Environment: &authorization.RequestEnvironment{
				RiskScore:          0.15, // Risco muito baixo
				IPAddress:          "192.168.1.100", // IP interno
				IsRecognizedDevice: true,
				AccessTime:         time.Date(2025, 8, 5, 14, 30, 0, 0, time.UTC), // Horário comercial
				Location:           "Angola",
			},
		},
		// Requisição de risco médio
		{
			Subject:  "user:operator:test",
			Resource: "mobile_money:account:transfer",
			Action:   "write",
			Context:  "normal",
			Tenant:   "tenant_angola_1",
			Environment: &authorization.RequestEnvironment{
				RiskScore:          0.45, // Risco médio
				IPAddress:          "203.0.113.5", // IP externo
				IsRecognizedDevice: true,
				AccessTime:         time.Date(2025, 8, 5, 17, 45, 0, 0, time.UTC), // Final do expediente
				Location:           "Angola",
			},
			Attributes: map[string]interface{}{
				"transfer_amount": 5000, // Valor médio
			},
		},
		// Requisição de alto risco
		{
			Subject:  "user:operator:test",
			Resource: "mobile_money:account:mass_transfer",
			Action:   "write",
			Context:  "normal",
			Tenant:   "tenant_angola_1",
			Environment: &authorization.RequestEnvironment{
				RiskScore:          0.85, // Risco alto
				IPAddress:          "198.51.100.42", // IP não reconhecido
				IsRecognizedDevice: false, // Dispositivo desconhecido
				AccessTime:         time.Date(2025, 8, 5, 2, 15, 0, 0, time.UTC), // Madrugada
				Location:           "China", // Local incomum
			},
			Attributes: map[string]interface{}{
				"transfer_amount": 100000, // Valor alto
				"recipients":      50,     // Muitos destinatários
			},
		},
	}
	
	// Configurar mock para diferentes respostas com base no risco
	mockPolicyEngine := new(MockPolicyEngine)
	
	// Baixo risco: permitido sem condições
	mockPolicyEngine.On("EvaluatePolicy", mock.Anything, testRequests[0]).
		Return(&policy.Decision{
			Allow:    true,
			Reason:   "Acesso de baixo risco permitido",
			PolicyID: "risk_adaptive_policy",
		}, nil)
	
	// Médio risco: permitido com MFA
	mockPolicyEngine.On("EvaluatePolicy", mock.Anything, testRequests[1]).
		Return(&policy.Decision{
			Allow:    true,
			Reason:   "Acesso de risco médio permitido com verificação adicional",
			PolicyID: "risk_adaptive_policy",
			Conditions: []*policy.Condition{
				{
					Type:       policy.ConditionTypeMFA,
					Parameters: map[string]interface{}{"required": true},
				},
			},
		}, nil)
	
	// Alto risco: negado
	mockPolicyEngine.On("EvaluatePolicy", mock.Anything, testRequests[2]).
		Return(&policy.Decision{
			Allow:    false,
			Reason:   "Acesso de alto risco bloqueado por segurança",
			PolicyID: "risk_adaptive_policy",
		}, nil)
	
	// Criar o autorizador com a policy engine mockada
	authorizer := authorization.NewCoreAuthorizer(mockPolicyEngine)
	
	// Testar requisições com diferentes níveis de risco
	t.Run("Autorização adaptativa ao risco", func(t *testing.T) {
		ctx := context.Background()
		testCtx := obs.RecordTestStart(ctx, "adaptive_risk")
		
		// Testar requisição de baixo risco
		allow, reason, err := authorizer.Authorize(testCtx, testRequests[0])
		assert.NoError(t, err, "Não deveria retornar erro para baixo risco")
		assert.True(t, allow, "Deveria permitir acesso para baixo risco")
		assert.Contains(t, reason, "baixo risco", "Razão deveria mencionar baixo risco")
		
		// Testar requisição de médio risco
		allow, reason, err = authorizer.Authorize(testCtx, testRequests[1])
		assert.NoError(t, err, "Não deveria retornar erro para médio risco")
		assert.True(t, allow, "Deveria permitir acesso para médio risco, mas com condições")
		assert.Contains(t, reason, "verificação adicional", "Razão deveria mencionar verificação adicional")
		
		// Testar requisição de alto risco
		allow, reason, err = authorizer.Authorize(testCtx, testRequests[2])
		assert.NoError(t, err, "Não deveria retornar erro para alto risco")
		assert.False(t, allow, "Não deveria permitir acesso para alto risco")
		assert.Contains(t, reason, "alto risco", "Razão deveria mencionar alto risco")
		
		obs.RecordTestEnd(testCtx, "adaptive_risk", true, time.Since(time.Now()))
	})
}

// TestCoreAuthorizer_MultiTenant testa o isolamento de tenants
// Verifica se as decisões respeitam fronteiras de tenant e implementam
// controles de isolamento e compartilhamento adequados
func TestCoreAuthorizer_MultiTenant(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("core_authorizer_multitenant_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Casos de teste para isolamento multi-tenant
	testCases := []struct {
		name           string
		subject        string
		subjectTenant  string
		resource       string
		resourceTenant string
		action         string
		expectAllow    bool
		expectReason   string
	}{
		{
			name:           "Mesmo tenant - permitido",
			subject:        "user:admin:test",
			subjectTenant:  "tenant_angola_1",
			resource:       "crm:customer:list",
			resourceTenant: "tenant_angola_1",
			action:         "read",
			expectAllow:    true,
			expectReason:   "Acesso dentro do mesmo tenant",
		},
		{
			name:           "Tenants diferentes - negado",
			subject:        "user:admin:test",
			subjectTenant:  "tenant_brasil_1",
			resource:       "crm:customer:list",
			resourceTenant: "tenant_angola_1",
			action:         "read",
			expectAllow:    false,
			expectReason:   "Acesso entre tenants diferentes não permitido",
		},
		{
			name:           "Admin global - permitido",
			subject:        "user:global_admin:test",
			subjectTenant:  "tenant_global",
			resource:       "crm:customer:list",
			resourceTenant: "tenant_angola_1",
			action:         "read",
			expectAllow:    true,
			expectReason:   "Acesso entre tenants permitido para administrador global",
		},
		{
			name:           "Recurso compartilhado - permitido",
			subject:        "user:operator:test",
			subjectTenant:  "tenant_angola_2",
			resource:       "marketplace:product:view",
			resourceTenant: "tenant_angola_1",
			action:         "read",
			expectAllow:    true,
			expectReason:   "Acesso a recurso compartilhado permitido",
		},
		{
			name:           "Escrita em recurso de outro tenant - negado",
			subject:        "user:operator:test",
			subjectTenant:  "tenant_angola_2",
			resource:       "marketplace:product:update",
			resourceTenant: "tenant_angola_1",
			action:         "write",
			expectAllow:    false,
			expectReason:   "Operação de escrita entre tenants não permitida",
		},
	}
	
	// Configurar mock para decisões multi-tenant
	mockPolicyEngine := new(MockPolicyEngine)
	
	// Configurar comportamento esperado para cada caso de teste
	for i, tc := range testCases {
		request := &authorization.AuthorizationRequest{
			Subject:  tc.subject,
			Resource: tc.resource,
			Action:   tc.action,
			Tenant:   tc.subjectTenant,
			Attributes: map[string]interface{}{
				"resource_tenant": tc.resourceTenant,
			},
			Environment: &authorization.RequestEnvironment{
				RiskScore: 0.3, // Risco médio-baixo padrão
			},
		}
		
		decision := &policy.Decision{
			Allow:    tc.expectAllow,
			Reason:   tc.expectReason,
			PolicyID: "tenant_isolation_policy",
		}
		
		mockPolicyEngine.On("EvaluatePolicy", mock.Anything, request).
			Return(decision, nil)
	}
	
	// Criar o autorizador com a policy engine mockada
	authorizer := authorization.NewCoreAuthorizer(mockPolicyEngine)
	
	// Executar os casos de teste
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, tc.name)
			
			request := &authorization.AuthorizationRequest{
				Subject:  tc.subject,
				Resource: tc.resource,
				Action:   tc.action,
				Tenant:   tc.subjectTenant,
				Attributes: map[string]interface{}{
					"resource_tenant": tc.resourceTenant,
				},
				Environment: &authorization.RequestEnvironment{
					RiskScore: 0.3, // Risco médio-baixo padrão
				},
			}
			
			allow, reason, err := authorizer.Authorize(testCtx, request)
			
			assert.NoError(t, err, "Não deveria retornar erro")
			assert.Equal(t, tc.expectAllow, allow, "Resultado de autorização incorreto para caso %s", tc.name)
			assert.Contains(t, reason, tc.expectReason, "Razão de autorização incorreta")
			
			obs.RecordTestEnd(testCtx, tc.name, err == nil && allow == tc.expectAllow, time.Since(time.Now()))
		})
	}
	
	// Verificar se o mock foi chamado como esperado
	mockPolicyEngine.AssertExpectations(t)
}