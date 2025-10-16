/**
 * @file risk-management-mock.ts
 * @description Mocks para testes de integração com o módulo RiskManagement
 * @author INNOVABIZ Development Team
 * @copyright 2025 INNOVABIZ
 * @version 1.0.0
 */

/**
 * Mock para uma avaliação de risco padrão (baixo risco)
 */
export const mockStandardRiskAssessment = {
  assessmentId: 'f8c4e51d-3c7b-4a12-8b85-5d689ad7b23c',
  userId: 'a8f7e52c-3f0d-4c12-9a85-4d789ad7b23c',
  tenantId: 'test-tenant',
  timestamp: new Date().toISOString(),
  sessionId: '5f9e8a3d-1b9c-4e8f-8b1c-3d9e8a5f1b9c',
  riskLevel: 'low',
  score: 0.25,
  factors: [
    {
      name: 'location',
      value: 'Luanda, Angola',
      score: 0.2,
      weight: 0.3
    },
    {
      name: 'device',
      value: 'known',
      score: 0.1,
      weight: 0.25
    },
    {
      name: 'time_pattern',
      value: 'normal',
      score: 0.15,
      weight: 0.15
    },
    {
      name: 'auth_method',
      value: 'password',
      score: 0.3,
      weight: 0.3
    }
  ],
  actions: [],
  recommendations: {
    authLevel: 'standard',
    monitoring: 'normal',
    restrictions: []
  }
};

/**
 * Mock para uma avaliação de risco elevado
 */
export const mockHighRiskAssessment = {
  assessmentId: 'a2b3c4d5-6e7f-8a9b-0c1d-2e3f4a5b6c7d',
  userId: 'a8f7e52c-3f0d-4c12-9a85-4d789ad7b23c',
  tenantId: 'test-tenant',
  timestamp: new Date().toISOString(),
  sessionId: '7a8b9c0d-1e2f-3a4b-5c6d-7e8f9a0b1c2d',
  riskLevel: 'high',
  score: 0.85,
  factors: [
    {
      name: 'location',
      value: 'São Paulo, Brasil',
      score: 0.7,
      weight: 0.3,
      reason: 'Localização incomum para este usuário'
    },
    {
      name: 'device',
      value: 'unknown',
      score: 0.9,
      weight: 0.25,
      reason: 'Dispositivo não reconhecido'
    },
    {
      name: 'time_pattern',
      value: 'unusual',
      score: 0.8,
      weight: 0.15,
      reason: 'Horário fora do padrão de uso'
    },
    {
      name: 'auth_method',
      value: 'password',
      score: 0.9,
      weight: 0.3,
      reason: 'Método de autenticação fraco para o risco detectado'
    }
  ],
  actions: [
    {
      type: 'STEP_UP_AUTH',
      required: true,
      description: 'Solicitar autenticação adicional por FIDO2'
    },
    {
      type: 'NOTIFY',
      required: true,
      description: 'Notificar equipe de segurança'
    },
    {
      type: 'LIMIT_ACCESS',
      required: false,
      description: 'Limitar acesso a funções sensíveis'
    }
  ],
  recommendations: {
    authLevel: 'strong',
    monitoring: 'enhanced',
    restrictions: [
      {
        feature: 'financial_transactions',
        limit: 500.00,
        currency: 'USD',
        duration: 24 // horas
      }
    ]
  }
};

/**
 * Mock para um relatório de evento de autenticação
 */
export const mockAuthEventReport = {
  eventId: 'c1d2e3f4-5a6b-7c8d-9e0f-1a2b3c4d5e6f',
  userId: 'a8f7e52c-3f0d-4c12-9a85-4d789ad7b23c',
  tenantId: 'test-tenant',
  eventType: 'AUTH_SUCCESS',
  timestamp: new Date().toISOString(),
  sessionId: '5f9e8a3d-1b9c-4e8f-8b1c-3d9e8a5f1b9c',
  ipAddress: '41.223.112.245',
  userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
  authMethod: 'webauthn',
  deviceId: 'device-id-123',
  success: true,
  failureReason: null,
  riskAssessment: {
    id: 'f8c4e51d-3c7b-4a12-8b85-5d689ad7b23c',
    level: 'low',
    score: 0.25
  },
  contextData: {
    location: 'Luanda, Angola',
    authenticationType: 'webauthn',
    authenticator: {
      type: 'platform',
      attestation: 'none'
    },
    previousLogin: '2025-01-01T12:00:00Z'
  }
};

/**
 * Mock para uma falha de autenticação com alto risco
 */
