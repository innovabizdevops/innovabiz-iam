package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"innovabiz/iam/identity-service/internal/application"
	"innovabiz/iam/identity-service/internal/domain/model"
	"innovabiz/iam/identity-service/internal/interface/api/dto"
	"innovabiz/iam/identity-service/internal/interface/api/middleware"
)

var tracer = otel.Tracer("innovabiz.iam.interface.api.handlers.role")

// RoleHandler é responsável por gerenciar requisições HTTP relacionadas a funções (roles)
type RoleHandler struct {
	roleService application.RoleService
}

// NewRoleHandler cria uma nova instância de RoleHandler
func NewRoleHandler(roleService application.RoleService) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

// RegisterRoutes registra as rotas do handler no router
func (h *RoleHandler) RegisterRoutes(router *mux.Router) {
	// Rotas para gerenciamento básico de funções
	router.HandleFunc("/api/v1/tenants/{tenant_id}/roles", h.CreateRole).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/tenants/{tenant_id}/roles", h.ListRoles).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/tenants/{tenant_id}/roles/{role_id}", h.GetRole).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/tenants/{tenant_id}/roles/code/{code}", h.GetRoleByCode).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/tenants/{tenant_id}/roles/{role_id}", h.UpdateRole).Methods(http.MethodPut, http.MethodPatch)
	router.HandleFunc("/api/v1/tenants/{tenant_id}/roles/{role_id}", h.DeleteRole).Methods(http.MethodDelete)
	router.HandleFunc("/api/v1/tenants/{tenant_id}/roles/{role_id}/clone", h.CloneRole).Methods(http.MethodPost)

	// Rotas para gerenciamento de permissões
	router.HandleFunc("/api/v1/tenants/{tenant_id}/roles/{role_id}/permissions", h.GetRolePermissions).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/tenants/{tenant_id}/roles/{role_id}/permissions/{permission_id}", h.AssignPermission).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/tenants/{tenant_id}/roles/{role_id}/permissions/{permission_id}", h.RevokePermission).Methods(http.MethodDelete)

	// Rotas para gerenciamento de hierarquia de funções
	router.HandleFunc("/api/v1/tenants/{tenant_id}/roles/{role_id}/children", h.GetChildRoles).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/tenants/{tenant_id}/roles/{role_id}/parents", h.GetParentRoles).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/tenants/{tenant_id}/roles/{role_id}/ancestors", h.GetAncestorRoles).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/tenants/{tenant_id}/roles/{role_id}/descendants", h.GetDescendantRoles).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/tenants/{tenant_id}/roles/{parent_id}/children/{child_id}", h.AssignChildRole).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/tenants/{tenant_id}/roles/{parent_id}/children/{child_id}", h.RemoveChildRole).Methods(http.MethodDelete)

	// Rotas para gerenciamento de usuários de funções
	router.HandleFunc("/api/v1/tenants/{tenant_id}/roles/{role_id}/users", h.GetRoleUsers).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/tenants/{tenant_id}/roles/{role_id}/users/{user_id}", h.AssignUserToRole).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/tenants/{tenant_id}/roles/{role_id}/users/{user_id}", h.RevokeUserFromRole).Methods(http.MethodDelete)
	router.HandleFunc("/api/v1/tenants/{tenant_id}/users/{user_id}/roles", h.GetUserRoles).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/tenants/{tenant_id}/users/{user_id}/roles/active", h.GetUserActiveRoles).Methods(http.MethodGet)

	// Rotas administrativas para funções do sistema
	router.HandleFunc("/api/v1/tenants/{tenant_id}/system/roles", h.SyncSystemRoles).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/tenants/{tenant_id}/system/roles", h.GetSystemRoles).Methods(http.MethodGet)
}

