// Modelos de usuário para o serviço de identidade - INNOVABIZ Platform
// Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
// PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III, PSD2, AML/KYC
package models

import (
	"time"

	"github.com/google/uuid"
)

// Status possíveis para um usuário
const (
	UserStatusActive   = "ACTIVE"
	UserStatusInactive = "INACTIVE"
	UserStatusLocked   = "LOCKED"
	UserStatusPending  = "PENDING"
	UserStatusSuspended = "SUSPENDED"
)

// Tipos de usuário
const (
	UserTypeIndividual = "INDIVIDUAL"
	UserTypeCorporate  = "CORPORATE"
	UserTypeSystem     = "SYSTEM"
	UserTypePartner    = "PARTNER"
	UserTypeIntegrator = "INTEGRATOR"
)

// Origem do usuário
const (
	UserSourceDirect    = "DIRECT"
	UserSourceImported  = "IMPORTED"
	UserSourceMigrated  = "MIGRATED"
	UserSourceFederated = "FEDERATED"
)

// Níveis de acesso
const (
	AccessLevelLow     = "LOW"
	AccessLevelMedium  = "MEDIUM"
	AccessLevelHigh    = "HIGH"
	AccessLevelCritical = "CRITICAL"
)

// User representa um usuário no sistema de IAM
type User struct {
	ID               string           `json:"id" bson:"_id"`
	TenantID         string           `json:"tenant_id" bson:"tenant_id"`
	Username         string           `json:"username" bson:"username"`
	Email            string           `json:"email" bson:"email"`
	EmailVerified    bool             `json:"email_verified" bson:"email_verified"`
	Phone            string           `json:"phone,omitempty" bson:"phone,omitempty"`
	PhoneVerified    bool             `json:"phone_verified" bson:"phone_verified"`
	FirstName        string           `json:"first_name,omitempty" bson:"first_name,omitempty"`
	LastName         string           `json:"last_name,omitempty" bson:"last_name,omitempty"`
	DisplayName      string           `json:"display_name,omitempty" bson:"display_name,omitempty"`
	Status           string           `json:"status" bson:"status"`
	UserType         string           `json:"user_type" bson:"user_type"`
	Source           string           `json:"source" bson:"source"`
	AccessLevel      string           `json:"access_level" bson:"access_level"`
	Attributes       map[string]interface{} `json:"attributes,omitempty" bson:"attributes,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
	PasswordLastChanged *time.Time     `json:"password_last_changed,omitempty" bson:"password_last_changed,omitempty"`
	MFAEnabled       bool             `json:"mfa_enabled" bson:"mfa_enabled"`
	MFAMethods       []MFAMethod      `json:"mfa_methods,omitempty" bson:"mfa_methods,omitempty"`
	LastLoginAt      *time.Time       `json:"last_login_at,omitempty" bson:"last_login_at,omitempty"`
	LastLoginIP      string           `json:"last_login_ip,omitempty" bson:"last_login_ip,omitempty"`
	FailedLoginAttempts int            `json:"failed_login_attempts" bson:"failed_login_attempts"`
	LastFailedLoginAt *time.Time       `json:"last_failed_login_at,omitempty" bson:"last_failed_login_at,omitempty"`
	LockExpiresAt    *time.Time       `json:"lock_expires_at,omitempty" bson:"lock_expires_at,omitempty"`
	CreatedAt        time.Time        `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at" bson:"updated_at"`
	CreatedBy        string           `json:"created_by" bson:"created_by"`
	UpdatedBy        string           `json:"updated_by" bson:"updated_by"`
	Deleted          bool             `json:"deleted" bson:"deleted"`
	DeletedAt        *time.Time       `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	DeletedBy        string           `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
}

// MFAMethod representa um método de autenticação multi-fator
type MFAMethod struct {
	Type        string    `json:"type" bson:"type"` // SMS, EMAIL, TOTP, FIDO2, etc.
	Identifier  string    `json:"identifier" bson:"identifier"` // número de telefone, email, etc.
	Enabled     bool      `json:"enabled" bson:"enabled"`
	VerifiedAt  time.Time `json:"verified_at" bson:"verified_at"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty" bson:"last_used_at,omitempty"`
}

// UserRole representa uma atribuição de função a um usuário
type UserRole struct {
	ID           string     `json:"id" bson:"_id"`
	UserID       string     `json:"user_id" bson:"user_id"`
	RoleID       string     `json:"role_id" bson:"role_id"`
	TenantID     string     `json:"tenant_id" bson:"tenant_id"`
	AssignedAt   time.Time  `json:"assigned_at" bson:"assigned_at"`
	AssignedBy   string     `json:"assigned_by" bson:"assigned_by"`
	ExpiresAt    time.Time  `json:"expires_at" bson:"expires_at"`
	Justification string     `json:"justification" bson:"justification"`
	ApprovedBy   string     `json:"approved_by,omitempty" bson:"approved_by,omitempty"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty" bson:"approved_at,omitempty"`
	Status       string     `json:"status" bson:"status"` // ACTIVE, EXPIRED, REVOKED
	RevokedAt    *time.Time `json:"revoked_at,omitempty" bson:"revoked_at,omitempty"`
	RevokedBy    string     `json:"revoked_by,omitempty" bson:"revoked_by,omitempty"`
	RevokeReason string     `json:"revoke_reason,omitempty" bson:"revoke_reason,omitempty"`
}

// AuthenticatedUser representa um usuário autenticado com seus dados de sessão e autorizações
type AuthenticatedUser struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	TenantID    string   `json:"tenant_id"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	AccessLevel string   `json:"access_level"`
	SessionID   string   `json:"session_id"`
	MFAVerified bool     `json:"mfa_verified"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
	ExpiresAt   time.Time `json:"expires_at"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	DeviceID    string    `json:"device_id,omitempty"`
}

// NewAuthenticatedUser cria uma nova instância de usuário autenticado
func NewAuthenticatedUser(user User, roles []string, permissions []string, sessionID string, ipAddress string, userAgent string) AuthenticatedUser {
	return AuthenticatedUser{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		TenantID:    user.TenantID,
		Roles:       roles,
		Permissions: permissions,
		AccessLevel: user.AccessLevel,
		SessionID:   sessionID,
		MFAVerified: user.MFAEnabled,
		Attributes:  user.Attributes,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		DeviceID:    uuid.New().String(),
	}
}