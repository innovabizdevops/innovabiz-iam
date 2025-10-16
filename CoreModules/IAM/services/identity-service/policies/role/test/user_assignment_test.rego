# Testes unitários para políticas de atribuição de função a usuários - INNOVABIZ Platform
#
# Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
# PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III
package innovabiz.iam.role.user_assignment_test

import data.innovabiz.iam.role.user_assignment
import data.innovabiz.iam.role.constants

# ---------------------------------------------------------
# Testes para decisão de atribuição de função a usuário
# ---------------------------------------------------------
test_super_admin_can_assign_any_role {
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
            "path": "/users/user-1/roles",
            "user_id": "user-1",
            "role_id": "1",
            "expiration": "2023-12-31T23:59:59Z",
            "justification": "Necessário para projeto XYZ"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "user_assignment.role_assignment_decision"
    }
    
    # Verificar se a decisão é permitida
    user_assignment.role_assignment_decision == true with input as input
}

test_tenant_admin_can_assign_custom_role_to_user_in_own_tenant {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "CUSTOM_ROLE", "type": "CUSTOM"} {
        id == "2"
    }
    
    mock_get_user_by_id(id) = {"id": id, "email": "user@example.com", "tenant_id": "tenant-1"} {
        id == "user-1"
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
            "path": "/users/user-1/roles",
            "user_id": "user-1",
            "role_id": "2",
            "expiration": "2023-12-31T23:59:59Z",
            "justification": "Necessário para projeto XYZ"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "user_assignment.role_assignment_decision"
    }
    
    # Verificar se a decisão é permitida
    user_assignment.role_assignment_decision == true with input as input with data.innovabiz.iam.role.user_assignment.get_role_by_id as mock_get_role_by_id with data.innovabiz.iam.role.user_assignment.get_user_by_id as mock_get_user_by_id
}

test_tenant_admin_cannot_assign_system_role {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "SYSTEM_ADMIN", "type": "SYSTEM"} {
        id == "1"
    }
    
    mock_get_user_by_id(id) = {"id": id, "email": "user@example.com", "tenant_id": "tenant-1"} {
        id == "user-1"
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
            "path": "/users/user-1/roles",
            "user_id": "user-1",
            "role_id": "1",
            "expiration": "2023-12-31T23:59:59Z",
            "justification": "Necessário para projeto XYZ"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "user_assignment.role_assignment_decision"
    }
    
    # Verificar se a decisão é negada
    not user_assignment.role_assignment_decision with input as input with data.innovabiz.iam.role.user_assignment.get_role_by_id as mock_get_role_by_id with data.innovabiz.iam.role.user_assignment.get_user_by_id as mock_get_user_by_id
}

test_tenant_admin_cannot_assign_role_to_user_in_other_tenant {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "CUSTOM_ROLE", "type": "CUSTOM"} {
        id == "2"
    }
    
    mock_get_user_by_id(id) = {"id": id, "email": "user@example.com", "tenant_id": "tenant-2"} {  # Tenant diferente
        id == "user-1"
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
            "path": "/users/user-1/roles",
            "user_id": "user-1",
            "role_id": "2",
            "expiration": "2023-12-31T23:59:59Z",
            "justification": "Necessário para projeto XYZ"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "user_assignment.role_assignment_decision"
    }
    
    # Verificar se a decisão é negada
    not user_assignment.role_assignment_decision with input as input with data.innovabiz.iam.role.user_assignment.get_role_by_id as mock_get_role_by_id with data.innovabiz.iam.role.user_assignment.get_user_by_id as mock_get_user_by_id
}