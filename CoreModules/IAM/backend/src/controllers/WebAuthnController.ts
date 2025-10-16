/**
 * ============================================================================
 * INNOVABIZ IAM - WebAuthn Controller
 * ============================================================================
 * Version: 1.0.0
 * Date: 31/07/2025
 * Author: Equipe de Segurança INNOVABIZ
 * Classification: Confidencial - Interno
 * 
 * Description: Controlador REST para operações WebAuthn/FIDO2
 * Standards: W3C WebAuthn Level 3, FIDO2 CTAP2.1, NIST SP 800-63B
 * Compliance: PCI DSS 4.0, GDPR/LGPD, PSD2, ISO 27001
 * ============================================================================
 */

import { Request, Response, NextFunction } from 'express';
import { Logger } from 'winston';
import Joi from 'joi';
import { v4 as uuidv4 } from 'uuid';

import { WebAuthnService } from '../services/WebAuthnService';
import { CredentialService } from '../services/CredentialService';
import { webauthnMetrics } from '../metrics/webauthn';
import { errorTemplates } from '../config/webauthn';
import {
  WebAuthnContext,
  WebAuthnAPIResponse,
  WebAuthnError,
  SignCountAnomalyError,
  RegistrationOptionsRequest,
  AuthenticationOptionsRequest
} from '../types/webauthn';

/**
 * Controlador para endpoints WebAuthn
 */
export class WebAuthnController {
  private readonly logger: Logger;
  private readonly webauthnService: WebAuthnService;
  private readonly credentialService: CredentialService;

  constructor(
    logger: Logger,
    webauthnService: WebAuthnService,
    credentialService: CredentialService
  ) {
    this.logger = logger;
    this.webauthnService = webauthnService;
    this.credentialService = credentialService;
  }

  /**
   * POST /api/v1/webauthn/registration/options
   * Gera opções para registro de credencial WebAuthn
   */
  public generateRegistrationOptions = async (
    req: Request,
    res: Response,
    next: NextFunction
  ): Promise<void> => {
    const timer = webauthnMetrics.httpRequestDuration.startTimer({
      method: 'POST',
      endpoint: '/webauthn/registration/options',
      status_code: '200'
    });

    try {
      // Validar entrada
      const { error, value } = this.validateRegistrationOptionsRequest(req.body);
      if (error) {
        timer({ status_code: '400' });
        this.sendErrorResponse(res, 'VALIDATION_ERROR', error.details[0].message, 400);
        return;
      }

      // Extrair contexto
      const context = this.extractContext(req);
      
      // Gerar opções
      const options = await this.webauthnService.generateRegistrationOptions(context, value);

      // Métricas
      webauthnMetrics.httpRequestsTotal.inc({
        method: 'POST',
        endpoint: '/webauthn/registration/options',
        status_code: '200',
        tenant_id: context.tenantId || 'unknown'
      });

      this.sendSuccessResponse(res, options, 'Registration options generated successfully');

    } catch (error) {
      timer({ status_code: '500' });
      this.handleError(error, res, next);
    } finally {
      timer();
    }
  };

  /**
   * POST /api/v1/webauthn/registration/verify
   * Verifica e registra credencial WebAuthn
   */
  public verifyRegistration = async (
    req: Request,
    res: Response,
    next: NextFunction
  ): Promise<void> => {
    const timer = webauthnMetrics.httpRequestDuration.startTimer({
      method: 'POST',
      endpoint: '/webauthn/registration/verify',
      status_code: '200'
    });

    try {
      // Validar entrada
      const { error, value } = this.validateRegistrationVerificationRequest(req.body);
      if (error) {
        timer({ status_code: '400' });
        this.sendErrorResponse(res, 'VALIDATION_ERROR', error.details[0].message, 400);
        return;
      }

      // Extrair contexto
      const context = this.extractContext(req);

      // Verificar registro
      const result = await this.webauthnService.verifyRegistration(context, value);

      // Métricas
      webauthnMetrics.httpRequestsTotal.inc({
        method: 'POST',
        endpoint: '/webauthn/registration/verify',
        status_code: '200',
        tenant_id: context.tenantId || 'unknown'
      });

      this.sendSuccessResponse(res, result, 'Registration verified successfully', 201);

    } catch (error) {
      timer({ status_code: error instanceof WebAuthnError ? '400' : '500' });
      this.handleError(error, res, next);
    } finally {
      timer();
    }
  };

