/**
 * @file notification-adapter-factory.ts
 * @description Fábrica para criação e gerenciamento de adaptadores de notificação
 * 
 * Esta fábrica é responsável por criar, inicializar e gerenciar os diferentes
 * adaptadores de notificação suportados pelo sistema.
 */

import { Logger } from '../../../infrastructure/observability/logger';
import { NotificationAdapter } from './notification-adapter';
import { NotificationChannel, ChannelConfig } from '../core/notification-channel';
import { EmailAdapter, EmailAdapterConfig } from './email-adapter';
import { SmsAdapter, SmsAdapterConfig } from './sms-adapter';
import { PushAdapter, PushAdapterConfig } from './push-adapter';
import { WebhookAdapter, WebhookAdapterConfig } from './webhook-adapter';

/**
 * Interface para configuração geral da fábrica de adaptadores
 */
export interface NotificationAdapterFactoryConfig {
  /**
   * Canais de notificação habilitados
   */
  enabledChannels: NotificationChannel[];
  
  /**
   * Configurações específicas por canal
   */
  channelConfigs: Map<NotificationChannel, ChannelConfig>;
  
  /**
   * Timeout para inicialização de adaptadores (ms)
   */
  initializationTimeoutMs?: number;
  
  /**
   * Número de tentativas para inicialização de adaptadores
   */
  initializationMaxRetries?: number;
  
  /**
   * Indicador se deve tentar recuperar adaptadores com falha
   */
  autoRecoveryEnabled?: boolean;
  
  /**
   * Intervalo para verificação automática de adaptadores (ms)
   */
  healthCheckIntervalMs?: number;
}

/**
 * Fábrica para criação e gerenciamento de adaptadores de notificação
 */
export class NotificationAdapterFactory {
  private adapters: Map<NotificationChannel, NotificationAdapter> = new Map();
  private config: NotificationAdapterFactoryConfig;
  private healthCheckInterval?: NodeJS.Timeout;
  private logger = new Logger('NotificationAdapterFactory');
  private initializing = false;
  
  /**
   * Construtor
   * @param config Configuração da fábrica
   */
  constructor(config: NotificationAdapterFactoryConfig) {
    this.config = config;
  }
  
  /**
   * Inicializa todos os adaptadores configurados
   */
  async initialize(): Promise<void> {
    if (this.initializing) {
      throw new Error('A fábrica de adaptadores já está sendo inicializada');
    }
    
    this.initializing = true;
    
    try {
      this.logger.info('Iniciando fábrica de adaptadores de notificação', {
        enabledChannels: this.config.enabledChannels
      });
      
      // Inicializar adaptadores para cada canal habilitado
      const initPromises = this.config.enabledChannels.map(async channel => {
        try {
          await this.initializeChannel(channel);
        } catch (error) {
          this.logger.error(`Falha ao inicializar adaptador para canal ${channel}: ${error}`, {
            channel,
            error
          });
        }
      });
      
      // Aguardar inicialização de todos os adaptadores
      await Promise.all(initPromises);
      
      // Verificar quantos adaptadores foram inicializados com sucesso
      const initializedCount = [...this.adapters.values()].filter(adapter => 
        adapter.isReady && adapter.isReady()).length;
      
      if (initializedCount === 0) {
        throw new Error('Nenhum adaptador de notificação pôde ser inicializado');
      }
      
      this.logger.info(`Fábrica de adaptadores inicializada com sucesso: ${initializedCount}/${this.config.enabledChannels.length} adaptadores disponíveis`);
      
      // Iniciar monitoramento de saúde dos adaptadores se configurado
      if (this.config.autoRecoveryEnabled && this.config.healthCheckIntervalMs) {
        this.startHealthCheck();
      }
    } finally {
      this.initializing = false;
    }
  }
  
