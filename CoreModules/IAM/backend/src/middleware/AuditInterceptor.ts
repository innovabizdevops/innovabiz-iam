/**
 * 📋 AUDIT INTERCEPTOR - INNOVABIZ IAM
 * Interceptor avançado de auditoria para compliance e governança
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: GDPR, LGPD, SOX 404, PCI DSS, ISO 27001, NIST SP 800-53
 * Standards: OWASP Logging, SANS Audit Guidelines, ISACA COBIT
 */

import {
  Injectable,
  NestInterceptor,
  ExecutionContext,
  CallHandler,
  Logger,
  Inject
} from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { Reflector } from '@nestjs/core';
import { Observable, throwError } from 'rxjs';
import { tap, catchError, finalize } from 'rxjs/operators';
import { Request, Response } from 'express';
import { Cache } from 'cache-manager';
import { CACHE_MANAGER } from '@nestjs/cache-manager';
import * as crypto from 'crypto';

import { AuditService } from '../services/AuditService';
import { RiskAssessmentService } from '../services/RiskAssessmentService';

// Interfaces para Auditoria
interface AuditContext {
  requestId: string;
  sessionId?: string;
  userId?: string;
  tenantId?: string;
  ipAddress: string;
  userAgent: string;
  timestamp: Date;
  endpoint: string;
  method: string;
  headers: Record<string, any>;
  body?: any;
  query?: any;
  params?: any;
}

interface AuditResult {
  statusCode: number;
  responseTime: number;
  responseSize?: number;
  error?: any;
  riskScore?: number;
  complianceFlags: string[];
}

interface SensitiveDataPattern {
  field: string;
  pattern: RegExp;
  maskingStrategy: 'full' | 'partial' | 'hash' | 'remove';
  complianceFramework: string[];
}

@Injectable()
export class AuditInterceptor implements NestInterceptor {
  private readonly logger = new Logger(AuditInterceptor.name);
  
  // Padrões para dados sensíveis
  private readonly sensitiveDataPatterns: SensitiveDataPattern[] = [
    {
      field: 'password',
      pattern: /password|pwd|pass/i,
      maskingStrategy: 'full',
      complianceFramework: ['GDPR', 'LGPD', 'PCI_DSS']
    },
    {
      field: 'email',
      pattern: /email|e-mail/i,
      maskingStrategy: 'partial',
      complianceFramework: ['GDPR', 'LGPD']
    },
    {
      field: 'phone',
      pattern: /phone|telephone|mobile|celular/i,
      maskingStrategy: 'partial',
      complianceFramework: ['GDPR', 'LGPD']
    },
    {
      field: 'document',
      pattern: /cpf|cnpj|ssn|passport|rg|identity/i,
      maskingStrategy: 'hash',
      complianceFramework: ['GDPR', 'LGPD', 'HIPAA']
    },
    {
      field: 'creditCard',
      pattern: /card|credit|debit|pan/i,
      maskingStrategy: 'full',
      complianceFramework: ['PCI_DSS']
    },
    {
      field: 'token',
      pattern: /token|jwt|bearer|api[_-]?key/i,
      maskingStrategy: 'full',
      complianceFramework: ['OWASP', 'NIST']
    }
  ];

  constructor(
    private readonly configService: ConfigService,
    private readonly reflector: Reflector,
    private readonly auditService: AuditService,
    private readonly riskAssessmentService: RiskAssessmentService,
    @Inject(CACHE_MANAGER) private readonly cacheManager: Cache
  ) {}

  intercept(context: ExecutionContext, next: CallHandler): Observable<any> {
    const request = context.switchToHttp().getRequest<Request>();
    const response = context.switchToHttp().getResponse<Response>();
    
    // Verificar se auditoria está habilitada
    const skipAudit = this.reflector.get<boolean>('skipAudit', context.getHandler());
    if (skipAudit || !this.isAuditEnabled()) {
      return next.handle();
    }

    // Criar contexto de auditoria
    const auditContext = this.createAuditContext(request, context);
    const startTime = Date.now();

    // Log de início da requisição
    this.logRequestStart(auditContext);

    return next.handle().pipe(
      tap((data) => {
        // Requisição bem-sucedida
        const responseTime = Date.now() - startTime;
        const auditResult: AuditResult = {
          statusCode: response.statusCode,
          responseTime,
          responseSize: this.calculateResponseSize(data),
          complianceFlags: this.identifyComplianceFlags(auditContext, data)
        };
        
        this.logRequestSuccess(auditContext, auditResult, data);
      }),
      catchError((error) => {
        // Requisição com erro
        const responseTime = Date.now() - startTime;
        const auditResult: AuditResult = {
          statusCode: error.status || 500,
          responseTime,
          error: this.sanitizeError(error),
          complianceFlags: this.identifyComplianceFlags(auditContext, null, error)
        };
        
        this.logRequestError(auditContext, auditResult, error);
        return throwError(error);
      }),
      finalize(() => {
        // Limpeza e processamento final
        this.finalizeAudit(auditContext);
      })
    );
  }

