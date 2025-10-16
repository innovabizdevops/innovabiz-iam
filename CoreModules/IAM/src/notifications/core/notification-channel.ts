/**
 * @file notification-channel.ts
 * @description Define os canais de notificação suportados e suas configurações
 * 
 * Este módulo implementa a estrutura para os diferentes canais de notificação
 * suportados pela plataforma INNOVABIZ, incluindo configurações específicas
 * para cada canal e estratégias de entrega.
 */

/**
 * Enumeração dos canais de notificação suportados
 */
export enum NotificationChannel {
  // Canais de comunicação direta
  EMAIL = 'EMAIL',
  SMS = 'SMS',
  PUSH = 'PUSH',
  WHATSAPP = 'WHATSAPP',
  IN_APP = 'IN_APP',
  WEB_SOCKET = 'WEB_SOCKET',
  
  // Canais de mídia social e mensageria
  TELEGRAM = 'TELEGRAM',
  MESSENGER = 'MESSENGER',
  SIGNAL = 'SIGNAL',
  
  // Canais baseados em voz
  VOICE_CALL = 'VOICE_CALL',
  IVR = 'IVR', // Interactive Voice Response
  
  // Canais para bancos/financeiras
  MOBILE_BANKING = 'MOBILE_BANKING',
  INTERNET_BANKING = 'INTERNET_BANKING',
  ATM = 'ATM',
  POS = 'POS', // Point of Sale
  
  // Canais para integração com sistemas externos
  API_WEBHOOK = 'API_WEBHOOK',
  KAFKA = 'KAFKA',
  MESSAGE_QUEUE = 'MESSAGE_QUEUE',
  
  // Canais físicos (para registro/completude)
  MAIL = 'MAIL',
  BRANCH_NOTIFICATION = 'BRANCH_NOTIFICATION',
  
  // Canais para cenários específicos
  EMERGENCY = 'EMERGENCY' // Canal de emergência (combina múltiplos)
}

/**
 * Interface para configuração base de canal de notificação
 */
export interface ChannelConfig {
  enabled: boolean;
  rateLimits?: {
    maxPerMinute?: number;
    maxPerHour?: number;
    maxPerDay?: number;
    burstLimit?: number;
  };
  retryPolicy?: {
    maxRetries: number;
    initialDelayMs: number;
    backoffMultiplier: number;
    maxDelayMs: number;
  };
  deliveryPriority?: number; // 1-100, maior número = maior prioridade
  deliveryTimeConstraints?: {
    businessHoursOnly?: boolean;
    startTime?: string; // formato "HH:MM"
    endTime?: string; // formato "HH:MM"
    timeZone?: string; // formato IANA (ex: "Africa/Luanda")
    blockedDays?: number[]; // 0 = domingo, 6 = sábado
  };
}

/**
 * Configuração para canal de email
 */
export interface EmailChannelConfig extends ChannelConfig {
  templates: {
    html?: boolean;
    plainText?: boolean;
    mjml?: boolean;
    amp4Email?: boolean;
  };
  sender: {
    email: string;
    name?: string;
    replyTo?: string;
  };
  attachments?: {
    allowed: boolean;
    maxSizeKb?: number;
    allowedTypes?: string[];
  };
  trackingOptions?: {
    openTracking?: boolean;
    clickTracking?: boolean;
    unsubscribeTracking?: boolean;
  };
  provider?: string;
}

/**
 * Configuração para canal de SMS
 */
export interface SmsChannelConfig extends ChannelConfig {
  sender: string;
  maxLength?: number;
  unicode?: boolean;
  provider?: string;
  deliveryReport?: boolean;
}

/**
 * Configuração para notificações push
 */
export interface PushChannelConfig extends ChannelConfig {
  platforms: {
    android?: boolean;
    ios?: boolean;
    web?: boolean;
    huawei?: boolean;
  };
  options: {
    sound?: boolean;
    badge?: boolean;
    actionButtons?: boolean;
    images?: boolean;
    silentPush?: boolean;
  };
  ttlSeconds?: number;
  collapseKey?: string;
  provider?: string;
}

/**
 * Configuração para canal de WhatsApp
 */
export interface WhatsAppChannelConfig extends ChannelConfig {
  businessAccountId: string;
  phoneNumberId: string;
  templateNamespace?: string;
  supportedMessageTypes: {
    text?: boolean;
    media?: boolean;
    template?: boolean;
    interactive?: boolean;
    location?: boolean;
  };
  provider?: string;
}

