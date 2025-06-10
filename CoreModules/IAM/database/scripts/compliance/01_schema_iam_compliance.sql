-- INNOVABIZ IAM Module - Compliance Schema
-- Version: 1.0.0
-- Date: 09/05/2025
-- Author: Eduardo Jeremias
-- Description: Schema and tables for IAM compliance validation

-- Create IAM Compliance Schema
CREATE SCHEMA IF NOT EXISTS iam_compliance;
COMMENT ON SCHEMA iam_compliance IS 'Schema for IAM compliance validation and reporting';

-- Set search path
SET search_path TO iam_compliance, iam, public;

-- Create Compliance Policy table
CREATE TABLE compliance_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(id) ON DELETE CASCADE,
    policy_name TEXT NOT NULL,
    policy_code TEXT NOT NULL,
    version TEXT NOT NULL,
    description TEXT,
    framework TEXT NOT NULL,
    jurisdiction TEXT NOT NULL,
    industry_sector TEXT[],
    effective_date DATE NOT NULL,
    expiration_date DATE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID NOT NULL REFERENCES iam.users(id),
    updated_by UUID NOT NULL REFERENCES iam.users(id),
    metadata JSONB,
    validation_rules JSONB NOT NULL,
    severity_level TEXT NOT NULL CHECK (severity_level IN ('low', 'medium', 'high', 'critical')),
    UNIQUE(tenant_id, policy_code, version)
);

COMMENT ON TABLE compliance_policies IS 'Policies defining compliance requirements for IAM configurations';

-- Create Compliance Check Results table
CREATE TABLE compliance_check_results (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES iam.tenants(id) ON DELETE CASCADE,
    policy_id UUID NOT NULL REFERENCES compliance_policies(id),
    check_date TIMESTAMPTZ NOT NULL DEFAULT now(),
    initiated_by UUID NOT NULL REFERENCES iam.users(id),
    status TEXT NOT NULL CHECK (status IN ('passed', 'failed', 'warning', 'error')),
    details JSONB,
    score NUMERIC(5,2),
    failure_count INTEGER DEFAULT 0,
    warning_count INTEGER DEFAULT 0,
    remediation_steps JSONB,
    next_check_date TIMESTAMPTZ,
    execution_time_ms INTEGER,
    report_id UUID
);

COMMENT ON TABLE compliance_check_results IS 'Results of compliance checks run against IAM configurations';

-- Create Compliance Validator Registry table
CREATE TABLE compliance_validators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    validator_name TEXT NOT NULL,
    validator_type TEXT NOT NULL,
    description TEXT,
    framework TEXT NOT NULL,
    jurisdiction TEXT[] NOT NULL,
    industry_sector TEXT[],
    function_schema TEXT NOT NULL,
    function_name TEXT NOT NULL,
    parameters JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_active BOOLEAN NOT NULL DEFAULT true,
    version TEXT NOT NULL,
    dependencies TEXT[],
    author TEXT,
    documentation_url TEXT,
    UNIQUE(validator_name, version)
);

COMMENT ON TABLE compliance_validators IS 'Registry of available compliance validators for IAM';

-- Create Compliance Exemptions table
CREATE TABLE compliance_exemptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(id) ON DELETE CASCADE,
    policy_id UUID NOT NULL REFERENCES compliance_policies(id),
    resource_type TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    justification TEXT NOT NULL,
    approved_by UUID NOT NULL REFERENCES iam.users(id),
    approved_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expiration_date DATE,
    evidence_document_id UUID,
    risk_assessment JSONB,
    mitigation_controls TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('pending', 'active', 'expired', 'revoked')),
    UNIQUE(tenant_id, policy_id, resource_type, resource_id)
);

COMMENT ON TABLE compliance_exemptions IS 'Exemptions from specific compliance policies for IAM resources';