  /**
   * Criar contexto de auditoria
   */
  private createAuditContext(request: Request, context: ExecutionContext): AuditContext {
    const requestId = this.generateRequestId();
    const userId = request['user']?.sub;
    const tenantId = request['user']?.tenantId;
    const sessionId = request.get('X-Session-Id') || request['sessionId'];

    return {
      requestId,
      sessionId,
      userId,
      tenantId,
      ipAddress: this.getClientIP(request),
      userAgent: request.get('User-Agent') || 'unknown',
      timestamp: new Date(),
      endpoint: request.path,
      method: request.method,
      headers: this.sanitizeHeaders(request.headers),
      body: this.sanitizeRequestData(request.body),
      query: this.sanitizeRequestData(request.query),
      params: this.sanitizeRequestData(request.params)
    };
  }

  /**
   * Log de início da requisição
   */
  private async logRequestStart(context: AuditContext): Promise<void> {
    try {
      // Log estruturado para início da requisição
      this.logger.log(`Request started: ${context.method} ${context.endpoint}`, {
        requestId: context.requestId,
        userId: context.userId,
        tenantId: context.tenantId,
        ip: context.ipAddress,
        userAgent: context.userAgent,
        timestamp: context.timestamp.toISOString()
      });

      // Armazenar contexto no cache para correlação
      await this.cacheManager.set(
        `audit_context:${context.requestId}`,
        context,
        300 // 5 minutos
      );

      // Verificar se é uma operação sensível
      if (this.isSensitiveOperation(context)) {
        await this.auditService.logEvent({
          userId: context.userId,
          tenantId: context.tenantId,
          action: 'SENSITIVE_OPERATION_STARTED',
          resource: 'API',
          resourceId: context.endpoint,
          severity: 'info',
          metadata: {
            requestId: context.requestId,
            method: context.method,
            endpoint: context.endpoint,
            ip: context.ipAddress,
            userAgent: context.userAgent
          }
        });
      }
    } catch (error) {
      this.logger.error(`Failed to log request start: ${error.message}`, error.stack);
    }
  }

  /**
   * Log de requisição bem-sucedida
   */
  private async logRequestSuccess(
    context: AuditContext,
    result: AuditResult,
    responseData: any
  ): Promise<void> {
    try {
      // Log estruturado para sucesso
      this.logger.log(`Request completed successfully: ${context.method} ${context.endpoint}`, {
        requestId: context.requestId,
        userId: context.userId,
        statusCode: result.statusCode,
        responseTime: result.responseTime,
        responseSize: result.responseSize
      });

      // Avaliar risco da operação
      const riskScore = await this.assessOperationRisk(context, result);
      result.riskScore = riskScore;

      // Log de auditoria detalhado
      await this.auditService.logEvent({
        userId: context.userId,
        tenantId: context.tenantId,
        action: this.mapEndpointToAction(context.method, context.endpoint),
        resource: this.extractResourceFromEndpoint(context.endpoint),
        resourceId: context.params?.id || context.requestId,
        severity: this.determineSeverity(context, result),
        metadata: {
          requestId: context.requestId,
          sessionId: context.sessionId,
          method: context.method,
          endpoint: context.endpoint,
          statusCode: result.statusCode,
          responseTime: result.responseTime,
          responseSize: result.responseSize,
          ip: context.ipAddress,
          userAgent: context.userAgent,
          riskScore,
          complianceFlags: result.complianceFlags,
          sanitizedResponse: this.sanitizeResponseData(responseData)
        }
      });

      // Métricas de performance
      await this.recordPerformanceMetrics(context, result);

    } catch (error) {
      this.logger.error(`Failed to log request success: ${error.message}`, error.stack);
    }
  }

