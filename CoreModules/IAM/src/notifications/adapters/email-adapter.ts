/**
 * @file email-adapter.ts
 * @description Implementação de adaptador de notificação por email
 * 
 * Este adaptador permite o envio de emails utilizando diferentes provedores
 * (SMTP, SendGrid, AWS SES, etc.), com suporte a múltiplos formatos,
 * anexos, rastreamento de abertura e clique.
 */

import * as nodemailer from 'nodemailer';
import { SES } from 'aws-sdk';
import axios from 'axios';
import { createReadStream } from 'fs';
import { extname } from 'path';
import { Logger } from '../../../infrastructure/observability/logger';
import { 
  NotificationAdapter, 
  NotificationResult, 
  NotificationRecipient,
  NotificationContent,
  BaseNotificationOptions,
  Attachment
} from './notification-adapter';
import { BaseEvent } from '../core/base-event';
import { NotificationChannel, ChannelConfig } from '../core/notification-channel';

/**
 * Configuração do adaptador de email
 */
export interface EmailAdapterConfig extends ChannelConfig {
  /**
   * Tipo de provedor de email a ser utilizado
   */
  provider: 'SMTP' | 'SENDGRID' | 'SES' | 'POSTMARK' | 'MAILGUN' | 'MOCK';
  
  /**
   * Configurações específicas para SMTP
   */
  smtp?: {
    host: string;
    port: number;
    secure: boolean; // true para 465, false para outros portos
    auth: {
      user: string;
      pass: string;
    };
    tls?: {
      rejectUnauthorized?: boolean;
      ciphers?: string;
    };
  };
  
  /**
   * Configurações específicas para SendGrid
   */
  sendgrid?: {
    apiKey: string;
  };
  
  /**
   * Configurações específicas para Amazon SES
   */
  ses?: {
    region: string;
    accessKeyId?: string;
    secretAccessKey?: string;
    apiVersion?: string;
  };
  
  /**
   * Configurações específicas para Postmark
   */
  postmark?: {
    serverToken: string;
  };
  
  /**
   * Configurações específicas para Mailgun
   */
  mailgun?: {
    apiKey: string;
    domain: string;
    host?: string;
  };
  
  /**
   * Endereço de email padrão para o remetente
   */
  defaultSender: string;
  
  /**
   * Nome do remetente padrão
   */
  defaultSenderName?: string;
  
  /**
   * Endereços de email para receber cópia (CC)
   */
  defaultCc?: string[];
  
  /**
   * Endereços de email para receber cópia oculta (BCC)
   */
  defaultBcc?: string[];
  
  /**
   * Endereço de email para respostas (reply-to)
   */
  defaultReplyTo?: string;
  
  /**
   * URL base para links de rastreamento
   */
  trackingBaseUrl?: string;
  
  /**
   * Habilita rastreamento de abertura de email
   */
  enableOpenTracking?: boolean;
  
  /**
   * Habilita rastreamento de clique em links
   */
  enableClickTracking?: boolean;
  
  /**
   * Prefixo para assunto dos emails (ex: [INNOVABIZ])
   */
  subjectPrefix?: string;
  
  /**
   * Rodapé padrão para emails
   */
  defaultFooter?: string;
  
  /**
   * Limites para throttling de envio
   */
  throttling?: {
    maxEmailsPerSecond: number;
    maxEmailsPerMinute: number;
    maxEmailsPerHour: number;
  };
}

/**
 * Implementação de adaptador de notificação para canal de email
 */
export class EmailAdapter implements NotificationAdapter {
  readonly channelType = NotificationChannel.EMAIL;
  
