/**
 * Motor de Regras para Análise de Fraude
 * 
 * Este módulo implementa um motor de regras especializado na detecção
 * de fraudes em transações financeiras, com foco em detecção de padrões,
 * anomalias e comportamentos suspeitos.
 * 
 * @module FraudDetectionEngine
 */

import { Logger } from '../../../observability/logging/hook_logger';
import { Metrics } from '../../../observability/metrics/hook_metrics';
import { Tracer } from '../../../observability/tracing/hook_tracing';
import { 
  FinancialTransactionType,
  TransactionRiskData,
  RiskLevel
} from './risk-assessment-model';

/**
 * Tipos de detecção de fraude
 */
export enum FraudDetectionType {
  VELOCITY_CHECK = 'VELOCITY_CHECK',
  AMOUNT_ANOMALY = 'AMOUNT_ANOMALY',
  LOCATION_MISMATCH = 'LOCATION_MISMATCH',
  DEVICE_ANOMALY = 'DEVICE_ANOMALY',
  PATTERN_RECOGNITION = 'PATTERN_RECOGNITION',
  NETWORK_ANALYSIS = 'NETWORK_ANALYSIS',
  TIME_ANOMALY = 'TIME_ANOMALY',
  BEHAVIORAL_ANALYSIS = 'BEHAVIORAL_ANALYSIS',
  ACCOUNT_TAKEOVER = 'ACCOUNT_TAKEOVER',
  SYNTHETIC_IDENTITY = 'SYNTHETIC_IDENTITY'
}

/**
 * Níveis de confiança na detecção de fraude
 */
export enum FraudConfidenceLevel {
  VERY_LOW = 'VERY_LOW',
  LOW = 'LOW',
  MEDIUM = 'MEDIUM',
  HIGH = 'HIGH',
  VERY_HIGH = 'VERY_HIGH'
}

/**
 * Interface para uma regra de detecção de fraude
 */
export interface FraudRule {
  id: string;
  name: string;
  description: string;
  detectionType: FraudDetectionType;
  enabled: boolean;
  priority: number; // 1-100, quanto maior, maior a prioridade
  evaluator: (data: TransactionRiskData, context: FraudDetectionContext) => Promise<FraudRuleResult>;
  applicableTransactionTypes: FinancialTransactionType[];
  requiredDataPoints: string[];
  regulatoryFrameworks?: string[];
  version: string;
  createdAt: Date;
  updatedAt: Date;
  cooldownPeriodMinutes?: number;
}

/**
 * Interface para o contexto da detecção de fraude
 */
export interface FraudDetectionContext {
  recentTransactions?: Array<{
    transactionId: string;
    timestamp: Date;
    amount: number;
    currency: string;
    status: string;
    type: string;
  }>;
  userDevices?: Array<{
    deviceId: string;
    firstSeen: Date;
    lastSeen: Date;
    riskScore: number;
    isTrusted: boolean;
  }>;
  knownLocations?: Array<{
    countryCode: string;
    city?: string;
    firstSeen: Date;
    lastSeen: Date;
    frequency: number;
  }>;
  previousFraudAttempts?: Array<{
    timestamp: Date;
    fraudType: FraudDetectionType;
    confidenceLevel: FraudConfidenceLevel;
    resolution: string;
  }>;
  transactionVelocity?: {
    hourly: number;
    daily: number;
    weekly: number;
    hourlyAmount: number;
    dailyAmount: number;
    weeklyAmount: number;
  };
  userSegment?: string;
  riskTier?: string;
}

/**
 * Interface para o resultado da avaliação de uma regra de fraude
 */
export interface FraudRuleResult {
  ruleId: string;
  triggered: boolean;
  confidenceLevel: FraudConfidenceLevel;
  score: number; // 0-100
  reason: string;
  suggestedActions: string[];
  metadata?: Record<string, any>;
}

/**
 * Interface para o resultado da detecção de fraude
 */
export interface FraudDetectionResult {
  transactionId: string;
  timestamp: Date;
  fraudDetected: boolean;
  overallConfidenceLevel: FraudConfidenceLevel;
  overallScore: number; // 0-100
  triggeredRules: FraudRuleResult[];
  evaluatedRules: string[];
  processingTimeMs: number;
  suggestedActions: string[];
  requiresManualReview: boolean;
  fraudTypes: FraudDetectionType[];
  riskLevel: RiskLevel;
  additionalContext?: Record<string, any>;
}

/**
 * Motor de detecção de fraudes
 */