-- Create Compliance Reports table
CREATE TABLE compliance_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(id) ON DELETE CASCADE,
    report_name TEXT NOT NULL,
    report_type TEXT NOT NULL,
    generation_date TIMESTAMPTZ NOT NULL DEFAULT now(),
    generated_by UUID NOT NULL REFERENCES iam.users(id),
    parameters JSONB,
    format TEXT NOT NULL CHECK (format IN ('pdf', 'html', 'json', 'csv', 'excel')),
    file_path TEXT,
    file_hash TEXT,
    size_bytes INTEGER,
    status TEXT NOT NULL DEFAULT 'completed' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    error_message TEXT,
    summary JSONB,
    expiration_date TIMESTAMPTZ,
    is_archived BOOLEAN NOT NULL DEFAULT false,
    shared_with UUID[]
);

COMMENT ON TABLE compliance_reports IS 'Generated compliance reports for IAM configurations';

-- Create Compliance Issue Tracking table
CREATE TABLE compliance_issues (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES iam.tenants(id) ON DELETE CASCADE,
    policy_id UUID NOT NULL REFERENCES compliance_policies(id),
    check_result_id BIGINT REFERENCES compliance_check_results(id),
    issue_type TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    description TEXT NOT NULL,
    severity TEXT NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    detected_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'in_progress', 'resolved', 'wontfix', 'false_positive')),
    assigned_to UUID REFERENCES iam.users(id),
    resolved_by UUID REFERENCES iam.users(id),
    resolved_at TIMESTAMPTZ,
    resolution_steps TEXT,
    remediation_plan TEXT,
    due_date DATE,
    exemption_id UUID REFERENCES compliance_exemptions(id),
    metadata JSONB
);

COMMENT ON TABLE compliance_issues IS 'Tracking of compliance issues detected in IAM configurations';

-- Create indexes
CREATE INDEX idx_compliance_policies_tenant ON compliance_policies(tenant_id);
CREATE INDEX idx_compliance_policies_framework ON compliance_policies(framework);
CREATE INDEX idx_compliance_policies_jurisdiction ON compliance_policies(jurisdiction);
CREATE INDEX idx_compliance_check_results_tenant ON compliance_check_results(tenant_id);
CREATE INDEX idx_compliance_check_results_policy ON compliance_check_results(policy_id);
CREATE INDEX idx_compliance_check_results_date ON compliance_check_results(check_date);
CREATE INDEX idx_compliance_check_results_status ON compliance_check_results(status);
CREATE INDEX idx_compliance_validators_framework ON compliance_validators(framework);
CREATE INDEX idx_compliance_validators_jurisdiction ON compliance_validators(jurisdiction);
CREATE INDEX idx_compliance_exemptions_tenant ON compliance_exemptions(tenant_id);
CREATE INDEX idx_compliance_exemptions_policy ON compliance_exemptions(policy_id);
CREATE INDEX idx_compliance_exemptions_resource ON compliance_exemptions(resource_type, resource_id);
CREATE INDEX idx_compliance_exemptions_status ON compliance_exemptions(status);
CREATE INDEX idx_compliance_reports_tenant ON compliance_reports(tenant_id);
CREATE INDEX idx_compliance_reports_date ON compliance_reports(generation_date);
CREATE INDEX idx_compliance_reports_type ON compliance_reports(report_type);
CREATE INDEX idx_compliance_issues_tenant ON compliance_issues(tenant_id);
CREATE INDEX idx_compliance_issues_policy ON compliance_issues(policy_id);
CREATE INDEX idx_compliance_issues_severity ON compliance_issues(severity);
CREATE INDEX idx_compliance_issues_status ON compliance_issues(status);
CREATE INDEX idx_compliance_issues_resource ON compliance_issues(resource_type, resource_id);

-- Create view for active compliance policies
CREATE OR REPLACE VIEW active_compliance_policies AS
SELECT
    cp.*,
    COUNT(DISTINCT ce.id) AS exemption_count,
    COUNT(DISTINCT ci.id) FILTER (WHERE ci.status = 'open') AS open_issues_count
FROM 
    compliance_policies cp
LEFT JOIN 
    compliance_exemptions ce ON cp.id = ce.policy_id AND ce.status = 'active'
LEFT JOIN 
    compliance_issues ci ON cp.id = ci.policy_id
WHERE 
    cp.is_active = true
    AND (cp.expiration_date IS NULL OR cp.expiration_date >= CURRENT_DATE)
