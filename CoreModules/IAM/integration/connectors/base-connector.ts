/**
 * @file base-connector.ts
 * @description Interface base para todos os conectores de integração do IAM com outros módulos
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

import { Logger } from '../../observability/logging/hook_logger';
import { MetricsCollector } from '../../observability/metrics/hook_metrics';
import { TracingProvider } from '../../observability/tracing/hook_tracing';

/**
 * Configuração base para conectores
 */
export interface BaseConnectorConfig {
  /** Endpoint base para o serviço */
  baseUrl: string;
  
  /** Configurações de autenticação */
  auth: {
    /** Tipo de autenticação a ser utilizada */
    type: 'apiKey' | 'oauth2' | 'jwt' | 'basic' | 'mTLS';
    /** Chaves de autenticação baseadas no tipo */
    credentials: Record<string, any>;
  };
  
  /** Timeout para requisições em milisegundos */
  timeoutMs?: number;
  
  /** Estratégia de retry */
  retry?: {
    /** Número máximo de tentativas */
    maxAttempts: number;
    /** Delay inicial entre tentativas em ms */
    initialDelayMs: number;
    /** Fator de crescimento exponencial do delay */
    backoffFactor: number;
    /** Delay máximo entre tentativas em ms */
    maxDelayMs: number;
  };
  
  /** Configurações específicas de circuitbreaker */
  circuitBreaker?: {
    /** Percentagem de falhas que aciona o circuitbreaker */
    failureThreshold: number;
    /** Tempo de reset do circuitbreaker em ms */
    resetTimeoutMs: number;
  };
  
  /** Configurações de cache */
  cache?: {
    /** Se o cache está habilitado */
    enabled: boolean;
    /** Tempo de vida do cache em ms */
    ttlMs: number;
  };
  
  /** Configurações de observabilidade */
  observability?: {
    /** Se o rastreamento está habilitado */
    tracingEnabled: boolean;
    /** Se as métricas estão habilitadas */
    metricsEnabled: boolean;
    /** Tags adicionais para telemetria */
    tags?: Record<string, string>;
  };
}

/**
 * Status de conexão do conector
 */
export enum ConnectorStatus {
  /** Conector não inicializado */
  NOT_INITIALIZED = 'NOT_INITIALIZED',
  /** Conector inicializado e conectado */
  CONNECTED = 'CONNECTED',
  /** Conector inicializado mas desconectado */
  DISCONNECTED = 'DISCONNECTED',
  /** Conector em estado de erro */
  ERROR = 'ERROR',
  /** Conector em estado de degradação */
  DEGRADED = 'DEGRADED',
}

/**
 * Contexto de telemetria para operações do conector
 */
export interface TelemetryContext {
  /** ID da transação */
  transactionId?: string;
  /** ID da correlação */
  correlationId?: string;
  /** ID do tenant */
  tenantId?: string;
  /** ID do usuário */
  userId?: string;
  /** Origem da requisição */
  source?: string;
  /** Tags adicionais */
  tags?: Record<string, string>;
}

/**
 * Interface base para conectores de integração
 */
export interface BaseConnector {
  /**
   * Inicializa o conector
   * @returns Promise resolvida quando o conector estiver inicializado
   */
  initialize(): Promise<boolean>;
  
  /**
   * Verifica o status atual do conector
   * @returns Status atual do conector
   */
  getStatus(): ConnectorStatus;
  
  /**
   * Realiza health check na conexão com o serviço
   * @returns Resultado do health check
   */
  healthCheck(): Promise<{
    status: ConnectorStatus;
    details?: Record<string, any>;
    latencyMs?: number;
  }>;
  
  /**
   * Fecha a conexão com o serviço
   * @returns Promise resolvida quando a conexão for fechada
   */
  shutdown(): Promise<void>;
}

/**
 * Classe base abstrata para implementação de conectores
 */