  /**
   * Inicializa um adaptador para um canal específico
   * @param channel Canal de notificação
   */
  private async initializeChannel(channel: NotificationChannel): Promise<void> {
    const channelConfig = this.config.channelConfigs.get(channel);
    
    if (!channelConfig) {
      throw new Error(`Configuração não encontrada para canal ${channel}`);
    }
    
    // Criar instância do adaptador apropriado
    const adapter = this.createAdapter(channel);
    
    if (!adapter) {
      throw new Error(`Adaptador não suportado para canal ${channel}`);
    }
    
    // Tentar inicializar o adaptador com retentativas
    const maxRetries = this.config.initializationMaxRetries || 3;
    const timeout = this.config.initializationTimeoutMs || 30000;
    
    for (let attempt = 0; attempt < maxRetries; attempt++) {
      try {
        // Aplicar timeout para evitar travamento na inicialização
        await Promise.race([
          adapter.initialize(channelConfig),
          new Promise((_, reject) => {
            setTimeout(() => reject(new Error(`Timeout ao inicializar adaptador para canal ${channel}`)), timeout);
          })
        ]);
        
        // Se a inicialização foi bem-sucedida, armazenar o adaptador
        this.adapters.set(channel, adapter);
        this.logger.info(`Adaptador inicializado com sucesso para canal ${channel}`);
        return;
      } catch (error) {
        if (attempt === maxRetries - 1) {
          // Última tentativa falhou
          throw error;
        }
        
        // Tentar novamente após um breve intervalo
        this.logger.warn(`Tentativa ${attempt + 1}/${maxRetries} falhou para canal ${channel}: ${error}. Tentando novamente...`);
        await this.delay(1000 * Math.pow(2, attempt)); // Backoff exponencial
      }
    }
  }
  
  /**
   * Cria uma instância de adaptador para o canal especificado
   * @param channel Canal de notificação
   */
  private createAdapter(channel: NotificationChannel): NotificationAdapter | null {
    switch (channel) {
      case NotificationChannel.EMAIL:
        return new EmailAdapter();
        
      case NotificationChannel.SMS:
        return new SmsAdapter();
        
      case NotificationChannel.PUSH:
        return new PushAdapter();
        
      case NotificationChannel.WEBHOOK:
        return new WebhookAdapter();
        
      default:
        this.logger.error(`Tipo de canal não suportado: ${channel}`);
        return null;
    }
  }
  
  /**
   * Obtém um adaptador para o canal especificado
   * @param channel Canal de notificação
   */
  async getAdapter<T extends NotificationAdapter>(channel: NotificationChannel): Promise<T> {
    const adapter = this.adapters.get(channel);
    
    if (!adapter) {
      throw new Error(`Adaptador não disponível para canal ${channel}`);
    }
    
    // Verificar se o adaptador está pronto
    const isReady = await adapter.isReady();
    
    if (!isReady) {
      if (this.config.autoRecoveryEnabled) {
        this.logger.warn(`Adaptador para canal ${channel} não está pronto. Tentando reinicializar...`);
        try {
          await this.initializeChannel(channel);
          return this.adapters.get(channel) as T;
        } catch (error) {
          throw new Error(`Falha ao reinicializar adaptador para canal ${channel}: ${error}`);
        }
      } else {
        throw new Error(`Adaptador para canal ${channel} não está pronto`);
      }
    }
    
    return adapter as T;
  }
  
  /**
   * Retorna todos os adaptadores disponíveis
   */
  async getAllAdapters(): Promise<Map<NotificationChannel, NotificationAdapter>> {
    return new Map(this.adapters);
  }
  
  /**
   * Verifica se um canal está habilitado e disponível
   * @param channel Canal de notificação
   */
  isChannelAvailable(channel: NotificationChannel): boolean {
    return this.adapters.has(channel);
  }
  
  /**
   * Retorna os canais disponíveis
   */
  getAvailableChannels(): NotificationChannel[] {
    return Array.from(this.adapters.keys());
  }
  
  /**
   * Inicia verificação periódica da saúde dos adaptadores
   */
  private startHealthCheck(): void {
    if (this.healthCheckInterval) {
      clearInterval(this.healthCheckInterval);
    }
    
    this.healthCheckInterval = setInterval(async () => {
      this.logger.debug('Verificando saúde dos adaptadores de notificação');
      
      for (const [channel, adapter] of this.adapters.entries()) {
        try {
          const isReady = await adapter.isReady();
          
          if (!isReady && this.config.autoRecoveryEnabled) {
            this.logger.warn(`Adaptador para canal ${channel} não está saudável. Tentando recuperar...`);
            try {
              await this.initializeChannel(channel);
              this.logger.info(`Adaptador para canal ${channel} recuperado com sucesso`);
            } catch (error) {
              this.logger.error(`Falha ao recuperar adaptador para canal ${channel}: ${error}`);
            }
          }
        } catch (error) {
          this.logger.error(`Erro ao verificar saúde do adaptador para canal ${channel}: ${error}`);
        }
      }
    }, this.config.healthCheckIntervalMs);
  }
  
  /**
   * Finaliza a fábrica e libera recursos
   */
  dispose(): void {
    if (this.healthCheckInterval) {
      clearInterval(this.healthCheckInterval);
    }
    
    this.adapters.clear();
    this.logger.info('Fábrica de adaptadores finalizada');
  }
  
  /**
   * Cria um delay (promise) por um tempo específico
   * @param ms Tempo em milissegundos
   */
  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}