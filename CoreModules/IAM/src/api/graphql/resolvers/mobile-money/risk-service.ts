/**
 * Serviço de enriquecimento de dados de risco para transações Mobile Money
 * 
 * Integra-se com o RiskManagement para avaliação de transações em tempo real
 * Implementa verificações baseadas em ML e regras para detecção de fraudes
 * Compatível com requisitos de compliance AML/CFT globais
 */

import { tracer } from '../../../../observability/tracing';
import { logger } from '../../../../observability/logging';
import { metrics } from '../../../../observability/metrics';
import { DeviceService } from '../../../../services/device-service';
import { LocationService } from '../../../../services/location-service';
import { UserBehaviorService } from '../../../../services/user-behavior-service';

// Integrações com serviços de análise de risco
import { RiskScoreService } from '../../../../integration/risk-management/risk-score-service';
import { FraudDetectionService } from '../../../../integration/risk-management/fraud-detection-service';
import { TransactionPatternService } from '../../../../integration/risk-management/transaction-pattern-service';
import { AmlService } from '../../../../integration/risk-management/aml-service';

/**
 * Enriquece uma transação Mobile Money com dados de análise de risco
 */
export const enrichTransactionWithRiskData = async (
  transaction: any,
  context: any
): Promise<any> => {
  const span = tracer.startSpan('services.enrichTransactionWithRiskData');
  
  try {
    logger.info('Enriquecendo transação com dados de risco', {
      transactionId: transaction.referenceId,
      userId: transaction.userId,
      provider: transaction.provider,
      type: transaction.type,
      amount: transaction.amount,
      currency: transaction.currency,
      tenantId: transaction.tenantId
    });

    metrics.increment('risk.transaction_enrichment', {
      provider: transaction.provider,
      type: transaction.type,
      currency: transaction.currency,
      tenantId: transaction.tenantId
    });
    
    // Obter informações do dispositivo a partir do contexto
    const deviceInfo = await extractDeviceInfo(context);
    
    // Obter informações de localização
    const locationInfo = await extractLocationInfo(context);
    
    // Obter dados de comportamento do usuário
    const userBehaviorService = new UserBehaviorService(transaction.tenantId);
    const behaviorData = await userBehaviorService.getUserBehaviorProfile(transaction.userId);
    
    // Combinar dados para avaliação de risco
    const riskData = {
      transaction: {
        id: transaction.referenceId,
        amount: transaction.amount,
        currency: transaction.currency,
        provider: transaction.provider,
        type: transaction.type,
        timestamp: new Date(),
        phoneNumber: transaction.phoneNumber,
        userId: transaction.userId
      },
      user: {
        id: transaction.userId,
        behaviorProfile: behaviorData,
        riskHistory: await userBehaviorService.getUserRiskHistory(transaction.userId)
      },
      context: {
        device: deviceInfo,
        location: locationInfo,
        sessionData: context.session || {},
        channel: context.channel || 'API',
        ipAddress: context.ipAddress || context.ip,
        tenantId: transaction.tenantId,
        regionCode: transaction.regionCode
      }
    };

    // Analisar padrões de transação
    const transactionPatternService = new TransactionPatternService(transaction.tenantId);
    const patternAnalysis = await transactionPatternService.analyzeTransaction(riskData);
    
    // Avaliar risco de fraude
    const fraudDetectionService = new FraudDetectionService(transaction.tenantId);
    const fraudRiskAssessment = await fraudDetectionService.assessTransaction(riskData);
    
    // Verificar conformidade AML
    const amlService = new AmlService(transaction.tenantId);
    const amlCompliance = await amlService.checkTransactionCompliance({
      userId: transaction.userId,
      amount: transaction.amount,
      currency: transaction.currency,
      counterparty: transaction.phoneNumber,
      transactionType: transaction.type,
      timestamp: new Date(),
      regionCode: transaction.regionCode
    });
    
    // Calcular pontuação de risco geral
    const riskScoreService = new RiskScoreService(transaction.tenantId);
    const riskScore = await riskScoreService.calculateScore({
      fraudRisk: fraudRiskAssessment.riskScore,
      patternRisk: patternAnalysis.riskScore,
      amlRisk: amlCompliance.riskScore,
      transactionData: riskData.transaction,
      userData: riskData.user,
      contextData: riskData.context
    });

    // Adicionar dados de risco à transação
    const enrichedTransaction = {
      ...transaction,
      riskScore: riskScore.score,
      riskLevel: riskScore.level,
      riskAssessment: {
        score: riskScore.score,
        level: riskScore.level,
        factors: riskScore.factors,
        fraudRisk: fraudRiskAssessment.riskScore,
        patternRisk: patternAnalysis.riskScore,
        amlRisk: amlCompliance.riskScore,
        timestamp: new Date(),
        requiresReview: riskScore.score > 75,
        requiresApproval: riskScore.score > 85,
        automaticallyDeclined: riskScore.score > 95
      },
      complianceData: {
        ...transaction.complianceData,
        amlChecks: amlCompliance.checks,
        amlStatus: amlCompliance.status,
        sanctionsScreening: amlCompliance.sanctionsScreening,
        pefStatus: amlCompliance.pefStatus,
        watchlistMatches: amlCompliance.watchlistMatches
      }
    };

    logger.info('Transação enriquecida com dados de risco', {
      transactionId: transaction.referenceId,
      riskScore: riskScore.score,
      riskLevel: riskScore.level,
      fraudRisk: fraudRiskAssessment.riskScore,
      amlStatus: amlCompliance.status
    });

    // Aumentar métricas com base na avaliação de risco
    metrics.histogram('risk.transaction_score', riskScore.score, {
      provider: transaction.provider,
      type: transaction.type,
      riskLevel: riskScore.level,
      tenantId: transaction.tenantId
    });

    return enrichedTransaction;
  } catch (error) {
    logger.error('Erro ao enriquecer transação com dados de risco', {
      error: error.message,
      transactionId: transaction.referenceId,
      stack: error.stack
    });

    metrics.increment('risk.transaction_enrichment.error', {
      error: error.name,
      tenantId: transaction.tenantId
    });

    // Em caso de erro, retornar a transação original sem interromper o fluxo
    return {
      ...transaction,
      riskScore: 50, // Pontuação de risco padrão/média em caso de falha
      riskLevel: 'MEDIUM', // Nível de risco padrão em caso de falha
      riskAssessment: {
        score: 50,
        level: 'MEDIUM',
        factors: ['FALLBACK_RISK_ASSESSMENT'],
        timestamp: new Date(),
        requiresReview: false,
        requiresApproval: false,
        automaticallyDeclined: false
      }
    };
  } finally {
    span.end();
  }
};

