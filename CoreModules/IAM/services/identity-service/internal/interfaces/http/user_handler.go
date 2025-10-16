/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Handler HTTP para rotas de gerenciamento de usuários.
 * Implementa os endpoints REST para operações CRUD e outras operações relacionadas a usuários.
 */

package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/innovabiz/iam/services/identity-service/internal/application"
	"github.com/innovabiz/iam/services/identity-service/internal/infrastructure/middleware"
)

// UserHandler gerencia os endpoints HTTP relacionados a usuários
type UserHandler struct {
	userService application.UserService
	validator   *validator.Validate
	tracer      trace.Tracer
}

// NewUserHandler cria uma nova instância do UserHandler
func NewUserHandler(userService application.UserService, tracer trace.Tracer) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   validator.New(),
		tracer:      tracer,
	}
}

// RegisterRoutes registra as rotas HTTP para o UserHandler
func (h *UserHandler) RegisterRoutes(r chi.Router) {
	// Rotas protegidas por autenticação
	r.Route("/users", func(r chi.Router) {
		// Middleware para autenticação
		r.Use(middleware.AuthMiddleware)

		// Métodos GET
		r.Get("/", h.ListUsers)                  // Listar todos os usuários
		r.Get("/{userId}", h.GetUserByID)        // Obter usuário por ID
		r.Get("/username/{username}", h.GetUserByUsername) // Obter usuário por nome de usuário
		r.Get("/email/{email}", h.GetUserByEmail)         // Obter usuário por email

		// Métodos POST
		r.Post("/", h.CreateUser)                // Criar novo usuário
		
		// Métodos PUT
		r.Put("/{userId}", h.UpdateUser)         // Atualizar usuário
		
		// Métodos DELETE
		r.Delete("/{userId}", h.DeleteUser)      // Excluir usuário (soft delete)
		
		// Rotas para endereços
		r.Route("/{userId}/addresses", func(r chi.Router) {
			r.Get("/", h.ListUserAddresses)
			r.Post("/", h.AddUserAddress)
			r.Put("/{addressId}", h.UpdateUserAddress)
			r.Delete("/{addressId}", h.DeleteUserAddress)
		})
		
		// Rotas para contatos
		r.Route("/{userId}/contacts", func(r chi.Router) {
			r.Get("/", h.ListUserContacts)
			r.Post("/", h.AddUserContact)
			r.Put("/{contactId}", h.UpdateUserContact)
			r.Delete("/{contactId}", h.DeleteUserContact)
		})
		
		// Rotas para funções
		r.Route("/{userId}/roles", func(r chi.Router) {
			r.Get("/", h.ListUserRoles)
			r.Post("/", h.AssignRolesToUser)
			r.Delete("/", h.RevokeRolesFromUser)
		})
	})
}

// validateRequest valida uma estrutura de requisição com o validador
func (h *UserHandler) validateRequest(req interface{}) error {
	if err := h.validator.Struct(req); err != nil {
		return err
	}
	return nil
}

// getQueryUUID extrai um UUID de um parâmetro de consulta
func (h *UserHandler) getQueryUUID(r *http.Request, param string) (uuid.UUID, error) {
	idStr := chi.URLParam(r, param)
	if idStr == "" {
		return uuid.Nil, errors.New("parâmetro não fornecido")
	}
	
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, err
	}
	
	return id, nil
}

// getTenantIDFromContext extrai o ID do tenant do contexto
func (h *UserHandler) getTenantIDFromContext(r *http.Request) (uuid.UUID, error) {
	tenantIDStr := r.Context().Value(middleware.TenantIDKey)
	if tenantIDStr == nil {
		return uuid.Nil, errors.New("tenant ID não encontrado no contexto")
	}
	
	tenantID, ok := tenantIDStr.(string)
	if !ok {
		return uuid.Nil, errors.New("tenant ID inválido")
	}
	
	return uuid.Parse(tenantID)
}

// getUserIDFromContext extrai o ID do usuário do contexto
func (h *UserHandler) getUserIDFromContext(r *http.Request) (uuid.UUID, error) {
	userIDStr := r.Context().Value(middleware.UserIDKey)
	if userIDStr == nil {
		return uuid.Nil, errors.New("user ID não encontrado no contexto")
	}
	
	userID, ok := userIDStr.(string)
	if !ok {
		return uuid.Nil, errors.New("user ID inválido")
	}
	
	return uuid.Parse(userID)
}

