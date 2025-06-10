-- =========================================================================
-- INNOVABIZ - IAM - Healthcare Compliance Database Functions
-- =========================================================================
-- Autor: Eduardo Jeremias
-- Data: 08/05/2025
-- Versão: 1.0
--
-- Descrição: Funções e procedimentos para gerenciamento de dados de compliance
-- do módulo IAM Healthcare, incluindo funções para registro de validações,
-- atualização de status e geração de relatórios.
-- =========================================================================

SET search_path TO iam_healthcare, public;

-- =========================================================================
-- Function: Register Validation Result
-- Records a new validation result and associated issues
-- =========================================================================
CREATE OR REPLACE FUNCTION register_validation_result(
    p_validator_id VARCHAR(50),
    p_target_id VARCHAR(255),
    p_target_type VARCHAR(100),
    p_status VARCHAR(50),
    p_score NUMERIC(5, 2),
    p_issues_count INTEGER,
    p_result_data JSONB,
    p_issues JSONB
) RETURNS UUID AS $$
DECLARE
    v_result_id UUID;
    v_issue JSONB;
BEGIN
    -- Insert validation result
    INSERT INTO validation_results(
        validator_id,
        target_id,
        target_type,
        status,
        score,
        issues_count,
        result_data,
        issues
    ) VALUES (
        p_validator_id,
        p_target_id,
        p_target_type,
        p_status,
        p_score,
        p_issues_count,
        p_result_data,
        p_issues
    ) RETURNING id INTO v_result_id;
    
    -- Insert individual issues
    IF p_issues IS NOT NULL AND jsonb_array_length(p_issues) > 0 THEN
        FOR v_issue IN SELECT * FROM jsonb_array_elements(p_issues)
        LOOP
            INSERT INTO compliance_issues(
                validation_result_id,
                issue_id,
                severity,
                description,
                recommendation,
                reference
            ) VALUES (
                v_result_id,
                v_issue->>'id',
                v_issue->>'severity',
                v_issue->>'description',
                v_issue->>'recommendation',
                v_issue->>'reference'
            );
        END LOOP;
    END IF;
    
    -- Create alerts for critical and high severity issues
    INSERT INTO compliance_alerts(
        validator_id,
        target_id,
        target_type,
        alert_level,
        message,
        reference,
        recommendation
    )
    SELECT
        p_validator_id,
        p_target_id,
        p_target_type,
        CASE WHEN severity = 'critical' THEN 'high' ELSE severity END,
        description,
        reference,
        recommendation
    FROM compliance_issues
    WHERE validation_result_id = v_result_id
    AND severity IN ('critical', 'high');
    
    RETURN v_result_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION register_validation_result IS 'Records a new validation result and associated issues, generating alerts for high severity issues';

-- =========================================================================
-- Function: Get Compliance Summary
-- Returns a summary of compliance status for a specific target
-- =========================================================================
CREATE OR REPLACE FUNCTION get_compliance_summary(
    p_target_id VARCHAR(255),
    p_target_type VARCHAR(100)
) RETURNS TABLE (
    framework_id VARCHAR(50),
    framework_name VARCHAR(100),
    region VARCHAR(5),
    validator_id VARCHAR(50),
    validator_name VARCHAR(100),
    status VARCHAR(50),
    score NUMERIC(5, 2),
    issues_count INTEGER,
    validation_time TIMESTAMP WITH TIME ZONE,
    compliance_level VARCHAR(20)
) AS $$
BEGIN
    RETURN QUERY
    WITH latest_validations AS (
        SELECT 
            vr.validator_id,
            ROW_NUMBER() OVER (PARTITION BY cv.framework_id ORDER BY vr.validation_time DESC) AS rn,
            vr.id
        FROM 
            validation_results vr
        JOIN 
            compliance_validators cv ON vr.validator_id = cv.id
        WHERE 
            vr.target_id = p_target_id
        AND 
            vr.target_type = p_target_type
    )
    SELECT 
        rf.id AS framework_id,
        rf.name AS framework_name,
        rf.region,
        cv.id AS validator_id,
        cv.name AS validator_name,
        vr.status,
        vr.score,
        vr.issues_count,
        vr.validation_time,
        CASE 
            WHEN vr.score >= 90 THEN 'COMPLIANT'
            WHEN vr.score >= 70 THEN 'PARTIAL_COMPLIANCE'
            ELSE 'NON_COMPLIANT'
        END AS compliance_level
    FROM 
        latest_validations lv
    JOIN 
        validation_results vr ON lv.id = vr.id
    JOIN 
        compliance_validators cv ON vr.validator_id = cv.id
    JOIN 
        regulatory_frameworks rf ON cv.framework_id = rf.id
    WHERE 
        lv.rn = 1
    ORDER BY 
        vr.score DESC;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_compliance_summary IS 'Retrieves the latest compliance status for each regulatory framework for a specific target';

