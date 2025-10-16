import { ComplianceValidatorIntegrator, ComplianceValidationRequest } from './compliance-validator-integrator';
import { GDPRValidationRequest, GDPRProcessingPurpose } from './gdpr-validator';
import { LGPDValidationRequest, LGPDLegalBasis } from './lgpd-validator';
import { POPIAValidationRequest, POPIALawfulBasis } from './popia-validator';

/**
 * Métodos de mapeamento para o integrador de validadores de conformidade.
 * Este arquivo contém os métodos que mapeiam a requisição unificada para formatos específicos.
 */

/**
 * Mapeia a solicitação de validação unificada para o formato GDPR
 * 
 * @param request Solicitação de validação unificada
 * @returns Solicitação formatada para o validador GDPR
 */
export function mapToGdprRequest(this: ComplianceValidatorIntegrator, request: ComplianceValidationRequest): GDPRValidationRequest {
  // Mapear base legal para o formato GDPR
  let gdprLegalBasis;
  
  switch (request.processingLegalBasis?.toLowerCase()) {
    case 'consent':
      gdprLegalBasis = 'consent';
      break;
    case 'contract':
    case 'contract_performance':
      gdprLegalBasis = 'contract';
      break;
    case 'legal_obligation':
      gdprLegalBasis = 'legal_obligation';
      break;
    case 'vital_interest':
      gdprLegalBasis = 'vital_interest';
      break;
    case 'public_interest':
      gdprLegalBasis = 'public_task';
      break;
    case 'legitimate_interest':
      gdprLegalBasis = 'legitimate_interest';
      break;
    default:
      gdprLegalBasis = 'consent'; // Padrão mais seguro
  }
  
  // Mapear finalidade para o formato GDPR
  let gdprPurpose: GDPRProcessingPurpose;
  
  if (request.operationType.includes('credit_assessment')) {
    gdprPurpose = GDPRProcessingPurpose.CREDIT_SCORING;
  } else if (request.operationType.includes('fraud')) {
    gdprPurpose = GDPRProcessingPurpose.FRAUD_PREVENTION;
  } else if (request.operationType.includes('marketing')) {
    gdprPurpose = GDPRProcessingPurpose.MARKETING;
  } else if (request.operationType.includes('research')) {
    gdprPurpose = GDPRProcessingPurpose.RESEARCH;
  } else {
    gdprPurpose = GDPRProcessingPurpose.SERVICE_PROVISION;
  }
  
  // Mapear para o formato da requisição GDPR
  const gdprRequest: GDPRValidationRequest = {
    userId: request.userId,
    tenantId: request.tenantId,
    processingMetadata: {
      controller: request.tenantId,
      processor: 'bureau-credito-service',
      legalBasis: gdprLegalBasis,
      purpose: gdprPurpose,
      dataCategories: request.dataCategories,
      retentionPeriod: request.retentionPeriodDays,
      thirdPartySharing: request.crossBorderTransfer || false,
      thirdParties: request.destinationCountries || [],
      securityMeasures: request.securityMeasures || ['encryption', 'access_control'],
      specialCategories: request.specialCategories || false,
      automatedDecision: request.automatedDecisionMaking || false
    },
    dataSubjectEUResident: request.dataSubjectCountry?.toLowerCase() === 'eu',
    consentReference: request.consentReferences?.gdpr,
    dataFields: request.dataFields,
    validationTypes: undefined // Use os padrões
  };
  
  return gdprRequest;
}

/**
 * Mapeia a solicitação de validação unificada para o formato LGPD
 * 
 * @param request Solicitação de validação unificada
 * @returns Solicitação formatada para o validador LGPD
 */
