package resolvers

import (
	"context"
	"time"

	"github.com/innovabiz/iam/internal/domain/model"
	"github.com/innovabiz/iam/internal/domain/model/errors"
	"github.com/innovabiz/iam/internal/infrastructure/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// SecurityEvent resolve a query para buscar um evento de segurança por ID
func (r *queryResolver) SecurityEvent(ctx context.Context, id string) (*model.SecurityEvent, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.securityEvent",
		trace.WithAttributes(attribute.String("security_event.id", id)))
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verificar permissão para acessar eventos de segurança
	if !authInfo.HasPermission("IAM:ViewSecurityEvents") {
		r.logger.Warn(ctx, "Permission denied for viewing security event",
			"requester_id", authInfo.UserID,
			"event_id", id)
		return nil, errors.NewForbiddenError("security_event_access_denied", "Permissão insuficiente para visualizar eventos de segurança")
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: securityEvent",
		"event_id", id,
		"requester_id", authInfo.UserID,
		"tenant_id", authInfo.TenantID)

	// Obter evento de segurança do serviço
	event, err := r.securityService.GetEventByID(ctx, id)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to get security event by ID",
			"error", err.Error(),
			"event_id", id)

		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))

		return nil, err
	}

	// Verificar acesso cross-tenant
	if event.TenantID != authInfo.TenantID && !r.config.EnableCrossTenantAccess && !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant access denied",
			"requester_tenant", authInfo.TenantID,
			"resource_tenant", event.TenantID)

		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", "cross_tenant_access_denied"))

		return nil, errors.NewForbiddenError("cross_tenant_access", "Acesso a recursos de outro tenant não permitido")
	}

	// Registrar acesso bem-sucedido
	span.SetAttributes(attribute.Bool("success", true))

	return event, nil
}

