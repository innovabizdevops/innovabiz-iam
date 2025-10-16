/**
 * Adaptador Regional para Framework de Autenticação
 * 
 * Este componente é responsável por adaptar o comportamento do framework de autenticação
 * de acordo com a região geográfica do usuário, aplicando regras de conformidade
 * específicas para cada região suportada (UE/Portugal, Brasil, Angola, EUA).
 * 
 * @module regional-adapter
 * @author INNOVABIZ
 * @copyright 2025 INNOVABIZ
 * @license Proprietary
 */

const { logger } = require('../utils/logging');
const geoip = require('geoip-lite'); // Simulação - seria substituído por solução robusta

/**
 * Classe de adaptação regional
 */
class RegionalAdapter {
  /**
   * Construtor do adaptador regional
   * 
   * @param {Object} regionMapping Mapeamento de países para regiões
   */
  constructor(regionMapping = {}) {
    // Definição padrão de mapeamento de países para regiões regulatórias
    this.regionMapping = {
      // Europa
      'PT': 'EU', 'ES': 'EU', 'FR': 'EU', 'DE': 'EU', 'IT': 'EU',
      'BE': 'EU', 'NL': 'EU', 'LU': 'EU', 'DK': 'EU', 'SE': 'EU',
      'FI': 'EU', 'AT': 'EU', 'IE': 'EU', 'GR': 'EU', 'CY': 'EU',
      
      // Brasil
      'BR': 'BR',
      
      // Angola e outros países africanos
      'AO': 'AO', 'MZ': 'AO', 'CV': 'AO', 'GW': 'AO', 'ST': 'AO',
      
      // Estados Unidos
      'US': 'US', 'CA': 'US',
      
      // Mapeamento personalizado substituirá os padrões
      ...regionMapping
    };
    
    // Configurações de conformidade por região
    this.complianceSettings = {
      'EU': {
        dataProtection: 'GDPR',
        identityVerification: 'eIDAS',
        dataRetention: {
          default: 30 * 24 * 60 * 60, // 30 dias em segundos
          maxAllowed: 365 * 24 * 60 * 60, // 1 ano em segundos
          requiresExplicitConsent: true
        },
        userConsent: {
          required: true,
          granular: true,
          withdrawable: true
        },
        biometricDataRules: {
          explicitConsent: true,
          specialCategory: true,
          encryption: 'required',
          localProcessingPreferred: true
        },
        childrenRestrictions: {
          minimumAge: 16,
          parentalConsent: true
        }
      },
      'BR': {
        dataProtection: 'LGPD',
        identityVerification: 'ICP-Brasil',
        dataRetention: {
          default: 30 * 24 * 60 * 60, // 30 dias em segundos
          maxAllowed: 365 * 24 * 60 * 60, // 1 ano em segundos
          requiresExplicitConsent: true
        },
        userConsent: {
          required: true,
          granular: true,
          withdrawable: true
        },
        biometricDataRules: {
          explicitConsent: true,
          specialCategory: true,
          encryption: 'required',
          localProcessingPreferred: true
        },
        childrenRestrictions: {
          minimumAge: 12,
          parentalConsent: true
        }
      },
      'AO': {
        dataProtection: 'PNDSB',
        identityVerification: 'National',
        dataRetention: {
          default: 60 * 24 * 60 * 60, // 60 dias em segundos
          maxAllowed: 365 * 24 * 60 * 60, // 1 ano em segundos
          requiresExplicitConsent: true
        },
        userConsent: {
          required: true,
          granular: false,
          withdrawable: true
        },
        biometricDataRules: {
          explicitConsent: true,
          specialCategory: true,
          encryption: 'required',
          localProcessingPreferred: false
        },
        childrenRestrictions: {
          minimumAge: 18,
          parentalConsent: true
        }
      },
      'US': {
        dataProtection: 'SectorSpecific',
        identityVerification: 'NIST',
        dataRetention: {
          default: 90 * 24 * 60 * 60, // 90 dias em segundos
          maxAllowed: 730 * 24 * 60 * 60, // 2 anos em segundos
          requiresExplicitConsent: false
        },
        userConsent: {
          required: true,
          granular: false,
          withdrawable: true
        },
        biometricDataRules: {
          explicitConsent: true,
          specialCategory: false,
          encryption: 'required',
          localProcessingPreferred: false
        },
        childrenRestrictions: {
          minimumAge: 13,
          parentalConsent: true
        }
      }
    };
    
    // Configurações de autenticação por região
    this.authSettings = {
      'EU': {
        mfaPolicy: 'recommended',
        passwordRequirements: {
          minLength: 12,
          requireSpecialChars: true,
          requireNumbers: true,
          requireUpperLower: true,
          maxAge: 90 // dias
        },
        sessionDuration: 30 * 60, // 30 minutos em segundos
        failedAttemptsLimit: 5,
        cooldownPeriod: 15 * 60, // 15 minutos em segundos
        ipRateLimit: {
          enabled: true,
          limit: 10, // tentativas
          window: 5 * 60 // 5 minutos em segundos
        },
        preferredAuthMethods: ['FIDO2', 'SmartCard', 'OTP'],
        privacyByDefault: true,
        userNotifications: {
          loginAttempts: true,
          passwordChanges: true,
          profileUpdates: true
        }
      },
      'BR': {
        mfaPolicy: 'recommended',
        passwordRequirements: {
          minLength: 10,
          requireSpecialChars: true,
          requireNumbers: true,
          requireUpperLower: true,
          maxAge: 60 // dias
        },
        sessionDuration: 45 * 60, // 45 minutos em segundos
        failedAttemptsLimit: 5,
        cooldownPeriod: 10 * 60, // 10 minutos em segundos
        ipRateLimit: {
          enabled: true,
          limit: 15, // tentativas
          window: 5 * 60 // 5 minutos em segundos
        },
        preferredAuthMethods: ['OTP', 'FIDO2', 'PushNotification'],
        privacyByDefault: true,
        userNotifications: {
          loginAttempts: true,
          passwordChanges: true,
          profileUpdates: true
        }
      },
      'AO': {
        mfaPolicy: 'optional',
        passwordRequirements: {
          minLength: 8,
          requireSpecialChars: false,
          requireNumbers: true,
          requireUpperLower: true,
          maxAge: 90 // dias
        },
        sessionDuration: 60 * 60, // 60 minutos em segundos
        failedAttemptsLimit: 7,
        cooldownPeriod: 10 * 60, // 10 minutos em segundos
        ipRateLimit: {
          enabled: true,
          limit: 20, // tentativas
          window: 10 * 60 // 10 minutos em segundos
        },
        preferredAuthMethods: ['OTP', 'Password', 'PushNotification'],
        privacyByDefault: false,
        userNotifications: {
          loginAttempts: true,
          passwordChanges: true,
          profileUpdates: false
        }
      },
      'US': {
        mfaPolicy: 'configurable',
        passwordRequirements: {
          minLength: 8,
          requireSpecialChars: true,
          requireNumbers: true,
          requireUpperLower: true,
          maxAge: 120 // dias
        },
        sessionDuration: 120 * 60, // 120 minutos em segundos
        failedAttemptsLimit: 10,
        cooldownPeriod: 5 * 60, // 5 minutos em segundos
        ipRateLimit: {
          enabled: true,
          limit: 30, // tentativas
          window: 15 * 60 // 15 minutos em segundos
        },
        preferredAuthMethods: ['FIDO2', 'OTP', 'Password'],
        privacyByDefault: false,
        userNotifications: {
          loginAttempts: false,
          passwordChanges: true,
          profileUpdates: false
        }
      }
    };
    
    logger.info('[RegionalAdapter] Adaptador regional inicializado');
  }
  
