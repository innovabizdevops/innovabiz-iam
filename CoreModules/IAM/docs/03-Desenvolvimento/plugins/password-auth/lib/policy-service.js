/**
 * Serviço de Políticas de Senha
 * 
 * Este módulo gerencia as políticas de senha, incluindo validação,
 * histórico, e regras de complexidade adaptadas para diferentes regiões
 * e requisitos setoriais.
 * 
 * @module policy-service
 * @author INNOVABIZ
 * @copyright 2025 INNOVABIZ
 * @license Proprietary
 */

const { logger } = require('@innovabiz/auth-framework/utils/logging');
const DatabaseService = require('@innovabiz/auth-framework/services/database');

/**
 * Serviço de gerenciamento de políticas de senha
 */
class PolicyService {
  /**
   * Cria uma nova instância do serviço de políticas
   */
  constructor() {
    this.db = new DatabaseService();
    this.commonPasswordCache = new Set();
    this.lockoutCache = new Map(); // Cache em memória para bloqueios temporários

    // Carrega senhas comuns para cache (versão simplificada)
    this.loadCommonPasswords();
    
    logger.info('[PolicyService] Inicializado');
  }
  
  /**
   * Carrega uma lista de senhas comuns para cache
   * Em produção, isso carregaria de um arquivo ou banco de dados
   */
  async loadCommonPasswords() {
    // Lista muito reduzida para demonstração
    const commonPasswords = [
      'password', '123456', 'qwerty', 'admin', 'welcome',
      'senha', '123456789', 'abc123', 'password1', '1234567890',
      'letmein', 'passw0rd', 'trustno1', 'senha123', 'admin123'
    ];
    
    commonPasswords.forEach(pwd => this.commonPasswordCache.add(pwd));
    logger.info(`[PolicyService] Carregadas ${this.commonPasswordCache.size} senhas comuns para verificação`);
  }
  
  /**
   * Obtém a política de senha aplicável a um usuário
   * 
   * @param {string} tenantId ID do tenant
   * @param {string} userId ID do usuário
   * @param {Object} context Contexto adicional (região, indústria, etc.)
   * @returns {Promise<Object>} Política de senha
   */
  async getPasswordPolicy(tenantId, userId, context = {}) {
    try {
      // Busca a política no banco de dados baseada em prioridade
      const query = `
        SELECT p.* 
        FROM iam.password_policies p
        JOIN iam.password_policy_assignments pa 
          ON p.policy_id = pa.policy_id AND p.tenant_id = pa.tenant_id
        WHERE p.tenant_id = $1
        AND p.status = 'active'
        AND (
          (pa.target_type = 'user' AND pa.target_id = $2) OR
          (pa.target_type = 'group' AND pa.target_id IN (
            SELECT group_id FROM iam.group_memberships 
            WHERE user_id = $2 AND status = 'active'
          )) OR
          (pa.target_type = 'org_unit' AND pa.target_id IN (
            SELECT org_unit_id FROM iam.users 
            WHERE user_id = $2
          )) OR
          (pa.target_type = 'all' AND pa.target_id IS NULL)
        )
        ORDER BY CASE 
          WHEN pa.target_type = 'user' THEN 1
          WHEN pa.target_type = 'group' THEN 2
          WHEN pa.target_type = 'org_unit' THEN 3
          WHEN pa.target_type = 'all' THEN 4
        END
        LIMIT 1
      `;
      
      const result = await this.db.query(query, [tenantId, userId]);
      
      if (result.rows.length > 0) {
        const policy = result.rows[0];
        
        // Aplica adaptações regionais se necessário
        if (context.region && policy.regional_settings && 
            policy.regional_settings[context.region]) {
          return this.applyRegionalSettings(policy, context.region);
        }
        
        // Aplica adaptações setoriais se necessário
        if (context.industry && policy.industry_settings && 
            policy.industry_settings[context.industry]) {
          return this.applyIndustrySettings(policy, context.industry);
        }
        
        return policy;
      }
      
      // Se não encontrar política específica, busca a política padrão do tenant
      const defaultQuery = `
        SELECT * FROM iam.password_policies
        WHERE tenant_id = $1
        AND status = 'active'
        ORDER BY created_at DESC
        LIMIT 1
      `;
      
      const defaultResult = await this.db.query(defaultQuery, [tenantId]);
      
      if (defaultResult.rows.length > 0) {
        const policy = defaultResult.rows[0];
        
        // Aplica adaptações regionais/setoriais como acima
        if (context.region && policy.regional_settings && 
            policy.regional_settings[context.region]) {
          return this.applyRegionalSettings(policy, context.region);
        }
        
        if (context.industry && policy.industry_settings && 
            policy.industry_settings[context.industry]) {
          return this.applyIndustrySettings(policy, context.industry);
        }
        
        return policy;
      }
      
      // Se ainda não encontrar, usa política do sistema com adaptações regionais
      let systemPolicy;
      
      if (context.region) {
        const regionQuery = `
          SELECT * FROM iam.password_policies
          WHERE tenant_id = (SELECT tenant_id FROM iam.tenants WHERE name = 'system' LIMIT 1)
          AND name = $1
          LIMIT 1
        `;
        
        const regionName = this.getRegionalPolicyName(context.region);
        const regionResult = await this.db.query(regionQuery, [regionName]);
        
        if (regionResult.rows.length > 0) {
          systemPolicy = regionResult.rows[0];
        }
      }
      
      // Se não encontrar política regional, usa a mais restritiva (EU)
      if (!systemPolicy) {
        const systemQuery = `
          SELECT * FROM iam.password_policies
          WHERE tenant_id = (SELECT tenant_id FROM iam.tenants WHERE name = 'system' LIMIT 1)
          AND name = 'EU Standard Password Policy'
          LIMIT 1
        `;
        
        const systemResult = await this.db.query(systemQuery);
        
        if (systemResult.rows.length > 0) {
          systemPolicy = systemResult.rows[0];
        }
      }
      
      // Se ainda não encontrar, retorna política padrão codificada
      if (!systemPolicy) {
        return this.getDefaultPolicy(context);
      }
      
      return systemPolicy;
      
    } catch (error) {
      logger.error(`[PolicyService] Erro ao obter política de senha: ${error.message}`);
      
      // Em caso de erro, retorna política padrão
      return this.getDefaultPolicy(context);
    }
  }
  
