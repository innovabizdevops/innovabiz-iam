/**
 * @file mobile-money-connector.ts
 * @description Conector de integração entre IAM e Mobile Money para autenticação e validação de identidade
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';
import { AbstractConnector, BaseConnectorConfig, ConnectorStatus, TelemetryContext } from './base-connector';
import { Logger } from '../../observability/logging/hook_logger';
import { MetricsCollector } from '../../observability/metrics/hook_metrics';
import { TracingProvider } from '../../observability/tracing/hook_tracing';

/**
 * Configuração específica para o conector Mobile Money
 */
export interface MobileMoneyConnectorConfig extends BaseConnectorConfig {
  /** ID do provedor Mobile Money */
  providerId: string;
  
  /** Versão da API do Mobile Money */
  apiVersion: string;
  
  /** Código do país (ISO) */
  countryCode: string;
  
  /** Configurações específicas por operadora */
  operators?: Record<string, {
    /** Nome da operadora */
    name: string;
    /** Código da operadora */
    code: string;
    /** API específica da operadora */
    apiEndpoint?: string;
    /** Configurações específicas da operadora */
    settings?: Record<string, any>;
  }>;
  
  /** Configurações de verificação de identidade */
  identityVerification?: {
    /** Se a verificação de identidade está ativa */
    enabled: boolean;
    /** Nível de verificação requerido */
    requiredLevel: 'basic' | 'medium' | 'strict';
    /** Se deve verificar documentos */
    documentVerificationEnabled: boolean;
  };
  
  /** Configurações de validação de número */
  phoneValidation?: {
    /** Se a validação de número está ativa */
    enabled: boolean;
    /** Se deve verificar se o número existe */
    existenceCheck: boolean;
    /** Se deve verificar o status do número */
    statusCheck: boolean;
  };
  
  /** Configurações para Angola (específico) */
  angolaSettings?: {
    /** Integração com SIAC */
    siacIntegration?: {
      enabled: boolean;
      endpoint?: string;
      apiKey?: string;
    };
    /** Conformidade com BNA */
    bnaCompliance?: {
      enabled: boolean;
      level: 'basic' | 'enhanced';
    };
  };
}

/**
 * Interface para verificação de número de telefone
 */
export interface PhoneVerificationRequest {
  /** Número de telefone com código do país */
  phoneNumber: string;
  
  /** ID do usuário */
  userId: string;
  
  /** Operadora (opcional) */
  operator?: string;
  
  /** Tipo de verificação */
  verificationType: 'otp' | 'flash_call' | 'silent' | 'push';
  
  /** Configurações específicas do tipo de verificação */
  verificationOptions?: {
    /** Expiração em segundos */
    expirationSeconds?: number;
    /** Comprimento do OTP */
    otpLength?: number;
    /** Texto personalizado para SMS */
    customMessage?: string;
    /** Canal de entrega preferido */
    preferredChannel?: 'sms' | 'voice' | 'whatsapp';
    /** Idioma */
    language?: string;
  };
}

/**
 * Resposta de início de verificação de telefone
 */
export interface PhoneVerificationStartResponse {
  /** ID da verificação */
  verificationId: string;
  
  /** Status da verificação */
  status: 'initiated' | 'failed' | 'expired';
  
  /** Número de telefone parcialmente mascarado */
  maskedPhoneNumber: string;
  
  /** Tipo de verificação utilizado */
  verificationType: string;
  
  /** Tempo de expiração */
  expiresAt: string;
  
  /** Metadados adicionais */
  metadata?: Record<string, any>;
}

/**
 * Interface para validação de código OTP
 */
export interface OtpValidationRequest {
  /** ID da verificação */
  verificationId: string;
  
  /** Código OTP */
  code: string;
  
  /** ID do usuário */
  userId: string;
}

/**
 * Resposta de validação de código OTP
 */
export interface OtpValidationResponse {
  /** Status da validação */
  valid: boolean;
  
  /** ID da verificação */
  verificationId: string;
  
  /** Número de telefone parcialmente mascarado */
  maskedPhoneNumber: string;
  
