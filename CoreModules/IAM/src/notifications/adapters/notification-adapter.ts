/**
 * @file notification-adapter.ts
 * @description Define a interface base para adaptadores de canais de notificação
 * 
 * Este módulo implementa o padrão Adapter para permitir o envio de notificações
 * através de diferentes canais (email, SMS, push, etc.) com uma interface unificada,
 * garantindo extensibilidade e facilidade de integração com novos canais.
 */

import { BaseEvent } from '../core/base-event';
import { NotificationChannel, ChannelConfig } from '../core/notification-channel';

/**
 * Interface para o resultado de uma tentativa de envio de notificação
 */
export interface NotificationResult {
  /**
   * Indica se o envio foi bem-sucedido
   */
  success: boolean;
  
  /**
   * Identificador da notificação enviada (quando bem-sucedida)
   */
  notificationId?: string;
  
  /**
   * Mensagem de erro (quando houve falha)
   */
  errorMessage?: string;
  
  /**
   * Código de erro (quando houve falha)
   */
  errorCode?: string;
  
  /**
   * Detalhes adicionais sobre o envio
   */
  details?: Record<string, any>;
  
  /**
   * Timestamp do envio
   */
  timestamp: Date;
}

/**
 * Interface para destinatário de notificação
 */
export interface NotificationRecipient {
  /**
   * Identificador do destinatário
   */
  id: string;
  
  /**
   * Tipo do destinatário (usuário, sistema, etc.)
   */
  type: 'USER' | 'SYSTEM' | 'SERVICE' | 'GROUP' | 'ROLE' | 'EXTERNAL';
  
  /**
   * Nome do destinatário (opcional)
   */
  name?: string;
  
  /**
   * Endereços para envio por canal
   * Mapa de canal para endereço (email, número de telefone, token de dispositivo, etc.)
   */
  addresses?: Map<NotificationChannel, string[]>;
  
  /**
   * Preferências de notificação
   */
  preferences?: {
    /**
     * Canais preferidos em ordem de prioridade
     */
    preferredChannels?: NotificationChannel[];
    
    /**
     * Canais desativados para este destinatário
     */
    disabledChannels?: NotificationChannel[];
    
    /**
     * Horários de entrega preferidos (formato "HH:MM")
     */
    deliveryTimes?: {
      start: string;
      end: string;
      timeZone?: string;
    };
    
    /**
     * Dias da semana preferidos (0 = domingo, 6 = sábado)
     */
    deliveryDays?: number[];
    
    /**
     * Formato preferido para notificações
     */
    format?: 'TEXT' | 'HTML' | 'MARKDOWN';
    
    /**
     * Idioma preferido
     */
    language?: string;
  };
  
  /**
   * Metadados adicionais do destinatário
   */
  metadata?: Record<string, any>;
}

/**
 * Interface para conteúdo de notificação
 */
export interface NotificationContent {
  /**
   * Título da notificação
   */
  title?: string;
  
  /**
   * Corpo da mensagem
   */
  body: string;
  
  /**
   * Formato do conteúdo
   */
  format?: 'TEXT' | 'HTML' | 'MARKDOWN' | 'JSON';
  
  /**
   * Dados para template (se aplicável)
   */
  templateData?: Record<string, any>;
  
  /**
   * ID do template (se aplicável)
   */
  templateId?: string;
  
  /**
   * Anexos (se aplicável)
   */
  attachments?: Array<{
    filename: string;
    content: string | Buffer;
    contentType: string;
  }>;
  
  /**
   * URLs de recursos relacionados
   */
  resourceUrls?: Array<{
    url: string;
    type: string;
    description?: string;
  }>;
  
  /**
   * Ações disponíveis para a notificação
   */
  actions?: Array<{
    id: string;
    label: string;
    url?: string;
    actionType: 'LINK' | 'BUTTON' | 'REPLY' | 'DISMISS' | 'CUSTOM';
    payload?: Record<string, any>;
  }>;
  
