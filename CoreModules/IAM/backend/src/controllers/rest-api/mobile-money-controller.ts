/**
 * Controlador REST para Mobile Money
 * 
 * Este controlador implementa endpoints REST para compatibilidade legada
 * relacionados à integração com serviços Mobile Money.
 */

import { Request, Response } from 'express';
import { BaseRestController } from './base-controller';
import { Logger } from '../../../../observability/logging/hook_logger';
import { Metrics } from '../../../../observability/metrics/hook_metrics';
import { Tracer } from '../../../../observability/tracing/hook_tracing';
import { MobileMoneyProvider, TransactionType } from '../../../../services/mobile-money/types';

/**
 * Controlador REST para operações de Mobile Money
 */
export class MobileMoneyRestController extends BaseRestController {
  private readonly mobileMoneyService: any;
  private readonly iamService: any;
  private readonly validationService: any;
  
  /**
   * Construtor do controlador de Mobile Money
   */
  constructor(
    logger: Logger,
    metrics: Metrics,
    tracer: Tracer,
    mobileMoneyService: any,
    iamService: any,
    validationService: any
  ) {
    super(logger, metrics, tracer);
    this.mobileMoneyService = mobileMoneyService;
    this.iamService = iamService;
    this.validationService = validationService;
  }
  
  /**
   * Verifica e valida o token de autenticação
   */
  private async validateAuthToken(req: Request, res: Response): Promise<{userId: string} | null> {
    const token = this.getAuthToken(req);
    
    if (!token) {
      this.sendErrorResponse(res, 401, 'Token de autenticação ausente', 'UNAUTHORIZED');
      return null;
    }
    
    try {
      const validation = await this.iamService.validateToken(token);
      
      if (!validation.valid) {
        this.sendErrorResponse(res, 401, 'Token de autenticação inválido', 'INVALID_TOKEN');
        return null;
      }
      
      return { userId: validation.userId };
    } catch (error) {
      this.handleError(error, req, res, 'token_validation');
      return null;
    }
  }
  
  /**
   * Inicia uma transação Mobile Money
   * 
   * POST /api/v1/mobile-money/transactions
   */
  public async initiateTransaction(req: Request, res: Response): Promise<void> {
    const endpointName = 'mobile_money.initiate_transaction';
    const tenantId = this.getTenantId(req);
    const span = this.tracer.startSpan(endpointName);
    
    try {
      this.recordRequestStart(endpointName, tenantId);
      
      // Validar autenticação
      const authResult = await this.validateAuthToken(req, res);
      if (!authResult) {
        return;
      }
      
      // Validar payload
      const validationResult = this.validationService.validateSchema(
        'mobileMoneyTransaction', 
        req.body
      );
      
      if (!validationResult.valid) {
        this.sendErrorResponse(
          res, 
          400, 
          'Dados inválidos para transação', 
          'VALIDATION_ERROR',
          validationResult.errors
        );
        return;
      }
      
      const { 
        provider, 
        amount, 
        currency, 
        phoneNumber, 
        type,
        description,
        metadata = {}
      } = req.body;
      
      // Verificar se o provedor é suportado
      if (!Object.values(MobileMoneyProvider).includes(provider)) {
        this.sendErrorResponse(res, 400, `Provedor "${provider}" não suportado`, 'INVALID_PROVIDER');
        return;
      }
      
      // Verificar se o tipo de transação é suportado
      if (!Object.values(TransactionType).includes(type)) {
        this.sendErrorResponse(res, 400, `Tipo de transação "${type}" não suportado`, 'INVALID_TRANSACTION_TYPE');
        return;
      }
      
      // Extrair informações do dispositivo
      const deviceInfo = {
        deviceId: req.header('X-Device-ID') || null,
        ipAddress: req.ip,
        userAgent: req.header('User-Agent') || null,
      };
      
      // Iniciar transação
      const transactionResult = await this.mobileMoneyService.initiateTransaction({
        userId: authResult.userId,
        tenantId,
        provider,
        amount,
        currency,
        phoneNumber,
        type,
        description,
        metadata: {
          ...metadata,
          apiSource: 'rest_api',
          ipAddress: req.ip
        },
        deviceInfo
      });
      
      // Retornar resultado
      this.sendSuccessResponse(res, {
        transactionId: transactionResult.transactionId,
        status: transactionResult.status,
        provider: transactionResult.provider,
        requiresOtp: transactionResult.requiresOtp,
        otpSent: transactionResult.otpSent,
        amount: transactionResult.amount,
        currency: transactionResult.currency,
        expiresAt: transactionResult.expiresAt
      }, 201);
      
      this.recordRequestResult(endpointName, tenantId, 201);
    } catch (error) {
      this.handleError(error, req, res, endpointName);
    } finally {
      span.end();
    }
  }
  
