# Arquitetura Técnica do MCP-IAM INNOVABIZ

## 1. Visão Geral da Arquitetura

O MCP-IAM (Model Context Protocol - Identity and Access Management) representa o núcleo de segurança da plataforma INNOVABIZ, implementando uma arquitetura de identidade e acesso multi-dimensional que se adapta dinamicamente aos contextos regulatórios, operacionais e de negócio em múltiplos mercados e clientes.

### 1.1. Princípios Arquiteturais

1. **Segurança por Design**: Segurança implementada em todas as camadas e componentes
2. **Multi-Dimensionalidade**: Adaptação a múltiplos contextos, tenants e mercados
3. **Zero-Trust**: Verificação contínua de identidade e autorizações
4. **Privacy by Default**: Proteção de dados incorporada na arquitetura
5. **Resiliência**: Alta disponibilidade e tolerância a falhas
6. **Observabilidade Total**: Visibilidade completa de todas operações de segurança
7. **Extensibilidade**: Arquitetura modular e baseada em hooks
8. **Compliance Automatizada**: Adaptação dinâmica a requisitos regulatórios

### 1.2. Visão de Alto Nível

```mermaid
graph TD
    A[Aplicações Cliente] --> B[API Gateway / Krakend]
    B --> C[IAM Core Services]
    
    subgraph "IAM Core Services"
        C1[Authentication Service]
        C2[Authorization Service]
        C3[User Management]
        C4[Privilege Elevation]
        C5[MFA Service]
        C6[Audit Service]
        C7[Policy Engine]
        C8[Compliance Service]
    end
    
    C --> D[Storage Layer]
    
    subgraph "Storage Layer"
        D1[PostgreSQL - User Data]
        D2[Redis - Token Cache]
        D3[TimescaleDB - Audit Logs]
        D4[Neo4j - Authorization Graph]
    end
    
    C --> E[Integration Layer]
    
    subgraph "Integration Layer"
        E1[MCP Hooks]
        E2[GraphQL Federation]
        E3[REST APIs]
        E4[gRPC Services]
        E5[Event Streaming]
    end
    
    E --> F[Módulos Core INNOVABIZ]
    
    subgraph "Módulos Core INNOVABIZ"
        F1[Payment Gateway]
        F2[Mobile Money]
        F3[Marketplace]
        F4[E-Commerce]
        F5[Microcrédito]
        F6[Seguros]
    end
```

## 2. Componentes da Arquitetura

### 2.1. Autenticação

#### 2.1.1. Serviços de Autenticação

| Serviço | Responsabilidade | Tecnologia |
|---------|------------------|------------|
| **Token Service** | Geração, validação e revogação de tokens JWT | Go + PASETO |
| **Credential Manager** | Validação de credenciais e senhas | Argon2id + Go |
| **Federation Service** | Integração com IdPs externos (OIDC, SAML) | Go + node-oidc-provider |
| **Session Manager** | Gestão de sessões e dispositivos confiáveis | Redis + PostgreSQL |
| **Biometric Auth** | Autenticação biométrica (facial, impressão digital) | Go + AWS Rekognition |

#### 2.1.2. Fluxos de Autenticação

```mermaid
sequenceDiagram
    participant U as Usuário
    participant C as Cliente (App/Web)
    participant AG as API Gateway
    participant AS as Auth Service
    participant MFA as MFA Service
    participant DS as Device Service
    participant AuS as Audit Service
    
    U->>C: Inicia login
    C->>AG: Autenticação (username/password)
    AG->>AS: Valida credenciais
    AS->>DS: Verifica dispositivo
    
    alt Dispositivo conhecido
        DS->>AS: Baixo risco - Prosseguir
    else Dispositivo desconhecido
        DS->>AS: Alto risco - Requisitar MFA
        AS->>MFA: Solicita verificação
        MFA->>C: Envia desafio MFA
        C->>U: Apresenta desafio MFA
        U->>C: Fornece MFA
        C->>MFA: Envia resposta MFA
        MFA->>AS: Confirma verificação
    end
    
    AS->>AuS: Registra evento de login
    AS->>AG: Retorna tokens (access, refresh)
    AG->>C: Tokens autenticados
    C->>U: Acesso concedido
```

#### 2.1.3. Estrutura de Token JWT

```json
{
  "header": {
    "alg": "ES384",
    "typ": "JWT",
    "kid": "key-id-1"
  },
  "payload": {
    "iss": "https://iam.innovabiz.com",
    "sub": "user123",
    "aud": ["payment-gateway", "mobile-money"],
    "exp": 1627968000,
    "iat": 1627964400,
    "jti": "random-unique-id",
    "tenant_id": "tenant-abc",
    "market": "angola",
    "roles": ["payment_agent", "support_l1"],
    "scopes": ["payment:read", "transaction:list"],
    "context": {
      "device_id": "dev-xyz",
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "mfa_verified": true,
      "risk_score": 15,
      "elevation_id": null
    }
  }
}
```

