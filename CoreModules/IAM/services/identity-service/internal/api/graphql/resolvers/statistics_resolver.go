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

// SystemStatistics resolve a query para obter estatísticas globais do sistema IAM
func (r *queryResolver) SystemStatistics(ctx context.Context, filter *model.StatisticsFilter) (*model.SystemStatistics, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.systemStatistics")
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verificar permissão para visualizar estatísticas do sistema
	if !authInfo.HasPermission("IAM:ViewSystemStatistics") {
		r.logger.Warn(ctx, "Permission denied for viewing system statistics",
			"requester_id", authInfo.UserID)
		return nil, errors.NewForbiddenError("system_statistics_access_denied", "Permissão insuficiente para visualizar estatísticas do sistema")
	}

	// Aplicar filtro se não fornecido
	if filter == nil {
		filter = &model.StatisticsFilter{}
	}

	// Restringir acesso a dados de outros tenants a menos que tenha permissão específica
	if !authInfo.HasPermission("IAM:CrossTenantAccess") && filter.IncludeAllTenants != nil && *filter.IncludeAllTenants {
		r.logger.Warn(ctx, "Cross-tenant statistics access denied",
			"requester_id", authInfo.UserID)
		return nil, errors.NewForbiddenError("cross_tenant_statistics_denied", "Sem permissão para acessar estatísticas de todos os tenants")
	}

	// Ajustar filtro para usuário sem acesso cross-tenant
	if !authInfo.HasPermission("IAM:CrossTenantAccess") {
		// Forçar inclusão apenas do tenant atual
		includeAllTenants := false
		filter.IncludeAllTenants = &includeAllTenants
		filter.TenantID = &authInfo.TenantID
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: systemStatistics",
		"filter", filter,
		"requester_id", authInfo.UserID,
		"tenant_id", authInfo.TenantID)

	// Adicionar atributos ao span
	if filter.TenantID != nil {
		span.SetAttributes(attribute.String("filter.tenant_id", *filter.TenantID))
	}
	if filter.IncludeAllTenants != nil {
		span.SetAttributes(attribute.Bool("filter.include_all_tenants", *filter.IncludeAllTenants))
	}

	// Obter estatísticas via serviço agregado
	stats, err := r.statisticsService.GetSystemStatistics(ctx, filter)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to get system statistics", "error", err.Error())

		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))

		return nil, err
	}

	// Registrar acesso bem-sucedido
	span.SetAttributes(attribute.Bool("success", true))

	return stats, nil
}

// UserActivityStatistics resolve a query para estatísticas de atividade de usuários
func (r *queryResolver) UserActivityStatistics(ctx context.Context, timeRange *model.TimeRangeInput, filter *model.StatisticsFilter) (*model.UserActivityStatistics, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.userActivityStatistics")
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verificar permissão para visualizar estatísticas de atividade
	if !authInfo.HasPermission("IAM:ViewUserActivityStatistics") {
		r.logger.Warn(ctx, "Permission denied for viewing user activity statistics",
			"requester_id", authInfo.UserID)
		return nil, errors.NewForbiddenError("user_activity_statistics_denied", "Permissão insuficiente para visualizar estatísticas de atividade")
	}

	// Aplicar filtro se não fornecido
	if filter == nil {
		filter = &model.StatisticsFilter{}
	}

	// Definir intervalo de tempo padrão se não fornecido (últimos 30 dias)
	if timeRange == nil {
		now := time.Now()
		startDate := now.AddDate(0, 0, -30) // 30 dias atrás
		timeRange = &model.TimeRangeInput{
			StartDate: &startDate,
			EndDate:   &now,
		}
	}

	// Restringir acesso a dados de outros tenants
	if !authInfo.HasPermission("IAM:CrossTenantAccess") {
		if filter.IncludeAllTenants != nil && *filter.IncludeAllTenants {
			r.logger.Warn(ctx, "Cross-tenant activity statistics access denied",
				"requester_id", authInfo.UserID)
			return nil, errors.NewForbiddenError("cross_tenant_statistics_denied", "Sem permissão para acessar estatísticas de todos os tenants")
		}

		// Forçar inclusão apenas do tenant atual
		includeAllTenants := false
		filter.IncludeAllTenants = &includeAllTenants
		filter.TenantID = &authInfo.TenantID
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: userActivityStatistics",
		"filter", filter,
		"timeRange", map[string]interface{}{
			"startDate": timeRange.StartDate,
			"endDate":   timeRange.EndDate,
		},
		"requester_id", authInfo.UserID)

	// Adicionar atributos ao span
	span.SetAttributes(attribute.String("time_range.start", timeRange.StartDate.Format(time.RFC3339)))
	span.SetAttributes(attribute.String("time_range.end", timeRange.EndDate.Format(time.RFC3339)))
	if filter.TenantID != nil {
		span.SetAttributes(attribute.String("filter.tenant_id", *filter.TenantID))
	}

	// Obter estatísticas via serviço
	stats, err := r.statisticsService.GetUserActivityStatistics(ctx, timeRange, filter)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to get user activity statistics", "error", err.Error())

		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))

		return nil, err
	}

	// Registrar acesso bem-sucedido
	span.SetAttributes(attribute.Bool("success", true))

	return stats, nil
}

