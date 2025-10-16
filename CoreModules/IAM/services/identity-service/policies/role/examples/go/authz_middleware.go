// Package authz fornece integração com políticas OPA para autorização
//
// Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
// PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III
package authz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/innovabiz/iam/internal/domain/model"
	"github.com/innovabiz/iam/internal/infrastructure/observability"
	"github.com/innovabiz/iam/internal/infrastructure/telemetry"
)

// Config define as configurações para o middleware de autorização
type Config struct {
	OPAServerURL      string
	PolicyPackage     string
	DefaultTenant     string
	Timeout           time.Duration
	EnableCache       bool
	CacheTTL          time.Duration
	EnableFallback    bool
	AuditService      AuditService
	ObservabilityOpts []telemetry.Option
}

// AuditService interface para serviço de auditoria
type AuditService interface {
	RecordEvent(ctx context.Context, event model.AuditEvent) error
}

// DefaultConfig retorna configurações padrão para o middleware
func DefaultConfig() Config {
	return Config{
		OPAServerURL:   "http://localhost:8181",
		PolicyPackage:  "innovabiz.iam.role",
		DefaultTenant:  "00000000-0000-0000-0000-000000000000",
		Timeout:        500 * time.Millisecond,
		EnableCache:    true,
		CacheTTL:       5 * time.Minute,
		EnableFallback: true,
	}
}

// Middleware cria um middleware Gin para autorização baseada em políticas OPA
func Middleware(config Config, policy string) gin.HandlerFunc {
	logger := observability.GetLogger(context.Background())
	meter := telemetry.NewMeter("authz.middleware", config.ObservabilityOpts...)

	logger.Info("inicializando middleware de autorização OPA",
		"policy", policy,
		"opa_server", config.OPAServerURL,
		"timeout", config.Timeout,
		"cache_enabled", config.EnableCache)

	return func(c *gin.Context) {
		start := time.Now()
		ctx := c.Request.Context()
		reqLogger := logger.With("request_id", getOrCreateRequestID(c))

		reqLogger.Debug("processando autorização", 
			"path", c.Request.URL.Path, 
			"method", c.Request.Method)

		// 1. Extrair informações do usuário autenticado
		userClaims, exists := c.Get("userClaims")
		if !exists {
			reqLogger.Warn("usuário não autenticado")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "usuário não autenticado",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		// 2. Extrair tenantID do header ou definir padrão
		tenantID := c.GetHeader("X-Tenant-ID")
		if tenantID == "" {
			tenantID = config.DefaultTenant
		}

		// 3. Obter dados do recurso (definido pelos controllers)
		resourceData, exists := c.Get("resourceData")
		if !exists {
			resourceData = map[string]interface{}{}
		}

		// 4. Preparar input para OPA
		input := map[string]interface{}{
			"user": userClaims,
			"resource": map[string]interface{}{
				"tenant_id": tenantID,
				"data":      resourceData,
			},
			"method":  c.Request.Method,
			"context": buildRequestContext(c),
		}

		// 5. Registrar decisão de autorização
		defer func() {
			duration := time.Since(start)
			meter.RecordAuthorizationLatency(duration, policy)
			reqLogger.Debug("autorização processada", 
				"policy", policy, 
				"duration_ms", duration.Milliseconds())
		}()

		// 6. Consultar OPA para decisão
		decision, err := queryOPA(ctx, config, policy, input)
		if err != nil {
			reqLogger.Error("falha ao consultar OPA", "error", err)
			
			// 6.1. Usar política de fallback se habilitado
			if config.EnableFallback {
				reqLogger.Info("aplicando política de fallback")
				if fallbackDecision := applyFallbackPolicy(policy, input); fallbackDecision.Allow {
					c.Next()
					return
				}
			}
			
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "falha no serviço de autorização",
				"code":  "AUTHZ_ERROR",
			})
			meter.IncrementAuthorizationErrors(policy, "service_error")
			return
		}

		// 7. Processar decisão
		if !decision.Allow {
			// 7.1. Registrar tentativa de acesso não autorizada
			auditUnauthorizedAccess(ctx, config, c, input, decision)
			
			// 7.2. Incrementar métricas
			meter.IncrementAuthorizationDenials(policy, decision.Reason)
			
			// 7.3. Responder com acesso negado
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":  decision.Reason,
				"code":   "FORBIDDEN",
				"policy": policy,
			})
			return
		}

		// 8. Acesso permitido
		meter.IncrementAuthorizationGrants(policy)
		c.Next()
	}
}

