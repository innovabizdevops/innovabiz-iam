# ADR-007: Sistema de Alertas Prometheus e Dashboards Grafana para IAM Audit Service

## Status

Aprovado

## Data

2025-07-31

## Contexto

O IAM Audit Service necessita de um sistema robusto e configurável de alertas, integrado com dashboards visuais, para detectar proativamente anomalias, violações de compliance e problemas operacionais. Os alertas e visualizações devem atender aos seguintes requisitos:

- Detectar eventos críticos e anomalias em tempo hábil
- Suportar contextos múltiplos (tenant, região, ambiente)
- Oferecer diferentes níveis de severidade e canais de notificação
- Integrar-se ao sistema de on-call e gestão de incidentes da plataforma
- Apresentar visualizações claras para diferentes stakeholders
- Atender requisitos regulatórios de auditoria e compliance
- Minimizar falsos positivos e fadiga de alertas
- Garantir extensibilidade para novos cenários de alerta

## Decisão

Implementar um **Sistema Integrado de Alertas Prometheus e Dashboards Grafana** com as seguintes características:

### 1. Alertas Prometheus

#### 1.1. Configuração de Alertas por Categoria

Organizar alertas em categorias específicas para o domínio de auditoria IAM:

```yaml
groups:
- name: iam-audit-service-alerts
  rules:
  
  # Alertas de Disponibilidade
  - alert: IAMAuditServiceDown
    expr: up{job="iam-audit-service"} == 0
    for: 1m
    labels:
      severity: critical
      service: iam-audit
      category: availability
    annotations:
      summary: "IAM Audit Service indisponível"
      description: "O serviço de auditoria IAM está indisponível há mais de 1 minuto."
      runbook_url: "https://wiki.innovabiz.com/iam/audit/alerts/service-down"

  # Alertas de Performance
  - alert: IAMAuditHighLatency
    expr: histogram_quantile(0.95, sum(rate(iam_audit_http_request_duration_seconds_bucket{job="iam-audit-service"}[5m])) by (tenant, region, le)) > 0.5
    for: 5m
    labels:
      severity: warning
      service: iam-audit
      category: performance
    annotations:
      summary: "Alta latência no IAM Audit Service"
      description: "Latência P95 > 500ms para {{ $labels.tenant }} na região {{ $labels.region }}."
      runbook_url: "https://wiki.innovabiz.com/iam/audit/alerts/high-latency"

  # Alertas de Eventos de Auditoria
  - alert: IAMAuditHighErrorRate
    expr: sum(rate(iam_audit_event_processed_total{job="iam-audit-service", status="error"}[5m])) by (tenant, region) / sum(rate(iam_audit_event_processed_total{job="iam-audit-service"}[5m])) by (tenant, region) > 0.05
    for: 5m
    labels:
      severity: warning
      service: iam-audit
      category: reliability
    annotations:
      summary: "Alta taxa de erros em eventos de auditoria"
      description: "Taxa de erros > 5% para eventos de auditoria no tenant {{ $labels.tenant }} na região {{ $labels.region }}."
      runbook_url: "https://wiki.innovabiz.com/iam/audit/alerts/high-error-rate"

  # Alertas de Compliance
  - alert: IAMAuditComplianceViolation
    expr: iam_audit_compliance_violation_total{job="iam-audit-service", severity="high"} > 0
    for: 1m
    labels:
      severity: critical
      service: iam-audit
      category: compliance
    annotations:
      summary: "Violação de compliance detectada"
      description: "Violação de compliance {{ $labels.compliance_type }} detectada para {{ $labels.tenant }} na região {{ $labels.region }}."
      runbook_url: "https://wiki.innovabiz.com/iam/audit/alerts/compliance-violation"

  # Alertas de Políticas de Retenção
  - alert: IAMAuditRetentionPolicyFailure
    expr: iam_audit_retention_execution_total{job="iam-audit-service", status="failure"} > 0
    for: 30m
    labels:
      severity: warning
      service: iam-audit
      category: data-governance
    annotations:
      summary: "Falha na execução de política de retenção"
      description: "Política de retenção {{ $labels.retention_policy }} falhou para {{ $labels.tenant }} na região {{ $labels.region }}."
      runbook_url: "https://wiki.innovabiz.com/iam/audit/alerts/retention-policy-failure"

  # Alertas de Recursos e Infraestrutura
  - alert: IAMAuditDatabaseHighConnections
    expr: iam_audit_resource_utilization_ratio{job="iam-audit-service", resource_type="database_connections"} > 0.8
    for: 10m
    labels:
      severity: warning
      service: iam-audit
      category: resources
    annotations:
      summary: "Alto uso de conexões de banco de dados"
      description: "Mais de 80% das conexões de banco de dados em uso para {{ $labels.tenant }} na região {{ $labels.region }}."
      runbook_url: "https://wiki.innovabiz.com/iam/audit/alerts/high-db-connections"
```

