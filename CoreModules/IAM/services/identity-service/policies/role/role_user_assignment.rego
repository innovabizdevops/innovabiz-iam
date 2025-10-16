# role_user_assignment.rego
# Políticas de autorização para gestão de atribuição de funções a usuários
# Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53, PCI DSS v4.0
package innovabiz.iam.role.user_assignment

import data.innovabiz.iam.role.base
import future.keywords
import time

# ==========================================
# === Atribuição de Função a um Usuário ===
# ==========================================

# Decisão final para atribuição de função
role_assignment_decision := {
    "allow": allow,
    "reason": reason
} {
    # Computar decisão de autorização
    allow := role_assignment_allowed
    reason := role_assignment_reason
}

# Verifica se o usuário tem permissão para atribuir uma função a outro usuário
role_assignment_allowed {
    # Verificação de autenticação
    base.user_authenticated

    # Verificação de escopo de tenant
    base.tenant_scope_valid
    
    # Verificação de integridade dos dados
    base.valid_uuid(input.resource.role_id)
    base.valid_uuid(input.resource.user_id)
    
    # Acesso especial para super admin
    base.is_super_admin
    
    # Validações adicionais de dados e negócio
    valid_expiration_date
    not user_already_has_role
} else {
    # Verificação de autenticação
    base.user_authenticated

    # Verificação de escopo de tenant
    base.tenant_scope_valid
    
    # Verificação de integridade dos dados
    base.valid_uuid(input.resource.role_id)
    base.valid_uuid(input.resource.user_id)
    
    # Usuário tem permissão explícita para atribuir funções
    base.has_permission("role:assign_to_user")
    
    # A função existe e está ativa
    role_exists
    role_is_active
    
    # Usuários estão no mesmo tenant
    same_tenant_context
    
    # Validações de negócio
    not assigning_protected_role
    not assigning_to_self
    valid_expiration_date
    not user_already_has_role
    
    # Não é uma atribuição de risco elevado
    not is_high_risk_assignment
}

# Recupera a razão pela qual a atribuição de função foi permitida ou negada
role_assignment_reason = reason {
    not role_assignment_allowed
    reason := concat(" ", [
        get_reason_auth_failed,
        get_reason_tenant_mismatch,
        get_reason_role_status,
        get_reason_protected_role,
        get_reason_self_assignment,
        get_reason_expiration_date,
        get_reason_existing_assignment,
        get_reason_high_risk
    ])
} else {
    role_assignment_allowed
    reason := "atribuição de função autorizada"
}

# === Funções auxiliares para atribuição de função ===

# Verifica se a função existe
role_exists {
    role_id := input.resource.role_id
    _ = data.roles[role_id]
}

# Verifica se a função está ativa
role_is_active {
    role_id := input.resource.role_id
    role := data.roles[role_id]
    role.status == "ACTIVE"
}

# Verifica se os usuários estão no mesmo tenant
same_tenant_context {
    role_id := input.resource.role_id
    role := data.roles[role_id]
    role.tenant_id == input.resource.tenant_id
    role.tenant_id == input.user.tenant_id
}

# Verifica se é uma tentativa de atribuir uma função protegida
assigning_protected_role {
    role_id := input.resource.role_id
    role := data.roles[role_id]
    
    # Funções protegidas requerem permissões especiais
    protected_roles := [
        "SUPER_ADMIN",
        "TENANT_ADMIN",
        "IAM_ADMIN",
        "SECURITY_OFFICER",
        "COMPLIANCE_OFFICER",
        "AUDITOR",
        "SYSTEM_ADMIN"
    ]
    
    # Se a função for protegida e o usuário não for admin
    some i
    protected_roles[i] == role.name
    not (base.is_super_admin or base.is_tenant_admin)
}

# Verifica se o usuário está tentando atribuir uma função a si mesmo
assigning_to_self {
    input.resource.user_id == input.user.id
    
    # Atribuição a si mesmo não é permitida, exceto para super admin
    not base.is_super_admin
}

# Verifica se a data de expiração é válida
valid_expiration_date {
    # Se não houver data de expiração, é válido
    not input.resource.expires_at
} else {
    # Se houver data de expiração, verificar se está no futuro
    expires_at := time.parse_rfc3339_ns(input.resource.expires_at)
    now := time.now_ns()
    expires_at > now
}

# Verifica se o usuário já possui a função
user_already_has_role {
    role_id := input.resource.role_id
    user_id := input.resource.user_id
    
    # Verificar nas atribuições existentes
    some i
    user_role := data.user_roles[i]
    user_role.role_id == role_id
    user_role.user_id == user_id
    user_role.status == "ACTIVE"
}

