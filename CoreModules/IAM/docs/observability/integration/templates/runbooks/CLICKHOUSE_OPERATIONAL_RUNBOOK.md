# Runbook Operacional: ClickHouse INNOVABIZ

![Status](https://img.shields.io/badge/Status-Ativo-brightgreen)
![Versão](https://img.shields.io/badge/Versão-1.0.0-blue)
![Plataforma](https://img.shields.io/badge/Plataforma-INNOVABIZ-purple)
![Compliance](https://img.shields.io/badge/Compliance-ISO%2027001%20|%20PCI%20DSS%20|%20LGPD-orange)

## 1. Visão Geral

### 1.1 Propósito do Documento

Este runbook operacional fornece procedimentos detalhados para operação, monitoramento, troubleshooting e manutenção de instâncias ClickHouse na plataforma INNOVABIZ. O documento foi desenvolvido seguindo as melhores práticas internacionais de SRE, DevOps e governança corporativa, garantindo conformidade com normas ISO 27001, PCI DSS 4.0, GDPR/LGPD e frameworks ITIL v4.

### 1.2 Escopo e Aplicabilidade

Este runbook aplica-se a todas as instâncias ClickHouse operadas na plataforma INNOVABIZ, incluindo:

- **Ambientes**: Produção, Homologação, Desenvolvimento, Sandbox
- **Topologias**: Single-node, Cluster distribuído, Réplicas
- **Contextos**: Multi-tenant, Multi-região, Multi-ambiente
- **Integrações**: Kafka, PostgreSQL, Redis, APIs, ML Pipelines

### 1.3 Arquitetura ClickHouse na Plataforma INNOVABIZ

O ClickHouse é implementado como data warehouse analítico de alto desempenho, processando grandes volumes de dados em tempo real para suporte a decisões de negócio, analytics avançados e machine learning.

**Componentes principais:**
- **ClickHouse Server**: Motor principal de processamento analítico
- **ZooKeeper**: Coordenação de cluster e replicação
- **ClickHouse Keeper**: Alternativa nativa ao ZooKeeper (em implementação)
- **Exporters**: Prometheus exporters para métricas
- **Load Balancers**: Distribuição de carga entre nós

**Integrações críticas:**
- **Kafka**: Ingestão de dados em tempo real
- **PostgreSQL**: Sincronização de dados transacionais
- **Redis**: Cache de consultas frequentes
- **API Gateway**: Exposição de endpoints analíticos
- **ML Pipelines**: Alimentação de modelos de machine learning

## 2. Monitoramento e Alertas

### 2.1 Métricas Críticas

#### 2.1.1 Disponibilidade
- **up{job="clickhouse"}**: Status da instância (0/1)
- **clickhouse_uptime_seconds**: Tempo de atividade
- **clickhouse_http_connections**: Conexões HTTP ativas
- **clickhouse_tcp_connections**: Conexões TCP ativas

#### 2.1.2 Performance
- **clickhouse_query_total**: Total de consultas executadas
- **clickhouse_query_time_ms_sum/count**: Latência média de consultas
- **clickhouse_query_duration_seconds_bucket**: Histograma de latência
- **clickhouse_failed_query_total**: Consultas com falha

#### 2.1.3 Recursos
- **node_cpu_seconds_total**: Utilização de CPU
- **node_memory_***: Métricas de memória
- **clickhouse_disk_***: Utilização de disco
- **clickhouse_read_bytes/written_bytes**: Throughput de I/O

#### 2.1.4 Replicação
- **clickhouse_replicas**: Número total de réplicas
- **clickhouse_replicas_readonly**: Réplicas somente leitura
- **clickhouse_zookeeper_session_expiration**: Expiração de sessão ZK

### 2.2 Thresholds de Alerta

| Métrica | Warning | Critical | Duração |
|---------|---------|----------|---------|
| Instância Down | N/A | up == 0 | 1m |
| CPU Usage | > 80% | > 90% | 15m |
| Memory Usage | > 85% | > 95% | 15m |
| Disk Space | < 20% | < 10% | 15m/5m |
| Query Latency | > 1s | > 3s | 5m |
| Failed Queries | > 5% | > 10% | 5m |
| Readonly Replicas | > 50% | > 80% | 10m |

### 2.3 Dashboards Relacionados

- **Dashboard Principal**: [ClickHouse INNOVABIZ](https://grafana.innovabiz.com/d/innovabiz-clickhouse)
- **Dashboard de Infraestrutura**: [Infraestrutura Multi-Contexto](https://grafana.innovabiz.com/d/innovabiz-infrastructure)
- **Dashboard de Alertas**: [AlertManager INNOVABIZ](https://grafana.innovabiz.com/d/innovabiz-alertmanager)

## 3. Procedimentos Operacionais

### 3.1 Verificações de Saúde

#### 3.1.1 Verificação Básica de Status

```bash
# Verificar se o serviço está rodando
sudo systemctl status clickhouse-server

# Verificar conectividade HTTP
curl -s "http://clickhouse-server:8123/ping"

# Verificar conectividade TCP
echo "SELECT 1" | clickhouse-client --host clickhouse-server --port 9000

# Verificar versão
clickhouse-client --query "SELECT version()"
```

#### 3.1.2 Verificação de Cluster

```sql
-- Verificar status do cluster
SELECT * FROM system.clusters;

-- Verificar réplicas
SELECT * FROM system.replicas;

-- Verificar processos ativos
SELECT * FROM system.processes;

-- Verificar métricas do sistema
SELECT * FROM system.metrics WHERE metric LIKE '%Query%';
```

#### 3.1.3 Verificação de ZooKeeper

```bash
# Verificar conectividade com ZooKeeper
echo "ls /" | zkCli.sh -server zookeeper:2181

# Verificar nós ClickHouse no ZooKeeper
echo "ls /clickhouse" | zkCli.sh -server zookeeper:2181
```

### 3.2 Troubleshooting por Tipo de Alerta

#### 3.2.1 ClickHouseInstanceDown

**Sintomas:**
- Instância não responde a conexões
- Dashboard mostra status offline
- Aplicações reportam erros de conexão

**Diagnóstico:**
```bash
# Verificar logs do ClickHouse
sudo tail -f /var/log/clickhouse-server/clickhouse-server.log
sudo tail -f /var/log/clickhouse-server/clickhouse-server.err.log

# Verificar status do processo
ps aux | grep clickhouse

# Verificar uso de recursos
top -p $(pgrep clickhouse-server)

# Verificar conectividade de rede
netstat -tlnp | grep :8123
netstat -tlnp | grep :9000
```

**Ações de Remediação:**
```bash
# Tentar restart do serviço
sudo systemctl restart clickhouse-server

# Se falhar, verificar configuração
sudo clickhouse-server --config-file=/etc/clickhouse-server/config.xml --check-config

# Verificar permissões de arquivos
sudo chown -R clickhouse:clickhouse /var/lib/clickhouse
sudo chown -R clickhouse:clickhouse /var/log/clickhouse-server

# Em caso extremo, restart completo
sudo systemctl stop clickhouse-server
sudo systemctl start clickhouse-server
```

#### 3.2.2 ClickHouseSlowQueries

**Sintomas:**
- Latência média de consultas > 1000ms
- Aplicações reportam timeouts
- Usuários reportam lentidão

**Diagnóstico:**
```sql
-- Identificar consultas lentas
SELECT 
    query,
    user,
    query_duration_ms,
    memory_usage,
    read_rows,
    written_rows
FROM system.query_log 
WHERE query_duration_ms > 1000 
ORDER BY query_duration_ms DESC 
LIMIT 10;

-- Verificar processos em execução
SELECT 
    query_id,
    user,
    query,
    elapsed,
    memory_usage
FROM system.processes 
ORDER BY elapsed DESC;

-- Verificar uso de índices
SELECT 
    table,
    name,
    type,
    granularity
FROM system.data_skipping_indices;
```

**Ações de Remediação:**
```sql
-- Matar consultas problemáticas
KILL QUERY WHERE query_id = 'query_id_problematico';

-- Otimizar tabelas
OPTIMIZE TABLE nome_da_tabela FINAL;

-- Verificar e criar índices necessários
ALTER TABLE nome_da_tabela ADD INDEX idx_nome (coluna) TYPE minmax GRANULARITY 1;
```

#### 3.2.3 ClickHouseDiskSpaceCritical

**Sintomas:**
- Espaço em disco < 10%
- Falhas em inserções
- Alertas críticos de armazenamento

**Diagnóstico:**
```bash
# Verificar uso de disco
df -h /var/lib/clickhouse

# Identificar tabelas que mais consomem espaço
du -sh /var/lib/clickhouse/data/*/* | sort -hr | head -20

# Verificar logs de tamanho
clickhouse-client --query "
SELECT 
    database,
    table,
    formatReadableSize(sum(bytes)) as size
FROM system.parts 
GROUP BY database, table 
ORDER BY sum(bytes) DESC 
LIMIT 10"
```

**Ações de Remediação:**
```sql
-- Limpar dados antigos baseado em TTL
ALTER TABLE nome_da_tabela MODIFY TTL data_coluna + INTERVAL 30 DAY;

-- Executar limpeza manual
OPTIMIZE TABLE nome_da_tabela FINAL;

-- Dropar partições antigas
ALTER TABLE nome_da_tabela DROP PARTITION 'partição_antiga';

-- Comprimir dados
ALTER TABLE nome_da_tabela MODIFY COLUMN coluna CODEC(ZSTD);
```

#### 3.2.4 ClickHouseReplicaReadOnly

**Sintomas:**
- Réplicas em modo somente leitura
- Falhas em inserções em réplicas
- Inconsistências entre nós

**Diagnóstico:**
```sql
-- Verificar status das réplicas
SELECT 
    database,
    table,
    replica_name,
    is_readonly,
    absolute_delay,
    queue_size
FROM system.replicas;

-- Verificar logs de replicação
SELECT * FROM system.replication_queue;

-- Verificar conectividade ZooKeeper
SELECT * FROM system.zookeeper WHERE path = '/';
```

**Ações de Remediação:**
```sql
-- Forçar sincronização de réplica
SYSTEM SYNC REPLICA nome_da_tabela;

-- Reinicializar réplica problemática
SYSTEM RESTART REPLICA nome_da_tabela;

-- Em casos extremos, recriar réplica
DETACH TABLE nome_da_tabela;
ATTACH TABLE nome_da_tabela;
```

### 3.3 Operações de Rotina

#### 3.3.1 Backup e Restore

```bash
# Backup completo
clickhouse-backup create backup_$(date +%Y%m%d_%H%M%S)

# Listar backups
clickhouse-backup list

# Restore de backup
clickhouse-backup restore backup_20250131_140000

# Backup incremental
clickhouse-backup create_remote backup_incremental_$(date +%Y%m%d_%H%M%S)
```

#### 3.3.2 Manutenção de Tabelas

```sql
-- Otimização de tabelas
OPTIMIZE TABLE nome_da_tabela FINAL;

-- Verificar integridade
CHECK TABLE nome_da_tabela;

-- Recomputar estatísticas
ANALYZE TABLE nome_da_tabela;

-- Limpeza de dados antigos
ALTER TABLE nome_da_tabela DELETE WHERE data_coluna < today() - 90;
```

#### 3.3.3 Monitoramento de Performance

```sql
-- Top consultas por tempo
SELECT 
    query,
    count() as executions,
    avg(query_duration_ms) as avg_duration,
    max(query_duration_ms) as max_duration
FROM system.query_log 
WHERE event_date >= today() - 1
GROUP BY query 
ORDER BY avg_duration DESC 
LIMIT 10;

-- Uso de memória por consulta
SELECT 
    user,
    query_id,
    memory_usage,
    peak_memory_usage
FROM system.query_log 
WHERE event_date >= today()
ORDER BY peak_memory_usage DESC 
LIMIT 10;
```

## 4. Recuperação de Desastres

### 4.1 Cenários de Falha

#### 4.1.1 Falha de Nó Único

**Procedimento:**
1. Identificar causa da falha
2. Tentar restart do serviço
3. Se necessário, migrar para nó backup
4. Restaurar dados do backup mais recente
5. Sincronizar com cluster

#### 4.1.2 Falha de Cluster Completo

**Procedimento:**
1. Avaliar extensão da falha
2. Verificar integridade dos dados
3. Restaurar nós em ordem de prioridade
4. Reestabelecer replicação
5. Validar consistência dos dados

#### 4.1.3 Corrupção de Dados

**Procedimento:**
1. Isolar nós afetados
2. Verificar backups disponíveis
3. Restaurar dados do último backup válido
4. Reprocessar dados perdidos
5. Validar integridade completa

### 4.2 RTO e RPO

- **RTO (Recovery Time Objective)**: 4 horas
- **RPO (Recovery Point Objective)**: 1 hora
- **Backup Frequency**: A cada 6 horas
- **Retenção**: 30 dias local, 90 dias remoto

## 5. Configurações Recomendadas

### 5.1 Configuração de Servidor

```xml
<!-- /etc/clickhouse-server/config.xml -->
<clickhouse>
    <logger>
        <level>information</level>
        <log>/var/log/clickhouse-server/clickhouse-server.log</log>
        <errorlog>/var/log/clickhouse-server/clickhouse-server.err.log</errorlog>
        <size>1000M</size>
        <count>10</count>
    </logger>
    
    <http_port>8123</http_port>
    <tcp_port>9000</tcp_port>
    <interserver_http_port>9009</interserver_http_port>
    
    <max_connections>4096</max_connections>
    <keep_alive_timeout>3</keep_alive_timeout>
    <max_concurrent_queries>100</max_concurrent_queries>
    
    <uncompressed_cache_size>8589934592</uncompressed_cache_size>
    <mark_cache_size>5368709120</mark_cache_size>
    
    <timezone>UTC</timezone>
</clickhouse>
```

### 5.2 Configuração de Usuários

```xml
<!-- /etc/clickhouse-server/users.xml -->
<clickhouse>
    <users>
        <default>
            <password></password>
            <networks incl="networks" replace="replace">
                <ip>::/0</ip>
            </networks>
            <profile>default</profile>
            <quota>default</quota>
        </default>
        
        <readonly>
            <password_sha256_hex>hash_da_senha</password_sha256_hex>
            <networks>
                <ip>::1</ip>
                <ip>127.0.0.1</ip>
            </networks>
            <profile>readonly</profile>
            <quota>default</quota>
        </readonly>
    </users>
    
    <profiles>
        <default>
            <max_memory_usage>10000000000</max_memory_usage>
            <use_uncompressed_cache>1</use_uncompressed_cache>
            <load_balancing>random</load_balancing>
        </default>
        
        <readonly>
            <readonly>1</readonly>
            <max_memory_usage>5000000000</max_memory_usage>
        </readonly>
    </profiles>
</clickhouse>
```

## 6. Segurança e Governança

### 6.1 Controles de Segurança

#### 6.1.1 Autenticação e Autorização
- Integração com IAM INNOVABIZ
- Autenticação baseada em certificados TLS
- RBAC granular por tenant/região
- Auditoria completa de acessos

#### 6.1.2 Criptografia
- TLS 1.3 para todas as conexões
- Criptografia de dados em repouso
- Rotação automática de chaves
- HSM para chaves críticas

#### 6.1.3 Isolamento Multi-Tenant
- Segregação lógica por labels
- Políticas de rede restritivas
- Quotas por tenant
- Monitoramento de vazamentos de dados

### 6.2 Compliance e Auditoria

#### 6.2.1 Requisitos Regulatórios
- **PCI DSS 4.0**: Controles 2.2, 8.1-8.3, 10.1-10.3
- **GDPR/LGPD**: Artigos 25, 32, 35
- **ISO 27001**: Controles A.12.1, A.12.6, A.14.1
- **NIST CSF**: PR.AC, PR.DS, DE.AE, RS.RP

#### 6.2.2 Logs de Auditoria
```sql
-- Configurar auditoria de consultas
SET log_queries = 1;
SET log_query_threads = 1;
SET log_profile_events = 1;

-- Verificar logs de auditoria
SELECT * FROM system.query_log 
WHERE user != 'default' 
AND event_date >= today() - 1;
```

## 7. Ciclo de Vida e Manutenção

### 7.1 Atualizações e Patches

#### 7.1.1 Processo de Atualização
1. **Planejamento**: Análise de impacto e janela de manutenção
2. **Backup**: Backup completo pré-atualização
3. **Teste**: Validação em ambiente de homologação
4. **Execução**: Atualização rolling em produção
5. **Validação**: Testes de funcionalidade e performance

#### 7.1.2 Cronograma de Manutenção
- **Patches de Segurança**: Imediato (< 24h)
- **Updates Menores**: Mensal
- **Updates Maiores**: Trimestral
- **Manutenção Preventiva**: Semanal

### 7.2 Capacity Planning

#### 7.2.1 Métricas de Crescimento
- Taxa de crescimento de dados: 15% ao mês
- Aumento de consultas: 10% ao mês
- Novos tenants: 5% ao mês
- Expansão geográfica: Trimestral

#### 7.2.2 Thresholds de Expansão
- CPU > 70% sustentado por 7 dias
- Memória > 80% sustentado por 7 dias
- Disco > 75% de utilização
- Latência P95 > 2s por 3 dias consecutivos

## 8. Contatos e Escalação

### 8.1 Matriz de Escalação

| Nível | Responsável | Tempo de Resposta | Contato |
|-------|-------------|-------------------|---------|
| L1 | NOC INNOVABIZ | 5 minutos | noc@innovabiz.com |
| L2 | SRE Team | 15 minutos | sre@innovabiz.com |
| L3 | Platform Engineering | 30 minutos | platform@innovabiz.com |
| L4 | Arquiteto de Dados | 1 hora | data-architect@innovabiz.com |

### 8.2 Canais de Comunicação

- **Slack**: #innovabiz-alerts, #clickhouse-ops
- **PagerDuty**: Integração automática para alertas críticos
- **Jira**: Tickets para problemas não urgentes
- **Confluence**: Documentação e post-mortems

### 8.3 Fornecedores e Suporte

- **ClickHouse Inc.**: Suporte enterprise 24/7
- **Cloud Provider**: Suporte de infraestrutura
- **Monitoring Vendor**: Suporte de observabilidade

## 9. Referências e Recursos

### 9.1 Documentação Técnica
- [ClickHouse Official Documentation](https://clickhouse.com/docs/)
- [INNOVABIZ Platform Architecture](https://wiki.innovabiz.com/architecture)
- [Multi-Context Monitoring Standards](https://wiki.innovabiz.com/monitoring)

### 9.2 Ferramentas e Scripts
- [ClickHouse Backup Tool](https://github.com/AlexAkulov/clickhouse-backup)
- [INNOVABIZ Automation Scripts](https://git.innovabiz.com/platform/automation)
- [Monitoring Templates](https://git.innovabiz.com/platform/monitoring)

### 9.3 Treinamentos e Certificações
- ClickHouse Certified Developer
- INNOVABIZ Platform Operations
- SRE Best Practices

---

**Documento Controlado**  
**Classificação**: Confidencial - Uso Interno  
**Autor**: Equipe SRE INNOVABIZ  
**Aprovado por**: Arquiteto de Plataforma  
**Próxima Revisão**: Maio 2025