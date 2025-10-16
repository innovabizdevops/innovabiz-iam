/**
 * ============================================================================
 * INNOVABIZ IAM - Express Application
 * ============================================================================
 * Version: 1.0.0
 * Date: 31/07/2025
 * Author: Equipe de Segurança INNOVABIZ
 * Classification: Confidencial - Interno
 * 
 * Description: Aplicação Express principal para WebAuthn/FIDO2 IAM
 * Standards: W3C WebAuthn Level 3, FIDO2 CTAP2.1, NIST SP 800-63B
 * Compliance: PCI DSS 4.0, GDPR/LGPD, PSD2, ISO 27001
 * ============================================================================
 */

import express, { Application, Request, Response, NextFunction } from 'express';
import cors from 'cors';
import helmet from 'helmet';
import compression from 'compression';
import rateLimit from 'express-rate-limit';
import { Pool } from 'pg';
import { Redis } from 'ioredis';
import { Kafka } from 'kafkajs';
import winston from 'winston';
import promClient from 'prom-client';

// Serviços
import { WebAuthnService } from './services/WebAuthnService';
import { CredentialService } from './services/CredentialService';
import { RiskAssessmentService } from './services/RiskAssessmentService';
import { AuditService } from './services/AuditService';
import { AttestationService } from './services/AttestationService';

// Controladores
import { WebAuthnController } from './controllers/WebAuthnController';

// Configurações
import { webauthnConfig } from './config/webauthn';
import { webauthnMetrics } from './metrics/webauthn';

// Tipos
import { WebAuthnAPIResponse } from './types/webauthn';

/**
 * Classe principal da aplicação
 */
export class WebAuthnApp {
  private readonly app: Application;
  private readonly logger: Logger;
  private readonly db: Pool;
  private readonly redis: Redis;
  private readonly kafka?: Kafka;
  
  // Serviços
  private webauthnService: WebAuthnService;
  private credentialService: CredentialService;
  private riskService: RiskAssessmentService;
  private auditService: AuditService;
  private attestationService: AttestationService;
  
  // Controladores
  private webauthnController: WebAuthnController;

  constructor(
    logger: Logger,
    db: Pool,
    redis: Redis,
    kafka?: Kafka
  ) {
    this.app = express();
    this.logger = logger;
    this.db = db;
    this.redis = redis;
    this.kafka = kafka;

    this.initializeServices();
    this.initializeMiddlewares();
    this.initializeRoutes();
    this.initializeErrorHandling();
  }

  /**
   * Inicializa os serviços
   */
  private initializeServices(): void {
    this.logger.info('Initializing WebAuthn services');

    // Serviços base
    this.credentialService = new CredentialService(this.logger, this.db, this.redis);
    this.riskService = new RiskAssessmentService(this.logger, this.db, this.redis);
    this.auditService = new AuditService(this.logger, this.db, this.kafka);
    this.attestationService = new AttestationService(this.logger, this.db, this.redis);

    // Serviço principal WebAuthn
    this.webauthnService = new WebAuthnService(
      this.logger,
      this.redis,
      this.db,
      this.credentialService,
      this.attestationService,
      this.riskService,
      this.auditService
    );

    // Controladores
    this.webauthnController = new WebAuthnController(
      this.logger,
      this.webauthnService,
      this.credentialService
    );

    this.logger.info('WebAuthn services initialized successfully');
  }

