/**
 * Modelo de Avaliação de Risco para Transações Financeiras
 * 
 * Este modelo implementa os algoritmos e regras para avaliação de risco
 * de transações financeiras, integrando com o Bureau de Créditos e outras
 * fontes de dados para análise de risco.
 * 
 * @module RiskAssessmentModel
 */

import { Logger } from '../../../observability/logging/hook_logger';
import { Metrics } from '../../../observability/metrics/hook_metrics';
import { Tracer } from '../../../observability/tracing/hook_tracing';

/**
 * Tipos de transação financeira para avaliação de risco
 */
export enum FinancialTransactionType {
  PAYMENT = 'PAYMENT',
  TRANSFER = 'TRANSFER',
  WITHDRAWAL = 'WITHDRAWAL',
  DEPOSIT = 'DEPOSIT',
  LOAN_REQUEST = 'LOAN_REQUEST',
  CREDIT_CARD_PURCHASE = 'CREDIT_CARD_PURCHASE',
  MOBILE_MONEY_TRANSFER = 'MOBILE_MONEY_TRANSFER',
  EXCHANGE = 'EXCHANGE',
  INVESTMENT = 'INVESTMENT',
  RECURRING_PAYMENT = 'RECURRING_PAYMENT',
  BILL_PAYMENT = 'BILL_PAYMENT'
}

/**
 * Níveis de risco para classificação
 */
export enum RiskLevel {
  VERY_LOW = 'VERY_LOW',
  LOW = 'LOW',
  MEDIUM = 'MEDIUM',
  HIGH = 'HIGH',
  VERY_HIGH = 'VERY_HIGH',
  CRITICAL = 'CRITICAL'
}

/**
 * Categorias de fatores de risco
 */
export enum RiskFactorCategory {
  TRANSACTION_HISTORY = 'TRANSACTION_HISTORY',
  USER_PROFILE = 'USER_PROFILE',
  DEVICE_INFORMATION = 'DEVICE_INFORMATION',
  LOCATION = 'LOCATION',
  AMOUNT = 'AMOUNT',
  TIME_PATTERN = 'TIME_PATTERN',
  MERCHANT_CATEGORY = 'MERCHANT_CATEGORY',
  EXTERNAL_DATA = 'EXTERNAL_DATA',
  BEHAVIORAL = 'BEHAVIORAL',
  NETWORK = 'NETWORK',
  REGULATORY = 'REGULATORY'
}

/**
 * Ações recomendadas com base na avaliação de risco
 */
export enum RiskAction {
  APPROVE = 'APPROVE',
  REVIEW = 'REVIEW',
  ADDITIONAL_VERIFICATION = 'ADDITIONAL_VERIFICATION',
  STEP_UP_AUTH = 'STEP_UP_AUTH',
  BLOCK = 'BLOCK',
  REPORT = 'REPORT'
}

/**
 * Interface para a regra de avaliação de risco
 */
export interface RiskRule {
  id: string;
  name: string;
  description: string;
  category: RiskFactorCategory;
  weight: number;
  evaluator: (data: TransactionRiskData) => RiskEvaluation;
  enabled: boolean;
  applicableTransactionTypes: FinancialTransactionType[];
  requiredDataPoints: string[];
  version: string;
  createdAt: Date;
  updatedAt: Date;
  regulatoryRequirement?: string;
}

/**
 * Interface para o resultado da avaliação de risco
 */
export interface RiskEvaluation {
  score: number; // 0-100, onde 0 é risco mínimo e 100 é risco máximo
  level: RiskLevel;
  details: string;
  triggers: string[];
  confidence: number; // 0-100, confiança na avaliação
}

/**
 * Interface para dados da transação usados na avaliação de risco
 */
export interface TransactionRiskData {
  // Informações da transação
  transactionId: string;
  transactionType: FinancialTransactionType;
  amount: number;
  currency: string;
  timestamp: Date;
  channel: string;
  description?: string;
  
