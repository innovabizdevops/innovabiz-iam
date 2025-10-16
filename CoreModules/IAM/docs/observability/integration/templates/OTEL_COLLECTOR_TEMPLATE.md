# Template para Configuração do OpenTelemetry Collector

## Visão Geral

Este documento fornece o template padrão INNOVABIZ para configuração do OpenTelemetry Collector, componente central que recebe, processa e exporta telemetria de todos os serviços da plataforma. Esta configuração segue as melhores práticas, padrões e requisitos estabelecidos no Framework de Integração de Observabilidade INNOVABIZ.

## Arquitetura de Referência

O OpenTelemetry Collector na plataforma INNOVABIZ segue um modelo de implantação em camadas:

1. **Collector Agente** - Implantado como sidecar ou daemonset próximo aos serviços, com foco em recepção e pré-processamento
2. **Collector Gateway** - Implantado como deployment centralizado por namespace, responsável por processamento, filtragem e roteamento
3. **Collector Central** - Deployment global que gerencia a exportação final para os backends de armazenamento

Esta arquitetura escalável garante resiliência, processamento eficiente e suporte ao contexto multi-dimensional da plataforma.

## Template de Configuração YAML

```yaml
# otel-collector-config.yaml
# Template de configuração do OpenTelemetry Collector para INNOVABIZ
# Compatível com OpenTelemetry Collector v0.83.0+

# Definição dos receptores - como o collector recebe dados
receivers:
  # Receptor OTLP para traces, métricas e logs
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
        tls:
          cert_file: /etc/otel/certs/tls.crt
          key_file: /etc/otel/certs/tls.key
      http:
        endpoint: 0.0.0.0:4318
        cors:
          allowed_origins:
            - "*"
          allowed_headers:
            - "*"
        tls:
          cert_file: /etc/otel/certs/tls.crt
          key_file: /etc/otel/certs/tls.key

  # Receptor Prometheus para compatibilidade com endpoints /metrics
  prometheus:
    config:
      scrape_configs:
        - job_name: 'otel-collector'
          scrape_interval: 15s
          static_configs:
            - targets: ['0.0.0.0:8888']

  # Receptor Jaeger para compatibilidade com clientes Jaeger existentes
  jaeger:
    protocols:
      grpc:
        endpoint: 0.0.0.0:14250
      thrift_http:
        endpoint: 0.0.0.0:14268

  # Receptor Zipkin para compatibilidade com clientes Zipkin existentes
  zipkin:
    endpoint: 0.0.0.0:9411

  # Receptor para logs via Fluentforward
  fluentforward:
    endpoint: 0.0.0.0:8006

# Definição dos processadores - como o collector processa os dados
processors:
  # Processador para gerenciamento de lotes
  batch:
    timeout: 10s
    send_batch_max_size: 1024
    send_batch_size: 512
  
  # Filtro de memória para evitar OOM
  memory_limiter:
    check_interval: 5s
    limit_mib: 1024
    spike_limit_mib: 128
  
  # Processador para contexto multi-dimensional INNOVABIZ
  attributes:
    actions:
      # Adicionar atributos padrão se não existirem
      - key: innovabiz.tenant.id
        value: default
        action: insert
      - key: innovabiz.region.id
        value: default
        action: insert
      - key: innovabiz.deployment.environment
        value: ${ENVIRONMENT}
        action: insert
  
  # Filtro de recursos para otimizar armazenamento
  resource:
    attributes:
      # Filtragem de atributos para manter os essenciais e multi-dimensionais
      - key: service.name
        action: keep
      - key: service.version
        action: keep
      - key: innovabiz.module.id
        action: keep
      - key: innovabiz.tenant.id
        action: keep
      - key: innovabiz.region.id
        action: keep
      - key: innovabiz.deployment.environment
        action: keep
      - key: k8s.namespace.name
        action: keep
      - key: k8s.pod.name
        action: keep
      - key: host.name
        action: keep
  
  # Processador para transformar atributos e aderência à padronização INNOVABIZ
  transform:
    trace_statements:
      - context: span
        statements:
          - set(resource.attributes["innovabiz.processed.timestamp"], now())
          - set(resource.attributes["innovabiz.collector.version"], "${COLLECTOR_VERSION}")
    metric_statements:
      - context: datapoint
        statements:
          - set(resource.attributes["innovabiz.processed.timestamp"], now())
          - set(resource.attributes["innovabiz.collector.version"], "${COLLECTOR_VERSION}")
    log_statements:
      - context: log
        statements:
          - set(resource.attributes["innovabiz.processed.timestamp"], now())
          - set(resource.attributes["innovabiz.collector.version"], "${COLLECTOR_VERSION}")

  # Filtro de Span para reduzir volume de dados em produção
  # Desativado em ambientes de dev/qa
  filter:
    metrics:
      include:
        match_type: regexp
        regexp:
          - ^innovabiz\..*
          - ^service\..*
          - ^api\..*
          - ^db\..*
          - ^process\..*
    spans:
      exclude:
        match_type: strict
        attributes:
          - key: http.url
            value: /health
          - key: http.url
            value: /ready
          - key: http.url
            value: /metrics

  # Amostrador probabilístico para ambientes de alta carga
  probabilistic_sampler:
    hash_seed: 22
    sampling_percentage: ${SAMPLING_PERCENTAGE}

  # Mascaramento de dados sensíveis para compliance
  redaction:
    allowed_keys:
      match_type: strict
      values: ["user.id", "payment.method.id"]
    blocked_values:
      match_type: regexp
      regexp:
        - "\\d{13,19}" # Números de cartão
        - "\\d{3}" # CVV
        - "\\d{3}-\\d{2}-\\d{4}" # SSN format
        - "[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}"  # Email
    redaction_text: "[REDACTED]"

# Definição dos exportadores - para onde o collector envia os dados
exporters:
  # Exportador para Prometheus
  prometheus:
    endpoint: 0.0.0.0:8889
    namespace: innovabiz
    const_labels:
      innovabiz.tenant.id: ${TENANT_ID}
      innovabiz.region.id: ${REGION_ID}
    send_timestamps: true
    metric_expiration: 180m
    enable_open_metrics: true
    resource_to_telemetry_conversion:
      enabled: true
  
  # Exportador para OTLP (cascading collectors)
  otlp:
    endpoint: ${OTLP_ENDPOINT}
    tls:
      cert_file: /etc/otel/certs/tls.crt
      key_file: /etc/otel/certs/tls.key
      ca_file: /etc/otel/certs/ca.crt
      insecure: false
    headers:
      x-innovabiz-tenant-id: ${TENANT_ID}
      x-innovabiz-region-id: ${REGION_ID}
      x-innovabiz-environment: ${ENVIRONMENT}
  
  # Exportador para Jaeger
  jaeger:
    endpoint: ${JAEGER_ENDPOINT}
    tls:
      insecure: false
      ca_file: /etc/otel/certs/ca.crt
  
  # Exportador para Elasticsearch (logs)
  elasticsearch:
    endpoints: [${ELASTICSEARCH_ENDPOINT}]
    index: "innovabiz-logs-${TENANT_ID}-${REGION_ID}-%{today}"
    mapping:
      mode: ecs
    tls:
      insecure: false
      ca_file: /etc/otel/certs/ca.crt
    sending_queue:
      enabled: true
      num_consumers: 4
      queue_size: 100
    retry_on_failure:
      enabled: true
      initial_interval: 5s
      max_interval: 30s
      max_elapsed_time: 5m
  
  # Exportador para ClickHouse (analíticos)
  clickhouse:
    endpoint: ${CLICKHOUSE_ENDPOINT}
    database: innovabiz_telemetry
    timeout: 10s
    retry_on_failure:
      enabled: true
      initial_interval: 5s
      max_interval: 30s
      max_elapsed_time: 5m
    logs_table_name: "logs_${TENANT_ID}_${REGION_ID}"
    traces_table_name: "traces_${TENANT_ID}_${REGION_ID}"
    tls:
      insecure: false
      ca_file: /etc/otel/certs/ca.crt
  
  # Exportador para Loki (logs)
  loki:
    endpoint: ${LOKI_ENDPOINT}
    tenant_id: ${TENANT_ID}
    labels:
      resource:
        service.name: "service_name"
        service.version: "service_version"
        innovabiz.tenant.id: "tenant_id"
        innovabiz.region.id: "region_id"
        innovabiz.module.id: "module_id"
        innovabiz.deployment.environment: "environment"
    tls:
      insecure: false
      ca_file: /etc/otel/certs/ca.crt
  
  # Exportador para arquivo local (debug, principalmente para ambientes não-prod)
  file:
    path: /var/log/otel/collector-export.json
    rotation:
      max_size: 100
      max_backups: 5
      max_age: 7
      compress: true

  # Exportador para debug (console)
  debug:
    verbosity: detailed

# Extensões para funcionalidades adicionais
extensions:
  # Health check
  health_check:
    endpoint: 0.0.0.0:13133
  
  # Métricas internas do collector
  pprof:
    endpoint: 0.0.0.0:1777
  
  # Métricas Prometheus do próprio collector
  zpages:
    endpoint: 0.0.0.0:55679

  # Gestão de credenciais seguras
  file_storage:
    directory: /etc/otel/storage

  # Autenticação para endpoints de API
  basicauth:
    htpasswd:
      file: /etc/otel/auth/htpasswd
      inline: ${HTPASSWD_CONTENT}

# Configuração de serviços - como os componentes se conectam
service:
  # Extensões ativas
  extensions: [health_check, pprof, zpages, file_storage, basicauth]
  
  # Pipeline para traces
  pipelines:
    # Pipeline de traces para processamento padrão
    traces:
      receivers: [otlp, jaeger, zipkin]
      processors: [memory_limiter, attributes, resource, redaction, transform, probabilistic_sampler, batch]
      exporters: [otlp, jaeger, clickhouse]
      # Configuração condicional por ambiente
      # ${TRACES_EXPORTERS} seria substituído em tempo de implantação
      exporters: ${TRACES_EXPORTERS}
    
    # Pipeline de métricas para processamento padrão
    metrics:
      receivers: [otlp, prometheus]
      processors: [memory_limiter, attributes, resource, filter, transform, batch]
      exporters: [prometheus, otlp, clickhouse]
      # Configuração condicional por ambiente
      exporters: ${METRICS_EXPORTERS}
    
    # Pipeline de logs para processamento padrão
    logs:
      receivers: [otlp, fluentforward]
      processors: [memory_limiter, attributes, resource, redaction, transform, batch]
      exporters: [elasticsearch, loki, clickhouse]
      # Configuração condicional por ambiente
      exporters: ${LOGS_EXPORTERS}
    
    # Pipeline de debug para validação (habilitado via variável de ambiente)
    traces/debug:
      receivers: [otlp]
      processors: [memory_limiter, attributes, batch]
      exporters: [debug, file]
      # Ativo apenas quando DEBUG=true
      enabled: ${DEBUG}

  # Configuração de telemetria do próprio collector
  telemetry:
    logs:
      level: ${LOG_LEVEL}
      development: false
      encoding: json
      output_paths: [stdout, /var/log/otel/collector.log]
      error_output_paths: [stderr, /var/log/otel/collector-error.log]
      initial_fields:
        innovabiz.tenant.id: ${TENANT_ID}
        innovabiz.region.id: ${REGION_ID}
        innovabiz.deployment.environment: ${ENVIRONMENT}
        collector.type: ${COLLECTOR_TYPE}
    metrics:
      level: detailed
      address: 0.0.0.0:8888
```

