/**
 * INNOVABIZ IAM - Resolvers GraphQL para Consultas de Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação dos resolvers GraphQL para consultas relacionadas a grupos
 * no módulo IAM da plataforma INNOVABIZ, seguindo os princípios de
 * multi-tenant, multi-dimensional, multi-contextual e observabilidade total.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7)
 * - LGPD/GDPR/PDPA/CCPA (Proteção de dados e controle de acesso)
 * - BNA Instrução 7/2021 (Segurança cibernética para instituições financeiras)
 * - SOX (Rastreabilidade e auditoria)
 * - NIST CSF (Proteção de identidade)
 */

package resolvers

import (
	"context"
	"errors"
	"strconv"

	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
	"github.com/innovabiz/iam/internal/domain/group"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
)

// Group resolve uma consulta para buscar um grupo por ID
func (r *GroupResolver) Group(ctx context.Context, params graphql.ResolveParams) (interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.Group")
	defer span.End()

	// Registrar métrica de início da resolução
	timer := r.metrics.Timer("resolver.group.duration")
	defer timer.ObserveDuration()

	// Extrair parâmetros
	idStr, ok := params.Args["id"].(string)
	if !ok {
		return nil, errors.New("ID do grupo não fornecido ou inválido")
	}

	tenantIDStr, ok := params.Args["tenantId"].(string)
	if !ok {
		return nil, errors.New("ID do tenant não fornecido ou inválido")
	}

	// Converter para UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, errors.New("formato de ID de grupo inválido")
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, errors.New("formato de ID de tenant inválido")
	}

	// Buscar grupo
	group, err := r.groupService.GetGroupByID(ctx, id, tenantID)
	if err != nil {
		r.logger.Error(ctx, "Erro ao buscar grupo por ID", logging.Fields{
			"error":    err.Error(),
			"groupId":  id.String(),
			"tenantId": tenantID.String(),
		})

		// Incrementar contador de erro
		r.metrics.Counter("resolver.group.error").Inc(1)

		// Verificar tipo de erro para tratamento específico
		if errors.Is(err, group.ErrGroupNotFound) {
			return nil, nil // GraphQL retorna null para não encontrado
		}
		
		return nil, err
	}

	// Incrementar contador de sucesso
	r.metrics.Counter("resolver.group.success").Inc(1)

	return group, nil
}

// GroupByCode resolve uma consulta para buscar um grupo por código
func (r *GroupResolver) GroupByCode(ctx context.Context, params graphql.ResolveParams) (interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.GroupByCode")
	defer span.End()

	// Registrar métrica de início da resolução
	timer := r.metrics.Timer("resolver.groupByCode.duration")
	defer timer.ObserveDuration()

	// Extrair parâmetros
	code, ok := params.Args["code"].(string)
	if !ok {
		return nil, errors.New("código do grupo não fornecido ou inválido")
	}

	tenantIDStr, ok := params.Args["tenantId"].(string)
	if !ok {
		return nil, errors.New("ID do tenant não fornecido ou inválido")
	}

	// Converter para UUID
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, errors.New("formato de ID de tenant inválido")
	}

	// Buscar grupo
	group, err := r.groupService.GetGroupByCode(ctx, code, tenantID)
	if err != nil {
		r.logger.Error(ctx, "Erro ao buscar grupo por código", logging.Fields{
			"error":    err.Error(),
			"code":     code,
			"tenantId": tenantID.String(),
		})

		// Incrementar contador de erro
		r.metrics.Counter("resolver.groupByCode.error").Inc(1)

		// Verificar tipo de erro para tratamento específico
		if errors.Is(err, group.ErrGroupNotFound) {
			return nil, nil // GraphQL retorna null para não encontrado
		}
		
		return nil, err
	}

	// Incrementar contador de sucesso
	r.metrics.Counter("resolver.groupByCode.success").Inc(1)

	return group, nil
}

