/**
 * @fileoverview Produtor de eventos de autenticação para o Apache Kafka
 * @author INNOVABIZ Dev Team
 * @version 1.0.0
 * @module iam/auth-framework/kafka/auth-event-producer
 * @description Responsável por publicar eventos de autenticação no Kafka
 */

const { Kafka, CompressionTypes, logLevel } = require('kafkajs');
const { SchemaRegistry } = require('@kafkajs/confluent-schema-registry');
const uuid = require('uuid');
const config = require('../../config/kafka-config');
const logger = require('../../core/utils/logging');
const { maskSensitiveData } = require('../../core/utils/data-masking');

// Configurações regionais
const REGIONAL_CONFIGS = {
  EU: {
    data_masking: true,
    pii_fields: ['username', 'ip_address', 'email'],
    batch_size: 100,
    require_acks: -1 // all
  },
  BR: {
    data_masking: true,
    pii_fields: ['username', 'ip_address', 'email'],
    batch_size: 100,
    require_acks: -1
  },
  AO: {
    data_masking: false,
    pii_fields: [],
    batch_size: 50,
    require_acks: 1
  },
  US: {
    data_masking: false,
    pii_fields: [],
    batch_size: 100,
    require_acks: -1,
    sector_specific: {
      HEALTHCARE: {
        data_masking: true,
        pii_fields: ['username', 'ip_address', 'email', 'device_id']
      }
    }
  }
};

/**
 * Classe responsável por produzir eventos de autenticação para o Kafka
 */
class AuthEventProducer {
  /**
   * Inicializa o produtor de eventos de autenticação
   * @param {Object} options - Opções de configuração
   * @param {string} options.regionCode - Código da região (EU, BR, AO, US)
   * @param {string} options.schemaRegistryUrl - URL do Schema Registry
   * @param {Object} options.kafkaConfig - Configurações do Kafka
   * @param {Object} options.sectorConfig - Configurações específicas do setor (opcional)
   */
  constructor(options = {}) {
    this.regionCode = options.regionCode || process.env.REGION_CODE || 'EU';
    this.schemaRegistryUrl = options.schemaRegistryUrl || process.env.SCHEMA_REGISTRY_URL || 'http://iam-schema-registry:8081';
    this.kafkaConfig = options.kafkaConfig || config.kafka;
    this.sectorConfig = options.sectorConfig || null;
    
    // Inicializando o cliente Kafka e Schema Registry
    this.initKafkaClient();
    this.initSchemaRegistry();
    
    // Carregando configurações regionais
    this.regionalConfig = REGIONAL_CONFIGS[this.regionCode] || REGIONAL_CONFIGS.EU;
    
    // Se houver configuração específica do setor, aplicar
    if (this.sectorConfig && this.regionalConfig.sector_specific && 
        this.regionalConfig.sector_specific[this.sectorConfig.sector]) {
      this.regionalConfig = {
        ...this.regionalConfig,
        ...this.regionalConfig.sector_specific[this.sectorConfig.sector]
      };
    }
    
    // Inicializando IDs dos schemas
    this.schemaIds = {};
    
    logger.info(`AuthEventProducer inicializado para a região ${this.regionCode}`);
  }
  
  /**
   * Inicializa o cliente Kafka
   * @private
   */
  initKafkaClient() {
    this.kafka = new Kafka({
      clientId: `iam-auth-producer-${this.regionCode.toLowerCase()}`,
      brokers: this.kafkaConfig.brokers,
      ssl: this.kafkaConfig.ssl,
      sasl: this.kafkaConfig.sasl,
      connectionTimeout: 3000,
      requestTimeout: 25000,
      retry: {
        initialRetryTime: 100,
        retries: 8
      },
      logLevel: logLevel.ERROR
    });
    
    this.producer = this.kafka.producer({
      allowAutoTopicCreation: false,
      transactionalId: `iam-auth-producer-${this.regionCode.toLowerCase()}-${uuid.v4()}`,
      maxInFlightRequests: 5,
      idempotent: true
    });
  }
  
  /**
   * Inicializa o cliente do Schema Registry
   * @private
   */
  initSchemaRegistry() {
    this.registry = new SchemaRegistry({
      host: this.schemaRegistryUrl,
      auth: this.kafkaConfig.schemaRegistry.auth
    });
  }
  
