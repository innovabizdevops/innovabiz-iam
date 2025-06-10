-- INNOVABIZ IAM Module - Multi-Tenant RLS (Row-Level Security) Policy Functions
-- Version: 1.0.0
-- Date: 09/05/2025
-- Author: Eduardo Jeremias
-- Description: Functions for implementing advanced RLS policies for multi-tenant isolation

-- Set search path
SET search_path TO iam_multi_tenant, iam, public;

-- Function to apply standard RLS policy to a table
CREATE OR REPLACE FUNCTION apply_tenant_rls_policy(
    p_schema_name TEXT,
    p_table_name TEXT,
    p_tenant_column TEXT DEFAULT 'tenant_id',
    p_policy_name TEXT DEFAULT NULL
) RETURNS BOOLEAN AS $$
DECLARE
    v_policy_name TEXT;
    v_sql TEXT;
BEGIN
    -- Generate policy name if not provided
    IF p_policy_name IS NULL THEN
        v_policy_name := format('rls_tenant_policy_%s', p_table_name);
    ELSE
        v_policy_name := p_policy_name;
    END IF;
    
    -- First, enable RLS on the table
    v_sql := format('ALTER TABLE %I.%I ENABLE ROW LEVEL SECURITY', 
        p_schema_name, p_table_name);
    EXECUTE v_sql;
    
    -- Check if policy already exists and drop it if it does
    IF EXISTS (
        SELECT 1
        FROM pg_catalog.pg_policies
        WHERE schemaname = p_schema_name
        AND tablename = p_table_name
        AND policyname = v_policy_name
    ) THEN
        v_sql := format('DROP POLICY %I ON %I.%I', 
            v_policy_name, p_schema_name, p_table_name);
        EXECUTE v_sql;
    END IF;
    
    -- Create the RLS policy
    v_sql := format(
        'CREATE POLICY %I ON %I.%I 
         USING (%I::UUID = current_setting(''app.current_tenant'')::UUID 
                OR pg_has_role(current_user, ''iam_admin'', ''member''))',
        v_policy_name, p_schema_name, p_table_name, p_tenant_column
    );
    EXECUTE v_sql;
    
    -- Log the policy application
    INSERT INTO tenant_isolation_audit (
        operation_type,
        tenant_id,
        schema_name,
        performed_by,
        details,
        operation_status
    ) VALUES (
        'rls_policy_applied',
        NULL, -- Not specific to a tenant
        p_schema_name,
        current_setting('app.current_user_id', TRUE)::UUID,
        jsonb_build_object(
            'policy_name', v_policy_name,
            'table_name', p_table_name,
            'tenant_column', p_tenant_column
        ),
        'success'
    );
    
    RETURN TRUE;
EXCEPTION
    WHEN OTHERS THEN
        -- Log failure
        INSERT INTO tenant_isolation_audit (
            operation_type,
            tenant_id,
            schema_name,
            performed_by,
            details,
            operation_status,
            error_details
        ) VALUES (
            'rls_policy_applied',
            NULL, -- Not specific to a tenant
            p_schema_name,
            current_setting('app.current_user_id', TRUE)::UUID,
            jsonb_build_object(
                'policy_name', v_policy_name,
                'table_name', p_table_name,
                'tenant_column', p_tenant_column
            ),
            'failure',
            SQLERRM
        );
        
        RAISE EXCEPTION 'Failed to apply RLS policy to %.%: %', p_schema_name, p_table_name, SQLERRM;
        RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION apply_tenant_rls_policy(TEXT, TEXT, TEXT, TEXT) IS 
'Applies a standard tenant isolation RLS policy to a table';

-- Function to apply RLS policies in bulk to a schema
CREATE OR REPLACE FUNCTION apply_rls_policies_to_schema(
    p_schema_name TEXT,
    p_tenant_column TEXT DEFAULT 'tenant_id',
    p_exclude_tables TEXT[] DEFAULT '{}'
) RETURNS TABLE(
    table_name TEXT,
    success BOOLEAN,
    message TEXT
) AS $$
DECLARE
    v_table RECORD;
    v_result BOOLEAN;
    v_error TEXT;
BEGIN
    FOR v_table IN
        SELECT table_name
        FROM information_schema.tables
        WHERE table_schema = p_schema_name
        AND table_type = 'BASE TABLE'
        AND table_name <> ALL(p_exclude_tables)
    LOOP
        -- Check if the table has the tenant column
        IF EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_schema = p_schema_name
            AND table_name = v_table.table_name
            AND column_name = p_tenant_column
        ) THEN
            BEGIN
                v_result := apply_tenant_rls_policy(p_schema_name, v_table.table_name, p_tenant_column);
                table_name := v_table.table_name;
                success := TRUE;
                message := 'RLS policy applied successfully';
                RETURN NEXT;
            EXCEPTION
                WHEN OTHERS THEN
                    v_error := SQLERRM;
                    table_name := v_table.table_name;
                    success := FALSE;
                    message := v_error;
                    RETURN NEXT;
            END;
        ELSE
            -- Table doesn't have tenant column, skip it
            table_name := v_table.table_name;
            success := FALSE;
            message := format('Table does not have tenant column "%s"', p_tenant_column);
            RETURN NEXT;
        END IF;
    END LOOP;
    
    RETURN;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION apply_rls_policies_to_schema(TEXT, TEXT, TEXT[]) IS 
