/**
 * Serviço de Recuperação de Senha
 * 
 * Este módulo gerencia os fluxos de recuperação e redefinição de senha,
 * incluindo geração de tokens, validação e comunicação com usuários.
 * 
 * @module recovery-service
 * @author INNOVABIZ
 * @copyright 2025 INNOVABIZ
 * @license Proprietary
 */

const crypto = require('crypto');
const { v4: uuidv4 } = require('uuid');
const { logger } = require('@innovabiz/auth-framework/utils/logging');
const DatabaseService = require('@innovabiz/auth-framework/services/database');
const NotificationService = require('@innovabiz/auth-framework/services/notification');
const RegionalAdapter = require('./regional-adapter');

/**
 * Serviço de recuperação de senhas
 */
class RecoveryService {
  /**
   * Cria uma nova instância do serviço de recuperação
   */
  constructor() {
    this.db = new DatabaseService();
    this.notification = new NotificationService();
    this.regionalAdapter = new RegionalAdapter();
    
    logger.info('[RecoveryService] Inicializado');
  }
  
  /**
   * Inicia o processo de recuperação de senha
   * 
   * @param {string} tenantId ID do tenant
   * @param {string} identifier Identificador do usuário (email ou username)
   * @param {Object} context Contexto da requisição
   * @returns {Promise<Object>} Resultado da operação
   */
  async initiateRecovery(tenantId, identifier, context = {}) {
    try {
      // Busca o usuário com base no identificador
      const user = await this.findUser(tenantId, identifier);
      
      if (!user) {
        logger.info(`[RecoveryService] Tentativa de recuperação para identificador inexistente: ${identifier}`);
        return {
          success: false,
          message: 'Um email com instruções de recuperação será enviado se o endereço fornecido estiver associado a uma conta.',
          code: 'USER_NOT_FOUND'
        };
      }
      
      // Verifica se o usuário pode recuperar a senha
      if (user.status !== 'active') {
        logger.info(`[RecoveryService] Tentativa de recuperação para usuário não ativo: ${user.user_id}`);
        return {
          success: false,
          message: 'Um email com instruções de recuperação será enviado se o endereço fornecido estiver associado a uma conta.',
          code: 'USER_INACTIVE'
        };
      }
      
      // Obtém a configuração regional
      const region = context.region || await this.getUserRegion(user.user_id, tenantId);
      const regionConfig = this.regionalAdapter.getRegionConfig(region);
      
      // Gera token de recuperação
      const tokenData = await this.generateRecoveryToken(
        tenantId, 
        user.user_id, 
        regionConfig.passwordRecovery.tokenExpiryMinutes || 15,
        context
      );
      
      if (!tokenData.success) {
        return {
          success: false,
          message: 'Não foi possível iniciar o processo de recuperação de senha.',
          code: 'TOKEN_GENERATION_FAILED'
        };
      }
      
      // Envia notificação baseada em configuração regional
      const notificationResult = await this.sendRecoveryNotification(
        tenantId,
        user,
        tokenData.token,
        regionConfig,
        context
      );
      
      if (!notificationResult.success) {
        return {
          success: false,
          message: 'Falha ao enviar notificação de recuperação.',
          code: 'NOTIFICATION_FAILED'
        };
      }
      
      // Resposta de segurança - não confirma se o usuário existe
      return {
        success: true,
        message: 'Se o endereço fornecido estiver associado a uma conta, você receberá instruções para recuperação de senha.',
        code: 'RECOVERY_INITIATED'
      };
      
    } catch (error) {
      logger.error(`[RecoveryService] Erro ao iniciar recuperação: ${error.message}`);
      
      // Resposta de segurança - não revela se houve erro
      return {
        success: true,
        message: 'Se o endereço fornecido estiver associado a uma conta, você receberá instruções para recuperação de senha.',
        code: 'RECOVERY_INITIATED'
      };
    }
  }
  
  /**
   * Busca um usuário pelo identificador
   * 
   * @param {string} tenantId ID do tenant
   * @param {string} identifier Identificador (email ou username)
   * @returns {Promise<Object>} Dados do usuário
   */
  async findUser(tenantId, identifier) {
    try {
      // Primeiro tenta buscar por email
      const isEmail = identifier.includes('@');
      
      let query;
      if (isEmail) {
        query = `
          SELECT u.*, e.verified as email_verified
          FROM iam.users u
          LEFT JOIN iam.user_emails e ON u.user_id = e.user_id AND e.email = $2
          WHERE u.tenant_id = $1
          AND e.email = $2
          LIMIT 1
        `;
      } else {
        query = `
          SELECT u.*, true as email_verified
          FROM iam.users u
          WHERE u.tenant_id = $1
          AND u.username = $2
          LIMIT 1
        `;
      }
      
      const result = await this.db.query(query, [tenantId, identifier]);
      
      if (result.rows.length === 0) {
        return null;
      }
      
      return result.rows[0];
    } catch (error) {
      logger.error(`[RecoveryService] Erro ao buscar usuário: ${error.message}`);
      return null;
    }
  }
  
