/*
 * IAM Module Database Schema - INNOVABIZ
 *
 * This script creates the database structure for the IAM module
 * supporting 70 authentication methods, regional adaptations,
 * multi-tenancy, and other framework requirements.
 *
 * @author: INNOVABIZ
 * @copyright: 2025 INNOVABIZ
 * @version: 1.0.0
 * @date: 2025-05-10
 */

-- Create schema for IAM module
CREATE SCHEMA IF NOT EXISTS iam;

COMMENT ON SCHEMA iam IS 'Schema for Identity and Access Management (IAM) module';

-- Tenants table (multi-tenancy)
CREATE TABLE iam.tenants (
    tenant_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    domain VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    region VARCHAR(10) NOT NULL DEFAULT 'EU', -- EU, BR, AO, US
    settings JSONB DEFAULT '{}'::jsonb,
    plan VARCHAR(50) DEFAULT 'standard',
    limits JSONB DEFAULT '{}'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT ck_tenant_status CHECK (status IN ('active', 'inactive', 'blocked', 'trial'))
);

COMMENT ON TABLE iam.tenants IS 'Registry of platform tenants (clients)';

-- Applications per tenant table
CREATE TABLE iam.applications (
    application_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(tenant_id),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    app_type VARCHAR(50) NOT NULL DEFAULT 'web', -- web, mobile, desktop, api
    client_id VARCHAR(100) UNIQUE NOT NULL,
    client_secret TEXT NOT NULL,
    redirect_uris TEXT[],
    allowed_origins TEXT[],
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    settings JSONB DEFAULT '{}'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT ck_application_status CHECK (status IN ('active', 'inactive', 'blocked', 'development'))
);

CREATE INDEX idx_applications_tenant ON iam.applications(tenant_id);
COMMENT ON TABLE iam.applications IS 'Registry of applications using the authentication system';

-- Users table
CREATE TABLE iam.users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(tenant_id),
    username VARCHAR(100) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    password_hash TEXT,
    full_name VARCHAR(200),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    verified_data BOOLEAN DEFAULT false,
    verified_email BOOLEAN DEFAULT false,
    verified_phone BOOLEAN DEFAULT false,
    mfa_required BOOLEAN DEFAULT false,
    failed_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    last_password_at TIMESTAMP WITH TIME ZONE,
    preferences JSONB DEFAULT '{}'::jsonb,
    profile_data JSONB DEFAULT '{}'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT ck_user_status CHECK (status IN ('active', 'inactive', 'blocked', 'pending', 'deleted')),
    CONSTRAINT uq_user_tenant_username UNIQUE (tenant_id, username),
    CONSTRAINT uq_user_tenant_email UNIQUE (tenant_id, email)
);

CREATE INDEX idx_users_tenant ON iam.users(tenant_id);
CREATE INDEX idx_users_email ON iam.users(email) WHERE email IS NOT NULL;
CREATE INDEX idx_users_phone ON iam.users(phone) WHERE phone IS NOT NULL;

COMMENT ON TABLE iam.users IS 'Registry of platform users';

