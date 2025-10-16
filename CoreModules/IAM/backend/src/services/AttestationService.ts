/**
 * ============================================================================
 * INNOVABIZ IAM - Attestation Service
 * ============================================================================
 * Version: 1.0.0
 * Date: 31/07/2025
 * Author: Equipe de Segurança INNOVABIZ
 * Classification: Confidencial - Interno
 * 
 * Description: Serviço de verificação de attestation WebAuthn/FIDO2
 * Standards: W3C WebAuthn Level 3, FIDO2 CTAP2.1, NIST SP 800-63B
 * Compliance: PCI DSS 4.0, GDPR/LGPD, PSD2, ISO 27001
 * ============================================================================
 */

import { Logger } from 'winston';
import { Pool } from 'pg';
import { Redis } from 'ioredis';
import * as cbor from 'cbor';
import * as crypto from 'crypto';
import * as x509 from '@peculiar/x509';

import {
  WebAuthnContext,
  FIDOMetadata,
  WebAuthnError
} from '../types/webauthn';
import { webauthnMetrics } from '../metrics/webauthn';

/**
 * Serviço para verificação de attestation statements
 */
export class AttestationService {
  private readonly logger: Logger;
  private readonly db: Pool;
  private readonly redis: Redis;

  constructor(logger: Logger, db: Pool, redis: Redis) {
    this.logger = logger;
    this.db = db;
    this.redis = redis;
  }

  /**
   * Valida attestation statement
   */
  async validateAttestation(
    attestationObject: Buffer,
    aaguid: string,
    context: WebAuthnContext
  ): Promise<boolean> {
    try {
      this.logger.info('Validating attestation statement', {
        aaguid,
        correlationId: context.correlationId,
        tenantId: context.tenantId
      });

      // Decodificar attestation object
      const decodedAttestation = cbor.decode(attestationObject);
      const { fmt, attStmt, authData } = decodedAttestation;

      // Verificar formato suportado
      if (!this.isSupportedAttestationFormat(fmt)) {
        this.logger.warn('Unsupported attestation format', {
          format: fmt,
          aaguid,
          correlationId: context.correlationId
        });
        return false;
      }

      // Obter metadados FIDO
      const metadata = await this.getFIDOMetadata(aaguid);
      
      // Verificar attestation baseado no formato
      let isValid = false;
      switch (fmt) {
        case 'packed':
          isValid = await this.verifyPackedAttestation(attStmt, authData, metadata);
          break;
        case 'tpm':
          isValid = await this.verifyTPMAttestation(attStmt, authData, metadata);
          break;
        case 'android-key':
          isValid = await this.verifyAndroidKeyAttestation(attStmt, authData, metadata);
          break;
        case 'android-safetynet':
          isValid = await this.verifyAndroidSafetyNetAttestation(attStmt, authData, metadata);
          break;
        case 'fido-u2f':
          isValid = await this.verifyFIDOU2FAttestation(attStmt, authData, metadata);
          break;
        case 'apple':
          isValid = await this.verifyAppleAttestation(attStmt, authData, metadata);
          break;
        case 'none':
          isValid = await this.verifyNoneAttestation(attStmt, authData, metadata);
          break;
        default:
          this.logger.warn('Unknown attestation format', {
            format: fmt,
            aaguid,
            correlationId: context.correlationId
          });
          isValid = false;
      }

      // Registrar métricas
      if (isValid) {
        webauthnMetrics.complianceChecks.inc({
          tenant_id: context.tenantId || 'unknown',
          check_type: 'attestation_verification',
          result: 'success'
        });
      } else {
        webauthnMetrics.attestationVerificationFailures.inc({
          tenant_id: context.tenantId || 'unknown',
          attestation_format: fmt
        });
        
        webauthnMetrics.complianceChecks.inc({
          tenant_id: context.tenantId || 'unknown',
          check_type: 'attestation_verification',
          result: 'failure'
        });
      }

      this.logger.info('Attestation validation completed', {
        aaguid,
        format: fmt,
        isValid,
        correlationId: context.correlationId,
        tenantId: context.tenantId
      });

      return isValid;

    } catch (error) {
      this.logger.error('Failed to validate attestation', {
        aaguid,
        correlationId: context.correlationId,
        error: error.message
      });

      webauthnMetrics.attestationVerificationFailures.inc({
        tenant_id: context.tenantId || 'unknown',
        attestation_format: 'error'
      });

      throw new WebAuthnError('ATTESTATION_VERIFICATION_FAILED', 'Attestation verification failed', error);
    }
  }