  /**
   * Inicializa middlewares
   */
  private initializeMiddlewares(): void {
    this.logger.info('Initializing middlewares');

    // Segurança básica
    this.app.use(helmet({
      contentSecurityPolicy: {
        directives: {
          defaultSrc: ["'self'"],
          scriptSrc: ["'self'", "'unsafe-inline'"],
          styleSrc: ["'self'", "'unsafe-inline'"],
          imgSrc: ["'self'", "data:", "https:"],
          connectSrc: ["'self'"],
          fontSrc: ["'self'"],
          objectSrc: ["'none'"],
          mediaSrc: ["'self'"],
          frameSrc: ["'none'"]
        }
      },
      crossOriginEmbedderPolicy: false
    }));

    // CORS
    this.app.use(cors({
      origin: webauthnConfig.origin,
      credentials: true,
      methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
      allowedHeaders: [
        'Content-Type',
        'Authorization',
        'X-Correlation-ID',
        'X-Tenant-ID',
        'X-Region',
        'X-Device-Fingerprint'
      ]
    }));

    // Compressão
    this.app.use(compression());

    // Parsing
    this.app.use(express.json({ limit: '1mb' }));
    this.app.use(express.urlencoded({ extended: true, limit: '1mb' }));

    // Rate limiting
    this.setupRateLimiting();

    // Logging de requisições
    this.app.use(this.requestLoggingMiddleware);

    // Correlation ID
    this.app.use(this.correlationIdMiddleware);

    // Métricas
    this.app.use(this.metricsMiddleware);

    this.logger.info('Middlewares initialized successfully');
  }

  /**
   * Configura rate limiting
   */
  private setupRateLimiting(): void {
    // Rate limiting geral
    const generalLimiter = rateLimit({
      windowMs: 15 * 60 * 1000, // 15 minutos
      max: 100, // 100 requests por IP
      message: {
        success: false,
        error: {
          code: 'RATE_LIMIT_EXCEEDED',
          message: 'Too many requests from this IP'
        }
      },
      standardHeaders: true,
      legacyHeaders: false,
      handler: (req, res) => {
        webauthnMetrics.rateLimitHits.inc({
          tenant_id: req.headers['x-tenant-id'] as string || 'unknown',
          limit_type: 'general',
          identifier: req.ip
        });
        
        res.status(429).json({
          success: false,
          error: {
            code: 'RATE_LIMIT_EXCEEDED',
            message: 'Too many requests from this IP'
          },
          metadata: {
            correlationId: res.locals.correlationId,
            timestamp: new Date().toISOString(),
            version: '1.0.0'
          }
        });
      }
    });

    // Rate limiting para registro
    const registrationLimiter = rateLimit({
      windowMs: webauthnConfig.rateLimiting.windowMinutes * 60 * 1000,
      max: webauthnConfig.rateLimiting.registrationPerIP,
      keyGenerator: (req) => `registration:${req.ip}`,
      message: {
        success: false,
        error: {
          code: 'REGISTRATION_RATE_LIMIT_EXCEEDED',
          message: 'Too many registration attempts'
        }
      }
    });

    // Rate limiting para autenticação
    const authenticationLimiter = rateLimit({
      windowMs: webauthnConfig.rateLimiting.windowMinutes * 60 * 1000,
      max: webauthnConfig.rateLimiting.authenticationPerIP,
      keyGenerator: (req) => `authentication:${req.ip}`,
      message: {
        success: false,
        error: {
          code: 'AUTHENTICATION_RATE_LIMIT_EXCEEDED',
          message: 'Too many authentication attempts'
        }
      }
    });

    this.app.use(generalLimiter);
    this.app.use('/api/v1/webauthn/registration', registrationLimiter);
    this.app.use('/api/v1/webauthn/authentication', authenticationLimiter);
  }

  /**
   * Middleware de logging de requisições
   */
  private requestLoggingMiddleware = (req: Request, res: Response, next: NextFunction): void => {
    const start = Date.now();
    
    res.on('finish', () => {
      const duration = Date.now() - start;
      
      this.logger.info('HTTP Request', {
        method: req.method,
        url: req.url,
        statusCode: res.statusCode,
        duration,
        userAgent: req.headers['user-agent'],
        ip: req.ip,
        correlationId: res.locals.correlationId,
        tenantId: req.headers['x-tenant-id']
      });
    });

    next();
  };

