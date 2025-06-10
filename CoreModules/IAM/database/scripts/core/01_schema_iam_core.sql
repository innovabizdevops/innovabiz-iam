-- INNOVABIZ - IAM Core Schema
-- Author: Eduardo Jeremias
-- Date: 08/05/2025
-- Version: 1.0
-- Description: Script base para criação do esquema IAM com tabelas fundamentais.

-- Criação do esquema
CREATE SCHEMA IF NOT EXISTS iam;

-- Configurar caminho de busca
SET search_path TO iam, public;

-- Criar extensões necessárias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "hstore";
CREATE EXTENSION IF NOT EXISTS "ltree";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- Tabela de Organizações
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    industry VARCHAR(100),
    sector VARCHAR(100),
    country_code VARCHAR(3),
    region_code VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    settings JSONB DEFAULT '{}'::JSONB,
    compliance_settings JSONB DEFAULT '{}'::JSONB,
    metadata JSONB DEFAULT '{}'::JSONB
);

CREATE INDEX IF NOT EXISTS idx_organizations_industry ON organizations(industry);
CREATE INDEX IF NOT EXISTS idx_organizations_sector ON organizations(sector);
CREATE INDEX IF NOT EXISTS idx_organizations_country_code ON organizations(country_code);
CREATE INDEX IF NOT EXISTS idx_organizations_region_code ON organizations(region_code);

-- Tabela de Usuários
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES iam.organizations(id) ON DELETE RESTRICT,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    full_name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login TIMESTAMP WITH TIME ZONE,
    preferences JSONB DEFAULT '{}'::JSONB,
    created_by UUID REFERENCES iam.users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES iam.users(id) ON DELETE SET NULL,
    metadata JSONB DEFAULT '{}'::JSONB,
    CONSTRAINT users_status_valid_values CHECK (status IN ('active', 'inactive', 'suspended', 'locked'))
);

CREATE INDEX IF NOT EXISTS idx_users_organization_id ON users(organization_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

-- Tabela de Roles
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES iam.organizations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system_role BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES iam.users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES iam.users(id) ON DELETE SET NULL,
    metadata JSONB DEFAULT '{}'::JSONB,
    UNIQUE(organization_id, name)
);

CREATE INDEX IF NOT EXISTS idx_roles_organization_id ON roles(organization_id);
CREATE INDEX IF NOT EXISTS idx_roles_is_system_role ON roles(is_system_role);

-- Tabela de Permissões
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::JSONB
);

CREATE INDEX IF NOT EXISTS idx_permissions_resource ON permissions(resource);
CREATE INDEX IF NOT EXISTS idx_permissions_action ON permissions(action);

