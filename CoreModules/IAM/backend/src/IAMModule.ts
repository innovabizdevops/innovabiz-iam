/**
 * 🏗️ IAM MODULE - INNOVABIZ PLATFORM
 * Módulo principal do sistema de gestão de identidade e acesso
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: NestJS Best Practices, Dependency Injection, Modular Architecture
 * Standards: SOLID Principles, Clean Architecture, Domain-Driven Design
 */

import { Module, Global } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { JwtModule } from '@nestjs/jwt';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { CacheModule } from '@nestjs/cache-manager';
import { ThrottlerModule } from '@nestjs/throttler';
import { PrometheusModule } from '@willsoto/nestjs-prometheus';
import { HealthModule } from '@nestjs/terminus';

// Entities
import { User } from './models/User.entity';
import { Session } from './models/Session.entity';
import { Credential } from './models/Credential.entity';
import { AuditLog } from './models/AuditLog.entity';
import { RiskProfile } from './models/RiskProfile.entity';
import { RiskEvent } from './models/RiskEvent.entity';

// Services
import { IAMService } from './services/IAMService';
import { WebAuthnService } from './services/WebAuthnService';
import { CredentialService } from './services/CredentialService';
import { RiskAssessmentService } from './services/RiskAssessmentService';
import { AuditService } from './services/AuditService';
import { AttestationService } from './services/AttestationService';

// Controllers
import { IAMController } from './controllers/IAMController';

// Guards
import { JwtAuthGuard } from './middleware/JwtAuthGuard';
import { RateLimitGuard } from './middleware/RateLimitGuard';
import { TenantGuard } from './middleware/TenantGuard';

// Interceptors
import { SecurityHeadersInterceptor } from './middleware/SecurityHeadersInterceptor';
import { AuditInterceptor } from './middleware/AuditInterceptor';
import { MetricsInterceptor } from './middleware/MetricsInterceptor';

// Strategies
import { JwtStrategy } from './strategies/JwtStrategy';

// Health Indicators
import { IAMHealthIndicator } from './health/IAMHealthIndicator';

// Metrics
import { IAMMetricsService } from './metrics/IAMMetricsService';

// Configuration
import { iamConfig } from './config/iam.config';
import { webauthnConfig } from './config/webauthn.config';
import { riskConfig } from './config/risk.config';

@Global()
@Module({
  imports: [
    // Configuration
    ConfigModule.forRoot({
      load: [iamConfig, webauthnConfig, riskConfig],
      isGlobal: true,
      cache: true,
      expandVariables: true
    }),

    // Database
    TypeOrmModule.forFeature([
      User,
      Session,
      Credential,
      AuditLog,
      RiskProfile,
      RiskEvent
    ]),

    // JWT Configuration
    JwtModule.registerAsync({
      imports: [ConfigModule],
      useFactory: async (configService: ConfigService) => ({
        secret: configService.get<string>('jwt.secret'),
        signOptions: {
          expiresIn: configService.get<string>('jwt.expiresIn', '1h'),
          issuer: configService.get<string>('jwt.issuer', 'innovabiz-iam'),
          audience: configService.get<string>('jwt.audience', 'innovabiz-platform')
        },
        verifyOptions: {
          issuer: configService.get<string>('jwt.issuer', 'innovabiz-iam'),
          audience: configService.get<string>('jwt.audience', 'innovabiz-platform')
        }
      }),
      inject: [ConfigService]
    }),

    // Cache Configuration
    CacheModule.registerAsync({
      imports: [ConfigModule],
      useFactory: async (configService: ConfigService) => ({
        store: 'redis',
        host: configService.get<string>('redis.host', 'localhost'),
        port: configService.get<number>('redis.port', 6379),
        password: configService.get<string>('redis.password'),
        db: configService.get<number>('redis.db', 0),
        ttl: configService.get<number>('cache.ttl', 300), // 5 minutes
        max: configService.get<number>('cache.max', 1000),
        keyPrefix: 'iam:',
        retryDelayOnFailover: 100,
        enableReadyCheck: true,
        maxRetriesPerRequest: 3
      }),
      inject: [ConfigService]
    }),

    // Rate Limiting
    ThrottlerModule.forRootAsync({
      imports: [ConfigModule],
      useFactory: async (configService: ConfigService) => ({
        ttl: configService.get<number>('throttle.ttl', 60),
        limit: configService.get<number>('throttle.limit', 100),
        storage: 'redis',
        redis: {
          host: configService.get<string>('redis.host', 'localhost'),
          port: configService.get<number>('redis.port', 6379),
          password: configService.get<string>('redis.password'),
          db: configService.get<number>('redis.throttle_db', 1)
        }
      }),
      inject: [ConfigService]
    }),

    // Metrics
    PrometheusModule.register({
      path: '/metrics',
      defaultMetrics: {
        enabled: true,
        config: {
          prefix: 'innovabiz_iam_'
        }
      }
    }),

    // Health Checks
    HealthModule
  ],

  controllers: [
    IAMController
  ],

  providers: [
    // Core Services
    IAMService,
    WebAuthnService,
    CredentialService,
    RiskAssessmentService,
    AuditService,
    AttestationService,

    // Authentication Strategy
    JwtStrategy,

    // Guards
    JwtAuthGuard,
    RateLimitGuard,
    TenantGuard,

    // Interceptors
    SecurityHeadersInterceptor,
    AuditInterceptor,
    MetricsInterceptor,

    // Health Indicators
    IAMHealthIndicator,

    // Metrics Service
    IAMMetricsService,

    // Configuration Providers
    {
      provide: 'IAM_CONFIG',
      useFactory: (configService: ConfigService) => configService.get('iam'),
      inject: [ConfigService]
    },
    {
      provide: 'WEBAUTHN_CONFIG',
      useFactory: (configService: ConfigService) => configService.get('webauthn'),
      inject: [ConfigService]
    },
    {
      provide: 'RISK_CONFIG',
      useFactory: (configService: ConfigService) => configService.get('risk'),
      inject: [ConfigService]
    }
  ],

  exports: [
    // Export services for use in other modules
    IAMService,
    WebAuthnService,
    CredentialService,
    RiskAssessmentService,
    AuditService,
    AttestationService,
    
    // Export guards for use in other modules
    JwtAuthGuard,
    RateLimitGuard,
    TenantGuard,
    
    // Export interceptors
    SecurityHeadersInterceptor,
    AuditInterceptor,
    MetricsInterceptor,
    
    // Export strategy
    JwtStrategy,
    
    // Export metrics service
    IAMMetricsService,
    
    // Export TypeORM repositories
    TypeOrmModule
  ]
})
export class IAMModule {
  constructor(
    private readonly configService: ConfigService,
    private readonly iamMetricsService: IAMMetricsService
  ) {
    this.initializeModule();
  }

