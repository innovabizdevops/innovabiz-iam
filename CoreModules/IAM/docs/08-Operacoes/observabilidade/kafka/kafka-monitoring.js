/**
 * @fileoverview Sistema de monitoramento e observabilidade para integração Kafka do IAM
 * @author INNOVABIZ Dev Team
 * @version 1.0.0
 * @module iam/operacoes/observabilidade/kafka/kafka-monitoring
 * @description Responsável pelo monitoramento, métricas e observabilidade da integração Kafka
 */

const { MeterProvider } = require('@opentelemetry/sdk-metrics');
const { PrometheusExporter } = require('@opentelemetry/exporter-prometheus');
const { Resource } = require('@opentelemetry/resources');
const { SemanticResourceAttributes } = require('@opentelemetry/semantic-conventions');
const { trace, context, propagation } = require('@opentelemetry/api');
const { NodeTracerProvider } = require('@opentelemetry/sdk-trace-node');
const { SimpleSpanProcessor } = require('@opentelemetry/sdk-trace-base');
const { JaegerExporter } = require('@opentelemetry/exporter-jaeger');
const logger = require('../../../core/utils/logging');

/**
 * Configurações regionais para observabilidade
 */
const REGIONAL_CONFIGS = {
  EU: {
    metrics_interval_ms: 15000,
    trace_sampling_ratio: 0.3,
    detailed_compliance_metrics: true,
    compliance_log_level: 'INFO',
    jaeger_endpoint: process.env.JAEGER_ENDPOINT_EU || 'http://jaeger-collector-eu:14268/api/traces',
    sector_specific: {
      HEALTHCARE: {
        gdpr_healthcare_metrics: true,
        phi_access_tracking: true,
        consent_verification_metrics: true
      }
    }
  },
  BR: {
    metrics_interval_ms: 15000,
    trace_sampling_ratio: 0.3,
    detailed_compliance_metrics: true,
    compliance_log_level: 'INFO',
    jaeger_endpoint: process.env.JAEGER_ENDPOINT_BR || 'http://jaeger-collector-br:14268/api/traces',
    sector_specific: {
      HEALTHCARE: {
        lgpd_healthcare_metrics: true,
        phi_access_tracking: true,
        consent_verification_metrics: true
      }
    }
  },
  AO: {
    metrics_interval_ms: 30000,
    trace_sampling_ratio: 0.1,
    detailed_compliance_metrics: false,
    compliance_log_level: 'WARN',
    jaeger_endpoint: process.env.JAEGER_ENDPOINT_AO || 'http://jaeger-collector-ao:14268/api/traces',
    sector_specific: {
      HEALTHCARE: {
        pndsb_healthcare_metrics: true,
        phi_access_tracking: false,
        consent_verification_metrics: false
      }
    }
  },
  US: {
    metrics_interval_ms: 15000,
    trace_sampling_ratio: 0.5,
    detailed_compliance_metrics: true,
    compliance_log_level: 'INFO',
    jaeger_endpoint: process.env.JAEGER_ENDPOINT_US || 'http://jaeger-collector-us:14268/api/traces',
    sector_specific: {
      HEALTHCARE: {
        hipaa_metrics: true,
        phi_access_tracking: true,
        consent_verification_metrics: false
      }
    }
  }
};

/**
 * Classe para monitoramento e observabilidade da integração Kafka
 */
class KafkaMonitoring {
  /**
   * Inicializa o sistema de monitoramento Kafka
   * @param {Object} options - Opções de configuração
   * @param {string} options.regionCode - Código da região (EU, BR, AO, US)
   * @param {string} options.serviceName - Nome do serviço para identificação
   * @param {boolean} options.enableMetrics - Habilitar coleta de métricas
   * @param {boolean} options.enableTracing - Habilitar tracing distribuído
   * @param {Object} options.sectorConfig - Configurações específicas do setor (opcional)
   * @param {Object} options.customLabels - Labels personalizados para métricas
   */
  constructor(options = {}) {
    this.regionCode = options.regionCode || process.env.REGION_CODE || 'EU';
    this.serviceName = options.serviceName || 'iam-kafka-service';
    this.enableMetrics = options.enableMetrics !== false;
    this.enableTracing = options.enableTracing !== false;
    this.sectorConfig = options.sectorConfig || null;
    this.customLabels = options.customLabels || {};
    
    // Carregar configurações regionais
    this.config = REGIONAL_CONFIGS[this.regionCode] || REGIONAL_CONFIGS.EU;
    
    // Se houver configuração específica do setor, aplicar
    if (this.sectorConfig && this.config.sector_specific && 
        this.config.sector_specific[this.sectorConfig.sector]) {
      this.sectorMetrics = this.config.sector_specific[this.sectorConfig.sector];
    }
    
    // Inicializar observabilidade
    if (this.enableMetrics) {
      this.initializeMetrics();
    }
    
    if (this.enableTracing) {
      this.initializeTracing();
    }
    
    logger.info(`KafkaMonitoring inicializado para a região ${this.regionCode}`, {
      service: this.serviceName,
      sector: this.sectorConfig?.sector || 'GENERIC'
    });
  }
  
