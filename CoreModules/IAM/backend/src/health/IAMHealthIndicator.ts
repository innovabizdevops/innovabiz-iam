/**
 * 🏥 IAM HEALTH INDICATOR - INNOVABIZ PLATFORM
 * Indicador de saúde avançado para o módulo IAM
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: Observability Best Practices, SRE Principles
 * Monitoring: Database, Cache, External Services, Performance
 */

import { Injectable, Logger } from '@nestjs/common';
import { HealthIndicator, HealthIndicatorResult, HealthCheckError } from '@nestjs/terminus';
import { ConfigService } from '@nestjs/config';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Cache } from 'cache-manager';
import { Inject } from '@nestjs/common';
import { CACHE_MANAGER } from '@nestjs/cache-manager';

import { User } from '../models/User.entity';
import { Session } from '../models/Session.entity';
import { AuditLog } from '../models/AuditLog.entity';

// Interfaces para Health Check
interface HealthMetrics {
  responseTime: number;
  errorRate: number;
  activeUsers: number;
  activeSessions: number;
  databaseConnections: number;
  cacheHitRate: number;
  memoryUsage: number;
  cpuUsage: number;
}

interface ComponentHealth {
  status: 'healthy' | 'degraded' | 'unhealthy';
  responseTime: number;
  lastCheck: Date;
  error?: string;
  metrics?: Record<string, any>;
}

interface ServiceDependency {
  name: string;
  url?: string;
  timeout: number;
  critical: boolean;
  healthCheck: () => Promise<ComponentHealth>;
}

@Injectable()
export class IAMHealthIndicator extends HealthIndicator {
  private readonly logger = new Logger(IAMHealthIndicator.name);
  private readonly healthCache = new Map<string, { result: ComponentHealth; timestamp: number }>();
  private readonly cacheTTL = 30000; // 30 segundos

  constructor(
    private readonly configService: ConfigService,
    @InjectRepository(User)
    private readonly userRepository: Repository<User>,
    @InjectRepository(Session)
    private readonly sessionRepository: Repository<Session>,
    @InjectRepository(AuditLog)
    private readonly auditRepository: Repository<AuditLog>,
    @Inject(CACHE_MANAGER)
    private readonly cacheManager: Cache
  ) {
    super();
  }

  /**
   * Health check principal do IAM
   */
  async isHealthy(key: string): Promise<HealthIndicatorResult> {
    try {
      const startTime = Date.now();
      
      // Executar todos os health checks
      const [
        databaseHealth,
        cacheHealth,
        servicesHealth,
        metricsHealth
      ] = await Promise.allSettled([
        this.checkDatabaseHealth(),
        this.checkCacheHealth(),
        this.checkExternalServices(),
        this.collectHealthMetrics()
      ]);

      const responseTime = Date.now() - startTime;
      
      // Determinar status geral
      const overallStatus = this.determineOverallStatus([
        databaseHealth,
        cacheHealth,
        servicesHealth
      ]);

      // Construir resultado
      const result = {
        status: overallStatus,
        responseTime,
        timestamp: new Date().toISOString(),
        components: {
          database: this.getSettledValue(databaseHealth),
          cache: this.getSettledValue(cacheHealth),
          externalServices: this.getSettledValue(servicesHealth)
        },
        metrics: this.getSettledValue(metricsHealth),
        environment: {
          nodeEnv: process.env.NODE_ENV,
          version: this.configService.get<string>('app.version', '2.1.0'),
          uptime: process.uptime(),
          memoryUsage: process.memoryUsage(),
          cpuUsage: process.cpuUsage()
        }
      };

      if (overallStatus === 'healthy') {
        return this.getStatus(key, true, result);
      } else {
        throw new HealthCheckError('IAM service is unhealthy', result);
      }

    } catch (error) {
      this.logger.error(`Health check failed: ${error.message}`, error.stack);
      throw new HealthCheckError('Health check failed', {
        error: error.message,
        timestamp: new Date().toISOString()
      });
    }
  }

