/**
 * @file push-adapter.ts
 * @description Implementação de adaptador de notificação push
 * 
 * Este adaptador permite o envio de notificações push utilizando
 * diferentes provedores (Firebase, OneSignal, etc.), com suporte a
 * múltiplos formatos, ações e rastreamento de entrega.
 */

import * as admin from 'firebase-admin';
import { SNS } from 'aws-sdk';
import axios from 'axios';
import { Logger } from '../../../infrastructure/observability/logger';
import { 
  NotificationAdapter, 
  NotificationResult, 
  NotificationRecipient,
  NotificationContent,
  BaseNotificationOptions
} from './notification-adapter';
import { BaseEvent } from '../core/base-event';
import { NotificationChannel, ChannelConfig } from '../core/notification-channel';

/**
 * Configuração do adaptador de push
 */
export interface PushAdapterConfig extends ChannelConfig {
  /**
   * Tipo de provedor de push a ser utilizado
   */
  provider: 'FIREBASE' | 'ONE_SIGNAL' | 'SNS' | 'MOCK';
  
  /**
   * Configurações específicas para Firebase
   */
  firebase?: {
    projectId: string;
    clientEmail?: string;
    privateKey?: string;
    databaseURL?: string;
    credential?: admin.credential.Credential;
  };
  
  /**
   * Configurações específicas para OneSignal
   */
  oneSignal?: {
    appId: string;
    apiKey: string;
    restApiKey: string;
  };
  
  /**
   * Configurações específicas para Amazon SNS
   */
  sns?: {
    region: string;
    accessKeyId?: string;
    secretAccessKey?: string;
    applicationArn?: string;
    platformApplicationArn?: string;
  };
  
  /**
   * URL do ícone padrão para notificações
   */
  defaultIconUrl?: string;
  
  /**
   * URL da imagem padrão para notificações (quando suportado)
   */
  defaultImageUrl?: string;
  
  /**
   * Som padrão para notificações
   */
  defaultSound?: string;
  
  /**
   * Cor de destaque da notificação (hexadecimal)
   */
  defaultColor?: string;
  
  /**
   * Canal padrão para Android (Android 8.0+)
   */
  defaultAndroidChannel?: string;
  
  /**
   * Tags padrão para categorização
   */
  defaultTags?: string[];
  
  /**
   * TTL padrão para notificações (em segundos)
   */
  defaultTtlSeconds?: number;
  
  /**
   * URL base para deep links
   */
  deepLinkBaseUrl?: string;
}

/**
 * Implementação de adaptador de notificação para canal push
 */
export class PushAdapter implements NotificationAdapter {
  readonly channelType = NotificationChannel.PUSH;
  
  private firebaseApp?: admin.app.App;
  private snsClient?: SNS;
  private config: PushAdapterConfig;
  private initialized = false;
  private logger = new Logger('PushAdapter');
  
  /**
   * Construtor
   */
  constructor() {}
  
  /**
   * Verifica se o adaptador está inicializado e pronto para uso
   */
  async isReady(): Promise<boolean> {
    return this.initialized;
  }
  
  /**
   * Inicializa o adaptador com a configuração
   * @param config Configuração do canal
   */
  async initialize(config: PushAdapterConfig): Promise<void> {
    this.config = config;
    
    try {
      switch (config.provider) {
        case 'FIREBASE':
          await this.initializeFirebase(config);
          break;
          
        case 'ONE_SIGNAL':
          // OneSignal não requer inicialização especial além da configuração
          if (!config.oneSignal) {
            throw new Error('Configuração OneSignal não fornecida');
          }
          break;
          
        case 'SNS':
          if (!config.sns) {
            throw new Error('Configuração AWS SNS não fornecida');
          }
          
          this.snsClient = new SNS({
            region: config.sns.region,
            accessKeyId: config.sns.accessKeyId,
            secretAccessKey: config.sns.secretAccessKey
          });
          break;
          
        case 'MOCK':
          // Nenhuma inicialização necessária para o adaptador de mock
          break;
          
        default:
          throw new Error(`Provedor de push não suportado: ${config.provider}`);
      }
      
      this.initialized = true;
      this.logger.info(`Adaptador de push inicializado com sucesso: ${config.provider}`);
    } catch (error) {
      this.initialized = false;
      this.logger.error(`Falha ao inicializar adaptador de push: ${error}`);
      throw error;
    }
  }
  
