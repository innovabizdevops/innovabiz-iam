package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Chaves do contexto para informações do usuário autenticado
type contextKey string

const (
	TenantIDContextKey contextKey = "tenant_id"
	UserIDContextKey   contextKey = "user_id"
	UsernameContextKey contextKey = "username"
	RolesContextKey    contextKey = "roles"
)

// Claims representa as reivindicações (claims) customizadas do JWT
type Claims struct {
	jwt.RegisteredClaims
	TenantID  string   `json:"tid,omitempty"`
	Username  string   `json:"preferred_username,omitempty"`
	Roles     []string `json:"roles,omitempty"`
	Email     string   `json:"email,omitempty"`
	FirstName string   `json:"given_name,omitempty"`
	LastName  string   `json:"family_name,omitempty"`
}

// AuthConfig representa a configuração do middleware de autenticação
type AuthConfig struct {
	JWTSecret            string
	JWTIssuer            string
	JWTAudience          string
	AllowedAlgorithms    []string
	TokenLookup          string
	TokenHeaderName      string
	TokenQueryParamName  string
	TokenCookieName      string
	TokenPrefix          string
	DisableAuthentication bool
	SkipPaths           []string
}

// DefaultAuthConfig retorna uma configuração padrão para autenticação
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		JWTSecret:           os.Getenv("JWT_SECRET"),
		JWTIssuer:           os.Getenv("JWT_ISSUER"),
		JWTAudience:         os.Getenv("JWT_AUDIENCE"),
		AllowedAlgorithms:   []string{"HS256", "RS256"},
		TokenLookup:         "header:Authorization",
		TokenHeaderName:     "Authorization",
		TokenQueryParamName: "token",
		TokenCookieName:     "jwt",
		TokenPrefix:         "Bearer ",
		DisableAuthentication: os.Getenv("DISABLE_AUTH") == "true",
		SkipPaths:          []string{"/health", "/ready", "/docs/"},
	}
}

