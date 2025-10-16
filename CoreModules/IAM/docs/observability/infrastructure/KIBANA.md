# KIBANA - Interface de Visualização e Analytics

## 1. Visão Geral

O Kibana serve como a interface principal de visualização e análise para logs e métricas armazenados no Elasticsearch dentro da infraestrutura de observabilidade do IAM Audit Service. Esta documentação detalha a implementação, configuração e utilização do Kibana como componente integral da estratégia de observabilidade multi-dimensional da plataforma INNOVABIZ.

### 1.1 Função na Arquitetura de Observabilidade

O Kibana atua como:

- Interface de visualização para dados armazenados no Elasticsearch
- Plataforma de criação de dashboards interativos
- Ferramenta de análise investigativa (logs, métricas, APM)
- Sistema de alertas e notificações baseado em padrões nos dados
- Portal de visualização para eventos de segurança e auditoria

### 1.2 Recursos e Capacidades

- **Discover**: Exploração interativa de logs e eventos em tempo real
- **Dashboard**: Visualizações customizáveis com mais de 20 tipos de gráficos
- **Canvas**: Apresentações dinâmicas baseadas em dados para relatórios
- **Maps**: Visualização geoespacial para eventos com contexto geográfico
- **ML & Analytics**: Detecção de anomalias e análise estatística avançada
- **Alerting**: Sistema robusto de alertas baseado em condições e thresholds
- **Security**: Integração com OIDC e RBAC para controle de acesso granular
- **Multi-Tenancy**: Isolamento completo por tenant via espaços

## 2. Arquitetura de Implantação

### 2.1 Diagrama de Arquitetura

```
                           ┌───────────────────────────────┐
                           │         Load Balancer         │
                           │   (ingress.innovabiz.cloud)   │
                           └───────────────┬───────────────┘
                                           │
                                           ▼
                           ┌───────────────────────────────┐
                           │          OIDC Proxy           │
                           │ (autenticação/autorização)    │
                           └───────────────┬───────────────┘
                                           │
                 ┌─────────────────────────┼─────────────────────────┐
                 │                         │                         │
                 ▼                         ▼                         ▼
        ┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
        │                 │       │                 │       │                 │
        │  Kibana-1       │       │  Kibana-2       │       │  Kibana-3       │
        │  (Réplica)      │◄─────►│  (Réplica)      │◄─────►│  (Réplica)      │
        │                 │       │                 │       │                 │
        └────────┬────────┘       └────────┬────────┘       └────────┬────────┘
                 │                         │                         │
                 │                         │                         │
                 ▼                         ▼                         ▼
        ┌─────────────────────────────────────────────────────────────────────┐
        │                                                                     │
        │                    Cluster Elasticsearch                            │
        │                                                                     │
        └─────────────────────────────────────────────────────────────────────┘
```

### 2.2 Implantação Kubernetes

O Kibana é implantado como um Deployment Kubernetes com as seguintes características:

- **Namespace**: `innovabiz-observability`
- **Deployment**: StatefulSet de 3 réplicas
- **Service**: LoadBalancer interno + Ingress para acesso externo
- **Pods**: Configuração anti-affinity para alta disponibilidade
- **TLS**: Certificados gerenciados via cert-manager
- **PersistentVolume**: Para reports e dados salvos (10GB SSD)

### 2.3 Especificação de Recursos

| Componente | Réplicas | CPU (Request/Limit) | Memória (Request/Limit) | Armazenamento |
|------------|----------|---------------------|-------------------------|--------------|
| Kibana Server | 3 | 1 CPU / 2 CPU | 2Gi / 4Gi | 10Gi (SavedObjects) |
| Ingress Controller | 2 | 0.5 CPU / 1 CPU | 512Mi / 1Gi | N/A |
| OIDC Proxy | 2 | 0.2 CPU / 0.5 CPU | 256Mi / 512Mi | N/A |

*Nota: Escalabilidade automática configurada baseada em utilização de CPU >70%*

### 2.4 Networking e Conectividade

- **Ingress**: `kibana.innovabiz.cloud` (externo), `kibana.observability.svc.cluster.local` (interno)
- **Portas**: 5601 (HTTP/HTTPS)
- **Network Policies**: Acesso restrito de acordo com política zero-trust
- **Comunicação com Elasticsearch**: Canal criptografado via TLS mútuo

### 2.5 Persistência

- **Armazenamento**: PersistentVolumeClaims para armazenamento de SavedObjects
- **Backup**: Snapshots diários para repositório S3
- **Retenção**: 30 dias para snapshots (retenção estendida de 1 ano para compliance)

## 3. Configuração Base

### 3.1 ConfigMap Principal

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kibana-config
  namespace: innovabiz-observability
  labels:
    app: kibana
    component: visualization
    tier: frontend
