# Testes unitários para políticas CRUD de RoleService - INNOVABIZ Platform
#
# Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
# PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III
package innovabiz.iam.role.crud_test

import data.innovabiz.iam.role.crud
import data.innovabiz.iam.role.constants

# ---------------------------------------------------------
# Testes para decisão de criação de função
# ---------------------------------------------------------
test_super_admin_can_create_any_role {
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
            "path": "/roles",
            "data": {
                "name": "NEW_ROLE",
                "type": "SYSTEM",
                "description": "Nova função de sistema"
            }
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "crud.create_decision"
    }
    
    # Verificar se a decisão é permitida
    crud.create_decision == true with input as input
}

test_tenant_admin_can_create_custom_role {
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
            "path": "/roles",
            "data": {
                "name": "CUSTOM_ROLE",
                "type": "CUSTOM",
                "description": "Nova função customizada"
            }
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "crud.create_decision"
    }
    
    # Verificar se a decisão é permitida
    crud.create_decision == true with input as input
}

test_tenant_admin_cannot_create_system_role {
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
            "path": "/roles",
            "data": {
                "name": "NEW_ROLE",
                "type": "SYSTEM",  # Tipo de função de sistema
                "description": "Nova função de sistema"
            }
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "crud.create_decision"
    }
    
    # Verificar se a decisão é negada
    not crud.create_decision with input as input
}

test_tenant_admin_cannot_create_protected_role_name {
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
            "path": "/roles",
            "data": {
                "name": "SUPER_ADMIN",  # Nome protegido
                "type": "CUSTOM",
                "description": "Tentativa de criar função com nome protegido"
            }
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "crud.create_decision"
    }
    
    # Verificar se a decisão é negada
    not crud.create_decision with input as input
}

test_iam_admin_can_create_custom_role_with_permission {
    # Configurar entrada para teste
    input := {
        "http_method": "POST",
        "tenant_id": "tenant-1",
        "user": {
            "id": "iam-admin-user",
            "roles": ["IAM_ADMIN"],
            "permissions": ["role:create"],
            "tenant_id": "tenant-1"
        },
        "resource": {
            "path": "/roles",
            "data": {
                "name": "CUSTOM_ROLE",
                "type": "CUSTOM",
                "description": "Nova função customizada"
            }
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "crud.create_decision"
    }
    
    # Verificar se a decisão é permitida
    crud.create_decision == true with input as input
}

test_iam_admin_cannot_create_custom_role_without_permission {
    # Configurar entrada para teste
    input := {
        "http_method": "POST",
        "tenant_id": "tenant-1",
        "user": {
            "id": "iam-admin-user",
            "roles": ["IAM_ADMIN"],
            "permissions": [],  # Sem a permissão necessária
            "tenant_id": "tenant-1"
        },
        "resource": {
            "path": "/roles",
            "data": {
                "name": "CUSTOM_ROLE",
                "type": "CUSTOM",
                "description": "Nova função customizada"
            }
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "crud.create_decision"
    }
    
    # Verificar se a decisão é negada
    not crud.create_decision with input as input
}

# ---------------------------------------------------------
# Testes para decisão de leitura de função
# ---------------------------------------------------------
test_super_admin_can_read_any_role {
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
            "path": "/roles/1",
            "id": "1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "crud.read_decision"
    }
    
    # Verificar se a decisão é permitida
    crud.read_decision == true with input as input
}

test_tenant_admin_can_read_role_in_own_tenant {
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
            "path": "/roles/1",
            "id": "1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "crud.read_decision"
    }
    
    # Verificar se a decisão é permitida
    crud.read_decision == true with input as input
}

test_tenant_admin_cannot_read_role_in_other_tenant {
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
            "path": "/roles/1",
            "id": "1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "crud.read_decision"
    }
    
    # Verificar se a decisão é negada
    not crud.read_decision with input as input
}

# ---------------------------------------------------------
# Testes para decisão de exclusão de função
# ---------------------------------------------------------
test_super_admin_can_delete_any_role {
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
            "path": "/roles/2",
            "id": "2"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "crud.delete_decision"
    }
    
    # Verificar se a decisão é permitida
    crud.delete_decision == true with input as input
}

test_tenant_admin_can_delete_custom_role {
    # Função auxiliar para simular get_role_by_id
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
            "path": "/roles/2",
            "id": "2"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "crud.delete_decision"
    }
    
    # Verificar se a decisão é permitida
    crud.delete_decision == true with input as input with data.innovabiz.iam.role.crud.get_role_by_id as mock_get_role_by_id
}

test_tenant_admin_cannot_delete_system_role {
    # Função auxiliar para simular get_role_by_id
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
            "path": "/roles/1",
            "id": "1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "crud.delete_decision"
    }
    
    # Verificar se a decisão é negada
    not crud.delete_decision with input as input with data.innovabiz.iam.role.crud.get_role_by_id as mock_get_role_by_id
}

# ---------------------------------------------------------
# Testes para decisão de exclusão permanente de função
# ---------------------------------------------------------
test_super_admin_can_permanent_delete_any_role {
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
            "path": "/roles/2/permanent",
            "id": "2"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "crud.permanent_delete_decision"
    }
    
    # Verificar se a decisão é permitida
    crud.permanent_delete_decision == true with input as input
}