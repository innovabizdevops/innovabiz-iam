/**
 * Testes para o Integrador de Validadores de Conformidade
 * 
 * Este arquivo contém testes automatizados para validar a implementação do integrador
 * que coordena validações de conformidade com múltiplas regulamentações (GDPR, LGPD, POPIA).
 * 
 * @module ComplianceValidatorIntegratorTest
 */

import { 
  ComplianceValidatorIntegrator, 
  ComplianceValidationRequest, 
  Region,
  ConsolidatedComplianceResult
} from '../../services/bureau-credito/compliance/compliance-validator-integrator';

import { GDPRValidator } from '../../services/bureau-credito/compliance/gdpr-validator';
import { LGPDValidator } from '../../services/bureau-credito/compliance/lgpd-validator';
import { POPIAValidator } from '../../services/bureau-credito/compliance/popia-validator';

import { Logger } from '../../observability/logging/hook_logger';
import { Metrics } from '../../observability/metrics/hook_metrics';
import { Tracer } from '../../observability/tracing/hook_tracing';

// Mock dos validadores individuais
jest.mock('../../services/bureau-credito/compliance/gdpr-validator');
jest.mock('../../services/bureau-credito/compliance/lgpd-validator');
jest.mock('../../services/bureau-credito/compliance/popia-validator');

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

