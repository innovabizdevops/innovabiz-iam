# ADR-004: Padrões de Métricas para Observabilidade do IAM Audit Service

## Status

Aprovado

## Data

2025-07-31

## Contexto

Para garantir uma observabilidade completa e padronizada do IAM Audit Service, é necessário estabelecer padrões claros para as métricas coletadas e expostas. Estes padrões devem:

- Seguir as convenções da plataforma INNOVABIZ
- Atender aos requisitos específicos de um serviço de auditoria IAM
- Suportar contextos múltiplos (tenant, região, ambiente)
- Facilitar a detecção de problemas e anomalias
- Garantir compatibilidade com Prometheus e Grafana
- Permitir agregações e análises significativas
- Atender requisitos regulatórios e de compliance
- Balancear granularidade com performance e custo de armazenamento

## Decisão

Adotar um conjunto padronizado de métricas para observabilidade do IAM Audit Service, organizadas em categorias funcionais e coletadas via instrumentação automática e manual.

### 1. Padrão de Nomenclatura

Adotar o padrão `domain_component_unit_suffix` para todas as métricas:

- **Domain**: `iam_audit`
- **Component**: Funcionalidade específica (ex: `event`, `retention`, `http`, `compliance`)
- **Unit**: Unidade sendo medida (ex: `requests`, `errors`, `latency`, `size`)
- **Suffix**: Tipo de métrica (ex: `total`, `seconds`, `bytes`, `ratio`)

Exemplos:
- `iam_audit_event_processed_total`
- `iam_audit_retention_purge_seconds`
- `iam_audit_http_request_duration_seconds`

### 2. Categorias de Métricas

#### 2.1. Métricas de Eventos de Auditoria

```python
# Total de eventos de auditoria processados
iam_audit_event_processed_total = Counter(
    "iam_audit_event_processed_total",
    "Total de eventos de auditoria processados",
    ["tenant", "region", "event_type", "severity"]
)

# Latência no processamento de eventos
iam_audit_event_processing_seconds = Histogram(
    "iam_audit_event_processing_seconds",
    "Latência no processamento de eventos de auditoria",
    ["tenant", "region", "event_type"],
    buckets=[0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0]
)

# Tamanho dos eventos de auditoria
iam_audit_event_size_bytes = Histogram(
    "iam_audit_event_size_bytes",
    "Tamanho dos eventos de auditoria em bytes",
    ["tenant", "region", "event_type"],
    buckets=[64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384]
)
```

#### 2.2. Métricas de Retenção e Armazenamento

```python
# Total de eventos expurgados por política de retenção
iam_audit_retention_purge_total = Counter(
    "iam_audit_retention_purge_total",
    "Total de eventos expurgados por política de retenção",
    ["tenant", "region", "retention_policy"]
)

# Duração da execução de políticas de retenção
iam_audit_retention_execution_seconds = Histogram(
    "iam_audit_retention_execution_seconds",
    "Duração da execução de políticas de retenção",
    ["tenant", "region", "retention_policy"],
    buckets=[0.1, 0.5, 1.0, 5.0, 10.0, 30.0, 60.0, 300.0, 600.0]
)

# Volume de armazenamento por tenant/região
iam_audit_storage_used_bytes = Gauge(
    "iam_audit_storage_used_bytes",
    "Volume de armazenamento utilizado",
    ["tenant", "region"]
)
```

#### 2.3. Métricas de Compliance e Verificação

```python
# Total de verificações de compliance
iam_audit_compliance_check_total = Counter(
    "iam_audit_compliance_check_total",
    "Total de verificações de compliance realizadas",
    ["tenant", "region", "compliance_type", "status"]
)

# Duração das verificações de compliance
iam_audit_compliance_check_seconds = Histogram(
    "iam_audit_compliance_check_seconds",
    "Duração das verificações de compliance",
    ["tenant", "region", "compliance_type"],
    buckets=[0.01, 0.05, 0.1, 0.5, 1.0, 5.0, 10.0, 30.0, 60.0]
)

# Número de violações de compliance detectadas
iam_audit_compliance_violation_total = Counter(
    "iam_audit_compliance_violation_total",
    "Número de violações de compliance detectadas",
    ["tenant", "region", "compliance_type", "severity"]
)
```

#### 2.4. Métricas HTTP e API

```python
# Total de requisições HTTP
iam_audit_http_request_total = Counter(
    "iam_audit_http_request_total",
    "Total de requisições HTTP",
    ["tenant", "region", "method", "path", "status_code"]
)

# Duração das requisições HTTP
iam_audit_http_request_duration_seconds = Histogram(
    "iam_audit_http_request_duration_seconds",
    "Duração das requisições HTTP",
    ["tenant", "region", "method", "path"],
    buckets=[0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0]
)

# Tamanho das requisições e respostas
iam_audit_http_request_size_bytes = Histogram(
    "iam_audit_http_request_size_bytes",
    "Tamanho das requisições HTTP",
    ["tenant", "region", "method", "path"],
    buckets=[64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384]
)

iam_audit_http_response_size_bytes = Histogram(
    "iam_audit_http_response_size_bytes",
    "Tamanho das respostas HTTP",
    ["tenant", "region", "method", "path", "status_code"],
    buckets=[64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384]
)
```

#### 2.5. Métricas de Recursos e Dependências

