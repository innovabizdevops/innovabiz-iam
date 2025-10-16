# Integração do Apache Kafka com o Módulo IAM INNOVABIZ

## Visão Geral

Este documento descreve a arquitetura de integração do Apache Kafka com o módulo de Gerenciamento de Identidade e Acesso (IAM) da plataforma INNOVABIZ. O Apache Kafka é utilizado como a espinha dorsal da comunicação assíncrona entre os diversos componentes do módulo IAM, permitindo escalabilidade, resiliência e observabilidade em tempo real.

## Arquitetura de Eventos

O módulo IAM do INNOVABIZ adota uma arquitetura orientada a eventos (Event-Driven Architecture - EDA) para:

1. **Desacoplamento de Serviços**: Permitir que os microserviços operem de forma independente
2. **Escalabilidade Horizontal**: Facilitar o escalonamento de componentes individuais
3. **Resiliência**: Garantir que falhas em um componente não afetem todo o sistema
4. **Auditoria Completa**: Manter um registro imutável de todas as operações de autenticação e autorização
5. **Adaptação Regional**: Atender aos requisitos específicos de cada região de implementação

### Componentes da Infraestrutura Kafka

A infraestrutura Kafka do módulo IAM inclui:

- **Brokers Kafka**: Responsáveis pelo armazenamento e distribuição de eventos
- **Zookeeper**: Gerencia a coordenação entre os brokers
- **Schema Registry**: Garante a consistência dos esquemas de eventos
- **Kafka Connect**: Facilita a integração com sistemas externos
- **KSQLDB**: Permite o processamento de fluxos de eventos em tempo real
- **Kafka UI**: Interface de gerenciamento e monitoramento

## Tópicos e Domínios de Eventos

Os tópicos do Kafka são organizados por domínios funcionais, conforme descrito abaixo:

### Domínio de Autenticação

| Tópico | Descrição | Partições | Retenção |
|--------|-----------|-----------|----------|
| `iam-auth-events` | Eventos de autenticação (login, logout) | 12 | 7 dias |
| `iam-token-events` | Ciclo de vida de tokens | 12 | 12 horas |
| `iam-mfa-challenges` | Desafios de autenticação multifator | 6 | 24 horas |
| `iam-risk-scores` | Pontuações de risco para autenticação adaptativa | 6 | 7 dias |

### Domínio de Usuários e Inquilinos

| Tópico | Descrição | Partições | Retenção |
|--------|-----------|-----------|----------|
| `iam-user-events` | Operações de usuários | 6 | 14 dias |
| `iam-tenant-events` | Operações de inquilinos | 3 | 30 dias |
| `iam-sessions` | Gerenciamento de sessões | 12 | 24 horas |

### Domínio de Configuração e Gestão

| Tópico | Descrição | Partições | Retenção |
|--------|-----------|-----------|----------|
| `iam-method-updates` | Atualizações de métodos de autenticação | 3 | Compactado |
| `iam-auth-configurations` | Configurações de autenticação | 3 | Compactado |

### Domínio de Segurança e Compliance

| Tópico | Descrição | Partições | Retenção |
|--------|-----------|-----------|----------|
| `iam-security-alerts` | Alertas de segurança | 6 | 30 dias |
| `iam-audit-logs` | Logs de auditoria | 12 | 90 dias |
| `iam-user-deletion-requests` | Solicitações de exclusão (GDPR, LGPD) | 3 | 365 dias |
| `iam-data-subject-requests` | Solicitações de titulares de dados | 3 | 365 dias |
| `iam-consent-events` | Eventos de consentimento | 6 | 730 dias |

### Domínio de Operações

| Tópico | Descrição | Partições | Retenção |
|--------|-----------|-----------|----------|
| `iam-notification-events` | Eventos para notificações | 6 | 3 dias |
| `iam-healthchecks` | Verificações de saúde | 3 | 24 horas |
| `iam-deadletter` | Mensagens não processadas | 3 | 90 dias |

### Domínios Específicos por Região/Setor

| Tópico | Descrição | Partições | Retenção | Regiões Ativas |
|--------|-----------|-----------|----------|----------------|
| `iam-offline-auth-events` | Autenticação offline | 6 | 90 dias | AO |
| `iam-healthcare-auth-events` | Eventos específicos de saúde | 6 | 365 dias | US, EU |

## Fluxos de Eventos e Casos de Uso

### 1. Fluxo de Autenticação com MFA

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│    Login    │────▶│ Avaliação   │────▶│  Desafio    │────▶│ Verificação │
│   Inicial   │     │  de Risco   │     │    MFA      │     │    MFA      │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
       │                   │                   │                   │
       ▼                   ▼                   ▼                   ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│iam-auth-eve.│────▶│iam-risk-sco.│────▶│iam-mfa-chal.│────▶│iam-auth-eve.│
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                                                                    │
                                                                    ▼
                                                            ┌─────────────┐
                                                            │iam-token-ev.│
                                                            └─────────────┘
```

### 2. Fluxo de Auditoria e Compliance

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Eventos   │────▶│ Agregação e │────▶│ Relatórios  │
│     IAM     │     │ Normalização│     │   GDPR      │
└─────────────┘     └─────────────┘     └─────────────┘
       │                   │
       ▼                   ▼
┌─────────────┐     ┌─────────────┐
│iam-*-events │────▶│iam-audit-lo.│
└─────────────┘     └─────────────┘
```

