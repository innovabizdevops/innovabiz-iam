# API e Integração - IAM INNOVABIZ

**Documento:** Especificação de API e Integração do IAM
**Versão:** 1.0.0
**Data:** 2025-08-06
**Classificação:** Confidencial
**Autor:** Equipa INNOVABIZ DevSecOps

## 1. Visão Geral da API

O módulo IAM da plataforma INNOVABIZ expõe uma API completa e flexível para gestão de identidades e acessos, seguindo os princípios de API-first, segurança zero-trust e conformidade regulatória multi-regional. Este documento especifica as interfaces, protocolos e padrões de integração para consumidores internos e externos.

### 1.1 Princípios da API

- **API-First**: Design centrado na API antes da implementação
- **Padronização**: Interfaces consistentes em toda a plataforma
- **Versionamento Semântico**: Evolução não disruptiva das APIs
- **Segurança em Camadas**: Múltiplos níveis de proteção
- **Observabilidade Total**: Monitoramento completo de utilização e performance
- **Documentação como Código**: Especificações OpenAPI/GraphQL atualizadas automaticamente
- **Conformidade por Design**: Adaptação automática aos requisitos regulatórios regionais

## 2. GraphQL API (Principal)

### 2.1 Estrutura do Schema GraphQL

A API GraphQL é a interface principal do IAM, fornecendo acesso completo a todas as funcionalidades de gestão de identidade e acesso.

#### 2.1.1 Queries Principais

```graphql
type Query {
  # Usuários
  user(id: ID!): User @auth(requires: ["IAM:ReadUser"])
  users(filter: UserFilter, pagination: PaginationInput): UserConnection! @auth(requires: ["IAM:ReadUser"])
  
  # Grupos
  group(id: ID!): Group @auth(requires: ["IAM:ReadGroup"])
  groups(filter: GroupFilter, pagination: PaginationInput): GroupConnection! @auth(requires: ["IAM:ReadGroup"])
  
  # Hierarquia de Grupos
  groupHierarchy(rootGroupId: ID!): GroupHierarchy @auth(requires: ["IAM:ReadGroup"])
  
  # Roles e Permissões
  role(id: ID!): Role @auth(requires: ["IAM:ReadRole"])
  roles(filter: RoleFilter, pagination: PaginationInput): RoleConnection! @auth(requires: ["IAM:ReadRole"])
  permission(id: ID!): Permission @auth(requires: ["IAM:ReadPermission"])
  
  # Tenant
  tenant(id: ID!): Tenant @auth(requires: ["IAM:ReadTenant"])
  tenants(filter: TenantFilter, pagination: PaginationInput): TenantConnection! @auth(requires: ["IAM:ReadTenant"])
  
  # Verificações de Acesso
  checkPermission(permission: String!, resourceId: ID): Boolean! @auth
  checkRole(role: String!): Boolean! @auth
  
  # Estatísticas e Métricas
  userStatistics(tenantId: ID): UserStatistics! @auth(requires: ["IAM:ReadStatistics"])
  groupStatistics(groupId: ID): GroupStatistics! @auth(requires: ["IAM:ReadStatistics"])
}
```

#### 2.1.2 Mutations Principais

```graphql
type Mutation {
  # Autenticação
  login(input: LoginInput!): AuthPayload!
  refreshToken(input: RefreshTokenInput!): AuthPayload!
  logout: Boolean!
  
  # Usuários
  createUser(input: CreateUserInput!): User! @auth(requires: ["IAM:CreateUser"]) @validateInput
  updateUser(id: ID!, input: UpdateUserInput!): User! @auth(requires: ["IAM:UpdateUser"]) @validateInput
  deleteUser(id: ID!): Boolean! @auth(requires: ["IAM:DeleteUser"])
  
  # Grupos
  createGroup(input: CreateGroupInput!): Group! @auth(requires: ["IAM:CreateGroup"]) @validateInput
  updateGroup(id: ID!, input: UpdateGroupInput!): Group! @auth(requires: ["IAM:UpdateGroup"]) @validateInput
  deleteGroup(id: ID!): Boolean! @auth(requires: ["IAM:DeleteGroup"])
  
  # Membros de Grupo
  addGroupMember(groupId: ID!, userId: ID!, role: GroupMemberRole): GroupMembership! @auth(requires: ["IAM:UpdateGroup"])
  removeGroupMember(groupId: ID!, userId: ID!): Boolean! @auth(requires: ["IAM:UpdateGroup"])
  updateGroupMemberRole(groupId: ID!, userId: ID!, role: GroupMemberRole!): GroupMembership! @auth(requires: ["IAM:UpdateGroup"])
  
  # Hierarquia de Grupos
  addGroupToParent(childId: ID!, parentId: ID!): Boolean! @auth(requires: ["IAM:UpdateGroup"])
  removeGroupFromParent(childId: ID!, parentId: ID!): Boolean! @auth(requires: ["IAM:UpdateGroup"])
  
  # Roles e Permissões
  assignRoleToUser(userId: ID!, roleId: ID!): Boolean! @auth(requires: ["IAM:AssignRole"])
  revokeRoleFromUser(userId: ID!, roleId: ID!): Boolean! @auth(requires: ["IAM:RevokeRole"])
  assignRoleToGroup(groupId: ID!, roleId: ID!): Boolean! @auth(requires: ["IAM:AssignRole"])
  revokeRoleFromGroup(groupId: ID!, roleId: ID!): Boolean! @auth(requires: ["IAM:RevokeRole"])
}
```