# Verifica se é uma atribuição de alto risco
is_high_risk_assignment {
    role_id := input.resource.role_id
    user_id := input.resource.user_id
    role := data.roles[role_id]
    
    # Qualquer atribuição de função de sistema é considerada de alto risco
    role.type == "SYSTEM"
    
    # Requer aprovação elevada, a menos que seja super admin ou tenha permissão específica
    not base.is_super_admin
    not base.has_permission("role:assign_system_role")
}

# Razões para falha na atribuição de função
get_reason_auth_failed = msg {
    not base.user_authenticated
    msg := "usuário não autenticado"
} else = ""

get_reason_tenant_mismatch = msg {
    not base.tenant_scope_valid
    msg := "escopo de tenant inválido"
} else = msg {
    role_exists
    not same_tenant_context
    msg := "usuário não pode atribuir funções em outro tenant"
} else = ""

get_reason_role_status = msg {
    role_exists
    not role_is_active
    msg := "apenas funções ativas podem ser atribuídas"
} else = ""

get_reason_protected_role = msg {
    assigning_protected_role
    msg := "esta função só pode ser atribuída por administradores"
} else = ""

get_reason_self_assignment = msg {
    assigning_to_self
    msg := "usuários não podem atribuir funções a si mesmos"
} else = ""

get_reason_expiration_date = msg {
    not valid_expiration_date
    msg := "a data de expiração deve estar no futuro"
} else = ""

get_reason_existing_assignment = msg {
    user_already_has_role
    msg := "esta função já está atribuída a este utilizador"
} else = ""

get_reason_high_risk = msg {
    is_high_risk_assignment
    msg := "esta atribuição é considerada de alto risco e requer aprovação elevada"
} else = ""

# ==========================================
# === Remoção de Função de um Usuário ===
# ==========================================

# Decisão final para remoção de função
role_removal_decision := {
    "allow": allow,
    "reason": reason
} {
    # Computar decisão de autorização
    allow := role_removal_allowed
    reason := role_removal_reason
}

# Verifica se o usuário tem permissão para remover uma função de outro usuário
role_removal_allowed {
    # Verificação de autenticação
    base.user_authenticated

    # Verificação de escopo de tenant
    base.tenant_scope_valid
    
    # Verificação de integridade dos dados
    base.valid_uuid(input.resource.user_role_id)
    
    # Acesso especial para super admin
    base.is_super_admin
} else {
    # Verificação de autenticação
    base.user_authenticated

    # Verificação de escopo de tenant
    base.tenant_scope_valid
    
    # Verificação de integridade dos dados
    base.valid_uuid(input.resource.user_role_id)
    
    # Usuário tem permissão explícita para gerenciar atribuições
    base.has_permission("role:remove_from_user")
    
    # A atribuição existe
    assignment_exists
    
    # Usuário está no mesmo tenant da atribuição
    assignment_in_user_tenant
    
    # Não é uma atribuição de papel protegido
    not is_protected_role_assignment
    
    # Não é uma auto-remoção
    not removing_own_role
}

# Recupera a razão pela qual a remoção de função foi permitida ou negada
role_removal_reason = reason {
    not role_removal_allowed
    reason := concat(" ", [
        get_reason_auth_failed,
        get_reason_assignment_not_exists,
        get_reason_tenant_mismatch_removal,
        get_reason_protected_assignment,
        get_reason_self_removal
    ])
} else {
    role_removal_allowed
    reason := "remoção de função autorizada"
}

# === Funções auxiliares para remoção de função ===

# Verifica se a atribuição existe
assignment_exists {
    user_role_id := input.resource.user_role_id
    
    some i
    user_role := data.user_roles[i]
    user_role.id == user_role_id
}

# Verifica se o usuário está no mesmo tenant da atribuição
assignment_in_user_tenant {
    user_role_id := input.resource.user_role_id
    
    # Encontrar a atribuição
    some i
    user_role := data.user_roles[i]
    user_role.id == user_role_id
    
    # Verificar se o tenant da atribuição é o mesmo do usuário
    user_role.tenant_id == input.user.tenant_id
}

# Verifica se é uma atribuição de função protegida
is_protected_role_assignment {
    user_role_id := input.resource.user_role_id
    
    # Encontrar a atribuição
    some i
    user_role := data.user_roles[i]
    user_role.id == user_role_id
    
    # Obter a função
    role := data.roles[user_role.role_id]
    
    # Lista de funções protegidas que requerem permissões especiais para remover
    protected_roles := [
        "SUPER_ADMIN",
        "TENANT_ADMIN",
        "IAM_ADMIN",
        "SECURITY_OFFICER"
    ]
    
    # Verificar se a função está na lista de protegidas
    some j
    protected_roles[j] == role.name
    
    # Não é super admin
    not base.is_super_admin
}