// CreateRole cria uma nova função
func (h *RoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.CreateRole")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}

	// Obter usuário atual do contexto
	currentUserID := middleware.GetUserIDFromContext(ctx)
	if currentUserID == uuid.Nil {
		respondWithError(w, http.StatusUnauthorized, "Usuário não autenticado")
		return
	}

	// Decodificar requisição
	var req dto.CreateRoleRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido: "+err.Error())
		return
	}

	// Validar campos obrigatórios
	if req.Code == "" {
		respondWithError(w, http.StatusBadRequest, "Código da função é obrigatório")
		return
	}

	if req.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Nome da função é obrigatório")
		return
	}

	if req.Type == "" {
		respondWithError(w, http.StatusBadRequest, "Tipo da função é obrigatório")
		return
	}

	// Mapear para DTO de serviço
	serviceReq := application.CreateRoleRequest{
		TenantID:              tenantID,
		Code:                  req.Code,
		Name:                  req.Name,
		Description:           req.Description,
		Type:                  req.Type,
		Metadata:              req.Metadata,
		CreatedBy:             currentUserID,
		IsSystem:              req.IsSystem,
		SyncSystemPermissions: req.SyncSystemPermissions,
		PermissionCodes:       req.PermissionCodes,
	}

	// Chamar serviço
	role, err := h.roleService.CreateRole(ctx, serviceReq)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Mapear resposta
	response := dto.RoleResponse{
		ID:          role.ID().String(),
		TenantID:    role.TenantID().String(),
		Code:        role.Code(),
		Name:        role.Name(),
		Description: role.Description(),
		Type:        role.Type(),
		IsActive:    role.IsActive(),
		IsSystem:    role.IsSystem(),
		Metadata:    role.Metadata(),
		CreatedAt:   role.CreatedAt(),
		CreatedBy:   role.CreatedBy().String(),
		UpdatedAt:   role.UpdatedAt(),
		UpdatedBy:   role.UpdatedBy().String(),
	}

	// Responder
	respondWithJSON(w, http.StatusCreated, response)
}

// GetRole recupera uma função pelo ID
func (h *RoleHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.GetRole")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	roleIDStr := vars["role_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID da função inválido")
		return
	}

	// Chamar serviço
	role, err := h.roleService.GetRole(ctx, tenantID, roleID)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Mapear resposta
	response := dto.RoleResponse{
		ID:          role.ID().String(),
		TenantID:    role.TenantID().String(),
		Code:        role.Code(),
		Name:        role.Name(),
		Description: role.Description(),
		Type:        role.Type(),
		IsActive:    role.IsActive(),
		IsSystem:    role.IsSystem(),
		Metadata:    role.Metadata(),
		CreatedAt:   role.CreatedAt(),
		CreatedBy:   role.CreatedBy().String(),
		UpdatedAt:   role.UpdatedAt(),
		UpdatedBy:   role.UpdatedBy().String(),
	}

	// Responder
	respondWithJSON(w, http.StatusOK, response)
}

// GetRoleByCode recupera uma função pelo código
func (h *RoleHandler) GetRoleByCode(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.GetRoleByCode")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	code := vars["code"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}

	// Chamar serviço
	role, err := h.roleService.GetRoleByCode(ctx, tenantID, code)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Mapear resposta
	response := dto.RoleResponse{
		ID:          role.ID().String(),
		TenantID:    role.TenantID().String(),
		Code:        role.Code(),
		Name:        role.Name(),
		Description: role.Description(),
		Type:        role.Type(),
		IsActive:    role.IsActive(),
		IsSystem:    role.IsSystem(),
		Metadata:    role.Metadata(),
		CreatedAt:   role.CreatedAt(),
		CreatedBy:   role.CreatedBy().String(),
		UpdatedAt:   role.UpdatedAt(),
		UpdatedBy:   role.UpdatedBy().String(),
	}

	// Responder
	respondWithJSON(w, http.StatusOK, response)
}// ListRoles lista funções com filtros e paginação
func (h *RoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.ListRoles")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}

	// Extrair parâmetros de consulta para filtros
	query := r.URL.Query()
	
	// Construir filtro
	filter := application.RoleFilter{
		NameOrCodeContains: query.Get("q"),
	}

	// Processar tipos de funções (podem ser múltiplos)
	if typeValues, ok := query["type"]; ok {
		filter.Types = typeValues
	}

	// Processar filtro de ativação
	if isActiveStr := query.Get("is_active"); isActiveStr != "" {
		isActive, err := strconv.ParseBool(isActiveStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Valor inválido para is_active")
			return
		}
		filter.IsActive = &isActive
	}

	// Processar filtro de função de sistema
	if isSystemStr := query.Get("is_system"); isSystemStr != "" {
		isSystem, err := strconv.ParseBool(isSystemStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Valor inválido para is_system")
			return
		}
		filter.IsSystem = &isSystem
	}

	// Extrair parâmetros de paginação
	page := 1
	pageSize := 20 // valor padrão
	
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	pagination := application.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	// Chamar serviço
	roles, total, err := h.roleService.ListRoles(ctx, tenantID, filter, pagination)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Mapear resposta
	items := make([]dto.RoleResponse, 0, len(roles))
	for _, role := range roles {
		items = append(items, dto.RoleResponse{
			ID:          role.ID().String(),
			TenantID:    role.TenantID().String(),
			Code:        role.Code(),
			Name:        role.Name(),
			Description: role.Description(),
			Type:        role.Type(),
			IsActive:    role.IsActive(),
			IsSystem:    role.IsSystem(),
			Metadata:    role.Metadata(),
			CreatedAt:   role.CreatedAt(),
			CreatedBy:   role.CreatedBy().String(),
			UpdatedAt:   role.UpdatedAt(),
			UpdatedBy:   role.UpdatedBy().String(),
		})
	}

	// Calcular total de páginas
	totalPages := total / int64(pageSize)
	if total%int64(pageSize) > 0 {
		totalPages++
	}

	response := dto.RoleListResponse{
		Items:      items,
		TotalItems: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(totalPages),
	}

	// Responder
	respondWithJSON(w, http.StatusOK, response)
}

