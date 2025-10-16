/**
 * @file get_identity.go
 * @description Query de aplicação para recuperação de identidade
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	
	"innovabiz/iam/src/multi-context/domain/models"
	"innovabiz/iam/src/multi-context/domain/services"
)

// IdentityDTO representa dados de identidade para transferência
type IdentityDTO struct {
	ID             string                 `json:"identity_id"`
	PrimaryKeyType string                 `json:"primary_key_type"`
	PrimaryKeyValue string                `json:"primary_key_value"`
	MasterPersonID string                 `json:"master_person_id,omitempty"`
	Status         string                 `json:"status"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Contexts       []ContextDTO           `json:"contexts,omitempty"`
}

// ContextDTO representa dados de contexto para transferência
type ContextDTO struct {
	ID               string                 `json:"context_id"`
	ContextType      string                 `json:"context_type"`
	Status           string                 `json:"status"`
	VerificationLevel string                `json:"verification_level"`
	TrustScore       float64                `json:"trust_score"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	Attributes       []AttributeDTO         `json:"attributes,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// AttributeDTO representa dados de atributo para transferência
type AttributeDTO struct {
	ID                string                 `json:"attribute_id"`
	AttributeKey      string                 `json:"attribute_key"`
	AttributeValue    string                 `json:"attribute_value"`
	SensitivityLevel  string                 `json:"sensitivity_level"`
	VerificationStatus string                `json:"verification_status"`
	VerificationSource string                `json:"verification_source,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// GetIdentityQuery representa a query para obter uma identidade
type GetIdentityQuery struct {
	IdentityID      string `json:"identity_id,omitempty"`
	PrimaryKeyType  string `json:"primary_key_type,omitempty"`
	PrimaryKeyValue string `json:"primary_key_value,omitempty"`
	IncludeContexts bool   `json:"include_contexts"`
	IncludeAttributes bool `json:"include_attributes"`
}

// GetIdentityHandler é o handler para a query de obtenção de identidade
type GetIdentityHandler struct {
	identityService *services.IdentityService
}

// NewGetIdentityHandler cria uma nova instância do handler
func NewGetIdentityHandler(identityService *services.IdentityService) *GetIdentityHandler {
	return &GetIdentityHandler{
		identityService: identityService,
	}
}

// Handle executa a query de obtenção de identidade
func (h *GetIdentityHandler) Handle(ctx context.Context, query GetIdentityQuery) (*IdentityDTO, error) {
	var identity *models.Identity
	var err error
	
	// Buscar por ID ou por chave primária
	if query.IdentityID != "" {
		identityID, err := uuid.Parse(query.IdentityID)
		if err != nil {
			return nil, fmt.Errorf("ID de identidade inválido: %w", err)
		}
		
		identity, err = h.identityService.GetIdentity(ctx, identityID, query.IncludeContexts)
	} else if query.PrimaryKeyType != "" && query.PrimaryKeyValue != "" {
		identity, err = h.identityService.GetIdentityByPrimaryKey(
			ctx, 
			models.PrimaryKeyType(query.PrimaryKeyType),
			query.PrimaryKeyValue,
			query.IncludeContexts,
		)
	} else {
		return nil, fmt.Errorf("ID de identidade ou chave primária (tipo e valor) são obrigatórios")
	}
	
	if err != nil {
		return nil, err
	}
	
	// Converter para DTO
	identityDTO := mapIdentityToDTO(identity, query.IncludeContexts, query.IncludeAttributes)
	
	return identityDTO, nil
}

// mapIdentityToDTO converte um modelo de identidade para DTO
func mapIdentityToDTO(
	identity *models.Identity, 
	includeContexts bool,
	includeAttributes bool,
) *IdentityDTO {
	dto := &IdentityDTO{
		ID:             identity.ID.String(),
		PrimaryKeyType: string(identity.PrimaryKeyType),
		PrimaryKeyValue: identity.PrimaryKeyValue,
		Status:         string(identity.Status),
		CreatedAt:      identity.CreatedAt,
		UpdatedAt:      identity.UpdatedAt,
		Metadata:       identity.Metadata,
	}
	
	if identity.MasterPersonID != nil {
		dto.MasterPersonID = identity.MasterPersonID.String()
	}
	
	if includeContexts && len(identity.Contexts) > 0 {
		dto.Contexts = make([]ContextDTO, 0, len(identity.Contexts))
		
		for _, ctx := range identity.Contexts {
			contextDTO := ContextDTO{
				ID:               ctx.ID.String(),
				ContextType:      string(ctx.ContextType),
				Status:           string(ctx.Status),
				VerificationLevel: string(ctx.VerificationLevel),
				TrustScore:       ctx.TrustScore,
				CreatedAt:        ctx.CreatedAt,
				UpdatedAt:        ctx.UpdatedAt,
				Metadata:         ctx.Metadata,
			}
			
			if includeAttributes && len(ctx.Attributes) > 0 {
				contextDTO.Attributes = make([]AttributeDTO, 0, len(ctx.Attributes))
				
				for _, attr := range ctx.Attributes {
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
			
			dto.Contexts = append(dto.Contexts, contextDTO)
		}
	}
	
	return dto
}