  /**
   * Inicializa o provedor de métricas e exportadores
   * @private
   */
  initializeMetrics() {
    // Configurar resource com atributos do serviço
    const resource = new Resource({
      [SemanticResourceAttributes.SERVICE_NAME]: this.serviceName,
      [SemanticResourceAttributes.SERVICE_VERSION]: process.env.SERVICE_VERSION || '1.0.0',
      [SemanticResourceAttributes.DEPLOYMENT_ENVIRONMENT]: process.env.NODE_ENV || 'development',
      'region.code': this.regionCode,
      ...(this.sectorConfig ? { 'sector': this.sectorConfig.sector } : {})
    });
    
    // Configurar exporter Prometheus
    const exporter = new PrometheusExporter({
      endpoint: process.env.PROMETHEUS_ENDPOINT || '/metrics',
      port: parseInt(process.env.PROMETHEUS_PORT || '9464', 10),
      startServer: true
    });
    
    // Criar provedor de métricas
    this.meterProvider = new MeterProvider({
      resource,
      exporter,
      interval: this.config.metrics_interval_ms
    });
    
    // Criar medidores para diferentes domínios
    this.meter = this.meterProvider.getMeter('iam-kafka-metrics');
    
    // Inicializar contadores e medidores
    this.initializeCounters();
    this.initializeHistograms();
    
    // Inicializar métricas específicas de setor, se aplicável
    if (this.sectorConfig?.sector === 'HEALTHCARE') {
      this.initializeHealthcareMetrics();
    }
    
    logger.debug('Sistema de métricas Kafka inicializado', {
      endpoint: exporter.endpoint,
      port: exporter.port
    });
  }
  
  /**
   * Inicializa os contadores para métricas Kafka
   * @private
   */
  initializeCounters() {
    // Contadores básicos de eventos Kafka
    this.eventCounter = this.meter.createCounter('iam.kafka.events.total', {
      description: 'Total de eventos Kafka processados',
      unit: '1'
    });
    
    this.eventsByTypeCounter = this.meter.createCounter('iam.kafka.events.by_type', {
      description: 'Total de eventos Kafka por tipo',
      unit: '1'
    });
    
    this.eventErrorCounter = this.meter.createCounter('iam.kafka.events.errors', {
      description: 'Erros no processamento de eventos Kafka',
      unit: '1'
    });
    
    // Contadores de operações de autenticação
    this.authCounter = this.meter.createCounter('iam.auth.operations', {
      description: 'Operações de autenticação realizadas',
      unit: '1'
    });
    
    this.mfaCounter = this.meter.createCounter('iam.auth.mfa', {
      description: 'Operações de MFA realizadas',
      unit: '1'
    });
    
    // Contadores de conformidade
    if (this.config.detailed_compliance_metrics) {
      this.complianceIssueCounter = this.meter.createCounter('iam.compliance.issues', {
        description: 'Problemas de conformidade detectados',
        unit: '1'
      });
      
      this.dataPrivacyCounter = this.meter.createCounter('iam.data.privacy_operations', {
        description: 'Operações de privacidade de dados realizadas',
        unit: '1'
      });
    }
  }
  
  /**
   * Inicializa os histogramas para métricas Kafka
   * @private
   */
  initializeHistograms() {
    // Histograma de latência de processamento de eventos
    this.eventProcessingTime = this.meter.createHistogram('iam.kafka.processing_time', {
      description: 'Tempo de processamento de eventos Kafka em ms',
      unit: 'ms',
      boundaries: [5, 10, 25, 50, 100, 250, 500, 1000]
    });
    
    // Histograma de tamanho de eventos
    this.eventSizeHistogram = this.meter.createHistogram('iam.kafka.event_size', {
      description: 'Tamanho dos eventos Kafka em bytes',
      unit: 'bytes',
      boundaries: [100, 500, 1000, 5000, 10000, 50000, 100000]
    });
  }
  
