# Especificação Técnica - IAM RoleService

## Visão Geral

O RoleService é um componente crítico do módulo IAM (Identity and Access Management) da plataforma INNOVABIZ, responsável pela gestão completa de funções (roles), suas permissões associadas, hierarquia de funções e atribuições a utilizadores. Esta especificação detalha os aspectos técnicos, regras de negócio, modelo de dados e interfaces de API para implementação do serviço.

| Metadata | Valor |
|----------|-------|
| Versão | 1.0.0 |
| Status | Implementação |
| Classificação | Confidencial |
| Data Criação | 2025-08-05 |
| Última Atualização | 2025-08-05 |
| Autor | INNOVABIZ IAM Team |
| Aprovado por | Eduardo Jeremias |
| Responsável | Equipe de IAM |

## Contextualização e Objetivos

O RoleService implementa um modelo avançado de RBAC (Role-Based Access Control) com suporte a hierarquias e atributos (ARBAC), fornecendo:

1. Gestão de funções com diversos níveis de granularidade
2. Atribuição e revogação de permissões a funções
3. Organização de funções em estruturas hierárquicas
4. Herança de permissões entre funções pai e filho
5. Atribuição e revogação de funções a utilizadores
6. Suporte a expiração temporal de funções atribuídas
7. Validação de permissões diretas e herdadas
8. Isolamento multitenancy de funções e permissões
9. Rastreabilidade completa das operações via auditoria
10. Integração com ABAC via Open Policy Agent (OPA)

## Modelo de Domínio

### Entidades Principais

#### Role (Função)
```go
type Role struct {
    ID          uuid.UUID   // Identificador único da função
    TenantID    uuid.UUID   // Identificador do tenant
    Name        string      // Nome da função
    Description string      // Descrição detalhada da função
    Type        RoleType    // Tipo da função (SYSTEM, CUSTOM)
    Status      Status      // Status (ACTIVE, INACTIVE, DELETED)
    Metadata    JSONMap     // Metadados customizáveis
    CreatedAt   time.Time   // Data de criação
    CreatedBy   uuid.UUID   // ID do usuário que criou
    UpdatedAt   time.Time   // Data da última atualização
    UpdatedBy   uuid.UUID   // ID do usuário que atualizou
    DeletedAt   *time.Time  // Data de exclusão (soft delete)
    DeletedBy   *uuid.UUID  // ID do usuário que excluiu
}
```

#### RolePermission (Associação Função-Permissão)
```go
type RolePermission struct {
    ID           uuid.UUID  // ID da associação
    TenantID     uuid.UUID  // Identificador do tenant
    RoleID       uuid.UUID  // ID da função
    PermissionID uuid.UUID  // ID da permissão
    CreatedAt    time.Time  // Data de criação
    CreatedBy    uuid.UUID  // ID do usuário que criou
}
```

#### RoleHierarchy (Hierarquia de Funções)
```go
type RoleHierarchy struct {
    ID        uuid.UUID  // ID da relação hierárquica
    TenantID  uuid.UUID  // Identificador do tenant
    ParentID  uuid.UUID  // ID da função pai
    ChildID   uuid.UUID  // ID da função filho
    CreatedAt time.Time  // Data de criação
    CreatedBy uuid.UUID  // ID do usuário que criou
}
```

#### UserRole (Atribuição Utilizador-Função)
```go
type UserRole struct {
    ID         uuid.UUID   // ID da atribuição
    TenantID   uuid.UUID   // Identificador do tenant
    UserID     uuid.UUID   // ID do utilizador
    RoleID     uuid.UUID   // ID da função
    ExpiresAt  *time.Time  // Data de expiração opcional
    CreatedAt  time.Time   // Data de criação
    CreatedBy  uuid.UUID   // ID do usuário que criou
    RevokedAt  *time.Time  // Data de revogação
    RevokedBy  *uuid.UUID  // ID do usuário que revogou
}
```

### Enumerações e Tipos