describe('ComplianceValidatorIntegrator', () => {
  let integrator: ComplianceValidatorIntegrator;
  let validRequest: ComplianceValidationRequest;

  // Implementação mock para validadores individuais
  const mockValidatorResponse = (compliant: boolean, processingAllowed: boolean) => {
    return {
      compliant,
      processingAllowed,
      validationResults: [
        {
          validationType: 'test_validation',
          valid: compliant,
          details: compliant ? 'Valid test' : 'Invalid test',
          warnings: [],
          errors: compliant ? [] : ['Test error']
        }
      ],
      requiredActions: compliant ? [] : ['Fix compliance issues']
    };
  };
  
  beforeEach(() => {
    // Resetar mocks
    jest.clearAllMocks();
    
    // Mock das implementações dos validadores
    (GDPRValidator as jest.MockedClass<typeof GDPRValidator>).mockImplementation(() => {
      return {
        validate: jest.fn().mockResolvedValue(mockValidatorResponse(true, true))
      } as unknown as GDPRValidator;
    });
    
    (LGPDValidator as jest.MockedClass<typeof LGPDValidator>).mockImplementation(() => {
      return {
        validate: jest.fn().mockResolvedValue(mockValidatorResponse(true, true))
      } as unknown as LGPDValidator;
    });
    
    (POPIAValidator as jest.MockedClass<typeof POPIAValidator>).mockImplementation(() => {
      return {
        validate: jest.fn().mockResolvedValue(mockValidatorResponse(true, true))
      } as unknown as POPIAValidator;
    });
    
    // Instanciar integrador
    integrator = new ComplianceValidatorIntegrator(mockLogger, mockMetrics, mockTracer);
    
    // Criar uma requisição válida padrão
    validRequest = {
      userId: 'user-12345',
      tenantId: 'tenant-abc',
      operationId: 'op-12345',
      operationType: 'credit_assessment',
      dataSubjectCountry: Region.EUROPEAN_UNION,
      dataProcessingCountry: 'angola',
      businessTargetCountries: ['angola', 'brazil', 'south_africa', 'eu'],
      consentReferences: {
        gdpr: 'consent-gdpr-123',
        lgpd: 'consent-lgpd-456',
        popia: 'consent-popia-789'
      },
      dataPurpose: 'Avaliação de crédito para análise de risco financeiro',
      dataCategories: ['personal_details', 'financial_information'],
      dataFields: ['full_name', 'email', 'income', 'credit_score'],
      retentionPeriodDays: 730,
      processingLegalBasis: 'consent',
      specialCategories: false,
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
    it('deve determinar corretamente as regulamentações aplicáveis com base no país do titular dos dados', async () => {
      // Configurar
      validRequest.dataSubjectCountry = Region.EUROPEAN_UNION;
      
      // Executar
      await integrator.validate(validRequest);
      
      // Verificar que o validador GDPR foi chamado
      expect(GDPRValidator).toHaveBeenCalled();
      expect((GDPRValidator as jest.Mock).mock.instances[0].validate).toHaveBeenCalled();
    });

    it('deve aplicar validadores com base no país de processamento', async () => {
      // Configurar
      validRequest.dataSubjectCountry = 'angola';
      validRequest.dataProcessingCountry = 'brazil';
      
      // Executar
      await integrator.validate(validRequest);
      
      // Verificar que o validador LGPD foi chamado porque os dados são processados no Brasil
      expect(LGPDValidator).toHaveBeenCalled();
      expect((LGPDValidator as jest.Mock).mock.instances[0].validate).toHaveBeenCalled();
    });

    it('deve aplicar validadores com base nos países-alvo de negócios', async () => {
      // Configurar
      validRequest.dataSubjectCountry = 'angola';
      validRequest.dataProcessingCountry = 'angola';
      validRequest.businessTargetCountries = ['south_africa', 'angola'];
      
      // Executar
      await integrator.validate(validRequest);
      
      // Verificar que o validador POPIA foi chamado porque África do Sul é um país-alvo
      expect(POPIAValidator).toHaveBeenCalled();
      expect((POPIAValidator as jest.Mock).mock.instances[0].validate).toHaveBeenCalled();
    });

    it('deve aplicar múltiplos validadores quando apropriado', async () => {
      // Configurar
      validRequest.dataSubjectCountry = 'brazil';
      validRequest.dataProcessingCountry = 'angola';
      validRequest.businessTargetCountries = ['south_africa', 'eu', 'brazil'];
      
      // Executar
      await integrator.validate(validRequest);
      
      // Verificar que todos os validadores foram chamados
      expect(GDPRValidator).toHaveBeenCalled();
      expect(LGPDValidator).toHaveBeenCalled();
      expect(POPIAValidator).toHaveBeenCalled();
      
      expect((GDPRValidator as jest.Mock).mock.instances[0].validate).toHaveBeenCalled();
      expect((LGPDValidator as jest.Mock).mock.instances[0].validate).toHaveBeenCalled();
      expect((POPIAValidator as jest.Mock).mock.instances[0].validate).toHaveBeenCalled();
    });

    it('deve retornar resultado consolidado com conformidade geral quando todos os validadores aprovam', async () => {
      // Configurar
      validRequest.dataSubjectCountry = 'brazil';
      validRequest.businessTargetCountries = ['south_africa', 'eu'];
      
      // Todos os validadores retornam conformidade
      
      // Executar
      const result = await integrator.validate(validRequest);
      
      // Verificar
      expect(result.overallCompliant).toBe(true);
      expect(result.processingAllowed).toBe(true);
      expect(result.applicableRegulations.length).toBeGreaterThan(0);
    });

    it('deve retornar não-conformidade geral quando qualquer validador falha', async () => {
      // Configurar
      validRequest.dataSubjectCountry = 'brazil';
      validRequest.businessTargetCountries = ['south_africa', 'eu'];
      
      // Fazer com que o validador GDPR retorne não-conformidade
      (GDPRValidator as jest.MockedClass<typeof GDPRValidator>).mockImplementation(() => {
        return {
          validate: jest.fn().mockResolvedValue(mockValidatorResponse(false, false))
        } as unknown as GDPRValidator;
      });
      
      // Executar
      const result = await integrator.validate(validRequest);
      
      // Verificar
      expect(result.overallCompliant).toBe(false);
      expect(result.processingAllowed).toBe(false);
    });

    it('deve permitir processamento com restrições quando validadores aprovam com condições', async () => {
      // Configurar
      validRequest.dataSubjectCountry = 'brazil';
      validRequest.businessTargetCountries = ['south_africa', 'eu'];
      
      // O validador GDPR permite processamento mas com restrições
      (GDPRValidator as jest.MockedClass<typeof GDPRValidator>).mockImplementation(() => {
        return {
          validate: jest.fn().mockResolvedValue({
            compliant: false,
            processingAllowed: true, // Pode processar com restrições
            validationResults: [
              {
                validationType: 'retention_validation',
                valid: false,
                details: 'Período de retenção muito longo',
                warnings: ['Considere reduzir o período de retenção'],
                errors: []
              }
            ],
            requiredActions: ['Reduzir período de retenção']
          })
        } as unknown as GDPRValidator;
      });
      
      // Executar
      const result = await integrator.validate(validRequest);
      
      // Verificar
      expect(result.overallCompliant).toBe(false);
      expect(result.processingAllowed).toBe(true);
      expect(result.processingRestrictions).toBeDefined();
      expect(result.processingRestrictions.length).toBeGreaterThan(0);
      expect(result.requiredActions).toContain('Reduzir período de retenção');
    });

    it('deve consolidar ações requeridas de todos os validadores', async () => {
      // Configurar
      validRequest.dataSubjectCountry = 'brazil';
      validRequest.businessTargetCountries = ['south_africa'];
      
      // O validador LGPD requer ações
      (LGPDValidator as jest.MockedClass<typeof LGPDValidator>).mockImplementation(() => {
        return {
          validate: jest.fn().mockResolvedValue({
            compliant: false,
            processingAllowed: true,
            validationResults: [
              {
                validationType: 'consent_validation',
                valid: false,
                details: 'Consentimento insuficiente',
                warnings: [],
                errors: ['Consentimento não específico o suficiente']
              }
            ],
            requiredActions: ['Obter consentimento específico']
          })
        } as unknown as LGPDValidator;
      });
      
      // O validador POPIA também requer ações
      (POPIAValidator as jest.MockedClass<typeof POPIAValidator>).mockImplementation(() => {
        return {
          validate: jest.fn().mockResolvedValue({
            compliant: false,
            processingAllowed: true,
            validationResults: [
              {
                validationType: 'security_measures_validation',
                valid: false,
                details: 'Medidas de segurança insuficientes',
                warnings: [],
                errors: ['Faltam medidas de segurança organizacionais']
              }
            ],
            requiredActions: ['Implementar medidas de segurança organizacionais']
          })
        } as unknown as POPIAValidator;
      });
      
      // Executar
      const result = await integrator.validate(validRequest);
      
      // Verificar
      expect(result.overallCompliant).toBe(false);
      expect(result.requiredActions.length).toBe(2);
      expect(result.requiredActions).toContain('Obter consentimento específico');
      expect(result.requiredActions).toContain('Implementar medidas de segurança organizacionais');
    });

    it('deve registrar métricas de tempo de validação', async () => {
      // Executar
      await integrator.validate(validRequest);
      
      // Verificar
      expect(mockMetrics.timing).toHaveBeenCalledWith(
        'compliance_validation_time',
        expect.any(Number),
        expect.any(Object)
      );
    });

    it('deve registrar métricas de conformidade', async () => {
      // Executar
      await integrator.validate(validRequest);
      
      // Verificar
      expect(mockMetrics.increment).toHaveBeenCalledWith(
        'compliance_validation_total',
        expect.any(Object)
      );
    });
    
    it('deve criar um span de rastreamento durante a validação', async () => {
      // Executar
      await integrator.validate(validRequest);
      
      // Verificar
      expect(mockTracer.startSpan).toHaveBeenCalledWith(
        'compliance.validate',
        expect.any(Object)
      );
      
      // Verificar que o span foi finalizado
      const mockSpan = mockTracer.startSpan.mock.results[0].value;
      expect(mockSpan.end).toHaveBeenCalled();
    });
  });
});