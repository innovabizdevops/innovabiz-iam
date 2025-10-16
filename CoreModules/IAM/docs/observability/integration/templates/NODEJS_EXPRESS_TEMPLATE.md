# Template para Instrumentação OpenTelemetry - Node.js (Express)

## Visão Geral

Este documento fornece o template padrão INNOVABIZ para instrumentação de aplicações Node.js (Express) com OpenTelemetry, garantindo consistência nas métricas, logs e traces coletados em toda a plataforma.

## Pré-requisitos

- Node.js versão 16.x ou superior
- Express.js 4.x ou superior
- Acesso ao OpenTelemetry Collector (via variáveis de ambiente)

## Instalação das Dependências

```bash
# Instalar dependências principais
npm install --save @opentelemetry/sdk-node @opentelemetry/resources @opentelemetry/semantic-conventions @opentelemetry/api

# Instalar exportadores
npm install --save @opentelemetry/exporter-trace-otlp-http @opentelemetry/exporter-metrics-otlp-http

# Instalar instrumentações automáticas
npm install --save @opentelemetry/instrumentation-express @opentelemetry/instrumentation-http @opentelemetry/instrumentation-mongodb @opentelemetry/instrumentation-pg @opentelemetry/instrumentation-redis @opentelemetry/instrumentation-dns

# Instalar propagadores
npm install --save @opentelemetry/core @opentelemetry/propagator-b3 @opentelemetry/propagator-jaeger

# Instalar SDK INNOVABIZ (opcional, mas recomendado)
npm install --save @innovabiz/observability-sdk
```

## Template de Implementação

### 1. Arquivo de Configuração - `src/config/telemetry.js`

