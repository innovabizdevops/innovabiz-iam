// ================================================================================================
// INNOVABIZ IAM - REST API CONTROLLER
// ================================================================================================
// Módulo: IAM - REST API Controller
// Autor: Eduardo Jeremias - INNOVABIZ DevOps Team
// Data: 2025-08-03
// Versão: 2.0.0
// Descrição: Controller REST para APIs do módulo IAM
// ================================================================================================

import { Request, Response } from 'express';
import { body, validationResult } from 'express-validator';
import winston from 'winston';
import { IAMService } from '../services/IAMService';

/**
 * Controller para APIs REST do IAM
 */
export class IAMController {
  private readonly iamService: IAMService;
  private readonly logger: winston.Logger;

  constructor(iamService: IAMService, logger: winston.Logger) {
    this.iamService = iamService;
    this.logger = logger;
  }

  /**
   * Validações para criação de usuário
   */
  static createUserValidation = [
    body('tenantId').isUUID().withMessage('tenantId deve ser um UUID válido'),
    body('email').isEmail().withMessage('Email deve ser válido'),
    body('firstName').notEmpty().withMessage('firstName é obrigatório'),
    body('lastName').notEmpty().withMessage('lastName é obrigatório'),
    body('password').isLength({ min: 8 }).withMessage('Password deve ter pelo menos 8 caracteres'),
    body('roles').isArray().withMessage('roles deve ser um array')
  ];

  /**
   * Validações para autenticação
   */
  static authenticateValidation = [
    body('tenantId').isUUID().withMessage('tenantId deve ser um UUID válido'),
    body('email').isEmail().withMessage('Email deve ser válido'),
    body('password').notEmpty().withMessage('Password é obrigatório')
  ];

  /**
   * Validações para validação de token
   */
  static validateTokenValidation = [
    body('token').notEmpty().withMessage('Token é obrigatório')
  ];

  /**
   * Validações para verificação de permissão
   */
  static checkPermissionValidation = [
    body('userId').isUUID().withMessage('userId deve ser um UUID válido'),
    body('tenantId').isUUID().withMessage('tenantId deve ser um UUID válido'),
    body('resource').notEmpty().withMessage('resource é obrigatório'),
    body('action').notEmpty().withMessage('action é obrigatório')
  ];

  /**
   * Criar novo usuário
   * POST /api/v1/users
   */
  async createUser(req: Request, res: Response): Promise<void> {
    try {
      // Verificar validações
      const errors = validationResult(req);
      if (!errors.isEmpty()) {
        res.status(400).json({
          success: false,
          error: 'Dados inválidos',
          details: errors.array()
        });
        return;
      }

      const {
        tenantId,
        email,
        username,
        firstName,
        lastName,
        password,
        roles,
        metadata
      } = req.body;

      const user = await this.iamService.createUser({
        tenantId,
        email,
        username,
        firstName,
        lastName,
        password,
        roles,
        metadata
      });

      // Remover dados sensíveis da resposta
      const userResponse = {
        userId: user.userId,
        tenantId: user.tenantId,
        email: user.email,
        username: user.username,
        firstName: user.firstName,
        lastName: user.lastName,
        roles: user.roles,
        permissions: user.permissions,
        isActive: user.isActive,
        createdAt: user.createdAt,
        updatedAt: user.updatedAt
      };

      this.logger.info('Usuário criado via API', {
        userId: user.userId,
        tenantId,
        email,
        requestId: req.headers['x-request-id']
      });

      res.status(201).json({
        success: true,
        data: userResponse,
        message: 'Usuário criado com sucesso'
      });

    } catch (error) {
      this.logger.error('Erro na criação de usuário via API', {
        error: error.message,
        requestId: req.headers['x-request-id'],
        body: req.body
      });

      res.status(500).json({
        success: false,
        error: error.message || 'Erro interno do servidor'
      });
    }
  }

