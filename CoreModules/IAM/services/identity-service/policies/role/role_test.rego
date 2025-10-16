# role_test.rego
# Testes unitários para as políticas de autorização do RoleService
# Conformidade: ISO/IEC 27001, TOGAF 10.0, COBIT 2019, NIST SP 800-53, PCI DSS v4.0
package innovabiz.iam.role_test

import data.innovabiz.iam.role.crud
import data.innovabiz.iam.role.permissions
import data.innovabiz.iam.role.hierarchy
import data.innovabiz.iam.role.user_assignment
import future.keywords

# === Dados de Teste ===

# Usuários
test_users := {
    "super_admin": {
        "id": "00000000-0000-0000-0000-000000000001",
        "tenant_id": "10000000-0000-0000-0000-000000000001",
        "username": "super.admin",
        "roles": ["SUPER_ADMIN"],
        "permissions": ["*:*"]
    },
    "tenant_admin": {
        "id": "00000000-0000-0000-0000-000000000002",
        "tenant_id": "20000000-0000-0000-0000-000000000001",
        "username": "tenant.admin",
        "roles": ["TENANT_ADMIN"],
        "permissions": ["tenant:manage", "role:*"]
    },
    "iam_admin": {
        "id": "00000000-0000-0000-0000-000000000003",
        "tenant_id": "20000000-0000-0000-0000-000000000001",
        "username": "iam.admin",
        "roles": ["IAM_ADMIN"],
        "permissions": ["role:create", "role:read", "role:update", "role:delete"]
    },
    "iam_operator": {
        "id": "00000000-0000-0000-0000-000000000004",
        "tenant_id": "20000000-0000-0000-0000-000000000001",
        "username": "iam.operator",
        "roles": ["IAM_OPERATOR"],
        "permissions": ["role:read", "role:assign_to_user", "role:check_permission"]
    },
    "regular_user": {
        "id": "00000000-0000-0000-0000-000000000005",
        "tenant_id": "20000000-0000-0000-0000-000000000001",
        "username": "regular.user",
        "roles": ["USER"],
        "permissions": ["role:read"]
    },
    "different_tenant_admin": {
        "id": "00000000-0000-0000-0000-000000000006",
        "tenant_id": "30000000-0000-0000-0000-000000000001",
        "username": "other.admin",
        "roles": ["TENANT_ADMIN"],
        "permissions": ["tenant:manage", "role:*"]
    }
}

# Funções
test_roles := {
    "super_admin_role": {
        "id": "10000000-0000-0000-0000-000000000001",
        "tenant_id": "10000000-0000-0000-0000-000000000001",
        "name": "SUPER_ADMIN",
        "description": "Super Administrator",
        "type": "SYSTEM",
        "status": "ACTIVE"
    },
    "tenant_admin_role": {
        "id": "20000000-0000-0000-0000-000000000001",
        "tenant_id": "20000000-0000-0000-0000-000000000001",
        "name": "TENANT_ADMIN",
        "description": "Tenant Administrator",
        "type": "SYSTEM",
        "status": "ACTIVE"
    },
    "custom_role": {
        "id": "20000000-0000-0000-0000-000000000002",
        "tenant_id": "20000000-0000-0000-0000-000000000001",
        "name": "CUSTOM_ROLE",
        "description": "Custom Role",
        "type": "CUSTOM",
        "status": "ACTIVE"
    },
    "inactive_role": {
        "id": "20000000-0000-0000-0000-000000000003",
        "tenant_id": "20000000-0000-0000-0000-000000000001",
        "name": "INACTIVE_ROLE",
        "description": "Inactive Role",
        "type": "CUSTOM",
        "status": "INACTIVE"
    },
    "other_tenant_role": {
        "id": "30000000-0000-0000-0000-000000000001",
        "tenant_id": "30000000-0000-0000-0000-000000000001",
        "name": "OTHER_TENANT_ROLE",
        "description": "Role in another tenant",
        "type": "CUSTOM",
        "status": "ACTIVE"
    }
}

# Permissões
test_permissions := {
    "role_create": {
        "id": "00000000-0000-0000-0000-000000000001",
        "name": "role:create",
        "description": "Create roles"
    },
    "role_read": {
        "id": "00000000-0000-0000-0000-000000000002",
        "name": "role:read",
        "description": "Read roles"
    },
    "critical_permission": {
        "id": "00000000-0000-0000-0000-000000000003",
        "name": "iam:super_admin",
        "description": "Super admin permission"
    }
}

# Hierarquia de funções
test_role_hierarchies := [
    {
        "id": "00000000-0000-0000-0000-000000000001",
        "tenant_id": "20000000-0000-0000-0000-000000000001",
        "parent_id": "20000000-0000-0000-0000-000000000001", # tenant_admin_role
        "child_id": "20000000-0000-0000-0000-000000000002"  # custom_role
    }
]

