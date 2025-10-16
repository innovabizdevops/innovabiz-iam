/**
 * Serviço de Bureau de Créditos
 * 
 * Este serviço orquestra a análise de risco, detecção de fraude e consulta
 * a dados externos de crédito, fornecendo uma interface unificada para
 * avaliação de transações financeiras.
 * 
 * @module BureauCreditoService
 */

import { Logger } from '../../observability/logging/hook_logger';
import { Metrics } from '../../observability/metrics/hook_metrics';
import { Tracer } from '../../observability/tracing/hook_tracing';
import { 
  RiskAssessmentModel, 
  TransactionRiskData,
  RiskAssessmentResult,
  RiskAction,
  RiskLevel
} from './models/risk-assessment-model';
import {
  FraudDetectionEngine,
  FraudDetectionResult
} from './models/fraud-detection-engine';
import {
  CreditAdapterFactory,
  CreditDataProviderType,
  CreditQueryParams,
  CreditQueryResult,
  ICreditAdapter
} from './adapters/external-credit-adapter';
import { v4 as uuidv4 } from 'uuid';

/**
 * Nível de verificação de identidade
 */
export enum IdentityVerificationLevel {
  NONE = 'NONE',
  BASIC = 'BASIC',
  MEDIUM = 'MEDIUM',
  STRONG = 'STRONG',
  VERY_STRONG = 'VERY_STRONG'
}

/**
 * Configuração do serviço de Bureau de Créditos
 */
export interface BureauCreditoConfig {
  riskAssessmentConfig: any;
  fraudDetectionConfig: any;
  creditAdapterConfigs: {
    [key in CreditDataProviderType]?: {
      baseUrl: string;
      apiKey: string;
      apiSecret?: string;
      timeout?: number;
      additionalConfig?: any;
    };
  };
  enableFraudDetection: boolean;
  enableRiskAssessment: boolean;
  enableCreditData: boolean;
  defaultCreditProvider?: CreditDataProviderType;
  cacheTimeMs?: number;
  auditOptions?: {
    enabled: boolean;
    detailLevel: string;
    retentionDays: number;
  };
  thresholds?: {
    highRiskThreshold: number;
    fraudSuspicionThreshold: number;
    maxCreditScoreThreshold: number;
    minCreditScoreThreshold: number;
  };
}

/**
 * Parâmetros para avaliação de transação
 */
export interface TransactionEvaluationParams {
  transactionId: string;
  userId: string;
  tenantId: string;
  documentType: string;
  documentNumber: string;
  transactionType: string;
  amount: number;
  currency: string;
  channel?: string;
  deviceId?: string;
  deviceFingerprint?: string;
  ipAddress?: string;
  countryCode?: string;
  location?: {
    latitude: number;
    longitude: number;
    accuracy: number;
  };
  timestamp?: Date;
  userAgent?: string;
  userMetadata?: Record<string, any>;
  transactionMetadata?: Record<string, any>;
  options?: {
    performRiskAssessment?: boolean;
    performFraudDetection?: boolean;
    fetchCreditData?: boolean;
    creditProviderType?: CreditDataProviderType;
    includeRawData?: boolean;
  };
}

/**
 * Resultado da avaliação de transação
 */
export interface TransactionEvaluationResult {
  evaluationId: string;
  transactionId: string;
  userId: string;
  timestamp: Date;
  approved: boolean;
  requiresReview: boolean;
  requiresAdditionalVerification: boolean;
  recommendedActions: string[];
  riskAssessment?: RiskAssessmentResult;
  fraudDetection?: FraudDetectionResult;
  creditData?: CreditQueryResult;
  overallRiskLevel: RiskLevel;
  overallRiskScore: number;
  identityVerificationLevel: IdentityVerificationLevel;
  processingTimeMs: number;
  errors?: Array<{
    component: string;
    code: string;
    message: string;
  }>;
}

/**
 * Serviço de Bureau de Créditos
 */
export class BureauCreditoService {
  private logger: Logger;
  private metrics: Metrics;
  private tracer: Tracer;
  private config: BureauCreditoConfig;
  private riskModel: RiskAssessmentModel;
  private fraudEngine: FraudDetectionEngine;
  private creditAdapters: Map<CreditDataProviderType, ICreditAdapter> = new Map();
  private dataProvider: any; // Serviço para obter dados de contexto
  private notificationService: any; // Serviço para notificações
  private cache: Map<string, { timestamp: number, data: any }> = new Map();
  
