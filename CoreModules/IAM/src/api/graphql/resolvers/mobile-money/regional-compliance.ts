/**
 * Serviço de resolução de conformidade regional para transações Mobile Money
 * 
 * Implementa verificações específicas de cada mercado/jurisdição:
 * - PALOP (Angola, Moçambique, Guiné-Bissau, Cabo Verde, São Tomé e Príncipe)
 * - SADC (África do Sul, Zimbabwe, Namíbia, Botswana, etc.)
 * - CPLP (Brasil, Portugal)
 * - Outros mercados globais
 */

import { logger } from '../../../../observability/logging';
import { metrics } from '../../../../observability/metrics';
import { tracer } from '../../../../observability/tracing';
import { ComplianceService } from './compliance-service';

// Tipos de verificação de conformidade
type ComplianceResult = {
  compliant: boolean;
  message?: string;
  details?: any;
};

type RegionalRules = {
  [regionCode: string]: {
    requiredFields: string[];
    transactionLimits: {
      [transactionType: string]: {
        [kycLevel: string]: number;
      };
    };
    kycLevels: {
      [level: string]: {
        requiredDocuments: string[];
        dailyLimit: number;
        monthlyLimit: number;
        transactionLimit: number;
      };
    };
    purposeCodes: string[];
    regulations: string[];
    consentTypes: string[];
    documentationUrl?: string;
  };
};

