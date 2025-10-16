/**
 * @file verify_attribute.go
 * @description Comando e handler para verificação de atributos contextuais
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

// VerifyAttributeCommand representa o comando para verificar um atributo contextual
type VerifyAttributeCommand struct {
	AttributeID        uuid.UUID                  // ID do atributo a ser verificado
	VerificationStatus models.VerificationStatus   // Resultado da verificação
	VerificationSource string                      // Fonte da verificação
	Notes              string                      // Notas ou comentários sobre a verificação
	EvidenceMetadata   map[string]interface{}      // Metadados de evidência da verificação
	RequestedBy        string                      // Utilizador ou sistema que solicitou a verificação
}

// VerifyAttributeHandler gerencia a verificação de atributos contextuais
type VerifyAttributeHandler struct {
	attributeService *services.AttributeService
	contextService   *services.ContextService
	trustGuardClient services.TrustGuardClient
	auditLogger      services.AuditLogger
}

// NewVerifyAttributeHandler cria uma nova instância do handler
func NewVerifyAttributeHandler(
	attributeService *services.AttributeService,
	contextService *services.ContextService,
	trustGuardClient services.TrustGuardClient,
	auditLogger services.AuditLogger,
) *VerifyAttributeHandler {
	return &VerifyAttributeHandler{
		attributeService: attributeService,
		contextService:   contextService,
		trustGuardClient: trustGuardClient,
		auditLogger:      auditLogger,
	}
}

// Handle processa o comando de verificação de atributo
func (h *VerifyAttributeHandler) Handle(ctx context.Context, cmd VerifyAttributeCommand) (*models.ContextAttribute, error) {
	// Registrar início da operação para rastreabilidade
	startTime := time.Now()
	operationID := uuid.New()
	
	h.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "VERIFY_ATTRIBUTE_INITIATED",
		ResourceID:  cmd.AttributeID.String(),
		ResourceType: "CONTEXT_ATTRIBUTE",
		UserID:      cmd.RequestedBy,
		Timestamp:   startTime,
		Details: map[string]interface{}{
			"operation_id":        operationID,
			"attribute_id":        cmd.AttributeID,
			"verification_status": cmd.VerificationStatus,
			"verification_source": cmd.VerificationSource,
		},
	})
	
	// Validar status de verificação
	if !isValidVerificationStatus(cmd.VerificationStatus) {
		err := fmt.Errorf("status de verificação inválido: %s", cmd.VerificationStatus)
		
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "VERIFY_ATTRIBUTE_FAILED",
			ResourceID:  cmd.AttributeID.String(),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":        operationID,
				"attribute_id":        cmd.AttributeID,
				"verification_status": cmd.VerificationStatus,
				"error":               err.Error(),
				"duration_ms":         time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, err
	}
	
	// Validar que o status não é "não verificado" (não faz sentido neste comando)
	if cmd.VerificationStatus == models.VerificationStatusUnverified {
		err := fmt.Errorf("não é possível definir o status como 'não verificado' através deste comando")
		
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "VERIFY_ATTRIBUTE_FAILED",
			ResourceID:  cmd.AttributeID.String(),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":        operationID,
				"attribute_id":        cmd.AttributeID,
				"verification_status": cmd.VerificationStatus,
				"error":               err.Error(),
				"duration_ms":         time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, err
	}
	
	// Validar fonte de verificação
	if cmd.VerificationSource == "" {
		err := fmt.Errorf("a fonte de verificação é obrigatória")
		
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "VERIFY_ATTRIBUTE_FAILED",
			ResourceID:  cmd.AttributeID.String(),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":        operationID,
				"attribute_id":        cmd.AttributeID,
				"verification_status": cmd.VerificationStatus,
				"error":               err.Error(),
				"duration_ms":         time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, err
	}
	
	// Buscar o atributo existente
	attribute, err := h.attributeService.GetAttributeByID(ctx, cmd.AttributeID)
	if err != nil {
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "VERIFY_ATTRIBUTE_FAILED",
			ResourceID:  cmd.AttributeID.String(),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":        operationID,
				"attribute_id":        cmd.AttributeID,
				"verification_status": cmd.VerificationStatus,
				"error":               err.Error(),
				"duration_ms":         time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, fmt.Errorf("erro ao buscar atributo: %w", err)
	}
	
	// Verificar se o contexto associado está ativo
	context, err := h.contextService.GetContextByID(ctx, attribute.ContextID)
	if err != nil {
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "VERIFY_ATTRIBUTE_FAILED",
			ResourceID:  cmd.AttributeID.String(),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":        operationID,
				"attribute_id":        cmd.AttributeID,
				"context_id":          attribute.ContextID,
				"error":               err.Error(),
				"duration_ms":         time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, fmt.Errorf("erro ao verificar contexto: %w", err)
	}
	
	if context.Status != models.ContextStatusActive {
		err := fmt.Errorf("não é possível verificar atributos de um contexto não ativo")
		
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "VERIFY_ATTRIBUTE_FAILED",
			ResourceID:  cmd.AttributeID.String(),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":        operationID,
				"attribute_id":        cmd.AttributeID,
				"context_id":          attribute.ContextID,
				"context_status":      context.Status,
				"error":               err.Error(),
				"duration_ms":         time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, err
	}
	
	// Registrar o status anterior para auditoria
	previousStatus := attribute.VerificationStatus
	
	// Atualizar o status e fonte de verificação
	attribute.VerificationStatus = cmd.VerificationStatus
	attribute.VerificationSource = cmd.VerificationSource
	attribute.UpdatedAt = time.Now().UTC()
	
	// Atualizar metadados com informações de verificação
	if attribute.Metadata == nil {
		attribute.Metadata = make(map[string]interface{})
	}
	
	// Adicionar histórico de verificação
	verificationHistory := []map[string]interface{}{}
	if history, ok := attribute.Metadata["verification_history"].([]map[string]interface{}); ok {
		verificationHistory = history
	}
	
	verificationEntry := map[string]interface{}{
		"timestamp":          time.Now().UTC(),
		"verification_status": cmd.VerificationStatus,
		"verification_source": cmd.VerificationSource,
		"requested_by":        cmd.RequestedBy,
	}
	
	if cmd.Notes != "" {
		verificationEntry["notes"] = cmd.Notes
	}
	
	if cmd.EvidenceMetadata != nil && len(cmd.EvidenceMetadata) > 0 {
		verificationEntry["evidence"] = cmd.EvidenceMetadata
	}
	
	verificationHistory = append(verificationHistory, verificationEntry)
	attribute.Metadata["verification_history"] = verificationHistory
	attribute.Metadata["last_verification"] = verificationEntry
	
	// Persistir as alterações
	if err := h.attributeService.UpdateAttribute(ctx, attribute); err != nil {
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "VERIFY_ATTRIBUTE_FAILED",
			ResourceID:  cmd.AttributeID.String(),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":        operationID,
				"attribute_id":        cmd.AttributeID,
				"error":               err.Error(),
				"duration_ms":         time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, fmt.Errorf("erro ao atualizar atributo: %w", err)
	}
	
	// Recalcular pontuação de confiança do contexto via TrustGuard
	go func() {
		// Usar um novo contexto para a operação em background
		bgCtx := context.Background()
		
		// Obter todos os atributos do contexto
		attributes, err := h.attributeService.ListAttributesByContext(bgCtx, attribute.ContextID)
		if err != nil {
			// Apenas registrar o erro, mas não falhar a operação
			h.auditLogger.LogEvent(bgCtx, services.AuditEvent{
				EventType:   "TRUST_SCORE_CALCULATION_FAILED",
				ResourceID:  attribute.ContextID.String(),
				ResourceType: "IDENTITY_CONTEXT",
				UserID:      "system",
				Timestamp:   time.Now(),
				Details: map[string]interface{}{
					"trigger_operation": operationID,
					"attribute_id":      attribute.ID,
					"context_id":        attribute.ContextID,
					"error":             err.Error(),
				},
			})
			return
		}
		
		// Avaliar a pontuação de confiança com base nos atributos
		trustScore, err := h.trustGuardClient.EvaluateContextTrust(bgCtx, context, attributes)
		if err != nil {
			h.auditLogger.LogEvent(bgCtx, services.AuditEvent{
				EventType:   "TRUST_SCORE_CALCULATION_FAILED",
				ResourceID:  attribute.ContextID.String(),
				ResourceType: "IDENTITY_CONTEXT",
				UserID:      "system",
				Timestamp:   time.Now(),
				Details: map[string]interface{}{
					"trigger_operation": operationID,
					"attribute_id":      attribute.ID,
					"context_id":        attribute.ContextID,
					"error":             err.Error(),
				},
			})
			return
		}
		
		// Atualizar a pontuação de confiança do contexto
		if err := h.contextService.UpdateTrustScore(bgCtx, attribute.ContextID, trustScore); err != nil {
			h.auditLogger.LogEvent(bgCtx, services.AuditEvent{
				EventType:   "TRUST_SCORE_UPDATE_FAILED",
				ResourceID:  attribute.ContextID.String(),
				ResourceType: "IDENTITY_CONTEXT",
				UserID:      "system",
				Timestamp:   time.Now(),
				Details: map[string]interface{}{
					"trigger_operation": operationID,
					"attribute_id":      attribute.ID,
					"context_id":        attribute.ContextID,
					"trust_score":       trustScore,
					"error":             err.Error(),
				},
			})
			return
		}
		
		h.auditLogger.LogEvent(bgCtx, services.AuditEvent{
			EventType:   "TRUST_SCORE_UPDATED",
			ResourceID:  attribute.ContextID.String(),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      "system",
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"trigger_operation":    operationID,
				"attribute_id":         attribute.ID,
				"context_id":           attribute.ContextID,
				"trust_score":          trustScore,
				"verification_trigger": cmd.AttributeID,
			},
		})
	}()
	
	// Registrar sucesso da operação
	h.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "VERIFY_ATTRIBUTE_SUCCEEDED",
		ResourceID:  attribute.ID.String(),
		ResourceType: "CONTEXT_ATTRIBUTE",
		UserID:      cmd.RequestedBy,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"operation_id":           operationID,
			"attribute_id":           attribute.ID,
			"attribute_key":          attribute.AttributeKey,
			"previous_status":        previousStatus,
			"new_status":             attribute.VerificationStatus,
			"verification_source":    attribute.VerificationSource,
			"context_id":             attribute.ContextID,
			"duration_ms":            time.Since(startTime).Milliseconds(),
		},
	})
	
	return attribute, nil
}