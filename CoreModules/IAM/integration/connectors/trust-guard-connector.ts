/**
 * @file trust-guard-connector.ts
 * @description Conector de integração entre IAM e TrustGuard para avaliação de risco e segurança
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';
import { AbstractConnector, BaseConnectorConfig, ConnectorStatus, TelemetryContext } from './base-connector';
import { Logger } from '../../observability/logging/hook_logger';
import { MetricsCollector } from '../../observability/metrics/hook_metrics';
import { TracingProvider } from '../../observability/tracing/hook_tracing';
import { RiskLevel, SecurityScore, TrustGuardAssessment } from '../../src/app/trust_guard_models';

/**
 * Configuração específica para o conector TrustGuard
 */
export interface TrustGuardConnectorConfig extends BaseConnectorConfig {
  /** Versão da API do TrustGuard */
  apiVersion: string;
  
  /** ID do tenant para comunicação com TrustGuard */
  tenantId: string;
  
  /** Configurações específicas de avaliação de risco */
  riskAssessment?: {
    /** Limiar de risco para bloqueio automático */
    blockThreshold: number;
    /** Se deve aplicar políticas adaptativas */
    adaptivePoliciesEnabled: boolean;
    /** Se deve usar machine learning para detecção */
    mlDetectionEnabled: boolean;
  };
  
  /** Configurações de compartilhamento de dados */
  dataSharing?: {
    /** Se compartilhamento de dados está habilitado */
    enabled: boolean;
    /** Categorias de dados permitidas para compartilhamento */
    allowedCategories: string[];
  };
  
  /** Configurações de contexto multi-regional */
  multiRegional?: {
    /** Região principal */
    primaryRegion: string;
    /** Regiões secundárias para failover */
    secondaryRegions: string[];
    /** Estratégia de roteamento */
    routingStrategy: 'latency' | 'geo' | 'failover';
  };
}

/**
 * Interface para avaliação de usuário
 */
export interface UserAssessmentRequest {
  /** ID do usuário */
  userId: string;
  
  /** Nome do usuário (opcional) */
  userName?: string;
  
  /** Email do usuário (opcional) */
  email?: string;
  
  /** Número de telefone (opcional) */
  phoneNumber?: string;
  
  /** Endereço IP do usuário */
  ipAddress: string;
  
  /** Informações do dispositivo */
  device: {
    /** ID do dispositivo */
    id?: string;
    /** Tipo do dispositivo */
    type: 'mobile' | 'desktop' | 'tablet' | 'other';
    /** Sistema operacional */
    os?: string;
    /** Browser utilizado */
    browser?: string;
    /** Hash de fingerprint do dispositivo */
    fingerprint?: string;
  };
  
  /** Geolocalização */
  geolocation?: {
    /** País */
    country?: string;
    /** Cidade */
    city?: string;
    /** Latitude */
    latitude?: number;
    /** Longitude */
    longitude?: number;
  };
  
  /** Contexto da ação */
  context: {
    /** Tipo da ação */
    actionType: 'login' | 'registration' | 'transaction' | 'profile_update' | 'password_reset' | string;
    /** Valor da transação (se aplicável) */
    transactionValue?: number;
    /** Moeda da transação (se aplicável) */
    currency?: string;
    /** ID da sessão */
    sessionId?: string;
    /** Fatores de autenticação utilizados */
    authFactors?: string[];
    /** Metadados adicionais específicos do contexto */
    metadata?: Record<string, any>;
  };
}

/**
 * Resposta de avaliação de usuário do TrustGuard
 */
export interface UserAssessmentResponse {
  /** ID da avaliação */
  assessmentId: string;
  
  /** Pontuação de risco de 0 (baixo) a 100 (alto) */
  riskScore: number;
  
  /** Nível de risco categorizado */
  riskLevel: RiskLevel;
  
  /** Pontuação de confiança/segurança */
  trustScore: SecurityScore;
  
