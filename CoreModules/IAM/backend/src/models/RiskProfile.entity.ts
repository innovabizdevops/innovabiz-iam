/**
 * ⚠️ RISK PROFILE ENTITY - INNOVABIZ IAM
 * Entidade de perfil de risco com ML e análise comportamental
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: NIST AI RMF, ISO/IEC 42001, Basel III, COSO ERM
 * Standards: NIST Cybersecurity Framework, ISO 31000, PCI DSS 4.0
 */

import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
  ManyToOne,
  OneToMany,
  JoinColumn,
  Index,
  BeforeInsert,
  BeforeUpdate
} from 'typeorm';
import { IsUUID, IsString, IsEnum, IsOptional, IsNumber, IsArray, IsDateString, Min, Max } from 'class-validator';

import { User } from './User.entity';
import { RiskEvent } from './RiskEvent.entity';

export enum RiskLevel {
  VERY_LOW = 'very_low',
  LOW = 'low',
  MEDIUM = 'medium',
  HIGH = 'high',
  VERY_HIGH = 'very_high',
  CRITICAL = 'critical'
}

export enum RiskCategory {
  AUTHENTICATION = 'authentication',
  DEVICE = 'device',
  LOCATION = 'location',
  BEHAVIORAL = 'behavioral',
  TEMPORAL = 'temporal',
  VELOCITY = 'velocity',
  ANOMALY = 'anomaly'
}

export enum RiskTrend {
  DECREASING = 'decreasing',
  STABLE = 'stable',
  INCREASING = 'increasing',
  VOLATILE = 'volatile'
}

@Entity('risk_profiles')
@Index(['userId', 'tenantId'], { unique: true })
@Index(['currentRiskScore'])
@Index(['riskLevel'])
@Index(['lastAssessmentAt'])
@Index(['isActive'])
export class RiskProfile {
  @PrimaryGeneratedColumn('uuid')
  @IsUUID()
  id: string;

  @Column({ type: 'uuid', name: 'user_id' })
  @IsUUID()
  @Index()
  userId: string;

  @Column({ type: 'uuid', name: 'tenant_id' })
  @IsUUID()
  @Index()
  tenantId: string;

  @Column({ type: 'float', default: 0, name: 'baseline_risk_score' })
  @IsNumber()
  @Min(0)
  @Max(100)
  baselineRiskScore: number;

  @Column({ type: 'float', default: 0, name: 'current_risk_score' })
  @IsNumber()
  @Min(0)
  @Max(100)
  currentRiskScore: number;

  @Column({ type: 'float', default: 0, name: 'peak_risk_score' })
  @IsNumber()
  @Min(0)
  @Max(100)
  peakRiskScore: number;

  @Column({ 
    type: 'enum', 
    enum: RiskLevel, 
    default: RiskLevel.LOW,
    name: 'risk_level'
  })
  @IsEnum(RiskLevel)
  riskLevel: RiskLevel;

  @Column({ 
    type: 'enum', 
    enum: RiskTrend, 
    default: RiskTrend.STABLE,
    name: 'risk_trend'
  })
  @IsEnum(RiskTrend)
  riskTrend: RiskTrend;

  @Column({ type: 'float', default: 0, name: 'confidence_score' })
  @IsNumber()
  @Min(0)
  @Max(1)
  confidenceScore: number;

  @Column({ type: 'simple-array', nullable: true, name: 'device_fingerprints' })
  @IsOptional()
  @IsArray()
  deviceFingerprints?: string[];

  @Column({ type: 'simple-array', nullable: true, name: 'trusted_locations' })
  @IsOptional()
  @IsArray()
  trustedLocations?: string[];

  @Column({ type: 'simple-array', nullable: true, name: 'suspicious_ips' })
  @IsOptional()
  @IsArray()
  suspiciousIps?: string[];

  @Column({ type: 'jsonb', nullable: true, name: 'behavior_patterns' })
  behaviorPatterns?: {
    averageSessionDuration?: number;
    typicalLoginHours?: number[];
    commonUserAgents?: string[];
    frequentLocations?: string[];
    devicePreferences?: string[];
    activityPatterns?: Record<string, number>;
  };

  @Column({ type: 'jsonb', nullable: true, name: 'risk_factors' })
  riskFactors?: {
    deviceRisk?: number;
    locationRisk?: number;
    behavioralRisk?: number;
    temporalRisk?: number;
    velocityRisk?: number;
    anomalyRisk?: number;
  };

  @Column({ type: 'jsonb', nullable: true, name: 'ml_features' })
  mlFeatures?: {
    loginFrequency?: number;
    sessionDuration?: number;
    locationVariability?: number;
    deviceConsistency?: number;
    timePatternScore?: number;
    velocityScore?: number;
    anomalyScore?: number;
  };

