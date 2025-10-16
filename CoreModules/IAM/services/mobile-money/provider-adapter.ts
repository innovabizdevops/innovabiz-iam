/**
 * Adaptador para provedores de Mobile Money
 * 
 * Esta classe implementa a comunicação com APIs de provedores específicos
 * de Mobile Money como MPesa, Airtel, MTN, etc.
 */

import axios from 'axios';
import { v4 as uuidv4 } from 'uuid';
import { Logger } from '../../observability/logging/hook_logger';
import { MobileMoneyProvider } from './types';
import { MobileMoneyProviderInstance } from './provider-factory';

/**
 * Implementação do adaptador para provedores Mobile Money
 */
export class MobileMoneyProviderAdapter implements MobileMoneyProviderInstance {
  private readonly provider: MobileMoneyProvider;
  private readonly config: any;
  private readonly logger: Logger;
  private authToken: string | null = null;
  private tokenExpiry: Date | null = null;
  
  constructor(provider: MobileMoneyProvider, config: any, logger: Logger) {
    this.provider = provider;
    this.config = config;
    this.logger = logger;
  }
  
  /**
   * Inicializa uma transação com o provedor
   */
  async initiateTransaction(params: {
    transactionId: string;
    amount: number;
    currency: string;
    phoneNumber: string;
    recipientPhone?: string;
    recipientName?: string;
    description?: string;
    type: string;
    metadata?: Record<string, any>;
    notifyRecipient?: boolean;
  }): Promise<{
    referenceNumber: string;
    providerReference?: string;
    otpRequired: boolean;
    otpSent: boolean;
    otpPhoneNumber?: string;
    kycStatus?: 'VERIFIED' | 'PENDING' | 'NOT_VERIFIED';
    requiredAction?: {
      type: string;
      instructions?: string;
      url?: string;
      timeoutSeconds?: number;
    };
  }> {
    try {
      this.logger.info(`Initiating ${this.provider} transaction`, { 
        transactionId: params.transactionId,
        type: params.type,
        amount: params.amount,
        currency: params.currency
      });
      
      // Garantir que temos um token de autenticação válido
      await this.ensureAuthToken();
      
      // Construir payload de acordo com especificações do provedor
      const payload = this.buildInitiateTransactionPayload(params);
      
      // Fazer chamada para a API do provedor
      const response = await axios({
        method: 'POST',
        url: `${this.config.apiEndpoint}/transactions`,
        headers: {
          'Authorization': `Bearer ${this.authToken}`,
          'Content-Type': 'application/json',
          'X-Provider-Api-Key': this.config.apiKey,
          'X-Correlation-ID': params.transactionId
        },
        data: payload,
        timeout: this.config.timeoutMs || 15000
      });
      
      // Processar resposta
      const result = response.data;
      
      this.logger.info(`${this.provider} transaction initiated successfully`, {
        transactionId: params.transactionId,
        providerReference: result.referenceId,
        otpRequired: result.otpRequired
      });
      
      return {
        referenceNumber: result.referenceId || uuidv4(),
        providerReference: result.providerTransactionId,
        otpRequired: result.otpRequired || this.config.features.requiresOTP,
        otpSent: result.otpSent || false,
        otpPhoneNumber: result.otpSent ? params.phoneNumber : undefined,
        kycStatus: result.kycStatus || 'VERIFIED',
        requiredAction: result.requiredAction
      };
    } catch (error) {
      this.logger.error(`Failed to initiate ${this.provider} transaction`, { 
        error,
        transactionId: params.transactionId
      });
      
      // Transformar erro específico do provedor em formato padronizado
      throw this.normalizeProviderError(error);
    }
  }
  
