-- =========================================================================
-- INNOVABIZ - IAM - Healthcare Compliance Database Schema
-- =========================================================================
-- Autor: Eduardo Jeremias
-- Data: 08/05/2025
-- Versão: 1.0
--
-- Descrição: Esquema de banco de dados para armazenar informações de compliance
-- relacionadas ao módulo IAM Healthcare, incluindo resultados de validação,
-- histórico de auditorias e configurações para as diferentes regulamentações
-- regionais (HIPAA, GDPR, LGPD, PNDSB).
-- =========================================================================

-- Verify and create the schema if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'iam_healthcare') THEN
        CREATE SCHEMA iam_healthcare;
        
        -- Set proper permissions (adjust as needed based on your security model)
        GRANT USAGE ON SCHEMA iam_healthcare TO innovabiz_app;
        GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA iam_healthcare TO innovabiz_app;
        GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA iam_healthcare TO innovabiz_app;
        
        -- Create comment for documentation
        COMMENT ON SCHEMA iam_healthcare IS 'Schema for IAM Healthcare compliance data including validation results, audits, and regulatory configurations';
    END IF;
END
$$;

SET search_path TO iam_healthcare, public;

-- =========================================================================
-- Regulatory Frameworks
-- Stores information about the various regulatory frameworks supported
-- =========================================================================
CREATE TABLE IF NOT EXISTS regulatory_frameworks (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    region VARCHAR(5) NOT NULL,
    sector VARCHAR(50) NOT NULL DEFAULT 'HEALTHCARE',
    version VARCHAR(20),
    implementation_date DATE,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    active BOOLEAN DEFAULT TRUE,
    metadata JSONB
);

COMMENT ON TABLE regulatory_frameworks IS 'Reference table for healthcare regulatory frameworks supported by the system';

-- Create index for efficient searches
CREATE INDEX IF NOT EXISTS idx_reg_framework_region ON regulatory_frameworks(region);
CREATE INDEX IF NOT EXISTS idx_reg_framework_sector ON regulatory_frameworks(sector);

-- =========================================================================
-- Compliance Validators
-- Stores information about the validators implemented in the system
-- =========================================================================
CREATE TABLE IF NOT EXISTS compliance_validators (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    validator_class VARCHAR(255) NOT NULL,
    version VARCHAR(20) NOT NULL,
    framework_id VARCHAR(50) REFERENCES regulatory_frameworks(id),
    region VARCHAR(5) NOT NULL,
    sector VARCHAR(50) NOT NULL DEFAULT 'HEALTHCARE',
    configuration JSONB,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    active BOOLEAN DEFAULT TRUE,
    metadata JSONB
);

COMMENT ON TABLE compliance_validators IS 'Configuration for compliance validators including their mapping to regulatory frameworks';

-- Create indexes for efficient searches
CREATE INDEX IF NOT EXISTS idx_validator_framework ON compliance_validators(framework_id);
CREATE INDEX IF NOT EXISTS idx_validator_region ON compliance_validators(region);
CREATE INDEX IF NOT EXISTS idx_validator_active ON compliance_validators(active);

-- =========================================================================
-- Validation Results
-- Stores the results of compliance validations
-- =========================================================================
CREATE TABLE IF NOT EXISTS validation_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    validator_id VARCHAR(50) REFERENCES compliance_validators(id),
    target_id VARCHAR(255) NOT NULL,
    target_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,
    score NUMERIC(5, 2),
    issues_count INTEGER,
    validation_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    result_data JSONB,
    issues JSONB,
    metadata JSONB
);

COMMENT ON TABLE validation_results IS 'Stores the results of compliance validations performed by validators';

-- Create indexes for efficient searches and time-series queries
CREATE INDEX IF NOT EXISTS idx_val_result_validator ON validation_results(validator_id);
CREATE INDEX IF NOT EXISTS idx_val_result_target ON validation_results(target_id, target_type);
CREATE INDEX IF NOT EXISTS idx_val_result_status ON validation_results(status);
CREATE INDEX IF NOT EXISTS idx_val_result_time ON validation_results(validation_time);

-- =========================================================================
-- Compliance Issues
-- Stores detailed information about compliance issues found
-- =========================================================================
CREATE TABLE IF NOT EXISTS compliance_issues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    validation_result_id UUID REFERENCES validation_results(id) ON DELETE CASCADE,
    issue_id VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    description TEXT NOT NULL,
    recommendation TEXT,
    reference TEXT,
    status VARCHAR(20) DEFAULT 'OPEN',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolution_notes TEXT,
    metadata JSONB
);

