// Package mocks fornece implementações simuladas dos componentes de tracing
// para uso em testes unitários da plataforma INNOVABIZ.
//
// Conformidades: ISO/IEC 29119, ISO 9001, ISO 27001, W3C Trace Context, COBIT 2019, TOGAF 10.0
// Frameworks de Teste: TDD, BDD, FIRST, OpenTelemetry
package mocks

import (
	"context"
	"time"

	"github.com/golang/mock/gomock"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// MockHookTracerInterface é uma interface mockável para HookTracer
// usada em testes unitários para simular comportamentos do sistema de tracing.
type MockHookTracerInterface interface {
	TraceHookOperation(ctx context.Context, hookType, market, tenantType, operation string, attributes []attribute.KeyValue, fn func(context.Context) error) error
	TraceValidateScope(ctx context.Context, hookType, market, tenantType, scope, userId string, validateFunc func(context.Context) error) error
	TraceValidateMFA(ctx context.Context, hookType, market, tenantType, mfaLevel, userId string, validateFunc func(context.Context) error) error
	TraceValidateToken(ctx context.Context, hookType, market, tenantType, tokenId, scope, userId string, validateFunc func(context.Context) error) error
	TraceComplianceCheck(ctx context.Context, hookType, market, tenantType, complianceLevel, framework, userId string, checkFunc func(context.Context) error) error
	TraceAuditEvent(ctx context.Context, hookType, market, tenantType, userId, eventType, eventDetails string) 
	ExtractTraceInfo(ctx context.Context) map[string]string
}

// MockHookTracer é uma implementação mock de HookTracerInterface.
type MockHookTracer struct {
	ctrl *gomock.Controller
}

// NewMockHookTracerInterface cria um novo mock da interface de tracing.
func NewMockHookTracerInterface(ctrl *gomock.Controller) *MockHookTracer {
	return &MockHookTracer{ctrl: ctrl}
}

// TraceHookOperation traça uma operação de hook.
func (m *MockHookTracer) TraceHookOperation(ctx context.Context, hookType, market, tenantType, operation string, attributes []attribute.KeyValue, fn func(context.Context) error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TraceHookOperation", ctx, hookType, market, tenantType, operation, attributes, fn)
	return ret.Error(0)
}

// TraceValidateScope traça uma validação de escopo.
func (m *MockHookTracer) TraceValidateScope(ctx context.Context, hookType, market, tenantType, scope, userId string, validateFunc func(context.Context) error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TraceValidateScope", ctx, hookType, market, tenantType, scope, userId, validateFunc)
	return ret.Error(0)
}

// TraceValidateMFA traça uma validação de MFA.
func (m *MockHookTracer) TraceValidateMFA(ctx context.Context, hookType, market, tenantType, mfaLevel, userId string, validateFunc func(context.Context) error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TraceValidateMFA", ctx, hookType, market, tenantType, mfaLevel, userId, validateFunc)
	return ret.Error(0)
}

// TraceValidateToken traça uma validação de token.
func (m *MockHookTracer) TraceValidateToken(ctx context.Context, hookType, market, tenantType, tokenId, scope, userId string, validateFunc func(context.Context) error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TraceValidateToken", ctx, hookType, market, tenantType, tokenId, scope, userId, validateFunc)
	return ret.Error(0)
}

// TraceComplianceCheck traça uma verificação de compliance.
func (m *MockHookTracer) TraceComplianceCheck(ctx context.Context, hookType, market, tenantType, complianceLevel, framework, userId string, checkFunc func(context.Context) error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TraceComplianceCheck", ctx, hookType, market, tenantType, complianceLevel, framework, userId, checkFunc)
	return ret.Error(0)
}

// TraceAuditEvent traça um evento de auditoria.
func (m *MockHookTracer) TraceAuditEvent(ctx context.Context, hookType, market, tenantType, userId, eventType, eventDetails string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "TraceAuditEvent", ctx, hookType, market, tenantType, userId, eventType, eventDetails)
}

// ExtractTraceInfo extrai informações de trace do contexto.
func (m *MockHookTracer) ExtractTraceInfo(ctx context.Context) map[string]string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExtractTraceInfo", ctx)
	if ret.Get(0) != nil {
		return ret.Get(0).(map[string]string)
	}
	return nil
}

// MockTracer é uma implementação mock da interface trace.Tracer.
type MockTracer struct {
	ctrl *gomock.Controller
}

// NewMockTracer cria um novo mock da interface trace.Tracer.
func NewMockTracer(ctrl *gomock.Controller) *MockTracer {
	return &MockTracer{ctrl: ctrl}
}

// Start cria um novo span.
func (m *MockTracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, spanName}
	for _, opt := range opts {
		varargs = append(varargs, opt)
	}
	ret := m.ctrl.Call(m, "Start", varargs...)
	return ret.Get(0).(context.Context), ret.Get(1).(trace.Span)
}

// MockSpan é uma implementação mock da interface trace.Span.
type MockSpan struct {
	ctrl *gomock.Controller
}

// NewMockSpan cria um novo mock da interface trace.Span.
func NewMockSpan(ctrl *gomock.Controller) *MockSpan {
	return &MockSpan{ctrl: ctrl}
}

// End finaliza o span.
func (m *MockSpan) End(options ...trace.SpanEndOption) {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, option := range options {
		varargs = append(varargs, option)
	}
	m.ctrl.Call(m, "End", varargs...)
}

// AddEvent adiciona um evento ao span.
func (m *MockSpan) AddEvent(name string, options ...trace.EventOption) {
	m.ctrl.T.Helper()
	varargs := []interface{}{name}
	for _, option := range options {
		varargs = append(varargs, option)
	}
	m.ctrl.Call(m, "AddEvent", varargs...)
}

// IsRecording retorna se o span está gravando.
func (m *MockSpan) IsRecording() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsRecording")
	return ret.Bool(0)
}

// RecordError registra um erro no span.
func (m *MockSpan) RecordError(err error, options ...trace.EventOption) {
	m.ctrl.T.Helper()
	varargs := []interface{}{err}
	for _, option := range options {
		varargs = append(varargs, option)
	}
	m.ctrl.Call(m, "RecordError", varargs...)
}

// SetStatus define o status do span.
func (m *MockSpan) SetStatus(code codes.Code, description string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetStatus", code, description)
}

// SetName define o nome do span.
func (m *MockSpan) SetName(name string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetName", name)
}

// SetAttributes define atributos para o span.
func (m *MockSpan) SetAttributes(attributes ...attribute.KeyValue) {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, attr := range attributes {
		varargs = append(varargs, attr)
	}
	m.ctrl.Call(m, "SetAttributes", varargs...)
}

// SpanContext retorna o contexto do span.
func (m *MockSpan) SpanContext() trace.SpanContext {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SpanContext")
	if ret.Get(0) != nil {
		return ret.Get(0).(trace.SpanContext)
	}
	return trace.SpanContext{}
}

// TracerProvider retorna o provedor de tracer.
func (m *MockSpan) TracerProvider() trace.TracerProvider {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TracerProvider")
	if ret.Get(0) != nil {
		return ret.Get(0).(trace.TracerProvider)
	}
	return nil
}