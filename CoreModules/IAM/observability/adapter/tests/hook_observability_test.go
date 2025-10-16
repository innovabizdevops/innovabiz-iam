// Package adapter_test implementa testes unitários para o adaptador de observabilidade
// dos hooks MCP-IAM da plataforma INNOVABIZ, validando integração de métricas,
// tracing e logging em conformidade com padrões e frameworks internacionais.
//
// Conformidade: ISO/IEC 29119, ISO 9001, ISO 27001, COBIT 2019, TOGAF 10.0
// Frameworks de Teste: TDD, BDD, AAA (Arrange-Act-Assert), FIRST
package adapter_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/innovabiz/iam/constants"
	"github.com/innovabiz/iam/observability/adapter"
	"github.com/innovabiz/iam/observability/logging"
	mock_logging "github.com/innovabiz/iam/observability/logging/mocks"
	mock_metrics "github.com/innovabiz/iam/observability/metrics/mocks"
	mock_tracing "github.com/innovabiz/iam/observability/tracing/mocks"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TestAdapter é uma estrutura de teste que contém mocks para os componentes
// de observabilidade usados pelo adaptador.
type TestAdapter struct {
	ctrl          *gomock.Controller
	mockMetrics   *mock_metrics.MockHookMetricsInterface
	mockTracer    *mock_tracing.MockHookTracerInterface
	mockLogger    *mock_logging.MockHookLoggerInterface
	adapter       *adapter.HookObservability
	ctx           context.Context
	registry      *prometheus.Registry
	testMarketCtx adapter.MarketContext
}

// setupTestAdapter prepara um ambiente de teste com mocks para todos os componentes.
func setupTestAdapter(t *testing.T) *TestAdapter {
	ctrl := gomock.NewController(t)
	
	mockMetrics := mock_metrics.NewMockHookMetricsInterface(ctrl)
	mockTracer := mock_tracing.NewMockHookTracerInterface(ctrl)
	mockLogger := mock_logging.NewMockHookLoggerInterface(ctrl)
	
	registry := prometheus.NewRegistry()
	
	adapterConfig := adapter.Config{
		Environment:           constants.EnvTest,
		ServiceName:           "mcp-iam-hook-test",
		OTLPEndpoint:          "localhost:4317",
		MetricsPort:           9090,
		ComplianceLogsPath:    "/tmp/compliance",
		EnableComplianceAudit: true,
		StructuredLogging:     true,
		LogLevel:              "info",
	}
	
	// Criar um contexto de mercado para testes
	testMarketCtx := adapter.MarketContext{
		Market:                constants.MarketAngola,
		TenantType:            constants.TenantFinancial,
		HookType:              constants.HookTypePrivilegeElevation,
		ComplianceLevel:       constants.ComplianceStrict,
		ApplicableRegulations: []string{"BNA", "LGPD"},
	}
	
	// Criar adaptador com os mocks
	mockAdapter := &adapter.HookObservability{
		Metrics:     mockMetrics,
		Tracer:      mockTracer,
		Logger:      mockLogger,
		Env:         constants.EnvTest,
		ServiceName: "mcp-iam-hook-test",
	}
	
	return &TestAdapter{
		ctrl:          ctrl,
		mockMetrics:   mockMetrics,
		mockTracer:    mockTracer,
		mockLogger:    mockLogger,
		adapter:       mockAdapter,
		ctx:           context.Background(),
		registry:      registry,
		testMarketCtx: testMarketCtx,
	}
}

