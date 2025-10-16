/**
 * 📊 METRICS INTERCEPTOR - INNOVABIZ IAM
 * Interceptor para coleta avançada de métricas e observabilidade
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: Prometheus, OpenTelemetry, Grafana, DataDog
 * Mercados: Angola (BNA), Europa, América, China, BRICS, Brasil
 * Observabilidade: Métricas, Traces, Logs, Alertas, Dashboards
 */

import {
  Injectable,
  NestInterceptor,
  ExecutionContext,
  CallHandler,
  Logger
} from '@nestjs/common';
import { Observable, throwError } from 'rxjs';
import { tap, catchError, finalize } from 'rxjs/operators';
import { Reflector } from '@nestjs/core';
import { ConfigService } from '@nestjs/config';
import { Request, Response } from 'express';
import * as prometheus from 'prom-client';

/**
 * Configurações do Metrics Interceptor
 */
export interface MetricsConfig {
  // Métricas básicas
  basic: {
    requestCount: boolean;
    requestDuration: boolean;
    responseSize: boolean;
    errorRate: boolean;
  };
  
  // Métricas de negócio
  business: {
    userActions: boolean;
    tenantActivity: boolean;
    featureUsage: boolean;
    conversionFunnels: boolean;
  };
  
  // Métricas de segurança
  security: {
    authenticationAttempts: boolean;
    authorizationFailures: boolean;
    riskScores: boolean;
    suspiciousActivity: boolean;
  };
  
  // Métricas de performance
  performance: {
    databaseQueries: boolean;
    cacheHitRatio: boolean;
    externalApiCalls: boolean;
    memoryUsage: boolean;
  };
  
  // Métricas de compliance
  compliance: {
    auditEvents: boolean;
    dataAccess: boolean;
    privacyEvents: boolean;
    regulatoryReporting: boolean;
  };
  
  // Configurações de coleta
  collection: {
    sampleRate: number; // 0.0 - 1.0
    batchSize: number;
    flushInterval: number; // em segundos
    retentionPeriod: number; // em dias
  };
  
  // Configurações de exportação
  export: {
    prometheus: boolean;
    datadog: boolean;
    newrelic: boolean;
    customEndpoint?: string;
  };
  
  // Configurações de alertas
  alerting: {
    enabled: boolean;
    thresholds: {
      errorRate: number;
      responseTime: number;
      throughput: number;
    };
    channels: string[];
  };
  
  // Labels e dimensões
  labels: {
    includeUserId: boolean;
    includeTenantId: boolean;
    includeEndpoint: boolean;
    includeMethod: boolean;
    includeStatusCode: boolean;
    includeUserAgent: boolean;
    includeRegion: boolean;
    customLabels?: Record<string, string>;
  };
}

/**
 * Contexto de métricas
 */
export interface MetricsContext {
  // Identificadores
  requestId: string;
  userId?: string;
  tenantId?: string;
  sessionId?: string;
  
  // Request info
  method: string;
  endpoint: string;
  path: string;
  statusCode?: number;
  
  // Timing
  startTime: number;
  endTime?: number;
  duration?: number;
  
  // Sizes
  requestSize: number;
  responseSize?: number;
  
  // User context
  userAgent: string;
  ipAddress: string;
  region?: string;
  
  // Business context
  feature?: string;
  action?: string;
  category?: string;
  
  // Security context
  riskScore?: number;
  authMethod?: string;
  mfaUsed?: boolean;
  
  // Performance context
  dbQueries?: number;
  dbDuration?: number;
  cacheHits?: number;
  cacheMisses?: number;
  
  // Error context
  error?: {
    type: string;
    message: string;
    stack?: string;
  };
  
  // Custom metrics
  customMetrics?: Record<string, number>;
  customLabels?: Record<string, string>;
}

/**
 * Configuração padrão
 */
