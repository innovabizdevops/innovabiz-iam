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

// RemoveGroupMember resolve a mutation para remover um membro de um grupo
func (r *mutationResolver) RemoveGroupMember(ctx context.Context, groupID string, userID string) (bool, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.removeGroupMember",
		trace.WithAttributes(
			attribute.String("group.id", groupID),
			attribute.String("user.id", userID),
		))
	defer span.End()
	
	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return false, errors.ErrUnauthorized
	}
	
	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: removeGroupMember", 
		"group_id", groupID,
		"user_id", userID,
		"requester_id", authInfo.UserID)
	
	// Obter o grupo para verificar o tenant
	group, err := r.groupService.GetByID(ctx, groupID)
	if err != nil {
		r.logger.Error(ctx, "Failed to get group", "error", err.Error(), "group_id", groupID)
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		return false, err
	}
	
	// Verificar acesso cross-tenant
	if group.TenantID != authInfo.TenantID && !r.config.EnableCrossTenantAccess && !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant operation denied", 
			"requester_tenant", authInfo.TenantID,
			"resource_tenant", group.TenantID)
			
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", "cross_tenant_operation_denied"))
		
		return false, errors.NewForbiddenError("cross_tenant_operation", "Operação em outro tenant não permitida")
	}
	
	// Verificar se o usuário é membro do grupo
	isMember, err := r.groupService.IsMember(ctx, groupID, userID)
	if err != nil {
		r.logger.Error(ctx, "Failed to check group membership", "error", err.Error())
		return false, err
	}
	
	if !isMember {
		r.logger.Warn(ctx, "User is not a member of the group",
			"user_id", userID,
			"group_id", groupID)
			
		return false, errors.NewBusinessError("not_a_member", "O usuário não é membro deste grupo")
	}
	
	// Remover membro do grupo via serviço
	err = r.groupService.RemoveMember(ctx, groupID, userID)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to remove group member", "error", err.Error())
		
		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		
		return false, err
	}
	
	// Registrar operação bem-sucedida
	span.SetAttributes(attribute.Bool("success", true))
	
	// Publicar evento de segurança
	if r.config.EnableAuditLogging {
		// Buscar informações do usuário para o evento
		user, _ := r.userService.GetByID(ctx, userID)
		username := "unknown"
		if user != nil {
			username = user.Username
		}
		
		r.securityService.LogEvent(ctx, &model.SecurityEvent{
			EventType:   "GROUP_MEMBER_REMOVED",
			Severity:    "INFO",
			UserID:      authInfo.UserID,
			TenantID:    group.TenantID,
			Description: fmt.Sprintf("Usuário %s removido do grupo %s", username, group.Name),
			IPAddress:   auth.GetClientIPFromContext(ctx),
			Metadata: map[string]interface{}{
				"group_id": groupID,
				"user_id":  userID,
			},
		})
	}
	
	return true, nil
}

// UpdateGroupMemberRole resolve a mutation para atualizar o papel de um membro no grupo
func (r *mutationResolver) UpdateGroupMemberRole(ctx context.Context, groupID string, userID string, role model.GroupMemberRole) (*model.GroupMembership, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.updateGroupMemberRole",
		trace.WithAttributes(
			attribute.String("group.id", groupID),
			attribute.String("user.id", userID),
			attribute.String("role", string(role)),
		))
	defer span.End()
	
	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}
	
	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: updateGroupMemberRole", 
		"group_id", groupID,
		"user_id", userID,
		"role", role,
		"requester_id", authInfo.UserID)
	
	// Obter o grupo para verificar o tenant
	group, err := r.groupService.GetByID(ctx, groupID)
	if err != nil {
		r.logger.Error(ctx, "Failed to get group", "error", err.Error(), "group_id", groupID)
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, err
	}
	
	// Verificar acesso cross-tenant
	if group.TenantID != authInfo.TenantID && !r.config.EnableCrossTenantAccess && !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant operation denied", 
			"requester_tenant", authInfo.TenantID,
			"resource_tenant", group.TenantID)
			
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", "cross_tenant_operation_denied"))
		
		return nil, errors.NewForbiddenError("cross_tenant_operation", "Operação em outro tenant não permitida")
	}
	
	// Verificar se o usuário é membro do grupo
	isMember, err := r.groupService.IsMember(ctx, groupID, userID)
	if err != nil {
		r.logger.Error(ctx, "Failed to check group membership", "error", err.Error())
		return nil, err
	}
	
	if !isMember {
		r.logger.Warn(ctx, "User is not a member of the group",
			"user_id", userID,
			"group_id", groupID)
			
		return nil, errors.NewBusinessError("not_a_member", "O usuário não é membro deste grupo")
	}
	
	// Verificar permissão para OWNER e ADMIN (papéis especiais)
	if (role == model.GroupMemberRoleOWNER || role == model.GroupMemberRoleADMIN) && 
	   !authInfo.HasPermission("IAM:AssignGroupAdminRole") {
		r.logger.Warn(ctx, "Permission denied for assigning admin/owner role", 
			"requester_id", authInfo.UserID,
			"role", role)
			
		return nil, errors.NewForbiddenError("role_assignment_denied", "Permissão insuficiente para atribuir papel administrativo")
	}
	
	// Atualizar papel do membro via serviço
	membership, err := r.groupService.UpdateMemberRole(ctx, groupID, userID, role)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to update group member role", "error", err.Error())
		
		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		
		return nil, err
	}
	
	// Registrar operação bem-sucedida
	span.SetAttributes(attribute.Bool("success", true))
	
	// Publicar evento de segurança
	if r.config.EnableAuditLogging {
		// Buscar informações do usuário para o evento
		user, _ := r.userService.GetByID(ctx, userID)
		username := "unknown"
		if user != nil {
			username = user.Username
		}
		
		r.securityService.LogEvent(ctx, &model.SecurityEvent{
			EventType:   "GROUP_MEMBER_ROLE_UPDATED",
			Severity:    "INFO",
			UserID:      authInfo.UserID,
			TenantID:    group.TenantID,
			Description: fmt.Sprintf("Papel do usuário %s no grupo %s atualizado para %s", username, group.Name, role),
			IPAddress:   auth.GetClientIPFromContext(ctx),
			Metadata: map[string]interface{}{
				"group_id": groupID,
				"user_id":  userID,
				"role":     role,
			},
		})
	}
	
	return membership, nil
}

