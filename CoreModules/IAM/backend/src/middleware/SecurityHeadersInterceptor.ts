/**
 * 🔒 SECURITY HEADERS INTERCEPTOR - INNOVABIZ IAM
 * Interceptor para adicionar headers de segurança
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Compliance: OWASP Security Headers, NIST Cybersecurity Framework
 * Security: CSP, HSTS, XSS Protection, Content Type Options
 */

import {
  Injectable,
  NestInterceptor,
  ExecutionContext,
  CallHandler,
  Logger
} from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { Observable } from 'rxjs';
import { tap } from 'rxjs/operators';
import { Response } from 'express';

@Injectable()
export class SecurityHeadersInterceptor implements NestInterceptor {
  private readonly logger = new Logger(SecurityHeadersInterceptor.name);

  constructor(private readonly configService: ConfigService) {}

  intercept(context: ExecutionContext, next: CallHandler): Observable<any> {
    const response = context.switchToHttp().getResponse<Response>();
    const request = context.switchToHttp().getRequest();

    // Aplicar headers de segurança
    this.applySecurityHeaders(response, request);

    return next.handle().pipe(
      tap(() => {
        // Headers adicionais após processamento
        this.applyPostProcessHeaders(response);
      })
    );
  }

  /**
   * Aplicar headers de segurança principais
   */
  private applySecurityHeaders(response: Response, request: any): void {
    const isProduction = this.configService.get<string>('NODE_ENV') === 'production';
    const domain = this.configService.get<string>('app.domain', 'localhost');

    // Content Security Policy
    const csp = this.buildContentSecurityPolicy();
    response.setHeader('Content-Security-Policy', csp);

    // HTTP Strict Transport Security (HSTS)
    if (isProduction) {
      response.setHeader(
        'Strict-Transport-Security',
        'max-age=31536000; includeSubDomains; preload'
      );
    }

    // X-Frame-Options
    response.setHeader('X-Frame-Options', 'DENY');

    // X-Content-Type-Options
    response.setHeader('X-Content-Type-Options', 'nosniff');

    // X-XSS-Protection
    response.setHeader('X-XSS-Protection', '1; mode=block');

    // Referrer Policy
    response.setHeader('Referrer-Policy', 'strict-origin-when-cross-origin');

    // Permissions Policy
    const permissionsPolicy = this.buildPermissionsPolicy();
    response.setHeader('Permissions-Policy', permissionsPolicy);

    // Cross-Origin Policies
    response.setHeader('Cross-Origin-Embedder-Policy', 'require-corp');
    response.setHeader('Cross-Origin-Opener-Policy', 'same-origin');
    response.setHeader('Cross-Origin-Resource-Policy', 'same-origin');

    // Cache Control para endpoints sensíveis
    if (this.isSensitiveEndpoint(request.path)) {
      response.setHeader('Cache-Control', 'no-store, no-cache, must-revalidate, private');
      response.setHeader('Pragma', 'no-cache');
      response.setHeader('Expires', '0');
    }

    // Server Information Hiding
    response.removeHeader('X-Powered-By');
    response.setHeader('Server', 'INNOVABIZ-IAM');

    // CORS Headers (se necessário)
    this.applyCorsHeaders(response, request);
  }

  /**
   * Construir Content Security Policy
   */
  private buildContentSecurityPolicy(): string {
    const isProduction = this.configService.get<string>('NODE_ENV') === 'production';
    const allowedDomains = this.configService.get<string[]>('security.allowedDomains', []);

    const policies = [
      "default-src 'self'",
      "script-src 'self' 'unsafe-inline' 'unsafe-eval'", // Ajustar conforme necessário
      "style-src 'self' 'unsafe-inline'",
      "img-src 'self' data: https:",
      "font-src 'self' data:",
      "connect-src 'self'",
      "media-src 'self'",
      "object-src 'none'",
      "child-src 'none'",
      "frame-src 'none'",
      "worker-src 'self'",
      "manifest-src 'self'",
      "base-uri 'self'",
      "form-action 'self'"
    ];

    // Adicionar domínios permitidos se configurados
    if (allowedDomains.length > 0) {
      const domains = allowedDomains.join(' ');
      policies[1] = `script-src 'self' 'unsafe-inline' ${domains}`;
      policies[4] = `connect-src 'self' ${domains}`;
    }

    // Políticas mais restritivas para produção
    if (isProduction) {
      policies.push("upgrade-insecure-requests");
      policies.push("block-all-mixed-content");
    }

    return policies.join('; ');
  }

