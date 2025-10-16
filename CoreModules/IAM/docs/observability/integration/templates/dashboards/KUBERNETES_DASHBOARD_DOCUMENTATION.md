# Dashboard Kubernetes - Documentação Técnica

## 📊 Visão Geral

O **Dashboard Kubernetes** da plataforma INNOVABIZ fornece monitoramento abrangente e em tempo real de clusters Kubernetes, oferecendo visibilidade completa sobre recursos, performance, saúde dos workloads e utilização de infraestrutura. Este dashboard é essencial para operações DevOps, SRE e gestão de infraestrutura cloud-native.

### 🎯 Objetivos Principais
- **Monitoramento de Cluster**: Status geral e saúde dos clusters Kubernetes
- **Gestão de Recursos**: CPU, memória, disco e rede por node e pod
- **Observabilidade de Workloads**: Deployments, pods, containers e serviços
- **Detecção Proativa**: Identificação de problemas antes do impacto ao usuário
- **Otimização de Performance**: Insights para tuning e scaling
- **Compliance e Governança**: Aderência a políticas e limites de recursos

---

## 🏗️ Arquitetura e Componentes

### 📡 Fontes de Dados
```yaml
Prometheus Metrics:
  - kube-state-metrics: Estado dos objetos Kubernetes
  - node-exporter: Métricas de sistema operacional
  - cadvisor: Métricas de containers
  - kubelet: Métricas do runtime Kubernetes

Coletores Principais:
  - kube_pod_info: Informações de pods
  - kube_node_info: Informações de nodes
  - kube_deployment_status: Status de deployments
  - container_cpu_usage_seconds_total: CPU usage
  - container_memory_working_set_bytes: Memory usage
  - node_filesystem_avail_bytes: Disk usage
```

### 🎛️ Variáveis Multi-Contexto
```yaml
tenant_id:
  - Descrição: Identificador único do tenant
  - Tipo: Multi-select com "All"
  - Query: label_values(kube_pod_info, tenant_id)
  - Dependências: Nenhuma

region_id:
  - Descrição: Região geográfica do cluster
  - Tipo: Multi-select com "All"
  - Query: label_values(kube_pod_info{tenant_id=~"$tenant_id"}, region_id)
  - Dependências: tenant_id

environment:
  - Descrição: Ambiente (dev, staging, prod)
  - Tipo: Multi-select com "All"
  - Query: label_values(kube_pod_info{tenant_id=~"$tenant_id", region_id=~"$region_id"}, environment)
  - Dependências: tenant_id, region_id

cluster:
  - Descrição: Nome do cluster Kubernetes
  - Tipo: Multi-select com "All"
  - Query: label_values(kube_pod_info{tenant_id=~"$tenant_id", region_id=~"$region_id", environment=~"$environment"}, cluster)
  - Dependências: tenant_id, region_id, environment

namespace:
  - Descrição: Namespace Kubernetes
  - Tipo: Multi-select com "All"
  - Query: label_values(kube_pod_info{tenant_id=~"$tenant_id", region_id=~"$region_id", environment=~"$environment", cluster=~"$cluster"}, namespace)
  - Dependências: tenant_id, region_id, environment, cluster
```

---

## 📊 Painéis e Métricas

### 🏢 Cluster Overview
Visão geral do status e recursos do cluster.

#### Cluster Status
```yaml
Tipo: Stat Panel
Métrica: up{job="kube-state-metrics"}
Objetivo: Verificar disponibilidade do cluster
Thresholds:
  - Red (0): Cluster indisponível
  - Green (1): Cluster operacional
Alertas: Crítico se cluster down por >2min
```

#### Total Nodes
```yaml
Tipo: Stat Panel
Métrica: count(kube_node_info)
Objetivo: Quantidade total de nodes no cluster
Baseline: Varia por cluster (3-100+ nodes)
Tendência: Crescimento conforme scaling
```

#### Ready Nodes
```yaml
Tipo: Stat Panel
Métrica: sum(kube_node_status_condition{condition="Ready", status="true"})
Objetivo: Nodes prontos para receber workloads
Target: 100% dos nodes ready
Alerta: <95% nodes ready
```

#### Total Pods / Running Pods
```yaml
Tipo: Stat Panel
Métricas:
  - count(kube_pod_info): Total de pods
  - sum(kube_pod_status_phase{phase="Running"}): Pods em execução
Objetivo: Visão geral dos workloads
Target: >95% pods running
```

