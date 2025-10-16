/**
 * Controlador REST para E-Commerce
 * 
 * Este controlador implementa endpoints REST para compatibilidade legada
 * relacionados à integração com o módulo de E-Commerce.
 */

import { Request, Response } from 'express';
import { BaseRestController } from './base-controller';
import { Logger } from '../../../../observability/logging/hook_logger';
import { Metrics } from '../../../../observability/metrics/hook_metrics';
import { Tracer } from '../../../../observability/tracing/hook_tracing';

/**
 * Controlador REST para operações de E-Commerce
 */
export class EcommerceRestController extends BaseRestController {
  private readonly ecommerceService: any;
  private readonly iamService: any;
  private readonly validationService: any;
  
  /**
   * Construtor do controlador de E-Commerce
   */
  constructor(
    logger: Logger,
    metrics: Metrics,
    tracer: Tracer,
    ecommerceService: any,
    iamService: any,
    validationService: any
  ) {
    super(logger, metrics, tracer);
    this.ecommerceService = ecommerceService;
    this.iamService = iamService;
    this.validationService = validationService;
  }
  
  /**
   * Verifica e valida o token de autenticação
   */
  private async validateAuthToken(req: Request, res: Response): Promise<{userId: string} | null> {
    const token = this.getAuthToken(req);
    
    if (!token) {
      this.sendErrorResponse(res, 401, 'Token de autenticação ausente', 'UNAUTHORIZED');
      return null;
    }
    
    try {
      const validation = await this.iamService.validateToken(token);
      
      if (!validation.valid) {
        this.sendErrorResponse(res, 401, 'Token de autenticação inválido', 'INVALID_TOKEN');
        return null;
      }
      
      return { userId: validation.userId };
    } catch (error) {
      this.handleError(error, req, res, 'token_validation');
      return null;
    }
  }
  
  /**
   * Autenticação com e-mail e senha para E-Commerce
   * 
   * POST /api/v1/ecommerce/auth/login
   */
  public async login(req: Request, res: Response): Promise<void> {
    const endpointName = 'ecommerce.login';
    const tenantId = this.getTenantId(req);
    const span = this.tracer.startSpan(endpointName);
    
    try {
      this.recordRequestStart(endpointName, tenantId);
      
      // Validar payload
      const { email, password } = req.body;
      
      if (!email || !password) {
        this.sendErrorResponse(res, 400, 'E-mail e senha são obrigatórios', 'MISSING_CREDENTIALS');
        return;
      }
      
      // Autenticar usuário
      const authResult = await this.ecommerceService.authenticate(
        email,
        password,
        tenantId
      );
      
      // Retornar resultado
      this.sendSuccessResponse(res, {
        userId: authResult.userId,
        accessToken: authResult.accessToken,
        refreshToken: authResult.refreshToken,
        expiresIn: authResult.expiresIn,
        roles: authResult.roles,
        tokenType: authResult.tokenType || 'Bearer',
        permissions: authResult.permissions
      });
      
      this.recordRequestResult(endpointName, tenantId, 200);
    } catch (error) {
      this.handleError(error, req, res, endpointName);
    } finally {
      span.end();
    }
  }
  
  /**
   * Login SSO usando token IAM
   * 
   * POST /api/v1/ecommerce/auth/sso-login
   */
  public async ssoLogin(req: Request, res: Response): Promise<void> {
    const endpointName = 'ecommerce.sso_login';
    const tenantId = this.getTenantId(req);
    const span = this.tracer.startSpan(endpointName);
    
    try {
      this.recordRequestStart(endpointName, tenantId);
      
      // Validar payload
      const { iamToken } = req.body;
      const shopId = req.body.shopId || req.query.shopId as string;
      
      if (!iamToken) {
        this.sendErrorResponse(res, 400, 'Token IAM é obrigatório', 'MISSING_TOKEN');
        return;
      }
      
      // Realizar login SSO
      const loginResult = await this.ecommerceService.ssoLogin(
        iamToken,
        tenantId,
        shopId
      );
      
      // Retornar resultado
      this.sendSuccessResponse(res, {
        userId: loginResult.userId,
        accessToken: loginResult.accessToken,
        refreshToken: loginResult.refreshToken,
        expiresIn: loginResult.expiresIn,
        shopId: loginResult.shopId,
        roles: loginResult.roles
      });
      
      this.recordRequestResult(endpointName, tenantId, 200);
    } catch (error) {
      this.handleError(error, req, res, endpointName);
    } finally {
      span.end();
    }
  }
  
