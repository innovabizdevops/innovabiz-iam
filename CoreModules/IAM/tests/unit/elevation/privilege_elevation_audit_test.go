// Package elevation_test contém testes unitários para o componente de elevação de privilégios do MCP-IAM
package elevation_test

import (
	"context"
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	
	"github.com/innovabizdevops/innovabiz-iam/audit"
	"github.com/innovabizdevops/innovabiz-iam/authorization/elevation"
	"github.com/innovabizdevops/innovabiz-iam/observability"
	"github.com/innovabizdevops/innovabiz-iam/tests/testutil"
)

// MockAuditStore é um mock do repositório de auditoria
type MockAuditStore struct {
	mock.Mock
}

func (m *MockAuditStore) StoreAuditEvent(ctx context.Context, event *audit.AuditEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockAuditStore) QueryAuditEvents(ctx context.Context, filter audit.AuditQueryFilter) ([]*audit.AuditEvent, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*audit.AuditEvent), args.Error(1)
}

// MockTracer é um mock do tracer para observabilidade
type MockTracer struct {
	mock.Mock
}

func (m *MockTracer) StartSpan(ctx context.Context, operationName string) (context.Context, observability.Span) {
	args := m.Called(ctx, operationName)
	if args.Get(0) == nil {
		return ctx, nil
	}
	return args.Get(0).(context.Context), args.Get(1).(observability.Span)
}

// MockSpan é um mock de um span de observabilidade
type MockSpan struct {
	mock.Mock
}

func (m *MockSpan) End() {
	m.Called()
}

func (m *MockSpan) SetAttribute(key string, value interface{}) {
	m.Called(key, value)
}

func (m *MockSpan) AddEvent(name string, attributes map[string]interface{}) {
	m.Called(name, attributes)
}

func (m *MockSpan) RecordError(err error) {
	m.Called(err)
}

