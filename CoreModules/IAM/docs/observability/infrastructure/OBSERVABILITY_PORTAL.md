# OBSERVABILITY PORTAL - Interface Unificada de Observabilidade

## 1. Visão Geral

O Observability Portal é a interface centralizada de observabilidade da plataforma INNOVABIZ, agregando dados de logs, métricas e traces de todos os componentes da infraestrutura em uma única visualização coerente. Este documento detalha a implementação, configuração e utilização do Portal como ponto central para monitoramento, troubleshooting e análise do IAM Audit Service.

### 1.1 Função na Arquitetura de Observabilidade

O Observability Portal atua como:

- **Interface Unificada**: Single pane of glass para toda a stack de observabilidade
- **Agregador Multi-Fonte**: Integração com Prometheus, Elasticsearch, Loki, Grafana e Kibana
- **Gateway de Observabilidade**: Ponto de entrada único para todas as ferramentas
- **Federação de Contexto**: Propagação de contexto entre ferramentas integradas
- **Plataforma de Alerta Centralizada**: Unificação de alertas de múltiplas fontes
- **Centro de Operações**: Interface para diagnóstico e resolução de incidentes

### 1.2 Recursos e Capacidades

- **Dashboard Unificado**: Visão consolidada de saúde, performance e status
- **Navegação Contextual**: Drill-down entre logs, métricas e traces preservando contexto
- **Correlação Automática**: Relacionamento entre eventos de diferentes fontes
- **Alertas Centralizados**: Consolidação e gestão de alertas de múltiplos sistemas
- **Visualização Multi-Dimensional**: Adaptação automática ao contexto (tenant, região, ambiente)
- **Service Map**: Visualização de topologia e dependências entre serviços
- **Health Status**: Status de saúde em tempo real de todos os componentes
- **Analytics**: Análise de tendências e anomalias

## 2. Arquitetura de Implementação

### 2.1 Diagrama de Arquitetura

```
                          ┌────────────────────────────────┐
                          │        Load Balancer           │
                          │  (portal.innovabiz.cloud)      │
                          └──────────────┬─────────────────┘
                                         │
                                         ▼
                          ┌────────────────────────────────┐
                          │          OIDC Proxy            │
                          │    (Auth + Context Injection)   │
                          └──────────────┬─────────────────┘
                                         │
                                         ▼
           ┌──────────────────────────────────────────────────────────────┐
           │                  Observability Portal                         │
           │                                                              │
           │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐           │
           │  │ Portal UI   │  │ Backend API │  │ Federation  │           │
           │  │ (React/     │  │ (Node.js)   │  │ Service     │           │
           │  │  TypeScript)│  │             │  │             │           │
           │  └─────────────┘  └─────────────┘  └─────────────┘           │
           │         │               │                │                   │
           └─────────┼───────────────┼────────────────┼───────────────────┘
                     │               │                │
        ┌────────────┼───────────────┼────────────────┼───────────────────┐
        │            │               │                │                   │
        │   ┌────────▼─────┐  ┌──────▼───────┐  ┌─────▼────────┐  ┌───────▼───────┐
        │   │              │  │              │  │              │  │               │
        │   │  Grafana     │  │  Kibana      │  │ Prometheus   │  │ Jaeger/Tempo  │
        │   │  (Métricas/  │  │  (Logs/      │  │ (Métricas)   │  │ (Traces)      │
        │   │   Loki)      │  │   Elastic)   │  │              │  │               │
        │   │              │  │              │  │              │  │               │
        │   └──────────────┘  └──────────────┘  └──────────────┘  └───────────────┘
        │            │               │                │                   │         
        └────────────┼───────────────┼────────────────┼───────────────────┘         
                     │               │                │                            
        ┌────────────┼───────────────┼────────────────┼───────────────────┐         
        │   ┌────────▼─────┐  ┌──────▼───────┐  ┌─────▼────────┐  ┌───────▼───────┐
        │   │              │  │              │  │              │  │               │
        │   │  Loki        │  │ Elasticsearch │  │ Prometheus   │  │ Tempo/Jaeger  │
        │   │  (Logs)      │  │  (Logs)       │  │ (Storage)    │  │ (Storage)     │
        │   │              │  │              │  │              │  │               │
        │   └──────────────┘  └──────────────┘  └──────────────┘  └───────────────┘
        │                                                                         │
        └─────────────────────────────────────────────────────────────────────────┘
```

### 2.2 Componentes Principais

#### 2.2.1 Portal UI (Frontend)

- **Framework**: React 18 com TypeScript
- **Design System**: Material UI com tema customizado INNOVABIZ
- **Estado**: Redux Toolkit + RTK Query
- **Visualizações**: D3.js, ECharts, React Flow
- **Internacionalização**: i18n com suporte a pt-BR, en-US, pt-AO, es-ES
- **Comunicação**: REST API + GraphQL + WebSockets para atualizações em tempo real

#### 2.2.2 Backend API

- **Runtime**: Node.js 18+ com Express
- **API**: REST + GraphQL (Apollo Server)
- **Cache**: Redis para cache distribuído e sessão
- **Comunicação**: Axios para integração com outros sistemas
- **Autenticação**: Passport.js com estratégias OIDC e JWT
- **Logging**: Winston com formato estruturado (JSON)

#### 2.2.3 Federation Service

- **Agregação**: Agregador de dados de múltiplas fontes
- **Tradução**: Normalização de formatos entre sistemas
- **Contexto**: Propagação de contexto multi-dimensional
- **Correlação**: Identificação de relações entre eventos
- **Proxy Reverso**: Reescrita e encaminhamento de requisições

### 2.3 Implantação Kubernetes

O Observability Portal é implantado como um conjunto de Deployments Kubernetes:

- **Namespace**: `innovabiz-observability`
- **Deployments**:
  - `portal-frontend`: UI React (3 réplicas)
  - `portal-backend`: API Node.js (3 réplicas)
  - `portal-federation`: Serviço de Federação (2 réplicas)
- **Services**: ClusterIP internos + Ingress para acesso externo
- **Pods**: Configuração anti-affinity para alta disponibilidade
- **TLS**: Certificados gerenciados via cert-manager
- **ConfigMaps**: Configuração externalizada e versionada no Git

### 2.4 Especificação de Recursos

| Componente | Réplicas | CPU (Request/Limit) | Memória (Request/Limit) | Armazenamento |
|------------|----------|---------------------|-------------------------|--------------|
| Portal Frontend | 3 | 0.5 CPU / 1 CPU | 512Mi / 1Gi | N/A |
| Portal Backend | 3 | 1 CPU / 2 CPU | 1Gi / 2Gi | N/A |
| Federation Service | 2 | 1 CPU / 2 CPU | 1Gi / 2Gi | N/A |
| Redis Cache | 3 | 0.5 CPU / 1 CPU | 1Gi / 2Gi | 10Gi (PVC) |
| Ingress | 2 | 0.5 CPU / 1 CPU | 512Mi / 1Gi | N/A |

*Nota: Escalabilidade automática configurada baseada em utilização de CPU >70% e RPS >100*

