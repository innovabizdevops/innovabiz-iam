/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Handler HTTP para gerenciamento de funções (roles).
 * Define endpoints REST para CRUD de funções e operações relacionadas.
 */

package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/innovabiz/iam/services/identity-service/internal/application"
	"github.com/innovabiz/iam/services/identity-service/internal/infrastructure/middleware"
)

// RoleHandler gerencia endpoints HTTP para operações de funções
type RoleHandler struct {
	roleService application.RoleService
	validator   *validator.Validate
	tracer      trace.Tracer
}

// NewRoleHandler cria uma nova instância de RoleHandler
func NewRoleHandler(roleService application.RoleService) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
		validator:   validator.New(),
		tracer:      otel.Tracer("role-handler"),
	}
}

// RegisterRoutes registra as rotas do handler no router fornecido
func (h *RoleHandler) RegisterRoutes(r *mux.Router, authMiddleware *middleware.AuthMiddleware) {
	// Grupo de rotas para funções
	rolesRouter := r.PathPrefix("/v1/roles").Subrouter()
	
	// Aplicar middleware de autenticação em todas as rotas
	rolesRouter.Use(authMiddleware.Authenticate)
	
	// Rotas básicas CRUD
	rolesRouter.HandleFunc("", h.ListRoles).Methods(http.MethodGet)
	rolesRouter.HandleFunc("", h.CreateRole).Methods(http.MethodPost)
	rolesRouter.HandleFunc("/{id}", h.GetRoleByID).Methods(http.MethodGet)
	rolesRouter.HandleFunc("/{id}", h.UpdateRole).Methods(http.MethodPut)
	rolesRouter.HandleFunc("/{id}", h.DeleteRole).Methods(http.MethodDelete)
	
	// Rotas para busca por código
	rolesRouter.HandleFunc("/code/{code}", h.GetRoleByCode).Methods(http.MethodGet)
	
	// Rotas para gerenciamento de permissões em funções
	rolesRouter.HandleFunc("/{id}/permissions", h.GetRolePermissions).Methods(http.MethodGet)
	rolesRouter.HandleFunc("/{id}/permissions", h.AssignPermissionsToRole).Methods(http.MethodPost)
	rolesRouter.HandleFunc("/{id}/permissions", h.RevokePermissionsFromRole).Methods(http.MethodDelete)
	
	// Rotas para gerenciamento de usuários em funções
	rolesRouter.HandleFunc("/{id}/users", h.GetRoleUsers).Methods(http.MethodGet)
	rolesRouter.HandleFunc("/{id}/users", h.AssignRoleToUsers).Methods(http.MethodPost)
	rolesRouter.HandleFunc("/{id}/users", h.RevokeRoleFromUsers).Methods(http.MethodDelete)
	
	// Proteger rotas de administração com middleware de verificação de permissões
	adminRouter := rolesRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(authMiddleware.RequirePermissions([]string{"roles:admin:manage"}))
	adminRouter.HandleFunc("/sync", h.SyncSystemRoles).Methods(http.MethodPost)
}

// CreateRole cria uma nova função
func (h *RoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.CreateRole")
	defer span.End()
	
	tenantID, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter tenant ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "Tenant ID não encontrado")
		return
	}
	
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter user ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "User ID não encontrado")
		return
	}
	
	// Adicionar informações ao span do tracer
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("user_id", userID.String()),
	)
	
	var req application.CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Erro ao decodificar corpo da requisição")
		middleware.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido")
		return
	}
	
	// Validar a requisição
	if err := h.validator.Struct(req); err != nil {
		log.Error().Err(err).Interface("request", req).Msg("Requisição inválida")
		middleware.RespondWithValidationErrors(w, err)
		return
	}
	
	// Definir o tenant ID da requisição com o valor do contexto
	req.TenantID = tenantID
	
	// Chamar o serviço para criar a função
	resp, err := h.roleService.Create(ctx, &req)
	if err != nil {
		log.Error().Err(err).Interface("request", req).Msg("Erro ao criar função")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusCreated, resp)
}

// GetRoleByID recupera uma função pelo seu ID
func (h *RoleHandler) GetRoleByID(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.GetRoleByID")
	defer span.End()
	
	tenantID, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter tenant ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "Tenant ID não encontrado")
		return
	}
	
	// Extrair ID da URL
	vars := mux.Vars(r)
	roleIDStr := vars["id"]
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		log.Error().Err(err).Str("role_id", roleIDStr).Msg("ID de função inválido")
		middleware.RespondWithError(w, http.StatusBadRequest, "ID de função inválido")
		return
	}
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("role_id", roleID.String()),
	)
	
	// Chamar o serviço para recuperar a função
	resp, err := h.roleService.GetByID(ctx, tenantID, roleID)
	if err != nil {
		log.Error().Err(err).Str("role_id", roleID.String()).Msg("Erro ao recuperar função")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusOK, resp)
}

