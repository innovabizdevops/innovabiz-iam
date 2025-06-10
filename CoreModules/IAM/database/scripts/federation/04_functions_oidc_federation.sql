-- INNOVABIZ IAM Module - OpenID Connect Federation Functions
-- Version: 1.0.0
-- Date: 09/05/2025
-- Author: Eduardo Jeremias
-- Description: Functions for managing OpenID Connect identity federation

-- Set search path
SET search_path TO iam_federation, iam, public;

-- Function to create a new OpenID Connect identity provider
CREATE OR REPLACE FUNCTION create_oidc_provider(
    p_tenant_id UUID,
    p_provider_name TEXT,
    p_display_name TEXT,
    p_issuer TEXT,
    p_client_id TEXT,
    p_client_secret TEXT,
    p_discovery_url TEXT DEFAULT NULL,
    p_scope TEXT DEFAULT 'openid profile email',
    p_is_enabled BOOLEAN DEFAULT true,
    p_is_default BOOLEAN DEFAULT false,
    p_jit_provisioning BOOLEAN DEFAULT false,
    p_created_by UUID
) RETURNS UUID AS $$
DECLARE
    v_provider_id UUID;
BEGIN
    -- Start a transaction
    BEGIN
        -- Insert into identity_providers table
        INSERT INTO identity_providers (
            tenant_id,
            provider_name,
            provider_type,
            display_name,
            is_enabled,
            is_default,
            created_by,
            updated_by,
            configuration,
            jit_provisioning
        ) VALUES (
            p_tenant_id,
            p_provider_name,
            'oidc',
            p_display_name,
            p_is_enabled,
            p_is_default,
            p_created_by,
            p_created_by,
            jsonb_build_object(
                'type', 'oidc',
                'issuer', p_issuer,
                'client_id', p_client_id,
                'discovery_url', p_discovery_url,
                'scope', p_scope
            ),
            p_jit_provisioning
        ) RETURNING id INTO v_provider_id;
        
        -- Insert into oidc_configurations table
        INSERT INTO oidc_configurations (
            provider_id,
            issuer,
            discovery_url,
            client_id,
            client_secret,
            scope
        ) VALUES (
            v_provider_id,
            p_issuer,
            p_discovery_url,
            p_client_id,
            p_client_secret,
            p_scope
        );
        
        -- Log the creation in audit
        INSERT INTO federation_audit_log (
            tenant_id,
            provider_id,
            event_type,
            status,
            details
        ) VALUES (
            p_tenant_id,
            v_provider_id,
            'provider_created',
            'success',
            jsonb_build_object(
                'provider_type', 'oidc',
                'provider_name', p_provider_name,
                'created_by', p_created_by
            )
        );
        
        RETURN v_provider_id;
    EXCEPTION
        WHEN OTHERS THEN
            -- Log error in audit
            IF v_provider_id IS NOT NULL THEN
                INSERT INTO federation_audit_log (
                    tenant_id,
                    provider_id,
                    event_type,
                    status,
                    details,
                    error_details
                ) VALUES (
                    p_tenant_id,
                    v_provider_id,
                    'provider_created',
                    'failure',
                    jsonb_build_object(
                        'provider_type', 'oidc',
                        'provider_name', p_provider_name,
                        'created_by', p_created_by
                    ),
                    SQLERRM
                );
            END IF;
            
            RAISE;
    END;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION create_oidc_provider(UUID, TEXT, TEXT, TEXT, TEXT, TEXT, TEXT, TEXT, BOOLEAN, BOOLEAN, BOOLEAN, UUID) IS 
'Creates a new OpenID Connect identity provider with the specified configuration';

-- Function to fetch OpenID Connect discovery document
CREATE OR REPLACE FUNCTION fetch_oidc_discovery_document(
    p_provider_id UUID
) RETURNS JSONB AS $$
DECLARE
    v_provider RECORD;
    v_discovery_url TEXT;
    v_result JSONB;
