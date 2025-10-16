/**
 * Testes de integração para o serviço de processamento de transações Mobile Money
 */

import { MobileMoneyTransactionServiceImpl } from '../../../services/mobile-money/transaction-service';
import { MobileMoneyProvider, TransactionStatus, TransactionType } from '../../../services/mobile-money/types';
import { mockLogger } from '../../mocks/logger.mock';
import { mockMetrics } from '../../mocks/metrics.mock';
import { mockTracer } from '../../mocks/tracer.mock';
import { mockDatabaseClient } from '../../mocks/database-client.mock';
import { mockRegionalComplianceService } from '../../mocks/regional-compliance-service.mock';
import { mockRiskService } from '../../mocks/risk-service.mock';
import { mockConfigService } from '../../mocks/config-service.mock';
import { mockEventBus } from '../../mocks/event-bus.mock';
import { mockCacheService } from '../../mocks/cache-service.mock';
import { mockIamService } from '../../mocks/iam-service.mock';
import { mockProviderFactory } from '../../mocks/mobile-money-provider-factory.mock';

// Mock do provider adapter
const mockProviderAdapter = {
  initiateTransaction: jest.fn(),
  verifyOTP: jest.fn(),
  checkStatus: jest.fn(),
  cancelTransaction: jest.fn(),
  getDetails: jest.fn()
};

// Configurar factory para retornar o adapter mock
mockProviderFactory.getProvider.mockReturnValue(mockProviderAdapter);

