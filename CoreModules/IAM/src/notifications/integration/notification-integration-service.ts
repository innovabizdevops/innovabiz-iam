/**
 * @file notification-integration-service.ts
 * @description Serviço de integração do sistema de notificações com outros módulos
 * 
 * Este serviço facilita a integração do sistema de notificações com outros
 * módulos da plataforma InnovaBiz, permitindo que eventos de diferentes
 * sistemas gerem notificações de forma padronizada e configurável.
 */

import { Logger } from '../../../infrastructure/observability/logger';
import { NotificationService } from '../services/notification-service';
import { TemplateService } from '../core/notification-template';
import { NotificationChannel } from '../core/notification-channel';
import { BaseEvent } from '../core/base-event';
import { EventProcessor, validateEvent } from './event-processor';
import { NotificationRecipient } from '../adapters/notification-adapter';

/**
 * Configuração para o serviço de integração
 */
export interface NotificationIntegrationConfig {
  /**
   * Diretório para regras de mapeamento de eventos
   */
  eventMappingsPath?: string;
  
  /**
   * Configuração de canais padrão por módulo
   */
  defaultChannelsByModule: Record<string, NotificationChannel[]>;
  
  /**
   * Templates padrão por tipo de evento
   */
  defaultTemplatesByEventType: Record<string, string>;
  
  /**
   * Modo de execução (síncrono ou assíncrono)
   */
  executionMode?: 'sync' | 'async';
  
  /**
   * Timeout para processamento de eventos (ms)
   */
  processingTimeoutMs?: number;
  
  /**
   * Número máximo de tentativas de processamento
   */
  maxProcessingAttempts?: number;
}

/**
 * Serviço de integração de notificações
 */
export class NotificationIntegrationService {
  private notificationService: NotificationService;
  private templateService: TemplateService;
  private eventProcessors: Map<string, EventProcessor> = new Map();
  private config: NotificationIntegrationConfig;
  private logger = new Logger('NotificationIntegrationService');
  
  /**
   * Construtor
   * @param notificationService Serviço de notificações
   * @param templateService Serviço de templates
   * @param config Configuração de integração
   */
  constructor(
    notificationService: NotificationService,
    templateService: TemplateService,
    config: NotificationIntegrationConfig
  ) {
    this.notificationService = notificationService;
    this.templateService = templateService;
    this.config = config;
  }
  
  /**
   * Inicializa o serviço de integração
   */
  async initialize(): Promise<void> {
    this.logger.info('Inicializando serviço de integração de notificações');
    
    // Carregar processadores de eventos padrão
    await this.loadEventProcessors();
    
    // Inicializar conexões com outros módulos
    await this.initializeModuleConnections();
    
    this.logger.info('Serviço de integração de notificações inicializado com sucesso');
  }
  
  /**
   * Carrega os processadores de eventos
   */
  private async loadEventProcessors(): Promise<void> {
    // Registrar processadores básicos
    this.registerEventProcessors();
  }
  
  /**
   * Registra os processadores de eventos básicos
   */
  private registerEventProcessors(): void {
    // Registrar processadores específicos por módulo
    this.registerIamEventProcessors();
    this.registerPaymentGatewayEventProcessors();
    this.registerMobileMoneyEventProcessors();
    this.registerECommerceEventProcessors();
    this.registerBureauCreditoEventProcessors();
  }
  
  /**
   * Inicializa conexões com outros módulos
   */
  private async initializeModuleConnections(): Promise<void> {
    // TODO: Implementar inicialização de conexões com outros módulos
    // Por exemplo, configurar listeners para eventos de outros módulos
  }
  
