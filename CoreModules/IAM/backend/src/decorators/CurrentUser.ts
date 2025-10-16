/**
 * 👤 CURRENT USER DECORATOR - INNOVABIZ IAM
 * Decorator para injeção do usuário atual autenticado
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: NIST SP 800-63B, OWASP Authentication Guidelines
 * Mercados: Angola (BNA), Europa, América, China, BRICS, Brasil
 * Multi-tenant: Isolamento de dados por tenant, contexto multi-dimensional
 */

import { createParamDecorator, ExecutionContext, UnauthorizedException } from '@nestjs/common';
import { Request } from 'express';

/**
 * Interface do usuário autenticado com contexto multi-dimensional
 */
export interface AuthenticatedUser {
  // Identificação básica
  id: string;
  email: string;
  username?: string;
  displayName?: string;
  
  // Contexto multi-tenant
  tenantId: string;
  tenantName?: string;
  tenantType?: 'enterprise' | 'government' | 'financial' | 'healthcare' | 'education';
  
  // Contexto geográfico e regulatório
  jurisdiction: 'angola' | 'europe' | 'america' | 'china' | 'brics' | 'brazil' | 'global';
  regulatoryFramework: string[]; // BNA, ECB, FED, PBOC, BCB, etc.
  dataResidency: string; // País onde os dados devem residir
  
  // Perfis e permissões
  roles: string[];
  permissions: string[];
  scope: string[];
  
  // Contexto de sessão
  sessionId?: string;
  deviceId?: string;
  ipAddress?: string;
  userAgent?: string;
  
  // Contexto de segurança
  riskScore: number;
  riskLevel: 'low' | 'medium' | 'high' | 'critical';
  requiresMFA: boolean;
  lastActivity: Date;
  
  // Contexto de negócio
  organizationId?: string;
  departmentId?: string;
  businessUnit?: string;
  costCenter?: string;
  
  // Contexto financeiro (para instituições financeiras)
  financialLicense?: string; // Licença bancária, corretora, etc.
  supervisoryAuthority?: string; // BNA, ECB, FED, etc.
  tier?: 'retail' | 'corporate' | 'investment' | 'private';
  
  // Contexto de compliance
  complianceLevel: 'basic' | 'enhanced' | 'premium';
  kycStatus: 'pending' | 'verified' | 'enhanced' | 'rejected';
  amlRisk: 'low' | 'medium' | 'high';
  
  // Metadados
  createdAt: Date;
  lastLogin: Date;
  timezone: string;
  locale: string;
  currency: string;
  
  // Contexto de integração
  externalIds?: Record<string, string>; // IDs em sistemas externos
  integrationContext?: Record<string, any>;
}

/**
 * Opções para o decorator CurrentUser
 */
export interface CurrentUserOptions {
  // Se deve lançar exceção quando usuário não encontrado
  required?: boolean;
  
  // Campos específicos a serem incluídos/excluídos
  include?: (keyof AuthenticatedUser)[];
  exclude?: (keyof AuthenticatedUser)[];
  
  // Validações adicionais
  requireRoles?: string[];
  requirePermissions?: string[];
  maxRiskLevel?: 'low' | 'medium' | 'high' | 'critical';
  
  // Contexto específico
  requireJurisdiction?: string[];
  requireComplianceLevel?: string[];
  requireKYCStatus?: string[];
  
  // Auditoria
  auditAccess?: boolean;
  sensitiveOperation?: boolean;
}

/**
 * Decorator para obter o usuário atual autenticado
 * 
 * @param options Opções de configuração
 * @returns Dados do usuário autenticado
 * 
 * @example
 * // Uso básico
 * async getProfile(@CurrentUser() user: AuthenticatedUser) {
 *   return user;
 * }
 * 
 * @example
 * // Com validações específicas
 * async adminOperation(
 *   @CurrentUser({ 
 *     requireRoles: ['admin'], 
 *     maxRiskLevel: 'medium',
 *     requireJurisdiction: ['angola', 'brazil'],
 *     auditAccess: true 
 *   }) user: AuthenticatedUser
 * ) {
 *   // Operação administrativa
 * }
 * 
 * @example
 * // Para operações financeiras
 * async transferFunds(
 *   @CurrentUser({ 
 *     requirePermissions: ['transfer:funds'],
 *     requireKYCStatus: ['verified', 'enhanced'],
 *     maxRiskLevel: 'low',
 *     sensitiveOperation: true
 *   }) user: AuthenticatedUser
 * ) {
 *   // Transferência de fundos
 * }
 */
