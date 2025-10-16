# Módulo de Observabilidade do INNOVABIZ IAM Audit Service

Este módulo implementa recursos abrangentes de observabilidade para o IAM Audit Service da plataforma INNOVABIZ, conforme definido nos seguintes ADRs:

- **ADR-005**: Framework de Decoradores e Middleware para Instrumentação Automática
- **ADR-006**: Health Checks e Endpoints de Diagnóstico
- **ADR-007**: Sistema de Alertas Prometheus e Dashboards Grafana
- **ADR-008**: Integração de Rastreamento Distribuído com OpenTelemetry

## Recursos Principais

- **Métricas Prometheus**: Instrumentação automática de APIs FastAPI, eventos de auditoria, verificações de compliance e operações de retenção
- **Health Checks**: Endpoints padronizados para health, readiness e liveness com suporte multi-tenant e multi-região
- **Diagnósticos**: Endpoint detalhado de diagnóstico para administradores com verificação profunda de dependências
- **Rastreamento Distribuído**: Integração com OpenTelemetry para rastreamento de operações entre serviços
- **Suporte Multi-contexto**: Isolamento completo de dados por tenant e região em todas as métricas e spans
- **Decoradores**: Framework de decoradores para instrumentação consistente e simplificada de código de negócio

## Instalação

```bash
pip install -r requirements-observability.txt
```

## Guia de Uso Rápido

### Integração com FastAPI

```python
from fastapi import FastAPI
from src.observability import configure_observability

# Cria a aplicação FastAPI
app = FastAPI()

# Configura observabilidade com um único comando
obs = configure_observability(
    app,
    service_name="iam-audit-service",
    service_version="1.0.0",
    default_tenant="innovabiz",
    default_region="global"
)
```

### Instrumentação de Eventos de Auditoria

```python
from src.observability import instrument_audit_event

class AuditService:
    @instrument_audit_event(event_type="login")
    async def process_login_event(self, event, tenant, region):
        # Seu código aqui
        pass
```

### Instrumentação de Funções com Métricas

```python
from src.observability import instrument_function

# Usa decorador para instrumentar uma função com histograma de duração
@instrument_function(
    histogram=metrics.compliance_check_seconds,
    labels={"compliance_type": "gdpr"},
    extract_labels_from_args={"tenant": "tenant_id"}
)
async def check_gdpr_compliance(tenant_id, record_id):
    # Seu código aqui
    pass
```

### Rastreamento Distribuído

```python
from src.observability import traced

# Usa decorador para rastreamento distribuído
@traced(
    name="retention.policy.execute",
    attributes={"policy_type": "gdpr"}
)
async def execute_gdpr_retention(tenant, region):
    # Seu código aqui
    pass
```

### Health Checks Customizados

```python
from src.observability import HealthChecker, HealthStatus, ComponentCheckResult

# Obtém o health checker da integração
health_checker = obs.health

# Registra um verificador customizado
async def check_external_service(tenant, region):
    # Verificação personalizada
    return ComponentCheckResult(
        status=HealthStatus.HEALTHY,
        description="External service is working",
        latency_ms=42.5
    )

health_checker.register_dependency_checker(
    "external_service", 
    check_external_service,
    {
        "name": "External API Service",
        "type": "external_api"
    }
)
```

## Endpoints de Observabilidade

A integração expõe os seguintes endpoints:

- **GET /metrics**: Métricas no formato Prometheus
- **GET /health**: Health check básico (verificações rápidas)
- **GET /ready**: Readiness check para decisões de roteamento
- **GET /live**: Liveness check simples para verificação de processo
- **GET /diagnostic**: Diagnóstico detalhado com estado de dependências

## Contexto Multi-tenant e Multi-região

O módulo dá suporte total à propagação de contexto multi-tenant e multi-regional:

- Via headers HTTP: `X-Tenant-ID`, `X-Region` e `X-Environment`
- Automaticamente capturados e propagados em métricas, health checks e spans
- Isolamento de dados por tenant em todas as visualizações e alertas

## Exemplos

Veja o diretório `examples/` para implementações completas:

- `fastapi_app.py`: Aplicação FastAPI completa com instrumentação
- `custom_health_checks.py`: Como adicionar verificações de saúde customizadas
- `prometheus_rules.yml`: Exemplos de regras de alertas para Prometheus

## Estrutura do Código

```
src/observability/
├── __init__.py           # Exporta as APIs públicas
├── config.py             # Configurações baseadas em Pydantic
├── integration.py        # Classe central de integração
├── metrics.py            # Gerenciamento de métricas Prometheus
├── health.py             # Health checks e diagnósticos
└── tracing.py            # Integração OpenTelemetry
```

## Integração com Infraestrutura

### Prometheus e Grafana

As métricas expostas podem ser coletadas pelo Prometheus e visualizadas no Grafana. Os dashboards estão disponíveis em:

```
config/observability/grafana-dashboards/
├── audit_service_dashboard.json     # Dashboard principal
├── audit_compliance_dashboard.json  # Dashboard de compliance
└── audit_alerts_dashboard.json      # Dashboard de alertas
```

### OpenTelemetry Collector

Para tracing distribuído, configure o OpenTelemetry Collector para coletar e exportar spans:

```yaml
# Exemplo de config para otel-collector
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

processors:
  batch:
    timeout: 1s
    send_batch_size: 1024

exporters:
  jaeger:
    endpoint: jaeger:14250
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [jaeger]
```

## Boas Práticas

1. **Use labels com moderação**: Evite alta cardinalidade em métricas
2. **Nomeie spans adequadamente**: Use convenção `service.operation.sub_operation`
3. **Implemente health checks eficientes**: Mantenha verificações leves (<100ms)
4. **Configure alertas adequadamente**: Evite alertas excessivos
5. **Utilize decoradores para consistência**: Padronize a instrumentação

## Configuração Avançada

Para personalizar a configuração, você pode usar variáveis de ambiente com prefixo `OBSERVABILITY_`:

```bash
# Configuração de métricas
export OBSERVABILITY_SERVICE_NAME="iam-audit-service"
export OBSERVABILITY_DEFAULT_TENANT="innovabiz"
export OBSERVABILITY_METRICS__NAMESPACE="innovabiz"
export OBSERVABILITY_METRICS__SUBSYSTEM="iam_audit"

# Configuração de tracing
export OBSERVABILITY_TRACING__ENABLED=true
export OBSERVABILITY_TRACING__OTLP_ENDPOINT="http://otel-collector:4317"
export OBSERVABILITY_TRACING__SAMPLE_RATIO=0.5
```

## Licenciamento

Este módulo é parte da Suíte INNOVABIZ e está sujeito às políticas de licenciamento definidas pela INNOVABIZ. Uso exclusivamente interno e confidencial.

## Desenvolvido por

INNOVABIZ DevOps Team - 2025