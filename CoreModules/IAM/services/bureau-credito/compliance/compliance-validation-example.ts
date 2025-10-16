/**
 * Exemplo de uso do Integrador de Validadores de Conformidade
 * 
 * Este arquivo demonstra como utilizar o ComplianceValidatorIntegrator para realizar
 * validações de conformidade com múltiplas regulamentações (GDPR, LGPD, POPIA)
 * em operações do Bureau de Créditos.
 * 
 * @module ComplianceValidationExample
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

// Configuração das ferramentas de observabilidade
const logger = new Logger({ service: 'bureau-credito-compliance' });
const metrics = new Metrics({ service: 'bureau-credito-compliance' });
const tracer = new Tracer({ service: 'bureau-credito-compliance' });

/**
 * Realiza uma validação de conformidade para uma operação de avaliação de crédito
 * 
 * @param userId ID do usuário
 * @param tenantId ID do tenant
 * @param dataSubjectCountry País de residência do titular dos dados
 * @param consentId Referência ao consentimento (se existir)
 * @returns Resultado da validação de conformidade
 */
export async function validateCreditAssessmentCompliance(
  userId: string,
  tenantId: string,
  dataSubjectCountry: Region | string,
  consentId?: string
): Promise<ConsolidatedComplianceResult> {
  // Inicializar o integrador de validadores
  const complianceIntegrator = new ComplianceValidatorIntegrator(logger, metrics, tracer);
  
  // Criar a solicitação de validação
  const validationRequest: ComplianceValidationRequest = {
    userId,
    tenantId,
    operationId: `credit-assessment-${Date.now()}`,
    operationType: 'credit_assessment',
    dataSubjectCountry,
    dataProcessingCountry: 'angola', // País onde o processamento ocorre
    businessTargetCountries: ['angola', 'brazil', 'south_africa', 'eu', 'mozambique'],
    consentReferences: consentId ? {
      gdpr: consentId,
      lgpd: consentId,
      popia: consentId
    } : undefined,
    dataPurpose: 'Avaliação de crédito para análise de risco financeiro e elegibilidade para empréstimos',
    dataCategories: ['personal_details', 'financial_information', 'credit_history'],
    dataFields: [
      'id_number', 'full_name', 'address', 'phone_number', 
      'monthly_income', 'employment_status', 'credit_score', 
      'previous_loans', 'payment_history'
    ],
    retentionPeriodDays: 730, // 2 anos
    processingLegalBasis: 'legitimate_interest',
    specialCategories: false,
    automatedDecisionMaking: true,
    crossBorderTransfer: false,
    securityMeasures: [
      'encryption', 'access_control', 'logging', 'data_minimization',
      'human_review', 'appeal_process', 'incident_response'
    ]
  };
  
  // Executar validação de conformidade
  const span = tracer.startSpan('credit_assessment.validate_compliance');
  
  try {
    logger.info({
      message: 'Iniciando validação de conformidade para avaliação de crédito',
      userId,
      tenantId,
      dataSubjectCountry
    });
    
    // Executar validação
    const complianceResult = await complianceIntegrator.validate(validationRequest);
    
    // Registrar resultado
    logger.info({
      message: `Validação de conformidade concluída: ${complianceResult.overallCompliant ? 'Conforme' : 'Não conforme'}`,
      userId,
      tenantId,
      overallCompliant: complianceResult.overallCompliant,
      processingAllowed: complianceResult.processingAllowed,
      applicableRegulations: complianceResult.applicableRegulations
    });
    
    // Registrar métricas de resultado
    metrics.increment('bureau_credito.credit_assessment.compliance_check', {
      tenant_id: tenantId,
      compliant: complianceResult.overallCompliant.toString(),
      processing_allowed: complianceResult.processingAllowed.toString()
    });
    
    return complianceResult;
  } catch (error) {
    // Registrar erro
    logger.error({
      message: 'Erro durante validação de conformidade para avaliação de crédito',
      error: error.message,
      stack: error.stack,
      userId,
      tenantId
    });
    
    // Registrar métrica de erro
    metrics.increment('bureau_credito.credit_assessment.compliance_errors', {
      tenant_id: tenantId,
      error_type: error.name || 'unknown'
    });
    
    throw error;
  } finally {
    span.end();
  }
}

/**
 * Processa o resultado da validação de conformidade e determina se a operação pode prosseguir
 * 
 * @param complianceResult Resultado da validação de conformidade
 * @returns Objeto com informações sobre permissão e restrições
 */
export function processComplianceResult(
  complianceResult: ConsolidatedComplianceResult
): { canProceed: boolean; restrictions: string[]; requiresHumanReview: boolean } {
  // Verificar se o processamento é permitido
  const canProceed = complianceResult.processingAllowed;
  
  // Obter restrições de processamento
  const restrictions = complianceResult.processingRestrictions || [];
  
  // Determinar se revisão humana é necessária (se há não-conformidades críticas)
  const requiresHumanReview = !complianceResult.overallCompliant;
  
  return {
    canProceed,
    restrictions,
    requiresHumanReview
  };
}

/**
 * Exemplo de uso do integrador de conformidade em um fluxo completo de avaliação de crédito
 */
async function exampleCreditAssessmentFlow(): Promise<void> {
  try {
    console.log('Iniciando fluxo de avaliação de crédito com validação de conformidade');
    
    // Dados de exemplo
    const userId = 'user-12345';
    const tenantId = 'financial-institution-xyz';
    const dataSubjectCountry = Region.BRAZIL; // Cliente brasileiro
    const consentId = 'consent-abc-123'; // ID de um consentimento previamente obtido
    
    console.log('Executando validação de conformidade...');
    
    // Executar validação de conformidade
    const complianceResult = await validateCreditAssessmentCompliance(
      userId,
      tenantId,
      dataSubjectCountry,
      consentId
    );
    
    // Processar resultado
    const { canProceed, restrictions, requiresHumanReview } = processComplianceResult(complianceResult);
    
    console.log(`Resultado da validação de conformidade:`);
    console.log(`- Pode prosseguir: ${canProceed}`);
    console.log(`- Requer revisão humana: ${requiresHumanReview}`);
    
    if (restrictions.length > 0) {
      console.log('Restrições aplicáveis:');
      restrictions.forEach((restriction, index) => {
        console.log(`  ${index + 1}. ${restriction}`);
      });
    }
    
    // Simulação da decisão baseada na conformidade
    if (canProceed) {
      console.log('Avaliação de crédito autorizada a prosseguir com as restrições aplicáveis');
      
      if (requiresHumanReview) {
        console.log('ATENÇÃO: Revisão humana necessária antes da decisão final');
        
        // Aqui executaríamos a lógica para encaminhar para revisão humana
        // ...
      } else {
        console.log('Processamento totalmente automatizado permitido');
        
        // Aqui executaríamos a lógica completa de avaliação de crédito
        // ...
      }
    } else {
      console.log('Avaliação de crédito bloqueada devido a não-conformidades críticas');
      console.log('Ações requeridas:');
      
      complianceResult.requiredActions.forEach((action, index) => {
        console.log(`  ${index + 1}. ${action}`);
      });
    }
    
    console.log('Fluxo de avaliação de crédito concluído');
  } catch (error) {
    console.error('Erro durante o fluxo de avaliação de crédito:', error);
  }
}

// Executar exemplo se chamado diretamente (não importado como módulo)
if (require.main === module) {
  exampleCreditAssessmentFlow()
    .then(() => console.log('Exemplo concluído com sucesso'))
    .catch(err => console.error('Erro no exemplo:', err));
}