export function mapToLgpdRequest(this: ComplianceValidatorIntegrator, request: ComplianceValidationRequest): LGPDValidationRequest {
  // Mapear base legal para o formato LGPD
  let lgpdLegalBasis: LGPDLegalBasis;
  
  switch (request.processingLegalBasis?.toLowerCase()) {
    case 'consent':
      lgpdLegalBasis = LGPDLegalBasis.CONSENT;
      break;
    case 'contract':
    case 'contract_performance':
      lgpdLegalBasis = LGPDLegalBasis.CONTRACT_EXECUTION;
      break;
    case 'legal_obligation':
      lgpdLegalBasis = LGPDLegalBasis.LEGAL_OBLIGATION;
      break;
    case 'vital_interest':
      lgpdLegalBasis = LGPDLegalBasis.LIFE_PROTECTION;
      break;
    case 'public_interest':
      lgpdLegalBasis = LGPDLegalBasis.PUBLIC_POLICY;
      break;
    case 'legitimate_interest':
      lgpdLegalBasis = LGPDLegalBasis.LEGITIMATE_INTEREST;
      break;
    case 'credit_reporting':
      lgpdLegalBasis = LGPDLegalBasis.CREDIT_PROTECTION;
      break;
    default:
      lgpdLegalBasis = LGPDLegalBasis.CONSENT; // Padrão mais seguro
  }
  
  // Mapear para o formato da requisição LGPD
  const lgpdRequest: LGPDValidationRequest = {
    userId: request.userId,
    tenantId: request.tenantId,
    processingMetadata: {
      controller: request.tenantId,
      processor: 'bureau-credito-service',
      legalBasis: lgpdLegalBasis,
      purpose: request.dataPurpose,
      dataCategories: request.dataCategories,
      retentionPeriod: request.retentionPeriodDays,
      internationalTransfer: request.crossBorderTransfer || false,
      destinationCountries: request.destinationCountries || [],
      securityMeasures: request.securityMeasures || ['encryption', 'access_control'],
      sensitiveData: request.specialCategories || false,
      automatedDecision: request.automatedDecisionMaking || false
    },
    isBrazilianResident: request.dataSubjectCountry?.toLowerCase() === 'brazil',
    consentReference: request.consentReferences?.lgpd,
    dataFields: request.dataFields,
    validationTypes: undefined // Use os padrões
  };
  
  return lgpdRequest;
}

/**
 * Mapeia a solicitação de validação unificada para o formato POPIA
 * 
 * @param request Solicitação de validação unificada
 * @returns Solicitação formatada para o validador POPIA
 */
export function mapToPopiaRequest(this: ComplianceValidatorIntegrator, request: ComplianceValidationRequest): POPIAValidationRequest {
  // Mapear base legal para o formato POPIA
  let popiaLawfulBasis: POPIALawfulBasis;
  
  switch (request.processingLegalBasis?.toLowerCase()) {
    case 'consent':
      popiaLawfulBasis = POPIALawfulBasis.CONSENT;
      break;
    case 'contract':
    case 'contract_performance':
      popiaLawfulBasis = POPIALawfulBasis.CONTRACT_PERFORMANCE;
      break;
    case 'legal_obligation':
      popiaLawfulBasis = POPIALawfulBasis.LEGAL_OBLIGATION;
      break;
    case 'vital_interest':
      popiaLawfulBasis = POPIALawfulBasis.VITAL_INTEREST;
      break;
    case 'public_interest':
      popiaLawfulBasis = POPIALawfulBasis.PUBLIC_INTEREST;
      break;
    case 'legitimate_interest':
      popiaLawfulBasis = POPIALawfulBasis.LEGITIMATE_INTEREST;
      break;
    case 'credit_reporting':
      popiaLawfulBasis = POPIALawfulBasis.CREDIT_REPORTING;
      break;
    default:
      popiaLawfulBasis = POPIALawfulBasis.CONSENT; // Padrão mais seguro
  }
  
  // Mapear para o formato da requisição POPIA
  const popiaRequest: POPIAValidationRequest = {
    userId: request.userId,
    tenantId: request.tenantId,
    processingMetadata: {
      operationId: request.operationId,
      operationType: request.operationType,
      responsibleParty: request.tenantId,
      operator: 'bureau-credito-service',
      informationOfficer: `info-officer-${request.tenantId}`, // Placeholder - normalmente viria de uma configuração
      lawfulBasis: popiaLawfulBasis,
      purpose: request.dataPurpose,
      dataCategories: request.dataCategories,
      retentionPeriod: request.retentionPeriodDays,
      thirdPartySharing: request.crossBorderTransfer || false,
      crossBorderTransfer: request.crossBorderTransfer || false,
      destinationCountries: request.destinationCountries,
      securityMeasures: request.securityMeasures || ['encryption', 'access_control'],
      automatedDecisionMaking: request.automatedDecisionMaking || false
    },
    isSouthAfricanResident: request.dataSubjectCountry?.toLowerCase() === 'south_africa',
    consentReference: request.consentReferences?.popia,
    dataFields: request.dataFields,
    validationTypes: undefined // Use os padrões
  };
  
  return popiaRequest;
}

