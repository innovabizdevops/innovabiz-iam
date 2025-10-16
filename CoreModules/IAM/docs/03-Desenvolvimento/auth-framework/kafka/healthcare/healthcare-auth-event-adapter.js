/**
 * @fileoverview Adaptador especializado para eventos de autenticação do setor de saúde
 * @author INNOVABIZ Dev Team
 * @version 1.0.0
 * @module iam/auth-framework/kafka/healthcare/healthcare-auth-event-adapter
 * @description Responsável por adaptar eventos de autenticação para o contexto específico de saúde
 */

const { AuthEventProducer } = require('../auth-event-producer');
const { AuthEventConsumer } = require('../auth-event-consumer');
const { MCPKafkaAdapter } = require('../mcp/mcp-kafka-adapter');
const logger = require('../../../core/utils/logging');
const { maskSensitiveData } = require('../../../core/utils/data-masking');
const { EventObservability } = require('../../../core/utils/observability');

// Importar validadores específicos de saúde
const { 
  HIPAAHealthcareValidator,
  GDPRHealthcareValidator,
  LGPDHealthcareValidator
} = require('../../../integrations/healthcare/validators');

// Configurações por região para o setor de saúde
const HEALTHCARE_CONFIGS = {
  EU: {
    topic: 'iam-healthcare-auth-events',
    data_masking: true,
    phi_fields: ['username', 'ip_address', 'email', 'device_id', 'location', 'user_agent'],
    validators: ['GDPRHealthcareValidator'],
    compliance: {
      gdpr_healthcare: true,
      data_minimization: true,
      consent_tracking: true,
      audit_retention_days: 365,
      data_residency: 'EU',
      cross_border_transfer: false
    },
    mcp_channel: 'healthcare.auth.eu'
  },
  BR: {
    topic: 'iam-healthcare-auth-events',
    data_masking: true,
    phi_fields: ['username', 'ip_address', 'email', 'device_id', 'location'],
    validators: ['LGPDHealthcareValidator'],
    compliance: {
      lgpd_healthcare: true,
      data_minimization: true,
      consent_tracking: true,
      audit_retention_days: 365,
      data_residency: 'BR',
      cross_border_transfer: false
    },
    mcp_channel: 'healthcare.auth.br'
  },
  AO: {
    topic: 'iam-healthcare-auth-events',
    data_masking: false,
    phi_fields: ['username', 'email'],
    validators: [],
    compliance: {
      pndsb_healthcare: true,
      data_minimization: false,
      consent_tracking: false,
      audit_retention_days: 180,
      data_residency: 'AO',
      cross_border_transfer: false
    },
    mcp_channel: 'healthcare.auth.ao'
  },
  US: {
    topic: 'iam-healthcare-auth-events',
    data_masking: true,
    phi_fields: ['username', 'ip_address', 'email', 'device_id', 'location', 'user_agent'],
    validators: ['HIPAAHealthcareValidator'],
    compliance: {
      hipaa: true,
      data_minimization: true,
      consent_tracking: true,
      audit_retention_days: 730,
      data_residency: 'US',
      cross_border_transfer: true
    },
    mcp_channel: 'healthcare.auth.us'
  }
};

/**
 * Classe para adaptar eventos de autenticação específicos do setor de saúde
 */
