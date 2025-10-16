# INNOVABIZ IAM Audit Service - Documentação Elasticsearch

**Versão:** 7.0.0  
**Data:** 31/07/2025  
**Classificação:** Oficial  
**Status:** ✅ Implementado  
**Autor:** Equipe INNOVABIZ DevOps  
**Aprovado por:** Eduardo Jeremias  

## 1. Visão Geral

O Elasticsearch é o componente central de armazenamento e indexação de logs na arquitetura de observabilidade do IAM Audit Service da INNOVABIZ. Implementado como parte da estratégia "dual-write" de logs (Elasticsearch + Loki), fornece capacidades avançadas de pesquisa, análise e visualização de logs estruturados e não estruturados.

### 1.1 Funcionalidades Principais

- **Armazenamento Distribuído**: Indexação e armazenamento de logs em cluster escalável
- **Pesquisa Full-Text**: Capacidades avançadas de busca e filtragem
- **Análise em Tempo Real**: Aggregations para métricas derivadas de logs
- **Multi-tenant**: Isolamento completo de dados por tenant
- **Conformidade Regulatória**: Retenção configurável e auditoria completa
- **Integração**: Fluentd para coleta, Kibana para visualização

### 1.2 Posicionamento na Arquitetura

O Elasticsearch atua como repositório primário de logs, recebendo dados de:

- Fluentd (coletores de logs)
- Aplicações via API REST
- Filebeat/Logstash para logs específicos
- Audit logs do Kubernetes

E fornecendo dados para:

- Kibana (visualização e análise)
- Portal de Observabilidade (API unificada)
- Sistemas de alerta baseados em conteúdo de logs
- Storage de longo prazo para conformidade regulatória

## 2. Implementação Técnica

### 2.1 Manifesto Kubernetes

O Elasticsearch é implementado como um StatefulSet no Kubernetes, conforme definido em `observability/elasticsearch.yaml`. Os principais componentes incluem:

```yaml
# Trecho exemplificativo do manifesto
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: elasticsearch-master
  namespace: iam-system
  labels:
    app.kubernetes.io/name: elasticsearch
    app.kubernetes.io/part-of: innovabiz-observability
    innovabiz.com/module: iam-audit
    innovabiz.com/tier: observability
spec:
  replicas: 3
  serviceName: elasticsearch
  # ... outras configurações
  template:
    spec:
      securityContext:
        fsGroup: 1000
        runAsUser: 1000
      containers:
      - name: elasticsearch
        image: docker.elastic.co/elasticsearch/elasticsearch:8.10.0
        resources:
          limits:
            cpu: 2000m
            memory: 4Gi
          requests:
            cpu: 1000m
            memory: 2Gi
        env:
        - name: cluster.name
          value: innovabiz-iam-audit
        - name: node.name
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        # ... outras configurações
```

### 2.2 Recursos Computacionais

| Nó | Recurso | Requisito | Limite | Observações |
|----|---------|-----------|--------|-------------|
| **Master** | CPU | 1000m | 2000m | Otimizado para gerenciamento de cluster |
| **Master** | Memória | 2Gi | 4Gi | JVM heap definido em 2Gi |
| **Data** | CPU | 2000m | 4000m | Escalável conforme volume de dados |
| **Data** | Memória | 8Gi | 16Gi | JVM heap definido em 8Gi |
| **Ingest** | CPU | 1000m | 2000m | Para processamento de pipelines |
| **Ingest** | Memória | 2Gi | 4Gi | JVM heap definido em 2Gi |
| **Coordinating** | CPU | 1000m | 2000m | Para balanceamento de consultas |
| **Coordinating** | Memória | 2Gi | 4Gi | JVM heap definido em 2Gi |

### 2.3 Arquitetura de Cluster

- **Topologia**: Separação de funções em nós especializados
  - Master nodes (3x): Gerenciamento do cluster
  - Data nodes (min 3x): Armazenamento de dados, escalável horizontalmente
  - Ingest nodes (2x): Processamento de pipelines
  - Coordinating nodes (2x): Balanceamento de carga de consultas
- **Sharding**: Estratégia adaptativa baseada em volume de dados por tenant
  - Primários: 5 shards por índice padrão
  - Réplicas: 1 réplica por shard (configurável por tenant)
