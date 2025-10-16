# Manual de Integração para Parceiros - Bureau de Créditos

**Documento**: Manual de Integração para Parceiros - Bureau de Créditos  
**Versão**: 1.0.0  
**Data**: 07/08/2025  
**Classificação**: Restrito

## 1. Visão Geral

Este documento descreve o processo de integração com o módulo IAM da INNOVABIZ para acesso aos serviços de Bureau de Créditos nos mercados PALOP e SADC. O manual orienta parceiros estratégicos na implementação segura e conforme das APIs necessárias para consulta e gestão de informações de crédito.

### 1.1. Objetivos

- Fornecer instruções técnicas detalhadas para integração com o IAM
- Documentar fluxos de autenticação e autorização para acesso ao Bureau de Créditos
- Especificar requisitos de segurança e conformidade regulatória
- Orientar sobre validação e testes da integração

### 1.2. Público-Alvo

Este manual destina-se a:

- Equipes de desenvolvimento de parceiros estratégicos
- Arquitetos de integração de sistemas
- Equipes de compliance e segurança da informação
- Gestores técnicos responsáveis pela implementação

## 2. Requisitos Técnicos

### 2.1. Pré-requisitos

Para integração com a plataforma IAM para acesso ao Bureau de Créditos, são necessários:

1. **Credenciais de API**
   - Client ID e Client Secret fornecidos pela INNOVABIZ
   - Certificados digitais para canais seguros (quando aplicável)

2. **Infraestrutura**
   - TLS 1.3 ou superior
   - Suporte para OAuth 2.1 e OpenID Connect 1.0
   - Capacidade de processamento de tokens JWT
   - Armazenamento seguro para credenciais

3. **Conformidade**
   - Aderência aos requisitos específicos da jurisdição (BNA, Banco de Moçambique, NCR, etc.)
   - Políticas de privacidade compatíveis com GDPR, POPIA ou regulamentações locais
   - Processos documentados de gestão de consentimento

### 2.2. Ambientes Disponíveis

| Ambiente | URL Base | Finalidade |
|----------|----------|------------|
| Sandbox | `https://sandbox.api.iam.innovabiz.com/bureau` | Desenvolvimento e testes iniciais |
| Homologação | `https://homolog.api.iam.innovabiz.com/bureau` | Certificação e validação |
| Produção | `https://api.iam.innovabiz.com/bureau` | Ambiente produtivo |

## 3. Arquitetura de Integração

### 3.1. Visão Geral da Arquitetura

A integração com o Bureau de Créditos através do IAM segue um modelo de API segura em múltiplas camadas:

```
Parceiro → API Gateway (Krakend) → IAM → Adaptadores Bureau → Bureau de Créditos
```

### 3.2. Fluxos de Autenticação

#### 3.2.1. OAuth 2.1 com PKCE (Recomendado)

```mermaid
sequenceDiagram
    participant App as Aplicação Parceiro
    participant Auth as IAM Auth Server
    participant API as API Bureau
    participant Bureau as Bureau de Créditos

    App->>App: Gera code_verifier e code_challenge
    App->>Auth: Authorization Request + code_challenge
    Auth->>App: Authorization Code
    App->>Auth: Token Request + code_verifier
    Auth->>App: Access Token + Refresh Token
    App->>API: API Request + Access Token
    API->>Bureau: Consulta Bureaus
    Bureau->>API: Resultado Consulta
    API->>App: API Response
```#### 3.2.2. Autenticação com Cliente Confidencial

```mermaid
sequenceDiagram
    participant App as Sistema Parceiro
    participant Auth as IAM Auth Server
    participant API as API Bureau
    participant Bureau as Bureau de Créditos

    App->>Auth: Token Request (client_id + client_secret)
    Auth->>App: Access Token
    App->>API: API Request + Access Token
    API->>Bureau: Consulta Bureaus
    Bureau->>API: Resultado Consulta
    API->>App: API Response
```

### 3.3. Contexto Multi-Regulatório

A plataforma aplica automaticamente contextos regulatórios específicos baseados na jurisdição do tenant:

