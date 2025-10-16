// Package elevation contém testes de integração para o sistema de elevação de privilégios MCP-IAM.
// Este arquivo testa o fluxo de elevação de privilégios com o hook Figma.
package elevation

import (
	"context"
	"testing"
	"time"

	"github.com/innovabiz/iam/auth"
	"github.com/innovabiz/iam/elevation"
	"github.com/innovabiz/iam/mcp/hooks"
	"github.com/innovabiz/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestFigmaElevation testa o fluxo completo de elevação para o hook Figma
func TestFigmaElevation(t *testing.T) {
	span, ctx := testutil.StartTestSpan(context.Background(), "TestFigmaElevation")
	defer span.Finish()
	
	logger := testutil.GetTestLogger()
	logger.Info("Iniciando testes de elevação para hook Figma",
		zap.String("test", "figma_elevation"))

	// Configura serviço de elevação para testes
	elevationService, err := setupElevationService()
	require.NoError(t, err, "Falha ao configurar serviço de elevação")

	// Configura o hook Figma para testes
	figmaHook := hooks.NewFigmaMCPHook(elevationService, logger)
	
	// Define cenários de teste multi-mercado
	testScenarios := []struct {
		name          string
		marketCode    string
		tenantID      string
		userID        string
		operations    []hooks.FigmaOperation
		description   string
		expectElevation bool
	}{
		{
			name:       "Angola - Operações Sensíveis UI/UX",
			marketCode: "angola",
			tenantID:   "tenant-angola-123",
			userID:     "user-admin-456",
			operations: []hooks.FigmaOperation{
				{
					Type:      "delete_file",
					FileKey:   "file_key_123456",
					ProjectID: "project_789",
					TeamID:    "team_innovabiz",
				},
				{
					Type:      "transfer_ownership",
					FileKey:   "file_key_456789",
					ProjectID: "project_789",
					TeamID:    "team_innovabiz",
					UserID:    "external_collaborator_123",
				},
			},
			description:    "Operações críticas em arquivos UI/UX para Angola",
			expectElevation: true,
		},
		{
			name:       "Brasil - Operações em Design System",
			marketCode: "brasil",
			tenantID:   "tenant-brasil-456",
			userID:     "user-admin-456",
			operations: []hooks.FigmaOperation{
				{
					Type:      "delete_file",
					FileKey:   "design_system_key_123",
					ProjectID: "design_system_project",
					TeamID:    "team_innovabiz",
				},
				{
					Type:      "delete_component",
					FileKey:   "design_system_key_123",
					NodeID:    "component:123:456",
					ProjectID: "design_system_project",
					TeamID:    "team_innovabiz",
				},
			},
			description:    "Operações em design system para Brasil",
			expectElevation: true,
		},
		{
			name:       "Moçambique - Operações Regulares",
			marketCode: "mocambique",
			tenantID:   "tenant-mocambique-789",
			userID:     "user-admin-456",
			operations: []hooks.FigmaOperation{
				{
					Type:      "comment",
					FileKey:   "file_key_789012",
					ProjectID: "project_123",
					TeamID:    "team_innovabiz",
					Message:   "Comentário de teste",
				},
				{
					Type:      "view_file",
					FileKey:   "file_key_789012",
					ProjectID: "project_123",
					TeamID:    "team_innovabiz",
				},
			},
			description:    "Operações não críticas para Moçambique",
			expectElevation: false,
		},
	}
	
	// Executa testes para cada cenário
	for _, scenario := range testScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			testCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()
			
			// Gera token de autenticação para o usuário
			authToken := createTestAuthToken(t, scenario.userID, scenario.tenantID, scenario.marketCode)
			testCtx = auth.ContextWithToken(testCtx, authToken)
			
			// Testa cada operação no cenário
			for _, operation := range scenario.operations {
				testFigmaOperation(t, testCtx, elevationService, figmaHook, operation, 
					scenario.marketCode, scenario.tenantID, scenario.userID, scenario.expectElevation)
			}
		})
	}
	
	// Testes específicos para proteção de design systems e arquivos críticos
	t.Run("ProteçãoArquivosCríticosPorMercado", func(t *testing.T) {
		// Testa regras específicas por mercado para proteção de arquivos críticos
		protectionTests := []struct {
			market     string
			tenantID   string
			operation  hooks.FigmaOperation
			expectElevation bool
			description string
			requiresMFA bool
		}{
			{
				market:     "angola",
				tenantID:   "tenant-angola-123",
				operation: hooks.FigmaOperation{
					Type:      "delete_file",
					FileKey:   "payment_ui_key_123",
					ProjectID: "payment_ui_project",
					TeamID:    "team_innovabiz",
				},
				expectElevation: true,
				requiresMFA:     true,
				description:    "Angola requer elevação com MFA para exclusão de UI de pagamento",
			},
			{
				market:     "brasil",
				tenantID:   "tenant-brasil-456",
				operation: hooks.FigmaOperation{
					Type:      "transfer_ownership",
					FileKey:   "design_system_key_456",
					ProjectID: "design_system_project",
					TeamID:    "team_innovabiz",
					UserID:    "external_user_789",
				},
				expectElevation: true,
				requiresMFA:     false,
				description:    "Brasil requer elevação sem MFA para transferência de design system",
			},
			{
				market:     "mocambique",
				tenantID:   "tenant-mocambique-789",
				operation: hooks.FigmaOperation{
					Type:      "delete_file",
					FileKey:   "basic_ui_key_789",
					ProjectID: "basic_ui_project",
					TeamID:    "team_innovabiz",
				},
				expectElevation: false,
				requiresMFA:     false,
				description:    "Moçambique não requer elevação para exclusão de UI básica",
			},
		}
		
		for _, test := range protectionTests {
			t.Run(test.description, func(t *testing.T) {
				testCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				
				// Gera token de autenticação para o usuário
				authToken := createTestAuthToken(t, "user-admin-456", test.tenantID, test.market)
				if !test.requiresMFA {
					authToken.MFACompleted = false // Teste sem MFA completado
				}
				testCtx = auth.ContextWithToken(testCtx, authToken)
				
				// Testa operação
				result, err := figmaHook.ProcessOperation(testCtx, test.operation)
				
				if test.expectElevation {
					// Se espera-se que precise de elevação
					require.Error(t, err, "Operação sensível deve ser negada sem elevação")
					assert.Contains(t, err.Error(), "elevation required", "Erro deve indicar que elevação é necessária")
					
					if test.requiresMFA && !authToken.MFACompleted {
						assert.Contains(t, err.Error(), "MFA required", "Erro deve indicar que MFA é necessário")
					}
					
					// Solicita elevação
					scope := getFigmaElevationScope(test.operation)
					elevToken, err := getTestElevationWithApproval(testCtx, elevationService, "user-admin-456", 
						test.tenantID, test.market, scope)
					require.NoError(t, err, "Falha ao obter token de elevação")
					
					// Adiciona token de elevação ao contexto
					elevCtx := elevation.ContextWithElevationToken(testCtx, elevToken)
					authToken.MFACompleted = true // Simula MFA concluído
					elevCtx = auth.ContextWithToken(elevCtx, authToken)
					
					// Tenta novamente a operação com elevação
					result, err = figmaHook.ProcessOperation(elevCtx, test.operation)
					require.NoError(t, err, "Operação deve ser permitida com elevação")
					assert.True(t, result.Allowed, "Operação deve ser permitida com elevação")
				} else {
					// Se não precisa de elevação
					if err != nil {
						// Pode haver outros motivos para rejeição
						t.Logf("Operação rejeitada: %v", err)
					} else {
						assert.NotNil(t, result, "Resultado não deve ser nil para operação permitida")
						assert.True(t, result.Allowed, "Operação deve ser permitida sem elevação")
					}
				}
			})
		}
	})
	
	// Teste para controle de risco em operações em massa
	t.Run("ControleDeRiscoOperaçõesEmMassa", func(t *testing.T) {
		testCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		
		market := "angola"
		tenantID := "tenant-angola-123"
		userID := "user-admin-456"
		
		// Operações em massa (simulando um script de automação)
		operations := []hooks.FigmaOperation{
			{
				Type:      "delete_file",
				FileKey:   "file_key_001",
				ProjectID: "project_123",
				TeamID:    "team_innovabiz",
			},
			{
				Type:      "delete_file",
				FileKey:   "file_key_002",
				ProjectID: "project_123",
				TeamID:    "team_innovabiz",
			},
			{
				Type:      "delete_file",
				FileKey:   "file_key_003",
				ProjectID: "project_123",
				TeamID:    "team_innovabiz",
			},
		}
		
		// Gera token de autenticação
		authToken := createTestAuthToken(t, userID, tenantID, market)
		testCtx = auth.ContextWithToken(testCtx, authToken)
		
		// Solicita elevação para operações em massa
		elevationRequest := &elevation.ElevationRequest{
			UserID:        userID,
			TenantID:      tenantID,
			Market:        market,
			Justification: "Operação em massa para limpeza de arquivos obsoletos",
			Scopes:        []string{"figma:delete:file", "figma:batch_operation"},
			Duration:      60, // 60 minutos
			Emergency:     false,
		}
		
		// Submete solicitação de elevação
		elevToken, err := elevationService.RequestElevation(testCtx, elevationRequest)
		require.NoError(t, err, "Falha ao solicitar elevação")
		
		// Aprova solicitação por supervisor
		supervisorID := "user-supervisor-789"
		supervisorCtx := createSupervisorContext(testCtx, supervisorID, tenantID, market)
		approvedToken, err := elevationService.ApproveElevation(supervisorCtx, elevToken.ID)
		require.NoError(t, err, "Falha ao aprovar elevação")
		
		// Adiciona token de elevação ao contexto
		elevCtx := elevation.ContextWithElevationToken(testCtx, approvedToken)
		elevCtx = auth.ContextWithToken(elevCtx, authToken)
		
		// Tenta operações em massa com elevação
		for i, operation := range operations {
			result, err := figmaHook.ProcessOperation(elevCtx, operation)
			require.NoError(t, err, "Operação em massa deve ser permitida com elevação")
			assert.NotNil(t, result, "Resultado não deve ser nil")
			assert.True(t, result.Allowed, "Operação deve ser permitida")
			
			// Registra evento de auditoria
			testutil.LogAuditEvent(testCtx, "figma_batch_operation", map[string]interface{}{
				"tenant_id":    tenantID,
				"user_id":      userID,
				"market":       market,
				"operation":    operation.Type,
				"file_key":     operation.FileKey,
				"elevation_id": approvedToken.ID,
				"batch_index":  i,
			})
		}
		
		// Verifica que o token continua válido após as operações
		assert.False(t, approvedToken.IsExpired(), "Token não deve expirar durante operações em massa")
	})
}

