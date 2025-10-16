/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Migração inicial do banco de dados para o Identity Service
 * Implementa o esquema de dados para suportar gerenciamento de identidade e acesso
 * com isolamento multi-tenant, segurança e auditoria completa.
 */

-- Extensões
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "hstore";

-- Esquema principal
CREATE SCHEMA IF NOT EXISTS iam;

-- Configuração de Row-Level Security (RLS)
-- Habilita políticas de segurança em nível de linha para isolamento de tenant
ALTER DATABASE CURRENT SET row_security = on;

-- Tabela de Tenant (Multi-tenancy)
CREATE TABLE iam.tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255),
    display_name VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    plan VARCHAR(50) NOT NULL DEFAULT 'standard',
    subscription_id VARCHAR(255),
    max_users INTEGER NOT NULL DEFAULT 5,
    settings JSONB NOT NULL DEFAULT '{}',
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uk_tenant_name UNIQUE(name),
    CONSTRAINT uk_tenant_domain UNIQUE(domain)
);

COMMENT ON TABLE iam.tenants IS 'Armazena informações de tenant para suporte multi-tenant';

-- Tabela de Usuários
CREATE TABLE iam.users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(id),
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    phone_number VARCHAR(50),
    phone_verified BOOLEAN NOT NULL DEFAULT FALSE,
    profile_picture_url TEXT,
    locale VARCHAR(20),
    timezone VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    metadata JSONB,
    login_count INTEGER NOT NULL DEFAULT 0,
    last_login_at TIMESTAMPTZ,
    last_token_issued_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uk_user_tenant_username UNIQUE(tenant_id, username),
    CONSTRAINT uk_user_tenant_email UNIQUE(tenant_id, email)
);

CREATE INDEX idx_users_tenant_id ON iam.users(tenant_id);
CREATE INDEX idx_users_email ON iam.users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_status ON iam.users(status) WHERE deleted_at IS NULL;

COMMENT ON TABLE iam.users IS 'Armazena informações de usuários no sistema';

-- Aplicar RLS para isolamento de tenant na tabela de usuários
ALTER TABLE iam.users ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_policy ON iam.users
    USING (tenant_id = current_setting('app.tenant_id')::UUID);

-- Tabela de Credenciais de Usuário
CREATE TABLE iam.user_credentials (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    password_hash BYTEA,
    password_last_change TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    password_temp_expiry TIMESTAMPTZ,
    provider VARCHAR(50) NOT NULL DEFAULT 'local',
    provider_user_id VARCHAR(255),
    failed_attempts INTEGER NOT NULL DEFAULT 0,
    last_failed_attempt TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_credentials_user_id ON iam.user_credentials(user_id);

COMMENT ON TABLE iam.user_credentials IS 'Armazena credenciais de autenticação de usuários';

-- Tabela de Configurações MFA
CREATE TABLE iam.user_mfa_settings (
    user_id UUID PRIMARY KEY REFERENCES iam.users(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    default_method VARCHAR(50) NOT NULL DEFAULT 'none',
    methods VARCHAR(50)[] NOT NULL DEFAULT '{}',
    totp_secret BYTEA,
    phone_number VARCHAR(50),
    recovery_codes TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE iam.user_mfa_settings IS 'Armazena configurações de autenticação multi-fator';

-- Tabela de Endereços
CREATE TABLE iam.user_addresses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL DEFAULT 'residential',
    street VARCHAR(255) NOT NULL,
    number VARCHAR(50) NOT NULL,
    complement VARCHAR(255),
    district VARCHAR(255) NOT NULL,
    city VARCHAR(255) NOT NULL,
    state VARCHAR(255) NOT NULL,
    country VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_addresses_user_id ON iam.user_addresses(user_id);

COMMENT ON TABLE iam.user_addresses IS 'Armazena endereços físicos dos usuários';

-- Tabela de Contatos
CREATE TABLE iam.user_contacts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    value VARCHAR(255) NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_contacts_user_id ON iam.user_contacts(user_id);

COMMENT ON TABLE iam.user_contacts IS 'Armazena informações de contato dos usuários';

-- Tabela de Sessões
CREATE TABLE iam.user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES iam.tenants(id),
    token TEXT NOT NULL,
    refresh_token TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    refresh_expires_at TIMESTAMPTZ,
    ip_address VARCHAR(50),
    user_agent TEXT,
    device_info JSONB,
    location VARCHAR(255),
    last_activity TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_sessions_user_id ON iam.user_sessions(user_id);
CREATE INDEX idx_user_sessions_token ON iam.user_sessions(token);
CREATE INDEX idx_user_sessions_refresh_token ON iam.user_sessions(refresh_token);

COMMENT ON TABLE iam.user_sessions IS 'Armazena sessões ativas de usuários';

-- Tabela de Permissões
CREATE TABLE iam.permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(id),
    resource VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_permissions_tenant_resource_action UNIQUE(tenant_id, resource, action)
);

CREATE INDEX idx_permissions_tenant_id ON iam.permissions(tenant_id);

COMMENT ON TABLE iam.permissions IS 'Armazena permissões disponíveis no sistema';

-- Tabela de Funções/Papéis
CREATE TABLE iam.roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_roles_tenant_name UNIQUE(tenant_id, name)
);

CREATE INDEX idx_roles_tenant_id ON iam.roles(tenant_id);

COMMENT ON TABLE iam.roles IS 'Armazena funções/papéis disponíveis no sistema';

-- Tabela de Associação entre Funções e Permissões
CREATE TABLE iam.role_permissions (
    role_id UUID NOT NULL REFERENCES iam.roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES iam.permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, permission_id)
);

