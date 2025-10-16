// Package mocks fornece implementações simuladas dos componentes de métricas
// para uso em testes unitários da plataforma INNOVABIZ.
//
// Conformidades: ISO/IEC 29119, ISO 9001, ISO 27001, COBIT 2019, TOGAF 10.0
// Frameworks de Teste: TDD, BDD, FIRST
package mocks

import (
	"context"

	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
)

// MockHookMetricsInterface é uma interface mockável para HookMetrics
// usada em testes unitários para simular comportamentos do sistema de métricas.
type MockHookMetricsInterface interface {
	// Métodos principais para registro de métricas
	RecordElevationRequest(ctx context.Context, market, tenantType, hookType string)
	RecordElevationRequestRejected(ctx context.Context, market, tenantType, hookType, reason string)
	RecordComplianceCheck(ctx context.Context, market, complianceLevel, hookType string)
	RecordMFACheck(ctx context.Context, market, mfaLevel string, success bool, hookType string)
	RecordTokenOperation(ctx context.Context, market, operation, hookType string)
	UpdateActiveElevations(market, tenantType, hookType string, count int)
	RecordTestCoverage(hookType string, coverage float64)
	RegisterComplianceMetadata(market, framework string, requiresDualApproval bool, mfaLevel string, retentionYears int)
	
	// Acesso aos contadores e medidores Prometheus
	ElevationRequestsRejectedTotal() *prometheus.CounterVec
}

// MockHookMetrics é uma implementação mock de HookMetricsInterface.
type MockHookMetrics struct {
	ctrl                           *gomock.Controller
	elevationRequestsTotal         *prometheus.CounterVec
	elevationRequestsRejectedTotal *prometheus.CounterVec
	complianceChecksTotal          *prometheus.CounterVec
	mfaChecksTotal                 *prometheus.CounterVec
	tokenOperationsTotal           *prometheus.CounterVec
	validationDuration             *prometheus.HistogramVec
	activeElevations               *prometheus.GaugeVec
	testCoveragePercent            *prometheus.GaugeVec
}

// NewMockHookMetricsInterface cria um novo mock da interface de métricas.
func NewMockHookMetricsInterface(ctrl *gomock.Controller) *MockHookMetrics {
	return &MockHookMetrics{ctrl: ctrl}
}

// RecordElevationRequest registra uma solicitação de elevação.
func (m *MockHookMetrics) RecordElevationRequest(ctx context.Context, market, tenantType, hookType string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordElevationRequest", ctx, market, tenantType, hookType)
}

// RecordElevationRequestRejected registra uma solicitação de elevação rejeitada.
func (m *MockHookMetrics) RecordElevationRequestRejected(ctx context.Context, market, tenantType, hookType, reason string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordElevationRequestRejected", ctx, market, tenantType, hookType, reason)
}

// RecordComplianceCheck registra uma verificação de compliance.
func (m *MockHookMetrics) RecordComplianceCheck(ctx context.Context, market, complianceLevel, hookType string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordComplianceCheck", ctx, market, complianceLevel, hookType)
}

// RecordMFACheck registra uma verificação de MFA.
func (m *MockHookMetrics) RecordMFACheck(ctx context.Context, market, mfaLevel string, success bool, hookType string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordMFACheck", ctx, market, mfaLevel, success, hookType)
}

// RecordTokenOperation registra uma operação de token.
func (m *MockHookMetrics) RecordTokenOperation(ctx context.Context, market, operation, hookType string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordTokenOperation", ctx, market, operation, hookType)
}

// UpdateActiveElevations atualiza o contador de elevações ativas.
func (m *MockHookMetrics) UpdateActiveElevations(market, tenantType, hookType string, count int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "UpdateActiveElevations", market, tenantType, hookType, count)
}

// RecordTestCoverage registra a porcentagem de cobertura de testes.
func (m *MockHookMetrics) RecordTestCoverage(hookType string, coverage float64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordTestCoverage", hookType, coverage)
}

// RegisterComplianceMetadata registra metadados de compliance para um mercado.
func (m *MockHookMetrics) RegisterComplianceMetadata(market, framework string, requiresDualApproval bool, mfaLevel string, retentionYears int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RegisterComplianceMetadata", market, framework, requiresDualApproval, mfaLevel, retentionYears)
}

// ElevationRequestsRejectedTotal retorna o contador de rejeições de elevação.
func (m *MockHookMetrics) ElevationRequestsRejectedTotal() *prometheus.CounterVec {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ElevationRequestsRejectedTotal")
	if ret.Get(0) != nil {
		return ret.Get(0).(*prometheus.CounterVec)
	}
	return nil
}