class HealthcareAuthEventAdapter {
  /**
   * Inicializa o adaptador de eventos de saúde
   * @param {Object} options - Opções de configuração
   * @param {string} options.regionCode - Código da região (EU, BR, AO, US)
   * @param {Object} options.kafkaConfig - Configurações do Kafka
   * @param {Object} options.mcpConfig - Configurações do MCP
   * @param {boolean} options.enableComplianceValidation - Habilitar validação de conformidade
   * @param {boolean} options.enableObservability - Habilitar métricas de observabilidade
   */
  constructor(options = {}) {
    this.regionCode = options.regionCode || process.env.REGION_CODE || 'EU';
    this.kafkaConfig = options.kafkaConfig || {};
    this.mcpConfig = options.mcpConfig || {};
    this.enableComplianceValidation = options.enableComplianceValidation !== false;
    this.enableObservability = options.enableObservability !== false;
    
    // Carregar configurações específicas de saúde para a região
    this.healthcareConfig = HEALTHCARE_CONFIGS[this.regionCode] || HEALTHCARE_CONFIGS.US;
    
    // Inicializar componentes
    this.initializeComponents();
    
    // Inicializar validadores de conformidade se habilitado
    if (this.enableComplianceValidation) {
      this.initializeValidators();
    }
    
    // Inicializar observabilidade se habilitado
    if (this.enableObservability) {
      this.observability = new EventObservability({
        serviceName: 'healthcare-auth-adapter',
        region: this.regionCode,
        sectorSpecific: {
          sector: 'HEALTHCARE',
          hipaaCompliant: this.healthcareConfig.compliance.hipaa || false,
          gdprHealthcareCompliant: this.healthcareConfig.compliance.gdpr_healthcare || false,
          lgpdHealthcareCompliant: this.healthcareConfig.compliance.lgpd_healthcare || false
        }
      });
    }
    
    logger.info(`HealthcareAuthEventAdapter inicializado para a região ${this.regionCode}`);
  }
  
  /**
   * Inicializa componentes Kafka e MCP
   * @private
   */
  initializeComponents() {
    // Inicializar produtor de eventos de saúde
    this.producer = new AuthEventProducer({
      regionCode: this.regionCode,
      kafkaConfig: this.kafkaConfig,
      sectorConfig: {
        sector: 'HEALTHCARE'
      }
    });
    
    // Inicializar consumidor de eventos de saúde
    this.consumer = new AuthEventConsumer({
      regionCode: this.regionCode,
      groupId: `healthcare-auth-consumer-${this.regionCode.toLowerCase()}`,
      kafkaConfig: this.kafkaConfig,
      sectorConfig: {
        sector: 'HEALTHCARE'
      },
      eventHandlers: this.getCustomEventHandlers()
    });
    
    // Inicializar adaptador MCP para eventos de saúde
    this.mcpAdapter = new MCPKafkaAdapter({
      regionCode: this.regionCode,
      mcpConfig: this.mcpConfig,
      sectorConfig: {
        sector: 'HEALTHCARE'
      },
      contextEnrichment: true
    });
  }
  
  /**
   * Inicializa validadores de conformidade para o setor de saúde
   * @private
   */
  initializeValidators() {
    this.validators = [];
    
    // Carrega os validadores baseados na configuração regional
    if (this.healthcareConfig.validators && this.healthcareConfig.validators.length > 0) {
      for (const validatorName of this.healthcareConfig.validators) {
        let validator;
        
        switch (validatorName) {
          case 'HIPAAHealthcareValidator':
            validator = new HIPAAHealthcareValidator({
              region: this.regionCode,
              strictMode: true,
              auditRetentionDays: this.healthcareConfig.compliance.audit_retention_days
            });
            break;
            
          case 'GDPRHealthcareValidator':
            validator = new GDPRHealthcareValidator({
              region: this.regionCode,
              strictMode: true,
              dataMinimization: this.healthcareConfig.compliance.data_minimization,
              consentTracking: this.healthcareConfig.compliance.consent_tracking,
              crossBorderTransfer: this.healthcareConfig.compliance.cross_border_transfer
            });
            break;
            
          case 'LGPDHealthcareValidator':
            validator = new LGPDHealthcareValidator({
              region: this.regionCode,
              strictMode: true,
              dataMinimization: this.healthcareConfig.compliance.data_minimization,
              consentTracking: this.healthcareConfig.compliance.consent_tracking
            });
            break;
            
          default:
            logger.warn(`Validador desconhecido: ${validatorName}`);
            continue;
        }
        
        if (validator) {
          this.validators.push(validator);
          logger.info(`Validador ${validatorName} inicializado para a região ${this.regionCode}`);
        }
      }
    }
  }
  
