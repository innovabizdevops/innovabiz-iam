/**
 * 🔑 JWT STRATEGY - INNOVABIZ IAM
 * Estratégia de autenticação JWT com validação avançada
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: NIST SP 800-63B, OWASP JWT Security, RFC 7519
 * Security: Token Validation, Session Management, Risk Assessment
 */

import { Injectable, UnauthorizedException, Logger, Inject } from '@nestjs/common';
import { PassportStrategy } from '@nestjs/passport';
import { ConfigService } from '@nestjs/config';
import { Strategy, ExtractJwt, StrategyOptions } from 'passport-jwt';
import { Request } from 'express';
import { Cache } from 'cache-manager';
import { CACHE_MANAGER } from '@nestjs/cache-manager';

import { IAMService } from '../services/IAMService';
import { RiskAssessmentService } from '../services/RiskAssessmentService';
import { AuditService } from '../services/AuditService';

// Interfaces
interface JwtPayload {
  sub: string; // User ID
  email?: string;
  tenantId: string;
  sessionId?: string;
  type: 'access' | 'refresh';
  iat: number;
  exp: number;
  iss: string;
  aud: string;
  jti?: string; // JWT ID
  scope?: string[];
  deviceId?: string;
  ipAddress?: string;
  riskScore?: number;
}

interface ValidatedUser {
  id: string;
  email: string;
  tenantId: string;
  sessionId?: string;
  scope: string[];
  deviceId?: string;
  riskScore: number;
  lastActivity: Date;
  isActive: boolean;
  requiresMFA?: boolean;
  metadata?: Record<string, any>;
}

@Injectable()
export class JwtStrategy extends PassportStrategy(Strategy, 'jwt') {
  private readonly logger = new Logger(JwtStrategy.name);

  constructor(
    private readonly configService: ConfigService,
    private readonly iamService: IAMService,
    private readonly riskAssessmentService: RiskAssessmentService,
    private readonly auditService: AuditService,
    @Inject(CACHE_MANAGER) private readonly cacheManager: Cache
  ) {
    super({
      jwtFromRequest: ExtractJwt.fromExtractors([
        JwtStrategy.extractJWTFromCookie,
        JwtStrategy.extractJWTFromHeader,
        JwtStrategy.extractJWTFromQuery
      ]),
      ignoreExpiration: false,
      secretOrKey: configService.get<string>('jwt.secret'),
      issuer: configService.get<string>('jwt.issuer', 'innovabiz-iam'),
      audience: configService.get<string>('jwt.audience', 'innovabiz-platform'),
      algorithms: [configService.get<string>('jwt.algorithm', 'HS256')],
      passReqToCallback: true, // Permitir acesso ao request
      clockTolerance: configService.get<number>('jwt.clockTolerance', 30)
    } as StrategyOptions);
  }

  /**
   * Validar payload JWT e retornar usuário autenticado
   */
  async validate(request: Request, payload: JwtPayload): Promise<ValidatedUser> {
    try {
      this.logger.debug(`Validating JWT for user: ${payload.sub}`);

      // Validações básicas do payload
      this.validatePayloadStructure(payload);

      // Verificar se token não está na blacklist
      await this.checkTokenBlacklist(payload);

      // Validar sessão se presente
      await this.validateSession(payload, request);

      // Obter dados do usuário
      const user = await this.getUserData(payload);

      // Avaliar risco da requisição
      const riskScore = await this.assessRequestRisk(payload, request, user);

      // Validar contexto de segurança
      await this.validateSecurityContext(payload, request, user, riskScore);

      // Atualizar atividade do usuário
      await this.updateUserActivity(user, request);

      // Construir objeto do usuário validado
      const validatedUser: ValidatedUser = {
        id: user.id,
        email: user.email,
        tenantId: payload.tenantId,
        sessionId: payload.sessionId,
        scope: payload.scope || ['user'],
        deviceId: payload.deviceId,
        riskScore,
        lastActivity: new Date(),
        isActive: user.isActive,
        requiresMFA: this.shouldRequireMFA(riskScore, user),
        metadata: {
          ipAddress: this.getClientIP(request),
          userAgent: request.get('User-Agent'),
          tokenIssued: new Date(payload.iat * 1000),
          tokenExpires: new Date(payload.exp * 1000)
        }
      };

      // Log de autenticação bem-sucedida
      await this.logSuccessfulAuthentication(validatedUser, request);

      return validatedUser;

    } catch (error) {
      this.logger.warn(`JWT validation failed: ${error.message}`, {
        userId: payload?.sub,
        error: error.message,
        ip: this.getClientIP(request)
      });

      // Log de tentativa de autenticação falhada
      await this.logFailedAuthentication(payload, request, error);

      throw new UnauthorizedException('Invalid token');
    }
  }