// ListUsers lista usuários com filtros e paginação
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "users.list",
		trace.WithAttributes(
			attribute.String("handler", "user_handler"),
			attribute.String("method", "ListUsers"),
		),
	)
	defer span.End()
	
	reqID := getRequestID(r)
	
	// Extrair tenant ID do contexto
	tenantID, err := h.getTenantIDFromContext(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_tenant", "Tenant ID inválido", reqID)
		return
	}
	
	// Extrair parâmetros de paginação e filtros
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10 // Tamanho padrão da página
	}
	
	// Extrair outros filtros
	filter := application.UserFilter{
		TenantID:   tenantID,
		Page:       page,
		PageSize:   pageSize,
		Username:   r.URL.Query().Get("username"),
		Email:      r.URL.Query().Get("email"),
		Status:     r.URL.Query().Get("status"),
		FirstName:  r.URL.Query().Get("firstName"),
		LastName:   r.URL.Query().Get("lastName"),
		SearchTerm: r.URL.Query().Get("search"),
		OrderBy:    r.URL.Query().Get("orderBy"),
		Order:      r.URL.Query().Get("order"),
	}
	
	// Validar ordem
	if filter.Order != "" && filter.Order != "asc" && filter.Order != "desc" {
		filter.Order = "asc"
	}
	
	// Chamar o serviço para listar usuários
	response, err := h.userService.ListUsers(ctx, filter)
	if err != nil {
		// Verificar o tipo de erro para enviar a resposta apropriada
		var appErr application.AppError
		if errors.As(err, &appErr) {
			respondWithError(w, appErr.StatusCode, appErr.Code, appErr.Message, reqID)
		} else {
			respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno no servidor", reqID)
		}
		log.Error().Err(err).Str("request_id", reqID).Msg("Erro ao listar usuários")
		return
	}
	
	// Responder com a lista de usuários
	respondWithJSON(w, http.StatusOK, response)
}

// GetUserByID obtém um usuário pelo ID
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "users.get_by_id",
		trace.WithAttributes(
			attribute.String("handler", "user_handler"),
			attribute.String("method", "GetUserByID"),
		),
	)
	defer span.End()
	
	reqID := getRequestID(r)
	
	// Extrair tenant ID do contexto
	tenantID, err := h.getTenantIDFromContext(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_tenant", "Tenant ID inválido", reqID)
		return
	}
	
	// Extrair user ID da URL
	userID, err := h.getQueryUUID(r, "userId")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_user_id", "ID de usuário inválido", reqID)
		return
	}
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("user_id", userID.String()),
	)
	
	// Chamar o serviço para obter o usuário
	user, err := h.userService.GetUserByID(ctx, tenantID, userID)
	if err != nil {
		// Verificar o tipo de erro para enviar a resposta apropriada
		var appErr application.AppError
		if errors.As(err, &appErr) {
			respondWithError(w, appErr.StatusCode, appErr.Code, appErr.Message, reqID)
		} else {
			respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno no servidor", reqID)
		}
		log.Error().Err(err).Str("request_id", reqID).Msg("Erro ao obter usuário por ID")
		return
	}
	
	// Responder com o usuário
	respondWithJSON(w, http.StatusOK, user)
}

// GetUserByUsername obtém um usuário pelo nome de usuário
func (h *UserHandler) GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "users.get_by_username",
		trace.WithAttributes(
			attribute.String("handler", "user_handler"),
			attribute.String("method", "GetUserByUsername"),
		),
	)
	defer span.End()
	
	reqID := getRequestID(r)
	
	// Extrair tenant ID do contexto
	tenantID, err := h.getTenantIDFromContext(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_tenant", "Tenant ID inválido", reqID)
		return
	}
	
	// Extrair username da URL
	username := chi.URLParam(r, "username")
	if username == "" {
		respondWithError(w, http.StatusBadRequest, "invalid_username", "Nome de usuário não fornecido", reqID)
		return
	}
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("username", username),
	)
	
	// Chamar o serviço para obter o usuário
	user, err := h.userService.GetUserByUsername(ctx, tenantID, username)
	if err != nil {
		// Verificar o tipo de erro para enviar a resposta apropriada
		var appErr application.AppError
		if errors.As(err, &appErr) {
			respondWithError(w, appErr.StatusCode, appErr.Code, appErr.Message, reqID)
		} else {
			respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno no servidor", reqID)
		}
		log.Error().Err(err).Str("request_id", reqID).Msg("Erro ao obter usuário por nome de usuário")
		return
	}
	
	// Responder com o usuário
	respondWithJSON(w, http.StatusOK, user)
}