GROUP BY
    cp.id;

COMMENT ON VIEW active_compliance_policies IS 'Active compliance policies with counts of exemptions and open issues';

-- Create view for compliance overview by tenant
CREATE OR REPLACE VIEW tenant_compliance_overview AS
SELECT
    t.id AS tenant_id,
    t.tenant_name,
    COUNT(DISTINCT cp.id) AS total_policies,
    COUNT(DISTINCT cp.id) FILTER (WHERE cp.is_active = true) AS active_policies,
    COUNT(DISTINCT ccr.id) FILTER (WHERE ccr.status = 'passed') AS passed_checks,
    COUNT(DISTINCT ccr.id) FILTER (WHERE ccr.status = 'failed') AS failed_checks,
    COUNT(DISTINCT ci.id) FILTER (WHERE ci.status = 'open') AS open_issues,
    COUNT(DISTINCT ci.id) FILTER (WHERE ci.status = 'open' AND ci.severity = 'critical') AS critical_issues,
    AVG(ccr.score) FILTER (WHERE ccr.check_date >= (CURRENT_DATE - INTERVAL '90 days')) AS avg_compliance_score,
    MAX(ccr.check_date) AS last_check_date,
    COUNT(DISTINCT ce.id) FILTER (WHERE ce.status = 'active') AS active_exemptions
FROM 
    iam.tenants t
LEFT JOIN 
    compliance_policies cp ON t.id = cp.tenant_id
LEFT JOIN 
    compliance_check_results ccr ON t.id = ccr.tenant_id
LEFT JOIN 
    compliance_issues ci ON t.id = ci.tenant_id
LEFT JOIN 
    compliance_exemptions ce ON t.id = ce.tenant_id
GROUP BY
    t.id, t.tenant_name;

COMMENT ON VIEW tenant_compliance_overview IS 'Compliance overview statistics by tenant';

-- Enable RLS on tables
ALTER TABLE compliance_policies ENABLE ROW LEVEL SECURITY;
ALTER TABLE compliance_check_results ENABLE ROW LEVEL SECURITY;
ALTER TABLE compliance_exemptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE compliance_reports ENABLE ROW LEVEL SECURITY;
ALTER TABLE compliance_issues ENABLE ROW LEVEL SECURITY;

-- Create RLS policies
CREATE POLICY tenant_compliance_policies_policy ON compliance_policies
    USING (tenant_id = current_setting('app.current_tenant')::uuid OR pg_has_role(current_user, 'iam_compliance_admin', 'member'));

CREATE POLICY tenant_compliance_check_results_policy ON compliance_check_results
    USING (tenant_id = current_setting('app.current_tenant')::uuid OR pg_has_role(current_user, 'iam_compliance_admin', 'member'));

CREATE POLICY tenant_compliance_exemptions_policy ON compliance_exemptions
    USING (tenant_id = current_setting('app.current_tenant')::uuid OR pg_has_role(current_user, 'iam_compliance_admin', 'member'));

CREATE POLICY tenant_compliance_reports_policy ON compliance_reports
    USING (tenant_id = current_setting('app.current_tenant')::uuid OR pg_has_role(current_user, 'iam_compliance_admin', 'member'));

CREATE POLICY tenant_compliance_issues_policy ON compliance_issues
    USING (tenant_id = current_setting('app.current_tenant')::uuid OR pg_has_role(current_user, 'iam_compliance_admin', 'member'));

-- Create roles
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'iam_compliance_admin') THEN
        CREATE ROLE iam_compliance_admin;
    END IF;
    
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'iam_compliance_reader') THEN
        CREATE ROLE iam_compliance_reader;
    END IF;
END
$$;

-- Grant permissions
GRANT USAGE ON SCHEMA iam_compliance TO iam_compliance_admin;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA iam_compliance TO iam_compliance_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA iam_compliance TO iam_compliance_admin;

GRANT USAGE ON SCHEMA iam_compliance TO iam_compliance_reader;
GRANT SELECT ON ALL TABLES IN SCHEMA iam_compliance TO iam_compliance_reader;

-- Reset search path
RESET search_path;