```json
{
  "angola": {
    "regulatoryFramework": "BNA",
    "consentRequirements": ["explicit", "written", "purpose_limited"],
    "dataRetentionPeriod": "5 years",
    "applicableLaws": ["Lei 22/11", "Aviso 12/2016"]
  },
  "mozambique": {
    "regulatoryFramework": "Banco de Moçambique",
    "consentRequirements": ["explicit", "multi_language", "purpose_limited"],
    "dataRetentionPeriod": "7 years",
    "applicableLaws": ["Lei 15/2020", "Aviso 10/GBM/2016"]
  },
  "south_africa": {
    "regulatoryFramework": "NCR",
    "consentRequirements": ["explicit", "revocable", "purpose_limited"],
    "dataRetentionPeriod": "5 years",
    "applicableLaws": ["National Credit Act", "POPIA"]
  }
}
```

## 4. APIs Disponíveis

### 4.1. Visão Geral das APIs

A plataforma oferece duas interfaces de API:

1. **REST API** - Para operações simples e diretas
2. **GraphQL API** - Para consultas complexas e relacionadas

Todas as APIs requerem autenticação via OAuth 2.1 e seguem os princípios RESTful:

- Versionamento via URL ou cabeçalho
- Respostas paginadas para conjuntos grandes
- Suporte a HATEOAS para navegação de recursos
- Padronização de códigos de erro

### 4.2. Endpoints REST Principais

#### 4.2.1. Gerenciamento de Vínculos de Identidade

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| GET | `/v1/bureau/identity-links` | Lista vínculos de identidade |
| GET | `/v1/bureau/identity-links/{id}` | Obtém detalhes de um vínculo |
| POST | `/v1/bureau/identity-links` | Cria novo vínculo de identidade |
| PUT | `/v1/bureau/identity-links/{id}` | Atualiza vínculo existente |
| DELETE | `/v1/bureau/identity-links/{id}` | Revoga vínculo de identidade |

#### 4.2.2. Autorizações de Acesso

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| GET | `/v1/bureau/authorizations` | Lista autorizações |
| GET | `/v1/bureau/authorizations/{id}` | Obtém detalhes de autorização |
| POST | `/v1/bureau/authorizations` | Cria nova autorização |
| PUT | `/v1/bureau/authorizations/{id}` | Atualiza autorização existente |
| DELETE | `/v1/bureau/authorizations/{id}` | Revoga autorização |

#### 4.2.3. Consultas ao Bureau

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| GET | `/v1/bureau/credit-score/{userId}` | Obtém score de crédito |
| GET | `/v1/bureau/credit-report/{userId}` | Obtém relatório completo |
| GET | `/v1/bureau/credit-history/{userId}` | Obtém histórico de crédito |
| GET | `/v1/bureau/loan-offers/{userId}` | Obtém ofertas de crédito disponíveis |

#### Exemplo de Requisição REST

```http
GET /v1/bureau/credit-score/550e8400-e29b-41d4-a716-446655440000 HTTP/1.1
Host: api.iam.innovabiz.com
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
X-Tenant-ID: 7890abcd-ef12-34gh-i567-jklmnopqrstu
Content-Type: application/json
```#### Exemplo de Resposta REST

```json
{
  "data": {
    "userId": "550e8400-e29b-41d4-a716-446655440000",
    "scoreDate": "2025-08-07T10:15:30Z",
    "bureauProvider": "CIRC-BNA",
    "scoreValue": 725,
    "scoreRange": {
      "min": 0,
      "max": 1000
    },
    "riskCategory": "low",
    "factors": [
      {
        "factor": "payment_history",
        "impact": "positive",
        "description": "Histórico de pagamentos em dia"
      },
      {
        "factor": "credit_utilization",
        "impact": "neutral",
        "description": "Utilização de 65% do crédito disponível"
      }
    ],
    "metadata": {
      "regulatoryContext": "angola",
      "consentId": "a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6",
      "queryId": "q1w2e3r4-t5y6-u7i8-o9p0-a1s2d3f4g5h6"
    }
  }
}
```

### 4.3. API GraphQL

A API GraphQL permite consultas personalizadas e mais granulares aos dados do Bureau de Créditos.

#### 4.3.1. Endpoint GraphQL

```
POST https://api.iam.innovabiz.com/bureau/graphql
```