  /**
   * Log de requisição com erro
   */
  private async logRequestError(
    context: AuditContext,
    result: AuditResult,
    error: any
  ): Promise<void> {
    try {
      // Log estruturado para erro
      this.logger.error(`Request failed: ${context.method} ${context.endpoint}`, {
        requestId: context.requestId,
        userId: context.userId,
        statusCode: result.statusCode,
        responseTime: result.responseTime,
        error: result.error
      });

      // Avaliar risco elevado para erros
      const riskScore = await this.assessOperationRisk(context, result, error);
      result.riskScore = riskScore;

      // Log de auditoria para erro
      await this.auditService.logSecurityEvent({
        userId: context.userId,
        tenantId: context.tenantId,
        action: 'REQUEST_FAILED',
        severity: this.determineErrorSeverity(error),
        metadata: {
          requestId: context.requestId,
          sessionId: context.sessionId,
          method: context.method,
          endpoint: context.endpoint,
          statusCode: result.statusCode,
          responseTime: result.responseTime,
          ip: context.ipAddress,
          userAgent: context.userAgent,
          error: result.error,
          riskScore,
          complianceFlags: result.complianceFlags
        }
      });

      // Detectar padrões de ataque
      await this.detectAttackPatterns(context, error);

    } catch (auditError) {
      this.logger.error(`Failed to log request error: ${auditError.message}`, auditError.stack);
    }
  }

  /**
   * Finalizar auditoria
   */
  private async finalizeAudit(context: AuditContext): Promise<void> {
    try {
      // Limpar contexto do cache
      await this.cacheManager.del(`audit_context:${context.requestId}`);
      
      // Processar alertas se necessário
      await this.processAuditAlerts(context);
      
    } catch (error) {
      this.logger.error(`Failed to finalize audit: ${error.message}`, error.stack);
    }
  }

  /**
   * Sanitizar dados da requisição
   */
  private sanitizeRequestData(data: any): any {
    if (!data || typeof data !== 'object') {
      return data;
    }

    const sanitized = { ...data };
    
    for (const pattern of this.sensitiveDataPatterns) {
      this.applySanitization(sanitized, pattern);
    }

    return sanitized;
  }

  /**
   * Sanitizar dados da resposta
   */
  private sanitizeResponseData(data: any): any {
    if (!data || typeof data !== 'object') {
      return null; // Não logar dados de resposta por padrão
    }

    // Apenas logar metadados básicos da resposta
    return {
      type: Array.isArray(data) ? 'array' : 'object',
      size: Array.isArray(data) ? data.length : Object.keys(data).length,
      hasData: !!data
    };
  }

  /**
   * Sanitizar headers
   */
  private sanitizeHeaders(headers: any): Record<string, any> {
    const sanitized = { ...headers };
    
    // Remover headers sensíveis
    const sensitiveHeaders = [
      'authorization',
      'cookie',
      'x-api-key',
      'x-auth-token',
      'x-session-id'
    ];

    sensitiveHeaders.forEach(header => {
      if (sanitized[header]) {
        sanitized[header] = '***REDACTED***';
      }
    });

    return sanitized;
  }

  /**
   * Aplicar sanitização baseada em padrões
   */
  private applySanitization(obj: any, pattern: SensitiveDataPattern): void {
    if (!obj || typeof obj !== 'object') return;

    Object.keys(obj).forEach(key => {
      if (pattern.pattern.test(key)) {
        switch (pattern.maskingStrategy) {
          case 'full':
            obj[key] = '***REDACTED***';
            break;
          case 'partial':
            obj[key] = this.partialMask(obj[key]);
            break;
          case 'hash':
            obj[key] = this.hashValue(obj[key]);
            break;
          case 'remove':
            delete obj[key];
            break;
        }
      } else if (typeof obj[key] === 'object') {
        this.applySanitization(obj[key], pattern);
      }
    });
  }

  /**
   * Mascaramento parcial
   */
  private partialMask(value: any): string {
    if (!value || typeof value !== 'string') return '***';
    
    if (value.includes('@')) {
      // Email
      const [local, domain] = value.split('@');
      return `${local.charAt(0)}***@${domain}`;
    }
    
    // Outros valores
    const len = value.length;
    if (len <= 3) return '***';
    
    return `${value.substring(0, 2)}***${value.substring(len - 2)}`;
  }

  /**
   * Hash de valor
   */
  private hashValue(value: any): string {
    if (!value) return '***HASHED***';
    
    return crypto
      .createHash('sha256')
      .update(value.toString())
      .digest('hex')
      .substring(0, 16) + '***';
  }