'Applies RLS policies to all tables in a schema that have the specified tenant column';

-- Function to create custom RLS policy from a template
CREATE OR REPLACE FUNCTION create_custom_rls_policy(
    p_template_id BIGINT,
    p_schema_name TEXT,
    p_table_name TEXT,
    p_parameters JSONB DEFAULT NULL
) RETURNS BOOLEAN AS $$
DECLARE
    v_template RECORD;
    v_policy_definition TEXT;
    v_policy_name TEXT;
    v_sql TEXT;
BEGIN
    -- Get template details
    SELECT * INTO v_template
    FROM rls_policy_templates
    WHERE id = p_template_id
    AND is_active = TRUE;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Policy template not found or not active: %', p_template_id;
    END IF;
    
    -- Check if table matches template pattern
    IF NOT (p_table_name ~ v_template.table_pattern) THEN
        RAISE EXCEPTION 'Table % does not match template pattern %', 
            p_table_name, v_template.table_pattern;
    END IF;
    
    -- Generate policy name based on template
    v_policy_name := format('custom_rls_%s_%s', v_template.template_name, p_table_name);
    
    -- Process template definition with parameters if needed
    IF v_template.is_parameterized AND p_parameters IS NOT NULL THEN
        v_policy_definition := v_template.policy_definition;
        
        -- Replace parameters in policy definition
        FOR key, value IN SELECT * FROM jsonb_each_text(p_parameters)
        LOOP
            v_policy_definition := replace(v_policy_definition, 
                '{{' || key || '}}', value);
        END LOOP;
    ELSE
        v_policy_definition := v_template.policy_definition;
    END IF;
    
    -- Enable RLS on table if not already enabled
    v_sql := format('ALTER TABLE %I.%I ENABLE ROW LEVEL SECURITY', 
        p_schema_name, p_table_name);
    EXECUTE v_sql;
    
    -- Drop policy if it already exists
    IF EXISTS (
        SELECT 1
        FROM pg_catalog.pg_policies
        WHERE schemaname = p_schema_name
        AND tablename = p_table_name
        AND policyname = v_policy_name
    ) THEN
        v_sql := format('DROP POLICY %I ON %I.%I', 
            v_policy_name, p_schema_name, p_table_name);
        EXECUTE v_sql;
    END IF;
    
    -- Create the custom policy
    v_sql := format('CREATE POLICY %I ON %I.%I %s %s %s', 
        v_policy_name,
        p_schema_name,
        p_table_name,
        CASE 
            WHEN v_template.policy_type = 'permissive' THEN 'AS PERMISSIVE'
            ELSE 'AS RESTRICTIVE'
        END,
        CASE 
            WHEN v_template.command_type = 'all' THEN 'FOR ALL'
            WHEN v_template.command_type = 'select' THEN 'FOR SELECT'
            WHEN v_template.command_type = 'insert' THEN 'FOR INSERT'
            WHEN v_template.command_type = 'update' THEN 'FOR UPDATE'
            WHEN v_template.command_type = 'delete' THEN 'FOR DELETE'
        END,
        v_policy_definition
    );
    
    EXECUTE v_sql;
    
    -- Log the custom policy creation
    INSERT INTO tenant_isolation_audit (
        operation_type,
        tenant_id,
        schema_name,
        performed_by,
        details,
        operation_status
    ) VALUES (
        'custom_rls_policy_applied',
        NULL, -- Not specific to a tenant
        p_schema_name,
        current_setting('app.current_user_id', TRUE)::UUID,
        jsonb_build_object(
            'policy_name', v_policy_name,
            'table_name', p_table_name,
            'template_id', p_template_id,
            'template_name', v_template.template_name,
            'parameters', p_parameters
        ),
        'success'
    );
    
    RETURN TRUE;
EXCEPTION
    WHEN OTHERS THEN
        -- Log failure
        INSERT INTO tenant_isolation_audit (
            operation_type,
            tenant_id,
            schema_name,
            performed_by,
            details,
            operation_status,
            error_details
        ) VALUES (
            'custom_rls_policy_applied',
            NULL, -- Not specific to a tenant
            p_schema_name,
            current_setting('app.current_user_id', TRUE)::UUID,
            jsonb_build_object(
                'template_id', p_template_id,
                'table_name', p_table_name,
                'parameters', p_parameters
            ),
            'failure',
            SQLERRM
        );
        
        RAISE EXCEPTION 'Failed to apply custom RLS policy to %.%: %', 
            p_schema_name, p_table_name, SQLERRM;
        RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION create_custom_rls_policy(BIGINT, TEXT, TEXT, JSONB) IS 
