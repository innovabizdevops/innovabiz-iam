/**
 * INNOVABIZ IAM - Resolvers de Campos Relacionados a Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação dos resolvers para campos complexos e relacionamentos do tipo Group,
 * seguindo a arquitetura multi-dimensional, multi-tenant, multi-camada e multi-contexto
 * da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7)
 * - LGPD/GDPR/PDPA (Proteção de dados)
 * - BNA Instrução 7/2021 (Segurança cibernética)
 * - SOX (Rastreabilidade e auditoria)
 * - NIST CSF (Proteção de identidade)
 */

package resolvers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
	"github.com/innovabiz/iam/internal/domain/group"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
)

// RegisterGroupFieldResolvers registra os resolvers de campos relacionados a grupos
func RegisterGroupFieldResolvers(schema *graphql.Schema, resolver *GroupResolver) {
	groupType := schema.Type("Group").(*graphql.Object)

	// Resolver para o campo "parentGroup"
	groupType.AddFieldConfig("parentGroup", &graphql.Field{
		Type: graphql.NewNonNull(graphql.NewObject(graphql.ObjectConfig{
			Name:        "Group",
			Description: "Grupo pai",
		})),
		Resolve: resolver.ParentGroup,
	})

	// Resolver para o campo "users" (membros)
	groupType.AddFieldConfig("users", &graphql.Field{
		Type: graphql.NewNonNull(graphql.NewObject(graphql.ObjectConfig{
			Name:        "UserConnection",
			Description: "Conexão paginada de usuários",
		})),
		Args: graphql.FieldConfigArgument{
			"searchTerm": &graphql.ArgumentConfig{
				Type:        graphql.String,
				Description: "Termo de busca para filtragem",
			},
			"status": &graphql.ArgumentConfig{
				Type:        graphql.NewList(graphql.NewNonNull(graphql.NewEnum(graphql.EnumConfig{
					Name:        "UserStatus",
					Description: "Status possíveis para um usuário",
					Values: graphql.EnumValueConfigMap{
						"ACTIVE":    &graphql.EnumValueConfig{Value: "ACTIVE"},
						"INACTIVE":  &graphql.EnumValueConfig{Value: "INACTIVE"},
						"PENDING":   &graphql.EnumValueConfig{Value: "PENDING"},
						"LOCKED":    &graphql.EnumValueConfig{Value: "LOCKED"},
						"DELETED":   &graphql.EnumValueConfig{Value: "DELETED"},
					},
				}))),
				Description: "Status dos usuários a incluir",
			},
			"sortBy": &graphql.ArgumentConfig{
				Type:         graphql.String,
				Description:  "Campo para ordenação",
				DefaultValue: "displayName",
			},
			"sortDirection": &graphql.ArgumentConfig{
				Type:         graphql.String,
				Description:  "Direção da ordenação (ASC/DESC)",
				DefaultValue: "ASC",
			},
			"page": &graphql.ArgumentConfig{
				Type:         graphql.Int,
				Description:  "Número da página (inicia em 1)",
				DefaultValue: 1,
			},
			"pageSize": &graphql.ArgumentConfig{
				Type:         graphql.Int,
				Description:  "Tamanho da página",
				DefaultValue: 20,
			},
			"recursive": &graphql.ArgumentConfig{
				Type:         graphql.Boolean,
				Description:  "Incluir usuários de subgrupos",
				DefaultValue: false,
			},
		},
		Resolve: resolver.GroupUsers,
	})

	// Resolver para o campo "childGroups" (subgrupos diretos)
	groupType.AddFieldConfig("childGroups", &graphql.Field{
		Type: graphql.NewNonNull(graphql.NewObject(graphql.ObjectConfig{
			Name:        "GroupConnection",
			Description: "Conexão paginada de grupos",
		})),
		Args: graphql.FieldConfigArgument{
			"status": &graphql.ArgumentConfig{
				Type:        graphql.NewList(graphql.NewNonNull(graphql.NewEnum(graphql.EnumConfig{
					Name:        "GroupStatus",
					Description: "Status possíveis para um grupo",
					Values: graphql.EnumValueConfigMap{
						"ACTIVE":   &graphql.EnumValueConfig{Value: "ACTIVE"},
						"INACTIVE": &graphql.EnumValueConfig{Value: "INACTIVE"},
						"PENDING":  &graphql.EnumValueConfig{Value: "PENDING"},
						"DELETED":  &graphql.EnumValueConfig{Value: "DELETED"},
					},
				}))),
				Description: "Status dos grupos a incluir",
			},
			"searchTerm": &graphql.ArgumentConfig{
				Type:        graphql.String,
				Description: "Termo de busca para filtragem",
			},
			"page": &graphql.ArgumentConfig{
				Type:         graphql.Int,
				Description:  "Número da página (inicia em 1)",
				DefaultValue: 1,
			},
			"pageSize": &graphql.ArgumentConfig{
				Type:         graphql.Int,
				Description:  "Tamanho da página",
				DefaultValue: 20,
			},
		},
		Resolve: resolver.ChildGroups,
	})

	// Resolver para o campo "parentGroups" (grupos ancestrais)
	groupType.AddFieldConfig("parentGroups", &graphql.Field{
		Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.NewObject(graphql.ObjectConfig{
			Name:        "Group",
			Description: "Grupo ancestral",
		})))),
		Resolve: resolver.ParentGroups,
	})

	// Resolver para o campo "roles" (funções/papéis associados ao grupo)
	groupType.AddFieldConfig("roles", &graphql.Field{
		Type: graphql.NewNonNull(graphql.NewObject(graphql.ObjectConfig{
			Name:        "RoleConnection",
			Description: "Conexão paginada de funções/papéis",
		})),
		Args: graphql.FieldConfigArgument{
			"limit": &graphql.ArgumentConfig{
				Type:         graphql.Int,
				Description:  "Limite de resultados",
				DefaultValue: 10,
			},
			"offset": &graphql.ArgumentConfig{
				Type:         graphql.Int,
				Description:  "Deslocamento para paginação",
				DefaultValue: 0,
			},
		},
		Resolve: resolver.GroupRoles,
	})

	// Resolver para o campo "stats" (estatísticas detalhadas do grupo)
	groupType.AddFieldConfig("stats", &graphql.Field{
		Type: graphql.NewObject(graphql.ObjectConfig{
			Name:        "GroupStats",
			Description: "Estatísticas detalhadas de um grupo",
		}),
		Resolve: resolver.GroupStats,
	})

	// Resolver para os campos de contagem
	groupType.AddFieldConfig("membersCount", &graphql.Field{
		Type:        graphql.NewNonNull(graphql.Int),
		Description: "Contagem total de membros (diretos)",
		Resolve:     resolver.MembersCount,
	})

	groupType.AddFieldConfig("subgroupsCount", &graphql.Field{
		Type:        graphql.NewNonNull(graphql.Int),
		Description: "Contagem total de subgrupos (diretos)",
		Resolve:     resolver.SubgroupsCount,
	})
}