-- =========================================================================
-- Function: Resolve Compliance Issue
-- Marks a compliance issue as resolved
-- =========================================================================
CREATE OR REPLACE FUNCTION resolve_compliance_issue(
    p_issue_id UUID,
    p_resolution_notes TEXT,
    p_resolved_by VARCHAR(100)
) RETURNS BOOLEAN AS $$
DECLARE
    v_target_id VARCHAR(255);
    v_target_type VARCHAR(100);
    v_validator_id VARCHAR(50);
BEGIN
    -- Get issue details
    SELECT 
        vr.target_id, 
        vr.target_type,
        vr.validator_id
    INTO 
        v_target_id, 
        v_target_type,
        v_validator_id
    FROM 
        compliance_issues ci
    JOIN 
        validation_results vr ON ci.validation_result_id = vr.id
    WHERE 
        ci.id = p_issue_id;
    
    IF NOT FOUND THEN
        RETURN FALSE;
    END IF;
    
    -- Update issue status
    UPDATE compliance_issues
    SET 
        status = 'RESOLVED',
        resolved_at = CURRENT_TIMESTAMP,
        resolution_notes = p_resolution_notes
    WHERE 
        id = p_issue_id;
    
    -- Update related alerts
    UPDATE compliance_alerts
    SET 
        status = 'RESOLVED',
        acknowledged_at = CURRENT_TIMESTAMP,
        acknowledged_by = p_resolved_by
    WHERE 
        target_id = v_target_id
    AND 
        target_type = v_target_type
    AND 
        validator_id = v_validator_id
    AND 
        message LIKE (SELECT '%' || description || '%' FROM compliance_issues WHERE id = p_issue_id);
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION resolve_compliance_issue IS 'Marks a compliance issue as resolved and updates related alerts';

-- =========================================================================
-- Function: Generate Compliance Report
-- Creates a new compliance report for specified frameworks and targets
-- =========================================================================
CREATE OR REPLACE FUNCTION generate_compliance_report(
    p_report_type VARCHAR(100),
    p_framework_ids VARCHAR(50)[],
    p_report_format VARCHAR(20),
    p_generated_by VARCHAR(100)
) RETURNS UUID AS $$
DECLARE
    v_report_id UUID;
    v_report_data JSONB;
    v_summary TEXT;
    v_formatted_date TEXT;
