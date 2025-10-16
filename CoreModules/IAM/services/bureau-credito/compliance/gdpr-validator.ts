/**
 * Validador de conformidade com GDPR (General Data Protection Regulation) da União Europeia
 * 
 * Este módulo implementa validações para garantir que o processamento de dados
 * no módulo Bureau de Créditos esteja em conformidade com as exigências do GDPR.
 * 
 * @module GDPRValidator
 */

import { Logger } from '../../../observability/logging/hook_logger';
import { Metrics } from '../../../observability/metrics/hook_metrics';
import { Tracer } from '../../../observability/tracing/hook_tracing';

// Tipos de validações GDPR
export enum GDPRValidationType {
  CONSENT_CHECK = 'consent_check',
  PURPOSE_LIMITATION = 'purpose_limitation',
  DATA_MINIMIZATION = 'data_minimization',
  ACCURACY = 'accuracy',
  STORAGE_LIMITATION = 'storage_limitation',
  RIGHT_TO_ACCESS = 'right_to_access',
  RIGHT_TO_RECTIFICATION = 'right_to_rectification',
  RIGHT_TO_ERASURE = 'right_to_erasure',
  RIGHT_TO_RESTRICTION = 'right_to_restriction',
  RIGHT_TO_PORTABILITY = 'right_to_portability',
  RIGHT_TO_OBJECT = 'right_to_object',
  CROSS_BORDER_TRANSFER = 'cross_border_transfer',
  CHILDREN_DATA = 'children_data',
  SPECIAL_CATEGORIES = 'special_categories',
  DATA_BREACH_NOTIFICATION = 'data_breach_notification'
}

// Finalidades legítimas de processamento conforme GDPR
export enum GDPRProcessingPurpose {
  CONSENT = 'consent',
  CONTRACT_PERFORMANCE = 'contract_performance',
  LEGAL_OBLIGATION = 'legal_obligation',
  VITAL_INTERESTS = 'vital_interests',
  PUBLIC_INTEREST = 'public_interest',
  LEGITIMATE_INTEREST = 'legitimate_interest',
  EXPLICIT_CONSENT_SPECIAL = 'explicit_consent_special'
}

// Interface para informações de consentimento GDPR
export interface GDPRConsent {
  consentId: string;
  userId: string;
  tenantId: string;
  purposes: GDPRProcessingPurpose[];
  dataCategories: string[];
  consentDate: Date;
  expiryDate?: Date;
  withdrawnDate?: Date;
  isActive: boolean;
  proofOfConsent: string;
  consentVersion: string;
}

// Interface para meta-informações sobre processamento GDPR
export interface GDPRProcessingMetadata {
  operationId: string;
  operationType: string;
  dataController: string;
  dataProcessor: string;
  legalBasis: GDPRProcessingPurpose;
  purpose: string;
  dataCategories: string[];
  retentionPeriod: number; // em dias
  thirdPartySharing: boolean;
  crossBorderTransfer: boolean;
  destinationCountries?: string[];
  securityMeasures: string[];
}

// Interface para validação de conformidade GDPR
export interface GDPRValidationRequest {
  userId: string;
  tenantId: string;
  processingMetadata: GDPRProcessingMetadata;
  consentReference?: string;
  dataFields: string[];
  ipAddress?: string;
  countryCode?: string;
  isEUResident?: boolean;
  validationTypes?: GDPRValidationType[];
}

// Interface para resultado de validação individual
export interface GDPRValidationResult {
  validationType: GDPRValidationType;
  passed: boolean;
  failureReason?: string;
  requiredActions?: string[];
  severity: 'low' | 'medium' | 'high' | 'critical';
}

// Interface para resultado completo da validação GDPR
export interface GDPRCompleteValidationResult {
  requestId: string;
  timestamp: Date;
  userId: string;
  tenantId: string;
  overallCompliant: boolean;
  validationResults: GDPRValidationResult[];
  requiredActions: string[];
  processingAllowed: boolean;
  processingRestrictions?: string[];
  auditRecord: {
    validatedBy: string;
    validationTimestamp: Date;
    version: string;
  };
}

/**
 * Classe que implementa validações de conformidade com GDPR
 */
export class GDPRValidator {
  private logger: Logger;
  private metrics: Metrics;
  private tracer: Tracer;
  