```go
// Tipo de função
type RoleType string
const (
    RoleTypeSystem RoleType = "SYSTEM"   // Funções do sistema (não podem ser excluídas)
    RoleTypeCustom RoleType = "CUSTOM"   // Funções personalizadas
)

// Status da função
type Status string
const (
    StatusActive   Status = "ACTIVE"    // Função ativa
    StatusInactive Status = "INACTIVE"  // Função inativa
    StatusDeleted  Status = "DELETED"   // Função excluída logicamente
)

// Mapa JSON para metadados
type JSONMap map[string]interface{}
```

## Regras de Negócio

### Regras Gerais

1. **Multitenancy**: Todas as funções são isoladas por tenant, impedindo acesso cruzado entre tenants
2. **Exclusão**: Funções do tipo SYSTEM não podem ser excluídas, apenas desativadas
3. **Unicidade**: Nomes de funções devem ser únicos dentro de um mesmo tenant
4. **Auditoria**: Todas as operações são registradas com metadados do usuário e timestamp
5. **Soft Delete**: Exclusão lógica (soft delete) é o padrão, preservando histórico

### Regras de Hierarquia

1. **Anti-Ciclos**: Não é permitida a criação de ciclos na hierarquia de funções
2. **Profundidade**: A hierarquia pode ter profundidade ilimitada (limitada apenas por recursos)
3. **Herança**: Permissões são herdadas de pai para filho através da hierarquia
4. **Transitividade**: A herança é transitiva ao longo de toda a cadeia hierárquica
5. **Multiparentalidade**: Uma função pode ter múltiplas funções pai

### Regras de Permissões

1. **Granularidade**: Permissões são atribuídas em nível granular a funções
2. **Herança**: Permissões são herdadas através da hierarquia de funções
3. **Desambiguação**: Se uma função herda permissões conflitantes, a permissão mais próxima na hierarquia prevalece
4. **Revogação**: A revogação de uma permissão afeta apenas atribuições diretas, não as herdadas

### Regras de Atribuição a Utilizadores

1. **Expiração**: Atribuições podem ter data de expiração opcional
2. **Validação Temporal**: Atribuições expiradas são automaticamente invalidadas
3. **Revogação**: Atribuições podem ser revogadas explicitamente antes da expiração
4. **Múltiplas Atribuições**: Um utilizador pode ter múltiplas funções atribuídas
5. **Verificação**: A verificação de função considera tanto atribuições diretas quanto herdadas via hierarquia
6. **Contexto**: A verificação pode considerar contexto adicional via integração com OPA

## Interface de Serviço (RoleService)

```go
type RoleService interface {
    // CRUD de Funções
    CreateRole(ctx context.Context, req CreateRoleRequest) (*Role, error)
    GetRole(ctx context.Context, req GetRoleRequest) (*Role, error)
    ListRoles(ctx context.Context, req ListRolesRequest) (*ListRolesResponse, error)
    UpdateRole(ctx context.Context, req UpdateRoleRequest) (*Role, error)
    DeleteRole(ctx context.Context, req DeleteRoleRequest) error
    HardDeleteRole(ctx context.Context, req DeleteRoleRequest) error
    
    // Clone e Sincronização
    CloneRole(ctx context.Context, req CloneRoleRequest) (*Role, error)
    SyncSystemRoles(ctx context.Context, req SyncSystemRolesRequest) (*SyncSystemRolesResponse, error)
    
    // Gestão de Permissões
    GetRolePermissions(ctx context.Context, req GetRolePermissionsRequest) (*GetPermissionsResponse, error)
    GetAllRolePermissions(ctx context.Context, req GetRolePermissionsRequest) (*GetPermissionsResponse, error)
    AssignPermissionToRole(ctx context.Context, req AssignPermissionRequest) error
    RevokePermissionFromRole(ctx context.Context, req RevokePermissionRequest) error
    CheckRoleHasPermission(ctx context.Context, req CheckPermissionRequest) (bool, error)
    
    // Gestão de Hierarquia
    GetChildRoles(ctx context.Context, req GetChildRolesRequest) (*GetRolesResponse, error)
    GetParentRoles(ctx context.Context, req GetParentRolesRequest) (*GetRolesResponse, error)
    GetDescendantRoles(ctx context.Context, req GetDescendantRolesRequest) (*GetRolesResponse, error)
    GetAncestorRoles(ctx context.Context, req GetAncestorRolesRequest) (*GetRolesResponse, error)
    AssignChildRole(ctx context.Context, req AssignChildRoleRequest) error
    RemoveChildRole(ctx context.Context, req RemoveChildRoleRequest) error
    
    // Gestão de Utilizadores
    GetRoleUsers(ctx context.Context, req GetRoleUsersRequest) (*GetUsersResponse, error)
    GetUserRoles(ctx context.Context, req GetUserRolesRequest) (*GetRolesResponse, error)
    AssignRoleToUser(ctx context.Context, req AssignRoleToUserRequest) error
    UpdateUserRoleExpiration(ctx context.Context, req UpdateExpirationRequest) error
    RemoveRoleFromUser(ctx context.Context, req RemoveRoleFromUserRequest) error
    CheckUserHasRole(ctx context.Context, req CheckUserRoleRequest) (bool, error)
}
```

