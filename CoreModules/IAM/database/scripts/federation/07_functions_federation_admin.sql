-- INNOVABIZ IAM Module - Federation Administration Functions
-- Version: 1.0.0
-- Date: 09/05/2025
-- Author: Eduardo Jeremias
-- Description: Core administration functions for identity federation management

-- Set search path
SET search_path TO iam_federation, iam, public;

-- Function to list all identity providers for a tenant
CREATE OR REPLACE FUNCTION list_identity_providers(
    p_tenant_id UUID
) RETURNS TABLE (
    id UUID,
    provider_name TEXT,
    provider_type TEXT,
    display_name TEXT,
    is_enabled BOOLEAN,
    is_default BOOLEAN,
    jit_provisioning BOOLEAN,
    status TEXT,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    last_sync_date TIMESTAMPTZ,
    user_count BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        ip.id,
        ip.provider_name,
        ip.provider_type,
        ip.display_name,
        ip.is_enabled,
        ip.is_default,
        ip.jit_provisioning,
        ip.status,
        ip.created_at,
        ip.updated_at,
        ip.last_sync_date,
        COUNT(DISTINCT fi.user_id) AS user_count
    FROM
        identity_providers ip
    LEFT JOIN
        federated_identities fi ON ip.id = fi.provider_id
    WHERE
        ip.tenant_id = p_tenant_id
    GROUP BY
        ip.id
    ORDER BY
        ip.is_default DESC,
        ip.display_name ASC;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION list_identity_providers(UUID) IS 
'Lists all identity providers for a tenant with user counts';

-- Function to toggle identity provider status
CREATE OR REPLACE FUNCTION toggle_identity_provider_status(
    p_provider_id UUID,
    p_is_enabled BOOLEAN,
    p_updated_by UUID
) RETURNS BOOLEAN AS $$
DECLARE
    v_tenant_id UUID;
BEGIN
    -- Get tenant ID and update provider
    UPDATE identity_providers
    SET 
        is_enabled = p_is_enabled,
        updated_at = now(),
        updated_by = p_updated_by
    WHERE id = p_provider_id
    RETURNING tenant_id INTO v_tenant_id;
    
    IF NOT FOUND THEN
        RETURN false;
    END IF;
    
    -- Log the action
    INSERT INTO federation_audit_log (
        tenant_id,
        provider_id,
        event_type,
        status,
        details
    ) VALUES (
        v_tenant_id,
        p_provider_id,
        'provider_status_changed',
        'success',
        jsonb_build_object(
            'is_enabled', p_is_enabled,
            'updated_by', p_updated_by
        )
    );
    
    RETURN true;
EXCEPTION
    WHEN OTHERS THEN
        RETURN false;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION toggle_identity_provider_status(UUID, BOOLEAN, UUID) IS 
'Enables or disables an identity provider';

-- Function to delete an identity provider
CREATE OR REPLACE FUNCTION delete_identity_provider(
    p_provider_id UUID
) RETURNS BOOLEAN AS $$
DECLARE
    v_tenant_id UUID;
    v_provider_type TEXT;
    v_provider_name TEXT;
BEGIN
    -- Get provider details before deletion
    SELECT tenant_id, provider_type, provider_name 
    INTO v_tenant_id, v_provider_type, v_provider_name
    FROM identity_providers
    WHERE id = p_provider_id;
    
    IF NOT FOUND THEN
        RETURN false;
    END IF;
    
    -- Delete the provider (cascade will handle related records)
    DELETE FROM identity_providers
    WHERE id = p_provider_id;
    
    -- Log the deletion
    INSERT INTO federation_audit_log (
        tenant_id,
        event_type,
        status,
        details
    ) VALUES (
        v_tenant_id,
        'provider_deleted',
        'success',
        jsonb_build_object(
            'provider_id', p_provider_id,
            'provider_type', v_provider_type,
            'provider_name', v_provider_name
        )
    );
    
    RETURN true;
EXCEPTION
    WHEN OTHERS THEN
        -- Log the error
        IF v_tenant_id IS NOT NULL THEN
            INSERT INTO federation_audit_log (
                tenant_id,
                event_type,
                status,
                details,
                error_details
            ) VALUES (
                v_tenant_id,
                'provider_deleted',
                'failure',
                jsonb_build_object(
                    'provider_id', p_provider_id
                ),
                SQLERRM
            );
        END IF;
        
        RETURN false;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION delete_identity_provider(UUID) IS 
'Deletes an identity provider and all associated records';

-- Function to list federated users
CREATE OR REPLACE FUNCTION list_federated_users(
    p_tenant_id UUID,
    p_provider_id UUID DEFAULT NULL,
    p_limit INTEGER DEFAULT 100,
    p_offset INTEGER DEFAULT 0
) RETURNS TABLE (
    user_id UUID,
    username TEXT,
    email TEXT,
    first_name TEXT,
    last_name TEXT,
    provider_id UUID,
    provider_name TEXT,
    provider_type TEXT,
    external_id TEXT,
    last_login_at TIMESTAMPTZ,
    status TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        u.id AS user_id,
        u.username,
        u.email,
        u.first_name,
        u.last_name,
        ip.id AS provider_id,
        ip.provider_name,
        ip.provider_type,
        fi.external_id,
        fi.last_login_at,
        fi.status
    FROM
        iam.users u
    JOIN
        federated_identities fi ON u.id = fi.user_id
    JOIN
        identity_providers ip ON fi.provider_id = ip.id
    WHERE
        u.tenant_id = p_tenant_id
        AND (p_provider_id IS NULL OR fi.provider_id = p_provider_id)
    ORDER BY
        fi.last_login_at DESC NULLS LAST
    LIMIT p_limit
    OFFSET p_offset;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION list_federated_users(UUID, UUID, INTEGER, INTEGER) IS 
'Lists all federated users for a tenant, optionally filtered by provider';

-- Function to link an existing user to a federated identity
CREATE OR REPLACE FUNCTION link_user_to_identity(
    p_user_id UUID,
    p_provider_id UUID,
    p_external_id TEXT,
    p_claims JSONB DEFAULT NULL
) RETURNS BOOLEAN AS $$
DECLARE
    v_tenant_id UUID;
    v_provider_tenant_id UUID;
BEGIN
    -- Get user tenant ID
    SELECT tenant_id INTO v_tenant_id
    FROM iam.users
    WHERE id = p_user_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'User not found';
    END IF;
    
    -- Check if provider belongs to the same tenant
    SELECT tenant_id INTO v_provider_tenant_id
    FROM identity_providers
    WHERE id = p_provider_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Identity provider not found';
    END IF;
    
    IF v_tenant_id != v_provider_tenant_id THEN
        RAISE EXCEPTION 'User and identity provider must belong to the same tenant';
    END IF;
    
    -- Check if the identity already exists
    IF EXISTS (
        SELECT 1
        FROM federated_identities
        WHERE provider_id = p_provider_id
        AND external_id = p_external_id
    ) THEN
        RAISE EXCEPTION 'Federated identity already exists';
    END IF;
    
    -- Create the federated identity link
    INSERT INTO federated_identities (
        user_id,
        provider_id,
        external_id,
        claims,
        metadata
    ) VALUES (
        p_user_id,
        p_provider_id,
        p_external_id,
        p_claims,
        jsonb_build_object(
            'manually_linked', true,
            'linked_at', now()
        )
    );
    
    -- Log the link creation
    INSERT INTO federation_audit_log (
        tenant_id,
        provider_id,
        user_id,
        event_type,
        status,
        external_id,
        details
    ) VALUES (
        v_tenant_id,
        p_provider_id,
        p_user_id,
        'identity_linked',
        'success',
        p_external_id,
        jsonb_build_object(
            'manually_linked', true
        )
    );
    
    RETURN true;
EXCEPTION
    WHEN OTHERS THEN
        -- Log the error
        IF v_tenant_id IS NOT NULL THEN
            INSERT INTO federation_audit_log (
                tenant_id,
                provider_id,
                user_id,
                event_type,
                status,
                external_id,
                details,
                error_details
            ) VALUES (
                v_tenant_id,
                p_provider_id,
                p_user_id,
                'identity_linked',
                'failure',
                p_external_id,
                jsonb_build_object(
                    'manually_linked', true
                ),
                SQLERRM
            );
        END IF;
        
        RAISE;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION link_user_to_identity(UUID, UUID, TEXT, JSONB) IS 
'Links an existing user to a federated identity';

-- Function to unlink a federated identity from a user
CREATE OR REPLACE FUNCTION unlink_federated_identity(
    p_user_id UUID,
    p_identity_id BIGINT
) RETURNS BOOLEAN AS $$
DECLARE
    v_tenant_id UUID;
    v_provider_id UUID;
    v_external_id TEXT;
BEGIN
    -- Get tenant ID and identity details
    SELECT 
        u.tenant_id, 
        fi.provider_id, 
        fi.external_id 
    INTO 
        v_tenant_id, 
        v_provider_id, 
        v_external_id
    FROM 
        federated_identities fi
        JOIN iam.users u ON fi.user_id = u.id
    WHERE 
        fi.id = p_identity_id
        AND fi.user_id = p_user_id;
    
    IF NOT FOUND THEN
        RETURN false;
    END IF;
    
    -- Delete the federated identity
    DELETE FROM federated_identities
    WHERE id = p_identity_id
    AND user_id = p_user_id;
    
    -- Log the unlink
    INSERT INTO federation_audit_log (
        tenant_id,
        provider_id,
        user_id,
        event_type,
        status,
        external_id,
        details
    ) VALUES (
        v_tenant_id,
        v_provider_id,
        p_user_id,
        'identity_unlinked',
        'success',
        v_external_id,
        jsonb_build_object(
            'identity_id', p_identity_id
        )
    );
    
    RETURN true;
EXCEPTION
    WHEN OTHERS THEN
        -- Log the error
        IF v_tenant_id IS NOT NULL THEN
            INSERT INTO federation_audit_log (
                tenant_id,
                provider_id,
                user_id,
                event_type,
                status,
                external_id,
                details,
                error_details
            ) VALUES (
                v_tenant_id,
                v_provider_id,
                p_user_id,
                'identity_unlinked',
                'failure',
                v_external_id,
                jsonb_build_object(
                    'identity_id', p_identity_id
                ),
                SQLERRM
            );
        END IF;
        
        RETURN false;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION unlink_federated_identity(UUID, BIGINT) IS 
'Unlinks a federated identity from a user';

-- Function to configure JIT provisioning settings
CREATE OR REPLACE FUNCTION configure_jit_provisioning(
    p_provider_id UUID,
    p_jit_provisioning BOOLEAN,
    p_auto_link_accounts BOOLEAN,
    p_updated_by UUID
) RETURNS BOOLEAN AS $$
DECLARE
    v_tenant_id UUID;
BEGIN
    -- Update provider settings
    UPDATE identity_providers
    SET 
        jit_provisioning = p_jit_provisioning,
        auto_link_accounts = p_auto_link_accounts,
        updated_at = now(),
        updated_by = p_updated_by
    WHERE id = p_provider_id
    RETURNING tenant_id INTO v_tenant_id;
    
    IF NOT FOUND THEN
        RETURN false;
    END IF;
    
    -- Log the configuration change
    INSERT INTO federation_audit_log (
        tenant_id,
        provider_id,
        event_type,
        status,
        details
    ) VALUES (
        v_tenant_id,
        p_provider_id,
        'jit_provisioning_configured',
        'success',
        jsonb_build_object(
            'jit_provisioning', p_jit_provisioning,
            'auto_link_accounts', p_auto_link_accounts,
            'updated_by', p_updated_by
        )
    );
    
    RETURN true;
EXCEPTION
    WHEN OTHERS THEN
        RETURN false;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION configure_jit_provisioning(UUID, BOOLEAN, BOOLEAN, UUID) IS 
'Configures JIT provisioning settings for an identity provider';

-- Function to get federation audit log
CREATE OR REPLACE FUNCTION get_federation_audit_log(
    p_tenant_id UUID,
    p_provider_id UUID DEFAULT NULL,
    p_user_id UUID DEFAULT NULL,
    p_event_type TEXT DEFAULT NULL,
    p_status TEXT DEFAULT NULL,
    p_start_date TIMESTAMPTZ DEFAULT NULL,
    p_end_date TIMESTAMPTZ DEFAULT NULL,
    p_limit INTEGER DEFAULT 100,
    p_offset INTEGER DEFAULT 0
) RETURNS TABLE (
    id BIGINT,
    provider_id UUID,
    provider_name TEXT,
    user_id UUID,
    username TEXT,
    event_type TEXT,
    event_time TIMESTAMPTZ,
    status TEXT,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    error_details TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        fal.id,
        fal.provider_id,
        ip.provider_name,
        fal.user_id,
        u.username,
        fal.event_type,
        fal.event_time,
        fal.status,
        fal.details,
        fal.ip_address,
        fal.user_agent,
        fal.error_details
    FROM
        federation_audit_log fal
    LEFT JOIN
        identity_providers ip ON fal.provider_id = ip.id
    LEFT JOIN
        iam.users u ON fal.user_id = u.id
    WHERE
        fal.tenant_id = p_tenant_id
        AND (p_provider_id IS NULL OR fal.provider_id = p_provider_id)
        AND (p_user_id IS NULL OR fal.user_id = p_user_id)
        AND (p_event_type IS NULL OR fal.event_type = p_event_type)
        AND (p_status IS NULL OR fal.status = p_status)
        AND (p_start_date IS NULL OR fal.event_time >= p_start_date)
        AND (p_end_date IS NULL OR fal.event_time <= p_end_date)
    ORDER BY
        fal.event_time DESC
    LIMIT p_limit
    OFFSET p_offset;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_federation_audit_log(UUID, UUID, UUID, TEXT, TEXT, TIMESTAMPTZ, TIMESTAMPTZ, INTEGER, INTEGER) IS 
'Gets federation audit log filtered by various criteria';

-- Function to create group mappings between external groups and local roles
CREATE OR REPLACE FUNCTION create_group_mapping(
    p_federation_group_id BIGINT,
    p_local_role_id UUID,
    p_mapping_type TEXT DEFAULT 'exact',
    p_priority INTEGER DEFAULT 100,
    p_created_by UUID
) RETURNS BOOLEAN AS $$
DECLARE
    v_provider_id UUID;
    v_tenant_id UUID;
BEGIN
    -- Get provider details
    SELECT 
        fg.provider_id, 
        ip.tenant_id
    INTO 
        v_provider_id, 
        v_tenant_id
    FROM 
        federation_groups fg
        JOIN identity_providers ip ON fg.provider_id = ip.id
    WHERE 
        fg.id = p_federation_group_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Federation group not found';
    END IF;
    
    -- Check if role belongs to the same tenant
    IF NOT EXISTS (
        SELECT 1
        FROM iam.roles
        WHERE id = p_local_role_id
        AND tenant_id = v_tenant_id
    ) THEN
        RAISE EXCEPTION 'Role does not belong to the same tenant as the identity provider';
    END IF;
    
    -- Create the mapping
    INSERT INTO group_mappings (
        federation_group_id,
        local_role_id,
        created_by,
        mapping_type,
        priority
    ) VALUES (
        p_federation_group_id,
        p_local_role_id,
        p_created_by,
        p_mapping_type,
        p_priority
    );
    
    RETURN true;
EXCEPTION
    WHEN OTHERS THEN
        RAISE;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION create_group_mapping(BIGINT, UUID, TEXT, INTEGER, UUID) IS 
'Creates a mapping between a federation group and a local role';

-- Reset search path
RESET search_path;
