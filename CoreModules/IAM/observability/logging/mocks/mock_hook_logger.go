// Package mocks fornece implementações simuladas dos componentes de logging
// para uso em testes unitários da plataforma INNOVABIZ.
//
// Conformidades: ISO/IEC 29119, ISO 9001, ISO 27001, COBIT 2019, TOGAF 10.0, DMBOK 2.0
// Frameworks de Teste: TDD, BDD, FIRST
package mocks

import (
	"context"

	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
)

// MockHookLoggerInterface é uma interface mockável para HookLogger
// usada em testes unitários para simular comportamentos do sistema de logging.
type MockHookLoggerInterface interface {
	// Métodos de logging contextualizados
	WithContext(ctx context.Context) *zap.Logger
	
	// Métodos de logging específicos
	LogAuditEvent(ctx context.Context, market, tenantType, hookType, operation, userId, eventType, eventDetails string)
	LogSecurityEvent(ctx context.Context, market, tenantType, hookType, operation, userId, severity, eventDetails string)
	LogComplianceEvent(ctx context.Context, market, tenantType, hookType, operation, userId, regulation, status, details string)
	LogHookError(ctx context.Context, market, tenantType, hookType, operation, userId string, err error, fields ...zap.Field)
	
	// Métodos de configuração
	SetLogLevel(level string) error
	EnableComplianceAudit(complianceLogPath string) error
}

// MockHookLogger é uma implementação mock de HookLoggerInterface.
type MockHookLogger struct {
	ctrl *gomock.Controller
}

// NewMockHookLoggerInterface cria um novo mock da interface de logging.
func NewMockHookLoggerInterface(ctrl *gomock.Controller) *MockHookLogger {
	return &MockHookLogger{ctrl: ctrl}
}

// WithContext adiciona contexto ao logger.
func (m *MockHookLogger) WithContext(ctx context.Context) *zap.Logger {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithContext", ctx)
	if ret.Get(0) != nil {
		return ret.Get(0).(*zap.Logger)
	}
	return zap.NewNop()
}

// LogAuditEvent registra um evento de auditoria.
func (m *MockHookLogger) LogAuditEvent(ctx context.Context, market, tenantType, hookType, operation, userId, eventType, eventDetails string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "LogAuditEvent", ctx, market, tenantType, hookType, operation, userId, eventType, eventDetails)
}

// LogSecurityEvent registra um evento de segurança.
func (m *MockHookLogger) LogSecurityEvent(ctx context.Context, market, tenantType, hookType, operation, userId, severity, eventDetails string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "LogSecurityEvent", ctx, market, tenantType, hookType, operation, userId, severity, eventDetails)
}

// LogComplianceEvent registra um evento de compliance.
func (m *MockHookLogger) LogComplianceEvent(ctx context.Context, market, tenantType, hookType, operation, userId, regulation, status, details string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "LogComplianceEvent", ctx, market, tenantType, hookType, operation, userId, regulation, status, details)
}

// LogHookError registra um erro de hook.
func (m *MockHookLogger) LogHookError(ctx context.Context, market, tenantType, hookType, operation, userId string, err error, fields ...zap.Field) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, market, tenantType, hookType, operation, userId, err}
	for _, field := range fields {
		varargs = append(varargs, field)
	}
	m.ctrl.Call(m, "LogHookError", varargs...)
}

// SetLogLevel define o nível de log.
func (m *MockHookLogger) SetLogLevel(level string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetLogLevel", level)
	return ret.Error(0)
}

// EnableComplianceAudit ativa o audit de compliance.
func (m *MockHookLogger) EnableComplianceAudit(complianceLogPath string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnableComplianceAudit", complianceLogPath)
	return ret.Error(0)
}