  /**
   * Registra processadores de eventos do módulo IAM
   */
  private registerIamEventProcessors(): void {
    // Processador para eventos de autenticação
    this.registerEventProcessor('iam', 'authentication', async (event) => {
      const eventType = event.type || 'unknown';
      const templateId = this.config.defaultTemplatesByEventType[`iam.authentication.${eventType}`] || 
                         this.config.defaultTemplatesByEventType['iam.authentication.default'];
                         
      if (!templateId) {
        this.logger.warn(`Template não encontrado para evento IAM.Authentication.${eventType}`);
        return null;
      }
      
      // Determinar destinatários com base no evento
      const recipients = await this.extractRecipientsFromEvent(event);
      
      if (!recipients || recipients.length === 0) {
        this.logger.warn(`Nenhum destinatário encontrado para evento ${event.id}`);
        return null;
      }
      
      // Determinar canais para o módulo IAM
      const channels = this.config.defaultChannelsByModule['iam'] || [NotificationChannel.EMAIL];
      
      // Para cada destinatário, enviar notificação
      const results = [];
      for (const recipient of recipients) {
        // Renderizar template com dados do evento
        const rendered = await this.templateService.render(templateId, {
          targetChannel: channels[0], // Usar primeiro canal como principal
          variables: {
            ...event.data,
            recipientId: recipient.id,
            eventId: event.id,
            eventType: event.type,
            timestamp: event.timestamp
          }
        });
        
        // Enviar notificação
        const result = await this.notificationService.send(
          recipient,
          rendered.content,
          {
            preferredChannels: channels,
            notificationId: `iam-auth-${event.id}-${recipient.id}`,
            tracking: {
              source: 'iam',
              category: 'authentication',
              tags: ['security', 'account', event.type || ''],
              metadata: {
                eventId: event.id,
                moduleSource: 'iam'
              }
            }
          },
          event
        );
        
        results.push(result);
      }
      
      return results;
    });
    
    // Processador para eventos de autorização
    this.registerEventProcessor('iam', 'authorization', async (event) => {
      const eventType = event.type || 'unknown';
      const templateId = this.config.defaultTemplatesByEventType[`iam.authorization.${eventType}`] || 
                         this.config.defaultTemplatesByEventType['iam.authorization.default'];
      
      if (!templateId) {
        this.logger.warn(`Template não encontrado para evento IAM.Authorization.${eventType}`);
        return null;
      }
      
      // Determinar destinatários
      const recipients = await this.extractRecipientsFromEvent(event);
      
      if (!recipients || recipients.length === 0) {
        return null;
      }
      
      // Determinar canais (pode priorizar diferentes canais para autorização)
      const channels = this.config.defaultChannelsByModule['iam.authorization'] || 
                       this.config.defaultChannelsByModule['iam'] ||
                       [NotificationChannel.EMAIL];
      
      // Para cada destinatário, enviar notificação
      const results = [];
      for (const recipient of recipients) {
        const rendered = await this.templateService.render(templateId, {
          targetChannel: channels[0],
          variables: {
            ...event.data,
            recipientId: recipient.id,
            eventId: event.id,
            eventType: event.type,
            timestamp: event.timestamp
          }
        });
        
        const result = await this.notificationService.send(
          recipient,
          rendered.content,
          {
            preferredChannels: channels,
            notificationId: `iam-auth-${event.id}-${recipient.id}`,
            tracking: {
              source: 'iam',
              category: 'authorization',
              tags: ['security', 'permissions', event.type || ''],
              metadata: {
                eventId: event.id,
                moduleSource: 'iam'
              }
            }
          },
          event
        );
        
        results.push(result);
      }
      
      return results;
    });
    
    // Processador para eventos de gestão de usuários
    this.registerEventProcessor('iam', 'user-management', async (event) => {
      const eventType = event.type || 'unknown';
      const templateId = this.config.defaultTemplatesByEventType[`iam.user-management.${eventType}`] || 
                         this.config.defaultTemplatesByEventType['iam.user-management.default'];
      
      if (!templateId) {
        this.logger.warn(`Template não encontrado para evento IAM.UserManagement.${eventType}`);
        return null;
      }
      
      // Restante da implementação semelhante aos outros processadores
      return null;
    });
  }
  
  /**
   * Registra processadores de eventos do módulo Payment Gateway
   */
  private registerPaymentGatewayEventProcessors(): void {
    // Processador para eventos de pagamento
    this.registerEventProcessor('payment-gateway', 'payment', async (event) => {
      const eventType = event.type || 'unknown';
      const templateId = this.config.defaultTemplatesByEventType[`payment-gateway.payment.${eventType}`] || 
                         this.config.defaultTemplatesByEventType['payment-gateway.payment.default'];
                         
      if (!templateId) {
        this.logger.warn(`Template não encontrado para evento PaymentGateway.Payment.${eventType}`);
        return null;
      }
      
      // Determinar destinatários com base no evento
      const recipients = await this.extractRecipientsFromEvent(event);
      
      if (!recipients || recipients.length === 0) {
        this.logger.warn(`Nenhum destinatário encontrado para evento ${event.id}`);
        return null;
      }
      
      // Determinar canais para o módulo Payment Gateway
      const channels = this.config.defaultChannelsByModule['payment-gateway'] || [
        NotificationChannel.EMAIL,
        NotificationChannel.SMS
      ];
      
      // Para cada destinatário, enviar notificação
      const results = [];
      for (const recipient of recipients) {
        // Renderizar template com dados do evento
        const rendered = await this.templateService.render(templateId, {
          targetChannel: channels[0], // Usar primeiro canal como principal
          variables: {
            ...event.data,
            recipientId: recipient.id,
            eventId: event.id,
            eventType: event.type,
            timestamp: event.timestamp
          }
        });
        
        // Enviar notificação
        const result = await this.notificationService.send(
          recipient,
          rendered.content,
          {
            preferredChannels: channels,
            notificationId: `pg-payment-${event.id}-${recipient.id}`,
            tracking: {
              source: 'payment-gateway',
              category: 'payment',
              tags: ['payment', 'transaction', event.type || ''],
              metadata: {
                eventId: event.id,
                moduleSource: 'payment-gateway'
              }
            }
          },
          event
        );
        
        results.push(result);
      }
      
      return results;
    });
  }
  
