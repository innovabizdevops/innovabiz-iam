/**
 * 📊 IAM METRICS SERVICE - INNOVABIZ PLATFORM
 * Serviço avançado de métricas e observabilidade para IAM
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: Prometheus Standards, OpenTelemetry, SRE Best Practices
 * Monitoring: Business Metrics, Technical Metrics, Security Metrics
 */

import { Injectable, Logger, OnModuleInit } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { 
  Counter, 
  Histogram, 
  Gauge, 
  Summary,
  register as prometheusRegister 
} from 'prom-client';

import { User } from '../models/User.entity';
import { Session } from '../models/Session.entity';
import { AuditLog } from '../models/AuditLog.entity';
import { RiskEvent } from '../models/RiskEvent.entity';

// Interfaces para Métricas
interface BusinessMetrics {
  totalUsers: number;
  activeUsers: number;
  newUsersToday: number;
  activeSessions: number;
  averageSessionDuration: number;
  authenticationSuccessRate: number;
  webauthnAdoptionRate: number;
}

interface SecurityMetrics {
  failedLoginAttempts: number;
  blockedIPs: number;
  highRiskEvents: number;
  suspiciousActivities: number;
  mfaUsageRate: number;
  passwordlessRate: number;
}

interface TechnicalMetrics {
  responseTime: number;
  errorRate: number;
  throughput: number;
  databaseConnections: number;
  cacheHitRate: number;
  memoryUsage: number;
  cpuUsage: number;
}

interface CustomMetricOptions {
  name: string;
  help: string;
  labelNames?: string[];
  buckets?: number[];
  percentiles?: number[];
}

@Injectable()
export class IAMMetricsService implements OnModuleInit {
  private readonly logger = new Logger(IAMMetricsService.name);

  // Contadores de negócio
  private readonly userRegistrationsCounter: Counter<string>;
  private readonly loginAttemptsCounter: Counter<string>;
  private readonly authenticationCounter: Counter<string>;
  private readonly webauthnUsageCounter: Counter<string>;
  private readonly sessionCounter: Counter<string>;

  // Contadores de segurança
  private readonly securityEventsCounter: Counter<string>;
  private readonly riskEventsCounter: Counter<string>;
  private readonly blockedRequestsCounter: Counter<string>;
  private readonly mfaEventsCounter: Counter<string>;

  // Contadores técnicos
  private readonly httpRequestsCounter: Counter<string>;
  private readonly httpErrorsCounter: Counter<string>;
  private readonly databaseOperationsCounter: Counter<string>;
  private readonly cacheOperationsCounter: Counter<string>;

  // Histogramas para latência
  private readonly httpRequestDuration: Histogram<string>;
  private readonly databaseQueryDuration: Histogram<string>;
  private readonly authenticationDuration: Histogram<string>;
  private readonly webauthnOperationDuration: Histogram<string>;

  // Gauges para estado atual
  private readonly activeUsersGauge: Gauge<string>;
  private readonly activeSessionsGauge: Gauge<string>;
  private readonly databaseConnectionsGauge: Gauge<string>;
  private readonly memoryUsageGauge: Gauge<string>;
  private readonly cpuUsageGauge: Gauge<string>;

  // Summaries para percentis
  private readonly responseTimeSummary: Summary<string>;
  private readonly sessionDurationSummary: Summary<string>;