#### 4.3.2. Esquema GraphQL Simplificado

```graphql
type Query {
  # Consulta de vínculos de identidade
  bureauIdentityLinks(
    filter: BureauLinkFilterInput
    pagination: PaginationInput
  ): BureauLinkConnection!
  
  # Consulta de autorizações
  bureauAuthorizations(
    filter: BureauAuthFilterInput
    pagination: PaginationInput
  ): BureauAuthConnection!
  
  # Consulta de score de crédito
  bureauCreditScore(
    userId: ID!
    authorizationId: ID!
  ): CreditScore!
  
  # Consulta de relatório completo
  bureauCreditReport(
    userId: ID!
    authorizationId: ID!
  ): CreditReport!
  
  # Consulta histórico de crédito
  bureauCreditHistory(
    userId: ID!
    authorizationId: ID!
    timeframe: TimeframeInput
  ): CreditHistoryConnection!
}

# Objetos principais
type BureauLink {
  id: ID!
  userId: ID!
  bureauProvider: String!
  bureauId: String!
  status: LinkStatus!
  verificationStatus: VerificationStatus
  createdAt: DateTime!
  updatedAt: DateTime
  metadata: JSONObject
}

type Authorization {
  id: ID!
  linkId: ID!
  purpose: String!
  scope: [String!]!
  status: AuthStatus!
  createdAt: DateTime!
  expiresAt: DateTime!
  createdBy: ID
  consentId: ID
  regulatoryContext: JSONObject
}

type CreditScore {
  userId: ID!
  scoreDate: DateTime!
  bureauProvider: String!
  scoreValue: Int!
  scoreRange: ScoreRange!
  riskCategory: RiskCategory!
  factors: [ScoreFactor!]!
  metadata: JSONObject
}
```

#### 4.3.3. Exemplo de Consulta GraphQL

```graphql
query GetUserCreditInfo {
  bureauCreditScore(
    userId: "550e8400-e29b-41d4-a716-446655440000",
    authorizationId: "a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6"
  ) {
    scoreValue
    riskCategory
    factors {
      factor
      impact
      description
    }
  }
  
  bureauIdentityLinks(
    filter: { 
      userId: "550e8400-e29b-41d4-a716-446655440000",
      status: ACTIVE
    }
  ) {
    edges {
      node {
        bureauProvider
        status
        createdAt
      }
    }
  }
}
```## 5. Requisitos de Segurança

### 5.1. Autenticação e Autorização

#### 5.1.1. Gestão de Tokens

- **Armazenamento**: Tokens devem ser armazenados de forma segura, preferencialmente em armazenamento criptografado
- **Ciclo de Vida**: Implementar renovação automática de tokens expirados usando refresh tokens
- **Revogação**: Implementar mecanismo para revogação imediata de tokens comprometidos

#### 5.1.2. Escopos de Acesso

| Escopo | Descrição | Permissões |
|--------|-----------|------------|
| `bureau:identity:read` | Leitura de vínculos de identidade | GET em `/identity-links` |
| `bureau:identity:write` | Gestão de vínculos de identidade | POST, PUT, DELETE em `/identity-links` |
| `bureau:auth:read` | Leitura de autorizações | GET em `/authorizations` |
| `bureau:auth:write` | Gestão de autorizações | POST, PUT, DELETE em `/authorizations` |
| `bureau:score:read` | Consulta de score de crédito | GET em `/credit-score` |
| `bureau:report:read` | Consulta de relatórios completos | GET em `/credit-report` |
| `bureau:history:read` | Consulta de histórico de crédito | GET em `/credit-history` |
| `bureau:offers:read` | Consulta de ofertas disponíveis | GET em `/loan-offers` |

### 5.2. Criptografia e Proteção de Dados

#### 5.2.1. Requisitos de Criptografia

- TLS 1.3 ou superior para todas as comunicações
- Criptografia AES-256 para dados armazenados
- Assinatura digital de mensagens críticas
- HMAC para verificação de integridade

#### 5.2.2. Proteção de Dados Sensíveis

- Dados de identificação pessoal devem ser criptografados em trânsito e em repouso
- Implementar mascaramento de dados sensíveis em logs e interfaces
- Dados de cartão de crédito e biométricos nunca devem ser armazenados localmente