  /**
   * Verifica o código OTP para uma transação
   * 
   * POST /api/v1/mobile-money/transactions/{transactionId}/verify-otp
   */
  public async verifyOtp(req: Request, res: Response): Promise<void> {
    const endpointName = 'mobile_money.verify_otp';
    const tenantId = this.getTenantId(req);
    const span = this.tracer.startSpan(endpointName);
    
    try {
      this.recordRequestStart(endpointName, tenantId);
      
      // Validar autenticação
      const authResult = await this.validateAuthToken(req, res);
      if (!authResult) {
        return;
      }
      
      const transactionId = req.params.transactionId;
      const { otpCode } = req.body;
      
      // Validar código OTP
      if (!otpCode || typeof otpCode !== 'string' || otpCode.trim().length === 0) {
        this.sendErrorResponse(res, 400, 'Código OTP é obrigatório', 'MISSING_OTP');
        return;
      }
      
      // Extrair informações do dispositivo
      const deviceInfo = {
        deviceId: req.header('X-Device-ID') || null,
        ipAddress: req.ip,
        userAgent: req.header('User-Agent') || null,
      };
      
      // Verificar OTP
      const verificationResult = await this.mobileMoneyService.verifyOTP({
        transactionId,
        tenantId,
        otpCode,
        deviceInfo
      });
      
      // Retornar resultado
      this.sendSuccessResponse(res, {
        transactionId,
        verified: verificationResult.verified,
        status: verificationResult.verified ? 'PROCESSING' : 'PENDING',
        remainingAttempts: verificationResult.remainingAttempts
      });
      
      this.recordRequestResult(endpointName, tenantId, 200);
    } catch (error) {
      this.handleError(error, req, res, endpointName);
    } finally {
      span.end();
    }
  }
  
  /**
   * Verifica o status de uma transação
   * 
   * GET /api/v1/mobile-money/transactions/{transactionId}
   */
  public async checkTransactionStatus(req: Request, res: Response): Promise<void> {
    const endpointName = 'mobile_money.check_status';
    const tenantId = this.getTenantId(req);
    const span = this.tracer.startSpan(endpointName);
    
    try {
      this.recordRequestStart(endpointName, tenantId);
      
      // Validar autenticação
      const authResult = await this.validateAuthToken(req, res);
      if (!authResult) {
        return;
      }
      
      const transactionId = req.params.transactionId;
      
      // Verificar status da transação
      const transactionStatus = await this.mobileMoneyService.checkTransactionStatus({
        transactionId,
        tenantId
      });
      
      // Retornar resultado
      this.sendSuccessResponse(res, {
        transactionId,
        status: transactionStatus.status,
        provider: transactionStatus.provider,
        amount: transactionStatus.amount,
        currency: transactionStatus.currency,
        completedAt: transactionStatus.completedAt,
        failureReason: transactionStatus.failureReason
      });
      
      this.recordRequestResult(endpointName, tenantId, 200);
    } catch (error) {
      this.handleError(error, req, res, endpointName);
    } finally {
      span.end();
    }
  }
  
  /**
   * Cancela uma transação pendente
   * 
   * POST /api/v1/mobile-money/transactions/{transactionId}/cancel
   */
  public async cancelTransaction(req: Request, res: Response): Promise<void> {
    const endpointName = 'mobile_money.cancel_transaction';
    const tenantId = this.getTenantId(req);
    const span = this.tracer.startSpan(endpointName);
    
    try {
      this.recordRequestStart(endpointName, tenantId);
      
      // Validar autenticação
      const authResult = await this.validateAuthToken(req, res);
      if (!authResult) {
        return;
      }
      
      const transactionId = req.params.transactionId;
      const { reason } = req.body;
      
      // Cancelar transação
      const cancelResult = await this.mobileMoneyService.cancelTransaction({
        transactionId,
        tenantId,
        reason
      });
      
      // Retornar resultado
      this.sendSuccessResponse(res, {
        transactionId,
        cancelled: cancelResult.cancelled,
        status: cancelResult.status
      });
      
      this.recordRequestResult(endpointName, tenantId, 200);
    } catch (error) {
      this.handleError(error, req, res, endpointName);
    } finally {
      span.end();
    }
  }
  