  /**
   * Renovação de token
   * 
   * POST /api/v1/ecommerce/auth/refresh-token
   */
  public async refreshToken(req: Request, res: Response): Promise<void> {
    const endpointName = 'ecommerce.refresh_token';
    const tenantId = this.getTenantId(req);
    const span = this.tracer.startSpan(endpointName);
    
    try {
      this.recordRequestStart(endpointName, tenantId);
      
      // Validar payload
      const { refreshToken } = req.body;
      
      if (!refreshToken) {
        this.sendErrorResponse(res, 400, 'Refresh token é obrigatório', 'MISSING_REFRESH_TOKEN');
        return;
      }
      
      // Renovar token
      const tokenResult = await this.ecommerceService.refreshToken(
        refreshToken,
        tenantId
      );
      
      // Retornar resultado
      this.sendSuccessResponse(res, {
        accessToken: tokenResult.accessToken,
        refreshToken: tokenResult.refreshToken,
        expiresIn: tokenResult.expiresIn,
        tokenType: tokenResult.tokenType || 'Bearer'
      });
      
      this.recordRequestResult(endpointName, tenantId, 200);
    } catch (error) {
      this.handleError(error, req, res, endpointName);
    } finally {
      span.end();
    }
  }
  
  /**
   * Logout
   * 
   * POST /api/v1/ecommerce/auth/logout
   */
  public async logout(req: Request, res: Response): Promise<void> {
    const endpointName = 'ecommerce.logout';
    const tenantId = this.getTenantId(req);
    const span = this.tracer.startSpan(endpointName);
    
    try {
      this.recordRequestStart(endpointName, tenantId);
      
      // Validar autenticação
      const token = this.getAuthToken(req);
      
      if (!token) {
        this.sendErrorResponse(res, 401, 'Token de autenticação ausente', 'UNAUTHORIZED');
        return;
      }
      
      // Realizar logout
      await this.ecommerceService.logout(token, tenantId);
      
      // Retornar resultado
      this.sendSuccessResponse(res, {
        success: true,
        message: 'Logout realizado com sucesso'
      });
      
      this.recordRequestResult(endpointName, tenantId, 200);
    } catch (error) {
      this.handleError(error, req, res, endpointName);
    } finally {
      span.end();
    }
  }
  
  /**
   * Validação de token
   * 
   * POST /api/v1/ecommerce/auth/validate-token
   */
  public async validateToken(req: Request, res: Response): Promise<void> {
    const endpointName = 'ecommerce.validate_token';
    const span = this.tracer.startSpan(endpointName);
    
    try {
      this.recordRequestStart(endpointName, 'system');
      
      // Validar payload
      const { token } = req.body;
      
      if (!token) {
        this.sendErrorResponse(res, 400, 'Token é obrigatório', 'MISSING_TOKEN');
        return;
      }
      
      // Validar token
      const validation = await this.ecommerceService.validateToken(token);
      
      // Retornar resultado
      this.sendSuccessResponse(res, {
        valid: validation.valid,
        userId: validation.userId,
        roles: validation.roles,
        permissions: validation.permissions,
        expiresAt: validation.expiresAt,
        reason: validation.reason
      });
      
      this.recordRequestResult(endpointName, 'system', 200);
    } catch (error) {
      this.handleError(error, req, res, endpointName);
    } finally {
      span.end();
    }
  }
  
