/**
 * @file context_repository.go
 * @description Interface de repositório para gerenciamento de contextos de identidade
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package repositories

import (
	"context"

	"github.com/google/uuid"
	
	"innovabiz/iam/src/multi-context/domain/models"
)

// ContextRepository define as operações de persistência para contextos de identidade
type ContextRepository interface {
	// Create persiste um novo contexto de identidade
	Create(ctx context.Context, identityContext *models.IdentityContext) error
	
	// Update atualiza um contexto existente
	Update(ctx context.Context, identityContext *models.IdentityContext) error
	
	// GetByID recupera um contexto por seu ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.IdentityContext, error)
	
	// GetByIdentityAndType recupera um contexto pela identidade e tipo
	GetByIdentityAndType(ctx context.Context, identityID uuid.UUID, contextType models.ContextType) (*models.IdentityContext, error)
	
	// ListByIdentity lista contextos de uma identidade específica
	ListByIdentity(ctx context.Context, identityID uuid.UUID) ([]*models.IdentityContext, error)
	
	// Delete remove um contexto (logicamente)
	Delete(ctx context.Context, id uuid.UUID) error
	
	// LoadAttributes carrega os atributos para um contexto específico
	LoadAttributes(ctx context.Context, identityContext *models.IdentityContext) error
	
	// List lista contextos com filtros e paginação
	List(ctx context.Context, filter ContextFilter, page, pageSize int) ([]*models.IdentityContext, int, error)
	
	// UpdateTrustScore atualiza apenas a pontuação de confiança de um contexto
	UpdateTrustScore(ctx context.Context, contextID uuid.UUID, score float64) error
	
	// UpdateVerificationLevel atualiza apenas o nível de verificação de um contexto
	UpdateVerificationLevel(ctx context.Context, contextID uuid.UUID, level models.VerificationLevel) error
}

// ContextFilter define filtros para busca de contextos
type ContextFilter struct {
	// ID da identidade para filtro
	IdentityID *uuid.UUID
	
	// Tipos de contexto para filtro
	ContextTypes []models.ContextType
	
	// Status do contexto para filtro
	Status []models.ContextStatus
	
	// Nível de verificação mínimo
	MinVerificationLevel *models.VerificationLevel
	
	// Pontuação de confiança mínima
	MinTrustScore *float64
	
	// Filtro por data de criação (início)
	CreatedStart *string
	
	// Filtro por data de criação (fim)
	CreatedEnd *string
	
	// Ordem da listagem
	OrderBy string
	
	// Direção da ordenação (asc/desc)
	OrderDirection string
}