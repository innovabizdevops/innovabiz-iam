# role_hierarchy.rego
# Políticas de autorização para gestão de hierarquia de funções
# Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53, PCI DSS v4.0
package innovabiz.iam.role.hierarchy

import data.innovabiz.iam.role.base
import future.keywords

# ==========================================
# === Adição de Relação Hierárquica entre Funções ===
# ==========================================

# Decisão final para adição de hierarquia
hierarchy_addition_decision := {
    "allow": allow,
    "reason": reason
} {
    # Computar decisão de autorização
    allow := hierarchy_addition_allowed
    reason := hierarchy_addition_reason
}

# Verifica se o usuário tem permissão para adicionar uma relação hierárquica entre funções
hierarchy_addition_allowed {
    # Verificação de autenticação
    base.user_authenticated

    # Verificação de escopo de tenant
    base.tenant_scope_valid
    
    # Verificação de integridade dos dados
    base.valid_uuid(input.resource.parent_id)
    base.valid_uuid(input.resource.child_id)
    
    # Acesso especial para super admin
    base.is_super_admin
} else {
    # Verificação de autenticação
    base.user_authenticated

    # Verificação de escopo de tenant
    base.tenant_scope_valid
    
    # Verificação de integridade dos dados
    base.valid_uuid(input.resource.parent_id)
    base.valid_uuid(input.resource.child_id)
    
    # Usuário tem permissão explícita para gerenciar hierarquia de funções
    base.has_permission("role:manage_hierarchy")
    
    # As funções pai e filho existem
    parent_exists
    child_exists
    
    # As funções estão no mesmo tenant
    same_tenant
    
    # O usuário está no mesmo tenant das funções
    user_in_same_tenant
    
    # A relação hierárquica ainda não existe
    not hierarchy_already_exists
    
    # A adição não criará um ciclo na hierarquia
    not would_create_cycle
    
    # A hierarquia não excede a profundidade máxima permitida para o tenant
    not would_exceed_max_depth
    
    # Não é uma tentativa de adicionar uma função de sistema como filho de uma função personalizada
    not system_role_under_custom_role
    
    # Não é uma tentativa de subordinar uma função administrativa crítica
    not subordinating_critical_role
}

# Recupera a razão pela qual a adição de hierarquia foi permitida ou negada
hierarchy_addition_reason = reason {
    not hierarchy_addition_allowed
    reason := concat(" ", [
        get_reason_auth_failed,
        get_reason_tenant_mismatch,
        get_reason_role_not_exists,
        get_reason_hierarchy_exists,
        get_reason_cycle,
        get_reason_max_depth,
        get_reason_system_under_custom,
        get_reason_critical_role
    ])
} else {
    hierarchy_addition_allowed
    reason := "adição de relação hierárquica autorizada"
}

# === Funções auxiliares para adição de hierarquia ===

# Verifica se a função pai existe
parent_exists {
    parent_id := input.resource.parent_id
    _ = data.roles[parent_id]
}

# Verifica se a função filho existe
child_exists {
    child_id := input.resource.child_id
    _ = data.roles[child_id]
}

# Verifica se as funções pertencem ao mesmo tenant
same_tenant {
    parent_id := input.resource.parent_id
    child_id := input.resource.child_id
    parent := data.roles[parent_id]
    child := data.roles[child_id]
    parent.tenant_id == child.tenant_id
}

# Verifica se o usuário está no mesmo tenant das funções
user_in_same_tenant {
    parent_id := input.resource.parent_id
    parent := data.roles[parent_id]
    parent.tenant_id == input.user.tenant_id
}

# Verifica se a relação hierárquica já existe
hierarchy_already_exists {
    parent_id := input.resource.parent_id
    child_id := input.resource.child_id
    
    some i
    hierarchy := data.role_hierarchies[i]
    hierarchy.parent_id == parent_id
    hierarchy.child_id == child_id
}

# Verifica se a adição da relação criaria um ciclo na hierarquia
would_create_cycle {
    parent_id := input.resource.parent_id
    child_id := input.resource.child_id
    
    # Se o filho já é ancestral do pai (direta ou indiretamente),
    # adicionar esta relação criaria um ciclo
    is_ancestor(child_id, parent_id)
}

