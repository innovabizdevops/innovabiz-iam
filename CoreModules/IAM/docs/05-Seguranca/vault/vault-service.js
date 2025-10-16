/**
 * @fileoverview Serviço HashiCorp Vault para a plataforma INNOVABIZ
 * @author INNOVABIZ Dev Team
 * @version 1.0.0
 * @module iam/seguranca/vault/vault-service
 * @description Serviço para gerenciamento de segredos e operações criptográficas usando HashiCorp Vault
 */

const VaultConfig = require('./vault-config');
const logger = require('../../core/utils/logging');
const { maskSensitiveData } = require('../../core/utils/data-masking');
const retry = require('async-retry');

/**
 * Classe de serviço para interação com HashiCorp Vault
 */
class VaultService {
  /**
   * Inicializa o serviço Vault
   * @param {Object} options - Opções de configuração
   * @param {string} options.regionCode - Código da região (EU, BR, AO, US)
   * @param {string} options.serviceName - Nome do serviço
   * @param {string} options.environment - Ambiente (production, development, etc.)
   * @param {Object} options.sectorConfig - Configurações específicas do setor
   * @param {boolean} options.enableCaching - Habilitar cache de segredos
   * @param {boolean} options.enableAudit - Habilitar auditoria detalhada
   */
  constructor(options = {}) {
    this.vaultConfig = new VaultConfig(options);
    this.client = this.vaultConfig.getClient();
    this.authenticated = false;
    this.tokenRenewalTimeout = null;
    this.secretCache = new Map();
    this.enableCaching = options.enableCaching !== false;
    this.enableAudit = options.enableAudit !== false;
    this.regionCode = options.regionCode || process.env.REGION_CODE || 'EU';
    this.serviceName = options.serviceName || process.env.SERVICE_NAME || 'innovabiz-service';
    this.sectorConfig = options.sectorConfig || null;
    
    // Inicializar estado
    this.initialized = false;
    
    logger.info(`VaultService inicializado para a região ${this.regionCode}`, {
      service: this.serviceName,
      caching: this.enableCaching ? 'enabled' : 'disabled',
      audit: this.enableAudit ? 'enabled' : 'disabled'
    });
  }
  
  /**
   * Inicializa o serviço Vault
   * @returns {Promise<boolean>} Se a inicialização foi bem-sucedida
   */
  async initialize() {
    try {
      // Verificar status do Vault
      const { initialized, sealed } = await this.client.status();
      
      if (!initialized) {
        logger.error('Vault não inicializado nesta região', {
          region: this.regionCode,
          service: this.serviceName
        });
        return false;
      }
      
      if (sealed) {
        logger.error('Vault selado nesta região', {
          region: this.regionCode,
          service: this.serviceName
        });
        return false;
      }
      
      // Autenticar
      await this.authenticate();
      
      // Verifica se as engines necessárias estão habilitadas
      await this.checkRequiredEngines();
      
      this.initialized = true;
      logger.info(`Vault inicializado com sucesso na região ${this.regionCode}`);
      
      return true;
    } catch (error) {
      logger.error(`Erro ao inicializar Vault: ${error.message}`, {
        error,
        region: this.regionCode,
        service: this.serviceName
      });
      
      return false;
    }
  }
  
  /**
   * Autentica no Vault usando o método apropriado
   * @returns {Promise<void>}
   */
  async authenticate() {
    // Se já autenticado, renovar token
    if (this.authenticated) {
      await this.renewToken();
      return;
    }
    
    try {
      // Tentar autenticação via AppRole (para serviços)
      if (process.env.VAULT_ROLE_ID && process.env.VAULT_SECRET_ID) {
        await this.authenticateWithAppRole();
      }
      // Tentar autenticação via Kubernetes
      else if (process.env.VAULT_K8S_AUTH) {
        await this.authenticateWithKubernetes();
      }
      // Tentar autenticação via JWT/OIDC
      else if (process.env.VAULT_JWT_PATH) {
        await this.authenticateWithJwt();
      }
      // Tentar autenticação via token direto (menos seguro)
      else if (process.env.VAULT_TOKEN) {
        await this.authenticateWithToken();
      }
      else {
        throw new Error('Nenhum método de autenticação Vault configurado');
      }
      
      this.authenticated = true;
      logger.info('Autenticação Vault bem-sucedida', {
        region: this.regionCode,
        service: this.serviceName
      });
      
      // Configurar renovação automática de token
      this.setupTokenRenewal();
    } catch (error) {
      this.authenticated = false;
      logger.error(`Falha na autenticação Vault: ${error.message}`, {
        error,
        region: this.regionCode,
        service: this.serviceName
      });
      
      throw error;
    }
  }
  
