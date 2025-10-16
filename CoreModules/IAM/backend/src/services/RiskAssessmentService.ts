/**
 * ============================================================================
 * INNOVABIZ IAM - Risk Assessment Service
 * ============================================================================
 * Version: 1.0.0
 * Date: 31/07/2025
 * Author: Equipe de Segurança INNOVABIZ
 * Classification: Confidencial - Interno
 * 
 * Description: Serviço de avaliação de risco para operações WebAuthn/FIDO2
 * Standards: W3C WebAuthn Level 3, FIDO2 CTAP2.1, NIST SP 800-63B
 * Compliance: PCI DSS 4.0, GDPR/LGPD, PSD2, ISO 27001
 * ============================================================================
 */

import { Logger } from 'winston';
import { Pool } from 'pg';
import { Redis } from 'ioredis';
import geoip from 'geoip-lite';
import UAParser from 'ua-parser-js';

import {
  WebAuthnContext,
  RiskAssessment,
  RiskFactor
} from '../types/webauthn';
import { webauthnMetrics } from '../metrics/webauthn';

/**
 * Serviço para avaliação de risco em operações WebAuthn
 */
export class RiskAssessmentService {
  private readonly logger: Logger;
  private readonly db: Pool;
  private readonly redis: Redis;

  constructor(logger: Logger, db: Pool, redis: Redis) {
    this.logger = logger;
    this.db = db;
    this.redis = redis;
  }

  /**
   * Avalia risco durante registro de credencial
   */
  async assessRegistrationRisk(
    context: WebAuthnContext,
    registrationInfo: any
  ): Promise<number> {
    try {
      const factors: RiskFactor[] = [];

      // Análise geográfica
      const geoRisk = await this.assessGeographicRisk(context);
      factors.push(...geoRisk);

      // Análise de dispositivo
      const deviceRisk = await this.assessDeviceRisk(context, registrationInfo);
      factors.push(...deviceRisk);

      // Análise comportamental
      const behaviorRisk = await this.assessBehavioralRisk(context, 'registration');
      factors.push(...behaviorRisk);

      // Análise de autenticador
      const authenticatorRisk = await this.assessAuthenticatorRisk(registrationInfo);
      factors.push(...authenticatorRisk);

      // Calcular score final
      const riskScore = this.calculateRiskScore(factors);

      // Registrar métricas
      webauthnMetrics.riskScoreDistribution.observe(
        { tenant_id: context.tenantId || 'unknown', operation_type: 'registration' },
        riskScore
      );

      if (riskScore > 0.7) {
        webauthnMetrics.highRiskEvents.inc({
          tenant_id: context.tenantId || 'unknown',
          risk_level: 'high',
          event_type: 'registration'
        });
      }

      this.logger.info('Registration risk assessment completed', {
        userId: context.userId,
        tenantId: context.tenantId,
        correlationId: context.correlationId,
        riskScore,
        factorCount: factors.length
      });

      return riskScore;

    } catch (error) {
      this.logger.error('Failed to assess registration risk', {
        userId: context.userId,
        tenantId: context.tenantId,
        correlationId: context.correlationId,
        error: error.message
      });

      // Retornar score médio em caso de erro
      return 0.5;
    }
  }

