/**
 * ⚠️ RISK ASSESSMENT DECORATOR - INNOVABIZ IAM
 * Decorator para avaliação de risco em tempo real com IA/ML
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: Basel III, NIST Risk Management Framework, ISO 31000
 * Mercados: Angola (BNA), Europa, América, China, BRICS, Brasil
 * IA/ML: Modelos preditivos, Análise comportamental, Detecção de anomalias
 */

import { 
  createMethodDecorator, 
  SetMetadata, 
  applyDecorators,
  UseInterceptors,
  ExecutionContext,
  Injectable,
  NestInterceptor,
  CallHandler,
  Logger,
  Inject
} from '@nestjs/common';
import { Observable, throwError } from 'rxjs';
import { tap, catchError } from 'rxjs/operators';
import { Reflector } from '@nestjs/core';
import { ConfigService } from '@nestjs/config';
import { Cache } from 'cache-manager';
import { CACHE_MANAGER } from '@nestjs/cache-manager';
import { Request } from 'express';

/**
 * Configurações de avaliação de risco
 */
export interface RiskAssessmentConfig {
  level: 'basic' | 'standard' | 'advanced' | 'ml_enhanced' | 'real_time';
  
  thresholds: {
    low: number;
    medium: number;
    high: number;
    critical: number;
  };
  
  actions: {
    onLowRisk?: RiskAction[];
    onMediumRisk?: RiskAction[];
    onHighRisk?: RiskAction[];
    onCriticalRisk?: RiskAction[];
  };
  
  factors: {
    behavioralAnalysis: boolean;
    deviceFingerprinting: boolean;
    locationAnalysis: boolean;
    timePatternAnalysis: boolean;
    ipReputationCheck: boolean;
    userAgentAnalysis: boolean;
    requestPatternAnalysis: boolean;
    velocityChecks: boolean;
    transactionAmount?: boolean;
    accountAge: boolean;
    previousIncidents: boolean;
    complianceHistory: boolean;
    jurisdictionRisk: boolean;
    sanctionListCheck: boolean;
    pepsCheck: boolean;
    amlScreening: boolean;
  };
  
  mlConfig?: {
    modelVersion: string;
    features: string[];
    confidenceThreshold: number;
    retrainInterval: number;
    anomalyDetection: boolean;
    ensembleModels: boolean;
  };
  
  cacheConfig?: {
    enabled: boolean;
    ttl: number;
    keyStrategy: 'user' | 'session' | 'ip' | 'composite';
  };
  
  auditConfig?: {
    logAllAssessments: boolean;
    logHighRiskOnly: boolean;
    includeFeatures: boolean;
    includeModelOutput: boolean;
  };
  
  jurisdictionConfig?: {
    [key: string]: Partial<RiskAssessmentConfig>;
  };
}

export type RiskAction = 
  | 'allow'
  | 'log'
  | 'alert'
  | 'require_mfa'
  | 'require_additional_auth'
  | 'rate_limit'
  | 'block_temporarily'
  | 'block_permanently'
  | 'escalate_to_human'
  | 'trigger_investigation'
  | 'notify_compliance'
  | 'freeze_account'
  | 'require_kyc_update';

export interface RiskAssessmentResult {
  riskScore: number;
  riskLevel: 'low' | 'medium' | 'high' | 'critical';
  
  riskFactors: {
    factor: string;
    weight: number;
    score: number;
    description: string;
  }[];
  
  recommendedActions: RiskAction[];
  confidence?: number;
  modelVersion?: string;
  
  assessmentId: string;
  timestamp: Date;
  processingTime: number;
  
  metadata: {
    userId?: string;
    tenantId?: string;
    sessionId?: string;
    ipAddress: string;
    userAgent: string;
    endpoint: string;
    method: string;
    jurisdiction: string;
    deviceFingerprint?: string;
    location?: {
      country: string;
      region: string;
      city?: string;
      coordinates?: [number, number];
    };
    previousAssessments?: number;
    recentIncidents?: number;
    accountAge?: number;
    sanctionListMatch?: boolean;
    pepsMatch?: boolean;
    amlFlags?: string[];
  };
  
