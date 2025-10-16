# Constantes para as políticas do RoleService - INNOVABIZ Platform
#
# Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
# PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III
package innovabiz.iam.role.constants

# Constantes para funções do sistema
super_admin_role := "SUPER_ADMIN"
tenant_admin_role := "TENANT_ADMIN"
iam_admin_role := "IAM_ADMIN"
iam_operator_role := "IAM_OPERATOR"

# Tipos de função
system_role_type := "SYSTEM"
custom_role_type := "CUSTOM"

# Constantes para configuração de hierarquia
max_hierarchy_depth := 5

# Tenant default
default_tenant := "00000000-0000-0000-0000-000000000000"

# Tempo máximo de cache para decisões (em segundos)
decision_cache_ttl := 60

# Número máximo de tentativas de acesso antes de bloqueio
max_access_attempts := 5

# Tempo de bloqueio após exceder tentativas máximas (em segundos)
lockout_duration := 1800  # 30 minutos

# Estados de função
role_status := {
    "active": "ACTIVE",
    "inactive": "INACTIVE",
    "deleted": "DELETED"
}

# Permissões críticas que requerem aprovação especial
critical_permissions := [
    "role:create",
    "role:delete",
    "role:permanent_delete",
    "role:assign_permission",
    "role:revoke_permission",
    "role:manage_hierarchy",
    "role:assign_to_user",
    "system:config",
    "system:logs",
    "tenant:create",
    "tenant:delete",
    "audit:delete"
]

# Funções protegidas que não podem ser modificadas exceto por super admin
protected_roles := [
    "SUPER_ADMIN",
    "TENANT_ADMIN",
    "IAM_ADMIN",
    "SYSTEM_ADMIN",
    "SECURITY_ADMIN",
    "AUDIT_ADMIN"
]

# Constantes para operações
operations := {
    "create": "CREATE",
    "read": "READ",
    "update": "UPDATE",
    "delete": "DELETE",
    "list": "LIST",
    "assign": "ASSIGN",
    "revoke": "REVOKE",
    "check": "CHECK"
}

# Constantes para normas de conformidade
compliance_standards := [
    "ISO/IEC 27001:2022",
    "TOGAF 10.0",
    "COBIT 2019",
    "NIST SP 800-53",
    "PCI DSS v4.0",
    "GDPR",
    "APD Angola",
    "BNA",
    "Basel III"
]

# Duração padrão de uma atribuição de função (em segundos)
# 90 dias = 7776000 segundos
default_role_assignment_duration := 7776000

# Duração máxima de uma atribuição de função (em segundos)
# 1 ano = 31536000 segundos
max_role_assignment_duration := 31536000