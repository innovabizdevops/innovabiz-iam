/**
 * Funções específicas para integrações com E-Commerce no orquestrador financeiro
 */

import { FinancialOrchestratorHelpers } from './financial-orchestrator-service-helpers';
import { 
  FinancialServiceType, 
  FinancialOperationType, 
  OrchestrationStatus, 
  FinancialOperationResult
} from './types';

/**
 * Extensão do FinancialOrchestratorServiceImpl com funções específicas para E-Commerce
 */
export class FinancialOrchestratorEcommerceProcessor {
  private readonly logger: any;
  private readonly metrics: any;
  private readonly tracer: any;
  private readonly ecommerceService: any;
  private readonly mobileMoneyService: any;
  private readonly paymentGatewayService: any;
  private readonly databaseClient: any;
  private readonly eventBus: any;
  private readonly helpers: any;
  
  constructor(
    logger: any,
    metrics: any,
    tracer: any,
    ecommerceService: any,
    mobileMoneyService: any,
    paymentGatewayService: any,
    databaseClient: any,
    eventBus: any
  ) {
    this.logger = logger;
    this.metrics = metrics;
    this.tracer = tracer;
    this.ecommerceService = ecommerceService;
    this.mobileMoneyService = mobileMoneyService;
    this.paymentGatewayService = paymentGatewayService;
    this.databaseClient = databaseClient;
    this.eventBus = eventBus;
    this.helpers = FinancialOrchestratorHelpers;
  }
  
  /**
   * Processa um checkout do E-Commerce usando diferentes métodos de pagamento
   */
  async processCheckout(params: {
    operationId: string;
    userId: string;
    tenantId: string;
    shopId: string;
    checkoutSessionId: string;
    paymentMethod: string;
    deviceInfo?: any;
  }): Promise<FinancialOperationResult> {
    const span = this.tracer.startSpan('financial_orchestrator.process_checkout');
    
    try {
      this.logger.info('FinancialOrchestratorService: Processando checkout de E-Commerce', {
        operationId: params.operationId,
        userId: params.userId,
        tenantId: params.tenantId,
        shopId: params.shopId,
        checkoutSessionId: params.checkoutSessionId,
        paymentMethod: params.paymentMethod
      });
      
      // Registrar em métricas
      this.metrics.increment('financial_orchestrator.checkout.initiated', {
        tenantId: params.tenantId,
        paymentMethod: params.paymentMethod
      });
      
      // 1. Validar sessão de checkout
      const sessionValidation = await this.ecommerceService.validateCheckoutSession(
        params.checkoutSessionId,
        params.tenantId
      );
      
      if (!sessionValidation.valid) {
        throw new Error('Sessão de checkout inválida ou expirada');
      }
      
      // 2. Verificar informações da loja
      const shopInfo = await this.ecommerceService.getShopInfo(params.shopId, params.tenantId);
      
      if (!shopInfo || shopInfo.status !== 'ACTIVE') {
        throw new Error('Loja não encontrada ou inativa');
      }
      
      // 3. Verificar se o método de pagamento é aceito pela loja
      if (!shopInfo.settings?.paymentMethods?.includes(params.paymentMethod)) {
        throw new Error(`Método de pagamento "${params.paymentMethod}" não é aceito por esta loja`);
      }
      
      // 4. Iniciar pagamento baseado no método escolhido
      let paymentResult;
      
      switch (params.paymentMethod) {
        case 'MOBILE_MONEY':
          paymentResult = await this.processMobileMoneyCheckout(
            params.operationId,
            params.userId,
            params.tenantId,
            params.checkoutSessionId,
            sessionValidation,
            params.deviceInfo
          );
          break;
        
        case 'CREDIT_CARD':
        case 'DEBIT_CARD':
          paymentResult = await this.processCardCheckout(
            params.operationId,
            params.userId,
            params.tenantId,
            params.checkoutSessionId,
            sessionValidation,
            params.deviceInfo
          );
          break;
          
        case 'BANK_TRANSFER':
          paymentResult = await this.processBankTransferCheckout(
            params.operationId,
            params.userId,
            params.tenantId,
            params.checkoutSessionId,
            sessionValidation
          );
          break;
          
        default:
          throw new Error(`Método de pagamento "${params.paymentMethod}" não suportado`);
      }
      
      // 5. Atualizar o status da operação no banco de dados
      await this.databaseClient.financialOperation.update(params.operationId, {
        status: paymentResult.status,
        transactionId: paymentResult.transactionId,
        providerReference: paymentResult.providerReference,
        metadata: {
          checkoutSessionId: params.checkoutSessionId,
          shopId: params.shopId,
          paymentMethod: params.paymentMethod,
          ...paymentResult.metadata
        },
        updatedAt: new Date()
      });
      
      // 6. Notificar E-Commerce sobre o pagamento (para casos de sucesso ou em processamento)
      if ([OrchestrationStatus.COMPLETED, OrchestrationStatus.PROCESSING].includes(paymentResult.status)) {
        await this.ecommerceService.notifyPaymentStatus({
          checkoutSessionId: params.checkoutSessionId,
          tenantId: params.tenantId,
          status: paymentResult.status === OrchestrationStatus.COMPLETED ? 'PAID' : 'PROCESSING',
          transactionId: paymentResult.transactionId,
          operationId: params.operationId
        });
      }
      
      // 7. Publicar evento de checkout processado
      this.eventBus.publish('financial.ecommerce.checkout_processed', {
        operationId: params.operationId,
        checkoutSessionId: params.checkoutSessionId,
        userId: params.userId,
        tenantId: params.tenantId,
        shopId: params.shopId,
        paymentMethod: params.paymentMethod,
        status: paymentResult.status,
        amount: sessionValidation.amount,
        currency: sessionValidation.currency
      });
      
      this.logger.info('FinancialOrchestratorService: Checkout de E-Commerce processado', {
        operationId: params.operationId,
        status: paymentResult.status
      });
      
      return paymentResult;
    } catch (error) {
      this.logger.error('FinancialOrchestratorService: Erro ao processar checkout de E-Commerce', {
        operationId: params.operationId,
        error,
        userId: params.userId,
        tenantId: params.tenantId
      });
      
      // Registrar falha em métricas
      this.metrics.increment('financial_orchestrator.checkout.failed', {
        tenantId: params.tenantId,
        paymentMethod: params.paymentMethod,
        reason: error.name
      });
      
      // Atualizar status da operação no banco de dados
      try {
        await this.databaseClient.financialOperation.update(params.operationId, {
          status: OrchestrationStatus.FAILED,
          failureReason: error.message,
          updatedAt: new Date()
        });
      } catch (dbError) {
        this.logger.error('FinancialOrchestratorService: Erro ao atualizar status da operação', { dbError });
      }
      
      // Publicar evento de falha no checkout
      this.eventBus.publish('financial.ecommerce.checkout_failed', {
        operationId: params.operationId,
        userId: params.userId,
        tenantId: params.tenantId,
        error: error.message
      });
      
      throw new Error(`Erro ao processar checkout: ${error.message}`);
    } finally {
      span.end();
    }
  }
  
