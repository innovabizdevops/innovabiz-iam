# Matriz de Integração Avançada: MCP-IAM Elevation Hooks

**Documento**: INNOVABIZ-IAM-MATRIX-MCP-HOOKS-INT-v1.0.0  
**Classificação**: Confidencial-Interno  
**Data**: 06/08/2025  
**Estado**: Aprovado  
**Âmbito**: Multi-Mercado, Multi-Camada, Multi-Dimensional, Multi-Contexto  
**Elaborado por**: Equipa de Arquitetura INNOVABIZ

## Índice

1. [Visão Geral da Integração](#visão-geral-da-integração)
2. [Arquitetura de Integração Total](#arquitetura-de-integração-total)
3. [Matriz de Integração Detalhada](#matriz-de-integração-detalhada)
4. [Especificação de APIs e Interfaces](#especificação-de-apis-e-interfaces)
5. [Modelo Multi-Dimensional](#modelo-multi-dimensional)
6. [Extensibilidade e Customização](#extensibilidade-e-customização)
7. [Observabilidade Avançada](#observabilidade-avançada)
8. [Conformidade Multi-Regulatória](#conformidade-multi-regulatória)
9. [Anexos Técnicos](#anexos-técnicos)

## Visão Geral da Integração

A Matriz de Integração Avançada define a implementação técnica detalhada para a integração dos hooks MCP-IAM de elevação de privilégios com todos os módulos core da plataforma INNOVABIZ. Esta integração segue os princípios de Arquitetura de Integração Total, garantindo interoperabilidade completa, isolamento adequado, conformidade regulatória específica por mercado e monitorização avançada.

Este documento fornece as especificações técnicas necessárias para a integração em múltiplas dimensões, considerando os contextos específicos de cada módulo, mercado, tenant e camada tecnológica.

### Princípios Fundamentais de Integração

1. **Integração por Contrato**: APIs e interfaces claramente definidas com contratos explícitos
2. **Isolamento de Falhas**: Padrão de Circuit Breaker para evitar propagação de falhas
3. **Observabilidade Total**: Instrumentação em todas as camadas e pontos de integração
4. **Configuração Dinâmica**: Políticas e regras ajustáveis por mercado e tenant
5. **Segurança por Desenho**: Validação de segurança em cada ponto de integração
6. **Conformidade Incorporada**: Conformidade regulatória como requisito funcional
7. **Adaptabilidade Contextual**: Comportamento adaptativo baseado no contexto de execução

## Arquitetura de Integração Total

A arquitetura de integração dos hooks MCP-IAM segue o modelo de Integração Total Avançada, baseado nos seguintes componentes:

### Camadas da Arquitetura

1. **Camada de Apresentação**
   - Dashboards de monitoramento e auditoria
   - Interfaces de aprovação para operações privilegiadas
   - Portais de administração para configuração de políticas

2. **Camada de APIs**
   - API Gateway (KrakenD) para roteamento e controle de acesso
   - APIs REST para serviços síncronos
   - APIs GraphQL para consultas complexas
   - APIs gRPC para comunicação interna de alta performance

3. **Camada de Serviços**
   - Serviço core de elevação de privilégios
   - Registro e gestão de hooks MCP
   - Serviços de validação e autorização
   - Serviços de auditoria e logging

4. **Camada de Comunicação**
   - Event Bus (Kafka) para comunicação assíncrona
   - Sistema de mensageria para notificações
   - WebSockets para atualizações em tempo real

5. **Camada de Persistência**
   - PostgreSQL para armazenamento relacional
   - Redis para caching e dados temporários
   - Elasticsearch para logs e dados de auditoria

### Fluxo de Dados de Integração

```
┌────────────┐    ┌────────────┐    ┌────────────┐    ┌────────────┐
│  Módulo    │    │ API Gateway│    │  Serviço   │    │ Hook MCP   │
│  Cliente   │───>│  (KrakenD) │───>│  Elevação  │───>│ Específico │
└────────────┘    └────────────┘    └────────────┘    └────────────┘
                                          │                  │
                                          ▼                  ▼
                                    ┌────────────┐    ┌────────────┐
                                    │  Sistema   │    │  Sistema   │
                                    │  Auditoria │    │    MFA     │
                                    └────────────┘    └────────────┘
```

### Integração com Infraestrutura Multi-Cloud

A arquitetura suporta deployment em múltiplas nuvens e ambientes on-premise, garantindo:

- **Portabilidade**: Containerização com Docker e orquestração Kubernetes
- **Resiliência**: Replicação em múltiplos clusters e regiões
- **Localização de Dados**: Conformidade com requisitos de residência de dados
- **Recuperação**: Estratégias de DR com RPO < 15min e RTO < 30min

## Matriz de Integração Detalhada

### 1. Payment Gateway

#### 1.1 Pontos de Integração

| Componente | Método de Integração | Protocolo | Padrão de Comunicação |
|------------|----------------------|-----------|------------------------|
| ConfigService | API REST | HTTPS | Síncrono/Request-Response |
| TransactionService | gRPC | HTTP/2 | Síncrono/Request-Response |
| RefundService | API REST | HTTPS | Síncrono/Request-Response |
| LimitService | API REST | HTTPS | Síncrono/Request-Response |
| AuditService | Kafka | TCP | Assíncrono/Pub-Sub |

#### 1.2 Mapeamento de Escopos para Operações

| Escopo | Operações | Sensibilidade | MFA Requerido | Aprovação |
|--------|-----------|---------------|---------------|-----------|
| payment:config | updateGateway, createProcessor, updateKeys | Alta | Forte | Sim |
| payment:refund | createRefund, bulkRefund | Média | Básico | Condicional |
| payment:limits | updateLimits, overrideLimit | Média | Básico | Sim |
| payment:fees | updateFee, createPromo | Média | Básico | Sim |
| payment:admin | allOperations | Muito Alta | Forte | Sim |

#### 1.3 Adaptações por Mercado

| Mercado | Adaptações Específicas | Conformidade |
|---------|------------------------|--------------|
| Angola | Aprovação dupla para transações > 1M AKZ | BNA |
| Moçambique | Validação adicional para operações em divisas | Banco de Moçambique |
| Brasil | Logs específicos para transações PIX | BACEN |
| UE | Validação adicional para GDPR | EBA, GDPR |
| EUA | Conformidade para transações internacionais | FinCEN, OFAC |
| China | Localização de dados e validações específicas | Regulações locais |

### 2. Mobile Money

#### 2.1 Pontos de Integração

| Componente | Método de Integração | Protocolo | Padrão de Comunicação |
|------------|----------------------|-----------|------------------------|
| WalletService | gRPC | HTTP/2 | Síncrono/Request-Response |
| AgentService | API REST | HTTPS | Síncrono/Request-Response |
| KYCService | API REST | HTTPS | Síncrono/Request-Response |
| TransactionService | gRPC | HTTP/2 | Síncrono/Request-Response |
| NotificationService | Kafka | TCP | Assíncrono/Pub-Sub |

#### 2.2 Mapeamento de Escopos para Operações

| Escopo | Operações | Sensibilidade | MFA Requerido | Aprovação |
|--------|-----------|---------------|---------------|-----------|
| mobile:wallet | createWallet, blockWallet, updateLimit | Média | Básico | Não |
| mobile:agent | createAgent, updateCommission, blockAgent | Alta | Forte | Sim |
| mobile:limits | updateSystemLimits, createLimitGroup | Alta | Forte | Sim |
| mobile:kyc | updateKYCRules, overrideKYCLevel | Alta | Forte | Sim |
| mobile:admin | allOperations | Muito Alta | Forte | Sim |

#### 2.3 Adaptações por Mercado

| Mercado | Adaptações Específicas | Conformidade |
|---------|------------------------|--------------|
| Angola | Validação especial para operações de câmbio | BNA |
| Moçambique | Limite específico para transações por agente | Banco de Moçambique |
| Brasil | Validações específicas para arranjos de pagamento | BACEN |
| UE | Verificações AML adicionais | EBA, 5AMLD |
| EUA | Verificação em listas de sanções | OFAC |
| SADC | Validações para transações transfronteiriças | Regulações regionais |

### 3. E-Commerce & Marketplace

#### 3.1 Pontos de Integração

| Componente | Método de Integração | Protocolo | Padrão de Comunicação |
|------------|----------------------|-----------|------------------------|
| SellerService | API REST | HTTPS | Síncrono/Request-Response |
| ProductService | API REST | HTTPS | Síncrono/Request-Response |
| OrderService | gRPC | HTTP/2 | Síncrono/Request-Response |
| PricingService | API REST | HTTPS | Síncrono/Request-Response |
| PromotionService | Kafka | TCP | Assíncrono/Pub-Sub |

#### 3.2 Mapeamento de Escopos para Operações

| Escopo | Operações | Sensibilidade | MFA Requerido | Aprovação |
|--------|-----------|---------------|---------------|-----------|
| ecommerce:seller | approveVerification, updateFees | Alta | Básico | Sim |
| ecommerce:product | bulkUpdate, approveRestrictedCategory | Média | Básico | Condicional |
| ecommerce:pricing | updatePriceStrategy, bulkDiscount | Média | Básico | Condicional |
| ecommerce:order | cancelConfirmedOrder, modifyShipping | Média | Básico | Não |
| ecommerce:admin | allOperations | Muito Alta | Forte | Sim |

#### 3.3 Adaptações por Mercado

| Mercado | Adaptações Específicas | Conformidade |
|---------|------------------------|--------------|
| Angola | Verificações adicionais para produtos importados | Regulações alfandegárias |
| Moçambique | Validação de licenças para produtos específicos | Regulações de comércio |
| Brasil | Validações específicas para Nota Fiscal | Requisitos SEFAZ |
| UE | Verificações de conformidade de produtos | Diretivas de proteção ao consumidor |
| EUA | Validações específicas para produtos regulados | Regulações FDA, CPSC |
| China | Requisitos específicos para marketplace | Regulações de e-commerce locais |

### 4. Risk Management

#### 4.1 Pontos de Integração

| Componente | Método de Integração | Protocolo | Padrão de Comunicação |
|------------|----------------------|-----------|------------------------|
| RiskEngineService | gRPC | HTTP/2 | Síncrono/Request-Response |
| FraudDetectionService | gRPC | HTTP/2 | Síncrono/Request-Response |
| RuleManagerService | API REST | HTTPS | Síncrono/Request-Response |
| AlertService | Kafka | TCP | Assíncrono/Pub-Sub |
| ReportingService | API REST | HTTPS | Síncrono/Request-Response |

#### 4.2 Mapeamento de Escopos para Operações

| Escopo | Operações | Sensibilidade | MFA Requerido | Aprovação |
|--------|-----------|---------------|---------------|-----------|
| risk:rules | updateRiskRules, createRiskModel | Alta | Forte | Sim |
| risk:thresholds | updateThresholds, createRiskLevel | Média | Básico | Sim |
| risk:override | overrideRiskScore, whitelistEntity | Alta | Forte | Sim |
| risk:reporting | generateReport, configureAlert | Baixa | Básico | Não |
| risk:admin | allOperations | Muito Alta | Forte | Sim |

#### 4.3 Adaptações por Mercado

| Mercado | Adaptações Específicas | Conformidade |
|---------|------------------------|--------------|
| Angola | Regras específicas para operações em USD | BNA |
| Moçambique | Monitoramento especial para transações de alto valor | Banco de Moçambique |
| Brasil | Integração com listas de PEP locais | COAF, BACEN |
| UE | Verificações reforçadas para AML/CFT | 5AMLD, 6AMLD |
| EUA | Verificações OFAC e FinCEN | BSA, OFAC |
| Global | Monitoramento para sanções internacionais | FATF |

## Especificação de APIs e Interfaces

### 1. API Gateway (KrakenD)

O API Gateway KrakenD atua como ponto central de entrada para todas as solicitações de elevação, fornecendo:

- Roteamento inteligente para serviços de backend
- Validação de JWT e autenticação
- Rate limiting e controle de tráfego
- Transformação de requisições e respostas
- Agregação de múltiplos endpoints
- Métricas e telemetria em tempo real

#### Configuração por Módulo

```json
{
  "endpoints": [
    {
      "endpoint": "/v1/iam/elevation/{service}",
      "method": "POST",
      "backend": [
        {
          "url_pattern": "/elevation/{service}",
          "host": ["http://elevation-service:8080"],
          "method": "POST"
        }
      ],
      "extra_config": {
        "auth/validator": {
          "alg": "RS256",
          "jwk_url": "http://keycloak:8080/auth/realms/innovabiz/protocol/openid-connect/certs",
          "disable_jwk_security": false
        },
        "qos/ratelimit/router": {
          "max_rate": 100,
          "client_ip_strategy": "X-Forwarded-For"
        }
      }
    }
  ]
}
```

### 2. Serviço de Elevação

O serviço core de elevação expõe as seguintes APIs para integração com outros módulos:

#### 2.1 API REST

```
POST /v1/elevation/request
POST /v1/elevation/approve/{requestId}
POST /v1/elevation/reject/{requestId}
POST /v1/elevation/validate/{tokenId}
GET  /v1/elevation/requests?status={status}
GET  /v1/elevation/request/{requestId}
GET  /v1/elevation/token/{tokenId}
```

#### 2.2 API GraphQL

```graphql
type Query {
  elevationRequest(id: ID!): ElevationRequest
  elevationRequests(
    status: RequestStatus, 
    market: String, 
    tenantId: String, 
    hookType: String
  ): [ElevationRequest]
  elevationToken(id: ID!): ElevationToken
  elevationPolicies(
    market: String, 
    tenantId: String, 
    hookType: String
  ): [PolicyConfig]
}

type Mutation {
  createElevationRequest(input: ElevationRequestInput!): ElevationRequest
  approveElevationRequest(id: ID!, approverNotes: String): ElevationRequest
  rejectElevationRequest(id: ID!, reason: String!): ElevationRequest
  validateElevationToken(id: ID!, scope: String!, metadata: JSON!): ValidationResult
}
```

#### 2.3 API gRPC

```protobuf
service ElevationService {
  rpc RequestElevation(ElevationRequest) returns (ElevationRequestResponse);
  rpc ApproveRequest(ApprovalRequest) returns (ElevationRequestResponse);
  rpc RejectRequest(RejectionRequest) returns (ElevationRequestResponse);
  rpc ValidateToken(ValidationRequest) returns (ValidationResponse);
  rpc GetElevationRequests(RequestsQuery) returns (ElevationRequestsResponse);
  rpc GetElevationRequest(RequestQuery) returns (ElevationRequestResponse);
  rpc GetElevationToken(TokenQuery) returns (ElevationTokenResponse);
}
```

### 3. Event Bus (Kafka)

#### 3.1 Tópicos

```
innovabiz.iam.elevation.request.created
innovabiz.iam.elevation.request.approved
innovabiz.iam.elevation.request.rejected
innovabiz.iam.elevation.token.created
innovabiz.iam.elevation.token.used
innovabiz.iam.elevation.token.expired
```

#### 3.2 Formato de Evento

```json
{
  "eventId": "uuid",
  "eventType": "elevation.request.created",
  "timestamp": "2025-08-06T17:42:13Z",
  "version": "1.0",
  "tenantId": "tenant123",
  "market": "angola",
  "source": "elevation-service",
  "data": {
    "requestId": "req123",
    "userId": "user456",
    "scopes": ["docker:run"],
    "hookType": "docker",
    "emergency": false,
    "metadata": {}
  },
  "traceId": "trace123",
  "spanId": "span456"
}
```

## Modelo Multi-Dimensional

A integração dos hooks MCP-IAM segue um modelo multi-dimensional que considera simultaneamente:

### 1. Dimensão de Mercado

Adaptações específicas por região geográfica e jurisdição regulatória:

| Dimensão | Parâmetros Configuráveis | Exemplo |
|----------|--------------------------|---------|
| Requisitos MFA | mfa.level, mfa.step_up | {"angola": {"mfa.level": "STRONG"}} |
| Aprovações | approval.levels, approval.roles | {"brasil": {"approval.levels": 2}} |
| Limites Temporais | duration.max, duration.default | {"eu": {"duration.max": 240}} |
| Auditoria | audit.detail, audit.retention | {"angola": {"audit.detail": "FULL"}} |
| Notificações | notification.channels | {"mozambique": {"notification.channels": ["email", "sms"]}} |

### 2. Dimensão de Tenant

Configurações específicas por organização ou unidade de negócio:

| Dimensão | Parâmetros Configuráveis | Exemplo |
|----------|--------------------------|---------|
| Papéis | roles.approvers, roles.admins | {"tenant123": {"roles.approvers": ["security-officer"]}} |
| Escopos Permitidos | scopes.allowed, scopes.restricted | {"tenant456": {"scopes.restricted": ["docker:admin"]}} |
| Políticas de Emergência | emergency.allowed, emergency.approvers | {"tenant789": {"emergency.allowed": false}} |
| SLA | sla.approval, sla.processing | {"tenant123": {"sla.approval": 240}} |
| Limites | limits.active_tokens, limits.requests | {"tenant456": {"limits.active_tokens": 10}} |

### 3. Dimensão de Módulo

Adaptações específicas para cada módulo da plataforma:

| Dimensão | Parâmetros Configuráveis | Exemplo |
|----------|--------------------------|---------|
| Operações Sensíveis | sensitive.operations | {"payment": {"sensitive.operations": ["refund"]}} |
| Categorias de Risco | risk.categories | {"mobile": {"risk.categories": {"agent": "HIGH"}}} |
| Aprovadores | module.approvers | {"ecommerce": {"module.approvers": ["commerce-lead"]}} |
| Limites Específicos | module.limits | {"credit": {"module.limits": {"approval": 5000}}} |
| Metadados Específicos | module.metadata | {"insurance": {"module.metadata": ["policy_id"]}} |

### 4. Dimensão Temporal

Configurações que variam com base em parâmetros temporais:

| Dimensão | Parâmetros Configuráveis | Exemplo |
|----------|--------------------------|---------|
| Período do Dia | time.restrictions | {"payment": {"time.restrictions": {"after_hours": "APPROVAL_REQUIRED"}}} |
| Feriados | holiday.restrictions | {"angola": {"holiday.restrictions": "EMERGENCY_ONLY"}} |
| Horário Comercial | business.hours | {"brasil": {"business.hours": "09:00-18:00"}} |
| Períodos de Manutenção | maintenance.windows | {"global": {"maintenance.windows": ["SAT 22:00-02:00"]}} |
| Janelas de Mudança | change.windows | {"payment": {"change.windows": ["MON-THU 10:00-16:00"]}} |

## Extensibilidade e Customização

### 1. Framework de Plugins

O sistema de hooks MCP-IAM inclui um framework de plugins que permite estender funcionalidades sem modificação do código core:

```go
type MCPHookPlugin interface {
    Initialize(config map[string]interface{}) error
    Name() string
    Version() string
    PreProcessRequest(ctx context.Context, request *ElevationRequest) error
    PostProcessRequest(ctx context.Context, request *ElevationRequest, result *ElevationResult) error
    PreValidateToken(ctx context.Context, tokenID string, scope string, metadata map[string]interface{}) error
    PostValidateToken(ctx context.Context, tokenID string, scope string, metadata map[string]interface{}, result *ValidationResult) error
}
```

### 2. Templates de Configuração

Templates Jsonnet para geração dinâmica de configurações específicas por mercado e tenant:

```jsonnet
local base = import 'base.libsonnet';
local market = import 'markets/angola.libsonnet';
local tenant = import 'tenants/financialInstitution.libsonnet';

base {
  market: market,
  tenant: tenant,
  combined: {
    mfa: {
      level: if market.highRisk then 'STRONG' else tenant.mfaLevel,
    },
    approval: {
      required: market.requiresApproval || tenant.requiresApproval,
      levels: std.max(market.approvalLevels, tenant.approvalLevels),
    },
  },
}
```

### 3. Mecanismo de Extensão para Novos Hooks MCP

Processo para registrar novos hooks MCP no sistema:

1. Implementar a interface `MCPHook`
2. Criar arquivo de configuração específico
3. Registrar no `HookRegistry` durante inicialização
4. Implementar testes específicos de validação
5. Documentar escopos e operações suportadas

## Observabilidade Avançada

### 1. Telemetria OpenTelemetry

Instrumentação completa com OpenTelemetry para rastreamento distribuído:

```go
func (h *DockerHook) ValidateScope(ctx context.Context, scope string, tenantID string, market string) (*ScopeDetails, error) {
    ctx, span := h.tracer.Start(ctx, "DockerHook.ValidateScope",
        trace.WithAttributes(
            attribute.String("scope", scope),
            attribute.String("tenant_id", tenantID),
            attribute.String("market", market),
        ))
    defer span.End()
    
    // Lógica de validação...
    
    span.SetAttributes(attribute.Bool("approved", true))
    return details, nil
}
```

### 2. Métricas Prometheus

Métricas exportadas para monitoramento em tempo real:

```
# HELP innovabiz_iam_elevation_requests_total Total number of elevation requests
# TYPE innovabiz_iam_elevation_requests_total counter
innovabiz_iam_elevation_requests_total{hook_type="docker",market="angola",tenant="tenant123"} 45
innovabiz_iam_elevation_requests_total{hook_type="github",market="angola",tenant="tenant123"} 23

# HELP innovabiz_iam_elevation_request_duration_seconds Duration of elevation request processing
# TYPE innovabiz_iam_elevation_request_duration_seconds histogram
innovabiz_iam_elevation_request_duration_seconds_bucket{hook_type="docker",market="angola",tenant="tenant123",le="0.1"} 32
```

### 3. Dashboards Operacionais

Painéis de controle para monitoramento em tempo real:

- Dashboard de solicitações pendentes por hook e mercado
- Dashboard de uso de tokens por escopo e tenant
- Dashboard de duração de aprovações por mercado
- Dashboard de auditoria de operações privilegiadas
- Dashboard de conformidade regulatória por mercado

### 4. Alertas e Notificações

Sistema de alertas para detecção de anomalias e eventos críticos:

- Alertas para solicitações de elevação de alta sensibilidade
- Alertas para uso excessivo de modo emergencial
- Notificações para aprovações pendentes há mais de 2 horas
- Alertas para operações em recursos protegidos
- Notificações para expiração iminente de tokens

## Conformidade Multi-Regulatória

### 1. Matriz de Conformidade

| Regulamento | Mercados | Requisitos Implementados | Validação |
|-------------|----------|--------------------------|-----------|
| GDPR | UE, Global | Logging detalhado, Justificativas, Minimização | Revisão trimestral |
| LGPD | Brasil | DPO Approval, Justificativas específicas | Auditoria semestral |
| PCI-DSS | Global | Segregação de funções, Aprovação dupla | Certificação anual |
| BNA | Angola | Aprovações específicas para operações financeiras | Revisão regulatória |
| BACEN | Brasil | Logging específico para transações financeiras | Reporte mensal |
| SADC | Moçambique | Controles para operações transfronteiriças | Validação semestral |

### 2. Logs de Auditoria Específicos

Formato de logs específicos para requisitos regulatórios:

```json
{
  "timestamp": "2025-08-06T18:42:13Z",
  "level": "INFO",
  "market": "angola",
  "regulation": "BNA",
  "event": "ELEVATION_APPROVAL",
  "details": {
    "requestId": "req123",
    "approver": "security-officer-1",
    "justification": "Authorized for system maintenance as per ticket #4567",
    "scope": "docker:admin",
    "duration": 60,
    "resources": ["container-name:finance-app"],
    "emergency": false,
    "approval_time": "2025-08-06T18:40:22Z",
    "regulatory_tags": ["financial_system", "core_banking"]
  },
  "user_id": "user456",
  "tenant_id": "bank123",
  "client_ip": "192.168.1.100",
  "trace_id": "trace123",
  "retention_period": "7y"
}
```

### 3. Relatórios de Conformidade Automáticos

Relatórios automatizados para requisitos regulatórios específicos:

- Relatórios de operações privilegiadas para reguladores financeiros
- Relatórios de acesso a dados pessoais para autoridades de proteção de dados
- Relatórios de operações emergenciais para auditoria interna
- Relatórios de aprovações e rejeições para governança de TI
- Relatórios de exceções e violações de política para comitê de segurança

## Anexos Técnicos

1. Diagramas de Sequência para Integração
2. Modelo de Dados Completo
3. Contratos de API em OpenAPI 3.1
4. Esquemas gRPC Completos
5. Especificação de Eventos Kafka