  /** Se o número foi verificado */
  numberVerified: boolean;
  
  /** Data da verificação */
  verifiedAt?: string;
  
  /** Tentativas restantes (se falhou) */
  attemptsRemaining?: number;
  
  /** Mensagem de erro (se falhou) */
  errorMessage?: string;
}

/**
 * Detalhes de um usuário Mobile Money
 */
export interface MobileMoneyUserDetails {
  /** ID do usuário no Mobile Money */
  mobileMoneyId: string;
  
  /** Número de telefone */
  phoneNumber: string;
  
  /** Nome completo */
  fullName?: string;
  
  /** Status da conta */
  accountStatus: 'active' | 'suspended' | 'blocked' | 'pending_verification';
  
  /** Nível KYC */
  kycLevel: 'basic' | 'medium' | 'full';
  
  /** Operadora */
  operator: string;
  
  /** País */
  country: string;
  
  /** Saldo disponível */
  availableBalance?: {
    amount: number;
    currency: string;
  };
  
  /** Limite da conta */
  limits?: {
    daily: number;
    monthly: number;
    perTransaction: number;
    currency: string;
  };
  
  /** Verificações realizadas */
  verifications?: {
    phone: boolean;
    email?: boolean;
    identity?: boolean;
    address?: boolean;
    biometric?: boolean;
  };
}

/**
 * Conector para integração com Mobile Money
 */
export class MobileMoneyConnector extends AbstractConnector {
  /** Cliente HTTP para comunicação com Mobile Money */
  private client: AxiosInstance;
  
  /** Configuração específica do Mobile Money */
  private mobileMoneyConfig: MobileMoneyConnectorConfig;
  
  /**
   * Construtor do conector Mobile Money
   * @param config Configuração do conector
   * @param logger Logger para o conector
   * @param metrics Coletor de métricas opcional
   * @param tracer Provedor de rastreamento opcional
   */
  constructor(
    config: MobileMoneyConnectorConfig,
    logger: Logger,
    metrics?: MetricsCollector,
    tracer?: TracingProvider
  ) {
    super(config, logger, metrics, tracer);
    this.mobileMoneyConfig = config;
  }
  
  /**
   * Valida a configuração específica do Mobile Money
   * @throws Error se a configuração for inválida
   */
  protected validateConfig(): void {
    super.validateConfig();
    
    if (!this.mobileMoneyConfig.providerId) {
      throw new Error('providerId é obrigatório na configuração do Mobile Money');
    }
    
    if (!this.mobileMoneyConfig.apiVersion) {
      throw new Error('apiVersion é obrigatório na configuração do Mobile Money');
    }
    
    if (!this.mobileMoneyConfig.countryCode) {
      throw new Error('countryCode é obrigatório na configuração do Mobile Money');
    }
  }
  
  /**
   * Inicializa o conector Mobile Money
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
          'X-API-Version': this.mobileMoneyConfig.apiVersion,
          'X-Provider-ID': this.mobileMoneyConfig.providerId,
          'X-Country-Code': this.mobileMoneyConfig.countryCode
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
      this.logger.error(`Erro ao inicializar conector Mobile Money: ${error instanceof Error ? error.message : String(error)}`);
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
          this.metrics.recordValue('mobileMoney.request.latency', latencyMs, {
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
            this.metrics.recordValue('mobileMoney.request.latency', latencyMs, {
              path: error.config?.url || '',
              method: error.config?.method || 'unknown',
              status: error.response.status.toString(),
              error: 'true'
            });
          }
          
          this.logger.warn(`Requisição Mobile Money falhou: ${error.response.status} ${JSON.stringify(error.response.data)}`);
        } else {
          this.logger.error(`Erro na requisição Mobile Money: ${error.message}`);
        }
        
        return Promise.reject(error);
      }
    );
  }
  
  /**
   * Realiza health check no serviço Mobile Money
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
          operators: response.data.operators
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
   * Fecha conexão com o serviço Mobile Money
   */
  public async shutdown(): Promise<void> {
    this.logger.info('Encerrando conexão com Mobile Money');
    this.status = ConnectorStatus.DISCONNECTED;
  }
  
