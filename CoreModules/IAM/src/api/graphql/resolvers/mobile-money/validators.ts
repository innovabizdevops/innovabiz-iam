/**
 * Validadores para operações de Mobile Money
 * Implementa verificações específicas de cada região/mercado
 * Garante conformidade com regulações bancárias e financeiras
 */

import { UserInputError } from 'apollo-server-express';
import { LimitsRepository } from './limits-repository';
import { PhoneNumberValidator } from '../../../../utils/phone-validator';
import { logger } from '../../../../observability/logging';
import { ComplianceService } from './compliance-service';

/**
 * Validações de requisitos gerais
 */
export const validateTransactionRequest = async (input: any, tenantId: string): Promise<void> => {
  // Validar campos básicos obrigatórios
  if (!input.provider) {
    throw new UserInputError('Provedor é obrigatório');
  }

  if (!input.type) {
    throw new UserInputError('Tipo de transação é obrigatório');
  }

  if (!input.amount || input.amount <= 0) {
    throw new UserInputError('Valor da transação deve ser maior que zero');
  }

  if (!input.currency) {
    throw new UserInputError('Moeda é obrigatória');
  }

  if (!input.phoneNumber) {
    throw new UserInputError('Número de telefone é obrigatório');
  }

  if (!input.userId) {
    throw new UserInputError('ID do usuário é obrigatório');
  }

  if (!input.regionCode) {
    throw new UserInputError('Código da região é obrigatório');
  }

  // Validar formato do número de telefone
  const phoneValidator = new PhoneNumberValidator();
  if (!phoneValidator.isValid(input.phoneNumber, input.regionCode)) {
    throw new UserInputError(`Número de telefone inválido para a região ${input.regionCode}`);
  }

  // Validar limites de transação
  await validateTransactionLimits(input, tenantId);

  // Validar requisitos regulatórios específicos por região
  await validateRegionalRequirements(input, tenantId);

  logger.info('Validação de transação concluída com sucesso', {
    provider: input.provider,
    type: input.type,
    regionCode: input.regionCode,
    tenantId
  });
};

/**
 * Validação de limites de transação
 */
const validateTransactionLimits = async (input: any, tenantId: string): Promise<void> => {
  const limitsRepo = new LimitsRepository(tenantId);
  
  // Obter limites aplicáveis
  const kycLevel = input.complianceData?.kycLevel || 'BASIC';
  const limits = await limitsRepo.getLimits({
    provider: input.provider,
    regionCode: input.regionCode,
    kycLevel,
    transactionType: input.type
  });

  // Validar valor máximo por transação
  if (limits.transactionMax && input.amount > limits.transactionMax) {
    throw new UserInputError(
      `O valor da transação excede o limite máximo de ${limits.transactionMax} ${input.currency} para o nível KYC ${kycLevel}`
    );
  }

  // Validar limites diários
  const dailyTotal = await limitsRepo.getDailyTotal(input.userId, input.regionCode);
  if (limits.dailyMax && (dailyTotal + input.amount) > limits.dailyMax) {
    throw new UserInputError(
      `Esta transação excederia o limite diário de ${limits.dailyMax} ${input.currency}. ` +
      `Total atual: ${dailyTotal} ${input.currency}`
    );
  }

  // Validar limites mensais
  const monthlyTotal = await limitsRepo.getMonthlyTotal(input.userId, input.regionCode);
  if (limits.monthlyMax && (monthlyTotal + input.amount) > limits.monthlyMax) {
    throw new UserInputError(
      `Esta transação excederia o limite mensal de ${limits.monthlyMax} ${input.currency}. ` +
      `Total atual: ${monthlyTotal} ${input.currency}`
    );
  }

  // Validar número máximo de transações por dia
  const dailyCount = await limitsRepo.getDailyTransactionCount(input.userId, input.regionCode);
  if (limits.dailyTransactionCount && dailyCount >= limits.dailyTransactionCount) {
    throw new UserInputError(
      `Número máximo de transações diárias (${limits.dailyTransactionCount}) atingido para o usuário`
    );
  }

  logger.info('Validação de limites concluída com sucesso', {
    amount: input.amount,
    currency: input.currency,
    kycLevel,
    dailyTotal,
    monthlyTotal,
    dailyCount,
    tenantId
  });
};

/**
 * Validações específicas por região
 */
const validateRegionalRequirements = async (input: any, tenantId: string): Promise<void> => {
  const complianceService = new ComplianceService(tenantId);
  
  switch (input.regionCode) {
    case 'AO': // Angola
      await validateAngolanRequirements(input, complianceService);
      break;
      
    case 'MZ': // Moçambique
      await validateMozambiqueRequirements(input, complianceService);
      break;
      
    case 'ZA': // África do Sul
      await validateSouthAfricanRequirements(input, complianceService);
      break;
      
    case 'BR': // Brasil
      await validateBrazilianRequirements(input, complianceService);
      break;
      
    // Adicionar outros países conforme necessário
    
    default:
      // Validações gerais para regiões não explicitamente especificadas
      if (!input.complianceData || !input.complianceData.kycLevel) {
        throw new UserInputError('Nível KYC é obrigatório para transações financeiras');
      }
  }
};

