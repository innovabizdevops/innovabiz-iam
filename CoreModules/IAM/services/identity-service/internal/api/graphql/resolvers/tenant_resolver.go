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

// Tenant resolve a query para buscar um tenant por ID
func (r *queryResolver) Tenant(ctx context.Context, id string) (*model.Tenant, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.tenant",
		trace.WithAttributes(attribute.String("tenant.id", id)))
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verificar se usuário pode acessar informações de tenants
	if !authInfo.HasPermission("IAM:ViewTenants") && id != authInfo.TenantID {
		r.logger.Warn(ctx, "Permission denied for viewing tenant details",
			"requester_id", authInfo.UserID,
			"requester_tenant", authInfo.TenantID,
			"target_tenant", id)
		return nil, errors.NewForbiddenError("tenant_access_denied", "Sem permissão para visualizar detalhes do tenant")
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: tenant",
		"tenant_id", id,
		"requester_id", authInfo.UserID,
		"requester_tenant", authInfo.TenantID)

	// Obter tenant do serviço
	tenant, err := r.tenantService.GetByID(ctx, id)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to get tenant by ID",
			"error", err.Error(),
			"tenant_id", id)

		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))

		return nil, err
	}

	// Verificar acesso cross-tenant se não for administrador global
	if tenant.ID != authInfo.TenantID && !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant access denied",
			"requester_tenant", authInfo.TenantID,
			"resource_tenant", tenant.ID)

		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", "cross_tenant_access_denied"))

		return nil, errors.NewForbiddenError("cross_tenant_access", "Acesso a recursos de outro tenant não permitido")
	}

	// Registrar acesso bem-sucedido
	span.SetAttributes(attribute.Bool("success", true))

	return tenant, nil
}

// Tenants resolve a query para listar tenants com filtros e paginação
func (r *queryResolver) Tenants(ctx context.Context, filter *model.TenantFilter, pagination *model.PaginationInput) (*model.TenantConnection, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.tenants")
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verificar permissão para listar tenants
	if !authInfo.HasPermission("IAM:ViewTenants") {
		r.logger.Warn(ctx, "Permission denied for listing tenants",
			"requester_id", authInfo.UserID)
		return nil, errors.NewForbiddenError("tenant_access_denied", "Sem permissão para listar tenants")
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
		filter = &model.TenantFilter{}
	}

	// Restringir tenants visíveis se não tiver acesso cross-tenant
	if !authInfo.HasPermission("IAM:CrossTenantAccess") {
		// Se não for admin global, só pode ver seu próprio tenant
		filter.ID = authInfo.TenantID
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: tenants",
		"filter", filter,
		"pagination", pagination,
		"requester_id", authInfo.UserID,
		"requester_tenant", authInfo.TenantID)

	// Adicionar atributos ao span
	span.SetAttributes(attribute.Int("pagination.page", pagination.Page))
	span.SetAttributes(attribute.Int("pagination.per_page", pagination.PerPage))

	// Obter tenants do serviço
	result, err := r.tenantService.ListTenants(ctx, filter, pagination)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to list tenants", "error", err.Error())

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

// TenantByCode resolve a query para buscar um tenant por código
func (r *queryResolver) TenantByCode(ctx context.Context, code string) (*model.Tenant, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.tenantByCode",
		trace.WithAttributes(attribute.String("tenant.code", code)))
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verificar se usuário pode acessar informações de tenants
	if !authInfo.HasPermission("IAM:ViewTenants") {
		r.logger.Warn(ctx, "Permission denied for viewing tenant details",
			"requester_id", authInfo.UserID)
		return nil, errors.NewForbiddenError("tenant_access_denied", "Sem permissão para visualizar detalhes do tenant")
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: tenantByCode",
		"tenant_code", code,
		"requester_id", authInfo.UserID,
		"requester_tenant", authInfo.TenantID)

	// Obter tenant do serviço
	tenant, err := r.tenantService.GetByCode(ctx, code)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to get tenant by code",
			"error", err.Error(),
			"tenant_code", code)

		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))

		return nil, err
	}

	// Verificar acesso cross-tenant se não for administrador global
	if tenant.ID != authInfo.TenantID && !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant access denied",
			"requester_tenant", authInfo.TenantID,
			"resource_tenant", tenant.ID)

		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", "cross_tenant_access_denied"))

		return nil, errors.NewForbiddenError("cross_tenant_access", "Acesso a recursos de outro tenant não permitido")
	}

	// Registrar acesso bem-sucedido
	span.SetAttributes(attribute.Bool("success", true))

	return tenant, nil
}

