/**
 * Conector de Integração entre IAM e Validadores de Conformidade
 * 
 * Este módulo estabelece a conexão entre o sistema IAM (Identity and Access Management)
 * e o sistema de validação de conformidade regulatória, permitindo validações de conformidade
 * automatizadas durante operações de autenticação e autorização.
 * 
 * @module IAMComplianceConnector
 */

import { Logger } from '../../../observability/logging/hook_logger';
import { Metrics } from '../../../observability/metrics/hook_metrics';
import { Tracer } from '../../../observability/tracing/hook_tracing';

import { 
  ComplianceValidatorIntegrator,
  ComplianceValidationRequest,
  Region,
  ConsolidatedComplianceResult
} from './compliance-validator-integrator';

/**
 * Interface para requisições de autorização de acesso
 */
export interface AccessAuthorizationRequest {
  userId: string;
  tenantId: string;
  requestId: string;
  resourceType: string;
  resourceId: string;
  operationType: 'read' | 'write' | 'delete' | 'credit_assessment' | 'identity_verification' | 'transaction_processing';
  dataCategories: string[];
  userContext: {
    roles: string[];
    permissions: string[];
    authenticationLevel: number;
    authenticationFactors: string[];
    consentReferences?: {
      [key: string]: string;
    };
    geoLocation?: {
      country: string;
      region?: string;
    };
  };
}

/**
 * Interface para resultados de autorização de acesso
 */
export interface AccessAuthorizationResult {
  authorized: boolean;
  complianceVerified: boolean;
  complianceResult?: ConsolidatedComplianceResult;
  requiredAuthenticationLevel?: number;
  requiredAuthenticationFactors?: string[];
  restrictionReasons?: string[];
  requiredActions?: string[];
  requestId: string;
  evaluationTimestamp: number;
}

/**
 * Classe responsável pela integração entre IAM e validadores de conformidade
 */
export class IAMComplianceConnector {
  private complianceIntegrator: ComplianceValidatorIntegrator;
  
  constructor(
    private logger: Logger,
    private metrics: Metrics,
    private tracer: Tracer
  ) {
    this.complianceIntegrator = new ComplianceValidatorIntegrator(logger, metrics, tracer);
  }
  
  /**
   * Verifica conformidade regulatória como parte do processo de autorização de acesso
   * 
   * @param request Requisição de autorização de acesso
   * @returns Resultado de autorização incluindo avaliação de conformidade
   */
  public async authorizeWithComplianceCheck(request: AccessAuthorizationRequest): Promise<AccessAuthorizationResult> {
    const span = this.tracer.startSpan('iam_compliance.authorize_with_compliance_check', {
      attributes: {
        'request.id': request.requestId,
        'user.id': request.userId,
        'tenant.id': request.tenantId,
        'resource.type': request.resourceType,
        'operation.type': request.operationType
      }
    });
    
    try {
      this.logger.info({
        message: 'Iniciando verificação de conformidade para autorização de acesso',
        requestId: request.requestId,
        userId: request.userId,
        tenantId: request.tenantId,
        operationType: request.operationType
      });
      
      const startTime = Date.now();
      
      // Mapeamento para requisição de validação de conformidade
      const validationRequest = this.mapToComplianceRequest(request);
      
      // Executa validação de conformidade
      const complianceResult = await this.complianceIntegrator.validate(validationRequest);
      
      // Avalia resultado e aplica políticas de autorização
      const authorizationResult = this.evaluateAuthorizationWithCompliance(request, complianceResult);
      
      const processingTime = Date.now() - startTime;
      
      // Registra métricas
      this.metrics.timing('iam_compliance.authorization_time', processingTime, {
        tenant_id: request.tenantId,
        operation_type: request.operationType,
        authorized: authorizationResult.authorized.toString()
      });
      
      this.metrics.increment('iam_compliance.authorization_requests', {
        tenant_id: request.tenantId,
        operation_type: request.operationType,
        authorized: authorizationResult.authorized.toString(),
        compliance_verified: authorizationResult.complianceVerified.toString()
      });
      
      this.logger.info({
        message: `Verificação de conformidade concluída: ${authorizationResult.authorized ? 'Autorizado' : 'Negado'}`,
        requestId: request.requestId,
        userId: request.userId,
        tenantId: request.tenantId,
        authorized: authorizationResult.authorized,
        complianceVerified: authorizationResult.complianceVerified,
        processingTimeMs: processingTime
      });
      
      return authorizationResult;
    } catch (error) {
      this.logger.error({
        message: 'Erro durante verificação de conformidade para autorização',
        error: error.message,
        stack: error.stack,
        requestId: request.requestId,
        userId: request.userId,
        tenantId: request.tenantId
      });
      
      this.metrics.increment('iam_compliance.authorization_errors', {
        tenant_id: request.tenantId,
        error_type: error.name || 'unknown'
      });
      
      // Retorna negação de autorização em caso de erro
      return {
        authorized: false,
        complianceVerified: false,
        restrictionReasons: ['Erro durante verificação de conformidade regulatória'],
        requiredActions: ['Contatar administrador do sistema'],
        requestId: request.requestId,
        evaluationTimestamp: Date.now()
      };
    } finally {
      span.end();
    }
  }
  