// UpdateRole atualiza uma função existente
func (h *RoleHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.UpdateRole")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	roleIDStr := vars["role_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID da função inválido")
		return
	}

	// Obter usuário atual do contexto
	currentUserID := middleware.GetUserIDFromContext(ctx)
	if currentUserID == uuid.Nil {
		respondWithError(w, http.StatusUnauthorized, "Usuário não autenticado")
		return
	}

	// Decodificar requisição
	var req dto.UpdateRoleRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido: "+err.Error())
		return
	}

	// Mapear para DTO de serviço
	serviceReq := application.UpdateRoleRequest{
		ID:              roleID,
		TenantID:        tenantID,
		Code:            req.Code,
		Name:            req.Name,
		Description:     req.Description,
		Type:            req.Type,
		IsActive:        req.IsActive,
		IsSystem:        req.IsSystem,
		Metadata:        req.Metadata,
		SyncPermissions: req.SyncPermissions,
		PermissionCodes: req.PermissionCodes,
		UpdatedBy:       currentUserID,
	}

	// Chamar serviço
	role, err := h.roleService.UpdateRole(ctx, serviceReq)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Mapear resposta
	response := dto.RoleResponse{
		ID:          role.ID().String(),
		TenantID:    role.TenantID().String(),
		Code:        role.Code(),
		Name:        role.Name(),
		Description: role.Description(),
		Type:        role.Type(),
		IsActive:    role.IsActive(),
		IsSystem:    role.IsSystem(),
		Metadata:    role.Metadata(),
		CreatedAt:   role.CreatedAt(),
		CreatedBy:   role.CreatedBy().String(),
		UpdatedAt:   role.UpdatedAt(),
		UpdatedBy:   role.UpdatedBy().String(),
	}

	// Responder
	respondWithJSON(w, http.StatusOK, response)
}

// DeleteRole exclui uma função
func (h *RoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.DeleteRole")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	roleIDStr := vars["role_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID da função inválido")
		return
	}

	// Obter usuário atual do contexto
	currentUserID := middleware.GetUserIDFromContext(ctx)
	if currentUserID == uuid.Nil {
		respondWithError(w, http.StatusUnauthorized, "Usuário não autenticado")
		return
	}

	// Verificar parâmetros de consulta para opções de exclusão
	query := r.URL.Query()
	
	// Verificar se é exclusão permanente
	hardDelete := false
	if hardDeleteStr := query.Get("hard_delete"); hardDeleteStr != "" {
		var err error
		hardDelete, err = strconv.ParseBool(hardDeleteStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Valor inválido para hard_delete")
			return
		}
	}
	
	// Verificar se é exclusão forçada
	force := false
	if forceStr := query.Get("force"); forceStr != "" {
		var err error
		force, err = strconv.ParseBool(forceStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Valor inválido para force")
			return
		}
	}

	// Construir requisição
	serviceReq := application.DeleteRoleRequest{
		ID:         roleID,
		TenantID:   tenantID,
		DeletedBy:  currentUserID,
		HardDelete: hardDelete,
		Force:      force,
	}

	// Chamar serviço
	err = h.roleService.DeleteRole(ctx, serviceReq)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Responder
	w.WriteHeader(http.StatusNoContent)
}

