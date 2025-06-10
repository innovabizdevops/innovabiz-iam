-- INNOVABIZ IAM Module - GDPR Compliance Validators
-- Version: 1.0.0
-- Date: 09/05/2025
-- Author: Eduardo Jeremias
-- Description: Implementation of GDPR compliance validators for IAM

-- Set search path
SET search_path TO iam_compliance, iam, public;

-- Register GDPR validators in the registry
INSERT INTO compliance_validators (
    validator_name,
    validator_type,
    description,
    framework,
    jurisdiction,
    industry_sector,
    function_schema,
    function_name,
    parameters,
    version,
    dependencies,
    author,
    documentation_url
) VALUES
(
    'GDPR Password Policy Validator',
    'password_policy',
    'Validates that password policies meet GDPR recommendations for strong authentication',
    'GDPR',
    ARRAY['EU'],
    ARRAY['all'],
    'iam_compliance',
    'validate_gdpr_password_policy',
    jsonb_build_object(
        'min_length', 12,
        'require_uppercase', true,
        'require_lowercase', true,
        'require_numbers', true,
        'require_special_chars', true,
        'max_age_days', 90
    ),
    '1.0',
    ARRAY['iam.password_policies'],
    'Eduardo Jeremias',
    'https://innovabiz.com/docs/compliance/gdpr/password'
),
(
    'GDPR Data Access Control Validator',
    'access_control',
    'Validates that IAM access controls meet GDPR requirements for data protection',
    'GDPR',
    ARRAY['EU'],
    ARRAY['all'],
    'iam_compliance',
    'validate_gdpr_access_controls',
    jsonb_build_object(
        'require_justification', true,
        'max_access_duration_days', 90,
        'privileged_access_review_days', 30
    ),
    '1.0',
    ARRAY['iam.roles', 'iam.permissions', 'iam.user_roles'],
    'Eduardo Jeremias',
    'https://innovabiz.com/docs/compliance/gdpr/access-control'
),
(
    'GDPR Consent Management Validator',
    'consent',
    'Validates that consent mechanisms meet GDPR requirements',
    'GDPR',
    ARRAY['EU'],
    ARRAY['all'],
    'iam_compliance',
    'validate_gdpr_consent_management',
    jsonb_build_object(
        'require_explicit_consent', true,
        'support_withdrawal', true,
        'consent_tracking', true
    ),
    '1.0',
    ARRAY['iam.user_consents'],
    'Eduardo Jeremias',
    'https://innovabiz.com/docs/compliance/gdpr/consent'
),
(
    'GDPR Data Subject Rights Validator',
    'subject_rights',
    'Validates that systems support GDPR data subject rights',
    'GDPR',
    ARRAY['EU'],
    ARRAY['all'],
    'iam_compliance',
    'validate_gdpr_data_subject_rights',
    jsonb_build_object(
        'right_to_access', true,
        'right_to_rectification', true,
        'right_to_erasure', true,
        'right_to_portability', true,
        'right_to_object', true
    ),
    '1.0',
    ARRAY['iam.users', 'iam.user_data'],
    'Eduardo Jeremias',
    'https://innovabiz.com/docs/compliance/gdpr/subject-rights'
),
(
    'GDPR Data Breach Notification Validator',
    'breach_notification',
    'Validates that breach notification mechanisms meet GDPR requirements',
    'GDPR',
    ARRAY['EU'],
    ARRAY['all'],
    'iam_compliance',
    'validate_gdpr_breach_notification',
    jsonb_build_object(
        'notification_timeline_hours', 72,
        'documentation_required', true
    ),
    '1.0',
    ARRAY['iam.security_incidents'],
    'Eduardo Jeremias',
    'https://innovabiz.com/docs/compliance/gdpr/breach-notification'
);

