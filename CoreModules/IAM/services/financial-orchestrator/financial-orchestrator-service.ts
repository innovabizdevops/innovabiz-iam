/**
 * Serviço orquestrador de serviços financeiros
 * 
 * Este middleware coordena as interações entre os diferentes serviços financeiros
 * da plataforma INNOVABIZ, como Mobile Money, E-Commerce, Bureau de Crédito, etc.
 */

import { v4 as uuidv4 } from 'uuid';
import { Logger } from '../../observability/logging/hook_logger';
import { Metrics } from '../../observability/metrics/hook_metrics';
import { Tracer } from '../../observability/tracing/hook_tracing';
import { 
  FinancialOperationRequest,
  FinancialOperationResult,
  FinancialOperationHistory,
  AvailableFinancialServices,
  FinancialServiceType,
  FinancialOperationType,
  OrchestrationStatus,
  OperationPriority
} from './types';
import { MobileMoneyProvider, TransactionStatus, TransactionType } from '../mobile-money/types';

/**
 * Implementação do serviço orquestrador de serviços financeiros
 */
export class FinancialOrchestratorServiceImpl {
  private readonly logger: Logger;
  private readonly metrics: Metrics;
  private readonly tracer: Tracer;
  private readonly databaseClient: any;
  private readonly complianceService: any;
  private readonly riskService: any;
  private readonly configService: any;
  private readonly eventBus: any;
  private readonly cacheService: any;
  private readonly iamService: any;
  
  // Serviços específicos a serem orquestrados
  private readonly mobileMoneyService: any;
  private readonly ecommerceService: any;
  private readonly bureauCreditService: any;
  
  constructor(
    logger: Logger,
    metrics: Metrics,
    tracer: Tracer,
    databaseClient: any,
    complianceService: any,
    riskService: any,
    configService: any,
    eventBus: any,
    cacheService: any,
    iamService: any,
    mobileMoneyService: any,
    ecommerceService: any,
    bureauCreditService: any
  ) {
    this.logger = logger;
    this.metrics = metrics;
    this.tracer = tracer;
    this.databaseClient = databaseClient;
    this.complianceService = complianceService;
    this.riskService = riskService;
    this.configService = configService;
    this.eventBus = eventBus;
    this.cacheService = cacheService;
    this.iamService = iamService;
    this.mobileMoneyService = mobileMoneyService;
    this.ecommerceService = ecommerceService;
    this.bureauCreditService = bureauCreditService;
  }
  
