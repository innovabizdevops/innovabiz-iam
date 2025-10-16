/**
 * @file logging_hook.go
 * @description Hook para logging estruturado do Bureau de Crédito
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package hooks

import (
	"context"
	"strings"
	"time"
	
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LoggingHook implementa logging estruturado para operações do Bureau de Crédito
type LoggingHook struct {
	BaseHook
	logLevel         zerolog.Level
	sensitiveFields  []string
	includePayload   bool
	maxPayloadSize   int
	sanitizerEnabled bool
}

// LoggingConfig contém configurações para o hook de logging
type LoggingConfig struct {
	LogLevel         zerolog.Level
	SensitiveFields  []string
	IncludePayload   bool
	MaxPayloadSize   int
	SanitizerEnabled bool
}

// NewLoggingHook cria uma nova instância do hook de logging
func NewLoggingHook(config *LoggingConfig) *LoggingHook {
	if config == nil {
		config = &LoggingConfig{
			LogLevel:         zerolog.InfoLevel,
			SensitiveFields:  []string{"document", "cpf", "cnpj", "senha", "password", "token", "credit_card", "cartao", "secret"},
			IncludePayload:   true,
			MaxPayloadSize:   4096, // 4KB
			SanitizerEnabled: true,
		}
	}
	
	return &LoggingHook{
		BaseHook: BaseHook{
			Name:     "logging_hook",
			Priority: 0, // Maior prioridade, executado primeiro
		},
		logLevel:         config.LogLevel,
		sensitiveFields:  config.SensitiveFields,
		includePayload:   config.IncludePayload,
		maxPayloadSize:   config.MaxPayloadSize,
		sanitizerEnabled: config.SanitizerEnabled,
	}
}

// Execute implementa a interface Hook para logging
func (h *LoggingHook) Execute(ctx context.Context, hookType HookType, metadata HookMetadata, payload interface{}) error {
	// Criar evento de log base
	event := log.WithLevel(h.logLevel).
		Str("hook", h.GetName()).
		Str("hook_type", string(hookType)).
		Str("request_id", metadata.RequestID).
		Str("correlation_id", metadata.CorrelationID).
		Str("operation", string(metadata.OperationType)).
		Str("provider", metadata.ProviderID).
		Str("tenant", metadata.TenantID).
		Str("user_id", metadata.UserID).
		Str("environment", metadata.Environment).
		Str("region", metadata.Region).
		Str("version", metadata.Version).
		Time("timestamp", metadata.Timestamp)
	
	// Adicionar labels personalizados
	for k, v := range metadata.Labels {
		event = event.Str(k, v)
	}
	
	// Adicionar payload sanitizado, se configurado
	if h.includePayload && payload != nil {
		sanitizedPayload := payload
		if h.sanitizerEnabled {
			sanitizedPayload = h.sanitizePayload(payload)
		}
		event = event.Interface("payload", sanitizedPayload)
	}
	
	// Personalizar log por tipo de hook
	switch hookType {
	case HookBefore:
		event.Msg("Bureau de Crédito: iniciando operação")
		
	case HookAfter:
		event.Dur("duration_ms", metadata.Duration).
			Msg("Bureau de Crédito: operação concluída")
		
	case HookError:
		errMsg := "erro desconhecido"
		if err, ok := payload.(error); ok {
			errMsg = err.Error()
		}
		event.Str("error", errMsg).
			Msg("Bureau de Crédito: erro na operação")
	}
	
	return nil
}

// sanitizePayload sanitiza dados sensíveis do payload
func (h *LoggingHook) sanitizePayload(payload interface{}) interface{} {
	// Esta é uma implementação simplificada
	// Uma implementação completa percorreria estruturas complexas recursivamente
	
	// Para demonstração, apenas sanitizamos strings em maps
	if payloadMap, ok := payload.(map[string]interface{}); ok {
		result := make(map[string]interface{})
		
		for k, v := range payloadMap {
			// Verificar se o campo é sensível
			if h.isFieldSensitive(k) {
				result[k] = "[REDACTED]"
			} else if strValue, ok := v.(string); ok {
				// Para strings, verificar se contém dados sensíveis
				result[k] = h.sanitizeString(strValue)
			} else if nestedMap, ok := v.(map[string]interface{}); ok {
				// Recursivamente sanitizar maps aninhados
				result[k] = h.sanitizePayload(nestedMap)
			} else {
				// Manter outros tipos inalterados
				result[k] = v
			}
		}
		
		return result
	}
	
	// Se não for um map, retornar inalterado
	return payload
}

// isFieldSensitive verifica se um campo é sensível com base nas regras configuradas
func (h *LoggingHook) isFieldSensitive(fieldName string) bool {
	fieldNameLower := strings.ToLower(fieldName)
	
	for _, sensitive := range h.sensitiveFields {
		if strings.Contains(fieldNameLower, strings.ToLower(sensitive)) {
			return true
		}
	}
	
	return false
}

// sanitizeString sanitiza uma string para remover dados sensíveis
func (h *LoggingHook) sanitizeString(input string) string {
	// Implementação simplificada que limita o tamanho da string
	if len(input) > h.maxPayloadSize {
		return input[:h.maxPayloadSize] + "... [truncated]"
	}
	
	// Poderia implementar outras regras de sanitização aqui
	// como ofuscação de números de cartão de crédito, etc.
	
	return input
}