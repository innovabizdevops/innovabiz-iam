/**
 * ============================================================================
 * INNOVABIZ IAM - WebAuthn Service
 * ============================================================================
 * Version: 1.0.0
 * Date: 31/07/2025
 * Author: Equipe de Segurança INNOVABIZ
 * Classification: Confidencial - Interno
 * 
 * Description: Serviço principal para operações WebAuthn/FIDO2
 * Standards: W3C WebAuthn Level 3, FIDO2 CTAP2.1, NIST SP 800-63B
 * Compliance: PCI DSS 4.0, GDPR/LGPD, PSD2, ISO 27001
 * ============================================================================
 */

import {
  generateRegistrationOptions,
  verifyRegistrationResponse,
  generateAuthenticationOptions,
  verifyAuthenticationResponse,
  VerifiedRegistrationResponse,
  VerifiedAuthenticationResponse
} from '@simplewebauthn/server';
import {
  RegistrationResponseJSON,
  AuthenticationResponseJSON,
  PublicKeyCredentialCreationOptionsJSON,
  PublicKeyCredentialRequestOptionsJSON,
  AuthenticatorTransportFuture
} from '@simplewebauthn/typescript-types';
import { Logger } from 'winston';
import { Redis } from 'ioredis';
import { Pool } from 'pg';

import { CredentialService } from './CredentialService';
import { AttestationService } from './AttestationService';
import { RiskAssessmentService } from './RiskAssessmentService';
import { AuditService } from './AuditService';
import { webauthnConfig } from '../config/webauthn';
import { webauthnMetrics } from '../metrics/webauthn';
import {
  WebAuthnCredential,
  AuthenticationResult,
  CredentialRegistrationResult,
  RegistrationOptionsRequest,
  AuthenticationOptionsRequest,
  WebAuthnContext,
  SignCountAnomalyError,
  WebAuthnError
} from '../types/webauthn';

/**
 * Serviço principal para operações WebAuthn/FIDO2
 * Implementa os padrões W3C WebAuthn Level 3 e FIDO2 CTAP2.1
 */
export class WebAuthnService {
  private readonly logger: Logger;
  private readonly redis: Redis;
  private readonly db: Pool;
  private readonly credentialService: CredentialService;
  private readonly attestationService: AttestationService;
  private readonly riskService: RiskAssessmentService;
  private readonly auditService: AuditService;

  constructor(
    logger: Logger,
    redis: Redis,
    db: Pool,
    credentialService: CredentialService,
    attestationService: AttestationService,
    riskService: RiskAssessmentService,
    auditService: AuditService
  ) {
    this.logger = logger;
    this.redis = redis;
    this.db = db;
    this.credentialService = credentialService;
    this.attestationService = attestationService;
    this.riskService = riskService;
    this.auditService = auditService;
  }

  /**
   * Gera opções para registro de nova credencial WebAuthn
   */
  async generateRegistrationOptions(
    context: WebAuthnContext,
    options: RegistrationOptionsRequest
  ): Promise<PublicKeyCredentialCreationOptionsJSON> {
    const timer = webauthnMetrics.verificationDuration.startTimer({ operation_type: 'registration_options' });
    
    try {
      this.logger.info('Generating WebAuthn registration options', {
        userId: context.userId,
        tenantId: context.tenantId,
        correlationId: context.correlationId
      });

      // Buscar credenciais existentes para exclusão
      const existingCredentials = await this.credentialService.getUserCredentials(
        context.userId,
        context.tenantId
      );

      const excludeCredentials = existingCredentials.map(cred => ({
        id: Buffer.from(cred.credentialId, 'base64url'),
        type: 'public-key' as const,
        transports: cred.transports as AuthenticatorTransportFuture[]
      }));

      // Obter configurações regionais
      const regionalConfig = await this.getRegionalConfig(context.regionCode);

      // Gerar opções de registro
      const registrationOptions = await generateRegistrationOptions({
        rpName: webauthnConfig.rpName,
        rpID: webauthnConfig.rpID,
        userID: Buffer.from(context.userId),
        userName: options.username || context.userEmail || context.userId,
        userDisplayName: options.displayName || context.userDisplayName || options.username,
        timeout: regionalConfig.registrationTimeoutMs || webauthnConfig.timeout,
        attestationType: options.attestation || regionalConfig.attestationRequirement || webauthnConfig.attestation,
        excludeCredentials,
        authenticatorSelection: {
          authenticatorAttachment: options.authenticatorSelection?.authenticatorAttachment,
          userVerification: options.authenticatorSelection?.userVerification || 
                           (regionalConfig.requireUserVerification ? 'required' : 'preferred'),
          residentKey: options.authenticatorSelection?.residentKey || 
                      (regionalConfig.requireResidentKey ? 'required' : 'preferred')
        },
        supportedAlgorithmIDs: [-7, -257, -8, -37, -38, -39] // ES256, RS256, EdDSA, PS256, PS384, PS512
      });

      // Armazenar challenge temporariamente
      await this.storeChallenge(
        context.userId,
        context.tenantId,
        'registration',
        registrationOptions.challenge,
        context
      );

      // Métricas
      webauthnMetrics.registrationAttempts.inc({
        tenant_id: context.tenantId,
        result: 'options_generated',
        attestation_format: options.attestation || 'default'
      });

      this.logger.info('WebAuthn registration options generated successfully', {
        userId: context.userId,
        tenantId: context.tenantId,
        correlationId: context.correlationId,
        excludedCredentials: excludeCredentials.length
      });

      return registrationOptions;

    } catch (error) {
      this.logger.error('Failed to generate WebAuthn registration options', {
        userId: context.userId,
        tenantId: context.tenantId,
        correlationId: context.correlationId,
        error: error.message
      });

      webauthnMetrics.registrationAttempts.inc({
        tenant_id: context.tenantId,
        result: 'options_failed',
        attestation_format: 'error'
      });

      throw new WebAuthnError('REGISTRATION_OPTIONS_FAILED', 'Failed to generate registration options', error);
    } finally {
      timer();
    }
  }

