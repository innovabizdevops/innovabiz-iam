/**
 * 📊 RISK EVENT ENTITY - INNOVABIZ IAM
 * Entidade de evento de risco para análise e ML
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: NIST AI RMF, ISO/IEC 42001, COSO ERM, Basel III
 * Standards: NIST Cybersecurity Framework, ISO 31000
 */

import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  ManyToOne,
  JoinColumn,
  Index,
  BeforeInsert
} from 'typeorm';
import { IsUUID, IsString, IsEnum, IsOptional, IsNumber, IsArray, IsDateString, IsIP, Min, Max } from 'class-validator';

import { User } from './User.entity';
import { RiskProfile } from './RiskProfile.entity';

export enum RiskEventType {
  AUTHENTICATION = 'authentication',
  REGISTRATION = 'registration',
  LOGIN_ATTEMPT = 'login_attempt',
  SESSION_CREATION = 'session_creation',
  CREDENTIAL_USAGE = 'credential_usage',
  DEVICE_CHANGE = 'device_change',
  LOCATION_CHANGE = 'location_change',
  BEHAVIOR_ANOMALY = 'behavior_anomaly',
  VELOCITY_VIOLATION = 'velocity_violation',
  SECURITY_VIOLATION = 'security_violation',
  SUSPICIOUS_ACTIVITY = 'suspicious_activity'
}

export enum RiskEventSeverity {
  VERY_LOW = 'very_low',
  LOW = 'low',
  MEDIUM = 'medium',
  HIGH = 'high',
  VERY_HIGH = 'very_high',
  CRITICAL = 'critical'
}

export enum RiskEventStatus {
  DETECTED = 'detected',
  ANALYZING = 'analyzing',
  CONFIRMED = 'confirmed',
  FALSE_POSITIVE = 'false_positive',
  MITIGATED = 'mitigated',
  RESOLVED = 'resolved'
}

@Entity('risk_events')
@Index(['userId', 'tenantId'])
@Index(['eventType'])
@Index(['severity'])
@Index(['riskScore'])
@Index(['timestamp'])
@Index(['ipAddress'])
@Index(['status'])
export class RiskEvent {
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

  @Column({ type: 'uuid', name: 'risk_profile_id', nullable: true })
  @IsOptional()
  @IsUUID()
  riskProfileId?: string;

  @Column({ 
    type: 'enum', 
    enum: RiskEventType,
    name: 'event_type'
  })
  @IsEnum(RiskEventType)
  eventType: RiskEventType;

  @Column({ 
    type: 'enum', 
    enum: RiskEventSeverity,
    default: RiskEventSeverity.LOW
  })
  @IsEnum(RiskEventSeverity)
  severity: RiskEventSeverity;

  @Column({ 
    type: 'enum', 
    enum: RiskEventStatus,
    default: RiskEventStatus.DETECTED
  })
  @IsEnum(RiskEventStatus)
  status: RiskEventStatus;

  @Column({ type: 'float', name: 'risk_score' })
  @IsNumber()
  @Min(0)
  @Max(100)
  riskScore: number;

  @Column({ type: 'float', default: 0, name: 'confidence_score' })
  @IsNumber()
  @Min(0)
  @Max(1)
  confidenceScore: number;

  @Column({ type: 'simple-array', name: 'risk_factors' })
  @IsArray()
  riskFactors: string[];

  @Column({ type: 'inet', nullable: true, name: 'ip_address' })
  @IsOptional()
  @IsIP()
  ipAddress?: string;

  @Column({ type: 'varchar', length: 1000, nullable: true, name: 'user_agent' })
  @IsOptional()
  @IsString()
  userAgent?: string;

  @Column({ type: 'varchar', length: 255, nullable: true, name: 'device_fingerprint' })
  @IsOptional()
  @IsString()
  deviceFingerprint?: string;

  @Column({ type: 'varchar', length: 255, nullable: true, name: 'session_id' })
  @IsOptional()
  @IsString()
  sessionId?: string;

  @Column({ type: 'varchar', length: 255, nullable: true, name: 'credential_id' })
  @IsOptional()
  @IsString()
  credentialId?: string;

  @Column({ type: 'varchar', length: 100, nullable: true })
  @IsOptional()
  @IsString()
  country?: string;

  @Column({ type: 'varchar', length: 100, nullable: true })
  @IsOptional()
  @IsString()
  region?: string;