### 2.5 Networking e Conectividade

- **Ingress**: `observability.innovabiz.cloud` (externo), `portal.observability.svc.cluster.local` (interno)
- **Portas**:
  - Frontend: 80/443 (HTTP/HTTPS)
  - Backend API: 3000 (interno)
  - Federation Service: 4000 (interno)
  - Redis: 6379 (interno)
- **Network Policies**: Acesso restrito de acordo com política zero-trust
- **Comunicação com Serviços**: TLS mútuo e tokens JWT para autenticação

## 3. Configuração Base

### 3.1 ConfigMap Principal

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: observability-portal-config
  namespace: innovabiz-observability
  labels:
    app: observability-portal
    component: configuration
data:
  config.json: |
    {
      "general": {
        "applicationName": "INNOVABIZ Observability Portal",
        "defaultLanguage": "pt-BR",
        "supportedLanguages": ["pt-BR", "en-US", "pt-AO", "es-ES"],
        "defaultTimezone": "America/Sao_Paulo",
        "refreshInterval": 60000,
        "maxCacheAge": 300000,
        "telemetry": {
          "enabled": true,
          "anonymousData": true
        }
      },
      "authentication": {
        "providers": [
          {
            "type": "oidc",
            "name": "INNOVABIZ SSO",
            "enabled": true,
            "primary": true,
            "config": {
              "clientId": "observability-portal",
              "authority": "https://auth.innovabiz.cloud/",
              "redirectUri": "https://observability.innovabiz.cloud/auth/callback",
              "postLogoutRedirectUri": "https://observability.innovabiz.cloud/",
              "scope": "openid profile email roles"
            }
          }
        ],
        "session": {
          "idleTimeout": 1800000,
          "absoluteTimeout": 28800000
        }
      },
      "datasources": {
        "prometheus": {
          "url": "https://prometheus.innovabiz-observability.svc.cluster.local:9090",
          "auth": {
            "type": "bearer"
          }
        },
        "elasticsearch": {
          "url": "https://elasticsearch-master.innovabiz-observability.svc.cluster.local:9200",
          "auth": {
            "type": "bearer"
          }
        },
        "loki": {
          "url": "https://loki.innovabiz-observability.svc.cluster.local:3100",
          "auth": {
            "type": "bearer"
          }
        },
        "jaeger": {
          "url": "https://jaeger-query.innovabiz-observability.svc.cluster.local:16686",
          "auth": {
            "type": "bearer"
          }
        },
        "grafana": {
          "url": "https://grafana.innovabiz-observability.svc.cluster.local:3000",
          "auth": {
            "type": "bearer"
          },
          "embedding": {
            "enabled": true,
            "allowedDashboards": ["*"]
          }
        },
        "kibana": {
          "url": "https://kibana.innovabiz-observability.svc.cluster.local:5601",
          "auth": {
            "type": "bearer"
          },
          "embedding": {
            "enabled": true,
            "allowedDashboards": ["*"]
          }
        }
      },
      "federation": {
        "enabled": true,
        "contextPropagation": true,
        "correlationEnabled": true,
        "correlationStrategies": ["time-based", "id-based", "semantic"],
        "maxResults": 1000,
        "timeout": 30000
      },
      "alerting": {
        "enabled": true,
        "providers": ["prometheus", "elasticsearch", "loki", "custom"],
        "aggregation": true,
        "deduplication": true,
        "silenceEnabled": true,
        "notificationChannels": ["email", "slack", "webhook", "pagerduty"]
      },
      "ui": {
        "theme": "innovabiz-dark",
        "alternativeThemes": ["innovabiz-light"],
        "defaultDashboard": "overview",
        "autoRefresh": true,
        "timeRangeOptions": ["5m", "15m", "30m", "1h", "3h", "6h", "12h", "24h", "7d", "30d", "custom"],
        "defaultTimeRange": "3h"
      },
      "features": {
        "logs": {
          "enabled": true,
          "defaultProvider": "elasticsearch",
          "alternatives": ["loki"]
        },
        "metrics": {
          "enabled": true,
          "defaultProvider": "prometheus"
        },
        "traces": {
          "enabled": true,
          "defaultProvider": "jaeger"
        },
        "dashboards": {
          "enabled": true
        },
        "serviceMaps": {
          "enabled": true,
          "autoRefresh": true,
          "refreshInterval": 60000
        },
        "alerts": {
          "enabled": true,
          "showOnDashboard": true
        },
        "reports": {
          "enabled": true,
          "exportFormats": ["pdf", "csv", "json"]
        },
        "healthChecks": {
          "enabled": true,
          "refreshInterval": 60000
        }
      }
    }
```

### 3.2 Secrets

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: observability-portal-secrets
  namespace: innovabiz-observability
type: Opaque
data:
  OIDC_CLIENT_SECRET: ${BASE64_CLIENT_SECRET}
  SESSION_SECRET: ${BASE64_SESSION_SECRET}
  ENCRYPTION_KEY: ${BASE64_ENCRYPTION_KEY}
  PROMETHEUS_TOKEN: ${BASE64_PROMETHEUS_TOKEN}
  ELASTICSEARCH_TOKEN: ${BASE64_ELASTICSEARCH_TOKEN}
  LOKI_TOKEN: ${BASE64_LOKI_TOKEN}
  GRAFANA_TOKEN: ${BASE64_GRAFANA_TOKEN}
  KIBANA_TOKEN: ${BASE64_KIBANA_TOKEN}
  JAEGER_TOKEN: ${BASE64_JAEGER_TOKEN}
  REDIS_PASSWORD: ${BASE64_REDIS_PASSWORD}
  SLACK_WEBHOOK_URL: ${BASE64_SLACK_URL}
  PAGERDUTY_SERVICE_KEY: ${BASE64_PAGERDUTY_KEY}
```# OBSERVABILITY PORTAL - Interface Unificada de Observabilidade

## 1. Visão Geral

O Observability Portal é a interface centralizada de observabilidade da plataforma INNOVABIZ, agregando dados de logs, métricas e traces de todos os componentes da infraestrutura em uma única visualização coerente. Este documento detalha a implementação, configuração e utilização do Portal como ponto central para monitoramento, troubleshooting e análise do IAM Audit Service.

### 1.1 Função na Arquitetura de Observabilidade

O Observability Portal atua como:

- **Interface Unificada**: Single pane of glass para toda a stack de observabilidade
- **Agregador Multi-Fonte**: Integração com Prometheus, Elasticsearch, Loki, Grafana e Kibana
- **Gateway de Observabilidade**: Ponto de entrada único para todas as ferramentas
- **Federação de Contexto**: Propagação de contexto entre ferramentas integradas
- **Plataforma de Alerta Centralizada**: Unificação de alertas de múltiplas fontes
- **Centro de Operações**: Interface para diagnóstico e resolução de incidentes

### 1.2 Recursos e Capacidades

