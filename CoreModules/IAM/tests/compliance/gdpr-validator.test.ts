/**
 * Testes para o validador GDPR (General Data Protection Regulation)
 * 
 * Este arquivo contém testes automatizados para validar a implementação do validador
 * de conformidade com o GDPR da União Europeia.
 * 
 * @module GDPRValidatorTest
 */

import { GDPRValidator, GDPRValidationRequest, LegalBasisGDPR } from '../../services/bureau-credito/compliance/gdpr-validator';
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

describe('GDPRValidator', () => {
  let validator: GDPRValidator;
  let validRequest: GDPRValidationRequest;
  
  beforeEach(() => {
    // Resetar mocks
    jest.clearAllMocks();
    
    // Instanciar validador
    validator = new GDPRValidator(mockLogger, mockMetrics, mockTracer);
    
    // Criar uma requisição válida padrão
    validRequest = {
      userId: 'user-12345',
      tenantId: 'tenant-abc',
      operationId: 'op-12345',
      operationType: 'credit_assessment',
      consentId: 'consent-gdpr-123',
      dataPurpose: 'Avaliação de crédito para análise de risco financeiro',
      dataCategories: ['personal_data', 'financial_data'],
      dataFields: ['full_name', 'email', 'income', 'credit_score'],
      retentionPeriodDays: 730,
      legalBasis: LegalBasisGDPR.CONSENT,
      specialCategoryData: false,
      automatedDecisionMaking: false,
      crossBorderTransfer: false,
      adequacyDecision: false,
      appropriateSafeguards: [],
      securityMeasures: [
        'encryption',
        'access_control',
        'logging',
        'data_minimization'
      ],
      dpia: false,
      dpiaResults: null
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
        vr => vr.validationType === 'lawfulness_validation'
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
        vr => vr.validationType === 'purpose_limitation_validation'
      );
      
      expect(purposeValidation).toBeDefined();
      expect(purposeValidation?.valid).toBe(false);
    });

    it('deve rejeitar quando categorias especiais são processadas sem base legal específica', async () => {
      // Configurar
      const invalidRequest = {
        ...validRequest,
        specialCategoryData: true,
        legalBasis: LegalBasisGDPR.LEGITIMATE_INTEREST // Base legal insuficiente para dados sensíveis
      };
      
      // Executar
      const result = await validator.validate(invalidRequest);
      
      // Verificar
      expect(result.compliant).toBe(false);
      
      // Encontrar validação específica que falhou
      const specialCategoryValidation = result.validationResults.find(
        vr => vr.validationType === 'special_categories_validation'
      );
      
      expect(specialCategoryValidation).toBeDefined();
      expect(specialCategoryValidation?.valid).toBe(false);
    });

    it('deve rejeitar período de retenção indefinido', async () => {
      // Configurar
      const invalidRequest = {
        ...validRequest,
        retentionPeriodDays: undefined
      };
      
      // Executar
      const result = await validator.validate(invalidRequest);
      
      // Verificar
      expect(result.compliant).toBe(false);
      
      // Encontrar validação específica que falhou
      const retentionValidation = result.validationResults.find(
        vr => vr.validationType === 'storage_limitation_validation'
      );
      
      expect(retentionValidation).toBeDefined();
      expect(retentionValidation?.valid).toBe(false);
    });

    it('deve requerer DPIA para tomada de decisão automatizada de alto risco', async () => {
      // Configurar
      const automatedRequest = {
        ...validRequest,
        automatedDecisionMaking: true,
        dpia: false,
        dpiaResults: null
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

    it('deve aprovar tomada de decisão automatizada com DPIA e salvaguardas adequadas', async () => {
      // Configurar
      const automatedWithSafeguards = {
        ...validRequest,
        automatedDecisionMaking: true,
        dpia: true,
        dpiaResults: {
          highRiskIdentified: true,
          mitigationMeasures: [
            'human_intervention',
            'algorithmic_transparency',
            'appeal_process',
            'regular_accuracy_review'
          ],
          riskLevel: 'medium',
          acceptableRisk: true
        },
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

    it('deve validar transferências internacionais com decisão de adequação', async () => {
      // Configurar
      const transferRequest = {
        ...validRequest,
        crossBorderTransfer: true,
        adequacyDecision: true,
        transferDestinations: ['Switzerland', 'New Zealand']
      };
      
      // Executar
      const result = await validator.validate(transferRequest);
      
      // Verificar
      expect(result.compliant).toBe(true);
      
      // Encontrar validação específica
      const transferValidation = result.validationResults.find(
        vr => vr.validationType === 'cross_border_transfer_validation'
      );
      
      expect(transferValidation).toBeDefined();
      expect(transferValidation?.valid).toBe(true);
    });

    it('deve validar transferências internacionais com salvaguardas apropriadas', async () => {
      // Configurar
      const transferRequest = {
        ...validRequest,
        crossBorderTransfer: true,
        adequacyDecision: false,
        appropriateSafeguards: ['binding_corporate_rules', 'standard_contractual_clauses'],
        transferDestinations: ['United States', 'India']
      };
      
      // Executar
      const result = await validator.validate(transferRequest);
      
      // Verificar
      expect(result.compliant).toBe(true);
      
      // Encontrar validação específica
      const transferValidation = result.validationResults.find(
        vr => vr.validationType === 'cross_border_transfer_validation'
      );
      
      expect(transferValidation).toBeDefined();
      expect(transferValidation?.valid).toBe(true);
    });

    it('deve rejeitar transferências internacionais sem proteção adequada', async () => {
      // Configurar
      const transferRequest = {
        ...validRequest,
        crossBorderTransfer: true,
        adequacyDecision: false,
        appropriateSafeguards: [],
        transferDestinations: ['United States', 'Brazil']
      };
      
      // Executar
      const result = await validator.validate(transferRequest);
      
      // Verificar
      expect(result.compliant).toBe(false);
      
      // Encontrar validação específica
      const transferValidation = result.validationResults.find(
        vr => vr.validationType === 'cross_border_transfer_validation'
      );
      
      expect(transferValidation).toBeDefined();
      expect(transferValidation?.valid).toBe(false);
    });

    it('deve verificar registro de métricas durante a validação', async () => {
      // Executar
      await validator.validate(validRequest);
      
      // Verificar
      expect(mockMetrics.increment).toHaveBeenCalledWith(
        'gdpr_validations_total',
        expect.any(Object)
      );
      expect(mockMetrics.timing).toHaveBeenCalledWith(
        'gdpr_validation_time',
        expect.any(Number),
        expect.any(Object)
      );
    });
    
    it('deve criar um span de rastreamento durante a validação', async () => {
      // Executar
      await validator.validate(validRequest);
      
      // Verificar
      expect(mockTracer.startSpan).toHaveBeenCalledWith(
        'gdpr.validate',
        expect.any(Object)
      );
      
      // Verificar que o span foi finalizado
      const mockSpan = mockTracer.startSpan.mock.results[0].value;
      expect(mockSpan.end).toHaveBeenCalled();
    });
  });
});