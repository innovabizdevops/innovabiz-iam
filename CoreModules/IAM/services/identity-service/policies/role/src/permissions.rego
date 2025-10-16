# Políticas para atribuição e gestão de permissões - INNOVABIZ Platform
#
# Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
# PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III
package innovabiz.iam.role.permissions

import data.innovabiz.iam.role.common
import data.innovabiz.iam.role.constants

# ---------------------------------------------------------
# Decisão para atribuição de permissões
# ---------------------------------------------------------
permission_assignment_decision := true {
    # Verificar método HTTP
    input.http_method == "POST"
    
    # Super Admin pode atribuir qualquer permissão a qualquer função em qualquer tenant
    common.has_role(constants.super_admin_role)
}

permission_assignment_decision := true {
    # Verificar método HTTP
    input.http_method == "POST"
    
    # Tenant Admin pode atribuir permissões a funções customizadas no seu tenant
    common.has_role(constants.tenant_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Obter dados da função
    role_id := input.resource.role_id
    role := get_role_by_id(role_id)
    
    # Verificar se não é uma função de sistema
    role.type == constants.custom_role_type
    
    # Verificar se a permissão não é crítica/restrita
    permission_id := input.resource.permission_id
    permission := get_permission_by_id(permission_id)
    not is_restricted_permission(permission.name)
}

permission_assignment_decision := true {
    # Verificar método HTTP
    input.http_method == "POST"
    
    # IAM Admin pode atribuir permissões a funções customizadas no seu tenant
    common.has_role(constants.iam_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Verificar se o usuário tem permissão específica
    common.has_permission("role:assign_permission")
    
    # Obter dados da função
    role_id := input.resource.role_id
    role := get_role_by_id(role_id)
    
    # Verificar se não é uma função de sistema
    role.type == constants.custom_role_type
    
    # Verificar se a permissão não é crítica/restrita
    permission_id := input.resource.permission_id
    permission := get_permission_by_id(permission_id)
    not is_restricted_permission(permission.name)
}

# ---------------------------------------------------------
# Decisão para revogação de permissões
# ---------------------------------------------------------
permission_revocation_decision := true {
    # Verificar método HTTP
    input.http_method == "DELETE"
    
    # Super Admin pode revogar qualquer permissão de qualquer função em qualquer tenant
    common.has_role(constants.super_admin_role)
}

permission_revocation_decision := true {
    # Verificar método HTTP
    input.http_method == "DELETE"
    
    # Tenant Admin pode revogar permissões de funções customizadas no seu tenant
    common.has_role(constants.tenant_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Obter dados da função
    role_id := input.resource.role_id
    role := get_role_by_id(role_id)
    
    # Verificar se não é uma função de sistema
    role.type == constants.custom_role_type
}

permission_revocation_decision := true {
    # Verificar método HTTP
    input.http_method == "DELETE"
    
    # IAM Admin pode revogar permissões de funções customizadas no seu tenant
    common.has_role(constants.iam_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Verificar se o usuário tem permissão específica
    common.has_permission("role:revoke_permission")
    
    # Obter dados da função
    role_id := input.resource.role_id
    role := get_role_by_id(role_id)
    
    # Verificar se não é uma função de sistema
    role.type == constants.custom_role_type
}

# ---------------------------------------------------------
# Decisão para listar permissões de uma função
# ---------------------------------------------------------
permission_check_decision := true {
    # Verificar método HTTP
    input.http_method == "GET"
    
    # Super Admin pode verificar qualquer permissão de qualquer função em qualquer tenant
    common.has_role(constants.super_admin_role)
}

permission_check_decision := true {
    # Verificar método HTTP
    input.http_method == "GET"
    
    # Qualquer usuário com uma das funções administrativas pode verificar permissões no seu tenant
    admin_roles := {constants.tenant_admin_role, constants.iam_admin_role, constants.iam_operator_role}
    common.has_any_role(admin_roles)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
}

# Função para verificar se uma permissão é restrita
is_restricted_permission(permission_name) {
    restricted_permissions := [
        "super_admin:*",
        "tenant:create",
        "tenant:delete",
        "system:config",
        "system:logs",
        "system:security",
        "iam:all_tenants",
        "audit:delete",
        "role:permanent_delete"
    ]
    
    permission := restricted_permissions[_]
    glob.match(permission, [], permission_name)
}

# Função auxiliar para obter dados de uma função pelo ID (simulação)
# Em produção, isso seria obtido de um banco de dados ou outro serviço
get_role_by_id(role_id) = role {
    # Simular recuperação de dados da função pelo ID
    # Para um ambiente real, isso seria obtido de uma fonte externa
    role_id == "1"
    role := {"id": "1", "name": "ADMIN", "type": "SYSTEM"}
} else = role {
    # Função customizada para testes
    role_id == "2"
    role := {"id": "2", "name": "CUSTOM_ROLE", "type": "CUSTOM"}
} else = {
    "id": role_id,
    "name": "UNKNOWN",
    "type": "CUSTOM"
}

# Função auxiliar para obter dados de uma permissão pelo ID (simulação)
get_permission_by_id(permission_id) = permission {
    permission_id == "1"
    permission := {"id": "1", "name": "role:create", "description": "Permite criar funções"}
} else = permission {
    permission_id == "2"
    permission := {"id": "2", "name": "role:read", "description": "Permite ler funções"}
} else = permission {
    permission_id == "3"
    permission := {"id": "3", "name": "role:update", "description": "Permite atualizar funções"}
} else = permission {
    permission_id == "4"
    permission := {"id": "4", "name": "role:delete", "description": "Permite excluir funções"}
} else = permission {
    permission_id == "5"
    permission := {"id": "5", "name": "role:assign_permission", "description": "Permite atribuir permissões"}
} else = {
    "id": permission_id,
    "name": "unknown:permission",
    "description": "Permissão desconhecida"
}