- **Dashboard Unificado**: Visão consolidada de saúde, performance e status
- **Navegação Contextual**: Drill-down entre logs, métricas e traces preservando contexto
- **Correlação Automática**: Relacionamento entre eventos de diferentes fontes
- **Alertas Centralizados**: Consolidação e gestão de alertas de múltiplos sistemas
- **Visualização Multi-Dimensional**: Adaptação automática ao contexto (tenant, região, ambiente)
- **Service Map**: Visualização de topologia e dependências entre serviços
- **Health Status**: Status de saúde em tempo real de todos os componentes
- **Analytics**: Análise de tendências e anomalias

## 2. Arquitetura de Implementação

### 2.1 Diagrama de Arquitetura

```
                          ┌────────────────────────────────┐
                          │        Load Balancer           │
                          │  (portal.innovabiz.cloud)      │
                          └──────────────┬─────────────────┘
                                         │
                                         ▼
                          ┌────────────────────────────────┐
                          │          OIDC Proxy            │
                          │    (Auth + Context Injection)   │
                          └──────────────┬─────────────────┘
                                         │
                                         ▼
           ┌──────────────────────────────────────────────────────────────┐
           │                  Observability Portal                         │
           │                                                              │
           │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐           │
           │  │ Portal UI   │  │ Backend API │  │ Federation  │           │
           │  │ (React/     │  │ (Node.js)   │  │ Service     │           │
           │  │  TypeScript)│  │             │  │             │           │
           │  └─────────────┘  └─────────────┘  └─────────────┘           │
           │         │               │                │                   │
           └─────────┼───────────────┼────────────────┼───────────────────┘
                     │               │                │
        ┌────────────┼───────────────┼────────────────┼───────────────────┐
        │            │               │                │                   │
        │   ┌────────▼─────┐  ┌──────▼───────┐  ┌─────▼────────┐  ┌───────▼───────┐
        │   │              │  │              │  │              │  │               │
        │   │  Grafana     │  │  Kibana      │  │ Prometheus   │  │ Jaeger/Tempo  │
        │   │  (Métricas/  │  │  (Logs/      │  │ (Métricas)   │  │ (Traces)      │
        │   │   Loki)      │  │   Elastic)   │  │              │  │               │
        │   │              │  │              │  │              │  │               │
        │   └──────────────┘  └──────────────┘  └──────────────┘  └───────────────┘
        │            │               │                │                   │         
        └────────────┼───────────────┼────────────────┼───────────────────┘         
                     │               │                │                            
        ┌────────────┼───────────────┼────────────────┼───────────────────┐         
        │   ┌────────▼─────┐  ┌──────▼───────┐  ┌─────▼────────┐  ┌───────▼───────┐
        │   │              │  │              │  │              │  │               │
        │   │  Loki        │  │ Elasticsearch │  │ Prometheus   │  │ Tempo/Jaeger  │
        │   │  (Logs)      │  │  (Logs)       │  │ (Storage)    │  │ (Storage)     │
        │   │              │  │              │  │              │  │               │
        │   └──────────────┘  └──────────────┘  └──────────────┘  └───────────────┘
        │                                                                         │
        └─────────────────────────────────────────────────────────────────────────┘
```

### 2.2 Componentes Principais

#### 2.2.1 Portal UI (Frontend)

- **Framework**: React 18 com TypeScript
- **Design System**: Material UI com tema customizado INNOVABIZ
- **Estado**: Redux Toolkit + RTK Query
- **Visualizações**: D3.js, ECharts, React Flow
- **Internacionalização**: i18n com suporte a pt-BR, en-US, pt-AO, es-ES
- **Comunicação**: REST API + GraphQL + WebSockets para atualizações em tempo real

#### 2.2.2 Backend API

- **Runtime**: Node.js 18+ com Express
- **API**: REST + GraphQL (Apollo Server)
- **Cache**: Redis para cache distribuído e sessão
- **Comunicação**: Axios para integração com outros sistemas
- **Autenticação**: Passport.js com estratégias OIDC e JWT
- **Logging**: Winston com formato estruturado (JSON)

#### 2.2.3 Federation Service

- **Agregação**: Agregador de dados de múltiplas fontes
- **Tradução**: Normalização de formatos entre sistemas
- **Contexto**: Propagação de contexto multi-dimensional
- **Correlação**: Identificação de relações entre eventos
- **Proxy Reverso**: Reescrita e encaminhamento de requisições

### 2.3 Implantação Kubernetes

O Observability Portal é implantado como um conjunto de Deployments Kubernetes:

- **Namespace**: `innovabiz-observability`
- **Deployments**:
  - `portal-frontend`: UI React (3 réplicas)
  - `portal-backend`: API Node.js (3 réplicas)
  - `portal-federation`: Serviço de Federação (2 réplicas)
- **Services**: ClusterIP internos + Ingress para acesso externo
- **Pods**: Configuração anti-affinity para alta disponibilidade
- **TLS**: Certificados gerenciados via cert-manager
- **ConfigMaps**: Configuração externalizada e versionada no Git

### 2.4 Especificação de Recursos

| Componente | Réplicas | CPU (Request/Limit) | Memória (Request/Limit) | Armazenamento |
|------------|----------|---------------------|-------------------------|--------------|
| Portal Frontend | 3 | 0.5 CPU / 1 CPU | 512Mi / 1Gi | N/A |
| Portal Backend | 3 | 1 CPU / 2 CPU | 1Gi / 2Gi | N/A |
| Federation Service | 2 | 1 CPU / 2 CPU | 1Gi / 2Gi | N/A |
| Redis Cache | 3 | 0.5 CPU / 1 CPU | 1Gi / 2Gi | 10Gi (PVC) |
| Ingress | 2 | 0.5 CPU / 1 CPU | 512Mi / 1Gi | N/A |

*Nota: Escalabilidade automática configurada baseada em utilização de CPU >70% e RPS >100*

### 2.5 Networking e Conectividade

- **Ingress**: `observability.innovabiz.cloud` (externo), `portal.observability.svc.cluster.local` (interno)
- **Portas**:
  - Frontend: 80/443 (HTTP/HTTPS)
  - Backend API: 3000 (interno)
  - Federation Service: 4000 (interno)
  - Redis: 6379 (interno)
- **Network Policies**: Acesso restrito de acordo com política zero-trust
- **Comunicação com Serviços**: TLS mútuo e tokens JWT para autenticação

## 3. Configuração Base

### 3.1 ConfigMap Principal

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: observability-portal-config
  namespace: innovabiz-observability
  labels:
    app: observability-portal
    component: configuration