  /**
   * Verifica e registra uma credencial WebAuthn
   */
  async verifyRegistration(
    context: WebAuthnContext,
    response: RegistrationResponseJSON
  ): Promise<CredentialRegistrationResult> {
    const timer = webauthnMetrics.verificationDuration.startTimer({ operation_type: 'registration_verify' });
    
    try {
      this.logger.info('Verifying WebAuthn registration', {
        userId: context.userId,
        tenantId: context.tenantId,
        correlationId: context.correlationId,
        credentialId: response.id
      });

      // Recuperar challenge
      const expectedChallenge = await this.getChallenge(
        context.userId,
        context.tenantId,
        'registration'
      );

      if (!expectedChallenge) {
        throw new WebAuthnError('CHALLENGE_NOT_FOUND', 'Challenge not found or expired');
      }

      // Obter configurações regionais
      const regionalConfig = await this.getRegionalConfig(context.regionCode);

      // Verificar resposta de registro
      const verification = await verifyRegistrationResponse({
        response,
        expectedChallenge,
        expectedOrigin: this.getExpectedOrigins(context),
        expectedRPID: webauthnConfig.rpID,
        requireUserVerification: regionalConfig.requireUserVerification || false
      });

      if (!verification.verified || !verification.registrationInfo) {
        await this.auditService.logWebAuthnEvent({
          type: 'registration',
          userId: context.userId,
          tenantId: context.tenantId,
          credentialId: response.id,
          result: 'failure',
          errorCode: 'REGISTRATION_VERIFICATION_FAILED',
          ipAddress: context.ipAddress,
          userAgent: context.userAgent,
          correlationId: context.correlationId
        });

        webauthnMetrics.registrationAttempts.inc({
          tenant_id: context.tenantId,
          result: 'verification_failed',
          attestation_format: 'unknown'
        });

        throw new WebAuthnError('REGISTRATION_VERIFICATION_FAILED', 'Registration verification failed');
      }

      const { registrationInfo } = verification;

      // Validar attestation se necessário
      if (registrationInfo.attestationObject && regionalConfig.requireAttestationVerification) {
        await this.attestationService.validateAttestation(
          registrationInfo.attestationObject,
          registrationInfo.aaguid,
          context
        );
      }

      // Verificar limites de credenciais por usuário
      const userCredentialCount = await this.credentialService.getUserCredentialCount(
        context.userId,
        context.tenantId
      );

      if (userCredentialCount >= (regionalConfig.maxCredentialsPerUser || 10)) {
        throw new WebAuthnError('MAX_CREDENTIALS_EXCEEDED', 'Maximum credentials per user exceeded');
      }

      // Armazenar credencial
      const credential = await this.credentialService.storeCredential({
        userId: context.userId,
        tenantId: context.tenantId,
        credentialId: Buffer.from(registrationInfo.credentialID).toString('base64url'),
        publicKey: registrationInfo.credentialPublicKey,
        signCount: registrationInfo.counter,
        aaguid: registrationInfo.aaguid,
        attestationFormat: registrationInfo.fmt,
        attestationData: {
          fmt: registrationInfo.fmt,
          attStmt: registrationInfo.attestationObject,
          aaguid: registrationInfo.aaguid
        },
        userVerified: registrationInfo.userVerified,
        backupEligible: registrationInfo.credentialBackedUp,
        backupState: registrationInfo.credentialDeviceType === 'multiDevice',
        transports: response.response.transports || [],
        authenticatorType: registrationInfo.credentialDeviceType === 'multiDevice' ? 'cross-platform' : 'platform',
        deviceType: this.determineDeviceType(registrationInfo.aaguid),
        friendlyName: options.displayName || 'WebAuthn Credential',
        complianceLevel: this.determineAAL(registrationInfo, regionalConfig),
        riskScore: await this.riskService.assessRegistrationRisk(context, registrationInfo),
        metadata: {
          registrationContext: context,
          userAgent: context.userAgent,
          ipAddress: context.ipAddress,
          registeredAt: new Date().toISOString()
        }
      });

      // Limpar challenge
      await this.clearChallenge(context.userId, context.tenantId, 'registration');

      // Registrar evento de auditoria
      await this.auditService.logWebAuthnEvent({
        type: 'registration',
        userId: context.userId,
        tenantId: context.tenantId,
        credentialId: credential.id,
        result: 'success',
        clientData: response.response.clientDataJSON,
        userVerified: registrationInfo.userVerified,
        complianceLevel: credential.complianceLevel,
        ipAddress: context.ipAddress,
        userAgent: context.userAgent,
        correlationId: context.correlationId,
        metadata: {
          attestationFormat: registrationInfo.fmt,
          aaguid: registrationInfo.aaguid,
          deviceType: credential.authenticatorType
        }
      });

      // Métricas
      webauthnMetrics.registrationAttempts.inc({
        tenant_id: context.tenantId,
        result: 'success',
        attestation_format: registrationInfo.fmt
      });

      this.logger.info('WebAuthn registration completed successfully', {
        userId: context.userId,
        tenantId: context.tenantId,
        correlationId: context.correlationId,
        credentialId: credential.credentialId,
        attestationFormat: registrationInfo.fmt,
        complianceLevel: credential.complianceLevel
      });

      return {
        credentialId: credential.credentialId,
        verified: true,
        attestationFormat: registrationInfo.fmt,
        userVerified: registrationInfo.userVerified,
        complianceLevel: credential.complianceLevel,
        deviceType: credential.authenticatorType,
        friendlyName: credential.friendlyName
      };

    } catch (error) {
      this.logger.error('WebAuthn registration verification failed', {
        userId: context.userId,
        tenantId: context.tenantId,
        correlationId: context.correlationId,
        error: error.message
      });

      webauthnMetrics.registrationAttempts.inc({
        tenant_id: context.tenantId,
        result: 'failed',
        attestation_format: 'error'
      });

      if (error instanceof WebAuthnError) {
        throw error;
      }

      throw new WebAuthnError('REGISTRATION_VERIFICATION_FAILED', 'Registration verification failed', error);
    } finally {
      timer();
    }
  }

