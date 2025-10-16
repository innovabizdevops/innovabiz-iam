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

// MockElevationApprover é um mock do aprovador de elevações
type MockElevationApprover struct {
	mock.Mock
}

func (m *MockElevationApprover) ApproveElevation(ctx context.Context, request *elevation.ElevationRequest) (*elevation.ElevationApproval, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*elevation.ElevationApproval), args.Error(1)
}

// MockAuditLogger é um mock do logger de auditoria
type MockAuditLogger struct {
	mock.Mock
}

func (m *MockAuditLogger) LogElevationEvent(ctx context.Context, event *elevation.ElevationEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockAuditLogger) LogSecurityEvent(ctx context.Context, eventType string, details map[string]interface{}) error {
	args := m.Called(ctx, eventType, details)
	return args.Error(0)
}

// MockNotifier é um mock do notificador de elevações
type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) NotifyElevationRequest(ctx context.Context, request *elevation.ElevationRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockNotifier) NotifyElevationApproval(ctx context.Context, approval *elevation.ElevationApproval) error {
	args := m.Called(ctx, approval)
	return args.Error(0)
}

func (m *MockNotifier) NotifyElevationExpiration(ctx context.Context, elevationID string, userID string) error {
	args := m.Called(ctx, elevationID, userID)
	return args.Error(0)
}

