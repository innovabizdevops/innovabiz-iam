/**
 * Tipos e interfaces para o serviço de processamento de transações Mobile Money
 * 
 * Este arquivo define todas as interfaces e tipos necessários para o processamento
 * de transações Mobile Money no ecossistema INNOVABIZ, seguindo os padrões
 * internacionais e requisitos regulatórios regionais.
 */

import { RegionalRules } from '../../infrastructure/compliance/types';

/**
 * Status possíveis para uma transação Mobile Money
 */
export enum TransactionStatus {
  INITIATED = 'INITIATED',
  OTP_SENT = 'OTP_SENT', 
  OTP_VERIFIED = 'OTP_VERIFIED',
  PROCESSING = 'PROCESSING',
  COMPLETED = 'COMPLETED',
  FAILED = 'FAILED',
  CANCELLED = 'CANCELLED',
  EXPIRED = 'EXPIRED',
  PENDING_APPROVAL = 'PENDING_APPROVAL',
  REJECTED = 'REJECTED'
}

/**
 * Tipos de transação Mobile Money suportados
 */
export enum TransactionType {
  PAYMENT = 'PAYMENT',
  TRANSFER = 'TRANSFER',
  DEPOSIT = 'DEPOSIT',
  WITHDRAWAL = 'WITHDRAWAL',
  REFUND = 'REFUND',
  BILL_PAYMENT = 'BILL_PAYMENT',
  MERCHANT_PAYMENT = 'MERCHANT_PAYMENT',
  AIRTIME = 'AIRTIME'
}

/**
 * Motivos de falha de transação
 */
export enum FailureReason {
  INSUFFICIENT_FUNDS = 'INSUFFICIENT_FUNDS',
  INVALID_ACCOUNT = 'INVALID_ACCOUNT',
  NETWORK_ERROR = 'NETWORK_ERROR',
  INVALID_OTP = 'INVALID_OTP',
  TIMEOUT = 'TIMEOUT',
  LIMIT_EXCEEDED = 'LIMIT_EXCEEDED',
  SECURITY_CHECK_FAILED = 'SECURITY_CHECK_FAILED',
  COMPLIANCE_CHECK_FAILED = 'COMPLIANCE_CHECK_FAILED',
  RISK_CHECK_FAILED = 'RISK_CHECK_FAILED',
  PROVIDER_ERROR = 'PROVIDER_ERROR',
  USER_CANCELLED = 'USER_CANCELLED',
  OTHER = 'OTHER'
}

/**
 * Provedores de Mobile Money suportados
 */
export enum MobileMoneyProvider {
  MPESA = 'MPESA',
  AIRTEL = 'AIRTEL',
  ORANGE = 'ORANGE',
  MTN = 'MTN',
  UNITEL = 'UNITEL',
  ECO_CASH = 'ECO_CASH',
  VODAFONE = 'VODAFONE',
  TIGO = 'TIGO',
  MOVICEL = 'MOVICEL',
  TMN = 'TMN'
}

/**
 * Interface para configuração de provedores Mobile Money
 */
export interface ProviderConfig {
  providerId: MobileMoneyProvider;
  apiEndpoint: string;
  apiKey: string;
  apiSecret: string;
  callbackUrl?: string;
  timeoutMs: number;
  retryCount: number;
  retryDelayMs: number;
  supportedCountries: string[];
  supportedTransactionTypes: TransactionType[];
  limits: {
    minAmount: number;
    maxAmount: number;
    dailyLimit?: number;
    monthlyLimit?: number;
  };
  features: {
    supportsOTP: boolean;
    requiresOTP: boolean;
    supportsPush: boolean;
    supportsCallback: boolean;
    supportsStatusCheck: boolean;
    supportsBulkPayments: boolean;
  };
}

/**
 * Interface para informações de dispositivo
 */
export interface DeviceInfo {
  deviceId?: string;
  deviceType?: string;
  operatingSystem?: string;
  osVersion?: string;
  browserType?: string;
  browserVersion?: string;
  ipAddress?: string;
  isMobile?: boolean;
  geolocation?: {
    latitude?: number;
    longitude?: number;
    accuracy?: number;
  };
  fingerprint?: string;
  isTrustedDevice?: boolean;
  lastUsed?: Date;
}

/**
 * Interface para dados de risco
 */