  /**
   * Obtém o nome da política regional baseado no código da região
   * 
   * @param {string} regionCode Código da região
   * @returns {string} Nome da política
   */
  getRegionalPolicyName(regionCode) {
    switch (regionCode) {
      case 'EU':
        return 'EU Standard Password Policy';
      case 'BR':
        return 'BR Standard Password Policy';
      case 'AO':
        return 'AO Standard Password Policy';
      case 'US':
        return 'US Standard Password Policy';
      default:
        return 'EU Standard Password Policy'; // Mais restritiva como padrão
    }
  }
  
  /**
   * Aplica configurações regionais a uma política
   * 
   * @param {Object} policy Política base
   * @param {string} region Código da região
   * @returns {Object} Política com configurações regionais
   */
  applyRegionalSettings(policy, region) {
    if (!policy.regional_settings || !policy.regional_settings[region]) {
      return policy;
    }
    
    const regionalSettings = policy.regional_settings[region];
    const result = { ...policy };
    
    // Sobrescreve atributos com configurações regionais
    Object.keys(regionalSettings).forEach(key => {
      if (key !== 'region' && key !== 'compliance') {
        result[key] = regionalSettings[key];
      }
    });
    
    // Registra a região aplicada para referência
    result.applied_region = region;
    result.regional_compliance = regionalSettings.compliance || [];
    
    return result;
  }
  
  /**
   * Aplica configurações setoriais a uma política
   * 
   * @param {Object} policy Política base
   * @param {string} industry Setor
   * @returns {Object} Política com configurações setoriais
   */
  applyIndustrySettings(policy, industry) {
    if (!policy.industry_settings || !policy.industry_settings[industry]) {
      return policy;
    }
    
    const industrySettings = policy.industry_settings[industry];
    const result = { ...policy };
    
    // Sobrescreve atributos com configurações setoriais
    Object.keys(industrySettings).forEach(key => {
      if (key !== 'industry' && key !== 'compliance') {
        result[key] = industrySettings[key];
      }
    });
    
    // Registra o setor aplicado para referência
    result.applied_industry = industry;
    result.industry_compliance = industrySettings.compliance || [];
    
    return result;
  }
  
