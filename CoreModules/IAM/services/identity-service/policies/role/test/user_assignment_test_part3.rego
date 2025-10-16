# Testes unitários para políticas de atribuição de função a usuários (Parte 3) - INNOVABIZ Platform
#
# Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
# PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III
package innovabiz.iam.role.user_assignment_test

import data.innovabiz.iam.role.user_assignment
import data.innovabiz.iam.role.constants

# ---------------------------------------------------------
# Testes para decisão de verificação de funções de usuário
# ---------------------------------------------------------
test_super_admin_can_check_any_user_roles {
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
            "path": "/users/user-1/roles",
            "user_id": "user-1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "user_assignment.role_check_decision"
    }
    
    # Verificar se a decisão é permitida
    user_assignment.role_check_decision == true with input as input
}

test_tenant_admin_can_check_user_roles_in_own_tenant {
    # Funções auxiliares para simular obtenção de dados
    mock_get_user_by_id(id) = {"id": id, "email": "user@example.com", "tenant_id": "tenant-1"} {
        id == "user-1"
    }
    
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
            "path": "/users/user-1/roles",
            "user_id": "user-1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "user_assignment.role_check_decision"
    }
    
    # Verificar se a decisão é permitida
    user_assignment.role_check_decision == true with input as input with data.innovabiz.iam.role.user_assignment.get_user_by_id as mock_get_user_by_id
}

test_tenant_admin_cannot_check_user_roles_in_other_tenant {
    # Funções auxiliares para simular obtenção de dados
    mock_get_user_by_id(id) = {"id": id, "email": "user@example.com", "tenant_id": "tenant-2"} {  # Tenant diferente
        id == "user-1"
    }
    
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
            "path": "/users/user-1/roles",
            "user_id": "user-1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "user_assignment.role_check_decision"
    }
    
    # Verificar se a decisão é negada
    not user_assignment.role_check_decision with input as input with data.innovabiz.iam.role.user_assignment.get_user_by_id as mock_get_user_by_id
}

test_iam_operator_can_check_user_roles_in_own_tenant {
    # Funções auxiliares para simular obtenção de dados
    mock_get_user_by_id(id) = {"id": id, "email": "user@example.com", "tenant_id": "tenant-1"} {
        id == "user-1"
    }
    
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
            "path": "/users/user-1/roles",
            "user_id": "user-1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "user_assignment.role_check_decision"
    }
    
    # Verificar se a decisão é permitida
    user_assignment.role_check_decision == true with input as input with data.innovabiz.iam.role.user_assignment.get_user_by_id as mock_get_user_by_id
}

test_user_can_check_own_roles {
    # Configurar entrada para teste onde o usuário consulta suas próprias funções
    input := {
        "http_method": "GET",
        "tenant_id": "tenant-1",
        "user": {
            "id": "user-1",  # Mesmo ID do recurso
            "roles": ["USER"],
            "tenant_id": "tenant-1"
        },
        "resource": {
            "path": "/users/user-1/roles",
            "user_id": "user-1"  # Mesmo ID do usuário que faz a requisição
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "user_assignment.role_check_decision"
    }
    
    # Verificar se a decisão é permitida
    user_assignment.role_check_decision == true with input as input
}

test_user_cannot_check_other_user_roles {
    # Configurar entrada para teste onde o usuário tenta consultar funções de outro usuário
    input := {
        "http_method": "GET",
        "tenant_id": "tenant-1",
        "user": {
            "id": "user-2",  # ID diferente do recurso
            "roles": ["USER"],
            "tenant_id": "tenant-1"
        },
        "resource": {
            "path": "/users/user-1/roles",
            "user_id": "user-1"  # ID diferente do usuário que faz a requisição
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "user_assignment.role_check_decision"
    }
    
    # Verificar se a decisão é negada
    not user_assignment.role_check_decision with input as input
}

test_user_with_permission_can_check_user_roles {
    # Configurar entrada para teste de usuário com permissão específica
    input := {
        "http_method": "GET",
        "tenant_id": "tenant-1",
        "user": {
            "id": "special-user",
            "roles": ["SUPPORT"],
            "permissions": ["user:roles:read"],
            "tenant_id": "tenant-1"
        },
        "resource": {
            "path": "/users/user-1/roles",
            "user_id": "user-1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "user_assignment.role_check_decision"
    }
    
    # Funções auxiliares para simular obtenção de dados
    mock_get_user_by_id(id) = {"id": id, "email": "user@example.com", "tenant_id": "tenant-1"} {
        id == "user-1"
    }
    
    # Verificar se a decisão é permitida
    user_assignment.role_check_decision == true with input as input with data.innovabiz.iam.role.user_assignment.get_user_by_id as mock_get_user_by_id
}

# ---------------------------------------------------------
# Testes para verificação específica de função de usuário
# ---------------------------------------------------------
test_super_admin_can_check_specific_user_role {
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
            "path": "/users/user-1/roles/1",
            "user_id": "user-1",
            "role_id": "1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "user_assignment.specific_role_check_decision"
    }
    
    # Verificar se a decisão é permitida
    user_assignment.specific_role_check_decision == true with input as input
}