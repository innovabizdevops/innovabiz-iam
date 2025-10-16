/**
 * Plugin de Senha Tradicional (K01)
 * 
 * Implementa o método de autenticação baseado em senha tradicional,
 * com suporte a diferentes algoritmos de hash, verificação de força
 * e adaptações regionais.
 * 
 * @module traditional-password
 * @author INNOVABIZ
 * @copyright 2025 INNOVABIZ
 * @license Proprietary
 */

const argon2 = require('argon2');
const crypto = require('crypto');
const zxcvbn = require('zxcvbn');
const { logger } = require('../../core/utils/logging');

/**
 * Classe principal do plugin de senha tradicional
 */
class TraditionalPasswordPlugin {
  /**
   * Construtor do plugin
   * 
   * @param {Object} manifest Manifesto do plugin
   */
  constructor(manifest) {
    this.manifest = manifest;
    this.settings = {
      hashSettings: {
        // Configurações do Argon2id (seguro para 2025+)
        argon2: {
          memoryCost: 65536, // 64 MB
          timeCost: 3, // 3 iterações
          parallelism: 2, // 2 threads
          type: argon2.argon2id, // variante mais segura para senhas
          saltLength: 16, // tamanho do salt em bytes
          hashLength: 32 // tamanho do hash em bytes
        }
      },
      passwordPolicies: {
        default: {
          min_length: 10,
          require_special_chars: true,
          require_numbers: true,
          require_upper_lower: true,
          max_age_days: 90,
          history_check: 5,
          min_strength_score: 3,
          common_password_check: true,
          breach_check: true
        }
      },
      breachAPI: {
        enabled: true,
        endpoint: 'https://api.pwnedpasswords.com/range/',
        timeout: 2000
      }
    };
    
    logger.info('[TraditionalPassword] Plugin de senha tradicional inicializado');
  }
  
