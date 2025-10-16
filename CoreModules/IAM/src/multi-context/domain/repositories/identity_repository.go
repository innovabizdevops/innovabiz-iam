/**
 * @file identity_repository.go
 * @description Interface de repositório para gerenciamento de identidades multi-contexto
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package repositories

import (
	"context"

	"github.com/google/uuid"
	
	"innovabiz/iam/src/multi-context/domain/models"
)

// IdentityRepository define as operações de persistência para identidades
type IdentityRepository interface {
	// Create persiste uma nova identidade
	Create(ctx context.Context, identity *models.Identity) error
	
	// Update atualiza uma identidade existente
	Update(ctx context.Context, identity *models.Identity) error
	
	// GetByID recupera uma identidade por seu ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.Identity, error)
	
	// GetByPrimaryKey recupera uma identidade por sua chave primária
	GetByPrimaryKey(ctx context.Context, keyType models.PrimaryKeyType, keyValue string) (*models.Identity, error)
	
	// Delete remove uma identidade (logicamente)
	Delete(ctx context.Context, id uuid.UUID) error
	
	// List lista identidades com filtros e paginação
	List(ctx context.Context, filter IdentityFilter, page, pageSize int) ([]*models.Identity, int, error)
	
	// LoadContexts carrega os contextos para uma identidade específica
	LoadContexts(ctx context.Context, identity *models.Identity) error
}

// IdentityFilter define filtros para busca de identidades
type IdentityFilter struct {
	// Status da identidade para filtro
	Status []models.IdentityStatus
	
	// Tipo de chave primária para filtro
	PrimaryKeyType []models.PrimaryKeyType
	
	// Texto para busca em campos de texto
	SearchText string
	
	// Filtro por data de criação (início)
	CreatedStart *string
	
	// Filtro por data de criação (fim)
	CreatedEnd *string
	
	// Ordem da listagem
	OrderBy string
	
	// Direção da ordenação (asc/desc)
	OrderDirection string
}