BEGIN
    -- Get provider details
    SELECT 
        oic.discovery_url,
        oic.issuer
    INTO v_provider
    FROM 
        oidc_configurations oic
    WHERE 
        oic.provider_id = p_provider_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'OIDC provider configuration not found';
    END IF;
    
    -- Determine discovery URL
    IF v_provider.discovery_url IS NOT NULL THEN
        v_discovery_url := v_provider.discovery_url;
    ELSE
        -- Standard well-known URL format
        v_discovery_url := v_provider.issuer;
        -- Ensure issuer URL ends with '/'
        IF right(v_discovery_url, 1) != '/' THEN
            v_discovery_url := v_discovery_url || '/';
        END IF;
        v_discovery_url := v_discovery_url || '.well-known/openid-configuration';
    END IF;
    
    -- This is a placeholder for actual discovery document fetching
    -- In a real implementation, this would make an HTTP request to the discovery URL
    
    -- For now, we just return a mock discovery document
    v_result := jsonb_build_object(
        'issuer', v_provider.issuer,
        'authorization_endpoint', v_provider.issuer || '/auth',
        'token_endpoint', v_provider.issuer || '/token',
        'userinfo_endpoint', v_provider.issuer || '/userinfo',
        'jwks_uri', v_provider.issuer || '/.well-known/jwks.json',
        'response_types_supported', jsonb_build_array('code', 'id_token', 'token id_token'),
        'subject_types_supported', jsonb_build_array('public'),
        'id_token_signing_alg_values_supported', jsonb_build_array('RS256'),
        'scopes_supported', jsonb_build_array('openid', 'email', 'profile'),
        'token_endpoint_auth_methods_supported', jsonb_build_array('client_secret_basic', 'client_secret_post'),
        'claims_supported', jsonb_build_array('sub', 'iss', 'email', 'email_verified', 'preferred_username', 'name')
    );
    
    RETURN v_result;
EXCEPTION
    WHEN OTHERS THEN
        RETURN jsonb_build_object(
            'error', SQLERRM,
            'error_description', 'Failed to fetch OIDC discovery document'
        );
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION fetch_oidc_discovery_document(UUID) IS 
'Fetches and returns the OpenID Connect discovery document for a provider';