  /**
   * Avalia risco durante autenticação
   */
  async assessAuthenticationRisk(params: {
    userId: string;
    credentialId: string;
    context: WebAuthnContext;
    authenticationInfo: any;
    userVerified: boolean;
  }): Promise<number> {
    try {
      const { userId, credentialId, context, authenticationInfo, userVerified } = params;
      const factors: RiskFactor[] = [];

      // Análise geográfica
      const geoRisk = await this.assessGeographicRisk(context);
      factors.push(...geoRisk);

      // Análise de dispositivo
      const deviceRisk = await this.assessDeviceRisk(context, authenticationInfo);
      factors.push(...deviceRisk);

      // Análise comportamental
      const behaviorRisk = await this.assessBehavioralRisk(context, 'authentication');
      factors.push(...behaviorRisk);

      // Análise de padrão de uso
      const usageRisk = await this.assessUsagePatternRisk(userId, credentialId, context);
      factors.push(...usageRisk);

      // Análise de verificação de usuário
      const uvRisk = this.assessUserVerificationRisk(userVerified, context);
      factors.push(...uvRisk);

      // Análise de velocidade (velocity)
      const velocityRisk = await this.assessVelocityRisk(userId, context);
      factors.push(...velocityRisk);

      // Calcular score final
      const riskScore = this.calculateRiskScore(factors);

      // Registrar métricas
      webauthnMetrics.riskScoreDistribution.observe(
        { tenant_id: context.tenantId || 'unknown', operation_type: 'authentication' },
        riskScore
      );

      if (riskScore > 0.7) {
        webauthnMetrics.highRiskEvents.inc({
          tenant_id: context.tenantId || 'unknown',
          risk_level: 'high',
          event_type: 'authentication'
        });
      }

      this.logger.info('Authentication risk assessment completed', {
        userId,
        credentialId,
        tenantId: context.tenantId,
        correlationId: context.correlationId,
        riskScore,
        factorCount: factors.length,
        userVerified
      });

      return riskScore;

    } catch (error) {
      this.logger.error('Failed to assess authentication risk', {
        userId: params.userId,
        credentialId: params.credentialId,
        tenantId: params.context.tenantId,
        correlationId: params.context.correlationId,
        error: error.message
      });

      // Retornar score médio em caso de erro
      return 0.5;
    }
  }

  /**
   * Análise de risco geográfico
   */
  private async assessGeographicRisk(context: WebAuthnContext): Promise<RiskFactor[]> {
    const factors: RiskFactor[] = [];

    try {
      // Obter informações geográficas do IP
      const geoInfo = geoip.lookup(context.ipAddress);
      
      if (geoInfo) {
        // Verificar se é um país de alto risco
        const highRiskCountries = ['CN', 'RU', 'KP', 'IR']; // Exemplo
        if (highRiskCountries.includes(geoInfo.country)) {
          factors.push({
            type: 'geographic_high_risk_country',
            description: `Access from high-risk country: ${geoInfo.country}`,
            weight: 0.3,
            value: 1.0,
            impact: 'negative'
          });
        }

        // Verificar distância da localização usual
        const usualLocation = await this.getUserUsualLocation(context.userId, context.tenantId);
        if (usualLocation) {
          const distance = this.calculateDistance(
            geoInfo.ll[0], geoInfo.ll[1],
            usualLocation.latitude, usualLocation.longitude
          );

          if (distance > 1000) { // Mais de 1000km
            factors.push({
              type: 'geographic_unusual_location',
              description: `Access from unusual location (${distance.toFixed(0)}km away)`,
              weight: 0.2,
              value: Math.min(distance / 10000, 1.0), // Normalizar para 0-1
              impact: 'negative'
            });
          }
        }

        // Verificar se é um proxy/VPN conhecido
        if (await this.isKnownProxy(context.ipAddress)) {
          factors.push({
            type: 'geographic_proxy_vpn',
            description: 'Access through known proxy/VPN',
            weight: 0.25,
            value: 0.8,
            impact: 'negative'
          });
        }
      }

    } catch (error) {
      this.logger.warn('Failed to assess geographic risk', {
        ipAddress: context.ipAddress,
        error: error.message
      });
    }

    return factors;
  }

  /**
   * Análise de risco de dispositivo
   */
  private async assessDeviceRisk(context: WebAuthnContext, authInfo: any): Promise<RiskFactor[]> {
    const factors: RiskFactor[] = [];

    try {
      // Analisar User Agent
      const parser = new UAParser(context.userAgent);
      const result = parser.getResult();

      // Verificar se é um dispositivo conhecido
      const isKnownDevice = await this.isKnownDevice(context.userId, context.deviceFingerprint);
      if (!isKnownDevice) {
        factors.push({
          type: 'device_unknown',
          description: 'Access from unknown device',
          weight: 0.2,
          value: 0.7,
          impact: 'negative'
        });
      }

      // Verificar se é um browser desatualizado
      if (result.browser.version) {
        const isOutdated = await this.isBrowserOutdated(result.browser.name, result.browser.version);
        if (isOutdated) {
          factors.push({
            type: 'device_outdated_browser',
            description: `Outdated browser: ${result.browser.name} ${result.browser.version}`,
            weight: 0.15,
            value: 0.6,
            impact: 'negative'
          });
        }
      }

      // Verificar se é um OS desatualizado
      if (result.os.version) {
        const isOutdated = await this.isOSOutdated(result.os.name, result.os.version);
        if (isOutdated) {
          factors.push({
            type: 'device_outdated_os',
            description: `Outdated OS: ${result.os.name} ${result.os.version}`,
            weight: 0.15,
            value: 0.6,
            impact: 'negative'
          });
        }
      }

      // Verificar características do autenticador
      if (authInfo.credentialDeviceType === 'multiDevice') {
        factors.push({
          type: 'device_multi_device_authenticator',
          description: 'Multi-device authenticator used',
          weight: 0.1,
          value: 0.3,
          impact: 'negative'
        });
      }

    } catch (error) {
      this.logger.warn('Failed to assess device risk', {
        userAgent: context.userAgent,
        error: error.message
      });
    }

    return factors;
  }

