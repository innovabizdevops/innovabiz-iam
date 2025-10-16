/**
 * @fileoverview Configuração do HashiCorp Vault para a plataforma INNOVABIZ
 * @author INNOVABIZ Dev Team
 * @version 1.0.0
 * @module iam/seguranca/vault/vault-config
 * @description Configuração centralizada do HashiCorp Vault para gerenciamento de segredos
 */

const vault = require('node-vault');
const logger = require('../../core/utils/logging');
const fs = require('fs');
const path = require('path');

/**
 * Configurações regionais para Vault
 */
const REGIONAL_CONFIGS = {
  EU: {
    vault_addr: process.env.VAULT_ADDR_EU || 'https://vault-eu.innovabiz.internal:8200',
    vault_namespace: process.env.VAULT_NAMESPACE_EU || 'innovabiz/eu',
    jwt_auth_path: 'jwt-eu',
    approle_path: 'approle-eu',
    kv_mount_path: 'kv-eu',
    transit_mount_path: 'transit-eu',
    pki_mount_path: 'pki-eu',
    auth_mount_path: 'auth-eu',
    data_residency: 'EU',
    token_ttl: '24h',
    secret_engines: {
      kv: { version: 2 },
      transit: true,
      pki: true,
      database: true
    },
    compliance: {
      gdpr: true,
      eidas: true
    },
    sector_specific: {
      HEALTHCARE: {
        secrets_path: 'healthcare-eu',
        strict_access_control: true,
        data_masking: true
      }
    }
  },
  BR: {
    vault_addr: process.env.VAULT_ADDR_BR || 'https://vault-br.innovabiz.internal:8200',
    vault_namespace: process.env.VAULT_NAMESPACE_BR || 'innovabiz/br',
    jwt_auth_path: 'jwt-br',
    approle_path: 'approle-br',
    kv_mount_path: 'kv-br',
    transit_mount_path: 'transit-br',
    pki_mount_path: 'pki-br',
    auth_mount_path: 'auth-br',
    data_residency: 'BR',
    token_ttl: '24h',
    secret_engines: {
      kv: { version: 2 },
      transit: true,
      pki: true,
      database: true
    },
    compliance: {
      lgpd: true,
      icp_brasil: true
    },
    sector_specific: {
      HEALTHCARE: {
        secrets_path: 'healthcare-br',
        strict_access_control: true,
        data_masking: true
      }
    }
  },
  AO: {
    vault_addr: process.env.VAULT_ADDR_AO || 'https://vault-ao.innovabiz.internal:8200',
    vault_namespace: process.env.VAULT_NAMESPACE_AO || 'innovabiz/ao',
    jwt_auth_path: 'jwt-ao',
    approle_path: 'approle-ao',
    kv_mount_path: 'kv-ao',
    transit_mount_path: 'transit-ao',
    pki_mount_path: 'pki-ao',
    auth_mount_path: 'auth-ao',
    data_residency: 'AO',
    token_ttl: '24h',
    secret_engines: {
      kv: { version: 2 },
      transit: true,
      pki: true,
      database: true
    },
    compliance: {
      pndsb: true
    },
    sector_specific: {
      HEALTHCARE: {
        secrets_path: 'healthcare-ao',
        strict_access_control: false,
        data_masking: false
      }
    }
  },
  US: {
    vault_addr: process.env.VAULT_ADDR_US || 'https://vault-us.innovabiz.internal:8200',
    vault_namespace: process.env.VAULT_NAMESPACE_US || 'innovabiz/us',
    jwt_auth_path: 'jwt-us',
    approle_path: 'approle-us',
    kv_mount_path: 'kv-us',
    transit_mount_path: 'transit-us',
    pki_mount_path: 'pki-us',
    auth_mount_path: 'auth-us',
    data_residency: 'US',
    token_ttl: '24h',
    secret_engines: {
      kv: { version: 2 },
      transit: true,
      pki: true,
      database: true
    },
    compliance: {
      hipaa: true,
      nist: true,
      pci_dss: true
    },
    sector_specific: {
      HEALTHCARE: {
        secrets_path: 'healthcare-us',
        strict_access_control: true,
        data_masking: true
      }
    }
  }
};

/**
 * Classe de configuração do HashiCorp Vault
 */
