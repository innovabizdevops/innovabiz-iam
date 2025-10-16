# Template Kubernetes YAML para AlertManager - INNOVABIZ

## Visão Geral

Este documento fornece templates YAML para implantação do AlertManager no ambiente Kubernetes da plataforma INNOVABIZ. O AlertManager é responsável por gerenciar, agrupar e encaminhar alertas gerados pelo Prometheus para os canais de notificação apropriados. Os templates seguem as melhores práticas de segurança, dimensionamento e configuração multi-dimensional conforme os padrões INNOVABIZ.

## Sumário

1. [Namespace](#namespace)
2. [ConfigMap](#configmap)
3. [Secret](#secret)
4. [PersistentVolumeClaim](#persistentvolumeclaim)
5. [ServiceAccount e RBAC](#serviceaccount-e-rbac)
6. [Deployment](#deployment)
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
    innovabiz.com/module: "alerting"
```

## ConfigMap

```yaml
# alertmanager-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: alertmanager-config
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: alertmanager
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "alerting"
data:
  alertmanager.yml: |
    global:
      # Configuração global de templates para notificações
      smtp_smarthost: '${SMTP_SERVER}:${SMTP_PORT}'
      smtp_from: '${SMTP_FROM}'
      smtp_auth_username: '${SMTP_USER}'
      smtp_auth_password: '${SMTP_PASSWORD}'
      smtp_require_tls: true
      slack_api_url: '${SLACK_API_URL}'
      http_config:
        follow_redirects: true
        enable_http2: true
      resolve_timeout: 5m

    # Configuração de encaminhamentos, armazenamento e templates
    route:
      # Estratégia global de agrupamento
      group_by: ['tenant_id', 'region_id', 'environment', 'alertname', 'service', 'severity']
      # Espera entre notificações
      group_wait: 30s
      # Intervalo entre notificações do mesmo grupo
      group_interval: 5m
      # Intervalo para repetir alertas não resolvidos
      repeat_interval: ${REPEAT_INTERVAL}
      # Receptor padrão
      receiver: 'ops-${TENANT_ID}-${REGION_ID}-${ENVIRONMENT}'

      # Rotas específicas por severidade, serviço e ambiente
      routes:
      - receiver: 'critical-${TENANT_ID}-${REGION_ID}-${ENVIRONMENT}'
        matchers:
          - severity="critical"
          - tenant_id="${TENANT_ID}"
          - region_id="${REGION_ID}"
          - environment="${ENVIRONMENT}"
        continue: true

      - receiver: 'warning-${TENANT_ID}-${REGION_ID}-${ENVIRONMENT}'
        matchers:
          - severity="warning"
          - tenant_id="${TENANT_ID}"
          - region_id="${REGION_ID}"
          - environment="${ENVIRONMENT}"
        continue: true

      - receiver: 'info-${TENANT_ID}-${REGION_ID}-${ENVIRONMENT}'
        matchers:
          - severity="info"
          - tenant_id="${TENANT_ID}"
          - region_id="${REGION_ID}"
          - environment="${ENVIRONMENT}"
        continue: true

      # Rotas adicionais baseadas em serviços específicos
      - receiver: 'infrastructure-${TENANT_ID}-${REGION_ID}-${ENVIRONMENT}'
        matchers:
          - service=~"kubernetes|node|cluster"
          - tenant_id="${TENANT_ID}"
          - region_id="${REGION_ID}"
          - environment="${ENVIRONMENT}"
        continue: true

      - receiver: 'database-${TENANT_ID}-${REGION_ID}-${ENVIRONMENT}'
        matchers:
          - service=~"postgresql|mongodb|redis"
          - tenant_id="${TENANT_ID}"
          - region_id="${REGION_ID}"
          - environment="${ENVIRONMENT}"
        continue: true

      - receiver: 'application-${TENANT_ID}-${REGION_ID}-${ENVIRONMENT}'
        matchers:
          - service=~"api|backend|frontend"
          - tenant_id="${TENANT_ID}"
          - region_id="${REGION_ID}"
          - environment="${ENVIRONMENT}"
        continue: true

      - receiver: 'business-${TENANT_ID}-${REGION_ID}-${ENVIRONMENT}'
        matchers:
          - category="business"
          - tenant_id="${TENANT_ID}"
          - region_id="${REGION_ID}"
          - environment="${ENVIRONMENT}"
        continue: true

    # Inibidores para evitar tempestade de alertas
    inhibit_rules:
      # Se houver um alerta crítico para um serviço, inibir alertas de warning do mesmo serviço
      - source_matchers:
        - severity="critical"
        target_matchers:
        - severity="warning"
        - severity="info"
        # Aplicar apenas se esses labels forem iguais
        equal: ['tenant_id', 'region_id', 'environment', 'service', 'instance']

    # Configuração de receptores (canais de notificação)
    receivers:
      - name: 'ops-${TENANT_ID}-${REGION_ID}-${ENVIRONMENT}'
        email_configs:
          - to: '${OPS_EMAIL}'
            send_resolved: true
            html: '{{ template "email.innovabiz.html" . }}'
        slack_configs:
          - channel: '#ops-alerts-${ENVIRONMENT}'
            send_resolved: true
            title: '[{{ .Status | toUpper }}{{ if eq .Status "firing" }}:{{ .Alerts.Firing | len }}{{ end }}] {{ .GroupLabels.SortedPairs.Values | join " " }} - {{ .CommonLabels.alertname }}'
            text: "{{ range .Alerts }}*Alert:* {{ .Annotations.summary }}\n*Description:* {{ .Annotations.description }}\n*Severity:* {{ .Labels.severity }}\n*Tenant:* {{ .Labels.tenant_id }}\n*Region:* {{ .Labels.region_id }}\n*Environment:* {{ .Labels.environment }}\n*Service:* {{ .Labels.service }}\n*Started:* {{ .StartsAt | since }}\n{{ if ne .Status \"firing\" }}*Ended:* {{ .EndsAt | since }}{{ end }}\n{{ end }}"
        webhook_configs:
          - url: 'https://hooks.opsgenie.com/v2/api/alertmanager/${OPSGENIE_KEY}'
            send_resolved: true

      - name: 'critical-${TENANT_ID}-${REGION_ID}-${ENVIRONMENT}'
        pagerduty_configs:
          - service_key: ${PAGERDUTY_SERVICE_KEY}
            send_resolved: true
            severity: critical
            description: '{{ .CommonAnnotations.summary }}'
            details:
              tenant: '{{ .CommonLabels.tenant_id }}'
              region: '{{ .CommonLabels.region_id }}'
              environment: '{{ .CommonLabels.environment }}'
              service: '{{ .CommonLabels.service }}'
        slack_configs:
          - channel: '#critical-alerts-${ENVIRONMENT}'
            send_resolved: true
            color: '#FF0000'
            title: '[CRITICAL] {{ .GroupLabels.SortedPairs.Values | join " " }} - {{ .CommonLabels.alertname }}'
            text: "{{ range .Alerts }}*Alert:* {{ .Annotations.summary }}\n*Description:* {{ .Annotations.description }}\n*Tenant:* {{ .Labels.tenant_id }}\n*Region:* {{ .Labels.region_id }}\n*Environment:* {{ .Labels.environment }}\n*Service:* {{ .Labels.service }}\n*Started:* {{ .StartsAt | since }}\n{{ if ne .Status \"firing\" }}*Ended:* {{ .EndsAt | since }}{{ end }}\n{{ end }}"

      - name: 'warning-${TENANT_ID}-${REGION_ID}-${ENVIRONMENT}'
        email_configs:
          - to: '${WARNING_EMAIL}'
            send_resolved: true
            html: '{{ template "email.innovabiz.html" . }}'
        slack_configs:
          - channel: '#warning-alerts-${ENVIRONMENT}'
            send_resolved: true
            color: '#FFBF00'
            title: '[WARNING] {{ .GroupLabels.SortedPairs.Values | join " " }} - {{ .CommonLabels.alertname }}'
            text: "{{ range .Alerts }}*Alert:* {{ .Annotations.summary }}\n*Description:* {{ .Annotations.description }}\n*Tenant:* {{ .Labels.tenant_id }}\n*Region:* {{ .Labels.region_id }}\n*Environment:* {{ .Labels.environment }}\n*Service:* {{ .Labels.service }}\n*Started:* {{ .StartsAt | since }}\n{{ end }}"

      - name: 'info-${TENANT_ID}-${REGION_ID}-${ENVIRONMENT}'
        slack_configs:
          - channel: '#info-alerts-${ENVIRONMENT}'
            send_resolved: true
            color: '#36A64F'
            title: '[INFO] {{ .GroupLabels.SortedPairs.Values | join " " }} - {{ .CommonLabels.alertname }}'
            text: "{{ range .Alerts }}*Alert:* {{ .Annotations.summary }}\n*Description:* {{ .Annotations.description }}\n*Tenant:* {{ .Labels.tenant_id }}\n*Region:* {{ .Labels.region_id }}\n*Environment:* {{ .Labels.environment }}\n*Service:* {{ .Labels.service }}\n*Started:* {{ .StartsAt | since }}\n{{ end }}"

      - name: 'infrastructure-${TENANT_ID}-${REGION_ID}-${ENVIRONMENT}'
        email_configs:
          - to: '${INFRA_EMAIL}'
            send_resolved: true
            html: '{{ template "email.innovabiz.html" . }}'
        slack_configs:
          - channel: '#infra-alerts-${ENVIRONMENT}'
            send_resolved: true
            title: '[INFRA] {{ .GroupLabels.SortedPairs.Values | join " " }} - {{ .CommonLabels.alertname }}'
            text: "{{ range .Alerts }}*Alert:* {{ .Annotations.summary }}\n*Description:* {{ .Annotations.description }}\n*Severity:* {{ .Labels.severity }}\n*Tenant:* {{ .Labels.tenant_id }}\n*Region:* {{ .Labels.region_id }}\n*Environment:* {{ .Labels.environment }}\n*Service:* {{ .Labels.service }}\n*Started:* {{ .StartsAt | since }}\n{{ end }}"

      - name: 'database-${TENANT_ID}-${REGION_ID}-${ENVIRONMENT}'
        email_configs:
          - to: '${DB_EMAIL}'
            send_resolved: true
            html: '{{ template "email.innovabiz.html" . }}'
        slack_configs:
          - channel: '#db-alerts-${ENVIRONMENT}'
            send_resolved: true
            title: '[DB] {{ .GroupLabels.SortedPairs.Values | join " " }} - {{ .CommonLabels.alertname }}'
            text: "{{ range .Alerts }}*Alert:* {{ .Annotations.summary }}\n*Description:* {{ .Annotations.description }}\n*Severity:* {{ .Labels.severity }}\n*Tenant:* {{ .Labels.tenant_id }}\n*Region:* {{ .Labels.region_id }}\n*Environment:* {{ .Labels.environment }}\n*Service:* {{ .Labels.service }}\n*Started:* {{ .StartsAt | since }}\n{{ end }}"

      - name: 'application-${TENANT_ID}-${REGION_ID}-${ENVIRONMENT}'
        email_configs:
          - to: '${APP_EMAIL}'
            send_resolved: true
            html: '{{ template "email.innovabiz.html" . }}'
        slack_configs:
          - channel: '#app-alerts-${ENVIRONMENT}'
            send_resolved: true
            title: '[APP] {{ .GroupLabels.SortedPairs.Values | join " " }} - {{ .CommonLabels.alertname }}'
            text: "{{ range .Alerts }}*Alert:* {{ .Annotations.summary }}\n*Description:* {{ .Annotations.description }}\n*Severity:* {{ .Labels.severity }}\n*Tenant:* {{ .Labels.tenant_id }}\n*Region:* {{ .Labels.region_id }}\n*Environment:* {{ .Labels.environment }}\n*Service:* {{ .Labels.service }}\n*Started:* {{ .StartsAt | since }}\n{{ end }}"

      - name: 'business-${TENANT_ID}-${REGION_ID}-${ENVIRONMENT}'
        email_configs:
          - to: '${BUSINESS_EMAIL}'
            send_resolved: true
            html: '{{ template "email.innovabiz.html" . }}'
        slack_configs:
          - channel: '#business-alerts-${ENVIRONMENT}'
            send_resolved: true
            title: '[BUSINESS] {{ .GroupLabels.SortedPairs.Values | join " " }} - {{ .CommonLabels.alertname }}'
            text: "{{ range .Alerts }}*Alert:* {{ .Annotations.summary }}\n*Description:* {{ .Annotations.description }}\n*Severity:* {{ .Labels.severity }}\n*Tenant:* {{ .Labels.tenant_id }}\n*Region:* {{ .Labels.region_id }}\n*Environment:* {{ .Labels.environment }}\n*Service:* {{ .Labels.service }}\n*Started:* {{ .StartsAt | since }}\n{{ end }}"

  # Templates personalizados para notificações
  email.tmpl: |
    {{ define "email.innovabiz.html" }}
    <!DOCTYPE html>
    <html>
    <head>
      <meta charset="UTF-8">
      <title>{{ template "__subject" . }}</title>
      <style>
        body {
          font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif;
          font-size: 14px;
          color: #333;
          margin: 0;
          padding: 0;
        }
        .container {
          max-width: 800px;
          margin: 0 auto;
          padding: 20px;
        }
        .header {
          background-color: #2c3e50;
          color: white;
          padding: 15px;
          border-top-left-radius: 5px;
          border-top-right-radius: 5px;
        }
        .content {
          background-color: #f9f9f9;
          padding: 15px;
          border: 1px solid #ddd;
        }
        .alert {
          margin-bottom: 15px;
          padding: 10px;
          background-color: #fff;
          border-left: 5px solid #ddd;
        }
        .alert-firing {
          border-left-color: #e74c3c;
        }
        .alert-resolved {
          border-left-color: #2ecc71;
        }
        .alert-critical {
          border-left-color: #e74c3c;
        }
        .alert-warning {
          border-left-color: #f39c12;
        }
        .alert-info {
          border-left-color: #3498db;
        }
        .footer {
          margin-top: 20px;
          text-align: center;
          font-size: 12px;
          color: #777;
        }
        table {
          width: 100%;
          border-collapse: collapse;
          margin-top: 10px;
        }
        th, td {
          padding: 8px;
          text-align: left;
          border-bottom: 1px solid #ddd;
        }
        th {
          background-color: #f2f2f2;
        }
      </style>
    </head>
    <body>
      <div class="container">
        <div class="header">
          <h2>{{ template "__subject" . }}</h2>
        </div>
        <div class="content">
          <p>{{ .Alerts | len }} alerta(s) {{ .Status }}</p>
          
          {{ range .Alerts }}
          <div class="alert {{ if eq .Status "firing" }}alert-firing{{ else }}alert-resolved{{ end }} {{ if .Labels.severity }}alert-{{ .Labels.severity }}{{ end }}">
            <h3>{{ .Annotations.summary }}</h3>
            <p>{{ .Annotations.description }}</p>
            
            <table>
              <tr>
                <th>Status:</th>
                <td>{{ .Status | toUpper }}</td>
              </tr>
              <tr>
                <th>Tenant:</th>
                <td>{{ .Labels.tenant_id }}</td>
              </tr>
              <tr>
                <th>Região:</th>
                <td>{{ .Labels.region_id }}</td>
              </tr>
              <tr>
                <th>Ambiente:</th>
                <td>{{ .Labels.environment }}</td>
              </tr>
              <tr>
                <th>Serviço:</th>
                <td>{{ .Labels.service }}</td>
              </tr>
              <tr>
                <th>Severidade:</th>
                <td>{{ .Labels.severity }}</td>
              </tr>
              <tr>
                <th>Início:</th>
                <td>{{ .StartsAt }}</td>
              </tr>
              {{ if ne .Status "firing" }}
              <tr>
                <th>Fim:</th>
                <td>{{ .EndsAt }}</td>
              </tr>
              {{ end }}
              {{ if .GeneratorURL }}
              <tr>
                <th>Source:</th>
                <td><a href="{{ .GeneratorURL }}">{{ .GeneratorURL }}</a></td>
              </tr>
              {{ end }}
            </table>
            
            <!-- Labels extras -->
            {{ if gt (len .Labels.SortedPairs) 0 }}
            <h4>Labels Adicionais:</h4>
            <table>
              {{ range .Labels.SortedPairs }}
              {{ if and (ne .Name "tenant_id") (ne .Name "region_id") (ne .Name "environment") (ne .Name "service") (ne .Name "severity") (ne .Name "alertname") }}
              <tr>
                <th>{{ .Name }}:</th>
                <td>{{ .Value }}</td>
              </tr>
              {{ end }}
              {{ end }}
            </table>
            {{ end }}
          </div>
          {{ end }}
        </div>
        <div class="footer">
          <p>Este é um alerta gerado automaticamente pelo sistema de monitoramento INNOVABIZ.</p>
          <p>© {{ now.Format "2006" }} INNOVABIZ - Framework de Observabilidade</p>
        </div>
      </div>
    </body>
    </html>
    {{ end }}
```

## Secret

```yaml
# alertmanager-secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: alertmanager-secrets
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: alertmanager
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "alerting"
type: Opaque
stringData:
  # Credenciais para SMTP
  smtp_password: "${SMTP_PASSWORD}"
  # Credenciais para PagerDuty
  pagerduty_service_key: "${PAGERDUTY_SERVICE_KEY}"
  # Token Slack
  slack_api_url: "${SLACK_API_URL}"
  # Chave OpsGenie
  opsgenie_api_key: "${OPSGENIE_KEY}"
  # Credenciais para webhook receiver
  webhook_auth_token: "${WEBHOOK_AUTH_TOKEN}"
  # Credenciais para API de administração
  admin_password: "${ADMIN_PASSWORD}"
  # Chave para criptografia de silences
  silence_encryption_key: "${SILENCE_KEY}"
```

## PersistentVolumeClaim

```yaml
# alertmanager-pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: alertmanager-storage
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: alertmanager
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "alerting"
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: "${STORAGE_CLASS}"
  resources:
    requests:
      storage: ${ALERTMANAGER_STORAGE_SIZE}
```

## ServiceAccount e RBAC

```yaml
# alertmanager-rbac.yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: alertmanager
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: alertmanager
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "alerting"
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: alertmanager
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: alertmanager
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "alerting"
rules:
  # Permissões mínimas necessárias para operação
  - apiGroups: [""]
    resources: ["configmaps", "secrets"]
    verbs: ["get", "list", "watch"]
  # Permissões para serviço de descoberta de endpoints
  - apiGroups: [""]
    resources: ["services", "endpoints"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: alertmanager
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: alertmanager
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "alerting"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: alertmanager
subjects:
  - kind: ServiceAccount
    name: alertmanager
    namespace: observability-${TENANT_ID}-${REGION_ID}
```## Deployment

```yaml
# alertmanager-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: alertmanager
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: alertmanager
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "alerting"
spec:
  replicas: ${ALERTMANAGER_REPLICAS}
  selector:
    matchLabels:
      app: alertmanager
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app: alertmanager
        innovabiz.com/tenant: "${TENANT_ID}"
        innovabiz.com/region: "${REGION_ID}"
        innovabiz.com/environment: "${ENVIRONMENT}"
        innovabiz.com/component: "observability"
        innovabiz.com/module: "alerting"
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9093"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: alertmanager
      securityContext:
        fsGroup: 65534  # nobody
        runAsNonRoot: true
        runAsUser: 65534  # nobody
        runAsGroup: 65534  # nobody
      containers:
      - name: alertmanager
        image: prom/alertmanager:${ALERTMANAGER_VERSION}
        imagePullPolicy: IfNotPresent
        args:
        - "--config.file=/etc/alertmanager/alertmanager.yml"
        - "--storage.path=/alertmanager"
        - "--cluster.listen-address=[$(POD_IP)]:9094"
        - "--cluster.advertise-address=$(POD_IP):9094"
        - "--web.external-url=https://alertmanager-${TENANT_ID}-${REGION_ID}.${DOMAIN}/alertmanager"
        - "--web.route-prefix=/alertmanager"
        - "--log.level=${LOG_LEVEL}"
        - "--log.format=json"
        - "--cluster.peer=alertmanager-0.alertmanager:9094"
        - "--cluster.peer=alertmanager-1.alertmanager:9094"
        - "--cluster.settle-timeout=1m"
        env:
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        - name: TENANT_ID
          value: "${TENANT_ID}"
        - name: REGION_ID
          value: "${REGION_ID}"
        - name: ENVIRONMENT
          value: "${ENVIRONMENT}"
        ports:
        - name: http
          containerPort: 9093
          protocol: TCP
        - name: mesh
          containerPort: 9094
          protocol: TCP
        resources:
          requests:
            cpu: ${ALERTMANAGER_CPU_REQUEST}
            memory: ${ALERTMANAGER_MEMORY_REQUEST}
          limits:
            cpu: ${ALERTMANAGER_CPU_LIMIT}
            memory: ${ALERTMANAGER_MEMORY_LIMIT}
        readinessProbe:
          httpGet:
            path: /alertmanager/-/ready
            port: 9093
          initialDelaySeconds: 30
          timeoutSeconds: 5
          periodSeconds: 10
          failureThreshold: 3
        livenessProbe:
          httpGet:
            path: /alertmanager/-/healthy
            port: 9093
          initialDelaySeconds: 60
          timeoutSeconds: 5
          periodSeconds: 15
          failureThreshold: 6
        volumeMounts:
        - name: config-volume
          mountPath: /etc/alertmanager
        - name: storage-volume
          mountPath: /alertmanager
        - name: templates-volume
          mountPath: /etc/alertmanager/templates
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
          readOnlyRootFilesystem: true
      volumes:
      - name: config-volume
        configMap:
          name: alertmanager-config
      - name: templates-volume
        configMap:
          name: alertmanager-config
          items:
            - key: email.tmpl
              path: email.tmpl
      - name: storage-volume
        persistentVolumeClaim:
          claimName: alertmanager-storage
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
                  - alertmanager
              topologyKey: kubernetes.io/hostname
      terminationGracePeriodSeconds: 60
```

## Service

```yaml
# alertmanager-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: alertmanager
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: alertmanager
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "alerting"
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "9093"
    prometheus.io/path: "/metrics"
spec:
  type: ClusterIP
  ports:
  - name: http
    port: 9093
    targetPort: 9093
    protocol: TCP
  - name: mesh
    port: 9094
    targetPort: 9094
    protocol: TCP
  selector:
    app: alertmanager
```

## Ingress

```yaml
# alertmanager-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: alertmanager
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: alertmanager
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "alerting"
  annotations:
    kubernetes.io/ingress.class: "nginx"
    cert-manager.io/cluster-issuer: "${CLUSTER_ISSUER}"
    nginx.ingress.kubernetes.io/auth-type: "basic"
    nginx.ingress.kubernetes.io/auth-secret: "alertmanager-basic-auth"
    nginx.ingress.kubernetes.io/auth-realm: "Authentication Required"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-connect-timeout: "60"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "60"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "60"
    nginx.ingress.kubernetes.io/configuration-snippet: |
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-Proto $scheme;
      proxy_set_header X-Tenant-ID "${TENANT_ID}";
      proxy_set_header X-Region-ID "${REGION_ID}";
      proxy_set_header X-Environment "${ENVIRONMENT}";
      
      # Security headers
      add_header X-Content-Type-Options "nosniff" always;
      add_header X-Frame-Options "DENY" always;
      add_header X-XSS-Protection "1; mode=block" always;
      add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; font-src 'self'; object-src 'none'; media-src 'self'; frame-src 'none'; sandbox allow-forms allow-scripts; base-uri 'self'; worker-src 'none';" always;
      add_header Referrer-Policy "strict-origin-when-cross-origin" always;
      add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
      add_header Permissions-Policy "camera=(), microphone=(), geolocation=(), interest-cohort=()" always;
      
      # Custom error pages
      error_page 401 = /401.html;
      error_page 403 = /403.html;
      error_page 404 = /404.html;
spec:
  tls:
  - hosts:
    - alertmanager-${TENANT_ID}-${REGION_ID}.${DOMAIN}
    secretName: alertmanager-tls
  rules:
  - host: alertmanager-${TENANT_ID}-${REGION_ID}.${DOMAIN}
    http:
      paths:
      - path: /alertmanager
        pathType: Prefix
        backend:
          service:
            name: alertmanager
            port:
              number: 9093
```

## NetworkPolicy

```yaml
# alertmanager-networkpolicy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: alertmanager
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: alertmanager
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "alerting"
spec:
  podSelector:
    matchLabels:
      app: alertmanager
  policyTypes:
  - Ingress
  - Egress
  ingress:
  # Permitir tráfego do Ingress Controller
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
      podSelector:
        matchLabels:
          app.kubernetes.io/name: ingress-nginx
    ports:
    - protocol: TCP
      port: 9093
  
  # Permitir tráfego de outros pods do AlertManager (cluster mesh)
  - from:
    - podSelector:
        matchLabels:
          app: alertmanager
    ports:
    - protocol: TCP
      port: 9094
  
  # Permitir tráfego do Prometheus
  - from:
    - namespaceSelector:
        matchLabels:
          innovabiz.com/component: observability
      podSelector:
        matchLabels:
          app: prometheus
    ports:
    - protocol: TCP
      port: 9093
  
  # Permitir tráfego do Grafana
  - from:
    - namespaceSelector:
        matchLabels:
          innovabiz.com/component: observability
      podSelector:
        matchLabels:
          app: grafana
    ports:
    - protocol: TCP
      port: 9093
  
  egress:
  # DNS
  - to:
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: kube-system
      podSelector:
        matchLabels:
          k8s-app: kube-dns
    ports:
    - protocol: UDP
      port: 53
    - protocol: TCP
      port: 53
  
  # SMTP para envio de alertas
  - to:
    - ipBlock:
        cidr: 0.0.0.0/0
        except:
        - 10.0.0.0/8
        - 172.16.0.0/12
        - 192.168.0.0/16
    ports:
    - protocol: TCP
      port: 25
    - protocol: TCP
      port: 465
    - protocol: TCP
      port: 587
  
  # HTTP(S) para webhooks e integrações
  - to:
    - ipBlock:
        cidr: 0.0.0.0/0
        except:
        - 10.0.0.0/8
        - 172.16.0.0/12
        - 192.168.0.0/16
    ports:
    - protocol: TCP
      port: 80
    - protocol: TCP
      port: 443
  
  # Tráfego entre AlertManager (cluster mesh)
  - to:
    - podSelector:
        matchLabels:
          app: alertmanager
    ports:
    - protocol: TCP
      port: 9094
```

## PodDisruptionBudget

```yaml
# alertmanager-pdb.yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: alertmanager
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: alertmanager
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "alerting"
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: alertmanager
```

## ServiceMonitor

```yaml
# alertmanager-servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: alertmanager
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: alertmanager
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "alerting"
spec:
  selector:
    matchLabels:
      app: alertmanager
  namespaceSelector:
    matchNames:
    - observability-${TENANT_ID}-${REGION_ID}
  endpoints:
  - port: http
    interval: 30s
    path: /metrics
    honorLabels: true
    metricRelabelings:
    - sourceLabels: [__meta_kubernetes_pod_label_innovabiz_com_tenant]
      targetLabel: tenant_id
      replacement: "${TENANT_ID}"
    - sourceLabels: [__meta_kubernetes_pod_label_innovabiz_com_region]
      targetLabel: region_id
      replacement: "${REGION_ID}"
    - sourceLabels: [__meta_kubernetes_pod_label_innovabiz_com_environment]
      targetLabel: environment
      replacement: "${ENVIRONMENT}"
```

## Checklist de Validação

A implementação do AlertManager na plataforma INNOVABIZ deve atender aos seguintes critérios:

### Propagação de Contexto Multi-dimensional

- [ ] Todos os recursos têm labels para `tenant_id`, `region_id`, `environment`
- [ ] Configuração de AlertManager inclui multi-tenant e multi-região nos agrupamentos
- [ ] Templates de notificação incluem informações de contexto (tenant, região, ambiente)
- [ ] Rotas de alerta respeitam isolamento multi-tenant
- [ ] Métricas expõem labels de contexto (tenant, região, ambiente)

### Segurança

- [ ] Configuração de TLS via cert-manager para ingress
- [ ] Autenticação básica configurada para acesso via UI
- [ ] Execução como usuário não-root (65534 / nobody)
- [ ] NetworkPolicy implementada limitando comunicação de entrada/saída
- [ ] Secrets armazenando credenciais sensíveis
- [ ] Headers de segurança configurados no Ingress
- [ ] Permissões RBAC mínimas para operação
- [ ] Root filesystem montado como somente leitura
- [ ] Capacidades de contêiner limitadas (drop ALL)

### Alta Disponibilidade

- [ ] Replicação configurada (mínimo 2)
- [ ] Pod anti-affinity para evitar single-point-of-failure
- [ ] PodDisruptionBudget garantindo disponibilidade mínima
- [ ] Readiness e liveness probes configurados
- [ ] RollingUpdate com zero downtime
- [ ] Armazenamento persistente para silenciamentos e estado

### Observabilidade

- [ ] Logging em formato JSON estruturado
- [ ] Métricas Prometheus expostas e configuradas para scraping
- [ ] ServiceMonitor para integração com Prometheus Operator
- [ ] Monitoramento do próprio AlertManager configurado
- [ ] Templates de notificação personalizados para contexto INNOVABIZ
- [ ] Levels de log configuráveis

### Integração

- [ ] Integração com múltiplos canais (email, Slack, PagerDuty, OpsGenie)
- [ ] Templates HTML personalizados para emails
- [ ] Rotas de alerta para diferentes serviços/times
- [ ] Agrupamento inteligente para redução de ruído
- [ ] Regras de inibição para evitar tempestade de alertas
- [ ] Configuração de timeouts e repetição adequados

## Melhores Práticas

### Design de Alertas

1. **Hierarquia de Severidade**: Utilizar consistentemente os níveis de severidade (critical, warning, info) em todos os alertas para priorização adequada.

2. **Agrupamento Inteligente**: Agrupar alertas por dimensões relevantes (tenant, região, ambiente, serviço) para reduzir ruído e facilitar triagem.

3. **Notificações Contextuais**: Incluir informações de contexto suficientes nas notificações para permitir diagnóstico rápido sem necessidade de acessar outras ferramentas.

4. **Silenciamentos Temporários**: Utilizar silenciamentos para manutenções planejadas com duração específica e comentários explicativos.

5. **Documentação de Alertas**: Manter templates de alerta com descrições claras, impacto esperado e ações de remediação sugeridas.

### Operação de Alta Disponibilidade

1. **Cluster Mesh**: Configurar AlertManager em modo de cluster para garantir consistência nas notificações mesmo durante upgrades.

2. **Monitoramento Meta**: Implementar alertas para o próprio AlertManager para garantir sua operacionalidade.

3. **Backup de Configuração**: Manter a configuração do AlertManager versionada em Git e aplicada via CD.

4. **Gerenciamento de Capacidade**: Monitorar uso de recursos e escalar horizontalmente conforme necessário.

5. **Testes de Failover**: Realizar testes regulares de falha para garantir continuidade de notificações.

### Integração com Ecossistema

1. **Dashboards Grafana**: Criar dashboards no Grafana para visualizar métricas de AlertManager (alertas ativos, taxa de notificações, etc).

2. **Exportação de Métricas**: Utilizar ServiceMonitor para exportar métricas de performance do AlertManager para o Prometheus.

3. **Rotas Específicas**: Criar rotas específicas por equipe/serviço/severidade para direcionar alertas aos responsáveis corretos.

4. **Múltiplos Canais**: Configurar múltiplos canais de notificação por rota para redundância.

5. **API Integration**: Utilizar a API do AlertManager para integração com sistemas de ticketing e runbooks automatizados.

### Segurança e Compliance

1. **Autenticação e Autorização**: Implementar autenticação para acesso à UI do AlertManager e RBAC para operações.

2. **Isolamento Multi-Tenant**: Garantir que alertas e rotas de um tenant não sejam visíveis para outros tenants.

3. **Auditoria**: Registrar todas as operações de gerenciamento de alertas (criação, silenciamento, resolução).

4. **Segurança de Rede**: Implementar NetworkPolicy para limitar comunicação apenas para serviços necessários.

5. **Gestão de Segredos**: Armazenar tokens e chaves de API em secrets Kubernetes com rotação regular.

## Exemplos de Uso

### Variáveis de Ambiente

Para ambientes de produção no Brasil para o tenant principal:

```bash
export TENANT_ID="primary"
export REGION_ID="br"
export ENVIRONMENT="production"
export DOMAIN="observability.innovabiz.com"
export CLUSTER_ISSUER="letsencrypt-prod"
export STORAGE_CLASS="ssd-replicated"
export ALERTMANAGER_STORAGE_SIZE="10Gi"
export ALERTMANAGER_VERSION="v0.25.0"
export ALERTMANAGER_REPLICAS="2"
export ALERTMANAGER_CPU_REQUEST="100m"
export ALERTMANAGER_MEMORY_REQUEST="256Mi"
export ALERTMANAGER_CPU_LIMIT="500m"
export ALERTMANAGER_MEMORY_LIMIT="512Mi"
export LOG_LEVEL="info"
export REPEAT_INTERVAL="4h"
export SMTP_SERVER="smtp.innovabiz.com"
export SMTP_PORT="587"
export SMTP_FROM="alertmanager@innovabiz.com"
export SMTP_USER="alertmanager@innovabiz.com"
export SMTP_PASSWORD="SECURE_PASSWORD_HERE"
export OPS_EMAIL="ops@innovabiz.com"
export WARNING_EMAIL="alerts@innovabiz.com"
export INFRA_EMAIL="infra@innovabiz.com"
export DB_EMAIL="db@innovabiz.com"
export APP_EMAIL="app@innovabiz.com"
export BUSINESS_EMAIL="business@innovabiz.com"
export SLACK_API_URL="https://hooks.slack.com/services/XXXX/YYYY/ZZZZ"
export PAGERDUTY_SERVICE_KEY="PAGERDUTY_KEY_HERE"
export OPSGENIE_KEY="OPSGENIE_KEY_HERE"
```

### Criar Secrets de Autenticação

```bash
# Criar secret para autenticação básica no Ingress
kubectl create secret generic alertmanager-basic-auth \
  --namespace=observability-${TENANT_ID}-${REGION_ID} \
  --from-literal=auth=$(htpasswd -bn admin "${ADMIN_PASSWORD}")
```

### Aplicar Configuração

```bash
# Aplicar todos os manifestos usando envsubst para substituição de variáveis
for file in namespace.yaml alertmanager-configmap.yaml alertmanager-secrets.yaml alertmanager-pvc.yaml alertmanager-rbac.yaml alertmanager-deployment.yaml alertmanager-service.yaml alertmanager-ingress.yaml alertmanager-networkpolicy.yaml alertmanager-pdb.yaml alertmanager-servicemonitor.yaml; do
  envsubst < $file | kubectl apply -f -
done
```

### Validar Implantação

```bash
# Verificar status dos pods
kubectl get pods -n observability-${TENANT_ID}-${REGION_ID} -l app=alertmanager

# Verificar logs
kubectl logs -n observability-${TENANT_ID}-${REGION_ID} -l app=alertmanager

# Testar API
kubectl port-forward -n observability-${TENANT_ID}-${REGION_ID} svc/alertmanager 9093:9093
curl http://localhost:9093/api/v2/status
```

### Integração com Prometheus

Para configurar o Prometheus para enviar alertas para o AlertManager:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
  namespace: observability-${TENANT_ID}-${REGION_ID}
data:
  prometheus.yml: |
    # ... outras configurações ...
    
    alerting:
      alertmanagers:
      - kubernetes_sd_configs:
        - role: endpoints
          namespaces:
            names:
            - observability-${TENANT_ID}-${REGION_ID}
        relabel_configs:
        - source_labels: [__meta_kubernetes_service_name]
          regex: alertmanager
          action: keep
        - source_labels: [__meta_kubernetes_namespace]
          target_label: kubernetes_namespace
        - source_labels: [__meta_kubernetes_service_name]
          target_label: kubernetes_name
        - source_labels: [__meta_kubernetes_pod_label_innovabiz_com_tenant]
          target_label: tenant_id
        - source_labels: [__meta_kubernetes_pod_label_innovabiz_com_region]
          target_label: region_id
        - source_labels: [__meta_kubernetes_pod_label_innovabiz_com_environment]
          target_label: environment
```

### Integração com Grafana

Para adicionar o AlertManager como fonte de dados no Grafana:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-datasources
  namespace: observability-${TENANT_ID}-${REGION_ID}
data:
  alertmanager.yaml: |
    apiVersion: 1
    datasources:
    - name: AlertManager
      type: alertmanager
      url: http://alertmanager.observability-${TENANT_ID}-${REGION_ID}.svc.cluster.local:9093
      access: proxy
      isDefault: false
      jsonData:
        implementation: prometheus
      version: 1
      editable: false
```