## Kubernetes ConfigMap e Deployment

### ConfigMap para o Collector

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: otel-collector-config
  namespace: observability
  labels:
    app: otel-collector
    innovabiz.module: observability
data:
  otel-collector-config.yaml: |
    # Cópia completa da configuração YAML acima
  collector-env: |
    # Variáveis de ambiente para substituição
    ENVIRONMENT=production
    TENANT_ID=default
    REGION_ID=default
    OTLP_ENDPOINT=otel-collector-gateway.observability:4317
    JAEGER_ENDPOINT=jaeger-collector.observability:14250
    ELASTICSEARCH_ENDPOINT=elasticsearch-master.observability:9200
    LOKI_ENDPOINT=loki-gateway.observability:3100
    CLICKHOUSE_ENDPOINT=clickhouse.observability:9000
    COLLECTOR_VERSION=0.83.0
    COLLECTOR_TYPE=agent
    SAMPLING_PERCENTAGE=10
    LOG_LEVEL=info
    DEBUG=false
    TRACES_EXPORTERS=[otlp, clickhouse]
    METRICS_EXPORTERS=[prometheus, otlp, clickhouse]
    LOGS_EXPORTERS=[elasticsearch, loki]
```

### Deployment do Collector como DaemonSet

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: otel-collector-agent
  namespace: observability
  labels:
    app: otel-collector-agent
    innovabiz.module: observability
spec:
  selector:
    matchLabels:
      app: otel-collector-agent
  template:
    metadata:
      labels:
        app: otel-collector-agent
        innovabiz.module: observability
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8888"
    spec:
      serviceAccountName: otel-collector
      securityContext:
        fsGroup: 10001
        runAsUser: 10001
        runAsGroup: 10001
        runAsNonRoot: true
      containers:
      - name: otel-collector
        image: otel/opentelemetry-collector-contrib:0.83.0
        args:
        - --config=/etc/otel/config/otel-collector-config.yaml
        - --set=service.telemetry.logs.initial_fields.host.name=$(NODE_NAME)
        env:
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        # Variáveis de ambiente do ConfigMap
        envFrom:
        - configMapRef:
            name: otel-collector-config
            key: collector-env
        resources:
          limits:
            cpu: 500m
            memory: 1Gi
          requests:
            cpu: 100m
            memory: 256Mi
        ports:
        - containerPort: 4317 # OTLP gRPC
          name: otlp-grpc
          protocol: TCP
        - containerPort: 4318 # OTLP HTTP
          name: otlp-http
          protocol: TCP
        - containerPort: 8888 # Prometheus metrics
          name: metrics
          protocol: TCP
        - containerPort: 13133 # Health check
          name: health
          protocol: TCP
        livenessProbe:
          httpGet:
            path: /
            port: 13133
          initialDelaySeconds: 15
          timeoutSeconds: 5
        readinessProbe:
          httpGet:
            path: /
            port: 13133
          initialDelaySeconds: 10
          timeoutSeconds: 5
        volumeMounts:
        - mountPath: /etc/otel/config
          name: otel-collector-config-vol
        - mountPath: /etc/otel/certs
          name: otel-collector-certs
          readOnly: true
        - mountPath: /var/log/otel
          name: otel-collector-logs
        - mountPath: /etc/otel/storage
          name: otel-collector-storage
        - mountPath: /etc/otel/auth
          name: otel-collector-auth
          readOnly: true
      volumes:
      - name: otel-collector-config-vol
        configMap:
          name: otel-collector-config
          items:
          - key: otel-collector-config.yaml
            path: otel-collector-config.yaml
      - name: otel-collector-certs
        secret:
          secretName: otel-collector-certs
      - name: otel-collector-logs
        emptyDir: {}
      - name: otel-collector-storage
        emptyDir: {}
      - name: otel-collector-auth
        secret:
          secretName: otel-collector-auth
```

