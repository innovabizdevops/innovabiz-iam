# role_crud.rego
# Políticas detalhadas para operações CRUD no RoleService
# Conformidade: ISO/IEC 27001, TOGAF 10.0, COBIT 2019, PCI DSS, Basel III
package innovabiz.iam.role.crud

import data.innovabiz.iam.common
import data.innovabiz.iam.role.permissions
import future.keywords

# === Política para Criação de Funções ===

# Regra padrão: negar criação
default allow_create = false

# Permite criação para administradores do sistema e do tenant
allow_create {
    # Super administradores podem criar qualquer tipo de função
    common.has_role(input.user, "SUPER_ADMIN")
}

allow_create {
    # Administradores de tenant podem criar apenas funções customizadas em seu tenant
    common.has_role(input.user, "TENANT_ADMIN")
    common.same_tenant(input.user.tenant_id, input.resource.tenant_id)
    input.resource.data.type == "CUSTOM" # Apenas funções customizadas
}

allow_create {
    # Administradores IAM podem criar apenas funções customizadas em seu tenant
    common.has_role(input.user, "IAM_ADMIN")
    common.same_tenant(input.user.tenant_id, input.resource.tenant_id)
    input.resource.data.type == "CUSTOM" # Apenas funções customizadas
}

allow_create {
    # Usuários com permissão específica podem criar funções customizadas
    common.has_permission(input.user, "role:create")
    common.same_tenant(input.user.tenant_id, input.resource.tenant_id)
    input.resource.data.type == "CUSTOM" # Apenas funções customizadas
}

# Validações de conteúdo para criação de função
valid_role_creation {
    # Nome da função é obrigatório e não deve ser vazio
    input.resource.data.name != null
    count(input.resource.data.name) > 0
    count(input.resource.data.name) <= 100 # Tamanho máximo de 100 caracteres
    
    # Nome de função não deve conter caracteres especiais (apenas alfanuméricos, espaço e underscore)
    regex.match("^[a-zA-Z0-9_\\s]+$", input.resource.data.name)
    
    # Tipo da função deve ser válido
    input.resource.data.type == "SYSTEM" or input.resource.data.type == "CUSTOM"
    
    # Se metadata estiver presente, deve ser um objeto válido
    not input.resource.data.metadata != null or is_object(input.resource.data.metadata)
}

# Verificação de nomes reservados (nomes de funções do sistema)
is_reserved_name {
    reserved_names := [
        "SUPER_ADMIN", 
        "TENANT_ADMIN", 
        "IAM_ADMIN", 
        "IAM_OPERATOR", 
        "IAM_AUDITOR",
        "SYSTEM", 
        "ADMIN"
    ]
    
    upper_name := upper(input.resource.data.name)
    some reserved in reserved_names
    upper_name == reserved
}

# Verifica se já existe uma função com o mesmo nome no tenant
role_name_exists {
    # Na implementação real, isso consultaria o banco de dados
    # Para políticas, usamos dados simulados
    some existing_role in data.roles
    existing_role.tenant_id == input.resource.tenant_id
    existing_role.name == input.resource.data.name
    existing_role.status != "DELETED"
}

# Validação final para criação de função
validate_role_creation = response {
    not valid_role_creation
    response := {
        "valid": false,
        "reason": "dados da função inválidos: nome obrigatório e deve conter apenas caracteres válidos"
    }
} else = response {
    input.resource.data.type == "CUSTOM"
    is_reserved_name
    response := {
        "valid": false,
        "reason": "nome da função é reservado para uso do sistema"
    }
} else = response {
    role_name_exists
    response := {
        "valid": false,
        "reason": "já existe uma função com este nome neste tenant"
    }
} else = response {
    response := {
        "valid": true
    }
}

# Decisão final para criação de função
create_decision = decision {
    not allow_create
    decision := {
        "allow": false,
        "reason": "permissão insuficiente para criar funções"
    }
} else = decision {
    validation := validate_role_creation
    not validation.valid
    decision := {
        "allow": false,
        "reason": validation.reason
    }
} else = decision {
    decision := {
        "allow": true,
        "reason": "criação de função permitida"
    }
}

