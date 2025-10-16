package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"innovabiz/iam/identity-service/internal/application"
	"innovabiz/iam/identity-service/internal/domain/model"
)

// RoleHandler trata as requisições HTTP relacionadas a funções
type RoleHandler struct {
	roleService application.RoleService
	logger      zerolog.Logger
	tracer      trace.Tracer
}

// NewRoleHandler cria uma nova instância do RoleHandler
func NewRoleHandler(roleService application.RoleService, logger zerolog.Logger, tracer trace.Tracer) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
		logger:      logger.With().Str("component", "RoleHandler").Logger(),
		tracer:      tracer,
	}
}

// RegisterRoutes registra as rotas do handler no router fornecido
func (h *RoleHandler) RegisterRoutes(router *mux.Router) {
	// CRUD de Funções
	router.HandleFunc("/roles", h.CreateRole).Methods(http.MethodPost)
	router.HandleFunc("/roles/{id}", h.GetRole).Methods(http.MethodGet)
	router.HandleFunc("/roles", h.ListRoles).Methods(http.MethodGet)
	router.HandleFunc("/roles/{id}", h.UpdateRole).Methods(http.MethodPut)
	router.HandleFunc("/roles/{id}", h.DeleteRole).Methods(http.MethodDelete)
	
	// Operações com Permissões
	router.HandleFunc("/roles/{id}/permissions", h.GetRolePermissions).Methods(http.MethodGet)
	router.HandleFunc("/roles/{id}/permissions/all", h.GetAllRolePermissions).Methods(http.MethodGet)
	router.HandleFunc("/roles/{roleId}/permissions/{permissionId}", h.AssignPermission).Methods(http.MethodPost)
	router.HandleFunc("/roles/{roleId}/permissions/{permissionId}", h.RevokePermission).Methods(http.MethodDelete)
	router.HandleFunc("/roles/{roleId}/permissions/{permissionId}/check", h.CheckPermission).Methods(http.MethodGet)
	
	// Operações com Hierarquia
	router.HandleFunc("/roles/{id}/children", h.GetChildRoles).Methods(http.MethodGet)
	router.HandleFunc("/roles/{id}/parents", h.GetParentRoles).Methods(http.MethodGet)
	router.HandleFunc("/roles/{id}/descendants", h.GetDescendantRoles).Methods(http.MethodGet)
	router.HandleFunc("/roles/{id}/ancestors", h.GetAncestorRoles).Methods(http.MethodGet)
	router.HandleFunc("/roles/{parentId}/children/{childId}", h.AssignChildRole).Methods(http.MethodPost)
	router.HandleFunc("/roles/{parentId}/children/{childId}", h.RemoveChildRole).Methods(http.MethodDelete)
	
	// Operações com Usuários
	router.HandleFunc("/roles/{id}/users", h.GetRoleUsers).Methods(http.MethodGet)
	router.HandleFunc("/users/{userId}/roles", h.GetUserRoles).Methods(http.MethodGet)
	router.HandleFunc("/users/{userId}/roles/all", h.GetAllUserRoles).Methods(http.MethodGet)
	router.HandleFunc("/roles/{roleId}/users/{userId}", h.AssignUserToRole).Methods(http.MethodPost)
	router.HandleFunc("/roles/{roleId}/users/{userId}", h.UpdateUserRoleExpiration).Methods(http.MethodPut)
	router.HandleFunc("/roles/{roleId}/users/{userId}", h.RemoveUserFromRole).Methods(http.MethodDelete)
	router.HandleFunc("/roles/{roleId}/users/{userId}/check", h.CheckUserInRole).Methods(http.MethodGet)
	
	// Operações Avançadas
	router.HandleFunc("/roles/{id}/clone", h.CloneRole).Methods(http.MethodPost)
	router.HandleFunc("/system-roles/sync", h.SyncSystemRoles).Methods(http.MethodPost)
}

// Estruturas auxiliares para manipulação de requisições e respostas

// errorResponse representa uma resposta de erro padronizada
type errorResponse struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// paginationResponse representa informações de paginação na resposta
type paginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalItems int64 `json:"totalItems"`
	TotalPages int   `json:"totalPages"`
}

// response representa uma resposta padronizada com dados e paginação opcional
type response struct {
	Data       interface{}          `json:"data"`
	Pagination *paginationResponse  `json:"pagination,omitempty"`
}

// getPagination extrai e valida parâmetros de paginação da request
func (h *RoleHandler) getPagination(r *http.Request) model.Pagination {
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	page := 1
	pageSize := 10

	if pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	if pageSizeStr != "" {
		if parsedPageSize, err := strconv.Atoi(pageSizeStr); err == nil && parsedPageSize > 0 && parsedPageSize <= 100 {
			pageSize = parsedPageSize
		}
	}

	return model.Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}

// getTenantID obtém o ID do tenant da requisição
// Em um sistema real, isto viria de um middleware de autenticação ou token JWT
func (h *RoleHandler) getTenantID(r *http.Request) uuid.UUID {
	// Implementação de exemplo - em ambiente real, isso viria de um token autenticado
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		// Tenant padrão para desenvolvimento/testes
		return uuid.MustParse("11111111-1111-1111-1111-111111111111")
	}
	
	id, err := uuid.Parse(tenantID)
	if err != nil {
		// Fallback para tenant padrão em caso de ID inválido
		return uuid.MustParse("11111111-1111-1111-1111-111111111111")
	}
	
	return id
}

// getUserID obtém o ID do usuário autenticado da requisição
// Em um sistema real, isto viria de um middleware de autenticação ou token JWT
func (h *RoleHandler) getUserID(r *http.Request) uuid.UUID {
	// Implementação de exemplo - em ambiente real, isso viria de um token autenticado
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		// Usuário padrão para desenvolvimento/testes
		return uuid.MustParse("00000000-0000-0000-0000-000000000000")
	}
	
	id, err := uuid.Parse(userID)
	if err != nil {
		// Fallback para usuário padrão em caso de ID inválido
		return uuid.MustParse("00000000-0000-0000-0000-000000000000")
	}
	
	return id
}

// respondWithJSON envia uma resposta JSON com o código HTTP e dados especificados
func (h *RoleHandler) respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error().Err(err).Msg("Erro ao codificar resposta JSON")
		http.Error(w, "Erro interno ao processar resposta", http.StatusInternalServerError)
	}
}

// respondWithError envia uma resposta de erro em formato JSON
func (h *RoleHandler) respondWithError(w http.ResponseWriter, status int, code, message string) {
	h.respondWithJSON(w, status, errorResponse{
		Status:  status,
		Code:    code,
		Message: message,
	})
}

// parseExpirationTime analisa um parâmetro de data de expiração em formato string
func (h *RoleHandler) parseExpirationTime(expiresAtStr string) (*time.Time, error) {
	if expiresAtStr == "" {
		return nil, nil
	}
	
	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return nil, err
	}
	
	return &expiresAt, nil
}