/**
 * @file delivery-tracker.ts
 * @description Sistema de rastreamento e confirmação de entrega de notificações
 * 
 * Este módulo implementa o sistema de rastreamento que monitora o estado das 
 * notificações, gerencia confirmações de entrega, e fornece insights sobre
 * o status atual das notificações em diferentes canais.
 */

import { v4 as uuidv4 } from 'uuid';
import { Logger } from '../../../infrastructure/observability/logger';
import { NotificationChannel } from './notification-channel';
import { DeliveryStatus } from './notification-orchestrator';
import { NotificationRepository } from '../repositories/notification-repository';
import { EventRepository } from '../repositories/event-repository';

/**
 * Status detalhado de rastreamento da notificação
 */
export enum TrackingStatus {
  /**
   * Notificação foi gerada mas ainda não foi processada
   */
  QUEUED = 'QUEUED',
  
  /**
   * Notificação em processo de envio
   */
  SENDING = 'SENDING',
  
  /**
   * Notificação enviada pelo provedor
   */
  SENT = 'SENT',
  
  /**
   * Recebimento confirmado pelo dispositivo/servidor do destinatário
   */
  DELIVERED = 'DELIVERED',
  
  /**
   * Notificação foi aberta/visualizada pelo destinatário
   */
  OPENED = 'OPENED',
  
  /**
   * Usuário interagiu com a notificação (clique, resposta)
   */
  INTERACTED = 'INTERACTED',
  
  /**
   * Falha permanente na entrega
   */
  FAILED = 'FAILED',
  
  /**
   * Falha temporária (possível reenvio)
   */
  BOUNCED = 'BOUNCED',
  
  /**
   * Notificação bloqueada (spam, política do provedor, etc.)
   */
  BLOCKED = 'BLOCKED',
  
  /**
   * Rejeitada pelo destinatário (opt-out, regras de filtro, etc.)
   */
  REJECTED = 'REJECTED',
  
  /**
   * Notificação expirou sem confirmação
   */
  EXPIRED = 'EXPIRED',
  
  /**
   * Cancelada antes da entrega
   */
  CANCELLED = 'CANCELLED',
  
  /**
   * Status desconhecido ou não rastreável
   */
  UNKNOWN = 'UNKNOWN'
}

/**
 * Detalhes do evento de rastreamento
 */
export interface TrackingEvent {
  /**
   * ID único do evento de rastreamento
   */
  trackingId: string;
  
  /**
   * ID da notificação relacionada
   */
  notificationId: string;
  
  /**
   * Timestamp do evento
   */
  timestamp: Date;
  
  /**
   * Status do rastreamento
   */
  status: TrackingStatus;
  
  /**
   * Canal onde o evento ocorreu
   */
  channel: NotificationChannel;
  
  /**
   * Tipo de evento
   */
  eventType: 'STATUS_CHANGE' | 'DELIVERY_RECEIPT' | 'READ_RECEIPT' | 'INTERACTION' | 'ERROR';
  
  /**
   * Dados adicionais específicos do evento
   */
  metadata?: Record<string, any>;
  
  /**
   * IP ou identificação do dispositivo que gerou o evento
   */
  deviceInfo?: string;
  
  /**
   * Localização aproximada (quando disponível)
   */
  location?: {
    country?: string;
    region?: string;
    city?: string;
    coordinates?: [number, number]; // [longitude, latitude]
  };
  
  /**
   * Informações do agente (browser, app, etc.)
   */
  userAgent?: string;
}

/**
 * Status consolidado de rastreamento por canal
 */
export interface ChannelTrackingStatus {
  /**
   * Canal de notificação
   */
  channel: NotificationChannel;
  
  /**
   * Status atual no canal
   */
  status: TrackingStatus;
  
  /**
   * Timestamp da última atualização
   */
  lastUpdated: Date;
  
  /**
   * Histórico de eventos
   */
  events: TrackingEvent[];
}

/**
 * Status consolidado de rastreamento da notificação
 */
export interface NotificationTrackingStatus {
  /**
   * ID da notificação
   */
  notificationId: string;
  
