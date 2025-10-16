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

// CreateGroup resolve a mutation para criar um novo grupo
func (r *mutationResolver) CreateGroup(ctx context.Context, input model.CreateGroupInput) (*model.Group, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.createGroup")
	defer span.End()
	
	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}
	
	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: createGroup", 
		"input", map[string]interface{}{
			"name": input.Name,
			"code": input.Code,
			"description": input.Description,
			"tenantId": input.TenantID,
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
	
	// Adicionar atributos ao span
	span.SetAttributes(attribute.String("group.name", input.Name))
	span.SetAttributes(attribute.String("group.code", input.Code))
	span.SetAttributes(attribute.String("tenant.id", input.TenantID))

	// Criar grupo via serviço
	group, err := r.groupService.Create(ctx, &input)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to create group", "error", err.Error())
		
		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		
		return nil, err
	}
	
	// Registrar operação bem-sucedida
	span.SetAttributes(attribute.Bool("success", true))
	span.SetAttributes(attribute.String("group.id", group.ID))
	
	// Publicar evento de segurança
	if r.config.EnableAuditLogging {
		r.securityService.LogEvent(ctx, &model.SecurityEvent{
			EventType:   "GROUP_CREATED",
			Severity:    "INFO",
			UserID:      authInfo.UserID,
			TenantID:    input.TenantID,
			Description: fmt.Sprintf("Grupo %s criado", input.Name),
			IPAddress:   auth.GetClientIPFromContext(ctx),
			Metadata: map[string]interface{}{
				"group_name": input.Name,
				"group_code": input.Code,
			},
		})
	}
	
	return group, nil
}

// UpdateGroup resolve a mutation para atualizar um grupo existente
func (r *mutationResolver) UpdateGroup(ctx context.Context, id string, input model.UpdateGroupInput) (*model.Group, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.updateGroup", 
		trace.WithAttributes(attribute.String("group.id", id)))
	defer span.End()
	
	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}
	
	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: updateGroup", 
		"group_id", id, 
		"input", input,
		"requester_id", authInfo.UserID)
	
	// Obter o grupo atual para verificar o tenant
	existingGroup, err := r.groupService.GetByID(ctx, id)
	if err != nil {
		r.logger.Error(ctx, "Failed to get group for update", "error", err.Error(), "group_id", id)
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, err
	}
	
	// Verificar acesso cross-tenant
	if existingGroup.TenantID != authInfo.TenantID && !r.config.EnableCrossTenantAccess && !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant operation denied", 
			"requester_tenant", authInfo.TenantID,
			"resource_tenant", existingGroup.TenantID)
			
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", "cross_tenant_operation_denied"))
		
		return nil, errors.NewForbiddenError("cross_tenant_operation", "Operação em outro tenant não permitida")
	}
	
	// Atualizar grupo via serviço
	group, err := r.groupService.Update(ctx, id, &input)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to update group", "error", err.Error(), "group_id", id)
		
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
			EventType:   "GROUP_UPDATED",
			Severity:    "INFO",
			UserID:      authInfo.UserID,
			TenantID:    existingGroup.TenantID,
			Description: fmt.Sprintf("Grupo %s atualizado", existingGroup.Name),
			IPAddress:   auth.GetClientIPFromContext(ctx),
			Metadata: map[string]interface{}{
				"group_id": id,
				"changes": input,
			},
		})
	}
	
	return group, nil
}