  private transporter?: nodemailer.Transporter;
  private sesClient?: SES;
  private config: EmailAdapterConfig;
  private initialized = false;
  private logger = new Logger('EmailAdapter');
  
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
  async initialize(config: EmailAdapterConfig): Promise<void> {
    this.config = config;
    
    try {
      switch (config.provider) {
        case 'SMTP':
          if (!config.smtp) {
            throw new Error('Configuração SMTP não fornecida');
          }
          
          this.transporter = nodemailer.createTransport(config.smtp);
          
          // Verificar conexão
          await this.transporter.verify();
          break;
          
        case 'SENDGRID':
          if (!config.sendgrid) {
            throw new Error('Configuração SendGrid não fornecida');
          }
          
          this.transporter = nodemailer.createTransport({
            host: 'smtp.sendgrid.net',
            port: 587,
            secure: false,
            auth: {
              user: 'apikey',
              pass: config.sendgrid.apiKey
            }
          });
          break;
          
        case 'SES':
          if (!config.ses) {
            throw new Error('Configuração Amazon SES não fornecida');
          }
          
          this.sesClient = new SES({
            region: config.ses.region,
            accessKeyId: config.ses.accessKeyId,
            secretAccessKey: config.ses.secretAccessKey,
            apiVersion: config.ses.apiVersion || '2010-12-01'
          });
          
          // Para compatibilidade com nodemailer
          this.transporter = nodemailer.createTransport({
            SES: this.sesClient
          });
          break;
          
        case 'POSTMARK':
          if (!config.postmark) {
            throw new Error('Configuração Postmark não fornecida');
          }
          
          this.transporter = nodemailer.createTransport({
            host: 'smtp.postmarkapp.com',
            port: 587,
            secure: false,
            auth: {
              user: config.postmark.serverToken,
              pass: config.postmark.serverToken
            }
          });
          break;
          
        case 'MAILGUN':
          if (!config.mailgun) {
            throw new Error('Configuração Mailgun não fornecida');
          }
          
          this.transporter = nodemailer.createTransport({
            host: config.mailgun.host || 'smtp.mailgun.org',
            port: 587,
            secure: false,
            auth: {
              user: 'postmaster@' + config.mailgun.domain,
              pass: config.mailgun.apiKey
            }
          });
          break;
          
        case 'MOCK':
          // Criar um transportador de teste que não envia emails reais
          this.transporter = nodemailer.createTransport({
            jsonTransport: true
          });
          break;
          
        default:
          throw new Error(`Provedor de email não suportado: ${config.provider}`);
      }
      
      this.initialized = true;
      this.logger.info(`Adaptador de email inicializado com sucesso: ${config.provider}`);
    } catch (error) {
      this.initialized = false;
      this.logger.error(`Falha ao inicializar adaptador de email: ${error}`);
      throw error;
    }
  }
  
