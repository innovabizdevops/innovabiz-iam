# role_permissions.rego
# Políticas de autorização para gestão de permissões em funções
# Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53, PCI DSS v4.0
package innovabiz.iam.role.permissions

import data.innovabiz.iam.role.base
import future.keywords

# ==========================================
# === Atribuição de Permissão a uma Função ===
# ==========================================

# Decisão final para atribuição de permissão
permission_assignment_decision := {
    "allow": allow,
    "reason": reason
} {
    # Computar decisão de autorização
    allow := permission_assignment_allowed
    reason := permission_assignment_reason
}

# Verifica se o usuário tem permissão para atribuir uma permissão a uma função
permission_assignment_allowed {
    # Verificação de autenticação
    base.user_authenticated

    # Verificação de escopo de tenant
    base.tenant_scope_valid
    
    # Verificação de integridade dos dados
    base.valid_uuid(input.resource.role_id)
    base.valid_uuid(input.resource.permission_id)
    
    # Acesso especial para super admin
    base.is_super_admin
} else {
    # Verificação de autenticação
    base.user_authenticated

    # Verificação de escopo de tenant
    base.tenant_scope_valid
    
    # Verificação de integridade dos dados
    base.valid_uuid(input.resource.role_id)
    base.valid_uuid(input.resource.permission_id)
    
    # Usuário tem permissão explícita para atribuir permissões a funções
    base.has_permission("role:assign_permission")
    
    # A função e a permissão existem
    role_exists
    permission_exists
    
    # Usuário está no mesmo tenant da função
    same_tenant_as_role
    
    # Permissão não é crítica (somente super admin pode atribuir permissões críticas)
    not is_critical_permission
}

# Recupera a razão pela qual a atribuição de permissão foi permitida ou negada
permission_assignment_reason = reason {
    not permission_assignment_allowed
    reason := concat(" ", [
        get_reason_auth_failed,
        get_reason_tenant_mismatch,
        get_reason_role_not_exists,
        get_reason_permission_not_exists,
        get_reason_critical_permission
    ])
} else {
    permission_assignment_allowed
    reason := "atribuição de permissão autorizada"
}

# === Funções auxiliares para atribuição de permissão ===

# Verifica se a função existe
role_exists {
    role_id := input.resource.role_id
    _ = data.roles[role_id]
}

# Verifica se a permissão existe
permission_exists {
    permission_id := input.resource.permission_id
    _ = data.permissions[permission_id]
}

# Verifica se o usuário está no mesmo tenant da função
same_tenant_as_role {
    role_id := input.resource.role_id
    role := data.roles[role_id]
    role.tenant_id == input.user.tenant_id
}

# Verifica se a permissão é considerada crítica
is_critical_permission {
    permission_id := input.resource.permission_id
    permission := data.permissions[permission_id]
    
    # Lista de padrões de permissões críticas que só podem ser atribuídas por super admin
    critical_patterns := [
        "iam:super_admin",
        "iam:admin",
        "tenant:*",
        "system:*",
        "security:*"
    ]
    
    # Verifica se o nome da permissão corresponde a algum padrão crítico
    pattern := critical_patterns[_]
    glob.match(pattern, [], permission.name)
}

# Razões para falha na atribuição de permissão
get_reason_auth_failed = msg {
    not base.user_authenticated
    msg := "usuário não autenticado"
} else = ""

get_reason_tenant_mismatch = msg {
    not base.tenant_scope_valid
    msg := "escopo de tenant inválido"
} else = msg {
    role_exists
    not same_tenant_as_role
    msg := "usuário não pode atribuir permissões a funções em outro tenant"
} else = ""

get_reason_role_not_exists = msg {
    not role_exists
    msg := "a função especificada não existe"
} else = ""

get_reason_permission_not_exists = msg {
    not permission_exists
    msg := "a permissão especificada não existe"
} else = ""

get_reason_critical_permission = msg {
    role_exists
    permission_exists
    is_critical_permission
    not base.is_super_admin
    msg := "apenas super admin pode atribuir esta permissão crítica"
} else = ""

# ==========================================
# === Revogação de Permissão de uma Função ===
# ==========================================

# Decisão final para revogação de permissão
permission_revocation_decision := {
    "allow": allow,
    "reason": reason
} {
    # Computar decisão de autorização
    allow := permission_revocation_allowed
    reason := permission_revocation_reason
}

# Verifica se o usuário tem permissão para revogar uma permissão de uma função
permission_revocation_allowed {
    # Verificação de autenticação
    base.user_authenticated

    # Verificação de escopo de tenant
    base.tenant_scope_valid
    
    # Verificação de integridade dos dados
    base.valid_uuid(input.resource.role_id)
    base.valid_uuid(input.resource.permission_id)
    
    # Acesso especial para super admin
    base.is_super_admin
} else {
    # Verificação de autenticação
    base.user_authenticated

    # Verificação de escopo de tenant
    base.tenant_scope_valid
    
    # Verificação de integridade dos dados
    base.valid_uuid(input.resource.role_id)
    base.valid_uuid(input.resource.permission_id)
    
    # Usuário tem permissão explícita para revogar permissões de funções
    base.has_permission("role:revoke_permission")
    
    # A função existe
    role_exists
    
    # Usuário está no mesmo tenant da função
    same_tenant_as_role
    
    # Não é uma permissão crítica do sistema
    not is_essential_system_permission
}

