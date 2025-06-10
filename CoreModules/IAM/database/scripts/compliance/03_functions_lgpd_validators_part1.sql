-- ===============================================================================
-- IAM LGPD Compliance Validators (Part 1)
-- INNOVABIZ Platform
-- Version: 1.0
-- Date: January 2025
-- Description: Implementation of LGPD compliance validators for the IAM module
-- ===============================================================================

-- Schema reference for IAM LGPD validators
SET search_path TO iam_compliance, iam, public;

-- =======================================================
-- Base LGPD Validator Function
-- =======================================================

/**
 * Base function for all LGPD validation functions
 * Creates a standardized approach to validator registration and execution
 */
CREATE OR REPLACE FUNCTION register_lgpd_validator(
    p_validator_name TEXT,
    p_description TEXT,
    p_category TEXT,
    p_article_reference TEXT,
    p_severity TEXT DEFAULT 'high'
) RETURNS UUID AS $$
DECLARE
    v_validator_id UUID;
BEGIN
    -- Check if the validator already exists
    SELECT validator_id INTO v_validator_id
    FROM compliance_validators
    WHERE validator_name = p_validator_name
    AND regulatory_framework = 'LGPD';
    
    -- If not, create a new validator
    IF v_validator_id IS NULL THEN
        v_validator_id := gen_random_uuid();
        
        INSERT INTO compliance_validators (
            validator_id,
            validator_name,
            description,
            category,
            regulatory_framework,
            article_reference,
            severity,
            created_at,
            last_updated_at,
            active
        ) VALUES (
            v_validator_id,
            p_validator_name,
            p_description,
            p_category,
            'LGPD',
            p_article_reference,
            p_severity,
            NOW(),
            NOW(),
            TRUE
        );
    ELSE
        -- Update the existing validator
        UPDATE compliance_validators
        SET description = p_description,
            category = p_category,
            article_reference = p_article_reference,
            severity = p_severity,
            last_updated_at = NOW()
        WHERE validator_id = v_validator_id;
    END IF;
    
    RETURN v_validator_id;
END;
$$ LANGUAGE plpgsql;

-- Register all LGPD validators during database initialization
CREATE OR REPLACE FUNCTION initialize_lgpd_validators() RETURNS VOID AS $$
DECLARE
    v_validator_id UUID;
BEGIN
    -- Legal Basis Validator
    v_validator_id := register_lgpd_validator(
        'LGPD-Legal-Basis-Validator',
        'Validates that personal data processing has a valid legal basis recorded',
        'Data Protection Principles',
        'Art. 7, 8, 11',
        'critical'
    );
    
    -- Consent Management Validator
    v_validator_id := register_lgpd_validator(
        'LGPD-Consent-Management-Validator',
        'Validates consent management practices for user data processing',
        'Individual Rights',
        'Art. 7, 8, 9',
        'high'
    );
    
    -- Data Subject Rights Validator
    v_validator_id := register_lgpd_validator(
        'LGPD-Data-Subject-Rights-Validator',
        'Validates implementation of data subject rights',
        'Individual Rights',
        'Art. 18, 19, 20',
        'high'
    );
    
    -- Minimal Data Collection Validator
    v_validator_id := register_lgpd_validator(
        'LGPD-Minimal-Data-Collection-Validator',
        'Validates that only necessary personal data is collected',
        'Data Protection Principles',
        'Art. 6, 10',
        'medium'
    );
    
    -- Data Retention Validator
    v_validator_id := register_lgpd_validator(
        'LGPD-Data-Retention-Validator',
        'Validates data retention policies and procedures',
        'Data Lifecycle',
        'Art. 15, 16',
        'high'
    );
    
    -- DPO Appointment Validator
    v_validator_id := register_lgpd_validator(
        'LGPD-DPO-Appointment-Validator',
        'Validates appointment of Data Protection Officer',
        'Governance',
        'Art. 41',
        'medium'
    );
    
    -- International Transfer Validator
    v_validator_id := register_lgpd_validator(
        'LGPD-International-Transfer-Validator',
        'Validates controls for international data transfers',
        'Data Security',
        'Art. 33, 34, 35',
        'high'
    );
    
    -- Sensitive Data Protection Validator
    v_validator_id := register_lgpd_validator(
        'LGPD-Sensitive-Data-Validator',
        'Validates enhanced protection for sensitive personal data',
        'Data Security',
        'Art. 11, 12, 13',
        'critical'
    );
    
    -- Children and Adolescent Data Validator
    v_validator_id := register_lgpd_validator(
        'LGPD-Children-Data-Validator',
        'Validates special protections for children and adolescent data',
        'Special Categories',
        'Art. 14',
        'critical'
    );
    
    -- Data Breach Notification Validator
    v_validator_id := register_lgpd_validator(
        'LGPD-Data-Breach-Notification-Validator',
        'Validates data breach notification processes',
        'Incident Management',
        'Art. 48',
        'high'
    );
    
    -- Privacy by Design Validator
    v_validator_id := register_lgpd_validator(
        'LGPD-Privacy-By-Design-Validator',
        'Validates implementation of privacy by design principles',
        'System Design',
        'Art. 6, 46',
        'medium'
    );
    
    -- Security Measures Validator
    v_validator_id := register_lgpd_validator(
        'LGPD-Security-Measures-Validator',
        'Validates technical and organizational security measures',
        'Data Security',
        'Art. 46, 47, 48, 49',
        'high'
    );

    -- Record of Processing Activities Validator
    v_validator_id := register_lgpd_validator(
        'LGPD-Record-Processing-Activities-Validator',
        'Validates maintenance of records of processing activities',
        'Documentation',
        'Art. 37, 38',
        'medium'
    );
    
    -- Impact Assessment Validator
    v_validator_id := register_lgpd_validator(
        'LGPD-Impact-Assessment-Validator',
        'Validates performance of data protection impact assessments',
        'Risk Management',
        'Art. 5, 38',
        'medium'
    );
    
    -- Accountability Validator
    v_validator_id := register_lgpd_validator(
        'LGPD-Accountability-Validator',
        'Validates demonstration of compliance and accountability',
        'Governance',
        'Art. 6, 50',
        'high'
    );
