// Middleware de Integração OPA para RoleService da Plataforma INNOVABIZ
// Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
// PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III, PSD2, AML/KYC
package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/innovabiz/iam/logging"
	"github.com/innovabiz/iam/models"
)

// Configuração do middleware OPA
type OPAConfig struct {
	// Endereço do serviço OPA
	OPAEndpoint string `json:"opa_endpoint"`
	
	// Timeout para requisições ao OPA em segundos
	TimeoutSeconds int `json:"timeout_seconds"`
	
	// Habilita cache de decisões
	EnableCache bool `json:"enable_cache"`
	
	// TTL do cache em segundos
	CacheTTL int `json:"cache_ttl"`
	
	// Modo de falha (permitir ou negar)
	FailOpen bool `json:"fail_open"`
	
	// Habilita logs detalhados de decisões
	VerboseLogging bool `json:"verbose_logging"`
}

// Middleware OPA para autorização
type OPAMiddleware struct {
	config OPAConfig
	logger logging.Logger
	cache  DecisionCache
}

// Interface para cache de decisões
type DecisionCache interface {
	Get(key string) (bool, bool)
	Set(key string, decision bool, ttl time.Duration)
	Clear()
}

// Contexto da requisição para autorização
type AuthorizationContext struct {
	ClientIP   string `json:"client_ip"`
	UserAgent  string `json:"user_agent"`
	RequestID  string `json:"request_id"`
	Origin     string `json:"origin,omitempty"`
	SessionID  string `json:"session_id,omitempty"`
	DeviceInfo string `json:"device_info,omitempty"`
}

// Input para decisão OPA
type OPAInput struct {
	HTTPMethod string                 `json:"http_method"`
	TenantID   string                 `json:"tenant_id"`
	User       models.AuthenticatedUser `json:"user"`
	Resource   map[string]interface{} `json:"resource"`
	Context    AuthorizationContext   `json:"context"`
	Path       string                 `json:"path"`
}

// Resposta do OPA
type OPADecision struct {
	Result bool `json:"result"`
}

// Criar nova instância do middleware OPA
func NewOPAMiddleware(config OPAConfig, logger logging.Logger, cache DecisionCache) *OPAMiddleware {
	if cache == nil && config.EnableCache {
		cache = NewInMemoryCache()
	}
	
	return &OPAMiddleware{
		config: config,
		logger: logger,
		cache:  cache,
	}
}

