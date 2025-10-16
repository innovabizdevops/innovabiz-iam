/**
 * @file update_context_verification_level.go
 * @description Comando e handler para atualização do nível de verificação de um contexto
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	
	"innovabiz/iam/src/multi-context/domain/models"
	"innovabiz/iam/src/multi-context/domain/services"
)

// UpdateContextVerificationLevelCommand representa o comando para atualizar o nível de verificação de um contexto
type UpdateContextVerificationLevelCommand struct {
	ContextID          uuid.UUID                // ID do contexto a ser atualizado
	VerificationLevel  models.VerificationLevel // Novo nível de verificação
	VerificationSource string                   // Fonte da verificação (opcional)
	Reason             string                   // Motivo da atualização para fins de auditoria
	RequestedBy        string                   // Utilizador ou sistema que solicitou a atualização
}

// UpdateContextVerificationLevelHandler gerencia a atualização do nível de verificação de um contexto
type UpdateContextVerificationLevelHandler struct {
	contextService *services.ContextService
	auditLogger    services.AuditLogger
}

// NewUpdateContextVerificationLevelHandler cria uma nova instância do handler
func NewUpdateContextVerificationLevelHandler(
	contextService *services.ContextService,
	auditLogger services.AuditLogger,
) *UpdateContextVerificationLevelHandler {
	return &UpdateContextVerificationLevelHandler{
		contextService: contextService,
		auditLogger:    auditLogger,
	}
}

// Handle processa o comando de atualização do nível de verificação
func (h *UpdateContextVerificationLevelHandler) Handle(ctx context.Context, cmd UpdateContextVerificationLevelCommand) error {
	// Registrar início da operação para rastreabilidade
	startTime := time.Now()
	operationID := uuid.New()
	
	h.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "UPDATE_VERIFICATION_LEVEL_INITIATED",
		ResourceID:  cmd.ContextID.String(),
		ResourceType: "IDENTITY_CONTEXT",
		UserID:      cmd.RequestedBy,
		Timestamp:   startTime,
		Details: map[string]interface{}{
			"operation_id":       operationID,
			"context_id":         cmd.ContextID,
			"verification_level": cmd.VerificationLevel,
			"reason":             cmd.Reason,
		},
	})
	
	// Validar o nível de verificação
	if !isValidVerificationLevel(cmd.VerificationLevel) {
		err := fmt.Errorf("nível de verificação inválido: %s", cmd.VerificationLevel)
		
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "UPDATE_VERIFICATION_LEVEL_FAILED",
			ResourceID:  cmd.ContextID.String(),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":       operationID,
				"context_id":         cmd.ContextID,
				"verification_level": cmd.VerificationLevel,
				"error":              err.Error(),
				"duration_ms":        time.Since(startTime).Milliseconds(),
			},
		})
		
		return err
	}
	
	// Buscar o contexto atual para verificar o estado e registrar mudanças
	currentContext, err := h.contextService.GetContextByID(ctx, cmd.ContextID)
	if err != nil {
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "UPDATE_VERIFICATION_LEVEL_FAILED",
			ResourceID:  cmd.ContextID.String(),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":       operationID,
				"context_id":         cmd.ContextID,
				"verification_level": cmd.VerificationLevel,
				"error":              err.Error(),
				"duration_ms":        time.Since(startTime).Milliseconds(),
			},
		})
		
		return fmt.Errorf("erro ao buscar contexto: %w", err)
	}
	
	// Verificar se o contexto está ativo
	if currentContext.Status != models.ContextStatusActive {
		err := fmt.Errorf("não é possível atualizar o nível de verificação de um contexto não ativo")
		
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "UPDATE_VERIFICATION_LEVEL_FAILED",
			ResourceID:  cmd.ContextID.String(),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":       operationID,
				"context_id":         cmd.ContextID,
				"verification_level": cmd.VerificationLevel,
				"current_status":     currentContext.Status,
				"error":              err.Error(),
				"duration_ms":        time.Since(startTime).Milliseconds(),
			},
		})
		
		return err
	}
	
	// Verificar se a mudança é válida (não pode diminuir o nível de verificação, apenas aumentar)
	currentLevel := verificationLevelValue(currentContext.VerificationLevel)
	newLevel := verificationLevelValue(cmd.VerificationLevel)
	
	if newLevel < currentLevel {
		err := fmt.Errorf(
			"não é permitido diminuir o nível de verificação (atual: %s, solicitado: %s)",
			currentContext.VerificationLevel,
			cmd.VerificationLevel,
		)
		
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "UPDATE_VERIFICATION_LEVEL_FAILED",
			ResourceID:  cmd.ContextID.String(),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":              operationID,
				"context_id":                cmd.ContextID,
				"current_verification_level": currentContext.VerificationLevel,
				"requested_verification_level": cmd.VerificationLevel,
				"error":                     err.Error(),
				"duration_ms":               time.Since(startTime).Milliseconds(),
			},
		})
		
		return err
	}
	
	// Atualizar o nível de verificação no serviço de domínio
	if err := h.contextService.UpdateVerificationLevel(ctx, cmd.ContextID, cmd.VerificationLevel); err != nil {
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "UPDATE_VERIFICATION_LEVEL_FAILED",
			ResourceID:  cmd.ContextID.String(),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":       operationID,
				"context_id":         cmd.ContextID,
				"verification_level": cmd.VerificationLevel,
				"error":              err.Error(),
				"duration_ms":        time.Since(startTime).Milliseconds(),
			},
		})
		
		return fmt.Errorf("erro ao atualizar nível de verificação: %w", err)
	}
	
	// Se foi fornecida uma fonte de verificação, atualizar os metadados do contexto
	if cmd.VerificationSource != "" {
		// Buscar o contexto novamente para obter a versão atualizada
		updatedContext, err := h.contextService.GetContextByID(ctx, cmd.ContextID)
		if err != nil {
			// Apenas registrar o erro, mas não falhar a operação já que o nível foi atualizado
			h.auditLogger.LogEvent(ctx, services.AuditEvent{
				EventType:   "UPDATE_VERIFICATION_SOURCE_FAILED",
				ResourceID:  cmd.ContextID.String(),
				ResourceType: "IDENTITY_CONTEXT",
				UserID:      cmd.RequestedBy,
				Timestamp:   time.Now(),
				Details: map[string]interface{}{
					"operation_id":       operationID,
					"context_id":         cmd.ContextID,
					"error":              err.Error(),
				},
			})
		} else {
			// Adicionar fonte de verificação aos metadados
			if updatedContext.Metadata == nil {
				updatedContext.Metadata = make(map[string]interface{})
			}
			
			verificationHistory := []map[string]interface{}{}
			
			// Verificar se já existe histórico de verificação
			if history, ok := updatedContext.Metadata["verification_history"].([]map[string]interface{}); ok {
				verificationHistory = history
			}
			
			// Adicionar nova entrada ao histórico
			verificationHistory = append(verificationHistory, map[string]interface{}{
				"timestamp":          time.Now().UTC(),
				"verification_level": cmd.VerificationLevel,
				"source":             cmd.VerificationSource,
				"reason":             cmd.Reason,
				"requested_by":       cmd.RequestedBy,
			})
			
			updatedContext.Metadata["verification_history"] = verificationHistory
			updatedContext.Metadata["last_verification_source"] = cmd.VerificationSource
			updatedContext.Metadata["last_verification_timestamp"] = time.Now().UTC()
			
			// Atualizar o contexto com os novos metadados
			if err := h.contextService.UpdateContext(ctx, updatedContext); err != nil {
				// Apenas registrar o erro, mas não falhar a operação já que o nível foi atualizado
				h.auditLogger.LogEvent(ctx, services.AuditEvent{
					EventType:   "UPDATE_VERIFICATION_METADATA_FAILED",
					ResourceID:  cmd.ContextID.String(),
					ResourceType: "IDENTITY_CONTEXT",
					UserID:      cmd.RequestedBy,
					Timestamp:   time.Now(),
					Details: map[string]interface{}{
						"operation_id": operationID,
						"context_id":   cmd.ContextID,
						"error":        err.Error(),
					},
				})
			}
		}
	}
	
	// Registrar sucesso da operação
	h.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "UPDATE_VERIFICATION_LEVEL_SUCCEEDED",
		ResourceID:  cmd.ContextID.String(),
		ResourceType: "IDENTITY_CONTEXT",
		UserID:      cmd.RequestedBy,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"operation_id":              operationID,
			"context_id":                cmd.ContextID,
			"previous_verification_level": currentContext.VerificationLevel,
			"new_verification_level":    cmd.VerificationLevel,
			"reason":                    cmd.Reason,
			"verification_source":       cmd.VerificationSource,
			"duration_ms":               time.Since(startTime).Milliseconds(),
		},
	})
	
	return nil
}

// isValidVerificationLevel verifica se o nível de verificação é válido
func isValidVerificationLevel(level models.VerificationLevel) bool {
	validLevels := map[models.VerificationLevel]bool{
		models.VerificationNone:     true,
		models.VerificationBasic:    true,
		models.VerificationStandard: true,
		models.VerificationEnhanced: true,
		models.VerificationComplete: true,
	}
	
	return validLevels[level]
}

// verificationLevelValue retorna um valor numérico para o nível de verificação para comparação
func verificationLevelValue(level models.VerificationLevel) int {
	levelValues := map[models.VerificationLevel]int{
		models.VerificationNone:     0,
		models.VerificationBasic:    1,
		models.VerificationStandard: 2,
		models.VerificationEnhanced: 3,
		models.VerificationComplete: 4,
	}
	
	return levelValues[level]
}