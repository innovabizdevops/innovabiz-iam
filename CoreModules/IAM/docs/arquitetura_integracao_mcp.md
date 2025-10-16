# Arquitetura de Integração IAM-MCP (Model Context Protocol)

## Visão Geral

Este documento técnico detalha a arquitetura de integração entre o módulo IAM (Identity and Access Management) e o MCP (Model Context Protocol) na plataforma INNOVABIZ. O MCP atua como padrão fundamental para conectar sistemas de IA e serviços, permitindo que o IAM forneça contexto seguro e adaptativo para decisões baseadas em modelos de IA, garantindo compliance, segurança e observabilidade em todos os fluxos de informação.

## Objetivos

1. **Integração Contextual**: Fornecer contexto de identidade, permissões e tenant para sistemas de IA
2. **Segurança Adaptativa**: Implementar controles de acesso dinâmicos baseados em análise de comportamento
3. **Propagação de Contexto**: Garantir consistência de contexto de segurança em todas as camadas da aplicação
4. **Rastreabilidade Total**: Assegurar auditoria completa de todas as interações com modelos de IA
5. **Governança de IA**: Implementar controles para uso responsável e ético de sistemas de IA
6. **Compliance Automatizada**: Verificar conformidade regulatória em tempo real para operações de IA
7. **Multi-tenancy Rigoroso**: Manter isolamento completo entre contextos de diferentes tenants

## Arquitetura MCP na Plataforma INNOVABIZ

### Componentes Fundamentais

![Arquitetura MCP-IAM](../assets/mcp_iam_architecture.png)

| Componente | Descrição | Responsabilidade |
|------------|-----------|-----------------|
| MCP Server | Implementação core do protocolo MCP | Gerenciamento de contexto, roteamento de ferramentas, logs |
| IAM MCP Provider | Adaptador IAM para MCP | Contexto de identidade, validação de autorização |
| MCP Tool Registry | Catálogo de ferramentas disponíveis | Registro, versionamento, descoberta de capacidades |
| MCP Context Store | Armazenamento de contextos | Persistência, versionamento, recuperação |
| Security Gateway | Interceptor de chamadas MCP | Validação, autorização, auditoria |
| Observability Collector | Coleta de telemetria MCP | Métricas, traces, logs de operações |

### Integração com API Gateway (KrakenD)

```
Cliente → KrakenD API Gateway → Security Filter → MCP Gateway → MCP Server → IAM MCP Provider
                                                                          → Outros Providers
```

O KrakenD como API Gateway realiza:

1. Roteamento inicial de requisições
2. Autenticação preliminar via JWT/OAuth2
3. Rate limiting por tenant e usuário
4. Transformações básicas de payload
5. Logging e métricas de primeiro nível

### Fluxo de Processamento MCP-IAM

1. Requisição entra via KrakenD com token JWT
2. Security Filter valida token e extrai claims
3. MCP Gateway enriquece contexto com informações de identidade
4. MCP Server roteia para o provider adequado
5. IAM MCP Provider valida permissões específicas para a operação
6. Resposta é enriquecida com contexto de segurança
7. Telemetria completa é registrada para auditoria

## Implementação do IAM MCP Provider

### Responsabilidades do Provider

1. **Validação Contextual**: Verificar permissões no contexto específico da requisição
2. **Enriquecimento de Contexto**: Adicionar informações de identidade e tenant
3. **Filtragem de Dados**: Aplicar controles de acesso a nível de campo
4. **Propagação de Políticas**: Garantir que políticas de segurança são respeitadas
5. **Registro de Auditoria**: Gravar eventos de segurança para todas interações

### Interface do IAM MCP Provider

