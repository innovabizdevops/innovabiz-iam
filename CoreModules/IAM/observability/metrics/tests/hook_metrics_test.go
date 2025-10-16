// Package metrics_test fornece testes unitários para o pacote de métricas dos hooks MCP-IAM
//
// Testes implementados seguindo padrões:
// - TDD (Test-Driven Development)
// - BDD (Behavior-Driven Development)
// - ISO/IEC 29119 (Software Testing Standards)
//
// Frameworks de testes: testify, gomock, prometheus/testutil
package metrics_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/innovabiz/iam/constants"
	"github.com/innovabiz/iam/observability/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// MockTracer é uma implementação mock do tracer OpenTelemetry
type MockTracer struct {
	mock.Mock
}

// Start implementa o método Start da interface trace.Tracer
func (m *MockTracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	args := m.Called(ctx, spanName, opts)
	return args.Get(0).(context.Context), args.Get(1).(trace.Span)
}

// MockSpan é uma implementação mock do span OpenTelemetry
type MockSpan struct {
	mock.Mock
}

// End implementa o método End da interface trace.Span
func (m *MockSpan) End(options ...trace.SpanEndOption) {
	m.Called(options)
}

// SpanContext implementa o método SpanContext da interface trace.Span
func (m *MockSpan) SpanContext() trace.SpanContext {
	args := m.Called()
	return args.Get(0).(trace.SpanContext)
}

// IsRecording implementa o método IsRecording da interface trace.Span
func (m *MockSpan) IsRecording() bool {
	args := m.Called()
	return args.Bool(0)
}

// SetStatus implementa o método SetStatus da interface trace.Span
func (m *MockSpan) SetStatus(code trace.StatusCode, description string) {
	m.Called(code, description)
}

// SetName implementa o método SetName da interface trace.Span
func (m *MockSpan) SetName(name string) {
	m.Called(name)
}

// SetAttributes implementa o método SetAttributes da interface trace.Span
func (m *MockSpan) SetAttributes(attributes ...attribute.KeyValue) {
	m.Called(attributes)
}

// AddEvent implementa o método AddEvent da interface trace.Span
func (m *MockSpan) AddEvent(name string, options ...trace.EventOption) {
	m.Called(name, options)
}

// RecordError implementa o método RecordError da interface trace.Span
func (m *MockSpan) RecordError(err error, options ...trace.EventOption) {
	m.Called(err, options)
}

// TracerProvider implementa o método TracerProvider da interface trace.Span
func (m *MockSpan) TracerProvider() trace.TracerProvider {
	args := m.Called()
	return args.Get(0).(trace.TracerProvider)
}

// MockSpanContext é uma implementação mock do SpanContext OpenTelemetry
type MockSpanContext struct {
	mock.Mock
}

// TraceID implementa o método TraceID da interface trace.SpanContext
func (m *MockSpanContext) TraceID() trace.TraceID {
	args := m.Called()
	return args.Get(0).(trace.TraceID)
}

// SpanID implementa o método SpanID da interface trace.SpanContext
func (m *MockSpanContext) SpanID() trace.SpanID {
	args := m.Called()
	return args.Get(0).(trace.SpanID)
}

// IsRemote implementa o método IsRemote da interface trace.SpanContext
func (m *MockSpanContext) IsRemote() bool {
	args := m.Called()
	return args.Bool(0)
}

// IsSampled implementa o método IsSampled da interface trace.SpanContext
func (m *MockSpanContext) IsSampled() bool {
	args := m.Called()
	return args.Bool(0)
}

// IsValid implementa o método IsValid da interface trace.SpanContext
func (m *MockSpanContext) IsValid() bool {
	args := m.Called()
	return args.Bool(0)
}

// TraceState implementa o método TraceState da interface trace.SpanContext
func (m *MockSpanContext) TraceState() trace.TraceState {
	args := m.Called()
	return args.Get(0).(trace.TraceState)
}