  /**
   * Inicializa métricas específicas para o setor de saúde
   * @private
   */
  initializeHealthcareMetrics() {
    // Métricas comuns para saúde
    this.phiAccessCounter = this.meter.createCounter('healthcare.phi.access', {
      description: 'Acessos a informações de saúde protegidas (PHI)',
      unit: '1'
    });
    
    // Métricas específicas por região/regulamentação
    if (this.sectorMetrics.hipaa_metrics) {
      this.hipaaComplianceCounter = this.meter.createCounter('healthcare.hipaa.compliance', {
        description: 'Métricas de conformidade HIPAA',
        unit: '1'
      });
    }
    
    if (this.sectorMetrics.gdpr_healthcare_metrics) {
      this.gdprHealthcareCounter = this.meter.createCounter('healthcare.gdpr.compliance', {
        description: 'Métricas de conformidade GDPR para saúde',
        unit: '1'
      });
      
      this.consentVerificationCounter = this.meter.createCounter('healthcare.consent.verification', {
        description: 'Verificações de consentimento para dados de saúde',
        unit: '1'
      });
    }
    
    if (this.sectorMetrics.lgpd_healthcare_metrics) {
      this.lgpdHealthcareCounter = this.meter.createCounter('healthcare.lgpd.compliance', {
        description: 'Métricas de conformidade LGPD para saúde',
        unit: '1'
      });
    }
    
    logger.debug('Métricas específicas de saúde inicializadas', {
      sector: 'HEALTHCARE',
      hipaa: !!this.sectorMetrics.hipaa_metrics,
      gdpr: !!this.sectorMetrics.gdpr_healthcare_metrics,
      lgpd: !!this.sectorMetrics.lgpd_healthcare_metrics
    });
  }
  
  /**
   * Inicializa o sistema de tracing distribuído
   * @private
   */
  initializeTracing() {
    // Configurar resource com atributos do serviço
    const resource = new Resource({
      [SemanticResourceAttributes.SERVICE_NAME]: this.serviceName,
      [SemanticResourceAttributes.SERVICE_VERSION]: process.env.SERVICE_VERSION || '1.0.0',
      [SemanticResourceAttributes.DEPLOYMENT_ENVIRONMENT]: process.env.NODE_ENV || 'development',
      'region.code': this.regionCode,
      ...(this.sectorConfig ? { 'sector': this.sectorConfig.sector } : {})
    });
    
    // Configurar exporter Jaeger
    const jaegerExporter = new JaegerExporter({
      endpoint: this.config.jaeger_endpoint,
      maxPacketSize: 65000
    });
    
    // Criar provedor de tracer
    this.tracerProvider = new NodeTracerProvider({
      resource,
      sampler: {
        shouldSample: () => {
          // Aplicar amostragem baseada na configuração regional
          return Math.random() < this.config.trace_sampling_ratio;
        }
      }
    });
    
    // Adicionar processador de spans
    this.tracerProvider.addSpanProcessor(
      new SimpleSpanProcessor(jaegerExporter)
    );
    
    // Registrar provedor globalmente
    this.tracerProvider.register();
    
    // Criar tracer para este componente
    this.tracer = trace.getTracer('iam-kafka-tracer');
    
    logger.debug('Sistema de tracing Kafka inicializado', {
      jaegerEndpoint: this.config.jaeger_endpoint.replace(/\/api\/traces$/, '')
    });
  }
  