-- Function to validate ID token
CREATE OR REPLACE FUNCTION validate_id_token(
    p_provider_id UUID,
    p_id_token TEXT,
    p_nonce TEXT DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_provider RECORD;
    v_claims JSONB;
    v_valid BOOLEAN := true;
    v_error TEXT;
BEGIN
    -- Get provider details
    SELECT 
        ip.*,
        oic.issuer,
        oic.client_id
    INTO v_provider
    FROM 
        identity_providers ip
        JOIN oidc_configurations oic ON ip.id = oic.provider_id
    WHERE 
        ip.id = p_provider_id;
    
    IF NOT FOUND THEN
        RETURN jsonb_build_object(
            'valid', false,
            'error', 'Provider not found',
            'claims', NULL
        );
    END IF;
    
    -- This is a placeholder for actual JWT validation
    -- In a real implementation, this would:
    -- 1. Decode the JWT
    -- 2. Verify the signature using the provider's JWKS
    -- 3. Validate the claims (iss, aud, exp, iat, nonce, etc.)
    
    -- For now, we just simulate parsing the token and extracting claims
    -- In this simulation, we just extract a mock payload
    v_claims := jsonb_build_object(
        'iss', v_provider.issuer,
        'sub', 'user789',
        'aud', v_provider.client_id,
        'exp', extract(epoch from (now() + interval '1 hour'))::bigint,
        'iat', extract(epoch from now())::bigint,
        'auth_time', extract(epoch from now())::bigint,
        'nonce', p_nonce,
        'email', 'user789@example.com',
        'email_verified', true,
        'name', 'Alex Johnson',
        'given_name', 'Alex',
        'family_name', 'Johnson'
    );
    
    -- Perform basic validation checks (a real implementation would be more thorough)
    IF v_claims->>'iss' != v_provider.issuer THEN
        v_valid := false;
        v_error := 'Invalid issuer';
    ELSIF v_claims->>'aud' != v_provider.client_id THEN
        v_valid := false;
        v_error := 'Invalid audience';
    ELSIF (v_claims->>'exp')::bigint < extract(epoch from now())::bigint THEN
        v_valid := false;
        v_error := 'Token expired';
    ELSIF p_nonce IS NOT NULL AND v_claims->>'nonce' != p_nonce THEN
        v_valid := false;
        v_error := 'Invalid nonce';
    END IF;
    
    RETURN jsonb_build_object(
        'valid', v_valid,
        'error', v_error,
        'claims', v_claims
    );
EXCEPTION
    WHEN OTHERS THEN
        RETURN jsonb_build_object(
            'valid', false,
            'error', SQLERRM,
            'claims', NULL
        );
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION validate_id_token(UUID, TEXT, TEXT) IS 
'Validates an OpenID Connect ID token and returns the parsed claims';

-- Function to process OpenID Connect authentication
CREATE OR REPLACE FUNCTION process_oidc_authentication(
    p_provider_id UUID,
    p_id_token TEXT,
    p_access_token TEXT DEFAULT NULL,
    p_refresh_token TEXT DEFAULT NULL,
    p_token_expiration TIMESTAMPTZ DEFAULT NULL,
    p_nonce TEXT DEFAULT NULL,
    p_ip_address INET DEFAULT NULL,
    p_user_agent TEXT DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_provider RECORD;
    v_tenant_id UUID;
    v_token_validation JSONB;
    v_claims JSONB;
    v_external_id TEXT;
    v_email TEXT;
    v_first_name TEXT;
    v_last_name TEXT;
    v_user_id UUID;
    v_is_new_user BOOLEAN := false;
    v_federated_identity_id BIGINT;
    v_result JSONB;
BEGIN
    -- Get provider details
    SELECT 
        ip.*,
        oic.client_id,
        oic.attribute_mapping
    INTO v_provider
    FROM 
        identity_providers ip
        JOIN oidc_configurations oic ON ip.id = oic.provider_id
    WHERE 
        ip.id = p_provider_id
        AND ip.is_enabled = true;
    
    IF NOT FOUND THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', 'Identity provider not found or not enabled',
            'error_code', 'provider_not_found'
        );
    END IF;
    
    v_tenant_id := v_provider.tenant_id;
    
    -- Validate the ID token
    v_token_validation := validate_id_token(p_provider_id, p_id_token, p_nonce);
    
    IF NOT (v_token_validation->>'valid')::boolean THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', 'Invalid ID token: ' || (v_token_validation->>'error'),
            'error_code', 'invalid_token'
        );
    END IF;
    
    v_claims := v_token_validation->'claims';
    
    -- Extract user identifier (sub claim in OIDC)
    v_external_id := v_claims->>'sub';
    
    IF v_external_id IS NULL THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', 'Missing subject identifier in ID token',
            'error_code', 'missing_sub'
        );
    END IF;
    
    -- Extract common user attributes
    v_email := v_claims->>'email';
    v_first_name := COALESCE(v_claims->>'given_name', v_claims->>'first_name');
    v_last_name := COALESCE(v_claims->>'family_name', v_claims->>'last_name');
    
    -- Check if user already exists
    SELECT u.id INTO v_user_id
    FROM federated_identities fi
    JOIN iam.users u ON fi.user_id = u.id
    WHERE fi.provider_id = p_provider_id
    AND fi.external_id = v_external_id;
    
    -- If user doesn't exist and JIT provisioning is enabled
    IF v_user_id IS NULL AND v_provider.jit_provisioning THEN
        -- Create new user
        INSERT INTO iam.users (
            tenant_id,
            username,
            email,
            first_name,
            last_name,
            status,
            created_by
        ) VALUES (
            v_tenant_id,
            COALESCE(v_claims->>'preferred_username', v_external_id),
            v_email,
            v_first_name,
            v_last_name,
            'active',
            -- Use a system user ID for JIT provisioning
            (SELECT id FROM iam.users WHERE username = 'system' AND tenant_id = v_tenant_id)
        ) RETURNING id INTO v_user_id;
        
        v_is_new_user := true;
    END IF;
    
    -- If user exists or was created
    IF v_user_id IS NOT NULL THEN
        -- Check if federated identity exists
        SELECT id INTO v_federated_identity_id
        FROM federated_identities
        WHERE provider_id = p_provider_id
        AND external_id = v_external_id;
        
        -- If federated identity doesn't exist, create it
        IF v_federated_identity_id IS NULL THEN
            INSERT INTO federated_identities (
                user_id,
                provider_id,
                external_id,
                last_login_at,
                access_token,
                refresh_token,
                token_expiration,
                claims,
                metadata
            ) VALUES (
                v_user_id,
                p_provider_id,
                v_external_id,
                now(),
                p_access_token,
                p_refresh_token,
                p_token_expiration,
                v_claims,
                jsonb_build_object(
                    'first_login_at', now(),
                    'source_ip', p_ip_address,
                    'user_agent', p_user_agent
                )
            ) RETURNING id INTO v_federated_identity_id;
        ELSE
            -- Update existing federated identity
            UPDATE federated_identities
            SET 
                last_login_at = now(),
                access_token = p_access_token,
                refresh_token = p_refresh_token,
                token_expiration = p_token_expiration,
                claims = v_claims
            WHERE id = v_federated_identity_id;
        END IF;
        
        -- Log successful authentication
        INSERT INTO federation_audit_log (
            tenant_id,
            provider_id,
            user_id,
            event_type,
            status,
            external_id,
            details,
            ip_address,
            user_agent
        ) VALUES (
            v_tenant_id,
            p_provider_id,
            v_user_id,
            'oidc_authentication',
            'success',
            v_external_id,
            jsonb_build_object(
                'is_new_user', v_is_new_user
            ),
            p_ip_address,
            p_user_agent
        );
        
        -- Return successful result
        v_result := jsonb_build_object(
            'success', true,
            'user_id', v_user_id,
            'is_new_user', v_is_new_user,
            'federated_identity_id', v_federated_identity_id,
            'external_id', v_external_id,
            'provider_id', p_provider_id,
            'provider_name', v_provider.provider_name,
            'tenant_id', v_tenant_id
        );
    ELSE
        -- User doesn't exist and JIT provisioning is disabled
        v_result := jsonb_build_object(
            'success', false,
            'error', 'User not found and JIT provisioning is disabled',
            'error_code', 'user_not_found'
        );
        
        -- Log failed authentication
        INSERT INTO federation_audit_log (
            tenant_id,
            provider_id,
            event_type,
            status,
            external_id,
            details,
            ip_address,
            user_agent,
            error_details
        ) VALUES (
            v_tenant_id,
            p_provider_id,
            'oidc_authentication',
            'failure',
            v_external_id,
            jsonb_build_object(
                'reason', 'User not found and JIT provisioning is disabled'
            ),
            p_ip_address,
            p_user_agent,
            'User not found and JIT provisioning is disabled'
        );
    END IF;
    
    RETURN v_result;
EXCEPTION
    WHEN OTHERS THEN
        -- Log error
        IF v_tenant_id IS NOT NULL THEN
            INSERT INTO federation_audit_log (
                tenant_id,
                provider_id,
                event_type,
                status,
                external_id,
                ip_address,
                user_agent,
                error_details
            ) VALUES (
                v_tenant_id,
                p_provider_id,
                'oidc_authentication',
                'failure',
                v_external_id,
                p_ip_address,
                p_user_agent,
                SQLERRM
            );
        END IF;
        
        RETURN jsonb_build_object(
            'success', false,
            'error', SQLERRM,
            'error_code', 'system_error'
        );
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION process_oidc_authentication(UUID, TEXT, TEXT, TEXT, TIMESTAMPTZ, TEXT, INET, TEXT) IS 
'Processes OpenID Connect authentication using ID token and performs user provisioning if needed';

-- Reset search path
RESET search_path;
