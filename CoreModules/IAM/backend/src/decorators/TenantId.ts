/**
 * 🏢 TENANT ID DECORATOR - INNOVABIZ IAM
 * Decorator para injeção do ID do tenant com contexto multi-dimensional
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: Multi-tenancy Best Practices, Data Isolation Standards
 * Mercados: Angola (BNA), Europa, América, China, BRICS, Brasil
 * Arquitetura: Multi-tenant, Multi-camada, Multi-contexto, Multi-dimensional
 */

import { 
  createParamDecorator, 
  ExecutionContext, 
  BadRequestException, 
  ForbiddenException 
} from '@nestjs/common';
import { Request } from 'express';

/**
 * Contexto completo do tenant com informações regulatórias e de negócio
 */
export interface TenantContext {
  // Identificação básica
  id: string;
  name: string;
  displayName: string;
  description?: string;
  
  // Tipo e categoria
  type: 'enterprise' | 'government' | 'financial' | 'healthcare' | 'education' | 'retail' | 'technology';
  category: 'bank' | 'insurance' | 'investment' | 'fintech' | 'government' | 'corporate' | 'startup';
  tier: 'basic' | 'standard' | 'premium' | 'enterprise' | 'government';
  
  // Contexto geográfico e regulatório
  jurisdiction: 'angola' | 'europe' | 'america' | 'china' | 'brics' | 'brazil' | 'global';
  country: string;
  region: string;
  city?: string;
  
  // Framework regulatório específico
  regulatoryFramework: string[];
  supervisoryAuthority: string[];
  licenses: string[];
  certifications: string[];
  
  // Configurações de compliance
  complianceLevel: 'basic' | 'enhanced' | 'premium' | 'government';
  dataClassification: 'public' | 'internal' | 'confidential' | 'restricted' | 'top-secret';
  dataResidency: string[];
  encryptionRequired: boolean;
  auditLevel: 'standard' | 'enhanced' | 'forensic';
  
  // Configurações financeiras (para instituições financeiras)
  financialLicenses?: {
    bankingLicense?: string;
    insuranceLicense?: string;
    investmentLicense?: string;
    paymentLicense?: string;
    cryptoLicense?: string;
  };
  
  // Configurações de negócio
  businessModel: 'b2b' | 'b2c' | 'b2g' | 'b2b2c' | 'marketplace' | 'platform';
  industry: string[];
  sectors: string[];
  
  // Configurações técnicas
  features: string[];
  modules: string[];
  integrations: string[];
  apiLimits: {
    requestsPerMinute: number;
    requestsPerDay: number;
    concurrentConnections: number;
  };
  
  // Configurações de segurança
  securityLevel: 'standard' | 'high' | 'critical' | 'maximum';
  mfaRequired: boolean;
  passwordPolicy: string;
  sessionTimeout: number;
  ipWhitelist?: string[];
  
  // Configurações de localização
  timezone: string;
  locale: string;
  currency: string;
  dateFormat: string;
  numberFormat: string;
  
  // Status e metadados
  status: 'active' | 'inactive' | 'suspended' | 'terminated' | 'pending';
  createdAt: Date;
  updatedAt: Date;
  lastActivity?: Date;
  
  // Configurações de cobrança
  billingModel: 'free' | 'subscription' | 'usage' | 'enterprise' | 'government';
  billingCurrency: string;
  billingCycle: 'monthly' | 'quarterly' | 'yearly' | 'custom';
  
  // Limites e quotas
  limits: {
    users: number;
    storage: number; // em GB
    bandwidth: number; // em GB/mês
    transactions: number; // por mês
    apiCalls: number; // por mês
  };
  
  // Configurações de integração
  externalIds?: Record<string, string>;
  webhooks?: string[];
  callbacks?: string[];
  
  // Metadados customizados
  customFields?: Record<string, any>;
  tags?: string[];
}

/**
 * Opções para o decorator TenantId
 */
export interface TenantIdOptions {
  // Se deve retornar apenas o ID ou o contexto completo
  fullContext?: boolean;
  
  // Validações específicas
  requireStatus?: ('active' | 'inactive' | 'suspended' | 'terminated' | 'pending')[];
  requireType?: string[];
  requireJurisdiction?: string[];
  requireComplianceLevel?: string[];
  requireSecurityLevel?: string[];
  
  // Campos específicos a incluir/excluir (quando fullContext = true)
  include?: (keyof TenantContext)[];
  exclude?: (keyof TenantContext)[];
  