  /**
   * Análise de risco comportamental
   */
  private async assessBehavioralRisk(context: WebAuthnContext, operationType: string): Promise<RiskFactor[]> {
    const factors: RiskFactor[] = [];

    try {
      // Verificar padrão de horário
      const hour = new Date().getHours();
      const isUnusualTime = await this.isUnusualTime(context.userId, hour);
      if (isUnusualTime) {
        factors.push({
          type: 'behavioral_unusual_time',
          description: `Access at unusual time: ${hour}:00`,
          weight: 0.1,
          value: 0.4,
          impact: 'negative'
        });
      }

      // Verificar frequência de tentativas
      const recentAttempts = await this.getRecentAttempts(context.userId, operationType, 300); // 5 minutos
      if (recentAttempts > 5) {
        factors.push({
          type: 'behavioral_high_frequency',
          description: `High frequency attempts: ${recentAttempts} in 5 minutes`,
          weight: 0.3,
          value: Math.min(recentAttempts / 10, 1.0),
          impact: 'negative'
        });
      }

      // Verificar padrão de navegação
      const sessionDuration = await this.getSessionDuration(context.sessionId);
      if (sessionDuration && sessionDuration < 10) { // Menos de 10 segundos
        factors.push({
          type: 'behavioral_short_session',
          description: `Very short session duration: ${sessionDuration}s`,
          weight: 0.15,
          value: 0.6,
          impact: 'negative'
        });
      }

    } catch (error) {
      this.logger.warn('Failed to assess behavioral risk', {
        userId: context.userId,
        operationType,
        error: error.message
      });
    }

    return factors;
  }

  /**
   * Análise de risco do autenticador
   */
  private async assessAuthenticatorRisk(authInfo: any): Promise<RiskFactor[]> {
    const factors: RiskFactor[] = [];

    try {
      // Verificar se o AAGUID está na lista de bloqueados
      const isBlocked = await this.isBlockedAAGUID(authInfo.aaguid);
      if (isBlocked) {
        factors.push({
          type: 'authenticator_blocked',
          description: `Blocked authenticator AAGUID: ${authInfo.aaguid}`,
          weight: 0.5,
          value: 1.0,
          impact: 'negative'
        });
      }

      // Verificar se é um autenticador de baixa segurança
      const securityLevel = await this.getAuthenticatorSecurityLevel(authInfo.aaguid);
      if (securityLevel === 'low') {
        factors.push({
          type: 'authenticator_low_security',
          description: 'Low security authenticator',
          weight: 0.2,
          value: 0.7,
          impact: 'negative'
        });
      }

      // Verificar se suporta user verification
      if (!authInfo.userVerified) {
        factors.push({
          type: 'authenticator_no_user_verification',
          description: 'No user verification performed',
          weight: 0.15,
          value: 0.5,
          impact: 'negative'
        });
      }

    } catch (error) {
      this.logger.warn('Failed to assess authenticator risk', {
        aaguid: authInfo.aaguid,
        error: error.message
      });
    }

    return factors;
  }