// TestPrivilegeElevationAudit testa os mecanismos de auditoria para elevação de privilégios
func TestPrivilegeElevationAudit(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("privilege_elevation_audit_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Configurar mocks
	mockApprover := new(MockElevationApprover)
	mockLogger := new(MockAuditLogger)
	mockNotifier := new(MockNotifier)
	mockAuditStore := new(MockAuditStore)
	mockTracer := new(MockTracer)
	mockSpan := new(MockSpan)
	
	// Configurar comportamentos dos mocks
	baseTime := time.Now()
	
	// Criar o gerenciador de elevação de privilégios com os mocks
	elevationManager := elevation.NewPrivilegeElevationManager(
		mockApprover,
		mockLogger,
		mockNotifier,
	)
	
	// Configurar armazenamento de auditoria
	elevationManager.ConfigureAuditStore(mockAuditStore)
	
	// Configurar tracer para observabilidade
	elevationManager.ConfigureTracer(mockTracer)
	
	// Casos de teste para auditoria de elevação
	testCases := []struct {
		name           string
		elevationEvent *elevation.ElevationEvent
		auditEvent     *audit.AuditEvent
		expectSuccess  bool
		expectError    string
	}{
		{
			name: "Auditoria de solicitação de elevação",
			elevationEvent: &elevation.ElevationEvent{
				EventType:    elevation.EventTypeElevationRequested,
				ElevationID:  "elev-audit-001",
				UserID:       "user:operator:123",
				TenantID:     "tenant_angola_1",
				Timestamp:    baseTime,
				Market:       "angola",
				BusinessUnit: "operations",
				Details: map[string]interface{}{
					"requested_roles":  []string{"admin"},
					"requested_scopes": []string{"k8s:production:pods:delete"},
					"justification":    "Incidente de produção #INC-2025-42",
					"source_ip":        "192.168.1.100",
					"user_agent":       "MCP Client/1.0",
				},
			},
			auditEvent: &audit.AuditEvent{
				EventType:    audit.EventTypeAuthorizationElevation,
				EventSubtype: string(elevation.EventTypeElevationRequested),
				UserID:       "user:operator:123",
				TenantID:     "tenant_angola_1",
				Timestamp:    baseTime,
				Market:       "angola",
				BusinessUnit: "operations",
				ResourceID:   "elev-audit-001",
				ResourceType: "elevation",
				Action:       "request",
				Result:       "pending",
				RequestData: map[string]interface{}{
					"requested_roles":  []string{"admin"},
					"requested_scopes": []string{"k8s:production:pods:delete"},
					"justification":    "Incidente de produção #INC-2025-42",
				},
				ClientData: map[string]interface{}{
					"source_ip":  "192.168.1.100",
					"user_agent": "MCP Client/1.0",
				},
				Severity:        "medium",
				ComplianceFlags: []string{"SOX", "ISO27001", "PCI-DSS"},
			},
			expectSuccess: true,
			expectError:   "",
		},
		{
			name: "Auditoria de aprovação de elevação",
			elevationEvent: &elevation.ElevationEvent{
				EventType:    elevation.EventTypeElevationApproved,
				ElevationID:  "elev-audit-002",
				UserID:       "user:operator:123",
				TenantID:     "tenant_angola_1",
				Timestamp:    baseTime,
				Market:       "angola",
				BusinessUnit: "operations",
				Details: map[string]interface{}{
					"approver_id":         "user:manager:456",
					"elevated_roles":      []string{"admin"},
					"elevated_scopes":     []string{"k8s:production:pods:delete"},
					"approval_evidence":   "ticket:INC-2025-42",
					"expiration_time":     baseTime.Add(30 * time.Minute),
					"emergency_approval":  true,
					"mfa_verified":        true,
					"approval_conditions": "Limitado a pods específicos",
				},
			},
			auditEvent: &audit.AuditEvent{
				EventType:    audit.EventTypeAuthorizationElevation,
				EventSubtype: string(elevation.EventTypeElevationApproved),
				UserID:       "user:operator:123",
				TenantID:     "tenant_angola_1",
				Timestamp:    baseTime,
				Market:       "angola",
				BusinessUnit: "operations",
				ResourceID:   "elev-audit-002",
				ResourceType: "elevation",
				Action:       "approve",
				Result:       "success",
				RequestData: map[string]interface{}{
					"elevated_roles":      []string{"admin"},
					"elevated_scopes":     []string{"k8s:production:pods:delete"},
					"approval_evidence":   "ticket:INC-2025-42",
					"expiration_time":     baseTime.Add(30 * time.Minute),
					"emergency_approval":  true,
					"mfa_verified":        true,
					"approval_conditions": "Limitado a pods específicos",
				},
				RelatedActorID:   "user:manager:456",
				RelatedActorType: "approver",
				Severity:         "high",
				ComplianceFlags:  []string{"SOX", "ISO27001", "PCI-DSS"},
			},
			expectSuccess: true,
			expectError:   "",
		},
		{
			name: "Auditoria de uso de elevação",
			elevationEvent: &elevation.ElevationEvent{
				EventType:    elevation.EventTypeElevationUsed,
				ElevationID:  "elev-audit-003",
				UserID:       "user:operator:123",
				TenantID:     "tenant_angola_1",
				Timestamp:    baseTime,
				Market:       "angola",
				BusinessUnit: "operations",
				Details: map[string]interface{}{
					"operation":        "kubectl delete pod nginx-123",
					"resource":         "pod/nginx-123",
					"namespace":        "production",
					"command_id":       "cmd-789",
					"client_ip":        "192.168.1.100",
					"elevated_roles":   []string{"admin"},
					"elevated_scopes":  []string{"k8s:production:pods:delete"},
					"operation_result": "success",
				},
			},
			auditEvent: &audit.AuditEvent{
				EventType:    audit.EventTypeAuthorizationElevation,
				EventSubtype: string(elevation.EventTypeElevationUsed),
				UserID:       "user:operator:123",
				TenantID:     "tenant_angola_1",
				Timestamp:    baseTime,
				Market:       "angola",
				BusinessUnit: "operations",
				ResourceID:   "elev-audit-003",
				ResourceType: "elevation",
				Action:       "use",
				Result:       "success",
				RequestData: map[string]interface{}{
					"operation":        "kubectl delete pod nginx-123",
					"resource":         "pod/nginx-123",
					"namespace":        "production",
					"command_id":       "cmd-789",
					"elevated_roles":   []string{"admin"},
					"elevated_scopes":  []string{"k8s:production:pods:delete"},
					"operation_result": "success",
				},
				ClientData: map[string]interface{}{
					"client_ip": "192.168.1.100",
				},
				Severity:        "high",
				ComplianceFlags: []string{"SOX", "ISO27001", "PCI-DSS"},
				TargetResource:  "pod/nginx-123",
				TargetAction:    "delete",
			},
			expectSuccess: true,
			expectError:   "",
		},
		{
			name: "Auditoria de revogação de elevação",
			elevationEvent: &elevation.ElevationEvent{
				EventType:    elevation.EventTypeElevationRevoked,
				ElevationID:  "elev-audit-004",
				UserID:       "user:operator:123",
				TenantID:     "tenant_angola_1",
				Timestamp:    baseTime,
				Market:       "angola",
				BusinessUnit: "operations",
				Details: map[string]interface{}{
					"revoked_by":        "user:security_admin:789",
					"revocation_reason": "Atividade suspeita detectada",
					"original_expiry":   baseTime.Add(30 * time.Minute),
					"revocation_time":   baseTime.Add(10 * time.Minute),
				},
			},
			auditEvent: &audit.AuditEvent{
				EventType:    audit.EventTypeAuthorizationElevation,
				EventSubtype: string(elevation.EventTypeElevationRevoked),
				UserID:       "user:operator:123",
				TenantID:     "tenant_angola_1",
				Timestamp:    baseTime,
				Market:       "angola",
				BusinessUnit: "operations",
				ResourceID:   "elev-audit-004",
				ResourceType: "elevation",
				Action:       "revoke",
				Result:       "success",
				RequestData: map[string]interface{}{
					"revocation_reason": "Atividade suspeita detectada",
					"original_expiry":   baseTime.Add(30 * time.Minute),
					"revocation_time":   baseTime.Add(10 * time.Minute),
				},
				RelatedActorID:   "user:security_admin:789",
				RelatedActorType: "security_admin",
				Severity:         "high",
				ComplianceFlags:  []string{"SOX", "ISO27001", "PCI-DSS"},
			},
			expectSuccess: true,
			expectError:   "",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, tc.name)
			
			// Configurar o tracer mock para este teste específico
			mockTracer.On("StartSpan", mock.Anything, "AuditElevationEvent").
				Return(testCtx, mockSpan)
				
			mockSpan.On("SetAttribute", "elevation_id", tc.elevationEvent.ElevationID).Return()
			mockSpan.On("SetAttribute", "event_type", tc.elevationEvent.EventType).Return()
			mockSpan.On("SetAttribute", "user_id", tc.elevationEvent.UserID).Return()
			mockSpan.On("SetAttribute", "tenant_id", tc.elevationEvent.TenantID).Return()
			mockSpan.On("End").Return()
			
			// Configurar logger para o evento de elevação
			mockLogger.On("LogElevationEvent", mock.Anything, tc.elevationEvent).
				Return(nil)
			
			// Configurar armazenamento de auditoria
			mockAuditStore.On("StoreAuditEvent", mock.Anything, mock.MatchedBy(func(event *audit.AuditEvent) bool {
				return event.EventSubtype == string(tc.elevationEvent.EventType) &&
				       event.ResourceID == tc.elevationEvent.ElevationID &&
				       event.UserID == tc.elevationEvent.UserID
			})).Return(nil)
			
			// Executar o teste - registrar evento de elevação
			err := elevationManager.AuditElevationEvent(testCtx, tc.elevationEvent)
			
			// Verificar resultados
			if tc.expectSuccess {
				require.NoError(t, err, "Registro de auditoria não deveria falhar")
			} else {
				require.Error(t, err, "Registro de auditoria deveria falhar")
				assert.Contains(t, err.Error(), tc.expectError, "Mensagem de erro incorreta")
			}
			
			// Verificar chamadas aos mocks
			mockLogger.AssertExpectations(t)
			mockAuditStore.AssertExpectations(t)
			mockTracer.AssertExpectations(t)
			mockSpan.AssertExpectations(t)
			
			// Registrar conclusão do teste
			obs.RecordTestEnd(testCtx, tc.name, err == nil, time.Since(time.Now()))
		})
	}
}

