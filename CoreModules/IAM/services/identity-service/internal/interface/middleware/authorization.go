package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// AuthzConfig representa a configuração do middleware de autorização
type AuthzConfig struct {
	OPAEndpoint        string
	PolicyPath         string
	DecisionPath       string
	Timeout            time.Duration
	DisableAuthorization bool
	SkipPaths          []string
}

// DefaultAuthzConfig retorna uma configuração padrão para autorização
func DefaultAuthzConfig() AuthzConfig {
	return AuthzConfig{
		OPAEndpoint:        "http://opa:8181/v1/data",
		PolicyPath:         "innovabiz/iam/authz",
		DecisionPath:       "allow",
		Timeout:            500 * time.Millisecond,
		DisableAuthorization: false,
		SkipPaths:          []string{"/health", "/ready", "/docs/"},
	}
}

// AuthorizationMiddleware cria um middleware para autorização baseada em políticas (ABAC)
// utilizando Open Policy Agent (OPA), seguindo as melhores práticas de TOGAF, COBIT e ISO 27001
func AuthorizationMiddleware(logger zerolog.Logger, config AuthzConfig) func(http.Handler) http.Handler {
	tracer := otel.GetTracerProvider().Tracer("innovabiz.iam.middleware")
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracer.Start(r.Context(), "authz.middleware")
			defer span.End()
			
			// Verificar se o caminho deve ser ignorado
			for _, path := range config.SkipPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			
			// Se a autorização estiver desabilitada, continuar com o processamento
			if config.DisableAuthorization {
				span.AddEvent("authz.disabled")
				logger.Warn().Msg("Autorização desabilitada. Continuando sem verificação de permissões.")
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			
			// Extrair informações de autenticação do contexto
			tenantID, err := GetTenantID(ctx)
			if err != nil {
				span.SetStatus(codes.Error, "TenantID não encontrado no contexto")
				span.RecordError(err)
				handleAuthError(w, http.StatusUnauthorized, "missing_tenant", "TenantID não encontrado", logger)
				return
			}
			
			userID, err := GetUserID(ctx)
			if err != nil {
				span.SetStatus(codes.Error, "UserID não encontrado no contexto")
				span.RecordError(err)
				handleAuthError(w, http.StatusUnauthorized, "missing_user", "UserID não encontrado", logger)
				return
			}
			
			roles, err := GetRoles(ctx)
			if err != nil {
				span.SetStatus(codes.Error, "Roles não encontrados no contexto")
				span.RecordError(err)
				handleAuthError(w, http.StatusUnauthorized, "missing_roles", "Roles não encontrados", logger)
				return
			}

			// Extrair variáveis de rota para o contexto de autorização
			vars := mux.Vars(r)
			
			// Construir input para o OPA
			input := map[string]interface{}{
				"user": map[string]interface{}{
					"id":     userID.String(),
					"roles":  roles,
				},
				"tenant": map[string]interface{}{
					"id": tenantID.String(),
				},
				"request": map[string]interface{}{
					"method": r.Method,
					"path":   r.URL.Path,
					"vars":   vars,
					"query":  parseQueryParams(r),
				},
			}

			// Adicionar corpo da requisição para métodos POST, PUT e PATCH, se necessário
			// Nota: Isso pode impactar a performance e aumentar o consumo de memória
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				if r.Body != nil && r.ContentLength > 0 && r.ContentLength < 10_000 { // Limitar tamanho do corpo
					bodyBytes, _ := io.ReadAll(r.Body)
					// Restaurar o corpo para leitura posterior
					r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
					
					var bodyJSON map[string]interface{}
					if err := json.Unmarshal(bodyBytes, &bodyJSON); err == nil {
						input["request"].(map[string]interface{})["body"] = bodyJSON
					}
				}
			}

			// Registrar input para debugging (versão sanitizada)
			if logger.GetLevel() <= zerolog.DebugLevel {
				inputCopy := sanitizeMap(input)
				logger.Debug().
					Interface("authorization_input", inputCopy).
					Str("policy_path", config.PolicyPath).
					Msg("Avaliando autorização com OPA")
			}
			
			// Adicionar atributos ao span para observabilidade
			span.SetAttributes(
				attribute.String("policy_path", config.PolicyPath),
				attribute.String("decision_path", config.DecisionPath),
				attribute.String("request.method", r.Method),
				attribute.String("request.path", r.URL.Path),
			)
			
			// Criar timeout para a requisição
			authzCtx, cancel := context.WithTimeout(ctx, config.Timeout)
			defer cancel()
			
			// Criar payload para o OPA
			opaInput := map[string]interface{}{
				"input": input,
			}
			
			// Converter payload para JSON
			opaPayload, err := json.Marshal(opaInput)
			if err != nil {
				span.SetStatus(codes.Error, "Erro ao serializar input para OPA")
				span.RecordError(err)
				logger.Error().Err(err).Msg("Falha ao serializar input para OPA")
				handleAuthError(w, http.StatusInternalServerError, "authz_error", "Erro interno de autorização", logger)
				return
			}
			
			// Construir URL para o OPA
			opaURL := fmt.Sprintf("%s/%s", config.OPAEndpoint, config.PolicyPath)
			
			// Criar requisição para o OPA
			req, err := http.NewRequestWithContext(authzCtx, "POST", opaURL, bytes.NewBuffer(opaPayload))
			if err != nil {
				span.SetStatus(codes.Error, "Erro ao criar requisição para OPA")
				span.RecordError(err)
				logger.Error().Err(err).Msg("Falha ao criar requisição para OPA")
				handleAuthError(w, http.StatusInternalServerError, "authz_error", "Erro interno de autorização", logger)
				return
			}
			
			// Configurar headers
			req.Header.Set("Content-Type", "application/json")
			
			// Enviar requisição para o OPA
			httpClient := &http.Client{
				Timeout: config.Timeout,
			}
			
			span.AddEvent("opa.request.start")
			startTime := time.Now()
			
			resp, err := httpClient.Do(req)
			
			latency := time.Since(startTime)
			span.SetAttributes(attribute.String("opa.latency", latency.String()))
			span.AddEvent("opa.request.end")
			
			// Verificar erro de comunicação com o OPA
			if err != nil {
				span.SetStatus(codes.Error, "Erro ao comunicar com o OPA")
				span.RecordError(err)
				logger.Error().Err(err).Msg("Falha ao comunicar com o OPA")
				handleAuthError(w, http.StatusInternalServerError, "authz_error", "Erro interno de autorização", logger)
				return
			}
			defer resp.Body.Close()
			
			// Verificar código de status da resposta do OPA
			if resp.StatusCode != http.StatusOK {
				span.SetStatus(codes.Error, fmt.Sprintf("OPA retornou status %d", resp.StatusCode))
				logger.Error().
					Int("status", resp.StatusCode).
					Msg("OPA retornou status não-OK")
				handleAuthError(w, http.StatusInternalServerError, "authz_error", "Erro interno de autorização", logger)
				return
			}
			
			// Ler e processar resposta do OPA
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				span.SetStatus(codes.Error, "Erro ao ler resposta do OPA")
				span.RecordError(err)
				logger.Error().Err(err).Msg("Falha ao ler resposta do OPA")
				handleAuthError(w, http.StatusInternalServerError, "authz_error", "Erro interno de autorização", logger)
				return
			}
			
			// Parse da resposta
			var opaResp map[string]interface{}
			if err := json.Unmarshal(respBody, &opaResp); err != nil {
				span.SetStatus(codes.Error, "Erro ao deserializar resposta do OPA")
				span.RecordError(err)
				logger.Error().Err(err).Msg("Falha ao deserializar resposta do OPA")
				handleAuthError(w, http.StatusInternalServerError, "authz_error", "Erro interno de autorização", logger)
				return
			}
			
			// Verificar decisão do OPA
			result, exists := opaResp["result"]
			if !exists {
				span.SetStatus(codes.Error, "Resposta do OPA não contém campo 'result'")
				logger.Error().Msg("Resposta do OPA não contém campo 'result'")
				handleAuthError(w, http.StatusInternalServerError, "authz_error", "Erro interno de autorização", logger)
				return
			}
			
			// Navegar pelo caminho da decisão
			decision := result
			if config.DecisionPath != "" {
				for _, part := range strings.Split(config.DecisionPath, ".") {
					if m, ok := decision.(map[string]interface{}); ok {
						if v, exists := m[part]; exists {
							decision = v
						} else {
							span.SetStatus(codes.Error, "Caminho da decisão não encontrado na resposta do OPA")
							logger.Error().
								Str("decision_path", config.DecisionPath).
								Msg("Caminho da decisão não encontrado na resposta do OPA")
							handleAuthError(w, http.StatusInternalServerError, "authz_error", "Erro interno de autorização", logger)
							return
						}
					} else {
						span.SetStatus(codes.Error, "Formato inválido para navegação no caminho da decisão")
						logger.Error().
							Str("decision_path", config.DecisionPath).
							Msg("Formato inválido para navegação no caminho da decisão")
						handleAuthError(w, http.StatusInternalServerError, "authz_error", "Erro interno de autorização", logger)
						return
					}
				}
			}
			
			// Verificar se o acesso foi permitido
			allowed, ok := decision.(bool)
			if !ok {
				span.SetStatus(codes.Error, "Decisão do OPA não é um booleano")
				logger.Error().
					Interface("decision", decision).
					Msg("Decisão do OPA não é um booleano")
				handleAuthError(w, http.StatusInternalServerError, "authz_error", "Erro interno de autorização", logger)
				return
			}
			
			// Registrar resultado da autorização
			span.SetAttributes(attribute.Bool("authorization.allowed", allowed))
			logger.Debug().
				Bool("allowed", allowed).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("user_id", userID.String()).
				Str("tenant_id", tenantID.String()).
				Msg("Resultado da autorização")
			
			if !allowed {
				span.SetStatus(codes.Unauthenticated, "Acesso negado")
				handleAuthError(w, http.StatusForbidden, "access_denied", "Acesso negado. Você não tem permissão para realizar esta operação.", logger)
				return
			}
			
			// Continuar com a requisição
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// parseQueryParams extrai os parâmetros de consulta da requisição
func parseQueryParams(r *http.Request) map[string]interface{} {
	queryParams := make(map[string]interface{})
	for k, v := range r.URL.Query() {
		if len(v) == 1 {
			queryParams[k] = v[0]
		} else {
			queryParams[k] = v
		}
	}
	return queryParams
}