- **Hot-Warm Architecture**: Migração automática de índices antigos
  - Hot nodes: SSDs para índices ativos (7 dias)
  - Warm nodes: HDDs para índices mais antigos (30-90 dias)
  - Cold storage: Object storage para arquivamento (1-7 anos)

### 2.4 Estratégia de Persistência

- **Volumes Persistentes**: StorageClass SSD para nós hot, HDD para warm
- **Snapshots**: Automáticos a cada 6 horas para object storage
- **Retenção**:
  - Hot: 7-15 dias (baseado em SLA do tenant)
  - Warm: 30-90 dias
  - Cold: 1-7 anos (baseado em requisitos regulatórios)
- **ILM (Index Lifecycle Management)**: Políticas por tenant e tipo de log
  - Rollover baseado em tamanho (50GB) ou tempo (1 dia)
  - Migration hot→warm após 7 dias
  - Migration warm→cold após 30-90 dias

### 2.5 Segurança e Controle de Acesso

- **Criptografia**: TLS 1.3 para comunicações em trânsito
- **Autenticação**: 
  - X.509 para comunicação entre nós
  - OAuth2 + OIDC para acesso externo
- **Autorização**:
  - RBAC nativo com integração ao IAM
  - Field-level security para dados sensíveis
- **Isolamento Multi-tenant**:
  - Índices separados por tenant (`tenant-id-*`)
  - Document-level security baseado em tenant_id
- **Auditoria**: Logs completos de acesso e alterações

## 3. Configuração Multi-dimensional

### 3.1 Modelo de Índices Multi-contexto

Estratégia de naming para índices que garante isolamento e performance:

```
{tenant_id}-{module}-{log_type}-{region_id}-{YYYY.MM.DD}
```

Exemplos:
- `tenant1-iam-audit-events-br-east-1-2025.07.31`
- `tenant2-iam-auth-logs-us-west-1-2025.07.31`

### 3.2 Isolamento por Tenant

- **Separação de Dados**: Índices dedicados por tenant
- **Templates Customizados**: Mapeamentos específicos por tenant
- **Política de Recursos**: Limites de shards/índices por tenant
- **SLA Diferenciado**: Performance garantida baseada em tier de serviço
- **Modelo de Custo**: Chargeback baseado em uso real

### 3.3 Contexto Regional

- **Índices Regionais**: Separados por região para conformidade legal
- **Cross-Region Search**: Através do Cross-Cluster Search
- **Compliance**: Garantia de soberania de dados por região
- **Disaster Recovery**: Replicação cross-region para tenants premium

### 3.4 Index Templates

Templates pré-configurados para diferentes tipos de logs:

```json
{
  "index_patterns": ["*-iam-audit-events-*"],
  "template": {
    "settings": {
      "number_of_shards": 5,
      "number_of_replicas": 1,
      "index.mapping.total_fields.limit": 2000,
      "index.refresh_interval": "5s"
    },
    "mappings": {
      "dynamic_templates": [
        {
          "strings_as_keywords": {
            "match_mapping_type": "string",
            "mapping": {
              "type": "keyword",
              "ignore_above": 256,
              "fields": {
                "text": {
                  "type": "text"
                }
              }
            }
          }
        }
      ],
      "properties": {
        "@timestamp": { "type": "date" },
        "tenant_id": { "type": "keyword" },
        "region_id": { "type": "keyword" },
        "environment": { "type": "keyword" },
        "service": { "type": "keyword" },
        "host": { "type": "keyword" },
        "level": { "type": "keyword" },
        "trace_id": { "type": "keyword" },
        "span_id": { "type": "keyword" },
        "user_id": { "type": "keyword" },
        "resource_id": { "type": "keyword" },
        "event": {
          "properties": {
            "type": { "type": "keyword" },
            "action": { "type": "keyword" },
            "outcome": { "type": "keyword" },
            "severity": { "type": "keyword" }
          }
        },
        "message": { "type": "text" }
      }
    }
  }
}
```

## 4. Pipelines de Ingestão

### 4.1 Pipeline de Enriquecimento

```json
{
  "description": "Pipeline padrão para logs de auditoria IAM",
  "processors": [
    {
      "geoip": {
        "field": "source.ip",
        "target_field": "source.geo"
      }
    },
    {
      "user_agent": {
        "field": "user_agent.original",
        "target_field": "user_agent"
      }
    },
    {
      "script": {
        "lang": "painless",
        "source": "ctx.event.risk_score = calculateRiskScore(ctx)"
      }
    },
    {
      "set": {
        "field": "event.ingested",
        "value": "{{_ingest.timestamp}}"
      }
    }
  ]
}
```

