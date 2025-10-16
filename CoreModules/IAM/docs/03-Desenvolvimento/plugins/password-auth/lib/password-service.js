/**
 * Serviço de Gerenciamento de Senhas
 * 
 * Este módulo fornece funcionalidades para hashing e verificação de senhas
 * com suporte a múltiplos algoritmos de hashing conforme as melhores práticas
 * de segurança.
 * 
 * @module password-service
 * @author INNOVABIZ
 * @copyright 2025 INNOVABIZ
 * @license Proprietary
 */

const crypto = require('crypto');
const bcrypt = require('bcryptjs');
const argon2 = require('argon2');
const { logger } = require('@innovabiz/auth-framework/utils/logging');

/**
 * Serviço de gerenciamento de senhas
 */
class PasswordService {
  /**
   * Cria uma nova instância do serviço de senha
   * 
   * @param {string} defaultAlgorithm Algoritmo de hash padrão
   * @param {Object} defaultOptions Opções padrão para o algoritmo de hash
   */
  constructor(defaultAlgorithm = 'argon2id', defaultOptions = {}) {
    this.defaultAlgorithm = defaultAlgorithm;
    this.defaultOptions = {
      // Configurações padrão para bcrypt
      bcrypt: {
        rounds: defaultOptions.iterations || 12
      },
      // Configurações padrão para argon2id
      argon2id: {
        type: argon2.argon2id,
        memoryCost: defaultOptions.memoryCost || 65536, // 64MB
        timeCost: defaultOptions.timeCost || 3,
        parallelism: defaultOptions.parallelism || 1
      },
      // Configurações padrão para PBKDF2 (SHA-256)
      'pbkdf2-sha256': {
        iterations: defaultOptions.iterations || 600000,
        keylen: 32,
        digest: 'sha256'
      },
      // Configurações padrão para PBKDF2 (SHA-512)
      'pbkdf2-sha512': {
        iterations: defaultOptions.iterations || 210000,
        keylen: 64,
        digest: 'sha512'
      },
      // Configurações padrão para scrypt
      scrypt: {
        cost: defaultOptions.memoryCost || 16384,
        blockSize: 8,
        parallelization: defaultOptions.parallelism || 1,
        keylen: 64
      }
    };
    
    logger.info(`[PasswordService] Inicializado com algoritmo padrão: ${defaultAlgorithm}`);
  }
  
  /**
   * Gera um salt aleatório
   * 
   * @param {number} length Comprimento do salt em bytes
   * @returns {string} Salt em formato Base64
   */
  generateSalt(length = 16) {
    const salt = crypto.randomBytes(length);
    return salt.toString('base64');
  }
  
  /**
   * Gera um hash de senha utilizando o algoritmo especificado
   * 
   * @param {string} password Senha em texto plano
   * @param {string} algorithm Algoritmo de hash a utilizar
   * @param {Object} options Opções específicas do algoritmo
   * @returns {Promise<Object>} Hash e metadados
   */
  async hashPassword(password, algorithm = this.defaultAlgorithm, options = {}) {
    // Normaliza algoritmo para caixa minúscula
    algorithm = algorithm.toLowerCase();
    
    try {
      switch (algorithm) {
        case 'bcrypt':
          return await this.hashBcrypt(password, options);
        
        case 'argon2id':
          return await this.hashArgon2id(password, options);
        
        case 'pbkdf2-sha256':
          return await this.hashPbkdf2(password, 'sha256', options);
        
        case 'pbkdf2-sha512':
          return await this.hashPbkdf2(password, 'sha512', options);
        
        case 'scrypt':
          return await this.hashScrypt(password, options);
        
        default:
          throw new Error(`Algoritmo de hash não suportado: ${algorithm}`);
      }
    } catch (error) {
      logger.error(`[PasswordService] Erro ao gerar hash de senha: ${error.message}`);
      throw error;
    }
  }
  
  /**
   * Gera hash de senha utilizando bcrypt
   * 
   * @param {string} password Senha em texto plano
   * @param {Object} options Opções de bcrypt
   * @returns {Promise<Object>} Hash e metadados
   */
  async hashBcrypt(password, options = {}) {
    const rounds = options.rounds || this.defaultOptions.bcrypt.rounds;
    const hash = await bcrypt.hash(password, rounds);
    
    return {
      hash,
      salt: null, // O salt está incorporado no hash
      algorithm: 'bcrypt',
      iterations: rounds
    };
  }
  
  /**
   * Gera hash de senha utilizando Argon2id
   * 
   * @param {string} password Senha em texto plano
   * @param {Object} options Opções de Argon2
   * @returns {Promise<Object>} Hash e metadados
   */
  async hashArgon2id(password, options = {}) {
    const salt = Buffer.from(options.salt || this.generateSalt(), 'base64');
    
    const argonOptions = {
      type: argon2.argon2id,
      memoryCost: options.memoryCost || this.defaultOptions.argon2id.memoryCost,
      timeCost: options.timeCost || this.defaultOptions.argon2id.timeCost,
      parallelism: options.parallelism || this.defaultOptions.argon2id.parallelism,
      salt: salt
    };
    
    const hash = await argon2.hash(password, argonOptions);
    
    return {
      hash,
      salt: salt.toString('base64'),
      algorithm: 'argon2id',
      memoryCost: argonOptions.memoryCost,
      timeCost: argonOptions.timeCost,
      parallelism: argonOptions.parallelism
    };
  }
  
