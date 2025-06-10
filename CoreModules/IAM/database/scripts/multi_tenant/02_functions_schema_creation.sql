-- INNOVABIZ IAM Module - Multi-Tenant Schema Creation Functions
-- Version: 1.0.0
-- Date: 09/05/2025
-- Author: Eduardo Jeremias
-- Description: Functions for creating isolated schemas for multi-tenant functionality

-- Set search path
SET search_path TO iam_multi_tenant, iam, public;

-- Function to create tenant schema
CREATE OR REPLACE FUNCTION create_tenant_schema(
    p_tenant_id UUID,
    p_schema_name TEXT,
    p_created_by UUID
) RETURNS BOOLEAN AS $$
DECLARE
    v_job_id BIGINT;
    v_exists BOOLEAN;
BEGIN
    -- Verify if tenant exists in config
    SELECT EXISTS(
        SELECT 1 FROM tenant_isolation_config 
        WHERE tenant_id = p_tenant_id
    ) INTO v_exists;
    
    IF NOT v_exists THEN
        -- Insert new tenant configuration
        INSERT INTO tenant_isolation_config (
            tenant_id,
            schema_name,
            isolation_level,
            created_by,
            updated_by
        ) VALUES (
            p_tenant_id,
            p_schema_name,
            'isolated',
            p_created_by,
            p_created_by
        );
    ELSE
        -- Update existing tenant configuration
        UPDATE tenant_isolation_config
        SET 
            schema_name = p_schema_name,
            isolation_level = 'isolated',
            updated_at = now(),
            updated_by = p_created_by,
            migration_status = 'pending'
        WHERE tenant_id = p_tenant_id;
    END IF;
    
    -- Create a job for schema creation
    INSERT INTO tenant_migration_jobs (
        tenant_id,
        job_type,
        status,
        created_by
    ) VALUES (
        p_tenant_id,
        'create_schema',
        'pending',
        p_created_by
    ) RETURNING id INTO v_job_id;
    
    -- Log the operation in audit
    INSERT INTO tenant_isolation_audit (
        operation_type,
        tenant_id,
        schema_name,
        performed_by,
        details,
        operation_status
    ) VALUES (
        'schema_creation_initiated',
        p_tenant_id,
        p_schema_name,
        p_created_by,
        jsonb_build_object(
            'job_id', v_job_id,
            'isolation_level', 'isolated'
        ),
        'success'
    );
    
    -- Execute the schema creation job immediately
    PERFORM execute_schema_creation_job(v_job_id);
    
    RETURN TRUE;
EXCEPTION
    WHEN OTHERS THEN
        -- Log error in audit
        INSERT INTO tenant_isolation_audit (
            operation_type,
            tenant_id,
            schema_name,
            performed_by,
            details,
            operation_status,
            error_details
        ) VALUES (
            'schema_creation_initiated',
            p_tenant_id,
            p_schema_name,
            p_created_by,
            jsonb_build_object(
                'isolation_level', 'isolated'
            ),
            'failure',
            SQLERRM
        );
        
        RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION create_tenant_schema(UUID, TEXT, UUID) IS 
'Creates an isolated schema for a tenant and sets up the necessary configuration';

-- Function to execute schema creation job
CREATE OR REPLACE FUNCTION execute_schema_creation_job(
    p_job_id BIGINT
) RETURNS BOOLEAN AS $$
DECLARE
    v_tenant_id UUID;
    v_schema_name TEXT;
    v_sql TEXT;