  @Column({ type: 'jsonb', nullable: true, name: 'threat_indicators' })
  threatIndicators?: {
    bruteForceAttempts?: number;
    suspiciousLocations?: number;
    unknownDevices?: number;
    velocityViolations?: number;
    anomalousPatterns?: number;
    compromiseIndicators?: number;
  };

  @Column({ type: 'int', default: 0, name: 'assessment_count' })
  @IsNumber()
  assessmentCount: number;

  @Column({ type: 'int', default: 0, name: 'high_risk_events' })
  @IsNumber()
  highRiskEvents: number;

  @Column({ type: 'int', default: 0, name: 'security_violations' })
  @IsNumber()
  securityViolations: number;

  @Column({ type: 'boolean', default: true, name: 'is_active' })
  isActive: boolean;

  @Column({ type: 'boolean', default: false, name: 'requires_monitoring' })
  requiresMonitoring: boolean;

  @Column({ type: 'boolean', default: false, name: 'is_flagged' })
  isFlagged: boolean;

  @Column({ type: 'timestamp', nullable: true, name: 'last_assessment_at' })
  @IsOptional()
  @IsDateString()
  lastAssessmentAt?: Date;

  @Column({ type: 'timestamp', nullable: true, name: 'last_high_risk_at' })
  @IsOptional()
  @IsDateString()
  lastHighRiskAt?: Date;

  @Column({ type: 'timestamp', nullable: true, name: 'flagged_at' })
  @IsOptional()
  @IsDateString()
  flaggedAt?: Date;

  @Column({ type: 'varchar', length: 500, nullable: true, name: 'flagged_reason' })
  @IsOptional()
  @IsString()
  flaggedReason?: string;

  @Column({ type: 'jsonb', nullable: true })
  metadata?: Record<string, any>;

  @CreateDateColumn({ name: 'created_at' })
  createdAt: Date;

  @UpdateDateColumn({ name: 'updated_at' })
  updatedAt: Date;

  // ========================================
  // RELATIONSHIPS
  // ========================================

  @ManyToOne(() => User, user => user.riskProfiles, {
    onDelete: 'CASCADE',
    eager: false
  })
  @JoinColumn({ name: 'user_id' })
  user: User;

  @OneToMany(() => RiskEvent, riskEvent => riskEvent.riskProfile, {
    cascade: true,
    eager: false
  })
  riskEvents: RiskEvent[];

  // ========================================
  // LIFECYCLE HOOKS
  // ========================================

  @BeforeInsert()
  beforeInsert() {
    this.calculateRiskLevel();
    this.createdAt = new Date();
    this.updatedAt = new Date();
  }

  @BeforeUpdate()
  beforeUpdate() {
    this.calculateRiskLevel();
    this.calculateRiskTrend();
    this.updatedAt = new Date();
  }

  // ========================================
  // BUSINESS METHODS
  // ========================================

  /**
   * Calcular nível de risco baseado na pontuação
   */
  private calculateRiskLevel(): void {
    const score = this.currentRiskScore;
    
    if (score >= 90) {
      this.riskLevel = RiskLevel.CRITICAL;
    } else if (score >= 75) {
      this.riskLevel = RiskLevel.VERY_HIGH;
    } else if (score >= 60) {
      this.riskLevel = RiskLevel.HIGH;
    } else if (score >= 40) {
      this.riskLevel = RiskLevel.MEDIUM;
    } else if (score >= 20) {
      this.riskLevel = RiskLevel.LOW;
    } else {
      this.riskLevel = RiskLevel.VERY_LOW;
    }
  }

  /**
   * Calcular tendência de risco
   */
  private calculateRiskTrend(): void {
    if (!this.baselineRiskScore) {
      this.riskTrend = RiskTrend.STABLE;
      return;
    }

    const difference = this.currentRiskScore - this.baselineRiskScore;
    const threshold = 5; // 5 pontos de diferença

    if (Math.abs(difference) < threshold) {
      this.riskTrend = RiskTrend.STABLE;
    } else if (difference > threshold) {
      this.riskTrend = RiskTrend.INCREASING;
    } else {
      this.riskTrend = RiskTrend.DECREASING;
    }

    // Verificar volatilidade baseada em eventos recentes
    if (this.hasVolatilePattern()) {
      this.riskTrend = RiskTrend.VOLATILE;
    }
  }

  /**
   * Verificar se há padrão volátil
   */
  private hasVolatilePattern(): boolean {
    // Implementar lógica para detectar volatilidade
    // baseada em variações recentes de risco
    return false; // Placeholder
  }