describe('Mobile Money Transaction Service Integration Tests', () => {
  let transactionService: MobileMoneyTransactionServiceImpl;

  beforeEach(() => {
    // Resetar todos os mocks
    jest.clearAllMocks();

    // Configurar mock do provider adapter
    mockProviderAdapter.initiateTransaction.mockResolvedValue({
      referenceNumber: 'ref-123',
      providerReference: 'provider-tx-123',
      otpRequired: true,
      otpSent: true,
      otpPhoneNumber: '+244123456789'
    });

    mockProviderAdapter.verifyOTP.mockResolvedValue({
      verified: true,
      status: 'OTP_VERIFIED'
    });

    mockProviderAdapter.checkStatus.mockResolvedValue({
      status: 'COMPLETED',
      completedAt: new Date(),
      providerReference: 'provider-tx-123',
      receiptNumber: 'receipt-123'
    });

    mockProviderAdapter.cancelTransaction.mockResolvedValue({
      cancelled: true,
      status: 'CANCELLED'
    });

    // Instanciar o serviço
    transactionService = new MobileMoneyTransactionServiceImpl(
      mockLogger,
      mockMetrics,
      mockTracer,
      mockDatabaseClient,
      mockRegionalComplianceService,
      mockRiskService,
      mockConfigService,
      mockEventBus,
      mockCacheService,
      mockIamService,
      mockProviderFactory
    );
  });

  describe('initiateTransaction', () => {
    it('deve iniciar uma transação com sucesso', async () => {
      // Configurar mock do serviço de risco
      mockRiskService.evaluateRisk.mockResolvedValue({
        score: 30,
        level: 'LOW',
        factors: ['TRUSTED_DEVICE', 'REGULAR_TRANSACTION_PATTERN'],
        requiresReview: false,
        requiresApproval: false,
        automaticallyDeclined: false
      });

      // Preparar dados de entrada
      const input = {
        userId: 'user-123',
        tenantId: 'tenant-456',
        type: TransactionType.PAYMENT,
        amount: 1000,
        currency: 'AOA',
        provider: MobileMoneyProvider.MPESA,
        phoneNumber: '+244123456789',
        description: 'Pagamento de teste',
        deviceInfo: {
          deviceId: 'device-123',
          ipAddress: '192.168.1.1'
        }
      };

      // Chamar o serviço
      const result = await transactionService.initiateTransaction(input);

      // Verificações
      expect(result).toBeDefined();
      expect(result.transactionId).toBeDefined();
      expect(result.status).toBe(TransactionStatus.INITIATED);
      expect(result.otpRequired).toBe(true);
      expect(result.otpSent).toBe(true);

      // Verificar se provider adapter foi chamado
      expect(mockProviderFactory.getProvider).toHaveBeenCalledWith(MobileMoneyProvider.MPESA);
      expect(mockProviderAdapter.initiateTransaction).toHaveBeenCalled();
      
      // Verificar se serviço de risco foi chamado
      expect(mockRiskService.evaluateRisk).toHaveBeenCalledWith(expect.objectContaining({
        userId: 'user-123',
        tenantId: 'tenant-456',
        transactionAmount: 1000
      }));

      // Verificar se a transação foi registrada no banco de dados
      expect(mockDatabaseClient.transaction.create).toHaveBeenCalledWith(expect.objectContaining({
        userId: 'user-123',
        tenantId: 'tenant-456',
        type: TransactionType.PAYMENT,
        status: TransactionStatus.INITIATED,
        amount: 1000,
        currency: 'AOA',
        provider: MobileMoneyProvider.MPESA
      }));

      // Verificar se evento foi publicado
      expect(mockEventBus.publish).toHaveBeenCalledWith(
        'mobile-money.transaction.initiated',
        expect.objectContaining({
          transactionId: expect.any(String),
          userId: 'user-123',
          tenantId: 'tenant-456',
          status: TransactionStatus.INITIATED
        })
      );
    });

    it('deve rejeitar transação de alto risco', async () => {
      // Configurar mock do serviço de risco para retornar alto risco
      mockRiskService.evaluateRisk.mockResolvedValue({
        score: 90,
        level: 'CRITICAL',
        factors: ['SUSPICIOUS_LOCATION', 'UNUSUAL_AMOUNT', 'NEW_DEVICE'],
        requiresReview: true,
        requiresApproval: true,
        automaticallyDeclined: true
      });

      // Preparar dados de entrada
      const input = {
        userId: 'user-123',
        tenantId: 'tenant-456',
        type: TransactionType.PAYMENT,
        amount: 50000, // Valor alto
        currency: 'AOA',
        provider: MobileMoneyProvider.MPESA,
        phoneNumber: '+244123456789',
        description: 'Pagamento de teste',
        deviceInfo: {
          deviceId: 'new-device',
          ipAddress: '203.0.113.1' // IP suspeito
        }
      };

      // Esperar que a chamada seja rejeitada
      await expect(transactionService.initiateTransaction(input)).rejects.toThrow(
        /Transaction automatically declined due to high risk score/
      );

      // Verificar que o provider não foi chamado
      expect(mockProviderAdapter.initiateTransaction).not.toHaveBeenCalled();
    });
  });

  describe('verifyOTP', () => {
    it('deve verificar OTP com sucesso', async () => {
      // Mock de busca de transação existente
      mockDatabaseClient.transaction.findById.mockResolvedValue({
        id: 'tx-123',
        userId: 'user-123',
        tenantId: 'tenant-456',
        status: TransactionStatus.OTP_SENT,
        provider: MobileMoneyProvider.MPESA,
        phoneNumber: '+244123456789'
      });

      // Preparar dados de entrada
      const input = {
        transactionId: 'tx-123',
        tenantId: 'tenant-456',
        otpCode: '123456',
        deviceInfo: {
          deviceId: 'device-123',
          ipAddress: '192.168.1.1'
        }
      };

      // Chamar o serviço
      const result = await transactionService.verifyOTP(input);

      // Verificações
      expect(result).toBeDefined();
      expect(result.verified).toBe(true);
      expect(result.status).toBe(TransactionStatus.OTP_VERIFIED);

      // Verificar se provider adapter foi chamado
      expect(mockProviderAdapter.verifyOTP).toHaveBeenCalledWith({
        transactionId: 'tx-123',
        otpCode: '123456'
      });

      // Verificar se a transação foi atualizada no banco de dados
      expect(mockDatabaseClient.transaction.update).toHaveBeenCalledWith(
        'tx-123',
        expect.objectContaining({
          status: TransactionStatus.OTP_VERIFIED,
          updatedAt: expect.any(Date)
        })
      );

      // Verificar se evento foi publicado
      expect(mockEventBus.publish).toHaveBeenCalledWith(
        'mobile-money.transaction.otp-verified',
        expect.objectContaining({
          transactionId: 'tx-123',
          status: TransactionStatus.OTP_VERIFIED
        })
      );
    });
  });

  describe('checkTransactionStatus', () => {
    it('deve verificar status de uma transação com sucesso', async () => {
      // Mock de busca de transação existente
      mockDatabaseClient.transaction.findById.mockResolvedValue({
        id: 'tx-123',
        userId: 'user-123',
        tenantId: 'tenant-456',
        status: TransactionStatus.PROCESSING,
        provider: MobileMoneyProvider.MPESA
      });

      // Preparar dados de entrada
      const input = {
        transactionId: 'tx-123',
        tenantId: 'tenant-456',
        includeDetails: true
      };

      // Chamar o serviço
      const result = await transactionService.checkTransactionStatus(input);

      // Verificações
      expect(result).toBeDefined();
      expect(result.status).toBe(TransactionStatus.COMPLETED);
      expect(result.completedAt).toBeDefined();
      expect(result.providerReference).toBe('provider-tx-123');
      expect(result.receiptNumber).toBe('receipt-123');

      // Verificar se provider adapter foi chamado
      expect(mockProviderAdapter.checkStatus).toHaveBeenCalledWith({
        transactionId: 'tx-123'
      });

      // Verificar se a transação foi atualizada no banco de dados se o status mudou
      expect(mockDatabaseClient.transaction.update).toHaveBeenCalledWith(
        'tx-123',
        expect.objectContaining({
          status: TransactionStatus.COMPLETED,
          completedAt: expect.any(Date),
          updatedAt: expect.any(Date)
        })
      );
    });
  });

  describe('cancelTransaction', () => {
    it('deve cancelar uma transação com sucesso', async () => {
      // Mock de busca de transação existente
      mockDatabaseClient.transaction.findById.mockResolvedValue({
        id: 'tx-123',
        userId: 'user-123',
        tenantId: 'tenant-456',
        status: TransactionStatus.INITIATED,
        provider: MobileMoneyProvider.MPESA
      });

      // Preparar dados de entrada
      const input = {
        transactionId: 'tx-123',
        tenantId: 'tenant-456',
        reason: 'Cancelado pelo usuário'
      };

      // Chamar o serviço
      const result = await transactionService.cancelTransaction(input);

      // Verificações
      expect(result).toBeDefined();
      expect(result.cancelled).toBe(true);
      expect(result.status).toBe(TransactionStatus.CANCELLED);

      // Verificar se provider adapter foi chamado
      expect(mockProviderAdapter.cancelTransaction).toHaveBeenCalledWith({
        transactionId: 'tx-123',
        reason: 'Cancelado pelo usuário'
      });

      // Verificar se a transação foi atualizada no banco de dados
      expect(mockDatabaseClient.transaction.update).toHaveBeenCalledWith(
        'tx-123',
        expect.objectContaining({
          status: TransactionStatus.CANCELLED,
          updatedAt: expect.any(Date)
        })
      );

      // Verificar se evento foi publicado
      expect(mockEventBus.publish).toHaveBeenCalledWith(
        'mobile-money.transaction.cancelled',
        expect.objectContaining({
          transactionId: 'tx-123',
          status: TransactionStatus.CANCELLED
        })
      );
    });
  });

  describe('checkEligibility', () => {
    it('deve verificar elegibilidade com sucesso', async () => {
      // Mock dos dados de KYC do usuário
      mockDatabaseClient.userKyc.findByUserId.mockResolvedValue({
        kycLevel: 'FULL',
        verificationDate: new Date(),
        documentsVerified: ['ID_CARD', 'PROOF_OF_ADDRESS']
      });

      // Preparar dados de entrada
      const input = {
        userId: 'user-123',
        tenantId: 'tenant-456',
        phoneNumber: '+244123456789',
        transactionType: TransactionType.PAYMENT,
        currency: 'AOA',
        provider: MobileMoneyProvider.MPESA
      };

      // Mock dos limites de transação
      jest.spyOn(transactionService as any, 'getTransactionLimits').mockResolvedValue({
        singleTransactionLimit: 10000,
        dailyLimit: 25000,
        monthlyLimit: 100000,
        remainingDailyLimit: 20000,
        remainingMonthlyLimit: 80000,
        currency: 'AOA'
      });

      // Chamar o serviço
      const result = await transactionService.checkEligibility(input);

      // Verificações
      expect(result).toBeDefined();
      expect(result.eligible).toBe(true);
      expect(result.services).toContain(TransactionType.PAYMENT);
      expect(result.kycRequired).toBe(false);
      expect(result.limits).toBeDefined();
      expect(result.limits.singleTransactionLimit).toBe(10000);
      expect(result.providers).toContain(MobileMoneyProvider.MPESA);
    });
  });

  describe('getTransactionHistory', () => {
    it('deve retornar histórico de transações', async () => {
      // Mock de busca de transações
      mockDatabaseClient.transaction.findAll.mockResolvedValue({
        transactions: [
          {
            id: 'tx-123',
            userId: 'user-123',
            tenantId: 'tenant-456',
            type: TransactionType.PAYMENT,
            status: TransactionStatus.COMPLETED,
            amount: 1000,
            currency: 'AOA',
            provider: MobileMoneyProvider.MPESA,
            phoneNumber: '+244123456789',
            createdAt: new Date(),
            completedAt: new Date()
          },
          {
            id: 'tx-456',
            userId: 'user-123',
            tenantId: 'tenant-456',
            type: TransactionType.TRANSFER,
            status: TransactionStatus.FAILED,
            amount: 500,
            currency: 'AOA',
            provider: MobileMoneyProvider.AIRTEL,
            phoneNumber: '+244123456789',
            createdAt: new Date(),
            failureReason: 'INSUFFICIENT_FUNDS'
          }
        ],
        totalCount: 2,
        hasMore: false
      });

      // Preparar dados de entrada
      const input = {
        userId: 'user-123',
        tenantId: 'tenant-456',
        limit: 10,
        offset: 0
      };

      // Chamar o serviço
      const result = await transactionService.getTransactionHistory(input);

      // Verificações
      expect(result).toBeDefined();
      expect(result.transactions).toHaveLength(2);
      expect(result.totalCount).toBe(2);
      expect(result.hasMore).toBe(false);
    });
  });
});