// CloneRole clona uma função existente
func (h *RoleHandler) CloneRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.CloneRole")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	roleIDStr := vars["role_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID da função inválido")
		return
	}

	// Obter usuário atual do contexto
	currentUserID := middleware.GetUserIDFromContext(ctx)
	if currentUserID == uuid.Nil {
		respondWithError(w, http.StatusUnauthorized, "Usuário não autenticado")
		return
	}

	// Decodificar requisição
	var req dto.CloneRoleRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido: "+err.Error())
		return
	}

	// Mapear para DTO de serviço
	serviceReq := application.CloneRoleRequest{
		TenantID:       tenantID,
		SourceRoleID:   roleID,
		TargetCode:     req.TargetCode,
		TargetName:     req.TargetName,
		CopyPermissions: req.CopyPermissions,
		CopyHierarchy:  req.CopyHierarchy,
		CreatedBy:      currentUserID,
	}

	// Chamar serviço
	clonedRole, err := h.roleService.CloneRole(ctx, serviceReq)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Mapear resposta
	response := dto.RoleResponse{
		ID:          clonedRole.ID().String(),
		TenantID:    clonedRole.TenantID().String(),
		Code:        clonedRole.Code(),
		Name:        clonedRole.Name(),
		Description: clonedRole.Description(),
		Type:        clonedRole.Type(),
		IsActive:    clonedRole.IsActive(),
		IsSystem:    clonedRole.IsSystem(),
		Metadata:    clonedRole.Metadata(),
		CreatedAt:   clonedRole.CreatedAt(),
		CreatedBy:   clonedRole.CreatedBy().String(),
		UpdatedAt:   clonedRole.UpdatedAt(),
		UpdatedBy:   clonedRole.UpdatedBy().String(),
	}

	// Responder
	respondWithJSON(w, http.StatusCreated, response)
}// GetRolePermissions lista as permissões atribuídas a uma função
func (h *RoleHandler) GetRolePermissions(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.GetRolePermissions")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	roleIDStr := vars["role_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID da função inválido")
		return
	}

	// Extrair parâmetros de paginação
	query := r.URL.Query()
	page := 1
	pageSize := 20 // valor padrão
	
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	pagination := application.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	// Chamar serviço
	permissions, total, err := h.roleService.GetRolePermissions(ctx, tenantID, roleID, pagination)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Mapear resposta
	items := make([]dto.PermissionResponse, 0, len(permissions))
	for _, perm := range permissions {
		items = append(items, dto.PermissionResponse{
			ID:           perm.ID().String(),
			TenantID:     perm.TenantID().String(),
			Code:         perm.Code(),
			Name:         perm.Name(),
			Description:  perm.Description(),
			ResourceType: perm.ResourceType(),
			Action:       perm.Action(),
			Effect:       perm.Effect(),
			Conditions:   perm.Conditions(),
			Metadata:     perm.Metadata(),
			IsActive:     perm.IsActive(),
			IsSystem:     perm.IsSystem(),
			CreatedAt:    perm.CreatedAt(),
			CreatedBy:    perm.CreatedBy().String(),
		})
	}

	// Calcular total de páginas
	totalPages := total / int64(pageSize)
	if total%int64(pageSize) > 0 {
		totalPages++
	}

	response := struct {
		Items      []dto.PermissionResponse `json:"items"`
		TotalItems int64                    `json:"total_items"`
		Page       int                      `json:"page"`
		PageSize   int                      `json:"page_size"`
		TotalPages int                      `json:"total_pages"`
	}{
		Items:      items,
		TotalItems: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(totalPages),
	}

	// Responder
	respondWithJSON(w, http.StatusOK, response)
}

// AssignPermission atribui uma permissão a uma função
func (h *RoleHandler) AssignPermission(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.AssignPermission")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	roleIDStr := vars["role_id"]
	permissionIDStr := vars["permission_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID da função inválido")
		return
	}
	
	permissionID, err := uuid.Parse(permissionIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID da permissão inválido")
		return
	}

	// Obter usuário atual do contexto
	currentUserID := middleware.GetUserIDFromContext(ctx)
	if currentUserID == uuid.Nil {
		respondWithError(w, http.StatusUnauthorized, "Usuário não autenticado")
		return
	}

	// Chamar serviço
	err = h.roleService.AssignPermission(ctx, tenantID, roleID, permissionID, currentUserID)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Responder
	w.WriteHeader(http.StatusNoContent)
}