  /**
   * Conecta ao Kafka e inicializa o produtor
   * @returns {Promise<void>}
   */
  async connect() {
    try {
      await this.producer.connect();
      logger.info('Produtor Kafka conectado com sucesso');
      
      // Pré-carregando IDs dos schemas para melhor performance
      await this.loadSchemaIds();
    } catch (error) {
      logger.error(`Erro ao conectar ao Kafka: ${error.message}`, { error });
      throw error;
    }
  }
  
  /**
   * Carrega os IDs dos schemas do registro
   * @private
   * @returns {Promise<void>}
   */
  async loadSchemaIds() {
    try {
      this.schemaIds.authEvent = await this.registry.getLatestSchemaId('io.innovabiz.iam.events.AuthenticationEvent');
      this.schemaIds.tokenEvent = await this.registry.getLatestSchemaId('io.innovabiz.iam.events.TokenEvent');
      this.schemaIds.mfaEvent = await this.registry.getLatestSchemaId('io.innovabiz.iam.events.MfaChallengeEvent');
      
      logger.debug('IDs dos schemas carregados com sucesso', { schemaIds: this.schemaIds });
    } catch (error) {
      logger.warn(`Erro ao carregar IDs dos schemas: ${error.message}. Serão carregados sob demanda.`, { error });
    }
  }
  
  /**
   * Processa e prepara um evento de autenticação
   * @private
   * @param {Object} event - Evento de autenticação
   * @returns {Object} Evento processado
   */
  processAuthEvent(event) {
    // Adicionar campos obrigatórios se não existirem
    const processedEvent = {
      event_id: event.event_id || uuid.v4(),
      timestamp: event.timestamp || Date.now(),
      correlation_id: event.correlation_id || uuid.v4(),
      region_code: this.regionCode,
      ...event
    };
    
    // Aplicar mascaramento de dados se necessário
    if (this.regionalConfig.data_masking && this.regionalConfig.pii_fields.length > 0) {
      return maskSensitiveData(processedEvent, this.regionalConfig.pii_fields);
    }
    
    return processedEvent;
  }
  
  /**
   * Publica um evento de autenticação no Kafka
   * @param {Object} event - Evento de autenticação
   * @param {Object} options - Opções adicionais
   * @returns {Promise<Object>} Resultado da operação
   */
  async publishAuthEvent(event, options = {}) {
    if (!this.producer) {
      throw new Error('Produtor não inicializado. Chame connect() primeiro.');
    }
    
    const processedEvent = this.processAuthEvent(event);
    const topicName = options.topic || 'iam-auth-events';
    
    try {
      // Codificar evento com Avro
      let value;
      if (this.schemaIds.authEvent) {
        // Usar ID de schema pré-carregado
        value = await this.registry.encode(this.schemaIds.authEvent, processedEvent);
      } else {
        // Codificar com o schema completo
        value = await this.registry.encode({
          type: 'io.innovabiz.iam.events.AuthenticationEvent',
          schema: processedEvent
        });
      }
      
      // Publicar no Kafka
      const result = await this.producer.send({
        topic: topicName,
        compression: CompressionTypes.GZIP,
        acks: this.regionalConfig.require_acks,
        messages: [{
          key: processedEvent.user_id || processedEvent.event_id,
          value,
          headers: {
            'event-type': processedEvent.event_type,
            'correlation-id': processedEvent.correlation_id,
            'region-code': this.regionCode,
            'tenant-id': processedEvent.tenant_id
          },
          timestamp: processedEvent.timestamp.toString()
        }]
      });
      
      logger.debug(`Evento de autenticação publicado com sucesso: ${processedEvent.event_id}`, {
        topic: topicName,
        eventType: processedEvent.event_type
      });
      
      return {
        success: true,
        topic: topicName,
        partition: result[0].partition,
        offset: result[0].baseOffset,
        event_id: processedEvent.event_id
      };
      
    } catch (error) {
      logger.error(`Erro ao publicar evento de autenticação: ${error.message}`, {
        error,
        eventId: processedEvent.event_id,
        topic: topicName
      });
      
      throw error;
    }
  }
  