  /**
   * Registra um evento Kafka processado
   * @param {Object} event - Evento Kafka
   * @param {Object} metadata - Metadados adicionais
   */
  recordEventProcessed(event, metadata = {}) {
    if (!this.enableMetrics) return;
    
    const labels = {
      topic: metadata.topic || 'unknown',
      event_type: event.event_type || 'unknown',
      region: this.regionCode,
      tenant: event.tenant_id || 'unknown',
      ...this.customLabels
    };
    
    // Incrementar contador geral
    this.eventCounter.add(1, labels);
    
    // Incrementar contador por tipo
    this.eventsByTypeCounter.add(1, {
      ...labels,
      event_type: event.event_type || 'unknown'
    });
    
    // Registrar tamanho do evento
    const eventSize = this.calculateEventSize(event);
    this.eventSizeHistogram.record(eventSize, labels);
    
    // Se for evento de autenticação, registrar métricas específicas
    if (event.event_type && event.event_type.startsWith('LOGIN_')) {
      this.recordAuthEvent(event, metadata);
    }
    
    // Se for evento de MFA, registrar métricas específicas
    if (event.event_type && event.event_type.startsWith('MFA_')) {
      this.recordMfaEvent(event, metadata);
    }
    
    // Registrar métricas específicas de saúde, se aplicável
    if (this.sectorConfig?.sector === 'HEALTHCARE') {
      this.recordHealthcareEvent(event, metadata);
    }
  }
  
  /**
   * Registra um evento de autenticação
   * @private
   * @param {Object} event - Evento de autenticação
   * @param {Object} metadata - Metadados adicionais
   */
  recordAuthEvent(event, metadata = {}) {
    const labels = {
      method: event.method_code || 'unknown',
      status: event.status || 'unknown',
      region: this.regionCode,
      tenant: event.tenant_id || 'unknown',
      success: event.event_type === 'LOGIN_SUCCESS'
    };
    
    this.authCounter.add(1, labels);
    
    // Registrar métricas de conformidade, se habilitado
    if (this.config.detailed_compliance_metrics && event.security_flags) {
      for (const flag of event.security_flags) {
        this.complianceIssueCounter.add(1, {
          ...labels,
          issue_type: flag
        });
      }
    }
  }
  
  /**
   * Registra um evento de MFA
   * @private
   * @param {Object} event - Evento de MFA
   * @param {Object} metadata - Metadados adicionais
   */
  recordMfaEvent(event, metadata = {}) {
    const labels = {
      method: event.method_code || 'unknown',
      method_category: event.method_category || 'unknown',
      status: event.status || 'unknown',
      region: this.regionCode,
      tenant: event.tenant_id || 'unknown',
      success: event.event_type === 'MFA_CHALLENGE_VERIFIED'
    };
    
    this.mfaCounter.add(1, labels);
  }
  
  /**
   * Registra um evento específico de saúde
   * @private
   * @param {Object} event - Evento relacionado a saúde
   * @param {Object} metadata - Metadados adicionais
   */
  recordHealthcareEvent(event, metadata = {}) {
    // Verificar acesso a PHI
    const phiAccess = event.additional_context?.phi_access === true || 
                      event.healthcare_metadata?.phi_access === true;
    
    if (phiAccess && this.sectorMetrics.phi_access_tracking) {
      this.phiAccessCounter.add(1, {
        region: this.regionCode,
        tenant: event.tenant_id || 'unknown',
        event_type: event.event_type || 'unknown',
        auth_method: event.method_code || 'unknown'
      });
    }
    
    // Métricas específicas por regulamentação
    if (this.sectorMetrics.hipaa_metrics) {
      this.recordHipaaMetrics(event, metadata);
    }
    
    if (this.sectorMetrics.gdpr_healthcare_metrics) {
      this.recordGdprHealthcareMetrics(event, metadata);
    }
    
    if (this.sectorMetrics.lgpd_healthcare_metrics) {
      this.recordLgpdHealthcareMetrics(event, metadata);
    }
  }
  
  /**
   * Registra métricas específicas HIPAA
   * @private
   * @param {Object} event - Evento relacionado a saúde
   * @param {Object} metadata - Metadados adicionais
   */
  recordHipaaMetrics(event, metadata = {}) {
    // Verificar aspectos específicos HIPAA
    const accessType = event.additional_context?.access_type || 'view';
    const department = event.additional_context?.department || 'unknown';
    
    this.hipaaComplianceCounter.add(1, {
      region: this.regionCode,
      tenant: event.tenant_id || 'unknown',
      event_type: event.event_type || 'unknown',
      access_type: accessType,
      department: department,
      mfa_compliant: this.isMfaCompliant(event.method_code) ? 'yes' : 'no'
    });
  }
  
