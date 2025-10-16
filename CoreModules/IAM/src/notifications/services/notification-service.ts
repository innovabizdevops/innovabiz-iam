/**
 * @file notification-service.ts
 * @description Serviço de notificação responsável por orquestrar os adaptadores
 * 
 * Este serviço centraliza a lógica de envio de notificações, gerencia preferências
 * de usuários, priorização de canais, agendamento, retry e análise de entrega.
 */

import { v4 as uuidv4 } from 'uuid';
import { Logger } from '../../../infrastructure/observability/logger';
import { 
  NotificationAdapter, 
  NotificationResult,
  NotificationRecipient,
  NotificationContent,
  BaseNotificationOptions
} from '../adapters/notification-adapter';
import { NotificationChannel } from '../core/notification-channel';
import { NotificationAdapterFactory } from '../adapters/notification-adapter-factory';
import { BaseEvent } from '../core/base-event';

/**
 * Opções de envio para o serviço de notificação
 */
export interface NotificationServiceOptions extends BaseNotificationOptions {
  /**
   * Canais prioritários para envio
   */
  preferredChannels?: NotificationChannel[];
  
  /**
   * Se verdadeiro, tenta todos os canais disponíveis até sucesso
   */
  fallbackEnabled?: boolean;
  
  /**
   * Número máximo de tentativas para cada canal
   */
  maxRetriesPerChannel?: number;
  
  /**
   * Intervalo de tempo entre tentativas (ms)
   */
  retryIntervalMs?: number;
  
  /**
   * Intervalo de backoff entre tentativas
   */
  retryBackoffFactor?: number;
  
  /**
   * Se verdadeiro, envia para todos os canais simultâneamente
   */
  sendToAllChannels?: boolean;
  
  /**
   * Agendamento para envio futuro
   */
  schedule?: {
    /**
     * Data e hora para envio
     */
    scheduledTime: Date;
    
    /**
     * Timezone do agendamento
     */
    timezone?: string;
  };
  
  /**
   * Dados para rastreamento da notificação
   */
  tracking?: {
    /**
     * Origem da notificação
     */
    source?: string;
    
    /**
     * Categoria da notificação
     */
    category?: string;
    
    /**
     * Tags para categorização
     */
    tags?: string[];
    
    /**
     * Metadados adicionais
     */
    metadata?: Record<string, any>;
  };
  
  /**
   * Opções específicas por canal
   */
  channelOptions?: Partial<Record<NotificationChannel, Record<string, any>>>;
  
  /**
   * Condições para expiração da notificação
   */
  expiryConditions?: {
    /**
     * Expirar após este timestamp
     */
    expiryTime?: Date;
    
    /**
     * Expirar após este número de horas
     */
    ttlHours?: number;
  };
}

/**
 * Resultado de envio de notificação pelo serviço
 */
export interface NotificationServiceResult {
  /**
   * Identificador único da notificação
   */
  notificationId: string;
  
  /**
   * Sucesso geral do envio (true se ao menos um canal teve sucesso)
   */
  success: boolean;
  
  /**
   * Canais para os quais a notificação foi enviada com sucesso
   */
  successfulChannels: NotificationChannel[];
  
  /**
   * Canais para os quais a notificação falhou
   */
  failedChannels: NotificationChannel[];
  
  /**
   * Canais ignorados por falta de informação (ex: sem email ou telefone)
   */
  skippedChannels: NotificationChannel[];
  
  /**
   * Resultados detalhados por canal
   */
  channelResults: Record<NotificationChannel, NotificationResult>;
  
  /**
   * Timestamp da conclusão do envio
   */
  completedAt: Date;
  
  /**
   * Status final da notificação
   */
  status: 'DELIVERED' | 'PARTIAL_DELIVERY' | 'FAILED' | 'SCHEDULED' | 'CANCELLED';
}

/**
 * Serviço de notificação que gerencia o envio através dos diferentes canais
 */
export class NotificationService {
  private adapterFactory: NotificationAdapterFactory;
  private logger = new Logger('NotificationService');
  
  /**
   * Construtor
   * @param adapterFactory Fábrica de adaptadores de notificação
   */
  constructor(adapterFactory: NotificationAdapterFactory) {
    this.adapterFactory = adapterFactory;
  }
  
