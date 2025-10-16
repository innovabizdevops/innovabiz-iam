/**
 * @file webhook-adapter.ts
 * @description Implementação de adaptador de notificação via webhook
 * 
 * Este adaptador permite o envio de notificações via webhook para sistemas
 * externos, APIs e serviços de integração.
 */

import axios, { AxiosRequestConfig } from 'axios';
import { createHmac } from 'crypto';
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
 * Configuração do adaptador webhook
 */
export interface WebhookAdapterConfig extends ChannelConfig {
  /**
   * Timeout para requisições (ms)
   */
  requestTimeoutMs?: number;
  
  /**
   * Número de tentativas em caso de falha
   */
  maxRetries?: number;
  
  /**
   * Intervalo entre tentativas (ms)
   */
  retryIntervalMs?: number;
  
  /**
   * Cabeçalhos HTTP padrão
   */
  defaultHeaders?: Record<string, string>;
  
  /**
   * Método HTTP padrão
   */
  defaultMethod?: 'POST' | 'PUT' | 'PATCH';
  
  /**
   * Secret para assinatura das requisições
   */
  signingSecret?: string;
  
  /**
   * Nome do cabeçalho para assinatura
   */
  signatureHeaderName?: string;
  
  /**
   * Algoritmo de hash para assinatura
   */
  signatureAlgorithm?: 'sha256' | 'sha512' | 'md5';
  
  /**
   * Formato do payload padrão
   */
  payloadFormat?: 'JSON' | 'FORM' | 'XML';
  
  /**
   * URL base para fallback se o destinatário não tiver URL
   */
  fallbackWebhookUrl?: string;
}

/**
 * Implementação de adaptador de notificação para webhook
 */
export class WebhookAdapter implements NotificationAdapter {
  readonly channelType = NotificationChannel.WEBHOOK;
  
  private config: WebhookAdapterConfig;
  private initialized = false;
  private logger = new Logger('WebhookAdapter');
  
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
  async initialize(config: WebhookAdapterConfig): Promise<void> {
    this.config = config;
    
    // Validar configuração
    if (!config.fallbackWebhookUrl) {
      this.logger.warn('Nenhuma URL de fallback configurada para o adaptador webhook');
    }
    
    this.initialized = true;
    this.logger.info('Adaptador webhook inicializado com sucesso');
  }
  
