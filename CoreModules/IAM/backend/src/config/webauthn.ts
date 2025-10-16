/**
 * ============================================================================
 * INNOVABIZ IAM - WebAuthn Configuration
 * ============================================================================
 * Version: 1.0.0
 * Date: 31/07/2025
 * Author: Equipe de Segurança INNOVABIZ
 * Classification: Confidencial - Interno
 * 
 * Description: Configuração do WebAuthn/FIDO2 para INNOVABIZ IAM
 * Standards: W3C WebAuthn Level 3, FIDO2 CTAP2.1, NIST SP 800-63B
 * Compliance: PCI DSS 4.0, GDPR/LGPD, PSD2, ISO 27001
 * ============================================================================
 */

import {
  AttestationConveyancePreference,
  UserVerificationRequirement,
  AuthenticatorAttachment,
  ResidentKeyRequirement
} from '@simplewebauthn/typescript-types';

/**
 * Configuração principal do WebAuthn
 */
export const webauthnConfig = {
  // Identificação do Relying Party (RP)
  rpName: process.env.WEBAUTHN_RP_NAME || 'INNOVABIZ IAM',
  rpID: process.env.WEBAUTHN_RP_ID || 'innovabiz.com',
  
  // Origens permitidas
  origin: (process.env.WEBAUTHN_ORIGINS || 'https://innovabiz.com,https://app.innovabiz.com,https://api.innovabiz.com')
    .split(',')
    .map(origin => origin.trim()),
  
  // Timeout padrão (60 segundos)
  timeout: parseInt(process.env.WEBAUTHN_TIMEOUT || '60000'),
  
  // Preferência de attestation
  attestation: (process.env.WEBAUTHN_ATTESTATION || 'indirect') as AttestationConveyancePreference,
  
  // Seleção de autenticador
  authenticatorSelection: {
    authenticatorAttachment: process.env.WEBAUTHN_AUTHENTICATOR_ATTACHMENT as AuthenticatorAttachment || undefined,
    userVerification: (process.env.WEBAUTHN_USER_VERIFICATION || 'preferred') as UserVerificationRequirement,
    residentKey: (process.env.WEBAUTHN_RESIDENT_KEY || 'preferred') as ResidentKeyRequirement
  },
  
  // Tamanho do challenge (32 bytes = 256 bits)
  challengeLength: parseInt(process.env.WEBAUTHN_CHALLENGE_LENGTH || '32'),
  
  // Configurações de rate limiting
  rateLimiting: {
    registrationPerUser: parseInt(process.env.WEBAUTHN_RATE_LIMIT_REG_USER || '5'),
    authenticationPerUser: parseInt(process.env.WEBAUTHN_RATE_LIMIT_AUTH_USER || '10'),
    registrationPerIP: parseInt(process.env.WEBAUTHN_RATE_LIMIT_REG_IP || '20'),
    authenticationPerIP: parseInt(process.env.WEBAUTHN_RATE_LIMIT_AUTH_IP || '50'),
    windowMinutes: parseInt(process.env.WEBAUTHN_RATE_LIMIT_WINDOW || '15')
  },
  
  // Configurações de segurança
  security: {
    requireOriginValidation: process.env.WEBAUTHN_REQUIRE_ORIGIN_VALIDATION !== 'false',
    requireRPIDValidation: process.env.WEBAUTHN_REQUIRE_RPID_VALIDATION !== 'false',
    allowInsecureOrigins: process.env.NODE_ENV === 'development' && 
                          process.env.WEBAUTHN_ALLOW_INSECURE_ORIGINS === 'true',
    signCountAnomalyThreshold: parseInt(process.env.WEBAUTHN_SIGN_COUNT_ANOMALY_THRESHOLD || '0')
  },
  
  // Configurações de cache Redis
  cache: {
    challengeTTL: parseInt(process.env.WEBAUTHN_CHALLENGE_TTL || '300'), // 5 minutos
    credentialCacheTTL: parseInt(process.env.WEBAUTHN_CREDENTIAL_CACHE_TTL || '600'), // 10 minutos
    userCredentialsCacheTTL: parseInt(process.env.WEBAUTHN_USER_CREDENTIALS_CACHE_TTL || '300') // 5 minutos
  },
  
  // Configurações de compliance
  compliance: {
    // Níveis mínimos de AAL por contexto
    minimumAAL: {
      default: process.env.WEBAUTHN_MINIMUM_AAL_DEFAULT || 'AAL1',
      financial: process.env.WEBAUTHN_MINIMUM_AAL_FINANCIAL || 'AAL2',
      administrative: process.env.WEBAUTHN_MINIMUM_AAL_ADMIN || 'AAL3'
    },
    
    // Requisitos de verificação de usuário
    requireUserVerification: {
      registration: process.env.WEBAUTHN_REQUIRE_UV_REGISTRATION === 'true',
      authentication: process.env.WEBAUTHN_REQUIRE_UV_AUTHENTICATION === 'true',
      stepUp: process.env.WEBAUTHN_REQUIRE_UV_STEP_UP !== 'false'
    },
    
    // Requisitos de attestation
    requireAttestation: {
      registration: process.env.WEBAUTHN_REQUIRE_ATTESTATION_REGISTRATION === 'true',
      verification: process.env.WEBAUTHN_REQUIRE_ATTESTATION_VERIFICATION === 'true'
    }
  },
  
  // Configurações de observabilidade
  observability: {
    metrics: {
      enabled: process.env.WEBAUTHN_METRICS_ENABLED !== 'false',
      prefix: process.env.WEBAUTHN_METRICS_PREFIX || 'innovabiz_webauthn_',
      labels: ['tenant_id', 'result', 'attestation_format', 'aal_level']
    },
    
    logging: {
      level: process.env.WEBAUTHN_LOG_LEVEL || 'info',
      includeCredentialId: process.env.WEBAUTHN_LOG_INCLUDE_CREDENTIAL_ID === 'true',
      includeSensitiveData: process.env.NODE_ENV === 'development' && 
                           process.env.WEBAUTHN_LOG_INCLUDE_SENSITIVE === 'true'
    },
    
    tracing: {
      enabled: process.env.WEBAUTHN_TRACING_ENABLED === 'true',
      serviceName: process.env.WEBAUTHN_TRACING_SERVICE_NAME || 'innovabiz-iam-webauthn',
      sampleRate: parseFloat(process.env.WEBAUTHN_TRACING_SAMPLE_RATE || '0.1')
    }
  }
};

