/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Migração inicial do banco de dados para o Identity Service - Parte 2
 * Triggers, índices e dados iniciais
 */

-- Aplicar triggers de auditoria
CREATE TRIGGER audit_tenants_trigger
AFTER INSERT OR UPDATE OR DELETE ON iam.tenants
FOR EACH ROW EXECUTE FUNCTION iam.trigger_audit_log();

CREATE TRIGGER audit_users_trigger
AFTER INSERT OR UPDATE OR DELETE ON iam.users
FOR EACH ROW EXECUTE FUNCTION iam.trigger_audit_log();

CREATE TRIGGER audit_roles_trigger
AFTER INSERT OR UPDATE OR DELETE ON iam.roles
FOR EACH ROW EXECUTE FUNCTION iam.trigger_audit_log();

CREATE TRIGGER audit_permissions_trigger
AFTER INSERT OR UPDATE OR DELETE ON iam.permissions
FOR EACH ROW EXECUTE FUNCTION iam.trigger_audit_log();

-- Índices para busca textual
CREATE INDEX idx_users_full_text ON iam.users
USING gin((
    setweight(to_tsvector('portuguese', coalesce(first_name, '')), 'A') ||
    setweight(to_tsvector('portuguese', coalesce(last_name, '')), 'A') ||
    setweight(to_tsvector('portuguese', coalesce(email, '')), 'B') ||
    setweight(to_tsvector('portuguese', coalesce(username, '')), 'B')
));

-- Função para atualização de timestamp automática
CREATE OR REPLACE FUNCTION iam.update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Aplicar triggers para atualização automática de timestamps
CREATE TRIGGER update_tenants_updated_at
BEFORE UPDATE ON iam.tenants
FOR EACH ROW EXECUTE FUNCTION iam.update_updated_at();

CREATE TRIGGER update_users_updated_at
BEFORE UPDATE ON iam.users
FOR EACH ROW EXECUTE FUNCTION iam.update_updated_at();

CREATE TRIGGER update_user_credentials_updated_at
BEFORE UPDATE ON iam.user_credentials
FOR EACH ROW EXECUTE FUNCTION iam.update_updated_at();

CREATE TRIGGER update_user_mfa_settings_updated_at
BEFORE UPDATE ON iam.user_mfa_settings
FOR EACH ROW EXECUTE FUNCTION iam.update_updated_at();

CREATE TRIGGER update_user_addresses_updated_at
BEFORE UPDATE ON iam.user_addresses
FOR EACH ROW EXECUTE FUNCTION iam.update_updated_at();

CREATE TRIGGER update_user_contacts_updated_at
BEFORE UPDATE ON iam.user_contacts
FOR EACH ROW EXECUTE FUNCTION iam.update_updated_at();

CREATE TRIGGER update_user_sessions_updated_at
BEFORE UPDATE ON iam.user_sessions
FOR EACH ROW EXECUTE FUNCTION iam.update_updated_at();

CREATE TRIGGER update_roles_updated_at
BEFORE UPDATE ON iam.roles
FOR EACH ROW EXECUTE FUNCTION iam.update_updated_at();

CREATE TRIGGER update_permissions_updated_at
BEFORE UPDATE ON iam.permissions
FOR EACH ROW EXECUTE FUNCTION iam.update_updated_at();

-- Função para gerar histograma de logins (para analytics)
CREATE OR REPLACE FUNCTION iam.user_login_histogram(
    p_tenant_id UUID,
    p_start_date TIMESTAMPTZ,
    p_end_date TIMESTAMPTZ,
    p_interval INTERVAL
)
RETURNS TABLE(time_bucket TIMESTAMPTZ, login_count BIGINT) AS $$
BEGIN
    RETURN QUERY
    SELECT
        date_trunc('hour', last_login_at) AS time_bucket,
        COUNT(*) AS login_count
    FROM
        iam.users
    WHERE
        tenant_id = p_tenant_id
        AND last_login_at BETWEEN p_start_date AND p_end_date
    GROUP BY
        time_bucket
    ORDER BY
        time_bucket ASC;
END;
$$ LANGUAGE plpgsql;

-- Inserção de dados iniciais para o sistema
INSERT INTO iam.tenants (
    id, name, domain, display_name, status, plan, settings
) VALUES (
    '00000000-0000-0000-0000-000000000001',
    'system',
    'system.innovabiz.com',
    'Sistema INNOVABIZ',
    'active',
    'enterprise',
    '{"features": {"advanced_security": true, "sso": true, "mfa": true}}'
);

-- Inserção de permissões do sistema
INSERT INTO iam.permissions (
    tenant_id, resource, action, description, is_system
) VALUES
    ('00000000-0000-0000-0000-000000000001', 'users', 'read', 'Ler usuários', true),
    ('00000000-0000-0000-0000-000000000001', 'users', 'create', 'Criar usuários', true),
    ('00000000-0000-0000-0000-000000000001', 'users', 'update', 'Atualizar usuários', true),
    ('00000000-0000-0000-0000-000000000001', 'users', 'delete', 'Excluir usuários', true),
    ('00000000-0000-0000-0000-000000000001', 'roles', 'read', 'Ler funções', true),
    ('00000000-0000-0000-0000-000000000001', 'roles', 'create', 'Criar funções', true),
    ('00000000-0000-0000-0000-000000000001', 'roles', 'update', 'Atualizar funções', true),
    ('00000000-0000-0000-0000-000000000001', 'roles', 'delete', 'Excluir funções', true),
    ('00000000-0000-0000-0000-000000000001', 'permissions', 'read', 'Ler permissões', true),
    ('00000000-0000-0000-0000-000000000001', 'permissions', 'create', 'Criar permissões', true),
    ('00000000-0000-0000-0000-000000000001', 'permissions', 'update', 'Atualizar permissões', true),
    ('00000000-0000-0000-0000-000000000001', 'permissions', 'delete', 'Excluir permissões', true),
    ('00000000-0000-0000-0000-000000000001', 'tenants', 'read', 'Ler tenants', true),
    ('00000000-0000-0000-0000-000000000001', 'tenants', 'create', 'Criar tenants', true),
    ('00000000-0000-0000-0000-000000000001', 'tenants', 'update', 'Atualizar tenants', true),
    ('00000000-0000-0000-0000-000000000001', 'tenants', 'delete', 'Excluir tenants', true);

-- Inserção de funções do sistema
INSERT INTO iam.roles (
    tenant_id, name, description, is_system
) VALUES
    ('00000000-0000-0000-0000-000000000001', 'admin', 'Administrador do sistema', true),
    ('00000000-0000-0000-0000-000000000001', 'user', 'Usuário comum', true),
    ('00000000-0000-0000-0000-000000000001', 'guest', 'Usuário convidado', true);

-- Associar todas as permissões à função de administrador
INSERT INTO iam.role_permissions (role_id, permission_id)
SELECT 
    r.id, p.id 
FROM 
    iam.roles r, iam.permissions p 
WHERE 
    r.name = 'admin' 
    AND r.tenant_id = '00000000-0000-0000-0000-000000000001'
    AND p.tenant_id = '00000000-0000-0000-0000-000000000001';