  /**
   * Autenticar usuário
   * POST /api/v1/auth/login
   */
  async authenticate(req: Request, res: Response): Promise<void> {
    try {
      // Verificar validações
      const errors = validationResult(req);
      if (!errors.isEmpty()) {
        res.status(400).json({
          success: false,
          error: 'Dados inválidos',
          details: errors.array()
        });
        return;
      }

      const { tenantId, email, password } = req.body;
      const ipAddress = req.ip || req.connection.remoteAddress;
      const userAgent = req.get('User-Agent');

      const result = await this.iamService.authenticate({
        tenantId,
        email,
        password,
        ipAddress,
        userAgent
      });

      if (!result.success) {
        res.status(401).json({
          success: false,
          error: result.error
        });
        return;
      }

      // Resposta de sucesso
      const response = {
        success: true,
        data: {
          user: {
            userId: result.user!.userId,
            tenantId: result.user!.tenantId,
            email: result.user!.email,
            username: result.user!.username,
            firstName: result.user!.firstName,
            lastName: result.user!.lastName,
            roles: result.user!.roles,
            permissions: result.user!.permissions,
            lastLoginAt: result.user!.lastLoginAt
          },
          accessToken: result.accessToken,
          refreshToken: result.refreshToken,
          expiresIn: result.expiresIn,
          tokenType: 'Bearer'
        },
        message: 'Autenticação realizada com sucesso'
      };

      this.logger.info('Autenticação bem-sucedida via API', {
        userId: result.user!.userId,
        tenantId,
        email,
        ipAddress,
        requestId: req.headers['x-request-id']
      });

      res.status(200).json(response);

    } catch (error) {
      this.logger.error('Erro na autenticação via API', {
        error: error.message,
        requestId: req.headers['x-request-id'],
        body: { ...req.body, password: '[REDACTED]' }
      });

      res.status(500).json({
        success: false,
        error: 'Erro interno do servidor'
      });
    }
  }

  /**
   * Validar token
   * POST /api/v1/auth/validate
   */
  async validateToken(req: Request, res: Response): Promise<void> {
    try {
      // Verificar validações
      const errors = validationResult(req);
      if (!errors.isEmpty()) {
        res.status(400).json({
          success: false,
          error: 'Dados inválidos',
          details: errors.array()
        });
        return;
      }

      const { token, requiredScopes } = req.body;

      const result = await this.iamService.validateToken({
        token,
        requiredScopes
      });

      if (!result.valid) {
        res.status(401).json({
          success: false,
          error: result.error
        });
        return;
      }

      const response = {
        success: true,
        data: {
          valid: true,
          user: {
            userId: result.user!.userId,
            tenantId: result.user!.tenantId,
            email: result.user!.email,
            username: result.user!.username,
            firstName: result.user!.firstName,
            lastName: result.user!.lastName,
            roles: result.user!.roles,
            permissions: result.user!.permissions,
            isActive: result.user!.isActive
          },
          tenant: {
            tenantId: result.tenant!.tenantId,
            name: result.tenant!.name,
            domain: result.tenant!.domain,
            isActive: result.tenant!.isActive,
            subscriptionPlan: result.tenant!.subscriptionPlan
          },
          scopes: result.scopes
        },
        message: 'Token válido'
      };

      this.logger.debug('Token validado via API', {
        userId: result.user!.userId,
        tenantId: result.user!.tenantId,
        requestId: req.headers['x-request-id']
      });

      res.status(200).json(response);

    } catch (error) {
      this.logger.error('Erro na validação de token via API', {
        error: error.message,
        requestId: req.headers['x-request-id']
      });

      res.status(500).json({
        success: false,
        error: 'Erro interno do servidor'
      });
    }
  }

  /**
   * Verificar permissão
   * POST /api/v1/auth/check-permission
   */
  async checkPermission(req: Request, res: Response): Promise<void> {
    try {
      // Verificar validações
      const errors = validationResult(req);
      if (!errors.isEmpty()) {
        res.status(400).json({
          success: false,
          error: 'Dados inválidos',
          details: errors.array()
        });
        return;
      }

      const { userId, tenantId, resource, action, context } = req.body;

      const result = await this.iamService.checkPermission({
        userId,
        tenantId,
        resource,
        action,
        context
      });

      const response = {
        success: true,
        data: {
          allowed: result.allowed,
          reason: result.reason,
          conditions: result.conditions
        },
        message: result.allowed ? 'Permissão concedida' : 'Permissão negada'
      };

      this.logger.debug('Permissão verificada via API', {
        userId,
        tenantId,
        resource,
        action,
        allowed: result.allowed,
        requestId: req.headers['x-request-id']
      });

      res.status(200).json(response);

    } catch (error) {
      this.logger.error('Erro na verificação de permissão via API', {
        error: error.message,
        requestId: req.headers['x-request-id'],
        body: req.body
      });

      res.status(500).json({
        success: false,
        error: 'Erro interno do servidor'
      });
    }
  }

