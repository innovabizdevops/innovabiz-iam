-- INNOVABIZ IAM Module - FIDO2/WebAuthn Functions
-- Version: 1.0.0
-- Date: 09/05/2025
-- Author: Eduardo Jeremias
-- Description: Core functions for managing FIDO2/WebAuthn passwordless authentication

-- Set search path
SET search_path TO iam_federation, iam, public;

-- Function to create a new FIDO2 configuration for a tenant
CREATE OR REPLACE FUNCTION create_fido2_configuration(
    p_tenant_id UUID,
    p_rp_id TEXT,
    p_rp_name TEXT,
    p_rp_icon TEXT DEFAULT NULL,
    p_attestation_preference TEXT DEFAULT 'direct',
    p_user_verification TEXT DEFAULT 'preferred',
    p_timeout INTEGER DEFAULT 60000,
    p_created_by UUID
) RETURNS UUID AS $$
DECLARE
    v_config_id UUID;
BEGIN
    -- Check if configuration already exists for tenant
    SELECT id INTO v_config_id
    FROM fido2_configurations
    WHERE tenant_id = p_tenant_id;
    
    IF v_config_id IS NOT NULL THEN
        -- Update existing configuration
        UPDATE fido2_configurations
        SET 
            rp_id = p_rp_id,
            rp_name = p_rp_name,
            rp_icon = p_rp_icon,
            attestation_preference = p_attestation_preference,
            user_verification = p_user_verification,
            timeout = p_timeout,
            updated_at = now(),
            updated_by = p_created_by
        WHERE id = v_config_id;
    ELSE
        -- Create new configuration
        INSERT INTO fido2_configurations (
            tenant_id,
            rp_id,
            rp_name,
            rp_icon,
            attestation_preference,
            user_verification,
            timeout,
            created_by,
            updated_by
        ) VALUES (
            p_tenant_id,
            p_rp_id,
            p_rp_name,
            p_rp_icon,
            p_attestation_preference,
            p_user_verification,
            p_timeout,
            p_created_by,
            p_created_by
        ) RETURNING id INTO v_config_id;
    END IF;
    
    RETURN v_config_id;
EXCEPTION
    WHEN OTHERS THEN
        RAISE;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION create_fido2_configuration(UUID, TEXT, TEXT, TEXT, TEXT, TEXT, INTEGER, UUID) IS 
'Creates or updates a FIDO2/WebAuthn configuration for a tenant';

