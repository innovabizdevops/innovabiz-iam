// Package elevation_test contém testes unitários para o componente de elevação de privilégios do MCP-IAM
package elevation_test

import (
	"context"
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	
	"github.com/innovabizdevops/innovabiz-iam/authorization"
	"github.com/innovabizdevops/innovabiz-iam/authorization/elevation"
	"github.com/innovabizdevops/innovabiz-iam/policy"
	"github.com/innovabizdevops/innovabiz-iam/tests/testutil"
)

// MockPolicyEngine é um mock do motor de políticas
type MockPolicyEngine struct {
	mock.Mock
}

func (m *MockPolicyEngine) EvaluatePolicy(ctx context.Context, policyID string, input map[string]interface{}) (*policy.PolicyDecision, error) {
	args := m.Called(ctx, policyID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*policy.PolicyDecision), args.Error(1)
}

// TestPrivilegeElevationPolicy testa a avaliação de políticas para elevação de privilégios
func TestPrivilegeElevationPolicy(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("privilege_elevation_policy_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Configurar mocks
	mockApprover := new(MockElevationApprover)
	mockLogger := new(MockAuditLogger)
	mockNotifier := new(MockNotifier)
	mockPolicyEngine := new(MockPolicyEngine)
	
	// Criar o gerenciador de elevação de privilégios com os mocks
	elevationManager := elevation.NewPrivilegeElevationManager(
		mockApprover,
		mockLogger,
		mockNotifier,
	)
	
	// Configurar motor de políticas
	elevationManager.ConfigurePolicyEngine(mockPolicyEngine)
	
	// Configurar IDs de política para diferentes cenários
	elevationManager.ConfigurePolicyIDs(&elevation.PolicyConfiguration{
		ElevationRequestPolicy:  "innovabiz:iam:elevation:request",
		ElevationApprovalPolicy: "innovabiz:iam:elevation:approval",
		ElevationUsagePolicy:    "innovabiz:iam:elevation:usage",
		ElevationScopePolicy:    "innovabiz:iam:elevation:scope",
	})
	
	// Casos de teste para avaliação de políticas
	testCases := []struct {
		name             string
		policyID         string
		policyInput      map[string]interface{}
		policyDecision   *policy.PolicyDecision
		policyError      error
		expectAllowed    bool
		expectReasons    []string
		expectConditions map[string]interface{}
	}{
		{
			name:     "Política permite elevação sem condições",
			policyID: "innovabiz:iam:elevation:request",
			policyInput: map[string]interface{}{
				"user_id":          "user:developer:123",
				"tenant_id":        "tenant_angola_1",
				"market":           "angola",
				"business_unit":    "technology",
				"requested_roles":  []string{"deployer"},
				"requested_scopes": []string{"deployment:production:deploy"},
				"justification":    "Implementação de feature #FEAT-2025-10",
				"duration":         float64(60), // minutos
				"emergency":        false,
			},
			policyDecision: &policy.PolicyDecision{
				Allowed: true,
				Reasons: []string{"Usuário autorizado para elevação"},
				Conditions: map[string]interface{}{
					"max_duration": float64(120), // minutos
				},
			},
			policyError:   nil,
			expectAllowed: true,
			expectReasons: []string{"Usuário autorizado para elevação"},
			expectConditions: map[string]interface{}{
				"max_duration": float64(120),
			},
		},
		{
			name:     "Política nega elevação - escopo não permitido",
			policyID: "innovabiz:iam:elevation:scope",
			policyInput: map[string]interface{}{
				"user_id":          "user:developer:456",
				"tenant_id":        "tenant_angola_1",
				"market":           "angola",
				"business_unit":    "technology",
				"requested_roles":  []string{"security_admin"},
				"requested_scopes": []string{"security:keys:manage"},
				"justification":    "Rotação de chaves de criptografia",
				"duration":         float64(30),
				"emergency":        false,
			},
			policyDecision: &policy.PolicyDecision{
				Allowed: false,
				Reasons: []string{
					"Escopo security:keys:manage requer papel segurança especializado",
					"Usuário não tem permissão para solicitar este escopo",
				},
				Conditions: nil,
			},
			policyError:      nil,
			expectAllowed:    false,
			expectReasons:    []string{"Escopo security:keys:manage requer papel segurança especializado", "Usuário não tem permissão para solicitar este escopo"},
			expectConditions: nil,
		},
		{
			name:     "Política permite elevação com condições",
			policyID: "innovabiz:iam:elevation:approval",
			policyInput: map[string]interface{}{
				"user_id":          "user:operator:789",
				"tenant_id":        "tenant_angola_1",
				"market":           "angola",
				"business_unit":    "operations",
				"requested_roles":  []string{"admin"},
				"requested_scopes": []string{"k8s:production:pods:delete"},
				"justification":    "Incidente de produção #INC-2025-42",
				"duration":         float64(30),
				"emergency":        true,
			},
			policyDecision: &policy.PolicyDecision{
				Allowed: true,
				Reasons: []string{"Aprovado por emergência para operador de produção"},
				Conditions: map[string]interface{}{
					"require_mfa":            true,
					"max_duration":           float64(60),
					"notify_security_team":   true,
					"namespace_restrictions": []string{"prod-frontend", "prod-api"},
					"pod_name_pattern":       "nginx-*",
				},
			},
			policyError:   nil,
			expectAllowed: true,
			expectReasons: []string{"Aprovado por emergência para operador de produção"},
			expectConditions: map[string]interface{}{
				"require_mfa":            true,
				"max_duration":           float64(60),
				"notify_security_team":   true,
				"namespace_restrictions": []string{"prod-frontend", "prod-api"},
				"pod_name_pattern":       "nginx-*",
			},
		},
		{
			name:     "Política de uso - horário não permitido",
			policyID: "innovabiz:iam:elevation:usage",
			policyInput: map[string]interface{}{
				"user_id":          "user:developer:101",
				"tenant_id":        "tenant_angola_1",
				"market":           "angola",
				"business_unit":    "technology",
				"command":          "kubectl delete deployment frontend-app",
				"resource":         "deployment/frontend-app",
				"namespace":        "production",
				"time_of_day":      23, // 23h - fora do horário comercial
				"elevated_roles":   []string{"deployer"},
				"elevated_scopes":  []string{"deployment:production:delete"},
				"elevation_id":     "elev-2025-010",
				"day_of_week":      "Saturday", // Fim de semana
			},
			policyDecision: &policy.PolicyDecision{
				Allowed: false,
				Reasons: []string{
					"Operações em produção não permitidas fora do horário comercial",
					"Operações em produção não permitidas aos finais de semana",
					"Requer aprovação adicional do gerente de produção",
				},
				Conditions: nil,
			},
			policyError:      nil,
			expectAllowed:    false,
			expectReasons:    []string{"Operações em produção não permitidas fora do horário comercial", "Operações em produção não permitidas aos finais de semana", "Requer aprovação adicional do gerente de produção"},
			expectConditions: nil,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, tc.name)
			
			// Configurar mock do motor de políticas
			mockPolicyEngine.On("EvaluatePolicy", mock.Anything, tc.policyID, mock.MatchedBy(func(input map[string]interface{}) bool {
				// Verificar apenas se os campos principais estão presentes
				_, hasUserID := input["user_id"]
				_, hasTenantID := input["tenant_id"]
				return hasUserID && hasTenantID
			})).Return(tc.policyDecision, tc.policyError)
			
			// Configurar mock do logger
			mockLogger.On("LogSecurityEvent", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			
			// Executar avaliação de política
			allowed, reasons, conditions, err := elevationManager.EvaluateElevationPolicy(testCtx, tc.policyID, tc.policyInput)
			
			// Verificar resultados
			require.NoError(t, err, "Avaliação de política não deveria falhar")
			assert.Equal(t, tc.expectAllowed, allowed, "Resultado da política incorreto")
			
			if tc.expectReasons != nil {
				assert.ElementsMatch(t, tc.expectReasons, reasons, "Razões da política incorretas")
			}
			
			if tc.expectConditions != nil {
				for key, expectedValue := range tc.expectConditions {
					actualValue, exists := conditions[key]
					assert.True(t, exists, "Condição %s não encontrada", key)
					assert.Equal(t, expectedValue, actualValue, "Valor da condição %s incorreto", key)
				}
			}
			
			// Verificar chamadas aos mocks
			mockPolicyEngine.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
			
			// Registrar conclusão do teste
			obs.RecordTestEnd(testCtx, tc.name, true, time.Since(time.Now()))
		})
	}
}

