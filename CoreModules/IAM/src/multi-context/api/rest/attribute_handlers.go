/**
 * @file attribute_handlers.go
 * @description Handlers REST para operações com atributos de contexto
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package rest

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"innovabiz/iam/src/multi-context/application/commands"
	"innovabiz/iam/src/multi-context/application/queries"
	"innovabiz/iam/src/multi-context/domain/models"
)

// GetAttribute retorna um atributo específico pelo ID
func (c *MultiContextController) GetAttribute(w http.ResponseWriter, r *http.Request) {
	// Extrair ID do atributo
	params := mux.Vars(r)
	attributeID, err := parseUUID(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_ID", "ID de atributo inválido")
		return
	}

	// Extrair informações do usuário
	userInfo := extractUserInfo(r)

	// Registrar tentativa de acesso
	c.auditLogger.LogSecurityEvent(r.Context(), "attribute.get.attempt", map[string]interface{}{
		"attributeId": attributeID.String(),
		"userId":      userInfo.UserID,
		"tenantId":    userInfo.TenantID,
		"timestamp":   time.Now(),
	})

	// Buscar o atributo
	attribute, err := c.attributeService.FindByID(r.Context(), attributeID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao buscar atributo")
		c.auditLogger.LogSecurityEvent(r.Context(), "attribute.get.error", map[string]interface{}{
			"attributeId": attributeID.String(),
			"userId":      userInfo.UserID,
			"error":       err.Error(),
			"timestamp":   time.Now(),
		})
		return
	}

	// Verificar se o atributo foi encontrado
	if attribute == nil {
		respondWithError(w, http.StatusNotFound, "ATTRIBUTE_NOT_FOUND", "Atributo não encontrado")
		c.auditLogger.LogSecurityEvent(r.Context(), "attribute.get.not_found", map[string]interface{}{
			"attributeId": attributeID.String(),
			"userId":      userInfo.UserID,
			"timestamp":   time.Now(),
		})
		return
	}

	// Buscar o contexto do atributo
	context, err := c.contextService.FindByID(r.Context(), attribute.ContextID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao buscar contexto do atributo")
		return
	}

	// Verificar autorização
	if !c.isUserAuthorizedForContext(userInfo, *context) {
		respondWithError(w, http.StatusForbidden, "FORBIDDEN", "Acesso negado ao atributo")
		c.auditLogger.LogSecurityEvent(r.Context(), "attribute.get.forbidden", map[string]interface{}{
			"attributeId": attributeID.String(),
			"contextId":   attribute.ContextID.String(),
			"userId":      userInfo.UserID,
			"timestamp":   time.Now(),
		})
		return
	}

	// Mapear para DTO e responder
	response := mapAttributeToResponse(*attribute)
	respondWithJSON(w, http.StatusOK, response)

	// Registrar sucesso
	c.auditLogger.LogSecurityEvent(r.Context(), "attribute.get.success", map[string]interface{}{
		"attributeId": attributeID.String(),
		"userId":      userInfo.UserID,
		"timestamp":   time.Now(),
	})
}

// ListAttributes lista os atributos com filtros e paginação
func (c *MultiContextController) ListAttributes(w http.ResponseWriter, r *http.Request) {
	// Extrair informações do usuário
	userInfo := extractUserInfo(r)

	// Extrair parâmetros de paginação e filtros
	pagination := parsePaginationParams(r)
	filters := parseAttributeFilters(r)

	// Mapear para query
	query := models.AttributeQuery{
		UserID:         userInfo.UserID,
		TenantID:       userInfo.TenantID,
		IncludeDeleted: false,
		Page:           pagination.Page,
		PageSize:       pagination.PageSize,
		SortBy:         pagination.SortBy,
		SortDirection:  pagination.SortDir,
	}

	// Aplicar filtros específicos
	if contextID, err := parseUUID(filters.ContextID); err == nil {
		query.ContextID = &contextID
	}
	query.Name = filters.Name
	query.AttributeType = filters.Type
	query.VerificationStatus = filters.VerificationStatus

	// Aplicar filtros avançados se o usuário for admin
	if userInfo.IsAdmin {
		query.IncludeDeleted = filters.IncludeDeleted
	}

	// Registrar tentativa de listagem
	c.auditLogger.LogSecurityEvent(r.Context(), "attributes.list.attempt", map[string]interface{}{
		"userId":   userInfo.UserID,
		"tenantId": userInfo.TenantID,
		"filters":  filters,
		"timestamp": time.Now(),
	})

	// Executar consulta
	result, err := c.listAttributesHandler.Handle(r.Context(), query)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao listar atributos")
		c.auditLogger.LogSecurityEvent(r.Context(), "attributes.list.error", map[string]interface{}{
			"userId":   userInfo.UserID,
			"error":    err.Error(),
			"timestamp": time.Now(),
		})
		return
	}

	// Mapear resultado para resposta
	response := mapAttributesToResponse(result)
	respondWithJSON(w, http.StatusOK, response)

	// Registrar sucesso
	c.auditLogger.LogSecurityEvent(r.Context(), "attributes.list.success", map[string]interface{}{
		"userId":   userInfo.UserID,
		"count":    len(result.Items),
		"total":    result.TotalCount,
		"timestamp": time.Now(),
	})
}