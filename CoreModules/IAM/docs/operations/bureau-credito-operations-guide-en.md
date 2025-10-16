# Operations Guide - Credit Bureau Module

![INNOVABIZ](../../../assets/images/logo.png)

**Version:** 1.0.0  
**Date:** 2025-08-06  
**Classification:** Private  
**Author:** INNOVABIZ Operations Team

## Table of Contents

1. [Introduction](#1-introduction)
2. [Operational Overview](#2-operational-overview)
3. [Infrastructure and Deployment](#3-infrastructure-and-deployment)
4. [Monitoring and Observability](#4-monitoring-and-observability)
5. [Troubleshooting](#5-troubleshooting)
6. [Maintenance Procedures](#6-maintenance-procedures)
7. [Backup and Recovery Procedures](#7-backup-and-recovery-procedures)
8. [Operational Security Policies](#8-operational-security-policies)
9. [Market-Specific Compliance](#9-market-specific-compliance)
10. [Configuration Management](#10-configuration-management)
11. [References](#11-references)

## 1. Introduction

This operations guide provides detailed instructions for administering, maintaining, monitoring, and troubleshooting the Credit Bureau module on the INNOVABIZ platform. The document is designed for operations teams, DevSecOps, and support staff who manage the production environment and ensure the availability, performance, security, and regulatory compliance of the module.

### 1.1 Module Objectives

The Credit Bureau is a critical component of the INNOVABIZ platform that:
- Provides credit inquiries at different depth levels
- Executes market-specific compliance validations
- Manages consents for financial data queries
- Integrates with other core modules (IAM, Payment Gateway, Risk Management, Mobile Money, Marketplace)
- Maintains audit records for all queries and actions

### 1.2 Operational Context

The module operates in an environment that is:
- **Multi-market**: Supports specific rules for Angola, Brazil, European Union, USA, and global market
- **Multi-tenant**: Allows configurations by tenant type and specific needs
- **Multi-context**: Adapts behavior based on market context and application
- **Highly observable**: Generates complete telemetry for auditing, compliance, and operations

### 1.3 Operational Requirements

| Requirement | Description | Target |
|-------------|-------------|--------|
| Availability | Percentage of time the service is accessible | 99.95% |
| MTTR | Mean time to recovery after failure | < 15 minutes |
| P95 Latency | Response time for 95% of queries | < 500ms |
| Error Rate | Percentage of queries with technical error | < 0.1% |
| RTO | Recovery Time Objective | < 1 hour |
| RPO | Recovery Point Objective | < 5 minutes |
| Alerts | Time for incident notification | < 1 minute |

## 2. Operational Overview

### 2.1 Operational Architecture

The Credit Bureau is implemented as a containerized service, deployed on a managed Kubernetes cluster. The operational architecture includes:

```
                                       ┌────────────────┐
                                       │   API Gateway  │
                                       │    KrakenD     │
                                       └───────┬────────┘
                                               │
                ┌──────────────────────────────┴───────────────────────────┐
                │                                                          │
        ┌───────┴───────┐                                        ┌────────┴──────────┐
        │  Credit       │                                        │                   │
        │  Bureau       │◄─────────────┐             ┌──────────►│  IAM Service      │
        │  Service      │              │             │           │                   │
        └───┬───────────┘              │             │           └───────────────────┘
            │                          │             │
            │                 ┌────────┴─────────┐   │
        ┌───┴───────────┐     │                  │   │        ┌─────────────────────┐
        │ Prometheus/   │     │  OpenTelemetry   │   │        │                     │
        │ Grafana       │     │  Collector       │◄──┴────────┤  Logs/Audit Storage │
        │               │     │                  │            │                     │
        └───────────────┘     └──────────────────┘            └─────────────────────┘
```

### 2.2 Operational Components

| Component | Description | Responsibility |
|-----------|-------------|------------------|
| Credit Bureau Service | Main module service | Credit inquiry processing, compliance rules application |
| API Gateway (KrakenD) | API Gateway | Secure API exposure, authentication, authorization and rate limiting |
| IAM Service | Identity service | Authentication, authorization and consent management |
| OpenTelemetry Collector | Telemetry collector | Centralized collection of traces, metrics and logs |
| Prometheus/Grafana | Monitoring stack | Metrics storage, visualization and alerts |
| Logs/Audit Storage | Log and audit storage | Centralized storage of logs and audit events |

### 2.3 Operational Flows

#### 2.3.1 Credit Inquiry Flow

1. Request is received by the KrakenD API Gateway
2. Gateway validates JWT token and access scopes
3. Request is forwarded to the Credit Bureau Service
4. Service validates authentication, authorization, consent and compliance
5. Daily query limits are applied
6. The inquiry is processed and recorded for auditing
7. Regulatory notifications are sent when necessary
8. Result is returned to the client

#### 2.3.2 Telemetry Flow

1. Each operation generates OpenTelemetry spans
2. Audit events are recorded for all inquiries
3. Security events are generated for invalid attempts
4. Business and operational metrics are collected
5. The OpenTelemetry Collector aggregates and forwards telemetry
6. Prometheus stores metrics for alerts and dashboards
7. Grafana presents operational and business dashboards

### 2.4 Support Models

| Level | Response Time | Coverage Hours | Responsible Team |
|-------|-------------------|----------------------|-------------------|
| L1 | 15 minutes | 24x7 | NOC / Operations |
| L2 | 1 hour | 24x7 | DevSecOps |
| L3 | 4 hours | Business hours | Development |

## 3. Infrastructure and Deployment

### 3.1 Infrastructure Requirements

| Resource | Minimum | Recommended | Notes |
|---------|--------|-------------|-------------|
| CPU | 2 cores | 4 cores | Per replica |
| Memory | 1 GB | 2 GB | Per replica |
| Storage | 10 GB | 20 GB | For local cache |
| Replicas | 3 | 5+ | Distributed across AZs |
| Network | 100 Mbps | 1 Gbps | Low latency |

### 3.2 Kubernetes Manifests

#### 3.2.1 Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: credit-bureau
  namespace: innovabiz
  labels:
    app: credit-bureau
    module: core
spec:
  replicas: 3
  selector:
    matchLabels:
      app: credit-bureau
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: credit-bureau
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
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
                  - credit-bureau
              topologyKey: kubernetes.io/hostname
      containers:
      - name: credit-bureau
        image: innovabiz/credit-bureau:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: ENVIRONMENT
          valueFrom:
            configMapKeyRef:
              name: credit-bureau-config
              key: ENVIRONMENT
        - name: MARKET
          valueFrom:
            configMapKeyRef:
              name: credit-bureau-config
              key: MARKET
        - name: TENANT_TYPE
          valueFrom:
            configMapKeyRef:
              name: credit-bureau-config
              key: TENANT_TYPE
        - name: SERVICE_VERSION
          valueFrom:
            configMapKeyRef:
              name: credit-bureau-config
              key: SERVICE_VERSION
        - name: LOG_LEVEL
          valueFrom:
            configMapKeyRef:
              name: credit-bureau-config
              key: LOG_LEVEL
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          valueFrom:
            configMapKeyRef:
              name: credit-bureau-config
              key: OTEL_EXPORTER_OTLP_ENDPOINT
        resources:
          limits:
            cpu: "1"
            memory: "1Gi"
          requests:
            cpu: "500m"
            memory: "512Mi"
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
        volumeMounts:
        - name: config-volume
          mountPath: /app/config
        - name: certs-volume
          mountPath: /app/certs
      volumes:
      - name: config-volume
        configMap:
          name: credit-bureau-config
      - name: certs-volume
        secret:
          secretName: credit-bureau-certs
```

#### 3.2.2 Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: credit-bureau
  namespace: innovabiz
  labels:
    app: credit-bureau
    module: core
spec:
  ports:
  - port: 8080
    targetPort: 8080
    protocol: TCP
    name: http
  selector:
    app: credit-bureau
  type: ClusterIP
```

#### 3.2.3 ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: credit-bureau-config
  namespace: innovabiz
data:
  ENVIRONMENT: "production"
  MARKET: "global"
  TENANT_TYPE: "default"
  SERVICE_VERSION: "1.0.0"
  LOG_LEVEL: "info"
  OTEL_EXPORTER_OTLP_ENDPOINT: "http://otel-collector.observability:4317"
  REDIS_HOST: "redis-master.cache"
  REDIS_PORT: "6379"
  DAILY_QUERY_LIMIT_DEFAULT: "100"
  DAILY_QUERY_LIMIT_ANGOLA: "50"
  DAILY_QUERY_LIMIT_BRAZIL: "200"
  DAILY_QUERY_LIMIT_EU: "100"
  DAILY_QUERY_LIMIT_USA: "150"
  NOTIFICATION_REQUIRED_ANGOLA: "true"
  NOTIFICATION_REQUIRED_BRAZIL: "false"
  NOTIFICATION_REQUIRED_EU: "true"
  NOTIFICATION_REQUIRED_USA: "false"
```

#### 3.2.4 HPA (Horizontal Pod Autoscaler)

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: credit-bureau-hpa
  namespace: innovabiz
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: credit-bureau
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Pods
        value: 1
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 15
      - type: Pods
        value: 4
        periodSeconds: 15
      selectPolicy: Max
```

### 3.3 Deployment Process

#### 3.3.1 Deployment Strategy

The Credit Bureau module uses Blue/Green deployment through ArgoCD with the following phases:

1. **Build & Test**: Compilation and automated testing (unit, integration, performance)
2. **Image Publication**: Publishing Docker image to the corporate registry
3. **Deploy to Staging**: Automated deployment to staging environment
4. **Acceptance Tests**: Automated acceptance tests
5. **Approval**: Manual or automated approval based on metrics
6. **Deploy to Production**: Deployment to production using Blue/Green
7. **Validation**: Integrity and performance validation
8. **Finalization**: Promoting the new version and removing the old one

#### 3.3.2 Deployment Commands

For manual deployment (only in emergencies):

```bash
# Apply ConfigMap
kubectl apply -f credit-bureau-configmap.yaml -n innovabiz

# Apply Deployment
kubectl apply -f credit-bureau-deployment.yaml -n innovabiz

# Monitor rollout
kubectl rollout status deployment/credit-bureau -n innovabiz

# In case of problems, rollback
kubectl rollout undo deployment/credit-bureau -n innovabiz
```

For deployment via ArgoCD:

```bash
# Synchronize application
argocd app sync credit-bureau

# Check synchronization status
argocd app get credit-bureau

# Promote version (Blue/Green)
argocd app actions run credit-bureau promote --kind Rollout
```## 4. Monitoring and Observability

### 4.1 Metrics Strategy

The Credit Bureau module follows the RED method (Rate, Errors, Duration) for service monitoring, complemented by business metrics and compliance indicators:

#### 4.1.1 Core Metrics

| Metric Type | Metric Name | Description | Alert Threshold |
|------------|-------------|-------------|-----------------|
| Rate | `credit_bureau_queries_total` | Total number of credit queries | N/A |
| Errors | `credit_bureau_queries_error_total` | Failed credit queries | >1% error rate |
| Duration | `credit_bureau_query_duration_seconds` | Query execution time | P95 > 1s |

#### 4.1.2 Business Metrics

| Metric Name | Description | Alert Threshold |
|-------------|-------------|-----------------|
| `credit_bureau_queries_by_market` | Query count by market | N/A |
| `credit_bureau_queries_by_type` | Query count by type (basic, complete, score) | N/A |
| `credit_bureau_daily_quota_used_percent` | Percentage of daily quota used | >85% |
| `credit_bureau_consent_rejection_rate` | Rate of consent rejections | >10% |
| `credit_bureau_avg_score` | Average credit score by market | Significant deviation |

#### 4.1.3 Operational Metrics

| Metric Name | Description | Alert Threshold |
|-------------|-------------|-----------------|
| `credit_bureau_cache_hit_rate` | Cache hit rate | <70% |
| `credit_bureau_integration_errors` | Integration errors with other services | >0 |
| `credit_bureau_compliance_rule_violations` | Compliance rule violations | >0 |
| `credit_bureau_daily_limit_exceeded` | Number of daily limit exceeded attempts | >10 |

### 4.2 Grafana Dashboards

The following Grafana dashboards are available for operational monitoring:

#### 4.2.1 Operations Dashboard

- **URL**: https://grafana.innovabiz.com/d/credit-bureau-ops
- **Content**:
  - Query rate, errors, and latency
  - Service instance health and resource usage
  - Integration status with dependent services
  - Cache effectiveness metrics

```bash
# Export dashboard for reference
curl -X GET "https://grafana.innovabiz.com/api/dashboards/uid/credit-bureau-ops" \
  -H "Authorization: Bearer $GRAFANA_API_KEY" \
  -H "Accept: application/json" \
  > credit-bureau-ops-dashboard.json
```

#### 4.2.2 Business Dashboard

- **URL**: https://grafana.innovabiz.com/d/credit-bureau-business
- **Content**:
  - Query volume by market and type
  - Daily quota usage percentage
  - Credit scores distribution by market
  - Consent rejection rates

#### 4.2.3 Compliance Dashboard

- **URL**: https://grafana.innovabiz.com/d/credit-bureau-compliance
- **Content**:
  - Regulatory notifications sent by market
  - Compliance rule violations
  - Consent validation metrics
  - Authorization and authentication metrics

### 4.3 Alerting Rules

The following Prometheus alerting rules are configured:

#### 4.3.1 High Severity Alerts

```yaml
groups:
- name: credit-bureau-critical
  rules:
  - alert: CreditBureauHighErrorRate
    expr: sum(rate(credit_bureau_queries_error_total[5m])) / sum(rate(credit_bureau_queries_total[5m])) > 0.05
    for: 5m
    labels:
      severity: critical
      service: credit-bureau
    annotations:
      summary: "High error rate in Credit Bureau"
      description: "Credit Bureau experiencing {{ $value | humanizePercentage }} error rate (> 5%)"
      runbook: "https://wiki.innovabiz.com/ops/runbooks/credit-bureau-high-error-rate"

  - alert: CreditBureauServiceDown
    expr: up{app="credit-bureau"} == 0
    for: 2m
    labels:
      severity: critical
      service: credit-bureau
    annotations:
      summary: "Credit Bureau service down"
      description: "Credit Bureau instance {{ $labels.instance }} is down"
      runbook: "https://wiki.innovabiz.com/ops/runbooks/credit-bureau-service-down"

  - alert: CreditBureauHighLatency
    expr: histogram_quantile(0.95, sum(rate(credit_bureau_query_duration_seconds_bucket[5m])) by (le)) > 1.5
    for: 5m
    labels:
      severity: critical
      service: credit-bureau
    annotations:
      summary: "High query latency in Credit Bureau"
      description: "Credit Bureau P95 latency is {{ $value }}s (threshold: 1.5s)"
      runbook: "https://wiki.innovabiz.com/ops/runbooks/credit-bureau-high-latency"
```

#### 4.3.2 Medium Severity Alerts

```yaml
groups:
- name: credit-bureau-warning
  rules:
  - alert: CreditBureauQuotaNearLimit
    expr: credit_bureau_daily_quota_used_percent{market="angola"} > 85
    for: 5m
    labels:
      severity: warning
      service: credit-bureau
      market: "{{ $labels.market }}"
    annotations:
      summary: "Daily quota near limit for {{ $labels.market }}"
      description: "Credit Bureau {{ $labels.market }} at {{ $value | humanizePercentage }} of daily quota"
      runbook: "https://wiki.innovabiz.com/ops/runbooks/credit-bureau-quota-near-limit"

  - alert: CreditBureauComplianceViolation
    expr: sum(increase(credit_bureau_compliance_rule_violations[1h])) by (market) > 0
    for: 5m
    labels:
      severity: warning
      service: credit-bureau
    annotations:
      summary: "Compliance violations detected"
      description: "{{ $value }} compliance violations in the last hour for {{ $labels.market }}"
      runbook: "https://wiki.innovabiz.com/ops/runbooks/credit-bureau-compliance-violation"
```

### 4.4 Logging Strategy

The Credit Bureau module uses structured JSON logging with the following standards:

#### 4.4.1 Log Levels

| Level | Usage | Examples |
|-------|-------|----------|
| ERROR | System failures, data integrity issues | Service unavailable, database connection failure |
| WARN | Recoverable operational issues | Daily limit approaching, high latency, retries |
| INFO | Normal operations, business events | Query processed, notification sent |
| DEBUG | Detailed operations for troubleshooting | Request details, integration responses |
| TRACE | Extremely detailed for development | Code paths, variable values |

#### 4.4.2 Log Structure

```json
{
  "timestamp": "2025-08-06T14:35:22.357Z",
  "level": "INFO",
  "service": "credit-bureau",
  "instance": "credit-bureau-7d8f9c6b5-abcd1",
  "version": "1.0.0",
  "environment": "production",
  "market": "angola",
  "tenant": "fintech",
  "request_id": "f7d8e9c6-b5a4-3c2d-1e0f-9a8b7c6d5e4f",
  "trace_id": "0af7651916cd43dd8448eb211c80319c",
  "span_id": "b7ad6b7169203331",
  "user_id": "4f3e2d1c-b5a4-3c2d-1e0f-9a8b7c6d5e4f",
  "message": "Credit inquiry successfully processed",
  "details": {
    "inquiry_type": "score",
    "processing_time_ms": 237,
    "document_type": "national_id",
    "has_consent": true,
    "compliance_rules_applied": [
      "consent_validation",
      "angola_notification_required",
      "purpose_validation"
    ],
    "notification_sent": true
  }
}
```

#### 4.4.3 Log Querying

The logs can be queried in Kibana using structured fields:

```
service: "credit-bureau" AND level: "ERROR" AND market: "angola"
```

Or using the CLI for local testing:

```bash
# Query service logs
kubectl logs -l app=credit-bureau -n innovabiz | jq 'select(.level == "ERROR")'

# Get logs for specific request
kubectl logs -l app=credit-bureau -n innovabiz | jq 'select(.request_id == "f7d8e9c6-b5a4-3c2d-1e0f-9a8b7c6d5e4f")'
```

### 4.5 Distributed Tracing

The Credit Bureau module uses OpenTelemetry for distributed tracing with the following spans:

#### 4.5.1 Main Spans

| Span Name | Description | Child Spans |
|-----------|-------------|------------|
| `credit-bureau.query` | Overall credit query | Auth, validation, processing, notification |
| `credit-bureau.auth` | Authentication and authorization | Token validation, scope check |
| `credit-bureau.validation` | Validation of query parameters | Document check, consent validation |
| `credit-bureau.compliance` | Compliance rules checking | Market-specific rule checks |
| `credit-bureau.processing` | Core query processing | Data retrieval, scoring |
| `credit-bureau.notification` | Regulatory notifications | Notification preparation, sending |

#### 4.5.2 Example Trace Hierarchy

```
credit-bureau.query
├── credit-bureau.auth
│   ├── iam-service.validate-token
│   └── iam-service.check-scopes
├── credit-bureau.validation
│   ├── credit-bureau.validate-document
│   └── credit-bureau.validate-consent
├── credit-bureau.compliance
│   ├── credit-bureau.check-daily-limit
│   └── credit-bureau.apply-market-rules
├── credit-bureau.processing
│   ├── credit-bureau.retrieve-data
│   └── credit-bureau.calculate-score
└── credit-bureau.notification
    └── notification-service.send
```

#### 4.5.3 Critical Tags

| Tag Name | Purpose | Example Values |
|----------|---------|----------------|
| `market` | Identifies market context | `angola`, `brazil`, `eu`, `usa` |
| `inquiry_type` | Type of credit inquiry | `basic`, `complete`, `score` |
| `tenant_type` | Type of tenant | `bank`, `fintech`, `telco` |
| `document_type` | Type of document | `national_id`, `passport`, `tax_id` |
| `has_consent` | Presence of consent | `true`, `false` |
| `notification_required` | Notification requirement | `true`, `false` |

## 5. Troubleshooting

### 5.1 Common Issues and Resolution

#### 5.1.1 Authentication and Authorization Failures

| Issue | Symptoms | Diagnostic Commands | Resolution |
|-------|----------|---------------------|----------|
| Invalid JWT token | 401 Unauthorized, `auth_failures` metric increase | `curl -v -H "Authorization: Bearer $TOKEN" https://api.innovabiz.com/v1/credit-bureau/health/auth-test` | Verify token expiration, issuer configuration, or IAM service status |
| Insufficient scopes | 403 Forbidden, `scope_validation_failures` metric increase | `curl -v -H "Authorization: Bearer $TOKEN" https://api.innovabiz.com/v1/credit-bureau/health/scope-test` | Add required scopes to client registration or token request |
| Missing consent | 403 Forbidden, `consent_validation_failures` metric increase | Check logs: `kubectl logs -l app=credit-bureau -n innovabiz \| jq 'select(.details.has_consent == false)'` | Register consent via IAM consent endpoints |

#### 5.1.2 Performance Degradation

| Issue | Symptoms | Diagnostic Commands | Resolution |
|-------|----------|---------------------|-----------|
| High CPU usage | Increased latency, `credit_bureau_query_duration_seconds` increase | `kubectl top pods -l app=credit-bureau -n innovabiz` | Scale up resources, optimize queries, check for inefficient code |
| Memory leaks | Increasing memory usage over time | `kubectl describe pod -l app=credit-bureau -n innovabiz` | Identify leaking components, restart service, deploy fix |
| Slow integration services | Increased latency in integration spans | `kubectl exec -it $(kubectl get pods -l app=credit-bureau -n innovabiz -o jsonpath='{.items[0].metadata.name}') -- curl http://localhost:8080/internal/health/integrations` | Implement circuit breaker, request timeout, or verify integration service health |

#### 5.1.3 Compliance Errors

| Issue | Symptoms | Diagnostic Commands | Resolution |
|-------|----------|---------------------|-----------|
| Incorrect market rules | `compliance_rule_violations` metric increase | `kubectl exec -it $(kubectl get pods -l app=credit-bureau -n innovabiz -o jsonpath='{.items[0].metadata.name}') -- curl http://localhost:8080/internal/compliance/rules/status` | Update rules configuration, verify market context |
| Notification failures | `notification_errors` metric increase | `kubectl logs -l app=credit-bureau -n innovabiz \| jq 'select(.details.notification_sent == false)'` | Check notification service, verify templates, address connectivity issues |
| Data protection violations | GDPR or LGPD related errors | `kubectl logs -l app=credit-bureau -n innovabiz \| jq 'select(.details.compliance_rules_applied | contains(["data_protection"]))'` | Review data handling, apply masking rules, update consent requirements |

#### 5.1.4 Daily Limit Exceedance

| Issue | Symptoms | Diagnostic Commands | Resolution |
|-------|----------|---------------------|-----------|
| Quota exceeded | 429 Too Many Requests, `daily_limit_exceeded` metric increase | `kubectl exec -it $(kubectl get pods -l app=credit-bureau -n innovabiz -o jsonpath='{.items[0].metadata.name}') -- curl http://localhost:8080/internal/quotas/status` | Review usage patterns, increase limits if justified, implement client rate limiting |
| Counter inconsistencies | Incorrect quota tracking | `kubectl exec -it $(kubectl get pods -l app=credit-bureau -n innovabiz -o jsonpath='{.items[0].metadata.name}') -- curl http://localhost:8080/internal/quotas/reset-test` | Reset counters, check for distributed counter issues |
| Reset job failure | Quotas not resetting at midnight | `kubectl logs -l app=credit-bureau -n innovabiz -c quota-reset \| jq 'select(.message == "Quota reset")'` | Check cron job status, manually reset if necessary, repair reset mechanism |

### 5.2 Diagnostic Tools

#### 5.2.1 Health Check Endpoints

| Endpoint | Purpose | Example Command |
|----------|---------|----------------|
| `/health/live` | Basic liveness check | `curl https://api.innovabiz.com/v1/credit-bureau/health/live` |
| `/health/ready` | Readiness check including dependencies | `curl https://api.innovabiz.com/v1/credit-bureau/health/ready` |
| `/internal/health/deep` | Detailed health of all subsystems | `kubectl exec -it $(kubectl get pods -l app=credit-bureau -n innovabiz -o jsonpath='{.items[0].metadata.name}') -- curl http://localhost:8080/internal/health/deep` |
| `/internal/health/integrations` | Status of all integrations | `kubectl exec -it $(kubectl get pods -l app=credit-bureau -n innovabiz -o jsonpath='{.items[0].metadata.name}') -- curl http://localhost:8080/internal/health/integrations` |

#### 5.2.2 Profiling Endpoints

| Endpoint | Purpose | Example Command |
|----------|---------|----------------|
| `/internal/debug/pprof` | Go pprof profiling interface | `kubectl port-forward $(kubectl get pods -l app=credit-bureau -n innovabiz -o jsonpath='{.items[0].metadata.name}') 8080:8080` then access locally |
| `/internal/metrics/details` | Detailed internal metrics | `kubectl exec -it $(kubectl get pods -l app=credit-bureau -n innovabiz -o jsonpath='{.items[0].metadata.name}') -- curl http://localhost:8080/internal/metrics/details` |

#### 5.2.3 Configuration Verification

```bash
# Verify config in pod
kubectl exec -it $(kubectl get pods -l app=credit-bureau -n innovabiz -o jsonpath='{.items[0].metadata.name}') -n innovabiz -- env | grep CREDIT_BUREAU

# Check ConfigMap
kubectl get configmap credit-bureau-config -n innovabiz -o yaml

# Verify secrets mounting
kubectl exec -it $(kubectl get pods -l app=credit-bureau -n innovabiz -o jsonpath='{.items[0].metadata.name}') -n innovabiz -- ls -la /app/certs

# Test config validation
kubectl exec -it $(kubectl get pods -l app=credit-bureau -n innovabiz -o jsonpath='{.items[0].metadata.name}') -n innovabiz -- curl http://localhost:8080/internal/config/validate
```## 6. Maintenance Procedures

### 6.1 Scheduled Maintenance

Credit Bureau scheduled maintenance should follow the approved maintenance window calendar:

| Environment | Standard Window | Frequency | Prior Notice |
|------------|----------------|-----------|-------------|
| Development | Any time | As needed | None |
| Quality | Tuesday, 09:00-12:00 | Weekly | 24 hours |
| Staging | Wednesday, 09:00-12:00 | Bi-weekly | 48 hours |
| Production | Sunday, 01:00-05:00 | Monthly | 7 days |

#### 6.1.1 Standard Maintenance Procedure

1. **Announcement**:
   - Create Jira ticket (`MAINT-XXX`)
   - Notify stakeholders via email and Slack
   - Update maintenance calendar

2. **Preparation**:
   - Verify current service state
   - Ensure recent backups are available
   - Prepare rollback scripts
   - Validate updates in staging environment

3. **Execution**:
   - Enable maintenance banner through API Gateway
   - Gradually reduce traffic (if necessary)
   - Execute maintenance procedures
   - Verify functionality after maintenance
   - Run automated sanity tests

4. **Completion**:
   - Remove maintenance banner
   - Notify stakeholders of completion
   - Update maintenance ticket
   - Document any issues or lessons learned

### 6.2 Compliance Rules Update

The procedure for updating market-specific compliance rules should be executed when there are regulatory changes or business adjustments:

1. **Rules Preparation**:
   ```bash
   # Export current rules
   kubectl exec -it $(kubectl get pods -n innovabiz -l app=credit-bureau -o jsonpath='{.items[0].metadata.name}') -n innovabiz -- curl http://localhost:8080/internal/compliance/rules > compliance_rules_backup.json
   
   # Prepare file with new rules
   vim compliance_rules_new.json
   ```

2. **Staging Validation**:
   ```bash
   # Apply new rules in staging
   curl -X POST -H "Content-Type: application/json" -d @compliance_rules_new.json https://credit-bureau-staging.innovabiz.com/internal/compliance/update
   
   # Run automated compliance tests
   cd /path/to/tests && go test -v -tags=compliance ./...
   ```

3. **Production Deployment**:
   ```bash
   # Apply new rules in production
   curl -X POST -H "Content-Type: application/json" -d @compliance_rules_new.json https://credit-bureau.innovabiz.com/internal/compliance/update
   
   # Verify rules application
   curl https://credit-bureau.innovabiz.com/internal/compliance/rules/status
   ```

### 6.3 Secrets Rotation

Credit Bureau secrets (certificates, tokens, API keys) should be rotated regularly:

| Secret | Rotation Frequency | Procedure | Impact |
|--------|-------------------|-----------|--------|
| TLS Certificates | 12 months | Vault automated | No downtime |
| API Keys | 3 months | Manual procedure | Possible downtime |
| Service Tokens | 1 month | Automated rotation | No downtime |

Manual API Key rotation procedure:

1. **Generate new credentials**:
   ```bash
   # Generate new API key
   vault write secret/credit-bureau/api-keys/provider-xyz rotation=true
   
   # Retrieve new API key
   NEW_API_KEY=$(vault read -field=api_key secret/credit-bureau/api-keys/provider-xyz)
   ```

2. **Update service**:
   ```bash
   # Update secret in Kubernetes
   kubectl create secret generic credit-bureau-api-keys \
     --from-literal=provider-xyz=$NEW_API_KEY \
     -n innovabiz \
     --dry-run=client -o yaml | kubectl apply -f -
   
   # Restart pods to apply new configuration
   kubectl rollout restart deployment/credit-bureau -n innovabiz
   ```

3. **Verify functionality**:
   ```bash
   # Test integration with provider
   kubectl exec -it $(kubectl get pods -n innovabiz -l app=credit-bureau -o jsonpath='{.items[0].metadata.name}') -n innovabiz -- curl http://localhost:8080/internal/providers/test?provider=xyz
   ```

4. **Revoke old credentials**:
   ```bash
   # Revoke old API key after overlap period
   vault write secret/credit-bureau/api-keys/provider-xyz/revoke confirmed=true
   ```

### 6.4 Capacity Management

Credit Bureau capacity should be monitored and adjusted regularly to meet business demands:

#### 6.4.1 Capacity Monitoring

| Resource | Metric | Alert Threshold | Critical Threshold |
|---------|--------|-----------------|-------------------|
| CPU | `credit_bureau_cpu_usage_percent` | 70% for 30min | 85% for 15min |
| Memory | `credit_bureau_memory_usage_percent` | 75% for 30min | 90% for 15min |
| Disk | `credit_bureau_disk_usage_percent` | 75% | 90% |
| Network | `credit_bureau_network_saturation` | 70% | 85% |
| Requests | `credit_bureau_requests_per_second` | 80% of capacity | 90% of capacity |

#### 6.4.2 Capacity Planning

Capacity should be reviewed monthly, considering:
- Historical growth in query volume
- Business forecasts for new customers
- Expansion to new markets
- Seasonal changes (e.g., high volume periods)
- New query types or features

#### 6.4.3 Capacity Adjustment

```yaml
# HPA adjustment to increase capacity
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: credit-bureau-hpa
  namespace: innovabiz
spec:
  minReplicas: 5  # Increased from 3 to 5
  maxReplicas: 15  # Increased from 10 to 15
```

```yaml
# Pod resource adjustment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: credit-bureau
  namespace: innovabiz
spec:
  template:
    spec:
      containers:
      - name: credit-bureau
        resources:
          limits:
            cpu: "2"    # Increased from 1 to 2
            memory: "2Gi"  # Increased from 1Gi to 2Gi
          requests:
            cpu: "1"    # Increased from 500m to 1
            memory: "1Gi"  # Increased from 512Mi to 1Gi
```

## 7. Backup and Recovery Procedures

### 7.1 Backup Strategy

Credit Bureau uses a multi-layered approach for backups:

| Data | Frequency | Retention | Method | Responsible |
|------|-----------|----------|--------|------------|
| Configurations | Daily | 90 days | GitOps (ArgoCD) | DevSecOps |
| Cache Data | Daily | 7 days | Redis snapshot | Infrastructure |
| Metrics | Weekly | 1 year | Prometheus snapshot | Infrastructure |
| Logs | Continuous | 1 year | Log shipping | Infrastructure |
| Secrets | Weekly | 1 year | Vault export | Security |

### 7.2 Backup Procedures

#### 7.2.1 Configuration Backup

Configurations are managed via GitOps and stored in the Git repository:

```bash
# Check configuration repository status
argocd app get credit-bureau

# Export current configurations for validation
kubectl get cm,secret -l app=credit-bureau -n innovabiz -o yaml > bureau_config_backup.yaml
```

#### 7.2.2 Redis Cache Backup

```bash
# Initiate Redis snapshot
kubectl exec -it redis-master-0 -n cache -- redis-cli SAVE

# Copy RDB file to secure storage
kubectl cp cache/redis-master-0:/data/dump.rdb credit-bureau-redis-backup-$(date +%Y%m%d).rdb
```

### 7.3 Recovery Procedures

#### 7.3.1 Pod Failure Recovery

```bash
# Check pod status
kubectl get pods -n innovabiz -l app=credit-bureau

# Restart specific pod
kubectl delete pod credit-bureau-5d4f8c7b68-abcd1 -n innovabiz

# Check new pod logs
kubectl logs -f $(kubectl get pods -n innovabiz -l app=credit-bureau -o jsonpath='{.items[0].metadata.name}') -n innovabiz
```

#### 7.3.2 Configuration Recovery

```bash
# Roll back to previous version via GitOps
argocd app history credit-bureau
argocd app rollback credit-bureau 15  # Roll back to version 15

# Manually apply configuration backup (emergencies only)
kubectl apply -f bureau_config_backup.yaml
```

#### 7.3.3 Cache Recovery

```bash
# Copy RDB backup to Redis pod
kubectl cp credit-bureau-redis-backup-20250805.rdb cache/redis-master-0:/data/dump.rdb.restore

# Restore data within the pod
kubectl exec -it redis-master-0 -n cache -- bash
mv /data/dump.rdb.restore /data/dump.rdb
redis-cli SHUTDOWN SAVE
exit
```

### 7.4 Disaster Recovery Plan (DRP)

| Scenario | RTO | RPO | Procedure |
|---------|-----|-----|-----------|
| Pod Failure | 5 min | 0 | Automatic recovery via Kubernetes |
| Node Failure | 10 min | 0 | Automatic pod rescheduling |
| Zone Failure | 30 min | 5 min | Alternate zone activation |
| Region Failure | 1 hour | 15 min | Failover to secondary region |
| Data Corruption | 2 hours | 24 hours | Restore from backup |

#### 7.4.1 Regional Failover

```bash
# Check secondary region status
kubectl --context=gcp-europe-west4 get pods -n innovabiz

# Promote secondary region to primary
kubectl --context=gcp-europe-west4 patch configmap global-config -n innovabiz --type merge -p '{"data":{"PRIMARY_REGION":"europe-west4"}}'

# Update DNS to point to secondary region
kubectl --context=gcp-europe-west4 apply -f dns-failover.yaml

# Notify stakeholders
./scripts/notify-disaster-recovery.sh --event=regional-failover --region=europe-west4
```

## 8. Operational Security Policies

### 8.1 Operational Security Principles

Credit Bureau follows these operational security principles:

1. **Defense in Depth**: Multiple layers of security controls
2. **Principle of Least Privilege**: Minimum necessary access for each role
3. **Segregation of Duties**: Separation of responsibilities to prevent fraud
4. **Security by Design**: Security controls built into architecture
5. **Zero Trust**: No implicit trust, continuous verification

### 8.2 Access Controls

#### 8.2.1 Environment Access

| Access Level | Group | Permissions | Authentication Method |
|--------------|-------|-------------|----------------------|
| Read | credit-bureau-viewers | View logs, metrics, status | Standard MFA |
| Operation | credit-bureau-operators | Restart service, adjust settings | High MFA |
| Administration | credit-bureau-admins | Full access including secrets | High MFA + Approval |

#### 8.2.2 Credential Rotation

- Service credentials: Automatic rotation every 30 days
- User credentials: Expiration in 90 days
- Emergency credentials: Expiration in 24 hours

#### 8.2.3 Session Management

- Session timeout: 30 minutes of inactivity
- Maximum session duration: 8 hours
- Lockout after 5 unsuccessful attempts

### 8.3 Data Security

#### 8.3.1 Data Classification

| Classification | Examples | Controls |
|---------------|----------|----------|
| Public | Public documentation | No restrictions |
| Internal | Non-sensitive configurations | Basic authentication |
| Confidential | Query data, responses | Encryption, controlled access |
| Restricted | Personal documents, financial data | Strong encryption, high MFA, masking |

#### 8.3.2 Encryption

- Data in transit: TLS 1.3
- Data at rest: AES-256
- Sensitive keys: Managed via Vault with HSM

#### 8.3.3 Data Masking

```go
// Example implementation of sensitive data masking
func maskSensitiveData(data *QueryResult, market string) {
    switch market {
    case "brazil":
        // CPF: show only the last 3 digits
        if data.CustomerDocument != "" {
            data.CustomerDocument = "***.***.***-" + data.CustomerDocument[len(data.CustomerDocument)-2:]
        }
        
        // Financial data: exact values only with special scope
        if !hasScope(ctx, "credit_bureau:financial_data:complete") {
            for i := range data.CreditRecords {
                data.CreditRecords[i].Value = roundToRange(data.CreditRecords[i].Value)
            }
        }
    case "eu":
        // GDPR: removal of personal data without explicit consent
        if !hasConsent(ctx, "personal_data_processing") {
            data.FullAddress = ""
            data.PhoneNumbers = nil
        }
    }
    // Other market-specific rules...
}
```

### 8.4 Communications Security

#### 8.4.1 Network Policy

- Network segmentation via Network Policies
- Inter-pod traffic encrypted via Service Mesh
- External exposure only through API Gateway
- Blocking of unauthorized egress traffic

```yaml
# Network Policy for Credit Bureau
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: credit-bureau-network-policy
  namespace: innovabiz
spec:
  podSelector:
    matchLabels:
      app: credit-bureau
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: kube-system
      podSelector:
        matchLabels:
          k8s-app: kube-proxy
    - podSelector:
        matchLabels:
          app: api-gateway
    - podSelector:
        matchLabels:
          app: otel-collector
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: UDP
      port: 53
    - protocol: TCP
      port: 53
  - to:
    - podSelector:
        matchLabels:
          app: iam-service
    ports:
    - protocol: TCP
      port: 8080
  - to:
    - podSelector:
        matchLabels:
          app: otel-collector
    ports:
    - protocol: TCP
      port: 4317
```

#### 8.4.2 API Security

- Input validation via JSON Schema
- Protection against injection attacks
- Client-based rate limiting
- OWASP Top 10 protection

### 8.5 Vulnerability Management

#### 8.5.1 Patch Lifecycle

| Severity | Application SLA | Maintenance Window |
|----------|----------------|-------------------|
| Critical | 24 hours | Immediate |
| High | 7 days | Next window |
| Medium | 30 days | Monthly window |
| Low | 90 days | Quarterly window |

#### 8.5.2 Management Process

```bash
# Check image vulnerabilities
trivy image innovabiz/credit-bureau:latest

# Check deployment vulnerabilities
kubectl-trivy deployment credit-bureau -n innovabiz

# Generate compliance report
trivy image --format json --output credit-bureau-vulnerabilities.json innovabiz/credit-bureau:latest
```## 9. Market-Specific Compliance

### 9.1 Angola (BNA)

#### 9.1.1 Regulatory Requirements

- **Legal Basis**: Notice No. 05/2021 of Banco Nacional de Angola (BNA)
- **Scope**: Banking and non-banking financial institutions
- **Key Requirements**:
  - Explicit consent for inquiries
  - Mandatory notification for all inquiries
  - High-level MFA for complete inquiries
  - Record retention for 5 years

#### 9.1.2 Specific Configuration

```yaml
# Angola-specific ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: credit-bureau-angola-config
  namespace: innovabiz
data:
  CONSENT_REQUIRED: "true"
  NOTIFICATION_REQUIRED: "true"
  MIN_MFA_LEVEL_COMPLETE_INQUIRY: "high"
  MIN_MFA_LEVEL_SCORE_INQUIRY: "standard"
  RETENTION_PERIOD_DAYS: "1825" # 5 years
  DAILY_QUERY_LIMIT: "50"
  MASKING_RULES: |
    {
      "CustomerDocument": "mask-partial",
      "CreditValue": "no-mask",
      "Address": "mask-partial"
    }
```

#### 9.1.3 Audit Procedures

Procedures to be executed quarterly:

1. **Consent Verification**:
   ```bash
   # Export consent logs for analysis
   kubectl exec -it $(kubectl get pods -n innovabiz -l app=credit-bureau -o jsonpath='{.items[0].metadata.name}') -n innovabiz -- curl -o /tmp/consent-audit.json http://localhost:8080/internal/audit/consent?market=angola&startDate=2025-05-01
   
   # Analyze compliance rates
   kubectl exec -it $(kubectl get pods -n innovabiz -l app=credit-bureau -o jsonpath='{.items[0].metadata.name}') -n innovabiz -- curl http://localhost:8080/internal/compliance/stats?market=angola | jq '.consentStats'
   ```

2. **Notification Verification**:
   ```bash
   # Check notification records
   kubectl exec -it $(kubectl get pods -n innovabiz -l app=credit-bureau -o jsonpath='{.items[0].metadata.name}') -n innovabiz -- curl http://localhost:8080/internal/notifications/stats?market=angola
   ```

### 9.2 Brazil (BACEN/LGPD)

#### 9.2.1 Regulatory Requirements

- **Legal Basis**: Lei Geral de Proteção de Dados (LGPD), Resolution No. 4,737 of the Central Bank of Brazil
- **Scope**: Financial institutions, credit bureaus
- **Key Requirements**:
  - Specific purpose for each inquiry
  - Mandatory notification for credit restrictions
  - Consent for sharing personal data
  - Right of access, correction, and deletion of data

#### 9.2.2 Specific Configuration

```yaml
# Brazil-specific ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: credit-bureau-brazil-config
  namespace: innovabiz
data:
  PURPOSE_REQUIRED: "true"
  NOTIFICATION_FOR_RESTRICTIONS: "true"
  CONSENT_REQUIRED: "true"
  MIN_MFA_LEVEL_COMPLETE_INQUIRY: "standard"
  RETENTION_PERIOD_DAYS: "730" # 2 years
  DATA_SUBJECT_RIGHTS_ENABLED: "true"
  DAILY_QUERY_LIMIT: "200"
  LGPD_DATA_CATEGORIES: |
    {
      "personal_data": ["name", "cpf", "address", "phone"],
      "financial_data": ["score", "restrictions", "credit_history"],
      "sensitive_data": []
    }
```

### 9.3 European Union (GDPR/PSD2)

#### 9.3.1 Regulatory Requirements

- **Legal Basis**: General Data Protection Regulation (GDPR), Payment Services Directive 2 (PSD2)
- **Scope**: Entities processing data of EU citizens
- **Key Requirements**:
  - Data minimization for specific purpose
  - Explicit and specific consent
  - Right to be forgotten
  - Data breach notification within 72 hours

#### 9.3.2 Specific Configuration

```yaml
# EU-specific ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: credit-bureau-eu-config
  namespace: innovabiz
data:
  DATA_MINIMIZATION: "true"
  EXPLICIT_CONSENT_REQUIRED: "true"
  RIGHT_TO_BE_FORGOTTEN_ENABLED: "true"
  DATA_BREACH_NOTIFICATION_ENABLED: "true"
  MIN_MFA_LEVEL_COMPLETE_INQUIRY: "high"
  RETENTION_PERIOD_DAYS: "365" # 1 year
  CONSENT_EXPIRY_DAYS: "90" # Expiration in 90 days
  DATA_PORTABILITY_ENABLED: "true"
  DAILY_QUERY_LIMIT: "100"
```

### 9.4 USA (FCRA/GLBA)

#### 9.4.1 Regulatory Requirements

- **Legal Basis**: Fair Credit Reporting Act (FCRA), Gramm-Leach-Bliley Act (GLBA)
- **Scope**: Credit bureaus, financial institutions
- **Key Requirements**:
  - Permissible purpose for inquiries
  - Adverse action notification
  - Dispute resolution right
  - Financial data security

#### 9.4.2 Specific Configuration

```yaml
# USA-specific ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: credit-bureau-usa-config
  namespace: innovabiz
data:
  PERMISSIBLE_PURPOSE_REQUIRED: "true"
  ADVERSE_ACTION_NOTIFICATION: "true"
  DISPUTE_RESOLUTION_ENABLED: "true"
  SAFEGUARDS_RULE_COMPLIANCE: "true"
  MIN_MFA_LEVEL_COMPLETE_INQUIRY: "standard"
  RETENTION_PERIOD_DAYS: "2555" # 7 years
  IDENTITY_THEFT_PROTECTION: "true"
  DAILY_QUERY_LIMIT: "150"
```

## 10. Configuration Management

### 10.1 Configuration Strategy

Credit Bureau uses a layered approach to configuration:

1. **Base Configuration**: Applied to all environments and markets
2. **Environment Configuration**: Overrides base configuration per environment (dev, qa, staging, prod)
3. **Market Configuration**: Overrides configuration per market (angola, brazil, eu, usa, global)
4. **Tenant Configuration**: Tenant-specific adjustments when necessary

### 10.2 Configuration Sources

| Priority | Source | Type | Usage |
|----------|-------|------|-------|
| 1 (lowest) | Hardcoded defaults | Code values | Last-resort defaults |
| 2 | Base ConfigMap | Global values | Common settings |
| 3 | Environment ConfigMap | Per-environment values | Service endpoints |
| 4 | Market ConfigMap | Per-market values | Specific rules |
| 5 | Secret | Sensitive values | Credentials, keys |
| 6 (highest) | Environment variables | Override | Emergency settings |

### 10.3 Configuration Keys

| Key | Description | Default Values | Applicable Environments |
|-----|------------|----------------|------------------------|
| `ENVIRONMENT` | Execution environment | `production` | All |
| `MARKET` | Default market | `global` | All |
| `TENANT_TYPE` | Tenant type | `default` | All |
| `SERVICE_VERSION` | Service version | `1.0.0` | All |
| `LOG_LEVEL` | Log level | `info` | All |
| `DAILY_QUERY_LIMIT_*` | Daily limit per market | Varies | All |
| `NOTIFICATION_REQUIRED_*` | Notification requirements | Varies | All |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OpenTelemetry collector endpoint | `http://otel-collector:4317` | All |
| `REDIS_HOST` | Redis host for cache | `redis-master.cache` | All |
| `REDIS_PORT` | Redis port | `6379` | All |
| `IAM_SERVICE_URL` | IAM service URL | Varies | All |
| `COMPLIANCE_RULES_PATH` | Rules file path | `/app/config/compliance-rules.json` | All |
| `MFA_REQUIRED_LEVELS` | MFA levels per operation | JSON with settings | All |

### 10.4 Secrets Management

Credit Bureau uses HashiCorp Vault integrated with Kubernetes for secrets management:

```bash
# Check secrets status
kubectl get secrets -n innovabiz -l app=credit-bureau

# Rotate secrets
vault write -f secret/credit-bureau/rotate

# Synchronize secrets with Kubernetes
vault-k8s-sync credit-bureau
```

## 11. References

### 11.1 Internal Documentation

- [ADR: Credit Bureau](../adr/bureau-credito-adr.md)
- [Technical Specification](../technical/bureau-credito-technical-spec.md)
- [Integration Guide](../integration/bureau-credito-integration-guide.md)
- [Test Plan](../testing/bureau-credito-test-plan.md)
- [Runbooks](../runbooks/)

### 11.2 Tools Documentation

- [Kubernetes](https://kubernetes.io/docs/)
- [OpenTelemetry](https://opentelemetry.io/docs/)
- [Prometheus](https://prometheus.io/docs/)
- [Grafana](https://grafana.com/docs/)
- [KrakenD API Gateway](https://www.krakend.io/docs/)
- [HashiCorp Vault](https://www.vaultproject.io/docs)

### 11.3 Regulations

- [BNA - Notice No. 05/2021](https://www.bna.ao/)
- [BACEN - Resolution No. 4,737](https://www.bcb.gov.br/)
- [LGPD - Lei Geral de Proteção de Dados](https://www.lgpdbrasil.com.br/)
- [GDPR](https://gdpr.eu/)
- [PSD2](https://ec.europa.eu/info/law/payment-services-psd-2-directive-eu-2015-2366_en)
- [FCRA](https://www.ftc.gov/enforcement/statutes/fair-credit-reporting-act)
- [GLBA](https://www.ftc.gov/business-guidance/privacy-security/gramm-leach-bliley-act)

---

**Revision History**

| Version | Date | Author | Description |
|--------|------|-------|-------------|
| 0.1 | 2025-07-15 | Operations Team | Initial version |
| 0.2 | 2025-07-28 | Security Team | Addition of security policies |
| 0.3 | 2025-08-01 | Compliance Team | Addition of regulatory requirements |
| 1.0 | 2025-08-06 | Operations Team | Final approved version |

**Classification: Private - Internal Use**