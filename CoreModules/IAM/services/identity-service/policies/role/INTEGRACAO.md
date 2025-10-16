# Guia de Integração das Políticas OPA do RoleService

## Visão Geral

Este documento descreve como integrar as políticas de autorização OPA (Open Policy Agent) do módulo RoleService ao middleware de autorização da plataforma INNOVABIZ. A arquitetura de autorização utiliza uma abordagem baseada em políticas descentralizadas e escaláveis, com isolamento total entre tenants e alta conformidade com padrões internacionais de segurança.

| Metadata | Valor |
|----------|-------|
| Versão | 1.0.0 |
| Status | Homologação |
| Classificação | Restrito |
| Data Criação | 2025-08-05 |
| Última Atualização | 2025-08-05 |
| Autor | INNOVABIZ IAM Team |
| Aprovado por | Eduardo Jeremias |

## Arquitetura de Autorização

A arquitetura de autorização da plataforma INNOVABIZ segue o modelo **Policy Enforcement Point (PEP) / Policy Decision Point (PDP)**, implementado da seguinte forma:

1. **Policy Enforcement Point (PEP)**: Middleware de autorização Go implementado no servidor HTTP
2. **Policy Decision Point (PDP)**: Servidor OPA distribuído com políticas sincronizadas
3. **Policy Information Point (PIP)**: Camada de acesso a dados para consulta de contexto
4. **Policy Administration Point (PAP)**: Interface de administração para gestão de políticas
5. **Policy Retrieval Point (PRP)**: Repositório Git e CI/CD para versionamento e distribuição de políticas

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│ HTTP Router │──1──>│  Auth       │──2──>│  OPA        │
│  (Gin)      │      │  Middleware │      │  Server     │
└─────────────┘      └─────────────┘      └─────────────┘
                           │                    │
                           │                    │
                          3│                   5│
                           ▼                    ▼
                     ┌─────────────┐      ┌─────────────┐
                     │  Data       │      │  Policy     │
                     │  Services   │      │  Repository │
                     └─────────────┘      └─────────────┘
                           ▲                    ▲
                           │                    │
                          4│                   6│
                           │                    │
                     ┌─────────────┐      ┌─────────────┐
                     │  Database   │      │  CI/CD      │
                     │             │      │  Pipeline   │
                     └─────────────┘      └─────────────┘
```

## Componentes de Integração

### 1. Middleware de Autorização (Go)

O middleware de autorização deve ser implementado como parte dos handlers HTTP do módulo RoleService:

```go
// authz/middleware.go
package authz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/innovabiz/iam/internal/domain/model"
	"github.com/innovabiz/iam/internal/observability"
)

// AuthzMiddleware integra as políticas OPA com o framework Gin
func AuthzMiddleware(policy string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		logger := observability.GetLogger(c.Request.Context())
		
		// 1. Extrair informações do usuário autenticado (do JWT)
		userClaims, exists := c.Get("userClaims")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "usuário não autenticado",
				"code":  "UNAUTHORIZED",
			})
			return
		}
		
		// 2. Preparar input para OPA
		input := map[string]interface{}{
			"user": userClaims,
			"resource": map[string]interface{}{
				"tenant_id": c.GetHeader("X-Tenant-ID"),
				"data":      c.MustGet("resourceData"),
			},
			"method":  c.Request.Method,
			"context": buildRequestContext(c),
		}
		
		// 3. Consultar OPA para decisão de autorização
		decision, err := queryOPA(c.Request.Context(), policy, input)
		if err != nil {
			logger.Error("falha ao consultar OPA", "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "falha no serviço de autorização",
				"code":  "AUTHZ_ERROR",
			})
			return
		}
		
		// 4. Processar decisão
		allow, ok := decision["allow"].(bool)
		if !ok || !allow {
			reason := "acesso negado"
			if r, exists := decision["reason"].(string); exists {
				reason = r
			}
			
			// Registrar tentativa de acesso não autorizado
			auditUnauthorizedAccess(c, input, reason)
			
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":  reason,
				"code":   "FORBIDDEN",
				"policy": policy,
			})
			return
		}
		
		// 5. Registrar métricas
		observability.RecordAuthzDecision(policy, true, time.Since(start))
		
		// Continuar para o próximo handler
		c.Next()
	}
}

