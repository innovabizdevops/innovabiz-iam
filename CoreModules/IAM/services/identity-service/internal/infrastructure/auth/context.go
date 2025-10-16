/**
 * INNOVABIZ IAM - Componente de Autenticação e Contexto
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação do componente de autenticação e gerenciamento de contexto
 * para o módulo Core IAM, seguindo a arquitetura multi-dimensional, multi-tenant
 * e com segurança total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (A.5.15, A.9.4 - Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7.2, 8.3 - Autenticação e autorização)
 * - LGPD/GDPR/PDPA (Arts. 46, 47, 48 - Segurança de dados)
 * - BNA Instrução 7/2021 (Art. 9 - Autenticação e autorização)
 * - SOX (Sec. 404 - Controles internos)
 * - NIST CSF (PR.AC - Gestão de identidades)
 * - OWASP ASVS 4.0 (V2 - Autenticação, V3 - Gerenciamento de Sessão)
 */

package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// Chaves para o contexto
type contextKey string

const (
	// UserContextKey é a chave para o usuário no contexto
	UserContextKey contextKey = "user"
	
	// AuthInfoContextKey é a chave para informações de autenticação no contexto
	AuthInfoContextKey contextKey = "auth_info"
)

// Erros comuns de autenticação
var (
	ErrNoUserInContext = errors.New("nenhum usuário encontrado no contexto")
	ErrInvalidToken    = errors.New("token de autenticação inválido ou expirado")
	ErrInsufficientPermissions = errors.New("permissões insuficientes para realizar esta operação")
)

// User representa um usuário autenticado
type User struct {
	ID       uuid.UUID
	Username string
	Email    string
	TenantID uuid.UUID
	Roles    []string
	Permissions []string
	SessionID uuid.UUID
	AuthMethod string // 'password', 'certificate', 'mfa', 'sso', etc.
}

// AuthInfo contém informações adicionais sobre a autenticação
type AuthInfo struct {
	TokenID        uuid.UUID
	TokenType      string  // 'access', 'refresh', 'apikey', etc.
	ExpiresAt      int64
	IssuedAt       int64
	DeviceID       string
	IPAddress      string
	UserAgent      string
	RequestSource  string // 'api', 'web', 'mobile', etc.
	CorrelationID  string
	SessionContext map[string]interface{}
}

// GetUserFromContext extrai o usuário do contexto
func GetUserFromContext(ctx context.Context) (*User, error) {
	user, ok := ctx.Value(UserContextKey).(*User)
	if !ok || user == nil {
		return nil, ErrNoUserInContext
	}
	return user, nil
}

// EnrichContextWithUser adiciona um usuário ao contexto
func EnrichContextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

// GetAuthInfoFromContext extrai informações de autenticação do contexto
func GetAuthInfoFromContext(ctx context.Context) (*AuthInfo, error) {
	authInfo, ok := ctx.Value(AuthInfoContextKey).(*AuthInfo)
	if !ok || authInfo == nil {
		return nil, errors.New("nenhuma informação de autenticação encontrada no contexto")
	}
	return authInfo, nil
}

// EnrichContextWithAuthInfo adiciona informações de autenticação ao contexto
func EnrichContextWithAuthInfo(ctx context.Context, authInfo *AuthInfo) context.Context {
	return context.WithValue(ctx, AuthInfoContextKey, authInfo)
}

// UserHasRoles verifica se o usuário tem pelo menos um dos roles especificados
func UserHasRoles(ctx context.Context, userID uuid.UUID, requiredRoles []string) (bool, error) {
	// Obter o usuário do contexto
	user, err := GetUserFromContext(ctx)
	if err != nil {
		return false, err
	}
	
	// Verificar se é o mesmo usuário (prevenção contra spoofing)
	if user.ID != userID {
		return false, fmt.Errorf("ID de usuário não corresponde: esperado %s, recebido %s", userID, user.ID)
	}
	
	// Admin sempre tem acesso a tudo
	for _, role := range user.Roles {
		if strings.ToUpper(role) == "ADMIN" || strings.ToUpper(role) == "SUPERADMIN" {
			return true, nil
		}
	}
	
	// Verificar se o usuário tem pelo menos um dos roles necessários
	if len(requiredRoles) == 0 {
		return true, nil // Sem roles requeridos, acesso permitido
	}
	
	for _, requiredRole := range requiredRoles {
		for _, userRole := range user.Roles {
			if strings.ToUpper(userRole) == strings.ToUpper(requiredRole) {
				return true, nil
			}
		}
	}
	
	return false, nil
}

// UserHasPermissions verifica se o usuário tem todas as permissões especificadas
func UserHasPermissions(ctx context.Context, requiredPermissions []string) (bool, error) {
	// Obter o usuário do contexto
	user, err := GetUserFromContext(ctx)
	if err != nil {
		return false, err
	}
	
	// Admin sempre tem acesso a tudo
	for _, role := range user.Roles {
		if strings.ToUpper(role) == "ADMIN" || strings.ToUpper(role) == "SUPERADMIN" {
			return true, nil
		}
	}
	
	// Verificar se o usuário tem todas as permissões necessárias
	for _, requiredPerm := range requiredPermissions {
		found := false
		for _, userPerm := range user.Permissions {
			if strings.ToUpper(userPerm) == strings.ToUpper(requiredPerm) {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}
	
	return true, nil
}

// ContextWithCorrelationID adiciona um ID de correlação ao contexto
func ContextWithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, "correlation_id", correlationID)
}

// GetCorrelationIDFromContext extrai o ID de correlação do contexto
func GetCorrelationIDFromContext(ctx context.Context) string {
	correlationID, ok := ctx.Value("correlation_id").(string)
	if !ok {
		return ""
	}
	return correlationID
}