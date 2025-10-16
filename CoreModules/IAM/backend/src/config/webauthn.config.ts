/**
 * 🔐 WEBAUTHN CONFIGURATION - INNOVABIZ IAM
 * Configurações avançadas para WebAuthn/FIDO2
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: W3C WebAuthn Level 3, FIDO2 CTAP2.1, NIST SP 800-63B
 * Security: Attestation, Resident Keys, User Verification
 */

import { registerAs } from '@nestjs/config';

export default registerAs('webauthn', () => ({
  // ========================================
  // RELYING PARTY CONFIGURATION
  // ========================================
  rpID: process.env.WEBAUTHN_RP_ID || 'localhost',
  rpName: process.env.WEBAUTHN_RP_NAME || 'INNOVABIZ Platform',
  rpIcon: process.env.WEBAUTHN_RP_ICON || 'https://innovabiz.com/icon.png',
  
  // Origins permitidas
  origin: process.env.WEBAUTHN_ORIGIN || 'http://localhost:3000',
  allowedOrigins: process.env.WEBAUTHN_ALLOWED_ORIGINS?.split(',') || [
    'http://localhost:3000',
    'https://localhost:3000',
    'https://innovabiz.com',
    'https://app.innovabiz.com'
  ],

  // ========================================
  // TIMEOUT CONFIGURATION
  // ========================================
  timeout: parseInt(process.env.WEBAUTHN_TIMEOUT) || 300000, // 5 minutos
  challengeTimeout: parseInt(process.env.WEBAUTHN_CHALLENGE_TIMEOUT) || 300000, // 5 minutos
  registrationTimeout: parseInt(process.env.WEBAUTHN_REGISTRATION_TIMEOUT) || 600000, // 10 minutos
  authenticationTimeout: parseInt(process.env.WEBAUTHN_AUTHENTICATION_TIMEOUT) || 300000, // 5 minutos

  // ========================================
  // ATTESTATION CONFIGURATION
  // ========================================
  attestation: process.env.WEBAUTHN_ATTESTATION as 'none' | 'indirect' | 'direct' | 'enterprise' || 'none',
  
  // Verificação de attestation
  enableAttestationVerification: process.env.WEBAUTHN_ENABLE_ATTESTATION_VERIFICATION === 'true',
  trustedAttestationRoots: process.env.WEBAUTHN_TRUSTED_ROOTS?.split(',') || [],
  
  // Configurações de attestation por ambiente
  attestationByEnvironment: {
    development: 'none' as const,
    testing: 'none' as const,
    staging: 'indirect' as const,
    production: process.env.WEBAUTHN_PRODUCTION_ATTESTATION as 'none' | 'indirect' | 'direct' | 'enterprise' || 'indirect'
  },

  // ========================================
  // AUTHENTICATOR SELECTION CRITERIA
  // ========================================
  authenticatorSelection: {
    // Anexo do autenticador
    authenticatorAttachment: process.env.WEBAUTHN_AUTHENTICATOR_ATTACHMENT as 'platform' | 'cross-platform' | undefined,
    
    // Chaves residentes (discoverable credentials)
    residentKey: process.env.WEBAUTHN_RESIDENT_KEY as 'discouraged' | 'preferred' | 'required' || 'preferred',
    requireResidentKey: process.env.WEBAUTHN_REQUIRE_RESIDENT_KEY === 'true',
    
    // Verificação do usuário
    userVerification: process.env.WEBAUTHN_USER_VERIFICATION as 'required' | 'preferred' | 'discouraged' || 'preferred'
  },

  // ========================================
  // SUPPORTED ALGORITHMS
  // ========================================
  supportedAlgorithmIDs: [
    -7,   // ES256 (ECDSA w/ SHA-256)
    -257, // RS256 (RSASSA-PKCS1-v1_5 w/ SHA-256)
    -8,   // EdDSA
    -37,  // PS256 (RSASSA-PSS w/ SHA-256)
    -38,  // PS384 (RSASSA-PSS w/ SHA-384)
    -39,  // PS512 (RSASSA-PSS w/ SHA-512)
    -35,  // ES384 (ECDSA w/ SHA-384)
    -36   // ES512 (ECDSA w/ SHA-512)
  ],

  // Preferência de algoritmos por segurança
  preferredAlgorithms: {
    high: [-8, -7, -35, -36], // EdDSA, ES256, ES384, ES512
    medium: [-257, -37, -38, -39], // RS256, PS256, PS384, PS512
    legacy: [-7, -257] // ES256, RS256 para compatibilidade
  },

  // ========================================
  // TRANSPORT CONFIGURATION
  // ========================================
  supportedTransports: [
    'usb',
    'nfc',
    'ble',
    'smart-card',
    'hybrid',
    'internal'
  ] as const,

  // Transporte preferido por plataforma
  preferredTransportsByPlatform: {
    desktop: ['usb', 'nfc', 'hybrid', 'internal'],
    mobile: ['internal', 'hybrid', 'nfc', 'ble'],
    tablet: ['internal', 'hybrid', 'nfc', 'ble', 'usb']
  },

  // ========================================
  // SECURITY POLICIES
  // ========================================
  security: {
    // Verificação de origem rigorosa
    strictOriginValidation: process.env.WEBAUTHN_STRICT_ORIGIN_VALIDATION !== 'false',
    
    // Verificação de RP ID
    strictRPIDValidation: process.env.WEBAUTHN_STRICT_RPID_VALIDATION !== 'false',
    
    // Verificação de contador de assinatura
    enableSignCountVerification: process.env.WEBAUTHN_ENABLE_SIGN_COUNT_VERIFICATION !== 'false',
    signCountTolerance: parseInt(process.env.WEBAUTHN_SIGN_COUNT_TOLERANCE) || 0,
    
    // Verificação de clonagem
    enableCloneDetection: process.env.WEBAUTHN_ENABLE_CLONE_DETECTION !== 'false',
    
    // Verificação de backup eligibility
    allowBackupEligibleCredentials: process.env.WEBAUTHN_ALLOW_BACKUP_ELIGIBLE !== 'false',
    allowBackedUpCredentials: process.env.WEBAUTHN_ALLOW_BACKED_UP !== 'false',
    
    // Verificação de device public key
    requireDevicePublicKey: process.env.WEBAUTHN_REQUIRE_DEVICE_PUBLIC_KEY === 'true',
    
    // Verificação de enterprise attestation
    allowEnterpriseAttestation: process.env.WEBAUTHN_ALLOW_ENTERPRISE_ATTESTATION === 'true'
  },

  // ========================================
  // CHALLENGE CONFIGURATION
  // ========================================
  challenge: {
    // Tamanho do challenge em bytes
    size: parseInt(process.env.WEBAUTHN_CHALLENGE_SIZE) || 32,
    
    // Algoritmo de geração
    algorithm: process.env.WEBAUTHN_CHALLENGE_ALGORITHM || 'random',
    
    // Armazenamento do challenge
    storage: {
      type: process.env.WEBAUTHN_CHALLENGE_STORAGE || 'redis', // redis, memory, database
      ttl: parseInt(process.env.WEBAUTHN_CHALLENGE_TTL) || 300, // 5 minutos
      keyPrefix: process.env.WEBAUTHN_CHALLENGE_KEY_PREFIX || 'webauthn:challenge:'
    },
    
    // Verificação de replay
    enableReplayProtection: process.env.WEBAUTHN_ENABLE_REPLAY_PROTECTION !== 'false',
    replayWindowSeconds: parseInt(process.env.WEBAUTHN_REPLAY_WINDOW) || 30
  },

  // ========================================
  // USER CONFIGURATION
  // ========================================
  user: {
    // Tamanho do user handle em bytes
    handleSize: parseInt(process.env.WEBAUTHN_USER_HANDLE_SIZE) || 32,
    
    // Formato do user handle
    handleFormat: process.env.WEBAUTHN_USER_HANDLE_FORMAT || 'base64url', // base64url, hex, uuid
    
    // Verificação de user handle
    enableUserHandleVerification: process.env.WEBAUTHN_ENABLE_USER_HANDLE_VERIFICATION !== 'false',
    
    // Display name personalizado
    enableCustomDisplayName: process.env.WEBAUTHN_ENABLE_CUSTOM_DISPLAY_NAME !== 'false',
    maxDisplayNameLength: parseInt(process.env.WEBAUTHN_MAX_DISPLAY_NAME_LENGTH) || 100
  },

  // ========================================
  // CREDENTIAL CONFIGURATION
  // ========================================
  credentials: {
    // Máximo de credenciais por usuário
    maxCredentialsPerUser: parseInt(process.env.WEBAUTHN_MAX_CREDENTIALS_PER_USER) || 10,
    
    // Exclusão automática de credenciais antigas
    enableAutoCleanup: process.env.WEBAUTHN_ENABLE_AUTO_CLEANUP === 'true',
    maxCredentialAge: parseInt(process.env.WEBAUTHN_MAX_CREDENTIAL_AGE) || 31536000, // 1 ano
    
    // Verificação de credencial duplicada
    preventDuplicateCredentials: process.env.WEBAUTHN_PREVENT_DUPLICATE_CREDENTIALS !== 'false',
    
    // Backup de credenciais
    enableCredentialBackup: process.env.WEBAUTHN_ENABLE_CREDENTIAL_BACKUP === 'true',
    backupEncryption: process.env.WEBAUTHN_BACKUP_ENCRYPTION !== 'false'
  },

  // ========================================
  // INTEGRATION CONFIGURATION
  // ========================================
  integration: {
    // Integração com serviços de risco
    enableRiskAssessment: process.env.WEBAUTHN_ENABLE_RISK_ASSESSMENT !== 'false',
    riskServiceUrl: process.env.WEBAUTHN_RISK_SERVICE_URL,
    riskServiceTimeout: parseInt(process.env.WEBAUTHN_RISK_SERVICE_TIMEOUT) || 5000,
    
    // Integração com auditoria
    enableAuditLogging: process.env.WEBAUTHN_ENABLE_AUDIT_LOGGING !== 'false',
    auditDetailLevel: process.env.WEBAUTHN_AUDIT_DETAIL_LEVEL || 'standard', // minimal, standard, detailed
    
    // Integração com métricas
    enableMetrics: process.env.WEBAUTHN_ENABLE_METRICS !== 'false',
    metricsPrefix: process.env.WEBAUTHN_METRICS_PREFIX || 'webauthn_',
    
    // Integração com notificações
    enableNotifications: process.env.WEBAUTHN_ENABLE_NOTIFICATIONS === 'true',
    notificationServiceUrl: process.env.WEBAUTHN_NOTIFICATION_SERVICE_URL
  },

  // ========================================
  // BROWSER COMPATIBILITY
  // ========================================
  compatibility: {
    // Suporte a navegadores legados
    enableLegacySupport: process.env.WEBAUTHN_ENABLE_LEGACY_SUPPORT === 'true',
    
    // Detecção de capacidades do navegador
    enableCapabilityDetection: process.env.WEBAUTHN_ENABLE_CAPABILITY_DETECTION !== 'false',
    
    // Fallback para autenticação tradicional
    enableFallback: process.env.WEBAUTHN_ENABLE_FALLBACK !== 'false',
    fallbackMethods: process.env.WEBAUTHN_FALLBACK_METHODS?.split(',') || ['password', 'otp'],
    
    // Suporte a extensões específicas do navegador
    browserExtensions: {
      chrome: {
        enableEnterpriseAttestation: process.env.WEBAUTHN_CHROME_ENTERPRISE_ATTESTATION === 'true',
        enableLargeBlob: process.env.WEBAUTHN_CHROME_LARGE_BLOB === 'true'
      },
      firefox: {
        enableResidentKeys: process.env.WEBAUTHN_FIREFOX_RESIDENT_KEYS !== 'false'
      },
      safari: {
        enableTouchID: process.env.WEBAUTHN_SAFARI_TOUCHID !== 'false',
        enableFaceID: process.env.WEBAUTHN_SAFARI_FACEID !== 'false'
      },
      edge: {
        enableWindowsHello: process.env.WEBAUTHN_EDGE_WINDOWS_HELLO !== 'false'
      }
    }
  },

  // ========================================
  // PLATFORM SPECIFIC CONFIGURATION
  // ========================================
  platform: {
    // Configurações específicas para iOS
    ios: {
      enableTouchID: process.env.WEBAUTHN_IOS_TOUCHID !== 'false',
      enableFaceID: process.env.WEBAUTHN_IOS_FACEID !== 'false',
      requireBiometrics: process.env.WEBAUTHN_IOS_REQUIRE_BIOMETRICS === 'true'
    },
    
    // Configurações específicas para Android
    android: {
      enableFingerprint: process.env.WEBAUTHN_ANDROID_FINGERPRINT !== 'false',
      enableFace: process.env.WEBAUTHN_ANDROID_FACE !== 'false',
      enableIris: process.env.WEBAUTHN_ANDROID_IRIS === 'true',
      requireBiometrics: process.env.WEBAUTHN_ANDROID_REQUIRE_BIOMETRICS === 'true'
    },
    
    // Configurações específicas para Windows
    windows: {
      enableWindowsHello: process.env.WEBAUTHN_WINDOWS_HELLO !== 'false',
      enablePIN: process.env.WEBAUTHN_WINDOWS_PIN !== 'false',
      requireTPM: process.env.WEBAUTHN_WINDOWS_REQUIRE_TPM === 'true'
    },
    
    // Configurações específicas para macOS
    macos: {
      enableTouchID: process.env.WEBAUTHN_MACOS_TOUCHID !== 'false',
      enableSecureEnclave: process.env.WEBAUTHN_MACOS_SECURE_ENCLAVE !== 'false'
    }
  },

  // ========================================
  // DEVELOPMENT & TESTING
  // ========================================
  development: {
    // Modo de desenvolvimento
    enableDevMode: process.env.NODE_ENV === 'development' && 
      process.env.WEBAUTHN_ENABLE_DEV_MODE === 'true',
    
    // Logging detalhado
    enableVerboseLogging: process.env.WEBAUTHN_ENABLE_VERBOSE_LOGGING === 'true',
    
    // Simulação de autenticadores
    enableSimulator: process.env.WEBAUTHN_ENABLE_SIMULATOR === 'true',
    simulatorConfig: {
      defaultAuthenticator: process.env.WEBAUTHN_SIMULATOR_DEFAULT || 'platform',
      enableAllTransports: process.env.WEBAUTHN_SIMULATOR_ALL_TRANSPORTS === 'true'
    },
    
    // Bypass de verificações em desenvolvimento
    bypassOriginValidation: process.env.NODE_ENV === 'development' && 
      process.env.WEBAUTHN_BYPASS_ORIGIN_VALIDATION === 'true',
    bypassAttestationVerification: process.env.NODE_ENV === 'development' && 
      process.env.WEBAUTHN_BYPASS_ATTESTATION_VERIFICATION === 'true'
  },

  // ========================================
  // MONITORING & OBSERVABILITY
  // ========================================
  monitoring: {
    // Métricas de performance
    enablePerformanceMetrics: process.env.WEBAUTHN_ENABLE_PERFORMANCE_METRICS !== 'false',
    performanceThresholds: {
      registrationTime: parseInt(process.env.WEBAUTHN_REGISTRATION_TIME_THRESHOLD) || 10000, // 10s
      authenticationTime: parseInt(process.env.WEBAUTHN_AUTHENTICATION_TIME_THRESHOLD) || 5000, // 5s
      challengeGenerationTime: parseInt(process.env.WEBAUTHN_CHALLENGE_GENERATION_THRESHOLD) || 100 // 100ms
    },
    
    // Alertas
    enableAlerting: process.env.WEBAUTHN_ENABLE_ALERTING === 'true',
    alertThresholds: {
      failureRate: parseFloat(process.env.WEBAUTHN_FAILURE_RATE_THRESHOLD) || 0.1, // 10%
      responseTime: parseInt(process.env.WEBAUTHN_RESPONSE_TIME_THRESHOLD) || 5000, // 5s
      errorRate: parseFloat(process.env.WEBAUTHN_ERROR_RATE_THRESHOLD) || 0.05 // 5%
    },
    
    // Health checks
    enableHealthChecks: process.env.WEBAUTHN_ENABLE_HEALTH_CHECKS !== 'false',
    healthCheckInterval: parseInt(process.env.WEBAUTHN_HEALTH_CHECK_INTERVAL) || 30000 // 30s
  },

  // ========================================
  // COMPLIANCE & REGULATORY
  // ========================================
  compliance: {
    // Frameworks de compliance
    enabledFrameworks: process.env.WEBAUTHN_COMPLIANCE_FRAMEWORKS?.split(',') || [
      'FIDO2',
      'NIST_SP_800_63B',
      'PCI_DSS',
      'GDPR',
      'LGPD'
    ],
    
    // Retenção de dados
    dataRetention: {
      challengeRetention: parseInt(process.env.WEBAUTHN_CHALLENGE_RETENTION) || 86400, // 24 horas
      auditLogRetention: parseInt(process.env.WEBAUTHN_AUDIT_RETENTION) || 2555, // 7 anos
      credentialMetadataRetention: parseInt(process.env.WEBAUTHN_CREDENTIAL_METADATA_RETENTION) || 2555 // 7 anos
    },
    
    // Privacidade
    enablePrivacyMode: process.env.WEBAUTHN_ENABLE_PRIVACY_MODE === 'true',
    anonymizeUserData: process.env.WEBAUTHN_ANONYMIZE_USER_DATA === 'true',
    
    // Auditoria
    enableComplianceAudit: process.env.WEBAUTHN_ENABLE_COMPLIANCE_AUDIT !== 'false',
    complianceReportingInterval: parseInt(process.env.WEBAUTHN_COMPLIANCE_REPORTING_INTERVAL) || 86400 // 24 horas
  }
}));