/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Modelo de domínio para Usuário no sistema IAM.
 * Implementa conceitos de Domain-Driven Design e segue os princípios SOLID.
 */

package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

// Erros específicos do domínio de usuário
var (
	ErrInvalidEmail         = errors.New("endereço de email inválido")
	ErrWeakPassword         = errors.New("senha não atende aos requisitos mínimos de segurança")
	ErrInvalidUserName      = errors.New("nome de usuário inválido")
	ErrRequiredField        = errors.New("campo obrigatório não fornecido")
	ErrUserAccountLocked    = errors.New("conta de usuário bloqueada")
	ErrUserAccountSuspended = errors.New("conta de usuário suspensa")
	ErrUserAccountDisabled  = errors.New("conta de usuário desativada")
	ErrInvalidUserStatus    = errors.New("status de usuário inválido")
	ErrInvalidTenantID      = errors.New("ID do tenant inválido")
)

// UserStatus define os possíveis estados de uma conta de usuário
type UserStatus string

// Constantes para o status de usuário
const (
	UserStatusActive    UserStatus = "active"
	UserStatusLocked    UserStatus = "locked"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusDisabled  UserStatus = "disabled"
	UserStatusPending   UserStatus = "pending"
)

// AuthProvider representa um provedor de autenticação externo
type AuthProvider string

// Constantes para os provedores de autenticação
const (
	AuthProviderLocal      AuthProvider = "local"
	AuthProviderGoogle     AuthProvider = "google"
	AuthProviderMicrosoft  AuthProvider = "microsoft"
	AuthProviderFacebook   AuthProvider = "facebook"
	AuthProviderApple      AuthProvider = "apple"
	AuthProviderTwitter    AuthProvider = "twitter"
	AuthProviderGitHub     AuthProvider = "github"
	AuthProviderLinkedIn   AuthProvider = "linkedin"
	AuthProviderOktaSAML   AuthProvider = "okta_saml"
	AuthProviderAzureAD    AuthProvider = "azure_ad"
	AuthProviderCustomOIDC AuthProvider = "custom_oidc"
	AuthProviderCustomSAML AuthProvider = "custom_saml"
)

// MFAMethod representa um método de autenticação multi-fator
type MFAMethod string

// Constantes para os métodos de MFA
const (
	MFAMethodNone       MFAMethod = "none"
	MFAMethodTOTP       MFAMethod = "totp"
	MFAMethodSMS        MFAMethod = "sms"
	MFAMethodEmail      MFAMethod = "email"
	MFAMethodFIDO2      MFAMethod = "fido2"
	MFAMethodWebAuthn   MFAMethod = "webauthn"
	MFAMethodPush       MFAMethod = "push_notification"
	MFAMethodRecoveryCodes MFAMethod = "recovery_codes"
)

// UserCredential representa as credenciais de autenticação de um usuário
type UserCredential struct {
	ID                 uuid.UUID     `json:"id"`
	UserID             uuid.UUID     `json:"user_id"`
	PasswordHash       []byte        `json:"-"`                     // Nunca exportar via JSON
	PasswordLastChange time.Time     `json:"password_last_change"`
	PasswordTempExpiry *time.Time    `json:"password_temp_expiry"` // Nulo se não for temporária
	Provider           AuthProvider  `json:"provider"`
	ProviderUserID     string        `json:"provider_user_id"`
	FailedAttempts     int           `json:"failed_attempts"`
	LastFailedAttempt  *time.Time    `json:"last_failed_attempt"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
}

// MFASettings representa as configurações de MFA de um usuário
type MFASettings struct {
	Enabled      bool        `json:"enabled"`
	DefaultMethod MFAMethod   `json:"default_method"`
	Methods      []MFAMethod  `json:"methods"`
	TOTPSecret   []byte       `json:"-"` // Nunca exportar via JSON
	PhoneNumber  string       `json:"phone_number"`
	RecoveryCodes []string    `json:"-"` // Nunca exportar via JSON
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// Address representa um endereço físico associado a um usuário
type Address struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Type        string     `json:"type"`       // residencial, comercial, etc.
	Street      string     `json:"street"`
	Number      string     `json:"number"`
	Complement  string     `json:"complement"`
	District    string     `json:"district"`
	City        string     `json:"city"`
	State       string     `json:"state"`
	Country     string     `json:"country"`
	PostalCode  string     `json:"postal_code"`
	IsDefault   bool       `json:"is_default"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Contact representa informações de contato do usuário
type Contact struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	Type       string     `json:"type"`      // email, telefone, etc.
	Value      string     `json:"value"`
	Verified   bool       `json:"verified"`
	IsDefault  bool       `json:"is_default"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// UserSession representa uma sessão ativa de um usuário
type UserSession struct {
	ID               uuid.UUID    `json:"id"`
	UserID           uuid.UUID    `json:"user_id"`
	TenantID         uuid.UUID    `json:"tenant_id"`
	Token            string       `json:"-"`           // Nunca exportar via JSON
	RefreshToken     string       `json:"-"`           // Nunca exportar via JSON
	ExpiresAt        time.Time    `json:"expires_at"`
	RefreshExpiresAt time.Time    `json:"refresh_expires_at"`
	IPAddress        string       `json:"ip_address"`
	UserAgent        string       `json:"user_agent"`
	DeviceInfo       string       `json:"device_info"`
	Location         string       `json:"location"`
	LastActivity     time.Time    `json:"last_activity"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}