  /**
   * Autentica usando AppRole
   * @private
   */
  async authenticateWithAppRole() {
    const result = await this.client.approleLogin({
      role_id: process.env.VAULT_ROLE_ID,
      secret_id: process.env.VAULT_SECRET_ID
    });
    
    this.client.token = result.auth.client_token;
    this.tokenExpiry = Date.now() + (result.auth.lease_duration * 1000);
  }
  
  /**
   * Autentica usando JWT/OIDC
   * @private
   */
  async authenticateWithJwt() {
    const jwtPath = process.env.VAULT_JWT_PATH || 'jwt';
    const jwt = process.env.VAULT_JWT_TOKEN;
    const role = process.env.VAULT_JWT_ROLE || this.serviceName;
    
    const result = await this.client.write(`auth/${jwtPath}/login`, {
      jwt,
      role
    });
    
    this.client.token = result.auth.client_token;
    this.tokenExpiry = Date.now() + (result.auth.lease_duration * 1000);
  }
  
  /**
   * Autentica usando Kubernetes
   * @private
   */
  async authenticateWithKubernetes() {
    const fs = require('fs');
    const k8sTokenPath = process.env.VAULT_K8S_TOKEN_PATH || '/var/run/secrets/kubernetes.io/serviceaccount/token';
    const k8sToken = fs.readFileSync(k8sTokenPath, 'utf8');
    const k8sRole = process.env.VAULT_K8S_ROLE || this.serviceName;
    
    const result = await this.client.kubernetesLogin({
      role: k8sRole,
      jwt: k8sToken
    });
    
    this.client.token = result.auth.client_token;
    this.tokenExpiry = Date.now() + (result.auth.lease_duration * 1000);
  }
  
  /**
   * Autentica usando token direto
   * @private
   */
  async authenticateWithToken() {
    this.client.token = process.env.VAULT_TOKEN;
    
    // Verificar token
    const result = await this.client.tokenLookupSelf();
    this.tokenExpiry = Date.now() + (result.data.ttl * 1000);
  }
  
  /**
   * Configura renovação automática de token
   * @private
   */
  setupTokenRenewal() {
    const now = Date.now();
    const tokenTtl = this.tokenExpiry - now;
    
    // Renovar quando restarem 20% do tempo de vida
    const renewalTime = tokenTtl * 0.8;
    
    // Limpar timeout anterior, se existir
    if (this.tokenRenewalTimeout) {
      clearTimeout(this.tokenRenewalTimeout);
    }
    
    this.tokenRenewalTimeout = setTimeout(async () => {
      try {
        await this.renewToken();
      } catch (error) {
        logger.error(`Falha ao renovar token: ${error.message}`, {
          error,
          region: this.regionCode,
          service: this.serviceName
        });
        
        // Tentar reautenticar em caso de falha
        try {
          await this.authenticate();
        } catch (authError) {
          logger.error(`Falha na reautenticação: ${authError.message}`, {
            error: authError,
            region: this.regionCode,
            service: this.serviceName
          });
        }
      }
    }, renewalTime);
  }
  
