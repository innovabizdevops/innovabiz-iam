# Documentação da API IAM

## Visão Geral

Este documento descreve a API de Gerenciamento de Identidade e Acesso (IAM) para a plataforma INNOVABIZ. A API fornece serviços abrangentes de identidade, autenticação, autorização e conformidade para consumo interno e externo.

## Princípios de Design da API

- **Design API-First**: Toda a funcionalidade é exposta através de APIs consistentes
- **RESTful e GraphQL**: Suporte para ambos os paradigmas com base no caso de uso
- **Segurança por Padrão**: Segurança incorporada no design central da API
- **Versionamento**: Suporte à evolução da API com compatibilidade retroativa
- **Experiência do Desenvolvedor**: Interfaces claras, previsíveis e bem documentadas
- **Multi-Tenancy**: Contexto de tenant aplicado em todas as operações

## URLs Base

- **API REST**: `https://{tenant-id}.api.innovabiz.com/iam/v1`
- **API GraphQL**: `https://{tenant-id}.api.innovabiz.com/iam/graphql`
- **OAuth/OIDC**: `https://{tenant-id}.auth.innovabiz.com`

## Autenticação e Autorização

### Autenticação na API

Todas as chamadas de API devem ser autenticadas usando um dos seguintes métodos:

1. **Token Bearer OAuth 2.1**:
   ```
   Authorization: Bearer {access_token}
   ```

2. **Chave de API** (para serviço-a-serviço):
   ```
   X-API-Key: {api_key}
   ```

3. **TLS Mútuo** (para integrações de alta segurança)

### Contexto de Tenant

Todas as requisições devem incluir o contexto do tenant:

1. **Baseado em URL**: Usando o subdomínio ou ID do tenant na URL
2. **Baseado em Cabeçalho**:
   ```
   X-Tenant-ID: {tenant-id}
   ```

## Limitação de Taxa

Limites de taxa são aplicados por tenant e chave de API:

- **Nível Padrão**: 100 requisições/minuto
- **Nível Empresarial**: 1000 requisições/minuto
- **Nível Personalizado**: Configurável com base nos requisitos

Cabeçalhos de limite de taxa são incluídos em todas as respostas:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1614556800
```

## Tratamento de Erros

Todos os erros seguem um formato consistente:

```json
{
  "error": {
    "code": "AUTHENTICATION_FAILED",
    "message": "Credenciais inválidas fornecidas",
    "details": "O token JWT fornecido expirou",
    "request_id": "f7a8c39e-8dfc-42f9-9738-f82c6a99a354"
  }
}
```

Códigos de erro comuns:
- `AUTHENTICATION_FAILED`: Problemas de autenticação
- `AUTHORIZATION_FAILED`: Problemas de permissão
- `VALIDATION_ERROR`: Entrada inválida
- `RESOURCE_NOT_FOUND`: Recurso solicitado não existe
- `RATE_LIMIT_EXCEEDED`: Limite de taxa da API alcançado
- `TENANT_CONTEXT_MISSING`: Nenhum contexto de tenant fornecido
- `INTERNAL_ERROR`: Erro no lado do servidor

## Serviços Principais da API

### API de Gerenciamento de Identidade

Endpoints para gerenciar usuários, grupos, papéis e permissões.

#### Gerenciamento de Usuários

| Método | Endpoint | Descrição |
|--------|---------|-------------|
| `GET` | `/users` | Listar usuários com paginação e filtragem |
| `POST` | `/users` | Criar um novo usuário |
| `GET` | `/users/{id}` | Obter um usuário específico |
| `PUT` | `/users/{id}` | Atualizar um usuário |
| `DELETE` | `/users/{id}` | Excluir um usuário |
| `GET` | `/users/{id}/groups` | Listar grupos para um usuário |
| `GET` | `/users/{id}/roles` | Listar papéis para um usuário |
| `GET` | `/users/{id}/permissions` | Listar permissões efetivas |

Exemplo de requisição:
```
POST /users
Content-Type: application/json
Authorization: Bearer {token}
X-Tenant-ID: acme