// Middleware para integração com OPA
func (m *OPAMiddleware) Authorize(decisionPath string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extrair usuário do contexto (presumindo que foi definido por um middleware de autenticação anterior)
			user, ok := r.Context().Value("authenticatedUser").(models.AuthenticatedUser)
			if !ok {
				m.logger.Error("user not found in request context")
				http.Error(w, "Unauthorized: User not authenticated", http.StatusUnauthorized)
				return
			}
			
			// Obter dados específicos do recurso com base no caminho da API
			resource, err := extractResourceFromRequest(r)
			if err != nil {
				m.logger.Error("failed to extract resource data", "error", err.Error())
				http.Error(w, "Bad Request: Invalid resource data", http.StatusBadRequest)
				return
			}
			
			// Obter tenant ID do cabeçalho ou contexto
			tenantID := r.Header.Get("X-Tenant-ID")
			if tenantID == "" {
				tenantID = user.TenantID
			}
			
			// Construir contexto de autorização
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
				r.Header.Set("X-Request-ID", requestID)
			}
			
			authCtx := AuthorizationContext{
				ClientIP:   getClientIP(r),
				UserAgent:  r.UserAgent(),
				RequestID:  requestID,
				Origin:     r.Header.Get("Origin"),
				SessionID:  r.Header.Get("X-Session-ID"),
				DeviceInfo: r.Header.Get("X-Device-Info"),
			}
			
			// Construir input para OPA
			input := OPAInput{
				HTTPMethod: r.Method,
				TenantID:   tenantID,
				User:       user,
				Resource:   resource,
				Context:    authCtx,
				Path:       decisionPath,
			}
			
			// Verificar cache para decisões repetidas
			if m.config.EnableCache {
				cacheKey := generateCacheKey(input)
				if decision, found := m.cache.Get(cacheKey); found {
					if decision {
						next.ServeHTTP(w, r)
					} else {
						http.Error(w, "Forbidden: Access denied by policy", http.StatusForbidden)
					}
					return
				}
			}
			
			// Decisão de autorização via OPA
			allowed, err := m.checkOPADecision(r.Context(), input)
			
			// Tratar erros de comunicação com OPA
			if err != nil {
				m.logger.Error("opa authorization error", "error", err.Error())
				
				// Decidir com base no modo de falha configurado
				if m.config.FailOpen {
					m.logger.Warn("failing open due to OPA error", "path", r.URL.Path)
					next.ServeHTTP(w, r)
					return
				}
				
				http.Error(w, "Internal Server Error: Authorization service unavailable", http.StatusInternalServerError)
				return
			}
			
			// Cache da decisão se habilitado
			if m.config.EnableCache {
				cacheKey := generateCacheKey(input)
				m.cache.Set(cacheKey, allowed, time.Duration(m.config.CacheTTL)*time.Second)
			}
			
			// Aplicar decisão de autorização
			if allowed {
				// Adicionar contexto de auditoria para o handler subsequente
				ctx := context.WithValue(r.Context(), "auditInfo", map[string]interface{}{
					"decision_path": decisionPath,
					"authorized_at": time.Now().UTC(),
					"request_id":    requestID,
				})
				
				// Prosseguir com a requisição autorizada
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				// Registrar negação de acesso
				m.logger.Warn("access denied",
					"user_id", user.ID,
					"tenant_id", tenantID,
					"path", r.URL.Path,
					"method", r.Method,
					"decision_path", decisionPath,
				)
				
				http.Error(w, "Forbidden: Access denied by policy", http.StatusForbidden)
			}
		})
	}
}