  /**
   * Inicia uma operação financeira
   */
  async initiateOperation(request: FinancialOperationRequest): Promise<FinancialOperationResult> {
    const span = this.tracer.startSpan('financial_orchestrator.initiate_operation');
    const operationId = uuidv4();
    
    try {
      this.logger.info('FinancialOrchestratorService: Iniciando operação financeira', {
        operationType: request.operationType,
        serviceType: request.serviceType,
        userId: request.userId,
        tenantId: request.tenantId,
        operationId,
        amount: request.amount,
        currency: request.currency
      });
      
      // Registrar início da operação em métricas
      this.metrics.increment('financial_orchestrator.operation.initiated', {
        operationType: request.operationType,
        serviceType: request.serviceType,
        tenantId: request.tenantId
      });
      
      // 1. Verificar permissões do usuário
      const hasPermission = await this.checkUserPermissions(
        request.userId,
        request.tenantId,
        request.serviceType,
        request.operationType
      );
      
      if (!hasPermission) {
        throw new Error('Usuário não possui permissão para realizar esta operação');
      }
      
      // 2. Verificar conformidade regulatória
      const complianceResult = await this.checkCompliance(request);
      
      if (!complianceResult.compliant) {
        return {
          operationId,
          status: OrchestrationStatus.FAILED,
          serviceType: request.serviceType,
          userId: request.userId,
          tenantId: request.tenantId,
          timestamp: new Date(),
          failureReason: `Não conformidade regulatória: ${complianceResult.reason}`
        };
      }
      
      // 3. Avaliar risco da operação
      const riskResult = await this.evaluateRisk(request);
      
      if (riskResult.automaticallyDeclined) {
        return {
          operationId,
          status: OrchestrationStatus.FAILED,
          serviceType: request.serviceType,
          userId: request.userId,
          tenantId: request.tenantId,
          timestamp: new Date(),
          failureReason: `Operação recusada pelo sistema de risco: ${riskResult.reason}`
        };
      }
      
      if (riskResult.requiresApproval) {
        // Registrar operação pendente de aprovação no banco de dados
        await this.databaseClient.financialOperation.create({
          id: operationId,
          userId: request.userId,
          tenantId: request.tenantId,
          serviceType: request.serviceType,
          operationType: request.operationType,
          status: OrchestrationStatus.PENDING_APPROVAL,
          amount: request.amount,
          currency: request.currency,
          provider: request.provider,
          referenceId: request.referenceId,
          metadata: request.metadata,
          riskScore: riskResult.score,
          riskLevel: riskResult.level,
          requiredApprovals: request.requiredApprovals || 1,
          currentApprovals: 0
        });
        
        // Publicar evento de operação pendente de aprovação
        this.eventBus.publish('financial.operation.pending_approval', {
          operationId,
          userId: request.userId,
          tenantId: request.tenantId,
          serviceType: request.serviceType,
          operationType: request.operationType,
          amount: request.amount,
          currency: request.currency,
          riskLevel: riskResult.level,
          requiredApprovals: request.requiredApprovals || 1
        });
        
        return {
          operationId,
          status: OrchestrationStatus.PENDING_APPROVAL,
          serviceType: request.serviceType,
          userId: request.userId,
          tenantId: request.tenantId,
          timestamp: new Date(),
          amount: request.amount,
          currency: request.currency,
          metadata: request.metadata
        };
      }
      
      // 4. Delegar para o serviço específico
      const result = await this.delegateToService(operationId, request);
      
      // 5. Registrar operação no banco de dados
      await this.databaseClient.financialOperation.create({
        id: operationId,
        userId: request.userId,
        tenantId: request.tenantId,
        serviceType: request.serviceType,
        operationType: request.operationType,
        status: result.status,
        amount: request.amount,
        currency: request.currency,
        provider: request.provider,
        transactionId: result.transactionId,
        referenceId: request.referenceId || result.referenceId,
        providerReference: result.providerReference,
        metadata: {
          ...request.metadata,
          ...result.metadata
        },
        riskScore: riskResult.score,
        riskLevel: riskResult.level
      });
      
      // 6. Publicar evento de operação
      this.eventBus.publish('financial.operation.initiated', {
        operationId,
        userId: request.userId,
        tenantId: request.tenantId,
        serviceType: request.serviceType,
        operationType: request.operationType,
        status: result.status,
        transactionId: result.transactionId
      });
      
      this.logger.info('FinancialOrchestratorService: Operação financeira iniciada com sucesso', {
        operationId,
        status: result.status
      });
      
      return result;
    } catch (error) {
      this.logger.error('FinancialOrchestratorService: Erro ao iniciar operação financeira', {
        operationId,
        error,
        userId: request.userId,
        tenantId: request.tenantId,
        serviceType: request.serviceType,
        operationType: request.operationType
      });
      
      // Registrar falha em métricas
      this.metrics.increment('financial_orchestrator.operation.failed', {
        operationType: request.operationType,
        serviceType: request.serviceType,
        tenantId: request.tenantId,
        reason: error.name
      });
      
      // Registrar operação falha no banco de dados
      try {
        await this.databaseClient.financialOperation.create({
          id: operationId,
          userId: request.userId,
          tenantId: request.tenantId,
          serviceType: request.serviceType,
          operationType: request.operationType,
          status: OrchestrationStatus.FAILED,
          amount: request.amount,
          currency: request.currency,
          provider: request.provider,
          referenceId: request.referenceId,
          failureReason: error.message
        });
      } catch (dbError) {
        // Log em caso de falha no banco de dados
        this.logger.error('FinancialOrchestratorService: Erro ao registrar falha no banco de dados', { dbError });
      }
      
      // Publicar evento de operação falha
      this.eventBus.publish('financial.operation.failed', {
        operationId,
        userId: request.userId,
        tenantId: request.tenantId,
        serviceType: request.serviceType,
        operationType: request.operationType,
        error: error.message
      });
      
      throw new Error(`Erro ao iniciar operação financeira: ${error.message}`);
    } finally {
      span.end();
    }
  }
  