  /**
   * Obtém manipuladores personalizados para eventos do setor de saúde
   * @private
   * @returns {Object} Manipuladores de eventos
   */
  getCustomEventHandlers() {
    return {
      // Manipulador para tentativas de login em sistemas de saúde
      LOGIN_ATTEMPT: async (event) => {
        logger.debug(`Handler para LOGIN_ATTEMPT em sistema de saúde: ${event.event_id}`);
        
        // Registrar métrica de observabilidade
        if (this.observability) {
          this.observability.recordHealthcareAuthAttempt({
            success: false,
            pending: true,
            tenantId: event.tenant_id,
            regionCode: event.region_code
          });
        }
        
        // Validar conformidade
        if (this.enableComplianceValidation && this.validators.length > 0) {
          for (const validator of this.validators) {
            const validationResult = validator.validate(event);
            if (!validationResult.valid) {
              logger.warn(`Evento falhou na validação ${validator.constructor.name}: ${validationResult.reason}`);
            }
          }
        }
        
        return { processed: true, action: 'log' };
      },
      
      // Manipulador para login bem-sucedido em sistemas de saúde
      LOGIN_SUCCESS: async (event) => {
        logger.debug(`Handler para LOGIN_SUCCESS em sistema de saúde: ${event.event_id}`);
        
        // Registrar métrica de observabilidade
        if (this.observability) {
          this.observability.recordHealthcareAuthAttempt({
            success: true,
            pending: false,
            tenantId: event.tenant_id,
            regionCode: event.region_code,
            authMethod: event.method_code
          });
        }
        
        // Publicar evento no MCP para integração com sistemas de saúde
        if (this.mcpAdapter && await this.mcpAdapter.isConnected()) {
          await this.mcpAdapter.publishToMCP(
            this.healthcareConfig.topic,
            event,
            { healthcare_specific: true }
          );
        }
        
        // Gerar relatório de compliance se necessário
        if (event.additional_context?.generate_compliance_report) {
          // Lógica para gerar relatório de compliance
          logger.info(`Gerando relatório de compliance para login em sistema de saúde: ${event.event_id}`);
          
          // Essa funcionalidade seria implementada em outro componente
          // await this.generateComplianceReport(event);
        }
        
        return { processed: true, action: 'integrate' };
      },
      
      // Manipulador para falha de login em sistemas de saúde
      LOGIN_FAILURE: async (event) => {
        logger.debug(`Handler para LOGIN_FAILURE em sistema de saúde: ${event.event_id}`);
        
        // Registrar métrica de observabilidade
        if (this.observability) {
          this.observability.recordHealthcareAuthAttempt({
            success: false,
            pending: false,
            tenantId: event.tenant_id,
            regionCode: event.region_code,
            failure: true,
            failureReason: event.error_info?.error_code || 'UNKNOWN'
          });
        }
        
        // Verificar tentativas consecutivas de falha
        // Essa lógica seria implementada com Redis ou similar
        
        return { processed: true, action: 'alert' };
      },
      
      // Manipulador para desafio MFA em sistemas de saúde
      MFA_CHALLENGE_ISSUED: async (event) => {
        logger.debug(`Handler para MFA_CHALLENGE_ISSUED em sistema de saúde: ${event.event_id}`);
        
        // Validação específica para MFA em saúde
        if (this.enableComplianceValidation && this.validators.length > 0) {
          // Verificar se o método MFA é forte o suficiente para acesso a dados de saúde
          const isMFAStrongEnough = this.isHealthcareMFACompliant(event.method_code);
          
          if (!isMFAStrongEnough) {
            logger.warn(`Método MFA ${event.method_code} não é forte o suficiente para acesso a dados de saúde`);
            
            // Registrar não-conformidade
            if (this.observability) {
              this.observability.recordComplianceIssue({
                issueType: 'WEAK_MFA',
                tenantId: event.tenant_id,
                regionCode: event.region_code,
                severity: 'HIGH'
              });
            }
          }
        }
        
        return { processed: true, action: 'none' };
      }
    };
  }
  