```go
// IAMMCPProvider implementa a interface MCP Provider
type IAMMCPProvider struct {
    authService       auth.Service
    policyEngine     policy.Engine
    contextManager   context.Manager
    auditLogger      audit.Logger
    telemetryTracer  telemetry.Tracer
}

// ProcessRequest valida e enriquece requisições MCP
func (p *IAMMCPProvider) ProcessRequest(ctx context.Context, req *mcp.Request) (*mcp.Response, error) {
    // Extrair identidade do contexto
    identity, err := p.extractIdentity(ctx, req)
    if err != nil {
        return nil, p.handleAuthError(ctx, err, req)
    }
    
    // Verificar permissões para a operação
    decision, err := p.policyEngine.Evaluate(ctx, identity, req.Operation, req.Resource)
    if err != nil || !decision.Allowed {
        return nil, p.handleAccessDenied(ctx, decision, req)
    }
    
    // Enriquecer contexto da requisição
    enrichedCtx, err := p.contextManager.Enrich(ctx, identity, req)
    if err != nil {
        return nil, p.handleEnrichmentError(ctx, err, req)
    }
    
    // Processar a requisição com contexto enriquecido
    resp, err := p.processWithContext(enrichedCtx, req)
    
    // Registrar auditoria da operação
    p.auditLogger.LogAccess(ctx, identity, req, resp, err)
    
    return resp, err
}

// Métodos adicionais de suporte...
```

### Modelo de Contexto MCP-IAM

O contexto transferido entre sistemas inclui:

```json
{
  "request_id": "req-1234567890",
  "timestamp": "2025-08-06T15:04:05.123Z",
  "identity": {
    "tenant_id": "tenant-12345",
    "tenant_type": "financial_institution",
    "tenant_tier": "enterprise",
    "user_id": "user-98765",
    "username": "joao.silva",
    "roles": ["finance-manager", "payment-approver"],
    "groups": ["finance-department", "high-value-approvers"],
    "permissions": ["payments:approve:high", "reports:view:financial"],
    "session_id": "sess-abcdef",
    "authentication_level": "mfa_verified",
    "risk_score": 92
  },
  "security": {
    "classification_level": "confidential",
    "integrity_level": "high",
    "encryption_required": true,
    "audit_level": "detailed",
    "data_residency": ["BR", "AO", "PT"],
    "compliance_frameworks": ["pci-dss", "gdpr", "lgpd"]
  },
  "resources": {
    "quotas": {
      "api_calls_per_minute": 1000,
      "model_tokens_per_hour": 50000,
      "storage_mb": 5000
    },
    "priority": "high",
    "rate_limiting": {
      "tokens_per_interval": 100,
      "interval_seconds": 60
    }
  },
  "environment": {
    "deployment": "production",
    "region": "south-america-east1",
    "ip_address": "192.168.1.50",
    "user_agent": "Mozilla/5.0...",
    "device_id": "dev-abc123",
    "channel": "web_portal"
  },
  "tracing": {
    "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
    "span_id": "00f067aa0ba902b7",
    "parent_id": "df42f8223b3w21a9",
    "sampling_priority": 1
  }
}
```

## Servidores MCP Suportados

O IAM integra-se com os seguintes servidores MCP para diferentes casos de uso:

### 1. MCP_DOCKER

Responsável por operações de infraestrutura e containerização segura.

**Integrações IAM:**
- Autenticação e autorização para operações Kubernetes/Docker
- Gerenciamento de segredos por tenant para deployments
- Políticas de segurança para contêineres e namespaces
- Auditoria de operações de infraestrutura

**Fluxos Principais:**
1. Provisionamento seguro de infraestrutura por tenant
2. Isolamento de contêineres e recursos por tenant
3. Deployments com contexto de segurança

### 2. Desktop-Commander

Interface de operações para administradores de sistema.

**Integrações IAM:**
- Controle granular de comandos por role/tenant
- Auditoria de operações administrativas
- Sessões com contexto de segurança
- Isolamento de recursos de file system por tenant

**Fluxos Principais:**
1. Operações administrativas com elevação de privilégios
2. Automação de tarefas com contexto de segurança
3. Troubleshooting com garantia de isolamento

### 3. Figma MCP

Integração para design e prototipação segura.

**Integrações IAM:**
- Controle de acesso a assets de design por role/tenant
- Versionamento seguro de recursos de UI/UX
- Colaboração com contexto de segurança
- Aprovações baseadas em workflows