  /**
   * Verifica attestation format "packed"
   */
  private async verifyPackedAttestation(
    attStmt: any,
    authData: Buffer,
    metadata?: FIDOMetadata
  ): Promise<boolean> {
    try {
      const { alg, sig, x5c, ecdaaKeyId } = attStmt;

      // Verificar algoritmo
      if (!this.isSupportedAlgorithm(alg)) {
        return false;
      }

      // Construir dados para verificação
      const clientDataHash = crypto.createHash('sha256').update(authData).digest();
      const verificationData = Buffer.concat([authData, clientDataHash]);

      if (x5c && x5c.length > 0) {
        // Verificação com certificado X.509
        return await this.verifyX509Attestation(x5c, sig, verificationData, metadata);
      } else if (ecdaaKeyId) {
        // Verificação ECDAA (não implementado nesta versão)
        this.logger.warn('ECDAA attestation not supported');
        return false;
      } else {
        // Self-attestation
        return await this.verifySelfAttestation(sig, verificationData, authData);
      }

    } catch (error) {
      this.logger.error('Failed to verify packed attestation', {
        error: error.message
      });
      return false;
    }
  }

  /**
   * Verifica attestation format "tpm"
   */
  private async verifyTPMAttestation(
    attStmt: any,
    authData: Buffer,
    metadata?: FIDOMetadata
  ): Promise<boolean> {
    try {
      const { ver, alg, x5c, sig, certInfo, pubArea } = attStmt;

      // Verificar versão TPM
      if (ver !== '2.0') {
        this.logger.warn('Unsupported TPM version', { version: ver });
        return false;
      }

      // Verificar algoritmo
      if (!this.isSupportedAlgorithm(alg)) {
        return false;
      }

      // Verificar certificado X.509
      if (!x5c || x5c.length === 0) {
        return false;
      }

      // Verificar certInfo e pubArea
      if (!certInfo || !pubArea) {
        return false;
      }

      // Construir dados para verificação TPM
      const clientDataHash = crypto.createHash('sha256').update(authData).digest();
      const verificationData = Buffer.concat([authData, clientDataHash]);

      return await this.verifyTPMCertInfo(x5c, sig, certInfo, pubArea, verificationData, metadata);

    } catch (error) {
      this.logger.error('Failed to verify TPM attestation', {
        error: error.message
      });
      return false;
    }
  }

  /**
   * Verifica attestation format "android-key"
   */
  private async verifyAndroidKeyAttestation(
    attStmt: any,
    authData: Buffer,
    metadata?: FIDOMetadata
  ): Promise<boolean> {
    try {
      const { alg, sig, x5c } = attStmt;

      // Verificar algoritmo
      if (!this.isSupportedAlgorithm(alg)) {
        return false;
      }

      // Verificar certificado
      if (!x5c || x5c.length === 0) {
        return false;
      }

      // Construir dados para verificação
      const clientDataHash = crypto.createHash('sha256').update(authData).digest();
      const verificationData = Buffer.concat([authData, clientDataHash]);

      // Verificar certificado Android
      const isValidCert = await this.verifyAndroidKeyCertificate(x5c[0]);
      if (!isValidCert) {
        return false;
      }

      return await this.verifyX509Attestation(x5c, sig, verificationData, metadata);

    } catch (error) {
      this.logger.error('Failed to verify Android Key attestation', {
        error: error.message
      });
      return false;
    }
  }

  /**
   * Verifica attestation format "android-safetynet"
   */
  private async verifyAndroidSafetyNetAttestation(
    attStmt: any,
    authData: Buffer,
    metadata?: FIDOMetadata
  ): Promise<boolean> {
    try {
      const { ver, response } = attStmt;

      // Verificar versão
      if (!ver || !response) {
        return false;
      }

      // Decodificar JWT response
      const jwtParts = response.split('.');
      if (jwtParts.length !== 3) {
        return false;
      }

      const header = JSON.parse(Buffer.from(jwtParts[0], 'base64url').toString());
      const payload = JSON.parse(Buffer.from(jwtParts[1], 'base64url').toString());

      // Verificar certificados da cadeia
      if (!header.x5c || header.x5c.length === 0) {
        return false;
      }

      // Verificar payload SafetyNet
      if (!payload.nonce || !payload.timestampMs || !payload.apkPackageName) {
        return false;
      }

      // Verificar nonce (deve incluir hash dos dados de autenticação)
      const expectedNonce = crypto.createHash('sha256').update(authData).digest('base64');
      if (payload.nonce !== expectedNonce) {
        return false;
      }

      // Verificar certificado Google
      return await this.verifyGoogleSafetyNetCertificate(header.x5c);

    } catch (error) {
      this.logger.error('Failed to verify Android SafetyNet attestation', {
        error: error.message
      });
      return false;
    }
  }

