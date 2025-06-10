-- =========================================================================
-- INNOVABIZ - IAM - Healthcare Compliance Database Views
-- =========================================================================
-- Autor: Eduardo Jeremias
-- Data: 08/05/2025
-- Versão: 1.0
--
-- Descrição: Views para facilitar o acesso e análise de dados de compliance
-- do módulo IAM Healthcare, seguindo as melhores práticas de governança
-- de dados e conformidade com regulamentações internacionais.
-- =========================================================================

SET search_path TO iam_healthcare, public;

-- =========================================================================
-- View: Validation Results Overview
-- Provides a consolidated overview of validation results with framework details
-- =========================================================================
CREATE OR REPLACE VIEW vw_validation_results_overview AS
SELECT 
    vr.id AS validation_id,
    vr.target_id,
    vr.target_type,
    vr.status,
    vr.score,
    vr.issues_count,
    vr.validation_time,
    cv.id AS validator_id,
    cv.name AS validator_name,
    rf.id AS framework_id,
    rf.name AS framework_name,
    rf.region,
    rf.sector,
    CASE 
        WHEN vr.score >= 90 THEN 'COMPLIANT'
        WHEN vr.score >= 70 THEN 'PARTIAL_COMPLIANCE'
        ELSE 'NON_COMPLIANT'
    END AS compliance_level,
    vr.metadata
FROM 
    validation_results vr
JOIN 
    compliance_validators cv ON vr.validator_id = cv.id
JOIN 
    regulatory_frameworks rf ON cv.framework_id = rf.id;

COMMENT ON VIEW vw_validation_results_overview IS 'Consolidated view of validation results with framework details for compliance reporting';

-- =========================================================================
-- View: Open Compliance Issues
-- Lists all unresolved compliance issues with severity and framework details
-- =========================================================================
CREATE OR REPLACE VIEW vw_open_compliance_issues AS
SELECT 
    ci.id AS issue_id,
    ci.issue_id AS issue_code,
    ci.severity,
    ci.description,
    ci.recommendation,
    ci.reference,
    ci.created_at,
    vr.target_id,
    vr.target_type,
    cv.id AS validator_id,
    cv.name AS validator_name,
    rf.id AS framework_id,
    rf.name AS framework_name,
    rf.region
FROM 
    compliance_issues ci
JOIN 
    validation_results vr ON ci.validation_result_id = vr.id
JOIN 
    compliance_validators cv ON vr.validator_id = cv.id
JOIN 
    regulatory_frameworks rf ON cv.framework_id = rf.id
WHERE 
    ci.status = 'OPEN'
ORDER BY 
    CASE 
        WHEN ci.severity = 'critical' THEN 1
        WHEN ci.severity = 'high' THEN 2
        WHEN ci.severity = 'medium' THEN 3
        WHEN ci.severity = 'low' THEN 4
        ELSE 5
    END,
    ci.created_at;

COMMENT ON VIEW vw_open_compliance_issues IS 'Lists all open compliance issues sorted by severity and creation date';

-- =========================================================================
-- View: Compliance Trends
-- Shows compliance trends over time by framework and region
-- =========================================================================
CREATE OR REPLACE VIEW vw_compliance_trends AS
SELECT 
    DATE_TRUNC('day', vr.validation_time) AS validation_date,
    rf.id AS framework_id,
    rf.name AS framework_name,
    rf.region,
    COUNT(vr.id) AS validations_count,
    AVG(vr.score) AS avg_score,
    SUM(CASE WHEN vr.score >= 90 THEN 1 ELSE 0 END) AS compliant_count,
    SUM(CASE WHEN vr.score >= 70 AND vr.score < 90 THEN 1 ELSE 0 END) AS partial_compliance_count,
    SUM(CASE WHEN vr.score < 70 THEN 1 ELSE 0 END) AS non_compliant_count,
    SUM(vr.issues_count) AS total_issues
FROM 
    validation_results vr
JOIN 
    compliance_validators cv ON vr.validator_id = cv.id
JOIN 
    regulatory_frameworks rf ON cv.framework_id = rf.id
GROUP BY 
    DATE_TRUNC('day', vr.validation_time),
    rf.id,
    rf.name,
    rf.region
ORDER BY 
    validation_date DESC, 
    rf.region;

