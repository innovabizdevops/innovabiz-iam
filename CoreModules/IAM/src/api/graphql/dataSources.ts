/**
 * @file dataSources.ts
 * @description Data Sources para Apollo GraphQL - Conecta com os serviços backend
 * @author INNOVABIZ Development Team
 * @copyright 2025 INNOVABIZ
 * @version 1.0.0
 */

import { DataSource, DataSourceConfig } from 'apollo-datasource';
import { UserService } from '../../services/user/user.service';
import { TenantService } from '../../services/tenant/tenant.service';
import { AuthService } from '../../services/auth/auth.service';
import { SessionService } from '../../services/session/session.service';
import { MFAService } from '../../services/mfa/mfa.service';
import { RiskService } from '../../services/risk/risk.service';
import { logger } from '../../utils/logger';

// Interface para context da aplicação
export interface IContext {
  user?: any;
  ip?: string;
}

// Classe base para DataSources
abstract class BaseDataSource extends DataSource {
  context: IContext = {};

  initialize(config: DataSourceConfig<IContext>) {
    this.context = config.context;
  }

  protected handleError(error: any, operation: string): never {
    logger.error(`Erro em ${operation}`, { 
      error, 
      userId: this.context.user?.id,
      tenantId: this.context.user?.tenantId 
    });
    throw error;
  }
}

/**
 * DataSource para operações relacionadas a usuários
 */
export class UserAPI extends BaseDataSource {
  private userService: UserService;

  constructor(userService: UserService) {
    super();
    this.userService = userService;
  }

  async getUserById(id: string) {
    try {
      return await this.userService.findById(id);
    } catch (error) {
      return this.handleError(error, 'getUserById');
    }
  }

  async getUsers(filter?: any) {
    try {
      return await this.userService.findAll(filter);
    } catch (error) {
      return this.handleError(error, 'getUsers');
    }
  }

  async createUser(userData: any) {
    try {
      return await this.userService.create({
        ...userData,
        createdBy: this.context.user?.id
      });
    } catch (error) {
      return this.handleError(error, 'createUser');
    }
  }

  async updateUser(id: string, userData: any) {
    try {
      return await this.userService.update(id, {
        ...userData,
        updatedBy: this.context.user?.id
      });
    } catch (error) {
      return this.handleError(error, 'updateUser');
    }
  }

  async deleteUser(id: string) {
    try {
      return await this.userService.delete(id, this.context.user?.id);
    } catch (error) {
      return this.handleError(error, 'deleteUser');
    }
  }

  async getUserPermissions(userId: string, tenantId: string) {
    try {
      return await this.userService.getUserPermissions(userId, tenantId);
    } catch (error) {
      return this.handleError(error, 'getUserPermissions');
    }
  }
}

/**
 * DataSource para operações relacionadas a tenants
 */
export class TenantAPI extends BaseDataSource {
  private tenantService: TenantService;

  constructor(tenantService: TenantService) {
    super();
    this.tenantService = tenantService;
  }

  async getTenantById(id: string) {
    try {
      return await this.tenantService.findById(id);
    } catch (error) {
      return this.handleError(error, 'getTenantById');
    }
  }

  async getTenants(filter?: any) {
    try {
      return await this.tenantService.findAll(filter);
    } catch (error) {
      return this.handleError(error, 'getTenants');
    }
  }

  async createTenant(tenantData: any) {
    try {
      return await this.tenantService.create({
        ...tenantData,
        createdBy: this.context.user?.id
      });
    } catch (error) {
      return this.handleError(error, 'createTenant');
    }
  }

  async updateTenant(id: string, tenantData: any) {
    try {
      return await this.tenantService.update(id, {
        ...tenantData,
        updatedBy: this.context.user?.id
      });
    } catch (error) {
      return this.handleError(error, 'updateTenant');
    }
  }

  async deleteTenant(id: string) {
    try {
      return await this.tenantService.delete(id);
    } catch (error) {
      return this.handleError(error, 'deleteTenant');
    }
  }
}

/**
 * DataSource para operações relacionadas a autenticação
 */
export class AuthAPI extends BaseDataSource {
  private authService: AuthService;

  constructor(authService: AuthService) {
    super();
    this.authService = authService;
  }

  async login(username: string, password: string, tenantId: string) {
    try {
      return await this.authService.login(username, password, tenantId, this.context.ip);
    } catch (error) {
      return this.handleError(error, 'login');
    }
  }

  async refreshToken(refreshToken: string) {
    try {
      return await this.authService.refreshToken(refreshToken, this.context.ip);
    } catch (error) {
      return this.handleError(error, 'refreshToken');
    }
  }

