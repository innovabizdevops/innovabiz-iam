/**
 * @file notification-orchestrator.ts
 * @description Orquestrador de notificações para entrega multi-canal
 * 
 * Este módulo implementa o orquestrador que coordena o envio de notificações
 * através de múltiplos canais, gerencia prioridades, implementa estratégias
 * de fallback e controla a entrega de notificações.
 */

import { v4 as uuidv4 } from 'uuid';
import { Logger } from '../../../infrastructure/observability/logger';
import { NotificationAdapter, NotificationContent, NotificationRecipient, NotificationResult } from '../adapters/notification-adapter';
import { NotificationChannel } from './notification-channel';
import { DeliveryStrategy } from './delivery-strategy';
import { BaseEvent } from './base-event';
import { NotificationPreferences } from './notification-preferences';
import { NotificationRepository } from '../repositories/notification-repository';
import { EventRepository } from '../repositories/event-repository';
import { RecipientRepository } from '../repositories/recipient-repository';

/**
 * Status de entrega de uma notificação
 */
export enum DeliveryStatus {
  /**
   * Notificação pendente de envio
   */
  PENDING = 'PENDING',
  
  /**
   * Notificação em processamento
   */
  PROCESSING = 'PROCESSING',
  
  /**
   * Notificação entregue com sucesso em pelo menos um canal
   */
  DELIVERED = 'DELIVERED',
  
  /**
   * Notificação falhou em todos os canais
   */
  FAILED = 'FAILED',
  
  /**
   * Notificação cancelada antes do envio
   */
  CANCELLED = 'CANCELLED',
  
  /**
   * Notificação expirada antes de ser entregue
   */
  EXPIRED = 'EXPIRED'
}

/**
 * Resultado de entrega em um canal específico
 */
export interface ChannelDeliveryResult extends NotificationResult {
  /**
   * Canal usado para entrega
   */
  channel: NotificationChannel;
}

/**
 * Resultado da tentativa de entrega de uma notificação
 */
export interface DeliveryAttemptResult {
  /**
   * ID único da tentativa
   */
  attemptId: string;
  
  /**
   * ID da notificação
   */
  notificationId: string;
  
  /**
   * Timestamp da tentativa
   */
  timestamp: Date;
  
  /**
   * Resultados por canal
   */
  channelResults: ChannelDeliveryResult[];
  
  /**
   * Status geral da tentativa
   */
  status: DeliveryStatus;
  
  /**
   * Canais com entrega bem-sucedida
   */
  successfulChannels: NotificationChannel[];
  
  /**
   * Canais com falha na entrega
   */
  failedChannels: NotificationChannel[];
  
  /**
   * Mensagem de erro consolidada (se houver falhas)
   */
  errorMessage?: string;
}

/**
 * Opções para envio de notificações
 */
export interface NotificationOrchestratorOptions {
  /**
   * Estratégia de entrega a ser utilizada
   */
  deliveryStrategy?: DeliveryStrategy;
  
  /**
   * Lista ordenada de canais preferenciais
   */
  preferredChannels?: NotificationChannel[];
  
  /**
   * Tentativa automática em canais alternativos em caso de falha
   */
  enableFallback?: boolean;
  
  /**
   * Tempo máximo (em segundos) para considerar a notificação válida
   */
  expirationSeconds?: number;
  
  /**
   * Número máximo de tentativas de entrega
   */
  maxAttempts?: number;
  
  /**
   * Canais a serem excluídos da entrega
   */
  excludedChannels?: NotificationChannel[];
  
  /**
   * ID da transação relacionada para rastreamento
   */
  transactionId?: string;
  
  /**
   * Metadados adicionais
   */
  metadata?: Record<string, any>;
  
  /**
   * Forçar entrega em todos os canais disponíveis (ignorando preferências)
   */
  deliverToAllChannels?: boolean;
  
  /**
   * Agendar entrega para uma data futura
   */
  scheduleFor?: Date;
  
  /**
   * Grupo de notificações para agregação
   */
  notificationGroup?: string;
  
  /**
   * Ignorar configurações de horário silencioso (DND)
   */
  bypassDoNotDisturb?: boolean;
  
