# Políticas para auditoria de decisões de autorização - INNOVABIZ Platform
#
# Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
# PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III
package innovabiz.iam.role.audit

import data.innovabiz.iam.role.common
import data.innovabiz.iam.role.constants

# ---------------------------------------------------------
# Decisões de auditoria para operações de funções
# ---------------------------------------------------------

# Gera metadados de auditoria para cada decisão
audit_log(decision_id, path, allow, user, tenant_id, resource, context) = log {
    timestamp := time.now_ns()
    
    log := {
        "id": decision_id,
        "timestamp": timestamp,
        "timestamp_formatted": time.format(timestamp, "2006-01-02T15:04:05Z07:00"),
        "decision": {
            "path": path,
            "allow": allow
        },
        "user": user_info(user),
        "tenant_id": tenant_id,
        "resource": resource,
        "context": context,
        "compliance": compliance_info(path),
        "tags": ["IAM", "RoleService", "Authorization"]
    }
}

# Informações do usuário para auditoria
user_info(user) = info {
    info := {
        "id": user.id,
        "email": user.email,
        "roles": user.roles,
        "tenant_id": user.tenant_id,
        "ip_address": user.ip_address,
        "user_agent": user.user_agent
    }
}

# Informações de conformidade para auditoria
compliance_info(decision_path) = info {
    info := {
        "standards": [
            "ISO/IEC 27001:2022",
            "TOGAF 10.0",
            "COBIT 2019",
            "NIST SP 800-53",
            "PCI DSS v4.0",
            "GDPR",
            "APD Angola",
            "BNA"
        ],
        "controls": get_controls_for_decision(decision_path)
    }
}