### 4.2 Pipeline de Normalização

```json
{
  "description": "Normalização de campos para compatibilidade ECS",
  "processors": [
    {
      "rename": {
        "field": "log_level",
        "target_field": "log.level",
        "ignore_missing": true
      }
    },
    {
      "rename": {
        "field": "req_id",
        "target_field": "trace.id",
        "ignore_missing": true
      }
    },
    {
      "date": {
        "field": "timestamp",
        "formats": ["ISO8601", "yyyy-MM-dd'T'HH:mm:ss.SSSZ"],
        "target_field": "@timestamp"
      }
    }
  ]
}
```

### 4.3 Pipeline de PII Handling

```json
{
  "description": "Tratamento de dados sensíveis para compliance",
  "processors": [
    {
      "redact": {
        "field": "message",
        "patterns": ["\\d{3}\\.\\d{3}\\.\\d{3}-\\d{2}", "\\d{16}"],
        "replacement": "[REDACTED]"
      }
    },
    {
      "script": {
        "lang": "painless",
        "source": "if (ctx.containsKey('pii_fields')) { for (field in ctx.pii_fields) { ctx[field] = maskPII(ctx[field]) } }"
      }
    }
  ]
}
```

## 5. Indexação e Performance

### 5.1 Estratégias de Indexação

- **Bulk Indexing**: Otimizado para alta throughput
  - Tamanho de batch: 5-10MB
  - Concorrência: Ajustada dinamicamente
  - Refresh interval: 5s (balanceamento entre latência e performance)
- **Mapeamentos Otimizados**:
  - Campos keyword para termos exatos
  - Campos text com analisadores específicos para pesquisa
  - Dynamic templates para controle de cardinalidade
- **Aliases**: Utilizados para acesso transparente
  - `tenant1-iam-audit-events-write`: Para escrita no índice atual
  - `tenant1-iam-audit-events-read`: Para leitura (pode apontar para múltiplos índices)
  - `tenant1-iam-audit-events-*`: Padrão de acesso a todos os índices do tipo

### 5.2 Otimizações de Performance

| Configuração | Valor | Propósito |
|--------------|-------|-----------|
| `refresh_interval` | 5s | Balanceia freshness x performance |
| `index.number_of_routing_shards` | 30 | Permite split futuro |
| `index.codec` | best_compression | Otimização de armazenamento |
| `index.queries.cache.enabled` | true | Caching de resultados |
| `indices.fielddata.cache.size` | 20% | Para aggregations |
| `indices.memory.index_buffer_size` | 15% | Performance de indexação |

### 5.3 Políticas de Caching

- **Query Cache**: 10% da heap para resultados de consultas frequentes
- **Fielddata Cache**: 20% da heap para agregações
- **Request Cache**: Ativado para pesquisas read-only
- **Shard Request Cache**: Ativado em índices de baixa mudança

### 5.4 Consultas Otimizadas

```json
// Exemplo de consulta otimizada com filtros pre-filter
{
  "query": {
    "bool": {
      "filter": [
        { "term": { "tenant_id": "tenant1" } },
        { "term": { "region_id": "br-east-1" } },
        { "range": { "@timestamp": { "gte": "now-24h" } } }
      ],
      "must": [
        { "match": { "message": "authentication failed" } }
      ]
    }
  }
}
```

## 6. Monitoramento e Alerta

### 6.1 Métricas Expostas

O Elasticsearch expõe métricas detalhadas através de:

1. **API REST**: `/_nodes/stats`, `/_cluster/health`, `/_cat/*`
2. **Exporter Prometheus**: Métricas formatadas para Prometheus
3. **Metricbeat**: Coleta automática de métricas internas

### 6.2 Métricas Principais

| Métrica | Descrição | Threshold de Alerta |
|---------|-----------|---------------------|
| `cluster_health_status` | Estado geral do cluster | != green por >15min |
| `unassigned_shards` | Shards não atribuídos | >0 por >10min |
| `jvm_heap_used_percent` | Uso de heap JVM | >85% por >5min |
| `disk_used_percent` | Uso de disco | >85% |
| `cpu_usage` | Uso de CPU | >90% por >15min |
| `index_latency` | Latência de indexação | >500ms média por 5min |
| `search_latency` | Latência de busca | >200ms média por 5min |
| `rejected_threads` | Threads rejeitadas | >0 em 5min |
| `index_rate` | Taxa de indexação | Queda >50% em 5min |

