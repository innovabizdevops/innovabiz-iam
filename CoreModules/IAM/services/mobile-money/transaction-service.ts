/**
 * Serviço de processamento de transações Mobile Money
 * 
 * Este serviço gerencia o ciclo de vida completo das transações Mobile Money,
 * incluindo validação, compliance regional, processamento, gerenciamento de OTP,
 * e integração com provedores externos.
 */

import { v4 as uuidv4 } from 'uuid';
import { Logger } from '../../observability/logging/hook_logger';
import { Metrics } from '../../observability/metrics/hook_metrics';
import { Tracer } from '../../observability/tracing/hook_tracing';

import { 
  MobileMoneyTransactionService,
  InitiateTransactionInput, 
  InitiateTransactionOutput,
  VerifyOTPInput,
  VerifyOTPOutput,
  CheckTransactionStatusInput,
  CheckTransactionStatusOutput,
  CancelTransactionInput,
  CancelTransactionOutput,
  TransactionHistoryInput,
  Transaction,
  TransactionLimits,
  EligibilityCheckInput,
  EligibilityCheckOutput,
  TransactionEvent,
  TransactionStatus,
  FailureReason,
  MobileMoneyProvider,
  ProviderConfig,
  TransactionType
} from './types';

import { RegionalComplianceService } from '../../infrastructure/compliance/regional_compliance_service';
import { RiskEnrichmentService } from '../../infrastructure/adaptive/risk_engine';
import { DatabaseClient } from '../../infrastructure/common/database_client';
import { ConfigService } from '../../infrastructure/common/config_service';
import { EventBus } from '../../infrastructure/common/event_bus';
import { CacheService } from '../../infrastructure/common/cache_service';
import { IamService } from '../../services/iam-service';
import { MobileMoneyProviderFactory } from './provider-factory';

/**
 * Implementação do serviço de processamento de transações Mobile Money
 */
export class MobileMoneyTransactionServiceImpl implements MobileMoneyTransactionService {
  private readonly logger: Logger;
  private readonly metrics: Metrics;
  private readonly tracer: Tracer;
  private readonly db: DatabaseClient;
  private readonly regionalComplianceService: RegionalComplianceService;
  private readonly riskService: RiskEnrichmentService;
  private readonly configService: ConfigService;
  private readonly eventBus: EventBus;
  private readonly cacheService: CacheService;
  private readonly iamService: IamService;
  private readonly providerFactory: MobileMoneyProviderFactory;
  private providerConfigurations: Map<string, ProviderConfig> = new Map();
  
  constructor(
    logger: Logger,
    metrics: Metrics,
    tracer: Tracer,
    db: DatabaseClient,
    regionalComplianceService: RegionalComplianceService,
    riskService: RiskEnrichmentService,
    configService: ConfigService,
    eventBus: EventBus,
    cacheService: CacheService,
    iamService: IamService,
    providerFactory: MobileMoneyProviderFactory
  ) {
    this.logger = logger;
    this.metrics = metrics;
    this.tracer = tracer;
    this.db = db;
    this.regionalComplianceService = regionalComplianceService;
    this.riskService = riskService;
    this.configService = configService;
    this.eventBus = eventBus;
    this.cacheService = cacheService;
    this.iamService = iamService;
    this.providerFactory = providerFactory;
    
    // Inicializar configurações de provedores
    this.refreshProviderConfigurations();
  }
  
  /**
   * Carrega as configurações dos provedores de Mobile Money
   */
  async refreshProviderConfigurations(): Promise<void> {
    const span = this.tracer.startSpan('MobileMoneyService.refreshProviderConfigurations');
    
    try {
      this.logger.info('Refreshing Mobile Money provider configurations');
      
      const configs = await this.configService.getProviderConfigurations('mobile-money');
      
      this.providerConfigurations.clear();
      configs.forEach(config => {
        this.providerConfigurations.set(config.providerId, config as ProviderConfig);
      });
      
      this.logger.info(`Loaded ${this.providerConfigurations.size} provider configurations`);
      this.metrics.gauge('mobile_money.providers.count', this.providerConfigurations.size);
    } catch (error) {
      this.logger.error('Failed to refresh provider configurations', { error });
      throw error;
    } finally {
      span.end();
    }
  }

