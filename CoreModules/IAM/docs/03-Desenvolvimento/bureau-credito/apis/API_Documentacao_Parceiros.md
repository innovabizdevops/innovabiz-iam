# Documentação de API - Bureau de Créditos INNOVABIZ

## Visão Geral

O serviço de Bureau de Créditos da INNOVABIZ fornece uma API abrangente para avaliação de risco financeiro, detecção de fraude e consulta de dados de crédito. Esta documentação destina-se aos parceiros técnicos que integrarão seus sistemas com nossa plataforma.

**Versão da API:** v1.0  
**Ambiente de Produção:** https://api.innovabiz.com/v1/bureau-credito  
**Ambiente de Sandbox:** https://sandbox-api.innovabiz.com/v1/bureau-credito

## Autenticação e Autorização

### Autenticação

Todas as solicitações à API devem incluir um token de acesso OAuth 2.0 no cabeçalho HTTP:

```http
Authorization: Bearer {seu_token_de_acesso}
```

Para obter um token de acesso:

1. Registre-se como parceiro através do portal de desenvolvedores: https://developers.innovabiz.com
2. Crie credenciais OAuth 2.0 no console do desenvolvedor
3. Solicite um token usando o fluxo de concessão de credenciais de cliente:

```http
POST https://auth.innovabiz.com/oauth2/token
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials&
client_id={seu_client_id}&
client_secret={seu_client_secret}&
scope=bureau_credito:read bureau_credito:write
```

### Autorização

O acesso às APIs é controlado através de escopos OAuth. Os escopos disponíveis são:

- `bureau_credito:read` - Permissão para consultar dados de crédito
- `bureau_credito:write` - Permissão para avaliar transações e atualizar dados
- `bureau_credito:admin` - Permissão para operações administrativas (apenas parceiros especiais)

## Formato de Dados

A API suporta dois formatos:

1. **REST** - Utilizando JSON para requisições e respostas
2. **GraphQL** - Oferecendo maior flexibilidade para consultas complexas

Para todas as solicitações REST, utilize o cabeçalho:

```http
Content-Type: application/json
Accept: application/json
```

Para solicitações GraphQL:

```http
Content-Type: application/json
Accept: application/json
```

### Convenções de Data e Hora

Todos os timestamps são fornecidos no formato ISO 8601 em UTC:

```
YYYY-MM-DDThh:mm:ss.sssZ
```

Exemplo: `2025-08-07T14:30:00.000Z`

## Endpoints REST

### Avaliação de Transação

#### Requisição

```http
POST /transactions/evaluate
Content-Type: application/json
Authorization: Bearer {seu_token_de_acesso}

{
  "transactionId": "tx-123456789",
  "userId": "user-987654321",
  "tenantId": "tenant-12345",
  "documentType": "CPF",
  "documentNumber": "12345678901",
  "transactionType": "PAYMENT",
  "amount": 1500.00,
  "currency": "BRL",
  "channel": "mobile_app",
  "deviceId": "device-abcdef",
  "deviceFingerprint": "a1b2c3d4e5f6g7h8i9j0",
  "ipAddress": "192.168.1.100",
  "countryCode": "BR",
  "location": {
    "latitude": -23.5505,
    "longitude": -46.6333,
    "accuracy": 10.5
  },
  "timestamp": "2025-08-07T14:30:00.000Z",
  "options": {
    "performRiskAssessment": true,
    "performFraudDetection": true,
    "fetchCreditData": true,
    "creditProviderType": "BUREAU_CREDITO",
    "includeRawData": false
  },
  "userMetadata": {
    "registrationDate": "2024-01-15T10:00:00.000Z",
    "verificationLevel": "STRONG",
    "accountType": "PREMIUM"
  },
  "transactionMetadata": {
    "merchantId": "merchant-1234",
    "merchantCategory": "5499",
    "recurringPayment": false
  }
}
```

#### Resposta (200 OK)

