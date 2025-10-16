/**
 * Adaptador para Fontes Externas de Dados de Crédito
 * 
 * Este módulo implementa adaptadores para conexão com diferentes
 * provedores de dados de crédito externos, oferecendo uma interface
 * unificada para consulta e integração.
 * 
 * @module ExternalCreditAdapter
 */

import { Logger } from '../../../observability/logging/hook_logger';
import { Metrics } from '../../../observability/metrics/hook_metrics';
import { Tracer } from '../../../observability/tracing/hook_tracing';
import axios, { AxiosInstance } from 'axios';

/**
 * Tipos de provedores de dados de crédito
 */
export enum CreditDataProviderType {
  BUREAU_CREDITO = 'BUREAU_CREDITO',
  CENTRAL_BANCO = 'CENTRAL_BANCO',
  SERASA = 'SERASA',
  SCPC = 'SCPC',
  EXPERIAN = 'EXPERIAN',
  TRANSUNION = 'TRANSUNION',
  EQUIFAX = 'EQUIFAX',
  CREDIT_INFO = 'CREDIT_INFO',
  CUSTOM = 'CUSTOM'
}

/**
 * Tipos de dados de crédito
 */
export enum CreditDataType {
  CREDIT_SCORE = 'CREDIT_SCORE',
  CREDIT_HISTORY = 'CREDIT_HISTORY',
  CREDIT_LIMITS = 'CREDIT_LIMITS',
  PAYMENT_HISTORY = 'PAYMENT_HISTORY',
  DEFAULTS = 'DEFAULTS',
  BANKRUPTCIES = 'BANKRUPTCIES',
  LEGAL_PROCEEDINGS = 'LEGAL_PROCEEDINGS',
  INCOME_VERIFICATION = 'INCOME_VERIFICATION',
  EMPLOYMENT_HISTORY = 'EMPLOYMENT_HISTORY',
  ADDRESS_HISTORY = 'ADDRESS_HISTORY',
  IDENTITY_VERIFICATION = 'IDENTITY_VERIFICATION',
  FRAUD_ALERTS = 'FRAUD_ALERTS',
  CREDIT_UTILIZATION = 'CREDIT_UTILIZATION',
  ACCOUNT_DETAILS = 'ACCOUNT_DETAILS',
  CREDIT_INQUIRIES = 'CREDIT_INQUIRIES'
}

/**
 * Configuração do adaptador de crédito
 */
export interface CreditAdapterConfig {
  providerType: CreditDataProviderType;
  baseUrl: string;
  apiKey: string;
  apiSecret?: string;
  timeout?: number;
  retryAttempts?: number;
  retryDelayMs?: number;
  cacheTimeMs?: number;
  region?: string;
  version?: string;
  additionalHeaders?: Record<string, string>;
  providerSpecificConfig?: Record<string, any>;
}

/**
 * Parâmetros de consulta de crédito
 */
export interface CreditQueryParams {
  userId: string;
  tenantId: string;
  documentType: 'CPF' | 'CNPJ' | 'PASSPORT' | 'ID' | 'TAX_ID' | string;
  documentNumber: string;
  name?: string;
  birthDate?: Date | string;
  requestId: string;
  dataTypes?: CreditDataType[];
  additionalParams?: Record<string, any>;
  includeRawData?: boolean;
  requestReason?: string;
  consentId?: string;
}

/**
 * Estrutura do resultado de consulta de crédito
 */
export interface CreditQueryResult {
  requestId: string;
  userId: string;
  tenantId: string;
  timestamp: Date;
  providerType: CreditDataProviderType;
  responseCode: string;
  responseStatus: 'SUCCESS' | 'PARTIAL' | 'ERROR' | 'NO_DATA';
  creditScore?: number;
  creditScoreScale?: {
    min: number;
    max: number;
    provider: string;
    category?: string;
  };
  creditScoreHistory?: Array<{
    score: number;
    timestamp: Date;
    provider: string;
  }>;
  riskCategory?: string;
  activeCreditAccounts?: number;
  totalCreditLimit?: number;
  totalBalance?: number;
  creditUtilizationRate?: number;
  paymentDefaults?: Array<{
    creditor: string;
    amount: number;
    currency: string;
    daysOverdue: number;
    date: Date;
  }>;
  legalProceedings?: Array<{
    type: string;
    court: string;
    caseNumber: string;
    status: string;
    filingDate: Date;
    amount?: number;
    currency?: string;
  }>;
  fraudAlerts?: Array<{
    type: string;
    description: string;
    severity: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
    timestamp: Date;
  }>;
  inquiries?: Array<{
    requestor: string;
    reason: string;
    date: Date;
  }>;
  addressVerification?: {
    verified: boolean;
    score: number;
    details?: string;
  };
  identityVerification?: {
    verified: boolean;
    score: number;
    details?: string;
  };
  recommendations?: string[];
  warnings?: string[];
  dataCompleteness?: number;
  dataFreshness?: Date;
  processingTimeMs: number;
  rawData?: Record<string, any>;
  errors?: Array<{
    code: string;
    message: string;
    field?: string;
  }>;
}