  explanation?: {
    primaryReasons: string[];
    contributingFactors: string[];
    mitigatingFactors: string[];
    recommendations: string[];
  };
}

/**
 * Configurações padrão por tipo de operação
 */
const DEFAULT_CONFIGS: Record<string, RiskAssessmentConfig> = {
  authentication: {
    level: 'advanced',
    thresholds: { low: 0.2, medium: 0.4, high: 0.7, critical: 0.9 },
    actions: {
      onMediumRisk: ['log', 'require_mfa'],
      onHighRisk: ['log', 'require_additional_auth', 'alert'],
      onCriticalRisk: ['block_temporarily', 'trigger_investigation', 'notify_compliance']
    },
    factors: {
      behavioralAnalysis: true,
      deviceFingerprinting: true,
      locationAnalysis: true,
      timePatternAnalysis: true,
      ipReputationCheck: true,
      userAgentAnalysis: true,
      requestPatternAnalysis: true,
      velocityChecks: true,
      accountAge: true,
      previousIncidents: true,
      complianceHistory: true,
      jurisdictionRisk: true,
      sanctionListCheck: false,
      pepsCheck: false,
      amlScreening: false
    }
  },
  
  financial: {
    level: 'ml_enhanced',
    thresholds: { low: 0.1, medium: 0.3, high: 0.6, critical: 0.8 },
    actions: {
      onMediumRisk: ['log', 'require_mfa', 'alert'],
      onHighRisk: ['log', 'require_additional_auth', 'escalate_to_human'],
      onCriticalRisk: ['block_temporarily', 'trigger_investigation', 'notify_compliance', 'freeze_account']
    },
    factors: {
      behavioralAnalysis: true,
      deviceFingerprinting: true,
      locationAnalysis: true,
      timePatternAnalysis: true,
      ipReputationCheck: true,
      userAgentAnalysis: true,
      requestPatternAnalysis: true,
      velocityChecks: true,
      transactionAmount: true,
      accountAge: true,
      previousIncidents: true,
      complianceHistory: true,
      jurisdictionRisk: true,
      sanctionListCheck: true,
      pepsCheck: true,
      amlScreening: true
    },
    mlConfig: {
      modelVersion: 'financial_v2.1',
      features: ['amount', 'frequency', 'location', 'time', 'counterparty'],
      confidenceThreshold: 0.85,
      retrainInterval: 24,
      anomalyDetection: true,
      ensembleModels: true
    }
  },
  
  administrative: {
    level: 'standard',
    thresholds: { low: 0.3, medium: 0.5, high: 0.7, critical: 0.9 },
    actions: {
      onMediumRisk: ['log'],
      onHighRisk: ['log', 'alert'],
      onCriticalRisk: ['log', 'alert', 'require_additional_auth']
    },
    factors: {
      behavioralAnalysis: true,
      deviceFingerprinting: false,
      locationAnalysis: true,
      timePatternAnalysis: true,
      ipReputationCheck: true,
      userAgentAnalysis: false,
      requestPatternAnalysis: true,
      velocityChecks: true,
      accountAge: true,
      previousIncidents: true,
      complianceHistory: false,
      jurisdictionRisk: false,
      sanctionListCheck: false,
      pepsCheck: false,
      amlScreening: false
    }
  }
};

/**
 * Decorator para avaliação de risco automática
 */
export function RiskAssessment(
  configOrType: string | RiskAssessmentConfig = 'authentication'
): MethodDecorator {
  const config = typeof configOrType === 'string' 
    ? DEFAULT_CONFIGS[configOrType] || DEFAULT_CONFIGS.authentication
    : configOrType;

  return applyDecorators(
    SetMetadata('riskAssessment', config),
    UseInterceptors(RiskAssessmentInterceptor)
  );
}

/**
 * Interceptor para executar avaliação de risco
 */
@Injectable()
export class RiskAssessmentInterceptor implements NestInterceptor {
  private readonly logger = new Logger(RiskAssessmentInterceptor.name);

  constructor(
    private readonly reflector: Reflector,
    private readonly configService: ConfigService,
    @Inject(CACHE_MANAGER) private readonly cacheManager: Cache
  ) {}