// GetRoleByCode recupera uma função pelo seu código
func (h *RoleHandler) GetRoleByCode(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.GetRoleByCode")
	defer span.End()
	
	tenantID, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter tenant ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "Tenant ID não encontrado")
		return
	}
	
	// Extrair código da URL
	vars := mux.Vars(r)
	code := vars["code"]
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("role_code", code),
	)
	
	// Chamar o serviço para recuperar a função
	resp, err := h.roleService.GetByCode(ctx, tenantID, code)
	if err != nil {
		log.Error().Err(err).Str("code", code).Msg("Erro ao recuperar função por código")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusOK, resp)
}

// ListRoles lista funções com filtros e paginação
func (h *RoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.ListRoles")
	defer span.End()
	
	tenantID, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter tenant ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "Tenant ID não encontrado")
		return
	}
	
	// Extrair parâmetros de consulta
	query := r.URL.Query()
	
	// Parâmetros de paginação
	page, _ := strconv.Atoi(query.Get("page"))
	if page <= 0 {
		page = 1
	}
	
	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20 // Valor padrão
	}
	
	// Filtros
	isActiveStr := query.Get("is_active")
	var isActive *bool
	if isActiveStr != "" {
		active := isActiveStr == "true"
		isActive = &active
	}
	
	req := &application.ListRolesRequest{
		TenantID:   tenantID,
		Page:       page,
		PageSize:   pageSize,
		Code:       query.Get("code"),
		Name:       query.Get("name"),
		Type:       query.Get("type"),
		IsActive:   isActive,
		SearchTerm: query.Get("search"),
		OrderBy:    query.Get("order_by"),
		Order:      query.Get("order"),
	}
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.Int("page", page),
		attribute.Int("page_size", pageSize),
		attribute.String("search_term", req.SearchTerm),
	)
	
	// Chamar o serviço para listar funções
	resp, err := h.roleService.List(ctx, req)
	if err != nil {
		log.Error().Err(err).Interface("request", req).Msg("Erro ao listar funções")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusOK, resp)
}

// UpdateRole atualiza uma função existente
func (h *RoleHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.UpdateRole")
	defer span.End()
	
	tenantID, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter tenant ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "Tenant ID não encontrado")
		return
	}
	
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter user ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "User ID não encontrado")
		return
	}
	
	// Extrair ID da URL
	vars := mux.Vars(r)
	roleIDStr := vars["id"]
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		log.Error().Err(err).Str("role_id", roleIDStr).Msg("ID de função inválido")
		middleware.RespondWithError(w, http.StatusBadRequest, "ID de função inválido")
		return
	}
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("user_id", userID.String()),
		attribute.String("role_id", roleID.String()),
	)
	
	var req application.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Erro ao decodificar corpo da requisição")
		middleware.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido")
		return
	}
	
	// Validar a requisição
	if err := h.validator.Struct(req); err != nil {
		log.Error().Err(err).Interface("request", req).Msg("Requisição inválida")
		middleware.RespondWithValidationErrors(w, err)
		return
	}
	
	// Definir IDs da requisição com os valores do contexto e da URL
	req.TenantID = tenantID
	req.ID = roleID
	
	// Chamar o serviço para atualizar a função
	resp, err := h.roleService.Update(ctx, &req)
	if err != nil {
		log.Error().Err(err).Interface("request", req).Msg("Erro ao atualizar função")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusOK, resp)
}

// DeleteRole exclui uma função
func (h *RoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.DeleteRole")
	defer span.End()
	
	tenantID, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter tenant ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "Tenant ID não encontrado")
		return
	}
	
	// Extrair ID da URL
	vars := mux.Vars(r)
	roleIDStr := vars["id"]
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		log.Error().Err(err).Str("role_id", roleIDStr).Msg("ID de função inválido")
		middleware.RespondWithError(w, http.StatusBadRequest, "ID de função inválido")
		return
	}
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("role_id", roleID.String()),
	)
	
	// Extrair parâmetros de consulta
	query := r.URL.Query()
	force := query.Get("force") == "true"
	
	// Chamar o serviço para excluir a função
	err = h.roleService.Delete(ctx, tenantID, roleID, force)
	if err != nil {
		log.Error().Err(err).Str("role_id", roleID.String()).Bool("force", force).Msg("Erro ao excluir função")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusNoContent, nil)
}