// RevokePermission remove uma permissão de uma função
func (h *RoleHandler) RevokePermission(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.RevokePermission")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	roleIDStr := vars["role_id"]
	permissionIDStr := vars["permission_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID da função inválido")
		return
	}
	
	permissionID, err := uuid.Parse(permissionIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID da permissão inválido")
		return
	}

	// Obter usuário atual do contexto
	currentUserID := middleware.GetUserIDFromContext(ctx)
	if currentUserID == uuid.Nil {
		respondWithError(w, http.StatusUnauthorized, "Usuário não autenticado")
		return
	}

	// Chamar serviço
	err = h.roleService.RevokePermission(ctx, tenantID, roleID, permissionID, currentUserID)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Responder
	w.WriteHeader(http.StatusNoContent)
}

// GetChildRoles obtém as funções filhas de uma função
func (h *RoleHandler) GetChildRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.GetChildRoles")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	roleIDStr := vars["role_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID da função inválido")
		return
	}

	// Extrair parâmetros de paginação
	query := r.URL.Query()
	page := 1
	pageSize := 20 // valor padrão
	
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	pagination := application.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	// Chamar serviço
	childRoles, total, err := h.roleService.GetChildRoles(ctx, tenantID, roleID, pagination)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Mapear resposta
	items := make([]dto.RoleResponse, 0, len(childRoles))
	for _, role := range childRoles {
		items = append(items, dto.RoleResponse{
			ID:          role.ID().String(),
			TenantID:    role.TenantID().String(),
			Code:        role.Code(),
			Name:        role.Name(),
			Description: role.Description(),
			Type:        role.Type(),
			IsActive:    role.IsActive(),
			IsSystem:    role.IsSystem(),
			Metadata:    role.Metadata(),
			CreatedAt:   role.CreatedAt(),
			CreatedBy:   role.CreatedBy().String(),
			UpdatedAt:   role.UpdatedAt(),
			UpdatedBy:   role.UpdatedBy().String(),
		})
	}

	// Calcular total de páginas
	totalPages := total / int64(pageSize)
	if total%int64(pageSize) > 0 {
		totalPages++
	}

	response := dto.RoleListResponse{
		Items:      items,
		TotalItems: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(totalPages),
	}

	// Responder
	respondWithJSON(w, http.StatusOK, response)
}

// GetParentRoles obtém as funções pai de uma função
func (h *RoleHandler) GetParentRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.GetParentRoles")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	roleIDStr := vars["role_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID da função inválido")
		return
	}

	// Extrair parâmetros de paginação
	query := r.URL.Query()
	page := 1
	pageSize := 20 // valor padrão
	
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	pagination := application.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	// Chamar serviço
	parentRoles, total, err := h.roleService.GetParentRoles(ctx, tenantID, roleID, pagination)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Mapear resposta
	items := make([]dto.RoleResponse, 0, len(parentRoles))
	for _, role := range parentRoles {
		items = append(items, dto.RoleResponse{
			ID:          role.ID().String(),
			TenantID:    role.TenantID().String(),
			Code:        role.Code(),
			Name:        role.Name(),
			Description: role.Description(),
			Type:        role.Type(),
			IsActive:    role.IsActive(),
			IsSystem:    role.IsSystem(),
			Metadata:    role.Metadata(),
			CreatedAt:   role.CreatedAt(),
			CreatedBy:   role.CreatedBy().String(),
			UpdatedAt:   role.UpdatedAt(),
			UpdatedBy:   role.UpdatedBy().String(),
		})
	}

	// Calcular total de páginas
	totalPages := total / int64(pageSize)
	if total%int64(pageSize) > 0 {
		totalPages++
	}

	response := dto.RoleListResponse{
		Items:      items,
		TotalItems: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(totalPages),
	}

	// Responder
	respondWithJSON(w, http.StatusOK, response)
}// GetAncestorRoles obtém todas as funções ancestrais de uma função
func (h *RoleHandler) GetAncestorRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.GetAncestorRoles")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	roleIDStr := vars["role_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID da função inválido")
		return
	}

	// Chamar serviço
	ancestorRoles, err := h.roleService.GetAncestorRoles(ctx, tenantID, roleID)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Mapear resposta
	items := make([]dto.RoleResponse, 0, len(ancestorRoles))
	for _, role := range ancestorRoles {
		items = append(items, dto.RoleResponse{
			ID:          role.ID().String(),
			TenantID:    role.TenantID().String(),
			Code:        role.Code(),
			Name:        role.Name(),
			Description: role.Description(),
			Type:        role.Type(),
			IsActive:    role.IsActive(),
			IsSystem:    role.IsSystem(),
			Metadata:    role.Metadata(),
			CreatedAt:   role.CreatedAt(),
			CreatedBy:   role.CreatedBy().String(),
			UpdatedAt:   role.UpdatedAt(),
			UpdatedBy:   role.UpdatedBy().String(),
		})
	}

	// Responder
	respondWithJSON(w, http.StatusOK, struct {
		Items []dto.RoleResponse `json:"items"`
		Total int                `json:"total"`
	}{
		Items: items,
		Total: len(items),
	})
}