-- Tabela de Atribuição de Roles a Usuários
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES iam.roles(id) ON DELETE CASCADE,
    assigned_by UUID REFERENCES iam.users(id) ON DELETE SET NULL,
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    metadata JSONB DEFAULT '{}'::JSONB,
    UNIQUE(user_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_is_active ON user_roles(is_active);

-- Tabela de Permissões de Roles (Associação Role-Permission)
CREATE TABLE IF NOT EXISTS iam.role_permissions (
    organization_id UUID REFERENCES iam.organizations(id) ON DELETE CASCADE NOT NULL, -- Chave estrangeira para organização, parte da PK
    role_id UUID NOT NULL REFERENCES iam.roles(id) ON DELETE CASCADE, -- Chave estrangeira para roles
    permission_id UUID NOT NULL REFERENCES iam.permissions(id) ON DELETE CASCADE, -- Chave estrangeira para permissions
    assigned_at TIMESTAMPTZ DEFAULT NOW(), -- Timestamp de quando a permissão foi atribuída à role
    assigned_by UUID REFERENCES iam.users(id) ON DELETE SET NULL, -- Usuário que atribuiu a permissão
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::JSONB,
    PRIMARY KEY (organization_id, role_id, permission_id) -- Chave primária composta
);

COMMENT ON TABLE iam.role_permissions IS 'Tabela de associação entre Roles e Permissions, definindo quais permissões cada role concede dentro de uma organização.';
COMMENT ON COLUMN iam.role_permissions.organization_id IS 'Identificador da organização à qual esta atribuição de permissão de role se aplica.';
COMMENT ON COLUMN iam.role_permissions.role_id IS 'Identificador da role.';
COMMENT ON COLUMN iam.role_permissions.permission_id IS 'Identificador da permissão atribuída à role.';
COMMENT ON COLUMN iam.role_permissions.assigned_at IS 'Timestamp de quando a permissão foi efetivamente atribuída à role.';
COMMENT ON COLUMN iam.role_permissions.assigned_by IS 'Identificador do usuário que realizou a atribuição da permissão à role.';
COMMENT ON COLUMN iam.role_permissions.created_at IS 'Timestamp da criação do registo de atribuição.';
COMMENT ON COLUMN iam.role_permissions.updated_at IS 'Timestamp da última atualização do registo de atribuição.';
COMMENT ON COLUMN iam.role_permissions.metadata IS 'Metadados adicionais sobre a atribuição da permissão à role, em formato JSONB.'

CREATE INDEX IF NOT EXISTS idx_role_permissions_organization_id ON iam.role_permissions(organization_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON iam.role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON iam.role_permissions(permission_id);

-- Tabela de Sessões
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    ip_address VARCHAR(50),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::JSONB,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_is_active ON sessions(is_active);

-- Tabela de Audit Log
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES iam.organizations(id) ON DELETE RESTRICT,
    user_id UUID REFERENCES iam.users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ip_address VARCHAR(50),
    request_id VARCHAR(255),
    status VARCHAR(50) NOT NULL,
    details JSONB DEFAULT '{}'::JSONB,
    session_id UUID REFERENCES iam.sessions(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_organization_id ON audit_logs(organization_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_type ON audit_logs(resource_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_id ON audit_logs(resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_logs_status ON audit_logs(status);

-- Tabela de Política de Segurança
CREATE TABLE IF NOT EXISTS security_policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES iam.organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    policy_type VARCHAR(100) NOT NULL,
    settings JSONB NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES iam.users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES iam.users(id) ON DELETE SET NULL,
    UNIQUE(organization_id, name, policy_type)
);

CREATE INDEX IF NOT EXISTS idx_security_policies_organization_id ON security_policies(organization_id);
CREATE INDEX IF NOT EXISTS idx_security_policies_policy_type ON security_policies(policy_type);
CREATE INDEX IF NOT EXISTS idx_security_policies_is_active ON security_policies(is_active);

-- Tabela de Frameworks Regulatórios
CREATE TABLE IF NOT EXISTS regulatory_frameworks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    region VARCHAR(100),
    sector VARCHAR(100),
    version VARCHAR(50),
    effective_date DATE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::JSONB
);

CREATE INDEX IF NOT EXISTS idx_regulatory_frameworks_region ON regulatory_frameworks(region);
CREATE INDEX IF NOT EXISTS idx_regulatory_frameworks_sector ON regulatory_frameworks(sector);
CREATE INDEX IF NOT EXISTS idx_regulatory_frameworks_is_active ON regulatory_frameworks(is_active);

-- Tabela de Configuração de Validadores
CREATE TABLE IF NOT EXISTS compliance_validators (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    validator_class VARCHAR(255) NOT NULL,
    framework_id UUID NOT NULL REFERENCES iam.regulatory_frameworks(id) ON DELETE CASCADE,
    configuration JSONB DEFAULT '{}'::JSONB,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    version VARCHAR(50) NOT NULL,
    metadata JSONB DEFAULT '{}'::JSONB
);

CREATE INDEX IF NOT EXISTS idx_compliance_validators_framework_id ON compliance_validators(framework_id);
CREATE INDEX IF NOT EXISTS idx_compliance_validators_is_active ON compliance_validators(is_active);

-- Tabela de Especialistas Humanos
CREATE TABLE IF NOT EXISTS equipe_especialistas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(255) NOT NULL,
    funcao VARCHAR(100) NOT NULL,
    area_atuacao VARCHAR(100) NOT NULL,
    certificacoes TEXT,
    contato VARCHAR(255),
    disponibilidade BOOLEAN DEFAULT TRUE,
    observacoes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Adiciona campo owner (responsável humano) em compliance_validators para governança híbrida
ALTER TABLE compliance_validators ADD COLUMN IF NOT EXISTS owner VARCHAR(255);

-- Função para atualizar o timestamp 'updated_at'
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Aplicar a função aos triggers
CREATE TRIGGER update_organizations_updated_at
BEFORE UPDATE ON organizations
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_roles_updated_at
BEFORE UPDATE ON roles
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_permissions_updated_at
BEFORE UPDATE ON permissions
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_security_policies_updated_at
BEFORE UPDATE ON security_policies
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_regulatory_frameworks_updated_at
BEFORE UPDATE ON regulatory_frameworks
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_compliance_validators_updated_at
BEFORE UPDATE ON compliance_validators
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Inserir frameworks regulatórios iniciais
INSERT INTO regulatory_frameworks (code, name, description, region, sector, version, effective_date, is_active)
VALUES
    ('HIPAA', 'Health Insurance Portability and Accountability Act', 'Regulamentação dos EUA para proteção de dados de saúde', 'US', 'healthcare', '1996', '1996-08-21', TRUE),
    ('GDPR', 'General Data Protection Regulation', 'Regulamentação da UE para proteção de dados pessoais', 'EU', 'general', '2016/679', '2018-05-25', TRUE),
    ('LGPD', 'Lei Geral de Proteção de Dados', 'Regulamentação brasileira para proteção de dados pessoais', 'BR', 'general', '13.709/2018', '2020-09-18', TRUE),
    ('PNDSB', 'Política Nacional de Desenvolvimento de Sistemas de Saúde', 'Regulamentação angolana para sistemas de saúde', 'AO', 'healthcare', '2010', '2010-01-01', TRUE),
    ('ISO27001', 'ISO/IEC 27001 Information Security Management', 'Padrão internacional para gestão de segurança da informação', 'GLOBAL', 'general', '2013', '2013-10-01', TRUE),
    ('HIMSS-EMRAM', 'HIMSS Electronic Medical Record Adoption Model', 'Modelo de maturidade para registros médicos eletrônicos', 'GLOBAL', 'healthcare', '2018', '2018-01-01', TRUE),
    ('HL7-FHIR', 'HL7 Fast Healthcare Interoperability Resources', 'Padrão global para interoperabilidade de dados de saúde', 'GLOBAL', 'healthcare', 'R4', '2019-10-30', TRUE)
ON CONFLICT (code) DO UPDATE
    SET name = EXCLUDED.name,
        description = EXCLUDED.description,
        region = EXCLUDED.region,
        sector = EXCLUDED.sector,
        version = EXCLUDED.version,
        effective_date = EXCLUDED.effective_date,
        is_active = EXCLUDED.is_active,
        updated_at = NOW();

-- Inserir validadores iniciais
INSERT INTO compliance_validators (code, name, description, validator_class, framework_id, is_active, version)
VALUES
    ('HIPAA-HEALTHCARE', 'HIPAA Healthcare Validator', 'Validador para HIPAA específico para healthcare', 'backend.iam.compliance.validators.healthcare.usa_validator_hipaa.HIPAAHealthcareValidator', (SELECT id FROM regulatory_frameworks WHERE code = 'HIPAA'), TRUE, '1.0'),
    ('GDPR-HEALTHCARE', 'GDPR Healthcare Validator', 'Validador para GDPR específico para healthcare', 'backend.iam.compliance.validators.healthcare.eu_validator_gdpr.GDPRHealthcareValidator', (SELECT id FROM regulatory_frameworks WHERE code = 'GDPR'), TRUE, '1.0'),
    ('LGPD-HEALTHCARE', 'LGPD Healthcare Validator', 'Validador para LGPD específico para healthcare', 'backend.iam.compliance.validators.healthcare.brazil_validator.LGPDHealthcareValidator', (SELECT id FROM regulatory_frameworks WHERE code = 'LGPD'), TRUE, '1.0'),
    ('PNDSB-HEALTHCARE', 'PNDSB Healthcare Validator', 'Validador para PNDSB específico para healthcare', 'backend.iam.compliance.validators.healthcare.angola_validator.AngolaHealthcareValidator', (SELECT id FROM regulatory_frameworks WHERE code = 'PNDSB'), TRUE, '1.0'),
    ('FHIR-VALIDATOR-R4', 'FHIR R4 Validator', 'Validador para HL7 FHIR R4', 'backend.iam.compliance.validators.healthcare.fhir_validator.FHIRValidatorR4', (SELECT id FROM regulatory_frameworks WHERE code = 'HL7-FHIR'), TRUE, '1.0'),
    ('ISO27001-HEALTHCARE', 'ISO 27001 Healthcare Validator', 'Validador ISO 27001 específico para healthcare', 'backend.iam.compliance.validators.healthcare.iso.iso27001_validator.ISO27001HealthcareValidator', (SELECT id FROM regulatory_frameworks WHERE code = 'ISO27001'), TRUE, '1.0'),
    ('HIMSS-EMRAM', 'HIMSS EMRAM Validator', 'Validador HIMSS EMRAM para maturidade de registros médicos eletrônicos', 'backend.iam.compliance.validators.healthcare.himss.emram_validator.HIMSSEMRAMValidator', (SELECT id FROM regulatory_frameworks WHERE code = 'HIMSS-EMRAM'), TRUE, '1.0')
ON CONFLICT (code) DO UPDATE
    SET name = EXCLUDED.name,
        description = EXCLUDED.description,
        validator_class = EXCLUDED.validator_class,
        framework_id = EXCLUDED.framework_id,
        is_active = EXCLUDED.is_active,
        version = EXCLUDED.version,
        updated_at = NOW();

COMMENT ON SCHEMA iam IS 'Esquema para o módulo de Identity and Access Management (IAM) do INNOVABIZ';
COMMENT ON TABLE organizations IS 'Organizações gerenciadas pela plataforma INNOVABIZ';
COMMENT ON TABLE users IS 'Usuários do sistema com suas informações básicas e credenciais';
COMMENT ON TABLE roles IS 'Papéis que podem ser atribuídos a usuários definindo suas permissões';
COMMENT ON TABLE permissions IS 'Permissões disponíveis no sistema que podem ser agrupadas em roles';
COMMENT ON TABLE user_roles IS 'Associação entre usuários e papéis atribuídos';
COMMENT ON TABLE sessions IS 'Sessões ativas de usuários no sistema';
COMMENT ON TABLE audit_logs IS 'Registro de auditorias de ações dos usuários no sistema';
COMMENT ON TABLE security_policies IS 'Políticas de segurança configuráveis por organização';
COMMENT ON TABLE regulatory_frameworks IS 'Frameworks regulatórios suportados pelo sistema';
COMMENT ON TABLE compliance_validators IS 'Configuração dos validadores de compliance disponíveis';
