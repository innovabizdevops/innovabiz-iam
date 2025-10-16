# Observabilidade para MCP-IAM Hooks

## Visão Geral

O módulo de observabilidade para hooks MCP-IAM da plataforma INNOVABIZ fornece uma camada unificada e extensível para instrumentação, monitoramento e auditoria de operações de identidade e acesso, integrando:

- **Métricas** (Prometheus)
- **Tracing** (OpenTelemetry)
- **Logging Estruturado** (Zap)

O adaptador de observabilidade é projetado para ser:

- **Multi-mercado**: Configurações específicas por mercado (Angola, Brasil, UE, China, etc.)
- **Multi-tenant**: Suporte a diferentes tipos de organizações (Financeiro, Varejo, Saúde, etc.)
- **Multi-contexto**: Adaptável a diversos tipos de hooks e operações IAM
- **Compliant**: Alinhado com regulações locais e normas internacionais

## Arquitetura

```
┌─────────────────────────────────────────────────────┐
│              Adaptador de Observabilidade           │
├─────────────────┬─────────────────┬─────────────────┤
│    Métricas     │     Tracing     │    Logging      │
│   (Prometheus)  │ (OpenTelemetry) │     (Zap)       │
├─────────────────┴─────────────────┴─────────────────┤
│             Contexto de Mercado & Tenant            │
├─────────────────────────────────────────────────────┤
│              Hooks MCP-IAM (Consumidores)           │
└─────────────────────────────────────────────────────┘
```

### Componentes Principais

1. **Hook Observability (`adapter/hook_observability.go`)**: Adaptador principal que unifica métricas, tracing e logging
2. **Market Context**: Contexto que encapsula informações de mercado, tenant, tipo de hook e requisitos regulatórios
3. **Metrics Interface**: Abstração para métricas Prometheus
4. **Tracing Interface**: Abstração para tracing OpenTelemetry
5. **Logger Interface**: Abstração para logging estruturado Zap
6. **Configurador de Observabilidade**: Mecanismo para configurar observabilidade específica por mercado

## Funcionalidades

### Observabilidade de Operações de Hook

- **ObserveHookOperation**: Instrumentação genérica para qualquer operação de hook
- **ObserveValidateScope**: Instrumentação específica para validação de escopo
- **ObserveValidateMFA**: Instrumentação específica para validação MFA
- **ObserveValidateToken**: Instrumentação específica para validação de token
- **ObserveGetApprovers**: Instrumentação específica para obtenção de aprovadores
- **ObserveGenerateAuditData**: Instrumentação específica para geração de dados de auditoria
- **ObserveCompleteElevation**: Instrumentação específica para conclusão de elevação

### Auditoria e Segurança

- **TraceAuditEvent**: Registro de eventos de auditoria com correlação via tracing
- **TraceSecurity**: Registro de eventos de segurança com severidade e correlação
- **UpdateActiveElevations**: Atualização de métricas de elevações ativas
- **RecordTestCoverage**: Registro de cobertura de testes para hooks
- **RegisterComplianceMetadata**: Registro de metadados de compliance

## Compliance e Regulações Suportadas

| Mercado      | Regulações                   | Retenção (Anos) | Aprovação Dual | Nível MFA |
|--------------|------------------------------|-----------------|----------------|-----------|
| Angola       | BNA, ISO 27001              | 7               | Sim            | Alto      |
| Brasil       | LGPD, BACEN                 | 5-10            | Sim            | Alto      |
| UE           | GDPR, PSD2                  | 7-10            | Sim            | Alto      |
| China        | PIPL                        | 5               | Sim            | Alto      |
| EUA          | SOX, HIPAA, PCI-DSS         | 5-7             | Variável       | Médio     |
| Moçambique   | Banco de Moçambique         | 5               | Sim            | Médio     |
| Global       | ISO 27001, NIST, COBIT      | 3               | Variável       | Médio     |

## Configuração

O adaptador de observabilidade é configurado através da estrutura `Config`:

```go
type Config struct {
    Environment           string // development, test, staging, production
    ServiceName           string // Nome do serviço
    OTLPEndpoint          string // Endpoint para OpenTelemetry
    MetricsPort           int    // Porta para métricas Prometheus
    ComplianceLogsPath    string // Caminho para logs de compliance
    EnableComplianceAudit bool   // Ativar auditoria de compliance
    StructuredLogging     bool   // Usar logging estruturado
    LogLevel              string // Nível de log (debug, info, warn, error)
}
```

## Uso Básico

### Inicialização do Adaptador

```go
// Configuração para Angola (Produção)
config := adapter.Config{
    Environment:           constants.EnvProduction,
    ServiceName:           "mcp-iam-hooks-angola",
    OTLPEndpoint:          "otel-collector:4317",
    MetricsPort:           9090,
    ComplianceLogsPath:    "/compliance/angola/bna",
    EnableComplianceAudit: true,
    StructuredLogging:     true,
    LogLevel:              "info",
}

// Criar adaptador de observabilidade
obs, err := adapter.NewHookObservability(config)
if err != nil {
    log.Fatalf("Falha ao configurar observabilidade: %v", err)
}

// Registrar metadados de compliance para Angola
obs.RegisterComplianceMetadata(
    constants.MarketAngola,
    "BNA",
    true, // Requer aprovação dual
    constants.MFALevelHigh,
    7, // 7 anos de retenção
)
```

