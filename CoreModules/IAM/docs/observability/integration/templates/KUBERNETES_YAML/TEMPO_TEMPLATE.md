# Template Kubernetes YAML para Tempo - INNOVABIZ

## Visão Geral

Este documento fornece templates YAML para implantação do Grafana Tempo no ambiente Kubernetes da plataforma INNOVABIZ. O Tempo é um backend de rastreamento distribuído de alto volume e baixo custo, altamente escalável, projetado para integração nativa com o Grafana para visualização. Os templates seguem as melhores práticas de segurança, dimensionamento e configuração multi-dimensional conforme os padrões INNOVABIZ.

## Sumário

1. [Namespace](#namespace)
2. [ConfigMap](#configmap)
3. [Secret](#secret)
4. [PersistentVolumeClaim](#persistentvolumeclaim)
5. [ServiceAccount e RBAC](#serviceaccount-e-rbac)
6. [StatefulSet](#statefulset)
7. [Service](#service)
8. [Ingress](#ingress)
9. [NetworkPolicy](#networkpolicy)
10. [PodDisruptionBudget](#poddisruptionbudget)
11. [ServiceMonitor](#servicemonitor)
12. [Checklist de Validação](#checklist-de-validação)
13. [Melhores Práticas](#melhores-práticas)
14. [Exemplos de Uso](#exemplos-de-uso)

## Namespace

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: observability-${TENANT_ID}-${REGION_ID}
  labels:
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "tracing"
```

## ConfigMap

```yaml
# tempo-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tempo-config
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: tempo
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "tracing"
data:
  tempo.yaml: |
    auth_enabled: ${TEMPO_AUTH_ENABLED}

    server:
      http_listen_port: 3200
      grpc_listen_port: 9096
      log_format: json
      log_level: ${TEMPO_LOG_LEVEL}

    tenant_federation:
      enabled: true
      tenant_label: tenant_id
      default_tenant: ${TENANT_ID}_${REGION_ID}_${ENVIRONMENT}

    distributor:
      receivers:
        jaeger:
          protocols:
            thrift_http:
              endpoint: 0.0.0.0:14268
            thrift_binary:
              endpoint: 0.0.0.0:6832
            thrift_compact:
              endpoint: 0.0.0.0:6831
            grpc:
              endpoint: 0.0.0.0:14250
        zipkin:
          endpoint: 0.0.0.0:9411
        otlp:
          protocols:
            grpc:
              endpoint: 0.0.0.0:4317
            http:
              endpoint: 0.0.0.0:4318
        opencensus:
          endpoint: 0.0.0.0:55678
      log_received_spans:
        enabled: ${TEMPO_LOG_RECEIVED_SPANS}
        format: ${TEMPO_LOG_SPAN_FORMAT}
        filter_by_status_error: true
      queue_size: 10000
      max_recv_msg_size: 10485760  # 10MB
      span_ttl: 1h

    ingester:
      trace_idle_period: 10s
      max_block_duration: 5m
      complete_block_timeout: 30m
      lifecycler:
        ring:
          replication_factor: 1
          kvstore:
            store: inmemory

    compactor:
      compaction:
        block_retention: ${TEMPO_RETENTION_PERIOD}
      ring:
        kvstore:
          store: inmemory

    storage:
      trace:
        backend: local
        local:
          path: /var/tempo/traces
        pool:
          max_workers: 100
          queue_depth: 10000

    query_frontend:
      max_outstanding_per_tenant: 2000
      search:
        max_duration: 24h  # Maximum duration for search
        max_bytes_per_tag_values: 10485760
        span_max_limit: 1000
      timeout: 1m

    metrics_generator:
      registry:
        external_labels:
          source: tempo
          tenant_id: ${TENANT_ID}
          region_id: ${REGION_ID}
          environment: ${ENVIRONMENT}
      storage:
        path: /var/tempo/generator/wal
        remote_write:
          - url: http://prometheus-${TENANT_ID}-${REGION_ID}.observability-${TENANT_ID}-${REGION_ID}.svc.cluster.local:9090/api/v1/write
            send_exemplars: true
      processor:
        service_graphs:
          max_items: 10000
          dimensions:
            - tenant_id
            - region_id
            - environment
            - service
            - span_name
        span_metrics:
          dimensions:
            - tenant_id
            - region_id
            - environment
            - service
            - span_name
            - status_code
          metrics_type:
            - latency
            - call_count
          histogram:
            explicit_boundaries_ms:
              - 1
              - 2
              - 5
              - 10
              - 20
              - 50
              - 100
              - 200
              - 500
              - 1000
              - 2000
              - 5000
              - 10000
            max_value_ms: 60000
            min_value_ms: 0

    overrides:
      ${TENANT_ID}_${REGION_ID}_${ENVIRONMENT}:
        metrics_generator_processors:
          - service-graphs
          - span-metrics
        metrics_generator_max_active_series: 1000
        ingestion_rate_strategy: local
        ingestion_rate_limit_bytes: 15_000_000
        ingestion_burst_size_bytes: 30_000_000
```

## Secret

```yaml
# tempo-secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: tempo-secrets
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: tempo
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "tracing"
type: Opaque
stringData:
  tempo-auth-token: "${TEMPO_AUTH_TOKEN}"
```

## PersistentVolumeClaim

```yaml
# tempo-pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: tempo-storage
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: tempo
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "tracing"
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: "${STORAGE_CLASS_NAME}"
  resources:
    requests:
      storage: ${TEMPO_STORAGE_SIZE}  # ex: 50Gi para produção
```

## ServiceAccount e RBAC

```yaml
# tempo-rbac.yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: tempo
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: tempo
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "tracing"
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: tempo
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: tempo
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "tracing"
rules:
- apiGroups: [""]
  resources: ["configmaps", "secrets"]
  verbs: ["get", "watch", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: tempo
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: tempo
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "tracing"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: tempo
subjects:
- kind: ServiceAccount
  name: tempo
  namespace: observability-${TENANT_ID}-${REGION_ID}
```

## StatefulSet

```yaml
# tempo-statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: tempo
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: tempo
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "tracing"
spec:
  replicas: ${TEMPO_REPLICAS}  # Ajuste baseado no ambiente (1 para dev/staging, 3+ para production)
  serviceName: tempo
  podManagementPolicy: Parallel
  updateStrategy:
    type: RollingUpdate
  selector:
    matchLabels:
      app: tempo
  template:
    metadata:
      labels:
        app: tempo
        innovabiz.com/tenant: "${TENANT_ID}"
        innovabiz.com/region: "${REGION_ID}"
        innovabiz.com/environment: "${ENVIRONMENT}"
        innovabiz.com/component: "observability"
        innovabiz.com/module: "tracing"
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "3200"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: tempo
      securityContext:
        fsGroup: 10001
        runAsGroup: 10001
        runAsNonRoot: true
        runAsUser: 10001
      terminationGracePeriodSeconds: 300
      initContainers:
        - name: init-chown-data
          image: busybox:1.35
          imagePullPolicy: IfNotPresent
          command: ['chown', '-R', '10001:10001', '/var/tempo']
          securityContext:
            runAsNonRoot: false
            runAsUser: 0
          volumeMounts:
            - name: storage
              mountPath: /var/tempo
      containers:
        - name: tempo
          image: grafana/tempo:${TEMPO_VERSION}  # Use versões específicas, ex: 2.2.3
          imagePullPolicy: IfNotPresent
          args:
            - -config.file=/etc/tempo/tempo.yaml
          env:
            - name: TENANT_ID
              value: "${TENANT_ID}"
            - name: REGION_ID
              value: "${REGION_ID}"
            - name: ENVIRONMENT
              value: "${ENVIRONMENT}"
          ports:
            - name: http
              containerPort: 3200
              protocol: TCP
            - name: grpc
              containerPort: 9096
              protocol: TCP
            - name: jaeger-thrift
              containerPort: 14268
              protocol: TCP
            - name: jaeger-binary
              containerPort: 6832
              protocol: UDP
            - name: jaeger-compact
              containerPort: 6831
              protocol: UDP
            - name: jaeger-grpc
              containerPort: 14250
              protocol: TCP
            - name: zipkin
              containerPort: 9411
              protocol: TCP
            - name: otlp-grpc
              containerPort: 4317
              protocol: TCP
            - name: otlp-http
              containerPort: 4318
              protocol: TCP
            - name: opencensus
              containerPort: 55678
              protocol: TCP
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
            readOnlyRootFilesystem: true
          volumeMounts:
            - name: config
              mountPath: /etc/tempo
            - name: storage
              mountPath: /var/tempo
          livenessProbe:
            httpGet:
              path: /ready
              port: http
            initialDelaySeconds: 45
            timeoutSeconds: 5
            failureThreshold: 10
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /ready
              port: http
            initialDelaySeconds: 45
            timeoutSeconds: 5
            periodSeconds: 10
          resources:
            limits:
              cpu: ${TEMPO_CPU_LIMIT}  # ex: 1000m
              memory: ${TEMPO_MEMORY_LIMIT}  # ex: 2Gi
            requests:
              cpu: ${TEMPO_CPU_REQUEST}  # ex: 200m
              memory: ${TEMPO_MEMORY_REQUEST}  # ex: 1Gi
      nodeSelector:
        kubernetes.io/os: linux
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            - labelSelector:
                matchExpressions:
                  - key: app
                    operator: In
                    values:
                      - tempo
              topologyKey: "kubernetes.io/hostname"
      volumes:
        - name: config
          configMap:
            name: tempo-config
  volumeClaimTemplates:
    - metadata:
        name: storage
        labels:
          app: tempo
          innovabiz.com/tenant: "${TENANT_ID}"
          innovabiz.com/region: "${REGION_ID}"
          innovabiz.com/environment: "${ENVIRONMENT}"
      spec:
        accessModes: [ "ReadWriteOnce" ]
        storageClassName: "${STORAGE_CLASS_NAME}"
        resources:
          requests:
            storage: ${TEMPO_STORAGE_SIZE}  # ex: 50Gi
```## Service

```yaml
# tempo-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: tempo
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: tempo
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "tracing"
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "3200"
    prometheus.io/path: "/metrics"
spec:
  type: ClusterIP
  ports:
    - name: http
      port: 3200
      targetPort: http
      protocol: TCP
    - name: grpc
      port: 9096
      targetPort: grpc
      protocol: TCP
    - name: jaeger-thrift
      port: 14268
      targetPort: jaeger-thrift
      protocol: TCP
    - name: jaeger-binary
      port: 6832
      targetPort: jaeger-binary
      protocol: UDP
    - name: jaeger-compact
      port: 6831
      targetPort: jaeger-compact
      protocol: UDP
    - name: jaeger-grpc
      port: 14250
      targetPort: jaeger-grpc
      protocol: TCP
    - name: zipkin
      port: 9411
      targetPort: zipkin
      protocol: TCP
    - name: otlp-grpc
      port: 4317
      targetPort: otlp-grpc
      protocol: TCP
    - name: otlp-http
      port: 4318
      targetPort: otlp-http
      protocol: TCP
    - name: opencensus
      port: 55678
      targetPort: opencensus
      protocol: TCP
  selector:
    app: tempo
```

## Ingress

```yaml
# tempo-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: tempo
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: tempo
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "tracing"
  annotations:
    # TLS e certificado
    cert-manager.io/cluster-issuer: "${CLUSTER_ISSUER}"
    # Configurações de segurança
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/auth-type: "basic"
    nginx.ingress.kubernetes.io/auth-secret: "tempo-basic-auth"
    nginx.ingress.kubernetes.io/auth-realm: "Authentication Required"
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "300"
    nginx.ingress.kubernetes.io/proxy-buffer-size: "16k"
    # Segurança adicional
    nginx.ingress.kubernetes.io/configuration-snippet: |
      more_set_headers "X-Frame-Options: SAMEORIGIN";
      more_set_headers "X-Content-Type-Options: nosniff";
      more_set_headers "X-XSS-Protection: 1; mode=block";
      more_set_headers "Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:";
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - tempo-${TENANT_ID}-${REGION_ID}.${DOMAIN}
      secretName: tempo-tls-${TENANT_ID}-${REGION_ID}
  rules:
    - host: tempo-${TENANT_ID}-${REGION_ID}.${DOMAIN}
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: tempo
                port:
                  name: http
```

## NetworkPolicy

```yaml
# tempo-networkpolicy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: tempo-network-policy
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: tempo
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "tracing"
spec:
  podSelector:
    matchLabels:
      app: tempo
  policyTypes:
    - Ingress
    - Egress
  ingress:
    # Permitir ingress dos ingress controllers
    - from:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: ingress-nginx
      ports:
        - port: 3200
          protocol: TCP
    # Permitir ingress dos coletores OpenTelemetry e clientes
    - from:
        - namespaceSelector:
            matchLabels:
              innovabiz.com/component: "observability"
        - podSelector:
            matchLabels:
              app: opentelemetry-collector
      ports:
        - port: 4317  # OTLP gRPC
          protocol: TCP
        - port: 4318  # OTLP HTTP
          protocol: TCP
        - port: 9411  # Zipkin
          protocol: TCP
        - port: 14268  # Jaeger Thrift HTTP
          protocol: TCP
        - port: 14250  # Jaeger gRPC
          protocol: TCP
        - port: 6831  # Jaeger Thrift Compact
          protocol: UDP
        - port: 6832  # Jaeger Thrift Binary
          protocol: UDP
        - port: 55678  # OpenCensus
          protocol: TCP
    # Permitir ingress do Grafana
    - from:
        - namespaceSelector:
            matchLabels:
              innovabiz.com/component: "observability"
        - podSelector:
            matchLabels:
              app: grafana
      ports:
        - port: 3200
          protocol: TCP
    # Permitir ingress do Prometheus para métricas
    - from:
        - namespaceSelector:
            matchLabels:
              innovabiz.com/component: "observability"
        - podSelector:
            matchLabels:
              app: prometheus
      ports:
        - port: 3200
          protocol: TCP
  egress:
    # Permitir egress para DNS
    - to:
        - namespaceSelector: {}
          podSelector:
            matchLabels:
              k8s-app: kube-dns
      ports:
        - port: 53
          protocol: UDP
        - port: 53
          protocol: TCP
    # Permitir egress para Prometheus
    - to:
        - namespaceSelector:
            matchLabels:
              innovabiz.com/component: "observability"
        - podSelector:
            matchLabels:
              app: prometheus
      ports:
        - port: 9090
          protocol: TCP
    # Permitir egress para outros pods Tempo (comunicação interna)
    - to:
        - podSelector:
            matchLabels:
              app: tempo
      ports:
        - port: 9096  # gRPC
          protocol: TCP
```

## PodDisruptionBudget

```yaml
# tempo-pdb.yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: tempo-pdb
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: tempo
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "tracing"
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: tempo
```

## ServiceMonitor

```yaml
# tempo-servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: tempo
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: tempo
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "tracing"
    release: prometheus  # Importante para ser descoberto pelo Prometheus Operator
spec:
  endpoints:
    - port: http
      path: /metrics
      interval: 30s
      scrapeTimeout: 10s
      metricRelabelings:
        # Adiciona labels de tenant, região e ambiente
        - sourceLabels: [__name__]
          targetLabel: tenant_id
          replacement: "${TENANT_ID}"
        - sourceLabels: [__name__]
          targetLabel: region_id
          replacement: "${REGION_ID}"
        - sourceLabels: [__name__]
          targetLabel: environment
          replacement: "${ENVIRONMENT}"
  selector:
    matchLabels:
      app: tempo
  namespaceSelector:
    matchNames:
      - observability-${TENANT_ID}-${REGION_ID}
```## Checklist de Validação

Antes de implantar o Tempo em produção, verifique se os seguintes requisitos foram atendidos:

### Multi-dimensionalidade e Isolamento

- [ ] **Namespace isolado** configurado com o padrão `observability-${TENANT_ID}-${REGION_ID}`
- [ ] **Labels multi-dimensionais** aplicadas em todos os recursos:
  - `innovabiz.com/tenant`: Identificador do tenant
  - `innovabiz.com/region`: Região de operação
  - `innovabiz.com/environment`: Ambiente (dev, qa, staging, prod)
  - `innovabiz.com/component`: "observability"
  - `innovabiz.com/module`: "tracing"
- [ ] **Federação de tenants** configurada no Tempo para suportar múltiplos contextos
- [ ] **Métricas enriquecidas** com dimensões de tenant, região e ambiente
- [ ] **NetworkPolicy** implementada para isolar o tráfego entre tenants

### Segurança

- [ ] **ServiceAccount** dedicado com permissões mínimas necessárias
- [ ] **RBAC** configurado com princípio de privilégio mínimo
- [ ] **SecurityContext** do pod configurado para:
  - Execução como usuário não-root (10001:10001)
  - Sistema de arquivos raiz somente leitura
  - Desativação de privilege escalation
  - Remoção de todas as capabilities desnecessárias
- [ ] **Ingress** protegido com:
  - TLS habilitado
  - Autenticação básica
  - Headers de segurança apropriados (X-Frame-Options, CSP, etc.)
- [ ] **NetworkPolicy** restritiva para tráfego de entrada e saída
- [ ] **Secrets** armazenando dados sensíveis (tokens de autenticação)

### Alta Disponibilidade e Resiliência

- [ ] **Replicas** configuradas de acordo com o ambiente e carga de trabalho
- [ ] **PodDisruptionBudget** para garantir disponibilidade durante atualizações
- [ ] **Anti-afinidade** configurada para distribuir pods em diferentes nós
- [ ] **PersistentVolumeClaim** para armazenamento durável de dados
- [ ] **Resources (CPU/Memory)** configurados adequadamente
- [ ] **Probes de liveness e readiness** implementadas corretamente
- [ ] **gracePeriod** de terminação adequado configurado

### Observabilidade do próprio Tempo

- [ ] **ServiceMonitor** para coletar métricas com o Prometheus Operator
- [ ] **Annotations** para scraping do Prometheus adicionadas
- [ ] **Log format** configurado como JSON para facilitar a análise
- [ ] **Labels de relabeling** para adicionar contexto multidimensional às métricas

### Integração

- [ ] **Portas** para todos os formatos de trace expostas (OTLP, Jaeger, Zipkin)
- [ ] **Service** configurado para expor todas as portas necessárias
- [ ] **Integração com Prometheus** para remote_write configurada
- [ ] **Gerador de métricas** configurado para service graphs e span metrics
- [ ] **NetworkPolicy** permite comunicação com OpenTelemetry Collector, Grafana e Prometheus

### Retenção e Escalabilidade

- [ ] **Período de retenção** configurado de acordo com a política de retenção de dados
- [ ] **Storage size** dimensionado com base no volume de traces esperado
- [ ] **Limites de ingestion rate** configurados por tenant

## Melhores Práticas

### Design de Sistema de Tracing

1. **Estratégia de Amostragem**
   - Implemente amostragem baseada em caudas para capturar traces mais lentos
   - Configure taxas de amostragem diferentes por serviço e ambiente
   - Assegure que erros sejam sempre capturados, independentemente da amostragem

2. **Contextualização de Traces**
   - Propague contexto sempre incluindo tenant_id, region_id, environment
   - Implemente propagadores consistentes em toda a plataforma (W3C TraceContext recomendado)
   - Adicione atributos padronizados para facilitar consultas (módulo, serviço, operação)

3. **Qualidade de Dados**
   - Defina uma nomenclatura consistente para serviços e spans
   - Implemente spans com semântica padronizada (OpenTelemetry Semantic Conventions)
   - Enriqueça spans com atributos relevantes para o negócio

4. **Geração de Métricas**
   - Ative o metrics_generator para gerar RED metrics (Rate, Error, Duration) e service graphs
   - Configure dimensions consistentes para métricas derivadas de traces
   - Use métricas para alertas e traces para investigação detalhada

### Operação em Produção

1. **Escalabilidade**
   - Monitore a utilização de recursos e dimensione o Tempo de acordo
   - Considere a estratégia de retenção em relação ao espaço de armazenamento
   - Implemente limites de ingestion rate por tenant para evitar "noisy neighbor"

2. **Monitoramento**
   - Configure alertas para falhas no Tempo e métricas de saúde
   - Monitore métricas-chave:
     - Latência de ingestão
     - Erro de ingestion rate
     - Uso de recursos (CPU, memória, disco)
     - Falhas de write/read
     - Latência de consulta

3. **Backups e DR**
   - Implemente backup periódico do armazenamento
   - Considere replicação geográfica para dados críticos
   - Documente procedimentos de DR

### Integração com Dashboard Grafana

1. **Datasource Padrão**
   - Configure o Tempo como fonte de dados padrão para traces no Grafana
   - Habilite relações Trace-to-Logs e Trace-to-Metrics

2. **Visualizações Recomendadas**
   - Service Graph para visualização das dependências
   - Red metrics para monitorar performance por serviço
   - Dashboard de operação do Tempo

3. **Drill-down Integrado**
   - Configure links de métricas para traces relacionados
   - Habilite correlação entre logs, métricas e traces

### Segurança e Compliance

1. **Proteção de Dados**
   - Assegure que dados sensíveis não sejam incluídos em spans
   - Implemente uma política de sanitização para atributos de traces
   - Configure períodos de retenção de acordo com LGPD/GDPR

2. **Auditoria**
   - Registre ações administrativas nas configurações do Tempo
   - Monitore tentativas de acesso não autorizado

3. **Isolamento**
   - Garanta estrito isolamento multi-tenant através de federação
   - Aplique limites de recursos por tenant

## Exemplos de Uso

### Preparação de Variáveis de Ambiente

```bash
# Variáveis de contexto multi-dimensional
export TENANT_ID="tenant001"
export REGION_ID="br-sp"
export ENVIRONMENT="prod"
export DOMAIN="observability.innovabiz.com"

# Configurações de armazenamento
export STORAGE_CLASS_NAME="premium-rwo"
export TEMPO_STORAGE_SIZE="100Gi"

# Configurações de recursos
export TEMPO_REPLICAS="3"
export TEMPO_CPU_REQUEST="500m"
export TEMPO_CPU_LIMIT="2000m"
export TEMPO_MEMORY_REQUEST="2Gi"
export TEMPO_MEMORY_LIMIT="4Gi"

# Configurações do Tempo
export TEMPO_VERSION="2.2.3"
export TEMPO_AUTH_ENABLED="true"
export TEMPO_AUTH_TOKEN=$(openssl rand -hex 16)
export TEMPO_LOG_LEVEL="info"
export TEMPO_LOG_RECEIVED_SPANS="true"
export TEMPO_LOG_SPAN_FORMAT="json"
export TEMPO_RETENTION_PERIOD="336h" # 14 dias

# Configurações de segurança
export CLUSTER_ISSUER="letsencrypt-prod"
```

### Geração de Arquivos com Envsubst

```bash
# Crie um arquivo de secrets para autenticação básica
TEMPO_BASIC_AUTH_USER="admin"
TEMPO_BASIC_AUTH_PASS=$(openssl rand -base64 12)
TEMPO_BASIC_AUTH=$(htpasswd -nb $TEMPO_BASIC_AUTH_USER $TEMPO_BASIC_AUTH_PASS)
kubectl create secret generic tempo-basic-auth -n observability-${TENANT_ID}-${REGION_ID} \
  --from-literal=auth="$TEMPO_BASIC_AUTH"

# Gere e aplique cada arquivo YAML
for template in namespace.yaml tempo-configmap.yaml tempo-secrets.yaml tempo-pvc.yaml \
                tempo-rbac.yaml tempo-statefulset.yaml tempo-service.yaml tempo-ingress.yaml \
                tempo-networkpolicy.yaml tempo-pdb.yaml tempo-servicemonitor.yaml; do
  envsubst < $template | kubectl apply -f -
done
```

### Validação da Instalação

```bash
# Verifique a implantação do Tempo
kubectl get statefulset -n observability-${TENANT_ID}-${REGION_ID}
kubectl get pods -n observability-${TENANT_ID}-${REGION_ID} -l app=tempo

# Verifique os serviços
kubectl get svc -n observability-${TENANT_ID}-${REGION_ID} -l app=tempo

# Verifique se o ingress está configurado corretamente
kubectl get ingress -n observability-${TENANT_ID}-${REGION_ID} -l app=tempo

# Teste o acesso ao Tempo (substitua a URL conforme necessário)
curl -u "${TEMPO_BASIC_AUTH_USER}:${TEMPO_BASIC_AUTH_PASS}" \
  https://tempo-${TENANT_ID}-${REGION_ID}.${DOMAIN}/api/status
```

### Integração do Tempo com OpenTelemetry Collector

```yaml
# Exemplo de configuração de OpenTelemetry Collector para envio de traces para o Tempo
apiVersion: v1
kind: ConfigMap
metadata:
  name: otel-collector-config
  namespace: observability-${TENANT_ID}-${REGION_ID}
data:
  otel-collector-config.yaml: |
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
      attributes:
        actions:
          - key: tenant_id
            value: "${TENANT_ID}"
            action: upsert
          - key: region_id
            value: "${REGION_ID}"
            action: upsert
          - key: environment
            value: "${ENVIRONMENT}"
            action: upsert

    exporters:
      otlp:
        endpoint: "tempo.observability-${TENANT_ID}-${REGION_ID}.svc.cluster.local:4317"
        tls:
          insecure: true
        headers:
          Authorization: "Bearer ${TEMPO_AUTH_TOKEN}"

    service:
      telemetry:
        logs:
          level: "info"
      pipelines:
        traces:
          receivers: [otlp]
          processors: [attributes, batch]
          exporters: [otlp]
```

### Configuração da Fonte de Dados no Grafana

```yaml
# Exemplo de ConfigMap para configuração de datasource Tempo no Grafana
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-tempo-datasource
  namespace: observability-${TENANT_ID}-${REGION_ID}
data:
  tempo-datasource.yaml: |
    apiVersion: 1
    datasources:
      - name: Tempo-${TENANT_ID}-${REGION_ID}
        type: tempo
        access: proxy
        url: http://tempo.observability-${TENANT_ID}-${REGION_ID}.svc.cluster.local:3200
        version: 1
        isDefault: false
        jsonData:
          httpMethod: GET
          tracesToLogs:
            datasourceUid: loki-${TENANT_ID}-${REGION_ID}
            tags: ['tenant_id', 'region_id', 'environment', 'service.name', 'k8s.namespace.name', 'k8s.pod.name']
            mappedTags: [{ key: 'service.name', value: 'service' }]
            mapTagNamesEnabled: true
            spanStartTimeShift: '-1h'
            spanEndTimeShift: '1h'
            filterByTraceID: true
            filterBySpanID: false
          tracesToMetrics:
            datasourceUid: prometheus-${TENANT_ID}-${REGION_ID}
            spanStartTimeShift: '-1h'
            spanEndTimeShift: '1h'
            tags: [{ key: 'service.name', value: 'service' }, { key: 'span.name', value: 'operation' }]
            queries:
              - name: 'Request Rate'
                query: 'sum(rate(tempo_spanmetrics_calls_total{service="$service",operation="$operation"}[$__rate_interval]))'
              - name: 'Error Rate'
                query: 'sum(rate(tempo_spanmetrics_calls_total{service="$service",operation="$operation",status_code="STATUS_CODE_ERROR"}[$__rate_interval])) / sum(rate(tempo_spanmetrics_calls_total{service="$service",operation="$operation"}[$__rate_interval]))'
              - name: 'Latency'
                query: 'histogram_quantile(0.95, sum(rate(tempo_spanmetrics_latency_bucket{service="$service",operation="$operation"}[$__rate_interval])) by (le))'
          serviceMap:
            datasourceUid: prometheus-${TENANT_ID}-${REGION_ID}
        secureJsonData:
          httpHeaderValue1: "${TEMPO_AUTH_TOKEN}"
```

Com este template completo, a equipe INNOVABIZ pode implantar e operar o Grafana Tempo como solução de rastreamento distribuído com todas as garantias de segurança, isolamento multi-dimensional e melhores práticas operacionais para ambientes de alta disponibilidade.