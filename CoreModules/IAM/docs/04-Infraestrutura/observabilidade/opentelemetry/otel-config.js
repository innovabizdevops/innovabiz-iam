/**
 * @fileoverview Configuração principal do OpenTelemetry para a plataforma INNOVABIZ
 * @author INNOVABIZ Dev Team
 * @version 1.0.0
 * @module iam/infraestrutura/observabilidade/opentelemetry/otel-config
 * @description Configuração centralizada do OpenTelemetry para métricas, traces e logs
 */

const opentelemetry = require('@opentelemetry/sdk-node');
const { getNodeAutoInstrumentations } = require('@opentelemetry/auto-instrumentations-node');
const { OTLPTraceExporter } = require('@opentelemetry/exporter-trace-otlp-proto');
const { OTLPMetricExporter } = require('@opentelemetry/exporter-metrics-otlp-proto');
const { Resource } = require('@opentelemetry/resources');
const { SemanticResourceAttributes } = require('@opentelemetry/semantic-conventions');
const { JaegerExporter } = require('@opentelemetry/exporter-jaeger');
const { PrometheusExporter } = require('@opentelemetry/exporter-prometheus');
const { CompressionAlgorithm } = require('@opentelemetry/otlp-exporter-base');
const { HttpInstrumentation } = require('@opentelemetry/instrumentation-http');
const { ExpressInstrumentation } = require('@opentelemetry/instrumentation-express');
const { KafkaJsInstrumentation } = require('@opentelemetry/instrumentation-kafkajs');
const { PgInstrumentation } = require('@opentelemetry/instrumentation-pg');
const { MongoDBInstrumentation } = require('@opentelemetry/instrumentation-mongodb');
const { RedisInstrumentation } = require('@opentelemetry/instrumentation-redis');
const { GraphQLInstrumentation } = require('@opentelemetry/instrumentation-graphql');
const { AwsInstrumentation } = require('@opentelemetry/instrumentation-aws-sdk');
const { B3Propagator } = require('@opentelemetry/propagator-b3');
const { W3CTraceContextPropagator } = require('@opentelemetry/core');

/**
 * Configurações regionais para observabilidade
 */
const REGIONAL_CONFIGS = {
  EU: {
    jaeger_endpoint: process.env.JAEGER_ENDPOINT_EU || 'http://jaeger-collector-eu:14268/api/traces',
    otlp_endpoint: process.env.OTLP_ENDPOINT_EU || 'http://otel-collector-eu:4318',
    prometheus_endpoint: process.env.PROMETHEUS_ENDPOINT_EU || '/metrics',
    prometheus_port: parseInt(process.env.PROMETHEUS_PORT_EU || '9464', 10),
    sampling_ratio: 0.3,
    log_level: 'info',
    health_checks: true,
    compliance_metrics: {
      gdpr: true,
      eidas: true
    },
    sector_specific: {
      HEALTHCARE: {
        gdpr_healthcare_metrics: true,
        phi_tracking: true,
        data_residency_tracking: true
      }
    }
  },
  BR: {
    jaeger_endpoint: process.env.JAEGER_ENDPOINT_BR || 'http://jaeger-collector-br:14268/api/traces',
    otlp_endpoint: process.env.OTLP_ENDPOINT_BR || 'http://otel-collector-br:4318',
    prometheus_endpoint: process.env.PROMETHEUS_ENDPOINT_BR || '/metrics',
    prometheus_port: parseInt(process.env.PROMETHEUS_PORT_BR || '9464', 10),
    sampling_ratio: 0.3,
    log_level: 'info',
    health_checks: true,
    compliance_metrics: {
      lgpd: true,
      icp_brasil: true
    },
    sector_specific: {
      HEALTHCARE: {
        lgpd_healthcare_metrics: true,
        phi_tracking: true,
        consent_tracking: true
      }
    }
  },
  AO: {
    jaeger_endpoint: process.env.JAEGER_ENDPOINT_AO || 'http://jaeger-collector-ao:14268/api/traces',
    otlp_endpoint: process.env.OTLP_ENDPOINT_AO || 'http://otel-collector-ao:4318',
    prometheus_endpoint: process.env.PROMETHEUS_ENDPOINT_AO || '/metrics',
    prometheus_port: parseInt(process.env.PROMETHEUS_PORT_AO || '9464', 10),
    sampling_ratio: 0.1,
    log_level: 'info',
    health_checks: true,
    compliance_metrics: {
      pndsb: true
    },
    sector_specific: {
      HEALTHCARE: {
        pndsb_healthcare_metrics: true,
        offline_tracking: true
      }
    }
  },
  US: {
    jaeger_endpoint: process.env.JAEGER_ENDPOINT_US || 'http://jaeger-collector-us:14268/api/traces',
    otlp_endpoint: process.env.OTLP_ENDPOINT_US || 'http://otel-collector-us:4318',
    prometheus_endpoint: process.env.PROMETHEUS_ENDPOINT_US || '/metrics',
    prometheus_port: parseInt(process.env.PROMETHEUS_PORT_US || '9464', 10),
    sampling_ratio: 0.5,
    log_level: 'info',
    health_checks: true,
    compliance_metrics: {
      hipaa: true,
      nist: true,
      pci_dss: true
    },
    sector_specific: {
      HEALTHCARE: {
        hipaa_metrics: true,
        phi_tracking: true,
        data_access_audit: true
      }
    }
  }
};