data:
  kibana.yml: |
    server.name: kibana
    server.host: "0.0.0.0"
    
    # Elasticsearch Connection
    elasticsearch.hosts: ["https://elasticsearch-master.innovabiz-observability.svc.cluster.local:9200"]
    elasticsearch.ssl.verificationMode: certificate
    elasticsearch.ssl.certificateAuthorities: ["/etc/kibana/certs/ca.crt"]
    elasticsearch.ssl.certificate: "/etc/kibana/certs/kibana.crt"
    elasticsearch.ssl.key: "/etc/kibana/certs/kibana.key"
    
    # Authentication
    xpack.security.enabled: true
    xpack.security.authc.providers:
      oidc.oidc1:
        order: 0
        realm: oidc1
        description: "Login com SSO INNOVABIZ"
      basic.basic1:
        order: 1
        icon: "logoKibana"
        hint: "Usuários locais para emergências"
    
    # Multi-Tenancy
    xpack.spaces.enabled: true
    xpack.spaces.maxSpaces: 100
    
    # Telemetry
    telemetry.enabled: false
    telemetry.optIn: false
    
    # Monitoramento
    xpack.monitoring.enabled: true
    xpack.monitoring.collection.enabled: true
    
    # APM Integration
    apm_oss.apmEnabled: true
    apm_oss.serverUrl: "https://apm.innovabiz-observability.svc.cluster.local:8200"
    
    # Reporting
    xpack.reporting.enabled: true
    xpack.reporting.encryptionKey: "${REPORTING_ENCRYPTION_KEY}"
    xpack.reporting.csv.maxSizeBytes: 10485760
    
    # Alerting
    xpack.alerting.enabled: true
    xpack.actions.enabled: true
    xpack.actions.preconfiguredConnectors:
      - id: slack_observability
        name: "Slack Observability"
        actionTypeId: .slack
        config:
          webhookUrl: "${SLACK_WEBHOOK_URL}"
    
    # Timezone e localização
    dateFormat:tz: "America/Sao_Paulo"
    i18n.locale: "pt-BR"
```

### 3.2 Parâmetros de Segurança

Configuração do Secret contendo credenciais e chaves:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kibana-secrets
  namespace: innovabiz-observability
type: Opaque
data:
  ELASTICSEARCH_PASSWORD: ${BASE64_ELASTIC_PASSWORD}
  REPORTING_ENCRYPTION_KEY: ${BASE64_REPORTING_KEY}
  SLACK_WEBHOOK_URL: ${BASE64_SLACK_URL}
  ENCRYPTION_KEY: ${BASE64_ENCRYPTION_KEY}
  APM_TOKEN: ${BASE64_APM_TOKEN}
```

### 3.3 OIDC e Integração SSO

Configuração de integração com provedor OIDC (Keycloak):

```yaml
xpack.security.authc.providers.oidc.oidc1.realm: oidc1
xpack.security.authc.realms.oidc.oidc1:
  order: 0
  realm: oidc1
  issuer: "https://auth.innovabiz.cloud/auth/realms/innovabiz"
  client_id: "kibana-observability"
  client_secret: "${OIDC_CLIENT_SECRET}"
  redirect_uri: "https://kibana.innovabiz.cloud/api/security/oidc/callback"
  post_logout_redirect_uri: "https://kibana.innovabiz.cloud/logged_out"
  claims:
    principal: preferred_username
    groups: groups
    name: name
    email: email
  authorization_endpoint: "https://auth.innovabiz.cloud/auth/realms/innovabiz/protocol/openid-connect/auth"
  token_endpoint: "https://auth.innovabiz.cloud/auth/realms/innovabiz/protocol/openid-connect/token"
  userinfo_endpoint: "https://auth.innovabiz.cloud/auth/realms/innovabiz/protocol/openid-connect/userinfo"
  jwkset_endpoint: "https://auth.innovabiz.cloud/auth/realms/innovabiz/protocol/openid-connect/certs"
  jwks_path: path/to/local/jwks.json
  logout_endpoint: "https://auth.innovabiz.cloud/auth/realms/innovabiz/protocol/openid-connect/logout"
```

### 3.4 Configuração de Proxy

O Kibana é acessado através de um proxy OIDC que adiciona headers de contexto:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kibana-proxy-config
  namespace: innovabiz-observability
data:
  nginx.conf: |
    server {
      listen 8080;
      server_name kibana.innovabiz.cloud;
      
      location / {
        # Adiciona contexto multi-tenant e região
        proxy_set_header X-Tenant-ID $http_x_tenant_id;
        proxy_set_header X-Region $http_x_region;
        proxy_set_header X-Environment $http_x_environment;
        
        # Headers de segurança
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Passa para o Kibana
        proxy_pass http://kibana:5601;
        proxy_read_timeout 90s;
        
        # Limites para uploads de relatórios
        client_max_body_size 50M;
      }
    }
