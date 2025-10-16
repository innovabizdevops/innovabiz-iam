# WebAuthn Monitoring & Operations Guide

**Documento:** Guia de Monitoramento e Operações WebAuthn/FIDO2  
**Versão:** 1.0.0  
**Data:** 31/07/2025  
**Autor:** Equipe SRE INNOVABIZ  
**Classificação:** Confidencial - Operacional  

## Índice

1. [Visão Geral](#1-visão-geral)
2. [Métricas e KPIs](#2-métricas-e-kpis)
3. [Dashboards](#3-dashboards)
4. [Alertas](#4-alertas)
5. [Logs e Auditoria](#5-logs-e-auditoria)
6. [Procedimentos Operacionais](#6-procedimentos-operacionais)
7. [Incident Response](#7-incident-response)
8. [Capacity Planning](#8-capacity-planning)

## 1. Visão Geral

### 1.1 Stack de Monitoramento

| Componente | Função | Endpoint |
|------------|--------|----------|
| **Prometheus** | Coleta de métricas | https://prometheus.innovabiz.com |
| **Grafana** | Visualização | https://grafana.innovabiz.com |
| **AlertManager** | Gestão de alertas | https://alerts.innovabiz.com |
| **Jaeger** | Distributed tracing | https://jaeger.innovabiz.com |
| **ELK Stack** | Logs centralizados | https://kibana.innovabiz.com |

### 1.2 SLOs (Service Level Objectives)

| Métrica | SLO | Período | Consequência |
|---------|-----|---------|--------------|
| **Disponibilidade** | 99.9% | Mensal | Error budget: 43.2 min |
| **Latência P95** | <500ms | Semanal | Performance review |
| **Taxa de Erro** | <0.1% | Diário | Incident response |
| **MTTR** | <15min | Por incident | Escalation |

## 2. Métricas e KPIs

### 2.1 Métricas de Negócio

#### Autenticação
```promql
# Taxa de sucesso de autenticação
rate(webauthn_authentication_total{status="success"}[5m]) / 
rate(webauthn_authentication_total[5m]) * 100

# Tempo médio de autenticação
histogram_quantile(0.5, rate(webauthn_authentication_duration_seconds_bucket[5m]))

# Autenticações por tenant
sum(rate(webauthn_authentication_total[5m])) by (tenant_id)
```

#### Registro
```promql
# Taxa de sucesso de registro
rate(webauthn_registration_total{status="success"}[5m]) / 
rate(webauthn_registration_total[5m]) * 100

# Registros por tipo de autenticador
sum(rate(webauthn_registration_total[5m])) by (authenticator_type)

# Tempo médio de registro
histogram_quantile(0.95, rate(webauthn_registration_duration_seconds_bucket[5m]))
```

### 2.2 Métricas Técnicas

#### Performance
```promql
# CPU utilization
rate(container_cpu_usage_seconds_total{pod=~"webauthn-.*"}[5m]) * 100

# Memory utilization
container_memory_usage_bytes{pod=~"webauthn-.*"} / 
container_spec_memory_limit_bytes * 100

# Request rate
rate(webauthn_http_requests_total[5m])

# Error rate
rate(webauthn_http_requests_total{status=~"5.."}[5m]) / 
rate(webauthn_http_requests_total[5m]) * 100
```

#### Infraestrutura
```promql
# Database connections
webauthn_database_connections_active

# Redis hit rate
rate(webauthn_redis_hits_total[5m]) / 
(rate(webauthn_redis_hits_total[5m]) + rate(webauthn_redis_misses_total[5m])) * 100

# Kafka lag
webauthn_kafka_consumer_lag_sum
```

### 2.3 Métricas de Segurança

```promql
# Tentativas de autenticação suspeitas
rate(webauthn_security_suspicious_attempts_total[5m])

# Rate limiting ativado
rate(webauthn_rate_limit_exceeded_total[5m])

# Falhas de verificação de attestation
rate(webauthn_attestation_verification_total{status="failed"}[5m])
```

## 3. Dashboards

### 3.1 Dashboard Executivo

```json
{
  "dashboard": {
    "title": "WebAuthn - Executive Overview",
    "panels": [
      {
        "title": "Service Health",
        "type": "stat",
        "targets": [
          {
            "expr": "up{job=\"webauthn\"}"
          }
        ]
      },
      {
        "title": "Daily Active Users",
        "type": "graph",
        "targets": [
          {
            "expr": "count(count by (user_id) (increase(webauthn_authentication_total[1d])))"
          }
        ]
      },
      {
        "title": "Success Rate",
        "type": "gauge",
        "targets": [
          {
            "expr": "rate(webauthn_authentication_total{status=\"success\"}[5m]) / rate(webauthn_authentication_total[5m]) * 100"
          }
        ]
      }
    ]
  }
}
```

### 3.2 Dashboard Operacional

```json
{
  "dashboard": {
    "title": "WebAuthn - Operations",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(webauthn_http_requests_total[5m])"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(webauthn_request_duration_seconds_bucket[5m]))"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(webauthn_http_requests_total{status=~\"5..\"}[5m])"
          }
        ]
      }
    ]
  }
}
```

### 3.3 Dashboard de Segurança

```json
{
  "dashboard": {
    "title": "WebAuthn - Security",
    "panels": [
      {
        "title": "Authentication Attempts",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(webauthn_authentication_total[5m])"
          }
        ]
      },
      {
        "title": "Suspicious Activity",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(webauthn_security_suspicious_attempts_total[5m])"
          }
        ]
      },
      {
        "title": "Rate Limiting",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(webauthn_rate_limit_exceeded_total[5m])"
          }
        ]
      }
    ]
  }
}
```

## 4. Alertas

### 4.1 Alertas Críticos

```yaml
groups:
  - name: webauthn.critical
    rules:
      - alert: WebAuthnServiceDown
        expr: up{job="webauthn"} == 0
        for: 1m
        labels:
          severity: critical
          service: webauthn
        annotations:
          summary: "WebAuthn service is down"
          description: "WebAuthn service has been down for more than 1 minute"
          runbook_url: "https://runbooks.innovabiz.com/webauthn/service-down"
          
      - alert: WebAuthnHighErrorRate
        expr: rate(webauthn_http_requests_total{status=~"5.."}[5m]) / rate(webauthn_http_requests_total[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
          service: webauthn
        annotations:
          summary: "High error rate in WebAuthn service"
          description: "Error rate is {{ $value | humanizePercentage }}"
          
      - alert: WebAuthnDatabaseDown
        expr: webauthn_database_up == 0
        for: 2m
        labels:
          severity: critical
          service: webauthn
        annotations:
          summary: "WebAuthn database is down"
          description: "Database connection has been down for more than 2 minutes"
```

### 4.2 Alertas de Warning

```yaml
groups:
  - name: webauthn.warning
    rules:
      - alert: WebAuthnHighLatency
        expr: histogram_quantile(0.95, rate(webauthn_request_duration_seconds_bucket[5m])) > 1
        for: 10m
        labels:
          severity: warning
          service: webauthn
        annotations:
          summary: "High latency in WebAuthn service"
          description: "95th percentile latency is {{ $value }}s"
          
      - alert: WebAuthnHighMemoryUsage
        expr: container_memory_usage_bytes{pod=~"webauthn-.*"} / container_spec_memory_limit_bytes > 0.8
        for: 15m
        labels:
          severity: warning
          service: webauthn
        annotations:
          summary: "High memory usage in WebAuthn pods"
          description: "Memory usage is {{ $value | humanizePercentage }}"
          
      - alert: WebAuthnSuspiciousActivity
        expr: rate(webauthn_security_suspicious_attempts_total[5m]) > 10
        for: 5m
        labels:
          severity: warning
          service: webauthn
        annotations:
          summary: "Suspicious authentication activity detected"
          description: "{{ $value }} suspicious attempts per second"
```

### 4.3 Configuração AlertManager

```yaml
global:
  smtp_smarthost: 'smtp.innovabiz.com:587'
  smtp_from: 'alerts@innovabiz.com'

route:
  group_by: ['alertname', 'service']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 12h
  receiver: 'web.hook'
  routes:
  - match:
      severity: critical
    receiver: 'critical-alerts'
  - match:
      severity: warning
    receiver: 'warning-alerts'

receivers:
- name: 'critical-alerts'
  slack_configs:
  - api_url: 'https://hooks.slack.com/services/CRITICAL_WEBHOOK'
    channel: '#alerts-critical'
    title: 'CRITICAL: {{ .GroupLabels.alertname }}'
    text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
  email_configs:
  - to: 'sre-oncall@innovabiz.com'
    subject: 'CRITICAL: WebAuthn Alert'
    
- name: 'warning-alerts'
  slack_configs:
  - api_url: 'https://hooks.slack.com/services/WARNING_WEBHOOK'
    channel: '#alerts-warning'
    title: 'WARNING: {{ .GroupLabels.alertname }}'
```

## 5. Logs e Auditoria

### 5.1 Estrutura de Logs

```json
{
  "timestamp": "2025-01-31T10:00:00.000Z",
  "level": "info",
  "service": "webauthn",
  "version": "1.0.0",
  "environment": "production",
  "tenant_id": "tenant-123",
  "user_id": "user-456",
  "correlation_id": "corr-789",
  "event_type": "webauthn_authentication_success",
  "message": "User authenticated successfully",
  "metadata": {
    "authenticator_type": "platform",
    "user_verification": true,
    "duration_ms": 1250,
    "ip_address": "192.168.1.100",
    "user_agent": "Mozilla/5.0..."
  }
}
```

### 5.2 Queries Úteis (Kibana)

```lucene
# Erros de autenticação
service:webauthn AND level:error AND event_type:webauthn_authentication_failed

# Atividade suspeita
service:webauthn AND event_type:webauthn_suspicious_activity

# Performance por tenant
service:webauthn AND event_type:webauthn_authentication_success AND tenant_id:"tenant-123"

# Erros de attestation
service:webauthn AND event_type:webauthn_attestation_verification_failed
```

### 5.3 Retenção de Logs

| Tipo de Log | Retenção | Localização |
|-------------|----------|-------------|
| **Application Logs** | 90 dias | Elasticsearch |
| **Audit Logs** | 7 anos | S3 + Glacier |
| **Security Logs** | 2 anos | S3 |
| **Performance Logs** | 30 dias | Elasticsearch |

## 6. Procedimentos Operacionais

### 6.1 Health Checks Diários

```bash
#!/bin/bash
# daily-health-check.sh

echo "=== WebAuthn Daily Health Check ==="
echo "Date: $(date)"

# Check service status
kubectl get pods -n innovabiz-production -l app.kubernetes.io/name=webauthn

# Check resource usage
kubectl top pods -n innovabiz-production -l app.kubernetes.io/name=webauthn

# Check certificate expiry
kubectl get certificates -n innovabiz-production

# Check database connections
kubectl exec -n innovabiz-production deployment/webauthn -- \
  curl -s http://localhost:3000/health | jq '.database.status'

# Check Redis connectivity
kubectl exec -n innovabiz-production deployment/webauthn -- \
  curl -s http://localhost:3000/health | jq '.redis.status'

# Check recent errors
kubectl logs -n innovabiz-production -l app.kubernetes.io/name=webauthn \
  --since=24h | grep -i error | wc -l

echo "=== Health Check Complete ==="
```

### 6.2 Performance Review Semanal

```bash
#!/bin/bash
# weekly-performance-review.sh

echo "=== WebAuthn Weekly Performance Review ==="

# Calculate SLO compliance
AVAILABILITY=$(curl -s "http://prometheus:9090/api/v1/query?query=avg_over_time(up{job=\"webauthn\"}[7d])" | jq -r '.data.result[0].value[1]')
echo "Availability: $(echo "$AVAILABILITY * 100" | bc)%"

# Calculate average response time
AVG_LATENCY=$(curl -s "http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95, avg_over_time(webauthn_request_duration_seconds_bucket[7d]))" | jq -r '.data.result[0].value[1]')
echo "P95 Latency: ${AVG_LATENCY}s"

# Calculate error rate
ERROR_RATE=$(curl -s "http://prometheus:9090/api/v1/query?query=rate(webauthn_http_requests_total{status=~\"5..\"}[7d]) / rate(webauthn_http_requests_total[7d])" | jq -r '.data.result[0].value[1]')
echo "Error Rate: $(echo "$ERROR_RATE * 100" | bc)%"

# Generate report
cat > weekly-report.md << EOF
# WebAuthn Weekly Performance Report

**Week of:** $(date -d '7 days ago' +%Y-%m-%d) to $(date +%Y-%m-%d)

## SLO Compliance
- **Availability:** ${AVAILABILITY}% (Target: 99.9%)
- **P95 Latency:** ${AVG_LATENCY}s (Target: <0.5s)
- **Error Rate:** ${ERROR_RATE}% (Target: <0.1%)

## Recommendations
$(if (( $(echo "$AVAILABILITY < 0.999" | bc -l) )); then echo "- Investigate availability issues"; fi)
$(if (( $(echo "$AVG_LATENCY > 0.5" | bc -l) )); then echo "- Optimize response times"; fi)
$(if (( $(echo "$ERROR_RATE > 0.001" | bc -l) )); then echo "- Reduce error rate"; fi)
EOF

echo "Report generated: weekly-report.md"
```

### 6.3 Capacity Planning Mensal

```bash
#!/bin/bash
# monthly-capacity-planning.sh

echo "=== WebAuthn Monthly Capacity Planning ==="

# Get current resource usage
CPU_USAGE=$(kubectl top pods -n innovabiz-production -l app.kubernetes.io/name=webauthn --no-headers | awk '{sum+=$2} END {print sum}')
MEMORY_USAGE=$(kubectl top pods -n innovabiz-production -l app.kubernetes.io/name=webauthn --no-headers | awk '{sum+=$3} END {print sum}')

# Get request volume trend
CURRENT_RPS=$(curl -s "http://prometheus:9090/api/v1/query?query=rate(webauthn_http_requests_total[1h])" | jq -r '.data.result[0].value[1]')
LAST_MONTH_RPS=$(curl -s "http://prometheus:9090/api/v1/query?query=rate(webauthn_http_requests_total[30d])" | jq -r '.data.result[0].value[1]')

# Calculate growth rate
GROWTH_RATE=$(echo "scale=2; ($CURRENT_RPS - $LAST_MONTH_RPS) / $LAST_MONTH_RPS * 100" | bc)

echo "Current CPU Usage: ${CPU_USAGE}m"
echo "Current Memory Usage: ${MEMORY_USAGE}Mi"
echo "Current RPS: $CURRENT_RPS"
echo "Growth Rate: ${GROWTH_RATE}%"

# Recommendations
if (( $(echo "$GROWTH_RATE > 20" | bc -l) )); then
    echo "RECOMMENDATION: Consider scaling up resources"
fi

if (( $(echo "$CPU_USAGE > 1500" | bc -l) )); then
    echo "RECOMMENDATION: CPU usage is high, consider horizontal scaling"
fi
```

## 7. Incident Response

### 7.1 Playbook: Service Down

```markdown
# Incident Response: WebAuthn Service Down

## Immediate Actions (0-5 minutes)
1. Acknowledge alert in PagerDuty
2. Check service status: `kubectl get pods -n innovabiz-production -l app.kubernetes.io/name=webauthn`
3. Check recent deployments: `helm history webauthn-production -n innovabiz-production`
4. Notify stakeholders in #incident-response

## Investigation (5-15 minutes)
1. Check pod logs: `kubectl logs -n innovabiz-production -l app.kubernetes.io/name=webauthn --tail=100`
2. Check events: `kubectl get events -n innovabiz-production --sort-by='.lastTimestamp'`
3. Check resource usage: `kubectl top pods -n innovabiz-production`
4. Check dependencies (DB, Redis, Kafka)

## Mitigation (15-30 minutes)
1. If recent deployment: `helm rollback webauthn-production -n innovabiz-production`
2. If resource issue: Scale up replicas
3. If dependency issue: Fix dependency or failover
4. Verify service recovery

## Post-Incident
1. Update incident timeline
2. Schedule post-mortem meeting
3. Document lessons learned
```

### 7.2 Playbook: High Error Rate

```markdown
# Incident Response: High Error Rate

## Immediate Actions
1. Check error distribution by endpoint
2. Check recent changes or deployments
3. Verify database and Redis connectivity
4. Check for rate limiting issues

## Investigation
1. Analyze error logs for patterns
2. Check external dependencies
3. Review recent configuration changes
4. Verify certificate validity

## Mitigation
1. Rollback if caused by deployment
2. Scale resources if capacity issue
3. Fix configuration if config issue
4. Implement circuit breaker if dependency issue
```

## 8. Capacity Planning

### 8.1 Métricas de Capacidade

| Métrica | Atual | Limite | Ação |
|---------|-------|--------|------|
| **CPU Usage** | 45% | 70% | Monitor |
| **Memory Usage** | 60% | 80% | Monitor |
| **Request Rate** | 500 req/s | 1000 req/s | Monitor |
| **Database Connections** | 25 | 50 | Monitor |

### 8.2 Projeções de Crescimento

```python
# growth-projection.py
import pandas as pd
import numpy as np
from datetime import datetime, timedelta

# Historical data (example)
dates = pd.date_range(start='2024-01-01', end='2025-01-31', freq='D')
requests = np.random.normal(500, 50, len(dates)) * (1 + 0.001 * np.arange(len(dates)))

# Linear regression for growth prediction
from sklearn.linear_model import LinearRegression
X = np.arange(len(dates)).reshape(-1, 1)
model = LinearRegression().fit(X, requests)

# Predict next 6 months
future_days = 180
future_X = np.arange(len(dates), len(dates) + future_days).reshape(-1, 1)
future_requests = model.predict(future_X)

print(f"Current average: {requests[-30:].mean():.0f} req/s")
print(f"Projected in 6 months: {future_requests[-1]:.0f} req/s")
print(f"Growth rate: {((future_requests[-1] / requests[-1]) - 1) * 100:.1f}%")
```

### 8.3 Scaling Recommendations

| Cenário | CPU | Memory | Replicas | Database |
|---------|-----|--------|----------|----------|
| **Current** | 1 core | 1GB | 5 | 2 cores |
| **+50% traffic** | 1.5 cores | 1.5GB | 7 | 3 cores |
| **+100% traffic** | 2 cores | 2GB | 10 | 4 cores |
| **+200% traffic** | 3 cores | 3GB | 15 | 6 cores |

---

**Desenvolvido pela equipe INNOVABIZ**  
**© 2025 INNOVABIZ. Todos os direitos reservados.**