  /**
   * Obtém a região do usuário
   * 
   * @param {string} userId ID do usuário
   * @param {string} tenantId ID do tenant
   * @returns {Promise<string>} Código da região
   */
  async getUserRegion(userId, tenantId) {
    try {
      const query = `
        SELECT u.region, t.default_region
        FROM iam.users u
        JOIN iam.tenants t ON u.tenant_id = t.tenant_id
        WHERE u.user_id = $1
        AND u.tenant_id = $2
        LIMIT 1
      `;
      
      const result = await this.db.query(query, [userId, tenantId]);
      
      if (result.rows.length === 0) {
        return 'EU'; // Região padrão
      }
      
      // Usa a região do usuário ou a região padrão do tenant
      return result.rows[0].region || result.rows[0].default_region || 'EU';
    } catch (error) {
      logger.error(`[RecoveryService] Erro ao obter região do usuário: ${error.message}`);
      return 'EU'; // Região padrão em caso de erro
    }
  }
  
  /**
   * Gera um token de recuperação de senha
   * 
   * @param {string} tenantId ID do tenant
   * @param {string} userId ID do usuário
   * @param {number} expiryMinutes Minutos até expiração
   * @param {Object} context Contexto da requisição
   * @returns {Promise<Object>} Token e resultado
   */
  async generateRecoveryToken(tenantId, userId, expiryMinutes = 15, context = {}) {
    try {
      // Invalida tokens existentes
      await this.invalidateExistingTokens(tenantId, userId);
      
      // Gera token aleatório
      const token = crypto.randomBytes(32).toString('hex');
      const tokenHash = crypto.createHash('sha256').update(token).digest('hex');
      
      // Cria registro de token
      const query = `
        INSERT INTO iam.password_reset_tokens (
          token_id, tenant_id, user_id, token_hash, 
          expires_at, client_ip, user_agent, status
        ) VALUES (
          $1, $2, $3, $4, $5, $6, $7, 'active'
        )
      `;
      
      const tokenId = uuidv4();
      const expiresAt = new Date(Date.now() + (expiryMinutes * 60 * 1000));
      
      await this.db.query(query, [
        tokenId,
        tenantId,
        userId,
        tokenHash,
        expiresAt,
        context.clientIp || null,
        context.userAgent || null
      ]);
      
      return {
        success: true,
        token: token,
        tokenId: tokenId,
        expiresAt: expiresAt
      };
      
    } catch (error) {
      logger.error(`[RecoveryService] Erro ao gerar token: ${error.message}`);
      return {
        success: false,
        error: error.message
      };
    }
  }
  
  /**
   * Invalida tokens existentes para um usuário
   * 
   * @param {string} tenantId ID do tenant
   * @param {string} userId ID do usuário
   * @returns {Promise<boolean>} Sucesso da operação
   */
  async invalidateExistingTokens(tenantId, userId) {
    try {
      const query = `
        UPDATE iam.password_reset_tokens
        SET status = 'invalidated',
            updated_at = NOW()
        WHERE tenant_id = $1
        AND user_id = $2
        AND status = 'active'
      `;
      
      await this.db.query(query, [tenantId, userId]);
      return true;
    } catch (error) {
      logger.error(`[RecoveryService] Erro ao invalidar tokens: ${error.message}`);
      return false;
    }
  }
  
