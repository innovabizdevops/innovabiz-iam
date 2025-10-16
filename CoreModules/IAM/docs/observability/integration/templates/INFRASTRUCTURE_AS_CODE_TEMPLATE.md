# Template de Infraestrutura como Código (IaC) para Observabilidade - INNOVABIZ

## Visão Geral

Este documento fornece templates padronizados INNOVABIZ para Infraestrutura como Código (IaC) relacionada aos componentes de observabilidade. Esses templates seguem as melhores práticas para implantar, configurar e gerenciar componentes de observabilidade em ambientes Kubernetes, considerando os requisitos multi-dimensionais (tenant, região, módulo) da plataforma INNOVABIZ.

## Sumário

1. [Componentes de Observabilidade](#componentes-de-observabilidade)
2. [Templates Terraform](#templates-terraform)
3. [Templates Kubernetes (Helm)](#templates-kubernetes-helm)
4. [Templates Kubernetes (YAML)](#templates-kubernetes-yaml)
5. [Checklist de Validação](#checklist-de-validação)
6. [Melhores Práticas](#melhores-práticas)

## Componentes de Observabilidade

A infraestrutura de observabilidade INNOVABIZ é composta pelos seguintes componentes principais:

| Componente | Função | Template Principal |
|------------|--------|-------------------|
| OpenTelemetry Collector | Coleta, processamento e exportação de telemetria | [YAML](#opentelemetry-collector) / [Helm](#opentelemetry-collector-helm) |
| Prometheus | Armazenamento e consulta de métricas | [YAML](#prometheus) / [Helm](#prometheus-helm) |
| Grafana | Visualização e dashboards | [YAML](#grafana) / [Helm](#grafana-helm) |
| Loki | Armazenamento e consulta de logs | [YAML](#loki) / [Helm](#loki-helm) |
| Tempo | Armazenamento e consulta de traces | [YAML](#tempo) / [Helm](#tempo-helm) |
| AlertManager | Gerenciamento e roteamento de alertas | [YAML](#alertmanager) / [Helm](#alertmanager-helm) |
| Prometheus Operator | CRDs para gerenciamento declarativo | [Helm](#prometheus-operator-helm) |

## Templates Terraform

### Provisionamento de Infraestrutura Base

```hcl
module "observability_infrastructure" {
  source = "git::https://github.com/innovabiz/terraform-modules//observability?ref=v1.0.0"
  
  # Configurações Multi-dimensionais
  tenant_id         = var.tenant_id
  region_id         = var.region_id
  environment       = var.environment
  
  # Configurações do Cluster
  cluster_name      = "${var.tenant_id}-${var.region_id}-${var.environment}"
  namespace         = "observability"
  
  # Dimensionamento e Recursos
  prometheus_storage_size = var.prometheus_storage_size # ex: "100Gi"
  loki_storage_size       = var.loki_storage_size       # ex: "200Gi"
  tempo_storage_size      = var.tempo_storage_size      # ex: "100Gi"
  
  # Rede e Segurança
  ingress_enabled         = var.ingress_enabled
  ingress_domain          = var.ingress_domain
  tls_enabled             = var.tls_enabled
  tls_secret_name         = var.tls_secret_name
  
  # Retenção de Dados
  prometheus_retention    = var.prometheus_retention    # ex: "15d"
  loki_retention          = var.loki_retention          # ex: "30d"
  tempo_retention         = var.tempo_retention         # ex: "7d"
  
  # Integrações
  alertmanager_receivers  = var.alertmanager_receivers
  grafana_oauth_enabled   = var.grafana_oauth_enabled
  grafana_oauth_config    = var.grafana_oauth_config
  
  # Tags
  tags = {
    tenant      = var.tenant_id
    region      = var.region_id
    environment = var.environment
    module      = "observability"
    managed-by  = "terraform"
  }
}
```

### Variables.tf para Observabilidade

```hcl
# Variáveis Multi-dimensionais INNOVABIZ
variable "tenant_id" {
  description = "ID do tenant na plataforma INNOVABIZ"
  type        = string
  validation {
    condition     = can(regex("^[a-z0-9-]{3,16}$", var.tenant_id))
    error_message = "O tenant_id deve conter entre 3 e 16 caracteres alfanuméricos e hífens."
  }
}

variable "region_id" {
  description = "ID da região INNOVABIZ (br, us, eu, ao)"
  type        = string
  validation {
    condition     = contains(["br", "us", "eu", "ao"], var.region_id)
    error_message = "A região deve ser uma das seguintes: br, us, eu, ao."
  }
}

variable "environment" {
  description = "Ambiente (production, staging, development, sandbox)"
  type        = string
  validation {
    condition     = contains(["production", "staging", "development", "sandbox"], var.environment)
    error_message = "O ambiente deve ser um dos seguintes: production, staging, development, sandbox."
  }
}

# Configurações do Cluster
variable "cluster_name" {
  description = "Nome do cluster Kubernetes"
  type        = string
  default     = ""
}

variable "namespace" {
  description = "Namespace Kubernetes para componentes de observabilidade"
  type        = string
  default     = "observability"
}

# Dimensionamento e Recursos
variable "prometheus_storage_size" {
  description = "Tamanho de armazenamento para Prometheus"
  type        = string
  default     = "100Gi"
}

variable "loki_storage_size" {
  description = "Tamanho de armazenamento para Loki"
  type        = string
  default     = "200Gi"
}

variable "tempo_storage_size" {
  description = "Tamanho de armazenamento para Tempo"
  type        = string
  default     = "100Gi"
}

# Rede e Segurança
variable "ingress_enabled" {
  description = "Habilitar Ingress para componentes de observabilidade"
  type        = bool
  default     = true
}

variable "ingress_domain" {
  description = "Domínio base para Ingress de observabilidade"
  type        = string
  default     = "observability.innovabiz.com"
}

variable "tls_enabled" {
  description = "Habilitar TLS para Ingress"
  type        = bool
  default     = true
}

variable "tls_secret_name" {
  description = "Nome do secret TLS para Ingress"
  type        = string
  default     = "observability-tls"
}

# Retenção de Dados
variable "prometheus_retention" {
  description = "Período de retenção de dados do Prometheus"
  type        = string
  default     = "15d"
}

variable "loki_retention" {
  description = "Período de retenção de dados do Loki"
  type        = string
  default     = "30d"
}

variable "tempo_retention" {
  description = "Período de retenção de dados do Tempo"
  type        = string
  default     = "7d"
}

# Integrações
variable "alertmanager_receivers" {
  description = "Configuração de receivers do AlertManager"
  type = list(object({
    name              = string
    slack_configs     = optional(list(object({
      channel         = string
      api_url         = string
      title           = optional(string)
      text            = optional(string)
      send_resolved   = optional(bool)
    })))
    email_configs     = optional(list(object({
      to              = string
      from            = optional(string)
      send_resolved   = optional(bool)
    })))
    webhook_configs   = optional(list(object({
      url             = string
      send_resolved   = optional(bool)
    })))
    pagerduty_configs = optional(list(object({
      service_key     = string
      routing_key     = optional(string)
      send_resolved   = optional(bool)
    })))
  }))
  default = []
}

variable "grafana_oauth_enabled" {
  description = "Habilitar OAuth para Grafana"
  type        = bool
  default     = false
}

variable "grafana_oauth_config" {
  description = "Configuração OAuth para Grafana"
  type        = object({
    client_id         = string
    client_secret     = string
    auth_url          = string
    token_url         = string
    api_url           = string
    allowed_domains   = optional(list(string))
    allowed_groups    = optional(list(string))
  })
  default = null
}
```

## Templates Kubernetes (Helm)

### <a name="opentelemetry-collector-helm"></a> OpenTelemetry Collector (Helm)

```yaml
# values.yaml para o chart OpenTelemetry Collector

# Configurações Multi-dimensionais INNOVABIZ
global:
  tenant: "${TENANT_ID}"
  region: "${REGION_ID}"
  environment: "${ENVIRONMENT}"

# Modo de operação (deployment, daemonset, statefulset)
mode: "deployment"

# Configuração do Collector
config:
  receivers:
    otlp:
      protocols:
        grpc:
          endpoint: 0.0.0.0:4317
        http:
          endpoint: 0.0.0.0:4318
    prometheus:
      config:
        scrape_configs:
          - job_name: 'otel-collector'
            scrape_interval: 10s
            static_configs:
              - targets: ['${POD_IP}:8888']

  processors:
    batch:
      timeout: 1s
      send_batch_size: 1024
    memory_limiter:
      check_interval: 1s
      limit_percentage: 80
      spike_limit_percentage: 25
    resourcedetection:
      detectors: [env, kubernetes]
      timeout: 2s
    k8sattributes:
      auth_type: "serviceAccount"
      passthrough: false
      filter:
        node_from_env_var: KUBE_NODE_NAME
      extract:
        metadata:
          - k8s.namespace.name
          - k8s.pod.name
          - k8s.deployment.name
          - k8s.node.name
      pod_association:
        - sources:
          - from: resource_attribute
            name: k8s.pod.ip
          - from: resource_attribute
            name: k8s.pod.name
    attributes:
      actions:
        - key: tenant.id
          value: "${TENANT_ID}"
          action: insert
        - key: region.id
          value: "${REGION_ID}"
          action: insert
        - key: environment
          value: "${ENVIRONMENT}"
          action: insert
    filter:
      metrics:
        include:
          match_type: regexp
          metric_names:
            - .*
        exclude:
          match_type: regexp
          metric_names:
            - ^internal\..*$

  exporters:
    prometheus:
      endpoint: 0.0.0.0:8889
      namespace: "${TENANT_ID}_${REGION_ID}"
    otlp:
      endpoint: "otlp-gateway.observability.svc.cluster.local:4317"
      tls:
        insecure: false
        ca_file: /etc/ssl/certs/ca-certificates.crt
    otlp/jaeger:
      endpoint: "tempo.observability.svc.cluster.local:4317"
      tls:
        insecure: false
    logging:
      verbosity: detailed
      sampling_initial: 5
      sampling_thereafter: 200

  service:
    pipelines:
      traces:
        receivers: [otlp]
        processors: [memory_limiter, k8sattributes, resourcedetection, attributes, batch]
        exporters: [otlp/jaeger, logging]
      metrics:
        receivers: [otlp, prometheus]
        processors: [memory_limiter, k8sattributes, resourcedetection, filter, attributes, batch]
        exporters: [prometheus, otlp]

# Configurações do Deployment
replicaCount: 2

# Recursos
resources:
  limits:
    cpu: 1
    memory: 2Gi
  requests:
    cpu: 200m
    memory: 400Mi

# Afinidade e Anti-afinidade
affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchExpressions:
            - key: app.kubernetes.io/name
              operator: In
              values:
                - opentelemetry-collector
        topologyKey: "kubernetes.io/hostname"

# Segurança
securityContext:
  runAsNonRoot: true
  runAsUser: 10001
  capabilities:
    drop:
      - ALL
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false

# Configurações de Service
service:
  type: ClusterIP
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "8889"
  ports:
    - name: otlp-grpc
      port: 4317
      targetPort: 4317
    - name: otlp-http
      port: 4318
      targetPort: 4318
    - name: prometheus
      port: 8889
      targetPort: 8889
    - name: metrics
      port: 8888
      targetPort: 8888

# Variáveis de ambiente
env:
  - name: TENANT_ID
    value: "${TENANT_ID}"
  - name: REGION_ID
    value: "${REGION_ID}"
  - name: ENVIRONMENT
    value: "${ENVIRONMENT}"
  - name: KUBE_NODE_NAME
    valueFrom:
      fieldRef:
        fieldPath: spec.nodeName
  - name: POD_IP
    valueFrom:
      fieldRef:
        fieldPath: status.podIP

# Labels específicas INNOVABIZ
extraLabels:
  innovabiz.com/tenant: "${TENANT_ID}"
  innovabiz.com/region: "${REGION_ID}"
  innovabiz.com/environment: "${ENVIRONMENT}"
  innovabiz.com/component: "observability"
  innovabiz.com/module: "telemetry-collector"
```

### <a name="prometheus-helm"></a> Prometheus (Helm)

```yaml
# values.yaml para o chart Prometheus

# Configurações Multi-dimensionais INNOVABIZ
global:
  tenant: "${TENANT_ID}"
  region: "${REGION_ID}"
  environment: "${ENVIRONMENT}"

# Configuração do Servidor Prometheus
server:
  image:
    repository: quay.io/prometheus/prometheus
    tag: v2.45.0
  
  # Configurações de Persistência
  persistentVolume:
    enabled: true
    size: "${PROMETHEUS_STORAGE_SIZE}"
    storageClass: "${STORAGE_CLASS}"
  
  # Configuração de Retenção
  retention: "${PROMETHEUS_RETENTION}"
  
  # Segurança
  securityContext:
    runAsNonRoot: true
    runAsUser: 65534
    fsGroup: 65534
  
  # Recursos
  resources:
    limits:
      cpu: 1
      memory: 2Gi
    requests:
      cpu: 500m
      memory: 500Mi
  
  # Configurações de Alta Disponibilidade
  replicaCount: 2
  affinity:
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        - labelSelector:
            matchExpressions:
              - key: app
                operator: In
                values:
                  - prometheus
          topologyKey: "kubernetes.io/hostname"
  
  # Configurações específicas INNOVABIZ
  extraFlags:
    - web.enable-lifecycle
    - web.enable-admin-api
    - storage.tsdb.allow-overlapping-blocks
  
  # Configurações adicionais para suporte multi-dimensional
  extraArgs:
    - --query.timeout=2m
    - --query.max-samples=100000000
  
  # Ingress
  ingress:
    enabled: true
    ingressClassName: nginx
    annotations:
      kubernetes.io/tls-acme: "true"
      nginx.ingress.kubernetes.io/ssl-redirect: "true"
      nginx.ingress.kubernetes.io/auth-type: basic
      nginx.ingress.kubernetes.io/auth-secret: prometheus-basic-auth
      nginx.ingress.kubernetes.io/auth-realm: "Authentication Required"
    hosts:
      - prometheus-${TENANT_ID}-${REGION_ID}.${INGRESS_DOMAIN}
    tls:
      - secretName: prometheus-tls
        hosts:
          - prometheus-${TENANT_ID}-${REGION_ID}.${INGRESS_DOMAIN}

# Configuração do Alertmanager
alertmanager:
  enabled: false  # Configurado separadamente

# Configuração do Exportador de Nós
nodeExporter:
  enabled: true
  tolerations:
    - effect: NoSchedule
      operator: Exists
  securityContext:
    runAsNonRoot: true
    runAsUser: 65534

# Configuração de Service Discovery Kubernetes
kubeStateMetrics:
  enabled: true

# Configurações globais de scrape
serverFiles:
  prometheus.yml:
    global:
      scrape_interval: 30s
      scrape_timeout: 10s
      evaluation_interval: 30s
    rule_files:
      - /etc/prometheus/rules/*.rules
    scrape_configs:
      # Configurações para Kubernetes
      - job_name: 'kubernetes-apiservers'
        kubernetes_sd_configs:
          - role: endpoints
        scheme: https
        tls_config:
          ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
        relabel_configs:
          - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
            action: keep
            regex: default;kubernetes;https
          - target_label: tenant_id
            replacement: ${TENANT_ID}
          - target_label: region_id
            replacement: ${REGION_ID}
          - target_label: environment
            replacement: ${ENVIRONMENT}

      # Configurações para Nodes
      - job_name: 'kubernetes-nodes'
        kubernetes_sd_configs:
          - role: node
        relabel_configs:
          - target_label: tenant_id
            replacement: ${TENANT_ID}
          - target_label: region_id
            replacement: ${REGION_ID}
          - target_label: environment
            replacement: ${ENVIRONMENT}
          - action: labelmap
            regex: __meta_kubernetes_node_label_(.+)

      # Configurações para Pods
      - job_name: 'kubernetes-pods'
        kubernetes_sd_configs:
          - role: pod
        relabel_configs:
          - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
            action: keep
            regex: true
          - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
            action: replace
            target_label: __metrics_path__
            regex: (.+)
          - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
            action: replace
            regex: (.+):(?:\d+);(\d+)
            replacement: ${1}:${2}
            target_label: __address__
          - target_label: tenant_id
            replacement: ${TENANT_ID}
          - target_label: region_id
            replacement: ${REGION_ID}
          - target_label: environment
            replacement: ${ENVIRONMENT}
          - action: labelmap
            regex: __meta_kubernetes_pod_label_(.+)
          - source_labels: [__meta_kubernetes_namespace]
            action: replace
            target_label: kubernetes_namespace
          - source_labels: [__meta_kubernetes_pod_name]
            action: replace
            target_label: kubernetes_pod_name
          - source_labels: [__meta_kubernetes_pod_annotation_innovabiz_com_tenant]
            action: replace
            target_label: tenant_id
          - source_labels: [__meta_kubernetes_pod_annotation_innovabiz_com_region]
            action: replace
            target_label: region_id

# Labels específicas INNOVABIZ
extraLabels:
  innovabiz.com/tenant: "${TENANT_ID}"
  innovabiz.com/region: "${REGION_ID}"
  innovabiz.com/environment: "${ENVIRONMENT}"
  innovabiz.com/component: "observability"
  innovabiz.com/module: "metrics"
```

## <a name="checklist-de-validação"></a> Checklist de Validação

Ao implementar a Infraestrutura como Código para observabilidade, utilize este checklist para validar a conformidade com os padrões INNOVABIZ:

### Validação de Multi-dimensionalidade

- [ ] **Isolamento de Tenant**
  - [ ] Cada tenant possui namespace ou cluster isolado
  - [ ] Labels de tenant aplicados em todos os recursos
  - [ ] Políticas RBAC configuradas para isolamento de tenant

- [ ] **Contexto Regional**
  - [ ] Configurações específicas por região implementadas
  - [ ] Labels de região aplicados em todos os recursos
  - [ ] Recursos dimensionados de acordo com necessidades regionais

- [ ] **Separação de Ambientes**
  - [ ] Ambientes (production, staging, development) claramente separados
  - [ ] Configurações específicas por ambiente implementadas
  - [ ] Políticas de retenção de dados diferenciadas por ambiente

### Validação de Segurança

- [ ] **TLS/mTLS**
  - [ ] Comunicação entre componentes protegida por TLS
  - [ ] Autenticação mTLS configurada onde necessário
  - [ ] Certificados gerenciados adequadamente (validade, renovação)

- [ ] **RBAC**
  - [ ] Serviços usando contas de serviço com privilégios mínimos
  - [ ] Políticas de RBAC aplicadas para acesso a recursos
  - [ ] Permissões de namespace adequadamente configuradas

- [ ] **Segredos**
  - [ ] Credenciais armazenadas como Kubernetes Secrets
  - [ ] Secrets não expostos em logs ou configurações
  - [ ] Rotação automática de credenciais configurada

### Validação de Recursos e Escalabilidade

- [ ] **Recursos de Computação**
  - [ ] Limites e requisições de CPU/memória definidos para todos os pods
  - [ ] HPA (Horizontal Pod Autoscaler) configurado para componentes críticos
  - [ ] Políticas de afinidade e anti-afinidade implementadas

- [ ] **Armazenamento**
  - [ ] PVCs configurados com tamanho adequado
  - [ ] StorageClass apropriada selecionada
  - [ ] Políticas de retenção de dados configuradas

- [ ] **Rede**
  - [ ] Network Policies aplicadas para isolamento
  - [ ] Ingress configurado com autenticação
  - [ ] Roteamento entre componentes otimizado

### Validação de Observabilidade

- [ ] **Coleta de Métricas**
  - [ ] Scraping configurado para todos os componentes
  - [ ] Labels padrão INNOVABIZ aplicados a todas as métricas
  - [ ] Métricas de sistema e negócio separadas

- [ ] **Coleta de Logs**
  - [ ] Logs estruturados em formato JSON
  - [ ] Labels de contexto incluídos nos logs
  - [ ] Níveis de log adequados por ambiente

- [ ] **Coleta de Traces**
  - [ ] Propagação de contexto configurada
  - [ ] Sampling rate adequado por ambiente
  - [ ] Integração com spans do OpenTelemetry

- [ ] **Alertas**
  - [ ] Regras de alerta com thresholds adequados
  - [ ] Roteamento de alertas configurado
  - [ ] Silenciamentos e inibições para manutenção

## <a name="melhores-práticas"></a> Melhores Práticas

### Padrões de Implementação

1. **Estrutura GitOps**
   - Mantenha toda a configuração de IaC em repositórios Git
   - Utilize CD automatizado para implantação baseada em alterações de Git
   - Implemente validação de configuração em pipelines CI/CD

2. **Estratégia de Versionamento**
   - Versione todos os módulos Terraform e charts Helm
   - Utilize tags imutáveis para imagens de contêiner
   - Documente mudanças em CHANGELOG.md

3. **Modularização**
   - Divida configurações complexas em módulos reutilizáveis
   - Separe preocupações (ex: coleta vs armazenamento vs visualização)
   - Crie abstrações para padrões comuns

### Padrões Multi-dimensionais INNOVABIZ

1. **Nomenclatura**
   - Padronize prefixos/sufixos para incluir tenant e região
   - Use namespace hierárquico para métricas: `tenant_regiao_aplicacao_metrica`
   - Aplique rotulagem consistente em todos os recursos

2. **Isolamento**
   - Namespaces separados por tenant/função
   - Visualizações filtradas automaticamente por contexto
   - Controle de acesso granular baseado em dimensões

3. **Configuração Dinâmica**
   - Utilize ConfigMaps para definições específicas de contexto
   - Implemente substituição de variáveis na implantação
   - Crie operadores customizados para gerenciamento de CRDs específicos INNOVABIZ

### Segurança e Compliance

1. **Gestão de Credenciais**
   - Utilize sistemas de gestão de segredos (Vault, AWS Secrets Manager)
   - Automatize rotação de credenciais
   - Implemente revogação de tokens de emergência

2. **Auditoria**
   - Habilite logs de auditoria em todos os componentes
   - Configure retenção de logs de auditoria conforme requisitos regulatórios
   - Integre com sistemas de alerta para eventos de segurança

3. **Conformidade**
   - Implemente políticas de mascaramento/redação de dados sensíveis
   - Configure retenção de dados conforme LGPD/GDPR
   - Documente práticas de segurança e compliance

### Operação e Manutenção

1. **Atualizações**
   - Defina janelas de manutenção programadas
   - Implemente estratégias de atualização sem tempo de inatividade
   - Teste atualizações em ambientes não-produtivos

2. **Backups**
   - Configure backups para dados de longo prazo
   - Teste restaurações periodicamente
   - Documente procedimentos de recuperação

3. **Monitoramento da Infraestrutura**
   - Monitore a própria infraestrutura de observabilidade
   - Configure alertas para problemas na stack de observabilidade
   - Defina runbooks para recuperação da infraestrutura de observabilidade

## Recursos Adicionais

- [Portal de Documentação INNOVABIZ](https://docs.innovabiz.com/observability)
- [Repositório de Templates IaC](https://github.com/innovabiz/observability-templates)
- [Biblioteca de Módulos Terraform INNOVABIZ](https://github.com/innovabiz/terraform-modules)
- [Charts Helm Customizados INNOVABIZ](https://github.com/innovabiz/helm-charts)
- [Políticas de Segurança e Compliance](https://docs.innovabiz.com/security/policies)