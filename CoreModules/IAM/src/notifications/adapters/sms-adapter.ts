/**
 * @file sms-adapter.ts
 * @description Implementação de adaptador de notificação por SMS
 * 
 * Este adaptador permite o envio de notificações por SMS utilizando
 * diferentes provedores (Twilio, AWS SNS, etc.), com suporte a múltiplos
 * formatos e rastreamento de entrega.
 */

import { Twilio } from 'twilio';
import { SNS } from 'aws-sdk';
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
 * Configuração do adaptador de SMS
 */
export interface SmsAdapterConfig extends ChannelConfig {
  /**
   * Tipo de provedor de SMS a ser utilizado
   */
  provider: 'TWILIO' | 'SNS' | 'MOCK';
  
  /**
   * Configurações específicas para Twilio
   */
  twilio?: {
    accountSid: string;
    authToken: string;
    fromNumber: string;
  };
  
  /**
   * Configurações específicas para Amazon SNS
   */
  sns?: {
    region: string;
    accessKeyId?: string;
    secretAccessKey?: string;
    applicationArn?: string;
  };
  
  /**
   * Número de telefone de origem padrão
   */
  defaultSender?: string;
  
  /**
   * Prefixo para adicionar ao início das mensagens (ex: [INNOVABIZ])
   */
  messagePrefix?: string;
  
  /**
   * URL base para rastreamento de SMS (encurtamento de URLs)
   */
  trackingBaseUrl?: string;
  
  /**
   * Configurações para throttling de envio
   */
  throttling?: {
    maxMessagesPerSecond: number;
    maxMessagesPerMinute: number;
  };
  
  /**
   * Template padrão para mensagens ({{variáveis}} serão substituídas)
   */
  defaultTemplate?: string;
  
  /**
   * Número máximo de caracteres por mensagem
   */
  maxMessageLength?: number;
  
  /**
   * Gerenciar automaticamente longas mensagens (dividir em múltiplas)
   */
  autoSplitLongMessages?: boolean;
}

/**
 * Implementação de adaptador de notificação para canal de SMS
 */
export class SmsAdapter implements NotificationAdapter {
  readonly channelType = NotificationChannel.SMS;
  
  private twilioClient?: Twilio;
  private snsClient?: SNS;
  private config: SmsAdapterConfig;
  private initialized = false;
  private logger = new Logger('SmsAdapter');
  
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
  async initialize(config: SmsAdapterConfig): Promise<void> {
    this.config = config;
    
    try {
      switch (config.provider) {
        case 'TWILIO':
          if (!config.twilio) {
            throw new Error('Configuração Twilio não fornecida');
          }
          this.twilioClient = new Twilio(
            config.twilio.accountSid,
            config.twilio.authToken
          );
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
          throw new Error(`Provedor de SMS não suportado: ${config.provider}`);
      }
      
      this.initialized = true;
      this.logger.info(`Adaptador de SMS inicializado com sucesso: ${config.provider}`);
    } catch (error) {
      this.initialized = false;
      this.logger.error(`Falha ao inicializar adaptador de SMS: ${error}`);
      throw error;
    }
  }
  