  /**
   * Atualizar pontuação de risco
   */
  updateRiskScore(newScore: number, confidence: number = 0.8): void {
    // Atualizar baseline se for a primeira avaliação
    if (this.assessmentCount === 0) {
      this.baselineRiskScore = newScore;
    }

    // Atualizar pontuação atual
    this.currentRiskScore = Math.max(0, Math.min(100, newScore));
    this.confidenceScore = Math.max(0, Math.min(1, confidence));

    // Atualizar pico se necessário
    if (this.currentRiskScore > this.peakRiskScore) {
      this.peakRiskScore = this.currentRiskScore;
    }

    // Incrementar contador de avaliações
    this.assessmentCount += 1;
    this.lastAssessmentAt = new Date();

    // Marcar eventos de alto risco
    if (this.isHighRisk()) {
      this.highRiskEvents += 1;
      this.lastHighRiskAt = new Date();
    }

    // Determinar se requer monitoramento
    this.requiresMonitoring = this.shouldRequireMonitoring();
  }

  /**
   * Verificar se é alto risco
   */
  isHighRisk(): boolean {
    return this.currentRiskScore >= 60 || 
           this.riskLevel === RiskLevel.HIGH ||
           this.riskLevel === RiskLevel.VERY_HIGH ||
           this.riskLevel === RiskLevel.CRITICAL;
  }

  /**
   * Verificar se é risco crítico
   */
  isCriticalRisk(): boolean {
    return this.currentRiskScore >= 90 || this.riskLevel === RiskLevel.CRITICAL;
  }

  /**
   * Determinar se requer monitoramento
   */
  private shouldRequireMonitoring(): boolean {
    return this.isHighRisk() || 
           this.riskTrend === RiskTrend.INCREASING ||
           this.riskTrend === RiskTrend.VOLATILE ||
           this.securityViolations > 0;
  }

  /**
   * Adicionar dispositivo confiável
   */
  addTrustedDevice(fingerprint: string): void {
    if (!this.deviceFingerprints) {
      this.deviceFingerprints = [];
    }
    
    if (!this.deviceFingerprints.includes(fingerprint)) {
      this.deviceFingerprints.push(fingerprint);
    }
  }

  /**
   * Remover dispositivo confiável
   */
  removeTrustedDevice(fingerprint: string): void {
    if (this.deviceFingerprints) {
      this.deviceFingerprints = this.deviceFingerprints.filter(
        fp => fp !== fingerprint
      );
    }
  }

  /**
   * Verificar se dispositivo é confiável
   */
  isTrustedDevice(fingerprint: string): boolean {
    return this.deviceFingerprints?.includes(fingerprint) || false;
  }

  /**
   * Adicionar localização confiável
   */
  addTrustedLocation(location: string): void {
    if (!this.trustedLocations) {
      this.trustedLocations = [];
    }
    
    if (!this.trustedLocations.includes(location)) {
      this.trustedLocations.push(location);
    }
  }

  /**
   * Verificar se localização é confiável
   */
  isTrustedLocation(location: string): boolean {
    return this.trustedLocations?.includes(location) || false;
  }

  /**
   * Adicionar IP suspeito
   */
  addSuspiciousIp(ip: string): void {
    if (!this.suspiciousIps) {
      this.suspiciousIps = [];
    }
    
    if (!this.suspiciousIps.includes(ip)) {
      this.suspiciousIps.push(ip);
    }
  }

  /**
   * Verificar se IP é suspeito
   */
  isSuspiciousIp(ip: string): boolean {
    return this.suspiciousIps?.includes(ip) || false;
  }

  /**
   * Atualizar padrões comportamentais
   */
  updateBehaviorPatterns(patterns: Partial<typeof this.behaviorPatterns>): void {
    this.behaviorPatterns = {
      ...this.behaviorPatterns,
      ...patterns
    };
  }

  /**
   * Atualizar fatores de risco
   */
  updateRiskFactors(factors: Partial<typeof this.riskFactors>): void {
    this.riskFactors = {
      ...this.riskFactors,
      ...factors
    };
  }

  /**
   * Atualizar features de ML
   */
  updateMLFeatures(features: Partial<typeof this.mlFeatures>): void {
    this.mlFeatures = {
      ...this.mlFeatures,
      ...features
    };
  }

  /**
   * Marcar como flagged
   */
  flag(reason: string): void {
    this.isFlagged = true;
    this.flaggedAt = new Date();
    this.flaggedReason = reason;
    this.requiresMonitoring = true;
  }

  /**
   * Remover flag
   */
  unflag(): void {
    this.isFlagged = false;
    this.flaggedAt = null;
    this.flaggedReason = null;
  }

  /**
   * Registrar violação de segurança
   */
  recordSecurityViolation(): void {
    this.securityViolations += 1;
    this.requiresMonitoring = true;
    
    // Auto-flag se muitas violações
    if (this.securityViolations >= 3) {
      this.flag('Multiple security violations detected');
    }
  }

