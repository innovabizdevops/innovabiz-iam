/**
 * @file attribute_handlers_part2.go
 * @description Handlers REST para operações avançadas com atributos de contexto
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
	"innovabiz/iam/src/multi-context/domain/models"
)

// SearchAttributes realiza busca avançada por atributos
func (c *MultiContextController) SearchAttributes(w http.ResponseWriter, r *http.Request) {
	// Extrair informações do usuário
	userInfo := extractUserInfo(r)

	// Verificar se usuário tem permissão para busca avançada (somente admin)
	if !userInfo.IsAdmin {
		respondWithError(w, http.StatusForbidden, "FORBIDDEN", "Apenas administradores podem realizar busca avançada")
		c.auditLogger.LogSecurityEvent(r.Context(), "attributes.search.forbidden", map[string]interface{}{
			"userId":    userInfo.UserID,
			"timestamp": time.Now(),
		})
		return
	}

	// Decodificar payload de busca avançada
	var searchInput struct {
		Query         string   `json:"query"`
		ContextIDs    []string `json:"contextIds"`
		Types         []string `json:"types"`
		VerifiedOnly  bool     `json:"verifiedOnly"`
		ExcludedNames []string `json:"excludedNames"`
		Pagination    struct {
			Page     int    `json:"page"`
			PageSize int    `json:"pageSize"`
			SortBy   string `json:"sortBy"`
			SortDir  string `json:"sortDir"`
		} `json:"pagination"`
	}

	if err := json.NewDecoder(r.Body).Decode(&searchInput); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Payload de busca inválido")
		return
	}

	// Converter IDs de contexto para UUIDs
	var contextIDs []uuid.UUID
	for _, idStr := range searchInput.ContextIDs {
		if id, err := uuid.Parse(idStr); err == nil {
			contextIDs = append(contextIDs, id)
		}
	}

	// Converter tipos de atributos para enum
	var types []models.AttributeType
	for _, typeStr := range searchInput.Types {
		switch typeStr {
		case "TEXT":
			types = append(types, models.AttributeTypeText)
		case "NUMBER":
			types = append(types, models.AttributeTypeNumber)
		case "BOOLEAN":
			types = append(types, models.AttributeTypeBoolean)
		case "DATE":
			types = append(types, models.AttributeTypeDate)
		case "DOCUMENT":
			types = append(types, models.AttributeTypeDocument)
		case "BIOMETRIC":
			types = append(types, models.AttributeTypeBiometric)
		case "ADDRESS":
			types = append(types, models.AttributeTypeAddress)
		case "EMAIL":
			types = append(types, models.AttributeTypeEmail)
		case "PHONE":
			types = append(types, models.AttributeTypePhone)
		}
	}

	// Construir query de busca
	query := models.AttributeSearchQuery{
		Query:         searchInput.Query,
		ContextIDs:    contextIDs,
		Types:         types,
		VerifiedOnly:  searchInput.VerifiedOnly,
		ExcludedNames: searchInput.ExcludedNames,
		UserID:        userInfo.UserID,
		TenantID:      userInfo.TenantID,
		Page:          searchInput.Pagination.Page,
		PageSize:      searchInput.Pagination.PageSize,
		SortBy:        searchInput.Pagination.SortBy,
		SortDirection: searchInput.Pagination.SortDir,
	}

	// Executar busca avançada
	result, err := c.searchAttributesHandler.Handle(r.Context(), query)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao buscar atributos")
		c.auditLogger.LogSecurityEvent(r.Context(), "attributes.search.error", map[string]interface{}{
			"userId":    userInfo.UserID,
			"error":     err.Error(),
			"timestamp": time.Now(),
		})
		return
	}

	// Mapear resultado para resposta
	response := mapAttributesToResponse(result)
	respondWithJSON(w, http.StatusOK, response)

	// Registrar sucesso
	c.auditLogger.LogSecurityEvent(r.Context(), "attributes.search.success", map[string]interface{}{
		"userId":    userInfo.UserID,
		"count":     len(result.Items),
		"total":     result.TotalCount,
		"timestamp": time.Now(),
	})
}

// GetAttributeVerificationHistory retorna o histórico de verificação de um atributo
func (c *MultiContextController) GetAttributeVerificationHistory(w http.ResponseWriter, r *http.Request) {
	// Extrair ID do atributo
	params := mux.Vars(r)
	attributeID, err := parseUUID(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_ID", "ID de atributo inválido")
		return
	}

	// Extrair informações do usuário
	userInfo := extractUserInfo(r)

	// Buscar o atributo
	attribute, err := c.attributeService.FindByID(r.Context(), attributeID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao buscar atributo")
		return
	}

	// Verificar se o atributo foi encontrado
	if attribute == nil {
		respondWithError(w, http.StatusNotFound, "ATTRIBUTE_NOT_FOUND", "Atributo não encontrado")
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
		return
	}

	// Extrair histórico de verificação
	history, err := c.attributeService.GetVerificationHistory(r.Context(), attributeID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao buscar histórico de verificação")
		return
	}

	// Mapear para DTO e responder
	response := mapVerificationHistoryToResponse(history)
	respondWithJSON(w, http.StatusOK, response)
}

// CreateAttribute cria um novo atributo para um contexto
func (c *MultiContextController) CreateAttribute(w http.ResponseWriter, r *http.Request) {
	// Extrair informações do usuário
	userInfo := extractUserInfo(r)

	// Decodificar payload
	var input struct {
		ContextID     string                 `json:"contextId"`
		Name          string                 `json:"name"`
		Type          string                 `json:"type"`
		Value         interface{}            `json:"value"`
		Sensitivity   string                 `json:"sensitivity"`
		Metadata      map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Payload inválido")
		return
	}

	// Validar entrada básica
	if input.ContextID == "" || input.Name == "" || input.Type == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "ContextID, nome e tipo são obrigatórios")
		return
	}

	// Converter contextID para UUID
	contextID, err := uuid.Parse(input.ContextID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_CONTEXT_ID", "ID de contexto inválido")
		return
	}

	// Verificar se o contexto existe
	context, err := c.contextService.FindByID(r.Context(), contextID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao buscar contexto")
		return
	}
	if context == nil {
		respondWithError(w, http.StatusNotFound, "CONTEXT_NOT_FOUND", "Contexto não encontrado")
		return
	}

	// Verificar autorização (usuário deve ser dono do contexto ou admin)
	if context.UserID != userInfo.UserID && !userInfo.IsAdmin {
		respondWithError(w, http.StatusForbidden, "FORBIDDEN", "Apenas o dono do contexto ou administradores podem criar atributos")
		return
	}

	// Converter tipo para enum
	var attrType models.AttributeType
	switch input.Type {
	case "TEXT":
		attrType = models.AttributeTypeText
	case "NUMBER":
		attrType = models.AttributeTypeNumber
	case "BOOLEAN":
		attrType = models.AttributeTypeBoolean
	case "DATE":
		attrType = models.AttributeTypeDate
	case "DOCUMENT":
		attrType = models.AttributeTypeDocument
	case "BIOMETRIC":
		attrType = models.AttributeTypeBiometric
	case "ADDRESS":
		attrType = models.AttributeTypeAddress
	case "EMAIL":
		attrType = models.AttributeTypeEmail
	case "PHONE":
		attrType = models.AttributeTypePhone
	default:
		respondWithError(w, http.StatusBadRequest, "INVALID_TYPE", "Tipo de atributo inválido")
		return
	}

	// Converter sensibilidade para enum
	var sensitivity models.AttributeSensitivity
	switch input.Sensitivity {
	case "LOW":
		sensitivity = models.SensitivityLow
	case "MEDIUM":
		sensitivity = models.SensitivityMedium
	case "HIGH":
		sensitivity = models.SensitivityHigh
	default:
		sensitivity = models.SensitivityLow // Valor padrão se não especificado
	}

	// Criar comando
	cmd := commands.CreateAttributeCommand{
		ContextID:   contextID,
		Name:        input.Name,
		Type:        attrType,
		Value:       input.Value,
		Sensitivity: sensitivity,
		Metadata:    input.Metadata,
		CreatorID:   userInfo.UserID,
	}

	// Executar comando
	attribute, err := c.createAttributeHandler.Handle(r.Context(), cmd)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao criar atributo: "+err.Error())
		return
	}

	// Mapear resultado e responder
	response := mapAttributeToResponse(attribute)
	respondWithJSON(w, http.StatusCreated, response)
}