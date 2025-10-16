/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Serviço de aplicação para autenticação e autorização.
 * Implementa casos de uso relacionados à autenticação de usuários,
 * seguindo os princípios da Clean Architecture e Domain-Driven Design.
 */

package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	
	"github.com/innovabiz/iam/services/identity-service/internal/domain/model"
)

// Erros específicos do serviço de autenticação
var (
	ErrInvalidCredentials     = NewAppError("credenciais inválidas", "AUTH_001", 401)
	ErrAccountLocked          = NewAppError("conta bloqueada", "AUTH_002", 403)
	ErrAccountDisabled        = NewAppError("conta desativada", "AUTH_003", 403)
	ErrPasswordExpired        = NewAppError("senha expirada", "AUTH_004", 403)
	ErrMFARequired            = NewAppError("autenticação multi-fator requerida", "AUTH_005", 403)
	ErrInvalidMFACode         = NewAppError("código MFA inválido", "AUTH_006", 401)
	ErrInvalidToken           = NewAppError("token inválido ou expirado", "AUTH_007", 401)
	ErrInsufficientPermission = NewAppError("permissão insuficiente", "AUTH_008", 403)
	ErrSessionExpired         = NewAppError("sessão expirada", "AUTH_009", 401)
	ErrSessionRevoked         = NewAppError("sessão revogada", "AUTH_010", 401)
)

// LoginRequest representa uma solicitação de login
type LoginRequest struct {
	TenantID      uuid.UUID        `json:"tenant_id"`
	Username      string           `json:"username,omitempty"` // Email ou username
	Password      string           `json:"password,omitempty"`
	Provider      model.AuthProvider `json:"provider"`
	ProviderToken string           `json:"provider_token,omitempty"` // Token de provedor externo
	IPAddress     string           `json:"ip_address"`
	UserAgent     string           `json:"user_agent"`
	DeviceInfo    string           `json:"device_info,omitempty"`
	Location      string           `json:"location,omitempty"`
}

// MFAVerifyRequest representa uma solicitação de verificação MFA
type MFAVerifyRequest struct {
	MFAToken  string         `json:"mfa_token"`  // Token temporário recebido após login
	Method    model.MFAMethod `json:"method"`
	Code      string         `json:"code"`       // Código OTP ou outro tipo de desafio
}

// TokenResponse representa os tokens retornados após autenticação
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`    // Tempo em segundos
	RefreshToken string    `json:"refresh_token"`
	Scope        string    `json:"scope"`
	MFARequired  bool      `json:"mfa_required"`
	MFAToken     string    `json:"mfa_token,omitempty"` // Token temporário para completar MFA
	IssuedAt     time.Time `json:"issued_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// UserInfoResponse contém informações do usuário após autenticação
type UserInfoResponse struct {
	ID            string                 `json:"id"`
	TenantID      string                 `json:"tenant_id"`
	Username      string                 `json:"username"`
	Email         string                 `json:"email"`
	FirstName     string                 `json:"first_name"`
	LastName      string                 `json:"last_name"`
	DisplayName   string                 `json:"display_name"`
	EmailVerified bool                   `json:"email_verified"`
	PhoneVerified bool                   `json:"phone_verified"`
	Roles         []string               `json:"roles"`
	Permissions   []string               `json:"permissions"`
	Metadata      map[string]interface{} `json:"metadata"`
	Provider      model.AuthProvider     `json:"provider"`
}

// RefreshTokenRequest representa uma solicitação de renovação de token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
	IPAddress    string `json:"ip_address"`
	UserAgent    string `json:"user_agent"`
}

// LogoutRequest representa uma solicitação de logout
type LogoutRequest struct {
	AccessToken string `json:"access_token"`
	SessionID   string `json:"session_id,omitempty"`
	All         bool   `json:"all"` // Se deve encerrar todas as sessões
}