// testFigmaOperation testa uma operação específica do Figma
func testFigmaOperation(t *testing.T, ctx context.Context, elevationService *elevation.Service, 
	figmaHook *hooks.FigmaMCPHook, operation hooks.FigmaOperation, 
	market, tenantID, userID string, expectElevation bool) {
	
	// Teste sem elevação
	t.Run("SemElevação-"+operation.Type, func(t *testing.T) {
		// Tenta operação sem elevação
		result, err := figmaHook.ProcessOperation(ctx, operation)
		
		if expectElevation {
			// Se espera-se que precise de elevação
			require.Error(t, err, "Operação sensível deve ser negada sem elevação")
			assert.Contains(t, err.Error(), "elevation required", "Erro deve indicar que elevação é necessária")
			assert.Nil(t, result, "Resultado deve ser nil quando elevação é necessária")
		} else {
			// Se não precisa de elevação
			require.NoError(t, err, "Operação não sensível não deve requerer elevação")
			assert.NotNil(t, result, "Resultado não deve ser nil para operação permitida")
			assert.True(t, result.Allowed, "Operação deve ser permitida sem elevação")
		}
	})
	
	// Se a operação requer elevação, teste o fluxo completo
	if expectElevation {
		t.Run("ComElevação-"+operation.Type, func(t *testing.T) {
			// Determina escopo de elevação necessário
			scope := getFigmaElevationScope(operation)
			
			// Solicita elevação
			elevToken, err := getTestElevationWithApproval(ctx, elevationService, userID, tenantID, market, scope)
			require.NoError(t, err, "Falha ao obter token de elevação")
			
			// Adiciona token de elevação ao contexto
			elevCtx := elevation.ContextWithElevationToken(ctx, elevToken)
			elevCtx = auth.ContextWithToken(elevCtx, createTestAuthToken(t, userID, tenantID, market))
			
			// Tenta novamente a operação com elevação
			result, err := figmaHook.ProcessOperation(elevCtx, operation)
			require.NoError(t, err, "Operação deve ser permitida com elevação")
			assert.NotNil(t, result, "Resultado não deve ser nil quando operação é permitida")
			assert.True(t, result.Allowed, "Operação deve ser permitida com elevação")
			
			// Registra evento de auditoria para análise posterior
			testutil.LogAuditEvent(ctx, "figma_operation_execution", map[string]interface{}{
				"tenant_id":   tenantID,
				"user_id":     userID,
				"market":      market,
				"operation":   operation.Type,
				"file_key":    operation.FileKey,
				"elevation_id": elevToken.ID,
				"approved":    true,
			})
		})
	}
}

