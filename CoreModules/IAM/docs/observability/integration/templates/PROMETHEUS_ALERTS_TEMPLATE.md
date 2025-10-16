# Template para Alertas Prometheus - INNOVABIZ

## Visão Geral

Este documento fornece o template padrão INNOVABIZ para configuração de alertas Prometheus, assegurando consistência nas práticas de monitoramento e alertas em toda a plataforma. As regras de alerta seguem os princípios multi-dimensionais da plataforma INNOVABIZ, permitindo seletividade por tenant, região, módulo e outros contextos relevantes.

## Estrutura de Alertas

A plataforma INNOVABIZ organiza alertas Prometheus em uma hierarquia padronizada:

1. **Alertas Críticos** - Impacto imediato no serviço, requer ação urgente (página)
2. **Alertas de Aviso** - Potencial degradação de serviço, requer investigação
3. **Alertas Informativos** - Mudanças significativas que devem ser notadas, mas não requerem ação imediata

Cada módulo deve implementar alertas nestas categorias, adaptando métricas e limiares específicos às suas necessidades.

## Template de Configuração YAML

Abaixo está o template YAML para regras de alertas Prometheus, seguindo as práticas recomendadas INNOVABIZ:

```yaml
# prometheus-rules.yaml
# Template de regras de alerta para módulos INNOVABIZ
# Compatível com Prometheus 2.40.0+

groups:
  - name: innovabiz_{{module_id}}_availability
    rules:
    # ALERTA CRÍTICO: Serviço indisponível
    - alert: ServiceDown
      expr: up{job="{{service_name}}", innovabiz_tenant_id="$TENANT_ID", innovabiz_region_id="$REGION_ID", innovabiz_module_id="{{module_id}}"} == 0
      for: 1m
      labels:
        severity: critical
        category: availability
        team: sre
        innovabiz_tenant_id: $TENANT_ID
        innovabiz_region_id: $REGION_ID
        innovabiz_module_id: "{{module_id}}"
      annotations:
        summary: "Serviço {{service_name}} indisponível"
        description: "O serviço {{service_name}} está indisponível por pelo menos 1 minuto para tenant={{ $labels.innovabiz_tenant_id }}, região={{ $labels.innovabiz_region_id }}."
        runbook_url: "https://wiki.innovabiz.com/observability/runbooks/{{module_id}}/service-down"
        dashboard_url: "https://grafana.innovabiz.com/d/innovabiz-{{module_id}}-operational/{{module_id}}-operational?var-service={{service_name}}&var-tenant={{ $labels.innovabiz_tenant_id }}&var-region={{ $labels.innovabiz_region_id }}"
        priority: P1

    # ALERTA CRÍTICO: Alta taxa de erros
    - alert: HighErrorRate
      expr: |
        sum(rate(http_server_requests_seconds_count{status_code=~"5..", job="{{service_name}}", innovabiz_tenant_id="$TENANT_ID", innovabiz_region_id="$REGION_ID", innovabiz_module_id="{{module_id}}"}[5m]))
        /
        sum(rate(http_server_requests_seconds_count{job="{{service_name}}", innovabiz_tenant_id="$TENANT_ID", innovabiz_region_id="$REGION_ID", innovabiz_module_id="{{module_id}}"}[5m]))
        > 0.05
      for: 2m
      labels:
        severity: critical
        category: errors
        team: devops
        innovabiz_tenant_id: $TENANT_ID
        innovabiz_region_id: $REGION_ID
        innovabiz_module_id: "{{module_id}}"
      annotations:
        summary: "Alta taxa de erros em {{service_name}}"
        description: "O serviço {{service_name}} está apresentando >5% de erros 5xx nos últimos 2 minutos para tenant={{ $labels.innovabiz_tenant_id }}, região={{ $labels.innovabiz_region_id }}."
        runbook_url: "https://wiki.innovabiz.com/observability/runbooks/{{module_id}}/high-error-rate"
        dashboard_url: "https://grafana.innovabiz.com/d/innovabiz-{{module_id}}-errors/{{module_id}}-troubleshooting?var-service={{service_name}}&var-tenant={{ $labels.innovabiz_tenant_id }}&var-region={{ $labels.innovabiz_region_id }}"
        priority: P1

  - name: innovabiz_{{module_id}}_latency
    rules:
    # ALERTA DE AVISO: Alta latência
    - alert: HighLatency
      expr: |
        histogram_quantile(0.95, sum(rate(http_server_requests_seconds_bucket{job="{{service_name}}", innovabiz_tenant_id="$TENANT_ID", innovabiz_region_id="$REGION_ID", innovabiz_module_id="{{module_id}}"}[5m])) by (le, route)) > 1.0
      for: 5m
      labels:
        severity: warning
        category: performance
        team: developers
        innovabiz_tenant_id: $TENANT_ID
        innovabiz_region_id: $REGION_ID
        innovabiz_module_id: "{{module_id}}"
      annotations:
        summary: "Alta latência em {{service_name}} ({{ $labels.route }})"
        description: "O endpoint {{ $labels.route }} está apresentando latência p95 > 1s nos últimos 5 minutos para tenant={{ $labels.innovabiz_tenant_id }}, região={{ $labels.innovabiz_region_id }}."
        runbook_url: "https://wiki.innovabiz.com/observability/runbooks/{{module_id}}/high-latency"
        dashboard_url: "https://grafana.innovabiz.com/d/innovabiz-{{module_id}}-performance/{{module_id}}-performance?var-service={{service_name}}&var-route={{ $labels.route }}&var-tenant={{ $labels.innovabiz_tenant_id }}&var-region={{ $labels.innovabiz_region_id }}"
        priority: P2

  - name: innovabiz_{{module_id}}_resources
    rules:
    # ALERTA DE AVISO: Alto uso de CPU
    - alert: HighCpuUsage
      expr: |
        avg(rate(process_cpu_seconds_total{job="{{service_name}}", innovabiz_tenant_id="$TENANT_ID", innovabiz_region_id="$REGION_ID", innovabiz_module_id="{{module_id}}"}[5m])) by (pod) > 0.8
      for: 10m
      labels:
        severity: warning
        category: resources
        team: sre
        innovabiz_tenant_id: $TENANT_ID
        innovabiz_region_id: $REGION_ID
        innovabiz_module_id: "{{module_id}}"
      annotations:
        summary: "Alto consumo de CPU em {{service_name}} ({{ $labels.pod }})"
        description: "O pod {{ $labels.pod }} está utilizando >80% de CPU por 10 minutos para tenant={{ $labels.innovabiz_tenant_id }}, região={{ $labels.innovabiz_region_id }}."
        runbook_url: "https://wiki.innovabiz.com/observability/runbooks/{{module_id}}/high-cpu"
        dashboard_url: "https://grafana.innovabiz.com/d/innovabiz-{{module_id}}-resources/{{module_id}}-resources?var-service={{service_name}}&var-pod={{ $labels.pod }}&var-tenant={{ $labels.innovabiz_tenant_id }}&var-region={{ $labels.innovabiz_region_id }}"
        priority: P2

    # ALERTA DE AVISO: Alto uso de memória
    - alert: HighMemoryUsage
      expr: |
        max(process_resident_memory_bytes{job="{{service_name}}", innovabiz_tenant_id="$TENANT_ID", innovabiz_region_id="$REGION_ID", innovabiz_module_id="{{module_id}}"} / container_spec_memory_limit_bytes{job="{{service_name}}"}) by (pod) > 0.85
      for: 10m
      labels:
        severity: warning
        category: resources
        team: sre
        innovabiz_tenant_id: $TENANT_ID
        innovabiz_region_id: $REGION_ID
        innovabiz_module_id: "{{module_id}}"
      annotations:
        summary: "Alto consumo de memória em {{service_name}} ({{ $labels.pod }})"
        description: "O pod {{ $labels.pod }} está utilizando >85% da memória alocada por 10 minutos para tenant={{ $labels.innovabiz_tenant_id }}, região={{ $labels.innovabiz_region_id }}."
        runbook_url: "https://wiki.innovabiz.com/observability/runbooks/{{module_id}}/high-memory"
        dashboard_url: "https://grafana.innovabiz.com/d/innovabiz-{{module_id}}-resources/{{module_id}}-resources?var-service={{service_name}}&var-pod={{ $labels.pod }}&var-tenant={{ $labels.innovabiz_tenant_id }}&var-region={{ $labels.innovabiz_region_id }}"
        priority: P2

  - name: innovabiz_{{module_id}}_business
    rules:
    # ALERTA DE AVISO: Alta taxa de rejeição de transações
    - alert: HighTransactionRejectionRate
      expr: |
        sum(rate(innovabiz_transactions_count{status="rejected", job="{{service_name}}", innovabiz_tenant_id="$TENANT_ID", innovabiz_region_id="$REGION_ID", innovabiz_module_id="{{module_id}}"}[5m]))
        /
        sum(rate(innovabiz_transactions_count{job="{{service_name}}", innovabiz_tenant_id="$TENANT_ID", innovabiz_region_id="$REGION_ID", innovabiz_module_id="{{module_id}}"}[5m]))
        > 0.1
      for: 15m
      labels:
        severity: warning
        category: business
        team: product
        innovabiz_tenant_id: $TENANT_ID
        innovabiz_region_id: $REGION_ID
        innovabiz_module_id: "{{module_id}}"
      annotations:
        summary: "Alta taxa de rejeição em {{service_name}}"
        description: "O serviço {{service_name}} está apresentando >10% de transações rejeitadas nos últimos 15 minutos para tenant={{ $labels.innovabiz_tenant_id }}, região={{ $labels.innovabiz_region_id }}."
        runbook_url: "https://wiki.innovabiz.com/observability/runbooks/{{module_id}}/high-rejection-rate"
        dashboard_url: "https://grafana.innovabiz.com/d/innovabiz-{{module_id}}-business/{{module_id}}-business?var-service={{service_name}}&var-tenant={{ $labels.innovabiz_tenant_id }}&var-region={{ $labels.innovabiz_region_id }}"
        priority: P2

    # ALERTA INFORMATIVO: Volume de transações anômalo
    - alert: AnomalousTransactionVolume
      expr: |
        abs(
          sum(rate(innovabiz_transactions_count{job="{{service_name}}", innovabiz_tenant_id="$TENANT_ID", innovabiz_region_id="$REGION_ID", innovabiz_module_id="{{module_id}}"}[30m]))
          /
          sum(avg_over_time(innovabiz_transactions_count{job="{{service_name}}", innovabiz_tenant_id="$TENANT_ID", innovabiz_region_id="$REGION_ID", innovabiz_module_id="{{module_id}}"}[7d:30m] offset 30m))
          - 1
        ) > 0.3
      for: 30m
      labels:
        severity: info
        category: business
        team: analyst
        innovabiz_tenant_id: $TENANT_ID
        innovabiz_region_id: $REGION_ID
        innovabiz_module_id: "{{module_id}}"
      annotations:
        summary: "Volume anômalo de transações em {{service_name}}"
        description: "O serviço {{service_name}} está apresentando volume de transações 30% diferente da média histórica nos últimos 30 minutos para tenant={{ $labels.innovabiz_tenant_id }}, região={{ $labels.innovabiz_region_id }}."
        dashboard_url: "https://grafana.innovabiz.com/d/innovabiz-{{module_id}}-business/{{module_id}}-business?var-service={{service_name}}&var-tenant={{ $labels.innovabiz_tenant_id }}&var-region={{ $labels.innovabiz_region_id }}"
        priority: P3
```

