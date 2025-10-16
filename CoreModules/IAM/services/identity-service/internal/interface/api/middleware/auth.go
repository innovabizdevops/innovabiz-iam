package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type contextKey string

const (
	// UserIDKey é a chave usada para armazenar o ID do usuário no contexto
	UserIDKey contextKey = "user_id"
	// TenantIDKey é a chave usada para armazenar o ID do tenant no contexto
	TenantIDKey contextKey = "tenant_id"
)

var tracer = otel.Tracer("innovabiz.iam.interface.api.middleware.auth")

// AuthMiddleware é um middleware para autenticação de usuários
type AuthMiddleware struct {
	// Dependências podem ser adicionadas aqui, como serviços de autenticação
}

// NewAuthMiddleware cria uma nova instância de AuthMiddleware
func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

// Authenticate autentica o usuário da requisição
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "AuthMiddleware.Authenticate", trace.WithAttributes(
			attribute.String("path", r.URL.Path),
			attribute.String("method", r.Method),
		))
		defer span.End()

		// Obter token de autenticação do cabeçalho
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Error().Str("path", r.URL.Path).Msg("Token de autenticação não encontrado")
			http.Error(w, "Não autorizado: token ausente", http.StatusUnauthorized)
			return
		}

		// Validar formato do token (Bearer)
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Error().Str("path", r.URL.Path).Msg("Formato de token inválido")
			http.Error(w, "Não autorizado: formato de token inválido", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// TODO: Implementar lógica real de validação do token
		// Este é apenas um exemplo para desenvolvimento inicial

		// Extrair userID e tenantID do token
		// Em uma implementação real, isso seria feito verificando o token JWT
		// e extraindo as claims necessárias
		
		// Para fins de demonstração, vamos usar IDs fixos
		userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")  // ID de exemplo para desenvolvimento
		
		// Armazenar no contexto
		ctx = context.WithValue(ctx, UserIDKey, userID)
		
		// Adicionar informações do usuário no span para observabilidade
		span.SetAttributes(attribute.String("user_id", userID.String()))

		// Continuar a requisição com o contexto atualizado
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext recupera o ID do usuário do contexto
func GetUserIDFromContext(ctx context.Context) uuid.UUID {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return userID
}

// GetTenantIDFromContext recupera o ID do tenant do contexto
func GetTenantIDFromContext(ctx context.Context) uuid.UUID {
	tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return tenantID
}

// RequireTenantAccess verifica se o usuário tem acesso ao tenant
func RequireTenantAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "AuthMiddleware.RequireTenantAccess")
		defer span.End()

		userID := GetUserIDFromContext(ctx)
		if userID == uuid.Nil {
			log.Error().Str("path", r.URL.Path).Msg("Usuário não autenticado")
			http.Error(w, "Não autorizado", http.StatusUnauthorized)
			return
		}

		// TODO: Implementar verificação de acesso ao tenant
		// Neste ponto, verificaríamos se o usuário tem acesso ao tenant especificado
		// Para desenvolvimento inicial, vamos permitir o acesso

		next.ServeHTTP(w, r)
	})
}

// RequirePermission verifica se o usuário tem a permissão necessária
func RequirePermission(permissionCode string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracer.Start(r.Context(), "AuthMiddleware.RequirePermission", trace.WithAttributes(
				attribute.String("permission_code", permissionCode),
			))
			defer span.End()

			userID := GetUserIDFromContext(ctx)
			if userID == uuid.Nil {
				log.Error().Str("path", r.URL.Path).Str("permission", permissionCode).Msg("Usuário não autenticado")
				http.Error(w, "Não autorizado", http.StatusUnauthorized)
				return
			}

			// TODO: Implementar verificação de permissão
			// Aqui verificaríamos se o usuário tem a permissão específica
			// Para desenvolvimento inicial, vamos permitir o acesso

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}