  /**
   * Envia notificação de recuperação de senha
   * 
   * @param {string} tenantId ID do tenant
   * @param {Object} user Dados do usuário
   * @param {string} token Token de recuperação
   * @param {Object} regionConfig Configuração regional
   * @param {Object} context Contexto da requisição
   * @returns {Promise<Object>} Resultado do envio
   */
  async sendRecoveryNotification(tenantId, user, token, regionConfig, context = {}) {
    try {
      // Obtém detalhes do tenant
      const tenantInfo = await this.getTenantInfo(tenantId);
      
      // Prepara dados para o template
      const templateData = {
        userName: user.full_name || user.username,
        resetLink: this.buildResetLink(tenantInfo, token, context),
        companyName: tenantInfo.display_name || tenantInfo.name,
        expiryHours: regionConfig.passwordRecovery.tokenExpiryMinutes / 60,
        supportEmail: tenantInfo.support_email,
        requestIp: context.clientIp || 'desconhecido',
        requestLocation: context.geolocation ? `${context.geolocation.city}, ${context.geolocation.country}` : 'desconhecido',
        requestTime: new Date().toLocaleString('pt-BR', { 
          timeZone: regionConfig.timeZone || 'UTC'
        })
      };
      
      // Envia email
      const emailData = {
        to: user.email,
        templateId: regionConfig.passwordRecovery.emailTemplate,
        data: templateData,
        metadata: {
          tenantId,
          userId: user.user_id,
          category: 'password_recovery',
          region: regionConfig.region
        }
      };
      
      await this.notification.sendEmail(emailData);
      
      // Se configurado, envia SMS também
      if (regionConfig.passwordRecovery.sendSms && user.phone_number) {
        const smsData = {
          to: user.phone_number,
          templateId: regionConfig.passwordRecovery.smsTemplate,
          data: {
            resetLink: this.buildShortResetLink(tenantInfo, token),
            expiryHours: regionConfig.passwordRecovery.tokenExpiryMinutes / 60
          },
          metadata: {
            tenantId,
            userId: user.user_id,
            category: 'password_recovery',
            region: regionConfig.region
          }
        };
        
        await this.notification.sendSms(smsData);
      }
      
      return { success: true };
      
    } catch (error) {
      logger.error(`[RecoveryService] Erro ao enviar notificação: ${error.message}`);
      return { success: false, error: error.message };
    }
  }
  
  /**
   * Obtém informações do tenant
   * 
   * @param {string} tenantId ID do tenant
   * @returns {Promise<Object>} Dados do tenant
   */
  async getTenantInfo(tenantId) {
    try {
      const query = `
        SELECT * FROM iam.tenants
        WHERE tenant_id = $1
        LIMIT 1
      `;
      
      const result = await this.db.query(query, [tenantId]);
      
      if (result.rows.length === 0) {
        throw new Error(`Tenant não encontrado: ${tenantId}`);
      }
      
      return result.rows[0];
    } catch (error) {
      logger.error(`[RecoveryService] Erro ao obter info do tenant: ${error.message}`);
      throw error;
    }
  }
  
  /**
   * Constrói link de redefinição de senha
   * 
   * @param {Object} tenant Dados do tenant
   * @param {string} token Token de redefinição
   * @param {Object} context Contexto da requisição
   * @returns {string} URL de redefinição
   */
  buildResetLink(tenant, token, context = {}) {
    // Base URL configurada para o tenant ou padrão
    const baseUrl = tenant.auth_settings?.password_reset_url || 
                    `https://${tenant.domain}/auth/reset-password`;
    
    // Constrói URL com token
    return `${baseUrl}?token=${token}`;
  }
  
  /**
   * Constrói link curto para SMS
   * 
   * @param {Object} tenant Dados do tenant
   * @param {string} token Token de redefinição
   * @returns {string} URL curto
   */
  buildShortResetLink(tenant, token) {
    // Em produção, usaria um serviço de URL curta
    return `https://${tenant.subdomain || 'app'}.innovabiz.com/r/${token.substring(0, 10)}`;
  }
  
  /**
   * Valida um token de redefinição de senha
   * 
   * @param {string} token Token a validar
   * @returns {Promise<Object>} Resultado da validação
   */
  async validateToken(token) {
    try {
      // Calcula hash do token
      const tokenHash = crypto.createHash('sha256').update(token).digest('hex');
      
      const query = `
        SELECT t.*, u.user_id, u.username, u.email, u.status as user_status,
               tn.tenant_id, tn.name as tenant_name, tn.status as tenant_status
        FROM iam.password_reset_tokens t
        JOIN iam.users u ON t.user_id = u.user_id
        JOIN iam.tenants tn ON t.tenant_id = tn.tenant_id
        WHERE t.token_hash = $1
        AND t.status = 'active'
        AND t.expires_at > NOW()
        LIMIT 1
      `;
      
      const result = await this.db.query(query, [tokenHash]);
      
      if (result.rows.length === 0) {
        return {
          valid: false,
          message: 'Token inválido ou expirado.',
          code: 'INVALID_TOKEN'
        };
      }
      
      const tokenData = result.rows[0];
      
      // Verifica status do tenant
      if (tokenData.tenant_status !== 'active') {
        return {
          valid: false,
          message: 'Tenant inativo.',
          code: 'INACTIVE_TENANT'
        };
      }
      
      // Verifica status do usuário
      if (tokenData.user_status !== 'active') {
        return {
          valid: false,
          message: 'Usuário inativo.',
          code: 'INACTIVE_USER'
        };
      }
      
      return {
        valid: true,
        userId: tokenData.user_id,
        tenantId: tokenData.tenant_id,
        tokenId: tokenData.token_id,
        username: tokenData.username,
        email: tokenData.email
      };
      
    } catch (error) {
      logger.error(`[RecoveryService] Erro ao validar token: ${error.message}`);
      return {
        valid: false,
        message: 'Erro ao validar token.',
        code: 'VALIDATION_ERROR'
      };
    }
  }
  