  /**
   * Construtor do serviço
   */
  constructor(
    logger: Logger,
    metrics: Metrics,
    tracer: Tracer,
    config: BureauCreditoConfig,
    dataProvider: any,
    notificationService: any
  ) {
    this.logger = logger;
    this.metrics = metrics;
    this.tracer = tracer;
    this.config = config;
    this.dataProvider = dataProvider;
    this.notificationService = notificationService;
    
    // Inicializar componentes
    this.riskModel = new RiskAssessmentModel(
      logger,
      metrics,
      tracer,
      config.riskAssessmentConfig
    );
    
    this.fraudEngine = new FraudDetectionEngine(
      logger,
      metrics,
      tracer,
      dataProvider,
      notificationService,
      config.fraudDetectionConfig
    );
    
    this.initializeAdapters();
  }
  
  /**
   * Inicializa os adaptadores de crédito
   */
  private async initializeAdapters(): Promise<void> {
    try {
      for (const [provider, config] of Object.entries(this.config.creditAdapterConfigs)) {
        const providerType = provider as CreditDataProviderType;
        
        this.logger.info({
          message: `Inicializando adaptador para ${providerType}`,
          provider: providerType
        });
        
        const adapter = await CreditAdapterFactory.getAdapter(
          providerType,
          {
            providerType,
            baseUrl: config.baseUrl,
            apiKey: config.apiKey,
            apiSecret: config.apiSecret,
            timeout: config.timeout,
            ...config.additionalConfig
          },
          this.logger,
          this.metrics,
          this.tracer
        );
        
        this.creditAdapters.set(providerType, adapter);
        
        this.logger.info({
          message: `Adaptador para ${providerType} inicializado com sucesso`,
          provider: providerType
        });
      }
    } catch (error) {
      this.logger.error({
        message: 'Erro ao inicializar adaptadores de crédito',
        error: error.message,
        stack: error.stack
      });
    }
  }
  
