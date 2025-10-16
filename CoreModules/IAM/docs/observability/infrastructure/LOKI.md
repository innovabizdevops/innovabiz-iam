# INNOVABIZ IAM Audit Service - Documentação Loki

**Versão:** 3.0.0  
**Data:** 31/07/2025  
**Classificação:** Oficial  
**Status:** ✅ Implementado  
**Autor:** Equipe INNOVABIZ DevOps  
**Aprovado por:** Eduardo Jeremias  

## 1. Visão Geral

O Loki é o segundo componente central da estratégia "dual-write" de armazenamento de logs na arquitetura de observabilidade do IAM Audit Service da INNOVABIZ. Projetado como um sistema de agregação de logs altamente eficiente e econômico, o Loki complementa o Elasticsearch, fornecendo capacidades de armazenamento e consulta com ênfase em eficiência de recursos e integração com o ecossistema Grafana.

### 1.1 Funcionalidades Principais

- **Indexação Eficiente**: Indexação apenas de metadados (labels) em vez do conteúdo completo
- **Armazenamento Otimizado**: Compressão avançada e sistema de chunks para economia de espaço
- **LogQL**: Linguagem de consulta poderosa similar a PromQL para filtragem e análise
- **Multi-tenant**: Isolamento completo de dados por tenant
- **Integração Nativa**: Com Grafana para visualização e alerta
- **Escalabilidade**: Arquitetura microserviço para escala horizontal

### 1.2 Posicionamento na Arquitetura

O Loki atua como repositório secundário de logs, recebendo dados de:

- Fluentd (via dual-write com Elasticsearch)
- Promtail para logs de aplicações específicas
- Vector para processamento avançado de logs
- OpenTelemetry Collector (via exporter OTLP)

E fornecendo dados para:

- Grafana (visualização e análise)
- Portal de Observabilidade (API unificada)
- AlertManager (via regras LogQL)
- Correlação com métricas e traces (via Grafana)

## 2. Implementação Técnica

### 2.1 Manifesto Kubernetes

O Loki é implementado como um conjunto de microserviços no Kubernetes, conforme definido em `observability/loki.yaml`. Os principais componentes incluem:

```yaml
# Trecho exemplificativo do manifesto
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: loki-distributor
  namespace: iam-system
  labels:
    app.kubernetes.io/name: loki
    app.kubernetes.io/component: distributor
    app.kubernetes.io/part-of: innovabiz-observability
    innovabiz.com/module: iam-audit
    innovabiz.com/tier: observability
spec:
  replicas: 2
  serviceName: loki-distributor
  # ... outras configurações
  template:
    spec:
      containers:
      - name: loki
        image: grafana/loki:2.9.0
        args:
        - -config.file=/etc/loki/config.yaml
        - -target=distributor
        resources:
          limits:
            cpu: 1000m
            memory: 2Gi
          requests:
            cpu: 500m
            memory: 1Gi
        # ... outras configurações
---
# Outros componentes: ingester, querier, compactor, etc.
```

### 2.2 Arquitetura de Microserviços

A implementação do Loki segue o padrão de microserviços, com os seguintes componentes:

| Componente | Função | Replicas | Recursos (req/limits) |
|------------|--------|----------|------------------------|
| **Distributor** | Recebe e distribui logs para ingesters | 2 | 500m/1000m CPU, 1Gi/2Gi Mem |
| **Ingester** | Processa e armazena chunks de logs | 3 | 1000m/2000m CPU, 2Gi/4Gi Mem |
| **Querier** | Executa consultas nos dados armazenados | 2 | 500m/1000m CPU, 1Gi/2Gi Mem |
| **Query Frontend** | Gerencia e otimiza consultas | 2 | 500m/1000m CPU, 1Gi/2Gi Mem |
| **Compactor** | Compacta dados antigos | 1 | 500m/1000m CPU, 1Gi/2Gi Mem |
| **Index Gateway** | Cache de índices para consultas | 2 | 500m/1000m CPU, 2Gi/4Gi Mem |
| **Ruler** | Executa regras de alerta | 2 | 500m/1000m CPU, 1Gi/2Gi Mem |

