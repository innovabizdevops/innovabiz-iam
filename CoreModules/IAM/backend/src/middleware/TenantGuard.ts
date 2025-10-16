/**
 * 🏢 TENANT GUARD - INNOVABIZ IAM
 * Guard para isolamento e validação de tenant multi-dimensional
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: GDPR, LGPD, SOX 404, ISO 27001, Multi-tenancy Security
 * Mercados: Angola (BNA), Europa, América, China, BRICS, Brasil
 * Arquitetura: Multi-tenant, Multi-layer, Multi-context, Multi-dimensional
 */

import {
  Injectable,
  CanActivate,
  ExecutionContext,
  ForbiddenException,
  UnauthorizedException,
  BadRequestException,
  Logger,
  Inject
} from '@nestjs/common';
import { Reflector } from '@nestjs/core';
import { ConfigService } from '@nestjs/config';
import { Cache } from 'cache-manager';
import { CACHE_MANAGER } from '@nestjs/cache-manager';
import { Request } from 'express';

/**
 * Configurações do Tenant Guard
 */
export interface TenantGuardConfig {
  // Estratégias de identificação do tenant
  identificationStrategies: {
    header: boolean;           // X-Tenant-ID header
    subdomain: boolean;        // tenant.domain.com
    path: boolean;            // /tenant/{id}/...
    jwt: boolean;             // tenant no JWT payload
    query: boolean;           // ?tenant=id
    cookie: boolean;          // tenant cookie
  };
  
  // Validações de tenant
  validation: {
    required: boolean;         // Tenant obrigatório
    strict: boolean;          // Validação estrita
    allowSuperTenant: boolean; // Permitir super tenant
    validateStatus: boolean;   // Validar status ativo
    validateLicense: boolean;  // Validar licença
    validateJurisdiction: boolean; // Validar jurisdição
  };
  
  // Isolamento de dados
  dataIsolation: {
    enforceRowLevelSecurity: boolean;
    validateDataAccess: boolean;
    auditCrossTenatAccess: boolean;
    blockCrossTenatQueries: boolean;
  };
  
  // Cache e performance
  cacheConfig: {
    enabled: boolean;
    ttl: number; // em segundos
    keyPrefix: string;
    invalidateOnUpdate: boolean;
  };
  
  // Auditoria e compliance
  auditConfig: {
    logAllAccess: boolean;
    logViolations: boolean;
    logCrossTenatAttempts: boolean;
    includeRequestDetails: boolean;
    complianceMode: 'basic' | 'gdpr' | 'lgpd' | 'sox' | 'full';
  };
  
  // Configurações específicas por jurisdição
  jurisdictionConfig: {
    [key: string]: {
      dataResidency: string[];
      complianceFrameworks: string[];
      auditRetention: number;
      encryptionRequired: boolean;
    };
  };
}

/**
 * Contexto do tenant enriquecido
 */
export interface EnrichedTenantContext {
  // Identificação básica
  id: string;
  name: string;
  code: string;
  
  // Status e configuração
  status: 'active' | 'inactive' | 'suspended' | 'pending';
  type: 'enterprise' | 'business' | 'individual' | 'trial';
  tier: 'basic' | 'standard' | 'premium' | 'enterprise';
  
  // Configurações de negócio
  settings: {
    timezone: string;
    locale: string;
    currency: string;
    dateFormat: string;
    numberFormat: string;
  };
  
  // Contexto regulatório
  regulatory: {
    jurisdiction: string;
    primaryRegulator: string;
    complianceFrameworks: string[];
    dataResidency: string[];
    licenseNumber?: string;
    licenseType?: string;
    licenseExpiry?: Date;
  };
  
  // Contexto de segurança
  security: {
    encryptionLevel: 'basic' | 'standard' | 'high' | 'maximum';
    mfaRequired: boolean;
    sessionTimeout: number;
    ipWhitelist?: string[];
    allowedCountries?: string[];
    blockedCountries?: string[];
  };
  
  // Limites e quotas
  limits: {
    maxUsers: number;
    maxSessions: number;
    maxApiCalls: number;
    maxStorage: number;
    maxTransactions?: number;
  };
  
  // Contexto de auditoria
  audit: {
    retentionPeriod: number;
    complianceLevel: string;
    auditFrequency: string;
    lastAudit?: Date;
    nextAudit?: Date;
  };
  