  /**
   * Extrair JWT do cookie
   */
  private static extractJWTFromCookie(request: Request): string | null {
    if (request.cookies && request.cookies.access_token) {
      return request.cookies.access_token;
    }
    return null;
  }

  /**
   * Extrair JWT do header Authorization
   */
  private static extractJWTFromHeader(request: Request): string | null {
    const authHeader = request.get('Authorization');
    if (authHeader && authHeader.startsWith('Bearer ')) {
      return authHeader.substring(7);
    }
    return null;
  }

  /**
   * Extrair JWT do query parameter (apenas em desenvolvimento)
   */
  private static extractJWTFromQuery(request: Request): string | null {
    if (process.env.NODE_ENV !== 'production' && request.query?.token) {
      return request.query.token as string;
    }
    return null;
  }

  /**
   * Validar estrutura do payload
   */
  private validatePayloadStructure(payload: JwtPayload): void {
    if (!payload.sub) {
      throw new UnauthorizedException('Token missing subject (user ID)');
    }

    if (!payload.tenantId) {
      throw new UnauthorizedException('Token missing tenant ID');
    }

    if (payload.type === 'refresh') {
      throw new UnauthorizedException('Refresh token cannot be used for authentication');
    }

    // Validar timestamps
    const now = Math.floor(Date.now() / 1000);
    const clockTolerance = this.configService.get<number>('jwt.clockTolerance', 30);

    if (payload.iat > now + clockTolerance) {
      throw new UnauthorizedException('Token used before issued');
    }

    if (payload.exp < now - clockTolerance) {
      throw new UnauthorizedException('Token has expired');
    }

    // Validar idade máxima do token
    const maxAge = this.configService.get<number>('jwt.maxAge', 86400); // 24 horas
    if (now - payload.iat > maxAge) {
      throw new UnauthorizedException('Token is too old');
    }
  }

  /**
   * Verificar se token está na blacklist
   */
  private async checkTokenBlacklist(payload: JwtPayload): Promise<void> {
    try {
      // Verificar por JTI se disponível
      if (payload.jti) {
        const isBlacklisted = await this.cacheManager.get(`blacklist_jti:${payload.jti}`);
        if (isBlacklisted) {
          throw new UnauthorizedException('Token has been revoked');
        }
      }

      // Verificar por usuário e timestamp
      const userBlacklistKey = `blacklist_user:${payload.sub}:${payload.iat}`;
      const isUserTokenBlacklisted = await this.cacheManager.get(userBlacklistKey);
      if (isUserTokenBlacklisted) {
        throw new UnauthorizedException('Token has been revoked');
      }

    } catch (error) {
      if (error instanceof UnauthorizedException) {
        throw error;
      }
      this.logger.warn(`Blacklist check failed: ${error.message}`);
      // Continuar se não conseguir verificar blacklist
    }
  }

  /**
   * Validar sessão ativa
   */
  private async validateSession(payload: JwtPayload, request: Request): Promise<void> {
    if (!payload.sessionId) {
      return; // Sessão não obrigatória para todos os tokens
    }

    try {
      const isSessionValid = await this.iamService.validateSession(payload.sessionId);
      if (!isSessionValid) {
        throw new UnauthorizedException('Session is not active');
      }

      // Verificar se IP da sessão coincide (se configurado)
      const enforceIPBinding = this.configService.get<boolean>('security.enforceIPBinding', false);
      if (enforceIPBinding && payload.ipAddress) {
        const currentIP = this.getClientIP(request);
        if (payload.ipAddress !== currentIP) {
          throw new UnauthorizedException('IP address mismatch');
        }
      }

    } catch (error) {
      if (error instanceof UnauthorizedException) {
        throw error;
      }
      this.logger.warn(`Session validation failed: ${error.message}`);
      // Não falhar autenticação por problemas de sessão não críticos
    }
  }