#### 5.2.3. Segurança de Endpoints

- Implementar proteção contra ataques de injeção
- Validar todos os inputs do cliente
- Implementar rate limiting por tenant e por IP
- Monitorar e alertar sobre padrões de acesso anômalos

## 6. Requisitos de Conformidade

### 6.1. Gestão de Consentimento

A integração deve implementar um processo robusto de gestão de consentimento que atenda aos requisitos regulatórios do mercado específico:

#### 6.1.1. Requisitos Básicos de Consentimento

- Consentimento deve ser explícito e específico
- Finalidade da consulta claramente informada ao usuário
- Data de expiração do consentimento claramente indicada
- Mecanismo para revogação de consentimento a qualquer momento
- Registro de auditoria de todos os consentimentos

#### 6.1.2. Modelo de Implementação

```typescript
interface ConsentRecord {
  id: string;
  userId: string;
  purpose: string;
  scopes: string[];
  createdAt: Date;
  expiresAt: Date;
  consentText: string;
  consentVersion: string;
  userIpAddress: string;
  userAgent: string;
  regulatoryContext: string;
}
```

#### 6.1.3. Exemplo de Fluxo de Consentimento

```mermaid
sequenceDiagram
    participant User as Usuário
    participant App as Aplicação Parceiro
    participant IAM as IAM INNOVABIZ
    participant Bureau as Bureau de Créditos
    
    App->>User: Apresenta tela de consentimento
    User->>App: Concede consentimento
    App->>IAM: Registra consentimento (POST /consent)
    IAM->>App: ID do consentimento
    App->>IAM: Solicita autorização (com consentimento_id)
    IAM->>App: ID de autorização
    App->>IAM: Consulta Bureau (com authorization_id)
    IAM->>Bureau: Consulta com contexto regulatório
    Bureau->>IAM: Dados de crédito
    IAM->>App: Resposta da consulta
```### 6.2. Requisitos Específicos por Jurisdição

#### 6.2.1. Angola (BNA)

- **Base Legal**: Lei 22/11 e Aviso 12/2016 do BNA
- **Requisitos Específicos**:
  - Consentimento explícito por escrito do titular dos dados
  - Retenção de dados por 5 anos
  - Relatórios mensais para o BNA sobre consultas realizadas
  - Dados devem ser armazenados em território angolano
  - Notificação ao usuário antes de cada consulta

#### 6.2.2. Moçambique (Banco de Moçambique)

- **Base Legal**: Lei 15/2020 e Aviso 10/GBM/2016
- **Requisitos Específicos**:
  - Consentimento em múltiplos idiomas (português e línguas locais)
  - Retenção de dados por 7 anos
  - Relatórios trimestrais para o Banco de Moçambique
  - Interface de consulta deve estar disponível em português

#### 6.2.3. África do Sul (NCR)

- **Base Legal**: National Credit Act e POPIA
- **Requisitos Específicos**:
  - Conformidade com POPIA para processamento de dados pessoais
  - Avaliação de impacto de proteção de dados necessária
  - Mecanismo de correção de informações incorretas
  - Notificação de incidentes de segurança em 72 horas

## 7. Implementação e Guias Práticos

### 7.1. Processo de Onboarding