```javascript
// Configuração do OpenTelemetry para Node.js (Express)
const { NodeSDK } = require('@opentelemetry/sdk-node');
const { Resource } = require('@opentelemetry/resources');
const { SemanticResourceAttributes } = require('@opentelemetry/semantic-conventions');
const { OTLPTraceExporter } = require('@opentelemetry/exporter-trace-otlp-http');
const { OTLPMetricExporter } = require('@opentelemetry/exporter-metrics-otlp-http');
const { ExpressInstrumentation } = require('@opentelemetry/instrumentation-express');
const { HttpInstrumentation } = require('@opentelemetry/instrumentation-http');
const { MongoDBInstrumentation } = require('@opentelemetry/instrumentation-mongodb');
const { PgInstrumentation } = require('@opentelemetry/instrumentation-pg');
const { RedisInstrumentation } = require('@opentelemetry/instrumentation-redis');
const { PeriodicExportingMetricReader } = require('@opentelemetry/sdk-metrics');
const { DnsInstrumentation } = require('@opentelemetry/instrumentation-dns');
const { W3CTraceContextPropagator } = require('@opentelemetry/core');
const { B3Propagator } = require('@opentelemetry/propagator-b3');
const { JaegerPropagator } = require('@opentelemetry/propagator-jaeger');

// Se disponível, importar propagador de contexto INNOVABIZ
let InnovabizContextPropagator;
try {
  const { InnovabizContextPropagator: ICP } = require('@innovabiz/observability-sdk');
  InnovabizContextPropagator = ICP;
} catch (e) {
  // Implementação fallback para o propagador INNOVABIZ
  class DefaultInnovabizContextPropagator {
    inject(context, carrier, setter) {
      const tenantId = process.env.TENANT_ID || 'default';
      const regionId = process.env.REGION_ID || 'default';
      
      setter(carrier, 'x-innovabiz-tenant-id', tenantId);
      setter(carrier, 'x-innovabiz-region-id', regionId);
      setter(carrier, 'x-innovabiz-context-version', '1.0');
    }
    
    extract(context, carrier, getter) {
      const tenantId = getter(carrier, 'x-innovabiz-tenant-id') || process.env.TENANT_ID || 'default';
      const regionId = getter(carrier, 'x-innovabiz-region-id') || process.env.REGION_ID || 'default';
      
      // Em uma implementação completa, estas informações seriam anexadas ao contexto
      return context;
    }
    
    fields() {
      return ['x-innovabiz-tenant-id', 'x-innovabiz-region-id', 'x-innovabiz-context-version'];
    }
  }
  
  InnovabizContextPropagator = DefaultInnovabizContextPropagator;
}

// Função de inicialização da telemetria
function initializeOpenTelemetry(serviceName, moduleId, serviceVersion) {
  // Informações multi-contexto INNOVABIZ
  const resourceAttributes = {
    [SemanticResourceAttributes.SERVICE_NAME]: serviceName,
    [SemanticResourceAttributes.SERVICE_VERSION]: serviceVersion,
    'innovabiz.module.id': moduleId,
    'innovabiz.deployment.environment': process.env.ENVIRONMENT || 'development',
    'innovabiz.tenant.id': process.env.TENANT_ID || 'default',
    'innovabiz.region.id': process.env.REGION_ID || 'default',
  };

  // Configuração do coletor
  const collectorOptions = {
    url: process.env.OTEL_EXPORTER_OTLP_ENDPOINT || 'http://localhost:4318/v1/traces',
    headers: {
      'x-innovabiz-tenant-id': process.env.TENANT_ID || 'default',
      'x-innovabiz-region-id': process.env.REGION_ID || 'default',
    },
  };

  // Configuração do pipeline de métricas
  const metricReaders = [
    new PeriodicExportingMetricReader({
      exporter: new OTLPMetricExporter({
        url: process.env.OTEL_EXPORTER_OTLP_METRICS_ENDPOINT || 'http://localhost:4318/v1/metrics',
        headers: {
          'x-innovabiz-tenant-id': process.env.TENANT_ID || 'default',
          'x-innovabiz-region-id': process.env.REGION_ID || 'default',
        },
      }),
      exportIntervalMillis: 15000,
    }),
  ];

  // Configuração de instrumentação automática
  const instrumentations = [
    new HttpInstrumentation({
      ignoreIncomingPaths: ['/health', '/metrics', '/ready'],
    }),
    new ExpressInstrumentation(),
    new MongoDBInstrumentation(),
    new PgInstrumentation(),
    new RedisInstrumentation(),
    new DnsInstrumentation(),
  ];

  // Configuração de propagadores de contexto
  const propagators = {
    registeredPropagators: [
      new W3CTraceContextPropagator(),
      new B3Propagator(),
      new JaegerPropagator(),
      new InnovabizContextPropagator(), // Propagador personalizado INNOVABIZ
    ],
  };

  // Inicializa o SDK
  const sdk = new NodeSDK({
    resource: new Resource(resourceAttributes),
    traceExporter: new OTLPTraceExporter(collectorOptions),
    metricReaders: metricReaders,
    instrumentations: instrumentations,
    propagators: propagators,
  });

  // Inicializa e registra manipuladores de erro/desligamento
  sdk.start()
    .then(() => console.log('Telemetria iniciada com sucesso'))
    .catch((error) => console.error('Erro ao iniciar telemetria:', error));

  // Manipulador de desligamento
  process.on('SIGTERM', () => {
    sdk.shutdown()
      .then(() => console.log('Telemetria encerrada com sucesso'))
      .catch((error) => console.error('Erro ao encerrar telemetria:', error))
      .finally(() => process.exit(0));
  });

  return sdk;
}

// Exemplos de uso de métricas customizadas
function registerCustomMetrics() {
  const meter = metrics.getMeter('innovabiz-custom-metrics');
  
  // Counter para transações
  const transactionCounter = meter.createCounter('transactions.count', {
    description: 'Contador de transações processadas',
    unit: '1',
  });
  
  // Histogram para latência de transações
  const transactionDuration = meter.createHistogram('transaction.duration', {
    description: 'Duração das transações',
    unit: 'ms',
  });
  
  // Up/Down Counter para usuários ativos
  const activeUsers = meter.createUpDownCounter('users.active', {
    description: 'Usuários ativos no momento',
    unit: '1',
  });

  // Gauge para utilização de recursos
  const cpuUsage = meter.createObservableGauge('system.cpu.usage', {
    description: 'Uso de CPU',
    unit: '%',
    callback: (result) => {
      // Lógica para obter uso de CPU
      const usage = process.cpuUsage();
      result.observe(((usage.user + usage.system) / 1000000) * 100, {
        'innovabiz.resource.type': 'cpu',
      });
    },
  });

  return {
    transactionCounter,
    transactionDuration,
    activeUsers,
    cpuUsage,
  };
}

// Exporta funções
module.exports = {
  initializeOpenTelemetry,
  registerCustomMetrics,
};
```

### 2. Integração no Arquivo Principal - `src/server.js`