  /**
   * Obtém uma política padrão codificada
   * 
   * @param {Object} context Contexto para customização
   * @returns {Object} Política padrão
   */
  getDefaultPolicy(context = {}) {
    // Política base
    const policy = {
      complexity: 'medium',
      min_length: 8,
      max_length: 128,
      require_uppercase: true,
      require_lowercase: true,
      require_numbers: true,
      require_special_chars: false,
      special_chars_set: '@#$%^&*()_+-=[]{}|;:,.<>?',
      max_age_days: 90,
      history_count: 5,
      prevent_common_passwords: true,
      prevent_username_in_password: true,
      prevent_user_info_in_password: false,
      max_attempts: 5,
      lockout_duration_minutes: 30
    };
    
    // Adaptações regionais
    if (context.region) {
      switch (context.region) {
        case 'EU': // União Europeia (GDPR, eIDAS)
          policy.min_length = 12;
          policy.require_special_chars = true;
          policy.history_count = 10;
          policy.max_age_days = 90;
          break;
          
        case 'BR': // Brasil (LGPD)
          policy.min_length = 10;
          policy.require_special_chars = true;
          policy.history_count = 8;
          policy.max_age_days = 90;
          break;
          
        case 'AO': // Angola (PNDSB)
          policy.min_length = 8;
          policy.max_age_days = 120;
          policy.history_count = 5;
          break;
          
        case 'US': // EUA (NIST 800-63B)
          policy.min_length = 12;
          policy.max_age_days = null; // NIST não recomenda expiração regular
          policy.require_special_chars = true;
          policy.history_count = 10;
          break;
      }
    }
    
    // Adaptações setoriais
    if (context.industry) {
      switch (context.industry) {
        case 'finance': // Setor financeiro
          policy.complexity = 'high';
          policy.min_length = Math.max(policy.min_length, 12);
          policy.require_special_chars = true;
          policy.history_count = Math.max(policy.history_count, 10);
          policy.prevent_user_info_in_password = true;
          break;
          
        case 'healthcare': // Setor de saúde
          policy.complexity = 'high';
          policy.min_length = Math.max(policy.min_length, 12);
          policy.require_special_chars = true;
          policy.prevent_user_info_in_password = true;
          
          // Adaptações específicas de região para saúde
          if (context.region === 'US') {
            // HIPAA
            policy.max_age_days = 60; // Mais restritivo para saúde nos EUA
          }
          break;
          
        case 'government': // Setor governamental
          policy.complexity = 'very_high';
          policy.min_length = Math.max(policy.min_length, 14);
          policy.require_special_chars = true;
          policy.history_count = Math.max(policy.history_count, 24);
          policy.prevent_user_info_in_password = true;
          break;
      }
    }
    
    return policy;
  }
  