#### 1.2. Níveis de Severidade Padronizados

Adotar níveis de severidade consistentes com a plataforma INNOVABIZ:

| Severidade | Descrição | Tempo de Resposta | Canais |
|------------|-----------|-------------------|--------|
| critical | Impacto severo em produção, serviço indisponível | 15 minutos | Slack, Email, SMS, Telefone |
| warning | Potencial problema, degradação ou risco | 2 horas | Slack, Email |
| info | Informativo, sem impacto imediato | 8 horas | Slack |

#### 1.3. Canais de Notificação Configuráveis

```yaml
# Configuração do Alertmanager
global:
  resolve_timeout: 5m
  slack_api_url: '${SLACK_WEBHOOK_URL}'  # Configure via environment variable

templates:
- '/etc/alertmanager/templates/*.tmpl'

route:
  receiver: 'slack-notifications'
  group_by: ['tenant', 'region', 'alertname', 'severity']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h
  
  routes:
  - match:
      severity: critical
    receiver: 'pagerduty-critical'
    repeat_interval: 1h
    continue: true
    
  - match:
      severity: warning
    receiver: 'slack-warnings'
    repeat_interval: 2h
    
  - match:
      category: compliance
    receiver: 'compliance-team'
    group_by: ['tenant', 'region', 'compliance_type']
    
  - match:
      category: data-governance
    receiver: 'data-governance-team'
    
receivers:
- name: 'slack-notifications'
  slack_configs:
  - channel: '#iam-audit-alerts'
    send_resolved: true
    title: '[{{ .Status | toUpper }}{{ if eq .Status "firing" }}:{{ .Alerts.Firing | len }}{{ end }}] {{ .CommonLabels.alertname }}'
    text: >-
      {{ range .Alerts }}
        *Tenant:* {{ .Labels.tenant }}
        *Região:* {{ .Labels.region }}
        *Descrição:* {{ .Annotations.description }}
        *Severidade:* {{ .Labels.severity | toUpper }}
        *Runbook:* {{ .Annotations.runbook_url }}
        *Início:* {{ .StartsAt | since }}
      {{ end }}
    
- name: 'pagerduty-critical'
  pagerduty_configs:
  - service_key: '<service_key>'
    send_resolved: true
    description: '{{ .CommonLabels.alertname }}'
    client: 'IAM Audit Service'
    client_url: '{{ .CommonAnnotations.runbook_url }}'
    details:
      tenant: '{{ .CommonLabels.tenant }}'
      region: '{{ .CommonLabels.region }}'
      description: '{{ .CommonAnnotations.description }}'
      
- name: 'compliance-team'
  email_configs:
  - to: 'compliance-team@innovabiz.com'
    send_resolved: true
    subject: '[{{ .Status | toUpper }}] Alerta de Compliance: {{ .CommonLabels.alertname }}'
    html: |
      <h1>Alerta de Compliance</h1>
      <p><strong>Alerta:</strong> {{ .CommonLabels.alertname }}</p>
      <p><strong>Tenant:</strong> {{ .CommonLabels.tenant }}</p>
      <p><strong>Região:</strong> {{ .CommonLabels.region }}</p>
      <p><strong>Descrição:</strong> {{ .CommonAnnotations.description }}</p>
      <p><a href="{{ .CommonAnnotations.runbook_url }}">Link para o Runbook</a></p>
```

### 2. Dashboards Grafana