const DEFAULT_CONFIG: MetricsConfig = {
  basic: {
    requestCount: true,
    requestDuration: true,
    responseSize: true,
    errorRate: true
  },
  business: {
    userActions: true,
    tenantActivity: true,
    featureUsage: true,
    conversionFunnels: false
  },
  security: {
    authenticationAttempts: true,
    authorizationFailures: true,
    riskScores: true,
    suspiciousActivity: true
  },
  performance: {
    databaseQueries: true,
    cacheHitRatio: true,
    externalApiCalls: true,
    memoryUsage: false
  },
  compliance: {
    auditEvents: true,
    dataAccess: true,
    privacyEvents: true,
    regulatoryReporting: false
  },
  collection: {
    sampleRate: 1.0,
    batchSize: 100,
    flushInterval: 60,
    retentionPeriod: 30
  },
  export: {
    prometheus: true,
    datadog: false,
    newrelic: false
  },
  alerting: {
    enabled: true,
    thresholds: {
      errorRate: 0.05, // 5%
      responseTime: 1000, // 1s
      throughput: 100 // req/s
    },
    channels: ['email', 'slack']
  },
  labels: {
    includeUserId: false, // Privacy concern
    includeTenantId: true,
    includeEndpoint: true,
    includeMethod: true,
    includeStatusCode: true,
    includeUserAgent: false,
    includeRegion: true
  }
};

/**
 * Decorator para configurar métricas
 */
export const CollectMetrics = (config?: Partial<MetricsConfig>) => {
  return (target: any, propertyKey?: string, descriptor?: PropertyDescriptor) => {
    const mergedConfig = { ...DEFAULT_CONFIG, ...config };
    Reflect.defineMetadata('metrics-config', mergedConfig, target, propertyKey);
  };
};

/**
 * Interceptor de métricas
 */
@Injectable()
export class MetricsInterceptor implements NestInterceptor {
  private readonly logger = new Logger(MetricsInterceptor.name);
  
  // Métricas Prometheus
  private readonly requestCounter: prometheus.Counter<string>;
  private readonly requestDuration: prometheus.Histogram<string>;
  private readonly responseSize: prometheus.Histogram<string>;
  private readonly errorCounter: prometheus.Counter<string>;
  private readonly businessMetrics: prometheus.Counter<string>;
  private readonly securityMetrics: prometheus.Counter<string>;
  private readonly performanceMetrics: prometheus.Histogram<string>;
  private readonly complianceMetrics: prometheus.Counter<string>;
  
  // Buffer para métricas em lote
  private metricsBuffer: MetricsContext[] = [];
  private flushTimer?: NodeJS.Timeout;

  constructor(
    private readonly reflector: Reflector,
    private readonly configService: ConfigService
  ) {
    // Inicializar métricas Prometheus
    this.requestCounter = new prometheus.Counter({
      name: 'iam_http_requests_total',
      help: 'Total number of HTTP requests',
      labelNames: ['method', 'endpoint', 'status_code', 'tenant_id', 'region']
    });

    this.requestDuration = new prometheus.Histogram({
      name: 'iam_http_request_duration_seconds',
      help: 'HTTP request duration in seconds',
      labelNames: ['method', 'endpoint', 'tenant_id', 'region'],
      buckets: [0.1, 0.3, 0.5, 0.7, 1, 3, 5, 7, 10]
    });

    this.responseSize = new prometheus.Histogram({
      name: 'iam_http_response_size_bytes',
      help: 'HTTP response size in bytes',
      labelNames: ['method', 'endpoint', 'tenant_id'],
      buckets: [100, 1000, 10000, 100000, 1000000]
    });

    this.errorCounter = new prometheus.Counter({
      name: 'iam_http_errors_total',
      help: 'Total number of HTTP errors',
      labelNames: ['method', 'endpoint', 'status_code', 'error_type', 'tenant_id']
    });

    this.businessMetrics = new prometheus.Counter({
      name: 'iam_business_events_total',
      help: 'Total number of business events',
      labelNames: ['event_type', 'feature', 'action', 'tenant_id', 'user_id']
    });

    this.securityMetrics = new prometheus.Counter({
      name: 'iam_security_events_total',
      help: 'Total number of security events',
      labelNames: ['event_type', 'risk_level', 'auth_method', 'tenant_id']
    });

    this.performanceMetrics = new prometheus.Histogram({
      name: 'iam_performance_metrics',
      help: 'Performance metrics',
      labelNames: ['metric_type', 'endpoint', 'tenant_id'],
      buckets: [1, 5, 10, 25, 50, 100, 250, 500, 1000]
    });

    this.complianceMetrics = new prometheus.Counter({
      name: 'iam_compliance_events_total',
      help: 'Total number of compliance events',
      labelNames: ['event_type', 'framework', 'severity', 'tenant_id']
    });

    // Registrar métricas
    prometheus.register.registerMetric(this.requestCounter);
    prometheus.register.registerMetric(this.requestDuration);
    prometheus.register.registerMetric(this.responseSize);
    prometheus.register.registerMetric(this.errorCounter);
    prometheus.register.registerMetric(this.businessMetrics);
    prometheus.register.registerMetric(this.securityMetrics);
    prometheus.register.registerMetric(this.performanceMetrics);
    prometheus.register.registerMetric(this.complianceMetrics);
  }