data:
  config.json: |
    {
      "general": {
        "applicationName": "INNOVABIZ Observability Portal",
        "defaultLanguage": "pt-BR",
        "supportedLanguages": ["pt-BR", "en-US", "pt-AO", "es-ES"],
        "defaultTimezone": "America/Sao_Paulo",
        "refreshInterval": 60000,
        "maxCacheAge": 300000,
        "telemetry": {
          "enabled": true,
          "anonymousData": true
        }
      },
      "authentication": {
        "providers": [
          {
            "type": "oidc",
            "name": "INNOVABIZ SSO",
            "enabled": true,
            "primary": true,
            "config": {
              "clientId": "observability-portal",
              "authority": "https://auth.innovabiz.cloud/",
              "redirectUri": "https://observability.innovabiz.cloud/auth/callback",
              "postLogoutRedirectUri": "https://observability.innovabiz.cloud/",
              "scope": "openid profile email roles"
            }
          }
        ],
        "session": {
          "idleTimeout": 1800000,
          "absoluteTimeout": 28800000
        }
      },
      "datasources": {
        "prometheus": {
          "url": "https://prometheus.innovabiz-observability.svc.cluster.local:9090",
          "auth": {
            "type": "bearer"
          }
        },
        "elasticsearch": {
          "url": "https://elasticsearch-master.innovabiz-observability.svc.cluster.local:9200",
          "auth": {
            "type": "bearer"
          }
        },
        "loki": {
          "url": "https://loki.innovabiz-observability.svc.cluster.local:3100",
          "auth": {
            "type": "bearer"
          }
        },
        "jaeger": {
          "url": "https://jaeger-query.innovabiz-observability.svc.cluster.local:16686",
          "auth": {
            "type": "bearer"
          }
        },
        "grafana": {
          "url": "https://grafana.innovabiz-observability.svc.cluster.local:3000",
          "auth": {
            "type": "bearer"
          },
          "embedding": {
            "enabled": true,
            "allowedDashboards": ["*"]
          }
        },
        "kibana": {
          "url": "https://kibana.innovabiz-observability.svc.cluster.local:5601",
          "auth": {
            "type": "bearer"
          },
          "embedding": {
            "enabled": true,
            "allowedDashboards": ["*"]
          }
        }
      },
      "federation": {
        "enabled": true,
        "contextPropagation": true,
        "correlationEnabled": true,
        "correlationStrategies": ["time-based", "id-based", "semantic"],
        "maxResults": 1000,
        "timeout": 30000
      },
      "alerting": {
        "enabled": true,
        "providers": ["prometheus", "elasticsearch", "loki", "custom"],
        "aggregation": true,
        "deduplication": true,
        "silenceEnabled": true,
        "notificationChannels": ["email", "slack", "webhook", "pagerduty"]
      },
      "ui": {
        "theme": "innovabiz-dark",
        "alternativeThemes": ["innovabiz-light"],
        "defaultDashboard": "overview",
        "autoRefresh": true,
        "timeRangeOptions": ["5m", "15m", "30m", "1h", "3h", "6h", "12h", "24h", "7d", "30d", "custom"],
        "defaultTimeRange": "3h"
      },
      "features": {
        "logs": {
          "enabled": true,
          "defaultProvider": "elasticsearch",
          "alternatives": ["loki"]
        },
        "metrics": {
          "enabled": true,
          "defaultProvider": "prometheus"
        },
        "traces": {
          "enabled": true,
          "defaultProvider": "jaeger"
        },
        "dashboards": {
          "enabled": true
        },
        "serviceMaps": {
          "enabled": true,
          "autoRefresh": true,
          "refreshInterval": 60000
        },
        "alerts": {
          "enabled": true,
          "showOnDashboard": true
        },
        "reports": {
          "enabled": true,
          "exportFormats": ["pdf", "csv", "json"]
        },
        "healthChecks": {
          "enabled": true,
          "refreshInterval": 60000
        }
      }
    }
```

### 3.2 Secrets

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: observability-portal-secrets
  namespace: innovabiz-observability
type: Opaque
data:
  OIDC_CLIENT_SECRET: ${BASE64_CLIENT_SECRET}
  SESSION_SECRET: ${BASE64_SESSION_SECRET}
  ENCRYPTION_KEY: ${BASE64_ENCRYPTION_KEY}
  PROMETHEUS_TOKEN: ${BASE64_PROMETHEUS_TOKEN}
  ELASTICSEARCH_TOKEN: ${BASE64_ELASTICSEARCH_TOKEN}
  LOKI_TOKEN: ${BASE64_LOKI_TOKEN}
  GRAFANA_TOKEN: ${BASE64_GRAFANA_TOKEN}
  KIBANA_TOKEN: ${BASE64_KIBANA_TOKEN}
  JAEGER_TOKEN: ${BASE64_JAEGER_TOKEN}
  REDIS_PASSWORD: ${BASE64_REDIS_PASSWORD}
  SLACK_WEBHOOK_URL: ${BASE64_SLACK_URL}
  PAGERDUTY_SERVICE_KEY: ${BASE64_PAGERDUTY_KEY}
```

### 3.3 Deployment Federation Service

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: observability-portal-federation
  namespace: innovabiz-observability
  labels:
    app: observability-portal
    component: federation
spec:
  replicas: 2
  selector:
    matchLabels:
      app: observability-portal
      component: federation
  template:
    metadata:
      labels:
        app: observability-portal
        component: federation
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
                  - observability-portal
              topologyKey: kubernetes.io/hostname
      containers:
      - name: federation
        image: innovabiz/observability-portal-federation:${VERSION}
        imagePullPolicy: Always
        ports:
        - containerPort: 4000
        resources:
          requests:
            cpu: 1
            memory: 1Gi
          limits:
            cpu: 2
            memory: 2Gi
        env:
        - name: NODE_ENV
          value: "production"
        - name: PORT
          value: "4000"
        - name: PROMETHEUS_URL
          value: "https://prometheus.innovabiz-observability.svc.cluster.local:9090"
        - name: ELASTICSEARCH_URL
          value: "https://elasticsearch-master.innovabiz-observability.svc.cluster.local:9200"
        - name: LOKI_URL
          value: "https://loki.innovabiz-observability.svc.cluster.local:3100"
        - name: JAEGER_URL
          value: "https://jaeger-query.innovabiz-observability.svc.cluster.local:16686"
        - name: PROMETHEUS_TOKEN
          valueFrom:
            secretKeyRef:
              name: observability-portal-secrets
              key: PROMETHEUS_TOKEN
        - name: ELASTICSEARCH_TOKEN
          valueFrom:
            secretKeyRef:
              name: observability-portal-secrets
              key: ELASTICSEARCH_TOKEN
        - name: LOKI_TOKEN
          valueFrom:
            secretKeyRef:
              name: observability-portal-secrets
              key: LOKI_TOKEN
        - name: JAEGER_TOKEN
          valueFrom:
            secretKeyRef:
              name: observability-portal-secrets
              key: JAEGER_TOKEN
        volumeMounts:
        - name: config-volume
          mountPath: /app/config
        - name: certs-volume
          mountPath: /app/certs
        livenessProbe:
          httpGet:
            path: /health
            port: 4000
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 4000
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config-volume
        configMap:
          name: observability-portal-config
      - name: certs-volume
        secret:
          secretName: observability-portal-certs