  /**
   * Construtor para o validador GDPR
   */
  constructor(logger: Logger, metrics: Metrics, tracer: Tracer) {
    this.logger = logger;
    this.metrics = metrics;
    this.tracer = tracer;
  }
  
  /**
   * Executa validações de conformidade GDPR
   * 
   * @param request Requisição de validação GDPR
   * @returns Resultado da validação
   */
  public async validate(request: GDPRValidationRequest): Promise<GDPRCompleteValidationResult> {
    const span = this.tracer.startSpan('gdpr.validate');
    
    try {
      // Registrar início da validação
      this.logger.info({
        message: 'Iniciando validação de conformidade GDPR',
        userId: request.userId,
        tenantId: request.tenantId,
        processingType: request.processingMetadata.operationType
      });
      
      // Timestamp para métricas
      const startTime = Date.now();
      
      // Determinar quais validações executar
      const validationTypes = request.validationTypes || Object.values(GDPRValidationType);
      
      // Executar todas as validações solicitadas
      const validationResults: GDPRValidationResult[] = [];
      const requiredActions: string[] = [];
      
      // Executar validações individualmente
      for (const validationType of validationTypes) {
        const validationResult = await this.executeValidation(validationType, request);
        validationResults.push(validationResult);
        
        // Se a validação falhou, adicionar ações requeridas
        if (!validationResult.passed && validationResult.requiredActions) {
          requiredActions.push(...validationResult.requiredActions);
        }
      }
      
      // Determinar resultado geral
      const overallCompliant = validationResults.every(result => result.passed);
      
      // Determinar se o processamento pode continuar
      const blockingFailures = validationResults.some(
        result => !result.passed && (result.severity === 'high' || result.severity === 'critical')
      );
      
      const processingAllowed = overallCompliant || !blockingFailures;
      
      // Construir resultado completo
      const result: GDPRCompleteValidationResult = {
        requestId: `gdpr-val-${Date.now()}`,
        timestamp: new Date(),
        userId: request.userId,
        tenantId: request.tenantId,
        overallCompliant,
        validationResults,
        requiredActions: [...new Set(requiredActions)], // Remover duplicatas
        processingAllowed,
        processingRestrictions: processingAllowed && !overallCompliant ? 
          this.determineProcessingRestrictions(validationResults) : undefined,
        auditRecord: {
          validatedBy: 'bureau-credito-gdpr-validator',
          validationTimestamp: new Date(),
          version: '1.0.0'
        }
      };
      
      // Registrar métricas
      const validationTime = Date.now() - startTime;
      
      this.metrics.histogram('bureau_credito.gdpr.validation_time', validationTime, {
        tenant_id: request.tenantId,
        compliant: overallCompliant.toString()
      });
      
      this.metrics.increment('bureau_credito.gdpr.validations_performed', {
        tenant_id: request.tenantId,
        compliant: overallCompliant.toString(),
        processing_allowed: processingAllowed.toString()
      });
      
      // Registrar resultado
      this.logger.info({
        message: `Validação GDPR concluída: ${overallCompliant ? 'Conforme' : 'Não conforme'}`,
        userId: request.userId,
        tenantId: request.tenantId,
        processingType: request.processingMetadata.operationType,
        overallCompliant,
        processingAllowed,
        validationTime
      });
      
      return result;
    } catch (error) {
      // Registrar erro
      this.logger.error({
        message: 'Erro durante validação de conformidade GDPR',
        error: error.message,
        stack: error.stack,
        userId: request.userId,
        tenantId: request.tenantId
      });
      
      // Registrar métrica de erro
      this.metrics.increment('bureau_credito.gdpr.validation_errors', {
        tenant_id: request.tenantId,
        error_type: error.name || 'unknown'
      });
      
      throw error;
    } finally {
      span.end();
    }
  }
  