/**
 * Configurações regionais padrão
 */
export const defaultRegionalConfigs = {
  // Brasil (LGPD, Banco Central)
  BR: {
    requireUserVerification: true,
    requireResidentKey: false,
    attestationRequirement: 'indirect' as AttestationConveyancePreference,
    registrationTimeoutMs: 60000,
    authenticationTimeoutMs: 60000,
    maxCredentialsPerUser: 10,
    requireAttestationVerification: true,
    riskThresholds: { low: 0.3, medium: 0.6, high: 0.8 },
    complianceRequirements: {
      minimumAAL: 'AAL2' as const,
      requireBiometrics: false,
      requireSecurityKey: false
    }
  },
  
  // Estados Unidos (NIST, FIDO Alliance)
  US: {
    requireUserVerification: true,
    requireResidentKey: false,
    attestationRequirement: 'indirect' as AttestationConveyancePreference,
    registrationTimeoutMs: 60000,
    authenticationTimeoutMs: 60000,
    maxCredentialsPerUser: 15,
    requireAttestationVerification: false,
    riskThresholds: { low: 0.25, medium: 0.5, high: 0.75 },
    complianceRequirements: {
      minimumAAL: 'AAL2' as const,
      requireBiometrics: false,
      requireSecurityKey: false
    }
  },
  
  // União Europeia (GDPR, PSD2, eIDAS)
  EU: {
    requireUserVerification: true,
    requireResidentKey: true,
    attestationRequirement: 'direct' as AttestationConveyancePreference,
    registrationTimeoutMs: 90000,
    authenticationTimeoutMs: 60000,
    maxCredentialsPerUser: 8,
    requireAttestationVerification: true,
    riskThresholds: { low: 0.2, medium: 0.4, high: 0.7 },
    complianceRequirements: {
      minimumAAL: 'AAL3' as const,
      requireBiometrics: true,
      requireSecurityKey: false
    }
  },
  
  // Reino Unido (UK GDPR, FCA)
  GB: {
    requireUserVerification: true,
    requireResidentKey: false,
    attestationRequirement: 'indirect' as AttestationConveyancePreference,
    registrationTimeoutMs: 60000,
    authenticationTimeoutMs: 60000,
    maxCredentialsPerUser: 12,
    requireAttestationVerification: true,
    riskThresholds: { low: 0.25, medium: 0.5, high: 0.75 },
    complianceRequirements: {
      minimumAAL: 'AAL2' as const,
      requireBiometrics: false,
      requireSecurityKey: false
    }
  },
  
  // Configuração padrão global
  DEFAULT: {
    requireUserVerification: true,
    requireResidentKey: false,
    attestationRequirement: 'indirect' as AttestationConveyancePreference,
    registrationTimeoutMs: 60000,
    authenticationTimeoutMs: 60000,
    maxCredentialsPerUser: 10,
    requireAttestationVerification: false,
    riskThresholds: { low: 0.3, medium: 0.6, high: 0.8 },
    complianceRequirements: {
      minimumAAL: 'AAL1' as const,
      requireBiometrics: false,
      requireSecurityKey: false
    }
  }
};