// Consulta o serviço OPA para decisão de autorização
func (m *OPAMiddleware) checkOPADecision(ctx context.Context, input OPAInput) (bool, error) {
	// Criar contexto com timeout
	timeoutDuration := time.Duration(m.config.TimeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeoutDuration)
	defer cancel()
	
	// Serializar input
	inputJSON, err := json.Marshal(map[string]interface{}{
		"input": input,
	})
	if err != nil {
		return false, fmt.Errorf("failed to marshal OPA input: %w", err)
	}
	
	// Log detalhado de decisões se habilitado
	if m.config.VerboseLogging {
		m.logger.Debug("opa authorization request", "input", string(inputJSON))
	}
	
	// Criar requisição para OPA
	endpoint := fmt.Sprintf("%s/v1/data/innovabiz/iam/role/allow", m.config.OPAEndpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(inputJSON))
	if err != nil {
		return false, fmt.Errorf("failed to create OPA request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	// Enviar requisição para OPA
	client := &http.Client{Timeout: timeoutDuration}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to call OPA: %w", err)
	}
	defer resp.Body.Close()
	
	// Verificar código de resposta
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("OPA returned non-200 status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}
	
	// Processar resposta
	var decision struct {
		Result OPADecision `json:"result"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&decision); err != nil {
		return false, fmt.Errorf("failed to decode OPA response: %w", err)
	}
	
	// Log detalhado da decisão se habilitado
	if m.config.VerboseLogging {
		m.logger.Debug("opa authorization decision", "allowed", decision.Result.Result)
	}
	
	return decision.Result.Result, nil
}

// Extrai dados relevantes do recurso a partir da requisição HTTP
func extractResourceFromRequest(r *http.Request) (map[string]interface{}, error) {
	resource := map[string]interface{}{
		"path": r.URL.Path,
	}
	
	// Extrair parâmetros da URL
	pathParts := strings.Split(r.URL.Path, "/")
	
	// Análise por tipo de recurso (exemplos)
	if len(pathParts) >= 3 && pathParts[1] == "roles" {
		// Ex: /roles/{id}
		if len(pathParts) >= 3 && pathParts[2] != "" {
			resource["id"] = pathParts[2]
		}
		
		// Ex: /roles/{id}/permissions/{permission_id}
		if len(pathParts) >= 5 && pathParts[3] == "permissions" {
			resource["permission_id"] = pathParts[4]
		}
	} else if len(pathParts) >= 4 && pathParts[1] == "users" && pathParts[3] == "roles" {
		// Ex: /users/{user_id}/roles/{role_id}
		resource["user_id"] = pathParts[2]
		if len(pathParts) >= 5 && pathParts[4] != "" {
			resource["role_id"] = pathParts[4]
		}
	}
	
	// Extrair dados do corpo para métodos que enviam dados
	if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
		if r.Body != nil && r.Header.Get("Content-Type") == "application/json" {
			var bodyData map[string]interface{}
			
			// Preservar o body para leitura posterior
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				return nil, err
			}
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			
			// Deserializar body
			if err := json.Unmarshal(bodyBytes, &bodyData); err != nil {
				return nil, err
			}
			
			resource["data"] = bodyData
		}
	}
	
	// Processar parâmetros de query
	queryParams := make(map[string]interface{})
	for k, v := range r.URL.Query() {
		if len(v) == 1 {
			queryParams[k] = v[0]
		} else {
			queryParams[k] = v
		}
	}
	
	if len(queryParams) > 0 {
		resource["query_params"] = queryParams
	}
	
	return resource, nil
}

// Obter endereço IP do cliente com suporte a proxies
func getClientIP(r *http.Request) string {
	// Verificar cabeçalhos X-Forwarded-For, X-Real-IP em ordem
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		if ip := r.Header.Get(header); ip != "" {
			// Para X-Forwarded-For, pegar o primeiro IP (cliente original)
			if header == "X-Forwarded-For" {
				ips := strings.Split(ip, ",")
				if len(ips) > 0 {
					return strings.TrimSpace(ips[0])
				}
			}
			return ip
		}
	}
	
	// Extrair IP do endereço remoto
	ip := r.RemoteAddr
	// Remover a porta se presente
	if i := strings.LastIndex(ip, ":"); i != -1 {
		ip = ip[:i]
	}
	return ip
}

// Gerar chave para cache com base no input OPA
func generateCacheKey(input OPAInput) string {
	// Simplificação: na implementação real, usar hash criptográfico
	key := fmt.Sprintf("%s:%s:%s:%s:%s",
		input.HTTPMethod,
		input.TenantID,
		input.User.ID,
		input.Path,
		fmt.Sprintf("%v", input.Resource["path"]),
	)
	return key
}

// Implementação simples de cache em memória
type InMemoryCache struct {
	items map[string]cacheItem
}

type cacheItem struct {
	decision  bool
	expiresAt time.Time
}

// Criar nova instância de cache em memória
func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		items: make(map[string]cacheItem),
	}
}

// Obter decisão do cache
func (c *InMemoryCache) Get(key string) (bool, bool) {
	item, exists := c.items[key]
	if !exists {
		return false, false
	}
	
	// Verificar expiração
	if time.Now().After(item.expiresAt) {
		delete(c.items, key)
		return false, false
	}
	
	return item.decision, true
}

// Armazenar decisão no cache
func (c *InMemoryCache) Set(key string, decision bool, ttl time.Duration) {
	c.items[key] = cacheItem{
		decision:  decision,
		expiresAt: time.Now().Add(ttl),
	}
}

// Limpar cache
func (c *InMemoryCache) Clear() {
	c.items = make(map[string]cacheItem)
}