### 3. Fluxo de Autenticação Adaptativa

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│    Login    │────▶│  Análise    │────▶│  Decisão    │────▶│   Fluxo de  │
│   Inicial   │     │ Contextual  │     │ Adaptativa  │     │Autenticação │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
       │                   │                   │                   │
       ▼                   ▼                   ▼                   ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│iam-auth-eve.│────▶│iam-risk-sco.│────▶│iam-auth-con.│────▶│iam-auth-eve.│
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
```

## Adaptações Regionais

A infraestrutura Kafka do módulo IAM é adaptada para atender aos requisitos específicos de cada região de implementação:

### União Europeia/Portugal (EU)

- **GDPR**: Habilitação de mascaramento de dados pessoais nos eventos
- **Retenção**: Políticas de retenção mais curtas para dados pessoais
- **Segurança**: Criptografia aprimorada e validação rigorosa de esquema
- **Auditoria**: Logs extensivos de acesso a dados

### Brasil (BR)

- **LGPD**: Conformidade com a Lei Geral de Proteção de Dados
- **ICP-Brasil**: Tópicos específicos para validação de certificados
- **Retenção**: Políticas conforme requisitos regulatórios locais

### Angola (AO)

- **Conectividade Intermitente**: Suporte a autenticação offline
- **PNDSB**: Conformidade com a regulamentação local
- **Armazenamento Estendido**: Maior tempo de retenção para eventos

### Estados Unidos (US)

- **Setores Específicos**: Tópicos dedicados para saúde (HIPAA) e finanças
- **Alta Performance**: Maior número de partições para maior paralelismo
- **Retenção Prolongada**: Conformidade com requisitos de auditoria

## Segurança e Criptografia

A segurança da infraestrutura Kafka do IAM inclui:

1. **Autenticação SASL/SSL**: Garante que apenas clientes autorizados possam se conectar
2. **Autorização ACL**: Controla quais clientes podem produzir/consumir de quais tópicos
3. **Criptografia TLS**: Protege os dados em trânsito
4. **Criptografia de Dados**: Protege dados sensíveis nos eventos
5. **Mascaramento de PII**: Não expõe informações de identificação pessoal nos eventos

## Observabilidade e Monitoramento

O sistema Kafka do IAM é monitorado através de:

1. **Métricas JMX**: Para performance dos brokers e clients
2. **Logs de Auditoria**: Para rastreamento de atividades sensíveis
3. **Alertas de Anomalias**: Para detecção de comportamentos inesperados
4. **Dashboards Operacionais**: Para visualização em tempo real
5. **Integração com OpenTelemetry**: Para rastreamento distribuído

## Procedimentos Operacionais

### Gerenciamento de Tópicos

```bash
# Criar um novo tópico
kafka-topics --bootstrap-server iam-kafka:9092 --create --topic iam-auth-events --partitions 12 --replication-factor 3

# Listar tópicos
kafka-topics --bootstrap-server iam-kafka:9092 --list

# Descrever um tópico
kafka-topics --bootstrap-server iam-kafka:9092 --describe --topic iam-auth-events
```

### Gerenciamento de Consumidores

```bash
# Listar grupos de consumidores
kafka-consumer-groups --bootstrap-server iam-kafka:9092 --list

# Descrever um grupo de consumidores
kafka-consumer-groups --bootstrap-server iam-kafka:9092 --describe --group iam-auth-consumer
```

### Backup e Recuperação

```bash
# Backup de tópico
kafka-mirror-maker --consumer.config consumer.properties --producer.config producer.properties --whitelist iam-audit-logs

# Restauração de tópico
kafka-console-consumer --bootstrap-server iam-kafka:9092 --topic backup.iam-audit-logs --from-beginning | \
kafka-console-producer --bootstrap-server iam-kafka:9092 --topic iam-audit-logs
```

## Integração com Outros Sistemas

### GraphQL e APIs REST

Os eventos Kafka são integrados com o gateway GraphQL e APIs REST do INNOVABIZ através de:

1. **Kafka Connect**: Para ingestão/exposição de dados para sistemas externos
2. **KSQLDB**: Para transformação de eventos em tempo real
3. **Schema Registry**: Para garantir compatibilidade de dados

### MCP (Model Context Protocol)

A integração com o MCP permite:

1. **Contextualização de Eventos**: Enriquecimento de eventos com informações contextuais
2. **Distribuição Inteligente**: Roteamento de eventos com base em metadados
3. **Rastreabilidade**: Correlação de eventos através de fluxos completos

## Melhores Práticas

1. **Desenvolvimento**:
   - Use as bibliotecas cliente oficiais do Kafka
   - Implemente tratamento de erros e retentativas para produtores
   - Defina claramente a semântica de processamento (at-least-once, exactly-once)

2. **Operação**:
   - Monitore o lag de consumidores em tempo real
   - Realize backups regulares dos tópicos críticos
   - Mantenha as versões dos brokers e clientes atualizadas

3. **Segurança**:
   - Revise as ACLs periodicamente
   - Monitore tentativas de acesso não autorizado
   - Atualize certificados antes da expiração

## Conformidade e Governança

A implementação do Kafka no módulo IAM atende aos seguintes requisitos de conformidade:

- **ISO/IEC 27001**: Segurança da informação
- **GDPR**: Proteção de dados na União Europeia
- **LGPD**: Proteção de dados no Brasil
- **HIPAA**: Para dados de saúde nos EUA
- **PCI DSS**: Para dados de pagamento
- **SOC 2**: Controles de segurança e disponibilidade

## Referências e Documentação Adicional

- [Apache Kafka Documentation](https://kafka.apache.org/documentation/)
- [Confluent Platform Documentation](https://docs.confluent.io/platform/current/overview.html)
- [Schema Registry Documentation](https://docs.confluent.io/platform/current/schema-registry/index.html)
- [KSQLDB Documentation](https://docs.ksqldb.io/)
- [Kafka Connect Documentation](https://docs.confluent.io/platform/current/connect/index.html)
