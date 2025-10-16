/**
 * INNOVABIZ IAM - Resolver GraphQL para Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação do resolver GraphQL para integração do serviço de domínio 
 * de grupos com a API GraphQL no módulo Core IAM, seguindo a arquitetura 
 * multi-dimensional, multi-tenant e com observabilidade total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (A.5.15 - Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7.2.4 - Gestão de grupos)
 * - LGPD/GDPR/PDPA (Controle de acesso)
 * - BNA Instrução 7/2021 (Segurança cibernética)
 * - SOX (Rastreabilidade e auditoria)
 * - NIST CSF (Proteção de identidade)
 */

package resolvers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	
	"github.com/innovabiz/iam/internal/domain/group"
	"github.com/innovabiz/iam/internal/domain/services"
	"github.com/innovabiz/iam/internal/domain/user"
	"github.com/innovabiz/iam/internal/domain/validation"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
	"github.com/innovabiz/iam/internal/infrastructure/metrics"
	"github.com/innovabiz/iam/internal/infrastructure/tracing"
	"github.com/innovabiz/iam/internal/interfaces/graphql/model"
	"github.com/innovabiz/iam/internal/interfaces/graphql/scalars"
)

// GroupResolver implementa os resolvers GraphQL para operações de grupos
type GroupResolver struct {
	groupService services.GroupService
	logger       logging.Logger
	metrics      metrics.Client
	tracer       tracing.Tracer
}

// NewGroupResolver cria uma nova instância do resolver de grupos
func NewGroupResolver(
	groupService services.GroupService,
	logger logging.Logger,
	metrics metrics.Client,
	tracer tracing.Tracer,
) *GroupResolver {
	return &GroupResolver{
		groupService: groupService,
		logger:       logger,
		metrics:      metrics,
		tracer:       tracer,
	}
}

// Resolvers para Query

// Group resolve a query para buscar um grupo por ID
func (r *GroupResolver) Group(ctx context.Context, id string, tenantID string) (*model.Group, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.Group")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", id),
		attribute.String("tenant.id", tenantID),
	)

	timer := r.metrics.Timer("resolver.group.getById.duration")
	defer timer.ObserveDuration()

	r.logger.Debug(ctx, "Resolvendo query para buscar grupo por ID", logging.Fields{
		"groupId":  id,
		"tenantId": tenantID,
	})

	// Converter IDs de string para UUID
	groupID, err := uuid.Parse(id)
	if err != nil {
		r.logger.Error(ctx, "ID do grupo inválido", logging.Fields{
			"error":   err.Error(),
			"groupId": id,
		})
		r.metrics.Counter("resolver.group.getById.invalidId").Inc(1)
		return nil, fmt.Errorf("ID do grupo inválido: %w", err)
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		r.logger.Error(ctx, "ID do tenant inválido", logging.Fields{
			"error":    err.Error(),
			"tenantId": tenantID,
		})
		r.metrics.Counter("resolver.group.getById.invalidTenantId").Inc(1)
		return nil, fmt.Errorf("ID do tenant inválido: %w", err)
	}

	// Chamar o serviço de domínio
	domainGroup, err := r.groupService.GetByID(ctx, groupID, tenantUUID)
	if err != nil {
		if errors.Is(err, group.ErrGroupNotFound) {
			r.logger.Debug(ctx, "Grupo não encontrado", logging.Fields{
				"groupId":  id,
				"tenantId": tenantID,
			})
			r.metrics.Counter("resolver.group.getById.notFound").Inc(1)
			return nil, nil // Retornar nil sem erro para representar "não encontrado" no GraphQL
		}

		r.logger.Error(ctx, "Erro ao buscar grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  id,
			"tenantId": tenantID,
		})
		r.metrics.Counter("resolver.group.getById.error").Inc(1)
		return nil, fmt.Errorf("erro ao buscar grupo: %w", err)
	}

	// Converter o modelo de domínio para o modelo GraphQL
	result, err := domainGroupToGraphQL(domainGroup)
	if err != nil {
		r.logger.Error(ctx, "Erro ao converter grupo para modelo GraphQL", logging.Fields{
			"error":    err.Error(),
			"groupId":  id,
			"tenantId": tenantID,
		})
		r.metrics.Counter("resolver.group.getById.conversionError").Inc(1)
		return nil, fmt.Errorf("erro ao converter grupo para resposta: %w", err)
	}

	r.metrics.Counter("resolver.group.getById.success").Inc(1)
	return result, nil
}