```

## 4. Multi-Dimensionalidade

O Observability Portal implementa suporte completo a múltiplas dimensões de contexto, permitindo visualização e filtragem dos dados de observabilidade de acordo com os diferentes contextos operacionais da plataforma INNOVABIZ.

### 4.1 Dimensões Suportadas

- **Tenant (Multi-Tenant)**: Isolamento e visualização por inquilino
- **Região (Multi-Região)**: Filtragem por regiões geográficas (BR, US, EU, AO)
- **Ambiente (Multi-Ambiente)**: Segregação por ambiente (DEV, QA, STG, PROD)
- **Módulo (Multi-Módulo)**: Navegação por módulo funcional da plataforma
- **Componente (Multi-Componente)**: Detalhamento por componente técnico

### 4.2 Implementação de Contexto

#### 4.2.1 Propagação de Contexto

O Portal implementa propagação de contexto consistente entre todas as ferramentas integradas:

```javascript
// Exemplo de objeto de contexto propagado
const context = {
  tenant: {
    id: "tenant-123",
    name: "Organização ABC"
  },
  region: "BR",
  environment: "PROD",
  module: "IAM",
  component: "AuthService",
  timeRange: {
    from: "2025-07-30T10:00:00Z",
    to: "2025-07-30T16:00:00Z"
  },
  filters: [
    { key: "level", value: "error", operator: "=" },
    { key: "user", value: "admin", operator: "!=" }
  ]
};
```

#### 4.2.2 Seleção de Contexto

A interface oferece um seletor de contexto persistente que permite:

- Visualização e edição do contexto atual
- Salvamento de contextos favoritos
- Restauração de contextos anteriores
- Propagação automática para todas as ferramentas

#### 4.2.3 Labels Multi-Dimensionais

Todas as métricas, logs e traces são automaticamente filtrados pelos seletores de contexto usando labels padronizados:

- `tenant`: ID do tenant (ex: `tenant-123`)
- `region`: Código da região (ex: `BR`, `US`, `EU`, `AO`)
- `environment`: Ambiente (ex: `DEV`, `QA`, `STG`, `PROD`)
- `module`: Módulo funcional (ex: `IAM`, `Payment`, `Marketplace`)
- `component`: Componente técnico (ex: `AuthService`, `AuditService`, `TokenService`)

### 4.3 Tradução de Contexto

O Federation Service implementa tradução de contexto entre as diferentes ferramentas:

| Plataforma | Tenant | Região | Ambiente | Módulo | Componente |
|------------|--------|--------|----------|--------|------------|
| Prometheus | tenant_id | region | env | module | component |
| Elasticsearch | tenant.id | region.keyword | environment.keyword | module.keyword | component.keyword |
| Loki | tenant | region | environment | module | component |
| Grafana | tenant | region | environment | module | component |
| Jaeger | tenant | region | environment | service.module | service.name |

## 5. Dashboards e Visualizações

### 5.1 Dashboard Principal

O dashboard principal oferece uma visão consolidada de todo o ambiente IAM, destacando:

- **Status de Saúde**: Estado atual de todos os componentes do IAM
- **SLIs/SLOs**: Indicadores de nível de serviço com status atual
- **Alertas Ativos**: Alertas não resolvidos por severidade
- **Métricas-Chave**: Taxa de autenticação, autorização, latência de resposta
- **Top Erros**: Erros mais frequentes nas últimas 24 horas
- **Atividade de Auditoria**: Volume de eventos de auditoria por tipo

### 5.2 Dashboards Específicos do IAM

#### 5.2.1 Autenticação

Dashboard focado em operações de autenticação:

- Taxa de sucesso/falha por método de autenticação
- Latência média de autenticação
- Volume de autenticações por protocolo (OIDC, SAML, etc.)
- Distribuição geográfica de tentativas de login
- Top 10 usuários por volume de autenticação
- Tentativas de login inválidas por origem

#### 5.2.2 Autorização

Dashboard para monitoramento de autorização:

- Volume de decisões de autorização (permitidas/negadas)
- Latência do serviço de autorização
- Recursos mais acessados
- Top 10 políticas acionadas
- Distribuição de decisões por roles
- Cache hit/miss ratio

#### 5.2.3 Auditoria

Dashboard específico para auditoria:

- Volume de eventos por categoria
- Eventos de alta severidade
- Distribuição de eventos por usuário/serviço
- Taxa de ingestão de logs
- Volume de dados por tenant
- Latência de indexação

#### 5.2.4 Segurança

Dashboard de monitoramento de segurança:

- Tentativas de acesso suspeitas
- Alterações em permissões críticas
- Atividade administrativa
- Tentativas de escalonamento de privilégios
- Mudanças em políticas de segurança
- Alertas de vulnerabilidades

### 5.3 Service Maps

O Portal inclui visualização de topologia dos serviços do IAM:

- Mapeamento de dependências entre serviços
- Estado atual de cada serviço
- Volume de tráfego entre componentes
- Latência entre serviços
- Taxas de erro nas comunicações
- Detecção automática de bottlenecks

### 5.4 Exemplo de Visualização: Análise de Falhas de Autenticação

```json
{
  "title": "Análise de Falhas de Autenticação",
  "type": "visualization",
  "visualizationType": "composite",
  "components": [
    {
      "title": "Taxa de Falha de Autenticação (últimas 24h)",
      "type": "timeseries",
      "data": {
        "metric": "iam_authentication_failures_total",
        "aggregation": "rate",
        "period": "5m",
        "filters": [
          {"tenant": "${tenant}"},
          {"region": "${region}"},
          {"environment": "${environment}"}
        ]
      },
      "thresholds": [
        {"value": 0.01, "color": "yellow", "label": "Warning"},
        {"value": 0.05, "color": "red", "label": "Critical"}
      ]
    },
    {
      "title": "Falhas por Método de Autenticação",
      "type": "piechart",
      "data": {
        "metric": "iam_authentication_failures_total",
        "dimension": "auth_method",
        "filters": [
          {"tenant": "${tenant}"},
          {"region": "${region}"},
          {"environment": "${environment}"}
        ]
      }
    },
    {
      "title": "Top 5 Razões de Falha",
      "type": "barchart",
      "data": {
        "metric": "iam_authentication_failures_total",
        "dimension": "reason",
        "limit": 5,
        "filters": [
          {"tenant": "${tenant}"},
          {"region": "${region}"},
          {"environment": "${environment}"}
        ]
      }
    },
    {
      "title": "Logs de Falha Recentes",
      "type": "logs",
      "data": {
        "query": "level=error AND module=IAM AND component=AuthService",
        "source": "elasticsearch",
        "limit": 10,
        "filters": [
          {"tenant": "${tenant}"},
          {"region": "${region}"},
          {"environment": "${environment}"}
        ]
      }
    }
  ]
}
```## 6. Segurança e Compliance

### 6.1 Modelo de Segurança

O Observability Portal implementa um modelo de segurança robusto em múltiplas camadas:

1. **Autenticação**: OIDC integrado com INNOVABIZ SSO
2. **Autorização**: RBAC com permissões granulares
3. **Comunicação**: TLS mútuo para todas as comunicações
4. **Tokens**: JWT com assinatura RS256
5. **Auditoria**: Logging completo de todas as ações
6. **Secrets**: Gerenciados via Kubernetes Secrets
7. **Network**: Zero-trust network policies

### 6.2 RBAC - Controle de Acesso Baseado em Papéis

O portal implementa um modelo RBAC multinível que controla o acesso de acordo com:

#### 6.2.1 Papéis Padrão

| Papel | Descrição | Permissões |
|-------|-----------|------------|
| `observability-viewer` | Acesso somente leitura | Visualizar dashboards e alertas |
| `observability-editor` | Capacidade de edição | Visualizar + editar dashboards e visualizações |
| `observability-admin` | Administrador do portal | Acesso completo incluindo configuração |
| `observability-auditor` | Auditor do sistema | Visualizar logs, eventos de auditoria e relatórios |
| `observability-operator` | Operador de alertas | Gerenciar alertas e notificações |

#### 6.2.2 Matriz de Permissões

| Recurso | Viewer | Editor | Operator | Auditor | Admin |
|---------|--------|--------|----------|---------|-------|
| Visualizar dashboards | ✅ | ✅ | ✅ | ✅ | ✅ |
| Editar dashboards | ❌ | ✅ | ❌ | ❌ | ✅ |
| Visualizar alertas | ✅ | ✅ | ✅ | ✅ | ✅ |
| Gerenciar alertas | ❌ | ❌ | ✅ | ❌ | ✅ |
| Silenciar alertas | ❌ | ❌ | ✅ | ❌ | ✅ |
| Acessar configuração | ❌ | ❌ | ❌ | ❌ | ✅ |
| Visualizar logs | ✅ | ✅ | ✅ | ✅ | ✅ |
| Acessar dados sensíveis | ❌ | ❌ | ❌ | ✅ | ✅ |
| Exportar relatórios | ✅ | ✅ | ✅ | ✅ | ✅ |
| Gerenciar usuários | ❌ | ❌ | ❌ | ❌ | ✅ |

#### 6.2.3 Permissões Multi-Tenant

Todas as permissões são aplicadas com escopo de tenant, permitindo configurações diferentes por tenant:

```yaml
# Exemplo de configuração RBAC multi-tenant
tenants:
  - id: "tenant-123"
    name: "Organização ABC"
    roles:
      - name: "observability-viewer"
        subjects: ["user:analyst@abc.com", "group:developers"]
      - name: "observability-admin"
        subjects: ["user:admin@abc.com", "group:platform-team"]
  
  - id: "tenant-456"
    name: "Empresa XYZ"
    roles:
      - name: "observability-viewer"
        subjects: ["user:support@xyz.com", "group:it-team"]
      - name: "observability-editor"
        subjects: ["user:devops@xyz.com"]