**Fluxos Principais:**
1. Design de interfaces com restrições de branding por tenant
2. Aprovações de design com contexto de identidade
3. Exportação segura de assets

### 4. GitHub MCP

Gestão de código e CI/CD com contexto de segurança.

**Integrações IAM:**
- Autenticação federada para repositórios
- Controle de acesso a branches e repositórios
- Code review com contexto de segurança
- Auditoria de operações de código

**Fluxos Principais:**
1. Integração contínua com contexto de segurança
2. Gestão de código com políticas por tenant
3. Automação de workflows de release segura

### 5. Memory MCP

Gestão de informações persistentes entre sessões.

**Integrações IAM:**
- Isolamento de memória por tenant
- Persistência segura de contextos de usuário
- Criptografia de dados por tenant
- Políticas de retenção por tipo de dado

**Fluxos Principais:**
1. Persistência de preferências de usuário com segurança
2. Recuperação de contexto entre sessões
3. Compartilhamento seguro de conhecimento entre equipes

## Padrões de Integração MCP-IAM

### 1. Circuit Breaking

Implementação de mecanismos de circuit breaking para evitar cascata de falhas:

```go
// Exemplo simplificado de circuit breaker para chamadas MCP
func (p *IAMMCPProvider) callWithCircuitBreaker(ctx context.Context, req *mcp.Request) (*mcp.Response, error) {
    tenant := extractTenantFromContext(ctx)
    
    // Obter ou criar circuit breaker para este tenant
    cb := p.circuitBreakers.ForTenant(tenant.ID)
    
    // Executar com circuit breaker
    return cb.Execute(func() (*mcp.Response, error) {
        // Chamada real para o serviço MCP
        return p.mcpClient.Execute(ctx, req)
    })
}
```

### 2. Backpressure & Rate Limiting

Controle de fluxo para proteger serviços downstream:

```go
// Rate limiting por tenant para chamadas MCP
func (p *IAMMCPProvider) applyRateLimiting(ctx context.Context, tenant *models.Tenant) error {
    // Obter limites configurados para o tenant
    limits := p.configService.GetTenantLimits(tenant.ID)
    
    // Aplicar rate limiting baseado em token bucket
    limiter := p.limiters.ForTenant(tenant.ID, limits)
    
    // Verificar se requisição pode prosseguir
    if !limiter.Allow() {
        metrics.RecordRateLimited(tenant.ID)
        return errors.NewRateLimitExceededError(tenant.ID)
    }
    
    return nil
}
```

### 3. Tenant Context Propagation

Propagação consistente de contexto de tenant entre serviços:

```go
// Middleware para propagação de contexto de tenant
func TenantContextPropagationMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extrair tenant_id do token JWT
        tenantID := extractTenantFromToken(c)
        if tenantID == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "Tenant context missing"})
            return
        }
        
        // Adicionar tenant_id aos headers para propagação
        c.Request.Header.Set("X-Tenant-ID", tenantID)
        
        // Adicionar ao contexto para uso interno
        ctx := context.WithValue(c.Request.Context(), TenantContextKey, tenantID)
        c.Request = c.Request.WithContext(ctx)
        
        // Adicionar ao contexto de tracing
        span := trace.SpanFromContext(ctx)
        span.SetAttributes(attribute.String("tenant.id", tenantID))
        
        c.Next()
    }
}
```

## Controle de Acesso a Ferramentas MCP

### 1. Modelo de Autorização para Ferramentas MCP

O sistema implementa um modelo granular de autorização para acesso a ferramentas MCP:

```json
{
  "tool_policy": {
    "tool_id": "docker_kubernetes_deploy",
    "tenant_id": "tenant-12345",
    "allowed_roles": ["devops-admin", "platform-engineer"],
    "denied_roles": ["developer-junior", "support-basic"],
    "conditions": {
      "environment": ["development", "staging"],
      "time_window": {
        "start": "08:00:00",
        "end": "18:00:00",
        "timezone": "America/Sao_Paulo"
      },
      "approval_required": {
        "roles": ["platform-manager"],
        "threshold": 1,
        "expiry_hours": 24
      },
      "risk_score_min": 80,
      "authentication_level": "mfa_verified"
    },
    "quotas": {
      "max_daily_invocations": 100,
      "max_resources": {
        "cpu": "4",
        "memory": "8Gi",
        "storage": "100Gi"
      }
    },
    "audit_level": "detailed"
  }
}
```