// TestPrivilegeElevationPolicyIntegration testa a integração entre políticas e fluxo de elevação
func TestPrivilegeElevationPolicyIntegration(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("privilege_elevation_policy_integration_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Configurar mocks
	mockApprover := new(MockElevationApprover)
	mockLogger := new(MockAuditLogger)
	mockNotifier := new(MockNotifier)
	mockPolicyEngine := new(MockPolicyEngine)
	
	// Criar o gerenciador de elevação de privilégios com os mocks
	elevationManager := elevation.NewPrivilegeElevationManager(
		mockApprover,
		mockLogger,
		mockNotifier,
	)
	
	// Configurar motor de políticas
	elevationManager.ConfigurePolicyEngine(mockPolicyEngine)
	
	// Configurar IDs de política
	elevationManager.ConfigurePolicyIDs(&elevation.PolicyConfiguration{
		ElevationRequestPolicy:  "innovabiz:iam:elevation:request",
		ElevationApprovalPolicy: "innovabiz:iam:elevation:approval",
		ElevationUsagePolicy:    "innovabiz:iam:elevation:usage",
		ElevationScopePolicy:    "innovabiz:iam:elevation:scope",
	})
	
	// Criar solicitação de elevação para teste
	baseTime := time.Now()
	elevationRequest := &elevation.ElevationRequest{
		UserID:          "user:operator:202",
		TenantID:        "tenant_angola_1",
		Justification:   "Manutenção de produção programada",
		RequestedRoles:  []string{"admin"},
		RequestedScopes: []string{"k8s:production:pods:restart"},
		Duration:        30 * time.Minute,
		EmergencyAccess: false,
		Market:          "angola",
		BusinessUnit:    "operations",
	}
	
	// Criar aprovação para teste
	elevationApproval := &elevation.ElevationApproval{
		ElevationID:      "elev-policy-001",
		UserID:           elevationRequest.UserID,
		ApprovedBy:       "user:manager:303",
		ApprovalTime:     baseTime,
		ExpirationTime:   baseTime.Add(30 * time.Minute),
		ElevatedRoles:    elevationRequest.RequestedRoles,
		ElevatedScopes:   elevationRequest.RequestedScopes,
		ApprovalEvidence: "ticket:MAINT-2025-05",
		AuditMetadata: map[string]interface{}{
			"request_ip":     "192.168.1.100",
			"approval_notes": "Aprovado para manutenção programada",
		},
	}
	
	// Configurar decisões de política para diferentes etapas do fluxo
	requestPolicyDecision := &policy.PolicyDecision{
		Allowed: true,
		Reasons: []string{"Solicitação válida para operador de produção"},
		Conditions: map[string]interface{}{
			"require_approval":      true,
			"allowed_approvers":     []string{"user:manager:303", "user:manager:404"},
			"max_duration":          float64(60),
			"require_justification": true,
		},
	}
	
	scopePolicyDecision := &policy.PolicyDecision{
		Allowed: true,
		Reasons: []string{"Escopo permitido para papel do usuário"},
		Conditions: map[string]interface{}{
			"require_mfa": true,
		},
	}
	
	approvalPolicyDecision := &policy.PolicyDecision{
		Allowed: true,
		Reasons: []string{"Aprovador autorizado"},
		Conditions: map[string]interface{}{
			"max_duration":         float64(45),
			"requires_ticket":      true,
			"namespace_limitation": "production",
			"pod_name_pattern":     "*",
		},
	}
	
	// Configurar mocks para avaliação de políticas
	mockPolicyEngine.On("EvaluatePolicy", mock.Anything, "innovabiz:iam:elevation:request", mock.Anything).
		Return(requestPolicyDecision, nil)
		
	mockPolicyEngine.On("EvaluatePolicy", mock.Anything, "innovabiz:iam:elevation:scope", mock.Anything).
		Return(scopePolicyDecision, nil)
		
	mockPolicyEngine.On("EvaluatePolicy", mock.Anything, "innovabiz:iam:elevation:approval", mock.Anything).
		Return(approvalPolicyDecision, nil)
	
	// Configurar mock do aprovador
	mockApprover.On("ApproveElevation", mock.Anything, elevationRequest).
		Return(elevationApproval, nil)
	
	// Configurar mock do logger
	mockLogger.On("LogElevationEvent", mock.Anything, mock.Anything).Return(nil)
	mockLogger.On("LogSecurityEvent", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	
	// Configurar mock do notificador
	mockNotifier.On("NotifyElevationRequest", mock.Anything, mock.Anything).Return(nil)
	mockNotifier.On("NotifyElevationApproval", mock.Anything, mock.Anything).Return(nil)
	
	// Executar teste - solicitar elevação com avaliação de políticas
	ctx := context.Background()
	testCtx := obs.RecordTestStart(ctx, "RequestElevationWithPolicyEvaluation")
	
	// Habilitar avaliação de políticas para o teste
	elevationManager.EnablePolicyEnforcement(true)
	
	// Solicitar elevação
	result, err := elevationManager.RequestElevation(testCtx, elevationRequest)
	require.NoError(t, err, "Solicitação de elevação com políticas não deveria falhar")
	assert.NotNil(t, result, "Resultado da elevação não deveria ser nulo")
	
	// Verificar que as condições da política foram aplicadas
	assert.Equal(t, elevationApproval.ElevationID, result.ElevationID, "ID de elevação incorreto")
	assert.Equal(t, elevationApproval.UserID, result.UserID, "UserID incorreto")
	assert.Equal(t, elevationApproval.ElevatedRoles, result.ElevatedRoles, "Papéis elevados incorretos")
	assert.Equal(t, elevationApproval.ElevatedScopes, result.ElevatedScopes, "Escopos elevados incorretos")
	assert.NotEmpty(t, result.ElevationToken, "Token de elevação não deveria ser vazio")
	
	// Verificar metadados de auditoria baseados nas políticas
	assert.Contains(t, result.AuditMetadata, "policy_conditions", "Metadados deveriam incluir condições de política")
	assert.Contains(t, result.AuditMetadata, "approval_evidence", "Metadados deveriam incluir evidência de aprovação")
	
	// Verificar que o escopo foi limitado conforme política
	if policyConditions, ok := result.AuditMetadata["policy_conditions"].(map[string]interface{}); ok {
		assert.Contains(t, policyConditions, "namespace_limitation", "Condições deveriam incluir limitação de namespace")
		assert.Contains(t, policyConditions, "pod_name_pattern", "Condições deveriam incluir padrão de nome de pod")
	} else {
		t.Errorf("policy_conditions não é um mapa como esperado")
	}
	
	// Verificar chamadas aos mocks
	mockPolicyEngine.AssertExpectations(t)
	mockApprover.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	mockNotifier.AssertExpectations(t)
	
	// Registrar conclusão do teste
	obs.RecordTestEnd(testCtx, "RequestElevationWithPolicyEvaluation", err == nil, time.Since(time.Now()))
}

// TestPrivilegeElevationMarketSpecificPolicies testa políticas específicas por mercado
func TestPrivilegeElevationMarketSpecificPolicies(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("privilege_elevation_market_policy_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Configurar mocks
	mockApprover := new(MockElevationApprover)
	mockLogger := new(MockAuditLogger)
	mockNotifier := new(MockNotifier)
	mockPolicyEngine := new(MockPolicyEngine)
	
	// Criar o gerenciador de elevação de privilégios com os mocks
	elevationManager := elevation.NewPrivilegeElevationManager(
		mockApprover,
		mockLogger,
		mockNotifier,
	)
	
	// Configurar motor de políticas
	elevationManager.ConfigurePolicyEngine(mockPolicyEngine)
	
	// Configurar IDs de política específicos por mercado
	marketPolicies := map[string]*elevation.PolicyConfiguration{
		"angola": {
			ElevationRequestPolicy:  "innovabiz:iam:angola:elevation:request",
			ElevationApprovalPolicy: "innovabiz:iam:angola:elevation:approval",
			ElevationUsagePolicy:    "innovabiz:iam:angola:elevation:usage",
			ElevationScopePolicy:    "innovabiz:iam:angola:elevation:scope",
		},
		"brazil": {
			ElevationRequestPolicy:  "innovabiz:iam:brazil:elevation:request",
			ElevationApprovalPolicy: "innovabiz:iam:brazil:elevation:approval",
			ElevationUsagePolicy:    "innovabiz:iam:brazil:elevation:usage",
			ElevationScopePolicy:    "innovabiz:iam:brazil:elevation:scope",
		},
		"global": {
			ElevationRequestPolicy:  "innovabiz:iam:global:elevation:request",
			ElevationApprovalPolicy: "innovabiz:iam:global:elevation:approval",
			ElevationUsagePolicy:    "innovabiz:iam:global:elevation:usage",
			ElevationScopePolicy:    "innovabiz:iam:global:elevation:scope",
		},
	}
	
	elevationManager.ConfigureMarketSpecificPolicies(marketPolicies)
	
	// Habilitar avaliação de políticas
	elevationManager.EnablePolicyEnforcement(true)
	
	// Configurar logger
	mockLogger.On("LogSecurityEvent", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	
	// Casos de teste para diferentes mercados
	testCases := []struct {
		name          string
		market        string
		expectedPolicyID string
		policyDecision *policy.PolicyDecision
	}{
		{
			name:   "Política específica de Angola",
			market: "angola",
			expectedPolicyID: "innovabiz:iam:angola:elevation:request",
			policyDecision: &policy.PolicyDecision{
				Allowed: true,
				Reasons: []string{"Regra de Angola aplicada"},
				Conditions: map[string]interface{}{
					"angola_specific_condition": true,
				},
			},
		},
		{
			name:   "Política específica do Brasil",
			market: "brazil",
			expectedPolicyID: "innovabiz:iam:brazil:elevation:request",
			policyDecision: &policy.PolicyDecision{
				Allowed: true,
				Reasons: []string{"Regra do Brasil aplicada"},
				Conditions: map[string]interface{}{
					"brazil_specific_condition": true,
				},
			},
		},
		{
			name:   "Política global para mercado não específico",
			market: "portugal", // Mercado sem política específica
			expectedPolicyID: "innovabiz:iam:global:elevation:request",
			policyDecision: &policy.PolicyDecision{
				Allowed: true,
				Reasons: []string{"Regra global aplicada"},
				Conditions: map[string]interface{}{
					"global_condition": true,
				},
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, tc.name)
			
			// Configurar contexto com mercado
			marketCtx := elevation.WithMarket(testCtx, tc.market)
			
			// Configurar mock do motor de políticas para este caso específico
			mockPolicyEngine.On("EvaluatePolicy", mock.Anything, tc.expectedPolicyID, mock.Anything).
				Return(tc.policyDecision, nil).Once()
			
			// Input para avaliação de política
			policyInput := map[string]interface{}{
				"user_id":          "user:test:123",
				"tenant_id":        "tenant_global_1",
				"market":           tc.market,
				"business_unit":    "test",
				"requested_roles":  []string{"tester"},
				"requested_scopes": []string{"test:read"},
				"justification":    "Teste de políticas de mercado",
			}
			
			// Avaliar política de solicitação
			policyID := elevationManager.GetMarketSpecificPolicyID(marketCtx, "ElevationRequestPolicy")
			assert.Equal(t, tc.expectedPolicyID, policyID, "ID de política incorreto para mercado %s", tc.market)
			
			allowed, reasons, conditions, err := elevationManager.EvaluateElevationPolicy(marketCtx, policyID, policyInput)
			require.NoError(t, err, "Avaliação de política não deveria falhar")
			
			// Verificar resultados
			assert.True(t, allowed, "Política deveria permitir")
			assert.ElementsMatch(t, tc.policyDecision.Reasons, reasons, "Razões incorretas")
			
			// Verificar condições específicas do mercado
			for key, expectedValue := range tc.policyDecision.Conditions {
				actualValue, exists := conditions[key]
				assert.True(t, exists, "Condição %s não encontrada", key)
				assert.Equal(t, expectedValue, actualValue, "Valor da condição %s incorreto", key)
			}
			
			// Registrar conclusão do teste
			obs.RecordTestEnd(testCtx, tc.name, err == nil, time.Since(time.Now()))
		})
	}
	
	// Verificar chamadas aos mocks
	mockPolicyEngine.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}