  /**
   * Envia uma notificação utilizando os canais apropriados
   * @param recipient Destinatário da notificação
   * @param content Conteúdo da notificação
   * @param options Opções de envio
   * @param event Evento que originou a notificação (opcional)
   */
  async send(
    recipient: NotificationRecipient,
    content: NotificationContent,
    options: NotificationServiceOptions = {},
    event?: BaseEvent
  ): Promise<NotificationServiceResult> {
    // Gerar ID único para a notificação se não fornecido
    const notificationId = options.notificationId || `notif-${uuidv4()}`;
    
    // Se agendado para o futuro, salvar na fila e retornar
    if (options.schedule && options.schedule.scheduledTime > new Date()) {
      // TODO: Implementar agendamento de notificações
      return {
        notificationId,
        success: true,
        successfulChannels: [],
        failedChannels: [],
        skippedChannels: [],
        channelResults: {},
        completedAt: new Date(),
        status: 'SCHEDULED'
      };
    }
    
    // Iniciar rastreamento de performance
    const startTime = Date.now();
    
    // Determinar canais para envio
    const channels = await this.determineTargetChannels(recipient, options);
    
    if (channels.length === 0) {
      this.logger.warn(`Nenhum canal disponível para notificação ${notificationId}`, {
        recipientId: recipient.id,
        notificationId
      });
      
      return {
        notificationId,
        success: false,
        successfulChannels: [],
        failedChannels: [],
        skippedChannels: [],
        channelResults: {},
        completedAt: new Date(),
        status: 'FAILED'
      };
    }
    
    // Registrar início do envio
    this.logger.info(`Iniciando envio de notificação ${notificationId} para ${recipient.id} via ${channels.join(', ')}`, {
      recipientId: recipient.id,
      notificationId,
      channels
    });
    
    // Resultados por canal
    const channelResults: Record<NotificationChannel, NotificationResult> = {};
    const successfulChannels: NotificationChannel[] = [];
    const failedChannels: NotificationChannel[] = [];
    const skippedChannels: NotificationChannel[] = [];
    
    // Verificar se deve enviar para todos os canais ou tentar até sucesso
    if (options.sendToAllChannels) {
      // Enviar para todos os canais simultâneamente
      const sendPromises = channels.map(channel => 
        this.sendToChannel(channel, recipient, content, notificationId, options, event)
      );
      
      const results = await Promise.allSettled(sendPromises);
      
      // Processar resultados
      for (let i = 0; i < channels.length; i++) {
        const channel = channels[i];
        const result = results[i];
        
        if (result.status === 'fulfilled') {
          if (result.value.success) {
            successfulChannels.push(channel);
            channelResults[channel] = result.value;
          } else if (result.value.errorCode === 'ADDRESS_MISSING') {
            skippedChannels.push(channel);
            channelResults[channel] = result.value;
          } else {
            failedChannels.push(channel);
            channelResults[channel] = result.value;
          }
        } else {
          failedChannels.push(channel);
          channelResults[channel] = {
            success: false,
            notificationId,
            errorMessage: result.reason?.message || 'Erro desconhecido',
            errorCode: 'SEND_ERROR',
            timestamp: new Date()
          };
        }
      }
    } else {
      // Tentar canais sequencialmente até sucesso
      let successAchieved = false;
      
      for (const channel of channels) {
        try {
          const result = await this.sendToChannel(channel, recipient, content, notificationId, options, event);
          channelResults[channel] = result;
          
          if (result.success) {
            successfulChannels.push(channel);
            successAchieved = true;
            
            // Se fallback não estiver habilitado, parar após primeiro sucesso
            if (!options.fallbackEnabled) {
              break;
            }
          } else if (result.errorCode === 'ADDRESS_MISSING') {
            skippedChannels.push(channel);
          } else {
            failedChannels.push(channel);
          }
        } catch (error) {
          failedChannels.push(channel);
          channelResults[channel] = {
            success: false,
            notificationId,
            errorMessage: error.message || 'Erro desconhecido',
            errorCode: 'SEND_ERROR',
            timestamp: new Date()
          };
        }
        
        // Parar se alcançou sucesso e fallback não está habilitado
        if (successAchieved && !options.fallbackEnabled) {
          break;
        }
      }
    }
    
    // Determinar status geral
    let status: 'DELIVERED' | 'PARTIAL_DELIVERY' | 'FAILED';
    if (successfulChannels.length === channels.length) {
      status = 'DELIVERED';
    } else if (successfulChannels.length > 0) {
      status = 'PARTIAL_DELIVERY';
    } else {
      status = 'FAILED';
    }
    
    const completedAt = new Date();
    const duration = Date.now() - startTime;
    
    // Registrar resultado
    this.logger.info(`Notificação ${notificationId} concluída em ${duration}ms - Status: ${status}`, {
      notificationId,
      recipientId: recipient.id,
      status,
      duration,
      successfulChannels,
      failedChannels,
      skippedChannels
    });
    
    return {
      notificationId,
      success: successfulChannels.length > 0,
      successfulChannels,
      failedChannels,
      skippedChannels,
      channelResults,
      completedAt,
      status
    };
  }
  
