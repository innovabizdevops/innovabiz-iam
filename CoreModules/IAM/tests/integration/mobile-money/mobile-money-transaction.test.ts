/**
 * Testes de integração para Mobile Money - Transações
 * 
 * Verifica se os resolvers GraphQL e serviços de backend funcionam corretamente juntos
 * Validando conformidade regulatória e análise de risco
 */

import { ApolloServer } from 'apollo-server-express';
import { createTestClient } from 'apollo-server-testing';
import { gql } from 'apollo-server-core';
import { typeDefs } from '../../../src/api/graphql/schema';
import { resolvers } from '../../../src/api/graphql/resolvers';
import { mockMobileMoneyService } from '../mocks/mobile-money-service.mock';
import { mockRiskScoreService } from '../mocks/risk-score-service.mock';
import { mockComplianceService } from '../mocks/compliance-service.mock';

// Mock dos módulos de serviço
jest.mock('../../../src/integration/mobile-money/service', () => mockMobileMoneyService);
jest.mock('../../../src/integration/risk-management/risk-score-service', () => ({
  RiskScoreService: mockRiskScoreService
}));
jest.mock('../../../src/api/graphql/resolvers/mobile-money/compliance-service', () => ({
  ComplianceService: mockComplianceService
}));

// Configuração do servidor de teste
const setupTestServer = (context = {}) => {
  const server = new ApolloServer({
    typeDefs,
    resolvers,
    context: () => context
  });

  return createTestClient(server);
};