  @Column({ type: 'varchar', length: 100, nullable: true })
  @IsOptional()
  @IsString()
  city?: string;

  @Column({ type: 'jsonb', nullable: true })
  metadata?: {
    // Device Information
    deviceType?: string;
    browserName?: string;
    browserVersion?: string;
    operatingSystem?: string;
    
    // Location Information
    timezone?: string;
    isp?: string;
    organization?: string;
    
    // Behavioral Information
    sessionDuration?: number;
    activityCount?: number;
    navigationPattern?: string[];
    
    // Security Flags
    isVpn?: boolean;
    isTor?: boolean;
    isProxy?: boolean;
    isMaliciousIp?: boolean;
    
    // ML Features
    mlFeatures?: Record<string, number>;
    mlPrediction?: number;
    mlConfidence?: number;
    
    // Additional Context
    previousRiskScore?: number;
    riskScoreDelta?: number;
    triggerRules?: string[];
    mitigationActions?: string[];
  };

  @Column({ type: 'jsonb', nullable: true, name: 'detection_rules' })
  detectionRules?: {
    ruleId: string;
    ruleName: string;
    ruleType: string;
    threshold: number;
    actualValue: number;
    severity: string;
  }[];

  @Column({ type: 'jsonb', nullable: true, name: 'ml_analysis' })
  mlAnalysis?: {
    modelVersion: string;
    features: Record<string, number>;
    prediction: number;
    confidence: number;
    featureImportance: Record<string, number>;
    anomalyScore: number;
    clusterLabel?: string;
  };

  @Column({ type: 'text', nullable: true })
  @IsOptional()
  @IsString()
  description?: string;

  @Column({ type: 'text', nullable: true, name: 'mitigation_notes' })
  @IsOptional()
  @IsString()
  mitigationNotes?: string;

  @Column({ type: 'varchar', length: 255, nullable: true, name: 'analyst_id' })
  @IsOptional()
  @IsString()
  analystId?: string;

  @Column({ type: 'timestamp', nullable: true, name: 'analyzed_at' })
  @IsOptional()
  @IsDateString()
  analyzedAt?: Date;

  @Column({ type: 'timestamp', nullable: true, name: 'resolved_at' })
  @IsOptional()
  @IsDateString()
  resolvedAt?: Date;

  @CreateDateColumn({ name: 'timestamp' })
  timestamp: Date;

  // ========================================
  // RELATIONSHIPS
  // ========================================

  @ManyToOne(() => User, user => user.auditLogs, {
    onDelete: 'CASCADE',
    eager: false
  })
  @JoinColumn({ name: 'user_id' })
  user: User;

  @ManyToOne(() => RiskProfile, riskProfile => riskProfile.riskEvents, {
    onDelete: 'SET NULL',
    eager: false
  })
  @JoinColumn({ name: 'risk_profile_id' })
  riskProfile?: RiskProfile;

  // ========================================
  // LIFECYCLE HOOKS
  // ========================================

  @BeforeInsert()
  beforeInsert() {
    this.calculateSeverity();
    this.timestamp = new Date();
  }

  // ========================================
  // BUSINESS METHODS
  // ========================================

  /**
   * Calcular severidade baseada na pontuação de risco
   */
  private calculateSeverity(): void {
    if (this.riskScore >= 90) {
      this.severity = RiskEventSeverity.CRITICAL;
    } else if (this.riskScore >= 75) {
      this.severity = RiskEventSeverity.VERY_HIGH;
    } else if (this.riskScore >= 60) {
      this.severity = RiskEventSeverity.HIGH;
    } else if (this.riskScore >= 40) {
      this.severity = RiskEventSeverity.MEDIUM;
    } else if (this.riskScore >= 20) {
      this.severity = RiskEventSeverity.LOW;
    } else {
      this.severity = RiskEventSeverity.VERY_LOW;
    }
  }

  /**
   * Verificar se é evento crítico
   */
  isCritical(): boolean {
    return this.severity === RiskEventSeverity.CRITICAL || this.riskScore >= 90;
  }

  /**
   * Verificar se é evento de alta severidade
   */
  isHighSeverity(): boolean {
    return this.severity === RiskEventSeverity.HIGH || 
           this.severity === RiskEventSeverity.VERY_HIGH ||
           this.severity === RiskEventSeverity.CRITICAL;
  }