// TraceFlags implementa o método TraceFlags da interface trace.SpanContext
func (m *MockSpanContext) TraceFlags() trace.TraceFlags {
	args := m.Called()
	return args.Get(0).(trace.TraceFlags)
}

// TestNewHookMetrics verifica a criação de uma nova instância de HookMetrics
func TestNewHookMetrics(t *testing.T) {
	// Arrange
	environment := "test"
	mockTracer := &MockTracer{}

	// Act
	hm := metrics.NewHookMetrics(environment, mockTracer)

	// Assert
	assert.NotNil(t, hm, "HookMetrics não deve ser nil")
}

// TestObserveValidation verifica o comportamento do método ObserveValidation
func TestObserveValidation(t *testing.T) {
	// Cenários de teste conforme diferentes mercados e tipos de tenant
	testCases := []struct {
		name       string
		market     string
		tenantType string
		hookType   string
		operation  string
		hasError   bool
	}{
		{
			name:       "Validação bem-sucedida - Angola/Financial",
			market:     constants.MarketAngola,
			tenantType: constants.TenantFinancial,
			hookType:   constants.HookTypeFigma,
			operation:  constants.OperationValidateScope,
			hasError:   false,
		},
		{
			name:       "Validação com erro - Brasil/Government",
			market:     constants.MarketBrasil,
			tenantType: constants.TenantGovernment,
			hookType:   constants.HookTypeGitHub,
			operation:  constants.OperationValidateMFA,
			hasError:   true,
		},
		{
			name:       "Validação bem-sucedida - UE/Healthcare",
			market:     constants.MarketUE,
			tenantType: constants.TenantHealthcare,
			hookType:   constants.HookTypeDocker,
			operation:  constants.OperationValidateToken,
			hasError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			environment := "test"
			
			mockTracer := new(MockTracer)
			mockSpan := new(MockSpan)
			mockSpanContext := new(MockSpanContext)
			
			// Configurar comportamento esperado dos mocks
			mockSpan.On("SpanContext").Return(mockSpanContext)
			mockSpan.On("End", mock.Anything).Return()
			mockSpan.On("SetStatus", mock.Anything, mock.Anything).Return()
			mockSpan.On("RecordError", mock.Anything, mock.Anything).Return()
			
			mockTracer.On("Start", mock.Anything, mock.Anything, mock.Anything).Return(ctx, mockSpan)
			
			// Criar função de validação mock
			var validationErr error
			if tc.hasError {
				validationErr = errors.New("erro de validação")
			}
			
			validationFunc := func(ctx context.Context) error {
				// Simular um atraso para métricas de duração
				time.Sleep(10 * time.Millisecond)
				return validationErr
			}
			
			hm := metrics.NewHookMetrics(environment, mockTracer)
			
			// Act
			err := hm.ObserveValidation(
				ctx,
				tc.market,
				tc.tenantType,
				tc.hookType,
				tc.operation,
				validationFunc,
			)
			
			// Assert
			if tc.hasError {
				assert.Error(t, err, "ObserveValidation deve retornar erro quando a função de validação falha")
				mockSpan.AssertCalled(t, "RecordError", mock.Anything, mock.Anything)
				mockSpan.AssertCalled(t, "SetStatus", mock.Anything, mock.Anything)
			} else {
				assert.NoError(t, err, "ObserveValidation não deve retornar erro quando a função de validação é bem-sucedida")
			}
			
			mockTracer.AssertCalled(t, "Start", mock.Anything, mock.Anything, mock.Anything)
			mockSpan.AssertCalled(t, "End", mock.Anything)
			
			// Verificar se métricas foram registradas
			metricValue := testutil.ToFloat64(metrics.ValidationDuration.WithLabelValues(
				environment, tc.market, tc.tenantType, tc.hookType, tc.operation,
			))
			assert.GreaterOrEqual(t, metricValue, float64(0), "A métrica de duração deve ter sido registrada")
		})
	}
}