  /**
   * ID do destinatário
   */
  recipientId: string;
  
  /**
   * Status atual de entrega
   */
  deliveryStatus: DeliveryStatus;
  
  /**
   * Status detalhado por canal
   */
  channelStatus: ChannelTrackingStatus[];
  
  /**
   * Timestamp da criação da notificação
   */
  createdAt: Date;
  
  /**
   * Timestamp da última atualização
   */
  lastUpdated: Date;
  
  /**
   * Status geral de rastreamento (o mais avançado entre os canais)
   */
  overallStatus: TrackingStatus;
  
  /**
   * Canal com status mais avançado
   */
  primaryChannel?: NotificationChannel;
  
  /**
   * Timestamp estimado para expiração
   */
  expiresAt?: Date;
}

/**
 * Tipo de webhook para rastreamento
 */
export enum WebhookType {
  /**
   * Confirmação de entrega (recebido pelo dispositivo/servidor)
   */
  DELIVERY_RECEIPT = 'DELIVERY_RECEIPT',
  
  /**
   * Confirmação de leitura (visualizado pelo destinatário)
   */
  READ_RECEIPT = 'READ_RECEIPT',
  
  /**
   * Interação do usuário (clique, resposta)
   */
  INTERACTION = 'INTERACTION',
  
  /**
   * Relatório de falha na entrega
   */
  FAILURE_REPORT = 'FAILURE_REPORT',
  
  /**
   * Status temporário (em trânsito, pendente)
   */
  STATUS_UPDATE = 'STATUS_UPDATE'
}

/**
 * Dados do webhook de rastreamento
 */
export interface WebhookPayload {
  /**
   * ID da notificação
   */
  notificationId: string;
  
  /**
   * Canal de notificação
   */
  channel: NotificationChannel;
  
  /**
   * Tipo do webhook
   */
  type: WebhookType;
  
  /**
   * Status reportado
   */
  status: string;
  
  /**
   * Timestamp do evento (UTC)
   */
  timestamp: string;
  
  /**
   * ID do provedor (SMS, push, etc.)
   */
  providerId?: string;
  
  /**
   * Dados adicionais específicos do provedor
   */
  providerData?: Record<string, any>;
  
  /**
   * Informações do dispositivo
   */
  device?: Record<string, any>;
  
  /**
   * Erro reportado (se houver)
   */
  error?: {
    code: string;
    message: string;
    details?: any;
  };
}

/**
 * Classe para rastreamento de entrega de notificações
 */
export class DeliveryTracker {
  private logger = new Logger('DeliveryTracker');
  
  /**
   * Construtor
   * @param notificationRepository Repositório de notificações
   * @param eventRepository Repositório de eventos
   */
  constructor(
    private notificationRepository: NotificationRepository,
    private eventRepository: EventRepository
  ) {}
  
  /**
   * Registra um novo evento de rastreamento
   * @param event Evento de rastreamento
   */
  async trackEvent(event: Omit<TrackingEvent, 'trackingId'>): Promise<TrackingEvent> {
    const trackingId = uuidv4();
    const trackingEvent: TrackingEvent = {
      ...event,
      trackingId,
      timestamp: event.timestamp || new Date()
    };
    
    try {
      await this.notificationRepository.addTrackingEvent(trackingEvent);
      
      // Atualizar o status geral da notificação se necessário
      await this.updateNotificationStatus(event.notificationId, event.status, event.channel);
      
      this.logger.info(`Evento de rastreamento registrado: ${trackingId}`, {
        notificationId: event.notificationId,
        channel: event.channel,
        status: event.status,
        eventType: event.eventType
      });
      
      return trackingEvent;
    } catch (error) {
      this.logger.error(`Erro ao registrar evento de rastreamento: ${error}`, {
        notificationId: event.notificationId,
        channel: event.channel,
        error
      });
      throw error;
    }
  }
  