  /**
   * Registra processadores de eventos do módulo Mobile Money
   */
  private registerMobileMoneyEventProcessors(): void {
    // Implementação para eventos do Mobile Money
    this.registerEventProcessor('mobile-money', 'transaction', async (event) => {
      const eventType = event.type || 'unknown';
      const templateId = this.config.defaultTemplatesByEventType[`mobile-money.transaction.${eventType}`] || 
                         this.config.defaultTemplatesByEventType['mobile-money.transaction.default'];
      
      if (!templateId) {
        this.logger.warn(`Template não encontrado para evento MobileMoney.Transaction.${eventType}`);
        return null;
      }
      
      // Priorizar SMS para Mobile Money
      const channels = this.config.defaultChannelsByModule['mobile-money'] || [
        NotificationChannel.SMS,
        NotificationChannel.PUSH,
        NotificationChannel.EMAIL
      ];
      
      const recipients = await this.extractRecipientsFromEvent(event);
      
      if (!recipients || recipients.length === 0) {
        return null;
      }
      
      // Implementar lógica específica para Mobile Money
      // Por exemplo, adicionar informações sobre agente se for transação presencial
      
      return null;
    });
  }
  
  /**
   * Registra processadores de eventos do módulo E-Commerce
   */
  private registerECommerceEventProcessors(): void {
    // Implementação para eventos do E-Commerce
    this.registerEventProcessor('e-commerce', 'order', async (event) => {
      const eventType = event.type || 'unknown';
      const templateId = this.config.defaultTemplatesByEventType[`e-commerce.order.${eventType}`] || 
                         this.config.defaultTemplatesByEventType['e-commerce.order.default'];
      
      if (!templateId) {
        this.logger.warn(`Template não encontrado para evento ECommerce.Order.${eventType}`);
        return null;
      }
      
      // Para pedidos, pode-se usar múltiplos canais com diferente conteúdo
      const channels = this.config.defaultChannelsByModule['e-commerce'] || [
        NotificationChannel.EMAIL,
        NotificationChannel.SMS,
        NotificationChannel.PUSH
      ];
      
      // Implementação específica para pedidos de e-commerce
      
      return null;
    });
  }
  
  /**
   * Registra processadores de eventos do módulo Bureau de Crédito
   */
  private registerBureauCreditoEventProcessors(): void {
    // Implementação para eventos do Bureau de Crédito
    this.registerEventProcessor('bureau-credito', 'credit-check', async (event) => {
      const eventType = event.type || 'unknown';
      const templateId = this.config.defaultTemplatesByEventType[`bureau-credito.credit-check.${eventType}`] || 
                         this.config.defaultTemplatesByEventType['bureau-credito.credit-check.default'];
      
      if (!templateId) {
        this.logger.warn(`Template não encontrado para evento BureauCredito.CreditCheck.${eventType}`);
        return null;
      }
      
      // Para bureau de crédito, pode-se priorizar canais seguros
      const channels = this.config.defaultChannelsByModule['bureau-credito'] || [
        NotificationChannel.EMAIL
      ];
      
      // Implementação específica para consultas de crédito
      
      return null;
    });
  }
  
  /**
   * Registra um processador de eventos
   * @param module Módulo de origem
   * @param eventCategory Categoria do evento
   * @param processor Função de processamento do evento
   */
  registerEventProcessor(
    module: string,
    eventCategory: string,
    processor: EventProcessor
  ): void {
    const key = `${module.toLowerCase()}.${eventCategory.toLowerCase()}`;
    this.eventProcessors.set(key, processor);
    this.logger.debug(`Processador de eventos registrado para ${key}`);
  }
  