-- Function to validate password policy compliance with GDPR
CREATE OR REPLACE FUNCTION validate_gdpr_password_policy(
    p_tenant_id UUID,
    p_parameters JSONB DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_result JSONB;
    v_issues JSONB[];
    v_policy RECORD;
    v_default_params JSONB;
    v_effective_params JSONB;
BEGIN
    -- Set default parameters if not provided
    v_default_params := jsonb_build_object(
        'min_length', 12,
        'require_uppercase', true,
        'require_lowercase', true,
        'require_numbers', true,
        'require_special_chars', true,
        'max_age_days', 90
    );
    
    -- Merge with provided parameters, preferring provided ones
    v_effective_params := COALESCE(p_parameters, '{}'::jsonb) || v_default_params;
    
    -- Retrieve password policies for the tenant
    SELECT * INTO v_policy
    FROM iam.password_policies
    WHERE tenant_id = p_tenant_id;
    
    IF NOT FOUND THEN
        RETURN jsonb_build_object(
            'status', 'error',
            'message', 'No password policy found for tenant',
            'compliance_score', 0,
            'issues', jsonb_build_array(
                jsonb_build_object(
                    'severity', 'critical',
                    'description', 'No password policy defined',
                    'recommendation', 'Create a password policy compliant with GDPR requirements'
                )
            )
        );
    END IF;
    
    -- Initialize empty issues array
    v_issues := ARRAY[]::JSONB[];
    
    -- Check minimum length
    IF v_policy.min_length < (v_effective_params->>'min_length')::INTEGER THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'severity', 'high',
            'description', format('Password minimum length (%s) is less than GDPR recommended (%s)', 
                                  v_policy.min_length, v_effective_params->>'min_length'),
            'recommendation', format('Increase minimum password length to at least %s characters', 
                                    v_effective_params->>'min_length')
        ));
    END IF;
    
    -- Check complexity requirements
    IF (v_effective_params->>'require_uppercase')::BOOLEAN AND NOT v_policy.require_uppercase THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'severity', 'medium',
            'description', 'Password policy does not require uppercase characters',
            'recommendation', 'Enable uppercase character requirement in password policy'
        ));
    END IF;
    
    IF (v_effective_params->>'require_lowercase')::BOOLEAN AND NOT v_policy.require_lowercase THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'severity', 'medium',
            'description', 'Password policy does not require lowercase characters',
            'recommendation', 'Enable lowercase character requirement in password policy'
        ));
    END IF;
    
    IF (v_effective_params->>'require_numbers')::BOOLEAN AND NOT v_policy.require_numbers THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'severity', 'medium',
            'description', 'Password policy does not require numeric characters',
            'recommendation', 'Enable numeric character requirement in password policy'
        ));
    END IF;
    
    IF (v_effective_params->>'require_special_chars')::BOOLEAN AND NOT v_policy.require_special_chars THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'severity', 'medium',
            'description', 'Password policy does not require special characters',
            'recommendation', 'Enable special character requirement in password policy'
        ));
    END IF;
    
    -- Check password expiration
    IF v_policy.max_age_days IS NULL OR v_policy.max_age_days > (v_effective_params->>'max_age_days')::INTEGER THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'severity', 'high',
            'description', format('Password expiration period (%s days) exceeds GDPR recommended maximum (%s days)',
                                 COALESCE(v_policy.max_age_days::TEXT, 'never'), v_effective_params->>'max_age_days'),
            'recommendation', format('Set password expiration to a maximum of %s days', 
                                    v_effective_params->>'max_age_days')
        ));
    END IF;
    
    -- Check history requirement
    IF v_policy.history_count < 4 THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'severity', 'medium',
            'description', format('Password history count (%s) is less than GDPR recommended minimum (4)',
                                 v_policy.history_count),
            'recommendation', 'Increase password history requirement to at least 4 previous passwords'
        ));
    END IF;
    
    -- Calculate compliance score based on issues
    DECLARE
        v_total_checks INTEGER := 7; -- Total number of checks performed
        v_passed_checks INTEGER := v_total_checks - array_length(v_issues, 1);
        v_score NUMERIC;
    BEGIN
        -- If there are no issues, score is 100
        IF array_length(v_issues, 1) IS NULL THEN
            v_score := 100;
        ELSE
            -- Calculate score as percentage of passed checks
            v_score := (v_passed_checks::NUMERIC / v_total_checks) * 100;
        END IF;
        
        -- Prepare result
        v_result := jsonb_build_object(
            'status', CASE WHEN array_length(v_issues, 1) IS NULL THEN 'passed' ELSE 'failed' END,
            'message', CASE 
                WHEN array_length(v_issues, 1) IS NULL THEN 'Password policy is GDPR compliant' 
                ELSE format('Password policy has %s compliance issues', array_length(v_issues, 1))
            END,
            'compliance_score', v_score,
            'issues', COALESCE(array_to_json(v_issues)::JSONB, '[]'::JSONB),
            'details', jsonb_build_object(
                'total_checks', v_total_checks,
                'passed_checks', v_passed_checks,
                'policy_id', v_policy.id,
                'tenant_id', p_tenant_id
            )
        );
    END;
    
    RETURN v_result;