COMMENT ON TABLE compliance_issues IS 'Detailed information about compliance issues identified during validation';

-- Create indexes for efficient searches
CREATE INDEX IF NOT EXISTS idx_issue_validation ON compliance_issues(validation_result_id);
CREATE INDEX IF NOT EXISTS idx_issue_status ON compliance_issues(status);
CREATE INDEX IF NOT EXISTS idx_issue_severity ON compliance_issues(severity);
CREATE INDEX IF NOT EXISTS idx_issue_created ON compliance_issues(created_at);

-- =========================================================================
-- Compliance Alerts
-- Stores alerts generated based on compliance validation results
-- =========================================================================
CREATE TABLE IF NOT EXISTS compliance_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    validator_id VARCHAR(50) REFERENCES compliance_validators(id),
    target_id VARCHAR(255) NOT NULL,
    target_type VARCHAR(100) NOT NULL,
    alert_level VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    reference TEXT,
    recommendation TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    acknowledged_by VARCHAR(100),
    status VARCHAR(20) DEFAULT 'ACTIVE',
    metadata JSONB
);

COMMENT ON TABLE compliance_alerts IS 'Alerts generated from compliance validation results that require attention';

-- Create indexes for efficient searches
CREATE INDEX IF NOT EXISTS idx_alert_validator ON compliance_alerts(validator_id);
CREATE INDEX IF NOT EXISTS idx_alert_target ON compliance_alerts(target_id, target_type);
CREATE INDEX IF NOT EXISTS idx_alert_level ON compliance_alerts(alert_level);
CREATE INDEX IF NOT EXISTS idx_alert_status ON compliance_alerts(status);
CREATE INDEX IF NOT EXISTS idx_alert_created ON compliance_alerts(created_at);

-- =========================================================================
-- Compliance Reports
-- Stores information about generated compliance reports
-- =========================================================================
CREATE TABLE IF NOT EXISTS compliance_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id VARCHAR(100) NOT NULL,
    report_type VARCHAR(100) NOT NULL,
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    generated_by VARCHAR(100),
    framework_ids VARCHAR(50)[] NOT NULL,
    report_format VARCHAR(20) DEFAULT 'JSON',
    report_data JSONB,
    summary TEXT,
    status VARCHAR(20) DEFAULT 'ACTIVE',
    metadata JSONB
);

COMMENT ON TABLE compliance_reports IS 'Information about generated compliance reports including their content and metadata';

-- Create indexes for efficient searches
CREATE INDEX IF NOT EXISTS idx_report_type ON compliance_reports(report_type);
CREATE INDEX IF NOT EXISTS idx_report_generated ON compliance_reports(generated_at);
CREATE INDEX IF NOT EXISTS idx_report_id ON compliance_reports(report_id);

-- =========================================================================
-- Regulatory Changes
-- Tracks changes to regulatory frameworks for compliance monitoring
-- =========================================================================
CREATE TABLE IF NOT EXISTS regulatory_changes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    framework_id VARCHAR(50) REFERENCES regulatory_frameworks(id),
    change_type VARCHAR(50) NOT NULL,
    change_summary TEXT NOT NULL,
    change_details TEXT,
    change_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    effective_date TIMESTAMP WITH TIME ZONE,
    source_url TEXT,
    impact_level VARCHAR(20) DEFAULT 'MEDIUM',
    affected_components JSONB,
    status VARCHAR(20) DEFAULT 'PENDING',
    metadata JSONB
);

COMMENT ON TABLE regulatory_changes IS 'Tracks changes to regulatory frameworks to monitor compliance requirements over time';

-- Create indexes for efficient searches
CREATE INDEX IF NOT EXISTS idx_reg_change_framework ON regulatory_changes(framework_id);
CREATE INDEX IF NOT EXISTS idx_reg_change_date ON regulatory_changes(change_date);
CREATE INDEX IF NOT EXISTS idx_reg_change_effective ON regulatory_changes(effective_date);
CREATE INDEX IF NOT EXISTS idx_reg_change_impact ON regulatory_changes(impact_level);

