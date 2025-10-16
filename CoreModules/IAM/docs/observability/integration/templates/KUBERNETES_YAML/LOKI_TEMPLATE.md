# Template Kubernetes YAML para Loki - INNOVABIZ

## Visão Geral

Este documento fornece templates YAML para implantação do Loki no ambiente Kubernetes da plataforma INNOVABIZ. O Loki é um sistema de agregação de logs inspirado no Prometheus, otimizado para armazenar e consultar logs de forma eficiente. Os templates seguem as melhores práticas de segurança, dimensionamento e configuração multi-dimensional conforme os padrões INNOVABIZ.

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
    innovabiz.com/module: "logging"
```

## ConfigMap

```yaml
# loki-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: loki-config
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: loki
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "logging"
data:
  loki.yaml: |
    auth_enabled: ${LOKI_AUTH_ENABLED}

    server:
      http_listen_port: 3100
      grpc_listen_port: 9095
      http_server_read_timeout: 120s
      http_server_write_timeout: 120s
      grpc_server_max_recv_msg_size: 8388608
      grpc_server_max_send_msg_size: 8388608
      log_format: json
      log_level: ${LOKI_LOG_LEVEL}

    common:
      path_prefix: /data/loki
      storage:
        filesystem:
          chunks_directory: /data/loki/chunks
          rules_directory: /data/loki/rules
      replication_factor: 1
      ring:
        instance_addr: 127.0.0.1
        kvstore:
          store: inmemory

    schema_config:
      configs:
        - from: ${LOKI_START_DATE}
          store: boltdb-shipper
          object_store: filesystem
          schema: v12
          index:
            prefix: index_
            period: 24h

    tenant_federation:
      enabled: true
      tenant_label: tenant_id
      default_tenant: ${TENANT_ID}_${REGION_ID}_${ENVIRONMENT}

    limits_config:
      enforce_metric_name: false
      max_entries_limit_per_query: 10000
      max_global_streams_per_user: 10000
      max_query_length: 24h
      max_query_parallelism: 32
      max_streams_per_user: 10000
      reject_old_samples: true
      reject_old_samples_max_age: 24h
      split_queries_by_interval: 30m
      ingestion_rate_mb: 8
      ingestion_burst_size_mb: 16
      retention_period: ${LOKI_RETENTION_PERIOD}

    chunk_store_config:
      max_look_back_period: ${LOKI_RETENTION_PERIOD}

    table_manager:
      retention_deletes_enabled: true
      retention_period: ${LOKI_RETENTION_PERIOD}

    ruler:
      storage:
        type: local
        local:
          directory: /data/loki/rules
      rule_path: /data/loki/rules
      alertmanager_url: http://alertmanager.observability-${TENANT_ID}-${REGION_ID}.svc.cluster.local:9093
      ring:
        kvstore:
          store: inmemory
      enable_api: true

    analytics:
      reporting_enabled: false

    compactor:
      working_directory: /data/loki/compactor
      shared_store: filesystem
      compaction_interval: 10m
      retention_enabled: true
      retention_delete_delay: 2h
      retention_delete_worker_count: 150

    frontend:
      log_queries_longer_than: 10s
      compress_responses: true
      max_outstanding_per_tenant: 4096

    frontend_worker:
      frontend_address: loki:9095
      parallelism: 16

    ingester:
      chunk_idle_period: 10m
      chunk_retain_period: 1m
      chunk_target_size: 1048576
      lifecycler:
        ring:
          kvstore:
            store: inmemory
          replication_factor: 1
      wal:
        enabled: true
        dir: /data/loki/wal

    distributor:
      ring:
        kvstore:
          store: inmemory

    querier:
      max_concurrent: 16
      engine:
        timeout: 3m

    query_range:
      results_cache:
        cache:
          enable_fifocache: true
          fifocache:
            size: 1024
            validity: 24h

    storage_config:
      boltdb_shipper:
        active_index_directory: /data/loki/index
        cache_location: /data/loki/boltdb-cache
        cache_ttl: 24h
        shared_store: filesystem
        resync_interval: 5s
