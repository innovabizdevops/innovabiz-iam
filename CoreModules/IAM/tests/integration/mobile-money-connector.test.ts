/**
 * @file mobile-money-connector.test.ts
 * @description Testes de unidade para o conector Mobile Money
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

import { MobileMoneyConnector, MobileMoneyConnectorConfig, PhoneVerificationRequest, OtpValidationRequest } from '../../integration/connectors/mobile-money-connector';
import axios from 'axios';
import { Logger } from '../../observability/logging/hook_logger';
import { MetricsCollector } from '../../observability/metrics/hook_metrics';
import { TracingProvider } from '../../observability/tracing/hook_tracing';
import { ConnectorStatus } from '../../integration/connectors/base-connector';

// Mock para o axios
jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

// Mocks para observabilidade
const mockLogger: jest.Mocked<Logger> = {
  debug: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
  error: jest.fn(),
  log: jest.fn(),
};

const mockMetrics: jest.Mocked<MetricsCollector> = {
  incrementCounter: jest.fn(),
  recordValue: jest.fn(),
  recordEvent: jest.fn(),
};

const mockTracer: jest.Mocked<TracingProvider> = {
  createSpan: jest.fn().mockReturnValue({
    addEvent: jest.fn(),
    end: jest.fn(),
  }),
};

describe('MobileMoneyConnector', () => {
  let connector: MobileMoneyConnector;
  let config: MobileMoneyConnectorConfig;
  const mockAxiosCreate = jest.fn();
  const mockAxiosInstance = {
    get: jest.fn(),
    post: jest.fn(),
    interceptors: {
      request: { use: jest.fn() },
      response: { use: jest.fn() },
    },
  };

  beforeEach(() => {
    jest.clearAllMocks();
    
    // Configuração padrão para testes
    config = {
      baseUrl: 'https://api.mobilemoney.test',
      name: 'MobileMoneyTest',
      providerId: 'test-provider',
      apiVersion: '1.0',
      countryCode: 'AO',
      timeoutMs: 5000,
      auth: {
        type: 'apiKey',
        credentials: {
          apiKey: 'test-api-key',
        },
      },
      observability: {
        loggingEnabled: true,
        metricsEnabled: true,
        tracingEnabled: true,
      },
      phoneValidation: {
        enabled: true,
        existenceCheck: true,
        statusCheck: true,
      },
      identityVerification: {
        enabled: true,
        requiredLevel: 'medium',
        documentVerificationEnabled: true,
      },
      angolaSettings: {
        siacIntegration: {
          enabled: true,
          endpoint: 'https://siac-api.test',
        },
        bnaCompliance: {
          enabled: true,
          level: 'enhanced',
        },
      },
    };

    // Configurar mock do axios.create
    mockedAxios.create.mockReturnValue(mockAxiosInstance as any);
    
    // Criar instância do conector
    connector = new MobileMoneyConnector(
      config,
      mockLogger,
      mockMetrics,
      mockTracer
    );
  });

  describe('initialize', () => {
    it('deve inicializar o conector com sucesso', async () => {
      // Mock para health check
      mockAxiosInstance.get.mockResolvedValueOnce({
        data: {
          status: 'UP',
          version: '1.0.0',
          services: { database: 'UP', cache: 'UP' },
          environment: 'test',
          operators: ['UNITEL', 'MOVICEL'],
        },
      });

      // Inicializar o conector
      const result = await connector.initialize();

      // Verificações
      expect(result).toBe(true);
      expect(mockedAxios.create).toHaveBeenCalledWith(
        expect.objectContaining({
          baseURL: config.baseUrl,
          timeout: config.timeoutMs,
          headers: expect.objectContaining({
            'X-API-Key': config.auth.credentials.apiKey,
            'X-API-Version': config.apiVersion,
            'X-Provider-ID': config.providerId,
            'X-Country-Code': config.countryCode,
          }),
        })
      );
      expect(mockAxiosInstance.interceptors.request.use).toHaveBeenCalled();
      expect(mockAxiosInstance.interceptors.response.use).toHaveBeenCalled();
      expect(mockLogger.info).toHaveBeenCalledWith(expect.stringContaining('inicializado'));
    });

    it('deve falhar na inicialização quando health check falha', async () => {
      // Mock para health check falhando
      mockAxiosInstance.get.mockRejectedValueOnce(new Error('Serviço indisponível'));

      // Inicializar o conector
      const result = await connector.initialize();

      // Verificações
      expect(result).toBe(false);
      expect(mockLogger.error).toHaveBeenCalledWith(expect.stringContaining('Erro'));
    });

    it('deve validar a configuração e lançar erro se inválida', async () => {
      // Configuração inválida sem providerId
      const invalidConfig = { ...config, providerId: '' };
      
      // Criar instância com configuração inválida
      const invalidConnector = new MobileMoneyConnector(
        invalidConfig as MobileMoneyConnectorConfig,
        mockLogger,
        mockMetrics,
        mockTracer
      );

      // Deve falhar na inicialização
      await expect(invalidConnector.initialize()).rejects.toThrow('providerId é obrigatório');
    });
  });

  describe('healthCheck', () => {
    it('deve retornar status CONNECTED quando o serviço está UP', async () => {
      // Mock para health check
      mockAxiosInstance.get.mockResolvedValueOnce({
        data: {
          status: 'UP',
          version: '1.0.0',
          services: { database: 'UP', cache: 'UP' },
          environment: 'test',
          operators: ['UNITEL', 'MOVICEL'],
        },
      });

      // Executar health check
      const result = await connector.healthCheck();

      // Verificações
      expect(result.status).toBe(ConnectorStatus.CONNECTED);
      expect(result.details).toMatchObject({
        apiVersion: '1.0.0',
        services: expect.any(Object),
        environment: 'test',
        operators: expect.any(Array),
      });
      expect(result.latencyMs).toBeGreaterThanOrEqual(0);
    });

    it('deve retornar status ERROR quando o health check falha', async () => {
      // Mock para health check falhando
      mockAxiosInstance.get.mockRejectedValueOnce(new Error('Serviço indisponível'));

      // Executar health check
      const result = await connector.healthCheck();

      // Verificações
      expect(result.status).toBe(ConnectorStatus.ERROR);
      expect(result.details).toMatchObject({
        error: expect.stringContaining('Serviço indisponível'),
      });
    });
  });

  describe('startPhoneVerification', () => {
    it('deve iniciar verificação de telefone com sucesso', async () => {
      // Dados de requisição
      const request: PhoneVerificationRequest = {
        phoneNumber: '+244912345678',
        userId: 'user-123',
        verificationType: 'otp',
        verificationOptions: {
          expirationSeconds: 300,
          otpLength: 6,
          preferredChannel: 'sms',
        },
      };

      // Mock para resposta
      const mockResponse = {
        data: {
          verificationId: 'verify-123',
          status: 'initiated',
          maskedPhoneNumber: '+24491****78',
          verificationType: 'otp',
          expiresAt: new Date().toISOString(),
        },
      };
      mockAxiosInstance.post.mockResolvedValueOnce(mockResponse);

      // Executar verificação
      const result = await connector.startPhoneVerification(request);

      // Verificações
      expect(result).toEqual(mockResponse.data);
      expect(mockAxiosInstance.post).toHaveBeenCalledWith(
        '/v1/phone/verify',
        expect.objectContaining({
          phoneNumber: request.phoneNumber,
          userId: request.userId,
          verificationType: request.verificationType,
          providerId: config.providerId,
          countryCode: config.countryCode,
        }),
        expect.anything()
      );
      expect(mockMetrics.incrementCounter).toHaveBeenCalledWith(
        'mobileMoney.phoneVerification.attempts',
        expect.any(Object)
      );
    });

    it('deve gerenciar erro na verificação de telefone', async () => {
      // Dados de requisição
      const request: PhoneVerificationRequest = {
        phoneNumber: '+244912345678',
        userId: 'user-123',
        verificationType: 'otp',
      };

      // Mock para erro
      mockAxiosInstance.post.mockRejectedValueOnce(new Error('Falha na verificação'));

      // Executar e verificar erro
      await expect(connector.startPhoneVerification(request)).rejects.toThrow('Falha na verificação');

      // Verificações
      expect(mockLogger.error).toHaveBeenCalledWith(expect.stringContaining('Erro ao iniciar verificação'));
      expect(mockMetrics.incrementCounter).toHaveBeenCalledWith(
        'mobileMoney.phoneVerification.error',
        expect.any(Object)
      );
    });
  });

  describe('validateOtp', () => {
    it('deve validar OTP com sucesso', async () => {
      // Dados de requisição
      const request: OtpValidationRequest = {
        verificationId: 'verify-123',
        code: '123456',
        userId: 'user-123',
      };

      // Mock para resposta
      const mockResponse = {
        data: {
          valid: true,
          verificationId: 'verify-123',
          maskedPhoneNumber: '+24491****78',
          numberVerified: true,
          verifiedAt: new Date().toISOString(),
        },
      };
      mockAxiosInstance.post.mockResolvedValueOnce(mockResponse);

      // Executar validação
      const result = await connector.validateOtp(request);

      // Verificações
      expect(result).toEqual(mockResponse.data);
      expect(mockAxiosInstance.post).toHaveBeenCalledWith(
        '/v1/phone/verify/validate',
        expect.objectContaining({
          verificationId: request.verificationId,
          code: request.code,
          userId: request.userId,
        }),
        expect.anything()
      );
      expect(mockMetrics.incrementCounter).toHaveBeenCalledWith(
        'mobileMoney.otp.validationResult',
        expect.objectContaining({
          valid: 'true',
        })
      );
    });

    it('deve gerenciar erro na validação de OTP', async () => {
      // Dados de requisição
      const request: OtpValidationRequest = {
        verificationId: 'verify-123',
        code: '123456',
        userId: 'user-123',
      };

      // Mock para erro
      mockAxiosInstance.post.mockRejectedValueOnce(new Error('Código inválido'));

      // Executar e verificar erro
      await expect(connector.validateOtp(request)).rejects.toThrow('Código inválido');

      // Verificações
      expect(mockLogger.error).toHaveBeenCalledWith(expect.stringContaining('Erro ao validar OTP'));
      expect(mockMetrics.incrementCounter).toHaveBeenCalledWith(
        'mobileMoney.otp.validationError',
        expect.any(Object)
      );
    });
  });

  describe('getUserDetails', () => {
    it('deve obter detalhes do usuário com sucesso', async () => {
      // Número de telefone
      const phoneNumber = '+244912345678';

      // Mock para resposta
      const mockResponse = {
        data: {
          mobileMoneyId: 'mm-user-123',
          phoneNumber: '+244912345678',
          fullName: 'João Silva',
          accountStatus: 'active',
          kycLevel: 'medium',
          operator: 'UNITEL',
          country: 'AO',
          availableBalance: {
            amount: 5000,
            currency: 'AOA',
          },
          limits: {
            daily: 100000,
            monthly: 1000000,
            perTransaction: 50000,
            currency: 'AOA',
          },
          verifications: {
            phone: true,
            email: true,
            identity: true,
          },
        },
      };
      mockAxiosInstance.get.mockResolvedValueOnce(mockResponse);

      // Executar obtenção de detalhes
      const result = await connector.getUserDetails(phoneNumber);

      // Verificações
      expect(result).toEqual(mockResponse.data);
      expect(mockAxiosInstance.get).toHaveBeenCalledWith(
        `/v1/users/phone/${phoneNumber}`,
        expect.anything()
      );
    });

    it('deve gerenciar erro na obtenção de detalhes', async () => {
      // Número de telefone
      const phoneNumber = '+244912345678';

      // Mock para erro
      mockAxiosInstance.get.mockRejectedValueOnce(new Error('Usuário não encontrado'));

      // Executar e verificar erro
      await expect(connector.getUserDetails(phoneNumber)).rejects.toThrow('Usuário não encontrado');

      // Verificações
      expect(mockLogger.error).toHaveBeenCalledWith(expect.stringContaining('Erro ao obter detalhes do usuário'));
    });
  });

  describe('validatePhoneNumberExists', () => {
    it('deve validar existência de número com sucesso', async () => {
      // Número de telefone
      const phoneNumber = '+244912345678';
      const operator = 'UNITEL';

      // Mock para resposta
      const mockResponse = {
        data: {
          exists: true,
          valid: true,
          operator: 'UNITEL',
          operatorName: 'Unitel Angola',
          type: 'mobile',
          countryCode: 'AO',
        },
      };
      mockAxiosInstance.get.mockResolvedValueOnce(mockResponse);

      // Executar validação
      const result = await connector.validatePhoneNumberExists(phoneNumber, operator);

      // Verificações
      expect(result).toEqual(mockResponse.data);
      expect(mockAxiosInstance.get).toHaveBeenCalledWith(
        '/v1/phone/validate',
        expect.objectContaining({
          params: expect.objectContaining({
            phone_number: phoneNumber,
            operator: operator,
          }),
        })
      );
      expect(mockMetrics.incrementCounter).toHaveBeenCalledWith(
        'mobileMoney.phoneValidation',
        expect.objectContaining({
          exists: 'true',
          valid: 'true',
        })
      );
    });

    it('deve gerenciar erro na validação de existência de número', async () => {
      // Número de telefone
      const phoneNumber = '+244912345678';

      // Configurar phoneValidation como desabilitado
      connector = new MobileMoneyConnector(
        {
          ...config,
          phoneValidation: {
            enabled: false,
            existenceCheck: false,
            statusCheck: false,
          },
        },
        mockLogger,
        mockMetrics,
        mockTracer
      );

      // Executar e verificar erro
      await expect(connector.validatePhoneNumberExists(phoneNumber)).rejects.toThrow(
        'Validação de número de telefone não está habilitada'
      );

      // Verificações
      expect(mockLogger.error).toHaveBeenCalledWith(expect.stringContaining('Erro ao validar existência'));
    });
  });

  describe('shutdown', () => {
    it('deve encerrar o conector com sucesso', async () => {
      // Executar shutdown
      await connector.shutdown();

      // Verificações
      expect(mockLogger.info).toHaveBeenCalledWith(expect.stringContaining('Encerrando conexão'));
    });
  });
});