// AuthMiddleware cria um middleware para autenticação baseada em JWT
// Implementado de acordo com as melhores práticas de segurança ISO/IEC 27001, PCI DSS e OWASP
func AuthMiddleware(logger zerolog.Logger, config AuthConfig) func(http.Handler) http.Handler {
	tracer := otel.GetTracerProvider().Tracer("innovabiz.iam.middleware")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracer.Start(r.Context(), "auth.middleware")
			defer span.End()

			// Verificar se o caminho deve ser ignorado (endpoints públicos)
			for _, path := range config.SkipPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Se a autenticação estiver desabilitada (ambiente de desenvolvimento), 
			// usar headers simulados para tenant_id e user_id
			if config.DisableAuthentication {
				span.AddEvent("auth.disabled")
				logger.Warn().Msg("Autenticação desabilitada. Utilizando headers simulados para ambiente de desenvolvimento.")

				tenantID := r.Header.Get("X-Tenant-ID")
				if tenantID == "" {
					tenantID = "11111111-1111-1111-1111-111111111111" // Tenant padrão para desenvolvimento
				}

				userID := r.Header.Get("X-User-ID")
				if userID == "" {
					userID = "00000000-0000-0000-0000-000000000000" // Usuário padrão para desenvolvimento
				}

				// Adicionar informações ao contexto
				ctx = context.WithValue(ctx, TenantIDContextKey, tenantID)
				ctx = context.WithValue(ctx, UserIDContextKey, userID)
				ctx = context.WithValue(ctx, UsernameContextKey, "dev_user")
				ctx = context.WithValue(ctx, RolesContextKey, []string{"admin"})

				// Continuar com o processamento da requisição
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Obter token da requisição
			tokenString, err := extractToken(r, config)
			if err != nil {
				span.SetStatus(codes.Error, "Token não fornecido ou em formato inválido")
				handleAuthError(w, http.StatusUnauthorized, "missing_token", "Token de autenticação não fornecido ou em formato inválido", logger)
				return
			}

			span.SetAttributes(attribute.Bool("token.present", true))

			// Validar o token
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				// Verificar se o algoritmo é permitido
				alg := token.Method.Alg()
				allowed := false
				for _, allowedAlg := range config.AllowedAlgorithms {
					if alg == allowedAlg {
						allowed = true
						break
					}
				}
				if !allowed {
					return nil, fmt.Errorf("algoritmo de assinatura não permitido: %s", alg)
				}

				// Para HMAC, retornar a chave secreta
				if strings.HasPrefix(alg, "HS") {
					return []byte(config.JWTSecret), nil
				}

				// Para RSA, retornar a chave pública (implementação pendente)
				return nil, fmt.Errorf("algoritmo não implementado: %s", alg)
			})

			if err != nil {
				span.SetStatus(codes.Error, fmt.Sprintf("Token inválido: %v", err))
				span.RecordError(err)
				handleAuthError(w, http.StatusUnauthorized, "invalid_token", "Token de autenticação inválido", logger)
				return
			}

			if !token.Valid {
				span.SetStatus(codes.Error, "Token inválido")
				handleAuthError(w, http.StatusUnauthorized, "invalid_token", "Token de autenticação inválido", logger)
				return
			}

			// Validar claims adicionais (issuer, audience, expiration)
			if config.JWTIssuer != "" && claims.Issuer != config.JWTIssuer {
				span.SetStatus(codes.Error, "Issuer inválido")
				handleAuthError(w, http.StatusUnauthorized, "invalid_issuer", "Emissor do token inválido", logger)
				return
			}

			if config.JWTAudience != "" {
				validAudience := false
				for _, aud := range claims.Audience {
					if aud == config.JWTAudience {
						validAudience = true
						break
					}
				}
				if !validAudience {
					span.SetStatus(codes.Error, "Audience inválida")
					handleAuthError(w, http.StatusUnauthorized, "invalid_audience", "Audiência do token inválida", logger)
					return
				}
			}

			// Verificar expiration time
			if claims.ExpiresAt != nil {
				if time.Now().After(claims.ExpiresAt.Time) {
					span.SetStatus(codes.Error, "Token expirado")
					handleAuthError(w, http.StatusUnauthorized, "token_expired", "Token de autenticação expirado", logger)
					return
				}
			}

			// Validar tenant_id e user_id
			tenantID, err := uuid.Parse(claims.TenantID)
			if err != nil {
				span.SetStatus(codes.Error, "TenantID inválido")
				handleAuthError(w, http.StatusUnauthorized, "invalid_tenant", "TenantID inválido no token", logger)
				return
			}

			// O sub (subject) do JWT deve ser o user_id
			userID, err := uuid.Parse(claims.Subject)
			if err != nil {
				span.SetStatus(codes.Error, "UserID inválido")
				handleAuthError(w, http.StatusUnauthorized, "invalid_user", "UserID inválido no token", logger)
				return
			}

			// Adicionar claims ao span para observabilidade
			span.SetAttributes(
				attribute.String("user.id", userID.String()),
				attribute.String("tenant.id", tenantID.String()),
				attribute.String("user.name", claims.Username),
				attribute.StringSlice("user.roles", claims.Roles),
			)

			// Adicionar informações ao contexto
			ctx = context.WithValue(ctx, TenantIDContextKey, tenantID)
			ctx = context.WithValue(ctx, UserIDContextKey, userID)
			ctx = context.WithValue(ctx, UsernameContextKey, claims.Username)
			ctx = context.WithValue(ctx, RolesContextKey, claims.Roles)

			// Continuar com o processamento da requisição
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractToken extrai o token JWT da requisição
func extractToken(r *http.Request, config AuthConfig) (string, error) {
	// Verificar no header
	if strings.HasPrefix(config.TokenLookup, "header:") {
		headerName := strings.TrimPrefix(config.TokenLookup, "header:")
		if headerName == "" {
			headerName = config.TokenHeaderName
		}
		
		authHeader := r.Header.Get(headerName)
		if authHeader == "" {
			return "", fmt.Errorf("header de autenticação não encontrado")
		}
		
		// Remover prefixo (ex: "Bearer ")
		if config.TokenPrefix != "" && strings.HasPrefix(authHeader, config.TokenPrefix) {
			return strings.TrimPrefix(authHeader, config.TokenPrefix), nil
		}
		
		return authHeader, nil
	}
	
	// Verificar em query param
	if strings.HasPrefix(config.TokenLookup, "query:") {
		paramName := strings.TrimPrefix(config.TokenLookup, "query:")
		if paramName == "" {
			paramName = config.TokenQueryParamName
		}
		
		token := r.URL.Query().Get(paramName)
		if token == "" {
			return "", fmt.Errorf("token não encontrado no parâmetro de consulta")
		}
		
		return token, nil
	}
	
	// Verificar em cookie
	if strings.HasPrefix(config.TokenLookup, "cookie:") {
		cookieName := strings.TrimPrefix(config.TokenLookup, "cookie:")
		if cookieName == "" {
			cookieName = config.TokenCookieName
		}
		
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			return "", fmt.Errorf("cookie de autenticação não encontrado")
		}
		
		return cookie.Value, nil
	}
	
	return "", fmt.Errorf("método de extração de token não suportado")
}

// handleAuthError responde com erro de autenticação
func handleAuthError(w http.ResponseWriter, status int, code, message string, logger zerolog.Logger) {
	logger.Info().
		Int("status", status).
		Str("code", code).
		Str("message", message).
		Msg("Erro de autenticação")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	type errorResponse struct {
		Status  int    `json:"status"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	
	json.NewEncoder(w).Encode(errorResponse{
		Status:  status,
		Code:    code,
		Message: message,
	})
}

// Helper functions to get values from context

// GetTenantID retorna o ID do tenant do contexto
func GetTenantID(ctx context.Context) (uuid.UUID, error) {
	tenantID, ok := ctx.Value(TenantIDContextKey).(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("tenant_id não encontrado no contexto")
	}
	
	return uuid.Parse(tenantID)
}

// GetUserID retorna o ID do usuário do contexto
func GetUserID(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(UserIDContextKey).(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("user_id não encontrado no contexto")
	}
	
	return uuid.Parse(userID)
}

// GetUsername retorna o nome do usuário do contexto
func GetUsername(ctx context.Context) (string, error) {
	username, ok := ctx.Value(UsernameContextKey).(string)
	if !ok {
		return "", fmt.Errorf("username não encontrado no contexto")
	}
	
	return username, nil
}

// GetRoles retorna as funções do usuário do contexto
func GetRoles(ctx context.Context) ([]string, error) {
	roles, ok := ctx.Value(RolesContextKey).([]string)
	if !ok {
		return nil, fmt.Errorf("roles não encontrados no contexto")
	}
	
	return roles, nil
}