  /** Se a ação deve ser bloqueada */
  shouldBlock: boolean;
  
  /** Se requer verificação adicional */
  requiresAdditionalVerification: boolean;
  
  /** Fatores adicionais de autenticação recomendados */
  recommendedAuthFactors?: string[];
  
  /** Limites adaptados com base no perfil */
  adaptiveThresholds?: {
    /** Limite para transações de baixo valor */
    lowValueThreshold?: number;
    /** Limite para transações de alto valor */
    highValueThreshold?: number;
  };
  
  /** Detalhes das verificações realizadas */
  assessmentDetails: TrustGuardAssessment;
  
  /** Timestamp da avaliação */
  timestamp: string;
}

/**
 * Conector para integração com o serviço TrustGuard
 */
export class TrustGuardConnector extends AbstractConnector {
  /** Cliente HTTP para comunicação com TrustGuard */
  private client: AxiosInstance;
  
  /** Configuração específica de TrustGuard */
  private trustGuardConfig: TrustGuardConnectorConfig;
  
  /**
   * Construtor do conector TrustGuard
   * @param config Configuração do conector
   * @param logger Logger para o conector
   * @param metrics Coletor de métricas opcional
   * @param tracer Provedor de rastreamento opcional
   */
  constructor(
    config: TrustGuardConnectorConfig,
    logger: Logger,
    metrics?: MetricsCollector,
    tracer?: TracingProvider
  ) {
    super(config, logger, metrics, tracer);
    this.trustGuardConfig = config;
  }
  
  /**
   * Valida a configuração específica de TrustGuard
   * @throws Error se a configuração for inválida
   */
  protected validateConfig(): void {
    super.validateConfig();
    
    if (!this.trustGuardConfig.apiVersion) {
      throw new Error('apiVersion é obrigatório na configuração do TrustGuard');
    }
    
    if (!this.trustGuardConfig.tenantId) {
      throw new Error('tenantId é obrigatório na configuração do TrustGuard');
    }
  }
  
  /**
   * Inicializa o conector TrustGuard
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
          'X-API-Version': this.trustGuardConfig.apiVersion,
          'X-Tenant-ID': this.trustGuardConfig.tenantId
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
      this.logger.error(`Erro ao inicializar conector TrustGuard: ${error instanceof Error ? error.message : String(error)}`);
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
          this.metrics.recordValue('trustGuard.request.latency', latencyMs, {
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
            this.metrics.recordValue('trustGuard.request.latency', latencyMs, {
              path: error.config?.url || '',
              method: error.config?.method || 'unknown',
              status: error.response.status.toString(),
              error: 'true'
            });
          }
          
          this.logger.warn(`Requisição TrustGuard falhou: ${error.response.status} ${JSON.stringify(error.response.data)}`);
        } else {
          this.logger.error(`Erro na requisição TrustGuard: ${error.message}`);
        }
        
        return Promise.reject(error);
      }
    );
  }
  
  /**
   * Realiza health check no serviço TrustGuard
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
          environment: response.data.environment
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
   * Fecha conexão com o serviço TrustGuard
   */
  public async shutdown(): Promise<void> {
    this.logger.info('Encerrando conexão com TrustGuard');
    this.status = ConnectorStatus.DISCONNECTED;
  }
  