// GetRolePermissions recupera as permissões de uma função
func (h *RoleHandler) GetRolePermissions(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.GetRolePermissions")
	defer span.End()
	
	tenantID, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter tenant ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "Tenant ID não encontrado")
		return
	}
	
	// Extrair ID da URL
	vars := mux.Vars(r)
	roleIDStr := vars["id"]
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		log.Error().Err(err).Str("role_id", roleIDStr).Msg("ID de função inválido")
		middleware.RespondWithError(w, http.StatusBadRequest, "ID de função inválido")
		return
	}
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("role_id", roleID.String()),
	)
	
	// Chamar o serviço para recuperar as permissões da função
	resp, err := h.roleService.GetRolePermissions(ctx, tenantID, roleID)
	if err != nil {
		log.Error().Err(err).Str("role_id", roleID.String()).Msg("Erro ao recuperar permissões da função")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusOK, resp)
}

// AssignPermissionsToRole atribui permissões a uma função
func (h *RoleHandler) AssignPermissionsToRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.AssignPermissionsToRole")
	defer span.End()
	
	tenantID, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter tenant ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "Tenant ID não encontrado")
		return
	}
	
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter user ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "User ID não encontrado")
		return
	}
	
	// Extrair ID da URL
	vars := mux.Vars(r)
	roleIDStr := vars["id"]
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		log.Error().Err(err).Str("role_id", roleIDStr).Msg("ID de função inválido")
		middleware.RespondWithError(w, http.StatusBadRequest, "ID de função inválido")
		return
	}
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("user_id", userID.String()),
		attribute.String("role_id", roleID.String()),
	)
	
	var req application.AssignPermissionsToRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Erro ao decodificar corpo da requisição")
		middleware.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido")
		return
	}
	
	// Validar a requisição
	if err := h.validator.Struct(req); err != nil {
		log.Error().Err(err).Interface("request", req).Msg("Requisição inválida")
		middleware.RespondWithValidationErrors(w, err)
		return
	}
	
	// Definir IDs da requisição com os valores do contexto e da URL
	req.TenantID = tenantID
	req.RoleID = roleID
	
	// Chamar o serviço para atribuir permissões à função
	err = h.roleService.AssignPermissionsToRole(ctx, &req)
	if err != nil {
		log.Error().Err(err).Interface("request", req).Msg("Erro ao atribuir permissões à função")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	// Recuperar as permissões atualizadas da função para retornar na resposta
	resp, err := h.roleService.GetRolePermissions(ctx, tenantID, roleID)
	if err != nil {
		log.Error().Err(err).Str("role_id", roleID.String()).Msg("Erro ao recuperar permissões da função após atribuição")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusOK, resp)
}// RevokePermissionsFromRole revoga permissões de uma função
func (h *RoleHandler) RevokePermissionsFromRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.RevokePermissionsFromRole")
	defer span.End()
	
	tenantID, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter tenant ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "Tenant ID não encontrado")
		return
	}
	
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter user ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "User ID não encontrado")
		return
	}
	
	// Extrair ID da URL
	vars := mux.Vars(r)
	roleIDStr := vars["id"]
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		log.Error().Err(err).Str("role_id", roleIDStr).Msg("ID de função inválido")
		middleware.RespondWithError(w, http.StatusBadRequest, "ID de função inválido")
		return
	}
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("user_id", userID.String()),
		attribute.String("role_id", roleID.String()),
	)
	
	var req application.RevokePermissionsFromRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Erro ao decodificar corpo da requisição")
		middleware.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido")
		return
	}
	
	// Validar a requisição
	if err := h.validator.Struct(req); err != nil {
		log.Error().Err(err).Interface("request", req).Msg("Requisição inválida")
		middleware.RespondWithValidationErrors(w, err)
		return
	}
	
	// Definir IDs da requisição com os valores do contexto e da URL
	req.TenantID = tenantID
	req.RoleID = roleID
	
	// Chamar o serviço para revogar permissões da função
	err = h.roleService.RevokePermissionsFromRole(ctx, &req)
	if err != nil {
		log.Error().Err(err).Interface("request", req).Msg("Erro ao revogar permissões da função")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	// Recuperar as permissões atualizadas da função para retornar na resposta
	resp, err := h.roleService.GetRolePermissions(ctx, tenantID, roleID)
	if err != nil {
		log.Error().Err(err).Str("role_id", roleID.String()).Msg("Erro ao recuperar permissões da função após revogação")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusOK, resp)
}// GetRoleUsers recupera os usuários atribuídos a uma função
func (h *RoleHandler) GetRoleUsers(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.GetRoleUsers")
	defer span.End()
	
	tenantID, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter tenant ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "Tenant ID não encontrado")
		return
	}
	
	// Extrair ID da URL
	vars := mux.Vars(r)
	roleIDStr := vars["id"]
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		log.Error().Err(err).Str("role_id", roleIDStr).Msg("ID de função inválido")
		middleware.RespondWithError(w, http.StatusBadRequest, "ID de função inválido")
		return
	}
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("role_id", roleID.String()),
	)
	
	// Extrair parâmetros de consulta para paginação
	query := r.URL.Query()
	
	// Parâmetros de paginação
	page, _ := strconv.Atoi(query.Get("page"))
	if page <= 0 {
		page = 1
	}
	
	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20 // Valor padrão
	}
	
	// Criar requisição para o serviço
	req := &application.GetRoleUsersRequest{
		TenantID: tenantID,
		RoleID:   roleID,
		Page:     page,
		PageSize: pageSize,
	}
	
	// Chamar o serviço para recuperar os usuários da função
	resp, err := h.roleService.GetRoleUsers(ctx, req)
	if err != nil {
		log.Error().Err(err).Interface("request", req).Msg("Erro ao recuperar usuários da função")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusOK, resp)
}// AssignRoleToUsers atribui uma função a múltiplos usuários
func (h *RoleHandler) AssignRoleToUsers(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.AssignRoleToUsers")
	defer span.End()
	
	tenantID, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter tenant ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "Tenant ID não encontrado")
		return
	}
	
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter user ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "User ID não encontrado")
		return
	}
	
	// Extrair ID da URL
	vars := mux.Vars(r)
	roleIDStr := vars["id"]
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		log.Error().Err(err).Str("role_id", roleIDStr).Msg("ID de função inválido")
		middleware.RespondWithError(w, http.StatusBadRequest, "ID de função inválido")
		return
	}
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("user_id", userID.String()),
		attribute.String("role_id", roleID.String()),
	)
	
	var req application.AssignRoleToUsersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Erro ao decodificar corpo da requisição")
		middleware.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido")
		return
	}
	
	// Validar a requisição
	if err := h.validator.Struct(req); err != nil {
		log.Error().Err(err).Interface("request", req).Msg("Requisição inválida")
		middleware.RespondWithValidationErrors(w, err)
		return
	}
	
	// Definir IDs da requisição com os valores do contexto e da URL
	req.TenantID = tenantID
	req.RoleID = roleID
	
	// Chamar o serviço para atribuir a função aos usuários
	err = h.roleService.AssignRoleToUsers(ctx, &req)
	if err != nil {
		log.Error().Err(err).Interface("request", req).Msg("Erro ao atribuir função aos usuários")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "Função atribuída aos usuários com sucesso",
		"role_id":    roleID.String(),
		"user_count": len(req.UserIDs),
	})
}// RevokeRoleFromUsers revoga uma função de múltiplos usuários
func (h *RoleHandler) RevokeRoleFromUsers(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.RevokeRoleFromUsers")
	defer span.End()
	
	tenantID, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter tenant ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "Tenant ID não encontrado")
		return
	}
	
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter user ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "User ID não encontrado")
		return
	}
	
	// Extrair ID da URL
	vars := mux.Vars(r)
	roleIDStr := vars["id"]
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		log.Error().Err(err).Str("role_id", roleIDStr).Msg("ID de função inválido")
		middleware.RespondWithError(w, http.StatusBadRequest, "ID de função inválido")
		return
	}
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("user_id", userID.String()),
		attribute.String("role_id", roleID.String()),
	)
	
	var req application.RevokeRoleFromUsersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Erro ao decodificar corpo da requisição")
		middleware.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido")
		return
	}
	
	// Validar a requisição
	if err := h.validator.Struct(req); err != nil {
		log.Error().Err(err).Interface("request", req).Msg("Requisição inválida")
		middleware.RespondWithValidationErrors(w, err)
		return
	}
	
	// Definir IDs da requisição com os valores do contexto e da URL
	req.TenantID = tenantID
	req.RoleID = roleID
	
	// Chamar o serviço para revogar a função dos usuários
	err = h.roleService.RevokeRoleFromUsers(ctx, &req)
	if err != nil {
		log.Error().Err(err).Interface("request", req).Msg("Erro ao revogar função dos usuários")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "Função revogada dos usuários com sucesso",
		"role_id":    roleID.String(),
		"user_count": len(req.UserIDs),
	})
}// SyncSystemRoles sincroniza as funções de sistema a partir da configuração
func (h *RoleHandler) SyncSystemRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.SyncSystemRoles")
	defer span.End()
	
	tenantID, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter tenant ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "Tenant ID não encontrado")
		return
	}
	
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter user ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "User ID não encontrado")
		return
	}
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("user_id", userID.String()),
	)
	
	// Criar requisição para o serviço
	req := &application.SyncSystemRolesRequest{
		TenantID: tenantID,
		UserID:   userID,
	}
	
	// Chamar o serviço para sincronizar as funções do sistema
	resp, err := h.roleService.SyncSystemRoles(ctx, req)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao sincronizar funções de sistema")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusOK, resp)
}