  /**
   * Mapeia requisição de autorização do IAM para requisição de validação de conformidade
   * 
   * @param request Requisição de autorização do IAM
   * @returns Requisição de validação de conformidade
   */
  private mapToComplianceRequest(request: AccessAuthorizationRequest): ComplianceValidationRequest {
    // Determina país do titular dos dados com base na geolocalização do usuário
    let dataSubjectCountry: Region | string = 'unknown';
    
    if (request.userContext.geoLocation) {
      dataSubjectCountry = this.mapCountryToRegion(request.userContext.geoLocation.country);
    }
    
    // Mapeia tipo de operação para finalidade do processamento
    const purposeMapping = {
      'read': 'Visualização de informações',
      'write': 'Atualização de informações',
      'delete': 'Exclusão de informações',
      'credit_assessment': 'Avaliação de crédito para análise de risco financeiro',
      'identity_verification': 'Verificação de identidade para fins de segurança',
      'transaction_processing': 'Processamento de transação financeira'
    };
    
    // Determina base legal com base no contexto
    let processingLegalBasis = 'consent';
    if (request.userContext.consentReferences && Object.keys(request.userContext.consentReferences).length > 0) {
      processingLegalBasis = 'consent';
    } else if (request.operationType === 'identity_verification') {
      processingLegalBasis = 'legal_obligation';
    } else {
      processingLegalBasis = 'legitimate_interest';
    }
    
    // Mapeia para formato de requisição de conformidade
    return {
      userId: request.userId,
      tenantId: request.tenantId,
      operationId: request.requestId,
      operationType: request.operationType,
      dataSubjectCountry,
      dataProcessingCountry: 'angola', // País onde o sistema está operando
      businessTargetCountries: ['angola', 'brazil', 'south_africa', 'eu', 'mozambique'],
      consentReferences: request.userContext.consentReferences,
      dataPurpose: purposeMapping[request.operationType] || 'Operação não especificada',
      dataCategories: request.dataCategories,
      dataFields: [], // Será preenchido conforme metadados do recurso
      retentionPeriodDays: 730, // Período padrão, poderia vir da configuração do tenant
      processingLegalBasis,
      specialCategories: request.dataCategories.some(cat => 
        ['biometric', 'health', 'political', 'racial', 'religious'].includes(cat)
      ),
      automatedDecisionMaking: request.operationType === 'credit_assessment',
      crossBorderTransfer: false, // Poderia ser determinado com base na localização dos serviços
      securityMeasures: [
        'encryption',
        'access_control',
        'logging',
        'data_minimization',
        'user_authentication'
      ]
    };
  }
  