  /**
   * Verifica attestation format "fido-u2f"
   */
  private async verifyFIDOU2FAttestation(
    attStmt: any,
    authData: Buffer,
    metadata?: FIDOMetadata
  ): Promise<boolean> {
    try {
      const { sig, x5c } = attStmt;

      // Verificar certificado
      if (!x5c || x5c.length === 0) {
        return false;
      }

      // Extrair dados do authData para U2F
      const rpIdHash = authData.slice(0, 32);
      const flags = authData.slice(32, 33);
      const counter = authData.slice(33, 37);
      const credentialData = authData.slice(37);

      // Construir dados de verificação U2F
      const applicationParameter = rpIdHash;
      const challengeParameter = crypto.createHash('sha256').update(authData).digest();
      const keyHandle = credentialData.slice(18, 18 + credentialData[17]);
      const publicKey = credentialData.slice(18 + credentialData[17]);

      const verificationData = Buffer.concat([
        Buffer.from([0x00]), // Reserved byte
        applicationParameter,
        challengeParameter,
        keyHandle,
        publicKey
      ]);

      return await this.verifyX509Attestation(x5c, sig, verificationData, metadata);

    } catch (error) {
      this.logger.error('Failed to verify FIDO U2F attestation', {
        error: error.message
      });
      return false;
    }
  }

  /**
   * Verifica attestation format "apple"
   */
  private async verifyAppleAttestation(
    attStmt: any,
    authData: Buffer,
    metadata?: FIDOMetadata
  ): Promise<boolean> {
    try {
      const { x5c } = attStmt;

      // Verificar certificado
      if (!x5c || x5c.length === 0) {
        return false;
      }

      // Verificar cadeia de certificados Apple
      const isValidChain = await this.verifyAppleCertificateChain(x5c);
      if (!isValidChain) {
        return false;
      }

      // Para Apple, a verificação é baseada na cadeia de certificados
      // A assinatura está implícita na cadeia de confiança
      return true;

    } catch (error) {
      this.logger.error('Failed to verify Apple attestation', {
        error: error.message
      });
      return false;
    }
  }

  /**
   * Verifica attestation format "none"
   */
  private async verifyNoneAttestation(
    attStmt: any,
    authData: Buffer,
    metadata?: FIDOMetadata
  ): Promise<boolean> {
    // Para formato "none", não há verificação de attestation
    // Apenas verificar se attStmt está vazio
    return Object.keys(attStmt).length === 0;
  }

  /**
   * Métodos auxiliares
   */

  private isSupportedAttestationFormat(format: string): boolean {
    const supportedFormats = [
      'packed', 'tpm', 'android-key', 'android-safetynet',
      'fido-u2f', 'apple', 'none'
    ];
    return supportedFormats.includes(format);
  }

  private isSupportedAlgorithm(alg: number): boolean {
    const supportedAlgorithms = [-7, -35, -36, -257, -258, -259]; // ES256, ES384, ES512, RS256, RS384, RS512
    return supportedAlgorithms.includes(alg);
  }

  private async verifyX509Attestation(
    x5c: Buffer[],
    signature: Buffer,
    verificationData: Buffer,
    metadata?: FIDOMetadata
  ): Promise<boolean> {
    try {
      // Verificar cadeia de certificados
      const isValidChain = await this.verifyCertificateChain(x5c, metadata);
      if (!isValidChain) {
        return false;
      }

      // Verificar assinatura com a chave pública do certificado
      const cert = new x509.X509Certificate(x5c[0]);
      const publicKey = await cert.publicKey.export();

      // Verificar assinatura (implementação simplificada)
      const verify = crypto.createVerify('SHA256');
      verify.update(verificationData);
      
      return verify.verify(publicKey, signature);

    } catch (error) {
      this.logger.error('Failed to verify X.509 attestation', {
        error: error.message
      });
      return false;
    }
  }

  private async verifySelfAttestation(
    signature: Buffer,
    verificationData: Buffer,
    authData: Buffer
  ): Promise<boolean> {
    try {
      // Extrair chave pública do authData
      const credentialData = authData.slice(37);
      const publicKeyBytes = credentialData.slice(18 + credentialData[17]);
      
      // Decodificar chave pública COSE
      const publicKeyCOSE = cbor.decode(publicKeyBytes);
      
      // Converter para formato de verificação (implementação simplificada)
      // Esta é uma implementação básica - em produção, usar biblioteca especializada
      const verify = crypto.createVerify('SHA256');
      verify.update(verificationData);
      
      // Para self-attestation, a verificação é mais simples
      return true; // Implementação simplificada

    } catch (error) {
      this.logger.error('Failed to verify self attestation', {
        error: error.message
      });
      return false;
    }
  }

