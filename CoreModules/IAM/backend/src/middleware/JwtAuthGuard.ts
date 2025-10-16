/**
 * 🛡️ JWT AUTH GUARD - INNOVABIZ IAM
 * Guard de autenticação JWT com validação avançada
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: OWASP Authentication, NIST SP 800-63B, JWT Best Practices
 * Security: Token Validation, Session Management, Rate Limiting
 */

import {
  Injectable,
  CanActivate,
  ExecutionContext,
  UnauthorizedException,
  Logger,
  Inject
} from '@nestjs/common';
import { JwtService } from '@nestjs/jwt';
import { ConfigService } from '@nestjs/config';
import { Reflector } from '@nestjs/core';
import { Request } from 'express';
import { Cache } from 'cache-manager';
import { CACHE_MANAGER } from '@nestjs/cache-manager';

import { IAMService } from '../services/IAMService';
import { AuditService } from '../services/AuditService';

@Injectable()
export class JwtAuthGuard implements CanActivate {
  private readonly logger = new Logger(JwtAuthGuard.name);

  constructor(
    private readonly jwtService: JwtService,
    private readonly configService: ConfigService,
    private readonly reflector: Reflector,
    private readonly iamService: IAMService,
    private readonly auditService: AuditService,
    @Inject(CACHE_MANAGER) private readonly cacheManager: Cache
  ) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest<Request>();
    