// DeleteGroup resolve a mutation para excluir um grupo
func (r *mutationResolver) DeleteGroup(ctx context.Context, id string) (bool, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.deleteGroup", 
		trace.WithAttributes(attribute.String("group.id", id)))
	defer span.End()
	
	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return false, errors.ErrUnauthorized
	}
	
	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: deleteGroup", 
		"group_id", id, 
		"requester_id", authInfo.UserID)
	
	// Obter o grupo atual para verificar o tenant
	existingGroup, err := r.groupService.GetByID(ctx, id)
	if err != nil {
		r.logger.Error(ctx, "Failed to get group for deletion", "error", err.Error(), "group_id", id)
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		return false, err
	}
	
	// Verificar acesso cross-tenant
	if existingGroup.TenantID != authInfo.TenantID && !r.config.EnableCrossTenantAccess && !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant operation denied", 
			"requester_tenant", authInfo.TenantID,
			"resource_tenant", existingGroup.TenantID)
			
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", "cross_tenant_operation_denied"))
		
		return false, errors.NewForbiddenError("cross_tenant_operation", "Operação em outro tenant não permitida")
	}
	
	// Verificar se há membros no grupo antes de excluir
	hasMembers, err := r.groupService.HasMembers(ctx, id)
	if err != nil {
		r.logger.Error(ctx, "Failed to check group members", "error", err.Error(), "group_id", id)
		return false, err
	}
	
	// Se o grupo tem membros, não permitir exclusão direta
	if hasMembers && !authInfo.HasPermission("IAM:ForceDeleteGroup") {
		r.logger.Warn(ctx, "Cannot delete group with members", "group_id", id)
		return false, errors.NewBusinessError("group_has_members", "Não é possível excluir um grupo que possui membros")
	}
	
	// Excluir grupo via serviço
	err = r.groupService.Delete(ctx, id)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to delete group", "error", err.Error(), "group_id", id)
		
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
			EventType:   "GROUP_DELETED",
			Severity:    "WARN",
			UserID:      authInfo.UserID,
			TenantID:    existingGroup.TenantID,
			Description: fmt.Sprintf("Grupo %s excluído", existingGroup.Name),
			IPAddress:   auth.GetClientIPFromContext(ctx),
			Metadata: map[string]interface{}{
				"group_id":   id,
				"group_name": existingGroup.Name,
				"group_code": existingGroup.Code,
			},
		})
	}
	
	return true, nil
}

// AddGroupMember resolve a mutation para adicionar um membro a um grupo
func (r *mutationResolver) AddGroupMember(ctx context.Context, groupID string, userID string, role *model.GroupMemberRole) (*model.GroupMembership, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.addGroupMember",
		trace.WithAttributes(
			attribute.String("group.id", groupID),
			attribute.String("user.id", userID),
		))
	defer span.End()
	
	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}
	
	// Definir papel padrão se não for especificado
	memberRole := model.GroupMemberRoleMEMBER
	if role != nil {
		memberRole = *role
	}
	
	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: addGroupMember", 
		"group_id", groupID,
		"user_id", userID,
		"role", memberRole,
		"requester_id", authInfo.UserID)
	
	// Obter o grupo para verificar o tenant
	group, err := r.groupService.GetByID(ctx, groupID)
	if err != nil {
		r.logger.Error(ctx, "Failed to get group", "error", err.Error(), "group_id", groupID)
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, err
	}
	
	// Verificar acesso cross-tenant para o grupo
	if group.TenantID != authInfo.TenantID && !r.config.EnableCrossTenantAccess && !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant operation denied", 
			"requester_tenant", authInfo.TenantID,
			"resource_tenant", group.TenantID)
			
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", "cross_tenant_operation_denied"))
		
		return nil, errors.NewForbiddenError("cross_tenant_operation", "Operação em outro tenant não permitida")
	}
	
	// Obter o usuário para verificar o tenant
	user, err := r.userService.GetByID(ctx, userID)
	if err != nil {
		r.logger.Error(ctx, "Failed to get user", "error", err.Error(), "user_id", userID)
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, err
	}
	
	// Verificar se o usuário e o grupo pertencem ao mesmo tenant
	if user.TenantID != group.TenantID {
		r.logger.Warn(ctx, "User and group belong to different tenants",
			"user_tenant", user.TenantID,
			"group_tenant", group.TenantID)
		
		return nil, errors.NewBusinessError("tenant_mismatch", "Usuário e grupo devem pertencer ao mesmo tenant")
	}
	
	// Adicionar membro ao grupo via serviço
	membership, err := r.groupService.AddMember(ctx, groupID, userID, memberRole)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to add group member", "error", err.Error())
		
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
			EventType:   "GROUP_MEMBER_ADDED",
			Severity:    "INFO",
			UserID:      authInfo.UserID,
			TenantID:    group.TenantID,
			Description: fmt.Sprintf("Usuário %s adicionado ao grupo %s", user.Username, group.Name),
			IPAddress:   auth.GetClientIPFromContext(ctx),
			Metadata: map[string]interface{}{
				"group_id": groupID,
				"user_id":  userID,
				"role":     memberRole,
			},
		})
	}
	
	return membership, nil
}