// SecurityEvents resolve a query para listar eventos de segurança com filtros e paginação
func (r *queryResolver) SecurityEvents(ctx context.Context, filter *model.SecurityEventFilter, pagination *model.PaginationInput) (*model.SecurityEventConnection, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.securityEvents")
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verificar permissão para acessar eventos de segurança
	if !authInfo.HasPermission("IAM:ViewSecurityEvents") {
		r.logger.Warn(ctx, "Permission denied for viewing security events",
			"requester_id", authInfo.UserID)
		return nil, errors.NewForbiddenError("security_event_access_denied", "Permissão insuficiente para visualizar eventos de segurança")
	}

	// Aplicar configurações de paginação padrão se não fornecidas
	if pagination == nil {
		pagination = &model.PaginationInput{
			Page:          0,
			PerPage:       r.config.DefaultPageSize,
			SortField:     "timestamp",
			SortDirection: model.SortDirectionDESC, // Default é ordem cronológica reversa
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
		filter = &model.SecurityEventFilter{}
	}

	// Forçar filtro de tenant para garantir isolamento, exceto para administradores com permissão especial
	if !authInfo.HasPermission("IAM:CrossTenantAccess") {
		filter.TenantID = authInfo.TenantID
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: securityEvents",
		"filter", filter,
		"pagination", pagination,
		"requester_id", authInfo.UserID)

	// Adicionar atributos ao span
	span.SetAttributes(attribute.Int("pagination.page", pagination.Page))
	span.SetAttributes(attribute.Int("pagination.per_page", pagination.PerPage))
	if filter != nil && filter.TenantID != "" {
		span.SetAttributes(attribute.String("filter.tenant_id", filter.TenantID))
	}

	// Obter eventos de segurança do serviço
	result, err := r.securityService.ListEvents(ctx, filter, pagination)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to list security events", "error", err.Error())

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

// SecurityEventSubscription handle a subscription para eventos de segurança em tempo real
func (r *subscriptionResolver) SecurityEventSubscription(ctx context.Context, filter *model.SecurityEventFilter) (<-chan *model.SecurityEvent, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.subscription.securityEventSubscription")
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verificar permissão para receber eventos de segurança
	if !authInfo.HasPermission("IAM:ReceiveSecurityEvents") {
		r.logger.Warn(ctx, "Permission denied for security event subscription",
			"requester_id", authInfo.UserID)
		return nil, errors.NewForbiddenError("security_event_subscription_denied", "Permissão insuficiente para assinar eventos de segurança")
	}

	// Aplicar filtro se não fornecido
	if filter == nil {
		filter = &model.SecurityEventFilter{}
	}

	// Forçar filtro de tenant para garantir isolamento, exceto para administradores com permissão especial
	if !authInfo.HasPermission("IAM:CrossTenantAccess") {
		filter.TenantID = authInfo.TenantID
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL subscription: securityEventSubscription",
		"filter", filter,
		"requester_id", authInfo.UserID,
		"tenant_id", authInfo.TenantID)

	// Adicionar atributos ao span
	if filter != nil && filter.TenantID != "" {
		span.SetAttributes(attribute.String("filter.tenant_id", filter.TenantID))
	}
	if filter != nil && filter.Severity != nil {
		span.SetAttributes(attribute.String("filter.severity", *filter.Severity))
	}

	// Criar canal para eventos
	events := make(chan *model.SecurityEvent, 100)

	// Registrar observador para receber eventos filtrados
	unsubscribe := r.securityService.SubscribeToEvents(ctx, events, filter)

	// Monitorar contexto para fechar a assinatura quando o cliente desconectar
	go func() {
		<-ctx.Done()
		r.logger.Info(ctx, "Security event subscription ended",
			"requester_id", authInfo.UserID)
		unsubscribe()
		close(events)
	}()

	// Registrar assinatura bem-sucedida
	span.SetAttributes(attribute.Bool("success", true))

	return events, nil
}

// CreateSecurityEvent resolve a mutation para criar um evento de segurança manualmente
// Geralmente usado por administradores para fins de auditoria ou por serviços externos
func (r *mutationResolver) CreateSecurityEvent(ctx context.Context, input model.CreateSecurityEventInput) (*model.SecurityEvent, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.mutation.createSecurityEvent")
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verificar permissão para criar eventos de segurança
	if !authInfo.HasPermission("IAM:CreateSecurityEvents") {
		r.logger.Warn(ctx, "Permission denied for creating security event", 
			"requester_id", authInfo.UserID)
		return nil, errors.NewForbiddenError("security_event_creation_denied", "Permissão insuficiente para criar eventos de segurança")
	}

	// Determinar tenant ID se não fornecido
	if input.TenantID == "" {
		input.TenantID = authInfo.TenantID
	}

	// Verificar acesso cross-tenant
	if input.TenantID != authInfo.TenantID && !r.config.EnableCrossTenantAccess && !authInfo.HasPermission("IAM:CrossTenantAccess") {
		r.logger.Warn(ctx, "Cross-tenant operation denied", 
			"requester_tenant", authInfo.TenantID,
			"target_tenant", input.TenantID)
			
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", "cross_tenant_operation_denied"))
		
		return nil, errors.NewForbiddenError("cross_tenant_operation", "Operação em outro tenant não permitida")
	}

	// Adicionar timestamp atual se não fornecido
	if input.Timestamp == nil {
		now := time.Now()
		input.Timestamp = &now
	}

	// Adicionar IP do cliente se não fornecido
	if input.IPAddress == "" {
		input.IPAddress = auth.GetClientIPFromContext(ctx)
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL mutation: createSecurityEvent", 
		"input", map[string]interface{}{
			"eventType": input.EventType,
			"severity": input.Severity,
			"description": input.Description,
			"tenantId": input.TenantID,
		},
		"requester_id", authInfo.UserID)

	// Adicionar atributos ao span
	span.SetAttributes(attribute.String("event.type", input.EventType))
	span.SetAttributes(attribute.String("event.severity", input.Severity))
	span.SetAttributes(attribute.String("tenant.id", input.TenantID))

	// Criar evento via serviço
	event := &model.SecurityEvent{
		EventType:   input.EventType,
		Severity:    input.Severity,
		UserID:      input.UserID,
		TenantID:    input.TenantID,
		Description: input.Description,
		IPAddress:   input.IPAddress,
		Timestamp:   input.Timestamp,
		Metadata:    input.Metadata,
	}

	createdEvent, err := r.securityService.LogEvent(ctx, event)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to create security event", "error", err.Error())
		
		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		
		return nil, err
	}

	// Registrar operação bem-sucedida
	span.SetAttributes(attribute.Bool("success", true))
	span.SetAttributes(attribute.String("event.id", createdEvent.ID))

	return createdEvent, nil
}

// SecurityEventStatistics resolve a query para obter estatísticas de eventos de segurança
func (r *queryResolver) SecurityEventStatistics(ctx context.Context, filter *model.SecurityEventFilter, timeRange *model.TimeRangeInput) (*model.SecurityEventStatistics, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.securityEventStatistics")
	defer span.End()
	
	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}
	
	// Verificar permissão para acessar estatísticas de eventos de segurança
	if !authInfo.HasPermission("IAM:ViewSecurityStatistics") {
		r.logger.Warn(ctx, "Permission denied for viewing security event statistics", 
			"requester_id", authInfo.UserID)
		return nil, errors.NewForbiddenError("statistics_access_denied", "Permissão insuficiente para visualizar estatísticas de segurança")
	}
	
	// Aplicar filtro se não fornecido
	if filter == nil {
		filter = &model.SecurityEventFilter{}
	}
	
	// Forçar filtro de tenant para garantir isolamento, exceto para administradores com permissão especial
	if !authInfo.HasPermission("IAM:CrossTenantAccess") {
		filter.TenantID = authInfo.TenantID
	}
	
	// Definir intervalo de tempo padrão se não fornecido (últimas 24 horas)
	if timeRange == nil {
		now := time.Now()
		startDate := now.Add(-24 * time.Hour)
		timeRange = &model.TimeRangeInput{
			StartDate: &startDate,
			EndDate:   &now,
		}
	}
	
	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: securityEventStatistics", 
		"filter", filter,
		"timeRange", map[string]interface{}{
			"startDate": timeRange.StartDate,
			"endDate":   timeRange.EndDate,
		},
		"requester_id", authInfo.UserID)
	
	// Adicionar atributos ao span
	if filter != nil && filter.TenantID != "" {
		span.SetAttributes(attribute.String("filter.tenant_id", filter.TenantID))
	}
	if timeRange.StartDate != nil && timeRange.EndDate != nil {
		span.SetAttributes(attribute.String("time_range.start", timeRange.StartDate.Format(time.RFC3339)))
		span.SetAttributes(attribute.String("time_range.end", timeRange.EndDate.Format(time.RFC3339)))
	}
	
	// Obter estatísticas via serviço
	stats, err := r.securityService.GetStatistics(ctx, filter, timeRange)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to get security event statistics", "error", err.Error())
		
		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))
		
		return nil, err
	}
	
	// Registrar acesso bem-sucedido
	span.SetAttributes(attribute.Bool("success", true))
	span.SetAttributes(attribute.Int("stats.total_events", stats.TotalEvents))
	
	return stats, nil
}