/**
 * @file payment-gateway-connector.test.ts
 * @description Testes de unidade para o conector Payment Gateway
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

import { PaymentGatewayConnector, PaymentGatewayConnectorConfig, AuthVerificationResponse, PaymentMethodDetails, TransactionStatus } from '../../integration/connectors/payment-gateway-connector';
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

describe('PaymentGatewayConnector', () => {
  let connector: PaymentGatewayConnector;
  let config: PaymentGatewayConnectorConfig;
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
      baseUrl: 'https://api.paymentgateway.test',
      merchantId: 'test-merchant-id',
      apiVersion: '2.1',
      auth: {
        type: 'apiKey',
        credentials: {
          apiKey: 'test-api-key',
        },
      },
      timeoutMs: 5000,
      transactionAuth: {
        strongAuthThreshold: 1000,
        threeDSecure: {
          enabled: true,
          version: '2.2',
          challengeIndicator: 'challenge-preferred',
        },
        mode: 'sync',
      },
      tokenization: {
        enabled: true,
        tokenType: 'non-pci',
        expirationSeconds: 86400 * 30, // 30 dias
      },
      observability: {
        tracingEnabled: true,
        metricsEnabled: true,
        tags: {
          service: 'payment-gateway',
          environment: 'test',
        },
      },
    };

    // Configurar mock do axios.create
    mockedAxios.create.mockReturnValue(mockAxiosInstance as any);
    
    // Criar instância do conector
    connector = new PaymentGatewayConnector(
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
          version: '2.1.0',
          services: { database: 'UP', processor: 'UP', fraud: 'UP' },
          environment: 'test',
          features: ['3DS', 'tokenization', 'fraud-detection'],
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
            'X-Merchant-ID': config.merchantId,
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
      // Configuração inválida sem merchantId
      const invalidConfig = { ...config, merchantId: '' };
      
      // Criar instância com configuração inválida
      const invalidConnector = new PaymentGatewayConnector(
        invalidConfig as PaymentGatewayConnectorConfig,
        mockLogger,
        mockMetrics,
        mockTracer
      );

      // Deve falhar na inicialização
      await expect(invalidConnector.initialize()).rejects.toThrow('merchantId é obrigatório');
    });
  });

  describe('healthCheck', () => {
    it('deve retornar status CONNECTED quando o serviço está UP', async () => {
      // Mock para health check
      mockAxiosInstance.get.mockResolvedValueOnce({
        data: {
          status: 'UP',
          version: '2.1.0',
          services: { database: 'UP', processor: 'UP', fraud: 'UP' },
          environment: 'test',
          features: ['3DS', 'tokenization', 'fraud-detection'],
        },
      });

      // Executar health check
      const result = await connector.healthCheck();

      // Verificações
      expect(result.status).toBe(ConnectorStatus.CONNECTED);
      expect(result.details).toMatchObject({
        apiVersion: '2.1.0',
        services: expect.any(Object),
        environment: 'test',
        features: expect.any(Array),
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

  describe('verifyTransactionAuthentication', () => {
    it('deve verificar autenticação de transação com sucesso', async () => {
      // Dados de transação
      const transactionData = {
        transactionId: 'txn-123',
        userId: 'user-456',
        amount: 1500,
        currency: 'AOA',
        paymentMethodId: 'pm-789',
        paymentMethodType: 'credit_card',
        merchantData: {
          name: 'Loja Teste',
          mcc: '5411',
          city: 'Luanda',
          country: 'AO',
        },
        ipAddress: '192.168.1.100',
      };

      // Opções de autenticação
      const authOptions = {
        challengePreference: 'challenge-preferred' as const,
        redirectUrl: 'https://app.example.com/auth-callback',
        authenticationMethod: 'otp' as const,
      };

      // Mock para resposta
      const mockResponse = {
        data: {
          status: 'approved',
          transactionId: 'txn-123',
          authenticationId: 'auth-123',
          factorsUsed: ['otp'],
          timestamp: new Date().toISOString(),
        },
      };
      mockAxiosInstance.post.mockResolvedValueOnce(mockResponse);

      // Executar verificação
      const result = await connector.verifyTransactionAuthentication(
        transactionData,
        authOptions
      );

      // Verificações
      expect(result).toEqual(mockResponse.data);
      expect(mockAxiosInstance.post).toHaveBeenCalledWith(
        '/v1/transactions/authentication/verify',
        expect.objectContaining({
          transactionId: transactionData.transactionId,
          userId: transactionData.userId,
          amount: transactionData.amount,
          merchantId: config.merchantId,
          authentication: expect.objectContaining({
            challengePreference: authOptions.challengePreference,
            redirectUrl: authOptions.redirectUrl,
            threeDSecureOptions: config.transactionAuth?.threeDSecure,
          }),
        }),
        expect.anything()
      );
      expect(mockMetrics.incrementCounter).toHaveBeenCalledWith(
        'paymentGateway.verifyAuth.result',
        expect.objectContaining({
          status: 'approved',
        })
      );
    });

    it('deve requerer autenticação forte para valores acima do limite', async () => {
      // Dados de transação com valor alto
      const transactionData = {
        transactionId: 'txn-123',
        userId: 'user-456',
        amount: 5000, // Acima do limite de autenticação forte (1000)
        currency: 'AOA',
        paymentMethodId: 'pm-789',
        paymentMethodType: 'credit_card',
      };

      // Opções de autenticação
      const authOptions = {
        challengePreference: 'no-challenge' as const, // Mesmo com preferência de não-desafio
      };

      // Mock para resposta
      const mockResponse = {
        data: {
          status: 'requires_action', // Requer ação adicional para autenticação forte
          transactionId: 'txn-123',
          authenticationId: 'auth-123',
          redirectUrl: 'https://3ds.example.com/challenge',
          challengeCode: 'chall-123',
          timestamp: new Date().toISOString(),
        },
      };
      mockAxiosInstance.post.mockResolvedValueOnce(mockResponse);

      // Executar verificação
      const result = await connector.verifyTransactionAuthentication(
        transactionData,
        authOptions
      );

      // Verificações
      expect(result).toEqual(mockResponse.data);
      expect(mockAxiosInstance.post).toHaveBeenCalledWith(
        '/v1/transactions/authentication/verify',
        expect.objectContaining({
          authentication: expect.objectContaining({
            requiresStrongAuth: true, // Deve enviar flag de autenticação forte
          }),
        }),
        expect.anything()
      );
    });

    it('deve gerenciar erro na verificação de autenticação', async () => {
      // Dados de transação
      const transactionData = {
        transactionId: 'txn-123',
        userId: 'user-456',
        amount: 1500,
        currency: 'AOA',
        paymentMethodId: 'pm-789',
      };

      // Opções de autenticação
      const authOptions = {};

      // Mock para erro
      mockAxiosInstance.post.mockRejectedValueOnce(new Error('Falha na autenticação'));

      // Executar e verificar erro
      await expect(
        connector.verifyTransactionAuthentication(transactionData, authOptions)
      ).rejects.toThrow('Falha na autenticação');

      // Verificações
      expect(mockLogger.error).toHaveBeenCalledWith(expect.stringContaining('Erro ao verificar autenticação'));
      expect(mockMetrics.incrementCounter).toHaveBeenCalledWith(
        'paymentGateway.verifyAuth.error',
        expect.any(Object)
      );
    });
  });

  describe('getPaymentMethodDetails', () => {
    it('deve obter detalhes do método de pagamento com sucesso', async () => {
      // IDs para consulta
      const userId = 'user-456';
      const paymentMethodId = 'pm-789';

      // Mock para resposta
      const mockResponse = {
        data: {
          id: paymentMethodId,
          type: 'credit_card',
          subtype: 'visa',
          data: {
            last4: '4242',
            brand: 'visa',
            expiryMonth: '12',
            expiryYear: '2025',
            holderName: 'João Silva',
            tokenized: true,
            token: 'tok-123456',
          },
          userId: userId,
          trustLevel: 'high',
          createdAt: '2023-01-01T00:00:00Z',
          updatedAt: '2023-01-01T00:00:00Z',
        },
      };
      mockAxiosInstance.get.mockResolvedValueOnce(mockResponse);

      // Executar consulta
      const result = await connector.getPaymentMethodDetails(userId, paymentMethodId);

      // Verificações
      expect(result).toEqual(mockResponse.data);
      expect(mockAxiosInstance.get).toHaveBeenCalledWith(
        `/v1/users/${userId}/payment-methods/${paymentMethodId}`,
        expect.anything()
      );
    });

    it('deve gerenciar erro na obtenção de detalhes do método de pagamento', async () => {
      // IDs para consulta
      const userId = 'user-456';
      const paymentMethodId = 'pm-789';

      // Mock para erro
      mockAxiosInstance.get.mockRejectedValueOnce(new Error('Método de pagamento não encontrado'));

      // Executar e verificar erro
      await expect(
        connector.getPaymentMethodDetails(userId, paymentMethodId)
      ).rejects.toThrow('Método de pagamento não encontrado');

      // Verificações
      expect(mockLogger.error).toHaveBeenCalledWith(expect.stringContaining('Erro ao obter detalhes do método de pagamento'));
    });
  });

  describe('listUserPaymentMethods', () => {
    it('deve listar métodos de pagamento do usuário com sucesso', async () => {
      // ID do usuário
      const userId = 'user-456';

      // Opções de listagem
      const options = {
        types: ['credit_card', 'debit_card'],
        limit: 5,
        includeTrustLevel: true,
      };

      // Mock para resposta
      const mockResponse = {
        data: {
          paymentMethods: [
            {
              id: 'pm-789',
              type: 'credit_card',
              subtype: 'visa',
              data: {
                last4: '4242',
                brand: 'visa',
                tokenized: true,
              },
              userId: userId,
              trustLevel: 'high',
              createdAt: '2023-01-01T00:00:00Z',
              updatedAt: '2023-01-01T00:00:00Z',
            },
            {
              id: 'pm-790',
              type: 'debit_card',
              subtype: 'mastercard',
              data: {
                last4: '5678',
                brand: 'mastercard',
                tokenized: true,
              },
              userId: userId,
              trustLevel: 'medium',
              createdAt: '2023-02-01T00:00:00Z',
              updatedAt: '2023-02-01T00:00:00Z',
            },
          ],
          total: 2,
          offset: 0,
          limit: 5,
        },
      };
      mockAxiosInstance.get.mockResolvedValueOnce(mockResponse);

      // Executar listagem
      const result = await connector.listUserPaymentMethods(userId, options);

      // Verificações
      expect(result).toEqual(mockResponse.data);
      expect(mockAxiosInstance.get).toHaveBeenCalledWith(
        `/v1/users/${userId}/payment-methods`,
        expect.objectContaining({
          params: expect.objectContaining({
            types: 'credit_card,debit_card',
            limit: 5,
            include_trust_level: true,
          }),
        })
      );
    });

    it('deve gerenciar erro na listagem de métodos de pagamento', async () => {
      // ID do usuário
      const userId = 'user-456';

      // Mock para erro
      mockAxiosInstance.get.mockRejectedValueOnce(new Error('Falha ao listar métodos de pagamento'));

      // Executar e verificar erro
      await expect(
        connector.listUserPaymentMethods(userId)
      ).rejects.toThrow('Falha ao listar métodos de pagamento');

      // Verificações
      expect(mockLogger.error).toHaveBeenCalledWith(expect.stringContaining('Erro ao listar métodos de pagamento'));
    });
  });

  describe('getTransactionStatus', () => {
    it('deve obter status de transação com sucesso', async () => {
      // ID da transação
      const transactionId = 'txn-123';

      // Mock para resposta
      const mockResponse = {
        data: {
          transactionId: transactionId,
          status: 'authorized',
          amount: 1500,
          currency: 'AOA',
          createdAt: '2023-01-01T00:00:00Z',
          updatedAt: '2023-01-01T00:01:00Z',
          authenticationStatus: 'verified',
          authenticationDetails: {
            method: 'otp',
            completed: true,
          },
        },
      };
      mockAxiosInstance.get.mockResolvedValueOnce(mockResponse);

      // Executar consulta
      const result = await connector.getTransactionStatus(transactionId);

      // Verificações
      expect(result).toEqual(mockResponse.data);
      expect(mockAxiosInstance.get).toHaveBeenCalledWith(
        `/v1/transactions/${transactionId}`,
        expect.anything()
      );
    });

    it('deve gerenciar erro na obtenção de status de transação', async () => {
      // ID da transação
      const transactionId = 'txn-123';

      // Mock para erro
      mockAxiosInstance.get.mockRejectedValueOnce(new Error('Transação não encontrada'));

      // Executar e verificar erro
      await expect(
        connector.getTransactionStatus(transactionId)
      ).rejects.toThrow('Transação não encontrada');

      // Verificações
      expect(mockLogger.error).toHaveBeenCalledWith(expect.stringContaining('Erro ao obter status da transação'));
    });
  });

  describe('reportAuthenticationResult', () => {
    it('deve reportar resultado de autenticação com sucesso', async () => {
      // ID da transação
      const transactionId = 'txn-123';

      // Resultado da autenticação
      const authResult = {
        status: 'success' as const,
        authenticationMethod: 'otp',
        factorsUsed: ['otp', 'device'],
        challengeCompleted: true,
        authenticationData: {
          deviceId: 'device-123',
          ipAddress: '192.168.1.100',
        },
      };

      // Mock para resposta
      const mockResponse = {
        data: {
          success: true,
          transactionId: transactionId,
          updatedStatus: 'authorized',
          timestamp: new Date().toISOString(),
        },
      };
      mockAxiosInstance.post.mockResolvedValueOnce(mockResponse);

      // Executar reporte
      const result = await connector.reportAuthenticationResult(transactionId, authResult);

      // Verificações
      expect(result).toEqual(mockResponse.data);
      expect(mockAxiosInstance.post).toHaveBeenCalledWith(
        `/v1/transactions/${transactionId}/authentication/result`,
        expect.objectContaining({
          status: 'success',
          authenticationMethod: 'otp',
          factorsUsed: ['otp', 'device'],
          challengeCompleted: true,
          merchantId: config.merchantId,
        }),
        expect.anything()
      );
      expect(mockMetrics.incrementCounter).toHaveBeenCalledWith(
        'paymentGateway.authResult',
        expect.objectContaining({
          status: 'success',
          challengeCompleted: 'true',
        })
      );
    });

    it('deve gerenciar erro no reporte de resultado de autenticação', async () => {
      // ID da transação
      const transactionId = 'txn-123';

      // Resultado da autenticação
      const authResult = {
        status: 'failure' as const,
        errorCode: 'AUTH_FAILED',
        errorMessage: 'Autenticação falhou',
      };

      // Mock para erro
      mockAxiosInstance.post.mockRejectedValueOnce(new Error('Falha ao reportar resultado'));

      // Executar e verificar erro
      await expect(
        connector.reportAuthenticationResult(transactionId, authResult)
      ).rejects.toThrow('Falha ao reportar resultado');

      // Verificações
      expect(mockLogger.error).toHaveBeenCalledWith(expect.stringContaining('Erro ao reportar resultado de autenticação'));
      expect(mockMetrics.incrementCounter).toHaveBeenCalledWith(
        'paymentGateway.authResult.error',
        expect.objectContaining({
          transactionId,
        })
      );
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
