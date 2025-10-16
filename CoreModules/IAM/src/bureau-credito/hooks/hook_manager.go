/**
 * @file hook_manager.go
 * @description Gerenciador de hooks de observabilidade do Bureau de Crédito
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package hooks

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// HookManager gerencia o registro e execução de hooks de observabilidade
type HookManager struct {
	hooks         map[string][]Hook
	globalHooks   []Hook
	mutex         sync.RWMutex
	asyncExecution bool
	errorHandler   func(error)
}

// NewHookManager cria uma nova instância do gerenciador de hooks
func NewHookManager(asyncExecution bool) *HookManager {
	return &HookManager{
		hooks:         make(map[string][]Hook),
		globalHooks:   make([]Hook, 0),
		asyncExecution: asyncExecution,
		errorHandler:   defaultErrorHandler,
	}
}

// RegisterHook registra um hook para tipos específicos de operação
func (m *HookManager) RegisterHook(hook Hook, operationTypes ...OperationType) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Se nenhum tipo de operação for especificado, registrar como global
	if len(operationTypes) == 0 {
		m.globalHooks = append(m.globalHooks, hook)
		// Ordenar hooks globais por prioridade
		sort.Slice(m.globalHooks, func(i, j int) bool {
			return m.globalHooks[i].GetPriority() < m.globalHooks[j].GetPriority()
		})
		return
	}

	// Registrar para cada tipo de operação especificado
	for _, opType := range operationTypes {
		opTypeStr := string(opType)
		if m.hooks[opTypeStr] == nil {
			m.hooks[opTypeStr] = make([]Hook, 0)
		}
		m.hooks[opTypeStr] = append(m.hooks[opTypeStr], hook)

		// Ordenar hooks por prioridade
		sort.Slice(m.hooks[opTypeStr], func(i, j int) bool {
			return m.hooks[opTypeStr][i].GetPriority() < m.hooks[opTypeStr][j].GetPriority()
		})
	}

	log.Info().
		Str("hook", hook.GetName()).
		Int("priority", hook.GetPriority()).
		Strs("operations", operationsToStrings(operationTypes)).
		Msg("Hook registrado com sucesso")
}

// ExecuteHooks executa todos os hooks aplicáveis para um tipo de hook e operação
func (m *HookManager) ExecuteHooks(ctx context.Context, hookType HookType, metadata HookMetadata, payload interface{}) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Coletar hooks aplicáveis
	var applicableHooks []Hook

	// Adicionar hooks específicos da operação
	opTypeStr := string(metadata.OperationType)
	if opHooks, exists := m.hooks[opTypeStr]; exists {
		applicableHooks = append(applicableHooks, opHooks...)
	}

	// Adicionar hooks globais
	applicableHooks = append(applicableHooks, m.globalHooks...)

	// Executar hooks
	for _, hook := range applicableHooks {
		// Verificar se o hook deve ser executado
		if !hook.ShouldExecute(ctx, hookType, metadata) {
			continue
		}

		// Executar hook
		if m.asyncExecution {
			go m.executeHookSafely(ctx, hook, hookType, metadata, payload)
		} else {
			m.executeHookSafely(ctx, hook, hookType, metadata, payload)
		}
	}
}

// executeHookSafely executa um hook com tratamento de erros
func (m *HookManager) executeHookSafely(ctx context.Context, hook Hook, hookType HookType, metadata HookMetadata, payload interface{}) {
	defer func() {
		if r := recover(); r != nil {
			err := recoverToError(r)
			log.Error().
				Err(err).
				Str("hook", hook.GetName()).
				Str("hook_type", string(hookType)).
				Str("operation", string(metadata.OperationType)).
				Msg("Pânico durante execução de hook")
			
			if m.errorHandler != nil {
				m.errorHandler(err)
			}
		}
	}()

	// Medir tempo de execução
	start := time.Now()

	// Executar hook
	if err := hook.Execute(ctx, hookType, metadata, payload); err != nil {
		log.Error().
			Err(err).
			Str("hook", hook.GetName()).
			Str("hook_type", string(hookType)).
			Str("operation", string(metadata.OperationType)).
			Msg("Erro durante execução de hook")
		
		if m.errorHandler != nil {
			m.errorHandler(err)
		}
	}

	// Registrar duração se for longa
	duration := time.Since(start)
	if duration > 100*time.Millisecond {
		log.Warn().
			Str("hook", hook.GetName()).
			Str("hook_type", string(hookType)).
			Str("operation", string(metadata.OperationType)).
			Dur("duration", duration).
			Msg("Execução de hook demorou mais que o esperado")
	}
}

// SetErrorHandler define uma função personalizada para tratamento de erros
func (m *HookManager) SetErrorHandler(handler func(error)) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.errorHandler = handler
}

// SetAsyncExecution define se a execução dos hooks deve ser assíncrona
func (m *HookManager) SetAsyncExecution(async bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.asyncExecution = async
}

// defaultErrorHandler é o tratador de erros padrão
func defaultErrorHandler(err error) {
	log.Error().Err(err).Msg("Erro não tratado em hook")
}

// recoverToError converte um valor recuperado de panic em um erro
func recoverToError(r interface{}) error {
	if err, ok := r.(error); ok {
		return err
	}
	return nil
}

// operationsToStrings converte OperationType para strings
func operationsToStrings(ops []OperationType) []string {
	result := make([]string, len(ops))
	for i, op := range ops {
		result[i] = string(op)
	}
	return result
}