#### 2.1. Dashboard Principal do IAM Audit Service

Dashboard principal com foco em métricas operacionais e de negócio:

```json
{
  "annotations": {...},
  "editable": true,
  "gnetId": null,
  "graphTooltip": 0,
  "id": 1234,
  "links": [],
  "panels": [
    {
      "title": "IAM Audit Service - Visão Geral",
      "type": "row",
      "panels": []
    },
    {
      "title": "Total de Eventos de Auditoria",
      "type": "stat",
      "datasource": "Prometheus",
      "targets": [
        {
          "expr": "sum(iam_audit_event_processed_total{tenant=\"$tenant\", region=\"$region\"})",
          "instant": true
        }
      ],
      "fieldConfig": {
        "defaults": {
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              { "color": "green", "value": null }
            ]
          },
          "unit": "short"
        }
      }
    },
    {
      "title": "Taxa de Eventos por Minuto",
      "type": "timeseries",
      "datasource": "Prometheus",
      "targets": [
        {
          "expr": "sum(rate(iam_audit_event_processed_total{tenant=\"$tenant\", region=\"$region\"}[5m])) by (event_type)",
          "legendFormat": "{{event_type}}"
        }
      ]
    },
    {
      "title": "Latência de Processamento de Eventos (P95)",
      "type": "gauge",
      "datasource": "Prometheus",
      "targets": [
        {
          "expr": "histogram_quantile(0.95, sum(rate(iam_audit_event_processing_seconds_bucket{tenant=\"$tenant\", region=\"$region\"}[5m])) by (le))",
          "instant": true
        }
      ],
      "fieldConfig": {
        "defaults": {
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              { "color": "green", "value": null },
              { "color": "yellow", "value": 0.1 },
              { "color": "red", "value": 0.5 }
            ]
          },
          "unit": "s"
        }
      }
    },
    {
      "title": "Verificações de Compliance",
      "type": "timeseries",
      "datasource": "Prometheus",
      "targets": [
        {
          "expr": "sum(rate(iam_audit_compliance_check_total{tenant=\"$tenant\", region=\"$region\"}[5m])) by (compliance_type, status)",
          "legendFormat": "{{compliance_type}} - {{status}}"
        }
      ]
    },
    {
      "title": "HTTP - Requests por Minuto",
      "type": "timeseries",
      "datasource": "Prometheus",
      "targets": [
        {
          "expr": "sum(rate(iam_audit_http_request_total{tenant=\"$tenant\", region=\"$region\"}[5m])) by (status_code)",
          "legendFormat": "Status {{status_code}}"
        }
      ]
    },
    {
      "title": "HTTP - Latência por Endpoint (P95)",
      "type": "heatmap",
      "datasource": "Prometheus",
      "targets": [
        {
          "expr": "histogram_quantile(0.95, sum(rate(iam_audit_http_request_duration_seconds_bucket{tenant=\"$tenant\", region=\"$region\"}[5m])) by (path, le))",
          "legendFormat": "{{path}}"
        }
      ]
    },
    {
      "title": "Políticas de Retenção",
      "type": "row",
      "panels": []
    },
    {
      "title": "Execuções de Políticas de Retenção",
      "type": "timeseries",
      "datasource": "Prometheus",
      "targets": [
        {
          "expr": "sum(rate(iam_audit_retention_execution_total{tenant=\"$tenant\", region=\"$region\"}[5m])) by (retention_policy, status)",
          "legendFormat": "{{retention_policy}} - {{status}}"
        }
      ]
    },
    {
      "title": "Registros Expurgados por Política",
      "type": "barchart",
      "datasource": "Prometheus",
      "targets": [
        {
          "expr": "sum(iam_audit_retention_purge_total{tenant=\"$tenant\", region=\"$region\"}) by (retention_policy)",
          "legendFormat": "{{retention_policy}}"
        }
      ]
    },
    {
      "title": "Volume de Armazenamento Utilizado",
      "type": "gauge",
      "datasource": "Prometheus",
      "targets": [
        {
          "expr": "sum(iam_audit_storage_used_bytes{tenant=\"$tenant\", region=\"$region\"}) by (tenant)",
          "instant": true
        }
      ],
      "fieldConfig": {
        "defaults": {
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              { "color": "green", "value": null },
              { "color": "yellow", "value": 85 },
              { "color": "red", "value": 95 }
            ]
          },
          "unit": "bytes"
        }
      }
    }
  ],
  "schemaVersion": 27,
  "templating": {
    "list": [
      {
        "name": "tenant",
        "type": "query",
        "datasource": "Prometheus",
        "query": "label_values(iam_audit_event_processed_total, tenant)",
        "current": {
          "selected": true,
          "text": "all",
          "value": ["$__all"]
        },
        "includeAll": true
      },
      {
        "name": "region",
        "type": "query",
        "datasource": "Prometheus",
        "query": "label_values(iam_audit_event_processed_total{tenant=\"$tenant\"}, region)",
        "current": {
          "selected": true,
          "text": "all",
          "value": ["$__all"]
        },
        "includeAll": true
      },
      {
        "name": "environment",
        "type": "custom",
        "query": "production,staging,development",
        "current": {
          "selected": true,
          "text": "production",
          "value": "production"
        }
      }
    ]
  },
  "time": {
    "from": "now-6h",
    "to": "now"
  },
  "title": "IAM Audit Service Dashboard",
  "uid": "iam-audit-main"
}
```