  /**
   * Prioridade da notificação (quanto maior, mais importante)
   */
  priority?: number;
}

/**
 * Orquestrador de notificações
 * 
 * Gerencia o ciclo de vida das notificações, coordena o envio através de múltiplos
 * canais e implementa estratégias de entrega.
 */
export class NotificationOrchestrator {
  // Adaptadores para cada canal de notificação
  private adapters: Map<NotificationChannel, NotificationAdapter> = new Map();
  private logger = new Logger('NotificationOrchestrator');
  
  /**
   * Construtor
   * @param notificationRepository Repositório de notificações
   * @param eventRepository Repositório de eventos
   * @param recipientRepository Repositório de destinatários
   */
  constructor(
    private notificationRepository: NotificationRepository,
    private eventRepository: EventRepository,
    private recipientRepository: RecipientRepository
  ) {}

  /**
   * Registra um adaptador para um canal específico
   * @param channel Canal de notificação
   * @param adapter Adaptador para o canal
   */
  registerAdapter(channel: NotificationChannel, adapter: NotificationAdapter): void {
    this.adapters.set(channel, adapter);
    this.logger.info(`Adaptador registrado para canal ${channel}`);
  }
  
  /**
   * Verifica se há um adaptador registrado e inicializado para o canal
   * @param channel Canal a verificar
   */
  private async isChannelAvailable(channel: NotificationChannel): Promise<boolean> {
    const adapter = this.adapters.get(channel);
    if (!adapter) return false;
    
    return await adapter.isReady();
  }
  
  /**
   * Determina os canais disponíveis para um destinatário específico
   * @param recipient Destinatário da notificação
   * @param options Opções de notificação
   */
  private async getAvailableChannels(
    recipient: NotificationRecipient,
    options?: NotificationOrchestratorOptions
  ): Promise<NotificationChannel[]> {
    const availableChannels: NotificationChannel[] = [];
    const recipientAddresses = recipient.addresses || new Map<NotificationChannel, string[]>();
    
    // Verificar cada canal registrado
    for (const [channel, adapter] of this.adapters.entries()) {
      // Pular canais excluídos
      if (options?.excludedChannels?.includes(channel)) {
        continue;
      }
      
      // Verificar se o adaptador está disponível
      const isAdapterAvailable = await adapter.isReady();
      if (!isAdapterAvailable) {
        continue;
      }
      
      // Verificar se o destinatário tem um endereço para este canal
      const addresses = recipientAddresses.get(channel);
      if (!addresses || addresses.length === 0) {
        continue;
      }
      
      // Canal disponível para uso
      availableChannels.push(channel);
    }
    
    return availableChannels;
  }
  
  /**
   * Determina a ordem de canais a serem usados para entrega
   * @param availableChannels Canais disponíveis para uso
   * @param preferences Preferências de notificação do destinatário
   * @param options Opções de notificação
   */
  private getPrioritizedChannels(
    availableChannels: NotificationChannel[],
    preferences: NotificationPreferences | null,
    options?: NotificationOrchestratorOptions
  ): NotificationChannel[] {
    // Se a opção de entregar em todos os canais estiver habilitada
    if (options?.deliverToAllChannels) {
      return availableChannels;
    }
    
    // Priorizar canais com base nas opções fornecidas
    if (options?.preferredChannels && options.preferredChannels.length > 0) {
      // Filtrar apenas canais disponíveis
      const filteredChannels = options.preferredChannels.filter(
        channel => availableChannels.includes(channel)
      );
      
      // Se houver canais disponíveis entre os preferenciais, usar eles
      if (filteredChannels.length > 0) {
        return filteredChannels;
      }
    }
    
    // Usar preferências do destinatário se disponíveis
    if (preferences && preferences.preferredChannels && preferences.preferredChannels.length > 0) {
      // Filtrar apenas canais disponíveis
      const filteredChannels = preferences.preferredChannels.filter(
        channel => availableChannels.includes(channel)
      );
      
      // Se houver canais disponíveis entre os preferenciais, usar eles
      if (filteredChannels.length > 0) {
        return filteredChannels;
      }
    }
    
    // Ordem padrão de preferência se nenhuma outra for especificada
    const defaultOrder = [
      NotificationChannel.PUSH, 
      NotificationChannel.SMS, 
      NotificationChannel.EMAIL,
      NotificationChannel.IN_APP,
      NotificationChannel.WEBHOOK
    ];
    
    return defaultOrder.filter(channel => availableChannels.includes(channel));
  }
  