# Função recursiva para verificar se role_id é ancestral de target_id
is_ancestor(role_id, target_id) {
    # Caso base: role_id é pai direto de target_id
    some i
    hierarchy := data.role_hierarchies[i]
    hierarchy.parent_id == role_id
    hierarchy.child_id == target_id
} else {
    # Caso recursivo: role_id é ancestral de um pai de target_id
    some i
    hierarchy := data.role_hierarchies[i]
    hierarchy.child_id == target_id
    is_ancestor(role_id, hierarchy.parent_id)
}

# Verifica se a adição excederia a profundidade máxima de hierarquia
would_exceed_max_depth {
    parent_id := input.resource.parent_id
    child_id := input.resource.child_id
    
    # Obter configuração de profundidade máxima para o tenant
    tenant_id := data.roles[parent_id].tenant_id
    settings := data.tenant_settings[tenant_id]
    max_depth := settings.max_hierarchy_depth
    
    # Calcular a profundidade atual da função pai
    parent_depth := role_depth(parent_id)
    
    # Se o pai já está na profundidade máxima, não podemos adicionar um filho
    parent_depth >= max_depth
}

# Calcula a profundidade de uma função na hierarquia
# Retorna 0 para funções de topo (sem pais)
role_depth(role_id) = max_depth {
    # Encontrar todas as profundidades de todos os caminhos para esta função
    depths := [depth |
        parent_id := get_parent(role_id)
        parent_id != null
        depth := 1 + role_depth(parent_id)
    ]
    
    # Se a função não tem pais, a profundidade é 0
    count(depths) == 0
    max_depth := 0
} else = max_depth {
    # Encontrar todas as profundidades de todos os caminhos para esta função
    depths := [depth |
        parent_id := get_parent(role_id)
        parent_id != null
        depth := 1 + role_depth(parent_id)
    ]
    
    # A profundidade da função é o caminho mais longo até ela
    max_depth := max(depths)
}

# Obtém o ID da função pai, ou null se não houver pai
get_parent(role_id) = parent_id {
    some i
    hierarchy := data.role_hierarchies[i]
    hierarchy.child_id == role_id
    parent_id := hierarchy.parent_id
} else = null {
    true
}

# Verifica se é uma tentativa de adicionar uma função de sistema como filho de uma função personalizada
system_role_under_custom_role {
    parent_id := input.resource.parent_id
    child_id := input.resource.child_id
    
    parent := data.roles[parent_id]
    child := data.roles[child_id]
    
    parent.type == "CUSTOM"
    child.type == "SYSTEM"
}

# Verifica se é uma tentativa de subordinar uma função administrativa crítica
subordinating_critical_role {
    child_id := input.resource.child_id
    child := data.roles[child_id]
    
    # Lista de nomes de funções administrativas críticas
    critical_roles := [
        "SUPER_ADMIN",
        "TENANT_ADMIN",
        "IAM_ADMIN",
        "SECURITY_OFFICER"
    ]
    
    # Verifica se a função filho é uma função crítica
    some i
    critical_roles[i] == child.name
    
    # Para funções críticas, apenas super admin pode adicionar relações hierárquicas
    not base.is_super_admin
}

# Razões para falha na adição de hierarquia
get_reason_auth_failed = msg {
    not base.user_authenticated
    msg := "usuário não autenticado"
} else = ""

get_reason_tenant_mismatch = msg {
    not base.tenant_scope_valid
    msg := "escopo de tenant inválido"
} else = msg {
    parent_exists
    child_exists
    not same_tenant
    msg := "as funções pai e filho devem pertencer ao mesmo tenant"
} else = msg {
    parent_exists
    not user_in_same_tenant
    msg := "usuário não tem permissão para gerenciar hierarquia em outro tenant"
} else = ""

get_reason_role_not_exists = msg {
    not parent_exists
    not child_exists
    msg := "as funções pai e filho especificadas não existem"
} else = msg {
    not parent_exists
    msg := "a função pai especificada não existe"
} else = msg {
    not child_exists
    msg := "a função filho especificada não existe"
} else = ""

get_reason_hierarchy_exists = msg {
    hierarchy_already_exists
    msg := "a relação hierárquica já existe"
} else = ""