  /**
   * Verifica se um método MFA é compatível com requisitos de saúde
   * @private
   * @param {string} methodCode - Código do método MFA
   * @returns {boolean} Se o método é compatível
   */
  isHealthcareMFACompliant(methodCode) {
    // Métodos MFA considerados seguros para sistemas de saúde
    const strongMFAMethods = ['K05', 'K06', 'K07', 'K10', 'K12', 'K13'];
    return strongMFAMethods.includes(methodCode);
  }
  
  /**
   * Processa e adapta um evento de autenticação para o contexto de saúde
   * @param {Object} event - Evento de autenticação
   * @returns {Object} Evento processado
   */
  processHealthcareAuthEvent(event) {
    // Adicionar campos específicos de saúde
    const enhancedEvent = {
      ...event,
      sector: 'HEALTHCARE',
      healthcare_metadata: {
        phi_access: event.additional_context?.phi_access || false,
        requires_hipaa_logging: this.healthcareConfig.compliance.hipaa || false,
        requires_gdpr_healthcare: this.healthcareConfig.compliance.gdpr_healthcare || false,
        requires_lgpd_healthcare: this.healthcareConfig.compliance.lgpd_healthcare || false,
        data_residency: this.healthcareConfig.compliance.data_residency,
        consent_validation: this.healthcareConfig.compliance.consent_tracking
      }
    };
    
    // Aplicar mascaramento em campos PHI se configurado
    if (this.healthcareConfig.data_masking && this.healthcareConfig.phi_fields.length > 0) {
      return maskSensitiveData(enhancedEvent, this.healthcareConfig.phi_fields);
    }
    
    return enhancedEvent;
  }
  
  /**
   * Publica um evento de autenticação específico para saúde
   * @param {Object} event - Evento de autenticação
   * @param {Object} options - Opções adicionais
   * @returns {Promise<Object>} Resultado da operação
   */
  async publishHealthcareAuthEvent(event, options = {}) {
    if (!this.producer) {
      throw new Error('Produtor não inicializado');
    }
    
    // Processar evento para contexto de saúde
    const healthcareEvent = this.processHealthcareAuthEvent(event);
    
    // Validar conformidade
    if (this.enableComplianceValidation && this.validators.length > 0) {
      for (const validator of this.validators) {
        const validationResult = validator.validate(healthcareEvent);
        if (!validationResult.valid) {
          logger.warn(`Evento falhou na validação ${validator.constructor.name}: ${validationResult.reason}`);
          
          // Registrar não-conformidade
          if (this.observability) {
            this.observability.recordComplianceIssue({
              issueType: validationResult.code || 'VALIDATION_FAILED',
              tenantId: healthcareEvent.tenant_id,
              regionCode: healthcareEvent.region_code,
              severity: validationResult.severity || 'MEDIUM',
              details: validationResult.reason
            });
          }
          
          // Se a validação for estrita, falhar
          if (options.strictValidation) {
            throw new Error(`Falha na validação de conformidade: ${validationResult.reason}`);
          }
        }
      }
    }
    
    try {
      // Publicar evento no tópico específico de saúde
      const result = await this.producer.publishAuthEvent(
        healthcareEvent,
        { topic: this.healthcareConfig.topic }
      );
      
      // Publicar também no MCP se configurado
      if (this.mcpAdapter && await this.mcpAdapter.isConnected() && options.publishToMCP !== false) {
        await this.mcpAdapter.publishToMCP(
          this.healthcareConfig.topic,
          healthcareEvent,
          { healthcare_specific: true }
        );
      }
      
      // Registrar métrica de evento publicado
      if (this.observability) {
        this.observability.recordHealthcareEventPublished({
          eventType: healthcareEvent.event_type,
          tenantId: healthcareEvent.tenant_id,
          regionCode: healthcareEvent.region_code,
          topic: this.healthcareConfig.topic
        });
      }
      
      return result;
      
    } catch (error) {
      logger.error(`Erro ao publicar evento de autenticação para saúde: ${error.message}`, {
        error,
        eventId: healthcareEvent.event_id
      });
      
      throw error;
    }
  }
  