// TestPrivilegeElevationManager_RequestElevation testa o fluxo de solicitação de elevação de privilégios
func TestPrivilegeElevationManager_RequestElevation(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("privilege_elevation_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Casos de teste para diferentes cenários de solicitação de elevação
	testCases := []struct {
		name              string
		request           *elevation.ElevationRequest
		mockApproval      *elevation.ElevationApproval
		mockApprovalError error
		expectSuccess     bool
		expectError       string
	}{
		{
			name: "Solicitação válida com aprovação automática",
			request: &elevation.ElevationRequest{
				UserID:          "user:operator:123",
				TenantID:        "tenant_angola_1",
				Justification:   "Incidente de produção #INC-2025-42",
				RequestedRoles:  []string{"admin"},
				RequestedScopes: []string{"k8s:production:pods:delete"},
				Duration:        30 * time.Minute,
				EmergencyAccess: true, // Acesso emergencial
				Market:          "angola",
				BusinessUnit:    "operations",
			},
			mockApproval: &elevation.ElevationApproval{
				ElevationID:      "elev-2025-001",
				UserID:           "user:operator:123",
				ApprovedBy:       "system:emergency:auto",
				ApprovalTime:     time.Now(),
				ExpirationTime:   time.Now().Add(30 * time.Minute),
				ElevatedRoles:    []string{"admin"},
				ElevatedScopes:   []string{"k8s:production:pods:delete"},
				ApprovalEvidence: "emergency_auto_approval:INC-2025-42",
				AuditMetadata: map[string]interface{}{
					"request_ip":      "192.168.1.100",
					"emergency_access": true,
				},
			},
			mockApprovalError: nil,
			expectSuccess:     true,
			expectError:       "",
		},
		{
			name: "Solicitação com duração excessiva",
			request: &elevation.ElevationRequest{
				UserID:          "user:developer:456",
				TenantID:        "tenant_angola_1",
				Justification:   "Acesso prolongado para manutenção",
				RequestedRoles:  []string{"system_admin"},
				RequestedScopes: []string{"system:config:write"},
				Duration:        24 * time.Hour, // 24 horas, muito tempo
				EmergencyAccess: false,
				Market:          "angola",
				BusinessUnit:    "technology",
			},
			mockApproval:      nil,
			mockApprovalError: elevation.ErrElevationDurationExceeded,
			expectSuccess:     false,
			expectError:       "duração excede o máximo permitido",
		},
		{
			name: "Solicitação com escopo não permitido",
			request: &elevation.ElevationRequest{
				UserID:          "user:developer:456",
				TenantID:        "tenant_angola_1",
				Justification:   "Necessidade de acesso a logs",
				RequestedRoles:  []string{"security_analyst"},
				RequestedScopes: []string{"security:logs:read", "security:keys:access"}, // Escopo sensível
				Duration:        30 * time.Minute,
				EmergencyAccess: false,
				Market:          "angola",
				BusinessUnit:    "technology",
			},
			mockApproval:      nil,
			mockApprovalError: elevation.ErrForbiddenElevationScope,
			expectSuccess:     false,
			expectError:       "escopo de elevação não permitido",
		},
		{
			name: "Solicitação válida com fluxo de aprovação normal",
			request: &elevation.ElevationRequest{
				UserID:          "user:developer:456",
				TenantID:        "tenant_angola_1",
				Justification:   "Implementação de feature #FEAT-2025-10",
				RequestedRoles:  []string{"deployer"},
				RequestedScopes: []string{"deployment:production:deploy"},
				Duration:        60 * time.Minute,
				EmergencyAccess: false, // Não emergencial, requer aprovação
				Market:          "angola",
				BusinessUnit:    "technology",
				ApproverID:      "user:manager:789", // Solicitação direcionada
			},
			mockApproval: &elevation.ElevationApproval{
				ElevationID:      "elev-2025-002",
				UserID:           "user:developer:456",
				ApprovedBy:       "user:manager:789",
				ApprovalTime:     time.Now(),
				ExpirationTime:   time.Now().Add(60 * time.Minute),
				ElevatedRoles:    []string{"deployer"},
				ElevatedScopes:   []string{"deployment:production:deploy"},
				ApprovalEvidence: "ticket:FEAT-2025-10",
				AuditMetadata: map[string]interface{}{
					"request_ip":     "192.168.1.101",
					"approval_notes": "Aprovado para implementação da feature",
				},
			},
			mockApprovalError: nil,
			expectSuccess:     true,
			expectError:       "",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, tc.name)
			
			// Configurar mocks
			mockApprover := new(MockElevationApprover)
			mockApprover.On("ApproveElevation", mock.Anything, tc.request).
				Return(tc.mockApproval, tc.mockApprovalError)
				
			mockLogger := new(MockAuditLogger)
			if tc.expectSuccess {
				mockLogger.On("LogElevationEvent", mock.Anything, mock.MatchedBy(func(event *elevation.ElevationEvent) bool {
					return event.EventType == elevation.EventTypeElevationRequested ||
					       event.EventType == elevation.EventTypeElevationApproved
				})).Return(nil)
				
				mockLogger.On("LogSecurityEvent", mock.Anything, "privilege_elevation_approved", mock.Anything).
					Return(nil)
			} else {
				mockLogger.On("LogElevationEvent", mock.Anything, mock.MatchedBy(func(event *elevation.ElevationEvent) bool {
					return event.EventType == elevation.EventTypeElevationRequested ||
					       event.EventType == elevation.EventTypeElevationDenied
				})).Return(nil)
				
				mockLogger.On("LogSecurityEvent", mock.Anything, "privilege_elevation_denied", mock.Anything).
					Return(nil)
			}
			
			mockNotifier := new(MockNotifier)
			mockNotifier.On("NotifyElevationRequest", mock.Anything, tc.request).
				Return(nil)
				
			if tc.expectSuccess {
				mockNotifier.On("NotifyElevationApproval", mock.Anything, tc.mockApproval).
					Return(nil)
			}
			
			// Criar o gerenciador de elevação de privilégios com os mocks
			elevationManager := elevation.NewPrivilegeElevationManager(
				mockApprover,
				mockLogger,
				mockNotifier,
			)			
			// Executar solicitação de elevação
			result, err := elevationManager.RequestElevation(testCtx, tc.request)
			
			// Verificar resultados
			if tc.expectSuccess {
				require.NoError(t, err, "Solicitação de elevação não deveria falhar")
				assert.NotNil(t, result, "Resultado não deveria ser nulo")
				assert.Equal(t, tc.mockApproval.ElevationID, result.ElevationID, "ID de elevação incorreto")
				assert.Equal(t, tc.mockApproval.UserID, result.UserID, "UserID incorreto")
				assert.Equal(t, tc.mockApproval.ElevatedRoles, result.ElevatedRoles, "Papéis elevados incorretos")
				assert.Equal(t, tc.mockApproval.ElevatedScopes, result.ElevatedScopes, "Escopos elevados incorretos")
				
				// Verificar que o token de elevação foi gerado
				assert.NotEmpty(t, result.ElevationToken, "Token de elevação deveria ser gerado")
			} else {
				require.Error(t, err, "Solicitação de elevação deveria falhar")
				assert.Contains(t, err.Error(), tc.expectError, "Mensagem de erro incorreta")
				assert.Nil(t, result, "Resultado deveria ser nulo em caso de erro")
			}
			
			// Verificar chamadas aos mocks
			mockApprover.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
			mockNotifier.AssertExpectations(t)
			
			// Registrar conclusão do teste
			obs.RecordTestEnd(testCtx, tc.name, err == nil == tc.expectSuccess, time.Since(time.Now()))
		})
	}
}

