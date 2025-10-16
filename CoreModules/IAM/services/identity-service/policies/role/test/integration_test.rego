# Testes integrados para políticas OPA do RoleService - INNOVABIZ Platform
#
# Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
# PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III
package innovabiz.iam.role.integration_test

import data.innovabiz.iam.role
import data.innovabiz.iam.role.constants

# ---------------------------------------------------------
# Testes de integração para o fluxo completo de decisões
# ---------------------------------------------------------

# Teste integrado: Super Admin cria, lê, atualiza, exclui função
test_super_admin_full_role_lifecycle {
    # 1. Criar uma função
    create_input := {
        "http_method": "POST",
        "tenant_id": "tenant-1",
        "user": {
            "id": "super-admin-user",
            "roles": ["SUPER_ADMIN"],
            "tenant_id": constants.default_tenant,
            "ip_address": "192.168.1.100",
            "user_agent": "Mozilla/5.0"
        },
        "resource": {
            "path": "/roles",
            "data": {
                "name": "TEST_ROLE",
                "type": "SYSTEM",
                "description": "Função de teste"
            }
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-123"
        },
        "path": "crud.create_decision"
    }
    
    # 2. Ler a função
    read_input := {
        "http_method": "GET",
        "tenant_id": "tenant-1",
        "user": {
            "id": "super-admin-user",
            "roles": ["SUPER_ADMIN"],
            "tenant_id": constants.default_tenant,
            "ip_address": "192.168.1.100",
            "user_agent": "Mozilla/5.0"
        },
        "resource": {
            "path": "/roles/1",
            "id": "1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-124"
        },
        "path": "crud.read_decision"
    }
    
    # 3. Atualizar a função
    update_input := {
        "http_method": "PUT",
        "tenant_id": "tenant-1",
        "user": {
            "id": "super-admin-user",
            "roles": ["SUPER_ADMIN"],
            "tenant_id": constants.default_tenant,
            "ip_address": "192.168.1.100",
            "user_agent": "Mozilla/5.0"
        },
        "resource": {
            "path": "/roles/1",
            "id": "1",
            "data": {
                "name": "TEST_ROLE_UPDATED",
                "description": "Função de teste atualizada"
            }
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-125"
        },
        "path": "crud.update_decision"
    }
    
    # 4. Excluir a função
    delete_input := {
        "http_method": "DELETE",
        "tenant_id": "tenant-1",
        "user": {
            "id": "super-admin-user",
            "roles": ["SUPER_ADMIN"],
            "tenant_id": constants.default_tenant,
            "ip_address": "192.168.1.100",
            "user_agent": "Mozilla/5.0"
        },
        "resource": {
            "path": "/roles/1",
            "id": "1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-126"
        },
        "path": "crud.delete_decision"
    }
    
    # Verificar se todas as decisões são permitidas
    role.allow with input as create_input
    role.allow with input as read_input
    role.allow with input as update_input
    role.allow with input as delete_input
}

# Teste integrado: Tenant Admin gerencia funções customizadas em seu tenant
test_tenant_admin_manage_custom_roles {
    # Dados comuns
    tenant_id := "tenant-1"
    user := {
        "id": "tenant-admin-user",
        "roles": ["TENANT_ADMIN"],
        "tenant_id": tenant_id,
        "ip_address": "192.168.1.100",
        "user_agent": "Mozilla/5.0"
    }
    context := {
        "client_ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0",
        "request_id": "req-200"
    }
    
    # 1. Criar uma função customizada
    create_input := {
        "http_method": "POST",
        "tenant_id": tenant_id,
        "user": user,
        "resource": {
            "path": "/roles",
            "data": {
                "name": "CUSTOM_ROLE",
                "type": "CUSTOM",
                "description": "Função customizada"
            }
        },
        "context": context,
        "path": "crud.create_decision"
    }
    
    # 2. Atribuir permissão à função
    permission_input := {
        "http_method": "POST",
        "tenant_id": tenant_id,
        "user": user,
        "resource": {
            "path": "/roles/1/permissions",
            "role_id": "1",
            "permission_id": "1"
        },
        "context": context,
        "path": "permissions.permission_assignment_decision"
    }
    
    # 3. Atribuir função a um usuário
    role_assignment_input := {
        "http_method": "POST",
        "tenant_id": tenant_id,
        "user": user,
        "resource": {
            "path": "/users/user-1/roles",
            "user_id": "user-1",
            "role_id": "1",
            "expiration": "2023-12-31T23:59:59Z",
            "justification": "Necessário para projeto XYZ"
        },
        "context": context,
        "path": "user_assignment.role_assignment_decision"
    }
    
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "CUSTOM_ROLE", "type": "CUSTOM", "tenant_id": tenant_id} {
        id == "1"
    }
    
    mock_get_permission_by_id(id) = {"id": id, "name": "role:read", "description": "Permite ler funções"} {
        id == "1"
    }
    
    mock_get_user_by_id(id) = {"id": id, "email": "user@example.com", "tenant_id": tenant_id} {
        id == "user-1"
    }
    
    # Verificar se todas as decisões são permitidas
    role.allow with input as create_input
    
    role.allow with input as permission_input with data.innovabiz.iam.role.permissions.get_role_by_id as mock_get_role_by_id with data.innovabiz.iam.role.permissions.get_permission_by_id as mock_get_permission_by_id
    
    role.allow with input as role_assignment_input with data.innovabiz.iam.role.user_assignment.get_role_by_id as mock_get_role_by_id with data.innovabiz.iam.role.user_assignment.get_user_by_id as mock_get_user_by_id
}

