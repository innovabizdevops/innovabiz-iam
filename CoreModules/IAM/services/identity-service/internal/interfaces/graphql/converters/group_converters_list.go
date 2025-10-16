/**
 * INNOVABIZ IAM - Conversores de Modelos para Grupos (Parte 2)
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação dos conversores entre modelos de domínio e modelos GraphQL
 * para listas e coleções de grupos, seguindo a arquitetura multi-dimensional 
 * e multi-tenant da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (A.5.15 - Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7.2.4 - Gestão de grupos)
 * - LGPD/GDPR/PDPA (Controle de acesso)
 */

package converters

import (
	"github.com/google/uuid"
	
	"github.com/innovabiz/iam/internal/domain/entities"
	"github.com/innovabiz/iam/internal/interfaces/graphql/model"
)

// DomainGroupsToGraphQL converte uma lista de grupos de domínio para modelos GraphQL
func DomainGroupsToGraphQL(domainGroups []*entities.Group) []*model.Group {
	if domainGroups == nil {
		return nil
	}

	graphqlGroups := make([]*model.Group, 0, len(domainGroups))
	for _, domainGroup := range domainGroups {
		graphqlGroups = append(graphqlGroups, DomainGroupToGraphQL(domainGroup))
	}

	return graphqlGroups
}

// DomainPageInfoToGraphQL converte informações de paginação do domínio para GraphQL
func DomainPageInfoToGraphQL(pageInfo *entities.PageInfo) *model.PageInfo {
	if pageInfo == nil {
		return nil
	}
	
	return &model.PageInfo{
		CurrentPage:      pageInfo.CurrentPage,
		PageSize:         pageInfo.PageSize,
		TotalPages:       pageInfo.TotalPages,
		HasNextPage:      pageInfo.HasNextPage,
		HasPreviousPage:  pageInfo.HasPreviousPage,
	}
}

// DomainGroupListResultToGraphQL converte um resultado de lista de grupos de domínio para GraphQL
func DomainGroupListResultToGraphQL(result *entities.GroupListResult) *model.GroupListResult {
	if result == nil {
		return nil
	}
	
	return &model.GroupListResult{
		Items:      DomainGroupsToGraphQL(result.Items),
		TotalCount: result.TotalCount,
		PageInfo:   DomainPageInfoToGraphQL(result.PageInfo),
	}
}

// DomainUserListResultToGraphQL converte um resultado de lista de usuários de domínio para GraphQL
func DomainUserListResultToGraphQL(result *entities.UserListResult) *model.UserListResult {
	if result == nil {
		return nil
	}
	
	users := make([]*model.User, 0, len(result.Items))
	for _, user := range result.Items {
		users = append(users, &model.User{
			ID:       user.ID.String(),
			Username: user.Username,
			Email:    user.Email,
		})
	}
	
	return &model.UserListResult{
		Items:      users,
		TotalCount: result.TotalCount,
		PageInfo:   DomainPageInfoToGraphQL(result.PageInfo),
	}
}

// GraphQLUserFilterToDomain converte um filtro de usuários GraphQL para filtro de domínio
func GraphQLUserFilterToDomain(filter *model.UserFilter) (*entities.UserFilter, error) {
	if filter == nil {
		return nil, nil
	}

	domainFilter := &entities.UserFilter{}

	// Converter IDs
	if len(filter.IDs) > 0 {
		domainFilter.IDs = make([]uuid.UUID, 0, len(filter.IDs))
		for _, idStr := range filter.IDs {
			id, err := uuid.Parse(idStr)
			if err != nil {
				return nil, err
			}
			domainFilter.IDs = append(domainFilter.IDs, id)
		}
	}

	// Converter filtros de texto
	if filter.UsernameContains != nil {
		domainFilter.UsernameContains = *filter.UsernameContains
	}

	if filter.EmailContains != nil {
		domainFilter.EmailContains = *filter.EmailContains
	}

	// Converter status e tipos
	if len(filter.Statuses) > 0 {
		domainFilter.Statuses = filter.Statuses
	}

	if len(filter.Types) > 0 {
		domainFilter.Types = filter.Types
	}

	// Converter datas
	if filter.CreatedAtStart != nil {
		domainFilter.CreatedAtStart = filter.CreatedAtStart.Time
	}

	if filter.CreatedAtEnd != nil {
		domainFilter.CreatedAtEnd = filter.CreatedAtEnd.Time
	}

	if filter.UpdatedAtStart != nil {
		domainFilter.UpdatedAtStart = filter.UpdatedAtStart.Time
	}

	if filter.UpdatedAtEnd != nil {
		domainFilter.UpdatedAtEnd = filter.UpdatedAtEnd.Time
	}

	return domainFilter, nil
}