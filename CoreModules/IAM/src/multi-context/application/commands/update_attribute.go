/**
 * @file update_attribute.go
 * @description Comando e handler para atualização de atributos contextuais
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

// UpdateAttributeCommand representa o comando para atualizar um atributo contextual
type UpdateAttributeCommand struct {
	AttributeID        uuid.UUID                  // ID do atributo a ser atualizado
	AttributeValue     *string                    // Novo valor do atributo (opcional)
	SensitivityLevel   *models.SensitivityLevel   // Novo nível de sensibilidade (opcional)
	VerificationStatus *models.VerificationStatus // Novo status de verificação (opcional)
	VerificationSource *string                    // Nova fonte de verificação (opcional)
	Metadata           map[string]interface{}     // Metadados a serem mesclados (opcional)
	Reason             string                     // Motivo da atualização
	RequestedBy        string                     // Utilizador ou sistema que solicitou a atualização
}

// UpdateAttributeHandler gerencia a atualização de atributos contextuais
type UpdateAttributeHandler struct {
	attributeService *services.AttributeService
	contextService   *services.ContextService
	auditLogger      services.AuditLogger
}

// NewUpdateAttributeHandler cria uma nova instância do handler
func NewUpdateAttributeHandler(
	attributeService *services.AttributeService,
	contextService *services.ContextService,
	auditLogger services.AuditLogger,
) *UpdateAttributeHandler {
	return &UpdateAttributeHandler{
		attributeService: attributeService,
		contextService:   contextService,
		auditLogger:      auditLogger,
	}
}

// Handle processa o comando de atualização de atributo
func (h *UpdateAttributeHandler) Handle(ctx context.Context, cmd UpdateAttributeCommand) (*models.ContextAttribute, error) {
	// Registrar início da operação para rastreabilidade
	startTime := time.Now()
	operationID := uuid.New()
	
	h.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "UPDATE_ATTRIBUTE_INITIATED",
		ResourceID:  cmd.AttributeID.String(),
		ResourceType: "CONTEXT_ATTRIBUTE",
		UserID:      cmd.RequestedBy,
		Timestamp:   startTime,
		Details: map[string]interface{}{
			"operation_id":      operationID,
			"attribute_id":      cmd.AttributeID,
			"reason":            cmd.Reason,
		},
	})
	
	// Verificar se pelo menos um campo para atualização foi fornecido
	if !hasUpdateFields(cmd) {
		err := fmt.Errorf("nenhum campo para atualização foi fornecido")
		
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "UPDATE_ATTRIBUTE_FAILED",
			ResourceID:  cmd.AttributeID.String(),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":      operationID,
				"attribute_id":      cmd.AttributeID,
				"error":             err.Error(),
				"duration_ms":       time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, err
	}
	
	// Buscar o atributo existente
	existingAttribute, err := h.attributeService.GetAttributeByID(ctx, cmd.AttributeID)
	if err != nil {
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "UPDATE_ATTRIBUTE_FAILED",
			ResourceID:  cmd.AttributeID.String(),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":      operationID,
				"attribute_id":      cmd.AttributeID,
				"error":             err.Error(),
				"duration_ms":       time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, fmt.Errorf("erro ao buscar atributo: %w", err)
	}
	
	// Verificar se o contexto associado está ativo
	contextExists, err := h.contextService.GetContextByID(ctx, existingAttribute.ContextID)
	if err != nil {
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "UPDATE_ATTRIBUTE_FAILED",
			ResourceID:  cmd.AttributeID.String(),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":      operationID,
				"attribute_id":      cmd.AttributeID,
				"context_id":        existingAttribute.ContextID,
				"error":             err.Error(),
				"duration_ms":       time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, fmt.Errorf("erro ao verificar contexto: %w", err)
	}
	
	if contextExists.Status != models.ContextStatusActive {
		err := fmt.Errorf("não é possível atualizar atributos de um contexto não ativo")
		
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "UPDATE_ATTRIBUTE_FAILED",
			ResourceID:  cmd.AttributeID.String(),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":      operationID,
				"attribute_id":      cmd.AttributeID,
				"context_id":        existingAttribute.ContextID,
				"context_status":    contextExists.Status,
				"error":             err.Error(),
				"duration_ms":       time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, err
	}
	
	// Registrar as alterações para auditoria
	changes := make(map[string]interface{})
	
	// Aplicar as alterações ao atributo
	if cmd.AttributeValue != nil {
		changes["previous_value"] = existingAttribute.AttributeValue
		changes["new_value"] = *cmd.AttributeValue
		existingAttribute.AttributeValue = *cmd.AttributeValue
	}
	
	if cmd.SensitivityLevel != nil {
		// Validar o nível de sensibilidade
		if !isValidSensitivityLevel(*cmd.SensitivityLevel) {
			err := fmt.Errorf("nível de sensibilidade inválido: %s", *cmd.SensitivityLevel)
			
			h.auditLogger.LogEvent(ctx, services.AuditEvent{
				EventType:   "UPDATE_ATTRIBUTE_FAILED",
				ResourceID:  cmd.AttributeID.String(),
				ResourceType: "CONTEXT_ATTRIBUTE",
				UserID:      cmd.RequestedBy,
				Timestamp:   time.Now(),
				Details: map[string]interface{}{
					"operation_id":      operationID,
					"attribute_id":      cmd.AttributeID,
					"error":             err.Error(),
					"duration_ms":       time.Since(startTime).Milliseconds(),
				},
			})
			
			return nil, err
		}
		
		changes["previous_sensitivity_level"] = existingAttribute.SensitivityLevel
		changes["new_sensitivity_level"] = *cmd.SensitivityLevel
		existingAttribute.SensitivityLevel = *cmd.SensitivityLevel
	}
	
	if cmd.VerificationStatus != nil {
		// Validar o status de verificação
		if !isValidVerificationStatus(*cmd.VerificationStatus) {
			err := fmt.Errorf("status de verificação inválido: %s", *cmd.VerificationStatus)
			
			h.auditLogger.LogEvent(ctx, services.AuditEvent{
				EventType:   "UPDATE_ATTRIBUTE_FAILED",
				ResourceID:  cmd.AttributeID.String(),
				ResourceType: "CONTEXT_ATTRIBUTE",
				UserID:      cmd.RequestedBy,
				Timestamp:   time.Now(),
				Details: map[string]interface{}{
					"operation_id":      operationID,
					"attribute_id":      cmd.AttributeID,
					"error":             err.Error(),
					"duration_ms":       time.Since(startTime).Milliseconds(),
				},
			})
			
			return nil, err
		}
		
		// Validar fonte de verificação quando status é "verificado"
		if *cmd.VerificationStatus == models.VerificationStatusVerified && 
		  (cmd.VerificationSource == nil || *cmd.VerificationSource == "") {
			err := fmt.Errorf("a fonte de verificação é obrigatória para atributos verificados")
			
			h.auditLogger.LogEvent(ctx, services.AuditEvent{
				EventType:   "UPDATE_ATTRIBUTE_FAILED",
				ResourceID:  cmd.AttributeID.String(),
				ResourceType: "CONTEXT_ATTRIBUTE",
				UserID:      cmd.RequestedBy,
				Timestamp:   time.Now(),
				Details: map[string]interface{}{
					"operation_id":      operationID,
					"attribute_id":      cmd.AttributeID,
					"error":             err.Error(),
					"duration_ms":       time.Since(startTime).Milliseconds(),
				},
			})
			
			return nil, err
		}
		
		changes["previous_verification_status"] = existingAttribute.VerificationStatus
		changes["new_verification_status"] = *cmd.VerificationStatus
		existingAttribute.VerificationStatus = *cmd.VerificationStatus
	}
	
	if cmd.VerificationSource != nil {
		changes["previous_verification_source"] = existingAttribute.VerificationSource
		changes["new_verification_source"] = *cmd.VerificationSource
		existingAttribute.VerificationSource = *cmd.VerificationSource
	}
	
	// Atualizar timestamp
	existingAttribute.UpdatedAt = time.Now().UTC()
	
	// Mesclar metadados se fornecidos
	if cmd.Metadata != nil {
		if existingAttribute.Metadata == nil {
			existingAttribute.Metadata = make(map[string]interface{})
		}
		
		// Armazenar histórico de atualizações
		updateHistory := []map[string]interface{}{}
		if history, ok := existingAttribute.Metadata["update_history"].([]map[string]interface{}); ok {
			updateHistory = history
		}
		
		updateEntry := map[string]interface{}{
			"timestamp":   time.Now().UTC(),
			"requested_by": cmd.RequestedBy,
			"reason":      cmd.Reason,
			"changes":     changes,
		}
		
		updateHistory = append(updateHistory, updateEntry)
		existingAttribute.Metadata["update_history"] = updateHistory
		
		// Mesclar os novos metadados com os existentes
		for key, value := range cmd.Metadata {
			existingAttribute.Metadata[key] = value
		}
	}
	
	// Persistir as alterações
	if err := h.attributeService.UpdateAttribute(ctx, existingAttribute); err != nil {
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "UPDATE_ATTRIBUTE_FAILED",
			ResourceID:  cmd.AttributeID.String(),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":      operationID,
				"attribute_id":      cmd.AttributeID,
				"error":             err.Error(),
				"duration_ms":       time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, fmt.Errorf("erro ao atualizar atributo: %w", err)
	}
	
	// Se houver mudança no nível de sensibilidade para alto/crítico, programar verificação
	if cmd.SensitivityLevel != nil && 
	   (*cmd.SensitivityLevel == models.SensitivityHigh || 
	    *cmd.SensitivityLevel == models.SensitivityCritical) {
		go h.attributeService.ScheduleAttributeVerification(context.Background(), existingAttribute.ID)
	}
	
	// Se o valor foi alterado e o atributo já estava verificado, retornar ao estado pendente
	if cmd.AttributeValue != nil && 
	   existingAttribute.VerificationStatus == models.VerificationStatusVerified && 
	   cmd.VerificationStatus == nil {
		// Não usar o contexto original para evitar cancelamento prematuro
		go func() {
			bgCtx := context.Background()
			pendingStatus := models.VerificationStatusPending
			
			// Criar comando para reverter o status
			revertCmd := UpdateAttributeCommand{
				AttributeID:        existingAttribute.ID,
				VerificationStatus: &pendingStatus,
				Reason:             "Valor alterado - verificação anterior invalidada",
				RequestedBy:        "system",
			}
			
			// Executar após um pequeno delay para garantir que a transação original foi concluída
			time.Sleep(100 * time.Millisecond)
			h.Handle(bgCtx, revertCmd)
		}()
	}
	
	// Registrar sucesso da operação
	h.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "UPDATE_ATTRIBUTE_SUCCEEDED",
		ResourceID:  existingAttribute.ID.String(),
		ResourceType: "CONTEXT_ATTRIBUTE",
		UserID:      cmd.RequestedBy,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"operation_id":      operationID,
			"attribute_id":      existingAttribute.ID,
			"changes":           changes,
			"reason":            cmd.Reason,
			"duration_ms":       time.Since(startTime).Milliseconds(),
		},
	})
	
	return existingAttribute, nil
}

// hasUpdateFields verifica se pelo menos um campo para atualização foi fornecido
func hasUpdateFields(cmd UpdateAttributeCommand) bool {
	return cmd.AttributeValue != nil ||
		cmd.SensitivityLevel != nil ||
		cmd.VerificationStatus != nil ||
		cmd.VerificationSource != nil ||
		(cmd.Metadata != nil && len(cmd.Metadata) > 0)
}

// isValidSensitivityLevel verifica se o nível de sensibilidade é válido
func isValidSensitivityLevel(level models.SensitivityLevel) bool {
	validLevels := map[models.SensitivityLevel]bool{
		models.SensitivityLow:      true,
		models.SensitivityMedium:   true,
		models.SensitivityHigh:     true,
		models.SensitivityCritical: true,
	}
	
	return validLevels[level]
}

// isValidVerificationStatus verifica se o status de verificação é válido
func isValidVerificationStatus(status models.VerificationStatus) bool {
	validStatus := map[models.VerificationStatus]bool{
		models.VerificationStatusUnverified:  true,
		models.VerificationStatusPending:     true,
		models.VerificationStatusVerified:    true,
		models.VerificationStatusRejected:    true,
	}
	
	return validStatus[status]
}