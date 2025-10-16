/**
 * Funções auxiliares para o serviço de processamento de transações Mobile Money
 * Implementação de métodos de validação, verificação e utilidades
 */

import { InitiateTransactionInput, RiskData, TransactionLimits } from './types';
import { RegionalComplianceService } from '../../infrastructure/compliance/regional_compliance_service';
import { RiskEnrichmentService } from '../../infrastructure/adaptive/risk_engine';
import { Logger } from '../../observability/logging/hook_logger';

/**
 * Classe que contém métodos auxiliares para o processamento de transações
 */
export class MobileMoneyTransactionHelpers {
  private readonly logger: Logger;
  private readonly regionalComplianceService: RegionalComplianceService;
  private readonly riskService: RiskEnrichmentService;
  
  constructor(
    logger: Logger,
    regionalComplianceService: RegionalComplianceService,
    riskService: RiskEnrichmentService
  ) {
    this.logger = logger;
    this.regionalComplianceService = regionalComplianceService;
    this.riskService = riskService;
  }

  /**
   * Verifica se o usuário tem permissões para realizar a operação
   */
  async verifyPermissions(userId: string, tenantId: string, permission: string): Promise<boolean> {
    try {
      this.logger.debug('Verifying user permissions', { userId, tenantId, permission });
      
      // Implementar verificação de permissões com o IAM
      // Esta seria uma integração com o módulo IAM para verificar permissões
      
      // Se a verificação falhar, lançar erro
      return true;
    } catch (error) {
      this.logger.error('Permission verification failed', { userId, tenantId, permission, error });
      throw new Error(`User does not have permission: ${permission}`);
    }
  }

  /**
   * Valida os limites de transação para o usuário
   */
  async validateTransactionLimits(input: InitiateTransactionInput): Promise<boolean> {
    const { userId, tenantId, amount, currency } = input;
    
    try {
      this.logger.debug('Validating transaction limits', { userId, tenantId, amount, currency });
      
      // Obter limites atuais do usuário
      const limits = await this.getTransactionLimits(userId, tenantId, currency);
      
      // Verificar limite por transação
      if (amount > limits.singleTransactionLimit) {
        throw new Error(`Transaction amount exceeds the single transaction limit of ${limits.singleTransactionLimit} ${currency}`);
      }
      
      // Verificar limite diário
      if (amount > limits.remainingDailyLimit) {
        throw new Error(`Transaction amount exceeds the remaining daily limit of ${limits.remainingDailyLimit} ${currency}`);
      }
      
      // Verificar limite mensal
      if (amount > limits.remainingMonthlyLimit) {
        throw new Error(`Transaction amount exceeds the remaining monthly limit of ${limits.remainingMonthlyLimit} ${currency}`);
      }
      
      return true;
    } catch (error) {
      this.logger.error('Transaction limit validation failed', { userId, tenantId, amount, currency, error });
      throw error;
    }
  }

  /**
   * Avalia o risco da transação
   */
  async assessTransactionRisk(input: InitiateTransactionInput): Promise<RiskData> {
    const { userId, tenantId, amount, currency, type, phoneNumber, deviceInfo } = input;
    
    try {
      this.logger.debug('Assessing transaction risk', { userId, tenantId, amount, currency, type });
      
      // Dados para avaliação de risco
      const riskContext = {
        userId,
        tenantId,
        transactionAmount: amount,
        transactionCurrency: currency,
        transactionType: type,
        phoneNumber,
        deviceInfo,
        timestamp: new Date(),
        ipAddress: deviceInfo?.ipAddress,
        location: deviceInfo?.geolocation,
        previousBehavior: {} // Dados de comportamento anterior seriam carregados aqui
      };
      
      // Chamar serviço de enriquecimento de risco
      const riskResult = await this.riskService.evaluateRisk(riskContext);
      
      return {
        score: riskResult.score,
        level: riskResult.level,
        factors: riskResult.factors,
        requiresReview: riskResult.requiresReview,
        requiresApproval: riskResult.requiresApproval,
        automaticallyDeclined: riskResult.automaticallyDeclined
      };
    } catch (error) {
      this.logger.error('Risk assessment failed', { userId, tenantId, error });
      
      // Em caso de falha na avaliação de risco, retornar um nível de risco padrão
      // para não bloquear a transação, mas marcar para revisão
      return {
        score: 50,
        level: 'MEDIUM',
        factors: ['RISK_SERVICE_UNAVAILABLE'],
        requiresReview: true,
        requiresApproval: false,
        automaticallyDeclined: false
      };
    }
  }