# Obtém controles aplicáveis para um determinado tipo de decisão
get_controls_for_decision(decision_path) = controls {
    # Decisões CRUD
    decision_path == "crud.create_decision"
    controls := [
        "AC-2 Account Management",
        "AC-3 Access Enforcement",
        "AC-6 Least Privilege",
        "AU-2 Audit Events",
        "AU-6 Audit Review, Analysis, and Reporting",
        "CM-3 Configuration Change Control",
        "ISO 27001 A.9.2.3 Management of privileged access rights",
        "COBIT DSS06.03 Manage roles, responsibilities, access privileges",
        "PCI DSS 7.1.2 Role-based access control"
    ]
} else = controls {
    decision_path == "crud.read_decision"
    controls := [
        "AC-3 Access Enforcement",
        "AC-6 Least Privilege",
        "AU-2 Audit Events",
        "ISO 27001 A.9.4.1 Information access restriction",
        "COBIT DSS06.03 Manage roles, responsibilities, access privileges"
    ]
} else = controls {
    decision_path == "crud.update_decision"
    controls := [
        "AC-2 Account Management",
        "AC-3 Access Enforcement",
        "AC-6 Least Privilege",
        "AU-2 Audit Events",
        "AU-6 Audit Review, Analysis, and Reporting",
        "CM-3 Configuration Change Control",
        "ISO 27001 A.9.2.3 Management of privileged access rights",
        "COBIT DSS06.03 Manage roles, responsibilities, access privileges",
        "PCI DSS 7.1.2 Role-based access control"
    ]
} else = controls {
    decision_path == "crud.delete_decision"
    controls := [
        "AC-2 Account Management",
        "AC-3 Access Enforcement",
        "AC-6 Least Privilege",
        "AU-2 Audit Events",
        "AU-6 Audit Review, Analysis, and Reporting",
        "CM-3 Configuration Change Control",
        "ISO 27001 A.9.2.6 Removal or adjustment of access rights",
        "COBIT DSS06.03 Manage roles, responsibilities, access privileges",
        "PCI DSS 7.1.4 Removal of access"
    ]
} else = controls {
    decision_path == "crud.permanent_delete_decision"
    controls := [
        "AC-2 Account Management",
        "AC-3 Access Enforcement",
        "AC-6(7) Least Privilege | Review of User Privileges",
        "AU-2 Audit Events",
        "AU-6 Audit Review, Analysis, and Reporting",
        "AU-12 Audit Generation",
        "CM-3 Configuration Change Control",
        "ISO 27001 A.9.2.6 Removal or adjustment of access rights",
        "COBIT DSS06.03 Manage roles, responsibilities, access privileges",
        "PCI DSS 7.1.4 Removal of access",
        "GDPR Article 17 Right to erasure"
    ]
} else = controls {
    decision_path == "crud.list_decision"
    controls := [
        "AC-3 Access Enforcement",
        "AC-6 Least Privilege",
        "AU-2 Audit Events",
        "ISO 27001 A.9.4.1 Information access restriction",
        "COBIT DSS06.03 Manage roles, responsibilities, access privileges"
    ]
} else = controls {
    # Decisões de permissões
    decision_path == "permissions.permission_assignment_decision"
    controls := [
        "AC-2 Account Management",
        "AC-3 Access Enforcement",
        "AC-6 Least Privilege",
        "AU-2 Audit Events",
        "AU-6 Audit Review, Analysis, and Reporting",
        "CM-3 Configuration Change Control",
        "ISO 27001 A.9.2.3 Management of privileged access rights",
        "COBIT DSS06.03 Manage roles, responsibilities, access privileges",
        "PCI DSS 7.1.2 Role-based access control",
        "PCI DSS 7.2.1 Access control systems"
    ]
} else = controls {
    decision_path == "permissions.permission_revocation_decision"
    controls := [
        "AC-2 Account Management",
        "AC-3 Access Enforcement",
        "AC-6 Least Privilege",
        "AU-2 Audit Events",
        "AU-6 Audit Review, Analysis, and Reporting",
        "CM-3 Configuration Change Control",
        "ISO 27001 A.9.2.6 Removal or adjustment of access rights",
        "COBIT DSS06.03 Manage roles, responsibilities, access privileges",
        "PCI DSS 7.1.4 Removal of access"
    ]
} else = controls {
    decision_path == "permissions.permission_check_decision"
    controls := [
        "AC-3 Access Enforcement",
        "AC-6 Least Privilege",
        "AU-2 Audit Events",
        "ISO 27001 A.9.4.1 Information access restriction",
        "COBIT DSS06.03 Manage roles, responsibilities, access privileges"
    ]
} else = controls {
    # Decisões de hierarquia
    decision_path == "hierarchy.hierarchy_addition_decision"
    controls := [
        "AC-2 Account Management",
        "AC-3 Access Enforcement",
        "AC-6 Least Privilege",
        "AU-2 Audit Events",
        "AU-6 Audit Review, Analysis, and Reporting",
        "CM-3 Configuration Change Control",
        "ISO 27001 A.9.2.3 Management of privileged access rights",
        "COBIT DSS06.03 Manage roles, responsibilities, access privileges",
        "PCI DSS 7.1.2 Role-based access control"
    ]
} else = controls {
    decision_path == "hierarchy.hierarchy_removal_decision"
    controls := [
        "AC-2 Account Management",
        "AC-3 Access Enforcement",
        "AC-6 Least Privilege",
        "AU-2 Audit Events",
        "AU-6 Audit Review, Analysis, and Reporting",
        "CM-3 Configuration Change Control",
        "ISO 27001 A.9.2.6 Removal or adjustment of access rights",
        "COBIT DSS06.03 Manage roles, responsibilities, access privileges"
    ]
} else = controls {
    decision_path == "hierarchy.hierarchy_query_decision"
    controls := [
        "AC-3 Access Enforcement",
        "AC-6 Least Privilege",
        "AU-2 Audit Events",
        "ISO 27001 A.9.4.1 Information access restriction",
        "COBIT DSS06.03 Manage roles, responsibilities, access privileges"
    ]
} else = controls {
    # Decisões de atribuição de função a usuário
    decision_path == "user_assignment.role_assignment_decision"
    controls := [
        "AC-2 Account Management",
        "AC-3 Access Enforcement",
        "AC-6 Least Privilege",
        "AU-2 Audit Events",
        "AU-6 Audit Review, Analysis, and Reporting",
        "CM-3 Configuration Change Control",
        "ISO 27001 A.9.2.2 User access provisioning",
        "ISO 27001 A.9.2.3 Management of privileged access rights",
        "COBIT DSS06.03 Manage roles, responsibilities, access privileges",
        "PCI DSS 7.1.2 Role-based access control",
        "PCI DSS 7.2.1 Access control systems"
    ]
} else = controls {
    decision_path == "user_assignment.role_removal_decision"
    controls := [
        "AC-2 Account Management",
        "AC-3 Access Enforcement",
        "AC-6 Least Privilege",
        "AU-2 Audit Events",
        "AU-6 Audit Review, Analysis, and Reporting",
        "CM-3 Configuration Change Control",
        "ISO 27001 A.9.2.6 Removal or adjustment of access rights",
        "COBIT DSS06.03 Manage roles, responsibilities, access privileges",
        "PCI DSS 7.1.4 Removal of access"
    ]
} else = controls {
    decision_path == "user_assignment.expiration_update_decision"
    controls := [
        "AC-2 Account Management",
        "AC-3 Access Enforcement",
        "AC-6 Least Privilege",
        "AU-2 Audit Events",
        "AU-6 Audit Review, Analysis, and Reporting",
        "CM-3 Configuration Change Control",
        "ISO 27001 A.9.2.3 Management of privileged access rights",
        "COBIT DSS06.03 Manage roles, responsibilities, access privileges",
        "PCI DSS 7.1.2 Role-based access control"
    ]
} else = controls {
    decision_path == "user_assignment.role_check_decision"
    controls := [
        "AC-3 Access Enforcement",
        "AC-6 Least Privilege",
        "AU-2 Audit Events",
        "ISO 27001 A.9.4.1 Information access restriction",
        "COBIT DSS06.03 Manage roles, responsibilities, access privileges"
    ]
} else = [
    "AC-3 Access Enforcement",
    "AU-2 Audit Events",
    "ISO 27001 A.9.4 System and application access control",
    "COBIT DSS06 Manage business process controls"
]

