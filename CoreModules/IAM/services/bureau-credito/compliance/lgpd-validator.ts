/**
 * Validador de conformidade com LGPD (Lei Geral de Proteção de Dados) do Brasil
 * 
 * Este módulo implementa validações para garantir que o processamento de dados
 * no módulo Bureau de Créditos esteja em conformidade com as exigências da LGPD.
 * 
 * @module LGPDValidator
 */

import { Logger } from '../../../observability/logging/hook_logger';
import { Metrics } from '../../../observability/metrics/hook_metrics';
import { Tracer } from '../../../observability/tracing/hook_tracing';

// Tipos de validações LGPD
export enum LGPDValidationType {
  LEGAL_BASIS_CHECK = 'legal_basis_check',
  PURPOSE_LIMITATION = 'purpose_limitation',
  DATA_MINIMIZATION = 'data_minimization',
  DATA_QUALITY = 'data_quality',
  RETENTION_LIMITATION = 'retention_limitation',
  RIGHT_TO_ACCESS = 'right_to_access',
  RIGHT_TO_CORRECTION = 'right_to_correction',
  RIGHT_TO_DELETION = 'right_to_deletion',
  RIGHT_TO_PORTABILITY = 'right_to_portability',
  RIGHT_TO_INFORMATION = 'right_to_information',
  RIGHT_TO_REVOCATION = 'right_to_revocation',
  RIGHT_TO_OBJECT = 'right_to_object',
  SECURITY_MEASURES = 'security_measures',
  SENSITIVE_DATA = 'sensitive_data',
  CHILDREN_DATA = 'children_data',
  DATA_PROCESSING_RECORDS = 'data_processing_records',
  ANPD_REGULATIONS = 'anpd_regulations'
}

// Bases legais para processamento conforme LGPD
export enum LGPDLegalBasis {
  CONSENT = 'consent',
  LEGAL_OBLIGATION = 'legal_obligation',
  CONTRACT_EXECUTION = 'contract_execution',
  LEGITIMATE_INTEREST = 'legitimate_interest',
  PUBLIC_POLICY_EXECUTION = 'public_policy_execution',
  STUDIES_BY_RESEARCH_ENTITY = 'studies_by_research_entity',
  CREDIT_PROTECTION = 'credit_protection',
  REGULAR_EXERCISE_OF_RIGHTS = 'regular_exercise_of_rights',
  HEALTH_PROTECTION = 'health_protection',
  VITAL_INTEREST_PROTECTION = 'vital_interest_protection'
}

// Interface para informações de consentimento LGPD
export interface LGPDConsent {
  consentId: string;
  userId: string;
  tenantId: string;
  purposes: string[];
  dataCategories: string[];
  consentDate: Date;
  expiryDate?: Date;
  withdrawnDate?: Date;
  isActive: boolean;
  proofOfConsent: string;
  consentVersion: string;
  freeSpecificInformed: boolean;
  unambiguous: boolean;
}

// Interface para meta-informações sobre processamento LGPD
export interface LGPDProcessingMetadata {
  operationId: string;
  operationType: string;
  dataController: string;
  dataOperator: string;
  dpo: string;
  legalBasis: LGPDLegalBasis;
  purpose: string;
  dataCategories: string[];
  retentionPeriod: number; // em dias
  thirdPartySharing: boolean;
  internationalTransfer: boolean;
  destinationCountries?: string[];
  securityMeasures: string[];
}

// Interface para validação de conformidade LGPD
export interface LGPDValidationRequest {
  userId: string;
  tenantId: string;
  processingMetadata: LGPDProcessingMetadata;
  consentReference?: string;
  dataFields: string[];
  validationTypes?: LGPDValidationType[];
  isBrazilianResident?: boolean;
}

// Interface para resultado de validação individual
export interface LGPDValidationResult {
  validationType: LGPDValidationType;
  passed: boolean;
  failureReason?: string;
  requiredActions?: string[];
  severity: 'low' | 'medium' | 'high' | 'critical';
}