  /**
   * Avalia resultado de conformidade e determina autorização final
   * 
   * @param request Requisição de autorização original
   * @param complianceResult Resultado da validação de conformidade
   * @returns Resultado de autorização
   */
  private evaluateAuthorizationWithCompliance(
    request: AccessAuthorizationRequest,
    complianceResult: ConsolidatedComplianceResult
  ): AccessAuthorizationResult {
    // Verifica se o processamento é permitido pelos validadores de conformidade
    const complianceAllows = complianceResult.processingAllowed;
    
    // Verifica se o usuário tem permissões de acesso básicas
    const hasPermission = this.userHasPermission(request);
    
    // Determina nível de autenticação necessário com base no tipo de operação e sensibilidade dos dados
    const requiredAuthLevel = this.getRequiredAuthLevel(request, complianceResult);
    
    // Verifica se o nível de autenticação atual é suficiente
    const sufficientAuthLevel = request.userContext.authenticationLevel >= requiredAuthLevel;
    
    // Fatores de autenticação necessários
    const requiredFactors = this.getRequiredAuthFactors(request, complianceResult);
    
    // Verifica se todos os fatores de autenticação necessários estão presentes
    const hasRequiredFactors = requiredFactors.every(factor => 
      request.userContext.authenticationFactors.includes(factor)
    );
    
    // Autorização final: todas as condições devem ser atendidas
    const authorized = complianceAllows && hasPermission && sufficientAuthLevel && hasRequiredFactors;
    
    return {
      authorized,
      complianceVerified: true,
      complianceResult,
      requiredAuthenticationLevel: requiredAuthLevel,
      requiredAuthenticationFactors: requiredFactors,
      restrictionReasons: this.getRestrictionReasons(
        complianceAllows, 
        hasPermission, 
        sufficientAuthLevel, 
        hasRequiredFactors,
        complianceResult
      ),
      requiredActions: complianceResult.requiredActions,
      requestId: request.requestId,
      evaluationTimestamp: Date.now()
    };
  }
  
  /**
   * Mapeia país para região regulatória
   * 
   * @param country Código do país
   * @returns Região regulatória
   */
  private mapCountryToRegion(country: string): Region | string {
    // Mapeamento de países para regiões regulatórias
    const regionMap: { [key: string]: Region } = {
      // União Europeia
      'at': Region.EUROPEAN_UNION, // Áustria
      'be': Region.EUROPEAN_UNION, // Bélgica
      'bg': Region.EUROPEAN_UNION, // Bulgária
      'hr': Region.EUROPEAN_UNION, // Croácia
      'cy': Region.EUROPEAN_UNION, // Chipre
      'cz': Region.EUROPEAN_UNION, // República Tcheca
      'dk': Region.EUROPEAN_UNION, // Dinamarca
      'ee': Region.EUROPEAN_UNION, // Estônia
      'fi': Region.EUROPEAN_UNION, // Finlândia
      'fr': Region.EUROPEAN_UNION, // França
      'de': Region.EUROPEAN_UNION, // Alemanha
      'gr': Region.EUROPEAN_UNION, // Grécia
      'hu': Region.EUROPEAN_UNION, // Hungria
      'ie': Region.EUROPEAN_UNION, // Irlanda
      'it': Region.EUROPEAN_UNION, // Itália
      'lv': Region.EUROPEAN_UNION, // Letônia
      'lt': Region.EUROPEAN_UNION, // Lituânia
      'lu': Region.EUROPEAN_UNION, // Luxemburgo
      'mt': Region.EUROPEAN_UNION, // Malta
      'nl': Region.EUROPEAN_UNION, // Holanda
      'pl': Region.EUROPEAN_UNION, // Polônia
      'pt': Region.EUROPEAN_UNION, // Portugal
      'ro': Region.EUROPEAN_UNION, // Romênia
      'sk': Region.EUROPEAN_UNION, // Eslováquia
      'si': Region.EUROPEAN_UNION, // Eslovênia
      'es': Region.EUROPEAN_UNION, // Espanha
      'se': Region.EUROPEAN_UNION, // Suécia
      
      // Brasil
      'br': Region.BRAZIL,
      'brazil': Region.BRAZIL,
      
      // África do Sul
      'za': Region.SOUTH_AFRICA,
      'south_africa': Region.SOUTH_AFRICA,
      
      // Angola (sem região específica)
      'ao': 'angola',
      'angola': 'angola',
      
      // Moçambique (sem região específica)
      'mz': 'mozambique',
      'mozambique': 'mozambique'
    };
    
    return regionMap[country.toLowerCase()] || country.toLowerCase();
  }
  