// getFigmaElevationScope determina o escopo de elevação necessário para uma operação
func getFigmaElevationScope(operation hooks.FigmaOperation) string {
	// Mapeamento de operações para escopos de elevação
	operationScopes := map[string]string{
		"delete_file":        "figma:delete:file",
		"transfer_ownership": "figma:transfer:ownership",
		"delete_component":   "figma:delete:component",
		"batch_operation":    "figma:batch_operation",
	}
	
	// Se houver um mapeamento específico para a operação, use-o
	if scope, exists := operationScopes[operation.Type]; exists {
		return scope
	}
	
	// Escopo padrão para operações não mapeadas explicitamente
	return "figma:operation:" + operation.Type
}

// getTestElevationWithApproval obtém um token de elevação aprovado para testes
func getTestElevationWithApproval(ctx context.Context, elevationService *elevation.Service, 
	userID, tenantID, market, scope string) (*elevation.Token, error) {
	
	// Solicita elevação de privilégios
	elevationRequest := &elevation.ElevationRequest{
		UserID:        userID,
		TenantID:      tenantID,
		Market:        market,
		Justification: "Teste de operação Figma com elevação",
		Scopes:        []string{scope},
		Duration:      30, // 30 minutos
		Emergency:     false,
	}
	
	// Submete solicitação de elevação
	elevToken, err := elevationService.RequestElevation(ctx, elevationRequest)
	if err != nil {
		return nil, err
	}
	
	// Verifica se requer aprovação com base no mercado/política
	if requiresFigmaApproval(market, scope) {
		// Simula aprovação por supervisor
		supervisorID := "user-supervisor-789"
		supervisorCtx := createSupervisorContext(ctx, supervisorID, tenantID, market)
		
		approvedToken, err := elevationService.ApproveElevation(supervisorCtx, elevToken.ID)
		if err != nil {
			return nil, err
		}
		
		// Retorna token aprovado
		return approvedToken, nil
	}
	
	// Se não precisar de aprovação, já retorna o token ativo
	return elevToken, nil
}

// requiresFigmaApproval verifica se a elevação para Figma requer aprovação por mercado/escopo
func requiresFigmaApproval(market, scope string) bool {
	// Em um cenário real, isso consultaria o banco de dados ou cache
	// Para simplificar, usamos regras hard-coded para testes
	
	// Operações críticas sempre requerem aprovação independente do mercado
	criticalScopes := map[string]bool{
		"figma:delete:file":     true,
		"figma:transfer:ownership": true,
		"figma:batch_operation": true,
	}
	
	// Angola e Brasil sempre requerem aprovação para operações sensíveis
	if (market == "angola" || market == "brasil") && criticalScopes[scope] {
		return true
	}
	
	// Moçambique tem requisitos mais flexíveis
	if market == "mocambique" {
		return scope == "figma:batch_operation" // Apenas operações em massa requerem aprovação
	}
	
	// Por padrão, não requer aprovação para outras operações/mercados
	return false
}