### 2. Validação de Ferramenta em Runtime

```go
func (v *MCPToolValidator) ValidateTool(ctx context.Context, toolRequest *mcp.ToolRequest) (*mcp.ValidationResult, error) {
    // Extrair identidade e tenant
    identity, err := auth.IdentityFromContext(ctx)
    if err != nil {
        return nil, err
    }
    
    // Carregar política para a ferramenta no contexto do tenant
    policy, err := v.policyStore.GetToolPolicy(identity.TenantID, toolRequest.ToolID)
    if err != nil {
        return nil, err
    }
    
    // Verificar roles permitidos/negados
    if !isRoleAllowed(identity.Roles, policy) {
        return &mcp.ValidationResult{
            Allowed: false,
            Reason: "role_not_allowed",
        }, nil
    }
    
    // Verificar condições contextuais
    if !v.contextValidator.ValidateConditions(ctx, policy.Conditions) {
        return &mcp.ValidationResult{
            Allowed: false,
            Reason: "conditions_not_met",
        }, nil
    }
    
    // Verificar quotas
    if exceeded, limit := v.quotaManager.CheckQuotaExceeded(ctx, identity.TenantID, toolRequest.ToolID); exceeded {
        return &mcp.ValidationResult{
            Allowed: false,
            Reason: "quota_exceeded",
            Details: fmt.Sprintf("Limit of %d invocations reached", limit),
        }, nil
    }
    
    // Registrar tentativa de uso para auditoria
    v.auditLogger.LogToolAccess(ctx, identity, toolRequest, true, "")
    
    return &mcp.ValidationResult{
        Allowed: true,
        AuditLevel: policy.AuditLevel,
    }, nil
}
```

## Monitoramento e Observabilidade

### 1. Métricas Específicas MCP

| Métrica | Tipo | Tags | Descrição |
|---------|------|------|-----------|
| `mcp.tools.invocations` | Counter | tool_id, tenant_id, status | Contagem de invocações de ferramentas |
| `mcp.tools.latency` | Histogram | tool_id, tenant_id | Latência de execução de ferramentas |
| `mcp.tools.errors` | Counter | tool_id, tenant_id, error_type | Contagem de erros por tipo |
| `mcp.auth.validations` | Counter | tenant_id, decision | Decisões de autorização |
| `mcp.quota.usage` | Gauge | tenant_id, tool_id, resource_type | Uso atual vs limite |
| `mcp.context.size` | Histogram | tenant_id | Tamanho dos contextos (bytes) |
| `mcp.tools.waiting_approval` | Gauge | tenant_id, tool_id | Invocações aguardando aprovação |

### 2. Logs Estruturados

```json
{
  "timestamp": "2025-08-06T15:04:05.123Z",
  "level": "INFO",
  "event": "mcp.tool.invocation",
  "tenant_id": "tenant-12345",
  "user_id": "user-98765",
  "tool_id": "github_create_repository",
  "session_id": "sess-abcdef",
  "request_id": "req-1234567890",
  "status": "success",
  "duration_ms": 345,
  "parameters": {
    "sensitive_redacted": true,
    "name": "novo-servico",
    "private": true
  },
  "authentication_context": {
    "method": "oauth",
    "mfa_verified": true,
    "location": "BR-SP",
    "ip_address": "192.168.1.50"
  },
  "authorization_context": {
    "roles": ["developer-senior"],
    "permissions": ["github:repo:create"],
    "policy_version": "v1.2.3"
  },
  "resource_usage": {
    "api_calls": 3,
    "quota_remaining": {
      "daily_invocations": 97,
      "repositories": 45
    }
  },
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "compliance": {
    "audit_trail_id": "audit-9876543"
  }
}
```

