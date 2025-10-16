/**
 * @file attribute_resolvers.go
 * @description Resolvers GraphQL para operações de consulta de atributos contextuais
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package resolvers

import (
	"context"
	"time"
	"fmt"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"

	"innovabiz/iam/src/multi-context/application/queries"
	"innovabiz/iam/src/multi-context/domain/models"
)

// AttributesQueryResolver implementa os resolvers para consultas relacionadas a atributos contextuais
type AttributesQueryResolver struct {
	listAttributesHandler    *queries.ListAttributesHandler
	searchAttributesHandler  *queries.SearchAttributesHandler
	attributeService         *services.AttributeService
	contextService           *services.ContextService
	auditLogger              services.AuditLogger
}

// NewAttributesQueryResolver cria uma nova instância do resolver para atributos
func NewAttributesQueryResolver(
	listAttributesHandler *queries.ListAttributesHandler,
	searchAttributesHandler *queries.SearchAttributesHandler,
	attributeService *services.AttributeService,
	contextService *services.ContextService,
	auditLogger services.AuditLogger,
) *AttributesQueryResolver {
	return &AttributesQueryResolver{
		listAttributesHandler:   listAttributesHandler,
		searchAttributesHandler: searchAttributesHandler,
		attributeService:        attributeService,
		contextService:          contextService,
		auditLogger:             auditLogger,
	}
}

// GraphQLAttribute representa o tipo de atributo para GraphQL
type GraphQLAttribute struct {
	ID                graphql.ID           `json:"id"`
	ContextID         graphql.ID           `json:"contextId"`
	AttributeKey      string               `json:"attributeKey"`
	AttributeValue    string               `json:"attributeValue"`
	SensitivityLevel  string               `json:"sensitivityLevel"`
	VerificationStatus string               `json:"verificationStatus"`
	VerificationSource *string              `json:"verificationSource"`
	CreatedAt         graphql.Time         `json:"createdAt"`
	UpdatedAt         graphql.Time         `json:"updatedAt"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// GraphQLAttributesResult representa o resultado paginado para listagem de atributos
type GraphQLAttributesResult struct {
	Attributes   []*GraphQLAttribute `json:"attributes"`
	TotalCount   int32               `json:"totalCount"`
	PageCount    int32               `json:"pageCount"`
	CurrentPage  int32               `json:"currentPage"`
	PageSize     int32               `json:"pageSize"`
	HasMore      bool                `json:"hasMore"`
}

// GraphQLAttributeFilters representa os filtros de atributo para GraphQL
type GraphQLAttributeFilters struct {
	ContextID          *graphql.ID    `json:"contextId"`
	KeyPattern         *string        `json:"keyPattern"`
	SensitivityLevel   *string        `json:"sensitivityLevel"`
	VerificationStatus *string        `json:"verificationStatus"`
	CreatedAfter       *graphql.Time  `json:"createdAfter"`
	CreatedBefore      *graphql.Time  `json:"createdBefore"`
}

// GraphQLAttributeSearchFilters representa os filtros avançados para busca de atributos
type GraphQLAttributeSearchFilters struct {
	ContextIDs          []graphql.ID   `json:"contextIds"`
	UserID              *string        `json:"userId"`
	TenantID            *string        `json:"tenantId"`
	SearchText          string         `json:"searchText"`
	SensitivityLevels   []string       `json:"sensitivityLevels"`
	VerificationStatuses []string       `json:"verificationStatuses"`
	ContextTypes        []string       `json:"contextTypes"`
	IncludeInactive     *bool          `json:"includeInactive"`
	CreatedAfter        *graphql.Time  `json:"createdAfter"`
	CreatedBefore       *graphql.Time  `json:"createdBefore"`
}

// Attribute resolve a consulta de atributo por ID
func (r *AttributesQueryResolver) Attribute(ctx context.Context, args struct {
	ID graphql.ID
}) (*GraphQLAttribute, error) {
	// Extrair informações de autenticação do contexto da requisição
	userInfo := extractUserInfo(ctx)
	
	// Converter o ID do formato GraphQL para UUID
	attributeID, err := uuid.Parse(string(args.ID))
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_ATTRIBUTE_QUERY_FAILED",
			ResourceID:  string(args.ID),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": "ID inválido",
			},
		})
		
		return nil, err
	}
	
	// Buscar o atributo pelo ID
	attribute, err := r.attributeService.GetAttributeByID(ctx, attributeID)
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_ATTRIBUTE_QUERY_FAILED",
			ResourceID:  string(args.ID),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		
		return nil, err
	}
	
	// Buscar o contexto relacionado para verificar permissões de acesso
	identityContext, err := r.contextService.GetContextByID(ctx, attribute.ContextID)
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_ATTRIBUTE_QUERY_FAILED",
			ResourceID:  string(args.ID),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": err.Error(),
				"context_id": attribute.ContextID.String(),
			},
		})
		
		return nil, err
	}
	
	// Verificar permissões de acesso
	if !canAccessContext(ctx, identityContext, userInfo) {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_ATTRIBUTE_ACCESS_DENIED",
			ResourceID:  string(args.ID),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"context_user_id": identityContext.UserID,
				"tenant_id":       identityContext.TenantID,
				"context_id":      identityContext.ID.String(),
			},
		})
		
		return nil, fmt.Errorf("acesso negado ao atributo")
	}
	
	// Converter para o tipo GraphQL
	result := mapAttributeToGraphQL(attribute)
	
	r.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "GRAPHQL_ATTRIBUTE_QUERY_SUCCEEDED",
		ResourceID:  string(args.ID),
		ResourceType: "CONTEXT_ATTRIBUTE",
		UserID:      userInfo.UserID,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"attribute_key": attribute.Key,
			"context_id":    attribute.ContextID.String(),
		},
	})
	
	return result, nil
}

// Attributes resolve a consulta para listar atributos com filtros e paginação
func (r *AttributesQueryResolver) Attributes(ctx context.Context, args struct {
	Filters    *GraphQLAttributeFilters
	Pagination *GraphQLPagination
	Sorting    *GraphQLSorting
}) (*GraphQLAttributesResult, error) {
	// Extrair informações de autenticação do contexto da requisição
	userInfo := extractUserInfo(ctx)
	
	// Preparar os filtros para a consulta
	queryFilters := queries.ListAttributesQuery{
		RequestedBy: userInfo.UserID,
	}
	
	// Aplicar filtros se fornecidos
	if args.Filters != nil {
		// Converter o ContextID se fornecido
		if args.Filters.ContextID != nil {
			contextID, err := uuid.Parse(string(*args.Filters.ContextID))
			if err != nil {
				r.auditLogger.LogEvent(ctx, services.AuditEvent{
					EventType:   "GRAPHQL_ATTRIBUTES_QUERY_FAILED",
					ResourceType: "CONTEXT_ATTRIBUTE",
					UserID:      userInfo.UserID,
					Timestamp:   time.Now(),
					Details: map[string]interface{}{
						"error": "ID de contexto inválido",
					},
				})
				
				return nil, err
			}
			queryFilters.ContextID = &contextID
		}
		
		queryFilters.KeyPattern = args.Filters.KeyPattern
		
		if args.Filters.SensitivityLevel != nil {
			sensLevel := mapGraphQLSensitivityLevelToModel(*args.Filters.SensitivityLevel)
			queryFilters.SensitivityLevel = &sensLevel
		}
		
		if args.Filters.VerificationStatus != nil {
			verStatus := mapGraphQLVerificationStatusToModel(*args.Filters.VerificationStatus)
			queryFilters.VerificationStatus = &verStatus
		}
		
		if args.Filters.CreatedAfter != nil {
			timeValue := time.Time(args.Filters.CreatedAfter.Time)
			queryFilters.CreatedAfter = &timeValue
		}
		
		if args.Filters.CreatedBefore != nil {
			timeValue := time.Time(args.Filters.CreatedBefore.Time)
			queryFilters.CreatedBefore = &timeValue
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
	
	// Executar a consulta
	result, err := r.listAttributesHandler.Handle(ctx, queryFilters)
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_ATTRIBUTES_QUERY_FAILED",
			ResourceType: "CONTEXT_ATTRIBUTE",
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
	graphqlResult := &GraphQLAttributesResult{
		TotalCount:  int32(result.TotalCount),
		PageCount:   int32(result.PageCount),
		CurrentPage: int32(result.CurrentPage),
		PageSize:    int32(result.PageSize),
		HasMore:     result.HasMore,
	}
	
	// Converter os atributos
	graphqlResult.Attributes = make([]*GraphQLAttribute, len(result.Attributes))
	for i, attr := range result.Attributes {
		graphqlResult.Attributes[i] = mapAttributeToGraphQL(attr)
	}
	
	r.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "GRAPHQL_ATTRIBUTES_QUERY_SUCCEEDED",
		ResourceType: "CONTEXT_ATTRIBUTE",
		UserID:      userInfo.UserID,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"result_count": len(result.Attributes),
			"total_count":  result.TotalCount,
			"page":         result.CurrentPage,
		},
	})
	
	return graphqlResult, nil
}

// SearchAttributes resolve a consulta para busca avançada de atributos
func (r *AttributesQueryResolver) SearchAttributes(ctx context.Context, args struct {
	Filters    GraphQLAttributeSearchFilters
	Pagination *GraphQLPagination
	Sorting    *GraphQLSorting
}) (*GraphQLAttributesResult, error) {
	// Extrair informações de autenticação do contexto da requisição
	userInfo := extractUserInfo(ctx)
	
	// Verificar se o usuário tem permissões administrativas para busca avançada
	if !userInfo.IsAdmin {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_SEARCH_ATTRIBUTES_ACCESS_DENIED",
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": "Permissão de administrador necessária para busca avançada",
			},
		})
		
		return nil, fmt.Errorf("acesso negado: permissão de administrador necessária para busca avançada")
	}
	
	// Preparar os filtros para a consulta
	queryFilters := queries.SearchAttributesQuery{
		RequestedBy: userInfo.UserID,
		SearchText:  args.Filters.SearchText,
	}
	
	// Converter os IDs de contexto
	if len(args.Filters.ContextIDs) > 0 {
		contextIDs := make([]uuid.UUID, len(args.Filters.ContextIDs))
		for i, id := range args.Filters.ContextIDs {
			contextID, err := uuid.Parse(string(id))
			if err != nil {
				r.auditLogger.LogEvent(ctx, services.AuditEvent{
					EventType:   "GRAPHQL_SEARCH_ATTRIBUTES_QUERY_FAILED",
					ResourceType: "CONTEXT_ATTRIBUTE",
					UserID:      userInfo.UserID,
					Timestamp:   time.Now(),
					Details: map[string]interface{}{
						"error": "ID de contexto inválido",
						"id":    string(id),
					},
				})
				
				return nil, fmt.Errorf("ID de contexto inválido: %s", id)
			}
			contextIDs[i] = contextID
		}
		queryFilters.ContextIDs = contextIDs
	}
	
	queryFilters.UserID = args.Filters.UserID
	queryFilters.TenantID = args.Filters.TenantID
	
	// Converter os níveis de sensibilidade
	if len(args.Filters.SensitivityLevels) > 0 {
		sensLevels := make([]models.SensitivityLevel, len(args.Filters.SensitivityLevels))
		for i, level := range args.Filters.SensitivityLevels {
			sensLevels[i] = mapGraphQLSensitivityLevelToModel(level)
		}
		queryFilters.SensitivityLevels = sensLevels
	}
	
	// Converter os status de verificação
	if len(args.Filters.VerificationStatuses) > 0 {
		verStatuses := make([]models.VerificationStatus, len(args.Filters.VerificationStatuses))
		for i, status := range args.Filters.VerificationStatuses {
			verStatuses[i] = mapGraphQLVerificationStatusToModel(status)
		}
		queryFilters.VerificationStatuses = verStatuses
	}
	
	queryFilters.ContextTypes = args.Filters.ContextTypes
	
	if args.Filters.IncludeInactive != nil {
		queryFilters.IncludeInactive = *args.Filters.IncludeInactive
	}
	
	if args.Filters.CreatedAfter != nil {
		timeValue := time.Time(args.Filters.CreatedAfter.Time)
		queryFilters.CreatedAfter = &timeValue
	}
	
	if args.Filters.CreatedBefore != nil {
		timeValue := time.Time(args.Filters.CreatedBefore.Time)
		queryFilters.CreatedBefore = &timeValue
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
		queryFilters.SortBy = "relevance" // Ordenação padrão para busca é por relevância
		queryFilters.SortDirection = "desc"
	}
	
	// Executar a consulta
	result, err := r.searchAttributesHandler.Handle(ctx, queryFilters)
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_SEARCH_ATTRIBUTES_QUERY_FAILED",
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error":   err.Error(),
				"search_text": args.Filters.SearchText,
			},
		})
		
		return nil, err
	}
	
	// Converter para o tipo GraphQL
	graphqlResult := &GraphQLAttributesResult{
		TotalCount:  int32(result.TotalCount),
		PageCount:   int32(result.PageCount),
		CurrentPage: int32(result.CurrentPage),
		PageSize:    int32(result.PageSize),
		HasMore:     result.HasMore,
	}
	
	// Converter os atributos
	graphqlResult.Attributes = make([]*GraphQLAttribute, len(result.Attributes))
	for i, attr := range result.Attributes {
		graphqlResult.Attributes[i] = mapAttributeToGraphQL(attr)
	}
	
	r.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "GRAPHQL_SEARCH_ATTRIBUTES_QUERY_SUCCEEDED",
		ResourceType: "CONTEXT_ATTRIBUTE",
		UserID:      userInfo.UserID,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"result_count": len(result.Attributes),
			"total_count":  result.TotalCount,
			"search_text":  args.Filters.SearchText,
		},
	})
	
	return graphqlResult, nil
}

// AttributeVerificationHistory resolve a consulta para histórico de verificações de um atributo
func (r *AttributesQueryResolver) AttributeVerificationHistory(ctx context.Context, args struct {
	AttributeID graphql.ID
}) ([]*GraphQLVerificationHistory, error) {
	// Extrair informações de autenticação do contexto da requisição
	userInfo := extractUserInfo(ctx)
	
	// Converter o ID do formato GraphQL para UUID
	attributeID, err := uuid.Parse(string(args.AttributeID))
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_ATTRIBUTE_HISTORY_QUERY_FAILED",
			ResourceID:  string(args.AttributeID),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": "ID inválido",
			},
		})
		
		return nil, err
	}
	
	// Buscar o atributo pelo ID
	attribute, err := r.attributeService.GetAttributeByID(ctx, attributeID)
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_ATTRIBUTE_HISTORY_QUERY_FAILED",
			ResourceID:  string(args.AttributeID),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		
		return nil, err
	}
	
	// Buscar o contexto relacionado para verificar permissões de acesso
	identityContext, err := r.contextService.GetContextByID(ctx, attribute.ContextID)
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_ATTRIBUTE_HISTORY_QUERY_FAILED",
			ResourceID:  string(args.AttributeID),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": err.Error(),
				"context_id": attribute.ContextID.String(),
			},
		})
		
		return nil, err
	}
	
	// Verificar permissões de acesso
	if !canAccessContext(ctx, identityContext, userInfo) {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_ATTRIBUTE_HISTORY_ACCESS_DENIED",
			ResourceID:  string(args.AttributeID),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"context_user_id": identityContext.UserID,
				"tenant_id":       identityContext.TenantID,
				"context_id":      identityContext.ID.String(),
			},
		})
		
		return nil, fmt.Errorf("acesso negado ao histórico do atributo")
	}
	
	// Buscar o histórico de verificações (a partir dos metadados do atributo)
	history := extractVerificationHistoryFromAttribute(attribute)
	
	r.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "GRAPHQL_ATTRIBUTE_HISTORY_QUERY_SUCCEEDED",
		ResourceID:  string(args.AttributeID),
		ResourceType: "CONTEXT_ATTRIBUTE",
		UserID:      userInfo.UserID,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"attribute_key": attribute.Key,
			"history_count": len(history),
		},
	})
	
	return history, nil
}

// mapAttributeToGraphQL converte um modelo de domínio para o tipo GraphQL
func mapAttributeToGraphQL(attribute *models.ContextAttribute) *GraphQLAttribute {
	result := &GraphQLAttribute{
		ID:                graphql.ID(attribute.ID.String()),
		ContextID:         graphql.ID(attribute.ContextID.String()),
		AttributeKey:      attribute.Key,
		AttributeValue:    attribute.Value,
		SensitivityLevel:  string(attribute.SensitivityLevel),
		VerificationStatus: string(attribute.VerificationStatus),
		CreatedAt:         graphql.Time{Time: attribute.CreatedAt},
		UpdatedAt:         graphql.Time{Time: attribute.UpdatedAt},
		Metadata:          attribute.Metadata,
	}
	
	if attribute.VerificationSource != "" {
		result.VerificationSource = &attribute.VerificationSource
	}
	
	return result
}

// mapGraphQLSensitivityLevelToModel converte um nível de sensibilidade de GraphQL para o modelo de domínio
func mapGraphQLSensitivityLevelToModel(level string) models.SensitivityLevel {
	switch level {
	case "LOW":
		return models.SensitivityLevelLow
	case "MEDIUM":
		return models.SensitivityLevelMedium
	case "HIGH":
		return models.SensitivityLevelHigh
	case "CRITICAL":
		return models.SensitivityLevelCritical
	default:
		return models.SensitivityLevelMedium
	}
}

// mapGraphQLVerificationStatusToModel converte um status de verificação de GraphQL para o modelo de domínio
func mapGraphQLVerificationStatusToModel(status string) models.VerificationStatus {
	switch status {
	case "UNVERIFIED":
		return models.VerificationStatusUnverified
	case "PENDING":
		return models.VerificationStatusPending
	case "VERIFIED":
		return models.VerificationStatusVerified
	case "REJECTED":
		return models.VerificationStatusRejected
	default:
		return models.VerificationStatusUnverified
	}
}

// extractVerificationHistoryFromAttribute extrai o histórico de verificações dos metadados do atributo
func extractVerificationHistoryFromAttribute(attribute *models.ContextAttribute) []*GraphQLVerificationHistory {
	var history []*GraphQLVerificationHistory
	
	if attribute.Metadata == nil {
		return history
	}
	
	// Buscar o histórico de verificações nos metadados
	rawHistory, ok := attribute.Metadata["verification_history"]
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
			if status, ok := entryMap["status"].(string); ok {
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