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

// Role resolve a query para buscar um papel por ID
func (r *queryResolver) Role(ctx context.Context, id string) (*model.Role, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.role",
		trace.WithAttributes(attribute.String("role.id", id)))
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: role",
		"role_id", id,
		"requester_id", authInfo.UserID,
		"tenant_id", authInfo.TenantID)

	// Obter papel do serviço
	role, err := r.roleService.GetByID(ctx, id)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to get role by ID",
			"error", err.Error(),
			"role_id", id)

		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))

		return nil, err
	}

	// Verificar acesso cross-tenant
	if role.TenantID != authInfo.TenantID && !r.config.EnableCrossTenantAccess && !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant access denied",
			"requester_tenant", authInfo.TenantID,
			"resource_tenant", role.TenantID)

		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", "cross_tenant_access_denied"))

		return nil, errors.NewForbiddenError("cross_tenant_access", "Acesso a recursos de outro tenant não permitido")
	}

	// Registrar acesso bem-sucedido
	span.SetAttributes(attribute.Bool("success", true))

	return role, nil
}

// Roles resolve a query para listar papéis com filtros e paginação
func (r *queryResolver) Roles(ctx context.Context, filter *model.RoleFilter, pagination *model.PaginationInput) (*model.RoleConnection, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.roles")
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

	// Aplicar filtro de tenant se não for administrador global
	if filter == nil {
		filter = &model.RoleFilter{}
	}

	// Forçar filtro de tenant para garantir isolamento, exceto para administradores com permissão especial
	if !authInfo.HasPermission("IAM:CrossTenantAccess") {
		filter.TenantID = authInfo.TenantID
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: roles",
		"filter", filter,
		"pagination", pagination,
		"requester_id", authInfo.UserID)

	// Adicionar atributos ao span
	span.SetAttributes(attribute.Int("pagination.page", pagination.Page))
	span.SetAttributes(attribute.Int("pagination.per_page", pagination.PerPage))
	if filter != nil && filter.TenantID != "" {
		span.SetAttributes(attribute.String("filter.tenant_id", filter.TenantID))
	}

	// Obter papéis do serviço
	result, err := r.roleService.ListRoles(ctx, filter, pagination)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to list roles", "error", err.Error())

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

// RoleByCode resolve a query para buscar um papel por código
func (r *queryResolver) RoleByCode(ctx context.Context, code string, tenantID *string) (*model.Role, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.roleByCode",
		trace.WithAttributes(attribute.String("role.code", code)))
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

		// Verificar acesso cross-tenant
		if targetTenantID != authInfo.TenantID && !r.config.EnableCrossTenantAccess && !authInfo.HasPermission("IAM:CrossTenantAccess") {
			r.logger.Warn(ctx, "Cross-tenant access denied",
				"requester_tenant", authInfo.TenantID,
				"target_tenant", targetTenantID)

			span.SetAttributes(attribute.Bool("success", false))
			span.SetAttributes(attribute.String("error", "cross_tenant_access_denied"))

			return nil, errors.NewForbiddenError("cross_tenant_access", "Acesso a recursos de outro tenant não permitido")
		}
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: roleByCode",
		"role_code", code,
		"tenant_id", targetTenantID,
		"requester_id", authInfo.UserID)

	// Adicionar atributos ao span
	span.SetAttributes(attribute.String("tenant.id", targetTenantID))

	// Obter papel do serviço
	role, err := r.roleService.GetByCode(ctx, code, targetTenantID)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to get role by code",
			"error", err.Error(),
			"role_code", code,
			"tenant_id", targetTenantID)

		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))

		return nil, err
	}

	// Registrar acesso bem-sucedido
	span.SetAttributes(attribute.Bool("success", true))

	return role, nil
}

// CreateRole resolve a mutation para criar um novo papel
func (r *mutationResolver) CreateRole(ctx context.Context, input model.CreateRoleInput) (*model.Role, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.createRole")
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: createRole",
		"input", map[string]interface{}{
			"name":        input.Name,
			"code":        input.Code,
			"description": input.Description,
			"tenantId":    input.TenantID,
		},
		"requester_id", authInfo.UserID,
		"tenant_id", authInfo.TenantID)

	// Verificar acesso cross-tenant
	if input.TenantID != authInfo.TenantID && !r.config.EnableCrossTenantAccess && !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant operation denied",
			"requester_tenant", authInfo.TenantID,
			"target_tenant", input.TenantID)

		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", "cross_tenant_operation_denied"))

		return nil, errors.NewForbiddenError("cross_tenant_operation", "Operação em outro tenant não permitida")
	}

	// Verificar permissões especiais para papéis com flags de sistema
	if input.IsSystem != nil && *input.IsSystem && !authInfo.HasPermission("IAM:ManageSystemRoles") {
		r.logger.Warn(ctx, "Permission denied for creating system role", 
			"requester_id", authInfo.UserID)
		return nil, errors.NewForbiddenError("system_role_creation", "Não possui permissão para criar papéis de sistema")
	}

	// Adicionar atributos ao span
	span.SetAttributes(attribute.String("role.name", input.Name))
	span.SetAttributes(attribute.String("role.code", input.Code))
	span.SetAttributes(attribute.String("tenant.id", input.TenantID))

	// Criar papel via serviço
	role, err := r.roleService.Create(ctx, &input)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to create role", "error", err.Error())

		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))

		return nil, err
	}

	// Registrar operação bem-sucedida
	span.SetAttributes(attribute.Bool("success", true))
	span.SetAttributes(attribute.String("role.id", role.ID))

	// Publicar evento de segurança
	if r.config.EnableAuditLogging {
		r.securityService.LogEvent(ctx, &model.SecurityEvent{
			EventType:   "ROLE_CREATED",
			Severity:    "INFO",
			UserID:      authInfo.UserID,
			TenantID:    input.TenantID,
			Description: fmt.Sprintf("Papel %s criado", input.Name),
			IPAddress:   auth.GetClientIPFromContext(ctx),
			Metadata: map[string]interface{}{
				"role_name": input.Name,
				"role_code": input.Code,
			},
		})
	}

	return role, nil
}

