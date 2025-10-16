/**
 * @file mcp_event.go
 * @description Definição da estrutura de eventos MCP para Bureau de Crédito
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package mcp

import (
	"time"
)

// MCPEvent representa um evento do Model Context Protocol
type MCPEvent struct {
	// Campos de identificação
	EventID       string                 `json:"eventId"`
	EventType     string                 `json:"eventType"`
	Path          string                 `json:"path"`
	RequestID     string                 `json:"requestId"`
	CorrelationID string                 `json:"correlationId"`
	
	// Campos de contexto
	ProviderID    string                 `json:"providerId"`
	TenantID      string                 `json:"tenantId"`
	UserID        string                 `json:"userId"`
	
	// Campos de tempo
	Timestamp     time.Time              `json:"timestamp"`
	Duration      int64                  `json:"duration"`  // em milissegundos
	
	// Dados do evento
	Data          interface{}            `json:"data"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// NewMCPEvent cria um novo evento MCP com valores padrão
func NewMCPEvent(eventType, path string) *MCPEvent {
	return &MCPEvent{
		EventID:   generateUUID(),
		EventType: eventType,
		Path:      path,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}

// generateUUID gera um UUID v4
func generateUUID() string {
	// Implementação simplificada para ilustração
	// Em produção, usar uma biblioteca de UUID
	return "event-" + time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString gera uma string aleatória do tamanho especificado
func randomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, n)
	
	// Implementação simplificada para ilustração
	// Em produção, usar crypto/rand
	for i := range result {
		result[i] = charset[time.Now().Nanosecond()%len(charset)]
	}
	
	return string(result)
}