// Regras de conformidade por região
const regionalRules: RegionalRules = {
  // Angola
  'AO': {
    requiredFields: ['complianceData.kycLevel', 'complianceData.consentId'],
    transactionLimits: {
      'TRANSFER': {
        'BASIC': 10000,
        'MEDIUM': 50000,
        'FULL': 250000
      },
      'WITHDRAWAL': {
        'BASIC': 5000,
        'MEDIUM': 25000,
        'FULL': 100000
      }
    },
    kycLevels: {
      'BASIC': {
        requiredDocuments: ['ID_CARD'],
        dailyLimit: 10000,
        monthlyLimit: 100000,
        transactionLimit: 10000
      },
      'MEDIUM': {
        requiredDocuments: ['ID_CARD', 'PROOF_OF_RESIDENCE'],
        dailyLimit: 50000,
        monthlyLimit: 500000,
        transactionLimit: 50000
      },
      'FULL': {
        requiredDocuments: ['ID_CARD', 'PROOF_OF_RESIDENCE', 'TAX_NUMBER'],
        dailyLimit: 250000,
        monthlyLimit: 1500000,
        transactionLimit: 250000
      }
    },
    purposeCodes: ['LIVING_EXPENSES', 'UTILITIES', 'EDUCATION', 'MEDICAL', 'FAMILY_SUPPORT'],
    regulations: [
      'Aviso nº 12/2016 do BNA', 
      'Lei 22/11 de Combate ao Branqueamento de Capitais',
      'Instrução 24/2020 do BNA sobre KYC Simplificado'
    ],
    consentTypes: ['DATA_PROCESSING', 'P2P_TRANSFER', 'BILL_PAYMENT'],
    documentationUrl: 'https://www.bna.ao/Servicos/pesquisa_legislacao.aspx'
  },
  
  // Moçambique
  'MZ': {
    requiredFields: ['complianceData.kycLevel', 'complianceData.purposeCode'],
    transactionLimits: {
      'TRANSFER': {
        'BASIC': 5000,
        'MEDIUM': 25000,
        'FULL': 100000
      },
      'WITHDRAWAL': {
        'BASIC': 2500,
        'MEDIUM': 15000,
        'FULL': 50000
      }
    },
    kycLevels: {
      'BASIC': {
        requiredDocuments: ['ID_CARD'],
        dailyLimit: 5000,
        monthlyLimit: 50000,
        transactionLimit: 5000
      },
      'MEDIUM': {
        requiredDocuments: ['ID_CARD', 'PROOF_OF_RESIDENCE'],
        dailyLimit: 25000,
        monthlyLimit: 250000,
        transactionLimit: 25000
      },
      'FULL': {
        requiredDocuments: ['ID_CARD', 'PROOF_OF_RESIDENCE', 'TAX_NUMBER'],
        dailyLimit: 100000,
        monthlyLimit: 1000000,
        transactionLimit: 100000
      }
    },
    purposeCodes: ['LIVING_EXPENSES', 'UTILITIES', 'EDUCATION', 'BUSINESS', 'FAMILY_SUPPORT'],
    regulations: [
      'Aviso 10/GBM/2016 do Banco de Moçambique',
      'Lei 14/2013 de Prevenção e Combate ao Branqueamento de Capitais'
    ],
    consentTypes: ['DATA_PROCESSING', 'P2P_TRANSFER', 'MERCHANT_PAYMENT'],
    documentationUrl: 'https://www.bancomoc.mz/fm_pgTab1.aspx?id=1'
  },
  
  // África do Sul
  'ZA': {
    requiredFields: ['complianceData.kycLevel', 'complianceData.regulatoryId'],
    transactionLimits: {
      'TRANSFER': {
        'BASIC': 3000,
        'MEDIUM': 15000,
        'FULL': 75000
      },
      'WITHDRAWAL': {
        'BASIC': 1500,
        'MEDIUM': 7500,
        'FULL': 25000
      }
    },
    kycLevels: {
      'BASIC': {
        requiredDocuments: ['ID_DOCUMENT'],
        dailyLimit: 3000,
        monthlyLimit: 25000,
        transactionLimit: 3000
      },
      'MEDIUM': {
        requiredDocuments: ['ID_DOCUMENT', 'PROOF_OF_RESIDENCE'],
        dailyLimit: 15000,
        monthlyLimit: 100000,
        transactionLimit: 15000
      },
      'FULL': {
        requiredDocuments: ['ID_DOCUMENT', 'PROOF_OF_RESIDENCE', 'TAX_NUMBER'],
        dailyLimit: 75000,
        monthlyLimit: 500000,
        transactionLimit: 75000
      }
    },
    purposeCodes: ['LIVING_EXPENSES', 'UTILITIES', 'EDUCATION', 'BUSINESS', 'REAL_ESTATE'],
    regulations: [
      'POPIA (Protection of Personal Information Act)',
      'FICA (Financial Intelligence Centre Act)',
      'National Payment System Act 78 of 1998',
      'Electronic Communications and Transactions Act 25 of 2002'
    ],
    consentTypes: ['POPIA_DATA_PROCESSING', 'FICA_VERIFICATION', 'CREDIT_CHECK'],
    documentationUrl: 'https://www.resbank.co.za/en/home/what-we-do/payments-and-settlements'
  },
  
  // Brasil
  'BR': {
    requiredFields: ['complianceData.kycLevel', 'complianceData.consentId'],
    transactionLimits: {
      'TRANSFER': {
        'BASIC': 1000,
        'MEDIUM': 5000,
        'FULL': 30000
      },
      'WITHDRAWAL': {
        'BASIC': 500,
        'MEDIUM': 2500,
        'FULL': 10000
      }
    },
    kycLevels: {
      'BASIC': {
        requiredDocuments: ['CPF'],
        dailyLimit: 1000,
        monthlyLimit: 5000,
        transactionLimit: 1000
      },
      'MEDIUM': {
        requiredDocuments: ['CPF', 'COMPROVANTE_RESIDENCIA'],
        dailyLimit: 5000,
        monthlyLimit: 30000,
        transactionLimit: 5000
      },
      'FULL': {
        requiredDocuments: ['CPF', 'COMPROVANTE_RESIDENCIA', 'SELFIE_COM_DOCUMENTO'],
        dailyLimit: 30000,
        monthlyLimit: 300000,
        transactionLimit: 30000
      }
    },
    purposeCodes: ['DESPESAS_GERAIS', 'CONTAS', 'EDUCACAO', 'NEGOCIOS', 'IMOVEIS'],
    regulations: [
      'LGPD (Lei Geral de Proteção de Dados)',
      'Resolução BCB Nº 1 de 2020 (PIX)',
      'Circular BACEN 3.680/2013 (Contas de Pagamento)',
      'Circular BACEN 4.032/2020'
    ],
    consentTypes: ['LGPD_DATA_PROCESSING', 'PIX_TRANSFER', 'OPEN_BANKING'],
    documentationUrl: 'https://www.bcb.gov.br/estabilidadefinanceira/pagamentosmoveis'
  },
  
  // Portugal (para suporte à CPLP)
  'PT': {
    requiredFields: ['complianceData.kycLevel', 'complianceData.consentId'],
    transactionLimits: {
      'TRANSFER': {
        'BASIC': 1000,
        'MEDIUM': 5000,
        'FULL': 25000
      },
      'WITHDRAWAL': {
        'BASIC': 500,
        'MEDIUM': 2500,
        'FULL': 10000
      }
    },
    kycLevels: {
      'BASIC': {
        requiredDocuments: ['CARTAO_CIDADAO'],
        dailyLimit: 1000,
        monthlyLimit: 10000,
        transactionLimit: 1000
      },
      'MEDIUM': {
        requiredDocuments: ['CARTAO_CIDADAO', 'COMPROVATIVO_MORADA'],
        dailyLimit: 5000,
        monthlyLimit: 50000,
        transactionLimit: 5000
      },
      'FULL': {
        requiredDocuments: ['CARTAO_CIDADAO', 'COMPROVATIVO_MORADA', 'COMPROVATIVO_RENDIMENTOS'],
        dailyLimit: 25000,
        monthlyLimit: 100000,
        transactionLimit: 25000
      }
    },
    purposeCodes: ['DESPESAS_GERAIS', 'SERVICOS_PUBLICOS', 'EDUCACAO', 'NEGOCIOS', 'HABITACAO'],
    regulations: [
      'RGPD (Regulamento Geral de Proteção de Dados)',
      'Lei nº 83/2017 - Prevenção do branqueamento de capitais',
      'Regulamento BCE nº 260/2012 - SEPA',
      'Decreto-Lei nº 91/2018 - Serviços de pagamento'
    ],
    consentTypes: ['RGPD_DATA_PROCESSING', 'SEPA_TRANSFER', 'OPEN_BANKING'],
    documentationUrl: 'https://www.bportugal.pt/page/sistemas-de-pagamentos'
  }
};

