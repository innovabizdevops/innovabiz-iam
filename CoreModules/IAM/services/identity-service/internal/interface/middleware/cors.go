package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// CORSConfig representa as configurações para o middleware CORS
type CORSConfig struct {
	// AllowedOrigins é a lista de origens permitidas para fazer requisições CORS
	// Valores especiais:
	// - "*": Permite todas as origens (não recomendado para produção)
	// - "null": Permite requisições de arquivos locais (file://)
	AllowedOrigins []string

	// AllowedMethods é a lista de métodos HTTP permitidos para requisições CORS
	AllowedMethods []string

	// AllowedHeaders é a lista de cabeçalhos HTTP permitidos para requisições CORS
	AllowedHeaders []string

	// ExposedHeaders é a lista de cabeçalhos HTTP que serão expostos na resposta
	ExposedHeaders []string

	// AllowCredentials indica se as requisições podem incluir credenciais (cookies, auth headers)
	AllowCredentials bool

	// MaxAge é o tempo máximo, em segundos, que o cliente deve armazenar a resposta preflight
	MaxAge int

	// OptionsPassthrough indica se o middleware deve passar requisições OPTIONS para os handlers
	OptionsPassthrough bool
}

// DefaultCORSConfig retorna uma configuração CORS padrão
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:     []string{"*"},
		AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Tenant-ID", "X-User-ID", "X-Request-ID"},
		ExposedHeaders:     []string{"Content-Length", "Content-Type", "X-Request-ID"},
		AllowCredentials:   true,
		MaxAge:             86400, // 24 horas
		OptionsPassthrough: false,
	}
}

