# Documentação Técnica: Dashboard ClickHouse INNOVABIZ

![Status](https://img.shields.io/badge/Status-Implementado-brightgreen)
![Versão](https://img.shields.io/badge/Versão-1.0.0-blue)
![Plataforma](https://img.shields.io/badge/Plataforma-INNOVABIZ-purple)
![Compliance](https://img.shields.io/badge/Compliance-ISO%2027001%20|%20PCI%20DSS%20|%20LGPD-orange)

## 1. Visão Geral

### 1.1 Objetivos do Dashboard

O dashboard ClickHouse INNOVABIZ foi desenvolvido para fornecer monitoramento abrangente de instâncias ClickHouse na plataforma, permitindo a observabilidade completa deste componente crítico do sistema analítico. O ClickHouse é utilizado como um data warehouse analítico de alto desempenho para processamento em tempo real de grandes volumes de dados, sendo fundamental para as operações analíticas da plataforma INNOVABIZ.

Este dashboard permite:
- Monitorar a saúde e disponibilidade de instâncias ClickHouse
- Visualizar métricas de performance de consultas e latência
- Acompanhar o uso de recursos (CPU, memória, disco)
- Identificar gargalos de performance e problemas de replicação
- Observar comportamentos anômalos que possam indicar falhas iminentes
- Manter visibilidade completa sobre as métricas do sistema analítico através dos vários contextos da plataforma

### 1.2 Arquitetura Multi-Contexto

O dashboard foi projetado seguindo a arquitetura multi-contexto da plataforma INNOVABIZ, permitindo filtrar todas as métricas por:

- **Tenant (tenant_id)**: Isolamento por tenant para ambientes multi-tenant
- **Região (region_id)**: Divisão geográfica para conformidade e baixa latência
- **Ambiente (environment)**: Segmentação por ambientes (prod, stage, dev)
- **Instância (instance)**: Seleção de servidores específicos
- **Banco de Dados (database)**: Filtro por bancos de dados específicos

Este modelo garante que operadores, administradores e stakeholders possam visualizar métricas relevantes para seus respectivos contextos de operação, mantendo conformidade com requisitos de governança e segregação de dados.

## 2. Estrutura do Dashboard

### 2.1 Seções Principais

O dashboard está organizado em quatro seções principais para facilitar o monitoramento e troubleshooting:

1. **Visão Geral**
   - Status da instância (online/offline)
   - Uptime
   - Número de tabelas e bancos de dados
   - Taxa de conexões (HTTP e TCP)
   - Processos ativos e em execução

2. **Métricas de Consulta**
   - Taxa de consultas por tipo (SELECT, INSERT, total, falhas)
   - Latência de consultas (média, p95, p99)
   - Threads aguardando por mutex
   - Tarefas em background

3. **Recursos e Performance**
   - Utilização de CPU
   - Utilização de memória
   - Uso de disco por disco
   - Throughput de I/O (leitura/escrita)

4. **Replicação e Sistema**
   - Status de replicação
   - Réplicas somente leitura
   - Expiração de sessão ZooKeeper
   - Eventos do sistema

### 2.2 Anotações

O dashboard inclui anotações automáticas para eventos críticos:

- **Falhas de Conexão Distribuída**: Detecta problemas na comunicação entre nós do cluster
- **Exceções ZooKeeper**: Identifica problemas de coordenação do cluster

### 2.3 Variáveis e Filtros

O dashboard implementa variáveis cascateadas para filtragem multi-contexto:

| Variável | Descrição | Dependências |
|----------|-----------|--------------|
| tenant_id | ID do tenant | Nenhuma |
| region_id | ID da região | tenant_id |
| environment | Ambiente (prod, stage, dev, etc.) | tenant_id, region_id |
| instance | Instância específica do ClickHouse | tenant_id, region_id, environment |
| database | Banco de dados específico | tenant_id, region_id, environment, instance |

As variáveis são configuradas para permitir seleção múltipla e incluir a opção "All" (todos), facilitando a navegação de contextos gerais para específicos.

## 3. Requisitos e Implementação

### 3.1 Métricas Prometheus Necessárias

O dashboard requer as seguintes métricas exportadas pelo ClickHouse-exporter:

**Métricas básicas:**
- `clickhouse_uptime_seconds`
- `clickhouse_tables`
- `clickhouse_databases`
- `clickhouse_processes_total`
- `clickhouse_processes_running`
- `up{job="clickhouse"}`

**Métricas de consulta:**
- `clickhouse_query_total`
- `clickhouse_select_query_total`
- `clickhouse_insert_query_total`
- `clickhouse_failed_query_total`
- `clickhouse_query_time_ms_sum`
- `clickhouse_query_time_ms_count`
- `clickhouse_query_duration_seconds_bucket`

**Métricas de recursos:**
- `node_cpu_seconds_total`
- `node_memory_MemTotal_bytes`
- `node_memory_MemFree_bytes`
- `node_memory_Cached_bytes`
- `node_memory_Buffers_bytes`
- `clickhouse_disk_data_bytes`
- `clickhouse_disk_free_bytes`
- `clickhouse_disk_total_bytes`
- `clickhouse_read_bytes`
- `clickhouse_written_bytes`

**Métricas de conexão:**
- `clickhouse_http_requests_total`
- `clickhouse_tcp_connections_total`

**Métricas de replicação:**
- `clickhouse_replicas`
- `clickhouse_replicas_readonly`
- `clickhouse_zookeeper_session_expiration`
- `clickhouse_events_total`
- `clickhouse_events_DistributedConnectionFailTry`
- `clickhouse_events_ZooKeeperExceptions`

**Métricas de performance:**
- `clickhouse_mutex_lock_waiting`
- `clickhouse_background_pool_tasks`
- `clickhouse_background_processing_pool_tasks`
- `clickhouse_distributed_processing_pool_tasks`

### 3.2 Requisitos de Labels

Para compatibilidade com a arquitetura multi-contexto INNOVABIZ, todas as métricas devem incluir os seguintes labels:

```yaml
- tenant_id: "<identificador_do_tenant>"
- region_id: "<identificador_da_região>"
- environment: "<ambiente>"
- instance: "<host>:<porta>"
```

### 3.3 Exporters e Configuração

#### 3.3.1 Configuração do ClickHouse-exporter

O ClickHouse-exporter deve ser configurado em cada servidor ClickHouse:

```yaml
# clickhouse-exporter-config.yml
endpoint: http://localhost:8123
credentials:
  user: prometheus
  password: <password_seguro>
metrics:
  collect_timeout: 30s
  namespace: clickhouse
  scrape_interval: 15s
labels:
  static:
    tenant_id: "${TENANT_ID}"
    region_id: "${REGION_ID}"
    environment: "${ENVIRONMENT}"
```

#### 3.3.2 Configuração do Prometheus

Adicione o seguinte scrape config ao Prometheus:

```yaml
scrape_configs:
  - job_name: 'clickhouse'
    scrape_interval: 15s
    metrics_path: '/metrics'
    static_configs:
      - targets: ['clickhouse-exporter:9116']
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
      - source_labels: [__meta_kubernetes_pod_label_tenant_id]
        target_label: tenant_id
      - source_labels: [__meta_kubernetes_pod_label_region_id]
        target_label: region_id
      - source_labels: [__meta_kubernetes_pod_label_environment]
        target_label: environment
```

### 3.4 Configuração do Grafana

Para importar o dashboard:

1. Acesse Grafana > Dashboards > Import
2. Carregue o arquivo JSON do dashboard
3. Selecione a fonte de dados Prometheus
4. Configure permissões conforme políticas de RBAC da plataforma INNOVABIZ
5. Salve o dashboard

## 4. Casos de Uso Operacional

### 4.1 Monitoramento em Tempo Real

**Cenário**: Acompanhamento contínuo do estado e performance do ClickHouse

**Painéis relevantes**:
- Status da instância
- Taxa de consultas por tipo
- Utilização de CPU e memória
- Throughput de I/O

**Procedimento**:
1. Verifique o status online/offline das instâncias
2. Observe a taxa de consultas para detectar padrões anormais
3. Monitore o uso de recursos para identificar possíveis saturações
4. Configure o intervalo de atualização automática para 30 segundos

### 4.2 Troubleshooting de Performance

**Cenário**: Investigação de lentidão reportada em consultas analíticas

**Painéis relevantes**:
- Latência de consultas
- Threads aguardando por mutex
- Tarefas em background
- Eventos do sistema

**Procedimento**:
1. Analise os percentis de latência (p95, p99) para determinar se o problema é generalizado ou em consultas específicas
2. Verifique se há threads esperando por mutex, indicando possíveis bloqueios
3. Observe se há aumento anormal de tarefas em background
4. Correlacione com eventos do sistema para identificar possíveis causas
5. Analise o throughput de I/O para verificar se há gargalos de disco

### 4.3 Planejamento de Capacidade

**Cenário**: Avaliar necessidade de expansão de recursos para o ClickHouse

**Painéis relevantes**:
- Uso de disco por disco
- Utilização de CPU e memória
- Número de tabelas e bancos de dados
- Taxa de conexões

**Procedimento**:
1. Analise tendências de crescimento no uso de disco
2. Correlacione com aumento no número de tabelas/bancos de dados
3. Verifique se a CPU está consistentemente acima de 70% ou memória acima de 85%
4. Observe tendências nas taxas de conexão e número de consultas
5. Utilize período de visualização de 7-30 dias para identificar tendências

### 4.4 Validação de Alta Disponibilidade

**Cenário**: Verificar integridade da replicação e alta disponibilidade do cluster

**Painéis relevantes**:
- Status de replicação
- Réplicas somente leitura
- Falhas de conexão distribuída (anotações)
- Exceções ZooKeeper (anotações)

**Procedimento**:
1. Verifique o número total de réplicas vs. réplicas somente leitura
2. Observe presença de anotações indicando falhas de conexão distribuída
3. Monitore exceções ZooKeeper que possam indicar problemas de coordenação
4. Verifique métricas de expiração de sessão ZooKeeper

## 5. Governança e Compliance

### 5.1 Requisitos de Segurança

O dashboard foi projetado considerando requisitos de segurança em conformidade com:

- **ISO 27001**: Controles de acesso e monitoramento de ativos de informação
- **PCI DSS**: Requisitos 10.1-10.3 para rastreamento de atividades e 2.2 para hardening
- **GDPR/LGPD**: Segregação de dados por tenant e região para conformidade com legislações de privacidade

### 5.2 Controle de Acesso

Recomenda-se a seguinte matriz de controle de acesso ao dashboard:

| Perfil | Permissão | Escopo |
|--------|-----------|--------|
| Operador NOC | Visualização | Todos os tenants/regiões |
| SRE/DevOps | Visualização | Todos os tenants/regiões |
| Admin Tenant | Visualização | Tenant específico |
| Analista de Segurança | Visualização | Todos os tenants/regiões |
| Dev/QA | Visualização | Apenas ambientes não-produtivos |

O controle de acesso deve ser implementado via integração com o IAM da plataforma INNOVABIZ, garantindo que os usuários só visualizem dados aos quais têm permissão.

### 5.3 Auditoria e Rastreabilidade

Todas as interações com o dashboard devem ser registradas no sistema de auditoria centralizado, incluindo:

- Quem acessou o dashboard
- Quais filtros foram aplicados
- Quando o acesso ocorreu
- Ações realizadas (exportação de dados, configurações alteradas)

Estes logs devem ser retidos conforme política de retenção da organização e requisitos regulatórios aplicáveis.

## 6. Alertas Recomendados

Baseado nas métricas visualizadas neste dashboard, recomenda-se configurar os seguintes alertas no Prometheus AlertManager:

### 6.1 Alertas de Disponibilidade

```yaml
- alert: ClickHouseInstanceDown
  expr: up{job="clickhouse"} == 0
  for: 1m
  labels:
    severity: critical
    category: availability
  annotations:
    summary: "ClickHouse Instance Down"
    description: "A instância {{ $labels.instance }} está offline por pelo menos 1 minuto"
```

### 6.2 Alertas de Performance

```yaml
- alert: ClickHouseHighQueryLatency
  expr: rate(clickhouse_query_time_ms_sum[5m]) / rate(clickhouse_query_time_ms_count[5m]) > 1000
  for: 5m
  labels:
    severity: warning
    category: performance
  annotations:
    summary: "Alta Latência em Consultas ClickHouse"
    description: "A latência média de consultas está acima de 1000ms por 5 minutos na instância {{ $labels.instance }}"
```

### 6.3 Alertas de Recursos

```yaml
- alert: ClickHouseDiskSpaceCritical
  expr: clickhouse_disk_free_bytes / clickhouse_disk_total_bytes * 100 < 10
  for: 5m
  labels:
    severity: critical
    category: resource
  annotations:
    summary: "Espaço em Disco Crítico"
    description: "Menos de 10% de espaço livre no disco {{ $labels.disk }} da instância {{ $labels.instance }}"
```

### 6.4 Alertas de Replicação

```yaml
- alert: ClickHouseHighReadOnlyReplicas
  expr: clickhouse_replicas_readonly / clickhouse_replicas * 100 > 50
  for: 5m
  labels:
    severity: warning
    category: replication
  annotations:
    summary: "Alto Número de Réplicas Somente-Leitura"
    description: "Mais de 50% das réplicas estão em modo somente leitura na instância {{ $labels.instance }}"
```

## 7. Integração com Outros Dashboards

Este dashboard se integra com outros dashboards da plataforma INNOVABIZ através das seguintes conexões:

### 7.1 Dashboards Relacionados

- **Dashboard de Infraestrutura**: Para correlacionar métricas de infraestrutura com comportamento do ClickHouse
- **Dashboard de APIs**: Para correlacionar consultas analíticas com chamadas de API
- **Dashboard Alerting**: Para visualização consolidada de alertas relacionados ao ClickHouse
- **Dashboard Multi-Contexto**: Para visão holística de todos os componentes por tenant/região

### 7.2 Navegação Entre Dashboards

Links diretos são fornecidos no dashboard para navegação contextualizada entre sistemas relacionados:

- Link para dashboard de infraestrutura com contexto da instância atual
- Link para dashboard de alertas filtrado para alertas ClickHouse
- Link para logs de sistema relacionados às instâncias ClickHouse

## 8. Melhorias Futuras

### 8.1 Próximas Iterações

- Adicionar métricas específicas de dicionários
- Implementar métricas de profile de consultas para identificação de consultas lentas
- Integrar métricas de qualidade de dados
- Expandir visualizações de réplicas e shards
- Adicionar métricas de compactação e garbage collection

### 8.2 Integrações Planejadas

- Integração com sistema de tracing distribuído para correlação de consultas
- Implementação de deep-links para logs específicos
- Correlação automática com incidentes
- Análise preditiva de tendências de uso e performance

## 9. Referências e Recursos Adicionais

- [Documentação Oficial ClickHouse](https://clickhouse.com/docs/)
- [Guia de Observabilidade INNOVABIZ](https://wiki.innovabiz.com/observability-guide)
- [Especificação do ClickHouse-exporter](https://github.com/ClickHouse/clickhouse_exporter)
- [RFC INNOVABIZ: Padrões de Monitoramento Multi-Contexto](https://wiki.innovabiz.com/rfc/monitoring-standards)
- [Requisitos de Governança INNOVABIZ](https://wiki.innovabiz.com/governance)

---

**Autor**: Equipe de Plataforma INNOVABIZ  
**Criado**: Fevereiro 2025  
**Última Atualização**: Fevereiro 2025  
**Revisão Programada**: Agosto 2025  
**Classificação**: Interno - Confidencial