### Deployment do Collector Gateway

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: otel-collector-gateway
  namespace: observability
  labels:
    app: otel-collector-gateway
    innovabiz.module: observability
spec:
  replicas: 2
  selector:
    matchLabels:
      app: otel-collector-gateway
  template:
    metadata:
      labels:
        app: otel-collector-gateway
        innovabiz.module: observability
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8888"
    spec:
      serviceAccountName: otel-collector
      securityContext:
        fsGroup: 10001
        runAsUser: 10001
        runAsGroup: 10001
        runAsNonRoot: true
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - otel-collector-gateway
              topologyKey: kubernetes.io/hostname
      containers:
      - name: otel-collector
        image: otel/opentelemetry-collector-contrib:0.83.0
        args:
        - --config=/etc/otel/config/otel-collector-config.yaml
        - --set=service.telemetry.logs.initial_fields.host.name=$(POD_NAME)
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        # Variáveis de ambiente modificadas para o gateway
        envFrom:
        - configMapRef:
            name: otel-collector-config
            key: collector-env
        env:
        - name: COLLECTOR_TYPE
          value: "gateway"
        - name: OTLP_ENDPOINT
          value: "otel-collector-central.observability:4317"
        resources:
          limits:
            cpu: 1000m
            memory: 2Gi
          requests:
            cpu: 200m
            memory: 512Mi
        # Outras configurações seguem o mesmo padrão do DaemonSet
        # ...
