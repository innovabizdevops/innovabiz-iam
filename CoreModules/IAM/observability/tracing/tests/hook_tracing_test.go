// Package tracing_test fornece testes unitários para o pacote de tracing dos hooks MCP-IAM
//
// Testes implementados seguindo padrões:
// - TDD (Test-Driven Development)
// - BDD (Behavior-Driven Development)
// - ISO/IEC 29119 (Software Testing Standards)
//
// Frameworks e normas de observabilidade:
// - OpenTelemetry
// - W3C Trace Context
// - TOGAF 10.0, DMBOK 2.0, COBIT 2019
package tracing_test

import (
	"context"
	"errors"
	"testing"

	"github.com/innovabiz/iam/constants"
	"github.com/innovabiz/iam/observability/tracing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
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

// newTestLogger cria um logger para testes
func newTestLogger(t *testing.T) *zap.Logger {
	return zaptest.NewLogger(t, zaptest.Level(zapcore.DebugLevel))
}

// TestNewHookTracer verifica a criação de uma nova instância de HookTracer
func TestNewHookTracer(t *testing.T) {
	// Arrange
	serviceName := "innovabiz-iam-hook-tests"
	logger := newTestLogger(t)

	// Act
	ht := tracing.NewHookTracer(serviceName, logger)

	// Assert
	assert.NotNil(t, ht, "HookTracer não deve ser nil")
}

// TestTraceHookOperation verifica o comportamento do método TraceHookOperation
func TestTraceHookOperation(t *testing.T) {
	// Cenários de teste para diferentes mercados e tipos de operações
	testCases := []struct {
		name       string
		hookType   string
		market     string
		tenantType string
		operation  string
		attributes []attribute.KeyValue
		hasError   bool
	}{
		{
			name:       "Operação bem-sucedida - Figma/Angola/Financial/ValidateScope",
			hookType:   constants.HookTypeFigma,
			market:     constants.MarketAngola,
			tenantType: constants.TenantFinancial,
			operation:  constants.OperationValidateScope,
			attributes: []attribute.KeyValue{
				attribute.String("scope", "figma:read"),
			},
			hasError:   false,
		},
		{
			name:       "Operação com erro - GitHub/Brasil/Government/ValidateMFA",
			hookType:   constants.HookTypeGitHub,
			market:     constants.MarketBrasil,
			tenantType: constants.TenantGovernment,
			operation:  constants.OperationValidateMFA,
			attributes: []attribute.KeyValue{
				attribute.String("mfa_level", constants.MFAStandard),
				attribute.String("user_id", "user123"),
			},
			hasError:   true,
		},
		{
			name:       "Operação bem-sucedida - Docker/UE/Healthcare/ValidateToken",
			hookType:   constants.HookTypeDocker,
			market:     constants.MarketUE,
			tenantType: constants.TenantHealthcare,
			operation:  constants.OperationValidateToken,
			attributes: []attribute.KeyValue{
				attribute.String("token_id", "token123"),
				attribute.String("scope", "docker:push"),
			},
			hasError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			logger := newTestLogger(t)
			
			mockTracer := new(MockTracer)
			mockSpan := new(MockSpan)
			
			// Configurar comportamento esperado dos mocks
			mockSpan.On("End").Return()
			mockSpan.On("SetAttributes", mock.Anything).Return()
			mockSpan.On("RecordError", mock.Anything, mock.Anything).Maybe()
			
			mockTracer.On("Start", mock.Anything, mock.Anything, mock.Anything).Return(ctx, mockSpan)
			
			// Criar função de operação mock
			var operationErr error
			if tc.hasError {
				operationErr = errors.New("erro na operação")
			}
			
			operationFunc := func(ctx context.Context) error {
				return operationErr
			}
			
			ht := tracing.NewHookTracer("test-service", mockTracer)
			
			// Act
			err := ht.TraceHookOperation(
				ctx,
				tc.hookType,
				tc.market,
				tc.tenantType,
				tc.operation,
				tc.attributes,
				operationFunc,
			)
			
			// Assert
			if tc.hasError {
				assert.Error(t, err, "TraceHookOperation deve retornar erro quando a função de operação falha")
				mockSpan.AssertCalled(t, "RecordError", mock.Anything, mock.Anything)
				mockSpan.AssertCalled(t, "SetAttributes", mock.MatchedBy(func(attrs []attribute.KeyValue) bool {
					// Verificar se os atributos de erro estão presentes
					for _, attr := range attrs {
						if attr.Key == "status" && attr.Value.AsString() == "error" {
							return true
						}
					}
					return false
				}))
			} else {
				assert.NoError(t, err, "TraceHookOperation não deve retornar erro quando a função de operação é bem-sucedida")
				mockSpan.AssertCalled(t, "SetAttributes", mock.MatchedBy(func(attrs []attribute.KeyValue) bool {
					// Verificar se os atributos de sucesso estão presentes
					for _, attr := range attrs {
						if attr.Key == "status" && attr.Value.AsString() == "success" {
							return true
						}
					}
					return false
				}))
			}
			
			mockTracer.AssertCalled(t, "Start", mock.Anything, mock.Anything, mock.Anything)
			mockSpan.AssertCalled(t, "End")
			mockSpan.AssertCalled(t, "SetAttributes", mock.Anything)
		})
	}
}