  /**
   * Middleware de correlation ID
   */
  private correlationIdMiddleware = (req: Request, res: Response, next: NextFunction): void => {
    const correlationId = req.headers['x-correlation-id'] as string || 
                         `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    
    res.locals.correlationId = correlationId;
    res.setHeader('X-Correlation-ID', correlationId);
    
    next();
  };

  /**
   * Middleware de métricas
   */
  private metricsMiddleware = (req: Request, res: Response, next: NextFunction): void => {
    const start = Date.now();
    
    res.on('finish', () => {
      const duration = (Date.now() - start) / 1000;
      
      webauthnMetrics.httpRequestsTotal.inc({
        method: req.method,
        endpoint: this.normalizeEndpoint(req.route?.path || req.path),
        status_code: res.statusCode.toString(),
        tenant_id: req.headers['x-tenant-id'] as string || 'unknown'
      });

      webauthnMetrics.httpRequestDuration.observe(
        {
          method: req.method,
          endpoint: this.normalizeEndpoint(req.route?.path || req.path),
          status_code: res.statusCode.toString()
        },
        duration
      );

      if (req.headers['content-length']) {
        webauthnMetrics.httpRequestSize.observe(
          {
            method: req.method,
            endpoint: this.normalizeEndpoint(req.route?.path || req.path)
          },
          parseInt(req.headers['content-length'] as string)
        );
      }

      const responseSize = res.get('content-length');
      if (responseSize) {
        webauthnMetrics.httpResponseSize.observe(
          {
            method: req.method,
            endpoint: this.normalizeEndpoint(req.route?.path || req.path),
            status_code: res.statusCode.toString()
          },
          parseInt(responseSize)
        );
      }
    });

    next();
  };

  /**
   * Normaliza endpoint para métricas
   */
  private normalizeEndpoint(path: string): string {
    return path
      .replace(/\/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/gi, '/:id')
      .replace(/\/[0-9]+/g, '/:id')
      .replace(/\/[a-zA-Z0-9_-]{20,}/g, '/:id');
  }

  /**
   * Inicializa rotas
   */
  private initializeRoutes(): void {
    this.logger.info('Initializing routes');

    // Health check
    this.app.get('/health', this.healthCheckHandler);
    this.app.get('/health/webauthn', this.webauthnHealthCheckHandler);

    // Métricas Prometheus
    this.app.get('/metrics', this.metricsHandler);

    // Rotas WebAuthn
    const webauthnRouter = express.Router();
    
    // Registro
    webauthnRouter.post('/registration/options', this.webauthnController.generateRegistrationOptions);
    webauthnRouter.post('/registration/verify', this.webauthnController.verifyRegistration);
    
    // Autenticação
    webauthnRouter.post('/authentication/options', this.webauthnController.generateAuthenticationOptions);
    webauthnRouter.post('/authentication/verify', this.webauthnController.verifyAuthentication);
    
    // Gerenciamento de credenciais
    webauthnRouter.get('/credentials', this.webauthnController.getUserCredentials);
    webauthnRouter.put('/credentials/:credentialId/name', this.webauthnController.updateCredentialName);
    webauthnRouter.delete('/credentials/:credentialId', this.webauthnController.deleteCredential);
    
    // Estatísticas
    webauthnRouter.get('/stats', this.webauthnController.getCredentialStats);

    this.app.use('/api/v1/webauthn', webauthnRouter);

    // Rota 404
    this.app.use('*', this.notFoundHandler);

    this.logger.info('Routes initialized successfully');
  }

  /**
   * Handler de health check
   */
  private healthCheckHandler = async (req: Request, res: Response): Promise<void> => {
    try {
      // Verificar banco de dados
      await this.db.query('SELECT 1');
      
      // Verificar Redis
      await this.redis.ping();

      res.json({
        status: 'healthy',
        timestamp: new Date().toISOString(),
        version: '1.0.0',
        services: {
          database: 'healthy',
          redis: 'healthy',
          webauthn: 'healthy'
        }
      });

    } catch (error) {
      this.logger.error('Health check failed', { error: error.message });
      
      res.status(503).json({
        status: 'unhealthy',
        timestamp: new Date().toISOString(),
        version: '1.0.0',
        error: error.message
      });
    }
  };

  /**
   * Handler de health check específico do WebAuthn
   */
  private webauthnHealthCheckHandler = async (req: Request, res: Response): Promise<void> => {
    try {
      // Verificar configuração WebAuthn
      const configValid = webauthnConfig.rpID && webauthnConfig.origin.length > 0;
      
      // Verificar métricas
      const metricsAvailable = webauthnMetrics.systemHealth !== undefined;

      if (configValid && metricsAvailable) {
        webauthnMetrics.updateSystemHealth('webauthn', true);
        
        res.json({
          status: 'healthy',
          timestamp: new Date().toISOString(),
          version: '1.0.0',
          webauthn: {
            rpID: webauthnConfig.rpID,
            originsConfigured: webauthnConfig.origin.length,
            metricsEnabled: webauthnConfig.observability.metrics.enabled
          }
        });
      } else {
        throw new Error('WebAuthn configuration invalid');
      }

    } catch (error) {
      webauthnMetrics.updateSystemHealth('webauthn', false);
      
      this.logger.error('WebAuthn health check failed', { error: error.message });
      
      res.status(503).json({
        status: 'unhealthy',
        timestamp: new Date().toISOString(),
        version: '1.0.0',
        error: error.message
      });
    }
  };

  /**
   * Handler de métricas Prometheus
   */
  private metricsHandler = async (req: Request, res: Response): Promise<void> => {
    try {
      res.set('Content-Type', promClient.register.contentType);
      res.end(await promClient.register.metrics());
    } catch (error) {
      this.logger.error('Failed to generate metrics', { error: error.message });
      res.status(500).end();
    }
  };

  /**
   * Handler 404
   */
  private notFoundHandler = (req: Request, res: Response): void => {
    const response: WebAuthnAPIResponse = {
      success: false,
      error: {
        code: 'NOT_FOUND',
        message: `Endpoint ${req.method} ${req.path} not found`
      },
      metadata: {
        correlationId: res.locals.correlationId,
        timestamp: new Date().toISOString(),
        version: '1.0.0'
      }
    };

    res.status(404).json(response);
  };

  /**
   * Inicializa tratamento de erros
   */
  private initializeErrorHandling(): void {
    // Handler de erro global
    this.app.use((error: Error, req: Request, res: Response, next: NextFunction) => {
      this.logger.error('Unhandled error', {
        error: error.message,
        stack: error.stack,
        url: req.url,
        method: req.method,
        correlationId: res.locals.correlationId
      });

      const response: WebAuthnAPIResponse = {
        success: false,
        error: {
          code: 'INTERNAL_ERROR',
          message: 'Internal server error'
        },
        metadata: {
          correlationId: res.locals.correlationId,
          timestamp: new Date().toISOString(),
          version: '1.0.0'
        }
      };

      res.status(500).json(response);
    });

    // Handler de processo não capturado
    process.on('uncaughtException', (error) => {
      this.logger.error('Uncaught exception', {
        error: error.message,
        stack: error.stack
      });
      
      // Graceful shutdown
      this.shutdown();
    });

    process.on('unhandledRejection', (reason, promise) => {
      this.logger.error('Unhandled rejection', {
        reason,
        promise
      });
    });

    this.logger.info('Error handling initialized successfully');
  }

  /**
   * Retorna a instância do Express
   */
  public getApp(): Application {
    return this.app;
  }

  /**
   * Inicia o servidor
   */
  public async start(port: number = 3000): Promise<void> {
    return new Promise((resolve) => {
      const server = this.app.listen(port, () => {
        this.logger.info(`WebAuthn server started on port ${port}`);
        resolve();
      });

      // Graceful shutdown
      process.on('SIGTERM', () => {
        this.logger.info('SIGTERM received, shutting down gracefully');
        server.close(() => {
          this.shutdown();
        });
      });

      process.on('SIGINT', () => {
        this.logger.info('SIGINT received, shutting down gracefully');
        server.close(() => {
          this.shutdown();
        });
      });
    });
  }

  /**
   * Shutdown graceful
   */
  private async shutdown(): Promise<void> {
    try {
      this.logger.info('Starting graceful shutdown');

      // Fechar conexões
      await this.auditService.disconnect();
      await this.redis.disconnect();
      await this.db.end();

      this.logger.info('Graceful shutdown completed');
      process.exit(0);
    } catch (error) {
      this.logger.error('Error during shutdown', { error: error.message });
      process.exit(1);
    }
  }
}