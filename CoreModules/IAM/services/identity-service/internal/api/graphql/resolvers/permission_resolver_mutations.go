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

// CreatePermission resolve a mutation para criar uma nova permissão
func (r *mutationResolver) CreatePermission(ctx context.Context, input model.CreatePermissionInput) (*model.Permission, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.createPermission")
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verificar permissão para criar permissões
	if !authInfo.HasPermission("IAM:ManagePermissions") {
		r.logger.Warn(ctx, "Permission denied for creating permission", 
			"requester_id", authInfo.UserID)
		return nil, errors.NewForbiddenError("permission_denied", "Permissão insuficiente para criar permissões")
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: createPermission", 
		"input", map[string]interface{}{
			"name": input.Name,
			"code": input.Code,
			"description": input.Description,
			"resource": input.Resource,
			"action": input.Action,
			"tenantId": input.TenantID,
		},
		"requester_id", authInfo.UserID)

	// Verificar acesso cross-tenant se a permissão for específica de tenant
	if input.TenantID != nil && *input.TenantID != "" && 
	   *input.TenantID != authInfo.TenantID && 
	   !r.config.EnableCrossTenantAccess && 
	   !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant operation denied", 
			"requester_tenant", authInfo.TenantID,
			"target_tenant", *input.TenantID)
			
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", "cross_tenant_operation_denied"))
		
		return nil, errors.NewForbiddenError("cross_tenant_operation", "Operação em outro tenant não permitida")
	}

	// Verificar permissão especial para criar permissões de sistema
	if input.IsSystem != nil && *input.IsSystem && !authInfo.HasPermission("IAM:ManageSystemPermissions") {
		r.logger.Warn(ctx, "Permission denied for creating system permission", 
			"requester_id", authInfo.UserID)
		return nil, errors.NewForbiddenError("system_permission_creation", "Não possui permissão para criar permissões de sistema")
	}

	// Adicionar atributos ao span
	span.SetAttributes(attribute.String("permission.name", input.Name))
	span.SetAttributes(attribute.String("permission.code", input.Code))
	span.SetAttributes(attribute.String("permission.resource", input.Resource))
	span.SetAttributes(attribute.String("permission.action", input.Action))
	if input.TenantID != nil {
		span.SetAttributes(attribute.String("tenant.id", *input.TenantID))
	}

	// Criar permissão via serviço
	permission, err := r.permissionService.Create(ctx, &input)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to create permission", "error", err.Error())
		
		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		
		return nil, err
	}

	// Registrar operação bem-sucedida
	span.SetAttributes(attribute.Bool("success", true))
	span.SetAttributes(attribute.String("permission.id", permission.ID))

	// Publicar evento de segurança
	if r.config.EnableAuditLogging {
		tenantID := ""
		if input.TenantID != nil {
			tenantID = *input.TenantID
		}
		
		r.securityService.LogEvent(ctx, &model.SecurityEvent{
			EventType:   "PERMISSION_CREATED",
			Severity:    "INFO",
			UserID:      authInfo.UserID,
			TenantID:    tenantID,
			Description: fmt.Sprintf("Permissão %s criada", input.Name),
			IPAddress:   auth.GetClientIPFromContext(ctx),
			Metadata: map[string]interface{}{
				"permission_name": input.Name,
				"permission_code": input.Code,
				"resource": input.Resource,
				"action": input.Action,
			},
		})
	}

	return permission, nil
}

// UpdatePermission resolve a mutation para atualizar uma permissão existente
func (r *mutationResolver) UpdatePermission(ctx context.Context, id string, input model.UpdatePermissionInput) (*model.Permission, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.updatePermission", 
		trace.WithAttributes(attribute.String("permission.id", id)))
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verificar permissão para gerenciar permissões
	if !authInfo.HasPermission("IAM:ManagePermissions") {
		r.logger.Warn(ctx, "Permission denied for updating permission", 
			"requester_id", authInfo.UserID)
		return nil, errors.NewForbiddenError("permission_denied", "Permissão insuficiente para atualizar permissões")
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: updatePermission", 
		"permission_id", id, 
		"input", input,
		"requester_id", authInfo.UserID)

	// Obter a permissão atual para verificar o tenant
	existingPermission, err := r.permissionService.GetByID(ctx, id)
	if err != nil {
		r.logger.Error(ctx, "Failed to get permission for update", "error", err.Error(), "permission_id", id)
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, err
	}

	// Verificar acesso cross-tenant se a permissão for específica de tenant
	if existingPermission.TenantID != nil && *existingPermission.TenantID != "" && 
	   *existingPermission.TenantID != authInfo.TenantID && 
	   !r.config.EnableCrossTenantAccess && 
	   !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant operation denied", 
			"requester_tenant", authInfo.TenantID,
			"resource_tenant", *existingPermission.TenantID)
			
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", "cross_tenant_operation_denied"))
		
		return nil, errors.NewForbiddenError("cross_tenant_operation", "Operação em outro tenant não permitida")
	}

	// Verificar permissão especial para atualizar permissões de sistema
	if existingPermission.IsSystem && !authInfo.HasPermission("IAM:ManageSystemPermissions") {
		r.logger.Warn(ctx, "Permission denied for updating system permission", 
			"requester_id", authInfo.UserID,
			"permission_id", id)
		return nil, errors.NewForbiddenError("system_permission_update", "Não possui permissão para modificar permissões de sistema")
	}

	// Atualizar permissão via serviço
	permission, err := r.permissionService.Update(ctx, id, &input)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to update permission", "error", err.Error(), "permission_id", id)
		
		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		
		return nil, err
	}

	// Registrar operação bem-sucedida
	span.SetAttributes(attribute.Bool("success", true))

	// Publicar evento de segurança
	if r.config.EnableAuditLogging {
		tenantID := ""
		if existingPermission.TenantID != nil {
			tenantID = *existingPermission.TenantID
		}
		
		r.securityService.LogEvent(ctx, &model.SecurityEvent{
			EventType:   "PERMISSION_UPDATED",
			Severity:    "INFO",
			UserID:      authInfo.UserID,
			TenantID:    tenantID,
			Description: fmt.Sprintf("Permissão %s atualizada", existingPermission.Name),
			IPAddress:   auth.GetClientIPFromContext(ctx),
			Metadata: map[string]interface{}{
				"permission_id": id,
				"changes": input,
			},
		})
	}

	return permission, nil
}