## Configuração em Kubernetes

Os alertas devem ser implantados como ConfigMaps no Kubernetes e referenciados na configuração do Prometheus:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-alerts-{{module_id}}
  namespace: observability
  labels:
    app: prometheus
    innovabiz.module: "{{module_id}}"
data:
  {{module_id}}-alerts.yaml: |
    # Conteúdo do arquivo prometheus-rules.yaml acima
```

## Alertas Padrão Requeridos por Categoria

### 1. Alertas de Disponibilidade

| Nome | Severidade | Descrição | Limiar Recomendado |
|------|------------|-----------|-------------------|
| ServiceDown | critical | Serviço totalmente indisponível | up == 0 por 1m |
| EndpointDown | critical | Endpoint específico indisponível | probe_success == 0 por 2m |
| HighErrorRate | critical | Alta taxa de erros 5xx | >5% por 2m |
| InstanceDown | critical | Instância específica indisponível | up == 0 por 1m |

### 2. Alertas de Performance

| Nome | Severidade | Descrição | Limiar Recomendado |
|------|------------|-----------|-------------------|
| HighLatency | warning | Latência acima do aceitável | p95 > 1s por 5m |
| SlowEndpoint | warning | Endpoint específico lento | endpoint p95 > 2s por 5m |
| SlowDatabase | warning | Consultas de banco lentas | db p95 > 500ms por 5m |
| SlowExternalAPI | warning | Chamadas para APIs externas lentas | external p95 > 2s por 5m |

### 3. Alertas de Recursos

| Nome | Severidade | Descrição | Limiar Recomendado |
|------|------------|-----------|-------------------|
| HighCpuUsage | warning | Alto consumo de CPU | >80% por 10m |
| HighMemoryUsage | warning | Alto consumo de memória | >85% por 10m |
| HighDiskUsage | warning | Alto uso de disco | >85% por 30m |
| ConnectionPoolSaturation | warning | Pool de conexões saturado | >85% por 5m |

### 4. Alertas de Negócios

| Nome | Severidade | Descrição | Limiar Recomendado |
|------|------------|-----------|-------------------|
| HighTransactionRejectionRate | warning | Taxa elevada de rejeições | >10% por 15m |
| LowTransactionApprovalRate | warning | Taxa baixa de aprovações | <70% por 15m |
| AnomalousTransactionVolume | info | Volume de transações anômalo | ±30% da média por 30m |
| BusinessSLABreach | warning | SLA de negócio violado | Específico do módulo |

## Melhores Práticas

1. **Nomenclatura e Organização**
   - Siga o padrão de naming `Modulo_Categoria_ProblemaEspecífico`
   - Agrupe alertas em arquivos separados por módulo
   - Organize alertas por categoria dentro de cada arquivo

2. **Configuração Multi-dimensional**
   - Inclua sempre labels para tenant e região
   - Filtre alertas usando variáveis `$TENANT_ID` e `$REGION_ID`
   - Adicione o label `innovabiz_module_id` para facilitar a organização

3. **Thresholds e Duração**
   - Defina limiares baseados em dados históricos reais
   - Evite alertas prematuros com durações apropriadas (`for`)
   - Calibre periodicamente os limiares conforme o serviço evolui

4. **Contexto e Metadados**
   - Inclua URLs para runbooks e dashboards nas anotações
   - Adicione informações de prioridade (P1-P4)
   - Forneça descrições claras e acionáveis
   - Indique a equipe responsável no label `team`

5. **Redução de Ruído**
   - Implemente alertas com duração apropriada para evitar flapping
   - Use silenciamento para manutenções planejadas
   - Agrupe alertas relacionados para evitar tempestades de alertas
   - Considere correlação de alertas para reduzir alertas duplicados

## Integração com Sistemas de Notificação

Os alertas devem ser roteados para diferentes destinos baseados em severidade e categoria:

1. **critical**
   - Notificações em tempo real (SMS, chamadas)
   - Criação automática de incidentes
   - Dashboard de Status atualizado

2. **warning**
   - Notificações via email e canais do Slack
   - Tickets criados (mas sem incidente completo)
   - Logs detalhados para investigação

3. **info**
   - Apenas notificações via Slack em canais específicos
   - Agregação em relatórios diários/semanais
   - Sem criação de tickets

## Exemplo de Configuração do AlertManager

```yaml
# alertmanager.yaml
global:
  resolve_timeout: 5m

