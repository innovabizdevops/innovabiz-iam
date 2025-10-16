/**
 * 🔐 SESSION ENTITY - INNOVABIZ IAM
 * Entidade de sessão de usuário com segurança avançada
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: OWASP Session Management, NIST SP 800-63B, PCI DSS 4.0
 * Standards: ISO 27001, GDPR/LGPD, Multi-tenant Architecture
 */

import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
  ManyToOne,
  JoinColumn,
  Index,
  BeforeInsert,
  BeforeUpdate
} from 'typeorm';
import { Exclude } from 'class-transformer';
import { IsUUID, IsString, IsBoolean, IsDateString, IsIP, IsOptional, IsEnum } from 'class-validator';
import { randomBytes, createHash } from 'crypto';

import { User } from './User.entity';

export enum SessionStatus {
  ACTIVE = 'active',
  EXPIRED = 'expired',
  REVOKED = 'revoked',
  TERMINATED = 'terminated'
}

export enum SessionType {
  WEB = 'web',
  MOBILE = 'mobile',
  API = 'api',
  DESKTOP = 'desktop'
}

@Entity('sessions')
@Index(['userId', 'tenantId'])
@Index(['sessionToken'], { unique: true })
@Index(['refreshToken'], { unique: true })
@Index(['isActive'])
@Index(['expiresAt'])
@Index(['createdAt'])
export class Session {
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

  @Column({ type: 'varchar', length: 512, name: 'session_token', unique: true })
  @IsString()
  @Exclude()
  sessionToken: string;

  @Column({ type: 'varchar', length: 512, name: 'refresh_token', unique: true })
  @IsString()
  @Exclude()
  refreshToken: string;

  @Column({ type: 'varchar', length: 64, name: 'session_hash' })
  @IsString()
  @Index()
  sessionHash: string; // Hash do session token para busca rápida

  @Column({ type: 'boolean', default: true, name: 'is_active' })
  @IsBoolean()
  isActive: boolean;

  @Column({ 
    type: 'enum', 
    enum: SessionStatus, 
    default: SessionStatus.ACTIVE 
  })
  @IsEnum(SessionStatus)
  status: SessionStatus;

  @Column({ 
    type: 'enum', 
    enum: SessionType, 
    default: SessionType.WEB 
  })
  @IsEnum(SessionType)
  type: SessionType;

  @Column({ type: 'timestamp', name: 'expires_at' })
  @IsDateString()
  expiresAt: Date;

  @Column({ type: 'timestamp', nullable: true, name: 'last_activity_at' })
  @IsOptional()
  @IsDateString()
  lastActivityAt?: Date;

  @Column({ type: 'inet', name: 'ip_address' })
  @IsIP()
  ipAddress: string;

  @Column({ type: 'varchar', length: 1000, name: 'user_agent' })
  @IsString()
  userAgent: string;

  @Column({ type: 'varchar', length: 255, nullable: true, name: 'device_fingerprint' })
  @IsOptional()
  @IsString()
  deviceFingerprint?: string;

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

  @Column({ type: 'float', default: 0, name: 'risk_score' })
  riskScore: number;

  @Column({ type: 'int', default: 0, name: 'activity_count' })
  activityCount: number;

  @Column({ type: 'int', default: 0, name: 'duration_seconds' })
  durationSeconds: number;

  @Column({ type: 'jsonb', nullable: true })
  metadata?: Record<string, any>;

  @Column({ type: 'jsonb', nullable: true, name: 'security_flags' })
  securityFlags?: {
    isSuspicious?: boolean;
    isVpn?: boolean;
    isTor?: boolean;
    isProxy?: boolean;
    requiresMfa?: boolean;
    isHighRisk?: boolean;
  };

  @CreateDateColumn({ name: 'created_at' })
  createdAt: Date;

  @UpdateDateColumn({ name: 'updated_at' })
  updatedAt: Date;

