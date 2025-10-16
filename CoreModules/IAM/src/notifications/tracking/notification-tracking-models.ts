/**
 * @file notification-tracking-models.ts
 * @description Modelos de dados para rastreamento de notificações
 * 
 * Define as estruturas de dados utilizadas pelo sistema de rastreamento
 * de notificações, permitindo o monitoramento detalhado do ciclo de vida
 * das notificações enviadas.
 */

import { NotificationChannel } from '../core/notification-channel';

/**
 * Status possíveis de uma notificação
 */
export enum NotificationStatus {
  SCHEDULED = 'SCHEDULED',       // Agendada para envio futuro
  SENDING = 'SENDING',           // Em processo de envio
  SENT = 'SENT',                 // Enviada com sucesso ao provedor
  DELIVERED = 'DELIVERED',       // Confirmação de entrega ao destinatário
  FAILED = 'FAILED',             // Falha no envio
  OPENED = 'OPENED',             // Aberta/visualizada pelo destinatário
  CLICKED = 'CLICKED',           // Links clicados pelo destinatário
  REPLIED = 'REPLIED',           // Destinatário respondeu (quando aplicável)
  BOUNCED = 'BOUNCED',           // Retornada por problemas de entrega
  REJECTED = 'REJECTED',         // Rejeitada pelo provedor ou destinatário
  BLOCKED = 'BLOCKED',           // Bloqueada por filtros ou políticas
  SPAM = 'SPAM',                 // Marcada como spam
  EXPIRED = 'EXPIRED',           // Expirou sem ser entregue ou após período válido
  CANCELLED = 'CANCELLED',       // Cancelada antes do envio
  UNKNOWN = 'UNKNOWN'            // Status desconhecido
}

/**
 * Interface para eventos de rastreamento de notificações
 */
export interface NotificationTrackingEvent {
  /**
   * ID único do evento de rastreamento
   */
  id: string;
  
  /**
   * ID da notificação relacionada
   */
  notificationId: string;
  
  /**
   * Tipo do evento
   */
  eventType: 'send' | 'deliver' | 'open' | 'click' | 'bounce' | 'spam' | 
              'reject' | 'fail' | 'expire' | 'reply' | 'cancel' | 'status_change';
  
  /**
   * Timestamp do evento
   */
  timestamp: Date;
  
  /**
   * ID do destinatário
   */
  recipientId: string;
  
  /**
   * Canal utilizado
   */
  channel: NotificationChannel;
  
  /**
   * Provedor de serviço utilizado
   */
  provider?: string;
  
  /**
   * Status resultante
   */
  status: NotificationStatus;
  
  /**
   * Detalhes específicos do evento (depende do tipo)
   */
  details?: {
    /**
     * URL clicada (para eventos de clique)
     */
    url?: string;
    
    /**
     * Agente de usuário (navegador, app)
     */
    userAgent?: string;
    
    /**
     * Endereço IP
     */
    ipAddress?: string;
    
    /**
     * Geolocalização (país/região)
     */
    geoLocation?: string;
    
    /**
     * Dispositivo utilizado
     */
    device?: string;
    
    /**
     * Código de erro
     */
    errorCode?: string;
    
    /**
     * Mensagem de erro ou detalhes adicionais
     */
    errorMessage?: string;
    
    /**
     * ID de resposta (webhook, id externo)
     */
    responseId?: string;
    
    /**
     * Tempo de resposta (ms)
     */
    responseTime?: number;
    
    /**
     * Dados adicionais específicos do provedor
     */
    providerData?: Record<string, any>;
  };
  
  /**
   * Contexto do evento
   */
  context?: {
    /**
     * ID do tenant
     */
    tenantId?: string;
    
    /**
     * Ambiente (dev, staging, prod)
     */
    environment?: string;
    
    /**
     * Metadados adicionais de contexto
     */
    metadata?: Record<string, any>;
  };
}

/**
 * Interface para resumo de rastreamento de notificação
 */
export interface NotificationTrackingSummary {
  /**
   * ID da notificação
   */
  notificationId: string;
  
  /**
   * ID do destinatário
   */
  recipientId: string;
  
  /**
   * Canal utilizado
   */
  channel: NotificationChannel;
  
  /**
   * Status atual
   */
  status: NotificationStatus;
  
  /**
   * Data de criação
   */
  createdAt: Date;
  
  /**
   * Data de envio
   */
  sentAt?: Date;
  
  /**
   * Data de entrega
   */
  deliveredAt?: Date;
  
  /**
   * Data da primeira abertura
   */
  firstOpenedAt?: Date;
  
  /**
   * Data da última abertura
   */
  lastOpenedAt?: Date;
  
  /**
   * Total de aberturas
   */
  openCount: number;
  
  /**
   * Data do primeiro clique
   */
  firstClickedAt?: Date;
  
  /**
   * Data do último clique
   */
  lastClickedAt?: Date;
  
  /**
   * Total de cliques
   */
  clickCount: number;
  
