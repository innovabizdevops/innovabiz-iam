/**
 * @file context_handlers.go
 * @description Handlers REST para operações com contextos de identidade
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package rest

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"innovabiz/iam/src/multi-context/domain/models"
)

// GetContext retorna um contexto específico pelo ID
func (c *MultiContextController) GetContext(w http.ResponseWriter, r *http.Request) {
	// Extrair ID do contexto
	params := mux.Vars(r)
	contextID, err := parseUUID(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_ID", "ID de contexto inválido")
		return
	}

	// Extrair informações do usuário
	userInfo := extractUserInfo(r)

	// Registrar tentativa de acesso
	c.auditLogger.LogSecurityEvent(r.Context(), "context.get.attempt", map[string]interface{}{
		"contextId": contextID.String(),
		"userId":    userInfo.UserID,
		"tenantId":  userInfo.TenantID,
		"timestamp": time.Now(),
	})

	// Buscar o contexto
	context, err := c.contextService.FindByID(r.Context(), contextID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao buscar contexto")
		c.auditLogger.LogSecurityEvent(r.Context(), "context.get.error", map[string]interface{}{
			"contextId": contextID.String(),
			"userId":    userInfo.UserID,
			"error":     err.Error(),
			"timestamp": time.Now(),
		})
		return
	}

	// Verificar se o contexto foi encontrado
	if context == nil {
		respondWithError(w, http.StatusNotFound, "CONTEXT_NOT_FOUND", "Contexto não encontrado")
		c.auditLogger.LogSecurityEvent(r.Context(), "context.get.not_found", map[string]interface{}{
			"contextId": contextID.String(),
			"userId":    userInfo.UserID,
			"timestamp": time.Now(),
		})
		return
	}

	// Verificar autorização
	if !c.isUserAuthorizedForContext(userInfo, *context) {
		respondWithError(w, http.StatusForbidden, "FORBIDDEN", "Acesso negado ao contexto")
		c.auditLogger.LogSecurityEvent(r.Context(), "context.get.forbidden", map[string]interface{}{
			"contextId": contextID.String(),
			"userId":    userInfo.UserID,
			"timestamp": time.Now(),
		})
		return
	}

	// Mapear para DTO e responder
	response := mapContextToResponse(*context)
	respondWithJSON(w, http.StatusOK, response)

	// Registrar sucesso
	c.auditLogger.LogSecurityEvent(r.Context(), "context.get.success", map[string]interface{}{
		"contextId": contextID.String(),
		"userId":    userInfo.UserID,
		"timestamp": time.Now(),
	})
}

// ListContexts lista os contextos com filtros e paginação
func (c *MultiContextController) ListContexts(w http.ResponseWriter, r *http.Request) {
	// Extrair informações do usuário
	userInfo := extractUserInfo(r)

	// Extrair parâmetros de paginação e filtros
	pagination := parsePaginationParams(r)
	filters := parseContextFilters(r)

	// Mapear para query
	query := models.ContextQuery{
		UserID:         userInfo.UserID,
		TenantID:       userInfo.TenantID,
		IncludeDeleted: false, // Default é não incluir deletados
		Page:           pagination.Page,
		PageSize:       pagination.PageSize,
		SortBy:         pagination.SortBy,
		SortDirection:  pagination.SortDir,
	}

	// Aplicar filtros avançados se o usuário for admin
	if userInfo.IsAdmin {
		query.ContextType = filters.Type
		query.VerificationLevel = filters.VerificationLevel
		query.MinTrustScore = filters.MinTrustScore
		query.MaxTrustScore = filters.MaxTrustScore
		query.IncludeDeleted = filters.IncludeDeleted
	}

	// Registrar tentativa de listagem
	c.auditLogger.LogSecurityEvent(r.Context(), "contexts.list.attempt", map[string]interface{}{
		"userId":   userInfo.UserID,
		"tenantId": userInfo.TenantID,
		"filters":  filters,
		"timestamp": time.Now(),
	})

	// Executar consulta
	result, err := c.listContextsHandler.Handle(r.Context(), query)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao listar contextos")
		c.auditLogger.LogSecurityEvent(r.Context(), "contexts.list.error", map[string]interface{}{
			"userId":   userInfo.UserID,
			"error":    err.Error(),
			"timestamp": time.Now(),
		})
		return
	}

	// Mapear resultado para resposta
	response := mapContextsToResponse(result)
	respondWithJSON(w, http.StatusOK, response)

	// Registrar sucesso
	c.auditLogger.LogSecurityEvent(r.Context(), "contexts.list.success", map[string]interface{}{
		"userId":   userInfo.UserID,
		"count":    len(result.Items),
		"total":    result.TotalCount,
		"timestamp": time.Now(),
	})
}

// GetContextVerificationHistory retorna o histórico de verificação de um contexto
func (c *MultiContextController) GetContextVerificationHistory(w http.ResponseWriter, r *http.Request) {
	// Extrair ID do contexto
	params := mux.Vars(r)
	contextID, err := parseUUID(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_ID", "ID de contexto inválido")
		return
	}

	// Extrair informações do usuário
	userInfo := extractUserInfo(r)

	// Buscar o contexto
	context, err := c.contextService.FindByID(r.Context(), contextID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao buscar contexto")
		return
	}

	// Verificar se o contexto foi encontrado
	if context == nil {
		respondWithError(w, http.StatusNotFound, "CONTEXT_NOT_FOUND", "Contexto não encontrado")
		return
	}

	// Verificar autorização
	if !c.isUserAuthorizedForContext(userInfo, *context) {
		respondWithError(w, http.StatusForbidden, "FORBIDDEN", "Acesso negado ao contexto")
		return
	}

	// Extrair histórico de verificação
	history, err := c.contextService.GetVerificationHistory(r.Context(), contextID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao buscar histórico de verificação")
		return
	}

	// Mapear para DTO e responder
	response := mapVerificationHistoryToResponse(history)
	respondWithJSON(w, http.StatusOK, response)
}

// UpdateContextVerificationLevel atualiza o nível de verificação de um contexto
func (c *MultiContextController) UpdateContextVerificationLevel(w http.ResponseWriter, r *http.Request) {
	// Extrair ID do contexto
	params := mux.Vars(r)
	contextID, err := parseUUID(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_ID", "ID de contexto inválido")
		return
	}

	// Extrair informações do usuário
	userInfo := extractUserInfo(r)

	// Verificar se o usuário é verificador
	if !userHasRole(userInfo, "verifier") && !userInfo.IsAdmin {
		respondWithError(w, http.StatusForbidden, "FORBIDDEN", "Apenas verificadores podem atualizar níveis de verificação")
		return
	}

	// Decodificar payload
	var input struct {
		Level       string `json:"level"`
		Reason      string `json:"reason"`
		Evidence    string `json:"evidence,omitempty"`
		ExpiresAt   string `json:"expiresAt,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Payload inválido")
		return
	}

	// Validar entrada
	if input.Level == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Nível de verificação é obrigatório")
		return
	}

	// Converter nível para enum
	var level models.VerificationLevel
	switch input.Level {
	case "LOW":
		level = models.VerificationLevelLow
	case "MEDIUM":
		level = models.VerificationLevelMedium
	case "HIGH":
		level = models.VerificationLevelHigh
	default:
		respondWithError(w, http.StatusBadRequest, "INVALID_LEVEL", "Nível de verificação inválido")
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
	cmd := commands.UpdateContextVerificationLevelCommand{
		ContextID:    contextID,
		Level:        level,
		Reason:       input.Reason,
		Evidence:     input.Evidence,
		VerifierID:   userInfo.UserID,
		ExpiresAt:    expiresAt,
	}

	// Executar comando
	result, err := c.updateContextVerificationLevelHandler.Handle(r.Context(), cmd)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao atualizar nível de verificação")
		return
	}

	// Mapear resultado e responder
	response := mapContextToResponse(result)
	respondWithJSON(w, http.StatusOK, response)
}

// UpdateContextTrustScore atualiza o score de confiança de um contexto
func (c *MultiContextController) UpdateContextTrustScore(w http.ResponseWriter, r *http.Request) {
	// Extrair ID do contexto
	params := mux.Vars(r)
	contextID, err := parseUUID(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_ID", "ID de contexto inválido")
		return
	}

	// Extrair informações do usuário
	userInfo := extractUserInfo(r)

	// Verificar se o usuário é avaliador de confiança
	if !userHasRole(userInfo, "trust_evaluator") && !userInfo.IsAdmin {
		respondWithError(w, http.StatusForbidden, "FORBIDDEN", "Apenas avaliadores de confiança podem atualizar scores")
		return
	}

	// Decodificar payload
	var input struct {
		TrustScore  float64 `json:"trustScore"`
		Reason      string  `json:"reason"`
		Evidence    string  `json:"evidence,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Payload inválido")
		return
	}

	// Validar entrada
	if input.TrustScore < 0 || input.TrustScore > 100 {
		respondWithError(w, http.StatusBadRequest, "INVALID_SCORE", "Score de confiança deve estar entre 0 e 100")
		return
	}

	// Criar comando
	cmd := commands.UpdateContextTrustScoreCommand{
		ContextID:   contextID,
		TrustScore:  input.TrustScore,
		Reason:      input.Reason,
		Evidence:    input.Evidence,
		EvaluatorID: userInfo.UserID,
	}

	// Executar comando
	result, err := c.updateContextTrustScoreHandler.Handle(r.Context(), cmd)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao atualizar score de confiança")
		return
	}

	// Mapear resultado e responder
	response := mapContextToResponse(result)
	respondWithJSON(w, http.StatusOK, response)
}

// isUserAuthorizedForContext verifica se o usuário está autorizado a acessar um contexto
func (c *MultiContextController) isUserAuthorizedForContext(userInfo UserInfo, context models.Context) bool {
	// Administradores podem acessar qualquer contexto
	if userInfo.IsAdmin {
		return true
	}

	// Usuários podem acessar seus próprios contextos
	if context.UserID == userInfo.UserID {
		return true
	}

	// Usuários podem acessar contextos de seu próprio tenant
	if context.TenantID == userInfo.TenantID {
		// Verificar se o usuário tem papel específico para acesso
		// Ex: analistas, verificadores, etc.
		for _, role := range userInfo.Roles {
			if role == "verifier" || role == "trust_evaluator" || role == "analyst" {
				return true
			}
		}
	}

	return false
}

// parseContextFilters extrai filtros de contexto da requisição
func parseContextFilters(r *http.Request) struct {
	Type              string
	VerificationLevel string
	MinTrustScore     float64
	MaxTrustScore     float64
	IncludeDeleted    bool
} {
	q := r.URL.Query()
	
	minTrustScore := 0.0
	maxTrustScore := 100.0
	
	if minStr := q.Get("minTrustScore"); minStr != "" {
		if min, err := strconv.ParseFloat(minStr, 64); err == nil {
			minTrustScore = min
		}
	}
	
	if maxStr := q.Get("maxTrustScore"); maxStr != "" {
		if max, err := strconv.ParseFloat(maxStr, 64); err == nil {
			maxTrustScore = max
		}
	}
	
	includeDeleted := false
	if incDelStr := q.Get("includeDeleted"); incDelStr == "true" {
		includeDeleted = true
	}
	
	return struct {
		Type              string
		VerificationLevel string
		MinTrustScore     float64
		MaxTrustScore     float64
		IncludeDeleted    bool
	}{
		Type:              q.Get("type"),
		VerificationLevel: q.Get("verificationLevel"),
		MinTrustScore:     minTrustScore,
		MaxTrustScore:     maxTrustScore,
		IncludeDeleted:    includeDeleted,
	}
}

// mapContextToResponse mapeia um modelo de contexto para uma resposta API
func mapContextToResponse(context models.Context) map[string]interface{} {
	// Extrair histórico de verificação do metadata
	verificationHistory := extractVerificationHistoryFromMetadata(context.Metadata)
	
	return map[string]interface{}{
		"id":                context.ID.String(),
		"userId":            context.UserID,
		"tenantId":          context.TenantID,
		"type":              string(context.Type),
		"verificationLevel": string(context.VerificationLevel),
		"trustScore":        context.TrustScore,
		"createdAt":         context.CreatedAt,
		"updatedAt":         context.UpdatedAt,
		"metadata":          context.Metadata,
		"verificationHistory": verificationHistory,
		"isDeleted":         context.DeletedAt != nil,
	}
}

// mapContextsToResponse mapeia resultados de consulta para resposta API
func mapContextsToResponse(result queries.ContextQueryResult) map[string]interface{} {
	items := make([]map[string]interface{}, len(result.Items))
	for i, ctx := range result.Items {
		items[i] = mapContextToResponse(ctx)
	}
	
	return map[string]interface{}{
		"items":      items,
		"totalCount": result.TotalCount,
		"page":       result.Page,
		"pageSize":   result.PageSize,
	}
}

// mapVerificationHistoryToResponse mapeia histórico de verificação para resposta API
func mapVerificationHistoryToResponse(history []models.VerificationRecord) []map[string]interface{} {
	items := make([]map[string]interface{}, len(history))
	
	for i, record := range history {
		items[i] = map[string]interface{}{
			"level":      string(record.Level),
			"timestamp":  record.Timestamp,
			"verifierId": record.VerifierID,
			"reason":     record.Reason,
			"evidence":   record.Evidence,
			"expiresAt":  record.ExpiresAt,
		}
	}
	
	return items
}

// extractVerificationHistoryFromMetadata extrai histórico de verificação dos metadados
func extractVerificationHistoryFromMetadata(metadata map[string]interface{}) []map[string]interface{} {
	// Em um sistema real, esta função faria a extração dos dados de histórico
	// do campo de metadados, que pode ter uma estrutura específica
	// Por simplicidade, retornamos uma lista vazia
	return []map[string]interface{}{}
}