### 6.3 Dashboards de Monitoramento

- **Cluster Overview**: Saúde geral, shards, nós, recursos
- **Index Performance**: Taxas de indexação, latência, rejeições
- **Query Performance**: Throughput de buscas, latências, cache hit ratio
- **Resources**: CPU, memória, disco por nó
- **Hot Threads**: Análise de threads com alto consumo
- **Alertas Ativos**: Visão de problemas atuais

### 6.4 Alertas Configurados

```yaml
# Exemplo de alerta para cluster health
- name: ElasticsearchClusterHealth
  rules:
  - alert: ElasticsearchClusterRed
    expr: elasticsearch_cluster_health_status{color="red"} == 1
    for: 5m
    labels:
      severity: critical
      component: elasticsearch
    annotations:
      summary: "Elasticsearch cluster em estado RED"
      description: "O cluster está em estado RED há 5+ minutos, indicando indisponibilidade de dados"
      runbook: "https://docs.innovabiz.com/observability/runbooks/es-cluster-red"
```

## 7. Backup e Recuperação

### 7.1 Estratégia de Backup

- **Snapshots Automáticos**:
  - Frequência: A cada 6 horas
  - Retenção: 30 snapshots (7.5 dias)
  - Tipo: Incremental após o primeiro snapshot completo
  - Armazenamento: Object Storage S3-compatible
- **Snapshot Repository**:
  - Tipo: S3 compatível
  - Compressão: Ativada
  - Encriptação: AES-256

### 7.2 Procedimento de Recuperação

1. **Recuperação Completa do Cluster**:
   - Instanciar novo cluster com manifesto Kubernetes
   - Registrar repository de snapshots
   - Restaurar snapshot mais recente
   - Verificar integridade dos índices
   - Reindexar caso necessário

2. **Recuperação de Índices Específicos**:
   - Identificar snapshot contendo o índice
   - Restaurar apenas os índices necessários
   - Aplicar aliases apropriados
   - Verificar acessibilidade dos dados

3. **Recuperação Point-in-Time**:
   - Utilizar snapshot mais próximo do ponto desejado
   - Restaurar com renomeação para evitar conflitos
   - Aplicar busca por timestamp para filtrar dados

### 7.3 RPO/RTO

| Nível de Serviço | RPO | RTO | Cobertura |
|------------------|-----|-----|-----------|
| **Standard** | 6 horas | 4 horas | Todos os tenants |
| **Premium** | 1 hora | 2 horas | Tenants premium |
| **Enterprise** | 15 minutos | 1 hora | Tenants enterprise |

## 8. Integração com Componentes

### 8.1 Integração com Fluentd

- **Protocolo**: HTTP Bulk API
- **Autenticação**: Certificado cliente + API key
- **Buffer**: Disk-based com retry exponential
- **Batch**: Configurado para 16MB ou 5 segundos
- **Failover**: Secondary endpoint + local spooling

### 8.2 Integração com Kibana

- **Datasource**: Configuração direta
- **Autenticação**: Single Sign-On via OIDC
- **Autorização**: Espaços Kibana por tenant
- **Indexes**: Index patterns pré-configurados
- **Dashboards**: Templates por tipo de serviço

### 8.3 Integração com Portal de Observabilidade

- **API**: GraphQL para consultas federadas
- **Autenticação**: JWT com escopo específico
- **Caching**: Resultados em Redis por 60 segundos
- **Rate Limiting**: Por tenant e tipo de consulta

### 8.4 Integração com sistemas de alertas

- **Consultas Agendadas**: Para alertas baseados em conteúdo
- **Destinos**: AlertManager, ticketing systems, webhooks
- **Formatação**: Templates por tipo de alerta
- **Agregação**: Deduplicação e correlação

## 9. Conformidade e Segurança

### 9.1 Requisitos de Conformidade

