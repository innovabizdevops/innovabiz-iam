/**
 * @file context_resolvers.go
 * @description Resolvers GraphQL para operações de consulta de contextos de identidade
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package resolvers

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"
	
	"innovabiz/iam/src/multi-context/application/queries"
	"innovabiz/iam/src/multi-context/domain/models"
)

// ContextsQueryResolver implementa os resolvers para consultas relacionadas a contextos de identidade
type ContextsQueryResolver struct {
	listContextsHandler      *queries.ListContextsHandler
	contextService           *services.ContextService
	auditLogger              services.AuditLogger
}

// NewContextsQueryResolver cria uma nova instância do resolver para contextos
func NewContextsQueryResolver(
	listContextsHandler *queries.ListContextsHandler,
	contextService *services.ContextService,
	auditLogger services.AuditLogger,
) *ContextsQueryResolver {
	return &ContextsQueryResolver{
		listContextsHandler: listContextsHandler,
		contextService:      contextService,
		auditLogger:         auditLogger,
	}
}

// GraphQLContext representa o tipo de contexto para GraphQL
type GraphQLContext struct {
	ID               graphql.ID           `json:"id"`
	UserID           string               `json:"userId"`
	TenantID         string               `json:"tenantId"`
	ContextType      string               `json:"contextType"`
	DisplayName      string               `json:"displayName"`
	Description      *string              `json:"description"`
	Status           string               `json:"status"`
	VerificationLevel string               `json:"verificationLevel"`
	TrustScore       float64              `json:"trustScore"`
	Tags             []string             `json:"tags"`
	CreatedAt        graphql.Time         `json:"createdAt"`
	UpdatedAt        graphql.Time         `json:"updatedAt"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// GraphQLContextsResult representa o resultado paginado para listagem de contextos
type GraphQLContextsResult struct {
	Contexts    []*GraphQLContext `json:"contexts"`
	TotalCount  int32             `json:"totalCount"`
	PageCount   int32             `json:"pageCount"`
	CurrentPage int32             `json:"currentPage"`
	PageSize    int32             `json:"pageSize"`
	HasMore     bool              `json:"hasMore"`
}

// GraphQLVerificationHistory representa o histórico de verificação para GraphQL
type GraphQLVerificationHistory struct {
	Timestamp   graphql.Time         `json:"timestamp"`
	Status      string               `json:"status"`
	Source      *string              `json:"source"`
	Notes       *string              `json:"notes"`
	RequestedBy *string              `json:"requestedBy"`
	Evidence    map[string]interface{} `json:"evidence"`
}

// GraphQLContextFilters representa os filtros de contexto para GraphQL
type GraphQLContextFilters struct {
	UserID               *string    `json:"userId"`
	TenantID             *string    `json:"tenantId"`
	Status               *string    `json:"status"`
	ContextType          *string    `json:"contextType"`
	MinTrustScore        *float64   `json:"minTrustScore"`
	MinVerificationLevel *string    `json:"minVerificationLevel"`
	Tags                 []string   `json:"tags"`
	CreatedAfter         *graphql.Time `json:"createdAfter"`
	CreatedBefore        *graphql.Time `json:"createdBefore"`
	IncludeInactive      *bool      `json:"includeInactive"`
	IncludeDeleted       *bool      `json:"includeDeleted"`
}

// GraphQLPagination representa a entrada de paginação para GraphQL
type GraphQLPagination struct {
	Page     *int32 `json:"page"`
	PageSize *int32 `json:"pageSize"`
}

// GraphQLSorting representa a entrada de ordenação para GraphQL
type GraphQLSorting struct {
	Field     *string `json:"field"`
	Direction *string `json:"direction"`
}

// Context resolve a consulta de contexto por ID
func (r *ContextsQueryResolver) Context(ctx context.Context, args struct {
	ID graphql.ID
}) (*GraphQLContext, error) {
	// Extrair informações de autenticação do contexto da requisição
	userInfo := extractUserInfo(ctx)
	
	// Converter o ID do formato GraphQL para UUID
	contextID, err := uuid.Parse(string(args.ID))
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_CONTEXT_QUERY_FAILED",
			ResourceID:  string(args.ID),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": "ID inválido",
			},
		})
		
		return nil, err
	}
	
	// Buscar o contexto pelo ID
	identityContext, err := r.contextService.GetContextByID(ctx, contextID)
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_CONTEXT_QUERY_FAILED",
			ResourceID:  string(args.ID),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		
		return nil, err
	}
	
	// Verificar permissões de acesso
	if !canAccessContext(ctx, identityContext, userInfo) {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_CONTEXT_ACCESS_DENIED",
			ResourceID:  string(args.ID),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"context_user_id": identityContext.UserID,
				"tenant_id":       identityContext.TenantID,
			},
		})
		
		return nil, fmt.Errorf("acesso negado ao contexto")
	}
	
	// Carregar atributos relacionados
	if err := r.contextService.LoadContextAttributes(ctx, identityContext); err != nil {
		// Apenas logar o erro, mas continuar com o processamento
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "LOAD_CONTEXT_ATTRIBUTES_FAILED",
			ResourceID:  contextID.String(),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
	}
	
	// Converter para o tipo GraphQL
	result := mapIdentityContextToGraphQL(identityContext)
	
	r.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "GRAPHQL_CONTEXT_QUERY_SUCCEEDED",
		ResourceID:  string(args.ID),
		ResourceType: "IDENTITY_CONTEXT",
		UserID:      userInfo.UserID,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"context_user_id": identityContext.UserID,
			"context_type":    identityContext.ContextType,
		},
	})
	
	return result, nil
}

// Contexts resolve a consulta para listar contextos com filtros e paginação
func (r *ContextsQueryResolver) Contexts(ctx context.Context, args struct {
	Filters          *GraphQLContextFilters
	Pagination       *GraphQLPagination
	Sorting          *GraphQLSorting
	IncludeAttributes *bool
}) (*GraphQLContextsResult, error) {
	// Extrair informações de autenticação do contexto da requisição
	userInfo := extractUserInfo(ctx)
	
	// Preparar os filtros para a consulta
	queryFilters := queries.ListContextsQuery{
		RequestedBy: userInfo.UserID,
	}
	
	// Aplicar filtros se fornecidos
	if args.Filters != nil {
		queryFilters.UserID = args.Filters.UserID
		queryFilters.TenantID = args.Filters.TenantID
		
		if args.Filters.Status != nil {
			status := mapGraphQLStatusToModel(*args.Filters.Status)
			queryFilters.Status = &status
		}
		
		queryFilters.ContextType = args.Filters.ContextType
		queryFilters.MinTrustScore = args.Filters.MinTrustScore
		
		if args.Filters.MinVerificationLevel != nil {
			verLevel := mapGraphQLVerificationLevelToModel(*args.Filters.MinVerificationLevel)
			queryFilters.MinVerificationLevel = &verLevel
		}
		
		queryFilters.Tags = args.Filters.Tags
		
		if args.Filters.CreatedAfter != nil {
			timeValue := time.Time(args.Filters.CreatedAfter.Time)
			queryFilters.CreatedAfter = &timeValue
		}
		
		if args.Filters.CreatedBefore != nil {
			timeValue := time.Time(args.Filters.CreatedBefore.Time)
			queryFilters.CreatedBefore = &timeValue
		}
		
		if args.Filters.IncludeInactive != nil {
			queryFilters.IncludeInactive = *args.Filters.IncludeInactive
		}
		
		if args.Filters.IncludeDeleted != nil {
			queryFilters.IncludeDeleted = *args.Filters.IncludeDeleted
		}
	}
	
	// Aplicar paginação
	if args.Pagination != nil {
		if args.Pagination.Page != nil {
			queryFilters.Page = int(*args.Pagination.Page)
		}
		
		if args.Pagination.PageSize != nil {
			queryFilters.PageSize = int(*args.Pagination.PageSize)
		}
	} else {
		queryFilters.Page = 0
		queryFilters.PageSize = 20
	}
	
	// Aplicar ordenação
	if args.Sorting != nil {
		if args.Sorting.Field != nil {
			queryFilters.SortBy = *args.Sorting.Field
		}
		
		if args.Sorting.Direction != nil {
			queryFilters.SortDirection = *args.Sorting.Direction
		}
	} else {
		queryFilters.SortBy = "created_at"
		queryFilters.SortDirection = "desc"
	}
	
	// Aplicar inclusão de atributos
	if args.IncludeAttributes != nil {
		queryFilters.IncludeAttributes = *args.IncludeAttributes
	}
	
	// Executar a consulta
	result, err := r.listContextsHandler.Handle(ctx, queryFilters)
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_CONTEXTS_QUERY_FAILED",
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error":   err.Error(),
				"filters": args.Filters,
			},
		})
		
		return nil, err
	}
	
	// Converter para o tipo GraphQL
	graphqlResult := &GraphQLContextsResult{
		TotalCount:  int32(result.TotalCount),
		PageCount:   int32(result.PageCount),
		CurrentPage: int32(result.CurrentPage),
		PageSize:    int32(result.PageSize),
		HasMore:     result.HasMore,
	}
	
	// Converter os contextos
	graphqlResult.Contexts = make([]*GraphQLContext, len(result.Contexts))
	for i, ctx := range result.Contexts {
		graphqlResult.Contexts[i] = mapIdentityContextToGraphQL(ctx)
	}
	
	r.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "GRAPHQL_CONTEXTS_QUERY_SUCCEEDED",
		ResourceType: "IDENTITY_CONTEXT",
		UserID:      userInfo.UserID,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"result_count": len(result.Contexts),
			"total_count":  result.TotalCount,
			"page":         result.CurrentPage,
		},
	})
	
	return graphqlResult, nil
}

// ContextVerificationHistory resolve a consulta para histórico de verificações de um contexto
func (r *ContextsQueryResolver) ContextVerificationHistory(ctx context.Context, args struct {
	ContextID graphql.ID
}) ([]*GraphQLVerificationHistory, error) {
	// Extrair informações de autenticação do contexto da requisição
	userInfo := extractUserInfo(ctx)
	
	// Converter o ID do formato GraphQL para UUID
	contextID, err := uuid.Parse(string(args.ContextID))
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_CONTEXT_HISTORY_QUERY_FAILED",
			ResourceID:  string(args.ContextID),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": "ID inválido",
			},
		})
		
		return nil, err
	}
	
	// Buscar o contexto pelo ID para verificar permissões de acesso
	identityContext, err := r.contextService.GetContextByID(ctx, contextID)
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_CONTEXT_HISTORY_QUERY_FAILED",
			ResourceID:  string(args.ContextID),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		
		return nil, err
	}
	
	// Verificar permissões de acesso
	if !canAccessContext(ctx, identityContext, userInfo) {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_CONTEXT_HISTORY_ACCESS_DENIED",
			ResourceID:  string(args.ContextID),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"context_user_id": identityContext.UserID,
				"tenant_id":       identityContext.TenantID,
			},
		})
		
		return nil, fmt.Errorf("acesso negado ao histórico do contexto")
	}
	
	// Buscar o histórico de verificações (a partir dos metadados)
	history := extractVerificationHistoryFromContext(identityContext)
	
	r.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "GRAPHQL_CONTEXT_HISTORY_QUERY_SUCCEEDED",
		ResourceID:  string(args.ContextID),
		ResourceType: "IDENTITY_CONTEXT",
		UserID:      userInfo.UserID,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"history_count": len(history),
		},
	})
	
	return history, nil
}

// mapIdentityContextToGraphQL converte um modelo de domínio para o tipo GraphQL
func mapIdentityContextToGraphQL(context *models.IdentityContext) *GraphQLContext {
	result := &GraphQLContext{
		ID:               graphql.ID(context.ID.String()),
		UserID:           context.UserID,
		TenantID:         context.TenantID,
		ContextType:      context.ContextType,
		DisplayName:      context.DisplayName,
		Status:           string(context.Status),
		VerificationLevel: string(context.VerificationLevel),
		TrustScore:       context.TrustScore,
		Tags:             context.Tags,
		CreatedAt:        graphql.Time{Time: context.CreatedAt},
		UpdatedAt:        graphql.Time{Time: context.UpdatedAt},
		Metadata:         context.Metadata,
	}
	
	if context.Description != "" {
		result.Description = &context.Description
	}
	
	return result
}

// mapGraphQLStatusToModel converte um status de GraphQL para o modelo de domínio
func mapGraphQLStatusToModel(status string) models.ContextStatus {
	switch status {
	case "ACTIVE":
		return models.ContextStatusActive
	case "SUSPENDED":
		return models.ContextStatusSuspended
	case "LOCKED":
		return models.ContextStatusLocked
	case "INACTIVE":
		return models.ContextStatusInactive
	case "DELETED":
		return models.ContextStatusDeleted
	default:
		return models.ContextStatusActive
	}
}

// mapGraphQLVerificationLevelToModel converte um nível de verificação de GraphQL para o modelo de domínio
func mapGraphQLVerificationLevelToModel(level string) models.VerificationLevel {
	switch level {
	case "UNVERIFIED":
		return models.VerificationLevelUnverified
	case "BASIC":
		return models.VerificationLevelBasic
	case "STANDARD":
		return models.VerificationLevelStandard
	case "ADVANCED":
		return models.VerificationLevelAdvanced
	case "HIGH":
		return models.VerificationLevelHigh
	default:
		return models.VerificationLevelUnverified
	}
}

// canAccessContext verifica se o usuário tem permissão para acessar o contexto
func canAccessContext(ctx context.Context, identityContext *models.IdentityContext, userInfo UserInfo) bool {
	// Se o usuário é administrador, tem acesso a todos os contextos
	if userInfo.IsAdmin {
		return true
	}
	
	// Se o contexto pertence ao usuário, ele tem acesso
	if identityContext.UserID == userInfo.UserID {
		return true
	}
	
	// Se o usuário pertence ao mesmo tenant e tem permissão para acessar contextos do tenant
	if identityContext.TenantID == userInfo.TenantID && userInfo.HasTenantAccess {
		return true
	}
	
	return false
}

// extractVerificationHistoryFromContext extrai o histórico de verificações dos metadados do contexto
func extractVerificationHistoryFromContext(identityContext *models.IdentityContext) []*GraphQLVerificationHistory {
	var history []*GraphQLVerificationHistory
	
	if identityContext.Metadata == nil {
		return history
	}
	
	// Buscar o histórico de verificações nos metadados
	rawHistory, ok := identityContext.Metadata["verification_history"]
	if !ok {
		return history
	}
	
	historyEntries, ok := rawHistory.([]interface{})
	if !ok {
		return history
	}
	
	for _, entry := range historyEntries {
		if entryMap, ok := entry.(map[string]interface{}); ok {
			historyEntry := &GraphQLVerificationHistory{}
			
			// Extrair o timestamp
			if timestamp, ok := entryMap["timestamp"].(time.Time); ok {
				historyEntry.Timestamp = graphql.Time{Time: timestamp}
			} else if timestampStr, ok := entryMap["timestamp"].(string); ok {
				if timestamp, err := time.Parse(time.RFC3339, timestampStr); err == nil {
					historyEntry.Timestamp = graphql.Time{Time: timestamp}
				}
			}
			
			// Extrair o status
			if status, ok := entryMap["verification_level"].(string); ok {
				historyEntry.Status = status
			}
			
			// Extrair a fonte
			if source, ok := entryMap["source"].(string); ok {
				historyEntry.Source = &source
			}
			
			// Extrair as notas
			if notes, ok := entryMap["notes"].(string); ok {
				historyEntry.Notes = &notes
			}
			
			// Extrair o solicitante
			if requestedBy, ok := entryMap["requested_by"].(string); ok {
				historyEntry.RequestedBy = &requestedBy
			}
			
			// Extrair evidências
			if evidence, ok := entryMap["evidence"].(map[string]interface{}); ok {
				historyEntry.Evidence = evidence
			}
			
			history = append(history, historyEntry)
		}
	}
	
	return history
}

// UserInfo representa informações do usuário autenticado
type UserInfo struct {
	UserID          string
	TenantID        string
	IsAdmin         bool
	HasTenantAccess bool
}

// extractUserInfo extrai informações do usuário do contexto da requisição
func extractUserInfo(ctx context.Context) UserInfo {
	// Em um sistema real, isso seria implementado para extrair informações
	// do token JWT ou outra fonte de autenticação
	
	// Implementação temporária para demonstração
	userID := ctx.Value("user_id")
	tenantID := ctx.Value("tenant_id")
	isAdmin := ctx.Value("is_admin")
	hasTenantAccess := ctx.Value("has_tenant_access")
	
	return UserInfo{
		UserID:          stringValue(userID, "anonymous"),
		TenantID:        stringValue(tenantID, ""),
		IsAdmin:         boolValue(isAdmin, false),
		HasTenantAccess: boolValue(hasTenantAccess, false),
	}
}

// stringValue extrai um valor string de uma interface com valor padrão
func stringValue(value interface{}, defaultValue string) string {
	if value == nil {
		return defaultValue
	}
	
	if strValue, ok := value.(string); ok {
		return strValue
	}
	
	return defaultValue
}

// boolValue extrai um valor booleano de uma interface com valor padrão
func boolValue(value interface{}, defaultValue bool) bool {
	if value == nil {
		return defaultValue
	}
	
	if boolValue, ok := value.(bool); ok {
		return boolValue
	}
	
	return defaultValue
}