#### 2.2. Dashboard de Compliance e Retenção

Dashboard especializado para métricas de compliance e políticas de retenção:

```json
{
  "annotations": {...},
  "editable": true,
  "gnetId": null,
  "graphTooltip": 0,
  "id": 1235,
  "panels": [
    {
      "title": "Compliance e Retenção",
      "type": "row",
      "panels": []
    },
    {
      "title": "Violações de Compliance por Tipo",
      "type": "piechart",
      "datasource": "Prometheus",
      "targets": [
        {
          "expr": "sum(iam_audit_compliance_violation_total{tenant=\"$tenant\", region=\"$region\"}) by (compliance_type)",
          "legendFormat": "{{compliance_type}}"
        }
      ]
    },
    {
      "title": "Tendência de Violações de Compliance",
      "type": "timeseries",
      "datasource": "Prometheus",
      "targets": [
        {
          "expr": "sum(rate(iam_audit_compliance_violation_total{tenant=\"$tenant\", region=\"$region\"}[1h])) by (compliance_type, severity)",
          "legendFormat": "{{compliance_type}} - {{severity}}"
        }
      ]
    },
    {
      "title": "Tempo de Execução das Verificações de Compliance",
      "type": "timeseries",
      "datasource": "Prometheus",
      "targets": [
        {
          "expr": "histogram_quantile(0.95, sum(rate(iam_audit_compliance_check_seconds_bucket{tenant=\"$tenant\", region=\"$region\"}[5m])) by (compliance_type, le))",
          "legendFormat": "{{compliance_type}} - P95"
        }
      ]
    },
    {
      "title": "Políticas de Retenção e Expurgo",
      "type": "row",
      "panels": []
    },
    {
      "title": "Eventos Expurgados por Política",
      "type": "timeseries",
      "datasource": "Prometheus",
      "targets": [
        {
          "expr": "sum(rate(iam_audit_retention_purge_total{tenant=\"$tenant\", region=\"$region\"}[1d])) by (retention_policy)",
          "legendFormat": "{{retention_policy}}"
        }
      ]
    },
    {
      "title": "Duração das Operações de Retenção",
      "type": "timeseries",
      "datasource": "Prometheus",
      "targets": [
        {
          "expr": "histogram_quantile(0.95, sum(rate(iam_audit_retention_execution_seconds_bucket{tenant=\"$tenant\", region=\"$region\"}[1d])) by (retention_policy, le))",
          "legendFormat": "{{retention_policy}} - P95"
        }
      ]
    },
    {
      "title": "Falhas de Retenção",
      "type": "stat",
      "datasource": "Prometheus",
      "targets": [
        {
          "expr": "sum(iam_audit_retention_execution_total{tenant=\"$tenant\", region=\"$region\", status=\"failure\"}) by (retention_policy)",
          "legendFormat": "{{retention_policy}}"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              { "color": "green", "value": null },
              { "color": "red", "value": 1 }
            ]
          }
        }
      }
    }
  ],
  "templating": {
    "list": [
      {
        "name": "tenant",
        "type": "query",
        "datasource": "Prometheus",
        "query": "label_values(iam_audit_event_processed_total, tenant)",
        "current": {
          "selected": true,
          "text": "all",
          "value": ["$__all"]
        },
        "includeAll": true
      },
      {
        "name": "region",
        "type": "query",
        "datasource": "Prometheus",
        "query": "label_values(iam_audit_event_processed_total{tenant=\"$tenant\"}, region)",
        "current": {
          "selected": true,
          "text": "all",
          "value": ["$__all"]
        },
        "includeAll": true
      }
    ]
  },
  "time": {
    "from": "now-7d",
    "to": "now"
  },
  "title": "IAM Audit Service - Compliance e Retenção",
  "uid": "iam-audit-compliance"
}
```

