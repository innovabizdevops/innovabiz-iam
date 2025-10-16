# Testes unitários para políticas de hierarquia do RoleService - INNOVABIZ Platform
#
# Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
# PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III
package innovabiz.iam.role.hierarchy_test

import data.innovabiz.iam.role.hierarchy
import data.innovabiz.iam.role.constants

# ---------------------------------------------------------
# Testes para decisão de adição de hierarquia
# ---------------------------------------------------------
test_super_admin_can_add_any_hierarchy {
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
            "path": "/roles/hierarchy",
            "parent_role_id": "1",
            "child_role_id": "2"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "hierarchy.hierarchy_addition_decision"
    }
    
    # Verificar se a decisão é permitida
    hierarchy.hierarchy_addition_decision == true with input as input
}

test_tenant_admin_can_add_hierarchy_for_custom_roles {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "CUSTOM_PARENT", "type": "CUSTOM"} {
        id == "3"
    } else = {"id": id, "name": "CUSTOM_CHILD", "type": "CUSTOM"} {
        id == "4"
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
            "path": "/roles/hierarchy",
            "parent_role_id": "3",
            "child_role_id": "4"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "hierarchy.hierarchy_addition_decision"
    }
    
    # Verificar se a decisão é permitida
    hierarchy.hierarchy_addition_decision == true with input as input with data.innovabiz.iam.role.hierarchy.get_role_by_id as mock_get_role_by_id
}

test_tenant_admin_cannot_add_hierarchy_involving_system_role {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "SYSTEM_ROLE", "type": "SYSTEM"} {
        id == "1"
    } else = {"id": id, "name": "CUSTOM_ROLE", "type": "CUSTOM"} {
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
            "path": "/roles/hierarchy",
            "parent_role_id": "1",  # Função do sistema
            "child_role_id": "2"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "hierarchy.hierarchy_addition_decision"
    }
    
    # Verificar se a decisão é negada
    not hierarchy.hierarchy_addition_decision with input as input with data.innovabiz.iam.role.hierarchy.get_role_by_id as mock_get_role_by_id
}

test_iam_admin_can_add_hierarchy_with_permission {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "CUSTOM_PARENT", "type": "CUSTOM"} {
        id == "3"
    } else = {"id": id, "name": "CUSTOM_CHILD", "type": "CUSTOM"} {
        id == "4"
    }
    
    # Configurar entrada para teste
    input := {
        "http_method": "POST",
        "tenant_id": "tenant-1",
        "user": {
            "id": "iam-admin-user",
            "roles": ["IAM_ADMIN"],
            "permissions": ["role:manage_hierarchy"],
            "tenant_id": "tenant-1"
        },
        "resource": {
            "path": "/roles/hierarchy",
            "parent_role_id": "3",
            "child_role_id": "4"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "hierarchy.hierarchy_addition_decision"
    }
    
    # Verificar se a decisão é permitida
    hierarchy.hierarchy_addition_decision == true with input as input with data.innovabiz.iam.role.hierarchy.get_role_by_id as mock_get_role_by_id
}

test_iam_admin_cannot_add_hierarchy_without_permission {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "CUSTOM_PARENT", "type": "CUSTOM"} {
        id == "3"
    } else = {"id": id, "name": "CUSTOM_CHILD", "type": "CUSTOM"} {
        id == "4"
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
            "path": "/roles/hierarchy",
            "parent_role_id": "3",
            "child_role_id": "4"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "hierarchy.hierarchy_addition_decision"
    }
    
    # Verificar se a decisão é negada
    not hierarchy.hierarchy_addition_decision with input as input with data.innovabiz.iam.role.hierarchy.get_role_by_id as mock_get_role_by_id
}

test_cannot_create_circular_hierarchy {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "CUSTOM_ROLE_A", "type": "CUSTOM"} {
        id == "3"
    } else = {"id": id, "name": "CUSTOM_ROLE_B", "type": "CUSTOM"} {
        id == "4"
    }
    
    # Função auxiliar para simular detecção de ciclo
    mock_would_create_cycle(parent_id, child_id) = true {
        parent_id == "3"
        child_id == "4"
    }
    
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
            "path": "/roles/hierarchy",
            "parent_role_id": "3",
            "child_role_id": "4"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "hierarchy.hierarchy_addition_decision"
    }
    
    # Verificar se a decisão é negada
    not hierarchy.hierarchy_addition_decision with input as input with data.innovabiz.iam.role.hierarchy.get_role_by_id as mock_get_role_by_id with data.innovabiz.iam.role.hierarchy.would_create_cycle as mock_would_create_cycle
}

