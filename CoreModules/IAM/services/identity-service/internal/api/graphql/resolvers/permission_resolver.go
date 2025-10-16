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

// Permission resolve a query para buscar uma permissão por ID
func (r *queryResolver) Permission(ctx context.Context, id string) (*model.Permission, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.permission",
		trace.WithAttributes(attribute.String("permission.id", id)))
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: permission",
		"permission_id", id,
		"requester_id", authInfo.UserID,
		"tenant_id", authInfo.TenantID)

	// Obter permissão do serviço
	permission, err := r.permissionService.GetByID(ctx, id)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to get permission by ID",
			"error", err.Error(),
			"permission_id", id)

		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))

		return nil, err
	}

	// Verificar acesso cross-tenant para permissões específicas de tenant
	if permission.TenantID != nil && *permission.TenantID != "" {
		if *permission.TenantID != authInfo.TenantID && !r.config.EnableCrossTenantAccess && !authInfo.HasPermission("IAM:CrossTenantAccess") {
			r.logger.Warn(ctx, "Cross-tenant access denied",
				"requester_tenant", authInfo.TenantID,
				"resource_tenant", *permission.TenantID)

			span.SetAttributes(attribute.Bool("success", false))
			span.SetAttributes(attribute.String("error", "cross_tenant_access_denied"))

			return nil, errors.NewForbiddenError("cross_tenant_access", "Acesso a recursos de outro tenant não permitido")
		}
	}

	// Registrar acesso bem-sucedido
	span.SetAttributes(attribute.Bool("success", true))

	return permission, nil
}

// Permissions resolve a query para listar permissões com filtros e paginação
func (r *queryResolver) Permissions(ctx context.Context, filter *model.PermissionFilter, pagination *model.PaginationInput) (*model.PermissionConnection, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.permissions")
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Aplicar configurações de paginação padrão se não fornecidas
	if pagination == nil {
		pagination = &model.PaginationInput{
			Page:          0,
			PerPage:       r.config.DefaultPageSize,
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

	// Aplicar filtro se não fornecido
	if filter == nil {
		filter = &model.PermissionFilter{}
	}

	// Forçar filtro de tenant para garantir isolamento, exceto para administradores com permissão especial
	// Permissões globais (sem tenantID) serão sempre visíveis independentemente do tenant do usuário
	if !authInfo.HasPermission("IAM:CrossTenantAccess") && filter.TenantID == nil {
		tenantID := authInfo.TenantID
		filter.TenantID = &tenantID
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: permissions",
		"filter", filter,
		"pagination", pagination,
		"requester_id", authInfo.UserID)

	// Adicionar atributos ao span
	span.SetAttributes(attribute.Int("pagination.page", pagination.Page))
	span.SetAttributes(attribute.Int("pagination.per_page", pagination.PerPage))
	if filter != nil && filter.TenantID != nil {
		span.SetAttributes(attribute.String("filter.tenant_id", *filter.TenantID))
	}

	// Obter permissões do serviço
	result, err := r.permissionService.ListPermissions(ctx, filter, pagination)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to list permissions", "error", err.Error())

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

// PermissionByCode resolve a query para buscar uma permissão por código
func (r *queryResolver) PermissionByCode(ctx context.Context, code string, tenantID *string) (*model.Permission, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.permissionByCode",
		trace.WithAttributes(attribute.String("permission.code", code)))
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Determinar tenant alvo
	targetTenantID := authInfo.TenantID
	if tenantID != nil {
		targetTenantID = *tenantID

		// Verificar acesso cross-tenant quando explicitamente solicitado outro tenant
		if targetTenantID != "" && targetTenantID != authInfo.TenantID && 
		   !r.config.EnableCrossTenantAccess && !authInfo.HasPermission("IAM:CrossTenantAccess") {
			r.logger.Warn(ctx, "Cross-tenant access denied",
				"requester_tenant", authInfo.TenantID,
				"target_tenant", targetTenantID)

			span.SetAttributes(attribute.Bool("success", false))
			span.SetAttributes(attribute.String("error", "cross_tenant_access_denied"))

			return nil, errors.NewForbiddenError("cross_tenant_access", "Acesso a recursos de outro tenant não permitido")
		}
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: permissionByCode",
		"permission_code", code,
		"tenant_id", targetTenantID,
		"requester_id", authInfo.UserID)

	// Adicionar atributos ao span
	if targetTenantID != "" {
		span.SetAttributes(attribute.String("tenant.id", targetTenantID))
	}

	// Obter permissão do serviço
	permission, err := r.permissionService.GetByCode(ctx, code, targetTenantID)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to get permission by code",
			"error", err.Error(),
			"permission_code", code,
			"tenant_id", targetTenantID)

		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))

		return nil, err
	}

	// Registrar acesso bem-sucedido
	span.SetAttributes(attribute.Bool("success", true))

	return permission, nil
}