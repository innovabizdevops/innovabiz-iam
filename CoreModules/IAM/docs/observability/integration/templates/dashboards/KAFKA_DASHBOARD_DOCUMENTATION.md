# Documentação Técnica: Dashboard Kafka INNOVABIZ

![Status](https://img.shields.io/badge/Status-Implementado-brightgreen)
![Versão](https://img.shields.io/badge/Versão-1.0.0-blue)
![Data](https://img.shields.io/badge/Data-2025--02-lightgrey)
![Plataforma](https://img.shields.io/badge/Plataforma-INNOVABIZ-purple)

## 1. Visão Geral

O Dashboard Kafka INNOVABIZ fornece monitoramento completo em tempo real para clusters Kafka, essenciais para a infraestrutura de mensageria e integração da plataforma INNOVABIZ. Este dashboard oferece observabilidade abrangente para operações de mensageria, incluindo tráfego de dados, saúde dos brokers, consumer groups, métricas de desempenho e recursos do sistema, com suporte completo ao modelo multi-contexto da plataforma.

## 2. Objetivos

- **Monitoramento em Tempo Real**: Acompanhar o estado operacional dos clusters Kafka em tempo real
- **Visibilidade Multi-contexto**: Permitir visualização por tenant, região, ambiente e instância
- **Detecção Proativa de Problemas**: Identificar gargalos de performance, issues de consumer lag e saturação de recursos
- **Análise de Tendências**: Visualizar padrões de tráfego e utilização ao longo do tempo
- **SLA Monitoring**: Facilitar o acompanhamento de métricas essenciais para cumprimento de SLAs
- **Troubleshooting**: Possibilitar diagnóstico rápido de problemas relacionados à mensageria
- **Capacity Planning**: Fornecer insights para planejamento de capacidade e escalabilidade

## 3. Estrutura do Dashboard

### 3.1 Seções e Painéis

O dashboard está estruturado em cinco seções principais:

#### 3.1.1 Visão Geral do Cluster Kafka
- **Broker Status**: Estado operacional (online/offline) de cada broker
- **Total de Brokers**: Número total de brokers ativos no cluster
- **Total de Tópicos**: Contagem total de tópicos configurados
- **Total de Partições**: Número global de partições no cluster

#### 3.1.2 Métricas de Tráfego
- **Bytes Recebidos**: Taxa de dados recebidos por broker (bytes/segundo)
- **Bytes Enviados**: Taxa de dados enviados por broker (bytes/segundo)
- **Mensagens por Segundo**: Taxa de mensagens processadas por broker
- **Requisições por Segundo**: Volume de requisições processadas por broker

#### 3.1.3 Performance e Latência
- **Tempo de Resposta por Tipo de Requisição**: Latência média para diferentes tipos de requisições (Produce, FetchConsumer, FetchFollower)
- **Tempo de Espera na Fila**: Percentis (P50, P95, P99) do tempo que as requisições aguardam na fila

#### 3.1.4 Consumer Groups e Offset
- **Total de Consumer Groups**: Número de consumer groups ativos
- **Lag Total dos Consumers**: Acúmulo total de mensagens não processadas
- **Top 10 Consumer Groups por Lag**: Ranking dos consumer groups com maior lag
- **Evolução de Lag por Consumer Group**: Gráfico temporal mostrando tendências de lag

#### 3.1.5 JVM e Recursos do Sistema
- **Uso de Memória JVM**: Utilização de memória heap e non-heap
- **Tempo de GC**: Tempo gasto em garbage collection
- **Uso de CPU**: Percentual de utilização de CPU por broker
- **Tamanho dos Logs por Tópico**: Utilização de disco por logs de tópicos

### 3.2 Anotações e Alertas Visuais

O dashboard inclui:
- **Anotações de Eventos Críticos**: Eleições de líder são automaticamente anotadas
- **Thresholds Visuais**: Limiares coloridos para indicar estados de alerta em métricas críticas
- **Refresh Automático**: Atualização a cada 30 segundos para monitoramento em tempo real

## 4. Variáveis e Filtragem Multi-Contexto

O dashboard implementa o modelo de filtragem multi-contexto INNOVABIZ com as seguintes variáveis:

| Variável | Descrição | Tipo | Dependência |
|----------|-----------|------|------------|
| `tenant_id` | Identificador do tenant | Multi-select | Nenhuma |
| `region_id` | Região geográfica | Multi-select | `tenant_id` |
| `environment` | Ambiente (prod, staging, dev, etc.) | Multi-select | `tenant_id`, `region_id` |
| `instance` | Instância Kafka específica | Multi-select | `tenant_id`, `region_id`, `environment` |
| `topic` | Tópico específico | Multi-select | `tenant_id`, `region_id`, `environment`, `instance` |

Esta estrutura de variáveis em cascata permite filtrar métricas de forma granular, respeitando o modelo multi-tenant e multi-região da plataforma INNOVABIZ, garantindo isolamento adequado e segurança contextual.

## 5. Métricas Prometheus Necessárias

### 5.1 Exporters Requeridos

O dashboard requer a coleta das seguintes métricas via JMX Exporter:

#### 5.1.1 Métricas Principais de Broker
- `kafka_server_brokertopicmetrics_up`
- `kafka_server_brokertopicmetrics_bytesinpersec_count`
- `kafka_server_brokertopicmetrics_bytesoutpersec_count`
- `kafka_server_brokertopicmetrics_messagesinpersec_count`
- `kafka_controller_kafkacontroller_globaltopiccount`
- `kafka_controller_kafkacontroller_globalpartitioncount`
- `kafka_server_replicamanager_leadercount`

#### 5.1.2 Métricas de Requisição e Performance
- `kafka_network_requestmetrics_requests_total`
- `kafka_network_requestmetrics_totaltimems_sum`
- `kafka_network_requestmetrics_totaltimems_count`
- `kafka_network_requestmetrics_requestqueuetimems`

#### 5.1.3 Métricas de Consumer Group
- `kafka_consumergroup_group_count`
- `kafka_consumergroup_group_lag`

#### 5.1.4 Métricas JVM e Sistema
- `jvm_memory_bytes_used`
- `jvm_gc_collection_seconds_sum`
- `node_cpu_seconds_total`
- `kafka_log_log_size`

### 5.2 Configuração de Labels Obrigatórios

É **imperativo** que todas as métricas incluam os seguintes labels para compatibilidade com o modelo multi-contexto:

```yaml
tenant_id: "<identificador do tenant>"
region_id: "<identificador da região>"
environment: "<ambiente de execução>"
```

### 5.3 Retenção e Resolução

- **Resolução recomendada**: 15s para métricas críticas, 30s para métricas gerais
- **Período de retenção**: 15 dias para dados brutos, 90 dias para dados agregados
- **Agregação**: Recomenda-se configurar regras de recording para pré-agregação de métricas de alta cardinalidade

## 6. Implementação e Configuração

### 6.1 Pré-requisitos

- Grafana versão 9.0+
- Prometheus versão 2.30+
- JMX Exporter versão 0.17.0+ configurado para Kafka
- Labels multi-contexto implementados em todas as métricas coletadas

### 6.2 Procedimento de Instalação

1. **Importe o Dashboard**:
   - Acesse Grafana > Dashboards > Import
   - Faça upload do arquivo JSON `KAFKA_DASHBOARD.json`
   - Selecione a fonte de dados Prometheus apropriada

2. **Configure Variáveis**:
   - Verifique se as consultas de variáveis estão retornando os valores esperados
   - Ajuste os filtros de variáveis conforme necessário para sua infraestrutura

3. **Ajuste Thresholds**:
   - Personalize os limiares de alerta visual de acordo com os requisitos específicos do ambiente

4. **Configuração de Anotações**:
   - Verifique se a consulta de anotação para eleições de líder está funcionando corretamente
   - Adicione outras anotações conforme necessário (ex: reinícios planejados, manutenções)

### 6.3 Integração com o Kafka Exporter

Configure o JMX Exporter com o seguinte formato básico:

```yaml
lowercaseOutputName: true
lowercaseOutputLabelNames: true
whitelistObjectNames:
  - kafka.server:*
  - kafka.controller:*
  - kafka.network:*
  - java.lang:type=GarbageCollector,name=*
  - java.lang:type=Memory
rules:
  # Adicione regra para incluir labels de tenant, região e ambiente
  - pattern: ".*"
    labels:
      tenant_id: "${tenant_id}"
      region_id: "${region_id}"
      environment: "${environment}"
```

## 7. Casos de Uso Operacionais

### 7.1 Monitoramento Diário

- **Verificação de Saúde**: Revisão da disponibilidade e status dos brokers
- **Análise de Tráfego**: Monitoramento dos padrões de entrada/saída de dados
- **Consumer Lag**: Verificação de atrasos no processamento de mensagens
- **Utilização de Recursos**: Acompanhamento da utilização de CPU, memória e disco

### 7.2 Troubleshooting

- **Detecção de Gargalos**: Identificação de bottlenecks em brokers específicos
- **Análise de Latência**: Investigação de aumentos no tempo de resposta
- **Diagnóstico de Consumer Issues**: Análise de consumidores lentos ou com problemas
- **Rebalanceamento**: Identificação da necessidade de redistribuição de partições

### 7.3 Capacity Planning

- **Tendências de Crescimento**: Análise de padrões de crescimento de tópicos
- **Previsão de Escala**: Identificação antecipada de necessidades de escalabilidade
- **Distribuição de Carga**: Avaliação da distribuição de carga entre brokers
- **Planejamento de Recursos**: Estimativa de requisitos futuros de hardware

## 8. Governança, Compliance e Segurança

### 8.1 Governança de Dados

O dashboard está alinhado com as políticas de governança de dados INNOVABIZ:
- **Segregação Multi-tenant**: Isolamento completo de dados entre tenants
- **Visibilidade Contextual**: Acesso a métricas baseado em contexto e permissões
- **Auditabilidade**: Rastreamento de uso e acesso às métricas operacionais

### 8.2 Compliance

Compatível com requisitos regulatórios e frameworks de compliance:
- **PCI DSS 4.0**: Monitoramento de sistemas críticos (Requisitos 10.2, 10.4.1)
- **ISO 27001**: Controles de monitoramento de segurança da informação (A.12.4)
- **GDPR/LGPD**: Suporte a proteção de dados através de isolamento multi-tenant
- **NIST CSF**: Alinhado com funções de Detecção e Resposta do framework

### 8.3 Segurança

- **Controle de Acesso**: Integração com IAM INNOVABIZ para RBAC granular
- **Exposição Limitada**: Sem exposição de dados sensíveis nas métricas
- **Autenticação**: Acesso restrito via SSO corporativo
- **Tenant Isolation**: Estrita separação de métricas entre tenants

## 9. Integração com Alertas

### 9.1 Regras de Alerta Recomendadas

Recomenda-se configurar as seguintes regras de alerta no Prometheus AlertManager:

```yaml
groups:
- name: innovabiz_kafka_alerts
  rules:
  - alert: KafkaBrokerDown
    expr: kafka_server_brokertopicmetrics_up == 0
    for: 2m
    labels:
      severity: critical
      service: kafka
    annotations:
      summary: "Kafka Broker Down"
      description: "Broker {{ $labels.instance }} está offline há pelo menos 2 minutos"
      
  - alert: KafkaHighConsumerLag
    expr: kafka_consumergroup_group_lag > 10000
    for: 5m
    labels:
      severity: warning
      service: kafka
    annotations:
      summary: "Consumer Lag Alto"
      description: "Consumer group {{ $labels.group }} apresenta lag de {{ $value }} mensagens por mais de 5 minutos"
      
  - alert: KafkaHighRequestLatency
    expr: sum by(request) (rate(kafka_network_requestmetrics_totaltimems_sum[5m])) / sum by(request) (rate(kafka_network_requestmetrics_totaltimems_count[5m])) > 100
    for: 3m
    labels:
      severity: warning
      service: kafka
    annotations:
      summary: "Latência Alta em Requisições"
      description: "Tipo de requisição {{ $labels.request }} apresenta latência média de {{ $value }}ms por mais de 3 minutos"
```

### 9.2 Integração com Sistema de Notificação INNOVABIZ

Os alertas devem ser integrados com:
- **INNOVABIZ AlertHub**: Central de gerenciamento de alertas da plataforma
- **PagerDuty/OpsGenie**: Para escalação de incidentes críticos
- **Webhooks para MS Teams/Slack**: Para notificações em canais de equipe
- **Email**: Para resumos diários e alertas de menor prioridade

## 10. Integração com o Ecossistema INNOVABIZ

### 10.1 Correlação com Outros Dashboards

Este dashboard complementa e se integra com:
- **Dashboard Kubernetes**: Para correlacionar issues de infraestrutura
- **Dashboard de Microserviços**: Para visualizar impactos de mensageria em serviços
- **Dashboard de APIs**: Para analisar efeitos de throttling e backpressure
- **Dashboard End-to-End**: Para rastreamento completo de fluxos de mensagens

### 10.2 INNOVABIZ IAM Integration

O acesso ao dashboard é controlado via:
- **RBAC**: Perfis específicos para operações, desenvolvimento e suporte
- **Permissões Contextuais**: Acesso limitado a tenants/regiões específicas
- **Auditoria de Acesso**: Registro de todas as visualizações e ações no dashboard

## 11. Referências e Recursos Adicionais

- [INNOVABIZ Kafka Operations Manual](/CoreModules/IAM/docs/operations/messaging/KAFKA_OPERATIONS.md)
- [Prometheus JMX Exporter Documentation](https://github.com/prometheus/jmx_exporter)
- [Kafka Monitoring Best Practices](/CoreModules/IAM/docs/observability/best-practices/KAFKA_MONITORING.md)
- [INNOVABIZ Alerting Strategy](/CoreModules/IAM/docs/observability/alerting/ALERTING_STRATEGY.md)
- [Multi-Context Monitoring Framework](/CoreModules/IAM/docs/observability/architecture/MULTI_CONTEXT_MONITORING.md)

---

**Autor**: Equipe de Observabilidade INNOVABIZ  
**Última Atualização**: 2025-02  
**Status**: Implementado  
**Classificação**: Público

---

*Este documento faz parte da documentação oficial da plataforma INNOVABIZ e está sujeito às políticas de controle de versão e revisão documentadas em [Governança de Documentação](/CoreModules/IAM/docs/governance/DOCUMENTATION_GOVERNANCE.md)*