// GetDescendantRoles obtém todas as funções descendentes de uma função
func (h *RoleHandler) GetDescendantRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.GetDescendantRoles")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	roleIDStr := vars["role_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID da função inválido")
		return
	}

	// Chamar serviço
	descendantRoles, err := h.roleService.GetDescendantRoles(ctx, tenantID, roleID)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Mapear resposta
	items := make([]dto.RoleResponse, 0, len(descendantRoles))
	for _, role := range descendantRoles {
		items = append(items, dto.RoleResponse{
			ID:          role.ID().String(),
			TenantID:    role.TenantID().String(),
			Code:        role.Code(),
			Name:        role.Name(),
			Description: role.Description(),
			Type:        role.Type(),
			IsActive:    role.IsActive(),
			IsSystem:    role.IsSystem(),
			Metadata:    role.Metadata(),
			CreatedAt:   role.CreatedAt(),
			CreatedBy:   role.CreatedBy().String(),
			UpdatedAt:   role.UpdatedAt(),
			UpdatedBy:   role.UpdatedBy().String(),
		})
	}

	// Responder
	respondWithJSON(w, http.StatusOK, struct {
		Items []dto.RoleResponse `json:"items"`
		Total int                `json:"total"`
	}{
		Items: items,
		Total: len(items),
	})
}

// AssignChildRole atribui uma função filha a uma função pai
func (h *RoleHandler) AssignChildRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.AssignChildRole")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	parentIDStr := vars["parent_id"]
	childIDStr := vars["child_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}
	
	parentID, err := uuid.Parse(parentIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID da função pai inválido")
		return
	}
	
	childID, err := uuid.Parse(childIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID da função filha inválido")
		return
	}

	// Obter usuário atual do contexto
	currentUserID := middleware.GetUserIDFromContext(ctx)
	if currentUserID == uuid.Nil {
		respondWithError(w, http.StatusUnauthorized, "Usuário não autenticado")
		return
	}

	// Chamar serviço
	err = h.roleService.AssignChildRole(ctx, tenantID, parentID, childID, currentUserID)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Responder
	w.WriteHeader(http.StatusNoContent)
}

// RemoveChildRole remove uma função filha de uma função pai
func (h *RoleHandler) RemoveChildRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.RemoveChildRole")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	parentIDStr := vars["parent_id"]
	childIDStr := vars["child_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}
	
	parentID, err := uuid.Parse(parentIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID da função pai inválido")
		return
	}
	
	childID, err := uuid.Parse(childIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID da função filha inválido")
		return
	}

	// Obter usuário atual do contexto
	currentUserID := middleware.GetUserIDFromContext(ctx)
	if currentUserID == uuid.Nil {
		respondWithError(w, http.StatusUnauthorized, "Usuário não autenticado")
		return
	}

	// Chamar serviço
	err = h.roleService.RemoveChildRole(ctx, tenantID, parentID, childID, currentUserID)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Responder
	w.WriteHeader(http.StatusNoContent)
}

// AssignUserToRole atribui um usuário a uma função
func (h *RoleHandler) AssignUserToRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.AssignUserToRole")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	roleIDStr := vars["role_id"]
	userIDStr := vars["user_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID da função inválido")
		return
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do usuário inválido")
		return
	}

	// Obter usuário atual do contexto
	currentUserID := middleware.GetUserIDFromContext(ctx)
	if currentUserID == uuid.Nil {
		respondWithError(w, http.StatusUnauthorized, "Usuário não autenticado")
		return
	}

	// Decodificar requisição
	var req dto.AssignUserToRoleRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido: "+err.Error())
		return
	}

	// Verificar e ajustar datas
	activatesAt := time.Now()
	if !req.ActivatesAt.IsZero() {
		activatesAt = req.ActivatesAt
	}

	var expiresAt *time.Time
	if !req.ExpiresAt.IsZero() {
		expiresAt = &req.ExpiresAt
	}

	// Chamar serviço
	err = h.roleService.AssignUserToRole(ctx, tenantID, roleID, userID, activatesAt, expiresAt, currentUserID)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Responder
	w.WriteHeader(http.StatusNoContent)
}

