/**
 * @file create_identity.go
 * @description Comando de aplicação para criação de identidade
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

// CreateIdentityCommand representa o comando para criação de identidade
type CreateIdentityCommand struct {
	PrimaryKeyType  models.PrimaryKeyType `json:"primary_key_type" validate:"required"`
	PrimaryKeyValue string                `json:"primary_key_value" validate:"required"`
	MasterPersonID  *uuid.UUID            `json:"master_person_id,omitempty"`
	InitialStatus   models.IdentityStatus `json:"initial_status,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	InitialContext  *InitialContextData   `json:"initial_context,omitempty"`
}

// InitialContextData representa dados para criação de contexto inicial
type InitialContextData struct {
	ContextType  models.ContextType                     `json:"context_type" validate:"required"`
	Attributes   map[string]InitialAttributeData        `json:"attributes,omitempty"`
}

// InitialAttributeData representa dados para criação de atributo inicial
type InitialAttributeData struct {
	Value           string                `json:"value" validate:"required"`
	SensitivityLevel models.SensitivityLevel `json:"sensitivity_level,omitempty"`
	VerifyAttribute bool                  `json:"verify_attribute,omitempty"`
}

// CreateIdentityResult representa o resultado da criação de identidade
type CreateIdentityResult struct {
	IdentityID   string    `json:"identity_id"`
	ContextID    string    `json:"context_id,omitempty"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	AttributeCount int     `json:"attribute_count,omitempty"`
}

// CreateIdentityHandler é o handler para o comando de criação de identidade
type CreateIdentityHandler struct {
	identityService *services.IdentityService
	contextService  *services.ContextService
}

// NewCreateIdentityHandler cria uma nova instância do handler
func NewCreateIdentityHandler(
	identityService *services.IdentityService,
	contextService *services.ContextService,
) *CreateIdentityHandler {
	return &CreateIdentityHandler{
		identityService: identityService,
		contextService:  contextService,
	}
}

// Handle executa o comando de criação de identidade
func (h *CreateIdentityHandler) Handle(ctx context.Context, cmd CreateIdentityCommand) (*CreateIdentityResult, error) {
	// Validar comando
	if cmd.PrimaryKeyType == "" || cmd.PrimaryKeyValue == "" {
		return nil, fmt.Errorf("tipo e valor de chave primária são obrigatórios")
	}
	
	// Definir status inicial se não foi especificado
	initialStatus := cmd.InitialStatus
	if initialStatus == "" {
		initialStatus = models.StatusPendingVerification
	}
	
	// Criar identidade
	identity, err := h.identityService.CreateIdentity(
		ctx,
		cmd.PrimaryKeyType,
		cmd.PrimaryKeyValue,
		cmd.MasterPersonID,
		cmd.Metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar identidade: %w", err)
	}
	
	// Definir status inicial se diferente do padrão
	if initialStatus != models.StatusPendingVerification {
		identity.UpdateStatus(initialStatus)
		if err := h.identityService.UpdateIdentityStatus(ctx, identity.ID, initialStatus); err != nil {
			// Registrar erro, mas continuar com a execução
			fmt.Printf("Erro ao atualizar status inicial: %v\n", err)
		}
	}
	
	result := &CreateIdentityResult{
		IdentityID: identity.ID.String(),
		Status:     string(identity.Status),
		CreatedAt:  identity.CreatedAt,
	}
	
	// Criar contexto inicial se especificado
	if cmd.InitialContext != nil && cmd.InitialContext.ContextType != "" {
		// Criar contexto
		context, err := h.identityService.AddContext(
			ctx,
			identity.ID,
			cmd.InitialContext.ContextType,
		)
		if err != nil {
			// Retornar identidade criada mesmo se houver erro no contexto
			return result, fmt.Errorf("identidade criada, mas erro ao criar contexto inicial: %w", err)
		}
		
		result.ContextID = context.ID.String()
		
		// Adicionar atributos ao contexto
		attributeCount := 0
		if cmd.InitialContext.Attributes != nil {
			for key, attrData := range cmd.InitialContext.Attributes {
				// Definir nível de sensibilidade padrão se não especificado
				sensitivityLevel := attrData.SensitivityLevel
				if sensitivityLevel == "" {
					sensitivityLevel = models.SensitivityLow
				}
				
				// Criar atributo
				_, err := h.identityService.AddContextAttribute(
					ctx,
					context.ID,
					key,
					attrData.Value,
					sensitivityLevel,
					attrData.VerifyAttribute,
				)
				if err != nil {
					fmt.Printf("Erro ao adicionar atributo %s: %v\n", key, err)
					continue
				}
				
				attributeCount++
			}
		}
		
		result.AttributeCount = attributeCount
	}
	
	return result, nil
}