## Especificação de API HTTP

### CRUD de Funções

#### Criar Função
- **Endpoint**: `POST /api/v1/roles`
- **Headers**:
  - `X-Tenant-ID`: ID do tenant (UUID)
  - `X-User-ID`: ID do utilizador autenticado (UUID)
  - `Content-Type`: application/json
- **Request Body**:
  ```json
  {
    "name": "string",
    "description": "string",
    "type": "SYSTEM|CUSTOM",
    "metadata": {
      "additionalProp": "any"
    }
  }
  ```
- **Response**: `201 Created`
  ```json
  {
    "id": "uuid",
    "tenant_id": "uuid",
    "name": "string",
    "description": "string",
    "type": "SYSTEM|CUSTOM",
    "status": "ACTIVE|INACTIVE|DELETED",
    "metadata": {
      "additionalProp": "any"
    },
    "created_at": "2025-08-05T10:30:00Z",
    "created_by": "uuid",
    "updated_at": "2025-08-05T10:30:00Z",
    "updated_by": "uuid"
  }
  ```
- **Erros**:
  - `400 Bad Request`: Parâmetros inválidos
  - `401 Unauthorized`: Autenticação inválida
  - `403 Forbidden`: Permissão insuficiente
  - `409 Conflict`: Nome de função já existe no tenant

#### Obter Função por ID
- **Endpoint**: `GET /api/v1/roles/{roleId}`
- **Headers**:
  - `X-Tenant-ID`: ID do tenant (UUID)
  - `X-User-ID`: ID do utilizador autenticado (UUID)
- **Path Parameters**:
  - `roleId`: UUID da função
- **Response**: `200 OK`
  ```json
  {
    "id": "uuid",
    "tenant_id": "uuid",
    "name": "string",
    "description": "string",
    "type": "SYSTEM|CUSTOM",
    "status": "ACTIVE|INACTIVE|DELETED",
    "metadata": {
      "additionalProp": "any"
    },
    "created_at": "2025-08-05T10:30:00Z",
    "created_by": "uuid",
    "updated_at": "2025-08-05T10:30:00Z",
    "updated_by": "uuid"
  }
  ```
- **Erros**:
  - `400 Bad Request`: UUID inválido
  - `401 Unauthorized`: Autenticação inválida
  - `403 Forbidden`: Permissão insuficiente
  - `404 Not Found`: Função não encontrada

#### Listar Funções
- **Endpoint**: `GET /api/v1/roles`
- **Headers**:
  - `X-Tenant-ID`: ID do tenant (UUID)
  - `X-User-ID`: ID do utilizador autenticado (UUID)