// ParentGroup resolve o campo "parentGroup"
func (r *GroupResolver) ParentGroup(ctx context.Context, params graphql.ResolveParams) (interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.ParentGroup")
	defer span.End()

	group, ok := params.Source.(*group.Group)
	if !ok || group == nil || group.ParentGroupID == nil {
		return nil, nil
	}

	parentGroup, err := r.groupService.GetGroupByID(ctx, *group.ParentGroupID, group.TenantID)
	if err != nil {
		if errors.Is(err, group.ErrGroupNotFound) {
			return nil, nil
		}
		r.logger.Error(ctx, "Erro ao buscar grupo pai", logging.Fields{
			"error":        err.Error(),
			"groupId":      group.ID.String(),
			"parentGroupId": group.ParentGroupID.String(),
		})
		return nil, err
	}

	return parentGroup, nil
}

// GroupUsers resolve o campo "users"
func (r *GroupResolver) GroupUsers(ctx context.Context, params graphql.ResolveParams) (interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.GroupUsers")
	defer span.End()

	group, ok := params.Source.(*group.Group)
	if !ok || group == nil {
		return nil, errors.New("grupo inválido")
	}

	// Extrair argumentos
	recursive := false
	if recursiveArg, ok := params.Args["recursive"].(bool); ok {
		recursive = recursiveArg
	}

	page := 1
	if pageArg, ok := params.Args["page"].(int); ok && pageArg > 0 {
		page = pageArg
	}

	pageSize := 20
	if pageSizeArg, ok := params.Args["pageSize"].(int); ok && pageSizeArg > 0 {
		pageSize = pageSizeArg
	}

	// Mapear filtros adicionais
	filter := map[string]interface{}{}
	
	if searchTerm, ok := params.Args["searchTerm"].(string); ok && searchTerm != "" {
		filter["searchTerm"] = searchTerm
	}
	
	if sortBy, ok := params.Args["sortBy"].(string); ok && sortBy != "" {
		filter["sortBy"] = sortBy
	}
	
	if sortDirection, ok := params.Args["sortDirection"].(string); ok && sortDirection != "" {
		filter["sortDirection"] = sortDirection
	}
	
	if statusArray, ok := params.Args["status"].([]interface{}); ok && len(statusArray) > 0 {
		statuses := make([]string, 0, len(statusArray))
		for _, s := range statusArray {
			if statusStr, ok := s.(string); ok {
				statuses = append(statuses, statusStr)
			}
		}
		filter["status"] = statuses
	}

	// Buscar membros
	result, err := r.groupService.ListGroupMembers(ctx, group.ID, group.TenantID, recursive, page, pageSize, filter)
	if err != nil {
		r.logger.Error(ctx, "Erro ao listar membros do grupo", logging.Fields{
			"error":   err.Error(),
			"groupId": group.ID.String(),
		})
		return nil, err
	}

	// Mapear resultado para o formato esperado pelo GraphQL
	return map[string]interface{}{
		"users":      result.Users,
		"totalCount": result.TotalCount,
		"pageInfo": map[string]interface{}{
			"currentPage":     result.Page,
			"pageSize":        result.PageSize,
			"totalPages":      result.TotalPages,
			"hasNextPage":     result.Page < result.TotalPages,
			"hasPreviousPage": result.Page > 1,
		},
	}, nil
}

