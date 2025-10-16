/**
 * ⚙️ IAM CONFIGURATION - INNOVABIZ PLATFORM
 * Configurações centralizadas do módulo IAM
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: NIST SP 800-63B, OWASP Configuration Guide
 * Security: Environment-based Configuration, Secure Defaults
 */

import { registerAs } from '@nestjs/config';

export default registerAs('iam', () => ({
  // ========================================
  // JWT CONFIGURATION
  // ========================================
  jwt: {
    secret: process.env.JWT_SECRET || 'your-super-secret-jwt-key-change-in-production',
    expiresIn: process.env.JWT_EXPIRES_IN || '1h',
    refreshExpiresIn: process.env.JWT_REFRESH_EXPIRES_IN || '7d',
    issuer: process.env.JWT_ISSUER || 'innovabiz-iam',
    audience: process.env.JWT_AUDIENCE || 'innovabiz-platform',
    algorithm: process.env.JWT_ALGORITHM || 'HS256',
    maxAge: parseInt(process.env.JWT_MAX_AGE) || 86400, // 24 horas
    clockTolerance: parseInt(process.env.JWT_CLOCK_TOLERANCE) || 30, // 30 segundos
    ignoreExpiration: process.env.NODE_ENV === 'development' ? 
      process.env.JWT_IGNORE_EXPIRATION === 'true' : false
  },

  // ========================================
  // SESSION CONFIGURATION
  // ========================================
  session: {
    maxConcurrentSessions: parseInt(process.env.MAX_CONCURRENT_SESSIONS) || 5,
    sessionTimeout: parseInt(process.env.SESSION_TIMEOUT) || 3600, // 1 hora
    extendedSessionTimeout: parseInt(process.env.EXTENDED_SESSION_TIMEOUT) || 86400, // 24 horas
    rememberMeTimeout: parseInt(process.env.REMEMBER_ME_TIMEOUT) || 2592000, // 30 dias
    sessionCleanupInterval: parseInt(process.env.SESSION_CLEANUP_INTERVAL) || 300, // 5 minutos
    enableSessionExtension: process.env.ENABLE_SESSION_EXTENSION !== 'false',
    requireReauthForSensitive: process.env.REQUIRE_REAUTH_SENSITIVE !== 'false'
  },

  // ========================================
  // SECURITY CONFIGURATION
  // ========================================
  security: {
    // Rate Limiting
    maxRequestsPerMinute: parseInt(process.env.MAX_REQUESTS_PER_MINUTE) || 100,
    maxLoginAttempts: parseInt(process.env.MAX_LOGIN_ATTEMPTS) || 5,
    lockoutDuration: parseInt(process.env.LOCKOUT_DURATION) || 900, // 15 minutos
    
    // Password Policy (se aplicável para fallback)
    minPasswordLength: parseInt(process.env.MIN_PASSWORD_LENGTH) || 12,
    requireUppercase: process.env.REQUIRE_UPPERCASE !== 'false',
    requireLowercase: process.env.REQUIRE_LOWERCASE !== 'false',
    requireNumbers: process.env.REQUIRE_NUMBERS !== 'false',
    requireSpecialChars: process.env.REQUIRE_SPECIAL_CHARS !== 'false',
    
    // HTTPS & Security Headers
    httpsOnly: process.env.HTTPS_ONLY !== 'false',
    allowedDomains: process.env.ALLOWED_DOMAINS?.split(',') || [],
    trustedProxies: process.env.TRUSTED_PROXIES?.split(',') || [],
    
    // Content Security Policy
    cspReportUri: process.env.CSP_REPORT_URI,
    enableCSPReporting: process.env.ENABLE_CSP_REPORTING === 'true'
  },

  // ========================================
  // AUDIT & LOGGING CONFIGURATION
  // ========================================
  audit: {
    enableAuditLogging: process.env.ENABLE_AUDIT_LOGGING !== 'false',
    auditLogLevel: process.env.AUDIT_LOG_LEVEL || 'info',
    auditRetentionDays: parseInt(process.env.AUDIT_RETENTION_DAYS) || 2555, // 7 anos
    enableRealTimeAudit: process.env.ENABLE_REALTIME_AUDIT === 'true',
    auditBatchSize: parseInt(process.env.AUDIT_BATCH_SIZE) || 100,
    auditFlushInterval: parseInt(process.env.AUDIT_FLUSH_INTERVAL) || 5000, // 5 segundos
    
    // Compliance Frameworks
    complianceFrameworks: process.env.COMPLIANCE_FRAMEWORKS?.split(',') || [
      'NIST_SP_800_63B',
      'OWASP_ASVS',
      'ISO_27001',
      'GDPR',
      'LGPD',
      'PCI_DSS'
    ],
    
    // Sensitive Data Masking
    enableDataMasking: process.env.ENABLE_DATA_MASKING !== 'false',
    maskingPatterns: {
      email: process.env.MASK_EMAIL_PATTERN || '***@***.***',
      phone: process.env.MASK_PHONE_PATTERN || '***-***-****',
      document: process.env.MASK_DOCUMENT_PATTERN || '***.***.***-**'
    }
  },

  // ========================================
  // MULTI-TENANT CONFIGURATION
  // ========================================
  multiTenant: {
    enableMultiTenancy: process.env.ENABLE_MULTI_TENANCY !== 'false',
    defaultTenantId: process.env.DEFAULT_TENANT_ID || 'default',
    tenantIsolationLevel: process.env.TENANT_ISOLATION_LEVEL || 'strict', // strict, moderate, basic
    enableTenantSubdomains: process.env.ENABLE_TENANT_SUBDOMAINS === 'true',
    maxTenantsPerUser: parseInt(process.env.MAX_TENANTS_PER_USER) || 10,
    tenantConfigCache: parseInt(process.env.TENANT_CONFIG_CACHE) || 300 // 5 minutos
  },

  // ========================================
  // PERFORMANCE & CACHING
  // ========================================
  performance: {
    enableCaching: process.env.ENABLE_CACHING !== 'false',
    cacheDefaultTTL: parseInt(process.env.CACHE_DEFAULT_TTL) || 300, // 5 minutos
    cacheMaxItems: parseInt(process.env.CACHE_MAX_ITEMS) || 1000,
    
    // Database Connection Pool
    dbPoolSize: parseInt(process.env.DB_POOL_SIZE) || 10,
    dbConnectionTimeout: parseInt(process.env.DB_CONNECTION_TIMEOUT) || 30000,
    dbQueryTimeout: parseInt(process.env.DB_QUERY_TIMEOUT) || 15000,
    
    // Request Processing
    requestTimeout: parseInt(process.env.REQUEST_TIMEOUT) || 30000,
    maxRequestSize: process.env.MAX_REQUEST_SIZE || '10mb',
    enableCompression: process.env.ENABLE_COMPRESSION !== 'false'
  },

  // ========================================
  // MONITORING & METRICS
  // ========================================
  monitoring: {
    enableMetrics: process.env.ENABLE_METRICS !== 'false',
    metricsPort: parseInt(process.env.METRICS_PORT) || 9090,
    metricsPath: process.env.METRICS_PATH || '/metrics',
    
    // Health Checks
    enableHealthChecks: process.env.ENABLE_HEALTH_CHECKS !== 'false',
    healthCheckInterval: parseInt(process.env.HEALTH_CHECK_INTERVAL) || 30000,
    healthCheckTimeout: parseInt(process.env.HEALTH_CHECK_TIMEOUT) || 5000,
    
    // Alerting
    enableAlerting: process.env.ENABLE_ALERTING === 'true',
    alertWebhookUrl: process.env.ALERT_WEBHOOK_URL,
    alertThresholds: {
      errorRate: parseFloat(process.env.ALERT_ERROR_RATE) || 0.05, // 5%
      responseTime: parseInt(process.env.ALERT_RESPONSE_TIME) || 2000, // 2 segundos
      memoryUsage: parseFloat(process.env.ALERT_MEMORY_USAGE) || 0.85 // 85%
    }
  },

  // ========================================
  // INTEGRATION CONFIGURATION
  // ========================================
  integration: {
    // API Gateway
    apiGatewayUrl: process.env.API_GATEWAY_URL,
    apiGatewayTimeout: parseInt(process.env.API_GATEWAY_TIMEOUT) || 10000,
    
    // External Services
    externalServices: {
      riskService: {
        url: process.env.RISK_SERVICE_URL,
        timeout: parseInt(process.env.RISK_SERVICE_TIMEOUT) || 5000,
        retries: parseInt(process.env.RISK_SERVICE_RETRIES) || 3
      },
      notificationService: {
        url: process.env.NOTIFICATION_SERVICE_URL,
        timeout: parseInt(process.env.NOTIFICATION_SERVICE_TIMEOUT) || 5000
      },
      analyticsService: {
        url: process.env.ANALYTICS_SERVICE_URL,
        timeout: parseInt(process.env.ANALYTICS_SERVICE_TIMEOUT) || 10000
      }
    },
    
    // Message Queue
    messageQueue: {
      enabled: process.env.ENABLE_MESSAGE_QUEUE === 'true',
      url: process.env.MESSAGE_QUEUE_URL,
      exchangeName: process.env.MQ_EXCHANGE_NAME || 'iam-events',
      queueName: process.env.MQ_QUEUE_NAME || 'iam-queue'
    }
  },

  // ========================================
  // DEVELOPMENT & DEBUGGING
  // ========================================
  development: {
    enableDebugMode: process.env.NODE_ENV === 'development' && 
      process.env.ENABLE_DEBUG_MODE === 'true',
    enableVerboseLogging: process.env.ENABLE_VERBOSE_LOGGING === 'true',
    enableTestEndpoints: process.env.NODE_ENV !== 'production' && 
      process.env.ENABLE_TEST_ENDPOINTS === 'true',
    mockExternalServices: process.env.MOCK_EXTERNAL_SERVICES === 'true',
    
    // Testing
    testDataRetention: parseInt(process.env.TEST_DATA_RETENTION) || 86400, // 24 horas
    enableTestDataCleanup: process.env.ENABLE_TEST_DATA_CLEANUP !== 'false'
  },

  // ========================================
  // FEATURE FLAGS
  // ========================================
  features: {
    enableWebAuthn: process.env.ENABLE_WEBAUTHN !== 'false',
    enableBiometrics: process.env.ENABLE_BIOMETRICS === 'true',
    enableSocialLogin: process.env.ENABLE_SOCIAL_LOGIN === 'true',
    enablePasswordless: process.env.ENABLE_PASSWORDLESS === 'true',
    enableRiskBasedAuth: process.env.ENABLE_RISK_BASED_AUTH !== 'false',
    enableMFA: process.env.ENABLE_MFA !== 'false',
    enableSSOIntegration: process.env.ENABLE_SSO_INTEGRATION === 'true',
    enableFederatedAuth: process.env.ENABLE_FEDERATED_AUTH === 'true',
    enableAdvancedAudit: process.env.ENABLE_ADVANCED_AUDIT === 'true',
    enableMLRiskScoring: process.env.ENABLE_ML_RISK_SCORING === 'true'
  },

  // ========================================
  // COMPLIANCE & REGULATORY
  // ========================================
  compliance: {
    dataResidency: process.env.DATA_RESIDENCY || 'global',
    encryptionAtRest: process.env.ENCRYPTION_AT_REST !== 'false',
    encryptionInTransit: process.env.ENCRYPTION_IN_TRANSIT !== 'false',
    keyRotationInterval: parseInt(process.env.KEY_ROTATION_INTERVAL) || 2592000, // 30 dias
    
    // GDPR/LGPD
    enableRightToBeDeleted: process.env.ENABLE_RIGHT_TO_BE_DELETED !== 'false',
    enableDataPortability: process.env.ENABLE_DATA_PORTABILITY !== 'false',
    dataProcessingLawfulBasis: process.env.DATA_PROCESSING_LAWFUL_BASIS || 'consent',
    
    // Industry Standards
    pciDssCompliance: process.env.PCI_DSS_COMPLIANCE === 'true',
    hipaaCompliance: process.env.HIPAA_COMPLIANCE === 'true',
    sox404Compliance: process.env.SOX_404_COMPLIANCE === 'true'
  }
}));