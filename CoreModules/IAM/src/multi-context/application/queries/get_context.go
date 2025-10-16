/**
 * @file get_context.go
 * @description Query de aplicação para recuperação de contexto de identidade
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package queries

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	
	"innovabiz/iam/src/multi-context/domain/services"
)

// GetContextQuery representa a query para obter um contexto
type GetContextQuery struct {
	ContextID        string `json:"context_id" validate:"required"`
	IncludeAttributes bool   `json:"include_attributes"`
}

// GetContextHandler é o handler para a query de obtenção de contexto
type GetContextHandler struct {
	contextService *services.ContextService
}

// NewGetContextHandler cria uma nova instância do handler
func NewGetContextHandler(contextService *services.ContextService) *GetContextHandler {
	return &GetContextHandler{
		contextService: contextService,
	}
}

// Handle executa a query de obtenção de contexto
func (h *GetContextHandler) Handle(ctx context.Context, query GetContextQuery) (*ContextDTO, error) {
	// Validar query
	if query.ContextID == "" {
		return nil, fmt.Errorf("ID de contexto é obrigatório")
	}
	
	// Converter ID do contexto
	contextID, err := uuid.Parse(query.ContextID)
	if err != nil {
		return nil, fmt.Errorf("ID de contexto inválido: %w", err)
	}
	
	// Obter contexto do serviço de domínio
	identityContext, err := h.contextService.GetContext(ctx, contextID)
	if err != nil {
		return nil, err
	}
	
	// Converter para DTO
	contextDTO := &ContextDTO{
		ID:               identityContext.ID.String(),
		ContextType:      string(identityContext.ContextType),
		Status:           string(identityContext.Status),
		VerificationLevel: string(identityContext.VerificationLevel),
		TrustScore:       identityContext.TrustScore,
		CreatedAt:        identityContext.CreatedAt,
		UpdatedAt:        identityContext.UpdatedAt,
		Metadata:         identityContext.Metadata,
	}
	
	// Adicionar atributos se solicitado
	if query.IncludeAttributes && len(identityContext.Attributes) > 0 {
		contextDTO.Attributes = make([]AttributeDTO, 0, len(identityContext.Attributes))
		
		for _, attr := range identityContext.Attributes {
			attributeDTO := AttributeDTO{
				ID:                attr.ID.String(),
				AttributeKey:      attr.AttributeKey,
				AttributeValue:    attr.AttributeValue,
				SensitivityLevel:  string(attr.SensitivityLevel),
				VerificationStatus: string(attr.VerificationStatus),
				VerificationSource: attr.VerificationSource,
				CreatedAt:         attr.CreatedAt,
				UpdatedAt:         attr.UpdatedAt,
				Metadata:          attr.Metadata,
			}
			
			contextDTO.Attributes = append(contextDTO.Attributes, attributeDTO)
		}
	}
	
	return contextDTO, nil
}