class VaultConfig {
  /**
   * Inicializa a configuração do Vault
   * @param {Object} options - Opções de configuração
   * @param {string} options.regionCode - Código da região (EU, BR, AO, US)
   * @param {string} options.serviceName - Nome do serviço
   * @param {string} options.environment - Ambiente (production, development, etc.)
   * @param {Object} options.sectorConfig - Configurações específicas do setor
   * @param {Object} options.customConfig - Configurações personalizadas
   */
  constructor(options = {}) {
    this.regionCode = options.regionCode || process.env.REGION_CODE || 'EU';
    this.serviceName = options.serviceName || process.env.SERVICE_NAME || 'innovabiz-service';
    this.environment = options.environment || process.env.NODE_ENV || 'development';
    this.sectorConfig = options.sectorConfig || null;
    this.customConfig = options.customConfig || {};
    
    // Carregar configurações regionais
    this.config = REGIONAL_CONFIGS[this.regionCode] || REGIONAL_CONFIGS.EU;
    
    // Mesclar com configurações personalizadas
    this.config = {
      ...this.config,
      ...this.customConfig
    };
    
    // Se houver configuração específica do setor, aplicar
    if (this.sectorConfig && this.config.sector_specific && 
        this.config.sector_specific[this.sectorConfig.sector]) {
      this.sectorSpecificConfig = this.config.sector_specific[this.sectorConfig.sector];
    }
    
    logger.info(`VaultConfig inicializado para a região ${this.regionCode}`, {
      service: this.serviceName,
      environment: this.environment
    });
  }
  
  /**
   * Obtém o cliente Vault configurado
   * @returns {Object} Cliente Vault
   */
  getClient() {
    const clientOptions = {
      apiVersion: 'v1',
      endpoint: this.config.vault_addr,
      namespace: this.config.vault_namespace,
      requestOptions: {
        timeout: 5000
      }
    };
    
    // Adicionar TLS se certificados estiverem disponíveis
    if (process.env.VAULT_CA_CERT) {
      clientOptions.ca = fs.readFileSync(process.env.VAULT_CA_CERT);
    }
    
    if (process.env.VAULT_CLIENT_CERT && process.env.VAULT_CLIENT_KEY) {
      clientOptions.cert = fs.readFileSync(process.env.VAULT_CLIENT_CERT);
      clientOptions.key = fs.readFileSync(process.env.VAULT_CLIENT_KEY);
    }
    
    // Criar cliente Vault
    const client = vault(clientOptions);
    
    // Adicionar headers específicos para logging de auditoria
    client.requestOptions = {
      headers: {
        'X-Vault-Namespace': this.config.vault_namespace,
        'X-Service-Name': this.serviceName,
        'X-Region-Code': this.regionCode
      }
    };
    
    return client;
  }
  
  /**
   * Constrói o caminho do segredo baseado no tipo e ambiente
   * @param {string} secretType - Tipo do segredo (db, api, auth, etc.)
   * @param {string} secretName - Nome do segredo
   * @returns {string} Caminho completo do segredo
   */
  buildSecretPath(secretType, secretName) {
    // Se for específico do setor, usar caminho específico
    if (this.sectorSpecificConfig) {
      return `${this.config.kv_mount_path}/data/${this.sectorSpecificConfig.secrets_path}/${secretType}/${this.environment}/${secretName}`;
    }
    
    return `${this.config.kv_mount_path}/data/${this.serviceName}/${secretType}/${this.environment}/${secretName}`;
  }
  
  /**
   * Constrói o caminho do trânsito para operações criptográficas
   * @param {string} keyName - Nome da chave de trânsito
   * @returns {string} Caminho completo do trânsito
   */
  buildTransitPath(keyName) {
    return `${this.config.transit_mount_path}/${keyName}`;
  }
  
  /**
   * Obtém a configuração específica de conformidade
   * @returns {Object} Configuração de conformidade
   */
  getComplianceConfig() {
    return this.config.compliance || {};
  }
  
  /**
   * Obtém informações específicas para o setor configurado
   * @returns {Object|null} Configuração específica do setor ou null
   */
  getSectorSpecificConfig() {
    return this.sectorSpecificConfig || null;
  }
  
  /**
   * Verifica se uma capacidade específica está habilitada
   * @param {string} engine - Engine do Vault
   * @param {string} capability - Capacidade específica
   * @returns {boolean} Se a capacidade está habilitada
   */
  isCapabilityEnabled(engine, capability) {
    if (!this.config.secret_engines[engine]) return false;
    
    if (typeof this.config.secret_engines[engine] === 'boolean') {
      return this.config.secret_engines[engine];
    }
    
    return !!this.config.secret_engines[engine][capability];
  }
  
  /**
   * Obtém a configuração completa
   * @returns {Object} Configuração atual
   */
  getConfig() {
    // Remover informações sensíveis
    const safeConfig = { ...this.config };
    delete safeConfig.vault_token;
    
    return {
      regionCode: this.regionCode,
      serviceName: this.serviceName,
      environment: this.environment,
      vaultAddr: safeConfig.vault_addr,
      vaultNamespace: safeConfig.vault_namespace,
      dataResidency: safeConfig.data_residency,
      secretEngines: safeConfig.secret_engines,
      compliance: safeConfig.compliance
    };
  }
}

module.exports = VaultConfig;