// Decision representa uma decisão de autorização OPA
type Decision struct {
	Allow  bool   `json:"allow"`
	Reason string `json:"reason,omitempty"`
}

// queryOPA envia uma consulta ao servidor OPA e retorna a decisão
func queryOPA(ctx context.Context, config Config, policy string, input interface{}) (Decision, error) {
	logger := observability.GetLogger(ctx)
	
	// 1. Verificar cache se habilitado
	if config.EnableCache {
		if cachedDecision, found := checkCache(policy, input); found {
			logger.Debug("decisão encontrada em cache", "policy", policy)
			return cachedDecision, nil
		}
	}

	// 2. Preparar URL para consulta OPA
	policyPath := fmt.Sprintf("/v1/data/%s/%s", config.PolicyPackage, policy)
	url := fmt.Sprintf("%s%s", config.OPAServerURL, policyPath)

	// 3. Preparar payload para OPA
	opaInput := map[string]interface{}{
		"input": input,
	}
	requestBody, err := json.Marshal(opaInput)
	if err != nil {
		return Decision{}, fmt.Errorf("falha ao serializar input: %w", err)
	}

	// 4. Criar contexto com timeout
	ctx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	// 5. Criar e enviar requisição HTTP
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return Decision{}, fmt.Errorf("falha ao criar requisição: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 6. Executar requisição
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Decision{}, fmt.Errorf("falha ao executar requisição: %w", err)
	}
	defer resp.Body.Close()

	// 7. Verificar código de resposta
	if resp.StatusCode != http.StatusOK {
		return Decision{}, fmt.Errorf("OPA retornou status inesperado: %d", resp.StatusCode)
	}

	// 8. Decodificar resposta
	var result struct {
		Result Decision `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Decision{}, fmt.Errorf("falha ao decodificar resposta: %w", err)
	}

	// 9. Atualizar cache se habilitado
	if config.EnableCache {
		updateCache(policy, input, result.Result, config.CacheTTL)
	}

	return result.Result, nil
}

// buildRequestContext constrói o contexto da requisição para avaliação da política
func buildRequestContext(c *gin.Context) map[string]interface{} {
	return map[string]interface{}{
		"timestamp":          time.Now().UTC().Format(time.RFC3339),
		"client_ip":          c.ClientIP(),
		"user_agent":         c.Request.UserAgent(),
		"request_id":         getOrCreateRequestID(c),
		"trace_id":           c.GetString("X-Trace-ID"),
		"geo":                getGeoInfo(c.ClientIP()),
		"requests_per_minute": getRateLimit(c),
	}
}

// getOrCreateRequestID obtém ou cria um ID de requisição
func getOrCreateRequestID(c *gin.Context) string {
	if reqID := c.GetString("X-Request-ID"); reqID != "" {
		return reqID
	}
	
	reqID := uuid.New().String()
	c.Set("X-Request-ID", reqID)
	c.Header("X-Request-ID", reqID)
	
	return reqID
}

// getGeoInfo obtém informações geográficas baseadas no IP (mock para exemplo)
func getGeoInfo(clientIP string) map[string]string {
	// Em um cenário real, usaria um serviço de geolocalização
	return map[string]string{
		"country": "AO",
		"region":  "Luanda",
	}
}

// getRateLimit obtém informações de limite de taxa para o cliente (mock para exemplo)
func getRateLimit(c *gin.Context) int {
	// Em um cenário real, consultaria um serviço de rate limiting
	return 5
}

// auditUnauthorizedAccess registra tentativas de acesso não autorizadas
func auditUnauthorizedAccess(ctx context.Context, config Config, c *gin.Context, input interface{}, decision Decision) {
	logger := observability.GetLogger(ctx)
	
	if config.AuditService == nil {
		logger.Warn("serviço de auditoria não configurado")
		return
	}
	
	// Criar evento de auditoria para acesso negado
	auditEvent := model.AuditEvent{
		EventType:    "AUTHORIZATION_DENIED",
		ResourceType: "role",
		Action:       c.Request.Method,
		ActorID:      c.GetString("user_id"),
		TenantID:     c.GetHeader("X-Tenant-ID"),
		Timestamp:    time.Now().UTC(),
		RequestID:    getOrCreateRequestID(c),
		ClientIP:     c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
		Details: map[string]interface{}{
			"policy":        c.FullPath(),
			"reason":        decision.Reason,
			"input":         input,
			"request_path":  c.Request.URL.Path,
			"request_query": c.Request.URL.RawQuery,
		},
		Severity: "WARNING",
	}
	
	// Registrar evento de auditoria de forma assíncrona
	go func() {
		if err := config.AuditService.RecordEvent(context.Background(), auditEvent); err != nil {
			logger.Error("falha ao registrar evento de auditoria", "error", err)
		}
	}()
}

// applyFallbackPolicy aplica uma política de fallback local em caso de falha do OPA
func applyFallbackPolicy(policy string, input interface{}) Decision {
	// Políticas de fallback devem ser muito restritas
	// e permitir apenas operações de leitura para usuários administrativos
	
	userMap, ok := input.(map[string]interface{})["user"].(map[string]interface{})
	if !ok {
		return Decision{Allow: false, Reason: "fallback: usuário inválido"}
	}
	
	method, ok := input.(map[string]interface{})["method"].(string)
	if !ok {
		return Decision{Allow: false, Reason: "fallback: método inválido"}
	}
	
	// Permitir apenas operações de leitura (GET)
	if method != "GET" {
		return Decision{Allow: false, Reason: "fallback: apenas operações de leitura permitidas"}
	}
	
	// Verificar se o usuário tem papel administrativo
	roles, ok := userMap["roles"].([]interface{})
	if !ok {
		return Decision{Allow: false, Reason: "fallback: papéis não encontrados"}
	}
	
	// Verificar se usuário é admin
	isAdmin := false
	for _, role := range roles {
		roleStr, ok := role.(string)
		if ok && (roleStr == "SUPER_ADMIN" || roleStr == "TENANT_ADMIN" || roleStr == "IAM_ADMIN") {
			isAdmin = true
			break
		}
	}
	
	if !isAdmin {
		return Decision{Allow: false, Reason: "fallback: acesso restrito a administradores"}
	}
	
	return Decision{Allow: true, Reason: "fallback: acesso de leitura permitido para administrador"}
}

// Funções de cache (implementação simplificada)
// Em um cenário real, usaria um sistema de cache distribuído como Redis

var decisionCache = map[string]cachedItem{}

type cachedItem struct {
	decision  Decision
	expiresAt time.Time
}

func cacheKey(policy string, input interface{}) string {
	inputJSON, _ := json.Marshal(input)
	return fmt.Sprintf("%s:%s", policy, string(inputJSON))
}

func checkCache(policy string, input interface{}) (Decision, bool) {
	key := cacheKey(policy, input)
	if item, found := decisionCache[key]; found && time.Now().Before(item.expiresAt) {
		return item.decision, true
	}
	return Decision{}, false
}

func updateCache(policy string, input interface{}, decision Decision, ttl time.Duration) {
	key := cacheKey(policy, input)
	decisionCache[key] = cachedItem{
		decision:  decision,
		expiresAt: time.Now().Add(ttl),
	}
}