```

### 6.3 Auditoria e Compliance

O Portal mantém registros de auditoria completos para todas as ações:

- Acessos ao sistema
- Visualização de dashboards
- Exportação de dados
- Configurações alteradas
- Alertas silenciados
- Consultas executadas

Estes logs são armazenados no Elasticsearch com retenção configurada de acordo com requisitos de compliance:

- Registros sensíveis: 12 meses
- Registros de autenticação: 6 meses
- Registros de visualização: 3 meses

### 6.4 Gestão de Vulnerabilidades

O Portal segue práticas rigorosas de segurança:

- Scan de dependências automatizado (npm audit, OWASP Dependency Check)
- Análise estática de código (ESLint security, SonarQube)
- Testes de penetração periódicos
- Container scanning (Trivy)
- Atualização regular de dependências
- Security headers configurados (CSP, HSTS, etc.)

## 7. Monitoramento e Alertas

### 7.1 Monitoramento do Portal

O Observability Portal inclui monitoramento próprio com métricas expostas via endpoint `/metrics` em formato Prometheus:

#### 7.1.1 Métricas Principais

| Métrica | Tipo | Descrição |
|---------|------|-----------|
| `portal_http_requests_total` | Counter | Total de requisições HTTP |
| `portal_http_request_duration_seconds` | Histogram | Latência das requisições HTTP |
| `portal_query_duration_seconds` | Histogram | Tempo de execução de queries |
| `portal_federation_requests_total` | Counter | Requisições ao serviço de federação |
| `portal_federation_errors_total` | Counter | Erros no serviço de federação |
| `portal_authentication_failures_total` | Counter | Falhas de autenticação |
| `portal_cache_hits_total` | Counter | Cache hits |
| `portal_cache_misses_total` | Counter | Cache misses |
| `portal_active_users` | Gauge | Usuários ativos simultaneamente |
| `portal_datasource_availability` | Gauge | Disponibilidade das fontes de dados (0-1) |

#### 7.1.2 Dashboard de Auto-Monitoramento

O Portal inclui um dashboard específico para monitorar sua própria saúde e performance:

- Saúde de componentes (Frontend, Backend, Federation)
- Latência de requisições
- Taxa de erros
- Uso de recursos (CPU, memória)
- Disponibilidade de fontes de dados
- Cache hit ratio
- Usuários ativos

### 7.2 Alertas Pré-Configurados

O Portal vem com alertas pré-configurados para detectar problemas comuns:

#### 7.2.1 Alertas de Disponibilidade

```yaml
- name: ObservabilityPortalComponentDown
  expr: up{job="observability-portal"} == 0
  for: 2m
  labels:
    severity: critical
    component: observability-portal
  annotations:
    summary: "Componente do Observability Portal indisponível"
    description: "O componente {{ $labels.instance }} está indisponível há mais de 2 minutos."

- name: ObservabilityDataSourceUnavailable
  expr: portal_datasource_availability{} == 0
  for: 5m
  labels:
    severity: critical
    component: observability-portal
  annotations:
    summary: "Fonte de dados indisponível"
    description: "A fonte de dados {{ $labels.datasource }} está indisponível há mais de 5 minutos."
```

#### 7.2.2 Alertas de Performance

```yaml
- name: ObservabilityPortalHighLatency
  expr: histogram_quantile(0.95, sum(rate(portal_http_request_duration_seconds_bucket{job="observability-portal"}[5m])) by (le)) > 2
  for: 10m
  labels:
    severity: warning
    component: observability-portal
  annotations:
    summary: "Alta latência no Observability Portal"
    description: "95% das requisições estão levando mais de 2 segundos para serem processadas."

- name: ObservabilityPortalHighErrorRate
  expr: sum(rate(portal_http_requests_total{job="observability-portal", status_code=~"5.."}[5m])) / sum(rate(portal_http_requests_total{job="observability-portal"}[5m])) > 0.05
  for: 5m
  labels:
    severity: warning
    component: observability-portal
  annotations:
    summary: "Alta taxa de erros no Observability Portal"
    description: "Mais de 5% das requisições estão resultando em erros 5xx."
```

#### 7.2.3 Alertas de Recursos

```yaml
- name: ObservabilityPortalHighMemoryUsage
  expr: container_memory_usage_bytes{container_name=~"observability-portal-.*"} / container_spec_memory_limit_bytes{container_name=~"observability-portal-.*"} > 0.85
  for: 15m
  labels:
    severity: warning
    component: observability-portal
  annotations:
    summary: "Alto uso de memória no Observability Portal"
    description: "O container {{ $labels.container_name }} está utilizando mais de 85% da memória disponível."