  /**
   * Renova o token atual
   * @private
   */
  async renewToken() {
    const result = await this.client.tokenRenewSelf();
    this.tokenExpiry = Date.now() + (result.auth.lease_duration * 1000);
    
    logger.debug('Token Vault renovado', {
      region: this.regionCode,
      service: this.serviceName,
      expiry: new Date(this.tokenExpiry).toISOString()
    });
    
    // Configurar próxima renovação
    this.setupTokenRenewal();
  }
  
  /**
   * Verifica se as engines necessárias estão habilitadas
   * @private
   */
  async checkRequiredEngines() {
    try {
      const mounts = await this.client.mounts();
      
      // Verificar KV
      const kvPath = `${this.vaultConfig.config.kv_mount_path}/`;
      if (!mounts[kvPath]) {
        logger.warn(`KV Engine não encontrada em ${kvPath}`, {
          region: this.regionCode,
          service: this.serviceName
        });
      }
      
      // Verificar Transit (se configurado)
      if (this.vaultConfig.isCapabilityEnabled('transit', true)) {
        const transitPath = `${this.vaultConfig.config.transit_mount_path}/`;
        if (!mounts[transitPath]) {
          logger.warn(`Transit Engine não encontrada em ${transitPath}`, {
            region: this.regionCode,
            service: this.serviceName
          });
        }
      }
      
      // Verificar PKI (se configurado)
      if (this.vaultConfig.isCapabilityEnabled('pki', true)) {
        const pkiPath = `${this.vaultConfig.config.pki_mount_path}/`;
        if (!mounts[pkiPath]) {
          logger.warn(`PKI Engine não encontrada em ${pkiPath}`, {
            region: this.regionCode,
            service: this.serviceName
          });
        }
      }
    } catch (error) {
      logger.error(`Erro ao verificar engines Vault: ${error.message}`, {
        error,
        region: this.regionCode
      });
    }
  }
  
  /**
   * Obtém um segredo do Vault
   * @param {string} secretType - Tipo do segredo (db, api, auth, etc.)
   * @param {string} secretName - Nome do segredo
   * @param {Object} options - Opções adicionais
   * @returns {Promise<Object>} Dados do segredo
   */
  async getSecret(secretType, secretName, options = {}) {
    // Verificar autenticação
    if (!this.authenticated) {
      await this.authenticate();
    }
    
    const cacheKey = `${secretType}:${secretName}`;
    
    // Verificar cache
    if (this.enableCaching && this.secretCache.has(cacheKey) && !options.bypassCache) {
      const cachedData = this.secretCache.get(cacheKey);
      if (cachedData.expiry > Date.now()) {
        return cachedData.data;
      }
      // Remover do cache se expirado
      this.secretCache.delete(cacheKey);
    }
    
    try {
      const secretPath = this.vaultConfig.buildSecretPath(secretType, secretName);
      
      // Registrar auditoria se habilitado
      if (this.enableAudit) {
        logger.info(`Obtendo segredo ${secretType}:${secretName}`, {
          region: this.regionCode,
          service: this.serviceName,
          secretType,
          secretName,
          operation: 'getSecret'
        });
      }
      
      // Obter segredo com retry para lidar com falhas transitórias
      const response = await retry(async (bail) => {
        try {
          return await this.client.read(secretPath);
        } catch (error) {
          // Não tentar novamente para erros de permissão ou de segredo não encontrado
          if (error.response && (error.response.statusCode === 403 || error.response.statusCode === 404)) {
            bail(error);
            return;
          }
          throw error;
        }
      }, {
        retries: 3,
        factor: 2,
        minTimeout: 1000,
        maxTimeout: 5000
      });
      
      if (!response || !response.data || !response.data.data) {
        throw new Error(`Segredo não encontrado: ${secretType}:${secretName}`);
      }
      
      const secretData = response.data.data;
      
      // Armazenar em cache se habilitado
      if (this.enableCaching) {
        const ttl = options.cacheTtl || 300000; // 5 minutos padrão
        this.secretCache.set(cacheKey, {
          data: secretData,
          expiry: Date.now() + ttl
        });
      }
      
      return secretData;
    } catch (error) {
      logger.error(`Erro ao obter segredo ${secretType}:${secretName}: ${error.message}`, {
        error,
        region: this.regionCode,
        service: this.serviceName,
        secretType,
        secretName
      });
      
      throw error;
    }
  }
  