  /**
   * Inicializa o cliente Firebase
   * @param config Configuração do adaptador
   */
  private async initializeFirebase(config: PushAdapterConfig): Promise<void> {
    if (!config.firebase) {
      throw new Error('Configuração Firebase não fornecida');
    }
    
    try {
      // Verificar se já existe uma app com este nome
      try {
        this.firebaseApp = admin.app('innovabiz-push-notifications');
      } catch (e) {
        // Se não existir, inicializar uma nova
        this.firebaseApp = admin.initializeApp({
          credential: config.firebase.credential || 
            admin.credential.cert({
              projectId: config.firebase.projectId,
              clientEmail: config.firebase.clientEmail,
              privateKey: config.firebase.privateKey?.replace(/\\n/g, '\n')
            }),
          databaseURL: config.firebase.databaseURL
        }, 'innovabiz-push-notifications');
      }
    } catch (error) {
      throw new Error(`Erro ao inicializar Firebase: ${error}`);
    }
  }
  
  /**
   * Envia uma notificação push para um único destinatário
   * @param recipient Destinatário da notificação
   * @param content Conteúdo da notificação
   * @param event Evento que originou a notificação (opcional)
   * @param options Opções adicionais para o envio
   */
  async send(
    recipient: NotificationRecipient,
    content: NotificationContent,
    event?: BaseEvent,
    options?: BaseNotificationOptions
  ): Promise<NotificationResult> {
    if (!this.initialized) {
      return {
        success: false,
        errorMessage: 'Adaptador de push não inicializado',
        errorCode: 'PUSH_ADAPTER_NOT_INITIALIZED',
        timestamp: new Date()
      };
    }
    
    const startTime = Date.now();
    const notificationId = options?.notificationId || `push-${Date.now()}-${Math.random().toString(36).substring(2, 10)}`;
    
    try {
      // Verificar se o destinatário tem tokens de dispositivo válidos
      const tokens = recipient.addresses?.get(NotificationChannel.PUSH);
      
      if (!tokens || tokens.length === 0) {
        return {
          success: false,
          notificationId,
          errorMessage: `Destinatário ${recipient.id} não possui tokens de dispositivo`,
          errorCode: 'DEVICE_TOKEN_MISSING',
          timestamp: new Date()
        };
      }
      
      // Preparar os dados da notificação
      const notificationData = this.prepareNotificationData(content, recipient, event, notificationId);
      
      let messageId: string;
      
      switch (this.config.provider) {
        case 'FIREBASE':
          messageId = await this.sendViaFirebase(tokens, notificationData, options);
          break;
          
        case 'ONE_SIGNAL':
          messageId = await this.sendViaOneSignal(tokens, notificationData, recipient, options);
          break;
          
        case 'SNS':
          messageId = await this.sendViaSns(tokens, notificationData, options);
          break;
          
        case 'MOCK':
          messageId = this.sendViaMock(tokens, notificationData);
          break;
          
        default:
          throw new Error(`Provedor de push não suportado: ${this.config.provider}`);
      }
      
      const deliveryTime = Date.now() - startTime;
      this.logger.info(`Push enviado para ${recipient.id} em ${deliveryTime}ms`, {
        messageId,
        notificationId,
        recipientId: recipient.id
      });
      
      return {
        success: true,
        notificationId,
        details: {
          messageId,
          provider: this.config.provider,
          tokensCount: tokens.length,
          deliveryTime
        },
        timestamp: new Date()
      };
    } catch (error) {
      this.logger.error(`Erro ao enviar push para ${recipient.id}: ${error}`, {
        notificationId,
        recipientId: recipient.id,
        error
      });
      
      return {
        success: false,
        notificationId,
        errorMessage: `Erro ao enviar push: ${error.message || error}`,
        errorCode: 'PUSH_SEND_FAILED',
        timestamp: new Date()
      };
    }
  }
  
