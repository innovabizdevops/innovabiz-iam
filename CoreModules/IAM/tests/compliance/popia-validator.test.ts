/**
 * Testes para o validador POPIA (Protection of Personal Information Act)
 * 
 * Este arquivo contém testes automatizados para validar a implementação do validador
 * de conformidade com a lei POPIA da África do Sul.
 * 
 * @module POPIAValidatorTest
 */

import { POPIAValidator, POPIAValidationRequest, LegalBasisPOPIA } from '../../services/bureau-credito/compliance/popia-validator';
import { Logger } from '../../observability/logging/hook_logger';
import { Metrics } from '../../observability/metrics/hook_metrics';
import { Tracer } from '../../observability/tracing/hook_tracing';

// Mock dos serviços de observabilidade
const mockLogger = {
  info: jest.fn(),
  error: jest.fn(),
  warn: jest.fn(),
  debug: jest.fn()
} as unknown as Logger;

const mockMetrics = {
  increment: jest.fn(),
  gauge: jest.fn(),
  histogram: jest.fn(),
  timing: jest.fn()
} as unknown as Metrics;

const mockTracer = {
  startSpan: jest.fn().mockReturnValue({
    end: jest.fn()
  }),
  injectSpanContext: jest.fn()
} as unknown as Tracer;