  /**
   * Verifica o status de uma operação financeira
   */
  async checkOperationStatus(operationId: string, tenantId: string): Promise<FinancialOperationResult> {
    const span = this.tracer.startSpan('financial_orchestrator.check_operation_status');
    
    try {
      this.logger.info('FinancialOrchestratorService: Verificando status de operação', {
        operationId,
        tenantId
      });
      
      // Buscar operação no banco de dados
      const operation = await this.databaseClient.financialOperation.findById(operationId);
      
      if (!operation) {
        throw new Error('Operação não encontrada');
      }
      
      // Verificar se o tenant corresponde
      if (operation.tenantId !== tenantId) {
        throw new Error('Acesso não autorizado a esta operação');
      }
      
      // Para operações ainda em andamento, verificar status atual no serviço específico
      if (this.isOperationInProgress(operation.status)) {
        const currentStatus = await this.getServiceStatus(
          operationId, 
          operation.serviceType,
          operation.transactionId
        );
        
        // Atualizar status no banco de dados se mudou
        if (currentStatus.status !== operation.status) {
          await this.databaseClient.financialOperation.update(operationId, {
            status: currentStatus.status,
            completedAt: this.isOperationCompleted(currentStatus.status) ? new Date() : undefined,
            providerReference: currentStatus.providerReference || operation.providerReference,
            failureReason: currentStatus.failureReason || operation.failureReason,
            metadata: {
              ...operation.metadata,
              ...currentStatus.metadata
            },
            updatedAt: new Date()
          });
          
          // Publicar evento de mudança de status
          this.eventBus.publish('financial.operation.status_changed', {
            operationId,
            userId: operation.userId,
            tenantId: operation.tenantId,
            previousStatus: operation.status,
            currentStatus: currentStatus.status
          });
          
          // Retornar status atualizado
          return {
            operationId,
            status: currentStatus.status,
            serviceType: operation.serviceType,
            transactionId: operation.transactionId,
            referenceId: operation.referenceId,
            providerReference: currentStatus.providerReference || operation.providerReference,
            amount: operation.amount,
            currency: operation.currency,
            timestamp: operation.createdAt,
            completedAt: this.isOperationCompleted(currentStatus.status) ? new Date() : undefined,
            userId: operation.userId,
            tenantId: operation.tenantId,
            requiredAction: currentStatus.requiredAction,
            receiptUrl: currentStatus.receiptUrl || operation.receiptUrl,
            failureReason: currentStatus.failureReason || operation.failureReason,
            metadata: {
              ...operation.metadata,
              ...currentStatus.metadata
            }
          };
        }
      }
      
      // Retornar dados da operação
      return {
        operationId,
        status: operation.status,
        serviceType: operation.serviceType,
        transactionId: operation.transactionId,
        referenceId: operation.referenceId,
        providerReference: operation.providerReference,
        amount: operation.amount,
        currency: operation.currency,
        timestamp: operation.createdAt,
        completedAt: operation.completedAt,
        userId: operation.userId,
        tenantId: operation.tenantId,
        requiredAction: operation.requiredAction,
        receiptUrl: operation.receiptUrl,
        failureReason: operation.failureReason,
        metadata: operation.metadata
      };
    } catch (error) {
      this.logger.error('FinancialOrchestratorService: Erro ao verificar status de operação', {
        operationId,
        tenantId,
        error
      });
      
      throw new Error(`Erro ao verificar status de operação: ${error.message}`);
    } finally {
      span.end();
    }
  }
  
  /**
   * Cancela uma operação financeira em andamento
   */
  async cancelOperation(operationId: string, tenantId: string, reason?: string): Promise<FinancialOperationResult> {
    const span = this.tracer.startSpan('financial_orchestrator.cancel_operation');
    
    try {
      this.logger.info('FinancialOrchestratorService: Cancelando operação financeira', {
        operationId,
        tenantId,
        reason
      });
      
      // Buscar operação no banco de dados
      const operation = await this.databaseClient.financialOperation.findById(operationId);
      
      if (!operation) {
        throw new Error('Operação não encontrada');
      }
      
      // Verificar se o tenant corresponde
      if (operation.tenantId !== tenantId) {
        throw new Error('Acesso não autorizado a esta operação');
      }
      
      // Verificar se operação pode ser cancelada
      if (!this.isOperationCancellable(operation.status)) {
        throw new Error(`Operação não pode ser cancelada no status atual: ${operation.status}`);
      }
      
      // Cancelar no serviço específico
      const cancelResult = await this.cancelServiceOperation(
        operationId,
        operation.serviceType,
        operation.transactionId,
        reason
      );
      
      // Atualizar no banco de dados
      await this.databaseClient.financialOperation.update(operationId, {
        status: OrchestrationStatus.CANCELLED,
        failureReason: reason || 'Cancelado pelo usuário',
        updatedAt: new Date()
      });
      
      // Publicar evento de cancelamento
      this.eventBus.publish('financial.operation.cancelled', {
        operationId,
        userId: operation.userId,
        tenantId: operation.tenantId,
        serviceType: operation.serviceType,
        operationType: operation.operationType,
        reason: reason || 'Cancelado pelo usuário'
      });
      
      return {
        operationId,
        status: OrchestrationStatus.CANCELLED,
        serviceType: operation.serviceType,
        transactionId: operation.transactionId,
        referenceId: operation.referenceId,
        providerReference: operation.providerReference,
        amount: operation.amount,
        currency: operation.currency,
        timestamp: operation.createdAt,
        completedAt: new Date(),
        userId: operation.userId,
        tenantId: operation.tenantId,
        failureReason: reason || 'Cancelado pelo usuário',
        metadata: operation.metadata
      };
    } catch (error) {
      this.logger.error('FinancialOrchestratorService: Erro ao cancelar operação', {
        operationId,
        tenantId,
        error
      });
      
      throw new Error(`Erro ao cancelar operação: ${error.message}`);
    } finally {
      span.end();
    }
  }
  