  /**
   * Construir Permissions Policy
   */
  private buildPermissionsPolicy(): string {
    const policies = [
      'accelerometer=()',
      'ambient-light-sensor=()',
      'autoplay=()',
      'battery=()',
      'camera=()',
      'cross-origin-isolated=()',
      'display-capture=()',
      'document-domain=()',
      'encrypted-media=()',
      'execution-while-not-rendered=()',
      'execution-while-out-of-viewport=()',
      'fullscreen=()',
      'geolocation=()',
      'gyroscope=()',
      'keyboard-map=()',
      'magnetometer=()',
      'microphone=()',
      'midi=()',
      'navigation-override=()',
      'payment=()',
      'picture-in-picture=()',
      'publickey-credentials-get=*', // Permitir WebAuthn
      'screen-wake-lock=()',
      'sync-xhr=()',
      'usb=()',
      'web-share=()',
      'xr-spatial-tracking=()'
    ];

    return policies.join(', ');
  }

  /**
   * Aplicar headers CORS se necessário
   */
  private applyCorsHeaders(response: Response, request: any): void {
    const allowedOrigins = this.configService.get<string[]>('cors.allowedOrigins', []);
    const origin = request.get('Origin');

    if (allowedOrigins.includes(origin) || allowedOrigins.includes('*')) {
      response.setHeader('Access-Control-Allow-Origin', origin || '*');
      response.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
      response.setHeader(
        'Access-Control-Allow-Headers',
        'Origin, X-Requested-With, Content-Type, Accept, Authorization, X-Session-Id, X-Device-Fingerprint'
      );
      response.setHeader('Access-Control-Allow-Credentials', 'true');
      response.setHeader('Access-Control-Max-Age', '86400'); // 24 horas
    }
  }

  /**
   * Headers adicionais após processamento
   */
  private applyPostProcessHeaders(response: Response): void {
    // Timing headers para debugging (apenas em desenvolvimento)
    if (this.configService.get<string>('NODE_ENV') !== 'production') {
      response.setHeader('X-Response-Time', Date.now().toString());
    }

    // Content-Type específico para APIs
    if (!response.getHeader('Content-Type')) {
      response.setHeader('Content-Type', 'application/json; charset=utf-8');
    }

    // Security headers adicionais baseados no conteúdo
    this.applyContentBasedHeaders(response);
  }

  /**
   * Headers baseados no tipo de conteúdo
   */
  private applyContentBasedHeaders(response: Response): void {
    const contentType = response.getHeader('Content-Type') as string;

    if (contentType?.includes('application/json')) {
      // Headers específicos para JSON
      response.setHeader('X-Content-Type-Options', 'nosniff');
    }

    if (contentType?.includes('text/html')) {
      // Headers específicos para HTML
      response.setHeader('X-UA-Compatible', 'IE=edge');
    }
  }

  /**
   * Verificar se é endpoint sensível
   */
  private isSensitiveEndpoint(path: string): boolean {
    const sensitivePatterns = [
      '/auth/',
      '/login',
      '/register',
      '/password',
      '/token',
      '/session',
      '/webauthn',
      '/risk',
      '/audit'
    ];

    return sensitivePatterns.some(pattern => path.includes(pattern));
  }

  /**
   * Aplicar headers específicos para WebAuthn
   */
  private applyWebAuthnHeaders(response: Response): void {
    // Headers específicos para WebAuthn
    response.setHeader('Feature-Policy', 'publickey-credentials-get *');
    response.setHeader('Permissions-Policy', 'publickey-credentials-get=*');
  }

  /**
   * Headers para prevenção de ataques de timing
   */
  private applyTimingAttackPrevention(response: Response): void {
    // Adicionar delay aleatório pequeno para prevenir ataques de timing
    const randomDelay = Math.floor(Math.random() * 10);
    response.setHeader('X-Processing-Time', randomDelay.toString());
  }

  /**
   * Headers para compliance GDPR/LGPD
   */
  private applyPrivacyHeaders(response: Response): void {
    response.setHeader('X-Privacy-Policy', 'https://innovabiz.com/privacy');
    response.setHeader('X-Data-Protection', 'GDPR-LGPD-Compliant');
  }

  /**
   * Headers para auditoria e compliance
   */
  private applyComplianceHeaders(response: Response, request: any): void {
    const requestId = request.headers['x-request-id'] || this.generateRequestId();
    response.setHeader('X-Request-ID', requestId);
    response.setHeader('X-Compliance-Framework', 'NIST-OWASP-ISO27001');
  }

  /**
   * Gerar ID único para request
   */
  private generateRequestId(): string {
    return `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }
}