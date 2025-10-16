/**
 * Integrador de Validadores de Conformidade
 * 
 * Este módulo unifica a interface para os diferentes validadores de conformidade
 * regulatória (GDPR, LGPD, POPIA) e implementa lógica para determinar quais
 * validadores devem ser aplicados com base no contexto da solicitação.
 * 
 * @module ComplianceValidatorIntegrator
 */

import { Logger } from '../../../observability/logging/hook_logger';
import { Metrics } from '../../../observability/metrics/hook_metrics';
import { Tracer } from '../../../observability/tracing/hook_tracing';

import { GDPRValidator, GDPRValidationRequest, GDPRCompleteValidationResult } from './gdpr-validator';
import { LGPDValidator, LGPDValidationRequest, LGPDCompleteValidationResult } from './lgpd-validator';
import { POPIAValidator, POPIAValidationRequest, POPIACompleteValidationResult } from './popia-validator';

// Define o país ou região para determinar as regulações aplicáveis
export enum Region {
  EU = 'eu',
  BRAZIL = 'brazil',
  SOUTH_AFRICA = 'south_africa',
  ANGOLA = 'angola',
  MOZAMBIQUE = 'mozambique',
  USA = 'usa',
  CHINA = 'china',
  OTHER = 'other'
}

// Interface para o resultado consolidado de múltiplas validações
export interface ConsolidatedComplianceResult {
  requestId: string;
  timestamp: Date;
  userId: string;
  tenantId: string;
  operationId: string;
  overallCompliant: boolean;
  resultsPerRegulation: {
    gdpr?: GDPRCompleteValidationResult;
    lgpd?: LGPDCompleteValidationResult;
    popia?: POPIACompleteValidationResult;
  };
  applicableRegulations: string[];
  requiredActions: string[];
  processingAllowed: boolean;
  processingRestrictions?: string[];
  auditRecord: {
    validatedBy: string;
    validationTimestamp: Date;
    version: string;
  };
}

// Interface para solicitação de validação unificada
export interface ComplianceValidationRequest {
  userId: string;
  tenantId: string;
  operationId: string;
  operationType: string;
  dataSubjectCountry?: Region | string;
  dataProcessingCountry?: Region | string;
  businessTargetCountries?: (Region | string)[];
  consentReferences?: {
    gdpr?: string;
    lgpd?: string;
    popia?: string;
  };
  dataPurpose: string;
  dataCategories: string[];
  dataFields: string[];
  retentionPeriodDays: number;
  processingLegalBasis?: string;
  specialCategories?: boolean;
  automatedDecisionMaking?: boolean;
  crossBorderTransfer?: boolean;
  destinationCountries?: string[];
  securityMeasures?: string[];
  skipRegulations?: string[];
}

/**
 * Classe que integra os diferentes validadores de conformidade
 */
export class ComplianceValidatorIntegrator {
  private logger: Logger;
  private metrics: Metrics;
  private tracer: Tracer;
  
  private gdprValidator: GDPRValidator;
  private lgpdValidator: LGPDValidator;
  private popiaValidator: POPIAValidator;
  
  /**
   * Construtor para o integrador de validadores de conformidade
   */
  constructor(logger: Logger, metrics: Metrics, tracer: Tracer) {
    this.logger = logger;
    this.metrics = metrics;
    this.tracer = tracer;
    
    // Inicializa os validadores individuais
    this.gdprValidator = new GDPRValidator(logger, metrics, tracer);
    this.lgpdValidator = new LGPDValidator(logger, metrics, tracer);
    this.popiaValidator = new POPIAValidator(logger, metrics, tracer);
  }
  