# === Política para Leitura de Funções ===

# Regra padrão: negar leitura
default allow_read = false

# Permite leitura para diversos perfis
allow_read {
    # Super administradores podem ler qualquer função
    common.has_role(input.user, "SUPER_ADMIN")
}

allow_read {
    # Administradores e operadores de tenant podem ler funções em seu tenant
    common.has_any_role(input.user, ["TENANT_ADMIN", "IAM_ADMIN", "IAM_OPERATOR", "IAM_AUDITOR"])
    common.same_tenant(input.user.tenant_id, input.resource.tenant_id)
}

allow_read {
    # Usuários com permissão específica podem ler funções em seu tenant
    common.has_permission(input.user, "role:read")
    common.same_tenant(input.user.tenant_id, input.resource.tenant_id)
}

# Decisão final para leitura de função
read_decision = decision {
    allow_read
    decision := {
        "allow": true,
        "reason": "leitura de função permitida"
    }
} else = decision {
    decision := {
        "allow": false,
        "reason": "permissão insuficiente para ler funções"
    }
}

# === Política para Atualização de Funções ===

# Regra padrão: negar atualização
default allow_update = false

# Permite atualização para administradores do sistema e do tenant
allow_update {
    # Super administradores podem atualizar qualquer tipo de função
    common.has_role(input.user, "SUPER_ADMIN")
}

allow_update {
    # Administradores de tenant podem atualizar apenas funções customizadas em seu tenant
    common.has_role(input.user, "TENANT_ADMIN")
    common.same_tenant(input.user.tenant_id, input.resource.tenant_id)
    input.resource.current.type == "CUSTOM" # Apenas funções customizadas
}

allow_update {
    # Administradores IAM podem atualizar apenas funções customizadas em seu tenant
    common.has_role(input.user, "IAM_ADMIN")
    common.same_tenant(input.user.tenant_id, input.resource.tenant_id)
    input.resource.current.type == "CUSTOM" # Apenas funções customizadas
}

allow_update {
    # Usuários com permissão específica podem atualizar funções customizadas
    common.has_permission(input.user, "role:update")
    common.same_tenant(input.user.tenant_id, input.resource.tenant_id)
    input.resource.current.type == "CUSTOM" # Apenas funções customizadas
}

# Validações para atualização de função
valid_role_update {
    # Não pode mudar o tipo da função
    input.resource.data.type == input.resource.current.type
    
    # Se o nome for fornecido, deve ser válido
    not input.resource.data.name != null or count(input.resource.data.name) > 0
    not input.resource.data.name != null or count(input.resource.data.name) <= 100
    not input.resource.data.name != null or regex.match("^[a-zA-Z0-9_\\s]+$", input.resource.data.name)
    
    # Se metadata estiver presente, deve ser um objeto válido
    not input.resource.data.metadata != null or is_object(input.resource.data.metadata)
}

# Verifica colisão de nomes na atualização
update_name_collision {
    # Nome está sendo alterado
    input.resource.data.name != null
    input.resource.data.name != input.resource.current.name
    
    # Verifica se o novo nome já existe
    some existing_role in data.roles
    existing_role.tenant_id == input.resource.tenant_id
    existing_role.id != input.resource.id
    existing_role.name == input.resource.data.name
    existing_role.status != "DELETED"
}

# Validação final para atualização
validate_role_update = response {
    not valid_role_update
    response := {
        "valid": false,
        "reason": "dados de atualização inválidos: não é permitido alterar o tipo da função"
    }
} else = response {
    update_name_collision
    response := {
        "valid": false,
        "reason": "já existe outra função com este nome neste tenant"
    }
} else = response {
    input.resource.data.name != null
    is_reserved_name
    input.resource.current.name != input.resource.data.name
    response := {
        "valid": false,
        "reason": "nome da função é reservado para uso do sistema"
    }
} else = response {
    response := {
        "valid": true
    }
}