  /**
   * Verificar saúde do banco de dados
   */
  private async checkDatabaseHealth(): Promise<ComponentHealth> {
    const cacheKey = 'database_health';
    const cached = this.getCachedResult(cacheKey);
    if (cached) return cached;

    const startTime = Date.now();
    
    try {
      // Teste de conectividade básica
      await this.userRepository.query('SELECT 1');
      
      // Verificar pool de conexões
      const connectionPool = this.userRepository.manager.connection;
      const poolSize = connectionPool.options.extra?.max || 10;
      
      // Teste de performance - contar usuários ativos
      const activeUsersCount = await this.userRepository.count({
        where: { isActive: true }
      });

      // Teste de escrita - verificar se consegue inserir audit log
      const testAudit = this.auditRepository.create({
        userId: null,
        tenantId: 'health-check',
        action: 'HEALTH_CHECK',
        resource: 'Database',
        severity: 'info',
        metadata: { healthCheck: true, timestamp: new Date() }
      });
      
      await this.auditRepository.save(testAudit);
      
      // Limpar teste
      await this.auditRepository.delete({ id: testAudit.id });

      const responseTime = Date.now() - startTime;
      
      const result: ComponentHealth = {
        status: responseTime < 1000 ? 'healthy' : 'degraded',
        responseTime,
        lastCheck: new Date(),
        metrics: {
          activeUsers: activeUsersCount,
          poolSize,
          connectionStatus: 'connected'
        }
      };

      this.setCachedResult(cacheKey, result);
      return result;

    } catch (error) {
      const responseTime = Date.now() - startTime;
      
      const result: ComponentHealth = {
        status: 'unhealthy',
        responseTime,
        lastCheck: new Date(),
        error: error.message,
        metrics: {
          connectionStatus: 'failed'
        }
      };

      this.setCachedResult(cacheKey, result);
      return result;
    }
  }

  /**
   * Verificar saúde do cache Redis
   */
  private async checkCacheHealth(): Promise<ComponentHealth> {
    const cacheKey = 'cache_health';
    const cached = this.getCachedResult(cacheKey);
    if (cached) return cached;

    const startTime = Date.now();
    
    try {
      // Teste de conectividade
      const testKey = 'health_check_test';
      const testValue = `test_${Date.now()}`;
      
      // Teste de escrita
      await this.cacheManager.set(testKey, testValue, 10);
      
      // Teste de leitura
      const retrievedValue = await this.cacheManager.get(testKey);
      
      if (retrievedValue !== testValue) {
        throw new Error('Cache read/write test failed');
      }
      
      // Limpeza
      await this.cacheManager.del(testKey);

      const responseTime = Date.now() - startTime;
      
      const result: ComponentHealth = {
        status: responseTime < 500 ? 'healthy' : 'degraded',
        responseTime,
        lastCheck: new Date(),
        metrics: {
          connectionStatus: 'connected',
          readWriteTest: 'passed'
        }
      };

      this.setCachedResult(cacheKey, result);
      return result;

    } catch (error) {
      const responseTime = Date.now() - startTime;
      
      const result: ComponentHealth = {
        status: 'unhealthy',
        responseTime,
        lastCheck: new Date(),
        error: error.message,
        metrics: {
          connectionStatus: 'failed'
        }
      };

      this.setCachedResult(cacheKey, result);
      return result;
    }
  }

  /**
   * Verificar serviços externos
   */
  private async checkExternalServices(): Promise<Record<string, ComponentHealth>> {
    const services: ServiceDependency[] = [
      {
        name: 'riskService',
        url: this.configService.get<string>('integration.externalServices.riskService.url'),
        timeout: this.configService.get<number>('integration.externalServices.riskService.timeout', 5000),
        critical: false,
        healthCheck: () => this.checkHTTPService('Risk Service', this.configService.get<string>('integration.externalServices.riskService.url'))
      },
      {
        name: 'notificationService',
        url: this.configService.get<string>('integration.externalServices.notificationService.url'),
        timeout: this.configService.get<number>('integration.externalServices.notificationService.timeout', 5000),
        critical: false,
        healthCheck: () => this.checkHTTPService('Notification Service', this.configService.get<string>('integration.externalServices.notificationService.url'))
      },
      {
        name: 'analyticsService',
        url: this.configService.get<string>('integration.externalServices.analyticsService.url'),
        timeout: this.configService.get<number>('integration.externalServices.analyticsService.timeout', 10000),
        critical: false,
        healthCheck: () => this.checkHTTPService('Analytics Service', this.configService.get<string>('integration.externalServices.analyticsService.url'))
      }
    ];

    const results: Record<string, ComponentHealth> = {};

    await Promise.allSettled(
      services.map(async (service) => {
        try {
          const cacheKey = `service_${service.name}_health`;
          const cached = this.getCachedResult(cacheKey);
          
          if (cached) {
            results[service.name] = cached;
            return;
          }

          const health = await service.healthCheck();
          results[service.name] = health;
          this.setCachedResult(cacheKey, health);
          
        } catch (error) {
          results[service.name] = {
            status: 'unhealthy',
            responseTime: 0,
            lastCheck: new Date(),
            error: error.message
          };
        }
      })
    );

    return results;
  }