/**
 * Verifica se a transação está em conformidade com as regras regionais
 */
export const resolveRegionalCompliance = async (
  transaction: any, 
  context: any
): Promise<ComplianceResult> => {
  const span = tracer.startSpan('services.resolveRegionalCompliance');
  
  try {
    const { regionCode } = transaction;
    const tenantId = transaction.tenantId || context.tenantId;
    
    logger.info('Verificando conformidade regional para transação', {
      regionCode,
      transactionType: transaction.type,
      tenantId
    });

    metrics.increment('compliance.regional_check', {
      regionCode,
      tenantId
    });

    // Se não houver regras específicas para a região, usar regras padrão
    if (!regionalRules[regionCode]) {
      logger.info(`Regras específicas não encontradas para região ${regionCode}, usando padrão global`);
      return validateGlobalCompliance(transaction, context);
    }

    // Verificar campos obrigatórios para a região
    const requiredFields = regionalRules[regionCode].requiredFields;
    const missingFields = requiredFields.filter(field => {
      const fieldPath = field.split('.');
      let value = transaction;
      
      for (const segment of fieldPath) {
        if (!value || !value[segment]) {
          return true;
        }
        value = value[segment];
      }
      
      return false;
    });

    if (missingFields.length > 0) {
      return {
        compliant: false,
        message: `Campos obrigatórios ausentes para a região ${regionCode}: ${missingFields.join(', ')}`,
        details: {
          missingFields,
          regulations: regionalRules[regionCode].regulations
        }
      };
    }

    // Verificar limites de transação específicos da região
    const kycLevel = transaction.complianceData?.kycLevel || 'BASIC';
    const transactionType = transaction.type;
    const regionRules = regionalRules[regionCode];
    
    if (
      regionRules.transactionLimits[transactionType] && 
      regionRules.transactionLimits[transactionType][kycLevel] && 
      transaction.amount > regionRules.transactionLimits[transactionType][kycLevel]
    ) {
      return {
        compliant: false,
        message: `Valor da transação excede o limite de ${regionRules.transactionLimits[transactionType][kycLevel]} para o nível KYC ${kycLevel} na região ${regionCode}`,
        details: {
          maxLimit: regionRules.transactionLimits[transactionType][kycLevel],
          requestedAmount: transaction.amount,
          kycLevel,
          regulations: regionRules.regulations
        }
      };
    }

    // Verificar requisitos específicos de documentos para o nível KYC
    const requiredDocuments = regionRules.kycLevels[kycLevel]?.requiredDocuments || [];
    
    // Verificar regras específicas por região
    let regionResult;
    
    switch (regionCode) {
      case 'AO':
        regionResult = await validateAngolaCompliance(transaction, context);
        break;
        
      case 'MZ':
        regionResult = await validateMozambiqueCompliance(transaction, context);
        break;
        
      case 'ZA':
        regionResult = await validateSouthAfricaCompliance(transaction, context);
        break;
        
      case 'BR':
        regionResult = await validateBrazilCompliance(transaction, context);
        break;
        
      case 'PT':
        regionResult = await validatePortugalCompliance(transaction, context);
        break;
        
      default:
        regionResult = { compliant: true };
    }
    
    if (!regionResult.compliant) {
      return regionResult;
    }

    // Se passou por todas as verificações
    return {
      compliant: true
    };
  } catch (error) {
    logger.error('Erro ao verificar conformidade regional', { 
      error: error.message,
      regionCode: transaction.regionCode,
      stack: error.stack
    });

    metrics.increment('compliance.regional_check.error', {
      error: error.name
    });

    return {
      compliant: false,
      message: `Erro ao verificar conformidade: ${error.message}`,
      details: {
        error: error.message
      }
    };
  } finally {
    span.end();
  }
};