### 3. Integração com ObservabilityIntegration

Integração com a classe principal de observabilidade:

```python
class ObservabilityIntegration:
    def __init__(self, config: ObservabilityConfig = None):
        self.config = config or ObservabilityConfig()
        self.registry = CollectorRegistry()
        self.metrics = self._setup_metrics()
        self.router = APIRouter(tags=["Observability"])
        self._setup_endpoints()
    
    def _setup_metrics(self):
        # Configuração de métricas já definida no ADR-004
        # ...
        
    def _setup_endpoints(self):
        # Configuração de endpoints já definida no ADR-006
        # ...
        
    def instrument_app(self, app: FastAPI):
        """
        Instrumenta uma aplicação FastAPI com observabilidade completa.
        Adiciona middlewares, endpoints e configura alertas.
        """
        # Adiciona middlewares
        app.add_middleware(HTTPMetricsMiddleware, exclude_paths=["/metrics", "/health"])
        app.add_middleware(ContextMiddleware)
        
        # Registra endpoints
        app.include_router(self.router)
        
        # Configura alertas básicos
        self._setup_default_alerts()
        
        # Registra handlers para eventos do ciclo de vida
        @app.on_event("startup")
        async def startup_event():
            pass  # Inicialização de métricas

        @app.on_event("shutdown")
        async def shutdown_event():
            pass  # Limpeza de recursos
    
    def _setup_default_alerts(self):
        """
        Configura alertas default baseados nas métricas disponíveis.
        Estes alertas podem ser ajustados via arquivo de configuração externo.
        """
        # Implementação de configuração de alertas programáticos
        pass
```

## Alternativas Consideradas

### 1. Monitoramento e Alertas via Solução SaaS (Datadog, NewRelic)

**Prós:**
- Interface unificada já pronta
- Menor esforço de configuração inicial
- Funcionalidades avançadas de ML/AI para detecção de anomalias

**Contras:**
- Custos operacionais contínuos e potencialmente altos
- Dependência de fornecedor externo
- Potenciais limitações para customização
- Desafios para atender requisitos específicos de compliance

### 2. Alertas e Dashboards Desenvolvidos Internamente

**Prós:**
- Controle total sobre a implementação
- Customização completa para necessidades específicas
- Independência de soluções externas

**Contras:**
- Alto custo de desenvolvimento e manutenção
- Reinvenção de funcionalidades existentes
- Necessidade de equipe especializada
- Tempo de desenvolvimento mais longo

### 3. Monitoramento Baseado Apenas em Logs

**Prós:**
- Simplicidade de implementação
- Menor overhead no serviço
- Foco apenas em eventos relevantes

**Contras:**
- Menor visibilidade de tendências e padrões
- Dificuldade para análises quantitativas
- Tempo de detecção potencialmente maior
- Limitações para métricas de performance

## Consequências

### Positivas

- **Detecção proativa**: Identificação rápida de problemas antes de afetarem os usuários
- **Visibilidade completa**: Dashboards intuitivos para diferentes stakeholders
- **Flexibilidade**: Configuração adaptável às necessidades específicas de cada tenant/região
- **Padronização**: Alinhamento com os padrões da plataforma INNOVABIZ
- **Auditabilidade**: Evidências para processos de compliance e auditoria
- **Operação eficiente**: Redução do tempo de resolução de problemas