  /**
   * Envia uma notificação para um único destinatário
   * @param recipient Destinatário da notificação
   * @param content Conteúdo da notificação
   * @param event Evento relacionado à notificação
   * @param options Opções de notificação
   */
  async sendNotification(
    recipient: NotificationRecipient,
    content: NotificationContent,
    event?: BaseEvent,
    options?: NotificationOrchestratorOptions
  ): Promise<DeliveryAttemptResult> {
    const notificationId = uuidv4();
    const attemptId = uuidv4();
    const timestamp = new Date();
    
    this.logger.info(`Enviando notificação ${notificationId} para ${recipient.id}`, {
      notificationId,
      recipientId: recipient.id,
      eventId: event?.eventId,
      eventType: event?.code,
      content: { title: content.title }
    });
    
    // Verificar canais disponíveis para este destinatário
    const availableChannels = await this.getAvailableChannels(recipient, options);
    
    if (availableChannels.length === 0) {
      const errorMsg = `Nenhum canal de notificação disponível para o destinatário ${recipient.id}`;
      this.logger.error(errorMsg, { notificationId, recipientId: recipient.id });
      
      return {
        attemptId,
        notificationId,
        timestamp,
        channelResults: [],
        status: DeliveryStatus.FAILED,
        successfulChannels: [],
        failedChannels: [],
        errorMessage: errorMsg
      };
    }
    
    // Buscar preferências do destinatário
    let preferences: NotificationPreferences | null = null;
    try {
      preferences = await this.recipientRepository.getNotificationPreferences(recipient.id);
    } catch (error) {
      this.logger.warn(`Falha ao obter preferências de notificação para ${recipient.id}`, {
        notificationId,
        recipientId: recipient.id,
        error
      });
      // Continuar sem as preferências
    }
    
    // Determinar ordem de canais para entrega
    const prioritizedChannels = this.getPrioritizedChannels(
      availableChannels,
      preferences,
      options
    );
    
    // Definir estratégia de entrega
    const strategy = options?.deliveryStrategy || DeliveryStrategy.SEQUENTIAL;
    
    // Inicializar resultados
    const channelResults: ChannelDeliveryResult[] = [];
    const successfulChannels: NotificationChannel[] = [];
    const failedChannels: NotificationChannel[] = [];
    
    try {
      // Registrar a notificação no repositório antes de enviar
      await this.notificationRepository.createNotification({
        id: notificationId,
        recipientId: recipient.id,
        content,
        eventId: event?.eventId,
        eventType: event?.code,
        status: DeliveryStatus.PROCESSING,
        createdAt: timestamp,
        channels: prioritizedChannels,
        transactionId: options?.transactionId,
        metadata: options?.metadata,
        priority: options?.priority || 1,
        scheduleFor: options?.scheduleFor
      });
      
      // Envio de notificação baseado na estratégia selecionada
      if (strategy === DeliveryStrategy.SEQUENTIAL) {
        // Envio sequencial - tenta um canal de cada vez até sucesso
        for (const channel of prioritizedChannels) {
          const result = await this.sendThroughChannel(channel, recipient, content, event, notificationId);
          channelResults.push({ ...result, channel });
          
          if (result.success) {
            successfulChannels.push(channel);
            // Se não estiver configurado para enviar em todos os canais, para no primeiro sucesso
            if (!options?.deliverToAllChannels) {
              break;
            }
          } else {
            failedChannels.push(channel);
          }
        }
      } else if (strategy === DeliveryStrategy.PARALLEL) {
        // Envio paralelo - tenta todos os canais simultaneamente
        const sendPromises = prioritizedChannels.map(async channel => {
          const result = await this.sendThroughChannel(channel, recipient, content, event, notificationId);
          return { result, channel };
        });
        
        const results = await Promise.all(sendPromises);
        
        for (const { result, channel } of results) {
          channelResults.push({ ...result, channel });
          
          if (result.success) {
            successfulChannels.push(channel);
          } else {
            failedChannels.push(channel);
          }
        }
      }
      
      // Determinar status final da tentativa
      let status: DeliveryStatus;
      let errorMessage: string | undefined;
      
      if (successfulChannels.length > 0) {
        status = DeliveryStatus.DELIVERED;
      } else {
        status = DeliveryStatus.FAILED;
        errorMessage = `Falha ao entregar notificação em todos os canais disponíveis: ${failedChannels.join(', ')}`;
        this.logger.error(errorMessage, { notificationId });
      }
      
      // Registrar resultado da tentativa
      const attemptResult: DeliveryAttemptResult = {
        attemptId,
        notificationId,
        timestamp,
        channelResults,
        status,
        successfulChannels,
        failedChannels,
        errorMessage
      };
      
      // Atualizar status da notificação no repositório
      await this.notificationRepository.updateNotificationStatus(
        notificationId,
        status,
        { attemptId, timestamp, channelResults }
      );
      
      return attemptResult;
    } catch (error) {
      const errorMsg = `Erro ao processar notificação: ${error}`;
      this.logger.error(errorMsg, { notificationId, error });
      
      // Atualizar status da notificação no repositório
      await this.notificationRepository.updateNotificationStatus(
        notificationId,
        DeliveryStatus.FAILED,
        { error: errorMsg }
      );
      
      return {
        attemptId,
        notificationId,
        timestamp,
        channelResults,
        status: DeliveryStatus.FAILED,
        successfulChannels,
        failedChannels,
        errorMessage: errorMsg
      };
    }
  }
  