  intercept(context: ExecutionContext, next: CallHandler): Observable<any> {
    const config = this.getMetricsConfig(context);
    
    if (!config) {
      return next.handle();
    }

    const request = context.switchToHttp().getRequest<Request>();
    const response = context.switchToHttp().getResponse<Response>();
    
    // Verificar sample rate
    if (Math.random() > config.collection.sampleRate) {
      return next.handle();
    }

    const metricsContext = this.createMetricsContext(request, context);
    
    return next.handle().pipe(
      tap((data) => {
        this.handleSuccess(metricsContext, response, data, config);
      }),
      catchError((error) => {
        this.handleError(metricsContext, response, error, config);
        return throwError(error);
      }),
      finalize(() => {
        this.finalizeMetrics(metricsContext, response, config);
      })
    );
  }

  /**
   * Obter configuração de métricas
   */
  private getMetricsConfig(context: ExecutionContext): MetricsConfig | null {
    return this.reflector.get<MetricsConfig>('metrics-config', context.getHandler()) ||
           this.reflector.get<MetricsConfig>('metrics-config', context.getClass()) ||
           DEFAULT_CONFIG;
  }

  /**
   * Criar contexto de métricas
   */
  private createMetricsContext(request: Request, context: ExecutionContext): MetricsContext {
    const user = (request as any).user;
    const tenant = (request as any).tenant;
    
    return {
      requestId: this.generateRequestId(),
      userId: user?.id,
      tenantId: tenant?.id,
      sessionId: user?.sessionId,
      method: request.method,
      endpoint: this.normalizeEndpoint(request.path),
      path: request.path,
      startTime: Date.now(),
      requestSize: this.calculateRequestSize(request),
      userAgent: request.get('User-Agent') || 'unknown',
      ipAddress: this.getClientIP(request),
      region: this.extractRegion(request),
      feature: this.extractFeature(context),
      action: this.extractAction(context),
      category: this.extractCategory(context)
    };
  }

  /**
   * Tratar sucesso da requisição
   */
  private handleSuccess(
    context: MetricsContext,
    response: Response,
    data: any,
    config: MetricsConfig
  ): void {
    context.endTime = Date.now();
    context.duration = context.endTime - context.startTime;
    context.statusCode = response.statusCode;
    context.responseSize = this.calculateResponseSize(data);
    
    // Coletar métricas básicas
    if (config.basic.requestCount) {
      this.collectRequestCount(context, config);
    }
    
    if (config.basic.requestDuration) {
      this.collectRequestDuration(context, config);
    }
    
    if (config.basic.responseSize) {
      this.collectResponseSize(context, config);
    }
    
    // Coletar métricas de negócio
    if (config.business.userActions) {
      this.collectBusinessMetrics(context, config);
    }
    
    // Coletar métricas de performance
    if (config.performance.databaseQueries) {
      this.collectPerformanceMetrics(context, config);
    }
    
    // Adicionar ao buffer
    this.addToBuffer(context, config);
  }

