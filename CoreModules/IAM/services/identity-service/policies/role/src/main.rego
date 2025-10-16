# Política principal do RoleService - INNOVABIZ Platform
#
# Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
# PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III
package innovabiz.iam.role

import data.innovabiz.iam.role.crud
import data.innovabiz.iam.role.permissions
import data.innovabiz.iam.role.hierarchy
import data.innovabiz.iam.role.user_assignment
import data.innovabiz.iam.role.common
import data.innovabiz.iam.role.constants
import data.innovabiz.iam.role.audit

# Decisão principal de autorização
default allow = false

# Regra geral de autorização baseada no caminho da decisão
allow {
    # Validações comuns para todas as requisições
    common.is_authenticated
    common.is_valid_request
    common.has_valid_tenant
    
    # Não está bloqueado por regras de proteção
    not common.is_rate_limited
    not common.is_outside_business_hours
    not common.is_from_suspicious_ip
    
    # Roteamento da decisão baseado no caminho especificado no input
    is_authorized_for_path
    
    # Gerar log de auditoria para a decisão
    audit_decision_id := uuid.rfc4122()
    audit_record := audit.audit_log(
        audit_decision_id,
        input.path,
        true,
        input.user,
        input.tenant_id,
        input.resource,
        input.context
    )
    
    # Registra a decisão de autorização (em um ambiente real, isso seria enviado para um sistema de logs)
    trace(sprintf("AUDIT: %v", [audit_record]))
}

# Determina se o usuário está autorizado com base no caminho da política
is_authorized_for_path {
    # Decisões CRUD
    input.path == "crud.create_decision"
    crud.create_decision
} else {
    input.path == "crud.read_decision"
    crud.read_decision
} else {
    input.path == "crud.update_decision"
    crud.update_decision
} else {
    input.path == "crud.delete_decision"
    crud.delete_decision
} else {
    input.path == "crud.permanent_delete_decision"
    crud.permanent_delete_decision
} else {
    input.path == "crud.list_decision"
    crud.list_decision
} else {
    # Decisões de permissões
    input.path == "permissions.permission_assignment_decision"
    permissions.permission_assignment_decision
} else {
    input.path == "permissions.permission_revocation_decision"
    permissions.permission_revocation_decision
} else {
    input.path == "permissions.permission_check_decision"
    permissions.permission_check_decision
} else {
    # Decisões de hierarquia
    input.path == "hierarchy.hierarchy_addition_decision"
    hierarchy.hierarchy_addition_decision
} else {
    input.path == "hierarchy.hierarchy_removal_decision"
    hierarchy.hierarchy_removal_decision
} else {
    input.path == "hierarchy.hierarchy_query_decision"
    hierarchy.hierarchy_query_decision
} else {
    # Decisões de atribuição de função a usuário
    input.path == "user_assignment.role_assignment_decision"
    user_assignment.role_assignment_decision
} else {
    input.path == "user_assignment.role_removal_decision"
    user_assignment.role_removal_decision
} else {
    input.path == "user_assignment.expiration_update_decision"
    user_assignment.expiration_update_decision
} else {
    input.path == "user_assignment.role_check_decision"
    user_assignment.role_check_decision
} else {
    input.path == "user_assignment.specific_role_check_decision"
    user_assignment.specific_role_check_decision
}

# Decisão para debug e diagnóstico
debug_info = {
    "user": input.user,
    "tenant_id": input.tenant_id,
    "resource": input.resource,
    "path": input.path,
    "http_method": input.http_method,
    "decision": allow,
    "timestamp": time.now_ns(),
    "decision_path": input.path,
    "reason": get_decision_reason()
}

# Função para determinar o motivo da decisão (usado para diagnóstico)
get_decision_reason() = reason {
    allow
    reason = "allowed"
} else = reason {
    not common.is_authenticated
    reason = "user_not_authenticated"
} else = reason {
    not common.is_valid_request
    reason = "invalid_request"
} else = reason {
    not common.has_valid_tenant
    reason = "invalid_tenant"
} else = reason {
    common.is_rate_limited
    reason = "rate_limited"
} else = reason {
    common.is_outside_business_hours
    reason = "outside_business_hours"
} else = reason {
    common.is_from_suspicious_ip
    reason = "suspicious_ip_address"
} else = reason {
    reason = sprintf("unauthorized_for_%v", [input.path])
}