  /**
   * Envia uma notificação através de um canal específico
   * @param channel Canal de notificação
   * @param recipient Destinatário da notificação
   * @param content Conteúdo da notificação
   * @param event Evento relacionado à notificação
   * @param notificationId ID da notificação
   */
  private async sendThroughChannel(
    channel: NotificationChannel,
    recipient: NotificationRecipient,
    content: NotificationContent,
    event?: BaseEvent,
    notificationId?: string
  ): Promise<NotificationResult> {
    const adapter = this.adapters.get(channel);
    
    if (!adapter) {
      return {
        success: false,
        notificationId,
        errorMessage: `Adaptador não encontrado para o canal ${channel}`,
        errorCode: 'ADAPTER_NOT_FOUND',
        timestamp: new Date()
      };
    }
    
    try {
      this.logger.debug(`Enviando notificação por ${channel}`, {
        recipientId: recipient.id,
        channel,
        notificationId
      });
      
      const result = await adapter.send(recipient, content, event, {
        notificationId,
        priority: 'NORMAL'
      });
      
      if (result.success) {
        this.logger.info(`Notificação enviada com sucesso por ${channel}`, {
          recipientId: recipient.id,
          channel,
          notificationId
        });
      } else {
        this.logger.warn(`Falha ao enviar notificação por ${channel}: ${result.errorMessage}`, {
          recipientId: recipient.id,
          channel,
          notificationId,
          errorCode: result.errorCode
        });
      }
      
      return result;
    } catch (error) {
      const errorMsg = `Erro ao enviar notificação por ${channel}: ${error}`;
      this.logger.error(errorMsg, {
        recipientId: recipient.id,
        channel,
        notificationId,
        error
      });
      
      return {
        success: false,
        notificationId,
        errorMessage: errorMsg,
        errorCode: 'CHANNEL_SEND_ERROR',
        timestamp: new Date()
      };
    }
  }
  