  /**
   * Envia uma notificação via webhook
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
        errorMessage: 'Adaptador webhook não inicializado',
        errorCode: 'WEBHOOK_ADAPTER_NOT_INITIALIZED',
        timestamp: new Date()
      };
    }
    
    const startTime = Date.now();
    const notificationId = options?.notificationId || `webhook-${Date.now()}-${Math.random().toString(36).substring(2, 10)}`;
    
    try {
      // Obter URL do webhook do destinatário
      const webhookUrls = recipient.addresses?.get(NotificationChannel.WEBHOOK);
      let targetUrl: string;
      
      if (!webhookUrls || webhookUrls.length === 0) {
        // Se o destinatário não tiver URL, usar fallback
        if (!this.config.fallbackWebhookUrl) {
          return {
            success: false,
            notificationId,
            errorMessage: `Destinatário ${recipient.id} não possui URL de webhook e nenhum fallback configurado`,
            errorCode: 'WEBHOOK_URL_MISSING',
            timestamp: new Date()
          };
        }
        
        targetUrl = this.config.fallbackWebhookUrl;
      } else {
        // Usar a primeira URL configurada
        targetUrl = webhookUrls[0];
      }
      
      // Preparar payload
      const payload = this.prepareWebhookPayload(recipient, content, event, notificationId, options);
      
      // Preparar configurações da requisição
      const requestConfig: AxiosRequestConfig = {
        method: this.config.defaultMethod || 'POST',
        url: targetUrl,
        timeout: this.config.requestTimeoutMs || 5000,
        headers: {
          'Content-Type': 'application/json',
          'User-Agent': 'InnovaBiz-Notification-Service/1.0',
          'X-Notification-ID': notificationId,
          ...this.config.defaultHeaders
        }
      };
      
      // Definir formato do payload
      switch (this.config.payloadFormat) {
        case 'XML':
          requestConfig.headers!['Content-Type'] = 'application/xml';
          requestConfig.data = this.convertToXml(payload);
          break;
        case 'FORM':
          requestConfig.headers!['Content-Type'] = 'application/x-www-form-urlencoded';
          requestConfig.data = new URLSearchParams(this.flattenObject(payload)).toString();
          break;
        case 'JSON':
        default:
          requestConfig.headers!['Content-Type'] = 'application/json';
          requestConfig.data = payload;
      }
      
      // Adicionar assinatura se configurado
      if (this.config.signingSecret && this.config.signatureHeaderName) {
        const signature = this.generateSignature(
          JSON.stringify(payload),
          this.config.signingSecret,
          this.config.signatureAlgorithm || 'sha256'
        );
        requestConfig.headers![this.config.signatureHeaderName] = signature;
      }
      
      // Enviar webhook com retentativas
      const response = await this.sendWithRetries(
        requestConfig,
        this.config.maxRetries || 3,
        this.config.retryIntervalMs || 1000
      );
      
      const deliveryTime = Date.now() - startTime;
      this.logger.info(`Webhook enviado para ${targetUrl} em ${deliveryTime}ms`, {
        statusCode: response.status,
        notificationId,
        recipientId: recipient.id
      });
      
      return {
        success: true,
        notificationId,
        details: {
          statusCode: response.status,
          responseData: response.data,
          url: targetUrl,
          deliveryTime
        },
        timestamp: new Date()
      };
    } catch (error) {
      this.logger.error(`Erro ao enviar webhook para ${recipient.id}: ${error}`, {
        notificationId,
        recipientId: recipient.id,
        error
      });
      
      return {
        success: false,
        notificationId,
        errorMessage: `Erro ao enviar webhook: ${error.message || error}`,
        errorCode: error.response?.status ? `HTTP_ERROR_${error.response.status}` : 'WEBHOOK_SEND_FAILED',
        details: error.response ? {
          statusCode: error.response.status,
          responseData: error.response.data
        } : undefined,
        timestamp: new Date()
      };
    }
  }
  
  /**
   * Prepara o payload do webhook
   * @param recipient Destinatário da notificação
   * @param content Conteúdo da notificação
   * @param event Evento que originou a notificação
   * @param notificationId ID da notificação
   * @param options Opções de envio
   */
  private prepareWebhookPayload(
    recipient: NotificationRecipient,
    content: NotificationContent,
    event?: BaseEvent,
    notificationId?: string,
    options?: BaseNotificationOptions
  ): Record<string, any> {
    // Dados básicos da notificação
    const payload: Record<string, any> = {
      notification: {
        id: notificationId,
        timestamp: new Date().toISOString(),
        title: content.title,
        body: content.body,
        format: content.format || 'TEXT'
      },
      recipient: {
        id: recipient.id,
        type: recipient.type
      }
    };
    
    // Adicionar dados do evento se existir
    if (event) {
      payload.event = {
        id: event.eventId,
        code: event.code,
        type: event.eventType,
        source: event.source,
        timestamp: event.timestamp.toISOString(),
        data: event.data
      };
    }
    
    // Adicionar metadados de rastreamento
    if (options?.tracking) {
      payload.tracking = options.tracking;
    }
    
    // Adicionar ações se existirem
    if (content.actions && content.actions.length > 0) {
      payload.notification.actions = content.actions;
    }
    
    // Adicionar anexos se existirem
    if (content.attachments && content.attachments.length > 0) {
      payload.notification.attachments = content.attachments.map(a => ({
        filename: a.filename,
        contentType: a.contentType,
        // Não incluir o conteúdo binário, apenas metadados
        size: typeof a.content === 'string' ? a.content.length : 
          Buffer.isBuffer(a.content) ? a.content.length : 0
      }));
    }
    
    // Adicionar recursos se existirem
    if (content.resourceUrls && content.resourceUrls.length > 0) {
      payload.notification.resources = content.resourceUrls;
    }
    
    return payload;
  }
  
