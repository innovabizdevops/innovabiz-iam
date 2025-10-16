package resolvers

import (
	"context"
	"fmt"

	"github.com/innovabiz/iam/internal/domain/model"
	"github.com/innovabiz/iam/internal/domain/model/errors"
	"github.com/innovabiz/iam/internal/infrastructure/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// GroupResolver contém os resolvers relacionados a grupos
type groupResolver struct {
	*Resolver
}

// Group resolve a query para buscar um grupo por ID
func (r *queryResolver) Group(ctx context.Context, id string) (*model.Group, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.group",
		trace.WithAttributes(attribute.String("group.id", id)))
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: group",
		"group_id", id,
		"requester_id", authInfo.UserID,
		"tenant_id", authInfo.TenantID)

	// Obter grupo do serviço
	group, err := r.groupService.GetByID(ctx, id)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to get group by ID",
			"error", err.Error(),
			"group_id", id)

		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))

		return nil, err
	}

	// Verificar se o usuário tem acesso cross-tenant
	if group.TenantID != authInfo.TenantID && !r.config.EnableCrossTenantAccess && !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant access denied",
			"requester_tenant", authInfo.TenantID,
			"resource_tenant", group.TenantID)

		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", "cross_tenant_access_denied"))

		return nil, errors.NewForbiddenError("cross_tenant_access", "Acesso a recursos de outro tenant não permitido")
	}

	// Registrar acesso bem-sucedido
	span.SetAttributes(attribute.Bool("success", true))

	return group, nil
}

// Groups resolve a query para listar grupos com filtros e paginação
func (r *queryResolver) Groups(ctx context.Context, filter *model.GroupFilter, pagination *model.PaginationInput) (*model.GroupConnection, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.groups")
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Aplicar configurações de paginação padrão se não fornecidas
	if pagination == nil {
		pagination = &model.PaginationInput{
			Page:         0,
			PerPage:      r.config.DefaultPageSize,
			SortDirection: model.SortDirectionASC,
		}
	}

	// Limitar o tamanho da página
	if pagination.PerPage > r.config.MaxPageSize {
		pagination.PerPage = r.config.MaxPageSize
		r.logger.Warn(ctx, "Requested page size exceeded maximum, using max page size instead",
			"requested_size", pagination.PerPage,
			"max_size", r.config.MaxPageSize)
	}

	// Aplicar filtro de tenant se não for administrador global
	if filter == nil {
		filter = &model.GroupFilter{}
	}

	// Forçar filtro de tenant para garantir isolamento, exceto para administradores com permissão especial
	if !authInfo.HasPermission("IAM:CrossTenantAccess") {
		filter.TenantID = authInfo.TenantID
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: groups",
		"filter", filter,
		"pagination", pagination,
		"requester_id", authInfo.UserID)

	// Adicionar atributos ao span
	span.SetAttributes(attribute.Int("pagination.page", pagination.Page))
	span.SetAttributes(attribute.Int("pagination.per_page", pagination.PerPage))
	if filter != nil && filter.TenantID != "" {
		span.SetAttributes(attribute.String("filter.tenant_id", filter.TenantID))
	}

	// Obter grupos do serviço
	result, err := r.groupService.ListGroups(ctx, filter, pagination)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to list groups", "error", err.Error())

		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))

		return nil, err
	}

	// Registrar acesso bem-sucedido
	span.SetAttributes(attribute.Bool("success", true))
	span.SetAttributes(attribute.Int("result.total_count", result.TotalCount))
	span.SetAttributes(attribute.Int("result.items_count", len(result.Items)))

	return result, nil
}

// GroupHierarchy resolve a query para obter a hierarquia de grupos a partir de um grupo raiz
func (r *queryResolver) GroupHierarchy(ctx context.Context, rootGroupID string) (*model.GroupHierarchy, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.groupHierarchy",
		trace.WithAttributes(attribute.String("root_group_id", rootGroupID)))
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: groupHierarchy",
		"root_group_id", rootGroupID,
		"requester_id", authInfo.UserID)

	// Obter grupo raiz para verificar o tenant
	rootGroup, err := r.groupService.GetByID(ctx, rootGroupID)
	if err != nil {
		r.logger.Error(ctx, "Failed to get root group",
			"error", err.Error(),
			"group_id", rootGroupID)

		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))

		return nil, err
	}

	// Verificar se o usuário tem acesso cross-tenant
	if rootGroup.TenantID != authInfo.TenantID && !r.config.EnableCrossTenantAccess && !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant access denied",
			"requester_tenant", authInfo.TenantID,
			"resource_tenant", rootGroup.TenantID)

		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", "cross_tenant_access_denied"))

		return nil, errors.NewForbiddenError("cross_tenant_access", "Acesso a recursos de outro tenant não permitido")
	}

	// Obter hierarquia de grupos do serviço
	hierarchy, err := r.groupService.GetHierarchy(ctx, rootGroupID)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to get group hierarchy",
			"error", err.Error(),
			"root_group_id", rootGroupID)

		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))

		return nil, err
	}

	// Registrar acesso bem-sucedido
	span.SetAttributes(attribute.Bool("success", true))
	span.SetAttributes(attribute.Int("hierarchy.levels", hierarchy.Levels))

	return hierarchy, nil
}