/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Middleware de autenticação e autorização.
 * Implementa a verificação de tokens JWT, extração de claims e aplicação de políticas de segurança.
 */

package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/innovabiz/iam/services/identity-service/internal/application"
	"github.com/innovabiz/iam/services/identity-service/internal/config"
)

// Context keys para armazenar informações do usuário no contexto da requisição
type contextKey string

const (
	UserIDKey       = contextKey("user_id")
	TenantIDKey     = contextKey("tenant_id")
	UsernameKey     = contextKey("username")
	RolesKey        = contextKey("roles")
	PermissionsKey  = contextKey("permissions")
	AuthTokenKey    = contextKey("auth_token")
	AuthorizedKey   = contextKey("authorized")
)

// AuthMiddleware verifica tokens JWT e extrai claims para o contexto
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extrair e validar o token de autorização
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			respondWithError(w, http.StatusUnauthorized, "unauthorized", "Token de autorização não fornecido")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		
		// Obter configuração
		cfg := config.GetConfig()
		
		// Validar o token
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			// Verificar algoritmo de assinatura
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("algoritmo de assinatura inesperado: %v", token.Header["alg"])
			}
			// Retornar a chave secreta para validação
			return []byte(cfg.JWT.Secret), nil
		})
		
		if err != nil {
			log.Error().Err(err).Msg("Erro ao validar token JWT")
			respondWithError(w, http.StatusUnauthorized, "invalid_token", "Token inválido ou expirado")
			return
		}
		
		// Verificar claims do token
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Extrair user ID do token
			userIDStr, ok := claims["sub"].(string)
			if !ok {
				respondWithError(w, http.StatusUnauthorized, "invalid_token_claims", "Token não contém ID do usuário")
				return
			}
			
			// Extrair tenant ID do token
			tenantIDStr, ok := claims["tenant_id"].(string)
			if !ok {
				respondWithError(w, http.StatusUnauthorized, "invalid_token_claims", "Token não contém ID do tenant")
				return
			}
			
			// Extrair username do token
			username, _ := claims["username"].(string)
			
			// Extrair roles do token (opcional)
			var roles []string
			if rolesInterface, ok := claims["roles"]; ok {
				if rolesArray, ok := rolesInterface.([]interface{}); ok {
					roles = make([]string, len(rolesArray))
					for i, role := range rolesArray {
						if roleStr, ok := role.(string); ok {
							roles[i] = roleStr
						}
					}
				}
			}
			
			// Extrair permissões do token (opcional)
			var permissions []string
			if permsInterface, ok := claims["permissions"]; ok {
				if permsArray, ok := permsInterface.([]interface{}); ok {
					permissions = make([]string, len(permsArray))
					for i, perm := range permsArray {
						if permStr, ok := perm.(string); ok {
							permissions[i] = permStr
						}
					}
				}
			}
			
			// Verificar tempo de expiração
			if exp, ok := claims["exp"].(float64); !ok || float64(jwt.NewNumericDate(config.GetCurrentTime()).Unix()) > exp {
				respondWithError(w, http.StatusUnauthorized, "token_expired", "Token expirado")
				return
			}

			// Adicionar informações ao contexto da requisição
			ctx := context.WithValue(r.Context(), UserIDKey, userIDStr)
			ctx = context.WithValue(ctx, TenantIDKey, tenantIDStr)
			ctx = context.WithValue(ctx, UsernameKey, username)
			ctx = context.WithValue(ctx, RolesKey, roles)
			ctx = context.WithValue(ctx, PermissionsKey, permissions)
			ctx = context.WithValue(ctx, AuthTokenKey, tokenStr)
			ctx = context.WithValue(ctx, AuthorizedKey, true)
			
			// Adicionar tenant ID ao header para o Row-Level Security do PostgreSQL
			// Este header será usado ao configurar a conexão com o banco
			r.Header.Set("X-Tenant-ID", tenantIDStr)
			
			// Propagar para o próximo handler com o contexto enriquecido
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			respondWithError(w, http.StatusUnauthorized, "invalid_token", "Token inválido")
			return
		}
	})
}