### 2.3 Configuração do Loki

A configuração do Loki é gerenciada via ConfigMap e segue o modelo abaixo:

```yaml
auth_enabled: true

server:
  http_listen_port: 3100

common:
  path_prefix: /loki
  storage:
    filesystem:
      chunks_directory: /var/loki/chunks
      rules_directory: /var/loki/rules
  replication_factor: 3

compactor:
  working_directory: /var/loki/compactor

limits_config:
  enforce_metric_name: false
  reject_old_samples: true
  reject_old_samples_max_age: 168h
  max_entries_limit_per_query: 5000
  split_queries_by_interval: 30m
  per_tenant_override_config: /etc/loki/overrides.yaml

memberlist:
  abort_if_cluster_join_fails: false
  join_members:
  - loki-memberlist

schema_config:
  configs:
  - from: 2023-01-01
    store: boltdb-shipper
    object_store: filesystem
    schema: v12
    index:
      prefix: index_
      period: 24h

storage_config:
  boltdb_shipper:
    active_index_directory: /var/loki/index
    cache_location: /var/loki/cache
    shared_store: filesystem
  filesystem:
    directory: /var/loki/chunks

ruler:
  storage:
    type: local
    local:
      directory: /etc/loki/rules
  rule_path: /tmp/loki/rules
  alertmanager_url: http://alertmanager.iam-system.svc:9093
  ring:
    kvstore:
      store: memberlist

ingester:
  lifecycler:
    ring:
      kvstore:
        store: memberlist
      replication_factor: 3
  chunk_idle_period: 15m
  chunk_block_size: 262144
  chunk_retain_period: 30s
  chunk_encoding: snappy
  wal:
    enabled: true
    dir: /var/loki/wal

query_range:
  align_queries_with_step: true
  max_retries: 5
  cache_results: true
  results_cache:
    cache:
      enable_fifocache: true
      fifocache:
        max_size_items: 1024
        ttl: 24h

frontend:
  log_queries_longer_than: 5s
  compress_responses: true
  tenant_query_timeout: 1m
  max_outstanding_per_tenant: 2048

frontend_worker:
  frontend_address: loki-query-frontend:9095

tenant_federation:
  enabled: true
```

### 2.4 Estratégia de Persistência

- **Volumes Persistentes**:
  - StorageClass SSD para dados de alta performance
  - PVCs separados para chunks, WAL e índices
- **Retenção**:
  - Padrão: 15 dias para todos os logs
  - Premium: 30 dias
  - Críticos: 90 dias
- **Compactação**:
  - Automática para chunks mais antigos que 12 horas
  - Redução de tamanho em ~5x após compactação
- **Índice**:
  - BoltDB-Shipper para índices
  - Período de 24h para rotação de índices

### 2.5 Segurança e Controle de Acesso

- **Autenticação**:
  - OIDC para integração com IAM
  - Tokens JWT com tenant_id
- **Multi-tenant**:
  - Isolamento completo por tenant_id
  - Limites configuráveis por tenant
- **TLS**:
  - Obrigatório para todas as comunicações externas
  - mTLS para comunicações entre componentes
- **Autorização**:
  - RBAC baseado em grupos do IAM
  - Capacidade de limitar escopo de consultas

## 3. Configuração Multi-dimensional

### 3.1 Estratégia de Labels Multi-contexto

O Loki utiliza um esquema padronizado de labels para garantir a capacidade multi-dimensional:

```
{tenant_id="tenant1", region_id="br-east-1", environment="production", module="iam-audit", component="authentication-service", level="error"}
```

### 3.2 Isolamento por Tenant