  /**
   * Valida conformidade regional para a transação
   */
  async validateRegionalCompliance(input: InitiateTransactionInput): Promise<boolean> {
    const { userId, tenantId, amount, currency, regionalData } = input;
    
    try {
      this.logger.debug('Validating regional compliance', { 
        userId, 
        tenantId, 
        currency,
        country: regionalData?.country 
      });
      
      // Determinar região/país baseado nos dados regionais ou na moeda
      const country = regionalData?.country || this.getCurrencyCountry(currency);
      
      // Obter regras de compliance regional
      const complianceRules = await this.regionalComplianceService.getComplianceRules(country);
      
      // Validar campos obrigatórios
      const missingFields = this.validateRequiredFields(input, complianceRules.requiredFields);
      if (missingFields.length > 0) {
        throw new Error(`Missing required fields for ${country}: ${missingFields.join(', ')}`);
      }
      
      // Validar código de finalidade (purpose code)
      if (complianceRules.requirePurposeCode && !input.purposeCode) {
        throw new Error(`Purpose code is required for transactions in ${country}`);
      }
      
      // Validar KYC mínimo exigido
      const userKycLevel = await this.getUserKycLevel(userId, tenantId);
      if (userKycLevel < complianceRules.minimumKycLevel) {
        throw new Error(`Insufficient KYC level for transaction in ${country}. Required: ${complianceRules.minimumKycLevel}, Current: ${userKycLevel}`);
      }
      
      // Validar consentimento
      if (complianceRules.requiresConsent && !input.consentId) {
        throw new Error(`Consent ID is required for transactions in ${country}`);
      }
      
      // Validar limites específicos da região
      if (amount > complianceRules.transactionLimits[input.type]) {
        throw new Error(`Transaction amount exceeds the ${input.type} limit of ${complianceRules.transactionLimits[input.type]} ${currency} for ${country}`);
      }
      
      return true;
    } catch (error) {
      this.logger.error('Regional compliance validation failed', { 
        userId, 
        tenantId, 
        country: regionalData?.country,
        error 
      });
      throw error;
    }
  }

  /**
   * Obtém país a partir do código da moeda
   */
  private getCurrencyCountry(currency: string): string {
    const currencyMap = {
      'AOA': 'ANGOLA',
      'MZN': 'MOZAMBIQUE',
      'ZAR': 'SOUTH_AFRICA',
      'BRL': 'BRAZIL',
      'EUR': 'PORTUGAL',
      // Adicionar mais moedas conforme necessário
    };
    
    return currencyMap[currency] || 'UNKNOWN';
  }

  /**
   * Valida campos obrigatórios para uma transação
   */
  private validateRequiredFields(input: any, requiredFields: string[]): string[] {
    const missingFields: string[] = [];
    
    for (const field of requiredFields) {
      const fieldParts = field.split('.');
      let value = input;
      
      for (const part of fieldParts) {
        value = value?.[part];
        if (value === undefined || value === null) {
          missingFields.push(field);
          break;
        }
      }
    }
    
    return missingFields;
  }

  /**
   * Obtém nível de KYC do usuário
   */
  private async getUserKycLevel(userId: string, tenantId: string): Promise<number> {
    // Implementar integração com serviço de KYC para obter o nível
    // Esta seria uma integração com o módulo de KYC
    
    // Nível KYC mock para fins de exemplo:
    // 0 - Não verificado
    // 1 - Básico
    // 2 - Médio
    // 3 - Completo
    return 3;
  }

  /**
   * Calcula taxas para uma transação
   */
  async calculateTransactionFees(input: InitiateTransactionInput): Promise<{
    amount: number;
    currency: string;
    description: string;
  } | undefined> {
    const { amount, currency, type, provider } = input;
    
    try {
      this.logger.debug('Calculating transaction fees', { amount, currency, type, provider });
      
      // Taxa fixa + percentual
      let feeAmount = 0;
      let feeDescription = '';
      
      // Lógica para cálculo de taxas baseada no tipo de transação e provedor
      switch (type) {
        case 'PAYMENT':
          feeAmount = Math.min(amount * 0.02, 50); // 2% até 50 unidades
          feeDescription = '2% de taxa de processamento';
          break;
        case 'TRANSFER':
          feeAmount = Math.min(amount * 0.015, 30); // 1.5% até 30 unidades
          feeDescription = '1.5% de taxa de transferência';
          break;
        case 'WITHDRAWAL':
          feeAmount = Math.min(amount * 0.025, 75); // 2.5% até 75 unidades
          feeDescription = '2.5% de taxa de saque';
          break;
        default:
          feeAmount = Math.min(amount * 0.01, 25); // 1% até 25 unidades
          feeDescription = '1% de taxa de serviço';
      }
      
      // Ajustar para valor mínimo se aplicável
      const minFee = 5;
      if (feeAmount < minFee) {
        feeAmount = minFee;
        feeDescription = `Taxa mínima de serviço: ${minFee} ${currency}`;
      }
      
      return {
        amount: feeAmount,
        currency,
        description: feeDescription
      };
    } catch (error) {
      this.logger.error('Failed to calculate transaction fees', { error });
      return undefined;
    }
  }

  /**
   * Obtém limites de transação para um usuário
   */
  async getTransactionLimits(userId: string, tenantId: string, currency: string): Promise<TransactionLimits> {
    try {
      this.logger.debug('Getting transaction limits', { userId, tenantId, currency });
      
      // Implementar integração com serviço de limites
      // Esta seria uma integração com o módulo de gestão de limites
      
      // Limites mock para fins de exemplo
      return {
        singleTransactionLimit: 10000,
        dailyLimit: 25000,
        monthlyLimit: 100000,
        remainingDailyLimit: 20000,
        remainingMonthlyLimit: 80000,
        currency
      };
    } catch (error) {
      this.logger.error('Failed to get transaction limits', { userId, tenantId, currency, error });
      throw error;
    }
  }
}