// RequireRoles verifica se o usuário possui pelo menos um dos roles especificados
func RequireRoles(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verificar se o usuário está autenticado
			authorized, ok := r.Context().Value(AuthorizedKey).(bool)
			if !ok || !authorized {
				respondWithError(w, http.StatusUnauthorized, "unauthorized", "Autenticação requerida")
				return
			}
			
			// Extrair roles do usuário do contexto
			userRoles, ok := r.Context().Value(RolesKey).([]string)
			if !ok {
				respondWithError(w, http.StatusForbidden, "forbidden", "Sem permissão para acessar este recurso")
				return
			}
			
			// Verificar se o usuário tem pelo menos um dos roles necessários
			hasRole := false
			for _, requiredRole := range roles {
				for _, userRole := range userRoles {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}
			
			if !hasRole {
				respondWithError(w, http.StatusForbidden, "forbidden", "Sem permissão para acessar este recurso")
				return
			}
			
			// Propagar para o próximo handler
			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermissions verifica se o usuário possui todas as permissões especificadas
func RequirePermissions(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verificar se o usuário está autenticado
			authorized, ok := r.Context().Value(AuthorizedKey).(bool)
			if !ok || !authorized {
				respondWithError(w, http.StatusUnauthorized, "unauthorized", "Autenticação requerida")
				return
			}
			
			// Extrair permissões do usuário do contexto
			userPerms, ok := r.Context().Value(PermissionsKey).([]string)
			if !ok {
				respondWithError(w, http.StatusForbidden, "forbidden", "Sem permissão para acessar este recurso")
				return
			}
			
			// Verificar se o usuário tem todas as permissões necessárias
			for _, requiredPerm := range permissions {
				hasPerm := false
				for _, userPerm := range userPerms {
					if userPerm == requiredPerm {
						hasPerm = true
						break
					}
				}
				
				if !hasPerm {
					respondWithError(w, http.StatusForbidden, "forbidden", "Sem permissão para acessar este recurso")
					return
				}
			}
			
			// Propagar para o próximo handler
			next.ServeHTTP(w, r)
		})
	}
}

// ValidateTenant garante que o tenant na requisição corresponda ao tenant do token
func ValidateTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extrair tenant ID do token (contexto)
		tokenTenantID, ok := r.Context().Value(TenantIDKey).(string)
		if !ok {
			respondWithError(w, http.StatusUnauthorized, "unauthorized", "Tenant ID não encontrado no token")
			return
		}
		
		// Extrair tenant ID da requisição (cabeçalho ou parâmetro)
		reqTenantID := r.Header.Get("X-Tenant-ID")
		if reqTenantID == "" {
			// Tentar extrair do parâmetro de consulta
			reqTenantID = r.URL.Query().Get("tenant_id")
		}
		
		// Se o tenant ID for fornecido na requisição, verificar se corresponde ao token
		if reqTenantID != "" && reqTenantID != tokenTenantID {
			log.Warn().
				Str("token_tenant_id", tokenTenantID).
				Str("request_tenant_id", reqTenantID).
				Msg("Tentativa de acessar tenant diferente do autenticado")
				
			respondWithError(w, http.StatusForbidden, "invalid_tenant", "Não é permitido acessar dados de outro tenant")
			return
		}
		
		// Propagar para o próximo handler
		next.ServeHTTP(w, r)
	})
}

// TracingMiddleware adiciona tracing distribuído às requisições
func TracingMiddleware(tracer trace.Tracer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Iniciar um novo span para a requisição
			ctx, span := tracer.Start(r.Context(), fmt.Sprintf("%s %s", r.Method, r.URL.Path))
			defer span.End()
			
			// Adicionar atributos úteis ao span
			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.String()),
				attribute.String("http.host", r.Host),
				attribute.String("http.user_agent", r.UserAgent()),
			)
			
			// Extrair ID da requisição
			requestID := r.Header.Get("X-Request-ID")
			if requestID != "" {
				span.SetAttributes(attribute.String("request_id", requestID))
			}
			
			// Propagar para o próximo handler com o contexto contendo o span
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TenantContextMiddleware injeta tenant ID no contexto do PostgreSQL para Row-Level Security
func TenantContextMiddleware(authService application.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extrair e validar o token de autorização
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
				
				// Verificar o token com o serviço de autenticação
				verifyReq := application.VerifyTokenRequest{
					Token: tokenStr,
				}
				
				response, err := authService.VerifyToken(r.Context(), verifyReq)
				if err == nil && response.Valid {
					// Token válido, extrair tenant ID
					tenantID, err := uuid.Parse(response.TenantID)
					if err == nil {
						// Definir tenant ID no header para uso pelo postgres_middleware
						r.Header.Set("X-Tenant-ID", tenantID.String())
					}
				}
			}
			
			// Propagar para o próximo handler
			next.ServeHTTP(w, r)
		})
	}
}

// respondWithError envia uma resposta de erro JSON
func respondWithError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	response := struct {
		Error   string `json:"error"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}{
		Error:   code,
		Code:    code,
		Message: message,
	}
	
	// Encode to JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().Err(err).Msg("Erro ao serializar resposta de erro")
	}
}