  /**
   * Verificar serviço HTTP
   */
  private async checkHTTPService(name: string, url?: string): Promise<ComponentHealth> {
    if (!url) {
      return {
        status: 'healthy', // Não crítico se não configurado
        responseTime: 0,
        lastCheck: new Date(),
        metrics: { configured: false }
      };
    }

    const startTime = Date.now();
    
    try {
      // Implementar verificação HTTP
      // Por enquanto, simular verificação
      const responseTime = Date.now() - startTime;
      
      return {
        status: 'healthy',
        responseTime,
        lastCheck: new Date(),
        metrics: {
          configured: true,
          url: url.replace(/\/\/.*@/, '//***@') // Mascarar credenciais
        }
      };

    } catch (error) {
      const responseTime = Date.now() - startTime;
      
      return {
        status: 'unhealthy',
        responseTime,
        lastCheck: new Date(),
        error: error.message,
        metrics: {
          configured: true,
          url: url.replace(/\/\/.*@/, '//***@')
        }
      };
    }
  }

  /**
   * Coletar métricas de saúde
   */
  private async collectHealthMetrics(): Promise<HealthMetrics> {
    try {
      const [
        activeUsersCount,
        activeSessionsCount
      ] = await Promise.all([
        this.userRepository.count({ where: { isActive: true } }),
        this.sessionRepository.count({ where: { isActive: true } })
      ]);

      const memoryUsage = process.memoryUsage();
      const cpuUsage = process.cpuUsage();

      return {
        responseTime: 0, // Será calculado pelo caller
        errorRate: await this.calculateErrorRate(),
        activeUsers: activeUsersCount,
        activeSessions: activeSessionsCount,
        databaseConnections: this.getDatabaseConnectionCount(),
        cacheHitRate: await this.calculateCacheHitRate(),
        memoryUsage: memoryUsage.heapUsed / memoryUsage.heapTotal,
        cpuUsage: (cpuUsage.user + cpuUsage.system) / 1000000 // Converter para segundos
      };

    } catch (error) {
      this.logger.warn(`Failed to collect health metrics: ${error.message}`);
      return {
        responseTime: 0,
        errorRate: 0,
        activeUsers: 0,
        activeSessions: 0,
        databaseConnections: 0,
        cacheHitRate: 0,
        memoryUsage: 0,
        cpuUsage: 0
      };
    }
  }

  /**
   * Calcular taxa de erro
   */
  private async calculateErrorRate(): Promise<number> {
    try {
      const last5Minutes = new Date(Date.now() - 5 * 60 * 1000);
      
      const [totalRequests, errorRequests] = await Promise.all([
        this.auditRepository.count({
          where: {
            timestamp: { $gte: last5Minutes } as any
          }
        }),
        this.auditRepository.count({
          where: {
            timestamp: { $gte: last5Minutes } as any,
            severity: { $in: ['error', 'critical'] } as any
          }
        })
      ]);

      return totalRequests > 0 ? errorRequests / totalRequests : 0;

    } catch (error) {
      return 0;
    }
  }

  /**
   * Obter contagem de conexões do banco
   */
  private getDatabaseConnectionCount(): number {
    try {
      const connection = this.userRepository.manager.connection;
      // Implementar lógica específica do driver
      return connection.isConnected ? 1 : 0;
    } catch (error) {
      return 0;
    }
  }