export interface RiskData {
  score: number;
  level: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
  factors: string[];
  requiresReview: boolean;
  requiresApproval: boolean;
  automaticallyDeclined: boolean;
}

/**
 * Interface para dados de conformidade
 */
export interface ComplianceData {
  kycStatus: 'VERIFIED' | 'PENDING' | 'NOT_VERIFIED' | 'EXPIRED';
  kycLevel: 'NONE' | 'BASIC' | 'MEDIUM' | 'FULL';
  consentIds: string[];
  purposeCode?: string;
  sanctionScreeningPassed?: boolean;
  amlChecksPassed?: boolean;
  regulatoryRequirementsMet?: boolean;
  documentsVerified?: string[];
  complianceNotes?: string;
  applicableRegulations?: string[];
}

/**
 * Interface para dados específicos de região/país
 */
export interface RegionalData {
  country: string;
  region?: string;
  currency: string;
  languageCode: string;
  regulatoryRegime: string;
  documentTypes: string[];
  requiredFields: string[];
  purposeCodes: string[];
}

/**
 * Interface para iniciar uma nova transação
 */
export interface InitiateTransactionInput {
  userId: string;
  tenantId: string;
  type: TransactionType;
  amount: number;
  currency: string;
  provider: MobileMoneyProvider;
  phoneNumber: string;
  recipient?: {
    phoneNumber?: string;
    accountId?: string;
    name?: string;
    bankCode?: string;
  };
  description?: string;
  referenceId?: string;
  metadata?: Record<string, any>;
  purposeCode?: string;
  callbackUrl?: string;
  deviceInfo?: DeviceInfo;
  consentId?: string;
  notifyRecipient?: boolean;
  notificationLanguage?: string;
  expiresInSeconds?: number;
  scheduledTime?: Date;
  immediateProcessing?: boolean;
  regionalData?: Record<string, any>;
}

/**
 * Interface para saída de iniciação de transação
 */
export interface InitiateTransactionOutput {
  transactionId: string;
  status: TransactionStatus;
  referenceNumber?: string;
  otpRequired: boolean;
  otpSent: boolean;
  otpPhoneNumber?: string;
  expiresAt: Date;
  processingEstimateSeconds?: number;
  fees?: {
    amount: number;
    currency: string;
    description: string;
  };
  totalAmount: number;
  requiredAction?: {
    type: 'OTP_VERIFICATION' | 'REDIRECT' | 'APPROVAL' | 'ADDITIONAL_INFO';
    instructions?: string;
    url?: string;
    timeoutSeconds?: number;
  };
  riskAssessment?: {
    level: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
    requiresAdditionalVerification: boolean;
  };
}

/**
 * Interface para verificação de OTP
 */
export interface VerifyOTPInput {
  transactionId: string;
  tenantId: string;
  otpCode: string;
  deviceInfo?: DeviceInfo;
}

/**
 * Interface para saída de verificação de OTP
 */
export interface VerifyOTPOutput {
  transactionId: string;
  verified: boolean;
  status: TransactionStatus;
  failureReason?: FailureReason;
  remainingAttempts?: number;
  nextAction?: {
    type: 'WAIT' | 'RETRY' | 'ADDITIONAL_VERIFICATION' | 'CONTACT_SUPPORT';
    instructions?: string;
  };
}

/**
 * Interface para verificar status de transação
 */
export interface CheckTransactionStatusInput {
  transactionId: string;
  tenantId: string;
  includeDetails?: boolean;
}

/**
 * Interface para saída de verificação de status
 */
export interface CheckTransactionStatusOutput {
  transactionId: string;
  status: TransactionStatus;
  timestamp: Date;
  completedAt?: Date;
  failureReason?: FailureReason;
  failureDescription?: string;
  providerReference?: string;
  receiptNumber?: string;
  fees?: {
    amount: number;
    currency: string;
    description: string;
  };
  exchangeRate?: {
    from: string;
    to: string;
    rate: number;
  };
  settlementInfo?: {
    status: 'PENDING' | 'PROCESSING' | 'COMPLETED' | 'FAILED';
    estimatedSettlementDate?: Date;
    settlementId?: string;
  };
}

/**
 * Interface para cancelar uma transação
 */