export const CurrentUser = createParamDecorator(
  (options: CurrentUserOptions = {}, ctx: ExecutionContext): AuthenticatedUser => {
    const request = ctx.switchToHttp().getRequest<Request>();
    const user = request.user as AuthenticatedUser;

    // Verificar se usuário existe
    if (!user && options.required !== false) {
      throw new UnauthorizedException('Usuário não autenticado');
    }

    if (!user) {
      return null;
    }

    // Enriquecer dados do usuário com contexto adicional
    const enrichedUser = enrichUserContext(user, request);

    // Aplicar validações
    validateUserContext(enrichedUser, options);

    // Filtrar campos se especificado
    const filteredUser = filterUserFields(enrichedUser, options);

    // Registrar acesso para auditoria se necessário
    if (options.auditAccess || options.sensitiveOperation) {
      setImmediate(() => {
        auditUserAccess(filteredUser, request, options);
      });
    }

    return filteredUser;
  }
);

/**
 * Enriquecer contexto do usuário com informações adicionais
 */
function enrichUserContext(user: AuthenticatedUser, request: Request): AuthenticatedUser {
  const enriched = { ...user };

  // Enriquecer com dados da requisição
  if (!enriched.ipAddress) {
    enriched.ipAddress = getClientIP(request);
  }

  if (!enriched.userAgent) {
    enriched.userAgent = request.get('User-Agent') || 'unknown';
  }

  // Determinar jurisdição se não definida
  if (!enriched.jurisdiction) {
    enriched.jurisdiction = determineJurisdiction(enriched, request);
  }

  // Definir framework regulatório baseado na jurisdição
  if (!enriched.regulatoryFramework || enriched.regulatoryFramework.length === 0) {
    enriched.regulatoryFramework = getRegulatoryFramework(enriched.jurisdiction);
  }

  // Definir residência de dados
  if (!enriched.dataResidency) {
    enriched.dataResidency = getDataResidency(enriched.jurisdiction);
  }

  // Definir timezone se não definido
  if (!enriched.timezone) {
    enriched.timezone = getTimezoneForJurisdiction(enriched.jurisdiction);
  }

  // Definir moeda padrão se não definida
  if (!enriched.currency) {
    enriched.currency = getCurrencyForJurisdiction(enriched.jurisdiction);
  }

  // Definir locale se não definido
  if (!enriched.locale) {
    enriched.locale = getLocaleForJurisdiction(enriched.jurisdiction);
  }

  return enriched;
}

/**
 * Validar contexto do usuário baseado nas opções
 */
function validateUserContext(user: AuthenticatedUser, options: CurrentUserOptions): void {
  // Validar roles obrigatórias
  if (options.requireRoles && options.requireRoles.length > 0) {
    const hasRequiredRole = options.requireRoles.some(role => user.roles?.includes(role));
    if (!hasRequiredRole) {
      throw new UnauthorizedException(`Acesso negado: roles obrigatórias [${options.requireRoles.join(', ')}]`);
    }
  }

  // Validar permissões obrigatórias
  if (options.requirePermissions && options.requirePermissions.length > 0) {
    const hasRequiredPermission = options.requirePermissions.some(permission => 
      user.permissions?.includes(permission)
    );
    if (!hasRequiredPermission) {
      throw new UnauthorizedException(`Acesso negado: permissões obrigatórias [${options.requirePermissions.join(', ')}]`);
    }
  }

  // Validar nível de risco máximo
  if (options.maxRiskLevel) {
    const riskLevels = { low: 1, medium: 2, high: 3, critical: 4 };
    const userRiskLevel = riskLevels[user.riskLevel] || 4;
    const maxRiskLevel = riskLevels[options.maxRiskLevel];
    
    if (userRiskLevel > maxRiskLevel) {
      throw new UnauthorizedException(`Acesso negado: nível de risco muito alto (${user.riskLevel})`);
    }
  }

  // Validar jurisdição
  if (options.requireJurisdiction && options.requireJurisdiction.length > 0) {
    if (!options.requireJurisdiction.includes(user.jurisdiction)) {
      throw new UnauthorizedException(`Acesso negado: jurisdição não autorizada (${user.jurisdiction})`);
    }
  }

  // Validar nível de compliance
  if (options.requireComplianceLevel && options.requireComplianceLevel.length > 0) {
    if (!options.requireComplianceLevel.includes(user.complianceLevel)) {
      throw new UnauthorizedException(`Acesso negado: nível de compliance insuficiente (${user.complianceLevel})`);
    }
  }

  // Validar status KYC
  if (options.requireKYCStatus && options.requireKYCStatus.length > 0) {
    if (!options.requireKYCStatus.includes(user.kycStatus)) {
      throw new UnauthorizedException(`Acesso negado: status KYC insuficiente (${user.kycStatus})`);
    }
  }
}

/**
 * Filtrar campos do usuário baseado nas opções
 */
