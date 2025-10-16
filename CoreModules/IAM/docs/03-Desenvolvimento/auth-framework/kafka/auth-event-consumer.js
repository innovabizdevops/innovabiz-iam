/**
 * @fileoverview Consumidor de eventos de autenticação para o Apache Kafka
 * @author INNOVABIZ Dev Team
 * @version 1.0.0
 * @module iam/auth-framework/kafka/auth-event-consumer
 * @description Responsável por consumir e processar eventos de autenticação do Kafka
 */

const { Kafka, logLevel } = require('kafkajs');
const { SchemaRegistry } = require('@kafkajs/confluent-schema-registry');
const uuid = require('uuid');
const config = require('../../config/kafka-config');
const logger = require('../../core/utils/logging');
const { EventObservability } = require('../../core/utils/observability');
const { applyRegionalCompliance } = require('../../core/utils/compliance');

// Configurações regionais
const REGIONAL_CONFIGS = {
  EU: {
    max_poll_records: 500,
    session_timeout_ms: 30000,
    heartbeat_interval_ms: 3000,
    auto_commit: true,
    compliance: {
      gdpr: true,
      eidas: true
    }
  },
  BR: {
    max_poll_records: 500,
    session_timeout_ms: 30000,
    heartbeat_interval_ms: 3000,
    auto_commit: true,
    compliance: {
      lgpd: true,
      icp_brasil: true
    }
  },
  AO: {
    max_poll_records: 200,
    session_timeout_ms: 45000,
    heartbeat_interval_ms: 5000,
    auto_commit: true,
    compliance: {
      pndsb: true
    }
  },
  US: {
    max_poll_records: 1000,
    session_timeout_ms: 30000,
    heartbeat_interval_ms: 3000,
    auto_commit: true,
    compliance: {
      nist: true,
      soc2: true
    },
    sector_specific: {
      HEALTHCARE: {
        compliance: {
          hipaa: true
        }
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
 * Manipuladores padrão para eventos específicos
 * @type {Object}
 */
const DEFAULT_HANDLERS = {
  LOGIN_ATTEMPT: async (event) => {
    logger.debug(`Handler padrão para LOGIN_ATTEMPT: ${event.event_id}`);
    return { processed: true, action: 'none' };
  },
  LOGIN_SUCCESS: async (event) => {
    logger.debug(`Handler padrão para LOGIN_SUCCESS: ${event.event_id}`);
    return { processed: true, action: 'none' };
  },
  LOGIN_FAILURE: async (event) => {
    logger.debug(`Handler padrão para LOGIN_FAILURE: ${event.event_id}`);
    return { processed: true, action: 'none' };
  },
  MFA_CHALLENGE_ISSUED: async (event) => {
    logger.debug(`Handler padrão para MFA_CHALLENGE_ISSUED: ${event.event_id}`);
    return { processed: true, action: 'none' };
  },
  TOKEN_ISSUED: async (event) => {
    logger.debug(`Handler padrão para TOKEN_ISSUED: ${event.event_id}`);
    return { processed: true, action: 'none' };
  },
  SUSPICIOUS_ACTIVITY: async (event) => {
    logger.info(`Handler padrão para SUSPICIOUS_ACTIVITY: ${event.event_id}`);
    return { processed: true, action: 'alert' };
  }
};

/**
 * Classe responsável por consumir eventos de autenticação do Kafka
 */
class AuthEventConsumer {
  /**
   * Inicializa o consumidor de eventos de autenticação
   * @param {Object} options - Opções de configuração
   * @param {string} options.regionCode - Código da região (EU, BR, AO, US)
   * @param {string} options.groupId - ID do grupo de consumidores
   * @param {string} options.schemaRegistryUrl - URL do Schema Registry
   * @param {Object} options.kafkaConfig - Configurações do Kafka
   * @param {Object} options.eventHandlers - Manipuladores personalizados para eventos
   * @param {Object} options.sectorConfig - Configurações específicas do setor (opcional)
   * @param {boolean} options.enableDLQ - Habilitar fila de mensagens mortas (DLQ)
   */
  constructor(options = {}) {
    this.regionCode = options.regionCode || process.env.REGION_CODE || 'EU';
    this.groupId = options.groupId || `iam-auth-consumer-group-${this.regionCode.toLowerCase()}`;
    this.schemaRegistryUrl = options.schemaRegistryUrl || process.env.SCHEMA_REGISTRY_URL || 'http://iam-schema-registry:8081';
    this.kafkaConfig = options.kafkaConfig || config.kafka;
    this.eventHandlers = { ...DEFAULT_HANDLERS, ...options.eventHandlers };
    this.sectorConfig = options.sectorConfig || null;
    this.enableDLQ = options.enableDLQ !== undefined ? options.enableDLQ : true;
    
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
        ...this.regionalConfig.sector_specific[this.sectorConfig.sector],
        compliance: {
          ...this.regionalConfig.compliance,
          ...this.regionalConfig.sector_specific[this.sectorConfig.sector].compliance
        }
      };
    }
    
    // Inicializar observabilidade
    this.observability = new EventObservability({
      serviceName: 'iam-auth-consumer',
      region: this.regionCode
    });
    
    logger.info(`AuthEventConsumer inicializado para a região ${this.regionCode}`);
  }
  
  /**
   * Inicializa o cliente Kafka
   * @private
   */
  initKafkaClient() {
    this.kafka = new Kafka({
      clientId: `iam-auth-consumer-${this.regionCode.toLowerCase()}-${uuid.v4().substring(0, 8)}`,
      brokers: this.kafkaConfig.brokers,
      ssl: this.kafkaConfig.ssl,
      sasl: this.kafkaConfig.sasl,
      connectionTimeout: 3000,
      requestTimeout: 30000,
      retry: {
        initialRetryTime: 100,
        retries: 8
      },
      logLevel: logLevel.ERROR
    });
    
    this.consumer = this.kafka.consumer({
      groupId: this.groupId,
      maxBytesPerPartition: 1048576, // 1MB
      sessionTimeout: this.regionalConfig.session_timeout_ms,
      heartbeatInterval: this.regionalConfig.heartbeat_interval_ms,
      maxWaitTimeInMs: 5000,
      allowAutoTopicCreation: false
    });
    
    // Inicializar produtor para DLQ se habilitado
    if (this.enableDLQ) {
      this.dlqProducer = this.kafka.producer({
        allowAutoTopicCreation: true
      });
    }
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
   * Conecta ao Kafka e inicializa o consumidor
   * @returns {Promise<void>}
   */
  async connect() {
    try {
      await this.consumer.connect();
      logger.info(`Consumidor Kafka conectado com sucesso (Grupo: ${this.groupId})`);
      
      // Conectar produtor DLQ se habilitado
      if (this.enableDLQ && this.dlqProducer) {
        await this.dlqProducer.connect();
        logger.info('Produtor DLQ conectado com sucesso');
      }
    } catch (error) {
      logger.error(`Erro ao conectar ao Kafka: ${error.message}`, { error });
      throw error;
    }
  }
  
  /**
   * Inscreve-se em tópicos do Kafka
   * @param {Array<string>} topics - Lista de tópicos para inscrição
   * @param {Object} options - Opções adicionais
   * @returns {Promise<void>}
   */
  async subscribe(topics, options = {}) {
    if (!this.consumer) {
      throw new Error('Consumidor não inicializado. Chame connect() primeiro.');
    }
    
    const defaultTopic = 'iam-auth-events';
    const topicsToSubscribe = topics && topics.length > 0 ? topics : [defaultTopic];
    
    try {
      for (const topic of topicsToSubscribe) {
        await this.consumer.subscribe({
          topic,
          fromBeginning: options.fromBeginning || false
        });
        logger.info(`Inscrito no tópico: ${topic}`);
      }
    } catch (error) {
      logger.error(`Erro ao inscrever nos tópicos: ${error.message}`, { error, topics: topicsToSubscribe });
      throw error;
    }
  }
  
  /**
   * Inicia o consumo de mensagens
   * @param {Object} options - Opções adicionais
   * @returns {Promise<void>}
   */
  async run(options = {}) {
    if (!this.consumer) {
      throw new Error('Consumidor não inicializado. Chame connect() primeiro.');
    }
    
    const autoCommit = options.autoCommit !== undefined ? options.autoCommit : this.regionalConfig.auto_commit;
    
    try {
      await this.consumer.run({
        eachBatchAutoResolve: autoCommit,
        autoCommit,
        partitionsConsumedConcurrently: options.concurrentPartitions || 3,
        eachBatch: async ({ batch, resolveOffset, heartbeat, isRunning, commitOffsetsIfNecessary }) => {
          const { topic, partition, messages } = batch;
          logger.debug(`Processando lote de ${messages.length} mensagens do tópico ${topic} [partição ${partition}]`);
          
          // Métricas para observabilidade
          this.observability.recordBatchReceived({
            topic,
            partition,
            messageCount: messages.length
          });
          
          for (const message of messages) {
            if (!isRunning() || this.shuttingDown) break;
            
            try {
              // Decodificar mensagem com Schema Registry
              const value = await this.registry.decode(message.value);
              
              // Extrair informações dos cabeçalhos
              const headers = {};
              for (const headerKey in message.headers) {
                headers[headerKey] = message.headers[headerKey]?.toString();
              }
              
              // Aplicar conformidade regional
              const event = applyRegionalCompliance(value, {
                region: this.regionCode,
                compliance: this.regionalConfig.compliance
              });
              
              // Iniciar span para rastreabilidade
              const span = this.observability.startSpan('process_auth_event', {
                topic,
                partition,
                offset: message.offset,
                eventType: event.event_type,
                eventId: event.event_id,
                correlationId: event.correlation_id
              });
              
              // Processar evento
              const result = await this.processEvent(event, headers);
              
              // Finalizar span com resultado
              span.finish({
                result: result.processed ? 'success' : 'failure',
                action: result.action
              });
              
              // Resolver o offset se processado com sucesso
              if (result.processed) {
                resolveOffset(message.offset);
              } else if (this.enableDLQ && this.dlqProducer) {
                // Enviar para DLQ se não processado
                await this.sendToDLQ(topic, event, message, result.error);
                resolveOffset(message.offset);
              }
              
            } catch (error) {
              logger.error(`Erro ao processar mensagem: ${error.message}`, {
                error,
                topic,
                partition,
                offset: message.offset
              });
              
              // Registrar erro na telemetria
              this.observability.recordError({
                topic,
                partition,
                offset: message.offset,
                error
              });
              
              // Enviar para DLQ se habilitado
              if (this.enableDLQ && this.dlqProducer) {
                await this.sendToDLQ(topic, null, message, error);
                resolveOffset(message.offset);
              }
            }
            
            // Heartbeat para manter a sessão ativa
            await heartbeat();
          }
          
          // Commit manual se não auto-commit
          if (!autoCommit) {
            await commitOffsetsIfNecessary();
          }
        }
      });
      
      logger.info('Consumidor Kafka iniciado e processando mensagens');
    } catch (error) {
      logger.error(`Erro ao iniciar o consumo de mensagens: ${error.message}`, { error });
      throw error;
    }
  }
  
  /**
   * Processa um evento de autenticação
   * @private
   * @param {Object} event - Evento de autenticação
   * @param {Object} headers - Cabeçalhos da mensagem
   * @returns {Promise<Object>} Resultado do processamento
   */
  async processEvent(event, headers = {}) {
    const eventType = event.event_type;
    logger.debug(`Processando evento tipo ${eventType}: ${event.event_id}`);
    
    try {
      // Verificar se existe um handler para o tipo de evento
      if (this.eventHandlers[eventType]) {
        const result = await this.eventHandlers[eventType](event, headers);
        
        // Registrar métricas
        this.observability.recordEventProcessed({
          eventType,
          success: true,
          tenantId: event.tenant_id,
          regionCode: event.region_code,
          processingTime: Date.now() - event.timestamp
        });
        
        return { processed: true, ...result };
      } else {
        logger.warn(`Sem handler definido para o tipo de evento: ${eventType}`);
        
        // Registrar evento sem handler
        this.observability.recordEventProcessed({
          eventType,
          success: false,
          reason: 'no_handler',
          tenantId: event.tenant_id,
          regionCode: event.region_code
        });
        
        return { processed: false, action: 'skip', error: 'No handler defined' };
      }
    } catch (error) {
      logger.error(`Erro ao processar evento ${eventType}: ${error.message}`, {
        error,
        eventId: event.event_id
      });
      
      // Registrar erro
      this.observability.recordEventProcessed({
        eventType,
        success: false,
        reason: 'processing_error',
        tenantId: event.tenant_id,
        regionCode: event.region_code,
        error: error.message
      });
      
      return { processed: false, action: 'retry', error: error.message };
    }
  }
  
  /**
   * Envia uma mensagem para a fila de mensagens mortas (DLQ)
   * @private
   * @param {string} srcTopic - Tópico de origem
   * @param {Object} event - Evento decodificado (se disponível)
   * @param {Object} message - Mensagem original
   * @param {Object|string} error - Erro ocorrido
   * @returns {Promise<void>}
   */
  async sendToDLQ(srcTopic, event, message, error) {
    if (!this.dlqProducer) {
      logger.warn('Produtor DLQ não disponível');
      return;
    }
    
    const dlqTopic = `${srcTopic}.dlq`;
    
    try {
      await this.dlqProducer.send({
        topic: dlqTopic,
        messages: [{
          // Preservar a chave original
          key: message.key,
          // Preservar o valor original
          value: message.value,
          // Adicionar informações de erro nos cabeçalhos
          headers: {
            ...message.headers,
            'error-message': Buffer.from(String(error instanceof Error ? error.message : error)),
            'error-time': Buffer.from(String(Date.now())),
            'source-topic': Buffer.from(srcTopic),
            'source-partition': Buffer.from(String(message.partition)),
            'source-offset': Buffer.from(String(message.offset))
          }
        }]
      });
      
      logger.info(`Mensagem enviada para DLQ: ${dlqTopic}`, {
        srcTopic,
        eventId: event?.event_id || 'unknown',
        error: error instanceof Error ? error.message : String(error)
      });
      
    } catch (dlqError) {
      logger.error(`Erro ao enviar mensagem para DLQ: ${dlqError.message}`, {
        error: dlqError,
        srcTopic,
        dlqTopic,
        eventId: event?.event_id || 'unknown'
      });
    }
  }
  
  /**
   * Adiciona ou atualiza um manipulador de evento
   * @param {string} eventType - Tipo de evento
   * @param {Function} handler - Função manipuladora
   */
  addEventHandler(eventType, handler) {
    if (typeof handler !== 'function') {
      throw new Error('Handler deve ser uma função');
    }
    
    this.eventHandlers[eventType] = handler;
    logger.info(`Handler adicionado para evento ${eventType}`);
  }
  
  /**
   * Inicia o desligamento do consumidor
   * @returns {Promise<void>}
   */
  async shutdown() {
    logger.info('Iniciando desligamento do consumidor Kafka...');
    this.shuttingDown = true;
    
    try {
      // Aguardar um curto período para permitir o processamento em andamento
      await new Promise(resolve => setTimeout(resolve, 5000));
      
      // Desconectar consumidor
      if (this.consumer) {
        await this.consumer.disconnect();
        logger.info('Consumidor Kafka desconectado');
      }
      
      // Desconectar produtor DLQ
      if (this.dlqProducer) {
        await this.dlqProducer.disconnect();
        logger.info('Produtor DLQ desconectado');
      }
      
    } catch (error) {
      logger.error(`Erro ao desligar consumidor Kafka: ${error.message}`, { error });
      throw error;
    } finally {
      this.shuttingDown = false;
    }
  }
}

module.exports = AuthEventConsumer;
