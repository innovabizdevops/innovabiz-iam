/**
 * @file attribute_repository.go
 * @description Interface de repositório para gerenciamento de atributos contextuais
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package repositories

import (
	"context"

	"github.com/google/uuid"
	
	"innovabiz/iam/src/multi-context/domain/models"
)

// AttributeRepository define as operações de persistência para atributos contextuais
type AttributeRepository interface {
	// Create persiste um novo atributo contextual
	Create(ctx context.Context, attribute *models.ContextAttribute) error
	
	// Update atualiza um atributo contextual existente
	Update(ctx context.Context, attribute *models.ContextAttribute) error
	
	// GetByID recupera um atributo por seu ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.ContextAttribute, error)
	
	// GetByContextAndKey recupera um atributo pelo contexto e chave
	GetByContextAndKey(ctx context.Context, contextID uuid.UUID, key string) (*models.ContextAttribute, error)
	
	// ListByContext lista atributos de um contexto específico
	ListByContext(ctx context.Context, contextID uuid.UUID) ([]*models.ContextAttribute, error)
	
	// Delete remove um atributo
	Delete(ctx context.Context, id uuid.UUID) error
	
	// BatchCreate cria múltiplos atributos em uma única operação
	BatchCreate(ctx context.Context, attributes []*models.ContextAttribute) error
	
	// List lista atributos com filtros e paginação
	List(ctx context.Context, filter AttributeFilter, page, pageSize int) ([]*models.ContextAttribute, int, error)
	
	// UpdateVerification atualiza apenas o status e fonte de verificação de um atributo
	UpdateVerification(ctx context.Context, attributeID uuid.UUID, status models.VerificationStatus, source string) error
	
	// SearchAttributes busca atributos por valor em todos os contextos
	SearchAttributes(ctx context.Context, searchValue string, sensitivityLevels []models.SensitivityLevel) ([]*models.ContextAttribute, error)
}

// AttributeFilter define filtros para busca de atributos
type AttributeFilter struct {
	// ID do contexto para filtro
	ContextID *uuid.UUID
	
	// Chaves de atributos para filtro
	AttributeKeys []string
	
	// Níveis de sensibilidade para filtro
	SensitivityLevels []models.SensitivityLevel
	
	// Status de verificação para filtro
	VerificationStatus []models.VerificationStatus
	
	// Fontes de verificação para filtro
	VerificationSources []string
	
	// Texto para busca em valores
	SearchValue string
	
	// Filtro por data de criação (início)
	CreatedStart *string
	
	// Filtro por data de criação (fim)
	CreatedEnd *string
	
	// Ordem da listagem
	OrderBy string
	
	// Direção da ordenação (asc/desc)
	OrderDirection string
}