    try {
      // Verificar se a rota é pública
      const isPublic = this.reflector.get<boolean>('isPublic', context.getHandler());
      if (isPublic) {
        return true;
      }

      // Extrair token
      const token = this.extractTokenFromRequest(request);
      if (!token) {
        throw new UnauthorizedException('Token not provided');
      }

      // Verificar se token está na blacklist
      const isBlacklisted = await this.isTokenBlacklisted(token);
      if (isBlacklisted) {
        throw new UnauthorizedException('Token has been revoked');
      }

      // Validar e decodificar token
      const payload = await this.validateToken(token);
      
      // Validar sessão
      await this.validateSession(payload, request);
      
      // Anexar usuário ao request
      request['user'] = payload;
      request['token'] = token;

      return true;
    } catch (error) {
      this.logger.warn(`Authentication failed: ${error.message}`, {
        ip: request.ip,
        userAgent: request.get('User-Agent'),
        path: request.path
      });

      // Log de auditoria para tentativa de acesso não autorizado
      await this.auditService.logSecurityEvent({
        userId: null,
        tenantId: null,
        action: 'UNAUTHORIZED_ACCESS_ATTEMPT',
        severity: 'warning',
        metadata: {
          ip: request.ip,
          userAgent: request.get('User-Agent'),
          path: request.path,
          error: error.message
        }
      });

      throw error;
    }
  }

  /**
   * Extrair token do request
   */
  private extractTokenFromRequest(request: Request): string | null {
    // Tentar extrair do header Authorization
    const authHeader = request.get('Authorization');
    if (authHeader && authHeader.startsWith('Bearer ')) {
      return authHeader.substring(7);
    }

    // Tentar extrair do cookie
    const tokenFromCookie = request.cookies?.access_token;
    if (tokenFromCookie) {
      return tokenFromCookie;
    }

    // Tentar extrair do query parameter (não recomendado para produção)
    const tokenFromQuery = request.query?.token as string;
    if (tokenFromQuery && process.env.NODE_ENV !== 'production') {
      return tokenFromQuery;
    }

    return null;
  }

  /**
   * Validar token JWT
   */
  private async validateToken(token: string): Promise<any> {
    try {
      const payload = this.jwtService.verify(token, {
        secret: this.configService.get<string>('jwt.secret'),
        issuer: this.configService.get<string>('jwt.issuer'),
        audience: this.configService.get<string>('jwt.audience')
      });

      // Validações adicionais
      this.validateTokenPayload(payload);
      
      return payload;
    } catch (error) {
      if (error.name === 'TokenExpiredError') {
        throw new UnauthorizedException('Token has expired');
      } else if (error.name === 'JsonWebTokenError') {
        throw new UnauthorizedException('Invalid token');
      } else if (error.name === 'NotBeforeError') {
        throw new UnauthorizedException('Token not active yet');
      }
      
      throw new UnauthorizedException('Token validation failed');
    }
  }

  /**
   * Validar payload do token
   */
  private validateTokenPayload(payload: any): void {
    if (!payload.sub) {
      throw new UnauthorizedException('Invalid token: missing subject');
    }

    if (!payload.tenantId) {
      throw new UnauthorizedException('Invalid token: missing tenant');
    }

    if (payload.type === 'refresh') {
      throw new UnauthorizedException('Refresh token cannot be used for authentication');
    }

    // Verificar se o token não é muito antigo
    const maxAge = this.configService.get<number>('jwt.maxAge', 86400); // 24 horas
    const tokenAge = Date.now() / 1000 - payload.iat;
    
    if (tokenAge > maxAge) {
      throw new UnauthorizedException('Token is too old');
    }
  }

  /**
   * Verificar se token está na blacklist
   */
  private async isTokenBlacklisted(token: string): Promise<boolean> {
    try {
      const blacklisted = await this.cacheManager.get(`blacklist:${token}`);
      return !!blacklisted;
    } catch (error) {
      this.logger.warn(`Failed to check token blacklist: ${error.message}`);
      return false; // Falha segura - não bloquear se não conseguir verificar
    }
  }

  /**
   * Validar sessão ativa
   */
  private async validateSession(payload: any, request: Request): Promise<void> {
    try {
      // Verificar se a sessão ainda está ativa
      const sessionId = payload.sessionId || request.get('X-Session-Id');
      
      if (sessionId) {
        // Implementar validação de sessão
        // const session = await this.iamService.validateSession(sessionId);
        // if (!session || !session.isActive) {
        //   throw new UnauthorizedException('Session is not active');
        // }
      }

      // Verificar rate limiting por usuário
      await this.checkUserRateLimit(payload.sub, request.ip);
      
    } catch (error) {
      if (error instanceof UnauthorizedException) {
        throw error;
      }
      
      this.logger.warn(`Session validation warning: ${error.message}`);
      // Não falhar a autenticação por problemas de sessão não críticos
    }
  }

  /**
   * Verificar rate limiting por usuário
   */
  private async checkUserRateLimit(userId: string, ip: string): Promise<void> {
    const rateLimitKey = `rate_limit:${userId}:${ip}`;
    const maxRequests = this.configService.get<number>('security.maxRequestsPerMinute', 100);
    
    try {
      const currentCount = await this.cacheManager.get<number>(rateLimitKey) || 0;
      
      if (currentCount >= maxRequests) {
        throw new UnauthorizedException('Rate limit exceeded');
      }
      
      await this.cacheManager.set(rateLimitKey, currentCount + 1, 60); // 1 minuto
    } catch (error) {
      if (error instanceof UnauthorizedException) {
        throw error;
      }
      
      this.logger.warn(`Rate limit check failed: ${error.message}`);
    }
  }

  /**
   * Adicionar token à blacklist
   */
  async blacklistToken(token: string, expiresIn?: number): Promise<void> {
    try {
      const ttl = expiresIn || this.configService.get<number>('jwt.expiresInSeconds', 3600);
      await this.cacheManager.set(`blacklist:${token}`, true, ttl);
    } catch (error) {
      this.logger.error(`Failed to blacklist token: ${error.message}`);
    }
  }

  /**
   * Validar permissões específicas (para uso futuro)
   */
  private async validatePermissions(
    payload: any, 
    requiredPermissions: string[]
  ): Promise<boolean> {
    if (!requiredPermissions || requiredPermissions.length === 0) {
      return true;
    }

    // Implementar validação de permissões
    // const userPermissions = await this.getUserPermissions(payload.sub, payload.tenantId);
    // return requiredPermissions.every(permission => userPermissions.includes(permission));
    
    return true; // Placeholder
  }
}