```

## Variáveis de Ambiente para Configuração

As variáveis de ambiente a seguir são usadas para personalizar a configuração do collector para diferentes ambientes e contextos:

| Variável | Descrição | Padrão |
|----------|-----------|--------|
| `ENVIRONMENT` | Ambiente de implantação (development, staging, production) | development |
| `TENANT_ID` | ID do tenant para isolamento multi-tenant | default |
| `REGION_ID` | ID da região para contexto multi-regional | default |
| `OTLP_ENDPOINT` | Endpoint para exportação OTLP (próximo collector na cadeia) | otel-collector-gateway:4317 |
| `JAEGER_ENDPOINT` | Endpoint do serviço Jaeger | jaeger-collector:14250 |
| `ELASTICSEARCH_ENDPOINT` | Endpoint do Elasticsearch para exportação de logs | elasticsearch-master:9200 |
| `LOKI_ENDPOINT` | Endpoint do Loki para exportação de logs | loki-gateway:3100 |
| `CLICKHOUSE_ENDPOINT` | Endpoint do ClickHouse para armazenamento analítico | clickhouse:9000 |
| `COLLECTOR_VERSION` | Versão do collector para rastreabilidade | 0.83.0 |
| `COLLECTOR_TYPE` | Tipo de collector (agent, gateway, central) | agent |
| `SAMPLING_PERCENTAGE` | Porcentagem de amostragem para traces | 100 |
| `LOG_LEVEL` | Nível de logging do collector | info |
| `DEBUG` | Habilitar pipeline de debug | false |
| `TRACES_EXPORTERS` | Lista de exportadores para traces | [otlp] |
| `METRICS_EXPORTERS` | Lista de exportadores para métricas | [prometheus, otlp] |
| `LOGS_EXPORTERS` | Lista de exportadores para logs | [elasticsearch] |

## Melhores Práticas

1. **Configuração por Tipo de Collector**
   - **Agent**: Foco em recepção e pré-processamento com baixo overhead
   - **Gateway**: Balanceamento entre processamento e agregação
   - **Central**: Otimizado para exportação confiável e processamento avançado

2. **Segurança**
   - Use TLS para todas as comunicações (mTLS quando possível)
   - Implemente autenticação para endpoints de API
   - Proteja credenciais usando secrets do Kubernetes
   - Aplique redação de dados sensíveis antes da exportação

3. **Performance e Resiliência**
   - Ajuste tamanhos de lote com base no volume de telemetria
   - Configure filas e retries para todos os exportadores
   - Implemente limitadores de memória para evitar OOM
   - Use anti-affinity para garantir alta disponibilidade do gateway

4. **Multi-dimensionalidade**
   - Preserve sempre os atributos de contexto multi-dimensional
   - Use processadores de atributos para garantir consistência
   - Adicione timestamps de processamento para rastreabilidade
   - Mantenha nomeação consistente para índices/tabelas

5. **Operação**
   - Monitore o próprio collector com exportação de métricas
   - Configure health checks para facilitar a detecção de problemas
   - Mantenha logs estruturados para troubleshooting
   - Implemente alarmes para gargalos de processamento

## Checklist de Validação

- [ ] TLS configurado para todas as comunicações
- [ ] Contexto multi-dimensional preservado em todo o pipeline
- [ ] Processadores de redação configurados para dados sensíveis
- [ ] Limitadores de recursos definidos apropriadamente
- [ ] Filas e retries configurados para todos os exportadores
- [ ] Health checks implementados
- [ ] Métricas do collector expostas
- [ ] Amostragem configurada conforme necessidades do ambiente
- [ ] Configuração de alta disponibilidade para gateway e central
- [ ] Variáveis de ambiente documentadas

## Recursos Adicionais

- [Documentação OpenTelemetry Collector](https://opentelemetry.io/docs/collector/)
- [Portal de Observabilidade INNOVABIZ](https://observability.innovabiz.com)
- [Guia de Operação do Collector](https://wiki.innovabiz.com/observability/collector-ops)
- [Calculadora de Recursos do Collector](https://tools.innovabiz.com/collector-sizing)