{
  "email": "joao.silva@exemplo.com",
  "firstName": "João",
  "lastName": "Silva",
  "attributes": {
    "department": "Engenharia",
    "location": "Lisboa"
  },
  "roles": ["desenvolvedor"]
}
```

Exemplo de resposta:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "joao.silva@exemplo.com",
  "firstName": "João",
  "lastName": "Silva",
  "status": "active",
  "createdAt": "2025-05-09T12:00:00Z",
  "updatedAt": "2025-05-09T12:00:00Z",
  "attributes": {
    "department": "Engenharia",
    "location": "Lisboa"
  },
  "roles": ["desenvolvedor"]
}
```

#### Gerenciamento de Papéis e Permissões

| Método | Endpoint | Descrição |
|--------|---------|-------------|
| `GET` | `/roles` | Listar papéis |
| `POST` | `/roles` | Criar um papel |
| `GET` | `/roles/{id}` | Obter um papel específico |
| `PUT` | `/roles/{id}` | Atualizar um papel |
| `DELETE` | `/roles/{id}` | Excluir um papel |
| `GET` | `/permissions` | Listar permissões |
| `POST` | `/roles/{id}/permissions` | Adicionar permissões a um papel |

### API de Autenticação

Endpoints para autenticação e gerenciamento de sessão.

#### Operações de Autenticação

| Método | Endpoint | Descrição |
|--------|---------|-------------|
| `POST` | `/auth/login` | Autenticar um usuário |
| `POST` | `/auth/logout` | Encerrar uma sessão |
| `POST` | `/auth/token` | Obter um novo token de acesso |
| `GET` | `/auth/status` | Verificar status de autenticação |
| `POST` | `/auth/mfa/initiate` | Iniciar processo de MFA |
| `POST` | `/auth/mfa/verify` | Verificar código MFA |

Exemplo de requisição de autenticação:
```
POST /auth/login
Content-Type: application/json
X-Tenant-ID: acme

{
  "username": "joao.silva@exemplo.com",
  "password": "senhaSegura123!",
  "factors": ["password"]
}
```

Exemplo de resposta:
```json
{
  "token": {
    "access_token": "eyJhbGciOiJSUzI1...",
    "refresh_token": "eyJhbGciOiJSUzI1...",
    "token_type": "Bearer",
    "expires_in": 3600
  },
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "joao.silva@exemplo.com",
    "firstName": "João",
    "lastName": "Silva"
  },
  "mfa_required": true,
  "mfa_options": ["totp", "sms"]
}
```

#### Autenticação AR/VR

| Método | Endpoint | Descrição |
|--------|---------|-------------|
| `POST` | `/auth/ar-vr/register` | Registrar autenticação espacial |
| `POST` | `/auth/ar-vr/authenticate` | Autenticar com padrões espaciais |
| `GET` | `/auth/ar-vr/methods` | Listar métodos disponíveis |

### API de Autorização

Endpoints para decisões de autorização e gerenciamento de políticas.

| Método | Endpoint | Descrição |
|--------|---------|-------------|
| `POST` | `/authz/check` | Verificar permissão para um recurso |
| `POST` | `/authz/bulk-check` | Verificar múltiplas permissões |
| `GET` | `/authz/policies` | Listar políticas de autorização |
| `POST` | `/authz/policies` | Criar política de autorização |

Exemplo de verificação de autorização:
```
POST /authz/check
Content-Type: application/json
Authorization: Bearer {token}
X-Tenant-ID: acme

{
  "principal": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "type": "USER"
  },
  "action": "READ",
  "resource": {
    "type": "REGISTRO_PACIENTE",
    "id": "123456"
  },
  "context": {
    "location": "hospital-ala-1",
    "emergency": false
  }
}
```

Exemplo de resposta:
```json
{
  "decision": "ALLOW",
  "policies_evaluated": [
    "politica-acesso-dados-saude",
    "politica-gdpr-saude"
  ],
  "reason": "Usuário tem papel de médico com relacionamento direto com o paciente"
}
```

### API de Compliance

Endpoints para validação e relatórios de conformidade.

| Método | Endpoint | Descrição |
|--------|---------|-------------|
| `POST` | `/compliance/validate` | Validar contra regras de conformidade |
| `GET` | `/compliance/reports` | Obter relatórios de conformidade |
| `GET` | `/compliance/regulations` | Listar regulamentações suportadas |