  /**
   * Verifica código OTP com o provedor
   */
  async verifyOTP(params: {
    transactionId: string;
    otpCode: string;
  }): Promise<{
    verified: boolean;
    status: string;
    failureReason?: string;
    remainingAttempts?: number;
  }> {
    try {
      this.logger.info(`Verifying OTP for ${this.provider} transaction`, { 
        transactionId: params.transactionId
      });
      
      // Garantir que temos um token de autenticação válido
      await this.ensureAuthToken();
      
      // Fazer chamada para a API do provedor
      const response = await axios({
        method: 'POST',
        url: `${this.config.apiEndpoint}/transactions/${params.transactionId}/verify-otp`,
        headers: {
          'Authorization': `Bearer ${this.authToken}`,
          'Content-Type': 'application/json',
          'X-Provider-Api-Key': this.config.apiKey
        },
        data: {
          otpCode: params.otpCode
        },
        timeout: this.config.timeoutMs || 15000
      });
      
      // Processar resposta
      const result = response.data;
      
      this.logger.info(`OTP verification result for ${this.provider} transaction`, {
        transactionId: params.transactionId,
        verified: result.verified,
        status: result.status
      });
      
      return {
        verified: result.verified,
        status: result.status,
        failureReason: result.failureReason,
        remainingAttempts: result.remainingAttempts
      };
    } catch (error) {
      this.logger.error(`Failed to verify OTP for ${this.provider} transaction`, { 
        error,
        transactionId: params.transactionId
      });
      
      // Transformar erro específico do provedor em formato padronizado
      throw this.normalizeProviderError(error);
    }
  }
  
  /**
   * Verifica status de uma transação com o provedor
   */
  async checkStatus(params: {
    transactionId: string;
  }): Promise<{
    status: string;
    completedAt?: Date;
    failureReason?: string;
    providerReference?: string;
    receiptNumber?: string;
  }> {
    try {
      this.logger.info(`Checking status for ${this.provider} transaction`, { 
        transactionId: params.transactionId
      });
      
      // Garantir que temos um token de autenticação válido
      await this.ensureAuthToken();
      
      // Fazer chamada para a API do provedor
      const response = await axios({
        method: 'GET',
        url: `${this.config.apiEndpoint}/transactions/${params.transactionId}`,
        headers: {
          'Authorization': `Bearer ${this.authToken}`,
          'Content-Type': 'application/json',
          'X-Provider-Api-Key': this.config.apiKey
        },
        timeout: this.config.timeoutMs || 15000
      });
      
      // Processar resposta
      const result = response.data;
      
      this.logger.info(`Status check result for ${this.provider} transaction`, {
        transactionId: params.transactionId,
        status: result.status
      });
      
      return {
        status: this.mapProviderStatus(result.status),
        completedAt: result.completedAt ? new Date(result.completedAt) : undefined,
        failureReason: result.failureReason,
        providerReference: result.providerTransactionId,
        receiptNumber: result.receiptNumber
      };
    } catch (error) {
      this.logger.error(`Failed to check status for ${this.provider} transaction`, { 
        error,
        transactionId: params.transactionId
      });
      
      // Transformar erro específico do provedor em formato padronizado
      throw this.normalizeProviderError(error);
    }
  }
  
  /**
   * Cancela uma transação com o provedor
   */
  async cancelTransaction(params: {
    transactionId: string;
    reason?: string;
  }): Promise<{
    cancelled: boolean;
    status: string;
    failureReason?: string;
  }> {
    try {
      this.logger.info(`Cancelling ${this.provider} transaction`, { 
        transactionId: params.transactionId
      });
      
      // Garantir que temos um token de autenticação válido
      await this.ensureAuthToken();
      
      // Fazer chamada para a API do provedor
      const response = await axios({
        method: 'POST',
        url: `${this.config.apiEndpoint}/transactions/${params.transactionId}/cancel`,
        headers: {
          'Authorization': `Bearer ${this.authToken}`,
          'Content-Type': 'application/json',
          'X-Provider-Api-Key': this.config.apiKey
        },
        data: {
          reason: params.reason
        },
        timeout: this.config.timeoutMs || 15000
      });
      
      // Processar resposta
      const result = response.data;
      
      this.logger.info(`Cancellation result for ${this.provider} transaction`, {
        transactionId: params.transactionId,
        cancelled: result.cancelled,
        status: result.status
      });
      
      return {
        cancelled: result.cancelled,
        status: this.mapProviderStatus(result.status),
        failureReason: result.failureReason
      };
    } catch (error) {
      this.logger.error(`Failed to cancel ${this.provider} transaction`, { 
        error,
        transactionId: params.transactionId
      });
      
      // Transformar erro específico do provedor em formato padronizado
      throw this.normalizeProviderError(error);
    }
  }
  
