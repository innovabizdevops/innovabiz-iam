/**
 * INNOVABIZ IAM - Resolvers GraphQL para Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação dos resolvers GraphQL para o serviço de grupos do IAM,
 * seguindo os princípios de multi-tenant, multi-dimensional e compliance
 * com normas internacionais de segurança e governança.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7)
 * - LGPD/GDPR (Minimização de dados e controle de acesso)
 * - BNA Instrução 7/2021 (Segurança cibernética)
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
	"github.com/innovabiz/iam/internal/infrastructure/observability"
	"github.com/innovabiz/iam/internal/infrastructure/security"
)

// GroupResolver é responsável por resolver consultas e mutações relacionadas a grupos
type GroupResolver struct {
	groupService group.Service
	logger       logging.Logger
	tracer       observability.Tracer
	authz        security.Authorizer
}

// NewGroupResolver cria uma nova instância de GroupResolver
func NewGroupResolver(
	groupService group.Service,
	logger logging.Logger,
	tracer observability.Tracer,
	authz security.Authorizer,
) *GroupResolver {
	return &GroupResolver{
		groupService: groupService,
		logger:       logger,
		tracer:       tracer,
		authz:        authz,
	}
}

// Group resolve uma consulta para obter um grupo específico por ID
func (r *GroupResolver) Group(ctx context.Context, params graphql.ResolveParams) (interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.Group")
	defer span.End()

	// Extrair parâmetros
	idStr, ok := params.Args["id"].(string)
	if !ok {
		return nil, errors.New("id inválido ou não fornecido")
	}

	tenantIDStr, ok := params.Args["tenantId"].(string)
	if !ok {
		return nil, errors.New("tenantId inválido ou não fornecido")
	}

	// Converter strings para UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		r.logger.Warn(ctx, "ID de grupo inválido", logging.Fields{
			"id":    idStr,
			"error": err.Error(),
		})
		return nil, errors.New("formato de ID de grupo inválido")
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		r.logger.Warn(ctx, "ID de tenant inválido", logging.Fields{
			"tenantId": tenantIDStr,
			"error":    err.Error(),
		})
		return nil, errors.New("formato de ID de tenant inválido")
	}

	// Buscar grupo
	result, err := r.groupService.GetGroupByID(ctx, id, tenantID)
	if err != nil {
		if errors.Is(err, group.ErrGroupNotFound) {
			return nil, nil // Retornar null em vez de erro para caso não encontrado
		}
		r.logger.Error(ctx, "Erro ao buscar grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  idStr,
			"tenantId": tenantIDStr,
		})
		return nil, err
	}

	// Registrar métrica de acesso
	r.tracer.RecordMetric(ctx, "innovabiz:iam:graphql", "group_query_count", 1, map[string]string{
		"tenantId": tenantID.String(),
		"operation": "group",
	})

	return result, nil
}

// GroupByCode resolve uma consulta para obter um grupo específico por código
func (r *GroupResolver) GroupByCode(ctx context.Context, params graphql.ResolveParams) (interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.GroupByCode")
	defer span.End()

	// Extrair parâmetros
	code, ok := params.Args["code"].(string)
	if !ok {
		return nil, errors.New("código inválido ou não fornecido")
	}

	tenantIDStr, ok := params.Args["tenantId"].(string)
	if !ok {
		return nil, errors.New("tenantId inválido ou não fornecido")
	}

	// Converter string para UUID
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		r.logger.Warn(ctx, "ID de tenant inválido", logging.Fields{
			"tenantId": tenantIDStr,
			"error":    err.Error(),
		})
		return nil, errors.New("formato de ID de tenant inválido")
	}

	// Buscar grupo
	result, err := r.groupService.GetGroupByCode(ctx, code, tenantID)
	if err != nil {
		if errors.Is(err, group.ErrGroupNotFound) {
			return nil, nil // Retornar null em vez de erro para caso não encontrado
		}
		r.logger.Error(ctx, "Erro ao buscar grupo por código", logging.Fields{
			"error":    err.Error(),
			"code":     code,
			"tenantId": tenantIDStr,
		})
		return nil, err
	}

	return result, nil
}

// Groups resolve uma consulta para listar grupos com filtros e paginação
func (r *GroupResolver) Groups(ctx context.Context, params graphql.ResolveParams) (interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.Groups")
	defer span.End()

	// Extrair parâmetros
	tenantIDStr, ok := params.Args["tenantId"].(string)
	if !ok {
		return nil, errors.New("tenantId inválido ou não fornecido")
	}

	// Converter string para UUID
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		r.logger.Warn(ctx, "ID de tenant inválido", logging.Fields{
			"tenantId": tenantIDStr,
			"error":    err.Error(),
		})
		return nil, errors.New("formato de ID de tenant inválido")
	}

	// Extrair filtro
	filterInput, ok := params.Args["filter"].(map[string]interface{})
	if !ok {
		return nil, errors.New("filtro inválido ou não fornecido")
	}

	// Mapear filtro para struct GroupFilter
	filter := group.GroupFilter{
		Page:          1,
		PageSize:      20,
		SortBy:        "name",
		SortDirection: "ASC",
	}

	// Extrair valores do filtro
	if page, ok := filterInput["page"].(int); ok && page > 0 {
		filter.Page = page
	}

	if pageSize, ok := filterInput["pageSize"].(int); ok && pageSize > 0 {
		filter.PageSize = pageSize
	}

	if sortBy, ok := filterInput["sortBy"].(string); ok && sortBy != "" {
		filter.SortBy = sortBy
	}

	if sortDirection, ok := filterInput["sortDirection"].(string); ok && sortDirection != "" {
		filter.SortDirection = sortDirection
	}

	if searchTerm, ok := filterInput["searchTerm"].(string); ok {
		filter.SearchTerm = searchTerm
	}

	if groupType, ok := filterInput["groupType"].(string); ok {
		filter.GroupType = groupType
	}

	if maxLevel, ok := filterInput["maxLevel"].(int); ok {
		filter.MaxLevel = &maxLevel
	}

	if parentGroupIDStr, ok := filterInput["parentGroupId"].(string); ok {
		parentGroupID, err := uuid.Parse(parentGroupIDStr)
		if err == nil {
			filter.ParentGroupID = &parentGroupID
		}
	}

	if memberUserIDStr, ok := filterInput["memberUserId"].(string); ok {
		memberUserID, err := uuid.Parse(memberUserIDStr)
		if err == nil {
			filter.MemberUserID = &memberUserID
		}
	}

	// Extrair status
	if statusArray, ok := filterInput["status"].([]interface{}); ok {
		statuses := make([]group.Status, 0, len(statusArray))
		for _, s := range statusArray {
			if statusStr, ok := s.(string); ok {
				statuses = append(statuses, group.Status(statusStr))
			}
		}
		filter.Status = statuses
	}

	// Buscar grupos
	result, err := r.groupService.ListGroups(ctx, tenantID, filter)
	if err != nil {
		r.logger.Error(ctx, "Erro ao listar grupos", logging.Fields{
			"error":    err.Error(),
			"tenantId": tenantIDStr,
			"filter":   filterInput,
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

// GroupMembers resolve uma consulta para listar membros de um grupo
func (r *GroupResolver) GroupMembers(ctx context.Context, params graphql.ResolveParams) (interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.GroupMembers")
	defer span.End()

	// Extrair parâmetros
	groupIDStr, ok := params.Args["groupId"].(string)
	if !ok {
		return nil, errors.New("groupId inválido ou não fornecido")
	}

	tenantIDStr, ok := params.Args["tenantId"].(string)
	if !ok {
		return nil, errors.New("tenantId inválido ou não fornecido")
	}

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

	// Converter strings para UUID
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		r.logger.Warn(ctx, "ID de grupo inválido", logging.Fields{
			"groupId": groupIDStr,
			"error":   err.Error(),
		})
		return nil, errors.New("formato de ID de grupo inválido")
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		r.logger.Warn(ctx, "ID de tenant inválido", logging.Fields{
			"tenantId": tenantIDStr,
			"error":    err.Error(),
		})
		return nil, errors.New("formato de ID de tenant inválido")
	}

	// Buscar membros
	result, err := r.groupService.ListGroupMembers(ctx, groupID, tenantID, recursive, page, pageSize)
	if err != nil {
		r.logger.Error(ctx, "Erro ao listar membros do grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupIDStr,
			"tenantId": tenantIDStr,
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