describe('POPIAValidator', () => {
  let validator: POPIAValidator;
  let validRequest: POPIAValidationRequest;
  
  beforeEach(() => {
    // Resetar mocks
    jest.clearAllMocks();
    
    // Instanciar validador
    validator = new POPIAValidator(mockLogger, mockMetrics, mockTracer);
    
    // Criar uma requisição válida padrão
    validRequest = {
      userId: 'user-12345',
      tenantId: 'tenant-abc',
      operationId: 'op-12345',
      operationType: 'credit_assessment',
      consentId: 'consent-popia-123',
      dataPurpose: 'Avaliação de crédito para análise de risco financeiro',
      dataCategories: ['personal_details', 'financial_information'],
      dataFields: ['full_name', 'id_number', 'income', 'credit_score'],
      retentionPeriodDays: 365,
      legalBasis: LegalBasisPOPIA.CONSENT,
      specialCategoriesData: false,
      automatedDecisionMaking: false,
      crossBorderTransfer: false,
      securityMeasures: [
        'encryption',
        'access_control',
        'logging',
        'data_minimization'
      ]
    };
  });

  describe('validate', () => {
    it('deve aprovar uma requisição totalmente compatível', async () => {
      // Executar
      const result = await validator.validate(validRequest);
      
      // Verificar
      expect(result.compliant).toBe(true);
      expect(result.processingAllowed).toBe(true);
      expect(result.validationResults.length).toBeGreaterThan(0);
      expect(result.validationResults.every(vr => vr.valid)).toBe(true);
    });

    it('deve rejeitar quando não há base legal válida', async () => {
      // Configurar
      const invalidRequest = {
        ...validRequest,
        legalBasis: undefined
      };
      
      // Executar
      const result = await validator.validate(invalidRequest);
      
      // Verificar
      expect(result.compliant).toBe(false);
      expect(result.processingAllowed).toBe(false);
      
      // Encontrar validação específica que falhou
      const legalBasisValidation = result.validationResults.find(
        vr => vr.validationType === 'legal_basis_validation'
      );
      
      expect(legalBasisValidation).toBeDefined();
      expect(legalBasisValidation?.valid).toBe(false);
    });

    it('deve rejeitar quando a finalidade não está claramente especificada', async () => {
      // Configurar
      const invalidRequest = {
        ...validRequest,
        dataPurpose: ''
      };
      
      // Executar
      const result = await validator.validate(invalidRequest);
      
      // Verificar
      expect(result.compliant).toBe(false);
      
      // Encontrar validação específica que falhou
      const purposeValidation = result.validationResults.find(
        vr => vr.validationType === 'purpose_specification_validation'
      );
      
      expect(purposeValidation).toBeDefined();
      expect(purposeValidation?.valid).toBe(false);
    });

    it('deve rejeitar quando dados especiais são processados sem base legal adequada', async () => {
      // Configurar
      const invalidRequest = {
        ...validRequest,
        specialCategoriesData: true,
        // Não fornecendo uma base legal adequada para dados sensíveis
      };
      
      // Executar
      const result = await validator.validate(invalidRequest);
      
      // Verificar
      expect(result.compliant).toBe(false);
      
      // Encontrar validação específica que falhou
      const specialDataValidation = result.validationResults.find(
        vr => vr.validationType === 'special_categories_validation'
      );
      
      expect(specialDataValidation).toBeDefined();
      expect(specialDataValidation?.valid).toBe(false);
    });

    it('deve alertar (mas não rejeitar) quando o período de retenção é longo', async () => {
      // Configurar
      const warningRequest = {
        ...validRequest,
        retentionPeriodDays: 1095 // 3 anos
      };
      
      // Executar
      const result = await validator.validate(warningRequest);
      
      // Verificar
      expect(result.compliant).toBe(true); // Ainda é compatível
      
      // Encontrar validação específica com aviso
      const retentionValidation = result.validationResults.find(
        vr => vr.validationType === 'retention_period_validation'
      );
      
      expect(retentionValidation).toBeDefined();
      expect(retentionValidation?.valid).toBe(true);
      expect(retentionValidation?.warnings.length).toBeGreaterThan(0);
    });

    it('deve requerer salvaguardas adicionais para tomada de decisão automatizada', async () => {
      // Configurar
      const automatedRequest = {
        ...validRequest,
        automatedDecisionMaking: true,
        securityMeasures: ['encryption'] // Insuficiente para decisão automatizada
      };
      
      // Executar
      const result = await validator.validate(automatedRequest);
      
      // Verificar
      expect(result.compliant).toBe(false);
      
      // Encontrar validação específica que falhou
      const automatedDecisionValidation = result.validationResults.find(
        vr => vr.validationType === 'automated_decision_making_validation'
      );
      
      expect(automatedDecisionValidation).toBeDefined();
      expect(automatedDecisionValidation?.valid).toBe(false);
    });

    it('deve aprovar tomada de decisão automatizada com salvaguardas adequadas', async () => {
      // Configurar
      const automatedWithSafeguards = {
        ...validRequest,
        automatedDecisionMaking: true,
        securityMeasures: [
          'encryption',
          'access_control',
          'logging',
          'human_review',
          'appeal_process',
          'explanation'
        ]
      };
      
      // Executar
      const result = await validator.validate(automatedWithSafeguards);
      
      // Verificar
      expect(result.compliant).toBe(true);
      
      // Encontrar validação específica
      const automatedDecisionValidation = result.validationResults.find(
        vr => vr.validationType === 'automated_decision_making_validation'
      );
      
      expect(automatedDecisionValidation).toBeDefined();
      expect(automatedDecisionValidation?.valid).toBe(true);
    });

    it('deve validar transferências transfronteiriças com proteção adequada', async () => {
      // Configurar
      const crossBorderRequest = {
        ...validRequest,
        crossBorderTransfer: true,
        transferDestinations: ['EU'], // Destino com proteção adequada
        transferSafeguards: ['binding_corporate_rules']
      };
      
      // Executar
      const result = await validator.validate(crossBorderRequest);
      
      // Verificar
      expect(result.compliant).toBe(true);
      
      // Encontrar validação específica
      const crossBorderValidation = result.validationResults.find(
        vr => vr.validationType === 'cross_border_transfer_validation'
      );
      
      expect(crossBorderValidation).toBeDefined();
      expect(crossBorderValidation?.valid).toBe(true);
    });

    it('deve rejeitar transferências transfronteiriças sem proteção adequada', async () => {
      // Configurar
      const crossBorderRequest = {
        ...validRequest,
        crossBorderTransfer: true,
        transferDestinations: ['UnsafeCountry'], // País sem proteção adequada
        transferSafeguards: [] // Sem salvaguardas
      };
      
      // Executar
      const result = await validator.validate(crossBorderRequest);
      
      // Verificar
      expect(result.compliant).toBe(false);
      
      // Encontrar validação específica
      const crossBorderValidation = result.validationResults.find(
        vr => vr.validationType === 'cross_border_transfer_validation'
      );
      
      expect(crossBorderValidation).toBeDefined();
      expect(crossBorderValidation?.valid).toBe(false);
    });

    it('deve registrar métricas ao processar validações', async () => {
      // Executar
      await validator.validate(validRequest);
      
      // Verificar
      expect(mockMetrics.increment).toHaveBeenCalledWith(
        'popia_validations_total',
        expect.any(Object)
      );
      expect(mockMetrics.timing).toHaveBeenCalledWith(
        'popia_validation_time',
        expect.any(Number),
        expect.any(Object)
      );
    });
    
    it('deve criar um span de rastreamento durante a validação', async () => {
      // Executar
      await validator.validate(validRequest);
      
      // Verificar
      expect(mockTracer.startSpan).toHaveBeenCalledWith(
        'popia.validate',
        expect.any(Object)
      );
      
      // Verificar que o span foi finalizado
      const mockSpan = mockTracer.startSpan.mock.results[0].value;
      expect(mockSpan.end).toHaveBeenCalled();
    });
  });
});