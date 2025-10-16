/**
 * @fileoverview Cliente OPA (Open Policy Agent) para a plataforma INNOVABIZ
 * @author INNOVABIZ Dev Team
 * @version 1.0.0
 * @module iam/seguranca/opa/opa-client
 * @description Cliente para integração com OPA para autorização baseada em políticas
 */

const axios = require('axios');
const logger = require('../../core/utils/logging');
const { maskSensitiveData } = require('../../core/utils/data-masking');

/**
 * Configurações regionais para OPA
 */
const REGIONAL_CONFIGS = {
  EU: {
    opa_url: process.env.OPA_URL_EU || 'http://opa-eu.innovabiz.internal:8181',
    policy_paths: {
      iam: 'v1/data/innovabiz/eu/iam',
      healthcare: 'v1/data/innovabiz/eu/healthcare',
      auth: 'v1/data/innovabiz/eu/auth'
    },
    cache_ttl_ms: 60000, // 1 minuto
    max_retries: 3,
    timeout_ms: 500,
    headers: {
      'X-Region-Code': 'EU'
    },
    compliance: {
      gdpr: true,
      eidas: true
    },
    sector_specific: {
      HEALTHCARE: {
        policy_paths: {
          phi_access: 'v1/data/innovabiz/eu/healthcare/phi_access',
          consent: 'v1/data/innovabiz/eu/healthcare/consent',
          data_residency: 'v1/data/innovabiz/eu/healthcare/data_residency'
        },
        strict_enforcement: true,
        require_consent_verification: true
      }
    }
  },
  BR: {
    opa_url: process.env.OPA_URL_BR || 'http://opa-br.innovabiz.internal:8181',
    policy_paths: {
      iam: 'v1/data/innovabiz/br/iam',
      healthcare: 'v1/data/innovabiz/br/healthcare',
      auth: 'v1/data/innovabiz/br/auth'
    },
    cache_ttl_ms: 60000, // 1 minuto
    max_retries: 3,
    timeout_ms: 500,
    headers: {
      'X-Region-Code': 'BR'
    },
    compliance: {
      lgpd: true,
      icp_brasil: true
    },
    sector_specific: {
      HEALTHCARE: {
        policy_paths: {
          phi_access: 'v1/data/innovabiz/br/healthcare/phi_access',
          consent: 'v1/data/innovabiz/br/healthcare/consent',
          data_residency: 'v1/data/innovabiz/br/healthcare/data_residency'
        },
        strict_enforcement: true,
        require_consent_verification: true
      }
    }
  },
  AO: {
    opa_url: process.env.OPA_URL_AO || 'http://opa-ao.innovabiz.internal:8181',
    policy_paths: {
      iam: 'v1/data/innovabiz/ao/iam',
      healthcare: 'v1/data/innovabiz/ao/healthcare',
      auth: 'v1/data/innovabiz/ao/auth'
    },
    cache_ttl_ms: 120000, // 2 minutos
    max_retries: 2,
    timeout_ms: 1000,
    headers: {
      'X-Region-Code': 'AO'
    },
    compliance: {
      pndsb: true
    },
    sector_specific: {
      HEALTHCARE: {
        policy_paths: {
          phi_access: 'v1/data/innovabiz/ao/healthcare/phi_access',
          offline_access: 'v1/data/innovabiz/ao/healthcare/offline_access'
        },
        strict_enforcement: false,
        require_consent_verification: false
      }
    }
  },
  US: {
    opa_url: process.env.OPA_URL_US || 'http://opa-us.innovabiz.internal:8181',
    policy_paths: {
      iam: 'v1/data/innovabiz/us/iam',
      healthcare: 'v1/data/innovabiz/us/healthcare',
      auth: 'v1/data/innovabiz/us/auth'
    },
    cache_ttl_ms: 30000, // 30 segundos
    max_retries: 3,
    timeout_ms: 500,
    headers: {
      'X-Region-Code': 'US'
    },
    compliance: {
      hipaa: true,
      nist: true,
      pci_dss: true
    },
    sector_specific: {
      HEALTHCARE: {
        policy_paths: {
          phi_access: 'v1/data/innovabiz/us/healthcare/phi_access',
          hipaa_privacy: 'v1/data/innovabiz/us/healthcare/hipaa_privacy',
          hipaa_security: 'v1/data/innovabiz/us/healthcare/hipaa_security'
        },
        strict_enforcement: true,
        require_consent_verification: true
      }
    }
  }
};

