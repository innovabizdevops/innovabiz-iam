/**
 * Fábrica de provedores Mobile Money
 * 
 * Esta classe é responsável por instanciar e gerenciar os adaptadores
 * para diferentes provedores de Mobile Money (MPesa, Airtel, MTN, etc.)
 */

import { Logger } from '../../observability/logging/hook_logger';
import { MobileMoneyProvider } from './types';
import { MobileMoneyProviderAdapter } from './provider-adapter';
import { ConfigService } from '../../infrastructure/common/config_service';

/**
 * Interface para adaptadores de provedores específicos
 */
export interface MobileMoneyProviderInstance {
  initiateTransaction(params: any): Promise<any>;
  verifyOTP(params: any): Promise<any>;
  checkStatus(params: any): Promise<any>;
  cancelTransaction(params: any): Promise<any>;
  getDetails(params: any): Promise<any>;
}

/**
 * Fábrica para provedores de Mobile Money
 */
export class MobileMoneyProviderFactory {
  private readonly logger: Logger;
  private readonly configService: ConfigService;
  private providerAdapters: Map<string, MobileMoneyProviderInstance> = new Map();
  
  constructor(logger: Logger, configService: ConfigService) {
    this.logger = logger;
    this.configService = configService;
    this.initializeProviderAdapters();
  }
  
  /**
   * Inicializa adaptadores para todos os provedores configurados
   */
  private async initializeProviderAdapters(): Promise<void> {
    try {
      this.logger.info('Initializing Mobile Money provider adapters');
      
      // Obter configurações de provedores
      const configs = await this.configService.getProviderConfigurations('mobile-money');
      
      // Criar adaptador para cada provedor configurado
      for (const config of configs) {
        const adapter = this.createProviderAdapter(config.providerId as MobileMoneyProvider, config);
        this.providerAdapters.set(config.providerId, adapter);
        this.logger.debug(`Initialized adapter for provider: ${config.providerId}`);
      }
      
      this.logger.info(`Successfully initialized ${this.providerAdapters.size} provider adapters`);
    } catch (error) {
      this.logger.error('Failed to initialize provider adapters', { error });
    }
  }
  
  /**
   * Cria adaptador para um provedor específico
   */
  private createProviderAdapter(provider: MobileMoneyProvider, config: any): MobileMoneyProviderInstance {
    switch (provider) {
      case MobileMoneyProvider.MPESA:
        return new MobileMoneyProviderAdapter(provider, config, this.logger);
      case MobileMoneyProvider.AIRTEL:
        return new MobileMoneyProviderAdapter(provider, config, this.logger);
      case MobileMoneyProvider.MTN:
        return new MobileMoneyProviderAdapter(provider, config, this.logger);
      case MobileMoneyProvider.UNITEL:
        return new MobileMoneyProviderAdapter(provider, config, this.logger);
      case MobileMoneyProvider.MOVICEL:
        return new MobileMoneyProviderAdapter(provider, config, this.logger);
      default:
        // Adaptador genérico para outros provedores
        return new MobileMoneyProviderAdapter(provider, config, this.logger);
    }
  }
  
  /**
   * Obtém um adaptador para o provedor especificado
   */
  public getProvider(provider: MobileMoneyProvider): MobileMoneyProviderInstance {
    const adapter = this.providerAdapters.get(provider);
    
    if (!adapter) {
      this.logger.error(`Provider adapter not found for: ${provider}`);
      throw new Error(`Provider not supported: ${provider}`);
    }
    
    return adapter;
  }
  
  /**
   * Atualiza as configurações de um provedor específico
   */
  public updateProviderConfig(provider: MobileMoneyProvider, config: any): void {
    const adapter = this.providerAdapters.get(provider);
    
    if (adapter) {
      // Criar novo adaptador com configuração atualizada
      const updatedAdapter = this.createProviderAdapter(provider, config);
      this.providerAdapters.set(provider, updatedAdapter);
      this.logger.info(`Updated configuration for provider: ${provider}`);
    } else {
      // Criar novo adaptador
      const newAdapter = this.createProviderAdapter(provider, config);
      this.providerAdapters.set(provider, newAdapter);
      this.logger.info(`Added new provider adapter: ${provider}`);
    }
  }
  
  /**
   * Remove um adaptador de provedor
   */
  public removeProvider(provider: MobileMoneyProvider): boolean {
    const result = this.providerAdapters.delete(provider);
    
    if (result) {
      this.logger.info(`Removed provider adapter: ${provider}`);
    } else {
      this.logger.warn(`Attempted to remove non-existent provider adapter: ${provider}`);
    }
    
    return result;
  }
  
  /**
   * Obtém lista de provedores suportados atualmente
   */
  public getSupportedProviders(): string[] {
    return Array.from(this.providerAdapters.keys());
  }
}