/**
 * @fileoverview Adaptador MCP (Model Context Protocol) para integração com Apache Kafka
 * @author INNOVABIZ Dev Team
 * @version 1.0.0
 * @module iam/auth-framework/kafka/mcp/mcp-kafka-adapter
 * @description Responsável pela integração entre eventos Kafka e o protocolo MCP da plataforma INNOVABIZ
 */

const { MCPClient, MCPMessage, MCPContext } = require('@innovabiz/mcp-client');
const uuid = require('uuid');
const logger = require('../../../core/utils/logging');
const { maskSensitiveData } = require('../../../core/utils/data-masking');
const { applyRegionalCompliance } = require('../../../core/utils/compliance');
const { HealthcareValidatorFactory } = require('../../../integrations/healthcare/validators');

// Configurações regionais para MCP
const REGIONAL_CONFIGS = {
  EU: {
    mcp_broker: process.env.MCP_BROKER_EU || 'amqp://innovabiz-rabbitmq-eu:5672',
    context_enrichment: true,
    compliance: {
      gdpr: true,
      eidas: true
    },
    sector_specific: {
      HEALTHCARE: {
        compliance: {
          gdpr_healthcare: true
        },
        validators: ['GDPRHealthcareValidator']
      }
    }
  },
  BR: {
    mcp_broker: process.env.MCP_BROKER_BR || 'amqp://innovabiz-rabbitmq-br:5672',
    context_enrichment: true,
    compliance: {
      lgpd: true,
      icp_brasil: true
    },
    sector_specific: {
      HEALTHCARE: {
        compliance: {
          lgpd_healthcare: true
        },
        validators: ['LGPDHealthcareValidator']
      }
    }
  },
  AO: {
    mcp_broker: process.env.MCP_BROKER_AO || 'amqp://innovabiz-rabbitmq-ao:5672',
    context_enrichment: true,
    compliance: {
      pndsb: true
    }
  },
  US: {
    mcp_broker: process.env.MCP_BROKER_US || 'amqp://innovabiz-rabbitmq-us:5672',
    context_enrichment: true,
    compliance: {
      nist: true,
      soc2: true
    },
    sector_specific: {
      HEALTHCARE: {
        compliance: {
          hipaa: true
        },
        validators: ['HIPAAHealthcareValidator']
      },
      FINANCE: {
        compliance: {
          pci_dss: true
        }
      }
    }
  }
};

/**
 * Mapeamento entre tópicos Kafka e canais MCP
 * @type {Object}
 */
const TOPIC_TO_MCP_CHANNEL = {
  'iam-auth-events': 'auth.events',
  'iam-token-events': 'auth.tokens',
  'iam-user-events': 'user.events',
  'iam-mfa-challenges': 'auth.mfa',
  'iam-audit-logs': 'audit.logs',
  'iam-security-alerts': 'security.alerts',
  'iam-notification-events': 'notifications',
  'iam-offline-auth-events': 'auth.offline',
  'iam-healthcare-auth-events': 'healthcare.auth'
};

/**
 * Classe responsável por adaptar eventos Kafka para o protocolo MCP
 */