// Interface para resultado completo da validação LGPD
export interface LGPDCompleteValidationResult {
  requestId: string;
  timestamp: Date;
  userId: string;
  tenantId: string;
  overallCompliant: boolean;
  validationResults: LGPDValidationResult[];
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
 * Classe que implementa validações de conformidade com LGPD
 */
export class LGPDValidator {
  private logger: Logger;
  private metrics: Metrics;
  private tracer: Tracer;
  
  /**
   * Construtor para o validador LGPD
   */
  constructor(logger: Logger, metrics: Metrics, tracer: Tracer) {
    this.logger = logger;
    this.metrics = metrics;
    this.tracer = tracer;
  }
  
  /**
   * Executa validações de conformidade LGPD
   * 
   * @param request Requisição de validação LGPD
   * @returns Resultado da validação
   */
  public async validate(request: LGPDValidationRequest): Promise<LGPDCompleteValidationResult> {
    const span = this.tracer.startSpan('lgpd.validate');
    
    try {
      // Registrar início da validação
      this.logger.info({
        message: 'Iniciando validação de conformidade LGPD',
        userId: request.userId,
        tenantId: request.tenantId,
        processingType: request.processingMetadata.operationType
      });
      
      // Timestamp para métricas
      const startTime = Date.now();
      
      // Determinar quais validações executar
      const validationTypes = request.validationTypes || Object.values(LGPDValidationType);
      
      // Executar todas as validações solicitadas
      const validationResults: LGPDValidationResult[] = [];
      const requiredActions: string[] = [];
      
      // Verificar se a LGPD é aplicável
      const isLGPDApplicable = this.isLGPDApplicable(request);
      
      if (!isLGPDApplicable) {
        // Se a LGPD não se aplica, retornar um resultado positivo simplificado
        const result: LGPDCompleteValidationResult = {
          requestId: `lgpd-val-${Date.now()}`,
          timestamp: new Date(),
          userId: request.userId,
          tenantId: request.tenantId,
          overallCompliant: true,
          validationResults: [{
            validationType: LGPDValidationType.LEGAL_BASIS_CHECK,
            passed: true,
            severity: 'low'
          }],
          requiredActions: [],
          processingAllowed: true,
          auditRecord: {
            validatedBy: 'bureau-credito-lgpd-validator',
            validationTimestamp: new Date(),
            version: '1.0.0'
          }
        };
        
        return result;
      }
      
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
      const result: LGPDCompleteValidationResult = {
        requestId: `lgpd-val-${Date.now()}`,
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
          validatedBy: 'bureau-credito-lgpd-validator',
          validationTimestamp: new Date(),
          version: '1.0.0'
        }
      };
      
      // Registrar métricas
      const validationTime = Date.now() - startTime;
      
      this.metrics.histogram('bureau_credito.lgpd.validation_time', validationTime, {
        tenant_id: request.tenantId,
        compliant: overallCompliant.toString()
      });
      
      this.metrics.increment('bureau_credito.lgpd.validations_performed', {
        tenant_id: request.tenantId,
        compliant: overallCompliant.toString(),
        processing_allowed: processingAllowed.toString()
      });
      
      // Registrar resultado
      this.logger.info({
        message: `Validação LGPD concluída: ${overallCompliant ? 'Conforme' : 'Não conforme'}`,
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
        message: 'Erro durante validação de conformidade LGPD',
        error: error.message,
        stack: error.stack,
        userId: request.userId,
        tenantId: request.tenantId
      });
      
      // Registrar métrica de erro
      this.metrics.increment('bureau_credito.lgpd.validation_errors', {
        tenant_id: request.tenantId,
        error_type: error.name || 'unknown'
      });
      
      throw error;
    } finally {
      span.end();
    }
  }
  