/**
 * Cliente para Open Policy Agent
 */
class OPAClient {
  /**
   * Inicializa o cliente OPA
   * @param {Object} options - Opções de configuração
   * @param {string} options.regionCode - Código da região (EU, BR, AO, US)
   * @param {string} options.serviceName - Nome do serviço
   * @param {boolean} options.enableCaching - Habilitar cache de decisões
   * @param {boolean} options.enableAudit - Habilitar auditoria detalhada
   * @param {Object} options.sectorConfig - Configurações específicas do setor
   */
  constructor(options = {}) {
    this.regionCode = options.regionCode || process.env.REGION_CODE || 'EU';
    this.serviceName = options.serviceName || process.env.SERVICE_NAME || 'innovabiz-service';
    this.enableCaching = options.enableCaching !== false;
    this.enableAudit = options.enableAudit !== false;
    this.sectorConfig = options.sectorConfig || null;
    
    // Carregar configuração regional
    this.config = REGIONAL_CONFIGS[this.regionCode] || REGIONAL_CONFIGS.EU;
    
    // Configurar API client
    this.client = axios.create({
      baseURL: this.config.opa_url,
      timeout: this.config.timeout_ms,
      headers: {
        ...this.config.headers,
        'X-Service-Name': this.serviceName
      }
    });
    
    // Inicializar cache se habilitado
    if (this.enableCaching) {
      this.cache = new Map();
    }
    
    // Se houver configuração específica do setor, aplicar
    if (this.sectorConfig && this.config.sector_specific && 
        this.config.sector_specific[this.sectorConfig.sector]) {
      this.sectorPolicyPaths = this.config.sector_specific[this.sectorConfig.sector].policy_paths;
      this.strictEnforcement = this.config.sector_specific[this.sectorConfig.sector].strict_enforcement;
    }
    
    logger.info(`OPAClient inicializado para a região ${this.regionCode}`, {
      service: this.serviceName,
      caching: this.enableCaching ? 'enabled' : 'disabled',
      audit: this.enableAudit ? 'enabled' : 'disabled',
      strictEnforcement: this.strictEnforcement
    });
  }
  
  /**
   * Verifica se o cliente está disponível
   * @returns {Promise<boolean>} Status do cliente
   */
  async checkHealth() {
    try {
      const response = await this.client.get('/health');
      return response.status === 200;
    } catch (error) {
      logger.error(`Erro ao verificar saúde do OPA: ${error.message}`, {
        error,
        region: this.regionCode,
        service: this.serviceName
      });
      
      return false;
    }
  }
  
  /**
   * Obtém a chave de cache para uma decisão
   * @private
   * @param {string} policyPath - Caminho da política
   * @param {Object} input - Dados de entrada
   * @returns {string} Chave de cache
   */
  getCacheKey(policyPath, input) {
    // Simplificar input para gerar chave de cache
    const simplifiedInput = {
      subject: input.subject,
      action: input.action,
      resource: input.resource,
      context: {
        tenant_id: input.context?.tenant_id,
        region_code: input.context?.region_code
      }
    };
    
    return `${policyPath}:${JSON.stringify(simplifiedInput)}`;
  }
  
