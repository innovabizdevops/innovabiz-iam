# Guia de Integração Kafka para o Módulo IAM

**Autor:** INNOVABIZ Dev Team  
**Versão:** 1.0.0  
**Data:** Maio/2025  
**Status:** Produção  
**Classificação:** Interno  

## Sumário

1. [Introdução](#introdução)
2. [Arquitetura de Eventos](#arquitetura-de-eventos)
3. [Estrutura de Tópicos](#estrutura-de-tópicos)
4. [Produtores e Consumidores](#produtores-e-consumidores)
5. [Adaptação Regional](#adaptação-regional)
6. [Integração MCP](#integração-mcp)
7. [Integração com Setor de Saúde](#integração-com-setor-de-saúde)
8. [Conformidade e Segurança](#conformidade-e-segurança)
9. [Observabilidade](#observabilidade)
10. [Troubleshooting](#troubleshooting)
11. [Referências](#referências)

## Introdução

Este documento descreve como integrar com a infraestrutura de eventos Apache Kafka do módulo IAM da plataforma INNOVABIZ. O Apache Kafka é utilizado como middleware de mensageria assíncrona para garantir comunicação confiável, distribuída e escalável entre os diversos componentes do módulo IAM e outros módulos da plataforma.

### Objetivos

- Fornecer um guia completo para desenvolvedores que precisam integrar-se com eventos IAM
- Explicar a estrutura de tópicos e esquemas de dados utilizados
- Detalhar as considerações específicas de cada região de implementação
- Descrever os componentes disponíveis para produção e consumo de eventos

### Pré-requisitos

- Conhecimento básico sobre Apache Kafka e processamento de eventos
- Acesso às configurações do ambiente INNOVABIZ
- Compreensão dos fluxos de autenticação e autorização do módulo IAM

## Arquitetura de Eventos

O módulo IAM adota uma arquitetura orientada a eventos (Event-Driven Architecture) que promove:

- **Desacoplamento**: Serviços podem evoluir independentemente
- **Escalabilidade**: Componentes podem ser escalados conforme necessidade
- **Resiliência**: Falhas são isoladas e não propagadas
- **Auditabilidade**: Todas as operações são registradas como eventos imutáveis
- **Adaptação Regional**: Conformidade com regulamentações específicas

### Fluxo de Eventos

![Fluxo de Eventos IAM](/docs/iam/03-Desenvolvimento/kafka/docs/images/event-flow-diagram.png)

O fluxo típico de eventos IAM segue estas etapas:

1. Um serviço produtor gera um evento (ex: tentativa de login)
2. O evento é serializado seguindo o esquema Avro registrado no Schema Registry
3. O evento é publicado no tópico Kafka apropriado
4. Múltiplos consumidores processam o evento conforme suas necessidades
5. Eventos são adaptados para o protocolo MCP para integração com outros módulos
6. Eventos de auditoria são gerados como subproduto do processamento

## Estrutura de Tópicos

Os tópicos Kafka do módulo IAM são organizados por domínios funcionais:

### Domínio de Autenticação

| Tópico | Descrição | Retenção | Partições |
|--------|-----------|----------|-----------|
| `iam-auth-events` | Eventos de autenticação (login, logout) | 7 dias | 12 |
| `iam-token-events` | Ciclo de vida de tokens | 12 horas | 12 |
| `iam-mfa-challenges` | Desafios de autenticação multifator | 24 horas | 6 |

### Domínio de Usuários

| Tópico | Descrição | Retenção | Partições |
|--------|-----------|----------|-----------|
| `iam-user-events` | Operações de usuários | 14 dias | 6 |
| `iam-sessions` | Gerenciamento de sessões | 24 horas | 12 |

### Domínio de Segurança

| Tópico | Descrição | Retenção | Partições |
|--------|-----------|----------|-----------|
| `iam-security-alerts` | Alertas de segurança | 30 dias | 6 |
| `iam-audit-logs` | Logs de auditoria | 90 dias | 12 |

### Domínios Específicos por Região/Setor

| Tópico | Descrição | Regiões Aplicáveis | Partições |
|--------|-----------|-------------------|-----------|
| `iam-offline-auth-events` | Autenticação offline | AO | 6 |
| `iam-healthcare-auth-events` | Eventos específicos de saúde | US, EU, BR | 6 |

## Produtores e Consumidores

O módulo IAM fornece componentes para simplificar a integração com a infraestrutura Kafka:

### Produtores

A classe principal para produção de eventos é a `AuthEventProducer`, que oferece:

- Gerenciamento automático de conexões
- Serialização Avro com Schema Registry
- Mascaramento regional de dados sensíveis
- Gerenciamento de transações
- Suporte para lotes de eventos

**Exemplo de uso:**

```javascript
const { AuthEventProducer } = require('iam/auth-framework/kafka/auth-event-producer');

// Criar instância com configuração regional
const producer = new AuthEventProducer({
  regionCode: 'BR',  // Região de execução
  schemaRegistryUrl: 'http://iam-schema-registry:8081'
});

// Conectar ao Kafka
await producer.connect();

// Publicar evento de autenticação
const result = await producer.publishAuthEvent({
  event_type: 'LOGIN_SUCCESS',
  tenant_id: 'acme-corp',
  user_id: '123e4567-e89b-12d3-a456-426614174000',
  method_code: 'K01',
  status: 'SUCCESS',
  timestamp: Date.now()
});

console.log(`Evento publicado: ${result.event_id}`);

// Desconectar ao finalizar
await producer.disconnect();
```

### Consumidores

A classe principal para consumo de eventos é a `AuthEventConsumer`, que oferece:

- Gerenciamento de grupos de consumidores
- Deserialização automática de eventos Avro
- Manipuladores de eventos configuráveis
- Processamento adaptado por região
- Suporte a DLQ (Dead Letter Queue)

**Exemplo de uso:**

```javascript
const { AuthEventConsumer } = require('iam/auth-framework/kafka/auth-event-consumer');

// Criar instância com configuração regional
const consumer = new AuthEventConsumer({
  regionCode: 'EU',
  groupId: 'my-service-consumer-group',
  eventHandlers: {
    // Handler personalizado para LOGIN_SUCCESS
    LOGIN_SUCCESS: async (event, headers) => {
      console.log(`Processando login bem-sucedido: ${event.user_id}`);
      // Lógica personalizada aqui
      return { processed: true, action: 'update-cache' };
    }
  }
});

// Conectar e iniciar consumo
await consumer.connect();
await consumer.subscribe(['iam-auth-events']);
await consumer.run();

// Para desligar adequadamente
process.on('SIGTERM', async () => {
  await consumer.shutdown();
});
```

## Adaptação Regional

A infraestrutura Kafka do IAM é adaptada para atender às especificidades regulatórias de cada região:

### União Europeia (EU)

- Mascaramento de dados pessoais em eventos (GDPR)
- Verificação estrita de consentimento
- Limitação de retenção para dados de autenticação
- Validação de autenticação eIDAS

### Brasil (BR)

- Conformidade com a LGPD
- Suporte ao ICP-Brasil para validação de certificados
- Políticas regionais de backup e retenção

### Angola (AO)

- Suporte a autenticação offline
- Menor exigência de mascaramento de dados
- Políticas adaptadas à PNDSB (Política Nacional de Dados)

### Estados Unidos (US)

- Validação HIPAA para eventos de saúde
- Conformidade SOC 2 e PCI DSS
- Políticas específicas por setor (saúde, finanças)

## Integração MCP

O Protocolo de Contexto de Modelo (MCP - Model Context Protocol) é utilizado para integrar eventos Kafka com outros serviços da plataforma INNOVABIZ.

### Adaptador MCP

A classe `MCPKafkaAdapter` facilita a conversão entre eventos Kafka e mensagens MCP:

- Associação automática entre tópicos Kafka e canais MCP
- Enriquecimento de contexto
- Roteamento baseado em atributos da mensagem
- Rastreabilidade entre eventos

**Exemplo de uso:**

```javascript
const { MCPKafkaAdapter } = require('iam/auth-framework/kafka/mcp/mcp-kafka-adapter');

// Criar adaptador MCP
const mcpAdapter = new MCPKafkaAdapter({
  regionCode: 'EU',
  contextEnrichment: true
});

// Conectar ao broker MCP
await mcpAdapter.connect();

// Publicar evento Kafka como mensagem MCP
await mcpAdapter.publishToMCP(
  'iam-auth-events',
  authEvent,
  { tenant: 'acme-corp' }
);

// Subscrever a um canal MCP e converter mensagens em eventos Kafka
await mcpAdapter.subscribeMCP('auth.events.eu', async (kafkaEvent) => {
  console.log(`Evento MCP recebido: ${kafkaEvent.event_id}`);
  // Processar o evento
});
```

## Integração com Setor de Saúde

A plataforma INNOVABIZ oferece adaptadores especializados para integração com o setor de saúde, considerando os requisitos regulatórios específicos.

### Adaptador para Eventos de Saúde

A classe `HealthcareAuthEventAdapter` fornece:

- Validação específica de compliance (HIPAA, GDPR Healthcare, LGPD Healthcare)
- Mascaramento de dados PHI (Protected Health Information)
- Integração com sistemas de saúde através do MCP
- Geração de relatórios de compliance

**Exemplo de uso:**

```javascript
const { HealthcareAuthEventAdapter } = 
  require('iam/auth-framework/kafka/healthcare/healthcare-auth-event-adapter');

// Criar adaptador para eventos de saúde
const healthcareAdapter = new HealthcareAuthEventAdapter({
  regionCode: 'US',
  enableComplianceValidation: true,
  enableObservability: true
});

// Publicar evento específico de saúde
await healthcareAdapter.publishHealthcareAuthEvent({
  event_type: 'LOGIN_SUCCESS',
  tenant_id: 'hospital-central',
  user_id: '123e4567-e89b-12d3-a456-426614174000',
  method_code: 'K05',  // OTP como segundo fator (requerido para HIPAA)
  status: 'SUCCESS',
  additional_context: {
    phi_access: true,
    department: 'radiology'
  }
});
```

## Conformidade e Segurança

A infraestrutura Kafka do IAM implementa diversas medidas de segurança:

### Autenticação e Autorização

- SASL/SSL para autenticação de clientes
- ACLs para controle de acesso a tópicos
- Autenticação mútua TLS

### Proteção de Dados

- Criptografia em trânsito (TLS 1.3)
- Mascaramento de dados sensíveis
- Sanitização de dados pessoais conforme regulações

### Auditoria

- Registro detalhado de todas as operações
- Rastreabilidade completa de eventos
- Possibilidade de reconstrução de estados históricos

## Observabilidade

A plataforma oferece métricas detalhadas para monitoramento:

### Métricas Kafka

- Latência de publicação/consumo
- Taxa de throughput por tópico
- Lag de consumidores
- Erros de produção/consumo

### Métricas de Domínio

- Taxa de eventos de autenticação
- Distribuição de métodos de autenticação
- Taxa de sucesso/falha
- Tentativas de MFA

### Integração OpenTelemetry

Todos os componentes Kafka integram-se com a infraestrutura de observabilidade OpenTelemetry da plataforma INNOVABIZ.

## Troubleshooting

### Problemas Comuns e Soluções

| Problema | Causa Provável | Solução |
|----------|----------------|---------|
| `SchemaRegistryError` | Incompatibilidade de schema | Verificar se o evento segue o schema registrado |
| `KafkaConnectionError` | Problemas de rede ou credenciais | Verificar conectividade e configurações SASL |
| `DeserializationError` | Problema no formato de dados | Validar os tipos de dados e codificação |
| `ConsumerGroupRebalance` | Adição/remoção de consumidores | Operação normal, verificar se todos instâncias se recuperam |

### Logs e Diagnóstico

Os componentes Kafka utilizam o logger padrão da plataforma INNOVABIZ. Para habilitar logs mais detalhados:

```javascript
// Configurar nível de log para componentes Kafka
logger.setLevel('kafka', 'DEBUG');
```

## Referências

- [Documentação Completa do Schema Registry](/docs/iam/04-Infraestrutura/kafka/schemas/README.md)
- [Guia de Conformidade Regional](/docs/iam/05-Seguranca/compliance/regional-compliance-guide.md)
- [Plano de Capacidade Kafka](/docs/iam/04-Infraestrutura/kafka/capacity-planning.md)
- [Guia de Integração MCP](/docs/iam/03-Desenvolvimento/mcp/integration-guide.md)
- [Documentação Apache Kafka](https://kafka.apache.org/documentation/)
