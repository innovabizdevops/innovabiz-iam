# INNOVABIZ-OBS-RULES01 - Regras de Alerta Padronizadas

## 🚨 Visão Geral

**Módulo:** INNOVABIZ Observability Framework  
**Componente:** Alert Rules  
**Versão:** 1.0.0  
**Equipe Responsável:** Observability Team  
**Contatos Primários:** observability@innovabiz.com, #observability-support  
**Repositório:** `CoreModules/IAM/docs/observability`  
**Status:** 🟢 Aprovado para Produção  

## 📋 Introdução

Este documento define o padrão INNOVABIZ para regras de alerta em todos os ambientes, serviços e produtos da plataforma. Todas as regras de alerta devem seguir estas diretrizes para garantir consistência, acionabilidade e eficácia no programa de observabilidade da organização.

## 🔍 Princípios Fundamentais

1. **Acionabilidade:** Cada alerta deve exigir intervenção humana imediata ou programada
2. **Precisão:** Minimizar falsos positivos/negativos com thresholds apropriados
3. **Contexto Completo:** Fornecer informações suficientes para diagnóstico sem consultas adicionais
4. **Multi-Dimensional:** Suportar isolamento por tenant, região, ambiente e componente
5. **Conformidade:** Aderência aos requisitos regulatórios (PCI DSS, GDPR, ISO 27001)
6. **Priorização Clara:** Severidade e urgência definidas objetivamente
7. **Escalonamento Adequado:** Roteamento baseado em severidade, horário e SLA
8. **Documentação Integrada:** Links diretos para runbooks e procedimentos

## 📊 Taxonomia e Classificação

### Severidades

| Nível | Código | Descrição | Tempo de Resposta | Exemplo |
|-------|--------|-----------|-------------------|---------|
| **Critical** | P1 | Impacto severo em produção, perda de serviço | 15 min | Serviço completamente indisponível |
| **High** | P2 | Impacto significativo, funcionalidade principal degradada | 30 min | API com alta taxa de erros |
| **Medium** | P3 | Impacto moderado, funcionalidade secundária afetada | 2 horas | Lentidão em operações não críticas |
| **Low** | P4 | Impacto menor, problemas potenciais futuros | 8 horas | Disk space acima de 75% |
| **Info** | P5 | Informacional, sem impacto direto | 24 horas | Alterações de configuração |

### Categorias

- **Availability:** Disponibilidade de serviços/componentes
- **Latency:** Tempo de resposta e performance
- **Saturation:** Uso de recursos (CPU, memória, disco, rede)
- **Errors:** Taxas de erro e falhas
- **Traffic:** Volume de requisições e transações
- **Configuration:** Problemas de configuração
- **Security:** Alertas de segurança e compliance
- **Business:** Métricas de negócio (transações, conversões)
- **Dependencies:** Problemas com dependências externas

### Labels Obrigatórios

| Label | Descrição | Exemplo |
|-------|-----------|---------|
| `tenant_id` | Identificador do tenant | `fintech_xyz` |
| `region_id` | Região de implantação | `br`, `us`, `eu`, `ao` |
| `environment` | Ambiente de execução | `production`, `staging`, `development` |
| `component` | Componente afetado | `payment_gateway`, `iam_service` |
| `module` | Módulo específico | `authentication`, `transaction_processing` |
| `severity` | Nível de severidade | `critical`, `high`, `medium`, `low`, `info` |
| `category` | Categoria do alerta | `availability`, `latency`, `saturation` |
| `priority` | Código de prioridade | `P1`, `P2`, `P3`, `P4`, `P5` |
| `sla` | Tempo de resposta esperado | `15m`, `30m`, `2h`, `8h`, `24h` |
| `team` | Equipe responsável | `platform`, `payments`, `security`, `infrastructure` |

## 🔧 Estrutura de Regras de Alerta Prometheus

### Formato Padrão

```yaml
groups:
- name: innovabiz_<component>_alerts
  rules:
  - alert: INNOVABIZ_<Componente>_<Condição>
    expr: <expressão_prometheus>
    for: <duração>
    labels:
      tenant_id: '{{ $labels.tenant_id }}'
      region_id: '{{ $labels.region_id }}'
      environment: '{{ $labels.environment }}'
      component: '<componente>'
      module: '<módulo>'
      severity: '<severidade>'
      category: '<categoria>'
      priority: '<código_prioridade>'
      sla: '<tempo_resposta>'
      team: '<equipe_responsável>'
    annotations:
      summary: "<resumo conciso do problema>"
      description: "<descrição detalhada com valores e contexto>"
      impact: "<impacto nos usuários/negócio>"
      action: "<ação recomendada>"
      dashboard_url: "https://grafana.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/d/<dashboard_id>"
      runbook_url: "https://docs.innovabiz.com/observability/runbooks/<runbook_id>"
      alert_details_url: "https://alertmanager.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/#/alerts?filter=%7Balertname%3D%22INNOVABIZ_<Componente>_<Condição>%22%7D"
```