get_reason_cycle = msg {
    would_create_cycle
    msg := "a adição desta relação hierárquica criaria um ciclo"
} else = ""

get_reason_max_depth = msg {
    would_exceed_max_depth
    parent_id := input.resource.parent_id
    tenant_id := data.roles[parent_id].tenant_id
    settings := data.tenant_settings[tenant_id]
    max_depth := settings.max_hierarchy_depth
    msg := sprintf("a adição desta relação hierárquica excederia a profundidade máxima permitida (%d)", [max_depth])
} else = ""

get_reason_system_under_custom = msg {
    system_role_under_custom_role
    msg := "não é permitido adicionar uma função de sistema como subordinada a uma função personalizada"
} else = ""

get_reason_critical_role = msg {
    subordinating_critical_role
    msg := "apenas super admin pode adicionar relações hierárquicas para funções administrativas críticas"
} else = ""

# ==========================================
# === Remoção de Relação Hierárquica entre Funções ===
# ==========================================

# Decisão final para remoção de hierarquia
hierarchy_removal_decision := {
    "allow": allow,
    "reason": reason
} {
    # Computar decisão de autorização
    allow := hierarchy_removal_allowed
    reason := hierarchy_removal_reason
}

# Verifica se o usuário tem permissão para remover uma relação hierárquica
hierarchy_removal_allowed {
    # Verificação de autenticação
    base.user_authenticated

    # Verificação de escopo de tenant
    base.tenant_scope_valid
    
    # Verificação de integridade dos dados
    base.valid_uuid(input.resource.hierarchy_id)
    
    # Acesso especial para super admin
    base.is_super_admin
} else {
    # Verificação de autenticação
    base.user_authenticated

    # Verificação de escopo de tenant
    base.tenant_scope_valid
    
    # Verificação de integridade dos dados
    base.valid_uuid(input.resource.hierarchy_id)
    
    # Usuário tem permissão explícita para gerenciar hierarquia
    base.has_permission("role:manage_hierarchy")
    
    # A relação hierárquica existe
    hierarchy_exists
    
    # O usuário está no mesmo tenant da hierarquia
    hierarchy_in_user_tenant
    
    # Não é uma relação hierárquica do sistema
    not is_system_hierarchy
}

# Recupera a razão pela qual a remoção de hierarquia foi permitida ou negada
hierarchy_removal_reason = reason {
    not hierarchy_removal_allowed
    reason := concat(" ", [
        get_reason_auth_failed,
        get_reason_hierarchy_not_exists,
        get_reason_tenant_mismatch_removal,
        get_reason_system_hierarchy
    ])
} else {
    hierarchy_removal_allowed
    reason := "remoção de relação hierárquica autorizada"
}

# === Funções auxiliares para remoção de hierarquia ===

# Verifica se a relação hierárquica existe
hierarchy_exists {
    hierarchy_id := input.resource.hierarchy_id
    
    some i
    hierarchy := data.role_hierarchies[i]
    hierarchy.id == hierarchy_id
}

# Verifica se o usuário está no mesmo tenant da hierarquia
hierarchy_in_user_tenant {
    hierarchy_id := input.resource.hierarchy_id
    
    # Encontrar a hierarquia
    some i
    hierarchy := data.role_hierarchies[i]
    hierarchy.id == hierarchy_id
    
    # Verificar se o tenant da hierarquia é o mesmo do usuário
    hierarchy.tenant_id == input.user.tenant_id
}

# Verifica se é uma relação hierárquica do sistema
is_system_hierarchy {
    hierarchy_id := input.resource.hierarchy_id
    
    # Encontrar a hierarquia
    some i
    hierarchy := data.role_hierarchies[i]
    hierarchy.id == hierarchy_id
    
    # Obter as funções pai e filho
    parent := data.roles[hierarchy.parent_id]
    child := data.roles[hierarchy.child_id]
    
    # Se ambas forem funções de sistema, é uma hierarquia do sistema
    parent.type == "SYSTEM"
    child.type == "SYSTEM"
}

# Razões para falha na remoção de hierarquia
get_reason_hierarchy_not_exists = msg {
    not hierarchy_exists
    msg := "a relação hierárquica especificada não existe"
} else = ""