```## 4. Configuração Multi-Dimensional

### 4.1 Estratégia de Multi-Tenancy

O Kibana implementa isolamento multi-tenant através da combinação de:

1. **Spaces**: Kibana Spaces para separação lógica por tenant
2. **Index Patterns**: Padrões de índice específicos por tenant
3. **Field-Level Security**: Segurança em nível de campo para dados compartilhados
4. **RBAC**: Controle de acesso baseado em função para cada tenant
5. **Tenant Context**: Injeção de tenant-id em todas as consultas

#### Exemplo de Criação de Space por Tenant:

```json
POST /_kibana/api/spaces/space
{
  "id": "tenant-${tenant_id}",
  "name": "Tenant ${tenant_name}",
  "description": "Espaço dedicado para o tenant ${tenant_name}",
  "color": "#${tenant_color}",
  "initials": "${tenant_initials}",
  "disabledFeatures": [
    "ml",
    "infrastructure",
    "apm",
    "uptime"
  ]
}
```

### 4.2 Configuração Multi-Regional

O Kibana suporta regionalização através de:

1. **Configuração de Idioma**: Interface localizada por região
2. **Timezone**: Configuração de fuso horário por região
3. **Formatos de Data**: Adaptados aos padrões regionais
4. **Filtros Regionais**: Filtros pré-configurados por região

#### Mapeamento de Regiões:

| Região | Idioma | Timezone | Formato de Data | Formato Numérico |
|--------|--------|----------|-----------------|------------------|
| BR | pt-BR | America/Sao_Paulo | DD/MM/YYYY | 1.000,00 |
| US | en-US | America/New_York | MM/DD/YYYY | 1,000.00 |
| EU | en-GB | Europe/Lisbon | DD/MM/YYYY | 1.000,00 |
| AO | pt-PT | Africa/Luanda | DD/MM/YYYY | 1.000,00 |

### 4.3 Configuração Multi-Ambiente

Cada ambiente (Desenvolvimento, Testes, Homologação, Produção, Sandbox) possui:

1. **Space Dedicado**: Separação completa de dashboards e visualizações
2. **Políticas de Index**: Retenção e lifecycle específicos por ambiente
3. **Alerting**: Regras de alerta adaptadas à criticidade do ambiente
4. **Branding**: Identificação visual clara do ambiente para evitar confusão

#### Script de Criação de Ambientes:

```javascript
// Provisionamento de ambientes via Kibana API
environments.forEach(env => {
  // Criar space para o ambiente
  createSpace(`${tenant_id}-${region}-${env}`, `${tenant_name} ${region} ${env}`, envColors[env]);
  
  // Aplicar template de dashboards
  applyDashboardTemplate(env);
  
  // Configurar alertas específicos
  setupEnvironmentAlerts(env);
  
  // Configurar retenção de índices
  setupIndexLifecyclePolicy(env);
});
```

### 4.4 Labeling e Contextualização

A estratégia de labeling é implementada com:

1. **Filtros Persistentes**: Por tenant, região, ambiente
2. **Templates de Visualização**: Pré-configurados com filtros contextuais
3. **Query Bar Templates**: Consultas pré-definidas com contexto
4. **Field Formatters**: Formatação de campos adaptada ao contexto

Exemplo de filtro multi-dimensional persistente:

```json
{
  "query": {
    "bool": {
      "filter": [
        { "term": { "tenant.id": "${tenant_id}" } },
        { "term": { "region": "${region}" } },
        { "term": { "environment": "${environment}" } },
        { "term": { "module": "IAM" } }
      ]
    }
  }
}
```

## 5. Dashboards e Visualizações

### 5.1 Dashboard Principal de Auditoria

Dashboard centralizado para auditoria IAM com:

- Visão geral de atividade de autenticação/autorização
- Distribuição de eventos por severidade
- Tendências temporais de eventos
- Top usuários por atividade
- Top recursos acessados
- Mapa de acesso geográfico

![Dashboard Auditoria IAM](https://innovabiz.cloud/assets/img/kibana-iam-audit-dashboard.png)

### 5.2 Visualizações Específicas

#### 5.2.1 Autenticação

- Sucesso/Falha de autenticação ao longo do tempo
- Métodos de autenticação utilizados
- Falhas por tipo de erro
- Distribuição geográfica de tentativas
- Análise de tentativas por hora do dia

#### 5.2.2 Autorização

- Acessos negados/permitidos por recurso
- Tendências de escalonamento de privilégios
- Anomalias em padrões de acesso
- Violações de política de acesso
- Top recursos com acesso negado

#### 5.2.3 Segurança

- Eventos suspeitos detectados
- Análise de brute-force
- Sessões simultâneas por usuário
- Alterações em permissões críticas
- Acessos fora do horário normal

#### 5.2.4 Conformidade

- Status de compliance por framework (PCI DSS, GDPR, etc)
- Relatórios periódicos automatizados
- Timeline de alterações em objetos auditados
- Logs de acesso a dados sensíveis
- Alertas de violação de política

### 5.3 Modelos de Visualização

#### 5.3.1 Gráficos de Eventos

```json
{
  "aggs": {
    "timeline": {
      "date_histogram": {
        "field": "@timestamp",
        "fixed_interval": "1h",
        "min_doc_count": 0
      }
    },
    "event_outcome": {
      "terms": {
        "field": "event.outcome",
        "size": 5
      }
    }
  },
  "size": 0,
  "query": {
    "bool": {
      "filter": [
        { "term": { "tenant.id": "${tenant_id}" } },
        { "term": { "event.module": "iam" } },
        { "term": { "event.category": "authentication" } }
      ],
      "must": [
        {
          "range": {
            "@timestamp": {
              "gte": "now-24h",
              "lte": "now"
            }
          }
        }
      ]
    }
  }
}
```

#### 5.3.2 Mapa de Calor de Atividade

```json
{
  "aggs": {
    "activity_by_hour": {
      "date_histogram": {
        "field": "@timestamp",
        "calendar_interval": "hour",
        "format": "HH:mm"
      },
      "aggs": {
        "activity_by_day": {
          "date_histogram": {
            "field": "@timestamp",
            "calendar_interval": "day",
            "format": "yyyy-MM-dd"
          }
        }
      }
    }
  },
  "size": 0,
  "query": {
    "bool": {
      "filter": [
        { "term": { "tenant.id": "${tenant_id}" } },
        { "term": { "event.module": "iam" } }
      ],
      "must": [
        {
          "range": {
            "@timestamp": {
              "gte": "now-30d",
              "lte": "now"
            }
          }
        }
      ]
    }
  }
}
```

### 5.4 Painéis Operacionais

#### 5.4.1 Painel de Status

Dashboard com status operacional do IAM:

- Saúde do serviço por componente
- Taxa de erro em autenticações/autorizações
- Latência de operações críticas
- Capacidade e limites utilizados
- Alertas ativos e recentes

#### 5.4.2 Painel de Performance

Dashboard focado em métricas de performance:

- Tempo de resposta por operação
- Taxa de throughput de autenticações
- Utilização de recursos (CPU, memória, rede)
- Cache hit/miss ratio
- Tempos de resposta de integrações externas

## 6. Segurança e Controle de Acesso

### 6.1 Modelo de RBAC

| Perfil | Descrição | Permissões |
|--------|-----------|------------|
| **Admin Global** | Administradores da plataforma | Acesso total, gerenciamento de usuários |
| **Admin Tenant** | Administrador de tenant específico | Gerenciamento limitado ao tenant |
| **Security Analyst** | Analista de segurança | Visualização e investigação de eventos |
| **Compliance Officer** | Responsável por compliance | Relatórios, alertas de conformidade |
| **Auditor** | Auditoria interna/externa | Acesso somente leitura a logs específicos |
| **DevOps** | Operações e monitoramento | Dashboards operacionais, não eventos |
| **Read Only** | Visualização básica | Dashboards pré-definidos apenas |

### 6.2 Feature Controls

Controle granular de recursos do Kibana:

```yaml
xpack.features:
  actions:
    enabled: true
    showInNavigation: true
    showInMenu: true
  alerting:
    enabled: true
    showInNavigation: true
  apm:
    enabled: false
    showInNavigation: false
  canvas:
    enabled: true
    showInNavigation: true
  dashboard:
    enabled: true
    showInNavigation: true
  dev_tools:
    enabled: false
    showInNavigation: false
  discover:
    enabled: true
    showInNavigation: true
  graphs:
    enabled: true
    showInNavigation: true
  logs:
    enabled: true
    showInNavigation: true
  maps:
    enabled: true
    showInNavigation: true
  ml:
    enabled: false
    showInNavigation: false
  monitoring:
    enabled: true
    showInNavigation: true
  reporting:
    enabled: true
    showInNavigation: true
  savedObjectsManagement:
    enabled: true
    showInNavigation: true
  security:
    enabled: true
    showInNavigation: true