  /**
   * Envia uma notificação para múltiplos destinatários
   * @param recipients Lista de destinatários
   * @param content Conteúdo da notificação
   * @param options Opções de envio
   * @param event Evento que originou a notificação (opcional)
   */
  async sendBulk(
    recipients: NotificationRecipient[],
    content: NotificationContent,
    options: NotificationServiceOptions = {},
    event?: BaseEvent
  ): Promise<{
    batchId: string;
    results: Record<string, NotificationServiceResult>;
    summary: {
      total: number;
      success: number;
      failed: number;
      skipped: number;
    };
  }> {
    const batchId = `batch-${uuidv4()}`;
    const startTime = Date.now();
    const results: Record<string, NotificationServiceResult> = {};
    
    this.logger.info(`Iniciando envio em lote ${batchId} para ${recipients.length} destinatários`, {
      recipientCount: recipients.length,
      batchId
    });
    
    // Para grandes volumes, podemos otimizar usando sendBulk dos adaptadores
    if (recipients.length > 10) {
      try {
        // Tentar usar a funcionalidade de envio em lote dos adaptadores
        await this.sendBulkViaAdapters(recipients, content, options, event, batchId, results);
      } catch (error) {
        this.logger.error(`Erro no envio em lote otimizado: ${error}. Voltando para envio individual.`);
        // Se falhar, voltar para o método padrão
      }
    }
    
    // Se ainda não temos resultados para todos destinatários, enviar individualmente
    if (Object.keys(results).length < recipients.length) {
      for (const recipient of recipients) {
        if (!results[recipient.id]) {
          try {
            const result = await this.send(recipient, content, {
              ...options,
              tracking: {
                ...(options.tracking || {}),
                metadata: {
                  ...(options.tracking?.metadata || {}),
                  batchId
                }
              }
            }, event);
            
            results[recipient.id] = result;
          } catch (error) {
            this.logger.error(`Erro ao enviar notificação para ${recipient.id}: ${error}`);
            results[recipient.id] = {
              notificationId: `error-${uuidv4()}`,
              success: false,
              successfulChannels: [],
              failedChannels: [],
              skippedChannels: [],
              channelResults: {},
              completedAt: new Date(),
              status: 'FAILED'
            };
          }
        }
      }
    }
    
    // Gerar resumo
    const summary = {
      total: recipients.length,
      success: Object.values(results).filter(r => r.success).length,
      failed: Object.values(results).filter(r => !r.success && r.status !== 'SCHEDULED').length,
      skipped: Object.values(results).filter(r => r.status === 'SCHEDULED').length
    };
    
    const duration = Date.now() - startTime;
    
    this.logger.info(`Envio em lote ${batchId} concluído em ${duration}ms - ${summary.success}/${summary.total} sucesso`, {
      batchId,
      duration,
      summary
    });
    
    return { batchId, results, summary };
  }
  