  /**
   * Envia uma notificação por email para um único destinatário
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
    if (!this.initialized || !this.transporter) {
      return {
        success: false,
        errorMessage: 'Adaptador de email não inicializado',
        errorCode: 'EMAIL_ADAPTER_NOT_INITIALIZED',
        timestamp: new Date()
      };
    }
    
    const startTime = Date.now();
    const notificationId = options?.notificationId || `email-${Date.now()}-${Math.random().toString(36).substring(2, 10)}`;
    
    try {
      // Verificar se o destinatário tem um endereço de email válido
      const emails = recipient.addresses?.get(NotificationChannel.EMAIL);
      
      if (!emails || emails.length === 0) {
        return {
          success: false,
          notificationId,
          errorMessage: `Destinatário ${recipient.id} não possui endereço de email`,
          errorCode: 'EMAIL_ADDRESS_MISSING',
          timestamp: new Date()
        };
      }
      
      // Usar o primeiro endereço de email disponível
      const toEmail = emails[0];
      
      // Preparar o assunto do email
      let subject = content.title || 'Notificação InnovaBiz';
      if (this.config.subjectPrefix) {
        subject = `${this.config.subjectPrefix} ${subject}`;
      }
      
      // Preparar o remetente
      const from = this.formatSender(
        options?.sender?.email || this.config.defaultSender,
        options?.sender?.name || this.config.defaultSenderName
      );
      
      // Preparar conteúdo do email com rastreamento
      const { html, text } = await this.prepareEmailContent(content, recipient, notificationId);
      
      // Preparar anexos
      const attachments = await this.prepareAttachments(content.attachments);
      
      // Configurar opções do email
      const mailOptions: nodemailer.SendMailOptions = {
        from,
        to: this.formatRecipient(toEmail, recipient.name),
        subject,
        html,
        text,
        attachments,
        headers: {
          'X-Notification-ID': notificationId,
          'X-SG-EID': notificationId, // Para compatibilidade com SendGrid
          'X-Message-ID': notificationId
        }
      };
      
      // Adicionar CC se configurado
      if (options?.cc?.length) {
        mailOptions.cc = options.cc;
      } else if (this.config.defaultCc?.length) {
        mailOptions.cc = this.config.defaultCc;
      }
      
      // Adicionar BCC se configurado
      if (options?.bcc?.length) {
        mailOptions.bcc = options.bcc;
      } else if (this.config.defaultBcc?.length) {
        mailOptions.bcc = this.config.defaultBcc;
      }
      
      // Adicionar Reply-To se configurado
      if (options?.replyTo) {
        mailOptions.replyTo = options.replyTo;
      } else if (this.config.defaultReplyTo) {
        mailOptions.replyTo = this.config.defaultReplyTo;
      }
      
      // Enviar o email
      const info = await this.transporter.sendMail(mailOptions);
      
      const messageId = this.extractMessageId(info);
      const deliveryTime = Date.now() - startTime;
      
      this.logger.info(`Email enviado para ${toEmail} em ${deliveryTime}ms`, {
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
          email: toEmail,
          subject,
          deliveryTime
        },
        timestamp: new Date()
      };
    } catch (error) {
      this.logger.error(`Erro ao enviar email para ${recipient.id}: ${error}`, {
        notificationId,
        recipientId: recipient.id,
        error
      });
      
      return {
        success: false,
        notificationId,
        errorMessage: `Erro ao enviar email: ${error.message || error}`,
        errorCode: 'EMAIL_SEND_FAILED',
        timestamp: new Date()
      };
    }
  }
  
  /**
   * Formata o remetente do email
   * @param email Endereço de email
   * @param name Nome (opcional)
   */
  private formatSender(email: string, name?: string): string {
    if (name) {
      return `"${name.replace(/"/g, '\\"')}" <${email}>`;
    }
    return email;
  }
  
  /**
   * Formata o destinatário do email
   * @param email Endereço de email
   * @param name Nome (opcional)
   */
  private formatRecipient(email: string, name?: string): string {
    if (name) {
      return `"${name.replace(/"/g, '\\"')}" <${email}>`;
    }
    return email;
  }
  
  /**
   * Extrai o ID da mensagem do resultado de envio
   * @param info Informações de envio do email
   */
  private extractMessageId(info: any): string {
    if (this.config.provider === 'MOCK') {
      return `mock-${Date.now()}`;
    }
    
    // Tentar extrair o ID da mensagem de várias formas possíveis
    return info.messageId || 
           info.response?.id || 
           info.response?.MessageId ||
           info.MessageId ||
           `${this.config.provider.toLowerCase()}-${Date.now()}`;
  }
  