  /**
   * POST /api/v1/webauthn/authentication/options
   * Gera opções para autenticação WebAuthn
   */
  public generateAuthenticationOptions = async (
    req: Request,
    res: Response,
    next: NextFunction
  ): Promise<void> => {
    const timer = webauthnMetrics.httpRequestDuration.startTimer({
      method: 'POST',
      endpoint: '/webauthn/authentication/options',
      status_code: '200'
    });

    try {
      // Validar entrada
      const { error, value } = this.validateAuthenticationOptionsRequest(req.body);
      if (error) {
        timer({ status_code: '400' });
        this.sendErrorResponse(res, 'VALIDATION_ERROR', error.details[0].message, 400);
        return;
      }

      // Extrair contexto
      const context = this.extractContext(req);

      // Gerar opções
      const options = await this.webauthnService.generateAuthenticationOptions(context, value);

      // Métricas
      webauthnMetrics.httpRequestsTotal.inc({
        method: 'POST',
        endpoint: '/webauthn/authentication/options',
        status_code: '200',
        tenant_id: context.tenantId || 'unknown'
      });

      this.sendSuccessResponse(res, options, 'Authentication options generated successfully');

    } catch (error) {
      timer({ status_code: '500' });
      this.handleError(error, res, next);
    } finally {
      timer();
    }
  };

  /**
   * POST /api/v1/webauthn/authentication/verify
   * Verifica autenticação WebAuthn
   */
  public verifyAuthentication = async (
    req: Request,
    res: Response,
    next: NextFunction
  ): Promise<void> => {
    const timer = webauthnMetrics.httpRequestDuration.startTimer({
      method: 'POST',
      endpoint: '/webauthn/authentication/verify',
      status_code: '200'
    });

    try {
      // Validar entrada
      const { error, value } = this.validateAuthenticationVerificationRequest(req.body);
      if (error) {
        timer({ status_code: '400' });
        this.sendErrorResponse(res, 'VALIDATION_ERROR', error.details[0].message, 400);
        return;
      }

      // Extrair contexto
      const context = this.extractContext(req);

      // Verificar autenticação
      const result = await this.webauthnService.verifyAuthentication(context, value);

      // Métricas
      webauthnMetrics.httpRequestsTotal.inc({
        method: 'POST',
        endpoint: '/webauthn/authentication/verify',
        status_code: '200',
        tenant_id: context.tenantId || 'unknown'
      });

      this.sendSuccessResponse(res, result, 'Authentication verified successfully');

    } catch (error) {
      timer({ status_code: error instanceof WebAuthnError ? '401' : '500' });
      this.handleError(error, res, next);
    } finally {
      timer();
    }
  };

  /**
   * GET /api/v1/webauthn/credentials
   * Lista credenciais do usuário
   */
  public getUserCredentials = async (
    req: Request,
    res: Response,
    next: NextFunction
  ): Promise<void> => {
    const timer = webauthnMetrics.httpRequestDuration.startTimer({
      method: 'GET',
      endpoint: '/webauthn/credentials',
      status_code: '200'
    });

    try {
      // Extrair contexto
      const context = this.extractContext(req);

      if (!context.userId || !context.tenantId) {
        timer({ status_code: '400' });
        this.sendErrorResponse(res, 'MISSING_USER_CONTEXT', 'User ID and Tenant ID are required', 400);
        return;
      }

      // Extrair filtros da query string
      const filter = this.extractCredentialFilter(req.query);

      // Buscar credenciais
      const credentials = await this.credentialService.getUserCredentials(
        context.userId,
        context.tenantId,
        filter
      );

      // Filtrar dados sensíveis para resposta
      const sanitizedCredentials = credentials.map(cred => ({
        id: cred.id,
        credentialId: cred.credentialId,
        friendlyName: cred.friendlyName,
        deviceType: cred.deviceType,
        authenticatorType: cred.authenticatorType,
        complianceLevel: cred.complianceLevel,
        status: cred.status,
        createdAt: cred.createdAt,
        lastUsedAt: cred.lastUsedAt,
        transports: cred.transports,
        userVerified: cred.userVerified,
        backupEligible: cred.backupEligible,
        backupState: cred.backupState
      }));

      // Métricas
      webauthnMetrics.httpRequestsTotal.inc({
        method: 'GET',
        endpoint: '/webauthn/credentials',
        status_code: '200',
        tenant_id: context.tenantId
      });

      this.sendSuccessResponse(res, {
        credentials: sanitizedCredentials,
        total: sanitizedCredentials.length
      }, 'Credentials retrieved successfully');

    } catch (error) {
      timer({ status_code: '500' });
      this.handleError(error, res, next);
    } finally {
      timer();
    }
  };