```### 6.3 Autenticação e Autorização

A segurança do Kibana é implementada em múltiplas camadas:

1. **Autenticação via OIDC**: Integração com Keycloak para SSO
2. **Autenticação de Fallback**: Basic auth para casos de emergência
3. **Autorização via Roles**: Mapeamento de grupos OIDC para roles Kibana
4. **Spaces Segregation**: Isolamento por tenant com spaces dedicados
5. **Field-Level Security**: Controle de acesso granular por campo
6. **Document-Level Security**: Restrição a documentos específicos por consulta
7. **Audit Trail**: Registro detalhado de ações no Kibana

#### 6.3.1 Mapeamento de Roles

```yaml
# Role mapping do Kibana para grupos Keycloak
role_mapping:
  - role: kibana_admin
    rules:
      groups: ["innovabiz:global_administrators"]
  - role: tenant_admin
    rules:
      groups: ["innovabiz:${tenant_id}:administrators"]
  - role: security_analyst
    rules:
      groups: ["innovabiz:${tenant_id}:security_analysts"]
  - role: compliance_officer
    rules:
      groups: ["innovabiz:${tenant_id}:compliance_officers"]
  - role: auditor
    rules:
      groups: ["innovabiz:${tenant_id}:auditors"]
  - role: devops_engineer
    rules:
      groups: ["innovabiz:${tenant_id}:devops"]
  - role: read_only
    rules:
      groups: ["innovabiz:${tenant_id}:viewers"]