### Convenção de Nomenclatura

- **Nome do Alerta:** `INNOVABIZ_<Componente>_<Condição>`
  - Componente: Serviço ou sistema (ex: IAM, PAYMENT_GATEWAY)
  - Condição: O que está sendo violado (ex: HIGH_ERROR_RATE, CPU_SATURATION)

- **Grupos:** `innovabiz_<component>_alerts`
  - Agrupar alertas relacionados ao mesmo componente

## 📚 Exemplos por Domínio

### 1. Infraestrutura

```yaml
groups:
- name: innovabiz_infrastructure_alerts
  rules:
  - alert: INNOVABIZ_Node_HighCpuUsage
    expr: (1 - avg by(tenant_id, region_id, environment, instance) (rate(node_cpu_seconds_total{mode="idle"}[5m]))) * 100 > 90
    for: 10m
    labels:
      tenant_id: '{{ $labels.tenant_id }}'
      region_id: '{{ $labels.region_id }}'
      environment: '{{ $labels.environment }}'
      component: 'infrastructure'
      module: 'compute'
      severity: 'high'
      category: 'saturation'
      priority: 'P2'
      sla: '30m'
      team: 'infrastructure'
    annotations:
      summary: "Alta utilização de CPU em {{ $labels.instance }}"
      description: "Node {{ $labels.instance }} tem utilização de CPU {{ $value | humanizePercentage }} por mais de 10 minutos."
      impact: "Performance degradada para todos os serviços no node"
      action: "Verificar processos consumindo CPU, escalar horizontalmente se necessário"
      dashboard_url: "https://grafana.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/d/infrastructure-nodes"
      runbook_url: "https://docs.innovabiz.com/observability/runbooks/infrastructure/high-cpu-usage"
  - alert: INNOVABIZ_Node_DiskSpaceRunningLow
    expr: (node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"} * 100) < 15
    for: 5m
    labels:
      tenant_id: '{{ $labels.tenant_id }}'
      region_id: '{{ $labels.region_id }}'
      environment: '{{ $labels.environment }}'
      component: 'infrastructure'
      module: 'storage'
      severity: 'medium'
      category: 'saturation'
      priority: 'P3'
      sla: '2h'
      team: 'infrastructure'
    annotations:
      summary: "Espaço em disco baixo em {{ $labels.instance }}"
      description: "Node {{ $labels.instance }} tem apenas {{ $value | humanizePercentage }} de espaço em disco disponível."
      impact: "Risco de falha em operações de escrita e logs"
      action: "Limpar arquivos temporários/logs ou aumentar capacidade de armazenamento"
      dashboard_url: "https://grafana.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/d/infrastructure-storage"
      runbook_url: "https://docs.innovabiz.com/observability/runbooks/infrastructure/disk-space-low"
      
  - alert: INNOVABIZ_Network_HighLatency
    expr: avg_over_time(probe_http_duration_seconds{job="blackbox"}[5m]) > 1
    for: 5m
    labels:
      tenant_id: '{{ $labels.tenant_id }}'
      region_id: '{{ $labels.region_id }}'
      environment: '{{ $labels.environment }}'
      component: 'infrastructure'
      module: 'network'
      severity: 'high'
      category: 'latency'
      priority: 'P2'
      sla: '30m'
      team: 'infrastructure'
    annotations:
      summary: "Alta latência de rede para {{ $labels.target }}"
      description: "Latência média de {{ $value | humanizeDuration }} para {{ $labels.target }} nos últimos 5 minutos."
      impact: "Degradação de performance para usuários"
      action: "Verificar conectividade, DNS e firewalls"
      dashboard_url: "https://grafana.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/d/network-overview"
      runbook_url: "https://docs.innovabiz.com/observability/runbooks/infrastructure/network-latency"

### 2. IAM Service

```yaml
groups:
- name: innovabiz_iam_service_alerts
  rules:
  - alert: INNOVABIZ_IAM_HighErrorRate
    expr: sum(rate(iam_http_requests_total{status=~"5.."}[5m])) by (tenant_id, region_id, environment, service) / sum(rate(iam_http_requests_total[5m])) by (tenant_id, region_id, environment, service) > 0.05
    for: 5m
    labels:
      tenant_id: '{{ $labels.tenant_id }}'
      region_id: '{{ $labels.region_id }}'
      environment: '{{ $labels.environment }}'
      component: 'iam'
      module: 'authentication'
      severity: 'critical'
      category: 'errors'
      priority: 'P1'
      sla: '15m'
      team: 'identity'
    annotations:
      summary: "Alta taxa de erros no serviço IAM"
      description: "IAM service em {{ $labels.tenant_id }}/{{ $labels.region_id }}/{{ $labels.environment }} tem taxa de erros de {{ $value | humanizePercentage }} (>5%)."
      impact: "Falhas de autenticação/autorização afetando todos os serviços"
      action: "Verificar logs, escalar para equipe de identidade imediatamente"
      dashboard_url: "https://grafana.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/d/iam-service-overview"
      runbook_url: "https://docs.innovabiz.com/observability/runbooks/iam/high-error-rate"
      
  - alert: INNOVABIZ_IAM_HighAuthLatency
    expr: histogram_quantile(0.95, sum(rate(iam_authentication_duration_seconds_bucket[5m])) by (le, tenant_id, region_id, environment)) > 1
    for: 5m
    labels:
      tenant_id: '{{ $labels.tenant_id }}'
      region_id: '{{ $labels.region_id }}'
      environment: '{{ $labels.environment }}'
      component: 'iam'
      module: 'authentication'
      severity: 'high'
      category: 'latency'
      priority: 'P2'
      sla: '30m'
      team: 'identity'
    annotations:
      summary: "Alta latência de autenticação no IAM"
      description: "95º percentil de latência de autenticação no IAM para {{ $labels.tenant_id }}/{{ $labels.region_id }}/{{ $labels.environment }} é {{ $value | humanizeDuration }} (>1s)."
      impact: "Login e autorizações lentas afetando UX"
      action: "Verificar connection pools, cache e dependências externas"
      dashboard_url: "https://grafana.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/d/iam-service-performance"
      runbook_url: "https://docs.innovabiz.com/observability/runbooks/iam/auth-latency"

  - alert: INNOVABIZ_IAM_HighTokenRejectionRate
    expr: sum(rate(iam_token_validations_rejected_total[5m])) by (tenant_id, region_id, environment) / sum(rate(iam_token_validations_total[5m])) by (tenant_id, region_id, environment) > 0.1
    for: 5m
    labels:
      tenant_id: '{{ $labels.tenant_id }}'
      region_id: '{{ $labels.region_id }}'
      environment: '{{ $labels.environment }}'
      component: 'iam'
      module: 'authorization'
      severity: 'high'
      category: 'security'
      priority: 'P2'
      sla: '30m'
      team: 'identity'
    annotations:
      summary: "Alta taxa de rejeição de tokens no IAM"
      description: "IAM em {{ $labels.tenant_id }}/{{ $labels.region_id }}/{{ $labels.environment }} está rejeitando {{ $value | humanizePercentage }} dos tokens (>10%)."
      impact: "Possível tentativa de acesso indevido ou expiração prematura de tokens"
      action: "Verificar relógios sincronizados e possíveis ataques"
      dashboard_url: "https://grafana.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/d/iam-service-security"
      runbook_url: "https://docs.innovabiz.com/observability/runbooks/iam/token-rejection"

