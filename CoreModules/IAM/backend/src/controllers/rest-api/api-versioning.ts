/**
 * Configuração de Versionamento da API REST
 * 
 * Este arquivo configura o sistema de versionamento da API, registrando
 * diferentes versões e configurando as rotas correspondentes.
 */

import { Router } from 'express';
import { ApiVersionManager } from './version-manager';
import { setupRestApiRoutes } from './rest-api-routes';
import { Logger } from '../../../../observability/logging/hook_logger';
import { Metrics } from '../../../../observability/metrics/hook_metrics';

/**
 * Configura o sistema de versionamento da API REST
 * 
 * @param services Serviços necessários para os controladores
 * @returns Router configurado com todas as versões da API
 */
export function setupVersionedApi(services: any): Router {
  const { logger, metrics } = services;
  
  // Criar gerenciador de versões
  const versionManager = new ApiVersionManager({
    basePath: '/api',
    defaultVersion: 'v1',
    allowVersionOverride: true,
    versionHeaderName: 'X-API-Version',
    versionQueryParam: 'api-version',
    logger,
    metrics
  });
  
  // Configurar versão atual (v1)
  const v1Router = Router();
  setupRestApiRoutes(v1Router, services);
  
  versionManager.registerVersion({
    version: 'v1',
    router: v1Router,
    deprecated: false
  });
  
  // Configurar versão de teste (v2) - para demonstração
  // Em um ambiente de produção, esta versão teria implementações diferentes
  const v2Router = Router();
  
  // Endpoint de exemplo exclusivo da v2
  v2Router.get('/status', (req, res) => {
    res.json({
      status: 'ok',
      version: 'v2',
      features: ['nova_autenticacao', 'rate_limiting_avancado'],
      timestamp: new Date().toISOString()
    });
  });
  
  // Para demonstração, também incluímos algumas rotas da v1 na v2
  // (Na prática, a v2 teria suas próprias implementações)
  setupRestApiRoutes(v2Router, services);
  
  versionManager.registerVersion({
    version: 'v2',
    router: v2Router,
    deprecated: false
  });
  
  // Configurar versão legada (v0) - marcada como obsoleta
  const v0Router = Router();
  
  v0Router.get('/status', (req, res) => {
    res.json({
      status: 'ok',
      version: 'v0 (obsoleta)',
      timestamp: new Date().toISOString()
    });
  });
  
  // Adicionar endpoints legados básicos para compatibilidade
  v0Router.get('/mobile-money/providers', (req, res) => {
    // Redirecionar para a implementação da v1
    req.url = '/api/v1/mobile-money/providers';
    services.app._router.handle(req, res);
  });
  
  // Definir data de descontinuação para daqui a 6 meses
  const sunsetDate = new Date();
  sunsetDate.setMonth(sunsetDate.getMonth() + 6);
  
  versionManager.registerVersion({
    version: 'v0',
    router: v0Router,
    deprecated: true,
    sunset: sunsetDate
  });
  
  // Marcar a versão v0 como obsoleta
  versionManager.deprecateVersion('v0', sunsetDate);
  
  // Criar e retornar o router configurado
  return versionManager.createRouter();
}