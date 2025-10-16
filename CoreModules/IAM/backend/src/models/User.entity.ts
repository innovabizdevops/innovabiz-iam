/**
 * 👤 USER ENTITY - INNOVABIZ IAM
 * Entidade principal de usuário do sistema
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: GDPR/LGPD, PCI DSS 4.0, Multi-tenant Architecture
 */

import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
  OneToMany,
  Index,
  BeforeInsert,
  BeforeUpdate
} from 'typeorm';
import { Exclude } from 'class-transformer';
import { IsEmail, IsUUID, IsString, IsBoolean, IsOptional, IsDateString } from 'class-validator';

import { Credential } from './Credential.entity';
import { Session } from './Session.entity';
import { AuditLog } from './AuditLog.entity';
import { RiskProfile } from './RiskProfile.entity';

@Entity('users')
@Index(['email', 'tenantId'], { unique: true })
@Index(['username', 'tenantId'], { unique: true })
@Index(['tenantId'])
@Index(['isActive'])
@Index(['createdAt'])
export class User {
  @PrimaryGeneratedColumn('uuid')
  @IsUUID()
  id: string;

  @Column({ type: 'uuid', name: 'tenant_id' })
  @IsUUID()
  @Index()
  tenantId: string;

  @Column({ type: 'varchar', length: 255, unique: false })
  @IsEmail()
  email: string;

  @Column({ type: 'varchar', length: 100, unique: false })
  @IsString()
  username: string;

  @Column({ type: 'varchar', length: 255, name: 'display_name' })
  @IsString()
  displayName: string;

  @Column({ type: 'boolean', default: true, name: 'is_active' })
  @IsBoolean()
  isActive: boolean;

  @Column({ type: 'boolean', default: false, name: 'is_verified' })
  @IsBoolean()
  isVerified: boolean;

  @Column({ type: 'boolean', default: false, name: 'is_locked' })
  @IsBoolean()
  isLocked: boolean;

  @Column({ type: 'timestamp', nullable: true, name: 'last_login_at' })
  @IsOptional()
  @IsDateString()
  lastLoginAt?: Date;

  @Column({ type: 'timestamp', nullable: true, name: 'locked_until' })
  @IsOptional()
  @IsDateString()
  lockedUntil?: Date;

  @Column({ type: 'int', default: 0, name: 'failed_login_attempts' })
  failedLoginAttempts: number;

  @Column({ type: 'varchar', length: 10, nullable: true })
  @IsOptional()
  @IsString()
  locale?: string;

  @Column({ type: 'varchar', length: 50, nullable: true })
  @IsOptional()
  @IsString()
  timezone?: string;

  @Column({ type: 'jsonb', nullable: true })
  preferences?: Record<string, any>;

  @Column({ type: 'jsonb', nullable: true })
  @Exclude()
  metadata?: Record<string, any>;

  @CreateDateColumn({ name: 'created_at' })
  createdAt: Date;

  @UpdateDateColumn({ name: 'updated_at' })
  updatedAt: Date;

  @Column({ type: 'timestamp', nullable: true, name: 'deleted_at' })
  @IsOptional()
  @IsDateString()
  @Exclude()
  deletedAt?: Date;

  // ========================================
  // RELATIONSHIPS
  // ========================================

  @OneToMany(() => Credential, credential => credential.user, {
    cascade: true,
    eager: false
  })
  credentials: Credential[];

  @OneToMany(() => Session, session => session.user, {
    cascade: true,
    eager: false
  })
  sessions: Session[];

  @OneToMany(() => AuditLog, auditLog => auditLog.user, {
    cascade: true,
    eager: false
  })
  auditLogs: AuditLog[];

  @OneToMany(() => RiskProfile, riskProfile => riskProfile.user, {
    cascade: true,
    eager: false
  })
  riskProfiles: RiskProfile[];

  // ========================================
  // LIFECYCLE HOOKS
  // ========================================

  @BeforeInsert()
  beforeInsert() {
    this.email = this.email?.toLowerCase();
    this.username = this.username?.toLowerCase();
    this.createdAt = new Date();
    this.updatedAt = new Date();
  }

  @BeforeUpdate()
  beforeUpdate() {
    this.email = this.email?.toLowerCase();
    this.username = this.username?.toLowerCase();
    this.updatedAt = new Date();
  }

