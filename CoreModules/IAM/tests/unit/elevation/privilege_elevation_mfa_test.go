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
	"github.com/innovabizdevops/innovabiz-iam/mfa"
	"github.com/innovabizdevops/innovabiz-iam/tests/testutil"
)

// MockMFAProvider é um mock do provedor de MFA
type MockMFAProvider struct {
	mock.Mock
}

func (m *MockMFAProvider) StartMFAChallenge(ctx context.Context, userID string, challengeType string) (string, error) {
	args := m.Called(ctx, userID, challengeType)
	return args.String(0), args.Error(1)
}

func (m *MockMFAProvider) VerifyMFAToken(ctx context.Context, userID string, challengeID string, token string) (bool, error) {
	args := m.Called(ctx, userID, challengeID, token)
	return args.Bool(0), args.Error(1)
}

func (m *MockMFAProvider) GetUserMFAStatus(ctx context.Context, userID string) (*mfa.UserMFAStatus, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mfa.UserMFAStatus), args.Error(1)
}

// TestElevationMFAIntegration testa a integração de MFA com elevação de privilégios
func TestElevationMFAIntegration(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("privilege_elevation_mfa_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Tempo base para os testes
	baseTime := time.Now()
	
	// Casos de teste para integração MFA
	testCases := []struct {
		name             string
		userID           string
		mfaStatus        *mfa.UserMFAStatus
		operationRisk    string // "low", "medium", "high", "critical"
		mfaChallenge     string
		mfaVerifySuccess bool
		expectElevationSuccess bool
		expectError      string
	}{
		{
			name:   "Operação crítica com MFA configurado e validado",
			userID: "user:manager:123",
			mfaStatus: &mfa.UserMFAStatus{
				UserID:             "user:manager:123",
				MFAEnabled:         true,
				PrimaryMethod:      "totp",
				BackupMethodsSetup: []string{"sms"},
				LastVerified:       baseTime.Add(-24 * time.Hour),
			},
			operationRisk:         "critical",
			mfaChallenge:          "challenge-123",
			mfaVerifySuccess:      true,
			expectElevationSuccess: true,
			expectError:           "",
		},
		{
			name:   "Operação crítica com MFA configurado mas não validado",
			userID: "user:manager:456",
			mfaStatus: &mfa.UserMFAStatus{
				UserID:             "user:manager:456",
				MFAEnabled:         true,
				PrimaryMethod:      "totp",
				BackupMethodsSetup: []string{"sms"},
				LastVerified:       baseTime.Add(-24 * time.Hour),
			},
			operationRisk:         "critical",
			mfaChallenge:          "challenge-456",
			mfaVerifySuccess:      false, // Falha na validação MFA
			expectElevationSuccess: false,
			expectError:           "falha na validação MFA",
		},
		{
			name:   "Operação crítica sem MFA configurado",
			userID: "user:manager:789",
			mfaStatus: &mfa.UserMFAStatus{
				UserID:             "user:manager:789",
				MFAEnabled:         false, // MFA não habilitado
				PrimaryMethod:      "",
				BackupMethodsSetup: []string{},
				LastVerified:       time.Time{},
			},
			operationRisk:         "critical",
			mfaChallenge:          "",
			mfaVerifySuccess:      false,
			expectElevationSuccess: false,
			expectError:           "MFA obrigatório não configurado",
		},
		{
			name:   "Operação de baixo risco sem MFA",
			userID: "user:analyst:101",
			mfaStatus: &mfa.UserMFAStatus{
				UserID:             "user:analyst:101",
				MFAEnabled:         false,
				PrimaryMethod:      "",
				BackupMethodsSetup: []string{},
				LastVerified:       time.Time{},
			},
			operationRisk:         "low", // Operação de baixo risco
			mfaChallenge:          "",
			mfaVerifySuccess:      false, 
			expectElevationSuccess: true, // Mesmo sem MFA, é permitido por ser baixo risco
			expectError:           "",
		},
		{
			name:   "Operação de médio risco com MFA expirado",
			userID: "user:developer:202",
			mfaStatus: &mfa.UserMFAStatus{
				UserID:             "user:developer:202",
				MFAEnabled:         true,
				PrimaryMethod:      "totp",
				BackupMethodsSetup: []string{},
				LastVerified:       baseTime.Add(-72 * time.Hour), // MFA expirado (>48h)
			},
			operationRisk:         "medium",
			mfaChallenge:          "challenge-202",
			mfaVerifySuccess:      true,
			expectElevationSuccess: true, // Operação de risco médio com verificação MFA
			expectError:           "",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, tc.name)
			
			// Configurar mocks
			mockMFA := new(MockMFAProvider)
			mockApprover := new(MockElevationApprover)
			mockLogger := new(MockAuditLogger)
			mockNotifier := new(MockNotifier)
			
			// Configurar mock MFA
			mockMFA.On("GetUserMFAStatus", mock.Anything, tc.userID).
				Return(tc.mfaStatus, nil)
				
			if tc.mfaStatus.MFAEnabled && tc.operationRisk != "low" {
				mockMFA.On("StartMFAChallenge", mock.Anything, tc.userID, "elevation").
					Return(tc.mfaChallenge, nil)
					
				mockMFA.On("VerifyMFAToken", mock.Anything, tc.userID, tc.mfaChallenge, mock.Anything).
					Return(tc.mfaVerifySuccess, nil)
			}
			
			// Configurar mock do logger
			mockLogger.On("LogSecurityEvent", mock.Anything, mock.Anything, mock.Anything).
				Return(nil)
			
			// Aprovação condicional com base no caso de teste
			elevationRequest := &elevation.ElevationRequest{
				UserID:          tc.userID,
				TenantID:        "tenant_angola_1",
				Justification:   "Teste MFA",
				RequestedRoles:  []string{"admin"},
				RequestedScopes: []string{"system:config:write"},
				Duration:        30 * time.Minute,
				EmergencyAccess: tc.operationRisk == "critical",
				Market:          "angola",
				BusinessUnit:    "technology",
			}
			
			if tc.expectElevationSuccess {
				mockApprover.On("ApproveElevation", mock.Anything, elevationRequest).
					Return(&elevation.ElevationApproval{
						ElevationID:      "elev-mfa-test-001",
						UserID:           tc.userID,
						ApprovedBy:       "system:emergency:auto",
						ApprovalTime:     baseTime,
						ExpirationTime:   baseTime.Add(30 * time.Minute),
						ElevatedRoles:    []string{"admin"},
						ElevatedScopes:   []string{"system:config:write"},
						ApprovalEvidence: "mfa_test",
						AuditMetadata: map[string]interface{}{
							"mfa_verified": tc.mfaVerifySuccess,
							"risk_level":   tc.operationRisk,
						},
					}, nil)
					
				mockLogger.On("LogElevationEvent", mock.Anything, mock.MatchedBy(func(event *elevation.ElevationEvent) bool {
					return event.EventType == elevation.EventTypeElevationRequested ||
						   event.EventType == elevation.EventTypeElevationApproved
				})).Return(nil)
				
				mockNotifier.On("NotifyElevationRequest", mock.Anything, mock.Anything).
					Return(nil)
				mockNotifier.On("NotifyElevationApproval", mock.Anything, mock.Anything).
					Return(nil)
			} else {
				if tc.operationRisk == "critical" && !tc.mfaStatus.MFAEnabled {
					mockLogger.On("LogElevationEvent", mock.Anything, mock.MatchedBy(func(event *elevation.ElevationEvent) bool {
						return event.EventType == elevation.EventTypeElevationRequested ||
							   event.EventType == elevation.EventTypeElevationDenied
					})).Return(nil)
				}
			}
			
			// Criar o gerenciador de elevação de privilégios com os mocks
			elevationManager := elevation.NewPrivilegeElevationManager(
				mockApprover,
				mockLogger,
				mockNotifier,
			)
			
			// Configurar a integração MFA
			elevationManager.ConfigureMFAProvider(mockMFA)
			
			// Configurar a política de MFA com base no nível de risco
			elevationManager.ConfigureMFAPolicy(&elevation.MFAPolicy{
				RequireMFAForRiskLevels: map[string]bool{
					"low":      false,
					"medium":   true,
					"high":     true,
					"critical": true,
				},
				MFAChallengeType:      "elevation",
				MFAVerificationExpiry: 48 * time.Hour,
			})
			
			// Definir o nível de risco da operação no contexto
			elevationCtx := elevation.WithRiskLevel(testCtx, tc.operationRisk)
			
			// Testar solicitação de elevação com verificação MFA
			mfaToken := "123456" // Token MFA fictício
			result, err := elevationManager.RequestElevationWithMFA(elevationCtx, elevationRequest, mfaToken)
			
			if tc.expectElevationSuccess {
				require.NoError(t, err, "Solicitação de elevação com MFA não deveria falhar")
				assert.NotNil(t, result, "Resultado da elevação não deveria ser nulo")
				assert.NotEmpty(t, result.ElevationToken, "Token de elevação deveria ser gerado")
			} else {
				require.Error(t, err, "Solicitação de elevação com MFA deveria falhar")
				assert.Contains(t, err.Error(), tc.expectError, "Mensagem de erro incorreta")
				assert.Nil(t, result, "Resultado deveria ser nulo em caso de erro")
			}
			
			// Verificar chamadas aos mocks
			mockMFA.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
			
			// Registrar conclusão do teste
			obs.RecordTestEnd(testCtx, tc.name, err == nil == tc.expectElevationSuccess, time.Since(time.Now()))
		})
	}
}