export class FraudDetectionEngine {
  private logger: Logger;
  private metrics: Metrics;
  private tracer: Tracer;
  private rules: FraudRule[] = [];
  private dataProvider: any; // Serviço para obter dados de contexto
  private notificationService: any; // Serviço para notificações de fraude
  private config: any;
  
  /**
   * Construtor do motor de detecção de fraude
   */
  constructor(
    logger: Logger,
    metrics: Metrics,
    tracer: Tracer,
    dataProvider: any,
    notificationService: any,
    config: any
  ) {
    this.logger = logger;
    this.metrics = metrics;
    this.tracer = tracer;
    this.dataProvider = dataProvider;
    this.notificationService = notificationService;
    this.config = config;
    
    this.initDefaultRules();
  }
  
  /**
   * Inicializa as regras padrão de detecção de fraude
   */
  private initDefaultRules(): void {
    // Regra 1: Verificação de velocidade de transação
    this.addRule({
      id: 'VELOCITY_CHECK_01',
      name: 'Verificação de Velocidade de Transação',
      description: 'Detecta múltiplas transações em curto período de tempo',
      detectionType: FraudDetectionType.VELOCITY_CHECK,
      enabled: true,
      priority: 80,
      evaluator: async (data: TransactionRiskData, context: FraudDetectionContext): Promise<FraudRuleResult> => {
        if (!context.transactionVelocity) {
          return {
            ruleId: 'VELOCITY_CHECK_01',
            triggered: false,
            confidenceLevel: FraudConfidenceLevel.VERY_LOW,
            score: 0,
            reason: 'Dados de velocidade de transação indisponíveis',
            suggestedActions: []
          };
        }
        
        const hourlyThreshold = this.config.velocityThresholds?.hourly || 5;
        const dailyThreshold = this.config.velocityThresholds?.daily || 20;
        const hourlyAmountThreshold = this.config.velocityThresholds?.hourlyAmount || 10000;
        
        if (context.transactionVelocity.hourly > hourlyThreshold) {
          return {
            ruleId: 'VELOCITY_CHECK_01',
            triggered: true,
            confidenceLevel: FraudConfidenceLevel.HIGH,
            score: 75,
            reason: `Velocidade de transação elevada: ${context.transactionVelocity.hourly} transações na última hora`,
            suggestedActions: ['STEP_UP_AUTH', 'MANUAL_REVIEW']
          };
        }
        
        if (context.transactionVelocity.hourlyAmount > hourlyAmountThreshold) {
          return {
            ruleId: 'VELOCITY_CHECK_01',
            triggered: true,
            confidenceLevel: FraudConfidenceLevel.MEDIUM,
            score: 60,
            reason: `Volume financeiro elevado: ${context.transactionVelocity.hourlyAmount} ${data.currency} na última hora`,
            suggestedActions: ['STEP_UP_AUTH']
          };
        }
        
        if (context.transactionVelocity.daily > dailyThreshold) {
          return {
            ruleId: 'VELOCITY_CHECK_01',
            triggered: true,
            confidenceLevel: FraudConfidenceLevel.MEDIUM,
            score: 55,
            reason: `Velocidade diária elevada: ${context.transactionVelocity.daily} transações nas últimas 24h`,
            suggestedActions: ['STEP_UP_AUTH']
          };
        }
        
        return {
          ruleId: 'VELOCITY_CHECK_01',
          triggered: false,
          confidenceLevel: FraudConfidenceLevel.LOW,
          score: 0,
          reason: 'Velocidade de transação dentro de limites normais',
          suggestedActions: []
        };
      },
      applicableTransactionTypes: Object.values(FinancialTransactionType),
      requiredDataPoints: ['transactionId', 'userId', 'timestamp'],
      version: '1.0.0',
      createdAt: new Date(),
      updatedAt: new Date(),
      cooldownPeriodMinutes: 60
    });
    
    // Regra 2: Detecção de mudança de dispositivo
    this.addRule({
      id: 'DEVICE_ANOMALY_01',
      name: 'Detecção de Dispositivo Não Reconhecido',
      description: 'Detecta transações realizadas em dispositivos não reconhecidos ou suspeitos',
      detectionType: FraudDetectionType.DEVICE_ANOMALY,
      enabled: true,
      priority: 70,
      evaluator: async (data: TransactionRiskData, context: FraudDetectionContext): Promise<FraudRuleResult> => {
        if (!data.deviceId || !context.userDevices) {
          return {
            ruleId: 'DEVICE_ANOMALY_01',
            triggered: false,
            confidenceLevel: FraudConfidenceLevel.VERY_LOW,
            score: 0,
            reason: 'Dados de dispositivo indisponíveis',
            suggestedActions: []
          };
        }
        
        // Verificar se o dispositivo é conhecido
        const knownDevice = context.userDevices.find(
          device => device.deviceId === data.deviceId
        );
        
        if (!knownDevice) {
          return {
            ruleId: 'DEVICE_ANOMALY_01',
            triggered: true,
            confidenceLevel: FraudConfidenceLevel.MEDIUM,
            score: 65,
            reason: 'Transação realizada em dispositivo não reconhecido',
            suggestedActions: ['VERIFY_DEVICE', 'STEP_UP_AUTH']
          };
        }
        
        // Verificar se o dispositivo é considerado de risco
        if (knownDevice.riskScore > 70) {
          return {
            ruleId: 'DEVICE_ANOMALY_01',
            triggered: true,
            confidenceLevel: FraudConfidenceLevel.HIGH,
            score: 75,
            reason: `Dispositivo com pontuação de risco elevada: ${knownDevice.riskScore}`,
            suggestedActions: ['VERIFY_DEVICE', 'STEP_UP_AUTH', 'MANUAL_REVIEW']
          };
        }
        
        return {
          ruleId: 'DEVICE_ANOMALY_01',
          triggered: false,
          confidenceLevel: FraudConfidenceLevel.LOW,
          score: 0,
          reason: 'Dispositivo reconhecido e com baixo risco',
          suggestedActions: []
        };
      },
      applicableTransactionTypes: Object.values(FinancialTransactionType),
      requiredDataPoints: ['deviceId', 'userId'],
      version: '1.0.0',
      createdAt: new Date(),
      updatedAt: new Date(),
      cooldownPeriodMinutes: 1440 // 24 horas
    });
  }
  
