-- INNOVABIZ IAM Module - Identity Federation Schema
-- Version: 1.0.0
-- Date: 09/05/2025
-- Author: Eduardo Jeremias
-- Description: Creates schema and tables for advanced identity federation functionality

-- Create Identity Federation Schema
CREATE SCHEMA IF NOT EXISTS iam_federation;
COMMENT ON SCHEMA iam_federation IS 'Schema for IAM identity federation functionality';

-- Set search path
SET search_path TO iam_federation, iam, public;

-- Create Identity Providers table
CREATE TABLE identity_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(id) ON DELETE CASCADE,
    provider_name TEXT NOT NULL,
    provider_type TEXT NOT NULL CHECK (provider_type IN ('saml', 'oauth2', 'oidc', 'ldap', 'custom')),
    provider_icon TEXT,
    display_name TEXT NOT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID NOT NULL REFERENCES iam.users(id),
    updated_by UUID NOT NULL REFERENCES iam.users(id),
    metadata JSONB,
    configuration JSONB NOT NULL,
    scopes TEXT[],
    jit_provisioning BOOLEAN NOT NULL DEFAULT false,
    auto_link_accounts BOOLEAN NOT NULL DEFAULT false,
    access_control_policy TEXT DEFAULT 'permissive' CHECK (access_control_policy IN ('permissive', 'restrictive', 'custom')),
    certificate_expiration DATE,
    last_sync_date TIMESTAMPTZ,
    status TEXT DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'testing', 'error')),
    error_message TEXT,
    UNIQUE(tenant_id, provider_name)
);

COMMENT ON TABLE identity_providers IS 'External identity providers for federation';

-- Create Federated Identities table
CREATE TABLE federated_identities (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    provider_id UUID NOT NULL REFERENCES identity_providers(id) ON DELETE CASCADE,
    external_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_login_at TIMESTAMPTZ,
    metadata JSONB,
    access_token TEXT,
    refresh_token TEXT,
    token_expiration TIMESTAMPTZ,
    claims JSONB,
    status TEXT DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'blocked')),
    UNIQUE(provider_id, external_id)
);

COMMENT ON TABLE federated_identities IS 'Links between users and their external identity provider accounts';

-- Create Federation Groups table
CREATE TABLE federation_groups (
    id BIGSERIAL PRIMARY KEY,
    provider_id UUID NOT NULL REFERENCES identity_providers(id) ON DELETE CASCADE,
    external_group_id TEXT NOT NULL,
    group_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    metadata JSONB,
    UNIQUE(provider_id, external_group_id)
);

COMMENT ON TABLE federation_groups IS 'External groups from identity providers';

-- Create Group Mappings table
CREATE TABLE group_mappings (
    id BIGSERIAL PRIMARY KEY,
    federation_group_id BIGINT NOT NULL REFERENCES federation_groups(id) ON DELETE CASCADE,
    local_role_id UUID NOT NULL REFERENCES iam.roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID NOT NULL REFERENCES iam.users(id),
    mapping_type TEXT DEFAULT 'exact' CHECK (mapping_type IN ('exact', 'contains', 'regex')),
    priority INTEGER DEFAULT 100,
    UNIQUE(federation_group_id, local_role_id)
);

COMMENT ON TABLE group_mappings IS 'Mappings between external groups and local roles';

-- Create Federation Audit Log table
CREATE TABLE federation_audit_log (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    provider_id UUID NOT NULL REFERENCES identity_providers(id),
    user_id UUID REFERENCES iam.users(id),
    event_type TEXT NOT NULL,
    event_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    external_id TEXT,
    status TEXT NOT NULL CHECK (status IN ('success', 'failure', 'warning')),
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    session_id UUID,
    error_details TEXT
);

COMMENT ON TABLE federation_audit_log IS 'Audit trail for identity federation events';