  /**
   * Gera opções para autenticação WebAuthn
   */
  async generateAuthenticationOptions(
    context: WebAuthnContext,
    options: AuthenticationOptionsRequest = {}
  ): Promise<PublicKeyCredentialRequestOptionsJSON> {
    const timer = webauthnMetrics.verificationDuration.startTimer({ operation_type: 'authentication_options' });
    
    try {
      this.logger.info('Generating WebAuthn authentication options', {
        userId: context.userId,
        tenantId: context.tenantId,
        correlationId: context.correlationId,
        usernameless: !context.userId
      });

      let allowCredentials: Array<{
        id: Buffer;
        type: 'public-key';
        transports: AuthenticatorTransportFuture[];
      }> = [];

      // Se usuário especificado, buscar suas credenciais
      if (context.userId && context.tenantId) {
        const credentials = await this.credentialService.getUserCredentials(
          context.userId,
          context.tenantId
        );
        
        allowCredentials = credentials.map(cred => ({
          id: Buffer.from(cred.credentialId, 'base64url'),
          type: 'public-key' as const,
          transports: cred.transports as AuthenticatorTransportFuture[]
        }));
      }

      // Obter configurações regionais
      const regionalConfig = await this.getRegionalConfig(context.regionCode);

      // Gerar opções de autenticação
      const authenticationOptions = await generateAuthenticationOptions({
        rpID: webauthnConfig.rpID,
        timeout: regionalConfig.authenticationTimeoutMs || webauthnConfig.timeout,
        allowCredentials: allowCredentials.length > 0 ? allowCredentials : undefined,
        userVerification: options.userVerification || 
                         (regionalConfig.requireUserVerification ? 'required' : 'preferred')
      });

      // Armazenar challenge
      const challengeKey = context.userId ? `${context.userId}:${context.tenantId}` : 'usernameless';
      await this.storeChallenge(
        challengeKey,
        'authentication',
        'authentication',
        authenticationOptions.challenge,
        context
      );

      // Métricas
      webauthnMetrics.authenticationAttempts.inc({
        tenant_id: context.tenantId || 'unknown',
        result: 'options_generated',
        aal_level: 'unknown'
      });

      this.logger.info('WebAuthn authentication options generated successfully', {
        userId: context.userId,
        tenantId: context.tenantId,
        correlationId: context.correlationId,
        allowedCredentials: allowCredentials.length
      });

      return authenticationOptions;

    } catch (error) {
      this.logger.error('Failed to generate WebAuthn authentication options', {
        userId: context.userId,
        tenantId: context.tenantId,
        correlationId: context.correlationId,
        error: error.message
      });

      webauthnMetrics.authenticationAttempts.inc({
        tenant_id: context.tenantId || 'unknown',
        result: 'options_failed',
        aal_level: 'error'
      });

      throw new WebAuthnError('AUTHENTICATION_OPTIONS_FAILED', 'Failed to generate authentication options', error);
    } finally {
      timer();
    }
  }