/**
 * Classe de configuração do OpenTelemetry
 */
class OpenTelemetryConfig {
  /**
   * Inicializa a configuração do OpenTelemetry
   * @param {Object} options - Opções de configuração
   * @param {string} options.serviceName - Nome do serviço
   * @param {string} options.serviceVersion - Versão do serviço
   * @param {string} options.regionCode - Código da região (EU, BR, AO, US)
   * @param {string} options.environment - Ambiente (production, development, etc.)
   * @param {Object} options.customAttributes - Atributos personalizados
   * @param {boolean} options.enableTracing - Habilitar tracing
   * @param {boolean} options.enableMetrics - Habilitar métricas
   * @param {Object} options.sectorConfig - Configurações específicas do setor
   */
  constructor(options = {}) {
    this.serviceName = options.serviceName || process.env.SERVICE_NAME || 'innovabiz-service';
    this.serviceVersion = options.serviceVersion || process.env.SERVICE_VERSION || '1.0.0';
    this.regionCode = options.regionCode || process.env.REGION_CODE || 'EU';
    this.environment = options.environment || process.env.NODE_ENV || 'development';
    this.customAttributes = options.customAttributes || {};
    this.enableTracing = options.enableTracing !== false;
    this.enableMetrics = options.enableMetrics !== false;
    this.sectorConfig = options.sectorConfig || null;
    
    // Carregar configurações regionais
    this.config = REGIONAL_CONFIGS[this.regionCode] || REGIONAL_CONFIGS.EU;
    
    // Se houver configuração específica do setor, aplicar
    if (this.sectorConfig && this.config.sector_specific && 
        this.config.sector_specific[this.sectorConfig.sector]) {
      this.sectorMetrics = this.config.sector_specific[this.sectorConfig.sector];
    }
    
    // Criar recurso de telemetria
    this.resource = new Resource({
      [SemanticResourceAttributes.SERVICE_NAME]: this.serviceName,
      [SemanticResourceAttributes.SERVICE_VERSION]: this.serviceVersion,
      [SemanticResourceAttributes.DEPLOYMENT_ENVIRONMENT]: this.environment,
      'region.code': this.regionCode,
      ...(this.sectorConfig ? { 'sector': this.sectorConfig.sector } : {}),
      ...this.customAttributes
    });
    
    console.log(`OpenTelemetry configurado para serviço ${this.serviceName} na região ${this.regionCode}`);
  }
  
  /**
   * Configura instrumentações automáticas para Node.js
   * @private
   * @returns {Array} Array de instrumentações configuradas
   */
  getInstrumentations() {
    // Instrumentações padrão para Node.js
    const instrumentations = [
      new HttpInstrumentation({
        ignoreIncomingPaths: ['/health', '/metrics'],
        headersToSpanAttributes: {
          requestHeaders: ['x-tenant-id', 'x-correlation-id', 'x-region-code'],
          responseHeaders: ['content-length', 'content-type']
        }
      }),
      new ExpressInstrumentation(),
      new KafkaJsInstrumentation({
        // Capturar conteúdo das mensagens em ambiente de desenvolvimento
        // para facilitar depuração
        messageAttributes: this.environment === 'development'
      }),
      new PgInstrumentation(),
      new MongoDBInstrumentation(),
      new RedisInstrumentation(),
      new GraphQLInstrumentation(),
      new AwsInstrumentation({
        suppressInternalInstrumentation: false
      })
    ];
    
    return instrumentations;
  }
  
  /**
   * Configura propagadores para contexto de tracing
   * @private
   * @returns {Object} Objeto com propagadores configurados
   */
  getPropagators() {
    return {
      b3: new B3Propagator(),
      w3c: new W3CTraceContextPropagator()
    };
  }
  