-- Function to generate a registration challenge
CREATE OR REPLACE FUNCTION generate_registration_challenge(
    p_tenant_id UUID,
    p_user_id UUID,
    p_credential_name TEXT DEFAULT NULL,
    p_authenticator_attachment TEXT DEFAULT NULL,
    p_require_resident_key BOOLEAN DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_config RECORD;
    v_user RECORD;
    v_challenge TEXT;
    v_existing_credentials JSONB;
    v_result JSONB;
BEGIN
    -- Get FIDO2 configuration for tenant
    SELECT * INTO v_config
    FROM fido2_configurations
    WHERE tenant_id = p_tenant_id
    AND is_enabled = true;
    
    IF NOT FOUND THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', 'FIDO2 configuration not found for tenant',
            'error_code', 'missing_configuration'
        );
    END IF;
    
    -- Get user details
    SELECT id, username, email, first_name, last_name INTO v_user
    FROM iam.users
    WHERE id = p_user_id
    AND tenant_id = p_tenant_id
    AND status = 'active';
    
    IF NOT FOUND THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', 'User not found or not active',
            'error_code', 'invalid_user'
        );
    END IF;
    
    -- Generate random challenge
    v_challenge := encode(gen_random_bytes(v_config.challenge_size), 'base64');
    
    -- Get existing credentials for user to exclude them
    SELECT jsonb_agg(jsonb_build_object(
        'type', 'public-key',
        'id', credential_id
    )) INTO v_existing_credentials
    FROM fido2_credentials
    WHERE user_id = p_user_id
    AND is_active = true;
    
    -- If no existing credentials, initialize to empty array
    IF v_existing_credentials IS NULL THEN
        v_existing_credentials := '[]'::jsonb;
    END IF;
    
    -- Build registration options
    v_result := jsonb_build_object(
        'success', true,
        'options', jsonb_build_object(
            'rp', jsonb_build_object(
                'id', v_config.rp_id,
                'name', v_config.rp_name,
                'icon', v_config.rp_icon
            ),
            'user', jsonb_build_object(
                'id', encode(v_user.id::text::bytea, 'base64'),
                'name', v_user.username,
                'displayName', COALESCE(v_user.first_name || ' ' || v_user.last_name, v_user.username),
                'icon', NULL
            ),
            'challenge', v_challenge,
            'pubKeyCredParams', jsonb_build_array(
                jsonb_build_object('type', 'public-key', 'alg', -7), -- ES256
                jsonb_build_object('type', 'public-key', 'alg', -257) -- RS256
            ),
            'timeout', v_config.timeout,
            'excludeCredentials', v_existing_credentials,
            'attestation', v_config.attestation_preference,
            'userVerification', v_config.user_verification
        ),
        'credential_name', p_credential_name
    );
    
    -- Add authenticator attachment if specified
    IF p_authenticator_attachment IS NOT NULL THEN
        v_result := jsonb_set(v_result, '{options, authenticatorSelection, authenticatorAttachment}', to_jsonb(p_authenticator_attachment));
    END IF;
    
    -- Add resident key requirement if specified
    IF p_require_resident_key IS NOT NULL THEN
        v_result := jsonb_set(v_result, '{options, authenticatorSelection, requireResidentKey}', to_jsonb(p_require_resident_key));
    END IF;
    
    -- Store challenge in session or cache (This would be implemented externally)
    -- ...
    
    RETURN v_result;
EXCEPTION
    WHEN OTHERS THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', SQLERRM,
            'error_code', 'system_error'
        );
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION generate_registration_challenge(UUID, UUID, TEXT, TEXT, BOOLEAN) IS 
'Generates a registration challenge for FIDO2/WebAuthn credential enrollment';