// User representa a entidade principal do modelo de domínio para usuários no IAM
type User struct {
	ID                uuid.UUID      `json:"id"`
	TenantID          uuid.UUID      `json:"tenant_id"`
	Username          string         `json:"username"`
	Email             string         `json:"email"`
	EmailVerified     bool           `json:"email_verified"`
	FirstName         string         `json:"first_name"`
	LastName          string         `json:"last_name"`
	DisplayName       string         `json:"display_name"`
	PhoneNumber       string         `json:"phone_number"`
	PhoneVerified     bool           `json:"phone_verified"`
	ProfilePictureURL string         `json:"profile_picture_url"`
	Locale            string         `json:"locale"`
	Timezone          string         `json:"timezone"`
	Metadata          map[string]interface{} `json:"metadata"`
	Status            UserStatus     `json:"status"`
	LoginCount        int            `json:"login_count"`
	LastLoginAt       *time.Time     `json:"last_login_at"`
	LastTokenIssuedAt *time.Time     `json:"last_token_issued_at"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         *time.Time     `json:"deleted_at"`
	
	// Campos de associação com outros objetos de domínio
	Credentials *UserCredential  `json:"-"` // Não exportado no JSON
	MFA         *MFASettings     `json:"mfa,omitempty"`
	Addresses   []*Address       `json:"addresses,omitempty"`
	Contacts    []*Contact       `json:"contacts,omitempty"`
	Sessions    []*UserSession   `json:"-"` // Não exportado no JSON
}

// NewUser cria uma nova instância de usuário com validações
func NewUser(
	tenantID uuid.UUID,
	username, email, firstName, lastName string,
) (*User, error) {
	// Validações
	if tenantID == uuid.Nil {
		return nil, ErrInvalidTenantID
	}
	
	if username == "" {
		return nil, ErrInvalidUserName
	}
	
	if email == "" {
		return nil, ErrInvalidEmail
	}
	
	// Em um cenário real, teríamos validação mais robusta de email
	
	now := time.Now().UTC()
	
	// Criação do usuário
	user := &User{
		ID:            uuid.New(),
		TenantID:      tenantID,
		Username:      username,
		Email:         email,
		FirstName:     firstName,
		LastName:      lastName,
		DisplayName:   firstName + " " + lastName,
		Status:        UserStatusPending, // Usuários começam com status pendente
		Metadata:      make(map[string]interface{}),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	
	return user, nil
}

// SetCredentials define as credenciais de autenticação do usuário
func (u *User) SetCredentials(password string, provider AuthProvider) error {
	if provider == AuthProviderLocal && password == "" {
		return ErrRequiredField
	}
	
	// Em um cenário real, validaríamos a força da senha
	
	// Hash da senha usando Argon2id (algoritmo recomendado)
	salt := make([]byte, 16) // Em um cenário real, geraríamos um salt aleatório
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	
	u.Credentials = &UserCredential{
		ID:                 uuid.New(),
		UserID:             u.ID,
		PasswordHash:       hash,
		Provider:           provider,
		PasswordLastChange: time.Now().UTC(),
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}
	
	return nil
}

// SetMFASettings define as configurações de MFA para o usuário
func (u *User) SetMFASettings(enabled bool, defaultMethod MFAMethod) {
	now := time.Now().UTC()
	
	if u.MFA == nil {
		u.MFA = &MFASettings{
			Enabled:      enabled,
			DefaultMethod: defaultMethod,
			Methods:      []MFAMethod{defaultMethod},
			CreatedAt:    now,
			UpdatedAt:    now,
		}
	} else {
		u.MFA.Enabled = enabled
		u.MFA.DefaultMethod = defaultMethod
		u.MFA.Methods = []MFAMethod{defaultMethod}
		u.MFA.UpdatedAt = now
	}
}

// Activate muda o status do usuário para ativo
func (u *User) Activate() error {
	if u.Status == UserStatusDisabled {
		return ErrUserAccountDisabled
	}
	
	u.Status = UserStatusActive
	u.UpdatedAt = time.Now().UTC()
	return nil
}

// Lock bloqueia a conta do usuário
func (u *User) Lock() {
	u.Status = UserStatusLocked
	u.UpdatedAt = time.Now().UTC()
}

// Unlock desbloqueia a conta do usuário
func (u *User) Unlock() error {
	if u.Status == UserStatusDisabled {
		return ErrUserAccountDisabled
	}
	
	u.Status = UserStatusActive
	u.UpdatedAt = time.Now().UTC()
	return nil
}

// Disable desativa permanentemente a conta do usuário
func (u *User) Disable() {
	u.Status = UserStatusDisabled
	u.UpdatedAt = time.Now().UTC()
	now := time.Now().UTC()
	u.DeletedAt = &now
}

// IsActive verifica se o usuário está ativo
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// CanAuthenticate verifica se o usuário pode se autenticar
func (u *User) CanAuthenticate() bool {
	return u.Status == UserStatusActive || u.Status == UserStatusPending
}

// FullName retorna o nome completo do usuário
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// RequiresMFA verifica se o usuário deve usar MFA para login
func (u *User) RequiresMFA() bool {
	return u.MFA != nil && u.MFA.Enabled
}