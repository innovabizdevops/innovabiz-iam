/**
 * Password Authentication Provider
 * 
 * Este provedor implementa autenticação tradicional baseada em nome de usuário/senha
 * com suporte a políticas avançadas, adaptações regionais e segurança aprimorada.
 * 
 * @module password-auth
 * @author INNOVABIZ
 * @copyright 2025 INNOVABIZ
 * @license Proprietary
 */

const { v4: uuidv4 } = require('uuid');
const { AuthenticationProvider, AuthMethodCategory, AssuranceLevel } = require('@innovabiz/auth-framework');
const CryptoService = require('@innovabiz/auth-framework/services/crypto');
const SessionManager = require('@innovabiz/auth-framework/services/session');
const AuditService = require('@innovabiz/auth-framework/services/audit');
const { logger } = require('@innovabiz/auth-framework/utils/logging');

const PasswordService = require('./lib/password-service');
const PolicyService = require('./lib/policy-service');
const PasswordRepository = require('./lib/password-repository');
const RegionalAdapter = require('./lib/regional-adapter');

/**
 * Implementação do Provedor de Autenticação por Senha
 */
class PasswordAuthProvider extends AuthenticationProvider {
  /**
   * Identificador único para este provedor
   */
  get id() {
    return 'password-auth';
  }
  
  /**
   * Metadados do provedor
   */
  get metadata() {
    return {
      name: 'Password Authentication',
      description: 'Traditional username/password authentication',
      version: '1.0.0',
      category: AuthMethodCategory.KNOWLEDGE,
      assuranceLevel: AssuranceLevel.LOW,
      capabilities: {
        supportsPasswordless: false,
        supportsFederatedLogin: false,
        supportsCrossPlatform: true,
        supportsOfflineMode: true,
        supportsSilentAuthentication: false,
        requiresUserInteraction: true,
        isPhishingResistant: false,
        supportsBruteForceProtection: true,
        supportsComplexityRules: true,
        supportsHistoryPolicies: true,
        supportsAgeRules: true,
        supportsAdaptiveThrottling: true
      }
    };
  }
  
  /**
   * Inicializa o provedor com configurações
   * 
   * @param {Object} config Configurações do provedor
   * @returns {Promise<void>}
   */
  async initialize(config) {
    this.config = this.validateConfig(config);
    
    // Inicializa serviços
    this.cryptoService = new CryptoService();
    this.sessionManager = new SessionManager();
    this.auditService = new AuditService();
    
    // Inicializa serviços específicos de senha
    this.passwordRepository = new PasswordRepository();
    this.passwordService = new PasswordService(this.config.hashingAlgorithm, this.config.hashingOptions);
    this.policyService = new PolicyService();
    
    // Inicializa adaptador regional
    this.regionalAdapter = new RegionalAdapter(this.config.regionalSettings);
    
    logger.info(`[PasswordAuthProvider] Inicializado com configurações: algoritmo=${this.config.hashingAlgorithm}`);
  }
  
  /**
   * Valida as configurações do provedor
   * 
   * @param {Object} config Configurações a validar
   * @returns {Object} Configurações validadas
   */
  validateConfig(config) {
    // Valores padrão para parâmetros opcionais
    return {
      ...config,
      hashingAlgorithm: config.hashingAlgorithm || 'argon2id',
      hashingOptions: config.hashingOptions || {
        iterations: 12,
        memoryCost: 65536,
        timeCost: 3,
        parallelism: 1
      },
      lockoutOptions: config.lockoutOptions || {
        enabled: true,
        maxAttempts: 5,
        lockoutDurationMinutes: 30,
        progressiveThrottling: true
      },
      selfServiceEnabled: config.selfServiceEnabled !== undefined ? config.selfServiceEnabled : true,
      authUiTemplate: config.authUiTemplate || 'standard',
      regionalSettings: config.regionalSettings || {},
      industrySettings: config.industrySettings || {}
    };
  }
  