  /**
   * Grava um segredo no Vault
   * @param {string} secretType - Tipo do segredo (db, api, auth, etc.)
   * @param {string} secretName - Nome do segredo
   * @param {Object} secretData - Dados do segredo
   * @param {Object} options - Opções adicionais
   * @returns {Promise<Object>} Resultado da operação
   */
  async writeSecret(secretType, secretName, secretData, options = {}) {
    // Verificar autenticação
    if (!this.authenticated) {
      await this.authenticate();
    }
    
    try {
      const secretPath = this.vaultConfig.buildSecretPath(secretType, secretName);
      
      // Mascarar dados sensíveis para logging
      const maskedData = maskSensitiveData({ ...secretData }, ['password', 'secret', 'key', 'token', 'credential']);
      
      // Registrar auditoria se habilitado
      if (this.enableAudit) {
        logger.info(`Gravando segredo ${secretType}:${secretName}`, {
          region: this.regionCode,
          service: this.serviceName,
          secretType,
          secretName,
          operation: 'writeSecret',
          data: JSON.stringify(maskedData)
        });
      }
      
      // No KV v2, os dados precisam estar dentro de um objeto data
      const payload = {
        data: secretData,
        options: {
          cas: options.cas
        }
      };
      
      const result = await this.client.write(secretPath, payload);
      
      // Invalidar cache para este segredo
      const cacheKey = `${secretType}:${secretName}`;
      this.secretCache.delete(cacheKey);
      
      return result;
    } catch (error) {
      logger.error(`Erro ao gravar segredo ${secretType}:${secretName}: ${error.message}`, {
        error,
        region: this.regionCode,
        service: this.serviceName,
        secretType,
        secretName
      });
      
      throw error;
    }
  }
  
  /**
   * Remove um segredo do Vault
   * @param {string} secretType - Tipo do segredo (db, api, auth, etc.)
   * @param {string} secretName - Nome do segredo
   * @returns {Promise<Object>} Resultado da operação
   */
  async deleteSecret(secretType, secretName) {
    // Verificar autenticação
    if (!this.authenticated) {
      await this.authenticate();
    }
    
    try {
      const secretPath = this.vaultConfig.buildSecretPath(secretType, secretName);
      
      // Registrar auditoria se habilitado
      if (this.enableAudit) {
        logger.info(`Removendo segredo ${secretType}:${secretName}`, {
          region: this.regionCode,
          service: this.serviceName,
          secretType,
          secretName,
          operation: 'deleteSecret'
        });
      }
      
      const result = await this.client.delete(secretPath);
      
      // Invalidar cache para este segredo
      const cacheKey = `${secretType}:${secretName}`;
      this.secretCache.delete(cacheKey);
      
      return result;
    } catch (error) {
      logger.error(`Erro ao remover segredo ${secretType}:${secretName}: ${error.message}`, {
        error,
        region: this.regionCode,
        service: this.serviceName,
        secretType,
        secretName
      });
      
      throw error;
    }
  }
  