  /**
   * Valida uma senha contra uma política
   * 
   * @param {string} password Senha a validar
   * @param {Object} policy Política de senha
   * @param {Object} userData Dados do usuário para verificar informações pessoais
   * @returns {Object} Resultado da validação
   */
  validatePassword(password, policy, userData = {}) {
    const result = {
      valid: true,
      reasons: []
    };
    
    // Verifica comprimento
    if (password.length < policy.min_length) {
      result.valid = false;
      result.reasons.push({
        code: 'TOO_SHORT',
        message: `A senha deve ter pelo menos ${policy.min_length} caracteres`
      });
    }
    
    if (policy.max_length && password.length > policy.max_length) {
      result.valid = false;
      result.reasons.push({
        code: 'TOO_LONG',
        message: `A senha não pode ter mais que ${policy.max_length} caracteres`
      });
    }
    
    // Verifica caracteres
    if (policy.require_uppercase && !/[A-Z]/.test(password)) {
      result.valid = false;
      result.reasons.push({
        code: 'NO_UPPERCASE',
        message: 'A senha deve conter pelo menos uma letra maiúscula'
      });
    }
    
    if (policy.require_lowercase && !/[a-z]/.test(password)) {
      result.valid = false;
      result.reasons.push({
        code: 'NO_LOWERCASE',
        message: 'A senha deve conter pelo menos uma letra minúscula'
      });
    }
    
    if (policy.require_numbers && !/[0-9]/.test(password)) {
      result.valid = false;
      result.reasons.push({
        code: 'NO_NUMBERS',
        message: 'A senha deve conter pelo menos um número'
      });
    }
    
    if (policy.require_special_chars) {
      const specialCharsRegex = new RegExp(`[${policy.special_chars_set.replace(/[-[\]{}()*+?.,\\^$|#\s]/g, '\\$&')}]`);
      if (!specialCharsRegex.test(password)) {
        result.valid = false;
        result.reasons.push({
          code: 'NO_SPECIAL_CHARS',
          message: 'A senha deve conter pelo menos um caractere especial'
        });
      }
    }
    
    // Verifica senhas comuns
    if (policy.prevent_common_passwords && this.isCommonPassword(password)) {
      result.valid = false;
      result.reasons.push({
        code: 'COMMON_PASSWORD',
        message: 'Esta senha é muito comum e não é segura'
      });
    }
    
    // Verifica se contém o nome de usuário
    if (policy.prevent_username_in_password && 
        userData.username && 
        password.toLowerCase().includes(userData.username.toLowerCase())) {
      result.valid = false;
      result.reasons.push({
        code: 'CONTAINS_USERNAME',
        message: 'A senha não pode conter seu nome de usuário'
      });
    }
    
    // Verifica se contém informações pessoais
    if (policy.prevent_user_info_in_password) {
      const personalInfo = [];
      
      if (userData.first_name && userData.first_name.length > 2) {
        personalInfo.push(userData.first_name);
      }
      
      if (userData.last_name && userData.last_name.length > 2) {
        personalInfo.push(userData.last_name);
      }
      
      if (userData.email) {
        const emailParts = userData.email.split('@')[0].split('.');
        emailParts.forEach(part => {
          if (part.length > 2) {
            personalInfo.push(part);
          }
        });
      }
      
      // Verifica cada informação pessoal
      for (const info of personalInfo) {
        if (password.toLowerCase().includes(info.toLowerCase())) {
          result.valid = false;
          result.reasons.push({
            code: 'CONTAINS_PERSONAL_INFO',
            message: 'A senha não pode conter suas informações pessoais'
          });
          break;
        }
      }
    }
    
    // Se não houver razões de falha, a senha é válida
    return result;
  }
  
  /**
   * Verifica se uma senha está na lista de senhas comuns
   * 
   * @param {string} password Senha a verificar
   * @returns {boolean} Verdadeiro se for uma senha comum
   */
  isCommonPassword(password) {
    return this.commonPasswordCache.has(password.toLowerCase());
  }
  
  /**
   * Verifica o histórico de senhas de um usuário
   * 
   * @param {string} userId ID do usuário
   * @param {string} tenantId ID do tenant
   * @param {string} password Nova senha proposta
   * @param {number} historyCount Quantidade de senhas no histórico a verificar
   * @returns {Promise<boolean>} Verdadeiro se a senha não estiver no histórico
   */
  async checkPasswordHistory(userId, tenantId, password, historyCount) {
    try {
      // Obtém o ID da credencial de senha do usuário
      const credQuery = `
        SELECT pc.credential_id 
        FROM iam.password_credentials pc
        JOIN iam.user_auth_methods uam ON pc.auth_method_id = uam.auth_method_id
        WHERE uam.user_id = $1
        AND uam.tenant_id = $2
        AND uam.status = 'active'
        AND uam.plugin_id = 'password-auth'
        LIMIT 1
      `;
      
      const credResult = await this.db.query(credQuery, [userId, tenantId]);
      
      if (credResult.rows.length === 0) {
        // Se o usuário não tem credencial, não há histórico
        return true;
      }
      
      const credentialId = credResult.rows[0].credential_id;
      
      // Obtém as senhas do histórico
      // Nota: Em uma implementação real, a verificação seria feita no lado do servidor
      // usando funções de hash específicas de cada algoritmo
      const historyQuery = `
        SELECT password_hash, password_salt, algorithm
        FROM iam.password_history
        WHERE user_id = $1
        AND tenant_id = $2
        AND credential_id = $3
        ORDER BY changed_at DESC
        LIMIT $4
      `;
      
      const historyResult = await this.db.query(historyQuery, [
        userId, tenantId, credentialId, historyCount || 5
      ]);
      
      // Em uma implementação real, verificaríamos o hash da senha contra cada hash do histórico
      // Aqui, como exemplo, retornamos sempre verdadeiro
      
      // const passwordService = new PasswordService();
      // for (const historyItem of historyResult.rows) {
      //   const match = await passwordService.verifyPassword(
      //     password,
      //     historyItem.password_hash,
      //     historyItem.algorithm,
      //     { salt: historyItem.password_salt }
      //   );
      //   
      //   if (match) {
      //     return false; // Senha encontrada no histórico
      //   }
      // }
      
      return true; // Senha não está no histórico
      
    } catch (error) {
      logger.error(`[PolicyService] Erro ao verificar histórico de senhas: ${error.message}`);
      return true; // Em caso de erro, permite a senha
    }
  }
  