### 3. Payment Gateway

```yaml
groups:
- name: innovabiz_payment_gateway_alerts
  rules:
  - alert: INNOVABIZ_PaymentGateway_HighTransactionFailureRate
    expr: sum(rate(payment_transactions_total{status="failed"}[5m])) by (tenant_id, region_id, environment, payment_method) / sum(rate(payment_transactions_total[5m])) by (tenant_id, region_id, environment, payment_method) > 0.05
    for: 5m
    labels:
      tenant_id: '{{ $labels.tenant_id }}'
      region_id: '{{ $labels.region_id }}'
      environment: '{{ $labels.environment }}'
      component: 'payment_gateway'
      module: 'transaction_processing'
      severity: 'critical'
      category: 'business'
      priority: 'P1'
      sla: '15m'
      team: 'payments'
    annotations:
      summary: "Alta taxa de falha em transações de pagamento"
      description: "{{ $labels.payment_method }} em {{ $labels.tenant_id }}/{{ $labels.region_id }}/{{ $labels.environment }} tem taxa de falha de {{ $value | humanizePercentage }} (>5%)."
      impact: "Perda de receita e deterioração da experiência do cliente"
      action: "Verificar logs, dependências de processadores e infraestrutura"
      dashboard_url: "https://grafana.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/d/payment-gateway-transactions"
      runbook_url: "https://docs.innovabiz.com/observability/runbooks/payments/transaction-failures"  - alert: INNOVABIZ_PaymentGateway_HighProcessingLatency
    expr: histogram_quantile(0.95, sum(rate(payment_transaction_duration_seconds_bucket[5m])) by (le, tenant_id, region_id, environment, payment_method)) > 3
    for: 5m
    labels:
      tenant_id: '{{ $labels.tenant_id }}'
      region_id: '{{ $labels.region_id }}'
      environment: '{{ $labels.environment }}'
      component: 'payment_gateway'
      module: 'transaction_processing'
      severity: 'high'
      category: 'latency'
      priority: 'P2'
      sla: '30m'
      team: 'payments'
    annotations:
      summary: "Alta latência de processamento de pagamentos"
      description: "95º percentil da latência de processamento para {{ $labels.payment_method }} em {{ $labels.tenant_id }}/{{ $labels.region_id }}/{{ $labels.environment }} é {{ $value | humanizeDuration }} (>3s)."
      impact: "Deterioração da experiência do cliente, possíveis timeouts"
      action: "Verificar processadores de pagamento, conexões de rede e configurações"
      dashboard_url: "https://grafana.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/d/payment-gateway-performance"
      runbook_url: "https://docs.innovabiz.com/observability/runbooks/payments/processing-latency"

  - alert: INNOVABIZ_PaymentGateway_HighFraudDetectionRate
    expr: sum(rate(payment_fraud_detection_total{result="suspected"}[1h])) by (tenant_id, region_id, environment) / sum(rate(payment_transactions_total[1h])) by (tenant_id, region_id, environment) > 0.1
    for: 15m
    labels:
      tenant_id: '{{ $labels.tenant_id }}'
      region_id: '{{ $labels.region_id }}'
      environment: '{{ $labels.environment }}'
      component: 'payment_gateway'
      module: 'fraud_detection'
      severity: 'medium'
      category: 'security'
      priority: 'P3'
      sla: '2h'
      team: 'risk'
    annotations:
      summary: "Taxa elevada de detecção de fraude"
      description: "{{ $value | humanizePercentage }} das transações em {{ $labels.tenant_id }}/{{ $labels.region_id }}/{{ $labels.environment }} estão sendo sinalizadas como potenciais fraudes (>10%)."
      impact: "Possível ataque em andamento ou falso positivo afetando conversões"
      action: "Analisar padrões de fraude, revisar regras de detecção"
      dashboard_url: "https://grafana.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/d/payment-gateway-security"
      runbook_url: "https://docs.innovabiz.com/observability/runbooks/payments/fraud-detection"

### 4. KrakenD API Gateway

```yaml
groups:
- name: innovabiz_krakend_gateway_alerts
  rules:
  - alert: INNOVABIZ_KrakenD_HighErrorRate
    expr: sum(rate(krakend_router_response_code_count{status=~"5.."}[5m])) by (tenant_id, region_id, environment, endpoint) / sum(rate(krakend_router_response_code_count[5m])) by (tenant_id, region_id, environment, endpoint) > 0.05
    for: 5m
    labels:
      tenant_id: '{{ $labels.tenant_id }}'
      region_id: '{{ $labels.region_id }}'
      environment: '{{ $labels.environment }}'
      component: 'api_gateway'
      module: 'krakend'
      severity: 'critical'
      category: 'errors'
      priority: 'P1'
      sla: '15m'
      team: 'platform'
    annotations:
      summary: "Alta taxa de erros no endpoint {{ $labels.endpoint }}"
      description: "API Gateway para {{ $labels.endpoint }} em {{ $labels.tenant_id }}/{{ $labels.region_id }}/{{ $labels.environment }} tem taxa de erros de {{ $value | humanizePercentage }} (>5%)."
      impact: "Falhas de API afetando múltiplos clientes e serviços"
      action: "Verificar backends, logs e configurações do endpoint"
      dashboard_url: "https://grafana.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/d/krakend-gateway-overview"
      runbook_url: "https://docs.innovabiz.com/observability/runbooks/api-gateway/high-error-rate"
      
  - alert: INNOVABIZ_KrakenD_CircuitBreakerOpen
    expr: sum(krakend_service_errors_count{err="circuit-open"}) by (tenant_id, region_id, environment, backend) > 0
    for: 1m
    labels:
      tenant_id: '{{ $labels.tenant_id }}'
      region_id: '{{ $labels.region_id }}'
      environment: '{{ $labels.environment }}'
      component: 'api_gateway'
      module: 'krakend'
      severity: 'high'
      category: 'availability'
      priority: 'P2'
      sla: '30m'
      team: 'platform'
    annotations:
      summary: "Circuit breaker aberto para {{ $labels.backend }}"
      description: "API Gateway detectou falhas e abriu circuit breaker para {{ $labels.backend }} em {{ $labels.tenant_id }}/{{ $labels.region_id }}/{{ $labels.environment }}."
      impact: "Requisições para este backend estão sendo rejeitadas"
      action: "Verificar saúde do backend e corrigir problemas subjacentes"
      dashboard_url: "https://grafana.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/d/krakend-circuit-breakers"
      runbook_url: "https://docs.innovabiz.com/observability/runbooks/api-gateway/circuit-breaker"