// DeletePermission resolve a mutation para excluir uma permissão
func (r *mutationResolver) DeletePermission(ctx context.Context, id string) (bool, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.deletePermission", 
		trace.WithAttributes(attribute.String("permission.id", id)))
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return false, errors.ErrUnauthorized
	}

	// Verificar permissão para gerenciar permissões
	if !authInfo.HasPermission("IAM:ManagePermissions") {
		r.logger.Warn(ctx, "Permission denied for deleting permission", 
			"requester_id", authInfo.UserID)
		return false, errors.NewForbiddenError("permission_denied", "Permissão insuficiente para excluir permissões")
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: deletePermission", 
		"permission_id", id, 
		"requester_id", authInfo.UserID)

	// Obter a permissão atual para verificar o tenant e outros atributos
	existingPermission, err := r.permissionService.GetByID(ctx, id)
	if err != nil {
		r.logger.Error(ctx, "Failed to get permission for deletion", "error", err.Error(), "permission_id", id)
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		return false, err
	}

	// Verificar acesso cross-tenant se a permissão for específica de tenant
	if existingPermission.TenantID != nil && *existingPermission.TenantID != "" && 
	   *existingPermission.TenantID != authInfo.TenantID && 
	   !r.config.EnableCrossTenantAccess && 
	   !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant operation denied", 
			"requester_tenant", authInfo.TenantID,
			"resource_tenant", *existingPermission.TenantID)
			
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", "cross_tenant_operation_denied"))
		
		return false, errors.NewForbiddenError("cross_tenant_operation", "Operação em outro tenant não permitida")
	}

	// Verificar permissão especial para excluir permissões de sistema
	if existingPermission.IsSystem && !authInfo.HasPermission("IAM:ManageSystemPermissions") {
		r.logger.Warn(ctx, "Permission denied for deleting system permission", 
			"requester_id", authInfo.UserID,
			"permission_id", id)
		return false, errors.NewForbiddenError("system_permission_deletion", "Não possui permissão para excluir permissões de sistema")
	}

	// Verificar se há papéis ou usuários usando esta permissão
	hasReferences, err := r.permissionService.HasReferences(ctx, id)
	if err != nil {
		r.logger.Error(ctx, "Failed to check permission references", "error", err.Error(), "permission_id", id)
		return false, err
	}

	// Se a permissão está sendo usada e não tem permissão para forçar exclusão
	if hasReferences && !authInfo.HasPermission("IAM:ForceDeletePermission") {
		r.logger.Warn(ctx, "Cannot delete permission with references", "permission_id", id)
		return false, errors.NewBusinessError("permission_has_references", "Não é possível excluir uma permissão que está sendo usada por papéis ou usuários")
	}

	// Excluir permissão via serviço
	err = r.permissionService.Delete(ctx, id)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to delete permission", "error", err.Error(), "permission_id", id)
		
		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		
		return false, err
	}

	// Registrar operação bem-sucedida
	span.SetAttributes(attribute.Bool("success", true))

	// Publicar evento de segurança
	if r.config.EnableAuditLogging {
		tenantID := ""
		if existingPermission.TenantID != nil {
			tenantID = *existingPermission.TenantID
		}
		
		r.securityService.LogEvent(ctx, &model.SecurityEvent{
			EventType:   "PERMISSION_DELETED",
			Severity:    "WARN",
			UserID:      authInfo.UserID,
			TenantID:    tenantID,
			Description: fmt.Sprintf("Permissão %s excluída", existingPermission.Name),
			IPAddress:   auth.GetClientIPFromContext(ctx),
			Metadata: map[string]interface{}{
				"permission_id":   id,
				"permission_name": existingPermission.Name,
				"permission_code": existingPermission.Code,
			},
		})
	}

	return true, nil
}