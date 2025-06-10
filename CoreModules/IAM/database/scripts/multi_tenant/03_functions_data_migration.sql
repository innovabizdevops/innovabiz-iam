-- INNOVABIZ IAM Module - Multi-Tenant Data Migration Functions
-- Version: 1.0.0
-- Date: 09/05/2025
-- Author: Eduardo Jeremias
-- Description: Functions for migrating data between schemas for tenant isolation

-- Set search path
SET search_path TO iam_multi_tenant, iam, public;

-- Function to initiate data migration for a tenant
CREATE OR REPLACE FUNCTION migrate_tenant_data(
    p_tenant_id UUID,
    p_created_by UUID
) RETURNS BOOLEAN AS $$
DECLARE
    v_schema_name TEXT;
    v_job_id BIGINT;
    v_migration_status TEXT;
BEGIN
    -- Get tenant details
    SELECT 
        schema_name,
        migration_status
    INTO 
        v_schema_name,
        v_migration_status
    FROM 
        tenant_isolation_config
    WHERE 
        tenant_id = p_tenant_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Tenant not found: %', p_tenant_id;
    END IF;
    
    -- Check if schema exists
    IF NOT schema_exists(v_schema_name) THEN
        RAISE EXCEPTION 'Target schema does not exist: %', v_schema_name;
    END IF;
    
    -- Check if migration is already in progress
    IF v_migration_status = 'in_progress' THEN
        RAISE EXCEPTION 'Migration already in progress for tenant: %', p_tenant_id;
    END IF;
    
    -- Create a job for data migration
    INSERT INTO tenant_migration_jobs (
        tenant_id,
        job_type,
        status,
        created_by
    ) VALUES (
        p_tenant_id,
        'migrate_data',
        'pending',
        p_created_by
    ) RETURNING id INTO v_job_id;
    
    -- Update tenant configuration
    UPDATE tenant_isolation_config
    SET 
        migration_status = 'in_progress',
        updated_at = now(),
        updated_by = p_created_by
    WHERE tenant_id = p_tenant_id;
    
    -- Log the operation in audit
    INSERT INTO tenant_isolation_audit (
        operation_type,
        tenant_id,
        schema_name,
        performed_by,
        details,
        operation_status
    ) VALUES (
        'data_migration_initiated',
        p_tenant_id,
        v_schema_name,
        p_created_by,
        jsonb_build_object(
            'job_id', v_job_id
        ),
        'success'
    );
    
    -- Execute data migration asynchronously
    PERFORM pg_notify('tenant_migration_channel', v_job_id::text);
    
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
            'data_migration_initiated',
            p_tenant_id,
            v_schema_name,
            p_created_by,
            jsonb_build_object(
                'error', SQLERRM
            ),
            'failure',
            SQLERRM
        );
        
        RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION migrate_tenant_data(UUID, UUID) IS 
'Initiates the data migration process for a tenant from shared to isolated schema';

-- Function to clone table structure to tenant schema
CREATE OR REPLACE FUNCTION clone_table_structure(
    p_source_schema TEXT,
    p_target_schema TEXT,
    p_table_name TEXT,
    p_include_indexes BOOLEAN DEFAULT TRUE,
    p_include_constraints BOOLEAN DEFAULT TRUE
) RETURNS BOOLEAN AS $$
DECLARE
    v_sql TEXT;
    v_column_defs TEXT;
    v_index_defs TEXT[];
    v_constraint_defs TEXT[];
    v_index_def TEXT;
    v_constraint_def TEXT;
BEGIN
    -- Get column definitions
    SELECT string_agg(
        column_name || ' ' || 
        data_type || 
        CASE 
            WHEN character_maximum_length IS NOT NULL 
                THEN '(' || character_maximum_length || ')'
            ELSE ''
        END || 
        CASE 
            WHEN is_nullable = 'NO' 
                THEN ' NOT NULL'
            ELSE ''
        END,
        ', '
    )
    INTO v_column_defs
    FROM information_schema.columns
    WHERE table_schema = p_source_schema
    AND table_name = p_table_name;

    -- Create table structure
    v_sql := format('CREATE TABLE IF NOT EXISTS %I.%I (%s)', 
        p_target_schema, p_table_name, v_column_defs);
    EXECUTE v_sql;

    -- Add indexes if requested
    IF p_include_indexes THEN
        FOR v_index_def IN
            SELECT pg_get_indexdef(indexrelid)
            FROM pg_index i
            JOIN pg_class c ON i.indexrelid = c.oid
            JOIN pg_namespace n ON c.relnamespace = n.oid
            JOIN pg_class t ON i.indrelid = t.oid
            JOIN pg_namespace tn ON t.relnamespace = tn.oid
            WHERE tn.nspname = p_source_schema
            AND t.relname = p_table_name
            AND NOT i.indisprimary -- Exclude primary key, handled with constraints
        LOOP
            -- Modify the index definition to point to the target schema
            v_index_def := regexp_replace(
                v_index_def, 
                'ON ' || p_source_schema || '\.', 
                'ON ' || p_target_schema || '.');
            
            BEGIN
                EXECUTE v_index_def;
            EXCEPTION WHEN OTHERS THEN
                -- Log error but continue with other indexes
                RAISE WARNING 'Error creating index: % - %', v_index_def, SQLERRM;
            END;
        END LOOP;
    END IF;

    -- Add constraints if requested
    IF p_include_constraints THEN
        FOR v_constraint_def IN
            SELECT pg_get_constraintdef(c.oid)
            FROM pg_constraint c
            JOIN pg_class t ON c.conrelid = t.oid
            JOIN pg_namespace n ON t.relnamespace = n.oid
            WHERE n.nspname = p_source_schema
            AND t.relname = p_table_name
            AND c.contype IN ('p', 'u', 'f') -- primary, unique, foreign key
        LOOP
            BEGIN
                v_sql := format('ALTER TABLE %I.%I ADD %s',
                    p_target_schema, p_table_name, v_constraint_def);
                EXECUTE v_sql;
            EXCEPTION WHEN OTHERS THEN
                -- Log error but continue with other constraints
                RAISE WARNING 'Error adding constraint: % - %', v_sql, SQLERRM;
            END;
        END LOOP;
    END IF;

    RETURN TRUE;