// TestPrivilegeElevationManager_VerifyElevation testa a verificação de elevações de privilégios
func TestPrivilegeElevationManager_VerifyElevation(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("privilege_elevation_verify_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Tempo base para os testes
	baseTime := time.Now()
	
	// Casos de teste para verificação de elevação
	testCases := []struct {
		name            string
		elevationToken  string
		elevationRecord *elevation.ElevationRecord
		expectValid     bool
		expectReason    string
	}{
		{
			name:           "Elevação válida e não expirada",
			elevationToken: "valid-elevation-token-123",
			elevationRecord: &elevation.ElevationRecord{
				ElevationID:      "elev-2025-003",
				UserID:           "user:operator:123",
				TenantID:         "tenant_angola_1",
				ApprovedBy:       "user:manager:789",
				ApprovalTime:     baseTime.Add(-15 * time.Minute),
				ExpirationTime:   baseTime.Add(15 * time.Minute), // Expira em 15 minutos
				ElevatedRoles:    []string{"admin"},
				ElevatedScopes:   []string{"k8s:production:pods:delete"},
				ApprovalEvidence: "ticket:INC-2025-42",
				Status:           elevation.StatusActive,
				Market:           "angola",
				BusinessUnit:     "operations",
			},
			expectValid:  true,
			expectReason: "",
		},
		{
			name:           "Elevação expirada",
			elevationToken: "expired-elevation-token-456",
			elevationRecord: &elevation.ElevationRecord{
				ElevationID:      "elev-2025-004",
				UserID:           "user:developer:456",
				TenantID:         "tenant_angola_1",
				ApprovedBy:       "user:manager:789",
				ApprovalTime:     baseTime.Add(-2 * time.Hour),
				ExpirationTime:   baseTime.Add(-1 * time.Hour), // Expirou há 1 hora
				ElevatedRoles:    []string{"deployer"},
				ElevatedScopes:   []string{"deployment:production:deploy"},
				ApprovalEvidence: "ticket:FEAT-2025-10",
				Status:           elevation.StatusExpired,
				Market:           "angola",
				BusinessUnit:     "technology",
			},
			expectValid:  false,
			expectReason: "elevação de privilégios expirada",
		},
		{
			name:           "Elevação revogada",
			elevationToken: "revoked-elevation-token-789",
			elevationRecord: &elevation.ElevationRecord{
				ElevationID:      "elev-2025-005",
				UserID:           "user:security:789",
				TenantID:         "tenant_angola_1",
				ApprovedBy:       "user:manager:101",
				ApprovalTime:     baseTime.Add(-30 * time.Minute),
				ExpirationTime:   baseTime.Add(30 * time.Minute), // Ainda válida
				ElevatedRoles:    []string{"security_admin"},
				ElevatedScopes:   []string{"security:firewall:update"},
				ApprovalEvidence: "ticket:SEC-2025-15",
				Status:           elevation.StatusRevoked,
				RevocationReason: "Acesso não mais necessário",
				RevokedBy:        "user:manager:101",
				RevocationTime:   baseTime.Add(-10 * time.Minute),
				Market:           "angola",
				BusinessUnit:     "security",
			},
			expectValid:  false,
			expectReason: "elevação de privilégios foi revogada",
		},
		{
			name:           "Token de elevação inválido",
			elevationToken: "invalid-elevation-token",
			elevationRecord: nil, // Nenhum registro encontrado
			expectValid:    false,
			expectReason:   "token de elevação inválido ou não encontrado",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, tc.name)
			
			// Configurar mocks
			mockApprover := new(MockElevationApprover)
			mockLogger := new(MockAuditLogger)
			mockNotifier := new(MockNotifier)
			
			// Logger para verificação de elevação
			if tc.elevationRecord != nil {
				mockLogger.On("LogSecurityEvent", mock.Anything, "privilege_elevation_verification", 
					mock.MatchedBy(func(details map[string]interface{}) bool {
						// Verificar se contém campos relevantes
						id, hasID := details["elevation_id"].(string)
						return hasID && id == tc.elevationRecord.ElevationID
					})).Return(nil)
			} else {
				mockLogger.On("LogSecurityEvent", mock.Anything, "invalid_elevation_token", mock.Anything).
					Return(nil)
			}
			
			// Criar o gerenciador de elevação de privilégios com os mocks e configuração em memória
			elevationManager := elevation.NewPrivilegeElevationManager(
				mockApprover,
				mockLogger,
				mockNotifier,
			)
			
			// Se houver um registro de elevação, adicioná-lo à memória do gerenciador de elevação
			if tc.elevationRecord != nil {
				err := elevationManager.AddElevationToStore(tc.elevationToken, tc.elevationRecord)
				require.NoError(t, err, "Falha ao adicionar elevação à memória")
			}
			
			// Executar verificação de elevação
			valid, reason, record, err := elevationManager.VerifyElevation(testCtx, tc.elevationToken)
			
			// Verificar resultados
			require.NoError(t, err, "Verificação de elevação não deveria falhar com erro")
			assert.Equal(t, tc.expectValid, valid, "Validade da elevação incorreta")
			
			if !tc.expectValid {
				assert.Contains(t, reason, tc.expectReason, "Razão de invalidação incorreta")
				assert.Nil(t, record, "Registro deveria ser nulo para elevações inválidas")
			} else {
				assert.NotNil(t, record, "Registro não deveria ser nulo para elevações válidas")
				assert.Equal(t, tc.elevationRecord.ElevationID, record.ElevationID, "ID de elevação incorreto")
				assert.Equal(t, tc.elevationRecord.UserID, record.UserID, "UserID incorreto")
				assert.Equal(t, elevation.StatusActive, record.Status, "Status da elevação incorreto")
			}
			
			// Verificar chamadas aos mocks
			mockLogger.AssertExpectations(t)
			
			// Registrar conclusão do teste
			obs.RecordTestEnd(testCtx, tc.name, err == nil && valid == tc.expectValid, time.Since(time.Now()))
		})
	}
}