  /**
   * Obter dados do usuário
   */
  private async getUserData(payload: JwtPayload): Promise<any> {
    try {
      // Tentar cache primeiro
      const cacheKey = `user:${payload.sub}:${payload.tenantId}`;
      let user = await this.cacheManager.get(cacheKey);

      if (!user) {
        // Buscar no banco de dados
        user = await this.iamService.getUserById(payload.sub, payload.tenantId);
        
        if (!user) {
          throw new UnauthorizedException('User not found');
        }

        // Cache por 5 minutos
        await this.cacheManager.set(cacheKey, user, 300);
      }

      // Verificar se usuário está ativo
      if (!user.isActive) {
        throw new UnauthorizedException('User account is disabled');
      }

      return user;

    } catch (error) {
      if (error instanceof UnauthorizedException) {
        throw error;
      }
      throw new UnauthorizedException('Failed to validate user');
    }
  }

  /**
   * Avaliar risco da requisição
   */
  private async assessRequestRisk(
    payload: JwtPayload, 
    request: Request, 
    user: any
  ): Promise<number> {
    try {
      const context = {
        userId: payload.sub,
        tenantId: payload.tenantId,
        ipAddress: this.getClientIP(request),
        userAgent: request.get('User-Agent') || '',
        deviceFingerprint: request.get('X-Device-Fingerprint'),
        timestamp: new Date()
      };

      // Usar serviço de avaliação de risco
      const riskAssessment = await this.riskAssessmentService.assessAuthenticationRisk(
        payload.sub,
        payload.tenantId,
        context
      );

      return riskAssessment.riskScore;

    } catch (error) {
      this.logger.warn(`Risk assessment failed: ${error.message}`);
      return 0.5; // Risco médio como fallback
    }
  }

  /**
   * Validar contexto de segurança
   */
  private async validateSecurityContext(
    payload: JwtPayload,
    request: Request,
    user: any,
    riskScore: number
  ): Promise<void> {
    // Verificar se risco é muito alto
    const maxAllowedRisk = this.configService.get<number>('security.maxRiskScore', 0.8);
    if (riskScore > maxAllowedRisk) {
      // Log evento de alto risco
      await this.auditService.logSecurityEvent({
        userId: payload.sub,
        tenantId: payload.tenantId,
        action: 'HIGH_RISK_ACCESS_BLOCKED',
        severity: 'high',
        metadata: {
          riskScore,
          maxAllowed: maxAllowedRisk,
          ip: this.getClientIP(request),
          userAgent: request.get('User-Agent')
        }
      });

      throw new UnauthorizedException('Access denied due to high risk score');
    }

    // Verificar rate limiting por usuário
    await this.checkUserRateLimit(payload.sub, request);

    // Verificar horário de acesso se configurado
    this.validateAccessHours(user);
  }

  /**
   * Verificar rate limiting por usuário
   */
  private async checkUserRateLimit(userId: string, request: Request): Promise<void> {
    const rateLimitKey = `user_rate_limit:${userId}`;
    const maxRequestsPerMinute = this.configService.get<number>('security.userMaxRequestsPerMinute', 60);

    try {
      const currentCount = await this.cacheManager.get<number>(rateLimitKey) || 0;
      
      if (currentCount >= maxRequestsPerMinute) {
        throw new UnauthorizedException('User rate limit exceeded');
      }
      
      await this.cacheManager.set(rateLimitKey, currentCount + 1, 60); // 1 minuto
    } catch (error) {
      if (error instanceof UnauthorizedException) {
        throw error;
      }
      this.logger.warn(`User rate limit check failed: ${error.message}`);
    }
  }