export const mockHighRiskAuthFailure = {
  eventId: 'd2e3f4a5-6b7c-8d9e-0f1a-2b3c4d5e6f7a',
  userId: 'a8f7e52c-3f0d-4c12-9a85-4d789ad7b23c',
  tenantId: 'test-tenant',
  eventType: 'AUTH_FAILURE',
  timestamp: new Date().toISOString(),
  sessionId: '7a8b9c0d-1e2f-3a4b-5c6d-7e8f9a0b1c2d',
  ipAddress: '185.189.103.254',
  userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
  authMethod: 'password',
  deviceId: 'unknown-device-xyz',
  success: false,
  failureReason: 'INVALID_CREDENTIALS',
  attemptCount: 3,
  riskAssessment: {
    id: 'a2b3c4d5-6e7f-8a9b-0c1d-2e3f4a5b6c7d',
    level: 'high',
    score: 0.85
  },
  contextData: {
    location: 'São Paulo, Brasil',
    authenticationType: 'password',
    previousLogin: '2025-01-01T12:00:00Z'
  },
  actionsTaken: [
    {
      action: 'ACCOUNT_LOCKOUT',
      timestamp: new Date().toISOString(),
      duration: 30, // minutos
      reason: 'Múltiplas falhas de autenticação em contexto de alto risco'
    },
    {
      action: 'SECURITY_NOTIFICATION',
      timestamp: new Date().toISOString(),
      channels: ['email', 'sms']
    }
  ]
};

/**
 * Gerador de respostas simuladas para o Risk Management
 * @param params Parâmetros para gerar uma resposta simulada
 * @returns Uma avaliação de risco simulada
 */
export function generateMockRiskResponse(params: {
  userId: string;
  tenantId: string;
  riskLevel?: 'low' | 'medium' | 'high';
  location?: string;
  deviceKnown?: boolean;
  authMethod?: string;
  timeUnusual?: boolean;
}) {
  const {
    userId,
    tenantId,
    riskLevel = 'low',
    location = 'Luanda, Angola',
    deviceKnown = true,
    authMethod = 'password',
    timeUnusual = false
  } = params;

  const assessmentId = `risk-${Date.now()}-${Math.floor(Math.random() * 1000)}`;
  const sessionId = `session-${Date.now()}-${Math.floor(Math.random() * 1000)}`;
  
  // Determina o score baseado no nível de risco
  let score: number;
  switch (riskLevel) {
    case 'low':
      score = 0.1 + Math.random() * 0.3; // 0.1 a 0.4
      break;
    case 'medium':
      score = 0.4 + Math.random() * 0.3; // 0.4 a 0.7
      break;
    case 'high':
      score = 0.7 + Math.random() * 0.3; // 0.7 a 1.0
      break;
  }

  // Constrói os fatores de risco
  const factors = [
    {
      name: 'location',
      value: location,
      score: riskLevel === 'high' ? 0.8 : 0.2,
      weight: 0.3,
      ...(riskLevel === 'high' ? { reason: 'Localização incomum' } : {})
    },
    {
      name: 'device',
      value: deviceKnown ? 'known' : 'unknown',
      score: deviceKnown ? 0.1 : 0.9,
      weight: 0.25,
      ...(!deviceKnown ? { reason: 'Dispositivo não reconhecido' } : {})
    },
    {
      name: 'time_pattern',
      value: timeUnusual ? 'unusual' : 'normal',
      score: timeUnusual ? 0.8 : 0.15,
      weight: 0.15,
      ...(timeUnusual ? { reason: 'Horário fora do padrão de uso' } : {})
    },
    {
      name: 'auth_method',
      value: authMethod,
      score: authMethod === 'password' ? (riskLevel === 'high' ? 0.9 : 0.3) : 0.1,
      weight: 0.3,
      ...(authMethod === 'password' && riskLevel === 'high' 
         ? { reason: 'Método de autenticação fraco para o risco detectado' } 
         : {})
    }
  ];

  // Define ações baseadas no nível de risco
  const actions = [];
  if (riskLevel === 'high') {
    actions.push({
      type: 'STEP_UP_AUTH',
      required: true,
      description: 'Solicitar autenticação adicional por FIDO2'
    });
    actions.push({
      type: 'NOTIFY',
      required: true,
      description: 'Notificar equipe de segurança'
    });
  } else if (riskLevel === 'medium') {
    actions.push({
      type: 'STEP_UP_AUTH',
      required: false,
      description: 'Considerar autenticação adicional'
    });
  }

  // Define recomendações baseadas no nível de risco
  const recommendations = {
    authLevel: riskLevel === 'high' ? 'strong' : riskLevel === 'medium' ? 'enhanced' : 'standard',
    monitoring: riskLevel === 'high' ? 'enhanced' : 'normal',
    restrictions: [] as any[]
  };

  // Adiciona restrições para risco alto
  if (riskLevel === 'high') {
    recommendations.restrictions.push({
      feature: 'financial_transactions',
      limit: 500.00,
      currency: 'USD',
      duration: 24 // horas
    });
  }

  return {
    assessmentId,
    userId,
    tenantId,
    timestamp: new Date().toISOString(),
    sessionId,
    riskLevel,
    score,
    factors,
    actions,
    recommendations
  };
}