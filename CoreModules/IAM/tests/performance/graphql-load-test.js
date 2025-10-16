import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { SharedArray } from 'k6/data';
import encoding from 'k6/encoding';

// Métricas personalizadas
const failureRate = new Rate('graphql_failures');
const queryResponseTimes = new Trend('graphql_query_response_times');
const mutationResponseTimes = new Trend('graphql_mutation_response_times');

// Configurações do teste
export const options = {
  vus: __ENV.VUS || 50,
  duration: __ENV.DURATION || '5m',
  
  thresholds: {
    'graphql_failures': ['rate<0.02'], // Máximo 2% de falha
    'graphql_query_response_times': ['p(95)<800'], // 95% das queries abaixo de 800ms
    'graphql_mutation_response_times': ['p(95)<1000'], // 95% das mutations abaixo de 1000ms
    'http_reqs': ['rate>80'], // Mínimo 80 req/s
  },
  
  // Rampa de carga
  stages: [
    { duration: '30s', target: Math.round(__ENV.VUS * 0.2) },
    { duration: '1m', target: Math.round(__ENV.VUS * 0.5) },
    { duration: '2m', target: parseInt(__ENV.VUS) },
    { duration: '1m', target: parseInt(__ENV.VUS) },
    { duration: '30s', target: 0 },
  ],
};

// Dados para testes
const testQueries = new SharedArray('graphql_queries', function() {
  return [
    // Query para verificar status de conformidade
    {
      type: 'query',
      name: 'getComplianceStatus',
      query: `
        query GetComplianceStatus($userId: ID!, $regulationType: String!) {
          complianceStatus(userId: $userId, regulationType: $regulationType) {
            userId
            regulationType
            status
            lastValidated
            expiresAt
            validationDetails {
              validatorVersion
              result
              warnings
              requiredActions
            }
          }
        }
      `,
      variables: {
        userId: 'user-123456',
        regulationType: 'GDPR'
      }
    },
    // Query para verificar histórico de validação
    {
      type: 'query',
      name: 'getValidationHistory',
      query: `
        query GetValidationHistory($userId: ID!, $limit: Int!) {
          complianceValidationHistory(userId: $userId, limit: $limit) {
            records {
              timestamp
              regulationType
              operation
              status
              requester
              details
            }
            pagination {
              totalCount
              hasMore
            }
          }
        }
      `,
      variables: {
        userId: 'user-123456',
        limit: 10
      }
    },
    // Query para Dashboard de Conformidade
    {
      type: 'query',
      name: 'getComplianceDashboardData',
      query: `
        query GetComplianceDashboard($period: String!, $regulationType: String) {
          complianceDashboard(period: $period, regulationType: $regulationType) {
            summary {
              totalValidations
              passRate
              failRate
              averageResponseTime
            }
            byRegulationType {
              type
              count
              passRate
            }
            byRequestType {
              type
              count
              averageResponseTime
            }
            trends {
              date
              count
              passRate
            }
          }
        }
      `,
      variables: {
        period: 'LAST_7_DAYS',
        regulationType: null
      }
    }
  ];
});