  /**
   * Informações de localização (se aplicável)
   */
  location?: {
    latitude: number;
    longitude: number;
    name?: string;
    address?: string;
  };
}

/**
 * Interface base para todas as opções de notificação
 */
export interface BaseNotificationOptions {
  /**
   * ID único da notificação
   */
  notificationId?: string;
  
  /**
   * Prioridade de envio
   */
  priority?: 'LOWEST' | 'LOW' | 'NORMAL' | 'HIGH' | 'HIGHEST';
  
  /**
   * Tempo de vida da notificação em segundos
   */
  ttlSeconds?: number;
  
  /**
   * Política de novas tentativas
   */
  retryPolicy?: {
    maxRetries: number;
    initialDelayMs: number;
    backoffMultiplier: number;
    maxDelayMs: number;
  };
  
  /**
   * Dados de rastreamento
   */
  tracking?: {
    /**
     * ID de correlação para rastreamento entre sistemas
     */
    correlationId?: string;
    
    /**
     * ID da transação
     */
    transactionId?: string;
    
    /**
     * Tags para categorização e filtragem
     */
    tags?: string[];
    
    /**
     * Metadados adicionais de rastreamento
     */
    metadata?: Record<string, any>;
  };
  
  /**
   * Opções de agendamento
   */
  scheduling?: {
    /**
     * Enviar em horário específico
     */
    sendAt?: Date;
    
    /**
     * Enviar recorrentemente (formato cron)
     */
    cronExpression?: string;
    
    /**
     * Fuso horário para agendamento
     */
    timeZone?: string;
  };
  
  /**
   * Manipuladores de eventos
   */
  eventHandlers?: {
    /**
     * Função chamada quando a notificação é entregue
     */
    onDelivered?: (result: NotificationResult) => void;
    
    /**
     * Função chamada quando há falha na entrega
     */
    onFailed?: (result: NotificationResult) => void;
    
    /**
     * Função chamada quando a notificação é aberta
     */
    onOpened?: (data: any) => void;
    
    /**
     * Função chamada quando uma ação é executada na notificação
     */
    onAction?: (actionId: string, data: any) => void;
  };
}

/**
 * Interface base para adaptadores de notificação
 */
export interface NotificationAdapter {
  /**
   * Tipo de canal que este adaptador suporta
   */
  readonly channelType: NotificationChannel;
  
  /**
   * Verifica se o adaptador está inicializado e pronto para uso
   */
  isReady(): Promise<boolean>;
  
  /**
   * Inicializa o adaptador com a configuração
   * @param config Configuração do canal
   */
  initialize(config: ChannelConfig): Promise<void>;
  
  /**
   * Envia uma notificação para um único destinatário
   * @param recipient Destinatário da notificação
   * @param content Conteúdo da notificação
   * @param event Evento que originou a notificação (opcional)
   * @param options Opções adicionais para o envio
   */
  send(
    recipient: NotificationRecipient,
    content: NotificationContent,
    event?: BaseEvent,
    options?: BaseNotificationOptions
  ): Promise<NotificationResult>;
  
  /**
   * Envia uma notificação para múltiplos destinatários
   * @param recipients Lista de destinatários
   * @param content Conteúdo da notificação
   * @param event Evento que originou a notificação (opcional)
   * @param options Opções adicionais para o envio
   */
  sendBulk(
    recipients: NotificationRecipient[],
    content: NotificationContent,
    event?: BaseEvent,
    options?: BaseNotificationOptions
  ): Promise<NotificationResult[]>;
  
  /**
   * Cancela uma notificação agendada
   * @param notificationId ID da notificação
   */
  cancel(notificationId: string): Promise<boolean>;
  
  /**
   * Verifica o status de uma notificação enviada
   * @param notificationId ID da notificação
   */
  getStatus(notificationId: string): Promise<{
    status: 'SCHEDULED' | 'SENT' | 'DELIVERED' | 'FAILED' | 'OPENED' | 'CLICKED' | 'EXPIRED' | 'CANCELLED';
    timestamp: Date;
    details?: Record<string, any>;
  }>;
}