  // Metadados
  metadata: {
    createdAt: Date;
    updatedAt: Date;
    lastAccess?: Date;
    version: string;
    tags?: string[];
    customFields?: Record<string, any>;
  };
  
  // Contexto de integração
  integration: {
    externalId?: string;
    parentTenant?: string;
    childTenants?: string[];
    partnerships?: string[];
    apiKeys?: string[];
  };
}

/**
 * Resultado da validação do tenant
 */
export interface TenantValidationResult {
  isValid: boolean;
  tenant?: EnrichedTenantContext;
  errors: string[];
  warnings: string[];
  
  // Contexto da validação
  validationId: string;
  timestamp: Date;
  processingTime: number;
  
  // Detalhes da validação
  validationDetails: {
    identificationMethod: string;
    tenantId: string;
    ipAddress: string;
    userAgent: string;
    endpoint: string;
    userId?: string;
    sessionId?: string;
  };
  
  // Compliance e auditoria
  compliance: {
    frameworks: string[];
    violations: string[];
    requirements: string[];
    recommendations: string[];
  };
}

/**
 * Configurações padrão do Tenant Guard
 */
const DEFAULT_CONFIG: TenantGuardConfig = {
  identificationStrategies: {
    header: true,
    subdomain: true,
    path: true,
    jwt: true,
    query: false,
    cookie: false
  },
  validation: {
    required: true,
    strict: true,
    allowSuperTenant: false,
    validateStatus: true,
    validateLicense: true,
    validateJurisdiction: true
  },
  dataIsolation: {
    enforceRowLevelSecurity: true,
    validateDataAccess: true,
    auditCrossTenatAccess: true,
    blockCrossTenatQueries: true
  },
  cacheConfig: {
    enabled: true,
    ttl: 300, // 5 minutos
    keyPrefix: 'tenant_guard:',
    invalidateOnUpdate: true
  },
  auditConfig: {
    logAllAccess: false,
    logViolations: true,
    logCrossTenatAttempts: true,
    includeRequestDetails: true,
    complianceMode: 'full'
  },
  jurisdictionConfig: {
    'angola': {
      dataResidency: ['AO', 'SADC'],
      complianceFrameworks: ['BNA', 'SADC', 'GDPR'],
      auditRetention: 2555, // 7 anos em dias
      encryptionRequired: true
    },
    'brazil': {
      dataResidency: ['BR'],
      complianceFrameworks: ['LGPD', 'BACEN', 'CVM'],
      auditRetention: 1825, // 5 anos em dias
      encryptionRequired: true
    },
    'europe': {
      dataResidency: ['EU'],
      complianceFrameworks: ['GDPR', 'PSD2', 'MiFID'],
      auditRetention: 2555, // 7 anos em dias
      encryptionRequired: true
    },
    'china': {
      dataResidency: ['CN'],
      complianceFrameworks: ['PIPL', 'CSL', 'PBOC'],
      auditRetention: 1825, // 5 anos em dias
      encryptionRequired: true
    }
  }
};

/**
 * Decorator para configurar validação de tenant
 */
export const RequireTenant = (config?: Partial<TenantGuardConfig>) => {
  return (target: any, propertyKey?: string, descriptor?: PropertyDescriptor) => {
    const mergedConfig = { ...DEFAULT_CONFIG, ...config };
    Reflect.defineMetadata('tenant-guard-config', mergedConfig, target, propertyKey);
  };
};

/**
 * Guard para validação e isolamento de tenant
 */
@Injectable()
export class TenantGuard implements CanActivate {
  private readonly logger = new Logger(TenantGuard.name);

  constructor(
    private readonly reflector: Reflector,
    private readonly configService: ConfigService,
    @Inject(CACHE_MANAGER) private readonly cacheManager: Cache
  ) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const config = this.getTenantGuardConfig(context);
    
    if (!config) {
      // Se não há configuração, permitir acesso (tenant opcional)
      return true;
    }

    const request = context.switchToHttp().getRequest<Request>();
    const startTime = Date.now();