  /**
   * Avalia uma transação financeira
   */
  public async evaluateTransaction(
    params: TransactionEvaluationParams
  ): Promise<TransactionEvaluationResult> {
    const startTime = Date.now();
    const evaluationId = uuidv4();
    const span = this.tracer.startSpan('bureau_credito.evaluate_transaction');
    
    try {
      this.logger.info({
        message: 'Iniciando avaliação de transação',
        evaluationId,
        transactionId: params.transactionId,
        userId: params.userId,
        tenantId: params.tenantId
      });
      
      // Inicializar resultado
      const result: TransactionEvaluationResult = {
        evaluationId,
        transactionId: params.transactionId,
        userId: params.userId,
        timestamp: new Date(),
        approved: false,
        requiresReview: false,
        requiresAdditionalVerification: false,
        recommendedActions: [],
        overallRiskLevel: RiskLevel.MEDIUM,
        overallRiskScore: 50,
        identityVerificationLevel: IdentityVerificationLevel.NONE,
        processingTimeMs: 0,
        errors: []
      };
      
      // Configurar opções
      const options = {
        performRiskAssessment: params.options?.performRiskAssessment ?? this.config.enableRiskAssessment,
        performFraudDetection: params.options?.performFraudDetection ?? this.config.enableFraudDetection,
        fetchCreditData: params.options?.fetchCreditData ?? this.config.enableCreditData,
        creditProviderType: params.options?.creditProviderType ?? this.config.defaultCreditProvider,
        includeRawData: params.options?.includeRawData ?? false
      };
      
      // Preparar dados para avaliação de risco
      const riskData: TransactionRiskData = this.prepareRiskData(params);
      
      // Executar avaliações em paralelo
      const [riskResult, fraudResult, creditResult] = await Promise.all([
        // Avaliação de risco
        options.performRiskAssessment ? this.performRiskAssessment(riskData) : Promise.resolve(null),
        // Detecção de fraude
        options.performFraudDetection ? this.performFraudDetection(riskData) : Promise.resolve(null),
        // Dados de crédito
        options.fetchCreditData && options.creditProviderType ? 
          this.fetchCreditData(params, options.creditProviderType) : Promise.resolve(null)
      ]);
      
      // Processar resultados da avaliação de risco
      if (riskResult) {
        result.riskAssessment = riskResult;
        
        // Adicionar ações recomendadas
        riskResult.recommendedActions.forEach(action => {
          if (!result.recommendedActions.includes(action)) {
            result.recommendedActions.push(action);
          }
        });
        
        // Verificar se requer revisão
        if (riskResult.requiresManualReview) {
          result.requiresReview = true;
        }
        
        // Verificar se requer verificação adicional
        if (riskResult.recommendedActions.includes(RiskAction.ADDITIONAL_VERIFICATION)) {
          result.requiresAdditionalVerification = true;
        }
        
        // Atualizar nível de risco geral
        result.overallRiskLevel = riskResult.riskLevel;
        result.overallRiskScore = riskResult.overallScore;
      }
      
      // Processar resultados da detecção de fraude
      if (fraudResult) {
        result.fraudDetection = fraudResult;
        
        // Adicionar ações recomendadas
        fraudResult.suggestedActions.forEach(action => {
          if (!result.recommendedActions.includes(action)) {
            result.recommendedActions.push(action);
          }
        });
        
        // Verificar se requer revisão
        if (fraudResult.requiresManualReview) {
          result.requiresReview = true;
        }
        
        // Aumentar nível de risco se fraude for detectada
        if (fraudResult.fraudDetected) {
          result.overallRiskLevel = this.upgradeRiskLevel(result.overallRiskLevel);
          result.overallRiskScore = Math.max(result.overallRiskScore, fraudResult.overallScore);
        }
      }
      
      // Processar resultados dos dados de crédito
      if (creditResult) {
        result.creditData = creditResult;
        
        // Ajustar nível de risco com base no score de crédito
        if (creditResult.creditScore !== undefined) {
          const scoreImpact = this.calculateCreditScoreImpact(creditResult.creditScore, creditResult.creditScoreScale);
          result.overallRiskScore = Math.max(0, Math.min(100, result.overallRiskScore - scoreImpact));
          result.overallRiskLevel = this.determineRiskLevel(result.overallRiskScore);
        }
        
        // Determinar nível de verificação de identidade
        result.identityVerificationLevel = this.determineIdentityVerificationLevel(
          creditResult.identityVerification?.verified ?? false,
          creditResult.identityVerification?.score ?? 0
        );
      }
      
      // Determinar aprovação com base nas avaliações
      result.approved = this.determineApproval(result);
      
      // Calcular tempo de processamento
      result.processingTimeMs = Date.now() - startTime;
      
      // Registrar métricas
      this.recordMetrics(params, result);
      
      // Registrar resultado no log
      this.logger.info({
        message: `Avaliação de transação concluída: ${result.approved ? 'Aprovada' : 'Reprovada'}`,
        evaluationId,
        transactionId: params.transactionId,
        approved: result.approved,
        overallRiskLevel: result.overallRiskLevel,
        overallRiskScore: result.overallRiskScore,
        processingTimeMs: result.processingTimeMs
      });
      
      return result;
    } catch (error) {
      this.logger.error({
        message: 'Erro na avaliação de transação',
        evaluationId,
        transactionId: params.transactionId,
        error: error.message,
        stack: error.stack
      });
      
      // Registrar métrica de erro
      this.metrics.increment('bureau_credito.evaluation_error', {
        tenant_id: params.tenantId,
        error_type: error.name || 'unknown'
      });
      
      // Retornar resultado com erro
      return {
        evaluationId,
        transactionId: params.transactionId,
        userId: params.userId,
        timestamp: new Date(),
        approved: false,
        requiresReview: true,
        requiresAdditionalVerification: false,
        recommendedActions: ['REVIEW', 'RETRY'],
        overallRiskLevel: RiskLevel.HIGH,
        overallRiskScore: 75,
        identityVerificationLevel: IdentityVerificationLevel.NONE,
        processingTimeMs: Date.now() - startTime,
        errors: [{
          component: 'bureau_credito_service',
          code: error.code || 'EVALUATION_ERROR',
          message: error.message || 'Erro na avaliação de transação'
        }]
      };
    } finally {
      span.end();
    }
  }
  
