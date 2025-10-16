/**
 * 📋 AUDIT LOG ENTITY - INNOVABIZ IAM
 * Entidade de log de auditoria com compliance avançado
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: SOX, GDPR/LGPD, PCI DSS 4.0, ISO 27001, NIST Cybersecurity Framework
 * Standards: IPPF, IIA Standards, COSO ERM, Basel III
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
import { IsUUID, IsString, IsEnum, IsOptional, IsIP, IsDateString, IsNumber } from 'class-validator';
import { createHash } from 'crypto';

import { User } from './User.entity';

export enum AuditEventType {
  // User Management
  USER_CREATED = 'user_created',
  USER_UPDATED = 'user_updated',
  USER_DELETED = 'user_deleted',
  USER_ACTIVATED = 'user_activated',
  USER_DEACTIVATED = 'user_deactivated',
  USER_LOCKED = 'user_locked',
  USER_UNLOCKED = 'user_unlocked',
  
  // Authentication Events
  LOGIN_SUCCESS = 'login_success',
  LOGIN_FAILED = 'login_failed',
  LOGOUT = 'logout',
  SESSION_CREATED = 'session_created',
  SESSION_EXPIRED = 'session_expired',
  SESSION_REVOKED = 'session_revoked',
  
  // WebAuthn Events
  WEBAUTHN_REGISTRATION_INITIATED = 'webauthn_registration_initiated',
  WEBAUTHN_REGISTRATION_SUCCESS = 'webauthn_registration_success',
  WEBAUTHN_REGISTRATION_FAILED = 'webauthn_registration_failed',
  WEBAUTHN_AUTHENTICATION_INITIATED = 'webauthn_authentication_initiated',
  WEBAUTHN_AUTHENTICATION_SUCCESS = 'webauthn_authentication_success',
  WEBAUTHN_AUTHENTICATION_FAILED = 'webauthn_authentication_failed',
  
  // Credential Management
  CREDENTIAL_CREATED = 'credential_created',
  CREDENTIAL_UPDATED = 'credential_updated',
  CREDENTIAL_REVOKED = 'credential_revoked',
  CREDENTIAL_COMPROMISED = 'credential_compromised',
  
  // Security Events
  SECURITY_VIOLATION = 'security_violation',
  SUSPICIOUS_ACTIVITY = 'suspicious_activity',
  BRUTE_FORCE_ATTEMPT = 'brute_force_attempt',
  ACCOUNT_LOCKOUT = 'account_lockout',
  PRIVILEGE_ESCALATION = 'privilege_escalation',
  
  // Risk Assessment
  RISK_ASSESSMENT_HIGH = 'risk_assessment_high',
  RISK_ASSESSMENT_CRITICAL = 'risk_assessment_critical',
  ANOMALY_DETECTED = 'anomaly_detected',
  
  // Compliance Events
  DATA_ACCESS = 'data_access',
  DATA_EXPORT = 'data_export',
  DATA_DELETION = 'data_deletion',
  CONSENT_GRANTED = 'consent_granted',
  CONSENT_REVOKED = 'consent_revoked',
  
  // Administrative Events
  ADMIN_ACTION = 'admin_action',
  CONFIGURATION_CHANGED = 'configuration_changed',
  POLICY_UPDATED = 'policy_updated',
  
  // System Events
  SYSTEM_ERROR = 'system_error',
  SYSTEM_WARNING = 'system_warning',
  SYSTEM_INFO = 'system_info'
}

export enum AuditSeverity {
  CRITICAL = 'critical',
  HIGH = 'high',
  MEDIUM = 'medium',
  LOW = 'low',
  INFO = 'info'
}

export enum AuditCategory {
  AUTHENTICATION = 'authentication',
  AUTHORIZATION = 'authorization',
  USER_MANAGEMENT = 'user_management',
  CREDENTIAL_MANAGEMENT = 'credential_management',
  SESSION_MANAGEMENT = 'session_management',
  SECURITY = 'security',
  COMPLIANCE = 'compliance',
  RISK_MANAGEMENT = 'risk_management',
  SYSTEM = 'system',
  ADMINISTRATIVE = 'administrative'
}

export enum ComplianceFramework {
  GDPR = 'gdpr',
  LGPD = 'lgpd',
  PCI_DSS = 'pci_dss',
  SOX = 'sox',
  HIPAA = 'hipaa',
  ISO_27001 = 'iso_27001',
  NIST = 'nist',
  BASEL_III = 'basel_iii',
  COSO = 'coso'
}

@Entity('audit_logs')
@Index(['userId', 'tenantId'])
@Index(['eventType'])
@Index(['severity'])
@Index(['category'])
@Index(['timestamp'])
@Index(['ipAddress'])
@Index(['complianceFrameworks'])
@Index(['resourceType', 'resourceId'])
export class AuditLog {
  @PrimaryGeneratedColumn('uuid')
  @IsUUID()
  id: string;

  @Column({ type: 'uuid', name: 'user_id', nullable: true })
  @IsOptional()
  @IsUUID()
  @Index()
  userId?: string;

  @Column({ type: 'uuid', name: 'tenant_id' })
  @IsUUID()
  @Index()
  tenantId: string;

  @Column({ 
    type: 'enum', 
    enum: AuditEventType,
    name: 'event_type'
  })
  @IsEnum(AuditEventType)
  eventType: AuditEventType;

  @Column({ 
    type: 'enum', 
    enum: AuditSeverity, 
    default: AuditSeverity.INFO 
  })
  @IsEnum(AuditSeverity)
  severity: AuditSeverity;

  @Column({ 
    type: 'enum', 
    enum: AuditCategory,
    default: AuditCategory.SYSTEM
  })
  @IsEnum(AuditCategory)
  category: AuditCategory;

  @Column({ type: 'varchar', length: 255 })
  @IsString()
  action: string;

  @Column({ type: 'varchar', length: 255, name: 'resource_type' })
  @IsString()
  resourceType: string;

  @Column({ type: 'varchar', length: 255, nullable: true, name: 'resource_id' })
  @IsOptional()
  @IsString()
  resourceId?: string;

  @Column({ type: 'text', nullable: true })
  @IsOptional()
  @IsString()
  description?: string;

  @Column({ type: 'inet', nullable: true, name: 'ip_address' })
  @IsOptional()
  @IsIP()
  ipAddress?: string;

  @Column({ type: 'varchar', length: 1000, nullable: true, name: 'user_agent' })
  @IsOptional()
  @IsString()
  userAgent?: string;

  @Column({ type: 'varchar', length: 255, nullable: true, name: 'session_id' })
  @IsOptional()
  @IsString()
  sessionId?: string;

  @Column({ type: 'varchar', length: 255, nullable: true, name: 'request_id' })
  @IsOptional()
  @IsString()
  requestId?: string;

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
  @IsNumber()
  riskScore: number;

  @Column({ type: 'boolean', default: false, name: 'is_successful' })
  isSuccessful: boolean;

  @Column({ type: 'varchar', length: 500, nullable: true, name: 'error_message' })
  @IsOptional()
  @IsString()
  errorMessage?: string;

  @Column({ type: 'varchar', length: 100, nullable: true, name: 'error_code' })
  @IsOptional()
  @IsString()
  errorCode?: string;

  @Column({ type: 'jsonb', nullable: true })
  metadata?: Record<string, any>;

  @Column({ type: 'jsonb', nullable: true, name: 'before_state' })
  beforeState?: Record<string, any>;

  @Column({ type: 'jsonb', nullable: true, name: 'after_state' })
  afterState?: Record<string, any>;

  @Column({ 
    type: 'simple-array', 
    nullable: true,
    name: 'compliance_frameworks'
  })
  @IsOptional()
  complianceFrameworks?: ComplianceFramework[];

  @Column({ type: 'jsonb', nullable: true, name: 'compliance_data' })
  complianceData?: {
    regulation?: string;
    article?: string;
    requirement?: string;
    controlId?: string;
    evidenceId?: string;
  };

  @Column({ type: 'varchar', length: 64, name: 'event_hash' })
  @IsString()
  eventHash: string; // Hash para integridade do evento

  @Column({ type: 'varchar', length: 64, name: 'chain_hash', nullable: true })
  @IsOptional()
  @IsString()
  chainHash?: string; // Hash da cadeia para auditoria sequencial

  @Column({ type: 'bigint', name: 'sequence_number' })
  @IsNumber()
  sequenceNumber: number;

  @Column({ type: 'boolean', default: false, name: 'is_sensitive' })
  isSensitive: boolean;

  @Column({ type: 'boolean', default: false, name: 'requires_retention' })
  requiresRetention: boolean;

  @Column({ type: 'timestamp', nullable: true, name: 'retention_until' })
  @IsOptional()
  @IsDateString()
  retentionUntil?: Date;

  @CreateDateColumn({ name: 'timestamp' })
  timestamp: Date;

  // ========================================
  // RELATIONSHIPS
  // ========================================

  @ManyToOne(() => User, user => user.auditLogs, {
    onDelete: 'SET NULL',
    eager: false
  })
  @JoinColumn({ name: 'user_id' })
  user?: User;

  // ========================================
  // LIFECYCLE HOOKS
  // ========================================

  @BeforeInsert()
  beforeInsert() {
    this.generateEventHash();
    this.setSequenceNumber();
    this.setRetentionPolicy();
    this.timestamp = new Date();
  }

  // ========================================
  // BUSINESS METHODS
  // ========================================

  /**
   * Gerar hash do evento para integridade
   */
  private generateEventHash(): void {
    const eventData = {
      userId: this.userId,
      tenantId: this.tenantId,
      eventType: this.eventType,
      action: this.action,
      resourceType: this.resourceType,
      resourceId: this.resourceId,
      timestamp: this.timestamp?.toISOString() || new Date().toISOString(),
      metadata: this.metadata
    };
    
    this.eventHash = createHash('sha256')
      .update(JSON.stringify(eventData))
      .digest('hex');
  }

  /**
   * Definir número sequencial do evento
   */
  private setSequenceNumber(): void {
    // Em implementação real, seria obtido de um contador global
    this.sequenceNumber = Date.now();
  }

  /**
   * Definir política de retenção baseada no tipo de evento
   */
  private setRetentionPolicy(): void {
    const retentionPolicies = {
      [AuditEventType.LOGIN_SUCCESS]: 90, // 90 dias
      [AuditEventType.LOGIN_FAILED]: 365, // 1 ano
      [AuditEventType.SECURITY_VIOLATION]: 2555, // 7 anos
      [AuditEventType.DATA_ACCESS]: 2555, // 7 anos (compliance)
      [AuditEventType.ADMIN_ACTION]: 2555, // 7 anos
      [AuditEventType.CONFIGURATION_CHANGED]: 2555, // 7 anos
    };

    const retentionDays = retentionPolicies[this.eventType] || 365;
    this.retentionUntil = new Date(Date.now() + (retentionDays * 24 * 60 * 60 * 1000));
    this.requiresRetention = retentionDays > 365;
  }

  /**
   * Verificar se o evento é crítico para segurança
   */
  isSecurityCritical(): boolean {
    const criticalEvents = [
      AuditEventType.SECURITY_VIOLATION,
      AuditEventType.BRUTE_FORCE_ATTEMPT,
      AuditEventType.PRIVILEGE_ESCALATION,
      AuditEventType.CREDENTIAL_COMPROMISED,
      AuditEventType.RISK_ASSESSMENT_CRITICAL
    ];
    
    return criticalEvents.includes(this.eventType) || 
           this.severity === AuditSeverity.CRITICAL;
  }

  /**
   * Verificar se requer notificação imediata
   */
  requiresImmediateNotification(): boolean {
    return this.isSecurityCritical() || 
           this.severity === AuditSeverity.CRITICAL ||
           this.severity === AuditSeverity.HIGH;
  }

  /**
   * Obter frameworks de compliance aplicáveis
   */
  getApplicableComplianceFrameworks(): ComplianceFramework[] {
    if (this.complianceFrameworks) {
      return this.complianceFrameworks;
    }

    // Auto-detectar frameworks baseado no tipo de evento
    const frameworks: ComplianceFramework[] = [];
    
    if (this.category === AuditCategory.AUTHENTICATION || 
        this.category === AuditCategory.AUTHORIZATION) {
      frameworks.push(ComplianceFramework.NIST, ComplianceFramework.ISO_27001);
    }
    
    if (this.eventType.includes('DATA_')) {
      frameworks.push(ComplianceFramework.GDPR, ComplianceFramework.LGPD);
    }
    
    if (this.category === AuditCategory.SECURITY) {
      frameworks.push(ComplianceFramework.PCI_DSS, ComplianceFramework.ISO_27001);
    }
    
    return frameworks;
  }

  /**
   * Verificar se o evento está dentro do período de retenção
   */
  isWithinRetentionPeriod(): boolean {
    if (!this.retentionUntil) return true;
    return new Date() <= this.retentionUntil;
  }

  /**
   * Marcar como sensível
   */
  markSensitive(reason?: string): void {
    this.isSensitive = true;
    if (reason) {
      this.metadata = {
        ...this.metadata,
        sensitiveReason: reason,
        markedSensitiveAt: new Date().toISOString()
      };
    }
  }

  /**
   * Adicionar contexto de compliance
   */
  addComplianceContext(
    framework: ComplianceFramework,
    regulation: string,
    article?: string,
    requirement?: string,
    controlId?: string
  ): void {
    if (!this.complianceFrameworks) {
      this.complianceFrameworks = [];
    }
    
    if (!this.complianceFrameworks.includes(framework)) {
      this.complianceFrameworks.push(framework);
    }
    
    this.complianceData = {
      ...this.complianceData,
      regulation,
      article,
      requirement,
      controlId
    };
  }

  /**
   * Obter resumo do evento para relatórios
   */
  getSummary(): string {
    const parts = [
      this.action,
      this.resourceType,
      this.resourceId ? `(${this.resourceId})` : '',
      this.isSuccessful ? 'SUCCESS' : 'FAILED'
    ].filter(Boolean);
    
    return parts.join(' ');
  }

  /**
   * Obter contexto geográfico
   */
  getGeographicContext(): string {
    const parts = [this.city, this.region, this.country].filter(Boolean);
    return parts.join(', ') || 'Unknown Location';
  }

  /**
   * Verificar integridade do evento
   */
  verifyIntegrity(): boolean {
    const currentHash = this.eventHash;
    this.generateEventHash();
    const recalculatedHash = this.eventHash;
    this.eventHash = currentHash;
    
    return currentHash === recalculatedHash;
  }

  /**
   * Criar evento de auditoria padrão
   */
  static createAuditEvent(
    tenantId: string,
    eventType: AuditEventType,
    action: string,
    resourceType: string,
    options: {
      userId?: string;
      resourceId?: string;
      description?: string;
      severity?: AuditSeverity;
      category?: AuditCategory;
      ipAddress?: string;
      userAgent?: string;
      sessionId?: string;
      metadata?: Record<string, any>;
      beforeState?: Record<string, any>;
      afterState?: Record<string, any>;
      isSuccessful?: boolean;
      errorMessage?: string;
      errorCode?: string;
    } = {}
  ): AuditLog {
    const auditLog = new AuditLog();
    
    auditLog.tenantId = tenantId;
    auditLog.eventType = eventType;
    auditLog.action = action;
    auditLog.resourceType = resourceType;
    auditLog.userId = options.userId;
    auditLog.resourceId = options.resourceId;
    auditLog.description = options.description;
    auditLog.severity = options.severity || AuditSeverity.INFO;
    auditLog.category = options.category || AuditCategory.SYSTEM;
    auditLog.ipAddress = options.ipAddress;
    auditLog.userAgent = options.userAgent;
    auditLog.sessionId = options.sessionId;
    auditLog.metadata = options.metadata;
    auditLog.beforeState = options.beforeState;
    auditLog.afterState = options.afterState;
    auditLog.isSuccessful = options.isSuccessful ?? true;
    auditLog.errorMessage = options.errorMessage;
    auditLog.errorCode = options.errorCode;
    auditLog.riskScore = 0;
    
    return auditLog;
  }

  /**
   * Criar evento de segurança
   */
  static createSecurityEvent(
    tenantId: string,
    userId: string,
    eventType: AuditEventType,
    action: string,
    severity: AuditSeverity,
    description: string,
    metadata?: Record<string, any>
  ): AuditLog {
    return AuditLog.createAuditEvent(tenantId, eventType, action, 'Security', {
      userId,
      description,
      severity,
      category: AuditCategory.SECURITY,
      metadata,
      isSuccessful: false
    });
  }

  /**
   * Criar evento de compliance
   */
  static createComplianceEvent(
    tenantId: string,
    userId: string,
    eventType: AuditEventType,
    action: string,
    resourceType: string,
    framework: ComplianceFramework,
    regulation: string,
    metadata?: Record<string, any>
  ): AuditLog {
    const auditLog = AuditLog.createAuditEvent(tenantId, eventType, action, resourceType, {
      userId,
      category: AuditCategory.COMPLIANCE,
      severity: AuditSeverity.MEDIUM,
      metadata
    });
    
    auditLog.addComplianceContext(framework, regulation);
    return auditLog;
  }

  /**
   * Serializar para JSON (excluindo dados sensíveis se necessário)
   */
  toJSON() {
    if (this.isSensitive) {
      const { metadata, beforeState, afterState, ...publicData } = this;
      return {
        ...publicData,
        metadata: '[REDACTED]',
        beforeState: '[REDACTED]',
        afterState: '[REDACTED]'
      };
    }
    
    return {
      id: this.id,
      eventType: this.eventType,
      severity: this.severity,
      category: this.category,
      action: this.action,
      resourceType: this.resourceType,
      resourceId: this.resourceId,
      description: this.description,
      isSuccessful: this.isSuccessful,
      timestamp: this.timestamp,
      riskScore: this.riskScore,
      geographicContext: this.getGeographicContext()
    };
  }

  /**
   * Obter dados para relatório de compliance
   */
  getComplianceReport() {
    return {
      eventId: this.id,
      timestamp: this.timestamp,
      eventType: this.eventType,
      severity: this.severity,
      category: this.category,
      action: this.action,
      resourceType: this.resourceType,
      userId: this.userId,
      isSuccessful: this.isSuccessful,
      complianceFrameworks: this.complianceFrameworks,
      complianceData: this.complianceData,
      retentionUntil: this.retentionUntil,
      eventHash: this.eventHash
    };
  }
}