// TestTraceScopeValidation verifica o comportamento do método TraceScopeValidation
func TestTraceScopeValidation(t *testing.T) {
	// Cenários de teste para diferentes tipos de validação de escopo
	testCases := []struct {
		name       string
		hookType   string
		market     string
		tenantType string
		scope      string
		hasError   bool
	}{
		{
			name:       "Validação de escopo bem-sucedida - Figma/Angola/Financial",
			hookType:   constants.HookTypeFigma,
			market:     constants.MarketAngola,
			tenantType: constants.TenantFinancial,
			scope:      "figma:read",
			hasError:   false,
		},
		{
			name:       "Validação de escopo com erro - GitHub/Brasil/Government",
			hookType:   constants.HookTypeGitHub,
			market:     constants.MarketBrasil,
			tenantType: constants.TenantGovernment,
			scope:      "github:admin",
			hasError:   true,
		},
		{
			name:       "Validação de escopo bem-sucedida - Docker/China/Telecom",
			hookType:   constants.HookTypeDocker,
			market:     constants.MarketChina,
			tenantType: constants.TenantTelecom,
			scope:      "docker:pull",
			hasError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			logger := newTestLogger(t)
			
			mockTracer := new(MockTracer)
			mockSpan := new(MockSpan)
			
			// Configurar comportamento esperado dos mocks
			mockSpan.On("End").Return()
			mockSpan.On("SetAttributes", mock.Anything).Return()
			mockSpan.On("RecordError", mock.Anything, mock.Anything).Maybe()
			
			mockTracer.On("Start", mock.Anything, mock.Anything, mock.Anything).Return(ctx, mockSpan)
			
			// Criar função de validação mock
			var validationErr error
			if tc.hasError {
				validationErr = errors.New("erro na validação de escopo")
			}
			
			validateFunc := func(ctx context.Context) error {
				return validationErr
			}
			
			ht := tracing.NewHookTracer("test-service", logger)
			
			// Act
			err := ht.TraceScopeValidation(
				ctx,
				tc.hookType,
				tc.market,
				tc.tenantType,
				tc.scope,
				validateFunc,
			)
			
			// Assert
			if tc.hasError {
				assert.Error(t, err, "TraceScopeValidation deve retornar erro quando a validação falha")
			} else {
				assert.NoError(t, err, "TraceScopeValidation não deve retornar erro quando a validação é bem-sucedida")
			}
			
			mockTracer.AssertExpectations(t)
			mockSpan.AssertExpectations(t)
		})
	}
}