  /**
   * Inicia o processo de autenticação
   * 
   * @param {Object} context Contexto de autenticação
   * @returns {Promise<Object>} Desafio de autenticação
   */
  async startAuthentication(context) {
    logger.debug(`[PasswordAuthProvider] Iniciando autenticação para contexto: ${JSON.stringify(context)}`);
    
    try {
      // Cria um ID de sessão para esta tentativa de autenticação
      const sessionId = uuidv4();
      
      // Aplica adaptações regionais baseadas no contexto
      const adaptedContext = await this.regionalAdapter.adaptAuthenticationContext(context);
      
      // Verifica bloqueio da conta se username for fornecido
      if (context.username) {
        const user = await this.getUserByUsername(context.username, context.tenantId);
        
        if (user) {
          const lockoutCheck = await this.policyService.checkAccountLockout(
            context.tenantId,
            user.user_id,
            context.clientIp
          );
          
          if (lockoutCheck.isLocked) {
            // Audita tentativa em conta bloqueada
            await this.auditService.logAuthEvent({
              eventType: 'authentication:failed',
              providerId: this.id,
              userId: user.user_id,
              tenantId: context.tenantId,
              clientIp: context.clientIp,
              userAgent: context.userAgent,
              details: {
                reason: 'account_locked',
                remainingMinutes: lockoutCheck.remainingMinutes
              },
              success: false
            });
            
            // Retorna erro de conta bloqueada
            return {
              sessionId,
              challengeType: 'error',
              challenge: {
                errorCode: 'account_locked',
                message: `Conta temporariamente bloqueada. Tente novamente em ${lockoutCheck.remainingMinutes} minutos.`
              },
              expiresAt: new Date(Date.now() + 60000), // 1 minuto
              uiOptions: {
                title: 'Conta Bloqueada',
                message: `Conta temporariamente bloqueada devido a múltiplas tentativas malsucedidas. Tente novamente em ${lockoutCheck.remainingMinutes} minutos.`,
                uiExtension: 'password-auth-lockout-ui'
              }
            };
          }
        }
      }
      
      // Armazena desafio para verificação posterior
      await this.sessionManager.setAuthenticationChallenge(
        sessionId,
        '',  // Não há challenge específico para autenticação por senha
        null, // Usuário ainda não identificado
        {
          tenantId: context.tenantId,
          origin: context.origin,
          clientIp: context.clientIp,
          userAgent: context.userAgent,
          region: adaptedContext.region,
          contextId: context.contextId
        },
        300 // 5 minutos
      );
      
      // Audita início da autenticação
      await this.auditService.logAuthEvent({
        eventType: 'authentication:started',
        providerId: this.id,
        tenantId: context.tenantId,
        sessionId,
        clientIp: context.clientIp,
        userAgent: context.userAgent,
        success: true
      });
      
      // Retorna desafio de autenticação
      return {
        sessionId,
        challengeType: 'password.login',
        challenge: {
          usernameRequired: true,
          passwordRequired: true,
          additionalFields: this.getAdditionalFields(adaptedContext)
        },
        expiresAt: new Date(Date.now() + 300000), // 5 minutos
        uiOptions: {
          title: 'Login',
          message: 'Entre com suas credenciais',
          uiExtension: `password-auth-${this.config.authUiTemplate}-ui`
        }
      };
    } catch (error) {
      logger.error(`[PasswordAuthProvider] Erro ao iniciar autenticação: ${error.message}`);
      throw error;
    }
  }
  
  /**
   * Determina campos adicionais baseados no contexto regional
   * 
   * @param {Object} context Contexto adaptado
   * @returns {Array} Campos adicionais
   */
  getAdditionalFields(context) {
    const fields = [];
    
    // Campos específicos por região
    if (context.region === 'BR' && context.industryType === 'finance') {
      fields.push({
        name: 'cpf',
        label: 'CPF',
        type: 'text',
        required: true,
        mask: '999.999.999-99'
      });
    }
    
    if (context.region === 'AO' && context.offlineAuthEnabled) {
      fields.push({
        name: 'offlineToken',
        label: 'Código Offline (se disponível)',
        type: 'text',
        required: false
      });
    }
    
    return fields;
  }
  