EXCEPTION
    WHEN OTHERS THEN
        RAISE EXCEPTION 'Error cloning table structure: % - %', p_table_name, SQLERRM;
        RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION clone_table_structure(TEXT, TEXT, TEXT, BOOLEAN, BOOLEAN) IS 
'Clones the structure of a table from source schema to target schema';

-- Function to migrate data for a specific table
CREATE OR REPLACE FUNCTION migrate_table_data(
    p_tenant_id UUID,
    p_source_schema TEXT,
    p_target_schema TEXT,
    p_table_name TEXT,
    p_batch_size INTEGER DEFAULT 1000
) RETURNS BIGINT AS $$
DECLARE
    v_tenant_column TEXT;
    v_where_clause TEXT;
    v_primary_key TEXT;
    v_sql TEXT;
    v_column_list TEXT;
    v_row_count BIGINT := 0;
    v_affected_rows BIGINT := 0;
    v_current_offset INTEGER := 0;
    v_table_has_tenant BOOLEAN;
    v_batch_affected INTEGER;
BEGIN
    -- Check if table exists in source schema
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.tables
        WHERE table_schema = p_source_schema
        AND table_name = p_table_name
    ) THEN
        RAISE EXCEPTION 'Table does not exist in source schema: %.%', p_source_schema, p_table_name;
    END IF;
    
    -- Get the tenant column if it exists
    SELECT column_name INTO v_tenant_column
    FROM information_schema.columns
    WHERE table_schema = p_source_schema
    AND table_name = p_table_name
    AND column_name IN ('tenant_id', 'organization_id');
    
    v_table_has_tenant := v_tenant_column IS NOT NULL;
    
    -- Get primary key if exists
    SELECT a.attname INTO v_primary_key
    FROM pg_index i
    JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
    JOIN pg_class c ON i.indrelid = c.oid
    JOIN pg_namespace n ON c.relnamespace = n.oid
    WHERE n.nspname = p_source_schema
    AND c.relname = p_table_name
    AND i.indisprimary
    LIMIT 1;
    
    -- Get column list (excluding identity columns)
    SELECT string_agg(column_name, ', ')
    INTO v_column_list
    FROM information_schema.columns
    WHERE table_schema = p_source_schema
    AND table_name = p_table_name
    AND is_identity = 'NO';
    
    -- Construct WHERE clause based on tenant column
    IF v_table_has_tenant THEN
        v_where_clause := format('%I = ''%s''', v_tenant_column, p_tenant_id);
    ELSE
        v_where_clause := '1=1'; -- No tenant filter
    END IF;
    
    -- Count rows to migrate
    EXECUTE format('SELECT COUNT(*) FROM %I.%I WHERE %s',
        p_source_schema, p_table_name, v_where_clause)
    INTO v_row_count;
    
    -- If no rows to migrate, return early
    IF v_row_count = 0 THEN
        RETURN 0;
    END IF;
    
    -- Migrate data in batches
    WHILE v_affected_rows < v_row_count LOOP
        -- Build SQL for batch insert
        IF v_primary_key IS NOT NULL THEN
            -- Use ordered batching by primary key
            v_sql := format(
                'INSERT INTO %I.%I (%s) 
                SELECT %s 
                FROM %I.%I 
                WHERE %s 
                ORDER BY %I 
                LIMIT %s OFFSET %s',
                p_target_schema, p_table_name, v_column_list,
                v_column_list,
                p_source_schema, p_table_name,
                v_where_clause,
                v_primary_key,
                p_batch_size, v_current_offset
            );
        ELSE
            -- No primary key, use simple batching
            v_sql := format(
                'INSERT INTO %I.%I (%s) 
                SELECT %s 
                FROM %I.%I 
                WHERE %s 
                LIMIT %s OFFSET %s',
                p_target_schema, p_table_name, v_column_list,
                v_column_list,
                p_source_schema, p_table_name,
                v_where_clause,
                p_batch_size, v_current_offset
            );
        END IF;
        
        -- Execute batch insert
        EXECUTE v_sql;
        GET DIAGNOSTICS v_batch_affected = ROW_COUNT;
        
        -- Update counters for next batch
        v_affected_rows := v_affected_rows + v_batch_affected;
        v_current_offset := v_current_offset + p_batch_size;
        
        -- Exit if batch was smaller than batch size (last batch)
        EXIT WHEN v_batch_affected < p_batch_size;
    END LOOP;
    
    -- Record the mapping in tenant_schema_mappings
    INSERT INTO tenant_schema_mappings (
        tenant_id,
        source_schema,
        target_schema,
        table_name,
        mapping_type,
        migration_version
    ) VALUES (
        p_tenant_id,
        p_source_schema,
        p_target_schema,
        p_table_name,
        'direct',
        '1.0'
    );
    
    RETURN v_affected_rows;
EXCEPTION
    WHEN OTHERS THEN
        RAISE EXCEPTION 'Error migrating data for table %.%: %', p_source_schema, p_table_name, SQLERRM;
        RETURN -1;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION migrate_table_data(UUID, TEXT, TEXT, TEXT, INTEGER) IS 
'Migrates data for a specific table from source schema to target schema for a tenant';

-- Reset search path
RESET search_path;