#### 2.1.3 Subscriptions

```graphql
type Subscription {
  userUpdated(id: ID): User! @auth(requires: ["IAM:ReadUser"])
  groupUpdated(id: ID): Group! @auth(requires: ["IAM:ReadGroup"])
  groupMembershipChanged(groupId: ID!): GroupMembershipEvent! @auth(requires: ["IAM:ReadGroup"])
  securityEvent(filter: SecurityEventFilter): SecurityEvent! @auth(requires: ["IAM:ReadSecurityEvents"])
}
```

### 2.2 Diretivas GraphQL

```graphql
# Controle de Autenticação e Autorização
directive @auth(
  requires: [String!] = [], 
  allowSameUser: Boolean = false,
  checkTenant: Boolean = true
) on FIELD_DEFINITION

# Validação de Input
directive @validateInput on FIELD_DEFINITION

# Marcação de Campos Sensíveis
directive @sensitive on FIELD_DEFINITION | OBJECT

# Redação de Dados em Logs
directive @redact(replacement: String = "[REDACTED]") on FIELD_DEFINITION
```

### 2.3 Tipos Principais

Os tipos completos estão definidos no schema GraphQL. Aqui estão os principais:

```graphql
type User {
  id: ID!
  username: String!
  email: String! @redact
  firstName: String
  lastName: String
  status: UserStatus!
  tenantId: ID!
  createdAt: DateTime!
  updatedAt: DateTime
  lastLogin: DateTime
  groups: [GroupMembership!]!
  roles: [Role!]!
  permissions: [Permission!]!
  attributes: JSONObject
  securityProfile: UserSecurityProfile @auth(requires: ["IAM:ReadSecurity"])
}

type Group {
  id: ID!
  name: String!
  code: String!
  description: String
  tenantId: ID!
  createdAt: DateTime!
  updatedAt: DateTime
  members: [GroupMembership!]!
  parentGroups: [Group!]!
  childGroups: [Group!]!
  roles: [Role!]!
  permissions: [Permission!]!
  attributes: JSONObject
}

type Role {
  id: ID!
  name: String!
  code: String!
  description: String
  tenantId: ID!
  createdAt: DateTime!
  permissions: [Permission!]!
}

type Permission {
  id: ID!
  name: String!
  code: String!
  description: String
  resource: String!
  action: String!
}

type AuthPayload {
  user: User!
  accessToken: String!
  refreshToken: String!
  expiresIn: Int!
  tokenType: String!
}
```

## 3. REST API (Complementar)

Além da API GraphQL principal, o IAM fornece APIs REST para casos específicos, como integração com sistemas legados e operações simples de autenticação.

### 3.1 Endpoints de Autenticação

| Método | Endpoint | Descrição | Parâmetros |
|--------|----------|-----------|------------|
| POST | `/api/v1/auth/token` | Obter token de acesso | `username`, `password`, `grant_type` |
| POST | `/api/v1/auth/refresh` | Renovar token expirado | `refresh_token` |
| POST | `/api/v1/auth/logout` | Invalidar sessão | `token` |
| GET | `/api/v1/auth/userinfo` | Obter informações do usuário | - |
| POST | `/api/v1/auth/mfa/initiate` | Iniciar autenticação MFA | `mfa_type` |
| POST | `/api/v1/auth/mfa/verify` | Verificar código MFA | `code`, `transaction_id` |

### 3.2 Endpoints de Usuário

| Método | Endpoint | Descrição | Parâmetros |
|--------|----------|-----------|------------|
| GET | `/api/v1/users` | Listar usuários | `page`, `limit`, `filter` |
| POST | `/api/v1/users` | Criar usuário | `user_data` |
| GET | `/api/v1/users/{id}` | Obter usuário | - |
| PUT | `/api/v1/users/{id}` | Atualizar usuário | `user_data` |
| DELETE | `/api/v1/users/{id}` | Excluir usuário | - |
| GET | `/api/v1/users/{id}/roles` | Listar roles do usuário | - |
| POST | `/api/v1/users/{id}/roles` | Atribuir role | `role_id` |
| DELETE | `/api/v1/users/{id}/roles/{role_id}` | Revogar role | - |