  /**
   * Verifica a resposta de autenticação
   * 
   * @param {Object} response Resposta de autenticação
   * @param {Object} context Contexto de autenticação
   * @returns {Promise<Object>} Resultado da autenticação
   */
  async verifyResponse(response, context) {
    logger.debug(`[PasswordAuthProvider] Verificando resposta de autenticação para sessão ${response.sessionId}`);
    
    try {
      // Obtém dados do desafio armazenado
      const challengeData = await this.sessionManager.getAuthenticationChallenge(response.sessionId);
      
      if (!challengeData) {
        throw new Error('Desafio de autenticação não encontrado ou expirado');
      }
      
      // Valida os campos obrigatórios
      if (!response.response.username || !response.response.password) {
        throw new Error('Nome de usuário e senha são obrigatórios');
      }
      
      // Busca o usuário por nome de usuário
      const user = await this.getUserByUsername(
        response.response.username,
        context.tenantId || challengeData.tenantId
      );
      
      // Usuário não encontrado
      if (!user) {
        // Registra tentativa malsucedida
        await this.policyService.recordFailedAttempt(
          context.tenantId || challengeData.tenantId,
          null,
          response.response.username,
          context.clientIp || challengeData.clientIp,
          context.userAgent || challengeData.userAgent,
          'user_not_found'
        );
        
        // Audita falha
        await this.auditService.logAuthEvent({
          eventType: 'authentication:failed',
          providerId: this.id,
          tenantId: context.tenantId || challengeData.tenantId,
          sessionId: response.sessionId,
          clientIp: context.clientIp || challengeData.clientIp,
          userAgent: context.userAgent || challengeData.userAgent,
          details: {
            reason: 'user_not_found',
            username: response.response.username
          },
          success: false
        });
        
        // Retorna erro de credenciais inválidas (sem revelar que o usuário não existe)
        return {
          success: false,
          error: {
            code: 'invalid_credentials',
            message: 'Nome de usuário ou senha inválidos'
          }
        };
      }
      
      // Verifica se a conta está bloqueada
      const lockoutCheck = await this.policyService.checkAccountLockout(
        context.tenantId || challengeData.tenantId,
        user.user_id,
        context.clientIp || challengeData.clientIp
      );
      
      if (lockoutCheck.isLocked) {
        // Audita tentativa em conta bloqueada
        await this.auditService.logAuthEvent({
          eventType: 'authentication:failed',
          providerId: this.id,
          userId: user.user_id,
          tenantId: context.tenantId || challengeData.tenantId,
          sessionId: response.sessionId,
          clientIp: context.clientIp || challengeData.clientIp,
          userAgent: context.userAgent || challengeData.userAgent,
          details: {
            reason: 'account_locked',
            remainingMinutes: lockoutCheck.remainingMinutes
          },
          success: false
        });
        
        return {
          success: false,
          error: {
            code: 'account_locked',
            message: `Conta temporariamente bloqueada. Tente novamente em ${lockoutCheck.remainingMinutes} minutos.`
          }
        };
      }
      
      // Obtém a credencial de senha do usuário
      const credential = await this.passwordRepository.getPasswordCredential(user.user_id);
      
      if (!credential) {
        // Caso raro: usuário existe mas não tem credencial de senha
        logger.warn(`[PasswordAuthProvider] Usuário ${user.user_id} não tem credencial de senha`);
        
        await this.auditService.logAuthEvent({
          eventType: 'authentication:failed',
          providerId: this.id,
          userId: user.user_id,
          tenantId: context.tenantId || challengeData.tenantId,
          sessionId: response.sessionId,
          clientIp: context.clientIp || challengeData.clientIp,
          userAgent: context.userAgent || challengeData.userAgent,
          details: {
            reason: 'credential_not_found'
          },
          success: false
        });
        
        return {
          success: false,
          error: {
            code: 'invalid_credentials',
            message: 'Nome de usuário ou senha inválidos'
          }
        };
      }
      
      // Verifica se a senha está correta
      const passwordValid = await this.passwordService.verifyPassword(
        response.response.password,
        credential.password_hash,
        credential.algorithm,
        {
          salt: credential.password_salt,
          iterations: credential.iterations,
          memoryCost: credential.memory_cost,
          timeCost: credential.time_cost,
          parallelism: credential.parallelism
        }
      );
      
      if (!passwordValid) {
        // Senha inválida
        await this.policyService.recordFailedAttempt(
          context.tenantId || challengeData.tenantId,
          user.user_id,
          response.response.username,
          context.clientIp || challengeData.clientIp,
          context.userAgent || challengeData.userAgent,
          'invalid_password'
        );
        
        await this.auditService.logAuthEvent({
          eventType: 'authentication:failed',
          providerId: this.id,
          userId: user.user_id,
          tenantId: context.tenantId || challengeData.tenantId,
          sessionId: response.sessionId,
          clientIp: context.clientIp || challengeData.clientIp,
          userAgent: context.userAgent || challengeData.userAgent,
          details: {
            reason: 'invalid_password'
          },
          success: false
        });
        
        return {
          success: false,
          error: {
            code: 'invalid_credentials',
            message: 'Nome de usuário ou senha inválidos'
          }
        };
      }
      
      // Verifica se a senha precisa ser atualizada
      let passwordChangeRequired = false;
      let passwordExpirationDays = null;
      
      if (credential.must_change) {
        passwordChangeRequired = true;
      } else if (credential.expires_at && new Date(credential.expires_at) <= new Date()) {
        passwordChangeRequired = true;
      } else if (credential.last_changed_at) {
        const policy = await this.policyService.getPasswordPolicy(context.tenantId, user.user_id);
        if (policy && policy.max_age_days) {
          const lastChanged = new Date(credential.last_changed_at);
          const expirationDate = new Date(lastChanged);
          expirationDate.setDate(expirationDate.getDate() + policy.max_age_days);
          
          if (expirationDate <= new Date()) {
            passwordChangeRequired = true;
          } else {
            // Calcula dias restantes para expiração
            const daysRemaining = Math.ceil((expirationDate - new Date()) / (1000 * 60 * 60 * 24));
            if (daysRemaining <= 7) { // Avisa se faltam 7 dias ou menos
              passwordExpirationDays = daysRemaining;
            }
          }
        }
      }
      
      // Atualiza data de último uso
      await this.passwordRepository.updateLastUsed(credential.credential_id);
      
      // Audita autenticação bem-sucedida
      await this.auditService.logAuthEvent({
        eventType: 'authentication:completed',
        providerId: this.id,
        userId: user.user_id,
        tenantId: context.tenantId || challengeData.tenantId,
        sessionId: response.sessionId,
        clientIp: context.clientIp || challengeData.clientIp,
        userAgent: context.userAgent || challengeData.userAgent,
        details: {
          passwordChangeRequired,
          passwordExpirationDays
        },
        success: true
      });
      
      // Retorna resultado da autenticação
      return {
        success: true,
        userId: user.user_id,
        authTime: new Date(),
        expiresAt: new Date(Date.now() + 3600 * 1000), // 1 hora
        amr: ['pwd'],
        acr: 'urn:innovabiz:ac:classes:password',
        sessionId: response.sessionId,
        identityAttributes: {
          username: user.username,
          displayName: user.display_name || user.username,
          email: user.email
        },
        stepUp: passwordChangeRequired ? {
          required: true,
          methods: ['password-change'],
          reason: 'password_expired'
        } : undefined,
        notifications: passwordExpirationDays ? [{
          type: 'warning',
          message: `Sua senha expirará em ${passwordExpirationDays} dias.`
        }] : undefined
      };
    } catch (error) {
      logger.error(`[PasswordAuthProvider] Erro na verificação de autenticação: ${error.message}`);
      
      // Audita erro na autenticação
      await this.auditService.logAuthEvent({
        eventType: 'authentication:failed',
        providerId: this.id,
        tenantId: context.tenantId,
        sessionId: response.sessionId,
        clientIp: context.clientIp,
        userAgent: context.userAgent,
        success: false,
        error: error.message
      });
      
      return {
        success: false,
        error: {
          code: 'auth_verification_failed',
          message: 'Falha na verificação de autenticação',
          details: error.message
        }
      };
    }
  }
  