# Atribuições de função a usuários
test_user_roles := [
    {
        "id": "00000000-0000-0000-0000-000000000001",
        "tenant_id": "10000000-0000-0000-0000-000000000001",
        "user_id": "00000000-0000-0000-0000-000000000001", # super_admin
        "role_id": "10000000-0000-0000-0000-000000000001", # super_admin_role
        "status": "ACTIVE"
    },
    {
        "id": "00000000-0000-0000-0000-000000000002",
        "tenant_id": "20000000-0000-0000-0000-000000000001",
        "user_id": "00000000-0000-0000-0000-000000000002", # tenant_admin
        "role_id": "20000000-0000-0000-0000-000000000001", # tenant_admin_role
        "status": "ACTIVE"
    }
]

# === Funções auxiliares para teste ===

# Mock de dados para testes
mock_data = {
    "roles": {
        "10000000-0000-0000-0000-000000000001": test_roles.super_admin_role,
        "20000000-0000-0000-0000-000000000001": test_roles.tenant_admin_role,
        "20000000-0000-0000-0000-000000000002": test_roles.custom_role,
        "20000000-0000-0000-0000-000000000003": test_roles.inactive_role,
        "30000000-0000-0000-0000-000000000001": test_roles.other_tenant_role
    },
    "permissions": {
        "00000000-0000-0000-0000-000000000001": test_permissions.role_create,
        "00000000-0000-0000-0000-000000000002": test_permissions.role_read,
        "00000000-0000-0000-0000-000000000003": test_permissions.critical_permission
    },
    "role_permissions": [],
    "role_hierarchies": test_role_hierarchies,
    "user_roles": test_user_roles,
    "users": {
        "00000000-0000-0000-0000-000000000001": test_users.super_admin,
        "00000000-0000-0000-0000-000000000002": test_users.tenant_admin,
        "00000000-0000-0000-0000-000000000003": test_users.iam_admin,
        "00000000-0000-0000-0000-000000000004": test_users.iam_operator,
        "00000000-0000-0000-0000-000000000005": test_users.regular_user,
        "00000000-0000-0000-0000-000000000006": test_users.different_tenant_admin
    },
    "tenant_settings": {
        "10000000-0000-0000-0000-000000000001": {
            "max_hierarchy_depth": 10
        },
        "20000000-0000-0000-0000-000000000001": {
            "max_hierarchy_depth": 5
        },
        "30000000-0000-0000-0000-000000000001": {
            "max_hierarchy_depth": 3
        }
    }
}

# === TESTES PARA CRIAR FUNÇÃO ===

test_allow_super_admin_create_role if {
    # Super admin pode criar qualquer tipo de função
    mock_data.roles
    
    decision := crud.create_decision with input as {
        "user": test_users.super_admin,
        "resource": {
            "tenant_id": "10000000-0000-0000-0000-000000000001",
            "data": {
                "name": "NEW_SYSTEM_ROLE",
                "description": "New System Role",
                "type": "SYSTEM"
            }
        }
    } with data.roles as mock_data.roles
    
    decision.allow == true
}

test_allow_tenant_admin_create_custom_role if {
    # Tenant admin pode criar função customizada em seu tenant
    decision := crud.create_decision with input as {
        "user": test_users.tenant_admin,
        "resource": {
            "tenant_id": "20000000-0000-0000-0000-000000000001",
            "data": {
                "name": "NEW_CUSTOM_ROLE",
                "description": "New Custom Role",
                "type": "CUSTOM"
            }
        }
    } with data.roles as mock_data.roles
    
    decision.allow == true
}

test_deny_tenant_admin_create_system_role if {
    # Tenant admin não pode criar função de sistema
    decision := crud.create_decision with input as {
        "user": test_users.tenant_admin,
        "resource": {
            "tenant_id": "20000000-0000-0000-0000-000000000001",
            "data": {
                "name": "NEW_SYSTEM_ROLE",
                "description": "New System Role",
                "type": "SYSTEM"
            }
        }
    } with data.roles as mock_data.roles
    
    decision.allow == false
}

test_deny_tenant_admin_create_role_other_tenant if {
    # Tenant admin não pode criar função em outro tenant
    decision := crud.create_decision with input as {
        "user": test_users.tenant_admin,
        "resource": {
            "tenant_id": "30000000-0000-0000-0000-000000000001",
            "data": {
                "name": "NEW_CUSTOM_ROLE",
                "description": "New Custom Role",
                "type": "CUSTOM"
            }
        }
    } with data.roles as mock_data.roles
    
    decision.allow == false
}

