-- ===============================================================================
-- IAM LGPD Compliance Validators (Part 3)
-- INNOVABIZ Platform
-- Version: 1.0
-- Date: January 2025
-- Description: Implementation of LGPD data breach and security validators
-- ===============================================================================

-- Schema reference for IAM LGPD validators
SET search_path TO iam_compliance, iam, public;

-- =======================================================
-- LGPD Data Breach Notification Validator
-- =======================================================

/**
 * Validates data breach notification processes
 * LGPD Article 48
 */
CREATE OR REPLACE FUNCTION validate_lgpd_data_breach_notification(
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
    v_max_points NUMERIC := 7;
    v_percentage NUMERIC;
    
    -- Breach notification metrics
    v_breach_count INTEGER;
    v_timely_notification_count INTEGER;
    v_complete_notification_count INTEGER;
    v_without_notification_count INTEGER;
    v_avg_notification_time INTERVAL;
    v_plan_exists BOOLEAN;
    v_team_assigned BOOLEAN;
    v_notification_template_exists BOOLEAN;
    v_healthcare_specific_plan BOOLEAN;
BEGIN
    -- Get validator ID
    SELECT validator_id INTO v_validator_id 
    FROM compliance_validators 
    WHERE validator_name = 'LGPD-Data-Breach-Notification-Validator' 
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

    -- 1. Check breach notification plan existence
    SELECT 
        COALESCE(setting_value::JSONB->>'plan_exists', 'false')::BOOLEAN,
        COALESCE(setting_value::JSONB->>'team_assigned', 'false')::BOOLEAN,
        COALESCE(setting_value::JSONB->>'notification_template_exists', 'false')::BOOLEAN,
        COALESCE(setting_value::JSONB->>'healthcare_specific_plan', 'false')::BOOLEAN
    INTO 
        v_plan_exists,
        v_team_assigned,
        v_notification_template_exists,
        v_healthcare_specific_plan
    FROM iam.tenant_settings
    WHERE tenant_id = p_tenant_id
    AND setting_key = 'data_breach_notification_plan';
    
    -- Add points for breach notification plan
    IF v_plan_exists THEN
        v_points := v_points + 2;
    END IF;
    
    IF v_team_assigned THEN
        v_points := v_points + 1;
    END IF;
    
    IF v_notification_template_exists THEN
        v_points := v_points + 1;
    END IF;
    
    -- Add additional point for healthcare-specific breach plan
    IF v_healthcare_specific_plan THEN
        v_points := v_points + 1;
    END IF;
    
    -- 2. Check breach notification implementation
    SELECT
        COUNT(*),
        COUNT(*) FILTER (WHERE notification_date - detection_date <= INTERVAL '2 days'),
        COUNT(*) FILTER (WHERE is_notification_complete),
        COUNT(*) FILTER (WHERE notification_date IS NULL),
        AVG(notification_date - detection_date) FILTER (WHERE notification_date IS NOT NULL)
    INTO
        v_breach_count,
        v_timely_notification_count,
        v_complete_notification_count,
        v_without_notification_count,
        v_avg_notification_time
    FROM iam.data_breaches
    WHERE tenant_id = p_tenant_id
    AND detection_date > NOW() - INTERVAL '24 months';
    
    -- Add points based on breach notification implementation
    IF v_breach_count > 0 THEN
        -- Add points for timely notification
        IF (v_timely_notification_count::FLOAT / v_breach_count::FLOAT) > 0.95 THEN
            v_points := v_points + 1;
        END IF;
        
        -- Add points for complete notification
        IF (v_complete_notification_count::FLOAT / v_breach_count::FLOAT) > 0.95 THEN
            v_points := v_points + 1;
        END IF;
    ELSE
        -- If no breaches reported, award points for having a plan in place
        IF v_plan_exists THEN
            v_points := v_points + 2;
        END IF;
    END IF;
    
    -- Add issues based on findings
    IF NOT v_plan_exists THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'issue_id', gen_random_uuid(),
            'issue_type', 'missing_breach_notification_plan',
            'description', 'No data breach notification plan documented',
            'severity', 'critical',
            'remediation', 'Develop and document a data breach notification plan in compliance with LGPD Article 48'
        ));
    END IF;
    
    IF NOT v_team_assigned THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'issue_id', gen_random_uuid(),
            'issue_type', 'missing_breach_response_team',
            'description', 'No data breach response team assigned',
            'severity', 'high',
            'remediation', 'Assign and train a dedicated team responsible for data breach response'
        ));
    END IF;
    
    IF v_breach_count > 0 AND v_without_notification_count > 0 THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'issue_id', gen_random_uuid(),
            'issue_type', 'breaches_without_notification',
            'description', 'Data breaches without required notification',
            'count', v_without_notification_count,
            'severity', 'critical',
            'remediation', 'Notify data protection authority about all data breaches in a reasonable timeframe'
        ));
    END IF;
    
    IF v_breach_count > 0 AND v_breach_count - v_complete_notification_count > 0 THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'issue_id', gen_random_uuid(),
            'issue_type', 'incomplete_breach_notifications',
            'description', 'Data breach notifications with incomplete information',
            'count', v_breach_count - v_complete_notification_count,
            'severity', 'high',
            'remediation', 'Ensure all breach notifications include complete information required by LGPD'
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
    IF v_breach_count > 0 THEN
        v_summary := format(
            'Data breach notification validation: %s%% compliant. %s of %s criteria met. %s data breaches analyzed with avg notification time of %s.',
            v_percentage,
            v_points,
            v_max_points,
            v_breach_count,
            v_avg_notification_time
        );
    ELSE
        v_summary := format(
            'Data breach notification validation: %s%% compliant. %s of %s criteria met. No data breaches reported in the last 24 months.',
            v_percentage,
            v_points,
            v_max_points
        );
    END IF;
    
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
        'validator', 'LGPD-Data-Breach-Notification-Validator',
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
-- LGPD Security Measures Validator
-- =======================================================