- name: ObservabilityPortalHighCPUUsage
  expr: sum(rate(container_cpu_usage_seconds_total{container_name=~"observability-portal-.*"}[5m])) by (container_name) / sum(container_spec_cpu_quota{container_name=~"observability-portal-.*"}) by (container_name) > 0.85
  for: 15m
  labels:
    severity: warning
    component: observability-portal
  annotations:
    summary: "Alto uso de CPU no Observability Portal"
    description: "O container {{ $labels.container_name }} está utilizando mais de 85% da CPU disponível."
```

## 8. Procedimentos Operacionais

### 8.1 Instalação e Atualização

#### 8.1.1 Pré-requisitos

- Kubernetes 1.23+
- Helm 3.8+
- Cert-manager instalado no cluster
- OIDC Provider configurado
- Redis disponível para cache
- Ferramentas de observabilidade já instaladas (Prometheus, Elasticsearch, etc.)

#### 8.1.2 Procedimento de Instalação

```bash
# 1. Adicione o repositório Helm do INNOVABIZ
helm repo add innovabiz https://charts.innovabiz.cloud/
helm repo update

# 2. Crie o namespace
kubectl create namespace innovabiz-observability

# 3. Configure os valores para o chart
cat > values.yaml << EOF
global:
  environment: production
  region: BR

image:
  repository: innovabiz/observability-portal
  tag: 1.5.0

replicas:
  frontend: 3
  backend: 3
  federation: 2

ingress:
  host: observability.innovabiz.cloud
  tls:
    enabled: true
    certManager: true
    issuer: letsencrypt-prod

oidc:
  enabled: true
  provider: innovabiz-sso
  clientId: observability-portal
  # clientSecret deve ser fornecido via --set-file

datasources:
  prometheus:
    url: https://prometheus.innovabiz-observability.svc.cluster.local:9090
  elasticsearch:
    url: https://elasticsearch-master.innovabiz-observability.svc.cluster.local:9200
  loki:
    url: https://loki.innovabiz-observability.svc.cluster.local:3100
  jaeger:
    url: https://jaeger-query.innovabiz-observability.svc.cluster.local:16686
  grafana:
    url: https://grafana.innovabiz-observability.svc.cluster.local:3000
  kibana:
    url: https://kibana.innovabiz-observability.svc.cluster.local:5601

redis:
  enabled: true
  architecture: replication
  auth:
    enabled: true
  persistence:
    enabled: true
    size: 10Gi
EOF

# 4. Instale o chart com os valores customizados
helm install observability-portal innovabiz/observability-portal \
  --namespace innovabiz-observability \
  --values values.yaml \
  --set-file oidc.clientSecret=./client_secret.txt

# 5. Verifique a instalação
kubectl get pods -n innovabiz-observability
```

#### 8.1.3 Procedimento de Atualização

```bash
# 1. Atualize o repositório
helm repo update

# 2. Atualize o chart
helm upgrade observability-portal innovabiz/observability-portal \
  --namespace innovabiz-observability \
  --values values.yaml \
  --set-file oidc.clientSecret=./client_secret.txt

# 3. Verifique o status do upgrade
kubectl get pods -n innovabiz-observability
```

### 8.2 Backup e Recuperação

#### 8.2.1 Componentes a Serem Backups

- **ConfigMaps**: Configurações do portal
- **Secrets**: Credenciais e tokens
- **Redis PVC**: Dados de cache e sessão
- **Dashboards customizados**: Exportação via API

#### 8.2.2 Procedimento de Backup

```bash
# 1. Backup de ConfigMaps e Secrets
kubectl get configmap -n innovabiz-observability -o yaml > observability-configmaps.yaml
kubectl get secret -n innovabiz-observability -o yaml > observability-secrets.yaml

# 2. Backup do Redis (utilizando Velero)
velero backup create observability-redis-backup \
  --include-namespaces innovabiz-observability \
  --selector "app=redis"

# 3. Backup de dashboards customizados
curl -X GET https://observability.innovabiz.cloud/api/v1/dashboards \
  -H "Authorization: Bearer $TOKEN" \
  -o dashboards-backup.json
```

#### 8.2.3 Procedimento de Recuperação

```bash
# 1. Restaure ConfigMaps e Secrets
kubectl apply -f observability-configmaps.yaml
kubectl apply -f observability-secrets.yaml

# 2. Recupere o Redis (utilizando Velero)
velero restore create --from-backup observability-redis-backup

# 3. Restaure dashboards customizados
curl -X POST https://observability.innovabiz.cloud/api/v1/dashboards/import \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d @dashboards-backup.json
```

### 8.3 Escalabilidade

O Portal foi projetado para escalar horizontalmente:

#### 8.3.1 Escalabilidade Horizontal

```bash
# Escalar frontend
kubectl scale deployment observability-portal-frontend \
  --replicas=5 -n innovabiz-observability

# Escalar backend
kubectl scale deployment observability-portal-backend \
  --replicas=5 -n innovabiz-observability

# Escalar federation service
kubectl scale deployment observability-portal-federation \
  --replicas=3 -n innovabiz-observability
```

#### 8.3.2 Autoscaling

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: observability-portal-frontend-hpa
  namespace: innovabiz-observability
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: observability-portal-frontend
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
    scaleUp:
      stabilizationWindowSeconds: 60
```

## 9. Troubleshooting

### 9.1 Problemas Comuns e Soluções

#### 9.1.1 Portal Indisponível

**Sintomas:**
- UI não carrega
- Erro HTTP 502/503
- Timeout ao acessar o portal

**Verificações:**
```bash
# Verifique se todos os pods estão rodando
kubectl get pods -n innovabiz-observability

# Verifique os logs do frontend
kubectl logs -l app=observability-portal,component=frontend -n innovabiz-observability

# Verifique os logs do backend
kubectl logs -l app=observability-portal,component=backend -n innovabiz-observability

# Verifique o ingress
kubectl get ingress -n innovabiz-observability
kubectl describe ingress observability-portal -n innovabiz-observability
```

**Soluções:**
- Reinicie os pods com problemas
- Verifique se o ingress está configurado corretamente
- Verifique se os certificados TLS estão válidos
- Confirme se o Redis está disponível

#### 9.1.2 Erros de Autenticação

**Sintomas:**
- Redirecionamentos infinitos
- Erro "Unauthorized"
- Falha ao fazer login

**Verificações:**
```bash
# Verifique os logs relacionados à autenticação
kubectl logs -l app=observability-portal,component=backend -n innovabiz-observability | grep -i auth

# Verifique o status do OIDC provider
curl -k https://auth.innovabiz.cloud/.well-known/openid-configuration

# Verifique os secrets
kubectl get secret observability-portal-secrets -n innovabiz-observability
```

**Soluções:**
- Verifique se o OIDC provider está configurado corretamente
- Confirme se o client secret está correto
- Verifique se os redirecionamentos estão autorizados no OIDC provider

#### 9.1.3 Fontes de Dados Indisponíveis

**Sintomas:**
- Dashboards não carregam dados
- Erro "Data source unavailable"
- Gráficos vazios