  /**
   * Lista transações do usuário
   * 
   * GET /api/v1/mobile-money/transactions
   */
  public async listTransactions(req: Request, res: Response): Promise<void> {
    const endpointName = 'mobile_money.list_transactions';
    const tenantId = this.getTenantId(req);
    const span = this.tracer.startSpan(endpointName);
    
    try {
      this.recordRequestStart(endpointName, tenantId);
      
      // Validar autenticação
      const authResult = await this.validateAuthToken(req, res);
      if (!authResult) {
        return;
      }
      
      // Extrair parâmetros de paginação e filtros
      const limit = parseInt(req.query.limit as string, 10) || 10;
      const offset = parseInt(req.query.offset as string, 10) || 0;
      const provider = req.query.provider as string;
      const status = req.query.status as string;
      const type = req.query.type as string;
      
      // Listar transações
      const transactions = await this.mobileMoneyService.getTransactionHistory({
        userId: authResult.userId,
        tenantId,
        limit,
        offset,
        provider: provider ? provider : undefined,
        status: status ? status : undefined,
        type: type ? type : undefined
      });
      
      // Retornar resultado
      this.sendSuccessResponse(res, {
        transactions: transactions.transactions,
        totalCount: transactions.totalCount,
        limit,
        offset,
        hasMore: transactions.hasMore
      });
      
      this.recordRequestResult(endpointName, tenantId, 200);
    } catch (error) {
      this.handleError(error, req, res, endpointName);
    } finally {
      span.end();
    }
  }
  
  /**
   * Obtém os provedores Mobile Money disponíveis
   * 
   * GET /api/v1/mobile-money/providers
   */
  public async getAvailableProviders(req: Request, res: Response): Promise<void> {
    const endpointName = 'mobile_money.get_providers';
    const tenantId = this.getTenantId(req);
    const span = this.tracer.startSpan(endpointName);
    
    try {
      this.recordRequestStart(endpointName, tenantId);
      
      // Validar autenticação
      const authResult = await this.validateAuthToken(req, res);
      if (!authResult) {
        return;
      }
      
      // Obter provedores disponíveis
      const providers = await this.mobileMoneyService.getAvailableProviders(tenantId);
      
      // Retornar resultado
      this.sendSuccessResponse(res, {
        providers: providers.map((p: any) => ({
          id: p.id,
          name: p.name,
          code: p.code,
          countries: p.countries,
          currencies: p.currencies,
          features: p.features,
          status: p.status
        }))
      });
      
      this.recordRequestResult(endpointName, tenantId, 200);
    } catch (error) {
      this.handleError(error, req, res, endpointName);
    } finally {
      span.end();
    }
  }
  
  /**
   * Verifica se uma transação é elegível antes de iniciá-la
   * 
   * POST /api/v1/mobile-money/check-eligibility
   */
  public async checkEligibility(req: Request, res: Response): Promise<void> {
    const endpointName = 'mobile_money.check_eligibility';
    const tenantId = this.getTenantId(req);
    const span = this.tracer.startSpan(endpointName);
    
    try {
      this.recordRequestStart(endpointName, tenantId);
      
      // Validar autenticação
      const authResult = await this.validateAuthToken(req, res);
      if (!authResult) {
        return;
      }
      
      // Validar payload
      const validationResult = this.validationService.validateSchema(
        'mobileMoneyEligibility', 
        req.body
      );
      
      if (!validationResult.valid) {
        this.sendErrorResponse(
          res, 
          400, 
          'Dados inválidos para verificação de elegibilidade', 
          'VALIDATION_ERROR',
          validationResult.errors
        );
        return;
      }
      
      const { provider, amount, currency, type } = req.body;
      
      // Verificar elegibilidade
      const eligibilityResult = await this.mobileMoneyService.checkTransactionEligibility({
        userId: authResult.userId,
        tenantId,
        provider,
        amount,
        currency,
        type
      });
      
      // Retornar resultado
      this.sendSuccessResponse(res, {
        eligible: eligibilityResult.eligible,
        limitReached: eligibilityResult.limitReached,
        kycRequired: eligibilityResult.kycRequired,
        reasons: eligibilityResult.reasons,
        limits: eligibilityResult.limits
      });
      
      this.recordRequestResult(endpointName, tenantId, 200);
    } catch (error) {
      this.handleError(error, req, res, endpointName);
    } finally {
      span.end();
    }
  }
}