'Creates a custom RLS policy on a table based on a template';

-- Function to create a hierarchical RLS policy for organizational structures
CREATE OR REPLACE FUNCTION create_hierarchical_rls_policy(
    p_schema_name TEXT,
    p_table_name TEXT,
    p_tenant_column TEXT DEFAULT 'tenant_id',
    p_org_column TEXT DEFAULT 'organization_id',
    p_policy_name TEXT DEFAULT NULL
) RETURNS BOOLEAN AS $$
DECLARE
    v_policy_name TEXT;
    v_sql TEXT;
BEGIN
    -- Generate policy name if not provided
    IF p_policy_name IS NULL THEN
        v_policy_name := format('rls_hierarchical_policy_%s', p_table_name);
    ELSE
        v_policy_name := p_policy_name;
    END IF;
    
    -- First, enable RLS on the table
    v_sql := format('ALTER TABLE %I.%I ENABLE ROW LEVEL SECURITY', 
        p_schema_name, p_table_name);
    EXECUTE v_sql;
    
    -- Check if policy already exists and drop it if it does
    IF EXISTS (
        SELECT 1
        FROM pg_catalog.pg_policies
        WHERE schemaname = p_schema_name
        AND tablename = p_table_name
        AND policyname = v_policy_name
    ) THEN
        v_sql := format('DROP POLICY %I ON %I.%I', 
            v_policy_name, p_schema_name, p_table_name);
        EXECUTE v_sql;
    END IF;
    
    -- Create the hierarchical RLS policy that considers organizational hierarchy
    v_sql := format(
        'CREATE POLICY %I ON %I.%I 
         USING (
            %I::UUID = current_setting(''app.current_tenant'')::UUID
            AND (
                -- Direct organization match
                %I::UUID = current_setting(''app.current_organization'')::UUID
                OR
                -- Organization is in user''s hierarchy (requires iam.organization_hierarchy function)
                EXISTS (
                    SELECT 1 
                    FROM iam.organization_hierarchy(%I::UUID, current_setting(''app.current_organization'')::UUID)
                    WHERE relationship_type IN (''parent'', ''ancestor'')
                )
                OR
                -- User has specific cross-organization permission
                EXISTS (
                    SELECT 1
                    FROM iam.user_permissions up
                    JOIN iam.permissions p ON up.permission_id = p.id
                    WHERE up.user_id = current_setting(''app.current_user_id'')::UUID
                    AND p.name = ''cross_organization_access''
                )
                OR
                -- Admin bypass
                pg_has_role(current_user, ''iam_admin'', ''member'')
            )
         )',
        v_policy_name, p_schema_name, p_table_name, 
        p_tenant_column, p_org_column, p_org_column
    );
    EXECUTE v_sql;
    
    -- Log the policy application
    INSERT INTO tenant_isolation_audit (
        operation_type,
        tenant_id,
        schema_name,
        performed_by,
        details,
        operation_status
    ) VALUES (
        'hierarchical_rls_policy_applied',
        NULL, -- Not specific to a tenant
        p_schema_name,
        current_setting('app.current_user_id', TRUE)::UUID,
        jsonb_build_object(
            'policy_name', v_policy_name,
            'table_name', p_table_name,
            'tenant_column', p_tenant_column,
            'org_column', p_org_column
        ),
        'success'
    );
    
    RETURN TRUE;
EXCEPTION
    WHEN OTHERS THEN
        -- Log failure
        INSERT INTO tenant_isolation_audit (
            operation_type,
            tenant_id,
            schema_name,
            performed_by,
            details,
            operation_status,
            error_details
        ) VALUES (
            'hierarchical_rls_policy_applied',
            NULL, -- Not specific to a tenant
            p_schema_name,
            current_setting('app.current_user_id', TRUE)::UUID,
            jsonb_build_object(
                'policy_name', v_policy_name,
                'table_name', p_table_name,
                'tenant_column', p_tenant_column,
                'org_column', p_org_column
            ),
            'failure',
            SQLERRM
        );
        
        RAISE EXCEPTION 'Failed to apply hierarchical RLS policy to %.%: %', 
            p_schema_name, p_table_name, SQLERRM;
        RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION create_hierarchical_rls_policy(TEXT, TEXT, TEXT, TEXT, TEXT) IS 
'Creates a hierarchical RLS policy that respects organizational structure';

-- Reset search path
RESET search_path;