EXCEPTION
    WHEN OTHERS THEN
        RETURN jsonb_build_object(
            'status', 'error',
            'message', 'Error executing GDPR password policy validator: ' || SQLERRM,
            'compliance_score', 0,
            'issues', jsonb_build_array(
                jsonb_build_object(
                    'severity', 'critical',
                    'description', 'Validator execution error',
                    'recommendation', 'Review password policy configuration and validator code'
                )
            )
        );
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION validate_gdpr_password_policy(UUID, JSONB) IS 
'Validates that a tenant''s password policy complies with GDPR recommendations';

-- Function to validate access control compliance with GDPR
CREATE OR REPLACE FUNCTION validate_gdpr_access_controls(
    p_tenant_id UUID,
    p_parameters JSONB DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_result JSONB;
    v_issues JSONB[];
    v_privileged_roles_count INTEGER;
    v_roles_with_justification_count INTEGER;
    v_roles_with_expiration_count INTEGER;
    v_roles_with_review_count INTEGER;
    v_total_roles_count INTEGER;
    v_default_params JSONB;
    v_effective_params JSONB;
BEGIN
    -- Set default parameters if not provided
    v_default_params := jsonb_build_object(
        'require_justification', true,
        'max_access_duration_days', 90,
        'privileged_access_review_days', 30
    );
    
    -- Merge with provided parameters, preferring provided ones
    v_effective_params := COALESCE(p_parameters, '{}'::jsonb) || v_default_params;
    
    -- Count total roles for the tenant
    SELECT COUNT(*) INTO v_total_roles_count
    FROM iam.roles
    WHERE tenant_id = p_tenant_id;
    
    IF v_total_roles_count = 0 THEN
        RETURN jsonb_build_object(
            'status', 'error',
            'message', 'No roles found for tenant',
            'compliance_score', 0,
            'issues', jsonb_build_array(
                jsonb_build_object(
                    'severity', 'critical',
                    'description', 'No roles defined for the tenant',
                    'recommendation', 'Create roles with appropriate access controls'
                )
            )
        );
    END IF;
    
    -- Count privileged roles (simplified example - in a real system, would use more sophisticated detection)
    SELECT COUNT(*) INTO v_privileged_roles_count
    FROM iam.roles
    WHERE tenant_id = p_tenant_id
    AND (
        role_type = 'admin' 
        OR name ILIKE '%admin%' 
        OR name ILIKE '%superuser%'
    );
    
    -- Count roles with justification requirement
    SELECT COUNT(*) INTO v_roles_with_justification_count
    FROM iam.roles
    WHERE tenant_id = p_tenant_id
    AND require_justification = true;
    
    -- Count roles with expiration
    SELECT COUNT(*) INTO v_roles_with_expiration_count
    FROM iam.roles
    WHERE tenant_id = p_tenant_id
    AND max_grant_days IS NOT NULL
    AND max_grant_days <= (v_effective_params->>'max_access_duration_days')::INTEGER;
    
    -- Count roles with review period
    SELECT COUNT(*) INTO v_roles_with_review_count
    FROM iam.roles
    WHERE tenant_id = p_tenant_id
    AND review_period_days IS NOT NULL
    AND review_period_days <= (v_effective_params->>'privileged_access_review_days')::INTEGER;
    
    -- Initialize empty issues array
    v_issues := ARRAY[]::JSONB[];
    
    -- Check justification requirements
    IF (v_effective_params->>'require_justification')::BOOLEAN AND v_roles_with_justification_count < v_privileged_roles_count THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'severity', 'high',
            'description', format('Only %s of %s privileged roles require justification for assignment',
                                 v_roles_with_justification_count, v_privileged_roles_count),
            'recommendation', 'Configure all privileged roles to require justification for assignment'
        ));
    END IF;
    
    -- Check expiration requirements
    IF v_roles_with_expiration_count < v_privileged_roles_count THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'severity', 'high',
            'description', format('Only %s of %s privileged roles have appropriate expiration limits',
                                 v_roles_with_expiration_count, v_privileged_roles_count),
            'recommendation', format('Configure all privileged roles to expire after a maximum of %s days',
                                   v_effective_params->>'max_access_duration_days')
        ));
    END IF;
    
    -- Check review period requirements
    IF v_roles_with_review_count < v_privileged_roles_count THEN
        v_issues := array_append(v_issues, jsonb_build_object(
            'severity', 'medium',
            'description', format('Only %s of %s privileged roles have appropriate review periods',
                                 v_roles_with_review_count, v_privileged_roles_count),
            'recommendation', format('Configure all privileged roles to be reviewed every %s days',
                                   v_effective_params->>'privileged_access_review_days')
        ));
    END IF;
    
    -- Calculate compliance score based on issues
    DECLARE
        v_total_checks INTEGER := 3; -- Total number of checks performed
        v_passed_checks INTEGER := v_total_checks - array_length(v_issues, 1);
        v_score NUMERIC;
    BEGIN
        -- If there are no issues, score is 100
        IF array_length(v_issues, 1) IS NULL THEN
            v_score := 100;
        ELSE
            -- Calculate score as percentage of passed checks
            v_score := (v_passed_checks::NUMERIC / v_total_checks) * 100;
        END IF;
        
        -- Prepare result
        v_result := jsonb_build_object(
            'status', CASE WHEN array_length(v_issues, 1) IS NULL THEN 'passed' ELSE 'failed' END,
            'message', CASE 
                WHEN array_length(v_issues, 1) IS NULL THEN 'Access controls are GDPR compliant' 
                ELSE format('Access controls have %s compliance issues', array_length(v_issues, 1))
            END,
            'compliance_score', v_score,
            'issues', COALESCE(array_to_json(v_issues)::JSONB, '[]'::JSONB),
            'details', jsonb_build_object(
                'total_checks', v_total_checks,
                'passed_checks', v_passed_checks,
                'total_roles', v_total_roles_count,
                'privileged_roles', v_privileged_roles_count,
                'tenant_id', p_tenant_id
            )
        );
    END;
    
    RETURN v_result;
EXCEPTION
    WHEN OTHERS THEN
        RETURN jsonb_build_object(
            'status', 'error',
            'message', 'Error executing GDPR access controls validator: ' || SQLERRM,
            'compliance_score', 0,
            'issues', jsonb_build_array(
                jsonb_build_object(
                    'severity', 'critical',
                    'description', 'Validator execution error',
                    'recommendation', 'Review access control configuration and validator code'
                )
            )
        );
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION validate_gdpr_access_controls(UUID, JSONB) IS 
'Validates that a tenant''s access controls comply with GDPR recommendations';

-- Reset search path
RESET search_path;