  /**
   * Obter informações do usuário
   * 
   * GET /api/v1/ecommerce/users/{userId}
   */
  public async getUserInfo(req: Request, res: Response): Promise<void> {
    const endpointName = 'ecommerce.get_user_info';
    const tenantId = this.getTenantId(req);
    const span = this.tracer.startSpan(endpointName);
    
    try {
      this.recordRequestStart(endpointName, tenantId);
      
      // Validar autenticação
      const authResult = await this.validateAuthToken(req, res);
      if (!authResult) {
        return;
      }
      
      const userId = req.params.userId;
      
      // Verificar se o usuário tem permissão para acessar os dados
      if (authResult.userId !== userId) {
        // Verificar se tem permissões para acessar outros usuários
        const hasPermission = await this.ecommerceService.checkUserAccess(
          authResult.userId,
          'ecommerce:user:read'
        );
        
        if (!hasPermission) {
          this.sendErrorResponse(res, 403, 'Acesso negado', 'FORBIDDEN');
          return;
        }
      }
      
      // Obter informações do usuário
      const userInfo = await this.ecommerceService.getUserInfo(userId, tenantId);
      
      // Retornar resultado (removendo campos sensíveis)
      this.sendSuccessResponse(res, {
        id: userInfo.id,
        name: userInfo.name,
        email: userInfo.email,
        status: userInfo.status,
        roles: userInfo.roles,
        createdAt: userInfo.createdAt,
        updatedAt: userInfo.updatedAt,
        shops: userInfo.shops,
        preferences: userInfo.preferences
      });
      
      this.recordRequestResult(endpointName, tenantId, 200);
    } catch (error) {
      this.handleError(error, req, res, endpointName);
    } finally {
      span.end();
    }
  }
  
  /**
   * Obter permissões do usuário
   * 
   * GET /api/v1/ecommerce/users/{userId}/permissions
   */
  public async getUserPermissions(req: Request, res: Response): Promise<void> {
    const endpointName = 'ecommerce.get_user_permissions';
    const tenantId = this.getTenantId(req);
    const span = this.tracer.startSpan(endpointName);
    
    try {
      this.recordRequestStart(endpointName, tenantId);
      
      // Validar autenticação
      const authResult = await this.validateAuthToken(req, res);
      if (!authResult) {
        return;
      }
      
      const userId = req.params.userId;
      
      // Verificar se o usuário tem permissão para acessar os dados
      if (authResult.userId !== userId) {
        // Verificar se tem permissões para acessar outros usuários
        const hasPermission = await this.ecommerceService.checkUserAccess(
          authResult.userId,
          'ecommerce:user:permissions:read'
        );
        
        if (!hasPermission) {
          this.sendErrorResponse(res, 403, 'Acesso negado', 'FORBIDDEN');
          return;
        }
      }
      
      // Obter permissões do usuário
      const permissions = await this.ecommerceService.getUserPermissions(userId, tenantId);
      
      // Retornar resultado
      this.sendSuccessResponse(res, permissions);
      
      this.recordRequestResult(endpointName, tenantId, 200);
    } catch (error) {
      this.handleError(error, req, res, endpointName);
    } finally {
      span.end();
    }
  }
  
  /**
   * Obter informações de uma loja
   * 
   * GET /api/v1/ecommerce/shops/{shopId}
   */
  public async getShopInfo(req: Request, res: Response): Promise<void> {
    const endpointName = 'ecommerce.get_shop_info';
    const tenantId = this.getTenantId(req);
    const span = this.tracer.startSpan(endpointName);
    
    try {
      this.recordRequestStart(endpointName, tenantId);
      
      // Validar autenticação
      const authResult = await this.validateAuthToken(req, res);
      if (!authResult) {
        return;
      }
      
      const shopId = req.params.shopId;
      
      // Obter informações da loja
      const shopInfo = await this.ecommerceService.getShopInfo(shopId, tenantId);
      
      // Retornar resultado
      this.sendSuccessResponse(res, shopInfo);
      
      this.recordRequestResult(endpointName, tenantId, 200);
    } catch (error) {
      this.handleError(error, req, res, endpointName);
    } finally {
      span.end();
    }
  }
  
