-- INNOVABIZ - IAM Authorization Engine (Parte 1)
-- Author: Eduardo Jeremias
-- Date: 09/05/2025
-- Version: 1.0
-- Description: Motor de autorização híbrido RBAC/ABAC para controle de acesso granular

-- Configuração do esquema
SET search_path TO iam, public;

-- Tipos enumerados para autorização
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'permission_effect') THEN
        CREATE TYPE iam.permission_effect AS ENUM (
            'allow', 
            'deny'
        );
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'permission_scope') THEN
        CREATE TYPE iam.permission_scope AS ENUM (
            'organization',
            'application',
            'module',
            'feature',
            'resource',
            'action'
        );
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'policy_evaluation_strategy') THEN
        CREATE TYPE iam.policy_evaluation_strategy AS ENUM (
            'first_applicable',
            'deny_overrides',
            'allow_overrides',
            'deny_unless_permit',
            'permit_unless_deny'
        );
    END IF;
END$$;

-- Tabela de permissões detalhada
CREATE TABLE IF NOT EXISTS iam.detailed_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    code VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    permission_scope iam.permission_scope NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    actions VARCHAR[] NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::JSONB,
    UNIQUE(organization_id, code)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_detailed_permissions_organization_id ON iam.detailed_permissions(organization_id);
CREATE INDEX IF NOT EXISTS idx_detailed_permissions_resource_type ON iam.detailed_permissions(resource_type);
CREATE INDEX IF NOT EXISTS idx_detailed_permissions_scope ON iam.detailed_permissions(permission_scope);

-- Tabela de políticas ABAC
CREATE TABLE IF NOT EXISTS iam.attribute_policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    effect iam.permission_effect NOT NULL,
    priority INTEGER NOT NULL DEFAULT 100,
    resource_type VARCHAR(100) NOT NULL,
    resource_pattern VARCHAR(255), -- Padrão para correspondência de recursos (regex ou glob)
    action_pattern VARCHAR(255), -- Padrão para correspondência de ações
    condition_expression JSONB NOT NULL, -- Expressão de condição ABAC
    condition_attributes JSONB NOT NULL, -- Atributos usados na condição
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    metadata JSONB DEFAULT '{}'::JSONB,
    UNIQUE(organization_id, name)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_attribute_policies_organization_id ON iam.attribute_policies(organization_id);
CREATE INDEX IF NOT EXISTS idx_attribute_policies_resource_type ON iam.attribute_policies(resource_type);
CREATE INDEX IF NOT EXISTS idx_attribute_policies_effect ON iam.attribute_policies(effect);
CREATE INDEX IF NOT EXISTS idx_attribute_policies_priority ON iam.attribute_policies(priority);
CREATE INDEX IF NOT EXISTS idx_attribute_policies_is_active ON iam.attribute_policies(is_active);

-- Tabela de conjuntos de políticas
CREATE TABLE IF NOT EXISTS iam.policy_sets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    evaluation_strategy iam.policy_evaluation_strategy NOT NULL DEFAULT 'deny_overrides',
    priority INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    metadata JSONB DEFAULT '{}'::JSONB,
    UNIQUE(organization_id, name)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_policy_sets_organization_id ON iam.policy_sets(organization_id);
CREATE INDEX IF NOT EXISTS idx_policy_sets_evaluation_strategy ON iam.policy_sets(evaluation_strategy);
CREATE INDEX IF NOT EXISTS idx_policy_sets_priority ON iam.policy_sets(priority);
CREATE INDEX IF NOT EXISTS idx_policy_sets_is_active ON iam.policy_sets(is_active);

-- Tabela de associação entre políticas e conjuntos de políticas
CREATE TABLE IF NOT EXISTS iam.policy_set_policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    policy_set_id UUID NOT NULL REFERENCES iam.policy_sets(id),
    attribute_policy_id UUID NOT NULL REFERENCES iam.attribute_policies(id),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    priority INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(policy_set_id, attribute_policy_id)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_policy_set_policies_policy_set_id ON iam.policy_set_policies(policy_set_id);
CREATE INDEX IF NOT EXISTS idx_policy_set_policies_attribute_policy_id ON iam.policy_set_policies(attribute_policy_id);
CREATE INDEX IF NOT EXISTS idx_policy_set_policies_organization_id ON iam.policy_set_policies(organization_id);
CREATE INDEX IF NOT EXISTS idx_policy_set_policies_priority ON iam.policy_set_policies(priority);

-- Tabela de papéis (roles) expandida
CREATE TABLE IF NOT EXISTS iam.detailed_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    code VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_system_role BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    parent_role_id UUID REFERENCES iam.detailed_roles(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::JSONB,
    UNIQUE(organization_id, code)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_detailed_roles_organization_id ON iam.detailed_roles(organization_id);
CREATE INDEX IF NOT EXISTS idx_detailed_roles_is_system_role ON iam.detailed_roles(is_system_role);
CREATE INDEX IF NOT EXISTS idx_detailed_roles_is_active ON iam.detailed_roles(is_active);
CREATE INDEX IF NOT EXISTS idx_detailed_roles_parent_role_id ON iam.detailed_roles(parent_role_id);

-- Tabela de associação entre papéis e permissões
CREATE TABLE IF NOT EXISTS iam.role_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id UUID NOT NULL REFERENCES iam.detailed_roles(id),
    permission_id UUID NOT NULL REFERENCES iam.detailed_permissions(id),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(role_id, permission_id)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON iam.role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON iam.role_permissions(permission_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_organization_id ON iam.role_permissions(organization_id);

-- Tabela de associação entre papéis e conjuntos de políticas
CREATE TABLE IF NOT EXISTS iam.role_policy_sets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id UUID NOT NULL REFERENCES iam.detailed_roles(id),
    policy_set_id UUID NOT NULL REFERENCES iam.policy_sets(id),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(role_id, policy_set_id)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_role_policy_sets_role_id ON iam.role_policy_sets(role_id);
CREATE INDEX IF NOT EXISTS idx_role_policy_sets_policy_set_id ON iam.role_policy_sets(policy_set_id);
CREATE INDEX IF NOT EXISTS idx_role_policy_sets_organization_id ON iam.role_policy_sets(organization_id);

-- Tabela de atribuição de papéis a usuários
CREATE TABLE IF NOT EXISTS iam.user_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam.users(id),
    role_id UUID NOT NULL REFERENCES iam.detailed_roles(id),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_by UUID REFERENCES iam.users(id),
    metadata JSONB DEFAULT '{}'::JSONB,
    UNIQUE(user_id, role_id)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON iam.user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON iam.user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_organization_id ON iam.user_roles(organization_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_expires_at ON iam.user_roles(expires_at);

-- Tabela de decisões de autorização em cache
CREATE TABLE IF NOT EXISTS iam.authorization_decisions_cache (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam.users(id),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    action VARCHAR(100) NOT NULL,
    decision iam.permission_effect NOT NULL,
    decision_context JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    metadata JSONB DEFAULT '{}'::JSONB
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_authorization_decisions_cache_user_id ON iam.authorization_decisions_cache(user_id);
CREATE INDEX IF NOT EXISTS idx_authorization_decisions_cache_organization_id ON iam.authorization_decisions_cache(organization_id);
CREATE INDEX IF NOT EXISTS idx_authorization_decisions_cache_resource_lookup ON iam.authorization_decisions_cache(resource_type, resource_id, action);
CREATE INDEX IF NOT EXISTS idx_authorization_decisions_cache_expires_at ON iam.authorization_decisions_cache(expires_at);