  /**
   * Realiza avaliação de risco de um usuário
   * @param request Dados para avaliação do usuário
   * @param telemetryContext Contexto de telemetria opcional
   * @returns Resultado da avaliação
   */
  public async assessUser(
    request: UserAssessmentRequest,
    telemetryContext?: Partial<TelemetryContext>
  ): Promise<UserAssessmentResponse> {
    try {
      const context = this.createTelemetryContext(telemetryContext);
      
      // Iniciar span de tracing se disponível
      const span = this.tracer?.createSpan('trustGuard.assessUser', context);
      
      try {
        // Configurar dados adicionais de contexto
        const requestData = {
          ...request,
          tenantId: this.trustGuardConfig.tenantId,
          assessmentConfig: {
            mlDetectionEnabled: this.trustGuardConfig.riskAssessment?.mlDetectionEnabled ?? true,
            adaptivePoliciesEnabled: this.trustGuardConfig.riskAssessment?.adaptivePoliciesEnabled ?? true
          }
        };
        
        // Registrar métricas de tentativa
        if (this.config.observability?.metricsEnabled && this.metrics) {
          this.metrics.incrementCounter('trustGuard.assessUser.attempts', {
            actionType: request.context.actionType,
            deviceType: request.device.type
          });
        }
        
        // Adicionar contexto de telemetria nos headers
        const headers = {
          'X-Telemetry-Context': JSON.stringify(context)
        };
        
        // Realizar a requisição
        const response = await this.client.post<UserAssessmentResponse>(
          '/v1/risk/assessment',
          requestData,
          { headers }
        );
        
        // Registrar métricas de sucesso
        if (this.config.observability?.metricsEnabled && this.metrics) {
          this.metrics.incrementCounter('trustGuard.assessUser.success', {
            actionType: request.context.actionType,
            riskLevel: response.data.riskLevel,
            shouldBlock: String(response.data.shouldBlock)
          });
          
          this.metrics.recordValue('trustGuard.assessUser.riskScore', response.data.riskScore, {
            actionType: request.context.actionType,
            deviceType: request.device.type
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
      this.logger.error(`Erro ao realizar avaliação de usuário: ${error instanceof Error ? error.message : String(error)}`);
      
      // Registrar métricas de falha
      if (this.config.observability?.metricsEnabled && this.metrics) {
        this.metrics.incrementCounter('trustGuard.assessUser.error', {
          actionType: request.context.actionType,
          error: error instanceof Error ? error.name : 'UnknownError'
        });
      }
      
      throw error;
    }
  }
  
  /**
   * Reporta um evento de fraude confirmada
   * @param userId ID do usuário
   * @param incidentDetails Detalhes do incidente
   * @param telemetryContext Contexto de telemetria opcional
   * @returns Confirmação do registro
   */
  public async reportFraud(
    userId: string,
    incidentDetails: {
      type: 'account_takeover' | 'identity_theft' | 'payment_fraud' | 'synthetic_identity' | 'other';
      description: string;
      severity: 'low' | 'medium' | 'high' | 'critical';
      evidenceData?: Record<string, any>;
      reportedBy?: string;
    },
    telemetryContext?: Partial<TelemetryContext>
  ): Promise<{
    success: boolean;
    incidentId: string;
    timestamp: string;
  }> {
    const context = this.createTelemetryContext(telemetryContext);
    
    try {
      const span = this.tracer?.createSpan('trustGuard.reportFraud', context);
      
      try {
        const response = await this.client.post('/v1/fraud/report', {
          userId,
          ...incidentDetails,
          tenantId: this.trustGuardConfig.tenantId,
          timestamp: new Date().toISOString()
        }, {
          headers: {
            'X-Telemetry-Context': JSON.stringify(context)
          }
        });
        
        // Registrar métricas de reporte de fraude
        if (this.config.observability?.metricsEnabled && this.metrics) {
          this.metrics.incrementCounter('trustGuard.reportFraud', {
            fraudType: incidentDetails.type,
            severity: incidentDetails.severity
          });
        }
        
        span?.end('success');
        return response.data;
      } catch (error) {
        span?.end('error', { error: error instanceof Error ? error.message : String(error) });
        throw error;
      }
    } catch (error) {
      this.logger.error(`Erro ao reportar fraude: ${error instanceof Error ? error.message : String(error)}`);
      
      if (this.config.observability?.metricsEnabled && this.metrics) {
        this.metrics.incrementCounter('trustGuard.reportFraud.error', {
          fraudType: incidentDetails.type
        });
      }
      
      throw error;
    }
  }
  
  /**
   * Obtém o histórico de avaliações de um usuário
   * @param userId ID do usuário
   * @param options Opções de consulta
   * @param telemetryContext Contexto de telemetria opcional
   * @returns Histórico de avaliações
   */
  public async getUserAssessmentHistory(
    userId: string,
    options: {
      startDate?: string;
      endDate?: string;
      limit?: number;
      offset?: number;
      includeDetails?: boolean;
    } = {},
    telemetryContext?: Partial<TelemetryContext>
  ): Promise<{
    userId: string;
    assessments: UserAssessmentResponse[];
    total: number;
    page: number;
    pageSize: number;
  }> {
    const context = this.createTelemetryContext(telemetryContext);
    
    try {
      const response = await this.client.get(`/v1/risk/user/${userId}/history`, {
        params: {
          startDate: options.startDate,
          endDate: options.endDate,
          limit: options.limit || 20,
          offset: options.offset || 0,
          includeDetails: options.includeDetails || false
        },
        headers: {
          'X-Telemetry-Context': JSON.stringify(context)
        }
      });
      
      return response.data;
    } catch (error) {
      this.logger.error(`Erro ao obter histórico de avaliações: ${error instanceof Error ? error.message : String(error)}`);
      throw error;
    }
  }
  
  /**
   * Obtém o perfil de risco atual do usuário
   * @param userId ID do usuário
   * @param telemetryContext Contexto de telemetria opcional
   * @returns Perfil de risco atual
   */
  public async getUserRiskProfile(
    userId: string,
    telemetryContext?: Partial<TelemetryContext>
  ): Promise<{
    userId: string;
    trustScore: SecurityScore;
    riskLevel: RiskLevel;
    lastUpdated: string;
    behavioralPatterns: {
      loginLocations: Array<{ country: string; city: string; frequency: number }>;
      deviceUsage: Array<{ deviceType: string; frequency: number }>;
      activityHours: Array<{ hour: number; frequency: number }>;
      transactionPatterns?: {
        averageValue: number;
        currency: string;
        frequency: string;
      };
    };
    adaptiveThresholds: {
      lowValueThreshold: number;
      highValueThreshold: number;
      requiredAuthFactorsByRisk: Record<RiskLevel, string[]>;
    };
  }> {
    const context = this.createTelemetryContext(telemetryContext);
    
    try {
      const response = await this.client.get(`/v1/risk/user/${userId}/profile`, {
        headers: {
          'X-Telemetry-Context': JSON.stringify(context)
        }
      });
      
      return response.data;
    } catch (error) {
      this.logger.error(`Erro ao obter perfil de risco do usuário: ${error instanceof Error ? error.message : String(error)}`);
      throw error;
    }
  }
  
  /**
   * Verifica se um usuário está em uma lista de bloqueio
   * @param userId ID do usuário
   * @param telemetryContext Contexto de telemetria opcional
   * @returns Status do bloqueio
   */
  public async checkUserBlocklist(
    userId: string,
    telemetryContext?: Partial<TelemetryContext>
  ): Promise<{
    isBlocked: boolean;
    reason?: string;
    expiresAt?: string;
    blockId?: string;
  }> {
    const context = this.createTelemetryContext(telemetryContext);
    
    try {
      const response = await this.client.get(`/v1/blocklist/user/${userId}`, {
        headers: {
          'X-Telemetry-Context': JSON.stringify(context)
        }
      });
      
      return response.data;
    } catch (error) {
      if (axios.isAxiosError(error) && error.response?.status === 404) {
        // 404 significa que o usuário não está na blocklist
        return { isBlocked: false };
      }
      
      this.logger.error(`Erro ao verificar blocklist: ${error instanceof Error ? error.message : String(error)}`);
      throw error;
    }
  }
}