  /**
   * Gera hash de senha utilizando PBKDF2
   * 
   * @param {string} password Senha em texto plano
   * @param {string} digest Algoritmo de hash (sha256 ou sha512)
   * @param {Object} options Opções de PBKDF2
   * @returns {Promise<Object>} Hash e metadados
   */
  async hashPbkdf2(password, digest = 'sha256', options = {}) {
    const salt = Buffer.from(options.salt || this.generateSalt(), 'base64');
    
    // Seleciona as opções corretas dependendo do digest
    const defaultOpts = (digest === 'sha512') ? 
      this.defaultOptions['pbkdf2-sha512'] : 
      this.defaultOptions['pbkdf2-sha256'];
    
    const iterations = options.iterations || defaultOpts.iterations;
    const keylen = options.keylen || defaultOpts.keylen;
    
    // Gera o hash PBKDF2
    return new Promise((resolve, reject) => {
      crypto.pbkdf2(password, salt, iterations, keylen, digest, (err, derivedKey) => {
        if (err) {
          reject(err);
          return;
        }
        
        // Formata o resultado
        const hashResult = `$pbkdf2-${digest}$i=${iterations},l=${keylen}$${salt.toString('base64')}$${derivedKey.toString('base64')}`;
        
        resolve({
          hash: hashResult,
          salt: salt.toString('base64'),
          algorithm: `pbkdf2-${digest}`,
          iterations,
          keylen
        });
      });
    });
  }
  
  /**
   * Gera hash de senha utilizando scrypt
   * 
   * @param {string} password Senha em texto plano
   * @param {Object} options Opções de scrypt
   * @returns {Promise<Object>} Hash e metadados
   */
  async hashScrypt(password, options = {}) {
    const salt = Buffer.from(options.salt || this.generateSalt(), 'base64');
    
    const cost = options.cost || this.defaultOptions.scrypt.cost;
    const blockSize = options.blockSize || this.defaultOptions.scrypt.blockSize;
    const parallelization = options.parallelization || this.defaultOptions.scrypt.parallelization;
    const keylen = options.keylen || this.defaultOptions.scrypt.keylen;
    
    // Gera o hash scrypt
    return new Promise((resolve, reject) => {
      crypto.scrypt(password, salt, keylen, {
        cost,
        blockSize,
        parallelization
      }, (err, derivedKey) => {
        if (err) {
          reject(err);
          return;
        }
        
        // Formata o resultado
        const hashResult = `$scrypt$n=${cost},r=${blockSize},p=${parallelization},l=${keylen}$${salt.toString('base64')}$${derivedKey.toString('base64')}`;
        
        resolve({
          hash: hashResult,
          salt: salt.toString('base64'),
          algorithm: 'scrypt',
          cost,
          blockSize,
          parallelization,
          keylen
        });
      });
    });
  }
  
  /**
   * Verifica se uma senha corresponde a um hash
   * 
   * @param {string} password Senha em texto plano a verificar
   * @param {string} storedHash Hash armazenado
   * @param {string} algorithm Algoritmo utilizado
   * @param {Object} options Opções do algoritmo
   * @returns {Promise<boolean>} Verdadeiro se a senha for válida
   */
  async verifyPassword(password, storedHash, algorithm, options = {}) {
    // Normaliza algoritmo para caixa minúscula
    algorithm = (algorithm || '').toLowerCase();
    
    try {
      switch (algorithm) {
        case 'bcrypt':
          return await bcrypt.compare(password, storedHash);
        
        case 'argon2id':
          // Para Argon2, o hash já contém todas as informações necessárias
          return await argon2.verify(storedHash, password);
        
        case 'pbkdf2-sha256':
        case 'pbkdf2-sha512':
          return await this.verifyPbkdf2(password, storedHash);
        
        case 'scrypt':
          return await this.verifyScrypt(password, storedHash);
        
        default:
          // Se o algoritmo não for reconhecido, tenta detectar o formato
          if (storedHash.startsWith('$2')) {
            // Formato bcrypt
            return await bcrypt.compare(password, storedHash);
          } else if (storedHash.startsWith('$argon2id$')) {
            // Formato Argon2id
            return await argon2.verify(storedHash, password);
          } else if (storedHash.startsWith('$pbkdf2-')) {
            // Formato PBKDF2
            return await this.verifyPbkdf2(password, storedHash);
          } else if (storedHash.startsWith('$scrypt$')) {
            // Formato scrypt
            return await this.verifyScrypt(password, storedHash);
          }
          
          throw new Error(`Algoritmo de hash não suportado ou formato inválido: ${algorithm}`);
      }
    } catch (error) {
      logger.error(`[PasswordService] Erro ao verificar senha: ${error.message}`);
      return false;
    }
  }
  