/**
 * Interface para o adaptador de crédito
 */
export interface ICreditAdapter {
  initialize(): Promise<void>;
  queryCreditData(params: CreditQueryParams): Promise<CreditQueryResult>;
  testConnection(): Promise<boolean>;
  getProviderType(): CreditDataProviderType;
  getSupportedDataTypes(): CreditDataType[];
  getProviderStatus(): Promise<{
    available: boolean;
    latencyMs?: number;
    message?: string;
  }>;
}

/**
 * Classe base para adaptadores de crédito
 */
export abstract class BaseCreditAdapter implements ICreditAdapter {
  protected logger: Logger;
  protected metrics: Metrics;
  protected tracer: Tracer;
  protected config: CreditAdapterConfig;
  protected initialized: boolean = false;
  protected httpClient: AxiosInstance;
  
  /**
   * Construtor do adaptador base
   */
  constructor(
    logger: Logger,
    metrics: Metrics,
    tracer: Tracer,
    config: CreditAdapterConfig
  ) {
    this.logger = logger;
    this.metrics = metrics;
    this.tracer = tracer;
    this.config = config;
    
    // Criar cliente HTTP
    this.httpClient = axios.create({
      baseURL: config.baseUrl,
      timeout: config.timeout || 15000,
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        'X-API-Key': config.apiKey,
        ...config.additionalHeaders
      }
    });
  }
  
  /**
   * Inicializa o adaptador
   */
  public async initialize(): Promise<void> {
    if (this.initialized) {
      return;
    }
    
    try {
      this.logger.info({
        message: `Inicializando adaptador para ${this.config.providerType}`,
        provider: this.config.providerType
      });
      
      // Verificar conexão com o provedor
      const isConnected = await this.testConnection();
      
      if (!isConnected) {
        this.logger.error({
          message: `Falha ao conectar com o provedor ${this.config.providerType}`,
          provider: this.config.providerType,
          baseUrl: this.config.baseUrl
        });
        
        throw new Error(`Falha ao conectar com o provedor ${this.config.providerType}`);
      }
      
      this.initialized = true;
      
      this.logger.info({
        message: `Adaptador para ${this.config.providerType} inicializado com sucesso`,
        provider: this.config.providerType
      });
    } catch (error) {
      this.logger.error({
        message: `Erro ao inicializar adaptador para ${this.config.providerType}`,
        error: error.message,
        stack: error.stack,
        provider: this.config.providerType
      });
      
      throw error;
    }
  }
  
  /**
   * Consulta dados de crédito (a ser implementado por subclasses)
   */
  public abstract queryCreditData(params: CreditQueryParams): Promise<CreditQueryResult>;
  
  /**
   * Testa a conexão com o provedor (a ser implementado por subclasses)
   */
  public abstract testConnection(): Promise<boolean>;
  
  /**
   * Retorna o tipo de provedor
   */
  public getProviderType(): CreditDataProviderType {
    return this.config.providerType;
  }
  
  /**
   * Retorna os tipos de dados suportados (a ser implementado por subclasses)
   */
  public abstract getSupportedDataTypes(): CreditDataType[];
  
  /**
   * Verifica o status do provedor (a ser implementado por subclasses)
   */
  public abstract getProviderStatus(): Promise<{
    available: boolean;
    latencyMs?: number;
    message?: string;
  }>;
  
  /**
   * Registra métricas da consulta
   */
  protected recordMetrics(
    params: CreditQueryParams,
    result: CreditQueryResult,
    startTime: number
  ): void {
    const endTime = Date.now();
    const duration = endTime - startTime;
    
    this.metrics.histogram('credit_adapter.query_duration', duration, {
      provider: this.config.providerType,
      status: result.responseStatus,
      tenant_id: params.tenantId
    });
    
    if (result.responseStatus === 'SUCCESS') {
      this.metrics.increment('credit_adapter.query_success', {
        provider: this.config.providerType,
        tenant_id: params.tenantId
      });
    } else {
      this.metrics.increment('credit_adapter.query_failure', {
        provider: this.config.providerType,
        status: result.responseStatus,
        tenant_id: params.tenantId
      });
    }
    
    if (result.creditScore) {
      this.metrics.gauge('credit_adapter.credit_score', result.creditScore, {
        provider: this.config.providerType,
        tenant_id: params.tenantId
      });
    }
  }
}