// AddGroupToParent resolve a mutation para adicionar um grupo como filho de outro grupo
func (r *mutationResolver) AddGroupToParent(ctx context.Context, childID string, parentID string) (bool, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.addGroupToParent",
		trace.WithAttributes(
			attribute.String("child_group.id", childID),
			attribute.String("parent_group.id", parentID),
		))
	defer span.End()
	
	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return false, errors.ErrUnauthorized
	}
	
	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: addGroupToParent", 
		"child_id", childID,
		"parent_id", parentID,
		"requester_id", authInfo.UserID)
	
	// Obter o grupo filho
	childGroup, err := r.groupService.GetByID(ctx, childID)
	if err != nil {
		r.logger.Error(ctx, "Failed to get child group", "error", err.Error(), "group_id", childID)
		return false, err
	}
	
	// Obter o grupo pai
	parentGroup, err := r.groupService.GetByID(ctx, parentID)
	if err != nil {
		r.logger.Error(ctx, "Failed to get parent group", "error", err.Error(), "group_id", parentID)
		return false, err
	}
	
	// Verificar se os grupos pertencem ao mesmo tenant
	if childGroup.TenantID != parentGroup.TenantID {
		r.logger.Warn(ctx, "Groups belong to different tenants",
			"child_tenant", childGroup.TenantID,
			"parent_tenant", parentGroup.TenantID)
		
		return false, errors.NewBusinessError("tenant_mismatch", "Os grupos devem pertencer ao mesmo tenant")
	}
	
	// Verificar acesso cross-tenant
	if childGroup.TenantID != authInfo.TenantID && !r.config.EnableCrossTenantAccess && !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant operation denied", 
			"requester_tenant", authInfo.TenantID,
			"resource_tenant", childGroup.TenantID)
		
		return false, errors.NewForbiddenError("cross_tenant_operation", "Operação em outro tenant não permitida")
	}
	
	// Verificar se a operação criaria um ciclo
	wouldCreateCycle, err := r.groupService.WouldCreateCycle(ctx, childID, parentID)
	if err != nil {
		r.logger.Error(ctx, "Failed to check cycle detection", "error", err.Error())
		return false, err
	}
	
	if wouldCreateCycle {
		r.logger.Warn(ctx, "Operation would create a cycle in group hierarchy",
			"child_id", childID,
			"parent_id", parentID)
		
		return false, errors.NewBusinessError("would_create_cycle", "Esta operação criaria um ciclo na hierarquia de grupos")
	}
	
	// Adicionar grupo filho ao grupo pai via serviço
	err = r.groupService.AddChildGroup(ctx, parentID, childID)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to add child group", "error", err.Error())
		
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
			EventType:   "GROUP_HIERARCHY_UPDATED",
			Severity:    "INFO",
			UserID:      authInfo.UserID,
			TenantID:    childGroup.TenantID,
			Description: fmt.Sprintf("Grupo %s adicionado como filho do grupo %s", childGroup.Name, parentGroup.Name),
			IPAddress:   auth.GetClientIPFromContext(ctx),
			Metadata: map[string]interface{}{
				"child_id":  childID,
				"parent_id": parentID,
			},
		})
	}
	
	return true, nil
}