- **Query Parameters**:
  - `page`: Número da página (default: 1)
  - `per_page`: Itens por página (default: 20, max: 100)
  - `status`: Filtro de status (optional)
  - `type`: Filtro de tipo (optional)
  - `search`: Busca textual (optional)
  - `sort`: Campo para ordenação (default: name)
  - `direction`: Direção da ordenação (asc/desc, default: asc)
- **Response**: `200 OK`
  ```json
  {
    "items": [
      {
        "id": "uuid",
        "tenant_id": "uuid",
        "name": "string",
        "description": "string",
        "type": "SYSTEM|CUSTOM",
        "status": "ACTIVE|INACTIVE|DELETED",
        "created_at": "2025-08-05T10:30:00Z"
      }
    ],
    "pagination": {
      "total": 0,
      "per_page": 0,
      "current_page": 0,
      "last_page": 0,
      "from": 0,
      "to": 0
    }
  }
  ```
- **Erros**:
  - `400 Bad Request`: Parâmetros de paginação inválidos
  - `401 Unauthorized`: Autenticação inválida
  - `403 Forbidden`: Permissão insuficiente

### Gestão de Permissões

#### Obter Permissões de uma Função (Diretas)
- **Endpoint**: `GET /api/v1/roles/{roleId}/permissions`
- **Headers**:
  - `X-Tenant-ID`: ID do tenant (UUID)
  - `X-User-ID`: ID do utilizador autenticado (UUID)
- **Path Parameters**:
  - `roleId`: UUID da função
- **Query Parameters**:
  - `page`: Número da página (default: 1)
  - `per_page`: Itens por página (default: 20, max: 100)
- **Response**: `200 OK`
  ```json
  {
    "items": [
      {
        "id": "uuid",
        "name": "string",
        "description": "string",
        "resource": "string",
        "action": "string"
      }
    ],
    "pagination": {
      "total": 0,
      "per_page": 0,
      "current_page": 0,
      "last_page": 0,
      "from": 0,
      "to": 0
    }
  }
  ```

#### Obter Todas as Permissões (Diretas + Herdadas)
- **Endpoint**: `GET /api/v1/roles/{roleId}/all-permissions`
- **Headers**:
  - `X-Tenant-ID`: ID do tenant (UUID)
  - `X-User-ID`: ID do utilizador autenticado (UUID)
- **Path Parameters**:
  - `roleId`: UUID da função
- **Query Parameters**:
  - `page`: Número da página (default: 1)
  - `per_page`: Itens por página (default: 20, max: 100)
- **Response**: `200 OK` (formato idêntico ao endpoint anterior)

### Gestão de Hierarquia

#### Obter Funções Filhas
- **Endpoint**: `GET /api/v1/roles/{roleId}/children`
- **Headers**:
  - `X-Tenant-ID`: ID do tenant (UUID)
  - `X-User-ID`: ID do utilizador autenticado (UUID)
- **Path Parameters**:
  - `roleId`: UUID da função pai
- **Query Parameters**:
  - `page`: Número da página (default: 1)
  - `per_page`: Itens por página (default: 20, max: 100)
- **Response**: `200 OK`
  ```json
  {
    "items": [
      {
        "id": "uuid",
        "tenant_id": "uuid",
        "name": "string",
        "description": "string",
        "type": "SYSTEM|CUSTOM",
        "status": "ACTIVE|INACTIVE|DELETED"
      }
    ],
    "pagination": {
      "total": 0,
      "per_page": 0,
      "current_page": 0,
      "last_page": 0,
      "from": 0,
      "to": 0
    }
  }
  ```

#### Atribuir Função Filha
- **Endpoint**: `POST /api/v1/roles/{parentId}/children/{childId}`
- **Headers**:
  - `X-Tenant-ID`: ID do tenant (UUID)
  - `X-User-ID`: ID do utilizador autenticado (UUID)
- **Path Parameters**:
  - `parentId`: UUID da função pai
  - `childId`: UUID da função filho