  /**
   * PUT /api/v1/webauthn/credentials/:credentialId/name
   * Atualiza nome amigável da credencial
   */
  public updateCredentialName = async (
    req: Request,
    res: Response,
    next: NextFunction
  ): Promise<void> => {
    const timer = webauthnMetrics.httpRequestDuration.startTimer({
      method: 'PUT',
      endpoint: '/webauthn/credentials/:id/name',
      status_code: '200'
    });

    try {
      // Validar entrada
      const { error, value } = this.validateUpdateCredentialNameRequest(req.body);
      if (error) {
        timer({ status_code: '400' });
        this.sendErrorResponse(res, 'VALIDATION_ERROR', error.details[0].message, 400);
        return;
      }

      // Extrair contexto
      const context = this.extractContext(req);
      const credentialId = req.params.credentialId;

      if (!context.userId || !context.tenantId) {
        timer({ status_code: '400' });
        this.sendErrorResponse(res, 'MISSING_USER_CONTEXT', 'User ID and Tenant ID are required', 400);
        return;
      }

      // Atualizar nome
      await this.credentialService.updateCredentialName(
        credentialId,
        context.userId,
        context.tenantId,
        value.friendlyName
      );

      // Métricas
      webauthnMetrics.httpRequestsTotal.inc({
        method: 'PUT',
        endpoint: '/webauthn/credentials/:id/name',
        status_code: '200',
        tenant_id: context.tenantId
      });

      this.sendSuccessResponse(res, null, 'Credential name updated successfully');

    } catch (error) {
      timer({ status_code: '500' });
      this.handleError(error, res, next);
    } finally {
      timer();
    }
  };

  /**
   * DELETE /api/v1/webauthn/credentials/:credentialId
   * Remove credencial do usuário
   */
  public deleteCredential = async (
    req: Request,
    res: Response,
    next: NextFunction
  ): Promise<void> => {
    const timer = webauthnMetrics.httpRequestDuration.startTimer({
      method: 'DELETE',
      endpoint: '/webauthn/credentials/:id',
      status_code: '200'
    });

    try {
      // Extrair contexto
      const context = this.extractContext(req);
      const credentialId = req.params.credentialId;
      const reason = req.body.reason || 'User requested deletion';

      if (!context.userId || !context.tenantId) {
        timer({ status_code: '400' });
        this.sendErrorResponse(res, 'MISSING_USER_CONTEXT', 'User ID and Tenant ID are required', 400);
        return;
      }

      // Verificar se credencial pertence ao usuário
      const credential = await this.credentialService.getCredentialById(credentialId);
      if (!credential || credential.userId !== context.userId || credential.tenantId !== context.tenantId) {
        timer({ status_code: '404' });
        this.sendErrorResponse(res, 'CREDENTIAL_NOT_FOUND', 'Credential not found or access denied', 404);
        return;
      }

      // Deletar credencial
      await this.credentialService.deleteCredential(credentialId, reason);

      // Métricas
      webauthnMetrics.httpRequestsTotal.inc({
        method: 'DELETE',
        endpoint: '/webauthn/credentials/:id',
        status_code: '200',
        tenant_id: context.tenantId
      });

      this.sendSuccessResponse(res, null, 'Credential deleted successfully');

    } catch (error) {
      timer({ status_code: '500' });
      this.handleError(error, res, next);
    } finally {
      timer();
    }
  };

  /**
   * GET /api/v1/webauthn/stats
   * Estatísticas de credenciais do tenant
   */
  public getCredentialStats = async (
    req: Request,
    res: Response,
    next: NextFunction
  ): Promise<void> => {
    const timer = webauthnMetrics.httpRequestDuration.startTimer({
      method: 'GET',
      endpoint: '/webauthn/stats',
      status_code: '200'
    });

    try {
      // Extrair contexto
      const context = this.extractContext(req);

      if (!context.tenantId) {
        timer({ status_code: '400' });
        this.sendErrorResponse(res, 'MISSING_TENANT_CONTEXT', 'Tenant ID is required', 400);
        return;
      }

      // Buscar estatísticas
      const stats = await this.credentialService.getCredentialStats(context.tenantId);

      // Métricas
      webauthnMetrics.httpRequestsTotal.inc({
        method: 'GET',
        endpoint: '/webauthn/stats',
        status_code: '200',
        tenant_id: context.tenantId
      });

      this.sendSuccessResponse(res, stats, 'Statistics retrieved successfully');

    } catch (error) {
      timer({ status_code: '500' });
      this.handleError(error, res, next);
    } finally {
      timer();
    }
  };

  /**
   * Métodos privados de utilidade
   */

  private extractContext(req: Request): WebAuthnContext {
    const correlationId = req.headers['x-correlation-id'] as string || uuidv4();
    
    return {
      userId: req.user?.id,
      tenantId: req.user?.tenantId || req.headers['x-tenant-id'] as string,
      userEmail: req.user?.email,
      userDisplayName: req.user?.displayName,
      regionCode: req.headers['x-region'] as string || 'DEFAULT',
      origin: req.headers.origin || req.headers.referer || 'unknown',
      ipAddress: req.ip || req.connection.remoteAddress || 'unknown',
      userAgent: req.headers['user-agent'] || 'unknown',
      correlationId,
      sessionId: req.sessionID,
      deviceFingerprint: req.headers['x-device-fingerprint'] as string
    };
  }

