import { POPIAValidator, POPIAValidationType, POPIAValidationRequest, POPIAValidationResult } from './popia-validator';

/**
 * Extensão dos métodos de validação para o validador POPIA.
 * Este arquivo será importado pelo validador principal.
 */

/**
 * Validação de informações pessoais especiais
 * 
 * @param request Requisição de validação POPIA
 * @returns Resultado da validação
 */
export async function validateSpecialPersonalInfo(this: POPIAValidator, request: POPIAValidationRequest): Promise<POPIAValidationResult> {
  const specialDataFields = ['race', 'ethnic_origin', 'health', 'biometric', 'criminal_record', 'religion', 'political_views', 'trade_union', 'sexual_orientation'];
  
  // Verificar se há campos especiais de dados
  const specialFieldsUsed = request.dataFields.filter(field => specialDataFields.some(special => field.includes(special)));
  
  if (specialFieldsUsed.length === 0) {
    // Não há dados especiais, retornar sucesso
    return {
      validationType: POPIAValidationType.SPECIAL_PERSONAL_INFO,
      passed: true,
      severity: 'critical'
    };
  }
  
  // Verificar a base legal para processamento de dados especiais
  const hasValidBasis = request.processingMetadata.lawfulBasis === 'explicit_consent_special' || 
                        request.processingMetadata.lawfulBasis === 'legal_obligation';
  
  if (!hasValidBasis) {
    return {
      validationType: POPIAValidationType.SPECIAL_PERSONAL_INFO,
      passed: false,
      failureReason: 'Base legal inadequada para processamento de informações pessoais especiais',
      requiredActions: ['Obter consentimento explícito para processamento de dados sensíveis', 
                      'Ou verificar obrigação legal que permita o processamento'],
      severity: 'critical'
    };
  }
  
  // Verificar se há consentimento para dados especiais quando necessário
  if (request.processingMetadata.lawfulBasis === 'explicit_consent_special' && !request.consentReference) {
    return {
      validationType: POPIAValidationType.SPECIAL_PERSONAL_INFO,
      passed: false,
      failureReason: 'Consentimento explícito necessário para dados especiais não fornecido',
      requiredActions: ['Obter consentimento explícito para processamento de dados especiais'],
      severity: 'critical'
    };
  }
  
  // Verificar se as salvaguardas para dados especiais estão implementadas
  const requiredSafeguards = ['encryption', 'access_control', 'special_data_policy'];
  
  const missingSafeguards = requiredSafeguards.filter(
    safeguard => !request.processingMetadata.securityMeasures.some(
      measure => measure.toLowerCase().includes(safeguard)
    )
  );
  
  if (missingSafeguards.length > 0) {
    return {
      validationType: POPIAValidationType.SPECIAL_PERSONAL_INFO,
      passed: false,
      failureReason: 'Salvaguardas insuficientes para dados pessoais especiais',
      requiredActions: [`Implementar salvaguardas adicionais: ${missingSafeguards.join(', ')}`],
      severity: 'high'
    };
  }
  
  return {
    validationType: POPIAValidationType.SPECIAL_PERSONAL_INFO,
    passed: true,
    severity: 'critical'
  };
}

/**
 * Validação de tomada de decisão automatizada
 * 
 * @param request Requisição de validação POPIA
 * @returns Resultado da validação
 */