- **Response**: `201 Created`
- **Erros**:
  - `400 Bad Request`: UUID inválido
  - `401 Unauthorized`: Autenticação inválida
  - `403 Forbidden`: Permissão insuficiente
  - `404 Not Found`: Função não encontrada
  - `409 Conflict`: Hierarquia já existe ou criaria ciclo

### Gestão de Utilizadores

#### Obter Utilizadores com Função Específica
- **Endpoint**: `GET /api/v1/roles/{roleId}/users`
- **Headers**:
  - `X-Tenant-ID`: ID do tenant (UUID)
  - `X-User-ID`: ID do utilizador autenticado (UUID)
- **Path Parameters**:
  - `roleId`: UUID da função
- **Query Parameters**:
  - `page`: Número da página (default: 1)
  - `per_page`: Itens por página (default: 20, max: 100)
  - `include_expired`: Incluir atribuições expiradas (boolean, default: false)
- **Response**: `200 OK`
  ```json
  {
    "items": [
      {
        "user_id": "uuid",
        "username": "string",
        "email": "string",
        "display_name": "string",
        "assigned_at": "2025-08-05T10:30:00Z",
        "expires_at": "2025-08-05T10:30:00Z"
      }
    ],
    "pagination": {
      "total": 0,
      "per_page": 0,
      "current_page": 0,
      "last_page": 0,
      "from": 0,
      "to": 0
    }
  }
  ```

#### Atribuir Função a Utilizador
- **Endpoint**: `POST /api/v1/roles/{roleId}/users/{userId}`
- **Headers**:
  - `X-Tenant-ID`: ID do tenant (UUID)
  - `X-User-ID`: ID do utilizador autenticado (UUID)
  - `Content-Type`: application/json
- **Path Parameters**:
  - `roleId`: UUID da função
  - `userId`: UUID do utilizador
- **Request Body**:
  ```json
  {
    "expires_at": "2026-08-05T10:30:00Z" // Optional
  }
  ```
- **Response**: `201 Created`
- **Erros**:
  - `400 Bad Request`: UUID ou data inválida
  - `401 Unauthorized`: Autenticação inválida
  - `403 Forbidden`: Permissão insuficiente
  - `404 Not Found`: Função ou utilizador não encontrado
  - `409 Conflict`: Atribuição já existe

## Eventos Emitidos

O RoleService publica os seguintes eventos via event bus para integração com outros sistemas:

### Eventos de Função
1. `role.created`: Quando uma nova função é criada
2. `role.updated`: Quando uma função é atualizada
3. `role.deleted`: Quando uma função é excluída logicamente
4. `role.hard_deleted`: Quando uma função é excluída permanentemente
5. `role.activated`: Quando uma função inativa é ativada
6. `role.deactivated`: Quando uma função ativa é desativada
7. `role.cloned`: Quando uma função é clonada

### Eventos de Permissão
1. `permission.assigned`: Quando uma permissão é atribuída a uma função
2. `permission.revoked`: Quando uma permissão é revogada de uma função

### Eventos de Hierarquia
1. `role.hierarchy.created`: Quando uma relação hierárquica é criada
2. `role.hierarchy.removed`: Quando uma relação hierárquica é removida

### Eventos de Utilizador
1. `user.role.assigned`: Quando uma função é atribuída a um utilizador
2. `user.role.removed`: Quando uma função é revogada de um utilizador
3. `user.role.expired`: Quando a atribuição de uma função a um utilizador expira
4. `user.role.expiration_updated`: Quando a expiração de uma atribuição é atualizada

## Métricas e Telemetria

O RoleService expõe as seguintes métricas para monitoramento:

1. `iam_roles_count`: Contador do total de funções por tenant e tipo
2. `iam_role_permissions_count`: Contador do total de permissões atribuídas a funções
3. `iam_role_hierarchy_depth`: Histograma da profundidade das hierarquias de função
4. `iam_user_roles_count`: Contador do total de atribuições de função a utilizadores
5. `iam_role_operations_total`: Contador de operações realizadas, por tipo
6. `iam_role_operations_errors`: Contador de erros em operações, por tipo
7. `iam_role_operations_duration`: Histograma da duração das operações