  /**
   * Verifica se o usuário tem permissões básicas para o acesso solicitado
   * 
   * @param request Requisição de autorização
   * @returns Booleano indicando se o usuário tem permissão
   */
  private userHasPermission(request: AccessAuthorizationRequest): boolean {
    // Simplificado para este exemplo
    // Em uma implementação real, verificaria as permissões detalhadas
    
    const operationPermissionMap = {
      'read': `${request.resourceType}:read`,
      'write': `${request.resourceType}:write`,
      'delete': `${request.resourceType}:delete`,
      'credit_assessment': 'bureau:credit:assess',
      'identity_verification': 'identity:verify',
      'transaction_processing': 'finance:transaction:process'
    };
    
    const requiredPermission = operationPermissionMap[request.operationType];
    
    return request.userContext.permissions.some(permission => {
      return permission === requiredPermission || permission === '*' || permission === `${request.resourceType}:*`;
    });
  }
  
  /**
   * Determina o nível de autenticação necessário com base na operação e dados
   * 
   * @param request Requisição de autorização
   * @param complianceResult Resultado de conformidade
   * @returns Nível de autenticação necessário (1-4)
   */
  private getRequiredAuthLevel(
    request: AccessAuthorizationRequest,
    complianceResult: ConsolidatedComplianceResult
  ): number {
    // Nível base por tipo de operação
    const baseLevel = {
      'read': 1,
      'write': 2,
      'delete': 3,
      'credit_assessment': 3,
      'identity_verification': 2,
      'transaction_processing': 3
    }[request.operationType] || 1;
    
    // Incremento para dados sensíveis
    const hasSensitiveData = request.dataCategories.some(cat => 
      ['biometric', 'health', 'political', 'racial', 'religious', 'financial', 'identity'].includes(cat)
    );
    
    // Incremento para alta severidade nas validações de conformidade
    const hasHighSeverityCompliance = complianceResult.validationDetails?.some(
      detail => detail.severity === 'high' && !detail.valid
    );
    
    let finalLevel = baseLevel;
    
    if (hasSensitiveData) finalLevel += 1;
    if (hasHighSeverityCompliance) finalLevel += 1;
    
    // Limita o nível a 4 (máximo)
    return Math.min(finalLevel, 4);
  }
  
  /**
   * Determina os fatores de autenticação necessários
   * 
   * @param request Requisição de autorização
   * @param complianceResult Resultado de conformidade
   * @returns Lista de fatores de autenticação necessários
   */
  private getRequiredAuthFactors(
    request: AccessAuthorizationRequest,
    complianceResult: ConsolidatedComplianceResult
  ): string[] {
    const factors = ['password']; // Base: senha sempre necessária
    
    // Para operações de alto risco ou dados sensíveis, requer segundo fator
    if (['delete', 'credit_assessment', 'transaction_processing'].includes(request.operationType)) {
      factors.push('mfa');
    }
    
    // Para dados biométricos ou saúde, requer biometria
    if (request.dataCategories.some(cat => ['biometric', 'health'].includes(cat))) {
      factors.push('biometric');
    }
    
    return factors;
  }
  
  /**
   * Gera lista de razões para restrição de acesso, se houver
   */
  private getRestrictionReasons(
    complianceAllows: boolean,
    hasPermission: boolean,
    sufficientAuthLevel: boolean,
    hasRequiredFactors: boolean,
    complianceResult: ConsolidatedComplianceResult
  ): string[] {
    const reasons = [];
    
    if (!complianceAllows) {
      reasons.push('Restrições de conformidade regulatória impedem esta operação');
      
      if (complianceResult.processingRestrictions && complianceResult.processingRestrictions.length > 0) {
        reasons.push(...complianceResult.processingRestrictions);
      }
    }
    
    if (!hasPermission) {
      reasons.push('Usuário não possui permissão para esta operação');
    }
    
    if (!sufficientAuthLevel) {
      reasons.push('Nível de autenticação insuficiente para esta operação');
    }
    
    if (!hasRequiredFactors) {
      reasons.push('Fatores de autenticação adicionais são necessários');
    }
    
    return reasons;
  }
}