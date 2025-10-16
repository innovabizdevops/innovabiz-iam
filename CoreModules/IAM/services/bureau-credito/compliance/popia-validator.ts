/**
 * Validador de conformidade com POPIA (Protection of Personal Information Act) da África do Sul
 * 
 * Este módulo implementa validações para garantir que o processamento de dados
 * no módulo Bureau de Créditos esteja em conformidade com as exigências do POPIA.
 * 
 * @module POPIAValidator
 */

import { Logger } from '../../../observability/logging/hook_logger';
import { Metrics } from '../../../observability/metrics/hook_metrics';
import { Tracer } from '../../../observability/tracing/hook_tracing';

// Tipos de validações POPIA
export enum POPIAValidationType {
  LAWFULNESS_CHECK = 'lawfulness_check',
  MINIMALITY = 'minimality',
  PURPOSE_SPECIFICATION = 'purpose_specification',
  FURTHER_PROCESSING_LIMITATION = 'further_processing_limitation',
  INFORMATION_QUALITY = 'information_quality',
  OPENNESS = 'openness',
  SECURITY_SAFEGUARDS = 'security_safeguards',
  DATA_SUBJECT_PARTICIPATION = 'data_subject_participation',
  CROSS_BORDER_TRANSFER = 'cross_border_transfer',
  SPECIAL_PERSONAL_INFO = 'special_personal_info',
  CHILDRENS_PERSONAL_INFO = 'childrens_personal_info',
  DIRECT_MARKETING = 'direct_marketing',
  AUTOMATED_DECISION_MAKING = 'automated_decision_making',
  ACCOUNTABILITY = 'accountability'
}

// Bases legítimas para processamento conforme POPIA
export enum POPIALawfulBasis {
  CONSENT = 'consent',
  CONTRACT_PERFORMANCE = 'contract_performance',
  LEGAL_OBLIGATION = 'legal_obligation',
  LEGITIMATE_INTEREST = 'legitimate_interest',
  PUBLIC_INTEREST = 'public_interest',
  VITAL_INTEREST = 'vital_interest',
  EXPLICIT_CONSENT_SPECIAL = 'explicit_consent_special',
  CREDIT_REPORTING = 'credit_reporting'
}

// Interface para informações de consentimento POPIA
export interface POPIAConsent {
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
  voluntary: boolean;
  specific: boolean;
  informed: boolean;
}

// Interface para meta-informações sobre processamento POPIA
export interface POPIAProcessingMetadata {
  operationId: string;
  operationType: string;
  responsibleParty: string;
  operator: string;
  informationOfficer: string;
  lawfulBasis: POPIALawfulBasis;
  purpose: string;
  dataCategories: string[];
  retentionPeriod: number; // em dias
  thirdPartySharing: boolean;
  crossBorderTransfer: boolean;
  destinationCountries?: string[];
  securityMeasures: string[];
  automatedDecisionMaking: boolean;
}

// Interface para validação de conformidade POPIA
export interface POPIAValidationRequest {
  userId: string;
  tenantId: string;
  processingMetadata: POPIAProcessingMetadata;
  consentReference?: string;
  dataFields: string[];
  validationTypes?: POPIAValidationType[];
  isSouthAfricanResident?: boolean;
}

// Interface para resultado de validação individual
export interface POPIAValidationResult {
  validationType: POPIAValidationType;
  passed: boolean;
  failureReason?: string;
  requiredActions?: string[];
  severity: 'low' | 'medium' | 'high' | 'critical';
}

// Interface para resultado completo da validação POPIA
export interface POPIACompleteValidationResult {
  requestId: string;
  timestamp: Date;
  userId: string;
  tenantId: string;
  overallCompliant: boolean;
  validationResults: POPIAValidationResult[];
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
 * Classe que implementa validações de conformidade com POPIA
 */
export class POPIAValidator {
  private logger: Logger;
  private metrics: Metrics;
  private tracer: Tracer;
  