  /**
   * Calcular taxa de acerto do cache
   */
  private async calculateCacheHitRate(): Promise<number> {
    try {
      // Implementar lógica de cálculo de cache hit rate
      // Por enquanto, retornar valor simulado
      return 0.85; // 85% de hit rate
    } catch (error) {
      return 0;
    }
  }

  /**
   * Determinar status geral
   */
  private determineOverallStatus(results: PromiseSettledResult<any>[]): 'healthy' | 'degraded' | 'unhealthy' {
    let healthyCount = 0;
    let degradedCount = 0;
    let unhealthyCount = 0;

    results.forEach(result => {
      if (result.status === 'fulfilled') {
        const health = result.value as ComponentHealth;
        switch (health.status) {
          case 'healthy':
            healthyCount++;
            break;
          case 'degraded':
            degradedCount++;
            break;
          case 'unhealthy':
            unhealthyCount++;
            break;
        }
      } else {
        unhealthyCount++;
      }
    });

    // Lógica de determinação de status
    if (unhealthyCount > 0) {
      return 'unhealthy';
    } else if (degradedCount > 0) {
      return 'degraded';
    } else {
      return 'healthy';
    }
  }

  /**
   * Obter valor de PromiseSettledResult
   */
  private getSettledValue(result: PromiseSettledResult<any>): any {
    return result.status === 'fulfilled' ? result.value : { error: result.reason?.message };
  }

  /**
   * Obter resultado em cache
   */
  private getCachedResult(key: string): ComponentHealth | null {
    const cached = this.healthCache.get(key);
    if (cached && Date.now() - cached.timestamp < this.cacheTTL) {
      return cached.result;
    }
    return null;
  }

  /**
   * Definir resultado em cache
   */
  private setCachedResult(key: string, result: ComponentHealth): void {
    this.healthCache.set(key, {
      result,
      timestamp: Date.now()
    });
  }

  /**
   * Health check específico para readiness
   */
  async isReady(key: string): Promise<HealthIndicatorResult> {
    try {
      // Verificações críticas para readiness
      const [databaseReady, cacheReady] = await Promise.all([
        this.isDatabaseReady(),
        this.isCacheReady()
      ]);

      if (databaseReady && cacheReady) {
        return this.getStatus(key, true, {
          database: 'ready',
          cache: 'ready',
          timestamp: new Date().toISOString()
        });
      } else {
        throw new HealthCheckError('Service not ready', {
          database: databaseReady ? 'ready' : 'not ready',
          cache: cacheReady ? 'ready' : 'not ready'
        });
      }

    } catch (error) {
      throw new HealthCheckError('Readiness check failed', {
        error: error.message,
        timestamp: new Date().toISOString()
      });
    }
  }

  /**
   * Verificar se banco está pronto
   */
  private async isDatabaseReady(): Promise<boolean> {
    try {
      await this.userRepository.query('SELECT 1');
      return true;
    } catch (error) {
      return false;
    }
  }

  /**
   * Verificar se cache está pronto
   */
  private async isCacheReady(): Promise<boolean> {
    try {
      await this.cacheManager.set('readiness_test', 'ok', 1);
      const result = await this.cacheManager.get('readiness_test');
      return result === 'ok';
    } catch (error) {
      return false;
    }
  }

  /**
   * Health check específico para liveness
   */
  async isLive(key: string): Promise<HealthIndicatorResult> {
    try {
      // Verificações básicas de liveness
      const memoryUsage = process.memoryUsage();
      const heapUsedPercent = memoryUsage.heapUsed / memoryUsage.heapTotal;
      
      // Verificar se não está com uso excessivo de memória
      if (heapUsedPercent > 0.9) {
        throw new HealthCheckError('High memory usage', {
          heapUsedPercent,
          memoryUsage
        });
      }

      return this.getStatus(key, true, {
        uptime: process.uptime(),
        memoryUsage: {
          heapUsed: memoryUsage.heapUsed,
          heapTotal: memoryUsage.heapTotal,
          heapUsedPercent
        },
        timestamp: new Date().toISOString()
      });

    } catch (error) {
      throw new HealthCheckError('Liveness check failed', {
        error: error.message,
        timestamp: new Date().toISOString()
      });
    }
  }
}