  /**
   * Obtém detalhes adicionais sobre uma transação
   */
  async getDetails(params: {
    transactionId: string;
  }): Promise<any> {
    try {
      this.logger.info(`Getting details for ${this.provider} transaction`, { 
        transactionId: params.transactionId
      });
      
      // Garantir que temos um token de autenticação válido
      await this.ensureAuthToken();
      
      // Fazer chamada para a API do provedor
      const response = await axios({
        method: 'GET',
        url: `${this.config.apiEndpoint}/transactions/${params.transactionId}/details`,
        headers: {
          'Authorization': `Bearer ${this.authToken}`,
          'Content-Type': 'application/json',
          'X-Provider-Api-Key': this.config.apiKey
        },
        timeout: this.config.timeoutMs || 15000
      });
      
      // Processar resposta
      return response.data;
    } catch (error) {
      this.logger.error(`Failed to get details for ${this.provider} transaction`, { 
        error,
        transactionId: params.transactionId
      });
      
      // Transformar erro específico do provedor em formato padronizado
      throw this.normalizeProviderError(error);
    }
  }
  
  /**
   * Garante que temos um token de autenticação válido
   */
  private async ensureAuthToken(): Promise<void> {
    // Verificar se já temos um token válido
    if (this.authToken && this.tokenExpiry && this.tokenExpiry > new Date()) {
      return;
    }
    
    try {
      this.logger.debug(`Authenticating with ${this.provider}`);
      
      // Fazer chamada para obter token
      const response = await axios({
        method: 'POST',
        url: `${this.config.apiEndpoint}/auth/token`,
        headers: {
          'Content-Type': 'application/json'
        },
        data: {
          apiKey: this.config.apiKey,
          apiSecret: this.config.apiSecret
        },
        timeout: this.config.timeoutMs || 15000
      });
      
      // Armazenar token e data de expiração
      this.authToken = response.data.accessToken;
      this.tokenExpiry = new Date(Date.now() + response.data.expiresIn * 1000);
      
      this.logger.debug(`Successfully authenticated with ${this.provider}`);
    } catch (error) {
      this.logger.error(`Failed to authenticate with ${this.provider}`, { error });
      throw new Error(`Authentication failed with ${this.provider}: ${error.message}`);
    }
  }
  
  /**
   * Constrói payload para iniciação de transação baseado no provedor
   */
  private buildInitiateTransactionPayload(params: any): any {
    // Construir payload baseado nas especificações do provedor
    switch (this.provider) {
      case MobileMoneyProvider.MPESA:
        return {
          amount: params.amount,
          currency: params.currency,
          msisdn: params.phoneNumber.replace('+', ''),
          reference: params.transactionId,
          thirdPartyReference: params.metadata?.referenceId || params.transactionId,
          transactionType: this.mapTransactionType(params.type, 'MPESA'),
          description: params.description || 'Mobile Money transaction'
        };
        
      case MobileMoneyProvider.AIRTEL:
        return {
          amount: params.amount,
          currencyCode: params.currency,
          phoneNumber: params.phoneNumber,
          externalId: params.transactionId,
          payerNote: params.description || 'Airtel Money transaction',
          transactionType: this.mapTransactionType(params.type, 'AIRTEL')
        };
        
      case MobileMoneyProvider.MTN:
        return {
          amount: params.amount,
          currency: params.currency,
          mobileNumber: params.phoneNumber,
          externalId: params.transactionId,
          paymentDescription: params.description || 'MTN Mobile Money transaction',
          type: this.mapTransactionType(params.type, 'MTN')
        };
        
      case MobileMoneyProvider.UNITEL:
        return {
          valor: params.amount,
          moeda: params.currency,
          numero: params.phoneNumber.replace('+', ''),
          referencia: params.transactionId,
          descricao: params.description || 'Transação Unitel Money',
          tipoTransacao: this.mapTransactionType(params.type, 'UNITEL')
        };
        
      default:
        // Formato genérico para outros provedores
        return {
          amount: params.amount,
          currency: params.currency,
          phoneNumber: params.phoneNumber,
          transactionId: params.transactionId,
          description: params.description || 'Mobile Money transaction',
          type: params.type,
          metadata: params.metadata || {}
        };
    }
  }
  
