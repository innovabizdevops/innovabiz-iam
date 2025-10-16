# Documentação Técnica: Testes de Hooks MCP-IAM

**Versão:** 1.0.0  
**Data:** 2023-06-12  
**Autor:** INNOVABIZ DevOps  
**Classificação:** Interno  
**Mercados:** Angola, Brasil, União Europeia, China, Moçambique, BRICS, EUA, SADC, PALOP  
**Tenants:** Financeiro, Governo, Saúde, Varejo, Telecomunicações, Educação, Energia  

## 1. Visão Geral

Esta documentação descreve a arquitetura, implementação, execução e monitoramento dos testes para os hooks MCP-IAM (Model Context Protocol - Identity and Access Management) da plataforma INNOVABIZ. Os hooks implementam mecanismos de elevação de privilégios com validação contextual integrada aos protocolos de identidade e acesso para aplicações integradas (Docker, GitHub, Figma e outras).

### 1.1 Alinhamento com Frameworks e Normas

Os testes foram desenvolvidos seguindo os princípios e normas:
- **Arquitetura Empresarial:** TOGAF 10.0, IEEE 1471, ISO/IEC/IEEE 42010
- **Governança:** COBIT 2019, ISO/IEC 38500, ISO 27001, BIAN
- **Qualidade:** ISO 9001, ISO 25010, CMMI
- **Segurança:** ISO/IEC 27001, NIST Cybersecurity Framework, PCI DSS
- **Privacidade:** GDPR, LGPD, CCPA, PIPEDA, KVKK
- **Auditoria:** IPPF, IIA, COSO

### 1.2 Camadas de Testes

| Tipo de Teste | Objetivo | Ferramentas | Métricas | Normas Aplicadas |
|---------------|----------|------------|----------|------------------|
| Unitário | Validar componentes individuais | Testify, Go Testing | Cobertura >85%, Latência <5ms | ISO 25010, ISO 9001 |
| Integração | Validar fluxos completos | Testify, Mockery | Cobertura >80%, Sucesso >95% | TOGAF, ISO/IEC 42010 |
| Performance | Validar comportamento sob carga | Go Benchmarks, K6 | Latência <5ms, Throughput >1000 req/s | ISO 25010, NIST SP 800-53 |
| Conformidade | Validar requisitos regulatórios | Mocks específicos por mercado | 100% conformidade com requisitos | GDPR, LGPD, ISO 27001 |
| E2E | Validar sistema completo | Kubernetes, Helm | Disponibilidade >99.9%, Latência <200ms | ISO 25010, ISO 9241 |

## 2. Estrutura dos Testes

### 2.1 Organização dos Diretórios

```
CoreModules/IAM/
├── tests/
│   ├── hooks/
│   │   ├── figma_hook_test.go     # Testes unitários do Figma Hook
│   │   ├── docker_hook_test.go    # Testes unitários do Docker Hook
│   │   ├── github_hook_test.go    # Testes unitários do GitHub Hook
│   │   ├── integration_test.go    # Testes de integração entre hooks
│   │   └── README.md              # Documentação específica dos testes
│   ├── performance/
│   │   ├── hooks_performance_test.go  # Testes de performance para hooks
│   │   └── benchmarks/              # Benchmarks específicos por hook
│   └── e2e/
│       └── elevation_flows_test.go  # Testes de fluxos completos
├── .github/workflows/
│   └── mcp-iam-hooks-tests.yml    # Pipeline CI/CD para testes
└── infrastructure/
    └── grafana/
        └── dashboards/
            └── mcp-iam-hooks-monitoring.json  # Dashboard de monitoramento
```

### 2.2 Tipos de Testes Implementados

#### 2.2.1 Testes Unitários

Os testes unitários validam o comportamento isolado de cada hook MCP-IAM, com mocks para simular dependências externas:

- **Validação de Escopo**: Testa se o hook valida corretamente os escopos solicitados conforme regras específicas de cada mercado e tenant.
- **Requisitos MFA**: Verifica se o hook aplica corretamente os requisitos de MFA conforme níveis definidos por mercado.
- **Obtenção de Aprovadores**: Valida a lógica de seleção de aprovadores baseada em escopo e mercado.
- **Validação de Token**: Testa a validação de tokens considerando expiração, escopo e restrições específicas.
- **Metadados de Auditoria**: Verifica a geração correta de metadados de auditoria por mercado.

#### 2.2.2 Testes de Integração

Os testes de integração validam fluxos completos de elevação de privilégios em diferentes mercados:

- **Fluxo Angola**: Validação de dupla aprovação com MFA forte.
- **Fluxo UE**: Verificação de conformidade GDPR com requisitos específicos de logs.
- **Fluxo Brasil**: Cenário de rejeição para escopo restrito com validação LGPD.
- **Fluxo Moçambique**: Cenário de uso de token expirado com regras SADC.

#### 2.2.3 Testes de Performance

Os testes de performance medem a eficiência dos hooks sob diferentes cargas:

- **Latência**: Tempo de resposta médio em milissegundos para validação de escopo e token.
- **Throughput**: Capacidade de processamento de solicitações por segundo.
- **Escalabilidade Multi-tenant**: Comportamento sob múltiplos tenants simultâneos.
- **Consumo de Recursos**: Utilização de CPU, memória e I/O durante operações.

## 3. Mocks e Simulações

### 3.1 Componentes Mockados

| Componente | Descrição | Implementação |
|------------|-----------|--------------|
| `ComplianceMetadataProvider` | Fornece metadados específicos de compliance por mercado | Interface com implementações por mercado |
| `MFAProvider` | Simula verificações MFA com diferentes níveis de segurança | Interface com métodos `Validate` e `GetMFALevel` |
| `ApproverService` | Fornece listas de aprovadores conforme contexto | Interface com métodos `GetApprovers` por escopo |
| `AuditService` | Registra eventos de auditoria e valida conformidade | Interface com métodos de log e verificação |
| `TokenStore` | Gerencia armazenamento e validação de tokens | Interface com métodos CRUD para tokens |

### 3.2 Metadados de Compliance por Mercado

Os mocks simulam diferentes requisitos regulatórios por mercado:

```go
// Exemplo de metadados de compliance para Angola
angolaCompliance := &mocks.ComplianceMetadata{
    RequiresDualApproval:    true,
    RequiresMFA:            "strong",
    RetentionYears:         7,
    DataSovereignty:        "angola",
    Framework:              []string{"BNA", "ARSSI", "ISO27001"},
    AuditRequirements:      []string{"full_audit_trail", "dual_approval_logs"},
    ProhibitedScopes:       []string{"admin:full_access"},
    MaxElevationMinutes:    60,
}
```

## 4. Cobertura Multi-Mercado e Multi-Tenant

### 4.1 Dimensão Multi-Mercado

| Mercado | Regulações | Requisitos Específicos | Implementação |
|---------|------------|------------------------|---------------|
| Angola | BNA, ARSSI | Dupla aprovação, MFA forte, auditoria completa | Testes específicos para validação BNA |
| Brasil | LGPD, BACEN | Consentimento explícito, direito ao esquecimento | Validação de logs LGPD-compliant |
| União Europeia | GDPR, EBA | Minimização de dados, direito à portabilidade | Verificação de retenção e privacidade |
| China | PIPL, CSL | Localização de dados, aprovações governamentais | Testes para restrições de escopo |
| Moçambique | LPDP, BM | Requisitos SADC, retenção 5 anos | Validação de fluxos específicos |
| BRICS | Múltiplas | Interoperabilidade entre membros | Testes de conformidade cruzada |
| EUA | SOX, GLBA, CCPA | Divulgação financeira, privacidade | Testes específicos para EUA |

### 4.2 Dimensão Multi-Tenant

| Tipo de Tenant | Características | Requisitos Específicos | Implementação |
|----------------|----------------|------------------------|---------------|
| Financeiro | Alta segurança | MFA forte, múltiplas aprovações, auditoria | Mocks específicos bancários |
| Governo | Soberania de dados | Validação jurisdicional, retenção longa | Testes com restrições governamentais |
| Saúde | Dados sensíveis | PHI, confidencialidade, consentimento | Verificações específicas de saúde |
| Varejo | Alta escala | Performance, segurança de transações | Testes de carga para varejo |
| Telecomunicações | Infraestrutura crítica | Redundância, conformidade setorial | Validações específicas do setor |

## 5. Observabilidade e Monitoramento

### 5.1 Métricas Expostas

O sistema expõe métricas Prometheus para monitoramento em tempo real:

| Métrica | Tipo | Rótulos | Descrição |
|---------|------|--------|-----------|
| `iam_hook_elevation_requests_total` | Counter | `environment`, `market`, `tenant_type`, `hook_type` | Total de solicitações de elevação |
| `iam_hook_elevation_requests_rejected_total` | Counter | `environment`, `market`, `tenant_type`, `hook_type`, `reason` | Solicitações rejeitadas |
| `iam_hook_validation_duration_milliseconds` | Histogram | `environment`, `market`, `tenant_type`, `hook_type` | Tempo de validação |
| `iam_hook_test_coverage` | Gauge | `environment`, `hook_type` | Cobertura de testes por hook |
| `iam_hook_compliance_checks_total` | Counter | `environment`, `market`, `compliance_type` | Verificações de compliance |
| `iam_hook_mfa_checks_total` | Counter | `environment`, `market`, `mfa_level`, `success` | Verificações MFA |

### 5.2 Dashboard de Monitoramento

O dashboard Grafana `mcp-iam-hooks-monitoring` fornece visualização em tempo real:

- **Visão Geral**: Total de solicitações, tempo médio de validação, distribuição por mercado
- **Performance**: Latência por hook, taxa de rejeição, throughput
- **Multi-Mercado**: Distribuição de solicitações por mercado e tipo de tenant
- **Compliance**: Verificações de compliance por tipo, metadados por mercado
- **Alertas**: Configurados para latência > 5ms, taxa de rejeição > 5%, falhas > 1%

### 5.3 Tracing Distribuído

A instrumentação com OpenTelemetry permite tracing de ponta a ponta:

```go
// Exemplo de instrumentação para validação de escopo
func (h *FigmaHook) ValidateScope(ctx context.Context, scope string) error {
    ctx, span := tracer.Start(ctx, "FigmaHook.ValidateScope", 
        trace.WithAttributes(
            attribute.String("scope", scope),
            attribute.String("market", h.market),
            attribute.String("tenant_type", h.tenantType),
        ))
    defer span.End()
    
    // Lógica de validação...
    
    // Registrar resultado no span
    if err != nil {
        span.SetStatus(codes.Error, err.Error())
        span.RecordError(err)
        return err
    }
    
    return nil
}
```

## 6. Pipeline CI/CD

### 6.1 Workflow GitHub Actions

O arquivo `.github/workflows/mcp-iam-hooks-tests.yml` configura a integração contínua:

- **Lint**: Verificação de qualidade de código com `golangci-lint`
- **Testes Unitários**: Execução paralela por mercado com cobertura mínima de 85%
- **Testes de Performance**: Validação de limites de latência (5ms)
- **Testes de Integração**: Execução com serviços reais (PostgreSQL, Redis)
- **Testes E2E**: Implantação em cluster K3D e validação completa
- **Publicação**: Atualização de dashboards e notificações

### 6.2 Matrizes de Execução

Os testes são executados em matrizes para validar diferentes combinações:

```yaml
strategy:
  matrix:
    market: [angola, brasil, eu, china, mocambique, brics, eua, sadc, palop]
```

### 6.3 Thresholds e Verificações Automáticas

- **Cobertura**: Mínimo 85% de cobertura de código
- **Performance**: Latência máxima de 5ms por operação
- **Vulnerabilidades**: Análise com `govulncheck`
- **Qualidade**: Análise SonarQube

## 7. Executando os Testes

### 7.1 Testes Unitários

```bash
# Executar todos os testes unitários
cd CoreModules/IAM
go test -v -race ./tests/hooks/...

# Executar testes para um mercado específico
MARKET=angola go test -v -race ./tests/hooks/...

# Verificar cobertura
go test -coverprofile=coverage.out ./tests/hooks/...
go tool cover -html=coverage.out
```

### 7.2 Testes de Performance

```bash
# Executar testes de performance
cd CoreModules/IAM
go test -v -bench=. ./tests/performance/...

# Executar com threshold específico
THRESHOLD_MS=5 go test -v ./tests/performance/... -tags=performance
```

### 7.3 Testes de Integração

```bash
# Configurar ambiente
docker-compose up -d postgres redis

# Executar testes de integração
DATABASE_URL="postgres://postgres:postgres@localhost:5432/iam_test?sslmode=disable" \
REDIS_URL="redis://localhost:6379/0" \
MARKETS="angola,brasil,eu,china,mocambique" \
go test -v ./tests/hooks/integration_test.go
```

## 8. Extensibilidade e Manutenção

### 8.1 Adicionando Novos Hooks

1. Crie um arquivo de teste na pasta `tests/hooks/` seguindo o padrão existente
2. Implemente mocks específicos no pacote `mocks/`
3. Adicione métricas Prometheus no arquivo `metrics.go`
4. Atualize o dashboard Grafana para incluir o novo hook
5. Adicione testes de performance específicos

### 8.2 Adicionando Novo Mercado