  /**
   * Processa checkout usando Mobile Money
   */
  private async processMobileMoneyCheckout(
    operationId: string,
    userId: string,
    tenantId: string,
    checkoutSessionId: string,
    sessionValidation: any,
    deviceInfo?: any
  ): Promise<FinancialOperationResult> {
    const span = this.tracer.startSpan('financial_orchestrator.process_mobile_money_checkout');
    
    try {
      // Obter informações do usuário para Mobile Money
      const userInfo = await this.databaseClient.user.findById(userId);
      
      if (!userInfo || !userInfo.mobileNumber) {
        throw new Error('Número de telefone do usuário não encontrado para pagamento via Mobile Money');
      }
      
      // Determinar provedor baseado no número de telefone ou preferência do usuário
      const provider = userInfo.preferredMobileMoneyProvider || this.detectProviderFromPhoneNumber(userInfo.mobileNumber);
      
      if (!provider) {
        throw new Error('Não foi possível determinar o provedor de Mobile Money');
      }
      
      // Iniciar pagamento Mobile Money
      const paymentResult = await this.mobileMoneyService.initiateTransaction({
        userId,
        tenantId,
        provider,
        amount: sessionValidation.amount,
        currency: sessionValidation.currency,
        phoneNumber: userInfo.mobileNumber,
        type: 'PAYMENT',
        description: `Pagamento para ${sessionValidation.merchantName || 'E-Commerce'} - Pedido #${sessionValidation.orderId || checkoutSessionId}`,
        metadata: {
          checkoutSessionId,
          orderId: sessionValidation.orderId,
          ecommerceReference: sessionValidation.reference
        },
        deviceInfo
      });
      
      return {
        operationId,
        status: OrchestrationStatus.PENDING_USER_ACTION,
        serviceType: FinancialServiceType.MOBILE_MONEY,
        transactionId: paymentResult.transactionId,
        providerReference: paymentResult.providerReference,
        amount: sessionValidation.amount,
        currency: sessionValidation.currency,
        timestamp: new Date(),
        userId,
        tenantId,
        requiredAction: {
          type: 'OTP_VERIFICATION',
          instructions: 'Verifique o código OTP enviado para seu telefone',
          otpRequired: true,
          otpSent: true,
          otpPhoneNumber: this.maskPhoneNumber(userInfo.mobileNumber),
          timeoutSeconds: 300
        },
        metadata: {
          provider,
          phoneNumber: this.maskPhoneNumber(userInfo.mobileNumber),
          checkoutSessionId,
          ...paymentResult.metadata
        }
      };
    } finally {
      span.end();
    }
  }
  