  /**
   * Configura um consumidor para eventos de autenticação de saúde
   * @param {Function} handler - Função para processar eventos
   * @param {Object} options - Opções adicionais
   * @returns {Promise<void>}
   */
  async setupHealthcareConsumer(handler, options = {}) {
    if (!this.consumer) {
      throw new Error('Consumidor não inicializado');
    }
    
    try {
      // Conectar consumidor
      await this.consumer.connect();
      
      // Inscrever no tópico de eventos de saúde
      await this.consumer.subscribe(
        [this.healthcareConfig.topic],
        { fromBeginning: options.fromBeginning || false }
      );
      
      // Iniciar consumo com handler personalizado para saúde
      await this.consumer.run({
        autoCommit: options.autoCommit !== undefined ? options.autoCommit : true,
        processHandler: async (event, headers) => {
          // Se handler personalizado foi fornecido, chamá-lo
          if (typeof handler === 'function') {
            return handler(event, headers);
          }
          
          // Caso contrário, usar handlers padrão
          const eventType = event.event_type;
          const eventHandlers = this.getCustomEventHandlers();
          
          if (eventHandlers[eventType]) {
            return eventHandlers[eventType](event, headers);
          }
          
          return { processed: true, action: 'none' };
        }
      });
      
      logger.info(`Consumidor para eventos de autenticação de saúde iniciado: ${this.healthcareConfig.topic}`);
      
    } catch (error) {
      logger.error(`Erro ao configurar consumidor para eventos de autenticação de saúde: ${error.message}`, {
        error,
        topic: this.healthcareConfig.topic
      });
      
      throw error;
    }
  }
  
  /**
   * Configura um assinante MCP para integração com sistemas de saúde
   * @param {Function} handler - Função para processar mensagens MCP
   * @param {Object} options - Opções adicionais
   * @returns {Promise<string>} ID da inscrição
   */
  async setupMCPSubscriber(handler, options = {}) {
    if (!this.mcpAdapter) {
      throw new Error('Adaptador MCP não inicializado');
    }
    
    try {
      // Conectar ao MCP
      await this.mcpAdapter.connect();
      
      // Inscrever no canal MCP específico de saúde
      const subscriptionId = await this.mcpAdapter.subscribeMCP(
        this.healthcareConfig.mcp_channel,
        async (mcpMessage, originalMessage) => {
          // Converter para formato de evento Kafka se necessário
          const event = options.convertToKafkaEvent ? 
            this.mcpAdapter.transformToKafkaEvent(mcpMessage) : 
            mcpMessage;
          
          // Aplicar processamento específico de saúde
          const healthcareEvent = this.processHealthcareAuthEvent(event);
          
          // Chamar handler com o evento processado
          if (typeof handler === 'function') {
            await handler(healthcareEvent, mcpMessage);
          }
        },
        options
      );
      
      logger.info(`Assinante MCP configurado para canal: ${this.healthcareConfig.mcp_channel}`, {
        subscriptionId
      });
      
      return subscriptionId;
      
    } catch (error) {
      logger.error(`Erro ao configurar assinante MCP: ${error.message}`, {
        error,
        channel: this.healthcareConfig.mcp_channel
      });
      
      throw error;
    }
  }
  
  /**
   * Fecha todas as conexões e recursos
   * @returns {Promise<void>}
   */
  async shutdown() {
    logger.info('Iniciando desligamento do adaptador de eventos de saúde');
    
    const shutdownPromises = [];
    
    if (this.consumer) {
      shutdownPromises.push(this.consumer.disconnect());
    }
    
    if (this.producer) {
      shutdownPromises.push(this.producer.disconnect());
    }
    
    if (this.mcpAdapter) {
      shutdownPromises.push(this.mcpAdapter.disconnect());
    }
    
    try {
      await Promise.all(shutdownPromises);
      logger.info('Adaptador de eventos de saúde desligado com sucesso');
    } catch (error) {
      logger.error(`Erro durante o desligamento: ${error.message}`, { error });
      throw error;
    }
  }
}

module.exports = HealthcareAuthEventAdapter;
