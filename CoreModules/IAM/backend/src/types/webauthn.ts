/**
 * ============================================================================
 * INNOVABIZ IAM - WebAuthn Types
 * ============================================================================
 * Version: 1.0.0
 * Date: 31/07/2025
 * Author: Equipe de Segurança INNOVABIZ
 * Classification: Confidencial - Interno
 * 
 * Description: Definições de tipos TypeScript para WebAuthn/FIDO2
 * Standards: W3C WebAuthn Level 3, FIDO2 CTAP2.1, NIST SP 800-63B
 * Compliance: PCI DSS 4.0, GDPR/LGPD, PSD2, ISO 27001
 * ============================================================================
 */

import {
  AuthenticatorTransportFuture,
  UserVerificationRequirement,
  AttestationConveyancePreference,
  AuthenticatorAttachment,
  ResidentKeyRequirement
} from '@simplewebauthn/typescript-types';

/**
 * Status de uma credencial WebAuthn
 */
export type CredentialStatus = 'active' | 'suspended' | 'deleted' | 'expired';

/**
 * Níveis de garantia de autenticação (Authentication Assurance Level)
 */
export type AAL = 'AAL1' | 'AAL2' | 'AAL3';

/**
 * Tipos de autenticador
 */
export type AuthenticatorType = 'platform' | 'cross-platform';

/**
 * Contexto de execução WebAuthn
 */
export interface WebAuthnContext {
  userId?: string;
  tenantId?: string;
  userEmail?: string;
  userDisplayName?: string;
  regionCode: string;
  origin: string;
  ipAddress: string;
  userAgent: string;
  correlationId: string;
  sessionId?: string;
  deviceFingerprint?: string;
  geolocation?: {
    latitude: number;
    longitude: number;
    accuracy: number;
  };
}

/**
 * Credencial WebAuthn armazenada
 */
export interface WebAuthnCredential {
  id: string;
  userId: string;
  tenantId: string;
  credentialId: string;
  publicKey: Buffer;
  signCount: number;
  aaguid: string;
  attestationFormat: string;
  attestationData: AttestationData;
  userVerified: boolean;
  backupEligible: boolean;
  backupState: boolean;
  transports: AuthenticatorTransportFuture[];
  authenticatorType: AuthenticatorType;
  deviceType: string;
  friendlyName: string;
  complianceLevel: AAL;
  riskScore: number;
  status: CredentialStatus;
  createdAt: Date;
  updatedAt: Date;
  lastUsedAt?: Date;
  lastUsedIp?: string;
  lastUsedUserAgent?: string;
  suspendedAt?: Date;
  suspensionReason?: string;
  suspensionDetails?: string;
  deletedAt?: Date;
  deletionReason?: string;
  reactivatedAt?: Date;
  reactivationReason?: string;
  metadata?: CredentialMetadata;
}

/**
 * Dados de attestation
 */
export interface AttestationData {
  fmt: string;
  attStmt: any;
  aaguid: string;
  credentialPublicKey?: Buffer;
  credentialId?: Buffer;
}

/**
 * Metadados da credencial
 */
export interface CredentialMetadata {
  registrationContext?: WebAuthnContext;
  userAgent?: string;
  ipAddress?: string;
  registeredAt?: string;
  deviceInfo?: {
    platform?: string;
    browser?: string;
    version?: string;
  };
  securityKeys?: {
    isSecurityKey?: boolean;
    vendor?: string;
    model?: string;
    firmwareVersion?: string;
  };
  biometrics?: {
    hasBiometrics?: boolean;
    biometricType?: string[];
  };
  [key: string]: any;
}

/**
 * Solicitação para armazenar credencial
 */
export interface StoreCredentialRequest {
  userId: string;
  tenantId: string;
  credentialId: string;
  publicKey: Buffer;
  signCount: number;
  aaguid: string;
  attestationFormat: string;
  attestationData: AttestationData;
  userVerified: boolean;
  backupEligible: boolean;
  backupState: boolean;
  transports: AuthenticatorTransportFuture[];
  authenticatorType: AuthenticatorType;
  deviceType: string;
  friendlyName: string;
  complianceLevel: AAL;
  riskScore: number;
  metadata?: CredentialMetadata;
}

/**
 * Filtros para busca de credenciais
 */
export interface CredentialFilter {
  status?: CredentialStatus;
  authenticatorType?: AuthenticatorType;
  complianceLevel?: AAL;
  deviceType?: string;
  limit?: number;
  offset?: number;
}

/**
 * Atualização de uso de credencial
 */
export interface CredentialUsageUpdate {
  credentialId: string;
  signCount: number;
  ipAddress: string;
  userAgent: string;
  riskScore?: number;
  usedAt: Date;
}

/**
 * Opções para geração de registro
 */
export interface RegistrationOptionsRequest {
  username?: string;
  displayName?: string;
  attestation?: AttestationConveyancePreference;
  authenticatorSelection?: {
    authenticatorAttachment?: AuthenticatorAttachment;
    userVerification?: UserVerificationRequirement;
    residentKey?: ResidentKeyRequirement;
  };
  excludeCredentials?: boolean;
  timeout?: number;
}

/**
 * Opções para geração de autenticação
 */
export interface AuthenticationOptionsRequest {
  userVerification?: UserVerificationRequirement;
  allowCredentials?: string[];
  timeout?: number;
}

/**
 * Resultado do registro de credencial
 */
export interface CredentialRegistrationResult {
  credentialId: string;
  verified: boolean;
  attestationFormat: string;
  userVerified: boolean;
  complianceLevel: AAL;
  deviceType: AuthenticatorType;
  friendlyName: string;
}

/**
 * Resultado da autenticação
 */
export interface AuthenticationResult {
  userId: string;
  tenantId: string;
  credentialId: string;
  verified: boolean;
  userVerified: boolean;
  riskScore: number;
  authenticationLevel: AAL;
  deviceType: AuthenticatorType;
  friendlyName: string;
  lastUsed: Date;
}

/**
 * Evento de auditoria WebAuthn
 */
export interface WebAuthnAuditEvent {
  type: string;
  userId: string;
  tenantId: string;
  credentialId?: string;
  result: 'success' | 'failure' | 'warning';
  errorCode?: string;
  errorMessage?: string;
  clientData?: string;
  authenticatorData?: string;
  signature?: string;
  userVerified?: boolean;
  signCount?: number;
  riskScore?: number;
  complianceLevel?: AAL;
  ipAddress: string;
  userAgent: string;
  correlationId: string;
  metadata?: Record<string, any>;
}

/**
 * Erro específico de WebAuthn
 */
export class WebAuthnError extends Error {
  public readonly code: string;
  public readonly details?: any;

  constructor(code: string, message: string, details?: any) {
    super(message);
    this.name = 'WebAuthnError';
    this.code = code;
    this.details = details;
  }
}

/**
 * Erro de anomalia de sign count
 */
export class SignCountAnomalyError extends WebAuthnError {
  public readonly expectedCount: number;
  public readonly receivedCount: number;

  constructor(message: string, expectedCount: number, receivedCount: number) {
    super('SIGN_COUNT_ANOMALY', message);
    this.name = 'SignCountAnomalyError';
    this.expectedCount = expectedCount;
    this.receivedCount = receivedCount;
  }
}

/**
 * Resposta de API padronizada
 */
export interface WebAuthnAPIResponse<T = any> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: any;
  };
  metadata: {
    correlationId: string;
    timestamp: string;
    version: string;
  };
}