  /**
   * Verificar se requer análise imediata
   */
  requiresImmediateAnalysis(): boolean {
    return this.isCritical() || 
           this.isHighSeverity() ||
           this.riskFactors.some(factor => 
             ['credential_compromised', 'brute_force', 'malicious_ip'].includes(factor)
           );
  }

  /**
   * Marcar como analisado
   */
  markAnalyzed(analystId: string, notes?: string): void {
    this.status = RiskEventStatus.ANALYZING;
    this.analystId = analystId;
    this.analyzedAt = new Date();
    
    if (notes) {
      this.mitigationNotes = notes;
    }
  }

  /**
   * Confirmar como verdadeiro positivo
   */
  confirmThreat(analystId: string, mitigationNotes?: string): void {
    this.status = RiskEventStatus.CONFIRMED;
    this.analystId = analystId;
    this.analyzedAt = new Date();
    
    if (mitigationNotes) {
      this.mitigationNotes = mitigationNotes;
    }
  }

  /**
   * Marcar como falso positivo
   */
  markFalsePositive(analystId: string, reason?: string): void {
    this.status = RiskEventStatus.FALSE_POSITIVE;
    this.analystId = analystId;
    this.analyzedAt = new Date();
    this.resolvedAt = new Date();
    
    if (reason) {
      this.mitigationNotes = `False Positive: ${reason}`;
    }
  }

  /**
   * Marcar como mitigado
   */
  markMitigated(analystId: string, mitigationActions: string[]): void {
    this.status = RiskEventStatus.MITIGATED;
    this.analystId = analystId;
    this.analyzedAt = new Date();
    
    this.metadata = {
      ...this.metadata,
      mitigationActions,
      mitigatedAt: new Date().toISOString()
    };
  }

  /**
   * Resolver evento
   */
  resolve(analystId: string, resolutionNotes?: string): void {
    this.status = RiskEventStatus.RESOLVED;
    this.analystId = analystId;
    this.resolvedAt = new Date();
    
    if (resolutionNotes) {
      this.mitigationNotes = resolutionNotes;
    }
  }

  /**
   * Adicionar análise de ML
   */
  addMLAnalysis(analysis: typeof this.mlAnalysis): void {
    this.mlAnalysis = analysis;
    
    // Atualizar confiança baseada na análise ML
    if (analysis?.confidence) {
      this.confidenceScore = Math.max(this.confidenceScore, analysis.confidence);
    }
  }

  /**
   * Adicionar regras de detecção
   */
  addDetectionRules(rules: typeof this.detectionRules): void {
    this.detectionRules = rules;
  }

  /**
   * Obter contexto geográfico
   */
  getGeographicContext(): string {
    const parts = [this.city, this.region, this.country].filter(Boolean);
    return parts.join(', ') || 'Unknown Location';
  }

  /**
   * Obter informações do dispositivo
   */
  getDeviceInfo(): string {
    if (!this.metadata) return 'Unknown Device';
    
    const parts = [
      this.metadata.deviceType,
      this.metadata.browserName,
      this.metadata.operatingSystem
    ].filter(Boolean);
    
    return parts.join(' - ') || 'Unknown Device';
  }

  /**
   * Verificar se é de origem suspeita
   */
  isSuspiciousOrigin(): boolean {
    return !!(this.metadata?.isVpn || 
              this.metadata?.isTor || 
              this.metadata?.isProxy || 
              this.metadata?.isMaliciousIp);
  }

  /**
   * Obter fatores de risco principais
   */
  getPrimaryRiskFactors(): string[] {
    const criticalFactors = [
      'credential_compromised',
      'brute_force',
      'malicious_ip',
      'tor_network',
      'suspicious_device'
    ];
    
    return this.riskFactors.filter(factor => 
      criticalFactors.includes(factor)
    );
  }

  /**
   * Calcular idade do evento em horas
   */
  getAgeInHours(): number {
    const now = new Date();
    const diffTime = Math.abs(now.getTime() - this.timestamp.getTime());
    return Math.floor(diffTime / (1000 * 60 * 60));
  }

  /**
   * Verificar se é evento recente
   */
  isRecent(hoursThreshold: number = 24): boolean {
    return this.getAgeInHours() <= hoursThreshold;
  }

