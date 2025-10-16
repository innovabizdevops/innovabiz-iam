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

// UserResponse representa o modelo de dados para retorno de um usuário
type UserResponse struct {
	ID          uuid.UUID              `json:"id"`
	TenantID    uuid.UUID              `json:"tenantId"`
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	FirstName   string                 `json:"firstName,omitempty"`
	LastName    string                 `json:"lastName,omitempty"`
	DisplayName string                 `json:"displayName"`
	IsActive    bool                   `json:"isActive"`
	IsSystem    bool                   `json:"isSystem"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	CreatedBy   uuid.UUID              `json:"createdBy"`
	UpdatedAt   *time.Time             `json:"updatedAt,omitempty"`
	UpdatedBy   *uuid.UUID             `json:"updatedBy,omitempty"`
	Version     int                    `json:"version"`
}

// UserRoleResponse representa o modelo de dados para retorno de um usuário com informações de expiração
type UserRoleResponse struct {
	User      UserResponse `json:"user"`
	ExpiresAt *time.Time   `json:"expiresAt,omitempty"`
	AssignedAt time.Time   `json:"assignedAt"`
	AssignedBy uuid.UUID   `json:"assignedBy"`
}

// RoleUserResponse representa o modelo de dados para retorno de uma função com informações de expiração
type RoleUserResponse struct {
	Role      RoleResponse `json:"role"`
	ExpiresAt *time.Time   `json:"expiresAt,omitempty"`
	AssignedAt time.Time   `json:"assignedAt"`
	AssignedBy uuid.UUID   `json:"assignedBy"`
}

// toUserResponse converte um modelo de domínio User para UserResponse
func toUserResponse(user *model.User) UserResponse {
	response := UserResponse{
		ID:          user.ID,
		TenantID:    user.TenantID,
		Username:    user.Username,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DisplayName: user.DisplayName,
		IsActive:    user.IsActive,
		IsSystem:    user.IsSystem,
		Metadata:    user.Metadata,
		CreatedAt:   user.CreatedAt,
		CreatedBy:   user.CreatedBy,
		Version:     user.Version,
	}

	if !user.UpdatedAt.IsZero() {
		response.UpdatedAt = &user.UpdatedAt
		response.UpdatedBy = &user.UpdatedBy
	}

	return response
}

// toUserRoleResponseList converte uma lista de modelos de domínio UserWithExpiration para []UserRoleResponse
func toUserRoleResponseList(usersWithExpiration []*model.UserWithExpiration) []UserRoleResponse {
	responseList := make([]UserRoleResponse, len(usersWithExpiration))
	for i, userWithExp := range usersWithExpiration {
		responseList[i] = UserRoleResponse{
			User:      toUserResponse(userWithExp.User),
			ExpiresAt: userWithExp.ExpiresAt,
			AssignedAt: userWithExp.AssignedAt,
			AssignedBy: userWithExp.AssignedBy,
		}
	}
	return responseList
}

// toRoleUserResponseList converte uma lista de modelos de domínio RoleWithExpiration para []RoleUserResponse
func toRoleUserResponseList(rolesWithExpiration []*model.RoleWithExpiration) []RoleUserResponse {
	responseList := make([]RoleUserResponse, len(rolesWithExpiration))
	for i, roleWithExp := range rolesWithExpiration {
		responseList[i] = RoleUserResponse{
			Role:      toRoleResponse(roleWithExp.Role),
			ExpiresAt: roleWithExp.ExpiresAt,
			AssignedAt: roleWithExp.AssignedAt,
			AssignedBy: roleWithExp.AssignedBy,
		}
	}
	return responseList
}

// GetRoleUsers obtém os usuários atribuídos a uma função
func (h *RoleHandler) GetRoleUsers(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.GetRoleUsers")
	defer span.End()

	tenantID := h.getTenantID(r)
	vars := mux.Vars(r)
	roleID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_id", "ID da função inválido")
		return
	}

	pagination := h.getPagination(r)
	
	// Extrair parâmetros de filtro
	includeExpired := r.URL.Query().Get("includeExpired") == "true"

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("role.id", roleID.String()),
		attribute.Bool("include_expired", includeExpired),
		attribute.Int("pagination.page", pagination.Page),
		attribute.Int("pagination.pageSize", pagination.PageSize),
	)

	users, totalCount, err := h.roleService.GetUsersInRole(ctx, tenantID, roleID, includeExpired, pagination)
	if err != nil {
		span.SetStatus(codes.Error, "Falha ao obter usuários da função")
		span.RecordError(err)

		// Mapear erros específicos do domínio para códigos HTTP apropriados
		switch err.(type) {
		case *application.ResourceNotFoundError:
			h.respondWithError(w, http.StatusNotFound, "not_found", err.Error())
		default:
			h.logger.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("role_id", roleID.String()).
				Msg("Erro ao obter usuários da função")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Calcular total de páginas
	totalPages := totalCount / int64(pagination.PageSize)
	if totalCount%int64(pagination.PageSize) > 0 {
		totalPages++
	}

	// Responder com a lista de usuários e informações de paginação
	h.respondWithJSON(w, http.StatusOK, response{
		Data: toUserRoleResponseList(users),
		Pagination: &paginationResponse{
			Page:       pagination.Page,
			PageSize:   pagination.PageSize,
			TotalItems: totalCount,
			TotalPages: int(totalPages),
		},
	})
}

// GetUserRoles obtém as funções atribuídas a um usuário
func (h *RoleHandler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.GetUserRoles")
	defer span.End()

	tenantID := h.getTenantID(r)
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["userId"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_user_id", "ID do usuário inválido")
		return
	}

	pagination := h.getPagination(r)
	
	// Extrair parâmetros de filtro
	includeExpired := r.URL.Query().Get("includeExpired") == "true"

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("user.id", userID.String()),
		attribute.Bool("include_expired", includeExpired),
		attribute.Int("pagination.page", pagination.Page),
		attribute.Int("pagination.pageSize", pagination.PageSize),
	)

	roles, totalCount, err := h.roleService.GetUserRoles(ctx, tenantID, userID, includeExpired, pagination)
	if err != nil {
		span.SetStatus(codes.Error, "Falha ao obter funções do usuário")
		span.RecordError(err)

		// Mapear erros específicos do domínio para códigos HTTP apropriados
		switch err.(type) {
		case *application.ResourceNotFoundError:
			h.respondWithError(w, http.StatusNotFound, "not_found", err.Error())
		default:
			h.logger.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("user_id", userID.String()).
				Msg("Erro ao obter funções do usuário")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Calcular total de páginas
	totalPages := totalCount / int64(pagination.PageSize)
	if totalCount%int64(pagination.PageSize) > 0 {
		totalPages++
	}

	// Responder com a lista de funções e informações de paginação
	h.respondWithJSON(w, http.StatusOK, response{
		Data: toRoleUserResponseList(roles),
		Pagination: &paginationResponse{
			Page:       pagination.Page,
			PageSize:   pagination.PageSize,
			TotalItems: totalCount,
			TotalPages: int(totalPages),
		},
	})
}

// GetAllUserRoles obtém todas as funções atribuídas a um usuário (incluindo as herdadas)
func (h *RoleHandler) GetAllUserRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.GetAllUserRoles")
	defer span.End()

	tenantID := h.getTenantID(r)
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["userId"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_user_id", "ID do usuário inválido")
		return
	}

	// Extrair parâmetros de filtro
	includeExpired := r.URL.Query().Get("includeExpired") == "true"

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("user.id", userID.String()),
		attribute.Bool("include_expired", includeExpired),
	)

	roles, err := h.roleService.GetAllUserRoles(ctx, tenantID, userID, includeExpired)
	if err != nil {
		span.SetStatus(codes.Error, "Falha ao obter todas as funções do usuário")
		span.RecordError(err)

		// Mapear erros específicos do domínio para códigos HTTP apropriados
		switch err.(type) {
		case *application.ResourceNotFoundError:
			h.respondWithError(w, http.StatusNotFound, "not_found", err.Error())
		default:
			h.logger.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("user_id", userID.String()).
				Msg("Erro ao obter todas as funções do usuário")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Responder com a lista completa de funções
	h.respondWithJSON(w, http.StatusOK, toRoleResponseList(roles))
}

// AssignUserToRole atribui um usuário a uma função
func (h *RoleHandler) AssignUserToRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.AssignUserToRole")
	defer span.End()

	tenantID := h.getTenantID(r)
	assignedBy := h.getUserID(r)
	vars := mux.Vars(r)
	
	roleID, err := uuid.Parse(vars["roleId"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_role_id", "ID da função inválido")
		return
	}
	
	userID, err := uuid.Parse(vars["userId"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_user_id", "ID do usuário inválido")
		return
	}

	// Extrair parâmetros adicionais da requisição
	type assignRequest struct {
		ExpiresAt string `json:"expiresAt,omitempty"`
	}

	var req assignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != http.ErrBodyReadCloser {
		span.SetStatus(codes.Error, "Falha ao decodificar requisição")
		span.RecordError(err)
		h.respondWithError(w, http.StatusBadRequest, "invalid_request", "Formato de requisição inválido")
		return
	}

	// Analisar data de expiração, se fornecida
	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		parsedTime, err := h.parseExpirationTime(req.ExpiresAt)
		if err != nil {
			h.respondWithError(w, http.StatusBadRequest, "invalid_expiration", "Formato de data de expiração inválido")
			return
		}
		expiresAt = parsedTime
	}

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("role.id", roleID.String()),
		attribute.String("user.id", userID.String()),
		attribute.String("assigned_by", assignedBy.String()),
	)

	if expiresAt != nil {
		span.SetAttributes(attribute.String("expires_at", expiresAt.Format(time.RFC3339)))
	}

	err = h.roleService.AssignUserToRole(ctx, tenantID, roleID, userID, expiresAt, assignedBy)
	if err != nil {
		span.SetStatus(codes.Error, "Falha ao atribuir usuário à função")
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
				Str("role_id", roleID.String()).
				Str("user_id", userID.String()).
				Msg("Erro ao atribuir usuário à função")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Responder com sucesso sem conteúdo
	w.WriteHeader(http.StatusNoContent)
}

// UpdateUserRoleExpiration atualiza a data de expiração da atribuição de um usuário a uma função
func (h *RoleHandler) UpdateUserRoleExpiration(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.UpdateUserRoleExpiration")
	defer span.End()

	tenantID := h.getTenantID(r)
	updatedBy := h.getUserID(r)
	vars := mux.Vars(r)
	
	roleID, err := uuid.Parse(vars["roleId"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_role_id", "ID da função inválido")
		return
	}
	
	userID, err := uuid.Parse(vars["userId"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_user_id", "ID do usuário inválido")
		return
	}

	// Extrair parâmetros adicionais da requisição
	type updateExpirationRequest struct {
		ExpiresAt string `json:"expiresAt,omitempty"`
	}

	var req updateExpirationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.SetStatus(codes.Error, "Falha ao decodificar requisição")
		span.RecordError(err)
		h.respondWithError(w, http.StatusBadRequest, "invalid_request", "Formato de requisição inválido")
		return
	}

	// Analisar data de expiração
	var expiresAt *time.Time
	if req.ExpiresAt == "" {
		// Se não for fornecida, remover a expiração
		expiresAt = nil
	} else {
		parsedTime, err := h.parseExpirationTime(req.ExpiresAt)
		if err != nil {
			h.respondWithError(w, http.StatusBadRequest, "invalid_expiration", "Formato de data de expiração inválido")
			return
		}
		expiresAt = parsedTime
	}

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("role.id", roleID.String()),
		attribute.String("user.id", userID.String()),
		attribute.String("updated_by", updatedBy.String()),
	)

	if expiresAt != nil {
		span.SetAttributes(attribute.String("expires_at", expiresAt.Format(time.RFC3339)))
	}

	err = h.roleService.UpdateUserRoleExpiration(ctx, tenantID, roleID, userID, expiresAt, updatedBy)
	if err != nil {
		span.SetStatus(codes.Error, "Falha ao atualizar expiração da atribuição do usuário")
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
				Str("user_id", userID.String()).
				Msg("Erro ao atualizar expiração da atribuição do usuário")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Responder com sucesso sem conteúdo
	w.WriteHeader(http.StatusNoContent)
}

// RemoveUserFromRole remove um usuário de uma função
func (h *RoleHandler) RemoveUserFromRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.RemoveUserFromRole")
	defer span.End()

	tenantID := h.getTenantID(r)
	removedBy := h.getUserID(r)
	vars := mux.Vars(r)
	
	roleID, err := uuid.Parse(vars["roleId"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_role_id", "ID da função inválido")
		return
	}
	
	userID, err := uuid.Parse(vars["userId"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_user_id", "ID do usuário inválido")
		return
	}

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("role.id", roleID.String()),
		attribute.String("user.id", userID.String()),
		attribute.String("removed_by", removedBy.String()),
	)

	err = h.roleService.RemoveUserFromRole(ctx, tenantID, roleID, userID, removedBy)
	if err != nil {
		span.SetStatus(codes.Error, "Falha ao remover usuário da função")
		span.RecordError(err)

		// Mapear erros específicos do domínio para códigos HTTP apropriados
		switch err.(type) {
		case *application.ResourceNotFoundError:
			h.respondWithError(w, http.StatusNotFound, "not_found", err.Error())
		default:
			h.logger.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("role_id", roleID.String()).
				Str("user_id", userID.String()).
				Msg("Erro ao remover usuário da função")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Responder com sucesso sem conteúdo
	w.WriteHeader(http.StatusNoContent)
}

// CheckUserInRole verifica se um usuário está em uma função
func (h *RoleHandler) CheckUserInRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RoleHandler.CheckUserInRole")
	defer span.End()

	tenantID := h.getTenantID(r)
	vars := mux.Vars(r)
	
	roleID, err := uuid.Parse(vars["roleId"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_role_id", "ID da função inválido")
		return
	}
	
	userID, err := uuid.Parse(vars["userId"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_user_id", "ID do usuário inválido")
		return
	}

	// Parâmetros para filtros adicionais
	directOnly := r.URL.Query().Get("directOnly") == "true"
	includeExpired := r.URL.Query().Get("includeExpired") == "true"

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("role.id", roleID.String()),
		attribute.String("user.id", userID.String()),
		attribute.Bool("direct_only", directOnly),
		attribute.Bool("include_expired", includeExpired),
	)

	var isInRole bool
	var checkErr error

	if directOnly {
		isInRole, checkErr = h.roleService.IsUserDirectlyInRole(ctx, tenantID, userID, roleID, includeExpired)
	} else {
		isInRole, checkErr = h.roleService.IsUserInRole(ctx, tenantID, userID, roleID, includeExpired)
	}

	if checkErr != nil {
		span.SetStatus(codes.Error, "Falha ao verificar usuário na função")
		span.RecordError(checkErr)

		// Mapear erros específicos do domínio para códigos HTTP apropriados
		switch checkErr.(type) {
		case *application.ResourceNotFoundError:
			h.respondWithError(w, http.StatusNotFound, "not_found", checkErr.Error())
		default:
			h.logger.Error().Err(checkErr).
				Str("tenant_id", tenantID.String()).
				Str("role_id", roleID.String()).
				Str("user_id", userID.String()).
				Bool("direct_only", directOnly).
				Bool("include_expired", includeExpired).
				Msg("Erro ao verificar usuário na função")
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno ao processar a requisição")
		}
		return
	}

	// Responder com o resultado da verificação
	h.respondWithJSON(w, http.StatusOK, map[string]bool{
		"isInRole": isInRole,
	})
}