  constructor(
    private readonly configService: ConfigService,
    @InjectRepository(User)
    private readonly userRepository: Repository<User>,
    @InjectRepository(Session)
    private readonly sessionRepository: Repository<Session>,
    @InjectRepository(AuditLog)
    private readonly auditRepository: Repository<AuditLog>,
    @InjectRepository(RiskEvent)
    private readonly riskEventRepository: Repository<RiskEvent>
  ) {
    const prefix = 'innovabiz_iam_';

    // Inicializar contadores de negócio
    this.userRegistrationsCounter = new Counter({
      name: `${prefix}user_registrations_total`,
      help: 'Total number of user registrations',
      labelNames: ['tenant_id', 'registration_method', 'status']
    });

    this.loginAttemptsCounter = new Counter({
      name: `${prefix}login_attempts_total`,
      help: 'Total number of login attempts',
      labelNames: ['tenant_id', 'method', 'status', 'risk_level']
    });

    this.authenticationCounter = new Counter({
      name: `${prefix}authentications_total`,
      help: 'Total number of authentication events',
      labelNames: ['tenant_id', 'method', 'status', 'mfa_required']
    });

    this.webauthnUsageCounter = new Counter({
      name: `${prefix}webauthn_operations_total`,
      help: 'Total number of WebAuthn operations',
      labelNames: ['tenant_id', 'operation', 'status', 'authenticator_type']
    });

    this.sessionCounter = new Counter({
      name: `${prefix}sessions_total`,
      help: 'Total number of session events',
      labelNames: ['tenant_id', 'event_type', 'session_type']
    });

    // Inicializar contadores de segurança
    this.securityEventsCounter = new Counter({
      name: `${prefix}security_events_total`,
      help: 'Total number of security events',
      labelNames: ['tenant_id', 'event_type', 'severity', 'source']
    });

    this.riskEventsCounter = new Counter({
      name: `${prefix}risk_events_total`,
      help: 'Total number of risk events',
      labelNames: ['tenant_id', 'risk_type', 'severity', 'status']
    });

    this.blockedRequestsCounter = new Counter({
      name: `${prefix}blocked_requests_total`,
      help: 'Total number of blocked requests',
      labelNames: ['tenant_id', 'reason', 'endpoint', 'ip_address']
    });

    this.mfaEventsCounter = new Counter({
      name: `${prefix}mfa_events_total`,
      help: 'Total number of MFA events',
      labelNames: ['tenant_id', 'method', 'status', 'trigger_reason']
    });

    // Inicializar contadores técnicos
    this.httpRequestsCounter = new Counter({
      name: `${prefix}http_requests_total`,
      help: 'Total number of HTTP requests',
      labelNames: ['method', 'endpoint', 'status_code', 'tenant_id']
    });

    this.httpErrorsCounter = new Counter({
      name: `${prefix}http_errors_total`,
      help: 'Total number of HTTP errors',
      labelNames: ['method', 'endpoint', 'status_code', 'error_type']
    });

    this.databaseOperationsCounter = new Counter({
      name: `${prefix}database_operations_total`,
      help: 'Total number of database operations',
      labelNames: ['operation', 'table', 'status']
    });

    this.cacheOperationsCounter = new Counter({
      name: `${prefix}cache_operations_total`,
      help: 'Total number of cache operations',
      labelNames: ['operation', 'status']
    });

    // Inicializar histogramas
    this.httpRequestDuration = new Histogram({
      name: `${prefix}http_request_duration_seconds`,
      help: 'Duration of HTTP requests in seconds',
      labelNames: ['method', 'endpoint', 'status_code'],
      buckets: [0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10]
    });

    this.databaseQueryDuration = new Histogram({
      name: `${prefix}database_query_duration_seconds`,
      help: 'Duration of database queries in seconds',
      labelNames: ['operation', 'table'],
      buckets: [0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2]
    });

    this.authenticationDuration = new Histogram({
      name: `${prefix}authentication_duration_seconds`,
      help: 'Duration of authentication operations in seconds',
      labelNames: ['method', 'status'],
      buckets: [0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10]
    });

    this.webauthnOperationDuration = new Histogram({
      name: `${prefix}webauthn_operation_duration_seconds`,
      help: 'Duration of WebAuthn operations in seconds',
      labelNames: ['operation', 'status'],
      buckets: [0.1, 0.5, 1, 2, 5, 10, 30]
    });

    // Inicializar gauges
    this.activeUsersGauge = new Gauge({
      name: `${prefix}active_users`,
      help: 'Number of currently active users',
      labelNames: ['tenant_id', 'time_window']
    });

    this.activeSessionsGauge = new Gauge({
      name: `${prefix}active_sessions`,
      help: 'Number of currently active sessions',
      labelNames: ['tenant_id', 'session_type']
    });

    this.databaseConnectionsGauge = new Gauge({
      name: `${prefix}database_connections`,
      help: 'Number of active database connections',
      labelNames: ['pool_name', 'status']
    });

    this.memoryUsageGauge = new Gauge({
      name: `${prefix}memory_usage_bytes`,
      help: 'Memory usage in bytes',
      labelNames: ['type']
    });

    this.cpuUsageGauge = new Gauge({
      name: `${prefix}cpu_usage_percent`,
      help: 'CPU usage percentage',
      labelNames: ['type']
    });

    // Inicializar summaries
    this.responseTimeSummary = new Summary({
      name: `${prefix}response_time_seconds`,
      help: 'Response time in seconds',
      labelNames: ['endpoint', 'method'],
      percentiles: [0.5, 0.9, 0.95, 0.99]
    });

    this.sessionDurationSummary = new Summary({
      name: `${prefix}session_duration_seconds`,
      help: 'Session duration in seconds',
      labelNames: ['tenant_id', 'session_type'],
      percentiles: [0.5, 0.9, 0.95, 0.99]
    });
  }

