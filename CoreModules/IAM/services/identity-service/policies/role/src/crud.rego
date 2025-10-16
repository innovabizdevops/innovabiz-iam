# Políticas CRUD para gerenciamento de funções - INNOVABIZ Platform
#
# Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
# PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III
package innovabiz.iam.role.crud

import data.innovabiz.iam.role.common
import data.innovabiz.iam.role.constants
import data.innovabiz.iam.role.audit

# ---------------------------------------------------------
# Decisão para criação de função (create)
# ---------------------------------------------------------
create_decision := true {
    # Verificar método HTTP
    input.http_method == "POST"
    
    # Super Admin pode criar qualquer tipo de função em qualquer tenant
    common.has_role(constants.super_admin_role)
}

create_decision := true {
    # Verificar método HTTP
    input.http_method == "POST"
    
    # Tenant Admin pode criar funções customizadas apenas no seu próprio tenant
    common.has_role(constants.tenant_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Tenant Admin só pode criar funções customizadas, não funções de sistema
    role_data := input.resource.data
    role_data.type == constants.custom_role_type
    
    # Verificar se a função não tem nome reservado para funções de sistema
    not role_has_protected_name(role_data.name)
}

create_decision := true {
    # Verificar método HTTP
    input.http_method == "POST"
    
    # IAM Admin pode criar funções customizadas apenas no seu próprio tenant
    common.has_role(constants.iam_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # IAM Admin só pode criar funções customizadas, não funções de sistema
    role_data := input.resource.data
    role_data.type == constants.custom_role_type
    
    # Verificar se a função não tem nome reservado para funções de sistema
    not role_has_protected_name(role_data.name)
    
    # Verificar se o usuário tem permissão específica
    common.has_permission("role:create")
}

# Verificar se o nome da função é reservado para funções de sistema
role_has_protected_name(name) {
    protected_names := [
        "SUPER_ADMIN",
        "TENANT_ADMIN",
        "IAM_ADMIN",
        "IAM_OPERATOR",
        "SYSTEM_ADMIN",
        "SECURITY_ADMIN",
        "AUDIT_ADMIN"
    ]
    
    name_upper := upper(name)
    protected_name := protected_names[_]
    name_upper == protected_name
}

# ---------------------------------------------------------
# Decisão para leitura de funções (read)
# ---------------------------------------------------------
read_decision := true {
    # Verificar método HTTP
    input.http_method == "GET"
    
    # Super Admin pode ler qualquer função em qualquer tenant
    common.has_role(constants.super_admin_role)
}

read_decision := true {
    # Verificar método HTTP
    input.http_method == "GET"
    
    # Admin de tenant pode ler funções no seu próprio tenant
    common.has_role(constants.tenant_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
}

read_decision := true {
    # Verificar método HTTP
    input.http_method == "GET"
    
    # IAM Admin pode ler funções no seu próprio tenant
    common.has_role(constants.iam_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
}

read_decision := true {
    # Verificar método HTTP
    input.http_method == "GET"
    
    # IAM Operator pode ler funções no seu próprio tenant
    common.has_role(constants.iam_operator_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Verificar se o usuário tem permissão específica
    common.has_permission("role:read")
}

# ---------------------------------------------------------
# Decisão para atualização de função (update)
# ---------------------------------------------------------
update_decision := true {
    # Verificar método HTTP
    input.http_method == "PUT"
    
    # Super Admin pode atualizar qualquer função em qualquer tenant
    common.has_role(constants.super_admin_role)
}

update_decision := true {
    # Verificar método HTTP
    input.http_method == "PUT"
    
    # Tenant Admin pode atualizar funções customizadas no seu tenant
    common.has_role(constants.tenant_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Verificar se não é uma função de sistema (não pode alterar funções do sistema)
    role_data := input.resource.data
    role_data.type == constants.custom_role_type
    
    # Verificar se a função não tem nome reservado para funções de sistema
    not role_has_protected_name(role_data.name)
}

update_decision := true {
    # Verificar método HTTP
    input.http_method == "PUT"
    
    # IAM Admin pode atualizar funções customizadas no seu tenant
    common.has_role(constants.iam_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Verificar se não é uma função de sistema (não pode alterar funções do sistema)
    role_data := input.resource.data
    role_data.type == constants.custom_role_type
    
    # Verificar se a função não tem nome reservado para funções de sistema
    not role_has_protected_name(role_data.name)
    
    # Verificar se o usuário tem permissão específica
    common.has_permission("role:update")
}

# ---------------------------------------------------------
# Decisão para exclusão lógica de função (delete)
# ---------------------------------------------------------
delete_decision := true {
    # Verificar método HTTP
    input.http_method == "DELETE"
    
    # Super Admin pode excluir qualquer função em qualquer tenant
    common.has_role(constants.super_admin_role)
}

delete_decision := true {
    # Verificar método HTTP
    input.http_method == "DELETE"
    
    # Tenant Admin pode excluir funções customizadas no seu tenant
    common.has_role(constants.tenant_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Obter dados da função a ser excluída
    role_id := input.resource.id
    role := get_role_by_id(role_id)
    
    # Verificar se não é uma função de sistema (não pode excluir funções do sistema)
    role.type == constants.custom_role_type
    
    # Verificar se a função não tem nome reservado para funções de sistema
    not role_has_protected_name(role.name)
}

delete_decision := true {
    # Verificar método HTTP
    input.http_method == "DELETE"
    
    # IAM Admin pode excluir funções customizadas no seu tenant
    common.has_role(constants.iam_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Obter dados da função a ser excluída
    role_id := input.resource.id
    role := get_role_by_id(role_id)
    
    # Verificar se não é uma função de sistema (não pode excluir funções do sistema)
    role.type == constants.custom_role_type
    
    # Verificar se a função não tem nome reservado para funções de sistema
    not role_has_protected_name(role.name)
    
    # Verificar se o usuário tem permissão específica
    common.has_permission("role:delete")
}

# ---------------------------------------------------------
# Decisão para exclusão permanente de função (permanent delete)
# ---------------------------------------------------------
permanent_delete_decision := true {
    # Verificar método HTTP
    input.http_method == "DELETE"
    
    # Apenas Super Admin pode excluir permanentemente qualquer função em qualquer tenant
    common.has_role(constants.super_admin_role)
    
    # Verificar caminho específico para exclusão permanente
    contains(input.resource.path, "/permanent")
}

permanent_delete_decision := true {
    # Verificar método HTTP
    input.http_method == "DELETE"
    
    # Tenant Admin pode excluir permanentemente funções customizadas no seu tenant
    # desde que tenha permissão especial para isso
    common.has_role(constants.tenant_admin_role)
    common.has_permission("role:permanent_delete")
    
    # Verificar caminho específico para exclusão permanente
    contains(input.resource.path, "/permanent")
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Obter dados da função a ser excluída
    role_id := input.resource.id
    role := get_role_by_id(role_id)
    
    # Verificar se não é uma função de sistema (não pode excluir funções do sistema)
    role.type == constants.custom_role_type
    
    # Verificar se a função não tem nome reservado para funções de sistema
    not role_has_protected_name(role.name)
}

# ---------------------------------------------------------
# Decisão para listar funções (list)
# ---------------------------------------------------------
list_decision := true {
    # Verificar método HTTP
    input.http_method == "GET"
    
    # Verificar se é uma operação de listagem (sem ID específico na URL)
    not input.resource.id
    
    # Super Admin pode listar qualquer função em qualquer tenant
    common.has_role(constants.super_admin_role)
}

list_decision := true {
    # Verificar método HTTP
    input.http_method == "GET"
    
    # Verificar se é uma operação de listagem (sem ID específico na URL)
    not input.resource.id
    
    # Qualquer usuário com uma das funções administrativas pode listar funções no seu tenant
    admin_roles := {constants.tenant_admin_role, constants.iam_admin_role, constants.iam_operator_role}
    common.has_any_role(admin_roles)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
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