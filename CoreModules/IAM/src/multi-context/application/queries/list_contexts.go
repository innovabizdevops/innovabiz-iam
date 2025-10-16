/**
 * @file list_contexts.go
 * @description Consulta e handler para listagem de contextos de identidade
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	
	"innovabiz/iam/src/multi-context/domain/models"
	"innovabiz/iam/src/multi-context/domain/services"
)

// ListContextsQuery representa a consulta para listar contextos de identidade
type ListContextsQuery struct {
	UserID           *string                  // Filtrar por ID do usuário (opcional)
	TenantID         *string                  // Filtrar por ID do tenant (opcional)
	Status           *models.ContextStatus    // Filtrar por status (opcional)
	ContextType      *string                  // Filtrar por tipo de contexto (opcional)
	MinTrustScore    *float64                 // Filtrar por pontuação mínima de confiança (opcional)
	MinVerificationLevel *models.VerificationLevel // Filtrar por nível mínimo de verificação (opcional)
	Tags             []string                 // Filtrar por tags (opcional)
	CreatedAfter     *time.Time               // Filtrar por data de criação após (opcional)
	CreatedBefore    *time.Time               // Filtrar por data de criação antes (opcional)
	SortBy           string                   // Campo para ordenação (opcional, padrão: "created_at")
	SortDirection    string                   // Direção da ordenação (opcional, "asc" ou "desc", padrão: "desc")
	Page             int                      // Página a ser retornada (começando em 0)
	PageSize         int                      // Tamanho da página
	IncludeInactive  bool                     // Se deve incluir contextos inativos
	IncludeDeleted   bool                     // Se deve incluir contextos marcados como excluídos
	IncludeAttributes bool                    // Se deve incluir atributos relacionados
	RequestedBy       string                  // Utilizador ou sistema que solicitou a consulta
}

// ContextsResult representa o resultado da listagem de contextos
type ContextsResult struct {
	Contexts    []*models.IdentityContext  // Lista de contextos
	TotalCount  int                        // Número total de contextos (sem paginação)
	PageCount   int                        // Número total de páginas
	CurrentPage int                        // Página atual
	PageSize    int                        // Tamanho da página
	HasMore     bool                       // Se existem mais páginas
}

// ListContextsHandler gerencia a consulta para listar contextos de identidade
type ListContextsHandler struct {
	contextService *services.ContextService
	auditLogger    services.AuditLogger
}

// NewListContextsHandler cria uma nova instância do handler
func NewListContextsHandler(
	contextService *services.ContextService,
	auditLogger services.AuditLogger,
) *ListContextsHandler {
	return &ListContextsHandler{
		contextService: contextService,
		auditLogger:    auditLogger,
	}
}

// Handle processa a consulta de listagem de contextos
func (h *ListContextsHandler) Handle(ctx context.Context, query ListContextsQuery) (*ContextsResult, error) {
	// Registrar início da operação para rastreabilidade
	startTime := time.Now()
	operationID := uuid.New()
	
	h.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "LIST_CONTEXTS_INITIATED",
		ResourceType: "IDENTITY_CONTEXT",
		UserID:      query.RequestedBy,
		Timestamp:   startTime,
		Details: map[string]interface{}{
			"operation_id":      operationID,
			"filters":           getQueryFilters(query),
			"page":              query.Page,
			"page_size":         query.PageSize,
			"include_attributes": query.IncludeAttributes,
		},
	})
	
	// Validar parâmetros de paginação
	if query.Page < 0 {
		query.Page = 0
	}
	
	if query.PageSize <= 0 {
		query.PageSize = 20 // Padrão
	} else if query.PageSize > 100 {
		query.PageSize = 100 // Limite máximo
	}
	
	// Configurar a ordenação
	if query.SortBy == "" {
		query.SortBy = "created_at"
	}
	
	if query.SortDirection != "asc" && query.SortDirection != "desc" {
		query.SortDirection = "desc" // Padrão: mais recentes primeiro
	}
	
	// Preparar filtros
	filters := models.ContextFilters{
		UserID:            query.UserID,
		TenantID:          query.TenantID,
		Status:            query.Status,
		ContextType:       query.ContextType,
		MinTrustScore:     query.MinTrustScore,
		MinVerificationLevel: query.MinVerificationLevel,
		Tags:              query.Tags,
		CreatedAfter:      query.CreatedAfter,
		CreatedBefore:     query.CreatedBefore,
		IncludeInactive:   query.IncludeInactive,
		IncludeDeleted:    query.IncludeDeleted,
		SortBy:            query.SortBy,
		SortDirection:     query.SortDirection,
	}
	
	// Buscar o total de registros que atendem aos critérios (para paginação)
	totalCount, err := h.contextService.CountContexts(ctx, filters)
	if err != nil {
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "LIST_CONTEXTS_FAILED",
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      query.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":      operationID,
				"error":             err.Error(),
				"duration_ms":       time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, fmt.Errorf("erro ao contar contextos: %w", err)
	}
	
	// Calcular informações de paginação
	pageCount := calculatePageCount(totalCount, query.PageSize)
	
	// Ajustar página se necessário
	if query.Page >= pageCount && totalCount > 0 {
		query.Page = pageCount - 1
	}
	
	// Buscar os contextos
	offset := query.Page * query.PageSize
	contexts, err := h.contextService.ListContexts(ctx, filters, offset, query.PageSize)
	if err != nil {
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "LIST_CONTEXTS_FAILED",
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      query.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":      operationID,
				"error":             err.Error(),
				"duration_ms":       time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, fmt.Errorf("erro ao listar contextos: %w", err)
	}
	
	// Carregar atributos relacionados se solicitado
	if query.IncludeAttributes && len(contexts) > 0 {
		for _, context := range contexts {
			if err := h.contextService.LoadContextAttributes(ctx, context); err != nil {
				// Apenas logar o erro, mas continuar com o processamento
				h.auditLogger.LogEvent(ctx, services.AuditEvent{
					EventType:   "LOAD_CONTEXT_ATTRIBUTES_FAILED",
					ResourceID:  context.ID.String(),
					ResourceType: "IDENTITY_CONTEXT",
					UserID:      query.RequestedBy,
					Timestamp:   time.Now(),
					Details: map[string]interface{}{
						"operation_id":  operationID,
						"context_id":    context.ID,
						"error":         err.Error(),
					},
				})
			}
		}
	}
	
	// Preparar resultado
	result := &ContextsResult{
		Contexts:    contexts,
		TotalCount:  totalCount,
		PageCount:   pageCount,
		CurrentPage: query.Page,
		PageSize:    query.PageSize,
		HasMore:     (query.Page+1) < pageCount,
	}
	
	// Registrar sucesso da operação
	h.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "LIST_CONTEXTS_SUCCEEDED",
		ResourceType: "IDENTITY_CONTEXT",
		UserID:      query.RequestedBy,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"operation_id":      operationID,
			"result_count":      len(contexts),
			"total_count":       totalCount,
			"page":              query.Page,
			"page_size":         query.PageSize,
			"page_count":        pageCount,
			"duration_ms":       time.Since(startTime).Milliseconds(),
		},
	})
	
	return result, nil
}

// calculatePageCount calcula o número total de páginas com base no total de registros e tamanho da página
func calculatePageCount(totalCount, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	
	if totalCount == 0 {
		return 1
	}
	
	return (totalCount + pageSize - 1) / pageSize
}

// getQueryFilters retorna um mapa com os filtros aplicados para auditoria
func getQueryFilters(query ListContextsQuery) map[string]interface{} {
	filters := make(map[string]interface{})
	
	if query.UserID != nil {
		filters["user_id"] = *query.UserID
	}
	
	if query.TenantID != nil {
		filters["tenant_id"] = *query.TenantID
	}
	
	if query.Status != nil {
		filters["status"] = *query.Status
	}
	
	if query.ContextType != nil {
		filters["context_type"] = *query.ContextType
	}
	
	if query.MinTrustScore != nil {
		filters["min_trust_score"] = *query.MinTrustScore
	}
	
	if query.MinVerificationLevel != nil {
		filters["min_verification_level"] = *query.MinVerificationLevel
	}
	
	if len(query.Tags) > 0 {
		filters["tags"] = query.Tags
	}
	
	if query.CreatedAfter != nil {
		filters["created_after"] = query.CreatedAfter.Format(time.RFC3339)
	}
	
	if query.CreatedBefore != nil {
		filters["created_before"] = query.CreatedBefore.Format(time.RFC3339)
	}
	
	filters["sort_by"] = query.SortBy
	filters["sort_direction"] = query.SortDirection
	filters["include_inactive"] = query.IncludeInactive
	filters["include_deleted"] = query.IncludeDeleted
	
	return filters
}