  /**
   * Executa uma validação específica GDPR
   * 
   * @param validationType Tipo de validação a ser executada
   * @param request Requisição de validação GDPR
   * @returns Resultado da validação específica
   */
  private async executeValidation(
    validationType: GDPRValidationType,
    request: GDPRValidationRequest
  ): Promise<GDPRValidationResult> {
    const span = this.tracer.startSpan('gdpr.execute_validation', { validationType });
    
    try {
      // Escolher a função de validação apropriada com base no tipo
      let validationResult: GDPRValidationResult;
      
      switch (validationType) {
        case GDPRValidationType.CONSENT_CHECK:
          validationResult = await this.validateConsent(request);
          break;
        case GDPRValidationType.PURPOSE_LIMITATION:
          validationResult = await this.validatePurposeLimitation(request);
          break;
        case GDPRValidationType.DATA_MINIMIZATION:
          validationResult = await this.validateDataMinimization(request);
          break;
        case GDPRValidationType.ACCURACY:
          validationResult = await this.validateAccuracy(request);
          break;
        case GDPRValidationType.STORAGE_LIMITATION:
          validationResult = await this.validateStorageLimitation(request);
          break;
        case GDPRValidationType.CROSS_BORDER_TRANSFER:
          validationResult = await this.validateCrossBorderTransfer(request);
          break;
        case GDPRValidationType.SPECIAL_CATEGORIES:
          validationResult = await this.validateSpecialCategories(request);
          break;
        // Adicionar outros casos conforme necessário
        default:
          // Para validações não implementadas, retornar como passadas
          validationResult = {
            validationType,
            passed: true,
            severity: 'low'
          };
      }
      
      // Registrar métrica para a validação específica
      this.metrics.increment('bureau_credito.gdpr.validation_type', {
        validation_type: validationType,
        passed: validationResult.passed.toString(),
        severity: validationResult.severity
      });
      
      return validationResult;
    } finally {
      span.end();
    }
  }
  
  /**
   * Validação de consentimento
   * 
   * @param request Requisição de validação GDPR
   * @returns Resultado da validação
   */
  private async validateConsent(request: GDPRValidationRequest): Promise<GDPRValidationResult> {
    // Verificar se a base legal é consentimento
    if (request.processingMetadata.legalBasis !== GDPRProcessingPurpose.CONSENT) {
      // Se não for baseado em consentimento, esta validação não se aplica
      return {
        validationType: GDPRValidationType.CONSENT_CHECK,
        passed: true,
        severity: 'low'
      };
    }
    
    // Para fins de demonstração, vamos simular uma verificação de consentimento
    // Na implementação real, consultaríamos um serviço de gerenciamento de consentimento
    const consentExists = !!request.consentReference;
    const consentIsValid = consentExists;
    
    if (!consentExists) {
      return {
        validationType: GDPRValidationType.CONSENT_CHECK,
        passed: false,
        failureReason: 'Consentimento não encontrado',
        requiredActions: ['Obter consentimento explícito do usuário antes de continuar o processamento'],
        severity: 'critical'
      };
    }
    
    if (!consentIsValid) {
      return {
        validationType: GDPRValidationType.CONSENT_CHECK,
        passed: false,
        failureReason: 'Consentimento inválido ou expirado',
        requiredActions: ['Renovar consentimento com o usuário'],
        severity: 'critical'
      };
    }
    
    // Consentimento existe e é válido
    return {
      validationType: GDPRValidationType.CONSENT_CHECK,
      passed: true,
      severity: 'low'
    };
  }
  
  /**
   * Validação de limitação de propósito
   * 
   * @param request Requisição de validação GDPR
   * @returns Resultado da validação
   */
  private async validatePurposeLimitation(request: GDPRValidationRequest): Promise<GDPRValidationResult> {
    // Verificar se o propósito está claramente definido
    const hasPurpose = !!request.processingMetadata.purpose;
    
    if (!hasPurpose) {
      return {
        validationType: GDPRValidationType.PURPOSE_LIMITATION,
        passed: false,
        failureReason: 'Propósito de processamento não especificado',
        requiredActions: ['Definir claramente o propósito do processamento de dados'],
        severity: 'high'
      };
    }
    
    // Verificar se a operação é compatível com o propósito declarado
    // Na implementação real, teríamos um mapeamento de operações permitidas por propósito
    const purposeCompatible = true; // Simulado para este exemplo
    
    if (!purposeCompatible) {
      return {
        validationType: GDPRValidationType.PURPOSE_LIMITATION,
        passed: false,
        failureReason: 'Operação incompatível com o propósito declarado',
        requiredActions: ['Revisar e ajustar o propósito ou cancelar a operação'],
        severity: 'high'
      };
    }
    
    return {
      validationType: GDPRValidationType.PURPOSE_LIMITATION,
      passed: true,
      severity: 'medium'
    };
  }
  