COMMENT ON VIEW vw_compliance_trends IS 'Provides time-series data of compliance scores and issues for trend analysis';

-- =========================================================================
-- View: Regional Compliance Summary
-- Summarizes compliance status by region for geographic visualization
-- =========================================================================
CREATE OR REPLACE VIEW vw_regional_compliance_summary AS
WITH latest_validations AS (
    SELECT 
        vr.target_id,
        vr.target_type,
        cv.framework_id,
        rf.region,
        vr.score,
        vr.issues_count,
        ROW_NUMBER() OVER (PARTITION BY vr.target_id, vr.target_type, cv.framework_id ORDER BY vr.validation_time DESC) AS rn
    FROM 
        validation_results vr
    JOIN 
        compliance_validators cv ON vr.validator_id = cv.id
    JOIN 
        regulatory_frameworks rf ON cv.framework_id = rf.id
)
SELECT 
    region,
    COUNT(target_id) AS total_targets,
    AVG(score) AS avg_compliance_score,
    SUM(issues_count) AS total_issues,
    SUM(CASE WHEN score >= 90 THEN 1 ELSE 0 END) AS compliant_targets,
    SUM(CASE WHEN score >= 70 AND score < 90 THEN 1 ELSE 0 END) AS partial_compliance_targets,
    SUM(CASE WHEN score < 70 THEN 1 ELSE 0 END) AS non_compliant_targets,
    ROUND((SUM(CASE WHEN score >= 90 THEN 1 ELSE 0 END)::numeric / COUNT(target_id)::numeric) * 100, 2) AS compliance_percentage
FROM 
    latest_validations
WHERE 
    rn = 1
GROUP BY 
    region
ORDER BY 
    avg_compliance_score DESC;

COMMENT ON VIEW vw_regional_compliance_summary IS 'Summarizes compliance status by region for geographic dashboard visualization';

-- =========================================================================
-- View: Framework Compliance Details
-- Provides detailed compliance metrics by regulatory framework
-- =========================================================================
CREATE OR REPLACE VIEW vw_framework_compliance_details AS
WITH latest_validations AS (
    SELECT 
        vr.id,
        vr.target_id,
        vr.target_type,
        vr.score,
        vr.issues_count,
        vr.validation_time,
        cv.framework_id,
        ROW_NUMBER() OVER (PARTITION BY vr.target_id, vr.target_type, cv.framework_id ORDER BY vr.validation_time DESC) AS rn
    FROM 
        validation_results vr
    JOIN 
        compliance_validators cv ON vr.validator_id = cv.id
)
SELECT 
    rf.id AS framework_id,
    rf.name AS framework_name,
    rf.region,
    rf.sector,
    COUNT(lv.id) AS total_validations,
    AVG(lv.score) AS avg_compliance_score,
    SUM(lv.issues_count) AS total_issues,
    COUNT(DISTINCT lv.target_id) AS unique_targets,
    SUM(CASE WHEN lv.score >= 90 THEN 1 ELSE 0 END) AS compliant_count,
    SUM(CASE WHEN lv.score >= 70 AND lv.score < 90 THEN 1 ELSE 0 END) AS partial_compliance_count,
    SUM(CASE WHEN lv.score < 70 THEN 1 ELSE 0 END) AS non_compliant_count,
    ROUND((SUM(CASE WHEN lv.score >= 90 THEN 1 ELSE 0 END)::numeric / COUNT(lv.id)::numeric) * 100, 2) AS compliance_percentage,
    MAX(lv.validation_time) AS latest_validation
FROM 
    latest_validations lv
JOIN 
    regulatory_frameworks rf ON lv.framework_id = rf.id
WHERE 
    lv.rn = 1
GROUP BY 
    rf.id,
    rf.name,
    rf.region,
    rf.sector
ORDER BY 
    avg_compliance_score DESC;

COMMENT ON VIEW vw_framework_compliance_details IS 'Provides detailed compliance metrics by regulatory framework for executive reporting';