// RemoveGroupFromParent resolve a mutation para remover um grupo filho de um grupo pai
func (r *mutationResolver) RemoveGroupFromParent(ctx context.Context, childID string, parentID string) (bool, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.removeGroupFromParent",
		trace.WithAttributes(
			attribute.String("child_group.id", childID),
			attribute.String("parent_group.id", parentID),
		))
	defer span.End()
	
	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return false, errors.ErrUnauthorized
	}
	
	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: removeGroupFromParent", 
		"child_id", childID,
		"parent_id", parentID,
		"requester_id", authInfo.UserID)
	
	// Obter o grupo filho
	childGroup, err := r.groupService.GetByID(ctx, childID)
	if err != nil {
		r.logger.Error(ctx, "Failed to get child group", "error", err.Error(), "group_id", childID)
		return false, err
	}
	
	// Obter o grupo pai
	parentGroup, err := r.groupService.GetByID(ctx, parentID)
	if err != nil {
		r.logger.Error(ctx, "Failed to get parent group", "error", err.Error(), "group_id", parentID)
		return false, err
	}
	
	// Verificar acesso cross-tenant
	if childGroup.TenantID != authInfo.TenantID && !r.config.EnableCrossTenantAccess && !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant operation denied", 
			"requester_tenant", authInfo.TenantID,
			"resource_tenant", childGroup.TenantID)
		
		return false, errors.NewForbiddenError("cross_tenant_operation", "Operação em outro tenant não permitida")
	}
	
	// Verificar se o grupo filho realmente é filho do grupo pai
	isChild, err := r.groupService.IsChildGroup(ctx, parentID, childID)
	if err != nil {
		r.logger.Error(ctx, "Failed to check child relationship", "error", err.Error())
		return false, err
	}
	
	if !isChild {
		r.logger.Warn(ctx, "Group is not a child of the parent group",
			"child_id", childID,
			"parent_id", parentID)
		
		return false, errors.NewBusinessError("not_a_child_group", "O grupo não é um filho do grupo pai especificado")
	}
	
	// Remover grupo filho do grupo pai via serviço
	err = r.groupService.RemoveChildGroup(ctx, parentID, childID)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to remove child group", "error", err.Error())
		
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
			EventType:   "GROUP_HIERARCHY_UPDATED",
			Severity:    "INFO",
			UserID:      authInfo.UserID,
			TenantID:    childGroup.TenantID,
			Description: fmt.Sprintf("Grupo %s removido como filho do grupo %s", childGroup.Name, parentGroup.Name),
			IPAddress:   auth.GetClientIPFromContext(ctx),
			Metadata: map[string]interface{}{
				"child_id":  childID,
				"parent_id": parentID,
			},
		})
	}
	
	return true, nil
}

// GroupStatistics resolve a query para obter estatísticas de um grupo
func (r *queryResolver) GroupStatistics(ctx context.Context, groupID *string) (*model.GroupStatistics, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.groupStatistics")
	defer span.End()
	
	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}
	
	// Determinar o grupo para as estatísticas (opcional)
	targetGroupID := ""
	if groupID != nil {
		targetGroupID = *groupID
		
		// Se um grupo específico foi solicitado, verificar acesso
		group, err := r.groupService.GetByID(ctx, targetGroupID)
		if err != nil {
			r.logger.Error(ctx, "Failed to get group for statistics", "error", err.Error(), "group_id", targetGroupID)
			return nil, err
		}
		
		// Verificar acesso cross-tenant
		if group.TenantID != authInfo.TenantID && !authInfo.HasPermission("IAM:CrossTenantAccess") {
			r.logger.Warn(ctx, "Cross-tenant statistics access denied", 
				"requester_tenant", authInfo.TenantID,
				"target_tenant", group.TenantID)
				
			return nil, errors.NewForbiddenError("cross_tenant_access", "Acesso a estatísticas de outro tenant não permitido")
		}
	}
	
	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: groupStatistics", 
		"group_id", targetGroupID, 
		"requester_id", authInfo.UserID)
	
	// Adicionar atributos ao span
	if targetGroupID != "" {
		span.SetAttributes(attribute.String("group.id", targetGroupID))
	}
	
	// Obter estatísticas via serviço
	stats, err := r.groupService.GetStatistics(ctx, targetGroupID)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to get group statistics", "error", err.Error(), "group_id", targetGroupID)
		
		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		
		return nil, err
	}
	
	// Registrar acesso bem-sucedido
	span.SetAttributes(attribute.Bool("success", true))
	span.SetAttributes(attribute.Int("stats.total_members", stats.TotalMembers))
	span.SetAttributes(attribute.Int("stats.active_members", stats.ActiveMembers))
	span.SetAttributes(attribute.Float64("stats.growth_rate", stats.GrowthRate))
	
	return stats, nil
}