  /**
   * Inicia o processo de autenticação
   * 
   * @param {Object} request Requisição de autenticação
   * @param {Object} context Contexto da autenticação
   * @returns {Promise<Object>} Resultado da autenticação
   */
  async startAuthentication(request, context) {
    try {
      // Validação básica de parâmetros
      if (!request.username || !request.password) {
        return {
          success: false,
          code: 'MISSING_CREDENTIALS',
          message: 'Nome de usuário e senha são obrigatórios',
          nextAction: 'retry'
        };
      }
      
      // Adaptações regionais
      const region = context.region || 'default';
      const regionalSettings = this.getRegionalSettings(region);
      
      // Recuperação de dados do usuário (simulado)
      const user = await this.getUserData(request.username, request.tenantId);
      
      if (!user) {
        // Em produção: limitar informações para evitar enumeração de usuários
        // Mesmo tempo de resposta para usuários existentes e não existentes
        await this.simulateProcessingTime();
        
        logger.info(`[TraditionalPassword] Tentativa com usuário inexistente: ${request.username}`);
        
        return {
          success: false,
          code: 'INVALID_CREDENTIALS',
          message: 'Credenciais inválidas',
          nextAction: 'retry'
        };
      }
      
      // Verifica se o usuário está bloqueado
      if (user.status === 'locked') {
        logger.warn(`[TraditionalPassword] Tentativa em conta bloqueada: ${request.username}`);
        
        return {
          success: false,
          code: 'ACCOUNT_LOCKED',
          message: 'Conta temporariamente bloqueada. Tente novamente mais tarde.',
          nextAction: 'abort'
        };
      }
      
      // Verifica se a senha está expirada
      if (this.isPasswordExpired(user, regionalSettings)) {
        logger.info(`[TraditionalPassword] Senha expirada: ${request.username}`);
        
        return {
          success: false,
          code: 'PASSWORD_EXPIRED',
          message: 'Sua senha expirou e precisa ser alterada.',
          nextAction: 'password_reset',
          resetToken: await this.generatePasswordResetToken(user)
        };
      }
      
      // Verifica a senha
      const passwordValid = await this.verifyPassword(request.password, user.password_hash);
      
      if (!passwordValid) {
        // Registra tentativa inválida
        await this.registerFailedAttempt(user);
        
        logger.info(`[TraditionalPassword] Senha inválida para usuário: ${request.username}`);
        
        // Verifica se deve bloquear a conta
        const failedAttempts = user.failed_attempts || 0;
        const maxAttempts = regionalSettings.max_attempts || 5;
        
        if (failedAttempts >= maxAttempts) {
          // Bloqueia a conta temporariamente
          await this.lockAccount(user);
          
          logger.warn(`[TraditionalPassword] Conta bloqueada após ${failedAttempts} tentativas: ${request.username}`);
          
          return {
            success: false,
            code: 'ACCOUNT_LOCKED',
            message: 'Conta temporariamente bloqueada. Tente novamente mais tarde.',
            nextAction: 'abort'
          };
        }
        
        return {
          success: false,
          code: 'INVALID_CREDENTIALS',
          message: 'Credenciais inválidas',
          nextAction: 'retry',
          attemptsLeft: maxAttempts - failedAttempts
        };
      }
      
      // Autenticação bem-sucedida
      
      // Limpa contador de tentativas
      await this.clearFailedAttempts(user);
      
      // Registra login bem-sucedido
      await this.registerSuccessfulLogin(user, context);
      
      logger.info(`[TraditionalPassword] Autenticação bem-sucedida: ${request.username}`);
      
      // Verifica se senha precisa ser atualizada (mas não está expirada)
      let passwordStatus = 'valid';
      let passwordStrength = this.checkPasswordStrength(request.password);
      
      if (passwordStrength.score < regionalSettings.min_strength_score) {
        passwordStatus = 'weak';
      }
      
      return {
        success: true,
        user_id: user.id,
        username: user.username,
        display_name: user.display_name,
        auth_level: 'single_factor',
        timestamp: new Date(),
        session_data: {
          auth_method: 'K01',
          auth_time: Date.now(),
          requires_mfa: user.mfa_required || context.risk_level === 'high'
        },
        password_status: passwordStatus,
        requires_action: passwordStatus === 'weak' ? 'PASSWORD_UPGRADE_RECOMMENDED' : null
      };
    } catch (error) {
      logger.error(`[TraditionalPassword] Erro na autenticação: ${error.message}`);
      
      return {
        success: false,
        code: 'AUTHENTICATION_ERROR',
        message: 'Erro no processamento da autenticação',
        nextAction: 'retry'
      };
    }
  }
  
  /**
   * Simula tempo de processamento para evitar timing attacks
   * 
   * @returns {Promise<void>}
   */
  async simulateProcessingTime() {
    return new Promise(resolve => {
      const randomTime = 100 + Math.floor(Math.random() * 200);
      setTimeout(resolve, randomTime);
    });
  }
  
  /**
   * Obtém as configurações regionais
   * 
   * @param {string} region Código da região
   * @returns {Object} Configurações regionais
   */
  getRegionalSettings(region) {
    // Tenta obter do manifesto
    if (this.manifest.regional_adaptations && this.manifest.regional_adaptations[region]) {
      return {
        ...this.settings.passwordPolicies.default,
        ...this.manifest.regional_adaptations[region],
        max_attempts: 5, // Consistente em todas as regiões
        lockout_duration: 15 * 60 // 15 minutos em segundos
      };
    }
    
    // Padrão
    return {
      ...this.settings.passwordPolicies.default,
      max_attempts: 5,
      lockout_duration: 15 * 60
    };
  }
  
