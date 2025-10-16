# INNOVABIZ IAM Audit Service - Documentação Prometheus

**Versão:** 2.0.0  
**Data:** 31/07/2025  
**Classificação:** Oficial  
**Status:** ✅ Implementado  
**Autor:** Equipe INNOVABIZ DevOps  
**Aprovado por:** Eduardo Jeremias  

## 1. Visão Geral

O Prometheus é o componente central de coleta e armazenamento de métricas na arquitetura de observabilidade do IAM Audit Service da INNOVABIZ. Sua implementação segue os princípios de multi-tenancy, escalabilidade e conformidade regulatória, conforme definido nos requisitos da plataforma INNOVABIZ.

### 1.1 Funcionalidades Principais

- **Coleta de Métricas**: Pull-based de targets definidos via service discovery
- **Armazenamento de Séries Temporais**: TSDB otimizada para alta disponibilidade
- **Linguagem de Consulta**: PromQL para consultas e análises avançadas
- **Alertas**: Regras de alerta com roteamento via AlertManager
- **Integração**: Exportação de dados para Grafana e sistemas externos
- **Multi-Contextualidade**: Suporte a labels de tenant, região e ambiente

### 1.2 Posicionamento na Arquitetura

O Prometheus atua como o serviço central de métricas, recebendo dados de:

- Aplicações instrumentadas diretamente (IAM Audit Service)
- OpenTelemetry Collector (via exporters)
- Exporters específicos (Node Exporter, cAdvisor, etc.)
- ServiceMonitors (CRD do operador Prometheus)

E fornecendo dados para:

- Grafana (visualização)
- AlertManager (alertas)
- API de observabilidade (consultas externas)
- Long-term storage (retenção estendida)

## 2. Implementação Técnica

### 2.1 Manifesto Kubernetes

O Prometheus é implementado como um StatefulSet no Kubernetes, conforme definido em `observability/prometheus.yaml`. Os principais componentes incluem:

```yaml
# Trecho exemplificativo do manifesto
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: prometheus
  namespace: iam-system
  labels:
    app.kubernetes.io/name: prometheus
    app.kubernetes.io/part-of: innovabiz-observability
    innovabiz.com/module: iam-audit
    innovabiz.com/tier: observability
spec:
  replicas: 2
  # ... outras configurações
```

### 2.2 Recursos Computacionais

| Recurso | Requisito | Limite | Observações |
|---------|-----------|--------|-------------|
| **CPU** | 1000m | 2000m | Escalável conforme volume de métricas |
| **Memória** | 2Gi | 4Gi | Dimensionado para retenção de 15 dias |
| **Armazenamento** | 50Gi | - | PVC com StorageClass de alta performance |
| **Rede** | - | - | Tráfego de entrada limitado a clusters autorizados |

### 2.3 Estratégia de Persistência

- **Volume Persistente**: StorageClass SSD para TSDB
- **Retenção**: 15 dias de dados de alta resolução
- **Compactação**: Ativada para otimização de espaço
- **Backup**: Snapshots diários para armazenamento frio
- **Recuperação**: Procedimento documentado no runbook de DR

### 2.4 Segurança e Controle de Acesso

- **Authentication**: Básica via proxy reverso (Ingress)
- **Authorization**: RBAC Kubernetes + ACL interna
- **TLS**: Obrigatório para todos os endpoints externos
- **Rede**: Isolamento via NetworkPolicy Kubernetes
- **Auditoria**: Logging de todas as consultas e configurações

### 2.5 Escalabilidade e Alta Disponibilidade

- **Horizontal**: Múltiplas réplicas com sharding federado
- **Vertical**: Recursos alocados dinamicamente via análise de uso
- **Disponibilidade**: Distribuição em múltiplas zonas de disponibilidade
- **Consistência**: Consistência eventual entre réplicas
- **Degradação**: Modo graceful de falha com alerta automático

## 3. Configuração Multi-dimensional

### 3.1 Modelo de Labels Multi-contexto

O Prometheus utiliza um esquema padronizado de labels para garantir a capacidade multi-dimensional:

```
tenant_id="tenant1"
region_id="br-east-1"
environment="production"
module="iam-audit"
component="authentication-service"
instance="pod-name"
```

### 3.2 Isolamento por Tenant

- **Separação Lógica**: Labels de tenant em todas as métricas
- **Autorização**: Filtros automáticos por tenant na camada de API
- **Agregação**: Regras de recording separadas por tenant
- **Alertas**: Threshold configurável por tenant (standard/premium)
- **Retenção**: Políticas específicas para tenants VIP

### 3.3 Contexto Regional

- **Labels Regionais**: Identificação de região em todas as métricas
- **Consultas Cross-region**: Suporte via PromQL federado
- **Alertas Específicos**: Thresholds adaptados por região
- **Normalização**: Ajuste automático para fusos horários

### 3.4 Modelo de Federation

Para ambientes muito grandes, implementamos federação Prometheus:

