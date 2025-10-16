# Políticas para atribuição de funções a usuários - INNOVABIZ Platform
#
# Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
# PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III
package innovabiz.iam.role.user_assignment

import data.innovabiz.iam.role.common
import data.innovabiz.iam.role.constants
import data.innovabiz.iam.role.audit

# ---------------------------------------------------------
# Decisão para atribuição de função a usuário
# ---------------------------------------------------------
role_assignment_decision := true {
    # Verificar método HTTP
    input.http_method == "POST"
    
    # Super Admin pode atribuir qualquer função a qualquer usuário em qualquer tenant
    common.has_role(constants.super_admin_role)
}

role_assignment_decision := true {
    # Verificar método HTTP
    input.http_method == "POST"
    
    # Tenant Admin pode atribuir funções customizadas a usuários no seu tenant
    common.has_role(constants.tenant_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Obter dados da atribuição
    assignment_data := input.resource.data
    role_id := assignment_data.role_id
    
    # Obter dados da função
    role := get_role_by_id(role_id)
    
    # Verificar se não é uma função de sistema protegida
    not is_protected_role(role.name)
    
    # Verificar se a função pertence ao mesmo tenant
    role.tenant_id == input.tenant_id
    
    # Verificar se o usuário alvo pertence ao mesmo tenant
    target_user_id := assignment_data.user_id
    target_user := get_user_by_id(target_user_id)
    target_user.tenant_id == input.tenant_id
    
    # Verificar se o usuário que está fazendo a atribuição não está atribuindo a si mesmo
    input.user.id != target_user_id
}

role_assignment_decision := true {
    # Verificar método HTTP
    input.http_method == "POST"
    
    # IAM Admin pode atribuir funções customizadas a usuários no seu tenant
    common.has_role(constants.iam_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Verificar se o usuário tem permissão específica
    common.has_permission("role:assign_to_user")
    
    # Obter dados da atribuição
    assignment_data := input.resource.data
    role_id := assignment_data.role_id
    
    # Obter dados da função
    role := get_role_by_id(role_id)
    
    # Verificar se não é uma função de sistema protegida
    not is_protected_role(role.name)
    
    # Verificar se a função pertence ao mesmo tenant
    role.tenant_id == input.tenant_id
    
    # Verificar se o usuário alvo pertence ao mesmo tenant
    target_user_id := assignment_data.user_id
    target_user := get_user_by_id(target_user_id)
    target_user.tenant_id == input.tenant_id
    
    # Verificar se o usuário que está fazendo a atribuição não está atribuindo a si mesmo
    input.user.id != target_user_id
    
    # Verificar se há metadados de justificativa
    assignment_data.metadata.justification
}

# ---------------------------------------------------------
# Decisão para remoção de função de usuário
# ---------------------------------------------------------
role_removal_decision := true {
    # Verificar método HTTP
    input.http_method == "DELETE"
    
    # Super Admin pode remover qualquer atribuição de função em qualquer tenant
    common.has_role(constants.super_admin_role)
}

role_removal_decision := true {
    # Verificar método HTTP
    input.http_method == "DELETE"
    
    # Tenant Admin pode remover atribuições de funções no seu tenant
    common.has_role(constants.tenant_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Obter dados da atribuição
    assignment_id := input.resource.assignment_id
    assignment := get_assignment_by_id(assignment_id)
    
    # Verificar se a atribuição pertence ao mesmo tenant
    assignment.tenant_id == input.tenant_id
    
    # Verificar se não é uma função de sistema protegida
    role := get_role_by_id(assignment.role_id)
    not is_protected_role(role.name)
    
    # Verificar se o usuário que está removendo não está removendo de si mesmo
    input.user.id != assignment.user_id
}

role_removal_decision := true {
    # Verificar método HTTP
    input.http_method == "DELETE"
    
    # IAM Admin pode remover atribuições de funções no seu tenant
    common.has_role(constants.iam_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Verificar se o usuário tem permissão específica
    common.has_permission("role:remove_from_user")
    
    # Obter dados da atribuição
    assignment_id := input.resource.assignment_id
    assignment := get_assignment_by_id(assignment_id)
    
    # Verificar se a atribuição pertence ao mesmo tenant
    assignment.tenant_id == input.tenant_id
    
    # Verificar se não é uma função de sistema protegida
    role := get_role_by_id(assignment.role_id)
    not is_protected_role(role.name)
    
    # Verificar se o usuário que está removendo não está removendo de si mesmo
    input.user.id != assignment.user_id
}

# ---------------------------------------------------------
# Decisão para atualização da data de expiração
# ---------------------------------------------------------
expiration_update_decision := true {
    # Verificar método HTTP
    input.http_method == "PUT"
    
    # Super Admin pode atualizar qualquer data de expiração em qualquer tenant
    common.has_role(constants.super_admin_role)
}

expiration_update_decision := true {
    # Verificar método HTTP
    input.http_method == "PUT"
    
    # Tenant Admin pode atualizar datas de expiração no seu tenant
    common.has_role(constants.tenant_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Obter dados da atribuição
    assignment_id := input.resource.assignment_id
    assignment := get_assignment_by_id(assignment_id)
    
    # Verificar se a atribuição pertence ao mesmo tenant
    assignment.tenant_id == input.tenant_id
    
    # Verificar se não é uma função de sistema protegida
    role := get_role_by_id(assignment.role_id)
    not is_protected_role(role.name)
}

expiration_update_decision := true {
    # Verificar método HTTP
    input.http_method == "PUT"
    
    # IAM Admin pode atualizar datas de expiração no seu tenant
    common.has_role(constants.iam_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Verificar se o usuário tem permissão específica
    common.has_permission("role:update_expiration")
    
    # Obter dados da atribuição
    assignment_id := input.resource.assignment_id
    assignment := get_assignment_by_id(assignment_id)
    
    # Verificar se a atribuição pertence ao mesmo tenant
    assignment.tenant_id == input.tenant_id
    
    # Verificar se não é uma função de sistema protegida
    role := get_role_by_id(assignment.role_id)
    not is_protected_role(role.name)
}

# ---------------------------------------------------------
# Decisão para consulta de funções de usuário
# ---------------------------------------------------------
role_check_decision := true {
    # Verificar método HTTP
    input.http_method == "GET"
    
    # Super Admin pode consultar funções de qualquer usuário em qualquer tenant
    common.has_role(constants.super_admin_role)
}

role_check_decision := true {
    # Verificar método HTTP
    input.http_method == "GET"
    
    # Qualquer usuário com uma das funções administrativas pode consultar funções no seu tenant
    admin_roles := {constants.tenant_admin_role, constants.iam_admin_role, constants.iam_operator_role}
    common.has_any_role(admin_roles)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Obter ID do usuário alvo
    target_user_id := input.resource.user_id
    
    # Verificar se o usuário alvo pertence ao mesmo tenant
    target_user := get_user_by_id(target_user_id)
    target_user.tenant_id == input.tenant_id
}

role_check_decision := true {
    # Verificar método HTTP
    input.http_method == "GET"
    
    # Um usuário pode consultar suas próprias funções
    input.resource.user_id == input.user.id
}

# ---------------------------------------------------------
# Funções auxiliares
# ---------------------------------------------------------

# Função para verificar se uma função é protegida
is_protected_role(role_name) {
    protected_roles := [
        "SUPER_ADMIN",
        "TENANT_ADMIN",
        "IAM_ADMIN",
        "SYSTEM_ADMIN",
        "SECURITY_ADMIN",
        "AUDIT_ADMIN"
    ]
    
    name_upper := upper(role_name)
    protected_name := protected_roles[_]
    name_upper == protected_name
}

# Função auxiliar para obter dados de uma função pelo ID (simulação)
get_role_by_id(role_id) = role {
    # Simular recuperação de dados da função pelo ID
    # Para um ambiente real, isso seria obtido de uma fonte externa
    role_id == "1"
    role := {
        "id": "1",
        "name": "ADMIN",
        "type": "SYSTEM",
        "tenant_id": "00000000-0000-0000-0000-000000000000"
    }
} else = role {
    # Função customizada para testes
    role_id == "2"
    role := {
        "id": "2",
        "name": "CUSTOM_ROLE",
        "type": "CUSTOM",
        "tenant_id": "tenant-1"
    }
} else = {
    "id": role_id,
    "name": "UNKNOWN",
    "type": "CUSTOM",
    "tenant_id": "tenant-1"
}

# Função auxiliar para obter dados de um usuário pelo ID (simulação)
get_user_by_id(user_id) = user {
    # Simular recuperação de dados do usuário pelo ID
    user_id == "user-1"
    user := {
        "id": "user-1",
        "email": "admin@example.com",
        "tenant_id": "tenant-1"
    }
} else = user {
    user_id == "user-2"
    user := {
        "id": "user-2",
        "email": "user@example.com",
        "tenant_id": "tenant-1"
    }
} else = user {
    user_id == "super-admin"
    user := {
        "id": "super-admin",
        "email": "super@innovabiz.com",
        "tenant_id": "00000000-0000-0000-0000-000000000000"
    }
} else = {
    "id": user_id,
    "email": "unknown@example.com",
    "tenant_id": "tenant-1"
}

# Função auxiliar para obter dados de uma atribuição de função pelo ID (simulação)
get_assignment_by_id(assignment_id) = assignment {
    # Simular recuperação de dados da atribuição pelo ID
    assignment_id == "assign-1"
    assignment := {
        "id": "assign-1",
        "role_id": "1",
        "user_id": "user-1",
        "tenant_id": "tenant-1"
    }
} else = assignment {
    assignment_id == "assign-2"
    assignment := {
        "id": "assign-2",
        "role_id": "2",
        "user_id": "user-2",
        "tenant_id": "tenant-1"
    }
} else = {
    "id": assignment_id,
    "role_id": "unknown",
    "user_id": "unknown",
    "tenant_id": "tenant-1"
}