  /**
   * Prepara os dados da notificação para envio
   * @param content Conteúdo da notificação
   * @param recipient Destinatário da notificação
   * @param event Evento que originou a notificação
   * @param notificationId ID da notificação
   */
  private prepareNotificationData(
    content: NotificationContent,
    recipient: NotificationRecipient,
    event?: BaseEvent,
    notificationId?: string
  ): Record<string, any> {
    // Dados base da notificação
    const data: Record<string, any> = {
      title: content.title || 'InnovaBiz',
      body: content.body,
      icon: this.config.defaultIconUrl,
      sound: this.config.defaultSound || 'default',
      color: this.config.defaultColor,
      notificationId,
      recipientId: recipient.id
    };
    
    // Adicionar dados do evento se existir
    if (event) {
      data.eventId = event.eventId;
      data.eventType = event.code;
    }
    
    // Adicionar ações se existirem
    if (content.actions && content.actions.length > 0) {
      data.actions = JSON.stringify(content.actions);
      
      // Adicionar URL de ação principal (para deep link)
      const primaryAction = content.actions.find(a => a.actionType === 'LINK' || a.actionType === 'BUTTON');
      if (primaryAction && primaryAction.url) {
        data.clickAction = primaryAction.url;
        
        // Adicionar deep link se configurado
        if (this.config.deepLinkBaseUrl) {
          data.deepLink = `${this.config.deepLinkBaseUrl}/${primaryAction.id}`;
        }
      }
    }
    
    // Adicionar imagem se existir
    if (content.resourceUrls && content.resourceUrls.length > 0) {
      const image = content.resourceUrls.find(r => r.type === 'IMAGE');
      if (image && image.url) {
        data.image = image.url;
      }
    } else if (this.config.defaultImageUrl) {
      data.image = this.config.defaultImageUrl;
    }
    
    return data;
  }
  
  /**
   * Envia notificação via Firebase Cloud Messaging
   * @param tokens Tokens de dispositivos
   * @param notificationData Dados da notificação
   * @param options Opções de envio
   */
  private async sendViaFirebase(
    tokens: string[],
    notificationData: Record<string, any>,
    options?: BaseNotificationOptions
  ): Promise<string> {
    if (!this.firebaseApp) {
      throw new Error('Cliente Firebase não inicializado');
    }
    
    const messaging = this.firebaseApp.messaging();
    
    const message: admin.messaging.MulticastMessage = {
      tokens,
      notification: {
        title: notificationData.title,
        body: notificationData.body,
        imageUrl: notificationData.image
      },
      data: Object.entries(notificationData).reduce((acc, [key, value]) => {
        acc[key] = value === undefined || value === null ? '' : String(value);
        return acc;
      }, {} as Record<string, string>),
      android: {
        priority: 'high',
        notification: {
          sound: notificationData.sound,
          color: notificationData.color,
          channelId: this.config.defaultAndroidChannel,
          clickAction: notificationData.clickAction
        }
      },
      apns: {
        payload: {
          aps: {
            alert: {
              title: notificationData.title,
              body: notificationData.body
            },
            sound: notificationData.sound,
            badge: 1,
            'mutable-content': 1
          },
          notificationId: notificationData.notificationId
        }
      },
      webpush: {
        notification: {
          icon: notificationData.icon,
          badge: notificationData.icon
        }
      }
    };
    
    // Definir TTL se configurado
    if (options?.ttlSeconds || this.config.defaultTtlSeconds) {
      const ttlSeconds = options?.ttlSeconds || this.config.defaultTtlSeconds;
      message.android = {
        ...message.android,
        ttl: ttlSeconds! * 1000
      };
      message.apns = {
        ...message.apns,
        headers: {
          'apns-expiration': Math.floor(Date.now() / 1000 + ttlSeconds!)
        }
      };
      message.webpush = {
        ...message.webpush,
        headers: {
          TTL: String(ttlSeconds)
        }
      };
    }
    
    const response = await messaging.sendMulticast(message);
    return `fcm-batch-${Date.now()}-${response.successCount}/${tokens.length}`;
  }
  
