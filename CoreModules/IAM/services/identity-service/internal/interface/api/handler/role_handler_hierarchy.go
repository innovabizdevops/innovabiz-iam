package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"innovabiz/iam/identity-service/internal/application"
)

// GetChildRoles obtém as funções filhas diretas de uma função
func (h *RoleHandler) GetChildRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.GetChildRoles")
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

	childRoles, totalCount, err := h.roleService.GetChildRoles(ctx, tenantID, roleID, pagination)
	if err != nil {
		span.SetStatus(codes.Error, "Falha ao obter funções filhas")
		span.RecordError(err)

		// Mapear erros específicos do domínio para códigos HTTP apropriados
		switch err.(type) {
		case *application.ResourceNotFoundError:
			h.respondWithError(w, http.StatusNotFound, "not_found", err.Error())
		default:
			h.logger.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("role_id", roleID.String()).
				Msg("Erro ao obter funções filhas")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Calcular total de páginas
	totalPages := totalCount / int64(pagination.PageSize)
	if totalCount%int64(pagination.PageSize) > 0 {
		totalPages++
	}

	// Responder com a lista de funções filhas e informações de paginação
	h.respondWithJSON(w, http.StatusOK, response{
		Data: toRoleResponseList(childRoles),
		Pagination: &paginationResponse{
			Page:       pagination.Page,
			PageSize:   pagination.PageSize,
			TotalItems: totalCount,
			TotalPages: int(totalPages),
		},
	})
}

// GetParentRoles obtém as funções pais diretas de uma função
func (h *RoleHandler) GetParentRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.GetParentRoles")
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

	parentRoles, totalCount, err := h.roleService.GetParentRoles(ctx, tenantID, roleID, pagination)
	if err != nil {
		span.SetStatus(codes.Error, "Falha ao obter funções pais")
		span.RecordError(err)

		// Mapear erros específicos do domínio para códigos HTTP apropriados
		switch err.(type) {
		case *application.ResourceNotFoundError:
			h.respondWithError(w, http.StatusNotFound, "not_found", err.Error())
		default:
			h.logger.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("role_id", roleID.String()).
				Msg("Erro ao obter funções pais")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Calcular total de páginas
	totalPages := totalCount / int64(pagination.PageSize)
	if totalCount%int64(pagination.PageSize) > 0 {
		totalPages++
	}

	// Responder com a lista de funções pais e informações de paginação
	h.respondWithJSON(w, http.StatusOK, response{
		Data: toRoleResponseList(parentRoles),
		Pagination: &paginationResponse{
			Page:       pagination.Page,
			PageSize:   pagination.PageSize,
			TotalItems: totalCount,
			TotalPages: int(totalPages),
		},
	})
}

// GetDescendantRoles obtém todas as funções descendentes de uma função
func (h *RoleHandler) GetDescendantRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.GetDescendantRoles")
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

	// Parâmetro para incluir a informação de profundidade na resposta
	includeDepth := r.URL.Query().Get("includeDepth") == "true"

	descendantRoles, err := h.roleService.GetDescendantRoles(ctx, tenantID, roleID, includeDepth)
	if err != nil {
		span.SetStatus(codes.Error, "Falha ao obter funções descendentes")
		span.RecordError(err)

		// Mapear erros específicos do domínio para códigos HTTP apropriados
		switch err.(type) {
		case *application.ResourceNotFoundError:
			h.respondWithError(w, http.StatusNotFound, "not_found", err.Error())
		default:
			h.logger.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("role_id", roleID.String()).
				Msg("Erro ao obter funções descendentes")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Se includeDepth for true, precisamos formatar a resposta para incluir a profundidade
	if includeDepth {
		type roleWithDepth struct {
			Role  RoleResponse `json:"role"`
			Depth int          `json:"depth"`
		}

		// Formatar os resultados para incluir a profundidade
		rolesWithDepth := make([]roleWithDepth, len(descendantRoles))
		for i, descendant := range descendantRoles {
			rolesWithDepth[i] = roleWithDepth{
				Role:  toRoleResponse(descendant.Role),
				Depth: descendant.Depth,
			}
		}

		h.respondWithJSON(w, http.StatusOK, rolesWithDepth)
		return
	}

	// Caso contrário, retornar apenas as funções
	roles := make([]*model.Role, len(descendantRoles))
	for i, descendant := range descendantRoles {
		roles[i] = descendant.Role
	}

	h.respondWithJSON(w, http.StatusOK, toRoleResponseList(roles))
}

// GetAncestorRoles obtém todas as funções ancestrais de uma função
func (h *RoleHandler) GetAncestorRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.GetAncestorRoles")
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

	// Parâmetro para incluir a informação de profundidade na resposta
	includeDepth := r.URL.Query().Get("includeDepth") == "true"

	ancestorRoles, err := h.roleService.GetAncestorRoles(ctx, tenantID, roleID, includeDepth)
	if err != nil {
		span.SetStatus(codes.Error, "Falha ao obter funções ancestrais")
		span.RecordError(err)

		// Mapear erros específicos do domínio para códigos HTTP apropriados
		switch err.(type) {
		case *application.ResourceNotFoundError:
			h.respondWithError(w, http.StatusNotFound, "not_found", err.Error())
		default:
			h.logger.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("role_id", roleID.String()).
				Msg("Erro ao obter funções ancestrais")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Se includeDepth for true, precisamos formatar a resposta para incluir a profundidade
	if includeDepth {
		type roleWithDepth struct {
			Role  RoleResponse `json:"role"`
			Depth int          `json:"depth"`
		}

		// Formatar os resultados para incluir a profundidade
		rolesWithDepth := make([]roleWithDepth, len(ancestorRoles))
		for i, ancestor := range ancestorRoles {
			rolesWithDepth[i] = roleWithDepth{
				Role:  toRoleResponse(ancestor.Role),
				Depth: ancestor.Depth,
			}
		}

		h.respondWithJSON(w, http.StatusOK, rolesWithDepth)
		return
	}

	// Caso contrário, retornar apenas as funções
	roles := make([]*model.Role, len(ancestorRoles))
	for i, ancestor := range ancestorRoles {
		roles[i] = ancestor.Role
	}

	h.respondWithJSON(w, http.StatusOK, toRoleResponseList(roles))
}

// AssignChildRole atribui uma função filha a uma função pai
func (h *RoleHandler) AssignChildRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.AssignChildRole")
	defer span.End()

	tenantID := h.getTenantID(r)
	userID := h.getUserID(r)
	vars := mux.Vars(r)
	
	parentID, err := uuid.Parse(vars["parentId"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_parent_id", "ID da função pai inválido")
		return
	}
	
	childID, err := uuid.Parse(vars["childId"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_child_id", "ID da função filha inválido")
		return
	}

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("parent_role.id", parentID.String()),
		attribute.String("child_role.id", childID.String()),
		attribute.String("user.id", userID.String()),
	)

	err = h.roleService.AssignChildRole(ctx, tenantID, parentID, childID, userID)
	if err != nil {
		span.SetStatus(codes.Error, "Falha ao atribuir função filha")
		span.RecordError(err)

		// Mapear erros específicos do domínio para códigos HTTP apropriados
		switch err.(type) {
		case *application.ResourceNotFoundError:
			h.respondWithError(w, http.StatusNotFound, "not_found", err.Error())
		case *application.ValidationError:
			h.respondWithError(w, http.StatusBadRequest, "validation_error", err.Error())
		case *application.CyclicReferenceError:
			h.respondWithError(w, http.StatusBadRequest, "cyclic_reference", err.Error())
		case *application.IncompatibleTypesError:
			h.respondWithError(w, http.StatusBadRequest, "incompatible_types", err.Error())
		case *application.DuplicateResourceError:
			h.respondWithError(w, http.StatusConflict, "conflict", err.Error())
		default:
			h.logger.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("parent_id", parentID.String()).
				Str("child_id", childID.String()).
				Msg("Erro ao atribuir função filha")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Responder com sucesso sem conteúdo
	w.WriteHeader(http.StatusNoContent)
}

// RemoveChildRole remove uma função filha de uma função pai
func (h *RoleHandler) RemoveChildRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.RemoveChildRole")
	defer span.End()

	tenantID := h.getTenantID(r)
	userID := h.getUserID(r)
	vars := mux.Vars(r)
	
	parentID, err := uuid.Parse(vars["parentId"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_parent_id", "ID da função pai inválido")
		return
	}
	
	childID, err := uuid.Parse(vars["childId"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_child_id", "ID da função filha inválido")
		return
	}

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("parent_role.id", parentID.String()),
		attribute.String("child_role.id", childID.String()),
		attribute.String("user.id", userID.String()),
	)

	err = h.roleService.RemoveChildRole(ctx, tenantID, parentID, childID, userID)
	if err != nil {
		span.SetStatus(codes.Error, "Falha ao remover função filha")
		span.RecordError(err)

		// Mapear erros específicos do domínio para códigos HTTP apropriados
		switch err.(type) {
		case *application.ResourceNotFoundError:
			h.respondWithError(w, http.StatusNotFound, "not_found", err.Error())
		default:
			h.logger.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("parent_id", parentID.String()).
				Str("child_id", childID.String()).
				Msg("Erro ao remover função filha")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Responder com sucesso sem conteúdo
	w.WriteHeader(http.StatusNoContent)
}