### 5. Banco de Dados

```yaml
groups:
- name: innovabiz_database_alerts
  rules:
  - alert: INNOVABIZ_PostgreSQL_HighConnections
    expr: sum(pg_stat_activity_count) by (tenant_id, region_id, environment, datname) > (pg_settings_max_connections{setting="max_connections"} * 0.9)
    for: 5m
    labels:
      tenant_id: '{{ $labels.tenant_id }}'
      region_id: '{{ $labels.region_id }}'
      environment: '{{ $labels.environment }}'
      component: 'database'
      module: 'postgresql'
      severity: 'high'
      category: 'saturation'
      priority: 'P2'
      sla: '30m'
      team: 'database'
    annotations:
      summary: "Alto número de conexões PostgreSQL"
      description: "Database {{ $labels.datname }} em {{ $labels.tenant_id }}/{{ $labels.region_id }}/{{ $labels.environment }} está com {{ $value }} conexões ativas (>90% do máximo)."
      impact: "Risco de esgotamento de conexões e falhas em aplicações"
      action: "Verificar connection pools e possíveis vazamentos nas aplicações"
      dashboard_url: "https://grafana.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/d/postgresql-overview"
      runbook_url: "https://docs.innovabiz.com/observability/runbooks/database/high-connections"  - alert: INNOVABIZ_PostgreSQL_SlowQueries
    expr: rate(pg_stat_activity_max_tx_duration{datname!~"template.*|postgres"}[5m]) > 30
    for: 5m
    labels:
      tenant_id: '{{ $labels.tenant_id }}'
      region_id: '{{ $labels.region_id }}'
      environment: '{{ $labels.environment }}'
      component: 'database'
      module: 'postgresql'
      severity: 'medium'
      category: 'performance'
      priority: 'P3'
      sla: '2h'
      team: 'database'
    annotations:
      summary: "Consultas lentas detectadas no PostgreSQL"
      description: "Database {{ $labels.datname }} em {{ $labels.tenant_id }}/{{ $labels.region_id }}/{{ $labels.environment }} tem consultas com duração média de {{ $value }}s (>30s)."
      impact: "Degradação de desempenho e possível bloqueio de recursos"
      action: "Verificar explain plans, adicionar índices ou otimizar consultas"
      dashboard_url: "https://grafana.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/d/postgresql-query-performance"
      runbook_url: "https://docs.innovabiz.com/observability/runbooks/database/slow-queries"

  - alert: INNOVABIZ_Redis_HighMemoryUsage
    expr: redis_memory_used_bytes / redis_memory_max_bytes * 100 > 80
    for: 5m
    labels:
      tenant_id: '{{ $labels.tenant_id }}'
      region_id: '{{ $labels.region_id }}'
      environment: '{{ $labels.environment }}'
      component: 'database'
      module: 'redis'
      severity: 'high'
      category: 'saturation'
      priority: 'P2'
      sla: '30m'
      team: 'database'
    annotations:
      summary: "Alto uso de memória no Redis"
      description: "Instância Redis em {{ $labels.tenant_id }}/{{ $labels.region_id }}/{{ $labels.environment }} está utilizando {{ $value | humanizePercentage }} de memória (>80%)."
      impact: "Risco de evicções de chaves e degradação de performance"
      action: "Avaliar políticas de expiração, verificar vazamentos de memória, considerar escalamento"
      dashboard_url: "https://grafana.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/d/redis-overview"
      runbook_url: "https://docs.innovabiz.com/observability/runbooks/database/redis-memory"

### 6. Infraestrutura Kubernetes

```yaml
groups:
- name: innovabiz_kubernetes_alerts
  rules:
  - alert: INNOVABIZ_Kubernetes_PodCrashLooping
    expr: max_over_time(kube_pod_container_status_restarts_total{namespace=~"innovabiz-.*"}[1h]) - min_over_time(kube_pod_container_status_restarts_total{namespace=~"innovabiz-.*"}[1h]) > 3
    for: 5m
    labels:
      tenant_id: '{{ $labels.tenant_id }}'
      region_id: '{{ $labels.region_id }}'
      environment: '{{ $labels.environment }}'
      component: 'infrastructure'
      module: 'kubernetes'
      severity: 'high'
      category: 'availability'
      priority: 'P2'
      sla: '30m'
      team: 'platform'
    annotations:
      summary: "Pod em crash loop: {{ $labels.namespace }}/{{ $labels.pod }}"
      description: "O container {{ $labels.container }} no pod {{ $labels.pod }} no namespace {{ $labels.namespace }} reiniciou {{ $value }} vezes na última hora."
      impact: "Serviço instável ou indisponível"
      action: "Verificar logs do pod, eventos Kubernetes e resource limits"
      dashboard_url: "https://grafana.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/d/kubernetes-pod-overview"
      runbook_url: "https://docs.innovabiz.com/observability/runbooks/kubernetes/pod-crash-loop"
  
  - alert: INNOVABIZ_Kubernetes_NodeHighCPU
    expr: instance:node_cpu_utilisation:rate5m > 0.9
    for: 5m
    labels:
      tenant_id: '{{ $labels.tenant_id }}'
      region_id: '{{ $labels.region_id }}'
      environment: '{{ $labels.environment }}'
      component: 'infrastructure' 
      module: 'kubernetes'
      severity: 'high'
      category: 'saturation'
      priority: 'P2'
      sla: '30m'
      team: 'platform'
    annotations:
      summary: "Alta utilização de CPU no node {{ $labels.instance }}"
      description: "Node Kubernetes {{ $labels.instance }} em {{ $labels.tenant_id }}/{{ $labels.region_id }}/{{ $labels.environment }} está com {{ $value | humanizePercentage }} de utilização de CPU (>90%)."
      impact: "Performance degradada, possíveis throttles em containers"
      action: "Verificar cargas de trabalho, escalar horizontalmente ou mover pods"
      dashboard_url: "https://grafana.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/d/kubernetes-node-overview"
      runbook_url: "https://docs.innovabiz.com/observability/runbooks/kubernetes/node-high-cpu"

  - alert: INNOVABIZ_Kubernetes_PersistentVolumeLowSpace
    expr: kubelet_volume_stats_available_bytes / kubelet_volume_stats_capacity_bytes * 100 < 15
    for: 15m
    labels:
      tenant_id: '{{ $labels.tenant_id }}'
      region_id: '{{ $labels.region_id }}'
      environment: '{{ $labels.environment }}'
      component: 'infrastructure'
      module: 'kubernetes'
      severity: 'high'
      category: 'saturation'
      priority: 'P2'
      sla: '30m'
      team: 'platform'
    annotations:
      summary: "Espaço baixo em volume persistente"
      description: "Volume persistente {{ $labels.persistentvolumeclaim }} em {{ $labels.namespace }} tem apenas {{ $value | humanizePercentage }} de espaço disponível (<15%)."
      impact: "Possíveis falhas de gravação nas aplicações"
      action: "Limpar dados antigos, aumentar volume ou provisionar volume adicional"
      dashboard_url: "https://grafana.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/d/kubernetes-storage-overview"
      runbook_url: "https://docs.innovabiz.com/observability/runbooks/kubernetes/pv-low-space"