# Verifica se o usuário está tentando remover uma de suas próprias funções
removing_own_role {
    user_role_id := input.resource.user_role_id
    
    # Encontrar a atribuição
    some i
    user_role := data.user_roles[i]
    user_role.id == user_role_id
    
    # Verificar se o usuário da atribuição é o mesmo que está fazendo a solicitação
    user_role.user_id == input.user.id
}

# Razões para falha na remoção de função
get_reason_assignment_not_exists = msg {
    not assignment_exists
    msg := "a atribuição de função especificada não existe"
} else = ""

get_reason_tenant_mismatch_removal = msg {
    assignment_exists
    not assignment_in_user_tenant
    msg := "usuário não tem permissão para gerenciar atribuições em outro tenant"
} else = ""

get_reason_protected_assignment = msg {
    is_protected_role_assignment
    msg := "apenas administradores podem remover atribuições de funções protegidas"
} else = ""

get_reason_self_removal = msg {
    removing_own_role
    msg := "usuários não podem remover suas próprias funções"
} else = ""

# ==========================================
# === Atualização de Expiração de Atribuição ===
# ==========================================

# Decisão final para atualização de expiração
expiration_update_decision := {
    "allow": allow,
    "reason": reason
} {
    # Computar decisão de autorização
    allow := expiration_update_allowed
    reason := expiration_update_reason
}

# Verifica se o usuário tem permissão para atualizar a expiração de uma atribuição
expiration_update_allowed {
    # Verificação de autenticação
    base.user_authenticated

    # Verificação de escopo de tenant
    base.tenant_scope_valid
    
    # Verificação de integridade dos dados
    base.valid_uuid(input.resource.user_role_id)
    
    # Validação da data de expiração
    valid_update_expiration
    
    # Acesso especial para super admin
    base.is_super_admin
} else {
    # Verificação de autenticação
    base.user_authenticated

    # Verificação de escopo de tenant
    base.tenant_scope_valid
    
    # Verificação de integridade dos dados
    base.valid_uuid(input.resource.user_role_id)
    
    # Validação da data de expiração
    valid_update_expiration
    
    # Usuário tem permissão explícita para gerenciar atribuições
    base.has_permission("role:update_assignment")
    
    # A atribuição existe
    update_assignment_exists
    
    # Usuário está no mesmo tenant da atribuição
    update_assignment_in_user_tenant
    
    # Não é uma atribuição de função protegida
    not is_update_protected_assignment
}

# Recupera a razão pela qual a atualização de expiração foi permitida ou negada
expiration_update_reason = reason {
    not expiration_update_allowed
    reason := concat(" ", [
        get_reason_auth_failed,
        get_reason_update_assignment_not_exists,
        get_reason_tenant_mismatch_update,
        get_reason_protected_assignment_update,
        get_reason_invalid_expiration
    ])
} else {
    expiration_update_allowed
    reason := "atualização de expiração autorizada"
}

# === Funções auxiliares para atualização de expiração ===

# Verifica se a atribuição existe
update_assignment_exists {
    user_role_id := input.resource.user_role_id
    
    some i
    user_role := data.user_roles[i]
    user_role.id == user_role_id
}

# Verifica se o usuário está no mesmo tenant da atribuição
update_assignment_in_user_tenant {
    user_role_id := input.resource.user_role_id
    
    # Encontrar a atribuição
    some i
    user_role := data.user_roles[i]
    user_role.id == user_role_id
    
    # Verificar se o tenant da atribuição é o mesmo do usuário
    user_role.tenant_id == input.user.tenant_id
}

# Verifica se é uma atribuição de função protegida
is_update_protected_assignment {
    user_role_id := input.resource.user_role_id
    
    # Encontrar a atribuição
    some i
    user_role := data.user_roles[i]
    user_role.id == user_role_id
    
    # Obter a função
    role := data.roles[user_role.role_id]
    
    # Lista de funções protegidas
    protected_roles := [
        "SUPER_ADMIN",
        "TENANT_ADMIN",
        "IAM_ADMIN"
    ]
    
    # Verificar se a função está na lista de protegidas
    some j
    protected_roles[j] == role.name
    
    # Não é super admin nem admin do tenant
    not (base.is_super_admin or base.is_tenant_admin)
}

# Verifica se a nova data de expiração é válida
valid_update_expiration {
    # Se não houver data de expiração, é uma remoção de expiração (permitido)
    not input.resource.expires_at
} else {
    # Se houver data de expiração, verificar se está no futuro
    expires_at := time.parse_rfc3339_ns(input.resource.expires_at)
    now := time.now_ns()
    expires_at > now
}