  /**
   * Atualiza o status de uma notificação com base em evento de rastreamento
   * @param notificationId ID da notificação
   * @param status Status de rastreamento
   * @param channel Canal da notificação
   */
  private async updateNotificationStatus(
    notificationId: string, 
    status: TrackingStatus,
    channel: NotificationChannel
  ): Promise<void> {
    try {
      // Obter notificação atual
      const notification = await this.notificationRepository.getNotification(notificationId);
      
      if (!notification) {
        this.logger.warn(`Notificação ${notificationId} não encontrada para atualização de status`);
        return;
      }
      
      // Converter o status de rastreamento para status de entrega
      let deliveryStatus = notification.status;
      
      switch (status) {
        case TrackingStatus.DELIVERED:
        case TrackingStatus.OPENED:
        case TrackingStatus.INTERACTED:
          deliveryStatus = DeliveryStatus.DELIVERED;
          break;
          
        case TrackingStatus.FAILED:
        case TrackingStatus.BLOCKED:
        case TrackingStatus.REJECTED:
          // Só atualiza para falha se não tiver sido entregue em nenhum outro canal
          if (notification.status !== DeliveryStatus.DELIVERED) {
            deliveryStatus = DeliveryStatus.FAILED;
          }
          break;
          
        case TrackingStatus.EXPIRED:
          deliveryStatus = DeliveryStatus.EXPIRED;
          break;
          
        case TrackingStatus.CANCELLED:
          deliveryStatus = DeliveryStatus.CANCELLED;
          break;
      }
      
      // Se o status mudou, atualizar
      if (deliveryStatus !== notification.status) {
        await this.notificationRepository.updateNotificationStatus(
          notificationId, 
          deliveryStatus,
          {
            trackingStatus: status,
            channel,
            updatedAt: new Date()
          }
        );
      } else {
        // Mesmo que o status principal não tenha mudado, atualizar o status de rastreamento
        await this.notificationRepository.updateTrackingStatus(
          notificationId,
          channel,
          status
        );
      }
    } catch (error) {
      this.logger.error(`Erro ao atualizar status da notificação ${notificationId}: ${error}`);
    }
  }
  
  /**
   * Processa um webhook de rastreamento de um provedor externo
   * @param payload Dados do webhook
   */
  async processWebhook(payload: WebhookPayload): Promise<TrackingEvent | null> {
    try {
      // Verificar se a notificação existe
      const notification = await this.notificationRepository.getNotification(payload.notificationId);
      
      if (!notification) {
        this.logger.warn(`Webhook recebido para notificação inexistente: ${payload.notificationId}`);
        return null;
      }
      
      // Converter status do provedor para TrackingStatus interno
      const status = this.mapProviderStatus(payload.status, payload.type, payload.channel);
      
      // Criar evento de rastreamento
      const trackingEvent: Omit<TrackingEvent, 'trackingId'> = {
        notificationId: payload.notificationId,
        timestamp: new Date(payload.timestamp),
        status,
        channel: payload.channel,
        eventType: this.mapWebhookTypeToEventType(payload.type),
        metadata: {
          providerId: payload.providerId,
          providerData: payload.providerData,
          providerStatus: payload.status,
          webhookType: payload.type
        },
        deviceInfo: payload.device ? JSON.stringify(payload.device) : undefined
      };
      
      // Registrar o evento
      return await this.trackEvent(trackingEvent);
    } catch (error) {
      this.logger.error(`Erro ao processar webhook de rastreamento: ${error}`, {
        notificationId: payload.notificationId,
        channel: payload.channel,
        type: payload.type,
        error
      });
      return null;
    }
  }
  
  /**
   * Mapeia o tipo de webhook para tipo de evento interno
   * @param webhookType Tipo do webhook
   */
  private mapWebhookTypeToEventType(webhookType: WebhookType): TrackingEvent['eventType'] {
    switch (webhookType) {
      case WebhookType.DELIVERY_RECEIPT:
        return 'DELIVERY_RECEIPT';
      case WebhookType.READ_RECEIPT:
        return 'READ_RECEIPT';
      case WebhookType.INTERACTION:
        return 'INTERACTION';
      case WebhookType.FAILURE_REPORT:
        return 'ERROR';
      case WebhookType.STATUS_UPDATE:
      default:
        return 'STATUS_CHANGE';
    }
  }
  