  /**
   * Detecta a região de uma requisição
   * 
   * @param {Object} request Requisição de autenticação
   * @param {Object} context Contexto adicional
   * @returns {string} Código da região detectada
   */
  detectRegion(request, context = {}) {
    // Prioridade 1: Região explicitamente informada no contexto
    if (context.region && this.complianceSettings[context.region]) {
      logger.debug(`[RegionalAdapter] Região explícita no contexto: ${context.region}`);
      return context.region;
    }
    
    // Prioridade 2: Configuração do tenant
    if (request.tenantId && context.tenantRegion) {
      logger.debug(`[RegionalAdapter] Região do tenant: ${context.tenantRegion}`);
      return context.tenantRegion;
    }
    
    // Prioridade 3: Cabeçalho Accept-Language
    if (request.headers && request.headers['accept-language']) {
      const language = request.headers['accept-language'].split(',')[0].trim().split('-')[0];
      if (language === 'pt') {
        // Distingue entre português de Portugal e Brasil
        if (request.headers['accept-language'].includes('pt-BR')) {
          logger.debug('[RegionalAdapter] Região detectada por idioma: BR');
          return 'BR';
        } else if (request.headers['accept-language'].includes('pt-PT')) {
          logger.debug('[RegionalAdapter] Região detectada por idioma: EU');
          return 'EU';
        }
      }
    }
    
    // Prioridade 4: Geolocalização por IP
    if (request.ip) {
      try {
        // Em produção, usaria um serviço mais robusto
        const geo = geoip.lookup(request.ip);
        if (geo && geo.country) {
          const region = this.mapCountryToRegion(geo.country);
          logger.debug(`[RegionalAdapter] Região detectada por IP: ${region} (país: ${geo.country})`);
          return region;
        }
      } catch (error) {
        logger.error(`[RegionalAdapter] Erro ao detectar região por IP: ${error.message}`);
      }
    }
    
    // Fallback: região padrão
    logger.debug(`[RegionalAdapter] Usando região padrão: EU`);
    return 'EU';
  }
  