### 3.3 Endpoints de Verificação

| Método | Endpoint | Descrição | Parâmetros |
|--------|----------|-----------|------------|
| POST | `/api/v1/auth/check` | Verificar permissões | `permissions`, `resource_id` |
| GET | `/api/v1/auth/introspect` | Introspecção de token | `token` |
| POST | `/api/v1/auth/batch-check` | Verificar múltiplas permissões | `checks` |

## 4. Protocolos de Autenticação

O IAM implementa múltiplos protocolos de autenticação para atender diferentes cenários de integração:

### 4.1 OAuth 2.1 + OpenID Connect

* **Authorization Code Flow**: Para aplicações web tradicionais
* **PKCE**: Para aplicações móveis e SPAs
* **Client Credentials**: Para comunicação entre serviços
* **Resource Owner Password**: Para aplicações legacy (uso limitado)

#### Endpoints OAuth 2.1 + OIDC

| Endpoint | Descrição |
|----------|-----------|
| `/oauth2/authorize` | Endpoint de autorização |
| `/oauth2/token` | Endpoint de token |
| `/oauth2/userinfo` | Endpoint de informações do usuário |
| `/.well-known/openid-configuration` | Configuração OpenID Connect |
| `/.well-known/jwks.json` | JSON Web Key Set |
| `/oauth2/revoke` | Revogação de token |

### 4.2 SAML 2.0

Suporte a federação de identidade via SAML 2.0 para integração com sistemas empresariais.

| Endpoint | Descrição |
|----------|-----------|
| `/saml/metadata` | Metadata do Provedor de Serviço |
| `/saml/sso` | Endpoint de Single Sign-On |
| `/saml/slo` | Endpoint de Single Logout |
| `/saml/acs` | Assertion Consumer Service |

### 4.3 WebAuthn / FIDO2

Suporte a autenticação sem senha via WebAuthn para maior segurança.

| Endpoint | Descrição |
|----------|-----------|
| `/api/v1/auth/webauthn/register/begin` | Iniciar registro de credencial |
| `/api/v1/auth/webauthn/register/complete` | Completar registro de credencial |
| `/api/v1/auth/webauthn/authenticate/begin` | Iniciar autenticação |
| `/api/v1/auth/webauthn/authenticate/complete` | Completar autenticação |

## 5. Integração com Event Bus

O IAM publica e consome eventos através de uma arquitetura orientada a eventos para integração assíncrona.

### 5.1 Eventos Publicados

| Tópico | Evento | Descrição | Schema |
|--------|--------|-----------|--------|
| `iam.user` | `user.created` | Usuário criado | `UserCreatedEvent` |
| `iam.user` | `user.updated` | Usuário atualizado | `UserUpdatedEvent` |
| `iam.user` | `user.deleted` | Usuário excluído | `UserDeletedEvent` |
| `iam.user` | `user.activated` | Usuário ativado | `UserStatusEvent` |
| `iam.user` | `user.deactivated` | Usuário desativado | `UserStatusEvent` |
| `iam.auth` | `auth.login` | Login bem-sucedido | `AuthLoginEvent` |
| `iam.auth` | `auth.logout` | Logout explícito | `AuthLogoutEvent` |
| `iam.auth` | `auth.failed_login` | Tentativa de login falhou | `AuthFailedEvent` |
| `iam.group` | `group.created` | Grupo criado | `GroupCreatedEvent` |
| `iam.group` | `group.updated` | Grupo atualizado | `GroupUpdatedEvent` |
| `iam.group` | `group.deleted` | Grupo excluído | `GroupDeletedEvent` |
| `iam.group` | `group.member_added` | Membro adicionado ao grupo | `GroupMembershipEvent` |
| `iam.group` | `group.member_removed` | Membro removido do grupo | `GroupMembershipEvent` |
| `iam.role` | `role.assigned` | Role atribuído | `RoleAssignmentEvent` |
| `iam.role` | `role.revoked` | Role revogado | `RoleRevocationEvent` |
| `iam.security` | `security.suspicious_activity` | Atividade suspeita detectada | `SuspiciousActivityEvent` |

### 5.2 Eventos Consumidos