Exemplo de validação de conformidade:
```
POST /compliance/validate
Content-Type: application/json
Authorization: Bearer {token}
X-Tenant-ID: acme

{
  "regulation": "HIPAA",
  "context": {
    "data_type": "PHI",
    "operation": "TRANSFER",
    "recipient": {
      "type": "HEALTHCARE_PROVIDER",
      "id": "hospital-123"
    }
  }
}
```

Exemplo de resposta:
```json
{
  "compliant": true,
  "rules_evaluated": [
    "hipaa-minimo-necessario",
    "hipaa-divulgacao-autorizada"
  ],
  "validations": [
    {
      "rule": "hipaa-minimo-necessario",
      "status": "PASS",
      "details": "Transferência limitada aos elementos de dados necessários"
    },
    {
      "rule": "hipaa-divulgacao-autorizada",
      "status": "PASS",
      "details": "Destinatário é um provedor de saúde autorizado"
    }
  ]
}
```

## OAuth 2.1 / OpenID Connect

O módulo IAM implementa os padrões OAuth 2.1 e OpenID Connect 1.0 para autenticação federada.

### Endpoints OAuth

| Endpoint | Descrição |
|----------|-------------|
| `/oauth/authorize` | Endpoint de autorização |
| `/oauth/token` | Endpoint de token |
| `/oauth/revoke` | Revogação de token |
| `/oauth/introspect` | Introspecção de token |
| `/oauth/userinfo` | Informações do usuário |

### Descoberta OpenID Connect

Configuração OpenID está disponível em:
```
GET /.well-known/openid-configuration
```

## API GraphQL

Além da API REST, uma API GraphQL abrangente está disponível para operações complexas de dados.

Exemplo de consulta:
```graphql
query {
  user(id: "550e8400-e29b-41d4-a716-446655440000") {
    id
    email
    firstName
    lastName
    roles {
      name
      permissions {
        name
        description
      }
    }
    groups {
      name
      members {
        totalCount
      }
    }
    authenticationMethods {
      type
      isEnabled
      lastUsed
    }
  }
}
```

## Webhooks

O sistema IAM pode enviar notificações de eventos via webhooks:

1. **Registro**: Registre endpoints de webhook via API
2. **Tipos de Evento**: Eventos de segurança, eventos de ciclo de vida do usuário, alertas de conformidade
3. **Segurança**: Webhooks são assinados com HMAC para verificação

Exemplo de payload de webhook:
```json
{
  "event_type": "user.login_succeeded",
  "tenant_id": "acme",
  "timestamp": "2025-05-09T12:34:56Z",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "ip_address": "192.168.1.1",
    "user_agent": "Mozilla/5.0...",
    "authentication_method": "password",
    "mfa_used": true
  },
  "signature": "sha256=5d3997..."
}
```

## SDKs e Bibliotecas de Cliente

SDKs oficiais estão disponíveis para plataformas comuns:

- **JavaScript/TypeScript**: Para aplicações web
- **Python**: Para serviços backend
- **Java**: Para integrações empresariais
- **Swift/Kotlin**: Para aplicações móveis
- **C#/.NET**: Para ambientes Windows

## Versionamento de API

A API usa versionamento semântico:

- **Versionamento por Caminho**: `/v1/`, `/v2/` para versões principais
- **Versionamento por Cabeçalho**: `X-API-Version: 1.2` para versões menores
- **Avisos de Depreciação**: `X-API-Deprecated: true` com datas de desativação

## Considerações de Segurança

- Todos os endpoints da API usam TLS 1.3
- Payloads JSON são validados contra schemas
- Operações sensíveis requerem permissões elevadas
- Chaves de API têm escopos definidos para privilégio mínimo
- Limitação de taxa previne abuso
- Todas as chamadas de API são registradas para auditoria

## Conclusão

A API IAM fornece capacidades abrangentes de gerenciamento de identidade e acesso para a plataforma INNOVABIZ. Para definições detalhadas de schema, exemplos e testes interativos, consulte a documentação Swagger/OpenAPI disponível em: `https://{tenant-id}.docs.innovabiz.com/iam`.