/**
 * Adaptador específico para Bureau de Crédito
 */
export class BureauCreditoAdapter extends BaseCreditAdapter {
  /**
   * Consulta dados de crédito no Bureau de Crédito
   */
  public async queryCreditData(params: CreditQueryParams): Promise<CreditQueryResult> {
    const startTime = Date.now();
    const span = this.tracer.startSpan('bureau_credito.query_credit_data');
    
    try {
      if (!this.initialized) {
        await this.initialize();
      }
      
      this.logger.info({
        message: 'Consultando dados de crédito no Bureau de Crédito',
        requestId: params.requestId,
        userId: params.userId,
        documentType: params.documentType,
        dataTypes: params.dataTypes
      });
      
      // Preparar payload para requisição
      const payload = {
        requestId: params.requestId,
        documentType: params.documentType,
        documentNumber: params.documentNumber,
        name: params.name,
        birthDate: params.birthDate,
        dataTypes: params.dataTypes || Object.values(CreditDataType),
        additionalParams: params.additionalParams || {},
        requestReason: params.requestReason,
        consentId: params.consentId
      };
      
      // Enviar requisição para o Bureau de Crédito
      const response = await this.httpClient.post('/api/v1/credit-data', payload);
      
      // Processar resposta
      const creditData = response.data;
      
      // Mapear resposta para o formato padrão
      const result: CreditQueryResult = {
        requestId: params.requestId,
        userId: params.userId,
        tenantId: params.tenantId,
        timestamp: new Date(),
        providerType: CreditDataProviderType.BUREAU_CREDITO,
        responseCode: creditData.responseCode || '200',
        responseStatus: creditData.status || 'SUCCESS',
        creditScore: creditData.score?.value,
        creditScoreScale: creditData.score?.scale ? {
          min: creditData.score.scale.min,
          max: creditData.score.scale.max,
          provider: creditData.score.provider,
          category: creditData.score.category
        } : undefined,
        creditScoreHistory: creditData.scoreHistory?.map(item => ({
          score: item.value,
          timestamp: new Date(item.date),
          provider: item.provider
        })),
        riskCategory: creditData.riskCategory,
        activeCreditAccounts: creditData.accounts?.active,
        totalCreditLimit: creditData.accounts?.totalLimit,
        totalBalance: creditData.accounts?.totalBalance,
        creditUtilizationRate: creditData.accounts?.utilizationRate,
        paymentDefaults: creditData.defaults?.map(item => ({
          creditor: item.creditorName,
          amount: item.amount.value,
          currency: item.amount.currency,
          daysOverdue: item.daysOverdue,
          date: new Date(item.date)
        })),
        legalProceedings: creditData.legalProceedings?.map(item => ({
          type: item.type,
          court: item.court,
          caseNumber: item.caseId,
          status: item.status,
          filingDate: new Date(item.filingDate),
          amount: item.amount?.value,
          currency: item.amount?.currency
        })),
        fraudAlerts: creditData.alerts?.map(item => ({
          type: item.type,
          description: item.description,
          severity: item.severity,
          timestamp: new Date(item.timestamp)
        })),
        inquiries: creditData.inquiries?.map(item => ({
          requestor: item.requestor,
          reason: item.reason,
          date: new Date(item.date)
        })),
        addressVerification: creditData.verifications?.address ? {
          verified: creditData.verifications.address.verified,
          score: creditData.verifications.address.score,
          details: creditData.verifications.address.details
        } : undefined,
        identityVerification: creditData.verifications?.identity ? {
          verified: creditData.verifications.identity.verified,
          score: creditData.verifications.identity.score,
          details: creditData.verifications.identity.details
        } : undefined,
        recommendations: creditData.recommendations,
        warnings: creditData.warnings,
        dataCompleteness: creditData.metadata?.completeness,
        dataFreshness: creditData.metadata?.lastUpdate ? new Date(creditData.metadata.lastUpdate) : undefined,
        processingTimeMs: Date.now() - startTime,
        rawData: params.includeRawData ? creditData : undefined,
        errors: creditData.errors?.map(err => ({
          code: err.code,
          message: err.message,
          field: err.field
        }))
      };
      
      // Registrar métricas
      this.recordMetrics(params, result, startTime);
      
      // Log de sucesso
      this.logger.info({
        message: 'Consulta de dados de crédito concluída com sucesso',
        requestId: params.requestId,
        responseStatus: result.responseStatus,
        processingTimeMs: result.processingTimeMs
      });
      
      return result;
    } catch (error) {
      // Log de erro
      this.logger.error({
        message: 'Erro ao consultar dados de crédito no Bureau de Crédito',
        requestId: params.requestId,
        error: error.message,
        stack: error.stack
      });
      
      // Registrar métrica de falha
      this.metrics.increment('credit_adapter.query_error', {
        provider: this.config.providerType,
        error_type: error.name,
        tenant_id: params.tenantId
      });
      
      // Retornar resultado com erro
      return {
        requestId: params.requestId,
        userId: params.userId,
        tenantId: params.tenantId,
        timestamp: new Date(),
        providerType: CreditDataProviderType.BUREAU_CREDITO,
        responseCode: error.response?.status?.toString() || '500',
        responseStatus: 'ERROR',
        processingTimeMs: Date.now() - startTime,
        errors: [{
          code: error.code || 'ERR_BUREAU_CREDITO',
          message: error.message || 'Erro ao consultar dados de crédito'
        }]
      };
    } finally {
      span.end();
    }
  }
  