1. **Registro na Plataforma**
   - Solicitar acesso via [parceiros.innovabiz.com](https://parceiros.innovabiz.com)
   - Preencher formulário de avaliação de conformidade
   - Assinar contrato de parceria e termos de uso da API

2. **Configuração do Ambiente**
   - Receber credenciais para ambiente Sandbox
   - Configurar certificados e chaves de API
   - Implementar fluxos de autenticação

3. **Validação e Certificação**
   - Executar suite de testes fornecida
   - Validar conformidade com requisitos de segurança
   - Obter aprovação para acesso ao ambiente de Homologação

4. **Migração para Produção**
   - Validação final em ambiente de Homologação
   - Obter aprovação para acesso de Produção
   - Configurar monitoramento e suporte

### 7.2. Exemplos de Implementação

#### 7.2.1. Exemplo em Node.js

```javascript
// Exemplo de autenticação e consulta de score
const axios = require('axios');
const crypto = require('crypto');

// Configuração
const clientId = 'YOUR_CLIENT_ID';
const clientSecret = 'YOUR_CLIENT_SECRET';
const apiBaseUrl = 'https://api.iam.innovabiz.com/bureau';
const tenantId = 'YOUR_TENANT_ID';

// Função para obter token de acesso
async function getAccessToken() {
  try {
    const response = await axios.post('https://auth.innovabiz.com/oauth/token', {
      grant_type: 'client_credentials',
      client_id: clientId,
      client_secret: clientSecret,
      audience: 'bureau-api'
    });
    
    return response.data.access_token;
  } catch (error) {
    console.error('Error obtaining access token:', error.response?.data || error.message);
    throw error;
  }
}

// Função para consultar score de crédito
async function getCreditScore(userId, authorizationId) {
  try {
    const token = await getAccessToken();
    
    const response = await axios.get(`${apiBaseUrl}/v1/bureau/credit-score/${userId}`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'X-Tenant-ID': tenantId,
        'X-Authorization-ID': authorizationId
      }
    });
    
    return response.data;
  } catch (error) {
    console.error('Error fetching credit score:', error.response?.data || error.message);
    throw error;
  }
}

// Uso
async function main() {
  try {
    // Exemplo de consulta
    const userId = '550e8400-e29b-41d4-a716-446655440000';
    const authorizationId = 'a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6';
    
    const creditScore = await getCreditScore(userId, authorizationId);
    console.log('Credit Score:', creditScore);
  } catch (error) {
    console.error('Main error:', error);
  }
}

main();
```#### 7.2.2. Exemplo em Python

```python
import requests
import time
import jwt
import json
from datetime import datetime, timedelta

# Configuração
client_id = 'YOUR_CLIENT_ID'
client_secret = 'YOUR_CLIENT_SECRET'
api_base_url = 'https://api.iam.innovabiz.com/bureau'
tenant_id = 'YOUR_TENANT_ID'

# Função para obter token de acesso
def get_access_token():
    try:
        response = requests.post('https://auth.innovabiz.com/oauth/token', json={
            'grant_type': 'client_credentials',
            'client_id': client_id,
            'client_secret': client_secret,
            'audience': 'bureau-api'
        })
        
        response.raise_for_status()
        return response.json()['access_token']
    except requests.exceptions.RequestException as e:
        print(f"Erro ao obter token de acesso: {e}")
        raise

# Função para registrar consentimento
def register_consent(user_id, purpose, scope):
    try:
        token = get_access_token()
        
        # Data de expiração: 90 dias
        expires_at = (datetime.now() + timedelta(days=90)).isoformat()
        
        response = requests.post(
            f"{api_base_url}/v1/bureau/consent", 
            headers={
                'Authorization': f'Bearer {token}',
                'X-Tenant-ID': tenant_id
            },
            json={
                'userId': user_id,
                'purpose': purpose,
                'scope': scope,
                'expiresAt': expires_at,
                'consentText': 'Autorizo a consulta ao Bureau de Créditos para avaliação de crédito',
                'consentVersion': '1.0'
            }
        )
        
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as e:
        print(f"Erro ao registrar consentimento: {e}")
        raise

# Função para criar autorização
def create_authorization(user_id, consent_id, purpose, scope):
    try:
        token = get_access_token()
        
        # Data de expiração: 30 dias
        expires_at = (datetime.now() + timedelta(days=30)).isoformat()
        
        response = requests.post(
            f"{api_base_url}/v1/bureau/authorizations", 
            headers={
                'Authorization': f'Bearer {token}',
                'X-Tenant-ID': tenant_id
            },
            json={
                'userId': user_id,
                'consentId': consent_id,
                'purpose': purpose,
                'scope': scope,
                'expiresAt': expires_at
            }
        )
        
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as e:
        print(f"Erro ao criar autorização: {e}")
        raise

# Função para consultar score de crédito
def get_credit_score(user_id, authorization_id):
    try:
        token = get_access_token()
        
        response = requests.get(
            f"{api_base_url}/v1/bureau/credit-score/{user_id}", 
            headers={
                'Authorization': f'Bearer {token}',
                'X-Tenant-ID': tenant_id,
                'X-Authorization-ID': authorization_id
            }
        )
        
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as e:
        print(f"Erro ao consultar score de crédito: {e}")
        raise

# Exemplo de uso
if __name__ == "__main__":
    try:
        user_id = "550e8400-e29b-41d4-a716-446655440000"
        
        # 1. Registrar consentimento
        consent = register_consent(
            user_id, 
            "avaliação de crédito", 
            ["bureau:score:read", "bureau:history:read"]
        )
        print(f"Consentimento registrado: {consent['id']}")
        
        # 2. Criar autorização
        authorization = create_authorization(
            user_id,
            consent['id'],
            "avaliação de crédito",
            ["bureau:score:read"]
        )
        print(f"Autorização criada: {authorization['id']}")
        
        # 3. Consultar score
        credit_score = get_credit_score(user_id, authorization['id'])
        print(f"Score de crédito: {credit_score['data']['scoreValue']}")
        print(f"Categoria de risco: {credit_score['data']['riskCategory']}")
        
    except Exception as e:
        print(f"Erro na execução do exemplo: {e}")
```### 7.3. Boas Práticas

#### 7.3.1. Gestão de Erros e Resiliência

- Implementar retry com backoff exponencial para falhas temporárias
- Cachear tokens de acesso localmente para reduzir chamadas de autenticação
- Monitorar expiração de tokens e renovar proativamente
- Implementar circuit breaker para prevenir sobrecarga em falhas

#### 7.3.2. Monitoramento e Logging

- Registrar todas as chamadas de API para auditoria
- Implementar monitoramento de tempo de resposta
- Configurar alertas para erros frequentes ou degradação de performance
- Mascarar dados sensíveis em logs (PII, tokens, credenciais)

#### 7.3.3. Performance

- Minimizar o número de requisições através de consultas otimizadas
- Implementar cache local para dados não sensíveis ou frequentemente acessados
- Utilizar compressão para reduzir tamanho das requisições e respostas
- Manter conexões HTTP persistentes para reduzir overhead de handshake

## 8. Códigos de Erro e Resolução de Problemas

### 8.1. Códigos de Erro HTTP

| Código | Descrição | Resolução |
|--------|-----------|-----------|
| 400 | Bad Request | Verifique os parâmetros enviados na requisição |
| 401 | Unauthorized | Token expirado ou inválido; obtenha um novo token |
| 403 | Forbidden | O token não tem as permissões necessárias para a operação |
| 404 | Not Found | O recurso solicitado não existe |
| 409 | Conflict | Operação não pode ser concluída devido a conflito de dados |
| 422 | Unprocessable Entity | Dados válidos, mas semanticamente incorretos |
| 429 | Too Many Requests | Limite de taxa excedido; aguarde antes de tentar novamente |
| 500 | Internal Server Error | Erro no servidor; contate suporte se persistir |
| 503 | Service Unavailable | Serviço temporariamente indisponível; tente mais tarde |

### 8.2. Códigos de Erro Específicos

| Código | Mensagem | Descrição | Resolução |
|--------|----------|-----------|-----------|
| BC-001 | `invalid_tenant` | Tenant ID inválido ou inexistente | Verifique o Tenant ID fornecido |
| BC-002 | `missing_consent` | Consentimento não registrado ou expirado | Registre um novo consentimento do usuário |
| BC-003 | `invalid_authorization` | Autorização inválida ou expirada | Crie uma nova autorização |
| BC-004 | `bureau_unavailable` | Bureau de Crédito temporariamente indisponível | Tente novamente mais tarde |
| BC-005 | `user_not_found` | Usuário não encontrado no Bureau | Verifique dados de identificação |
| BC-006 | `insufficient_scope` | Token não possui os escopos necessários | Solicite token com escopos apropriados |
| BC-007 | `regulatory_violation` | Violação de requisito regulatório | Verifique conformidade com requisitos legais |
| BC-008 | `rate_limit_exceeded` | Limite de consultas por tenant excedido | Aguarde o período de reset ou aumente seu plano |
| BC-009 | `data_format_error` | Dados fornecidos em formato inválido | Corrija o formato dos dados conforme documentação |

### 8.3. Resolução de Problemas Comuns

#### 8.3.1. Autenticação

**Problema**: Falha ao obter token de acesso  
**Possíveis Causas**:
- Credenciais incorretas (Client ID ou Client Secret)
- IP não autorizado
- Cliente desativado ou expirado

**Resolução**:
1. Verifique as credenciais fornecidas
2. Confirme se o IP de origem está na lista de permitidos
3. Verifique status do cliente no portal de parceiros

#### 8.3.2. Consultas ao Bureau

**Problema**: Falha na consulta ao Bureau  
**Possíveis Causas**:
- Consentimento ou autorização expirados
- Dados de identificação insuficientes
- Bureau temporariamente indisponível

**Resolução**:
1. Verifique status da autorização e consentimento
2. Confirme que todos os dados necessários foram fornecidos
3. Implemente retry com backoff exponencial para falhas temporárias

## 9. Suporte e Contato

### 9.1. Canais de Suporte

- **Portal de Desenvolvedores**: [developers.innovabiz.com](https://developers.innovabiz.com)
- **Email de Suporte**: bureau-support@innovabiz.com
- **Telefone**: +244 123 456 789 (Angola) / +55 11 4567-8901 (Brasil)

### 9.2. Níveis de Serviço

| Ambiente | Horário de Suporte | Tempo de Resposta | Disponibilidade |
|----------|-------------------|-------------------|----------------|
| Sandbox | Segunda a Sexta, 9h-18h | Até 24 horas | 99.5% |
| Homologação | Segunda a Sexta, 9h-18h | Até 8 horas | 99.9% |
| Produção | 24x7 | Crítico: 1 hora<br>Alta: 4 horas<br>Média: 8 horas | 99.99% |

### 9.3. Processo de Escalonamento

1. **Nível 1**: Suporte inicial via portal ou email
2. **Nível 2**: Suporte técnico especializado
3. **Nível 3**: Engenharia de produto e especialistas em Bureau
4. **Nível 4**: Gestão executiva e relações com Bureau de Créditos

## 10. Glossário

| Termo | Definição |
|-------|-----------|
| **API** | Interface de Programação de Aplicações |
| **Bureau de Créditos** | Instituição que coleta e armazena histórico de crédito de consumidores |
| **IAM** | Identity and Access Management - Sistema de gestão de identidade e acesso |
| **JWT** | JSON Web Token - Token seguro para autenticação e transmissão de informações |
| **OAuth 2.1** | Protocolo padrão para autorização de acesso |
| **OIDC** | OpenID Connect - Camada de identidade sobre OAuth 2.0 |
| **PALOP** | Países Africanos de Língua Oficial Portuguesa |
| **PKCE** | Proof Key for Code Exchange - Extensão de segurança para OAuth |
| **SADC** | Southern African Development Community - Comunidade para o Desenvolvimento da África Austral |
| **Tenant** | Instância isolada de um cliente dentro da plataforma multi-tenant |
| **TLS** | Transport Layer Security - Protocolo de criptografia para comunicações seguras |

## Apêndice A: Requisitos Técnicos por Mercado

| País | Legislação Aplicável | Requisitos Específicos | Formatos de Documentos |
|------|----------------------|------------------------|------------------------|
| Angola | Lei 22/11, Aviso 12/2016 | Dados residentes em Angola, retenção por 5 anos | BI, NIF |
| Moçambique | Lei 15/2020, Aviso 10/GBM/2016 | Relatórios trimestrais, retenção por 7 anos | NUIT, BI |
| África do Sul | National Credit Act, POPIA | Avaliação de impacto, notificação de incidentes | ID Number, Passport |
| Namíbia | Banking Institutions Act | Consentimento explícito por escrito | ID Number, Passport |
| Botswana | Data Protection Act | Direito à correção de informações | Omang ID, Passport |

---

© 2025 INNOVABIZ - Todos os direitos reservados. Classificação: Restrito

**Aviso Legal**: Este documento é confidencial e seu conteúdo é destinado apenas aos parceiros autorizados da INNOVABIZ. Qualquer distribuição, cópia ou uso não autorizado deste material é estritamente proibido.