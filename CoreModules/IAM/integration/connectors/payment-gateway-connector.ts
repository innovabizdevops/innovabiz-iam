/**
 * @file payment-gateway-connector.ts
 * @description Conector de integração entre IAM e Payment Gateway para autenticação de transações
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';
import { AbstractConnector, BaseConnectorConfig, ConnectorStatus, TelemetryContext } from './base-connector';
import { Logger } from '../../observability/logging/hook_logger';
import { MetricsCollector } from '../../observability/metrics/hook_metrics';
import { TracingProvider } from '../../observability/tracing/hook_tracing';

/**
 * Configuração específica para o conector Payment Gateway
 */
export interface PaymentGatewayConnectorConfig extends BaseConnectorConfig {
  /** ID do comerciante/merchant */
  merchantId: string;
  
  /** Versão da API do Payment Gateway */
  apiVersion: string;
  
  /** Configurações de autenticação de transação */
  transactionAuth?: {
    /** Valor mínimo para autenticação forte */
    strongAuthThreshold: number;
    /** Configurações 3DS */
    threeDSecure: {
      enabled: boolean;
      version: '1.0' | '2.0' | '2.1' | '2.2';
      challengeIndicator?: 'no-preference' | 'no-challenge' | 'challenge-preferred' | 'challenge-required';
    };
    /** Modo de operação */
    mode: 'sync' | 'async' | 'hybrid';
  };
  
  /** Configurações de tokenização */
  tokenization?: {
    /** Se tokenização está ativa */
    enabled: boolean;
    /** Tipo de token */
    tokenType: 'pci' | 'non-pci' | 'network';
    /** Tempo de expiração do token em segundos */
    expirationSeconds?: number;
  };
  
  /** Configurações específicas por região */
  regionalSettings?: Record<string, {
    /** URL base regional */
    baseUrl: string;
    /** Configurações específicas da região */
    settings: Record<string, any>;
  }>;
}

/**
 * Resposta de verificação de autenticação
 */
export interface AuthVerificationResponse {
  /** Status da verificação */
  status: 'approved' | 'rejected' | 'pending' | 'requires_action';
  
  /** ID da transação */
  transactionId: string;
  
  /** ID da referência de autenticação */
  authenticationId?: string;
  
  /** Fatores usados na autenticação */
  factorsUsed?: string[];
  
  /** URL para redirecionamento (para 3DS ou outro fluxo) */
  redirectUrl?: string;
  
  /** Código do desafio se aplicável */
  challengeCode?: string;
  
  /** Detalhes adicionais de verificação */
  verificationDetails?: Record<string, any>;
  
  /** Timestamp da verificação */
  timestamp: string;
}

/**
 * Detalhes de um método de pagamento
 */
export interface PaymentMethodDetails {
  /** ID do método de pagamento */
  id: string;
  
  /** Tipo do método de pagamento */
  type: 'credit_card' | 'debit_card' | 'bank_account' | 'wallet' | 'crypto' | 'other';
  
  /** Subtipo do método de pagamento */
  subtype?: string;
  
  /** Dados do método de pagamento */
  data: {
    /** Últimos 4 dígitos (cartão/conta) */
    last4?: string;
    /** Bandeira (para cartões) */
    brand?: string;
    /** Mês de expiração */
    expiryMonth?: string;
    /** Ano de expiração */
    expiryYear?: string;
    /** Nome do titular */
    holderName?: string;
    /** Se é tokenizado */
    tokenized: boolean;
    /** Token (se tokenizado) */
    token?: string;
    /** Metadados adicionais */
    metadata?: Record<string, any>;
  };
  
  /** Referência do usuário dono do método de pagamento */
  userId: string;
  
  /** Nível de confiança do método de pagamento */
  trustLevel?: 'high' | 'medium' | 'low';
  
  /** Data de criação */
  createdAt: string;
  
  /** Última atualização */
  updatedAt: string;
}

/**
 * Status de uma transação
 */
export type TransactionStatus = 
  'initiated' | 'pending' | 'authorized' | 'captured' | 
  'settled' | 'failed' | 'cancelled' | 'refunded' | 'chargeback';

/**
 * Conector para integração com Payment Gateway
 */