### Negativas

- **Complexidade de configuração**: Necessidade de manutenção de regras e alertas
- **Potencial fadiga de alertas**: Risco de muitos alertas causarem dessensibilização
- **Custo de infraestrutura**: Necessidade de recursos para Prometheus e Grafana
- **Curva de aprendizado**: Equipe precisa familiarizar-se com PromQL e configurações

### Mitigação de Riscos

- Implementar thresholds graduais para minimizar falsos positivos
- Configurar agrupamento inteligente de alertas para reduzir fadiga
- Estabelecer períodos de silêncio para manutenções planejadas
- Documentar runbooks claros para cada tipo de alerta
- Revisar periodicamente a eficácia dos alertas
- Implementar rotação de dados para controle de custos de armazenamento
- Realizar testes regulares de failover e recuperação

## Conformidade com Padrões

- **SRE Best Practices**: Google SRE Handbook para alerting
- **PagerDuty Incident Response**: Framework para gestão de incidentes
- **ISO/IEC 20000**: Gestão de serviços de TI
- **PCI DSS 4.0**: Requisitos de monitoramento (10.2, 10.6, 11.4, 12.10)
- **GDPR/LGPD**: Detecção de violações de dados
- **INNOVABIZ Platform Observability Standards v2.5**

## Implementação

A implementação inclui:

1. **Configuração de Prometheus**:
   - Arquivos de configuração para regras de alertas
   - Integração com Alertmanager
   - Definição de políticas de retenção de métricas

2. **Configuração de Alertmanager**:
   - Rotas para diferentes tipos de alertas
   - Integrações com canais de notificação
   - Templates de mensagens de alerta

3. **Dashboards Grafana**:
   - Dashboard principal de operações
   - Dashboard específico para compliance e retenção
   - Dashboard para troubleshooting e diagnóstico

4. **Documentação**:
   - Runbooks para resposta a alertas
   - Guia de interpretação de dashboards
   - Procedimentos para ajuste de thresholds

## Exemplos de Alertas Críticos

### Alerta de Segurança: Tentativa de Acesso Não Autorizado

```yaml
- alert: IAMAuditUnauthorizedAccessAttempt
  expr: sum(rate(iam_audit_event_processed_total{event_type="unauthorized_access", severity="high"}[5m])) by (tenant, region) > 0
  for: 1m
  labels:
    severity: critical
    service: iam-audit
    category: security
  annotations:
    summary: "Tentativa de acesso não autorizado detectada"
    description: "Múltiplas tentativas de acesso não autorizado detectadas para {{ $labels.tenant }} na região {{ $labels.region }}."
    runbook_url: "https://wiki.innovabiz.com/iam/audit/alerts/unauthorized-access"
```

### Alerta de Compliance: Violação de Política de Retenção

```yaml
- alert: IAMAuditRetentionPolicyViolation
  expr: iam_audit_compliance_violation_total{compliance_type=~"retention_policy|data_retention"} > 0
  for: 5m
  labels:
    severity: critical
    service: iam-audit
    category: compliance
  annotations:
    summary: "Violação de política de retenção detectada"
    description: "Violação de política de retenção de dados detectada para {{ $labels.tenant }} na região {{ $labels.region }}. Tipo: {{ $labels.compliance_type }}"
    runbook_url: "https://wiki.innovabiz.com/iam/audit/alerts/retention-violation"
```

## Referências

1. Prometheus Alerting Rules - https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/
2. Alertmanager Configuration - https://prometheus.io/docs/alerting/latest/configuration/
3. Grafana Dashboard Best Practices - https://grafana.com/docs/grafana/latest/best-practices/
4. Google SRE Book: Practical Alerting - https://sre.google/sre-book/practical-alerting/
5. PagerDuty Incident Response Documentation - https://response.pagerduty.com/
6. INNOVABIZ Platform Observability Standards v2.5 (Internal Document)
7. Multi-tenant Alerting Best Practices - https://www.robustperception.io/multi-tenant-prometheus-and-alertmanager