  /**
   * Inicializar métricas no startup do módulo
   */
  async onModuleInit(): Promise<void> {
    try {
      this.logger.log('Initializing IAM metrics service...');
      
      // Registrar todas as métricas no Prometheus
      this.registerMetrics();
      
      // Iniciar coleta de métricas periódicas
      this.startPeriodicMetricsCollection();
      
      // Coletar métricas iniciais
      await this.collectInitialMetrics();
      
      this.logger.log('IAM metrics service initialized successfully');
    } catch (error) {
      this.logger.error(`Failed to initialize metrics service: ${error.message}`, error.stack);
    }
  }

  /**
   * Registrar todas as métricas no Prometheus
   */
  private registerMetrics(): void {
    const metrics = [
      // Contadores
      this.userRegistrationsCounter,
      this.loginAttemptsCounter,
      this.authenticationCounter,
      this.webauthnUsageCounter,
      this.sessionCounter,
      this.securityEventsCounter,
      this.riskEventsCounter,
      this.blockedRequestsCounter,
      this.mfaEventsCounter,
      this.httpRequestsCounter,
      this.httpErrorsCounter,
      this.databaseOperationsCounter,
      this.cacheOperationsCounter,
      
      // Histogramas
      this.httpRequestDuration,
      this.databaseQueryDuration,
      this.authenticationDuration,
      this.webauthnOperationDuration,
      
      // Gauges
      this.activeUsersGauge,
      this.activeSessionsGauge,
      this.databaseConnectionsGauge,
      this.memoryUsageGauge,
      this.cpuUsageGauge,
      
      // Summaries
      this.responseTimeSummary,
      this.sessionDurationSummary
    ];

    metrics.forEach(metric => {
      try {
        prometheusRegister.registerMetric(metric);
      } catch (error) {
        // Métrica já registrada, ignorar
      }
    });
  }

  /**
   * Iniciar coleta periódica de métricas
   */
  private startPeriodicMetricsCollection(): void {
    const interval = this.configService.get<number>('monitoring.metricsCollectionInterval', 30000); // 30 segundos
    
    setInterval(async () => {
      try {
        await this.collectPeriodicMetrics();
      } catch (error) {
        this.logger.warn(`Periodic metrics collection failed: ${error.message}`);
      }
    }, interval);
  }