  private extractCredentialFilter(query: any): any {
    return {
      status: query.status,
      authenticatorType: query.authenticatorType,
      complianceLevel: query.complianceLevel,
      deviceType: query.deviceType,
      limit: query.limit ? parseInt(query.limit) : undefined,
      offset: query.offset ? parseInt(query.offset) : undefined
    };
  }

  private sendSuccessResponse<T>(
    res: Response,
    data: T,
    message: string,
    statusCode: number = 200
  ): void {
    const response: WebAuthnAPIResponse<T> = {
      success: true,
      data,
      metadata: {
        correlationId: res.locals.correlationId || uuidv4(),
        timestamp: new Date().toISOString(),
        version: '1.0.0'
      }
    };

    if (message) {
      response.metadata.message = message;
    }

    res.status(statusCode).json(response);
  }

  private sendErrorResponse(
    res: Response,
    code: string,
    message: string,
    statusCode: number,
    details?: any
  ): void {
    const response: WebAuthnAPIResponse = {
      success: false,
      error: {
        code,
        message,
        details
      },
      metadata: {
        correlationId: res.locals.correlationId || uuidv4(),
        timestamp: new Date().toISOString(),
        version: '1.0.0'
      }
    };

    res.status(statusCode).json(response);
  }

  private handleError(error: any, res: Response, next: NextFunction): void {
    if (error instanceof WebAuthnError) {
      const template = errorTemplates[error.code];
      if (template) {
        this.sendErrorResponse(res, error.code, error.message, template.httpStatus, error.details);
      } else {
        this.sendErrorResponse(res, error.code, error.message, 400, error.details);
      }
    } else if (error instanceof SignCountAnomalyError) {
      this.sendErrorResponse(res, error.code, error.message, 403, {
        expectedCount: error.expectedCount,
        receivedCount: error.receivedCount
      });
    } else {
      this.logger.error('Unhandled error in WebAuthn controller', {
        error: error.message,
        stack: error.stack
      });
      this.sendErrorResponse(res, 'INTERNAL_ERROR', 'Internal server error', 500);
    }
  }

  /**
   * Esquemas de validação Joi
   */

  private validateRegistrationOptionsRequest(data: any) {
    const schema = Joi.object({
      username: Joi.string().min(3).max(64).pattern(/^[a-zA-Z0-9._-]+$/).optional(),
      displayName: Joi.string().min(1).max(128).optional(),
      attestation: Joi.string().valid('none', 'indirect', 'direct', 'enterprise').optional(),
      authenticatorSelection: Joi.object({
        authenticatorAttachment: Joi.string().valid('platform', 'cross-platform').optional(),
        userVerification: Joi.string().valid('required', 'preferred', 'discouraged').optional(),
        residentKey: Joi.string().valid('required', 'preferred', 'discouraged').optional()
      }).optional(),
      excludeCredentials: Joi.boolean().optional(),
      timeout: Joi.number().min(30000).max(300000).optional()
    });

    return schema.validate(data);
  }

  private validateRegistrationVerificationRequest(data: any) {
    const schema = Joi.object({
      id: Joi.string().required(),
      rawId: Joi.string().required(),
      response: Joi.object({
        clientDataJSON: Joi.string().required(),
        attestationObject: Joi.string().required(),
        transports: Joi.array().items(Joi.string()).optional()
      }).required(),
      type: Joi.string().valid('public-key').required(),
      clientExtensionResults: Joi.object().optional(),
      authenticatorAttachment: Joi.string().valid('platform', 'cross-platform').optional()
    });

    return schema.validate(data);
  }

  private validateAuthenticationOptionsRequest(data: any) {
    const schema = Joi.object({
      userVerification: Joi.string().valid('required', 'preferred', 'discouraged').optional(),
      allowCredentials: Joi.array().items(Joi.string()).optional(),
      timeout: Joi.number().min(30000).max(300000).optional()
    });

    return schema.validate(data);
  }

  private validateAuthenticationVerificationRequest(data: any) {
    const schema = Joi.object({
      id: Joi.string().required(),
      rawId: Joi.string().required(),
      response: Joi.object({
        clientDataJSON: Joi.string().required(),
        authenticatorData: Joi.string().required(),
        signature: Joi.string().required(),
        userHandle: Joi.string().allow(null).optional()
      }).required(),
      type: Joi.string().valid('public-key').required(),
      clientExtensionResults: Joi.object().optional(),
      authenticatorAttachment: Joi.string().valid('platform', 'cross-platform').optional()
    });

    return schema.validate(data);
  }

  private validateUpdateCredentialNameRequest(data: any) {
    const schema = Joi.object({
      friendlyName: Joi.string().min(1).max(64).required()
    });

    return schema.validate(data);
  }
}