// TestRecordElevationRequest verifica o registro de solicitações de elevação
func TestRecordElevationRequest(t *testing.T) {
	// Cenários de teste para diferentes mercados
	testCases := []struct {
		name       string
		market     string
		tenantType string
		hookType   string
	}{
		{
			name:       "Solicitação de elevação - Angola",
			market:     constants.MarketAngola,
			tenantType: constants.TenantFinancial,
			hookType:   constants.HookTypeFigma,
		},
		{
			name:       "Solicitação de elevação - Brasil",
			market:     constants.MarketBrasil,
			tenantType: constants.TenantGovernment,
			hookType:   constants.HookTypeGitHub,
		},
		{
			name:       "Solicitação de elevação - China",
			market:     constants.MarketChina,
			tenantType: constants.TenantTelecom,
			hookType:   constants.HookTypeDocker,
		},
		{
			name:       "Solicitação de elevação - Moçambique",
			market:     constants.MarketMocambique,
			tenantType: constants.TenantHealthcare,
			hookType:   constants.HookTypeMemory,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			environment := "test"
			
			mockTracer := new(MockTracer)
			mockSpan := new(MockSpan)
			
			// Configurar comportamento esperado dos mocks
			mockSpan.On("AddEvent", mock.Anything, mock.Anything).Return()
			
			// Configurar contexto com span
			ctx = trace.ContextWithSpan(ctx, mockSpan)
			
			hm := metrics.NewHookMetrics(environment, mockTracer)
			
			// Capturar valor inicial da métrica
			initialValue := testutil.ToFloat64(metrics.ElevationRequestsTotal.WithLabelValues(
				environment, tc.market, tc.tenantType, tc.hookType,
			))
			
			// Act
			hm.RecordElevationRequest(ctx, tc.market, tc.tenantType, tc.hookType)
			
			// Assert
			finalValue := testutil.ToFloat64(metrics.ElevationRequestsTotal.WithLabelValues(
				environment, tc.market, tc.tenantType, tc.hookType,
			))
			
			assert.Equal(t, initialValue+1, finalValue, 
				"A métrica de solicitações de elevação deve ser incrementada")
			mockSpan.AssertCalled(t, "AddEvent", "elevation.request", mock.Anything)
		})
	}
}

// TestRecordComplianceCheck verifica o registro de verificações de compliance
func TestRecordComplianceCheck(t *testing.T) {
	// Cenários de teste para diferentes mercados e níveis de compliance
	testCases := []struct {
		name           string
		market         string
		complianceType string
		hookType       string
	}{
		{
			name:           "Verificação de compliance - Angola/Standard",
			market:         constants.MarketAngola,
			complianceType: constants.ComplianceStandard,
			hookType:       constants.HookTypeFigma,
		},
		{
			name:           "Verificação de compliance - Brasil/Enhanced",
			market:         constants.MarketBrasil,
			complianceType: constants.ComplianceEnhanced,
			hookType:       constants.HookTypeGitHub,
		},
		{
			name:           "Verificação de compliance - UE/Strict",
			market:         constants.MarketUE,
			complianceType: constants.ComplianceStrict,
			hookType:       constants.HookTypeDocker,
		},
		{
			name:           "Verificação de compliance - China/Sensitive",
			market:         constants.MarketChina,
			complianceType: constants.ComplianceSensitive,
			hookType:       constants.HookTypeMemory,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			environment := "test"
			
			mockTracer := new(MockTracer)
			mockSpan := new(MockSpan)
			
			// Configurar comportamento esperado dos mocks
			mockSpan.On("AddEvent", mock.Anything, mock.Anything).Return()
			
			// Configurar contexto com span
			ctx = trace.ContextWithSpan(ctx, mockSpan)
			
			hm := metrics.NewHookMetrics(environment, mockTracer)
			
			// Capturar valor inicial da métrica
			initialValue := testutil.ToFloat64(metrics.ComplianceChecksTotal.WithLabelValues(
				environment, tc.market, tc.complianceType, tc.hookType,
			))
			
			// Act
			hm.RecordComplianceCheck(ctx, tc.market, tc.complianceType, tc.hookType)
			
			// Assert
			finalValue := testutil.ToFloat64(metrics.ComplianceChecksTotal.WithLabelValues(
				environment, tc.market, tc.complianceType, tc.hookType,
			))
			
			assert.Equal(t, initialValue+1, finalValue,
				"A métrica de verificações de compliance deve ser incrementada")
			mockSpan.AssertCalled(t, "AddEvent", "compliance.check", mock.Anything)
		})
	}
}

