-- Script de Teste para o Módulo IAM
-- Data: 19/05/2025
-- Descrição: Cria dados de teste para o módulo IAM

-- Configuração do ambiente
SET client_encoding = 'UTF8';
SET client_min_messages = warning;
SET search_path = iam, public;

-- Início da transação
BEGIN;

-- Inserir uma organização de exemplo
INSERT INTO iam.organizations (
    id, name, code, industry, sector, country_code, region_code, is_active, 
    settings, compliance_settings, metadata
) VALUES (
    '11111111-1111-1111-1111-111111111111',
    'InnovaBiz Solutions',
    'INNOVABIZ',
    'Tecnologia da Informação',
    'Desenvolvimento de Software',
    'PT',
    'Europa',
    TRUE,
    '{"theme": "light", "timezone": "Europe/Lisbon"}',
    '{"gdpr": true, "lgpd": true}',
    '{"trial": false, "plan": "enterprise"}'
);

-- Inserir um usuário administrador
INSERT INTO iam.users (
    id, organization_id, username, email, full_name, password_hash, status,
    last_login, metadata
) VALUES (
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111',
    'admin',
    'admin@innovabiz.com',
    'Administrador do Sistema',
    crypt('Admin@123', gen_salt('bf')), -- Senha: Admin@123
    'active',
    NOW(),
    '{"department": "TI", "position": "Administrador de Sistema", "is_email_verified": true, "is_phone_verified": true}'
);

-- Inserir funções básicas
INSERT INTO iam.roles (
    id, organization_id, name, description, is_system_role, metadata
) VALUES 
    ('33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 'Administrador', 'Acesso total ao sistema', TRUE, '{"is_active": true}'),
    ('44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 'Usuário', 'Usuário padrão do sistema', TRUE, '{"is_active": true}');

-- Associar usuário admin à função de administrador
INSERT INTO iam.user_roles (user_id, role_id, granted_by, granted_at) 
VALUES (
    '22222222-2222-2222-2222-222222222222',
    '33333333-3333-3333-3333-333333333333',
    '22222222-2222-2222-2222-222222222222',
    NOW()
);

-- Definir permissões para a função de administrador
UPDATE iam.roles 
SET permissions = '[
    {
        "id": "55555555-5555-5555-5555-555555555555",
        "name": "Gerenciar Usuários",
        "code": "users:manage",
        "description": "Permite gerenciar usuários",
        "resource": "users",
        "action": "manage",
        "is_active": true
    },
    {
        "id": "66666666-6666-6666-6666-666666666666",
        "name": "Visualizar Relatórios",
        "code": "reports:view",
        "description": "Permite visualizar relatórios",
        "resource": "reports",
        "action": "view",
        "is_active": true
    }
]'::jsonb
WHERE id = '33333333-3333-3333-3333-333333333333';

-- Definir permissões para a função de usuário
UPDATE iam.roles 
SET permissions = '[
    {
        "id": "66666666-6666-6666-6666-666666666666",
        "name": "Visualizar Relatórios",
        "code": "reports:view",
        "description": "Permite visualizar relatórios",
        "resource": "reports",
        "action": "view",
        "is_active": true
    }
]'::jsonb
WHERE id = '44444444-4444-4444-4444-444444444444';

-- Inserir uma política de segurança
INSERT INTO iam.security_policies (
    id, organization_id, name, description, policy_type, is_active, settings, created_by, updated_by
) VALUES (
    '77777777-7777-7777-7777-777777777777',
    '11111111-1111-1111-1111-111111111111',
    'Política de Senha Padrão',
    'Política de senha padrão para todos os usuários',
    'password',
    TRUE,
    '{"min_length": 8, "require_uppercase": true, "require_lowercase": true, "require_numbers": true, "require_special_chars": true, "max_age_days": 90, "history_size": 5, "applies_to": "all_users"}',
    '22222222-2222-2222-2222-222222222222',
    '22222222-2222-2222-2222-222222222222'
);

-- Registrar um log de auditoria de exemplo
INSERT INTO iam.audit_logs (
    id, organization_id, user_id, action, resource_type, resource_id, 
    ip_address, status, details, session_id
) VALUES (
    '88888888-8888-8888-8888-888888888888',
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    'create',
    'user',
    '22222222-2222-2222-2222-222222222222',
    '192.168.1.1',
    'success',
    '{"old_values": {}, "new_values": {"username": "admin", "email": "admin@innovabiz.com", "status": "active"}, "changed_fields": ["username", "email", "status"], "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", "source": "test_script"}',
    NULL
);

-- Commit da transação
COMMIT;

-- Mensagem de conclusão
DO $$
BEGIN
    RAISE NOTICE 'Dados de teste criados com sucesso!';
END
$$;