  /**
   * Construtor para o validador POPIA
   */
  constructor(logger: Logger, metrics: Metrics, tracer: Tracer) {
    this.logger = logger;
    this.metrics = metrics;
    this.tracer = tracer;
  }
  
  /**
   * Executa validações de conformidade POPIA
   * 
   * @param request Requisição de validação POPIA
   * @returns Resultado da validação
   */
  public async validate(request: POPIAValidationRequest): Promise<POPIACompleteValidationResult> {
    const span = this.tracer.startSpan('popia.validate');
    
    try {
      // Registrar início da validação
      this.logger.info({
        message: 'Iniciando validação de conformidade POPIA',
        userId: request.userId,
        tenantId: request.tenantId,
        processingType: request.processingMetadata.operationType
      });
      
      // Timestamp para métricas
      const startTime = Date.now();
      
      // Verificar se o POPIA é aplicável
      const isPOPIAApplicable = this.isPOPIAApplicable(request);
      
      if (!isPOPIAApplicable) {
        // Se o POPIA não se aplica, retornar um resultado positivo simplificado
        const result: POPIACompleteValidationResult = {
          requestId: `popia-val-${Date.now()}`,
          timestamp: new Date(),
          userId: request.userId,
          tenantId: request.tenantId,
          overallCompliant: true,
          validationResults: [{
            validationType: POPIAValidationType.LAWFULNESS_CHECK,
            passed: true,
            severity: 'low'
          }],
          requiredActions: [],
          processingAllowed: true,
          auditRecord: {
            validatedBy: 'bureau-credito-popia-validator',
            validationTimestamp: new Date(),
            version: '1.0.0'
          }
        };
        
        return result;
      }
      
      // Determinar quais validações executar
      const validationTypes = request.validationTypes || Object.values(POPIAValidationType);
      
      // Executar todas as validações solicitadas
      const validationResults: POPIAValidationResult[] = [];
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
      const result: POPIACompleteValidationResult = {
        requestId: `popia-val-${Date.now()}`,
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
          validatedBy: 'bureau-credito-popia-validator',
          validationTimestamp: new Date(),
          version: '1.0.0'
        }
      };
      
      // Registrar métricas
      const validationTime = Date.now() - startTime;
      
      this.metrics.histogram('bureau_credito.popia.validation_time', validationTime, {
        tenant_id: request.tenantId,
        compliant: overallCompliant.toString()
      });
      
      this.metrics.increment('bureau_credito.popia.validations_performed', {
        tenant_id: request.tenantId,
        compliant: overallCompliant.toString(),
        processing_allowed: processingAllowed.toString()
      });
      
      // Registrar resultado
      this.logger.info({
        message: `Validação POPIA concluída: ${overallCompliant ? 'Conforme' : 'Não conforme'}`,
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
        message: 'Erro durante validação de conformidade POPIA',
        error: error.message,
        stack: error.stack,
        userId: request.userId,
        tenantId: request.tenantId
      });
      
      // Registrar métrica de erro
      this.metrics.increment('bureau_credito.popia.validation_errors', {
        tenant_id: request.tenantId,
        error_type: error.name || 'unknown'
      });
      
      throw error;
    } finally {
      span.end();
    }
  }
  
  /**
   * Verifica se o POPIA é aplicável para o processamento
   * 
   * @param request Requisição de validação POPIA
   * @returns Verdadeiro se o POPIA for aplicável
   */
  private isPOPIAApplicable(request: POPIAValidationRequest): boolean {
    // POPIA é aplicável se:
    // 1. O usuário é residente sul-africano (se a informação estiver disponível)
    // 2. OU se o processamento ocorre na África do Sul
    // 3. OU se o serviço é oferecido para o mercado sul-africano
    
    // Para fins de demonstração, vamos considerar aplicável se explicitamente marcado como residente sul-africano
    // ou se não temos essa informação (assumimos que sim para segurança)
    return request.isSouthAfricanResident !== false;
  }
  