// TestTraceMFAValidation verifica o comportamento do método TraceMFAValidation
func TestTraceMFAValidation(t *testing.T) {
	// Cenários de teste para diferentes tipos de validação MFA
	testCases := []struct {
		name       string
		hookType   string
		market     string
		tenantType string
		mfaLevel   string
		userId     string
		hasError   bool
	}{
		{
			name:       "Validação MFA básica bem-sucedida - Figma/Angola/Financial",
			hookType:   constants.HookTypeFigma,
			market:     constants.MarketAngola,
			tenantType: constants.TenantFinancial,
			mfaLevel:   constants.MFABasic,
			userId:     "user123",
			hasError:   false,
		},
		{
			name:       "Validação MFA standard com erro - GitHub/Brasil/Government",
			hookType:   constants.HookTypeGitHub,
			market:     constants.MarketBrasil,
			tenantType: constants.TenantGovernment,
			mfaLevel:   constants.MFAStandard,
			userId:     "user456",
			hasError:   true,
		},
		{
			name:       "Validação MFA enhanced bem-sucedida - Docker/UE/Healthcare",
			hookType:   constants.HookTypeDocker,
			market:     constants.MarketUE,
			tenantType: constants.TenantHealthcare,
			mfaLevel:   constants.MFAEnhanced,
			userId:     "user789",
			hasError:   false,
		},
		{
			name:       "Validação MFA biométrica com erro - Memory/China/Telecom",
			hookType:   constants.HookTypeMemory,
			market:     constants.MarketChina,
			tenantType: constants.TenantTelecom,
			mfaLevel:   constants.MFABiometric,
			userId:     "user101",
			hasError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			logger := newTestLogger(t)
			
			mockTracer := new(MockTracer)
			mockSpan := new(MockSpan)
			
			// Configurar comportamento esperado dos mocks
			mockSpan.On("End").Return()
			mockSpan.On("SetAttributes", mock.Anything).Return()
			mockSpan.On("RecordError", mock.Anything, mock.Anything).Maybe()
			
			mockTracer.On("Start", mock.Anything, mock.Anything, mock.Anything).Return(ctx, mockSpan)
			
			// Criar função de validação mock
			var validationErr error
			if tc.hasError {
				validationErr = errors.New("erro na validação MFA")
			}
			
			validateFunc := func(ctx context.Context) error {
				return validationErr
			}
			
			ht := tracing.NewHookTracer("test-service", logger)
			
			// Act
			err := ht.TraceMFAValidation(
				ctx,
				tc.hookType,
				tc.market,
				tc.tenantType,
				tc.mfaLevel,
				tc.userId,
				validateFunc,
			)
			
			// Assert
			if tc.hasError {
				assert.Error(t, err, "TraceMFAValidation deve retornar erro quando a validação falha")
			} else {
				assert.NoError(t, err, "TraceMFAValidation não deve retornar erro quando a validação é bem-sucedida")
			}
			
			mockTracer.AssertExpectations(t)
			mockSpan.AssertExpectations(t)
		})
	}
}

// TestTraceTokenValidation verifica o comportamento do método TraceTokenValidation
func TestTraceTokenValidation(t *testing.T) {
	// Cenários de teste para diferentes tipos de validação de token
	testCases := []struct {
		name       string
		hookType   string
		market     string
		tenantType string
		tokenId    string
		scope      string
		hasError   bool
	}{
		{
			name:       "Validação de token bem-sucedida - Figma/Angola/Financial",
			hookType:   constants.HookTypeFigma,
			market:     constants.MarketAngola,
			tenantType: constants.TenantFinancial,
			tokenId:    "token123",
			scope:      "figma:read",
			hasError:   false,
		},
		{
			name:       "Validação de token com erro - GitHub/Brasil/Government",
			hookType:   constants.HookTypeGitHub,
			market:     constants.MarketBrasil,
			tenantType: constants.TenantGovernment,
			tokenId:    "token456",
			scope:      "github:admin",
			hasError:   true,
		},
		{
			name:       "Validação de token bem-sucedida - Docker/UE/Healthcare",
			hookType:   constants.HookTypeDocker,
			market:     constants.MarketUE,
			tenantType: constants.TenantHealthcare,
			tokenId:    "token789",
			scope:      "docker:push",
			hasError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			logger := newTestLogger(t)
			
			mockTracer := new(MockTracer)
			mockSpan := new(MockSpan)
			
			// Configurar comportamento esperado dos mocks
			mockSpan.On("End").Return()
			mockSpan.On("SetAttributes", mock.Anything).Return()
			mockSpan.On("RecordError", mock.Anything, mock.Anything).Maybe()
			
			mockTracer.On("Start", mock.Anything, mock.Anything, mock.Anything).Return(ctx, mockSpan)
			
			// Criar função de validação mock
			var validationErr error
			if tc.hasError {
				validationErr = errors.New("erro na validação de token")
			}
			
			validateFunc := func(ctx context.Context) error {
				return validationErr
			}
			
			ht := tracing.NewHookTracer("test-service", logger)
			
			// Act
			err := ht.TraceTokenValidation(
				ctx,
				tc.hookType,
				tc.market,
				tc.tenantType,
				tc.tokenId,
				tc.scope,
				validateFunc,
			)
			
			// Assert
			if tc.hasError {
				assert.Error(t, err, "TraceTokenValidation deve retornar erro quando a validação falha")
			} else {
				assert.NoError(t, err, "TraceTokenValidation não deve retornar erro quando a validação é bem-sucedida")
			}
			
			mockTracer.AssertExpectations(t)
			mockSpan.AssertExpectations(t)
		})
	}
}