-- =========================================================================
-- FHIR Resources Configuration
-- Stores configuration for FHIR resource validation
-- =========================================================================
CREATE TABLE IF NOT EXISTS fhir_resource_configurations (
    id VARCHAR(50) PRIMARY KEY,
    resource_type VARCHAR(50) NOT NULL,
    description TEXT,
    required_fields JSONB NOT NULL,
    region_specific_requirements JSONB,
    version VARCHAR(20) DEFAULT 'R4',
    validation_rules JSONB,
    profile_urls JSONB,
    terminology_bindings JSONB,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    active BOOLEAN DEFAULT TRUE,
    metadata JSONB
);

COMMENT ON TABLE fhir_resource_configurations IS 'Configuration for validating FHIR resources including required fields and regional variations';

-- Create indexes for efficient searches
CREATE INDEX IF NOT EXISTS idx_fhir_resource_type ON fhir_resource_configurations(resource_type);
CREATE INDEX IF NOT EXISTS idx_fhir_version ON fhir_resource_configurations(version);

-- =========================================================================
-- Dashboard Configurations
-- Stores configurations for the compliance dashboard
-- =========================================================================
CREATE TABLE IF NOT EXISTS dashboard_configurations (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    config_data JSONB NOT NULL,
    user_id VARCHAR(100),
    role_id VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    active BOOLEAN DEFAULT TRUE,
    metadata JSONB
);

COMMENT ON TABLE dashboard_configurations IS 'Configuration settings for the compliance dashboard including layouts and metrics';

-- Create indexes for efficient searches
CREATE INDEX IF NOT EXISTS idx_dashboard_user ON dashboard_configurations(user_id);
CREATE INDEX IF NOT EXISTS idx_dashboard_role ON dashboard_configurations(role_id);
CREATE INDEX IF NOT EXISTS idx_dashboard_active ON dashboard_configurations(active);

-- =========================================================================
-- Initial data insertion for regulatory frameworks
-- =========================================================================
INSERT INTO regulatory_frameworks (id, name, description, region, version, implementation_date)
VALUES
    ('HIPAA', 'Health Insurance Portability and Accountability Act', 'US healthcare privacy and security regulation', 'US', '1996', '1996-08-21'),
    ('GDPR', 'General Data Protection Regulation', 'EU data protection and privacy regulation', 'EU', '2016/679', '2018-05-25'),
    ('LGPD', 'Lei Geral de Proteção de Dados', 'Brazilian general data protection law', 'BR', '13.709/2018', '2020-09-18'),
    ('PNDSB', 'Política Nacional de Desenvolvimento da Saúde', 'Angola health development policy', 'AO', '2012', '2012-01-01'),
    ('FHIR-R4', 'HL7 FHIR Release 4', 'Healthcare interoperability standard', 'GLOBAL', '4.0.1', '2019-10-30')
ON CONFLICT (id) DO NOTHING;

-- =========================================================================
-- Initial data insertion for compliance validators
-- =========================================================================
INSERT INTO compliance_validators (id, name, description, validator_class, version, framework_id, region, sector)
VALUES
    ('HIPAA-HEALTHCARE', 'HIPAA Healthcare Validator', 'Validates compliance with HIPAA for healthcare data', 'HIPAAHealthcareValidator', '1.0', 'HIPAA', 'US', 'HEALTHCARE'),
    ('GDPR-HEALTHCARE', 'GDPR Healthcare Validator', 'Validates compliance with GDPR for healthcare data', 'GDPRHealthcareValidator', '1.0', 'GDPR', 'EU', 'HEALTHCARE'),
    ('LGPD-HEALTHCARE', 'LGPD Healthcare Validator', 'Validates compliance with LGPD for healthcare data', 'LGPDHealthcareValidator', '1.0', 'LGPD', 'BR', 'HEALTHCARE'),
    ('ANGOLA-HEALTHCARE', 'Angola PNDSB Validator', 'Validates compliance with PNDSB for healthcare data', 'AngolaHealthcareValidator', '1.0', 'PNDSB', 'AO', 'HEALTHCARE'),
    ('FHIR-VALIDATOR-R4', 'HL7 FHIR R4 Validator', 'Validates compliance with FHIR R4 standard', 'FHIRValidatorR4', '1.0', 'FHIR-R4', 'GLOBAL', 'HEALTHCARE')
ON CONFLICT (id) DO NOTHING;