// TestPrivilegeElevationAuditReporting testa consultas ao histórico de auditoria para relatórios de compliance
func TestPrivilegeElevationAuditReporting(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("privilege_elevation_audit_reporting_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Configurar mocks
	mockApprover := new(MockElevationApprover)
	mockLogger := new(MockAuditLogger)
	mockNotifier := new(MockNotifier)
	mockAuditStore := new(MockAuditStore)
	mockTracer := new(MockTracer)
	mockSpan := new(MockSpan)
	
	// Configurar o gerenciador de elevação de privilégios com os mocks
	elevationManager := elevation.NewPrivilegeElevationManager(
		mockApprover,
		mockLogger,
		mockNotifier,
	)
	
	// Configurar armazenamento de auditoria
	elevationManager.ConfigureAuditStore(mockAuditStore)
	elevationManager.ConfigureTracer(mockTracer)
	
	// Definir base de tempo para os testes
	baseTime := time.Now()
	
	// Configurar eventos de auditoria de exemplo para retornar nas consultas
	sampleAuditEvents := []*audit.AuditEvent{
		{
			EventType:      audit.EventTypeAuthorizationElevation,
			EventSubtype:   string(elevation.EventTypeElevationRequested),
			UserID:         "user:operator:123",
			TenantID:       "tenant_angola_1",
			Timestamp:      baseTime.Add(-3 * time.Hour),
			Market:         "angola",
			BusinessUnit:   "operations",
			ResourceID:     "elev-audit-005",
			ResourceType:   "elevation",
			Action:         "request",
			Result:         "success",
			Severity:       "medium",
			ComplianceFlags: []string{"SOX", "ISO27001"},
		},
		{
			EventType:      audit.EventTypeAuthorizationElevation,
			EventSubtype:   string(elevation.EventTypeElevationApproved),
			UserID:         "user:operator:123",
			TenantID:       "tenant_angola_1",
			Timestamp:      baseTime.Add(-2 * time.Hour),
			Market:         "angola",
			BusinessUnit:   "operations",
			ResourceID:     "elev-audit-005",
			ResourceType:   "elevation",
			Action:         "approve",
			Result:         "success",
			RelatedActorID: "user:manager:456",
			Severity:       "high",
			ComplianceFlags: []string{"SOX", "ISO27001"},
		},
		{
			EventType:      audit.EventTypeAuthorizationElevation,
			EventSubtype:   string(elevation.EventTypeElevationUsed),
			UserID:         "user:operator:123",
			TenantID:       "tenant_angola_1",
			Timestamp:      baseTime.Add(-1 * time.Hour),
			Market:         "angola",
			BusinessUnit:   "operations",
			ResourceID:     "elev-audit-005",
			ResourceType:   "elevation",
			Action:         "use",
			Result:         "success",
			TargetResource: "pod/nginx-123",
			TargetAction:   "delete",
			Severity:       "high",
			ComplianceFlags: []string{"SOX", "ISO27001"},
		},
	}
	
	// Casos de teste para consulta de auditoria
	testCases := []struct {
		name          string
		queryFilter   audit.AuditQueryFilter
		expectedCount int
	}{
		{
			name: "Consultar por ID de elevação",
			queryFilter: audit.AuditQueryFilter{
				ResourceID:   "elev-audit-005",
				ResourceType: "elevation",
			},
			expectedCount: 3, // Todos os eventos para esta elevação
		},
		{
			name: "Consultar por usuário",
			queryFilter: audit.AuditQueryFilter{
				UserID:       "user:operator:123",
				ResourceType: "elevation",
			},
			expectedCount: 3, // Todos os eventos deste usuário
		},
		{
			name: "Consultar por período de tempo",
			queryFilter: audit.AuditQueryFilter{
				StartTime:    baseTime.Add(-2 * time.Hour),
				EndTime:      baseTime,
				ResourceType: "elevation",
			},
			expectedCount: 2, // Apenas eventos das últimas 2 horas
		},
		{
			name: "Consultar por tipo de evento",
			queryFilter: audit.AuditQueryFilter{
				EventSubtype: string(elevation.EventTypeElevationUsed),
				ResourceType: "elevation",
			},
			expectedCount: 1, // Apenas eventos de uso
		},
		{
			name: "Consultar por mercado",
			queryFilter: audit.AuditQueryFilter{
				Market:       "angola",
				ResourceType: "elevation",
			},
			expectedCount: 3, // Todos os eventos de Angola
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, tc.name)
			
			// Configurar tracer mock
			mockTracer.On("StartSpan", mock.Anything, "QueryElevationAudit").
				Return(testCtx, mockSpan)
				
			mockSpan.On("SetAttribute", "filter", mock.Anything).Return()
			mockSpan.On("End").Return()
			
			// Filtrar eventos que correspondem ao caso de teste
			filteredEvents := make([]*audit.AuditEvent, 0)
			for _, event := range sampleAuditEvents {
				matches := true
				
				if tc.queryFilter.ResourceID != "" && tc.queryFilter.ResourceID != event.ResourceID {
					matches = false
				}
				
				if tc.queryFilter.UserID != "" && tc.queryFilter.UserID != event.UserID {
					matches = false
				}
				
				if tc.queryFilter.EventSubtype != "" && tc.queryFilter.EventSubtype != event.EventSubtype {
					matches = false
				}
				
				if tc.queryFilter.Market != "" && tc.queryFilter.Market != event.Market {
					matches = false
				}
				
				if !tc.queryFilter.StartTime.IsZero() && event.Timestamp.Before(tc.queryFilter.StartTime) {
					matches = false
				}
				
				if !tc.queryFilter.EndTime.IsZero() && event.Timestamp.After(tc.queryFilter.EndTime) {
					matches = false
				}
				
				if matches {
					filteredEvents = append(filteredEvents, event)
				}
			}
			
			// Configurar retorno do mock do armazenamento de auditoria
			mockAuditStore.On("QueryAuditEvents", mock.Anything, tc.queryFilter).
				Return(filteredEvents, nil)
			
			// Executar consulta de auditoria
			events, err := elevationManager.QueryElevationAudit(testCtx, tc.queryFilter)
			require.NoError(t, err, "Consulta de auditoria não deveria falhar")
			
			// Verificar resultados
			assert.Equal(t, tc.expectedCount, len(events), "Número incorreto de eventos retornados")
			
			for i, event := range events {
				assert.Equal(t, audit.EventTypeAuthorizationElevation, event.EventType, "Tipo de evento incorreto")
				assert.Equal(t, "elevation", event.ResourceType, "Tipo de recurso incorreto")
			}
			
			// Verificar chamadas aos mocks
			mockAuditStore.AssertExpectations(t)
			mockTracer.AssertExpectations(t)
			mockSpan.AssertExpectations(t)
			
			// Registrar conclusão do teste
			obs.RecordTestEnd(testCtx, tc.name, err == nil, time.Since(time.Now()))
		})
	}
}

