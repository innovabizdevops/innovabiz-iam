import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { SharedArray } from 'k6/data';
import encoding from 'k6/encoding';

// Métricas personalizadas
const failureRate = new Rate('failures');
const gdprValidationTimes = new Trend('gdpr_validation_times');
const lgpdValidationTimes = new Trend('lgpd_validation_times');
const popiaValidationTimes = new Trend('popia_validation_times');

// Configurações do teste
export const options = {
  // Parâmetros definidos via variáveis de ambiente
  vus: __ENV.VUS || 50,
  duration: __ENV.DURATION || '5m',
  
  thresholds: {
    // Validadores não podem ter mais que 1% de falhas
    'failures': ['rate<0.01'],
    // 95% das validações devem ser concluídas em menos de 500ms
    'gdpr_validation_times': ['p(95)<500'],
    'lgpd_validation_times': ['p(95)<500'],
    'popia_validation_times': ['p(95)<500'],
    // Taxa de resposta (requisições por segundo)
    'http_reqs': ['rate>100'],
  },
  
  // Rampa de usuários para simular carga crescente
  stages: [
    { duration: '30s', target: Math.round(__ENV.VUS * 0.2) },  // Início com 20% dos usuários
    { duration: '1m', target: Math.round(__ENV.VUS * 0.5) },   // Subir para 50% em 1 minuto
    { duration: '2m', target: parseInt(__ENV.VUS) },           // Chegar a 100% em 2 minutos
    { duration: '1m', target: parseInt(__ENV.VUS) },           // Manter carga máxima por 1 minuto
    { duration: '30s', target: 0 },                            // Reduzir para 0
  ],
};

// Dados de teste
const testData = new SharedArray('compliance_test_data', function() {
  return [
    // GDPR (Europa)
    {
      type: 'gdpr',
      payload: {
        operation: 'data_access',
        userData: {
          userId: 'EU12345',
          email: 'user@example.com',
          region: 'EU',
          dataCategories: ['personal', 'financial'],
          consentStatus: true,
          consentTimestamp: new Date().toISOString(),
          processingPurpose: 'credit_scoring'
        },
        requestMetadata: {
          requestId: 'req-gdpr-12345',
          timestamp: new Date().toISOString(),
          requesterType: 'data_controller',
          requesterEntityId: 'finance-app-eu',
          requesterRegion: 'EU'
        }
      }
    },
    // LGPD (Brasil)
    {
      type: 'lgpd',
      payload: {
        operation: 'data_processing',
        userData: {
          userId: 'BR54321',
          email: 'usuario@exemplo.com.br',
          region: 'BR',
          dataCategories: ['personal', 'financial', 'biometric'],
          consentStatus: true,
          consentTimestamp: new Date().toISOString(),
          processingPurpose: 'credit_analysis'
        },
        requestMetadata: {
          requestId: 'req-lgpd-54321',
          timestamp: new Date().toISOString(),
          requesterType: 'data_operator',
          requesterEntityId: 'fintech-br',
          requesterRegion: 'BR'
        }
      }
    },
    // POPIA (África do Sul)
    {
      type: 'popia',
      payload: {
        operation: 'data_sharing',
        userData: {
          userId: 'ZA98765',
          email: 'user@example.co.za',
          region: 'ZA',
          dataCategories: ['personal', 'financial'],
          consentStatus: true,
          consentTimestamp: new Date().toISOString(),
          processingPurpose: 'credit_bureau_reporting'
        },
        requestMetadata: {
          requestId: 'req-popia-98765',
          timestamp: new Date().toISOString(),
          requesterType: 'credit_bureau',
          requesterEntityId: 'bureau-za',
          requesterRegion: 'ZA'
        }
      }
    }
  ];
});

// Função auxiliar para autenticação
function getAuthToken() {
  const credentials = encoding.b64encode('test-client:test-secret');
  const response = http.post(`${__ENV.API_BASE_URL}/auth/token`, {
    grant_type: 'client_credentials',
    scope: 'compliance:validate'
  }, {
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
      'Authorization': `Basic ${credentials}`
    }
  });
  
  if (response.status === 200) {
    const token = JSON.parse(response.body).access_token;
    return token;
  } else {
    console.error(`Falha ao obter token: ${response.status} ${response.body}`);
    failureRate.add(true);
    return null;
  }
}

// Função para validar conformidade
function validateCompliance(token, data) {
  const startTime = new Date();
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  };
  
  let endpoint;
  switch (data.type) {
    case 'gdpr':
      endpoint = '/api/v1/compliance/gdpr/validate';
      break;
    case 'lgpd':
      endpoint = '/api/v1/compliance/lgpd/validate';
      break;
    case 'popia':
      endpoint = '/api/v1/compliance/popia/validate';
      break;
    default:
      endpoint = '/api/v1/compliance/validate';
  }
  
  const response = http.post(
    `${__ENV.API_BASE_URL}${endpoint}`,
    JSON.stringify(data.payload),
    { headers }
  );
  
  const duration = new Date() - startTime;
  
  // Registrar a duração nas métricas específicas
  switch (data.type) {
    case 'gdpr':
      gdprValidationTimes.add(duration);
      break;
    case 'lgpd':
      lgpdValidationTimes.add(duration);
      break;
    case 'popia':
      popiaValidationTimes.add(duration);
      break;
  }
  
  // Verificar resultados
  const success = check(response, {
    'Status é 200': (r) => r.status === 200,
    'Resposta contém resultado de validação': (r) => {
      const body = JSON.parse(r.body);
      return body.validationResult !== undefined && body.validationDetails !== undefined;
    }
  });
  
  failureRate.add(!success);
  
  return {
    status: response.status,
    body: JSON.parse(response.body),
    duration
  };
}

// Teste principal
export default function() {
  const token = getAuthToken();
  if (!token) return;
  
  group('Validação de Conformidade Individual', () => {
    // Selecionar um item aleatório de dados de teste
    const testItem = testData[Math.floor(Math.random() * testData.length)];
    
    const result = validateCompliance(token, testItem);
    
    // Log para debug (apenas em ambiente de desenvolvimento)
    if (__ENV.DEBUG) {
      console.log(`Tipo: ${testItem.type}, Status: ${result.status}, Duração: ${result.duration}ms`);
    }
  });
  
  group('Validação de Conformidade em Lote', () => {
    const batchPayload = {
      requests: testData.map(item => ({
        regulationType: item.type,
        ...item.payload
      }))
    };
    
    const startTime = new Date();
    const response = http.post(
      `${__ENV.API_BASE_URL}/api/v1/compliance/batch-validate`,
      JSON.stringify(batchPayload),
      {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        }
      }
    );
    
    const duration = new Date() - startTime;
    
    check(response, {
      'Status de lote é 200': (r) => r.status === 200,
      'Resposta de lote contém todos os resultados': (r) => {
        const body = JSON.parse(r.body);
        return body.results && body.results.length === testData.length;
      },
      'Tempo de resposta de lote < 1.5s': (r) => duration < 1500
    });
  });
  
  // Simulação de intervalo entre solicitações de usuário
  sleep(Math.random() * 3 + 1);
}