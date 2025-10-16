/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Handler HTTP para gerenciamento de permissões.
 * Define endpoints REST para CRUD de permissões e operações relacionadas.
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

// PermissionHandler gerencia endpoints HTTP para operações de permissões
type PermissionHandler struct {
	permissionService application.PermissionService
	validator         *validator.Validate
	tracer            trace.Tracer
}

// NewPermissionHandler cria uma nova instância de PermissionHandler
func NewPermissionHandler(permissionService application.PermissionService) *PermissionHandler {
	return &PermissionHandler{
		permissionService: permissionService,
		validator:         validator.New(),
		tracer:            otel.Tracer("permission-handler"),
	}
}

// RegisterRoutes registra as rotas do handler no router fornecido
func (h *PermissionHandler) RegisterRoutes(r *mux.Router, authMiddleware *middleware.AuthMiddleware) {
	// Grupo de rotas para permissões
	permissionsRouter := r.PathPrefix("/v1/permissions").Subrouter()
	
	// Aplicar middleware de autenticação em todas as rotas
	permissionsRouter.Use(authMiddleware.Authenticate)
	
	// Rotas básicas CRUD
	permissionsRouter.HandleFunc("", h.ListPermissions).Methods(http.MethodGet)
	permissionsRouter.HandleFunc("", h.CreatePermission).Methods(http.MethodPost)
	permissionsRouter.HandleFunc("/{id}", h.GetPermissionByID).Methods(http.MethodGet)
	permissionsRouter.HandleFunc("/{id}", h.UpdatePermission).Methods(http.MethodPut)
	permissionsRouter.HandleFunc("/{id}", h.DeletePermission).Methods(http.MethodDelete)
	
	// Rotas para busca por código
	permissionsRouter.HandleFunc("/code/{code}", h.GetPermissionByCode).Methods(http.MethodGet)
	
	// Rotas para atribuição de permissões a funções
	permissionsRouter.HandleFunc("/{id}/roles", h.GetPermissionRoles).Methods(http.MethodGet)
	
	// Rotas para gerenciamento de permissões por módulo, recurso ou ação
	permissionsRouter.HandleFunc("/modules/{module}", h.GetPermissionsByModule).Methods(http.MethodGet)
	permissionsRouter.HandleFunc("/resources/{resource}", h.GetPermissionsByResource).Methods(http.MethodGet)
	permissionsRouter.HandleFunc("/actions/{action}", h.GetPermissionsByAction).Methods(http.MethodGet)
	
	// Rota para verificação de permissão
	permissionsRouter.HandleFunc("/check/{code}", h.CheckUserPermission).Methods(http.MethodGet)
	
	// Proteger rotas de administração com middleware de verificação de permissões
	adminRouter := permissionsRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(authMiddleware.RequirePermissions([]string{"permissions:admin:manage"}))
	adminRouter.HandleFunc("/sync", h.SyncSystemPermissions).Methods(http.MethodPost)
}

// CreatePermission cria uma nova permissão
func (h *PermissionHandler) CreatePermission(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PermissionHandler.CreatePermission")
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
	
	var req application.CreatePermissionRequest
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
	
	// Chamar o serviço para criar a permissão
	resp, err := h.permissionService.Create(ctx, &req)
	if err != nil {
		log.Error().Err(err).Interface("request", req).Msg("Erro ao criar permissão")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusCreated, resp)
}

// GetPermissionByID recupera uma permissão pelo seu ID
func (h *PermissionHandler) GetPermissionByID(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PermissionHandler.GetPermissionByID")
	defer span.End()
	
	tenantID, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter tenant ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "Tenant ID não encontrado")
		return
	}
	
	// Extrair ID da URL
	vars := mux.Vars(r)
	permissionIDStr := vars["id"]
	
	permissionID, err := uuid.Parse(permissionIDStr)
	if err != nil {
		log.Error().Err(err).Str("permission_id", permissionIDStr).Msg("ID de permissão inválido")
		middleware.RespondWithError(w, http.StatusBadRequest, "ID de permissão inválido")
		return
	}
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("permission_id", permissionID.String()),
	)
	
	// Chamar o serviço para recuperar a permissão
	resp, err := h.permissionService.GetByID(ctx, tenantID, permissionID)
	if err != nil {
		log.Error().Err(err).Str("permission_id", permissionID.String()).Msg("Erro ao recuperar permissão")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusOK, resp)
}

