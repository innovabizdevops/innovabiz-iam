/**
 * @file list_identities.go
 * @description Query de aplicação para listagem de identidades com filtros
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package queries

import (
	"context"

	"innovabiz/iam/src/multi-context/domain/models"
	"innovabiz/iam/src/multi-context/domain/repositories"
	"innovabiz/iam/src/multi-context/domain/services"
)

// ListIdentitiesQuery representa a query para listar identidades
type ListIdentitiesQuery struct {
	Status          []string `json:"status,omitempty"`
	PrimaryKeyType  []string `json:"primary_key_type,omitempty"`
	SearchText      string   `json:"search_text,omitempty"`
	CreatedStart    string   `json:"created_start,omitempty"`
	CreatedEnd      string   `json:"created_end,omitempty"`
	OrderBy         string   `json:"order_by,omitempty"`
	OrderDirection  string   `json:"order_direction,omitempty"`
	Page            int      `json:"page,omitempty"`
	PageSize        int      `json:"page_size,omitempty"`
}

// IdentityListResult representa o resultado da listagem de identidades
type IdentityListResult struct {
	Items      []IdentityDTO `json:"items"`
	TotalCount int           `json:"total_count"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}

// ListIdentitiesHandler é o handler para a query de listagem de identidades
type ListIdentitiesHandler struct {
	identityService *services.IdentityService
}

// NewListIdentitiesHandler cria uma nova instância do handler
func NewListIdentitiesHandler(identityService *services.IdentityService) *ListIdentitiesHandler {
	return &ListIdentitiesHandler{
		identityService: identityService,
	}
}

// Handle executa a query de listagem de identidades
func (h *ListIdentitiesHandler) Handle(ctx context.Context, query ListIdentitiesQuery) (*IdentityListResult, error) {
	// Converter filtros de status para tipos de domínio
	statusFilter := make([]models.IdentityStatus, 0, len(query.Status))
	for _, s := range query.Status {
		statusFilter = append(statusFilter, models.IdentityStatus(s))
	}
	
	// Converter filtros de tipo de chave para tipos de domínio
	keyTypeFilter := make([]models.PrimaryKeyType, 0, len(query.PrimaryKeyType))
	for _, kt := range query.PrimaryKeyType {
		keyTypeFilter = append(keyTypeFilter, models.PrimaryKeyType(kt))
	}
	
	// Definir valores padrão para paginação
	page := query.Page
	if page <= 0 {
		page = 1
	}
	
	pageSize := query.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	
	// Criar filtro para o repositório
	filter := repositories.IdentityFilter{
		Status:         statusFilter,
		PrimaryKeyType: keyTypeFilter,
		SearchText:     query.SearchText,
		OrderBy:        query.OrderBy,
		OrderDirection: query.OrderDirection,
	}
	
	// Adicionar filtros de data se especificados
	if query.CreatedStart != "" {
		filter.CreatedStart = &query.CreatedStart
	}
	
	if query.CreatedEnd != "" {
		filter.CreatedEnd = &query.CreatedEnd
	}
	
	// Obter identidades do serviço de domínio
	identities, totalCount, err := h.identityService.ListIdentitiesByFilter(ctx, filter, page, pageSize)
	if err != nil {
		return nil, err
	}
	
	// Calcular total de páginas
	totalPages := (totalCount + pageSize - 1) / pageSize
	
	// Converter para DTOs
	items := make([]IdentityDTO, 0, len(identities))
	for _, identity := range identities {
		dto := mapIdentityToDTO(identity, false, false)
		items = append(items, *dto)
	}
	
	// Preparar resultado
	result := &IdentityListResult{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
	
	return result, nil
}