- **Prometheus Local**: Por região/ambiente (15 dias de retenção)
- **Prometheus Global**: Agregação multi-região (30 dias de métricas sumarizadas)
- **Thanos**: Long-term storage para retenção estendida (365 dias)

## 4. Catálogo de Métricas

### 4.1 Métricas Padrão

| Métrica | Tipo | Descrição | Labels |
|---------|------|-----------|--------|
| `http_requests_total` | Counter | Total de requisições HTTP | method, path, status, tenant_id, region_id |
| `http_request_duration_seconds` | Histogram | Latência de requisições HTTP | method, path, status, tenant_id, region_id |
| `authentication_attempts_total` | Counter | Tentativas de autenticação | result, method, tenant_id, region_id |
| `authorization_checks_total` | Counter | Verificações de autorização | result, resource, action, tenant_id, region_id |
| `database_operations_total` | Counter | Operações de banco de dados | operation, table, result, tenant_id, region_id |
| `database_operation_duration_seconds` | Histogram | Latência de operações de banco | operation, table, result, tenant_id, region_id |
| `audit_events_total` | Counter | Eventos de auditoria gerados | event_type, severity, tenant_id, region_id |
| `audit_processing_duration_seconds` | Histogram | Tempo de processamento de auditoria | event_type, tenant_id, region_id |

### 4.2 Métricas de Infraestrutura

| Métrica | Tipo | Descrição | Labels |
|---------|------|-----------|--------|
| `node_cpu_seconds_total` | Counter | Uso de CPU por nó | mode, cpu, tenant_id, region_id |
| `node_memory_MemAvailable_bytes` | Gauge | Memória disponível por nó | tenant_id, region_id |
| `container_cpu_usage_seconds_total` | Counter | Uso de CPU por container | container, pod, namespace, tenant_id, region_id |
| `container_memory_usage_bytes` | Gauge | Uso de memória por container | container, pod, namespace, tenant_id, region_id |
| `kube_pod_status_phase` | Gauge | Estado de pods Kubernetes | pod, namespace, phase, tenant_id, region_id |
| `kube_deployment_status_replicas_available` | Gauge | Réplicas disponíveis por deployment | deployment, namespace, tenant_id, region_id |

### 4.3 Métricas de Desempenho

| Métrica | Tipo | Descrição | Labels |
|---------|------|-----------|--------|
| `api_request_rate` | Gauge | Taxa de requisições por segundo | api, method, tenant_id, region_id |
| `api_error_rate` | Gauge | Taxa de erros por segundo | api, method, error_type, tenant_id, region_id |
| `api_latency_percentiles` | Gauge | Percentis de latência (p50, p95, p99) | api, method, percentile, tenant_id, region_id |
| `system_saturation` | Gauge | Nível de saturação do sistema (0-1) | component, tenant_id, region_id |
| `resource_utilization` | Gauge | Utilização de recursos (0-1) | resource_type, component, tenant_id, region_id |

### 4.4 Métricas de Negócio

| Métrica | Tipo | Descrição | Labels |
|---------|------|-----------|--------|
| `active_sessions` | Gauge | Sessões ativas no momento | auth_method, tenant_id, region_id |
| `user_activity` | Counter | Atividades de usuário registradas | activity_type, tenant_id, region_id |
| `security_events` | Counter | Eventos de segurança detectados | severity, event_type, tenant_id, region_id |
| `compliance_checks` | Counter | Verificações de compliance realizadas | result, check_type, standard, tenant_id, region_id |

## 5. Regras de Alertas

### 5.1 Alertas de Disponibilidade

```yaml
groups:
- name: availability
  rules:
  - alert: ServiceDown
    expr: up{job="iam-audit-service"} == 0
    for: 1m
    labels:
      severity: critical
      category: availability
    annotations:
      summary: "IAM Audit Service indisponível"
      description: "O serviço está inacessível há mais de 1 minuto."
```

### 5.2 Alertas de Performance

```yaml
groups:
- name: performance
  rules:
  - alert: HighLatency
    expr: histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{job="iam-audit-service"}[5m])) by (le, tenant_id, region_id)) > 0.5
    for: 5m
    labels:
      severity: warning
      category: performance
    annotations:
      summary: "Latência elevada no IAM Audit Service"
      description: "P95 da latência acima de 500ms por 5 minutos."
      runbook_url: "https://docs.innovabiz.com/runbooks/high-latency"
```

### 5.3 Alertas de Negócio

```yaml
groups:
- name: business
  rules:
  - alert: HighAuthFailureRate
    expr: sum(rate(authentication_attempts_total{result="failure"}[5m])) by (tenant_id) / sum(rate(authentication_attempts_total[5m])) by (tenant_id) > 0.2
    for: 5m
    labels:
      severity: critical
      category: security
    annotations:
      summary: "Taxa elevada de falhas de autenticação"
      description: "Mais de 20% das tentativas de autenticação estão falhando."
      runbook_url: "https://docs.innovabiz.com/runbooks/high-auth-failure"
```