  /**
   * Tratar erro da requisição
   */
  private handleError(
    context: MetricsContext,
    response: Response,
    error: any,
    config: MetricsConfig
  ): void {
    context.endTime = Date.now();
    context.duration = context.endTime - context.startTime;
    context.statusCode = response.statusCode || 500;
    context.error = {
      type: error.constructor.name,
      message: error.message,
      stack: error.stack
    };
    
    // Coletar métricas de erro
    if (config.basic.errorRate) {
      this.collectErrorMetrics(context, config);
    }
    
    // Coletar métricas de segurança se for erro de auth
    if (config.security.authorizationFailures && this.isAuthError(error)) {
      this.collectSecurityMetrics(context, config, 'authorization_failure');
    }
    
    // Adicionar ao buffer
    this.addToBuffer(context, config);
  }

  /**
   * Finalizar coleta de métricas
   */
  private finalizeMetrics(
    context: MetricsContext,
    response: Response,
    config: MetricsConfig
  ): void {
    // Coletar métricas de compliance
    if (config.compliance.auditEvents) {
      this.collectComplianceMetrics(context, config);
    }
    
    // Verificar alertas
    if (config.alerting.enabled) {
      this.checkAlerts(context, config);
    }
    
    // Log de debug
    this.logger.debug(`Metrics collected for ${context.method} ${context.endpoint}`, {
      requestId: context.requestId,
      duration: context.duration,
      statusCode: context.statusCode,
      tenantId: context.tenantId
    });
  }

  /**
   * Coletar métricas de contagem de requisições
   */
  private collectRequestCount(context: MetricsContext, config: MetricsConfig): void {
    const labels = this.buildLabels(context, config, [
      'method', 'endpoint', 'status_code', 'tenant_id', 'region'
    ]);
    
    this.requestCounter.inc(labels);
  }

  /**
   * Coletar métricas de duração de requisições
   */
  private collectRequestDuration(context: MetricsContext, config: MetricsConfig): void {
    const labels = this.buildLabels(context, config, [
      'method', 'endpoint', 'tenant_id', 'region'
    ]);
    
    this.requestDuration.observe(labels, (context.duration || 0) / 1000);
  }

  /**
   * Coletar métricas de tamanho de resposta
   */
  private collectResponseSize(context: MetricsContext, config: MetricsConfig): void {
    const labels = this.buildLabels(context, config, [
      'method', 'endpoint', 'tenant_id'
    ]);
    
    this.responseSize.observe(labels, context.responseSize || 0);
  }

  /**
   * Coletar métricas de erro
   */
  private collectErrorMetrics(context: MetricsContext, config: MetricsConfig): void {
    const labels = this.buildLabels(context, config, [
      'method', 'endpoint', 'status_code', 'error_type', 'tenant_id'
    ]);
    
    labels.error_type = context.error?.type || 'unknown';
    this.errorCounter.inc(labels);
  }

  /**
   * Coletar métricas de negócio
   */
  private collectBusinessMetrics(context: MetricsContext, config: MetricsConfig): void {
    if (!context.feature || !context.action) return;
    
    const labels = this.buildLabels(context, config, [
      'event_type', 'feature', 'action', 'tenant_id', 'user_id'
    ]);
    
    labels.event_type = 'user_action';
    labels.feature = context.feature;
    labels.action = context.action;
    
    // Remover user_id se não permitido por privacidade
    if (!config.labels.includeUserId) {
      delete labels.user_id;
    }
    
    this.businessMetrics.inc(labels);
  }

  /**
   * Coletar métricas de segurança
   */
  private collectSecurityMetrics(
    context: MetricsContext,
    config: MetricsConfig,
    eventType: string
  ): void {
    const labels = this.buildLabels(context, config, [
      'event_type', 'risk_level', 'auth_method', 'tenant_id'
    ]);
    
    labels.event_type = eventType;
    labels.risk_level = this.getRiskLevel(context.riskScore);
    labels.auth_method = context.authMethod || 'unknown';
    
    this.securityMetrics.inc(labels);
  }