- **Tenant ID Obrigatório**: Toda escrita e leitura exige tenant_id
- **Limites Específicos**: Configurados em `overrides.yaml` por tenant
- **Políticas de Retenção**: Customizáveis por tenant
- **Métricas de Uso**: Coletadas e expostas por tenant para billing

Exemplo de configuração de overrides por tenant:

```yaml
overrides:
  tenant1:
    ingestion_rate_mb: 10
    ingestion_burst_size_mb: 20
    max_global_streams_per_user: 5000
    max_query_length: 12h
    max_query_parallelism: 32
    retention_period: 360h
  tenant2:
    ingestion_rate_mb: 5
    ingestion_burst_size_mb: 10
    max_global_streams_per_user: 2000
    max_query_length: 6h
    max_query_parallelism: 16
    retention_period: 168h
```

### 3.3 Contexto Regional

- **Region ID como Label**: Todas as entradas possuem region_id
- **Consultas Cross-Region**: Suportadas via filtros LogQL
- **Compliance Regional**: Garante que dados permanecem na região apropriada
- **Replicação Cross-Region**: Opcional para tenants premium

### 3.4 Churn de Labels

Para controlar o "label churn" (proliferação excessiva de valores de labels), implementamos:

- **Label Normalização**: Normalização automática de valores para formatos padronizados
- **Blacklisting**: Bloqueio de labels com alta cardinalidade
- **Rate Limiting**: Por número de streams exclusivos

## 4. Integração com Coletores de Logs

### 4.1 Integração com Fluentd

Fluentd é configurado para dual-write para Elasticsearch e Loki:

```ruby
# Trecho da configuração do Fluentd
<match **>
  @type copy
  <store>
    @type elasticsearch
    host elasticsearch.iam-system.svc
    port 9200
    # ... configurações do Elasticsearch
  </store>
  <store>
    @type loki
    url "#{ENV['LOKI_URL']}"
    tenant "#{ENV['TENANT_ID']}"
    extra_labels {"region_id":"#{ENV['REGION_ID']}","environment":"#{ENV['ENVIRONMENT']}"}
    label_keys tenant_id,region_id,environment,module,component,level,host
    line_format json
    # ... outras configurações
    <buffer>
      @type file
      path /var/log/fluentd/loki
      flush_mode interval
      flush_interval 5s
      flush_thread_count 4
      retry_forever true
      retry_max_interval 30
      chunk_limit_size 2M
    </buffer>
  </store>
</match>
```

### 4.2 Integração com Promtail

Promtail é utilizado para coletar logs específicos de aplicação:

```yaml
server:
  http_listen_port: 9080

clients:
  - url: http://loki-distributor.iam-system.svc:3100/loki/api/v1/push
    tenant_id: ${TENANT_ID}
    bearer_token: ${LOKI_TOKEN}

positions:
  filename: /run/promtail/positions.yaml

scrape_configs:
  - job_name: kubernetes-pods
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app_kubernetes_io_name]
        target_label: app
      - source_labels: [__meta_kubernetes_pod_label_app_kubernetes_io_component]
        target_label: component
      - source_labels: [__meta_kubernetes_pod_node_name]
        target_label: node_name
      - source_labels: [__meta_kubernetes_namespace]
        target_label: namespace
      - source_labels: [__meta_kubernetes_pod_name]
        target_label: pod
      - source_labels: [__meta_kubernetes_pod_container_name]
        target_label: container
      - action: replace
        source_labels: [__meta_kubernetes_namespace]
        regex: iam-system
        target_label: tenant_id
        replacement: ${TENANT_ID}
      - action: replace
        source_labels: [__meta_kubernetes_namespace]
        regex: iam-system
        target_label: region_id
        replacement: ${REGION_ID}
    pipeline_stages:
      - json:
          expressions:
            level: level
            timestamp: timestamp
            message: message
      - labels:
          level:
      - timestamp:
          source: timestamp
          format: RFC3339
```

### 4.3 Integração com OpenTelemetry