  async intercept(context: ExecutionContext, next: CallHandler): Promise<Observable<any>> {
    const riskConfig = this.reflector.get<RiskAssessmentConfig>('riskAssessment', context.getHandler());
    
    if (!riskConfig) {
      return next.handle();
    }

    const request = context.switchToHttp().getRequest<Request>();

    try {
      const riskResult = await this.assessRisk(request, riskConfig, context);
      (request as any).riskAssessment = riskResult;
      
      await this.applyRiskActions(riskResult, riskConfig, request);
      this.logRiskAssessment(riskResult, riskConfig);
      
      return next.handle().pipe(
        tap(() => this.updateSuccessMetrics(riskResult)),
        catchError((error) => {
          this.updateErrorMetrics(riskResult, error);
          return throwError(error);
        })
      );

    } catch (error) {
      this.logger.error(`Risk assessment failed: ${error.message}`, error.stack);
      this.logger.warn('Proceeding with operation due to risk assessment failure');
      return next.handle();
    }
  }

  private async assessRisk(
    request: Request,
    config: RiskAssessmentConfig,
    context: ExecutionContext
  ): Promise<RiskAssessmentResult> {
    const startTime = Date.now();
    const assessmentId = this.generateAssessmentId();
    
    const requestContext = this.extractRequestContext(request, context);
    const adjustedConfig = this.adjustConfigForJurisdiction(config, requestContext.jurisdiction);
    
    if (adjustedConfig.cacheConfig?.enabled) {
      const cached = await this.getCachedAssessment(requestContext, adjustedConfig);
      if (cached) {
        this.logger.debug(`Using cached risk assessment: ${cached.assessmentId}`);
        return cached;
      }
    }
    
    const riskFactors = await this.calculateRiskFactors(requestContext, adjustedConfig);
    const riskScore = this.calculateFinalRiskScore(riskFactors, adjustedConfig);
    const riskLevel = this.determineRiskLevel(riskScore, adjustedConfig.thresholds);
    const recommendedActions = this.determineRecommendedActions(riskLevel, adjustedConfig);
    const explanation = this.generateExplanation(riskFactors, riskScore, riskLevel);
    
    const result: RiskAssessmentResult = {
      riskScore,
      riskLevel,
      riskFactors,
      recommendedActions,
      assessmentId,
      timestamp: new Date(),
      processingTime: Date.now() - startTime,
      metadata: requestContext,
      explanation
    };
    
    if (adjustedConfig.cacheConfig?.enabled) {
      await this.setCachedAssessment(requestContext, result, adjustedConfig);
    }
    
    return result;
  }

  private extractRequestContext(request: Request, context: ExecutionContext): RiskAssessmentResult['metadata'] {
    const user = (request as any).user;
    const endpoint = `${request.method} ${request.path}`;
    
    return {
      userId: user?.id,
      tenantId: user?.tenantId,
      sessionId: user?.sessionId || request.get('X-Session-Id'),
      ipAddress: this.getClientIP(request),
      userAgent: request.get('User-Agent') || 'unknown',
      endpoint,
      method: request.method,
      jurisdiction: user?.jurisdiction || 'global',
      deviceFingerprint: request.get('X-Device-Fingerprint'),
      location: this.extractLocation(request),
      accountAge: user?.accountAge,
      previousAssessments: 0,
      recentIncidents: 0
    };
  }

  private adjustConfigForJurisdiction(
    config: RiskAssessmentConfig,
    jurisdiction: string
  ): RiskAssessmentConfig {
    if (!config.jurisdictionConfig || !config.jurisdictionConfig[jurisdiction]) {
      return config;
    }
    
    const jurisdictionOverrides = config.jurisdictionConfig[jurisdiction];
    return {
      ...config,
      ...jurisdictionOverrides,
      thresholds: { ...config.thresholds, ...jurisdictionOverrides.thresholds },
      actions: { ...config.actions, ...jurisdictionOverrides.actions },
      factors: { ...config.factors, ...jurisdictionOverrides.factors }
    };
  }