// Groups resolve uma consulta para listar grupos com filtros e paginação
func (r *GroupResolver) Groups(ctx context.Context, params graphql.ResolveParams) (interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.Groups")
	defer span.End()

	// Registrar métrica de início da resolução
	timer := r.metrics.Timer("resolver.groups.duration")
	defer timer.ObserveDuration()

	// Extrair parâmetros
	tenantIDStr, ok := params.Args["tenantId"].(string)
	if !ok {
		return nil, errors.New("ID do tenant não fornecido ou inválido")
	}

	// Converter para UUID
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, errors.New("formato de ID de tenant inválido")
	}

	// Preparar filtro padrão
	filter := group.GroupFilter{}

	// Extrair página e tamanho da página
	page := 1
	if pageArg, ok := params.Args["page"].(int); ok && pageArg > 0 {
		page = pageArg
	}
	filter.Page = page

	pageSize := 20
	if pageSizeArg, ok := params.Args["pageSize"].(int); ok && pageSizeArg > 0 {
		pageSize = pageSizeArg
	}
	filter.PageSize = pageSize

	// Extrair campos de ordenação
	if sortBy, ok := params.Args["sortBy"].(string); ok && sortBy != "" {
		filter.SortBy = sortBy
	} else {
		filter.SortBy = "name"
	}

	if sortDirection, ok := params.Args["sortDirection"].(string); ok && sortDirection != "" {
		filter.SortDirection = sortDirection
	} else {
		filter.SortDirection = "ASC"
	}

	// Processar filtros adicionais, se fornecidos
	if filterArg, ok := params.Args["filter"].(map[string]interface{}); ok {
		// Filtro por termo de busca
		if searchTerm, ok := filterArg["searchTerm"].(string); ok && searchTerm != "" {
			filter.SearchTerm = searchTerm
		}

		// Filtro por tipo de grupo
		if groupType, ok := filterArg["groupType"].(string); ok && groupType != "" {
			filter.GroupType = groupType
		}

		// Filtro por região
		if regionCode, ok := filterArg["regionCode"].(string); ok && regionCode != "" {
			filter.RegionCode = regionCode
		}

		// Filtro por grupo pai
		if parentGroupIDStr, ok := filterArg["parentGroupId"].(string); ok && parentGroupIDStr != "" {
			parentGroupID, err := uuid.Parse(parentGroupIDStr)
			if err == nil {
				filter.ParentGroupID = &parentGroupID
			}
		}

		// Filtro por status
		if statusArray, ok := filterArg["status"].([]interface{}); ok && len(statusArray) > 0 {
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

		// Filtro por nível hierárquico
		if level, ok := filterArg["level"].(int); ok {
			filter.Level = &level
		}

		// Filtro por grupo raiz (grupos de primeiro nível apenas)
		if rootOnly, ok := filterArg["rootOnly"].(bool); ok && rootOnly {
			filter.RootOnly = true
		}
	} else {
		// Por padrão, incluir apenas grupos ativos
		filter.Status = []group.Status{group.StatusActive}
	}

	// Buscar grupos
	result, err := r.groupService.ListGroups(ctx, tenantID, filter)
	if err != nil {
		r.logger.Error(ctx, "Erro ao listar grupos", logging.Fields{
			"error":    err.Error(),
			"tenantId": tenantID.String(),
			"filter":   filter,
		})

		// Incrementar contador de erro
		r.metrics.Counter("resolver.groups.error").Inc(1)
		
		return nil, err
	}

	// Incrementar contador de sucesso
	r.metrics.Counter("resolver.groups.success").Inc(1)

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

// GroupMembers resolve uma consulta para listar membros de um grupo
func (r *GroupResolver) GroupMembers(ctx context.Context, params graphql.ResolveParams) (interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.GroupMembers")
	defer span.End()

	// Registrar métrica de início da resolução
	timer := r.metrics.Timer("resolver.groupMembers.duration")
	defer timer.ObserveDuration()

	// Extrair parâmetros obrigatórios
	groupIDStr, ok := params.Args["groupId"].(string)
	if !ok {
		return nil, errors.New("ID do grupo não fornecido ou inválido")
	}

	tenantIDStr, ok := params.Args["tenantId"].(string)
	if !ok {
		return nil, errors.New("ID do tenant não fornecido ou inválido")
	}

	// Converter para UUID
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		return nil, errors.New("formato de ID de grupo inválido")
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, errors.New("formato de ID de tenant inválido")
	}

	// Extrair parâmetros opcionais
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

	// Construir filtros adicionais
	filter := map[string]interface{}{}

	if searchTerm, ok := params.Args["searchTerm"].(string); ok && searchTerm != "" {
		filter["searchTerm"] = searchTerm
	}

	if sortBy, ok := params.Args["sortBy"].(string); ok && sortBy != "" {
		filter["sortBy"] = sortBy
	} else {
		filter["sortBy"] = "displayName"
	}

	if sortDirection, ok := params.Args["sortDirection"].(string); ok && sortDirection != "" {
		filter["sortDirection"] = sortDirection
	} else {
		filter["sortDirection"] = "ASC"
	}

	// Processar filtro de status
	if statusArray, ok := params.Args["status"].([]interface{}); ok && len(statusArray) > 0 {
		statuses := make([]string, 0, len(statusArray))
		for _, s := range statusArray {
			if statusStr, ok := s.(string); ok {
				statuses = append(statuses, statusStr)
			}
		}
		if len(statuses) > 0 {
			filter["status"] = statuses
		}
	}

	// Buscar membros
	result, err := r.groupService.ListGroupMembers(ctx, groupID, tenantID, recursive, page, pageSize, filter)
	if err != nil {
		r.logger.Error(ctx, "Erro ao listar membros do grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"tenantId": tenantID.String(),
			"recursive": strconv.FormatBool(recursive),
		})

		// Incrementar contador de erro
		r.metrics.Counter("resolver.groupMembers.error").Inc(1)
		
		return nil, err
	}

	// Incrementar contador de sucesso
	r.metrics.Counter("resolver.groupMembers.success").Inc(1)

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