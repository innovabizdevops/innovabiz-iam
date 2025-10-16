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

// UserResolver contém os resolvers relacionados a usuários
type userResolver struct {
	*Resolver
}

// User resolve a query para buscar um usuário por ID
func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.user", 
		trace.WithAttributes(attribute.String("user.id", id)))
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: user", 
		"user_id", id, 
		"requester_id", authInfo.UserID,
		"tenant_id", authInfo.TenantID)

	// Obter usuário do serviço
	user, err := r.userService.GetByID(ctx, id)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to get user by ID", 
			"error", err.Error(), 
			"user_id", id)
		
		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		
		return nil, err
	}

	// Verificar se o usuário tem acesso cross-tenant
	if user.TenantID != authInfo.TenantID && !r.config.EnableCrossTenantAccess && !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant access denied", 
			"requester_tenant", authInfo.TenantID,
			"resource_tenant", user.TenantID)
			
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", "cross_tenant_access_denied"))
		
		return nil, errors.NewForbiddenError("cross_tenant_access", "Acesso a recursos de outro tenant não permitido")
	}

	// Registrar acesso bem-sucedido
	span.SetAttributes(attribute.Bool("success", true))
	
	return user, nil
}

// Users resolve a query para listar usuários com filtros e paginação
func (r *queryResolver) Users(ctx context.Context, filter *model.UserFilter, pagination *model.PaginationInput) (*model.UserConnection, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.users")
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
		filter = &model.UserFilter{}
	}
	
	// Forçar filtro de tenant para garantir isolamento, exceto para administradores com permissão especial
	if !authInfo.HasPermission("IAM:CrossTenantAccess") {
		filter.TenantID = authInfo.TenantID
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: users", 
		"filter", filter,
		"pagination", pagination, 
		"requester_id", authInfo.UserID)
	
	// Adicionar atributos ao span
	span.SetAttributes(attribute.Int("pagination.page", pagination.Page))
	span.SetAttributes(attribute.Int("pagination.per_page", pagination.PerPage))
	if filter != nil && filter.TenantID != "" {
		span.SetAttributes(attribute.String("filter.tenant_id", filter.TenantID))
	}

	// Obter usuários do serviço
	result, err := r.userService.ListUsers(ctx, filter, pagination)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to list users", "error", err.Error())
		
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

// Me resolve a query para obter o usuário atualmente autenticado
func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.me")
	defer span.End()
	
	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}
	
	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: me", 
		"user_id", authInfo.UserID,
		"tenant_id", authInfo.TenantID)
	
	// Adicionar atributos ao span
	span.SetAttributes(attribute.String("user.id", authInfo.UserID))
	span.SetAttributes(attribute.String("tenant.id", authInfo.TenantID))

	// Obter usuário atual do serviço
	user, err := r.userService.GetByID(ctx, authInfo.UserID)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to get current user", "error", err.Error())
		
		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		
		return nil, err
	}
	
	// Registrar acesso bem-sucedido
	span.SetAttributes(attribute.Bool("success", true))
	
	return user, nil
}

// CreateUser resolve a mutation para criar um novo usuário
func (r *mutationResolver) CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.createUser")
	defer span.End()
	
	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}
	
	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: createUser", 
		"input", map[string]interface{}{
			"username": input.Username,
			"email": input.Email,
			"firstName": input.FirstName,
			"lastName": input.LastName,
			"status": input.Status,
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
	span.SetAttributes(attribute.String("user.username", input.Username))
	span.SetAttributes(attribute.String("user.email", input.Email))
	span.SetAttributes(attribute.String("tenant.id", input.TenantID))

	// Criar usuário via serviço
	user, err := r.userService.Create(ctx, &input)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to create user", "error", err.Error())
		
		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		
		return nil, err
	}
	
	// Registrar operação bem-sucedida
	span.SetAttributes(attribute.Bool("success", true))
	span.SetAttributes(attribute.String("user.id", user.ID))
	
	// Publicar evento de segurança
	if r.config.EnableAuditLogging {
		r.securityService.LogEvent(ctx, &model.SecurityEvent{
			EventType:   "USER_CREATED",
			Severity:    "INFO",
			UserID:      authInfo.UserID,
			TenantID:    input.TenantID,
			Description: fmt.Sprintf("Usuário %s criado", input.Username),
			IPAddress:   auth.GetClientIPFromContext(ctx),
			Metadata: map[string]interface{}{
				"username": input.Username,
				"email":    input.Email,
				"status":   input.Status,
			},
		})
	}
	
	return user, nil
}