  /**
   * Testa a conexão com o Bureau de Crédito
   */
  public async testConnection(): Promise<boolean> {
    try {
      const response = await this.httpClient.get('/api/v1/health');
      return response.status === 200 && response.data.status === 'UP';
    } catch (error) {
      this.logger.error({
        message: 'Erro ao testar conexão com Bureau de Crédito',
        error: error.message,
        stack: error.stack
      });
      return false;
    }
  }
  
  /**
   * Retorna os tipos de dados suportados pelo Bureau de Crédito
   */
  public getSupportedDataTypes(): CreditDataType[] {
    return [
      CreditDataType.CREDIT_SCORE,
      CreditDataType.CREDIT_HISTORY,
      CreditDataType.PAYMENT_HISTORY,
      CreditDataType.DEFAULTS,
      CreditDataType.LEGAL_PROCEEDINGS,
      CreditDataType.FRAUD_ALERTS,
      CreditDataType.IDENTITY_VERIFICATION,
      CreditDataType.ADDRESS_HISTORY,
      CreditDataType.CREDIT_INQUIRIES
    ];
  }
  
  /**
   * Verifica o status do provedor Bureau de Crédito
   */
  public async getProviderStatus(): Promise<{
    available: boolean;
    latencyMs?: number;
    message?: string;
  }> {
    try {
      const startTime = Date.now();
      const response = await this.httpClient.get('/api/v1/health');
      const latencyMs = Date.now() - startTime;
      
      return {
        available: response.status === 200 && response.data.status === 'UP',
        latencyMs,
        message: response.data.message || 'Serviço disponível'
      };
    } catch (error) {
      return {
        available: false,
        message: `Serviço indisponível: ${error.message}`
      };
    }
  }
}

/**
 * Fábrica de adaptadores de crédito
 */
export class CreditAdapterFactory {
  private static adapters: Map<string, ICreditAdapter> = new Map();
  
  /**
   * Cria ou obtém um adaptador para o provedor especificado
   */
  public static async getAdapter(
    providerType: CreditDataProviderType,
    config: CreditAdapterConfig,
    logger: Logger,
    metrics: Metrics,
    tracer: Tracer
  ): Promise<ICreditAdapter> {
    const adapterId = `${providerType}-${config.baseUrl}`;
    
    // Verificar se o adaptador já existe
    if (this.adapters.has(adapterId)) {
      return this.adapters.get(adapterId)!;
    }
    
    // Criar novo adaptador conforme o tipo
    let adapter: ICreditAdapter;
    
    switch (providerType) {
      case CreditDataProviderType.BUREAU_CREDITO:
        adapter = new BureauCreditoAdapter(logger, metrics, tracer, config);
        break;
        
      // Outros adaptadores podem ser adicionados aqui
        
      default:
        throw new Error(`Provedor de crédito não suportado: ${providerType}`);
    }
    
    // Inicializar adaptador
    await adapter.initialize();
    
    // Armazenar adaptador para reutilização
    this.adapters.set(adapterId, adapter);
    
    return adapter;
  }
}