  /**
   * Verifica autenticação WebAuthn
   */
  async verifyAuthentication(
    context: WebAuthnContext,
    response: AuthenticationResponseJSON
  ): Promise<AuthenticationResult> {
    const timer = webauthnMetrics.verificationDuration.startTimer({ operation_type: 'authentication_verify' });
    
    try {
      this.logger.info('Verifying WebAuthn authentication', {
        userId: context.userId,
        tenantId: context.tenantId,
        correlationId: context.correlationId,
        credentialId: response.id
      });

      // Buscar credencial
      const credential = await this.credentialService.getCredentialById(
        Buffer.from(response.id, 'base64url').toString('base64url')
      );

      if (!credential || credential.status !== 'active') {
        await this.auditService.logWebAuthnEvent({
          type: 'authentication_failed',
          userId: context.userId || 'unknown',
          tenantId: context.tenantId || 'unknown',
          credentialId: response.id,
          result: 'failure',
          errorCode: 'CREDENTIAL_NOT_FOUND',
          ipAddress: context.ipAddress,
          userAgent: context.userAgent,
          correlationId: context.correlationId
        });

        throw new WebAuthnError('CREDENTIAL_NOT_FOUND', 'Credential not found or inactive');
      }

      // Atualizar contexto com dados da credencial se não fornecidos
      if (!context.userId) {
        context.userId = credential.userId;
        context.tenantId = credential.tenantId;
      }

      // Recuperar challenge
      const challengeKey = context.userId ? `${context.userId}:${context.tenantId}` : 'usernameless';
      const expectedChallenge = await this.getChallenge(challengeKey, 'authentication', 'authentication');
      
      if (!expectedChallenge) {
        throw new WebAuthnError('CHALLENGE_NOT_FOUND', 'Challenge not found or expired');
      }

      // Verificar resposta de autenticação
      const verification = await verifyAuthenticationResponse({
        response,
        expectedChallenge,
        expectedOrigin: this.getExpectedOrigins(context),
        expectedRPID: webauthnConfig.rpID,
        authenticator: {
          credentialID: Buffer.from(credential.credentialId, 'base64url'),
          credentialPublicKey: credential.publicKey,
          counter: credential.signCount,
          transports: credential.transports as AuthenticatorTransportFuture[]
        },
        requireUserVerification: credential.userVerified
      });

      if (!verification.verified || !verification.authenticationInfo) {
        await this.auditService.logWebAuthnEvent({
          type: 'authentication_failed',
          userId: credential.userId,
          tenantId: credential.tenantId,
          credentialId: credential.id,
          result: 'failure',
          errorCode: 'AUTHENTICATION_VERIFICATION_FAILED',
          ipAddress: context.ipAddress,
          userAgent: context.userAgent,
          correlationId: context.correlationId
        });

        webauthnMetrics.authenticationAttempts.inc({
          tenant_id: credential.tenantId,
          result: 'verification_failed',
          aal_level: credential.complianceLevel
        });

        throw new WebAuthnError('AUTHENTICATION_VERIFICATION_FAILED', 'Authentication verification failed');
      }

      const { authenticationInfo } = verification;

      // Verificar sign count (detecção de clonagem)
      if (authenticationInfo.newCounter <= credential.signCount && credential.signCount > 0) {
        await this.handleSignCountAnomaly(credential, authenticationInfo.newCounter, context);
        throw new SignCountAnomalyError(
          'Sign count anomaly detected - possible credential cloning',
          credential.signCount,
          authenticationInfo.newCounter
        );
      }

      // Avaliação de risco
      const riskScore = await this.riskService.assessAuthenticationRisk({
        userId: credential.userId,
        credentialId: credential.id,
        context,
        authenticationInfo,
        userVerified: authenticationInfo.userVerified
      });

      // Atualizar credencial
      await this.credentialService.updateCredentialUsage(
        credential.id,
        authenticationInfo.newCounter,
        context.ipAddress,
        context.userAgent,
        riskScore
      );

      // Limpar challenge
      await this.clearChallenge(challengeKey, 'authentication', 'authentication');

      // Determinar nível de garantia de autenticação
      const authenticationLevel = this.determineAAL(credential, authenticationInfo.userVerified);

      // Registrar evento de auditoria
      await this.auditService.logWebAuthnEvent({
        type: 'authentication',
        userId: credential.userId,
        tenantId: credential.tenantId,
        credentialId: credential.id,
        result: 'success',
        clientData: response.response.clientDataJSON,
        authenticatorData: response.response.authenticatorData,
        signature: response.response.signature,
        userVerified: authenticationInfo.userVerified,
        signCount: authenticationInfo.newCounter,
        riskScore,
        complianceLevel: authenticationLevel,
        ipAddress: context.ipAddress,
        userAgent: context.userAgent,
        correlationId: context.correlationId
      });

      // Métricas
      webauthnMetrics.authenticationAttempts.inc({
        tenant_id: credential.tenantId,
        result: 'success',
        aal_level: authenticationLevel
      });

      this.logger.info('WebAuthn authentication completed successfully', {
        userId: credential.userId,
        tenantId: credential.tenantId,
        correlationId: context.correlationId,
        credentialId: credential.credentialId,
        authenticationLevel,
        riskScore,
        userVerified: authenticationInfo.userVerified
      });

      return {
        userId: credential.userId,
        tenantId: credential.tenantId,
        credentialId: credential.credentialId,
        verified: true,
        userVerified: authenticationInfo.userVerified,
        riskScore,
        authenticationLevel,
        deviceType: credential.authenticatorType,
        friendlyName: credential.friendlyName,
        lastUsed: new Date()
      };

    } catch (error) {
      this.logger.error('WebAuthn authentication verification failed', {
        userId: context.userId,
        tenantId: context.tenantId,
        correlationId: context.correlationId,
        error: error.message
      });

      webauthnMetrics.authenticationAttempts.inc({
        tenant_id: context.tenantId || 'unknown',
        result: 'failed',
        aal_level: 'error'
      });

      if (error instanceof WebAuthnError || error instanceof SignCountAnomalyError) {
        throw error;
      }

      throw new WebAuthnError('AUTHENTICATION_VERIFICATION_FAILED', 'Authentication verification failed', error);
    } finally {
      timer();
    }
  }

