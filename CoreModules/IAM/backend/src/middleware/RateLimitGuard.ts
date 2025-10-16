/**
 * 🚦 RATE LIMIT GUARD - INNOVABIZ IAM
 * Guard avançado de limitação de taxa com algoritmos adaptativos
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: OWASP Rate Limiting, NIST SP 800-63B, DDoS Protection
 * Algorithms: Token Bucket, Sliding Window, Adaptive Rate Limiting
 */

import {
  Injectable,
  CanActivate,
  ExecutionContext,
  HttpException,
  HttpStatus,
  Logger,
  Inject
} from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { Reflector } from '@nestjs/core';
import { Request } from 'express';
import { Cache } from 'cache-manager';
import { CACHE_MANAGER } from '@nestjs/cache-manager';

import { AuditService } from '../services/AuditService';
import { RiskAssessmentService } from '../services/RiskAssessmentService';

// Interfaces para Rate Limiting
interface RateLimitConfig {
  windowMs: number;
  maxRequests: number;
  skipSuccessfulRequests?: boolean;
  skipFailedRequests?: boolean;
  keyGenerator?: (req: Request) => string;
  onLimitReached?: (req: Request) => void;
}

interface RateLimitInfo {
  totalHits: number;
  totalHitsInWindow: number;
  remainingPoints: number;
  msBeforeNext: number;
  isFirstInWindow: boolean;
}

interface AdaptiveRateLimitData {
  baseLimit: number;
  currentLimit: number;
  successRate: number;
  errorRate: number;
  lastAdjustment: number;
  riskScore: number;
}

@Injectable()
export class RateLimitGuard implements CanActivate {
  private readonly logger = new Logger(RateLimitGuard.name);

  constructor(
    private readonly configService: ConfigService,
    private readonly reflector: Reflector,
    private readonly auditService: AuditService,
    private readonly riskAssessmentService: RiskAssessmentService,
    @Inject(CACHE_MANAGER) private readonly cacheManager: Cache
  ) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest<Request>();
    