  private async calculateRiskFactors(
    context: RiskAssessmentResult['metadata'],
    config: RiskAssessmentConfig
  ): Promise<RiskAssessmentResult['riskFactors']> {
    const factors: RiskAssessmentResult['riskFactors'] = [];
    
    if (config.factors.behavioralAnalysis) {
      const behaviorScore = await this.analyzeBehavior(context);
      factors.push({
        factor: 'behavioral_analysis',
        weight: 0.25,
        score: behaviorScore,
        description: 'Análise de padrões comportamentais do usuário'
      });
    }
    
    if (config.factors.locationAnalysis) {
      const locationScore = await this.analyzeLocation(context);
      factors.push({
        factor: 'location_analysis',
        weight: 0.20,
        score: locationScore,
        description: 'Análise de localização e geolocalização'
      });
    }
    
    if (config.factors.ipReputationCheck) {
      const ipScore = await this.analyzeIPReputation(context.ipAddress);
      factors.push({
        factor: 'ip_reputation',
        weight: 0.15,
        score: ipScore,
        description: 'Verificação de reputação do endereço IP'
      });
    }
    
    if (config.factors.velocityChecks) {
      const velocityScore = await this.analyzeVelocity(context);
      factors.push({
        factor: 'velocity_check',
        weight: 0.15,
        score: velocityScore,
        description: 'Análise de frequência e velocidade de requisições'
      });
    }
    
    if (config.factors.sanctionListCheck && context.userId) {
      const sanctionScore = await this.checkSanctionLists(context.userId);
      factors.push({
        factor: 'sanction_check',
        weight: 0.30,
        score: sanctionScore,
        description: 'Verificação em listas de sanções internacionais'
      });
    }
    
    return factors;
  }

  private calculateFinalRiskScore(
    factors: RiskAssessmentResult['riskFactors'],
    config: RiskAssessmentConfig
  ): number {
    if (factors.length === 0) return 0;
    
    const totalWeight = factors.reduce((sum, factor) => sum + factor.weight, 0);
    const weightedScore = factors.reduce((sum, factor) => sum + (factor.score * factor.weight), 0);
    
    return Math.min(1, Math.max(0, weightedScore / totalWeight));
  }

  private determineRiskLevel(
    score: number,
    thresholds: RiskAssessmentConfig['thresholds']
  ): RiskAssessmentResult['riskLevel'] {
    if (score >= thresholds.critical) return 'critical';
    if (score >= thresholds.high) return 'high';
    if (score >= thresholds.medium) return 'medium';
    return 'low';
  }

  private determineRecommendedActions(
    riskLevel: RiskAssessmentResult['riskLevel'],
    config: RiskAssessmentConfig
  ): RiskAction[] {
    const actionKey = `on${riskLevel.charAt(0).toUpperCase() + riskLevel.slice(1)}Risk` as keyof RiskAssessmentConfig['actions'];
    return config.actions[actionKey] || ['allow'];
  }

  private generateExplanation(
    factors: RiskAssessmentResult['riskFactors'],
    score: number,
    level: RiskAssessmentResult['riskLevel']
  ): RiskAssessmentResult['explanation'] {
    const sortedFactors = factors.sort((a, b) => (b.score * b.weight) - (a.score * a.weight));
    
    return {
      primaryReasons: sortedFactors.slice(0, 3).map(f => f.description),
      contributingFactors: sortedFactors.slice(3).map(f => f.description),
      mitigatingFactors: factors.filter(f => f.score < 0.3).map(f => f.description),
      recommendations: this.generateRecommendations(level, factors)
    };
  }

  private async applyRiskActions(
    result: RiskAssessmentResult,
    config: RiskAssessmentConfig,
    request: Request
  ): Promise<void> {
    for (const action of result.recommendedActions) {
      await this.executeRiskAction(action, result, request);
    }
  }

  private async executeRiskAction(
    action: RiskAction,
    result: RiskAssessmentResult,
    request: Request
  ): Promise<void> {
    switch (action) {
      case 'block_temporarily':
        throw new Error(`Acesso temporariamente bloqueado devido ao alto risco (${result.riskScore.toFixed(2)})`);
      
      case 'block_permanently':
        throw new Error(`Acesso permanentemente bloqueado devido ao risco crítico`);
      
      case 'require_mfa':
        (request as any).requireMFA = true;
        break;
      
      case 'require_additional_auth':
        (request as any).requireAdditionalAuth = true;
        break;
      
      case 'alert':
        await this.sendRiskAlert(result);
        break;
      
      case 'notify_compliance':
        await this.notifyCompliance(result);
        break;
      
      case 'trigger_investigation':
        await this.triggerInvestigation(result);
        break;
      
      default:
        this.logger.debug(`Risk action not implemented: ${action}`);
    }
  }

