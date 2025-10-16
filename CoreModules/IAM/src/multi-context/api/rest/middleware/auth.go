/**
 * @file auth.go
 * @description Middleware de autenticação e autorização para API REST
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// TokenClaims representa o conteúdo do token JWT
type TokenClaims struct {
	UserID    string   `json:"sub"`
	TenantID  string   `json:"tid"`
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	IsAdmin   bool     `json:"is_admin"`
	jwt.StandardClaims
}

// UserInfo representa informações do usuário autenticado
type UserInfo struct {
	UserID   string
	TenantID string
	Name     string
	Email    string
	Roles    []string
	IsAdmin  bool
}

// ErrorResponse representa a estrutura de resposta de erro
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// TokenValidator valida o token JWT e extrai informações do usuário
func TokenValidator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extrair token do cabeçalho Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithAuthError(w, "MISSING_TOKEN", "Token de autenticação ausente", http.StatusUnauthorized)
			return
		}

		// Verificar formato do token (Bearer token)
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondWithAuthError(w, "INVALID_TOKEN_FORMAT", "Formato de token inválido", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Validar token
		userInfo, err := validateToken(tokenString)
		if err != nil {
			respondWithAuthError(w, "INVALID_TOKEN", err.Error(), http.StatusUnauthorized)
			return
		}

		// Adicionar informações do usuário ao contexto da requisição
		ctx := context.WithValue(r.Context(), "userInfo", userInfo)
		
		// Adicionar cabeçalhos para camadas subsequentes
		r.Header.Set("X-User-ID", userInfo.UserID)
		r.Header.Set("X-Tenant-ID", userInfo.TenantID)
		if userInfo.IsAdmin {
			r.Header.Set("X-User-Role", "admin")
		}
		if len(userInfo.Roles) > 0 {
			r.Header.Set("X-User-Roles", strings.Join(userInfo.Roles, ","))
		}

		// Prosseguir para o próximo handler com contexto enriquecido
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RoleRequired middleware para verificar se usuário possui um papel específico
func RoleRequired(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extrair informações do usuário do contexto
			userInfo, ok := r.Context().Value("userInfo").(UserInfo)
			if !ok {
				respondWithAuthError(w, "MISSING_AUTH", "Autenticação ausente", http.StatusUnauthorized)
				return
			}

			// Verificar se usuário é admin (bypass)
			if userInfo.IsAdmin {
				next.ServeHTTP(w, r)
				return
			}

			// Verificar se usuário possui o papel requerido
			hasRole := false
			for _, r := range userInfo.Roles {
				if r == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				respondWithAuthError(w, "INSUFFICIENT_PERMISSIONS", fmt.Sprintf("Função %s necessária", role), http.StatusForbidden)
				return
			}

			// Prosseguir para o próximo handler
			next.ServeHTTP(w, r)
		})
	}
}

// AdminRequired middleware para verificar se usuário é administrador
func AdminRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extrair informações do usuário do contexto
		userInfo, ok := r.Context().Value("userInfo").(UserInfo)
		if !ok {
			respondWithAuthError(w, "MISSING_AUTH", "Autenticação ausente", http.StatusUnauthorized)
			return
		}

		// Verificar se usuário é admin
		if !userInfo.IsAdmin {
			respondWithAuthError(w, "ADMIN_REQUIRED", "Permissão de administrador necessária", http.StatusForbidden)
			return
		}

		// Prosseguir para o próximo handler
		next.ServeHTTP(w, r)
	})
}

// validateToken valida um token JWT e extrai as informações do usuário
// Em um ambiente de produção, este método usaria uma chave secreta configurada
// e validaria o token com biblioteca específica
func validateToken(tokenString string) (UserInfo, error) {
	// TODO: Em produção, usar uma chave secreta configurada
	secretKey := []byte("INNOVABIZ_SECRET_KEY_REPLACE_IN_PRODUCTION")

	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validar o algoritmo de assinatura
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inesperado: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return UserInfo{}, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		// Verificar expiração
		if claims.ExpiresAt < time.Now().Unix() {
			return UserInfo{}, errors.New("token expirado")
		}

		// Extrair informações do usuário
		return UserInfo{
			UserID:   claims.UserID,
			TenantID: claims.TenantID,
			Name:     claims.Name,
			Email:    claims.Email,
			Roles:    claims.Roles,
			IsAdmin:  claims.IsAdmin,
		}, nil
	}

	return UserInfo{}, errors.New("token inválido")
}

// respondWithAuthError responde com um erro de autenticação no formato padrão
func respondWithAuthError(w http.ResponseWriter, code string, message string, status int) {
	response := ErrorResponse{}
	response.Error.Code = code
	response.Error.Message = message

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}