-- INNOVABIZ - IAM Identity Federation
-- Author: Eduardo Jeremias
-- Date: 09/05/2025
-- Version: 1.0
-- Description: Esquema para federação de identidades via SAML e OIDC

-- Configuração do esquema
SET search_path TO iam, public;

-- Tipos enumerados para federação
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'identity_provider_protocol') THEN
        CREATE TYPE iam.identity_provider_protocol AS ENUM (
            'oidc', 
            'saml', 
            'oauth2',
            'ws_federation'
        );
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'identity_provider_status') THEN
        CREATE TYPE iam.identity_provider_status AS ENUM (
            'active',
            'inactive',
            'testing',
            'error'
        );
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'identity_mapping_strategy') THEN
        CREATE TYPE iam.identity_mapping_strategy AS ENUM (
            'just_in_time_provisioning',
            'pre_provisioned',
            'strict_mapping',
            'attribute_based'
        );
    END IF;
END$$;

-- Tabela de provedores de identidade
CREATE TABLE IF NOT EXISTS iam.identity_providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    protocol iam.identity_provider_protocol NOT NULL,
    issuer_url VARCHAR(255) NOT NULL,
    metadata_url VARCHAR(255),
    client_id VARCHAR(255),
    client_secret TEXT,
    certificate TEXT,
    private_key TEXT,
    authorization_endpoint VARCHAR(255),
    token_endpoint VARCHAR(255),
    userinfo_endpoint VARCHAR(255),
    jwks_uri VARCHAR(255),
    end_session_endpoint VARCHAR(255),
    status iam.identity_provider_status NOT NULL DEFAULT 'inactive',
    mapping_strategy iam.identity_mapping_strategy NOT NULL DEFAULT 'just_in_time_provisioning',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_verified_at TIMESTAMP WITH TIME ZONE,
    config_metadata JSONB DEFAULT '{}'::JSONB,
    UNIQUE(organization_id, name),
    UNIQUE(organization_id, issuer_url)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_identity_providers_organization_id ON iam.identity_providers(organization_id);
CREATE INDEX IF NOT EXISTS idx_identity_providers_protocol ON iam.identity_providers(protocol);
CREATE INDEX IF NOT EXISTS idx_identity_providers_status ON iam.identity_providers(status);

-- Tabela de mapeamento de atributos do provedor de identidade
CREATE TABLE IF NOT EXISTS iam.identity_provider_attribute_mappings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider_id UUID NOT NULL REFERENCES iam.identity_providers(id),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    external_attribute VARCHAR(255) NOT NULL,
    internal_attribute VARCHAR(255) NOT NULL,
    transformation_expression TEXT,
    is_required BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(provider_id, external_attribute)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_identity_provider_attribute_mappings_provider_id ON iam.identity_provider_attribute_mappings(provider_id);
CREATE INDEX IF NOT EXISTS idx_identity_provider_attribute_mappings_organization_id ON iam.identity_provider_attribute_mappings(organization_id);

-- Tabela de mapeamento de papéis do provedor de identidade
CREATE TABLE IF NOT EXISTS iam.identity_provider_role_mappings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider_id UUID NOT NULL REFERENCES iam.identity_providers(id),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    external_role VARCHAR(255) NOT NULL,
    internal_role_id UUID NOT NULL REFERENCES iam.detailed_roles(id),
    mapping_condition JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(provider_id, external_role, internal_role_id)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_identity_provider_role_mappings_provider_id ON iam.identity_provider_role_mappings(provider_id);
CREATE INDEX IF NOT EXISTS idx_identity_provider_role_mappings_organization_id ON iam.identity_provider_role_mappings(organization_id);
CREATE INDEX IF NOT EXISTS idx_identity_provider_role_mappings_internal_role_id ON iam.identity_provider_role_mappings(internal_role_id);

-- Tabela de identidades federadas do usuário
CREATE TABLE IF NOT EXISTS iam.federated_identities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam.users(id),
    provider_id UUID NOT NULL REFERENCES iam.identity_providers(id),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    external_id VARCHAR(255) NOT NULL,
    external_username VARCHAR(255),
    external_email VARCHAR(255),
    last_login TIMESTAMP WITH TIME ZONE,
    external_data JSONB DEFAULT '{}'::JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(provider_id, external_id)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_federated_identities_user_id ON iam.federated_identities(user_id);
CREATE INDEX IF NOT EXISTS idx_federated_identities_provider_id ON iam.federated_identities(provider_id);
CREATE INDEX IF NOT EXISTS idx_federated_identities_organization_id ON iam.federated_identities(organization_id);
CREATE INDEX IF NOT EXISTS idx_federated_identities_external_id ON iam.federated_identities(external_id);
CREATE INDEX IF NOT EXISTS idx_federated_identities_external_email ON iam.federated_identities(external_email);

-- Tabela de grupos federados
CREATE TABLE IF NOT EXISTS iam.federated_groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider_id UUID NOT NULL REFERENCES iam.identity_providers(id),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    external_group_id VARCHAR(255) NOT NULL,
    external_group_name VARCHAR(255) NOT NULL,
    internal_role_id UUID REFERENCES iam.detailed_roles(id),
    auto_role_assignment BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(provider_id, external_group_id)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_federated_groups_provider_id ON iam.federated_groups(provider_id);
CREATE INDEX IF NOT EXISTS idx_federated_groups_organization_id ON iam.federated_groups(organization_id);
CREATE INDEX IF NOT EXISTS idx_federated_groups_internal_role_id ON iam.federated_groups(internal_role_id);

-- Tabela de relação entre usuários e grupos federados
CREATE TABLE IF NOT EXISTS iam.federated_user_groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    federated_identity_id UUID NOT NULL REFERENCES iam.federated_identities(id),
    federated_group_id UUID NOT NULL REFERENCES iam.federated_groups(id),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(federated_identity_id, federated_group_id)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_federated_user_groups_federated_identity_id ON iam.federated_user_groups(federated_identity_id);
CREATE INDEX IF NOT EXISTS idx_federated_user_groups_federated_group_id ON iam.federated_user_groups(federated_group_id);
CREATE INDEX IF NOT EXISTS idx_federated_user_groups_organization_id ON iam.federated_user_groups(organization_id);

-- Tabela de sessões de federação
CREATE TABLE IF NOT EXISTS iam.federation_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam.users(id),
    provider_id UUID NOT NULL REFERENCES iam.identity_providers(id),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    session_token VARCHAR(255) NOT NULL UNIQUE,
    external_session_id VARCHAR(255),
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    revoked_reason VARCHAR(100),
    metadata JSONB DEFAULT '{}'::JSONB
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_federation_sessions_user_id ON iam.federation_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_federation_sessions_provider_id ON iam.federation_sessions(provider_id);
CREATE INDEX IF NOT EXISTS idx_federation_sessions_organization_id ON iam.federation_sessions(organization_id);
CREATE INDEX IF NOT EXISTS idx_federation_sessions_session_token ON iam.federation_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_federation_sessions_expires_at ON iam.federation_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_federation_sessions_revoked ON iam.federation_sessions(revoked);