  /**
   * Validação de minimização de dados
   * 
   * @param request Requisição de validação GDPR
   * @returns Resultado da validação
   */
  private async validateDataMinimization(request: GDPRValidationRequest): Promise<GDPRValidationResult> {
    // Verificar se todos os campos de dados são realmente necessários para o propósito declarado
    
    // Lista de campos necessários (simulada)
    const requiredFields = ['documentNumber', 'name', 'address'];
    
    // Verificar se há campos desnecessários
    const unnecessaryFields = request.dataFields.filter(field => !requiredFields.includes(field));
    
    if (unnecessaryFields.length > 0) {
      return {
        validationType: GDPRValidationType.DATA_MINIMIZATION,
        passed: false,
        failureReason: 'Campos desnecessários solicitados para o propósito',
        requiredActions: [`Remover os campos desnecessários: ${unnecessaryFields.join(', ')}`],
        severity: 'high'
      };
    }
    
    return {
      validationType: GDPRValidationType.DATA_MINIMIZATION,
      passed: true,
      severity: 'medium'
    };
  }
  
  /**
   * Validação de precisão de dados
   * 
   * @param request Requisição de validação GDPR
   * @returns Resultado da validação
   */
  private async validateAccuracy(request: GDPRValidationRequest): Promise<GDPRValidationResult> {
    // Na implementação real, verificaríamos a origem dos dados e mecanismos de verificação
    // Para este exemplo, vamos simular que todos os dados são precisos
    
    return {
      validationType: GDPRValidationType.ACCURACY,
      passed: true,
      severity: 'medium'
    };
  }
  
  /**
   * Validação de limitação de armazenamento
   * 
   * @param request Requisição de validação GDPR
   * @returns Resultado da validação
   */
  private async validateStorageLimitation(request: GDPRValidationRequest): Promise<GDPRValidationResult> {
    // Verificar se há um período de retenção definido
    if (!request.processingMetadata.retentionPeriod || request.processingMetadata.retentionPeriod <= 0) {
      return {
        validationType: GDPRValidationType.STORAGE_LIMITATION,
        passed: false,
        failureReason: 'Período de retenção não definido ou inválido',
        requiredActions: ['Definir um período de retenção adequado para os dados'],
        severity: 'high'
      };
    }
    
    // Verificar se o período de retenção é razoável para o propósito
    // Para fins de crédito, um período máximo razoável pode ser, por exemplo, 5 anos
    const maxRetentionDays = 5 * 365; // 5 anos em dias
    
    if (request.processingMetadata.retentionPeriod > maxRetentionDays) {
      return {
        validationType: GDPRValidationType.STORAGE_LIMITATION,
        passed: false,
        failureReason: 'Período de retenção excessivamente longo',
        requiredActions: ['Reduzir o período de retenção para um valor razoável'],
        severity: 'medium'
      };
    }
    
    return {
      validationType: GDPRValidationType.STORAGE_LIMITATION,
      passed: true,
      severity: 'medium'
    };
  }
  