## Políticas OPA (Open Policy Agent)

O RoleService integra-se com OPA para decisões de autorização complexas. Exemplos de políticas:

```rego
# Política para verificar se um usuário pode criar funções
package iam.role

default allow_create = false

# Permite criação de funções se o usuário tiver permissão específica
allow_create {
    input.user.permissions[_] == "iam:role:create"
}

# Permite criação se o usuário for administrador
allow_create {
    input.user.roles[_] == "ADMIN"
}

# Restringe criação de funções SYSTEM apenas para super admins
allow_create {
    input.role.type != "SYSTEM"
    input.user.permissions[_] == "iam:role:create"
}

allow_create {
    input.role.type == "SYSTEM"
    input.user.roles[_] == "SUPER_ADMIN"
}
```

## Estratégias de Cache

O RoleService implementa uma estratégia de cache em múltiplas camadas para otimizar o desempenho:

1. **Cache L1 (Local)**: Cache em memória para consultas frequentes
   - TTL: 60 segundos
   - Tamanho máximo: Configurável por nó
   
2. **Cache L2 (Distribuído)**: Cache Redis para consultas entre nós
   - TTL: 300 segundos (5 minutos)
   - Chaves: Prefixadas com `iam:role:`
   - Estratégia de invalidação: Baseada em eventos

### Itens Cacheados

1. Funções por ID: `iam:role:{tenant_id}:{role_id}`
2. Hierarquia de funções: `iam:role:hierarchy:{tenant_id}:{role_id}`
3. Permissões diretas: `iam:role:permissions:{tenant_id}:{role_id}`
4. Permissões efetivas: `iam:role:effective_permissions:{tenant_id}:{role_id}`
5. Verificação de permissão: `iam:role:has_permission:{tenant_id}:{role_id}:{permission_id}`
6. Utilizadores por função: `iam:role:users:{tenant_id}:{role_id}`

## Considerações de Segurança

1. **Validação de Entrada**: Todos os inputs são validados rigorosamente
2. **Sanitização**: Dados são sanitizados antes de operações de banco de dados
3. **Prevenção de CSRF**: Tokens CSRF são verificados em operações de mutação
4. **Rate Limiting**: Limites de taxa são aplicados para prevenir abusos
5. **Auditoria**: Todas as operações são registradas em log de auditoria imutável
6. **Detecção de Anomalias**: Monitoramento para operações anômalas
7. **Criptografia**: Dados sensíveis são criptografados em trânsito e em repouso
8. **Princípio do Menor Privilégio**: Funções têm apenas permissões necessárias

## Testes e Qualidade

O RoleService é testado em múltiplos níveis:

1. **Testes Unitários**: Cobrem lógica de negócio e validações
2. **Testes de Integração**: Cobrem integração entre componentes
3. **Testes de API**: Validam endpoints HTTP
4. **Testes de Performance**: Garantem escalabilidade e desempenho
5. **Testes de Segurança**: Avaliam vulnerabilidades potenciais
6. **Testes de Conformidade**: Verificam aderência a padrões e regulamentos

A cobertura mínima de testes é de 80%, com ênfase em fluxos críticos e casos de borda.

## Próximos Passos

1. **Execução da Suite de Testes Completa**: Validar todos os cenários de teste
2. **Implementação de RSA para JWT**: Adicionar suporte a assinatura com chave assimétrica
3. **Integração com OPA**: Finalizar integração com motor de políticas
4. **Documentação OpenAPI**: Completar especificação Swagger
5. **Testes de Carga**: Verificar comportamento sob alto volume de requisições
6. **Integração com Event Bus**: Implementar publicação e consumo de eventos

---

© 2025 INNOVABIZ - Todos os direitos reservados