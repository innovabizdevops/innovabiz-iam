/**
 * Controlador base para os endpoints REST
 * 
 * Este arquivo contém a classe base com funcionalidades comuns para
 * todos os controladores REST da API legada de compatibilidade.
 */

import { Request, Response } from 'express';
import { Logger } from '../../../../observability/logging/hook_logger';
import { Metrics } from '../../../../observability/metrics/hook_metrics';
import { Tracer } from '../../../../observability/tracing/hook_tracing';

/**
 * Classe base para os controladores REST
 */
export abstract class BaseRestController {
  protected readonly logger: Logger;
  protected readonly metrics: Metrics;
  protected readonly tracer: Tracer;
  
  /**
   * Construtor da classe base
   * 
   * @param logger Serviço de logging
   * @param metrics Serviço de métricas
   * @param tracer Serviço de tracing
   */
  constructor(logger: Logger, metrics: Metrics, tracer: Tracer) {
    this.logger = logger;
    this.metrics = metrics;
    this.tracer = tracer;
  }
  
  /**
   * Extrai o tenant ID do cabeçalho ou dos parâmetros da requisição
   * 
   * @param req Requisição Express
   * @returns Tenant ID
   */
  protected getTenantId(req: Request): string {
    // Ordem de prioridade: cabeçalho específico, parâmetro de query, valor padrão
    return (
      req.header('X-Tenant-ID') ||
      req.header('X-TenantId') ||
      req.header('X-Tenant') ||
      req.query.tenantId as string ||
      'default'
    );
  }
  
  /**
   * Extrai o token de autenticação dos cabeçalhos
   * 
   * @param req Requisição Express
   * @returns Token de autenticação ou null se não encontrado
   */
  protected getAuthToken(req: Request): string | null {
    const authHeader = req.header('Authorization');
    
    if (!authHeader) {
      return null;
    }
    
    // Formato esperado: "Bearer {token}"
    const parts = authHeader.split(' ');
    
    if (parts.length !== 2 || parts[0] !== 'Bearer') {
      return null;
    }
    
    return parts[1];
  }
  
  /**
   * Registra o início de uma requisição para métricas
   * 
   * @param endpointName Nome do endpoint
   * @param tenantId ID do tenant
   */
  protected recordRequestStart(endpointName: string, tenantId: string): void {
    this.metrics.increment(`api.rest.${endpointName}.request`, { tenantId });
  }
  
  /**
   * Registra o resultado de uma requisição para métricas
   * 
   * @param endpointName Nome do endpoint
   * @param tenantId ID do tenant
   * @param statusCode Código de status HTTP
   */
  protected recordRequestResult(endpointName: string, tenantId: string, statusCode: number): void {
    this.metrics.increment(`api.rest.${endpointName}.response`, { 
      tenantId, 
      statusCode: statusCode.toString(),
      success: statusCode < 400 ? 'true' : 'false'
    });
  }
  
  /**
   * Envia uma resposta de erro padronizada
   * 
   * @param res Resposta Express
   * @param statusCode Código de status HTTP
   * @param message Mensagem de erro
   * @param errorCode Código de erro (opcional)
   * @param details Detalhes adicionais do erro (opcional)
   */
  protected sendErrorResponse(
    res: Response, 
    statusCode: number, 
    message: string, 
    errorCode?: string, 
    details?: any
  ): void {
    res.status(statusCode).json({
      success: false,
      error: {
        message,
        code: errorCode || 'ERROR',
        details: details || null,
        timestamp: new Date().toISOString()
      }
    });
  }
  
  /**
   * Envia uma resposta de sucesso padronizada
   * 
   * @param res Resposta Express
   * @param data Dados da resposta
   * @param statusCode Código de status HTTP (padrão: 200)
   */
  protected sendSuccessResponse(res: Response, data: any, statusCode: number = 200): void {
    res.status(statusCode).json({
      success: true,
      data,
      timestamp: new Date().toISOString()
    });
  }
  
  /**
   * Manipulador de erros para controlar e formatar exceções
   * 
   * @param error Erro capturado
   * @param req Requisição Express
   * @param res Resposta Express
   * @param endpointName Nome do endpoint para log e métricas
   */
  protected handleError(error: any, req: Request, res: Response, endpointName: string): void {
    const tenantId = this.getTenantId(req);
    const requestId = req.header('X-Request-ID') || 'unknown';
    
    this.logger.error(`REST API Error: ${endpointName}`, {
      error,
      tenantId,
      requestId,
      path: req.path,
      method: req.method
    });
    
    // Registrar erro nas métricas
    this.metrics.increment(`api.rest.${endpointName}.error`, { 
      tenantId,
      errorType: error.name || 'Unknown'
    });
    
    // Determinar código de status HTTP e mensagem apropriados
    let statusCode = 500;
    let errorMessage = 'Erro interno do servidor';
    let errorCode = 'INTERNAL_ERROR';
    
    // Mapear tipos de erro para códigos HTTP apropriados
    if (error.name === 'ValidationError') {
      statusCode = 400;
      errorMessage = error.message || 'Dados de entrada inválidos';
      errorCode = 'VALIDATION_ERROR';
    } else if (error.name === 'AuthenticationError') {
      statusCode = 401;
      errorMessage = error.message || 'Autenticação necessária';
      errorCode = 'AUTHENTICATION_ERROR';
    } else if (error.name === 'AuthorizationError') {
      statusCode = 403;
      errorMessage = error.message || 'Acesso negado';
      errorCode = 'AUTHORIZATION_ERROR';
    } else if (error.name === 'NotFoundError') {
      statusCode = 404;
      errorMessage = error.message || 'Recurso não encontrado';
      errorCode = 'NOT_FOUND';
    } else if (error.name === 'ConflictError') {
      statusCode = 409;
      errorMessage = error.message || 'Conflito de recursos';
      errorCode = 'CONFLICT';
    } else if (error.name === 'RateLimitError') {
      statusCode = 429;
      errorMessage = error.message || 'Limite de requisições excedido';
      errorCode = 'RATE_LIMIT_EXCEEDED';
    }
    
    // Enviar resposta de erro formatada
    this.sendErrorResponse(res, statusCode, errorMessage, errorCode, {
      requestId,
      path: req.path
    });
  }
}