/**
 * Templates de erro padronizados
 */
export const errorTemplates = {
  REGISTRATION_OPTIONS_FAILED: {
    code: 'REGISTRATION_OPTIONS_FAILED',
    message: 'Failed to generate registration options',
    httpStatus: 500,
    retryable: true,
    category: 'server' as const
  },
  
  REGISTRATION_VERIFICATION_FAILED: {
    code: 'REGISTRATION_VERIFICATION_FAILED',
    message: 'Registration verification failed',
    httpStatus: 400,
    retryable: false,
    category: 'client' as const
  },
  
  AUTHENTICATION_OPTIONS_FAILED: {
    code: 'AUTHENTICATION_OPTIONS_FAILED',
    message: 'Failed to generate authentication options',
    httpStatus: 500,
    retryable: true,
    category: 'server' as const
  },
  
  AUTHENTICATION_VERIFICATION_FAILED: {
    code: 'AUTHENTICATION_VERIFICATION_FAILED',
    message: 'Authentication verification failed',
    httpStatus: 401,
    retryable: false,
    category: 'security' as const
  },
  
  CREDENTIAL_NOT_FOUND: {
    code: 'CREDENTIAL_NOT_FOUND',
    message: 'Credential not found or inactive',
    httpStatus: 404,
    retryable: false,
    category: 'client' as const
  },
  
  CHALLENGE_NOT_FOUND: {
    code: 'CHALLENGE_NOT_FOUND',
    message: 'Challenge not found or expired',
    httpStatus: 400,
    retryable: false,
    category: 'client' as const
  },
  
  SIGN_COUNT_ANOMALY: {
    code: 'SIGN_COUNT_ANOMALY',
    message: 'Sign count anomaly detected - possible credential cloning',
    httpStatus: 403,
    retryable: false,
    category: 'security' as const
  },
  
  MAX_CREDENTIALS_EXCEEDED: {
    code: 'MAX_CREDENTIALS_EXCEEDED',
    message: 'Maximum credentials per user exceeded',
    httpStatus: 429,
    retryable: false,
    category: 'validation' as const
  },
  
  RATE_LIMIT_EXCEEDED: {
    code: 'RATE_LIMIT_EXCEEDED',
    message: 'Rate limit exceeded',
    httpStatus: 429,
    retryable: true,
    category: 'client' as const
  },
  
  INVALID_ORIGIN: {
    code: 'INVALID_ORIGIN',
    message: 'Invalid origin',
    httpStatus: 403,
    retryable: false,
    category: 'security' as const
  },
  
  INVALID_RPID: {
    code: 'INVALID_RPID',
    message: 'Invalid RP ID',
    httpStatus: 403,
    retryable: false,
    category: 'security' as const
  },
  
  ATTESTATION_VERIFICATION_FAILED: {
    code: 'ATTESTATION_VERIFICATION_FAILED',
    message: 'Attestation verification failed',
    httpStatus: 400,
    retryable: false,
    category: 'security' as const
  },
  
  INSUFFICIENT_AAL: {
    code: 'INSUFFICIENT_AAL',
    message: 'Insufficient authentication assurance level',
    httpStatus: 403,
    retryable: false,
    category: 'security' as const
  },
  
  USER_VERIFICATION_REQUIRED: {
    code: 'USER_VERIFICATION_REQUIRED',
    message: 'User verification is required',
    httpStatus: 400,
    retryable: false,
    category: 'validation' as const
  },
  
  CREDENTIAL_SUSPENDED: {
    code: 'CREDENTIAL_SUSPENDED',
    message: 'Credential is suspended',
    httpStatus: 403,
    retryable: false,
    category: 'security' as const
  },
  
  HIGH_RISK_DETECTED: {
    code: 'HIGH_RISK_DETECTED',
    message: 'High risk authentication detected',
    httpStatus: 403,
    retryable: false,
    category: 'security' as const
  }
};