  /**
   * Registra uma tentativa malsucedida de autenticação
   * 
   * @param {string} tenantId ID do tenant
   * @param {string} userId ID do usuário (opcional)
   * @param {string} username Nome de usuário (opcional)
   * @param {string} clientIp IP do cliente
   * @param {string} userAgent Agente do usuário
   * @param {string} reason Motivo da falha
   * @param {Object} context Contexto adicional
   * @returns {Promise<void>}
   */
  async recordFailedAttempt(tenantId, userId, username, clientIp, userAgent, reason, context = {}) {
    try {
      // Insere no banco de dados
      const query = `
        INSERT INTO iam.password_auth_attempts (
          tenant_id, user_id, username, status, client_ip,
          user_agent, reason, region, geolocation, session_id,
          correlation_id
        ) VALUES (
          $1, $2, $3, 'failure', $4, $5, $6, $7, $8, $9, $10
        )
      `;
      
      await this.db.query(query, [
        tenantId,
        userId,
        username,
        clientIp,
        userAgent,
        reason,
        context.region || null,
        context.geolocation ? JSON.stringify(context.geolocation) : null,
        context.sessionId || null,
        context.correlationId || null
      ]);
      
      // Atualiza cache de bloqueio em memória
      if (userId) {
        const key = `${tenantId}:${userId}`;
        const attempts = this.lockoutCache.get(key) || 0;
        this.lockoutCache.set(key, attempts + 1);
      }
      
      // Se o IP está fornecido, rastreie também por IP
      if (clientIp) {
        const ipKey = `ip:${clientIp}`;
        const ipAttempts = this.lockoutCache.get(ipKey) || 0;
        this.lockoutCache.set(ipKey, ipAttempts + 1);
      }
      
    } catch (error) {
      logger.error(`[PolicyService] Erro ao registrar tentativa malsucedida: ${error.message}`);
    }
  }
  