```json
{
  "evaluationId": "eval-987654321",
  "transactionId": "tx-123456789",
  "userId": "user-987654321",
  "timestamp": "2025-08-07T14:30:01.234Z",
  "approved": true,
  "requiresReview": false,
  "requiresAdditionalVerification": false,
  "recommendedActions": ["APPROVE"],
  "overallRiskLevel": "LOW",
  "overallRiskScore": 25.5,
  "identityVerificationLevel": "STRONG",
  "processingTimeMs": 325,
  "riskAssessment": {
    "overallScore": 25.5,
    "riskLevel": "LOW",
    "recommendedActions": ["APPROVE"],
    "evaluationDetails": [
      {
        "ruleId": "amount_threshold_check",
        "ruleName": "Verificação de Limite de Valor",
        "category": "TRANSACTION_AMOUNT",
        "score": 15.0,
        "details": "Valor dentro dos limites esperados para o perfil",
        "triggered": false
      },
      {
        "ruleId": "velocity_check",
        "ruleName": "Verificação de Velocidade de Transação",
        "category": "TRANSACTION_VELOCITY",
        "score": 10.5,
        "details": "Frequência de transação normal",
        "triggered": false
      }
    ],
    "dataQuality": {
      "completeness": 0.95,
      "reliability": 0.92,
      "missingFields": []
    },
    "thresholds": {
      "low": 30.0,
      "medium": 50.0,
      "high": 70.0,
      "veryHigh": 85.0,
      "critical": 95.0
    },
    "decisionTime": 125,
    "requiresManualReview": false
  },
  "fraudDetection": {
    "fraudDetected": false,
    "overallConfidenceLevel": "VERY_LOW",
    "overallScore": 12.3,
    "triggeredRules": [],
    "evaluatedRuleCount": 15,
    "processingTimeMs": 180,
    "suggestedActions": ["APPROVE"],
    "requiresManualReview": false,
    "fraudTypes": []
  },
  "creditData": {
    "requestId": "cr-12345",
    "providerType": "BUREAU_CREDITO",
    "responseStatus": "SUCCESS",
    "creditScore": 750,
    "creditScoreScale": {
      "min": 300,
      "max": 900,
      "provider": "BUREAU_CREDITO",
      "category": "BOM"
    },
    "riskCategory": "BAIXO",
    "activeCreditAccounts": 3,
    "totalCreditLimit": 25000,
    "totalBalance": 8500,
    "creditUtilizationRate": 34,
    "paymentDefaults": [],
    "identityVerification": {
      "verified": true,
      "score": 85,
      "details": "Verificação realizada com sucesso"
    },
    "addressVerification": {
      "verified": true,
      "score": 90,
      "details": "Endereço confirmado"
    },
    "dataCompleteness": 95,
    "dataFreshness": "2025-08-05T10:15:30.000Z",
    "processingTimeMs": 220
  },
  "errors": []
}
```

### Consulta de Dados de Crédito

#### Requisição

```http
POST /credit-data/query
Content-Type: application/json
Authorization: Bearer {seu_token_de_acesso}

{
  "userId": "user-987654321",
  "tenantId": "tenant-12345",
  "documentType": "CPF",
  "documentNumber": "12345678901",
  "name": "João da Silva",
  "birthDate": "1985-03-15T00:00:00.000Z",
  "requestReason": "LOAN_APPLICATION",
  "includeRawData": false,
  "providerType": "BUREAU_CREDITO"
}
```

#### Resposta (200 OK)

```json
{
  "requestId": "cr-67890",
  "providerType": "BUREAU_CREDITO",
  "responseStatus": "SUCCESS",
  "creditScore": 750,
  "creditScoreScale": {
    "min": 300,
    "max": 900,
    "provider": "BUREAU_CREDITO",
    "category": "BOM"
  },
  "riskCategory": "BAIXO",
  "activeCreditAccounts": 3,
  "totalCreditLimit": 25000,
  "totalBalance": 8500,
  "creditUtilizationRate": 34,
  "paymentDefaults": [],
  "identityVerification": {
    "verified": true,
    "score": 85,
    "details": "Verificação realizada com sucesso"
  },
  "addressVerification": {
    "verified": true,
    "score": 90,
    "details": "Endereço confirmado"
  },
  "dataCompleteness": 95,
  "dataFreshness": "2025-08-05T10:15:30.000Z",
  "processingTimeMs": 220,
  "errors": []
}
```

### Verificação de Saúde do Serviço

#### Requisição

```http
GET /health
Authorization: Bearer {seu_token_de_acesso}
```

#### Resposta (200 OK)

```json
{
  "status": "UP",
  "version": "1.0.0",
  "components": [
    {
      "name": "risk_assessment",
      "status": "UP",
      "details": "Serviço de avaliação de risco operacional",
      "latencyMs": 12
    },
    {
      "name": "fraud_detection",
      "status": "UP",
      "details": "Serviço de detecção de fraude operacional",
      "latencyMs": 18
    },
    {
      "name": "bureau_credito_adapter",
      "status": "UP",
      "details": "Adaptador para Bureau de Crédito conectado",
      "latencyMs": 45
    }
  ],
  "timestamp": "2025-08-07T14:35:12.345Z"
}
```

## API GraphQL

Endpoint único para todas as operações GraphQL:

```http
POST /graphql
Content-Type: application/json
Authorization: Bearer {seu_token_de_acesso}
```

### Exemplos de Consultas GraphQL

#### Avaliação de Transação

```graphql
mutation EvaluateTransaction {
  evaluateTransaction(input: {
    transactionId: "tx-123456789",
    userId: "user-987654321",
    tenantId: "tenant-12345",
    documentType: "CPF",
    documentNumber: "12345678901",
    transactionType: PAYMENT,
    amount: 1500.00,
    currency: "BRL",
    channel: "mobile_app",
    deviceId: "device-abcdef",
    deviceFingerprint: "a1b2c3d4e5f6g7h8i9j0",
    ipAddress: "192.168.1.100",
    countryCode: "BR",
    location: {
      latitude: -23.5505,
      longitude: -46.6333,
      accuracy: 10.5
    },
    timestamp: "2025-08-07T14:30:00.000Z",
    options: {
      performRiskAssessment: true,
      performFraudDetection: true,
      fetchCreditData: true,
      creditProviderType: BUREAU_CREDITO,
      includeRawData: false
    }
  }) {
    evaluationId
    transactionId
    timestamp
    approved
    requiresReview
    overallRiskLevel
    overallRiskScore
    processingTimeMs
    riskAssessment {
      overallScore
      riskLevel
      recommendedActions
      evaluationDetails {
        ruleName
        category
        score
        triggered
      }
    }
    fraudDetection {
      fraudDetected
      overallConfidenceLevel
      overallScore
      processingTimeMs
    }
    creditData {
      creditScore
      creditScoreScale {
        min
        max
        category
      }
      riskCategory
    }
  }
}
```