```python
# Estado de saúde das dependências
iam_audit_dependency_health = Gauge(
    "iam_audit_dependency_health",
    "Estado de saúde das dependências (1=healthy, 0=unhealthy)",
    ["tenant", "region", "dependency_name", "dependency_type"]
)

# Latência das operações de dependências
iam_audit_dependency_latency_seconds = Histogram(
    "iam_audit_dependency_latency_seconds",
    "Latência das operações de dependências",
    ["tenant", "region", "dependency_name", "operation"],
    buckets=[0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0]
)

# Utilização de recursos
iam_audit_resource_utilization_ratio = Gauge(
    "iam_audit_resource_utilization_ratio",
    "Taxa de utilização de recursos (0-1)",
    ["tenant", "region", "resource_type"]
)
```

### 3. Dimensões Obrigatórias (Labels)

Todas as métricas devem incluir as seguintes dimensões mínimas:

1. **tenant**: Identificador do tenant
2. **region**: Região geográfica (ex: br-east, us-east, eu-west, ao-central)

Dimensões adicionais específicas para cada categoria de métrica podem ser adicionadas conforme necessário.

### 4. Estratégia de Instrumentação

1. **Instrumentação Automática**:
   - Middleware FastAPI para métricas HTTP
   - Decoradores para funções críticas
   - Hooks de ciclo de vida da aplicação

2. **Instrumentação Manual**:
   - Pontos críticos do código que requerem métricas específicas
   - Operações de longa duração ou alto impacto

3. **Padrões de Uso**:
   ```python
   # Exemplo de uso de decorador para instrumentação automática
   @metrics.instrument_audit_event(event_type="user_login")
   async def process_login_event(event_data: dict):
       # Processamento do evento
       pass
       
   # Exemplo de uso manual de métrica
   async def purge_old_records(tenant_id: str, region: str):
       start_time = time.time()
       deleted_count = await retention_service.execute_purge(tenant_id)
       duration = time.time() - start_time
       
       metrics.iam_audit_retention_purge_total.labels(
           tenant=tenant_id, 
           region=region,
           retention_policy="standard_30d"
       ).inc(deleted_count)
       
       metrics.iam_audit_retention_execution_seconds.labels(
           tenant=tenant_id,
           region=region,
           retention_policy="standard_30d"
       ).observe(duration)
   ```

## Alternativas Consideradas

### 1. Sistema de Métricas Personalizado

**Prós:**
- Extremamente adaptado às necessidades específicas
- Controle total sobre coleta e armazenamento

**Contras:**
- Reinventar a roda
- Incompatibilidade com ecossistema de ferramentas
- Maior custo de manutenção
- Complexidade adicional

### 2. Adoção de Padrão OpenTelemetry Completo

**Prós:**
- Padrão emergente da indústria
- Unificação de métricas, logs e traces

**Contras:**
- Maior complexidade de implementação inicial
- Ecossistema em evolução
- Overhead potencialmente maior
- Curva de aprendizado para equipe

### 3. Métricas Minimalistas (apenas essenciais)

**Prós:**
- Menor overhead
- Simplicidade de implementação
- Menor custo de armazenamento

**Contras:**
- Visibilidade limitada
- Dificuldade para diagnosticar problemas complexos
- Risco de pontos cegos operacionais

## Consequências

### Positivas

- **Visibilidade completa**: Métricas abrangem todos os aspectos críticos do serviço
- **Padronização**: Convenções claras facilitam desenvolvimento e manutenção
- **Multi-contexto**: Suporte nativo para ambientes multi-tenant e multi-regionais
- **Conformidade**: Capacidade de demonstrar compliance regulatório
- **Alertas eficazes**: Base sólida para detecção de anomalias e problemas
- **Análise dimensional**: Capacidade de drill-down por múltiplas dimensões

### Negativas

- **Overhead**: Impacto de performance pela coleta de métricas detalhadas
- **Complexidade**: Necessidade de manutenção cuidadosa da instrumentação
- **Cardinality explosion**: Risco de explosão de cardinalidade com muitos labels
- **Custo de armazenamento**: Maior volume de dados de métricas para armazenar

### Mitigação de Riscos

- Implementar rate limiting para métricas de alta cardinalidade
- Configurar agregação server-side para métricas frequentes
- Estabelecer políticas de retenção adequadas para dados de métricas
- Monitorar o overhead da instrumentação
- Revisar periodicamente o conjunto de métricas para relevância e utilização

## Conformidade com Padrões

- **Prometheus Naming Conventions**: https://prometheus.io/docs/practices/naming/
- **OpenMetrics Format**: https://openmetrics.io/
- **SRE Golden Signals**: Latency, Traffic, Errors, Saturation
- **ISO/IEC 27001 Monitoring Requirements**
- **PCI DSS 4.0 Requirements** (10.2, 10.7, 11.4)
- **INNOVABIZ Platform Observability Standards v2.5**

## Referências

1. Prometheus Best Practices - https://prometheus.io/docs/practices/instrumentation/
2. Google SRE Book: Monitoring Distributed Systems - https://sre.google/sre-book/monitoring-distributed-systems/
3. Grafana Labs: Metric and Label Naming - https://grafana.com/docs/grafana/latest/fundamentals/timeseries-dimensions/
4. INNOVABIZ Observability Standards v2.5 (Internal Document)
5. FastAPI Instrumentation Patterns - https://fastapi.tiangolo.com/advanced/middleware/