  /**
   * Envia uma notificação por SMS para um único destinatário
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
        errorMessage: 'Adaptador de SMS não inicializado',
        errorCode: 'SMS_ADAPTER_NOT_INITIALIZED',
        timestamp: new Date()
      };
    }
    
    const startTime = Date.now();
    const notificationId = options?.notificationId || `sms-${Date.now()}-${Math.random().toString(36).substring(2, 10)}`;
    
    try {
      // Verificar se o destinatário tem um número de telefone válido
      const phoneNumbers = recipient.addresses?.get(NotificationChannel.SMS);
      
      if (!phoneNumbers || phoneNumbers.length === 0) {
        return {
          success: false,
          notificationId,
          errorMessage: `Destinatário ${recipient.id} não possui número de telefone`,
          errorCode: 'PHONE_NUMBER_MISSING',
          timestamp: new Date()
        };
      }
      
      // Usar o primeiro número de telefone disponível
      const toPhoneNumber = phoneNumbers[0];
      
      // Preparar o corpo da mensagem
      let messageBody = content.body;
      
      // Adicionar prefixo se configurado
      if (this.config.messagePrefix) {
        messageBody = `${this.config.messagePrefix} ${messageBody}`;
      }
      
      // Verificar tamanho da mensagem
      const maxLength = this.config.maxMessageLength || 160;
      if (messageBody.length > maxLength && !this.config.autoSplitLongMessages) {
        this.logger.warn(`Mensagem SMS excede tamanho máximo (${messageBody.length} > ${maxLength})`, {
          notificationId,
          recipientId: recipient.id
        });
        
        // Truncar a mensagem se não estiver configurado para dividir
        messageBody = messageBody.substring(0, maxLength - 3) + '...';
      }
      
      // Adicionar tracking URL se configurado
      if (this.config.trackingBaseUrl && content.actions?.length) {
        const action = content.actions[0]; // Usar a primeira ação para SMS
        if (action.url) {
          const trackingUrl = `${this.config.trackingBaseUrl}/t/${notificationId}`;
          messageBody = messageBody.replace(action.url, trackingUrl);
        }
      }
      
      let messageId: string;
      
      switch (this.config.provider) {
        case 'TWILIO':
          messageId = await this.sendViaTwilio(toPhoneNumber, messageBody, notificationId);
          break;
          
        case 'SNS':
          messageId = await this.sendViaSns(toPhoneNumber, messageBody, notificationId);
          break;
          
        case 'MOCK':
          messageId = this.sendViaMock(toPhoneNumber, messageBody, notificationId);
          break;
          
        default:
          throw new Error(`Provedor de SMS não suportado: ${this.config.provider}`);
      }
      
      const deliveryTime = Date.now() - startTime;
      this.logger.info(`SMS enviado para ${toPhoneNumber} em ${deliveryTime}ms`, {
        messageId,
        notificationId,
        recipientId: recipient.id
      });
      
      return {
        success: true,
        notificationId,
        details: {
          messageId,
          phoneNumber: toPhoneNumber,
          provider: this.config.provider,
          deliveryTime
        },
        timestamp: new Date()
      };
    } catch (error) {
      this.logger.error(`Erro ao enviar SMS para ${recipient.id}: ${error}`, {
        notificationId,
        recipientId: recipient.id,
        error
      });
      
      return {
        success: false,
        notificationId,
        errorMessage: `Erro ao enviar SMS: ${error.message || error}`,
        errorCode: 'SMS_SEND_FAILED',
        timestamp: new Date()
      };
    }
  }
  
  /**
   * Envia SMS usando o provedor Twilio
   * @param toPhoneNumber Número de telefone do destinatário
   * @param messageBody Corpo da mensagem
   * @param notificationId ID da notificação
   */
  private async sendViaTwilio(
    toPhoneNumber: string, 
    messageBody: string,
    notificationId: string
  ): Promise<string> {
    if (!this.twilioClient || !this.config.twilio) {
      throw new Error('Cliente Twilio não inicializado');
    }
    
    // Normalizar o número de telefone para formato E.164 se necessário
    const normalizedNumber = this.normalizePhoneNumber(toPhoneNumber);
    
    const message = await this.twilioClient.messages.create({
      body: messageBody,
      from: this.config.twilio.fromNumber,
      to: normalizedNumber,
      statusCallback: this.config.trackingBaseUrl ? 
        `${this.config.trackingBaseUrl}/webhook/sms/status/${notificationId}` : undefined
    });
    
    return message.sid;
  }
  
  /**
   * Envia SMS usando Amazon SNS
   * @param toPhoneNumber Número de telefone do destinatário
   * @param messageBody Corpo da mensagem
   * @param notificationId ID da notificação
   */
  private async sendViaSns(
    toPhoneNumber: string, 
    messageBody: string,
    notificationId: string
  ): Promise<string> {
    if (!this.snsClient) {
      throw new Error('Cliente SNS não inicializado');
    }
    
    // Normalizar o número de telefone para formato E.164
    const normalizedNumber = this.normalizePhoneNumber(toPhoneNumber);
    
    const params: SNS.PublishInput = {
      Message: messageBody,
      PhoneNumber: normalizedNumber,
      MessageAttributes: {
        'AWS.SNS.SMS.SenderID': {
          DataType: 'String',
          StringValue: this.config.defaultSender || 'INNOVABIZ'
        },
        'AWS.SNS.SMS.SMSType': {
          DataType: 'String',
          StringValue: 'Transactional'
        },
        'NotificationId': {
          DataType: 'String',
          StringValue: notificationId
        }
      }
    };
    
    const result = await this.snsClient.publish(params).promise();
    return result.MessageId as string;
  }
  
  /**
   * Simula o envio de SMS para testes
   * @param toPhoneNumber Número de telefone do destinatário
   * @param messageBody Corpo da mensagem
   * @param notificationId ID da notificação
   */
  private sendViaMock(
    toPhoneNumber: string, 
    messageBody: string,
    notificationId: string
  ): string {
    const mockId = `mock-${Date.now()}-${Math.random().toString(36).substring(2, 10)}`;
    
    this.logger.info(`[MOCK] SMS enviado para ${toPhoneNumber}`, {
      messageBody,
      notificationId,
      mockId
    });
    
    return mockId;
  }
  