O OpenTelemetry Collector é configurado para enviar logs para o Loki:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 1s
    send_batch_size: 1024
  memory_limiter:
    check_interval: 1s
    limit_mib: 1024
  resource:
    attributes:
    - key: tenant_id
      value: ${TENANT_ID}
      action: upsert
    - key: region_id
      value: ${REGION_ID}
      action: upsert
    - key: environment
      value: ${ENVIRONMENT}
      action: upsert

exporters:
  loki:
    endpoint: http://loki-distributor.iam-system.svc:3100/loki/api/v1/push
    tenant_id: ${TENANT_ID}
    labels:
      resource:
        tenant_id: tenant_id
        region_id: region_id
        environment: environment
        service.name: service_name
      record:
        severity: severity
    auth:
      authenticator: bearer
      bearer:
        token: ${LOKI_TOKEN}

service:
  pipelines:
    logs:
      receivers: [otlp]
      processors: [memory_limiter, batch, resource]
      exporters: [loki]
```

## 5. Consultas e Análise de Logs

### 5.1 LogQL

O Loki utiliza LogQL, uma linguagem de consulta inspirada em PromQL:

```logql
# Filtrar logs por labels
{tenant_id="tenant1", region_id="br-east-1", component="auth-service", level="error"}

# Busca por conteúdo
{tenant_id="tenant1"} |= "authentication failed"

# Extração de valores
{tenant_id="tenant1"} | json | response_time > 500

# Agregação e métricas
sum by(component)(count_over_time({tenant_id="tenant1", level="error"}[5m]))

# Correlação com métricas
{tenant_id="tenant1"} | json | response_time > 500 and status_code=500

# Extração de padrões
{tenant_id="tenant1"} | pattern "<ip> - <user> [<timestamp>] \"<method> <url> <proto>\" <status> <bytes>"
```

### 5.2 Estratégias de Consulta Eficiente

- **Filtrar Primeiro por Labels**: Sempre iniciar consultas com filtros de labels
- **Limitar Janela de Tempo**: Usar intervalos menores para consultas pesadas
- **Agregações**: Usar métricas derivadas para análise de tendências
- **Extrações Seletivas**: Extrair apenas os campos necessários
- **Caching**: Consultas comuns são automaticamente cacheadas

### 5.3 Métricas Derivadas de Logs

O Loki permite criar métricas a partir de logs para análise em Prometheus/Grafana:

```logql
# Taxa de erros por componente
sum by (component)(rate({tenant_id="tenant1", level="error"}[5m]))

# Latência média extraída de logs
avg by (component)(
  sum_over_time(
    {tenant_id="tenant1"} 
    | json 
    | unwrap response_time [5m]
  )
)

# Contagem de falhas de autenticação
sum(count_over_time({tenant_id="tenant1", component="auth-service"} |= "authentication failed" [5m]))
```

## 6. Integrações

### 6.1 Integração com Grafana

- **Datasource**: Loki configurado como fonte nativa no Grafana
- **Exploração**: Interface LogQL com highlight e sugestões
- **Dashboards**: Templates específicos para análise de logs
- **Alerting**: Alertas baseados em padrões de logs
- **Correlação**: Links automáticos entre logs, métricas e traces

### 6.2 Integração com AlertManager

Alertas baseados em logs são configurados via regras do Loki:

```yaml
groups:
  - name: iam-audit-alerts
    rules:
      - alert: HighAuthFailureRate
        expr: sum by (tenant_id)(rate({tenant_id=~".+", component="auth-service"} |= "authentication failed" [5m])) > 10
        for: 5m
        labels:
          severity: warning
          category: security
        annotations:
          summary: "Alta taxa de falhas de autenticação"
          description: "Detectadas mais de 10 falhas de autenticação por minuto"
          
      - alert: UnauthorizedAccessAttempt
        expr: sum by (tenant_id)(count_over_time({tenant_id=~".+", level="ERROR"} |= "unauthorized access" [5m])) > 5
        for: 2m
        labels:
          severity: critical
          category: security
        annotations:
          summary: "Tentativas de acesso não autorizado"
          description: "Múltiplas tentativas de acesso não autorizado detectadas"
