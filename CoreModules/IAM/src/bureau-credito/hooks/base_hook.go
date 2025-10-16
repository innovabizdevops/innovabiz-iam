/**
 * @file base_hook.go
 * @description Interface base para hooks de observabilidade do Bureau de Crédito
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package hooks

import (
	"context"
	"time"
)

// HookType representa o tipo de hook
type HookType string

const (
	// HookBefore é executado antes da operação principal
	HookBefore HookType = "before"
	
	// HookAfter é executado após a operação principal
	HookAfter HookType = "after"
	
	// HookError é executado quando ocorre um erro
	HookError HookType = "error"
)

// OperationType representa o tipo de operação sendo monitorada
type OperationType string

const (
	// OpAssessment representa uma operação de avaliação de crédito
	OpAssessment OperationType = "assessment"
	
	// OpBatchAssessment representa uma operação de avaliação em lote
	OpBatchAssessment OperationType = "batch_assessment"
	
	// OpFraudDetection representa uma operação de detecção de fraude
	OpFraudDetection OperationType = "fraud_detection"
	
	// OpCreditScore representa uma operação de pontuação de crédito
	OpCreditScore OperationType = "credit_score"
	
	// OpRiskAssessment representa uma operação de avaliação de risco
	OpRiskAssessment OperationType = "risk_assessment"
)

// HookMetadata contém metadados comuns para todos os hooks
type HookMetadata struct {
	RequestID     string            // ID único da requisição
	CorrelationID string            // ID de correlação para rastreamento
	Timestamp     time.Time         // Momento da execução do hook
	Duration      time.Duration     // Duração da operação (apenas para hooks After)
	ProviderID    string            // ID do provedor de crédito
	OperationType OperationType     // Tipo de operação
	TenantID      string            // ID do tenant
	UserID        string            // ID do usuário
	Labels        map[string]string // Rótulos adicionais para categorização
	Environment   string            // Ambiente (dev, qa, prod, etc.)
	Region        string            // Região geográfica
	Version       string            // Versão do serviço
}

// Hook é a interface base para todos os hooks de observabilidade
type Hook interface {
	// Execute executa o hook com o contexto e metadados fornecidos
	Execute(ctx context.Context, hookType HookType, metadata HookMetadata, payload interface{}) error
	
	// GetName retorna o nome do hook
	GetName() string
	
	// GetPriority retorna a prioridade de execução do hook (menor valor = maior prioridade)
	GetPriority() int
	
	// ShouldExecute determina se o hook deve ser executado com base no contexto e metadados
	ShouldExecute(ctx context.Context, hookType HookType, metadata HookMetadata) bool
}

// BaseHook implementa a funcionalidade básica comum a todos os hooks
type BaseHook struct {
	Name     string
	Priority int
}

// GetName implementa a interface Hook
func (h *BaseHook) GetName() string {
	return h.Name
}

// GetPriority implementa a interface Hook
func (h *BaseHook) GetPriority() int {
	return h.Priority
}

// ShouldExecute implementa a interface Hook com comportamento padrão
func (h *BaseHook) ShouldExecute(ctx context.Context, hookType HookType, metadata HookMetadata) bool {
	// Por padrão, todos os hooks são executados
	return true
}