  // Informações do usuário
  userId: string;
  userCreatedAt: Date;
  userRiskProfile?: UserRiskProfile;
  
  // Informações KYC/AML
  kycLevel?: number;
  kycVerified?: boolean;
  amlScreeningResult?: any;
  
  // Informações geográficas
  countryCode?: string;
  ipAddress?: string;
  location?: {
    latitude: number;
    longitude: number;
    accuracy: number;
  };
  
  // Informações do dispositivo
  deviceId?: string;
  deviceFingerprint?: string;
  userAgent?: string;
  
  // Informações de comportamento
  userBehaviorProfile?: any;
  transactionVelocity?: any;
  
  // Dados externos
  bureauCreditScore?: number;
  externalDataSources?: Record<string, any>;
  
  // Metadados
  tenantId: string;
  metadata?: Record<string, any>;
}

/**
 * Interface para o perfil de risco do usuário
 */
export interface UserRiskProfile {
  userId: string;
  riskScore: number;
  riskLevel: RiskLevel;
  lastUpdated: Date;
  historicalScores: Array<{
    timestamp: Date;
    score: number;
    level: RiskLevel;
  }>;
  flags: string[];
  kycLevel: number;
  fraudAttempts: number;
  successfulTransactionsCount: number;
  failedTransactionsCount: number;
  averageTransactionAmount: number;
  maxTransactionAmount: number;
  totalTransactionsAmount: number;
  transactionFrequency: number;
  lastTransactionDate?: Date;
  creditScore?: number;
  trustScore?: number;
}

/**
 * Interface para o resultado detalhado da avaliação de risco
 */
export interface RiskAssessmentResult {
  transactionId: string;
  userId: string;
  tenantId: string;
  timestamp: Date;
  overallScore: number;
  riskLevel: RiskLevel;
  recommendedActions: RiskAction[];
  evaluations: Array<{
    ruleId: string;
    ruleName: string;
    category: RiskFactorCategory;
    score: number;
    details: string;
    triggered: boolean;
  }>;
  dataQuality: {
    completeness: number;
    reliability: number;
    missingFields: string[];
  };
  decisionTime: number; // tempo em ms para tomar a decisão
  requiresManualReview: boolean;
  thresholds: {
    low: number;
    medium: number;
    high: number;
    veryHigh: number;
    critical: number;
  };
  additionalInformation?: Record<string, any>;
}

/**
 * Classe para o modelo de avaliação de risco
 */
export class RiskAssessmentModel {
  private logger: Logger;
  private metrics: Metrics;
  private tracer: Tracer;
  private rules: RiskRule[] = [];
  private thresholds: {
    low: number;
    medium: number;
    high: number;
    veryHigh: number;
    critical: number;
  };
  
  /**
   * Construtor do modelo de avaliação de risco
   */
  constructor(
    logger: Logger,
    metrics: Metrics,
    tracer: Tracer,
    config: any
  ) {
    this.logger = logger;
    this.metrics = metrics;
    this.tracer = tracer;
    this.thresholds = config.riskThresholds || {
      low: 20,
      medium: 40,
      high: 60,
      veryHigh: 80,
      critical: 90
    };
    
    this.initDefaultRules();
  }
  