### Uso em Hook de Elevação de Privilégios

```go
// Criar contexto de mercado
marketCtx := adapter.NewMarketContext(constants.MarketAngola, constants.TenantFinancial, constants.HookTypePrivilegeElevation)

// Observar validação de escopo
err = obs.ObserveValidateScope(
    ctx,
    marketCtx,
    userId,
    "admin:read",
    func(ctx context.Context) error {
        // Implementação de validação de escopo
        return validateScope(ctx, userId, "admin:read")
    },
)

// Observar validação MFA
err = obs.ObserveValidateMFA(
    ctx,
    marketCtx,
    userId,
    constants.MFALevelHigh,
    func(ctx context.Context) error {
        // Implementação de validação MFA
        return validateMFA(ctx, userId, constants.MFALevelHigh)
    },
)

// Registrar evento de auditoria
obs.TraceAuditEvent(
    ctx,
    marketCtx,
    userId,
    "privilege_elevation",
    "Elevação de privilégio aprovada para acesso admin temporário",
)
```

## Integração com Sistema de Observabilidade

O adaptador é projetado para integrar facilmente com a infraestrutura de observabilidade existente:

1. **Métricas Prometheus**: Expostas na porta configurada (padrão: 9090)
2. **Tracing OpenTelemetry**: Exportados para o endpoint OTLP configurado
3. **Logs**: Gravados em formato estruturado (JSON) nos caminhos especificados

## Cenários de Uso

### Hooks MCP-IAM

- **Elevação de privilégios**: Validação de escopo, MFA, aprovação dual e auditoria completa
- **Validação de token**: Verificação de tokens com tracing e métricas de desempenho
- **Validação MFA**: Verificação de fatores MFA com registro de eventos de segurança
- **Aprovação de elevação**: Fluxo de aprovação dual com auditoria detalhada

### Exportação de Métricas e Tracing

- **Dashboards Grafana**: Métricas de operações IAM, elevações ativas, cobertura de testes
- **Jaeger/Zipkin**: Visualização de traces para operações de hook completas
- **Elasticsearch**: Indexação de logs estruturados para consulta e análise

## Normas e Frameworks

O design e implementação do adaptador de observabilidade estão alinhados com:

- **ISO/IEC 27001**: Sistema de Gestão de Segurança da Informação
- **ISO 20000**: Gestão de Serviços de TI
- **COBIT 2019**: Objetivos de Controle para Informação e Tecnologias Relacionadas
- **TOGAF 10.0**: Arquitetura Empresarial
- **DMBOK 2.0**: Governança de Dados
- **NIST SP 800-53**: Controles de Segurança e Privacidade
- **ISO/IEC 29119**: Teste de Software
- **CMMI**: Modelo de Maturidade de Capacidade Integrado

## Exemplos Práticos

Consulte os exemplos detalhados:

1. **Configuração por Mercado**: `examples/mcp-hooks/observability_setup.go`
2. **Hook de Elevação de Privilégios**: `examples/mcp-hooks/privilege_elevation_hook.go`
3. **Testes Unitários**: `observability/adapter/tests/hook_observability_test.go`
4. **Testes de Integração**: `integration-tests/observability/hook_observability_integration_test.go`

## Requisitos Não-Funcionais

O adaptador de observabilidade atende aos seguintes requisitos não-funcionais:

- **Desempenho**: Sobrecarga mínima em operações críticas de hook (< 5ms)
- **Escalabilidade**: Suporte a alto volume de operações concorrentes
- **Resiliência**: Falhas na observabilidade não afetam operações principais
- **Segurança**: Logs sensíveis são protegidos e criptografados
- **Manutenibilidade**: Design modular com interfaces bem definidas
- **Testabilidade**: Cobertura de testes > 90%
- **Conformidade**: Adaptável a requisitos regulatórios específicos

## Próximos Passos

1. **Integração com Serviços de Alerta**: Notificações automáticas para eventos críticos
2. **Métricas de Business Intelligence**: KPIs específicos para operações IAM
3. **Configuração Dinâmica**: Ajuste de níveis de log e sampling em tempo real
4. **Dashboards Predefinidos**: Templates Grafana específicos por mercado
5. **Testes de Caos**: Validação de resiliência do sistema de observabilidade

## Referências

- [Prometheus Best Practices](https://prometheus.io/docs/practices/naming/)
- [OpenTelemetry Specification](https://opentelemetry.io/docs/reference/specification/)
- [Zap Logger](https://github.com/uber-go/zap)
- [OWASP Security Logging](https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html)