// GetUserByEmail obtém um usuário pelo email
func (h *UserHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "users.get_by_email",
		trace.WithAttributes(
			attribute.String("handler", "user_handler"),
			attribute.String("method", "GetUserByEmail"),
		),
	)
	defer span.End()
	
	reqID := getRequestID(r)
	
	// Extrair tenant ID do contexto
	tenantID, err := h.getTenantIDFromContext(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_tenant", "Tenant ID inválido", reqID)
		return
	}
	
	// Extrair email da URL
	email := chi.URLParam(r, "email")
	if email == "" {
		respondWithError(w, http.StatusBadRequest, "invalid_email", "Email não fornecido", reqID)
		return
	}
	
	span.SetAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("email", email),
	)
	
	// Chamar o serviço para obter o usuário
	user, err := h.userService.GetUserByEmail(ctx, tenantID, email)
	if err != nil {
		// Verificar o tipo de erro para enviar a resposta apropriada
		var appErr application.AppError
		if errors.As(err, &appErr) {
			respondWithError(w, appErr.StatusCode, appErr.Code, appErr.Message, reqID)
		} else {
			respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno no servidor", reqID)
		}
		log.Error().Err(err).Str("request_id", reqID).Msg("Erro ao obter usuário por email")
		return
	}
	
	// Responder com o usuário
	respondWithJSON(w, http.StatusOK, user)
}

// CreateUser cria um novo usuário
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "users.create",
		trace.WithAttributes(
			attribute.String("handler", "user_handler"),
			attribute.String("method", "CreateUser"),
		),
	)
	defer span.End()
	
	reqID := getRequestID(r)
	
	// Extrair tenant ID do contexto
	tenantID, err := h.getTenantIDFromContext(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_tenant", "Tenant ID inválido", reqID)
		return
	}
	
	// Decodificar corpo da requisição
	var req application.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_request", "Requisição inválida", reqID)
		return
	}
	
	// Definir o tenant ID a partir do contexto
	req.TenantID = tenantID
	
	// Validar a requisição
	if err := h.validateRequest(req); err != nil {
		log.Error().Err(err).Msg("Erro de validação da requisição")
		respondWithError(w, http.StatusBadRequest, "validation_error", "Dados inválidos: "+err.Error(), reqID)
		return
	}
	
	// Chamar o serviço para criar o usuário
	user, err := h.userService.CreateUser(ctx, req)
	if err != nil {
		// Verificar o tipo de erro para enviar a resposta apropriada
		var appErr application.AppError
		if errors.As(err, &appErr) {
			respondWithError(w, appErr.StatusCode, appErr.Code, appErr.Message, reqID)
		} else {
			respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno no servidor", reqID)
		}
		log.Error().Err(err).Str("request_id", reqID).Msg("Erro ao criar usuário")
		return
	}
	
	// Responder com o usuário criado
	respondWithJSON(w, http.StatusCreated, user)
}