test_deny_invalid_role_name if {
    # Nome de função inválido (vazio)
    decision := crud.create_decision with input as {
        "user": test_users.super_admin,
        "resource": {
            "tenant_id": "10000000-0000-0000-0000-000000000001",
            "data": {
                "name": "",
                "description": "Invalid Role",
                "type": "CUSTOM"
            }
        }
    } with data.roles as mock_data.roles
    
    decision.allow == false
}

test_deny_reserved_role_name if {
    # Nome de função reservado
    decision := crud.create_decision with input as {
        "user": test_users.tenant_admin,
        "resource": {
            "tenant_id": "20000000-0000-0000-0000-000000000001",
            "data": {
                "name": "SUPER_ADMIN",
                "description": "Attempted Reserved Name",
                "type": "CUSTOM"
            }
        }
    } with data.roles as mock_data.roles
    
    decision.allow == false
}

# === TESTES PARA ATRIBUIR PERMISSÃO ===

test_allow_super_admin_assign_permission if {
    # Super admin pode atribuir qualquer permissão
    decision := permissions.permission_assignment_decision with input as {
        "user": test_users.super_admin,
        "resource": {
            "tenant_id": "10000000-0000-0000-0000-000000000001",
            "role_id": "10000000-0000-0000-0000-000000000001",
            "permission_id": "00000000-0000-0000-0000-000000000001"
        }
    } with data as mock_data
    
    decision.allow == true
}

test_allow_tenant_admin_assign_permission if {
    # Tenant admin pode atribuir permissão em seu tenant
    decision := permissions.permission_assignment_decision with input as {
        "user": test_users.tenant_admin,
        "resource": {
            "tenant_id": "20000000-0000-0000-0000-000000000001",
            "role_id": "20000000-0000-0000-0000-000000000002",
            "permission_id": "00000000-0000-0000-0000-000000000001"
        }
    } with data as mock_data
    
    decision.allow == true
}

test_deny_tenant_admin_assign_critical_permission if {
    # Tenant admin não pode atribuir permissão crítica
    decision := permissions.permission_assignment_decision with input as {
        "user": test_users.tenant_admin,
        "resource": {
            "tenant_id": "20000000-0000-0000-0000-000000000001",
            "role_id": "20000000-0000-0000-0000-000000000002",
            "permission_id": "00000000-0000-0000-0000-000000000003"
        }
    } with data as mock_data
    
    decision.allow == false
}

test_deny_tenant_admin_assign_permission_other_tenant if {
    # Tenant admin não pode atribuir permissão em outro tenant
    decision := permissions.permission_assignment_decision with input as {
        "user": test_users.tenant_admin,
        "resource": {
            "tenant_id": "30000000-0000-0000-0000-000000000001",
            "role_id": "30000000-0000-0000-0000-000000000001",
            "permission_id": "00000000-0000-0000-0000-000000000001"
        }
    } with data as mock_data
    
    decision.allow == false
}

# === TESTES PARA HIERARQUIA DE FUNÇÕES ===

test_allow_super_admin_add_hierarchy if {
    # Super admin pode adicionar hierarquia
    decision := hierarchy.hierarchy_addition_decision with input as {
        "user": test_users.super_admin,
        "resource": {
            "tenant_id": "10000000-0000-0000-0000-000000000001",
            "parent_id": "10000000-0000-0000-0000-000000000001",
            "child_id": "20000000-0000-0000-0000-000000000001"
        }
    } with data as mock_data
    
    decision.allow == true
}

test_allow_tenant_admin_add_hierarchy if {
    # Tenant admin pode adicionar hierarquia em seu tenant
    decision := hierarchy.hierarchy_addition_decision with input as {
        "user": test_users.tenant_admin,
        "resource": {
            "tenant_id": "20000000-0000-0000-0000-000000000001",
            "parent_id": "20000000-0000-0000-0000-000000000001",
            "child_id": "20000000-0000-0000-0000-000000000002"
        }
    } with data as mock_data
    
    # Esta hierarquia já existe no mock, então deve falhar
    decision.allow == false
    decision.reason == "a relação hierárquica já existe"
}

test_deny_hierarchy_cycle if {
    # Detecta ciclo na hierarquia
    modified_data := mock_data
    modified_data.role_hierarchies = [
        {
            "id": "00000000-0000-0000-0000-000000000001",
            "tenant_id": "20000000-0000-0000-0000-000000000001",
            "parent_id": "20000000-0000-0000-0000-000000000002", # custom_role
            "child_id": "20000000-0000-0000-0000-000000000003"  # inactive_role
        }
    ]
    
    decision := hierarchy.hierarchy_addition_decision with input as {
        "user": test_users.tenant_admin,
        "resource": {
            "tenant_id": "20000000-0000-0000-0000-000000000001",
            "parent_id": "20000000-0000-0000-0000-000000000003", # inactive_role
            "child_id": "20000000-0000-0000-0000-000000000002"  # custom_role, criaria ciclo
        }
    } with data as modified_data
    
    decision.allow == false
    decision.reason == "a adição desta relação hierárquica criaria um ciclo"
}

