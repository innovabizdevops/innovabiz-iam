/**
 * @file controllers.go
 * @description Controladores REST para o serviço de identidade multi-contexto
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package rest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"innovabiz/iam/src/multi-context/application/commands"
	"innovabiz/iam/src/multi-context/application/queries"
	"innovabiz/iam/src/multi-context/domain/models"
	"innovabiz/iam/src/multi-context/domain/services"
)

// MultiContextController fornece handlers HTTP para operações de identidade multi-contexto
type MultiContextController struct {
	// Query handlers
	listContextsHandler     *queries.ListContextsHandler
	listAttributesHandler   *queries.ListAttributesHandler
	searchAttributesHandler *queries.SearchAttributesHandler

	// Command handlers
	createAttributeHandler              *commands.CreateAttributeHandler
	updateAttributeHandler              *commands.UpdateAttributeHandler
	verifyAttributeHandler              *commands.VerifyAttributeHandler
	updateContextVerificationLevelHandler *commands.UpdateContextVerificationLevelHandler
	updateContextTrustScoreHandler      *commands.UpdateContextTrustScoreHandler

	// Services
	contextService   *services.ContextService
	attributeService *services.AttributeService
	auditLogger      services.AuditLogger
}

// NewMultiContextController cria uma nova instância do controlador
func NewMultiContextController(
	// Query handlers
	listContextsHandler *queries.ListContextsHandler,
	listAttributesHandler *queries.ListAttributesHandler,
	searchAttributesHandler *queries.SearchAttributesHandler,

	// Command handlers
	createAttributeHandler *commands.CreateAttributeHandler,
	updateAttributeHandler *commands.UpdateAttributeHandler,
	verifyAttributeHandler *commands.VerifyAttributeHandler,
	updateContextVerificationLevelHandler *commands.UpdateContextVerificationLevelHandler,
	updateContextTrustScoreHandler *commands.UpdateContextTrustScoreHandler,

	// Services
	contextService *services.ContextService,
	attributeService *services.AttributeService,
	auditLogger services.AuditLogger,
) *MultiContextController {
	return &MultiContextController{
		listContextsHandler:     listContextsHandler,
		listAttributesHandler:   listAttributesHandler,
		searchAttributesHandler: searchAttributesHandler,

		createAttributeHandler:              createAttributeHandler,
		updateAttributeHandler:              updateAttributeHandler,
		verifyAttributeHandler:              verifyAttributeHandler,
		updateContextVerificationLevelHandler: updateContextVerificationLevelHandler,
		updateContextTrustScoreHandler:      updateContextTrustScoreHandler,

		contextService:   contextService,
		attributeService: attributeService,
		auditLogger:      auditLogger,
	}
}

// ResponseError representa um erro padronizado para respostas API
type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ErrorResponse encapsula um erro para resposta HTTP
type ErrorResponse struct {
	Error ResponseError `json:"error"`
}

// PaginationParams representa parâmetros de paginação
type PaginationParams struct {
	Page     int
	PageSize int
	SortBy   string
	SortDir  string
}

// extractUserInfo extrai informações do usuário da requisição
func extractUserInfo(r *http.Request) UserInfo {
	// Em um sistema real, isso seria implementado para extrair informações
	// do token JWT ou outra fonte de autenticação
	userID := r.Header.Get("X-User-ID")
	tenantID := r.Header.Get("X-Tenant-ID")
	
	return UserInfo{
		UserID:   userID,
		TenantID: tenantID,
		IsAdmin:  r.Header.Get("X-User-Role") == "admin",
		Roles:    parseRoles(r.Header.Get("X-User-Roles")),
	}
}

// parseRoles converte uma string de papéis separados por vírgula em uma slice
func parseRoles(rolesStr string) []string {
	if rolesStr == "" {
		return nil
	}
	
	// Em um sistema real, isso seria mais robusto
	// Por simplicidade, assumimos uma lista separada por vírgula
	return strings.Split(rolesStr, ",")
}

// UserInfo representa informações do usuário autenticado
type UserInfo struct {
	UserID   string
	TenantID string
	IsAdmin  bool
	Roles    []string
}

// userHasRole verifica se o usuário possui o papel especificado
func userHasRole(userInfo UserInfo, role string) bool {
	if userInfo.IsAdmin {
		return true
	}
	
	for _, r := range userInfo.Roles {
		if r == role {
			return true
		}
	}
	
	return false
}

// parsePaginationParams extrai parâmetros de paginação da requisição
func parsePaginationParams(r *http.Request) PaginationParams {
	// Valores padrão
	params := PaginationParams{
		Page:     0,
		PageSize: 20,
		SortBy:   "created_at",
		SortDir:  "desc",
	}
	
	// Extrair da query string
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page >= 0 {
			params.Page = page
		}
	}
	
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			params.PageSize = pageSize
		}
	}
	
	if sortBy := r.URL.Query().Get("sortBy"); sortBy != "" {
		params.SortBy = sortBy
	}
	
	if sortDir := r.URL.Query().Get("sortDir"); sortDir != "" {
		params.SortDir = sortDir
	}
	
	return params
}

// respondWithJSON responde com um payload JSON e status HTTP
func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao serializar resposta")
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

// respondWithError responde com um erro JSON padronizado
func respondWithError(w http.ResponseWriter, status int, code string, message string) {
	respondWithJSON(w, status, ErrorResponse{
		Error: ResponseError{
			Code:    code,
			Message: message,
		},
	})
}

// respondWithValidationError responde com um erro de validação detalhado
func respondWithValidationError(w http.ResponseWriter, message string, details map[string]interface{}) {
	respondWithJSON(w, http.StatusBadRequest, ErrorResponse{
		Error: ResponseError{
			Code:    "VALIDATION_ERROR",
			Message: message,
			Details: details,
		},
	})
}

// parseUUID converte um ID de string para UUID e retorna erro se inválido
func parseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}