  /**
   * Inicia verificação de número de telefone
   * @param request Dados para verificação do telefone
   * @param telemetryContext Contexto de telemetria opcional
   * @returns Resposta do início da verificação
   */
  public async startPhoneVerification(
    request: PhoneVerificationRequest,
    telemetryContext?: Partial<TelemetryContext>
  ): Promise<PhoneVerificationStartResponse> {
    const context = this.createTelemetryContext(telemetryContext);
    
    try {
      const span = this.tracer?.createSpan('mobileMoney.startPhoneVerification', context);
      
      try {
        // Configurar dados adicionais
        const requestData = {
          ...request,
          providerId: this.mobileMoneyConfig.providerId,
          countryCode: this.mobileMoneyConfig.countryCode,
          timestamp: new Date().toISOString()
        };
        
        // Registrar métricas de tentativa
        if (this.config.observability?.metricsEnabled && this.metrics) {
          this.metrics.incrementCounter('mobileMoney.phoneVerification.attempts', {
            verificationType: request.verificationType,
            operator: request.operator || 'unknown'
          });
        }
        
        // Adicionar contexto de telemetria nos headers
        const headers = {
          'X-Telemetry-Context': JSON.stringify(context)
        };
        
        // Realizar requisição
        const response = await this.client.post<PhoneVerificationStartResponse>(
          '/v1/phone/verify',
          requestData,
          { headers }
        );
        
        // Registrar métricas de sucesso
        if (this.config.observability?.metricsEnabled && this.metrics) {
          this.metrics.incrementCounter('mobileMoney.phoneVerification.started', {
            status: response.data.status,
            verificationType: request.verificationType
          });
        }
        
        // Finalizar span com sucesso
        span?.end('success');
        
        return response.data;
      } catch (error) {
        // Finalizar span com erro
        span?.end('error', { error: error instanceof Error ? error.message : String(error) });
        throw error;
      }
    } catch (error) {
      this.logger.error(`Erro ao iniciar verificação de telefone: ${error instanceof Error ? error.message : String(error)}`);
      
      // Registrar métricas de falha
      if (this.config.observability?.metricsEnabled && this.metrics) {
        this.metrics.incrementCounter('mobileMoney.phoneVerification.error', {
          verificationType: request.verificationType,
          error: error instanceof Error ? error.name : 'UnknownError'
        });
      }
      
      throw error;
    }
  }
  
  /**
   * Valida um código OTP
   * @param request Dados para validação do OTP
   * @param telemetryContext Contexto de telemetria opcional
   * @returns Resposta da validação
   */
  public async validateOtp(
    request: OtpValidationRequest,
    telemetryContext?: Partial<TelemetryContext>
  ): Promise<OtpValidationResponse> {
    const context = this.createTelemetryContext(telemetryContext);
    
    try {
      const span = this.tracer?.createSpan('mobileMoney.validateOtp', context);
      
      try {
        // Configurar dados para envio
        const requestData = {
          ...request,
          providerId: this.mobileMoneyConfig.providerId,
          timestamp: new Date().toISOString()
        };
        
        // Registrar métricas de tentativa de validação
        if (this.config.observability?.metricsEnabled && this.metrics) {
          this.metrics.incrementCounter('mobileMoney.otp.validationAttempts', {
            verificationId: request.verificationId
          });
        }
        
        // Realizar requisição
        const response = await this.client.post<OtpValidationResponse>(
          '/v1/phone/verify/validate',
          requestData,
          {
            headers: {
              'X-Telemetry-Context': JSON.stringify(context)
            }
          }
        );
        
        // Registrar métricas do resultado
        if (this.config.observability?.metricsEnabled && this.metrics) {
          this.metrics.incrementCounter('mobileMoney.otp.validationResult', {
            valid: String(response.data.valid),
            verificationId: request.verificationId
          });
        }
        
        span?.end('success', {
          valid: response.data.valid
        });
        
        return response.data;
      } catch (error) {
        span?.end('error', { error: error instanceof Error ? error.message : String(error) });
        throw error;
      }
    } catch (error) {
      this.logger.error(`Erro ao validar OTP: ${error instanceof Error ? error.message : String(error)}`);
      
      if (this.config.observability?.metricsEnabled && this.metrics) {
        this.metrics.incrementCounter('mobileMoney.otp.validationError', {
          verificationId: request.verificationId
        });
      }
      
      throw error;
    }
  }
  