// TestPrivilegeElevationAuditCompliance testa conformidade regulatória nos registros de auditoria
func TestPrivilegeElevationAuditCompliance(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("privilege_elevation_audit_compliance_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Configurar mocks
	mockApprover := new(MockElevationApprover)
	mockLogger := new(MockAuditLogger)
	mockNotifier := new(MockNotifier)
	mockAuditStore := new(MockAuditStore)
	mockTracer := new(MockTracer)
	mockSpan := new(MockSpan)
	
	// Configurar o gerenciador de elevação de privilégios com os mocks
	elevationManager := elevation.NewPrivilegeElevationManager(
		mockApprover,
		mockLogger,
		mockNotifier,
	)
	
	// Configurar armazenamento de auditoria e tracer
	elevationManager.ConfigureAuditStore(mockAuditStore)
	elevationManager.ConfigureTracer(mockTracer)
	
	// Configurar requisitos de auditoria específicos por mercado e regulamento
	elevationManager.ConfigureComplianceRequirements(map[string][]string{
		"angola": {"SOX", "ISO27001", "PCI-DSS", "Angola-CNPD", "GDPR"},
		"brazil": {"SOX", "ISO27001", "PCI-DSS", "LGPD", "Brazil-BACEN"},
		"global": {"SOX", "ISO27001", "PCI-DSS", "GDPR", "HIPAA"},
	})
	
	// Definir base de tempo para os testes
	baseTime := time.Now()
	
	// Criar evento de elevação para teste
	elevationEvent := &elevation.ElevationEvent{
		EventType:    elevation.EventTypeElevationApproved,
		ElevationID:  "elev-compliance-001",
		UserID:       "user:operator:123",
		TenantID:     "tenant_angola_1",
		Timestamp:    baseTime,
		Market:       "angola",
		BusinessUnit: "operations",
		Details: map[string]interface{}{
			"approver_id":        "user:manager:456",
			"elevated_roles":     []string{"admin"},
			"elevated_scopes":    []string{"k8s:production:pods:delete"},
			"approval_evidence":  "ticket:INC-2025-42",
			"expiration_time":    baseTime.Add(30 * time.Minute),
			"emergency_approval": true,
		},
	}
	
	// Configurar tracer mock
	mockTracer.On("StartSpan", mock.Anything, "AuditElevationEvent").
		Return(context.Background(), mockSpan)
		
	mockSpan.On("SetAttribute", mock.Anything, mock.Anything).Return()
	mockSpan.On("End").Return()
	
	// Configurar logger para o evento de elevação
	mockLogger.On("LogElevationEvent", mock.Anything, elevationEvent).
		Return(nil)
	
	// Configurar armazenamento de auditoria para validar campos específicos de compliance
	mockAuditStore.On("StoreAuditEvent", mock.Anything, mock.MatchedBy(func(event *audit.AuditEvent) bool {
		// Verificar se os campos obrigatórios para compliance estão presentes
		if event.EventType != audit.EventTypeAuthorizationElevation {
			return false
		}
		if event.UserID != elevationEvent.UserID {
			return false
		}
		
		// Verificar flags de compliance específicas do mercado Angola
		hasAngolaFlags := false
		for _, flag := range event.ComplianceFlags {
			if flag == "Angola-CNPD" {
				hasAngolaFlags = true
				break
			}
		}
		
		// Verificar campos de retenção e integridade exigidos por regulamentos
		_, hasRetention := event.Metadata["retention_period"]
		_, hasIntegrity := event.Metadata["integrity_hash"]
		_, hasTimestamp := event.Metadata["utc_timestamp"]
		
		return hasAngolaFlags && hasRetention && hasIntegrity && hasTimestamp
	})).Return(nil)
	
	// Executar o teste - registrar evento de elevação com requisitos de compliance
	ctx := context.Background()
	testCtx := elevation.WithMarket(ctx, "angola") // Configurar contexto com mercado Angola
	
	err = elevationManager.AuditElevationEventWithCompliance(testCtx, elevationEvent)
	require.NoError(t, err, "Registro de auditoria com compliance não deveria falhar")
	
	// Verificar chamadas aos mocks
	mockLogger.AssertExpectations(t)
	mockAuditStore.AssertExpectations(t)
	mockTracer.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
}