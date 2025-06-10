-- INNOVABIZ - IAM MFA Authentication Framework (Parte 1)
-- Author: Eduardo Jeremias
-- Date: 09/05/2025
-- Version: 1.0
-- Description: Implementação do sistema de autenticação multi-fator

-- Configuração do esquema
SET search_path TO iam, public;

-- Tipos enumerados para métodos MFA
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'mfa_method_type') THEN
        CREATE TYPE iam.mfa_method_type AS ENUM (
            'totp', 
            'sms', 
            'email', 
            'push_notification',
            'biometric',
            'security_key',
            'backup_codes',
            'ar_spatial_gesture',
            'ar_gaze_pattern',
            'ar_spatial_password'
        );
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'mfa_status') THEN
        CREATE TYPE iam.mfa_status AS ENUM (
            'enabled',
            'disabled',
            'pending_activation',
            'suspended'
        );
    END IF;
END$$;

-- Tabela para configurações de MFA por organização
CREATE TABLE IF NOT EXISTS iam.mfa_organization_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    required_for_all BOOLEAN DEFAULT FALSE,
    allowed_methods iam.mfa_method_type[] NOT NULL DEFAULT ARRAY['totp', 'email']::iam.mfa_method_type[],
    min_required_methods INTEGER DEFAULT 1,
    remember_device_days INTEGER DEFAULT 30,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    custom_settings JSONB DEFAULT '{}'::JSONB,
    UNIQUE(organization_id)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_mfa_org_settings_organization_id ON iam.mfa_organization_settings(organization_id);

-- Tabela para configurações de MFA por usuário
CREATE TABLE IF NOT EXISTS iam.user_mfa_methods (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam.users(id),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    method_type iam.mfa_method_type NOT NULL,
    status iam.mfa_status NOT NULL DEFAULT 'pending_activation',
    name VARCHAR(100),
    secret TEXT, -- Encrypted secret
    phone_number VARCHAR(50),
    email VARCHAR(255),
    last_used TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::JSONB,
    UNIQUE(user_id, method_type, name)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_user_mfa_methods_user_id ON iam.user_mfa_methods(user_id);
CREATE INDEX IF NOT EXISTS idx_user_mfa_methods_organization_id ON iam.user_mfa_methods(organization_id);
CREATE INDEX IF NOT EXISTS idx_user_mfa_methods_method_type ON iam.user_mfa_methods(method_type);
CREATE INDEX IF NOT EXISTS idx_user_mfa_methods_status ON iam.user_mfa_methods(status);

-- Tabela para backup codes
CREATE TABLE IF NOT EXISTS iam.user_mfa_backup_codes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam.users(id),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    code_hash VARCHAR(255) NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_user_mfa_backup_codes_user_id ON iam.user_mfa_backup_codes(user_id);
CREATE INDEX IF NOT EXISTS idx_user_mfa_backup_codes_organization_id ON iam.user_mfa_backup_codes(organization_id);
CREATE INDEX IF NOT EXISTS idx_user_mfa_backup_codes_used ON iam.user_mfa_backup_codes(used);

-- Tabela para sessões MFA
CREATE TABLE IF NOT EXISTS iam.mfa_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam.users(id),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    session_token VARCHAR(255) NOT NULL UNIQUE,
    challenge_token VARCHAR(255),
    ip_address VARCHAR(45),
    user_agent TEXT,
    verified BOOLEAN DEFAULT FALSE,
    verified_method iam.mfa_method_type,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::JSONB
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_mfa_sessions_user_id ON iam.mfa_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_mfa_sessions_organization_id ON iam.mfa_sessions(organization_id);
CREATE INDEX IF NOT EXISTS idx_mfa_sessions_session_token ON iam.mfa_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_mfa_sessions_expires_at ON iam.mfa_sessions(expires_at);

-- Tabela para dispositivos confiáveis
CREATE TABLE IF NOT EXISTS iam.trusted_devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam.users(id),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    device_identifier VARCHAR(255) NOT NULL,
    device_name VARCHAR(255),
    ip_address VARCHAR(45),
    user_agent TEXT,
    last_used TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}'::JSONB,
    UNIQUE(user_id, device_identifier)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_trusted_devices_user_id ON iam.trusted_devices(user_id);
CREATE INDEX IF NOT EXISTS idx_trusted_devices_organization_id ON iam.trusted_devices(organization_id);
CREATE INDEX IF NOT EXISTS idx_trusted_devices_device_identifier ON iam.trusted_devices(device_identifier);
CREATE INDEX IF NOT EXISTS idx_trusted_devices_revoked ON iam.trusted_devices(revoked);
CREATE INDEX IF NOT EXISTS idx_trusted_devices_expires_at ON iam.trusted_devices(expires_at);