```

## 7. Integrações

### 7.1 Integração com Elasticsearch

A integração com Elasticsearch inclui:

- **Conexão Segura**: TLS mútuo para todas as comunicações
- **Client Certificate Authentication**: Autenticação por certificado
- **Cross Cluster Search**: Para consultas federadas entre clusters
- **Índices Específicos**: Acesso a índices de auditoria e logging
- **Index Pattern Management**: Padrões de índice otimizados para IAM

### 7.2 Integração com Loki via Grafana

Para visualização unificada entre Elasticsearch e Loki:

1. **Configuração de Data Source**: Kibana como iframe em Grafana
2. **Plugin de Integração**: Iframe seguro com token SSO
3. **Navegação Contextual**: Deep linking entre plataformas
4. **Alerting Federation**: Agregação de alertas das duas plataformas

### 7.3 Integração com APM

Para correlação entre logs e traces:

- **APM Index Pattern**: Padrão de índice específico para APM
- **Transaction Links**: Links diretos para transações nos logs
- **Service Map Integration**: Visualização de serviços e dependências
- **Distributed Tracing**: Correlação de traces entre serviços

### 7.4 Integração com Observability Portal

O Kibana se integra ao Portal de Observabilidade central via:

- **Single Sign-On**: Autenticação unificada
- **API Integration**: Exposição de dashboards via API
- **Embedded Visualizations**: Visualizações embutidas no portal
- **Shared Alerting**: Sistema de alertas compartilhado
- **Context Propagation**: Preservação de contexto entre ferramentas

### 7.5 Exportação de Dados

Mecanismos para exportação de dados:

- **Reporting API**: Geração programática de relatórios
- **CSV Export**: Exportação de resultados de consultas
- **PDF Reports**: Dashboards exportados como PDF
- **Scheduled Reports**: Relatórios agendados por email
- **Integration API**: APIs para consumo por sistemas externos

## 8. Monitoramento e Alertas

### 8.1 Métricas de Saúde

O Kibana expõe métricas sobre seu próprio funcionamento:

```
# HELP kibana_cluster_connected_nodes Número total de nodes conectados
# TYPE kibana_cluster_connected_nodes gauge
kibana_cluster_connected_nodes 3

# HELP kibana_concurrent_connections Número de conexões simultâneas
# TYPE kibana_concurrent_connections gauge
kibana_concurrent_connections{tenant="default"} 42
kibana_concurrent_connections{tenant="tenant-123"} 17

# HELP kibana_request_duration_seconds Duração das requisições em segundos
# TYPE kibana_request_duration_seconds histogram
kibana_request_duration_seconds_bucket{tenant="default",path="/api/security/v1/users",le="0.1"} 123
kibana_request_duration_seconds_bucket{tenant="default",path="/api/security/v1/users",le="0.5"} 185
kibana_request_duration_seconds_bucket{tenant="default",path="/api/security/v1/users",le="1.0"} 193
kibana_request_duration_seconds_sum{tenant="default",path="/api/security/v1/users"} 24.2
kibana_request_duration_seconds_count{tenant="default",path="/api/security/v1/users"} 193

# HELP kibana_response_error_count Contagem de erros nas respostas
# TYPE kibana_response_error_count counter
kibana_response_error_count{tenant="default",path="/api/saved_objects/_find",error="timeout"} 12