  /**
   * Registra métricas específicas GDPR para saúde
   * @private
   * @param {Object} event - Evento relacionado a saúde
   * @param {Object} metadata - Metadados adicionais
   */
  recordGdprHealthcareMetrics(event, metadata = {}) {
    // Verificar aspectos específicos GDPR
    const consentVerified = event.additional_context?.consent_verified === true;
    const dataMasked = event.data_masked === true;
    const dataMinimized = event.additional_context?.data_minimized === true;
    
    this.gdprHealthcareCounter.add(1, {
      region: this.regionCode,
      tenant: event.tenant_id || 'unknown',
      event_type: event.event_type || 'unknown',
      data_masked: dataMasked ? 'yes' : 'no',
      data_minimized: dataMinimized ? 'yes' : 'no'
    });
    
    // Métricas de verificação de consentimento
    if (this.sectorMetrics.consent_verification_metrics) {
      this.consentVerificationCounter.add(1, {
        region: this.regionCode,
        tenant: event.tenant_id || 'unknown',
        verified: consentVerified ? 'yes' : 'no',
        event_type: event.event_type || 'unknown'
      });
    }
  }
  
  /**
   * Registra métricas específicas LGPD para saúde
   * @private
   * @param {Object} event - Evento relacionado a saúde
   * @param {Object} metadata - Metadados adicionais
   */
  recordLgpdHealthcareMetrics(event, metadata = {}) {
    // Verificar aspectos específicos LGPD
    const hasConsentId = !!event.additional_context?.consent_id;
    const dataMasked = event.data_masked === true;
    
    this.lgpdHealthcareCounter.add(1, {
      region: this.regionCode,
      tenant: event.tenant_id || 'unknown',
      event_type: event.event_type || 'unknown',
      has_consent: hasConsentId ? 'yes' : 'no',
      data_masked: dataMasked ? 'yes' : 'no'
    });
  }
  
  /**
   * Registra um erro no processamento de evento
   * @param {Object} event - Evento que gerou erro
   * @param {Error} error - Erro ocorrido
   * @param {Object} metadata - Metadados adicionais
   */
  recordEventProcessingError(event, error, metadata = {}) {
    if (!this.enableMetrics) return;
    
    const errorType = error.name || 'UnknownError';
    const errorCode = error.code || 'UNKNOWN';
    
    const labels = {
      topic: metadata.topic || 'unknown',
      event_type: event?.event_type || 'unknown',
      error_type: errorType,
      error_code: errorCode,
      region: this.regionCode,
      tenant: event?.tenant_id || 'unknown'
    };
    
    this.eventErrorCounter.add(1, labels);
    
    // Registrar no logger
    logger.error(`Erro no processamento de evento Kafka: ${error.message}`, {
      error: {
        name: errorType,
        code: errorCode,
        message: error.message,
        stack: error.stack
      },
      event: {
        id: event?.event_id,
        type: event?.event_type,
        tenant: event?.tenant_id
      },
      metadata
    });
  }
  
  /**
   * Registra o tempo de processamento de um evento
   * @param {Object} event - Evento processado
   * @param {number} processingTimeMs - Tempo de processamento em ms
   * @param {Object} metadata - Metadados adicionais
   */
  recordProcessingTime(event, processingTimeMs, metadata = {}) {
    if (!this.enableMetrics) return;
    
    const labels = {
      topic: metadata.topic || 'unknown',
      event_type: event.event_type || 'unknown',
      region: this.regionCode,
      tenant: event.tenant_id || 'unknown'
    };
    
    this.eventProcessingTime.record(processingTimeMs, labels);
  }
  
  /**
   * Cria um span de tracing para o processamento de um evento
   * @param {Object} event - Evento a ser processado
   * @param {Object} metadata - Metadados adicionais
   * @returns {Object} Objeto de span para tracing
   */
  createEventProcessingSpan(event, metadata = {}) {
    if (!this.enableTracing) return null;
    
    const spanName = `process_${metadata.topic || 'kafka'}_event`;
    
    return this.tracer.startSpan(spanName, {
      attributes: {
        'messaging.system': 'kafka',
        'messaging.destination': metadata.topic || 'unknown',
        'messaging.destination_kind': 'topic',
        'messaging.operation': 'process',
        'event.id': event.event_id || 'unknown',
        'event.type': event.event_type || 'unknown',
        'tenant.id': event.tenant_id || 'unknown',
        'region.code': this.regionCode
      }
    });
  }
  
