-- ===============================================================================
-- IAM LGPD Compliance Validators (Part 2)
-- INNOVABIZ Platform
-- Version: 1.0
-- Date: January 2025
-- Description: Implementation of LGPD data subject rights validators
-- ===============================================================================

-- Schema reference for IAM LGPD validators
SET search_path TO iam_compliance, iam, public;

-- =======================================================
-- LGPD Data Subject Rights Validator
-- =======================================================

/**
 * Validates implementation of data subject rights
 * LGPD Article 18, 19, 20
 */
CREATE OR REPLACE FUNCTION validate_lgpd_data_subject_rights(
    p_tenant_id UUID,
    p_parameters JSONB DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_result JSONB;
    v_validator_id UUID;
    v_issues JSONB[] := '{}';
    v_check_id UUID;
    v_summary TEXT;
    v_status TEXT;
    v_points NUMERIC := 0;
    v_max_points NUMERIC := 10;
    v_percentage NUMERIC;
    
    -- Rights implementation check
    v_rights_implemented JSONB;
    v_missing_rights TEXT[];
    v_incomplete_rights TEXT[];
    
    -- Request handling metrics
    v_request_count INTEGER;
    v_overdue_count INTEGER;
    v_rejected_count INTEGER;
    v_request_types TEXT[];
    v_avg_completion_time INTERVAL;
    v_healthcare_specific_count INTEGER;
BEGIN
    -- Get validator ID
    SELECT validator_id INTO v_validator_id 
    FROM compliance_validators 
    WHERE validator_name = 'LGPD-Data-Subject-Rights-Validator' 
    AND regulatory_framework = 'LGPD';
    
    -- Create compliance check record
    v_check_id := gen_random_uuid();
    
    INSERT INTO compliance_checks (
        check_id,
        validator_id,
        tenant_id,
        execution_date,
        parameters,
        status
    ) VALUES (
        v_check_id,
        v_validator_id,
        p_tenant_id,
        NOW(),
        p_parameters,
        'in_progress'
    );

    -- 1. Check if all LGPD data subject rights are implemented
    SELECT setting_value::JSONB INTO v_rights_implemented
    FROM iam.tenant_settings
    WHERE tenant_id = p_tenant_id
    AND setting_key = 'data_subject_rights_implementation';
    
    -- If setting doesn't exist, create an issue
    IF v_rights_implemented IS NULL THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'issue_id', gen_random_uuid(),
            'issue_type', 'missing_rights_implementation',
            'description', 'No data subject rights implementation found',
            'severity', 'critical',
            'remediation', 'Implement all data subject rights required by LGPD Article 18'
        ));
    ELSE
        -- Check which rights are implemented
        WITH required_rights(right_name) AS (
            VALUES 
                ('confirmation_of_processing'),
                ('access_to_data'),
                ('correction_of_data'),
                ('anonymization'),
                ('portability'),
                ('deletion_of_data'),
                ('information_about_sharing'),
                ('information_about_consent'),
                ('revocation_of_consent')
        ),
        implemented_rights AS (
            SELECT 
                right_name,
                v_rights_implemented->>right_name AS status
            FROM required_rights
        )
        SELECT 
            ARRAY_AGG(right_name) FILTER (WHERE status IS NULL),
            ARRAY_AGG(right_name) FILTER (WHERE status = 'partial')
        INTO
            v_missing_rights,
            v_incomplete_rights
        FROM implemented_rights;
        
        -- Add points based on implementation
        IF array_length(v_missing_rights, 1) IS NULL THEN
            v_points := v_points + 5;
        ELSIF array_length(v_missing_rights, 1) <= 2 THEN
            v_points := v_points + 3;
        ELSIF array_length(v_missing_rights, 1) <= 4 THEN
            v_points := v_points + 1;
        END IF;
        
        IF array_length(v_incomplete_rights, 1) IS NULL THEN
            v_points := v_points + 2;
        ELSIF array_length(v_incomplete_rights, 1) <= 3 THEN
            v_points := v_points + 1;
        END IF;
        
        -- Add issues for missing or incomplete rights
        IF array_length(v_missing_rights, 1) > 0 THEN
            v_issues := array_append(v_issues, jsonb_build_object(
                'issue_id', gen_random_uuid(),
                'issue_type', 'missing_rights',
                'description', 'Required data subject rights not implemented',
                'affected_items', v_missing_rights,
                'severity', 'critical',
                'remediation', 'Implement all missing data subject rights required by LGPD Article 18'
            ));
        END IF;
        
        IF array_length(v_incomplete_rights, 1) > 0 THEN
            v_issues := array_append(v_issues, jsonb_build_object(
                'issue_id', gen_random_uuid(),
                'issue_type', 'incomplete_rights',
                'description', 'Data subject rights partially implemented',
                'affected_items', v_incomplete_rights,
                'severity', 'high',
                'remediation', 'Complete the implementation of partially implemented rights'
            ));
        END IF;
    END IF;
    
    -- 2. Check data subject request handling
    SELECT
        COUNT(*),
        ARRAY_AGG(DISTINCT request_type),
        COUNT(*) FILTER (WHERE status = 'overdue'),
        COUNT(*) FILTER (WHERE status = 'rejected'),
        AVG(completion_date - request_date) FILTER (WHERE completion_date IS NOT NULL),
        COUNT(*) FILTER (WHERE metadata->>'sector' = 'healthcare')
    INTO
        v_request_count,
        v_request_types,
        v_overdue_count,
        v_rejected_count,
        v_avg_completion_time,
        v_healthcare_specific_count
    FROM iam.data_subject_requests
    WHERE tenant_id = p_tenant_id
    AND request_date > NOW() - INTERVAL '12 months';
    
    -- Add points based on request handling
    IF v_request_count > 0 THEN
        -- Add points if requests are handled in a timely manner
        IF v_overdue_count::FLOAT / v_request_count::FLOAT < 0.05 THEN
            v_points := v_points + 1;
        END IF;
        
        -- Add points if average response time is reasonable
        IF v_avg_completion_time < INTERVAL '15 days' THEN
            v_points := v_points + 1;
        END IF;
        
        -- Add points for diverse request types handling
        IF array_length(v_request_types, 1) >= 5 THEN
            v_points := v_points + 1;
        END IF;
    ELSE
        -- If no requests exist but rights are implemented, still give some points
        IF v_rights_implemented IS NOT NULL THEN
            v_points := v_points + 1;
        END IF;
    END IF;
    
    -- 3. Check for healthcare-specific data subject rights
    -- Special checking for healthcare data access rights
    IF v_healthcare_specific_count > 0 OR EXISTS (
        SELECT 1 FROM iam.tenant_settings
        WHERE tenant_id = p_tenant_id
        AND setting_key = 'healthcare_data_subject_rights'
        AND setting_value::JSONB->>'enabled' = 'true'
    ) THEN
        v_points := v_points + 1;
    END IF;
    
    -- Add issues based on request handling
    IF v_request_count > 0 AND v_overdue_count > 0 THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'issue_id', gen_random_uuid(),
            'issue_type', 'overdue_requests',
            'description', 'Data subject requests not handled within required timeframe',
            'count', v_overdue_count,
            'severity', 'high',
            'remediation', 'Implement process to ensure requests are handled within 15 days per LGPD Article 19'
        ));
    END IF;
    
    -- Calculate compliance percentage
    v_percentage := (v_points / v_max_points * 100)::NUMERIC(5,2);
    
    -- Determine status
    IF v_percentage >= 90 THEN
        v_status := 'compliant';
    ELSIF v_percentage >= 70 THEN
        v_status := 'partially_compliant';
    ELSE
        v_status := 'non_compliant';
    END IF;
    
    -- Create summary
    v_summary := format(
        'Data subject rights validation: %s%% compliant. %s of %s points earned. %s data subject requests analyzed.',
        v_percentage,
        v_points,
        v_max_points,
        v_request_count
    );
    
    -- Update compliance check record
    UPDATE compliance_checks
    SET 
        status = v_status,
        summary = v_summary,
        compliance_score = v_percentage,
        issues = to_jsonb(v_issues)
    WHERE check_id = v_check_id;
    
    -- Build result
    v_result := jsonb_build_object(
        'check_id', v_check_id,
        'validator', 'LGPD-Data-Subject-Rights-Validator',
        'status', v_status,
        'summary', v_summary,
        'score', v_percentage,
        'issues', v_issues,
        'execution_date', NOW()
    );
    
    RETURN v_result;