  /**
   * Prepara os dados para avaliação de risco
   */
  private prepareRiskData(params: TransactionEvaluationParams): TransactionRiskData {
    return {
      transactionId: params.transactionId,
      userId: params.userId,
      tenantId: params.tenantId,
      transactionType: params.transactionType as any,
      amount: params.amount,
      currency: params.currency,
      timestamp: params.timestamp || new Date(),
      channel: params.channel || 'UNKNOWN',
      deviceId: params.deviceId,
      deviceFingerprint: params.deviceFingerprint,
      ipAddress: params.ipAddress,
      countryCode: params.countryCode,
      location: params.location,
      userAgent: params.userAgent,
      metadata: {
        ...params.userMetadata,
        ...params.transactionMetadata
      }
    };
  }
  
  /**
   * Executa avaliação de risco
   */
  private async performRiskAssessment(
    data: TransactionRiskData
  ): Promise<RiskAssessmentResult | null> {
    try {
      return await this.riskModel.evaluateRisk(data);
    } catch (error) {
      this.logger.error({
        message: 'Erro na avaliação de risco',
        transactionId: data.transactionId,
        error: error.message,
        stack: error.stack
      });
      
      return null;
    }
  }
  
  /**
   * Executa detecção de fraude
   */
  private async performFraudDetection(
    data: TransactionRiskData
  ): Promise<FraudDetectionResult | null> {
    try {
      return await this.fraudEngine.detectFraud(data);
    } catch (error) {
      this.logger.error({
        message: 'Erro na detecção de fraude',
        transactionId: data.transactionId,
        error: error.message,
        stack: error.stack
      });
      
      return null;
    }
  }
  
  /**
   * Busca dados de crédito
   */
  private async fetchCreditData(
    params: TransactionEvaluationParams,
    providerType: CreditDataProviderType
  ): Promise<CreditQueryResult | null> {
    try {
      const adapter = this.creditAdapters.get(providerType);
      
      if (!adapter) {
        this.logger.error({
          message: `Adaptador para ${providerType} não encontrado`,
          transactionId: params.transactionId
        });
        
        return null;
      }
      
      // Verificar cache
      const cacheKey = `credit_data:${params.userId}:${params.documentNumber}`;
      const cacheTimeMs = this.config.cacheTimeMs || 3600000; // 1 hora padrão
      const cachedData = this.cache.get(cacheKey);
      
      if (cachedData && (Date.now() - cachedData.timestamp) < cacheTimeMs) {
        this.logger.debug({
          message: 'Usando dados de crédito em cache',
          transactionId: params.transactionId,
          userId: params.userId
        });
        
        return cachedData.data;
      }
      
      // Preparar parâmetros de consulta
      const queryParams: CreditQueryParams = {
        userId: params.userId,
        tenantId: params.tenantId,
        documentType: params.documentType,
        documentNumber: params.documentNumber,
        requestId: params.transactionId,
        includeRawData: params.options?.includeRawData,
        requestReason: 'TRANSACTION_EVALUATION',
        additionalParams: params.transactionMetadata
      };
      
      // Consultar dados de crédito
      const result = await adapter.queryCreditData(queryParams);
      
      // Armazenar em cache
      if (result.responseStatus === 'SUCCESS') {
        this.cache.set(cacheKey, {
          timestamp: Date.now(),
          data: result
        });
      }
      
      return result;
    } catch (error) {
      this.logger.error({
        message: 'Erro ao buscar dados de crédito',
        transactionId: params.transactionId,
        error: error.message,
        stack: error.stack
      });
      
      return null;
    }
  }
  
  /**
   * Determina o nível de verificação de identidade
   */
  private determineIdentityVerificationLevel(
    verified: boolean,
    score: number
  ): IdentityVerificationLevel {
    if (!verified) return IdentityVerificationLevel.NONE;
    
    if (score >= 90) return IdentityVerificationLevel.VERY_STRONG;
    if (score >= 80) return IdentityVerificationLevel.STRONG;
    if (score >= 60) return IdentityVerificationLevel.MEDIUM;
    if (score >= 40) return IdentityVerificationLevel.BASIC;
    return IdentityVerificationLevel.NONE;
  }
  