  /**
   * Buscar usuário por ID
   * GET /api/v1/users/:userId
   */
  async getUserById(req: Request, res: Response): Promise<void> {
    try {
      const { userId } = req.params;
      const { tenantId } = req.query;

      if (!tenantId) {
        res.status(400).json({
          success: false,
          error: 'tenantId é obrigatório'
        });
        return;
      }

      const user = await this.iamService.getUserById(userId, tenantId as string);

      if (!user) {
        res.status(404).json({
          success: false,
          error: 'Usuário não encontrado'
        });
        return;
      }

      // Remover dados sensíveis
      const userResponse = {
        userId: user.userId,
        tenantId: user.tenantId,
        email: user.email,
        username: user.username,
        firstName: user.firstName,
        lastName: user.lastName,
        roles: user.roles,
        permissions: user.permissions,
        isActive: user.isActive,
        lastLoginAt: user.lastLoginAt,
        createdAt: user.createdAt,
        updatedAt: user.updatedAt
      };

      res.status(200).json({
        success: true,
        data: userResponse,
        message: 'Usuário encontrado'
      });

    } catch (error) {
      this.logger.error('Erro ao buscar usuário via API', {
        error: error.message,
        userId: req.params.userId,
        requestId: req.headers['x-request-id']
      });

      res.status(500).json({
        success: false,
        error: 'Erro interno do servidor'
      });
    }
  }

  /**
   * Buscar tenant por ID
   * GET /api/v1/tenants/:tenantId
   */
  async getTenantById(req: Request, res: Response): Promise<void> {
    try {
      const { tenantId } = req.params;

      const tenant = await this.iamService.getTenantById(tenantId);

      if (!tenant) {
        res.status(404).json({
          success: false,
          error: 'Tenant não encontrado'
        });
        return;
      }

      res.status(200).json({
        success: true,
        data: tenant,
        message: 'Tenant encontrado'
      });

    } catch (error) {
      this.logger.error('Erro ao buscar tenant via API', {
        error: error.message,
        tenantId: req.params.tenantId,
        requestId: req.headers['x-request-id']
      });

      res.status(500).json({
        success: false,
        error: 'Erro interno do servidor'
      });
    }
  }

  /**
   * Health check
   * GET /api/v1/health
   */
  async healthCheck(req: Request, res: Response): Promise<void> {
    try {
      const health = {
        status: 'healthy',
        timestamp: new Date().toISOString(),
        version: process.env.npm_package_version || '1.0.0',
        uptime: process.uptime(),
        environment: process.env.NODE_ENV || 'development'
      };

      res.status(200).json(health);

    } catch (error) {
      this.logger.error('Erro no health check', {
        error: error.message,
        requestId: req.headers['x-request-id']
      });

      res.status(500).json({
        status: 'unhealthy',
        error: error.message
      });
    }
  }

  /**
   * Métricas básicas
   * GET /api/v1/metrics
   */
  async getMetrics(req: Request, res: Response): Promise<void> {
    try {
      // Implementar coleta de métricas básicas
      const metrics = {
        timestamp: new Date().toISOString(),
        uptime: process.uptime(),
        memory: process.memoryUsage(),
        cpu: process.cpuUsage(),
        version: process.env.npm_package_version || '1.0.0'
      };

      res.status(200).json({
        success: true,
        data: metrics
      });

    } catch (error) {
      this.logger.error('Erro ao obter métricas', {
        error: error.message,
        requestId: req.headers['x-request-id']
      });

      res.status(500).json({
        success: false,
        error: 'Erro interno do servidor'
      });
    }
  }
}