  /**
   * Envia notificação via OneSignal
   * @param tokens IDs de jogadores OneSignal
   * @param notificationData Dados da notificação
   * @param recipient Destinatário da notificação
   * @param options Opções de envio
   */
  private async sendViaOneSignal(
    tokens: string[],
    notificationData: Record<string, any>,
    recipient: NotificationRecipient,
    options?: BaseNotificationOptions
  ): Promise<string> {
    if (!this.config.oneSignal) {
      throw new Error('Configuração OneSignal não fornecida');
    }
    
    const oneSignalPayload = {
      app_id: this.config.oneSignal.appId,
      include_player_ids: tokens,
      headings: { en: notificationData.title },
      contents: { en: notificationData.body },
      data: notificationData,
      android_channel_id: this.config.defaultAndroidChannel,
      small_icon: 'ic_notification',
      large_icon: notificationData.icon,
      ios_attachments: notificationData.image ? { id1: notificationData.image } : undefined,
      big_picture: notificationData.image,
      chrome_web_image: notificationData.image,
      buttons: notificationData.actions ? JSON.parse(notificationData.actions)
        .filter((a: any) => a.actionType === 'BUTTON')
        .map((a: any) => ({ id: a.id, text: a.label, url: a.url })) : undefined
    };
    
    // Definir prioridade baseada nas opções
    if (options?.priority) {
      oneSignalPayload.priority = options.priority === 'HIGHEST' || options.priority === 'HIGH' ? 10 : 
        options.priority === 'LOWEST' || options.priority === 'LOW' ? 5 : 8;
    }
    
    // Definir TTL se configurado
    if (options?.ttlSeconds || this.config.defaultTtlSeconds) {
      oneSignalPayload.ttl = options?.ttlSeconds || this.config.defaultTtlSeconds;
    }
    
    const response = await axios.post(
      'https://onesignal.com/api/v1/notifications',
      oneSignalPayload,
      {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Basic ${this.config.oneSignal.restApiKey}`
        }
      }
    );
    
    return response.data.id;
  }
  
  /**
   * Envia notificação via Amazon SNS
   * @param tokens ARNs dos endpoints SNS
   * @param notificationData Dados da notificação
   * @param options Opções de envio
   */
  private async sendViaSns(
    tokens: string[],
    notificationData: Record<string, any>,
    options?: BaseNotificationOptions
  ): Promise<string> {
    if (!this.snsClient || !this.config.sns) {
      throw new Error('Cliente SNS não inicializado');
    }
    
    const messageAttributes = {
      'AWS.SNS.MOBILE.APNS.PUSH_TYPE': {
        DataType: 'String',
        StringValue: 'alert'
      }
    };
    
    // Preparar mensagem para diferentes plataformas
    const apnsPayload = {
      aps: {
        alert: {
          title: notificationData.title,
          body: notificationData.body
        },
        sound: notificationData.sound,
        badge: 1,
        'mutable-content': 1
      },
      ...notificationData
    };
    
    const fcmPayload = {
      notification: {
        title: notificationData.title,
        body: notificationData.body,
        image: notificationData.image
      },
      data: notificationData
    };
    
    const message = {
      default: notificationData.body,
      APNS: JSON.stringify({ aps: apnsPayload }),
      APNS_SANDBOX: JSON.stringify({ aps: apnsPayload }),
      GCM: JSON.stringify(fcmPayload)
    };
    
    // Publicar para cada token de endpoint
    const publishPromises = tokens.map(async token => {
      try {
        const params: SNS.PublishInput = {
          MessageStructure: 'json',
          Message: JSON.stringify(message),
          MessageAttributes: messageAttributes,
          TargetArn: token.startsWith('arn:') ? token : `${this.config.sns!.platformApplicationArn}/token-${token}`
        };
        
        const result = await this.snsClient!.publish(params).promise();
        return result.MessageId;
      } catch (error) {
        this.logger.error(`Erro ao enviar para endpoint SNS ${token}: ${error}`);
        throw error;
      }
    });
    
    // Aguardar todas as publicações
    const results = await Promise.allSettled(publishPromises);
    const successCount = results.filter(r => r.status === 'fulfilled').length;
    
    return `sns-batch-${Date.now()}-${successCount}/${tokens.length}`;
  }
  
  /**
   * Simula o envio de notificação push para testes
   * @param tokens Tokens de dispositivos
   * @param notificationData Dados da notificação
   */
  private sendViaMock(
    tokens: string[],
    notificationData: Record<string, any>
  ): string {
    const mockId = `mock-${Date.now()}-${Math.random().toString(36).substring(2, 10)}`;
    
    this.logger.info(`[MOCK] Push enviado para ${tokens.length} dispositivos`, {
      mockId,
      notificationData,
      tokens: tokens.length > 3 ? 
        `${tokens.slice(0, 3).join(', ')}... (${tokens.length} total)` : 
        tokens.join(', ')
    });
    
    return mockId;
  }
  
  /**
   * Envia uma notificação push para múltiplos destinatários
   * @param recipients Lista de destinatários
   * @param content Conteúdo da notificação
   * @param event Evento que originou a notificação (opcional)
   * @param options Opções adicionais para o envio
   */
  async sendBulk(
    recipients: NotificationRecipient[],
    content: NotificationContent,
    event?: BaseEvent,
    options?: BaseNotificationOptions
  ): Promise<NotificationResult[]> {
    const results: NotificationResult[] = [];
    const batchId = `batch-${Date.now()}`;
    
    this.logger.info(`Iniciando envio em lote de ${recipients.length} notificações push`, {
      recipientCount: recipients.length,
      batchId
    });
    
    // Se for Firebase, podemos usar a API batch para melhor performance
    if (this.config.provider === 'FIREBASE' && this.firebaseApp) {
      try {
        // Coletar todos os tokens em um único array
        const allTokens: string[] = [];
        const tokenToRecipientMap = new Map<string, string>();
        
        for (const recipient of recipients) {
          const tokens = recipient.addresses?.get(NotificationChannel.PUSH) || [];
          for (const token of tokens) {
            allTokens.push(token);
            tokenToRecipientMap.set(token, recipient.id);
          }
        }
        
        if (allTokens.length === 0) {
          this.logger.warn('Nenhum token encontrado para envio em lote');
          return recipients.map(recipient => ({
            success: false,
            notificationId: `push-${Date.now()}-${recipient.id}`,
            errorMessage: 'Destinatário não possui tokens de dispositivo',
            errorCode: 'DEVICE_TOKEN_MISSING',
            timestamp: new Date()
          }));
        }
        
        // Preparar a mensagem base
        const notificationData = this.prepareNotificationData(content, 
          { id: 'bulk', type: 'GROUP' } as NotificationRecipient, 
          event, 
          `${options?.notificationId || 'push'}-batch-${batchId}`
        );
        
        // Enviar para todos os tokens de uma vez
        const batchResult = await this.sendViaFirebase(allTokens, notificationData, options);
        
        // Criar resultado individual para cada destinatário
        return recipients.map(recipient => ({
          success: true,
          notificationId: `${options?.notificationId || 'push'}-${recipient.id}-${Date.now()}`,
          details: {
            batchId,
            batchMessageId: batchResult,
            provider: this.config.provider
          },
          timestamp: new Date()
        }));
      } catch (error) {
        this.logger.error(`Erro no envio em lote via Firebase: ${error}`);
        // Continuar com o método padrão
      }
    }
    
    // Método padrão: enviar individualmente para cada destinatário
    const sendPromises = recipients.map(recipient => 
      this.send(recipient, content, event, {
        ...options,
        notificationId: `${options?.notificationId || 'push'}-${recipient.id}-${Date.now()}`,
        tracking: {
          ...(options?.tracking || {}),
          metadata: {
            ...(options?.tracking?.metadata || {}),
            batchId
          }
        }
      })
    );
    
    // Aguardar todas as operações
    const recipientResults = await Promise.allSettled(sendPromises);
    
    // Processar resultados
    for (let i = 0; i < recipientResults.length; i++) {
      const result = recipientResults[i];
      if (result.status === 'fulfilled') {
        results.push(result.value);
      } else {
        // Adicionar resultado de falha
        results.push({
          success: false,
          notificationId: `${options?.notificationId || 'push'}-${recipients[i].id}-${Date.now()}`,
          errorMessage: `Erro ao enviar push: ${result.reason}`,
          errorCode: 'PUSH_SEND_ERROR',
          timestamp: new Date()
        });
      }
    }
    
    // Contabilizar resultados
    const successCount = results.filter(r => r.success).length;
    const failCount = results.length - successCount;
    
    this.logger.info(`Concluído envio em lote de push: ${successCount} sucesso, ${failCount} falha`, {
      batchId,
      successCount,
      failCount
    });
    
    return results;
  }
  
  /**
   * Cancela uma notificação agendada
   * @param notificationId ID da notificação
   */
  async cancel(notificationId: string): Promise<boolean> {
    // Maioria dos provedores de push não suporta cancelamento após envio
    this.logger.warn(`Tentativa de cancelamento de push ${notificationId}. Push geralmente não pode ser cancelado após enviado.`);
    return false;
  }
  
  /**
   * Verifica o status de uma notificação enviada
   * @param notificationId ID da notificação
   */
  async getStatus(notificationId: string): Promise<{
    status: 'SCHEDULED' | 'SENT' | 'DELIVERED' | 'FAILED' | 'OPENED' | 'CLICKED' | 'EXPIRED' | 'CANCELLED';
    timestamp: Date;
    details?: Record<string, any>;
  }> {
    // Status de notificações push geralmente só pode ser obtido via webhooks
    return {
      status: 'SENT', // Status padrão assumido
      timestamp: new Date(),
      details: {
        note: 'Status real depende de webhooks de provedores ou rastreamento'
      }
    };
  }
}