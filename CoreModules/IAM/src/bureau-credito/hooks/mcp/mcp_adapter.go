/**
 * @file mcp_adapter.go
 * @description Adaptador para integração de hooks com MCP (Model Context Protocol)
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package mcp

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"

	"innovabiz/iam/src/bureau-credito/hooks"
)

// MCPAdapter é o adaptador para integrar hooks com o Model Context Protocol
type MCPAdapter struct {
	hookManager    *hooks.HookManager
	serviceName    string
	serviceVersion string
	environment    string
	region         string
}

// MCPAdapterConfig contém configurações para o adaptador MCP
type MCPAdapterConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Region         string
}

// NewMCPAdapter cria uma nova instância do adaptador MCP
func NewMCPAdapter(hookManager *hooks.HookManager, config *MCPAdapterConfig) *MCPAdapter {
	if config == nil {
		config = &MCPAdapterConfig{
			ServiceName:    "bureau-credito",
			ServiceVersion: "1.0.0",
			Environment:    "production",
			Region:         "global",
		}
	}

	return &MCPAdapter{
		hookManager:    hookManager,
		serviceName:    config.ServiceName,
		serviceVersion: config.ServiceVersion,
		environment:    config.Environment,
		region:         config.Region,
	}
}

// ProcessMCPEvent processa um evento MCP e executa os hooks apropriados
func (a *MCPAdapter) ProcessMCPEvent(ctx context.Context, event *MCPEvent) error {
	// Validar evento
	if event == nil {
		return nil // Ignorar eventos nulos
	}

	// Converter evento MCP para hooks.HookMetadata
	metadata := a.convertEventToMetadata(event)

	// Determinar tipo de hook com base no tipo de evento
	hookType := a.determineHookType(event.EventType)

	// Determinar tipo de operação com base nos dados do evento
	operationType := a.determineOperationType(event)
	metadata.OperationType = operationType

	// Executar hooks
	a.hookManager.ExecuteHooks(ctx, hookType, metadata, event.Data)

	return nil
}

// convertEventToMetadata converte um evento MCP em hooks.HookMetadata
func (a *MCPAdapter) convertEventToMetadata(event *MCPEvent) hooks.HookMetadata {
	metadata := hooks.HookMetadata{
		RequestID:     event.RequestID,
		CorrelationID: event.CorrelationID,
		Timestamp:     event.Timestamp,
		ProviderID:    event.ProviderID,
		TenantID:      event.TenantID,
		UserID:        event.UserID,
		Environment:   a.environment,
		Region:        a.region,
		Version:       a.serviceVersion,
		Labels:        make(map[string]string),
	}

	// Adicionar metadados extras como labels
	if event.Metadata != nil {
		for k, v := range event.Metadata {
			if strValue, ok := v.(string); ok {
				metadata.Labels[k] = strValue
			} else {
				// Tentar converter para JSON
				if jsonValue, err := json.Marshal(v); err == nil {
					metadata.Labels[k] = string(jsonValue)
				}
			}
		}
	}

	// Adicionar duração se presente
	if event.Duration > 0 {
		metadata.Duration = time.Duration(event.Duration) * time.Millisecond
	}

	return metadata
}

// determineHookType determina o tipo de hook com base no tipo de evento MCP
func (a *MCPAdapter) determineHookType(eventType string) hooks.HookType {
	switch eventType {
	case "before", "start", "begin":
		return hooks.HookBefore
	case "after", "complete", "end":
		return hooks.HookAfter
	case "error", "exception", "failure":
		return hooks.HookError
	default:
		// Log para evento desconhecido
		log.Warn().Str("event_type", eventType).Msg("Tipo de evento MCP desconhecido, usando HookAfter")
		return hooks.HookAfter
	}
}

// determineOperationType determina o tipo de operação com base nos dados do evento
func (a *MCPAdapter) determineOperationType(event *MCPEvent) hooks.OperationType {
	// Verificar se operationType está explicitamente definido no evento
	if opType, ok := event.Metadata["operationType"].(string); ok {
		switch opType {
		case "assessment":
			return hooks.OpAssessment
		case "batchAssessment":
			return hooks.OpBatchAssessment
		case "fraudDetection":
			return hooks.OpFraudDetection
		case "creditScore":
			return hooks.OpCreditScore
		case "riskAssessment":
			return hooks.OpRiskAssessment
		}
	}

	// Inferir com base no path do evento
	if event.Path != "" {
		switch {
		case contains(event.Path, "assessment") && contains(event.Path, "batch"):
			return hooks.OpBatchAssessment
		case contains(event.Path, "assessment"):
			return hooks.OpAssessment
		case contains(event.Path, "fraud"):
			return hooks.OpFraudDetection
		case contains(event.Path, "score"):
			return hooks.OpCreditScore
		case contains(event.Path, "risk"):
			return hooks.OpRiskAssessment
		}
	}

	// Valor padrão
	return hooks.OpAssessment
}

// contains verifica se uma string contém outra
func contains(s, substr string) bool {
	return s != "" && substr != "" && s != substr && s != substr+":"
}