-- Password history table
CREATE TABLE iam.password_history (
    history_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES iam.users(user_id),
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL DEFAULT CURRENT_USER,
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_password_history_user ON iam.password_history(user_id);

COMMENT ON TABLE iam.password_history IS 'Stores password history to prevent reuse';

-- Authentication methods table
CREATE TABLE iam.authentication_methods (
    method_id VARCHAR(10) PRIMARY KEY, -- K01, P01, B01, etc.
    method_code VARCHAR(50) NOT NULL UNIQUE,
    name_pt VARCHAR(100) NOT NULL,
    name_en VARCHAR(100) NOT NULL,
    description_pt TEXT,
    description_en TEXT,
    category VARCHAR(50) NOT NULL, -- knowledge, possession, biometric, context, etc.
    factor VARCHAR(20) NOT NULL, -- knowledge, possession, inherence
    complexity VARCHAR(20) NOT NULL DEFAULT 'medium', -- low, medium, high, very-high
    priority INTEGER NOT NULL DEFAULT 50, -- 0-100
    implementation_wave INTEGER NOT NULL DEFAULT 1, -- 1-7
    security_level VARCHAR(20) NOT NULL DEFAULT 'medium', -- low, medium, high, very-high
    score INTEGER NOT NULL DEFAULT 50, -- 0-100
    status VARCHAR(20) NOT NULL DEFAULT 'planned', -- planned, development, active, disabled, deprecated
    settings JSONB DEFAULT '{}'::jsonb,
    regional_adaptations JSONB DEFAULT '{}'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT ck_method_status CHECK (status IN ('planned', 'development', 'active', 'disabled', 'deprecated'))
);

COMMENT ON TABLE iam.authentication_methods IS 'Catalog of available authentication methods';

-- Tenant enabled methods table
CREATE TABLE iam.tenant_methods (
    tenant_method_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(tenant_id),
    method_id VARCHAR(10) NOT NULL REFERENCES iam.authentication_methods(method_id),
    enabled_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    enabled BOOLEAN NOT NULL DEFAULT true,
    default_application BOOLEAN NOT NULL DEFAULT false,
    settings JSONB DEFAULT '{}'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT uq_tenant_method UNIQUE (tenant_id, method_id)
);

CREATE INDEX idx_tenant_methods_tenant ON iam.tenant_methods(tenant_id);
CREATE INDEX idx_tenant_methods_method ON iam.tenant_methods(method_id);

COMMENT ON TABLE iam.tenant_methods IS 'Authentication methods enabled per tenant';

-- User configured methods table
CREATE TABLE iam.user_methods (
    user_method_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES iam.users(user_id),
    method_id VARCHAR(10) NOT NULL REFERENCES iam.authentication_methods(method_id),
    registered_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    enabled BOOLEAN NOT NULL DEFAULT true,
    verified BOOLEAN NOT NULL DEFAULT false,
    preferred BOOLEAN NOT NULL DEFAULT false,
    auth_data JSONB DEFAULT '{}'::jsonb, -- Stores method-specific data (e.g., TOTP secret, device tokens)
    device_name VARCHAR(200),
    metadata JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT uq_user_method UNIQUE (user_id, method_id, device_name)
);

CREATE INDEX idx_user_methods_user ON iam.user_methods(user_id);
CREATE INDEX idx_user_methods_method ON iam.user_methods(method_id);

COMMENT ON TABLE iam.user_methods IS 'Authentication methods registered by user';

-- Sessions table
CREATE TABLE iam.sessions (
    session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES iam.users(user_id),
    refresh_token TEXT,
    client_id VARCHAR(100) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    device_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_used_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    active BOOLEAN NOT NULL DEFAULT true,
    auth_factors JSONB DEFAULT '[]'::jsonb, -- List of methods used in authentication
    auth_level VARCHAR(20) NOT NULL DEFAULT 'single_factor', -- single_factor, two_factor, multi_factor
    device_info JSONB DEFAULT '{}'::jsonb,
    location_info JSONB DEFAULT '{}'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_sessions_user ON iam.sessions(user_id);
CREATE INDEX idx_sessions_token ON iam.sessions(refresh_token) WHERE refresh_token IS NOT NULL;
CREATE INDEX idx_sessions_expiration ON iam.sessions(expires_at);

COMMENT ON TABLE iam.sessions IS 'Active user sessions';

-- Authorization (groups) table
CREATE TABLE iam.groups (
    group_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(tenant_id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    metadata JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT uq_group_tenant_name UNIQUE (tenant_id, name),
    CONSTRAINT ck_group_status CHECK (status IN ('active', 'inactive'))
);

CREATE INDEX idx_groups_tenant ON iam.groups(tenant_id);

COMMENT ON TABLE iam.groups IS 'Groups for access control';

-- User-group association table
CREATE TABLE iam.user_groups (
    user_group_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES iam.users(user_id),
    group_id UUID NOT NULL REFERENCES iam.groups(group_id),
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    assigned_by VARCHAR(100) NOT NULL DEFAULT CURRENT_USER,
    metadata JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT uq_user_group UNIQUE (user_id, group_id)
);

CREATE INDEX idx_user_groups_user ON iam.user_groups(user_id);
CREATE INDEX idx_user_groups_group ON iam.user_groups(group_id);

COMMENT ON TABLE iam.user_groups IS 'Association of users to groups';

-- Authentication attempts table
CREATE TABLE iam.authentication_attempts (
    attempt_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(tenant_id),
    user_id UUID REFERENCES iam.users(user_id),
    method_id VARCHAR(10) REFERENCES iam.authentication_methods(method_id),
    attempt_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    success BOOLEAN NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    device_id VARCHAR(255),
    device_info JSONB DEFAULT '{}'::jsonb,
    location_info JSONB DEFAULT '{}'::jsonb,
    risk_info JSONB DEFAULT '{}'::jsonb,
    details JSONB DEFAULT '{}'::jsonb,
    error_code VARCHAR(50),
    error_message TEXT
);

CREATE INDEX idx_attempts_tenant ON iam.authentication_attempts(tenant_id);
CREATE INDEX idx_attempts_user ON iam.authentication_attempts(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_attempts_method ON iam.authentication_attempts(method_id) WHERE method_id IS NOT NULL;
CREATE INDEX idx_attempts_date ON iam.authentication_attempts(attempt_at);
CREATE INDEX idx_attempts_ip ON iam.authentication_attempts(ip_address) WHERE ip_address IS NOT NULL;

COMMENT ON TABLE iam.authentication_attempts IS 'Record of authentication attempts (successful or not)';

-- Authentication challenges table
CREATE TABLE iam.authentication_challenges (
    challenge_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES iam.users(user_id),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(tenant_id),
    method_id VARCHAR(10) NOT NULL REFERENCES iam.authentication_methods(method_id),
    code VARCHAR(100) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    verified_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, verified, expired, canceled
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    created_ip VARCHAR(45),
    created_deviceid VARCHAR(255),
    context_info JSONB DEFAULT '{}'::jsonb,
    challenge_data JSONB DEFAULT '{}'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT ck_challenge_status CHECK (status IN ('pending', 'verified', 'expired', 'canceled'))
);

CREATE INDEX idx_challenges_user ON iam.authentication_challenges(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_challenges_code ON iam.authentication_challenges(code) WHERE code IS NOT NULL;
CREATE INDEX idx_challenges_expiration ON iam.authentication_challenges(expires_at);
CREATE INDEX idx_challenges_tenant ON iam.authentication_challenges(tenant_id);

COMMENT ON TABLE iam.authentication_challenges IS 'Authentication challenges (OTPs, magic links, etc.)';

-- Authentication flows table
CREATE TABLE iam.authentication_flows (
    flow_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(tenant_id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    steps JSONB NOT NULL DEFAULT '[]'::jsonb,
    adaptive BOOLEAN NOT NULL DEFAULT false,
    security_level VARCHAR(20) NOT NULL DEFAULT 'medium', -- low, medium, high, very-high
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL DEFAULT CURRENT_USER,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    metadata JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT uq_flow_tenant_name UNIQUE (tenant_id, name),
    CONSTRAINT ck_flow_status CHECK (status IN ('active', 'inactive', 'draft'))
);

CREATE INDEX idx_flows_tenant ON iam.authentication_flows(tenant_id);

COMMENT ON TABLE iam.authentication_flows IS 'Configurable authentication flow definitions';

-- Authentication policies table
CREATE TABLE iam.authentication_policies (
    policy_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(tenant_id),
    application_id UUID REFERENCES iam.applications(application_id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    rules JSONB NOT NULL DEFAULT '{}'::jsonb,
    min_risk_level VARCHAR(20) NOT NULL DEFAULT 'low', -- low, medium, high
    requires_mfa BOOLEAN NOT NULL DEFAULT false,
    allowed_methods VARCHAR(10)[] DEFAULT NULL,
    denied_methods VARCHAR(10)[] DEFAULT NULL,
    default_flow UUID REFERENCES iam.authentication_flows(flow_id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL DEFAULT CURRENT_USER,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    metadata JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT uq_policy_tenant_app_name UNIQUE (tenant_id, application_id, name) DEFERRABLE INITIALLY DEFERRED,
    CONSTRAINT ck_policy_status CHECK (status IN ('active', 'inactive', 'draft'))
);

CREATE INDEX idx_policies_tenant ON iam.authentication_policies(tenant_id);
CREATE INDEX idx_policies_application ON iam.authentication_policies(application_id) WHERE application_id IS NOT NULL;

COMMENT ON TABLE iam.authentication_policies IS 'Authentication policies for tenants and applications';

-- Temporary tokens table
CREATE TABLE iam.temporary_tokens (
    token_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token VARCHAR(255) UNIQUE NOT NULL,
    tenant_id UUID NOT NULL REFERENCES iam.tenants(tenant_id),
    user_id UUID REFERENCES iam.users(user_id),
    type VARCHAR(50) NOT NULL, -- verification, reset_password, access, invite
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    used BOOLEAN NOT NULL DEFAULT false,
    canceled BOOLEAN NOT NULL DEFAULT false,
    scope TEXT[] DEFAULT NULL,
    metadata JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT ck_token_type CHECK (type IN ('verification', 'reset_password', 'access', 'invite'))
);

CREATE INDEX idx_tokens_token ON iam.temporary_tokens(token);
CREATE INDEX idx_tokens_user ON iam.temporary_tokens(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_tokens_expiration ON iam.temporary_tokens(expires_at);
CREATE INDEX idx_tokens_tenant ON iam.temporary_tokens(tenant_id);

COMMENT ON TABLE iam.temporary_tokens IS 'Temporary tokens for password reset, invitations, verifications, etc.';

-- Trusted devices table
CREATE TABLE iam.trusted_devices (
    device_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES iam.users(user_id),
    device_name VARCHAR(200),
    device_identifier VARCHAR(255) NOT NULL,
    device_type VARCHAR(50) NOT NULL, -- desktop, mobile, tablet, other
    operating_system VARCHAR(100),
    browser VARCHAR(100),
    registered_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    trusted BOOLEAN NOT NULL DEFAULT true,
    device_info JSONB DEFAULT '{}'::jsonb,
    location_info JSONB DEFAULT '{}'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT uq_user_device UNIQUE (user_id, device_identifier)
);

CREATE INDEX idx_devices_user ON iam.trusted_devices(user_id);
CREATE INDEX idx_devices_identifier ON iam.trusted_devices(device_identifier);

COMMENT ON TABLE iam.trusted_devices IS 'Registered and trusted user devices';

-- Risk profiles table
CREATE TABLE iam.risk_profiles (
    profile_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES iam.users(user_id),
    risk_score INTEGER NOT NULL DEFAULT 0, -- 0-100
    risk_level VARCHAR(20) NOT NULL DEFAULT 'medium', -- low, medium, high
    common_locations JSONB DEFAULT '[]'::jsonb,
    common_devices JSONB DEFAULT '[]'::jsonb,
    time_patterns JSONB DEFAULT '{}'::jsonb,
    behavior_patterns JSONB DEFAULT '{}'::jsonb,
    detected_anomalies JSONB DEFAULT '[]'::jsonb,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT ck_profile_risk_level CHECK (risk_level IN ('low', 'medium', 'high'))
);

CREATE INDEX idx_profiles_user ON iam.risk_profiles(user_id);
CREATE INDEX idx_profiles_risk_level ON iam.risk_profiles(risk_level);

COMMENT ON TABLE iam.risk_profiles IS 'User risk profiles for adaptive authentication';

-- Initial data insertion for authentication methods
INSERT INTO iam.authentication_methods (
    method_id, method_code, name_pt, name_en, category, factor, complexity, 
    priority, implementation_wave, security_level, score, status
) VALUES
    ('K01', 'traditional-password', 'Senha Tradicional', 'Traditional Password', 'knowledge', 'knowledge', 'low', 90, 1, 'low', 60, 'active'),
    ('K02', 'pin', 'PIN', 'PIN', 'knowledge', 'knowledge', 'low', 85, 1, 'low', 50, 'planned'),
    ('K05', 'otp', 'Senha de Uso Único (OTP)', 'One-Time Password (OTP)', 'knowledge', 'possession', 'medium', 85, 1, 'medium', 75, 'development'),
    ('P01', 'totp-hotp', 'TOTP/HOTP', 'TOTP/HOTP', 'possession', 'possession', 'medium', 80, 1, 'medium', 80, 'planned'),
    ('P02', 'fido2-webauthn', 'FIDO2/WebAuthn', 'FIDO2/WebAuthn', 'possession', 'possession', 'high', 90, 1, 'high', 95, 'planned'),
    ('P04', 'push-notification', 'Notificação Push', 'Push Notification', 'possession', 'possession', 'medium', 85, 1, 'medium', 85, 'planned'),
    ('B01', 'fingerprint', 'Reconhecimento de Impressão Digital', 'Fingerprint Recognition', 'biometric', 'inherence', 'high', 95, 1, 'high', 90, 'planned'),
    ('B02', 'facial-recognition', 'Reconhecimento Facial', 'Facial Recognition', 'biometric', 'inherence', 'high', 90, 1, 'high', 88, 'planned'),
    ('A01', 'geolocation', 'Geolocalização', 'Geolocation', 'adaptive', 'context', 'medium', 75, 1, 'medium', 70, 'planned'),
    ('A03', 'device-recognition', 'Reconhecimento de Dispositivo', 'Device Recognition', 'adaptive', 'context', 'medium', 80, 1, 'medium', 75, 'planned');

-- Useful views creation
CREATE OR REPLACE VIEW iam.vw_active_methods AS
SELECT m.* 
FROM iam.authentication_methods m
WHERE m.status = 'active';

CREATE OR REPLACE VIEW iam.vw_users_with_mfa AS
SELECT u.user_id, u.username, u.email, u.tenant_id, COUNT(um.method_id) AS num_mfa_methods
FROM iam.users u
JOIN iam.user_methods um ON u.user_id = um.user_id
JOIN iam.authentication_methods m ON um.method_id = m.method_id
WHERE um.enabled = true AND um.verified = true
  AND m.factor != 'knowledge'
GROUP BY u.user_id, u.username, u.email, u.tenant_id
HAVING COUNT(um.method_id) > 0;

-- Useful functions
CREATE OR REPLACE FUNCTION iam.fn_verify_token(p_token VARCHAR)
RETURNS TABLE (
    token_id UUID,
    type VARCHAR,
    valid BOOLEAN,
    user_id UUID,
    tenant_id UUID,
    metadata JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        t.token_id,
        t.type,
        (NOT t.used AND NOT t.canceled AND t.expires_at > CURRENT_TIMESTAMP) AS valid,
        t.user_id,
        t.tenant_id,
        t.metadata
    FROM iam.temporary_tokens t
    WHERE t.token = p_token;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Audit and logging functions
CREATE OR REPLACE FUNCTION iam.fn_log_modification()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Audit and logging triggers
CREATE TRIGGER trg_tenants_audit
BEFORE UPDATE ON iam.tenants
FOR EACH ROW
EXECUTE FUNCTION iam.fn_log_modification();

CREATE TRIGGER trg_applications_audit
BEFORE UPDATE ON iam.applications
FOR EACH ROW
EXECUTE FUNCTION iam.fn_log_modification();

CREATE TRIGGER trg_users_audit
BEFORE UPDATE ON iam.users
FOR EACH ROW
EXECUTE FUNCTION iam.fn_log_modification();

-- Indexes to improve JSONB search
CREATE INDEX idx_users_profile_gin ON iam.users USING GIN (profile_data jsonb_path_ops);
CREATE INDEX idx_auth_attempts_risk_gin ON iam.authentication_attempts USING GIN (risk_info jsonb_path_ops);
CREATE INDEX idx_auth_flows_steps_gin ON iam.authentication_flows USING GIN (steps jsonb_path_ops);

-- Permissions and security
ALTER DEFAULT PRIVILEGES IN SCHEMA iam GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO iam_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA iam GRANT USAGE ON SEQUENCES TO iam_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA iam GRANT EXECUTE ON FUNCTIONS TO iam_app;

COMMENT ON SCHEMA iam IS 'Schema for identity and access management of the INNOVABIZ platform';