# Razões para falha na atualização de expiração
get_reason_update_assignment_not_exists = msg {
    not update_assignment_exists
    msg := "a atribuição de função especificada não existe"
} else = ""

get_reason_tenant_mismatch_update = msg {
    update_assignment_exists
    not update_assignment_in_user_tenant
    msg := "usuário não tem permissão para gerenciar atribuições em outro tenant"
} else = ""

get_reason_protected_assignment_update = msg {
    is_update_protected_assignment
    msg := "apenas administradores podem atualizar atribuições de funções protegidas"
} else = ""

get_reason_invalid_expiration = msg {
    not valid_update_expiration
    msg := "a data de expiração deve estar no futuro"
} else = ""

# ==========================================
# === Verificação de Função de um Usuário ===
# ==========================================

# Decisão final para verificação de função
role_check_decision := {
    "allow": allow,
    "reason": reason,
    "data": result
} {
    # Computar decisão de autorização
    allow := role_check_allowed
    reason := role_check_reason
    result := role_check_data
}

# Verifica se o usuário pode verificar as funções de outro usuário
role_check_allowed {
    # Verificação de autenticação
    base.user_authenticated

    # Verificação de escopo de tenant
    base.tenant_scope_valid
    
    # Verificação de integridade dos dados
    base.valid_uuid(input.resource.user_id)
    
    # Super admin pode verificar funções de qualquer usuário
    base.is_super_admin
} else {
    # Verificação de autenticação
    base.user_authenticated

    # Verificação de escopo de tenant
    base.tenant_scope_valid
    
    # Verificação de integridade dos dados
    base.valid_uuid(input.resource.user_id)
    
    # Usuário tem permissão para verificar funções ou é o próprio usuário
    base.has_permission("role:check_user_roles") or input.resource.user_id == input.user.id
    
    # O usuário alvo existe
    target_user_exists
    
    # Ambos os usuários estão no mesmo tenant (exceto se o próprio usuário)
    same_tenant_as_target_user
}

# Recupera a razão pela qual a verificação de função foi permitida ou negada
role_check_reason = reason {
    not role_check_allowed
    reason := concat(" ", [
        get_reason_auth_failed,
        get_reason_user_not_exists,
        get_reason_tenant_mismatch_check
    ])
} else {
    role_check_allowed
    reason := "verificação de função autorizada"
}

# Dados retornados pela verificação de função
role_check_data = result {
    role_check_allowed
    
    user_id := input.resource.user_id
    
    # Recuperar todas as atribuições de função ativas para o usuário
    user_roles := [role_data |
        some i
        user_role := data.user_roles[i]
        user_role.user_id == user_id
        user_role.status == "ACTIVE"
        
        # Se a atribuição tiver expirado, ignorar
        not is_expired_assignment(user_role)
        
        role := data.roles[user_role.role_id]
        
        role_data := {
            "id": role.id,
            "name": role.name,
            "type": role.type,
            "status": role.status,
            "assignment_id": user_role.id,
            "expires_at": user_role.expires_at
        }
    ]
    
    # Filtrar dados sensíveis para usuários não autorizados
    can_see_all := base.is_super_admin || base.is_tenant_admin || base.has_permission("role:view_all_user_roles")
    
    result := {
        "user_id": user_id,
        "roles": user_roles,
        "count": count(user_roles),
        "filtered": not can_see_all
    }
}

# === Funções auxiliares para verificação de função ===

# Verifica se o usuário alvo existe
target_user_exists {
    user_id := input.resource.user_id
    _ = data.users[user_id]
}

# Verifica se o usuário atual está no mesmo tenant que o usuário alvo
same_tenant_as_target_user {
    # Se for o próprio usuário, está no mesmo tenant
    input.resource.user_id == input.user.id
} else {
    # Caso contrário, verificar tenant
    user_id := input.resource.user_id
    target_user := data.users[user_id]
    target_user.tenant_id == input.user.tenant_id
}

# Verifica se uma atribuição está expirada
is_expired_assignment(user_role) {
    # Se não houver data de expiração, não está expirado
    not user_role.expires_at
    false
} else {
    # Se houver data de expiração, verificar se já passou
    expires_at := time.parse_rfc3339_ns(user_role.expires_at)
    now := time.now_ns()
    expires_at <= now
}

# Razões para falha na verificação de função
get_reason_user_not_exists = msg {
    not target_user_exists
    msg := "o usuário especificado não existe"
} else = ""

get_reason_tenant_mismatch_check = msg {
    target_user_exists
    not same_tenant_as_target_user
    msg := "usuário não pode verificar funções de usuários em outro tenant"
} else = ""