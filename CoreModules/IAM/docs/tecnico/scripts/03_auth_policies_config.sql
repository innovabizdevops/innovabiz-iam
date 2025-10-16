-- ==========================================================================
-- INNOVABIZ - Scripts de Configuração de Políticas de Autenticação
-- Versão: 1.0.0
-- Data de Criação: 14/05/2025
-- Autor: INNOVABIZ DevOps
-- Descrição: Configuração de políticas e fatores de autenticação (Fase 1)
-- Regiões Suportadas: UE/Portugal, Brasil, Angola, EUA
-- ==========================================================================

-- ==========================================================================
-- Tenant para Ambiente de Desenvolvimento
-- ==========================================================================

-- Criação de tenant de desenvolvimento para testes
INSERT INTO iam_core.tenants (
    tenant_code, 
    tenant_name, 
    tenant_status, 
    tenant_type, 
    organization_name, 
    primary_region_code, 
    supported_regions,
    configurations,
    compliance_settings
) VALUES (
    'INNOVABIZ_DEV', 
    'INNOVABIZ Desenvolvimento', 
    'ACTIVE', 
    'DEVELOPMENT', 
    'INNOVABIZ DevOps', 
    'PT', 
    ARRAY['PT', 'BR', 'AO', 'US'],
    '{
        "auth_timeout_seconds": 900,
        "session_timeout_seconds": 3600,
        "password_policy": {
            "min_length": 8,
            "require_uppercase": true,
            "require_lowercase": true,
            "require_numbers": true,
            "require_special_chars": true,
            "password_history": 5,
            "max_age_days": 90,
            "lockout_threshold": 5,
            "lockout_duration_minutes": 30
        }
    }'::JSONB,
    '{
        "gdpr_compliant": true,
        "lgpd_compliant": true,
        "hipaa_compliant": true,
        "psd2_compliant": true
    }'::JSONB
);

-- Recuperar o tenant_id gerado para uso nas próximas inserções
DO $$
DECLARE
    dev_tenant_id UUID;