class MCPKafkaAdapter {
  /**
   * Inicializa o adaptador MCP para Kafka
   * @param {Object} options - Opções de configuração
   * @param {string} options.regionCode - Código da região (EU, BR, AO, US)
   * @param {string} options.mcpConfig - Configurações do cliente MCP
   * @param {Object} options.sectorConfig - Configurações específicas do setor (opcional)
   * @param {boolean} options.contextEnrichment - Habilitar enriquecimento de contexto
   */
  constructor(options = {}) {
    this.regionCode = options.regionCode || process.env.REGION_CODE || 'EU';
    this.mcpConfig = options.mcpConfig || {};
    this.sectorConfig = options.sectorConfig || null;
    
    // Carregando configurações regionais
    this.regionalConfig = REGIONAL_CONFIGS[this.regionCode] || REGIONAL_CONFIGS.EU;
    
    // Se houver configuração específica do setor, aplicar
    if (this.sectorConfig && this.regionalConfig.sector_specific && 
        this.regionalConfig.sector_specific[this.sectorConfig.sector]) {
      this.regionalConfig = {
        ...this.regionalConfig,
        ...this.regionalConfig.sector_specific[this.sectorConfig.sector],
        compliance: {
          ...this.regionalConfig.compliance,
          ...this.regionalConfig.sector_specific[this.sectorConfig.sector].compliance
        }
      };
      
      // Inicializar validadores específicos do setor, se necessário
      if (this.sectorConfig.sector === 'HEALTHCARE' && 
          this.regionalConfig.sector_specific.HEALTHCARE.validators) {
        this.healthcareValidators = this.regionalConfig.sector_specific.HEALTHCARE.validators
          .map(validatorName => HealthcareValidatorFactory.create(validatorName, {
            region: this.regionCode
          }));
      }
    }
    
    // Configuração de enriquecimento de contexto
    this.contextEnrichment = options.contextEnrichment !== undefined ? 
      options.contextEnrichment : this.regionalConfig.context_enrichment;
    
    // Inicializar cliente MCP
    this.initMCPClient();
    
    logger.info(`MCPKafkaAdapter inicializado para a região ${this.regionCode}`);
  }
  
  /**
   * Inicializa o cliente MCP
   * @private
   */
  initMCPClient() {
    const clientOptions = {
      clientId: `iam-kafka-mcp-adapter-${this.regionCode.toLowerCase()}`,
      brokerUrl: this.regionalConfig.mcp_broker,
      reconnect: true,
      maxReconnectAttempts: 10,
      reconnectInterval: 5000,
      ...this.mcpConfig
    };
    
    this.mcpClient = new MCPClient(clientOptions);
    
    // Configurar handlers de eventos
    this.mcpClient.on('connected', () => {
      logger.info('Adaptador MCP conectado ao broker');
    });
    
    this.mcpClient.on('error', (error) => {
      logger.error(`Erro no cliente MCP: ${error.message}`, { error });
    });
    
    this.mcpClient.on('disconnected', () => {
      logger.warn('Adaptador MCP desconectado do broker');
    });
  }
  
  /**
   * Conecta ao broker MCP
   * @returns {Promise<void>}
   */
  async connect() {
    try {
      await this.mcpClient.connect();
      logger.info('Adaptador MCP conectado com sucesso');
    } catch (error) {
      logger.error(`Erro ao conectar ao broker MCP: ${error.message}`, { error });
      throw error;
    }
  }
  
  /**
   * Transforma um evento Kafka em uma mensagem MCP
   * @private
   * @param {string} topic - Tópico Kafka de origem
   * @param {Object} event - Evento Kafka
   * @param {Object} headers - Cabeçalhos Kafka
   * @returns {MCPMessage} Mensagem MCP
   */
  transformToMCPMessage(topic, event, headers = {}) {
    // Determinar canal MCP baseado no tópico Kafka
    const channel = TOPIC_TO_MCP_CHANNEL[topic] || 'auth.events';
    
    // Aplicar conformidade regional
    const compliantEvent = applyRegionalCompliance(event, {
      region: this.regionCode,
      compliance: this.regionalConfig.compliance
    });
    
    // Se for do setor de saúde, aplicar validadores específicos
    if (this.healthcareValidators && topic === 'iam-healthcare-auth-events') {
      for (const validator of this.healthcareValidators) {
        validator.validate(compliantEvent);
      }
    }
    
    // Criar contexto MCP
    const context = new MCPContext({
      requestId: event.correlation_id || headers['correlation-id'] || uuid.v4(),
      tenantId: event.tenant_id || headers['tenant-id'],
      userId: event.user_id,
      regionCode: this.regionCode,
      sourceSystem: 'iam-kafka',
      sourceTopic: topic,
      eventType: event.event_type,
      timestamp: event.timestamp || Date.now()
    });
    
    // Enriquecer contexto se habilitado
    if (this.contextEnrichment) {
      this.enrichContext(context, compliantEvent);
    }
    
    // Criar mensagem MCP
    const mcpMessage = new MCPMessage({
      channel,
      type: event.event_type,
      payload: compliantEvent,
      context,
      metadata: {
        originalTopic: topic,
        kafkaTimestamp: event.timestamp,
        kafkaOffset: headers.offset,
        kafkaPartition: headers.partition,
        regionCode: this.regionCode
      }
    });
    
    return mcpMessage;
  }
  