  /**
   * Adiciona uma regra ao motor
   */
  public addRule(rule: FraudRule): void {
    this.rules.push(rule);
    
    this.logger.info({
      message: `Regra de detecção de fraude adicionada: ${rule.name}`,
      ruleId: rule.id,
      detectionType: rule.detectionType
    });
    
    // Registrar métrica
    this.metrics.gauge('fraud_detection.rules.count', this.rules.length, {
      detection_type: rule.detectionType
    });
  }
  
  /**
   * Remove uma regra do motor
   */
  public removeRule(ruleId: string): boolean {
    const initialLength = this.rules.length;
    this.rules = this.rules.filter(rule => rule.id !== ruleId);
    
    const removed = initialLength > this.rules.length;
    
    if (removed) {
      this.logger.info({
        message: `Regra de detecção de fraude removida: ${ruleId}`
      });
      
      // Atualizar métrica
      this.metrics.gauge('fraud_detection.rules.count', this.rules.length);
    }
    
    return removed;
  }
  
  /**
   * Avalia uma transação para detecção de fraude
   */
  public async detectFraud(
    transaction: TransactionRiskData
  ): Promise<FraudDetectionResult> {
    const startTime = Date.now();
    const span = this.tracer.startSpan('fraud_detection.evaluate');
    
    try {
      this.logger.debug({
        message: 'Iniciando detecção de fraude',
        transactionId: transaction.transactionId,
        userId: transaction.userId
      });
      
      // Obter contexto para avaliação
      const context = await this.getDetectionContext(transaction);
      
      // Filtrar regras aplicáveis
      const applicableRules = this.rules.filter(
        rule => rule.enabled && 
        rule.applicableTransactionTypes.includes(transaction.transactionType)
      );
      
      // Ordenar por prioridade (maior primeiro)
      applicableRules.sort((a, b) => b.priority - a.priority);
      
      // Avaliar cada regra
      const ruleResults: FraudRuleResult[] = [];
      const evaluatedRuleIds: string[] = [];
      
      for (const rule of applicableRules) {
        try {
          const result = await rule.evaluator(transaction, context);
          ruleResults.push(result);
          evaluatedRuleIds.push(rule.id);
          
          // Registrar métrica para regra avaliada
          this.metrics.increment('fraud_detection.rule.evaluated', {
            rule_id: rule.id,
            triggered: result.triggered ? 'true' : 'false'
          });
          
        } catch (error) {
          this.logger.error({
            message: `Erro ao avaliar regra de fraude ${rule.id}`,
            error: error.message,
            stack: error.stack,
            transactionId: transaction.transactionId
          });
        }
      }
      
      // Filtrar regras acionadas
      const triggeredRules = ruleResults.filter(result => result.triggered);
      
      // Calcular pontuação global
      let totalScore = 0;
      let maxConfidence = FraudConfidenceLevel.VERY_LOW;
      
      triggeredRules.forEach(rule => {
        totalScore = Math.max(totalScore, rule.score);
        
        // Determinar o nível de confiança mais alto
        const confidenceLevels = Object.values(FraudConfidenceLevel);
        const currentConfidenceIndex = confidenceLevels.indexOf(rule.confidenceLevel);
        const maxConfidenceIndex = confidenceLevels.indexOf(maxConfidence);
        
        if (currentConfidenceIndex > maxConfidenceIndex) {
          maxConfidence = rule.confidenceLevel;
        }
      });
      
      // Determinar tipos de fraude detectados
      const detectedFraudTypes = new Set<FraudDetectionType>();
      
      triggeredRules.forEach(rule => {
        const originalRule = applicableRules.find(r => r.id === rule.ruleId);
        if (originalRule) {
          detectedFraudTypes.add(originalRule.detectionType);
        }
      });
      
      // Coletar ações sugeridas
      const allSuggestedActions = new Set<string>();
      
      triggeredRules.forEach(rule => {
        rule.suggestedActions.forEach(action => allSuggestedActions.add(action));
      });
      
      // Determinar se requer revisão manual
      const requiresManualReview = Array.from(allSuggestedActions).includes('MANUAL_REVIEW');
      
      // Determinar nível de risco
      const riskLevel = this.mapScoreToRiskLevel(totalScore);
      
      // Construir resultado final
      const result: FraudDetectionResult = {
        transactionId: transaction.transactionId,
        timestamp: new Date(),
        fraudDetected: triggeredRules.length > 0,
        overallConfidenceLevel: maxConfidence,
        overallScore: totalScore,
        triggeredRules,
        evaluatedRules: evaluatedRuleIds,
        processingTimeMs: Date.now() - startTime,
        suggestedActions: Array.from(allSuggestedActions),
        requiresManualReview,
        fraudTypes: Array.from(detectedFraudTypes),
        riskLevel
      };
      
      // Registrar métricas
      this.metrics.gauge('fraud_detection.score', result.overallScore, {
        transaction_type: transaction.transactionType,
        tenant_id: transaction.tenantId,
        fraud_detected: result.fraudDetected ? 'true' : 'false'
      });
      
      this.metrics.histogram('fraud_detection.processing_time', result.processingTimeMs, {
        transaction_type: transaction.transactionType
      });
      
      // Enviar notificações se fraude foi detectada
      if (result.fraudDetected) {
        this.notifyFraudDetection(transaction, result);
      }
      
      this.logger.info({
        message: `Detecção de fraude concluída: ${result.fraudDetected ? 'Detectada' : 'Não detectada'}`,
        transactionId: transaction.transactionId,
        score: result.overallScore,
        confidence: result.overallConfidenceLevel,
        processingTime: result.processingTimeMs,
        triggeredRulesCount: triggeredRules.length
      });
      
      return result;
    } catch (error) {
      this.logger.error({
        message: 'Erro na detecção de fraude',
        error: error.message,
        stack: error.stack,
        transactionId: transaction.transactionId
      });
      
      throw error;
    } finally {
      span.end();
    }
  }
  