  /**
   * Inicializa as regras padrão de avaliação de risco
   */
  private initDefaultRules(): void {
    // As regras serão carregadas de configurações ou banco de dados
    // Para fins de demonstração, criaremos algumas regras básicas
    
    // Regra 1: Valor da transação acima do limite
    this.addRule({
      id: 'AMOUNT_THRESHOLD',
      name: 'Valor da Transação Elevado',
      description: 'Verifica se o valor da transação está acima do limite definido',
      category: RiskFactorCategory.AMOUNT,
      weight: 10,
      evaluator: (data: TransactionRiskData): RiskEvaluation => {
        const highAmountThreshold = data.userRiskProfile?.maxTransactionAmount || 5000;
        const veryHighAmountThreshold = highAmountThreshold * 2;
        
        if (data.amount > veryHighAmountThreshold) {
          return {
            score: 80,
            level: RiskLevel.HIGH,
            details: `Valor da transação (${data.amount} ${data.currency}) é significativamente maior que o padrão do usuário`,
            triggers: ['high_value_transaction'],
            confidence: 90
          };
        } else if (data.amount > highAmountThreshold) {
          return {
            score: 50,
            level: RiskLevel.MEDIUM,
            details: `Valor da transação (${data.amount} ${data.currency}) é maior que o padrão do usuário`,
            triggers: ['above_average_amount'],
            confidence: 80
          };
        }
        
        return {
          score: 10,
          level: RiskLevel.LOW,
          details: 'Valor da transação dentro dos limites normais',
          triggers: [],
          confidence: 90
        };
      },
      enabled: true,
      applicableTransactionTypes: Object.values(FinancialTransactionType),
      requiredDataPoints: ['amount', 'currency'],
      version: '1.0.0',
      createdAt: new Date(),
      updatedAt: new Date()
    });
    
    // Regra 2: Localização incomum
    this.addRule({
      id: 'UNUSUAL_LOCATION',
      name: 'Localização Incomum',
      description: 'Verifica se a transação está sendo realizada em uma localização incomum para o usuário',
      category: RiskFactorCategory.LOCATION,
      weight: 15,
      evaluator: (data: TransactionRiskData): RiskEvaluation => {
        // Implementação simplificada
        // Em produção, seria usado um histórico de localizações do usuário
        
        if (!data.countryCode || !data.ipAddress) {
          return {
            score: 30,
            level: RiskLevel.MEDIUM,
            details: 'Informações de localização incompletas',
            triggers: ['missing_location_data'],
            confidence: 50
          };
        }
        
        // Simulação de detecção de localização incomum
        const isUnusualLocation = data.metadata?.isUnusualLocation === true;
        
        if (isUnusualLocation) {
          return {
            score: 70,
            level: RiskLevel.HIGH,
            details: `Localização incomum detectada: ${data.countryCode}`,
            triggers: ['unusual_location'],
            confidence: 75
          };
        }
        
        return {
          score: 5,
          level: RiskLevel.VERY_LOW,
          details: 'Localização consistente com o histórico do usuário',
          triggers: [],
          confidence: 85
        };
      },
      enabled: true,
      applicableTransactionTypes: Object.values(FinancialTransactionType),
      requiredDataPoints: ['ipAddress'],
      version: '1.0.0',
      createdAt: new Date(),
      updatedAt: new Date()
    });
  }
  
  /**
   * Adiciona uma regra ao modelo
   */
  public addRule(rule: RiskRule): void {
    this.rules.push(rule);
    this.logger.info({
      message: `Regra de risco adicionada: ${rule.name}`,
      ruleId: rule.id,
      category: rule.category
    });
  }
  
  /**
   * Remove uma regra do modelo
   */
  public removeRule(ruleId: string): boolean {
    const initialLength = this.rules.length;
    this.rules = this.rules.filter(rule => rule.id !== ruleId);
    
    const removed = initialLength > this.rules.length;
    
    if (removed) {
      this.logger.info({
        message: `Regra de risco removida: ${ruleId}`
      });
    }
    
    return removed;
  }
  
  /**
   * Atualiza uma regra no modelo
   */
  public updateRule(ruleId: string, updatedRule: Partial<RiskRule>): boolean {
    const ruleIndex = this.rules.findIndex(rule => rule.id === ruleId);
    
    if (ruleIndex >= 0) {
      this.rules[ruleIndex] = {
        ...this.rules[ruleIndex],
        ...updatedRule,
        updatedAt: new Date()
      };
      
      this.logger.info({
        message: `Regra de risco atualizada: ${this.rules[ruleIndex].name}`,
        ruleId: ruleId
      });
      
      return true;
    }
    
    return false;
  }
  