  /**
   * Mapeia tipo de transação para formato específico do provedor
   */
  private mapTransactionType(type: string, provider: string): string {
    const typeMap: Record<string, Record<string, string>> = {
      'MPESA': {
        'PAYMENT': 'CustomerPayBillOnline',
        'TRANSFER': 'BusinessPayment',
        'WITHDRAWAL': 'CustomerWithdraw',
        'DEPOSIT': 'CustomerDeposit'
      },
      'AIRTEL': {
        'PAYMENT': 'PAYMENT',
        'TRANSFER': 'TRANSFER',
        'WITHDRAWAL': 'WITHDRAW',
        'DEPOSIT': 'DEPOSIT'
      },
      'MTN': {
        'PAYMENT': 'PAYMENT',
        'TRANSFER': 'TRANSFER',
        'WITHDRAWAL': 'CASHOUT',
        'DEPOSIT': 'CASHIN'
      },
      'UNITEL': {
        'PAYMENT': 'PAGAMENTO',
        'TRANSFER': 'TRANSFERENCIA',
        'WITHDRAWAL': 'LEVANTAMENTO',
        'DEPOSIT': 'DEPOSITO'
      }
    };
    
    return typeMap[provider]?.[type] || type;
  }
  
  /**
   * Mapeia status do provedor para formato padronizado
   */
  private mapProviderStatus(providerStatus: string): string {
    // Mapear status específicos do provedor para nossos status padronizados
    const statusMap: Record<string, string> = {
      // MPesa
      'SUCCESS': 'COMPLETED',
      'PENDING': 'PROCESSING',
      'FAILED': 'FAILED',
      'CANCELLED': 'CANCELLED',
      'EXPIRED': 'EXPIRED',
      'REJECTED': 'REJECTED',
      
      // Airtel
      'TS': 'COMPLETED',  // Transaction Success
      'TIP': 'PROCESSING', // Transaction In Progress
      'TF': 'FAILED',     // Transaction Failed
      
      // MTN
      'SUCCESSFUL': 'COMPLETED',
      'PENDING_APPROVAL': 'PENDING_APPROVAL',
      'PROCESSING': 'PROCESSING',
      'ONGOING': 'PROCESSING',
      'FAILED': 'FAILED',
      
      // Unitel
      'COMPLETA': 'COMPLETED',
      'EM_PROCESSAMENTO': 'PROCESSING',
      'FALHA': 'FAILED',
      'CANCELADA': 'CANCELLED',
      'EXPIRADA': 'EXPIRED',
      'REJEITADA': 'REJECTED'
    };
    
    return statusMap[providerStatus] || 'PROCESSING';
  }
  
  /**
   * Normaliza erros específicos do provedor para formato padrão
   */
  private normalizeProviderError(error: any): Error {
    // Erros de rede/timeout
    if (error.code === 'ECONNABORTED') {
      return new Error(`Connection timeout with ${this.provider} API`);
    }
    
    if (!error.response) {
      return new Error(`Network error connecting to ${this.provider} API: ${error.message}`);
    }
    
    // Erros de resposta HTTP
    const status = error.response.status;
    const data = error.response.data || {};
    
    // Mapear códigos de erro específicos do provedor
    const errorMapping: Record<string, string> = {
      'INSUFFICIENT_FUNDS': 'Insufficient funds in the account',
      'INVALID_ACCOUNT': 'Invalid account or phone number',
      'UNAUTHORIZED_ACCESS': 'Unauthorized access',
      'TRANSACTION_LIMIT_EXCEEDED': 'Transaction limit exceeded',
      'INVALID_OTP': 'Invalid OTP code',
      'EXPIRED_OTP': 'OTP code expired',
      'INVALID_TRANSACTION': 'Invalid transaction',
      'DUPLICATE_TRANSACTION': 'Duplicate transaction'
    };
    
    const errorCode = data.errorCode || data.error?.code || '';
    const errorMessage = errorMapping[errorCode] || data.message || data.error?.message || error.message;
    
    return new Error(`${this.provider} API error (${status}): ${errorMessage}`);
  }
}