1. Crie implementações de mock específicas para o mercado
2. Adicione constantes de mercado no arquivo `constants.go`
3. Adicione casos de teste para o novo mercado
4. Configure o pipeline CI/CD para incluir o mercado na matriz
5. Atualize a documentação com regulações específicas

## 9. Integração com Outras Ferramentas

### 9.1 Integração com Sistemas de Observabilidade

- **Prometheus**: Exportação de métricas via endpoint `/metrics`
- **Grafana**: Dashboards pré-configurados para monitoramento
- **OpenTelemetry**: Tracing distribuído para todos os fluxos
- **Zap**: Logging estruturado para análise e auditoria

### 9.2 Integração com Sistemas de Segurança

- **OPA (Open Policy Agent)**: Validação de políticas
- **Vault**: Armazenamento seguro de segredos para testes
- **Falco**: Monitoramento de comportamento anormal
- **Sysdig**: Captura e análise de eventos para auditoria

## 10. Próximos Passos e Roadmap

### 10.1 Próximas Implementações (Q3 2023)

- Testes de caos (Chaos Engineering)
- Expansão para mais hooks (JIRA, Slack, AWS)
- Automatização de testes de regressão regulatória
- Dashboards para compliance contínuo

### 10.2 Roadmap de Longo Prazo (2024)

- Implementação de IA para análise de padrões suspeitos
- Testes adaptativos baseados em comportamento histórico
- Simulação de ataques e avaliação de resiliência
- Integração com frameworks de IAC (Infrastructure as Code)

## 11. Glossário

| Termo | Descrição |
|------|-----------|
| **MCP** | Model Context Protocol - Protocolo de comunicação entre modelos de IA e sistemas externos |
| **Hook** | Ponto de integração para validação e autorização de elevação de privilégios |
| **Tenant** | Entidade organizacional isolada dentro da plataforma multi-tenant |
| **Elevação** | Processo de obtenção temporária de privilégios elevados |
| **MFA** | Multi-Factor Authentication - Autenticação por múltiplos fatores |
| **BRICS** | Brasil, Rússia, Índia, China e África do Sul - Bloco econômico |
| **SADC** | Southern African Development Community - Comunidade de Desenvolvimento da África Austral |
| **PALOP** | Países Africanos de Língua Oficial Portuguesa |

---

## Apêndice A: Configuração de Métricas Prometheus

```go
// metrics.go - Exemplo de configuração de métricas
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // ElevationRequestsTotal conta o número total de solicitações de elevação
    ElevationRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "iam_hook_elevation_requests_total",
            Help: "Total number of elevation requests processed by IAM hooks",
        },
        []string{"environment", "market", "tenant_type", "hook_type"},
    )
    
    // ValidationDuration mede o tempo de validação
    ValidationDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "iam_hook_validation_duration_milliseconds",
            Help:    "Duration of validation operations in milliseconds",
            Buckets: prometheus.ExponentialBuckets(0.1, 2, 10), // 0.1ms a ~100ms
        },
        []string{"environment", "market", "tenant_type", "hook_type"},
    )
)
```

## Apêndice B: Exemplo de Configuração de Tracing

```go
// tracing.go - Exemplo de configuração de OpenTelemetry
package tracing

import (
    "context"
    "log"
    
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func InitTracer(serviceName string) func() {
    ctx := context.Background()
    
    res, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceNameKey.String(serviceName),
        ),
    )
    if err != nil {
        log.Fatalf("failed to create resource: %v", err)
    }
    
    // Configurar exportador OTLP
    client := otlptracegrpc.NewClient(
        otlptracegrpc.WithInsecure(),
        otlptracegrpc.WithEndpoint("otel-collector:4317"),
    )
    exporter, err := otlptrace.New(ctx, client)
    if err != nil {
        log.Fatalf("failed to create exporter: %v", err)
    }
    
    // Configurar provedor de tracer
    bsp := sdktrace.NewBatchSpanProcessor(exporter)
    tracerProvider := sdktrace.NewTracerProvider(
        sdktrace.WithSampler(sdktrace.AlwaysSample()),
        sdktrace.WithResource(res),
        sdktrace.WithSpanProcessor(bsp),
    )
    
    // Configurar propagador
    otel.SetTracerProvider(tracerProvider)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))
    
    return func() {
        if err := tracerProvider.Shutdown(ctx); err != nil {
            log.Fatalf("failed to shutdown tracer provider: %v", err)
        }
    }
}
```

---

*Este documento é uma propriedade intelectual da INNOVABIZ e está protegido por direitos autorais. Todos os direitos reservados.*