  /**
   * Listar lojas do usuário
   * 
   * GET /api/v1/ecommerce/users/{userId}/shops
   */
  public async getUserShops(req: Request, res: Response): Promise<void> {
    const endpointName = 'ecommerce.get_user_shops';
    const tenantId = this.getTenantId(req);
    const span = this.tracer.startSpan(endpointName);
    
    try {
      this.recordRequestStart(endpointName, tenantId);
      
      // Validar autenticação
      const authResult = await this.validateAuthToken(req, res);
      if (!authResult) {
        return;
      }
      
      const userId = req.params.userId;
      
      // Verificar se o usuário tem permissão para acessar os dados
      if (authResult.userId !== userId) {
        // Verificar se tem permissões para acessar outros usuários
        const hasPermission = await this.ecommerceService.checkUserAccess(
          authResult.userId,
          'ecommerce:user:shops:read'
        );
        
        if (!hasPermission) {
          this.sendErrorResponse(res, 403, 'Acesso negado', 'FORBIDDEN');
          return;
        }
      }
      
      // Obter lojas do usuário
      const userShops = await this.ecommerceService.getUserShops(userId, tenantId);
      
      // Retornar resultado
      this.sendSuccessResponse(res, {
        shops: userShops,
        count: userShops.length
      });
      
      this.recordRequestResult(endpointName, tenantId, 200);
    } catch (error) {
      this.handleError(error, req, res, endpointName);
    } finally {
      span.end();
    }
  }
  
  /**
   * Validar sessão de checkout
   * 
   * GET /api/v1/ecommerce/checkout/sessions/{sessionId}/validate
   */
  public async validateCheckoutSession(req: Request, res: Response): Promise<void> {
    const endpointName = 'ecommerce.validate_checkout_session';
    const tenantId = this.getTenantId(req);
    const span = this.tracer.startSpan(endpointName);
    
    try {
      this.recordRequestStart(endpointName, tenantId);
      
      // Validar autenticação
      const authResult = await this.validateAuthToken(req, res);
      if (!authResult) {
        return;
      }
      
      const sessionId = req.params.sessionId;
      
      // Validar sessão de checkout
      const sessionValidation = await this.ecommerceService.validateCheckoutSession(sessionId, tenantId);
      
      // Retornar resultado
      this.sendSuccessResponse(res, sessionValidation);
      
      this.recordRequestResult(endpointName, tenantId, 200);
    } catch (error) {
      this.handleError(error, req, res, endpointName);
    } finally {
      span.end();
    }
  }
  
  /**
   * Registrar um novo usuário
   * 
   * POST /api/v1/ecommerce/users
   */
  public async registerUser(req: Request, res: Response): Promise<void> {
    const endpointName = 'ecommerce.register_user';
    const tenantId = this.getTenantId(req);
    const span = this.tracer.startSpan(endpointName);
    
    try {
      this.recordRequestStart(endpointName, tenantId);
      
      // Validar payload
      const validationResult = this.validationService.validateSchema(
        'ecommerceUserRegistration', 
        req.body
      );
      
      if (!validationResult.valid) {
        this.sendErrorResponse(
          res, 
          400, 
          'Dados inválidos para registro de usuário', 
          'VALIDATION_ERROR',
          validationResult.errors
        );
        return;
      }
      
      const { email, password, name, phone, shopId, deviceInfo } = req.body;
      
      // Registrar usuário
      const registrationResult = await this.ecommerceService.registerUser({
        email,
        password,
        name,
        phone,
        shopId,
        tenantId,
        deviceInfo: {
          ...deviceInfo,
          ipAddress: req.ip,
          userAgent: req.header('User-Agent') || null
        }
      });
      
      // Retornar resultado
      this.sendSuccessResponse(res, {
        userId: registrationResult.userId,
        email: registrationResult.email,
        name: registrationResult.name,
        accessToken: registrationResult.accessToken,
        refreshToken: registrationResult.refreshToken,
        expiresIn: registrationResult.expiresIn
      }, 201);
      
      this.recordRequestResult(endpointName, tenantId, 201);
    } catch (error) {
      this.handleError(error, req, res, endpointName);
    } finally {
      span.end();
    }
  }
}