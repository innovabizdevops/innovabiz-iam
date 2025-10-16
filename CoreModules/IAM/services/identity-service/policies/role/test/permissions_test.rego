# Testes unitários para políticas de permissões do RoleService - INNOVABIZ Platform
#
# Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
# PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III
package innovabiz.iam.role.permissions_test

import data.innovabiz.iam.role.permissions
import data.innovabiz.iam.role.constants

# ---------------------------------------------------------
# Testes para decisão de atribuição de permissões
# ---------------------------------------------------------
test_super_admin_can_assign_any_permission {
    # Configurar entrada para teste
    input := {
        "http_method": "POST",
        "tenant_id": "tenant-1",
        "user": {
            "id": "super-admin-user",
            "roles": ["SUPER_ADMIN"],
            "tenant_id": constants.default_tenant
        },
        "resource": {
            "path": "/roles/1/permissions",
            "role_id": "1",
            "permission_id": "1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "permissions.permission_assignment_decision"
    }
    
    # Verificar se a decisão é permitida
    permissions.permission_assignment_decision == true with input as input
}

test_tenant_admin_can_assign_normal_permission_to_custom_role {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "CUSTOM_ROLE", "type": "CUSTOM"} {
        id == "2"
    }
    
    mock_get_permission_by_id(id) = {"id": id, "name": "role:read", "description": "Permite ler funções"} {
        id == "2"
    }
    
    # Configurar entrada para teste
    input := {
        "http_method": "POST",
        "tenant_id": "tenant-1",
        "user": {
            "id": "tenant-admin-user",
            "roles": ["TENANT_ADMIN"],
            "tenant_id": "tenant-1"
        },
        "resource": {
            "path": "/roles/2/permissions",
            "role_id": "2",
            "permission_id": "2"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "permissions.permission_assignment_decision"
    }
    
    # Verificar se a decisão é permitida
    permissions.permission_assignment_decision == true with input as input with data.innovabiz.iam.role.permissions.get_role_by_id as mock_get_role_by_id with data.innovabiz.iam.role.permissions.get_permission_by_id as mock_get_permission_by_id
}

test_tenant_admin_cannot_assign_restricted_permission {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "CUSTOM_ROLE", "type": "CUSTOM"} {
        id == "2"
    }
    
    mock_get_permission_by_id(id) = {"id": id, "name": "super_admin:access", "description": "Acesso de super admin"} {
        id == "5"
    }
    
    # Configurar entrada para teste
    input := {
        "http_method": "POST",
        "tenant_id": "tenant-1",
        "user": {
            "id": "tenant-admin-user",
            "roles": ["TENANT_ADMIN"],
            "tenant_id": "tenant-1"
        },
        "resource": {
            "path": "/roles/2/permissions",
            "role_id": "2",
            "permission_id": "5"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "permissions.permission_assignment_decision"
    }
    
    # Verificar se a decisão é negada
    not permissions.permission_assignment_decision with input as input with data.innovabiz.iam.role.permissions.get_role_by_id as mock_get_role_by_id with data.innovabiz.iam.role.permissions.get_permission_by_id as mock_get_permission_by_id
}

test_tenant_admin_cannot_assign_permission_to_system_role {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "ADMIN", "type": "SYSTEM"} {
        id == "1"
    }
    
    mock_get_permission_by_id(id) = {"id": id, "name": "role:read", "description": "Permite ler funções"} {
        id == "2"
    }
    
    # Configurar entrada para teste
    input := {
        "http_method": "POST",
        "tenant_id": "tenant-1",
        "user": {
            "id": "tenant-admin-user",
            "roles": ["TENANT_ADMIN"],
            "tenant_id": "tenant-1"
        },
        "resource": {
            "path": "/roles/1/permissions",
            "role_id": "1",
            "permission_id": "2"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "permissions.permission_assignment_decision"
    }
    
    # Verificar se a decisão é negada
    not permissions.permission_assignment_decision with input as input with data.innovabiz.iam.role.permissions.get_role_by_id as mock_get_role_by_id with data.innovabiz.iam.role.permissions.get_permission_by_id as mock_get_permission_by_id
}

test_iam_admin_can_assign_permission_with_specific_permission {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "CUSTOM_ROLE", "type": "CUSTOM"} {
        id == "2"
    }
    
    mock_get_permission_by_id(id) = {"id": id, "name": "role:read", "description": "Permite ler funções"} {
        id == "2"
    }
    
    # Configurar entrada para teste
    input := {
        "http_method": "POST",
        "tenant_id": "tenant-1",
        "user": {
            "id": "iam-admin-user",
            "roles": ["IAM_ADMIN"],
            "permissions": ["role:assign_permission"],
            "tenant_id": "tenant-1"
        },
        "resource": {
            "path": "/roles/2/permissions",
            "role_id": "2",
            "permission_id": "2"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "permissions.permission_assignment_decision"
    }
    
    # Verificar se a decisão é permitida
    permissions.permission_assignment_decision == true with input as input with data.innovabiz.iam.role.permissions.get_role_by_id as mock_get_role_by_id with data.innovabiz.iam.role.permissions.get_permission_by_id as mock_get_permission_by_id
}

test_iam_admin_cannot_assign_permission_without_specific_permission {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "CUSTOM_ROLE", "type": "CUSTOM"} {
        id == "2"
    }
    
    mock_get_permission_by_id(id) = {"id": id, "name": "role:read", "description": "Permite ler funções"} {
        id == "2"
    }
    
    # Configurar entrada para teste
    input := {
        "http_method": "POST",
        "tenant_id": "tenant-1",
        "user": {
            "id": "iam-admin-user",
            "roles": ["IAM_ADMIN"],
            "permissions": [],  # Sem a permissão específica
            "tenant_id": "tenant-1"
        },
        "resource": {
            "path": "/roles/2/permissions",
            "role_id": "2",
            "permission_id": "2"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "permissions.permission_assignment_decision"
    }
    
    # Verificar se a decisão é negada
    not permissions.permission_assignment_decision with input as input with data.innovabiz.iam.role.permissions.get_role_by_id as mock_get_role_by_id with data.innovabiz.iam.role.permissions.get_permission_by_id as mock_get_permission_by_id
}

# ---------------------------------------------------------
# Testes para decisão de revogação de permissões
# ---------------------------------------------------------
test_super_admin_can_revoke_any_permission {
    # Configurar entrada para teste
    input := {
        "http_method": "DELETE",
        "tenant_id": "tenant-1",
        "user": {
            "id": "super-admin-user",
            "roles": ["SUPER_ADMIN"],
            "tenant_id": constants.default_tenant
        },
        "resource": {
            "path": "/roles/1/permissions/1",
            "role_id": "1",
            "permission_id": "1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "permissions.permission_revocation_decision"
    }
    
    # Verificar se a decisão é permitida
    permissions.permission_revocation_decision == true with input as input
}

test_tenant_admin_can_revoke_permission_from_custom_role {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "CUSTOM_ROLE", "type": "CUSTOM"} {
        id == "2"
    }
    
    # Configurar entrada para teste
    input := {
        "http_method": "DELETE",
        "tenant_id": "tenant-1",
        "user": {
            "id": "tenant-admin-user",
            "roles": ["TENANT_ADMIN"],
            "tenant_id": "tenant-1"
        },
        "resource": {
            "path": "/roles/2/permissions/2",
            "role_id": "2",
            "permission_id": "2"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "permissions.permission_revocation_decision"
    }
    
    # Verificar se a decisão é permitida
    permissions.permission_revocation_decision == true with input as input with data.innovabiz.iam.role.permissions.get_role_by_id as mock_get_role_by_id
}

test_tenant_admin_cannot_revoke_permission_from_system_role {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "ADMIN", "type": "SYSTEM"} {
        id == "1"
    }
    
    # Configurar entrada para teste
    input := {
        "http_method": "DELETE",
        "tenant_id": "tenant-1",
        "user": {
            "id": "tenant-admin-user",
            "roles": ["TENANT_ADMIN"],
            "tenant_id": "tenant-1"
        },
        "resource": {
            "path": "/roles/1/permissions/1",
            "role_id": "1",
            "permission_id": "1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "permissions.permission_revocation_decision"
    }
    
    # Verificar se a decisão é negada
    not permissions.permission_revocation_decision with input as input with data.innovabiz.iam.role.permissions.get_role_by_id as mock_get_role_by_id
}

# ---------------------------------------------------------
# Testes para decisão de verificação de permissões
# ---------------------------------------------------------
test_super_admin_can_check_any_permission {
    # Configurar entrada para teste
    input := {
        "http_method": "GET",
        "tenant_id": "tenant-1",
        "user": {
            "id": "super-admin-user",
            "roles": ["SUPER_ADMIN"],
            "tenant_id": constants.default_tenant
        },
        "resource": {
            "path": "/roles/1/permissions",
            "role_id": "1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "permissions.permission_check_decision"
    }
    
    # Verificar se a decisão é permitida
    permissions.permission_check_decision == true with input as input
}

test_tenant_admin_can_check_permissions_in_own_tenant {
    # Configurar entrada para teste
    input := {
        "http_method": "GET",
        "tenant_id": "tenant-1",
        "user": {
            "id": "tenant-admin-user",
            "roles": ["TENANT_ADMIN"],
            "tenant_id": "tenant-1"
        },
        "resource": {
            "path": "/roles/1/permissions",
            "role_id": "1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "permissions.permission_check_decision"
    }
    
    # Verificar se a decisão é permitida
    permissions.permission_check_decision == true with input as input
}

test_tenant_admin_cannot_check_permissions_in_other_tenant {
    # Configurar entrada para teste
    input := {
        "http_method": "GET",
        "tenant_id": "tenant-2",  # Tenant diferente
        "user": {
            "id": "tenant-admin-user",
            "roles": ["TENANT_ADMIN"],
            "tenant_id": "tenant-1"
        },
        "resource": {
            "path": "/roles/1/permissions",
            "role_id": "1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "permissions.permission_check_decision"
    }
    
    # Verificar se a decisão é negada
    not permissions.permission_check_decision with input as input
}

test_iam_operator_can_check_permissions {
    # Configurar entrada para teste
    input := {
        "http_method": "GET",
        "tenant_id": "tenant-1",
        "user": {
            "id": "iam-operator-user",
            "roles": ["IAM_OPERATOR"],
            "tenant_id": "tenant-1"
        },
        "resource": {
            "path": "/roles/1/permissions",
            "role_id": "1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "permissions.permission_check_decision"
    }
    
    # Verificar se a decisão é permitida
    permissions.permission_check_decision == true with input as input
}