/**
 * Validates technical and organizational security measures
 * LGPD Article 46, 47, 48, 49
 */
CREATE OR REPLACE FUNCTION validate_lgpd_security_measures(
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
    
    -- Security metrics
    v_security_measures JSONB;
    v_missing_measures TEXT[];
    v_total_measures INTEGER;
    v_implemented_measures INTEGER;
    v_partial_measures INTEGER;
    v_has_healthcare_measures BOOLEAN;
    v_has_ar_vr_measures BOOLEAN;
BEGIN
    -- Get validator ID
    SELECT validator_id INTO v_validator_id 
    FROM compliance_validators 
    WHERE validator_name = 'LGPD-Security-Measures-Validator' 
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

    -- 1. Get security measures information
    SELECT setting_value::JSONB INTO v_security_measures
    FROM iam.tenant_settings
    WHERE tenant_id = p_tenant_id
    AND setting_key = 'security_measures';
    
    -- Check if healthcare-specific measures are implemented
    SELECT COALESCE(setting_value::BOOLEAN, FALSE) INTO v_has_healthcare_measures
    FROM iam.tenant_settings
    WHERE tenant_id = p_tenant_id
    AND setting_key = 'healthcare_security_measures_implemented';
    
    -- Check if AR/VR security measures are implemented
    SELECT COALESCE(setting_value::BOOLEAN, FALSE) INTO v_has_ar_vr_measures
    FROM iam.tenant_settings
    WHERE tenant_id = p_tenant_id
    AND setting_key = 'ar_vr_security_measures_implemented';
    
    -- If security measures setting doesn't exist, create an issue
    IF v_security_measures IS NULL THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'issue_id', gen_random_uuid(),
            'issue_type', 'missing_security_measures',
            'description', 'No security measures documented',
            'severity', 'critical',
            'remediation', 'Document and implement security measures as required by LGPD Article 46'
        ));
    ELSE
        -- 2. Analyze implemented security measures
        WITH required_measures(measure_name) AS (
            VALUES 
                ('access_control'),
                ('authentication_mfa'),
                ('encryption_data_at_rest'),
                ('encryption_data_in_transit'),
                ('logging_auditing'),
                ('vulnerability_management'),
                ('incident_response'),
                ('data_backup'),
                ('secure_development'),
                ('regular_security_testing')
        ),
        implemented_measures AS (
            SELECT 
                measure_name,
                v_security_measures->>measure_name AS status
            FROM required_measures
        )
        SELECT 
            COUNT(*),
            COUNT(*) FILTER (WHERE status = 'implemented'),
            COUNT(*) FILTER (WHERE status = 'partial'),
            ARRAY_AGG(measure_name) FILTER (WHERE status IS NULL OR status = 'not_implemented')
        INTO
            v_total_measures,
            v_implemented_measures,
            v_partial_measures,
            v_missing_measures
        FROM implemented_measures;
        
        -- 3. Add points based on implementation
        -- One point for each 20% of measures fully implemented
        v_points := v_points + FLOOR(v_implemented_measures::FLOAT / v_total_measures::FLOAT * 5);
        
        -- Additional points for partial implementation
        v_points := v_points + FLOOR(v_partial_measures::FLOAT / v_total_measures::FLOAT * 2);
        
        -- Add additional points for healthcare-specific security measures
        IF v_has_healthcare_measures THEN
            v_points := v_points + 2;
        END IF;
        
        -- Add additional points for AR/VR security measures
        IF v_has_ar_vr_measures THEN
            v_points := v_points + 1;
        END IF;
        
        -- Add issues for missing measures
        IF array_length(v_missing_measures, 1) > 0 THEN
            v_issues := array_append(v_issues, jsonb_build_object(
                'issue_id', gen_random_uuid(),
                'issue_type', 'missing_security_measures',
                'description', 'Required security measures not implemented',
                'affected_items', v_missing_measures,
                'severity', 'high',
                'remediation', 'Implement all required security measures in compliance with LGPD Article 46'
            ));
        END IF;
    END IF;
    
    -- Add healthcare-specific security issues if needed
    IF NOT v_has_healthcare_measures AND EXISTS (
        SELECT 1 FROM iam.data_processing
        WHERE tenant_id = p_tenant_id
        AND is_sensitive = TRUE
        AND (data_category LIKE '%health%' OR data_category LIKE '%medical%')
    ) THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'issue_id', gen_random_uuid(),
            'issue_type', 'missing_healthcare_security',
            'description', 'Healthcare data processed without specific security measures',
            'severity', 'critical',
            'remediation', 'Implement healthcare-specific security controls for sensitive health data'
        ));
    END IF;
    
    -- Add AR/VR-specific security issues if needed
    IF NOT v_has_ar_vr_measures AND EXISTS (
        SELECT 1 FROM iam.tenant_settings
        WHERE tenant_id = p_tenant_id
        AND setting_key = 'ar_vr_features_enabled'
        AND setting_value::BOOLEAN = TRUE
    ) THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'issue_id', gen_random_uuid(),
            'issue_type', 'missing_ar_vr_security',
            'description', 'AR/VR features enabled without specific security measures',
            'severity', 'high',
            'remediation', 'Implement AR/VR-specific security controls to protect spatial and perceptual data'
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
        'Security measures validation: %s%% compliant. %s of %s measures fully implemented, %s partially implemented.',
        v_percentage,
        v_implemented_measures,
        v_total_measures,
        v_partial_measures
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
        'validator', 'LGPD-Security-Measures-Validator',
        'status', v_status,
        'summary', v_summary,
        'score', v_percentage,
        'issues', v_issues,
        'execution_date', NOW()
    );
    
    RETURN v_result;
END;
$$ LANGUAGE plpgsql;

-- Create a trigger to initialize validators on database creation/update
CREATE OR REPLACE FUNCTION initialize_lgpd_validators_trigger() RETURNS TRIGGER AS $$
BEGIN
    PERFORM initialize_lgpd_validators();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop existing trigger if exists
DROP TRIGGER IF EXISTS trg_initialize_lgpd_validators ON compliance_validators;

-- Create trigger
CREATE TRIGGER trg_initialize_lgpd_validators
    AFTER INSERT ON compliance_validators
    EXECUTE FUNCTION initialize_lgpd_validators_trigger();

-- Initialize LGPD validators
SELECT initialize_lgpd_validators();