  /**
   * Coletar métricas de performance
   */
  private collectPerformanceMetrics(context: MetricsContext, config: MetricsConfig): void {
    const labels = this.buildLabels(context, config, [
      'metric_type', 'endpoint', 'tenant_id'
    ]);
    
    // Métricas de banco de dados
    if (context.dbQueries) {
      labels.metric_type = 'db_queries';
      this.performanceMetrics.observe(labels, context.dbQueries);
    }
    
    if (context.dbDuration) {
      labels.metric_type = 'db_duration';
      this.performanceMetrics.observe(labels, context.dbDuration);
    }
    
    // Métricas de cache
    if (context.cacheHits || context.cacheMisses) {
      const hitRatio = (context.cacheHits || 0) / ((context.cacheHits || 0) + (context.cacheMisses || 0));
      labels.metric_type = 'cache_hit_ratio';
      this.performanceMetrics.observe(labels, hitRatio);
    }
  }

  /**
   * Coletar métricas de compliance
   */
  private collectComplianceMetrics(context: MetricsContext, config: MetricsConfig): void {
    const labels = this.buildLabels(context, config, [
      'event_type', 'framework', 'severity', 'tenant_id'
    ]);
    
    labels.event_type = 'data_access';
    labels.framework = 'gdpr'; // Seria determinado dinamicamente
    labels.severity = 'info';
    
    this.complianceMetrics.inc(labels);
  }

  /**
   * Construir labels para métricas
   */
  private buildLabels(
    context: MetricsContext,
    config: MetricsConfig,
    labelNames: string[]
  ): Record<string, string> {
    const labels: Record<string, string> = {};
    
    for (const labelName of labelNames) {
      switch (labelName) {
        case 'method':
          if (config.labels.includeMethod) {
            labels.method = context.method;
          }
          break;
        case 'endpoint':
          if (config.labels.includeEndpoint) {
            labels.endpoint = context.endpoint;
          }
          break;
        case 'status_code':
          if (config.labels.includeStatusCode) {
            labels.status_code = (context.statusCode || 0).toString();
          }
          break;
        case 'tenant_id':
          if (config.labels.includeTenantId && context.tenantId) {
            labels.tenant_id = context.tenantId;
          }
          break;
        case 'user_id':
          if (config.labels.includeUserId && context.userId) {
            labels.user_id = context.userId;
          }
          break;
        case 'region':
          if (config.labels.includeRegion && context.region) {
            labels.region = context.region;
          }
          break;
      }
    }
    
    // Adicionar labels customizados
    if (config.labels.customLabels) {
      Object.assign(labels, config.labels.customLabels);
    }
    
    if (context.customLabels) {
      Object.assign(labels, context.customLabels);
    }
    
    return labels;
  }

  /**
   * Adicionar contexto ao buffer
   */
  private addToBuffer(context: MetricsContext, config: MetricsConfig): void {
    this.metricsBuffer.push(context);
    
    if (this.metricsBuffer.length >= config.collection.batchSize) {
      this.flushBuffer(config);
    }
    
    // Configurar timer de flush se não existir
    if (!this.flushTimer) {
      this.flushTimer = setTimeout(() => {
        this.flushBuffer(config);
        this.flushTimer = undefined;
      }, config.collection.flushInterval * 1000);
    }
  }

  /**
   * Flush do buffer de métricas
   */
  private flushBuffer(config: MetricsConfig): void {
    if (this.metricsBuffer.length === 0) return;
    
    const batch = [...this.metricsBuffer];
    this.metricsBuffer = [];
    
    // Exportar para diferentes sistemas
    if (config.export.datadog) {
      this.exportToDatadog(batch);
    }
    
    if (config.export.newrelic) {
      this.exportToNewRelic(batch);
    }
    
    if (config.export.customEndpoint) {
      this.exportToCustomEndpoint(batch, config.export.customEndpoint);
    }
    
    this.logger.debug(`Flushed ${batch.length} metrics to external systems`);
  }

