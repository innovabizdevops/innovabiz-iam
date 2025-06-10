-- ==========================================================================
-- INNOVABIZ - IAM Core Schema
-- Versão: 1.0.0
-- Data de Criação: 14/05/2025
-- Autor: INNOVABIZ DevOps
-- Descrição: Esquema central do IAM para suporte aos métodos de autenticação
-- Regiões Suportadas: UE/Portugal, Brasil, Angola, EUA
-- ==========================================================================

-- Definição de Schema
CREATE SCHEMA IF NOT EXISTS iam_core;

-- Configuração de Extensões PostgreSQL necessárias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "jsonb_plperhook";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- ==========================================================================
-- Tabelas Fundamentais de Usuários e Identidades
-- ==========================================================================

-- Tabela de Tenants (Multi-tenancy)
CREATE TABLE iam_core.tenants (
    tenant_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_code VARCHAR(50) UNIQUE NOT NULL,
    tenant_name VARCHAR(255) NOT NULL,
    tenant_status VARCHAR(20) NOT NULL CHECK (tenant_status IN ('ACTIVE', 'INACTIVE', 'SUSPENDED', 'DELETED')),
    tenant_type VARCHAR(50) NOT NULL,
    organization_name VARCHAR(255) NOT NULL,
    primary_region_code VARCHAR(10) NOT NULL,
    supported_regions VARCHAR(10)[] NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    configurations JSONB NOT NULL DEFAULT '{}'::JSONB,
    compliance_settings JSONB NOT NULL DEFAULT '{}'::JSONB
);

-- Índices para Tabela de Tenants
CREATE INDEX idx_tenants_tenant_code ON iam_core.tenants(tenant_code);
CREATE INDEX idx_tenants_tenant_status ON iam_core.tenants(tenant_status);
CREATE INDEX idx_tenants_region_code ON iam_core.tenants(primary_region_code);

-- Tabela de Usuários 
CREATE TABLE iam_core.users (
    user_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES iam_core.tenants(tenant_id),
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    external_id VARCHAR(255),
    user_type VARCHAR(50) NOT NULL CHECK (user_type IN ('HUMAN', 'SERVICE', 'DEVICE', 'SYSTEM')),
    user_status VARCHAR(20) NOT NULL CHECK (user_status IN ('ACTIVE', 'INACTIVE', 'SUSPENDED', 'DELETED', 'PENDING_ACTIVATION')),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    display_name VARCHAR(255),
    primary_region_code VARCHAR(10) NOT NULL,
    preferred_language VARCHAR(10),
    timezone VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    attributes JSONB NOT NULL DEFAULT '{}'::JSONB,
    security_profile VARCHAR(50) NOT NULL DEFAULT 'DEFAULT',
    risk_profile VARCHAR(50) NOT NULL DEFAULT 'MEDIUM',
    irr_score INTEGER DEFAULT 3,
    UNIQUE(tenant_id, username),
    UNIQUE(tenant_id, email) 
);

-- Índices para Tabela de Usuários
CREATE INDEX idx_users_tenant_id ON iam_core.users(tenant_id);
CREATE INDEX idx_users_email ON iam_core.users(email) WHERE email IS NOT NULL;
CREATE INDEX idx_users_user_status ON iam_core.users(user_status);
CREATE INDEX idx_users_external_id ON iam_core.users(external_id) WHERE external_id IS NOT NULL;
CREATE INDEX idx_users_region_code ON iam_core.users(primary_region_code);
CREATE INDEX idx_users_security_profile ON iam_core.users(security_profile);
CREATE INDEX idx_users_risk_profile ON iam_core.users(risk_profile);

-- Tabela de Identidades Federadas
CREATE TABLE iam_core.federated_identities (
    federated_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam_core.users(user_id),
    provider_type VARCHAR(50) NOT NULL,
    provider_id VARCHAR(100) NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    provider_username VARCHAR(255),
    provider_email VARCHAR(255),
    identity_data JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_verified_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(user_id, provider_type, provider_id)
);

