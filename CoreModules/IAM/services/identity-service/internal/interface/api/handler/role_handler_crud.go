package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"innovabiz/iam/identity-service/internal/application"
	"innovabiz/iam/identity-service/internal/domain/model"
)

// RoleRequest representa o modelo de dados para criação/atualização de uma função
type RoleRequest struct {
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type"`
	IsActive    bool                   `json:"isActive"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RoleResponse representa o modelo de dados para retorno de uma função
type RoleResponse struct {
	ID          uuid.UUID              `json:"id"`
	TenantID    uuid.UUID              `json:"tenantId"`
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type"`
	IsSystem    bool                   `json:"isSystem"`
	IsActive    bool                   `json:"isActive"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	CreatedBy   uuid.UUID              `json:"createdBy"`
	UpdatedAt   *time.Time             `json:"updatedAt,omitempty"`
	UpdatedBy   *uuid.UUID             `json:"updatedBy,omitempty"`
	Version     int                    `json:"version"`
}

// toRoleResponse converte um modelo de domínio Role para RoleResponse
func toRoleResponse(role *model.Role) RoleResponse {
	response := RoleResponse{
		ID:          role.ID,
		TenantID:    role.TenantID,
		Code:        role.Code,
		Name:        role.Name,
		Description: role.Description,
		Type:        role.Type,
		IsSystem:    role.IsSystem,
		IsActive:    role.IsActive,
		Metadata:    role.Metadata,
		CreatedAt:   role.CreatedAt,
		CreatedBy:   role.CreatedBy,
		Version:     role.Version,
	}

	if !role.UpdatedAt.IsZero() {
		response.UpdatedAt = &role.UpdatedAt
		response.UpdatedBy = &role.UpdatedBy
	}

	return response
}

// toRoleResponseList converte uma lista de modelos de domínio Role para []RoleResponse
func toRoleResponseList(roles []*model.Role) []RoleResponse {
	responseList := make([]RoleResponse, len(roles))
	for i, role := range roles {
		responseList[i] = toRoleResponse(role)
	}
	return responseList
}

// CreateRole cria uma nova função
func (h *RoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.CreateRole")
	defer span.End()

	tenantID := h.getTenantID(r)
	userID := h.getUserID(r)

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("user.id", userID.String()),
	)

	var roleRequest RoleRequest
	if err := json.NewDecoder(r.Body).Decode(&roleRequest); err != nil {
		span.SetStatus(codes.Error, "Falha ao decodificar requisição")
		span.RecordError(err)
		h.respondWithError(w, http.StatusBadRequest, "invalid_request", "Formato de requisição inválido")
		return
	}

	// Validações básicas
	if roleRequest.Code == "" {
		h.respondWithError(w, http.StatusBadRequest, "invalid_request", "Código da função é obrigatório")
		return
	}
	if roleRequest.Name == "" {
		h.respondWithError(w, http.StatusBadRequest, "invalid_request", "Nome da função é obrigatório")
		return
	}
	if roleRequest.Type == "" {
		h.respondWithError(w, http.StatusBadRequest, "invalid_request", "Tipo da função é obrigatório")
		return
	}

	// Mapear para modelo de domínio
	role := &model.Role{
		TenantID:    tenantID,
		Code:        roleRequest.Code,
		Name:        roleRequest.Name,
		Description: roleRequest.Description,
		Type:        roleRequest.Type,
		IsActive:    roleRequest.IsActive,
		Metadata:    roleRequest.Metadata,
	}

	// Criar função no serviço
	createdRole, err := h.roleService.CreateRole(ctx, tenantID, role, userID)
	if err != nil {
		span.SetStatus(codes.Error, "Falha ao criar função")
		span.RecordError(err)

		// Mapear erros específicos do domínio para códigos HTTP apropriados
		switch err.(type) {
		case *application.ValidationError:
			h.respondWithError(w, http.StatusBadRequest, "validation_error", err.Error())
		case *application.DuplicateResourceError:
			h.respondWithError(w, http.StatusConflict, "conflict", err.Error())
		default:
			h.logger.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("code", roleRequest.Code).
				Msg("Erro ao criar função")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Responder com a função criada
	h.respondWithJSON(w, http.StatusCreated, toRoleResponse(createdRole))
}

// GetRole obtém uma função por ID
func (h *RoleHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.GetRole")
	defer span.End()

	tenantID := h.getTenantID(r)
	vars := mux.Vars(r)
	roleID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_id", "ID da função inválido")
		return
	}

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("role.id", roleID.String()),
	)

	role, err := h.roleService.GetRoleByID(ctx, tenantID, roleID)
	if err != nil {
		span.SetStatus(codes.Error, "Falha ao obter função")
		span.RecordError(err)

		// Mapear erros específicos do domínio para códigos HTTP apropriados
		switch err.(type) {
		case *application.ResourceNotFoundError:
			h.respondWithError(w, http.StatusNotFound, "not_found", err.Error())
		default:
			h.logger.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("role_id", roleID.String()).
				Msg("Erro ao obter função")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Responder com a função
	h.respondWithJSON(w, http.StatusOK, toRoleResponse(role))
}

// ListRoles lista funções com filtros e paginação
func (h *RoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.ListRoles")
	defer span.End()

	tenantID := h.getTenantID(r)
	pagination := h.getPagination(r)

	// Extrair parâmetros de filtro da query string
	query := r.URL.Query()
	filter := application.RoleFilter{
		Code:       query.Get("code"),
		Name:       query.Get("name"),
		Type:       query.Get("type"),
		IsActive:   nil,
		IsSystem:   nil,
	}

	// Converter parâmetros booleanos quando presentes
	if isActiveStr := query.Get("isActive"); isActiveStr != "" {
		isActive := isActiveStr == "true"
		filter.IsActive = &isActive
	}

	if isSystemStr := query.Get("isSystem"); isSystemStr != "" {
		isSystem := isSystemStr == "true"
		filter.IsSystem = &isSystem
	}

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.Int("pagination.page", pagination.Page),
		attribute.Int("pagination.pageSize", pagination.PageSize),
	)

	roles, totalCount, err := h.roleService.ListRoles(ctx, tenantID, filter, pagination)
	if err != nil {
		span.SetStatus(codes.Error, "Falha ao listar funções")
		span.RecordError(err)

		h.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("Erro ao listar funções")
		h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		return
	}

	// Calcular total de páginas
	totalPages := totalCount / int64(pagination.PageSize)
	if totalCount%int64(pagination.PageSize) > 0 {
		totalPages++
	}

	// Responder com a lista de funções e informações de paginação
	h.respondWithJSON(w, http.StatusOK, response{
		Data: toRoleResponseList(roles),
		Pagination: &paginationResponse{
			Page:       pagination.Page,
			PageSize:   pagination.PageSize,
			TotalItems: totalCount,
			TotalPages: int(totalPages),
		},
	})
}

// UpdateRole atualiza uma função existente
func (h *RoleHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.UpdateRole")
	defer span.End()

	tenantID := h.getTenantID(r)
	userID := h.getUserID(r)
	vars := mux.Vars(r)
	roleID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_id", "ID da função inválido")
		return
	}

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("role.id", roleID.String()),
		attribute.String("user.id", userID.String()),
	)

	var roleRequest RoleRequest
	if err := json.NewDecoder(r.Body).Decode(&roleRequest); err != nil {
		span.SetStatus(codes.Error, "Falha ao decodificar requisição")
		span.RecordError(err)
		h.respondWithError(w, http.StatusBadRequest, "invalid_request", "Formato de requisição inválido")
		return
	}

	// Validações básicas
	if roleRequest.Code == "" {
		h.respondWithError(w, http.StatusBadRequest, "invalid_request", "Código da função é obrigatório")
		return
	}
	if roleRequest.Name == "" {
		h.respondWithError(w, http.StatusBadRequest, "invalid_request", "Nome da função é obrigatório")
		return
	}
	if roleRequest.Type == "" {
		h.respondWithError(w, http.StatusBadRequest, "invalid_request", "Tipo da função é obrigatório")
		return
	}

	// Mapear para modelo de domínio
	role := &model.Role{
		ID:          roleID,
		TenantID:    tenantID,
		Code:        roleRequest.Code,
		Name:        roleRequest.Name,
		Description: roleRequest.Description,
		Type:        roleRequest.Type,
		IsActive:    roleRequest.IsActive,
		Metadata:    roleRequest.Metadata,
	}

	// Atualizar função no serviço
	updatedRole, err := h.roleService.UpdateRole(ctx, tenantID, role, userID)
	if err != nil {
		span.SetStatus(codes.Error, "Falha ao atualizar função")
		span.RecordError(err)

		// Mapear erros específicos do domínio para códigos HTTP apropriados
		switch err.(type) {
		case *application.ResourceNotFoundError:
			h.respondWithError(w, http.StatusNotFound, "not_found", err.Error())
		case *application.ValidationError:
			h.respondWithError(w, http.StatusBadRequest, "validation_error", err.Error())
		case *application.DuplicateResourceError:
			h.respondWithError(w, http.StatusConflict, "conflict", err.Error())
		case *application.ConcurrentModificationError:
			h.respondWithError(w, http.StatusPreconditionFailed, "concurrent_modification", err.Error())
		default:
			h.logger.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("role_id", roleID.String()).
				Msg("Erro ao atualizar função")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Responder com a função atualizada
	h.respondWithJSON(w, http.StatusOK, toRoleResponse(updatedRole))
}

// DeleteRole exclui uma função
func (h *RoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.DeleteRole")
	defer span.End()

	tenantID := h.getTenantID(r)
	userID := h.getUserID(r)
	vars := mux.Vars(r)
	roleID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_id", "ID da função inválido")
		return
	}

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("role.id", roleID.String()),
		attribute.String("user.id", userID.String()),
	)

	// Parâmetro para exclusão permanente (hard delete)
	permanent := r.URL.Query().Get("permanent") == "true"

	// Excluir função no serviço
	if permanent {
		err = h.roleService.HardDeleteRole(ctx, tenantID, roleID, userID)
	} else {
		err = h.roleService.SoftDeleteRole(ctx, tenantID, roleID, userID)
	}

	if err != nil {
		span.SetStatus(codes.Error, "Falha ao excluir função")
		span.RecordError(err)

		// Mapear erros específicos do domínio para códigos HTTP apropriados
		switch err.(type) {
		case *application.ResourceNotFoundError:
			h.respondWithError(w, http.StatusNotFound, "not_found", err.Error())
		case *application.ValidationError:
			h.respondWithError(w, http.StatusBadRequest, "validation_error", err.Error())
		case *application.ResourceInUseError:
			h.respondWithError(w, http.StatusConflict, "resource_in_use", err.Error())
		case *application.OperationNotAllowedError:
			h.respondWithError(w, http.StatusForbidden, "operation_not_allowed", err.Error())
		default:
			h.logger.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("role_id", roleID.String()).
				Bool("permanent", permanent).
				Msg("Erro ao excluir função")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Responder com sucesso sem conteúdo
	w.WriteHeader(http.StatusNoContent)
}

// CloneRole clona uma função existente
func (h *RoleHandler) CloneRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.CloneRole")
	defer span.End()

	tenantID := h.getTenantID(r)
	userID := h.getUserID(r)
	vars := mux.Vars(r)
	sourceRoleID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_id", "ID da função fonte inválido")
		return
	}

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("source_role.id", sourceRoleID.String()),
		attribute.String("user.id", userID.String()),
	)

	type cloneRequest struct {
		NewCode        string `json:"newCode"`
		NewName        string `json:"newName"`
		CloneHierarchy bool   `json:"cloneHierarchy"`
		CloneUsers     bool   `json:"cloneUsers"`
	}

	var req cloneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.SetStatus(codes.Error, "Falha ao decodificar requisição")
		span.RecordError(err)
		h.respondWithError(w, http.StatusBadRequest, "invalid_request", "Formato de requisição inválido")
		return
	}

	// Validações básicas
	if req.NewCode == "" {
		h.respondWithError(w, http.StatusBadRequest, "invalid_request", "Novo código da função é obrigatório")
		return
	}
	if req.NewName == "" {
		h.respondWithError(w, http.StatusBadRequest, "invalid_request", "Novo nome da função é obrigatório")
		return
	}

	// Clonar função no serviço
	clonedRole, err := h.roleService.CloneRole(
		ctx, tenantID, sourceRoleID, req.NewCode, req.NewName, 
		req.CloneHierarchy, req.CloneUsers, userID,
	)

	if err != nil {
		span.SetStatus(codes.Error, "Falha ao clonar função")
		span.RecordError(err)

		// Mapear erros específicos do domínio para códigos HTTP apropriados
		switch err.(type) {
		case *application.ResourceNotFoundError:
			h.respondWithError(w, http.StatusNotFound, "not_found", err.Error())
		case *application.ValidationError:
			h.respondWithError(w, http.StatusBadRequest, "validation_error", err.Error())
		case *application.DuplicateResourceError:
			h.respondWithError(w, http.StatusConflict, "conflict", err.Error())
		default:
			h.logger.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("source_role_id", sourceRoleID.String()).
				Str("new_code", req.NewCode).
				Msg("Erro ao clonar função")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Responder com a função clonada
	h.respondWithJSON(w, http.StatusCreated, toRoleResponse(clonedRole))
}

// SyncSystemRoles sincroniza as funções de sistema
func (h *RoleHandler) SyncSystemRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.SyncSystemRoles")
	defer span.End()

	tenantID := h.getTenantID(r)
	userID := h.getUserID(r)

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("user.id", userID.String()),
	)

	// Sincronizar funções de sistema no serviço
	syncedRoles, err := h.roleService.SyncSystemRoles(ctx, tenantID, userID)
	if err != nil {
		span.SetStatus(codes.Error, "Falha ao sincronizar funções de sistema")
		span.RecordError(err)

		h.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("Erro ao sincronizar funções de sistema")
		h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		return
	}

	// Responder com as funções sincronizadas
	h.respondWithJSON(w, http.StatusOK, toRoleResponseList(syncedRoles))
}