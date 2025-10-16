package innovabiz.iam.authz

# Política de autorização para o IAM INNOVABIZ
# Implementa controle de acesso baseado em atributos (ABAC) seguindo:
# - ISO/IEC 27001 (Segurança da Informação)
# - COBIT 2019 (Governança de TI)
# - TOGAF 10.0 (Arquitetura Empresarial)
# - PCI DSS (Segurança de Dados de Pagamento)
# - NIST Cybersecurity Framework

import future.keywords.in
import future.keywords.every
import future.keywords.if
import future.keywords.contains

# Regra principal de decisão
default allow = false

# Super administrador tem acesso completo
allow {
    "system_admin" in input.user.roles
}

# Administrador do tenant tem acesso às operações dentro de seu próprio tenant
allow {
    "tenant_admin" in input.user.roles
    input.tenant.id == tenant_id_from_path
    not is_system_critical_operation
}

# Permissões baseadas em papéis e recursos
allow {
    # Verifica se o usuário tem as permissões necessárias
    has_permission(input.user, required_permission)
    
    # Verifica se a operação está dentro do escopo do tenant
    tenant_scope_valid
    
    # Verifica limites adicionais de segurança
    not is_restricted_operation
}

# Determina a permissão necessária com base no caminho e método
required_permission = permission {
    # Mapeamento de rotas para permissões
    route_permission = route_permissions[_]
    route_permission.method = input.request.method
    regex.match(route_permission.path_pattern, input.request.path)
    permission = route_permission.permission
}

# Tabela de mapeamento de rotas para permissões
route_permissions = [
    # Operações CRUD de funções (roles)
    {"method": "POST", "path_pattern": "/api/v1/roles", "permission": "roles:create"},
    {"method": "GET", "path_pattern": "/api/v1/roles/[^/]+", "permission": "roles:read"},
    {"method": "GET", "path_pattern": "/api/v1/roles", "permission": "roles:list"},
    {"method": "PUT", "path_pattern": "/api/v1/roles/[^/]+", "permission": "roles:update"},
    {"method": "DELETE", "path_pattern": "/api/v1/roles/[^/]+", "permission": "roles:delete"},
    {"method": "DELETE", "path_pattern": "/api/v1/roles/[^/]+/hard", "permission": "roles:hard_delete"},
    
    # Operações de clonagem e sincronização de funções
    {"method": "POST", "path_pattern": "/api/v1/roles/[^/]+/clone", "permission": "roles:clone"},
    {"method": "POST", "path_pattern": "/api/v1/roles/sync", "permission": "roles:sync_system"},
    
    # Gerenciamento de permissões em funções
    {"method": "GET", "path_pattern": "/api/v1/roles/[^/]+/permissions", "permission": "roles:list_permissions"},
    {"method": "GET", "path_pattern": "/api/v1/roles/[^/]+/permissions/all", "permission": "roles:list_permissions"},
    {"method": "POST", "path_pattern": "/api/v1/roles/[^/]+/permissions", "permission": "roles:assign_permission"},
    {"method": "DELETE", "path_pattern": "/api/v1/roles/[^/]+/permissions/[^/]+", "permission": "roles:revoke_permission"},
    {"method": "GET", "path_pattern": "/api/v1/roles/[^/]+/permissions/[^/]+", "permission": "roles:check_permission"},
    
    # Gerenciamento de hierarquia de funções
    {"method": "GET", "path_pattern": "/api/v1/roles/[^/]+/children", "permission": "roles:list_hierarchy"},
    {"method": "GET", "path_pattern": "/api/v1/roles/[^/]+/parents", "permission": "roles:list_hierarchy"},
    {"method": "GET", "path_pattern": "/api/v1/roles/[^/]+/descendants", "permission": "roles:list_hierarchy"},
    {"method": "GET", "path_pattern": "/api/v1/roles/[^/]+/ancestors", "permission": "roles:list_hierarchy"},
    {"method": "POST", "path_pattern": "/api/v1/roles/[^/]+/children", "permission": "roles:modify_hierarchy"},
    {"method": "DELETE", "path_pattern": "/api/v1/roles/[^/]+/children/[^/]+", "permission": "roles:modify_hierarchy"},
    
    # Gerenciamento de usuários em funções
    {"method": "GET", "path_pattern": "/api/v1/roles/[^/]+/users", "permission": "roles:list_users"},
    {"method": "POST", "path_pattern": "/api/v1/roles/[^/]+/users", "permission": "roles:assign_user"},
    {"method": "PUT", "path_pattern": "/api/v1/roles/[^/]+/users/[^/]+/expiration", "permission": "roles:update_user_expiration"},
    {"method": "DELETE", "path_pattern": "/api/v1/roles/[^/]+/users/[^/]+", "permission": "roles:remove_user"},
    {"method": "GET", "path_pattern": "/api/v1/roles/[^/]+/users/[^/]+", "permission": "roles:check_user"}
]