export async function validateAutomatedDecisionMaking(this: POPIAValidator, request: POPIAValidationRequest): Promise<POPIAValidationResult> {
  // Verificar se o processamento envolve tomada de decisão automatizada
  if (!request.processingMetadata.automatedDecisionMaking) {
    return {
      validationType: POPIAValidationType.AUTOMATED_DECISION_MAKING,
      passed: true,
      severity: 'high'
    };
  }
  
  // Verificar se há uma base legal adequada para tomada de decisão automatizada
  // POPIA requer bases específicas para automatização, incluindo consentimento ou contrato
  const validBases = ['consent', 'contract_performance', 'legal_obligation'];
  
  if (!validBases.includes(request.processingMetadata.lawfulBasis)) {
    return {
      validationType: POPIAValidationType.AUTOMATED_DECISION_MAKING,
      passed: false,
      failureReason: 'Base legal insuficiente para tomada de decisão automatizada',
      requiredActions: ['Obter consentimento específico para tomada de decisão automatizada', 
                      'Ou estabelecer necessidade contratual para automatização'],
      severity: 'high'
    };
  }
  
  // Para consentimento, verificar se ele existe e é específico
  if (request.processingMetadata.lawfulBasis === 'consent') {
    if (!request.consentReference) {
      return {
        validationType: POPIAValidationType.AUTOMATED_DECISION_MAKING,
        passed: false,
        failureReason: 'Consentimento não fornecido para tomada de decisão automatizada',
        requiredActions: ['Obter consentimento específico para tomada de decisão automatizada'],
        severity: 'high'
      };
    }
    
    // Aqui verificaríamos se o consentimento é específico para tomada de decisão automatizada
    // Simulando para este exemplo
  }
  
  // Verificar se há mecanismos para intervenção humana
  // Simulado para este exemplo - verificaria processos documentados para revisão humana
  const hasHumanIntervention = request.processingMetadata.securityMeasures.some(
    measure => measure.toLowerCase().includes('human_review') || 
              measure.toLowerCase().includes('appeal_process')
  );
  
  if (!hasHumanIntervention) {
    return {
      validationType: POPIAValidationType.AUTOMATED_DECISION_MAKING,
      passed: false,
      failureReason: 'Falta de mecanismos para intervenção humana em decisões automatizadas',
      requiredActions: ['Implementar processo de revisão humana para decisões automatizadas', 
                      'Estabelecer canal para contestação de decisões automatizadas'],
      severity: 'high'
    };
  }
  
  return {
    validationType: POPIAValidationType.AUTOMATED_DECISION_MAKING,
    passed: true,
    severity: 'high'
  };
}

/**
 * Determina restrições de processamento com base em resultados de validação
 * 
 * @param validationResults Resultados individuais de validação
 * @returns Lista de restrições de processamento
 */
export function determineProcessingRestrictions(this: POPIAValidator, validationResults: POPIAValidationResult[]): string[] {
  const restrictions: string[] = [];
  
  // Mapear falhas de validação para restrições específicas
  for (const result of validationResults) {
    if (!result.passed) {
      switch (result.validationType) {
        case POPIAValidationType.LAWFULNESS_CHECK:
          restrictions.push('Processamento restrito devido a base legal inadequada');
          break;
        case POPIAValidationType.MINIMALITY:
          restrictions.push('Processamento limitado apenas aos campos de dados necessários');
          break;
        case POPIAValidationType.PURPOSE_SPECIFICATION:
          restrictions.push('Processamento restrito apenas à finalidade original especificada');
          break;
        case POPIAValidationType.SECURITY_SAFEGUARDS:
          restrictions.push('Implementação de medidas adicionais de segurança obrigatórias');
          break;
        case POPIAValidationType.CROSS_BORDER_TRANSFER:
          restrictions.push('Transferência transfronteiriça de dados proibida até adequação');
          break;
        case POPIAValidationType.SPECIAL_PERSONAL_INFO:
          restrictions.push('Processamento de dados sensíveis suspenso até conformidade');
          break;
        case POPIAValidationType.AUTOMATED_DECISION_MAKING:
          restrictions.push('Decisões automatizadas devem ter revisão humana obrigatória');
          break;
        default:
          // Para outros tipos de falha, adicionar restrição genérica
          if (result.severity === 'critical' || result.severity === 'high') {
            restrictions.push('Processamento restrito até resolução de não-conformidades');
          }
      }
    }
  }
  
  return [...new Set(restrictions)]; // Remover duplicatas
}