  /**
   * URLs clicadas
   */
  clickedUrls?: string[];
  
  /**
   * Data de resposta (se aplicável)
   */
  repliedAt?: Date;
  
  /**
   * Data de falha (se aplicável)
   */
  failedAt?: Date;
  
  /**
   * Razão da falha (se aplicável)
   */
  failureReason?: string;
  
  /**
   * Tempo total de entrega (ms)
   */
  deliveryTimeMs?: number;
  
  /**
   * Origem da notificação
   */
  source?: string;
  
  /**
   * Módulo de origem
   */
  module?: string;
  
  /**
   * Tipo de evento que originou a notificação
   */
  eventType?: string;
  
  /**
   * Tags associadas
   */
  tags?: string[];
  
  /**
   * Metadados específicos de rastreamento
   */
  metadata?: Record<string, any>;
}

/**
 * Interface para estatísticas agregadas de notificações
 */
export interface NotificationAggregateStats {
  /**
   * Período de tempo das estatísticas
   */
  period: {
    /**
     * Data de início
     */
    start: Date;
    
    /**
     * Data de fim
     */
    end: Date;
  };
  
  /**
   * Filtros aplicados
   */
  filters?: {
    /**
     * Canal
     */
    channel?: NotificationChannel;
    
    /**
     * Módulo de origem
     */
    module?: string;
    
    /**
     * Tipo de evento
     */
    eventType?: string;
    
    /**
     * Tags
     */
    tags?: string[];
    
    /**
     * ID do tenant
     */
    tenantId?: string;
  };
  
  /**
   * Totais
   */
  totals: {
    /**
     * Total de notificações
     */
    notifications: number;
    
    /**
     * Total enviado
     */
    sent: number;
    
    /**
     * Total entregue
     */
    delivered: number;
    
    /**
     * Total de falhas
     */
    failed: number;
    
    /**
     * Total aberto
     */
    opened: number;
    
    /**
     * Total de cliques
     */
    clicked: number;
    
    /**
     * Total de respostas
     */
    replied: number;
    
    /**
     * Total de rejeições
     */
    rejected: number;
  };
  
  /**
   * Taxas calculadas
   */
  rates: {
    /**
     * Taxa de entrega (entregues/enviados)
     */
    deliveryRate: number;
    
    /**
     * Taxa de abertura (abertos/entregues)
     */
    openRate: number;
    
    /**
     * Taxa de clique (clicados/abertos)
     */
    clickRate: number;
    
    /**
     * Taxa de rejeição (rejeitados/enviados)
     */
    rejectRate: number;
    
    /**
     * Taxa de falha (falhas/enviados)
     */
    failureRate: number;
    
    /**
     * Taxa de resposta (respondidos/entregues)
     */
    replyRate: number;
  };
  
  /**
   * Tempos médios (ms)
   */
  averageTimes: {
    /**
     * Tempo médio de entrega
     */
    deliveryTimeMs: number;
    
    /**
     * Tempo médio para primeira abertura
     */
    timeToFirstOpenMs: number;
    
    /**
     * Tempo médio para primeiro clique
     */
    timeToFirstClickMs: number;
    
    /**
     * Tempo médio para resposta
     */
    timeToReplyMs: number;
  };
  
  /**
   * Distribuição por canal
   */
  byChannel: Record<NotificationChannel, {
    count: number;
    deliveryRate: number;
    openRate: number;
    clickRate: number;
  }>;
  
  /**
   * Distribuição por hora do dia
   */
  byHourOfDay?: Record<number, number>;
  
  /**
   * Distribuição por dia da semana
   */
  byDayOfWeek?: Record<number, number>;
}

/**
 * Interface para repositório de dados de rastreamento
 */
export interface NotificationTrackingRepository {
  /**
   * Salva um evento de rastreamento
   * @param event Evento de rastreamento
   */
  saveTrackingEvent(event: NotificationTrackingEvent): Promise<void>;
  
  /**
   * Obtém eventos de rastreamento para uma notificação
   * @param notificationId ID da notificação
   */
  getTrackingEvents(notificationId: string): Promise<NotificationTrackingEvent[]>;
  
  /**
   * Obtém resumo de rastreamento para uma notificação
   * @param notificationId ID da notificação
   */
  getTrackingSummary(notificationId: string): Promise<NotificationTrackingSummary | null>;
  
  /**
   * Atualiza o status de uma notificação
   * @param notificationId ID da notificação
   * @param status Novo status
   * @param details Detalhes adicionais
   */
  updateNotificationStatus(
    notificationId: string,
    status: NotificationStatus,
    details?: Record<string, any>
  ): Promise<void>;
  
  /**
   * Obtém estatísticas agregadas
   * @param filter Filtros para as estatísticas
   */
  getAggregateStats(filter: {
    startDate: Date;
    endDate: Date;
    channel?: NotificationChannel;
    module?: string;
    eventType?: string;
    tags?: string[];
    tenantId?: string;
  }): Promise<NotificationAggregateStats>;
}