### 7. Observabilidade e Segurança

```yaml
groups:
- name: innovabiz_observability_security_alerts
  rules:
  - alert: INNOVABIZ_PrometheusHighCardinality
    expr: prometheus_tsdb_head_series > 10000000
    for: 15m
    labels:
      tenant_id: '{{ $labels.tenant_id }}'
      region_id: '{{ $labels.region_id }}'
      environment: '{{ $labels.environment }}'
      component: 'observability'
      module: 'prometheus'
      severity: 'high'
      category: 'saturation'
      priority: 'P2'
      sla: '1h'
      team: 'platform'
    annotations:
      summary: "Alta cardinalidade no Prometheus"
      description: "Prometheus em {{ $labels.tenant_id }}/{{ $labels.region_id }}/{{ $labels.environment }} tem {{ $value }} séries ativas (>10M)."
      impact: "Maior consumo de memória, possíveis falhas no serviço de monitoramento"
      action: "Revisar rótulos de alta cardinalidade, ajustar políticas de retenção, considerar escalamento"
      dashboard_url: "https://grafana.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/d/prometheus-overview"
      runbook_url: "https://docs.innovabiz.com/observability/runbooks/prometheus/high-cardinality"  - alert: INNOVABIZ_AlertManagerHighFailureRate
    expr: sum(rate(alertmanager_notifications_failed_total[5m])) by (tenant_id, region_id, environment, integration) / sum(rate(alertmanager_notifications_total[5m])) by (tenant_id, region_id, environment, integration) > 0.1
    for: 5m
    labels:
      tenant_id: '{{ $labels.tenant_id }}'
      region_id: '{{ $labels.region_id }}'
      environment: '{{ $labels.environment }}'
      component: 'observability'
      module: 'alertmanager'
      severity: 'critical'
      category: 'availability'
      priority: 'P1'
      sla: '15m'
      team: 'platform'
    annotations:
      summary: "Falhas na entrega de notificações para {{ $labels.integration }}"
      description: "AlertManager em {{ $labels.tenant_id }}/{{ $labels.region_id }}/{{ $labels.environment }} tem {{ $value | humanizePercentage }} de falhas ao enviar alertas para {{ $labels.integration }}."
      impact: "Incidentes podem não ser notificados adequadamente"
      action: "Verificar conectividade, credenciais e quotas do receptor de notificações"
      dashboard_url: "https://grafana.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/d/alertmanager-overview"
      runbook_url: "https://docs.innovabiz.com/observability/runbooks/alertmanager/notification-failures"

  - alert: INNOVABIZ_SecurityUnauthorizedAPIAccess
    expr: sum(rate(api_gateway_unauthorized_requests_total[5m])) by (tenant_id, region_id, environment, endpoint) > 10
    for: 15m
    labels:
      tenant_id: '{{ $labels.tenant_id }}'
      region_id: '{{ $labels.region_id }}'
      environment: '{{ $labels.environment }}'
      component: 'security'
      module: 'api_gateway'
      severity: 'high'
      category: 'security'
      priority: 'P2'
      sla: '30m'
      team: 'security'
    annotations:
      summary: "Tentativas de acesso não autorizado à API"
      description: "{{ $value }} tentativas de acesso não autorizado no endpoint {{ $labels.endpoint }} em {{ $labels.tenant_id }}/{{ $labels.region_id }}/{{ $labels.environment }} nos últimos 5 minutos."
      impact: "Possível tentativa de invasão ou problemas de configuração do cliente"
      action: "Revisar logs de acesso, verificar padrões de IP e comportamento"
      dashboard_url: "https://grafana.{{ $labels.tenant_id }}.{{ $labels.region_id }}.innovabiz.io/d/security-overview"
      runbook_url: "https://docs.innovabiz.com/observability/runbooks/security/unauthorized-access"

## IV. Implementação e Governança

### 1. Diretrizes de Implementação

A implementação de regras de alerta na plataforma INNOVABIZ deve seguir estas diretrizes:

1. **Consistência Multi-Dimensional**
   - Todas as regras de alerta DEVEM incluir rótulos de contexto multi-dimensional (`tenant_id`, `region_id`, `environment`)
   - Os alertas devem propagar corretamente esses rótulos através de expressões Prometheus
   - A nomenclatura deve seguir o padrão `INNOVABIZ_Componente_AlgoDescritivo`

2. **Granularidade Adequada**
   - Evite alertas muito genéricos que podem gerar ruído
   - Evite alertas excessivamente específicos que fragmentam a visibilidade
   - Calibre os limiares (`for` e expressões) para reduzir falsos positivos
   - Configure silenciamentos em janelas de manutenção via API AlertManager

3. **Metadados Informativos**
   - Inclua `summary` e `description` claros e acionáveis
   - Adicione `impact` para comunicar a severidade do negócio
   - Forneça `action` com passos iniciais de resolução
   - Inclua `dashboard_url` e `runbook_url` para diagnóstico rápido

4. **Conformidade Regulatória**
   - Implemente alertas específicos para métricas relacionadas à conformidade (PCI DSS, GDPR/LGPD)
   - Documente a rastreabilidade entre alertas e requisitos regulatórios
   - Configure retenção de histórico de alertas conforme requisitos de auditoria

### 2. Governança e Manutenção

1. **Processo de Aprovação**
   - Todas as novas regras de alerta devem passar por revisão por pares
   - Mudanças em alertas críticos exigem aprovação da equipe de SRE
   - Utilize GitOps para controle de versão e aprovação de alterações

2. **Ciclo de Revisão**
   - Revisão trimestral de todas as regras de alerta
   - Análise de eficácia com base em métricas de alerta (taxa de falsos positivos/negativos)
   - Ajuste de limiares com base em dados históricos e feedback operacional

3. **Integração com ITSM**
   - Alertas críticos (P1/P2) devem gerar tickets automaticamente no sistema ITSM
   - Implementar correlação de eventos para reduzir duplicação de tickets
   - Manter rastreabilidade entre alertas, tickets e resolução

4. **Documentação e Treinamento**
   - Manter este documento atualizado com todas as regras de alerta aprovadas
   - Desenvolver e atualizar runbooks para cada categoria de alerta
   - Conduzir treinamentos periódicos sobre resposta a alertas para as equipes operacionais

### 3. Métricas de Eficácia

A eficácia das regras de alerta deve ser medida pelos seguintes KPIs:

1. **Precisão de Alertas**
   - Taxa de falsos positivos < 5%
   - Taxa de falsos negativos < 1% (para incidentes P1/P2)

2. **Tempo de Resposta**
   - Tempo médio até reconhecimento (MTTA) < 5 minutos para P1, < 15 minutos para P2
   - Tempo médio até resolução (MTTR) dentro dos SLAs definidos por prioridade

3. **Qualidade Operacional**
   - % de alertas com runbooks associados > 98%
   - % de alertas resolvidos sem escalação > 85%
   - % de incidentes cobertos por alertas preditivos > 75%

## V. Conclusão

Este documento estabelece o padrão para regras de alerta na plataforma INNOVABIZ, garantindo que todas as equipes sigam uma abordagem consistente, completa e orientada a negócios para monitoramento e resposta a incidentes.

As regras e diretrizes aqui estabelecidas estão alinhadas com:

- Requisitos de conformidade (PCI DSS 4.0, GDPR/LGPD, ISO 27001)
- Padrões internacionais de observabilidade (SRE, DevOps, ITIL)
- Arquitetura hexagonal e modular da plataforma INNOVABIZ
- Propagação de contexto multi-dimensional em todos os componentes
- Estratégia de observabilidade orientada a negócios

Para sugestões de melhorias neste documento ou adição de novas regras de alerta, siga o processo de governança descrito na Seção IV.2.

## Apêndice A: Referência Rápida de Rótulos

| Rótulo | Descrição | Exemplo de Valores |
|--------|-----------|-------------------|
| `tenant_id` | Identificador único do tenant | `acme_corp`, `innovabiz_internal` |
| `region_id` | Região geográfica de operação | `br`, `us`, `eu`, `ao` |
| `environment` | Ambiente de implantação | `production`, `staging`, `development` |
| `component` | Componente principal | `iam`, `payment_gateway`, `api_gateway` |
| `module` | Submódulo específico | `authentication`, `transaction_processing` |
| `severity` | Gravidade do alerta | `critical`, `high`, `medium`, `low` |
| `category` | Categoria do problema | `availability`, `latency`, `errors`, `saturation` |
| `priority` | Prioridade de negócio | `P1`, `P2`, `P3`, `P4` |
| `sla` | Tempo máximo de resolução | `15m`, `30m`, `2h`, `8h` |
| `team` | Equipe responsável | `platform`, `security`, `payments`, `database` |

---

**Última Atualização:** 26 de Julho de 2025  
**Autor:** INNOVABIZ Platform Team  
**Versão:** 1.0

*Este documento é parte da documentação oficial de observabilidade da plataforma INNOVABIZ e está sujeito ao controle de versão e processo de governança documental.*