describe('Mobile Money Integration Tests', () => {
  // Contexto de autenticação padrão para os testes
  const defaultContext = {
    user: {
      id: 'user-123',
      permissions: ['mobile_money:read', 'mobile_money:write']
    },
    tenantId: 'tenant-123',
    session: {
      sessionId: 'session-123'
    },
    ip: '192.168.1.1'
  };

  beforeEach(() => {
    // Resetar todos os mocks antes de cada teste
    jest.clearAllMocks();
    
    // Configurar comportamento padrão dos mocks
    mockMobileMoneyService.initiateTransaction.mockResolvedValue({
      transactionId: 'tx-123',
      provider: 'mpesa',
      type: 'PAYMENT',
      amount: 1000,
      currency: 'AOA',
      status: 'PENDING',
      userId: 'user-123',
      phoneNumber: '+244123456789',
      createdAt: new Date(),
      updatedAt: new Date()
    });

    mockMobileMoneyService.getTransactionById.mockResolvedValue({
      transactionId: 'tx-123',
      provider: 'mpesa',
      type: 'PAYMENT',
      amount: 1000,
      currency: 'AOA',
      status: 'PENDING',
      userId: 'user-123',
      phoneNumber: '+244123456789',
      createdAt: new Date(),
      updatedAt: new Date()
    });

    mockMobileMoneyService.verifyOTP.mockResolvedValue({
      transactionId: 'tx-123',
      provider: 'mpesa',
      type: 'PAYMENT',
      amount: 1000,
      currency: 'AOA',
      status: 'COMPLETED',
      userId: 'user-123',
      phoneNumber: '+244123456789',
      createdAt: new Date(),
      updatedAt: new Date(),
      completedAt: new Date()
    });

    mockRiskScoreService.prototype.calculateScore.mockResolvedValue({
      score: 30,
      level: 'LOW',
      factors: ['TRUSTED_DEVICE', 'REGULAR_TRANSACTION_PATTERN']
    });

    mockComplianceService.prototype.checkValidConsent.mockResolvedValue(true);
    mockComplianceService.prototype.validateConsentById.mockResolvedValue(true);
  });

  describe('Mutation: initiateMobileMoneyTransaction', () => {
    // Definição da query GraphQL
    const INITIATE_TRANSACTION = gql`
      mutation InitiateTransaction($input: MobileMoneyTransactionInput!) {
        initiateMobileMoneyTransaction(input: $input) {
          success
          message
          errors
          transaction {
            id
            provider
            type
            amount
            currency
            status
            phoneNumber
            riskScore
            riskAssessment {
              score
              level
              requiresReview
            }
          }
        }
      }
    `;

    it('deve iniciar uma transação com sucesso em Angola', async () => {
      // Configurar o cliente de teste
      const { mutate } = setupTestServer(defaultContext);

      // Executar a mutação
      const transactionInput = {
        provider: 'MPESA',
        type: 'PAYMENT',
        amount: 1000,
        currency: 'AOA',
        phoneNumber: '+244123456789',
        userId: 'user-123',
        description: 'Pagamento de serviço',
        regionCode: 'AO',
        tenantId: 'tenant-123',
        complianceData: {
          kycLevel: 'FULL',
          consentId: 'consent-123',
          purposeCode: 'LIVING_EXPENSES'
        }
      };

      const response = await mutate({
        mutation: INITIATE_TRANSACTION,
        variables: { input: transactionInput }
      });

      // Verificar resposta
      expect(response.errors).toBeUndefined();
      expect(response.data.initiateMobileMoneyTransaction.success).toBeTruthy();
      expect(response.data.initiateMobileMoneyTransaction.transaction).toBeDefined();
      expect(response.data.initiateMobileMoneyTransaction.transaction.provider).toBe('MPESA');
      expect(response.data.initiateMobileMoneyTransaction.transaction.amount).toBe(1000);

      // Verificar se o serviço foi chamado corretamente
      expect(mockMobileMoneyService.initiateTransaction).toHaveBeenCalledTimes(1);
      const serviceCall = mockMobileMoneyService.initiateTransaction.mock.calls[0][0];
      expect(serviceCall.provider).toBe('mpesa');
      expect(serviceCall.amount).toBe(1000);
      expect(serviceCall.regionCode).toBe('AO');

      // Verificar se a análise de risco foi aplicada
      expect(mockRiskScoreService.prototype.calculateScore).toHaveBeenCalledTimes(1);
    });

    it('deve rejeitar transação com conformidade inadequada em Angola', async () => {
      // Configurar falha na verificação de consentimento
      mockComplianceService.prototype.checkValidConsent.mockResolvedValue(false);

      // Configurar o cliente de teste
      const { mutate } = setupTestServer(defaultContext);

      // Executar a mutação sem dados de compliance adequados
      const transactionInput = {
        provider: 'MPESA',
        type: 'TRANSFER', // Transferência requer consentimento P2P
        amount: 1000,
        currency: 'AOA',
        phoneNumber: '+244123456789',
        userId: 'user-123',
        description: 'Transferência para família',
        regionCode: 'AO',
        tenantId: 'tenant-123',
        complianceData: {
          kycLevel: 'BASIC',  // Nível KYC inadequado para o valor
          purposeCode: ''  // Falta código de finalidade
        }
      };

      const response = await mutate({
        mutation: INITIATE_TRANSACTION,
        variables: { input: transactionInput }
      });

      // Verificar resposta
      expect(response.data.initiateMobileMoneyTransaction.success).toBeFalsy();
      expect(response.data.initiateMobileMoneyTransaction.transaction).toBeNull();
      expect(response.data.initiateMobileMoneyTransaction.errors.length).toBeGreaterThan(0);
      
      // Verificar que a transação não foi iniciada
      expect(mockMobileMoneyService.initiateTransaction).not.toHaveBeenCalled();
    });
  });

  describe('Query: mobileMoneyTransaction', () => {
    // Definição da query GraphQL
    const GET_TRANSACTION = gql`
      query GetTransaction($id: ID!, $tenantId: ID) {
        mobileMoneyTransaction(id: $id, tenantId: $tenantId) {
          id
          provider
          type
          amount
          currency
          status
          phoneNumber
          createdAt
        }
      }
    `;

    it('deve retornar detalhes de uma transação específica', async () => {
      // Configurar o cliente de teste
      const { query } = setupTestServer(defaultContext);

      // Executar a query
      const response = await query({
        query: GET_TRANSACTION,
        variables: { 
          id: 'tx-123', 
          tenantId: 'tenant-123'
        }
      });

      // Verificar resposta
      expect(response.errors).toBeUndefined();
      expect(response.data.mobileMoneyTransaction).toBeDefined();
      expect(response.data.mobileMoneyTransaction.id).toBe('tx-123');
      expect(response.data.mobileMoneyTransaction.provider).toBe('MPESA');

      // Verificar se o serviço foi chamado corretamente
      expect(mockMobileMoneyService.getTransactionById).toHaveBeenCalledWith('tx-123', 'tenant-123');
    });

    it('deve rejeitar acesso a transação de outro usuário sem permissão', async () => {
      // Configurar serviço para retornar transação de outro usuário
      mockMobileMoneyService.getTransactionById.mockResolvedValue({
        transactionId: 'tx-123',
        userId: 'other-user-456', // Outro usuário
        provider: 'mpesa',
        type: 'PAYMENT',
        amount: 1000
      });

      // Configurar o cliente de teste
      const { query } = setupTestServer(defaultContext);

      // Executar a query
      const response = await query({
        query: GET_TRANSACTION,
        variables: { 
          id: 'tx-123', 
          tenantId: 'tenant-123'
        }
      });

      // Verificar erro de permissão
      expect(response.errors).toBeDefined();
      expect(response.errors[0].message).toContain('permissão');
    });
  });

  describe('Mutation: verifyMobileMoneyOTP', () => {
    // Definição da query GraphQL
    const VERIFY_OTP = gql`
      mutation VerifyOTP($input: VerifyOTPInput!) {
        verifyMobileMoneyOTP(input: $input) {
          success
          message
          transaction {
            id
            status
            completedAt
          }
        }
      }
    `;

    it('deve verificar OTP com sucesso', async () => {
      // Configurar o cliente de teste
      const { mutate } = setupTestServer(defaultContext);

      // Executar a mutação
      const response = await mutate({
        mutation: VERIFY_OTP,
        variables: { 
          input: {
            transactionId: 'tx-123',
            otp: '123456',
            tenantId: 'tenant-123'
          }
        }
      });

      // Verificar resposta
      expect(response.errors).toBeUndefined();
      expect(response.data.verifyMobileMoneyOTP.success).toBeTruthy();
      expect(response.data.verifyMobileMoneyOTP.transaction.status).toBe('COMPLETED');
      expect(response.data.verifyMobileMoneyOTP.transaction.completedAt).toBeDefined();

      // Verificar se o serviço foi chamado corretamente
      expect(mockMobileMoneyService.verifyOTP).toHaveBeenCalledWith(
        'tx-123',
        '123456',
        'tenant-123'
      );
    });
  });
});