  /**
   * Converte status do provedor para TrackingStatus interno
   * @param providerStatus Status reportado pelo provedor
   * @param webhookType Tipo do webhook
   * @param channel Canal de notificação
   */
  private mapProviderStatus(
    providerStatus: string, 
    webhookType: WebhookType,
    channel: NotificationChannel
  ): TrackingStatus {
    // Mapeamento de status baseado no tipo de webhook
    switch (webhookType) {
      case WebhookType.DELIVERY_RECEIPT:
        return TrackingStatus.DELIVERED;
        
      case WebhookType.READ_RECEIPT:
        return TrackingStatus.OPENED;
        
      case WebhookType.INTERACTION:
        return TrackingStatus.INTERACTED;
        
      case WebhookType.FAILURE_REPORT:
        // Determinar o tipo de falha
        return this.mapFailureStatus(providerStatus, channel);
        
      case WebhookType.STATUS_UPDATE:
        // Status detalhado baseado no status do provedor
        return this.mapDetailedStatus(providerStatus, channel);
    }
  }
  
  /**
   * Mapeia um status de falha reportado para TrackingStatus interno
   * @param providerStatus Status reportado pelo provedor
   * @param channel Canal de notificação
   */
  private mapFailureStatus(providerStatus: string, channel: NotificationChannel): TrackingStatus {
    // Normalizar o status para comparação
    const normalizedStatus = providerStatus.toLowerCase();
    
    // Status que indicam falha permanente
    if (normalizedStatus.includes('invalid') ||
        normalizedStatus.includes('reject') ||
        normalizedStatus.includes('block') ||
        normalizedStatus.includes('permanent') ||
        normalizedStatus.includes('unsubscribe')) {
      return TrackingStatus.REJECTED;
    }
    
    // Status que indicam spam ou bloqueio
    if (normalizedStatus.includes('spam') ||
        normalizedStatus.includes('filter') ||
        normalizedStatus.includes('policy') ||
        normalizedStatus.includes('prohibited')) {
      return TrackingStatus.BLOCKED;
    }
    
    // Status que indicam falha temporária
    if (normalizedStatus.includes('tempfail') ||
        normalizedStatus.includes('defer') ||
        normalizedStatus.includes('bounce') ||
        normalizedStatus.includes('retry') ||
        normalizedStatus.includes('throttle') ||
        normalizedStatus.includes('quota')) {
      return TrackingStatus.BOUNCED;
    }
    
    // Padrão: falha genérica
    return TrackingStatus.FAILED;
  }
  
  /**
   * Mapeia um status detalhado reportado para TrackingStatus interno
   * @param providerStatus Status reportado pelo provedor
   * @param channel Canal de notificação
   */
  private mapDetailedStatus(providerStatus: string, channel: NotificationChannel): TrackingStatus {
    // Normalizar o status para comparação
    const normalizedStatus = providerStatus.toLowerCase();
    
    // Mapeamentos específicos por canal
    if (channel === NotificationChannel.EMAIL) {
      if (normalizedStatus.includes('deliver') || normalizedStatus.includes('sent')) {
        return TrackingStatus.DELIVERED;
      }
      if (normalizedStatus.includes('open') || normalizedStatus.includes('read')) {
        return TrackingStatus.OPENED;
      }
      if (normalizedStatus.includes('click') || normalizedStatus.includes('interact')) {
        return TrackingStatus.INTERACTED;
      }
    } else if (channel === NotificationChannel.SMS) {
      if (normalizedStatus.includes('deliver')) {
        return TrackingStatus.DELIVERED;
      }
      if (normalizedStatus.includes('sent')) {
        return TrackingStatus.SENT;
      }
      if (normalizedStatus.includes('queue')) {
        return TrackingStatus.QUEUED;
      }
    } else if (channel === NotificationChannel.PUSH) {
      if (normalizedStatus.includes('deliver') || normalizedStatus.includes('received')) {
        return TrackingStatus.DELIVERED;
      }
      if (normalizedStatus.includes('open') || normalizedStatus.includes('viewed')) {
        return TrackingStatus.OPENED;
      }
      if (normalizedStatus.includes('click') || normalizedStatus.includes('interact')) {
        return TrackingStatus.INTERACTED;
      }
      if (normalizedStatus.includes('sent') || normalizedStatus.includes('accept')) {
        return TrackingStatus.SENT;
      }
    }
    
    // Status genéricos para todos os canais
    if (normalizedStatus.includes('sending') || normalizedStatus.includes('process')) {
      return TrackingStatus.SENDING;
    }
    if (normalizedStatus.includes('queue') || normalizedStatus.includes('pending')) {
      return TrackingStatus.QUEUED;
    }
    if (normalizedStatus.includes('expire')) {
      return TrackingStatus.EXPIRED;
    }
    if (normalizedStatus.includes('cancel')) {
      return TrackingStatus.CANCELLED;
    }
    
    // Padrão: status desconhecido
    return TrackingStatus.UNKNOWN;
  }
  