BEGIN
    -- Get job details
    SELECT 
        tmj.tenant_id, 
        tic.schema_name
    INTO 
        v_tenant_id, 
        v_schema_name
    FROM 
        tenant_migration_jobs tmj
        JOIN tenant_isolation_config tic ON tmj.tenant_id = tic.tenant_id
    WHERE 
        tmj.id = p_job_id
        AND tmj.status = 'pending'
        AND tmj.job_type = 'create_schema';
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Job not found or not in pending status: %', p_job_id;
    END IF;
    
    -- Update job status to in_progress
    UPDATE tenant_migration_jobs
    SET 
        status = 'in_progress',
        started_at = now(),
        progress = 0
    WHERE id = p_job_id;
    
    -- Dynamically create schema
    v_sql := format('CREATE SCHEMA IF NOT EXISTS %I', v_schema_name);
    EXECUTE v_sql;
    
    -- Update job progress
    UPDATE tenant_migration_jobs
    SET progress = 50
    WHERE id = p_job_id;
    
    -- Grant usage to necessary roles
    v_sql := format('GRANT USAGE ON SCHEMA %I TO iam_admin', v_schema_name);
    EXECUTE v_sql;
    
    v_sql := format('GRANT USAGE ON SCHEMA %I TO iam_reader', v_schema_name);
    EXECUTE v_sql;
    
    -- Update tenant configuration to reflect schema creation
    UPDATE tenant_isolation_config
    SET migration_status = 'in_progress'
    WHERE tenant_id = v_tenant_id;
    
    -- Mark job as completed
    UPDATE tenant_migration_jobs
    SET 
        status = 'completed',
        completed_at = now(),
        progress = 100,
        execution_log = jsonb_build_object(
            'schema_created', v_schema_name,
            'completed_at', now()
        )
    WHERE id = p_job_id;
    
    -- Log successful schema creation
    INSERT INTO tenant_isolation_audit (
        operation_type,
        tenant_id,
        schema_name,
        performed_by,
        details,
        operation_status
    ) VALUES (
        'schema_created',
        v_tenant_id,
        v_schema_name,
        (SELECT created_by FROM tenant_migration_jobs WHERE id = p_job_id),
        jsonb_build_object(
            'job_id', p_job_id,
            'completed_at', now()
        ),
        'success'
    );
    
    RETURN TRUE;
EXCEPTION
    WHEN OTHERS THEN
        -- Update job to failed status
        UPDATE tenant_migration_jobs
        SET 
            status = 'failed',
            completed_at = now(),
            error_message = SQLERRM,
            execution_log = jsonb_build_object(
                'error', SQLERRM,
                'context', 'Schema creation failed',
                'failed_at', now()
            )
        WHERE id = p_job_id;
        
        -- Log error in audit
        INSERT INTO tenant_isolation_audit (
            operation_type,
            tenant_id,
            schema_name,
            performed_by,
            details,
            operation_status,
            error_details
        ) VALUES (
            'schema_creation',
            v_tenant_id,
            v_schema_name,
            (SELECT created_by FROM tenant_migration_jobs WHERE id = p_job_id),
            jsonb_build_object(
                'job_id', p_job_id,
                'failed_at', now()
            ),
            'failure',
            SQLERRM
        );
        
        RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION execute_schema_creation_job(BIGINT) IS 
'Executes a schema creation job for tenant isolation';

-- Function to check if a schema exists
CREATE OR REPLACE FUNCTION schema_exists(
    p_schema_name TEXT
) RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1
        FROM information_schema.schemata
        WHERE schema_name = p_schema_name
    );
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION schema_exists(TEXT) IS 
'Checks if the specified schema exists in the database';

-- Function to get all tables in a schema
CREATE OR REPLACE FUNCTION get_schema_tables(
    p_schema_name TEXT
) RETURNS TABLE(
    table_name TEXT,
    has_primary_key BOOLEAN,
    estimated_row_count BIGINT,
    has_rls BOOLEAN
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        t.table_name::TEXT,
        EXISTS (
            SELECT 1 
            FROM information_schema.table_constraints tc
            WHERE tc.table_schema = p_schema_name
            AND tc.table_name = t.table_name
            AND tc.constraint_type = 'PRIMARY KEY'
        ) AS has_primary_key,
        pg_table_size(p_schema_name || '.' || t.table_name)::BIGINT AS estimated_row_count,
        EXISTS (
            SELECT 1
            FROM pg_catalog.pg_class c
            JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
            WHERE n.nspname = p_schema_name
            AND c.relname = t.table_name
            AND c.relrowsecurity
        ) AS has_rls
    FROM information_schema.tables t
    WHERE t.table_schema = p_schema_name
    AND t.table_type = 'BASE TABLE';
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_schema_tables(TEXT) IS 
'Returns all tables in the specified schema with metadata about each table';

-- Reset search path
RESET search_path;