-- Function to verify and register a credential
CREATE OR REPLACE FUNCTION verify_and_register_credential(
    p_tenant_id UUID,
    p_user_id UUID,
    p_credential_id TEXT,
    p_public_key TEXT,
    p_attestation_type TEXT,
    p_attestation_format TEXT,
    p_aaguid TEXT,
    p_credential_name TEXT DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_credential_id BIGINT;
BEGIN
    -- In a real implementation, this would verify the attestation response
    -- against the stored challenge and attestation requirements
    
    -- For now, we just register the credential
    INSERT INTO fido2_credentials (
        user_id,
        credential_id,
        public_key,
        attestation_type,
        attestation_format,
        aaguid,
        credential_name
    ) VALUES (
        p_user_id,
        p_credential_id,
        p_public_key,
        p_attestation_type,
        p_attestation_format,
        p_aaguid,
        COALESCE(p_credential_name, 'Security Key')
    ) RETURNING id INTO v_credential_id;
    
    -- Log successful registration
    INSERT INTO federation_audit_log (
        tenant_id,
        user_id,
        event_type,
        status,
        details
    ) VALUES (
        p_tenant_id,
        p_user_id,
        'fido2_credential_registered',
        'success',
        jsonb_build_object(
            'credential_id', p_credential_id,
            'attestation_type', p_attestation_type,
            'attestation_format', p_attestation_format,
            'aaguid', p_aaguid
        )
    );
    
    RETURN jsonb_build_object(
        'success', true,
        'credential_id', v_credential_id,
        'message', 'Credential successfully registered'
    );
EXCEPTION
    WHEN OTHERS THEN
        -- Log error
        INSERT INTO federation_audit_log (
            tenant_id,
            user_id,
            event_type,
            status,
            details,
            error_details
        ) VALUES (
            p_tenant_id,
            p_user_id,
            'fido2_credential_registered',
            'failure',
            jsonb_build_object(
                'credential_id', p_credential_id
            ),
            SQLERRM
        );
        
        RETURN jsonb_build_object(
            'success', false,
            'error', SQLERRM,
            'error_code', 'registration_failed'
        );
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION verify_and_register_credential(UUID, UUID, TEXT, TEXT, TEXT, TEXT, TEXT, TEXT) IS 
'Verifies an attestation response and registers a new FIDO2/WebAuthn credential';

-- Function to generate authentication challenge
CREATE OR REPLACE FUNCTION generate_authentication_challenge(
    p_tenant_id UUID,
    p_username TEXT DEFAULT NULL,
    p_user_id UUID DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_config RECORD;
    v_user_id UUID;
    v_challenge TEXT;
    v_allowed_credentials JSONB;
    v_result JSONB;
BEGIN
    -- Get FIDO2 configuration for tenant
    SELECT * INTO v_config
    FROM fido2_configurations
    WHERE tenant_id = p_tenant_id
    AND is_enabled = true;
    
    IF NOT FOUND THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', 'FIDO2 configuration not found for tenant',
            'error_code', 'missing_configuration'
        );
    END IF;
    
    -- If user_id is provided, use it directly
    IF p_user_id IS NOT NULL THEN
        v_user_id := p_user_id;
    -- Otherwise, look up user by username
    ELSIF p_username IS NOT NULL THEN
        SELECT id INTO v_user_id
        FROM iam.users
        WHERE username = p_username
        AND tenant_id = p_tenant_id
        AND status = 'active';
        
        IF NOT FOUND THEN
            RETURN jsonb_build_object(
                'success', false,
                'error', 'User not found or not active',
                'error_code', 'invalid_user'
            );
        END IF;
    END IF;
    
    -- Generate random challenge
    v_challenge := encode(gen_random_bytes(v_config.challenge_size), 'base64');
    
    -- If user is known, get their credentials
    IF v_user_id IS NOT NULL THEN
        SELECT jsonb_agg(jsonb_build_object(
            'type', 'public-key',
            'id', credential_id
        )) INTO v_allowed_credentials
        FROM fido2_credentials
        WHERE user_id = v_user_id
        AND is_active = true;
        
        -- If no credentials found, return error
        IF v_allowed_credentials IS NULL OR jsonb_array_length(v_allowed_credentials) = 0 THEN
            RETURN jsonb_build_object(
                'success', false,
                'error', 'No credentials found for user',
                'error_code', 'no_credentials'
            );
        END IF;
    END IF;
    
    -- Build authentication options
    v_result := jsonb_build_object(
        'success', true,
        'options', jsonb_build_object(
            'challenge', v_challenge,
            'timeout', v_config.timeout,
            'rpId', v_config.rp_id,
            'userVerification', v_config.user_verification
        )
    );
    
    -- Add allowed credentials if user is known
    IF v_allowed_credentials IS NOT NULL THEN
        v_result := jsonb_set(v_result, '{options, allowCredentials}', v_allowed_credentials);
    END IF;
    
    -- Store challenge in session or cache (This would be implemented externally)
    -- ...
    
    RETURN v_result;
EXCEPTION
    WHEN OTHERS THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', SQLERRM,
            'error_code', 'system_error'
        );
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION generate_authentication_challenge(UUID, TEXT, UUID) IS 
'Generates an authentication challenge for FIDO2/WebAuthn passwordless login';

