/**
 * @file add_context.go
 * @description Comando de aplicação para adicionar um contexto a uma identidade existente
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

// AddContextCommand representa o comando para adicionar contexto
type AddContextCommand struct {
	IdentityID    string                `json:"identity_id" validate:"required"`
	ContextType   models.ContextType     `json:"context_type" validate:"required"`
	InitialAttributes map[string]InitialAttributeData `json:"initial_attributes,omitempty"`
	MappedFromContextID string           `json:"mapped_from_context_id,omitempty"`
}

// AddContextResult representa o resultado da adição de contexto
type AddContextResult struct {
	ContextID       string    `json:"context_id"`
	IdentityID      string    `json:"identity_id"`
	ContextType     string    `json:"context_type"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	AttributeCount  int       `json:"attribute_count"`
	TrustScore      float64   `json:"trust_score"`
	VerificationLevel string  `json:"verification_level"`
}

// AddContextHandler é o handler para o comando de adição de contexto
type AddContextHandler struct {
	identityService *services.IdentityService
	contextService  *services.ContextService
}

// NewAddContextHandler cria uma nova instância do handler
func NewAddContextHandler(
	identityService *services.IdentityService,
	contextService *services.ContextService,
) *AddContextHandler {
	return &AddContextHandler{
		identityService: identityService,
		contextService:  contextService,
	}
}

// Handle executa o comando de adição de contexto
func (h *AddContextHandler) Handle(ctx context.Context, cmd AddContextCommand) (*AddContextResult, error) {
	// Validar comando
	if cmd.IdentityID == "" || cmd.ContextType == "" {
		return nil, fmt.Errorf("ID de identidade e tipo de contexto são obrigatórios")
	}
	
	// Converter ID da identidade
	identityID, err := uuid.Parse(cmd.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("ID de identidade inválido: %w", err)
	}
	
	var identityContext *models.IdentityContext
	
	// Verificar se estamos mapeando de outro contexto
	if cmd.MappedFromContextID != "" {
		// Converter ID do contexto origem
		mappedFromContextID, err := uuid.Parse(cmd.MappedFromContextID)
		if err != nil {
			return nil, fmt.Errorf("ID de contexto origem inválido: %w", err)
		}
		
		// Obter contexto origem para determinar seu tipo
		sourceContext, err := h.contextService.GetContext(ctx, mappedFromContextID)
		if err != nil {
			return nil, fmt.Errorf("erro ao obter contexto origem: %w", err)
		}
		
		// Mapear contextos
		identityContext, err = h.contextService.MapContexts(
			ctx,
			identityID,
			sourceContext.ContextType,
			cmd.ContextType,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao mapear contextos: %w", err)
		}
	} else {
		// Adicionar novo contexto sem mapeamento
		identityContext, err = h.identityService.AddContext(
			ctx,
			identityID,
			cmd.ContextType,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao adicionar contexto: %w", err)
		}
		
		// Adicionar atributos iniciais se especificados
		if cmd.InitialAttributes != nil {
			for key, attrData := range cmd.InitialAttributes {
				// Definir nível de sensibilidade padrão se não especificado
				sensitivityLevel := attrData.SensitivityLevel
				if sensitivityLevel == "" {
					sensitivityLevel = models.SensitivityLow
				}
				
				// Criar atributo
				_, err := h.identityService.AddContextAttribute(
					ctx,
					identityContext.ID,
					key,
					attrData.Value,
					sensitivityLevel,
					attrData.VerifyAttribute,
				)
				if err != nil {
					// Registrar erro, mas continuar com outros atributos
					fmt.Printf("Erro ao adicionar atributo %s: %v\n", key, err)
				}
			}
		}
	}
	
	// Carregar atributos para contagem
	if err := h.contextService.GetContext(ctx, identityContext.ID); err != nil {
		// Continuar mesmo se houver erro ao carregar atributos
		fmt.Printf("Aviso: erro ao recarregar contexto com atributos: %v\n", err)
	}
	
	// Preparar resultado
	result := &AddContextResult{
		ContextID:        identityContext.ID.String(),
		IdentityID:       identityID.String(),
		ContextType:      string(identityContext.ContextType),
		Status:           string(identityContext.Status),
		CreatedAt:        identityContext.CreatedAt,
		AttributeCount:   len(identityContext.Attributes),
		TrustScore:       identityContext.TrustScore,
		VerificationLevel: string(identityContext.VerificationLevel),
	}
	
	return result, nil
}