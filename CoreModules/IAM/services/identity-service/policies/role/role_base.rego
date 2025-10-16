# role_base.rego
# Política base para operações do RoleService com controles avançados de segurança
# Conformidade: ISO/IEC 27001, TOGAF 10.0, COBIT 2019, PCI DSS, NIST Cybersecurity Framework
package innovabiz.iam.role

import data.innovabiz.iam.common
import future.keywords

# Constantes e configurações
permissions := {
    "role:create": "Permissão para criar funções",
    "role:read": "Permissão para ler funções",
    "role:update": "Permissão para atualizar funções",
    "role:delete": "Permissão para excluir funções (soft delete)",
    "role:hard_delete": "Permissão para excluir funções permanentemente",
    "role:list": "Permissão para listar funções",
    "role:clone": "Permissão para clonar funções",
    "role:sync": "Permissão para sincronizar funções do sistema",
    "role:assign_permission": "Permissão para atribuir permissões a funções",
    "role:revoke_permission": "Permissão para revogar permissões de funções",
    "role:check_permission": "Permissão para verificar permissões de funções",
    "role:add_child": "Permissão para adicionar funções filhas",
    "role:remove_child": "Permissão para remover funções filhas",
    "role:assign_to_user": "Permissão para atribuir funções a usuários",
    "role:remove_from_user": "Permissão para remover funções de usuários",
    "role:update_expiration": "Permissão para atualizar expiração de funções",
}

system_roles := {
    "SUPER_ADMIN": "Super administrador com acesso total",
    "TENANT_ADMIN": "Administrador do tenant",
    "IAM_ADMIN": "Administrador do módulo IAM",
    "IAM_OPERATOR": "Operador do módulo IAM",
    "IAM_AUDITOR": "Auditor do módulo IAM com acesso somente leitura",
}

# Regra padrão: negar acesso
default allow = false

# === Regras de Autorização Gerais ===

# Permite acesso para super administradores
allow {
    common.has_role(input.user, "SUPER_ADMIN")
}

# Permite acesso para administradores de tenant em seu próprio tenant
allow {
    common.has_role(input.user, "TENANT_ADMIN")
    common.same_tenant(input.user.tenant_id, input.resource.tenant_id)
}

# Permite acesso se o usuário tiver a permissão específica para a operação
allow {
    # Obter a permissão necessária baseada no método e recurso
    permission := get_required_permission(input.method, input.resource.type)
    
    # Verificar se o usuário tem a permissão
    common.has_permission(input.user, permission)
    
    # Verificar se o contexto de tenant é válido
    common.valid_tenant_context(input.user, input.resource)
}

# === Funções Auxiliares ===

# Mapeia métodos HTTP para permissões necessárias
get_required_permission(method, resource_type) = permission {
    mapping := {
        # Mapeamento para roles
        "GET/roles": "role:list",
        "GET/roles/{id}": "role:read",
        "POST/roles": "role:create",
        "PUT/roles/{id}": "role:update",
        "DELETE/roles/{id}": "role:delete",
        "DELETE/roles/{id}/hard": "role:hard_delete",
        "POST/roles/{id}/clone": "role:clone",
        "POST/roles/sync": "role:sync",
        
        # Mapeamento para permissões de roles
        "GET/roles/{id}/permissions": "role:read",
        "GET/roles/{id}/all-permissions": "role:read",
        "POST/roles/{id}/permissions/{permission_id}": "role:assign_permission",
        "DELETE/roles/{id}/permissions/{permission_id}": "role:revoke_permission",
        "GET/roles/{id}/permissions/{permission_id}/check": "role:check_permission",
        
        # Mapeamento para hierarquia
        "GET/roles/{id}/children": "role:read",
        "GET/roles/{id}/parents": "role:read",
        "GET/roles/{id}/descendants": "role:read",
        "GET/roles/{id}/ancestors": "role:read",
        "POST/roles/{id}/children/{child_id}": "role:add_child",
        "DELETE/roles/{id}/children/{child_id}": "role:remove_child",
        
        # Mapeamento para usuários
        "GET/roles/{id}/users": "role:read",
        "POST/roles/{id}/users/{user_id}": "role:assign_to_user",
        "DELETE/roles/{id}/users/{user_id}": "role:remove_from_user",
        "PUT/roles/{id}/users/{user_id}/expiration": "role:update_expiration",
        "GET/roles/{id}/users/{user_id}/check": "role:check_permission",
    }
    
    key := concat("/", [method, resource_type])
    permission = mapping[key]
}

# === Regras Específicas para Tipos de Roles ===

# Apenas super admins podem gerenciar roles do sistema
allow_system_role_management {
    input.resource.data.type == "SYSTEM"
    common.has_role(input.user, "SUPER_ADMIN")
}

deny_system_role_update {
    input.method == "PUT"
    input.resource.type == "roles"
    input.resource.data.type == "SYSTEM"
    not common.has_role(input.user, "SUPER_ADMIN")
}

deny_system_role_delete {
    input.method == "DELETE"
    input.resource.type == "roles"
    input.resource.data.type == "SYSTEM"
}

# === Regras de Auditoria e Compliance ===

# Registra tentativa de acesso (para auditoria)
log_access {
    timestamp := time.now_ns() / 1000000
    
    audit_event := {
        "timestamp": timestamp,
        "user_id": input.user.id,
        "tenant_id": input.user.tenant_id,
        "action": input.method,
        "resource_type": input.resource.type,
        "resource_id": input.resource.id,
        "allowed": allow,
        "reason": reason,
    }
    
    # Na implementação real, este evento seria enviado para um sistema de auditoria
    true
}

# Razão para a decisão de autorização (para explicabilidade)
reason = r {
    allow
    r := "permission granted"
} else = r {
    deny_system_role_update
    r := "system roles can only be updated by super admin"
} else = r {
    deny_system_role_delete
    r := "system roles cannot be deleted"
} else = r {
    r := "insufficient permissions"
}

# === Regras de Contexto e Segurança Avançada ===

# Verifica limites de taxa baseados no contexto
rate_limit_exceeded {
    # Implementação de exemplo - na prática, consultaria um serviço de rate limiting
    input.context.requests_per_minute > 100
}

# Verifica origem geográfica suspeita
suspicious_location {
    # Lista de países permitidos para o tenant específico
    allowed_countries := data.tenants[input.user.tenant_id].allowed_countries
    
    # Verifica se o país de origem está na lista permitida
    not array.contains(allowed_countries, input.context.geo.country)
}

# Bloqueia acesso baseado em análise de risco
block_based_on_risk {
    rate_limit_exceeded
}

block_based_on_risk {
    suspicious_location
}

# === Regra Final de Decisão ===

# A autorização final considera todas as regras de segurança
final_decision = decision {
    block_based_on_risk
    decision := {
        "allow": false,
        "reason": "blocked due to security risk",
    }
} else = decision {
    decision := {
        "allow": allow,
        "reason": reason,
    }
}

# Executa log e retorna a decisão final
evaluate_request = result {
    log_access
    result := final_decision
}