get_reason_tenant_mismatch_removal = msg {
    hierarchy_exists
    not hierarchy_in_user_tenant
    msg := "usuário não tem permissão para gerenciar hierarquia em outro tenant"
} else = ""

get_reason_system_hierarchy = msg {
    is_system_hierarchy
    msg := "não é permitido remover relações hierárquicas entre funções de sistema"
} else = ""

# ==========================================
# === Consulta de Hierarquia de Funções ===
# ==========================================

# Decisão final para consulta de hierarquia
hierarchy_query_decision := {
    "allow": allow,
    "reason": reason,
    "data": result
} {
    # Computar decisão de autorização
    allow := hierarchy_query_allowed
    reason := hierarchy_query_reason
    result := hierarchy_query_data
}

# Verifica se o usuário pode consultar a hierarquia de uma função
hierarchy_query_allowed {
    # Verificação de autenticação
    base.user_authenticated

    # Verificação de escopo de tenant
    base.tenant_scope_valid
    
    # Verificação de integridade dos dados
    base.valid_uuid(input.resource.role_id)
    
    # Super admin pode consultar qualquer hierarquia
    base.is_super_admin
} else {
    # Verificação de autenticação
    base.user_authenticated

    # Verificação de escopo de tenant
    base.tenant_scope_valid
    
    # Verificação de integridade dos dados
    base.valid_uuid(input.resource.role_id)
    
    # Usuário tem permissão para consultar hierarquia ou gerenciar hierarquia
    base.has_permission("role:read_hierarchy") or base.has_permission("role:manage_hierarchy")
    
    # A função existe
    role_exists_for_query
    
    # O usuário está no mesmo tenant da função
    role_in_user_tenant_for_query
}

# Recupera a razão pela qual a consulta de hierarquia foi permitida ou negada
hierarchy_query_reason = reason {
    not hierarchy_query_allowed
    reason := concat(" ", [
        get_reason_auth_failed,
        get_reason_role_not_exists_query,
        get_reason_tenant_mismatch_query
    ])
} else {
    hierarchy_query_allowed
    reason := "consulta de hierarquia autorizada"
}

# Dados retornados pela consulta de hierarquia
hierarchy_query_data = result {
    hierarchy_query_allowed
    
    role_id := input.resource.role_id
    
    # Recuperar pais e filhos da função
    parents := [parent |
        some i
        hierarchy := data.role_hierarchies[i]
        hierarchy.child_id == role_id
        parent_role := data.roles[hierarchy.parent_id]
        
        parent := {
            "id": parent_role.id,
            "name": parent_role.name,
            "type": parent_role.type,
            "status": parent_role.status,
            "hierarchy_id": hierarchy.id
        }
    ]
    
    children := [child |
        some i
        hierarchy := data.role_hierarchies[i]
        hierarchy.parent_id == role_id
        child_role := data.roles[hierarchy.child_id]
        
        child := {
            "id": child_role.id,
            "name": child_role.name,
            "type": child_role.type,
            "status": child_role.status,
            "hierarchy_id": hierarchy.id
        }
    ]
    
    # Filtrar dados sensíveis se o usuário não for admin
    can_see_full := base.is_super_admin || base.is_tenant_admin || base.has_permission("role:view_all_hierarchy")
    
    result := {
        "role_id": role_id,
        "parents": parents,
        "children": children,
        "depth": role_depth(role_id),
        "filtered": not can_see_full
    }
}

# === Funções auxiliares para consulta de hierarquia ===

# Verifica se a função existe para consulta
role_exists_for_query {
    role_id := input.resource.role_id
    _ = data.roles[role_id]
}

# Verifica se o usuário está no mesmo tenant da função para consulta
role_in_user_tenant_for_query {
    role_id := input.resource.role_id
    role := data.roles[role_id]
    role.tenant_id == input.user.tenant_id
}

# Razões para falha na consulta de hierarquia
get_reason_role_not_exists_query = msg {
    not role_exists_for_query
    msg := "a função especificada não existe"
} else = ""

get_reason_tenant_mismatch_query = msg {
    role_exists_for_query
    not role_in_user_tenant_for_query
    msg := "usuário não tem permissão para consultar hierarquia em outro tenant"
} else = ""