  // Configurações de cache
  useCache?: boolean;
  cacheTTL?: number;
  
  // Auditoria
  auditAccess?: boolean;
  sensitiveOperation?: boolean;
  
  // Fallback
  allowFallback?: boolean;
  fallbackTenantId?: string;
}

/**
 * Decorator para obter o ID do tenant ou contexto completo
 * 
 * @param options Opções de configuração
 * @returns ID do tenant ou contexto completo
 * 
 * @example
 * // Uso básico - apenas ID
 * async getUsers(@TenantId() tenantId: string) {
 *   return this.userService.findByTenant(tenantId);
 * }
 * 
 * @example
 * // Contexto completo
 * async getTenantConfig(@TenantId({ fullContext: true }) tenant: TenantContext) {
 *   return {
 *     features: tenant.features,
 *     limits: tenant.limits,
 *     compliance: tenant.complianceLevel
 *   };
 * }
 * 
 * @example
 * // Com validações específicas
 * async financialOperation(
 *   @TenantId({ 
 *     requireType: ['financial'],
 *     requireJurisdiction: ['angola', 'brazil'],
 *     requireComplianceLevel: ['enhanced', 'premium'],
 *     auditAccess: true 
 *   }) tenantId: string
 * ) {
 *   // Operação financeira
 * }
 * 
 * @example
 * // Para operações governamentais
 * async governmentService(
 *   @TenantId({ 
 *     requireType: ['government'],
 *     requireSecurityLevel: ['critical', 'maximum'],
 *     fullContext: true,
 *     sensitiveOperation: true
 *   }) tenant: TenantContext
 * ) {
 *   // Serviço governamental
 * }
 */
export const TenantId = createParamDecorator(
  async (options: TenantIdOptions = {}, ctx: ExecutionContext): Promise<string | TenantContext> => {
    const request = ctx.switchToHttp().getRequest<Request>();
    
    // Extrair tenant ID de múltiplas fontes
    const tenantId = extractTenantId(request, options);
    
    if (!tenantId) {
      if (options.allowFallback && options.fallbackTenantId) {
        return options.fullContext 
          ? await getTenantContext(options.fallbackTenantId, options)
          : options.fallbackTenantId;
      }
      throw new BadRequestException('Tenant ID não encontrado');
    }
    
    // Se apenas ID é necessário e não há validações
    if (!options.fullContext && !hasValidations(options)) {
      return tenantId;
    }
    
    // Obter contexto completo do tenant
    const tenantContext = await getTenantContext(tenantId, options);
    
    // Aplicar validações
    validateTenantContext(tenantContext, options);
    
    // Registrar acesso para auditoria se necessário
    if (options.auditAccess || options.sensitiveOperation) {
      setImmediate(() => {
        auditTenantAccess(tenantContext, request, options);
      });
    }
    
    // Retornar ID ou contexto completo
    if (options.fullContext) {
      return filterTenantFields(tenantContext, options);
    }
    
    return tenantId;
  }
);

/**
 * Extrair tenant ID de múltiplas fontes
 */
function extractTenantId(request: Request, options: TenantIdOptions): string | null {
  // 1. Header X-Tenant-ID (prioridade alta)
  let tenantId = request.get('X-Tenant-ID');
  if (tenantId) return tenantId;
  
  // 2. Header Authorization Bearer token (extrair do JWT)
  const authHeader = request.get('Authorization');
  if (authHeader && authHeader.startsWith('Bearer ')) {
    tenantId = extractTenantFromJWT(authHeader.substring(7));
    if (tenantId) return tenantId;
  }
  
  // 3. Query parameter
  tenantId = request.query.tenantId as string;
  if (tenantId) return tenantId;
  
  // 4. Path parameter
  tenantId = request.params.tenantId;
  if (tenantId) return tenantId;
  
  // 5. Subdomain (para multi-tenant baseado em subdomain)
  const host = request.get('Host');
  if (host) {
    const subdomain = extractSubdomain(host);
    if (subdomain && subdomain !== 'www' && subdomain !== 'api') {
      return subdomain;
    }
  }
  
  // 6. Cookie
  tenantId = request.cookies?.tenantId;
  if (tenantId) return tenantId;
  
  // 7. User context (se usuário autenticado)
  const user = request.user as any;
  if (user?.tenantId) return user.tenantId;
  
  // 8. Custom header
  tenantId = request.get('X-Organization-ID') || request.get('X-Client-ID');
  if (tenantId) return tenantId;
  
  return null;
}