  /**
   * Análise de padrão de uso
   */
  private async assessUsagePatternRisk(
    userId: string,
    credentialId: string,
    context: WebAuthnContext
  ): Promise<RiskFactor[]> {
    const factors: RiskFactor[] = [];

    try {
      // Verificar última utilização
      const lastUsage = await this.getLastCredentialUsage(credentialId);
      if (lastUsage) {
        const daysSinceLastUse = (Date.now() - lastUsage.getTime()) / (1000 * 60 * 60 * 24);
        if (daysSinceLastUse > 90) { // Mais de 90 dias
          factors.push({
            type: 'usage_long_dormancy',
            description: `Credential unused for ${daysSinceLastUse.toFixed(0)} days`,
            weight: 0.2,
            value: Math.min(daysSinceLastUse / 365, 1.0),
            impact: 'negative'
          });
        }
      }

      // Verificar padrão de IP
      const usualIPs = await this.getUserUsualIPs(userId);
      if (usualIPs.length > 0 && !usualIPs.includes(context.ipAddress)) {
        factors.push({
          type: 'usage_unusual_ip',
          description: 'Access from unusual IP address',
          weight: 0.15,
          value: 0.6,
          impact: 'negative'
        });
      }

    } catch (error) {
      this.logger.warn('Failed to assess usage pattern risk', {
        userId,
        credentialId,
        error: error.message
      });
    }

    return factors;
  }

  /**
   * Análise de risco de verificação de usuário
   */
  private assessUserVerificationRisk(userVerified: boolean, context: WebAuthnContext): RiskFactor[] {
    const factors: RiskFactor[] = [];

    if (!userVerified) {
      factors.push({
        type: 'user_verification_not_performed',
        description: 'User verification was not performed',
        weight: 0.2,
        value: 0.6,
        impact: 'negative'
      });
    } else {
      factors.push({
        type: 'user_verification_performed',
        description: 'User verification was performed',
        weight: 0.1,
        value: 0.3,
        impact: 'positive'
      });
    }

    return factors;
  }

  /**
   * Análise de risco de velocidade
   */
  private async assessVelocityRisk(userId: string, context: WebAuthnContext): Promise<RiskFactor[]> {
    const factors: RiskFactor[] = [];

    try {
      // Verificar tentativas por minuto
      const attemptsPerMinute = await this.getAttemptsInWindow(userId, 60); // 1 minuto
      if (attemptsPerMinute > 10) {
        factors.push({
          type: 'velocity_high_attempts_per_minute',
          description: `High velocity: ${attemptsPerMinute} attempts per minute`,
          weight: 0.4,
          value: Math.min(attemptsPerMinute / 20, 1.0),
          impact: 'negative'
        });
      }

      // Verificar tentativas por hora
      const attemptsPerHour = await this.getAttemptsInWindow(userId, 3600); // 1 hora
      if (attemptsPerHour > 50) {
        factors.push({
          type: 'velocity_high_attempts_per_hour',
          description: `High velocity: ${attemptsPerHour} attempts per hour`,
          weight: 0.3,
          value: Math.min(attemptsPerHour / 100, 1.0),
          impact: 'negative'
        });
      }

    } catch (error) {
      this.logger.warn('Failed to assess velocity risk', {
        userId,
        error: error.message
      });
    }

    return factors;
  }

  /**
   * Calcula score de risco final baseado nos fatores
   */
  private calculateRiskScore(factors: RiskFactor[]): number {
    if (factors.length === 0) return 0.1; // Score baixo se não há fatores

    let weightedSum = 0;
    let totalWeight = 0;

    for (const factor of factors) {
      const adjustedValue = factor.impact === 'positive' ? (1 - factor.value) : factor.value;
      weightedSum += factor.weight * adjustedValue;
      totalWeight += factor.weight;
    }

    const baseScore = totalWeight > 0 ? weightedSum / totalWeight : 0;
    
    // Aplicar função de suavização para evitar scores extremos
    const smoothedScore = Math.max(0.01, Math.min(0.99, baseScore));
    
    return Math.round(smoothedScore * 100) / 100; // Arredondar para 2 casas decimais
  }

  /**
   * Métodos auxiliares
   */