// TestObserveHookOperation_Success testa o fluxo de sucesso na observação de operações de hook
func TestObserveHookOperation_Success(t *testing.T) {
	// Arrange
	ta := setupTestAdapter(t)
	defer ta.ctrl.Finish()
	
	userId := "user123"
	operation := constants.OperationValidateScope
	description := "Validação de escopo 'admin'"
	
	// Configurar expectativas dos mocks
	ta.mockMetrics.EXPECT().
		RecordElevationRequest(gomock.Any(), ta.testMarketCtx.Market, ta.testMarketCtx.TenantType, ta.testMarketCtx.HookType).
		Times(1)
	
	ta.mockMetrics.EXPECT().
		RecordComplianceCheck(gomock.Any(), ta.testMarketCtx.Market, ta.testMarketCtx.ComplianceLevel, ta.testMarketCtx.HookType).
		Times(1)
	
	ta.mockLogger.EXPECT().
		WithContext(gomock.Any()).
		Return(zap.NewNop()).
		Times(1)
	
	mockSpan := mock_tracing.NewMockSpan(ta.ctrl)
	mockTracer := mock_tracing.NewMockTracer(ta.ctrl)
	
	ta.mockTracer.EXPECT().
		TraceHookOperation(
			gomock.Any(),
			ta.testMarketCtx.HookType,
			ta.testMarketCtx.Market,
			ta.testMarketCtx.TenantType,
			operation,
			gomock.Any(),
			gomock.Any(),
		).DoAndReturn(func(ctx context.Context, hookType, market, tenantType, op string, attrs []attribute.KeyValue, fn func(context.Context) error) error {
			return fn(ctx)
		}).Times(1)
	
	ta.mockLogger.EXPECT().
		LogAuditEvent(
			gomock.Any(),
			ta.testMarketCtx.Market,
			ta.testMarketCtx.TenantType,
			ta.testMarketCtx.HookType,
			gomock.Any(),
			userId,
			"elevation_approved",
			gomock.Any(),
		).Times(1)
	
	// Act - Executar a operação de hook com sucesso
	operationCalled := false
	err := ta.adapter.ObserveHookOperation(
		ta.ctx,
		ta.testMarketCtx,
		operation,
		userId,
		description,
		[]attribute.KeyValue{
			attribute.String("scope", "admin"),
		},
		func(ctx context.Context) error {
			operationCalled = true
			return nil
		},
	)
	
	// Assert
	assert.NoError(t, err)
	assert.True(t, operationCalled, "A função de operação deveria ter sido chamada")
}