  /**
   * Coletar métricas iniciais
   */
  private async collectInitialMetrics(): Promise<void> {
    try {
      // Coletar métricas de usuários
      await this.collectUserMetrics();
      
      // Coletar métricas de sessões
      await this.collectSessionMetrics();
      
      // Coletar métricas de sistema
      this.collectSystemMetrics();
      
    } catch (error) {
      this.logger.warn(`Initial metrics collection failed: ${error.message}`);
    }
  }

  /**
   * Coletar métricas periódicas
   */
  private async collectPeriodicMetrics(): Promise<void> {
    await Promise.allSettled([
      this.collectUserMetrics(),
      this.collectSessionMetrics(),
      this.collectSecurityMetrics(),
      this.collectSystemMetrics()
    ]);
  }

  /**
   * Coletar métricas de usuários
   */
  private async collectUserMetrics(): Promise<void> {
    try {
      // Total de usuários ativos por tenant
      const activeUsersByTenant = await this.userRepository
        .createQueryBuilder('user')
        .select('user.tenantId, COUNT(*) as count')
        .where('user.isActive = :isActive', { isActive: true })
        .groupBy('user.tenantId')
        .getRawMany();

      activeUsersByTenant.forEach(({ tenantId, count }) => {
        this.activeUsersGauge.set({ tenant_id: tenantId, time_window: 'current' }, parseInt(count));
      });

      // Novos usuários hoje
      const today = new Date();
      today.setHours(0, 0, 0, 0);
      
      const newUsersToday = await this.userRepository.count({
        where: {
          createdAt: { $gte: today } as any
        }
      });

      this.userRegistrationsCounter.inc({ 
        tenant_id: 'all', 
        registration_method: 'all', 
        status: 'completed' 
      }, newUsersToday);

    } catch (error) {
      this.logger.warn(`User metrics collection failed: ${error.message}`);
    }
  }

  /**
   * Coletar métricas de sessões
   */
  private async collectSessionMetrics(): Promise<void> {
    try {
      // Sessões ativas por tenant
      const activeSessionsByTenant = await this.sessionRepository
        .createQueryBuilder('session')
        .select('session.tenantId, session.type, COUNT(*) as count')
        .where('session.isActive = :isActive', { isActive: true })
        .groupBy('session.tenantId, session.type')
        .getRawMany();

      activeSessionsByTenant.forEach(({ tenantId, type, count }) => {
        this.activeSessionsGauge.set({ 
          tenant_id: tenantId, 
          session_type: type 
        }, parseInt(count));
      });

    } catch (error) {
      this.logger.warn(`Session metrics collection failed: ${error.message}`);
    }
  }

  /**
   * Coletar métricas de segurança
   */
  private async collectSecurityMetrics(): Promise<void> {
    try {
      const last24Hours = new Date(Date.now() - 24 * 60 * 60 * 1000);

      // Eventos de segurança nas últimas 24 horas
      const securityEvents = await this.auditRepository
        .createQueryBuilder('audit')
        .select('audit.tenantId, audit.action, audit.severity, COUNT(*) as count')
        .where('audit.timestamp >= :since', { since: last24Hours })
        .andWhere('audit.severity IN (:...severities)', { severities: ['warning', 'error', 'critical'] })
        .groupBy('audit.tenantId, audit.action, audit.severity')
        .getRawMany();

      securityEvents.forEach(({ tenantId, action, severity, count }) => {
        this.securityEventsCounter.inc({
          tenant_id: tenantId,
          event_type: action,
          severity,
          source: 'audit_log'
        }, parseInt(count));
      });

      // Eventos de risco
      const riskEvents = await this.riskEventRepository
        .createQueryBuilder('risk')
        .select('risk.tenantId, risk.eventType, risk.severity, COUNT(*) as count')
        .where('risk.timestamp >= :since', { since: last24Hours })
        .groupBy('risk.tenantId, risk.eventType, risk.severity')
        .getRawMany();

      riskEvents.forEach(({ tenantId, eventType, severity, count }) => {
        this.riskEventsCounter.inc({
          tenant_id: tenantId,
          risk_type: eventType,
          severity,
          status: 'detected'
        }, parseInt(count));
      });

    } catch (error) {
      this.logger.warn(`Security metrics collection failed: ${error.message}`);
    }
  }