  /**
   * Processa checkout usando cartão de crédito/débito
   */
  private async processCardCheckout(
    operationId: string,
    userId: string,
    tenantId: string,
    checkoutSessionId: string,
    sessionValidation: any,
    deviceInfo?: any
  ): Promise<FinancialOperationResult> {
    const span = this.tracer.startSpan('financial_orchestrator.process_card_checkout');
    
    try {
      // Criar sessão de pagamento no gateway de pagamento
      const paymentSession = await this.paymentGatewayService.createPaymentSession({
        amount: sessionValidation.amount,
        currency: sessionValidation.currency,
        userId,
        tenantId,
        description: `Pagamento para ${sessionValidation.merchantName || 'E-Commerce'} - Pedido #${sessionValidation.orderId || checkoutSessionId}`,
        returnUrl: `${sessionValidation.returnUrl || 'https://ecommerce.innovabiz.com'}/checkout/confirm?session=${checkoutSessionId}`,
        cancelUrl: `${sessionValidation.cancelUrl || 'https://ecommerce.innovabiz.com'}/checkout/cancel?session=${checkoutSessionId}`,
        metadata: {
          checkoutSessionId,
          orderId: sessionValidation.orderId,
          ecommerceReference: sessionValidation.reference,
          operationId
        },
        deviceInfo
      });
      
      return {
        operationId,
        status: OrchestrationStatus.PENDING_USER_ACTION,
        serviceType: FinancialServiceType.PAYMENT_GATEWAY,
        transactionId: paymentSession.id,
        amount: sessionValidation.amount,
        currency: sessionValidation.currency,
        timestamp: new Date(),
        userId,
        tenantId,
        requiredAction: {
          type: 'REDIRECT',
          instructions: 'Redirecionando para a página de pagamento',
          url: paymentSession.redirectUrl,
          timeoutSeconds: 900 // 15 minutos
        },
        metadata: {
          checkoutSessionId,
          paymentSessionId: paymentSession.id,
          expiresAt: paymentSession.expiresAt
        }
      };
    } finally {
      span.end();
    }
  }
  
  /**
   * Processa checkout usando transferência bancária
   */
  private async processBankTransferCheckout(
    operationId: string,
    userId: string,
    tenantId: string,
    checkoutSessionId: string,
    sessionValidation: any
  ): Promise<FinancialOperationResult> {
    const span = this.tracer.startSpan('financial_orchestrator.process_bank_transfer_checkout');
    
    try {
      // Obter dados bancários do comerciante
      const merchantBankInfo = await this.ecommerceService.getMerchantBankInfo(
        sessionValidation.merchantId,
        tenantId
      );
      
      if (!merchantBankInfo || !merchantBankInfo.accountNumber) {
        throw new Error('Informações bancárias do comerciante não encontradas');
      }
      
      // Gerar referência única para a transferência
      const transferReference = `ECM-${Date.now().toString().substring(5)}-${operationId.substring(0, 6)}`;
      
      return {
        operationId,
        status: OrchestrationStatus.PENDING_USER_ACTION,
        serviceType: FinancialServiceType.E_COMMERCE,
        transactionId: transferReference,
        amount: sessionValidation.amount,
        currency: sessionValidation.currency,
        timestamp: new Date(),
        userId,
        tenantId,
        requiredAction: {
          type: 'BANK_TRANSFER',
          instructions: 'Por favor, realize uma transferência bancária com os dados abaixo e informe o número da transação',
          timeoutSeconds: 86400 // 24 horas
        },
        metadata: {
          checkoutSessionId,
          bankName: merchantBankInfo.bankName,
          accountName: merchantBankInfo.accountName,
          accountNumber: merchantBankInfo.accountNumber,
          reference: transferReference,
          instructions: merchantBankInfo.instructions || 'Inclua o número de referência na descrição da transferência'
        }
      };
    } finally {
      span.end();
    }
  }
  
  /**
   * Detecta o provedor de Mobile Money a partir do número de telefone
   */
  private detectProviderFromPhoneNumber(phoneNumber: string): string | null {
    // Implementação simplificada para detecção de provedor
    // Em um cenário real, isso seria mais complexo e configurável por país/região
    if (!phoneNumber) return null;
    
    const cleanNumber = phoneNumber.replace(/\D/g, '');
    
    // Códigos de exemplo para Angola
    if (cleanNumber.startsWith('923') || cleanNumber.startsWith('993')) {
      return 'MOVICEL';
    } else if (cleanNumber.startsWith('921') || cleanNumber.startsWith('991') || 
               cleanNumber.startsWith('931') || cleanNumber.startsWith('937')) {
      return 'UNITEL';
    } 
    // Outros países poderiam ser adicionados aqui
    else if (cleanNumber.startsWith('254')) {  // Quênia
      return 'MPESA';
    } else if (cleanNumber.startsWith('27')) {  // África do Sul
      return 'MTN';
    }
    
    return null;
  }
  
  /**
   * Mascara o número de telefone para exibição
   */
  private maskPhoneNumber(phoneNumber: string): string {
    if (!phoneNumber) return '***';
    
    const cleanNumber = phoneNumber.replace(/\D/g, '');
    if (cleanNumber.length <= 4) return '***' + cleanNumber;
    
    return cleanNumber.substring(0, 3) + '***' + cleanNumber.slice(-3);
  }
}