  /**
   * Verifica uma senha com hash PBKDF2
   * 
   * @param {string} password Senha em texto plano
   * @param {string} storedHash Hash armazenado no formato PBKDF2
   * @returns {Promise<boolean>} Verdadeiro se a senha for válida
   */
  async verifyPbkdf2(password, storedHash) {
    // Formato: $pbkdf2-{digest}$i={iterations},l={keylen}${base64_salt}${base64_hash}
    const parts = storedHash.split('$');
    if (parts.length !== 5) {
      throw new Error('Formato de hash PBKDF2 inválido');
    }
    
    // Extrai informações do hash
    const algorithm = parts[1]; // pbkdf2-sha256 ou pbkdf2-sha512
    const digest = algorithm.split('-')[1]; // sha256 ou sha512
    
    const params = {};
    parts[2].split(',').forEach(param => {
      const [key, value] = param.split('=');
      params[key] = parseInt(value, 10);
    });
    
    const salt = Buffer.from(parts[3], 'base64');
    const storedKey = Buffer.from(parts[4], 'base64');
    
    // Gera o hash com os mesmos parâmetros
    return new Promise((resolve, reject) => {
      crypto.pbkdf2(password, salt, params.i, params.l, digest, (err, derivedKey) => {
        if (err) {
          reject(err);
          return;
        }
        
        // Compara os hashes em tempo constante
        resolve(crypto.timingSafeEqual(derivedKey, storedKey));
      });
    });
  }
  
  /**
   * Verifica uma senha com hash scrypt
   * 
   * @param {string} password Senha em texto plano
   * @param {string} storedHash Hash armazenado no formato scrypt
   * @returns {Promise<boolean>} Verdadeiro se a senha for válida
   */
  async verifyScrypt(password, storedHash) {
    // Formato: $scrypt$n={cost},r={blockSize},p={parallelization},l={keylen}${base64_salt}${base64_hash}
    const parts = storedHash.split('$');
    if (parts.length !== 5) {
      throw new Error('Formato de hash scrypt inválido');
    }
    
    // Extrai informações do hash
    const params = {};
    parts[2].split(',').forEach(param => {
      const [key, value] = param.split('=');
      params[key] = parseInt(value, 10);
    });
    
    const salt = Buffer.from(parts[3], 'base64');
    const storedKey = Buffer.from(parts[4], 'base64');
    
    // Gera o hash com os mesmos parâmetros
    return new Promise((resolve, reject) => {
      crypto.scrypt(password, salt, params.l, {
        cost: params.n,
        blockSize: params.r,
        parallelization: params.p
      }, (err, derivedKey) => {
        if (err) {
          reject(err);
          return;
        }
        
        // Compara os hashes em tempo constante
        resolve(crypto.timingSafeEqual(derivedKey, storedKey));
      });
    });
  }
  
  /**
   * Verifica se uma senha precisa ser atualizada para um algoritmo mais seguro
   * 
   * @param {string} algorithm Algoritmo atual
   * @param {Object} options Opções atuais
   * @returns {boolean} Verdadeiro se a senha precisa ser atualizada
   */
  needsUpgrade(algorithm, options = {}) {
    algorithm = (algorithm || '').toLowerCase();
    
    // Verifica se está usando um algoritmo obsoleto
    if (['md5', 'sha1', 'sha256', 'sha512'].includes(algorithm)) {
      return true;
    }
    
    // Verifica configurações específicas de algoritmos
    switch (algorithm) {
      case 'bcrypt':
        return (options.rounds || 0) < 12;
        
      case 'pbkdf2-sha256':
        return (options.iterations || 0) < 600000;
        
      case 'pbkdf2-sha512':
        return (options.iterations || 0) < 210000;
        
      case 'argon2id':
        return false; // Argon2id é considerado seguro com as configurações padrão
        
      case 'scrypt':
        return (options.cost || 0) < 16384;
        
      default:
        // Se não reconhecer o algoritmo, recomenda atualização
        return true;
    }
  }
  
  /**
   * Sugestão de algoritmo e parâmetros com base no contexto
   * 
   * @param {Object} context Contexto para decisão
   * @returns {Object} Algoritmo e parâmetros recomendados
   */
  getRecommendedAlgorithm(context = {}) {
    // Contexto contém informações como região, indústria, etc.
    
    // Para setores de alta segurança como finanças e saúde
    if (context.industry === 'finance' || context.industry === 'healthcare' || 
        context.industry === 'government') {
      return {
        algorithm: 'argon2id',
        options: {
          memoryCost: 131072, // 128MB
          timeCost: 4,
          parallelism: 2
        }
      };
    }
    
    // Para regiões com dispositivos de menor capacidade
    if (context.region === 'AO') {
      return {
        algorithm: 'bcrypt',
        options: {
          rounds: 12
        }
      };
    }
    
    // Padrão para outras situações
    return {
      algorithm: 'argon2id',
      options: {
        memoryCost: 65536, // 64MB
        timeCost: 3,
        parallelism: 1
      }
    };
  }
}

module.exports = PasswordService;