    try {
      // Validar tenant
      const validationResult = await this.validateTenant(request, config, context);
      
      // Anexar contexto do tenant ao request
      (request as any).tenant = validationResult.tenant;
      (request as any).tenantValidation = validationResult;
      
      // Log da validação
      this.logTenantValidation(validationResult, config);
      
      // Verificar se validação foi bem-sucedida
      if (!validationResult.isValid) {
        this.handleValidationFailure(validationResult, config);
        return false;
      }
      
      // Aplicar configurações de isolamento de dados
      await this.applyDataIsolation(validationResult.tenant!, request, config);
      
      return true;

    } catch (error) {
      this.logger.error(`Tenant validation failed: ${error.message}`, error.stack);
      
      if (config.validation.strict) {
        throw new ForbiddenException('Tenant validation failed');
      }
      
      // Em modo não-estrito, permitir acesso com log de warning
      this.logger.warn('Proceeding with request due to non-strict mode');
      return true;
    }
  }

  /**
   * Obter configuração do Tenant Guard
   */
  private getTenantGuardConfig(context: ExecutionContext): TenantGuardConfig | null {
    return this.reflector.get<TenantGuardConfig>('tenant-guard-config', context.getHandler()) ||
           this.reflector.get<TenantGuardConfig>('tenant-guard-config', context.getClass());
  }

  /**
   * Validar tenant
   */
  private async validateTenant(
    request: Request,
    config: TenantGuardConfig,
    context: ExecutionContext
  ): Promise<TenantValidationResult> {
    const validationId = this.generateValidationId();
    const startTime = Date.now();
    
    // Identificar tenant
    const tenantId = await this.identifyTenant(request, config);
    
    if (!tenantId && config.validation.required) {
      return this.createValidationResult(validationId, startTime, false, [], ['Tenant ID é obrigatório'], request, context);
    }
    
    if (!tenantId) {
      // Tenant não obrigatório
      return this.createValidationResult(validationId, startTime, true, [], [], request, context);
    }
    
    // Verificar cache
    if (config.cacheConfig.enabled) {
      const cached = await this.getCachedTenant(tenantId, config);
      if (cached) {
        this.logger.debug(`Using cached tenant context: ${tenantId}`);
        return this.createValidationResult(validationId, startTime, true, [cached], [], request, context, tenantId);
      }
    }
    
    // Carregar contexto do tenant
    const tenant = await this.loadTenantContext(tenantId, config);
    
    if (!tenant) {
      return this.createValidationResult(validationId, startTime, false, [], [`Tenant não encontrado: ${tenantId}`], request, context, tenantId);
    }
    
    // Validar tenant
    const validationErrors = await this.performTenantValidation(tenant, config, request);
    
    if (validationErrors.length > 0) {
      return this.createValidationResult(validationId, startTime, false, [], validationErrors, request, context, tenantId);
    }
    
    // Armazenar em cache
    if (config.cacheConfig.enabled) {
      await this.setCachedTenant(tenant, config);
    }
    
    return this.createValidationResult(validationId, startTime, true, [tenant], [], request, context, tenantId);
  }

  /**
   * Identificar tenant usando múltiplas estratégias
   */
  private async identifyTenant(request: Request, config: TenantGuardConfig): Promise<string | null> {
    const strategies = config.identificationStrategies;
    
    // Estratégia 1: Header X-Tenant-ID
    if (strategies.header) {
      const headerTenant = request.get('X-Tenant-ID') || request.get('x-tenant-id');
      if (headerTenant) {
        this.logger.debug(`Tenant identified via header: ${headerTenant}`);
        return headerTenant;
      }
    }
    
    // Estratégia 2: JWT payload
    if (strategies.jwt) {
      const user = (request as any).user;
      if (user?.tenantId) {
        this.logger.debug(`Tenant identified via JWT: ${user.tenantId}`);
        return user.tenantId;
      }
    }
    
    // Estratégia 3: Subdomain
    if (strategies.subdomain) {
      const host = request.get('Host');
      if (host) {
        const subdomain = this.extractSubdomain(host);
        if (subdomain && subdomain !== 'www' && subdomain !== 'api') {
          this.logger.debug(`Tenant identified via subdomain: ${subdomain}`);
          return subdomain;
        }
      }
    }
    
    // Estratégia 4: Path parameter
    if (strategies.path) {
      const pathTenant = this.extractTenantFromPath(request.path);
      if (pathTenant) {
        this.logger.debug(`Tenant identified via path: ${pathTenant}`);
        return pathTenant;
      }
    }
    
    // Estratégia 5: Query parameter
    if (strategies.query) {
      const queryTenant = request.query.tenant as string;
      if (queryTenant) {
        this.logger.debug(`Tenant identified via query: ${queryTenant}`);
        return queryTenant;
      }
    }
    
    // Estratégia 6: Cookie
    if (strategies.cookie) {
      const cookieTenant = request.cookies?.tenant;
      if (cookieTenant) {
        this.logger.debug(`Tenant identified via cookie: ${cookieTenant}`);
        return cookieTenant;
      }
    }
    
    return null;
  }

  /**
   * Carregar contexto enriquecido do tenant
   */
  private async loadTenantContext(tenantId: string, config: TenantGuardConfig): Promise<EnrichedTenantContext | null> {
    try {
      // Em uma implementação real, isso carregaria do banco de dados
      // Por agora, retornamos um contexto simulado
      const mockTenant: EnrichedTenantContext = {
        id: tenantId,
        name: `Tenant ${tenantId}`,
        code: tenantId.toUpperCase(),
        status: 'active',
        type: 'enterprise',
        tier: 'premium',
        settings: {
          timezone: 'UTC',
          locale: 'pt-BR',
          currency: 'USD',
          dateFormat: 'DD/MM/YYYY',
          numberFormat: '1.234,56'
        },
        regulatory: {
          jurisdiction: 'global',
          primaryRegulator: 'GLOBAL',
          complianceFrameworks: ['ISO27001', 'GDPR'],
          dataResidency: ['GLOBAL']
        },
        security: {
          encryptionLevel: 'high',
          mfaRequired: true,
          sessionTimeout: 3600,
          allowedCountries: [],
          blockedCountries: []
        },
        limits: {
          maxUsers: 1000,
          maxSessions: 100,
          maxApiCalls: 10000,
          maxStorage: 1000000,
          maxTransactions: 50000
        },
        audit: {
          retentionPeriod: 2555,
          complianceLevel: 'high',
          auditFrequency: 'monthly'
        },
        metadata: {
          createdAt: new Date(),
          updatedAt: new Date(),
          version: '1.0.0'
        },
        integration: {
          externalId: `ext_${tenantId}`,
          childTenants: [],
          partnerships: [],
          apiKeys: []
        }
      };
      
      return mockTenant;
      
    } catch (error) {
      this.logger.error(`Failed to load tenant context: ${error.message}`, error.stack);
      return null;
    }
  }

  /**
   * Realizar validações específicas do tenant
   */
  private async performTenantValidation(
    tenant: EnrichedTenantContext,
    config: TenantGuardConfig,
    request: Request
  ): Promise<string[]> {
    const errors: string[] = [];
    
    // Validar status
    if (config.validation.validateStatus && tenant.status !== 'active') {
      errors.push(`Tenant está inativo: ${tenant.status}`);
    }
    
    // Validar licença
    if (config.validation.validateLicense && tenant.regulatory.licenseExpiry) {
      if (tenant.regulatory.licenseExpiry < new Date()) {
        errors.push('Licença do tenant expirada');
      }
    }
    
    // Validar jurisdição
    if (config.validation.validateJurisdiction) {
      const clientIP = this.getClientIP(request);
      const isAllowed = await this.validateJurisdictionAccess(tenant, clientIP);
      if (!isAllowed) {
        errors.push('Acesso negado para esta jurisdição');
      }
    }
    
    // Validar limites
    const currentUsage = await this.getCurrentUsage(tenant.id);
    if (currentUsage.sessions >= tenant.limits.maxSessions) {
      errors.push('Limite de sessões excedido');
    }
    
    return errors;
  }

  /**
   * Aplicar isolamento de dados
   */
  private async applyDataIsolation(
    tenant: EnrichedTenantContext,
    request: Request,
    config: TenantGuardConfig
  ): Promise<void> {
    if (config.dataIsolation.enforceRowLevelSecurity) {
      // Configurar contexto de RLS no request
      (request as any).tenantContext = {
        tenantId: tenant.id,
        dataIsolationLevel: tenant.security.encryptionLevel,
        allowedOperations: this.getAllowedOperations(tenant),
        auditRequired: config.auditConfig.logAllAccess
      };
    }
  }

  /**
   * Criar resultado de validação
   */
  private createValidationResult(
    validationId: string,
    startTime: number,
    isValid: boolean,
    tenants: EnrichedTenantContext[],
    errors: string[],
    request: Request,
    context: ExecutionContext,
    tenantId?: string
  ): TenantValidationResult {
    const user = (request as any).user;
    
    return {
      isValid,
      tenant: tenants[0],
      errors,
      warnings: [],
      validationId,
      timestamp: new Date(),
      processingTime: Date.now() - startTime,
      validationDetails: {
        identificationMethod: this.getIdentificationMethod(request),
        tenantId: tenantId || 'unknown',
        ipAddress: this.getClientIP(request),
        userAgent: request.get('User-Agent') || 'unknown',
        endpoint: `${request.method} ${request.path}`,
        userId: user?.id,
        sessionId: user?.sessionId
      },
      compliance: {
        frameworks: tenants[0]?.regulatory.complianceFrameworks || [],
        violations: errors,
        requirements: [],
        recommendations: []
      }
    };
  }

  /**
   * Tratar falha de validação
   */
  private handleValidationFailure(result: TenantValidationResult, config: TenantGuardConfig): void {
    const errorMessage = result.errors.join('; ');
    
    this.logger.error(`Tenant validation failed: ${errorMessage}`, {
      validationId: result.validationId,
      tenantId: result.validationDetails.tenantId,
      errors: result.errors
    });
    
    if (result.errors.some(error => error.includes('não encontrado'))) {
      throw new BadRequestException('Tenant inválido');
    }
    
    if (result.errors.some(error => error.includes('inativo') || error.includes('expirada'))) {
      throw new ForbiddenException('Tenant não autorizado');
    }
    
    throw new ForbiddenException('Acesso negado ao tenant');
  }

  /**
   * Log da validação do tenant
   */
  private logTenantValidation(result: TenantValidationResult, config: TenantGuardConfig): void {
    if (config.auditConfig.logAllAccess || 
        (config.auditConfig.logViolations && !result.isValid)) {
      
      const logLevel = result.isValid ? 'log' : 'error';
      const message = result.isValid ? 'Tenant validation successful' : 'Tenant validation failed';
      
      this.logger[logLevel](message, {
        validationId: result.validationId,
        tenantId: result.validationDetails.tenantId,
        isValid: result.isValid,
        processingTime: result.processingTime,
        errors: result.errors,
        compliance: result.compliance
      });
    }
  }

  // Métodos auxiliares
  private extractSubdomain(host: string): string | null {
    const parts = host.split('.');
    return parts.length > 2 ? parts[0] : null;
  }

  private extractTenantFromPath(path: string): string | null {
    const match = path.match(/^\/(?:api\/)?(?:v\d+\/)?tenant\/([^\/]+)/);
    return match ? match[1] : null;
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

  private getIdentificationMethod(request: Request): string {
    if (request.get('X-Tenant-ID')) return 'header';
    if ((request as any).user?.tenantId) return 'jwt';
    if (request.get('Host')?.includes('.')) return 'subdomain';
    if (request.path.includes('/tenant/')) return 'path';
    if (request.query.tenant) return 'query';
    if (request.cookies?.tenant) return 'cookie';
    return 'unknown';
  }

  private async validateJurisdictionAccess(tenant: EnrichedTenantContext, clientIP: string): Promise<boolean> {
    // Implementar validação de jurisdição baseada em IP
    return true; // Placeholder
  }

  private async getCurrentUsage(tenantId: string): Promise<{ sessions: number; apiCalls: number; storage: number }> {
    // Implementar consulta de uso atual
    return { sessions: 0, apiCalls: 0, storage: 0 }; // Placeholder
  }

  private getAllowedOperations(tenant: EnrichedTenantContext): string[] {
    // Implementar lógica de operações permitidas
    return ['read', 'write', 'delete']; // Placeholder
  }

  private async getCachedTenant(tenantId: string, config: TenantGuardConfig): Promise<EnrichedTenantContext | null> {
    if (!config.cacheConfig.enabled) return null;
    
    try {
      const cacheKey = `${config.cacheConfig.keyPrefix}${tenantId}`;
      return await this.cacheManager.get<EnrichedTenantContext>(cacheKey);
    } catch (error) {
      this.logger.warn(`Failed to get cached tenant: ${error.message}`);
      return null;
    }
  }

  private async setCachedTenant(tenant: EnrichedTenantContext, config: TenantGuardConfig): Promise<void> {
    if (!config.cacheConfig.enabled) return;
    
    try {
      const cacheKey = `${config.cacheConfig.keyPrefix}${tenant.id}`;
      await this.cacheManager.set(cacheKey, tenant, config.cacheConfig.ttl * 1000);
    } catch (error) {
      this.logger.warn(`Failed to cache tenant: ${error.message}`);
    }
  }

  private generateValidationId(): string {
    return `tenant_val_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }
}