  /**
   * Executa uma decisão de autorização
   * @param {string} policyPath - Caminho da política
   * @param {Object} input - Dados de entrada para a decisão
   * @param {Object} options - Opções adicionais
   * @returns {Promise<Object>} Resultado da decisão
   */
  async evaluate(policyPath, input, options = {}) {
    // Verificar cache se habilitado
    if (this.enableCaching && !options.bypassCache) {
      const cacheKey = this.getCacheKey(policyPath, input);
      
      if (this.cache.has(cacheKey)) {
        const cachedResult = this.cache.get(cacheKey);
        if (cachedResult.expiry > Date.now()) {
          if (this.enableAudit) {
            logger.debug('Decisão obtida do cache', {
              policyPath,
              result: cachedResult.result.allowed,
              region: this.regionCode
            });
          }
          return cachedResult.result;
        }
        this.cache.delete(cacheKey);
      }
    }
    
    // Mascarar dados sensíveis para logging
    const maskedInput = maskSensitiveData({ ...input }, ['password', 'token', 'credential']);
    
    try {
      // Registrar auditoria se habilitado
      if (this.enableAudit) {
        logger.info(`Avaliando política: ${policyPath}`, {
          policyPath,
          input: JSON.stringify(maskedInput),
          region: this.regionCode,
          service: this.serviceName
        });
      }
      
      // Formatação específica esperada pelo OPA
      const payload = { input };
      
      // Tentativas com retry
      let response;
      let attempts = 0;
      
      while (attempts < this.config.max_retries) {
        try {
          response = await this.client.post(policyPath, payload);
          break;
        } catch (error) {
          attempts++;
          
          // Se for o último retry, propagar erro
          if (attempts >= this.config.max_retries) {
            throw error;
          }
          
          // Esperar antes de tentar novamente
          await new Promise(resolve => setTimeout(resolve, 100 * attempts));
        }
      }
      
      // Processar resposta
      const result = response.data.result || {};
      
      // Adicionar metadados adicionais ao resultado
      const enhancedResult = {
        allowed: !!result.allow,
        reason: result.reason || 'No reason provided',
        decisions: result.decisions || [],
        obligations: result.obligations || [],
        policyName: policyPath
      };
      
      // Armazenar em cache se habilitado
      if (this.enableCaching) {
        const cacheKey = this.getCacheKey(policyPath, input);
        const ttl = options.cacheTtl || this.config.cache_ttl_ms;
        
        this.cache.set(cacheKey, {
          result: enhancedResult,
          expiry: Date.now() + ttl
        });
      }
      
      return enhancedResult;
    } catch (error) {
      const errorMessage = error.response?.data?.message || error.message;
      
      logger.error(`Erro ao avaliar política ${policyPath}: ${errorMessage}`, {
        error,
        policyPath,
        input: JSON.stringify(maskedInput),
        region: this.regionCode,
        service: this.serviceName
      });
      
      // Em caso de erro e strict enforcement, negar acesso
      if (this.strictEnforcement) {
        return {
          allowed: false,
          reason: `Policy evaluation error: ${errorMessage}`,
          decisions: [],
          obligations: [],
          policyName: policyPath,
          error: true
        };
      }
      
      // Em caso de erro sem strict enforcement, permitir acesso
      // mas registrar uma obrigação de auditoria
      return {
        allowed: !this.strictEnforcement,
        reason: `Policy evaluation error, access ${this.strictEnforcement ? 'denied' : 'allowed'} by default`,
        decisions: [],
        obligations: [{
          type: 'audit',
          details: `Failed policy evaluation for ${policyPath}: ${errorMessage}`
        }],
        policyName: policyPath,
        error: true
      };
    }
  }
  
  /**
   * Verifica autorização para operações de IAM
   * @param {Object} input - Dados para decisão
   * @param {Object} options - Opções adicionais
   * @returns {Promise<Object>} Resultado da decisão
   */
  async checkIAMAuthorization(input, options = {}) {
    const policyPath = this.config.policy_paths.iam;
    return this.evaluate(policyPath, input, options);
  }
  
  /**
   * Verifica autorização para operações de autenticação
   * @param {Object} input - Dados para decisão
   * @param {Object} options - Opções adicionais
   * @returns {Promise<Object>} Resultado da decisão
   */
  async checkAuthAuthorization(input, options = {}) {
    const policyPath = this.config.policy_paths.auth;
    return this.evaluate(policyPath, input, options);
  }
  
  /**
   * Verifica autorização para operações de saúde
   * @param {Object} input - Dados para decisão
   * @param {Object} options - Opções adicionais
   * @returns {Promise<Object>} Resultado da decisão
   */
  async checkHealthcareAuthorization(input, options = {}) {
    if (!this.sectorConfig || this.sectorConfig.sector !== 'HEALTHCARE') {
      throw new Error('Configuração de setor de saúde não disponível');
    }
    
    const policyPath = this.config.policy_paths.healthcare;
    return this.evaluate(policyPath, input, options);
  }
  
  /**
   * Verifica autorização para acesso a informações de saúde protegidas (PHI)
   * @param {Object} input - Dados para decisão
   * @param {Object} options - Opções adicionais
   * @returns {Promise<Object>} Resultado da decisão
   */
  async checkPHIAccess(input, options = {}) {
    if (!this.sectorConfig || this.sectorConfig.sector !== 'HEALTHCARE' || !this.sectorPolicyPaths) {
      throw new Error('Configuração de setor de saúde não disponível');
    }
    
    const policyPath = this.sectorPolicyPaths.phi_access;
    return this.evaluate(policyPath, input, options);
  }
  