  /**
   * Envia notificações em lote usando os adaptadores diretamente
   */
  private async sendBulkViaAdapters(
    recipients: NotificationRecipient[],
    content: NotificationContent,
    options: NotificationServiceOptions,
    event?: BaseEvent,
    batchId?: string,
    results?: Record<string, NotificationServiceResult>
  ): Promise<void> {
    // Agrupar destinatários por canal preferencial
    const recipientsByChannel = new Map<NotificationChannel, NotificationRecipient[]>();
    
    // Identificar canal preferencial para cada destinatário
    for (const recipient of recipients) {
      // Determinar canais possíveis
      const availableChannels = await this.determineTargetChannels(recipient, options);
      
      if (availableChannels.length === 0) {
        // Sem canais disponíveis
        if (results) {
          results[recipient.id] = {
            notificationId: `no-channel-${uuidv4()}`,
            success: false,
            successfulChannels: [],
            failedChannels: [],
            skippedChannels: [],
            channelResults: {},
            completedAt: new Date(),
            status: 'FAILED'
          };
        }
        continue;
      }
      
      // Usar o primeiro canal disponível
      const channel = availableChannels[0];
      
      // Agrupar destinatários por canal
      const channelRecipients = recipientsByChannel.get(channel) || [];
      channelRecipients.push(recipient);
      recipientsByChannel.set(channel, channelRecipients);
    }
    
    // Para cada canal, enviar para todos os destinatários de uma vez
    for (const [channel, channelRecipients] of recipientsByChannel.entries()) {
      try {
        const adapter = await this.adapterFactory.getAdapter(channel);
        
        // Usar o método de envio em massa do adaptador
        const bulkResults = await adapter.sendBulk(
          channelRecipients,
          content,
          event,
          {
            ...options,
            tracking: {
              ...(options.tracking || {}),
              metadata: {
                ...(options.tracking?.metadata || {}),
                batchId
              }
            }
          }
        );
        
        // Processar resultados individuais
        for (let i = 0; i < channelRecipients.length; i++) {
          const recipient = channelRecipients[i];
          const bulkResult = bulkResults[i];
          
          if (results) {
            results[recipient.id] = {
              notificationId: bulkResult.notificationId || `bulk-${uuidv4()}`,
              success: bulkResult.success,
              successfulChannels: bulkResult.success ? [channel] : [],
              failedChannels: bulkResult.success ? [] : [channel],
              skippedChannels: [],
              channelResults: { [channel]: bulkResult },
              completedAt: new Date(),
              status: bulkResult.success ? 'DELIVERED' : 'FAILED'
            };
          }
        }
      } catch (error) {
        this.logger.error(`Erro no envio em massa para canal ${channel}: ${error}`);
        
        // Marcar como falha para todos os destinatários deste canal
        for (const recipient of channelRecipients) {
          if (results) {
            results[recipient.id] = {
              notificationId: `error-${uuidv4()}`,
              success: false,
              successfulChannels: [],
              failedChannels: [channel],
              skippedChannels: [],
              channelResults: {},
              completedAt: new Date(),
              status: 'FAILED'
            };
          }
        }
      }
    }
  }
  
  /**
   * Envia uma notificação através de um canal específico
   * @param channel Canal de notificação
   * @param recipient Destinatário
   * @param content Conteúdo
   * @param notificationId ID da notificação
   * @param options Opções de envio
   * @param event Evento que originou a notificação
   */
  private async sendToChannel(
    channel: NotificationChannel,
    recipient: NotificationRecipient,
    content: NotificationContent,
    notificationId: string,
    options: NotificationServiceOptions,
    event?: BaseEvent
  ): Promise<NotificationResult> {
    try {
      // Obter adaptador para o canal
      const adapter = await this.adapterFactory.getAdapter(channel);
      
      // Verificar se está pronto para uso
      if (!await adapter.isReady()) {
        return {
          success: false,
          notificationId,
          errorMessage: `Adaptador para canal ${channel} não está pronto`,
          errorCode: 'ADAPTER_NOT_READY',
          timestamp: new Date()
        };
      }
      
      // Obter opções específicas do canal
      const channelOptions = options.channelOptions?.[channel] || {};
      
      // Tentar enviar com retentativas
      const maxRetries = options.maxRetriesPerChannel || 1;
      let lastError: Error | null = null;
      
      for (let attempt = 0; attempt < maxRetries; attempt++) {
        try {
          // Enviar notificação
          const result = await adapter.send(recipient, content, event, {
            ...options,
            ...channelOptions,
            notificationId
          });
          
          // Se for bem-sucedido, retornar o resultado
          if (result.success) {
            return result;
          }
          
          // Registrar falha
          this.logger.warn(`Tentativa ${attempt + 1}/${maxRetries} falhou para canal ${channel}: ${result.errorMessage}`);
          
          // Se for um erro relacionado a endereço ausente, não tentar novamente
          if (result.errorCode === 'EMAIL_ADDRESS_MISSING' ||
              result.errorCode === 'PHONE_NUMBER_MISSING' ||
              result.errorCode === 'DEVICE_TOKEN_MISSING' ||
              result.errorCode === 'WEBHOOK_URL_MISSING') {
            return {
              ...result,
              errorCode: 'ADDRESS_MISSING'
            };
          }
          
          // Se for a última tentativa, retornar o resultado
          if (attempt === maxRetries - 1) {
            return result;
          }
          
          // Esperar antes da próxima tentativa com backoff exponencial
          const backoffFactor = options.retryBackoffFactor || 2;
          const retryInterval = (options.retryIntervalMs || 1000) * Math.pow(backoffFactor, attempt);
          await this.delay(retryInterval);
        } catch (error) {
          lastError = error;
          
          if (attempt === maxRetries - 1) {
            throw error;
          }
          
          // Esperar antes da próxima tentativa
          const backoffFactor = options.retryBackoffFactor || 2;
          const retryInterval = (options.retryIntervalMs || 1000) * Math.pow(backoffFactor, attempt);
          await this.delay(retryInterval);
        }
      }
      
      // Se chegou aqui, todas as tentativas falharam
      throw lastError || new Error(`Todas as ${maxRetries} tentativas falharam`);
    } catch (error) {
      return {
        success: false,
        notificationId,
        errorMessage: `Erro ao enviar notificação via ${channel}: ${error.message}`,
        errorCode: 'CHANNEL_ERROR',
        timestamp: new Date()
      };
    }
  }
  