// SecurityStatistics resolve a query para obter estatísticas de segurança do sistema
func (r *queryResolver) SecurityStatistics(ctx context.Context, timeRange *model.TimeRangeInput, filter *model.StatisticsFilter) (*model.SecurityStatistics, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.securityStatistics")
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verificar permissão para visualizar estatísticas de segurança
	if !authInfo.HasPermission("IAM:ViewSecurityStatistics") {
		r.logger.Warn(ctx, "Permission denied for viewing security statistics",
			"requester_id", authInfo.UserID)
		return nil, errors.NewForbiddenError("security_statistics_denied", "Permissão insuficiente para visualizar estatísticas de segurança")
	}

	// Aplicar filtro se não fornecido
	if filter == nil {
		filter = &model.StatisticsFilter{}
	}

	// Definir intervalo de tempo padrão se não fornecido (últimos 30 dias)
	if timeRange == nil {
		now := time.Now()
		startDate := now.AddDate(0, 0, -30) // 30 dias atrás
		timeRange = &model.TimeRangeInput{
			StartDate: &startDate,
			EndDate:   &now,
		}
	}

	// Restringir acesso a dados de outros tenants
	if !authInfo.HasPermission("IAM:CrossTenantAccess") {
		if filter.IncludeAllTenants != nil && *filter.IncludeAllTenants {
			r.logger.Warn(ctx, "Cross-tenant security statistics access denied",
				"requester_id", authInfo.UserID)
			return nil, errors.NewForbiddenError("cross_tenant_statistics_denied", "Sem permissão para acessar estatísticas de todos os tenants")
		}

		// Forçar inclusão apenas do tenant atual
		includeAllTenants := false
		filter.IncludeAllTenants = &includeAllTenants
		filter.TenantID = &authInfo.TenantID
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: securityStatistics",
		"filter", filter,
		"timeRange", map[string]interface{}{
			"startDate": timeRange.StartDate,
			"endDate":   timeRange.EndDate,
		},
		"requester_id", authInfo.UserID)

	// Adicionar atributos ao span
	span.SetAttributes(attribute.String("time_range.start", timeRange.StartDate.Format(time.RFC3339)))
	span.SetAttributes(attribute.String("time_range.end", timeRange.EndDate.Format(time.RFC3339)))
	if filter.TenantID != nil {
		span.SetAttributes(attribute.String("filter.tenant_id", *filter.TenantID))
	}

	// Obter estatísticas via serviço
	stats, err := r.statisticsService.GetSecurityStatistics(ctx, timeRange, filter)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to get security statistics", "error", err.Error())

		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))

		return nil, err
	}

	// Registrar acesso bem-sucedido
	span.SetAttributes(attribute.Bool("success", true))

	return stats, nil
}

// IAMDashboard resolve uma query para obter dados consolidados para o dashboard do IAM
func (r *queryResolver) IAMDashboard(ctx context.Context, timeRange *model.TimeRangeInput, filter *model.StatisticsFilter) (*model.IAMDashboard, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.iamDashboard")
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verificar permissão para acessar o dashboard
	if !authInfo.HasPermission("IAM:ViewDashboard") {
		r.logger.Warn(ctx, "Permission denied for accessing IAM dashboard",
			"requester_id", authInfo.UserID)
		return nil, errors.NewForbiddenError("dashboard_access_denied", "Permissão insuficiente para acessar dashboard IAM")
	}

	// Aplicar filtro se não fornecido
	if filter == nil {
		filter = &model.StatisticsFilter{}
	}

	// Definir intervalo de tempo padrão se não fornecido (últimos 7 dias)
	if timeRange == nil {
		now := time.Now()
		startDate := now.AddDate(0, 0, -7) // 7 dias atrás
		timeRange = &model.TimeRangeInput{
			StartDate: &startDate,
			EndDate:   &now,
		}
	}

	// Restringir acesso a dados de outros tenants
	if !authInfo.HasPermission("IAM:CrossTenantAccess") {
		if filter.IncludeAllTenants != nil && *filter.IncludeAllTenants {
			r.logger.Warn(ctx, "Cross-tenant dashboard access denied",
				"requester_id", authInfo.UserID)
			return nil, errors.NewForbiddenError("cross_tenant_access_denied", "Sem permissão para acessar dashboard de todos os tenants")
		}

		// Forçar inclusão apenas do tenant atual
		includeAllTenants := false
		filter.IncludeAllTenants = &includeAllTenants
		filter.TenantID = &authInfo.TenantID
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: iamDashboard",
		"filter", filter,
		"timeRange", map[string]interface{}{
			"startDate": timeRange.StartDate,
			"endDate":   timeRange.EndDate,
		},
		"requester_id", authInfo.UserID)

	// Adicionar atributos ao span
	span.SetAttributes(attribute.String("time_range.start", timeRange.StartDate.Format(time.RFC3339)))
	span.SetAttributes(attribute.String("time_range.end", timeRange.EndDate.Format(time.RFC3339)))
	if filter.TenantID != nil {
		span.SetAttributes(attribute.String("filter.tenant_id", *filter.TenantID))
	}

	// Obter dados do dashboard via serviço agregado
	dashboard, err := r.statisticsService.GetIAMDashboard(ctx, timeRange, filter)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to get IAM dashboard data", "error", err.Error())

		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))

		return nil, err
	}

	// Registrar acesso bem-sucedido
	span.SetAttributes(attribute.Bool("success", true))

	return dashboard, nil
}

