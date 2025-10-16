/**
 * @file attribute_handlers_part3.go
 * @description Handlers REST para operações de atualização e verificação de atributos
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

// UpdateAttribute atualiza um atributo existente
func (c *MultiContextController) UpdateAttribute(w http.ResponseWriter, r *http.Request) {
	// Extrair ID do atributo
	params := mux.Vars(r)
	attributeID, err := parseUUID(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_ID", "ID de atributo inválido")
		return
	}

	// Extrair informações do usuário
	userInfo := extractUserInfo(r)

	// Buscar o atributo existente
	attribute, err := c.attributeService.FindByID(r.Context(), attributeID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao buscar atributo")
		return
	}
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
	if context == nil {
		respondWithError(w, http.StatusNotFound, "CONTEXT_NOT_FOUND", "Contexto do atributo não encontrado")
		return
	}

	// Verificar autorização (apenas o dono do contexto ou admin pode atualizar)
	if context.UserID != userInfo.UserID && !userInfo.IsAdmin {
		respondWithError(w, http.StatusForbidden, "FORBIDDEN", "Apenas o dono do contexto ou administradores podem atualizar atributos")
		return
	}

	// Decodificar payload
	var input struct {
		Value       interface{}            `json:"value"`
		Sensitivity string                 `json:"sensitivity,omitempty"`
		Metadata    map[string]interface{} `json:"metadata,omitempty"`
		Reason      string                 `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Payload inválido")
		return
	}

	// Validar entrada básica
	if input.Value == nil || input.Reason == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Valor e motivo da alteração são obrigatórios")
		return
	}

	// Converter sensibilidade para enum se presente
	var sensitivity *models.AttributeSensitivity
	if input.Sensitivity != "" {
		var s models.AttributeSensitivity
		switch input.Sensitivity {
		case "LOW":
			s = models.SensitivityLow
		case "MEDIUM":
			s = models.SensitivityMedium
		case "HIGH":
			s = models.SensitivityHigh
		default:
			respondWithError(w, http.StatusBadRequest, "INVALID_SENSITIVITY", "Sensibilidade inválida")
			return
		}
		sensitivity = &s
	}

	// Criar comando
	cmd := commands.UpdateAttributeCommand{
		AttributeID:  attributeID,
		Value:        input.Value,
		Sensitivity:  sensitivity,
		Metadata:     input.Metadata,
		Reason:       input.Reason,
		UpdaterID:    userInfo.UserID,
	}

	// Executar comando
	updatedAttribute, err := c.updateAttributeHandler.Handle(r.Context(), cmd)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao atualizar atributo: "+err.Error())
		return
	}

	// Mapear resultado e responder
	response := mapAttributeToResponse(updatedAttribute)
	respondWithJSON(w, http.StatusOK, response)

	// Registrar sucesso
	c.auditLogger.LogSecurityEvent(r.Context(), "attribute.update.success", map[string]interface{}{
		"attributeId": attributeID.String(),
		"contextId":   attribute.ContextID.String(),
		"userId":      userInfo.UserID,
		"timestamp":   time.Now(),
	})
}

// VerifyAttribute verifica um atributo
func (c *MultiContextController) VerifyAttribute(w http.ResponseWriter, r *http.Request) {
	// Extrair ID do atributo
	params := mux.Vars(r)
	attributeID, err := parseUUID(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_ID", "ID de atributo inválido")
		return
	}

	// Extrair informações do usuário
	userInfo := extractUserInfo(r)

	// Verificar se o usuário é verificador
	if !userHasRole(userInfo, "verifier") && !userInfo.IsAdmin {
		respondWithError(w, http.StatusForbidden, "FORBIDDEN", "Apenas verificadores podem verificar atributos")
		return
	}

	// Buscar o atributo existente
	attribute, err := c.attributeService.FindByID(r.Context(), attributeID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao buscar atributo")
		return
	}
	if attribute == nil {
		respondWithError(w, http.StatusNotFound, "ATTRIBUTE_NOT_FOUND", "Atributo não encontrado")
		return
	}

	// Decodificar payload
	var input struct {
		Status     string `json:"status"`
		Reason     string `json:"reason"`
		Evidence   string `json:"evidence,omitempty"`
		ExpiresAt  string `json:"expiresAt,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Payload inválido")
		return
	}

	// Validar entrada básica
	if input.Status == "" || input.Reason == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Status e motivo são obrigatórios")
		return
	}

	// Converter status para enum
	var status models.VerificationStatus
	switch input.Status {
	case "VERIFIED":
		status = models.StatusVerified
	case "REJECTED":
		status = models.StatusRejected
	case "PENDING":
		status = models.StatusPending
	default:
		respondWithError(w, http.StatusBadRequest, "INVALID_STATUS", "Status de verificação inválido")
		return
	}

	// Converter data de expiração se presente
	var expiresAt *time.Time
	if input.ExpiresAt != "" {
		expiry, err := time.Parse(time.RFC3339, input.ExpiresAt)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "INVALID_DATE", "Data de expiração inválida")
			return
		}
		expiresAt = &expiry
	}

	// Criar comando
	cmd := commands.VerifyAttributeCommand{
		AttributeID:  attributeID,
		Status:       status,
		Reason:       input.Reason,
		Evidence:     input.Evidence,
		VerifierID:   userInfo.UserID,
		ExpiresAt:    expiresAt,
	}

	// Executar comando
	verifiedAttribute, err := c.verifyAttributeHandler.Handle(r.Context(), cmd)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao verificar atributo: "+err.Error())
		return
	}

	// Mapear resultado e responder
	response := mapAttributeToResponse(verifiedAttribute)
	respondWithJSON(w, http.StatusOK, response)

	// Registrar sucesso
	c.auditLogger.LogSecurityEvent(r.Context(), "attribute.verify.success", map[string]interface{}{
		"attributeId": attributeID.String(),
		"status":      string(status),
		"userId":      userInfo.UserID,
		"timestamp":   time.Now(),
	})
}

// DeleteAttribute exclui logicamente um atributo
func (c *MultiContextController) DeleteAttribute(w http.ResponseWriter, r *http.Request) {
	// Extrair ID do atributo
	params := mux.Vars(r)
	attributeID, err := parseUUID(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_ID", "ID de atributo inválido")
		return
	}

	// Extrair informações do usuário
	userInfo := extractUserInfo(r)

	// Buscar o atributo existente
	attribute, err := c.attributeService.FindByID(r.Context(), attributeID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao buscar atributo")
		return
	}
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
	if context == nil {
		respondWithError(w, http.StatusNotFound, "CONTEXT_NOT_FOUND", "Contexto do atributo não encontrado")
		return
	}

	// Verificar autorização (apenas o dono do contexto ou admin pode deletar)
	if context.UserID != userInfo.UserID && !userInfo.IsAdmin {
		respondWithError(w, http.StatusForbidden, "FORBIDDEN", "Apenas o dono do contexto ou administradores podem excluir atributos")
		return
	}

	// Decodificar payload para obter o motivo da exclusão
	var input struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Reason == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Motivo da exclusão é obrigatório")
		return
	}

	// Excluir logicamente o atributo
	deleted, err := c.attributeService.SoftDelete(r.Context(), attributeID, userInfo.UserID, input.Reason)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao excluir atributo: "+err.Error())
		return
	}

	// Responder com sucesso
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": deleted,
		"message": "Atributo excluído com sucesso",
	})

	// Registrar sucesso
	c.auditLogger.LogSecurityEvent(r.Context(), "attribute.delete.success", map[string]interface{}{
		"attributeId": attributeID.String(),
		"contextId":   attribute.ContextID.String(),
		"userId":      userInfo.UserID,
		"reason":      input.Reason,
		"timestamp":   time.Now(),
	})
}

// parseAttributeFilters extrai filtros de atributos da requisição
func parseAttributeFilters(r *http.Request) struct {
	ContextID          string
	Name               string
	Type               string
	VerificationStatus string
	IncludeDeleted     bool
} {
	q := r.URL.Query()
	
	includeDeleted := false
	if incDelStr := q.Get("includeDeleted"); incDelStr == "true" {
		includeDeleted = true
	}
	
	return struct {
		ContextID          string
		Name               string
		Type               string
		VerificationStatus string
		IncludeDeleted     bool
	}{
		ContextID:          q.Get("contextId"),
		Name:               q.Get("name"),
		Type:               q.Get("type"),
		VerificationStatus: q.Get("verificationStatus"),
		IncludeDeleted:     includeDeleted,
	}
}

// mapAttributeToResponse mapeia um modelo de atributo para uma resposta API
func mapAttributeToResponse(attribute models.Attribute) map[string]interface{} {
	// Extrair histórico de verificação do metadata
	verificationHistory := extractVerificationHistoryFromMetadata(attribute.Metadata)
	
	return map[string]interface{}{
		"id":                 attribute.ID.String(),
		"contextId":          attribute.ContextID.String(),
		"name":               attribute.Name,
		"type":               string(attribute.Type),
		"value":              attribute.Value,
		"sensitivity":        string(attribute.Sensitivity),
		"verificationStatus": string(attribute.VerificationStatus),
		"createdAt":          attribute.CreatedAt,
		"updatedAt":          attribute.UpdatedAt,
		"verifiedAt":         attribute.VerifiedAt,
		"verifierId":         attribute.VerifierID,
		"metadata":           attribute.Metadata,
		"verificationHistory": verificationHistory,
		"isDeleted":          attribute.DeletedAt != nil,
	}
}

// mapAttributesToResponse mapeia resultados de consulta para resposta API
func mapAttributesToResponse(result queries.AttributeQueryResult) map[string]interface{} {
	items := make([]map[string]interface{}, len(result.Items))
	for i, attr := range result.Items {
		items[i] = mapAttributeToResponse(attr)
	}
	
	return map[string]interface{}{
		"items":      items,
		"totalCount": result.TotalCount,
		"page":       result.Page,
		"pageSize":   result.PageSize,
	}
}