# Teste integrado: Tentativa de violação de isolamento multitenant
test_tenant_isolation {
    # Tenant Admin do tenant-1
    user_tenant1 := {
        "id": "tenant-admin-user-1",
        "roles": ["TENANT_ADMIN"],
        "tenant_id": "tenant-1",
        "ip_address": "192.168.1.100",
        "user_agent": "Mozilla/5.0"
    }
    
    # Dados de usuário em outro tenant
    user_in_tenant2 := {
        "id": "user-tenant-2",
        "email": "user@tenant2.com",
        "tenant_id": "tenant-2"
    }
    
    # Tentar acessar usuário em outro tenant
    cross_tenant_input := {
        "http_method": "GET",
        "tenant_id": "tenant-2",  # Tenant diferente
        "user": user_tenant1,
        "resource": {
            "path": "/users/user-tenant-2/roles",
            "user_id": "user-tenant-2"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-300"
        },
        "path": "user_assignment.role_check_decision"
    }
    
    # Função auxiliar para simular obtenção de dados
    mock_get_user_by_id(id) = user_in_tenant2 {
        id == "user-tenant-2"
    }
    
    # Verificar se a decisão é negada (isolamento entre tenants)
    not role.allow with input as cross_tenant_input with data.innovabiz.iam.role.user_assignment.get_user_by_id as mock_get_user_by_id
}

# Teste integrado: IAM Admin com permissões específicas
test_iam_admin_with_permissions {
    # IAM Admin com permissões específicas
    iam_admin := {
        "id": "iam-admin-user",
        "roles": ["IAM_ADMIN"],
        "permissions": ["role:create", "role:assign_permission", "role:manage_hierarchy"],
        "tenant_id": "tenant-1",
        "ip_address": "192.168.1.100",
        "user_agent": "Mozilla/5.0"
    }
    
    # 1. Criar uma função
    create_input := {
        "http_method": "POST",
        "tenant_id": "tenant-1",
        "user": iam_admin,
        "resource": {
            "path": "/roles",
            "data": {
                "name": "NEW_CUSTOM_ROLE",
                "type": "CUSTOM",
                "description": "Nova função customizada"
            }
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-400"
        },
        "path": "crud.create_decision"
    }
    
    # 2. Adicionar hierarquia entre funções
    hierarchy_input := {
        "http_method": "POST",
        "tenant_id": "tenant-1",
        "user": iam_admin,
        "resource": {
            "path": "/roles/hierarchy",
            "parent_role_id": "2",
            "child_role_id": "3"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-401"
        },
        "path": "hierarchy.hierarchy_addition_decision"
    }
    
    # Funções auxiliares para simular obtenção de dados
    mock_get_role_by_id(id) = {"id": id, "name": "CUSTOM_PARENT", "type": "CUSTOM", "tenant_id": "tenant-1"} {
        id == "2"
    } else = {"id": id, "name": "CUSTOM_CHILD", "type": "CUSTOM", "tenant_id": "tenant-1"} {
        id == "3"
    }
    
    # Verificar se as decisões são permitidas
    role.allow with input as create_input
    role.allow with input as hierarchy_input with data.innovabiz.iam.role.hierarchy.get_role_by_id as mock_get_role_by_id
}

# Teste integrado: Usuário comum tentando acessar recursos não autorizados
test_regular_user_unauthorized_access {
    # Usuário comum com função básica
    regular_user := {
        "id": "regular-user",
        "roles": ["USER"],
        "tenant_id": "tenant-1",
        "ip_address": "192.168.1.100",
        "user_agent": "Mozilla/5.0"
    }
    
    # Tentar criar uma função
    create_attempt := {
        "http_method": "POST",
        "tenant_id": "tenant-1",
        "user": regular_user,
        "resource": {
            "path": "/roles",
            "data": {
                "name": "ATTEMPT_ROLE",
                "type": "CUSTOM",
                "description": "Tentativa de criar função"
            }
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-500"
        },
        "path": "crud.create_decision"
    }
    
    # Tentar atribuir permissão
    permission_attempt := {
        "http_method": "POST",
        "tenant_id": "tenant-1",
        "user": regular_user,
        "resource": {
            "path": "/roles/1/permissions",
            "role_id": "1",
            "permission_id": "1"
        },
        "context": {
            "client_ip": "192.168.1.100",
            "user_agent": "Mozilla/5.0",
            "request_id": "req-501"
        },
        "path": "permissions.permission_assignment_decision"
    }
    
    # Verificar se ambas as decisões são negadas
    not role.allow with input as create_attempt
    not role.allow with input as permission_attempt
}