// GroupByCode resolve a query para buscar um grupo por código
func (r *GroupResolver) GroupByCode(ctx context.Context, code string, tenantID string) (*model.Group, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.GroupByCode")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.code", code),
		attribute.String("tenant.id", tenantID),
	)

	timer := r.metrics.Timer("resolver.group.getByCode.duration")
	defer timer.ObserveDuration()

	r.logger.Debug(ctx, "Resolvendo query para buscar grupo por código", logging.Fields{
		"code":     code,
		"tenantId": tenantID,
	})

	// Validar e converter o ID do tenant
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		r.logger.Error(ctx, "ID do tenant inválido", logging.Fields{
			"error":    err.Error(),
			"tenantId": tenantID,
		})
		r.metrics.Counter("resolver.group.getByCode.invalidTenantId").Inc(1)
		return nil, fmt.Errorf("ID do tenant inválido: %w", err)
	}

	// Chamar o serviço de domínio
	domainGroup, err := r.groupService.GetByCode(ctx, code, tenantUUID)
	if err != nil {
		if errors.Is(err, group.ErrGroupNotFound) {
			r.logger.Debug(ctx, "Grupo não encontrado", logging.Fields{
				"code":     code,
				"tenantId": tenantID,
			})
			r.metrics.Counter("resolver.group.getByCode.notFound").Inc(1)
			return nil, nil // Retornar nil sem erro para representar "não encontrado" no GraphQL
		}

		r.logger.Error(ctx, "Erro ao buscar grupo por código", logging.Fields{
			"error":    err.Error(),
			"code":     code,
			"tenantId": tenantID,
		})
		r.metrics.Counter("resolver.group.getByCode.error").Inc(1)
		return nil, fmt.Errorf("erro ao buscar grupo por código: %w", err)
	}

	// Converter o modelo de domínio para o modelo GraphQL
	result, err := domainGroupToGraphQL(domainGroup)
	if err != nil {
		r.logger.Error(ctx, "Erro ao converter grupo para modelo GraphQL", logging.Fields{
			"error":    err.Error(),
			"code":     code,
			"tenantId": tenantID,
		})
		r.metrics.Counter("resolver.group.getByCode.conversionError").Inc(1)
		return nil, fmt.Errorf("erro ao converter grupo para resposta: %w", err)
	}

	r.metrics.Counter("resolver.group.getByCode.success").Inc(1)
	return result, nil
}

// Groups resolve a query para listar grupos com filtros e paginação
func (r *GroupResolver) Groups(
	ctx context.Context,
	tenantID string,
	filter *model.GroupFilter,
	page *int,
	pageSize *int,
	sortBy *string,
	sortDirection *model.SortDirection,
) (*model.GroupListResult, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.Groups")
	defer span.End()

	span.SetAttributes(
		attribute.String("tenant.id", tenantID),
		attribute.Int("page", *page),
		attribute.Int("pageSize", *pageSize),
	)

	timer := r.metrics.Timer("resolver.group.list.duration")
	defer timer.ObserveDuration()

	r.logger.Debug(ctx, "Resolvendo query para listar grupos", logging.Fields{
		"tenantId": tenantID,
		"page":     *page,
		"pageSize": *pageSize,
	})

	// Validar e converter o ID do tenant
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		r.logger.Error(ctx, "ID do tenant inválido", logging.Fields{
			"error":    err.Error(),
			"tenantId": tenantID,
		})
		r.metrics.Counter("resolver.group.list.invalidTenantId").Inc(1)
		return nil, fmt.Errorf("ID do tenant inválido: %w", err)
	}

	// Converter filtro do GraphQL para filtro de domínio
	domainFilter, err := graphQLFilterToDomain(filter)
	if err != nil {
		r.logger.Error(ctx, "Erro ao converter filtro para modelo de domínio", logging.Fields{
			"error": err.Error(),
		})
		r.metrics.Counter("resolver.group.list.filterConversionError").Inc(1)
		return nil, fmt.Errorf("erro no filtro: %w", err)
	}

	// Chamar o serviço de domínio
	domainResult, err := r.groupService.List(ctx, tenantUUID, *domainFilter, *page, *pageSize)
	if err != nil {
		r.logger.Error(ctx, "Erro ao listar grupos", logging.Fields{
			"error":    err.Error(),
			"tenantId": tenantID,
		})
		r.metrics.Counter("resolver.group.list.error").Inc(1)
		return nil, fmt.Errorf("erro ao listar grupos: %w", err)
	}

	// Converter resultado para o modelo GraphQL
	result, err := domainListResultToGraphQL(domainResult, *page, *pageSize)
	if err != nil {
		r.logger.Error(ctx, "Erro ao converter resultado para modelo GraphQL", logging.Fields{
			"error":    err.Error(),
			"tenantId": tenantID,
		})
		r.metrics.Counter("resolver.group.list.conversionError").Inc(1)
		return nil, fmt.Errorf("erro ao converter resultado: %w", err)
	}

	r.metrics.Counter("resolver.group.list.success").Inc(1)
	return result, nil
}