  /**
   * Busca um usuário pelo nome de usuário
   * 
   * @param {string} username Nome de usuário
   * @param {string} tenantId ID do tenant
   * @returns {Promise<Object>} Usuário encontrado ou null
   */
  async getUserByUsername(username, tenantId) {
    // Esta função seria implementada para consultar o banco de dados
    // Em uma implementação real, isso consultaria a tabela iam.users
    
    // Mock simplificado para fins de demonstração
    return await this.passwordRepository.getUserByUsername(username, tenantId);
  }
  
  /**
   * Cancela uma autenticação em andamento
   * 
   * @param {string} sessionId ID da sessão de autenticação a cancelar
   * @returns {Promise<void>}
   */
  async cancelAuthentication(sessionId) {
    logger.debug(`[PasswordAuthProvider] Cancelando autenticação para sessão ${sessionId}`);
    
    // Remove desafio da sessão
    await this.sessionManager.removeAuthenticationChallenge(sessionId);
    
    // Audita cancelamento
    await this.auditService.logAuthEvent({
      eventType: 'authentication:cancelled',
      providerId: this.id,
      sessionId,
      success: true
    });
  }
  
  /**
   * Verifica se este provedor suporta cadastramento
   * 
   * @returns {boolean} Se cadastramento é suportado
   */
  supportsEnrollment() {
    return true;
  }
  