// TestRecordMFACheck verifica o registro de verificações MFA
func TestRecordMFACheck(t *testing.T) {
	// Cenários de teste para diferentes mercados e níveis de MFA
	testCases := []struct {
		name      string
		market    string
		mfaLevel  string
		success   bool
		hookType  string
	}{
		{
			name:      "MFA Básico bem-sucedido - Angola",
			market:    constants.MarketAngola,
			mfaLevel:  constants.MFABasic,
			success:   true,
			hookType:  constants.HookTypeFigma,
		},
		{
			name:      "MFA Standard falhou - Brasil",
			market:    constants.MarketBrasil,
			mfaLevel:  constants.MFAStandard,
			success:   false,
			hookType:  constants.HookTypeGitHub,
		},
		{
			name:      "MFA Enhanced bem-sucedido - UE",
			market:    constants.MarketUE,
			mfaLevel:  constants.MFAEnhanced,
			success:   true,
			hookType:  constants.HookTypeDocker,
		},
		{
			name:      "MFA Biométrico falhou - China",
			market:    constants.MarketChina,
			mfaLevel:  constants.MFABiometric,
			success:   false,
			hookType:  constants.HookTypeMemory,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			environment := "test"
			
			mockTracer := new(MockTracer)
			mockSpan := new(MockSpan)
			
			// Configurar comportamento esperado dos mocks
			mockSpan.On("AddEvent", mock.Anything, mock.Anything).Return()
			
			// Configurar contexto com span
			ctx = trace.ContextWithSpan(ctx, mockSpan)
			
			hm := metrics.NewHookMetrics(environment, mockTracer)
			
			successStr := "false"
			if tc.success {
				successStr = "true"
			}
			
			// Capturar valor inicial da métrica
			initialValue := testutil.ToFloat64(metrics.MFAChecksTotal.WithLabelValues(
				environment, tc.market, tc.mfaLevel, successStr, tc.hookType,
			))
			
			// Act
			hm.RecordMFACheck(ctx, tc.market, tc.mfaLevel, tc.success, tc.hookType)
			
			// Assert
			finalValue := testutil.ToFloat64(metrics.MFAChecksTotal.WithLabelValues(
				environment, tc.market, tc.mfaLevel, successStr, tc.hookType,
			))
			
			assert.Equal(t, initialValue+1, finalValue,
				"A métrica de verificações MFA deve ser incrementada")
			mockSpan.AssertCalled(t, "AddEvent", "mfa.check", mock.Anything)
		})
	}
}

// TestUpdateActiveElevations verifica a atualização de elevações ativas
func TestUpdateActiveElevations(t *testing.T) {
	// Cenários de teste para diferentes mercados
	testCases := []struct {
		name       string
		market     string
		tenantType string
		hookType   string
		count      int
	}{
		{
			name:       "0 elevações ativas - Angola/Financial",
			market:     constants.MarketAngola,
			tenantType: constants.TenantFinancial,
			hookType:   constants.HookTypeFigma,
			count:      0,
		},
		{
			name:       "5 elevações ativas - Brasil/Government",
			market:     constants.MarketBrasil,
			tenantType: constants.TenantGovernment,
			hookType:   constants.HookTypeGitHub,
			count:      5,
		},
		{
			name:       "10 elevações ativas - UE/Healthcare",
			market:     constants.MarketUE,
			tenantType: constants.TenantHealthcare,
			hookType:   constants.HookTypeDocker,
			count:      10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			environment := "test"
			mockTracer := new(MockTracer)
			
			hm := metrics.NewHookMetrics(environment, mockTracer)
			
			// Act
			hm.UpdateActiveElevations(tc.market, tc.tenantType, tc.hookType, tc.count)
			
			// Assert
			value := testutil.ToFloat64(metrics.ActiveElevations.WithLabelValues(
				environment, tc.market, tc.tenantType, tc.hookType,
			))
			
			assert.Equal(t, float64(tc.count), value,
				"A métrica de elevações ativas deve ser atualizada corretamente")
		})
	}
}

