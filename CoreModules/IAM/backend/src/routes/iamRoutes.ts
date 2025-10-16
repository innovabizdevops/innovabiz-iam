// ================================================================================================
// INNOVABIZ IAM - API ROUTES
// ================================================================================================
// Módulo: IAM - API Routes
// Autor: Eduardo Jeremias - INNOVABIZ DevOps Team
// Data: 2025-08-03
// Versão: 2.0.0
// Descrição: Definição das rotas da API REST do módulo IAM
// ================================================================================================

import { Router } from 'express';
import { IAMController } from '../controllers/IAMController';

/**
 * Criar rotas do IAM
 */
export function createIAMRoutes(iamController: IAMController): Router {
  const router = Router();

  // ================================================================================================
  // ROTAS PÚBLICAS (sem autenticação)
  // ================================================================================================

  /**
   * Health Check
   * GET /api/v1/health
   */
  router.get('/health', iamController.healthCheck.bind(iamController));

  /**
   * Métricas básicas
   * GET /api/v1/metrics
   */
  router.get('/metrics', iamController.getMetrics.bind(iamController));

  // ================================================================================================
  // ROTAS DE AUTENTICAÇÃO
  // ================================================================================================

  /**
   * Autenticar usuário
   * POST /api/v1/auth/login
   */
  router.post(
    '/auth/login',
    IAMController.authenticateValidation,
    iamController.authenticate.bind(iamController)
  );

  /**
   * Validar token
   * POST /api/v1/auth/validate
   */
  router.post(
    '/auth/validate',
    IAMController.validateTokenValidation,
    iamController.validateToken.bind(iamController)
  );

  /**
   * Verificar permissão
   * POST /api/v1/auth/check-permission
   */
  router.post(
    '/auth/check-permission',
    IAMController.checkPermissionValidation,
    iamController.checkPermission.bind(iamController)
  );

  // ================================================================================================
  // ROTAS PROTEGIDAS (requerem autenticação)
  // ================================================================================================

  /**
   * Criar usuário
   * POST /api/v1/users
   */
  router.post(
    '/users',
    IAMController.createUserValidation,
    iamController.createUser.bind(iamController)
  );

  /**
   * Buscar usuário por ID
   * GET /api/v1/users/:userId
   */
  router.get(
    '/users/:userId',
    iamController.getUserById.bind(iamController)
  );

  /**
   * Buscar tenant por ID
   * GET /api/v1/tenants/:tenantId
   */
  router.get(
    '/tenants/:tenantId',
    iamController.getTenantById.bind(iamController)
  );

  return router;
}