route:
  group_by: ['alertname', 'job', 'innovabiz_tenant_id', 'innovabiz_region_id']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h
  receiver: 'default'
  routes:
  - match:
      severity: critical
    receiver: 'pagerduty'
    continue: true
  - match:
      severity: warning
    receiver: 'slack'
    continue: true
  - match:
      severity: info
    receiver: 'slack-info'

receivers:
- name: 'default'
  email_configs:
  - to: 'alerts@innovabiz.com'
    send_resolved: true

- name: 'pagerduty'
  pagerduty_configs:
  - service_key: <pagerduty-key>
    description: '{{ .CommonAnnotations.summary }}'
    client: 'alertmanager'
    client_url: '{{ .CommonAnnotations.dashboard_url }}'
    details:
      description: '{{ .CommonAnnotations.description }}'
      runbook: '{{ .CommonAnnotations.runbook_url }}'
      tenant: '{{ .CommonLabels.innovabiz_tenant_id }}'
      region: '{{ .CommonLabels.innovabiz_region_id }}'
      module: '{{ .CommonLabels.innovabiz_module_id }}'

- name: 'slack'
  slack_configs:
  - api_url: <slack-webhook-url>
    channel: '#alerts'
    title: '{{ .CommonAnnotations.summary }}'
    text: |-
      {{ .CommonAnnotations.description }}
      Tenant: {{ .CommonLabels.innovabiz_tenant_id }}
      Region: {{ .CommonLabels.innovabiz_region_id }}
      Prioridade: {{ .CommonAnnotations.priority }}
      {{ if ne .CommonAnnotations.runbook_url "" }}Runbook: {{ .CommonAnnotations.runbook_url }}{{ end }}
      {{ if ne .CommonAnnotations.dashboard_url "" }}Dashboard: {{ .CommonAnnotations.dashboard_url }}{{ end }}