  /**
   * Envia uma notificação para múltiplos destinatários
   * @param recipients Lista de destinatários
   * @param content Conteúdo da notificação
   * @param event Evento relacionado à notificação
   * @param options Opções de notificação
   */
  async sendBulkNotification(
    recipients: NotificationRecipient[],
    content: NotificationContent,
    event?: BaseEvent,
    options?: NotificationOrchestratorOptions
  ): Promise<DeliveryAttemptResult[]> {
    const results: DeliveryAttemptResult[] = [];
    const batchId = uuidv4();
    
    this.logger.info(`Iniciando envio em lote de notificações para ${recipients.length} destinatários`, {
      batchId,
      recipientsCount: recipients.length,
      eventId: event?.eventId,
      eventType: event?.code
    });
    
    // Se a estratégia for paralela, envia todas as notificações simultaneamente
    if (options?.deliveryStrategy === DeliveryStrategy.PARALLEL) {
      const sendPromises = recipients.map(recipient => 
        this.sendNotification(recipient, content, event, {
          ...options,
          metadata: {
            ...(options.metadata || {}),
            batchId
          }
        })
      );
      
      return await Promise.all(sendPromises);
    }
    
    // Caso contrário, envia sequencialmente
    for (const recipient of recipients) {
      const result = await this.sendNotification(recipient, content, event, {
        ...options,
        metadata: {
          ...(options?.metadata || {}),
          batchId
        }
      });
      
      results.push(result);
    }
    
    return results;
  }
  
  /**
   * Cancela uma notificação agendada
   * @param notificationId ID da notificação
   */
  async cancelNotification(notificationId: string): Promise<boolean> {
    try {
      // Buscar informações da notificação no repositório
      const notification = await this.notificationRepository.getNotification(notificationId);
      
      if (!notification) {
        this.logger.warn(`Notificação ${notificationId} não encontrada para cancelamento`);
        return false;
      }
      
      // Se já foi entregue ou expirou, não pode ser cancelada
      if (notification.status === DeliveryStatus.DELIVERED || 
          notification.status === DeliveryStatus.EXPIRED) {
        this.logger.warn(`Notificação ${notificationId} não pode ser cancelada no estado ${notification.status}`);
        return false;
      }
      
      // Atualizar status da notificação para cancelada
      await this.notificationRepository.updateNotificationStatus(
        notificationId,
        DeliveryStatus.CANCELLED,
        { cancelledAt: new Date() }
      );
      
      // Se estava agendada, cancelar nos adaptadores
      if (notification.scheduleFor && notification.scheduleFor > new Date()) {
        const channels = notification.channels || [];
        
        for (const channel of channels) {
          const adapter = this.adapters.get(channel);
          if (adapter) {
            await adapter.cancel(notificationId);
          }
        }
      }
      
      this.logger.info(`Notificação ${notificationId} cancelada com sucesso`);
      return true;
    } catch (error) {
      this.logger.error(`Erro ao cancelar notificação ${notificationId}: ${error}`);
      return false;
    }
  }
  
  /**
   * Obtém o status atual de uma notificação
   * @param notificationId ID da notificação
   */
  async getNotificationStatus(notificationId: string): Promise<{
    status: DeliveryStatus;
    details: any;
  }> {
    try {
      const notification = await this.notificationRepository.getNotification(notificationId);
      
      if (!notification) {
        throw new Error(`Notificação ${notificationId} não encontrada`);
      }
      
      return {
        status: notification.status,
        details: {
          recipientId: notification.recipientId,
          channels: notification.channels,
          createdAt: notification.createdAt,
          updatedAt: notification.updatedAt,
          attempts: notification.attempts
        }
      };
    } catch (error) {
      this.logger.error(`Erro ao obter status da notificação ${notificationId}: ${error}`);
      throw error;
    }
  }
  
  /**
   * Processa notificações expiradas e atualiza seus status
   */
  async processExpiredNotifications(): Promise<number> {
    try {
      const expiredNotifications = await this.notificationRepository.getExpiredNotifications();
      
      for (const notification of expiredNotifications) {
        await this.notificationRepository.updateNotificationStatus(
          notification.id,
          DeliveryStatus.EXPIRED,
          { expiredAt: new Date() }
        );
      }
      
      this.logger.info(`Processadas ${expiredNotifications.length} notificações expiradas`);
      return expiredNotifications.length;
    } catch (error) {
      this.logger.error(`Erro ao processar notificações expiradas: ${error}`);
      return 0;
    }
  }
}