#### CPU Usage by Node
```yaml
Tipo: Time Series
Métrica: 100 - (avg by (node) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)
Objetivo: Utilização de CPU por node
Thresholds:
  - Green: 0-70%
  - Yellow: 70-90%
  - Red: >90%
Alerta: >85% por 5min
```

### 💾 Pod Resources
Monitoramento detalhado de recursos dos pods.

#### Pod CPU Usage
```yaml
Tipo: Time Series
Métrica: sum(rate(container_cpu_usage_seconds_total{container!="POD"}[5m])) by (namespace, pod)
Objetivo: Consumo de CPU por pod
Unit: Cores
Thresholds:
  - Green: 0-0.5 cores
  - Yellow: 0.5-1 core
  - Red: >1 core
Display: Table legend com current/max values
```

#### Pod Memory Usage
```yaml
Tipo: Time Series
Métrica: sum(container_memory_working_set_bytes{container!="POD"}) by (namespace, pod)
Objetivo: Consumo de memória por pod
Unit: Bytes
Thresholds:
  - Green: 0-512MB
  - Yellow: 512MB-1GB
  - Red: >1GB
Display: Table legend com current/max values
```

#### Network I/O
```yaml
Tipo: Time Series
Métricas:
  - RX: sum(rate(container_network_receive_bytes_total[5m])) by (namespace, pod)
  - TX: sum(rate(container_network_transmit_bytes_total[5m])) by (namespace, pod)
Objetivo: Tráfego de rede por pod
Unit: Bytes per second
Baseline: Varia por aplicação
```

#### Disk I/O
```yaml
Tipo: Time Series
Métricas:
  - Read: sum(rate(container_fs_reads_bytes_total[5m])) by (namespace, pod)
  - Write: sum(rate(container_fs_writes_bytes_total[5m])) by (namespace, pod)
Objetivo: I/O de disco por pod
Unit: Bytes per second
Baseline: Varia por aplicação
```

### 🚀 Workload Status
Status e saúde dos workloads Kubernetes.

#### Deployment Status
```yaml
Tipo: Table
Métricas:
  - Desired: kube_deployment_status_replicas
  - Available: kube_deployment_status_replicas_available
Objetivo: Status de deployments
Colunas: Namespace, Deployment, Desired, Available
Filtros: Searchable e sortable
Alerta: Available < Desired
```

#### Pod Restarts
```yaml
Tipo: Time Series
Métrica: increase(kube_pod_container_status_restarts_total[1h])
Objetivo: Monitorar reinicializações de containers
Thresholds:
  - Green: 0 restarts
  - Yellow: 1-4 restarts
  - Red: >5 restarts
Alerta: >3 restarts em 1h
```

#### Pod Status Distribution
```yaml
Tipo: Pie Chart
Métrica: sum by (phase) (kube_pod_status_phase)
Objetivo: Distribuição de status dos pods
Phases: Running, Pending, Failed, Succeeded
Target: >90% Running
Display: Percentages e valores absolutos
```

---

## 🚨 Anotações e Alertas

### 📍 Anotações Automáticas
```yaml
Kubernetes Events:
  - Trigger: increase(kube_pod_container_status_restarts_total[5m]) > 0
  - Icon: Red
  - Title: "Pod Restart"
  - Description: "{{namespace}}/{{pod}} restarted"

Node Issues:
  - Trigger: kube_node_status_condition{condition="Ready", status="false"} == 1
  - Icon: Orange
  - Title: "Node Not Ready"
  - Description: "Node {{node}} is not ready"

Resource Alerts:
  - Trigger: (kube_pod_container_resource_requests{resource="memory"} / kube_pod_container_resource_limits{resource="memory"}) > 0.9
  - Icon: Yellow
  - Title: "High Memory Usage"
  - Description: "{{namespace}}/{{pod}} high memory usage"
```

### ⚠️ Alertas Críticos
```yaml
Cluster Down:
  - Condition: up{job="kube-state-metrics"} == 0
  - Severity: Critical
  - Duration: 2 minutes
  - Action: Immediate escalation

Node Not Ready:
  - Condition: kube_node_status_condition{condition="Ready", status="false"} == 1
  - Severity: High
  - Duration: 5 minutes
  - Action: Infrastructure team notification

High CPU Usage:
  - Condition: node_cpu_usage > 90%
  - Severity: Warning
  - Duration: 10 minutes
  - Action: Auto-scaling trigger

Memory Pressure:
  - Condition: node_memory_usage > 95%
  - Severity: High
  - Duration: 5 minutes
  - Action: Pod eviction warning

Pod Crash Loop:
  - Condition: increase(kube_pod_container_status_restarts_total[15m]) > 5
  - Severity: High
  - Duration: 0 minutes
  - Action: Development team alert
```