  /**
   * Obter recomendações de segurança
   */
  getSecurityRecommendations(): string[] {
    const recommendations: string[] = [];

    if (this.isCriticalRisk()) {
      recommendations.push('require_immediate_verification');
      recommendations.push('block_suspicious_activities');
      recommendations.push('escalate_to_security_team');
    } else if (this.isHighRisk()) {
      recommendations.push('require_step_up_authentication');
      recommendations.push('increase_monitoring');
      recommendations.push('limit_sensitive_operations');
    }

    if (this.riskTrend === RiskTrend.INCREASING) {
      recommendations.push('monitor_behavior_changes');
    }

    if (this.riskTrend === RiskTrend.VOLATILE) {
      recommendations.push('investigate_anomalous_patterns');
    }

    if (this.securityViolations > 0) {
      recommendations.push('review_recent_activities');
    }

    if (!this.deviceFingerprints?.length) {
      recommendations.push('establish_device_trust');
    }

    return recommendations;
  }

  /**
   * Obter estatísticas de risco
   */
  getRiskStatistics() {
    return {
      currentScore: this.currentRiskScore,
      baselineScore: this.baselineRiskScore,
      peakScore: this.peakRiskScore,
      level: this.riskLevel,
      trend: this.riskTrend,
      confidence: this.confidenceScore,
      assessmentCount: this.assessmentCount,
      highRiskEvents: this.highRiskEvents,
      securityViolations: this.securityViolations,
      trustedDevices: this.deviceFingerprints?.length || 0,
      trustedLocations: this.trustedLocations?.length || 0,
      isFlagged: this.isFlagged,
      requiresMonitoring: this.requiresMonitoring,
      lastAssessment: this.lastAssessmentAt,
      lastHighRisk: this.lastHighRiskAt
    };
  }

  /**
   * Calcular score de risco composto
   */
  calculateCompositeRisk(): number {
    if (!this.riskFactors) return this.currentRiskScore;

    const weights = {
      deviceRisk: 0.25,
      locationRisk: 0.20,
      behavioralRisk: 0.25,
      temporalRisk: 0.15,
      velocityRisk: 0.10,
      anomalyRisk: 0.05
    };

    let compositeScore = 0;
    let totalWeight = 0;

    Object.entries(this.riskFactors).forEach(([factor, score]) => {
      if (score !== undefined && weights[factor]) {
        compositeScore += score * weights[factor];
        totalWeight += weights[factor];
      }
    });

    return totalWeight > 0 ? compositeScore / totalWeight : this.currentRiskScore;
  }

  /**
   * Validar integridade do perfil
   */
  validateIntegrity(): string[] {
    const errors: string[] = [];

    if (!this.userId) {
      errors.push('User ID is required');
    }

    if (!this.tenantId) {
      errors.push('Tenant ID is required');
    }

    if (this.currentRiskScore < 0 || this.currentRiskScore > 100) {
      errors.push('Risk score must be between 0 and 100');
    }

    if (this.confidenceScore < 0 || this.confidenceScore > 1) {
      errors.push('Confidence score must be between 0 and 1');
    }

    if (this.assessmentCount < 0) {
      errors.push('Assessment count cannot be negative');
    }

    return errors;
  }

  /**
   * Criar perfil de risco inicial
   */
  static createInitialProfile(userId: string, tenantId: string): RiskProfile {
    const profile = new RiskProfile();
    profile.userId = userId;
    profile.tenantId = tenantId;
    profile.baselineRiskScore = 25; // Score inicial baixo
    profile.currentRiskScore = 25;
    profile.peakRiskScore = 25;
    profile.riskLevel = RiskLevel.LOW;
    profile.riskTrend = RiskTrend.STABLE;
    profile.confidenceScore = 0.5;
    profile.assessmentCount = 0;
    profile.highRiskEvents = 0;
    profile.securityViolations = 0;
    profile.isActive = true;
    profile.requiresMonitoring = false;
    profile.isFlagged = false;
    profile.deviceFingerprints = [];
    profile.trustedLocations = [];
    profile.suspiciousIps = [];
    
    return profile;
  }

  /**
   * Serializar para JSON
   */
  toJSON() {
    return {
      id: this.id,
      userId: this.userId,
      currentRiskScore: this.currentRiskScore,
      riskLevel: this.riskLevel,
      riskTrend: this.riskTrend,
      confidenceScore: this.confidenceScore,
      isActive: this.isActive,
      requiresMonitoring: this.requiresMonitoring,
      isFlagged: this.isFlagged,
      statistics: this.getRiskStatistics(),
      recommendations: this.getSecurityRecommendations()
    };
  }
}