  /**
   * Processa um evento de sistema para notificação
   * @param event Evento a ser processado
   */
  async processEvent(event: BaseEvent): Promise<any> {
    try {
      // Validar evento
      if (!validateEvent(event)) {
        throw new Error(`Evento inválido: ${event.module}.${event.category}.${event.type}`);
      }
      
      if (!event.module || !event.category) {
        throw new Error('Evento inválido: módulo e categoria são obrigatórios');
      }
      
      const processorKey = `${event.module.toLowerCase()}.${event.category.toLowerCase()}`;
      const processor = this.eventProcessors.get(processorKey);
      
      if (!processor) {
        this.logger.warn(`Processador não encontrado para evento ${processorKey}`);
        return null;
      }
      
      // Executar processador de evento
      this.logger.info(`Processando evento ${event.id} do tipo ${processorKey}`);
      
      // Executar de forma síncrona ou assíncrona
      if (this.config.executionMode === 'async') {
        // Executar assincronamente
        setImmediate(async () => {
          try {
            await this.executeEventProcessor(processor, event);
          } catch (error) {
            this.logger.error(`Erro no processamento assíncrono de evento ${event.id}: ${error}`);
          }
        });
        return { queued: true, eventId: event.id };
      } else {
        // Executar sincronamente
        return await this.executeEventProcessor(processor, event);
      }
    } catch (error) {
      this.logger.error(`Erro ao processar evento ${event.id}: ${error}`);
      throw error;
    }
  }
  
  /**
   * Executa um processador de eventos com timeout e retry
   * @param processor Processador de eventos
   * @param event Evento a ser processado
   */
  private async executeEventProcessor(
    processor: EventProcessor,
    event: BaseEvent
  ): Promise<any> {
    const timeout = this.config.processingTimeoutMs || 30000;
    const maxAttempts = this.config.maxProcessingAttempts || 3;
    
    for (let attempt = 0; attempt < maxAttempts; attempt++) {
      try {
        const result = await Promise.race([
          processor(event),
          new Promise((_, reject) => {
            setTimeout(() => reject(new Error('Timeout no processamento de evento')), timeout);
          })
        ]);
        
        return result;
      } catch (error) {
        if (attempt === maxAttempts - 1) {
          this.logger.error(`Todas as ${maxAttempts} tentativas falharam para evento ${event.id}: ${error}`);
          throw error;
        }
        
        this.logger.warn(`Tentativa ${attempt + 1}/${maxAttempts} falhou para evento ${event.id}: ${error}. Tentando novamente...`);
        await this.delay(1000 * Math.pow(2, attempt)); // Backoff exponencial
      }
    }
  }
  
  /**
   * Extrai destinatários de um evento de sistema
   * @param event Evento de sistema
   */
  private async extractRecipientsFromEvent(event: BaseEvent): Promise<NotificationRecipient[]> {
    // Esta é uma implementação básica que extrai destinatários do próprio evento
    // Em um sistema real, isso poderia consultar bancos de dados ou outros serviços
    
    const recipients: NotificationRecipient[] = [];
    
    // Verificar se o evento tem destinatários definidos
    if (event.recipients) {
      return event.recipients;
    }
    
    // Verificar se o evento tem usuário relacionado
    if (event.data?.userId) {
      // Aqui poderia haver uma consulta ao banco de dados para obter mais detalhes do usuário
      recipients.push({
        id: event.data.userId,
        metadata: {
          email: event.data.userEmail,
          phone: event.data.userPhone,
          name: event.data.userName || event.data.fullName || 'Usuário'
        }
      });
    }
    
    // Para eventos de pagamento, enviar para compradores e vendedores
    if (event.module === 'payment-gateway' && event.data?.buyerId) {
      recipients.push({
        id: event.data.buyerId,
        metadata: {
          email: event.data.buyerEmail,
          phone: event.data.buyerPhone,
          name: event.data.buyerName || 'Comprador'
        }
      });
    }
    
    if (event.module === 'payment-gateway' && event.data?.sellerId) {
      recipients.push({
        id: event.data.sellerId,
        metadata: {
          email: event.data.sellerEmail,
          phone: event.data.sellerPhone,
          name: event.data.sellerName || 'Vendedor'
        }
      });
    }
    
    // Para eventos mobile money
    if (event.module === 'mobile-money' && event.data?.userPhone) {
      recipients.push({
        id: event.data.userPhone,
        metadata: {
          phone: event.data.userPhone,
          email: event.data.userEmail,
          name: event.data.userName || 'Cliente'
        }
      });
    }
    
    // Para eventos de e-commerce
    if (event.module === 'e-commerce' && event.data?.customerId) {
      recipients.push({
        id: event.data.customerId,
        metadata: {
          email: event.data.customerEmail,
          phone: event.data.customerPhone,
          name: event.data.customerName || 'Cliente'
        }
      });
    }
    
    return recipients;
  }
  
  /**
   * Cria um delay (promise) por um tempo específico
   * @param ms Tempo em milissegundos
   */
  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}