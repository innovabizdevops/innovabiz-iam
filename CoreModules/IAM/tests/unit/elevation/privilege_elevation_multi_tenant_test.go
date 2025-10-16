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
	"github.com/innovabizdevops/innovabiz-iam/tests/testutil"
)

// TestPrivilegeElevationManager_MultiTenant testa o isolamento multi-tenant na elevação de privilégios
func TestPrivilegeElevationManager_MultiTenant(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("privilege_elevation_multi_tenant_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Tempo base para os testes
	baseTime := time.Now()
	
	// Configurar mocks
	mockApprover := new(MockElevationApprover)
	mockLogger := new(MockAuditLogger)
	mockNotifier := new(MockNotifier)
	
	// Configurar solicitações e aprovações para diferentes tenants
	tenantRequests := map[string]*elevation.ElevationRequest{
		"tenant_angola_1": {
			UserID:          "user:operator:123",
			TenantID:        "tenant_angola_1",
			Justification:   "Acesso emergencial Angola",
			RequestedRoles:  []string{"admin"},
			RequestedScopes: []string{"k8s:angola:production:pods:delete"},
			Duration:        30 * time.Minute,
			EmergencyAccess: true,
			Market:          "angola",
			BusinessUnit:    "operations",
		},
		"tenant_brazil_1": {
			UserID:          "user:operator:456",
			TenantID:        "tenant_brazil_1",
			Justification:   "Acesso emergencial Brasil",
			RequestedRoles:  []string{"admin"},
			RequestedScopes: []string{"k8s:brazil:production:pods:delete"},
			Duration:        30 * time.Minute,
			EmergencyAccess: true,
			Market:          "brazil",
			BusinessUnit:    "operations",
		},
	}
	
	tenantApprovals := map[string]*elevation.ElevationApproval{
		"tenant_angola_1": {
			ElevationID:      "elev-angola-001",
			UserID:           "user:operator:123",
			ApprovedBy:       "system:emergency:auto",
			ApprovalTime:     baseTime,
			ExpirationTime:   baseTime.Add(30 * time.Minute),
			ElevatedRoles:    []string{"admin"},
			ElevatedScopes:   []string{"k8s:angola:production:pods:delete"},
			ApprovalEvidence: "emergency_auto_approval:angola",
			AuditMetadata: map[string]interface{}{
				"request_ip":       "192.168.1.100",
				"emergency_access": true,
				"tenant_id":        "tenant_angola_1",
				"market":           "angola",
			},
		},
		"tenant_brazil_1": {
			ElevationID:      "elev-brazil-001",
			UserID:           "user:operator:456",
			ApprovedBy:       "system:emergency:auto",
			ApprovalTime:     baseTime,
			ExpirationTime:   baseTime.Add(30 * time.Minute),
			ElevatedRoles:    []string{"admin"},
			ElevatedScopes:   []string{"k8s:brazil:production:pods:delete"},
			ApprovalEvidence: "emergency_auto_approval:brazil",
			AuditMetadata: map[string]interface{}{
				"request_ip":       "192.168.1.200",
				"emergency_access": true,
				"tenant_id":        "tenant_brazil_1",
				"market":           "brazil",
			},
		},
	}
	
	// Configurar mock do aprovador para responder conforme o tenant
	for tenantID, request := range tenantRequests {
		mockApprover.On("ApproveElevation", mock.Anything, request).
			Return(tenantApprovals[tenantID], nil)
	}
	
	// Configurar mock do logger
	mockLogger.On("LogElevationEvent", mock.Anything, mock.Anything).Return(nil)
	mockLogger.On("LogSecurityEvent", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	
	// Configurar mock do notificador
	mockNotifier.On("NotifyElevationRequest", mock.Anything, mock.Anything).Return(nil)
	mockNotifier.On("NotifyElevationApproval", mock.Anything, mock.Anything).Return(nil)
	
	// Criar o gerenciador de elevação de privilégios com os mocks
	elevationManager := elevation.NewPrivilegeElevationManager(
		mockApprover,
		mockLogger,
		mockNotifier,
	)
	
	// Habilitar validação de isolamento multi-tenant
	elevationManager.ConfigureMultiTenantIsolation(true)
	
	// Testar solicitações e verificações para cada tenant
	tenantTokens := make(map[string]string)
	
	// 1. Processar solicitações de elevação para cada tenant
	for tenantID, request := range tenantRequests {
		ctx := context.Background()
		testCtx := obs.RecordTestStart(ctx, "Solicitar_elevação_"+tenantID)
		
		// Solicitar elevação para o tenant
		result, err := elevationManager.RequestElevation(testCtx, request)
		require.NoError(t, err, "Solicitação para %s falhou", tenantID)
		assert.NotEmpty(t, result.ElevationToken, "Token de elevação vazio para %s", tenantID)
		
		// Armazenar token para teste posterior
		tenantTokens[tenantID] = result.ElevationToken
		
		obs.RecordTestEnd(testCtx, "Solicitar_elevação_"+tenantID, err == nil, time.Since(time.Now()))
	}
	
	// 2. Testar isolamento multi-tenant nas verificações
	t.Run("Isolamento_na_verificação", func(t *testing.T) {
		for tenantID, token := range tenantTokens {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, "Verificar_isolamento_"+tenantID)
			
			// Testar acesso com tenant correto
			reqCtx := elevation.WithTenantID(testCtx, tenantID)
			valid, reason, record, err := elevationManager.VerifyElevation(reqCtx, token)
			require.NoError(t, err, "Verificação falhou para %s", tenantID)
			assert.True(t, valid, "Elevação deveria ser válida para tenant correto: %s", tenantID)
			assert.Empty(t, reason, "Não deveria haver razão de invalidação para %s", tenantID)
			assert.Equal(t, tenantApprovals[tenantID].ElevationID, record.ElevationID, "ID de elevação incorreto para %s", tenantID)
			
			// Testar acesso com tenant incorreto - para cada token, testar com os outros tenants
			for otherTenantID := range tenantRequests {
				if otherTenantID == tenantID {
					continue // Pular o mesmo tenant
				}
				
				// Contexto com tenant diferente
				wrongTenantCtx := elevation.WithTenantID(testCtx, otherTenantID)
				valid, reason, _, err := elevationManager.VerifyElevation(wrongTenantCtx, token)
				require.NoError(t, err, "Verificação com tenant errado não deveria causar erro")
				assert.False(t, valid, "Elevação não deveria ser válida com tenant diferente")
				assert.Contains(t, reason, "tenant não autorizado", "Razão de invalidação incorreta")
			}
			
			obs.RecordTestEnd(testCtx, "Verificar_isolamento_"+tenantID, true, time.Since(time.Now()))
		}
	})
	
	// 3. Testar uso cross-tenant não autorizado
	t.Run("Bloqueio_cross_tenant", func(t *testing.T) {
		ctx := context.Background()
		testCtx := obs.RecordTestStart(ctx, "Tentativa_cross_tenant")
		
		// Simular uma operação que tenta usar elevação de Angola para acessar recursos do Brasil
		angolaToken := tenantTokens["tenant_angola_1"]
		
		// Configurar um contexto com TenantID de Brasil
		crossTenantCtx := elevation.WithTenantID(testCtx, "tenant_brazil_1")
		
		// Verificar elevação - deve falhar por isolamento
		valid, reason, _, err := elevationManager.VerifyElevation(crossTenantCtx, angolaToken)
		require.NoError(t, err, "Verificação cross-tenant não deveria causar erro")
		assert.False(t, valid, "Elevação cross-tenant não deveria ser válida")
		assert.Contains(t, reason, "tenant não autorizado", "Razão de invalidação incorreta")
		
		obs.RecordTestEnd(testCtx, "Tentativa_cross_tenant", !valid, time.Since(time.Now()))
	})
	
	// Verificar chamadas aos mocks
	mockApprover.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	mockNotifier.AssertExpectations(t)
}

// TestPrivilegeElevationManager_MultiMarket testa o isolamento multi-market na elevação de privilégios
func TestPrivilegeElevationManager_MultiMarket(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("privilege_elevation_multi_market_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Tempo base para os testes
	baseTime := time.Now()
	
	// Configurar mocks
	mockApprover := new(MockElevationApprover)
	mockLogger := new(MockAuditLogger)
	mockNotifier := new(MockNotifier)
	
	// Criar elevações para diferentes mercados
	marketElevations := map[string]*elevation.ElevationRecord{
		"angola": {
			ElevationID:      "elev-market-angola-001",
			UserID:           "user:operator:123",
			TenantID:         "tenant_global_1", // Mesmo tenant
			ApprovedBy:       "user:manager:789",
			ApprovalTime:     baseTime.Add(-15 * time.Minute),
			ExpirationTime:   baseTime.Add(15 * time.Minute),
			ElevatedRoles:    []string{"admin"},
			ElevatedScopes:   []string{"k8s:angola:production:pods:delete"},
			ApprovalEvidence: "ticket:INC-2025-42",
			Status:           elevation.StatusActive,
			Market:           "angola",
			BusinessUnit:     "operations",
		},
		"brazil": {
			ElevationID:      "elev-market-brazil-001",
			UserID:           "user:operator:456",
			TenantID:         "tenant_global_1", // Mesmo tenant
			ApprovedBy:       "user:manager:789",
			ApprovalTime:     baseTime.Add(-15 * time.Minute),
			ExpirationTime:   baseTime.Add(15 * time.Minute),
			ElevatedRoles:    []string{"admin"},
			ElevatedScopes:   []string{"k8s:brazil:production:pods:delete"},
			ApprovalEvidence: "ticket:INC-2025-43",
			Status:           elevation.StatusActive,
			Market:           "brazil",
			BusinessUnit:     "operations",
		},
		"global": {
			ElevationID:      "elev-market-global-001",
			UserID:           "user:operator:789",
			TenantID:         "tenant_global_1", // Mesmo tenant
			ApprovedBy:       "user:manager:101",
			ApprovalTime:     baseTime.Add(-15 * time.Minute),
			ExpirationTime:   baseTime.Add(15 * time.Minute),
			ElevatedRoles:    []string{"global_admin"},
			ElevatedScopes:   []string{"k8s:global:production:pods:delete"},
			ApprovalEvidence: "ticket:INC-2025-44",
			Status:           elevation.StatusActive,
			Market:           "global", // Acesso global
			BusinessUnit:     "operations",
		},
	}
	
	// Configurar mock do logger
	mockLogger.On("LogSecurityEvent", mock.Anything, "privilege_elevation_verification", mock.Anything).
		Return(nil)
		
	mockLogger.On("LogSecurityEvent", mock.Anything, "privilege_elevation_market_restriction", mock.Anything).
		Return(nil)
	
	// Criar o gerenciador de elevação de privilégios com os mocks
	elevationManager := elevation.NewPrivilegeElevationManager(
		mockApprover,
		mockLogger,
		mockNotifier,
	)
	
	// Adicionar elevações ao gerenciador
	marketTokens := make(map[string]string)
	for market, record := range marketElevations {
		token := "market-token-" + market
		err := elevationManager.AddElevationToStore(token, record)
		require.NoError(t, err, "Falha ao adicionar elevação para mercado %s", market)
		
		elevationManager.AddElevationToIndex(record.ElevationID, token)
		marketTokens[market] = token
	}
	
	// Habilitar validação de isolamento multi-market
	elevationManager.ConfigureMultiMarketIsolation(true, []string{"global"}) // "global" é mercado especial com acesso universal
	
	// Testar verificação com diferentes contextos de mercado
	testCases := []struct {
		name          string
		elevationMarket string
		contextMarket string
		expectValid   bool
	}{
		{
			name:          "Mesmo mercado Angola",
			elevationMarket: "angola",
			contextMarket: "angola",
			expectValid:   true,
		},
		{
			name:          "Mesmo mercado Brasil",
			elevationMarket: "brazil",
			contextMarket: "brazil",
			expectValid:   true,
		},
		{
			name:          "Angola não pode acessar Brasil",
			elevationMarket: "angola",
			contextMarket: "brazil",
			expectValid:   false,
		},
		{
			name:          "Brasil não pode acessar Angola",
			elevationMarket: "brazil",
			contextMarket: "angola",
			expectValid:   false,
		},
		{
			name:          "Global pode acessar Angola",
			elevationMarket: "global",
			contextMarket: "angola",
			expectValid:   true,
		},
		{
			name:          "Global pode acessar Brasil",
			elevationMarket: "global",
			contextMarket: "brazil",
			expectValid:   true,
		},
		{
			name:          "Global pode acessar Global",
			elevationMarket: "global",
			contextMarket: "global",
			expectValid:   true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, tc.name)
			
			// Configurar contexto de mercado para a operação
			reqCtx := elevation.WithMarket(testCtx, tc.contextMarket)
			
			// Verificar elevação
			valid, reason, _, err := elevationManager.VerifyElevation(reqCtx, marketTokens[tc.elevationMarket])
			require.NoError(t, err, "Verificação não deveria causar erro")
			
			if tc.expectValid {
				assert.True(t, valid, "Elevação deveria ser válida para caso %s", tc.name)
				assert.Empty(t, reason, "Não deveria haver razão de invalidação")
			} else {
				assert.False(t, valid, "Elevação não deveria ser válida para caso %s", tc.name)
				assert.Contains(t, reason, "mercado não autorizado", "Razão de invalidação incorreta")
			}
			
			obs.RecordTestEnd(testCtx, tc.name, valid == tc.expectValid, time.Since(time.Now()))
		})
	}
	
	// Verificar chamadas aos mocks
	mockLogger.AssertExpectations(t)
}