function filterUserFields(user: AuthenticatedUser, options: CurrentUserOptions): AuthenticatedUser {
  if (!options.include && !options.exclude) {
    return user;
  }

  const filtered = { ...user };

  // Se include foi especificado, manter apenas esses campos
  if (options.include && options.include.length > 0) {
    const result = {} as AuthenticatedUser;
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
 * Registrar acesso do usuário para auditoria
 */
async function auditUserAccess(
  user: AuthenticatedUser, 
  request: Request, 
  options: CurrentUserOptions
): Promise<void> {
  try {
    // Implementar logging de auditoria
    const auditData = {
      userId: user.id,
      tenantId: user.tenantId,
      action: 'USER_ACCESS',
      resource: `${request.method} ${request.path}`,
      severity: options.sensitiveOperation ? 'high' : 'info',
      metadata: {
        jurisdiction: user.jurisdiction,
        riskLevel: user.riskLevel,
        complianceLevel: user.complianceLevel,
        kycStatus: user.kycStatus,
        ipAddress: user.ipAddress,
        userAgent: user.userAgent,
        roles: user.roles,
        permissions: user.permissions,
        sensitiveOperation: options.sensitiveOperation,
        timestamp: new Date().toISOString()
      }
    };

    // Aqui seria feita a chamada para o serviço de auditoria
    console.log('Audit Log:', JSON.stringify(auditData, null, 2));
  } catch (error) {
    console.error('Failed to audit user access:', error);
  }
}

/**
 * Determinar jurisdição baseada no usuário e contexto
 */
function determineJurisdiction(user: AuthenticatedUser, request: Request): AuthenticatedUser['jurisdiction'] {
  // Lógica para determinar jurisdição baseada em:
  // - Tenant do usuário
  // - IP de origem
  // - Configurações do usuário
  // - Headers específicos
  
  if (user.tenantId?.includes('bna') || user.tenantId?.includes('angola')) {
    return 'angola';
  }
  
  if (user.tenantId?.includes('bcb') || user.tenantId?.includes('brazil')) {
    return 'brazil';
  }
  
  if (user.tenantId?.includes('ecb') || user.tenantId?.includes('europe')) {
    return 'europe';
  }
  
  if (user.tenantId?.includes('fed') || user.tenantId?.includes('america')) {
    return 'america';
  }
  
  if (user.tenantId?.includes('pboc') || user.tenantId?.includes('china')) {
    return 'china';
  }
  
  if (user.tenantId?.includes('brics')) {
    return 'brics';
  }
  
  return 'global';
}

/**
 * Obter framework regulatório baseado na jurisdição
 */
function getRegulatoryFramework(jurisdiction: AuthenticatedUser['jurisdiction']): string[] {
  const frameworks: Record<string, string[]> = {
    angola: ['BNA', 'BODIVA', 'ARSEG', 'IRSEM'],
    brazil: ['BCB', 'CVM', 'SUSEP', 'PREVIC'],
    europe: ['ECB', 'EBA', 'ESMA', 'EIOPA', 'GDPR'],
    america: ['FED', 'SEC', 'CFTC', 'OCC', 'FDIC'],
    china: ['PBOC', 'CBIRC', 'CSRC', 'SAFE'],
    brics: ['BNA', 'BCB', 'RBI', 'PBOC', 'SARB'],
    global: ['BASEL_III', 'FATF', 'ISO_27001', 'NIST']
  };
  
  return frameworks[jurisdiction] || frameworks.global;
}

/**
 * Obter residência de dados baseada na jurisdição
 */
function getDataResidency(jurisdiction: AuthenticatedUser['jurisdiction']): string {
  const residency: Record<string, string> = {
    angola: 'AO',
    brazil: 'BR',
    europe: 'EU',
    america: 'US',
    china: 'CN',
    brics: 'BRICS',
    global: 'GLOBAL'
  };
  
  return residency[jurisdiction] || 'GLOBAL';
}

/**
 * Obter timezone baseado na jurisdição
 */
function getTimezoneForJurisdiction(jurisdiction: AuthenticatedUser['jurisdiction']): string {
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

/**
 * Obter moeda baseada na jurisdição
 */
function getCurrencyForJurisdiction(jurisdiction: AuthenticatedUser['jurisdiction']): string {
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

/**
 * Obter locale baseado na jurisdição
 */
function getLocaleForJurisdiction(jurisdiction: AuthenticatedUser['jurisdiction']): string {
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

/**
 * Obter IP real do cliente
 */
function getClientIP(request: Request): string {
  const forwarded = request.get('X-Forwarded-For');
  const realIP = request.get('X-Real-IP');
  const cfConnectingIP = request.get('CF-Connecting-IP');
  
  if (cfConnectingIP) return cfConnectingIP;
  if (realIP) return realIP;
  if (forwarded) return forwarded.split(',')[0].trim();
  
  return request.ip || request.connection.remoteAddress || 'unknown';
}