  /**
   * Avalia o risco de uma transação financeira
   */
  public async evaluateRisk(data: TransactionRiskData): Promise<RiskAssessmentResult> {
    const startTime = Date.now();
    const span = this.tracer.startSpan('risk_assessment.evaluate');
    
    try {
      this.logger.debug({
        message: 'Iniciando avaliação de risco',
        transactionId: data.transactionId,
        userId: data.userId,
        transactionType: data.transactionType
      });
      
      // Verificar campos obrigatórios
      const missingFields = this.validateRequiredFields(data);
      
      // Calcular qualidade dos dados
      const dataQuality = this.calculateDataQuality(data, missingFields);
      
      // Filtrar regras aplicáveis ao tipo de transação
      const applicableRules = this.rules.filter(
        rule => rule.enabled && 
        rule.applicableTransactionTypes.includes(data.transactionType)
      );
      
      // Avaliar cada regra
      const ruleEvaluations = await Promise.all(
        applicableRules.map(async rule => {
          try {
            const evaluation = rule.evaluator(data);
            
            return {
              ruleId: rule.id,
              ruleName: rule.name,
              category: rule.category,
              score: evaluation.score * (rule.weight / 10),
              details: evaluation.details,
              triggered: evaluation.triggers.length > 0
            };
          } catch (error) {
            this.logger.error({
              message: `Erro ao avaliar regra ${rule.id}`,
              error: error.message,
              stack: error.stack,
              transactionId: data.transactionId
            });
            
            return {
              ruleId: rule.id,
              ruleName: rule.name,
              category: rule.category,
              score: 0,
              details: 'Erro ao avaliar regra',
              triggered: false
            };
          }
        })
      );
      
      // Calcular pontuação global
      let totalWeight = 0;
      let weightedScore = 0;
      
      ruleEvaluations.forEach((evaluation, index) => {
        const rule = applicableRules[index];
        weightedScore += evaluation.score * rule.weight;
        totalWeight += rule.weight;
      });
      
      // Normalizar pontuação para 0-100
      const overallScore = totalWeight > 0 
        ? Math.min(100, Math.max(0, weightedScore / totalWeight * 10))
        : 0;
      
      // Determinar nível de risco
      const riskLevel = this.determineRiskLevel(overallScore);
      
      // Determinar ações recomendadas
      const recommendedActions = this.determineRecommendedActions(riskLevel, overallScore, data);
      
      // Construir resultado
      const result: RiskAssessmentResult = {
        transactionId: data.transactionId,
        userId: data.userId,
        tenantId: data.tenantId,
        timestamp: new Date(),
        overallScore: Math.round(overallScore * 10) / 10, // Arredondar para 1 casa decimal
        riskLevel,
        recommendedActions,
        evaluations: ruleEvaluations,
        dataQuality,
        decisionTime: Date.now() - startTime,
        requiresManualReview: recommendedActions.includes(RiskAction.REVIEW),
        thresholds: this.thresholds
      };
      
      // Registrar métricas
      this.metrics.gauge('risk_assessment.score', result.overallScore, {
        transaction_type: data.transactionType,
        tenant_id: data.tenantId,
        risk_level: result.riskLevel
      });
      
      this.metrics.histogram('risk_assessment.decision_time', result.decisionTime, {
        transaction_type: data.transactionType
      });
      
      this.logger.info({
        message: 'Avaliação de risco concluída',
        transactionId: data.transactionId,
        score: result.overallScore,
        riskLevel: result.riskLevel,
        decisionTime: result.decisionTime
      });
      
      return result;
    } catch (error) {
      this.logger.error({
        message: 'Erro na avaliação de risco',
        error: error.message,
        stack: error.stack,
        transactionId: data.transactionId
      });
      
      throw error;
    } finally {
      span.end();
    }
  }
  
  /**
   * Determina o nível de risco com base na pontuação
   */
  private determineRiskLevel(score: number): RiskLevel {
    if (score >= this.thresholds.critical) return RiskLevel.CRITICAL;
    if (score >= this.thresholds.veryHigh) return RiskLevel.VERY_HIGH;
    if (score >= this.thresholds.high) return RiskLevel.HIGH;
    if (score >= this.thresholds.medium) return RiskLevel.MEDIUM;
    if (score >= this.thresholds.low) return RiskLevel.LOW;
    return RiskLevel.VERY_LOW;
  }
  