  /**
   * Cria exportador OTLP para traces
   * @private
   * @returns {OTLPTraceExporter} Exportador OTLP configurado
   */
  createOtlpTraceExporter() {
    return new OTLPTraceExporter({
      url: `${this.config.otlp_endpoint}/v1/traces`,
      headers: {
        'x-region-code': this.regionCode
      },
      compression: CompressionAlgorithm.GZIP
    });
  }
  
  /**
   * Cria exportador Jaeger para traces
   * @private
   * @returns {JaegerExporter} Exportador Jaeger configurado
   */
  createJaegerExporter() {
    return new JaegerExporter({
      endpoint: this.config.jaeger_endpoint,
      maxPacketSize: 65000
    });
  }
  
  /**
   * Cria exportador OTLP para métricas
   * @private
   * @returns {OTLPMetricExporter} Exportador OTLP configurado
   */
  createOtlpMetricExporter() {
    return new OTLPMetricExporter({
      url: `${this.config.otlp_endpoint}/v1/metrics`,
      headers: {
        'x-region-code': this.regionCode
      },
      compression: CompressionAlgorithm.GZIP
    });
  }
  
  /**
   * Cria exportador Prometheus para métricas
   * @private
   * @returns {PrometheusExporter} Exportador Prometheus configurado
   */
  createPrometheusExporter() {
    return new PrometheusExporter({
      port: this.config.prometheus_port,
      endpoint: this.config.prometheus_endpoint,
      preventServerStart: false
    });
  }
  
  /**
   * Inicializa o SDK do OpenTelemetry
   * @returns {opentelemetry.NodeSDK} SDK configurado
   */
  initializeSDK() {
    const traceExporter = this.createOtlpTraceExporter();
    const metricExporter = this.createOtlpMetricExporter();
    const propagator = this.getPropagators();
    
    // Configurar SDK
    const sdk = new opentelemetry.NodeSDK({
      resource: this.resource,
      traceExporter,
      metricExporter,
      instrumentations: this.getInstrumentations(),
      textMapPropagator: propagator.w3c,
      sampler: {
        shouldSample: () => {
          // Aplicar amostragem baseada na configuração regional
          return Math.random() < this.config.sampling_ratio;
        }
      }
    });
    
    return sdk;
  }
  
  /**
   * Inicializa o OpenTelemetry
   * @returns {Promise<void>}
   */
  async initialize() {
    try {
      const sdk = this.initializeSDK();
      
      // Registrar instrumentação
      await sdk.start();
      
      // Configurar também exportador Prometheus para métricas
      if (this.enableMetrics) {
        this.prometheusExporter = this.createPrometheusExporter();
        console.log(`Prometheus exporter iniciado em ${this.config.prometheus_endpoint}:${this.config.prometheus_port}`);
      }
      
      // Registrar para shutdown adequado
      process.on('SIGTERM', () => {
        sdk.shutdown()
          .then(() => console.log('OpenTelemetry SDK desligado'))
          .catch((error) => console.error('Erro ao desligar OpenTelemetry SDK', error))
          .finally(() => process.exit(0));
      });
      
      console.log(`OpenTelemetry inicializado com sucesso para ${this.serviceName}`);
      
      return sdk;
    } catch (error) {
      console.error(`Erro ao inicializar OpenTelemetry: ${error}`);
      throw error;
    }
  }
  
  /**
   * Configura atributos específicos para compliance
   * @param {Object} attributes - Atributos base
   * @returns {Object} Atributos enriquecidos com informações de compliance
   */
  enrichWithComplianceAttributes(attributes = {}) {
    const complianceAttributes = { ...attributes };
    
    // Adicionar atributos de compliance com base na região
    if (this.config.compliance_metrics) {
      for (const [regulation, enabled] of Object.entries(this.config.compliance_metrics)) {
        if (enabled) {
          complianceAttributes[`compliance.${regulation}.enabled`] = true;
        }
      }
    }
    
    // Adicionar atributos específicos do setor, se disponíveis
    if (this.sectorMetrics) {
      for (const [metric, enabled] of Object.entries(this.sectorMetrics)) {
        if (enabled) {
          complianceAttributes[`sector.${this.sectorConfig.sector.toLowerCase()}.${metric}`] = true;
        }
      }
    }
    
    return complianceAttributes;
  }
  
  /**
   * Retorna a configuração atual para uso externo
   * @returns {Object} Configuração atual
   */
  getConfig() {
    return {
      serviceName: this.serviceName,
      serviceVersion: this.serviceVersion,
      regionCode: this.regionCode,
      environment: this.environment,
      regionalConfig: this.config,
      sectorConfig: this.sectorConfig,
      sectorMetrics: this.sectorMetrics
    };
  }
}

module.exports = OpenTelemetryConfig;
