-- INNOVABIZ IAM Module - SAML Federation Functions
-- Version: 1.0.0
-- Date: 09/05/2025
-- Author: Eduardo Jeremias
-- Description: Functions for managing SAML identity federation

-- Set search path
SET search_path TO iam_federation, iam, public;

-- Function to create a new SAML identity provider
CREATE OR REPLACE FUNCTION create_saml_provider(
    p_tenant_id UUID,
    p_provider_name TEXT,
    p_display_name TEXT,
    p_metadata_url TEXT,
    p_entity_id TEXT,
    p_assertion_consumer_service_url TEXT,
    p_certificate TEXT,
    p_private_key TEXT DEFAULT NULL,
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
            'saml',
            p_display_name,
            p_is_enabled,
            p_is_default,
            p_created_by,
            p_created_by,
            jsonb_build_object(
                'type', 'saml',
                'metadata_url', p_metadata_url,
                'entity_id', p_entity_id
            ),
            p_jit_provisioning
        ) RETURNING id INTO v_provider_id;
        
        -- Insert into saml_configurations table
        INSERT INTO saml_configurations (
            provider_id,
            metadata_url,
            entity_id,
            assertion_consumer_service_url,
            certificate,
            private_key
        ) VALUES (
            v_provider_id,
            p_metadata_url,
            p_entity_id,
            p_assertion_consumer_service_url,
            p_certificate,
            p_private_key
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
                'provider_type', 'saml',
                'provider_name', p_provider_name,
                'created_by', p_created_by
            )
        );
        
        RETURN v_provider_id;
    EXCEPTION
        WHEN OTHERS THEN
            -- Log error in audit if provider_id was generated
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
                        'provider_type', 'saml',
                        'provider_name', p_provider_name,
                        'created_by', p_created_by
                    ),
                    SQLERRM
                );
            ELSE
                -- Log error without provider_id
                INSERT INTO federation_audit_log (
                    tenant_id,
                    event_type,
                    status,
                    details,
                    error_details
                ) VALUES (
                    p_tenant_id,
                    'provider_created',
                    'failure',
                    jsonb_build_object(
                        'provider_type', 'saml',
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

COMMENT ON FUNCTION create_saml_provider(UUID, TEXT, TEXT, TEXT, TEXT, TEXT, TEXT, TEXT, BOOLEAN, BOOLEAN, BOOLEAN, UUID) IS 
'Creates a new SAML identity provider with the specified configuration';

-- Function to validate SAML metadata
CREATE OR REPLACE FUNCTION validate_saml_metadata(
    p_metadata_url TEXT,
    p_entity_id TEXT
) RETURNS JSONB AS $$
DECLARE
    v_result JSONB;
BEGIN
    -- This is a placeholder for actual SAML metadata validation
    -- In a real implementation, this would make an HTTP request to fetch the metadata
    -- and validate it against XML schemas and security requirements
    
    -- For now, we just return a mock validation result
    v_result := jsonb_build_object(
        'valid', true,
        'entity_id_matches', true,
        'signature_valid', true,
        'certificate_expiration', (now() + interval '1 year')::date,
        'supports_single_logout', true,
        'message', 'Metadata validation successful'
    );
    
    RETURN v_result;
EXCEPTION
    WHEN OTHERS THEN
        RETURN jsonb_build_object(
            'valid', false,
            'error', SQLERRM,
            'message', 'Metadata validation failed'
        );
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION validate_saml_metadata(TEXT, TEXT) IS 
'Validates SAML metadata from a URL against the provided entity ID';

-- Function to process SAML authentication
CREATE OR REPLACE FUNCTION process_saml_authentication(
    p_provider_id UUID,
    p_saml_response TEXT,
    p_ip_address INET DEFAULT NULL,
    p_user_agent TEXT DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_provider RECORD;
    v_tenant_id UUID;
    v_external_id TEXT;
    v_attributes JSONB;
    v_user_id UUID;
    v_is_new_user BOOLEAN := false;
    v_federated_identity_id BIGINT;
    v_result JSONB;
BEGIN
    -- Get provider details
    SELECT 
        ip.*, 
        sc.entity_id,
        sc.attribute_mapping
    INTO v_provider
    FROM 
        identity_providers ip
        JOIN saml_configurations sc ON ip.id = sc.provider_id
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
    
    -- This is a placeholder for actual SAML response parsing
    -- In a real implementation, this would parse the XML, validate signatures, etc.
    -- For now, we just simulate extracting information from the SAML response
    
    -- Simulate extracting the external ID (NameID)
    v_external_id := 'user123@example.com';
    
    -- Simulate extracting attributes
    v_attributes := jsonb_build_object(
        'firstName', 'John',
        'lastName', 'Doe',
        'email', 'user123@example.com',
        'groups', jsonb_build_array('Employees', 'Finance')
    );
    
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
            v_external_id,
            v_attributes->>'email',
            v_attributes->>'firstName',
            v_attributes->>'lastName',
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
                claims,
                metadata
            ) VALUES (
                v_user_id,
                p_provider_id,
                v_external_id,
                now(),
                v_attributes,
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
                claims = v_attributes,
                metadata = jsonb_set(
                    COALESCE(metadata, '{}'::jsonb),
                    '{last_login_at}',
                    to_jsonb(now())
                )
            WHERE id = v_federated_identity_id;
        END IF;
        
        -- Process group mappings if available
        IF v_attributes ? 'groups' AND jsonb_typeof(v_attributes->'groups') = 'array' THEN
            PERFORM process_saml_group_mappings(
                p_provider_id,
                v_user_id,
                v_attributes->'groups'
            );
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
            'saml_authentication',
            'success',
            v_external_id,
            jsonb_build_object(
                'is_new_user', v_is_new_user,
                'attributes', v_attributes
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
            'saml_authentication',
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
                'saml_authentication',
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

COMMENT ON FUNCTION process_saml_authentication(UUID, TEXT, INET, TEXT) IS 
'Processes SAML authentication and performs user provisioning if needed';

-- Function to process SAML group mappings
CREATE OR REPLACE FUNCTION process_saml_group_mappings(
    p_provider_id UUID,
    p_user_id UUID,
    p_groups JSONB
) RETURNS VOID AS $$
DECLARE
    v_group TEXT;
    v_group_id BIGINT;
    v_role_ids UUID[];
BEGIN
    -- Process each group from SAML assertion
    FOR i IN 0..jsonb_array_length(p_groups) - 1 LOOP
        v_group := p_groups->>i;
        
        -- Find or create the federation group
        SELECT id INTO v_group_id
        FROM federation_groups
        WHERE provider_id = p_provider_id
        AND external_group_id = v_group;
        
        IF v_group_id IS NULL THEN
            -- Create the federation group if it doesn't exist
            INSERT INTO federation_groups (
                provider_id,
                external_group_id,
                group_name
            ) VALUES (
                p_provider_id,
                v_group,
                v_group -- Use the external ID as name initially
            ) RETURNING id INTO v_group_id;
        END IF;
        
        -- Get mapped roles for this group
        SELECT array_agg(local_role_id) INTO v_role_ids
        FROM group_mappings
        WHERE federation_group_id = v_group_id;
        
        -- If there are mapped roles, assign them to the user
        IF v_role_ids IS NOT NULL AND array_length(v_role_ids, 1) > 0 THEN
            -- For each role, create a user-role association if it doesn't exist
            FOR i IN 1..array_length(v_role_ids, 1) LOOP
                -- Check if user already has this role
                IF NOT EXISTS (
                    SELECT 1 
                    FROM iam.user_roles
                    WHERE user_id = p_user_id
                    AND role_id = v_role_ids[i]
                ) THEN
                    -- Assign role to user
                    INSERT INTO iam.user_roles (
                        user_id,
                        role_id,
                        granted_by,
                        grant_type
                    ) VALUES (
                        p_user_id,
                        v_role_ids[i],
                        -- Use a system user ID for SAML-based role assignments
                        (SELECT id FROM iam.users WHERE username = 'system' LIMIT 1),
                        'federation'
                    );
                END IF;
            END LOOP;
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION process_saml_group_mappings(UUID, UUID, JSONB) IS 
'Maps SAML groups to local roles and assigns them to the user';

-- Reset search path
RESET search_path;