  /**
   * Sanitizar erro
   */
  private sanitizeError(error: any): any {
    return {
      name: error.name,
      message: error.message,
      status: error.status,
      code: error.code,
      timestamp: new Date().toISOString()
    };
  }

  /**
   * Mapear endpoint para ação
   */
  private mapEndpointToAction(method: string, endpoint: string): string {
    const actions: Record<string, string> = {
      'POST:/users': 'USER_CREATED',
      'GET:/users': 'USER_ACCESSED',
      'PUT:/users': 'USER_UPDATED',
      'DELETE:/users': 'USER_DELETED',
      'POST:/auth/login': 'USER_LOGIN',
      'POST:/auth/logout': 'USER_LOGOUT',
      'POST:/webauthn/register': 'WEBAUTHN_REGISTER',
      'POST:/webauthn/authenticate': 'WEBAUTHN_AUTHENTICATE'
    };

    const key = `${method}:${endpoint}`;
    return actions[key] || `${method}_${endpoint.replace(/[^a-zA-Z0-9]/g, '_').toUpperCase()}`;
  }

  /**
   * Extrair recurso do endpoint
   */
  private extractResourceFromEndpoint(endpoint: string): string {
    const segments = endpoint.split('/').filter(s => s);
    return segments[segments.length - 1] || 'unknown';
  }

  /**
   * Determinar severidade
   */
  private determineSeverity(context: AuditContext, result: AuditResult): string {
    if (result.riskScore && result.riskScore > 0.8) return 'critical';
    if (result.riskScore && result.riskScore > 0.6) return 'high';
    if (this.isSensitiveOperation(context)) return 'medium';
    return 'info';
  }

  /**
   * Determinar severidade do erro
   */
  private determineErrorSeverity(error: any): string {
    if (error.status >= 500) return 'critical';
    if (error.status === 401 || error.status === 403) return 'high';
    if (error.status === 429) return 'medium';
    return 'low';
  }

  /**
   * Verificar se é operação sensível
   */
  private isSensitiveOperation(context: AuditContext): boolean {
    const sensitiveEndpoints = [
      '/auth/',
      '/users',
      '/webauthn/',
      '/admin/',
      '/risk/',
      '/audit/'
    ];

    return sensitiveEndpoints.some(endpoint => context.endpoint.includes(endpoint));
  }

  /**
   * Identificar flags de compliance
   */
  private identifyComplianceFlags(
    context: AuditContext,
    responseData?: any,
    error?: any
  ): string[] {
    const flags: string[] = [];

    // GDPR/LGPD - Dados pessoais
    if (this.containsPersonalData(context)) {
      flags.push('GDPR_PERSONAL_DATA', 'LGPD_PERSONAL_DATA');
    }

    // PCI DSS - Dados de pagamento
    if (context.endpoint.includes('/payment') || context.endpoint.includes('/card')) {
      flags.push('PCI_DSS_PAYMENT_DATA');
    }

    // SOX 404 - Controles financeiros
    if (context.endpoint.includes('/financial') || context.endpoint.includes('/audit')) {
      flags.push('SOX_404_FINANCIAL_CONTROL');
    }

    // HIPAA - Dados de saúde (se aplicável)
    if (context.endpoint.includes('/health') || context.endpoint.includes('/medical')) {
      flags.push('HIPAA_HEALTH_DATA');
    }

    return flags;
  }

  /**
   * Verificar se contém dados pessoais
   */
  private containsPersonalData(context: AuditContext): boolean {
    const personalDataFields = ['email', 'phone', 'name', 'address', 'document'];
    
    const checkObject = (obj: any): boolean => {
      if (!obj || typeof obj !== 'object') return false;
      
      return Object.keys(obj).some(key => 
        personalDataFields.some(field => key.toLowerCase().includes(field))
      );
    };

    return checkObject(context.body) || checkObject(context.query) || checkObject(context.params);
  }