  /**
   * Determina as ações recomendadas com base no nível de risco
   */
  private determineRecommendedActions(
    riskLevel: RiskLevel, 
    score: number, 
    data: TransactionRiskData
  ): RiskAction[] {
    const actions: RiskAction[] = [];
    
    switch (riskLevel) {
      case RiskLevel.CRITICAL:
        actions.push(RiskAction.BLOCK, RiskAction.REPORT);
        break;
      case RiskLevel.VERY_HIGH:
        actions.push(RiskAction.STEP_UP_AUTH, RiskAction.REVIEW);
        break;
      case RiskLevel.HIGH:
        actions.push(RiskAction.ADDITIONAL_VERIFICATION);
        break;
      case RiskLevel.MEDIUM:
        // Para transações de alto valor, solicitar verificação adicional mesmo com risco médio
        if (data.amount > 10000) {
          actions.push(RiskAction.ADDITIONAL_VERIFICATION);
        } else {
          actions.push(RiskAction.APPROVE);
        }
        break;
      default:
        actions.push(RiskAction.APPROVE);
    }
    
    return actions;
  }
  
  /**
   * Valida os campos obrigatórios para avaliação
   */
  private validateRequiredFields(data: TransactionRiskData): string[] {
    const requiredFields = [
      'transactionId', 
      'userId', 
      'transactionType', 
      'amount', 
      'currency', 
      'timestamp', 
      'tenantId'
    ];
    
    const missingFields: string[] = [];
    
    requiredFields.forEach(field => {
      if (data[field] === undefined || data[field] === null) {
        missingFields.push(field);
      }
    });
    
    return missingFields;
  }
  
  /**
   * Calcula a qualidade dos dados para a avaliação
   */
  private calculateDataQuality(
    data: TransactionRiskData, 
    missingRequiredFields: string[]
  ): { completeness: number; reliability: number; missingFields: string[] } {
    // Lista de todos os possíveis campos para uma avaliação ideal
    const allPossibleFields = [
      'transactionId', 'transactionType', 'amount', 'currency', 'timestamp',
      'channel', 'description', 'userId', 'userCreatedAt', 'userRiskProfile',
      'kycLevel', 'kycVerified', 'amlScreeningResult', 'countryCode', 'ipAddress',
      'location', 'deviceId', 'deviceFingerprint', 'userAgent', 'userBehaviorProfile',
      'transactionVelocity', 'bureauCreditScore', 'externalDataSources'
    ];
    
    // Campos que contêm dados
    const presentFields = allPossibleFields.filter(
      field => data[field] !== undefined && data[field] !== null
    );
    
    // Calcular completude (porcentagem de campos preenchidos)
    const completeness = (presentFields.length / allPossibleFields.length) * 100;
    
    // Calcular confiabilidade
    let reliability = 100;
    
    // Reduzir confiabilidade se campos obrigatórios estiverem faltando
    if (missingRequiredFields.length > 0) {
      reliability -= missingRequiredFields.length * 20;
    }
    
    // Reduzir confiabilidade se dados importantes estiverem faltando
    const importantFields = ['userRiskProfile', 'kycVerified', 'ipAddress', 'deviceId'];
    const missingImportantFields = importantFields.filter(
      field => data[field] === undefined || data[field] === null
    );
    
    if (missingImportantFields.length > 0) {
      reliability -= missingImportantFields.length * 10;
    }
    
    // Garantir que a confiabilidade esteja entre 0 e 100
    reliability = Math.max(0, Math.min(100, reliability));
    
    // Todos os campos ausentes
    const allMissingFields = allPossibleFields.filter(
      field => data[field] === undefined || data[field] === null
    );
    
    return {
      completeness,
      reliability,
      missingFields: allMissingFields
    };
  }
}