| Tópico | Evento | Descrição | Ação |
|--------|--------|-----------|------|
| `payment.user` | `payment.user.status_changed` | Status de pagamento alterado | Atualização de atributos de usuário |
| `risk.user` | `risk.user.score_changed` | Score de risco alterado | Ajuste de políticas de segurança |
| `tenant.management` | `tenant.created` | Novo tenant criado | Provisionamento automático de IAM |
| `tenant.management` | `tenant.deleted` | Tenant removido | Limpeza de dados IAM |

## 6. Padrões de Integração

### 6.1 Autenticação de API

Todas as APIs são protegidas por múltiplos mecanismos de autenticação:

* **OAuth 2.1 Bearer Tokens**: JWT assinados com RS256
* **mTLS**: Para comunicação entre serviços internos
* **API Keys**: Para integrações simples (acesso limitado)

#### Header de Autenticação:

```
Authorization: Bearer [JWT_TOKEN]
```

### 6.2 Contexto Multi-Tenant

Todas as operações requerem identificação de tenant:

* Via token JWT (claim `tenant_id`)
* Via header HTTP para APIs REST:

```
X-Tenant-ID: [TENANT_ID]
```

### 6.3 Versionamento de API

* **GraphQL**: Evolução não disruptiva sem versionamento explícito
* **REST**: Versionamento no caminho (`/api/v1/...`)
* **Eventos**: Versionamento no schema (`UserCreatedEventV1`)

### 6.4 Limitação de Taxa (Rate Limiting)

* Baseado em token JWT e IP
* Limites configuráveis por tenant
* Headers de resposta para informar status:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1628607038
```

## 7. Segurança da API

### 7.1 Proteções Implementadas

* **JWT com Expiração Curta**: Tokens de acesso válidos por 15 minutos
* **CORS Configurável**: Política restritiva por padrão
* **Proteção CSRF**: Tokens anti-CSRF para fluxos baseados em cookie
* **Sanitização de Input**: Validação de todos os inputs
* **Rate Limiting**: Proteção contra abuso e brute force
* **WAF**: Integração com WAF para proteção adicional

### 7.2 Headers de Segurança

```
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
Content-Security-Policy: default-src 'self'
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
```

## 8. SDKs e Clientes

O IAM fornece SDKs oficiais para facilitar a integração:

### 8.1 SDK Frontend

* **@innovabiz/iam-react**: Hooks e componentes React para autenticação
* **@innovabiz/iam-vue**: Componentes Vue.js para autenticação
* **@innovabiz/iam-js**: SDK JavaScript vanilla

### 8.2 SDK Backend

* **@innovabiz/iam-node**: SDK Node.js/TypeScript
* **innovabiz-iam-python**: SDK Python
* **innovabiz-iam-go**: SDK Go
* **innovabiz-iam-java**: SDK Java
* **InnovaBiz.IAM**: SDK .NET

## 9. Documentação da API

### 9.1 Documentação Interativa

* GraphQL: GraphiQL disponível em `/graphql/explorer`
* REST: Swagger UI disponível em `/api/docs`

### 9.2 Especificações

* GraphQL: Schema disponível em `/graphql/schema.graphql`
* REST: OpenAPI 3.1 disponível em `/api/openapi.json`

## 10. Ambientes

| Ambiente | URL Base | Propósito |
|----------|----------|-----------|
| Desenvolvimento | `https://iam-dev.innovabiz.io` | Desenvolvimento e testes de integração |
| QA | `https://iam-qa.innovabiz.io` | Testes de qualidade |
| Staging | `https://iam-staging.innovabiz.io` | Validação pré-produção |
| Produção | `https://iam.innovabiz.io` | Produção |

### 10.1 Ambientes Regionais de Produção

| Região | URL Base |
|--------|----------|
| Angola (Luanda) | `https://ao.iam.innovabiz.io` |
| África (Joanesburgo) | `https://za.iam.innovabiz.io` |
| Brasil (São Paulo) | `https://br.iam.innovabiz.io` |
| Portugal (Lisboa) | `https://pt.iam.innovabiz.io` |
| UE (Frankfurt) | `https://eu.iam.innovabiz.io` |
| EUA (Nova York) | `https://us.iam.innovabiz.io` |
| China (Pequim) | `https://cn.iam.innovabiz.io` |
| Ásia (Singapura) | `https://sg.iam.innovabiz.io` |

## 11. Referências

* [OAuth 2.1 Specification](https://oauth.net/2.1/)
* [OpenID Connect Core 1.0](https://openid.net/specs/openid-connect-core-1_0.html)
* [GraphQL Specification](https://spec.graphql.org/October2021/)
* [JWT RFC 7519](https://datatracker.ietf.org/doc/html/rfc7519)
* [SAML 2.0 Specification](http://docs.oasis-open.org/security/saml/v2.0/saml-core-2.0-os.pdf)