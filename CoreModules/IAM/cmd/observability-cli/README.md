# CLI de Observabilidade MCP-IAM

## 🚀 Visão Geral

A CLI de Observabilidade MCP-IAM é uma ferramenta especializada para gerenciar, configurar e testar a instrumentação de observabilidade nos hooks MCP-IAM da plataforma INNOVABIZ, com suporte multi-mercado, multi-tenant e multi-contexto conforme normativas internacionais e requisitos de compliance.

## 📋 Funcionalidades

- **Configuração Multi-Mercado**: Adaptação automática para requisitos específicos de Angola, Brasil, UE, EUA, China, SADC e outros mercados
- **Simulação de Operações**: Testes de validação de escopo, MFA e eventos de segurança/auditoria
- **Exportação de Telemetria**: Integração com coletores OpenTelemetry
- **Exposição de Métricas**: Servidor HTTP para métricas Prometheus
- **Logs de Compliance**: Registro de eventos em formato auditável por mercado

## 🛠️ Requisitos

- Go 1.21 ou superior
- Acesso ao diretório de configuração para logs de compliance
- Opcional: Coletor OpenTelemetry (para tracing distribuído)
- Opcional: Prometheus (para scraping de métricas)

## ⚙️ Instalação

```bash
# Compilar a CLI
cd CoreModules/IAM
go build -o bin/observability-cli ./cmd/observability-cli

# Adicionar ao PATH (opcional)
# Para Windows: Adicione o caminho completo à variável PATH
# Para Linux/MacOS: cp bin/observability-cli /usr/local/bin/
```

## 📚 Comandos Disponíveis

### Configuração

```bash
# Mostrar configuração atual
observability-cli config show

# Validar configuração
observability-cli config validate
```

### Testes

```bash
# Simular operações de hook
observability-cli test hook-operations --market Brazil --tenant-type Financial --count 10

# Testar exportação de traces
observability-cli test trace-export --otlp-endpoint localhost:4317
```

### Métricas

```bash
# Expor métricas em servidor HTTP
observability-cli metrics expose --metrics-port 9090
```

## 🌐 Configuração por Mercado

A CLI suporta configurações específicas por mercado através de flags:

| Mercado | Frameworks Suportados | Níveis MFA | Retenção de Logs |
|---------|----------------------|------------|-----------------|
| Angola | BNA | Alto | 7 anos |
| Brasil | LGPD, BACEN | Alto | 5-10 anos |
| UE | GDPR | Alto | 7 anos |
| EUA | SOX | Médio | 7 anos |
| Global | ISO27001 | Médio | 3 anos |

## 📊 Métricas Disponíveis

- `innovabiz_iam_hook_calls_total`: Total de chamadas de hook por mercado/tenant/tipo
- `innovabiz_iam_hook_errors_total`: Total de erros de hook por mercado/tenant/tipo
- `innovabiz_iam_hook_duration_seconds`: Tempo de execução de hooks em segundos
- `innovabiz_iam_active_elevations`: Elevações de privilégio ativas por mercado
- `innovabiz_iam_mfa_validations_total`: Validações MFA por nível e resultado
- `innovabiz_iam_scope_validations_total`: Validações de escopo por resultado
- `innovabiz_iam_test_coverage_percent`: Percentual de cobertura de testes
- `innovabiz_iam_compliance_events_total`: Eventos de compliance por framework
- `innovabiz_iam_security_events_total`: Eventos de segurança por severidade

## 🔍 Exemplos de Uso

### Testar em Ambiente de Desenvolvimento com Mercado Específico

```bash
observability-cli test hook-operations \
  --market Brazil \
  --tenant-type Financial \
  --hook-type PrivilegeElevation \
  --count 5 \
  --delay 200 \
  --environment development \
  --structured-logging true \
  --logs-path ./logs/compliance
```

### Exportar Traces para Coletor OpenTelemetry

```bash
observability-cli test trace-export \
  --market Angola \
  --otlp-endpoint localhost:4317 \
  --service-name "mcp-iam-hooks-service"
```

### Expor Métricas para Prometheus

```bash
observability-cli metrics expose \
  --metrics-port 9090 \
  --service-name "mcp-iam-hooks-metrics"
```

## 🔒 Compliance e Segurança

- Todos os eventos são registrados em formato auditável conforme requisitos regulatórios
- Suporte a níveis de MFA específicos por mercado
- Metadados de compliance por operação
- Rastreamento completo de operações para investigação e auditoria

## 📝 Normas e Frameworks Suportados

- **Segurança**: ISO/IEC 27001, ISO 27018, PCI DSS, SOX
- **Gestão**: COBIT 2019, ITIL 4.0, ISO 20000
- **Privacidade**: GDPR, LGPD, CCPA
- **Financeiro**: SOX, PSD2, Open Banking, BNA
- **Arquitetura**: TOGAF 10.0, DMBOK 2.0

## ⚠️ Observações

- Em ambientes de produção, configure o endpoint OTLP com TLS
- Defina níveis apropriados de retenção de logs conforme requisitos regulatórios
- Monitore o crescimento dos logs de compliance para evitar problemas de armazenamento

## 🔗 Integração

Esta CLI é parte do ecossistema de observabilidade da plataforma INNOVABIZ e integra-se nativamente com:

- Coletores OpenTelemetry
- Prometheus
- Grafana
- Jaeger/Zipkin
- Elasticsearch/Kibana