  /**
   * Avaliar risco da operação
   */
  private async assessOperationRisk(
    context: AuditContext,
    result: AuditResult,
    error?: any
  ): Promise<number> {
    try {
      let riskScore = 0;

      // Risco base por tipo de operação
      if (this.isSensitiveOperation(context)) riskScore += 0.3;
      if (error) riskScore += 0.4;
      if (result.statusCode >= 400) riskScore += 0.2;

      // Risco por IP/localização
      const ipRisk = await this.assessIPRisk(context.ipAddress);
      riskScore += ipRisk * 0.3;

      // Risco por padrões de comportamento
      const behaviorRisk = await this.assessBehaviorRisk(context);
      riskScore += behaviorRisk * 0.4;

      return Math.min(1, riskScore);
    } catch (assessmentError) {
      this.logger.warn(`Risk assessment failed: ${assessmentError.message}`);
      return 0.5; // Risco médio como fallback
    }
  }

  /**
   * Avaliar risco do IP
   */
  private async assessIPRisk(ip: string): Promise<number> {
    try {
      // Verificar histórico de bloqueios
      const blockHistory = await this.cacheManager.get<number>(`ip_blocks:${ip}`) || 0;
      return Math.min(1, blockHistory * 0.2);
    } catch (error) {
      return 0.1;
    }
  }

  /**
   * Avaliar risco comportamental
   */
  private async assessBehaviorRisk(context: AuditContext): Promise<number> {
    try {
      if (!context.userId) return 0.3; // Usuário não autenticado

      // Usar serviço de avaliação de risco
      const riskProfile = await this.riskAssessmentService.getRiskProfile(
        context.userId,
        context.tenantId || 'default'
      );
      
      return riskProfile?.riskScore || 0.2;
    } catch (error) {
      return 0.2;
    }
  }

  /**
   * Detectar padrões de ataque
   */
  private async detectAttackPatterns(context: AuditContext, error: any): Promise<void> {
    try {
      const patterns = [
        { name: 'SQL_INJECTION', pattern: /(\bor\b|\band\b|union|select|insert|update|delete|drop)/i },
        { name: 'XSS_ATTEMPT', pattern: /<script|javascript:|onerror|onload/i },
        { name: 'PATH_TRAVERSAL', pattern: /\.\.|\/etc\/|\/proc\/|\/var\//i },
        { name: 'COMMAND_INJECTION', pattern: /;|\||&|`|\$\(/i }
      ];

      const requestString = JSON.stringify(context.body || context.query || context.params || '');
      
      for (const pattern of patterns) {
        if (pattern.pattern.test(requestString)) {
          await this.auditService.logSecurityEvent({
            userId: context.userId,
            tenantId: context.tenantId,
            action: 'ATTACK_PATTERN_DETECTED',
            severity: 'critical',
            metadata: {
              requestId: context.requestId,
              attackType: pattern.name,
              ip: context.ipAddress,
              userAgent: context.userAgent,
              endpoint: context.endpoint,
              method: context.method
            }
          });
        }
      }
    } catch (error) {
      this.logger.error(`Attack pattern detection failed: ${error.message}`);
    }
  }

  /**
   * Processar alertas de auditoria
   */
  private async processAuditAlerts(context: AuditContext): Promise<void> {
    // Implementar lógica de alertas baseada em regras
    // Por exemplo: múltiplas tentativas de login falhadas, acessos suspeitos, etc.
  }

  /**
   * Registrar métricas de performance
   */
  private async recordPerformanceMetrics(context: AuditContext, result: AuditResult): Promise<void> {
    try {
      // Implementar coleta de métricas para Prometheus
      // Exemplo: tempo de resposta, taxa de erro, throughput
    } catch (error) {
      this.logger.warn(`Failed to record performance metrics: ${error.message}`);
    }
  }

  /**
   * Calcular tamanho da resposta
   */
  private calculateResponseSize(data: any): number {
    if (!data) return 0;
    
    try {
      return JSON.stringify(data).length;
    } catch (error) {
      return 0;
    }
  }

  /**
   * Verificar se auditoria está habilitada
   */
  private isAuditEnabled(): boolean {
    return this.configService.get<boolean>('audit.enableAuditLogging', true);
  }

  /**
   * Gerar ID único para requisição
   */
  private generateRequestId(): string {
    return `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  /**
   * Obter IP real do cliente
   */
  private getClientIP(request: Request): string {
    const forwarded = request.get('X-Forwarded-For');
    const realIP = request.get('X-Real-IP');
    const cfConnectingIP = request.get('CF-Connecting-IP');
    
    if (cfConnectingIP) return cfConnectingIP;
    if (realIP) return realIP;
    if (forwarded) return forwarded.split(',')[0].trim();
    
    return request.ip || request.connection.remoteAddress || 'unknown';
  }
}