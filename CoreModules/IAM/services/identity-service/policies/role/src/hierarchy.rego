# Políticas para gestão de hierarquia de funções - INNOVABIZ Platform
#
# Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
# PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III
package innovabiz.iam.role.hierarchy

import data.innovabiz.iam.role.common
import data.innovabiz.iam.role.constants

# ---------------------------------------------------------
# Decisão para adição de hierarquia
# ---------------------------------------------------------
hierarchy_addition_decision := true {
    # Verificar método HTTP
    input.http_method == "POST"
    
    # Super Admin pode gerenciar hierarquia de qualquer função em qualquer tenant
    common.has_role(constants.super_admin_role)
}

hierarchy_addition_decision := true {
    # Verificar método HTTP
    input.http_method == "POST"
    
    # Tenant Admin pode gerenciar hierarquia de funções customizadas no seu tenant
    common.has_role(constants.tenant_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Obter dados da hierarquia
    parent_id := input.resource.data.parent_id
    child_id := input.resource.data.child_id
    
    # Obter dados das funções
    parent_role := get_role_by_id(parent_id)
    child_role := get_role_by_id(child_id)
    
    # Verificar se ambas são funções customizadas
    parent_role.type == constants.custom_role_type
    child_role.type == constants.custom_role_type
    
    # Verificar se não cria ciclos na hierarquia
    not would_create_hierarchy_cycle(parent_id, child_id)
    
    # Verificar se não excede a profundidade máxima de hierarquia
    not would_exceed_max_hierarchy_depth(parent_id, child_id)
}

hierarchy_addition_decision := true {
    # Verificar método HTTP
    input.http_method == "POST"
    
    # IAM Admin pode gerenciar hierarquia de funções customizadas no seu tenant
    common.has_role(constants.iam_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Verificar se o usuário tem permissão específica
    common.has_permission("role:manage_hierarchy")
    
    # Obter dados da hierarquia
    parent_id := input.resource.data.parent_id
    child_id := input.resource.data.child_id
    
    # Obter dados das funções
    parent_role := get_role_by_id(parent_id)
    child_role := get_role_by_id(child_id)
    
    # Verificar se ambas são funções customizadas
    parent_role.type == constants.custom_role_type
    child_role.type == constants.custom_role_type
    
    # Verificar se não cria ciclos na hierarquia
    not would_create_hierarchy_cycle(parent_id, child_id)
    
    # Verificar se não excede a profundidade máxima de hierarquia
    not would_exceed_max_hierarchy_depth(parent_id, child_id)
}

# ---------------------------------------------------------
# Decisão para remoção de hierarquia
# ---------------------------------------------------------
hierarchy_removal_decision := true {
    # Verificar método HTTP
    input.http_method == "DELETE"
    
    # Super Admin pode remover hierarquia de qualquer função em qualquer tenant
    common.has_role(constants.super_admin_role)
}

hierarchy_removal_decision := true {
    # Verificar método HTTP
    input.http_method == "DELETE"
    
    # Tenant Admin pode remover hierarquia de funções customizadas no seu tenant
    common.has_role(constants.tenant_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Obter dados da hierarquia
    parent_id := input.resource.parent_id
    child_id := input.resource.child_id
    
    # Obter dados das funções
    parent_role := get_role_by_id(parent_id)
    child_role := get_role_by_id(child_id)
    
    # Verificar se ambas são funções customizadas
    parent_role.type == constants.custom_role_type
    child_role.type == constants.custom_role_type
}

hierarchy_removal_decision := true {
    # Verificar método HTTP
    input.http_method == "DELETE"
    
    # IAM Admin pode remover hierarquia de funções customizadas no seu tenant
    common.has_role(constants.iam_admin_role)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
    
    # Verificar se o usuário tem permissão específica
    common.has_permission("role:manage_hierarchy")
    
    # Obter dados da hierarquia
    parent_id := input.resource.parent_id
    child_id := input.resource.child_id
    
    # Obter dados das funções
    parent_role := get_role_by_id(parent_id)
    child_role := get_role_by_id(child_id)
    
    # Verificar se ambas são funções customizadas
    parent_role.type == constants.custom_role_type
    child_role.type == constants.custom_role_type
}

# ---------------------------------------------------------
# Decisão para consulta de hierarquia
# ---------------------------------------------------------
hierarchy_query_decision := true {
    # Verificar método HTTP
    input.http_method == "GET"
    
    # Super Admin pode consultar hierarquia de qualquer função em qualquer tenant
    common.has_role(constants.super_admin_role)
}

hierarchy_query_decision := true {
    # Verificar método HTTP
    input.http_method == "GET"
    
    # Qualquer usuário com uma das funções administrativas pode consultar hierarquia no seu tenant
    admin_roles := {constants.tenant_admin_role, constants.iam_admin_role, constants.iam_operator_role}
    common.has_any_role(admin_roles)
    
    # Verificar se o tenant da requisição é o mesmo do usuário
    input.tenant_id == input.user.tenant_id
}

# ---------------------------------------------------------
# Funções auxiliares para validação de hierarquia
# ---------------------------------------------------------

# Função para verificar se a adição de uma relação criaria um ciclo
would_create_hierarchy_cycle(parent_id, child_id) {
    # Se o filho é igual ao pai, seria um ciclo direto
    parent_id == child_id
}

would_create_hierarchy_cycle(parent_id, child_id) {
    # Verifica recursivamente se o pai já é filho do filho em algum nível
    # (isto requer dados de hierarquia existentes)
    # Em um ambiente real, isso seria verificado com base nos dados do banco
    
    # Simulação: Supor que já existe uma relação child_id -> X -> parent_id
    # que tornaria parent_id -> child_id um ciclo
    
    parent_id == "role1"
    child_id == "role3"
}

# Função para verificar se a adição excederia a profundidade máxima de hierarquia
would_exceed_max_hierarchy_depth(parent_id, child_id) {
    # Em um ambiente real, isso calcularia a profundidade atual da árvore
    # e verificaria se adicionar esta relação excederia o limite
    
    # Simulação: Supor que algumas relações específicas excederiam o limite
    parent_id == "roleA"
    child_id == "roleD"
}

# Função auxiliar para obter dados de uma função pelo ID (simulação)
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