  /**
   * Validação de transferência internacional de dados
   * 
   * @param request Requisição de validação GDPR
   * @returns Resultado da validação
   */
  private async validateCrossBorderTransfer(request: GDPRValidationRequest): Promise<GDPRValidationResult> {
    // Verificar se há transferência internacional
    if (!request.processingMetadata.crossBorderTransfer) {
      return {
        validationType: GDPRValidationType.CROSS_BORDER_TRANSFER,
        passed: true,
        severity: 'medium'
      };
    }
    
    // Verificar se os países de destino estão definidos
    if (!request.processingMetadata.destinationCountries || request.processingMetadata.destinationCountries.length === 0) {
      return {
        validationType: GDPRValidationType.CROSS_BORDER_TRANSFER,
        passed: false,
        failureReason: 'Transferência internacional sem países de destino especificados',
        requiredActions: ['Especificar os países de destino para transferência internacional'],
        severity: 'high'
      };
    }
    
    // Lista de países com adequação GDPR (simplificada para o exemplo)
    const adequateCountries = ['AT', 'BE', 'BG', 'HR', 'CY', 'CZ', 'DK', 'EE', 'FI', 'FR', 'DE', 'GR', 'HU', 'IE', 'IT', 'LV', 'LT', 'LU', 'MT', 'NL', 'PL', 'PT', 'RO', 'SK', 'SI', 'ES', 'SE', 'GB', 'IS', 'LI', 'NO', 'CH', 'AR', 'CA', 'IL', 'JP', 'NZ', 'KR', 'UY'];
    
    // Verificar se todos os países de destino são adequados
    const inadequateCountries = request.processingMetadata.destinationCountries.filter(
      country => !adequateCountries.includes(country)
    );
    
    if (inadequateCountries.length > 0) {
      return {
        validationType: GDPRValidationType.CROSS_BORDER_TRANSFER,
        passed: false,
        failureReason: 'Transferência internacional para países sem adequação GDPR',
        requiredActions: [
          `Implementar salvaguardas adicionais para transferências para: ${inadequateCountries.join(', ')}`,
          'Considerar cláusulas contratuais padrão ou outros mecanismos aprovados'
        ],
        severity: 'high'
      };
    }
    
    return {
      validationType: GDPRValidationType.CROSS_BORDER_TRANSFER,
      passed: true,
      severity: 'medium'
    };
  }
  
  /**
   * Validação de categorias especiais de dados
   * 
   * @param request Requisição de validação GDPR
   * @returns Resultado da validação
   */
  private async validateSpecialCategories(request: GDPRValidationRequest): Promise<GDPRValidationResult> {
    // Lista de categorias especiais de dados conforme GDPR
    const specialCategories = [
      'racial_origin', 'ethnic_origin', 'political_opinions', 'religious_beliefs',
      'philosophical_beliefs', 'trade_union_membership', 'genetic_data', 'biometric_data',
      'health_data', 'sexual_orientation', 'sex_life'
    ];
    
    // Verificar se há categorias especiais sendo processadas
    const specialCategoriesProcessed = request.processingMetadata.dataCategories.some(
      category => specialCategories.includes(category.toLowerCase())
    );
    
    if (!specialCategoriesProcessed) {
      // Se não há categorias especiais, essa validação passa
      return {
        validationType: GDPRValidationType.SPECIAL_CATEGORIES,
        passed: true,
        severity: 'medium'
      };
    }
    
    // Se há categorias especiais, verificar se a base legal é adequada
    const hasValidLegalBasis = request.processingMetadata.legalBasis === GDPRProcessingPurpose.EXPLICIT_CONSENT_SPECIAL;
    
    if (!hasValidLegalBasis) {
      return {
        validationType: GDPRValidationType.SPECIAL_CATEGORIES,
        passed: false,
        failureReason: 'Processamento de categorias especiais de dados sem base legal adequada',
        requiredActions: [
          'Obter consentimento explícito para processamento de categorias especiais',
          'Ou cancelar o processamento dessas categorias de dados'
        ],
        severity: 'critical'
      };
    }
    
    return {
      validationType: GDPRValidationType.SPECIAL_CATEGORIES,
      passed: true,
      severity: 'high'
    };
  }
  
  /**
   * Determina restrições de processamento com base nos resultados da validação
   * 
   * @param validationResults Resultados da validação
   * @returns Lista de restrições de processamento
   */
  private determineProcessingRestrictions(validationResults: GDPRValidationResult[]): string[] {
    const restrictions: string[] = [];
    
    // Analisar falhas e determinar restrições apropriadas
    for (const result of validationResults) {
      if (!result.passed) {
        switch (result.validationType) {
          case GDPRValidationType.DATA_MINIMIZATION:
            restrictions.push('Processamento limitado apenas aos campos obrigatórios');
            break;
          case GDPRValidationType.STORAGE_LIMITATION:
            restrictions.push('Período de retenção reduzido para máximo permitido');
            break;
          case GDPRValidationType.SPECIAL_CATEGORIES:
            restrictions.push('Categorias especiais de dados excluídas do processamento');
            break;
          case GDPRValidationType.CROSS_BORDER_TRANSFER:
            restrictions.push('Transferência internacional limitada apenas a países com adequação');
            break;
          // Adicionar outros casos conforme necessário
        }
      }
    }
    
    return [...new Set(restrictions)]; // Remover duplicatas
  }
}