```

### 6.3 Integração com Portal de Observabilidade

- **API GraphQL**: Exposição de logs via API unificada
- **UI Customizada**: Interface de consulta simplificada
- **Dashboards**: Visualizações pré-configuradas
- **Correlação**: Relacionamento automático com outros sinais

## 7. Monitoramento e Alerta

### 7.1 Métricas Expostas

O Loki expõe métricas detalhadas via endpoint Prometheus:

| Métrica | Descrição | Threshold de Alerta |
|---------|-----------|---------------------|
| `loki_distributor_bytes_received_total` | Bytes recebidos pelo distributor | Queda >50% |
| `loki_ingester_memory_chunks` | Chunks na memória dos ingesters | >80% capacidade |
| `loki_ingester_chunk_entries` | Entradas por chunk | N/A (monitoramento) |
| `loki_ingester_chunk_size_bytes` | Tamanho médio de chunks | N/A (monitoramento) |
| `loki_ingester_chunk_utilization` | Utilização de chunks | <50% (ineficiente) |
| `loki_ingester_chunk_age_seconds` | Idade de chunks antes do flush | >4h (possível problema) |
| `loki_query_frontend_queries_total` | Total de consultas | N/A (monitoramento) |
| `loki_query_frontend_queue_length` | Tamanho da fila de consultas | >100 por 5min |
| `loki_querier_query_latency_seconds` | Latência de consultas | p95 >5s |

### 7.2 Alertas Configurados

```yaml
# Exemplo de alertas para monitorar o Loki
- name: LokiAlerts
  rules:
  - alert: LokiProcessErrors
    expr: sum(rate(loki_process_failures_total[5m])) by (job) > 0
    for: 10m
    labels:
      severity: critical
      component: loki
    annotations:
      summary: "Loki está enfrentando erros de processamento"
      description: "Detectados erros de processamento em {{ $labels.job }}"
      runbook: "https://docs.innovabiz.com/observability/runbooks/loki-process-errors"

  - alert: LokiRequestErrors
    expr: sum(rate(loki_request_duration_seconds_count{status_code=~"5.."}[5m])) by (route) > 1
    for: 5m
    labels:
      severity: critical
      component: loki
    annotations:
      summary: "Alta taxa de erros HTTP 5xx no Loki"
      description: "Rota {{ $labels.route }} está retornando erros 5xx"
      runbook: "https://docs.innovabiz.com/observability/runbooks/loki-request-errors"