BEGIN
    -- Generate report ID
    SELECT TO_CHAR(CURRENT_TIMESTAMP, 'YYYYMMDD-HH24MISS') INTO v_formatted_date;
    
    -- Build report data
    WITH latest_validations AS (
        SELECT 
            vr.id,
            vr.target_id,
            vr.target_type,
            cv.framework_id,
            vr.score,
            vr.status,
            vr.issues_count,
            vr.validation_time,
            ROW_NUMBER() OVER (PARTITION BY vr.target_id, vr.target_type, cv.framework_id ORDER BY vr.validation_time DESC) AS rn
        FROM 
            validation_results vr
        JOIN 
            compliance_validators cv ON vr.validator_id = cv.id
        WHERE 
            cv.framework_id = ANY(p_framework_ids)
    ),
    framework_stats AS (
        SELECT 
            framework_id,
            COUNT(DISTINCT target_id) AS targets_count,
            AVG(score) AS avg_score,
            SUM(issues_count) AS total_issues,
            COUNT(CASE WHEN score >= 90 THEN 1 END) AS compliant_count,
            COUNT(CASE WHEN score >= 70 AND score < 90 THEN 1 END) AS partial_count,
            COUNT(CASE WHEN score < 70 THEN 1 END) AS non_compliant_count
        FROM 
            latest_validations
        WHERE 
            rn = 1
        GROUP BY 
            framework_id
    ),
    issues_summary AS (
        SELECT 
            ci.severity,
            COUNT(*) AS count
        FROM 
            compliance_issues ci
        JOIN 
            latest_validations lv ON ci.validation_result_id = lv.id
        WHERE 
            lv.rn = 1
        AND 
            ci.status = 'OPEN'
        GROUP BY 
            ci.severity
    )
    SELECT 
        jsonb_build_object(
            'report_type', p_report_type,
            'generated_at', CURRENT_TIMESTAMP,
            'generated_by', p_generated_by,
            'frameworks', (
                SELECT jsonb_agg(
                    jsonb_build_object(
                        'id', rf.id,
                        'name', rf.name,
                        'region', rf.region,
                        'targets_count', fs.targets_count,
                        'avg_score', fs.avg_score,
                        'total_issues', fs.total_issues,
                        'compliant_count', fs.compliant_count,
                        'partial_count', fs.partial_count,
                        'non_compliant_count', fs.non_compliant_count
                    )
                )
                FROM framework_stats fs
                JOIN regulatory_frameworks rf ON fs.framework_id = rf.id
            ),
            'issues_summary', (
                SELECT jsonb_object_agg(
                    severity, count
                )
                FROM issues_summary
            ),
            'overall_compliance', (
                SELECT 
                    CASE 
                        WHEN AVG(fs.avg_score) >= 90 THEN 'COMPLIANT'
                        WHEN AVG(fs.avg_score) >= 70 THEN 'PARTIAL_COMPLIANCE'
                        ELSE 'NON_COMPLIANT'
                    END
                FROM 
                    framework_stats fs
            )
        ) INTO v_report_data;
    
    -- Generate summary
    SELECT 
        'Compliance Report: ' || 
        p_report_type || 
        ' - Overall Compliance: ' || 
        (v_report_data->>'overall_compliance') ||
        ' - Generated on: ' || 
        TO_CHAR(CURRENT_TIMESTAMP, 'YYYY-MM-DD HH24:MI:SS')
    INTO v_summary;
    
    -- Insert report
    INSERT INTO compliance_reports(
        report_id,
        report_type,
        generated_at,
        generated_by,
        framework_ids,
        report_format,
        report_data,
        summary
    ) VALUES (
        p_report_type || '-' || v_formatted_date,
        p_report_type,
        CURRENT_TIMESTAMP,
        p_generated_by,
        p_framework_ids,
        p_report_format,
        v_report_data,
        v_summary
    ) RETURNING id INTO v_report_id;
    
    RETURN v_report_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION generate_compliance_report IS 'Creates a new compliance report for specified regulatory frameworks';

-- =========================================================================
-- Function: Get Regional Compliance
-- Returns compliance metrics by region for geospatial visualization
-- =========================================================================
CREATE OR REPLACE FUNCTION get_regional_compliance() 
RETURNS TABLE (
    region VARCHAR(5),
    region_name TEXT,
    compliance_score NUMERIC,
    compliance_level TEXT,
    total_targets INTEGER,
    compliant_targets INTEGER,
    non_compliant_targets INTEGER,
    total_issues INTEGER,
    critical_issues INTEGER,
    high_issues INTEGER
) AS $$
BEGIN
    RETURN QUERY
    WITH region_mapping AS (
        SELECT 'US' AS region, 'United States' AS region_name
        UNION SELECT 'EU', 'European Union'
        UNION SELECT 'BR', 'Brazil'
        UNION SELECT 'AO', 'Angola'
        UNION SELECT 'GLOBAL', 'Global'
    ),
    latest_validations AS (
        SELECT 
            vr.target_id,
            vr.target_type,
            rf.region,
            vr.score,
            vr.issues_count,
            vr.id AS validation_id,
            ROW_NUMBER() OVER (PARTITION BY vr.target_id, vr.target_type, rf.region ORDER BY vr.validation_time DESC) AS rn
        FROM 
            validation_results vr
        JOIN 
            compliance_validators cv ON vr.validator_id = cv.id
        JOIN 
            regulatory_frameworks rf ON cv.framework_id = rf.id
    ),
    issues_by_region AS (
        SELECT 
            rf.region,
            ci.severity,
            COUNT(*) AS issue_count
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
        GROUP BY 
            rf.region, ci.severity
    )
    SELECT 
        rm.region,
        rm.region_name,
        COALESCE(AVG(lv.score), 0) AS compliance_score,
        CASE 
            WHEN AVG(lv.score) >= 90 THEN 'COMPLIANT'
            WHEN AVG(lv.score) >= 70 THEN 'PARTIAL_COMPLIANCE'
            ELSE 'NON_COMPLIANT'
        END AS compliance_level,
        COUNT(DISTINCT lv.target_id) AS total_targets,
        COUNT(DISTINCT CASE WHEN lv.score >= 90 THEN lv.target_id END) AS compliant_targets,
        COUNT(DISTINCT CASE WHEN lv.score < 70 THEN lv.target_id END) AS non_compliant_targets,
        COALESCE(SUM(lv.issues_count), 0) AS total_issues,
        COALESCE(SUM(CASE WHEN ibr.severity = 'critical' THEN ibr.issue_count ELSE 0 END), 0) AS critical_issues,
        COALESCE(SUM(CASE WHEN ibr.severity = 'high' THEN ibr.issue_count ELSE 0 END), 0) AS high_issues
    FROM 
        region_mapping rm
    LEFT JOIN 
        latest_validations lv ON rm.region = lv.region AND lv.rn = 1
    LEFT JOIN 
        issues_by_region ibr ON rm.region = ibr.region
    GROUP BY 
        rm.region, rm.region_name
    ORDER BY 
        compliance_score DESC;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_regional_compliance IS 'Returns compliance metrics by region for geospatial dashboard visualization';

