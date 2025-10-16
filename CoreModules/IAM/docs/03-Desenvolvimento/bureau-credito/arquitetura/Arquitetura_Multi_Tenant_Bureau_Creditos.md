# Arquitetura Multi-Tenant e Multi-Contexto para Integração com Bureau de Créditos

**Autor:** Eduardo Jeremias - InnovaBiz  
**Data:** 19/08/2025  
**Versão:** 1.0  
**Classificação:** Confidencial  

## Índice

1. [Introdução](#1-introdução)
2. [Princípios Arquiteturais](#2-princípios-arquiteturais)
3. [Modelo de Multi-Tenancy para Bureau de Créditos](#3-modelo-de-multi-tenancy-para-bureau-de-créditos)
4. [Arquitetura Multi-Contexto](#4-arquitetura-multi-contexto)
5. [Estrutura de Dados](#5-estrutura-de-dados)
6. [Isolamento de Segurança](#6-isolamento-de-segurança)
7. [Estratégia de API](#7-estratégia-de-api)
8. [Gestão de Identidades Cross-Bureau](#8-gestão-de-identidades-cross-bureau)
9. [Conformidade Regulatória Multi-Jurisdição](#9-conformidade-regulatória-multi-jurisdição)
10. [Considerações de Performance](#10-considerações-de-performance)
11. [Padrões de Implementação](#11-padrões-de-implementação)
12. [Monitoramento e Observabilidade](#12-monitoramento-e-observabilidade)
13. [Referências](#13-referências)

## 1. Introdução

Este documento detalha a arquitetura multi-tenant e multi-contexto implementada para a integração do módulo IAM com serviços de Bureau de Créditos em diferentes regiões e mercados. Esta arquitetura permite que uma única instância do sistema atenda a múltiplas organizações (tenants), mantendo isolamento completo entre seus dados, enquanto se adapta a diferentes contextos regulatórios, geográficos e operacionais.

### 1.1. Escopo

A arquitetura abrange:

- Modelo de dados multi-tenant para integrações com Bureau de Créditos
- Estratégias de isolamento em múltiplos níveis
- Adaptabilidade a diferentes contextos regulatórios
- Padrões de autorização e autenticação
- Gestão de identidades cross-bureau
- Mecanismos de auditoria e compliance

### 1.2. Objetivos Arquiteturais

1. **Isolamento Completo**: Garantir que dados de um tenant nunca sejam visíveis ou acessíveis por outro tenant
2. **Adaptabilidade Contextual**: Suportar diferentes requisitos regulatórios por região e país
3. **Eficiência Operacional**: Maximizar o compartilhamento de infraestrutura sem comprometer segurança
4. **Escalabilidade**: Suportar crescimento horizontal para milhares de tenants
5. **Auditabilidade**: Permitir trilhas de auditoria completas por tenant e por contexto
6. **Configurabilidade**: Possibilitar configurações específicas por tenant sem modificar código## 2. Princípios Arquiteturais

A arquitetura multi-tenant e multi-contexto para Bureau de Créditos é baseada nos seguintes princípios:

### 2.1. Princípio de Isolamento Completo

Toda operação, dado ou configuração relacionada a Bureau de Créditos deve ser explicitamente associada a um tenant específico, sem possibilidade de vazamento entre tenants diferentes.

### 2.2. Princípio da Adaptação Contextual

O sistema deve se adaptar automaticamente ao contexto regulatório, geográfico e operacional de cada tenant, aplicando regras específicas sem intervenção manual.

### 2.3. Princípio de Menor Privilégio

Usuários e processos devem ter acesso apenas aos dados mínimos necessários para realizar suas funções, mesmo dentro do mesmo tenant.

### 2.4. Princípio da Transparência

Todas as operações cross-tenant ou cross-contexto devem ser explicitamente autorizadas, registradas e auditáveis.

### 2.5. Princípio da Localidade de Dados

Dados sensíveis devem ser armazenados nas regiões geográficas apropriadas conforme requisitos regulatórios, com sharding geográfico automático.

### 2.6. Princípio da Eficiência Compartilhada

Código, configurações e infraestrutura devem ser compartilhados entre tenants sempre que não comprometer segurança e isolamento.

### 2.7. Princípio da Compliance por Design

Requisitos regulatórios devem ser implementados como parte da arquitetura, não como camada adicional.

## 3. Modelo de Multi-Tenancy para Bureau de Créditos

### 3.1. Hierarquia de Tenants

O sistema implementa uma estrutura hierárquica de tenants com quatro níveis principais:

```
Organização (Master Tenant)
  └─ Unidade de Negócio
      └─ Departamento
          └─ Projeto/Aplicação
```

Esta hierarquia permite:

- **Herança Controlada**: Configurações e políticas podem ser herdadas de níveis superiores
- **Delegação de Administração**: Gestão descentralizada sem comprometer segurança global
- **Isolamento Granular**: Possibilidade de isolamento até nível de projeto específico

### 3.2. Modelos de Tenancy Suportados

#### 3.2.1. Tenant Isolado (Database Dedicated)

- Database completamente separado por tenant
- Máximo isolamento e conformidade
- Recomendado para instituições financeiras de grande porte
- Suporta configurações de infraestrutura específicas

#### 3.2.2. Schema Isolado (Database Shared, Schema Dedicated)

- Schemas separados dentro do mesmo banco de dados
- Bom equilíbrio entre isolamento e eficiência
- Recomendado para a maioria dos clientes corporativos
- Implementado com PostgreSQL schema per tenant

#### 3.2.3. Row-Level Security (Database Shared, Schema Shared)

- Compartilhamento completo com isolamento via RLS
- Máxima densidade e eficiência
- Adequado para clientes menores ou menos sensíveis
- Implementado com políticas RLS do PostgreSQL### 3.3. Mecanismos de Controle de Tenant

#### 3.3.1. Resolução de Tenant

O sistema identifica o contexto do tenant através de múltiplas estratégias:

1. **JWT Claim**: Token de autenticação contém identificador de tenant
2. **Header HTTP**: `X-Tenant-ID` para comunicação entre serviços
3. **Subdomínio**: Resolução baseada em subdomínio (tenant.innovabiz.com)
4. **Path Parameter**: Como fallback para APIs específicas

#### 3.3.2. Middleware de Tenant

```typescript
// Middleware de resolução e validação de tenant
export const tenantMiddleware = async (req, res, next) => {
  try {
    // 1. Determinar tenant ID da requisição
    const tenantId = resolveTenantId(req);
    
    // 2. Validar se tenant existe e está ativo
    const tenant = await tenantService.validateTenant(tenantId);
    if (!tenant || !tenant.isActive) {
      return res.status(403).json({ 
        error: "TENANT_INVALID",
        message: "Invalid or inactive tenant" 
      });
    }
    
    // 3. Verificar se usuário tem acesso ao tenant
    const hasAccess = await tenantService.validateUserAccess(
      req.user.id, 
      tenantId
    );
    if (!hasAccess) {
      return res.status(403).json({ 
        error: "TENANT_ACCESS_DENIED",
        message: "User does not have access to this tenant" 
      });
    }
    
    // 4. Estabelecer contexto de tenant
    req.tenantId = tenantId;
    req.tenant = tenant;
    
    // 5. Configurar contexto de banco de dados
    await db.setTenantContext(tenantId);
    
    next();
  } catch (error) {
    next(error);
  }
};
```

#### 3.3.3. Contexto de Database

Para PostgreSQL, o contexto de tenant é estabelecido com:

```sql
-- Configurar variáveis de sessão para tenant
SELECT set_config('app.tenant_id', $1, false);
SELECT set_config('app.tenant_type', $2, false);
```

Para sistemas que usam schemas separados:

```sql
-- Definir schema search path para o tenant
SELECT set_config('search_path', 'tenant_' || $1 || ',public', false);
```

## 4. Arquitetura Multi-Contexto

### 4.1. Dimensões de Contexto

O sistema suporta adaptação contextual em múltiplas dimensões:

1. **Contexto Geográfico**: Adaptação a região/país específico
2. **Contexto Regulatório**: Conjunto de regras aplicáveis
3. **Contexto Operacional**: Modo de operação do tenant
4. **Contexto de Mercado**: Segmento de mercado do tenant
5. **Contexto Técnico**: Configurações técnicas específicas### 4.2. Implementação de Contexto Multi-Regulatório

O sistema implementa um modelo de políticas baseado em contexto regulatório que permite a adaptação dinâmica a diferentes jurisdições:

```typescript
// Exemplo de definição de contexto regulatório
const regulatoryContexts = {
  "angola": {
    "dataRetentionDays": 1825,  // 5 anos conforme BNA
    "requiredConsent": ["explicit", "written", "purpose_limited"],
    "dataResidency": "angola",
    "reportingRequirements": ["bna_monthly", "circ_quarterly"],
    "applicableLaws": ["lei_22_11", "aviso_12_2016"]
  },
  "mozambique": {
    "dataRetentionDays": 2555,  // 7 anos conforme Banco de Moçambique
    "requiredConsent": ["explicit", "multi_language", "purpose_limited"],
    "dataResidency": "mozambique",
    "reportingRequirements": ["bm_quarterly"],
    "applicableLaws": ["lei_15_2020", "aviso_10_GBM_2016"]
  },
  "south_africa": {
    "dataRetentionDays": 1825,  // 5 anos conforme NCR
    "requiredConsent": ["explicit", "revocable", "purpose_limited"],
    "dataResidency": "south_africa",
    "reportingRequirements": ["ncr_monthly", "popia_incidents"],
    "applicableLaws": ["national_credit_act", "popia"]
  }
};
```

### 4.3. Adaptadores de Contexto

O sistema utiliza adaptadores de contexto que modificam o comportamento dos serviços baseado no contexto atual:

```typescript
// Adaptador de contexto regulatório
class RegulatoryContextAdapter {
  constructor(tenantContext, regulatoryContext) {
    this.tenantContext = tenantContext;
    this.regulatoryContext = regulatoryContext;
  }
  
  // Aplica regras de retenção de dados específicas do contexto
  applyDataRetentionRules(data) {
    const retention = this.regulatoryContext.dataRetentionDays;
    const expiryDate = new Date();
    expiryDate.setDate(expiryDate.getDate() + retention);
    
    return {
      ...data,
      retentionExpiryDate: expiryDate,
      retentionPolicy: `Conform to ${this.regulatoryContext.applicableLaws.join(', ')}`
    };
  }
  
  // Valida consentimento conforme requisitos do contexto
  validateConsent(consentData) {
    const requiredTypes = this.regulatoryContext.requiredConsent;
    const missing = requiredTypes.filter(type => !consentData[type]);
    
    if (missing.length > 0) {
      throw new Error(`Consent validation failed. Missing types: ${missing.join(', ')}`);
    }
    
    return true;
  }
}
```

### 4.4. Resolução de Contexto Dinâmico

O sistema implementa um mecanismo de resolução de contexto que combina informações do tenant com configurações específicas:

```typescript
// Serviço de resolução de contexto
class ContextResolver {
  async resolveContext(tenantId, userId, operationType) {
    // 1. Buscar configurações básicas do tenant
    const tenant = await tenantRepository.findById(tenantId);
    
    // 2. Determinar contexto geográfico
    const geoContext = tenant.primaryLocation || 'global';
    
    // 3. Carregar contexto regulatório aplicável
    const regContext = await regulatoryContextRepository.findByLocation(
      geoContext, 
      tenant.industryType
    );
    
    // 4. Carregar configurações operacionais específicas
    const opContext = await operationalContextRepository.findForTenant(
      tenantId, 
      operationType
    );
    
    // 5. Combinar contextos em uma única configuração
    return this.mergeContexts(tenant, geoContext, regContext, opContext);
  }
  
  // Combina múltiplos contextos com precedência definida
  mergeContexts(tenant, geoContext, regContext, opContext) {
    // Implementação de lógica de precedência e mesclagem
    // ...
  }
}
```## 5. Estrutura de Dados

### 5.1. Modelo de Dados Multi-Tenant

A estrutura de dados para integração com Bureau de Créditos segue o padrão de discriminador de tenant em todas as tabelas:

#### 5.1.1. Tabelas Principais

```sql
-- Vínculos de identidade entre IAM e Bureau
CREATE TABLE bureau.identity_links (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES iam.tenants(id),
    user_id UUID NOT NULL REFERENCES iam.users(id),
    bureau_provider VARCHAR(50) NOT NULL,
    bureau_id VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    verification_status VARCHAR(20),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    metadata JSONB,
    
    CONSTRAINT uk_bureau_link UNIQUE(tenant_id, user_id, bureau_provider),
    CONSTRAINT check_status CHECK (status IN ('pending', 'active', 'revoked', 'expired'))
);

-- Autorizações de acesso ao Bureau
CREATE TABLE bureau.authorizations (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES iam.tenants(id),
    link_id UUID NOT NULL REFERENCES bureau.identity_links(id),
    purpose VARCHAR(100) NOT NULL,
    scope TEXT[] NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    created_by UUID REFERENCES iam.users(id),
    consent_id UUID,
    regulatory_context JSONB,
    
    CONSTRAINT check_status CHECK (status IN ('active', 'expired', 'revoked'))
);

-- Tokens de acesso para consultas ao Bureau
CREATE TABLE bureau.access_tokens (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES iam.tenants(id),
    authorization_id UUID NOT NULL REFERENCES bureau.authorizations(id),
    token_value TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    
    CONSTRAINT check_status CHECK (status IN ('active', 'expired', 'revoked'))
);

-- Histórico de consultas ao Bureau
CREATE TABLE bureau.queries (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES iam.tenants(id),
    token_id UUID NOT NULL REFERENCES bureau.access_tokens(id),
    query_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP,
    request_payload JSONB,
    response_payload JSONB,
    error_details JSONB,
    ip_address INET,
    user_agent TEXT,
    
    CONSTRAINT check_status CHECK (status IN ('pending', 'completed', 'failed'))
);
```

### 5.2. Políticas de Row-Level Security

Para cada tabela, políticas RLS são aplicadas para garantir isolamento:

```sql
-- Política RLS para identity_links
ALTER TABLE bureau.identity_links ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_policy ON bureau.identity_links
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- Política RLS para authorizations
ALTER TABLE bureau.authorizations ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_policy ON bureau.authorizations
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- Política RLS para access_tokens
ALTER TABLE bureau.access_tokens ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_policy ON bureau.access_tokens
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- Política RLS para queries
ALTER TABLE bureau.queries ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_policy ON bureau.queries
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);
```### 5.3. Metadados Específicos por Contexto

Cada registro relacionado ao Bureau de Créditos contém metadados específicos por contexto armazenados em campos JSONB:

```json
// Exemplo de metadados em identity_links
{
  "regulatoryContext": {
    "country": "Angola",
    "applicableLaws": ["lei_22_11", "aviso_12_2016"],
    "residencyRequirements": "local"
  },
  "verificationData": {
    "method": "document_validation",
    "validatedAt": "2025-08-15T14:30:00Z",
    "validatedBy": "system",
    "evidenceStored": true,
    "evidencePath": "tenant_123/verifications/link_456.enc"
  },
  "bureauSpecific": {
    "bureauId": "AO1234567890",
    "customerSince": "2023-05-12",
    "riskCategory": "low"
  }
}
```

## 6. Isolamento de Segurança

### 6.1. Estratégias de Isolamento em Múltiplas Camadas

O sistema implementa isolamento em múltiplas camadas para garantir a separação completa entre tenants:

#### 6.1.1. Isolamento de Dados

- **Row-Level Security (RLS)**: Filtros automáticos em nível de linha
- **Schema Isolation**: Para tenants com requisitos elevados de isolamento
- **Database Isolation**: Para tenants com máximo requisito de isolamento
- **Encrypted Data**: Dados sensíveis criptografados com chaves por tenant
- **Bucket Isolation**: Armazenamento de documentos e evidências em buckets separados

#### 6.1.2. Isolamento de Processamento

- **Contexto de Execução**: Variáveis de contexto em nível de thread/processo
- **Tenant-Aware Connection Pool**: Pools de conexão isolados por tenant
- **Circuit Breakers por Tenant**: Limites e proteções específicos por tenant

#### 6.1.3. Isolamento de API

- **Tenant-Aware Rate Limiting**: Limites de taxa específicos por tenant
- **Tenant Authorization**: Verificação de autorização específica por tenant
- **Tenant Headers**: Propagação segura de contexto entre serviços

### 6.2. Criptografia Multi-Tenant

O sistema implementa uma hierarquia de chaves de criptografia por tenant:

```
Chave Mestra (HSM)
  └─ Chave de Tenant (Derivada)
      └─ Chaves de Dados (Por tipo de dado)
          └─ Criptografia de Campo
```

Implementação em pseudocódigo:

```typescript
class TenantEncryptionService {
  // Derivar chave específica do tenant a partir da chave mestra
  async deriveTenantKey(tenantId) {
    const masterKey = await this.keyVault.getMasterKey();
    return crypto.hkdf(
      'sha256',
      masterKey,
      Buffer.from(tenantId, 'utf-8'),
      Buffer.from('TENANT_KEY_INFO'),
      32
    );
  }
  
  // Derivar chave para tipo específico de dado
  async deriveDataTypeKey(tenantId, dataType) {
    const tenantKey = await this.deriveTenantKey(tenantId);
    return crypto.hkdf(
      'sha256',
      tenantKey,
      Buffer.from(dataType, 'utf-8'),
      Buffer.from('DATA_TYPE_KEY_INFO'),
      32
    );
  }
  
  // Criptografar dados específicos do tenant
  async encryptData(tenantId, dataType, plaintext) {
    const dataTypeKey = await this.deriveDataTypeKey(tenantId, dataType);
    const iv = crypto.randomBytes(16);
    const cipher = crypto.createCipheriv('aes-256-gcm', dataTypeKey, iv);
    
    const encrypted = Buffer.concat([
      cipher.update(plaintext, 'utf8'),
      cipher.final()
    ]);
    
    const authTag = cipher.getAuthTag();
    
    return {
      iv: iv.toString('base64'),
      data: encrypted.toString('base64'),
      tag: authTag.toString('base64')
    };
  }
}
```## 7. Estratégia de API

### 7.1. Arquitetura da API Multi-Tenant

A API para integração com Bureau de Créditos segue uma arquitetura em camadas com isolamento de tenant em cada nível:

```
Cliente → API Gateway (Krakend) → Serviço Bureau → Adaptadores → Bureau de Créditos
  ↑                                     ↑                ↑
  └─ Tenant Header                      │                │
                                        └─ Tenant Context└─ Provider Específico
```

#### 7.1.1. Gateway API (Krakend)

Configuração Krakend para isolamento de tenant:

```json
{
  "endpoint": "/v1/bureau/credit-score/{userId}",
  "method": "GET",
  "extra_config": {
    "auth/validator": {
      "alg": "RS256",
      "jwk_url": "https://auth.innovabiz.com/.well-known/jwks.json",
      "disable_jwk_security": false,
      "operation_debug": false,
      "cache": true,
      "cache_duration": 3600,
      "propagate_claims": [
        ["tenant_id", "x-tenant-id"],
        ["roles", "x-roles"]
      ],
      "required_claims": ["tenant_id"]
    },
    "security/cors": {
      "allow_origins": ["*"],
      "allow_methods": ["GET", "POST", "PUT", "DELETE"],
      "allow_headers": ["Origin", "Authorization", "Content-Type", "X-Tenant-ID"],
      "expose_headers": ["Content-Length"],
      "max_age": "12h"
    },
    "qos/ratelimit/router": {
      "client_max_rate": 10,
      "tenant_identifier_strategy": "header:x-tenant-id",
      "tenant_rate_limits": {
        "default": 100,
        "premium_tenant": 500,
        "enterprise_tenant": 1000
      }
    }
  },
  "backend": [
    {
      "url_pattern": "/bureau-service/credit-score/{userId}",
      "method": "GET",
      "extra_config": {
        "modifier/martian": {
          "header.Set": {
            "scope": "request",
            "name": "X-Tenant-ID",
            "value": "{resp.header.x-tenant-id}"
          }
        }
      }
    }
  ]
}
```

### 7.2. API GraphQL Multi-Tenant

#### 7.2.1. Esquema GraphQL

```graphql
type Query {
  # Consultas específicas de tenant (requerem contexto de tenant)
  bureauIdentityLinks(
    filter: BureauLinkFilterInput
    pagination: PaginationInput
  ): BureauLinkConnection! @requireTenantContext
  
  bureauAuthorizations(
    filter: BureauAuthFilterInput
    pagination: PaginationInput
  ): BureauAuthConnection! @requireTenantContext
  
  bureauCreditScore(
    userId: ID!
    authorizationId: ID!
  ): CreditScore! @requireTenantContext @requireScope(scopes: ["bureau:score:read"])
  
  # Consulta multi-tenant (requer privilégios especiais)
  adminBureauStatistics(
    tenantId: ID!
  ): BureauStatistics! @requireScope(scopes: ["admin:bureau:read"])
}

# Diretiva para verificar contexto de tenant
directive @requireTenantContext on FIELD_DEFINITION

# Diretiva para verificar escopos de autorização
directive @requireScope(scopes: [String!]!) on FIELD_DEFINITION
```

#### 7.2.2. Resolvers com Isolamento

```typescript
// Resolver com isolamento de tenant
const bureauIdentityLinksResolver = async (parent, args, context, info) => {
  // 1. Validar contexto de tenant
  if (!context.tenant || !context.tenantId) {
    throw new AuthenticationError('Tenant context required');
  }
  
  // 2. Aplicar filtros específicos de tenant
  const filter = {
    ...args.filter,
    tenantId: context.tenantId // Garantir filtro por tenant
  };
  
  // 3. Executar query com contexto de tenant
  return await context.dataSources.bureauAPI.getIdentityLinks(filter, args.pagination);
};

// Implementação da diretiva requireTenantContext
class RequireTenantContextDirective extends SchemaDirectiveVisitor {
  visitFieldDefinition(field) {
    const { resolve = defaultFieldResolver } = field;
    
    field.resolve = async function (parent, args, context, info) {
      if (!context.tenant || !context.tenantId) {
        throw new AuthenticationError('This operation requires tenant context');
      }
      
      return resolve.call(this, parent, args, context, info);
    };
  }
}
```## 8. Gestão de Identidades Cross-Bureau

### 8.1. Modelo de Vinculação Multi-Bureau

O sistema suporta vinculação entre identidades do IAM e múltiplos Bureaus de Crédito, mantendo o isolamento por tenant:

```
IAM User ─┬─ Bureau Link (Tenant A, Bureau 1)
          ├─ Bureau Link (Tenant A, Bureau 2)
          ├─ Bureau Link (Tenant B, Bureau 1)
          └─ Bureau Link (Tenant B, Bureau 3)
```

#### 8.1.1. Estrutura de Vinculação

```typescript
interface BureauLink {
  id: string;
  tenantId: string;         // Tenant específico
  userId: string;           // Usuário IAM
  bureauProvider: string;   // Provedor de Bureau (e.g., "CIRC-BNA")
  bureauId: string;         // ID no provedor de Bureau
  status: LinkStatus;       // Estado do vínculo
  verificationStatus?: VerificationStatus;  // Estado de verificação
  createdAt: Date;
  updatedAt?: Date;
  metadata: Record<string, any>;  // Metadados específicos de contexto
}
```

### 8.2. Estratégias de Resolução de Identidade

O sistema implementa várias estratégias para resolução de identidades entre o IAM e os Bureaus de Crédito:

#### 8.2.1. Estratégia de Matching Direto

```typescript
class DirectMatchingStrategy implements IdentityResolutionStrategy {
  async resolveIdentity(userData: UserData, bureauProvider: string): Promise<string | null> {
    // Mapeia diretamente baseado em documentos oficiais
    const documentMapping = {
      "CIRC-BNA": userData.nationalIdNumber,       // Angola - BI
      "CIC-BM": userData.nationalIdNumber,         // Moçambique - BI
      "TransUnion-SA": userData.nationalIdNumber,  // África do Sul - ID Number
      // Outros mapeamentos específicos por país
    };
    
    const bureauId = documentMapping[bureauProvider];
    if (!bureauId) {
      return null;
    }
    
    // Validação adicional baseada em contexto regulatório
    return this.validateIdentity(bureauId, bureauProvider, userData);
  }
}
```

#### 8.2.2. Estratégia de Matching Fuzzy

Para casos onde não há correspondência exata (mercados diferentes, variações de nome, etc.):

```typescript
class FuzzyMatchingStrategy implements IdentityResolutionStrategy {
  async resolveIdentity(userData: UserData, bureauProvider: string): Promise<string | null> {
    // Gera candidatos para matching
    const candidates = await this.generateCandidates(userData, bureauProvider);
    
    // Calcula scores para cada candidato
    const scoredCandidates = await this.scoreCandidates(candidates, userData);
    
    // Aplica threshold baseado no contexto regulatório
    const regulatoryContext = await this.getRegulationContext(bureauProvider);
    const threshold = regulatoryContext.matchingThreshold || 0.9;
    
    // Retorna o melhor candidato acima do threshold
    const bestMatch = scoredCandidates[0];
    if (bestMatch && bestMatch.score >= threshold) {
      return bestMatch.bureauId;
    }
    
    return null;
  }
}
```

### 8.3. Gestão de Consentimento Multi-Tenant e Multi-Contexto

A gestão de consentimento adapta-se automaticamente ao contexto regulatório do tenant:

```typescript
class ConsentManager {
  async createConsent(userId: string, tenantId: string, purpose: string): Promise<Consent> {
    // 1. Determinar contexto regulatório
    const tenant = await tenantRepository.findById(tenantId);
    const regulatoryContext = await regulatoryContextRepository.findByCountry(
      tenant.country
    );
    
    // 2. Determinar requisitos de consentimento
    const consentRequirements = regulatoryContext.consentRequirements;
    
    // 3. Gerar modelo de consentimento específico do contexto
    const consentTemplate = await this.getConsentTemplate(
      tenant.language,
      purpose,
      regulatoryContext.id
    );
    
    // 4. Determinar validade baseada no contexto regulatório
    const validityDays = regulatoryContext.consentValidityDays || 90;
    const expiresAt = new Date();
    expiresAt.setDate(expiresAt.getDate() + validityDays);
    
    // 5. Criar registro de consentimento
    return await consentRepository.create({
      userId,
      tenantId,
      purpose,
      consentText: consentTemplate.text,
      templateVersion: consentTemplate.version,
      regulatoryContext: regulatoryContext.id,
      status: 'pending',
      createdAt: new Date(),
      expiresAt: expiresAt,
      dataCategories: consentTemplate.dataCategories
    });
  }
}
```## 9. Conformidade Regulatória Multi-Jurisdição

### 9.1. Modelo de Compliance Adaptativo

A arquitetura implementa um modelo de compliance adaptativo que aplica automaticamente regras específicas por jurisdição:

```
Tenant → País/Região → Regulamentações Aplicáveis → Regras de Compliance
```

#### 9.1.1. Catálogo de Regulamentações

```typescript
// Catálogo centralizado de regulamentos por jurisdição
const regulatoryCatalog = {
  "angola": {
    "primaryRegulator": "BNA",
    "regulations": [
      {
        "id": "lei_22_11",
        "name": "Lei 22/11 - Lei das Instituições Financeiras",
        "dataRetention": {
          "creditData": 60,  // meses
          "consentRecords": 120  // meses
        },
        "consentRequirements": ["explicit", "written", "purpose_limited"],
        "reportingRequirements": [
          {
            "type": "regulatory_report",
            "recipient": "BNA",
            "frequency": "monthly",
            "deadline": "15th day of following month"
          }
        ]
      },
      {
        "id": "aviso_12_2016",
        "name": "Aviso 12/2016 do BNA",
        "dataProtectionRequirements": ["encryption", "access_control", "audit_trail"],
        "applicableFrom": "2016-06-01"
      }
    ]
  },
  "mozambique": {
    "primaryRegulator": "Banco de Moçambique",
    "regulations": [
      {
        "id": "lei_15_2020",
        "name": "Lei 15/2020 - Proteção de Dados",
        "dataRetention": {
          "creditData": 84,  // meses
          "consentRecords": 120  // meses
        },
        "consentRequirements": ["explicit", "multi_language", "purpose_limited"],
        "reportingRequirements": [
          {
            "type": "regulatory_report",
            "recipient": "Banco de Moçambique",
            "frequency": "quarterly",
            "deadline": "30 days after quarter end"
          }
        ]
      }
    ]
  },
  "south_africa": {
    "primaryRegulator": "National Credit Regulator",
    "regulations": [
      {
        "id": "national_credit_act",
        "name": "National Credit Act",
        "dataRetention": {
          "creditData": 60,  // meses
          "consentRecords": 60  // meses
        },
        "consentRequirements": ["explicit", "clear_language", "purpose_limited"]
      },
      {
        "id": "popia",
        "name": "Protection of Personal Information Act",
        "dataProtectionRequirements": ["encryption", "access_control", "audit_trail", "impact_assessment"],
        "applicableFrom": "2021-07-01"
      }
    ]
  }
};
```

### 9.2. Validadores de Compliance

A arquitetura implementa validadores que verificam automaticamente a conformidade com requisitos regulatórios:

```typescript
class ComplianceValidator {
  // Validador principal que orquestra validações específicas
  async validateOperation(operation: BureauOperation): Promise<ValidationResult> {
    // 1. Determinar contexto regulatório
    const tenant = await tenantRepository.findById(operation.tenantId);
    const regulations = await this.getApplicableRegulations(tenant);
    
    // 2. Executar validações específicas
    const validationResults = await Promise.all(
      regulations.map(reg => this.validateAgainstRegulation(operation, reg))
    );
    
    // 3. Agregação de resultados
    const isValid = validationResults.every(r => r.isValid);
    const issues = validationResults.flatMap(r => r.issues);
    
    // 4. Registrar tentativa para auditoria
    await this.recordValidationAttempt(operation, {
      isValid,
      issues,
      regulations: regulations.map(r => r.id)
    });
    
    return { isValid, issues };
  }
  
  // Validador específico por regulamentação
  async validateAgainstRegulation(operation: BureauOperation, regulation: Regulation): Promise<ValidationResult> {
    const validators = {
      "lei_22_11": this.validateLei22_11,
      "aviso_12_2016": this.validateAviso12_2016,
      "lei_15_2020": this.validateLei15_2020,
      "national_credit_act": this.validateNationalCreditAct,
      "popia": this.validatePopia
    };
    
    const validator = validators[regulation.id];
    if (!validator) {
      return { isValid: true, issues: [] }; // Sem validador específico
    }
    
    return await validator.call(this, operation, regulation);
  }
}
```### 9.3. Auditoria Multi-Tenant e Multi-Contexto

A arquitetura implementa um sistema de auditoria que registra operações com contextualização completa:

```typescript
class AuditService {
  // Registra evento de auditoria com contexto completo
  async logAuditEvent(event: BureauAuditEvent): Promise<void> {
    // 1. Enriquecer evento com contexto de tenant
    const tenant = await tenantRepository.findById(event.tenantId);
    const enrichedEvent = {
      ...event,
      tenantDetails: {
        name: tenant.name,
        type: tenant.tenantType,
        country: tenant.country
      }
    };
    
    // 2. Adicionar contexto regulatório
    const regulatoryContext = await regulatoryContextRepository.findByCountry(
      tenant.country
    );
    enrichedEvent.regulatoryContext = {
      regulations: regulatoryContext.regulations.map(r => r.id),
      jurisdiction: regulatoryContext.jurisdiction
    };
    
    // 3. Armazenar com isolamento por tenant
    await auditLogRepository.create(enrichedEvent);
    
    // 4. Emitir evento para sistemas externos (se necessário)
    if (event.severity >= AuditSeverity.HIGH) {
      await this.notifySecurityTeam(enrichedEvent);
    }
    
    // 5. Verificar conformidade com requisitos de retenção
    await this.scheduleRetentionCheck(
      event.tenantId, 
      regulatoryContext.auditRetentionPeriod
    );
  }
}
```

## 10. Considerações de Performance

### 10.1. Estratégias de Cache Multi-Tenant

A arquitetura implementa estratégias de cache específicas para o ambiente multi-tenant:

#### 10.1.1. Particionamento de Cache por Tenant

```typescript
class TenantAwareCache {
  // Cache particionado por tenant
  private cache: Map<string, Map<string, CacheEntry>> = new Map();
  
  // Gera chave de cache com namespace por tenant
  private generateKey(tenantId: string, key: string): string {
    return `tenant:${tenantId}:${key}`;
  }
  
  // Armazena valor em cache específico do tenant
  async set(tenantId: string, key: string, value: any, ttlSeconds: number): Promise<void> {
    // Inicializar cache do tenant se não existir
    if (!this.cache.has(tenantId)) {
      this.cache.set(tenantId, new Map());
    }
    
    const tenantCache = this.cache.get(tenantId)!;
    const expiresAt = Date.now() + (ttlSeconds * 1000);
    
    tenantCache.set(key, {
      value,
      expiresAt
    });
  }
  
  // Recupera valor do cache específico do tenant
  async get<T>(tenantId: string, key: string): Promise<T | null> {
    const tenantCache = this.cache.get(tenantId);
    if (!tenantCache) {
      return null;
    }
    
    const entry = tenantCache.get(key);
    if (!entry) {
      return null;
    }
    
    // Verificar expiração
    if (entry.expiresAt < Date.now()) {
      tenantCache.delete(key);
      return null;
    }
    
    return entry.value as T;
  }
  
  // Limpa cache específico do tenant
  async invalidateTenantCache(tenantId: string): Promise<void> {
    this.cache.delete(tenantId);
  }
}
```

### 10.2. Otimizações para Consultas Multi-Tenant

A arquitetura implementa otimizações específicas para consultas em ambiente multi-tenant:

#### 10.2.1. Índices Específicos para Multi-Tenancy

```sql
-- Índices otimizados para consultas multi-tenant
CREATE INDEX idx_bureau_links_tenant_user ON bureau.identity_links (tenant_id, user_id);
CREATE INDEX idx_bureau_auth_tenant_status ON bureau.authorizations (tenant_id, status);
CREATE INDEX idx_bureau_tokens_tenant_expires ON bureau.access_tokens (tenant_id, expires_at);

-- Índices para consultas mais frequentes dentro do tenant
CREATE INDEX idx_bureau_links_provider_status ON bureau.identity_links (tenant_id, bureau_provider, status);
CREATE INDEX idx_bureau_auth_purpose ON bureau.authorizations (tenant_id, purpose, status);
```

#### 10.2.2. Query Planning Multi-Tenant

```typescript
class QueryOptimizer {
  // Otimiza consultas com base em estatísticas específicas do tenant
  optimizeQuery(query: QueryTemplate, tenantId: string): OptimizedQuery {
    // 1. Carregar estatísticas específicas do tenant
    const tenantStats = this.getTenantStatistics(tenantId);
    
    // 2. Decidir estratégia baseada em volume de dados
    if (tenantStats.recordCount > 1000000) {
      // Para tenants grandes, usar estratégia de paginação agressiva
      return this.applyLargeTenantOptimizations(query, tenantStats);
    } else if (tenantStats.recordCount > 100000) {
      // Para tenants médios
      return this.applyMediumTenantOptimizations(query, tenantStats);
    } else {
      // Para tenants pequenos
      return this.applySmallTenantOptimizations(query, tenantStats);
    }
  }
}
```## 11. Padrões de Implementação

### 11.1. Repository Pattern Multi-Tenant

```typescript
// Interface base para repositórios com awareness de tenant
interface TenantAwareRepository<T> {
  findById(id: string, tenantId: string): Promise<T | null>;
  findAll(tenantId: string, filter?: any): Promise<T[]>;
  create(data: Partial<T>, tenantId: string): Promise<T>;
  update(id: string, data: Partial<T>, tenantId: string): Promise<T>;
  delete(id: string, tenantId: string): Promise<boolean>;
}

// Implementação para repositório de vínculos do Bureau
class BureauLinkRepository implements TenantAwareRepository<BureauLink> {
  // Implementação base que garante isolamento por tenant
  async findById(id: string, tenantId: string): Promise<BureauLink | null> {
    // Garantir que a query inclui sempre o filtro de tenant
    const result = await db.query(
      'SELECT * FROM bureau.identity_links WHERE id = $1 AND tenant_id = $2',
      [id, tenantId]
    );
    
    return result.rows[0] || null;
  }
  
  async findAll(tenantId: string, filter?: any): Promise<BureauLink[]> {
    // Construir query dinâmica com filtros adicionais
    let query = 'SELECT * FROM bureau.identity_links WHERE tenant_id = $1';
    const params = [tenantId];
    
    // Adicionar filtros adicionais sempre mantendo tenant_id
    if (filter) {
      let paramIndex = 2;
      if (filter.status) {
        query += ` AND status = $${paramIndex}`;
        params.push(filter.status);
        paramIndex++;
      }
      
      if (filter.bureauProvider) {
        query += ` AND bureau_provider = $${paramIndex}`;
        params.push(filter.bureauProvider);
        paramIndex++;
      }
    }
    
    const result = await db.query(query, params);
    return result.rows;
  }
  
  async create(data: Partial<BureauLink>, tenantId: string): Promise<BureauLink> {
    // Garantir que o tenant_id seja sempre o do contexto atual
    const linkData = {
      ...data,
      tenant_id: tenantId,
      id: uuidv4(),
      created_at: new Date()
    };
    
    const columns = Object.keys(linkData);
    const placeholders = columns.map((_, i) => `$${i + 1}`).join(', ');
    const values = Object.values(linkData);
    
    const query = `
      INSERT INTO bureau.identity_links (${columns.join(', ')})
      VALUES (${placeholders})
      RETURNING *
    `;
    
    const result = await db.query(query, values);
    return result.rows[0];
  }
}
```

### 11.2. Service Layer Multi-Tenant

```typescript
// Service layer que implementa lógica de negócio com isolamento de tenant
class BureauService {
  constructor(
    private linkRepository: BureauLinkRepository,
    private authRepository: BureauAuthorizationRepository,
    private tokenRepository: BureauTokenRepository,
    private tenantService: TenantService,
    private contextResolver: ContextResolver
  ) {}
  
  // Cria vínculo de identidade com contexto de tenant
  async createIdentityLink(
    userId: string, 
    tenantId: string, 
    linkData: BureauLinkCreationData
  ): Promise<BureauLink> {
    // 1. Validar acesso do usuário ao tenant
    await this.tenantService.validateUserAccess(userId, tenantId);
    
    // 2. Resolver contexto regulatório para o tenant
    const context = await this.contextResolver.resolveContext(
      tenantId, 
      userId, 
      'identity_linking'
    );
    
    // 3. Validar dados conforme contexto regulatório
    this.validateLinkDataForContext(linkData, context);
    
    // 4. Criar vínculo com metadados de contexto
    return await this.linkRepository.create({
      user_id: userId,
      bureau_provider: linkData.bureauProvider,
      bureau_id: linkData.bureauId,
      status: 'pending',
      metadata: {
        regulatoryContext: context.regulatoryContextId,
        verificationData: linkData.verificationData
      }
    }, tenantId);
  }
}
```

### 11.3. Factory Pattern para Adaptadores de Bureau

```typescript
// Factory para criar adaptadores específicos por país/provedor
class BureauAdapterFactory {
  // Registra adaptadores disponíveis
  private adapters: Record<string, BureauAdapterConstructor> = {
    "angola_bna": AngolaBnaAdapter,
    "mozambique_bm": MozambiqueBmAdapter,
    "southafrica_transunion": SouthAfricaTransunionAdapter,
    // Outros adaptadores
  };
  
  // Cria adaptador baseado no provedor e contexto
  async createAdapter(
    bureauProvider: string,
    tenantId: string
  ): Promise<BureauAdapter> {
    // 1. Determinar adaptador correto
    const [country, provider] = bureauProvider.split('_');
    const adapterKey = `${country}_${provider}`;
    
    const AdapterClass = this.adapters[adapterKey];
    if (!AdapterClass) {
      throw new Error(`No adapter found for provider: ${bureauProvider}`);
    }
    
    // 2. Carregar configuração específica do tenant
    const config = await this.loadAdapterConfig(tenantId, adapterKey);
    
    // 3. Instanciar adaptador com configuração
    return new AdapterClass(config);
  }
  
  // Carrega configuração específica do tenant para o adaptador
  private async loadAdapterConfig(
    tenantId: string, 
    adapterKey: string
  ): Promise<BureauAdapterConfig> {
    // Buscar configuração no repositório de configurações de tenant
    const tenantConfig = await tenantConfigRepository.findByTenant(
      tenantId,
      `bureau_adapter_${adapterKey}`
    );
    
    // Mesclar com configuração padrão
    const defaultConfig = await defaultConfigRepository.findByKey(
      `bureau_adapter_${adapterKey}`
    );
    
    return {
      ...defaultConfig,
      ...tenantConfig,
      tenantId
    };
  }
}
```## 12. Considerações Finais

### 12.1. Benefícios da Arquitetura Multi-Tenant e Multi-Contexto

A arquitetura proposta oferece diversos benefícios:

1. **Isolamento Completo**: Garantia de separação entre dados de diferentes tenants em todos os níveis.
2. **Adaptabilidade Regulatória**: Capacidade de adaptar-se dinamicamente às exigências regulatórias de diferentes jurisdições.
3. **Eficiência Operacional**: Compartilhamento de infraestrutura e código entre tenants sem comprometer isolamento.
4. **Escalabilidade**: Capacidade de escalar horizontalmente para acomodar novos tenants e mercados.
5. **Conformidade Incorporada**: Mecanismos de validação e auditoria integrados em cada camada do sistema.

### 12.2. Recomendações de Implementação

Para uma implementação bem-sucedida da arquitetura multi-tenant e multi-contexto para Bureau de Créditos, recomenda-se:

1. **Implementação Faseada**:
   - Fase 1: Estrutura básica multi-tenant com isolamento de dados
   - Fase 2: Adaptadores de Bureau específicos por país
   - Fase 3: Framework de contexto regulatório dinâmico
   - Fase 4: Sistema avançado de auditoria e compliance

2. **Testes de Isolamento**:
   - Implementar testes automatizados para verificar isolamento entre tenants
   - Realizar auditorias de segurança periódicas para validar isolamento
   - Simular cenários de "tenant confusion" para testar mitigações

3. **Monitoramento Específico**:
   - Implementar métricas específicas por tenant
   - Criar dashboards de performance por tenant
   - Configurar alertas para anomalias específicas por tenant

### 12.3. Próximos Passos

Para avançar com a implementação da arquitetura multi-tenant e multi-contexto para Bureau de Créditos, os próximos passos incluem:

1. **Refinamento dos Adaptadores de Bureau**:
   - Desenvolver adaptadores específicos para cada Bureau nos mercados PALOP e SADC
   - Implementar testes específicos para cada adaptador
   - Documentar APIs específicas de cada Bureau

2. **Enriquecimento do Catálogo Regulatório**:
   - Expandir catálogo para incluir regulamentações adicionais
   - Implementar validadores específicos para cada regulamentação
   - Criar documentação detalhada sobre requisitos por país

3. **Automação de Compliance**:
   - Desenvolver ferramentas para verificação automática de compliance
   - Implementar dashboards de status de compliance por tenant
   - Criar sistema de alertas para não-conformidades

4. **Expansão para Novos Mercados**:
   - Mapear requisitos para mercados adicionais
   - Desenvolver adaptadores para provedores adicionais
   - Criar guias de implementação específicos por país

## 13. Glossário de Termos

| Termo | Definição |
|-------|-----------|
| **Tenant** | Organização ou entidade que utiliza o sistema como um serviço isolado. |
| **Multi-Tenancy** | Arquitetura onde uma única instância do software atende múltiplos tenants isolados. |
| **Contexto** | Conjunto de configurações, regras e comportamentos específicos para um determinado ambiente (país, regulação, tenant). |
| **Bureau de Créditos** | Entidade que coleta, armazena e distribui informações de crédito de indivíduos e empresas. |
| **RLS** | Row-Level Security - Mecanismo de segurança para filtrar acesso a dados no nível de linha. |
| **Adaptador de Bureau** | Componente que traduz entre o sistema IAM e um Bureau de Créditos específico. |
| **Isolamento de Tenant** | Garantia de que dados e operações de um tenant não sejam acessíveis por outro tenant. |
| **Contexto Regulatório** | Conjunto de regras e requisitos legais aplicáveis a um determinado mercado ou jurisdição. |
| **PALOP** | Países Africanos de Língua Oficial Portuguesa. |
| **SADC** | Southern African Development Community - Comunidade para o Desenvolvimento da África Austral. |