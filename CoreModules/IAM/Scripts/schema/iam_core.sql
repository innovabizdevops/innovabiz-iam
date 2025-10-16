-- ====================================================================================
-- Schema: iam (Identity and Access Management)
-- Description: Módulo central para gestão de utilizadores, papéis e permissões.
-- ====================================================================================

CREATE SCHEMA IF NOT EXISTS iam;

-- Ativar a extensão pgcrypto para hashing de senhas
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 1. Tabela de Utilizadores
CREATE TABLE IF NOT EXISTS iam.users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL, -- Armazena o hash da senha
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING_VERIFICATION', -- Ex: PENDING_VERIFICATION, ACTIVE, INACTIVE, SUSPENDED
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE iam.users IS 'Armazena as informações dos utilizadores do sistema.';
COMMENT ON COLUMN iam.users.password_hash IS 'Hash da senha do utilizador, gerado com pgcrypto.';

-- 2. Tabela de Papéis (Roles)
CREATE TABLE IF NOT EXISTS iam.roles (
    role_id SERIAL PRIMARY KEY,
    role_name VARCHAR(50) UNIQUE NOT NULL, -- Ex: 'admin', 'merchant_user', 'support_agent'
    description TEXT
);

COMMENT ON TABLE iam.roles IS 'Define os papéis que podem ser atribuídos aos utilizadores.';

-- 3. Tabela de Permissões
CREATE TABLE IF NOT EXISTS iam.permissions (
    permission_id SERIAL PRIMARY KEY,
    permission_name VARCHAR(100) UNIQUE NOT NULL, -- Ex: 'transactions:create', 'reports:view', 'users:manage'
    description TEXT
);

COMMENT ON TABLE iam.permissions IS 'Define as permissões granulares do sistema.';

-- 4. Tabela de Mapeamento: Papel-Permissões (Muitos-para-Muitos)
CREATE TABLE IF NOT EXISTS iam.role_permissions (
    role_id INT NOT NULL REFERENCES iam.roles(role_id) ON DELETE CASCADE,
    permission_id INT NOT NULL REFERENCES iam.permissions(permission_id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

COMMENT ON TABLE iam.role_permissions IS 'Associa permissões a papéis.';

-- 5. Tabela de Mapeamento: Utilizador-Papéis (Muitos-para-Muitos)
CREATE TABLE IF NOT EXISTS iam.user_roles (
    user_id UUID NOT NULL REFERENCES iam.users(user_id) ON DELETE CASCADE,
    role_id INT NOT NULL REFERENCES iam.roles(role_id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

COMMENT ON TABLE iam.user_roles IS 'Atribui papéis aos utilizadores.';

-- Adicionar alguns papéis e permissões padrão para começar
INSERT INTO iam.roles (role_name, description) VALUES
('super_admin', 'Acesso total a todas as funcionalidades do sistema.'),
('merchant_admin', 'Administrador de uma conta de comerciante.'),
('customer_support', 'Agente de suporte ao cliente.')
ON CONFLICT (role_name) DO NOTHING;

INSERT INTO iam.permissions (permission_name, description) VALUES
('users:create', 'Permite a criação de novos utilizadores.'),
('users:read', 'Permite a visualização de dados de utilizadores.'),
('users:update', 'Permite a atualização de dados de utilizadores.'),
('users:delete', 'Permite a remoção de utilizadores.'),
('transactions:read', 'Permite a visualização de transações.')
ON CONFLICT (permission_name) DO NOTHING;

-- ====================================================================================
-- Funções Principais do IAM
-- ====================================================================================

-- Função para registar um novo utilizador
CREATE OR REPLACE FUNCTION iam.register_user(
    p_username VARCHAR(100),
    p_email VARCHAR(255),
    p_password TEXT,
    p_first_name VARCHAR(100) DEFAULT NULL,
    p_last_name VARCHAR(100) DEFAULT NULL
)
RETURNS UUID AS $$
DECLARE
    v_user_id UUID;
BEGIN
    IF EXISTS (SELECT 1 FROM iam.users WHERE username = p_username) THEN
        RAISE EXCEPTION 'Nome de utilizador já existe.';
    END IF;

    IF EXISTS (SELECT 1 FROM iam.users WHERE email = p_email) THEN
        RAISE EXCEPTION 'E-mail já registado.';
    END IF;

    INSERT INTO iam.users (username, email, password_hash, first_name, last_name)
    VALUES (p_username, p_email, crypt(p_password, gen_salt('bf')), p_first_name, p_last_name)
    RETURNING user_id INTO v_user_id;

    RETURN v_user_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para autenticar um utilizador
CREATE OR REPLACE FUNCTION iam.authenticate_user(
    p_username VARCHAR(100),
    p_password TEXT
)
RETURNS UUID AS $$
DECLARE
    v_user_id UUID;
BEGIN
    SELECT user_id INTO v_user_id
    FROM iam.users
    WHERE username = p_username AND password_hash = crypt(p_password, password_hash);

    RETURN v_user_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para verificar se um utilizador tem uma permissão específica
CREATE OR REPLACE FUNCTION iam.check_permission(
    p_user_id UUID,
    p_permission_name VARCHAR(100)
)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1
        FROM iam.user_roles ur
        JOIN iam.role_permissions rp ON ur.role_id = rp.role_id
        JOIN iam.permissions p ON rp.permission_id = p.permission_id
        WHERE ur.user_id = p_user_id AND p.permission_name = p_permission_name
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;