/**
 * Extrai informações do dispositivo a partir do contexto
 */
const extractDeviceInfo = async (context: any): Promise<any> => {
  try {
    const deviceService = new DeviceService();
    
    if (context.deviceId) {
      return await deviceService.getDeviceInfo(context.deviceId);
    }
    
    // Se não houver deviceId específico, extrair informações do cabeçalho HTTP
    if (context.req?.headers) {
      const headers = context.req.headers;
      return {
        userAgent: headers['user-agent'] || '',
        deviceType: detectDeviceType(headers['user-agent'] || ''),
        fingerprint: headers['device-fingerprint'] || '',
        trusted: false // Dispositivo não identificado é considerado não confiável
      };
    }
    
    return {
      deviceType: 'UNKNOWN',
      trusted: false
    };
  } catch (error) {
    logger.error('Erro ao extrair informações do dispositivo', {
      error: error.message,
      stack: error.stack
    });
    
    return {
      deviceType: 'UNKNOWN',
      trusted: false,
      error: error.message
    };
  }
};

/**
 * Detecta o tipo de dispositivo a partir do User-Agent
 */
const detectDeviceType = (userAgent: string): string => {
  if (!userAgent) return 'UNKNOWN';
  
  userAgent = userAgent.toLowerCase();
  
  if (userAgent.includes('mobile') || userAgent.includes('android') || userAgent.includes('iphone')) {
    return 'MOBILE';
  }
  
  if (userAgent.includes('tablet') || userAgent.includes('ipad')) {
    return 'TABLET';
  }
  
  if (userAgent.includes('windows') || userAgent.includes('macintosh') || userAgent.includes('linux')) {
    return 'DESKTOP';
  }
  
  return 'OTHER';
};

/**
 * Extrai informações de localização a partir do contexto
 */
const extractLocationInfo = async (context: any): Promise<any> => {
  try {
    const locationService = new LocationService();
    
    if (context.locationId) {
      return await locationService.getLocationInfo(context.locationId);
    }
    
    // Se não houver locationId específico, extrair informações do endereço IP
    if (context.req?.ip || context.ip || context.ipAddress) {
      const ip = context.req?.ip || context.ip || context.ipAddress;
      return await locationService.getLocationByIp(ip);
    }
    
    return {
      country: 'UNKNOWN',
      riskLevel: 'MEDIUM'
    };
  } catch (error) {
    logger.error('Erro ao extrair informações de localização', {
      error: error.message,
      stack: error.stack
    });
    
    return {
      country: 'UNKNOWN',
      riskLevel: 'MEDIUM',
      error: error.message
    };
  }
};