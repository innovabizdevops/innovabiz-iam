/**
 * Mock do serviço de Mobile Money para testes de integração
 */

export const mockMobileMoneyService = {
  initiateTransaction: jest.fn(),
  getTransactionById: jest.fn(),
  getTransactionHistory: jest.fn(),
  verifyOTP: jest.fn(),
  cancelTransaction: jest.fn(),
  checkEligibility: jest.fn(),
  registerConsent: jest.fn()
};

// Mock de histórico de transações
mockMobileMoneyService.getTransactionHistory.mockResolvedValue({
  transactions: [
    {
      transactionId: 'tx-123',
      provider: 'mpesa',
      type: 'PAYMENT',
      amount: 1000,
      currency: 'AOA',
      status: 'COMPLETED',
      userId: 'user-123',
      phoneNumber: '+244123456789',
      createdAt: new Date(),
      completedAt: new Date()
    },
    {
      transactionId: 'tx-456',
      provider: 'unitel',
      type: 'TRANSFER',
      amount: 500,
      currency: 'AOA',
      status: 'PENDING',
      userId: 'user-123',
      phoneNumber: '+244987654321',
      createdAt: new Date()
    }
  ],
  totalCount: 2,
  pageInfo: {
    hasNextPage: false,
    hasPreviousPage: false,
    startCursor: 'cursor-1',
    endCursor: 'cursor-2'
  }
});

// Mock de elegibilidade
mockMobileMoneyService.checkEligibility.mockResolvedValue({
  eligible: true,
  services: ['PAYMENT', 'TRANSFER', 'WITHDRAWAL'],
  limits: {
    daily: 10000,
    monthly: 100000,
    transaction: 5000
  },
  kycRequired: false,
  requiresUpgrade: false,
  message: 'Usuário elegível para todos os serviços'
});