export class PaymentGatewayConnector extends AbstractConnector {
  /** Cliente HTTP para comunicação com Payment Gateway */
  private client: AxiosInstance;
  
  /** Configuração específica do Payment Gateway */
  private paymentGatewayConfig: PaymentGatewayConnectorConfig;
  
  /**
   * Construtor do conector Payment Gateway
   * @param config Configuração do conector
   * @param logger Logger para o conector
   * @param metrics Coletor de métricas opcional
   * @param tracer Provedor de rastreamento opcional
   */
  constructor(
    config: PaymentGatewayConnectorConfig,
    logger: Logger,
    metrics?: MetricsCollector,
    tracer?: TracingProvider
  ) {
    super(config, logger, metrics, tracer);
    this.paymentGatewayConfig = config;
  }
  
  /**
   * Valida a configuração específica do Payment Gateway
   * @throws Error se a configuração for inválida
   */
  protected validateConfig(): void {
    super.validateConfig();
    
    if (!this.paymentGatewayConfig.merchantId) {
      throw new Error('merchantId é obrigatório na configuração do Payment Gateway');
    }
    
    if (!this.paymentGatewayConfig.apiVersion) {
      throw new Error('apiVersion é obrigatório na configuração do Payment Gateway');
    }
  }
  
  /**
   * Inicializa o conector Payment Gateway
   * @returns Promise resolvida quando o conector estiver inicializado
   */
  protected async doInitialize(): Promise<boolean> {
    try {
      // Configuração do cliente HTTP
      const axiosConfig: AxiosRequestConfig = {
        baseURL: this.config.baseUrl,
        timeout: this.config.timeoutMs || 30000,
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
          'X-API-Version': this.paymentGatewayConfig.apiVersion,
          'X-Merchant-ID': this.paymentGatewayConfig.merchantId
        }
      };
      
      // Configuração de autenticação com base no tipo
      switch (this.config.auth.type) {
        case 'apiKey':
          axiosConfig.headers = {
            ...axiosConfig.headers,
            'X-API-Key': this.config.auth.credentials.apiKey,
          };
          break;
        case 'oauth2':
          // A autenticação será feita via interceptor
          break;
        case 'jwt':
          axiosConfig.headers = {
            ...axiosConfig.headers,
            'Authorization': `Bearer ${this.config.auth.credentials.token}`,
          };
          break;
        case 'basic':
          const { username, password } = this.config.auth.credentials;
          const auth = Buffer.from(`${username}:${password}`).toString('base64');
          axiosConfig.headers = {
            ...axiosConfig.headers,
            'Authorization': `Basic ${auth}`,
          };
          break;
        case 'mTLS':
          // Configuração de mTLS para cliente HTTP
          axiosConfig.httpsAgent = {
            cert: this.config.auth.credentials.cert,
            key: this.config.auth.credentials.key,
            ca: this.config.auth.credentials.ca,
          };
          break;
        default:
          throw new Error(`Tipo de autenticação não suportado: ${this.config.auth.type}`);
      }
      
      // Criação do cliente HTTP
      this.client = axios.create(axiosConfig);
      
      // Configuração de interceptors para observabilidade
      this.configureInterceptors();
      
      // Verificar conectividade com health check
      const healthCheckResult = await this.healthCheck();
      return healthCheckResult.status === ConnectorStatus.CONNECTED;
    } catch (error) {
      this.logger.error(`Erro ao inicializar conector Payment Gateway: ${error instanceof Error ? error.message : String(error)}`);
      return false;
    }
  }
  
  /**
   * Configura interceptors para o cliente HTTP
   */
  private configureInterceptors(): void {
    // Interceptor de requisição para observabilidade
    this.client.interceptors.request.use(
      (config) => {
        const requestStartTime = Date.now();
        config.headers = config.headers || {};
        config.headers['X-Request-ID'] = this.generateId('req');
        
        // Adicionar informações de telemetria se disponível
        if (config.headers['X-Telemetry-Context']) {
          const telemetryContext = JSON.parse(config.headers['X-Telemetry-Context'] as string);
          
          if (telemetryContext.transactionId) {
            config.headers['X-Transaction-ID'] = telemetryContext.transactionId;
          }
          
          if (telemetryContext.correlationId) {
            config.headers['X-Correlation-ID'] = telemetryContext.correlationId;
          }
          
          delete config.headers['X-Telemetry-Context'];
        }
        
        // Armazenar tempo de início para cálculo de latência
        config.metadata = {
          ...config.metadata,
          requestStartTime
        };
        
        return config;
      },
      (error) => {
        this.logger.error(`Erro no interceptor de requisição: ${error.message}`);
        return Promise.reject(error);
      }
    );
    
    // Interceptor de resposta para observabilidade
    this.client.interceptors.response.use(
      (response) => {
        const requestStartTime = response.config.metadata?.requestStartTime;
        const latencyMs = requestStartTime ? Date.now() - requestStartTime : undefined;
        
        if (this.config.observability?.metricsEnabled && this.metrics && latencyMs) {
          this.metrics.recordValue('paymentGateway.request.latency', latencyMs, {
            path: response.config.url || '',
            method: response.config.method || 'unknown',
            status: response.status.toString()
          });
        }
        
        return response;
      },
      (error) => {
        if (error.response) {
          const requestStartTime = error.config?.metadata?.requestStartTime;
          const latencyMs = requestStartTime ? Date.now() - requestStartTime : undefined;
          
          if (this.config.observability?.metricsEnabled && this.metrics && latencyMs) {
            this.metrics.recordValue('paymentGateway.request.latency', latencyMs, {
              path: error.config?.url || '',
              method: error.config?.method || 'unknown',
              status: error.response.status.toString(),
              error: 'true'
            });
          }
          
          this.logger.warn(`Requisição Payment Gateway falhou: ${error.response.status} ${JSON.stringify(error.response.data)}`);
        } else {
          this.logger.error(`Erro na requisição Payment Gateway: ${error.message}`);
        }
        
        return Promise.reject(error);
      }
    );
  }
  
  /**
   * Realiza health check no serviço Payment Gateway
   * @returns Resultado do health check
   */
  public async healthCheck(): Promise<{
    status: ConnectorStatus;
    details?: Record<string, any>;
    latencyMs?: number;
  }> {
    try {
      const startTime = Date.now();
      const response = await this.client.get('/health');
      const latencyMs = Date.now() - startTime;
      
      const status = response.data.status === 'UP' 
        ? ConnectorStatus.CONNECTED 
        : ConnectorStatus.DEGRADED;
      
      return {
        status,
        details: {
          apiVersion: response.data.version,
          services: response.data.services,
          environment: response.data.environment,
          features: response.data.features
        },
        latencyMs
      };
    } catch (error) {
      this.status = ConnectorStatus.ERROR;
      
      return {
        status: ConnectorStatus.ERROR,
        details: {
          error: error instanceof Error ? error.message : String(error)
        }
      };
    }
  }
  
  /**
   * Fecha conexão com o serviço Payment Gateway
   */
  public async shutdown(): Promise<void> {
    this.logger.info('Encerrando conexão com Payment Gateway');
    this.status = ConnectorStatus.DISCONNECTED;
  }
  
  /**
   * Verifica a autenticação para uma transação
   * @param transactionData Dados da transação
   * @param authenticationOptions Opções de autenticação
   * @param telemetryContext Contexto de telemetria opcional
   * @returns Resposta da verificação de autenticação
   */
  public async verifyTransactionAuthentication(
    transactionData: {
      transactionId: string;
      userId: string;
      amount: number;
      currency: string;
      paymentMethodId: string;
      paymentMethodType?: string;
      merchantData?: {
        name?: string;
        mcc?: string;
        city?: string;
        country?: string;
      };
      ipAddress?: string;
      deviceId?: string;
      browserData?: {
        userAgent?: string;
        acceptHeader?: string;
        language?: string;
        colorDepth?: number;
        screenHeight?: number;
        screenWidth?: number;
        timeZoneOffset?: number;
        javaEnabled?: boolean;
      };
    },
    authenticationOptions: {
      challengePreference?: 'no-preference' | 'no-challenge' | 'challenge-preferred' | 'challenge-required';
      redirectUrl?: string;
      requestedExemption?: 'low_value' | 'recurring' | 'trusted_merchant' | 'trusted_listing' | 'secure_corporate' | 'none';
      authenticationMethod?: 'password' | 'biometric' | 'otp' | 'federated' | 'token';
      authFactors?: string[];
      skipVerification?: boolean;
    },
    telemetryContext?: Partial<TelemetryContext>
  ): Promise<AuthVerificationResponse> {
    const context = this.createTelemetryContext(telemetryContext);
    
    try {
      const span = this.tracer?.createSpan('paymentGateway.verifyTransactionAuthentication', context);
      
      try {
        // Implementar lógica de verificação de valores para autenticação forte
        const requiresStrongAuth = this.paymentGatewayConfig.transactionAuth?.strongAuthThreshold &&
          transactionData.amount >= this.paymentGatewayConfig.transactionAuth.strongAuthThreshold;
        
        // Preparar dados para envio
        const requestData = {
          ...transactionData,
          merchantId: this.paymentGatewayConfig.merchantId,
          authentication: {
            ...authenticationOptions,
            threeDSecureOptions: this.paymentGatewayConfig.transactionAuth?.threeDSecure,
            requiresStrongAuth,
            requestTime: new Date().toISOString()
          }
        };
        
        // Registrar métricas de tentativa
        if (this.config.observability?.metricsEnabled && this.metrics) {
          this.metrics.incrementCounter('paymentGateway.verifyAuth.attempts', {
            paymentType: transactionData.paymentMethodType || 'unknown',
            requiresStrongAuth: String(requiresStrongAuth)
          });
        }
        
        // Realizar requisição
        const response = await this.client.post<AuthVerificationResponse>(
          '/v1/transactions/authentication/verify',
          requestData,
          {
            headers: {
              'X-Telemetry-Context': JSON.stringify(context)
            }
          }
        );
        
        // Registrar métricas de resultado
        if (this.config.observability?.metricsEnabled && this.metrics) {
          this.metrics.incrementCounter('paymentGateway.verifyAuth.result', {
            status: response.data.status,
            paymentType: transactionData.paymentMethodType || 'unknown'
          });
        }
        
        span?.end('success', {
          status: response.data.status,
          transactionId: response.data.transactionId
        });
        
        return response.data;
      } catch (error) {
        span?.end('error', { 
          error: error instanceof Error ? error.message : String(error),
          transactionId: transactionData.transactionId
        });
        throw error;
      }
    } catch (error) {
      this.logger.error(`Erro ao verificar autenticação de transação: ${error instanceof Error ? error.message : String(error)}`);
      
      // Registrar métricas de erro
      if (this.config.observability?.metricsEnabled && this.metrics) {
        this.metrics.incrementCounter('paymentGateway.verifyAuth.error', {
          paymentType: transactionData.paymentMethodType || 'unknown',
          error: error instanceof Error ? error.name : 'UnknownError'
        });
      }
      
      throw error;
    }
  }
  
  /**
   * Obter detalhes de um método de pagamento do usuário
   * @param userId ID do usuário
   * @param paymentMethodId ID do método de pagamento
   * @param telemetryContext Contexto de telemetria opcional
   * @returns Detalhes do método de pagamento
   */
  public async getPaymentMethodDetails(
    userId: string,
    paymentMethodId: string,
    telemetryContext?: Partial<TelemetryContext>
  ): Promise<PaymentMethodDetails> {
    const context = this.createTelemetryContext(telemetryContext);
    
    try {
      const response = await this.client.get<PaymentMethodDetails>(
        `/v1/users/${userId}/payment-methods/${paymentMethodId}`,
        {
          headers: {
            'X-Telemetry-Context': JSON.stringify(context)
          }
        }
      );
      
      return response.data;
    } catch (error) {
      this.logger.error(`Erro ao obter detalhes do método de pagamento: ${error instanceof Error ? error.message : String(error)}`);
      throw error;
    }
  }
  
  /**
   * Listar métodos de pagamento de um usuário
   * @param userId ID do usuário
   * @param options Opções de listagem
   * @param telemetryContext Contexto de telemetria opcional
   * @returns Lista de métodos de pagamento
   */
  public async listUserPaymentMethods(
    userId: string,
    options: {
      types?: Array<'credit_card' | 'debit_card' | 'bank_account' | 'wallet' | 'crypto' | 'other'>;
      limit?: number;
      offset?: number;
      includeExpired?: boolean;
      includeTrustLevel?: boolean;
    } = {},
    telemetryContext?: Partial<TelemetryContext>
  ): Promise<{
    paymentMethods: PaymentMethodDetails[];
    total: number;
    offset: number;
    limit: number;
  }> {
    const context = this.createTelemetryContext(telemetryContext);
    
    try {
      const response = await this.client.get(
        `/v1/users/${userId}/payment-methods`,
        {
          params: {
            types: options.types ? options.types.join(',') : undefined,
            limit: options.limit || 10,
            offset: options.offset || 0,
            include_expired: options.includeExpired || false,
            include_trust_level: options.includeTrustLevel || true
          },
          headers: {
            'X-Telemetry-Context': JSON.stringify(context)
          }
        }
      );
      
      return response.data;
    } catch (error) {
      this.logger.error(`Erro ao listar métodos de pagamento do usuário: ${error instanceof Error ? error.message : String(error)}`);
      throw error;
    }
  }
  
  /**
   * Verificar o status de uma transação
   * @param transactionId ID da transação
   * @param telemetryContext Contexto de telemetria opcional
   * @returns Status da transação
   */
  public async getTransactionStatus(
    transactionId: string,
    telemetryContext?: Partial<TelemetryContext>
  ): Promise<{
    transactionId: string;
    status: TransactionStatus;
    amount: number;
    currency: string;
    createdAt: string;
    updatedAt: string;
    authenticationStatus?: 'verified' | 'failed' | 'pending' | 'not_attempted';
    authenticationDetails?: Record<string, any>;
  }> {
    const context = this.createTelemetryContext(telemetryContext);
    
    try {
      const response = await this.client.get(
        `/v1/transactions/${transactionId}`,
        {
          headers: {
            'X-Telemetry-Context': JSON.stringify(context)
          }
        }
      );
      
      return response.data;
    } catch (error) {
      this.logger.error(`Erro ao obter status da transação: ${error instanceof Error ? error.message : String(error)}`);
      throw error;
    }
  }
  
  /**
   * Registrar resultado de autenticação de transação
   * @param transactionId ID da transação
   * @param authResult Resultado da autenticação
   * @param telemetryContext Contexto de telemetria opcional
   * @returns Confirmação do registro
   */
  public async reportAuthenticationResult(
    transactionId: string,
    authResult: {
      status: 'success' | 'failure' | 'abandoned';
      authenticationMethod?: string;
      factorsUsed?: string[];
      errorCode?: string;
      errorMessage?: string;
      challengeCompleted?: boolean;
      authenticationData?: Record<string, any>;
    },
    telemetryContext?: Partial<TelemetryContext>
  ): Promise<{
    success: boolean;
    transactionId: string;
    updatedStatus: TransactionStatus;
    timestamp: string;
  }> {
    const context = this.createTelemetryContext(telemetryContext);
    
    try {
      const span = this.tracer?.createSpan('paymentGateway.reportAuthResult', context);
      
      try {
        const response = await this.client.post(
          `/v1/transactions/${transactionId}/authentication/result`,
          {
            ...authResult,
            merchantId: this.paymentGatewayConfig.merchantId,
            timestamp: new Date().toISOString()
          },
          {
            headers: {
              'X-Telemetry-Context': JSON.stringify(context)
            }
          }
        );
        
        // Registrar métricas do resultado
        if (this.config.observability?.metricsEnabled && this.metrics) {
          this.metrics.incrementCounter('paymentGateway.authResult', {
            status: authResult.status,
            challengeCompleted: String(!!authResult.challengeCompleted)
          });
        }
        
        span?.end('success', {
          status: authResult.status,
          transactionId
        });
        
        return response.data;
      } catch (error) {
        span?.end('error', { 
          error: error instanceof Error ? error.message : String(error),
          transactionId
        });
        throw error;
      }
    } catch (error) {
      this.logger.error(`Erro ao reportar resultado de autenticação: ${error instanceof Error ? error.message : String(error)}`);
      
      if (this.config.observability?.metricsEnabled && this.metrics) {
        this.metrics.incrementCounter('paymentGateway.authResult.error', {
          transactionId
        });
      }
      
      throw error;
    }
  }
}