// EnhancedCORSMiddleware cria um middleware CORS melhorado com opções de segurança avançadas,
// conforme recomendações OWASP, ISO/IEC 27001 e melhores práticas de segurança web
func EnhancedCORSMiddleware(logger zerolog.Logger, config CORSConfig) func(http.Handler) http.Handler {
	tracer := otel.GetTracerProvider().Tracer("innovabiz.iam.middleware")
	
	// Normalizar as origens permitidas
	allowedOrigins := normalizeOrigins(config.AllowedOrigins)
	allowAll := containsWildcard(allowedOrigins)
	
	// Converter arrays para map para busca mais rápida
	allowedOriginsMap := makeMap(allowedOrigins)
	allowedMethodsMap := makeMap(config.AllowedMethods)
	allowedHeadersMap := makeMap(normalizeCORSHeaders(config.AllowedHeaders))
	exposedHeadersMap := makeMap(config.ExposedHeaders)
	
	// Serializar para Cache-Control
	maxAgeStr := ""
	if config.MaxAge > 0 {
		maxAgeStr = "max-age=" + string(config.MaxAge)
	}
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracer.Start(r.Context(), "cors.middleware")
			defer span.End()
			
			origin := r.Header.Get("Origin")
			
			// Se não é uma requisição CORS (sem Origin header), apenas continuar
			if origin == "" {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			
			span.SetAttributes(attribute.String("cors.origin", origin))
			
			// Se a requisição é de um iframe (verificar Sec-Fetch-Dest)
			if r.Header.Get("Sec-Fetch-Dest") == "iframe" {
				// Adicionar headers de segurança para iframes
				w.Header().Set("X-Frame-Options", "SAMEORIGIN")
				w.Header().Set("Content-Security-Policy", "frame-ancestors 'self'")
			}
			
			// Sempre adicionar headers de segurança básicos
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			
			// Verificar se a origem é permitida
			originAllowed := allowAll || allowedOriginsMap[origin] || allowedOriginsMap[originHostname(origin)]
			
			if !originAllowed {
				logger.Debug().
					Str("origin", origin).
					Strs("allowed_origins", config.AllowedOrigins).
					Msg("CORS: Origem não permitida")
				
				span.SetAttributes(attribute.Bool("cors.origin_allowed", false))
				
				// Se não é uma preflight, continuar sem adicionar headers CORS
				if r.Method != http.MethodOptions {
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				
				// Se é uma preflight, retornar 403 Forbidden
				w.WriteHeader(http.StatusForbidden)
				return
			}
			
			span.SetAttributes(attribute.Bool("cors.origin_allowed", true))
			
			// Configurar headers de resposta CORS
			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				if config.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				w.Header().Add("Vary", "Origin")
			}
			
			// Adicionar headers expostos, se houver
			if len(config.ExposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
			}
			
			// Processar requisições preflight (OPTIONS)
			if r.Method == http.MethodOptions {
				// Verificar se há cabeçalho Access-Control-Request-Method
				requestMethod := r.Header.Get("Access-Control-Request-Method")
				if requestMethod == "" {
					// Não é uma requisição preflight, pode ser uma requisição OPTIONS normal
					if config.OptionsPassthrough {
						next.ServeHTTP(w, r.WithContext(ctx))
					} else {
						w.WriteHeader(http.StatusNoContent)
					}
					return
				}
				
				span.SetAttributes(attribute.String("cors.request_method", requestMethod))
				
				// Verificar se o método solicitado é permitido
				if !allowedMethodsMap[requestMethod] {
					logger.Debug().
						Str("request_method", requestMethod).
						Strs("allowed_methods", config.AllowedMethods).
						Msg("CORS: Método não permitido")
					
					span.SetAttributes(attribute.Bool("cors.method_allowed", false))
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				
				span.SetAttributes(attribute.Bool("cors.method_allowed", true))
				
				// Verificar headers solicitados
				requestHeaders := r.Header.Get("Access-Control-Request-Headers")
				if requestHeaders != "" {
					headers := normalizeCORSHeaders(strings.Split(requestHeaders, ","))
					
					span.SetAttributes(attribute.StringSlice("cors.request_headers", headers))
					
					// Verificar se todos os cabeçalhos solicitados são permitidos
					for _, header := range headers {
						if !allowedHeadersMap[header] {
							logger.Debug().
								Str("request_header", header).
								Strs("allowed_headers", config.AllowedHeaders).
								Msg("CORS: Cabeçalho não permitido")
							
							span.SetAttributes(attribute.Bool("cors.headers_allowed", false))
							w.WriteHeader(http.StatusForbidden)
							return
						}
					}
					
					span.SetAttributes(attribute.Bool("cors.headers_allowed", true))
					
					// Adicionar cabeçalhos ao Vary para caching correto
					w.Header().Add("Vary", "Access-Control-Request-Headers")
				}
				
				// Configurar headers de resposta preflight
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
				if len(config.AllowedHeaders) > 0 {
					w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
				}
				if config.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", maxAgeStr)
				}
				
				// Responder à requisição preflight
				w.WriteHeader(http.StatusNoContent)
				return
			}
			
			// Para requisições não-OPTIONS, continuar com o processamento
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// normalizeOrigins normaliza uma lista de origens
func normalizeOrigins(origins []string) []string {
	if len(origins) == 0 {
		return []string{}
	}
	
	result := make([]string, 0, len(origins))
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			result = append(result, origin)
		}
	}
	
	return result
}

// containsWildcard verifica se a lista contém o caractere curinga "*"
func containsWildcard(values []string) bool {
	for _, v := range values {
		if v == "*" {
			return true
		}
	}
	return false
}

// makeMap converte uma lista para um mapa para busca mais rápida
func makeMap(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, v := range values {
		result[v] = true
	}
	return result
}

// originHostname extrai o hostname de uma origem
func originHostname(origin string) string {
	// Remover protocolo
	if i := strings.Index(origin, "://"); i >= 0 {
		origin = origin[i+3:]
	}
	
	// Remover porta
	if i := strings.Index(origin, ":"); i >= 0 {
		origin = origin[:i]
	}
	
	return origin
}

// normalizeCORSHeaders normaliza os cabeçalhos CORS
func normalizeCORSHeaders(headers []string) []string {
	if len(headers) == 0 {
		return []string{}
	}
	
	result := make([]string, 0, len(headers))
	for _, header := range headers {
		header = strings.TrimSpace(header)
		if header != "" {
			result = append(result, http.CanonicalHeaderKey(header))
		}
	}
	
	return result
}