  /**
   * Obter recomendações de ação
   */
  getActionRecommendations(): string[] {
    const recommendations: string[] = [];

    if (this.isCritical()) {
      recommendations.push('immediate_investigation');
      recommendations.push('block_suspicious_activity');
      recommendations.push('escalate_to_security_team');
    }

    if (this.isHighSeverity()) {
      recommendations.push('require_additional_verification');
      recommendations.push('increase_monitoring');
      recommendations.push('review_user_permissions');
    }

    if (this.isSuspiciousOrigin()) {
      recommendations.push('verify_user_identity');
      recommendations.push('check_account_compromise');
    }

    if (this.riskFactors.includes('new_device')) {
      recommendations.push('device_verification');
    }

    if (this.riskFactors.includes('unusual_location')) {
      recommendations.push('location_verification');
    }

    if (this.riskFactors.includes('velocity_violation')) {
      recommendations.push('rate_limiting');
      recommendations.push('temporary_account_restriction');
    }

    return recommendations;
  }

  /**
   * Obter estatísticas do evento
   */
  getEventStatistics() {
    return {
      id: this.id,
      eventType: this.eventType,
      severity: this.severity,
      status: this.status,
      riskScore: this.riskScore,
      confidenceScore: this.confidenceScore,
      riskFactors: this.riskFactors,
      location: this.getGeographicContext(),
      deviceInfo: this.getDeviceInfo(),
      isSuspiciousOrigin: this.isSuspiciousOrigin(),
      ageInHours: this.getAgeInHours(),
      isRecent: this.isRecent(),
      requiresAnalysis: this.requiresImmediateAnalysis(),
      timestamp: this.timestamp,
      analyzedAt: this.analyzedAt,
      resolvedAt: this.resolvedAt
    };
  }

  /**
   * Validar integridade do evento
   */
  validateIntegrity(): string[] {
    const errors: string[] = [];

    if (!this.userId) {
      errors.push('User ID is required');
    }

    if (!this.tenantId) {
      errors.push('Tenant ID is required');
    }

    if (this.riskScore < 0 || this.riskScore > 100) {
      errors.push('Risk score must be between 0 and 100');
    }

    if (this.confidenceScore < 0 || this.confidenceScore > 1) {
      errors.push('Confidence score must be between 0 and 1');
    }

    if (!this.riskFactors || this.riskFactors.length === 0) {
      errors.push('At least one risk factor is required');
    }

    return errors;
  }

  /**
   * Criar evento de risco
   */
  static createRiskEvent(
    userId: string,
    tenantId: string,
    eventType: RiskEventType,
    riskScore: number,
    riskFactors: string[],
    options: {
      riskProfileId?: string;
      confidenceScore?: number;
      ipAddress?: string;
      userAgent?: string;
      deviceFingerprint?: string;
      sessionId?: string;
      credentialId?: string;
      country?: string;
      region?: string;
      city?: string;
      description?: string;
      metadata?: any;
    } = {}
  ): RiskEvent {
    const event = new RiskEvent();
    
    event.userId = userId;
    event.tenantId = tenantId;
    event.eventType = eventType;
    event.riskScore = riskScore;
    event.riskFactors = riskFactors;
    event.riskProfileId = options.riskProfileId;
    event.confidenceScore = options.confidenceScore || 0.8;
    event.ipAddress = options.ipAddress;
    event.userAgent = options.userAgent;
    event.deviceFingerprint = options.deviceFingerprint;
    event.sessionId = options.sessionId;
    event.credentialId = options.credentialId;
    event.country = options.country;
    event.region = options.region;
    event.city = options.city;
    event.description = options.description;
    event.metadata = options.metadata;
    event.status = RiskEventStatus.DETECTED;
    
    return event;
  }

  /**
   * Serializar para JSON
   */
  toJSON() {
    return {
      id: this.id,
      eventType: this.eventType,
      severity: this.severity,
      status: this.status,
      riskScore: this.riskScore,
      confidenceScore: this.confidenceScore,
      riskFactors: this.riskFactors,
      location: this.getGeographicContext(),
      deviceInfo: this.getDeviceInfo(),
      description: this.description,
      timestamp: this.timestamp,
      statistics: this.getEventStatistics(),
      recommendations: this.getActionRecommendations()
    };
  }

  /**
   * Obter dados para análise ML
   */
  getMLTrainingData() {
    return {
      eventType: this.eventType,
      riskScore: this.riskScore,
      riskFactors: this.riskFactors,
      metadata: this.metadata,
      mlAnalysis: this.mlAnalysis,
      outcome: this.status === RiskEventStatus.CONFIRMED ? 1 : 0
    };
  }
}