/**
 * Validações globais aplicáveis a todas as regiões
 */
const validateGlobalCompliance = (transaction: any, context: any): ComplianceResult => {
  // Verificações básicas globais
  if (!transaction.complianceData) {
    return {
      compliant: false,
      message: 'Dados de conformidade são obrigatórios para todas as transações financeiras',
      details: {
        regulation: 'Global AML/CFT Standards'
      }
    };
  }

  // Verificar nível KYC
  if (!transaction.complianceData.kycLevel) {
    return {
      compliant: false,
      message: 'Nível KYC é obrigatório para todas as transações financeiras',
      details: {
        regulation: 'Global AML/CFT Standards - FATF Recommendations'
      }
    };
  }

  // Transações de alto valor exigem KYC avançado
  if (transaction.amount > 10000 && transaction.complianceData.kycLevel === 'BASIC') {
    return {
      compliant: false,
      message: 'Transações de alto valor requerem nível KYC mais elevado',
      details: {
        currentLevel: 'BASIC',
        requiredLevel: 'MEDIUM ou FULL',
        regulation: 'Global AML/CFT Standards - FATF Recommendations'
      }
    };
  }

  return { compliant: true };
};

/**
 * Validações específicas para Angola
 */
const validateAngolaCompliance = async (
  transaction: any,
  context: any
): Promise<ComplianceResult> => {
  const complianceService = new ComplianceService(transaction.tenantId || context.tenantId);
  
  // Verificar consentimento para transferências P2P (Lei 22/11)
  if (transaction.type === 'TRANSFER') {
    const hasValidConsent = await complianceService.checkValidConsent(
      transaction.userId, 
      'P2P_TRANSFER'
    );
    
    if (!hasValidConsent) {
      return {
        compliant: false,
        message: 'Consentimento válido para transferências é obrigatório em Angola',
        details: {
          regulation: 'Lei 22/11 de Combate ao Branqueamento de Capitais e Aviso nº 12/2016 do BNA',
          requiredConsent: 'P2P_TRANSFER'
        }
      };
    }
  }

  // Verificar necessidade de código de finalidade para transações maiores
  if (transaction.amount > 5000 && (!transaction.complianceData.purposeCode || transaction.complianceData.purposeCode === '')) {
    return {
      compliant: false,
      message: 'Código de finalidade é obrigatório para transações acima de 5.000 AOA em Angola',
      details: {
        regulation: 'Aviso nº 12/2016 do BNA'
      }
    };
  }

  return { compliant: true };
};

/**
 * Validações específicas para Moçambique
 */
const validateMozambiqueCompliance = async (
  transaction: any,
  context: any
): Promise<ComplianceResult> => {
  const complianceService = new ComplianceService(transaction.tenantId || context.tenantId);
  
  // Para transferências, verificar existência de consentimento
  if (transaction.type === 'TRANSFER') {
    if (!transaction.complianceData.consentId) {
      return {
        compliant: false,
        message: 'ID de consentimento é obrigatório para transferências em Moçambique',
        details: {
          regulation: 'Aviso 10/GBM/2016 do Banco de Moçambique'
        }
      };
    }
    
    const consentValid = await complianceService.validateConsentById(
      transaction.complianceData.consentId
    );
    
    if (!consentValid) {
      return {
        compliant: false,
        message: 'Consentimento inválido ou expirado',
        details: {
          consentId: transaction.complianceData.consentId,
          regulation: 'Aviso 10/GBM/2016 do Banco de Moçambique'
        }
      };
    }
  }

  // Verificar código de finalidade para saques
  if (transaction.type === 'WITHDRAWAL' && 
      transaction.amount > 2500 && 
      (!transaction.complianceData.purposeCode || transaction.complianceData.purposeCode === '')) {
    return {
      compliant: false,
      message: 'Código de finalidade é obrigatório para saques acima de 2.500 MZN',
      details: {
        regulation: 'Lei 14/2013 de Prevenção e Combate ao Branqueamento de Capitais'
      }
    };
  }

  return { compliant: true };
};

/**
 * Validações específicas para África do Sul
 */