  /**
   * Envia uma requisição com sistema de retentativas
   * @param config Configuração da requisição
   * @param maxRetries Número máximo de tentativas
   * @param retryIntervalMs Intervalo entre tentativas
   */
  private async sendWithRetries(
    config: AxiosRequestConfig,
    maxRetries: number,
    retryIntervalMs: number
  ): Promise<any> {
    let lastError: any;
    
    for (let attempt = 0; attempt <= maxRetries; attempt++) {
      try {
        // Tentar enviar a requisição
        return await axios(config);
      } catch (error) {
        lastError = error;
        
        // Verificar se é um erro que pode ser retentado
        const status = error.response?.status;
        const retryable = !status || // Erro de rede
          status >= 500 || // Erro de servidor
          status === 429 || // Rate limit
          status === 408; // Timeout
        
        if (!retryable || attempt === maxRetries) {
          // Não pode ser retentado ou alcançou o limite de tentativas
          break;
        }
        
        // Calcular intervalo de espera com backoff exponencial
        const waitTime = retryIntervalMs * Math.pow(2, attempt);
        
        this.logger.warn(`Tentativa ${attempt + 1} falhou, tentando novamente em ${waitTime}ms`, {
          url: config.url,
          status,
          error: error.message
        });
        
        // Aguardar antes de tentar novamente
        await this.delay(waitTime);
      }
    }
    
    // Se chegou aqui, todas as tentativas falharam
    throw lastError;
  }
  
  /**
   * Gera uma assinatura para o payload
   * @param payload Payload a ser assinado
   * @param secret Secret para assinatura
   * @param algorithm Algoritmo de hash
   */
  private generateSignature(
    payload: string,
    secret: string,
    algorithm: 'sha256' | 'sha512' | 'md5'
  ): string {
    return createHmac(algorithm, secret)
      .update(payload)
      .digest('hex');
  }
  
  /**
   * Converte objeto para XML (implementação simplificada)
   * @param obj Objeto a ser convertido
   */
  private convertToXml(obj: any): string {
    const convert = (item: any, key: string): string => {
      if (item === null || item === undefined) {
        return `<${key}></${key}>`;
      }
      
      if (Array.isArray(item)) {
        return item.map(i => convert(i, 'item')).join('');
      }
      
      if (typeof item === 'object') {
        const content = Object.entries(item)
          .map(([k, v]) => convert(v, k))
          .join('');
        return `<${key}>${content}</${key}>`;
      }
      
      return `<${key}>${String(item).replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&apos;')}</${key}>`;
    };
    
    return `<?xml version="1.0" encoding="UTF-8"?><root>${
      Object.entries(obj).map(([key, value]) => convert(value, key)).join('')
    }</root>`;
  }
  
  /**
   * Converte objeto aninhado para objeto plano (para form data)
   * @param obj Objeto a ser achatado
   */
  private flattenObject(obj: any, prefix: string = ''): Record<string, string> {
    return Object.entries(obj).reduce((acc: Record<string, string>, [key, value]) => {
      const newKey = prefix ? `${prefix}[${key}]` : key;
      
      if (value === null || value === undefined) {
        acc[newKey] = '';
      } else if (typeof value === 'object' && !Array.isArray(value)) {
        Object.assign(acc, this.flattenObject(value, newKey));
      } else if (Array.isArray(value)) {
        value.forEach((item, index) => {
          if (typeof item === 'object') {
            Object.assign(acc, this.flattenObject(item, `${newKey}[${index}]`));
          } else {
            acc[`${newKey}[${index}]`] = String(item);
          }
        });
      } else {
        acc[newKey] = String(value);
      }
      
      return acc;
    }, {});
  }
  