  @Column({ type: 'timestamp', nullable: true, name: 'terminated_at' })
  @IsOptional()
  @IsDateString()
  @Exclude()
  terminatedAt?: Date;

  // ========================================
  // RELATIONSHIPS
  // ========================================

  @ManyToOne(() => User, user => user.sessions, {
    onDelete: 'CASCADE',
    eager: false
  })
  @JoinColumn({ name: 'user_id' })
  user: User;

  // ========================================
  // LIFECYCLE HOOKS
  // ========================================

  @BeforeInsert()
  beforeInsert() {
    this.generateSessionHash();
    this.lastActivityAt = new Date();
    this.createdAt = new Date();
    this.updatedAt = new Date();
  }

  @BeforeUpdate()
  beforeUpdate() {
    this.updatedAt = new Date();
    this.calculateDuration();
  }

  // ========================================
  // BUSINESS METHODS
  // ========================================

  /**
   * Gerar hash do session token para indexação
   */
  private generateSessionHash(): void {
    if (this.sessionToken) {
      this.sessionHash = createHash('sha256')
        .update(this.sessionToken)
        .digest('hex');
    }
  }

  /**
   * Verificar se a sessão está válida
   */
  isValid(): boolean {
    if (!this.isActive) return false;
    if (this.status !== SessionStatus.ACTIVE) return false;
    if (this.expiresAt <= new Date()) return false;
    return true;
  }

  /**
   * Verificar se a sessão expirou
   */
  isExpired(): boolean {
    return this.expiresAt <= new Date();
  }

  /**
   * Revogar sessão
   */
  revoke(reason?: string): void {
    this.isActive = false;
    this.status = SessionStatus.REVOKED;
    this.terminatedAt = new Date();
    
    if (reason) {
      this.metadata = {
        ...this.metadata,
        revocationReason: reason,
        revokedAt: new Date().toISOString()
      };
    }
  }

  /**
   * Terminar sessão
   */
  terminate(reason?: string): void {
    this.isActive = false;
    this.status = SessionStatus.TERMINATED;
    this.terminatedAt = new Date();
    
    if (reason) {
      this.metadata = {
        ...this.metadata,
        terminationReason: reason,
        terminatedAt: new Date().toISOString()
      };
    }
  }

  /**
   * Marcar como expirada
   */
  markExpired(): void {
    this.isActive = false;
    this.status = SessionStatus.EXPIRED;
    this.terminatedAt = new Date();
  }

  /**
   * Atualizar atividade da sessão
   */
  updateActivity(): void {
    this.lastActivityAt = new Date();
    this.activityCount += 1;
    this.calculateDuration();
  }

  /**
   * Calcular duração da sessão
   */
  private calculateDuration(): void {
    if (this.createdAt) {
      const now = this.terminatedAt || new Date();
      this.durationSeconds = Math.floor((now.getTime() - this.createdAt.getTime()) / 1000);
    }
  }

  /**
   * Estender expiração da sessão
   */
  extendExpiration(additionalMinutes: number = 60): void {
    const newExpiration = new Date(this.expiresAt.getTime() + (additionalMinutes * 60 * 1000));
    this.expiresAt = newExpiration;
  }

  /**
   * Verificar se a sessão precisa de renovação
   */
  needsRenewal(thresholdMinutes: number = 15): boolean {
    const threshold = new Date(Date.now() + (thresholdMinutes * 60 * 1000));
    return this.expiresAt <= threshold;
  }

  /**
   * Atualizar pontuação de risco
   */
  updateRiskScore(newScore: number): void {
    this.riskScore = Math.max(0, Math.min(100, newScore));
    
    // Atualizar flags de segurança baseado no risco
    this.securityFlags = {
      ...this.securityFlags,
      isHighRisk: this.riskScore >= 70,
      requiresMfa: this.riskScore >= 50
    };
  }

