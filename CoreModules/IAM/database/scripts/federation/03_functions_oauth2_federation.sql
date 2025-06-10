-- INNOVABIZ IAM Module - OAuth2 Federation Functions
-- Version: 1.0.0
-- Date: 09/05/2025
-- Author: Eduardo Jeremias
-- Description: Core functions for managing OAuth2 identity federation

-- Set search path
SET search_path TO iam_federation, iam, public;

-- Function to create a new OAuth2 identity provider
CREATE OR REPLACE FUNCTION create_oauth2_provider(
    p_tenant_id UUID,
    p_provider_name TEXT,
    p_display_name TEXT,
    p_client_id TEXT,
    p_client_secret TEXT,
    p_authorization_endpoint TEXT,
    p_token_endpoint TEXT,
    p_userinfo_endpoint TEXT,
    p_redirect_uri TEXT,
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
            'oauth2',
            p_display_name,
            p_is_enabled,
            p_is_default,
            p_created_by,
            p_created_by,
            jsonb_build_object(
                'type', 'oauth2',
                'client_id', p_client_id,
                'authorization_endpoint', p_authorization_endpoint,
                'token_endpoint', p_token_endpoint,
                'scope', p_scope
            ),
            p_jit_provisioning
        ) RETURNING id INTO v_provider_id;
        
        -- Insert into oauth2_configurations table
        INSERT INTO oauth2_configurations (
            provider_id,
            client_id,
            client_secret,
            authorization_endpoint,
            token_endpoint,
            userinfo_endpoint,
            scope,
            redirect_uri
        ) VALUES (
            v_provider_id,
            p_client_id,
            p_client_secret,
            p_authorization_endpoint,
            p_token_endpoint,
            p_userinfo_endpoint,
            p_scope,
            p_redirect_uri
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
                'provider_type', 'oauth2',
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
                        'provider_type', 'oauth2',
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

COMMENT ON FUNCTION create_oauth2_provider(UUID, TEXT, TEXT, TEXT, TEXT, TEXT, TEXT, TEXT, TEXT, TEXT, BOOLEAN, BOOLEAN, BOOLEAN, UUID) IS 
'Creates a new OAuth2 identity provider with the specified configuration';

-- Function to generate OAuth2 authorization URL
CREATE OR REPLACE FUNCTION generate_oauth2_authorization_url(
    p_provider_id UUID,
    p_state TEXT,
    p_nonce TEXT DEFAULT NULL,
    p_redirect_uri TEXT DEFAULT NULL,
    p_additional_params JSONB DEFAULT NULL
) RETURNS TEXT AS $$
DECLARE
    v_provider RECORD;
    v_redirect_uri TEXT;
    v_url TEXT;
    v_param_key TEXT;
    v_param_value TEXT;
BEGIN
    -- Get provider details
    SELECT 
        ip.*,
        oc.authorization_endpoint,
        oc.client_id,
        oc.scope,
        oc.redirect_uri AS configured_redirect_uri,
        oc.response_type
    INTO v_provider
    FROM 
        identity_providers ip
        JOIN oauth2_configurations oc ON ip.id = oc.provider_id
    WHERE 
        ip.id = p_provider_id
        AND ip.is_enabled = true;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Identity provider not found or not enabled';
    END IF;
    
    -- Determine which redirect URI to use
    v_redirect_uri := COALESCE(p_redirect_uri, v_provider.configured_redirect_uri);
    
    -- Build the base URL with required parameters
    v_url := v_provider.authorization_endpoint || 
             '?client_id=' || v_provider.client_id ||
             '&response_type=' || COALESCE(v_provider.response_type, 'code') ||
             '&scope=' || v_provider.scope ||
             '&redirect_uri=' || v_redirect_uri ||
             '&state=' || p_state;
    
    -- Add nonce if provided
    IF p_nonce IS NOT NULL THEN
        v_url := v_url || '&nonce=' || p_nonce;
    END IF;
    
    -- Add any additional parameters
    IF p_additional_params IS NOT NULL AND jsonb_typeof(p_additional_params) = 'object' THEN
        FOR v_param_key, v_param_value IN SELECT * FROM jsonb_each_text(p_additional_params) LOOP
            v_url := v_url || '&' || v_param_key || '=' || v_param_value;
        END LOOP;
    END IF;
    
    RETURN v_url;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION generate_oauth2_authorization_url(UUID, TEXT, TEXT, TEXT, JSONB) IS 
'Generates an OAuth2 authorization URL for a given provider';

-- Function to exchange OAuth2 authorization code for tokens
CREATE OR REPLACE FUNCTION process_oauth2_code_exchange(
    p_provider_id UUID,
    p_code TEXT,
    p_redirect_uri TEXT DEFAULT NULL,
    p_state TEXT DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_result JSONB;
BEGIN
    -- This is a placeholder for actual OAuth2 code exchange
    -- In a real implementation, this would make an HTTP request to the token endpoint
    
    -- For now, we just return a mock token response
    v_result := jsonb_build_object(
        'success', true,
        'access_token', 'mock_access_token_' || md5(random()::text),
        'token_type', 'Bearer',
        'expires_in', 3600,
        'refresh_token', 'mock_refresh_token_' || md5(random()::text),
        'id_token', 'mock_id_token_' || md5(random()::text)
    );
    
    RETURN v_result;
EXCEPTION
    WHEN OTHERS THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', SQLERRM
        );
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION process_oauth2_code_exchange(UUID, TEXT, TEXT, TEXT) IS 
'Processes OAuth2 authorization code exchange and returns tokens';

-- Function to process OAuth2 authentication
CREATE OR REPLACE FUNCTION process_oauth2_authentication(
    p_provider_id UUID,
    p_access_token TEXT,
    p_id_token TEXT DEFAULT NULL,
    p_refresh_token TEXT DEFAULT NULL,
    p_token_expiration TIMESTAMPTZ DEFAULT NULL,
    p_user_info JSONB DEFAULT NULL,
    p_ip_address INET DEFAULT NULL,
    p_user_agent TEXT DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_provider RECORD;
    v_tenant_id UUID;
    v_external_id TEXT;
    v_email TEXT;
    v_first_name TEXT;
    v_last_name TEXT;
    v_user_id UUID;
    v_is_new_user BOOLEAN := false;
    v_federated_identity_id BIGINT;
    v_result JSONB;
    v_user_info JSONB;
BEGIN
    -- Get provider details
    SELECT 
        ip.*,
        oc.client_id,
        oc.user_id_attribute,
        oc.attribute_mapping
    INTO v_provider
    FROM 
        identity_providers ip
        JOIN oauth2_configurations oc ON ip.id = oc.provider_id
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
    
    -- Use provided user info or simulate fetching it
    IF p_user_info IS NOT NULL THEN
        v_user_info := p_user_info;
    ELSE
        -- This is a placeholder for actual userinfo fetching
        -- In a real implementation, this would make an HTTP request to the userinfo endpoint
        
        -- For now, we just create mock user info
        v_user_info := jsonb_build_object(
            'sub', 'user456',
            'email', 'user456@example.com',
            'email_verified', true,
            'name', 'Jane Smith',
            'given_name', 'Jane',
            'family_name', 'Smith',
            'preferred_username', 'jsmith',
            'picture', 'https://example.com/profile.jpg'
        );
    END IF;
    
    -- Extract user identifier based on configured attribute
    v_external_id := v_user_info->>COALESCE(v_provider.user_id_attribute, 'sub');
    
    IF v_external_id IS NULL THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', 'Could not extract user identifier from tokens/userinfo',
            'error_code', 'missing_user_id'
        );
    END IF;
    
    -- Extract common user attributes
    v_email := v_user_info->>'email';
    v_first_name := COALESCE(v_user_info->>'given_name', v_user_info->>'first_name');
    v_last_name := COALESCE(v_user_info->>'family_name', v_user_info->>'last_name');
    
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
            COALESCE(v_user_info->>'preferred_username', v_external_id),
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
                v_user_info,
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
                claims = v_user_info
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
            'oauth2_authentication',
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
            'oauth2_authentication',
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
                'oauth2_authentication',
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

COMMENT ON FUNCTION process_oauth2_authentication(UUID, TEXT, TEXT, TEXT, TIMESTAMPTZ, JSONB, INET, TEXT) IS 
'Processes OAuth2 authentication using access token and user info and performs user provisioning if needed';

-- Reset search path
RESET search_path;