-- Create SAML Configuration table
CREATE TABLE saml_configurations (
    provider_id UUID PRIMARY KEY REFERENCES identity_providers(id) ON DELETE CASCADE,
    metadata_url TEXT,
    entity_id TEXT NOT NULL,
    assertion_consumer_service_url TEXT NOT NULL,
    single_logout_service_url TEXT,
    name_id_format TEXT DEFAULT 'urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress',
    certificate TEXT NOT NULL,
    private_key TEXT,
    signature_algorithm TEXT DEFAULT 'http://www.w3.org/2001/04/xmldsig-more#rsa-sha256',
    digest_algorithm TEXT DEFAULT 'http://www.w3.org/2001/04/xmlenc#sha256',
    want_assertions_signed BOOLEAN DEFAULT true,
    allow_just_in_time_provisioning BOOLEAN DEFAULT false,
    attribute_mapping JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE saml_configurations IS 'SAML-specific configuration details for identity providers';

-- Create OAuth2 Configuration table
CREATE TABLE oauth2_configurations (
    provider_id UUID PRIMARY KEY REFERENCES identity_providers(id) ON DELETE CASCADE,
    client_id TEXT NOT NULL,
    client_secret TEXT NOT NULL,
    authorization_endpoint TEXT NOT NULL,
    token_endpoint TEXT NOT NULL,
    userinfo_endpoint TEXT,
    jwks_uri TEXT,
    revocation_endpoint TEXT,
    scope TEXT DEFAULT 'openid profile email',
    response_type TEXT DEFAULT 'code',
    grant_type TEXT DEFAULT 'authorization_code',
    redirect_uri TEXT NOT NULL,
    token_auth_method TEXT DEFAULT 'client_secret_basic',
    user_id_attribute TEXT DEFAULT 'sub',
    attribute_mapping JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE oauth2_configurations IS 'OAuth2-specific configuration details for identity providers';

-- Create OIDC Configuration table
CREATE TABLE oidc_configurations (
    provider_id UUID PRIMARY KEY REFERENCES identity_providers(id) ON DELETE CASCADE,
    issuer TEXT NOT NULL,
    discovery_url TEXT,
    client_id TEXT NOT NULL,
    client_secret TEXT NOT NULL,
    scope TEXT DEFAULT 'openid profile email',
    response_type TEXT DEFAULT 'code',
    response_mode TEXT DEFAULT 'query',
    nonce_required BOOLEAN DEFAULT true,
    acr_values TEXT,
    attribute_mapping JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE oidc_configurations IS 'OpenID Connect specific configuration details for identity providers';

-- Create LDAP Configuration table
CREATE TABLE ldap_configurations (
    provider_id UUID PRIMARY KEY REFERENCES identity_providers(id) ON DELETE CASCADE,
    host TEXT NOT NULL,
    port INTEGER NOT NULL DEFAULT 389,
    use_ssl BOOLEAN DEFAULT false,
    use_tls BOOLEAN DEFAULT false,
    bind_dn TEXT NOT NULL,
    bind_credential TEXT NOT NULL,
    users_dn TEXT NOT NULL,
    user_object_classes TEXT[] DEFAULT ARRAY['inetOrgPerson', 'organizationalPerson'],
    username_attribute TEXT DEFAULT 'uid',
    rdn_attribute TEXT DEFAULT 'uid',
    email_attribute TEXT DEFAULT 'mail',
    first_name_attribute TEXT DEFAULT 'givenName',
    last_name_attribute TEXT DEFAULT 'sn',
    groups_dn TEXT,
    group_object_classes TEXT[] DEFAULT ARRAY['groupOfNames'],
    group_name_attribute TEXT DEFAULT 'cn',
    group_member_attribute TEXT DEFAULT 'member',
    group_member_format TEXT,
    connection_timeout INTEGER DEFAULT 10000,
    read_timeout INTEGER DEFAULT 10000,
    pagination_size INTEGER DEFAULT 1000,
    referral_mode TEXT DEFAULT 'follow',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE ldap_configurations IS 'LDAP-specific configuration details for identity providers';

-- Create FIDO2/WebAuthn Configuration table
CREATE TABLE fido2_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(id) ON DELETE CASCADE,
    rp_id TEXT NOT NULL,
    rp_name TEXT NOT NULL,
    rp_icon TEXT,
    attestation_preference TEXT DEFAULT 'direct' CHECK (attestation_preference IN ('none', 'indirect', 'direct')),
    authenticator_attachment TEXT CHECK (authenticator_attachment IN ('platform', 'cross-platform', NULL)),
    require_resident_key BOOLEAN DEFAULT false,
    user_verification TEXT DEFAULT 'preferred' CHECK (user_verification IN ('required', 'preferred', 'discouraged')),
    timeout INTEGER DEFAULT 60000,
    challenge_size INTEGER DEFAULT 32,
    allowed_algorithms TEXT[] DEFAULT ARRAY['ES256', 'RS256'],
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID NOT NULL REFERENCES iam.users(id),
    updated_by UUID NOT NULL REFERENCES iam.users(id),
    is_enabled BOOLEAN DEFAULT true
);

COMMENT ON TABLE fido2_configurations IS 'FIDO2/WebAuthn configuration for passwordless authentication';

-- Create FIDO2 Credentials table
CREATE TABLE fido2_credentials (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    credential_id TEXT NOT NULL,
    public_key TEXT NOT NULL,
    attestation_type TEXT,
    attestation_format TEXT,
    aaguid TEXT,
    credential_name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at TIMESTAMPTZ,
    counter BIGINT DEFAULT 0,
    device_type TEXT,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB,
    UNIQUE(user_id, credential_id)
);

COMMENT ON TABLE fido2_credentials IS 'FIDO2 registered credentials for users';

-- Create indexes
CREATE INDEX idx_identity_providers_tenant ON identity_providers(tenant_id);
CREATE INDEX idx_federated_identities_user ON federated_identities(user_id);
CREATE INDEX idx_federated_identities_provider ON federated_identities(provider_id);
CREATE INDEX idx_federation_groups_provider ON federation_groups(provider_id);
CREATE INDEX idx_group_mappings_federation_group ON group_mappings(federation_group_id);
CREATE INDEX idx_federation_audit_log_tenant ON federation_audit_log(tenant_id);
CREATE INDEX idx_federation_audit_log_provider ON federation_audit_log(provider_id);
CREATE INDEX idx_federation_audit_log_user ON federation_audit_log(user_id);
CREATE INDEX idx_fido2_configurations_tenant ON fido2_configurations(tenant_id);
CREATE INDEX idx_fido2_credentials_user ON fido2_credentials(user_id);

-- Create view for active identity providers with configuration
CREATE OR REPLACE VIEW active_identity_providers AS
SELECT
    ip.id,
    ip.tenant_id,
    ip.provider_name,
    ip.provider_type,
    ip.display_name,
    ip.is_default,
    ip.metadata,
    CASE
        WHEN ip.provider_type = 'saml' THEN 
            jsonb_build_object(
                'type', 'saml',
                'metadata_url', sc.metadata_url,
                'entity_id', sc.entity_id,
                'acs_url', sc.assertion_consumer_service_url
            )
        WHEN ip.provider_type = 'oauth2' THEN
            jsonb_build_object(
                'type', 'oauth2',
                'client_id', oc.client_id,
                'authorization_endpoint', oc.authorization_endpoint,
                'scope', oc.scope
            )
        WHEN ip.provider_type = 'oidc' THEN
            jsonb_build_object(
                'type', 'oidc',
                'issuer', oic.issuer,
                'client_id', oic.client_id,
                'scope', oic.scope
            )
        WHEN ip.provider_type = 'ldap' THEN
            jsonb_build_object(
                'type', 'ldap',
                'host', lc.host,
                'port', lc.port,
                'use_ssl', lc.use_ssl
            )
        ELSE ip.configuration
    END AS simplified_config
FROM 
    identity_providers ip
LEFT JOIN 
    saml_configurations sc ON ip.id = sc.provider_id
LEFT JOIN 
    oauth2_configurations oc ON ip.id = oc.provider_id
LEFT JOIN 
    oidc_configurations oic ON ip.id = oic.provider_id
LEFT JOIN 
    ldap_configurations lc ON ip.id = lc.provider_id
WHERE 
    ip.is_enabled = true
    AND ip.status = 'active';

COMMENT ON VIEW active_identity_providers IS 'Active identity providers with simplified configuration information';

-- Create view for user federation details
CREATE OR REPLACE VIEW user_federation_details AS
SELECT
    u.id AS user_id,
    u.username,
    u.email,
    u.tenant_id,
    ip.id AS provider_id,
    ip.provider_name,
    ip.provider_type,
    fi.external_id,
    fi.last_login_at,
    fi.status AS federation_status,
    fi.claims,
    array_agg(DISTINCT r.name) FILTER (WHERE r.name IS NOT NULL) AS mapped_roles
FROM 
    iam.users u
JOIN 
    federated_identities fi ON u.id = fi.user_id
JOIN 
    identity_providers ip ON fi.provider_id = ip.id
LEFT JOIN 
    iam.user_roles ur ON u.id = ur.user_id
LEFT JOIN 
    iam.roles r ON ur.role_id = r.id
GROUP BY
    u.id, u.username, u.email, u.tenant_id, ip.id, ip.provider_name, ip.provider_type,
    fi.external_id, fi.last_login_at, fi.status, fi.claims;

COMMENT ON VIEW user_federation_details IS 'Detailed view of federated users with their identity providers and roles';

-- Set RLS on relevant tables
ALTER TABLE identity_providers ENABLE ROW LEVEL SECURITY;
ALTER TABLE federated_identities ENABLE ROW LEVEL SECURITY;
ALTER TABLE federation_groups ENABLE ROW LEVEL SECURITY;
ALTER TABLE group_mappings ENABLE ROW LEVEL SECURITY;
ALTER TABLE federation_audit_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE fido2_configurations ENABLE ROW LEVEL SECURITY;
ALTER TABLE fido2_credentials ENABLE ROW LEVEL SECURITY;

-- Create RLS policies
CREATE POLICY tenant_identity_providers_policy ON identity_providers
    USING (tenant_id = current_setting('app.current_tenant')::uuid OR pg_has_role(current_user, 'iam_admin', 'member'));

CREATE POLICY tenant_fido2_configurations_policy ON fido2_configurations
    USING (tenant_id = current_setting('app.current_tenant')::uuid OR pg_has_role(current_user, 'iam_admin', 'member'));

CREATE POLICY user_fido2_credentials_policy ON fido2_credentials
    USING (user_id = current_setting('app.current_user_id')::uuid OR pg_has_role(current_user, 'iam_admin', 'member'));

-- Grant permissions
GRANT USAGE ON SCHEMA iam_federation TO iam_admin;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA iam_federation TO iam_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA iam_federation TO iam_admin;
GRANT USAGE ON SCHEMA iam_federation TO iam_reader;
GRANT SELECT ON ALL TABLES IN SCHEMA iam_federation TO iam_reader;

-- Reset search path
RESET search_path;