// ChildGroups resolve o campo "childGroups"
func (r *GroupResolver) ChildGroups(ctx context.Context, params graphql.ResolveParams) (interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.ChildGroups")
	defer span.End()

	group, ok := params.Source.(*group.Group)
	if !ok || group == nil {
		return nil, errors.New("grupo inválido")
	}

	// Extrair argumentos
	page := 1
	if pageArg, ok := params.Args["page"].(int); ok && pageArg > 0 {
		page = pageArg
	}

	pageSize := 20
	if pageSizeArg, ok := params.Args["pageSize"].(int); ok && pageSizeArg > 0 {
		pageSize = pageSizeArg
	}

	// Criar filtro para buscar subgrupos
	filter := group.GroupFilter{
		ParentGroupID: &group.ID,
		Page:          page,
		PageSize:      pageSize,
		SortBy:        "name",
		SortDirection: "ASC",
	}

	// Adicionar filtros opcionais
	if searchTerm, ok := params.Args["searchTerm"].(string); ok && searchTerm != "" {
		filter.SearchTerm = searchTerm
	}

	// Filtrar por status
	if statusArray, ok := params.Args["status"].([]interface{}); ok && len(statusArray) > 0 {
		statuses := make([]group.Status, 0, len(statusArray))
		for _, s := range statusArray {
			if statusStr, ok := s.(string); ok {
				statuses = append(statuses, group.Status(statusStr))
			}
		}
		filter.Status = statuses
	} else {
		// Por padrão, incluir apenas grupos ativos
		filter.Status = []group.Status{group.StatusActive}
	}

	// Buscar subgrupos
	result, err := r.groupService.ListGroups(ctx, group.TenantID, filter)
	if err != nil {
		r.logger.Error(ctx, "Erro ao listar subgrupos", logging.Fields{
			"error":   err.Error(),
			"groupId": group.ID.String(),
		})
		return nil, err
	}

	// Mapear resultado para o formato esperado pelo GraphQL
	return map[string]interface{}{
		"groups":     result.Groups,
		"totalCount": result.TotalCount,
		"pageInfo": map[string]interface{}{
			"currentPage":     result.Page,
			"pageSize":        result.PageSize,
			"totalPages":      result.TotalPages,
			"hasNextPage":     result.Page < result.TotalPages,
			"hasPreviousPage": result.Page > 1,
		},
	}, nil
}