  /**
   * Obtém dados de contexto para avaliação de fraude
   */
  private async getDetectionContext(
    transaction: TransactionRiskData
  ): Promise<FraudDetectionContext> {
    try {
      const span = this.tracer.startSpan('fraud_detection.get_context');
      
      const context: FraudDetectionContext = {};
      
      // Obter transações recentes do usuário
      context.recentTransactions = await this.dataProvider.getRecentTransactions({
        userId: transaction.userId,
        tenantId: transaction.tenantId,
        limit: 20,
        hours: 24
      });
      
      // Obter dispositivos conhecidos do usuário
      context.userDevices = await this.dataProvider.getUserDevices({
        userId: transaction.userId,
        tenantId: transaction.tenantId
      });
      
      // Obter localizações conhecidas do usuário
      context.knownLocations = await this.dataProvider.getUserLocations({
        userId: transaction.userId,
        tenantId: transaction.tenantId
      });
      
      // Obter tentativas anteriores de fraude
      context.previousFraudAttempts = await this.dataProvider.getPreviousFraudAttempts({
        userId: transaction.userId,
        tenantId: transaction.tenantId,
        days: 90
      });
      
      // Calcular velocidade de transação
      context.transactionVelocity = await this.calculateTransactionVelocity(
        transaction, 
        context.recentTransactions || []
      );
      
      // Obter segmento do usuário e nível de risco
      const userSegmentInfo = await this.dataProvider.getUserSegmentInfo({
        userId: transaction.userId,
        tenantId: transaction.tenantId
      });
      
      context.userSegment = userSegmentInfo?.segment;
      context.riskTier = userSegmentInfo?.riskTier;
      
      span.end();
      
      return context;
    } catch (error) {
      this.logger.error({
        message: 'Erro ao obter contexto para detecção de fraude',
        error: error.message,
        stack: error.stack,
        transactionId: transaction.transactionId,
        userId: transaction.userId
      });
      
      // Retornar contexto vazio em caso de erro
      return {};
    }
  }
  