BEGIN
    SELECT tenant_id INTO dev_tenant_id FROM iam_core.tenants WHERE tenant_code = 'INNOVABIZ_DEV';

    -- ==========================================================================
    -- Configurações de Fatores de Autenticação
    -- ==========================================================================

    -- Configuração para senhas alfanuméricas
    INSERT INTO iam_core.factor_configurations (
        tenant_id,
        method_id,
        config_name,
        config_parameters,
        is_enabled,
        applies_to_user_types,
        required_for_security_profiles
    ) VALUES (
        dev_tenant_id,
        'KB-01-01',
        'Configuração Padrão de Senhas',
        '{
            "min_length": 8,
            "require_uppercase": true,
            "require_lowercase": true,
            "require_numbers": true,
            "require_special_chars": true,
            "password_history": 5,
            "max_age_days": 90,
            "lockout_threshold": 5,
            "lockout_duration_minutes": 30
        }'::JSONB,
        TRUE,
        ARRAY['HUMAN', 'SERVICE'],
        ARRAY['DEFAULT', 'RESTRICTED', 'ADMIN']
    );

    -- Configuração para senhas complexas
    INSERT INTO iam_core.factor_configurations (
        tenant_id,
        method_id,
        config_name,
        config_parameters,
        is_enabled,
        applies_to_user_types,
        required_for_security_profiles
    ) VALUES (
        dev_tenant_id,
        'KB-01-02',
        'Configuração de Senhas Avançadas',
        '{
            "min_length": 12,
            "require_uppercase": true,
            "require_lowercase": true,
            "require_numbers": true,
            "require_special_chars": true,
            "min_complexity_score": 80,
            "password_history": 10,
            "max_age_days": 60,
            "lockout_threshold": 3,
            "lockout_duration_minutes": 60,
            "password_blacklist": true,
            "breach_detection": true
        }'::JSONB,
        TRUE,
        ARRAY['HUMAN'],
        ARRAY['SENSITIVE', 'ADMIN', 'FINTECH']
    );

    -- Configuração para PIN numérico
    INSERT INTO iam_core.factor_configurations (
        tenant_id,
        method_id,
        config_name,
        config_parameters,
        is_enabled,
        applies_to_user_types,
        required_for_security_profiles
    ) VALUES (
        dev_tenant_id,
        'KB-01-03',
        'Configuração de PIN',
        '{
            "min_length": 6,
            "max_length": 8,
            "require_numbers_only": true,
            "no_sequential_digits": true,
            "no_repeating_digits": true,
            "lockout_threshold": 3,
            "lockout_duration_minutes": 15,
            "require_second_factor": true
        }'::JSONB,
        TRUE,
        ARRAY['HUMAN'],
        ARRAY['MOBILE', 'KIOSK']
    );

    -- Configuração para FIDO2/WebAuthn
    INSERT INTO iam_core.factor_configurations (
        tenant_id,
        method_id,
        config_name,
        config_parameters,
        is_enabled,
        applies_to_user_types,
        required_for_security_profiles
    ) VALUES (
        dev_tenant_id,
        'IN-03-01',
        'Configuração FIDO2/WebAuthn',
        '{
            "challenge_size": 32,
            "timeout": 60000,
            "attestation": "direct",
            "user_verification": "preferred",
            "allowed_algorithms": ["ES256", "RS256"],
            "authenticator_attachment": "cross-platform",
            "require_resident_key": false,
            "allowed_authenticator_types": ["security_key", "platform"]
        }'::JSONB,
        TRUE,
        ARRAY['HUMAN'],
        ARRAY['DEFAULT', 'SENSITIVE', 'ADMIN']
    );

    -- Configuração para Autenticação SMS OTP
    INSERT INTO iam_core.factor_configurations (
        tenant_id,
        method_id,
        config_name,
        config_parameters,
        is_enabled,
        applies_to_user_types,
        required_for_security_profiles
    ) VALUES (
        dev_tenant_id,
        'PB-03-02',
        'Configuração SMS OTP',
        '{
            "code_length": 6,
            "validity_seconds": 300,
            "sms_template": "Seu código de verificação INNOVABIZ é: {code}",
            "attempts_limit": 3,
            "daily_limit": 10,
            "throttle_seconds": 30,
            "allowed_regions": ["PT", "BR", "AO", "US"]
        }'::JSONB,
        TRUE,
        ARRAY['HUMAN'],
        NULL
    );

    -- Configuração para OAuth 2.0
    INSERT INTO iam_core.factor_configurations (
        tenant_id,
        method_id,
        config_name,
        config_parameters,
        is_enabled,
        applies_to_user_types,
        required_for_security_profiles
    ) VALUES (
        dev_tenant_id,
        'FS-01-01',
        'Configuração OAuth 2.0',
        '{
            "supported_flows": ["authorization_code", "refresh_token", "client_credentials"],
            "access_token_lifetime_seconds": 3600,
            "refresh_token_lifetime_seconds": 2592000,
            "require_pkce": true,
            "allowed_scopes": ["openid", "profile", "email", "phone"],
            "token_endpoint_auth_methods": ["client_secret_basic", "client_secret_post", "private_key_jwt"],
            "require_https": true
        }'::JSONB,
        TRUE,
        ARRAY['HUMAN', 'SERVICE'],
        NULL
    );

    -- Configuração para OpenID Connect
    INSERT INTO iam_core.factor_configurations (
        tenant_id,
        method_id,
        config_name,
        config_parameters,
        is_enabled,
        applies_to_user_types,
        required_for_security_profiles
    ) VALUES (
        dev_tenant_id,
        'FS-01-02',
        'Configuração OpenID Connect',
        '{
            "supported_flows": ["authorization_code", "hybrid", "implicit"],
            "id_token_lifetime_seconds": 3600,
            "supported_claims": ["sub", "name", "given_name", "family_name", "email", "email_verified", "phone_number", "address"],
            "supported_response_types": ["code", "id_token", "token", "code id_token", "code token", "id_token token", "code id_token token"],
            "jwt_signing_algs": ["RS256", "ES256"],
            "subject_identifier_types": ["public", "pairwise"]
        }'::JSONB,
        TRUE,
        ARRAY['HUMAN', 'SERVICE'],
        NULL
    );

    -- Configuração para Autenticação Multi-Fator para Profissionais de Saúde
    INSERT INTO iam_core.factor_configurations (
        tenant_id,
        method_id,
        config_name,
        config_parameters,
        is_enabled,
        applies_to_user_types,
        required_for_security_profiles
    ) VALUES (
        dev_tenant_id,
        'HC-01-01',
        'MFA para Profissionais de Saúde',
        '{
            "required_factors": ["knowledge", "possession"],
            "professional_validation": true,
            "professional_db_integration": true,
            "hipaa_compliant": true,
            "gdpr_compliant": true,
            "lgpd_compliant": true,
            "step_up_for_sensitive_operations": true,
            "max_session_minutes": 240,
            "extended_logging": true
        }'::JSONB,
        TRUE,
        ARRAY['HUMAN'],
        ARRAY['HEALTHCARE_PROFESSIONAL']
    );

    -- ==========================================================================
    -- Políticas de Autenticação
    -- ==========================================================================

    -- Política Base MFA
    INSERT INTO iam_core.authentication_policies (
        tenant_id,
        policy_name,
        description,
        policy_type,
        policy_rules,
        applies_to_user_types,
        applies_to_security_profiles,
        applies_to_regions,
        is_enabled,
        priority
    ) VALUES (
        dev_tenant_id,
        'Política MFA Base',
        'Política padrão para requisitos de autenticação multi-fator',
        'MFA',
        '{
            "required_factors": 2,
            "allowed_methods": ["KB-01-01", "KB-01-02", "PB-03-01", "PB-03-02", "IN-03-01"],
            "minimum_factor_strength": "INTERMEDIATE"
        }'::JSONB,
        ARRAY['HUMAN'],
        ARRAY['SENSITIVE', 'ADMIN'],
        NULL,
        TRUE,
        100
    );

    -- Política de Alto Risco
    INSERT INTO iam_core.authentication_policies (
        tenant_id,
        policy_name,
        description,
        policy_type,
        policy_rules,
        applies_to_user_types,
        applies_to_security_profiles,
        applies_to_regions,
        is_enabled,
        priority
    ) VALUES (
        dev_tenant_id,
        'Política para Operações de Alto Risco',
        'Política para operações de risco elevado que requerem autenticação reforçada',
        'STEP_UP',
        '{
            "required_factors": 2,
            "allowed_methods": ["KB-01-02", "PB-01-01", "PB-01-03", "PB-02-02", "IN-01-01"],
            "minimum_factor_strength": "ADVANCED",
            "high_risk_operations": ["transaction_approval", "limit_change", "admin_access", "pii_export"],
            "max_last_factor_age_minutes": 5,
            "require_fresh_authentication": true
        }'::JSONB,
        ARRAY['HUMAN'],
        ARRAY['FINTECH', 'ADMIN', 'SENSITIVE'],
        NULL,
        TRUE,
        50
    );

    -- Política Adaptativa
    INSERT INTO iam_core.authentication_policies (
        tenant_id,
        policy_name,
        description,
        policy_type,
        policy_rules,
        applies_to_user_types,
        applies_to_security_profiles,
        applies_to_regions,
        is_enabled,
        priority
    ) VALUES (
        dev_tenant_id,
        'Política de Autenticação Adaptativa',
        'Política que ajusta requisitos de autenticação com base em fatores de risco',
        'ADAPTIVE',
        '{
            "risk_factors": {
                "new_device": 40,
                "unusual_location": 30,
                "unusual_time": 20,
                "vpn_proxy": 30,
                "impossible_travel": 70,
                "unknown_ip": 20,
                "suspicious_behavior": 50
            },
            "risk_thresholds": {
                "low": 30,
                "medium": 60,
                "high": 80
            },
            "actions": {
                "low": {
                    "required_factors": 1,
                    "allowed_methods": ["KB-01-01", "KB-01-02"]
                },
                "medium": {
                    "required_factors": 2,
                    "allowed_methods": ["KB-01-01", "KB-01-02", "PB-03-01", "PB-03-02", "IN-03-01"]
                },
                "high": {
                    "required_factors": 2,
                    "allowed_methods": ["KB-01-02", "PB-01-01", "PB-01-03", "IN-01-01"],
                    "minimum_factor_strength": "ADVANCED",
                    "max_session_minutes": 30
                }
            }
        }'::JSONB,
        ARRAY['HUMAN'],
        NULL,
        NULL,
        TRUE,
        75
    );

    -- Política de Saúde
    INSERT INTO iam_core.authentication_policies (
        tenant_id,
        policy_name,
        description,
        policy_type,
        policy_rules,
        applies_to_user_types,
        applies_to_security_profiles,
        applies_to_regions,
        is_enabled,
        priority
    ) VALUES (
        dev_tenant_id,
        'Política para Profissionais de Saúde',
        'Política específica para acesso aos sistemas de saúde conformes com HIPAA/GDPR/LGPD',
        'CONDITIONAL',
        '{
            "base_requirements": {
                "required_factors": 2,
                "allowed_methods": ["KB-01-02", "PB-01-01", "PB-01-02", "PB-03-01", "HC-01-01"],
                "minimum_factor_strength": "ADVANCED"
            },
            "emergency_access": {
                "enabled": true,
                "requires_attestation": true,
                "limited_access_minutes": 60,
                "requires_post_attestation": true,
                "notification_targets": ["supervisor", "compliance", "audit"]
            },
            "context_rules": {
                "hospital_network": {
                    "required_factors": 1,
                    "session_extension_allowed": true
                },
                "sensitive_data_access": {
                    "required_factors": 2,
                    "minimum_factor_strength": "VERY_ADVANCED",
                    "allowed_methods": ["PB-01-02", "IN-01-01", "HC-01-01"]
                }
            }
        }'::JSONB,
        ARRAY['HUMAN'],
        ARRAY['HEALTHCARE_PROFESSIONAL'],
        NULL,
        TRUE,
        60
    );

    -- Política para Open Banking
    INSERT INTO iam_core.authentication_policies (
        tenant_id,
        policy_name,
        description,
        policy_type,
        policy_rules,
        applies_to_user_types,
        applies_to_security_profiles,
        applies_to_regions,
        is_enabled,
        priority
    ) VALUES (
        dev_tenant_id,
        'Política para Open Banking',
        'Política que implementa requisitos PSD2 para Strong Customer Authentication',
        'CONDITIONAL',
        '{
            "base_requirements": {
                "required_factors": 2,
                "mandatory_factor_categories": ["KNOWLEDGE", "POSSESSION"],
                "allowed_methods": ["KB-01-02", "PB-01-01", "PB-01-03", "PB-03-01", "OB-01-01", "OB-01-02"],
                "minimum_factor_strength": "ADVANCED"
            },
            "exemptions": {
                "low_value_payment": {
                    "threshold_amount": 30,
                    "cumulative_limit": 100,
                    "consecutive_transactions": 5
                },
                "trusted_beneficiary": {
                    "requires_initial_sca": true,
                    "trust_period_days": 90
                },
                "transaction_risk_analysis": {
                    "fraud_rate_threshold": 0.013,
                    "amount_thresholds": {
                        "remote_electronic": 500,
                        "contactless": 50
                    }
                }
            },
            "psd2_compliant": true,
            "region_specific_rules": {
                "EU": {
                    "require_dynamic_linking": true,
                    "require_dedicated_interface": true
                }
            }
        }'::JSONB,
        ARRAY['HUMAN'],
        ARRAY['FINTECH', 'BANKING'],
        ARRAY['PT', 'EU'],
        TRUE,
        40
    );