// GetPermissionByCode recupera uma permissão pelo seu código
func (h *PermissionHandler) GetPermissionByCode(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PermissionHandler.GetPermissionByCode")
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
		attribute.String("permission_code", code),
	)
	
	// Chamar o serviço para recuperar a permissão
	resp, err := h.permissionService.GetByCode(ctx, tenantID, code)
	if err != nil {
		log.Error().Err(err).Str("code", code).Msg("Erro ao recuperar permissão por código")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusOK, resp)
}

// ListPermissions lista permissões com filtros e paginação
func (h *PermissionHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PermissionHandler.ListPermissions")
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
	
	req := &application.ListPermissionsRequest{
		TenantID:   tenantID,
		Page:       page,
		PageSize:   pageSize,
		Code:       query.Get("code"),
		Module:     query.Get("module"),
		Resource:   query.Get("resource"),
		Action:     query.Get("action"),
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
	
	// Chamar o serviço para listar permissões
	resp, err := h.permissionService.List(ctx, req)
	if err != nil {
		log.Error().Err(err).Interface("request", req).Msg("Erro ao listar permissões")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusOK, resp)
}

// UpdatePermission atualiza uma permissão existente
func (h *PermissionHandler) UpdatePermission(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PermissionHandler.UpdatePermission")
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
	permissionIDStr := vars["id"]
	
	permissionID, err := uuid.Parse(permissionIDStr)
	if err != nil {
		log.Error().Err(err).Str("permission_id", permissionIDStr).Msg("ID de permissão inválido")
		middleware.RespondWithError(w, http.StatusBadRequest, "ID de permissão inválido")
		return
	}
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("user_id", userID.String()),
		attribute.String("permission_id", permissionID.String()),
	)
	
	var req application.UpdatePermissionRequest
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
	req.ID = permissionID
	
	// Chamar o serviço para atualizar a permissão
	resp, err := h.permissionService.Update(ctx, &req)
	if err != nil {
		log.Error().Err(err).Interface("request", req).Msg("Erro ao atualizar permissão")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusOK, resp)
}

// DeletePermission exclui uma permissão
func (h *PermissionHandler) DeletePermission(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PermissionHandler.DeletePermission")
	defer span.End()
	
	tenantID, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter tenant ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "Tenant ID não encontrado")
		return
	}
	
	// Extrair ID da URL
	vars := mux.Vars(r)
	permissionIDStr := vars["id"]
	
	permissionID, err := uuid.Parse(permissionIDStr)
	if err != nil {
		log.Error().Err(err).Str("permission_id", permissionIDStr).Msg("ID de permissão inválido")
		middleware.RespondWithError(w, http.StatusBadRequest, "ID de permissão inválido")
		return
	}
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("permission_id", permissionID.String()),
	)
	
	// Extrair parâmetros de consulta
	query := r.URL.Query()
	force := query.Get("force") == "true"
	
	// Chamar o serviço para excluir a permissão
	err = h.permissionService.Delete(ctx, tenantID, permissionID, force)
	if err != nil {
		log.Error().Err(err).Str("permission_id", permissionID.String()).Bool("force", force).Msg("Erro ao excluir permissão")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusNoContent, nil)
}

// GetPermissionsByModule recupera permissões por módulo
func (h *PermissionHandler) GetPermissionsByModule(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PermissionHandler.GetPermissionsByModule")
	defer span.End()
	
	tenantID, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter tenant ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "Tenant ID não encontrado")
		return
	}
	
	// Extrair módulo da URL
	vars := mux.Vars(r)
	module := vars["module"]
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("module", module),
	)
	
	// Chamar o serviço para recuperar permissões por módulo
	resp, err := h.permissionService.GetByModule(ctx, tenantID, module)
	if err != nil {
		log.Error().Err(err).Str("module", module).Msg("Erro ao recuperar permissões por módulo")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusOK, resp)
}