  /**
   * Verifica se há consentimento para operação específica
   * @param {Object} input - Dados para decisão
   * @param {Object} options - Opções adicionais
   * @returns {Promise<Object>} Resultado da decisão
   */
  async checkConsent(input, options = {}) {
    if (!this.sectorConfig || this.sectorConfig.sector !== 'HEALTHCARE' || !this.sectorPolicyPaths || !this.sectorPolicyPaths.consent) {
      throw new Error('Verificação de consentimento não disponível para esta região/setor');
    }
    
    const policyPath = this.sectorPolicyPaths.consent;
    return this.evaluate(policyPath, input, options);
  }
  
  /**
   * Verifica conformidade HIPAA para operação específica
   * @param {Object} input - Dados para decisão
   * @param {Object} options - Opções adicionais
   * @returns {Promise<Object>} Resultado da decisão
   */
  async checkHIPAACompliance(input, options = {}) {
    if (
      this.regionCode !== 'US' || 
      !this.sectorConfig || 
      this.sectorConfig.sector !== 'HEALTHCARE' || 
      !this.sectorPolicyPaths || 
      !this.sectorPolicyPaths.hipaa_privacy
    ) {
      throw new Error('Verificação de conformidade HIPAA não disponível para esta região/setor');
    }
    
    // Verificar regras de privacidade HIPAA
    const privacyResult = await this.evaluate(
      this.sectorPolicyPaths.hipaa_privacy,
      input,
      options
    );
    
    // Verificar regras de segurança HIPAA, se disponível
    if (this.sectorPolicyPaths.hipaa_security) {
      const securityResult = await this.evaluate(
        this.sectorPolicyPaths.hipaa_security,
        input,
        options
      );
      
      // Combinar resultados
      return {
        allowed: privacyResult.allowed && securityResult.allowed,
        reason: privacyResult.allowed ? securityResult.reason : privacyResult.reason,
        decisions: [...privacyResult.decisions, ...securityResult.decisions],
        obligations: [...privacyResult.obligations, ...securityResult.obligations],
        privacyCompliant: privacyResult.allowed,
        securityCompliant: securityResult.allowed,
        error: privacyResult.error || securityResult.error
      };
    }
    
    return privacyResult;
  }
  
  /**
   * Verifica residência de dados para operação específica
   * @param {Object} input - Dados para decisão
   * @param {Object} options - Opções adicionais
   * @returns {Promise<Object>} Resultado da decisão
   */
  async checkDataResidency(input, options = {}) {
    if (
      !this.sectorConfig || 
      this.sectorConfig.sector !== 'HEALTHCARE' || 
      !this.sectorPolicyPaths || 
      !this.sectorPolicyPaths.data_residency
    ) {
      throw new Error('Verificação de residência de dados não disponível para esta região/setor');
    }
    
    const policyPath = this.sectorPolicyPaths.data_residency;
    return this.evaluate(policyPath, input, options);
  }
  
  /**
   * Verifica múltiplas políticas em uma única chamada
   * @param {Object[]} policyChecks - Array de objetos { policyPath, input }
   * @param {Object} options - Opções adicionais
   * @returns {Promise<Object[]>} Array com resultados de decisões
   */
  async evaluateMultiple(policyChecks, options = {}) {
    const results = [];
    
    for (const check of policyChecks) {
      const result = await this.evaluate(check.policyPath, check.input, options);
      results.push({
        policyPath: check.policyPath,
        result
      });
    }
    
    return results;
  }
  
  /**
   * Registra evento de auditoria em OPA
   * @param {Object} auditEvent - Evento a ser registrado
   * @returns {Promise<void>}
   */
  async logAuditEvent(auditEvent) {
    try {
      const maskedEvent = maskSensitiveData({ ...auditEvent }, ['password', 'token', 'credential']);
      
      await this.client.post('/v1/data/innovabiz/audit', {
        input: {
          event: {
            ...maskedEvent,
            timestamp: Date.now(),
            service: this.serviceName,
            region: this.regionCode
          }
        }
      });
      
      logger.debug('Evento de auditoria registrado no OPA', {
        eventType: maskedEvent.type,
        region: this.regionCode,
        service: this.serviceName
      });
    } catch (error) {
      logger.error(`Erro ao registrar evento de auditoria: ${error.message}`, {
        error,
        region: this.regionCode,
        service: this.serviceName
      });
    }
  }
  
  /**
   * Limpa o cache de decisões
   */
  clearCache() {
    if (this.enableCaching && this.cache) {
      this.cache.clear();
      logger.debug('Cache de decisões OPA limpo', {
        region: this.regionCode,
        service: this.serviceName
      });
    }
  }
}

module.exports = OPAClient;
