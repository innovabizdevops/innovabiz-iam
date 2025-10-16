# Template Kubernetes YAML para Grafana - INNOVABIZ

## Visão Geral

Este documento fornece templates YAML para implantação do Grafana no ambiente Kubernetes da plataforma INNOVABIZ. Estes templates seguem as melhores práticas de segurança, dimensionamento e configuração multi-dimensional conforme os padrões INNOVABIZ.

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
10. [HorizontalPodAutoscaler](#horizontalpodautoscaler)
11. [PodDisruptionBudget](#poddisruptionbudget)
12. [ServiceMonitor](#servicemonitor)
13. [Checklist de Validação](#checklist-de-validação)
14. [Melhores Práticas](#melhores-práticas)
15. [Exemplos de Uso](#exemplos-de-uso)

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
    innovabiz.com/module: "visualization"
```

## ConfigMap

```yaml
# grafana-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-config
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: grafana
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "visualization"
data:
  grafana.ini: |
    [analytics]
    check_for_updates = false
    [grafana_net]
    url = https://grafana.net
    [log]
    mode = console
    level = info
    [paths]
    data = /var/lib/grafana/
    logs = /var/log/grafana
    plugins = /var/lib/grafana/plugins
    provisioning = /etc/grafana/provisioning
    [server]
    domain = grafana-${TENANT_ID}-${REGION_ID}.${DOMAIN}
    root_url = https://grafana-${TENANT_ID}-${REGION_ID}.${DOMAIN}
    serve_from_sub_path = false
    [security]
    cookie_secure = true
    cookie_samesite = lax
    disable_gravatar = true
    content_security_policy = true
    content_security_policy_template = "default-src 'self'; script-src 'self' 'unsafe-eval' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self' https://grafana.com"
    [users]
    allow_sign_up = false
    auto_assign_org = true
    auto_assign_org_role = Editor
    [auth]
    disable_login_form = false
    [auth.proxy]
    enabled = true
    header_name = X-WEBAUTH-USER
    header_property = username
    auto_sign_up = true
    [auth.basic]
    enabled = false
    [auth.oauth]
    enabled = true
    [dashboards]
    versions_to_keep = 20
    min_refresh_interval = 10s
    [dashboards.json]
    enabled = true
    path = /etc/grafana/provisioning/dashboards
    [unified_alerting]
    enabled = true
    [alerting]
    enabled = false
    [feature_toggles]
    enable = tempoSearch tempoBackendSearch tempoServiceGraph
    [database]
    wal = true
  
  # Datasources
  datasources.yaml: |
    apiVersion: 1
    datasources:
      - name: Prometheus
        type: prometheus
        access: proxy
        url: http://prometheus-${TENANT_ID}-${REGION_ID}.observability-${TENANT_ID}-${REGION_ID}.svc.cluster.local:9090
        isDefault: true
        jsonData:
          timeInterval: 30s
          queryTimeout: 120s
          httpMethod: POST
          exemplarTraceIdDestinations:
            - name: traceID
              datasourceUid: tempo
        secureJsonData:
          httpHeaderValue1: "${PROM_AUTH_TOKEN}"
        editable: false
        
      - name: Loki
        type: loki
        access: proxy
        url: http://loki-${TENANT_ID}-${REGION_ID}.observability-${TENANT_ID}-${REGION_ID}.svc.cluster.local:3100
        jsonData:
          derivedFields:
            - datasourceUid: tempo
              matcherRegex: "traceID=(\\w+)"
              name: TraceID
              url: "$${__value.raw}"
        editable: false

      - name: Tempo
        type: tempo
        access: proxy
        url: http://tempo-${TENANT_ID}-${REGION_ID}.observability-${TENANT_ID}-${REGION_ID}.svc.cluster.local:3200
        uid: tempo
        jsonData:
          nodeGraph:
            enabled: true
          serviceMap:
            datasourceUid: prometheus
          tracesToLogs:
            datasourceUid: loki
            mapTagNamesEnabled: true
            mappedTags:
              - key: service.name
                value: service
              - key: tenant.id
                value: tenant_id
              - key: region.id
                value: region_id
              - key: environment
                value: environment
        editable: false
  
  # Dashboard providers
  dashboard-providers.yaml: |
    apiVersion: 1
    providers:
      - name: 'INNOVABIZ Dashboards'
        orgId: 1
        folder: 'INNOVABIZ'
        type: file
        disableDeletion: true
        updateIntervalSeconds: 30
        allowUiUpdates: false
        options:
          path: /etc/grafana/provisioning/dashboards
          foldersFromFilesStructure: true

  # Dashboards
  platform-overview-dashboard.json: |
    {
      "annotations": {
        "list": [
          {
            "builtIn": 1,
            "datasource": {
              "type": "grafana",
              "uid": "-- Grafana --"
            },
            "enable": true,
            "hide": true,
            "iconColor": "rgba(0, 211, 255, 1)",
            "name": "Annotations & Alerts",
            "type": "dashboard"
          }
        ]
      },
      "editable": true,
      "fiscalYearStartMonth": 0,
      "graphTooltip": 0,
      "id": 1,
      "links": [],
      "liveNow": false,
      "panels": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "prometheus"
          },
          "fieldConfig": {
            "defaults": {
              "color": {
                "mode": "palette-classic"
              },
              "custom": {
                "axisCenteredZero": false,
                "axisColorMode": "text",
                "axisLabel": "",
                "axisPlacement": "auto",
                "barAlignment": 0,
                "drawStyle": "line",
                "fillOpacity": 10,
                "gradientMode": "none",
                "hideFrom": {
                  "legend": false,
                  "tooltip": false,
                  "viz": false
                },
                "insertNulls": false,
                "lineInterpolation": "smooth",
                "lineWidth": 2,
                "pointSize": 5,
                "scaleDistribution": {
                  "type": "linear"
                },
                "showPoints": "never",
                "spanNulls": false,
                "stacking": {
                  "group": "A",
                  "mode": "none"
                },
                "thresholdsStyle": {
                  "mode": "off"
                }
              },
              "mappings": [],
              "thresholds": {
                "mode": "absolute",
                "steps": [
                  {
                    "color": "green",
                    "value": null
                  },
                  {
                    "color": "red",
                    "value": 80
                  }
                ]
              },
              "unit": "reqps"
            },
            "overrides": []
          },
          "gridPos": {
            "h": 9,
            "w": 12,
            "x": 0,
            "y": 0
          },
          "id": 1,
          "options": {
            "legend": {
              "calcs": [
                "mean",
                "lastNotNull",
                "max"
              ],
              "displayMode": "table",
              "placement": "bottom",
              "showLegend": true
            },
            "tooltip": {
              "mode": "multi",
              "sort": "none"
            }
          },
          "pluginVersion": "10.0.3",
          "targets": [
            {
              "datasource": {
                "type": "prometheus",
                "uid": "prometheus"
              },
              "editorMode": "code",
              "expr": "sum by(service) (rate(http_server_requests_seconds_count{tenant_id=\"${tenant_id}\",region_id=\"${region_id}\",environment=\"${environment}\"}[5m]))",
              "instant": false,
              "legendFormat": "{{service}}",
              "range": true,
              "refId": "A"
            }
          ],
          "title": "Requisições HTTP por Serviço",
          "type": "timeseries"
        },
        {
          "datasource": {
            "type": "prometheus",
            "uid": "prometheus"
          },
          "fieldConfig": {
            "defaults": {
              "color": {
                "mode": "thresholds"
              },
              "mappings": [],
              "thresholds": {
                "mode": "absolute",
                "steps": [
                  {
                    "color": "green",
                    "value": null
                  },
                  {
                    "color": "yellow",
                    "value": 200
                  },
                  {
                    "color": "orange",
                    "value": 500
                  },
                  {
                    "color": "red",
                    "value": 1000
                  }
                ]
              },
              "unit": "ms"
            },
            "overrides": []
          },
          "gridPos": {
            "h": 9,
            "w": 12,
            "x": 12,
            "y": 0
          },
          "id": 2,
          "options": {
            "colorMode": "value",
            "graphMode": "area",
            "justifyMode": "auto",
            "orientation": "auto",
            "reduceOptions": {
              "calcs": [
                "mean"
              ],
              "fields": "",
              "values": false
            },
            "textMode": "auto"
          },
          "pluginVersion": "10.0.3",
          "targets": [
            {
              "datasource": {
                "type": "prometheus",
                "uid": "prometheus"
              },
              "editorMode": "code",
              "expr": "sum(rate(http_server_requests_seconds_sum{tenant_id=\"${tenant_id}\",region_id=\"${region_id}\",environment=\"${environment}\"}[5m])) / sum(rate(http_server_requests_seconds_count{tenant_id=\"${tenant_id}\",region_id=\"${region_id}\",environment=\"${environment}\"}[5m])) * 1000",
              "instant": false,
              "legendFormat": "__auto",
              "range": true,
              "refId": "A"
            }
          ],
          "title": "Latência Média HTTP (ms)",
          "type": "stat"
        }
      ],
      "refresh": "10s",
      "schemaVersion": 38,
      "style": "dark",
      "tags": [
        "innovabiz",
        "overview"
      ],
      "templating": {
        "list": [
          {
            "current": {
              "selected": false,
              "text": "tenant1",
              "value": "tenant1"
            },
            "datasource": {
              "type": "prometheus",
              "uid": "prometheus"
            },
            "definition": "label_values(tenant_id)",
            "hide": 0,
            "includeAll": false,
            "label": "Tenant",
            "multi": false,
            "name": "tenant_id",
            "options": [],
            "query": {
              "query": "label_values(tenant_id)",
              "refId": "StandardVariableQuery"
            },
            "refresh": 1,
            "regex": "",
            "skipUrlSync": false,
            "sort": 0,
            "type": "query"
          },
          {
            "current": {
              "selected": false,
              "text": "br",
              "value": "br"
            },
            "datasource": {
              "type": "prometheus",
              "uid": "prometheus"
            },
            "definition": "label_values(region_id)",
            "hide": 0,
            "includeAll": false,
            "label": "Região",
            "multi": false,
            "name": "region_id",
            "options": [],
            "query": {
              "query": "label_values(region_id)",
              "refId": "StandardVariableQuery"
            },
            "refresh": 1,
            "regex": "",
            "skipUrlSync": false,
            "sort": 0,
            "type": "query"
          },
          {
            "current": {
              "selected": false,
              "text": "production",
              "value": "production"
            },
            "datasource": {
              "type": "prometheus",
              "uid": "prometheus"
            },
            "definition": "label_values(environment)",
            "hide": 0,
            "includeAll": false,
            "label": "Ambiente",
            "multi": false,
            "name": "environment",
            "options": [],
            "query": {
              "query": "label_values(environment)",
              "refId": "StandardVariableQuery"
            },
            "refresh": 1,
            "regex": "",
            "skipUrlSync": false,
            "sort": 0,
            "type": "query"
          }
        ]
      },
      "time": {
        "from": "now-6h",
        "to": "now"
      },
      "timepicker": {},
      "timezone": "",
      "title": "INNOVABIZ Platform Overview",
      "uid": "innovabiz-platform-overview",
      "version": 1,
      "weekStart": ""
    }
```

## Secret

```yaml
# grafana-secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: grafana-secrets
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: grafana
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "visualization"
type: Opaque
stringData:
  admin-user: "${GRAFANA_ADMIN_USER}"
  admin-password: "${GRAFANA_ADMIN_PASSWORD}"
  prometheus-token: "${PROM_AUTH_TOKEN}"
```

## PersistentVolumeClaim

```yaml
# grafana-pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: grafana-storage
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: grafana
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "visualization"
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: "${STORAGE_CLASS_NAME}"
  resources:
    requests:
      storage: 10Gi
```

## ServiceAccount e RBAC

```yaml
# grafana-rbac.yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: grafana
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: grafana
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "visualization"
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: grafana
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: grafana
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "visualization"
rules:
- apiGroups: [""]
  resources: ["configmaps", "secrets"]
  verbs: ["get", "watch", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: grafana
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: grafana
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "visualization"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: grafana
subjects:
- kind: ServiceAccount
  name: grafana
  namespace: observability-${TENANT_ID}-${REGION_ID}
```## Deployment

```yaml
# grafana-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: grafana
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "visualization"
spec:
  replicas: ${GRAFANA_REPLICAS}  # Ajuste baseado no ambiente (1 para dev/staging, 2+ para production)
  revisionHistoryLimit: 3
  selector:
    matchLabels:
      app: grafana
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
  template:
    metadata:
      labels:
        app: grafana
        innovabiz.com/tenant: "${TENANT_ID}"
        innovabiz.com/region: "${REGION_ID}"
        innovabiz.com/environment: "${ENVIRONMENT}"
        innovabiz.com/component: "observability"
        innovabiz.com/module: "visualization"
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "3000"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: grafana
      securityContext:
        fsGroup: 472
        supplementalGroups:
          - 0
      containers:
        - name: grafana
          image: grafana/grafana:${GRAFANA_VERSION}  # Use versões específicas, ex: 10.0.3
          imagePullPolicy: IfNotPresent
          ports:
            - name: http
              containerPort: 3000
              protocol: TCP
          env:
            - name: GF_SECURITY_ADMIN_USER
              valueFrom:
                secretKeyRef:
                  name: grafana-secrets
                  key: admin-user
            - name: GF_SECURITY_ADMIN_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: grafana-secrets
                  key: admin-password
            - name: GF_PATHS_CONFIG
              value: /etc/grafana/grafana.ini
            - name: GF_PATHS_DATA
              value: /var/lib/grafana
            - name: GF_PATHS_HOME
              value: /usr/share/grafana
            - name: GF_PATHS_LOGS
              value: /var/log/grafana
            - name: GF_PATHS_PLUGINS
              value: /var/lib/grafana/plugins
            - name: GF_PATHS_PROVISIONING
              value: /etc/grafana/provisioning
            - name: TENANT_ID
              value: "${TENANT_ID}"
            - name: REGION_ID
              value: "${REGION_ID}"
            - name: ENVIRONMENT
              value: "${ENVIRONMENT}"
          securityContext:
            runAsNonRoot: true
            runAsUser: 472
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
            readOnlyRootFilesystem: true
          volumeMounts:
            - name: config
              mountPath: /etc/grafana/grafana.ini
              subPath: grafana.ini
            - name: storage
              mountPath: /var/lib/grafana
            - name: provisioning-datasources
              mountPath: /etc/grafana/provisioning/datasources
            - name: provisioning-dashboards
              mountPath: /etc/grafana/provisioning/dashboards
            - name: dashboards
              mountPath: /etc/grafana/dashboards
          livenessProbe:
            httpGet:
              path: /api/health
              port: http
            initialDelaySeconds: 60
            timeoutSeconds: 30
            failureThreshold: 10
          readinessProbe:
            httpGet:
              path: /api/health
              port: http
            initialDelaySeconds: 10
            timeoutSeconds: 10
          resources:
            limits:
              cpu: ${GRAFANA_CPU_LIMIT}  # ex: 1000m
              memory: ${GRAFANA_MEMORY_LIMIT}  # ex: 1Gi
            requests:
              cpu: ${GRAFANA_CPU_REQUEST}  # ex: 100m
              memory: ${GRAFANA_MEMORY_REQUEST}  # ex: 256Mi
      volumes:
        - name: config
          configMap:
            name: grafana-config
            items:
              - key: grafana.ini
                path: grafana.ini
        - name: storage
          persistentVolumeClaim:
            claimName: grafana-storage
        - name: provisioning-datasources
          configMap:
            name: grafana-config
            items:
              - key: datasources.yaml
                path: datasources.yaml
        - name: provisioning-dashboards
          configMap:
            name: grafana-config
            items:
              - key: dashboard-providers.yaml
                path: dashboard-providers.yaml
        - name: dashboards
          configMap:
            name: grafana-config
            items:
              - key: platform-overview-dashboard.json
                path: platform-overview.json
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
                        - grafana
                topologyKey: kubernetes.io/hostname
      nodeSelector:
        kubernetes.io/os: linux
      terminationGracePeriodSeconds: 30
```

## Service

```yaml
# grafana-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: grafana
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: grafana
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "visualization"
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "3000"
spec:
  type: ClusterIP
  ports:
    - port: 80
      targetPort: http
      protocol: TCP
      name: http
  selector:
    app: grafana
```

## Ingress

```yaml
# grafana-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: grafana
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: grafana
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "visualization"
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-connect-timeout: "30"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "600"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "600"
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
    cert-manager.io/cluster-issuer: "${CLUSTER_ISSUER}"
spec:
  tls:
    - hosts:
        - grafana-${TENANT_ID}-${REGION_ID}.${DOMAIN}
      secretName: grafana-tls
  rules:
    - host: grafana-${TENANT_ID}-${REGION_ID}.${DOMAIN}
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: grafana
                port:
                  name: http
```

## NetworkPolicy

```yaml
# grafana-network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: grafana
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: grafana
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "visualization"
spec:
  podSelector:
    matchLabels:
      app: grafana
  policyTypes:
    - Ingress
    - Egress
  ingress:
    # Permitir tráfego de entrada para o serviço Grafana
    - ports:
        - port: 3000
          protocol: TCP
      from:
        # Permitir tráfego de ingress-controllers
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: ingress-nginx
        # Permitir tráfego de pods no mesmo namespace
        - podSelector: {}
  egress:
    # Permitir acesso ao Prometheus
    - to:
        - podSelector:
            matchLabels:
              app: prometheus
      ports:
        - port: 9090
          protocol: TCP
    # Permitir acesso ao Loki
    - to:
        - podSelector:
            matchLabels:
              app: loki
      ports:
        - port: 3100
          protocol: TCP
    # Permitir acesso ao Tempo
    - to:
        - podSelector:
            matchLabels:
              app: tempo
      ports:
        - port: 3200
          protocol: TCP
    # Permitir acesso a Servidores DNS
    - to:
        - namespaceSelector: {}
          podSelector:
            matchLabels:
              k8s-app: kube-dns
      ports:
        - port: 53
          protocol: UDP
    # Permitir acesso para HTTPS externo (atualizações, plugins)
    - to:
        - ipBlock:
            cidr: 0.0.0.0/0
      ports:
        - port: 443
          protocol: TCP
```

## HorizontalPodAutoscaler

```yaml
# grafana-hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: grafana
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: grafana
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "visualization"
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: grafana
  minReplicas: ${GRAFANA_MIN_REPLICAS}  # ex: 2 para produção
  maxReplicas: ${GRAFANA_MAX_REPLICAS}  # ex: 5 para produção
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
        - type: Pods
          value: 1
          periodSeconds: 30
        - type: Percent
          value: 100
          periodSeconds: 60
```

## PodDisruptionBudget

```yaml
# grafana-pdb.yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: grafana
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: grafana
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "visualization"
spec:
  minAvailable: 1  # Sempre manter pelo menos 1 pod disponível
  selector:
    matchLabels:
      app: grafana
```

## ServiceMonitor

```yaml
# grafana-service-monitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: grafana
  namespace: observability-${TENANT_ID}-${REGION_ID}
  labels:
    app: grafana
    innovabiz.com/tenant: "${TENANT_ID}"
    innovabiz.com/region: "${REGION_ID}"
    innovabiz.com/environment: "${ENVIRONMENT}"
    innovabiz.com/component: "observability"
    innovabiz.com/module: "visualization"
spec:
  selector:
    matchLabels:
      app: grafana
  endpoints:
    - port: http
      path: /metrics
      interval: 30s
      scrapeTimeout: 10s
      honorLabels: true
      metricRelabelings:
        - sourceLabels: [__name__]
          regex: go_.*
          action: drop
        - targetLabel: tenant_id
          replacement: "${TENANT_ID}"
        - targetLabel: region_id
          replacement: "${REGION_ID}"
        - targetLabel: environment
          replacement: "${ENVIRONMENT}"
```

## Checklist de Validação

Use o checklist abaixo para validar a implementação do Grafana na plataforma INNOVABIZ:

### Multi-dimensionalidade

- [ ] **Isolamento por Tenant**
  - [ ] Namespace específico por tenant e região
  - [ ] Labels de tenant/região em todos os recursos
  - [ ] Variáveis do Grafana configuradas para filtrar por tenant/região
  
- [ ] **Contexto Regional**
  - [ ] URLs e domínios específicos por região
  - [ ] Integração com fontes de dados regionais
  - [ ] Perfis de acesso específicos por região

- [ ] **Separação de Ambientes**
  - [ ] Configurações e recursos adequados ao ambiente
  - [ ] Políticas de segurança específicas por ambiente
  - [ ] Visualização clara do ambiente nos dashboards

### Segurança

- [ ] **Autenticação e Autorização**
  - [ ] Integração SSO configurada (OAuth/SAML)
  - [ ] RBAC implementado com papéis apropriados
  - [ ] Segredos armazenados adequadamente (não hard-coded)

- [ ] **Proteção de Rede**
  - [ ] Ingress configurado com TLS
  - [ ] NetworkPolicy limitando comunicação
  - [ ] Acesso exposto apenas quando necessário

- [ ] **Hardening de Containers**
  - [ ] Container não-root
  - [ ] Sistema de arquivos somente leitura
  - [ ] Capacidades mínimas necessárias
  - [ ] Recursos adequadamente limitados

### Conformidade com Padrões INNOVABIZ

- [ ] **Nomenclatura**
  - [ ] URLs seguem padrão innovabiz.com
  - [ ] Prefixos/sufixos adequados para recursos
  - [ ] Dasboards seguem convenção de nomes

- [ ] **Instrumentação**
  - [ ] Métricas próprias do Grafana exportadas
  - [ ] Labels padrão em todas as métricas
  - [ ] Integrações configuradas com outras ferramentas

- [ ] **Documentação**
  - [ ] README.md completo com instruções
  - [ ] Dashboards documentados
  - [ ] Runbooks de manutenção

## Melhores Práticas

### Dashboards

1. **Organização de Dashboards**
   - Estruture dashboards em pastas por funcionalidade/módulo
   - Use variáveis de template para tenant, região, ambiente
   - Padronize unidades e escalas entre dashboards similares

2. **Design de Dashboards**
   - Forneça visão "de cima para baixo" (overview para detalhes)
   - Inclua links para documentação e runbooks
   - Utilize cores consistentes para severidades e estados

3. **Performance e Usabilidade**
   - Otimize queries Prometheus para eficiência
   - Configure intervalos de atualização adequados
   - Use painéis reutilizáveis para elementos comuns

### Alta Disponibilidade e Escalabilidade

1. **Arquitetura**
   - Use múltiplas réplicas em produção
   - Configure anti-afinidade para distribuir pods
   - Implemente HPA para escalar conforme necessário

2. **Armazenamento**
   - Use PostgreSQL externo para instalações de grande porte
   - Considere soluções de armazenamento em nuvem para persistência
   - Implemente backup regular de configurações e dashboards

3. **Rede**
   - Otimize timeouts para queries de longa duração
   - Configure adequadamente limites de conexão
   - Utilize balanceamento de carga para distribuição de tráfego

### Monitoramento do Próprio Grafana

1. **Métricas Chave**
   - Monitore uso de CPU/memória
   - Acompanhe tempos de resposta da API
   - Monitore contagem de sessões ativas

2. **Alertas**
   - Configure alertas para indisponibilidade
   - Monitore erros de datasource
   - Alerte sobre uso elevado de recursos

3. **Logging**
   - Envie logs para plataforma centralizada
   - Defina níveis de log apropriados por ambiente
   - Implemente rastreamento para troubleshooting

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

# Defina as variáveis do Grafana
export GRAFANA_VERSION="10.0.3"
export GRAFANA_REPLICAS="2"
export GRAFANA_CPU_LIMIT="1000m"
export GRAFANA_MEMORY_LIMIT="1Gi"
export GRAFANA_CPU_REQUEST="100m"
export GRAFANA_MEMORY_REQUEST="256Mi"
export GRAFANA_MIN_REPLICAS="2"
export GRAFANA_MAX_REPLICAS="5"

# Defina credenciais (use geração segura em produção)
export GRAFANA_ADMIN_USER="admin"
export GRAFANA_ADMIN_PASSWORD="$(openssl rand -base64 16)"
export PROM_AUTH_TOKEN="$(openssl rand -base64 32)"
```

2. **Substitua as Variáveis e Aplique os Templates**

```bash
# Crie um diretório temporário
mkdir -p /tmp/grafana-deploy

# Copie e substitua variáveis em todos os arquivos
for file in namespace.yaml grafana-configmap.yaml grafana-secrets.yaml grafana-pvc.yaml grafana-rbac.yaml \
            grafana-deployment.yaml grafana-service.yaml grafana-ingress.yaml grafana-network-policy.yaml \
            grafana-hpa.yaml grafana-pdb.yaml grafana-service-monitor.yaml; do
  envsubst < $file > /tmp/grafana-deploy/$file
done

# Aplique os recursos na ordem correta
kubectl apply -f /tmp/grafana-deploy/namespace.yaml
kubectl apply -f /tmp/grafana-deploy/grafana-configmap.yaml
kubectl apply -f /tmp/grafana-deploy/grafana-secrets.yaml
kubectl apply -f /tmp/grafana-deploy/grafana-pvc.yaml
kubectl apply -f /tmp/grafana-deploy/grafana-rbac.yaml
kubectl apply -f /tmp/grafana-deploy/grafana-deployment.yaml
kubectl apply -f /tmp/grafana-deploy/grafana-service.yaml
kubectl apply -f /tmp/grafana-deploy/grafana-network-policy.yaml
kubectl apply -f /tmp/grafana-deploy/grafana-hpa.yaml
kubectl apply -f /tmp/grafana-deploy/grafana-pdb.yaml
kubectl apply -f /tmp/grafana-deploy/grafana-service-monitor.yaml
kubectl apply -f /tmp/grafana-deploy/grafana-ingress.yaml

# Limpe os arquivos temporários com credenciais
rm -rf /tmp/grafana-deploy
```

3. **Verifique a Implantação**

```bash
# Verifique se todos os recursos foram criados corretamente
kubectl -n observability-${TENANT_ID}-${REGION_ID} get all -l app=grafana

# Verifique se o pod do Grafana está em execução
kubectl -n observability-${TENANT_ID}-${REGION_ID} get pods -l app=grafana

# Obtenha a URL de acesso
echo "Acesse o Grafana em: https://grafana-${TENANT_ID}-${REGION_ID}.${DOMAIN}"

# Obtenha a senha do admin (apenas para ambientes não-produtivos)
kubectl -n observability-${TENANT_ID}-${REGION_ID} get secret grafana-secrets -o jsonpath="{.data.admin-password}" | base64 --decode
```