test_cannot_exceed_max_hierarchy_depth {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "CUSTOM_ROLE_A", "type": "CUSTOM"} {
        id == "3"
    } else = {"id": id, "name": "CUSTOM_ROLE_B", "type": "CUSTOM"} {
        id == "4"
    }
    
    # Função auxiliar para simular verificação de profundidade
    mock_would_exceed_max_depth(parent_id, child_id) = true {
        parent_id == "3"
        child_id == "4"
    }
    
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
            "path": "/roles/hierarchy",
            "parent_role_id": "3",
            "child_role_id": "4"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "hierarchy.hierarchy_addition_decision"
    }
    
    # Verificar se a decisão é negada
    not hierarchy.hierarchy_addition_decision with input as input with data.innovabiz.iam.role.hierarchy.get_role_by_id as mock_get_role_by_id with data.innovabiz.iam.role.hierarchy.would_exceed_max_depth as mock_would_exceed_max_depth
}

# ---------------------------------------------------------
# Testes para decisão de remoção de hierarquia
# ---------------------------------------------------------
test_super_admin_can_remove_any_hierarchy {
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
            "path": "/roles/hierarchy/1/2",
            "parent_role_id": "1",
            "child_role_id": "2"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "hierarchy.hierarchy_removal_decision"
    }
    
    # Verificar se a decisão é permitida
    hierarchy.hierarchy_removal_decision == true with input as input
}

test_tenant_admin_can_remove_hierarchy_for_custom_roles {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "CUSTOM_PARENT", "type": "CUSTOM"} {
        id == "3"
    } else = {"id": id, "name": "CUSTOM_CHILD", "type": "CUSTOM"} {
        id == "4"
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
            "path": "/roles/hierarchy/3/4",
            "parent_role_id": "3",
            "child_role_id": "4"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "hierarchy.hierarchy_removal_decision"
    }
    
    # Verificar se a decisão é permitida
    hierarchy.hierarchy_removal_decision == true with input as input with data.innovabiz.iam.role.hierarchy.get_role_by_id as mock_get_role_by_id
}

test_tenant_admin_cannot_remove_hierarchy_involving_system_role {
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "SYSTEM_ROLE", "type": "SYSTEM"} {
        id == "1"
    } else = {"id": id, "name": "CUSTOM_ROLE", "type": "CUSTOM"} {
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
            "path": "/roles/hierarchy/1/2",
            "parent_role_id": "1",  # Função do sistema
            "child_role_id": "2"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "hierarchy.hierarchy_removal_decision"
    }
    
    # Verificar se a decisão é negada
    not hierarchy.hierarchy_removal_decision with input as input with data.innovabiz.iam.role.hierarchy.get_role_by_id as mock_get_role_by_id
}

# ---------------------------------------------------------
# Testes para decisão de consulta de hierarquia
# ---------------------------------------------------------
test_super_admin_can_query_any_hierarchy {
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
            "path": "/roles/hierarchy"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "hierarchy.hierarchy_query_decision"
    }
    
    # Verificar se a decisão é permitida
    hierarchy.hierarchy_query_decision == true with input as input
}

test_tenant_admin_can_query_hierarchy_in_own_tenant {
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
            "path": "/roles/hierarchy"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "hierarchy.hierarchy_query_decision"
    }
    
    # Verificar se a decisão é permitida
    hierarchy.hierarchy_query_decision == true with input as input
}

test_tenant_admin_cannot_query_hierarchy_in_other_tenant {
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
            "path": "/roles/hierarchy"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "hierarchy.hierarchy_query_decision"
    }
    
    # Verificar se a decisão é negada
    not hierarchy.hierarchy_query_decision with input as input
}

test_iam_operator_can_query_hierarchy {
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
            "path": "/roles/hierarchy"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "hierarchy.hierarchy_query_decision"
    }
    
    # Verificar se a decisão é permitida
    hierarchy.hierarchy_query_decision == true with input as input
}