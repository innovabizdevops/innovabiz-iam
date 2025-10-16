/**
 * Exemplo de uso do Conector IAM-Compliance
 * 
 * Este arquivo demonstra como utilizar o IAMComplianceConnector em um fluxo
 * de autorização real, integrando validações de conformidade com decisões de acesso.
 * 
 * @module IAMComplianceExample
 */

import { Logger } from '../../../observability/logging/hook_logger';
import { Metrics } from '../../../observability/metrics/hook_metrics';
import { Tracer } from '../../../observability/tracing/hook_tracing';

import { 
  IAMComplianceConnector,
  AccessAuthorizationRequest,
  AccessAuthorizationResult
} from './iam-compliance-connector';

// Configuração das ferramentas de observabilidade
const logger = new Logger({ service: 'iam-compliance-integration' });
const metrics = new Metrics({ service: 'iam-compliance-integration' });
const tracer = new Tracer({ service: 'iam-compliance-integration' });

/**
 * Simula um fluxo completo de autorização de acesso com verificação de conformidade
 * para uma operação de avaliação de crédito
 */
async function simulateCreditAssessmentAuthorization() {
  console.log('Simulando autorização para avaliação de crédito com verificação de conformidade...');
  
  // Inicializar o conector
  const complianceConnector = new IAMComplianceConnector(logger, metrics, tracer);
  
  // Criar uma requisição de autorização (simulando contexto de um usuário brasileiro)
  const authRequest: AccessAuthorizationRequest = {
    userId: 'user-12345',
    tenantId: 'financial-institution-xyz',
    requestId: 'auth-req-' + Date.now(),
    resourceType: 'credit_profile',
    resourceId: 'profile-98765',
    operationType: 'credit_assessment',
    dataCategories: ['personal_details', 'financial_information', 'credit_history'],
    userContext: {
      roles: ['credit_analyst'],
      permissions: ['bureau:credit:assess', 'credit_profile:read'],
      authenticationLevel: 3,
      authenticationFactors: ['password', 'mfa', 'device_fingerprint'],
      consentReferences: {
        gdpr: null,
        lgpd: 'consent-lgpd-123',
        popia: null
      },
      geoLocation: {
        country: 'brazil',
        region: 'sao_paulo'
      }
    }
  };
  
  console.log(`Processando requisição de autorização ${authRequest.requestId}...`);
  
  try {
    // Executar a verificação de autorização com validação de conformidade
    const authResult = await complianceConnector.authorizeWithComplianceCheck(authRequest);
    
    // Exibir o resultado
    console.log('\nResultado da autorização:');
    console.log(`- Autorizado: ${authResult.authorized}`);
    console.log(`- Conformidade verificada: ${authResult.complianceVerified}`);
    
    if (authResult.requiredAuthenticationLevel && 
        authResult.requiredAuthenticationLevel > authRequest.userContext.authenticationLevel) {
      console.log(`- Nível de autenticação necessário: ${authResult.requiredAuthenticationLevel}`);
    }
    
    if (authResult.requiredAuthenticationFactors && authResult.requiredAuthenticationFactors.length > 0) {
      console.log('- Fatores de autenticação necessários:');
      authResult.requiredAuthenticationFactors.forEach(factor => {
        const isMissing = !authRequest.userContext.authenticationFactors.includes(factor);
        console.log(`  - ${factor}${isMissing ? ' (ausente)' : ''}`);
      });
    }
    
    if (authResult.restrictionReasons && authResult.restrictionReasons.length > 0) {
      console.log('- Razões para restrição:');
      authResult.restrictionReasons.forEach((reason, index) => {
        console.log(`  ${index + 1}. ${reason}`);
      });
    }
    
    if (authResult.requiredActions && authResult.requiredActions.length > 0) {
      console.log('- Ações necessárias:');
      authResult.requiredActions.forEach((action, index) => {
        console.log(`  ${index + 1}. ${action}`);
      });
    }
    
    // Se autorizado, simular continuidade do fluxo
    if (authResult.authorized) {
      console.log('\nAutorização aprovada. Prosseguindo com a avaliação de crédito...');
      
      // Aqui executaríamos a lógica real de avaliação de crédito
      // ...
      
      console.log('Avaliação de crédito concluída com sucesso.');
    } else {
      console.log('\nAutorização negada. A avaliação de crédito não pode ser realizada.');
      console.log('Registrando tentativa de acesso negada para auditoria...');
    }
  } catch (error) {
    console.error('Erro durante o processo de autorização:', error);
  }
}

