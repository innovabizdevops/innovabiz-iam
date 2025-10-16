/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Handler HTTP para rotas de autenticação.
 * Implementa os endpoints REST para autenticação e autorização.
 */

package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/innovabiz/iam/services/identity-service/internal/application"
	"github.com/innovabiz/iam/services/identity-service/internal/domain/model"
)

// AuthHandler gerencia os endpoints HTTP relacionados à autenticação
type AuthHandler struct {
	authService application.AuthService
	tracer      trace.Tracer
}

// NewAuthHandler cria uma nova instância do AuthHandler
func NewAuthHandler(authService application.AuthService, tracer trace.Tracer) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		tracer:      tracer,
	}
}

// RegisterRoutes registra as rotas HTTP para o AuthHandler
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	// Configura CORS para as rotas de autenticação
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Em produção, isto seria limitado aos domínios permitidos
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Tenant-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Tempo máximo de cache para preflight requests
	}))

	// Rotas de autenticação
	r.Route("/auth", func(r chi.Router) {
		// Login e registro
		r.Post("/login", h.Login)
		r.Post("/verify-mfa", h.VerifyMFA)
		r.Post("/refresh-token", h.RefreshToken)
		r.Post("/logout", h.Logout)
		r.Post("/register", h.Register)
		
		// Gerenciamento de senhas
		r.Post("/password/reset-request", h.RequestPasswordReset)
		r.Post("/password/reset-complete", h.CompletePasswordReset)
		r.Post("/password/change", h.ChangePassword)
		
		// MFA
		r.Post("/mfa/setup", h.SetupMFA)
		r.Post("/mfa/generate-backup-codes", h.GenerateMFABackupCodes)
		r.Post("/mfa/verify-setup", h.VerifyMFASetup)
		r.Post("/mfa/disable", h.DisableMFA)
		
		// Verificação de token e informações do usuário
		r.Get("/userinfo", h.GetUserInfo)
		r.Post("/verify-token", h.VerifyToken)
		r.Get("/sessions", h.ListSessions)
		r.Delete("/sessions/{sessionId}", h.RevokeSession)
		r.Delete("/sessions", h.RevokeAllSessions)
	})
}

// resError é uma estrutura para resposta de erro padronizada
type resError struct {
	Error     string `json:"error"`
	Code      string `json:"code,omitempty"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
	Timestamp string `json:"timestamp"`
}

// respondWithJSON envia uma resposta JSON com o status HTTP especificado
func (h *AuthHandler) respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("Erro ao serializar resposta JSON")
		// Em caso de erro na serialização, tenta enviar uma resposta de erro simples
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal_server_error","message":"Erro ao processar resposta"}`))
	}
}

// respondWithError envia uma resposta de erro padronizada
func (h *AuthHandler) respondWithError(w http.ResponseWriter, status int, errorCode, message string, reqID string) {
	errRes := resError{
		Error:     errorCode,
		Code:      errorCode,
		Message:   message,
		RequestID: reqID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	
	h.respondWithJSON(w, status, errRes)
}

// getRequestID extrai o ID da requisição do contexto ou gera um novo
func (h *AuthHandler) getRequestID(r *http.Request) string {
	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		// Em uma implementação real, geraria um UUID aqui
		reqID = "generated-id"
	}
	return reqID
}

// getTenantID extrai o ID do tenant do cabeçalho ou do token
func (h *AuthHandler) getTenantID(r *http.Request) string {
	tenantID := r.Header.Get("X-Tenant-ID")
	
	// Se não estiver no cabeçalho, tenta extrair do token de autorização
	if tenantID == "" {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			// Em uma implementação real, extrairia o tenant ID do token JWT
		}
	}
	
	return tenantID
}