/**
 * Validações específicas para Angola (BNA)
 */
const validateAngolanRequirements = async (input: any, complianceService: ComplianceService): Promise<void> => {
  // Verificar dados de compliance
  if (!input.complianceData) {
    throw new UserInputError('Dados de compliance são obrigatórios para transações em Angola');
  }

  // Verificar nível KYC
  if (!input.complianceData.kycLevel) {
    throw new UserInputError('Nível KYC é obrigatório para transações em Angola');
  }

  // Transações acima de certos valores exigem nível KYC específico
  if (input.amount > 100000 && input.complianceData.kycLevel !== 'FULL') {
    throw new UserInputError('Transações acima de 100.000 AOA requerem KYC completo conforme Aviso 12/2016 do BNA');
  }

  // Verificar se existe consentimento válido para o usuário
  if (input.type === 'TRANSFER') {
    const hasValidConsent = await complianceService.checkValidConsent(input.userId, 'P2P_TRANSFER');
    if (!hasValidConsent) {
      throw new UserInputError('Consentimento válido para transferências é obrigatório conforme Lei 22/11 de Angola');
    }
  }
};

/**
 * Validações específicas para Moçambique (Banco de Moçambique)
 */
const validateMozambiqueRequirements = async (input: any, complianceService: ComplianceService): Promise<void> => {
  // Verificar dados de compliance
  if (!input.complianceData) {
    throw new UserInputError('Dados de compliance são obrigatórios para transações em Moçambique');
  }

  // Para transferências, verificar existência de consentimento
  if (input.type === 'TRANSFER') {
    if (!input.complianceData.consentId) {
      throw new UserInputError('ID de consentimento é obrigatório para transferências em Moçambique conforme Aviso 10/GBM/2016');
    }
    
    const consentValid = await complianceService.validateConsentById(input.complianceData.consentId);
    if (!consentValid) {
      throw new UserInputError('Consentimento inválido ou expirado');
    }
  }

  // Verificar limites específicos para Moçambique
  if (input.type === 'WITHDRAWAL' && input.amount > 5000 && (!input.complianceData.purposeCode || input.complianceData.purposeCode === '')) {
    throw new UserInputError('Código de finalidade é obrigatório para saques acima de 5.000 MZN');
  }
};

/**
 * Validações específicas para África do Sul (SARB/NCR)
 */
const validateSouthAfricanRequirements = async (input: any, complianceService: ComplianceService): Promise<void> => {
  // Verificar dados de compliance
  if (!input.complianceData) {
    throw new UserInputError('Dados de compliance são obrigatórios para transações na África do Sul');
  }

  // Verificar ID regulatório conforme POPIA e National Credit Act
  if (!input.complianceData.regulatoryId) {
    throw new UserInputError('ID regulatório é obrigatório para transações na África do Sul conforme POPIA');
  }
  
  // Verificar se o ID regulatório é válido
  const isValidRegulatoryId = await complianceService.validateSouthAfricanRegulatoryId(input.complianceData.regulatoryId);
  if (!isValidRegulatoryId) {
    throw new UserInputError('ID regulatório inválido conforme requisitos da POPIA');
  }
  
  // Verificar consentimento POPIA para processamento de dados pessoais
  const hasPopiaConsent = await complianceService.checkValidConsent(input.userId, 'POPIA_DATA_PROCESSING');
  if (!hasPopiaConsent) {
    throw new UserInputError('Consentimento POPIA para processamento de dados é obrigatório');
  }
};

/**
 * Validações específicas para Brasil (BACEN)
 */
const validateBrazilianRequirements = async (input: any, complianceService: ComplianceService): Promise<void> => {
  // Verificar dados de compliance
  if (!input.complianceData) {
    throw new UserInputError('Dados de compliance são obrigatórios para transações no Brasil');
  }

  // Verificar consentimento LGPD para processamento de dados pessoais
  const hasLgpdConsent = await complianceService.checkValidConsent(input.userId, 'LGPD_DATA_PROCESSING');
  if (!hasLgpdConsent) {
    throw new UserInputError('Consentimento LGPD para processamento de dados é obrigatório');
  }
  
  // Para Pix (transferências instantâneas), validações adicionais
  if (input.type === 'TRANSFER' && input.metadata?.transferType === 'PIX') {
    if (!input.metadata.pixKey) {
      throw new UserInputError('Chave Pix é obrigatória para transferências Pix');
    }
    
    const isValidPixKey = await complianceService.validatePixKey(input.metadata.pixKey);
    if (!isValidPixKey) {
      throw new UserInputError('Chave Pix inválida');
    }
  }
};