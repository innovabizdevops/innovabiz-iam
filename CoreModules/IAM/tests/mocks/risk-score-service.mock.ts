/**
 * Mock do serviço de avaliação de risco para testes de integração
 */

export const mockRiskScoreService = jest.fn().mockImplementation(() => {
  return {
    calculateScore: jest.fn().mockResolvedValue({
      score: 30,
      level: 'LOW',
      factors: ['TRUSTED_DEVICE', 'REGULAR_TRANSACTION_PATTERN'],
      requiresReview: false,
      requiresApproval: false,
      automaticallyDeclined: false
    }),
    
    calculateHighRiskScore: jest.fn().mockResolvedValue({
      score: 80,
      level: 'HIGH',
      factors: ['UNUSUAL_LOCATION', 'LARGE_AMOUNT', 'SUSPICIOUS_PATTERN'],
      requiresReview: true,
      requiresApproval: true,
      automaticallyDeclined: false
    }),
    
    calculateFraudRiskScore: jest.fn().mockResolvedValue({
      score: 95,
      level: 'CRITICAL',
      factors: ['FRAUD_PATTERN', 'BLACKLISTED_DEVICE', 'SUSPICIOUS_LOCATION'],
      requiresReview: true,
      requiresApproval: true,
      automaticallyDeclined: true
    }),
    
    getHistoricalRiskScores: jest.fn().mockResolvedValue([
      {
        date: new Date(Date.now() - 86400000 * 30), // 30 dias atrás
        score: 25,
        level: 'LOW'
      },
      {
        date: new Date(Date.now() - 86400000 * 15), // 15 dias atrás
        score: 40,
        level: 'MEDIUM'
      },
      {
        date: new Date(),
        score: 30,
        level: 'LOW'
      }
    ])
  };
});