/**
 * Extrair tenant ID do JWT token
 */
function extractTenantFromJWT(token: string): string | null {
  try {
    // Decodificar JWT sem verificar assinatura (apenas para extrair tenant)
    const parts = token.split('.');
    if (parts.length !== 3) return null;
    
    const payload = JSON.parse(Buffer.from(parts[1], 'base64').toString());
    return payload.tenantId || payload.tenant_id || payload.org_id || null;
  } catch (error) {
    return null;
  }
}

/**
 * Extrair subdomain do host
 */
function extractSubdomain(host: string): string | null {
  const parts = host.split('.');
  if (parts.length >= 3) {
    return parts[0];
  }
  return null;
}

/**
 * Verificar se há validações configuradas
 */
function hasValidations(options: TenantIdOptions): boolean {
  return !!(
    options.requireStatus ||
    options.requireType ||
    options.requireJurisdiction ||
    options.requireComplianceLevel ||
    options.requireSecurityLevel
  );
}

/**
 * Obter contexto completo do tenant
 */
async function getTenantContext(tenantId: string, options: TenantIdOptions): Promise<TenantContext> {
  try {
    // Verificar cache primeiro se habilitado
    if (options.useCache !== false) {
      const cached = await getCachedTenantContext(tenantId);
      if (cached) return cached;
    }
    
    // Buscar contexto do tenant no banco de dados
    const context = await fetchTenantContextFromDatabase(tenantId);
    
    if (!context) {
      throw new BadRequestException(`Tenant não encontrado: ${tenantId}`);
    }
    
    // Armazenar em cache se habilitado
    if (options.useCache !== false) {
      await setCachedTenantContext(tenantId, context, options.cacheTTL || 300);
    }
    
    return context;
  } catch (error) {
    if (error instanceof BadRequestException) {
      throw error;
    }
    throw new BadRequestException(`Erro ao obter contexto do tenant: ${error.message}`);
  }
}

/**
 * Buscar contexto do tenant no banco de dados
 */
async function fetchTenantContextFromDatabase(tenantId: string): Promise<TenantContext | null> {
  // Implementação seria feita aqui para buscar no banco
  // Por enquanto, retornar contexto simulado baseado no tenantId
  
  const mockContext: TenantContext = {
    id: tenantId,
    name: `Tenant ${tenantId}`,
    displayName: `Tenant ${tenantId}`,
    type: determineTenantType(tenantId),
    category: determineTenantCategory(tenantId),
    tier: 'standard',
    jurisdiction: determineJurisdiction(tenantId),
    country: determineCountry(tenantId),
    region: determineRegion(tenantId),
    regulatoryFramework: getRegulatoryFramework(tenantId),
    supervisoryAuthority: getSupervisoryAuthority(tenantId),
    licenses: [],
    certifications: [],
    complianceLevel: 'enhanced',
    dataClassification: 'confidential',
    dataResidency: [determineCountry(tenantId)],
    encryptionRequired: true,
    auditLevel: 'enhanced',
    businessModel: 'b2b',
    industry: ['financial'],
    sectors: ['banking'],
    features: ['iam', 'payments', 'risk'],
    modules: ['core', 'advanced'],
    integrations: [],
    apiLimits: {
      requestsPerMinute: 1000,
      requestsPerDay: 100000,
      concurrentConnections: 100
    },
    securityLevel: 'high',
    mfaRequired: true,
    passwordPolicy: 'strong',
    sessionTimeout: 3600,
    timezone: getTimezone(tenantId),
    locale: getLocale(tenantId),
    currency: getCurrency(tenantId),
    dateFormat: 'DD/MM/YYYY',
    numberFormat: 'pt-BR',
    status: 'active',
    createdAt: new Date(),
    updatedAt: new Date(),
    billingModel: 'subscription',
    billingCurrency: getCurrency(tenantId),
    billingCycle: 'monthly',
    limits: {
      users: 1000,
      storage: 100,
      bandwidth: 1000,
      transactions: 100000,
      apiCalls: 1000000
    }
  };
  
  return mockContext;
}

/**
 * Validar contexto do tenant
 */