-- Índices para Tabela de Identidades Federadas
CREATE INDEX idx_federated_identities_user_id ON iam_core.federated_identities(user_id);
CREATE INDEX idx_federated_identities_provider ON iam_core.federated_identities(provider_type, provider_id);
CREATE INDEX idx_federated_identities_provider_userid ON iam_core.federated_identities(provider_user_id);

-- ==========================================================================
-- Tabelas de Credenciais e Fatores de Autenticação
-- ==========================================================================

-- Catálogo de Métodos de Autenticação Suportados
CREATE TABLE iam_core.authentication_methods (
    method_id VARCHAR(20) PRIMARY KEY,
    method_name VARCHAR(255) NOT NULL,
    category_id VARCHAR(10) NOT NULL,
    security_level VARCHAR(50) NOT NULL CHECK (security_level IN ('BASIC', 'INTERMEDIATE', 'ADVANCED', 'VERY_ADVANCED', 'CRITICAL')),
    irr_value VARCHAR(10) NOT NULL,
    complexity VARCHAR(50) NOT NULL CHECK (complexity IN ('LOW', 'MEDIUM', 'HIGH', 'VERY_HIGH')),
    maturity VARCHAR(50) NOT NULL CHECK (maturity IN ('EXPERIMENTAL', 'EMERGING', 'ESTABLISHED')),
    implementation_status VARCHAR(50) NOT NULL CHECK (implementation_status IN ('IMPLEMENTED', 'IN_PROGRESS', 'PLANNED', 'RESEARCH')),
    primary_use_cases TEXT[],
    description TEXT,
    technical_requirements TEXT,
    security_considerations TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Índices para Catálogo de Métodos
CREATE INDEX idx_auth_methods_category ON iam_core.authentication_methods(category_id);
CREATE INDEX idx_auth_methods_security ON iam_core.authentication_methods(security_level);
CREATE INDEX idx_auth_methods_status ON iam_core.authentication_methods(implementation_status);
CREATE INDEX idx_auth_methods_irr ON iam_core.authentication_methods(irr_value);

-- Credenciais e Fatores de Autenticação
CREATE TABLE iam_core.authentication_factors (
    factor_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam_core.users(user_id),
    method_id VARCHAR(20) NOT NULL REFERENCES iam_core.authentication_methods(method_id),
    factor_type VARCHAR(50) NOT NULL,
    factor_category VARCHAR(50) NOT NULL CHECK (factor_category IN ('KNOWLEDGE', 'POSSESSION', 'INHERENCE', 'CONTEXT', 'BEHAVIOR')),
    factor_data JSONB NOT NULL,
    credential_hash VARCHAR(1024),
    credential_salt VARCHAR(255),
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    verification_required BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_verified_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) NOT NULL CHECK (status IN ('ACTIVE', 'DISABLED', 'REVOKED', 'EXPIRED', 'PENDING_ACTIVATION'))
);

-- Índices para Fatores de Autenticação
CREATE INDEX idx_auth_factors_user_id ON iam_core.authentication_factors(user_id);
CREATE INDEX idx_auth_factors_method_id ON iam_core.authentication_factors(method_id);
CREATE INDEX idx_auth_factors_category ON iam_core.authentication_factors(factor_category);
CREATE INDEX idx_auth_factors_status ON iam_core.authentication_factors(status);
CREATE INDEX idx_auth_factors_is_primary ON iam_core.authentication_factors(is_primary);

-- Configurações dos Fatores de Autenticação por Método
CREATE TABLE iam_core.factor_configurations (
    config_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES iam_core.tenants(tenant_id),
    method_id VARCHAR(20) NOT NULL REFERENCES iam_core.authentication_methods(method_id),
    config_name VARCHAR(100) NOT NULL,
    config_parameters JSONB NOT NULL DEFAULT '{}'::JSONB,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    applies_to_user_types VARCHAR(50)[] NOT NULL DEFAULT ARRAY['HUMAN'],
    required_for_security_profiles VARCHAR(50)[] DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    UNIQUE(tenant_id, method_id, config_name)
);