// TestTraceComplianceCheck verifica o comportamento do método TraceComplianceCheck
func TestTraceComplianceCheck(t *testing.T) {
	// Cenários de teste para diferentes verificações de compliance
	testCases := []struct {
		name            string
		hookType        string
		market          string
		tenantType      string
		complianceLevel string
		regulation      string
		hasError        bool
	}{
		{
			name:            "Verificação de compliance bem-sucedida - Figma/Angola/Financial/Standard/BNA",
			hookType:        constants.HookTypeFigma,
			market:          constants.MarketAngola,
			tenantType:      constants.TenantFinancial,
			complianceLevel: constants.ComplianceStandard,
			regulation:      constants.RegulationBNA,
			hasError:        false,
		},
		{
			name:            "Verificação de compliance com erro - GitHub/Brasil/Government/Enhanced/LGPD",
			hookType:        constants.HookTypeGitHub,
			market:          constants.MarketBrasil,
			tenantType:      constants.TenantGovernment,
			complianceLevel: constants.ComplianceEnhanced,
			regulation:      constants.RegulationLGPD,
			hasError:        true,
		},
		{
			name:            "Verificação de compliance bem-sucedida - Docker/UE/Healthcare/Strict/GDPR",
			hookType:        constants.HookTypeDocker,
			market:          constants.MarketUE,
			tenantType:      constants.TenantHealthcare,
			complianceLevel: constants.ComplianceStrict,
			regulation:      constants.RegulationGDPR,
			hasError:        false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			logger := newTestLogger(t)
			
			mockTracer := new(MockTracer)
			mockSpan := new(MockSpan)
			
			// Configurar comportamento esperado dos mocks
			mockSpan.On("End").Return()
			mockSpan.On("SetAttributes", mock.Anything).Return()
			mockSpan.On("RecordError", mock.Anything, mock.Anything).Maybe()
			
			mockTracer.On("Start", mock.Anything, mock.Anything, mock.Anything).Return(ctx, mockSpan)
			
			// Criar função de verificação mock
			var checkErr error
			if tc.hasError {
				checkErr = errors.New("erro na verificação de compliance")
			}
			
			checkFunc := func(ctx context.Context) error {
				return checkErr
			}
			
			ht := tracing.NewHookTracer("test-service", logger)
			
			// Act
			err := ht.TraceComplianceCheck(
				ctx,
				tc.hookType,
				tc.market,
				tc.tenantType,
				tc.complianceLevel,
				tc.regulation,
				checkFunc,
			)
			
			// Assert
			if tc.hasError {
				assert.Error(t, err, "TraceComplianceCheck deve retornar erro quando a verificação falha")
			} else {
				assert.NoError(t, err, "TraceComplianceCheck não deve retornar erro quando a verificação é bem-sucedida")
			}
			
			mockTracer.AssertExpectations(t)
			mockSpan.AssertExpectations(t)
		})
	}
}