  /**
   * Métodos privados de utilidade
   */

  private async storeChallenge(
    key: string,
    type: string,
    subType: string,
    challenge: string,
    context: WebAuthnContext
  ): Promise<void> {
    const challengeData = {
      challenge,
      userId: context.userId,
      tenantId: context.tenantId,
      origin: context.origin,
      userAgent: context.userAgent,
      ipAddress: context.ipAddress,
      createdAt: new Date().toISOString()
    };

    await this.redis.setex(
      `webauthn:challenge:${type}:${subType}:${key}`,
      300, // 5 minutos
      JSON.stringify(challengeData)
    );
  }

  private async getChallenge(key: string, type: string, subType: string): Promise<string | null> {
    const challengeData = await this.redis.get(`webauthn:challenge:${type}:${subType}:${key}`);
    if (!challengeData) return null;
    
    try {
      const parsed = JSON.parse(challengeData);
      return parsed.challenge;
    } catch {
      return challengeData; // Fallback para formato simples
    }
  }

  private async clearChallenge(key: string, type: string, subType: string): Promise<void> {
    await this.redis.del(`webauthn:challenge:${type}:${subType}:${key}`);
  }

  private getExpectedOrigins(context: WebAuthnContext): string[] {
    const origins = webauthnConfig.origin;
    if (context.origin && !origins.includes(context.origin)) {
      this.logger.warn('Origin not in allowed list', {
        origin: context.origin,
        allowedOrigins: origins,
        correlationId: context.correlationId
      });
    }
    return origins;
  }