END;
$$ LANGUAGE plpgsql;

-- =======================================================
-- LGPD Legal Basis Validator
-- =======================================================

/**
 * Validates that each personal data processing operation has a valid legal basis recorded
 * LGPD Article 7, 8, 11
 */
CREATE OR REPLACE FUNCTION validate_lgpd_legal_basis(
    p_tenant_id UUID,
    p_parameters JSONB DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_result JSONB;
    v_validator_id UUID;
    v_issues JSONB[] := '{}';
    v_check_id UUID;
    v_summary TEXT;
    v_legal_bases TEXT[];
    v_missing_legal_bases TEXT[];
    v_processed_data_count INTEGER;
    v_missing_legal_basis_count INTEGER;
    v_percentage NUMERIC;
    v_status TEXT;
BEGIN
    -- Get validator ID
    SELECT validator_id INTO v_validator_id 
    FROM compliance_validators 
    WHERE validator_name = 'LGPD-Legal-Basis-Validator' 
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

    -- Get valid LGPD legal bases
    v_legal_bases := ARRAY[
        'consent', 'legal_obligation', 'execution_of_contract', 
        'legitimate_interest', 'protection_of_life', 'protection_of_health',
        'public_task', 'research', 'exercise_of_rights',
        'credit_protection'
    ];
    
    -- 1. Check for data processing operations without legal basis
    WITH data_processing AS (
        SELECT 
            dp.processing_id,
            dp.data_category,
            dp.legal_basis,
            CASE WHEN dp.legal_basis IS NULL OR dp.legal_basis = '' 
                 THEN TRUE ELSE FALSE END AS missing_legal_basis,
            CASE WHEN dp.legal_basis IS NOT NULL AND 
                      dp.legal_basis != '' AND 
                      NOT (dp.legal_basis = ANY(v_legal_bases))
                 THEN TRUE ELSE FALSE END AS invalid_legal_basis
        FROM iam.data_processing dp
        WHERE dp.tenant_id = p_tenant_id
    )
    SELECT 
        COUNT(*),
        COUNT(*) FILTER (WHERE missing_legal_basis),
        ARRAY_AGG(DISTINCT data_category) FILTER (WHERE missing_legal_basis)
    INTO 
        v_processed_data_count,
        v_missing_legal_basis_count,
        v_missing_legal_bases
    FROM data_processing;
    
    -- Calculate compliance percentage
    IF v_processed_data_count > 0 THEN
        v_percentage := (100 - (v_missing_legal_basis_count::NUMERIC / v_processed_data_count::NUMERIC * 100))::NUMERIC(5,2);
    ELSE
        v_percentage := 0;
    END IF;
    
    -- Determine status
    IF v_percentage >= 95 THEN
        v_status := 'compliant';
    ELSIF v_percentage >= 80 THEN
        v_status := 'partially_compliant';
    ELSE
        v_status := 'non_compliant';
    END IF;
    
    -- 2. Add issues for missing legal bases
    IF v_missing_legal_basis_count > 0 THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'issue_id', gen_random_uuid(),
            'issue_type', 'missing_legal_basis',
            'description', 'Processing operations without specified legal basis',
            'affected_items', v_missing_legal_bases,
            'count', v_missing_legal_basis_count,
            'severity', 'critical',
            'remediation', 'Define and document a valid legal basis for each data processing operation'
        ));
    END IF;
    
    -- 3. Check for sensitive data processing
    WITH sensitive_data AS (
        SELECT 
            dp.processing_id,
            dp.data_category,
            dp.legal_basis,
            dp.is_sensitive
        FROM iam.data_processing dp
        WHERE dp.tenant_id = p_tenant_id
        AND dp.is_sensitive = TRUE
    )
    SELECT 
        COUNT(*),
        COUNT(*) FILTER (WHERE legal_basis NOT IN ('consent', 'legal_obligation', 'protection_of_life', 'protection_of_health', 'research', 'exercise_of_rights')),
        ARRAY_AGG(DISTINCT data_category) FILTER (WHERE legal_basis NOT IN ('consent', 'legal_obligation', 'protection_of_life', 'protection_of_health', 'research', 'exercise_of_rights'))
    INTO 
        v_processed_data_count,
        v_missing_legal_basis_count,
        v_missing_legal_bases
    FROM sensitive_data;
    
    -- 4. Add issues for sensitive data with invalid legal basis
    IF v_missing_legal_basis_count > 0 THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'issue_id', gen_random_uuid(),
            'issue_type', 'invalid_sensitive_data_legal_basis',
            'description', 'Sensitive data processed with invalid legal basis',
            'affected_items', v_missing_legal_bases,
            'count', v_missing_legal_basis_count,
            'severity', 'critical',
            'remediation', 'Ensure that sensitive data is only processed with proper legal basis according to LGPD Art. 11'
        ));
    END IF;
    
    -- Create summary
    v_summary := format(
        'Legal basis validation: %s%% compliant. %s processing operations reviewed, %s without valid legal basis.',
        v_percentage,
        v_processed_data_count,
        v_missing_legal_basis_count
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
        'validator', 'LGPD-Legal-Basis-Validator',
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
-- LGPD Consent Management Validator
-- =======================================================

/**
 * Validates consent management practices for user data processing
 * LGPD Article 7, 8, 9
 */
CREATE OR REPLACE FUNCTION validate_lgpd_consent_management(
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
    v_max_points NUMERIC := 8;
    v_percentage NUMERIC;
    
    -- Counters for consent validation
    v_consent_records_count INTEGER;
    v_invalid_count INTEGER;
    v_missing_info_count INTEGER;
    v_expired_count INTEGER;
    v_withdrawable_count INTEGER;
    v_child_consent_count INTEGER;
    v_child_consent_invalid_count INTEGER;
    v_categories TEXT[];
    v_invalid_categories TEXT[];
BEGIN
    -- Get validator ID
    SELECT validator_id INTO v_validator_id 
    FROM compliance_validators 
    WHERE validator_name = 'LGPD-Consent-Management-Validator' 
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

    -- 1. Check for explicit, specific, and informed consent
    SELECT
        COUNT(*),
        COUNT(*) FILTER (WHERE NOT is_explicit OR NOT is_specific),
        COUNT(*) FILTER (WHERE information_provided IS NULL OR NOT information_provided),
        ARRAY_AGG(DISTINCT data_category) FILTER (WHERE NOT is_explicit OR NOT is_specific)
    INTO
        v_consent_records_count,
        v_invalid_count,
        v_missing_info_count,
        v_invalid_categories
    FROM iam.consent_records
    WHERE tenant_id = p_tenant_id
    AND legal_basis = 'consent';
    
    -- Add points if consents are explicit and specific
    IF v_consent_records_count > 0 AND (v_invalid_count::FLOAT / v_consent_records_count::FLOAT) < 0.05 THEN
        v_points := v_points + 1;
    END IF;
    
    -- Add points if adequate information is provided
    IF v_consent_records_count > 0 AND (v_missing_info_count::FLOAT / v_consent_records_count::FLOAT) < 0.05 THEN
        v_points := v_points + 1;
    END IF;
    
    -- 2. Check for expired consent or without expiration
    SELECT
        COUNT(*) FILTER (WHERE expiration_date IS NOT NULL AND expiration_date < CURRENT_DATE)
    INTO
        v_expired_count
    FROM iam.consent_records
    WHERE tenant_id = p_tenant_id
    AND legal_basis = 'consent';
    
    -- Add points if consents are not expired
    IF v_consent_records_count > 0 AND (v_expired_count::FLOAT / v_consent_records_count::FLOAT) < 0.02 THEN
        v_points := v_points + 1;
    END IF;
    
    -- 3. Check for ability to withdraw consent
    SELECT
        COUNT(*) FILTER (WHERE can_withdraw)
    INTO
        v_withdrawable_count
    FROM iam.consent_records
    WHERE tenant_id = p_tenant_id
    AND legal_basis = 'consent';
    
    -- Add points if consents can be withdrawn
    IF v_consent_records_count > 0 AND (v_withdrawable_count::FLOAT / v_consent_records_count::FLOAT) > 0.98 THEN
        v_points := v_points + 1;
    END IF;
    
    -- 4. Check for consent specifically related to children's data
    SELECT
        COUNT(*),
        COUNT(*) FILTER (WHERE NOT has_parental_consent)
    INTO
        v_child_consent_count,
        v_child_consent_invalid_count
    FROM iam.consent_records
    WHERE tenant_id = p_tenant_id
    AND legal_basis = 'consent'
    AND data_subject_type = 'child';
    
    -- Add points for proper children's data consent handling
    IF v_child_consent_count = 0 OR (v_child_consent_invalid_count::FLOAT / v_child_consent_count::FLOAT) < 0.01 THEN
        v_points := v_points + 1;
    END IF;
    
    -- 5. Check for consent revocation mechanism
    IF EXISTS (
        SELECT 1 FROM iam.tenant_settings 
        WHERE tenant_id = p_tenant_id 
        AND setting_key = 'consent_revocation_mechanism' 
        AND setting_value::BOOLEAN = TRUE
    ) THEN
        v_points := v_points + 1;
    END IF;
    
    -- 6. Check for consent records retention
    IF EXISTS (
        SELECT 1 FROM iam.tenant_settings 
        WHERE tenant_id = p_tenant_id 
        AND setting_key = 'consent_records_retention' 
        AND setting_value::JSONB->>'enabled' = 'true'
    ) THEN
        v_points := v_points + 1;
    END IF;
    
    -- 7. Check for separate consent for each purpose
    IF EXISTS (
        SELECT 1 FROM iam.tenant_settings 
        WHERE tenant_id = p_tenant_id 
        AND setting_key = 'purpose_specific_consent' 
        AND setting_value::BOOLEAN = TRUE
    ) THEN
        v_points := v_points + 1;
    END IF;
    
    -- Add issues based on findings
    IF v_invalid_count > 0 THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'issue_id', gen_random_uuid(),
            'issue_type', 'invalid_consent',
            'description', 'Consent records are not explicit or specific',
            'affected_items', v_invalid_categories,
            'count', v_invalid_count,
            'severity', 'high',
            'remediation', 'Ensure that all consent is explicit, specific, and freely given'
        ));
    END IF;
    
    IF v_missing_info_count > 0 THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'issue_id', gen_random_uuid(),
            'issue_type', 'uninformed_consent',
            'description', 'Consent obtained without adequate information provided',
            'count', v_missing_info_count,
            'severity', 'high',
            'remediation', 'Provide clear information about data processing before obtaining consent'
        ));
    END IF;
    
    IF v_expired_count > 0 THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'issue_id', gen_random_uuid(),
            'issue_type', 'expired_consent',
            'description', 'Processing based on expired consent',
            'count', v_expired_count,
            'severity', 'high',
            'remediation', 'Obtain renewed consent or stop processing data with expired consent'
        ));
    END IF;
    
    IF v_withdrawable_count < v_consent_records_count THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'issue_id', gen_random_uuid(),
            'issue_type', 'non_withdrawable_consent',
            'description', 'Consent records without withdrawal option',
            'count', v_consent_records_count - v_withdrawable_count,
            'severity', 'high',
            'remediation', 'Implement mechanism to allow withdrawal of consent as easily as it was given'
        ));
    END IF;
    
    IF v_child_consent_invalid_count > 0 THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'issue_id', gen_random_uuid(),
            'issue_type', 'invalid_child_consent',
            'description', 'Processing of children\'s data without valid parental consent',
            'count', v_child_consent_invalid_count,
            'severity', 'critical',
            'remediation', 'Implement proper parental consent verification for children\'s data'
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
        'Consent management validation: %s%% compliant. %s of %s criteria met.',
        v_percentage,
        v_points,
        v_max_points
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
        'validator', 'LGPD-Consent-Management-Validator',
        'status', v_status,
        'summary', v_summary,
        'score', v_percentage,
        'issues', v_issues,
        'execution_date', NOW()
    );
    
    RETURN v_result;
END;
$$ LANGUAGE plpgsql;