```

## Secret

```yaml
# loki-secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: loki-secrets
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: loki
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "logging"
type: Opaque
stringData:
  loki-auth-token: "${LOKI_AUTH_TOKEN}"
```

## PersistentVolumeClaim

```yaml
# loki-pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: loki-storage
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: loki
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "logging"
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: "${STORAGE_CLASS_NAME}"
  resources:
    requests:
      storage: ${LOKI_STORAGE_SIZE}  # ex: 50Gi para produção
```## ServiceAccount e RBAC

```yaml
# loki-rbac.yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: loki
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: loki
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "logging"
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: loki
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: loki
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "logging"
rules:
- apiGroups: [""]
  resources: ["configmaps", "secrets"]
  verbs: ["get", "watch", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: loki
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: loki
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "logging"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: loki
subjects:
- kind: ServiceAccount
  name: loki
  namespace: observability-${TENANT_ID}-${REGION_ID}
```

## StatefulSet

```yaml
# loki-statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: loki
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: loki
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "logging"
spec:
  replicas: ${LOKI_REPLICAS}  # Ajuste baseado no ambiente (1 para dev/staging, 3+ para production)
  serviceName: loki
  podManagementPolicy: Parallel
  updateStrategy:
    type: RollingUpdate
  selector:
    matchLabels:
      app: loki
  template:
    metadata:
      labels:
        app: loki
        innovabiz.com/tenant: "${TENANT_ID}"
        innovabiz.com/region: "${REGION_ID}"
        innovabiz.com/environment: "${ENVIRONMENT}"
        innovabiz.com/component: "observability"
        innovabiz.com/module: "logging"
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "3100"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: loki
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
          command: ['chown', '-R', '10001:10001', '/data']
          securityContext:
            runAsNonRoot: false
            runAsUser: 0
          volumeMounts:
            - name: storage
              mountPath: /data
      containers:
        - name: loki
          image: grafana/loki:${LOKI_VERSION}  # Use versões específicas, ex: 2.9.0
          imagePullPolicy: IfNotPresent
          args:
            - -config.file=/etc/loki/loki.yaml
          env:
            - name: JAEGER_AGENT_HOST
              value: tempo-${TENANT_ID}-${REGION_ID}
            - name: JAEGER_AGENT_PORT
              value: "6831"
            - name: JAEGER_SAMPLER_TYPE
              value: const
            - name: JAEGER_SAMPLER_PARAM
              value: "1"
            - name: JAEGER_TAGS
              value: "tenant=${TENANT_ID},region=${REGION_ID},environment=${ENVIRONMENT}"
            - name: TENANT_ID
              value: "${TENANT_ID}"
            - name: REGION_ID
              value: "${REGION_ID}"
            - name: ENVIRONMENT
              value: "${ENVIRONMENT}"
          ports:
            - name: http
              containerPort: 3100
              protocol: TCP
            - name: grpc
              containerPort: 9095
              protocol: TCP
            - name: memberlist
              containerPort: 7946
              protocol: TCP
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
            readOnlyRootFilesystem: true
          volumeMounts:
            - name: config
              mountPath: /etc/loki
            - name: storage
              mountPath: /data
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
              cpu: ${LOKI_CPU_LIMIT}  # ex: 1000m
              memory: ${LOKI_MEMORY_LIMIT}  # ex: 2Gi
            requests:
              cpu: ${LOKI_CPU_REQUEST}  # ex: 200m
              memory: ${LOKI_MEMORY_REQUEST}  # ex: 1Gi
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
                      - loki
              topologyKey: "kubernetes.io/hostname"
      volumes:
        - name: config
          configMap:
            name: loki-config
  volumeClaimTemplates:
    - metadata:
        name: storage
        labels:
          app: loki
          innovabiz.com/tenant: "${TENANT_ID}"
          innovabiz.com/region: "${REGION_ID}"
          innovabiz.com/environment: "${ENVIRONMENT}"
      spec:
        accessModes: [ "ReadWriteOnce" ]
        storageClassName: "${STORAGE_CLASS_NAME}"
        resources:
          requests:
            storage: ${LOKI_STORAGE_SIZE}  # ex: 50Gi
```

## Service

```yaml
# loki-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: loki
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: loki
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "logging"
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "3100"
spec:
  type: ClusterIP
  ports:
    - port: 3100
      targetPort: http
      protocol: TCP
      name: http
    - port: 9095
      targetPort: grpc
      protocol: TCP
      name: grpc
  selector:
    app: loki
```

## Ingress

```yaml
# loki-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: loki
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: loki
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "logging"
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-connect-timeout: "300"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "300"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "300"
    nginx.ingress.kubernetes.io/proxy-body-size: "500m"
    cert-manager.io/cluster-issuer: "${CLUSTER_ISSUER}"
    nginx.ingress.kubernetes.io/auth-type: basic
    nginx.ingress.kubernetes.io/auth-secret: loki-basic-auth
    nginx.ingress.kubernetes.io/auth-realm: "Authentication Required"
spec:
  tls:
    - hosts:
        - loki-${TENANT_ID}-${REGION_ID}.${DOMAIN}
      secretName: loki-tls
  rules:
    - host: loki-${TENANT_ID}-${REGION_ID}.${DOMAIN}
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: loki
                port:
                  name: http
```

## NetworkPolicy

```yaml
# loki-network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: loki
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: loki
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "logging"
spec:
  podSelector:
    matchLabels:
      app: loki
  policyTypes:
    - Ingress
    - Egress
  ingress:
    # Permitir tráfego de entrada HTTP para API do Loki
    - ports:
        - port: 3100
          protocol: TCP
      from:
        # Permitir tráfego de ingress-controllers
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: ingress-nginx
        # Permitir tráfego de pods no mesmo namespace
        - podSelector: {}
        # Permitir tráfego de Promtail/Fluentbit/Vector de todos os namespaces
        - namespaceSelector: {}
          podSelector:
            matchLabels:
              app.kubernetes.io/name: promtail
        - namespaceSelector: {}
          podSelector:
            matchLabels:
              app.kubernetes.io/name: fluent-bit
    
    # Permitir tráfego de entrada gRPC
    - ports:
        - port: 9095
          protocol: TCP
      from:
        # Permitir tráfego de pods no mesmo namespace (distributor)
        - podSelector: {}
    
    # Permitir tráfego de entrada para memberlist
    - ports:
        - port: 7946
          protocol: TCP
      from:
        # Permitir apenas de outros Loki
        - podSelector:
            matchLabels:
              app: loki
  
  egress:
    # Permitir acesso ao Alertmanager
    - to:
        - podSelector:
            matchLabels:
              app: alertmanager
      ports:
        - port: 9093
          protocol: TCP
    
    # Permitir acesso ao Tempo
    - to:
        - podSelector:
            matchLabels:
              app: tempo
      ports:
        - port: 6831
          protocol: UDP
    
    # Permitir acesso a Servidores DNS
    - to:
        - namespaceSelector: {}
          podSelector:
            matchLabels:
              k8s-app: kube-dns
      ports:
        - port: 53
          protocol: UDP
    
    # Permitir acesso para comunicação entre instâncias Loki
    - to:
        - podSelector:
            matchLabels:
              app: loki
      ports:
        - port: 7946
          protocol: TCP
        - port: 3100
          protocol: TCP
        - port: 9095
          protocol: TCP
```

## PodDisruptionBudget

```yaml
# loki-pdb.yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: loki
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: loki
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "logging"
spec:
  maxUnavailable: 1
  selector:
    matchLabels:
      app: loki
```

## ServiceMonitor

```yaml
# loki-service-monitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: loki
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: loki
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "logging"
spec:
  selector:
    matchLabels:
      app: loki
  endpoints:
    - port: http
      path: /metrics
      interval: 30s
      scrapeTimeout: 10s
      honorLabels: true
      metricRelabelings:
        - targetLabel: tenant_id
          replacement: "${TENANT_ID}"
        - targetLabel: region_id
          replacement: "${REGION_ID}"
        - targetLabel: environment
          replacement: "${ENVIRONMENT}"
```

## Checklist de Validação

Use o checklist abaixo para validar a implementação do Loki na plataforma INNOVABIZ:

### Multi-dimensionalidade

- [ ] **Isolamento por Tenant**
  - [ ] Namespace específico por tenant e região
  - [ ] Labels de tenant/região em todos os recursos
  - [ ] Configuração de tenant federation habilitada
  - [ ] Separação de buckets de armazenamento por tenant
  
- [ ] **Contexto Regional**
  - [ ] Configurações específicas por região
  - [ ] Integração com serviços regionais
  - [ ] Tracing regional configurado
  
- [ ] **Separação de Ambientes**
  - [ ] Retenção de logs específica por ambiente
  - [ ] Recursos escalados conforme ambiente
  - [ ] Nível de log apropriado ao ambiente

### Segurança

- [ ] **Controle de Acesso**
  - [ ] Autenticação básica para API HTTP
  - [ ] TLS para Ingress
  - [ ] Acesso restrito por NetworkPolicy
  
- [ ] **Segurança de Container**
  - [ ] Container executando como não-root
  - [ ] Sistema de arquivos somente leitura
  - [ ] Capacidades mínimas necessárias
  
- [ ] **Proteção de Dados**
  - [ ] Persistência configurada adequadamente
  - [ ] Retenção automatizada de logs antigos
  - [ ] Filtragem de dados sensíveis

### Integração com Stack INNOVABIZ

- [ ] **Integração com OpenTelemetry**
  - [ ] Rastreamento via Jaeger/Tempo
  - [ ] Configuração multi-tenant para correlação
  
- [ ] **Integração com Alerting**
  - [ ] Conexão com AlertManager
  - [ ] Regras de alerta para componentes críticos
  
- [ ] **Integração com Visualização**
  - [ ] DataSource configurado no Grafana
  - [ ] Dashboards para análise de logs
  - [ ] Correlação de logs com métricas e traces

## Melhores Práticas

### Configuração e Dimensionamento

1. **Otimização de Consultas**
   - Configure `split_queries_by_interval` apropriadamente para consultas de longa duração
   - Ajuste `max_query_parallelism` com base na capacidade do sistema
   - Implemente caching para consultas frequentes

2. **Gerenciamento de Recursos**
   - Ajuste limites de memória conforme o volume de logs
   - Configure corretamente limites de ingestão por tenant
   - Dimensione armazenamento com base em volume e políticas de retenção

3. **Alta Disponibilidade**
   - Use múltiplas réplicas para tolerância a falhas
   - Configure anti-afinidade para distribuir pods
   - Implemente PodDisruptionBudget para manutenções seguras

### Ingestão de Logs

1. **Estruturação de Logs**
   - Padronize formato JSON para todos os logs
   - Inclua metadados de tenant, região, ambiente, módulo
   - Utilize labels consistentes para filtro eficiente

2. **Roteamento e Filtros**
   - Configure filtros no lado cliente para reduzir volume
   - Implemente compressão para transferência eficiente
   - Utilize shards baseados em tenant para escalar horizontalmente

3. **Controle de Taxa**
   - Implemente rate limiting por tenant
   - Configure políticas de descarte para picos extremos
   - Monitore taxas de ingestão para capacidade e faturamento

### Retenção e Armazenamento

1. **Políticas de Retenção**
   - Configure retenção diferente por ambiente e criticidade
   - Implemente retenção automática para compliance
   - Archive logs importantes para storage de longo prazo

2. **Compactação**
   - Configure compactação para otimizar armazenamento
   - Ajuste intervalos de compactação baseado em padrões de acesso
   - Monitore performance de compactação

3. **Escalabilidade**
   - Utilize PersistentVolumes com StorageClasses apropriadas
   - Considere armazenamento em objeto para grande escala
   - Implemente índices eficientes para pesquisa

## Exemplos de Uso

### Aplicação dos Templates Kubernetes

1. **Prepare as Variáveis de Ambiente**

```bash
# Defina as variáveis multi-dimensionais
export TENANT_ID="tenant1"
export REGION_ID="br"
export ENVIRONMENT="production"

# Defina as variáveis de infraestrutura
export DOMAIN="innovabiz.com"
export STORAGE_CLASS_NAME="standard"
export CLUSTER_ISSUER="letsencrypt-prod"

# Defina as variáveis do Loki
export LOKI_VERSION="2.9.0"
export LOKI_REPLICAS="3"
export LOKI_CPU_LIMIT="2000m"
export LOKI_MEMORY_LIMIT="4Gi"
export LOKI_CPU_REQUEST="500m"
export LOKI_MEMORY_REQUEST="1Gi"
export LOKI_STORAGE_SIZE="100Gi"
export LOKI_AUTH_ENABLED="true"
export LOKI_LOG_LEVEL="info"
export LOKI_START_DATE="2023-01-01"
export LOKI_RETENTION_PERIOD="720h"  # 30 dias

# Defina credenciais (use geração segura em produção)
export LOKI_AUTH_TOKEN="$(openssl rand -base64 32)"
```

2. **Substitua as Variáveis e Aplique os Templates**

```bash
# Crie um diretório temporário
mkdir -p /tmp/loki-deploy

# Copie e substitua variáveis em todos os arquivos
for file in namespace.yaml loki-configmap.yaml loki-secrets.yaml loki-rbac.yaml \
            loki-statefulset.yaml loki-service.yaml loki-ingress.yaml loki-network-policy.yaml \
            loki-pdb.yaml loki-service-monitor.yaml; do
  envsubst < $file > /tmp/loki-deploy/$file
done

# Aplique os recursos na ordem correta
kubectl apply -f /tmp/loki-deploy/namespace.yaml
kubectl apply -f /tmp/loki-deploy/loki-configmap.yaml
kubectl apply -f /tmp/loki-deploy/loki-secrets.yaml
kubectl apply -f /tmp/loki-deploy/loki-rbac.yaml
kubectl apply -f /tmp/loki-deploy/loki-statefulset.yaml
kubectl apply -f /tmp/loki-deploy/loki-service.yaml
kubectl apply -f /tmp/loki-deploy/loki-network-policy.yaml
kubectl apply -f /tmp/loki-deploy/loki-pdb.yaml
kubectl apply -f /tmp/loki-deploy/loki-service-monitor.yaml
kubectl apply -f /tmp/loki-deploy/loki-ingress.yaml

# Limpe os arquivos temporários com credenciais
rm -rf /tmp/loki-deploy
```

3. **Verifique a Implantação**

```bash
# Verifique se todos os recursos foram criados corretamente
kubectl -n observability-${TENANT_ID}-${REGION_ID} get all -l app=loki

# Verifique se os pods do Loki estão em execução
kubectl -n observability-${TENANT_ID}-${REGION_ID} get pods -l app=loki

# Obtenha a URL de acesso
echo "Acesse o Loki em: https://loki-${TENANT_ID}-${REGION_ID}.${DOMAIN}"
```

### Integração com Promtail

Para coletar logs de pods Kubernetes e enviá-los ao Loki, você pode implantar o Promtail com a seguinte configuração:

```yaml
# Trecho do configmap do Promtail
scrape_configs:
  - job_name: kubernetes-pods
    kubernetes_sd_configs:
      - role: pod
    pipeline_stages:
      - json:
          expressions:
            tenant_id: tenant_id
            region_id: region_id
            environment: environment
            module: module
      - labels:
          tenant_id:
          region_id:
          environment:
          module:
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_innovabiz_com_tenant]
        target_label: tenant_id
      - source_labels: [__meta_kubernetes_pod_annotation_innovabiz_com_region]
        target_label: region_id
      - source_labels: [__meta_kubernetes_pod_annotation_innovabiz_com_environment]
        target_label: environment
      - source_labels: [__meta_kubernetes_pod_annotation_innovabiz_com_module]
        target_label: module
```

### Integração com Grafana

Para adicionar o Loki como fonte de dados no Grafana, configure o seguinte datasource:

```yaml
# Trecho do configmap do Grafana para datasource Loki
datasources:
  - name: Loki
    type: loki
    access: proxy
    url: http://loki.observability-${TENANT_ID}-${REGION_ID}.svc.cluster.local:3100
    jsonData:
      maxLines: 1000
      derivedFields:
        - datasourceUid: tempo
          matcherRegex: "traceID=(\\w+)"
          name: TraceID
          url: "$${__value.raw}"
      timeout: 60
```