function validateTenantContext(context: TenantContext, options: TenantIdOptions): void {
  // Validar status
  if (options.requireStatus && options.requireStatus.length > 0) {
    if (!options.requireStatus.includes(context.status)) {
      throw new ForbiddenException(`Status do tenant não autorizado: ${context.status}`);
    }
  }
  
  // Validar tipo
  if (options.requireType && options.requireType.length > 0) {
    if (!options.requireType.includes(context.type)) {
      throw new ForbiddenException(`Tipo de tenant não autorizado: ${context.type}`);
    }
  }
  
  // Validar jurisdição
  if (options.requireJurisdiction && options.requireJurisdiction.length > 0) {
    if (!options.requireJurisdiction.includes(context.jurisdiction)) {
      throw new ForbiddenException(`Jurisdição não autorizada: ${context.jurisdiction}`);
    }
  }
  
  // Validar nível de compliance
  if (options.requireComplianceLevel && options.requireComplianceLevel.length > 0) {
    if (!options.requireComplianceLevel.includes(context.complianceLevel)) {
      throw new ForbiddenException(`Nível de compliance insuficiente: ${context.complianceLevel}`);
    }
  }
  
  // Validar nível de segurança
  if (options.requireSecurityLevel && options.requireSecurityLevel.length > 0) {
    if (!options.requireSecurityLevel.includes(context.securityLevel)) {
      throw new ForbiddenException(`Nível de segurança insuficiente: ${context.securityLevel}`);
    }
  }
}

/**
 * Filtrar campos do contexto do tenant
 */
function filterTenantFields(context: TenantContext, options: TenantIdOptions): TenantContext {
  if (!options.include && !options.exclude) {
    return context;
  }
  
  const filtered = { ...context };
  
  // Se include foi especificado, manter apenas esses campos
  if (options.include && options.include.length > 0) {
    const result = {} as TenantContext;
    options.include.forEach(field => {
      if (field in filtered) {
        (result as any)[field] = filtered[field];
      }
    });
    return result;
  }
  
  // Se exclude foi especificado, remover esses campos
  if (options.exclude && options.exclude.length > 0) {
    options.exclude.forEach(field => {
      delete (filtered as any)[field];
    });
  }
  
  return filtered;
}

/**
 * Registrar acesso do tenant para auditoria
 */
async function auditTenantAccess(
  context: TenantContext,
  request: Request,
  options: TenantIdOptions
): Promise<void> {
  try {
    const auditData = {
      tenantId: context.id,
      action: 'TENANT_ACCESS',
      resource: `${request.method} ${request.path}`,
      severity: options.sensitiveOperation ? 'high' : 'info',
      metadata: {
        tenantName: context.name,
        tenantType: context.type,
        jurisdiction: context.jurisdiction,
        complianceLevel: context.complianceLevel,
        securityLevel: context.securityLevel,
        sensitiveOperation: options.sensitiveOperation,
        ipAddress: getClientIP(request),
        userAgent: request.get('User-Agent'),
        timestamp: new Date().toISOString()
      }
    };
    
    // Aqui seria feita a chamada para o serviço de auditoria
    console.log('Tenant Audit Log:', JSON.stringify(auditData, null, 2));
  } catch (error) {
    console.error('Failed to audit tenant access:', error);
  }
}

// Funções auxiliares para determinar contexto baseado no tenantId
function determineTenantType(tenantId: string): TenantContext['type'] {
  if (tenantId.includes('bank') || tenantId.includes('bna') || tenantId.includes('bcb')) return 'financial';
  if (tenantId.includes('gov') || tenantId.includes('government')) return 'government';
  if (tenantId.includes('health') || tenantId.includes('hospital')) return 'healthcare';
  if (tenantId.includes('edu') || tenantId.includes('university')) return 'education';
  if (tenantId.includes('tech') || tenantId.includes('software')) return 'technology';
  return 'enterprise';
}

function determineTenantCategory(tenantId: string): TenantContext['category'] {
  if (tenantId.includes('bank') || tenantId.includes('bna') || tenantId.includes('bcb')) return 'bank';
  if (tenantId.includes('insurance') || tenantId.includes('seguro')) return 'insurance';
  if (tenantId.includes('investment') || tenantId.includes('invest')) return 'investment';
  if (tenantId.includes('fintech') || tenantId.includes('pay')) return 'fintech';
  if (tenantId.includes('gov') || tenantId.includes('government')) return 'government';
  return 'corporate';
}