  /**
   * Envia notificações para múltiplos destinatários
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
    
    this.logger.info(`Iniciando envio em lote de ${recipients.length} webhooks`, {
      recipientCount: recipients.length,
      batchId
    });
    
    // Agrupar destinatários por URL para otimizar envios
    const recipientsByUrl = new Map<string, NotificationRecipient[]>();
    
    for (const recipient of recipients) {
      const webhookUrls = recipient.addresses?.get(NotificationChannel.WEBHOOK);
      const url = webhookUrls && webhookUrls.length > 0 
        ? webhookUrls[0] 
        : this.config.fallbackWebhookUrl;
      
      if (!url) {
        // Destinatário sem URL e sem fallback
        results.push({
          success: false,
          notificationId: `webhook-${recipient.id}-${Date.now()}`,
          errorMessage: `Destinatário ${recipient.id} não possui URL de webhook e nenhum fallback configurado`,
          errorCode: 'WEBHOOK_URL_MISSING',
          timestamp: new Date()
        });
        continue;
      }
      
      const urlRecipients = recipientsByUrl.get(url) || [];
      urlRecipients.push(recipient);
      recipientsByUrl.set(url, urlRecipients);
    }
    
    // Para cada URL, enviar uma requisição com todos os destinatários
    for (const [url, urlRecipients] of recipientsByUrl.entries()) {
      if (urlRecipients.length === 1) {
        // Para um único destinatário, usar o método normal
        const result = await this.send(urlRecipients[0], content, event, {
          ...options,
          notificationId: `${options?.notificationId || 'webhook'}-${urlRecipients[0].id}-${Date.now()}`,
          tracking: {
            ...(options?.tracking || {}),
            metadata: {
              ...(options?.tracking?.metadata || {}),
              batchId
            }
          }
        });
        results.push(result);
      } else {
        // Para múltiplos destinatários com a mesma URL, enviar uma única requisição
        try {
          const bulkNotificationId = `webhook-bulk-${Date.now()}-${Math.random().toString(36).substring(2, 10)}`;
          
          // Preparar payload com dados de todos os destinatários
          const payload = {
            notification: {
              id: bulkNotificationId,
              timestamp: new Date().toISOString(),
              title: content.title,
              body: content.body,
              format: content.format || 'TEXT',
              batch: true,
              batchSize: urlRecipients.length
            },
            recipients: urlRecipients.map(r => ({
              id: r.id,
              type: r.type,
              name: r.name,
              metadata: r.metadata
            }))
          };
          
          // Adicionar dados do evento se existir
          if (event) {
            payload.event = {
              id: event.eventId,
              code: event.code,
              type: event.eventType,
              source: event.source,
              timestamp: event.timestamp.toISOString(),
              data: event.data
            };
          }
          
          // Preparar configurações da requisição
          const requestConfig: AxiosRequestConfig = {
            method: this.config.defaultMethod || 'POST',
            url,
            timeout: this.config.requestTimeoutMs || 5000,
            headers: {
              'Content-Type': 'application/json',
              'User-Agent': 'InnovaBiz-Notification-Service/1.0',
              'X-Notification-ID': bulkNotificationId,
              'X-Batch-ID': batchId,
              ...this.config.defaultHeaders
            },
            data: payload
          };
          
          // Adicionar assinatura se configurado
          if (this.config.signingSecret && this.config.signatureHeaderName) {
            const signature = this.generateSignature(
              JSON.stringify(payload),
              this.config.signingSecret,
              this.config.signatureAlgorithm || 'sha256'
            );
            requestConfig.headers![this.config.signatureHeaderName] = signature;
          }
          
          // Enviar webhook com retentativas
          const response = await this.sendWithRetries(
            requestConfig,
            this.config.maxRetries || 3,
            this.config.retryIntervalMs || 1000
          );
          
          // Criar um resultado de sucesso para cada destinatário
          for (const recipient of urlRecipients) {
            results.push({
              success: true,
              notificationId: `${bulkNotificationId}-${recipient.id}`,
              details: {
                statusCode: response.status,
                url,
                batchId,
                bulkNotificationId,
                batchSize: urlRecipients.length
              },
              timestamp: new Date()
            });
          }
        } catch (error) {
          // Criar um resultado de falha para cada destinatário
          for (const recipient of urlRecipients) {
            results.push({
              success: false,
              notificationId: `webhook-${recipient.id}-${Date.now()}`,
              errorMessage: `Erro ao enviar webhook em lote: ${error.message || error}`,
              errorCode: error.response?.status ? `HTTP_ERROR_${error.response.status}` : 'WEBHOOK_SEND_FAILED',
              details: error.response ? {
                statusCode: error.response.status,
                url,
                batchId
              } : { url, batchId },
              timestamp: new Date()
            });
          }
        }
      }
    }
    
    // Contabilizar resultados
    const successCount = results.filter(r => r.success).length;
    const failCount = results.length - successCount;
    
    this.logger.info(`Concluído envio em lote de webhooks: ${successCount} sucesso, ${failCount} falha`, {
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
    // Webhooks não podem ser cancelados após o envio
    this.logger.warn(`Tentativa de cancelamento de webhook ${notificationId}. Webhooks não podem ser cancelados após enviados.`);
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
    // Para webhooks, consideramos entregue quando a requisição foi bem-sucedida
    return {
      status: 'DELIVERED',
      timestamp: new Date(),
      details: {
        note: 'Status de webhook é baseado apenas no resultado da requisição HTTP'
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