export interface CancelTransactionInput {
  transactionId: string;
  tenantId: string;
  reason?: string;
  deviceInfo?: DeviceInfo;
}

/**
 * Interface para saída de cancelamento
 */
export interface CancelTransactionOutput {
  transactionId: string;
  cancelled: boolean;
  status: TransactionStatus;
  failureReason?: FailureReason;
  refundInitiated?: boolean;
  refundTransactionId?: string;
}

/**
 * Interface para histórico de transações
 */
export interface TransactionHistoryInput {
  userId: string;
  tenantId: string;
  phoneNumber?: string;
  startDate?: Date;
  endDate?: Date;
  status?: TransactionStatus[];
  type?: TransactionType[];
  provider?: MobileMoneyProvider[];
  minAmount?: number;
  maxAmount?: number;
  currency?: string;
  limit?: number;
  offset?: number;
  sortBy?: string;
  sortDirection?: 'ASC' | 'DESC';
}

/**
 * Interface para transação completa
 */
export interface Transaction {
  id: string;
  userId: string;
  tenantId: string;
  type: TransactionType;
  status: TransactionStatus;
  amount: number;
  currency: string;
  totalAmount: number;
  fees?: {
    amount: number;
    currency: string;
    description: string;
  };
  provider: MobileMoneyProvider;
  phoneNumber: string;
  recipient?: {
    phoneNumber?: string;
    accountId?: string;
    name?: string;
    bankCode?: string;
  };
  description?: string;
  referenceId?: string;
  providerReferenceId?: string;
  metadata?: Record<string, any>;
  purposeCode?: string;
  consentId?: string;
  createdAt: Date;
  updatedAt: Date;
  completedAt?: Date;
  expiresAt?: Date;
  failureReason?: FailureReason;
  failureDescription?: string;
  deviceInfo?: DeviceInfo;
  riskData?: RiskData;
  complianceData?: ComplianceData;
  regionalData?: Record<string, any>;
}

/**
 * Interface para limites de transação
 */
export interface TransactionLimits {
  singleTransactionLimit: number;
  dailyLimit: number;
  monthlyLimit: number;
  remainingDailyLimit: number;
  remainingMonthlyLimit: number;
  currency: string;
}

/**
 * Interface para verificação de elegibilidade
 */
export interface EligibilityCheckInput {
  userId: string;
  tenantId: string;
  phoneNumber: string;
  transactionType?: TransactionType;
  amount?: number;
  currency?: string;
  provider?: MobileMoneyProvider;
}

/**
 * Interface para saída de verificação de elegibilidade
 */
export interface EligibilityCheckOutput {
  eligible: boolean;
  services: TransactionType[];
  limits: TransactionLimits;
  kycRequired: boolean;
  requiredDocuments?: string[];
  requiresUpgrade: boolean;
  upgradeInstructions?: string;
  message?: string;
  providers: MobileMoneyProvider[];
}

/**
 * Interface para eventos de transação para assinaturas GraphQL
 */
export interface TransactionEvent {
  transactionId: string;
  tenantId: string;
  userId: string;
  status: TransactionStatus;
  type: TransactionType;
  timestamp: Date;
  description: string;
  amount?: number;
  currency?: string;
  provider?: MobileMoneyProvider;
}

/**
 * Interface para serviço de processamento de transações Mobile Money
 */
export interface MobileMoneyTransactionService {
  initiateTransaction(input: InitiateTransactionInput): Promise<InitiateTransactionOutput>;
  verifyOTP(input: VerifyOTPInput): Promise<VerifyOTPOutput>;
  checkTransactionStatus(input: CheckTransactionStatusInput): Promise<CheckTransactionStatusOutput>;
  cancelTransaction(input: CancelTransactionInput): Promise<CancelTransactionOutput>;
  getTransactionHistory(input: TransactionHistoryInput): Promise<{
    transactions: Transaction[];
    totalCount: number;
    hasMore: boolean;
  }>;
  getTransactionById(id: string, tenantId: string): Promise<Transaction>;
  checkEligibility(input: EligibilityCheckInput): Promise<EligibilityCheckOutput>;
  registerTransactionEvent(event: TransactionEvent): Promise<void>;
  getTransactionLimits(userId: string, tenantId: string, currency: string): Promise<TransactionLimits>;
  refreshProviderConfigurations(): Promise<void>;
}