// Login processa uma requisição de login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "auth.login",
		trace.WithAttributes(
			attribute.String("handler", "auth_handler"),
			attribute.String("method", "Login"),
		),
	)
	defer span.End()
	
	reqID := h.getRequestID(r)
	
	// Decodifica o corpo da requisição
	var req application.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_request", "Requisição inválida", reqID)
		return
	}
	
	// Adiciona informações do cliente à requisição
	req.IPAddress = r.RemoteAddr
	req.UserAgent = r.UserAgent()
	
	// Chama o serviço de autenticação
	resp, err := h.authService.Login(ctx, req)
	if err != nil {
		// Verifica o tipo de erro para enviar a resposta apropriada
		var appErr application.AppError
		if errors.As(err, &appErr) {
			h.respondWithError(w, appErr.StatusCode, appErr.Code, appErr.Message, reqID)
		} else {
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno no servidor", reqID)
		}
		log.Error().Err(err).Str("request_id", reqID).Msg("Erro no login")
		return
	}
	
	// Responde com os tokens de autenticação
	h.respondWithJSON(w, http.StatusOK, resp)
}

// VerifyMFA verifica um código MFA após login inicial
func (h *AuthHandler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "auth.verify_mfa")
	defer span.End()
	
	reqID := h.getRequestID(r)
	
	// Decodifica o corpo da requisição
	var req application.MFAVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_request", "Requisição inválida", reqID)
		return
	}
	
	// Chama o serviço de autenticação
	resp, err := h.authService.VerifyMFA(ctx, req)
	if err != nil {
		// Verifica o tipo de erro para enviar a resposta apropriada
		var appErr application.AppError
		if errors.As(err, &appErr) {
			h.respondWithError(w, appErr.StatusCode, appErr.Code, appErr.Message, reqID)
		} else {
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno no servidor", reqID)
		}
		log.Error().Err(err).Str("request_id", reqID).Msg("Erro na verificação MFA")
		return
	}
	
	// Responde com os tokens de autenticação
	h.respondWithJSON(w, http.StatusOK, resp)
}

// RefreshToken renova tokens usando um refresh token
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "auth.refresh_token")
	defer span.End()
	
	reqID := h.getRequestID(r)
	
	// Decodifica o corpo da requisição
	var req application.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_request", "Requisição inválida", reqID)
		return
	}
	
	// Adiciona informações do cliente à requisição
	req.IPAddress = r.RemoteAddr
	req.UserAgent = r.UserAgent()
	
	// Chama o serviço de autenticação
	resp, err := h.authService.RefreshToken(ctx, req)
	if err != nil {
		// Verifica o tipo de erro para enviar a resposta apropriada
		var appErr application.AppError
		if errors.As(err, &appErr) {
			h.respondWithError(w, appErr.StatusCode, appErr.Code, appErr.Message, reqID)
		} else {
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno no servidor", reqID)
		}
		log.Error().Err(err).Str("request_id", reqID).Msg("Erro na renovação de token")
		return
	}
	
	// Responde com os novos tokens
	h.respondWithJSON(w, http.StatusOK, resp)
}

// Logout encerra a sessão atual ou todas as sessões do usuário
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "auth.logout")
	defer span.End()
	
	reqID := h.getRequestID(r)
	
	// Decodifica o corpo da requisição
	var req application.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_request", "Requisição inválida", reqID)
		return
	}
	
	// Extrai o token se não foi fornecido no corpo
	if req.AccessToken == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			req.AccessToken = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}
	
	// Chama o serviço de autenticação
	if err := h.authService.Logout(ctx, req); err != nil {
		// Verifica o tipo de erro para enviar a resposta apropriada
		var appErr application.AppError
		if errors.As(err, &appErr) {
			h.respondWithError(w, appErr.StatusCode, appErr.Code, appErr.Message, reqID)
		} else {
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno no servidor", reqID)
		}
		log.Error().Err(err).Str("request_id", reqID).Msg("Erro no logout")
		return
	}
	
	// Responde com sucesso
	h.respondWithJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// GetUserInfo obtém informações do usuário a partir do token