  /**
   * Obtém detalhes do usuário Mobile Money
   * @param phoneNumber Número de telefone
   * @param telemetryContext Contexto de telemetria opcional
   * @returns Detalhes do usuário
   */
  public async getUserDetails(
    phoneNumber: string,
    telemetryContext?: Partial<TelemetryContext>
  ): Promise<MobileMoneyUserDetails> {
    const context = this.createTelemetryContext(telemetryContext);
    
    try {
      // Normalizar formato do número
      const normalizedPhone = this.normalizePhoneNumber(phoneNumber);
      
      const response = await this.client.get<MobileMoneyUserDetails>(
        `/v1/users/phone/${normalizedPhone}`,
        {
          headers: {
            'X-Telemetry-Context': JSON.stringify(context)
          }
        }
      );
      
      return response.data;
    } catch (error) {
      this.logger.error(`Erro ao obter detalhes do usuário: ${error instanceof Error ? error.message : String(error)}`);
      throw error;
    }
  }
  
  /**
   * Normaliza um número de telefone para o formato padrão
   * @param phoneNumber Número a ser normalizado
   * @returns Número normalizado
   */
  private normalizePhoneNumber(phoneNumber: string): string {
    // Remover caracteres não numéricos
    let normalized = phoneNumber.replace(/\D/g, '');
    
    // Assegurar que o número tem código do país
    if (normalized.length < 10) {
      throw new Error(`Número de telefone inválido: ${phoneNumber}`);
    }
    
    if (!normalized.startsWith('+')) {
      normalized = '+' + normalized;
    }
    
    return normalized;
  }
  
  /**
   * Verifica se um número de telefone existe
   * @param phoneNumber Número de telefone
   * @param operator Operadora (opcional)
   * @param telemetryContext Contexto de telemetria opcional
   * @returns Resultado da verificação
   */
  public async validatePhoneNumberExists(
    phoneNumber: string,
    operator?: string,
    telemetryContext?: Partial<TelemetryContext>
  ): Promise<{
    exists: boolean;
    valid: boolean;
    operator?: string;
    operatorName?: string;
    type?: 'mobile' | 'landline' | 'voip' | 'unknown';
    countryCode?: string;
  }> {
    const context = this.createTelemetryContext(telemetryContext);
    
    try {
      // Verificar se a validação está habilitada
      if (!this.mobileMoneyConfig.phoneValidation?.enabled) {
        throw new Error('Validação de número de telefone não está habilitada');
      }
      
      // Normalizar o número
      const normalizedPhone = this.normalizePhoneNumber(phoneNumber);
      
      const response = await this.client.get(
        '/v1/phone/validate',
        {
          params: {
            phone_number: normalizedPhone,
            operator,
            existence_check: this.mobileMoneyConfig.phoneValidation?.existenceCheck || true,
            status_check: this.mobileMoneyConfig.phoneValidation?.statusCheck || false
          },
          headers: {
            'X-Telemetry-Context': JSON.stringify(context)
          }
        }
      );
      
      // Registrar métricas do resultado
      if (this.config.observability?.metricsEnabled && this.metrics) {
        this.metrics.incrementCounter('mobileMoney.phoneValidation', {
          exists: String(response.data.exists),
          valid: String(response.data.valid),
          operator: response.data.operator || 'unknown'
        });
      }
      
      return response.data;
    } catch (error) {
      this.logger.error(`Erro ao validar existência do número de telefone: ${error instanceof Error ? error.message : String(error)}`);
      
      if (this.config.observability?.metricsEnabled && this.metrics) {
        this.metrics.incrementCounter('mobileMoney.phoneValidation.error', {
          error: error instanceof Error ? error.name : 'UnknownError'
        });
      }
      
      throw error;
    }
  }
}