    try {
      // Verificar se rate limiting está habilitado
      const skipRateLimit = this.reflector.get<boolean>('skipRateLimit', context.getHandler());
      if (skipRateLimit) {
        return true;
      }

      // Obter configuração de rate limit para o endpoint
      const rateLimitConfig = this.getRateLimitConfig(context, request);
      
      // Gerar chave única para o rate limit
      const key = this.generateRateLimitKey(request, rateLimitConfig);
      
      // Verificar rate limit com algoritmo adaptativo
      const rateLimitResult = await this.checkRateLimit(key, rateLimitConfig, request);
      
      // Adicionar headers de rate limit na resposta
      this.addRateLimitHeaders(context, rateLimitResult);
      
      if (!rateLimitResult.allowed) {
        await this.handleRateLimitExceeded(request, key, rateLimitResult);
        throw new HttpException(
          {
            error: 'Rate limit exceeded',
            message: 'Too many requests, please try again later',
            retryAfter: Math.ceil(rateLimitResult.msBeforeNext / 1000)
          },
          HttpStatus.TOO_MANY_REQUESTS
        );
      }

      // Registrar request bem-sucedido
      await this.recordSuccessfulRequest(key, request);
      
      return true;
    } catch (error) {
      if (error instanceof HttpException) {
        throw error;
      }
      
      this.logger.error(`Rate limit check failed: ${error.message}`, error.stack);
      // Fail open em caso de erro no sistema de rate limiting
      return true;
    }
  }

  /**
   * Obter configuração de rate limit baseada no contexto
   */
  private getRateLimitConfig(context: ExecutionContext, request: Request): RateLimitConfig {
    // Configuração específica do endpoint
    const endpointConfig = this.reflector.get<Partial<RateLimitConfig>>('rateLimit', context.getHandler());
    
    // Configurações padrão baseadas no tipo de endpoint
    const defaultConfig = this.getDefaultConfigForEndpoint(request.path, request.method);
    
    // Configurações globais
    const globalConfig: RateLimitConfig = {
      windowMs: this.configService.get<number>('throttle.windowMs', 60000), // 1 minuto
      maxRequests: this.configService.get<number>('throttle.maxRequests', 100),
      skipSuccessfulRequests: false,
      skipFailedRequests: false
    };

    return {
      ...globalConfig,
      ...defaultConfig,
      ...endpointConfig
    };
  }

  /**
   * Configurações padrão baseadas no tipo de endpoint
   */
  private getDefaultConfigForEndpoint(path: string, method: string): Partial<RateLimitConfig> {
    // Endpoints de autenticação - mais restritivos
    if (path.includes('/auth/') || path.includes('/login') || path.includes('/webauthn/')) {
      return {
        windowMs: 60000, // 1 minuto
        maxRequests: method === 'POST' ? 10 : 20
      };
    }

    // Endpoints de criação de usuário - muito restritivos
    if (path.includes('/users') && method === 'POST') {
      return {
        windowMs: 300000, // 5 minutos
        maxRequests: 3
      };
    }

    // Endpoints de consulta - mais permissivos
    if (method === 'GET') {
      return {
        windowMs: 60000,
        maxRequests: 200
      };
    }

    // Endpoints de modificação - moderadamente restritivos
    if (['POST', 'PUT', 'PATCH', 'DELETE'].includes(method)) {
      return {
        windowMs: 60000,
        maxRequests: 50
      };
    }

    return {};
  }

  /**
   * Gerar chave única para rate limiting
   */
  private generateRateLimitKey(request: Request, config: RateLimitConfig): string {
    if (config.keyGenerator) {
      return config.keyGenerator(request);
    }

    // Estratégia de chave baseada em múltiplos fatores
    const ip = this.getClientIP(request);
    const userAgent = request.get('User-Agent') || 'unknown';
    const endpoint = `${request.method}:${request.path}`;
    const userId = request['user']?.sub || 'anonymous';
    
    // Hash simples para reduzir tamanho da chave
    const keyComponents = [ip, userId, endpoint].join(':');
    return `rate_limit:${Buffer.from(keyComponents).toString('base64').slice(0, 32)}`;
  }

  /**
   * Verificar rate limit com algoritmo adaptativo
   */
  private async checkRateLimit(
    key: string, 
    config: RateLimitConfig, 
    request: Request
  ): Promise<RateLimitInfo & { allowed: boolean }> {
    const now = Date.now();
    const windowStart = now - config.windowMs;

    // Obter dados atuais do rate limit
    const currentData = await this.getCurrentRateLimitData(key);
    
    // Aplicar algoritmo adaptativo
    const adaptiveLimit = await this.calculateAdaptiveLimit(key, config, request);
    
    // Limpar requests antigos (sliding window)
    const validRequests = currentData.requests.filter(timestamp => timestamp > windowStart);
    
    // Calcular informações do rate limit
    const totalHitsInWindow = validRequests.length;
    const remainingPoints = Math.max(0, adaptiveLimit - totalHitsInWindow);
    const isFirstInWindow = totalHitsInWindow === 0;
    
    let msBeforeNext = 0;
    if (remainingPoints === 0 && validRequests.length > 0) {
      msBeforeNext = validRequests[0] + config.windowMs - now;
    }

    const allowed = totalHitsInWindow < adaptiveLimit;

    // Atualizar dados se permitido
    if (allowed) {
      validRequests.push(now);
      await this.updateRateLimitData(key, {
        requests: validRequests,
        lastRequest: now,
        totalRequests: currentData.totalRequests + 1
      });
    }

    return {
      totalHits: currentData.totalRequests,
      totalHitsInWindow,
      remainingPoints,
      msBeforeNext: Math.max(0, msBeforeNext),
      isFirstInWindow,
      allowed
    };
  }

  /**
   * Calcular limite adaptativo baseado em comportamento e risco
   */
  private async calculateAdaptiveLimit(
    key: string, 
    config: RateLimitConfig, 
    request: Request
  ): Promise<number> {
    try {
      // Obter dados adaptativos existentes
      const adaptiveData = await this.getAdaptiveRateLimitData(key);
      
      // Avaliar risco do usuário/IP
      const riskScore = await this.assessRequestRisk(request);
      
      // Calcular limite baseado no risco
      let adaptiveLimit = adaptiveData.baseLimit;
      
      if (riskScore > 0.8) {
        // Alto risco - reduzir limite drasticamente
        adaptiveLimit = Math.floor(adaptiveData.baseLimit * 0.2);
      } else if (riskScore > 0.6) {
        // Risco médio - reduzir limite moderadamente
        adaptiveLimit = Math.floor(adaptiveData.baseLimit * 0.5);
      } else if (riskScore < 0.2 && adaptiveData.successRate > 0.95) {
        // Baixo risco e alta taxa de sucesso - aumentar limite
        adaptiveLimit = Math.floor(adaptiveData.baseLimit * 1.5);
      }

      // Limites mínimos e máximos
      adaptiveLimit = Math.max(1, Math.min(adaptiveLimit, config.maxRequests * 2));
      
      // Atualizar dados adaptativos
      await this.updateAdaptiveRateLimitData(key, {
        ...adaptiveData,
        currentLimit: adaptiveLimit,
        riskScore,
        lastAdjustment: Date.now()
      });

      return adaptiveLimit;
    } catch (error) {
      this.logger.warn(`Failed to calculate adaptive limit: ${error.message}`);
      return config.maxRequests;
    }
  }

  /**
   * Avaliar risco da requisição
   */
  private async assessRequestRisk(request: Request): Promise<number> {
    try {
      const ip = this.getClientIP(request);
      const userAgent = request.get('User-Agent') || '';
      const userId = request['user']?.sub;

      // Fatores de risco
      let riskScore = 0;

      // Verificar IP em blacklist ou com histórico ruim
      const ipRisk = await this.assessIPRisk(ip);
      riskScore += ipRisk * 0.4;

      // Verificar padrões suspeitos no User-Agent
      const userAgentRisk = this.assessUserAgentRisk(userAgent);
      riskScore += userAgentRisk * 0.2;

      // Se usuário autenticado, verificar perfil de risco
      if (userId) {
        const userRisk = await this.assessUserRisk(userId);
        riskScore += userRisk * 0.4;
      } else {
        // Usuários não autenticados têm risco base maior
        riskScore += 0.3;
      }

      return Math.min(1, riskScore);
    } catch (error) {
      this.logger.warn(`Risk assessment failed: ${error.message}`);
      return 0.5; // Risco médio como fallback
    }
  }

  /**
   * Avaliar risco do IP
   */
  private async assessIPRisk(ip: string): Promise<number> {
    try {
      // Verificar histórico de bloqueios
      const blockHistory = await this.cacheManager.get<number>(`ip_blocks:${ip}`) || 0;
      
      // Verificar frequência de requests falhados
      const failedRequests = await this.cacheManager.get<number>(`ip_failures:${ip}`) || 0;
      
      let risk = 0;
      if (blockHistory > 0) risk += 0.5;
      if (failedRequests > 10) risk += 0.3;
      
      return Math.min(1, risk);
    } catch (error) {
      return 0.2; // Risco baixo como fallback
    }
  }

  /**
   * Avaliar risco do User-Agent
   */
  private assessUserAgentRisk(userAgent: string): number {
    const suspiciousPatterns = [
      /bot/i,
      /crawler/i,
      /spider/i,
      /scraper/i,
      /curl/i,
      /wget/i,
      /python/i,
      /java/i
    ];

    const hasSuspiciousPattern = suspiciousPatterns.some(pattern => pattern.test(userAgent));
    
    if (!userAgent || userAgent.length < 10) return 0.8;
    if (hasSuspiciousPattern) return 0.6;
    
    return 0.1;
  }

  /**
   * Avaliar risco do usuário
   */
  private async assessUserRisk(userId: string): Promise<number> {
    try {
      // Usar serviço de avaliação de risco se disponível
      const riskProfile = await this.riskAssessmentService.getRiskProfile(userId, 'default');
      return riskProfile?.riskScore || 0.2;
    } catch (error) {
      return 0.2; // Risco baixo como fallback
    }
  }

  /**
   * Obter dados atuais do rate limit
   */
  private async getCurrentRateLimitData(key: string): Promise<{
    requests: number[];
    lastRequest: number;
    totalRequests: number;
  }> {
    try {
      const data = await this.cacheManager.get<any>(key);
      return data || {
        requests: [],
        lastRequest: 0,
        totalRequests: 0
      };
    } catch (error) {
      return {
        requests: [],
        lastRequest: 0,
        totalRequests: 0
      };
    }
  }

  /**
   * Atualizar dados do rate limit
   */
  private async updateRateLimitData(key: string, data: any): Promise<void> {
    try {
      const ttl = this.configService.get<number>('throttle.windowMs', 60000) * 2; // 2x janela
      await this.cacheManager.set(key, data, ttl / 1000);
    } catch (error) {
      this.logger.warn(`Failed to update rate limit data: ${error.message}`);
    }
  }

  /**
   * Obter dados adaptativos do rate limit
   */
  private async getAdaptiveRateLimitData(key: string): Promise<AdaptiveRateLimitData> {
    try {
      const adaptiveKey = `adaptive:${key}`;
      const data = await this.cacheManager.get<AdaptiveRateLimitData>(adaptiveKey);
      
      return data || {
        baseLimit: this.configService.get<number>('throttle.maxRequests', 100),
        currentLimit: this.configService.get<number>('throttle.maxRequests', 100),
        successRate: 1.0,
        errorRate: 0.0,
        lastAdjustment: Date.now(),
        riskScore: 0.2
      };
    } catch (error) {
      return {
        baseLimit: 100,
        currentLimit: 100,
        successRate: 1.0,
        errorRate: 0.0,
        lastAdjustment: Date.now(),
        riskScore: 0.2
      };
    }
  }

  /**
   * Atualizar dados adaptativos
   */
  private async updateAdaptiveRateLimitData(key: string, data: AdaptiveRateLimitData): Promise<void> {
    try {
      const adaptiveKey = `adaptive:${key}`;
      const ttl = 24 * 60 * 60; // 24 horas
      await this.cacheManager.set(adaptiveKey, data, ttl);
    } catch (error) {
      this.logger.warn(`Failed to update adaptive rate limit data: ${error.message}`);
    }
  }

  /**
   * Adicionar headers de rate limit na resposta
   */
  private addRateLimitHeaders(context: ExecutionContext, rateLimitInfo: RateLimitInfo): void {
    const response = context.switchToHttp().getResponse();
    
    response.setHeader('X-RateLimit-Limit', rateLimitInfo.totalHits.toString());
    response.setHeader('X-RateLimit-Remaining', rateLimitInfo.remainingPoints.toString());
    response.setHeader('X-RateLimit-Reset', new Date(Date.now() + rateLimitInfo.msBeforeNext).toISOString());
    
    if (rateLimitInfo.msBeforeNext > 0) {
      response.setHeader('Retry-After', Math.ceil(rateLimitInfo.msBeforeNext / 1000).toString());
    }
  }

  /**
   * Lidar com rate limit excedido
   */
  private async handleRateLimitExceeded(
    request: Request, 
    key: string, 
    rateLimitInfo: RateLimitInfo
  ): Promise<void> {
    const ip = this.getClientIP(request);
    const userAgent = request.get('User-Agent') || 'unknown';
    const userId = request['user']?.sub;

    // Log de auditoria
    await this.auditService.logSecurityEvent({
      userId: userId || null,
      tenantId: request['user']?.tenantId || null,
      action: 'RATE_LIMIT_EXCEEDED',
      severity: 'warning',
      metadata: {
        ip,
        userAgent,
        endpoint: `${request.method} ${request.path}`,
        rateLimitKey: key,
        totalHitsInWindow: rateLimitInfo.totalHitsInWindow,
        limit: rateLimitInfo.totalHits
      }
    });

    // Incrementar contador de bloqueios para o IP
    const blockKey = `ip_blocks:${ip}`;
    const currentBlocks = await this.cacheManager.get<number>(blockKey) || 0;
    await this.cacheManager.set(blockKey, currentBlocks + 1, 3600); // 1 hora

    this.logger.warn(`Rate limit exceeded`, {
      ip,
      userId,
      endpoint: `${request.method} ${request.path}`,
      totalHitsInWindow: rateLimitInfo.totalHitsInWindow
    });
  }

  /**
   * Registrar request bem-sucedido
   */
  private async recordSuccessfulRequest(key: string, request: Request): Promise<void> {
    try {
      // Atualizar estatísticas de sucesso para algoritmo adaptativo
      const adaptiveData = await this.getAdaptiveRateLimitData(key);
      const newSuccessRate = (adaptiveData.successRate * 0.9) + (1.0 * 0.1); // Média móvel
      
      await this.updateAdaptiveRateLimitData(key, {
        ...adaptiveData,
        successRate: newSuccessRate,
        errorRate: Math.max(0, adaptiveData.errorRate * 0.95) // Decaimento do erro
      });
    } catch (error) {
      this.logger.warn(`Failed to record successful request: ${error.message}`);
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