  /**
   * Busca dados do usuário
   * 
   * @param {string} username Nome de usuário
   * @param {string} tenantId ID do tenant
   * @returns {Promise<Object>} Dados do usuário
   */
  async getUserData(username, tenantId) {
    try {
      // Em produção: busca no banco de dados ou serviço de usuários
      // Para esta demonstração, usuário simulado
      
      // Simulação de comunicação com banco de dados
      await new Promise(resolve => setTimeout(resolve, 50));
      
      // Usuário de demonstração
      return {
        id: 'user123',
        username: username,
        display_name: 'Usuário de Teste',
        password_hash: '$argon2id$v=19$m=65536,t=3,p=2$VlWJsrjPz/rXR+xrPgB1bg$oX9NkG8UJHg73jFiD9v/c5XLNZJWUrf4Sn6kXpdfPXQ', // "password123"
        status: 'active',
        tenant_id: tenantId || 'default',
        failed_attempts: 0,
        last_password_change: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000), // 30 dias atrás
        mfa_required: false,
        password_history: [],
        created_at: new Date(Date.now() - 180 * 24 * 60 * 60 * 1000) // 180 dias atrás
      };
    } catch (error) {
      logger.error(`[TraditionalPassword] Erro ao buscar usuário: ${error.message}`);
      return null;
    }
  }
  
  /**
   * Verifica se uma senha está expirada
   * 
   * @param {Object} user Dados do usuário
   * @param {Object} settings Configurações regionais
   * @returns {boolean} Verdadeiro se a senha está expirada
   */
  isPasswordExpired(user, settings) {
    if (!user.last_password_change) {
      return false;
    }
    
    const maxAgeDays = settings.max_age_days || 90;
    const passwordAge = (Date.now() - user.last_password_change.getTime()) / (1000 * 60 * 60 * 24);
    
    return passwordAge > maxAgeDays;
  }
  
  /**
   * Verifica uma senha contra um hash armazenado
   * 
   * @param {string} password Senha fornecida
   * @param {string} storedHash Hash armazenado
   * @returns {Promise<boolean>} Verdadeiro se a senha é válida
   */
  async verifyPassword(password, storedHash) {
    try {
      // Verifica se é um hash Argon2
      if (storedHash.startsWith('$argon2')) {
        return await argon2.verify(storedHash, password);
      }
      
      // Suporte a migração de hashes legados
      if (storedHash.startsWith('$pbkdf2$')) {
        // Implementação para PBKDF2
        // ...
        return false; // Placeholder
      }
      
      if (storedHash.startsWith('$2a$') || storedHash.startsWith('$2b$')) {
        // Implementação para bcrypt
        // ...
        return false; // Placeholder
      }
      
      // Hash não reconhecido
      logger.warn('[TraditionalPassword] Formato de hash não reconhecido');
      return false;
    } catch (error) {
      logger.error(`[TraditionalPassword] Erro ao verificar senha: ${error.message}`);
      return false;
    }
  }
  
  /**
   * Gera um novo hash para senha
   * 
   * @param {string} password Senha em texto plano
   * @returns {Promise<string>} Hash da senha
   */
  async hashPassword(password) {
    try {
      const { memoryCost, timeCost, parallelism, type, hashLength } = this.settings.hashSettings.argon2;
      
      const salt = crypto.randomBytes(this.settings.hashSettings.argon2.saltLength);
      
      const hash = await argon2.hash(password, {
        type,
        memoryCost,
        timeCost,
        parallelism,
        salt,
        hashLength
      });
      
      return hash;
    } catch (error) {
      logger.error(`[TraditionalPassword] Erro ao gerar hash: ${error.message}`);
      throw error;
    }
  }
  
  /**
   * Verifica se a senha já foi comprometida em vazamentos
   * 
   * @param {string} password Senha a verificar
   * @returns {Promise<boolean>} Verdadeiro se a senha foi comprometida
   */
  async checkPasswordBreached(password) {
    if (!this.settings.breachAPI.enabled) {
      return false;
    }
    
    try {
      // Calcula o hash SHA-1 da senha
      const sha1Hash = crypto
        .createHash('sha1')
        .update(password)
        .digest('hex')
        .toUpperCase();
      
      // Primeiros 5 caracteres para consulta k-anonimity
      const prefix = sha1Hash.substring(0, 5);
      const suffix = sha1Hash.substring(5);
      
      // Consulta a API (simulada para o exemplo)
      // Em produção, faria uma requisição HTTP real
      const response = this.simulateBreachAPIResponse(prefix);
      
      // Verifica se o sufixo está presente na resposta
      return response.includes(suffix);
    } catch (error) {
      logger.error(`[TraditionalPassword] Erro ao verificar vazamentos: ${error.message}`);
      return false;
    }
  }
  
  /**
   * Simulação da resposta da API de verificação de vazamentos
   * 
   * @param {string} prefix Prefixo do hash
   * @returns {string} Resposta simulada
   */
  simulateBreachAPIResponse(prefix) {
    // Em produção: faria uma requisição HTTP real
    // Para o exemplo, resposta simulada
    
    // Simulação: certos prefixos são considerados comprometidos
    const compromisedPrefixes = ['5BAA6', '1D72C', 'E38AD'];
    
    if (compromisedPrefixes.includes(prefix)) {
      return `
        1D72CD0736C3927C11C5C8216157FC15F:3
        1D72CD0736C3927C11C5C8216157FC15F:5
        5BAA61E4C9B93F3F0682250B6CF8331B7EE68FD8:3730471
      `;
    }
    
    return '';
  }
  
  /**
   * Avalia a força de uma senha
   * 
   * @param {string} password Senha a avaliar
   * @returns {Object} Resultado da avaliação
   */
  checkPasswordStrength(password) {
    // Utiliza a biblioteca zxcvbn para avaliação de força
    const result = zxcvbn(password);
    
    // Adapta o resultado
    return {
      score: result.score, // 0-4, onde 4 é a mais forte
      feedback: result.feedback,
      estimatedCrackTime: result.crack_times_seconds.offline_slow_hashing_1e4_per_second,
      estimatedCrackTimeText: result.crack_times_display.offline_slow_hashing_1e4_per_second
    };
  }
  
  /**
   * Verifica se uma senha atende aos requisitos da política
   * 
   * @param {string} password Senha a verificar
   * @param {Object} policy Política de senha
   * @returns {Object} Resultado da validação
   */
  validatePasswordPolicy(password, policy) {
    const result = {
      valid: true,
      issues: []
    };
    
    // Verifica comprimento mínimo
    if (password.length < policy.min_length) {
      result.valid = false;
      result.issues.push({
        code: 'LENGTH',
        message: `A senha deve ter pelo menos ${policy.min_length} caracteres`
      });
    }
    
    // Verifica caracteres especiais
    if (policy.require_special_chars && !/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password)) {
      result.valid = false;
      result.issues.push({
        code: 'SPECIAL_CHAR',
        message: 'A senha deve conter pelo menos um caractere especial'
      });
    }
    
    // Verifica números
    if (policy.require_numbers && !/\d/.test(password)) {
      result.valid = false;
      result.issues.push({
        code: 'NUMBER',
        message: 'A senha deve conter pelo menos um número'
      });
    }
    
    // Verifica maiúsculas e minúsculas
    if (policy.require_upper_lower && (!/[a-z]/.test(password) || !/[A-Z]/.test(password))) {
      result.valid = false;
      result.issues.push({
        code: 'UPPER_LOWER',
        message: 'A senha deve conter letras maiúsculas e minúsculas'
      });
    }
    
    // Verifica força da senha
    const strength = this.checkPasswordStrength(password);
    if (strength.score < policy.min_strength_score) {
      result.valid = false;
      result.issues.push({
        code: 'STRENGTH',
        message: 'A senha é muito fraca ou previsível',
        details: strength.feedback
      });
    }
    
    return result;
  }
  
  /**
   * Registra uma tentativa de login inválida
   * 
   * @param {Object} user Dados do usuário
   * @returns {Promise<void>}
   */
  async registerFailedAttempt(user) {
    // Em produção: atualiza no banco de dados
    // Para o exemplo, apenas log
    
    user.failed_attempts = (user.failed_attempts || 0) + 1;
    user.last_failed_attempt = new Date();
    
    logger.debug(`[TraditionalPassword] Tentativa falha registrada para ${user.username} (total: ${user.failed_attempts})`);
  }
  
  /**
   * Limpa o contador de tentativas inválidas
   * 
   * @param {Object} user Dados do usuário
   * @returns {Promise<void>}
   */
  async clearFailedAttempts(user) {
    // Em produção: atualiza no banco de dados
    
    user.failed_attempts = 0;
    user.last_failed_attempt = null;
    
    logger.debug(`[TraditionalPassword] Contador de tentativas resetado para ${user.username}`);
  }
  
  /**
   * Bloqueia temporariamente uma conta
   * 
   * @param {Object} user Dados do usuário
   * @returns {Promise<void>}
   */
  async lockAccount(user) {
    // Em produção: atualiza no banco de dados
    
    user.status = 'locked';
    user.lock_expires_at = new Date(Date.now() + (15 * 60 * 1000)); // 15 minutos
    
    logger.warn(`[TraditionalPassword] Conta bloqueada: ${user.username}`);
  }
  
  /**
   * Registra um login bem-sucedido
   * 
   * @param {Object} user Dados do usuário
   * @param {Object} context Contexto da autenticação
   * @returns {Promise<void>}
   */
  async registerSuccessfulLogin(user, context) {
    // Em produção: atualiza no banco de dados
    
    user.last_login = new Date();
    user.last_login_ip = context.ip;
    user.last_login_device = context.deviceId;
    
    logger.debug(`[TraditionalPassword] Login bem-sucedido registrado para ${user.username}`);
  }
  
  /**
   * Gera um token para redefinição de senha
   * 
   * @param {Object} user Dados do usuário
   * @returns {Promise<string>} Token de redefinição
   */
  async generatePasswordResetToken(user) {
    // Em produção: armazena em banco de dados
    
    const token = crypto.randomBytes(32).toString('hex');
    
    logger.debug(`[TraditionalPassword] Token de redefinição gerado para ${user.username}`);
    
    return token;
  }
  
  /**
   * Altera a senha de um usuário
   * 
   * @param {Object} request Requisição de alteração
   * @param {Object} context Contexto da operação
   * @returns {Promise<Object>} Resultado da operação
   */
  async changePassword(request, context) {
    try {
      // Validação básica
      if (!request.userId || !request.currentPassword || !request.newPassword) {
        return {
          success: false,
          code: 'MISSING_PARAMETERS',
          message: 'Parâmetros insuficientes para alteração de senha'
        };
      }
      
      // Recupera dados do usuário
      const user = await this.getUserById(request.userId);
      
      if (!user) {
        return {
          success: false,
          code: 'USER_NOT_FOUND',
          message: 'Usuário não encontrado'
        };
      }
      
      // Adaptações regionais
      const region = context.region || 'default';
      const regionalSettings = this.getRegionalSettings(region);
      
      // Verifica senha atual
      const passwordValid = await this.verifyPassword(request.currentPassword, user.password_hash);
      
      if (!passwordValid) {
        logger.info(`[TraditionalPassword] Senha atual inválida na alteração: ${user.username}`);
        
        return {
          success: false,
          code: 'INVALID_CURRENT_PASSWORD',
          message: 'Senha atual inválida'
        };
      }
      
      // Valida nova senha contra política
      const validation = this.validatePasswordPolicy(request.newPassword, regionalSettings);
      
      if (!validation.valid) {
        return {
          success: false,
          code: 'POLICY_VIOLATION',
          message: 'A nova senha não atende aos requisitos de segurança',
          details: validation.issues
        };
      }
      
      // Verifica histórico de senhas
      if (await this.isPasswordInHistory(user, request.newPassword, regionalSettings.history_check)) {
        return {
          success: false,
          code: 'PASSWORD_REUSE',
          message: 'A nova senha não pode ser igual às senhas anteriores'
        };
      }
      
      // Verifica se senha foi comprometida
      if (regionalSettings.breach_check && await this.checkPasswordBreached(request.newPassword)) {
        return {
          success: false,
          code: 'PASSWORD_COMPROMISED',
          message: 'Esta senha foi comprometida em vazamentos de dados. Escolha outra senha.'
        };
      }
      
      // Gera novo hash
      const newHash = await this.hashPassword(request.newPassword);
      
      // Atualiza senha (simulado)
      await this.updatePassword(user, newHash, request.currentPassword);
      
      logger.info(`[TraditionalPassword] Senha alterada com sucesso: ${user.username}`);
      
      return {
        success: true,
        message: 'Senha alterada com sucesso',
        timestamp: new Date()
      };
    } catch (error) {
      logger.error(`[TraditionalPassword] Erro na alteração de senha: ${error.message}`);
      
      return {
        success: false,
        code: 'CHANGE_ERROR',
        message: 'Erro no processamento da alteração de senha'
      };
    }
  }
  
  /**
   * Busca usuário por ID
   * 
   * @param {string} userId ID do usuário
   * @returns {Promise<Object>} Dados do usuário
   */
  async getUserById(userId) {
    // Em produção: busca no banco de dados
    // Para a demonstração, usuário simulado
    return {
      id: userId,
      username: 'usuario.teste',
      display_name: 'Usuário de Teste',
      password_hash: '$argon2id$v=19$m=65536,t=3,p=2$VlWJsrjPz/rXR+xrPgB1bg$oX9NkG8UJHg73jFiD9v/c5XLNZJWUrf4Sn6kXpdfPXQ', // "password123"
      status: 'active',
      tenant_id: 'default',
      failed_attempts: 0,
      last_password_change: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000), // 30 dias atrás
      mfa_required: false,
      password_history: [],
      created_at: new Date(Date.now() - 180 * 24 * 60 * 60 * 1000) // 180 dias atrás
    };
  }
  
  /**
   * Verifica se uma senha está no histórico do usuário
   * 
   * @param {Object} user Dados do usuário
   * @param {string} newPassword Nova senha
   * @param {number} historyLength Tamanho do histórico a verificar
   * @returns {Promise<boolean>} Verdadeiro se a senha está no histórico
   */
  async isPasswordInHistory(user, newPassword, historyLength) {
    // Em produção: verificaria histórico no banco de dados
    // Para demonstração, resultado simulado
    return false;
  }
  
  /**
   * Atualiza a senha de um usuário
   * 
   * @param {Object} user Dados do usuário
   * @param {string} newPasswordHash Novo hash da senha
   * @param {string} oldPassword Senha anterior (para histórico)
   * @returns {Promise<void>}
   */
  async updatePassword(user, newPasswordHash, oldPassword) {
    // Em produção: atualiza no banco de dados
    
    // Adiciona senha atual ao histórico
    if (user.password_hash) {
      if (!user.password_history) {
        user.password_history = [];
      }
      
      user.password_history.push({
        hash: user.password_hash,
        changed_at: user.last_password_change || new Date()
      });
    }
    
    // Atualiza hash
    user.password_hash = newPasswordHash;
    user.last_password_change = new Date();
    
    logger.debug(`[TraditionalPassword] Senha atualizada para ${user.username}`);
  }
  
  /**
   * Inicializa o plugin
   * 
   * @returns {Promise<void>}
   */
  async initialize() {
    // Carrega configurações, inicializa recursos, etc.
    logger.info('[TraditionalPassword] Plugin inicializado com sucesso');
  }
  
  /**
   * Finaliza o plugin
   * 
   * @returns {Promise<void>}
   */
  async shutdown() {
    // Libera recursos, fecha conexões, etc.
    logger.info('[TraditionalPassword] Plugin finalizado');
  }
}

module.exports = TraditionalPasswordPlugin;