  async changePassword(userId: string, currentPassword: string, newPassword: string) {
    try {
      return await this.authService.changePassword(userId, currentPassword, newPassword);
    } catch (error) {
      return this.handleError(error, 'changePassword');
    }
  }

  async requestPasswordReset(email: string, tenantId: string) {
    try {
      return await this.authService.requestPasswordReset(email, tenantId);
    } catch (error) {
      return this.handleError(error, 'requestPasswordReset');
    }
  }

  async confirmPasswordReset(token: string, newPassword: string) {
    try {
      return await this.authService.confirmPasswordReset(token, newPassword);
    } catch (error) {
      return this.handleError(error, 'confirmPasswordReset');
    }
  }

  async getUserAuthMethods(userId: string) {
    try {
      return await this.authService.getUserAuthMethods(userId);
    } catch (error) {
      return this.handleError(error, 'getUserAuthMethods');
    }
  }
}

/**
 * DataSource para operações relacionadas a sessões
 */
export class SessionAPI extends BaseDataSource {
  private sessionService: SessionService;

  constructor(sessionService: SessionService) {
    super();
    this.sessionService = sessionService;
  }

  async getSessionById(id: string) {
    try {
      return await this.sessionService.findById(id);
    } catch (error) {
      return this.handleError(error, 'getSessionById');
    }
  }

  async getSessions(filter?: any) {
    try {
      return await this.sessionService.findAll(filter);
    } catch (error) {
      return this.handleError(error, 'getSessions');
    }
  }

  async getUserActiveSessions(userId: string) {
    try {
      return await this.sessionService.findActiveByUserId(userId);
    } catch (error) {
      return this.handleError(error, 'getUserActiveSessions');
    }
  }

  async invalidateSession(id: string) {
    try {
      return await this.sessionService.invalidate(id);
    } catch (error) {
      return this.handleError(error, 'invalidateSession');
    }
  }

  async invalidateAllUserSessions(userId: string, exceptSessionId?: string) {
    try {
      return await this.sessionService.invalidateAllForUser(userId, exceptSessionId);
    } catch (error) {
      return this.handleError(error, 'invalidateAllUserSessions');
    }
  }
}

/**
 * DataSource para operações relacionadas a MFA
 */
export class MFAAPI extends BaseDataSource {
  private mfaService: MFAService;

  constructor(mfaService: MFAService) {
    super();
    this.mfaService = mfaService;
  }

  async getUserMFAMethods(userId: string) {
    try {
      return await this.mfaService.getUserMethods(userId);
    } catch (error) {
      return this.handleError(error, 'getUserMFAMethods');
    }
  }

  async enrollMFAMethod(userId: string, type: string, options: any) {
    try {
      return await this.mfaService.enrollMethod(userId, type, options);
    } catch (error) {
      return this.handleError(error, 'enrollMFAMethod');
    }
  }

  async verifyMFA(userId: string, type: string, code: string) {
    try {
      return await this.mfaService.verify(userId, type, code);
    } catch (error) {
      return this.handleError(error, 'verifyMFA');
    }
  }

  async disableMFAMethod(userId: string, type: string) {
    try {
      return await this.mfaService.disableMethod(userId, type);
    } catch (error) {
      return this.handleError(error, 'disableMFAMethod');
    }
  }
}

/**
 * DataSource para operações relacionadas a análise de risco
 */
export class RiskAPI extends BaseDataSource {
  private riskService: RiskService;

  constructor(riskService: RiskService) {
    super();
    this.riskService = riskService;
  }

  async getUserRiskProfile(userId: string, tenantId: string) {
    try {
      return await this.riskService.getUserRiskProfile(userId, tenantId);
    } catch (error) {
      return this.handleError(error, 'getUserRiskProfile');
    }
  }

  async assessRisk(assessment: any) {
    try {
      return await this.riskService.assessRisk(assessment);
    } catch (error) {
      return this.handleError(error, 'assessRisk');
    }
  }

  async reportEvent(userId: string, eventType: string, metadata: any) {
    try {
      return await this.riskService.reportEvent(userId, eventType, metadata);
    } catch (error) {
      return this.handleError(error, 'reportEvent');
    }
  }
}

// Função para criar e retornar todas as DataSources
export function createDataSources() {
  // Em uma implementação real, os serviços seriam injetados
  // via contêiner DI ou um factory pattern
  const userService = new UserService();
  const tenantService = new TenantService();
  const authService = new AuthService();
  const sessionService = new SessionService();
  const mfaService = new MFAService();
  const riskService = new RiskService();

  return {
    userAPI: new UserAPI(userService),
    tenantAPI: new TenantAPI(tenantService),
    authAPI: new AuthAPI(authService),
    sessionAPI: new SessionAPI(sessionService),
    mfaAPI: new MFAAPI(mfaService),
    riskAPI: new RiskAPI(riskService)
  };
}