END $$;

-- ==========================================================================
-- Funções Auxiliares para Validação de Políticas
-- ==========================================================================

CREATE OR REPLACE FUNCTION iam_core.validate_authentication_policy(
    policy_id UUID
) RETURNS BOOLEAN AS $$
DECLARE
    policy_record iam_core.authentication_policies%ROWTYPE;
    policy_rules JSONB;
    validation_errors TEXT[];
BEGIN
    -- Recuperar a política
    SELECT * INTO policy_record FROM iam_core.authentication_policies WHERE policy_id = validate_authentication_policy.policy_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Política não encontrada: %', policy_id;
    END IF;
    
    policy_rules := policy_record.policy_rules;
    validation_errors := ARRAY[]::TEXT[];
    
    -- Validações gerais para todos os tipos de políticas
    IF policy_rules IS NULL OR policy_rules = '{}'::JSONB THEN
        validation_errors := array_append(validation_errors, 'Regras de política vazias');
    END IF;
    
    -- Validações específicas por tipo de política
    CASE policy_record.policy_type
        WHEN 'MFA' THEN
            IF NOT policy_rules ? 'required_factors' THEN
                validation_errors := array_append(validation_errors, 'MFA: Falta definição de required_factors');
            END IF;
            
            IF NOT policy_rules ? 'allowed_methods' THEN
                validation_errors := array_append(validation_errors, 'MFA: Falta definição de allowed_methods');
            END IF;
            
        WHEN 'ADAPTIVE' THEN
            IF NOT policy_rules ? 'risk_factors' THEN
                validation_errors := array_append(validation_errors, 'ADAPTIVE: Falta definição de risk_factors');
            END IF;
            
            IF NOT policy_rules ? 'risk_thresholds' THEN
                validation_errors := array_append(validation_errors, 'ADAPTIVE: Falta definição de risk_thresholds');
            END IF;
            
            IF NOT policy_rules ? 'actions' THEN
                validation_errors := array_append(validation_errors, 'ADAPTIVE: Falta definição de actions');
            END IF;
            
        WHEN 'RISK_BASED' THEN
            IF NOT policy_rules ? 'risk_levels' THEN
                validation_errors := array_append(validation_errors, 'RISK_BASED: Falta definição de risk_levels');
            END IF;
            
        WHEN 'STEP_UP' THEN
            IF NOT policy_rules ? 'high_risk_operations' THEN
                validation_errors := array_append(validation_errors, 'STEP_UP: Falta definição de high_risk_operations');
            END IF;
            
            IF NOT policy_rules ? 'required_factors' THEN
                validation_errors := array_append(validation_errors, 'STEP_UP: Falta definição de required_factors');
            END IF;
            
        WHEN 'CONDITIONAL' THEN
            IF NOT policy_rules ? 'base_requirements' THEN
                validation_errors := array_append(validation_errors, 'CONDITIONAL: Falta definição de base_requirements');
            END IF;
            
        ELSE
            validation_errors := array_append(validation_errors, 'Tipo de política desconhecido: ' || policy_record.policy_type);
    END CASE;
    
    -- Retornar resultado da validação
    IF array_length(validation_errors, 1) > 0 THEN
        RAISE NOTICE 'Erros de validação na política %: %', policy_id, array_to_string(validation_errors, ', ');
        RETURN FALSE;
    ELSE
        RETURN TRUE;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ==========================================================================
-- Triggers para Validação
-- ==========================================================================

CREATE OR REPLACE FUNCTION iam_core.validate_policy_on_insert()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM iam_core.validate_authentication_policy(NEW.policy_id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER validate_policy_insert
AFTER INSERT ON iam_core.authentication_policies
FOR EACH ROW EXECUTE FUNCTION iam_core.validate_policy_on_insert();

CREATE OR REPLACE FUNCTION iam_core.validate_policy_on_update()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM iam_core.validate_authentication_policy(NEW.policy_id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER validate_policy_update
AFTER UPDATE ON iam_core.authentication_policies
FOR EACH ROW EXECUTE FUNCTION iam_core.validate_policy_on_update();

-- ==========================================================================
-- Fim do Script
-- ==========================================================================