  /**
   * Prepara o conteúdo do email com rastreamento
   * @param content Conteúdo da notificação
   * @param recipient Destinatário da notificação
   * @param notificationId ID da notificação
   */
  private async prepareEmailContent(
    content: NotificationContent,
    recipient: NotificationRecipient,
    notificationId: string
  ): Promise<{ html: string; text: string }> {
    // Texto simples (fallback)
    let text = content.body || '';
    
    // Adicionar rodapé ao texto
    if (this.config.defaultFooter) {
      text += `\n\n${this.config.defaultFooter}`;
    }
    
    // Versão HTML
    let html: string;
    
    // Se o conteúdo já for HTML, usá-lo diretamente
    if (content.format === 'HTML') {
      html = content.body;
    } else {
      // Converter texto para HTML
      html = this.textToHtml(content.body);
    }
    
    // Adicionar rastreamento de abertura se habilitado
    if (this.config.enableOpenTracking && this.config.trackingBaseUrl) {
      const trackingPixel = `<img src="${this.config.trackingBaseUrl}/t/open/${notificationId}" width="1" height="1" alt="" style="display:none;"/>`;
      html = html.replace(/<\/body>/, `${trackingPixel}</body>`);
    }
    
    // Adicionar rastreamento de clique se habilitado
    if (this.config.enableClickTracking && this.config.trackingBaseUrl && content.actions) {
      for (const action of content.actions) {
        if (action.url) {
          const trackingUrl = `${this.config.trackingBaseUrl}/t/click/${notificationId}?u=${encodeURIComponent(action.url)}&aid=${action.id}`;
          
          // Substituir URLs no HTML
          html = html.replace(
            new RegExp(`href=["']${action.url}["']`, 'g'),
            `href="${trackingUrl}"`
          );
          
          // Substituir URLs no texto
          text = text.replace(
            new RegExp(action.url, 'g'),
            trackingUrl
          );
        }
      }
    }
    
    // Adicionar rodapé HTML
    if (this.config.defaultFooter) {
      const footerHtml = `<div style="margin-top: 20px; padding-top: 20px; border-top: 1px solid #eee; color: #666; font-size: 12px;">${this.textToHtml(this.config.defaultFooter)}</div>`;
      html = html.replace(/<\/body>/, `${footerHtml}</body>`);
    }
    
    return { html, text };
  }
  
  /**
   * Converte texto simples para HTML básico
   * @param text Texto para converter
   */
  private textToHtml(text: string): string {
    if (!text) return '';
    
    return `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Notificação InnovaBiz</title>
  <style>
    body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
    a { color: #0066cc; }
    .container { max-width: 600px; margin: 0 auto; padding: 20px; }
  </style>
</head>
<body>
  <div class="container">
    ${text.replace(/\n/g, '<br>')}
  </div>
</body>
</html>`;
  }
  
  /**
   * Prepara os anexos para envio
   * @param attachments Anexos da notificação
   */
  private async prepareAttachments(
    attachments?: Attachment[]
  ): Promise<nodemailer.Attachment[]> {
    if (!attachments || attachments.length === 0) {
      return [];
    }
    
    const result: nodemailer.Attachment[] = [];
    
    for (const attachment of attachments) {
      try {
        if (typeof attachment.content === 'string') {
          // Conteúdo em texto ou base64
          result.push({
            filename: attachment.filename,
            content: attachment.content,
            contentType: attachment.contentType
          });
        } else if (Buffer.isBuffer(attachment.content)) {
          // Conteúdo em Buffer
          result.push({
            filename: attachment.filename,
            content: attachment.content,
            contentType: attachment.contentType
          });
        } else if (typeof attachment.content === 'object' && attachment.content.path) {
          // Caminho para arquivo
          result.push({
            filename: attachment.filename,
            path: attachment.content.path,
            contentType: attachment.contentType
          });
        }
      } catch (error) {
        this.logger.error(`Erro ao processar anexo ${attachment.filename}: ${error}`);
      }
    }
    
    return result;
  }
  
  /**
   * Envia uma notificação por email para múltiplos destinatários
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
    
    this.logger.info(`Iniciando envio em lote de ${recipients.length} emails`, {
      recipientCount: recipients.length,
      batchId
    });
    
    // Aplicar limites de throttling baseados na configuração
    const messagesPerMinute = this.config.throttling?.maxEmailsPerMinute || 100;
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
      
      // Enviar o email
      const result = await this.send(recipient, content, event, {
        ...options,
        notificationId: `${options?.notificationId || 'email'}-${recipient.id}-${Date.now()}`,
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
    
    this.logger.info(`Concluído envio em lote de emails: ${successCount} sucesso, ${failCount} falha`, {
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
    // A maioria dos provedores de email não permite cancelamento após o envio
    this.logger.warn(`Tentativa de cancelamento de email ${notificationId}. Emails geralmente não podem ser cancelados após enviados.`);
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
    // Status de emails geralmente só pode ser obtido via webhooks
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