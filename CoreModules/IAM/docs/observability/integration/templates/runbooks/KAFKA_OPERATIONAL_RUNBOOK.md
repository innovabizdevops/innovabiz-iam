# Runbook Operacional: Apache Kafka INNOVABIZ

![Status](https://img.shields.io/badge/Status-Produção-brightgreen)
![Versão](https://img.shields.io/badge/Versão-1.0.0-blue)
![Criticidade](https://img.shields.io/badge/Criticidade-Alta-red)
![Plataforma](https://img.shields.io/badge/Plataforma-INNOVABIZ-purple)

## 1. Visão Geral do Serviço

Apache Kafka é um componente crítico da infraestrutura de mensageria e event streaming da plataforma INNOVABIZ, sendo utilizado para integração entre módulos, processamento de eventos em tempo real, pipelines de dados e comunicação assíncrona entre serviços.

### 1.1 Importância e Impacto

- **Criticidade**: ALTA
- **RTO (Recovery Time Objective)**: 15 minutos
- **RPO (Recovery Point Objective)**: 0 (zero perda de dados)
- **Impactos em Caso de Falha**: 
  - Interrupção de processamento de transações financeiras
  - Falhas em integrações entre módulos da plataforma
  - Perda de dados em filas de processamento
  - Inconsistências em sistemas downstream

### 1.2 Arquitetura e Topologia

- **Versão**: Apache Kafka 3.4.0
- **Brokers**: Mínimo 3 brokers por cluster (n+1)
- **Replicação**: Fator de replicação 3 para todos os tópicos críticos
- **Clusters**: Segregados por ambiente (prod, stage, dev)
- **Zookeeper**: Cluster dedicado com 3 ou 5 nodes (a ser migrado para KRaft em breve)
- **Segurança**: TLS 1.3 para comunicação, SASL/SCRAM para autenticação

### 1.3 Dependências

- **Upstream**: 
  - ZooKeeper/KRaft para coordenação de cluster
  - Sistema de autenticação IAM INNOVABIZ
  - Network Layer (VPCs, Security Groups, NACLs)

- **Downstream**: 
  - Serviços de streaming e processamento (Kafka Streams, Kafka Connect)
  - Sistemas de persistência (PostgreSQL, ClickHouse)
  - Módulos de negócio da plataforma INNOVABIZ

## 2. Monitoramento e Alertas

### 2.1 Dashboard Principal

[Dashboard Kafka INNOVABIZ](https://grafana.innovabiz.com/d/innovabiz-kafka)

### 2.2 Métricas Críticas

| Métrica | Descrição | Threshold de Aviso | Threshold Crítico |
|---------|-----------|-------------------|------------------|
| `kafka_server_brokertopicmetrics_up` | Status dos brokers | N/A | 0 (offline) |
| `kafka_consumergroup_group_lag` | Lag de consumer groups | 10.000 msgs | 100.000 msgs |
| `kafka_server_replicamanager_underreplicatedpartitions` | Partições sub-replicadas | > 0 | > 10 |
| `jvm_memory_bytes_used{area="heap"}` | Uso de memória heap | 85% | 95% |
| `node_filesystem_avail_bytes` | Espaço em disco disponível | < 25% | < 10% |
| `kafka_network_requestmetrics_totaltimems` | Latência de requisições | > 100ms | > 500ms |

### 2.3 Alertas Configurados

Os alertas estão configurados em Prometheus AlertManager conforme o arquivo `KAFKA_ALERT_RULES.yml`. Principais alertas:

- KafkaBrokerDown
- KafkaHighConsumerLag
- KafkaCriticalConsumerLag
- KafkaUnderReplicatedPartitions
- KafkaHighCpuUsage
- KafkaHighMemoryUsage
- KafkaDiskSpaceWarning/Critical
- KafkaTopicWithoutRetention

### 2.4 Logs e Traces

- **Logs**: Centralizados no Loki com retenção de 30 dias
  - Path: `/var/log/kafka/server.log`
  - Consulta Loki: `{job="kafka", tenant_id="$TENANT_ID", environment="$ENV"}`

- **Métricas**: Coletadas via JMX Exporter para Prometheus
  - Port: 9999 (/metrics)
  - Intervalo de scrape: 15s
  - Retenção: 15 dias (dados brutos), 90 dias (dados agregados)

## 3. Procedimentos Operacionais

### 3.1 Verificações de Saúde

#### Verificação Básica de Status

```bash
# Verificar status de brokers (deve retornar todos os brokers configurados)
kafka-broker-api-versions.sh --bootstrap-server kafka-broker-1:9092,kafka-broker-2:9092,kafka-broker-3:9092

# Listar tópicos para verificar conectividade
kafka-topics.sh --bootstrap-server kafka-broker-1:9092 --list

# Verificar grupos de consumidores
kafka-consumer-groups.sh --bootstrap-server kafka-broker-1:9092 --list
```

#### Verificação de Replicação

```bash
# Verificar partições sub-replicadas
kafka-topics.sh --bootstrap-server kafka-broker-1:9092 --describe --under-replicated-partitions

# Verificar tópicos com fator de replicação insuficiente
kafka-topics.sh --bootstrap-server kafka-broker-1:9092 --describe | grep "replication-factor: [12]"
```

### 3.2 Troubleshooting de Problemas Comuns

#### Alerta: KafkaBrokerDown

1. **Verificar conectividade de rede**:
   ```bash
   ping <broker-hostname>
   telnet <broker-hostname> 9092
   ```

2. **Verificar logs do broker**:
   ```bash
   sudo tail -f /var/log/kafka/server.log
   ```

3. **Verificar uso de recursos**:
   ```bash
   top -u kafka
   df -h /data/kafka
   ```

4. **Reiniciar o serviço (se necessário)**:
   ```bash
   sudo systemctl restart kafka
   ```

5. **Verificar status após reinício**:
   ```bash
   sudo systemctl status kafka
   kafka-broker-api-versions.sh --bootstrap-server <broker-hostname>:9092
   ```

#### Alerta: KafkaHighConsumerLag / KafkaCriticalConsumerLag

1. **Identificar o consumer group e tópico afetados**:
   ```bash
   kafka-consumer-groups.sh --bootstrap-server <broker>:9092 --describe --group <group-name>
   ```

2. **Verificar métricas de consumo**:
   - Acessar o dashboard Kafka e filtrar por tópico e consumer group afetados
   - Verificar taxa de consumo vs. taxa de produção

3. **Verificar logs do consumidor**:
   - Localizar os serviços consumidores com base no consumer group
   - Verificar logs de erros ou lentidão

4. **Ações possíveis**:
   - Escalar horizontalmente o serviço consumidor
   - Verificar gargalos no processamento downstream
   - Ajustar configurações de consumo (fetch.min.bytes, max.poll.records)
   - Implementar back pressure se necessário

#### Alerta: KafkaUnderReplicatedPartitions

1. **Identificar as partições afetadas**:
   ```bash
   kafka-topics.sh --bootstrap-server <broker>:9092 --describe --under-replicated-partitions
   ```

2. **Verificar brokers com problemas**:
   - Dashboard Kafka > Seção "Visão Geral do Cluster"
   - Verificar métricas de disco, rede e CPU dos brokers

3. **Verificar se há manutenção em andamento**:
   - Consultar calendário de manutenção
   - Verificar se há operações de rebalanceamento em andamento

4. **Ações possíveis**:
   - Resolver problemas de recursos nos brokers afetados
   - Reiniciar brokers não-responsivos (um de cada vez)
   - Executar reassignment de partições se necessário

### 3.3 Operações de Rotina

#### Adicionar um Novo Broker

1. **Preparação**:
   - Provisionar servidor com recursos adequados
   - Instalar Java e dependências necessárias
   - Configurar diretórios de dados e logs

2. **Configuração**:
   ```properties
   # Adicionar ao server.properties
   broker.id=<novo_id_único>
   listeners=PLAINTEXT://<hostname>:9092,SASL_SSL://<hostname>:9093
   log.dirs=/data/kafka/logs
   zookeeper.connect=<zk1>:2181,<zk2>:2181,<zk3>:2181
   ```

3. **Segurança**:
   - Gerar/obter certificados TLS
   - Configurar SASL/SCRAM
   - Aplicar ACLs necessárias

4. **Iniciar o Broker**:
   ```bash
   sudo systemctl start kafka
   ```

5. **Verificar Inclusão no Cluster**:
   ```bash
   kafka-broker-api-versions.sh --bootstrap-server kafka-broker-1:9092
   ```

6. **Rebalancear Partições** (se necessário):
   - Gerar plano de reassignment
   - Executar reassignment gradualmente

#### Rebalanceamento de Partições

1. **Criar arquivo de tópicos para rebalanceamento**:
   ```json
   {"topics": [{"topic": "topic1"}, {"topic": "topic2"}], "version": 1}
   ```

2. **Gerar plano de rebalanceamento**:
   ```bash
   kafka-reassign-partitions.sh --bootstrap-server <broker>:9092 \
     --topics-to-move-json-file topics.json \
     --broker-list "1,2,3,4" \
     --generate
   ```

3. **Executar rebalanceamento**:
   ```bash
   kafka-reassign-partitions.sh --bootstrap-server <broker>:9092 \
     --reassignment-json-file reassignment.json \
     --execute
   ```

4. **Monitorar progresso**:
   ```bash
   kafka-reassign-partitions.sh --bootstrap-server <broker>:9092 \
     --reassignment-json-file reassignment.json \
     --verify
   ```

### 3.4 Recuperação de Desastres

#### Cenário 1: Falha de um Broker

1. **Avaliação**: Verificar se o cluster continua operacional
2. **Validação**: Confirmar ausência de partições offline (apenas sub-replicadas)
3. **Recuperação**: 
   - Tentar reiniciar o broker com problemas
   - Se persistir, substituir o servidor mantendo o mesmo broker.id
4. **Reintegração**: O broker se reintegrará automaticamente e sincronizará as partições

#### Cenário 2: Falha de Múltiplos Brokers

1. **Avaliação**: Determinar se o quórum foi perdido
2. **Contenção**:
   - Notificar stakeholders sobre indisponibilidade parcial
   - Implementar circuit breakers em produtores críticos
3. **Recuperação**:
   - Priorizar a recuperação do broker líder para tópicos críticos
   - Restaurar brokers na ordem do menos para o mais afetado
4. **Validação**:
   - Verificar consumer lag após recuperação
   - Monitorar taxa de processamento de mensagens acumuladas

#### Cenário 3: Perda de Datacenter

1. **Ativação de DR**: Ativar cluster secundário em região alternativa
2. **Redirecionamento**:
   - Atualizar configuração de clients para apontamento ao cluster DR
   - Implementar procedimentos de failover nos serviços dependentes
3. **Reconciliação**:
   - Identificar possíveis mensagens perdidas (se RPO > 0)
   - Implementar procedimentos de recuperação de dados se necessário
4. **Restabelecimento**: Reconstruir cluster primário e sincronizar quando possível

## 4. Configurações e Otimizações

### 4.1 Configurações Recomendadas por Ambiente

#### Produção

```properties
# Desempenho e Confiabilidade
num.network.threads=8
num.io.threads=16
socket.send.buffer.bytes=102400
socket.receive.buffer.bytes=102400
socket.request.max.bytes=104857600
num.replica.fetchers=4

# Retenção e Durabilidade
log.retention.hours=168
log.segment.bytes=1073741824
log.retention.check.interval.ms=300000
min.insync.replicas=2

# Otimização de Recursos
num.partitions=12
default.replication.factor=3
```

#### Staging/QA

```properties
# Configuração mais leve para ambientes não-produtivos
num.network.threads=4
num.io.threads=8
num.partitions=6
default.replication.factor=3
min.insync.replicas=1
```

### 4.2 Melhores Práticas de Configuração de Tópicos

```bash
# Tópicos de Alta Throughput
kafka-topics.sh --bootstrap-server <broker>:9092 --create \
  --topic high-throughput-topic \
  --partitions 24 \
  --replication-factor 3 \
  --config min.insync.replicas=2 \
  --config retention.ms=604800000 \
  --config segment.bytes=1073741824 \
  --config cleanup.policy=delete

# Tópicos de Processamento Crítico
kafka-topics.sh --bootstrap-server <broker>:9092 --create \
  --topic critical-processing-topic \
  --partitions 12 \
  --replication-factor 3 \
  --config min.insync.replicas=2 \
  --config retention.ms=259200000 \
  --config segment.bytes=536870912 \
  --config cleanup.policy=delete \
  --config unclean.leader.election.enable=false
```

### 4.3 Recomendações para Clientes

#### Produtores

```properties
# Confiabilidade
acks=all
retries=Integer.MAX_VALUE
max.in.flight.requests.per.connection=5
enable.idempotence=true

# Desempenho
batch.size=16384
linger.ms=5
compression.type=snappy
buffer.memory=33554432

# Multi-contexto INNOVABIZ
interceptor.classes=com.innovabiz.kafka.interceptors.MultiContextProducerInterceptor
```

#### Consumidores

```properties
# Confiabilidade
enable.auto.commit=false
isolation.level=read_committed
auto.offset.reset=earliest

# Desempenho
fetch.min.bytes=1024
fetch.max.wait.ms=500
max.partition.fetch.bytes=1048576
max.poll.records=500

# Multi-contexto INNOVABIZ
interceptor.classes=com.innovabiz.kafka.interceptors.MultiContextConsumerInterceptor
```

## 5. Governança e Segurança

### 5.1 Conformidade e Padrões

O serviço Kafka é operado em conformidade com:

- **PCI DSS 4.0**: Requisitos 2.2, 3.4, 4.1, 6.2, 10.2
- **ISO 27001**: Controles A.12.1.2, A.12.4, A.13.1
- **GDPR/LGPD**: Requisitos de proteção de dados e isolamento multi-tenant
- **NIST CSF**: Funções Identify, Protect, Detect

### 5.2 Políticas de Acesso

- **Autenticação**: SASL/SCRAM integrado com IAM INNOVABIZ
- **Autorização**: ACLs baseadas em role com princípio de menor privilégio
- **Auditoria**: Logs de acesso preservados por 90 dias
- **Rotação de Credenciais**: Trimestral para contas de serviço

### 5.3 Isolamento Multi-Contexto

- **Isolamento por Tenant**: Separação por prefixo de tópico e ACLs
- **Rastreabilidade**: Headers de mensagens incluem tenant_id, region_id, environment
- **Interceptors**: Implementação obrigatória de interceptors para enriquecimento de contexto
- **Monitoramento**: Métricas segregadas por tenant, região e ambiente

## 6. Ciclo de Vida e Evolução

### 6.1 Versionamento e Upgrades

- **Política de Upgrades**: Trimestral para patches de segurança, semestral para versões menores
- **Janela de Manutenção**: Domingos, 00:00-04:00 (horário local)
- **Processo de Rollout**: Rolling upgrade com testes de regressão em cada broker
- **Testabilidade**: Ambientes de QA espelham configurações de produção

### 6.2 Capacidade e Escalabilidade

- **Limites Atuais**:
  - Taxa máxima de mensagens: 50.000 msgs/s por broker
  - Tamanho máximo de mensagem: 10 MB
  - Número máximo de tópicos: 10.000
  - Número máximo de partições por broker: 4.000

- **Planejamento de Capacidade**:
  - Monitoramento de tendências de crescimento mensal
  - Threshold de alerta em 70% da capacidade
  - Expansão horizontal quando utilização consistente > 60%

### 6.3 Roadmap Técnico

- **Curto Prazo (3-6 meses)**:
  - Migração para KRaft (eliminação de dependência do Zookeeper)
  - Implementação de Tiered Storage para redução de custos
  - Melhoria nos interceptors para rastreabilidade end-to-end

- **Médio Prazo (6-12 meses)**:
  - Expansão para modelo multi-região ativo-ativo
  - Integração com Apache Flink para processamento avançado
  - Implementação de criptografia transparente para dados em repouso

## 7. Contatos e Escalação

### 7.1 Equipe Responsável

| Papel | Responsabilidade | Contato | Horário |
|------|-----------------|---------|---------|
| Engenheiro de Plantão (SRE) | Resposta inicial | sre-oncall@innovabiz.com | 24x7 |
| Especialista Kafka | Suporte nível 2 | kafka-support@innovabiz.com | 8x5 |
| Arquiteto de Plataforma | Decisões críticas | platform-architect@innovabiz.com | 8x5 |

### 7.2 Matriz de Escalação

| Severidade | Tempo Máximo de Resposta | Nível 1 | Nível 2 | Nível 3 |
|------------|--------------------------|---------|---------|---------|
| SEV-1 (Crítico) | 15 minutos | SRE de Plantão | Especialista Kafka | CTO |
| SEV-2 (Alto) | 30 minutos | SRE de Plantão | Especialista Kafka | Gerente de Plataforma |
| SEV-3 (Médio) | 2 horas | SRE de Plantão | Especialista Kafka | - |
| SEV-4 (Baixo) | 8 horas | Ticket de suporte | - | - |

## 8. Recursos Adicionais

- [Wiki Interna de Kafka](https://wiki.innovabiz.com/platform/kafka)
- [Dashboard Kafka INNOVABIZ](https://grafana.innovabiz.com/d/innovabiz-kafka)
- [Repositório de Configuração](https://github.com/innovabiz/kafka-config)
- [Política de Backup e DR](https://wiki.innovabiz.com/platform/dr-policy)
- [Apache Kafka Documentation](https://kafka.apache.org/documentation/)
- [Confluent Kafka Operations](https://docs.confluent.io/platform/current/kafka/operations.html)

---

**Autor**: Equipe de Plataforma INNOVABIZ  
**Última Atualização**: 2025-02  
**Revisão Programada**: 2025-08  
**Classificação**: Confidencial - Uso Interno

---

*Este documento faz parte da documentação oficial da plataforma INNOVABIZ e está sujeito às políticas de controle de versão e revisão documentadas em [Governança de Documentação](/CoreModules/IAM/docs/governance/DOCUMENTATION_GOVERNANCE.md)*