  /**
   * Mapeia código de país para região regulatória
   * 
   * @param {string} countryCode Código ISO do país
   * @returns {string} Código da região
   */
  mapCountryToRegion(countryCode) {
    return this.regionMapping[countryCode] || 'EU'; // Padrão para EU como mais restritivo
  }
  
  /**
   * Obtém configurações de conformidade para uma região
   * 
   * @param {string} region Código da região
   * @returns {Object} Configurações de conformidade
   */
  getComplianceSettings(region) {
    return this.complianceSettings[region] || this.complianceSettings['EU'];
  }
  
  /**
   * Obtém configurações de autenticação para uma região
   * 
   * @param {string} region Código da região
   * @returns {Object} Configurações de autenticação
   */
  getAuthSettings(region) {
    return this.authSettings[region] || this.authSettings['EU'];
  }
  
  /**
   * Adapta uma requisição de autenticação para conformidade regional
   * 
   * @param {Object} request Requisição original
   * @param {string} region Código da região
   * @returns {Object} Requisição adaptada
   */
  adaptRequest(request, region) {
    const settings = this.getAuthSettings(region);
    const compliance = this.getComplianceSettings(region);
    
    // Clona a requisição para não modificar o original
    const adaptedRequest = { ...request };
    
    // Adiciona informações regionais
    adaptedRequest.regional = {
      region,
      dataProtection: compliance.dataProtection,
      mfaPolicy: settings.mfaPolicy,
      sessionDuration: settings.sessionDuration,
      failedAttemptsLimit: settings.failedAttemptsLimit,
      privacyByDefault: settings.privacyByDefault
    };
    
    // Adapta validações de senha conforme região
    if (adaptedRequest.password) {
      adaptedRequest.passwordRequirements = settings.passwordRequirements;
    }
    
    // Adapta retenção de dados conforme região
    adaptedRequest.dataRetention = compliance.dataRetention;
    
    // Implementa regras especiais por região
    this.applyRegionalRules(adaptedRequest, region);
    
    logger.debug(`[RegionalAdapter] Requisição adaptada para região: ${region}`);
    return adaptedRequest;
  }
  
  /**
   * Aplica regras especiais específicas por região
   * 
   * @param {Object} request Requisição a adaptar
   * @param {string} region Código da região
   */
  applyRegionalRules(request, region) {
    switch (region) {
      case 'EU':
        // Regras específicas para GDPR/eIDAS
        request.explicitConsentRequired = true;
        request.rightToBeRemembered = false; // Por padrão, não lembrar
        request.dataMinimization = true;
        
        // Adaptações para métodos biométricos na UE
        if (request.methodType === 'biometric') {
          request.requireExplicitBiometricConsent = true;
          request.localProcessingPreferred = true;
        }
        break;
        
      case 'BR':
        // Regras específicas para LGPD
        request.explicitConsentRequired = true;
        request.rightToBeRemembered = false; // Por padrão, não lembrar
        
        // Adaptações para o Brasil (ex: integração com gov.br)
        if (request.identityProvider) {
          request.supportGovBr = true;
        }
        break;
        
      case 'AO':
        // Regras específicas para Angola
        request.explicitConsentRequired = true;
        request.alternativeAuthMethodsRequired = true; // Suporte a métodos alternativos
        request.offlineCapabilityPreferred = true; // Para áreas com conectividade limitada
        break;
        
      case 'US':
        // Regras específicas para EUA
        request.sectorSpecificRules = true;
        
        // Conformidade com indústrias específicas
        if (request.industry === 'healthcare') {
          request.hipaaCompliance = true;
        } else if (request.industry === 'finance') {
          request.glbaCompliance = true;
        }
        break;
    }
  }
  