  /**
   * Obtém o status consolidado de rastreamento de uma notificação
   * @param notificationId ID da notificação
   */
  async getTrackingStatus(notificationId: string): Promise<NotificationTrackingStatus> {
    try {
      // Obter a notificação
      const notification = await this.notificationRepository.getNotification(notificationId);
      
      if (!notification) {
        throw new Error(`Notificação ${notificationId} não encontrada`);
      }
      
      // Obter todos os eventos de rastreamento para esta notificação
      const trackingEvents = await this.notificationRepository.getTrackingEvents(notificationId);
      
      // Agrupar eventos por canal
      const eventsByChannel = new Map<NotificationChannel, TrackingEvent[]>();
      
      for (const event of trackingEvents) {
        const channelEvents = eventsByChannel.get(event.channel) || [];
        channelEvents.push(event);
        eventsByChannel.set(event.channel, channelEvents);
      }
      
      // Determinar status para cada canal
      const channelStatus: ChannelTrackingStatus[] = [];
      let overallStatus = TrackingStatus.QUEUED;
      let primaryChannel: NotificationChannel | undefined;
      let lastUpdated = notification.createdAt;
      
      for (const [channel, events] of eventsByChannel.entries()) {
        // Ordenar eventos por timestamp (mais recente primeiro)
        const sortedEvents = events.sort((a, b) => 
          b.timestamp.getTime() - a.timestamp.getTime()
        );
        
        // O status mais recente para o canal
        const currentStatus = sortedEvents[0]?.status || TrackingStatus.UNKNOWN;
        const channelLastUpdated = sortedEvents[0]?.timestamp || notification.createdAt;
        
        channelStatus.push({
          channel,
          status: currentStatus,
          lastUpdated: channelLastUpdated,
          events: sortedEvents
        });
        
        // Atualizar o timestamp global da última atualização
        if (channelLastUpdated > lastUpdated) {
          lastUpdated = channelLastUpdated;
        }
        
        // Determinar o status geral mais avançado
        if (this.isMoreAdvancedStatus(currentStatus, overallStatus)) {
          overallStatus = currentStatus;
          primaryChannel = channel;
        }
      }
      
      return {
        notificationId,
        recipientId: notification.recipientId,
        deliveryStatus: notification.status,
        channelStatus,
        createdAt: notification.createdAt,
        lastUpdated,
        overallStatus,
        primaryChannel,
        expiresAt: notification.expiresAt
      };
    } catch (error) {
      this.logger.error(`Erro ao obter status de rastreamento para ${notificationId}: ${error}`);
      throw error;
    }
  }
  