// sanitizeMap cria uma cópia segura de um mapa para logging,
// removendo ou ofuscando informações sensíveis
func sanitizeMap(input map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	for k, v := range input {
		switch val := v.(type) {
		case map[string]interface{}:
			result[k] = sanitizeMap(val)
		case []interface{}:
			sanitized := make([]interface{}, 0, len(val))
			for _, item := range val {
				if itemMap, ok := item.(map[string]interface{}); ok {
					sanitized = append(sanitized, sanitizeMap(itemMap))
				} else {
					sanitized = append(sanitized, item)
				}
			}
			result[k] = sanitized
		case string:
			// Sanitizar valores sensíveis
			if isSensitiveKey(k) {
				result[k] = maskSensitiveValue(val)
			} else {
				result[k] = val
			}
		default:
			result[k] = val
		}
	}
	
	return result
}

// isSensitiveKey verifica se a chave representa um campo sensível
func isSensitiveKey(key string) bool {
	sensitivePrefixes := []string{
		"password", "token", "secret", "credential", "api_key", "private",
		"auth", "key", "cert", "sign", "hash", "cipher", "crypt",
	}
	
	loweredKey := strings.ToLower(key)
	for _, prefix := range sensitivePrefixes {
		if strings.Contains(loweredKey, prefix) {
			return true
		}
	}
	
	return false
}

// maskSensitiveValue ofusca um valor sensível
func maskSensitiveValue(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	
	// Verificar se é um UUID válido
	if _, err := uuid.Parse(value); err == nil {
		// Mostrar apenas primeiros e últimos 4 caracteres de UUIDs
		return value[:4] + "..." + value[len(value)-4:]
	}
	
	// Para outros valores, ofuscar a maior parte
	visibleChars := 4
	if len(value) > 16 {
		visibleChars = len(value) / 4
	}
	
	return value[:visibleChars] + strings.Repeat("*", len(value)-visibleChars*2) + value[len(value)-visibleChars:]
}