  /**
   * Enriquece o contexto MCP com informações adicionais
   * @private
   * @param {MCPContext} context - Contexto MCP a ser enriquecido
   * @param {Object} event - Evento processado
   */
  enrichContext(context, event) {
    // Adicionar informações específicas com base no tipo de evento
    switch (event.event_type) {
      case 'LOGIN_ATTEMPT':
      case 'LOGIN_SUCCESS':
      case 'LOGIN_FAILURE':
        context.setProperty('authMethod', event.method_code);
        context.setProperty('clientIp', event.client_info?.ip_address);
        context.setProperty('userAgent', event.client_info?.user_agent);
        context.setProperty('deviceId', event.client_info?.device_id);
        break;
        
      case 'TOKEN_ISSUED':
      case 'TOKEN_REFRESHED':
      case 'TOKEN_REVOKED':
        context.setProperty('tokenType', event.token_type);
        context.setProperty('sessionId', event.session_id);
        context.setProperty('tokenExpiry', event.token_expiry);
        break;
        
      case 'MFA_CHALLENGE_ISSUED':
      case 'MFA_CHALLENGE_VERIFIED':
      case 'MFA_CHALLENGE_FAILED':
        context.setProperty('challengeId', event.challenge_id);
        context.setProperty('methodCode', event.method_code);
        context.setProperty('methodCategory', event.method_category);
        context.setProperty('deliveryChannel', event.delivery_channel);
        break;
    }
    
    // Adicionar flags de segurança como propriedades de contexto
    if (event.security_flags && Array.isArray(event.security_flags)) {
      event.security_flags.forEach(flag => {
        context.setProperty(`securityFlag_${flag}`, true);
      });
    }
    
    // Adicionar informações de risco, se disponíveis
    if (event.risk_score !== undefined && event.risk_score !== null) {
      context.setProperty('riskScore', event.risk_score);
      context.setProperty('riskLevel', this.getRiskLevel(event.risk_score));
    }
    
    // Adicionar metadados específicos de setor, se aplicável
    if (this.sectorConfig) {
      context.setProperty('sector', this.sectorConfig.sector);
      
      if (this.sectorConfig.sector === 'HEALTHCARE') {
        context.setProperty('hipaaCompliant', this.regionalConfig.compliance.hipaa || false);
        context.setProperty('gdprHealthcareCompliant', this.regionalConfig.compliance.gdpr_healthcare || false);
        context.setProperty('lgpdHealthcareCompliant', this.regionalConfig.compliance.lgpd_healthcare || false);
      }
    }
  }
  
  /**
   * Obtém o nível de risco com base na pontuação
   * @private
   * @param {number} score - Pontuação de risco (0-100)
   * @returns {string} Nível de risco
   */
  getRiskLevel(score) {
    if (score <= 20) return 'LOW';
    if (score <= 50) return 'MEDIUM';
    if (score <= 80) return 'HIGH';
    return 'CRITICAL';
  }
  