```

### 7.3 Dashboards de Monitoramento

- **Loki Operational**: Métricas operacionais do cluster
- **Loki Resource Usage**: Uso de recursos por componente
- **Loki Performance**: Latência e throughput
- **Loki Ingestion**: Taxa de ingestão e erros
- **Loki Queries**: Performance de consultas

## 8. Backup e Recuperação

### 8.1 Estratégia de Backup

- **Chunks**: Backup diário para object storage
- **Índices**: Backup diário dos índices BoltDB
- **Retenção de Backup**: 30 dias padrão
- **Consistência**: Snapshots consistentes com coordenação entre serviços

### 8.2 Procedimento de Recuperação

1. **Recuperação de Índices**:
   - Restaurar índices BoltDB do backup
   - Reiniciar queriers e ingesters
   - Verificar consistência dos índices

2. **Recuperação de Chunks**:
   - Restaurar chunks do backup para storage
   - Verificar integridade dos arquivos
   - Reiniciar componentes para reconhecer novos chunks

3. **Recuperação Completa**:
   - Restaurar configuração e overrides
   - Restaurar índices e chunks
   - Verificar e corrigir inconsistências

### 8.3 RPO/RTO

| Nível de Serviço | RPO | RTO | Cobertura |
|------------------|-----|-----|-----------|
| **Standard** | 24 horas | 6 horas | Todos os tenants |
| **Premium** | 12 horas | 4 horas | Tenants premium |
| **Enterprise** | 6 horas | 2 horas | Tenants enterprise |

## 9. Conformidade e Segurança

### 9.1 Requisitos de Conformidade

| Regulação | Requisito | Implementação |
|-----------|-----------|---------------|
| **PCI DSS 4.0** | 10.2 Trilhas de auditoria | Logs de auditoria com retenção apropriada |
| **GDPR/LGPD** | Art. 17 Direito ao esquecimento | Pipeline para pseudonimização + API de exclusão |
| **ISO 27001** | A.12.4 Registros de eventos | Captura abrangente e protegida |
| **NIST 800-53** | SI-4 Monitoramento do sistema | Correlação de eventos de segurança |

### 9.2 Controles de Segurança

- **Autenticação**: OIDC com IAM central
- **Autorização**: RBAC com controle granular
- **Multi-tenancy**: Isolamento completo por tenant
- **Criptografia**: TLS para comunicações, disco criptografado
- **Auditoria**: Logs de acesso e operações administrativas

## 10. Operação e Manutenção

### 10.1 Procedimentos Operacionais

- **Health Check**: Automatizado a cada 5 minutos
- **Compactação**: Verificação diária de jobs pendentes
- **Monitoramento de Recursos**: Alertas preditivos de capacidade
- **Verificação de Performance**: Análise semanal de métricas

### 10.2 Troubleshooting

| Problema | Possíveis Causas | Resolução |
|----------|-----------------|-----------|
| Alta latência de consulta | Range muito amplo, filtros ineficientes | Otimizar consultas, adicionar mais queriers |
| Erros de ingestão | Rate limit, formato inválido | Verificar limites, validar formato dos logs |
| Memória alta em ingesters | Muitos streams, flush lento | Ajustar flush interval, adicionar mais ingesters |
| Erros de compactação | Falta de recursos, arquivos corrompidos | Verificar storage, reiniciar compactor |
| Índices inconsistentes | Falha durante flush, corrupção de dados | Reconstruir índices a partir de chunks |

### 10.3 Escalabilidade

- **Vertical**: Aumentar recursos por componente
- **Horizontal**: Adicionar mais réplicas de cada componente
- **Otimização**:
  - Ajustar retenção por importância
  - Reduzir cardinalidade de labels
  - Implementar sampling para logs de alto volume

## 11. Considerações de Evolução

### 11.1 Roadmap

1. **Curto Prazo** (3 meses):
   - Implementação de sampling inteligente
   - Melhorias na interface de consulta
   - Otimização de storage para alta performance

2. **Médio Prazo** (6-12 meses):
   - Machine learning para detecção de anomalias
   - Maior integração com métricas e traces
   - Sistema de visualização aprimorado

3. **Longo Prazo** (12+ meses):
   - Federação global
   - Análise preditiva baseada em logs
   - Auto-otimização baseada em padrões de uso

## 12. Referências

1. [Grafana Loki Documentation](https://grafana.com/docs/loki/latest/)
2. [LogQL Query Language](https://grafana.com/docs/loki/latest/logql/)
3. [Loki Architecture](https://grafana.com/docs/loki/latest/fundamentals/architecture/)
4. [Best Practices for Labels](https://grafana.com/docs/loki/latest/best-practices/)
5. [PCI DSS 4.0 Requirements](https://www.pcisecuritystandards.org/)
6. [SRE Book: Logging and Monitoring](https://sre.google/sre-book/monitoring-distributed-systems/)
7. [Observability Engineering](https://www.oreilly.com/library/view/observability-engineering/9781492076438/)

---

*Documento aprovado pelo Comitê de Arquitetura INNOVABIZ em 28/07/2025*