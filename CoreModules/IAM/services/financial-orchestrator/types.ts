/**
 * Tipos e interfaces para o orquestrador de serviços financeiros
 * 
 * Define as estruturas de dados e interfaces para o middleware responsável
 * pela orquestração e integração entre os diferentes serviços financeiros
 * da plataforma INNOVABIZ.
 */

import { MobileMoneyProvider, TransactionStatus, TransactionType } from '../mobile-money/types';
import { EcommerceUserRole } from '../e-commerce/types';

/**
 * Tipos de serviços financeiros
 */
export enum FinancialServiceType {
  MOBILE_MONEY = 'MOBILE_MONEY',
  E_COMMERCE = 'E_COMMERCE',
  PAYMENT_GATEWAY = 'PAYMENT_GATEWAY',
  BUREAU_CREDIT = 'BUREAU_CREDIT',
  MICROFINANCE = 'MICROFINANCE',
  INSURANCE = 'INSURANCE'
}

/**
 * Tipos de operações financeiras
 */
export enum FinancialOperationType {
  PAYMENT = 'PAYMENT',
  TRANSFER = 'TRANSFER',
  WITHDRAWAL = 'WITHDRAWAL',
  DEPOSIT = 'DEPOSIT',
  CHECKOUT = 'CHECKOUT',
  REFUND = 'REFUND',
  CREDIT_CHECK = 'CREDIT_CHECK',
  INSURANCE_CLAIM = 'INSURANCE_CLAIM',
  CREDIT_APPLICATION = 'CREDIT_APPLICATION'
}

/**
 * Status das operações orquestradas
 */
export enum OrchestrationStatus {
  INITIATED = 'INITIATED',
  PROCESSING = 'PROCESSING',
  COMPLETED = 'COMPLETED',
  FAILED = 'FAILED',
  CANCELLED = 'CANCELLED',
  PENDING_USER_ACTION = 'PENDING_USER_ACTION',
  PENDING_APPROVAL = 'PENDING_APPROVAL',
  PENDING_PROVIDER = 'PENDING_PROVIDER',
  PENDING_COMPLIANCE = 'PENDING_COMPLIANCE',
  PENDING_RISK_ASSESSMENT = 'PENDING_RISK_ASSESSMENT'
}

/**
 * Níveis de prioridade das operações
 */
export enum OperationPriority {
  LOW = 'LOW',
  MEDIUM = 'MEDIUM',
  HIGH = 'HIGH',
  CRITICAL = 'CRITICAL'
}

/**
 * Interface para uma solicitação de operação financeira
 */
export interface FinancialOperationRequest {
  operationType: FinancialOperationType;
  serviceType: FinancialServiceType;
  provider?: MobileMoneyProvider | string;
  amount?: number;
  currency?: string;
  userId: string;
  tenantId: string;
  referenceId?: string;
  metadata?: Record<string, any>;
  callbackUrl?: string;
  priority?: OperationPriority;
  requiredApprovals?: number;
  deviceInfo?: {
    deviceId?: string;
    ipAddress?: string;
    userAgent?: string;
    location?: {
      latitude?: number;
      longitude?: number;
      country?: string;
      city?: string;
    };
  };
}

/**
 * Interface para o resultado de uma operação financeira
 */
export interface FinancialOperationResult {
  operationId: string;
  status: OrchestrationStatus;
  serviceType: FinancialServiceType;
  transactionId?: string;
  referenceId?: string;
  providerReference?: string;
  amount?: number;
  currency?: string;
  timestamp: Date;
  completedAt?: Date;
  userId: string;
  tenantId: string;
  requiredAction?: {
    type: string;
    instructions?: string;
    url?: string;
    timeoutSeconds?: number;
    otpRequired?: boolean;
    otpSent?: boolean;
    otpPhoneNumber?: string;
  };
  receiptUrl?: string;
  failureReason?: string;
  metadata?: Record<string, any>;
}

/**
 * Interface para histórico de operações financeiras
 */
export interface FinancialOperationHistory {
  operations: FinancialOperationResult[];
  totalCount: number;
  hasMore: boolean;
}

/**
 * Interface para consulta de serviços disponíveis
 */
export interface AvailableFinancialServices {
  services: {
    serviceType: FinancialServiceType;
    providers: string[];
    operations: FinancialOperationType[];
    available: boolean;
    requiresKyc: boolean;
  }[];
  eligibleOperations: FinancialOperationType[];
  kycStatus: {
    verified: boolean;
    level: string;
    missingRequirements?: string[];
  };
}

/**
 * Interface para o middleware orquestrador de serviços financeiros
 */
export interface FinancialOrchestratorService {
  // Orquestração de operações
  initiateOperation(request: FinancialOperationRequest): Promise<FinancialOperationResult>;
  checkOperationStatus(operationId: string, tenantId: string): Promise<FinancialOperationResult>;
  cancelOperation(operationId: string, tenantId: string, reason?: string): Promise<FinancialOperationResult>;
  
  // Mobile Money (orquestração específica)
  initiatePayment(params: {
    userId: string;
    tenantId: string;
    provider: MobileMoneyProvider;
    amount: number;
    currency: string;
    phoneNumber: string;
    description?: string;
    metadata?: Record<string, any>;
    deviceInfo?: any;
  }): Promise<FinancialOperationResult>;
  
  verifyTransactionOtp(params: {
    operationId: string;
    tenantId: string;
    otpCode: string;
    deviceInfo?: any;
  }): Promise<FinancialOperationResult>;
  
  // E-Commerce (orquestração específica)
  processCheckout(params: {
    userId: string;
    tenantId: string;
    shopId: string;
    checkoutSessionId: string;
    paymentMethod: string;
    deviceInfo?: any;
  }): Promise<FinancialOperationResult>;
  
  // Consulta e histórico
  getOperationHistory(userId: string, tenantId: string, options?: {
    serviceType?: FinancialServiceType;
    operationType?: FinancialOperationType;
    startDate?: Date;
    endDate?: Date;
    limit?: number;
    offset?: number;
  }): Promise<FinancialOperationHistory>;
  
  getOperationDetails(operationId: string, tenantId: string): Promise<FinancialOperationResult>;
  
  getAvailableServices(userId: string, tenantId: string): Promise<AvailableFinancialServices>;
}