### 3. Traces Distribuídos

Implementação de traces para visualizar fluxo completo de execução:

1. Início no IAM Provider
2. Propagação para MCP Server
3. Execução da ferramenta específica
4. Chamadas a sistemas externos
5. Retorno e processamento de resposta

Cada span inclui contexto de tenant, usuário e segurança para rastreabilidade completa.

## Governança de IA e MCP

### 1. Políticas de Uso Responsável

O IAM implementa controles para uso ético e responsável de ferramentas de IA:

1. **Classificação de Ferramentas**: Categorização por risco e impacto
2. **Controles por Categoria**: Requisitos específicos por nível de risco
3. **Limites de Uso**: Quotas e thresholds por tenant e usuário
4. **Aprovações Multi-nível**: Workflow para ferramentas sensíveis
5. **Monitoramento de Viés**: Detecção de padrões discriminatórios

### 2. Modelo de Responsabilidade Compartilhada

| Entidade | Responsabilidades |
|----------|-------------------|
| IAM Provider | Autenticação, autorização, auditoria, contexto de segurança |
| MCP Server | Routing, validação de parâmetros, sanitização, logging |
| Tool Provider | Implementação segura, validação específica, documentação |
| Client Application | UX responsável, feedback ao usuário, consentimento |
| Tenant Admin | Políticas específicas, controles personalizados, revisão |

### 3. Compliance Automatizada

Implementação de verificações contínuas de conformidade:

1. **Guardrails em Runtime**: Validação de conformidade em tempo real
2. **Monitoramento Contínuo**: Detecção de desvios de políticas
3. **Registros Imutáveis**: Logs criptograficamente verificáveis
4. **Explicabilidade**: Registro de justificativas para decisões
5. **Notificações de Risco**: Alertas para potenciais violações

## Plano de Implementação e Evolução

### Fase 1: Fundação (M1-M2)

1. Implementação do IAM MCP Provider básico
2. Integração com autenticação e autorização existentes
3. Logging estruturado para auditoria
4. Suporte inicial para MCP_DOCKER e Memory MCP

### Fase 2: Segurança Avançada (M3-M5)

1. Políticas granulares para ferramentas MCP
2. Contexto enriquecido com informações de segurança
3. Validação avançada de requisições
4. Integração com GitHub MCP e Desktop-Commander

### Fase 3: Inteligência e Automação (M6-M8)

1. Detecção de anomalias em uso de ferramentas
2. Políticas adaptativas baseadas em comportamento
3. Workflows de aprovação automatizados
4. Integração com Figma MCP

### Fase 4: Otimização e Governança (M9-M12)

1. Dashboards específicos para governança de IA
2. Relatórios automatizados de compliance
3. Self-service para configuração de políticas por tenant
4. Framework extensível para novos servidores MCP

## Considerações de Segurança

1. **Proteção de Contexto**: Prevenir manipulação ou falsificação de contexto
2. **Minimização de Dados**: Incluir apenas o necessário no contexto
3. **Validação Rigorosa**: Validar todas entradas e saídas de ferramentas
4. **Auditoria Completa**: Registrar todas operações para investigação
5. **Criptografia**: Proteger dados sensíveis em contexto
6. **Prevenção de Injeção**: Sanitizar parâmetros e entradas
7. **Revogação de Acesso**: Mecanismo para revogar acesso em tempo real

## Referências

1. Model Context Protocol Specification
2. NIST AI Risk Management Framework (AI RMF)
3. ISO/IEC 42001 - Artificial Intelligence Management Systems
4. OWASP API Security Top 10 2023
5. Zero Trust Architecture (NIST SP 800-207)
6. TOGAF 10.0 - The Open Group Architecture Framework
7. COBIT 2019 - Framework de Governança de TI
8. KrakenD API Gateway Documentation
9. OpenTelemetry Specification

---

*Este documento está em conformidade com os padrões de documentação técnica da INNOVABIZ e deve ser revisado e atualizado regularmente conforme a evolução do sistema.*

*Última atualização: 06/08/2025*