// TestTraceAuditEvent verifica o comportamento do método TraceAuditEvent
func TestTraceAuditEvent(t *testing.T) {
	// Cenários de teste para diferentes eventos de auditoria
	testCases := []struct {
		name        string
		hookType    string
		market      string
		tenantType  string
		userId      string
		eventType   string
		eventDetails string
	}{
		{
			name:        "Evento de auditoria - Figma/Angola/Financial/ElevationRequest",
			hookType:    constants.HookTypeFigma,
			market:      constants.MarketAngola,
			tenantType:  constants.TenantFinancial,
			userId:      "user123",
			eventType:   "elevation_request",
			eventDetails: "Solicitação de elevação para acesso a figma:read",
		},
		{
			name:        "Evento de auditoria - GitHub/Brasil/Government/TokenRevocation",
			hookType:    constants.HookTypeGitHub,
			market:      constants.MarketBrasil,
			tenantType:  constants.TenantGovernment,
			userId:      "user456",
			eventType:   "token_revocation",
			eventDetails: "Revogação de token para escopo github:admin",
		},
		{
			name:        "Evento de auditoria - Docker/UE/Healthcare/MFAEnforcement",
			hookType:    constants.HookTypeDocker,
			market:      constants.MarketUE,
			tenantType:  constants.TenantHealthcare,
			userId:      "user789",
			eventType:   "mfa_enforcement",
			eventDetails: "Aplicação de MFA Enhanced conforme GDPR para docker:push",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			logger := zaptest.NewLogger(t)
			
			mockTracer := new(MockTracer)
			mockSpan := new(MockSpan)
			
			// Configurar comportamento esperado dos mocks
			mockSpan.On("End").Return()
			
			mockTracer.On("Start", mock.Anything, mock.Anything, mock.Anything).Return(ctx, mockSpan)
			
			ht := tracing.NewHookTracer("test-service", logger)
			
			// Act - esta função não retorna erro
			ht.TraceAuditEvent(
				ctx,
				tc.hookType,
				tc.market,
				tc.tenantType,
				tc.userId,
				tc.eventType,
				tc.eventDetails,
			)
			
			// Assert
			mockTracer.AssertExpectations(t)
			mockSpan.AssertExpectations(t)
		})
	}
}

// TestExtractTraceInfo verifica a extração de informações de trace do contexto
func TestExtractTraceInfo(t *testing.T) {
	// Arrange
	ctx := context.Background()
	logger := newTestLogger(t)
	
	mockTracer := new(MockTracer)
	mockSpan := new(MockSpan)
	mockSpanContext := new(MockSpanContext)
	
	// Configurar IDs mock para o SpanContext
	traceID := trace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	spanID := trace.SpanID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	
	// Configurar comportamento esperado dos mocks
	mockSpanContext.On("IsValid").Return(true)
	mockSpanContext.On("TraceID").Return(traceID)
	mockSpanContext.On("SpanID").Return(spanID)
	
	mockSpan.On("SpanContext").Return(mockSpanContext)
	
	// Criar contexto com span
	ctx = trace.ContextWithSpan(ctx, mockSpan)
	
	ht := tracing.NewHookTracer("test-service", logger)
	
	// Act
	traceInfo := ht.ExtractTraceInfo(ctx)
	
	// Assert
	require.NotEmpty(t, traceInfo, "Informações de trace não devem estar vazias")
	assert.Equal(t, traceID.String(), traceInfo["trace_id"], "ID de trace deve ser extraído corretamente")
	assert.Equal(t, spanID.String(), traceInfo["span_id"], "ID de span deve ser extraído corretamente")
	
	mockSpan.AssertExpectations(t)
	mockSpanContext.AssertExpectations(t)
}

// TestExtractTraceInfoNoSpan verifica o comportamento quando não há span no contexto
func TestExtractTraceInfoNoSpan(t *testing.T) {
	// Arrange
	ctx := context.Background()
	logger := newTestLogger(t)
	
	ht := tracing.NewHookTracer("test-service", logger)
	
	// Act
	traceInfo := ht.ExtractTraceInfo(ctx)
	
	// Assert
	assert.Empty(t, traceInfo, "Informações de trace devem estar vazias quando não há span no contexto")
}