  /**
   * Verifica se um método MFA é compatível com requisitos específicos
   * @private
   * @param {string} methodCode - Código do método MFA
   * @returns {boolean} Se o método é compatível
   */
  isMfaCompliant(methodCode) {
    if (!methodCode) return false;
    
    // Métodos MFA considerados seguros para saúde
    const strongMFAMethods = ['K05', 'K06', 'K07', 'K10', 'K12', 'K13'];
    return strongMFAMethods.includes(methodCode);
  }
  
  /**
   * Calcula o tamanho aproximado de um evento em bytes
   * @private
   * @param {Object} event - Evento a ser medido
   * @returns {number} Tamanho aproximado em bytes
   */
  calculateEventSize(event) {
    if (!event) return 0;
    try {
      return Buffer.byteLength(JSON.stringify(event), 'utf8');
    } catch (e) {
      return 0;
    }
  }
  
  /**
   * Registra uma operação de mascaramento de dados
   * @param {Object} event - Evento relacionado
   * @param {Array<string>} fields - Campos mascarados
   * @param {Object} metadata - Metadados adicionais
   */
  recordDataMaskingOperation(event, fields = [], metadata = {}) {
    if (!this.enableMetrics || !this.config.detailed_compliance_metrics) return;
    
    const labels = {
      operation: 'masking',
      region: this.regionCode,
      tenant: event.tenant_id || 'unknown',
      event_type: event.event_type || 'unknown',
      field_count: fields.length.toString()
    };
    
    this.dataPrivacyCounter.add(1, labels);
  }
  
  /**
   * Registra uma operação de validação de conformidade
   * @param {Object} validationResult - Resultado da validação
   * @param {Object} event - Evento validado
   * @param {Object} metadata - Metadados adicionais
   */
  recordComplianceValidation(validationResult, event, metadata = {}) {
    if (!this.enableMetrics || !this.config.detailed_compliance_metrics) return;
    
    const regulationType = metadata.regulation || 'generic';
    const validationType = metadata.validationType || 'standard';
    
    const labels = {
      region: this.regionCode,
      tenant: event.tenant_id || 'unknown',
      regulation: regulationType,
      validation_type: validationType,
      result: validationResult.valid ? 'compliant' : 'non_compliant',
      severity: validationResult.severity || 'info'
    };
    
    this.complianceIssueCounter.add(1, labels);
    
    // Registrar no logger se não estiver em conformidade
    if (!validationResult.valid) {
      const logLevel = this.config.compliance_log_level.toLowerCase();
      logger[logLevel](`Problema de conformidade detectado: ${validationResult.reason}`, {
        tenant: event.tenant_id,
        eventId: event.event_id,
        eventType: event.event_type,
        regulation: regulationType,
        code: validationResult.code,
        severity: validationResult.severity
      });
    }
  }
  
  /**
   * Exporta a configuração atual de monitoramento
   * @returns {Object} Configuração de monitoramento
   */
  exportConfig() {
    return {
      regionCode: this.regionCode,
      serviceName: this.serviceName,
      metricsEnabled: this.enableMetrics,
      tracingEnabled: this.enableTracing,
      sectorSpecific: this.sectorConfig?.sector,
      regionalSettings: {
        metricsInterval: this.config.metrics_interval_ms,
        traceSamplingRatio: this.config.trace_sampling_ratio,
        detailedComplianceMetrics: this.config.detailed_compliance_metrics
      },
      endpoints: {
        prometheus: process.env.PROMETHEUS_ENDPOINT,
        jaeger: this.config.jaeger_endpoint?.replace(/\/api\/traces$/, '')
      }
    };
  }
  
  /**
   * Finaliza o sistema de monitoramento Kafka
   * @returns {Promise<void>}
   */
  async shutdown() {
    logger.info('Finalizando sistema de monitoramento Kafka');
    
    const shutdownPromises = [];
    
    if (this.tracerProvider) {
      shutdownPromises.push(this.tracerProvider.shutdown());
    }
    
    if (this.meterProvider) {
      shutdownPromises.push(this.meterProvider.shutdown());
    }
    
    try {
      await Promise.all(shutdownPromises);
      logger.info('Sistema de monitoramento Kafka finalizado com sucesso');
    } catch (error) {
      logger.error(`Erro ao finalizar sistema de monitoramento: ${error.message}`, { error });
      throw error;
    }
  }
}

module.exports = KafkaMonitoring;