-- =========================================================================
-- Function: Track Regulatory Change
-- Records a new regulatory change and assesses potential impact
-- =========================================================================
CREATE OR REPLACE FUNCTION track_regulatory_change(
    p_framework_id VARCHAR(50),
    p_change_type VARCHAR(50),
    p_change_summary TEXT,
    p_change_details TEXT,
    p_effective_date TIMESTAMP WITH TIME ZONE,
    p_source_url TEXT,
    p_impact_level VARCHAR(20),
    p_affected_components JSONB
) RETURNS UUID AS $$
DECLARE
    v_change_id UUID;
BEGIN
    -- Insert regulatory change
    INSERT INTO regulatory_changes(
        framework_id,
        change_type,
        change_summary,
        change_details,
        effective_date,
        source_url,
        impact_level,
        affected_components
    ) VALUES (
        p_framework_id,
        p_change_type,
        p_change_summary,
        p_change_details,
        p_effective_date,
        p_source_url,
        p_impact_level,
        p_affected_components
    ) RETURNING id INTO v_change_id;
    
    -- Create alerts for high impact changes
    IF p_impact_level = 'HIGH' THEN
        INSERT INTO compliance_alerts(
            validator_id,
            target_id,
            target_type,
            alert_level,
            message,
            reference,
            recommendation
        )
        SELECT 
            cv.id,
            'SYSTEM',
            'REGULATORY_CHANGE',
            'high',
            'High impact regulatory change: ' || p_change_summary,
            p_source_url,
            'Review ' || rf.name || ' compliance before effective date: ' || 
            TO_CHAR(p_effective_date, 'YYYY-MM-DD')
        FROM 
            compliance_validators cv
        JOIN 
            regulatory_frameworks rf ON cv.framework_id = rf.id
        WHERE 
            cv.framework_id = p_framework_id;
    END IF;
    
    RETURN v_change_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION track_regulatory_change IS 'Records a regulatory change and generates alerts for high impact changes';

-- =========================================================================
-- Function: Get Compliance Timeline
-- Returns compliance score changes over time for a target
-- =========================================================================
CREATE OR REPLACE FUNCTION get_compliance_timeline(
    p_target_id VARCHAR(255),
    p_target_type VARCHAR(100),
    p_days INTEGER DEFAULT 90
) RETURNS TABLE (
    validation_date TIMESTAMP WITH TIME ZONE,
    framework_id VARCHAR(50),
    framework_name VARCHAR(100),
    validator_id VARCHAR(50),
    score NUMERIC(5, 2),
    issues_count INTEGER,
    status VARCHAR(50)
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        vr.validation_time AS validation_date,
        rf.id AS framework_id,
        rf.name AS framework_name,
        cv.id AS validator_id,
        vr.score,
        vr.issues_count,
        vr.status
    FROM 
        validation_results vr
    JOIN 
        compliance_validators cv ON vr.validator_id = cv.id
    JOIN 
        regulatory_frameworks rf ON cv.framework_id = rf.id
    WHERE 
        vr.target_id = p_target_id
    AND 
        vr.target_type = p_target_type
    AND 
        vr.validation_time >= (CURRENT_TIMESTAMP - (p_days || ' days')::INTERVAL)
    ORDER BY 
        rf.id, vr.validation_time;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_compliance_timeline IS 'Returns compliance score changes over time for historical analysis';
