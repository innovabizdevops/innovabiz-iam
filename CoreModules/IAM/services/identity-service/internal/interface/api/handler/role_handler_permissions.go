package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"innovabiz/iam/identity-service/internal/application"
	"innovabiz/iam/identity-service/internal/domain/model"
)

// PermissionResponse representa o modelo de dados para retorno de uma permissão
type PermissionResponse struct {
	ID          uuid.UUID              `json:"id"`
	TenantID    uuid.UUID              `json:"tenantId"`
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Category    string                 `json:"category"`
	IsSystem    bool                   `json:"isSystem"`
	IsActive    bool                   `json:"isActive"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	CreatedBy   uuid.UUID              `json:"createdBy"`
	UpdatedAt   *time.Time             `json:"updatedAt,omitempty"`
	UpdatedBy   *uuid.UUID             `json:"updatedBy,omitempty"`
	Version     int                    `json:"version"`
}

// toPermissionResponse converte um modelo de domínio Permission para PermissionResponse
func toPermissionResponse(permission *model.Permission) PermissionResponse {
	response := PermissionResponse{
		ID:          permission.ID,
		TenantID:    permission.TenantID,
		Code:        permission.Code,
		Name:        permission.Name,
		Description: permission.Description,
		Category:    permission.Category,
		IsSystem:    permission.IsSystem,
		IsActive:    permission.IsActive,
		Metadata:    permission.Metadata,
		CreatedAt:   permission.CreatedAt,
		CreatedBy:   permission.CreatedBy,
		Version:     permission.Version,
	}

	if !permission.UpdatedAt.IsZero() {
		response.UpdatedAt = &permission.UpdatedAt
		response.UpdatedBy = &permission.UpdatedBy
	}

	return response
}

// toPermissionResponseList converte uma lista de modelos de domínio Permission para []PermissionResponse
func toPermissionResponseList(permissions []*model.Permission) []PermissionResponse {
	responseList := make([]PermissionResponse, len(permissions))
	for i, permission := range permissions {
		responseList[i] = toPermissionResponse(permission)
	}
	return responseList
}

// GetRolePermissions obtém as permissões diretamente atribuídas a uma função
func (h *RoleHandler) GetRolePermissions(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.GetRolePermissions")
	defer span.End()

	tenantID := h.getTenantID(r)
	vars := mux.Vars(r)
	roleID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_id", "ID da função inválido")
		return
	}

	pagination := h.getPagination(r)

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("role.id", roleID.String()),
		attribute.Int("pagination.page", pagination.Page),
		attribute.Int("pagination.pageSize", pagination.PageSize),
	)

	permissions, totalCount, err := h.roleService.GetRolePermissions(ctx, tenantID, roleID, pagination)
	if err != nil {
		span.SetStatus(codes.Error, "Falha ao obter permissões da função")
		span.RecordError(err)

		// Mapear erros específicos do domínio para códigos HTTP apropriados
		switch err.(type) {
		case *application.ResourceNotFoundError:
			h.respondWithError(w, http.StatusNotFound, "not_found", err.Error())
		default:
			h.logger.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("role_id", roleID.String()).
				Msg("Erro ao obter permissões da função")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Calcular total de páginas
	totalPages := totalCount / int64(pagination.PageSize)
	if totalCount%int64(pagination.PageSize) > 0 {
		totalPages++
	}

	// Responder com a lista de permissões e informações de paginação
	h.respondWithJSON(w, http.StatusOK, response{
		Data: toPermissionResponseList(permissions),
		Pagination: &paginationResponse{
			Page:       pagination.Page,
			PageSize:   pagination.PageSize,
			TotalItems: totalCount,
			TotalPages: int(totalPages),
		},
	})
}

// GetAllRolePermissions obtém todas as permissões de uma função (incluindo as herdadas)
func (h *RoleHandler) GetAllRolePermissions(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.GetAllRolePermissions")
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

	permissions, err := h.roleService.GetAllPermissionsForRole(ctx, tenantID, roleID)
	if err != nil {
		span.SetStatus(codes.Error, "Falha ao obter todas as permissões da função")
		span.RecordError(err)

		// Mapear erros específicos do domínio para códigos HTTP apropriados
		switch err.(type) {
		case *application.ResourceNotFoundError:
			h.respondWithError(w, http.StatusNotFound, "not_found", err.Error())
		default:
			h.logger.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("role_id", roleID.String()).
				Msg("Erro ao obter todas as permissões da função")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Responder com a lista completa de permissões
	h.respondWithJSON(w, http.StatusOK, toPermissionResponseList(permissions))
}

// AssignPermission atribui uma permissão a uma função
func (h *RoleHandler) AssignPermission(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.AssignPermission")
	defer span.End()

	tenantID := h.getTenantID(r)
	userID := h.getUserID(r)
	vars := mux.Vars(r)
	
	roleID, err := uuid.Parse(vars["roleId"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_role_id", "ID da função inválido")
		return
	}
	
	permissionID, err := uuid.Parse(vars["permissionId"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_permission_id", "ID da permissão inválido")
		return
	}

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("role.id", roleID.String()),
		attribute.String("permission.id", permissionID.String()),
		attribute.String("user.id", userID.String()),
	)

	err = h.roleService.AssignPermission(ctx, tenantID, roleID, permissionID, userID)
	if err != nil {
		span.SetStatus(codes.Error, "Falha ao atribuir permissão")
		span.RecordError(err)

		// Mapear erros específicos do domínio para códigos HTTP apropriados
		switch err.(type) {
		case *application.ResourceNotFoundError:
			h.respondWithError(w, http.StatusNotFound, "not_found", err.Error())
		case *application.DuplicateResourceError:
			h.respondWithError(w, http.StatusConflict, "conflict", err.Error())
		default:
			h.logger.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("role_id", roleID.String()).
				Str("permission_id", permissionID.String()).
				Msg("Erro ao atribuir permissão")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Responder com sucesso sem conteúdo
	w.WriteHeader(http.StatusNoContent)
}

// RevokePermission remove uma permissão de uma função
func (h *RoleHandler) RevokePermission(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.RevokePermission")
	defer span.End()

	tenantID := h.getTenantID(r)
	userID := h.getUserID(r)
	vars := mux.Vars(r)
	
	roleID, err := uuid.Parse(vars["roleId"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_role_id", "ID da função inválido")
		return
	}
	
	permissionID, err := uuid.Parse(vars["permissionId"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_permission_id", "ID da permissão inválido")
		return
	}

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("role.id", roleID.String()),
		attribute.String("permission.id", permissionID.String()),
		attribute.String("user.id", userID.String()),
	)

	err = h.roleService.RevokePermission(ctx, tenantID, roleID, permissionID, userID)
	if err != nil {
		span.SetStatus(codes.Error, "Falha ao revogar permissão")
		span.RecordError(err)

		// Mapear erros específicos do domínio para códigos HTTP apropriados
		switch err.(type) {
		case *application.ResourceNotFoundError:
			h.respondWithError(w, http.StatusNotFound, "not_found", err.Error())
		case *application.ValidationError:
			h.respondWithError(w, http.StatusBadRequest, "validation_error", err.Error())
		default:
			h.logger.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("role_id", roleID.String()).
				Str("permission_id", permissionID.String()).
				Msg("Erro ao revogar permissão")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Responder com sucesso sem conteúdo
	w.WriteHeader(http.StatusNoContent)
}

// CheckPermission verifica se uma função tem uma permissão específica
func (h *RoleHandler) CheckPermission(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.CheckPermission")
	defer span.End()

	tenantID := h.getTenantID(r)
	vars := mux.Vars(r)
	
	roleID, err := uuid.Parse(vars["roleId"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_role_id", "ID da função inválido")
		return
	}
	
	permissionID, err := uuid.Parse(vars["permissionId"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_permission_id", "ID da permissão inválido")
		return
	}

	// Parâmetro para verificar apenas permissões diretas
	directOnly := r.URL.Query().Get("directOnly") == "true"

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("role.id", roleID.String()),
		attribute.String("permission.id", permissionID.String()),
		attribute.Bool("direct_only", directOnly),
	)

	var hasPermission bool
	var checkErr error

	if directOnly {
		hasPermission, checkErr = h.roleService.HasDirectPermission(ctx, tenantID, roleID, permissionID)
	} else {
		hasPermission, checkErr = h.roleService.HasPermission(ctx, tenantID, roleID, permissionID)
	}

	if checkErr != nil {
		span.SetStatus(codes.Error, "Falha ao verificar permissão")
		span.RecordError(checkErr)

		// Mapear erros específicos do domínio para códigos HTTP apropriados
		switch checkErr.(type) {
		case *application.ResourceNotFoundError:
			h.respondWithError(w, http.StatusNotFound, "not_found", checkErr.Error())
		default:
			h.logger.Error().Err(checkErr).
				Str("tenant_id", tenantID.String()).
				Str("role_id", roleID.String()).
				Str("permission_id", permissionID.String()).
				Bool("direct_only", directOnly).
				Msg("Erro ao verificar permissão")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Responder com o resultado da verificação
	h.respondWithJSON(w, http.StatusOK, map[string]bool{
		"hasPermission": hasPermission,
	})
}