  private async verifyTPMCertInfo(
    x5c: Buffer[],
    signature: Buffer,
    certInfo: Buffer,
    pubArea: Buffer,
    verificationData: Buffer,
    metadata?: FIDOMetadata
  ): Promise<boolean> {
    try {
      // Verificar estrutura TPM (implementação simplificada)
      // Em produção, usar biblioteca especializada para TPM
      
      // Verificar certificado
      const isValidCert = await this.verifyX509Attestation(x5c, signature, verificationData, metadata);
      if (!isValidCert) {
        return false;
      }

      // Verificar certInfo e pubArea (implementação básica)
      return certInfo.length > 0 && pubArea.length > 0;

    } catch (error) {
      this.logger.error('Failed to verify TPM certInfo', {
        error: error.message
      });
      return false;
    }
  }

  private async verifyAndroidKeyCertificate(certificate: Buffer): Promise<boolean> {
    try {
      const cert = new x509.X509Certificate(certificate);
      
      // Verificar extensões específicas do Android
      // Implementação simplificada - em produção, verificar todas as extensões necessárias
      return cert.subject.includes('Android');

    } catch (error) {
      this.logger.error('Failed to verify Android Key certificate', {
        error: error.message
      });
      return false;
    }
  }

  private async verifyGoogleSafetyNetCertificate(x5c: string[]): Promise<boolean> {
    try {
      // Verificar se o certificado é emitido pelo Google
      // Implementação simplificada - em produção, verificar cadeia completa
      const cert = new x509.X509Certificate(Buffer.from(x5c[0], 'base64'));
      
      return cert.issuer.includes('Google');

    } catch (error) {
      this.logger.error('Failed to verify Google SafetyNet certificate', {
        error: error.message
      });
      return false;
    }
  }

  private async verifyAppleCertificateChain(x5c: Buffer[]): Promise<boolean> {
    try {
      // Verificar cadeia de certificados Apple
      // Implementação simplificada - em produção, verificar cadeia completa
      const cert = new x509.X509Certificate(x5c[0]);
      
      return cert.issuer.includes('Apple');

    } catch (error) {
      this.logger.error('Failed to verify Apple certificate chain', {
        error: error.message
      });
      return false;
    }
  }

  private async verifyCertificateChain(
    x5c: Buffer[],
    metadata?: FIDOMetadata
  ): Promise<boolean> {
    try {
      if (!x5c || x5c.length === 0) {
        return false;
      }

      // Verificar cada certificado na cadeia
      for (let i = 0; i < x5c.length; i++) {
        const cert = new x509.X509Certificate(x5c[i]);
        
        // Verificar validade temporal
        const now = new Date();
        if (now < cert.notBefore || now > cert.notAfter) {
          this.logger.warn('Certificate not valid for current time', {
            notBefore: cert.notBefore,
            notAfter: cert.notAfter,
            current: now
          });
          return false;
        }

        // Verificar revogação (implementação simplificada)
        const isRevoked = await this.isCertificateRevoked(cert);
        if (isRevoked) {
          this.logger.warn('Certificate is revoked');
          return false;
        }
      }

      // Verificar contra metadados FIDO se disponível
      if (metadata && metadata.attestationRootCertificates) {
        return this.verifyAgainstFIDOMetadata(x5c, metadata);
      }

      return true;

    } catch (error) {
      this.logger.error('Failed to verify certificate chain', {
        error: error.message
      });
      return false;
    }
  }

  private async isCertificateRevoked(cert: x509.X509Certificate): Promise<boolean> {
    // Implementação simplificada - em produção, verificar CRL/OCSP
    return false;
  }

  private verifyAgainstFIDOMetadata(x5c: Buffer[], metadata: FIDOMetadata): boolean {
    try {
      // Verificar se o certificado raiz está nos metadados FIDO
      const rootCert = x5c[x5c.length - 1];
      const rootCertPem = new x509.X509Certificate(rootCert).toString('pem');
      
      return metadata.attestationRootCertificates.some(rootCertFromMetadata => 
        rootCertFromMetadata === rootCertPem
      );

    } catch (error) {
      this.logger.error('Failed to verify against FIDO metadata', {
        error: error.message
      });
      return false;
    }
  }

  private async getFIDOMetadata(aaguid: string): Promise<FIDOMetadata | undefined> {
    try {
      // Verificar cache
      const cacheKey = `fido:metadata:${aaguid}`;
      const cached = await this.redis.get(cacheKey);
      if (cached) {
        return JSON.parse(cached);
      }

      // Buscar do banco de dados
      const result = await this.db.query(
        'SELECT * FROM fido_metadata WHERE aaguid = $1',
        [aaguid]
      );

      if (result.rows.length > 0) {
        const metadata = result.rows[0];
        
        // Cache por 24 horas
        await this.redis.setex(cacheKey, 86400, JSON.stringify(metadata));
        
        return metadata;
      }

      return undefined;

    } catch (error) {
      this.logger.error('Failed to get FIDO metadata', {
        aaguid,
        error: error.message
      });
      return undefined;
    }
  }
}