  private calculateDistance(lat1: number, lon1: number, lat2: number, lon2: number): number {
    const R = 6371; // Raio da Terra em km
    const dLat = this.deg2rad(lat2 - lat1);
    const dLon = this.deg2rad(lon2 - lon1);
    const a = 
      Math.sin(dLat/2) * Math.sin(dLat/2) +
      Math.cos(this.deg2rad(lat1)) * Math.cos(this.deg2rad(lat2)) * 
      Math.sin(dLon/2) * Math.sin(dLon/2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1-a));
    return R * c;
  }

  private deg2rad(deg: number): number {
    return deg * (Math.PI/180);
  }

  private async getUserUsualLocation(userId: string, tenantId: string): Promise<any> {
    // Implementar busca da localização usual do usuário
    const cacheKey = `risk:usual_location:${userId}:${tenantId}`;
    const cached = await this.redis.get(cacheKey);
    if (cached) return JSON.parse(cached);

    // Buscar do banco de dados (implementar)
    return null;
  }

  private async isKnownProxy(ipAddress: string): Promise<boolean> {
    // Implementar verificação de proxy/VPN
    const cacheKey = `risk:proxy_check:${ipAddress}`;
    const cached = await this.redis.get(cacheKey);
    if (cached) return cached === 'true';

    // Verificar em serviços externos ou base de dados local
    return false;
  }

  private async isKnownDevice(userId: string, deviceFingerprint?: string): Promise<boolean> {
    if (!deviceFingerprint) return false;
    
    const result = await this.db.query(
      'SELECT 1 FROM user_devices WHERE user_id = $1 AND device_fingerprint = $2 LIMIT 1',
      [userId, deviceFingerprint]
    );
    
    return result.rows.length > 0;
  }

  private async isBrowserOutdated(browserName: string, version: string): Promise<boolean> {
    // Implementar verificação de versões desatualizadas
    // Por simplicidade, retornando false
    return false;
  }

  private async isOSOutdated(osName: string, version: string): Promise<boolean> {
    // Implementar verificação de versões desatualizadas
    // Por simplicidade, retornando false
    return false;
  }

  private async isUnusualTime(userId: string, hour: number): Promise<boolean> {
    // Implementar análise de padrão de horário do usuário
    return hour < 6 || hour > 23; // Simplificado
  }

  private async getRecentAttempts(userId: string, operationType: string, windowSeconds: number): Promise<number> {
    const result = await this.db.query(`
      SELECT COUNT(*) as count 
      FROM webauthn_events 
      WHERE user_id = $1 AND event_type LIKE $2 AND created_at > NOW() - INTERVAL '${windowSeconds} seconds'
    `, [userId, `${operationType}%`]);
    
    return parseInt(result.rows[0].count);
  }

  private async getSessionDuration(sessionId?: string): Promise<number | null> {
    if (!sessionId) return null;
    
    // Implementar busca de duração da sessão
    return null;
  }

  private async isBlockedAAGUID(aaguid: string): Promise<boolean> {
    const result = await this.db.query(
      'SELECT 1 FROM blocked_aaguids WHERE aaguid = $1 LIMIT 1',
      [aaguid]
    );
    
    return result.rows.length > 0;
  }

  private async getAuthenticatorSecurityLevel(aaguid: string): Promise<string> {
    const result = await this.db.query(
      'SELECT security_level FROM fido_metadata WHERE aaguid = $1 LIMIT 1',
      [aaguid]
    );
    
    return result.rows[0]?.security_level || 'medium';
  }

  private async getLastCredentialUsage(credentialId: string): Promise<Date | null> {
    const result = await this.db.query(
      'SELECT last_used_at FROM webauthn_credentials WHERE id = $1',
      [credentialId]
    );
    
    return result.rows[0]?.last_used_at || null;
  }

  private async getUserUsualIPs(userId: string): Promise<string[]> {
    const result = await this.db.query(`
      SELECT DISTINCT ip_address 
      FROM webauthn_events 
      WHERE user_id = $1 AND created_at > NOW() - INTERVAL '30 days'
      GROUP BY ip_address 
      HAVING COUNT(*) > 5
      ORDER BY COUNT(*) DESC 
      LIMIT 10
    `, [userId]);
    
    return result.rows.map(row => row.ip_address);
  }

  private async getAttemptsInWindow(userId: string, windowSeconds: number): Promise<number> {
    const result = await this.db.query(`
      SELECT COUNT(*) as count 
      FROM webauthn_events 
      WHERE user_id = $1 AND created_at > NOW() - INTERVAL '${windowSeconds} seconds'
    `, [userId]);
    
    return parseInt(result.rows[0].count);
  }
}