  /**
   * Coletar métricas de sistema
   */
  private collectSystemMetrics(): void {
    try {
      // Métricas de memória
      const memoryUsage = process.memoryUsage();
      this.memoryUsageGauge.set({ type: 'heap_used' }, memoryUsage.heapUsed);
      this.memoryUsageGauge.set({ type: 'heap_total' }, memoryUsage.heapTotal);
      this.memoryUsageGauge.set({ type: 'external' }, memoryUsage.external);
      this.memoryUsageGauge.set({ type: 'rss' }, memoryUsage.rss);

      // Métricas de CPU
      const cpuUsage = process.cpuUsage();
      this.cpuUsageGauge.set({ type: 'user' }, cpuUsage.user / 1000000); // Converter para segundos
      this.cpuUsageGauge.set({ type: 'system' }, cpuUsage.system / 1000000);

    } catch (error) {
      this.logger.warn(`System metrics collection failed: ${error.message}`);
    }
  }

  // ========================================
  // MÉTODOS PÚBLICOS PARA INSTRUMENTAÇÃO
  // ========================================

  /**
   * Registrar tentativa de login
   */
  recordLoginAttempt(tenantId: string, method: string, status: string, riskLevel: string): void {
    this.loginAttemptsCounter.inc({ tenant_id: tenantId, method, status, risk_level: riskLevel });
  }

  /**
   * Registrar autenticação
   */
  recordAuthentication(tenantId: string, method: string, status: string, mfaRequired: boolean): void {
    this.authenticationCounter.inc({ 
      tenant_id: tenantId, 
      method, 
      status, 
      mfa_required: mfaRequired.toString() 
    });
  }

  /**
   * Registrar operação WebAuthn
   */
  recordWebAuthnOperation(
    tenantId: string, 
    operation: string, 
    status: string, 
    authenticatorType: string,
    duration: number
  ): void {
    this.webauthnUsageCounter.inc({ 
      tenant_id: tenantId, 
      operation, 
      status, 
      authenticator_type: authenticatorType 
    });
    
    this.webauthnOperationDuration.observe({ operation, status }, duration / 1000);
  }

  /**
   * Registrar evento de sessão
   */
  recordSessionEvent(tenantId: string, eventType: string, sessionType: string): void {
    this.sessionCounter.inc({ tenant_id: tenantId, event_type: eventType, session_type: sessionType });
  }

  /**
   * Registrar evento de segurança
   */
  recordSecurityEvent(tenantId: string, eventType: string, severity: string, source: string): void {
    this.securityEventsCounter.inc({ tenant_id: tenantId, event_type: eventType, severity, source });
  }

  /**
   * Registrar requisição HTTP
   */
  recordHttpRequest(
    method: string, 
    endpoint: string, 
    statusCode: number, 
    duration: number, 
    tenantId?: string
  ): void {
    const labels = { method, endpoint, status_code: statusCode.toString() };
    
    this.httpRequestsCounter.inc({ ...labels, tenant_id: tenantId || 'unknown' });
    this.httpRequestDuration.observe(labels, duration / 1000);
    this.responseTimeSummary.observe({ endpoint, method }, duration / 1000);
    
    if (statusCode >= 400) {
      this.httpErrorsCounter.inc({ 
        ...labels, 
        error_type: statusCode >= 500 ? 'server_error' : 'client_error' 
      });
    }
  }

  /**
   * Registrar operação de banco de dados
   */
  recordDatabaseOperation(operation: string, table: string, status: string, duration: number): void {
    this.databaseOperationsCounter.inc({ operation, table, status });
    this.databaseQueryDuration.observe({ operation, table }, duration / 1000);
  }