# Decisão final para atualização
update_decision = decision {
    not allow_update
    decision := {
        "allow": false,
        "reason": "permissão insuficiente para atualizar esta função"
    }
} else = decision {
    validation := validate_role_update
    not validation.valid
    decision := {
        "allow": false,
        "reason": validation.reason
    }
} else = decision {
    decision := {
        "allow": true,
        "reason": "atualização de função permitida"
    }
}

# === Política para Exclusão de Funções ===

# Regra padrão: negar exclusão
default allow_delete = false

# Permite exclusão para administradores do sistema e do tenant
allow_delete {
    # Super administradores podem excluir qualquer função customizada
    common.has_role(input.user, "SUPER_ADMIN")
    input.resource.current.type == "CUSTOM" # Apenas funções customizadas
}

allow_delete {
    # Administradores de tenant podem excluir apenas funções customizadas em seu tenant
    common.has_role(input.user, "TENANT_ADMIN")
    common.same_tenant(input.user.tenant_id, input.resource.tenant_id)
    input.resource.current.type == "CUSTOM" # Apenas funções customizadas
}

allow_delete {
    # Administradores IAM podem excluir apenas funções customizadas em seu tenant
    common.has_role(input.user, "IAM_ADMIN")
    common.same_tenant(input.user.tenant_id, input.resource.tenant_id)
    input.resource.current.type == "CUSTOM" # Apenas funções customizadas
}

allow_delete {
    # Usuários com permissão específica podem excluir funções customizadas
    common.has_permission(input.user, "role:delete")
    common.same_tenant(input.user.tenant_id, input.resource.tenant_id)
    input.resource.current.type == "CUSTOM" # Apenas funções customizadas
}

# Verifica se a função está sendo usada
role_in_use {
    # Na implementação real, isso consultaria relacionamentos no banco de dados
    # Simulando: verifica se a função tem usuários atribuídos
    some assignment in data.user_roles
    assignment.role_id == input.resource.id
    assignment.status == "ACTIVE"
}

role_in_use {
    # Verifica se a função é pai em alguma hierarquia
    some hierarchy in data.role_hierarchies
    hierarchy.parent_id == input.resource.id
}

# Validação final para exclusão
validate_role_deletion = response {
    input.resource.current.type == "SYSTEM"
    response := {
        "valid": false,
        "reason": "funções do sistema não podem ser excluídas"
    }
} else = response {
    role_in_use
    input.force != true
    response := {
        "valid": false,
        "reason": "a função está em uso e não pode ser excluída sem a flag force=true"
    }
} else = response {
    response := {
        "valid": true
    }
}

# Decisão final para exclusão
delete_decision = decision {
    not allow_delete
    decision := {
        "allow": false,
        "reason": "permissão insuficiente para excluir esta função"
    }
} else = decision {
    validation := validate_role_deletion
    not validation.valid
    decision := {
        "allow": false,
        "reason": validation.reason
    }
} else = decision {
    decision := {
        "allow": true,
        "reason": "exclusão de função permitida"
    }
}

# === Política para Exclusão Permanente (Hard Delete) ===

# Regra padrão: negar exclusão permanente
default allow_hard_delete = false

# Apenas super administradores podem realizar exclusão permanente
allow_hard_delete {
    common.has_role(input.user, "SUPER_ADMIN")
    input.resource.current.type == "CUSTOM" # Apenas funções customizadas
}

allow_hard_delete {
    # Administradores IAM com permissão específica
    common.has_role(input.user, "IAM_ADMIN")
    common.has_permission(input.user, "role:hard_delete")
    common.same_tenant(input.user.tenant_id, input.resource.tenant_id)
    input.resource.current.type == "CUSTOM" # Apenas funções customizadas
}

# Decisão final para hard delete
hard_delete_decision = decision {
    not allow_hard_delete
    decision := {
        "allow": false,
        "reason": "permissão insuficiente para excluir permanentemente esta função"
    }
} else = decision {
    validation := validate_role_deletion
    not validation.valid
    decision := {
        "allow": false,
        "reason": validation.reason
    }
} else = decision {
    decision := {
        "allow": true,
        "reason": "exclusão permanente de função permitida"
    }
}