const validateSouthAfricaCompliance = async (
  transaction: any,
  context: any
): Promise<ComplianceResult> => {
  const complianceService = new ComplianceService(transaction.tenantId || context.tenantId);
  
  // Verificar ID regulatório conforme POPIA e FICA
  if (!transaction.complianceData.regulatoryId) {
    return {
      compliant: false,
      message: 'ID regulatório é obrigatório para transações na África do Sul',
      details: {
        regulation: 'FICA (Financial Intelligence Centre Act) e POPIA'
      }
    };
  }
  
  // Verificar se o ID regulatório é válido
  const isValidRegulatoryId = await complianceService.validateSouthAfricanRegulatoryId(
    transaction.complianceData.regulatoryId
  );
  
  if (!isValidRegulatoryId) {
    return {
      compliant: false,
      message: 'ID regulatório inválido',
      details: {
        regulation: 'FICA (Financial Intelligence Centre Act) e POPIA'
      }
    };
  }
  
  // Verificar consentimento POPIA para processamento de dados pessoais
  const hasPopiaConsent = await complianceService.checkValidConsent(
    transaction.userId, 
    'POPIA_DATA_PROCESSING'
  );
  
  if (!hasPopiaConsent) {
    return {
      compliant: false,
      message: 'Consentimento POPIA para processamento de dados é obrigatório',
      details: {
        regulation: 'POPIA (Protection of Personal Information Act)',
        requiredConsent: 'POPIA_DATA_PROCESSING'
      }
    };
  }

  return { compliant: true };
};

/**
 * Validações específicas para Brasil
 */
const validateBrazilCompliance = async (
  transaction: any,
  context: any
): Promise<ComplianceResult> => {
  const complianceService = new ComplianceService(transaction.tenantId || context.tenantId);
  
  // Verificar consentimento LGPD para processamento de dados pessoais
  const hasLgpdConsent = await complianceService.checkValidConsent(
    transaction.userId, 
    'LGPD_DATA_PROCESSING'
  );
  
  if (!hasLgpdConsent) {
    return {
      compliant: false,
      message: 'Consentimento LGPD para processamento de dados é obrigatório',
      details: {
        regulation: 'LGPD (Lei Geral de Proteção de Dados)',
        requiredConsent: 'LGPD_DATA_PROCESSING'
      }
    };
  }
  
  // Para Pix (transferências instantâneas), validações adicionais
  if (transaction.type === 'TRANSFER' && transaction.metadata?.transferType === 'PIX') {
    if (!transaction.metadata.pixKey) {
      return {
        compliant: false,
        message: 'Chave Pix é obrigatória para transferências Pix',
        details: {
          regulation: 'Resolução BCB Nº 1 de 2020'
        }
      };
    }
    
    const isValidPixKey = await complianceService.validatePixKey(
      transaction.metadata.pixKey
    );
    
    if (!isValidPixKey) {
      return {
        compliant: false,
        message: 'Chave Pix inválida',
        details: {
          regulation: 'Resolução BCB Nº 1 de 2020'
        }
      };
    }
  }

  return { compliant: true };
};

/**
 * Validações específicas para Portugal (CPLP)
 */
const validatePortugalCompliance = async (
  transaction: any,
  context: any
): Promise<ComplianceResult> => {
  const complianceService = new ComplianceService(transaction.tenantId || context.tenantId);
  
  // Verificar consentimento RGPD para processamento de dados pessoais
  const hasRgpdConsent = await complianceService.checkValidConsent(
    transaction.userId, 
    'RGPD_DATA_PROCESSING'
  );
  
  if (!hasRgpdConsent) {
    return {
      compliant: false,
      message: 'Consentimento RGPD para processamento de dados é obrigatório',
      details: {
        regulation: 'RGPD (Regulamento Geral de Proteção de Dados)',
        requiredConsent: 'RGPD_DATA_PROCESSING'
      }
    };
  }
  
  // Para transferências SEPA, validações adicionais
  if (transaction.type === 'TRANSFER' && transaction.metadata?.transferType === 'SEPA') {
    if (!transaction.metadata.ibanDestination) {
      return {
        compliant: false,
        message: 'IBAN de destino é obrigatório para transferências SEPA',
        details: {
          regulation: 'Regulamento BCE nº 260/2012 - SEPA'
        }
      };
    }
    
    const isValidIban = await complianceService.validateIban(
      transaction.metadata.ibanDestination
    );
    
    if (!isValidIban) {
      return {
        compliant: false,
        message: 'IBAN inválido',
        details: {
          regulation: 'Regulamento BCE nº 260/2012 - SEPA'
        }
      };
    }
  }

  return { compliant: true };
};