const testMutations = new SharedArray('graphql_mutations', function() {
  return [
    // Mutation para validar conformidade
    {
      type: 'mutation',
      name: 'validateCompliance',
      query: `
        mutation ValidateCompliance($input: ComplianceValidationInput!) {
          validateCompliance(input: $input) {
            status
            validationId
            timestamp
            result {
              isCompliant
              regulationType
              warnings
              blockers
              suggestedActions
            }
          }
        }
      `,
      variables: {
        input: {
          userId: 'user-123456',
          regulationType: 'GDPR',
          operation: 'DATA_ACCESS',
          dataCategories: ['PERSONAL', 'FINANCIAL'],
          purpose: 'CREDIT_SCORING',
          requestMetadata: {
            requestId: 'req-12345',
            sourceSystem: 'BUREAU_CREDITO',
            requesterType: 'ORGANIZATION'
          }
        }
      }
    },
    // Mutation para atualizar preferências de conformidade
    {
      type: 'mutation',
      name: 'updateComplianceSettings',
      query: `
        mutation UpdateComplianceSettings($input: ComplianceSettingsInput!) {
          updateComplianceSettings(input: $input) {
            userId
            updated
            settings {
              gdprConsent
              lgpdConsent
              popiaConsent
              marketingConsent
              dataRetentionPeriod
              dataCategories
            }
          }
        }
      `,
      variables: {
        input: {
          userId: 'user-123456',
          settings: {
            gdprConsent: true,
            lgpdConsent: true,
            popiaConsent: false,
            marketingConsent: false,
            dataRetentionPeriod: 'STANDARD',
            dataCategories: ['PERSONAL', 'FINANCIAL']
          }
        }
      }
    },
    // Mutation para registrar consentimento
    {
      type: 'mutation',
      name: 'recordUserConsent',
      query: `
        mutation RecordUserConsent($input: UserConsentInput!) {
          recordUserConsent(input: $input) {
            userId
            consentId
            timestamp
            status
            expiresAt
          }
        }
      `,
      variables: {
        input: {
          userId: 'user-123456',
          regulationType: 'LGPD',
          consentType: 'DATA_PROCESSING',
          granted: true,
          purpose: ['CREDIT_ANALYSIS', 'RISK_ASSESSMENT'],
          dataCategories: ['PERSONAL', 'FINANCIAL', 'LOCATION'],
          metadata: {
            ipAddress: '192.168.1.1',
            userAgent: 'Mozilla/5.0',
            deviceId: 'device-abc123'
          }
        }
      }
    }
  ];
});

// Função para autenticação
function getAuthToken() {
  const credentials = encoding.b64encode('graphql-client:graphql-secret');
  const response = http.post(`${__ENV.API_BASE_URL}/auth/token`, {
    grant_type: 'client_credentials',
    scope: 'graphql:api'
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

// Função para executar query GraphQL
function executeGraphQLRequest(token, operationType, operation) {
  const startTime = new Date();
  const payload = JSON.stringify({
    query: operation.query,
    variables: operation.variables
  });
  
  const response = http.post(
    `${__ENV.API_BASE_URL}/graphql`,
    payload,
    {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    }
  );
  
  const duration = new Date() - startTime;
  
  // Registrar tempo de resposta
  if (operationType === 'query') {
    queryResponseTimes.add(duration);
  } else if (operationType === 'mutation') {
    mutationResponseTimes.add(duration);
  }
  
  const success = check(response, {
    'Status é 200': (r) => r.status === 200,
    'Não há erros na resposta': (r) => {
      const body = JSON.parse(r.body);
      return !body.errors;
    },
    'Dados presentes na resposta': (r) => {
      const body = JSON.parse(r.body);
      return body.data && body.data !== null;
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
  
  // 70% das requisições serão queries
  if (Math.random() <= 0.7) {
    group('GraphQL Queries', () => {
      const query = testQueries[Math.floor(Math.random() * testQueries.length)];
      
      const result = executeGraphQLRequest(token, 'query', query);
      
      if (__ENV.DEBUG) {
        console.log(`Query: ${query.name}, Status: ${result.status}, Duração: ${result.duration}ms`);
      }
    });
  } 
  // 30% das requisições serão mutations
  else {
    group('GraphQL Mutations', () => {
      const mutation = testMutations[Math.floor(Math.random() * testMutations.length)];
      
      const result = executeGraphQLRequest(token, 'mutation', mutation);
      
      if (__ENV.DEBUG) {
        console.log(`Mutation: ${mutation.name}, Status: ${result.status}, Duração: ${result.duration}ms`);
      }
    });
  }
  
  // Simular intervalo entre solicitações
  sleep(Math.random() * 2 + 0.5);
}