  /**
   * Completa a redefinição da senha
   * 
   * @param {string} token Token de redefinição
   * @param {string} newPassword Nova senha
   * @param {Object} context Contexto da requisição
   * @returns {Promise<Object>} Resultado da operação
   */
  async completeReset(token, newPassword, context = {}) {
    try {
      // Valida o token
      const validation = await this.validateToken(token);
      
      if (!validation.valid) {
        return {
          success: false,
          ...validation
        };
      }
      
      // Marca o token como usado
      await this.markTokenAsUsed(validation.tokenId);
      
      // Atualiza a senha do usuário
      const { tenantId, userId } = validation;
      
      // Em uma implementação real, usaria um serviço de senha para criar o hash
      const passwordChangeResult = await this.updateUserPassword(
        tenantId, 
        userId, 
        newPassword, 
        context
      );
      
      if (!passwordChangeResult.success) {
        return passwordChangeResult;
      }
      
      // Registra o evento de redefinição
      await this.recordResetEvent(tenantId, userId, validation.tokenId, context);
      
      // Notifica o usuário sobre a alteração
      await this.sendPasswordChangedNotification(tenantId, validation, context);
      
      return {
        success: true,
        message: 'Senha alterada com sucesso.',
        code: 'PASSWORD_RESET_SUCCESS'
      };
      
    } catch (error) {
      logger.error(`[RecoveryService] Erro ao completar redefinição: ${error.message}`);
      return {
        success: false,
        message: 'Erro ao redefinir senha.',
        code: 'RESET_ERROR'
      };
    }
  }
  
  /**
   * Marca um token como usado
   * 
   * @param {string} tokenId ID do token
   * @returns {Promise<boolean>} Sucesso da operação
   */
  async markTokenAsUsed(tokenId) {
    try {
      const query = `
        UPDATE iam.password_reset_tokens
        SET status = 'used',
            used_at = NOW(),
            updated_at = NOW()
        WHERE token_id = $1
      `;
      
      await this.db.query(query, [tokenId]);
      return true;
    } catch (error) {
      logger.error(`[RecoveryService] Erro ao marcar token como usado: ${error.message}`);
      return false;
    }
  }
  