  /**
   * Criptografa dados usando a engine Transit
   * @param {string} keyName - Nome da chave
   * @param {string} plaintext - Texto a ser criptografado (base64)
   * @param {Object} options - Opções adicionais
   * @returns {Promise<string>} Texto criptografado
   */
  async encrypt(keyName, plaintext, options = {}) {
    // Verificar autenticação
    if (!this.authenticated) {
      await this.authenticate();
    }
    
    if (!this.vaultConfig.isCapabilityEnabled('transit', true)) {
      throw new Error('Engine Transit não habilitada para esta região');
    }
    
    try {
      // Converter para base64 se não estiver
      const b64Text = Buffer.isBuffer(plaintext) ? 
        plaintext.toString('base64') : 
        (Buffer.from(plaintext).toString('base64'));
      
      const path = `${this.vaultConfig.config.transit_mount_path}/encrypt/${keyName}`;
      
      // Registrar auditoria se habilitado
      if (this.enableAudit) {
        logger.info(`Criptografando dados com chave ${keyName}`, {
          region: this.regionCode,
          service: this.serviceName,
          keyName,
          operation: 'encrypt'
        });
      }
      
      const result = await this.client.write(path, {
        plaintext: b64Text,
        context: options.context,
        key_version: options.keyVersion
      });
      
      return result.data.ciphertext;
    } catch (error) {
      logger.error(`Erro ao criptografar com chave ${keyName}: ${error.message}`, {
        error,
        region: this.regionCode,
        service: this.serviceName,
        keyName
      });
      
      throw error;
    }
  }
  
  /**
   * Descriptografa dados usando a engine Transit
   * @param {string} keyName - Nome da chave
   * @param {string} ciphertext - Texto criptografado
   * @param {Object} options - Opções adicionais
   * @returns {Promise<string>} Texto descriptografado (base64)
   */
  async decrypt(keyName, ciphertext, options = {}) {
    // Verificar autenticação
    if (!this.authenticated) {
      await this.authenticate();
    }
    
    if (!this.vaultConfig.isCapabilityEnabled('transit', true)) {
      throw new Error('Engine Transit não habilitada para esta região');
    }
    
    try {
      const path = `${this.vaultConfig.config.transit_mount_path}/decrypt/${keyName}`;
      
      // Registrar auditoria se habilitado
      if (this.enableAudit) {
        logger.info(`Descriptografando dados com chave ${keyName}`, {
          region: this.regionCode,
          service: this.serviceName,
          keyName,
          operation: 'decrypt'
        });
      }
      
      const result = await this.client.write(path, {
        ciphertext,
        context: options.context
      });
      
      return result.data.plaintext;
    } catch (error) {
      logger.error(`Erro ao descriptografar com chave ${keyName}: ${error.message}`, {
        error,
        region: this.regionCode,
        service: this.serviceName,
        keyName
      });
      
      throw error;
    }
  }
  
  /**
   * Gera uma chave dinâmica para acesso a banco de dados
   * @param {string} dbName - Nome do banco de dados
   * @param {string} role - Papel para acesso
   * @param {Object} options - Opções adicionais
   * @returns {Promise<Object>} Credenciais geradas
   */
  async getDatabaseCredentials(dbName, role, options = {}) {
    // Verificar autenticação
    if (!this.authenticated) {
      await this.authenticate();
    }
    
    if (!this.vaultConfig.isCapabilityEnabled('database', true)) {
      throw new Error('Engine Database não habilitada para esta região');
    }
    
    try {
      const path = `database/creds/${role}`;
      
      // Registrar auditoria se habilitado
      if (this.enableAudit) {
        logger.info(`Obtendo credenciais para banco ${dbName} com papel ${role}`, {
          region: this.regionCode,
          service: this.serviceName,
          dbName,
          role,
          operation: 'getDatabaseCredentials'
        });
      }
      
      const result = await this.client.read(path);
      
      return {
        username: result.data.username,
        password: result.data.password,
        leaseId: result.lease_id,
        leaseDuration: result.lease_duration,
        renewable: result.renewable
      };
    } catch (error) {
      logger.error(`Erro ao obter credenciais para banco ${dbName} papel ${role}: ${error.message}`, {
        error,
        region: this.regionCode,
        service: this.serviceName,
        dbName,
        role
      });
      
      throw error;
    }
  }
  
  /**
   * Faz shutdown do serviço Vault
   */
  shutdown() {
    // Limpar timeout de renovação de token
    if (this.tokenRenewalTimeout) {
      clearTimeout(this.tokenRenewalTimeout);
      this.tokenRenewalTimeout = null;
    }
    
    // Limpar cache
    this.secretCache.clear();
    
    logger.info('VaultService desligado');
  }
}

module.exports = VaultService;
