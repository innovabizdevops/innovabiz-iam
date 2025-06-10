-- INNOVABIZ IAM Module - Multi-Tenant Schema Definition
-- Version: 1.0.0
-- Date: 09/05/2025
-- Author: Eduardo Jeremias
-- Description: Creates schema and base tables for advanced multi-tenant functionality

-- Create Multi-Tenant Schema
CREATE SCHEMA IF NOT EXISTS iam_multi_tenant;
COMMENT ON SCHEMA iam_multi_tenant IS 'Schema for IAM multi-tenant advanced functionality and isolation';

-- Set search path
SET search_path TO iam_multi_tenant, iam, public;

-- Create tenant isolation table
CREATE TABLE tenant_isolation_config (
    tenant_id UUID PRIMARY KEY,
    schema_name TEXT NOT NULL UNIQUE,
    isolation_level TEXT NOT NULL CHECK (isolation_level IN ('shared', 'isolated', 'hybrid')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID NOT NULL REFERENCES iam.users(id),
    updated_by UUID NOT NULL REFERENCES iam.users(id),
    is_active BOOLEAN NOT NULL DEFAULT true,
    isolation_reason TEXT,
    custom_connection_params JSONB,
    migration_status TEXT DEFAULT 'pending' CHECK (migration_status IN ('pending', 'in_progress', 'completed', 'failed')),
    compliance_requirements TEXT[],
    data_residency_region TEXT,
    CONSTRAINT valid_schema_name CHECK (schema_name ~ '^[a-z][a-z0-9_]*$')
);

COMMENT ON TABLE tenant_isolation_config IS 'Configuration for tenant isolation levels and associated schemas';

-- Create tenant schema mapping
CREATE TABLE tenant_schema_mappings (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenant_isolation_config(tenant_id),
    source_schema TEXT NOT NULL,
    target_schema TEXT NOT NULL,
    table_name TEXT NOT NULL,
    mapping_type TEXT NOT NULL CHECK (mapping_type IN ('direct', 'view', 'function', 'policy')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_active BOOLEAN NOT NULL DEFAULT true,
    migration_version TEXT,
    migration_hash TEXT,
    CONSTRAINT unique_schema_mapping UNIQUE (tenant_id, source_schema, target_schema, table_name)
);

COMMENT ON TABLE tenant_schema_mappings IS 'Mappings between source schemas and tenant-specific schemas';

-- Create cross-tenant permissions table
CREATE TABLE cross_tenant_permissions (
    id BIGSERIAL PRIMARY KEY,
    source_tenant_id UUID NOT NULL REFERENCES tenant_isolation_config(tenant_id),
    target_tenant_id UUID NOT NULL REFERENCES tenant_isolation_config(tenant_id),
    permission_set TEXT[] NOT NULL,
    resource_pattern TEXT,
    operation_type TEXT[] NOT NULL,
    granted_by UUID NOT NULL REFERENCES iam.users(id),
    granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    valid_from TIMESTAMPTZ NOT NULL DEFAULT now(),
    valid_until TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true,
    approval_workflow_id UUID,
    audit_trail_id UUID,
    CONSTRAINT different_tenants CHECK (source_tenant_id <> target_tenant_id)
);

COMMENT ON TABLE cross_tenant_permissions IS 'Permissions granted between different tenants for B2B scenarios';

-- Create tenant migration jobs table
CREATE TABLE tenant_migration_jobs (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenant_isolation_config(tenant_id),
    job_type TEXT NOT NULL CHECK (job_type IN ('create_schema', 'migrate_data', 'create_views', 'create_policies', 'validation', 'rollback')),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed', 'failed', 'cancelled')),
    progress NUMERIC(5,2) DEFAULT 0,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error_message TEXT,
    created_by UUID NOT NULL REFERENCES iam.users(id),
    execution_log JSONB,
    affected_tables_count INTEGER DEFAULT 0,
    affected_rows_count BIGINT DEFAULT 0,
    next_job_id BIGINT,
    previous_job_id BIGINT,
    CONSTRAINT valid_next_job FOREIGN KEY (next_job_id) REFERENCES tenant_migration_jobs(id),
    CONSTRAINT valid_previous_job FOREIGN KEY (previous_job_id) REFERENCES tenant_migration_jobs(id)
);

COMMENT ON TABLE tenant_migration_jobs IS 'Jobs for tenant migration and schema isolation processes';

-- Create audit log for tenant isolation operations
CREATE TABLE tenant_isolation_audit (
    id BIGSERIAL PRIMARY KEY,
    operation_type TEXT NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenant_isolation_config(tenant_id),
    schema_name TEXT NOT NULL,
    performed_by UUID NOT NULL REFERENCES iam.users(id),
    performed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    affected_object TEXT,
    operation_status TEXT NOT NULL CHECK (operation_status IN ('success', 'failure', 'partial')),
    error_details TEXT
);

COMMENT ON TABLE tenant_isolation_audit IS 'Audit trail for all tenant isolation operations';

-- Create tenant data metrics table
CREATE TABLE tenant_data_metrics (
    tenant_id UUID NOT NULL REFERENCES tenant_isolation_config(tenant_id),
    collection_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    schema_name TEXT NOT NULL,
    table_count INTEGER NOT NULL DEFAULT 0,
    total_size_bytes BIGINT NOT NULL DEFAULT 0,
    row_count_estimate BIGINT NOT NULL DEFAULT 0,
    index_size_bytes BIGINT NOT NULL DEFAULT 0,
    query_count_24h BIGINT NOT NULL DEFAULT 0,
    avg_query_time_ms NUMERIC(10,2),
    max_query_time_ms NUMERIC(10,2),
    PRIMARY KEY (tenant_id, collection_time)
);

COMMENT ON TABLE tenant_data_metrics IS 'Metrics about tenant data size and usage for capacity planning';

-- Create RLS policy templates
CREATE TABLE rls_policy_templates (
    id BIGSERIAL PRIMARY KEY,
    template_name TEXT NOT NULL UNIQUE,
    description TEXT,
    table_pattern TEXT NOT NULL,
    schema_pattern TEXT,
    policy_definition TEXT NOT NULL,
    policy_type TEXT NOT NULL CHECK (policy_type IN ('permissive', 'restrictive')),
    command_type TEXT NOT NULL CHECK (command_type IN ('all', 'select', 'insert', 'update', 'delete')),
    is_parameterized BOOLEAN NOT NULL DEFAULT false,
    parameter_definitions JSONB,
    created_by UUID NOT NULL REFERENCES iam.users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_active BOOLEAN NOT NULL DEFAULT true
);

COMMENT ON TABLE rls_policy_templates IS 'Templates for RLS policies that can be applied to tenant schemas';

-- Create indexes
CREATE INDEX idx_tenant_isolation_config_schema ON tenant_isolation_config(schema_name);
CREATE INDEX idx_tenant_schema_mappings_tenant ON tenant_schema_mappings(tenant_id);
CREATE INDEX idx_cross_tenant_permissions_source ON cross_tenant_permissions(source_tenant_id);
CREATE INDEX idx_cross_tenant_permissions_target ON cross_tenant_permissions(target_tenant_id);
CREATE INDEX idx_tenant_migration_jobs_tenant ON tenant_migration_jobs(tenant_id, status);
CREATE INDEX idx_tenant_isolation_audit_tenant ON tenant_isolation_audit(tenant_id);
CREATE INDEX idx_tenant_data_metrics_tenant ON tenant_data_metrics(tenant_id);
CREATE INDEX idx_rls_policy_templates_active ON rls_policy_templates(is_active);

-- Create views for convenient querying
CREATE OR REPLACE VIEW tenant_migration_status AS
SELECT 
    tic.tenant_id,
    tic.schema_name,
    tic.isolation_level,
    tic.migration_status,
    COUNT(DISTINCT tmj.id) AS total_jobs,
    COUNT(DISTINCT CASE WHEN tmj.status = 'completed' THEN tmj.id END) AS completed_jobs,
    COUNT(DISTINCT CASE WHEN tmj.status = 'failed' THEN tmj.id END) AS failed_jobs,
    COUNT(DISTINCT CASE WHEN tmj.status IN ('pending', 'in_progress') THEN tmj.id END) AS pending_jobs,
    MAX(tmj.completed_at) AS last_job_completion,
    MIN(CASE WHEN tmj.status IN ('pending', 'in_progress') THEN tmj.started_at END) AS next_job_start,
    SUM(CASE WHEN tmj.status = 'completed' THEN tmj.affected_rows_count ELSE 0 END) AS total_migrated_rows
FROM 
    tenant_isolation_config tic
LEFT JOIN 
    tenant_migration_jobs tmj ON tic.tenant_id = tmj.tenant_id
GROUP BY 
    tic.tenant_id, tic.schema_name, tic.isolation_level, tic.migration_status;

COMMENT ON VIEW tenant_migration_status IS 'Aggregated view of tenant migration status and progress';

-- Grant permissions
GRANT USAGE ON SCHEMA iam_multi_tenant TO iam_admin;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA iam_multi_tenant TO iam_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA iam_multi_tenant TO iam_admin;
GRANT USAGE ON SCHEMA iam_multi_tenant TO iam_reader;
GRANT SELECT ON ALL TABLES IN SCHEMA iam_multi_tenant TO iam_reader;

-- Set RLS on relevant tables
ALTER TABLE tenant_isolation_config ENABLE ROW LEVEL SECURITY;
ALTER TABLE cross_tenant_permissions ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_isolation_audit ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_data_metrics ENABLE ROW LEVEL SECURITY;

-- Create RLS policies
CREATE POLICY tenant_isolation_admin_policy ON tenant_isolation_config
    USING (pg_has_role(current_user, 'iam_admin', 'member'));

CREATE POLICY cross_tenant_permissions_visibility ON cross_tenant_permissions
    USING (
        pg_has_role(current_user, 'iam_admin', 'member') OR
        source_tenant_id = current_setting('app.current_tenant')::uuid OR
        target_tenant_id = current_setting('app.current_tenant')::uuid
    );

CREATE POLICY tenant_isolation_audit_admin_policy ON tenant_isolation_audit
    USING (pg_has_role(current_user, 'iam_admin', 'member'));

CREATE POLICY tenant_data_metrics_visibility ON tenant_data_metrics
    USING (
        pg_has_role(current_user, 'iam_admin', 'member') OR
        tenant_id = current_setting('app.current_tenant')::uuid
    );

-- Reset search path
RESET search_path;