  // ========================================
  // BUSINESS METHODS
  // ========================================

  /**
   * Verificar se o usuário está ativo e não bloqueado
   */
  isActiveAndUnlocked(): boolean {
    if (!this.isActive) return false;
    if (this.isLocked && this.lockedUntil && this.lockedUntil > new Date()) {
      return false;
    }
    return true;
  }

  /**
   * Incrementar tentativas de login falhadas
   */
  incrementFailedLoginAttempts(): void {
    this.failedLoginAttempts += 1;
    
    // Bloquear após 5 tentativas falhadas
    if (this.failedLoginAttempts >= 5) {
      this.isLocked = true;
      this.lockedUntil = new Date(Date.now() + 30 * 60 * 1000); // 30 minutos
    }
  }

  /**
   * Resetar tentativas de login falhadas
   */
  resetFailedLoginAttempts(): void {
    this.failedLoginAttempts = 0;
    this.isLocked = false;
    this.lockedUntil = null;
  }

  /**
   * Atualizar último login
   */
  updateLastLogin(): void {
    this.lastLoginAt = new Date();
    this.resetFailedLoginAttempts();
  }

  /**
   * Verificar se o usuário pode fazer login
   */
  canLogin(): boolean {
    return this.isActiveAndUnlocked() && this.isVerified;
  }

  /**
   * Obter informações básicas do usuário (sem dados sensíveis)
   */
  getPublicInfo() {
    return {
      id: this.id,
      username: this.username,
      displayName: this.displayName,
      isActive: this.isActive,
      isVerified: this.isVerified,
      lastLoginAt: this.lastLoginAt,
      createdAt: this.createdAt,
      locale: this.locale,
      timezone: this.timezone
    };
  }

  /**
   * Verificar se o usuário tem credenciais ativas
   */
  hasActiveCredentials(): boolean {
    return this.credentials?.some(credential => credential.isActive) || false;
  }

  /**
   * Obter número de credenciais ativas
   */
  getActiveCredentialCount(): number {
    return this.credentials?.filter(credential => credential.isActive).length || 0;
  }

  /**
   * Verificar se o usuário tem sessões ativas
   */
  hasActiveSessions(): boolean {
    return this.sessions?.some(session => 
      session.isActive && session.expiresAt > new Date()
    ) || false;
  }

  /**
   * Obter número de sessões ativas
   */
  getActiveSessionCount(): number {
    return this.sessions?.filter(session => 
      session.isActive && session.expiresAt > new Date()
    ).length || 0;
  }

  /**
   * Soft delete do usuário
   */
  softDelete(): void {
    this.isActive = false;
    this.deletedAt = new Date();
    this.email = `deleted_${this.id}@deleted.local`;
    this.username = `deleted_${this.id}`;
  }

  /**
   * Restaurar usuário soft deleted
   */
  restore(): void {
    this.deletedAt = null;
    // Email e username precisam ser restaurados manualmente
  }

  /**
   * Verificar se o usuário foi soft deleted
   */
  isDeleted(): boolean {
    return this.deletedAt !== null;
  }

  /**
   * Atualizar preferências do usuário
   */
  updatePreferences(preferences: Record<string, any>): void {
    this.preferences = {
      ...this.preferences,
      ...preferences
    };
  }

  /**
   * Obter preferência específica
   */
  getPreference(key: string, defaultValue?: any): any {
    return this.preferences?.[key] ?? defaultValue;
  }

  /**
   * Verificar se o usuário pertence ao tenant
   */
  belongsToTenant(tenantId: string): boolean {
    return this.tenantId === tenantId;
  }

  /**
   * Validar dados do usuário
   */
  validate(): string[] {
    const errors: string[] = [];

    if (!this.email || !this.email.includes('@')) {
      errors.push('Invalid email format');
    }

    if (!this.username || this.username.length < 3) {
      errors.push('Username must be at least 3 characters');
    }

    if (!this.displayName || this.displayName.length < 2) {
      errors.push('Display name must be at least 2 characters');
    }

    if (!this.tenantId) {
      errors.push('Tenant ID is required');
    }

    return errors;
  }

  /**
   * Serializar para JSON (excluindo campos sensíveis)
   */
  toJSON() {
    const { metadata, deletedAt, ...publicData } = this;
    return publicData;
  }
}