  /**
   * Valida a conformidade de uma solicitação de autenticação
   * 
   * @param {Object} request Requisição a validar
   * @param {string} region Código da região
   * @returns {Object} Resultado da validação
   */
  validateRequestCompliance(request, region) {
    const settings = this.getAuthSettings(region);
    const compliance = this.getComplianceSettings(region);
    const validationResults = {
      compliant: true,
      issues: []
    };
    
    // Validações comuns a todas as regiões
    
    // Verificação de consentimento quando necessário
    if (compliance.userConsent.required && !request.userConsent) {
      validationResults.compliant = false;
      validationResults.issues.push({
        code: 'MISSING_CONSENT',
        message: 'Consentimento do usuário é obrigatório nesta região',
        severity: 'high'
      });
    }
    
    // Verificações específicas por região
    switch (region) {
      case 'EU':
        this.validateEUCompliance(request, validationResults, compliance, settings);
        break;
      case 'BR':
        this.validateBRCompliance(request, validationResults, compliance, settings);
        break;
      case 'AO':
        this.validateAOCompliance(request, validationResults, compliance, settings);
        break;
      case 'US':
        this.validateUSCompliance(request, validationResults, compliance, settings);
        break;
    }
    
    return validationResults;
  }
  
  /**
   * Valida conformidade para a União Europeia
   */
  validateEUCompliance(request, results, compliance, settings) {
    // Verificações específicas GDPR/eIDAS
    
    // Biometria requer consentimento explícito
    if (request.methodType === 'biometric' && !request.explicitBiometricConsent) {
      results.compliant = false;
      results.issues.push({
        code: 'EU_BIOMETRIC_CONSENT',
        message: 'Métodos biométricos requerem consentimento explícito sob GDPR',
        severity: 'high'
      });
    }
    
    // Verificação de nível de garantia para eIDAS
    if (request.assuranceLevel && !['low', 'substantial', 'high'].includes(request.assuranceLevel)) {
      results.issues.push({
        code: 'EIDAS_ASSURANCE_LEVEL',
        message: 'Nível de garantia não compatível com eIDAS',
        severity: 'medium'
      });
    }
  }
  
  /**
   * Valida conformidade para o Brasil
   */
  validateBRCompliance(request, results, compliance, settings) {
    // Verificações específicas LGPD
    
    // Verificação de consentimento granular
    if (compliance.userConsent.granular && request.userConsent && !request.granularConsent) {
      results.issues.push({
        code: 'BR_GRANULAR_CONSENT',
        message: 'Consentimento deve ser granular sob LGPD',
        severity: 'medium'
      });
    }
  }
  
  /**
   * Valida conformidade para Angola
   */
  validateAOCompliance(request, results, compliance, settings) {
    // Verificações específicas para Angola
    
    // Verificação de adaptações para conectividade limitada
    if (request.requiresConstantConnectivity && !request.alternativeAuthMethods) {
      results.issues.push({
        code: 'AO_CONNECTIVITY',
        message: 'Recomendado fornecer métodos alternativos para áreas com conectividade limitada',
        severity: 'medium'
      });
    }
  }
  
  /**
   * Valida conformidade para os Estados Unidos
   */
  validateUSCompliance(request, results, compliance, settings) {
    // Verificações específicas para EUA
    
    // Verificações específicas de indústria
    if (request.industry === 'healthcare' && !request.hipaaCompliant) {
      results.issues.push({
        code: 'US_HIPAA',
        message: 'Autenticação para saúde deve ser compatível com HIPAA',
        severity: 'high'
      });
    } else if (request.industry === 'finance' && !request.glbaCompliant) {
      results.issues.push({
        code: 'US_GLBA',
        message: 'Autenticação para finanças deve ser compatível com GLBA',
        severity: 'high'
      });
    }
  }
}

module.exports = RegionalAdapter;
