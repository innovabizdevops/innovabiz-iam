# Testes unitários para políticas de atribuição de função a usuários (Parte 2) - INNOVABIZ Platform
#
# Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
# PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III
package innovabiz.iam.role.user_assignment_test

import data.innovabiz.iam.role.user_assignment
import data.innovabiz.iam.role.constants

# ---------------------------------------------------------
# Testes para decisão de remoção de função de usuário
# ---------------------------------------------------------
test_super_admin_can_remove_any_role_assignment {
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
            "path": "/users/user-1/roles/1",
            "user_id": "user-1",
            "role_id": "1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "user_assignment.role_removal_decision"
    }
    
    # Verificar se a decisão é permitida
    user_assignment.role_removal_decision == true with input as input
}

test_tenant_admin_can_remove_custom_role_from_user_in_own_tenant {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "CUSTOM_ROLE", "type": "CUSTOM"} {
        id == "2"
    }
    
    mock_get_user_by_id(id) = {"id": id, "email": "user@example.com", "tenant_id": "tenant-1"} {
        id == "user-1"
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
            "path": "/users/user-1/roles/2",
            "user_id": "user-1",
            "role_id": "2"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "user_assignment.role_removal_decision"
    }
    
    # Verificar se a decisão é permitida
    user_assignment.role_removal_decision == true with input as input with data.innovabiz.iam.role.user_assignment.get_role_by_id as mock_get_role_by_id with data.innovabiz.iam.role.user_assignment.get_user_by_id as mock_get_user_by_id
}

test_tenant_admin_cannot_remove_system_role {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "SYSTEM_ADMIN", "type": "SYSTEM"} {
        id == "1"
    }
    
    mock_get_user_by_id(id) = {"id": id, "email": "user@example.com", "tenant_id": "tenant-1"} {
        id == "user-1"
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
            "path": "/users/user-1/roles/1",
            "user_id": "user-1",
            "role_id": "1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "user_assignment.role_removal_decision"
    }
    
    # Verificar se a decisão é negada
    not user_assignment.role_removal_decision with input as input with data.innovabiz.iam.role.user_assignment.get_role_by_id as mock_get_role_by_id with data.innovabiz.iam.role.user_assignment.get_user_by_id as mock_get_user_by_id
}

test_tenant_admin_cannot_remove_role_from_user_in_other_tenant {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "CUSTOM_ROLE", "type": "CUSTOM"} {
        id == "2"
    }
    
    mock_get_user_by_id(id) = {"id": id, "email": "user@example.com", "tenant_id": "tenant-2"} {  # Tenant diferente
        id == "user-1"
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
            "path": "/users/user-1/roles/2",
            "user_id": "user-1",
            "role_id": "2"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "user_assignment.role_removal_decision"
    }
    
    # Verificar se a decisão é negada
    not user_assignment.role_removal_decision with input as input with data.innovabiz.iam.role.user_assignment.get_role_by_id as mock_get_role_by_id with data.innovabiz.iam.role.user_assignment.get_user_by_id as mock_get_user_by_id
}

# ---------------------------------------------------------
# Testes para decisão de atualização de expiração de função
# ---------------------------------------------------------
test_super_admin_can_update_any_role_expiration {
    # Configurar entrada para teste
    input := {
        "http_method": "PATCH",
        "tenant_id": "tenant-1",
        "user": {
            "id": "super-admin-user",
            "roles": ["SUPER_ADMIN"],
            "tenant_id": constants.default_tenant
        },
        "resource": {
            "path": "/users/user-1/roles/1/expiration",
            "user_id": "user-1",
            "role_id": "1",
            "expiration": "2024-06-30T23:59:59Z"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "user_assignment.expiration_update_decision"
    }
    
    # Verificar se a decisão é permitida
    user_assignment.expiration_update_decision == true with input as input
}

test_tenant_admin_can_update_custom_role_expiration {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "CUSTOM_ROLE", "type": "CUSTOM"} {
        id == "2"
    }
    
    mock_get_user_by_id(id) = {"id": id, "email": "user@example.com", "tenant_id": "tenant-1"} {
        id == "user-1"
    }
    
    # Configurar entrada para teste
    input := {
        "http_method": "PATCH",
        "tenant_id": "tenant-1",
        "user": {
            "id": "tenant-admin-user",
            "roles": ["TENANT_ADMIN"],
            "tenant_id": "tenant-1"
        },
        "resource": {
            "path": "/users/user-1/roles/2/expiration",
            "user_id": "user-1",
            "role_id": "2",
            "expiration": "2024-06-30T23:59:59Z"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "user_assignment.expiration_update_decision"
    }
    
    # Verificar se a decisão é permitida
    user_assignment.expiration_update_decision == true with input as input with data.innovabiz.iam.role.user_assignment.get_role_by_id as mock_get_role_by_id with data.innovabiz.iam.role.user_assignment.get_user_by_id as mock_get_user_by_id
}

test_tenant_admin_cannot_update_system_role_expiration {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "SYSTEM_ADMIN", "type": "SYSTEM"} {
        id == "1"
    }
    
    mock_get_user_by_id(id) = {"id": id, "email": "user@example.com", "tenant_id": "tenant-1"} {
        id == "user-1"
    }
    
    # Configurar entrada para teste
    input := {
        "http_method": "PATCH",
        "tenant_id": "tenant-1",
        "user": {
            "id": "tenant-admin-user",
            "roles": ["TENANT_ADMIN"],
            "tenant_id": "tenant-1"
        },
        "resource": {
            "path": "/users/user-1/roles/1/expiration",
            "user_id": "user-1",
            "role_id": "1",
            "expiration": "2024-06-30T23:59:59Z"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "user_assignment.expiration_update_decision"
    }
    
    # Verificar se a decisão é negada
    not user_assignment.expiration_update_decision with input as input with data.innovabiz.iam.role.user_assignment.get_role_by_id as mock_get_role_by_id with data.innovabiz.iam.role.user_assignment.get_user_by_id as mock_get_user_by_id
}

test_invalid_expiration_date_is_rejected {
    # Configurar entrada para teste com data de expiração inválida (já expirada)
    input := {
        "http_method": "PATCH",
        "tenant_id": "tenant-1",
        "user": {
            "id": "super-admin-user",
            "roles": ["SUPER_ADMIN"],
            "tenant_id": constants.default_tenant
        },
        "resource": {
            "path": "/users/user-1/roles/1/expiration",
            "user_id": "user-1",
            "role_id": "1",
            "expiration": "2020-01-01T00:00:00Z"  # Data no passado
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "user_assignment.expiration_update_decision"
    }
    
    # Função auxiliar para simular validação de data
    mock_is_valid_expiration(exp) = false {
        exp == "2020-01-01T00:00:00Z"
    }
    
    # Verificar se a decisão é negada
    not user_assignment.expiration_update_decision with input as input with data.innovabiz.iam.role.user_assignment.is_valid_expiration as mock_is_valid_expiration
}