| Regulação | Requisito | Implementação |
|-----------|-----------|---------------|
| **PCI DSS 4.0** | 10.2 Implementar trilhas de auditoria | Índices dedicados de auditoria com retenção 1+ ano |
| **GDPR/LGPD** | Art. 17 Direito ao esquecimento | Pipeline de pseudonimização + APIs de exclusão |
| **ISO 27001** | A.12.4 Registros e monitoramento | Auditoria completa de acesso e operações |
| **SOX** | Seção 404 Controles Internos | Imutabilidade de logs de auditoria financeira |
| **NIST 800-53** | SI-4 Monitoramento do sistema | Captura abrangente de eventos de segurança |

### 9.2 Controles de Segurança

- **Criptografia em Trânsito**: TLS 1.3 obrigatório
- **Criptografia em Repouso**: Volumes criptografados + field encryption
- **Controle de Acesso**: RBAC + Field-level security
- **Segregação de Dados**: Isolamento completo por tenant
- **Auditoria**: Logs de acesso e alterações retidos por 1 ano
- **Proteção de PII**: Mascaramento e pseudonimização automáticos

### 9.3 Gerenciamento de Chaves

- **Chaves TLS**: Rotacionadas a cada 90 dias
- **API Keys**: Rotacionadas a cada 30 dias
- **Secrets**: Gerenciados via Kubernetes Secrets / Vault
- **Key Management**: Integração com KMS para chaves de criptografia

## 10. Operação e Manutenção

### 10.1 Procedimentos Operacionais

- **Health Check**: Verificação automatizada a cada 5 minutos
- **Manutenção Preventiva**: Janela semanal para otimizações
- **Análise de Tendências**: Revisão semanal de métricas
- **Capacity Planning**: Forecast mensal baseado em tendências
- **Tuning**: Ajustes baseados em métricas de uso real

### 10.2 Troubleshooting

| Problema | Possíveis Causas | Resolução |
|----------|-----------------|-----------|
| Cluster RED | Shards não atribuídos, falha de nó | Verificar nós falhos, realocar shards |
| Alta latência | Consultas pesadas, memória insuficiente | Otimizar consultas, aumentar recursos |
| Rejected requests | Thread pool esgotado | Aumentar thread pools, implementar backpressure |
| Alta CPU | Consultas complexas, indexação pesada | Otimizar consultas, escalar horizontalmente |
| Uso alto de disco | Muitos dados, shards desbalanceados | Cleanup, rebalanceamento, adicionar nós |

### 10.3 Runbooks

- **Falha de Nó**: Procedimento para substituição de nó
- **Split-brain**: Recuperação de cenário split-brain
- **Performance**: Análise e mitigação de problemas de latência
- **Recuperação de Desastre**: Procedimento completo de DR
- **Manutenção de Rotina**: Index cleanup, optimização, merges forçados

## 11. Considerações de Evolução

### 11.1 Escalabilidade

- **Vertical**: Aumento de recursos por nó até limites recomendados
- **Horizontal**: Adição de nós data para escalar linearmente
- **Sharding**: Estratégia adaptativa baseada em volume e throughput
- **Cross-cluster**: Federação para escala global

### 11.2 Roadmap

1. **Curto Prazo** (3 meses):
   - Implementação de cross-cluster search
   - Otimização de pipelines de ingestão
   - Melhorias em dashboards de monitoramento

2. **Médio Prazo** (6-12 meses):
   - Machine learning para detecção de anomalias
   - Arquitetura frozen para dados históricos
   - Integração melhorada com observabilidade multidimensional

3. **Longo Prazo** (12+ meses):
   - Federação global com latência reduzida
   - Automação completa de operações
   - Search-as-a-service para outros módulos

## 12. Referências

1. [Elasticsearch Official Documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html)
2. [Elasticsearch on Kubernetes Best Practices](https://www.elastic.co/blog/kubernetes-deployment-best-practices)
3. [PCI DSS 4.0 Logging Requirements](https://www.pcisecuritystandards.org/)
4. [NIST SP 800-53 Rev. 5](https://csrc.nist.gov/publications/detail/sp/800-53/rev-5/final)
5. [Elasticsearch: The Definitive Guide](https://www.elastic.co/guide/en/elasticsearch/guide/current/index.html)
6. [ElastAlert: Alerting With Elasticsearch](https://github.com/Yelp/elastalert)
7. [Designing Elasticsearch for Scale](https://www.elastic.co/blog/designing-elasticsearch-for-scale)

---

*Documento aprovado pelo Comitê de Arquitetura INNOVABIZ em 28/07/2025*