**Verificações:**
```bash
# Verifique o status das fontes de dados
kubectl exec -it $(kubectl get pods -l app=observability-portal,component=backend -n innovabiz-observability -o jsonpath='{.items[0].metadata.name}') -n innovabiz-observability -- curl -s localhost:3000/api/v1/datasources/health

# Verifique logs de conexão
kubectl logs -l app=observability-portal,component=federation -n innovabiz-observability | grep -i connection

# Teste conexão direta com a fonte
kubectl exec -it $(kubectl get pods -l app=observability-portal,component=backend -n innovabiz-observability -o jsonpath='{.items[0].metadata.name}') -n innovabiz-observability -- curl -k https://prometheus.innovabiz-observability.svc.cluster.local:9090/api/v1/status/config
```

**Soluções:**
- Verifique se as fontes de dados estão acessíveis
- Confirme se os tokens de autenticação estão corretos
- Verifique as network policies

### 9.2 Logs de Diagnóstico

O Portal implementa logging estruturado em formato JSON:

```json
{
  "timestamp": "2025-07-30T15:23:45.123Z",
  "level": "error",
  "component": "federation",
  "method": "GET",
  "path": "/api/v1/federation/metrics",
  "status": 500,
  "latency_ms": 1532,
  "error": "Timeout connecting to Prometheus",
  "trace_id": "abc123def456",
  "tenant": "tenant-123",
  "user": "admin@innovabiz.com"
}
```

**Níveis de logging configuráveis:**
- `debug`: Informações detalhadas para debugging
- `info`: Informações operacionais normais
- `warn`: Situações potencialmente problemáticas
- `error`: Erros que não interrompem o serviço
- `fatal`: Erros críticos que podem interromper o serviço

### 9.3 Verificações de Saúde

O Portal expõe endpoints de health check:

- `/health`: Status geral (200 OK se saudável)
- `/health/ready`: Readiness check (200 OK quando pronto para tráfego)
- `/health/live`: Liveness check (200 OK quando o serviço está respondendo)
- `/health/datasources`: Status de todas as fontes de dados
- `/health/components`: Status de todos os componentes internos

Exemplo de resposta do endpoint `/health/datasources`:

```json
{
  "status": "warning",
  "timestamp": "2025-07-30T15:25:30Z",
  "datasources": [
    {
      "name": "prometheus",
      "status": "healthy",
      "latency_ms": 45,
      "last_check": "2025-07-30T15:25:29Z"
    },
    {
      "name": "elasticsearch",
      "status": "healthy",
      "latency_ms": 120,
      "last_check": "2025-07-30T15:25:28Z"
    },
    {
      "name": "loki",
      "status": "degraded",
      "latency_ms": 1500,
      "last_check": "2025-07-30T15:25:27Z",
      "message": "High latency detected"
    },
    {
      "name": "jaeger",
      "status": "healthy",
      "latency_ms": 87,
      "last_check": "2025-07-30T15:25:26Z"
    }
  ]
}
```

## 10. Compliance e Certificações

### 10.1 Compliance Regulatória

O Observability Portal foi projetado para atender aos seguintes requisitos regulatórios:

| Regulação | Escopo | Requisitos Atendidos |
|-----------|--------|----------------------|
| LGPD | Brasil | Logs de auditoria, controle de acesso, proteção de dados |
| GDPR | União Europeia | Pseudonimização, controle de acesso, direito ao esquecimento |
| PCI DSS | Global (Pagamentos) | Logging de atividades, separação de ambientes, RBAC |
| SOC 2 | Global | Monitoramento, controles de segurança, gestão de acesso |
| ISO 27001 | Global | Controles de segurança, gestão de riscos, auditoria |

### 10.2 Controles de Segurança Implementados

- **Dados Sensíveis**: Mascaramento automático de PII nos logs
- **Separação de Funções**: RBAC com separação de responsabilidades
- **Audit Trail**: Registro imutável de todas as ações
- **Proteção de Sessão**: Timeouts de inatividade, proteção contra CSRF
- **TLS**: Comunicação criptografada em trânsito
- **Hardening**: Containers mínimos sem componentes desnecessários

## 11. Integrações e Extensões

### 11.1 Integrações com IAM

O Portal integra-se com o IAM para:

- **Autenticação**: Single Sign-On via OIDC
- **Autorização**: Roles e permissões sincronizadas
- **Contexto**: Propagação de tenant/região/ambiente
- **Auditoria**: Logs de acesso consolidados
- **Segurança**: Políticas de acesso consistentes

### 11.2 API para Integração Externa

O Portal expõe uma API REST para integrações externas:

```
BASE URL: https://observability.innovabiz.cloud/api/v1
```

| Endpoint | Método | Descrição |
|----------|--------|-----------|
| `/dashboards` | GET | Listar dashboards |
| `/dashboards/{id}` | GET | Obter dashboard por ID |
| `/alerts` | GET | Listar alertas ativos |
| `/alerts/{id}/silence` | POST | Silenciar alerta |
| `/metrics/query` | POST | Executar query PromQL |
| `/logs/search` | POST | Buscar logs |
| `/traces/search` | POST | Buscar traces |
| `/status` | GET | Status do sistema |

### 11.3 Extensão via Plugins

O Portal suporta extensão via plugins:

- **Fontes de Dados**: Adaptadores para novas fontes
- **Visualizações**: Novos tipos de gráficos e painéis
- **Alertas**: Novos tipos de notificação
- **Integrações**: Conexões com sistemas externos

## 12. Roadmap

### 12.1 Próximas Versões

| Versão | Data Prevista | Principais Recursos |
|--------|--------------|---------------------|
| 1.6.0 | Q3 2025 | Machine Learning para detecção de anomalias |
| 1.7.0 | Q4 2025 | Correlação avançada de eventos com IA |
| 2.0.0 | Q1 2026 | Nova UI com UX aprimorada |
| 2.1.0 | Q2 2026 | Suporte a OpenTelemetry nativo |

### 12.2 Recursos Planejados

#### 12.2.1 Curto Prazo (6 meses)

- Integração nativa com INNOVABIZ AI Platform
- Dashboard builder visual (drag-and-drop)
- Exportação de relatórios automatizada
- Alertas inteligentes com redução de ruído
- Expansão dos dashboards predefinidos

#### 12.2.2 Médio Prazo (12 meses)

- Machine Learning para detecção de anomalias
- Correlação automática de eventos com IA
- Predição de incidentes
- Recomendações automáticas de resolução
- API GraphQL completa

#### 12.2.3 Longo Prazo (24 meses)

- AIOps com automação de resolução
- Natural Language Query para dados de observabilidade
- Observabilidade preditiva
- Digital Twins para simulação de sistemas
- Integração com AR/VR para visualização avançada

## 13. Referências e Documentação Adicional

- [INNOVABIZ Observability Framework](https://docs.innovabiz.cloud/observability-framework/)
- [Kubernetes Oficial Documentation](https://kubernetes.io/docs/)
- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/naming/)
- [Elasticsearch Reference](https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html)
- [Loki Documentation](https://grafana.com/docs/loki/latest/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)