  /**
   * Calcula o impacto do score de crédito no risco
   */
  private calculateCreditScoreImpact(
    score: number,
    scale?: { min: number; max: number; }
  ): number {
    const min = scale?.min || 0;
    const max = scale?.max || 1000;
    const range = max - min;
    
    // Normalizar score para 0-100
    const normalizedScore = ((score - min) / range) * 100;
    
    // Calcular impacto (score alto reduz o risco)
    // Um score de crédito de 100% pode reduzir o risco em até 30 pontos
    return (normalizedScore / 100) * 30;
  }
  
  /**
   * Aumenta o nível de risco para o próximo nível
   */
  private upgradeRiskLevel(currentLevel: RiskLevel): RiskLevel {
    switch (currentLevel) {
      case RiskLevel.VERY_LOW:
        return RiskLevel.LOW;
      case RiskLevel.LOW:
        return RiskLevel.MEDIUM;
      case RiskLevel.MEDIUM:
        return RiskLevel.HIGH;
      case RiskLevel.HIGH:
        return RiskLevel.VERY_HIGH;
      default:
        return RiskLevel.CRITICAL;
    }
  }
  
  /**
   * Determina o nível de risco com base na pontuação
   */
  private determineRiskLevel(score: number): RiskLevel {
    if (score >= 90) return RiskLevel.CRITICAL;
    if (score >= 80) return RiskLevel.VERY_HIGH;
    if (score >= 60) return RiskLevel.HIGH;
    if (score >= 40) return RiskLevel.MEDIUM;
    if (score >= 20) return RiskLevel.LOW;
    return RiskLevel.VERY_LOW;
  }
  
  /**
   * Determina se a transação deve ser aprovada
   */
  private determineApproval(result: TransactionEvaluationResult): boolean {
    // Rejeitar se houver detecção de fraude
    if (result.fraudDetection?.fraudDetected) {
      return false;
    }
    
    // Rejeitar se o risco for muito alto ou crítico
    if (
      result.overallRiskLevel === RiskLevel.VERY_HIGH ||
      result.overallRiskLevel === RiskLevel.CRITICAL
    ) {
      return false;
    }
    
    // Rejeitar se o score for muito alto
    const highRiskThreshold = this.config.thresholds?.highRiskThreshold || 80;
    if (result.overallRiskScore >= highRiskThreshold) {
      return false;
    }
    
    // Verificar ações recomendadas
    const blockingActions = ['BLOCK', 'REJECT', 'REPORT'];
    for (const action of blockingActions) {
      if (result.recommendedActions.includes(action)) {
        return false;
      }
    }
    
    return true;
  }
  
  /**
   * Registra métricas da avaliação
   */
  private recordMetrics(
    params: TransactionEvaluationParams,
    result: TransactionEvaluationResult
  ): void {
    // Métrica do score de risco
    this.metrics.gauge('bureau_credito.risk_score', result.overallRiskScore, {
      tenant_id: params.tenantId,
      approved: result.approved ? 'true' : 'false'
    });
    
    // Métrica do tempo de processamento
    this.metrics.histogram('bureau_credito.processing_time', result.processingTimeMs, {
      tenant_id: params.tenantId
    });
    
    // Métrica de aprovação/rejeição
    this.metrics.increment('bureau_credito.evaluation_result', {
      tenant_id: params.tenantId,
      result: result.approved ? 'approved' : 'rejected',
      risk_level: result.overallRiskLevel
    });
    
    // Métricas de componentes utilizados
    this.metrics.increment('bureau_credito.components_used', {
      tenant_id: params.tenantId,
      component: 'risk_assessment',
      used: result.riskAssessment ? 'true' : 'false'
    });
    
    this.metrics.increment('bureau_credito.components_used', {
      tenant_id: params.tenantId,
      component: 'fraud_detection',
      used: result.fraudDetection ? 'true' : 'false'
    });
    
    this.metrics.increment('bureau_credito.components_used', {
      tenant_id: params.tenantId,
      component: 'credit_data',
      used: result.creditData ? 'true' : 'false'
    });
  }
}