export abstract class AbstractConnector implements BaseConnector {
  /** Status atual do conector */
  protected status: ConnectorStatus = ConnectorStatus.NOT_INITIALIZED;
  
  /** Configuração do conector */
  protected config: BaseConnectorConfig;
  
  /** Logger para o conector */
  protected logger: Logger;
  
  /** Coletor de métricas */
  protected metrics?: MetricsCollector;
  
  /** Provedor de rastreamento */
  protected tracer?: TracingProvider;
  
  /**
   * Construtor da classe base de conectores
   * @param config Configuração do conector
   * @param logger Logger para o conector
   * @param metrics Coletor de métricas opcional
   * @param tracer Provedor de rastreamento opcional
   */
  constructor(
    config: BaseConnectorConfig, 
    logger: Logger,
    metrics?: MetricsCollector,
    tracer?: TracingProvider
  ) {
    this.config = config;
    this.logger = logger;
    this.metrics = metrics;
    this.tracer = tracer;
    
    this.validateConfig();
  }
  
  /**
   * Valida a configuração do conector
   * @throws Error se a configuração for inválida
   */
  protected validateConfig(): void {
    if (!this.config.baseUrl) {
      throw new Error('BaseUrl é obrigatório na configuração do conector');
    }
    
    if (!this.config.auth || !this.config.auth.type) {
      throw new Error('Tipo de autenticação é obrigatório na configuração do conector');
    }
  }
  
  /**
   * Inicializa o conector
   * @returns Promise resolvida quando o conector estiver inicializado
   */
  public async initialize(): Promise<boolean> {
    try {
      this.logger.info(`Inicializando conector com baseUrl: ${this.config.baseUrl}`);
      
      // Implementação específica será fornecida pelas subclasses
      const result = await this.doInitialize();
      
      if (result) {
        this.status = ConnectorStatus.CONNECTED;
        this.logger.info('Conector inicializado com sucesso');
      } else {
        this.status = ConnectorStatus.ERROR;
        this.logger.error('Falha ao inicializar conector');
      }
      
      return result;
    } catch (error) {
      this.status = ConnectorStatus.ERROR;
      this.logger.error(`Erro ao inicializar conector: ${error instanceof Error ? error.message : String(error)}`);
      return false;
    }
  }
  
  /**
   * Método abstrato para inicialização específica do conector
   * @returns Promise resolvida quando a inicialização específica estiver concluída
   */
  protected abstract doInitialize(): Promise<boolean>;
  
  /**
   * Retorna o status atual do conector
   * @returns Status do conector
   */
  public getStatus(): ConnectorStatus {
    return this.status;
  }
  
  /**
   * Realiza health check na conexão com o serviço
   * @returns Resultado do health check
   */
  public abstract healthCheck(): Promise<{
    status: ConnectorStatus;
    details?: Record<string, any>;
    latencyMs?: number;
  }>;
  
  /**
   * Fecha a conexão com o serviço
   * @returns Promise resolvida quando a conexão for fechada
   */
  public abstract shutdown(): Promise<void>;
  
  /**
   * Cria um contexto de telemetria para uma operação
   * @param context Contexto de telemetria opcional
   * @returns Contexto de telemetria completo
   */
  protected createTelemetryContext(context?: Partial<TelemetryContext>): TelemetryContext {
    return {
      transactionId: context?.transactionId || this.generateId('txn'),
      correlationId: context?.correlationId,
      tenantId: context?.tenantId,
      userId: context?.userId,
      source: context?.source || 'iam-connector',
      tags: {
        ...(this.config.observability?.tags || {}),
        ...(context?.tags || {}),
      },
    };
  }
  
  /**
   * Gera um ID único para uso em telemetria
   * @param prefix Prefixo para o ID
   * @returns ID único gerado
   */
  protected generateId(prefix: string): string {
    return `${prefix}-${Date.now()}-${Math.random().toString(36).substring(2, 10)}`;
  }
}