  /**
   * Inicia o processo de cadastramento
   * 
   * @param {string} userId Usuário a cadastrar
   * @param {Object} context Contexto de cadastramento
   * @returns {Promise<Object>} Desafio de cadastramento
   */
  async startEnrollment(userId, context) {
    logger.debug(`[PasswordAuthProvider] Iniciando cadastramento para usuário ${userId}`);
    
    // Implementação do cadastramento de senha
    // ... (código omitido para brevidade) ...
    
    // Placeholder simplificado
    return {
      sessionId: uuidv4(),
      challengeType: 'password.enroll',
      challenge: {
        passwordRequirements: {
          minLength: 8,
          requireUppercase: true,
          requireLowercase: true,
          requireNumbers: true,
          requireSpecial: true
        }
      },
      expiresAt: new Date(Date.now() + 300000) // 5 minutos
    };
  }
  
  /**
   * Completa o processo de cadastramento
   * 
   * @param {Object} response Resposta de cadastramento
   * @param {Object} context Contexto de cadastramento
   * @returns {Promise<Object>} Resultado do cadastramento
   */
  async completeEnrollment(response, context) {
    // Implementação da conclusão do cadastramento de senha
    // ... (código omitido para brevidade) ...
    
    // Placeholder simplificado
    return {
      success: true,
      userId: context.userId,
      credentialId: uuidv4()
    };
  }
  
  /**
   * Verifica a saúde do provedor
   * 
   * @returns {Promise<Object>} Status de saúde
   */
  async checkHealth() {
    try {
      // Verifica conexão com repositório
      const repoHealth = await this.passwordRepository.checkHealth();
      
      // Verifica gerenciador de sessão
      const sessionHealth = await this.sessionManager.checkHealth();
      
      const isHealthy = repoHealth.status === 'healthy' && 
                       sessionHealth.status === 'healthy';
      
      return {
        status: isHealthy ? 'healthy' : 'unhealthy',
        components: {
          repository: repoHealth,
          sessionManager: sessionHealth
        },
        timestamp: new Date()
      };
    } catch (error) {
      logger.error(`[PasswordAuthProvider] Erro na verificação de saúde: ${error.message}`);
      
      return {
        status: 'unhealthy',
        error: error.message,
        timestamp: new Date()
      };
    }
  }
  
  /**
   * Obtém métricas do provedor
   * 
   * @returns {Object} Métricas de uso e desempenho
   */
  getMetrics() {
    return {
      activeEnrollments: this.sessionManager.getActiveEnrollmentCount(this.id),
      activeAuthentications: this.sessionManager.getActiveAuthenticationCount(this.id),
      averageAuthenticationTime: this.auditService.getAverageAuthenticationTime(this.id),
      successRate: this.auditService.getAuthenticationSuccessRate(this.id),
      enrollmentCount: this.auditService.getEnrollmentCount(this.id),
      failureCount: this.auditService.getAuthenticationFailureCount(this.id),
      lockoutCount: this.policyService.getLockoutCount(),
      timestamp: new Date()
    };
  }
  
  /**
   * Desliga o provedor
   * 
   * @returns {Promise<void>}
   */
  async shutdown() {
    logger.info('[PasswordAuthProvider] Desligando');
    
    // Limpa recursos
    await this.passwordRepository.close();
  }
}

/**
 * Registra provedor no framework de autenticação
 * 
 * @param {Object} registry O registro de provedores
 */
function register(registry) {
  registry.registerProvider(new PasswordAuthProvider());
}

module.exports = {
  PasswordAuthProvider,
  register
};