/**
 * Configuração para canal in-app
 */
export interface InAppChannelConfig extends ChannelConfig {
  displayOptions: {
    toast?: boolean;
    banner?: boolean;
    modal?: boolean;
    inbox?: boolean;
    statusBar?: boolean;
  };
  persistenceOptions: {
    saveToInbox: boolean;
    expiryDays?: number;
  };
  uiOptions?: {
    theme?: string;
    animations?: boolean;
    customCssClass?: string;
  };
}

/**
 * Configuração para canal de API/webhook
 */
export interface ApiWebhookChannelConfig extends ChannelConfig {
  endpoints: {
    url: string;
    method: 'GET' | 'POST' | 'PUT';
    headers?: Record<string, string>;
    authType?: 'NONE' | 'BASIC' | 'BEARER' | 'API_KEY' | 'OAUTH2';
    authCredentials?: Record<string, string>;
  }[];
  payloadFormat: 'JSON' | 'XML' | 'FORM' | 'CUSTOM';
  securityOptions?: {
    signPayload?: boolean;
    encryptPayload?: boolean;
    tlsVerification?: boolean;
  };
}

/**
 * Mapeamento de tipos de configuração por canal
 */
export type ChannelConfigType = {
  [NotificationChannel.EMAIL]: EmailChannelConfig;
  [NotificationChannel.SMS]: SmsChannelConfig;
  [NotificationChannel.PUSH]: PushChannelConfig;
  [NotificationChannel.WHATSAPP]: WhatsAppChannelConfig;
  [NotificationChannel.IN_APP]: InAppChannelConfig;
  [NotificationChannel.API_WEBHOOK]: ApiWebhookChannelConfig;
  [key: string]: ChannelConfig;
};

/**
 * Classe para gerenciar as configurações de canais
 */
export class ChannelConfigManager {
  private static instance: ChannelConfigManager;
  private configs: Map<NotificationChannel, ChannelConfig>;
  
  private constructor() {
    this.configs = new Map();
    this.initializeDefaultConfigs();
  }
  
  /**
   * Inicializa as configurações padrão para os canais
   */
  private initializeDefaultConfigs(): void {
    // Configuração padrão que se aplica a todos os canais
    const defaultConfig: ChannelConfig = {
      enabled: true,
      rateLimits: {
        maxPerMinute: 60,
        maxPerHour: 500,
        maxPerDay: 2000
      },
      retryPolicy: {
        maxRetries: 3,
        initialDelayMs: 1000,
        backoffMultiplier: 2,
        maxDelayMs: 60000
      }
    };
    
    // Inicializa cada canal com a configuração padrão e sobrescreve com específicas
    Object.values(NotificationChannel).forEach(channel => {
      this.configs.set(channel, { ...defaultConfig });
    });
  }
  
  /**
   * Obtém a instância única do gerenciador (padrão Singleton)
   */
  public static getInstance(): ChannelConfigManager {
    if (!ChannelConfigManager.instance) {
      ChannelConfigManager.instance = new ChannelConfigManager();
    }
    return ChannelConfigManager.instance;
  }
  
  /**
   * Obtém a configuração para um canal específico
   * @param channel Canal de notificação
   * @returns Configuração do canal
   */
  public getConfig<T extends NotificationChannel>(channel: T): ChannelConfigType[T] {
    return this.configs.get(channel) as ChannelConfigType[T];
  }
  
  /**
   * Define a configuração para um canal específico
   * @param channel Canal de notificação
   * @param config Configuração do canal
   */
  public setConfig<T extends NotificationChannel>(
    channel: T, 
    config: ChannelConfigType[T]
  ): void {
    this.configs.set(channel, config);
  }
  
  /**
   * Verifica se um canal está habilitado
   * @param channel Canal de notificação
   * @returns Verdadeiro se o canal estiver habilitado
   */
  public isChannelEnabled(channel: NotificationChannel): boolean {
    const config = this.configs.get(channel);
    return config ? config.enabled : false;
  }
  
  /**
   * Habilita um canal de notificação
   * @param channel Canal de notificação
   */
  public enableChannel(channel: NotificationChannel): void {
    const config = this.configs.get(channel);
    if (config) {
      config.enabled = true;
    }
  }
  
  /**
   * Desabilita um canal de notificação
   * @param channel Canal de notificação
   */
  public disableChannel(channel: NotificationChannel): void {
    const config = this.configs.get(channel);
    if (config) {
      config.enabled = false;
    }
  }
}