  /**
   * Determina quais regulamentos são aplicáveis com base no contexto da solicitação
   * 
   * @param request Solicitação de validação de conformidade
   * @returns Lista de regulamentos aplicáveis
   */
  private determineApplicableRegulations(request: ComplianceValidationRequest): string[] {
    const applicableRegulations: string[] = [];
    
    // Verificar regulamentos explicitamente ignorados
    const skipRegulations = request.skipRegulations || [];
    
    // Verificar aplicabilidade do GDPR
    // GDPR se aplica se: o sujeito dos dados estiver na UE, o processamento ocorrer na UE,
    // ou se o serviço for direcionado ao mercado da UE
    const gdprCountries = ['eu', 'uk', 'norway', 'iceland', 'liechtenstein', 'switzerland'];
    const isGdprApplicable = 
      (request.dataSubjectCountry && gdprCountries.includes(request.dataSubjectCountry.toLowerCase())) ||
      (request.dataProcessingCountry && gdprCountries.includes(request.dataProcessingCountry.toLowerCase())) ||
      (request.businessTargetCountries && 
       request.businessTargetCountries.some(country => gdprCountries.includes(country.toLowerCase())));
    
    if (isGdprApplicable && !skipRegulations.includes('gdpr')) {
      applicableRegulations.push('gdpr');
    }
    
    // Verificar aplicabilidade do LGPD
    // LGPD se aplica se: o sujeito dos dados estiver no Brasil, o processamento ocorrer no Brasil,
    // ou se o serviço for direcionado ao mercado brasileiro
    const isLgpdApplicable = 
      (request.dataSubjectCountry && request.dataSubjectCountry.toLowerCase() === 'brazil') ||
      (request.dataProcessingCountry && request.dataProcessingCountry.toLowerCase() === 'brazil') ||
      (request.businessTargetCountries && 
       request.businessTargetCountries.some(country => country.toLowerCase() === 'brazil'));
    
    if (isLgpdApplicable && !skipRegulations.includes('lgpd')) {
      applicableRegulations.push('lgpd');
    }
    
    // Verificar aplicabilidade do POPIA
    // POPIA se aplica se: o sujeito dos dados estiver na África do Sul, o processamento ocorrer na África do Sul,
    // ou se o serviço for direcionado ao mercado sul-africano
    const isPOPIAApplicable = 
      (request.dataSubjectCountry && request.dataSubjectCountry.toLowerCase() === 'south_africa') ||
      (request.dataProcessingCountry && request.dataProcessingCountry.toLowerCase() === 'south_africa') ||
      (request.businessTargetCountries && 
       request.businessTargetCountries.some(country => country.toLowerCase() === 'south_africa'));
    
    if (isPOPIAApplicable && !skipRegulations.includes('popia')) {
      applicableRegulations.push('popia');
    }
    
    return applicableRegulations;
  }
  