COMMENT ON TABLE iam.role_permissions IS 'Associação entre funções e permissões';

-- Tabela de Associação entre Usuários e Funções
CREATE TABLE iam.user_roles (
    user_id UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES iam.roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES iam.users(id),
    PRIMARY KEY (user_id, role_id)
);

COMMENT ON TABLE iam.user_roles IS 'Associação entre usuários e funções';

-- Tabela de Tokens de Recuperação de Senha
CREATE TABLE iam.password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_password_reset_tokens_token UNIQUE(token)
);

CREATE INDEX idx_password_reset_tokens_user_id ON iam.password_reset_tokens(user_id);

COMMENT ON TABLE iam.password_reset_tokens IS 'Armazena tokens para recuperação de senha';

-- Tabela de Tokens de Verificação de Email
CREATE TABLE iam.email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    token TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_email_verification_tokens_token UNIQUE(token)
);

CREATE INDEX idx_email_verification_tokens_user_id ON iam.email_verification_tokens(user_id);

COMMENT ON TABLE iam.email_verification_tokens IS 'Armazena tokens para verificação de email';

-- Tabela de Auditoria (Imutável)
CREATE TABLE iam.audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(id),
    event_type VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID NOT NULL,
    user_id UUID REFERENCES iam.users(id),
    ip_address VARCHAR(50),
    user_agent TEXT,
    old_value JSONB,
    new_value JSONB,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_tenant_id ON iam.audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_event_type ON iam.audit_logs(event_type);
CREATE INDEX idx_audit_logs_entity ON iam.audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_user_id ON iam.audit_logs(user_id);
CREATE INDEX idx_audit_logs_created_at ON iam.audit_logs(created_at);

COMMENT ON TABLE iam.audit_logs IS 'Registra todas as ações para auditoria (imutável)';

-- Funções auxiliares para auditoria
CREATE OR REPLACE FUNCTION iam.trigger_audit_log()
RETURNS TRIGGER AS $$
DECLARE
    tenant_id UUID;
    audit_data JSONB;
BEGIN
    -- Extrai o tenant_id da tabela
    IF TG_TABLE_NAME = 'tenants' THEN
        tenant_id := NEW.id;
    ELSE
        tenant_id := COALESCE(NEW.tenant_id, OLD.tenant_id);
    END IF;
    
    -- Constrói dados de auditoria
    audit_data := jsonb_build_object(
        'table', TG_TABLE_NAME,
        'operation', TG_OP,
        'schema', TG_TABLE_SCHEMA
    );
    
    -- Insere log de auditoria
    IF TG_OP = 'INSERT' THEN
        INSERT INTO iam.audit_logs (
            tenant_id, event_type, entity_type, entity_id, 
            user_id, new_value, metadata
        )
        VALUES (
            tenant_id,
            'ENTITY_CREATED',
            TG_TABLE_NAME,
            NEW.id,
            current_setting('app.current_user_id', true)::UUID,
            to_jsonb(NEW),
            audit_data
        );
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO iam.audit_logs (
            tenant_id, event_type, entity_type, entity_id, 
            user_id, old_value, new_value, metadata
        )
        VALUES (
            tenant_id,
            'ENTITY_UPDATED',
            TG_TABLE_NAME,
            NEW.id,
            current_setting('app.current_user_id', true)::UUID,
            to_jsonb(OLD),
            to_jsonb(NEW),
            audit_data
        );
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO iam.audit_logs (
            tenant_id, event_type, entity_type, entity_id, 
            user_id, old_value, metadata
        )
        VALUES (
            tenant_id,
            'ENTITY_DELETED',
            TG_TABLE_NAME,
            OLD.id,
            current_setting('app.current_user_id', true)::UUID,
            to_jsonb(OLD),
            audit_data
        );
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;