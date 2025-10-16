/**
 * @file search_attributes.go
 * @description Consulta e handler para busca avançada de atributos contextuais
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

// SearchAttributesQuery representa a consulta para busca avançada de atributos
type SearchAttributesQuery struct {
	ContextIDs        []uuid.UUID                    // Filtrar por IDs de contexto (opcional)
	UserID           *string                         // Filtrar por ID de usuário associado (opcional)
	TenantID         *string                         // Filtrar por ID de tenant (opcional)
	SearchText       string                          // Texto para busca (chave ou valor do atributo)
	SensitivityLevels []models.SensitivityLevel      // Níveis de sensibilidade a incluir (opcional)
	VerificationStatuses []models.VerificationStatus // Status de verificação a incluir (opcional)
	ContextTypes      []string                       // Tipos de contexto a incluir (opcional)
	IncludeInactive   bool                           // Se deve incluir contextos inativos
	CreatedAfter      *time.Time                     // Filtrar por data de criação após (opcional)
	CreatedBefore     *time.Time                     // Filtrar por data de criação antes (opcional)
	Page              int                            // Página a ser retornada (começando em 0)
	PageSize          int                            // Tamanho da página
	RequestedBy       string                         // Utilizador ou sistema que solicitou a consulta
}

// SearchAttributesHandler gerencia a consulta para busca avançada de atributos
type SearchAttributesHandler struct {
	attributeService *services.AttributeService
	contextService   *services.ContextService
	auditLogger      services.AuditLogger
}

// NewSearchAttributesHandler cria uma nova instância do handler
func NewSearchAttributesHandler(
	attributeService *services.AttributeService,
	contextService *services.ContextService,
	auditLogger services.AuditLogger,
) *SearchAttributesHandler {
	return &SearchAttributesHandler{
		attributeService: attributeService,
		contextService:   contextService,
		auditLogger:      auditLogger,
	}
}

// Handle processa a consulta de busca de atributos
func (h *SearchAttributesHandler) Handle(ctx context.Context, query SearchAttributesQuery) (*AttributesResult, error) {
	// Registrar início da operação para rastreabilidade
	startTime := time.Now()
	operationID := uuid.New()
	
	h.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "SEARCH_ATTRIBUTES_INITIATED",
		ResourceType: "CONTEXT_ATTRIBUTE",
		UserID:      query.RequestedBy,
		Timestamp:   startTime,
		Details: map[string]interface{}{
			"operation_id":      operationID,
			"search_text":       query.SearchText,
			"filters":           getSearchQueryFilters(query),
			"page":              query.Page,
			"page_size":         query.PageSize,
		},
	})
	
	// Validar parâmetros de busca
	if query.SearchText == "" {
		err := fmt.Errorf("texto de busca não pode estar vazio")
		
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "SEARCH_ATTRIBUTES_FAILED",
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      query.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":  operationID,
				"error":         err.Error(),
				"duration_ms":   time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, err
	}
	
	// Validar parâmetros de paginação
	if query.Page < 0 {
		query.Page = 0
	}
	
	if query.PageSize <= 0 {
		query.PageSize = 20 // Padrão
	} else if query.PageSize > 100 {
		query.PageSize = 100 // Limite máximo
	}
	
	// Preparar filtros de busca
	searchFilters := models.AttributeSearchFilters{
		ContextIDs:         query.ContextIDs,
		UserID:             query.UserID,
		TenantID:           query.TenantID,
		SearchText:         query.SearchText,
		SensitivityLevels:  query.SensitivityLevels,
		VerificationStatuses: query.VerificationStatuses,
		ContextTypes:       query.ContextTypes,
		IncludeInactive:    query.IncludeInactive,
		CreatedAfter:       query.CreatedAfter,
		CreatedBefore:      query.CreatedBefore,
	}
	
	// Buscar o total de registros que atendem aos critérios (para paginação)
	totalCount, err := h.attributeService.CountAttributesBySearch(ctx, searchFilters)
	if err != nil {
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "SEARCH_ATTRIBUTES_FAILED",
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      query.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":  operationID,
				"error":         err.Error(),
				"duration_ms":   time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, fmt.Errorf("erro ao contar atributos: %w", err)
	}
	
	// Calcular informações de paginação
	pageCount := calculatePageCount(totalCount, query.PageSize)
	
	// Ajustar página se necessário
	if query.Page >= pageCount && totalCount > 0 {
		query.Page = pageCount - 1
	}
	
	// Buscar os atributos
	offset := query.Page * query.PageSize
	attributes, err := h.attributeService.SearchAttributes(ctx, searchFilters, offset, query.PageSize)
	if err != nil {
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "SEARCH_ATTRIBUTES_FAILED",
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      query.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":  operationID,
				"error":         err.Error(),
				"duration_ms":   time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, fmt.Errorf("erro ao buscar atributos: %w", err)
	}
	
	// Enriquecer resultados com informações de contexto (opcional)
	if len(attributes) > 0 {
		contextIDs := make(map[uuid.UUID]bool)
		for _, attr := range attributes {
			contextIDs[attr.ContextID] = true
		}
		
		// Carregar informações de contexto em batch para melhor performance
		contextMap := make(map[uuid.UUID]*models.IdentityContext)
		if len(contextIDs) > 0 {
			ctxIDs := make([]uuid.UUID, 0, len(contextIDs))
			for id := range contextIDs {
				ctxIDs = append(ctxIDs, id)
			}
			
			contexts, err := h.contextService.GetContextsByIDs(ctx, ctxIDs)
			if err != nil {
				// Apenas logar o erro, mas continuar com os resultados
				h.auditLogger.LogEvent(ctx, services.AuditEvent{
					EventType:   "LOAD_CONTEXTS_FAILED",
					ResourceType: "IDENTITY_CONTEXT",
					UserID:      query.RequestedBy,
					Timestamp:   time.Now(),
					Details: map[string]interface{}{
						"operation_id":  operationID,
						"context_count": len(ctxIDs),
						"error":         err.Error(),
					},
				})
			} else {
				for _, ctx := range contexts {
					contextMap[ctx.ID] = ctx
				}
			}
		}
		
		// Adicionar informações de contexto aos metadados dos atributos para melhor visualização
		for _, attr := range attributes {
			if ctx, exists := contextMap[attr.ContextID]; exists {
				if attr.Metadata == nil {
					attr.Metadata = make(map[string]interface{})
				}
				
				// Adicionar informações básicas do contexto para facilitar a visualização
				attr.Metadata["_context_info"] = map[string]interface{}{
					"user_id":       ctx.UserID,
					"tenant_id":     ctx.TenantID,
					"context_type":  ctx.ContextType,
					"trust_score":   ctx.TrustScore,
					"status":        ctx.Status,
					"display_name":  ctx.DisplayName,
				}
			}
		}
	}
	
	// Preparar resultado
	result := &AttributesResult{
		Attributes:   attributes,
		TotalCount:   totalCount,
		PageCount:    pageCount,
		CurrentPage:  query.Page,
		PageSize:     query.PageSize,
		HasMore:      (query.Page+1) < pageCount,
	}
	
	// Registrar sucesso da operação
	h.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "SEARCH_ATTRIBUTES_SUCCEEDED",
		ResourceType: "CONTEXT_ATTRIBUTE",
		UserID:      query.RequestedBy,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"operation_id":   operationID,
			"search_text":    query.SearchText,
			"result_count":   len(attributes),
			"total_count":    totalCount,
			"page":           query.Page,
			"page_count":     pageCount,
			"duration_ms":    time.Since(startTime).Milliseconds(),
		},
	})
	
	return result, nil
}

// getSearchQueryFilters retorna um mapa com os filtros aplicados para auditoria
func getSearchQueryFilters(query SearchAttributesQuery) map[string]interface{} {
	filters := make(map[string]interface{})
	
	if len(query.ContextIDs) > 0 {
		contextIDStrings := make([]string, len(query.ContextIDs))
		for i, id := range query.ContextIDs {
			contextIDStrings[i] = id.String()
		}
		filters["context_ids"] = contextIDStrings
	}
	
	if query.UserID != nil {
		filters["user_id"] = *query.UserID
	}
	
	if query.TenantID != nil {
		filters["tenant_id"] = *query.TenantID
	}
	
	if len(query.SensitivityLevels) > 0 {
		filters["sensitivity_levels"] = query.SensitivityLevels
	}
	
	if len(query.VerificationStatuses) > 0 {
		filters["verification_statuses"] = query.VerificationStatuses
	}
	
	if len(query.ContextTypes) > 0 {
		filters["context_types"] = query.ContextTypes
	}
	
	if query.CreatedAfter != nil {
		filters["created_after"] = query.CreatedAfter.Format(time.RFC3339)
	}
	
	if query.CreatedBefore != nil {
		filters["created_before"] = query.CreatedBefore.Format(time.RFC3339)
	}
	
	filters["include_inactive"] = query.IncludeInactive
	
	return filters
}