// UpdateUser resolve a mutation para atualizar um usuário existente
func (r *mutationResolver) UpdateUser(ctx context.Context, id string, input model.UpdateUserInput) (*model.User, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.updateUser", 
		trace.WithAttributes(attribute.String("user.id", id)))
	defer span.End()
	
	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}
	
	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: updateUser", 
		"user_id", id, 
		"input", input,
		"requester_id", authInfo.UserID)
	
	// Verificar se é o próprio usuário ou tem permissão adequada
	isSelf := id == authInfo.UserID
	
	// Obter o usuário atual para verificar o tenant
	existingUser, err := r.userService.GetByID(ctx, id)
	if err != nil {
		r.logger.Error(ctx, "Failed to get user for update", "error", err.Error(), "user_id", id)
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, err
	}
	
	// Verificar acesso cross-tenant
	if existingUser.TenantID != authInfo.TenantID && !r.config.EnableCrossTenantAccess && !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant operation denied", 
			"requester_tenant", authInfo.TenantID,
			"resource_tenant", existingUser.TenantID)
			
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", "cross_tenant_operation_denied"))
		
		return nil, errors.NewForbiddenError("cross_tenant_operation", "Operação em outro tenant não permitida")
	}
	
	// Verificar permissões especiais para certas atualizações
	if input.Status != nil && !isSelf && !authInfo.HasPermission("IAM:UpdateUserStatus") {
		r.logger.Warn(ctx, "Permission denied to update user status", 
			"requester_id", authInfo.UserID,
			"target_user", id)
			
		return nil, errors.NewForbiddenError("update_status_denied", "Permissão para alterar status de usuário negada")
	}
	
	// Atualizar usuário via serviço
	user, err := r.userService.Update(ctx, id, &input)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to update user", "error", err.Error(), "user_id", id)
		
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
			EventType:   "USER_UPDATED",
			Severity:    "INFO",
			UserID:      authInfo.UserID,
			TenantID:    existingUser.TenantID,
			Description: fmt.Sprintf("Usuário %s atualizado", existingUser.Username),
			IPAddress:   auth.GetClientIPFromContext(ctx),
			Metadata: map[string]interface{}{
				"user_id": id,
				"changes": input,
			},
		})
	}
	
	return user, nil
}

// DeleteUser resolve a mutation para excluir um usuário
func (r *mutationResolver) DeleteUser(ctx context.Context, id string) (bool, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.deleteUser", 
		trace.WithAttributes(attribute.String("user.id", id)))
	defer span.End()
	
	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return false, errors.ErrUnauthorized
	}
	
	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: deleteUser", 
		"user_id", id, 
		"requester_id", authInfo.UserID)
	
	// Obter o usuário atual para verificar o tenant
	existingUser, err := r.userService.GetByID(ctx, id)
	if err != nil {
		r.logger.Error(ctx, "Failed to get user for deletion", "error", err.Error(), "user_id", id)
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		return false, err
	}
	
	// Verificar acesso cross-tenant
	if existingUser.TenantID != authInfo.TenantID && !r.config.EnableCrossTenantAccess && !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant operation denied", 
			"requester_tenant", authInfo.TenantID,
			"resource_tenant", existingUser.TenantID)
			
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", "cross_tenant_operation_denied"))
		
		return false, errors.NewForbiddenError("cross_tenant_operation", "Operação em outro tenant não permitida")
	}
	
	// Excluir usuário via serviço
	err = r.userService.Delete(ctx, id)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to delete user", "error", err.Error(), "user_id", id)
		
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
			EventType:   "USER_DELETED",
			Severity:    "WARN",
			UserID:      authInfo.UserID,
			TenantID:    existingUser.TenantID,
			Description: fmt.Sprintf("Usuário %s excluído", existingUser.Username),
			IPAddress:   auth.GetClientIPFromContext(ctx),
			Metadata: map[string]interface{}{
				"user_id":   id,
				"username":  existingUser.Username,
				"email":     existingUser.Email,
			},
		})
	}
	
	return true, nil
}

// UserStatistics resolve a query para obter estatísticas de usuários
func (r *queryResolver) UserStatistics(ctx context.Context, tenantID *string) (*model.UserStatistics, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.userStatistics")
	defer span.End()
	
	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}
	
	// Determinar o tenant para as estatísticas
	targetTenantID := authInfo.TenantID
	if tenantID != nil && *tenantID != "" {
		// Se um tenant específico foi solicitado
		if *tenantID != authInfo.TenantID && !authInfo.HasPermission("IAM:CrossTenantAccess") {
			r.logger.Warn(ctx, "Cross-tenant statistics access denied", 
				"requester_tenant", authInfo.TenantID,
				"target_tenant", *tenantID)
				
			return nil, errors.NewForbiddenError("cross_tenant_access", "Acesso a estatísticas de outro tenant não permitido")
		}
		targetTenantID = *tenantID
	}
	
	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: userStatistics", 
		"tenant_id", targetTenantID, 
		"requester_id", authInfo.UserID)
	
	// Adicionar atributos ao span
	span.SetAttributes(attribute.String("tenant.id", targetTenantID))
	
	// Obter estatísticas via serviço
	stats, err := r.userService.GetStatistics(ctx, targetTenantID)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to get user statistics", "error", err.Error(), "tenant_id", targetTenantID)
		
		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		
		return nil, err
	}
	
	// Registrar acesso bem-sucedido
	span.SetAttributes(attribute.Bool("success", true))
	span.SetAttributes(attribute.Int("stats.total_users", stats.TotalUsers))
	span.SetAttributes(attribute.Int("stats.active_users", stats.ActiveUsers))
	
	return stats, nil
}