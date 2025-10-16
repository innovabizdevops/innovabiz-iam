/**
 * 🔑 CREDENTIAL ENTITY - INNOVABIZ IAM
 * Entidade de credencial WebAuthn com segurança avançada
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: W3C WebAuthn Level 3, FIDO2 CTAP2.1, NIST SP 800-63B
 * Standards: ISO 27001, PCI DSS 4.0, GDPR/LGPD, Multi-tenant
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
import { IsUUID, IsString, IsBoolean, IsDateString, IsOptional, IsEnum, IsNumber, IsArray } from 'class-validator';
import { createHash } from 'crypto';

import { User } from './User.entity';

export enum CredentialType {
  WEBAUTHN = 'webauthn',
  FIDO2 = 'fido2',
  PLATFORM = 'platform',
  CROSS_PLATFORM = 'cross-platform'
}

export enum DeviceType {
  PLATFORM = 'platform',
  CROSS_PLATFORM = 'cross-platform',
  UNKNOWN = 'unknown'
}

export enum AttestationType {
  NONE = 'none',
  INDIRECT = 'indirect',
  DIRECT = 'direct',
  ENTERPRISE = 'enterprise'
}

export enum CredentialStatus {
  ACTIVE = 'active',
  INACTIVE = 'inactive',
  REVOKED = 'revoked',
  COMPROMISED = 'compromised',
  EXPIRED = 'expired'
}

@Entity('credentials')
@Index(['userId', 'tenantId'])
@Index(['credentialId'], { unique: true })
@Index(['credentialHash'])
@Index(['isActive'])
@Index(['deviceType'])
@Index(['createdAt'])
@Index(['lastUsedAt'])
export class Credential {
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

  @Column({ type: 'varchar', length: 1024, name: 'credential_id', unique: true })
  @IsString()
  credentialId: string;

  @Column({ type: 'varchar', length: 64, name: 'credential_hash' })
  @IsString()
  @Index()
  credentialHash: string; // Hash do credential ID para busca rápida

  @Column({ type: 'text', name: 'public_key' })
  @IsString()
  @Exclude()
  publicKey: string;

  @Column({ type: 'bigint', default: 0 })
  @IsNumber()
  counter: number;

  @Column({ 
    type: 'enum', 
    enum: DeviceType, 
    default: DeviceType.UNKNOWN,
    name: 'device_type'
  })
  @IsEnum(DeviceType)
  deviceType: DeviceType;

  @Column({ 
    type: 'enum', 
    enum: CredentialType, 
    default: CredentialType.WEBAUTHN 
  })
  @IsEnum(CredentialType)
  type: CredentialType;

  @Column({ 
    type: 'enum', 
    enum: AttestationType, 
    default: AttestationType.NONE 
  })
  @IsEnum(AttestationType)
  attestationType: AttestationType;

  @Column({ 
    type: 'enum', 
    enum: CredentialStatus, 
    default: CredentialStatus.ACTIVE 
  })
  @IsEnum(CredentialStatus)
  status: CredentialStatus;

  @Column({ type: 'simple-array', nullable: true })
  @IsOptional()
  @IsArray()
  transports?: string[];

  @Column({ type: 'boolean', default: true, name: 'is_active' })
  @IsBoolean()
  isActive: boolean;

  @Column({ type: 'boolean', default: false, name: 'is_backup_eligible' })
  @IsBoolean()
  isBackupEligible: boolean;

  @Column({ type: 'boolean', default: false, name: 'is_backup_state' })
  @IsBoolean()
  isBackupState: boolean;

  @Column({ type: 'varchar', length: 255, nullable: true })
  @IsOptional()
  @IsString()
  nickname?: string;

  @Column({ type: 'varchar', length: 500, nullable: true })
  @IsOptional()
  @IsString()
  description?: string;

  @Column({ type: 'varchar', length: 255, nullable: true, name: 'device_name' })
  @IsOptional()
  @IsString()
  deviceName?: string;

  @Column({ type: 'varchar', length: 255, nullable: true, name: 'device_model' })
  @IsOptional()
  @IsString()
  deviceModel?: string;

  @Column({ type: 'varchar', length: 255, nullable: true, name: 'device_os' })
  @IsOptional()
  @IsString()
  deviceOs?: string;

  @Column({ type: 'varchar', length: 255, nullable: true, name: 'browser_name' })
  @IsOptional()
  @IsString()
  browserName?: string;

  @Column({ type: 'varchar', length: 100, nullable: true, name: 'browser_version' })
  @IsOptional()
  @IsString()
  browserVersion?: string;

  @Column({ type: 'text', nullable: true, name: 'attestation_object' })
  @IsOptional()
  @IsString()
  @Exclude()
  attestationObject?: string;

  @Column({ type: 'text', nullable: true, name: 'client_data_json' })
  @IsOptional()
  @IsString()
  @Exclude()
  clientDataJson?: string;

  @Column({ type: 'varchar', length: 255, nullable: true, name: 'aaguid' })
  @IsOptional()
  @IsString()
  aaguid?: string; // Authenticator Attestation GUID

  @Column({ type: 'int', default: 0, name: 'usage_count' })
  @IsNumber()
  usageCount: number;

  @Column({ type: 'float', default: 0, name: 'risk_score' })
  @IsNumber()
  riskScore: number;

  @Column({ type: 'timestamp', nullable: true, name: 'last_used_at' })
  @IsOptional()
  @IsDateString()
  lastUsedAt?: Date;

  @Column({ type: 'timestamp', nullable: true, name: 'expires_at' })
  @IsOptional()
  @IsDateString()
  expiresAt?: Date;

  @Column({ type: 'jsonb', nullable: true })
  metadata?: Record<string, any>;

  @Column({ type: 'jsonb', nullable: true, name: 'security_flags' })
  securityFlags?: {
    isCompromised?: boolean;
    isSuspicious?: boolean;
    requiresVerification?: boolean;
    isHighRisk?: boolean;
    hasAnomalousUsage?: boolean;
  };

  @Column({ type: 'jsonb', nullable: true, name: 'attestation_data' })
  @Exclude()
  attestationData?: {
    fmt: string;
    attStmt: any;
    authData: string;
  };

  @CreateDateColumn({ name: 'created_at' })
  createdAt: Date;

  @UpdateDateColumn({ name: 'updated_at' })
  updatedAt: Date;

  @Column({ type: 'timestamp', nullable: true, name: 'revoked_at' })
  @IsOptional()
  @IsDateString()
  @Exclude()
  revokedAt?: Date;

  // ========================================
  // RELATIONSHIPS
  // ========================================

  @ManyToOne(() => User, user => user.credentials, {
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
    this.generateCredentialHash();
    this.createdAt = new Date();
    this.updatedAt = new Date();
  }

  @BeforeUpdate()
  beforeUpdate() {
    this.updatedAt = new Date();
  }

  // ========================================
  // BUSINESS METHODS
  // ========================================

  /**
   * Gerar hash do credential ID para indexação
   */
  private generateCredentialHash(): void {
    if (this.credentialId) {
      this.credentialHash = createHash('sha256')
        .update(this.credentialId)
        .digest('hex');
    }
  }

  /**
   * Verificar se a credencial está válida
   */
  isValid(): boolean {
    if (!this.isActive) return false;
    if (this.status !== CredentialStatus.ACTIVE) return false;
    if (this.expiresAt && this.expiresAt <= new Date()) return false;
    return true;
  }

  /**
   * Verificar se a credencial expirou
   */
  isExpired(): boolean {
    return this.expiresAt ? this.expiresAt <= new Date() : false;
  }

  /**
   * Revogar credencial
   */
  revoke(reason?: string): void {
    this.isActive = false;
    this.status = CredentialStatus.REVOKED;
    this.revokedAt = new Date();
    
    if (reason) {
      this.metadata = {
        ...this.metadata,
        revocationReason: reason,
        revokedAt: new Date().toISOString()
      };
    }
  }

  /**
   * Marcar como comprometida
   */
  markCompromised(reason: string): void {
    this.isActive = false;
    this.status = CredentialStatus.COMPROMISED;
    this.revokedAt = new Date();
    
    this.securityFlags = {
      ...this.securityFlags,
      isCompromised: true
    };
    
    this.metadata = {
      ...this.metadata,
      compromiseReason: reason,
      compromisedAt: new Date().toISOString()
    };
  }

  /**
   * Atualizar uso da credencial
   */
  updateUsage(newCounter: number): void {
    // Verificar rollback attack
    if (newCounter <= this.counter) {
      this.markSuspicious('Counter rollback detected');
      throw new Error('Counter rollback attack detected');
    }
    
    this.counter = newCounter;
    this.usageCount += 1;
    this.lastUsedAt = new Date();
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
   * Atualizar pontuação de risco
   */
  updateRiskScore(newScore: number): void {
    this.riskScore = Math.max(0, Math.min(100, newScore));
    
    // Atualizar flags de segurança baseado no risco
    this.securityFlags = {
      ...this.securityFlags,
      isHighRisk: this.riskScore >= 70,
      requiresVerification: this.riskScore >= 50
    };
  }

  /**
   * Verificar se precisa de verificação adicional
   */
  requiresAdditionalVerification(): boolean {
    return this.securityFlags?.requiresVerification || 
           this.securityFlags?.isSuspicious || 
           this.riskScore >= 50;
  }

  /**
   * Obter informações do dispositivo
   */
  getDeviceInfo(): string {
    const parts = [this.deviceName, this.deviceModel, this.deviceOs].filter(Boolean);
    return parts.join(' - ') || 'Unknown Device';
  }

  /**
   * Obter informações do navegador
   */
  getBrowserInfo(): string {
    if (this.browserName && this.browserVersion) {
      return `${this.browserName} ${this.browserVersion}`;
    }
    return this.browserName || 'Unknown Browser';
  }

  /**
   * Verificar se é credencial de plataforma
   */
  isPlatformCredential(): boolean {
    return this.deviceType === DeviceType.PLATFORM;
  }

  /**
   * Verificar se é credencial cross-platform
   */
  isCrossPlatformCredential(): boolean {
    return this.deviceType === DeviceType.CROSS_PLATFORM;
  }

  /**
   * Verificar se suporta backup
   */
  supportsBackup(): boolean {
    return this.isBackupEligible;
  }

  /**
   * Verificar se está em estado de backup
   */
  isInBackupState(): boolean {
    return this.isBackupState;
  }

  /**
   * Obter transportes suportados
   */
  getSupportedTransports(): string[] {
    return this.transports || [];
  }

  /**
   * Verificar se suporta transporte específico
   */
  supportsTransport(transport: string): boolean {
    return this.getSupportedTransports().includes(transport);
  }

  /**
   * Calcular idade da credencial em dias
   */
  getAgeInDays(): number {
    const now = new Date();
    const diffTime = Math.abs(now.getTime() - this.createdAt.getTime());
    return Math.ceil(diffTime / (1000 * 60 * 60 * 24));
  }

  /**
   * Verificar se é credencial antiga
   */
  isOld(daysThreshold: number = 365): boolean {
    return this.getAgeInDays() > daysThreshold;
  }

  /**
   * Obter estatísticas de uso
   */
  getUsageStats() {
    return {
      usageCount: this.usageCount,
      lastUsedAt: this.lastUsedAt,
      ageInDays: this.getAgeInDays(),
      riskScore: this.riskScore,
      counter: this.counter,
      isActive: this.isActive,
      status: this.status
    };
  }

  /**
   * Validar integridade da credencial
   */
  validateIntegrity(): string[] {
    const errors: string[] = [];

    if (!this.credentialId || this.credentialId.length < 16) {
      errors.push('Invalid credential ID');
    }

    if (!this.publicKey || this.publicKey.length < 32) {
      errors.push('Invalid public key');
    }

    if (!this.userId) {
      errors.push('User ID is required');
    }

    if (!this.tenantId) {
      errors.push('Tenant ID is required');
    }

    if (this.counter < 0) {
      errors.push('Counter cannot be negative');
    }

    if (this.expiresAt && this.expiresAt <= this.createdAt) {
      errors.push('Expiration date must be after creation date');
    }

    return errors;
  }

  /**
   * Criar nova credencial WebAuthn
   */
  static createWebAuthnCredential(
    userId: string,
    tenantId: string,
    credentialId: string,
    publicKey: string,
    counter: number,
    deviceType: DeviceType,
    transports: string[] = [],
    attestationData?: any
  ): Credential {
    const credential = new Credential();
    credential.userId = userId;
    credential.tenantId = tenantId;
    credential.credentialId = credentialId;
    credential.publicKey = publicKey;
    credential.counter = counter;
    credential.deviceType = deviceType;
    credential.type = CredentialType.WEBAUTHN;
    credential.transports = transports;
    credential.isActive = true;
    credential.status = CredentialStatus.ACTIVE;
    credential.usageCount = 0;
    credential.riskScore = 0;
    
    if (attestationData) {
      credential.attestationData = attestationData;
      credential.attestationType = attestationData.fmt || AttestationType.NONE;
    }
    
    return credential;
  }

  /**
   * Serializar para JSON (excluindo dados sensíveis)
   */
  toJSON() {
    const { 
      publicKey, 
      attestationObject, 
      clientDataJson, 
      attestationData,
      credentialHash,
      revokedAt,
      ...publicData 
    } = this;
    return publicData;
  }

  /**
   * Obter informações públicas da credencial
   */
  getPublicInfo() {
    return {
      id: this.id,
      nickname: this.nickname,
      description: this.description,
      deviceType: this.deviceType,
      type: this.type,
      status: this.status,
      isActive: this.isActive,
      deviceInfo: this.getDeviceInfo(),
      browserInfo: this.getBrowserInfo(),
      transports: this.getSupportedTransports(),
      usageCount: this.usageCount,
      lastUsedAt: this.lastUsedAt,
      createdAt: this.createdAt,
      ageInDays: this.getAgeInDays(),
      riskScore: this.riskScore
    };
  }

  /**
   * Obter resumo da credencial para auditoria
   */
  getAuditSummary() {
    return {
      credentialId: this.credentialId.substring(0, 16) + '...',
      deviceType: this.deviceType,
      type: this.type,
      status: this.status,
      usageCount: this.usageCount,
      counter: this.counter,
      riskScore: this.riskScore,
      createdAt: this.createdAt,
      lastUsedAt: this.lastUsedAt
    };
  }
}