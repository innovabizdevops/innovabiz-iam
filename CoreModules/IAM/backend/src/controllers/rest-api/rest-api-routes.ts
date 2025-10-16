/**
 * Configuração de rotas REST para compatibilidade legada
 * 
 * Este arquivo registra todas as rotas REST API para oferecer 
 * compatibilidade com sistemas legados que ainda não utilizam GraphQL.
 */

import { Router } from 'express';
import { MobileMoneyRestController } from './mobile-money-controller';
import { EcommerceRestController } from './ecommerce-controller';
import { Logger } from '../../../../observability/logging/hook_logger';
import { Metrics } from '../../../../observability/metrics/hook_metrics';
import { Tracer } from '../../../../observability/tracing/hook_tracing';

/**
 * Função para configurar todas as rotas REST API
 * 
 * @param router Express Router
 * @param services Objeto com todos os serviços necessários
 */
export function setupRestApiRoutes(
  router: Router,
  services: {
    logger: Logger,
    metrics: Metrics,
    tracer: Tracer,
    mobileMoneyService: any,
    ecommerceService: any,
    iamService: any,
    validationService: any
  }
): Router {
  const { 
    logger, 
    metrics, 
    tracer, 
    mobileMoneyService, 
    ecommerceService, 
    iamService,
    validationService
  } = services;
  
  // Instanciar controladores
  const mobileMoneyController = new MobileMoneyRestController(
    logger,
    metrics,
    tracer,
    mobileMoneyService,
    iamService,
    validationService
  );
  
  const ecommerceController = new EcommerceRestController(
    logger,
    metrics,
    tracer,
    ecommerceService,
    iamService,
    validationService
  );
  
  // Rota para verificar se a API está disponível
  router.get('/api/v1/status', (req, res) => {
    res.json({ 
      status: 'ok',
      version: '1.0.0',
      timestamp: new Date().toISOString()
    });
  });
  
  // --------------------------------------------------
  // Rotas para Mobile Money
  // --------------------------------------------------
  
  // Operações de transação
  router.post('/api/v1/mobile-money/transactions', 
    (req, res) => mobileMoneyController.initiateTransaction(req, res));
  
  router.get('/api/v1/mobile-money/transactions/:transactionId', 
    (req, res) => mobileMoneyController.checkTransactionStatus(req, res));
  
  router.post('/api/v1/mobile-money/transactions/:transactionId/verify-otp', 
    (req, res) => mobileMoneyController.verifyOtp(req, res));
  
  router.post('/api/v1/mobile-money/transactions/:transactionId/cancel', 
    (req, res) => mobileMoneyController.cancelTransaction(req, res));
  
  router.get('/api/v1/mobile-money/transactions', 
    (req, res) => mobileMoneyController.listTransactions(req, res));
  
  // Provedores e elegibilidade
  router.get('/api/v1/mobile-money/providers', 
    (req, res) => mobileMoneyController.getAvailableProviders(req, res));
  
  router.post('/api/v1/mobile-money/check-eligibility', 
    (req, res) => mobileMoneyController.checkEligibility(req, res));
  
  // --------------------------------------------------
  // Rotas para E-Commerce
  // --------------------------------------------------
  
  // Autenticação
  router.post('/api/v1/ecommerce/auth/login', 
    (req, res) => ecommerceController.login(req, res));
  
  router.post('/api/v1/ecommerce/auth/sso-login', 
    (req, res) => ecommerceController.ssoLogin(req, res));
  
  router.post('/api/v1/ecommerce/auth/refresh-token', 
    (req, res) => ecommerceController.refreshToken(req, res));
  
  router.post('/api/v1/ecommerce/auth/logout', 
    (req, res) => ecommerceController.logout(req, res));
  
  router.post('/api/v1/ecommerce/auth/validate-token', 
    (req, res) => ecommerceController.validateToken(req, res));
  
  // Usuários
  router.post('/api/v1/ecommerce/users', 
    (req, res) => ecommerceController.registerUser(req, res));
  
  router.get('/api/v1/ecommerce/users/:userId', 
    (req, res) => ecommerceController.getUserInfo(req, res));
  
  router.get('/api/v1/ecommerce/users/:userId/permissions', 
    (req, res) => ecommerceController.getUserPermissions(req, res));
  
  router.get('/api/v1/ecommerce/users/:userId/shops', 
    (req, res) => ecommerceController.getUserShops(req, res));
  
  // Lojas
  router.get('/api/v1/ecommerce/shops/:shopId', 
    (req, res) => ecommerceController.getShopInfo(req, res));
  
  // Checkout
  router.get('/api/v1/ecommerce/checkout/sessions/:sessionId/validate', 
    (req, res) => ecommerceController.validateCheckoutSession(req, res));
  
  return router;
}