test_deny_hierarchy_cross_tenant if {
    # Não permite hierarquia entre funções de tenants diferentes
    decision := hierarchy.hierarchy_addition_decision with input as {
        "user": test_users.super_admin,
        "resource": {
            "tenant_id": "10000000-0000-0000-0000-000000000001",
            "parent_id": "20000000-0000-0000-0000-000000000001", # tenant_admin_role (tenant 2)
            "child_id": "30000000-0000-0000-0000-000000000001"  # other_tenant_role (tenant 3)
        }
    } with data as mock_data
    
    decision.allow == false
    decision.reason == "as funções pai e filho devem pertencer ao mesmo tenant"
}

# === TESTES PARA ATRIBUIÇÃO DE FUNÇÃO A UTILIZADOR ===

test_allow_super_admin_assign_role if {
    # Super admin pode atribuir qualquer função
    decision := user_assignment.role_assignment_decision with input as {
        "user": test_users.super_admin,
        "resource": {
            "tenant_id": "10000000-0000-0000-0000-000000000001",
            "role_id": "10000000-0000-0000-0000-000000000001",
            "user_id": "00000000-0000-0000-0000-000000000005",
            "expires_at": null
        }
    } with data as mock_data
    
    decision.allow == true
}

test_allow_tenant_admin_assign_role if {
    # Tenant admin pode atribuir função em seu tenant
    decision := user_assignment.role_assignment_decision with input as {
        "user": test_users.tenant_admin,
        "resource": {
            "tenant_id": "20000000-0000-0000-0000-000000000001",
            "role_id": "20000000-0000-0000-0000-000000000002",
            "user_id": "00000000-0000-0000-0000-000000000005",
            "expires_at": null
        }
    } with data as mock_data
    
    decision.allow == true
}

test_deny_tenant_admin_assign_role_other_tenant if {
    # Tenant admin não pode atribuir função em outro tenant
    decision := user_assignment.role_assignment_decision with input as {
        "user": test_users.tenant_admin,
        "resource": {
            "tenant_id": "30000000-0000-0000-0000-000000000001",
            "role_id": "30000000-0000-0000-0000-000000000001",
            "user_id": "00000000-0000-0000-0000-000000000006",
            "expires_at": null
        }
    } with data as mock_data
    
    decision.allow == false
}

test_allow_iam_operator_assign_custom_role if {
    # IAM Operator pode atribuir função customizada
    decision := user_assignment.role_assignment_decision with input as {
        "user": test_users.iam_operator,
        "resource": {
            "tenant_id": "20000000-0000-0000-0000-000000000001",
            "role_id": "20000000-0000-0000-0000-000000000002", # custom_role
            "user_id": "00000000-0000-0000-0000-000000000005",
            "expires_at": null
        }
    } with data as mock_data
    
    decision.allow == true
}

test_deny_iam_operator_assign_system_role if {
    # IAM Operator não pode atribuir função de sistema sensível
    decision := user_assignment.role_assignment_decision with input as {
        "user": test_users.iam_operator,
        "resource": {
            "tenant_id": "20000000-0000-0000-0000-000000000001",
            "role_id": "20000000-0000-0000-0000-000000000001", # tenant_admin_role
            "user_id": "00000000-0000-0000-0000-000000000005",
            "expires_at": null
        }
    } with data as mock_data
    
    decision.allow == false
}

test_deny_invalid_expiration if {
    # Rejeita data de expiração no passado
    past_date := "2020-01-01T00:00:00Z"
    
    decision := user_assignment.role_assignment_decision with input as {
        "user": test_users.tenant_admin,
        "resource": {
            "tenant_id": "20000000-0000-0000-0000-000000000001",
            "role_id": "20000000-0000-0000-0000-000000000002",
            "user_id": "00000000-0000-0000-0000-000000000005",
            "expires_at": past_date
        }
    } with data as mock_data
    
    decision.allow == false
    decision.reason == "a data de expiração deve estar no futuro"
}

test_deny_assign_existing_role if {
    # Não permite atribuir função que já está atribuída
    decision := user_assignment.role_assignment_decision with input as {
        "user": test_users.super_admin,
        "resource": {
            "tenant_id": "10000000-0000-0000-0000-000000000001",
            "role_id": "10000000-0000-0000-0000-000000000001", # super_admin_role
            "user_id": "00000000-0000-0000-0000-000000000001", # super_admin (já tem essa função)
            "expires_at": null
        }
    } with data as mock_data
    
    decision.allow == false
    decision.reason == "esta função já está atribuída a este utilizador"
}