  /**
   * Inicia uma nova transação Mobile Money
   */
  async initiateTransaction(input: InitiateTransactionInput): Promise<InitiateTransactionOutput> {
    const span = this.tracer.startSpan('MobileMoneyService.initiateTransaction');
    const { tenantId, userId } = input;
    
    try {
      this.logger.info('Initiating Mobile Money transaction', { 
        tenantId, 
        userId, 
        type: input.type,
        provider: input.provider,
        amount: input.amount,
        currency: input.currency
      });
      
      this.metrics.increment('mobile_money.transaction.initiate', { tenantId, provider: input.provider, type: input.type });
      
      // Verificar permissões do usuário
      await this.verifyPermissions(userId, tenantId, 'mobile_money:initiate_transaction');
      
      // Validar os limites de transação
      await this.validateTransactionLimits(input);
      
      // Avaliar risco da transação
      const riskAssessment = await this.assessTransactionRisk(input);
      
      // Verificar conformidade regional
      await this.validateRegionalCompliance(input);
      
      // Gerar ID da transação
      const transactionId = uuidv4();
      
      // Calcular taxas
      const fees = await this.calculateTransactionFees(input);
      
      // Calcular valor total
      const totalAmount = input.amount + (fees?.amount || 0);
      
      // Obter provedor específico para o serviço
      const providerInstance = this.providerFactory.getProvider(input.provider);
      
      // Iniciar transação com o provedor
      const providerResponse = await providerInstance.initiateTransaction({
        transactionId,
        amount: input.amount,
        currency: input.currency,
        phoneNumber: input.phoneNumber,
        recipientPhone: input.recipient?.phoneNumber,
        recipientName: input.recipient?.name,
        description: input.description,
        type: input.type,
        metadata: input.metadata,
        notifyRecipient: input.notifyRecipient
      });
      
      // Registrar a transação no banco de dados
      const transaction: Transaction = {
        id: transactionId,
        userId: input.userId,
        tenantId: input.tenantId,
        type: input.type,
        status: TransactionStatus.INITIATED,
        amount: input.amount,
        currency: input.currency,
        totalAmount,
        fees: fees,
        provider: input.provider,
        phoneNumber: input.phoneNumber,
        recipient: input.recipient,
        description: input.description,
        referenceId: input.referenceId || providerResponse.referenceNumber,
        providerReferenceId: providerResponse.providerReference,
        metadata: input.metadata,
        purposeCode: input.purposeCode,
        consentId: input.consentId,
        createdAt: new Date(),
        updatedAt: new Date(),
        expiresAt: new Date(Date.now() + (input.expiresInSeconds || 900) * 1000),
        deviceInfo: input.deviceInfo,
        riskData: riskAssessment,
        complianceData: {
          kycStatus: providerResponse.kycStatus || 'VERIFIED',
          kycLevel: 'FULL',
          consentIds: input.consentId ? [input.consentId] : [],
          purposeCode: input.purposeCode,
          sanctionScreeningPassed: true,
          amlChecksPassed: true,
          regulatoryRequirementsMet: true
        },
        regionalData: input.regionalData
      };
      
      await this.db.transaction.create(transaction);
      
      // Publicar evento de transação iniciada
      await this.publishTransactionEvent({
        transactionId,
        tenantId: input.tenantId,
        userId: input.userId,
        status: TransactionStatus.INITIATED,
        type: input.type,
        timestamp: new Date(),
        description: `Transação ${input.type} iniciada`,
        amount: input.amount,
        currency: input.currency,
        provider: input.provider
      });
      
      // Preparar resposta
      const response: InitiateTransactionOutput = {
        transactionId,
        status: TransactionStatus.INITIATED,
        referenceNumber: providerResponse.referenceNumber,
        otpRequired: providerResponse.otpRequired || false,
        otpSent: providerResponse.otpSent || false,
        otpPhoneNumber: providerResponse.otpPhoneNumber,
        expiresAt: transaction.expiresAt,
        processingEstimateSeconds: 30,
        fees,
        totalAmount,
        requiredAction: providerResponse.requiredAction,
        riskAssessment: {
          level: riskAssessment.level,
          requiresAdditionalVerification: riskAssessment.requiresApproval
        }
      };
      
      this.logger.info('Successfully initiated Mobile Money transaction', { 
        transactionId, 
        status: response.status, 
        otpRequired: response.otpRequired 
      });
      
      return response;
    } catch (error) {
      this.logger.error('Failed to initiate Mobile Money transaction', { 
        error, 
        tenantId, 
        userId,
        provider: input.provider
      });
      
      this.metrics.increment('mobile_money.transaction.initiate.error', { 
        tenantId, 
        provider: input.provider, 
        error: error.name
      });
      
      throw error;
    } finally {
      span.end();
    }
  }