// TestElevationMFACompliance testa a conformidade do MFA com regulamentos específicos de mercados
func TestElevationMFACompliance(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("privilege_elevation_mfa_compliance_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Configurar mocks
	mockMFA := new(MockMFAProvider)
	mockApprover := new(MockElevationApprover)
	mockLogger := new(MockAuditLogger)
	mockNotifier := new(MockNotifier)
	
	// Criar o gerenciador de elevação de privilégios com os mocks
	elevationManager := elevation.NewPrivilegeElevationManager(
		mockApprover,
		mockLogger,
		mockNotifier,
	)
	
	// Configurar a integração MFA
	elevationManager.ConfigureMFAProvider(mockMFA)
	
	// Configurar políticas MFA específicas por mercado
	marketPolicies := map[string]*elevation.MFAPolicy{
		"angola": {
			RequireMFAForRiskLevels: map[string]bool{
				"low":      false,
				"medium":   true,
				"high":     true,
				"critical": true,
			},
			MFAChallengeType:      "elevation",
			MFAVerificationExpiry: 24 * time.Hour, // Mais restritivo - 24 horas
			EnforceMFASetup:       true,           // Exige configuração de MFA
		},
		"brazil": {
			RequireMFAForRiskLevels: map[string]bool{
				"low":      false,
				"medium":   false, // Menos restritivo
				"high":     true,
				"critical": true,
			},
			MFAChallengeType:      "elevation",
			MFAVerificationExpiry: 48 * time.Hour, // 48 horas
			EnforceMFASetup:       true,
		},
		"global": {
			RequireMFAForRiskLevels: map[string]bool{
				"low":      false,
				"medium":   true,
				"high":     true,
				"critical": true,
			},
			MFAChallengeType:      "elevation",
			MFAVerificationExpiry: 72 * time.Hour, // 72 horas
			EnforceMFASetup:       true,
		},
	}
	
	elevationManager.ConfigureMarketSpecificMFAPolicies(marketPolicies)
	
	// Configurar logger para todos os eventos
	mockLogger.On("LogSecurityEvent", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	
	// Casos de teste para diferentes regulamentos de mercado
	testCases := []struct {
		name              string
		userID            string
		market            string
		operationRisk     string
		lastMFAVerification time.Duration // Quanto tempo atrás o MFA foi verificado
		expectMFARequired bool
	}{
		{
			name:              "Angola - Operação média com MFA verificado há 20 horas",
			userID:            "user:manager:301",
			market:            "angola",
			operationRisk:     "medium",
			lastMFAVerification: -20 * time.Hour,
			expectMFARequired: false, // Dentro do limite de 24h
		},
		{
			name:              "Angola - Operação média com MFA verificado há 30 horas",
			userID:            "user:manager:302",
			market:            "angola",
			operationRisk:     "medium",
			lastMFAVerification: -30 * time.Hour,
			expectMFARequired: true, // Fora do limite de 24h
		},
		{
			name:              "Brasil - Operação média com MFA verificado há 30 horas",
			userID:            "user:manager:303",
			market:            "brazil",
			operationRisk:     "medium",
			lastMFAVerification: -30 * time.Hour,
			expectMFARequired: false, // Não requer MFA para risco médio no Brasil
		},
		{
			name:              "Brasil - Operação alta com MFA verificado há 40 horas",
			userID:            "user:manager:304",
			market:            "brazil",
			operationRisk:     "high",
			lastMFAVerification: -40 * time.Hour,
			expectMFARequired: true, // Dentro do limite de 48h mas requer para alto risco
		},
		{
			name:              "Global - Operação média com MFA verificado há 50 horas",
			userID:            "user:manager:305",
			market:            "global",
			operationRisk:     "medium",
			lastMFAVerification: -50 * time.Hour,
			expectMFARequired: true, // Fora do limite de 72h para global
		},
		{
			name:              "Global - Operação média com MFA verificado há 60 horas",
			userID:            "user:manager:306",
			market:            "global",
			operationRisk:     "medium",
			lastMFAVerification: -60 * time.Hour,
			expectMFARequired: true, // Dentro do limite mas requer para médio risco
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, tc.name)
			
			// Configurar status de MFA do usuário
			baseTime := time.Now()
			userMFAStatus := &mfa.UserMFAStatus{
				UserID:             tc.userID,
				MFAEnabled:         true,
				PrimaryMethod:      "totp",
				BackupMethodsSetup: []string{"sms"},
				LastVerified:       baseTime.Add(tc.lastMFAVerification),
			}
			
			// Configurar mock MFA
			mockMFA.On("GetUserMFAStatus", mock.Anything, tc.userID).
				Return(userMFAStatus, nil).Once()
			
			// Configurar contexto
			reqCtx := context.Background()
			reqCtx = elevation.WithMarket(reqCtx, tc.market)
			reqCtx = elevation.WithRiskLevel(reqCtx, tc.operationRisk)
			
			// Verificar se MFA é necessário
			needsMFA, reason, err := elevationManager.CheckMFARequirement(reqCtx, tc.userID)
			require.NoError(t, err, "Verificação de requisito MFA não deveria falhar")
			
			assert.Equal(t, tc.expectMFARequired, needsMFA, "Requisito de MFA incorreto para %s", tc.name)
			if needsMFA {
				assert.NotEmpty(t, reason, "Razão para requisito MFA deveria ser fornecida")
			}
			
			// Registrar conclusão do teste
			obs.RecordTestEnd(testCtx, tc.name, err == nil, time.Since(time.Now()))
		})
	}
	
	// Verificar chamadas aos mocks
	mockMFA.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestElevationMFAWithPrivacyRegulations testa a conformidade com regulamentos de privacidade
func TestElevationMFAWithPrivacyRegulations(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("privilege_elevation_mfa_privacy_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Testar conformidade com diferentes regulamentos de privacidade por mercado
	// (GDPR para mercado global, LGPD para Brasil, etc.)
	
	// Criar elevação para acesso a dados sensíveis
	elevationRequest := &elevation.ElevationRequest{
		UserID:          "user:data_analyst:501",
		TenantID:        "tenant_global_1",
		Justification:   "Análise de dados para relatório de compliance",
		RequestedRoles:  []string{"data_analyst"},
		RequestedScopes: []string{"data:pii:read", "data:financial:read"},
		Duration:        2 * time.Hour,
		EmergencyAccess: false,
		Market:          "global",
		BusinessUnit:    "analytics",
		DataCategories:  []string{"pii", "financial"}, // Categorias de dados sensíveis
	}
	
	// Configurar mocks
	mockMFA := new(MockMFAProvider)
	mockApprover := new(MockElevationApprover)
	mockLogger := new(MockAuditLogger)
	mockNotifier := new(MockNotifier)
	
	// Status de MFA do usuário
	userMFAStatus := &mfa.UserMFAStatus{
		UserID:             elevationRequest.UserID,
		MFAEnabled:         true,
		PrimaryMethod:      "totp",
		BackupMethodsSetup: []string{"sms"},
		LastVerified:       time.Now().Add(-12 * time.Hour),
	}
	
	mockMFA.On("GetUserMFAStatus", mock.Anything, elevationRequest.UserID).
		Return(userMFAStatus, nil)
		
	mockMFA.On("StartMFAChallenge", mock.Anything, elevationRequest.UserID, "elevation").
		Return("challenge-privacy-001", nil)
		
	mockMFA.On("VerifyMFAToken", mock.Anything, elevationRequest.UserID, "challenge-privacy-001", mock.Anything).
		Return(true, nil)
	
	mockLogger.On("LogSecurityEvent", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
		
	mockLogger.On("LogElevationEvent", mock.Anything, mock.Anything).
		Return(nil)
	
	baseTime := time.Now()
	mockApprover.On("ApproveElevation", mock.Anything, mock.MatchedBy(func(req *elevation.ElevationRequest) bool {
		return req.UserID == elevationRequest.UserID
	})).Return(&elevation.ElevationApproval{
		ElevationID:      "elev-privacy-001",
		UserID:           elevationRequest.UserID,
		ApprovedBy:       "user:compliance:606",
		ApprovalTime:     baseTime,
		ExpirationTime:   baseTime.Add(2 * time.Hour),
		ElevatedRoles:    elevationRequest.RequestedRoles,
		ElevatedScopes:   elevationRequest.RequestedScopes,
		ApprovalEvidence: "compliance_report:2025:Q1",
		AuditMetadata: map[string]interface{}{
			"data_categories":     elevationRequest.DataCategories,
			"privacy_regulations": []string{"GDPR", "LGPD"},
			"purpose_recorded":    "Análise de dados para relatório de compliance",
			"retention_period":    "30 dias",
		},
	}, nil)
	
	mockNotifier.On("NotifyElevationRequest", mock.Anything, mock.Anything).
		Return(nil)
		
	mockNotifier.On("NotifyElevationApproval", mock.Anything, mock.Anything).
		Return(nil)
	
	// Criar o gerenciador de elevação de privilégios com os mocks
	elevationManager := elevation.NewPrivilegeElevationManager(
		mockApprover,
		mockLogger,
		mockNotifier,
	)
	
	// Configurar a integração MFA
	elevationManager.ConfigureMFAProvider(mockMFA)
	
	// Configurar política de elevação para dados sensíveis
	elevationManager.ConfigurePrivacyRequirements(map[string][]string{
		"pii": {
			"mfa_required",
			"purpose_required", 
			"audit_detailed",
			"retention_limited"
		},
		"financial": {
			"mfa_required",
			"approval_required",
			"audit_detailed"
		},
	})
	
	// Definir o contexto com informações de privacidade
	ctx := context.Background()
	privacyCtx := elevation.WithPrivacyContext(ctx, map[string]interface{}{
		"regulations":     []string{"GDPR", "LGPD"},
		"purpose":         "Análise de dados para relatório de compliance",
		"retention_days":  30,
		"data_categories": []string{"pii", "financial"},
	})
	
	// Testar solicitação de elevação com requisitos de privacidade
	mfaToken := "123456" // Token MFA fictício
	result, err := elevationManager.RequestElevationWithMFA(privacyCtx, elevationRequest, mfaToken)
	require.NoError(t, err, "Solicitação de elevação com requisitos de privacidade não deveria falhar")
	assert.NotNil(t, result, "Resultado da elevação não deveria ser nulo")
	
	// Verificar que os metadados de privacidade foram incluídos na elevação
	assert.Contains(t, result.AuditMetadata, "privacy_regulations")
	assert.Contains(t, result.AuditMetadata, "purpose_recorded")
	assert.Contains(t, result.AuditMetadata, "retention_period")
	
	// Verificar chamadas aos mocks
	mockMFA.AssertExpectations(t)
	mockApprover.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	mockNotifier.AssertExpectations(t)
}