// CreateTenant resolve a mutation para criar um novo tenant
func (r *mutationResolver) CreateTenant(ctx context.Context, input model.CreateTenantInput) (*model.Tenant, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.createTenant")
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verificar permissão para criar tenants
	if !authInfo.HasPermission("IAM:ManageTenants") {
		r.logger.Warn(ctx, "Permission denied for creating tenant", 
			"requester_id", authInfo.UserID)
		return nil, errors.NewForbiddenError("tenant_management_denied", "Permissão insuficiente para criar tenants")
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: createTenant", 
		"input", map[string]interface{}{
			"name": input.Name,
			"code": input.Code,
			"domain": input.Domain,
			"region": input.Region,
		},
		"requester_id", authInfo.UserID,
		"requester_tenant", authInfo.TenantID)

	// Adicionar atributos ao span
	span.SetAttributes(attribute.String("tenant.name", input.Name))
	span.SetAttributes(attribute.String("tenant.code", input.Code))
	span.SetAttributes(attribute.String("tenant.domain", input.Domain))
	span.SetAttributes(attribute.String("tenant.region", input.Region))

	// Criar tenant via serviço
	tenant, err := r.tenantService.Create(ctx, &input)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to create tenant", "error", err.Error())
		
		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		
		return nil, err
	}

	// Registrar operação bem-sucedida
	span.SetAttributes(attribute.Bool("success", true))
	span.SetAttributes(attribute.String("tenant.id", tenant.ID))

	// Publicar evento de segurança
	if r.config.EnableAuditLogging {
		r.securityService.LogEvent(ctx, &model.SecurityEvent{
			EventType:   "TENANT_CREATED",
			Severity:    "INFO",
			UserID:      authInfo.UserID,
			TenantID:    authInfo.TenantID, // Evento é registrado no contexto do usuário que fez a operação
			Description: fmt.Sprintf("Tenant %s criado", input.Name),
			IPAddress:   auth.GetClientIPFromContext(ctx),
			Metadata: map[string]interface{}{
				"tenant_name": input.Name,
				"tenant_code": input.Code,
				"tenant_domain": input.Domain,
				"tenant_region": input.Region,
			},
		})
	}

	return tenant, nil
}

// UpdateTenant resolve a mutation para atualizar um tenant existente
func (r *mutationResolver) UpdateTenant(ctx context.Context, id string, input model.UpdateTenantInput) (*model.Tenant, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.updateTenant", 
		trace.WithAttributes(attribute.String("tenant.id", id)))
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verificar permissão para atualizar tenants
	if !authInfo.HasPermission("IAM:ManageTenants") && id != authInfo.TenantID {
		r.logger.Warn(ctx, "Permission denied for updating tenant", 
			"requester_id", authInfo.UserID,
			"target_tenant", id)
		return nil, errors.NewForbiddenError("tenant_management_denied", "Permissão insuficiente para atualizar tenants")
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: updateTenant", 
		"tenant_id", id, 
		"input", input,
		"requester_id", authInfo.UserID,
		"requester_tenant", authInfo.TenantID)

	// Obter o tenant atual para verificar restrições adicionais
	existingTenant, err := r.tenantService.GetByID(ctx, id)
	if err != nil {
		r.logger.Error(ctx, "Failed to get tenant for update", "error", err.Error(), "tenant_id", id)
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, err
	}

	// Verificar regras especiais para tenant do sistema
	if existingTenant.IsSystem && !authInfo.HasPermission("IAM:ManageSystemTenants") {
		r.logger.Warn(ctx, "Permission denied for updating system tenant", 
			"requester_id", authInfo.UserID,
			"tenant_id", id)
		return nil, errors.NewForbiddenError("system_tenant_update", "Não possui permissão para modificar tenants de sistema")
	}

	// Atualizar tenant via serviço
	tenant, err := r.tenantService.Update(ctx, id, &input)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to update tenant", "error", err.Error(), "tenant_id", id)
		
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
			EventType:   "TENANT_UPDATED",
			Severity:    "INFO",
			UserID:      authInfo.UserID,
			TenantID:    authInfo.TenantID,
			Description: fmt.Sprintf("Tenant %s atualizado", existingTenant.Name),
			IPAddress:   auth.GetClientIPFromContext(ctx),
			Metadata: map[string]interface{}{
				"tenant_id": id,
				"changes": input,
			},
		})
	}

	return tenant, nil
}