// queryOPA envia a consulta ao servidor OPA e retorna a decisão
func queryOPA(ctx context.Context, policy string, input interface{}) (map[string]interface{}, error) {
	logger := observability.GetLogger(ctx)
	
	// 1. Serializar input
	inputJSON, err := json.Marshal(input)
	if err != nil {
		logger.Error("erro ao serializar input para OPA", "error", err)
		return nil, err
	}
	
	// 2. Montar URL para a política específica
	policyPath := fmt.Sprintf("/v1/data/innovabiz/iam/role/%s", policy)
	url := fmt.Sprintf("%s%s", getOPAServerURL(), policyPath)
	
	// 3. Preparar requisição
	opaInput := map[string]interface{}{
		"input": input,
	}
	requestBody, err := json.Marshal(opaInput)
	if err != nil {
		return nil, err
	}
	
	// 4. Enviar requisição para OPA
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	
	// 5. Executar requisição com timeout
	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		// Fallback para política local em caso de falha
		return localFallbackPolicy(policy, input)
	}
	defer resp.Body.Close()
	
	// 6. Processar resposta
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OPA returned status: %d", resp.StatusCode)
	}
	
	var result struct {
		Result map[string]interface{} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result.Result, nil
}

// buildRequestContext constrói o contexto da requisição para avaliação de política
func buildRequestContext(c *gin.Context) map[string]interface{} {
	return map[string]interface{}{
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"client_ip":    c.ClientIP(),
		"user_agent":   c.Request.UserAgent(),
		"request_id":   c.GetString("X-Request-ID"),
		"trace_id":     c.GetString("X-Trace-ID"),
		"geo":          getGeoInfo(c.ClientIP()),
		"risk_factors": getRiskFactors(c),
	}
}

// auditUnauthorizedAccess registra tentativas de acesso não autorizadas
func auditUnauthorizedAccess(c *gin.Context, input interface{}, reason string) {
	logger := observability.GetLogger(c.Request.Context())
	auditEvent := model.AuditEvent{
		EventType:    "AUTHORIZATION_DENIED",
		ResourceType: "role",
		Action:       c.Request.Method,
		ActorID:      c.GetString("user_id"),
		TenantID:     c.GetHeader("X-Tenant-ID"),
		Timestamp:    time.Now().UTC(),
		RequestID:    c.GetString("X-Request-ID"),
		ClientIP:     c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
		Details: map[string]interface{}{
			"policy":        c.FullPath(),
			"reason":        reason,
			"input":         input,
			"request_path":  c.Request.URL.Path,
			"request_query": c.Request.URL.RawQuery,
		},
		Severity: "WARNING",
	}
	
	// Registrar evento de auditoria de forma assíncrona
	go func() {
		if err := services.AuditService.RecordEvent(context.Background(), auditEvent); err != nil {
			logger.Error("falha ao registrar evento de auditoria", "error", err)
		}
	}()
}
```

### 2. Handlers do RoleService

Os handlers HTTP do RoleService devem utilizar o middleware de autorização:

```go
// handler/role_handler.go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/innovabiz/iam/internal/application/service"
	"github.com/innovabiz/iam/internal/domain/model"
	"github.com/innovabiz/iam/internal/interface/api/authz"
)

type RoleHandler struct {
	roleService service.RoleService
}

func NewRoleHandler(roleService service.RoleService) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