/**
 * Exemplo de uso do conector para um cenário com dados sensíveis e problemas de conformidade
 */
async function simulateHighRiskAuthorization() {
  console.log('\n\nSimulando autorização para operação de alto risco com dados sensíveis...');
  
  // Inicializar o conector
  const complianceConnector = new IAMComplianceConnector(logger, metrics, tracer);
  
  // Criar uma requisição de autorização com dados sensíveis (da África do Sul)
  const authRequest: AccessAuthorizationRequest = {
    userId: 'user-6789',
    tenantId: 'financial-institution-xyz',
    requestId: 'auth-req-' + Date.now(),
    resourceType: 'customer_profile',
    resourceId: 'profile-54321',
    operationType: 'credit_assessment',
    dataCategories: ['personal_details', 'financial_information', 'health', 'biometric'],
    userContext: {
      roles: ['credit_manager'],
      permissions: ['bureau:credit:assess', 'customer_profile:read'],
      authenticationLevel: 2, // Insuficiente para dados sensíveis
      authenticationFactors: ['password', 'mfa'], // Falta fator biométrico
      consentReferences: {}, // Sem referências de consentimento
      geoLocation: {
        country: 'south_africa',
        region: 'gauteng'
      }
    }
  };
  
  console.log(`Processando requisição de autorização ${authRequest.requestId}...`);
  
  try {
    // Executar a verificação de autorização com validação de conformidade
    const authResult = await complianceConnector.authorizeWithComplianceCheck(authRequest);
    
    // Exibir o resultado
    console.log('\nResultado da autorização:');
    console.log(`- Autorizado: ${authResult.authorized}`);
    console.log(`- Conformidade verificada: ${authResult.complianceVerified}`);
    
    if (authResult.requiredAuthenticationLevel && 
        authResult.requiredAuthenticationLevel > authRequest.userContext.authenticationLevel) {
      console.log(`- Nível de autenticação necessário: ${authResult.requiredAuthenticationLevel} (atual: ${authRequest.userContext.authenticationLevel})`);
    }
    
    if (authResult.requiredAuthenticationFactors && authResult.requiredAuthenticationFactors.length > 0) {
      console.log('- Fatores de autenticação necessários:');
      authResult.requiredAuthenticationFactors.forEach(factor => {
        const isMissing = !authRequest.userContext.authenticationFactors.includes(factor);
        console.log(`  - ${factor}${isMissing ? ' (ausente)' : ''}`);
      });
    }
    
    if (authResult.restrictionReasons && authResult.restrictionReasons.length > 0) {
      console.log('- Razões para restrição:');
      authResult.restrictionReasons.forEach((reason, index) => {
        console.log(`  ${index + 1}. ${reason}`);
      });
    }
    
    if (authResult.requiredActions && authResult.requiredActions.length > 0) {
      console.log('- Ações necessárias:');
      authResult.requiredActions.forEach((action, index) => {
        console.log(`  ${index + 1}. ${action}`);
      });
    }
    
    // Se autorizado (improvável neste cenário), simular continuidade do fluxo
    if (authResult.authorized) {
      console.log('\nAutorização aprovada. Prosseguindo com a operação...');
    } else {
      console.log('\nAutorização negada devido a problemas de conformidade e/ou autenticação insuficiente.');
      console.log('Registrando evento de segurança para revisão...');
    }
  } catch (error) {
    console.error('Erro durante o processo de autorização:', error);
  }
}

/**
 * Função principal para executar os exemplos
 */
async function runExamples() {
  try {
    await simulateCreditAssessmentAuthorization();
    await simulateHighRiskAuthorization();
    console.log('\nExemplos concluídos com sucesso.');
  } catch (error) {
    console.error('Erro ao executar exemplos:', error);
  }
}

// Executar se chamado diretamente
if (require.main === module) {
  runExamples()
    .then(() => console.log('Execução dos exemplos finalizada.'))
    .catch(err => console.error('Erro na execução dos exemplos:', err));
}