// DeactivateTenant resolve a mutation para desativar um tenant
func (r *mutationResolver) DeactivateTenant(ctx context.Context, id string, reason string) (bool, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.deactivateTenant", 
		trace.WithAttributes(attribute.String("tenant.id", id)))
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return false, errors.ErrUnauthorized
	}

	// Verificar permissão para gerenciar tenants
	if !authInfo.HasPermission("IAM:ManageTenants") {
		r.logger.Warn(ctx, "Permission denied for deactivating tenant", 
			"requester_id", authInfo.UserID,
			"target_tenant", id)
		return false, errors.NewForbiddenError("tenant_management_denied", "Permissão insuficiente para desativar tenants")
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: deactivateTenant", 
		"tenant_id", id, 
		"reason", reason,
		"requester_id", authInfo.UserID)

	// Obter o tenant atual para verificação
	existingTenant, err := r.tenantService.GetByID(ctx, id)
	if err != nil {
		r.logger.Error(ctx, "Failed to get tenant for deactivation", "error", err.Error(), "tenant_id", id)
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		return false, err
	}

	// Não permitir desativar tenant do sistema
	if existingTenant.IsSystem {
		r.logger.Warn(ctx, "Attempt to deactivate system tenant", 
			"requester_id", authInfo.UserID,
			"tenant_id", id)
		return false, errors.NewBusinessError("system_tenant_protection", "Não é possível desativar tenants de sistema")
	}

	// Desativar tenant via serviço
	err = r.tenantService.Deactivate(ctx, id, reason)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to deactivate tenant", "error", err.Error(), "tenant_id", id)
		
		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		
		return false, err
	}

	// Registrar operação bem-sucedida
	span.SetAttributes(attribute.Bool("success", true))

	// Publicar evento de segurança (alta severidade)
	if r.config.EnableAuditLogging {
		r.securityService.LogEvent(ctx, &model.SecurityEvent{
			EventType:   "TENANT_DEACTIVATED",
			Severity:    "WARN",
			UserID:      authInfo.UserID,
			TenantID:    authInfo.TenantID,
			Description: fmt.Sprintf("Tenant %s desativado", existingTenant.Name),
			IPAddress:   auth.GetClientIPFromContext(ctx),
			Metadata: map[string]interface{}{
				"tenant_id": id,
				"tenant_name": existingTenant.Name,
				"reason": reason,
			},
		})
	}

	return true, nil
}

// ReactivateTenant resolve a mutation para reativar um tenant
func (r *mutationResolver) ReactivateTenant(ctx context.Context, id string) (bool, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.reactivateTenant", 
		trace.WithAttributes(attribute.String("tenant.id", id)))
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return false, errors.ErrUnauthorized
	}

	// Verificar permissão para gerenciar tenants
	if !authInfo.HasPermission("IAM:ManageTenants") {
		r.logger.Warn(ctx, "Permission denied for reactivating tenant", 
			"requester_id", authInfo.UserID,
			"target_tenant", id)
		return false, errors.NewForbiddenError("tenant_management_denied", "Permissão insuficiente para reativar tenants")
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: reactivateTenant", 
		"tenant_id", id, 
		"requester_id", authInfo.UserID)

	// Reativar tenant via serviço
	err := r.tenantService.Reactivate(ctx, id)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to reactivate tenant", "error", err.Error(), "tenant_id", id)
		
		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		
		return false, err
	}

	// Registrar operação bem-sucedida
	span.SetAttributes(attribute.Bool("success", true))

	// Obter informações do tenant para o evento
	tenant, _ := r.tenantService.GetByID(ctx, id)
	tenantName := "unknown"
	if tenant != nil {
		tenantName = tenant.Name
	}

	// Publicar evento de segurança
	if r.config.EnableAuditLogging {
		r.securityService.LogEvent(ctx, &model.SecurityEvent{
			EventType:   "TENANT_REACTIVATED",
			Severity:    "INFO",
			UserID:      authInfo.UserID,
			TenantID:    authInfo.TenantID,
			Description: fmt.Sprintf("Tenant %s reativado", tenantName),
			IPAddress:   auth.GetClientIPFromContext(ctx),
			Metadata: map[string]interface{}{
				"tenant_id": id,
			},
		})
	}

	return true, nil
}

// TenantStatistics resolve a query para obter estatísticas de um tenant
func (r *queryResolver) TenantStatistics(ctx context.Context, id string) (*model.TenantStatistics, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.tenantStatistics",
		trace.WithAttributes(attribute.String("tenant.id", id)))
	defer span.End()
	
	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}
	
	// Verificar permissão para visualizar estatísticas de tenants
	if id != authInfo.TenantID && !authInfo.HasPermission("IAM:ViewTenantStatistics") {
		r.logger.Warn(ctx, "Permission denied for viewing tenant statistics", 
			"requester_id", authInfo.UserID,
			"target_tenant", id)
		return nil, errors.NewForbiddenError("tenant_statistics_access_denied", "Permissão insuficiente para visualizar estatísticas do tenant")
	}
	
	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: tenantStatistics", 
		"tenant_id", id, 
		"requester_id", authInfo.UserID)
	
	// Obter estatísticas via serviço
	stats, err := r.tenantService.GetStatistics(ctx, id)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to get tenant statistics", "error", err.Error(), "tenant_id", id)
		
		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		
		return nil, err
	}
	
	// Registrar acesso bem-sucedido
	span.SetAttributes(attribute.Bool("success", true))
	
	return stats, nil
}