  /**
   * Marcar como suspeita
   */
  markSuspicious(reason: string): void {
    this.securityFlags = {
      ...this.securityFlags,
      isSuspicious: true
    };
    
    this.metadata = {
      ...this.metadata,
      suspiciousReason: reason,
      markedSuspiciousAt: new Date().toISOString()
    };
  }

  /**
   * Verificar se é de localização confiável
   */
  isFromTrustedLocation(trustedLocations: string[]): boolean {
    if (!this.country) return false;
    return trustedLocations.includes(this.country);
  }

  /**
   * Obter informações de geolocalização
   */
  getLocationInfo(): string {
    const parts = [this.city, this.region, this.country].filter(Boolean);
    return parts.join(', ') || 'Unknown';
  }

  /**
   * Verificar se é sessão de longa duração
   */
  isLongRunning(hoursThreshold: number = 8): boolean {
    return this.durationSeconds > (hoursThreshold * 3600);
  }

  /**
   * Obter estatísticas da sessão
   */
  getStats() {
    return {
      id: this.id,
      duration: this.durationSeconds,
      activityCount: this.activityCount,
      riskScore: this.riskScore,
      location: this.getLocationInfo(),
      deviceType: this.type,
      isActive: this.isActive,
      status: this.status,
      createdAt: this.createdAt,
      lastActivity: this.lastActivityAt,
      expiresAt: this.expiresAt
    };
  }

  /**
   * Validar integridade da sessão
   */
  validateIntegrity(): string[] {
    const errors: string[] = [];

    if (!this.sessionToken || this.sessionToken.length < 32) {
      errors.push('Invalid session token');
    }

    if (!this.refreshToken || this.refreshToken.length < 32) {
      errors.push('Invalid refresh token');
    }

    if (!this.userId) {
      errors.push('User ID is required');
    }

    if (!this.tenantId) {
      errors.push('Tenant ID is required');
    }

    if (!this.ipAddress) {
      errors.push('IP address is required');
    }

    if (!this.userAgent) {
      errors.push('User agent is required');
    }

    if (this.expiresAt <= this.createdAt) {
      errors.push('Expiration date must be after creation date');
    }

    return errors;
  }

  /**
   * Gerar tokens seguros
   */
  static generateSecureTokens(): { sessionToken: string; refreshToken: string } {
    return {
      sessionToken: randomBytes(64).toString('base64url'),
      refreshToken: randomBytes(64).toString('base64url')
    };
  }

  /**
   * Criar nova sessão
   */
  static createSession(
    userId: string,
    tenantId: string,
    ipAddress: string,
    userAgent: string,
    expirationMinutes: number = 60,
    type: SessionType = SessionType.WEB
  ): Session {
    const tokens = Session.generateSecureTokens();
    
    const session = new Session();
    session.userId = userId;
    session.tenantId = tenantId;
    session.sessionToken = tokens.sessionToken;
    session.refreshToken = tokens.refreshToken;
    session.ipAddress = ipAddress;
    session.userAgent = userAgent;
    session.type = type;
    session.expiresAt = new Date(Date.now() + (expirationMinutes * 60 * 1000));
    session.isActive = true;
    session.status = SessionStatus.ACTIVE;
    session.riskScore = 0;
    session.activityCount = 0;
    session.durationSeconds = 0;
    
    return session;
  }

  /**
   * Serializar para JSON (excluindo tokens sensíveis)
   */
  toJSON() {
    const { sessionToken, refreshToken, sessionHash, ...publicData } = this;
    return publicData;
  }

  /**
   * Obter informações públicas da sessão
   */
  getPublicInfo() {
    return {
      id: this.id,
      type: this.type,
      status: this.status,
      isActive: this.isActive,
      location: this.getLocationInfo(),
      deviceFingerprint: this.deviceFingerprint?.substring(0, 8) + '...',
      createdAt: this.createdAt,
      lastActivityAt: this.lastActivityAt,
      expiresAt: this.expiresAt,
      activityCount: this.activityCount,
      durationSeconds: this.durationSeconds
    };
  }
}