// GetPermissionsByResource recupera permissões por recurso
func (h *PermissionHandler) GetPermissionsByResource(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PermissionHandler.GetPermissionsByResource")
	defer span.End()
	
	tenantID, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter tenant ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "Tenant ID não encontrado")
		return
	}
	
	// Extrair recurso da URL
	vars := mux.Vars(r)
	resource := vars["resource"]
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("resource", resource),
	)
	
	// Chamar o serviço para recuperar permissões por recurso
	resp, err := h.permissionService.GetByResource(ctx, tenantID, resource)
	if err != nil {
		log.Error().Err(err).Str("resource", resource).Msg("Erro ao recuperar permissões por recurso")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusOK, resp)
}

// GetPermissionsByAction recupera permissões por ação
func (h *PermissionHandler) GetPermissionsByAction(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PermissionHandler.GetPermissionsByAction")
	defer span.End()
	
	tenantID, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter tenant ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "Tenant ID não encontrado")
		return
	}
	
	// Extrair ação da URL
	vars := mux.Vars(r)
	action := vars["action"]
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("action", action),
	)
	
	// Chamar o serviço para recuperar permissões por ação
	resp, err := h.permissionService.GetByAction(ctx, tenantID, action)
	if err != nil {
		log.Error().Err(err).Str("action", action).Msg("Erro ao recuperar permissões por ação")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusOK, resp)
}

// GetPermissionRoles recupera as funções associadas a uma permissão
func (h *PermissionHandler) GetPermissionRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PermissionHandler.GetPermissionRoles")
	defer span.End()
	
	tenantID, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao obter tenant ID do contexto")
		middleware.RespondWithError(w, http.StatusUnauthorized, "Tenant ID não encontrado")
		return
	}
	
	// Extrair ID da URL
	vars := mux.Vars(r)
	permissionIDStr := vars["id"]
	
	permissionID, err := uuid.Parse(permissionIDStr)
	if err != nil {
		log.Error().Err(err).Str("permission_id", permissionIDStr).Msg("ID de permissão inválido")
		middleware.RespondWithError(w, http.StatusBadRequest, "ID de permissão inválido")
		return
	}
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("permission_id", permissionID.String()),
	)
	
	// Chamar o serviço para recuperar as funções associadas à permissão
	resp, err := h.permissionService.GetRolesWithPermission(ctx, tenantID, permissionID)
	if err != nil {
		log.Error().Err(err).Str("permission_id", permissionID.String()).Msg("Erro ao recuperar funções com permissão")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusOK, resp)
}

// CheckUserPermission verifica se o usuário atual tem uma permissão específica
func (h *PermissionHandler) CheckUserPermission(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PermissionHandler.CheckUserPermission")
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
	
	// Extrair código de permissão da URL
	vars := mux.Vars(r)
	permissionCode := vars["code"]
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("user_id", userID.String()),
		attribute.String("permission_code", permissionCode),
	)
	
	// Chamar o serviço para verificar se o usuário tem a permissão
	hasPermission, err := h.permissionService.UserHasPermission(ctx, tenantID, userID, permissionCode)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Str("permission_code", permissionCode).
			Msg("Erro ao verificar permissão do usuário")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	// Construir a resposta
	resp := map[string]interface{}{
		"has_permission": hasPermission,
		"permission_code": permissionCode,
		"user_id": userID.String(),
	}
	
	middleware.RespondWithJSON(w, http.StatusOK, resp)
}

// SyncSystemPermissions sincroniza as permissões do sistema
func (h *PermissionHandler) SyncSystemPermissions(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PermissionHandler.SyncSystemPermissions")
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
	
	var req application.SyncSystemPermissionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if err.Error() != "EOF" { // Ignorar erro EOF (corpo vazio)
			log.Error().Err(err).Msg("Erro ao decodificar corpo da requisição")
			middleware.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido")
			return
		}
		req = application.SyncSystemPermissionsRequest{}
	}
	
	// Definir o tenant ID da requisição com o valor do contexto
	req.TenantID = tenantID
	
	// Chamar o serviço para sincronizar as permissões do sistema
	resp, err := h.permissionService.SyncSystemPermissions(ctx, &req)
	if err != nil {
		log.Error().Err(err).Interface("request", req).Msg("Erro ao sincronizar permissões do sistema")
		middleware.RespondWithServiceError(w, err)
		return
	}
	
	middleware.RespondWithJSON(w, http.StatusOK, resp)
}