# Recupera a razão pela qual a revogação de permissão foi permitida ou negada
permission_revocation_reason = reason {
    not permission_revocation_allowed
    reason := concat(" ", [
        get_reason_auth_failed,
        get_reason_tenant_mismatch,
        get_reason_role_not_exists,
        get_reason_essential_permission
    ])
} else {
    permission_revocation_allowed
    reason := "revogação de permissão autorizada"
}

# Verifica se a permissão é essencial para uma função de sistema
# Estas permissões não podem ser revogadas nem mesmo por administradores de tenant
is_essential_system_permission {
    role_id := input.resource.role_id
    permission_id := input.resource.permission_id
    role := data.roles[role_id]
    permission := data.permissions[permission_id]
    
    # Se for função de sistema
    role.type == "SYSTEM"
    
    # E permissão for essencial para a função
    essential_perms := essential_system_permissions[role.name]
    essential_perms[_] == permission.name
}

# Mapeamento de permissões essenciais para funções de sistema
essential_system_permissions := {
    "SUPER_ADMIN": [
        "iam:super_admin",
        "tenant:*"
    ],
    "TENANT_ADMIN": [
        "tenant:manage",
        "role:*"
    ],
    "IAM_ADMIN": [
        "role:create",
        "role:update", 
        "role:delete"
    ]
}

# Razão para falha na revogação de permissão essencial
get_reason_essential_permission = msg {
    is_essential_system_permission
    msg := "não é possível revogar uma permissão essencial de uma função de sistema"
} else = ""

# ==========================================
# === Verificação de Permissão em uma Função ===
# ==========================================

# Decisão final para verificação de permissão
permission_check_decision := {
    "allow": allow,
    "reason": reason,
    "data": result
} {
    # Computar decisão de autorização
    allow := permission_check_allowed
    reason := permission_check_reason
    result := permission_check_data
}

# Verifica se o usuário pode verificar se uma função tem uma determinada permissão
permission_check_allowed {
    # Verificação de autenticação
    base.user_authenticated

    # Verificação de escopo de tenant
    base.tenant_scope_valid
    
    # Verificação de integridade dos dados
    base.valid_uuid(input.resource.role_id)
    
    # Qualquer usuário pode verificar permissões se tiver a permissão específica
    base.has_permission("role:check_permission")
}

# Recupera a razão pela qual a verificação de permissão foi permitida ou negada
permission_check_reason = reason {
    not permission_check_allowed
    reason := concat(" ", [
        get_reason_auth_failed,
        get_reason_tenant_mismatch
    ])
} else {
    permission_check_allowed
    reason := "verificação de permissão autorizada"
}

# Dados retornados pela verificação de permissão
# Para segurança, não mostra todas as permissões para usuários não autorizados
permission_check_data = result {
    permission_check_allowed
    
    # Se for super admin ou tiver permissão específica para ver todas as permissões
    can_see_all := base.is_super_admin || base.has_permission("role:view_all_permissions")
    
    # Filtra dados sensíveis para usuários comuns
    result := can_see_all ? 
              get_full_permission_data :
              get_filtered_permission_data
}

# Retorna dados completos de permissões para usuários autorizados
get_full_permission_data = data {
    role_id := input.resource.role_id
    
    # Recupera todas as permissões da função
    role_permissions := [perm |
        some i
        role_perm := data.role_permissions[i]
        role_perm.role_id == role_id
        perm := data.permissions[role_perm.permission_id]
    ]
    
    data := {
        "role_id": role_id,
        "permissions": role_permissions,
        "count": count(role_permissions)
    }
}

# Retorna dados filtrados de permissões para usuários comuns
get_filtered_permission_data = data {
    role_id := input.resource.role_id
    
    # Lista de tipos de permissões sensíveis que não devem ser mostradas
    sensitive_patterns := [
        "iam:*",
        "security:*",
        "tenant:*",
        "system:*"
    ]
    
    # Recupera permissões, excluindo as sensíveis
    role_permissions := [perm |
        some i
        role_perm := data.role_permissions[i]
        role_perm.role_id == role_id
        permission := data.permissions[role_perm.permission_id]
        
        # Verifica se não é uma permissão sensível
        not is_sensitive_permission(permission.name, sensitive_patterns)
        
        perm := {
            "id": permission.id,
            "name": permission.name
            # Nota: Descrição e outros metadados sensíveis são omitidos
        }
    ]
    
    data := {
        "role_id": role_id,
        "permissions": role_permissions,
        "count": count(role_permissions),
        "filtered": true # Indica que os dados foram filtrados
    }
}

# Verifica se uma permissão é considerada sensível
is_sensitive_permission(permission_name, patterns) {
    pattern := patterns[_]
    glob.match(pattern, [], permission_name)
}