# HELP kibana_memory_heap_used_bytes Uso de memória heap
# TYPE kibana_memory_heap_used_bytes gauge
kibana_memory_heap_used_bytes 1073741824
```

### 8.2 Dashboard de Monitoramento

Dashboard específico para monitoramento do Kibana:

- **Saúde do Cluster**: Status dos nodes, latência entre nodes
- **Uso de Recursos**: CPU, memória, heap, conexões de rede
- **Performance de Consultas**: Tempo médio, erros, timeouts
- **Conexões de Usuários**: Total, por tenant, por espaço
- **Cache Stats**: Hit/miss ratio, evictions
- **Logs do Kibana**: Erros, warnings, auditoria

### 8.3 Alertas Configurados

| Nome do Alerta | Descrição | Threshold | Severidade | Notificação |
|----------------|-----------|-----------|------------|-------------|
| **Kibana Node Down** | Node do Kibana indisponível | >1 min | Crítica | Slack, PagerDuty |
| **High CPU Usage** | Uso elevado de CPU | >85% por 5min | Alta | Slack |
| **Memory Pressure** | Pressão de memória | >80% por 5min | Alta | Slack |
| **High Response Time** | Tempo de resposta elevado | >2s p95 | Média | Slack |
| **Error Rate** | Taxa de erros | >5% por 5min | Alta | Slack, Email |
| **Failed Logins** | Falhas de login consecutivas | >10 em 2min | Média | Slack |
| **Saved Objects Limit** | Limite de objetos salvos | >80% | Baixa | Email |
| **Index Pattern Missing** | Padrão de índice ausente | - | Média | Slack |

### 8.4 Regras de Watchdog

Regras de monitoramento automatizado:

```json
{
  "trigger": {
    "schedule": {
      "interval": "1m"
    }
  },
  "input": {
    "search": {
      "request": {
        "search_type": "query_then_fetch",
        "indices": [".monitoring-kibana-*"],
        "body": {
          "size": 0,
          "query": {
            "bool": {
              "filter": [
                {
                  "term": {
                    "kibana_stats.process.memory.heap.size_limit": "{{ ctx.metadata.heap_limit }}"
                  }
                },
                {
                  "range": {
                    "timestamp": {
                      "gte": "now-2m",
                      "lte": "now"
                    }
                  }
                }
              ]
            }
          },
          "aggs": {
            "nodes": {
              "terms": {
                "field": "kibana_stats.kibana.uuid"
              },
              "aggs": {
                "heap_used_percent": {
                  "max": {
                    "script": {
                      "source": "doc['kibana_stats.process.memory.heap.used_in_bytes'].value / doc['kibana_stats.process.memory.heap.size_limit'].value * 100"
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  },
  "condition": {
    "script": {
      "source": "return ctx.payload.aggregations.nodes.buckets.any(node -> node.heap_used_percent.value > params.threshold)",
      "params": {
        "threshold": 85
      }
    }
  },
  "actions": {
    "notify_slack": {
      "slack": {
        "message": {
          "from": "Kibana Monitoring",
          "to": ["#observability-alerts"],
          "text": "Alerta de memória: Kibana com uso de heap > 85%"
        }
      }
    }
  }
}
```

## 9. Backup e Recuperação

### 9.1 Estratégia de Backup

O Kibana mantém seus objetos salvos no Elasticsearch, com a seguinte estratégia:

1. **Snapshot Diário**: Captura completa de índices `.kibana*`
2. **Exportação de Objetos**: Export programático de dashboards críticos
3. **Versionamento**: Objetos salvos com versionamento no git
4. **Repository**: Armazenamento em bucket S3 com retenção configurada

#### 9.1.1 Configuração de Snapshot

```json
PUT _snapshot/kibana_backup
{
  "type": "s3",
  "settings": {
    "bucket": "innovabiz-observability-backup",
    "region": "us-east-1",
    "base_path": "kibana",
    "compress": true
  }
}

PUT _snapshot/kibana_backup/backup-daily-{{ now/d }}
{
  "indices": ".kibana*",
  "ignore_unavailable": true,
  "include_global_state": true
}
```

### 9.2 Procedimentos de Recuperação

Procedimentos documentados para cenários de recuperação:

1. **Restauração Completa**: Restauração total em caso de falha catastrófica
2. **Restauração Seletiva**: Recuperação de objetos específicos
3. **Rollback**: Reversão para estado anterior em caso de problemas
4. **DR Failover**: Procedimento para failover para site DR

#### 9.2.1 Script de Restauração

```bash
#!/bin/bash
# Restaurar Kibana Spaces e Objetos Salvos

# Definir variáveis
SNAPSHOT_NAME="backup-daily-$1"
ELASTICSEARCH_URL="https://elasticsearch-master.innovabiz-observability.svc.cluster.local:9200"
INDEX_PATTERN=".kibana*"

# Fechar índices existentes
curl -XPOST "$ELASTICSEARCH_URL/_cluster/settings" \
  -H 'Content-Type: application/json' \
  -u "elastic:$ELASTIC_PASSWORD" \
  --cacert /path/to/ca.crt \
  -d '{
    "persistent": {
      "cluster.blocks.read_only": true
    }
  }'

# Deletar índices atuais
curl -XDELETE "$ELASTICSEARCH_URL/$INDEX_PATTERN" \
  -H 'Content-Type: application/json' \
  -u "elastic:$ELASTIC_PASSWORD" \
  --cacert /path/to/ca.crt

# Restaurar do snapshot
curl -XPOST "$ELASTICSEARCH_URL/_snapshot/kibana_backup/$SNAPSHOT_NAME/_restore" \
  -H 'Content-Type: application/json' \
  -u "elastic:$ELASTIC_PASSWORD" \
  --cacert /path/to/ca.crt \
  -d '{
    "indices": "'"$INDEX_PATTERN"'",
    "ignore_unavailable": true,
    "include_global_state": true,
    "rename_pattern": ".kibana(.*)",
    "rename_replacement": ".kibana$1"
  }'

# Remover bloqueio de leitura
curl -XPUT "$ELASTICSEARCH_URL/_cluster/settings" \
  -H 'Content-Type: application/json' \
  -u "elastic:$ELASTIC_PASSWORD" \
  --cacert /path/to/ca.crt \
  -d '{
    "persistent": {
      "cluster.blocks.read_only": null
    }
  }'

# Restart do Kibana
kubectl rollout restart deployment kibana -n innovabiz-observability

echo "Restauração do Kibana concluída. Verificando status..."

# Verificar status dos índices
curl -XGET "$ELASTICSEARCH_URL/_cat/indices/$INDEX_PATTERN?v" \
  -u "elastic:$ELASTIC_PASSWORD" \
  --cacert /path/to/ca.crt
```

### 9.3 Retenção de Dados

Políticas de retenção de objetos:

- **Snapshots Diários**: Retidos por 30 dias
- **Snapshots Semanais**: Retidos por 90 dias
- **Snapshots Mensais**: Retidos por 1 ano
- **Snapshots de Conformidade**: Retidos por 5 anos (eventos críticos)

## 10. Conformidade e Segurança

### 10.1 Requisitos de Conformidade

| Regulação | Requisito | Implementação |
|-----------|-----------|---------------|
| **PCI DSS 4.0** | 10.2 Implementação de logs de auditoria | Dashboard de auditoria e alertas |
| **PCI DSS 4.0** | 8.2 Identificação e autenticação | OIDC + MFA |
| **GDPR/LGPD** | Art. 25 Privacy by Design | Field masking, RBAC |
| **GDPR/LGPD** | Art. 30 Registros de atividades | Dashboards de auditoria detalhados |
| **ISO 27001** | A.12.4 Logging and monitoring | Monitoramento completo |
| **ISO 27001** | A.9.2 User access management | RBAC granular |
| **NIST 800-53** | AU-2 Audit Events | Captura eventos requeridos |
| **NIST 800-53** | AU-6 Audit Review, Analysis, and Reporting | Dashboards analíticos |

### 10.2 Controles de Segurança

- **Hardening**: Configuração segura conforme benchmarks CIS
- **TLS**: Criptografia em trânsito para todas comunicações
- **OIDC**: Autenticação centralizada com MFA
- **RBAC**: Controle de acesso granular
- **Session Management**: Timeout, concurrent session limits
- **Audit Trail**: Registro de todas ações administrativas
- **Input Validation**: Proteção contra injeção em consultas
- **Security Headers**: CSP, HSTS, X-Frame-Options

### 10.3 Auditoria

Logs de auditoria são gerados para:

1. **Login/Logout**: Tentativas bem-sucedidas e falhas
2. **Alterações de Configuração**: Modificações em objetos salvos
3. **Consultas de Alto Impacto**: Consultas de alta carga
4. **Exportações de Dados**: Downloads e relatórios
5. **Acessos a Dashboards**: Visualização de dados sensíveis
6. **Ações Administrativas**: Alteração de permissões

```json
{
  "auditTrail": {
    "enabled": true,
    "appender": {
      "type": "index",
      "index": ".kibana-audit-${tenant_id}"
    },
    "ignore_filters": [
      {
        "actions": [
          "http:get:200:"
        ],
        "categories": [
          "api"
        ],
        "types": [
          "rest"
        ]
      }
    ]
  }
}
```

## 11. Operação e Manutenção

### 11.1 Procedimentos Operacionais

- **Deployment**: Procedimento de atualização e rollback
- **Scaling**: Adição de novos nodes ao cluster
- **Monitoramento**: Verificação diária de métricas
- **Manutenção**: Janela mensal para atualizações
- **Troubleshooting**: Procedimentos para cenários comuns

### 11.2 Troubleshooting

| Problema | Possíveis Causas | Resolução |
|----------|-----------------|-----------|
| **Erro 503** | Node Kibana down | Verificar logs, restart pod, escalar |
| **Lentidão nas consultas** | Query complexa, índice grande | Otimizar consulta, verificar ES |
| **Erro de autenticação** | OIDC indisponível | Verificar Keycloak, usar basic auth |
| **Visualização quebrada** | Mudança em mapping | Recriar visualização, ajustar script |
| **Espaço em disco** | Muitos reports ou exports | Limpar temp files, aumentar volume |
| **Consumo alto de memória** | Muitas consultas, leak | Restart, limitar concorrência |

### 11.3 Runbooks

Runbooks detalhados foram criados para cenários comuns:

1. **Inicialização e Verificação**:
   - Validação de configuração
   - Verificação de conectividade
   - Teste de autenticação

2. **Diagnóstico de Performance**:
   - Análise de gargalos
   - Verificação de uso de recursos
   - Otimização de configuração

3. **Recuperação de Falhas**:
   - Procedimentos de restauração
   - Verificação de integridade
   - Validação de objetos salvos

4. **Administração de Usuários**:
   - Criação de novos perfis
   - Ajuste de permissões
   - Auditoria de acessos

## 12. Considerações de Evolução

### 12.1 Roadmap

1. **Curto Prazo** (3-6 meses):
   - Implementação de ML para detecção de anomalias
   - Integração com sistema centralizado de alertas
   - Expansão de dashboards para novos casos de uso

2. **Médio Prazo** (6-12 meses):
   - Implementação de Canvas para relatórios executivos
   - Auto-provisioning de dashboards via API
   - Integração com sistema de ticketing

3. **Longo Prazo** (12+ meses):
   - Implementação de Analytics avançados
   - Federated query across observability tools
   - Observabilidade preditiva com ML

### 12.2 Oportunidades de Melhorias

- **Performance**: Otimização de consultas e caching
- **Usabilidade**: Templates para casos de uso comuns
- **Automação**: Provisionamento automático de dashboards
- **Integração**: Melhor correlação entre logs, métricas e traces
- **Analytics**: Análise avançada para detecção de problemas

## 13. Referências

1. [Kibana Documentation](https://www.elastic.co/guide/en/kibana/current/index.html)
2. [Elasticsearch Documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html)
3. [Observability with the Elastic Stack](https://www.elastic.co/guide/en/observability/current/index.html)
4. [Kibana Security](https://www.elastic.co/guide/en/kibana/current/kibana-security-overview.html)
5. [PCI DSS 4.0 Requirements](https://www.pcisecuritystandards.org/)
6. [GDPR Compliance](https://gdpr.eu/compliance/)
7. [ISO 27001 Information Security](https://www.iso.org/isoiec-27001-information-security.html)
8. [NIST SP 800-53](https://csrc.nist.gov/publications/detail/sp/800-53/rev-5/final)
9. [Multi-Tenancy in Kibana](https://www.elastic.co/guide/en/kibana/current/xpack-spaces.html)
10. [Kibana Alerting](https://www.elastic.co/guide/en/kibana/current/alerting-getting-started.html)

## 14. Anexos

### 14.1 Glossário

| Termo | Definição |
|-------|-----------|
| **Space** | Isolamento lógico de objetos salvos e dashboards |
| **Index Pattern** | Definição de como os índices são acessados |
| **Visualization** | Representação gráfica de dados do Elasticsearch |
| **Dashboard** | Coleção de visualizações para análise conjunta |
| **Saved Object** | Item persistido no Kibana (dashboard, visualização) |
| **Canvas** | Ferramenta de apresentação de dados dinâmicos |
| **Lens** | Criador de visualizações drag-and-drop |
| **KQL** | Kibana Query Language, para filtrar dados |
| **OIDC** | OpenID Connect, protocolo de autenticação |

### 14.2 Template de Objetos Salvos

Exemplo de objeto salvo (dashboard):

```json
{
  "attributes": {
    "title": "IAM Audit Overview",
    "hits": 0,
    "description": "Visão geral de eventos de auditoria IAM",
    "panelsJSON": "[{\"version\":\"8.5.0\",\"type\":\"visualization\",\"gridData\":{\"x\":0,\"y\":0,\"w\":24,\"h\":15,\"i\":\"1\"},\"panelIndex\":\"1\",\"embeddableConfig\":{\"savedVizId\":\"iam-events-timeline\"},\"title\":\"Linha do Tempo de Eventos IAM\"},{\"version\":\"8.5.0\",\"type\":\"visualization\",\"gridData\":{\"x\":24,\"y\":0,\"w\":24,\"h\":15,\"i\":\"2\"},\"panelIndex\":\"2\",\"embeddableConfig\":{\"savedVizId\":\"iam-events-by-type\"},\"title\":\"Eventos por Tipo\"}]",
    "optionsJSON": "{\"hidePanelTitles\":false,\"useMargins\":true}",
    "version": 1,
    "timeRestore": true,
    "timeTo": "now",
    "timeFrom": "now-24h",
    "refreshInterval": {
      "pause": false,
      "value": 300000
    },
    "kibanaSavedObjectMeta": {
      "searchSourceJSON": "{\"query\":{\"language\":\"kuery\",\"query\":\"\"},\"filter\":[{\"$state\":{\"store\":\"appState\"},\"meta\":{\"alias\":null,\"disabled\":false,\"key\":\"event.module\",\"negate\":false,\"params\":{\"query\":\"iam\"},\"type\":\"phrase\"},\"query\":{\"match_phrase\":{\"event.module\":\"iam\"}}}]}"
    }
  },
  "references": [
    {
      "name": "panel_1",
      "type": "visualization",
      "id": "iam-events-timeline"
    },
    {
      "name": "panel_2",
      "type": "visualization",
      "id": "iam-events-by-type"
    }
  ]
}
```

---

*Documento aprovado pelo Comitê de Arquitetura INNOVABIZ em 28/07/2025*