  // Métodos auxiliares simplificados
  private async analyzeBehavior(context: RiskAssessmentResult['metadata']): Promise<number> {
    return Math.random() * 0.5; // Placeholder - implementar análise real
  }

  private async analyzeLocation(context: RiskAssessmentResult['metadata']): Promise<number> {
    return Math.random() * 0.3; // Placeholder
  }

  private async analyzeIPReputation(ip: string): Promise<number> {
    return Math.random() * 0.4; // Placeholder
  }

  private async analyzeVelocity(context: RiskAssessmentResult['metadata']): Promise<number> {
    return Math.random() * 0.6; // Placeholder
  }

  private async checkSanctionLists(userId: string): Promise<number> {
    return Math.random() * 0.1; // Placeholder
  }

  private generateRecommendations(
    level: RiskAssessmentResult['riskLevel'],
    factors: RiskAssessmentResult['riskFactors']
  ): string[] {
    const recommendations: string[] = [];
    
    switch (level) {
      case 'critical':
        recommendations.push('Bloquear acesso imediatamente');
        recommendations.push('Iniciar investigação de segurança');
        break;
      case 'high':
        recommendations.push('Requerer autenticação adicional');
        recommendations.push('Monitorar atividade de perto');
        break;
      case 'medium':
        recommendations.push('Requerer MFA');
        break;
      case 'low':
        recommendations.push('Continuar monitoramento normal');
        break;
    }
    
    return recommendations;
  }

  private async sendRiskAlert(result: RiskAssessmentResult): Promise<void> {
    this.logger.warn(`Risk alert: ${result.riskLevel} risk detected`, result);
  }

  private async notifyCompliance(result: RiskAssessmentResult): Promise<void> {
    this.logger.error(`Compliance notification: Critical risk detected`, result);
  }

  private async triggerInvestigation(result: RiskAssessmentResult): Promise<void> {
    this.logger.error(`Investigation triggered for assessment: ${result.assessmentId}`, result);
  }

  private logRiskAssessment(result: RiskAssessmentResult, config: RiskAssessmentConfig): void {
    if (config.auditConfig?.logAllAssessments || 
        (config.auditConfig?.logHighRiskOnly && ['high', 'critical'].includes(result.riskLevel))) {
      this.logger.log(`Risk assessment completed`, {
        assessmentId: result.assessmentId,
        riskScore: result.riskScore,
        riskLevel: result.riskLevel,
        processingTime: result.processingTime
      });
    }
  }

  private updateSuccessMetrics(result: RiskAssessmentResult): void {
    // Implementar métricas de sucesso
  }

  private updateErrorMetrics(result: RiskAssessmentResult, error: any): void {
    // Implementar métricas de erro
  }

  private generateAssessmentId(): string {
    return `risk_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  private extractLocation(request: Request): RiskAssessmentResult['metadata']['location'] {
    // Implementar extração de localização
    return {
      country: 'Unknown',
      region: 'Unknown'
    };
  }

  private async getCachedAssessment(
    context: RiskAssessmentResult['metadata'],
    config: RiskAssessmentConfig
  ): Promise<RiskAssessmentResult | null> {
    // Implementar cache
    return null;
  }

  private async setCachedAssessment(
    context: RiskAssessmentResult['metadata'],
    result: RiskAssessmentResult,
    config: RiskAssessmentConfig
  ): Promise<void> {
    // Implementar cache
  }

  private getClientIP(request: Request): string {
    const forwarded = request.get('X-Forwarded-For');
    const realIP = request.get('X-Real-IP');
    const cfConnectingIP = request.get('CF-Connecting-IP');
    
    if (cfConnectingIP) return cfConnectingIP;
    if (realIP) return realIP;
    if (forwarded) return forwarded.split(',')[0].trim();
    
    return request.ip || request.connection.remoteAddress || 'unknown';
  }
}