#### Consulta de Dados de Crédito

```graphql
query GetCreditData {
  creditData(input: {
    userId: "user-987654321",
    tenantId: "tenant-12345",
    documentType: "CPF",
    documentNumber: "12345678901",
    name: "João da Silva",
    birthDate: "1985-03-15T00:00:00.000Z",
    includeRawData: false,
    providerType: BUREAU_CREDITO
  }) {
    requestId
    providerType
    responseStatus
    creditScore
    creditScoreScale {
      min
      max
      category
    }
    riskCategory
    activeCreditAccounts
    totalCreditLimit
    totalBalance
    creditUtilizationRate
    paymentDefaults {
      creditor
      amount
      currency
      daysOverdue
      date
    }
    identityVerification {
      verified
      score
      details
    }
    dataFreshness
    processingTimeMs
  }
}
```

#### Verificação de Saúde do Serviço

```graphql
query CheckHealth {
  healthCheck {
    status
    version
    components {
      name
      status
      details
      latencyMs
    }
    timestamp
  }
}
```

## Códigos de Erro

| Código | Descrição | Resolução Recomendada |
|--------|-----------|------------------------|
| 400 | Requisição inválida | Verifique se todos os campos obrigatórios foram fornecidos e estão no formato correto |
| 401 | Não autorizado | Verifique seu token de acesso ou solicite um novo |
| 403 | Acesso proibido | Verifique se você tem os escopos necessários para esta operação |
| 404 | Recurso não encontrado | Verifique se os identificadores estão corretos |
| 422 | Entidade não processável | Verifique se os valores estão dentro dos limites permitidos |
| 429 | Muitas requisições | Reduza a taxa de solicitações para respeitar os limites de throttling |
| 500 | Erro interno do servidor | Entre em contato com o suporte se o problema persistir |
| 503 | Serviço indisponível | Tente novamente mais tarde ou verifique o status do serviço |

## Limites de Taxa

O serviço implementa limites de taxa (throttling) para garantir a disponibilidade para todos os parceiros:

| Plano | Limite de Requisições | Período |
|-------|------------------------|---------|
| Básico | 10 | Por segundo |
| Standard | 50 | Por segundo |
| Premium | 200 | Por segundo |
| Enterprise | Personalizado | Personalizado |

Ao atingir o limite, as respostas incluirão os seguintes cabeçalhos:

```http
X-RateLimit-Limit: 50
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1628347854
```

## Melhores Práticas

1. **Cache Inteligente**: Armazene em cache dados de crédito para minimizar consultas repetidas em um curto período de tempo.

2. **Retry com Backoff Exponencial**: Implemente uma estratégia de retry com backoff exponencial para lidar com falhas transitórias.

3. **Validação Local**: Realize validações básicas antes de enviar para a API para economizar tempo e recursos.

4. **Monitoramento**: Implemente monitoramento de endpoints para detectar problemas de disponibilidade.

5. **Segurança**: Nunca exponha tokens de acesso ou dados sensíveis no frontend ou em código do cliente.

## Ambiente de Sandbox

O ambiente de sandbox está disponível para testes e desenvolvimento:

```
https://sandbox-api.innovabiz.com/v1/bureau-credito
```

Credenciais de teste para o sandbox:

```
Client ID: sandbox-test-client
Client Secret: sandbox-test-secret
```

O ambiente de sandbox contém dados de teste predefinidos para diferentes cenários:

| CPF | Cenário |
|-----|---------|
| 12345678900 | Usuário com bom score de crédito |
| 12345678901 | Usuário com score médio de crédito |
| 12345678902 | Usuário com score baixo de crédito |
| 12345678903 | Usuário com restrições e inadimplências |
| 12345678904 | Usuário não encontrado |

## Suporte

Para questões técnicas, entre em contato com nossa equipe de suporte:

- **Email**: api-support@innovabiz.com
- **Portal de Desenvolvedores**: https://developers.innovabiz.com
- **Documentação Completa**: https://docs.innovabiz.com/bureau-credito

## Versões da API

| Versão | Status | Fim do Suporte |
|--------|--------|----------------|
| v1.0 | Atual | N/A |
| Beta | Desenvolvimento | N/A |

## Mudanças Planejadas

- **Q3 2025**: Adição de novos provedores de dados de crédito
- **Q4 2025**: Suporte para análise comportamental avançada
- **Q1 2026**: Novo endpoint para detecção de fraude em tempo real