  /**
   * Executa validações de conformidade com base nas regulamentações aplicáveis
   * 
   * @param request Solicitação de validação de conformidade
   * @returns Resultado consolidado das validações
   */
  public async validate(request: ComplianceValidationRequest): Promise<ConsolidatedComplianceResult> {
    const span = this.tracer.startSpan('compliance_integrator.validate');
    
    try {
      // Registrar início da validação
      this.logger.info({
        message: 'Iniciando validação integrada de conformidade',
        userId: request.userId,
        tenantId: request.tenantId,
        operationId: request.operationId,
        operationType: request.operationType
      });
      
      // Timestamp para métricas
      const startTime = Date.now();
      
      // Determinar regulamentações aplicáveis
      const applicableRegulations = this.determineApplicableRegulations(request);
      
      this.logger.info({
        message: `Regulamentações aplicáveis: ${applicableRegulations.join(', ')}`,
        userId: request.userId,
        tenantId: request.tenantId,
        operationId: request.operationId
      });
      
      // Se nenhuma regulamentação for aplicável, retornar resultado positivo
      if (applicableRegulations.length === 0) {
        const result: ConsolidatedComplianceResult = {
          requestId: `compliance-val-${Date.now()}`,
          timestamp: new Date(),
          userId: request.userId,
          tenantId: request.tenantId,
          operationId: request.operationId,
          overallCompliant: true,
          resultsPerRegulation: {},
          applicableRegulations: [],
          requiredActions: [],
          processingAllowed: true,
          auditRecord: {
            validatedBy: 'compliance-validator-integrator',
            validationTimestamp: new Date(),
            version: '1.0.0'
          }
        };
        
        return result;
      }
      
      // Resultado por regulamentação
      const resultsPerRegulation: {
        gdpr?: GDPRCompleteValidationResult;
        lgpd?: LGPDCompleteValidationResult;
        popia?: POPIACompleteValidationResult;
      } = {};
      
      // Executar validações em paralelo
      const validationPromises: Promise<any>[] = [];
      
      // GDPR
      if (applicableRegulations.includes('gdpr')) {
        const gdprRequest = this.mapToGdprRequest(request);
        
        const gdprPromise = this.gdprValidator.validate(gdprRequest)
          .then(result => {
            resultsPerRegulation.gdpr = result;
          })
          .catch(error => {
            this.logger.error({
              message: 'Erro durante validação GDPR',
              error: error.message,
              userId: request.userId,
              tenantId: request.tenantId
            });
            
            // Criar um resultado de erro para GDPR
            resultsPerRegulation.gdpr = {
              requestId: `gdpr-error-${Date.now()}`,
              timestamp: new Date(),
              userId: request.userId,
              tenantId: request.tenantId,
              overallCompliant: false,
              validationResults: [{
                validationType: 'error' as any,
                passed: false,
                failureReason: `Erro de validação: ${error.message}`,
                severity: 'critical'
              }],
              requiredActions: ['Entrar em contato com o suporte técnico'],
              processingAllowed: false,
              auditRecord: {
                validatedBy: 'gdpr-validator-error',
                validationTimestamp: new Date(),
                version: '1.0.0'
              }
            };
          });
        
        validationPromises.push(gdprPromise);
      }
      
      // LGPD
      if (applicableRegulations.includes('lgpd')) {
        const lgpdRequest = this.mapToLgpdRequest(request);
        
        const lgpdPromise = this.lgpdValidator.validate(lgpdRequest)
          .then(result => {
            resultsPerRegulation.lgpd = result;
          })
          .catch(error => {
            this.logger.error({
              message: 'Erro durante validação LGPD',
              error: error.message,
              userId: request.userId,
              tenantId: request.tenantId
            });
            
            // Criar um resultado de erro para LGPD
            resultsPerRegulation.lgpd = {
              requestId: `lgpd-error-${Date.now()}`,
              timestamp: new Date(),
              userId: request.userId,
              tenantId: request.tenantId,
              overallCompliant: false,
              validationResults: [{
                validationType: 'error' as any,
                passed: false,
                failureReason: `Erro de validação: ${error.message}`,
                severity: 'critical'
              }],
              requiredActions: ['Entrar em contato com o suporte técnico'],
              processingAllowed: false,
              auditRecord: {
                validatedBy: 'lgpd-validator-error',
                validationTimestamp: new Date(),
                version: '1.0.0'
              }
            };
          });
        
        validationPromises.push(lgpdPromise);
      }
      
      // POPIA
      if (applicableRegulations.includes('popia')) {
        const popiaRequest = this.mapToPopiaRequest(request);
        
        const popiaPromise = this.popiaValidator.validate(popiaRequest)
          .then(result => {
            resultsPerRegulation.popia = result;
          })
          .catch(error => {
            this.logger.error({
              message: 'Erro durante validação POPIA',
              error: error.message,
              userId: request.userId,
              tenantId: request.tenantId
            });
            
            // Criar um resultado de erro para POPIA
            resultsPerRegulation.popia = {
              requestId: `popia-error-${Date.now()}`,
              timestamp: new Date(),
              userId: request.userId,
              tenantId: request.tenantId,
              overallCompliant: false,
              validationResults: [{
                validationType: 'error' as any,
                passed: false,
                failureReason: `Erro de validação: ${error.message}`,
                severity: 'critical'
              }],
              requiredActions: ['Entrar em contato com o suporte técnico'],
              processingAllowed: false,
              auditRecord: {
                validatedBy: 'popia-validator-error',
                validationTimestamp: new Date(),
                version: '1.0.0'
              }
            };
          });
        
        validationPromises.push(popiaPromise);
      }
      
      // Aguardar todas as validações
      await Promise.all(validationPromises);
      
      // Consolidar resultados
      let overallCompliant = true;
      let processingAllowed = true;
      const allRequiredActions: string[] = [];
      const allProcessingRestrictions: string[] = [];
      
      // Verificar conformidade GDPR
      if (resultsPerRegulation.gdpr) {
        overallCompliant = overallCompliant && resultsPerRegulation.gdpr.overallCompliant;
        processingAllowed = processingAllowed && resultsPerRegulation.gdpr.processingAllowed;
        
        if (resultsPerRegulation.gdpr.requiredActions?.length > 0) {
          allRequiredActions.push(...resultsPerRegulation.gdpr.requiredActions.map(action => `[GDPR] ${action}`));
        }
        
        if (resultsPerRegulation.gdpr.processingRestrictions?.length > 0) {
          allProcessingRestrictions.push(...resultsPerRegulation.gdpr.processingRestrictions.map(restriction => `[GDPR] ${restriction}`));
        }
      }
      
      // Verificar conformidade LGPD
      if (resultsPerRegulation.lgpd) {
        overallCompliant = overallCompliant && resultsPerRegulation.lgpd.overallCompliant;
        processingAllowed = processingAllowed && resultsPerRegulation.lgpd.processingAllowed;
        
        if (resultsPerRegulation.lgpd.requiredActions?.length > 0) {
          allRequiredActions.push(...resultsPerRegulation.lgpd.requiredActions.map(action => `[LGPD] ${action}`));
        }
        
        if (resultsPerRegulation.lgpd.processingRestrictions?.length > 0) {
          allProcessingRestrictions.push(...resultsPerRegulation.lgpd.processingRestrictions.map(restriction => `[LGPD] ${restriction}`));
        }
      }
      
      // Verificar conformidade POPIA
      if (resultsPerRegulation.popia) {
        overallCompliant = overallCompliant && resultsPerRegulation.popia.overallCompliant;
        processingAllowed = processingAllowed && resultsPerRegulation.popia.processingAllowed;
        
        if (resultsPerRegulation.popia.requiredActions?.length > 0) {
          allRequiredActions.push(...resultsPerRegulation.popia.requiredActions.map(action => `[POPIA] ${action}`));
        }
        
        if (resultsPerRegulation.popia.processingRestrictions?.length > 0) {
          allProcessingRestrictions.push(...resultsPerRegulation.popia.processingRestrictions.map(restriction => `[POPIA] ${restriction}`));
        }
      }
      
      // Construir resultado consolidado
      const result: ConsolidatedComplianceResult = {
        requestId: `compliance-val-${Date.now()}`,
        timestamp: new Date(),
        userId: request.userId,
        tenantId: request.tenantId,
        operationId: request.operationId,
        overallCompliant,
        resultsPerRegulation,
        applicableRegulations,
        requiredActions: [...new Set(allRequiredActions)], // Remove duplicatas
        processingAllowed,
        processingRestrictions: allProcessingRestrictions.length > 0 ? [...new Set(allProcessingRestrictions)] : undefined,
        auditRecord: {
          validatedBy: 'compliance-validator-integrator',
          validationTimestamp: new Date(),
          version: '1.0.0'
        }
      };
      
      // Registrar métricas
      const validationTime = Date.now() - startTime;
      
      this.metrics.histogram('bureau_credito.compliance.validation_time', validationTime, {
        tenant_id: request.tenantId,
        compliant: overallCompliant.toString()
      });
      
      this.metrics.increment('bureau_credito.compliance.validations_performed', {
        tenant_id: request.tenantId,
        compliant: overallCompliant.toString(),
        processing_allowed: processingAllowed.toString(),
        regulations_count: applicableRegulations.length.toString()
      });
      
      // Registrar resultado
      this.logger.info({
        message: `Validação de conformidade concluída: ${overallCompliant ? 'Conforme' : 'Não conforme'}`,
        userId: request.userId,
        tenantId: request.tenantId,
        operationId: request.operationId,
        operationType: request.operationType,
        applicableRegulations,
        overallCompliant,
        processingAllowed,
        validationTime
      });
      
      return result;
    } catch (error) {
      // Registrar erro
      this.logger.error({
        message: 'Erro durante validação integrada de conformidade',
        error: error.message,
        stack: error.stack,
        userId: request.userId,
        tenantId: request.tenantId,
        operationId: request.operationId
      });
      
      // Registrar métrica de erro
      this.metrics.increment('bureau_credito.compliance.validation_errors', {
        tenant_id: request.tenantId,
        error_type: error.name || 'unknown'
      });
      
      throw error;
    } finally {
      span.end();
    }
  }