# ---------------------------------------------------------
# Função para verificar se a auditoria está habilitada
# ---------------------------------------------------------
is_audit_enabled(context) {
    not context.skip_audit
}

# ---------------------------------------------------------
# Função para verificar nível de sensibilidade da operação
# ---------------------------------------------------------
operation_sensitivity_level(decision_path) = level {
    high_sensitivity_operations := {
        "crud.create_decision",
        "crud.delete_decision",
        "crud.permanent_delete_decision",
        "permissions.permission_assignment_decision",
        "permissions.permission_revocation_decision",
        "hierarchy.hierarchy_addition_decision",
        "hierarchy.hierarchy_removal_decision",
        "user_assignment.role_assignment_decision",
        "user_assignment.role_removal_decision"
    }
    
    decision_path == high_sensitivity_operations[_]
    level := "HIGH"
} else = level {
    medium_sensitivity_operations := {
        "crud.update_decision",
        "user_assignment.expiration_update_decision"
    }
    
    decision_path == medium_sensitivity_operations[_]
    level := "MEDIUM"
} else = "LOW"

# ---------------------------------------------------------
# Função para determinar se uma operação requer aprovação dupla
# ---------------------------------------------------------
requires_dual_approval(decision_path, resource) {
    # Operações de alta sensibilidade que afetam funções do sistema
    operation_sensitivity_level(decision_path) == "HIGH"
    
    # Verificar se a função é do sistema
    resource.type == constants.system_role_type
}

requires_dual_approval(decision_path, resource) {
    # Operações específicas que sempre requerem aprovação dupla
    critical_operations := {
        "crud.permanent_delete_decision",
        "permissions.permission_assignment_decision",
        "user_assignment.role_assignment_decision"
    }
    
    decision_path == critical_operations[_]
    
    # E a função ou permissão é crítica
    is_critical_resource(resource)
}

# Verifica se o recurso é crítico
is_critical_resource(resource) {
    # Se o recurso for uma função de sistema protegida
    resource.type == "ROLE"
    critical_roles := ["SUPER_ADMIN", "TENANT_ADMIN", "IAM_ADMIN", "SYSTEM_ADMIN"]
    name_upper := upper(resource.name)
    critical_roles[_] == name_upper
} else {
    # Se o recurso for uma permissão crítica
    resource.type == "PERMISSION"
    critical_permissions := ["role:create", "role:delete", "role:permanent_delete", "system:*", "tenant:*"]
    perm := critical_permissions[_]
    glob.match(perm, [], resource.name)
}