// RevokeUserFromRole remove um usuário de uma função
func (h *RoleHandler) RevokeUserFromRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.RevokeUserFromRole")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	roleIDStr := vars["role_id"]
	userIDStr := vars["user_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID da função inválido")
		return
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do usuário inválido")
		return
	}

	// Obter usuário atual do contexto
	currentUserID := middleware.GetUserIDFromContext(ctx)
	if currentUserID == uuid.Nil {
		respondWithError(w, http.StatusUnauthorized, "Usuário não autenticado")
		return
	}

	// Chamar serviço
	err = h.roleService.RevokeUserFromRole(ctx, tenantID, roleID, userID, currentUserID)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Responder
	w.WriteHeader(http.StatusNoContent)
}// GetRoleUsers obtém os usuários atribuídos a uma função
func (h *RoleHandler) GetRoleUsers(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.GetRoleUsers")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	roleIDStr := vars["role_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}
	
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID da função inválido")
		return
	}

	// Extrair parâmetros de paginação
	query := r.URL.Query()
	page := 1
	pageSize := 20 // valor padrão
	
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	pagination := application.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	// Verificar apenas ativos
	includeOnlyActive := false
	if activeStr := query.Get("active_only"); activeStr != "" {
		var err error
		includeOnlyActive, err = strconv.ParseBool(activeStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Valor inválido para active_only")
			return
		}
	}

	// Chamar serviço
	userRoles, total, err := h.roleService.GetRoleUsers(ctx, tenantID, roleID, includeOnlyActive, pagination)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Mapear resposta
	items := make([]dto.UserRoleDetailResponse, 0, len(userRoles))
	for _, ur := range userRoles {
		expires := time.Time{}
		if ur.ExpiresAt != nil {
			expires = *ur.ExpiresAt
		}

		items = append(items, dto.UserRoleDetailResponse{
			UserID:      ur.UserID.String(),
			ActivatesAt: ur.ActivatesAt,
			ExpiresAt:   expires,
			AssignedAt:  ur.AssignedAt,
			AssignedBy:  ur.AssignedBy.String(),
		})
	}

	// Calcular total de páginas
	totalPages := total / int64(pageSize)
	if total%int64(pageSize) > 0 {
		totalPages++
	}

	response := dto.UserRoleListResponse{
		Items:      items,
		TotalItems: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(totalPages),
	}

	// Responder
	respondWithJSON(w, http.StatusOK, response)
}

// GetUserRoles obtém todas as funções de um usuário
func (h *RoleHandler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.GetUserRoles")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	userIDStr := vars["user_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do usuário inválido")
		return
	}

	// Extrair parâmetros de paginação
	query := r.URL.Query()
	page := 1
	pageSize := 20 // valor padrão
	
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	pagination := application.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	// Chamar serviço
	roles, total, err := h.roleService.GetUserRoles(ctx, tenantID, userID, pagination)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Mapear resposta
	items := make([]dto.UserRoleResponse, 0, len(roles))
	for _, ur := range roles {
		expires := time.Time{}
		if ur.ExpiresAt != nil {
			expires = *ur.ExpiresAt
		}

		items = append(items, dto.UserRoleResponse{
			RoleID:      ur.Role.ID().String(),
			RoleCode:    ur.Role.Code(),
			RoleName:    ur.Role.Name(),
			Type:        ur.Role.Type(),
			IsActive:    ur.Role.IsActive(),
			ActivatesAt: ur.ActivatesAt,
			ExpiresAt:   expires,
			AssignedAt:  ur.AssignedAt,
			AssignedBy:  ur.AssignedBy.String(),
		})
	}

	// Calcular total de páginas
	totalPages := total / int64(pageSize)
	if total%int64(pageSize) > 0 {
		totalPages++
	}

	response := struct {
		Items      []dto.UserRoleResponse `json:"items"`
		TotalItems int64                  `json:"total_items"`
		Page       int                    `json:"page"`
		PageSize   int                    `json:"page_size"`
		TotalPages int                    `json:"total_pages"`
	}{
		Items:      items,
		TotalItems: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(totalPages),
	}

	// Responder
	respondWithJSON(w, http.StatusOK, response)
}