  /**
   * Calcula a velocidade de transação
   */
  private async calculateTransactionVelocity(
    transaction: TransactionRiskData,
    recentTransactions: Array<{
      timestamp: Date;
      amount: number;
      currency: string;
    }>
  ) {
    const now = transaction.timestamp || new Date();
    
    // Filtrar transações por período
    const lastHour = recentTransactions.filter(
      t => (now.getTime() - t.timestamp.getTime()) <= 3600000 // 1 hora em ms
    );
    
    const lastDay = recentTransactions.filter(
      t => (now.getTime() - t.timestamp.getTime()) <= 86400000 // 24 horas em ms
    );
    
    const lastWeek = recentTransactions.filter(
      t => (now.getTime() - t.timestamp.getTime()) <= 604800000 // 7 dias em ms
    );
    
    // Calcular somas de valores
    const hourlyAmount = lastHour.reduce(
      (sum, t) => sum + (t.currency === transaction.currency ? t.amount : 0), 
      0
    );
    
    const dailyAmount = lastDay.reduce(
      (sum, t) => sum + (t.currency === transaction.currency ? t.amount : 0), 
      0
    );
    
    const weeklyAmount = lastWeek.reduce(
      (sum, t) => sum + (t.currency === transaction.currency ? t.amount : 0), 
      0
    );
    
    return {
      hourly: lastHour.length,
      daily: lastDay.length,
      weekly: lastWeek.length,
      hourlyAmount,
      dailyAmount,
      weeklyAmount
    };
  }
  
  /**
   * Envia notificações sobre detecção de fraude
   */
  private async notifyFraudDetection(
    transaction: TransactionRiskData,
    detectionResult: FraudDetectionResult
  ): Promise<void> {
    try {
      // Enviar notificação para equipe de segurança
      if (detectionResult.overallScore > 70) {
        await this.notificationService.sendSecurityAlert({
          type: 'FRAUD_DETECTION',
          severity: detectionResult.overallScore > 90 ? 'HIGH' : 'MEDIUM',
          transactionId: transaction.transactionId,
          userId: transaction.userId,
          tenantId: transaction.tenantId,
          details: {
            score: detectionResult.overallScore,
            confidenceLevel: detectionResult.overallConfidenceLevel,
            fraudTypes: detectionResult.fraudTypes,
            triggeredRules: detectionResult.triggeredRules.map(r => ({
              id: r.ruleId,
              reason: r.reason
            }))
          }
        });
      }
      
      // Registrar no log de atividades suspeitas
      await this.dataProvider.recordFraudDetection({
        transactionId: transaction.transactionId,
        userId: transaction.userId,
        tenantId: transaction.tenantId,
        timestamp: new Date(),
        score: detectionResult.overallScore,
        confidenceLevel: detectionResult.overallConfidenceLevel,
        fraudTypes: detectionResult.fraudTypes,
        triggeredRules: detectionResult.triggeredRules
      });
    } catch (error) {
      this.logger.error({
        message: 'Erro ao enviar notificação de detecção de fraude',
        error: error.message,
        stack: error.stack,
        transactionId: transaction.transactionId
      });
    }
  }
  
  /**
   * Mapeia pontuação para nível de risco
   */
  private mapScoreToRiskLevel(score: number): RiskLevel {
    if (score >= 90) return RiskLevel.CRITICAL;
    if (score >= 80) return RiskLevel.VERY_HIGH;
    if (score >= 60) return RiskLevel.HIGH;
    if (score >= 40) return RiskLevel.MEDIUM;
    if (score >= 20) return RiskLevel.LOW;
    return RiskLevel.VERY_LOW;
  }
}