// GroupsByUser resolve a query para buscar grupos de um usuário
func (r *GroupResolver) GroupsByUser(
	ctx context.Context,
	userID string,
	tenantID string,
	recursive *bool,
	page *int,
	pageSize *int,
) (*model.GroupListResult, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.GroupsByUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID),
		attribute.String("tenant.id", tenantID),
		attribute.Bool("recursive", *recursive),
	)

	timer := r.metrics.Timer("resolver.group.groupsByUser.duration")
	defer timer.ObserveDuration()

	r.logger.Debug(ctx, "Resolvendo query para buscar grupos de um usuário", logging.Fields{
		"userId":    userID,
		"tenantId":  tenantID,
		"recursive": *recursive,
		"page":      *page,
		"pageSize":  *pageSize,
	})

	// Converter IDs de string para UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		r.logger.Error(ctx, "ID do usuário inválido", logging.Fields{
			"error":  err.Error(),
			"userId": userID,
		})
		r.metrics.Counter("resolver.group.groupsByUser.invalidUserId").Inc(1)
		return nil, fmt.Errorf("ID do usuário inválido: %w", err)
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		r.logger.Error(ctx, "ID do tenant inválido", logging.Fields{
			"error":    err.Error(),
			"tenantId": tenantID,
		})
		r.metrics.Counter("resolver.group.groupsByUser.invalidTenantId").Inc(1)
		return nil, fmt.Errorf("ID do tenant inválido: %w", err)
	}

	// Chamar o serviço de domínio
	domainResult, err := r.groupService.FindGroupsByUserID(ctx, userUUID, tenantUUID, *recursive, *page, *pageSize)
	if err != nil {
		r.logger.Error(ctx, "Erro ao buscar grupos do usuário", logging.Fields{
			"error":    err.Error(),
			"userId":   userID,
			"tenantId": tenantID,
		})
		r.metrics.Counter("resolver.group.groupsByUser.error").Inc(1)
		return nil, fmt.Errorf("erro ao buscar grupos do usuário: %w", err)
	}

	// Converter resultado para o modelo GraphQL
	result, err := domainListResultToGraphQL(domainResult, *page, *pageSize)
	if err != nil {
		r.logger.Error(ctx, "Erro ao converter resultado para modelo GraphQL", logging.Fields{
			"error":    err.Error(),
			"userId":   userID,
			"tenantId": tenantID,
		})
		r.metrics.Counter("resolver.group.groupsByUser.conversionError").Inc(1)
		return nil, fmt.Errorf("erro ao converter resultado: %w", err)
	}

	r.metrics.Counter("resolver.group.groupsByUser.success").Inc(1)
	return result, nil
}

// Funções de conversão entre modelos

