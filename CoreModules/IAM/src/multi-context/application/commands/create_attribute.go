/**
 * @file create_attribute.go
 * @description Comando e handler para criação de atributos contextuais
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

// CreateAttributeCommand representa o comando para criar um atributo contextual
type CreateAttributeCommand struct {
	ContextID         uuid.UUID                 // ID do contexto ao qual o atributo pertence
	AttributeKey      string                    // Chave do atributo
	AttributeValue    string                    // Valor do atributo
	SensitivityLevel  models.SensitivityLevel   // Nível de sensibilidade do atributo
	VerificationStatus models.VerificationStatus // Status de verificação inicial
	VerificationSource string                    // Fonte de verificação (opcional)
	Metadata          map[string]interface{}     // Metadados adicionais
	RequestedBy       string                     // Utilizador ou sistema que solicitou a criação
}

// CreateAttributeHandler gerencia a criação de atributos contextuais
type CreateAttributeHandler struct {
	attributeService *services.AttributeService
	contextService   *services.ContextService
	auditLogger      services.AuditLogger
}

// NewCreateAttributeHandler cria uma nova instância do handler
func NewCreateAttributeHandler(
	attributeService *services.AttributeService,
	contextService *services.ContextService,
	auditLogger services.AuditLogger,
) *CreateAttributeHandler {
	return &CreateAttributeHandler{
		attributeService: attributeService,
		contextService:   contextService,
		auditLogger:      auditLogger,
	}
}

// Handle processa o comando de criação de atributo
func (h *CreateAttributeHandler) Handle(ctx context.Context, cmd CreateAttributeCommand) (*models.ContextAttribute, error) {
	// Registrar início da operação para rastreabilidade
	startTime := time.Now()
	operationID := uuid.New()
	
	h.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "CREATE_ATTRIBUTE_INITIATED",
		ResourceID:  cmd.ContextID.String(),
		ResourceType: "IDENTITY_CONTEXT",
		UserID:      cmd.RequestedBy,
		Timestamp:   startTime,
		Details: map[string]interface{}{
			"operation_id":        operationID,
			"context_id":          cmd.ContextID,
			"attribute_key":       cmd.AttributeKey,
			"sensitivity_level":   cmd.SensitivityLevel,
			"verification_status": cmd.VerificationStatus,
		},
	})
	
	// Validar dados do comando
	if err := validateCreateAttributeCommand(cmd); err != nil {
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "CREATE_ATTRIBUTE_FAILED",
			ResourceID:  cmd.ContextID.String(),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":      operationID,
				"context_id":        cmd.ContextID,
				"attribute_key":     cmd.AttributeKey,
				"error":             err.Error(),
				"duration_ms":       time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, err
	}
	
	// Verificar se o contexto existe e está ativo
	contextExists, err := h.contextService.GetContextByID(ctx, cmd.ContextID)
	if err != nil {
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "CREATE_ATTRIBUTE_FAILED",
			ResourceID:  cmd.ContextID.String(),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":      operationID,
				"context_id":        cmd.ContextID,
				"attribute_key":     cmd.AttributeKey,
				"error":             err.Error(),
				"duration_ms":       time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, fmt.Errorf("erro ao verificar contexto: %w", err)
	}
	
	if contextExists.Status != models.ContextStatusActive {
		err := fmt.Errorf("não é possível adicionar atributos a um contexto não ativo")
		
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "CREATE_ATTRIBUTE_FAILED",
			ResourceID:  cmd.ContextID.String(),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":      operationID,
				"context_id":        cmd.ContextID,
				"attribute_key":     cmd.AttributeKey,
				"context_status":    contextExists.Status,
				"error":             err.Error(),
				"duration_ms":       time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, err
	}
	
	// Verificar se já existe um atributo com a mesma chave no contexto
	existingAttribute, err := h.attributeService.GetAttributeByKey(ctx, cmd.ContextID, cmd.AttributeKey)
	if err == nil {
		// Atributo já existe
		err := fmt.Errorf("já existe um atributo com a chave '%s' neste contexto", cmd.AttributeKey)
		
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "CREATE_ATTRIBUTE_FAILED",
			ResourceID:  cmd.ContextID.String(),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":       operationID,
				"context_id":         cmd.ContextID,
				"attribute_key":      cmd.AttributeKey,
				"existing_attribute": existingAttribute.ID,
				"error":              err.Error(),
				"duration_ms":        time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, err
	} else if err != models.ErrAttributeNotFound {
		// Ocorreu um erro diferente de "não encontrado"
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "CREATE_ATTRIBUTE_FAILED",
			ResourceID:  cmd.ContextID.String(),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":      operationID,
				"context_id":        cmd.ContextID,
				"attribute_key":     cmd.AttributeKey,
				"error":             err.Error(),
				"duration_ms":       time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, fmt.Errorf("erro ao verificar existência de atributo: %w", err)
	}
	
	// Criar o novo atributo
	now := time.Now().UTC()
	attribute := &models.ContextAttribute{
		ID:                uuid.New(),
		ContextID:         cmd.ContextID,
		AttributeKey:      cmd.AttributeKey,
		AttributeValue:    cmd.AttributeValue,
		SensitivityLevel:  cmd.SensitivityLevel,
		VerificationStatus: cmd.VerificationStatus,
		VerificationSource: cmd.VerificationSource,
		CreatedAt:         now,
		UpdatedAt:         now,
		Metadata:          cmd.Metadata,
	}
	
	// Adicionar metadados de auditoria
	if attribute.Metadata == nil {
		attribute.Metadata = make(map[string]interface{})
	}
	
	attribute.Metadata["created_by"] = cmd.RequestedBy
	attribute.Metadata["created_at_timestamp"] = now
	
	if cmd.VerificationStatus != models.VerificationStatusUnverified {
		attribute.Metadata["initial_verification"] = map[string]interface{}{
			"status": cmd.VerificationStatus,
			"source": cmd.VerificationSource,
			"timestamp": now,
		}
	}
	
	// Persistir o atributo
	if err := h.attributeService.CreateAttribute(ctx, attribute); err != nil {
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "CREATE_ATTRIBUTE_FAILED",
			ResourceID:  cmd.ContextID.String(),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      cmd.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":      operationID,
				"context_id":        cmd.ContextID,
				"attribute_key":     cmd.AttributeKey,
				"error":             err.Error(),
				"duration_ms":       time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, fmt.Errorf("erro ao criar atributo: %w", err)
	}
	
	// Se o atributo tem alta sensibilidade, programar uma verificação em background
	if attribute.SensitivityLevel == models.SensitivityHigh || 
	   attribute.SensitivityLevel == models.SensitivityCritical {
		go h.attributeService.ScheduleAttributeVerification(context.Background(), attribute.ID)
	}
	
	// Registrar sucesso da operação
	h.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "CREATE_ATTRIBUTE_SUCCEEDED",
		ResourceID:  attribute.ID.String(),
		ResourceType: "CONTEXT_ATTRIBUTE",
		UserID:      cmd.RequestedBy,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"operation_id":        operationID,
			"context_id":          cmd.ContextID,
			"attribute_id":        attribute.ID,
			"attribute_key":       attribute.AttributeKey,
			"sensitivity_level":   attribute.SensitivityLevel,
			"verification_status": attribute.VerificationStatus,
			"duration_ms":         time.Since(startTime).Milliseconds(),
		},
	})
	
	return attribute, nil
}

// validateCreateAttributeCommand valida os dados do comando
func validateCreateAttributeCommand(cmd CreateAttributeCommand) error {
	// Validar chave do atributo
	if cmd.AttributeKey == "" {
		return fmt.Errorf("a chave do atributo é obrigatória")
	}
	
	// Validar valor do atributo
	if cmd.AttributeValue == "" {
		return fmt.Errorf("o valor do atributo é obrigatório")
	}
	
	// Validar nível de sensibilidade
	validSensitivity := map[models.SensitivityLevel]bool{
		models.SensitivityLow:      true,
		models.SensitivityMedium:   true,
		models.SensitivityHigh:     true,
		models.SensitivityCritical: true,
	}
	
	if !validSensitivity[cmd.SensitivityLevel] {
		return fmt.Errorf("nível de sensibilidade inválido: %s", cmd.SensitivityLevel)
	}
	
	// Validar status de verificação
	validStatus := map[models.VerificationStatus]bool{
		models.VerificationStatusUnverified:  true,
		models.VerificationStatusPending:     true,
		models.VerificationStatusVerified:    true,
		models.VerificationStatusRejected:    true,
	}
	
	if !validStatus[cmd.VerificationStatus] {
		return fmt.Errorf("status de verificação inválido: %s", cmd.VerificationStatus)
	}
	
	// Validar fonte de verificação quando status é "verificado"
	if cmd.VerificationStatus == models.VerificationStatusVerified && cmd.VerificationSource == "" {
		return fmt.Errorf("a fonte de verificação é obrigatória para atributos verificados")
	}
	
	return nil
}