function determineJurisdiction(tenantId: string): TenantContext['jurisdiction'] {
  if (tenantId.includes('bna') || tenantId.includes('angola')) return 'angola';
  if (tenantId.includes('bcb') || tenantId.includes('brazil')) return 'brazil';
  if (tenantId.includes('ecb') || tenantId.includes('europe')) return 'europe';
  if (tenantId.includes('fed') || tenantId.includes('america')) return 'america';
  if (tenantId.includes('pboc') || tenantId.includes('china')) return 'china';
  if (tenantId.includes('brics')) return 'brics';
  return 'global';
}

function determineCountry(tenantId: string): string {
  const jurisdiction = determineJurisdiction(tenantId);
  const countries: Record<string, string> = {
    angola: 'AO',
    brazil: 'BR',
    europe: 'EU',
    america: 'US',
    china: 'CN',
    brics: 'BRICS',
    global: 'GLOBAL'
  };
  return countries[jurisdiction] || 'GLOBAL';
}

function determineRegion(tenantId: string): string {
  const jurisdiction = determineJurisdiction(tenantId);
  const regions: Record<string, string> = {
    angola: 'Africa',
    brazil: 'South America',
    europe: 'Europe',
    america: 'North America',
    china: 'Asia',
    brics: 'Global',
    global: 'Global'
  };
  return regions[jurisdiction] || 'Global';
}

function getRegulatoryFramework(tenantId: string): string[] {
  const jurisdiction = determineJurisdiction(tenantId);
  const frameworks: Record<string, string[]> = {
    angola: ['BNA', 'BODIVA', 'ARSEG'],
    brazil: ['BCB', 'CVM', 'SUSEP'],
    europe: ['ECB', 'EBA', 'GDPR'],
    america: ['FED', 'SEC', 'OCC'],
    china: ['PBOC', 'CBIRC', 'CSRC'],
    brics: ['BASEL_III', 'FATF'],
    global: ['ISO_27001', 'NIST']
  };
  return frameworks[jurisdiction] || frameworks.global;
}

function getSupervisoryAuthority(tenantId: string): string[] {
  const jurisdiction = determineJurisdiction(tenantId);
  const authorities: Record<string, string[]> = {
    angola: ['Banco Nacional de Angola'],
    brazil: ['Banco Central do Brasil'],
    europe: ['European Central Bank'],
    america: ['Federal Reserve'],
    china: ['People\'s Bank of China'],
    brics: ['BRICS Supervisory Committee'],
    global: ['International Monetary Fund']
  };
  return authorities[jurisdiction] || authorities.global;
}

function getTimezone(tenantId: string): string {
  const jurisdiction = determineJurisdiction(tenantId);
  const timezones: Record<string, string> = {
    angola: 'Africa/Luanda',
    brazil: 'America/Sao_Paulo',
    europe: 'Europe/Brussels',
    america: 'America/New_York',
    china: 'Asia/Shanghai',
    brics: 'UTC',
    global: 'UTC'
  };
  return timezones[jurisdiction] || 'UTC';
}

function getLocale(tenantId: string): string {
  const jurisdiction = determineJurisdiction(tenantId);
  const locales: Record<string, string> = {
    angola: 'pt-AO',
    brazil: 'pt-BR',
    europe: 'en-EU',
    america: 'en-US',
    china: 'zh-CN',
    brics: 'en-US',
    global: 'en-US'
  };
  return locales[jurisdiction] || 'en-US';
}

function getCurrency(tenantId: string): string {
  const jurisdiction = determineJurisdiction(tenantId);
  const currencies: Record<string, string> = {
    angola: 'AOA',
    brazil: 'BRL',
    europe: 'EUR',
    america: 'USD',
    china: 'CNY',
    brics: 'USD',
    global: 'USD'
  };
  return currencies[jurisdiction] || 'USD';
}

// Funções de cache (implementação seria feita com Redis)
async function getCachedTenantContext(tenantId: string): Promise<TenantContext | null> {
  // Implementação do cache Redis seria feita aqui
  return null;
}

async function setCachedTenantContext(tenantId: string, context: TenantContext, ttl: number): Promise<void> {
  // Implementação do cache Redis seria feita aqui
}

function getClientIP(request: Request): string {
  const forwarded = request.get('X-Forwarded-For');
  const realIP = request.get('X-Real-IP');
  const cfConnectingIP = request.get('CF-Connecting-IP');
  
  if (cfConnectingIP) return cfConnectingIP;
  if (realIP) return realIP;
  if (forwarded) return forwarded.split(',')[0].trim();
  
  return request.ip || request.connection.remoteAddress || 'unknown';
}