// VerifyTokenRequest representa uma solicitação para verificar a validade de um token
type VerifyTokenRequest struct {
	Token     string   `json:"token"`
	Resource  string   `json:"resource,omitempty"`  // Recurso que está sendo acessado
	Action    string   `json:"action,omitempty"`    // Ação que está sendo realizada
	Scopes    []string `json:"scopes,omitempty"`    // Escopos necessários
}

// VerifyTokenResponse representa a resposta da verificação de token
type VerifyTokenResponse struct {
	Valid       bool                   `json:"valid"`
	UserID      string                 `json:"user_id,omitempty"`
	TenantID    string                 `json:"tenant_id,omitempty"`
	Permissions []string               `json:"permissions,omitempty"`
	Roles       []string               `json:"roles,omitempty"`
	Scopes      []string               `json:"scopes,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ExpiresAt   time.Time              `json:"expires_at,omitempty"`
	IssuedAt    time.Time              `json:"issued_at,omitempty"`
}

// ChangePasswordRequest representa uma solicitação de alteração de senha
type ChangePasswordRequest struct {
	UserID         uuid.UUID `json:"user_id"`
	TenantID       uuid.UUID `json:"tenant_id"`
	CurrentPassword string    `json:"current_password"`
	NewPassword    string    `json:"new_password"`
}

// ResetPasswordRequest representa uma solicitação de redefinição de senha
type ResetPasswordRequest struct {
	Email    string    `json:"email"`
	TenantID uuid.UUID `json:"tenant_id"`
}

// CompletePasswordResetRequest representa uma solicitação para completar a redefinição de senha
type CompletePasswordResetRequest struct {
	Token       string    `json:"token"`
	NewPassword string    `json:"new_password"`
}

// AuthService define a interface para o serviço de autenticação
type AuthService interface {
	// Login autentica um usuário e retorna tokens de acesso
	Login(ctx context.Context, req LoginRequest) (*TokenResponse, error)
	
	// VerifyMFA verifica a autenticação multi-fator após o login inicial
	VerifyMFA(ctx context.Context, req MFAVerifyRequest) (*TokenResponse, error)
	
	// RefreshToken renova tokens de acesso usando um refresh token
	RefreshToken(ctx context.Context, req RefreshTokenRequest) (*TokenResponse, error)
	
	// Logout encerra uma ou todas as sessões de um usuário
	Logout(ctx context.Context, req LogoutRequest) error
	
	// VerifyToken verifica a validade de um token e retorna informações do usuário
	VerifyToken(ctx context.Context, req VerifyTokenRequest) (*VerifyTokenResponse, error)
	
	// GetUserInfo obtém informações do usuário a partir de um token válido
	GetUserInfo(ctx context.Context, token string) (*UserInfoResponse, error)
	
	// ChangePassword altera a senha de um usuário
	ChangePassword(ctx context.Context, req ChangePasswordRequest) error
	
	// RequestPasswordReset solicita uma redefinição de senha
	RequestPasswordReset(ctx context.Context, req ResetPasswordRequest) error
	
	// CompletePasswordReset completa o processo de redefinição de senha
	CompletePasswordReset(ctx context.Context, req CompletePasswordResetRequest) error
	
	// UpdateMFASettings atualiza as configurações de MFA de um usuário
	UpdateMFASettings(ctx context.Context, userID uuid.UUID, enabled bool, method model.MFAMethod) error
	
	// GenerateMFABackupCodes gera novos códigos de backup para MFA
	GenerateMFABackupCodes(ctx context.Context, userID uuid.UUID) ([]string, error)
}

// AppError representa um erro de aplicação com código e status HTTP
type AppError struct {
	Message    string `json:"message"`
	Code       string `json:"code"`
	StatusCode int    `json:"status_code"`
}

// Error implementa a interface error
func (e AppError) Error() string {
	return e.Message
}

// NewAppError cria um novo erro de aplicação
func NewAppError(message, code string, statusCode int) AppError {
	return AppError{
		Message:    message,
		Code:       code,
		StatusCode: statusCode,
	}
}