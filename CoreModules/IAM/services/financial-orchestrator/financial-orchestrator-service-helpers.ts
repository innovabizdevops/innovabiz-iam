/**
 * Funções auxiliares para o serviço orquestrador de serviços financeiros
 */

import { 
  FinancialOperationRequest, 
  FinancialOperationResult, 
  FinancialServiceType, 
  FinancialOperationType, 
  OrchestrationStatus 
} from './types';
import { MobileMoneyProvider } from '../mobile-money/types';

/**
 * Classe auxiliar com funções de suporte para o orquestrador
 */
export class FinancialOrchestratorHelpers {
  /**
   * Verifica se uma operação está em andamento (não finalizada)
   */
  static isOperationInProgress(status: OrchestrationStatus): boolean {
    const finalStatuses = [
      OrchestrationStatus.COMPLETED,
      OrchestrationStatus.FAILED,
      OrchestrationStatus.CANCELLED
    ];
    
    return !finalStatuses.includes(status);
  }
  
  /**
   * Verifica se uma operação está concluída
   */
  static isOperationCompleted(status: OrchestrationStatus): boolean {
    return status === OrchestrationStatus.COMPLETED;
  }
  
  /**
   * Verifica se uma operação pode ser cancelada
   */
  static isOperationCancellable(status: OrchestrationStatus): boolean {
    const cancellableStatuses = [
      OrchestrationStatus.INITIATED,
      OrchestrationStatus.PROCESSING,
      OrchestrationStatus.PENDING_USER_ACTION,
      OrchestrationStatus.PENDING_APPROVAL,
      OrchestrationStatus.PENDING_PROVIDER,
      OrchestrationStatus.PENDING_COMPLIANCE,
      OrchestrationStatus.PENDING_RISK_ASSESSMENT
    ];
    
    return cancellableStatuses.includes(status);
  }
  
  /**
   * Extrai o tempo de expiração em segundos de uma resposta
   */
  static getExpirationTimeFromResponse(response: any): number {
    if (response.expiresIn) {
      return response.expiresIn;
    }
    
    if (response.expiresAt) {
      const expiresAt = new Date(response.expiresAt).getTime();
      const now = Date.now();
      return Math.max(0, Math.floor((expiresAt - now) / 1000));
    }
    
    // Default: 1 hora
    return 3600;
  }
  
  /**
   * Normaliza erros de serviços financeiros
   */
  static normalizeError(error: any, serviceType: FinancialServiceType): Error {
    // Erro já normalizado
    if (error instanceof Error) {
      return error;
    }
    
    // Erro de resposta HTTP
    if (error.response) {
      const statusCode = error.response.status;
      const message = error.response.data?.message || error.response.statusText || 'Erro desconhecido';
      
      let errorMessage = `[${serviceType}] Erro ${statusCode}: ${message}`;
      
      // Categorizar erros comuns para diferentes serviços
      switch (statusCode) {
        case 401:
          return new Error(`[${serviceType}] Autenticação falhou: ${message}`);
        case 403:
          return new Error(`[${serviceType}] Acesso negado: ${message}`);
        case 404:
          return new Error(`[${serviceType}] Recurso não encontrado: ${message}`);
        case 422:
          return new Error(`[${serviceType}] Dados inválidos: ${message}`);
        case 429:
          return new Error(`[${serviceType}] Limite de requisições excedido. Tente novamente em alguns minutos.`);
        case 500:
        case 502:
        case 503:
        case 504:
          return new Error(`[${serviceType}] Erro no servidor: ${message}. Tente novamente mais tarde.`);
        default:
          return new Error(errorMessage);
      }
    }
    
    // Erro de rede
    if (error.request) {
      return new Error(`[${serviceType}] Erro de rede: O serviço não respondeu. Verifique sua conexão.`);
    }
    
    // Erro genérico
    return new Error(`[${serviceType}] ${error.message || 'Erro desconhecido'}`);
  }
  
  /**
   * Constrói um objeto de resposta padrão para falha
   */
  static buildFailureResponse(
    operationId: string,
    request: FinancialOperationRequest,
    errorMessage: string
  ): FinancialOperationResult {
    return {
      operationId,
      status: OrchestrationStatus.FAILED,
      serviceType: request.serviceType,
      amount: request.amount,
      currency: request.currency,
      userId: request.userId,
      tenantId: request.tenantId,
      timestamp: new Date(),
      failureReason: errorMessage,
      metadata: request.metadata
    };
  }
  
  /**
   * Mapeia nome de serviços para nomes adequados para logs e métricas
   */
  static getServiceMetricsName(serviceType: FinancialServiceType): string {
    const mapping = {
      [FinancialServiceType.MOBILE_MONEY]: 'mobile_money',
      [FinancialServiceType.E_COMMERCE]: 'ecommerce',
      [FinancialServiceType.PAYMENT_GATEWAY]: 'payment_gateway',
      [FinancialServiceType.BUREAU_CREDIT]: 'bureau_credit',
      [FinancialServiceType.MICROFINANCE]: 'microfinance',
      [FinancialServiceType.INSURANCE]: 'insurance'
    };
    
    return mapping[serviceType] || serviceType.toLowerCase();
  }
  
  /**
   * Mapeia nome de operações para nomes adequados para logs e métricas
   */
  static getOperationMetricsName(operationType: FinancialOperationType): string {
    const mapping = {
      [FinancialOperationType.PAYMENT]: 'payment',
      [FinancialOperationType.TRANSFER]: 'transfer',
      [FinancialOperationType.WITHDRAWAL]: 'withdrawal',
      [FinancialOperationType.DEPOSIT]: 'deposit',
      [FinancialOperationType.CHECKOUT]: 'checkout',
      [FinancialOperationType.REFUND]: 'refund',
      [FinancialOperationType.CREDIT_CHECK]: 'credit_check',
      [FinancialOperationType.INSURANCE_CLAIM]: 'insurance_claim',
      [FinancialOperationType.CREDIT_APPLICATION]: 'credit_application'
    };
    
    return mapping[operationType] || operationType.toLowerCase();
  }
}