// domainGroupToGraphQL converte um grupo do modelo de domínio para o modelo GraphQL
func domainGroupToGraphQL(g *group.Group) (*model.Group, error) {
	if g == nil {
		return nil, nil
	}

	result := &model.Group{
		ID:          g.ID.String(),
		Code:        g.Code,
		Name:        g.Name,
		Description: g.Description,
		Status:      mapStatusToGraphQL(g.Status),
		TenantID:    g.TenantID.String(),
		CreatedAt:   scalars.DateTime{Time: g.CreatedAt},
	}

	if g.Type != nil {
		result.Type = *g.Type
	}

	if g.ParentGroupID != nil {
		parentGroupID := g.ParentGroupID.String()
		result.ParentGroupID = &parentGroupID
	}

	if g.Attributes != nil {
		result.Attributes = scalars.JSONObject{Data: g.Attributes}
	}

	if g.Metadata != nil {
		result.Metadata = scalars.JSONObject{Data: g.Metadata}
	}

	if g.UpdatedAt != nil {
		result.UpdatedAt = &scalars.DateTime{Time: *g.UpdatedAt}
	}

	if g.CreatedBy != nil {
		createdBy := g.CreatedBy.String()
		result.CreatedBy = &createdBy
	}

	if g.UpdatedBy != nil {
		updatedBy := g.UpdatedBy.String()
		result.UpdatedBy = &updatedBy
	}

	if g.Path != nil {
		result.Path = g.Path
	}

	if g.Level != nil {
		result.Level = g.Level
	}

	return result, nil
}

// graphQLGroupToDomain converte um grupo do modelo GraphQL para o modelo de domínio
func graphQLGroupToDomain(g *model.CreateGroupInput) (*group.Group, error) {
	if g == nil {
		return nil, errors.New("entrada inválida: grupo não pode ser nulo")
	}

	tenantUUID, err := uuid.Parse(g.TenantID)
	if err != nil {
		return nil, fmt.Errorf("ID do tenant inválido: %w", err)
	}

	domainGroup := &group.Group{
		TenantID:    tenantUUID,
		Code:        g.Code,
		Name:        g.Name,
		Description: g.Description,
		Status:      group.StatusActive, // Padrão para novos grupos
	}

	if g.Type != nil {
		domainGroup.Type = g.Type
	}

	if g.ParentGroupID != nil {
		parentGroupUUID, err := uuid.Parse(*g.ParentGroupID)
		if err != nil {
			return nil, fmt.Errorf("ID do grupo pai inválido: %w", err)
		}
		domainGroup.ParentGroupID = &parentGroupUUID
	}

	if g.Attributes != nil {
		domainGroup.Attributes = g.Attributes.Data
	}

	if g.Metadata != nil {
		domainGroup.Metadata = g.Metadata.Data
	}

	return domainGroup, nil
}

// graphQLUpdateGroupToDomain converte uma atualização de grupo do modelo GraphQL para o modelo de domínio
func graphQLUpdateGroupToDomain(g *model.UpdateGroupInput) (*group.Group, error) {
	if g == nil {
		return nil, errors.New("entrada inválida: grupo não pode ser nulo")
	}

	groupID, err := uuid.Parse(g.ID)
	if err != nil {
		return nil, fmt.Errorf("ID do grupo inválido: %w", err)
	}

	tenantUUID, err := uuid.Parse(g.TenantID)
	if err != nil {
		return nil, fmt.Errorf("ID do tenant inválido: %w", err)
	}

	domainGroup := &group.Group{
		ID:       groupID,
		TenantID: tenantUUID,
	}

	if g.Code != nil {
		domainGroup.Code = *g.Code
	}

	if g.Name != nil {
		domainGroup.Name = *g.Name
	}

	if g.Description != nil {
		domainGroup.Description = *g.Description
	}

	if g.Type != nil {
		domainGroup.Type = g.Type
	}

	if g.ParentGroupID != nil {
		parentGroupUUID, err := uuid.Parse(*g.ParentGroupID)
		if err != nil {
			return nil, fmt.Errorf("ID do grupo pai inválido: %w", err)
		}
		domainGroup.ParentGroupID = &parentGroupUUID
	}

	if g.Attributes != nil {
		domainGroup.Attributes = g.Attributes.Data
	}

	if g.Metadata != nil {
		domainGroup.Metadata = g.Metadata.Data
	}

	return domainGroup, nil
}