-- =========================================================================
-- View: Alert Priority Dashboard
-- Consolidates alerts with priority scoring for dashboard display
-- =========================================================================
CREATE OR REPLACE VIEW vw_alert_priority_dashboard AS
SELECT 
    ca.id AS alert_id,
    ca.target_id,
    ca.target_type,
    ca.alert_level,
    ca.message,
    ca.reference,
    ca.recommendation,
    ca.created_at,
    ca.status,
    ca.acknowledged_at,
    ca.acknowledged_by,
    cv.id AS validator_id,
    cv.name AS validator_name,
    rf.id AS framework_id,
    rf.name AS framework_name,
    rf.region,
    CASE 
        WHEN ca.alert_level = 'high' THEN 1
        WHEN ca.alert_level = 'medium' THEN 2
        WHEN ca.alert_level = 'low' THEN 3
        ELSE 4
    END AS priority_score,
    EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - ca.created_at))/86400 AS days_open
FROM 
    compliance_alerts ca
JOIN 
    compliance_validators cv ON ca.validator_id = cv.id
JOIN 
    regulatory_frameworks rf ON cv.framework_id = rf.id
WHERE 
    ca.status = 'ACTIVE'
ORDER BY 
    priority_score,
    ca.created_at;

COMMENT ON VIEW vw_alert_priority_dashboard IS 'Prioritized view of active compliance alerts for dashboard display';

-- =========================================================================
-- View: FHIR Validation Results
-- Specific view for FHIR validation results
-- =========================================================================
CREATE OR REPLACE VIEW vw_fhir_validation_results AS
SELECT 
    vr.id AS validation_id,
    vr.target_id,
    vr.target_type,
    vr.status,
    vr.score,
    vr.issues_count,
    vr.validation_time,
    vr.result_data->>'fhir_version' AS fhir_version,
    vr.result_data->>'resource_type' AS resource_type,
    vr.result_data->>'profile' AS profile,
    CASE 
        WHEN vr.score >= 90 THEN 'COMPLIANT'
        WHEN vr.score >= 70 THEN 'PARTIAL_COMPLIANCE'
        ELSE 'NON_COMPLIANT'
    END AS compliance_level
FROM 
    validation_results vr
JOIN 
    compliance_validators cv ON vr.validator_id = cv.id
WHERE 
    cv.id = 'FHIR-VALIDATOR-R4';

COMMENT ON VIEW vw_fhir_validation_results IS 'Specific view for FHIR validation results with resource type details';

-- =========================================================================
-- View: Regulatory Change Impact
-- Shows the potential impact of regulatory changes on compliance
-- =========================================================================
CREATE OR REPLACE VIEW vw_regulatory_change_impact AS
WITH recent_validations AS (
    SELECT 
        vr.target_id,
        vr.target_type,
        cv.framework_id,
        vr.score,
        ROW_NUMBER() OVER (PARTITION BY vr.target_id, vr.target_type, cv.framework_id ORDER BY vr.validation_time DESC) AS rn
    FROM 
        validation_results vr
    JOIN 
        compliance_validators cv ON vr.validator_id = cv.id
)
SELECT 
    rc.id AS change_id,
    rc.framework_id,
    rf.name AS framework_name,
    rf.region,
    rc.change_type,
    rc.change_summary,
    rc.change_date,
    rc.effective_date,
    rc.impact_level,
    rc.status AS change_status,
    COUNT(rv.target_id) AS potentially_affected_targets,
    AVG(rv.score) AS current_avg_compliance,
    SUM(CASE WHEN rv.score >= 90 THEN 1 ELSE 0 END) AS currently_compliant,
    SUM(CASE WHEN rv.score < 90 THEN 1 ELSE 0 END) AS potentially_impacted
FROM 
    regulatory_changes rc
JOIN 
    regulatory_frameworks rf ON rc.framework_id = rf.id
LEFT JOIN 
    recent_validations rv ON rc.framework_id = rv.framework_id AND rv.rn = 1
WHERE 
    rc.effective_date > CURRENT_DATE OR rc.effective_date IS NULL
GROUP BY 
    rc.id,
    rc.framework_id,
    rf.name,
    rf.region,
    rc.change_type,
    rc.change_summary,
    rc.change_date,
    rc.effective_date,
    rc.impact_level,
    rc.status
ORDER BY 
    CASE 
        WHEN rc.impact_level = 'HIGH' THEN 1
        WHEN rc.impact_level = 'MEDIUM' THEN 2
        WHEN rc.impact_level = 'LOW' THEN 3
        ELSE 4
    END,
    rc.effective_date;

COMMENT ON VIEW vw_regulatory_change_impact IS 'Assesses potential impact of upcoming regulatory changes on current compliance status';