  /**
   * Verifica se um status é mais avançado que outro
   * @param status1 Primeiro status
   * @param status2 Segundo status
   */
  private isMoreAdvancedStatus(status1: TrackingStatus, status2: TrackingStatus): boolean {
    // Ordem de prioridade dos status
    const statusPriority = {
      [TrackingStatus.INTERACTED]: 7,
      [TrackingStatus.OPENED]: 6,
      [TrackingStatus.DELIVERED]: 5,
      [TrackingStatus.SENT]: 4,
      [TrackingStatus.SENDING]: 3,
      [TrackingStatus.QUEUED]: 2,
      [TrackingStatus.BOUNCED]: 1,
      [TrackingStatus.FAILED]: 0,
      [TrackingStatus.BLOCKED]: 0,
      [TrackingStatus.REJECTED]: 0,
      [TrackingStatus.EXPIRED]: 0,
      [TrackingStatus.CANCELLED]: 0,
      [TrackingStatus.UNKNOWN]: 0
    };
    
    return statusPriority[status1] > statusPriority[status2];
  }
  
  /**
   * Gera URLs de rastreamento para uma notificação
   * @param notificationId ID da notificação
   * @param channel Canal de notificação
   * @param baseUrl URL base para rastreamento
   */
  generateTrackingUrls(
    notificationId: string, 
    channel: NotificationChannel,
    baseUrl: string
  ): {
    openUrl?: string;
    clickUrl?: string;
  } {
    // Gerar URLs apenas para canais que suportam rastreamento
    if (![NotificationChannel.EMAIL, NotificationChannel.SMS, NotificationChannel.PUSH].includes(channel)) {
      return {};
    }
    
    const trackingToken = this.generateTrackingToken(notificationId, channel);
    
    const openUrl = `${baseUrl}/track/open/${notificationId}/${channel}/${trackingToken}`;
    const clickUrl = `${baseUrl}/track/click/${notificationId}/${channel}/${trackingToken}`;
    
    return { openUrl, clickUrl };
  }
  
  /**
   * Gera um token de rastreamento para uma notificação
   * @param notificationId ID da notificação
   * @param channel Canal de notificação
   */
  private generateTrackingToken(notificationId: string, channel: NotificationChannel): string {
    // Em uma implementação real, isso seria um token criptografado e verificável
    // Para este exemplo, usamos um formato simples baseado em UUID
    const uuid = uuidv4();
    return `${uuid.substring(0, 8)}-${channel.substring(0, 3).toUpperCase()}`;
  }
  
  /**
   * Valida um token de rastreamento para uma notificação
   * @param notificationId ID da notificação
   * @param channel Canal de notificação
   * @param token Token de rastreamento
   */
  async validateTrackingToken(
    notificationId: string,
    channel: NotificationChannel,
    token: string
  ): Promise<boolean> {
    // Em uma implementação real, verificaríamos a validade do token
    // Para este exemplo, verificamos apenas se a notificação existe
    try {
      const notification = await this.notificationRepository.getNotification(notificationId);
      return !!notification;
    } catch (error) {
      this.logger.error(`Erro ao validar token de rastreamento: ${error}`, {
        notificationId,
        channel,
        token
      });
      return false;
    }
  }
  
  /**
   * Registra uma abertura de notificação
   * @param notificationId ID da notificação
   * @param channel Canal de notificação
   * @param metadata Metadados adicionais
   */
  async trackOpen(
    notificationId: string, 
    channel: NotificationChannel,
    metadata?: Record<string, any>
  ): Promise<TrackingEvent> {
    return this.trackEvent({
      notificationId,
      status: TrackingStatus.OPENED,
      channel,
      eventType: 'READ_RECEIPT',
      timestamp: new Date(),
      metadata
    });
  }
  
  /**
   * Registra um clique/interação com a notificação
   * @param notificationId ID da notificação
   * @param channel Canal de notificação
   * @param metadata Metadados adicionais
   */
  async trackClick(
    notificationId: string, 
    channel: NotificationChannel,
    metadata?: Record<string, any>
  ): Promise<TrackingEvent> {
    return this.trackEvent({
      notificationId,
      status: TrackingStatus.INTERACTED,
      channel,
      eventType: 'INTERACTION',
      timestamp: new Date(),
      metadata
    });
  }
}