// AuditLogStatistics resolve a query para estatísticas de logs de auditoria
func (r *queryResolver) AuditLogStatistics(ctx context.Context, timeRange *model.TimeRangeInput, filter *model.StatisticsFilter) (*model.AuditLogStatistics, error) {
	// Iniciar span para observabilidade
	ctx, span := r.tracer.Start(ctx, "resolvers.query.auditLogStatistics")
	defer span.End()

	// Obter contexto de autenticação
	authInfo := auth.GetAuthInfoFromContext(ctx)
	if authInfo == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verificar permissão para visualizar logs de auditoria
	if !authInfo.HasPermission("IAM:ViewAuditLogs") {
		r.logger.Warn(ctx, "Permission denied for viewing audit log statistics",
			"requester_id", authInfo.UserID)
		return nil, errors.NewForbiddenError("audit_statistics_denied", "Permissão insuficiente para visualizar estatísticas de auditoria")
	}

	// Aplicar filtro se não fornecido
	if filter == nil {
		filter = &model.StatisticsFilter{}
	}

	// Definir intervalo de tempo padrão se não fornecido (últimos 30 dias)
	if timeRange == nil {
		now := time.Now()
		startDate := now.AddDate(0, 0, -30) // 30 dias atrás
		timeRange = &model.TimeRangeInput{
			StartDate: &startDate,
			EndDate:   &now,
		}
	}

	// Restringir acesso a dados de outros tenants
	if !authInfo.HasPermission("IAM:CrossTenantAccess") {
		if filter.IncludeAllTenants != nil && *filter.IncludeAllTenants {
			r.logger.Warn(ctx, "Cross-tenant audit statistics access denied",
				"requester_id", authInfo.UserID)
			return nil, errors.NewForbiddenError("cross_tenant_audit_denied", "Sem permissão para acessar estatísticas de auditoria de todos os tenants")
		}

		// Forçar inclusão apenas do tenant atual
		includeAllTenants := false
		filter.IncludeAllTenants = &includeAllTenants
		filter.TenantID = &authInfo.TenantID
	}

	// Logging para auditoria
	r.logger.Info(ctx, "GraphQL query: auditLogStatistics",
		"filter", filter,
		"timeRange", map[string]interface{}{
			"startDate": timeRange.StartDate,
			"endDate":   timeRange.EndDate,
		},
		"requester_id", authInfo.UserID)

	// Adicionar atributos ao span
	span.SetAttributes(attribute.String("time_range.start", timeRange.StartDate.Format(time.RFC3339)))
	span.SetAttributes(attribute.String("time_range.end", timeRange.EndDate.Format(time.RFC3339)))
	if filter.TenantID != nil {
		span.SetAttributes(attribute.String("filter.tenant_id", *filter.TenantID))
	}

	// Obter estatísticas via serviço
	stats, err := r.statisticsService.GetAuditLogStatistics(ctx, timeRange, filter)
	if err != nil {
		// Logging de erro
		r.logger.Error(ctx, "Failed to get audit log statistics", "error", err.Error())

		// Adicionar detalhes ao span
		span.SetAttributes(attribute.Bool("success", false))
		span.SetAttributes(attribute.String("error", err.Error()))

		return nil, err
	}

	// Registrar acesso bem-sucedido
	span.SetAttributes(attribute.Bool("success", true))

	return stats, nil
}