- name: 'slack-info'
  slack_configs:
  - api_url: <slack-webhook-url>
    channel: '#info-alerts'
    title: 'INFO: {{ .CommonAnnotations.summary }}'
    text: |-
      {{ .CommonAnnotations.description }}
      Tenant: {{ .CommonLabels.innovabiz_tenant_id }}
      Region: {{ .CommonLabels.innovabiz_region_id }}
      {{ if ne .CommonAnnotations.dashboard_url "" }}Dashboard: {{ .CommonAnnotations.dashboard_url }}{{ end }}
```

## Checklist de Validação

- [ ] Alertas configurados para todas as categorias requeridas
- [ ] Labels multi-dimensionais (tenant, região, módulo) aplicados
- [ ] Limiares apropriados definidos e baseados em dados reais
- [ ] Links para runbooks e dashboards incluídos
- [ ] Equipes responsáveis identificadas
- [ ] Severidades corretamente atribuídas
- [ ] Prioridades definidas para todos os alertas
- [ ] Descrições claras e acionáveis
- [ ] Integração com sistemas de notificação testada
- [ ] Alertas testados para verificar se disparam conforme esperado

## Recursos Adicionais

- [Documentação Prometheus Alerting](https://prometheus.io/docs/alerting/latest/overview/)
- [Portal de Observabilidade INNOVABIZ](https://observability.innovabiz.com)
- [Repositório de Alertas Padrão](https://github.com/innovabiz/observability-alerts)
- [Guia de Operação de Alertas](https://wiki.innovabiz.com/observability/alerts-ops)