  /**
   * Verificar alertas
   */
  private checkAlerts(context: MetricsContext, config: MetricsConfig): void {
    const thresholds = config.alerting.thresholds;
    
    // Verificar tempo de resposta
    if (context.duration && context.duration > thresholds.responseTime) {
      this.triggerAlert('high_response_time', context, config);
    }
    
    // Verificar taxa de erro (seria calculada baseada em histórico)
    if (context.error && this.calculateErrorRate() > thresholds.errorRate) {
      this.triggerAlert('high_error_rate', context, config);
    }
  }

  /**
   * Disparar alerta
   */
  private triggerAlert(
    alertType: string,
    context: MetricsContext,
    config: MetricsConfig
  ): void {
    this.logger.warn(`Alert triggered: ${alertType}`, {
      alertType,
      requestId: context.requestId,
      endpoint: context.endpoint,
      duration: context.duration,
      tenantId: context.tenantId
    });
    
    // Implementar envio de alertas para canais configurados
    for (const channel of config.alerting.channels) {
      this.sendAlert(channel, alertType, context);
    }
  }

  // Métodos auxiliares
  private generateRequestId(): string {
    return `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  private normalizeEndpoint(path: string): string {
    // Normalizar path removendo IDs específicos
    return path
      .replace(/\/\d+/g, '/:id')
      .replace(/\/[0-9a-f-]{36}/g, '/:uuid')
      .replace(/\/[0-9a-f]{24}/g, '/:objectid');
  }

  private calculateRequestSize(request: Request): number {
    const contentLength = request.get('Content-Length');
    return contentLength ? parseInt(contentLength, 10) : 0;
  }

  private calculateResponseSize(data: any): number {
    if (!data) return 0;
    return JSON.stringify(data).length;
  }

  private getClientIP(request: Request): string {
    const forwarded = request.get('X-Forwarded-For');
    const realIP = request.get('X-Real-IP');
    const cfConnectingIP = request.get('CF-Connecting-IP');
    
    if (cfConnectingIP) return cfConnectingIP;
    if (realIP) return realIP;
    if (forwarded) return forwarded.split(',')[0].trim();
    
    return request.ip || request.connection.remoteAddress || 'unknown';
  }

  private extractRegion(request: Request): string {
    return request.get('CF-IPCountry') || 
           request.get('X-Country-Code') || 
           'unknown';
  }

  private extractFeature(context: ExecutionContext): string {
    const handler = context.getHandler();
    return Reflect.getMetadata('feature', handler) || 'unknown';
  }

  private extractAction(context: ExecutionContext): string {
    const handler = context.getHandler();
    return Reflect.getMetadata('action', handler) || 'unknown';
  }

  private extractCategory(context: ExecutionContext): string {
    const handler = context.getHandler();
    return Reflect.getMetadata('category', handler) || 'unknown';
  }

  private isAuthError(error: any): boolean {
    return error.name === 'UnauthorizedException' || 
           error.name === 'ForbiddenException' ||
           error.status === 401 || 
           error.status === 403;
  }

  private getRiskLevel(riskScore?: number): string {
    if (!riskScore) return 'unknown';
    if (riskScore >= 0.8) return 'critical';
    if (riskScore >= 0.6) return 'high';
    if (riskScore >= 0.3) return 'medium';
    return 'low';
  }

  private calculateErrorRate(): number {
    // Implementar cálculo de taxa de erro baseado em histórico
    return 0; // Placeholder
  }

  private async exportToDatadog(batch: MetricsContext[]): Promise<void> {
    // Implementar exportação para DataDog
    this.logger.debug(`Exporting ${batch.length} metrics to DataDog`);
  }

  private async exportToNewRelic(batch: MetricsContext[]): Promise<void> {
    // Implementar exportação para New Relic
    this.logger.debug(`Exporting ${batch.length} metrics to New Relic`);
  }

  private async exportToCustomEndpoint(batch: MetricsContext[], endpoint: string): Promise<void> {
    // Implementar exportação para endpoint customizado
    this.logger.debug(`Exporting ${batch.length} metrics to ${endpoint}`);
  }

  private async sendAlert(channel: string, alertType: string, context: MetricsContext): Promise<void> {
    // Implementar envio de alertas
    this.logger.debug(`Sending alert ${alertType} to ${channel}`);
  }
}