```javascript
require('dotenv').config();
const express = require('express');
const { metrics } = require('@opentelemetry/api');
const { initializeOpenTelemetry, registerCustomMetrics } = require('./config/telemetry');

// Inicializa OpenTelemetry antes de importar outros módulos
const sdk = initializeOpenTelemetry(
  'payment-gateway-api', 
  'payment-gateway', 
  '1.0.0'
);

// Registra métricas customizadas
const customMetrics = registerCustomMetrics();

const app = express();

// Middlewares padrão
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// Middleware para injeção de contexto multi-dimensional
app.use((req, res, next) => {
  const tenantId = req.headers['x-tenant-id'] || process.env.TENANT_ID || 'default';
  const regionId = req.headers['x-region-id'] || process.env.REGION_ID || 'default';
  
  // Define valores para uso em toda a solicitação
  req.tenantId = tenantId;
  req.regionId = regionId;
  
  // Adiciona ao cabeçalho de resposta para propagação
  res.setHeader('x-innovabiz-tenant-id', tenantId);
  res.setHeader('x-innovabiz-region-id', regionId);
  
  next();
});

// Endpoint de exemplo com instrumentação
app.post('/api/v1/transactions', (req, res) => {
  const startTime = Date.now();
  
  // Incrementa contador de transações com atributos multi-dimensionais
  customMetrics.transactionCounter.add(1, {
    'innovabiz.tenant.id': req.tenantId,
    'innovabiz.region.id': req.regionId,
    'transaction.type': req.body.type || 'default',
    'payment.method': req.body.paymentMethod || 'unknown',
  });
  
  // Lógica de processamento...
  
  // Registra duração da transação
  const duration = Date.now() - startTime;
  customMetrics.transactionDuration.record(duration, {
    'innovabiz.tenant.id': req.tenantId,
    'innovabiz.region.id': req.regionId,
    'transaction.type': req.body.type || 'default',
    'payment.method': req.body.paymentMethod || 'unknown',
  });
  
  res.status(200).json({ status: 'success' });
});

// Health check endpoint
app.get('/health', (req, res) => {
  res.status(200).json({ status: 'ok' });
});

// Endpoint de métricas personalizado
app.get('/metrics', (req, res) => {
  // Neste ponto, o OpenTelemetry já está exportando métricas
  // Este endpoint é apenas para compatibilidade ou debug
  res.status(200).json({ status: 'metrics available via OpenTelemetry collector' });
});

// Inicializa o servidor
const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`Servidor rodando na porta ${PORT}`);
});

// Manipuladores de desligamento
process.on('SIGINT', () => {
  console.log('SIGINT recebido. Encerrando aplicação...');
  sdk.shutdown()
    .then(() => console.log('Telemetria encerrada com sucesso'))
    .catch((error) => console.error('Erro ao encerrar telemetria:', error))
    .finally(() => process.exit(0));
});
```

## Configuração de Variáveis de Ambiente

Crie um arquivo `.env` com as seguintes variáveis:

```
# Ambiente
ENVIRONMENT=development
NODE_ENV=development

# Contexto Multi-dimensional INNOVABIZ
TENANT_ID=default
REGION_ID=br

# OpenTelemetry
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318/v1/traces
OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=http://otel-collector:4318/v1/metrics
OTEL_LOG_LEVEL=info
OTEL_RESOURCE_ATTRIBUTES=service.name=payment-gateway-api,service.version=1.0.0,innovabiz.module.id=payment-gateway

# Configurações de amostragem
OTEL_TRACES_SAMPLER=parentbased_traceidratio
OTEL_TRACES_SAMPLER_ARG=1.0

# Configurações de segurança
OTEL_EXPORTER_OTLP_HEADERS=x-innovabiz-tenant-id=default,x-innovabiz-region-id=br
```

## Melhores Práticas

1. **Nomenclatura de métricas**
   - Use `snake_case` para nomes de métricas
   - Siga o padrão `domínio.entidade.ação` (ex: `transactions.count`, `api.requests.duration`)
   - Use unidades padronizadas (ms, bytes, %, 1)

2. **Atributos obrigatórios para contexto multi-dimensional**
   - `innovabiz.tenant.id` - Identificador do tenant
   - `innovabiz.region.id` - Identificador da região
   - `innovabiz.module.id` - Identificador do módulo
   - `innovabiz.deployment.environment` - Ambiente de implantação

3. **Métricas essenciais por serviço**
   - Latência/duração das operações (histograms)
   - Contadores de operações (counters)
   - Taxa de erros (counters com atributos de status)
   - Utilização de recursos (gauges)
   - Estado do serviço (gauges)

4. **Propagação de contexto**
   - Propague sempre os cabeçalhos de contexto multi-dimensional entre serviços
   - Use o propagador INNOVABIZ para compatibilidade com toda a plataforma
   - Verifique sempre os cabeçalhos de entrada para extração do contexto

5. **Segurança e Compliance**
   - Não inclua dados sensíveis (PCI DSS, GDPR, LGPD) em métricas, logs ou traces
   - Utilize mascaramento ou remoção de dados sensíveis
   - Implemente controles de acesso (RBAC/ABAC) para visualização dos dados

## Checklist de Validação

- [ ] SDK OpenTelemetry inicializado antes de qualquer outro código
- [ ] Atributos de contexto multi-dimensional configurados corretamente
- [ ] Instrumentação automática configurada para todas as bibliotecas relevantes
- [ ] Métricas customizadas registradas conforme padrões INNOVABIZ
- [ ] Propagadores de contexto configurados corretamente
- [ ] Manipuladores de desligamento implementados
- [ ] Variáveis de ambiente documentadas
- [ ] Endpoints de health check implementados
- [ ] Dados sensíveis protegidos em conformidade com políticas de segurança
- [ ] Testes de verificação de telemetria implementados

## Recursos Adicionais

- [Documentação OpenTelemetry](https://opentelemetry.io/docs/nodejs/)
- [Portal de Observabilidade INNOVABIZ](https://observability.innovabiz.com)
- [Repositório de Dashboards Padrão](https://github.com/innovabiz/observability-dashboards)
- [Guia de Troubleshooting](https://wiki.innovabiz.com/observability/troubleshooting)