-- Function to verify authentication assertion
CREATE OR REPLACE FUNCTION verify_authentication_assertion(
    p_tenant_id UUID,
    p_credential_id TEXT,
    p_signature TEXT,
    p_authenticator_data TEXT,
    p_client_data_json TEXT,
    p_user_handle TEXT DEFAULT NULL,
    p_ip_address INET DEFAULT NULL,
    p_user_agent TEXT DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_credential RECORD;
    v_user_id UUID;
    v_user RECORD;
    v_counter BIGINT;
BEGIN
    -- Get credential details
    SELECT fc.*, u.id AS user_id, u.username, u.email
    INTO v_credential
    FROM fido2_credentials fc
    JOIN iam.users u ON fc.user_id = u.id
    WHERE fc.credential_id = p_credential_id
    AND fc.is_active = true
    AND u.tenant_id = p_tenant_id;
    
    IF NOT FOUND THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', 'Credential not found or not active',
            'error_code', 'invalid_credential'
        );
    END IF;
    
    -- In a real implementation, this would verify the assertion
    -- against the stored challenge, public key, etc.
    
    -- Update the credential's counter and last used timestamp
    -- The counter value would come from the authenticator data
    v_counter := v_credential.counter + 1;
    
    UPDATE fido2_credentials
    SET 
        last_used_at = now(),
        counter = v_counter
    WHERE id = v_credential.id;
    
    -- Log successful authentication
    INSERT INTO federation_audit_log (
        tenant_id,
        user_id,
        event_type,
        status,
        details,
        ip_address,
        user_agent
    ) VALUES (
        p_tenant_id,
        v_credential.user_id,
        'fido2_authentication',
        'success',
        jsonb_build_object(
            'credential_id', p_credential_id,
            'counter', v_counter
        ),
        p_ip_address,
        p_user_agent
    );
    
    -- Return success with user details
    RETURN jsonb_build_object(
        'success', true,
        'user_id', v_credential.user_id,
        'username', v_credential.username,
        'email', v_credential.email,
        'credential_name', v_credential.credential_name,
        'message', 'Authentication successful'
    );
EXCEPTION
    WHEN OTHERS THEN
        -- Log error
        INSERT INTO federation_audit_log (
            tenant_id,
            event_type,
            status,
            details,
            ip_address,
            user_agent,
            error_details
        ) VALUES (
            p_tenant_id,
            'fido2_authentication',
            'failure',
            jsonb_build_object(
                'credential_id', p_credential_id
            ),
            p_ip_address,
            p_user_agent,
            SQLERRM
        );
        
        RETURN jsonb_build_object(
            'success', false,
            'error', SQLERRM,
            'error_code', 'verification_failed'
        );
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION verify_authentication_assertion(UUID, TEXT, TEXT, TEXT, TEXT, TEXT, INET, TEXT) IS 
'Verifies an authentication assertion for FIDO2/WebAuthn passwordless login';

-- Function to list user's credentials
CREATE OR REPLACE FUNCTION list_user_credentials(
    p_user_id UUID
) RETURNS JSONB AS $$
DECLARE
    v_credentials JSONB;
BEGIN
    SELECT jsonb_agg(jsonb_build_object(
        'id', id,
        'credential_id', credential_id,
        'credential_name', COALESCE(credential_name, 'Security Key'),
        'created_at', created_at,
        'last_used_at', last_used_at,
        'device_type', device_type,
        'is_active', is_active
    ))
    INTO v_credentials
    FROM fido2_credentials
    WHERE user_id = p_user_id;
    
    IF v_credentials IS NULL THEN
        v_credentials := '[]'::jsonb;
    END IF;
    
    RETURN jsonb_build_object(
        'success', true,
        'credentials', v_credentials
    );
EXCEPTION
    WHEN OTHERS THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', SQLERRM
        );
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION list_user_credentials(UUID) IS 
'Lists all FIDO2/WebAuthn credentials registered for a user';

-- Function to deactivate a credential
CREATE OR REPLACE FUNCTION deactivate_credential(
    p_user_id UUID,
    p_credential_id BIGINT
) RETURNS BOOLEAN AS $$
BEGIN
    UPDATE fido2_credentials
    SET is_active = false
    WHERE id = p_credential_id
    AND user_id = p_user_id;
    
    IF FOUND THEN
        RETURN true;
    ELSE
        RETURN false;
    END IF;
EXCEPTION
    WHEN OTHERS THEN
        RAISE;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION deactivate_credential(UUID, BIGINT) IS 
'Deactivates a FIDO2/WebAuthn credential';

-- Reset search path
RESET search_path;