  /**
   * Inicializar configurações do módulo
   */
  private initializeModule(): void {
    // Configurar métricas personalizadas
    this.setupCustomMetrics();
    
    // Configurar logs estruturados
    this.setupStructuredLogging();
    
    // Validar configurações críticas
    this.validateCriticalConfigurations();
  }

  /**
   * Configurar métricas personalizadas do Prometheus
   */
  private setupCustomMetrics(): void {
    this.iamMetricsService.initializeMetrics();
  }

  /**
   * Configurar logs estruturados
   */
  private setupStructuredLogging(): void {
    // Configurar formato de logs para compliance
    const logLevel = this.configService.get<string>('LOG_LEVEL', 'info');
    const environment = this.configService.get<string>('NODE_ENV', 'development');
    
    console.log(`IAM Module initialized - Environment: ${environment}, Log Level: ${logLevel}`);
  }

  /**
   * Validar configurações críticas
   */
  private validateCriticalConfigurations(): void {
    const requiredConfigs = [
      'jwt.secret',
      'webauthn.rpID',
      'webauthn.origin',
      'database.host',
      'redis.host'
    ];

    const missingConfigs = requiredConfigs.filter(config => 
      !this.configService.get(config)
    );

    if (missingConfigs.length > 0) {
      throw new Error(`Missing critical configurations: ${missingConfigs.join(', ')}`);
    }

    // Validar configurações de segurança
    this.validateSecurityConfigurations();
  }

  /**
   * Validar configurações de segurança
   */
  private validateSecurityConfigurations(): void {
    const jwtSecret = this.configService.get<string>('jwt.secret');
    if (jwtSecret && jwtSecret.length < 32) {
      throw new Error('JWT secret must be at least 32 characters long');
    }

    const environment = this.configService.get<string>('NODE_ENV');
    if (environment === 'production') {
      const httpsOnly = this.configService.get<boolean>('security.httpsOnly', true);
      if (!httpsOnly) {
        console.warn('WARNING: HTTPS is not enforced in production environment');
      }
    }

    // Validar configurações WebAuthn
    const rpID = this.configService.get<string>('webauthn.rpID');
    const origin = this.configService.get<string>('webauthn.origin');
    
    if (environment === 'production' && (!rpID || !origin)) {
      throw new Error('WebAuthn RP ID and Origin must be configured for production');
    }

    if (origin && !origin.startsWith('https://') && environment === 'production') {
      throw new Error('WebAuthn origin must use HTTPS in production');
    }
  }

  /**
   * Método para configuração dinâmica (se necessário)
   */
  static forRoot(options?: {
    database?: any;
    jwt?: any;
    webauthn?: any;
    risk?: any;
    cache?: any;
  }) {
    return {
      module: IAMModule,
      providers: [
        ...(options?.database ? [{
          provide: 'DATABASE_OPTIONS',
          useValue: options.database
        }] : []),
        ...(options?.jwt ? [{
          provide: 'JWT_OPTIONS',
          useValue: options.jwt
        }] : []),
        ...(options?.webauthn ? [{
          provide: 'WEBAUTHN_OPTIONS',
          useValue: options.webauthn
        }] : []),
        ...(options?.risk ? [{
          provide: 'RISK_OPTIONS',
          useValue: options.risk
        }] : []),
        ...(options?.cache ? [{
          provide: 'CACHE_OPTIONS',
          useValue: options.cache
        }] : [])
      ]
    };
  }

  /**
   * Método para configuração assíncrona
   */
  static forRootAsync(options: {
    imports?: any[];
    useFactory?: (...args: any[]) => Promise<any> | any;
    inject?: any[];
  }) {
    return {
      module: IAMModule,
      imports: options.imports || [],
      providers: [
        {
          provide: 'IAM_MODULE_OPTIONS',
          useFactory: options.useFactory,
          inject: options.inject || []
        }
      ]
    };
  }
}