END;
$$ LANGUAGE plpgsql;

-- =======================================================
-- LGPD Sensitive Data Validator
-- =======================================================

/**
 * Validates enhanced protection for sensitive personal data
 * LGPD Article 11, 12, 13
 */
CREATE OR REPLACE FUNCTION validate_lgpd_sensitive_data(
    p_tenant_id UUID,
    p_parameters JSONB DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_result JSONB;
    v_validator_id UUID;
    v_issues JSONB[] := '{}';
    v_check_id UUID;
    v_summary TEXT;
    v_status TEXT;
    v_points NUMERIC := 0;
    v_max_points NUMERIC := 9;
    v_percentage NUMERIC;
    
    -- Sensitive data metrics
    v_sensitive_data_count INTEGER;
    v_sensitive_categories TEXT[];
    v_unencrypted_count INTEGER;
    v_unencrypted_categories TEXT[];
    v_no_legal_basis_count INTEGER;
    v_restricted_access_count INTEGER;
    v_healthcare_data_count INTEGER;
    v_healthcare_compliance BOOLEAN;
BEGIN
    -- Get validator ID
    SELECT validator_id INTO v_validator_id 
    FROM compliance_validators 
    WHERE validator_name = 'LGPD-Sensitive-Data-Validator' 
    AND regulatory_framework = 'LGPD';
    
    -- Create compliance check record
    v_check_id := gen_random_uuid();
    
    INSERT INTO compliance_checks (
        check_id,
        validator_id,
        tenant_id,
        execution_date,
        parameters,
        status
    ) VALUES (
        v_check_id,
        v_validator_id,
        p_tenant_id,
        NOW(),
        p_parameters,
        'in_progress'
    );

    -- 1. Check sensitive data inventory
    SELECT
        COUNT(*),
        ARRAY_AGG(DISTINCT data_category),
        COUNT(*) FILTER (WHERE NOT is_encrypted),
        ARRAY_AGG(DISTINCT data_category) FILTER (WHERE NOT is_encrypted),
        COUNT(*) FILTER (WHERE legal_basis IS NULL OR legal_basis = ''),
        COUNT(*) FILTER (WHERE has_restricted_access),
        COUNT(*) FILTER (WHERE data_category LIKE '%health%' OR data_category LIKE '%medical%')
    INTO
        v_sensitive_data_count,
        v_sensitive_categories,
        v_unencrypted_count,
        v_unencrypted_categories,
        v_no_legal_basis_count,
        v_restricted_access_count,
        v_healthcare_data_count
    FROM iam.data_processing
    WHERE tenant_id = p_tenant_id
    AND is_sensitive = TRUE;
    
    -- Check for healthcare-specific compliance
    SELECT setting_value::BOOLEAN INTO v_healthcare_compliance
    FROM iam.tenant_settings
    WHERE tenant_id = p_tenant_id
    AND setting_key = 'healthcare_sensitive_data_compliance';
    
    -- Add points based on sensitive data protection
    -- Point 1: Having a sensitive data inventory
    IF v_sensitive_data_count > 0 THEN
        v_points := v_points + 1;
    END IF;
    
    -- Point 2: Encryption of sensitive data
    IF v_sensitive_data_count > 0 AND (v_unencrypted_count::FLOAT / v_sensitive_data_count::FLOAT) < 0.05 THEN
        v_points := v_points + 2;
    ELSIF v_sensitive_data_count > 0 AND (v_unencrypted_count::FLOAT / v_sensitive_data_count::FLOAT) < 0.15 THEN
        v_points := v_points + 1;
    END IF;
    
    -- Point 3: Legal basis for sensitive data
    IF v_sensitive_data_count > 0 AND (v_no_legal_basis_count::FLOAT / v_sensitive_data_count::FLOAT) < 0.02 THEN
        v_points := v_points + 2;
    ELSIF v_sensitive_data_count > 0 AND (v_no_legal_basis_count::FLOAT / v_sensitive_data_count::FLOAT) < 0.10 THEN
        v_points := v_points + 1;
    END IF;
    
    -- Point 4: Access restriction for sensitive data
    IF v_sensitive_data_count > 0 AND (v_restricted_access_count::FLOAT / v_sensitive_data_count::FLOAT) > 0.95 THEN
        v_points := v_points + 2;
    ELSIF v_sensitive_data_count > 0 AND (v_restricted_access_count::FLOAT / v_sensitive_data_count::FLOAT) > 0.85 THEN
        v_points := v_points + 1;
    END IF;
    
    -- Point 5: Healthcare-specific sensitive data handling
    IF v_healthcare_data_count > 0 AND v_healthcare_compliance = TRUE THEN
        v_points := v_points + 2;
    ELSIF v_healthcare_data_count > 0 THEN
        -- If we have healthcare data but no specific compliance setting
        v_points := v_points + 0;
        
        v_issues := array_append(v_issues, jsonb_build_object(
            'issue_id', gen_random_uuid(),
            'issue_type', 'healthcare_data_non_compliance',
            'description', 'Healthcare sensitive data without proper compliance measures',
            'count', v_healthcare_data_count,
            'severity', 'critical',
            'remediation', 'Implement healthcare-specific sensitive data protection controls'
        ));
    ELSIF v_healthcare_data_count = 0 THEN
        -- If we don't have healthcare data, we don't need specific compliance
        v_points := v_points + 2;
    END IF;
    
    -- Add issues based on sensitive data handling
    IF v_unencrypted_count > 0 THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'issue_id', gen_random_uuid(),
            'issue_type', 'unencrypted_sensitive_data',
            'description', 'Sensitive data stored without encryption',
            'affected_items', v_unencrypted_categories,
            'count', v_unencrypted_count,
            'severity', 'critical',
            'remediation', 'Implement encryption for all sensitive data categories'
        ));
    END IF;
    
    IF v_no_legal_basis_count > 0 THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'issue_id', gen_random_uuid(),
            'issue_type', 'sensitive_data_without_legal_basis',
            'description', 'Processing of sensitive data without valid legal basis',
            'count', v_no_legal_basis_count,
            'severity', 'critical',
            'remediation', 'Define and document a valid legal basis for each sensitive data processing operation per LGPD Article 11'
        ));
    END IF;
    
    IF v_sensitive_data_count > 0 AND v_restricted_access_count < v_sensitive_data_count THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'issue_id', gen_random_uuid(),
            'issue_type', 'unrestricted_sensitive_data_access',
            'description', 'Sensitive data without access restrictions',
            'count', v_sensitive_data_count - v_restricted_access_count,
            'severity', 'high',
            'remediation', 'Implement access controls and restrictions for all sensitive data'
        ));
    END IF;
    
    -- Calculate compliance percentage
    v_percentage := (v_points / v_max_points * 100)::NUMERIC(5,2);
    
    -- Determine status
    IF v_percentage >= 90 THEN
        v_status := 'compliant';
    ELSIF v_percentage >= 75 THEN
        v_status := 'partially_compliant';
    ELSE
        v_status := 'non_compliant';
    END IF;
    
    -- Create summary
    v_summary := format(
        'Sensitive data protection validation: %s%% compliant. %s of %s criteria met. %s sensitive data categories processed.',
        v_percentage,
        v_points,
        v_max_points,
        array_length(v_sensitive_categories, 1)
    );
    
    -- Update compliance check record
    UPDATE compliance_checks
    SET 
        status = v_status,
        summary = v_summary,
        compliance_score = v_percentage,
        issues = to_jsonb(v_issues)
    WHERE check_id = v_check_id;
    
    -- Build result
    v_result := jsonb_build_object(
        'check_id', v_check_id,
        'validator', 'LGPD-Sensitive-Data-Validator',
        'status', v_status,
        'summary', v_summary,
        'score', v_percentage,
        'issues', v_issues,
        'execution_date', NOW()
    );
    
    RETURN v_result;
END;
$$ LANGUAGE plpgsql;
