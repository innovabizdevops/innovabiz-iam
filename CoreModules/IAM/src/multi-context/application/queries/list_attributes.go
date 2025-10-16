/**
 * @file list_attributes.go
 * @description Consulta e handler para listagem de atributos contextuais
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package queries

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	
	"innovabiz/iam/src/multi-context/domain/models"
	"innovabiz/iam/src/multi-context/domain/services"
)

// ListAttributesQuery representa a consulta para listar atributos contextuais
type ListAttributesQuery struct {
	ContextID         *uuid.UUID                     // Filtrar por ID do contexto (opcional)
	KeyPattern        *string                        // Filtrar por padrão da chave (opcional)
	SensitivityLevel  *models.SensitivityLevel       // Filtrar por nível de sensibilidade (opcional)
	VerificationStatus *models.VerificationStatus     // Filtrar por status de verificação (opcional)
	CreatedAfter      *time.Time                     // Filtrar por data de criação após (opcional)
	CreatedBefore     *time.Time                     // Filtrar por data de criação antes (opcional)
	SortBy            string                         // Campo para ordenação (opcional, padrão: "created_at")
	SortDirection     string                         // Direção da ordenação (opcional, "asc" ou "desc", padrão: "desc")
	Page              int                            // Página a ser retornada (começando em 0)
	PageSize          int                            // Tamanho da página
	RequestedBy       string                         // Utilizador ou sistema que solicitou a consulta
}

// AttributesResult representa o resultado da listagem de atributos
type AttributesResult struct {
	Attributes   []*models.ContextAttribute  // Lista de atributos
	TotalCount   int                         // Número total de atributos (sem paginação)
	PageCount    int                         // Número total de páginas
	CurrentPage  int                         // Página atual
	PageSize     int                         // Tamanho da página
	HasMore      bool                        // Se existem mais páginas
}

// ListAttributesHandler gerencia a consulta para listar atributos contextuais
type ListAttributesHandler struct {
	attributeService *services.AttributeService
	contextService   *services.ContextService
	auditLogger      services.AuditLogger
}

// NewListAttributesHandler cria uma nova instância do handler
func NewListAttributesHandler(
	attributeService *services.AttributeService,
	contextService *services.ContextService,
	auditLogger services.AuditLogger,
) *ListAttributesHandler {
	return &ListAttributesHandler{
		attributeService: attributeService,
		contextService:   contextService,
		auditLogger:      auditLogger,
	}
}

// Handle processa a consulta de listagem de atributos
func (h *ListAttributesHandler) Handle(ctx context.Context, query ListAttributesQuery) (*AttributesResult, error) {
	// Registrar início da operação para rastreabilidade
	startTime := time.Now()
	operationID := uuid.New()
	
	h.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "LIST_ATTRIBUTES_INITIATED",
		ResourceType: "CONTEXT_ATTRIBUTE",
		UserID:      query.RequestedBy,
		Timestamp:   startTime,
		Details: map[string]interface{}{
			"operation_id":      operationID,
			"filters":           getAttributeQueryFilters(query),
			"page":              query.Page,
			"page_size":         query.PageSize,
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
	
	// Validar acesso ao contexto, se especificado
	if query.ContextID != nil {
		// Verificar se o contexto existe e está acessível
		context, err := h.contextService.GetContextByID(ctx, *query.ContextID)
		if err != nil {
			h.auditLogger.LogEvent(ctx, services.AuditEvent{
				EventType:   "LIST_ATTRIBUTES_FAILED",
				ResourceType: "CONTEXT_ATTRIBUTE",
				UserID:      query.RequestedBy,
				Timestamp:   time.Now(),
				Details: map[string]interface{}{
					"operation_id":      operationID,
					"error":             fmt.Sprintf("contexto não encontrado: %s", err.Error()),
					"context_id":        query.ContextID,
					"duration_ms":       time.Since(startTime).Milliseconds(),
				},
			})
			
			return nil, fmt.Errorf("contexto não encontrado: %w", err)
		}
		
		// Verificar se o contexto está ativo (a menos que seja uma consulta administrativa)
		if context.Status != models.ContextStatusActive {
			// Verificar se o solicitante tem permissões administrativas (implementação específica)
			isAdmin := hasAdminPermissions(ctx, query.RequestedBy)
			
			if !isAdmin {
				h.auditLogger.LogEvent(ctx, services.AuditEvent{
					EventType:   "LIST_ATTRIBUTES_FAILED",
					ResourceType: "CONTEXT_ATTRIBUTE",
					UserID:      query.RequestedBy,
					Timestamp:   time.Now(),
					Details: map[string]interface{}{
						"operation_id":      operationID,
						"error":             "acesso negado a contexto não ativo",
						"context_id":        query.ContextID,
						"context_status":    context.Status,
						"duration_ms":       time.Since(startTime).Milliseconds(),
					},
				})
				
				return nil, fmt.Errorf("acesso negado a contexto não ativo")
			}
		}
	}
	
	// Preparar filtros
	filters := models.AttributeFilters{
		ContextID:         query.ContextID,
		KeyPattern:        query.KeyPattern,
		SensitivityLevel:  query.SensitivityLevel,
		VerificationStatus: query.VerificationStatus,
		CreatedAfter:      query.CreatedAfter,
		CreatedBefore:     query.CreatedBefore,
		SortBy:            query.SortBy,
		SortDirection:     query.SortDirection,
	}
	
	// Buscar o total de registros que atendem aos critérios (para paginação)
	totalCount, err := h.attributeService.CountAttributes(ctx, filters)
	if err != nil {
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "LIST_ATTRIBUTES_FAILED",
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      query.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":      operationID,
				"error":             err.Error(),
				"duration_ms":       time.Since(startTime).Milliseconds(),
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
	attributes, err := h.attributeService.ListAttributes(ctx, filters, offset, query.PageSize)
	if err != nil {
		h.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "LIST_ATTRIBUTES_FAILED",
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      query.RequestedBy,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"operation_id":      operationID,
				"error":             err.Error(),
				"duration_ms":       time.Since(startTime).Milliseconds(),
			},
		})
		
		return nil, fmt.Errorf("erro ao listar atributos: %w", err)
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
		EventType:   "LIST_ATTRIBUTES_SUCCEEDED",
		ResourceType: "CONTEXT_ATTRIBUTE",
		UserID:      query.RequestedBy,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"operation_id":      operationID,
			"result_count":      len(attributes),
			"total_count":       totalCount,
			"page":              query.Page,
			"page_size":         query.PageSize,
			"page_count":        pageCount,
			"duration_ms":       time.Since(startTime).Milliseconds(),
		},
	})
	
	return result, nil
}

// getAttributeQueryFilters retorna um mapa com os filtros aplicados para auditoria
func getAttributeQueryFilters(query ListAttributesQuery) map[string]interface{} {
	filters := make(map[string]interface{})
	
	if query.ContextID != nil {
		filters["context_id"] = query.ContextID.String()
	}
	
	if query.KeyPattern != nil {
		filters["key_pattern"] = *query.KeyPattern
	}
	
	if query.SensitivityLevel != nil {
		filters["sensitivity_level"] = *query.SensitivityLevel
	}
	
	if query.VerificationStatus != nil {
		filters["verification_status"] = *query.VerificationStatus
	}
	
	if query.CreatedAfter != nil {
		filters["created_after"] = query.CreatedAfter.Format(time.RFC3339)
	}
	
	if query.CreatedBefore != nil {
		filters["created_before"] = query.CreatedBefore.Format(time.RFC3339)
	}
	
	filters["sort_by"] = query.SortBy
	filters["sort_direction"] = query.SortDirection
	
	return filters
}

// hasAdminPermissions verifica se o usuário tem permissões administrativas
// Esta é uma implementação simples para fins de demonstração
// Em um sistema real, isso seria integrado com o sistema de controle de acesso
func hasAdminPermissions(ctx context.Context, userID string) bool {
	// Em um sistema real, consultar o serviço de autorização ou IAM
	// para verificar as permissões do usuário
	
	// Implementação temporária para demonstração
	// Assumindo que usuários do sistema e administradores têm permissão
	return userID == "system" || 
	       strings.HasPrefix(userID, "admin") || 
	       ctx.Value("is_admin") == true
}