  /**
   * Publica um evento de token no Kafka
   * @param {Object} event - Evento de token
   * @param {Object} options - Opções adicionais
   * @returns {Promise<Object>} Resultado da operação
   */
  async publishTokenEvent(event, options = {}) {
    // Implementação similar ao publishAuthEvent, mas para eventos de token
    // Código omitido para brevidade
    // ...
  }
  
  /**
   * Publica um evento de desafio MFA no Kafka
   * @param {Object} event - Evento de desafio MFA
   * @param {Object} options - Opções adicionais
   * @returns {Promise<Object>} Resultado da operação
   */
  async publishMfaChallengeEvent(event, options = {}) {
    // Implementação similar ao publishAuthEvent, mas para eventos de desafio MFA
    // Código omitido para brevidade
    // ...
  }
  
  /**
   * Publica um lote de eventos no Kafka
   * @param {Array<Object>} events - Lista de eventos
   * @param {string} eventType - Tipo de evento ('auth', 'token', 'mfa')
   * @param {Object} options - Opções adicionais
   * @returns {Promise<Object>} Resultado da operação
   */
  async publishBatch(events, eventType = 'auth', options = {}) {
    if (!this.producer) {
      throw new Error('Produtor não inicializado. Chame connect() primeiro.');
    }
    
    if (!Array.isArray(events) || events.length === 0) {
      throw new Error('É necessário fornecer uma lista não vazia de eventos');
    }
    
    const batchSize = options.batchSize || this.regionalConfig.batch_size || 100;
    const topicName = options.topic || 
      (eventType === 'auth' ? 'iam-auth-events' : 
       eventType === 'token' ? 'iam-token-events' : 
       eventType === 'mfa' ? 'iam-mfa-challenges' : 'iam-auth-events');
    
    try {
      // Processar eventos em lotes
      const batches = [];
      for (let i = 0; i < events.length; i += batchSize) {
        batches.push(events.slice(i, i + batchSize));
      }
      
      const results = [];
      for (const batch of batches) {
        const messages = await Promise.all(batch.map(async (event) => {
          const processedEvent = this.processAuthEvent(event);
          
          // Codificar evento com Avro
          let value;
          const schemaIdKey = 
            eventType === 'auth' ? 'authEvent' : 
            eventType === 'token' ? 'tokenEvent' : 
            eventType === 'mfa' ? 'mfaEvent' : 'authEvent';
          
          if (this.schemaIds[schemaIdKey]) {
            value = await this.registry.encode(this.schemaIds[schemaIdKey], processedEvent);
          } else {
            value = await this.registry.encode({
              type: `io.innovabiz.iam.events.${
                eventType === 'auth' ? 'AuthenticationEvent' : 
                eventType === 'token' ? 'TokenEvent' : 
                eventType === 'mfa' ? 'MfaChallengeEvent' : 'AuthenticationEvent'
              }`,
              schema: processedEvent
            });
          }
          
          return {
            key: processedEvent.user_id || processedEvent.event_id,
            value,
            headers: {
              'event-type': processedEvent.event_type,
              'correlation-id': processedEvent.correlation_id,
              'region-code': this.regionCode,
              'tenant-id': processedEvent.tenant_id
            },
            timestamp: processedEvent.timestamp.toString()
          };
        }));
        
        // Enviar lote para o Kafka
        const result = await this.producer.send({
          topic: topicName,
          compression: CompressionTypes.GZIP,
          acks: this.regionalConfig.require_acks,
          messages
        });
        
        results.push(result);
      }
      
      logger.info(`Lote de ${events.length} eventos publicados com sucesso no tópico ${topicName}`);
      
      return {
        success: true,
        topic: topicName,
        batchCount: batches.length,
        eventCount: events.length,
        results
      };
      
    } catch (error) {
      logger.error(`Erro ao publicar lote de eventos: ${error.message}`, {
        error,
        eventType,
        eventCount: events.length,
        topic: topicName
      });
      
      throw error;
    }
  }
  
  /**
   * Fecha a conexão com o Kafka
   * @returns {Promise<void>}
   */
  async disconnect() {
    if (this.producer) {
      await this.producer.disconnect();
      logger.info('Produtor Kafka desconectado');
    }
  }
}

module.exports = AuthEventProducer;