  /**
   * Inicia um pagamento via Mobile Money
   */
  async initiatePayment(params: {
    userId: string;
    tenantId: string;
    provider: MobileMoneyProvider;
    amount: number;
    currency: string;
    phoneNumber: string;
    description?: string;
    metadata?: Record<string, any>;
    deviceInfo?: any;
  }): Promise<FinancialOperationResult> {
    // Converter para formato genérico de operação e chamar initiateOperation
    const request: FinancialOperationRequest = {
      operationType: FinancialOperationType.PAYMENT,
      serviceType: FinancialServiceType.MOBILE_MONEY,
      provider: params.provider,
      amount: params.amount,
      currency: params.currency,
      userId: params.userId,
      tenantId: params.tenantId,
      metadata: {
        ...params.metadata,
        phoneNumber: params.phoneNumber,
        description: params.description
      },
      deviceInfo: params.deviceInfo,
      priority: OperationPriority.MEDIUM
    };
    
    return this.initiateOperation(request);
  }
  
  /**
   * Verifica OTP de uma transação
   */
  async verifyTransactionOtp(params: {
    operationId: string;
    tenantId: string;
    otpCode: string;
    deviceInfo?: any;
  }): Promise<FinancialOperationResult> {
    const span = this.tracer.startSpan('financial_orchestrator.verify_transaction_otp');
    
    try {
      this.logger.info('FinancialOrchestratorService: Verificando OTP de transação', {
        operationId: params.operationId,
        tenantId: params.tenantId
      });
      
      // Buscar operação no banco de dados
      const operation = await this.databaseClient.financialOperation.findById(params.operationId);
      
      if (!operation) {
        throw new Error('Operação não encontrada');
      }
      
      // Verificar se o tenant corresponde
      if (operation.tenantId !== params.tenantId) {
        throw new Error('Acesso não autorizado a esta operação');
      }
      
      // Verificar se a operação é do tipo Mobile Money
      if (operation.serviceType !== FinancialServiceType.MOBILE_MONEY) {
        throw new Error('Esta operação não suporta verificação de OTP');
      }
      
      // Verificar OTP no serviço Mobile Money
      const verifyResult = await this.mobileMoneyService.verifyOTP({
        transactionId: operation.transactionId,
        tenantId: params.tenantId,
        otpCode: params.otpCode,
        deviceInfo: params.deviceInfo
      });
      
      // Atualizar status no banco de dados
      const newStatus = verifyResult.verified 
        ? OrchestrationStatus.PROCESSING 
        : OrchestrationStatus.PENDING_USER_ACTION;
      
      await this.databaseClient.financialOperation.update(params.operationId, {
        status: newStatus,
        updatedAt: new Date(),
        metadata: {
          ...operation.metadata,
          otpVerified: verifyResult.verified,
          otpVerificationAttempts: (operation.metadata?.otpVerificationAttempts || 0) + 1
        }
      });
      
      // Publicar evento de OTP verificado
      this.eventBus.publish('financial.operation.otp_verified', {
        operationId: params.operationId,
        userId: operation.userId,
        tenantId: params.tenantId,
        verified: verifyResult.verified
      });
      
      return {
        operationId: params.operationId,
        status: newStatus,
        serviceType: operation.serviceType,
        transactionId: operation.transactionId,
        referenceId: operation.referenceId,
        providerReference: operation.providerReference,
        amount: operation.amount,
        currency: operation.currency,
        timestamp: operation.createdAt,
        userId: operation.userId,
        tenantId: operation.tenantId,
        requiredAction: !verifyResult.verified ? {
          type: 'OTP_VERIFICATION',
          instructions: 'Código OTP inválido. Tente novamente.',
          otpRequired: true,
          otpSent: false,
          timeoutSeconds: 300
        } : undefined,
        failureReason: !verifyResult.verified ? verifyResult.failureReason : undefined,
        metadata: {
          ...operation.metadata,
          otpVerified: verifyResult.verified,
          remainingAttempts: verifyResult.remainingAttempts
        }
      };
    } catch (error) {
      this.logger.error('FinancialOrchestratorService: Erro ao verificar OTP', {
        operationId: params.operationId,
        tenantId: params.tenantId,
        error
      });
      
      throw new Error(`Erro ao verificar OTP: ${error.message}`);
    } finally {
      span.end();
    }
  }
}