/**
 * Cria e formata um relatório resumido de conformidade para uso em dashboards
 * 
 * @param result Resultado consolidado das validações
 * @returns Objeto com dados formatados para dashboard
 */
export function createComplianceSummary(result: any): any {
  // Calcula estatísticas por regulamentação
  const regulationStats = {};
  let totalChecks = 0;
  let passedChecks = 0;
  
  if (result.resultsPerRegulation.gdpr) {
    const gdprResults = result.resultsPerRegulation.gdpr.validationResults || [];
    regulationStats['gdpr'] = {
      total: gdprResults.length,
      passed: gdprResults.filter(r => r.passed).length,
      critical: gdprResults.filter(r => !r.passed && r.severity === 'critical').length,
      high: gdprResults.filter(r => !r.passed && r.severity === 'high').length,
      medium: gdprResults.filter(r => !r.passed && r.severity === 'medium').length,
      low: gdprResults.filter(r => !r.passed && r.severity === 'low').length,
    };
    totalChecks += gdprResults.length;
    passedChecks += gdprResults.filter(r => r.passed).length;
  }
  
  if (result.resultsPerRegulation.lgpd) {
    const lgpdResults = result.resultsPerRegulation.lgpd.validationResults || [];
    regulationStats['lgpd'] = {
      total: lgpdResults.length,
      passed: lgpdResults.filter(r => r.passed).length,
      critical: lgpdResults.filter(r => !r.passed && r.severity === 'critical').length,
      high: lgpdResults.filter(r => !r.passed && r.severity === 'high').length,
      medium: lgpdResults.filter(r => !r.passed && r.severity === 'medium').length,
      low: lgpdResults.filter(r => !r.passed && r.severity === 'low').length,
    };
    totalChecks += lgpdResults.length;
    passedChecks += lgpdResults.filter(r => r.passed).length;
  }
  
  if (result.resultsPerRegulation.popia) {
    const popiaResults = result.resultsPerRegulation.popia.validationResults || [];
    regulationStats['popia'] = {
      total: popiaResults.length,
      passed: popiaResults.filter(r => r.passed).length,
      critical: popiaResults.filter(r => !r.passed && r.severity === 'critical').length,
      high: popiaResults.filter(r => !r.passed && r.severity === 'high').length,
      medium: popiaResults.filter(r => !r.passed && r.severity === 'medium').length,
      low: popiaResults.filter(r => !r.passed && r.severity === 'low').length,
    };
    totalChecks += popiaResults.length;
    passedChecks += popiaResults.filter(r => r.passed).length;
  }
  
  // Cria resumo para dashboard
  return {
    timestamp: result.timestamp,
    userId: result.userId,
    tenantId: result.tenantId,
    operationId: result.operationId,
    complianceScore: totalChecks > 0 ? Math.round((passedChecks / totalChecks) * 100) : 100,
    processingAllowed: result.processingAllowed,
    regulationStats,
    applicableRegulations: result.applicableRegulations,
    actionItemsCount: result.requiredActions.length,
    restrictionsCount: result.processingRestrictions?.length || 0
  };
}