func (h *RoleHandler) RegisterRoutes(router *gin.RouterGroup) {
	roles := router.Group("/roles")
	
	// Rotas para gestão de funções
	roles.POST("", authz.AuthzMiddleware("crud.create_decision"), h.CreateRole)
	roles.GET("", authz.AuthzMiddleware("crud.list_decision"), h.ListRoles)
	roles.GET("/:id", authz.AuthzMiddleware("crud.read_decision"), h.GetRole)
	roles.PUT("/:id", authz.AuthzMiddleware("crud.update_decision"), h.UpdateRole)
	roles.DELETE("/:id", authz.AuthzMiddleware("crud.delete_decision"), h.DeleteRole)
	roles.DELETE("/:id/permanent", authz.AuthzMiddleware("crud.permanent_delete_decision"), h.PermanentDeleteRole)
	
	// Rotas para gestão de permissões
	roles.POST("/:id/permissions", authz.AuthzMiddleware("permissions.permission_assignment_decision"), h.AssignPermission)
	roles.DELETE("/:id/permissions/:permission_id", authz.AuthzMiddleware("permissions.permission_revocation_decision"), h.RevokePermission)
	roles.GET("/:id/permissions", authz.AuthzMiddleware("permissions.permission_check_decision"), h.ListPermissions)
	
	// Rotas para gestão de hierarquia
	roles.POST("/hierarchy", authz.AuthzMiddleware("hierarchy.hierarchy_addition_decision"), h.AddHierarchy)
	roles.DELETE("/hierarchy/:id", authz.AuthzMiddleware("hierarchy.hierarchy_removal_decision"), h.RemoveHierarchy)
	roles.GET("/:id/hierarchy", authz.AuthzMiddleware("hierarchy.hierarchy_query_decision"), h.GetHierarchy)
	
	// Rotas para atribuição de função a usuário
	roles.POST("/assignments", authz.AuthzMiddleware("user_assignment.role_assignment_decision"), h.AssignRoleToUser)
	roles.DELETE("/assignments/:id", authz.AuthzMiddleware("user_assignment.role_removal_decision"), h.RemoveRoleFromUser)
	roles.PUT("/assignments/:id/expiration", authz.AuthzMiddleware("user_assignment.expiration_update_decision"), h.UpdateAssignmentExpiration)
	roles.GET("/users/:user_id", authz.AuthzMiddleware("user_assignment.role_check_decision"), h.GetUserRoles)
}