// GetUserActiveRoles obtém as funções ativas de um usuário
func (h *RoleHandler) GetUserActiveRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.GetUserActiveRoles")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	userIDStr := vars["user_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do usuário inválido")
		return
	}

	// Chamar serviço
	roles, err := h.roleService.GetUserActiveRoles(ctx, tenantID, userID)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Mapear resposta
	items := make([]dto.UserRoleResponse, 0, len(roles))
	for _, ur := range roles {
		expires := time.Time{}
		if ur.ExpiresAt != nil {
			expires = *ur.ExpiresAt
		}

		items = append(items, dto.UserRoleResponse{
			RoleID:      ur.Role.ID().String(),
			RoleCode:    ur.Role.Code(),
			RoleName:    ur.Role.Name(),
			Type:        ur.Role.Type(),
			IsActive:    ur.Role.IsActive(),
			ActivatesAt: ur.ActivatesAt,
			ExpiresAt:   expires,
			AssignedAt:  ur.AssignedAt,
			AssignedBy:  ur.AssignedBy.String(),
		})
	}

	// Responder
	respondWithJSON(w, http.StatusOK, struct {
		Items []dto.UserRoleResponse `json:"items"`
		Total int                    `json:"total"`
	}{
		Items: items,
		Total: len(items),
	})
}

// SyncSystemRoles sincroniza as funções do sistema
func (h *RoleHandler) SyncSystemRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.SyncSystemRoles")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}

	// Obter usuário atual do contexto
	currentUserID := middleware.GetUserIDFromContext(ctx)
	if currentUserID == uuid.Nil {
		respondWithError(w, http.StatusUnauthorized, "Usuário não autenticado")
		return
	}

	// Decodificar requisição
	var req dto.SyncSystemRolesRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido: "+err.Error())
		return
	}

	// Mapear para DTO de serviço
	systemRoles := make([]model.SystemRoleDefinition, 0, len(req.SystemRoles))
	for _, sr := range req.SystemRoles {
		systemRoles = append(systemRoles, model.SystemRoleDefinition{
			Code:            sr.Code,
			Name:            sr.Name,
			Description:     sr.Description,
			Type:            sr.Type,
			PermissionCodes: sr.PermissionCodes,
			ParentCodes:     sr.ParentCodes,
			Metadata:        sr.Metadata,
		})
	}

	// Chamar serviço
	createdCount, updatedCount, err := h.roleService.SyncSystemRoles(ctx, tenantID, systemRoles, currentUserID)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Responder
	respondWithJSON(w, http.StatusOK, struct {
		CreatedCount int `json:"created_count"`
		UpdatedCount int `json:"updated_count"`
	}{
		CreatedCount: createdCount,
		UpdatedCount: updatedCount,
	})
}

// GetSystemRoles obtém todas as funções do sistema
func (h *RoleHandler) GetSystemRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "RoleHandler.GetSystemRoles")
	defer span.End()
	
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID do tenant inválido")
		return
	}

	// Extrair parâmetros de paginação
	query := r.URL.Query()
	page := 1
	pageSize := 50 // valor padrão para funções do sistema (pode ser maior)
	
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 200 {
			pageSize = ps
		}
	}

	pagination := application.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	// Construir filtro (apenas funções do sistema)
	isSystem := true
	filter := application.RoleFilter{
		IsSystem: &isSystem,
	}

	// Chamar serviço
	systemRoles, total, err := h.roleService.ListRoles(ctx, tenantID, filter, pagination)
	if err != nil {
		handleRoleServiceError(w, err)
		return
	}

	// Mapear resposta
	items := make([]dto.RoleResponse, 0, len(systemRoles))
	for _, role := range systemRoles {
		items = append(items, dto.RoleResponse{
			ID:          role.ID().String(),
			TenantID:    role.TenantID().String(),
			Code:        role.Code(),
			Name:        role.Name(),
			Description: role.Description(),
			Type:        role.Type(),
			IsActive:    role.IsActive(),
			IsSystem:    role.IsSystem(),
			Metadata:    role.Metadata(),
			CreatedAt:   role.CreatedAt(),
			CreatedBy:   role.CreatedBy().String(),
			UpdatedAt:   role.UpdatedAt(),
			UpdatedBy:   role.UpdatedBy().String(),
		})
	}

	// Calcular total de páginas
	totalPages := total / int64(pageSize)
	if total%int64(pageSize) > 0 {
		totalPages++
	}

	response := dto.RoleListResponse{
		Items:      items,
		TotalItems: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(totalPages),
	}

	// Responder
	respondWithJSON(w, http.StatusOK, response)
}