  /**
   * Verifica se a LGPD é aplicável para o processamento
   * 
   * @param request Requisição de validação LGPD
   * @returns Verdadeiro se a LGPD for aplicável
   */
  private isLGPDApplicable(request: LGPDValidationRequest): boolean {
    // LGPD é aplicável se:
    // 1. O usuário é residente brasileiro (se a informação estiver disponível)
    // 2. OU se o processamento ocorre em território brasileiro
    // 3. OU se o serviço é oferecido para o mercado brasileiro
    
    // Para fins de demonstração, vamos considerar aplicável se explicitamente marcado como residente brasileiro
    // ou se não temos essa informação (assumimos que sim para segurança)
    return request.isBrazilianResident !== false;
  }
  
  /**
   * Executa uma validação específica LGPD
   * 
   * @param validationType Tipo de validação a ser executada
   * @param request Requisição de validação LGPD
   * @returns Resultado da validação específica
   */
  private async executeValidation(
    validationType: LGPDValidationType,
    request: LGPDValidationRequest
  ): Promise<LGPDValidationResult> {
    const span = this.tracer.startSpan('lgpd.execute_validation', { validationType });
    
    try {
      // Escolher a função de validação apropriada com base no tipo
      let validationResult: LGPDValidationResult;
      
      switch (validationType) {
        case LGPDValidationType.LEGAL_BASIS_CHECK:
          validationResult = await this.validateLegalBasis(request);
          break;
        case LGPDValidationType.PURPOSE_LIMITATION:
          validationResult = await this.validatePurposeLimitation(request);
          break;
        case LGPDValidationType.DATA_MINIMIZATION:
          validationResult = await this.validateDataMinimization(request);
          break;
        case LGPDValidationType.DATA_QUALITY:
          validationResult = await this.validateDataQuality(request);
          break;
        case LGPDValidationType.RETENTION_LIMITATION:
          validationResult = await this.validateRetentionLimitation(request);
          break;
        case LGPDValidationType.SENSITIVE_DATA:
          validationResult = await this.validateSensitiveData(request);
          break;
        case LGPDValidationType.SECURITY_MEASURES:
          validationResult = await this.validateSecurityMeasures(request);
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
      this.metrics.increment('bureau_credito.lgpd.validation_type', {
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
   * Validação de base legal
   * 
   * @param request Requisição de validação LGPD
   * @returns Resultado da validação
   */
  private async validateLegalBasis(request: LGPDValidationRequest): Promise<LGPDValidationResult> {
    // Verificar se uma base legal foi especificada
    if (!request.processingMetadata.legalBasis) {
      return {
        validationType: LGPDValidationType.LEGAL_BASIS_CHECK,
        passed: false,
        failureReason: 'Base legal não especificada',
        requiredActions: ['Definir uma base legal válida para o processamento de dados'],
        severity: 'critical'
      };
    }
    
    // Verificar se a base legal é válida para Bureau de Créditos
    // Para crédito, as bases válidas são: consentimento, proteção ao crédito, legítimo interesse e obrigação legal
    const validBases = [
      LGPDLegalBasis.CONSENT,
      LGPDLegalBasis.CREDIT_PROTECTION,
      LGPDLegalBasis.LEGITIMATE_INTEREST,
      LGPDLegalBasis.LEGAL_OBLIGATION
    ];
    
    if (!validBases.includes(request.processingMetadata.legalBasis)) {
      return {
        validationType: LGPDValidationType.LEGAL_BASIS_CHECK,
        passed: false,
        failureReason: 'Base legal incompatível com operações de bureau de crédito',
        requiredActions: ['Revisar e ajustar a base legal para uma das bases válidas para operações de crédito'],
        severity: 'critical'
      };
    }
    
    // Se a base legal é consentimento, verificar se o consentimento existe
    if (request.processingMetadata.legalBasis === LGPDLegalBasis.CONSENT) {
      if (!request.consentReference) {
        return {
          validationType: LGPDValidationType.LEGAL_BASIS_CHECK,
          passed: false,
          failureReason: 'Base legal é consentimento, mas referência de consentimento não fornecida',
          requiredActions: ['Obter e registrar consentimento explícito do usuário'],
          severity: 'critical'
        };
      }
      
      // Na implementação real, verificaríamos a validade do consentimento
      // aqui consultando um serviço de gerenciamento de consentimento
    }
    
    // Se a base legal é proteção ao crédito, verificar finalidade específica
    if (request.processingMetadata.legalBasis === LGPDLegalBasis.CREDIT_PROTECTION) {
      const isCreditProtectionPurpose = request.processingMetadata.purpose.toLowerCase().includes('credito') ||
                                       request.processingMetadata.purpose.toLowerCase().includes('crédito') ||
                                       request.processingMetadata.purpose.toLowerCase().includes('fraude');
      
      if (!isCreditProtectionPurpose) {
        return {
          validationType: LGPDValidationType.LEGAL_BASIS_CHECK,
          passed: false,
          failureReason: 'Base legal é proteção ao crédito, mas finalidade não está claramente relacionada',
          requiredActions: ['Ajustar a finalidade para explicitar a relação com proteção ao crédito'],
          severity: 'high'
        };
      }
    }
    
    return {
      validationType: LGPDValidationType.LEGAL_BASIS_CHECK,
      passed: true,
      severity: 'critical'
    };
  }
  
  /**
   * Validação de limitação de finalidade
   * 
   * @param request Requisição de validação LGPD
   * @returns Resultado da validação
   */
  private async validatePurposeLimitation(request: LGPDValidationRequest): Promise<LGPDValidationResult> {
    // Verificar se a finalidade está claramente definida
    if (!request.processingMetadata.purpose || request.processingMetadata.purpose.trim().length === 0) {
      return {
        validationType: LGPDValidationType.PURPOSE_LIMITATION,
        passed: false,
        failureReason: 'Finalidade de processamento não especificada',
        requiredActions: ['Definir claramente a finalidade do processamento de dados'],
        severity: 'high'
      };
    }
    
    // Verificar se a finalidade é específica e explícita
    // (na implementação real, poderíamos verificar contra um catálogo de finalidades autorizadas)
    const isPurposeSpecific = request.processingMetadata.purpose.length > 10;
    
    if (!isPurposeSpecific) {
      return {
        validationType: LGPDValidationType.PURPOSE_LIMITATION,
        passed: false,
        failureReason: 'Finalidade de processamento não é específica o suficiente',
        requiredActions: ['Definir finalidade mais específica e explícita para o processamento'],
        severity: 'medium'
      };
    }
    
    // Verificar se a operação é compatível com a finalidade declarada
    // Na implementação real, teríamos um mapeamento de operações permitidas por finalidade
    const purposeCompatible = true; // Simulado para este exemplo
    
    if (!purposeCompatible) {
      return {
        validationType: LGPDValidationType.PURPOSE_LIMITATION,
        passed: false,
        failureReason: 'Operação incompatível com a finalidade declarada',
        requiredActions: ['Revisar e ajustar a finalidade ou cancelar a operação'],
        severity: 'high'
      };
    }
    
    return {
      validationType: LGPDValidationType.PURPOSE_LIMITATION,
      passed: true,
      severity: 'medium'
    };
  }
  
  /**
   * Validação de minimização de dados
   * 
   * @param request Requisição de validação LGPD
   * @returns Resultado da validação
   */
  private async validateDataMinimization(request: LGPDValidationRequest): Promise<LGPDValidationResult> {
    // Verificar se todos os campos de dados são realmente necessários para a finalidade declarada
    
    // Lista de campos necessários com base na finalidade (simulada)
    let requiredFields: string[] = [];
    
    // Determinar campos necessários com base na operação
    if (request.processingMetadata.operationType.includes('avaliacao_credito')) {
      requiredFields = ['documentNumber', 'name', 'income', 'address'];
    } else if (request.processingMetadata.operationType.includes('deteccao_fraude')) {
      requiredFields = ['documentNumber', 'deviceId', 'ipAddress', 'transactionHistory'];
    } else {
      // Default para outras operações
      requiredFields = ['documentNumber', 'name'];
    }
    
    // Verificar se há campos desnecessários
    const unnecessaryFields = request.dataFields.filter(field => !requiredFields.includes(field));
    
    if (unnecessaryFields.length > 0) {
      return {
        validationType: LGPDValidationType.DATA_MINIMIZATION,
        passed: false,
        failureReason: 'Campos desnecessários solicitados para a finalidade',
        requiredActions: [`Remover os campos desnecessários: ${unnecessaryFields.join(', ')}`],
        severity: 'high'
      };
    }
    
    return {
      validationType: LGPDValidationType.DATA_MINIMIZATION,
      passed: true,
      severity: 'medium'
    };
  }
  
  /**
   * Validação de qualidade de dados
   * 
   * @param request Requisição de validação LGPD
   * @returns Resultado da validação
   */
  private async validateDataQuality(request: LGPDValidationRequest): Promise<LGPDValidationResult> {
    // Na implementação real, verificaríamos mecanismos para garantir qualidade dos dados
    // Para este exemplo, vamos considerar que existem processos adequados
    
    return {
      validationType: LGPDValidationType.DATA_QUALITY,
      passed: true,
      severity: 'medium'
    };
  }
  
  /**
   * Validação de limitação de retenção
   * 
   * @param request Requisição de validação LGPD
   * @returns Resultado da validação
   */
  private async validateRetentionLimitation(request: LGPDValidationRequest): Promise<LGPDValidationResult> {
    // Verificar se há um período de retenção definido
    if (!request.processingMetadata.retentionPeriod || request.processingMetadata.retentionPeriod <= 0) {
      return {
        validationType: LGPDValidationType.RETENTION_LIMITATION,
        passed: false,
        failureReason: 'Período de retenção não definido ou inválido',
        requiredActions: ['Definir um período de retenção adequado para os dados'],
        severity: 'high'
      };
    }
    
    // Verificar se o período de retenção é razoável para a finalidade
    // Para dados de crédito no Brasil, o período máximo é geralmente 5 anos após
    // o vencimento da obrigação, conforme definido pelo CDC (Código de Defesa do Consumidor)
    const maxRetentionDays = 5 * 365; // 5 anos em dias
    
    if (request.processingMetadata.retentionPeriod > maxRetentionDays) {
      return {
        validationType: LGPDValidationType.RETENTION_LIMITATION,
        passed: false,
        failureReason: 'Período de retenção excessivamente longo',
        requiredActions: ['Reduzir o período de retenção para conformidade com o limite de 5 anos'],
        severity: 'high'
      };
    }
    
    return {
      validationType: LGPDValidationType.RETENTION_LIMITATION,
      passed: true,
      severity: 'medium'
    };
  }
  
  /**
   * Validação de dados sensíveis
   * 
   * @param request Requisição de validação LGPD
   * @returns Resultado da validação
   */
  private async validateSensitiveData(request: LGPDValidationRequest): Promise<LGPDValidationResult> {
    // Lista de categorias de dados sensíveis conforme LGPD
    const sensitiveCategories = [
      'origem_racial', 'origem_etnica', 'convicçao_religiosa', 'opiniao_politica',
      'filiacao_sindical', 'dados_geneticos', 'dados_biometricos', 'dados_saude',
      'vida_sexual', 'orientacao_sexual'
    ];
    
    // Verificar se há categorias sensíveis sendo processadas
    const sensitiveCategoriesProcessed = request.processingMetadata.dataCategories.some(
      category => sensitiveCategories.includes(category.toLowerCase())
    );
    
    if (!sensitiveCategoriesProcessed) {
      // Se não há categorias sensíveis, essa validação passa
      return {
        validationType: LGPDValidationType.SENSITIVE_DATA,
        passed: true,
        severity: 'high'
      };
    }
    
    // Se há categorias sensíveis, verificar se a base legal é adequada
    const hasValidLegalBasis = request.processingMetadata.legalBasis === LGPDLegalBasis.CONSENT ||
                              request.processingMetadata.legalBasis === LGPDLegalBasis.LEGAL_OBLIGATION;
    
    if (!hasValidLegalBasis) {
      return {
        validationType: LGPDValidationType.SENSITIVE_DATA,
        passed: false,
        failureReason: 'Processamento de dados sensíveis sem base legal adequada',
        requiredActions: [
          'Obter consentimento específico e destacado para processamento de dados sensíveis',
          'Ou verificar se há obrigação legal específica que permita o processamento',
          'Ou cancelar o processamento desses dados sensíveis'
        ],
        severity: 'critical'
      };
    }
    
    // Se a base é consentimento, verificar se existe referência ao consentimento
    if (request.processingMetadata.legalBasis === LGPDLegalBasis.CONSENT && !request.consentReference) {
      return {
        validationType: LGPDValidationType.SENSITIVE_DATA,
        passed: false,
        failureReason: 'Processamento de dados sensíveis sem referência ao consentimento específico',
        requiredActions: ['Obter e registrar consentimento específico e destacado para dados sensíveis'],
        severity: 'critical'
      };
    }
    
    return {
      validationType: LGPDValidationType.SENSITIVE_DATA,
      passed: true,
      severity: 'high'
    };
  }
  
  /**
   * Validação de medidas de segurança
   * 
   * @param request Requisição de validação LGPD
   * @returns Resultado da validação
   */
  private async validateSecurityMeasures(request: LGPDValidationRequest): Promise<LGPDValidationResult> {
    // Verificar se há medidas de segurança definidas
    if (!request.processingMetadata.securityMeasures || request.processingMetadata.securityMeasures.length === 0) {
      return {
        validationType: LGPDValidationType.SECURITY_MEASURES,
        passed: false,
        failureReason: 'Medidas de segurança não especificadas',
        requiredActions: ['Definir e implementar medidas técnicas e administrativas de segurança'],
        severity: 'high'
      };
    }
    
    // Medidas de segurança mínimas esperadas
    const requiredMeasures = ['encryption', 'access_control', 'logging'];
    
    // Verificar se todas as medidas mínimas estão presentes
    const missingMeasures = requiredMeasures.filter(
      measure => !request.processingMetadata.securityMeasures.some(
        m => m.toLowerCase().includes(measure)
      )
    );
    
    if (missingMeasures.length > 0) {
      return {
        validationType: LGPDValidationType.SECURITY_MEASURES,
        passed: false,
        failureReason: 'Medidas de segurança insuficientes',
        requiredActions: [`Implementar medidas de segurança adicionais: ${missingMeasures.join(', ')}`],
        severity: 'high'
      };
    }
    
    return {
      validationType: LGPDValidationType.SECURITY_MEASURES,
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
  private determineProcessingRestrictions(validationResults: LGPDValidationResult[]): string[] {
    const restrictions: string[] = [];
    
    // Analisar falhas e determinar restrições apropriadas
    for (const result of validationResults) {
      if (!result.passed) {
        switch (result.validationType) {
          case LGPDValidationType.DATA_MINIMIZATION:
            restrictions.push('Processamento limitado apenas aos campos estritamente necessários');
            break;
          case LGPDValidationType.RETENTION_LIMITATION:
            restrictions.push('Período de retenção reduzido para máximo permitido (5 anos)');
            break;
          case LGPDValidationType.SENSITIVE_DATA:
            restrictions.push('Dados sensíveis excluídos do processamento');
            break;
          case LGPDValidationType.SECURITY_MEASURES:
            restrictions.push('Acesso aos resultados limitado a usuários autorizados com autenticação forte');
            break;
          // Adicionar outros casos conforme necessário
        }
      }
    }
    
    return [...new Set(restrictions)]; // Remover duplicatas
  }
}