  /**
   * Verifica se uma conta está bloqueada
   * 
   * @param {string} tenantId ID do tenant
   * @param {string} userId ID do usuário
   * @param {string} clientIp IP do cliente (opcional)
   * @returns {Promise<Object>} Status de bloqueio
   */
  async checkAccountLockout(tenantId, userId, clientIp) {
    try {
      // Obtém a política de senha aplicável
      const policy = await this.getPasswordPolicy(tenantId, userId);
      
      // Se o bloqueio não estiver ativado na política, retorna desbloqueado
      if (!policy.lockout_enabled) {
        return {
          isLocked: false,
          remainingMinutes: 0,
          reason: null
        };
      }
      
      const maxAttempts = policy.max_attempts || 5;
      const lockoutDuration = policy.lockout_duration_minutes || 30;
      
      // Verifica no banco de dados
      const query = `
        SELECT COUNT(*) as attempt_count, MAX(timestamp) as last_attempt
        FROM iam.password_auth_attempts
        WHERE tenant_id = $1
        AND user_id = $2
        AND status = 'failure'
        AND timestamp > NOW() - INTERVAL '${lockoutDuration} minutes'
      `;
      
      let params = [tenantId, userId];
      let whereClause = '';
      
      // Se o IP for fornecido, filtra também por IP
      if (clientIp) {
        whereClause = ' AND client_ip = $3';
        params.push(clientIp);
      }
      
      const result = await this.db.query(query + whereClause, params);
      
      if (result.rows.length > 0) {
        const attemptCount = parseInt(result.rows[0].attempt_count);
        const lastAttempt = result.rows[0].last_attempt;
        
        if (attemptCount >= maxAttempts) {
          // Calcula o tempo restante de bloqueio
          const lockoutEnd = new Date(new Date(lastAttempt).getTime() + lockoutDuration * 60000);
          const now = new Date();
          const remainingMs = Math.max(0, lockoutEnd.getTime() - now.getTime());
          const remainingMinutes = Math.ceil(remainingMs / 60000);
          
          return {
            isLocked: true,
            remainingMinutes,
            reason: 'Conta temporariamente bloqueada devido a múltiplas tentativas malsucedidas.'
          };
        }
      }
      
      // Também verifica o cache em memória
      const key = `${tenantId}:${userId}`;
      const cachedAttempts = this.lockoutCache.get(key) || 0;
      
      if (cachedAttempts >= maxAttempts) {
        return {
          isLocked: true,
          remainingMinutes: lockoutDuration,
          reason: 'Conta temporariamente bloqueada devido a múltiplas tentativas malsucedidas.'
        };
      }
      
      // Verifica bloqueio por IP se IP fornecido
      if (clientIp) {
        const ipKey = `ip:${clientIp}`;
        const ipAttempts = this.lockoutCache.get(ipKey) || 0;
        
        // Limite mais alto para bloqueio por IP para evitar falsos positivos
        if (ipAttempts >= maxAttempts * 3) {
          return {
            isLocked: true,
            remainingMinutes: lockoutDuration,
            reason: 'Bloqueado temporariamente devido a múltiplas tentativas malsucedidas deste IP.'
          };
        }
      }
      
      // Não está bloqueado
      return {
        isLocked: false,
        remainingMinutes: 0,
        reason: null
      };
      
    } catch (error) {
      logger.error(`[PolicyService] Erro ao verificar bloqueio de conta: ${error.message}`);
      
      // Em caso de erro, assume não bloqueado
      return {
        isLocked: false,
        remainingMinutes: 0,
        reason: null
      };
    }
  }
  
  /**
   * Libera o bloqueio de uma conta
   * 
   * @param {string} tenantId ID do tenant
   * @param {string} userId ID do usuário
   * @param {string} adminId ID do administrador que está liberando
   * @returns {Promise<boolean>} Sucesso da operação
   */
  async unlockAccount(tenantId, userId, adminId) {
    try {
      // Limpa o cache de bloqueio
      const key = `${tenantId}:${userId}`;
      this.lockoutCache.delete(key);
      
      // Insere registro de liberação no banco
      const query = `
        INSERT INTO iam.password_auth_attempts (
          tenant_id, user_id, status, reason, 
          created_by, metadata
        ) VALUES (
          $1, $2, 'unlocked', 'admin_unlock',
          $3, $4
        )
      `;
      
      await this.db.query(query, [
        tenantId,
        userId,
        adminId,
        JSON.stringify({ unlocked_by: adminId })
      ]);
      
      return true;
      
    } catch (error) {
      logger.error(`[PolicyService] Erro ao liberar conta: ${error.message}`);
      return false;
    }
  }
  
  /**
   * Obtém contagem de bloqueios ativos
   * 
   * @returns {number} Número de contas bloqueadas
   */
  getLockoutCount() {
    return this.lockoutCache.size;
  }
  
  /**
   * Verifica a saúde do serviço
   * 
   * @returns {Promise<Object>} Status de saúde
   */
  async checkHealth() {
    try {
      // Verifica conexão com banco de dados
      await this.db.query('SELECT 1');
      
      return {
        status: 'healthy',
        details: {
          commonPasswordsLoaded: this.commonPasswordCache.size,
          activeLockouts: this.lockoutCache.size
        },
        timestamp: new Date()
      };
    } catch (error) {
      return {
        status: 'unhealthy',
        error: error.message,
        timestamp: new Date()
      };
    }
  }
}

module.exports = PolicyService;