  /**
   * Validar horário de acesso
   */
  private validateAccessHours(user: any): void {
    const enforceAccessHours = this.configService.get<boolean>('security.enforceAccessHours', false);
    if (!enforceAccessHours || !user.allowedAccessHours) {
      return;
    }

    const now = new Date();
    const currentHour = now.getHours();
    const currentDay = now.getDay(); // 0 = domingo, 1 = segunda, etc.

    // Implementar lógica de horário de acesso baseada na configuração do usuário
    // Por exemplo: user.allowedAccessHours = { start: 8, end: 18, days: [1,2,3,4,5] }
    if (user.allowedAccessHours.days && !user.allowedAccessHours.days.includes(currentDay)) {
      throw new UnauthorizedException('Access not allowed on this day');
    }

    if (currentHour < user.allowedAccessHours.start || currentHour > user.allowedAccessHours.end) {
      throw new UnauthorizedException('Access not allowed at this time');
    }
  }

  /**
   * Determinar se MFA é necessário
   */
  private shouldRequireMFA(riskScore: number, user: any): boolean {
    const mfaRiskThreshold = this.configService.get<number>('security.mfaRiskThreshold', 0.6);
    const forceMFAForAdmin = this.configService.get<boolean>('security.forceMFAForAdmin', true);

    // MFA obrigatório para administradores
    if (forceMFAForAdmin && user.roles?.includes('admin')) {
      return true;
    }

    // MFA baseado em risco
    if (riskScore > mfaRiskThreshold) {
      return true;
    }

    // MFA baseado em preferências do usuário
    if (user.preferences?.requireMFA) {
      return true;
    }

    return false;
  }

  /**
   * Atualizar atividade do usuário
   */
  private async updateUserActivity(user: any, request: Request): Promise<void> {
    try {
      const activityData = {
        lastActivity: new Date(),
        lastIP: this.getClientIP(request),
        lastUserAgent: request.get('User-Agent')
      };

      // Atualizar em cache para performance
      const cacheKey = `user_activity:${user.id}`;
      await this.cacheManager.set(cacheKey, activityData, 3600); // 1 hora

      // Atualizar no banco de dados de forma assíncrona
      setImmediate(async () => {
        try {
          await this.iamService.updateUserActivity(user.id, activityData);
        } catch (error) {
          this.logger.warn(`Failed to update user activity: ${error.message}`);
        }
      });

    } catch (error) {
      this.logger.warn(`Failed to update user activity: ${error.message}`);
    }
  }

  /**
   * Log de autenticação bem-sucedida
   */
  private async logSuccessfulAuthentication(user: ValidatedUser, request: Request): Promise<void> {
    try {
      await this.auditService.logEvent({
        userId: user.id,
        tenantId: user.tenantId,
        action: 'JWT_AUTHENTICATION_SUCCESS',
        resource: 'Authentication',
        severity: 'info',
        metadata: {
          sessionId: user.sessionId,
          ip: this.getClientIP(request),
          userAgent: request.get('User-Agent'),
          riskScore: user.riskScore,
          requiresMFA: user.requiresMFA
        }
      });
    } catch (error) {
      this.logger.warn(`Failed to log successful authentication: ${error.message}`);
    }
  }

  /**
   * Log de autenticação falhada
   */
  private async logFailedAuthentication(
    payload: JwtPayload | null,
    request: Request,
    error: any
  ): Promise<void> {
    try {
      await this.auditService.logSecurityEvent({
        userId: payload?.sub || null,
        tenantId: payload?.tenantId || null,
        action: 'JWT_AUTHENTICATION_FAILED',
        severity: 'warning',
        metadata: {
          ip: this.getClientIP(request),
          userAgent: request.get('User-Agent'),
          error: error.message,
          tokenType: payload?.type,
          tokenIssued: payload?.iat ? new Date(payload.iat * 1000) : null
        }
      });
    } catch (auditError) {
      this.logger.error(`Failed to log failed authentication: ${auditError.message}`);
    }
  }

  /**
   * Obter IP real do cliente
   */
  private getClientIP(request: Request): string {
    const forwarded = request.get('X-Forwarded-For');
    const realIP = request.get('X-Real-IP');
    const cfConnectingIP = request.get('CF-Connecting-IP');
    
    if (cfConnectingIP) return cfConnectingIP;
    if (realIP) return realIP;
    if (forwarded) return forwarded.split(',')[0].trim();
    
    return request.ip || request.connection.remoteAddress || 'unknown';
  }
}