func (h *AuthHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "auth.get_user_info")
	defer span.End()
	
	reqID := h.getRequestID(r)
	
	// Extrai o token de autorização
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		h.respondWithError(w, http.StatusUnauthorized, "unauthorized", "Token de autorização ausente ou inválido", reqID)
		return
	}
	
	token := strings.TrimPrefix(authHeader, "Bearer ")
	
	// Chama o serviço de autenticação
	userInfo, err := h.authService.GetUserInfo(ctx, token)
	if err != nil {
		// Verifica o tipo de erro para enviar a resposta apropriada
		var appErr application.AppError
		if errors.As(err, &appErr) {
			h.respondWithError(w, appErr.StatusCode, appErr.Code, appErr.Message, reqID)
		} else {
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno no servidor", reqID)
		}
		log.Error().Err(err).Str("request_id", reqID).Msg("Erro ao obter informações do usuário")
		return
	}
	
	// Responde com as informações do usuário
	h.respondWithJSON(w, http.StatusOK, userInfo)
}

// VerifyToken verifica a validade de um token de acesso
func (h *AuthHandler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "auth.verify_token")
	defer span.End()
	
	reqID := h.getRequestID(r)
	
	// Decodifica o corpo da requisição
	var req application.VerifyTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_request", "Requisição inválida", reqID)
		return
	}
	
	// Extrai o token se não foi fornecido no corpo
	if req.Token == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			req.Token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}
	
	// Chama o serviço de autenticação
	resp, err := h.authService.VerifyToken(ctx, req)
	if err != nil {
		// Verifica o tipo de erro para enviar a resposta apropriada
		var appErr application.AppError
		if errors.As(err, &appErr) {
			h.respondWithError(w, appErr.StatusCode, appErr.Code, appErr.Message, reqID)
		} else {
			h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Erro interno no servidor", reqID)
		}
		log.Error().Err(err).Str("request_id", reqID).Msg("Erro na verificação do token")
		return
	}
	
	// Responde com o resultado da verificação
	h.respondWithJSON(w, http.StatusOK, resp)
}

// Register registra um novo usuário
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    // Implementação será adicionada em outro arquivo
    h.respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", h.getRequestID(r))
}

// RequestPasswordReset solicita a redefinição de senha
func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
    // Implementação será adicionada em outro arquivo
    h.respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", h.getRequestID(r))
}

// CompletePasswordReset completa o processo de redefinição de senha
func (h *AuthHandler) CompletePasswordReset(w http.ResponseWriter, r *http.Request) {
    // Implementação será adicionada em outro arquivo
    h.respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", h.getRequestID(r))
}

// ChangePassword altera a senha do usuário
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
    // Implementação será adicionada em outro arquivo
    h.respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", h.getRequestID(r))
}

// SetupMFA configura a autenticação multi-fator para um usuário
func (h *AuthHandler) SetupMFA(w http.ResponseWriter, r *http.Request) {
    // Implementação será adicionada em outro arquivo
    h.respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", h.getRequestID(r))
}

// VerifyMFASetup verifica a configuração de MFA
func (h *AuthHandler) VerifyMFASetup(w http.ResponseWriter, r *http.Request) {
    // Implementação será adicionada em outro arquivo
    h.respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", h.getRequestID(r))
}

// DisableMFA desativa a autenticação multi-fator
func (h *AuthHandler) DisableMFA(w http.ResponseWriter, r *http.Request) {
    // Implementação será adicionada em outro arquivo
    h.respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", h.getRequestID(r))
}

// GenerateMFABackupCodes gera novos códigos de backup para MFA
func (h *AuthHandler) GenerateMFABackupCodes(w http.ResponseWriter, r *http.Request) {
    // Implementação será adicionada em outro arquivo
    h.respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", h.getRequestID(r))
}

// ListSessions lista as sessões ativas do usuário
func (h *AuthHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
    // Implementação será adicionada em outro arquivo
    h.respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", h.getRequestID(r))
}

// RevokeSession revoga uma sessão específica
func (h *AuthHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
    // Implementação será adicionada em outro arquivo
    h.respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", h.getRequestID(r))
}

// RevokeAllSessions revoga todas as sessões do usuário
func (h *AuthHandler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
    // Implementação será adicionada em outro arquivo
    h.respondWithError(w, http.StatusNotImplemented, "not_implemented", "Função não implementada", h.getRequestID(r))
}