  /**
   * Executa uma validação específica POPIA
   * 
   * @param validationType Tipo de validação a ser executada
   * @param request Requisição de validação POPIA
   * @returns Resultado da validação específica
   */
  private async executeValidation(
    validationType: POPIAValidationType,
    request: POPIAValidationRequest
  ): Promise<POPIAValidationResult> {
    const span = this.tracer.startSpan('popia.execute_validation', { validationType });
    
    try {
      // Escolher a função de validação apropriada com base no tipo
      let validationResult: POPIAValidationResult;
      
      switch (validationType) {
        case POPIAValidationType.LAWFULNESS_CHECK:
          validationResult = await this.validateLawfulness(request);
          break;
        case POPIAValidationType.MINIMALITY:
          validationResult = await this.validateMinimality(request);
          break;
        case POPIAValidationType.PURPOSE_SPECIFICATION:
          validationResult = await this.validatePurposeSpecification(request);
          break;
        case POPIAValidationType.SECURITY_SAFEGUARDS:
          validationResult = await this.validateSecuritySafeguards(request);
          break;
        case POPIAValidationType.CROSS_BORDER_TRANSFER:
          validationResult = await this.validateCrossBorderTransfer(request);
          break;
        case POPIAValidationType.SPECIAL_PERSONAL_INFO:
          validationResult = await this.validateSpecialPersonalInfo(request);
          break;
        case POPIAValidationType.AUTOMATED_DECISION_MAKING:
          validationResult = await this.validateAutomatedDecisionMaking(request);
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
      this.metrics.increment('bureau_credito.popia.validation_type', {
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
   * Validação de legalidade do processamento
   * 
   * @param request Requisição de validação POPIA
   * @returns Resultado da validação
   */
  private async validateLawfulness(request: POPIAValidationRequest): Promise<POPIAValidationResult> {
    // Verificar se uma base legal foi especificada
    if (!request.processingMetadata.lawfulBasis) {
      return {
        validationType: POPIAValidationType.LAWFULNESS_CHECK,
        passed: false,
        failureReason: 'Base legal não especificada',
        requiredActions: ['Definir uma base legal válida para o processamento de dados'],
        severity: 'critical'
      };
    }
    
    // Verificar se a base legal é válida para Bureau de Créditos
    // Para crédito, as bases válidas são: consentimento, interesse legítimo, obrigação legal e relatórios de crédito
    const validBases = [
      POPIALawfulBasis.CONSENT,
      POPIALawfulBasis.LEGITIMATE_INTEREST,
      POPIALawfulBasis.LEGAL_OBLIGATION,
      POPIALawfulBasis.CREDIT_REPORTING
    ];
    
    if (!validBases.includes(request.processingMetadata.lawfulBasis)) {
      return {
        validationType: POPIAValidationType.LAWFULNESS_CHECK,
        passed: false,
        failureReason: 'Base legal incompatível com operações de bureau de crédito',
        requiredActions: ['Revisar e ajustar a base legal para uma das bases válidas para operações de crédito'],
        severity: 'critical'
      };
    }
    
    // Se a base legal é consentimento, verificar se o consentimento existe
    if (request.processingMetadata.lawfulBasis === POPIALawfulBasis.CONSENT) {
      if (!request.consentReference) {
        return {
          validationType: POPIAValidationType.LAWFULNESS_CHECK,
          passed: false,
          failureReason: 'Base legal é consentimento, mas referência de consentimento não fornecida',
          requiredActions: ['Obter e registrar consentimento explícito do usuário'],
          severity: 'critical'
        };
      }
      
      // Na implementação real, verificaríamos a validade do consentimento
      // aqui consultando um serviço de gerenciamento de consentimento
    }
    
    // Verificar se os participantes obrigatórios estão definidos
    if (!request.processingMetadata.responsibleParty || !request.processingMetadata.informationOfficer) {
      return {
        validationType: POPIAValidationType.LAWFULNESS_CHECK,
        passed: false,
        failureReason: 'Parte responsável ou oficial de informação não especificados',
        requiredActions: ['Designar e registrar a parte responsável e o oficial de informação'],
        severity: 'high'
      };
    }
    
    return {
      validationType: POPIAValidationType.LAWFULNESS_CHECK,
      passed: true,
      severity: 'critical'
    };
  }
  
  /**
   * Validação do princípio de minimalidade
   * 
   * @param request Requisição de validação POPIA
   * @returns Resultado da validação
   */
  private async validateMinimality(request: POPIAValidationRequest): Promise<POPIAValidationResult> {
    // Verificar se todos os campos de dados são realmente necessários para a finalidade declarada
    
    // Lista de campos necessários com base na finalidade (simulada)
    let requiredFields: string[] = [];
    
    // Determinar campos necessários com base na operação
    if (request.processingMetadata.operationType.includes('credit_assessment')) {
      requiredFields = ['id_number', 'full_name', 'income', 'address'];
    } else if (request.processingMetadata.operationType.includes('fraud_detection')) {
      requiredFields = ['id_number', 'device_id', 'ip_address', 'transaction_history'];
    } else {
      // Default para outras operações
      requiredFields = ['id_number', 'full_name'];
    }
    
    // Verificar se há campos desnecessários
    const unnecessaryFields = request.dataFields.filter(field => !requiredFields.includes(field));
    
    if (unnecessaryFields.length > 0) {
      return {
        validationType: POPIAValidationType.MINIMALITY,
        passed: false,
        failureReason: 'Campos desnecessários solicitados para a finalidade',
        requiredActions: [`Remover os campos desnecessários: ${unnecessaryFields.join(', ')}`],
        severity: 'high'
      };
    }
    
    return {
      validationType: POPIAValidationType.MINIMALITY,
      passed: true,
      severity: 'high'
    };
  }
  
  /**
   * Validação de especificação de finalidade
   * 
   * @param request Requisição de validação POPIA
   * @returns Resultado da validação
   */
  private async validatePurposeSpecification(request: POPIAValidationRequest): Promise<POPIAValidationResult> {
    // Verificar se a finalidade está claramente definida
    if (!request.processingMetadata.purpose || request.processingMetadata.purpose.trim().length === 0) {
      return {
        validationType: POPIAValidationType.PURPOSE_SPECIFICATION,
        passed: false,
        failureReason: 'Finalidade de processamento não especificada',
        requiredActions: ['Definir claramente a finalidade do processamento de dados'],
        severity: 'high'
      };
    }
    
    // Verificar se a finalidade é específica e explícita
    const isPurposeSpecific = request.processingMetadata.purpose.length > 10;
    
    if (!isPurposeSpecific) {
      return {
        validationType: POPIAValidationType.PURPOSE_SPECIFICATION,
        passed: false,
        failureReason: 'Finalidade de processamento não é específica o suficiente',
        requiredActions: ['Definir finalidade mais específica e explícita para o processamento'],
        severity: 'medium'
      };
    }
    
    // Verificar se há período de retenção definido
    if (!request.processingMetadata.retentionPeriod || request.processingMetadata.retentionPeriod <= 0) {
      return {
        validationType: POPIAValidationType.PURPOSE_SPECIFICATION,
        passed: false,
        failureReason: 'Período de retenção não definido',
        requiredActions: ['Definir um período de retenção claro e adequado para os dados'],
        severity: 'high'
      };
    }
    
    return {
      validationType: POPIAValidationType.PURPOSE_SPECIFICATION,
      passed: true,
      severity: 'high'
    };
  }
  
  /**
   * Validação de salvaguardas de segurança
   * 
   * @param request Requisição de validação POPIA
   * @returns Resultado da validação
   */
  private async validateSecuritySafeguards(request: POPIAValidationRequest): Promise<POPIAValidationResult> {
    // Verificar se há medidas de segurança definidas
    if (!request.processingMetadata.securityMeasures || request.processingMetadata.securityMeasures.length === 0) {
      return {
        validationType: POPIAValidationType.SECURITY_SAFEGUARDS,
        passed: false,
        failureReason: 'Medidas de segurança não especificadas',
        requiredActions: ['Definir e implementar medidas técnicas e organizacionais de segurança'],
        severity: 'critical'
      };
    }
    
    // Medidas de segurança mínimas esperadas conforme POPIA
    const requiredMeasures = ['encryption', 'access_control', 'logging', 'incident_response'];
    
    // Verificar se todas as medidas mínimas estão presentes
    const missingMeasures = requiredMeasures.filter(
      measure => !request.processingMetadata.securityMeasures.some(
        m => m.toLowerCase().includes(measure)
      )
    );
    
    if (missingMeasures.length > 0) {
      return {
        validationType: POPIAValidationType.SECURITY_SAFEGUARDS,
        passed: false,
        failureReason: 'Medidas de segurança insuficientes',
        requiredActions: [`Implementar medidas de segurança adicionais: ${missingMeasures.join(', ')}`],
        severity: 'high'
      };
    }
    
    return {
      validationType: POPIAValidationType.SECURITY_SAFEGUARDS,
      passed: true,
      severity: 'critical'
    };
  }
  
  /**
   * Validação de transferência transfronteiriça de dados
   * 
   * @param request Requisição de validação POPIA
   * @returns Resultado da validação
   */
  private async validateCrossBorderTransfer(request: POPIAValidationRequest): Promise<POPIAValidationResult> {
    // Verificar se há transferência transfronteiriça
    if (!request.processingMetadata.crossBorderTransfer) {
      return {
        validationType: POPIAValidationType.CROSS_BORDER_TRANSFER,
        passed: true,
        severity: 'high'
      };
    }
    
    // Verificar se os países de destino estão definidos
    if (!request.processingMetadata.destinationCountries || request.processingMetadata.destinationCountries.length === 0) {
      return {
        validationType: POPIAValidationType.CROSS_BORDER_TRANSFER,
        passed: false,
        failureReason: 'Transferência transfronteiriça sem países de destino especificados',
        requiredActions: ['Especificar os países de destino para transferência transfronteiriça'],
        severity: 'high'
      };
    }
    
    // Lista de países com proteção adequada conforme POPIA (simplificada para o exemplo)
    // Para uma implementação real, seria necessário uma lista atualizada de países aprovados
    const adequateCountries = [
      'ZA', // África do Sul
      'AT', 'BE', 'BG', 'HR', 'CY', 'CZ', 'DK', 'EE', 'FI', 'FR', 
      'DE', 'GR', 'HU', 'IE', 'IT', 'LV', 'LT', 'LU', 'MT', 'NL', 
      'PL', 'PT', 'RO', 'SK', 'SI', 'ES', 'SE', // UE
      'GB', 'CH', 'CA', 'AR', 'UY', 'NZ', 'IL', 'JP' // Outros com leis adequadas
    ];
    
    // Verificar se todos os países de destino são adequados
    const inadequateCountries = request.processingMetadata.destinationCountries.filter(
      country => !adequateCountries.includes(country)
    );
    
    if (inadequateCountries.length > 0) {
      return {
        validationType: POPIAValidationType.CROSS_BORDER_TRANSFER,
        passed: false,
        failureReason: 'Transferência transfronteiriça para países sem proteção adequada',
        requiredActions: [
          `Implementar salvaguardas adicionais para transferências para: ${inadequateCountries.join(', ')}`,
          'Considerar termos contratuais específicos ou consentimento explícito para transferências internacionais'
        ],
        severity: 'high'
      };
    }
    
    return {
      validationType: POPIAValidationType.CROSS_BORDER_TRANSFER,
      passed: true,
      severity: 'high'
    };
  }