## 6. Integrações

### 6.1 Integração com Grafana

- **Datasource**: Prometheus configurado como fonte principal no Grafana
- **Dashboards**: Pré-configurados para visualização multi-dimensional
- **Alerting**: Regras sincronizadas com AlertManager
- **Exploração**: Interface de consulta PromQL integrada

### 6.2 Integração com OpenTelemetry

- **Receiver**: OTLP/HTTP e OTLP/gRPC para métricas externas
- **Processador**: Batch e filtragem para otimização
- **Exportador**: Prometheus Remote Write API
- **Contexto**: Propagação de contexto multi-dimensional

### 6.3 Integração com AlertManager

- **Roteamento**: Configuração por severidade e contexto
- **Agrupamento**: Inteligente para reduzir fadiga de alertas
- **Silenciamento**: Regras para manutenções programadas
- **Escalação**: Fluxo baseado em SLA por tipo de alerta

### 6.4 Integração com Long-term Storage

- **Thanos**: Utilizado para armazenamento de longa duração
- **Compactação**: Downsampling para otimização de espaço
- **Retenção**: Políticas graduais (15d → 30d → 90d → 365d)
- **Consulta**: Interface unificada para dados históricos

## 7. Monitoramento do Monitoramento

O próprio Prometheus é monitorado através de métricas expostas:

- `prometheus_tsdb_head_samples_appended_total`: Taxa de ingestão
- `prometheus_tsdb_storage_blocks_bytes`: Consumo de armazenamento
- `prometheus_engine_query_duration_seconds`: Latência de consulta
- `prometheus_target_scrape_pool_targets`: Total de targets
- `prometheus_target_scrape_pool_sync_total`: Sincronização de service discovery

## 8. Operação e Manutenção

### 8.1 Procedimentos Operacionais

- **Backup**: Automático a cada 6 horas para object storage
- **Verificação de Integridade**: Validação diária da TSDB
- **Compactação**: Agendada para períodos de baixo tráfego
- **Rotação de Logs**: A cada 24 horas com compressão
- **Verificação de Rules**: Validação automática pré-deployment

### 8.2 Troubleshooting

| Problema | Possíveis Causas | Resolução |
|----------|-----------------|-----------|
| Alto uso de CPU | Consultas complexas, muitos targets | Otimizar consultas, aumentar recursos |
| Alto uso de memória | Muitas séries temporais, regras ineficientes | Revisão de cardinalidade, otimização de regras |
| Falha na coleta | Target inacessível, timeout de rede | Verificar conectividade, aumentar timeout |
| Alertas duplicados | Múltiplas instâncias disparando o mesmo alerta | Implementar deduplicação no AlertManager |
| Consultas lentas | Excesso de cardinalidade, range muito amplo | Otimizar consultas, usar recording rules |

### 8.3 Upgrades

- **Planejamento**: Testes em ambiente não-produtivo
- **Janela**: Durante período de baixo tráfego
- **Rollout**: Gradual, uma réplica por vez
- **Validação**: Testes automatizados pós-upgrade
- **Rollback**: Procedimento documentado e testado

## 9. Conformidade e Segurança

### 9.1 Requisitos de Conformidade

| Regulação | Requisito | Implementação |
|-----------|-----------|---------------|
| **PCI DSS 4.0** | 10.2 Implementar trilhas de auditoria | Métricas de auditoria com retenção apropriada |
| **ISO 27001** | A.12.4 Registros e monitoramento | Alertas e dashboards para análise de segurança |
| **GDPR/LGPD** | Art. 46 Segurança do tratamento | Mascaramento de dados sensíveis em labels |
| **NIST 800-53** | SI-4 Monitoramento do sistema | Cobertura completa de componentes críticos |

### 9.2 Controles de Segurança

- **Criptografia**: TLS 1.3 para todas as comunicações
- **Autenticação**: mTLS para comunicações entre serviços
- **Autorização**: RBAC Kubernetes e filtros por tenant
- **Isolamento**: NetworkPolicies restritivas
- **Auditoria**: Logging de todas as operações administrativas

## 10. Referências

1. [Prometheus Official Documentation](https://prometheus.io/docs/introduction/overview/)
2. [Kubernetes Monitoring Best Practices](https://kubernetes.io/docs/tasks/debug-application-cluster/resource-usage-monitoring/)
3. [OpenTelemetry Metrics](https://opentelemetry.io/docs/reference/specification/metrics/)
4. [PCI DSS 4.0 Monitoring Requirements](https://www.pcisecuritystandards.org/)
5. [NIST SP 800-53 Rev. 5](https://csrc.nist.gov/publications/detail/sp/800-53/rev-5/final)
6. [SRE Book: Monitoring Distributed Systems](https://sre.google/sre-book/monitoring-distributed-systems/)
7. [Observability Engineering](https://www.oreilly.com/library/view/observability-engineering/9781492076438/)

---

*Documento aprovado pelo Comitê de Arquitetura INNOVABIZ em 28/07/2025*