// UpdateRole resolve a mutation para atualizar um papel existente
func (r *mutationResolver) UpdateRole(ctx context.Context, id string, input model.UpdateRoleInput) (*model.Role, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.updateRole",
		trace.WithAttributes(attribute.String("role.id", id)))
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: updateRole",
		"role_id", id,
		"input", input,
		"requester_id", authInfo.UserID)

	// Obter o papel atual para verificar o tenant
	existingRole, err := r.roleService.GetByID(ctx, id)
	if err != nil {
		r.logger.Error(ctx, "Failed to get role for update", "error", err.Error(), "role_id", id)
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, err
	}

	// Verificar acesso cross-tenant
	if existingRole.TenantID != authInfo.TenantID && !r.config.EnableCrossTenantAccess && !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant operation denied",
			"requester_tenant", authInfo.TenantID,
			"resource_tenant", existingRole.TenantID)

		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", "cross_tenant_operation_denied"))

		return nil, errors.NewForbiddenError("cross_tenant_operation", "Operação em outro tenant não permitida")
	}

	// Verificar se está tentando modificar um papel de sistema sem permissão
	if existingRole.IsSystem && !authInfo.HasPermission("IAM:ManageSystemRoles") {
		r.logger.Warn(ctx, "Permission denied for updating system role", 
			"requester_id", authInfo.UserID,
			"role_id", id)
		return nil, errors.NewForbiddenError("system_role_update", "Não possui permissão para modificar papéis de sistema")
	}

	// Atualizar papel via serviço
	role, err := r.roleService.Update(ctx, id, &input)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to update role", "error", err.Error(), "role_id", id)

		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))

		return nil, err
	}

	// Registrar operação bem-sucedida
	span.SetAttributes(attribute.Bool("success", true))

	// Publicar evento de segurança
	if r.config.EnableAuditLogging {
		r.securityService.LogEvent(ctx, &model.SecurityEvent{
			EventType:   "ROLE_UPDATED",
			Severity:    "INFO",
			UserID:      authInfo.UserID,
			TenantID:    existingRole.TenantID,
			Description: fmt.Sprintf("Papel %s atualizado", existingRole.Name),
			IPAddress:   auth.GetClientIPFromContext(ctx),
			Metadata: map[string]interface{}{
				"role_id": id,
				"changes": input,
			},
		})
	}

	return role, nil
}

// DeleteRole resolve a mutation para excluir um papel
func (r *mutationResolver) DeleteRole(ctx context.Context, id string) (bool, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.deleteRole",
		trace.WithAttributes(attribute.String("role.id", id)))
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return false, errors.ErrUnauthorized
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: deleteRole",
		"role_id", id,
		"requester_id", authInfo.UserID)

	// Obter o papel atual para verificar o tenant
	existingRole, err := r.roleService.GetByID(ctx, id)
	if err != nil {
		r.logger.Error(ctx, "Failed to get role for deletion", "error", err.Error(), "role_id", id)
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		return false, err
	}

	// Verificar acesso cross-tenant
	if existingRole.TenantID != authInfo.TenantID && !r.config.EnableCrossTenantAccess && !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant operation denied",
			"requester_tenant", authInfo.TenantID,
			"resource_tenant", existingRole.TenantID)

		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", "cross_tenant_operation_denied"))

		return false, errors.NewForbiddenError("cross_tenant_operation", "Operação em outro tenant não permitida")
	}

	// Verificar se está tentando excluir um papel de sistema sem permissão
	if existingRole.IsSystem && !authInfo.HasPermission("IAM:ManageSystemRoles") {
		r.logger.Warn(ctx, "Permission denied for deleting system role", 
			"requester_id", authInfo.UserID,
			"role_id", id)
		return false, errors.NewForbiddenError("system_role_deletion", "Não possui permissão para excluir papéis de sistema")
	}

	// Verificar se há usuários ou grupos associados ao papel antes de excluir
	hasAssignments, err := r.roleService.HasAssignments(ctx, id)
	if err != nil {
		r.logger.Error(ctx, "Failed to check role assignments", "error", err.Error(), "role_id", id)
		return false, err
	}

	// Se o papel está associado a usuários ou grupos e não tem permissão para forçar exclusão
	if hasAssignments && !authInfo.HasPermission("IAM:ForceDeleteRole") {
		r.logger.Warn(ctx, "Cannot delete role with assignments", "role_id", id)
		return false, errors.NewBusinessError("role_has_assignments", "Não é possível excluir um papel que está atribuído a usuários ou grupos")
	}

	// Excluir papel via serviço
	err = r.roleService.Delete(ctx, id)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to delete role", "error", err.Error(), "role_id", id)

		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))

		return false, err
	}

	// Registrar operação bem-sucedida
	span.SetAttributes(attribute.Bool("success", true))

	// Publicar evento de segurança
	if r.config.EnableAuditLogging {
		r.securityService.LogEvent(ctx, &model.SecurityEvent{
			EventType:   "ROLE_DELETED",
			Severity:    "WARN",
			UserID:      authInfo.UserID,
			TenantID:    existingRole.TenantID,
			Description: fmt.Sprintf("Papel %s excluído", existingRole.Name),
			IPAddress:   auth.GetClientIPFromContext(ctx),
			Metadata: map[string]interface{}{
				"role_id":   id,
				"role_name": existingRole.Name,
				"role_code": existingRole.Code,
			},
		})
	}

	return true, nil
}