  /**
   * Publica um evento Kafka como mensagem MCP
   * @param {string} topic - Tópico Kafka de origem
   * @param {Object} event - Evento Kafka
   * @param {Object} headers - Cabeçalhos Kafka
   * @returns {Promise<Object>} Resultado da publicação
   */
  async publishToMCP(topic, event, headers = {}) {
    if (!this.mcpClient || !this.mcpClient.isConnected()) {
      throw new Error('Cliente MCP não está conectado. Chame connect() primeiro.');
    }
    
    try {
      // Transformar evento Kafka em mensagem MCP
      const mcpMessage = this.transformToMCPMessage(topic, event, headers);
      
      // Publicar mensagem no MCP
      const result = await this.mcpClient.publish(mcpMessage);
      
      logger.debug(`Evento Kafka publicado como mensagem MCP: ${event.event_id}`, {
        topic,
        mcpChannel: mcpMessage.channel,
        eventType: event.event_type
      });
      
      return {
        success: true,
        messageId: result.messageId,
        channel: mcpMessage.channel,
        event_id: event.event_id
      };
      
    } catch (error) {
      logger.error(`Erro ao publicar evento Kafka como mensagem MCP: ${error.message}`, {
        error,
        topic,
        eventId: event.event_id
      });
      
      throw error;
    }
  }
  
  /**
   * Configura um consumidor MCP para receber mensagens e convertê-las para eventos Kafka
   * @param {string} mcpChannel - Canal MCP para inscrição
   * @param {Function} handler - Função para processar mensagens recebidas
   * @param {Object} options - Opções adicionais
   * @returns {Promise<string>} ID da inscrição
   */
  async subscribeMCP(mcpChannel, handler, options = {}) {
    if (!this.mcpClient || !this.mcpClient.isConnected()) {
      throw new Error('Cliente MCP não está conectado. Chame connect() primeiro.');
    }
    
    try {
      const subscriptionId = await this.mcpClient.subscribe(mcpChannel, async (mcpMessage) => {
        // Converter mensagem MCP para formato de evento Kafka
        const kafkaEvent = this.transformToKafkaEvent(mcpMessage, options);
        
        // Chamar handler com o evento convertido
        await handler(kafkaEvent, mcpMessage);
      }, options);
      
      logger.info(`Inscrito no canal MCP: ${mcpChannel}`, {
        subscriptionId,
        options
      });
      
      return subscriptionId;
      
    } catch (error) {
      logger.error(`Erro ao inscrever-se no canal MCP: ${error.message}`, {
        error,
        channel: mcpChannel
      });
      
      throw error;
    }
  }
  
  /**
   * Transforma uma mensagem MCP em um evento Kafka
   * @private
   * @param {MCPMessage} mcpMessage - Mensagem MCP
   * @param {Object} options - Opções adicionais
   * @returns {Object} Evento no formato Kafka
   */
  transformToKafkaEvent(mcpMessage, options = {}) {
    const payload = mcpMessage.payload;
    const context = mcpMessage.context;
    
    // Garantir que o evento tenha um ID e timestamp
    if (!payload.event_id) {
      payload.event_id = uuid.v4();
    }
    
    if (!payload.timestamp) {
      payload.timestamp = Date.now();
    }
    
    // Adicionar informações do contexto MCP ao evento
    if (options.includeContext !== false) {
      payload.correlation_id = context.requestId;
      payload.tenant_id = context.tenantId || payload.tenant_id;
      payload.region_code = context.regionCode || this.regionCode;
      
      // Adicionar propriedades específicas de contexto
      if (context.properties) {
        for (const [key, value] of Object.entries(context.properties)) {
          if (key.startsWith('securityFlag_')) {
            if (!payload.security_flags) {
              payload.security_flags = [];
            }
            payload.security_flags.push(key.substring('securityFlag_'.length));
          }
        }
      }
    }
    
    // Aplicar conformidade regional
    const compliantEvent = applyRegionalCompliance(payload, {
      region: this.regionCode,
      compliance: this.regionalConfig.compliance
    });
    
    return compliantEvent;
  }
  
  /**
   * Fecha a conexão com o broker MCP
   * @returns {Promise<void>}
   */
  async disconnect() {
    if (this.mcpClient) {
      await this.mcpClient.disconnect();
      logger.info('Adaptador MCP desconectado');
    }
  }
}

module.exports = MCPKafkaAdapter;