  /**
   * Atualiza a senha do usuário
   * 
   * @param {string} tenantId ID do tenant
   * @param {string} userId ID do usuário
   * @param {string} newPassword Nova senha
   * @param {Object} context Contexto da requisição
   * @returns {Promise<Object>} Resultado da operação
   */
  async updateUserPassword(tenantId, userId, newPassword, context = {}) {
    try {
      // Em uma implementação real, importaria os serviços de senha e política
      // const passwordService = new PasswordService();
      // const policyService = new PolicyService();
      
      // Validaria a política de senha
      // const policy = await policyService.getPasswordPolicy(tenantId, userId, context);
      // const validationResult = policyService.validatePassword(newPassword, policy, userData);
      // if (!validationResult.valid) { return { success: false, ...validationResult }; }
      
      // Verificaria o histórico
      // const historyCheck = await policyService.checkPasswordHistory(userId, tenantId, newPassword, policy.history_count);
      // if (!historyCheck) { return { success: false, message: 'A senha não pode ser igual às senhas anteriores.' }; }
      
      // Geraria o hash da senha
      // const hashResult = await passwordService.hashPassword(newPassword, policy.preferred_algorithm);
      
      // Simulação simplificada de atualização
      const passwordHash = 'hash_simulado_da_senha'; // Na implementação real: hashResult.hash
      const algorithm = 'argon2id'; // Na implementação real: hashResult.algorithm
      
      // Atualizaria a credencial
      const credQuery = `
        UPDATE iam.password_credentials pc
        SET password_hash = $1,
            algorithm = $2,
            previous_password_hash = password_hash,
            last_changed_at = NOW(),
            updated_at = NOW()
        FROM iam.user_auth_methods uam
        WHERE pc.auth_method_id = uam.auth_method_id
        AND uam.user_id = $3
        AND uam.tenant_id = $4
        AND uam.plugin_id = 'password-auth'
        AND uam.status = 'active'
        RETURNING pc.credential_id
      `;
      
      const credResult = await this.db.query(credQuery, [
        passwordHash,
        algorithm,
        userId,
        tenantId
      ]);
      
      if (credResult.rows.length === 0) {
        logger.error(`[RecoveryService] Credencial não encontrada para usuário: ${userId}`);
        return {
          success: false,
          message: 'Credencial não encontrada.',
          code: 'CREDENTIAL_NOT_FOUND'
        };
      }
      
      // Adiciona ao histórico
      const historyQuery = `
        INSERT INTO iam.password_history (
          user_id, tenant_id, credential_id, password_hash,
          password_salt, algorithm
        ) VALUES (
          $1, $2, $3, $4, $5, $6
        )
      `;
      
      await this.db.query(historyQuery, [
        userId,
        tenantId,
        credResult.rows[0].credential_id,
        passwordHash,
        null, // salt, se aplicável
        algorithm
      ]);
      
      return {
        success: true,
        code: 'PASSWORD_UPDATED'
      };
      
    } catch (error) {
      logger.error(`[RecoveryService] Erro ao atualizar senha: ${error.message}`);
      return {
        success: false,
        message: 'Erro ao atualizar senha.',
        code: 'UPDATE_ERROR'
      };
    }
  }
  
  /**
   * Registra evento de redefinição de senha
   * 
   * @param {string} tenantId ID do tenant
   * @param {string} userId ID do usuário
   * @param {string} tokenId ID do token
   * @param {Object} context Contexto da requisição
   * @returns {Promise<boolean>} Sucesso da operação
   */
  async recordResetEvent(tenantId, userId, tokenId, context = {}) {
    try {
      // Em uma implementação real, usaria o EventService
      // const eventService = new EventService();
      // return eventService.logPasswordReset({
      //   tenant_id: tenantId,
      //   user_id: userId,
      //   status: 'success',
      //   client_ip: context.clientIp,
      //   user_agent: context.userAgent,
      //   metadata: { tokenId, source: 'recovery' }
      // });
      
      // Simulação simplificada
      logger.info(`[RecoveryService] Senha redefinida para usuário ${userId} (tenant: ${tenantId})`);
      return true;
    } catch (error) {
      logger.error(`[RecoveryService] Erro ao registrar evento: ${error.message}`);
      return false;
    }
  }
  
  /**
   * Envia notificação de senha alterada
   * 
   * @param {string} tenantId ID do tenant
   * @param {Object} userData Dados do usuário
   * @param {Object} context Contexto da requisição
   * @returns {Promise<boolean>} Sucesso da operação
   */
  async sendPasswordChangedNotification(tenantId, userData, context = {}) {
    try {
      // Obtém a configuração regional
      const region = context.region || await this.getUserRegion(userData.userId, tenantId);
      const regionConfig = this.regionalAdapter.getRegionConfig(region);
      
      // Obtém detalhes do tenant
      const tenantInfo = await this.getTenantInfo(tenantId);
      
      // Envia email de confirmação
      const emailData = {
        to: userData.email,
        templateId: regionConfig.passwordChanged.emailTemplate,
        data: {
          userName: userData.username,
          companyName: tenantInfo.display_name || tenantInfo.name,
          supportEmail: tenantInfo.support_email,
          requestIp: context.clientIp || 'desconhecido',
          requestLocation: context.geolocation ? `${context.geolocation.city}, ${context.geolocation.country}` : 'desconhecido',
          requestTime: new Date().toLocaleString('pt-BR', { 
            timeZone: regionConfig.timeZone || 'UTC'
          })
        },
        metadata: {
          tenantId,
          userId: userData.userId,
          category: 'password_changed',
          region: regionConfig.region
        }
      };
      
      await this.notification.sendEmail(emailData);
      
      return true;
    } catch (error) {
      logger.error(`[RecoveryService] Erro ao enviar notificação de senha alterada: ${error.message}`);
      return false;
    }
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

module.exports = RecoveryService;