/**
 * Configuração de validação de entrada
 */
export const validationConfig = {
  username: {
    minLength: 3,
    maxLength: 64,
    pattern: /^[a-zA-Z0-9._-]+$/
  },
  
  displayName: {
    minLength: 1,
    maxLength: 128,
    pattern: /^[\p{L}\p{N}\p{P}\p{Z}]+$/u
  },
  
  friendlyName: {
    minLength: 1,
    maxLength: 64,
    pattern: /^[\p{L}\p{N}\p{P}\p{Z}]+$/u
  },
  
  correlationId: {
    pattern: /^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$/i
  },
  
  credentialId: {
    minLength: 16,
    maxLength: 1024
  },
  
  challenge: {
    minLength: 16,
    maxLength: 128
  }
};

/**
 * Configuração de monitoramento e alertas
 */
export const monitoringConfig = {
  healthCheck: {
    enabled: process.env.WEBAUTHN_HEALTH_CHECK_ENABLED !== 'false',
    interval: parseInt(process.env.WEBAUTHN_HEALTH_CHECK_INTERVAL || '30000'), // 30 segundos
    timeout: parseInt(process.env.WEBAUTHN_HEALTH_CHECK_TIMEOUT || '5000'), // 5 segundos
    endpoints: [
      '/health',
      '/health/webauthn',
      '/health/database',
      '/health/redis'
    ]
  },
  
  alerts: {
    enabled: process.env.WEBAUTHN_ALERTS_ENABLED === 'true',
    rules: [
      {
        name: 'high_failure_rate',
        condition: 'failure_rate > 0.1',
        threshold: 0.1,
        severity: 'high' as const
      },
      {
        name: 'sign_count_anomalies',
        condition: 'anomaly_rate > 0.05',
        threshold: 0.05,
        severity: 'critical' as const
      },
      {
        name: 'high_response_time',
        condition: 'avg_response_time > 2000',
        threshold: 2000,
        severity: 'medium' as const
      }
    ]
  }
};

/**
 * Validação de configuração
 */
export function validateConfig(): void {
  const errors: string[] = [];
  
  // Validar RP ID
  if (!webauthnConfig.rpID || webauthnConfig.rpID.length === 0) {
    errors.push('WEBAUTHN_RP_ID is required');
  }
  
  // Validar origens
  if (!webauthnConfig.origin || webauthnConfig.origin.length === 0) {
    errors.push('WEBAUTHN_ORIGINS is required');
  }
  
  // Validar timeout
  if (webauthnConfig.timeout < 30000 || webauthnConfig.timeout > 300000) {
    errors.push('WEBAUTHN_TIMEOUT must be between 30000 and 300000 milliseconds');
  }
  
  // Validar challenge length
  if (webauthnConfig.challengeLength < 16 || webauthnConfig.challengeLength > 64) {
    errors.push('WEBAUTHN_CHALLENGE_LENGTH must be between 16 and 64 bytes');
  }
  
  if (errors.length > 0) {
    throw new Error(`WebAuthn configuration validation failed:\n${errors.join('\n')}`);
  }
}

// Validar configuração na inicialização
if (process.env.NODE_ENV !== 'test') {
  validateConfig();
}