// UpdateUser atualiza um usuário existente
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "users.update",
		trace.WithAttributes(
			attribute.String("handler", "user_handler"),
			attribute.String("method", "UpdateUser"),
		),
	)
	defer span.End()
	
	reqID := getRequestID(r)
	
	// Extrair tenant ID do contexto
	tenantID, err := h.getTenantIDFromContext(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_tenant", "Tenant ID inválido", reqID)
		return
	}
	
	// Extrair user ID da URL
	userID, err := h.getQueryUUID(r, "userId")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_user_id", "ID de usuário inválido", reqID)
		return
	}
	
	// Decodificar corpo da requisição
	var req application.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_request", "Requisição inválida", reqID)
		return
	}
	
	// Definir o ID do usuário a partir da URL
	req.ID = userID
	
	// Validar a requisição
	if err := h.validateRequest(req); err != nil {
		respondWithError(w, http.StatusBadRequest, "validation_error", "Dados inválidos: "+err.Error(), reqID)
		return
	}
	
	// Chamar o serviço para atualizar o usuário
	user, err := h.userService.UpdateUser(ctx, req)
	if err != nil {
		// Verificar o tipo de erro para enviar a resposta apropriada
		var appErr application.AppError
		if errors.As(err, &appErr) {
			respondWithError(w, appErr.StatusCode, appErr.Code, appErr.Message, reqID)
		} else {
			respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno no servidor", reqID)
		}
		log.Error().Err(err).Str("request_id", reqID).Msg("Erro ao atualizar usuário")
		return
	}
	
	// Responder com o usuário atualizado
	respondWithJSON(w, http.StatusOK, user)
}

// DeleteUser exclui um usuário (soft delete por padrão)
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "users.delete",
		trace.WithAttributes(
			attribute.String("handler", "user_handler"),
			attribute.String("method", "DeleteUser"),
		),
	)
	defer span.End()
	
	reqID := getRequestID(r)
	
	// Extrair tenant ID do contexto
	tenantID, err := h.getTenantIDFromContext(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_tenant", "Tenant ID inválido", reqID)
		return
	}
	
	// Extrair user ID da URL
	userID, err := h.getQueryUUID(r, "userId")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_user_id", "ID de usuário inválido", reqID)
		return
	}
	
	// Verificar se é hard delete (remoção permanente)
	hardDelete := r.URL.Query().Get("hard") == "true"
	
	// Chamar o serviço para excluir o usuário
	if err := h.userService.DeleteUser(ctx, tenantID, userID, hardDelete); err != nil {
		// Verificar o tipo de erro para enviar a resposta apropriada
		var appErr application.AppError
		if errors.As(err, &appErr) {
			respondWithError(w, appErr.StatusCode, appErr.Code, appErr.Message, reqID)
		} else {
			respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno no servidor", reqID)
		}
		log.Error().Err(err).Str("request_id", reqID).Msg("Erro ao excluir usuário")
		return
	}
	
	// Responder com sucesso
	respondWithJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// Funções para manipular endereços, contatos e funções (roles) serão implementadas em arquivos posteriores

func (h *UserHandler) ListUserAddresses(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", getRequestID(r))
}

func (h *UserHandler) AddUserAddress(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", getRequestID(r))
}

func (h *UserHandler) UpdateUserAddress(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", getRequestID(r))
}

func (h *UserHandler) DeleteUserAddress(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", getRequestID(r))
}

func (h *UserHandler) ListUserContacts(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", getRequestID(r))
}

func (h *UserHandler) AddUserContact(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", getRequestID(r))
}

func (h *UserHandler) UpdateUserContact(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", getRequestID(r))
}

func (h *UserHandler) DeleteUserContact(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", getRequestID(r))
}

func (h *UserHandler) ListUserRoles(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", getRequestID(r))
}

func (h *UserHandler) AssignRolesToUser(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", getRequestID(r))
}

func (h *UserHandler) RevokeRolesFromUser(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", getRequestID(r))
}

// getRequestID extrai o ID da requisição do cabeçalho
func getRequestID(r *http.Request) string {
	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = "generated-id"
	}
	return reqID
}

// respondWithJSON envia uma resposta JSON com o status HTTP especificado
func respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("Erro ao serializar resposta JSON")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal_server_error","message":"Erro ao processar resposta"}`))
	}
}

// respondWithError envia uma resposta de erro padronizada
func respondWithError(w http.ResponseWriter, status int, errorCode, message string, reqID string) {
	errRes := struct {
		Error     string `json:"error"`
		Code      string `json:"code,omitempty"`
		Message   string `json:"message"`
		RequestID string `json:"request_id,omitempty"`
	}{
		Error:     errorCode,
		Code:      errorCode,
		Message:   message,
		RequestID: reqID,
	}
	
	respondWithJSON(w, status, errRes)
}