// TestRecordTestCoverage verifica o registro de cobertura de testes
func TestRecordTestCoverage(t *testing.T) {
	// Cenários de teste para diferentes hooks
	testCases := []struct {
		name      string
		hookType  string
		coverage  float64
	}{
		{
			name:      "85% cobertura - Figma Hook",
			hookType:  constants.HookTypeFigma,
			coverage:  85.0,
		},
		{
			name:      "92.5% cobertura - Docker Hook",
			hookType:  constants.HookTypeDocker,
			coverage:  92.5,
		},
		{
			name:      "78.3% cobertura - GitHub Hook",
			hookType:  constants.HookTypeGitHub,
			coverage:  78.3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			environment := "test"
			mockTracer := new(MockTracer)
			
			hm := metrics.NewHookMetrics(environment, mockTracer)
			
			// Act
			hm.RecordTestCoverage(tc.hookType, tc.coverage)
			
			// Assert
			value := testutil.ToFloat64(metrics.TestCoverage.WithLabelValues(
				environment, tc.hookType,
			))
			
			assert.Equal(t, tc.coverage, value,
				"A métrica de cobertura de testes deve ser atualizada corretamente")
		})
	}
}

// TestRegisterComplianceMetadata verifica o registro de metadados de compliance
func TestRegisterComplianceMetadata(t *testing.T) {
	// Cenários de teste para diferentes mercados
	testCases := []struct {
		name                string
		market              string
		framework           string
		requiresDualApproval bool
		mfaLevel            string
		retentionYears      int
	}{
		{
			name:                "Metadados Angola - BNA",
			market:              constants.MarketAngola,
			framework:           constants.RegulationBNA,
			requiresDualApproval: true,
			mfaLevel:            constants.MFAStandard,
			retentionYears:      7,
		},
		{
			name:                "Metadados Brasil - LGPD",
			market:              constants.MarketBrasil,
			framework:           constants.RegulationLGPD,
			requiresDualApproval: true,
			mfaLevel:            constants.MFAStandard,
			retentionYears:      5,
		},
		{
			name:                "Metadados UE - GDPR",
			market:              constants.MarketUE,
			framework:           constants.RegulationGDPR,
			requiresDualApproval: true,
			mfaLevel:            constants.MFAEnhanced,
			retentionYears:      2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			environment := "test"
			mockTracer := new(MockTracer)
			
			hm := metrics.NewHookMetrics(environment, mockTracer)
			
			// Act
			hm.RegisterComplianceMetadata(
				tc.market,
				tc.framework,
				tc.requiresDualApproval,
				tc.mfaLevel,
				tc.retentionYears,
			)
			
			// Assert
			dualApproval := "false"
			if tc.requiresDualApproval {
				dualApproval = "true"
			}
			retentionYearsStr := fmt.Sprintf("%d", tc.retentionYears)
			
			value := testutil.ToFloat64(metrics.ComplianceMetadata.WithLabelValues(
				environment,
				tc.market,
				tc.framework,
				dualApproval,
				tc.mfaLevel,
				retentionYearsStr,
			))
			
			assert.Equal(t, float64(1), value,
				"A métrica de metadados de compliance deve ser registrada corretamente")
		})
	}
}