---

## 🔧 Configuração e Personalização

### ⚙️ Configurações Recomendadas
```yaml
Refresh Rate: 30 segundos
Time Range: Última 1 hora (padrão)
Auto-refresh: Habilitado
Theme: Dark (melhor para NOC/SOC)
Timezone: UTC (padrão global)

Variáveis:
  - Include All: Habilitado para todas
  - Multi-value: Habilitado para todas
  - Refresh on Dashboard Load: Habilitado
```

### 🎨 Customizações por Ambiente
```yaml
Development:
  - Refresh: 1 minuto
  - Retention: 7 dias
  - Alertas: Reduzidos

Staging:
  - Refresh: 30 segundos
  - Retention: 30 dias
  - Alertas: Moderados

Production:
  - Refresh: 15 segundos
  - Retention: 90 dias
  - Alertas: Completos
  - Escalação: Automática
```

---

## 📋 Casos de Uso Operacionais

### 🔍 Troubleshooting
```yaml
High CPU Usage:
  1. Identificar pods com alto consumo
  2. Verificar métricas de aplicação
  3. Analisar logs dos containers
  4. Considerar scaling horizontal/vertical

Memory Leaks:
  1. Monitorar tendência de memória
  2. Identificar pods com crescimento constante
  3. Analisar heap dumps (se Java)
  4. Implementar memory limits

Network Issues:
  1. Verificar I/O de rede anômalo
  2. Analisar conectividade entre pods
  3. Verificar políticas de rede
  4. Monitorar latência de DNS

Storage Problems:
  1. Monitorar uso de disco
  2. Verificar PV/PVC status
  3. Analisar I/O patterns
  4. Considerar storage scaling
```

### 📊 Capacity Planning
```yaml
Node Scaling:
  - Trigger: CPU/Memory >80% por 30min
  - Action: Add nodes via auto-scaler
  - Validation: Workload distribution

Pod Scaling:
  - Trigger: HPA metrics threshold
  - Action: Horizontal pod autoscaling
  - Validation: Performance maintenance

Resource Optimization:
  - Analysis: Resource requests vs usage
  - Action: Adjust limits/requests
  - Validation: No performance degradation
```

### 🚨 Incident Response
```yaml
Cluster Outage:
  1. Verify cluster status panel
  2. Check node availability
  3. Analyze recent deployments
  4. Execute disaster recovery

Pod Failures:
  1. Identify failing pods
  2. Check restart patterns
  3. Analyze container logs
  4. Verify resource constraints

Performance Degradation:
  1. Compare current vs baseline
  2. Identify resource bottlenecks
  3. Check external dependencies
  4. Implement temporary fixes
```

---

## 📚 Governança e Compliance

### 🛡️ Segurança e Acesso
```yaml
RBAC Integration:
  - View Access: All authenticated users
  - Edit Access: DevOps/SRE teams only
  - Admin Access: Platform administrators
  - Audit: All access logged

Data Privacy:
  - Tenant Isolation: Enforced via labels
  - PII Protection: No sensitive data in metrics
  - Retention: Aligned with data governance
  - Export: Controlled and audited
```

### 📊 SLAs e Métricas
```yaml
Availability SLA:
  - Target: 99.9% cluster uptime
  - Measurement: up{job="kube-state-metrics"}
  - Reporting: Monthly SLA reports

Performance SLA:
  - CPU: <80% average utilization
  - Memory: <85% average utilization
  - Network: <1s pod-to-pod latency
  - Storage: <100ms I/O latency

Recovery SLA:
  - Detection: <5 minutes
  - Response: <15 minutes
  - Resolution: <1 hour (P1), <4 hours (P2)
```

### 🔄 Manutenção e Updates
```yaml
Dashboard Updates:
  - Frequency: Monthly
  - Testing: Staging environment first
  - Rollback: Automated if issues detected
  - Documentation: Updated with changes

Metric Retention:
  - Raw metrics: 15 days
  - 5min aggregation: 90 days
  - 1hour aggregation: 1 year
  - Daily aggregation: 3 years

Backup Strategy:
  - Dashboard config: Daily backup
  - Historical data: Weekly backup
  - Recovery testing: Monthly
  - DR procedures: Quarterly review
```

---

*Documento técnico aprovado pela equipe de SRE em: 2025-01-31*  
*Versão: 1.0*  
*Próxima revisão: 2025-04-30*  
*Responsável: SRE Team + Platform Engineering*  
*Classificação: Interno*