// ParentGroups resolve o campo "parentGroups"
func (r *GroupResolver) ParentGroups(ctx context.Context, params graphql.ResolveParams) (interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.ParentGroups")
	defer span.End()

	group, ok := params.Source.(*group.Group)
	if !ok || group == nil {
		return nil, errors.New("grupo inválido")
	}

	// Buscar hierarquia de grupos
	ancestors, err := r.groupService.GetGroupHierarchy(ctx, group.ID, group.TenantID)
	if err != nil {
		r.logger.Error(ctx, "Erro ao buscar hierarquia de grupos", logging.Fields{
			"error":   err.Error(),
			"groupId": group.ID.String(),
		})
		return nil, err
	}

	return ancestors, nil
}

// GroupRoles resolve o campo "roles"
func (r *GroupResolver) GroupRoles(ctx context.Context, params graphql.ResolveParams) (interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.GroupRoles")
	defer span.End()

	group, ok := params.Source.(*group.Group)
	if !ok || group == nil {
		return nil, errors.New("grupo inválido")
	}

	// Extrair argumentos
	limit := 10
	if limitArg, ok := params.Args["limit"].(int); ok && limitArg > 0 {
		limit = limitArg
	}

	offset := 0
	if offsetArg, ok := params.Args["offset"].(int); ok && offsetArg >= 0 {
		offset = offsetArg
	}

	// Buscar funções/papéis associados ao grupo
	roles, total, err := r.roleService.GetRolesByGroupID(ctx, group.ID, group.TenantID, limit, offset)
	if err != nil {
		r.logger.Error(ctx, "Erro ao buscar funções do grupo", logging.Fields{
			"error":   err.Error(),
			"groupId": group.ID.String(),
		})
		return nil, err
	}

	// Calcular número total de páginas
	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	// Mapear resultado para o formato esperado pelo GraphQL
	return map[string]interface{}{
		"roles":      roles,
		"totalCount": total,
		"pageInfo": map[string]interface{}{
			"hasNextPage":     offset+limit < total,
			"hasPreviousPage": offset > 0,
		},
	}, nil
}

// GroupStats resolve o campo "stats"
func (r *GroupResolver) GroupStats(ctx context.Context, params graphql.ResolveParams) (interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.GroupStats")
	defer span.End()

	group, ok := params.Source.(*group.Group)
	if !ok || group == nil {
		return nil, errors.New("grupo inválido")
	}

	// Buscar estatísticas do grupo
	stats, err := r.groupService.GetGroupStats(ctx, group.ID, group.TenantID)
	if err != nil {
		r.logger.Error(ctx, "Erro ao buscar estatísticas do grupo", logging.Fields{
			"error":   err.Error(),
			"groupId": group.ID.String(),
		})
		return nil, err
	}

	return stats, nil
}

// MembersCount resolve o campo "membersCount"
func (r *GroupResolver) MembersCount(ctx context.Context, params graphql.ResolveParams) (interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.MembersCount")
	defer span.End()

	group, ok := params.Source.(*group.Group)
	if !ok || group == nil {
		return 0, nil
	}

	// Buscar contagem de membros
	count, err := r.groupService.CountGroupMembers(ctx, group.ID, group.TenantID, false)
	if err != nil {
		r.logger.Error(ctx, "Erro ao contar membros do grupo", logging.Fields{
			"error":   err.Error(),
			"groupId": group.ID.String(),
		})
		return 0, err
	}

	return count, nil
}

// SubgroupsCount resolve o campo "subgroupsCount"
func (r *GroupResolver) SubgroupsCount(ctx context.Context, params graphql.ResolveParams) (interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.SubgroupsCount")
	defer span.End()

	group, ok := params.Source.(*group.Group)
	if !ok || group == nil {
		return 0, nil
	}

	// Buscar contagem de subgrupos
	count, err := r.groupService.CountSubgroups(ctx, group.ID, group.TenantID, false)
	if err != nil {
		r.logger.Error(ctx, "Erro ao contar subgrupos", logging.Fields{
			"error":   err.Error(),
			"groupId": group.ID.String(),
		})
		return 0, err
	}

	return count, nil
}