  private determineAAL(
    credentialOrInfo: any,
    userVerified: boolean
  ): 'AAL1' | 'AAL2' | 'AAL3' {
    // AAL3: Hardware authenticator com user verification
    if (credentialOrInfo.authenticatorType === 'cross-platform' && userVerified) {
      return 'AAL3';
    }
    
    // AAL2: Multi-factor authentication
    if (userVerified || credentialOrInfo.authenticatorType === 'platform') {
      return 'AAL2';
    }
    
    // AAL1: Single factor
    return 'AAL1';
  }

  private determineDeviceType(aaguid: string): string {
    // Mapear AAGUID para tipos de dispositivo conhecidos
    const deviceTypeMap: Record<string, string> = {
      '2fc0579f-8113-47ea-b116-bb5a8db9202a': 'YubiKey 5 Series',
      'adce0002-35bc-c60a-648b-0b25f1f05503': 'Touch ID',
      '389c9753-1e30-4c14-b321-dc447d4b5d94': 'Face ID',
      '08987058-cadc-4b81-b6e1-30de50dcbe96': 'Windows Hello',
      'bada5566-a7aa-401f-bd96-45619a55120d': 'Android Fingerprint'
    };

    return deviceTypeMap[aaguid] || 'Unknown Device';
  }

  private async getRegionalConfig(regionCode: string): Promise<any> {
    // Implementar cache e busca de configuração regional
    const cacheKey = `webauthn:regional_config:${regionCode}`;
    const cached = await this.redis.get(cacheKey);
    
    if (cached) {
      return JSON.parse(cached);
    }

    // Buscar do banco de dados
    const result = await this.db.query(
      'SELECT * FROM webauthn_regional_config WHERE region_code = $1',
      [regionCode]
    );

    const config = result.rows[0] || {
      requireUserVerification: true,
      requireResidentKey: false,
      attestationRequirement: 'indirect',
      registrationTimeoutMs: 60000,
      authenticationTimeoutMs: 60000,
      maxCredentialsPerUser: 10,
      requireAttestationVerification: false
    };

    // Cache por 1 hora
    await this.redis.setex(cacheKey, 3600, JSON.stringify(config));
    
    return config;
  }

  private async handleSignCountAnomaly(
    credential: WebAuthnCredential,
    newCounter: number,
    context: WebAuthnContext
  ): Promise<void> {
    // Suspender credencial
    await this.credentialService.suspendCredential(
      credential.id,
      'SIGN_COUNT_ANOMALY',
      `Expected: ${credential.signCount}, Received: ${newCounter}`
    );

    // Registrar evento de anomalia
    await this.auditService.logWebAuthnEvent({
      type: 'sign_count_anomaly',
      userId: credential.userId,
      tenantId: credential.tenantId,
      credentialId: credential.id,
      result: 'warning',
      errorCode: 'SIGN_COUNT_ANOMALY',
      errorMessage: 'Sign count anomaly detected - possible credential cloning',
      signCount: newCounter,
      ipAddress: context.ipAddress,
      userAgent: context.userAgent,
      correlationId: context.correlationId,
      metadata: {
        expectedCounter: credential.signCount,
        receivedCounter: newCounter,
        anomalyDetectedAt: new Date().toISOString()
      }
    });

    // Métrica de anomalia
    webauthnMetrics.signCountAnomalies.inc({
      tenant_id: credential.tenantId
    });

    this.logger.error('Sign count anomaly detected', {
      credentialId: credential.credentialId,
      userId: credential.userId,
      tenantId: credential.tenantId,
      expectedCounter: credential.signCount,
      receivedCounter: newCounter,
      correlationId: context.correlationId
    });

    // Alertar equipe de segurança (implementar notificação)
    // await this.notifySecurityTeam({...});
  }
}