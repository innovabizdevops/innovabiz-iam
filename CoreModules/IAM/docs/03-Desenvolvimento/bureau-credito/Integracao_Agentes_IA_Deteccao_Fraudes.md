# Integração com Agentes IA para Detecção de Fraudes Adaptativa

## Visão Geral

O módulo de **Detecção de Fraudes Adaptativa** é um componente avançado do Bureau de Crédito que utiliza múltiplos agentes de Inteligência Artificial para identificar e prevenir atividades fraudulentas em solicitações de crédito e transações financeiras. O sistema é adaptativo, aprendendo continuamente com novos dados e ajustando-se aos padrões emergentes de fraude.

Este documento fornece orientações técnicas para integradores externos que desejam utilizar ou estender a funcionalidade de detecção de fraudes.

## Arquitetura

### Componentes Principais

![Arquitetura de Detecção de Fraudes](../../../assets/images/fraud_detection_architecture.png)

1. **Orquestrador de Agentes**: Coordena a execução de múltiplos agentes de detecção, gerencia timeout e consolida os resultados.

2. **Agentes de Detecção**:
   - **Agente Baseado em Regras**: Utiliza regras configuráveis para identificar atividades suspeitas
   - **Agente de ML**: Aplica modelos de machine learning para detectar anomalias e padrões de fraude
   - **Agente Comportamental**: Analisa padrões de comportamento do usuário em relação a linhas de base estabelecidas

3. **Serviço de Detecção de Fraudes**: Integra os agentes IA com a API do Bureau de Crédito

4. **Contexto de Análise**: Compartilha informações entre agentes e mantém o estado durante a análise

## Integração via API

### Endpoint de Análise de Solicitação de Crédito

**Endpoint**: `/api/v1/bureau-credito/fraud-detection/credit-request`

**Método**: POST

**Headers**:
- `Content-Type: application/json`
- `Authorization: Bearer {token}`
- `X-Tenant-ID: {tenant_id}`

**Request Body**:
```json
{
  "request_id": "unique-request-id",
  "customer_id": "customer-123",
  "tenant_id": "tenant-456",
  "amount": 5000.00,
  "currency": "AOA",
  "term_months": 12,
  "purpose": "home_improvement",
  "customer_data": {
    "name": "Nome Completo",
    "document": "123456789",
    "document_type": "BI",
    "birth_date": "1985-05-15",
    "address": {
      "street": "Rua Principal",
      "city": "Luanda",
      "province": "Luanda",
      "postal_code": "0000",
      "country": "Angola"
    },
    "contact": {
      "email": "email@example.com",
      "phone": "+244123456789"
    }
  },
  "device_info": {
    "ip_address": "192.168.1.1",
    "user_agent": "Mozilla/5.0...",
    "fingerprint": "device-fingerprint-123",
    "geolocation": {
      "latitude": -8.8383,
      "longitude": 13.2344
    }
  },
  "application_info": {
    "timestamp": "2023-06-15T14:30:45Z",
    "channel": "web",
    "previous_applications": 2
  }
}
```

**Response**:
```json
{
  "status": "success",
  "transaction_id": "transaction-789",
  "timestamp": "2023-06-15T14:30:48Z",
  "risk_score": 0.35,
  "decision": "approve", // approve, reject, review
  "confidence": 0.85,
  "fraud_indicators": [
    {
      "indicator_type": "rule_violation",
      "severity": "medium",
      "description": "Múltiplas solicitações em período curto",
      "confidence": 0.65
    }
  ],
  "insights": {
    "rules_agent": {
      "rules_summary": {
        "total_rules": 15,
        "evaluated_rules": 12,
        "violated_rules": 1,
        "risk_score": 0.35
      }
    },
    "ml_agent": {
      "anomaly_score": 0.28
    }
  }
}
```

### Endpoint de Análise de Transação

**Endpoint**: `/api/v1/bureau-credito/fraud-detection/transaction`

**Método**: POST