// graphQLFilterToDomain converte um filtro do modelo GraphQL para o modelo de domínio
func graphQLFilterToDomain(filter *model.GroupFilter) (*group.GroupFilter, error) {
	if filter == nil {
		return &group.GroupFilter{}, nil
	}

	result := group.GroupFilter{}

	if filter.IDs != nil {
		ids := make([]uuid.UUID, 0, len(filter.IDs))
		for _, id := range filter.IDs {
			uid, err := uuid.Parse(id)
			if err != nil {
				return nil, fmt.Errorf("ID inválido no filtro: %w", err)
			}
			ids = append(ids, uid)
		}
		result.IDs = ids
	}

	if filter.Codes != nil {
		result.Codes = filter.Codes
	}

	if filter.NameContains != nil {
		result.NameContains = filter.NameContains
	}

	if filter.DescriptionContains != nil {
		result.DescriptionContains = filter.DescriptionContains
	}

	if filter.Statuses != nil {
		statuses := make([]string, 0, len(filter.Statuses))
		for _, status := range filter.Statuses {
			statuses = append(statuses, mapGraphQLStatusToDomain(status))
		}
		result.Statuses = statuses
	}

	if filter.Types != nil {
		result.Types = filter.Types
	}

	if filter.ParentGroupID != nil {
		parentID, err := uuid.Parse(*filter.ParentGroupID)
		if err != nil {
			return nil, fmt.Errorf("ID do grupo pai inválido no filtro: %w", err)
		}
		result.ParentGroupID = &parentID
	}

	if filter.CreatedAtStart != nil {
		result.CreatedAtStart = &filter.CreatedAtStart.Time
	}

	if filter.CreatedAtEnd != nil {
		result.CreatedAtEnd = &filter.CreatedAtEnd.Time
	}

	if filter.UpdatedAtStart != nil {
		result.UpdatedAtStart = &filter.UpdatedAtStart.Time
	}

	if filter.UpdatedAtEnd != nil {
		result.UpdatedAtEnd = &filter.UpdatedAtEnd.Time
	}

	if filter.CreatedBy != nil {
		createdBy, err := uuid.Parse(*filter.CreatedBy)
		if err != nil {
			return nil, fmt.Errorf("ID do criador inválido no filtro: %w", err)
		}
		result.CreatedBy = &createdBy
	}

	if filter.UpdatedBy != nil {
		updatedBy, err := uuid.Parse(*filter.UpdatedBy)
		if err != nil {
			return nil, fmt.Errorf("ID do atualizador inválido no filtro: %w", err)
		}
		result.UpdatedBy = &updatedBy
	}

	if filter.HasParent != nil {
		result.HasParent = filter.HasParent
	}

	return &result, nil
}

// domainListResultToGraphQL converte um resultado de listagem do modelo de domínio para o modelo GraphQL
func domainListResultToGraphQL(result *group.GroupListResult, page, pageSize int) (*model.GroupListResult, error) {
	if result == nil {
		return nil, errors.New("resultado não pode ser nulo")
	}

	items := make([]*model.Group, 0, len(result.Items))
	for _, domainGroup := range result.Items {
		graphQLGroup, err := domainGroupToGraphQL(domainGroup)
		if err != nil {
			return nil, fmt.Errorf("erro ao converter grupo: %w", err)
		}
		items = append(items, graphQLGroup)
	}

	totalPages := result.TotalCount / pageSize
	if result.TotalCount%pageSize > 0 {
		totalPages++
	}

	return &model.GroupListResult{
		Items:      items,
		TotalCount: result.TotalCount,
		PageInfo: &model.PageInfo{
			CurrentPage:      page,
			PageSize:         pageSize,
			TotalPages:       totalPages,
			HasNextPage:      page < totalPages,
			HasPreviousPage:  page > 1,
		},
	}, nil
}

// mapStatusToGraphQL mapeia um status do domínio para o modelo GraphQL
func mapStatusToGraphQL(status string) model.GroupStatus {
	switch status {
	case group.StatusActive:
		return model.GroupStatusActive
	case group.StatusInactive:
		return model.GroupStatusInactive
	case group.StatusLocked:
		return model.GroupStatusLocked
	default:
		return model.GroupStatusInactive
	}
}

// mapGraphQLStatusToDomain mapeia um status do GraphQL para o domínio
func mapGraphQLStatusToDomain(status model.GroupStatus) string {
	switch status {
	case model.GroupStatusActive:
		return group.StatusActive
	case model.GroupStatusInactive:
		return group.StatusInactive
	case model.GroupStatusLocked:
		return group.StatusLocked
	default:
		return group.StatusInactive
	}
}