  /**
   * Determina os canais para envio com base nas preferências e configurações
   * @param recipient Destinatário da notificação
   * @param options Opções de envio
   */
  private async determineTargetChannels(
    recipient: NotificationRecipient,
    options: NotificationServiceOptions
  ): Promise<NotificationChannel[]> {
    // Obter canais disponíveis na fábrica
    const availableChannels = this.adapterFactory.getAvailableChannels();
    
    if (availableChannels.length === 0) {
      return [];
    }
    
    // Lista de canais para tentar, em ordem de prioridade
    let targetChannels: NotificationChannel[] = [];
    
    // Primeiro, considerar canais preferenciais das opções
    if (options.preferredChannels && options.preferredChannels.length > 0) {
      // Filtrar apenas canais preferidos disponíveis
      targetChannels = options.preferredChannels.filter(
        channel => availableChannels.includes(channel)
      );
    }
    
    // Se não houver canais preferenciais definidos, usar as preferências do usuário
    if (targetChannels.length === 0 && recipient.preferredChannels?.length) {
      targetChannels = recipient.preferredChannels.filter(
        channel => availableChannels.includes(channel)
      );
    }
    
    // Se ainda não tivermos canais, usar todos disponíveis
    if (targetChannels.length === 0) {
      targetChannels = [...availableChannels];
    }
    
    // Verificar se o destinatário tem endereços para cada canal
    return targetChannels.filter(channel => {
      // Se o destinatário não tiver endereços, verificar canal por canal
      if (!recipient.addresses) {
        return this.hasDefaultAddressFor(channel, recipient);
      }
      
      // Verificar se há endereços para este canal
      const addresses = recipient.addresses.get(channel);
      return addresses && addresses.length > 0;
    });
  }
  
  /**
   * Verifica se há um endereço padrão para um canal
   * @param channel Canal de notificação
   * @param recipient Destinatário
   */
  private hasDefaultAddressFor(
    channel: NotificationChannel,
    recipient: NotificationRecipient
  ): boolean {
    switch (channel) {
      case NotificationChannel.EMAIL:
        // Verificar se o email está no ID ou metadados
        return typeof recipient.id === 'string' && 
               (recipient.id.includes('@') || (recipient.metadata?.email !== undefined));
        
      case NotificationChannel.SMS:
        // Verificar se há telefone nos metadados
        return recipient.metadata?.phone !== undefined || 
               recipient.metadata?.phoneNumber !== undefined || 
               recipient.metadata?.mobile !== undefined;
        
      case NotificationChannel.PUSH:
        // Push requer tokens de dispositivo
        return false;
        
      case NotificationChannel.WEBHOOK:
        // Webhook poderia usar uma URL padrão
        return true;
        
      default:
        return false;
    }
  }
  
  /**
   * Cancela uma notificação agendada
   * @param notificationId ID da notificação a ser cancelada
   */
  async cancelNotification(notificationId: string): Promise<boolean> {
    // TODO: Implementar cancelamento de notificações agendadas
    return false;
  }
  
  /**
   * Verifica o status de uma notificação
   * @param notificationId ID da notificação
   */
  async getNotificationStatus(notificationId: string): Promise<{
    status: 'SCHEDULED' | 'SENT' | 'DELIVERED' | 'FAILED' | 'OPENED' | 'CLICKED' | 'EXPIRED' | 'CANCELLED';
    details: Record<string, any>;
  }> {
    // TODO: Implementar verificação de status de notificação
    return {
      status: 'SENT',
      details: {}
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