**Headers**:
- `Content-Type: application/json`
- `Authorization: Bearer {token}`
- `X-Tenant-ID: {tenant_id}`

**Request Body**:
```json
{
  "transaction_id": "transaction-123",
  "customer_id": "customer-123",
  "tenant_id": "tenant-456",
  "amount": 1500.00,
  "currency": "AOA",
  "transaction_type": "payment",
  "payment_method": {
    "type": "credit_card",
    "last_digits": "1234",
    "expiry_date": "12/25"
  },
  "merchant_info": {
    "merchant_id": "merchant-789",
    "name": "Loja Online",
    "category": "retail",
    "country": "Angola"
  },
  "device_info": {
    "ip_address": "192.168.1.1",
    "user_agent": "Mozilla/5.0...",
    "fingerprint": "device-fingerprint-123",
    "geolocation": {
      "latitude": -8.8383,
      "longitude": 13.2344
    }
  },
  "transaction_info": {
    "timestamp": "2023-06-15T14:30:45Z",
    "channel": "web"
  }
}
```

**Response**: Mesmo formato da resposta de análise de solicitação de crédito

## Configuração dos Agentes

### Configuração do Agente Baseado em Regras

As regras são definidas em formato JSON e podem ser configuradas via API ou arquivo:

```json
{
  "id": "rule_id",
  "name": "Nome da Regra",
  "description": "Descrição da regra",
  "severity": "high", // high, medium, low
  "risk_score": 0.8,
  "condition_type": "simple", // simple, complex, regex, threshold
  "condition": {
    "field": "device_info.ip_address",
    "operator": "in", // eq, neq, gt, lt, gte, lte, in, contains, exists
    "value": ["192.168.1.1", "10.0.0.1"]
  },
  "enabled": true,
  "tags": ["device", "security"]
}
```

Para condições complexas:

```json
{
  "condition_type": "complex",
  "condition": {
    "logical_operator": "AND", // AND, OR
    "conditions": [
      {
        "type": "simple",
        "field": "amount",
        "operator": "gt",
        "value": 10000
      },
      {
        "type": "simple",
        "field": "customer_data.age",
        "operator": "lt",
        "value": 25
      }
    ]
  }
}
```

### Configuração do Agente ML

O agente ML pode ser configurado com diferentes tipos de modelos:

```json
{
  "agent_id": "ml_agent_1",
  "model_type": "isolation_forest", // isolation_forest, random_forest
  "model_path": "/path/to/model.pkl",
  "feature_mapping": {
    "amount": "amount",
    "age": "customer_data.age",
    "transaction_count": "application_info.previous_applications"
  },
  "required_features": ["amount", "age", "transaction_count"]
}
```

## Treinamento de Modelos

O sistema suporta treinamento contínuo dos modelos ML via API:

**Endpoint**: `/api/v1/bureau-credito/fraud-detection/train`

**Método**: POST

**Request Body**:
```json
{
  "agent_id": "ml_agent_1",
  "training_data": [
    {
      "amount": 5000.00,
      "customer_data": {
        "age": 35
      },
      "application_info": {
        "previous_applications": 2
      },
      "is_fraud": false
    },
    {
      "amount": 25000.00,
      "customer_data": {
        "age": 22
      },
      "application_info": {
        "previous_applications": 5
      },
      "is_fraud": true
    }
  ]
}
```

## Observabilidade e Monitoramento

### Métricas Prometheus

O sistema expõe as seguintes métricas via Prometheus:

1. **fraud_detection_requests_total{tenant_id, result}**
   - Contador total de solicitações de detecção de fraude por tenant e resultado

2. **fraud_detection_processing_time{tenant_id, agent_id}**
   - Histograma do tempo de processamento por agente e tenant

3. **fraud_detection_risk_score{tenant_id}**
   - Histograma dos scores de risco por tenant

4. **fraud_indicators_total{tenant_id, indicator_type, severity}**
   - Contador de indicadores de fraude por tipo, severidade e tenant

### Painéis Grafana

O módulo inclui dashboards Grafana pré-configurados para monitoramento:

1. **Dashboard de Visão Geral**
   - Taxas de solicitação por tenant e decisão
   - Distribuição de score de risco
   - Contagem de indicadores de fraude

2. **Dashboard de Desempenho de Agentes**
   - Tempo de execução por agente
   - Taxa de detecção por agente
   - Desempenho de aprendizado dos modelos ML

## Conformidade Regulatória

O sistema foi projetado para cumprir com os seguintes regulamentos:

1. **RGPD/GDPR**: Controles para processamento seguro de dados pessoais
2. **POPIA (África do Sul)**: Medidas para proteção de dados pessoais
3. **Lei de Proteção de Dados (Angola)**: Conformidade com a regulamentação local
4. **LGPD (Brasil)**: Princípios de minimização e propósito específico

### Registro de Auditorias

Todas as decisões de detecção de fraude são registradas em logs de auditoria com os seguintes detalhes:

- ID de transação
- Timestamp
- Tenant ID
- Decisão tomada
- Agentes utilizados
- Score de risco

## Limitações e Considerações

1. **Tempo de Resposta**: O timeout padrão é de 10 segundos para a análise completa
2. **Requisitos de Hardware**: Recomenda-se no mínimo 4 CPUs e 8GB RAM para o serviço
3. **Limites de Taxa**: Máximo de 100 solicitações por segundo por tenant
4. **Armazenamento de Dados**: Modelos ML podem ocupar até 500MB de espaço em disco

## Exemplos de Integração

### Integração com Sistemas Legados

Para sistemas legados sem suporte a REST:

```java
// Exemplo de cliente Java
import com.innovabiz.bureau.client.FraudDetectionClient;

FraudDetectionClient client = new FraudDetectionClient("https://api.example.com", "api-key");
CreditRequest request = new CreditRequest.Builder()
    .withCustomerId("customer-123")
    .withAmount(5000.00)
    .withCurrency("AOA")
    .withTerm(12)
    .build();
    
FraudAnalysisResult result = client.analyzeCreditRequest(request);
if (result.getDecision() == Decision.APPROVE) {
    // Proceder com aprovação
} else if (result.getDecision() == Decision.REVIEW) {
    // Encaminhar para revisão manual
} else {
    // Rejeitar solicitação
}
```

### Integração com Microsserviços

Para arquiteturas de microsserviços:

```typescript
// Exemplo TypeScript
import { FraudDetectionService } from '@innovabiz/fraud-detection';

const service = new FraudDetectionService({
  baseUrl: 'https://api.example.com',
  apiKey: 'your-api-key',
  tenantId: 'tenant-456'
});

async function processTransaction(transaction) {
  try {
    const result = await service.analyzeTransaction({
      transaction_id: transaction.id,
      customer_id: transaction.customerId,
      amount: transaction.amount,
      // ... outros campos
    });
    
    if (result.risk_score > 0.8) {
      return { status: 'rejected', reason: 'high_risk' };
    } else if (result.risk_score > 0.5) {
      return { status: 'manual_review', reason: 'medium_risk' };
    } else {
      return { status: 'approved' };
    }
  } catch (error) {
    console.error('Erro na detecção de fraude:', error);
    return { status: 'error', message: error.message };
  }
}
```

## Suporte e Contato

Para suporte técnico ou dúvidas sobre a integração:

- **Email**: support@innovabiz.com
- **Portal do Desenvolvedor**: https://developers.innovabiz.com
- **Documentação API**: https://api.innovabiz.com/docs/bureau-credito

## Changelog e Roadmap

### Versão Atual: 0.1.0

- Implementação inicial dos agentes IA
- Suporte a regras configuráveis
- Modelos ML básicos (Isolation Forest, Random Forest)
- API REST para integração

### Próximas Versões

- **0.2.0**: Suporte a redes neurais e aprendizado profundo
- **0.3.0**: Detecção de fraudes em tempo real com processamento de streaming
- **1.0.0**: Versão estável com suporte completo para todos os mercados-alvo