// TestPrivilegeElevationManager_RevokeElevation testa a revogação de elevações de privilégios
func TestPrivilegeElevationManager_RevokeElevation(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("privilege_elevation_revoke_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Tempo base para os testes
	baseTime := time.Now()
	
	// Casos de teste para revogação de elevação
	testCases := []struct {
		name            string
		elevationID     string
		revokerID       string
		reason          string
		elevationRecord *elevation.ElevationRecord
		expectSuccess   bool
		expectError     string
	}{
		{
			name:        "Revogação por aprovador original",
			elevationID: "elev-2025-006",
			revokerID:   "user:manager:789", // Mesmo que aprovou
			reason:      "Acesso não mais necessário",
			elevationRecord: &elevation.ElevationRecord{
				ElevationID:      "elev-2025-006",
				UserID:           "user:developer:456",
				TenantID:         "tenant_angola_1",
				ApprovedBy:       "user:manager:789",
				ApprovalTime:     baseTime.Add(-30 * time.Minute),
				ExpirationTime:   baseTime.Add(30 * time.Minute), // Ainda válida
				ElevatedRoles:    []string{"deployer"},
				ElevatedScopes:   []string{"deployment:production:deploy"},
				ApprovalEvidence: "ticket:FEAT-2025-10",
				Status:           elevation.StatusActive,
				Market:           "angola",
				BusinessUnit:     "technology",
			},
			expectSuccess: true,
			expectError:   "",
		},
		{
			name:        "Revogação por usuário não autorizado",
			elevationID: "elev-2025-007",
			revokerID:   "user:developer:101", // Nem aprovador nem admin
			reason:      "Tentativa não autorizada de revogação",
			elevationRecord: &elevation.ElevationRecord{
				ElevationID:      "elev-2025-007",
				UserID:           "user:security:789",
				TenantID:         "tenant_angola_1",
				ApprovedBy:       "user:manager:202",
				ApprovalTime:     baseTime.Add(-20 * time.Minute),
				ExpirationTime:   baseTime.Add(40 * time.Minute),
				ElevatedRoles:    []string{"security_analyst"},
				ElevatedScopes:   []string{"security:logs:read"},
				ApprovalEvidence: "ticket:SEC-2025-20",
				Status:           elevation.StatusActive,
				Market:           "angola",
				BusinessUnit:     "security",
			},
			expectSuccess: false,
			expectError:   "não autorizado a revogar esta elevação",
		},
		{
			name:        "Revogação por administrador de segurança",
			elevationID: "elev-2025-008",
			revokerID:   "user:security_admin:303", // Admin de segurança pode revogar qualquer elevação
			reason:      "Comportamento suspeito detectado",
			elevationRecord: &elevation.ElevationRecord{
				ElevationID:      "elev-2025-008",
				UserID:           "user:developer:404",
				TenantID:         "tenant_angola_1",
				ApprovedBy:       "user:manager:505",
				ApprovalTime:     baseTime.Add(-10 * time.Minute),
				ExpirationTime:   baseTime.Add(50 * time.Minute),
				ElevatedRoles:    []string{"admin"},
				ElevatedScopes:   []string{"system:config:write"},
				ApprovalEvidence: "ticket:TASK-2025-30",
				Status:           elevation.StatusActive,
				Market:           "angola",
				BusinessUnit:     "technology",
			},
			expectSuccess: true,
			expectError:   "",
		},
		{
			name:            "Revogação de elevação inexistente",
			elevationID:     "elev-non-existent",
			revokerID:       "user:manager:789",
			reason:          "Limpeza de elevações",
			elevationRecord: nil, // Nenhum registro encontrado
			expectSuccess:   false,
			expectError:     "elevação não encontrada",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, tc.name)
			
			// Configurar mocks
			mockApprover := new(MockElevationApprover)
			mockLogger := new(MockAuditLogger)
			mockNotifier := new(MockNotifier)
			
			// Logger para eventos de revogação
			if tc.expectSuccess {
				mockLogger.On("LogElevationEvent", mock.Anything, mock.MatchedBy(func(event *elevation.ElevationEvent) bool {
					return event.EventType == elevation.EventTypeElevationRevoked &&
					       event.ElevationID == tc.elevationID
				})).Return(nil)
				
				mockLogger.On("LogSecurityEvent", mock.Anything, "privilege_elevation_revoked", mock.Anything).
					Return(nil)
					
				// Notificação de revogação
				if tc.elevationRecord != nil {
					mockNotifier.On("NotifyElevationExpiration", mock.Anything, tc.elevationID, tc.elevationRecord.UserID).
						Return(nil)
				}
			} else {
				mockLogger.On("LogSecurityEvent", mock.Anything, "privilege_elevation_revocation_failed", mock.Anything).
					Return(nil)
			}
			
			// Criar o gerenciador de elevação de privilégios com os mocks
			elevationManager := elevation.NewPrivilegeElevationManager(
				mockApprover,
				mockLogger,
				mockNotifier,
			)
			
			// Configurar autorizações de revogação
			elevationManager.ConfigureRevocationRules(&elevation.RevocationRules{
				AdminRoles:       []string{"security_admin", "iam_admin"},
				SelfRevocation:   true,
				ApproverCanRevoke: true,
			})
			
			// Se houver um registro de elevação, adicioná-lo à memória do gerenciador
			if tc.elevationRecord != nil {
				// Adicionar ao store usando um token fictício
				err := elevationManager.AddElevationToStore("token-"+tc.elevationID, tc.elevationRecord)
				require.NoError(t, err, "Falha ao adicionar elevação à memória")
				
				// Adicionar ao índice por ID
				elevationManager.AddElevationToIndex(tc.elevationID, "token-"+tc.elevationID)
			}
			
			// Executar revogação de elevação
			err = elevationManager.RevokeElevation(testCtx, tc.elevationID, tc.revokerID, tc.reason)
			
			// Verificar resultados
			if tc.expectSuccess {
				require.NoError(t, err, "Revogação de elevação não deveria falhar")
				
				// Verificar se a elevação foi realmente revogada
				token, exists := elevationManager.GetTokenByElevationID(tc.elevationID)
				assert.True(t, exists, "Elevação deveria existir após revogação")
				
				// Verificar status usando verificação
				_, _, record, verifyErr := elevationManager.VerifyElevation(testCtx, token)
				require.NoError(t, verifyErr, "Verificação não deveria falhar")
				assert.NotNil(t, record, "Registro não deveria ser nulo")
				assert.Equal(t, elevation.StatusRevoked, record.Status, "Status deveria ser revogado")
				assert.Equal(t, tc.revokerID, record.RevokedBy, "ID do revogador incorreto")
				assert.Equal(t, tc.reason, record.RevocationReason, "Razão de revogação incorreta")
			} else {
				require.Error(t, err, "Revogação de elevação deveria falhar")
				assert.Contains(t, err.Error(), tc.expectError, "Mensagem de erro incorreta")
			}
			
			// Verificar chamadas aos mocks
			mockLogger.AssertExpectations(t)
			mockNotifier.AssertExpectations(t)
			
			// Registrar conclusão do teste
			obs.RecordTestEnd(testCtx, tc.name, err == nil == tc.expectSuccess, time.Since(time.Now()))
		})
	}
}