func (h *RoleHandler) CreateRole(c *gin.Context) {
	var createRoleDTO dto.CreateRoleDTO
	if err := c.ShouldBindJSON(&createRoleDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Disponibilizar dados do recurso para o middleware de autorização
	c.Set("resourceData", createRoleDTO)
	
	// Implementar lógica de criação de função...
}

// Demais métodos handler implementados similarmente...
```

### 3. Distribuição de Políticas OPA

As políticas OPA devem ser distribuídas para todos os servidores OPA através do processo de CI/CD:

```yaml
# .github/workflows/deploy-opa-policies.yml
name: Deploy OPA Policies

on:
  push:
    branches: [ main, develop ]
    paths:
      - 'CoreModules/IAM/services/identity-service/policies/**'

jobs:
  deploy-policies:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      
      - name: Set up OPA
        run: |
          curl -L -o opa https://openpolicyagent.org/downloads/latest/opa_linux_amd64
          chmod 755 opa
          sudo mv opa /usr/local/bin/opa
      
      - name: Validate policies
        run: |
          cd CoreModules/IAM/services/identity-service/policies
          opa check ./**/*.rego
      
      - name: Run policy tests
        run: |
          cd CoreModules/IAM/services/identity-service/policies
          opa test . --verbose
      
      - name: Bundle policies
        run: |
          cd CoreModules/IAM/services/identity-service/policies
          opa build -b role/ -o iam-role-policies.tar.gz
      
      - name: Deploy to OPA servers
        uses: innovabiz/opa-deploy-action@v1
        with:
          bundle-path: CoreModules/IAM/services/identity-service/policies/iam-role-policies.tar.gz
          environment: ${{ github.ref == 'refs/heads/main' && 'production' || 'staging' }}
          signature-key: ${{ secrets.OPA_SIGNATURE_KEY }}
```

### 4. Inicialização do Servidor OPA

O servidor OPA deve ser configurado para carregar as políticas e dados:

```yaml
# opa-config.yaml
services:
  - name: innovabiz-iam
    url: https://config.innovabiz.com/opa/bundles

bundles:
  innovabiz/iam/role:
    service: innovabiz-iam
    resource: bundles/iam-role-policies.tar.gz
    persist: true
    polling:
      min_delay_seconds: 60
      max_delay_seconds: 120

decision_logs:
  service: innovabiz-iam
  reporting:
    min_delay_seconds: 30
    max_delay_seconds: 60

status:
  service: innovabiz-iam
```

## Fluxo de Autorização

O fluxo completo de autorização segue estas etapas:

1. **Requisição HTTP recebida** pelo servidor API do RoleService
2. **Middleware de autenticação** valida o token JWT e extrai claims do usuário
3. **Handler da rota** extrai dados do recurso da requisição
4. **Middleware de autorização** prepara o input para OPA, incluindo:
   - Informações do usuário autenticado
   - Dados do recurso
   - Contexto da requisição (headers, IP, timestamp, etc.)
5. **Consulta ao OPA** para decisão de autorização
6. **Avaliação de políticas** pelo OPA baseada em:
   - Roles e permissões do usuário
   - Tipo de recurso e operação
   - Restrições de tenant
   - Dados de contexto
7. **Decisão retornada** com allow/deny e razão
8. **Middleware processa a decisão**:
   - Se autorizado: continua para o handler
   - Se negado: retorna 403 Forbidden com a razão
9. **Logging e auditoria** da decisão de autorização
10. **Execução da operação** pelo handler se autorizada

## Teste e Validação

### 1. Teste Local das Políticas

```bash
# No diretório policies/role
make test-policies
```

### 2. Simulação de Decisões

```bash
# Simular decisão usando exemplo
make simulate-decision POLICY=role/role_crud.rego INPUT=role/examples/super_admin_create_role.json
```

### 3. Teste de Integração

```go
// integration_test.go
func TestRoleAuthorizationFlow(t *testing.T) {
    // Configurar servidor OPA de teste
    // Configurar servidor HTTP de teste
    // Executar casos de teste para diversos perfis de usuário e operações
}
```

## Monitoramento e Observabilidade

Para garantir a eficácia das políticas de autorização, implemente:

1. **Métricas de decisão**:
   - Contagem de decisões por política
   - Latência de decisões
   - Taxa de autorização/negação

2. **Alertas**:
   - Alto volume de decisões negadas (possível ataque)
   - Latência elevada (possível problema de performance)
   - Falhas na consulta ao OPA (indisponibilidade)

3. **Dashboards**:
   - Visão geral de autorização por serviço/endpoint
   - Distribuição de decisões por tenant
   - Tendências de uso de permissões

## Melhores Práticas

1. **Cache de decisões**: Implemente cache para decisões frequentes e não sensíveis a mudanças de contexto
2. **Fallback local**: Em caso de indisponibilidade do OPA, use políticas locais simplificadas
3. **Atualizações atômicas**: Distribua políticas de forma que não quebrem compatibilidade durante updates
4. **Auditoria extensiva**: Registre todas as decisões de autorização para fins de compliance
5. **Testes de carga**: Valide a performance das políticas sob carga elevada

## Considerações de Segurança

1. **Proteção de dados**: Não inclua dados sensíveis nos inputs para OPA
2. **Verificação de integridade**: Assine bundles de políticas para garantir autenticidade
3. **Isolamento**: Garanta que políticas de diferentes tenants não interfiram entre si
4. **Monitoramento de anomalias**: Detecte padrões suspeitos de autorização
5. **Atualização de políticas**: Mantenha processo ágil para correção de vulnerabilidades

## Próximos Passos

1. Implementar rollout gradual por ambiente (dev, staging, prod)
2. Integrar com sistema de gestão de identidades para dados em tempo real
3. Desenvolver console de administração para visualização e auditoria de políticas
4. Implementar análise de impacto para mudanças em políticas
5. Expandir cobertura de testes de integração

---

© 2025 INNOVABIZ - Todos os direitos reservados