-- Índices para Configurações de Fatores
CREATE INDEX idx_factor_config_tenant ON iam_core.factor_configurations(tenant_id);
CREATE INDEX idx_factor_config_method ON iam_core.factor_configurations(method_id);
CREATE INDEX idx_factor_config_enabled ON iam_core.factor_configurations(is_enabled);

-- ==========================================================================
-- Tabelas de Sessões e Tokens
-- ==========================================================================

-- Sessões de Autenticação
CREATE TABLE iam_core.sessions (
    session_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam_core.users(user_id),
    tenant_id UUID NOT NULL REFERENCES iam_core.tenants(tenant_id),
    auth_token VARCHAR(1024) UNIQUE,
    refresh_token VARCHAR(1024) UNIQUE,
    authentication_level VARCHAR(50) NOT NULL CHECK (authentication_level IN ('NONE', 'SINGLE_FACTOR', 'TWO_FACTOR', 'MULTI_FACTOR', 'ADAPTIVE')),
    session_type VARCHAR(50) NOT NULL CHECK (session_type IN ('WEB', 'MOBILE', 'API', 'DEVICE', 'FEDERATION')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_activity_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    invalidated_at TIMESTAMP WITH TIME ZONE,
    invalidation_reason VARCHAR(100),
    ip_address VARCHAR(50),
    user_agent TEXT,
    device_info JSONB,
    location_info JSONB,
    context_data JSONB,
    authentication_methods JSONB[],
    risk_score INTEGER,
    irr_context VARCHAR(10),
    session_data JSONB DEFAULT '{}'::JSONB
);

-- Índices para Sessões
CREATE INDEX idx_sessions_user_id ON iam_core.sessions(user_id);
CREATE INDEX idx_sessions_tenant_id ON iam_core.sessions(tenant_id);
CREATE INDEX idx_sessions_expires_at ON iam_core.sessions(expires_at);
CREATE INDEX idx_sessions_auth_level ON iam_core.sessions(authentication_level);
CREATE INDEX idx_sessions_auth_token ON iam_core.sessions(auth_token);
CREATE INDEX idx_sessions_refresh_token ON iam_core.sessions(refresh_token);

-- Histórico de Autenticação
CREATE TABLE iam_core.authentication_history (
    history_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam_core.users(user_id),
    tenant_id UUID NOT NULL REFERENCES iam_core.tenants(tenant_id),
    session_id UUID REFERENCES iam_core.sessions(session_id),
    method_id VARCHAR(20) REFERENCES iam_core.authentication_methods(method_id),
    factor_id UUID REFERENCES iam_core.authentication_factors(factor_id),
    authentication_result VARCHAR(20) NOT NULL CHECK (authentication_result IN ('SUCCESS', 'FAILURE', 'ABANDONED', 'LOCKED', 'CHALLENGE_ISSUED')),
    failure_reason VARCHAR(100),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ip_address VARCHAR(50),
    user_agent TEXT,
    device_info JSONB,
    location_info JSONB,
    risk_score INTEGER,
    irr_context VARCHAR(10),
    event_details JSONB
);

-- Índices para Histórico de Autenticação
CREATE INDEX idx_auth_history_user_id ON iam_core.authentication_history(user_id);
CREATE INDEX idx_auth_history_tenant_id ON iam_core.authentication_history(tenant_id);
CREATE INDEX idx_auth_history_session_id ON iam_core.authentication_history(session_id);
CREATE INDEX idx_auth_history_timestamp ON iam_core.authentication_history(timestamp);
CREATE INDEX idx_auth_history_result ON iam_core.authentication_history(authentication_result);
CREATE INDEX idx_auth_history_method ON iam_core.authentication_history(method_id);

-- ==========================================================================
-- Tabelas de Políticas e Permissões
-- ==========================================================================

-- Políticas de Autenticação
CREATE TABLE iam_core.authentication_policies (
    policy_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES iam_core.tenants(tenant_id),
    policy_name VARCHAR(100) NOT NULL,
    description TEXT,
    policy_type VARCHAR(50) NOT NULL CHECK (policy_type IN ('MFA', 'ADAPTIVE', 'RISK_BASED', 'STEP_UP', 'CONDITIONAL')),
    policy_rules JSONB NOT NULL,
    applies_to_user_types VARCHAR(50)[] NOT NULL DEFAULT ARRAY['HUMAN'],
    applies_to_security_profiles VARCHAR(50)[] DEFAULT NULL,
    applies_to_regions VARCHAR(10)[] DEFAULT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    priority INTEGER NOT NULL DEFAULT 100,
    UNIQUE(tenant_id, policy_name)
);

-- Índices para Políticas de Autenticação
CREATE INDEX idx_auth_policies_tenant_id ON iam_core.authentication_policies(tenant_id);
CREATE INDEX idx_auth_policies_type ON iam_core.authentication_policies(policy_type);
CREATE INDEX idx_auth_policies_enabled ON iam_core.authentication_policies(is_enabled);
CREATE INDEX idx_auth_policies_priority ON iam_core.authentication_policies(priority);

-- ==========================================================================
-- Funções e Triggers
-- ==========================================================================

-- Função para atualizar o timestamp de 'updated_at'
CREATE OR REPLACE FUNCTION iam_core.update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers para atualizar timestamps
CREATE TRIGGER update_tenant_timestamp
BEFORE UPDATE ON iam_core.tenants
FOR EACH ROW EXECUTE FUNCTION iam_core.update_timestamp();

CREATE TRIGGER update_user_timestamp
BEFORE UPDATE ON iam_core.users
FOR EACH ROW EXECUTE FUNCTION iam_core.update_timestamp();

CREATE TRIGGER update_federated_identity_timestamp
BEFORE UPDATE ON iam_core.federated_identities
FOR EACH ROW EXECUTE FUNCTION iam_core.update_timestamp();

CREATE TRIGGER update_auth_method_timestamp
BEFORE UPDATE ON iam_core.authentication_methods
FOR EACH ROW EXECUTE FUNCTION iam_core.update_timestamp();

CREATE TRIGGER update_auth_factor_timestamp
BEFORE UPDATE ON iam_core.authentication_factors
FOR EACH ROW EXECUTE FUNCTION iam_core.update_timestamp();

CREATE TRIGGER update_factor_config_timestamp
BEFORE UPDATE ON iam_core.factor_configurations
FOR EACH ROW EXECUTE FUNCTION iam_core.update_timestamp();

CREATE TRIGGER update_auth_policy_timestamp
BEFORE UPDATE ON iam_core.authentication_policies
FOR EACH ROW EXECUTE FUNCTION iam_core.update_timestamp();

-- ==========================================================================
-- Comentários de Documentação
-- ==========================================================================

COMMENT ON SCHEMA iam_core IS 'Esquema central do IAM para INNOVABIZ';
COMMENT ON TABLE iam_core.tenants IS 'Registro de tenants para suportar arquitetura multi-tenant';
COMMENT ON TABLE iam_core.users IS 'Usuários, serviços e dispositivos que requerem autenticação';
COMMENT ON TABLE iam_core.federated_identities IS 'Identidades federadas externas vinculadas a usuários INNOVABIZ';
COMMENT ON TABLE iam_core.authentication_methods IS 'Catálogo de métodos de autenticação suportados pela plataforma';
COMMENT ON TABLE iam_core.authentication_factors IS 'Fatores de autenticação registrados para usuários';
COMMENT ON TABLE iam_core.factor_configurations IS 'Configurações específicas para métodos de autenticação por tenant';
COMMENT ON TABLE iam_core.sessions IS 'Sessões ativas de usuários autenticados';
COMMENT ON TABLE iam_core.authentication_history IS 'Histórico de tentativas de autenticação e resultados';
COMMENT ON TABLE iam_core.authentication_policies IS 'Políticas e regras para requisitos de autenticação';

-- ==========================================================================
-- Fim do Script
-- ==========================================================================