// TestObserveHookOperation_Error testa o fluxo de erro na observação de operações de hook
func TestObserveHookOperation_Error(t *testing.T) {
	// Arrange
	ta := setupTestAdapter(t)
	defer ta.ctrl.Finish()
	
	userId := "user123"
	operation := constants.OperationValidateScope
	description := "Validação de escopo 'admin'"
	expectedError := errors.New("acesso negado: permissões insuficientes")
	
	// Configurar expectativas dos mocks
	ta.mockMetrics.EXPECT().
		RecordElevationRequest(gomock.Any(), ta.testMarketCtx.Market, ta.testMarketCtx.TenantType, ta.testMarketCtx.HookType).
		Times(1)
	
	ta.mockMetrics.EXPECT().
		RecordComplianceCheck(gomock.Any(), ta.testMarketCtx.Market, ta.testMarketCtx.ComplianceLevel, ta.testMarketCtx.HookType).
		Times(1)
	
	ta.mockLogger.EXPECT().
		WithContext(gomock.Any()).
		Return(zap.NewNop()).
		Times(1)
	
	ta.mockMetrics.EXPECT().
		ElevationRequestsRejectedTotal.
		WithLabelValues(
			constants.EnvTest,
			ta.testMarketCtx.Market,
			ta.testMarketCtx.TenantType,
			ta.testMarketCtx.HookType,
			expectedError.Error(),
		).
		Return(prometheus.NewCounter(prometheus.CounterOpts{})).
		Times(1)
	
	ta.mockTracer.EXPECT().
		TraceHookOperation(
			gomock.Any(),
			ta.testMarketCtx.HookType,
			ta.testMarketCtx.Market,
			ta.testMarketCtx.TenantType,
			operation,
			gomock.Any(),
			gomock.Any(),
		).DoAndReturn(func(ctx context.Context, hookType, market, tenantType, op string, attrs []attribute.KeyValue, fn func(context.Context) error) error {
			return expectedError
		}).Times(1)
	
	ta.mockLogger.EXPECT().
		LogHookError(
			gomock.Any(),
			ta.testMarketCtx.Market,
			ta.testMarketCtx.TenantType,
			ta.testMarketCtx.HookType,
			operation,
			userId,
			expectedError,
			gomock.Any(),
		).Times(1)
	
	ta.mockTracer.EXPECT().
		TraceAuditEvent(
			gomock.Any(),
			ta.testMarketCtx.HookType,
			ta.testMarketCtx.Market,
			ta.testMarketCtx.TenantType,
			userId,
			"elevation_rejected",
			gomock.Any(),
		).Times(1)
	
	// Act - Executar a operação de hook com erro
	err := ta.adapter.ObserveHookOperation(
		ta.ctx,
		ta.testMarketCtx,
		operation,
		userId,
		description,
		[]attribute.KeyValue{
			attribute.String("scope", "admin"),
		},
		func(ctx context.Context) error {
			return expectedError
		},
	)
	
	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

// TestObserveValidateScope testa a instrumentação da validação de escopo
func TestObserveValidateScope(t *testing.T) {
	// Arrange
	ta := setupTestAdapter(t)
	defer ta.ctrl.Finish()
	
	userId := "user123"
	scope := "admin:read"
	
	// Configurar expectativas dos mocks para ObserveHookOperation
	// Esses serão chamados indiretamente através de ObserveValidateScope
	ta.mockMetrics.EXPECT().
		RecordElevationRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes()
	
	ta.mockMetrics.EXPECT().
		RecordComplianceCheck(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes()
	
	ta.mockLogger.EXPECT().
		WithContext(gomock.Any()).
		Return(zap.NewNop()).
		AnyTimes()
	
	// Expectativas específicas para validação de escopo
	ta.mockLogger.EXPECT().
		LogComplianceEvent(
			gomock.Any(),
			ta.testMarketCtx.Market,
			ta.testMarketCtx.TenantType,
			ta.testMarketCtx.HookType,
			constants.OperationValidateScope,
			userId,
			gomock.Any(), // qualquer regulação
			gomock.Any(), // status
			gomock.Any(), // detalhes
		).
		Times(len(ta.testMarketCtx.ApplicableRegulations))
	
	ta.mockTracer.EXPECT().
		TraceHookOperation(
			gomock.Any(),
			ta.testMarketCtx.HookType,
			ta.testMarketCtx.Market,
			ta.testMarketCtx.TenantType,
			constants.OperationValidateScope,
			gomock.Any(),
			gomock.Any(),
		).DoAndReturn(func(ctx context.Context, hookType, market, tenantType, op string, attrs []attribute.KeyValue, fn func(context.Context) error) error {
			// Verificar se os atributos contêm o escopo
			hasScope := false
			for _, attr := range attrs {
				if attr.Key == "scope" && attr.Value.AsString() == scope {
					hasScope = true
					break
				}
			}
			assert.True(t, hasScope, "Atributos deveriam incluir o escopo")
			return nil
		}).Times(1)
	
	// Act
	validateCalled := false
	err := ta.adapter.ObserveValidateScope(
		ta.ctx,
		ta.testMarketCtx,
		userId,
		scope,
		func(ctx context.Context) error {
			validateCalled = true
			return nil
		},
	)
	
	// Assert
	assert.NoError(t, err)
	assert.True(t, validateCalled, "A função de validação de escopo deveria ter sido chamada")
}

// TestObserveValidateMFA testa a instrumentação da validação MFA
func TestObserveValidateMFA(t *testing.T) {
	// Arrange
	ta := setupTestAdapter(t)
	defer ta.ctrl.Finish()
	
	userId := "user123"
	mfaLevel := constants.MFALevelHigh
	
	// Configurar expectativas dos mocks para ObserveHookOperation
	ta.mockMetrics.EXPECT().
		RecordElevationRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes()
	
	ta.mockMetrics.EXPECT().
		RecordComplianceCheck(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes()
	
	ta.mockLogger.EXPECT().
		WithContext(gomock.Any()).
		Return(zap.NewNop()).
		AnyTimes()
	
	// Expectativas específicas para validação MFA
	ta.mockTracer.EXPECT().
		TraceHookOperation(
			gomock.Any(),
			ta.testMarketCtx.HookType,
			ta.testMarketCtx.Market,
			ta.testMarketCtx.TenantType,
			constants.OperationValidateMFA,
			gomock.Any(),
			gomock.Any(),
		).DoAndReturn(func(ctx context.Context, hookType, market, tenantType, op string, attrs []attribute.KeyValue, fn func(context.Context) error) error {
			// Verificar se os atributos contêm o nível MFA
			hasMFALevel := false
			for _, attr := range attrs {
				if attr.Key == "mfa_level" && attr.Value.AsString() == mfaLevel {
					hasMFALevel = true
					break
				}
			}
			assert.True(t, hasMFALevel, "Atributos deveriam incluir o nível MFA")
			return nil
		}).Times(1)
	
	ta.mockMetrics.EXPECT().
		RecordMFACheck(
			gomock.Any(),
			ta.testMarketCtx.Market,
			mfaLevel,
			true, // success
			ta.testMarketCtx.HookType,
		).Times(1)
	
	// Act
	validateCalled := false
	err := ta.adapter.ObserveValidateMFA(
		ta.ctx,
		ta.testMarketCtx,
		userId,
		mfaLevel,
		func(ctx context.Context) error {
			validateCalled = true
			return nil
		},
	)
	
	// Assert
	assert.NoError(t, err)
	assert.True(t, validateCalled, "A função de validação MFA deveria ter sido chamada")
}

// TestObserveValidateToken testa a instrumentação da validação de token
func TestObserveValidateToken(t *testing.T) {
	// Arrange
	ta := setupTestAdapter(t)
	defer ta.ctrl.Finish()
	
	userId := "user123"
	tokenId := "token456"
	scope := "admin:write"
	
	// Configurar expectativas dos mocks
	ta.mockMetrics.EXPECT().
		RecordElevationRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes()
	
	ta.mockMetrics.EXPECT().
		RecordComplianceCheck(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes()
	
	ta.mockLogger.EXPECT().
		WithContext(gomock.Any()).
		Return(zap.NewNop()).
		AnyTimes()
	
	ta.mockMetrics.EXPECT().
		RecordTokenOperation(
			gomock.Any(),
			ta.testMarketCtx.Market,
			"validate",
			ta.testMarketCtx.HookType,
		).Times(1)
	
	ta.mockTracer.EXPECT().
		TraceHookOperation(
			gomock.Any(),
			ta.testMarketCtx.HookType,
			ta.testMarketCtx.Market,
			ta.testMarketCtx.TenantType,
			constants.OperationValidateToken,
			gomock.Any(),
			gomock.Any(),
		).DoAndReturn(func(ctx context.Context, hookType, market, tenantType, op string, attrs []attribute.KeyValue, fn func(context.Context) error) error {
			// Verificar se os atributos contêm token e escopo
			hasTokenId := false
			hasScope := false
			for _, attr := range attrs {
				if attr.Key == "token_id" && attr.Value.AsString() == tokenId {
					hasTokenId = true
				}
				if attr.Key == "scope" && attr.Value.AsString() == scope {
					hasScope = true
				}
			}
			assert.True(t, hasTokenId, "Atributos deveriam incluir o token ID")
			assert.True(t, hasScope, "Atributos deveriam incluir o escopo")
			return nil
		}).Times(1)
	
	// Act
	validateCalled := false
	err := ta.adapter.ObserveValidateToken(
		ta.ctx,
		ta.testMarketCtx,
		userId,
		tokenId,
		scope,
		func(ctx context.Context) error {
			validateCalled = true
			return nil
		},
	)
	
	// Assert
	assert.NoError(t, err)
	assert.True(t, validateCalled, "A função de validação de token deveria ter sido chamada")
}

// TestObserveCompleteElevation testa a instrumentação da conclusão de elevação
func TestObserveCompleteElevation(t *testing.T) {
	// Arrange
	ta := setupTestAdapter(t)
	defer ta.ctrl.Finish()
	
	userId := "user123"
	tokenId := "token456"
	scope := "admin:write"
	
	// Configurar expectativas dos mocks
	ta.mockMetrics.EXPECT().
		RecordElevationRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes()
	
	ta.mockMetrics.EXPECT().
		RecordComplianceCheck(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes()
	
	ta.mockLogger.EXPECT().
		WithContext(gomock.Any()).
		Return(zap.NewNop()).
		AnyTimes()
	
	ta.mockMetrics.EXPECT().
		RecordTokenOperation(
			gomock.Any(),
			ta.testMarketCtx.Market,
			"complete_elevation",
			ta.testMarketCtx.HookType,
		).Times(1)
	
	ta.mockTracer.EXPECT().
		TraceHookOperation(
			gomock.Any(),
			ta.testMarketCtx.HookType,
			ta.testMarketCtx.Market,
			ta.testMarketCtx.TenantType,
			constants.OperationCompleteElevation,
			gomock.Any(),
			gomock.Any(),
		).Times(1)
	
	// Act
	completeCalled := false
	err := ta.adapter.ObserveCompleteElevation(
		ta.ctx,
		ta.testMarketCtx,
		userId,
		tokenId,
		scope,
		func(ctx context.Context) error {
			completeCalled = true
			return nil
		},
	)
	
	// Assert
	assert.NoError(t, err)
	assert.True(t, completeCalled, "A função de conclusão de elevação deveria ter sido chamada")
}

// TestTraceAuditEvent testa o registro de eventos de auditoria
func TestTraceAuditEvent(t *testing.T) {
	// Arrange
	ta := setupTestAdapter(t)
	defer ta.ctrl.Finish()
	
	userId := "user123"
	eventType := "elevation_requested"
	eventDetails := "Solicitação de elevação para escopo admin:read"
	
	// Configurar expectativas dos mocks
	ta.mockTracer.EXPECT().
		TraceAuditEvent(
			gomock.Any(),
			ta.testMarketCtx.HookType,
			ta.testMarketCtx.Market,
			ta.testMarketCtx.TenantType,
			userId,
			eventType,
			eventDetails,
		).Times(1)
	
	ta.mockLogger.EXPECT().
		LogAuditEvent(
			gomock.Any(),
			ta.testMarketCtx.Market,
			ta.testMarketCtx.TenantType,
			ta.testMarketCtx.HookType,
			"audit",
			userId,
			eventType,
			eventDetails,
		).Times(1)
	
	// Act
	ta.adapter.TraceAuditEvent(
		ta.ctx,
		ta.testMarketCtx,
		userId,
		eventType,
		eventDetails,
	)
	
	// Não há assert explícito, pois a verificação é feita pelo gomock
}

// TestTraceSecurity testa o registro de eventos de segurança
func TestTraceSecurity(t *testing.T) {
	// Arrange
	ta := setupTestAdapter(t)
	defer ta.ctrl.Finish()
	
	userId := "user123"
	severity := "high"
	eventDetails := "Tentativa de acesso não autorizado"
	operation := constants.OperationValidateScope
	
	// Configurar expectativas dos mocks
	ta.mockLogger.EXPECT().
		LogSecurityEvent(
			gomock.Any(),
			ta.testMarketCtx.Market,
			ta.testMarketCtx.TenantType,
			ta.testMarketCtx.HookType,
			operation,
			userId,
			severity,
			eventDetails,
		).Times(1)
	
	// Act
	ta.adapter.TraceSecurity(
		ta.ctx,
		ta.testMarketCtx,
		userId,
		severity,
		eventDetails,
		operation,
	)
	
	// Não há assert explícito, pois a verificação é feita pelo gomock
}

// TestUpdateActiveElevations testa a atualização do contador de elevações ativas
func TestUpdateActiveElevations(t *testing.T) {
	// Arrange
	ta := setupTestAdapter(t)
	defer ta.ctrl.Finish()
	
	count := 5
	
	// Configurar expectativas dos mocks
	ta.mockMetrics.EXPECT().
		UpdateActiveElevations(
			ta.testMarketCtx.Market,
			ta.testMarketCtx.TenantType,
			ta.testMarketCtx.HookType,
			count,
		).Times(1)
	
	// Act
	ta.adapter.UpdateActiveElevations(
		ta.testMarketCtx,
		count,
	)
	
	// Não há assert explícito, pois a verificação é feita pelo gomock
}

// TestRecordTestCoverage testa o registro da cobertura de testes
func TestRecordTestCoverage(t *testing.T) {
	// Arrange
	ta := setupTestAdapter(t)
	defer ta.ctrl.Finish()
	
	hookType := constants.HookTypePrivilegeElevation
	coverage := 95.5
	
	// Configurar expectativas dos mocks
	ta.mockMetrics.EXPECT().
		RecordTestCoverage(
			hookType,
			coverage,
		).Times(1)
	
	// Act
	ta.adapter.RecordTestCoverage(
		hookType,
		coverage,
	)
	
	// Não há assert explícito, pois a verificação é feita pelo gomock
}

// TestRegisterComplianceMetadata testa o registro de metadados de compliance
func TestRegisterComplianceMetadata(t *testing.T) {
	// Arrange
	ta := setupTestAdapter(t)
	defer ta.ctrl.Finish()
	
	market := constants.MarketAngola
	framework := "BNA"
	requiresDualApproval := true
	mfaLevel := constants.MFALevelHigh
	retentionYears := 7
	
	// Configurar expectativas dos mocks
	ta.mockMetrics.EXPECT().
		RegisterComplianceMetadata(
			market,
			framework,
			requiresDualApproval,
			mfaLevel,
			retentionYears,
		).Times(1)
	
	// Act
	ta.adapter.RegisterComplianceMetadata(
		market,
		framework,
		requiresDualApproval,
		mfaLevel,
		retentionYears,
	)
	
	// Não há assert explícito, pois a verificação é feita pelo gomock
}

// TestNewMarketContext testa a criação de contexto de mercado
func TestNewMarketContext(t *testing.T) {
	// Testes para mercados específicos
	tests := []struct {
		name           string
		market         string
		tenantType     string
		hookType       string
		expectedLevel  string
		expectFrameworks bool
	}{
		{
			name:           "Angola Financial Tenant",
			market:         constants.MarketAngola,
			tenantType:     constants.TenantFinancial,
			hookType:       constants.HookTypePrivilegeElevation,
			expectedLevel:  constants.ComplianceStrict,
			expectFrameworks: true,
		},
		{
			name:           "Brazil Healthcare Tenant",
			market:         constants.MarketBrazil,
			tenantType:     constants.TenantHealthcare,
			hookType:       constants.HookTypeTokenValidation,
			expectedLevel:  constants.ComplianceStrict,
			expectFrameworks: true,
		},
		{
			name:           "Global Retail Tenant",
			market:         constants.MarketGlobal,
			tenantType:     constants.TenantRetail,
			hookType:       constants.HookTypeElevationApproval,
			expectedLevel:  constants.ComplianceStandard,
			expectFrameworks: true,
		},
		{
			name:           "EU Telecom Tenant",
			market:         constants.MarketEU,
			tenantType:     constants.TenantTelecom,
			hookType:       constants.HookTypeAudit,
			expectedLevel:  constants.ComplianceEnhanced,
			expectFrameworks: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			marketCtx := adapter.NewMarketContext(tt.market, tt.tenantType, tt.hookType)
			
			// Assert
			assert.Equal(t, tt.market, marketCtx.Market)
			assert.Equal(t, tt.tenantType, marketCtx.TenantType)
			assert.Equal(t, tt.hookType, marketCtx.HookType)
			assert.Equal(t, tt.expectedLevel, marketCtx.ComplianceLevel)
			
			if tt.expectFrameworks {
				assert.NotEmpty(t, marketCtx.ApplicableRegulations, "Deveria ter regulações aplicáveis")
			}
		})
	}
}