# Verifica se o usuário possui a permissão necessária
has_permission(user, permission) {
    # O usuário tem a permissão diretamente em seus papéis
    some role in user.roles
    some perm in role_permissions[role]
    perm = permission
}

# Se o sistema de permissões não estiver configurado, a verificação é ignorada
has_permission(_, _) {
    not role_permissions
}

# Definição temporária de permissões para funções (em produção, isso seria buscado de uma fonte dinâmica)
role_permissions = {
    "system_admin": ["*:*"],
    "tenant_admin": [
        "roles:create", "roles:read", "roles:list", "roles:update", "roles:delete",
        "roles:clone", "roles:list_permissions", "roles:assign_permission", "roles:revoke_permission",
        "roles:check_permission", "roles:list_hierarchy", "roles:modify_hierarchy",
        "roles:list_users", "roles:assign_user", "roles:update_user_expiration", "roles:remove_user",
        "roles:check_user"
    ],
    "role_manager": [
        "roles:read", "roles:list", "roles:update", 
        "roles:list_permissions", "roles:assign_permission", "roles:revoke_permission",
        "roles:list_hierarchy", "roles:list_users", "roles:assign_user", "roles:remove_user"
    ],
    "role_viewer": [
        "roles:read", "roles:list", "roles:list_permissions", "roles:check_permission",
        "roles:list_hierarchy", "roles:list_users", "roles:check_user"
    ]
}

# Extrai o tenant_id do caminho, se disponível
tenant_id_from_path = tenant_id {
    regex.match("/api/v1/tenants/([^/]+)/", input.request.path)
    captures := regex.find_all_string_submatch_n("/api/v1/tenants/([^/]+)/", input.request.path, 1)
    tenant_id := captures[0][1]
}

# Se não houver tenant_id no caminho, usar o tenant_id do contexto
tenant_id_from_path = input.tenant.id {
    not regex.match("/api/v1/tenants/([^/]+)/", input.request.path)
}

# Verifica se a operação é válida dentro do escopo do tenant
tenant_scope_valid {
    # Se a operação incluir um ID de tenant no caminho, deve coincidir com o tenant do usuário
    tenant_id_from_path
    tenant_id_from_path = input.tenant.id
}

tenant_scope_valid {
    # Se não houver tenant_id no caminho, permitir (gerenciado por outras regras)
    not tenant_id_from_path
}

# Verifica se é uma operação crítica do sistema
is_system_critical_operation {
    critical_operations = {
        "POST:/api/v1/roles/sync",
        "DELETE:/api/v1/roles/system_",
        "PUT:/api/v1/roles/system_"
    }
    
    operation = concat(":", [input.request.method, input.request.path])
    startswith_match(critical_operations, operation)
}

# Verifica se é uma operação restrita com controles adicionais
is_restricted_operation {
    # Hard delete é uma operação restrita
    input.request.method == "DELETE"
    endswith(input.request.path, "/hard")
    
    # Verificar controles adicionais aqui (por exemplo, aprovações, janelas de manutenção)
    not has_permission(input.user, "system:maintenance")
}

# Função auxiliar para verificar se uma string começa com algum elemento do conjunto
startswith_match(set, str) {
    some item in set
    startswith(str, item)
}

# Função auxiliar para verificar se uma string termina com algum elemento do conjunto
endswith_match(set, str) {
    some item in set
    endswith(str, item)
}