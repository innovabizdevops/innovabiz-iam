/**
 * Mock do serviço de conformidade para testes de integração
 * Simula verificações de compliance para diferentes regiões e regulações
 */

export const mockComplianceService = jest.fn().mockImplementation(() => {
  return {
    // Verifica se existe um consentimento válido para o usuário
    checkValidConsent: jest.fn().mockResolvedValue(true),
    
    // Valida consentimento pelo ID
    validateConsentById: jest.fn().mockResolvedValue(true),
    
    // Verifica ID regulatório da África do Sul
    validateSouthAfricanRegulatoryId: jest.fn().mockResolvedValue(true),
    
    // Valida chave PIX (Brasil)
    validatePixKey: jest.fn().mockResolvedValue(true),
    
    // Valida IBAN (Portugal/Europa)
    validateIban: jest.fn().mockResolvedValue(true),
    
    // Verifica conformidade com requisitos AML/CFT
    checkAmlCompliance: jest.fn().mockResolvedValue({
      compliant: true,
      riskScore: 20,
      status: 'APPROVED',
      sanctionsScreening: {
        status: 'CLEAR',
        matchDetails: []
      },
      pefStatus: 'NOT_LISTED',
      watchlistMatches: []
    }),
    
    // Verifica status de documentos para KYC
    checkKycDocumentStatus: jest.fn().mockResolvedValue({
      status: 'VERIFIED',
      documentsVerified: ['ID_CARD', 'PROOF_OF_RESIDENCE'],
      kycLevel: 'MEDIUM',
      verificationDate: new Date(),
      expiryDate: new Date(Date.now() + 86400000 * 365) // 1 ano no futuro
    }),
    
    // Verifica conformidade com GDPR
    checkGdprCompliance: jest.fn().mockResolvedValue({
      compliant: true,
      consentStatus: 'ACTIVE',
      dataProcessingPermissions: ['BASIC_INFO', 'TRANSACTION_HISTORY', 'RISK_ASSESSMENT'],
      dataRetentionCompliant: true
    }),
    
    // Verifica conformidade com LGPD (Brasil)
    checkLgpdCompliance: jest.fn().mockResolvedValue({
      compliant: true,
      consentStatus: 'ACTIVE',
      dataProcessingPermissions: ['DADOS_PESSOAIS', 'HISTORICO_TRANSACIONAL'],
      dataRetentionCompliant: true
    }),
    
    // Verifica conformidade com POPIA (África do Sul)
    checkPopiaCompliance: jest.fn().mockResolvedValue({
      compliant: true,
      consentStatus: 'ACTIVE',
      dataProcessingPermissions: ['PERSONAL_INFO', 'TRANSACTION_HISTORY'],
      dataRetentionCompliant: true
    }),
    
    // Verifica requisitos BNA (Angola)
    checkBnaCompliance: jest.fn().mockResolvedValue({
      compliant: true,
      kyc: {
        status: 'VERIFIED',
        level: 'FULL'
      },
      transactionLimitsCompliant: true,
      amlChecksCompliant: true
    }),
    
    // Verifica requisitos Banco de Moçambique
    checkBancoMocambiqueCompliance: jest.fn().mockResolvedValue({
      compliant: true,
      kyc: {
        status: 'VERIFIED',
        level: 'FULL'
      },
      transactionLimitsCompliant: true,
      amlChecksCompliant: true
    }),
    
    // Verifica se há pendências de compliance
    getPendingComplianceIssues: jest.fn().mockResolvedValue([])
  };
});