  /**
   * Normaliza um número de telefone para formato E.164
   * @param phoneNumber Número de telefone
   */
  private normalizePhoneNumber(phoneNumber: string): string {
    // Implementação simples - em produção usaria uma biblioteca mais robusta
    let normalized = phoneNumber.replace(/\D/g, '');
    
    // Adicionar o + se não existir
    if (!normalized.startsWith('+')) {
      normalized = '+' + normalized;
    }
    
    return normalized;
  }
  
  /**
   * Envia uma notificação por SMS para múltiplos destinatários
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
    
    this.logger.info(`Iniciando envio em lote de ${recipients.length} SMS`, {
      recipientCount: recipients.length,
      batchId
    });
    
    // Definir limite de mensagens por minuto baseado na configuração
    const messagesPerMinute = this.config.throttling?.maxMessagesPerMinute || 100;
    const messagesPerSecond = this.config.throttling?.maxMessagesPerSecond || 10;
    
    // Calcular intervalo entre mensagens para respeitar o limite
    const intervalMs = 1000 / messagesPerSecond;
    
    let sentInCurrentMinute = 0;
    let minuteStartTime = Date.now();
    
    for (const recipient of recipients) {
      // Verificar se atingimos o limite por minuto
      if (sentInCurrentMinute >= messagesPerMinute) {
        const elapsed = Date.now() - minuteStartTime;
        if (elapsed < 60000) {
          // Aguardar até completar um minuto
          await this.delay(60000 - elapsed);
        }
        // Reiniciar contador de minuto
        sentInCurrentMinute = 0;
        minuteStartTime = Date.now();
      }
      
      // Enviar a mensagem
      const result = await this.send(recipient, content, event, {
        ...options,
        notificationId: `${options?.notificationId || 'sms'}-${recipient.id}-${Date.now()}`,
        tracking: {
          ...(options?.tracking || {}),
          metadata: {
            ...(options?.tracking?.metadata || {}),
            batchId
          }
        }
      });
      
      results.push(result);
      sentInCurrentMinute++;
      
      // Aguardar o intervalo calculado para respeitar o limite por segundo
      await this.delay(intervalMs);
    }
    
    // Contabilizar resultados
    const successCount = results.filter(r => r.success).length;
    const failCount = results.length - successCount;
    
    this.logger.info(`Concluído envio em lote de SMS: ${successCount} sucesso, ${failCount} falha`, {
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
    // A maioria dos provedores de SMS não permite cancelamento após o envio
    this.logger.warn(`Tentativa de cancelamento de SMS ${notificationId}. SMS geralmente não podem ser cancelados após enviados.`);
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
    // Implementação básica - em sistemas reais, consultaria a API do provedor
    if (this.config.provider === 'TWILIO' && this.twilioClient && notificationId.startsWith('SM')) {
      try {
        // Tentar buscar o status de uma mensagem Twilio
        const message = await this.twilioClient.messages(notificationId).fetch();
        
        let status: 'SCHEDULED' | 'SENT' | 'DELIVERED' | 'FAILED' | 'OPENED' | 'CLICKED' | 'EXPIRED' | 'CANCELLED';
        
        // Mapear status do Twilio para nosso formato
        switch (message.status) {
          case 'delivered':
            status = 'DELIVERED';
            break;
          case 'failed':
          case 'undelivered':
            status = 'FAILED';
            break;
          case 'sent':
            status = 'SENT';
            break;
          case 'queued':
            status = 'SCHEDULED';
            break;
          default:
            status = 'SENT';
        }
        
        return {
          status,
          timestamp: new Date(message.dateUpdated),
          details: {
            provider: 'TWILIO',
            providerStatus: message.status,
            errorCode: message.errorCode,
            errorMessage: message.errorMessage
          }
        };
      } catch (error) {
        this.logger.error(`Erro ao obter status do SMS ${notificationId}: ${error}`);
      }
    }
    
    // Se não conseguir obter o status real, retornar um status padrão
    return {
      status: 'SENT', // Status padrão assumido
      timestamp: new Date(),
      details: {
        note: 'Status real depende de webhooks de provedores ou rastreamento'
      }
    };
  }
  
  /**
   * Cria um delay (promise) por um tempo específico
   * @param ms Tempo em milissegundos
   */
  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}