  /**
   * Registrar operação de cache
   */
  recordCacheOperation(operation: string, status: string): void {
    this.cacheOperationsCounter.inc({ operation, status });
  }

  /**
   * Registrar duração de sessão
   */
  recordSessionDuration(tenantId: string, sessionType: string, duration: number): void {
    this.sessionDurationSummary.observe({ tenant_id: tenantId, session_type: sessionType }, duration);
  }

  /**
   * Obter métricas de negócio
   */
  async getBusinessMetrics(): Promise<BusinessMetrics> {
    try {
      const [
        totalUsers,
        activeUsers,
        newUsersToday,
        activeSessions
      ] = await Promise.all([
        this.userRepository.count(),
        this.userRepository.count({ where: { isActive: true } }),
        this.userRepository.count({
          where: {
            createdAt: { $gte: new Date(Date.now() - 24 * 60 * 60 * 1000) } as any
          }
        }),
        this.sessionRepository.count({ where: { isActive: true } })
      ]);

      return {
        totalUsers,
        activeUsers,
        newUsersToday,
        activeSessions,
        averageSessionDuration: 0, // Calcular baseado nos dados
        authenticationSuccessRate: 0.95, // Calcular baseado nos logs
        webauthnAdoptionRate: 0.75 // Calcular baseado nos dados
      };

    } catch (error) {
      this.logger.error(`Failed to get business metrics: ${error.message}`);
      throw error;
    }
  }

  /**
   * Obter métricas de segurança
   */
  async getSecurityMetrics(): Promise<SecurityMetrics> {
    try {
      const last24Hours = new Date(Date.now() - 24 * 60 * 60 * 1000);

      const [
        failedLoginAttempts,
        highRiskEvents
      ] = await Promise.all([
        this.auditRepository.count({
          where: {
            action: 'LOGIN_FAILED',
            timestamp: { $gte: last24Hours } as any
          }
        }),
        this.riskEventRepository.count({
          where: {
            severity: { $in: ['high', 'critical'] } as any,
            timestamp: { $gte: last24Hours } as any
          }
        })
      ]);

      return {
        failedLoginAttempts,
        blockedIPs: 0, // Calcular baseado no cache
        highRiskEvents,
        suspiciousActivities: 0, // Calcular baseado nos eventos
        mfaUsageRate: 0.60, // Calcular baseado nos dados
        passwordlessRate: 0.80 // Calcular baseado nos dados
      };

    } catch (error) {
      this.logger.error(`Failed to get security metrics: ${error.message}`);
      throw error;
    }
  }

  /**
   * Criar métrica customizada
   */
  createCustomMetric(type: 'counter' | 'gauge' | 'histogram' | 'summary', options: CustomMetricOptions): any {
    const prefix = 'innovabiz_iam_custom_';
    const name = `${prefix}${options.name}`;

    try {
      switch (type) {
        case 'counter':
          return new Counter({
            name,
            help: options.help,
            labelNames: options.labelNames || []
          });
        
        case 'gauge':
          return new Gauge({
            name,
            help: options.help,
            labelNames: options.labelNames || []
          });
        
        case 'histogram':
          return new Histogram({
            name,
            help: options.help,
            labelNames: options.labelNames || [],
            buckets: options.buckets || [0.001, 0.01, 0.1, 1, 10]
          });
        
        case 'summary':
          return new Summary({
            name,
            help: options.help,
            labelNames: options.labelNames || [],
            percentiles: options.percentiles || [0.5, 0.9, 0.99]
          });
        
        default:
          throw new Error(`Unsupported metric type: ${type}`);
      }
    } catch (error) {
      this.logger.error(`Failed to create custom metric: ${error.message}`);
      throw error;
    }
  }

  /**
   * Inicializar métricas (chamado pelo módulo)
   */
  initializeMetrics(): void {
    this.logger.log('IAM metrics initialized');
  }
}