### 2.2. Autorização

#### 2.2.1. Serviços de Autorização

| Serviço | Responsabilidade | Tecnologia |
|---------|------------------|------------|
| **Policy Engine** | Avaliação de políticas de autorização | OPA (Rego) |
| **Authorization Graph** | Gerenciamento do grafo de autorização | Neo4j |
| **Permission Service** | Gerenciamento de permissões e papéis | PostgreSQL + Go |
| **Context Provider** | Enriquecimento de contexto para decisões | Go |
| **Role Manager** | Gerenciamento de RBAC e ABAC | PostgreSQL + Redis |

#### 2.2.2. Modelo de Autorização Multi-Dimensional

```mermaid
graph TD
    A[Requisição de Acesso] --> B{Policy Engine}
    B --> C{Quem?}
    B --> D{O quê?}
    B --> E{Onde?}
    B --> F{Quando?}
    B --> G{Como?}
    B --> H{Por quê?}
    
    C --> C1[Identity]
    C --> C2[Role]
    C --> C3[Groups]
    
    D --> D1[Resource Type]
    D --> D2[Operation]
    D --> D3[Data Classification]
    
    E --> E1[Tenant]
    E --> E2[Market]
    E --> E3[Region]
    
    F --> F1[Time Window]
    F --> F2[Date Restrictions]
    F --> F3[Business Hours]
    
    G --> G1[Device]
    G --> G2[Network]
    G --> G3[Authentication Level]
    
    H --> H1[Purpose]
    H --> H2[Business Context]
    H --> H3[Justification]
    
    C1 --> I[Decision Engine]
    C2 --> I
    C3 --> I
    D1 --> I
    D2 --> I
    D3 --> I
    E1 --> I
    E2 --> I
    E3 --> I
    F1 --> I
    F2 --> I
    F3 --> I
    G1 --> I
    G2 --> I
    G3 --> I
    H1 --> I
    H2 --> I
    H3 --> I
    
    I --> J[Decisão de Autorização]
    J --> K[Auditoria]
```

#### 2.2.3. Exemplo de Política OPA (Rego)

```rego
# Política para acesso a transações de pagamento
package payment

import data.roles
import data.tenants
import data.markets.regulations

# Regra default - negar acesso
default allow = false

# Permitir acesso a transações
allow {
    # Verificar tipo de operação
    input.action == "read"
    input.resource_type == "transaction"
    
    # Verificar role do usuário
    user_has_role("payment_agent")
    
    # Verificar tenant
    same_tenant
    
    # Verificar compliance com regulações do mercado
    compliant_with_market_regulations
}

# Permitir acesso completo a transações para suporte avançado
allow {
    # Verificar tipo de operação
    input.action == "read"
    input.resource_type == "transaction"
    
    # Verificar role do usuário e elevação de privilégios
    user_has_role("support_l2")
    has_active_elevation
    
    # Verificar tenant
    same_tenant
    
    # Verificar compliance com regulações do mercado
    compliant_with_market_regulations
    
    # Registrar acesso como sensível para auditoria
    audit_sensitive_access
}

# Funções auxiliares
user_has_role(role) {
    input.user.roles[_] == role
}

same_tenant {
    input.user.tenant_id == input.resource.tenant_id
}

has_active_elevation {
    input.user.context.elevation_id != null
    elevation := data.elevations[input.user.context.elevation_id]
    elevation.status == "active"
    elevation.expiration > current_time_ms
}

compliant_with_market_regulations {
    market := input.user.market
    regulation_checks := regulations[market].transaction_access
    
    # Executar todas verificações específicas do mercado
    regulation_checks.min_auth_level <= input.user.auth_level
    
    # Condições específicas para mercados que exigem verificações adicionais
    not requires_special_checks[market] or special_check_passed[market]
}

audit_sensitive_access {
    # Esta função tem efeito colateral via API para registrar acesso sensível
    http.send({
        "method": "POST",
        "url": "http://audit-service/log",
        "body": {
            "user_id": input.user.id,
            "action": input.action,
            "resource": input.resource,
            "sensitivity": "high",
            "elevation_id": input.user.context.elevation_id
        }
    })
}

# Verificações específicas por mercado
requires_special_checks = {
    "angola": true,
    "brazil": true
}

special_check_passed = {
    "angola": input.user.kyc_level == "full",
    "brazil": input.user.context.purpose_validated == true
}
```