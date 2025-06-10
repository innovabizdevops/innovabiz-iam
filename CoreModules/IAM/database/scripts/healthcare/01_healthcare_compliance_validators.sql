-- INNOVABIZ - IAM Healthcare Compliance Validators
-- Author: Eduardo Jeremias
-- Date: 09/05/2025
-- Version: 1.0
-- Description: Validadores de compliance específicos para integração do IAM com Healthcare

-- Configuração do esquema
SET search_path TO iam, public;

-- Tipos enumerados para healthcare
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'healthcare_data_category') THEN
        CREATE TYPE iam.healthcare_data_category AS ENUM (
            'phi', -- Protected Health Information (HIPAA)
            'pii', -- Personally Identifiable Information
            'payment',
            'clinical',
            'diagnostic',
            'treatment',
            'medication',
            'genomic',
            'research',
            'administrative'
        );
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'healthcare_regulation') THEN
        CREATE TYPE iam.healthcare_regulation AS ENUM (
            'hipaa', -- EUA
            'gdpr_health', -- UE
            'lgpd_health', -- Brasil
            'pndsb', -- Angola
            'phipa', -- Canadá
            'hitech', -- EUA
            'nist_health'
        );
    END IF;
END$$;

-- Tabela de mapeamento entre regulamentações e regiões
CREATE TABLE IF NOT EXISTS iam.healthcare_regulatory_requirements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    regulation iam.healthcare_regulation NOT NULL,
    country_code VARCHAR(3),
    region_code VARCHAR(50),
    data_category iam.healthcare_data_category NOT NULL,
    requirement_code VARCHAR(50) NOT NULL,
    requirement_name VARCHAR(255) NOT NULL,
    requirement_description TEXT,
    requirement_level VARCHAR(50), -- required, recommended, optional
    validation_criteria JSONB NOT NULL, -- critérios específicos para validação
    remediation_steps TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    metadata JSONB DEFAULT '{}'::JSONB,
    UNIQUE(regulation, requirement_code)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_healthcare_regulatory_country_region ON iam.healthcare_regulatory_requirements(country_code, region_code);
CREATE INDEX IF NOT EXISTS idx_healthcare_regulatory_regulation ON iam.healthcare_regulatory_requirements(regulation);
CREATE INDEX IF NOT EXISTS idx_healthcare_regulatory_data_category ON iam.healthcare_regulatory_requirements(data_category);
CREATE INDEX IF NOT EXISTS idx_healthcare_regulatory_is_active ON iam.healthcare_regulatory_requirements(is_active);

-- Tabela de resultados de validação de compliance
CREATE TABLE IF NOT EXISTS iam.healthcare_compliance_validations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    validation_timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    regulation iam.healthcare_regulation NOT NULL,
    validator_name VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL, -- passed, failed, warning, not_applicable
    score INTEGER, -- 0-100
    details JSONB NOT NULL,
    remediation_plan TEXT,
    validated_by UUID REFERENCES iam.users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::JSONB
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_healthcare_compliance_organization_id ON iam.healthcare_compliance_validations(organization_id);
CREATE INDEX IF NOT EXISTS idx_healthcare_compliance_regulation ON iam.healthcare_compliance_validations(regulation);
CREATE INDEX IF NOT EXISTS idx_healthcare_compliance_status ON iam.healthcare_compliance_validations(status);
CREATE INDEX IF NOT EXISTS idx_healthcare_compliance_validation_timestamp ON iam.healthcare_compliance_validations(validation_timestamp);

-- Função para registrar um requisito regulatório de Healthcare
CREATE OR REPLACE FUNCTION iam.register_healthcare_requirement(
    p_regulation iam.healthcare_regulation,
    p_country_code VARCHAR,
    p_region_code VARCHAR,
    p_data_category iam.healthcare_data_category,
    p_requirement_code VARCHAR,
    p_requirement_name VARCHAR,
    p_requirement_description TEXT,
    p_requirement_level VARCHAR,
    p_validation_criteria JSONB,
    p_remediation_steps TEXT DEFAULT NULL,
    p_metadata JSONB DEFAULT '{}'::JSONB
) RETURNS UUID AS $$
DECLARE
    requirement_id UUID;
BEGIN
    INSERT INTO iam.healthcare_regulatory_requirements (
        regulation,
        country_code,
        region_code,
        data_category,
        requirement_code,
        requirement_name,
        requirement_description,
        requirement_level,
        validation_criteria,
        remediation_steps,
        metadata
    ) VALUES (
        p_regulation,
        p_country_code,
        p_region_code,
        p_data_category,
        p_requirement_code,
        p_requirement_name,
        p_requirement_description,
        p_requirement_level,
        p_validation_criteria,
        p_remediation_steps,
        p_metadata
    ) RETURNING id INTO requirement_id;
    
    RETURN requirement_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para validar HIPAA - Autenticação e Controle de Acesso
CREATE OR REPLACE FUNCTION iam.validate_hipaa_access_controls(
    p_organization_id UUID,
    p_parameters JSONB DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    validation_result JSONB;
    validation_id UUID;
    mfa_enabled BOOLEAN;
    session_timeout_minutes INTEGER;
    access_review_days INTEGER;
    audit_controls_enabled BOOLEAN;
    emergency_access_configured BOOLEAN;
    phi_data_encryption_enabled BOOLEAN;
BEGIN
    -- Obter configurações atuais da organização
    SELECT 
        (settings->'security'->'mfa'->'required_for_all')::BOOLEAN,
        COALESCE((settings->'security'->'session'->'timeout_minutes')::INTEGER, 0),
        COALESCE((settings->'security'->'access_review'->'interval_days')::INTEGER, 0),
        (settings->'security'->'audit'->'enabled')::BOOLEAN,
        (settings->'security'->'emergency_access'->'configured')::BOOLEAN,
        (settings->'security'->'encryption'->'phi_data')::BOOLEAN
    INTO
        mfa_enabled,
        session_timeout_minutes,
        access_review_days,
        audit_controls_enabled,
        emergency_access_configured,
        phi_data_encryption_enabled
    FROM iam.organizations
    WHERE id = p_organization_id;
    
    -- Avaliar resultados
    validation_result := jsonb_build_object(
        'regulation', 'hipaa',
        'validator', 'access_controls',
        'timestamp', NOW(),
        'checks', jsonb_build_array(
            jsonb_build_object(
                'name', 'mfa_required',
                'requirement', 'A autenticação multifator deve ser obrigatória para acesso a PHI',
                'status', CASE WHEN mfa_enabled THEN 'passed' ELSE 'failed' END,
                'details', 'MFA configurado: ' || mfa_enabled::TEXT
            ),
            jsonb_build_object(
                'name', 'session_timeout',
                'requirement', 'Timeout de sessão deve ser configurado para menos de 30 minutos',
                'status', CASE WHEN session_timeout_minutes > 0 AND session_timeout_minutes <= 30 THEN 'passed' ELSE 'failed' END,
                'details', 'Timeout atual: ' || session_timeout_minutes::TEXT || ' minutos'
            ),
            jsonb_build_object(
                'name', 'access_review',
                'requirement', 'Revisões de acesso devem ocorrer pelo menos a cada 90 dias',
                'status', CASE WHEN access_review_days > 0 AND access_review_days <= 90 THEN 'passed' ELSE 'failed' END,
                'details', 'Intervalo atual: ' || access_review_days::TEXT || ' dias'
            ),
            jsonb_build_object(
                'name', 'audit_controls',
                'requirement', 'Controles de auditoria devem estar habilitados',
                'status', CASE WHEN audit_controls_enabled THEN 'passed' ELSE 'failed' END,
                'details', 'Auditoria configurada: ' || audit_controls_enabled::TEXT
            ),
            jsonb_build_object(
                'name', 'emergency_access',
                'requirement', 'Acesso de emergência deve estar configurado',
                'status', CASE WHEN emergency_access_configured THEN 'passed' ELSE 'failed' END,
                'details', 'Acesso de emergência: ' || emergency_access_configured::TEXT
            ),
            jsonb_build_object(
                'name', 'phi_encryption',
                'requirement', 'Dados PHI devem ser criptografados',
                'status', CASE WHEN phi_data_encryption_enabled THEN 'passed' ELSE 'failed' END,
                'details', 'Criptografia PHI: ' || phi_data_encryption_enabled::TEXT
            )
        )
    );
    
    -- Calcular status geral e pontuação
    DECLARE
        passed_count INTEGER := 0;
        total_count INTEGER := 0;
        overall_status VARCHAR;
        score INTEGER;
    BEGIN
        SELECT 
            COUNT(*) FILTER (WHERE x->>'status' = 'passed'),
            COUNT(*)
        INTO
            passed_count,
            total_count
        FROM jsonb_array_elements(validation_result->'checks') x;
        
        score := (passed_count::FLOAT / total_count::FLOAT * 100)::INTEGER;
        
        IF score >= 90 THEN
            overall_status := 'passed';
        ELSIF score >= 70 THEN
            overall_status := 'warning';
        ELSE
            overall_status := 'failed';
        END IF;
        
        validation_result := jsonb_set(validation_result, '{score}', to_jsonb(score));
        validation_result := jsonb_set(validation_result, '{status}', to_jsonb(overall_status));
    END;
    
    -- Registrar resultado da validação
    INSERT INTO iam.healthcare_compliance_validations (
        organization_id,
        regulation,
        validator_name,
        status,
        score,
        details
    ) VALUES (
        p_organization_id,
        'hipaa',
        'access_controls',
        validation_result->>'status',
        (validation_result->>'score')::INTEGER,
        validation_result
    ) RETURNING id INTO validation_id;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        NULL, -- user_id (será preenchido pelo contexto de execução)
        'compliance'::iam.audit_event_category,
        CASE 
            WHEN validation_result->>'status' = 'passed' THEN 'info'::iam.audit_severity_level
            WHEN validation_result->>'status' = 'warning' THEN 'medium'::iam.audit_severity_level
            ELSE 'high'::iam.audit_severity_level
        END,
        'VALIDATE_HIPAA_ACCESS_CONTROLS',
        'healthcare_compliance',
        validation_id::TEXT,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        validation_result->>'status',
        NULL, -- response_time
        jsonb_build_object(
            'score', validation_result->>'score',
            'regulation', 'hipaa'
        ),
        NULL, -- request_payload
        validation_result, -- response_payload
        ARRAY['compliance', 'healthcare', 'hipaa'], -- compliance_tags
        ARRAY['HIPAA 164.312(a)', 'HIPAA 164.312(d)'], -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN validation_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para validar LGPD - Proteção de Dados de Saúde
CREATE OR REPLACE FUNCTION iam.validate_lgpd_health_data_protection(
    p_organization_id UUID,
    p_parameters JSONB DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    validation_result JSONB;
    validation_id UUID;
    consent_management_enabled BOOLEAN;
    data_processing_registry_enabled BOOLEAN;
    dpo_assigned BOOLEAN;
    right_to_access_implemented BOOLEAN;
    right_to_delete_implemented BOOLEAN;
    data_portability_implemented BOOLEAN;
    specific_consent_for_health_data BOOLEAN;
BEGIN
    -- Obter configurações atuais da organização
    SELECT 
        (settings->'privacy'->'consent_management'->'enabled')::BOOLEAN,
        (settings->'privacy'->'data_processing_registry'->'enabled')::BOOLEAN,
        (settings->'privacy'->'dpo'->'assigned')::BOOLEAN,
        (settings->'privacy'->'data_subject_rights'->'access_implemented')::BOOLEAN,
        (settings->'privacy'->'data_subject_rights'->'deletion_implemented')::BOOLEAN,
        (settings->'privacy'->'data_subject_rights'->'portability_implemented')::BOOLEAN,
        (settings->'privacy'->'health_data'->'specific_consent_enabled')::BOOLEAN
    INTO
        consent_management_enabled,
        data_processing_registry_enabled,
        dpo_assigned,
        right_to_access_implemented,
        right_to_delete_implemented,
        data_portability_implemented,
        specific_consent_for_health_data
    FROM iam.organizations
    WHERE id = p_organization_id;
    
    -- Avaliar resultados
    validation_result := jsonb_build_object(
        'regulation', 'lgpd_health',
        'validator', 'health_data_protection',
        'timestamp', NOW(),
        'checks', jsonb_build_array(
            jsonb_build_object(
                'name', 'consent_management',
                'requirement', 'Gestão de consentimento deve estar habilitada',
                'status', CASE WHEN consent_management_enabled THEN 'passed' ELSE 'failed' END,
                'details', 'Gestão de consentimento: ' || consent_management_enabled::TEXT
            ),
            jsonb_build_object(
                'name', 'data_processing_registry',
                'requirement', 'Registro de operações de tratamento deve estar habilitado',
                'status', CASE WHEN data_processing_registry_enabled THEN 'passed' ELSE 'failed' END,
                'details', 'Registro de tratamento: ' || data_processing_registry_enabled::TEXT
            ),
            jsonb_build_object(
                'name', 'dpo_assigned',
                'requirement', 'Encarregado (DPO) deve estar designado',
                'status', CASE WHEN dpo_assigned THEN 'passed' ELSE 'failed' END,
                'details', 'DPO designado: ' || dpo_assigned::TEXT
            ),
            jsonb_build_object(
                'name', 'right_to_access',
                'requirement', 'Direito de acesso aos dados deve estar implementado',
                'status', CASE WHEN right_to_access_implemented THEN 'passed' ELSE 'failed' END,
                'details', 'Acesso implementado: ' || right_to_access_implemented::TEXT
            ),
            jsonb_build_object(
                'name', 'right_to_delete',
                'requirement', 'Direito de exclusão deve estar implementado',
                'status', CASE WHEN right_to_delete_implemented THEN 'passed' ELSE 'failed' END,
                'details', 'Exclusão implementada: ' || right_to_delete_implemented::TEXT
            ),
            jsonb_build_object(
                'name', 'data_portability',
                'requirement', 'Portabilidade de dados deve estar implementada',
                'status', CASE WHEN data_portability_implemented THEN 'passed' ELSE 'failed' END,
                'details', 'Portabilidade implementada: ' || data_portability_implemented::TEXT
            ),
            jsonb_build_object(
                'name', 'specific_consent_health',
                'requirement', 'Consentimento específico para dados de saúde deve estar habilitado',
                'status', CASE WHEN specific_consent_for_health_data THEN 'passed' ELSE 'failed' END,
                'details', 'Consentimento específico: ' || specific_consent_for_health_data::TEXT
            )
        )
    );
    
    -- Calcular status geral e pontuação
    DECLARE
        passed_count INTEGER := 0;
        total_count INTEGER := 0;
        overall_status VARCHAR;
        score INTEGER;
    BEGIN
        SELECT 
            COUNT(*) FILTER (WHERE x->>'status' = 'passed'),
            COUNT(*)
        INTO
            passed_count,
            total_count
        FROM jsonb_array_elements(validation_result->'checks') x;
        
        score := (passed_count::FLOAT / total_count::FLOAT * 100)::INTEGER;
        
        IF score >= 90 THEN
            overall_status := 'passed';
        ELSIF score >= 70 THEN
            overall_status := 'warning';
        ELSE
            overall_status := 'failed';
        END IF;
        
        validation_result := jsonb_set(validation_result, '{score}', to_jsonb(score));
        validation_result := jsonb_set(validation_result, '{status}', to_jsonb(overall_status));
    END;
    
    -- Registrar resultado da validação
    INSERT INTO iam.healthcare_compliance_validations (
        organization_id,
        regulation,
        validator_name,
        status,
        score,
        details
    ) VALUES (
        p_organization_id,
        'lgpd_health',
        'health_data_protection',
        validation_result->>'status',
        (validation_result->>'score')::INTEGER,
        validation_result
    ) RETURNING id INTO validation_id;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        NULL, -- user_id (será preenchido pelo contexto de execução)
        'compliance'::iam.audit_event_category,
        CASE 
            WHEN validation_result->>'status' = 'passed' THEN 'info'::iam.audit_severity_level
            WHEN validation_result->>'status' = 'warning' THEN 'medium'::iam.audit_severity_level
            ELSE 'high'::iam.audit_severity_level
        END,
        'VALIDATE_LGPD_HEALTH_DATA_PROTECTION',
        'healthcare_compliance',
        validation_id::TEXT,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        validation_result->>'status',
        NULL, -- response_time
        jsonb_build_object(
            'score', validation_result->>'score',
            'regulation', 'lgpd_health'
        ),
        NULL, -- request_payload
        validation_result, -- response_payload
        ARRAY['compliance', 'healthcare', 'lgpd'], -- compliance_tags
        ARRAY['LGPD Art. 5', 'LGPD Art. 11', 'LGPD Art. 41'], -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN validation_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para validar PNDSB (Angola) - Proteção de Dados de Saúde
CREATE OR REPLACE FUNCTION iam.validate_pndsb_health_data_protection(
    p_organization_id UUID,
    p_parameters JSONB DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    validation_result JSONB;
    validation_id UUID;
    patient_consent_enabled BOOLEAN;
    appropriate_security_measures BOOLEAN;
    health_data_classification_implemented BOOLEAN;
    local_data_storage_enabled BOOLEAN;
    cross_border_transfers_managed BOOLEAN;
    health_data_access_restricted BOOLEAN;
    data_retention_policies_defined BOOLEAN;
BEGIN
    -- Obter configurações atuais da organização
    SELECT 
        (settings->'privacy'->'consent_management'->'patient_consent_enabled')::BOOLEAN,
        (settings->'security'->'healthcare'->'appropriate_measures')::BOOLEAN,
        (settings->'security'->'data_classification'->'healthcare_implemented')::BOOLEAN,
        (settings->'privacy'->'data_residency'->'local_storage')::BOOLEAN,
        (settings->'privacy'->'data_transfers'->'cross_border_managed')::BOOLEAN,
        (settings->'security'->'access_control'->'health_data_restricted')::BOOLEAN,
        (settings->'privacy'->'data_retention'->'policies_defined')::BOOLEAN
    INTO
        patient_consent_enabled,
        appropriate_security_measures,
        health_data_classification_implemented,
        local_data_storage_enabled,
        cross_border_transfers_managed,
        health_data_access_restricted,
        data_retention_policies_defined
    FROM iam.organizations
    WHERE id = p_organization_id;
    
    -- Avaliar resultados
    validation_result := jsonb_build_object(
        'regulation', 'pndsb',
        'validator', 'health_data_protection',
        'timestamp', NOW(),
        'checks', jsonb_build_array(
            jsonb_build_object(
                'name', 'patient_consent',
                'requirement', 'Consentimento do paciente deve estar habilitado',
                'status', CASE WHEN patient_consent_enabled THEN 'passed' ELSE 'failed' END,
                'details', 'Consentimento do paciente: ' || patient_consent_enabled::TEXT
            ),
            jsonb_build_object(
                'name', 'security_measures',
                'requirement', 'Medidas de segurança apropriadas devem ser implementadas',
                'status', CASE WHEN appropriate_security_measures THEN 'passed' ELSE 'failed' END,
                'details', 'Medidas de segurança: ' || appropriate_security_measures::TEXT
            ),
            jsonb_build_object(
                'name', 'data_classification',
                'requirement', 'Classificação de dados de saúde deve estar implementada',
                'status', CASE WHEN health_data_classification_implemented THEN 'passed' ELSE 'failed' END,
                'details', 'Classificação implementada: ' || health_data_classification_implemented::TEXT
            ),
            jsonb_build_object(
                'name', 'local_storage',
                'requirement', 'Armazenamento local de dados deve estar habilitado',
                'status', CASE WHEN local_data_storage_enabled THEN 'passed' ELSE 'failed' END,
                'details', 'Armazenamento local: ' || local_data_storage_enabled::TEXT
            ),
            jsonb_build_object(
                'name', 'cross_border_transfers',
                'requirement', 'Transferências transfronteiriças devem ser gerenciadas',
                'status', CASE WHEN cross_border_transfers_managed THEN 'passed' ELSE 'failed' END,
                'details', 'Transferências gerenciadas: ' || cross_border_transfers_managed::TEXT
            ),
            jsonb_build_object(
                'name', 'access_restriction',
                'requirement', 'Acesso a dados de saúde deve ser restrito',
                'status', CASE WHEN health_data_access_restricted THEN 'passed' ELSE 'failed' END,
                'details', 'Acesso restrito: ' || health_data_access_restricted::TEXT
            ),
            jsonb_build_object(
                'name', 'retention_policies',
                'requirement', 'Políticas de retenção de dados devem estar definidas',
                'status', CASE WHEN data_retention_policies_defined THEN 'passed' ELSE 'failed' END,
                'details', 'Políticas definidas: ' || data_retention_policies_defined::TEXT
            )
        )
    );
    
    -- Calcular status geral e pontuação
    DECLARE
        passed_count INTEGER := 0;
        total_count INTEGER := 0;
        overall_status VARCHAR;
        score INTEGER;
    BEGIN
        SELECT 
            COUNT(*) FILTER (WHERE x->>'status' = 'passed'),
            COUNT(*)
        INTO
            passed_count,
            total_count
        FROM jsonb_array_elements(validation_result->'checks') x;
        
        score := (passed_count::FLOAT / total_count::FLOAT * 100)::INTEGER;
        
        IF score >= 90 THEN
            overall_status := 'passed';
        ELSIF score >= 70 THEN
            overall_status := 'warning';
        ELSE
            overall_status := 'failed';
        END IF;
        
        validation_result := jsonb_set(validation_result, '{score}', to_jsonb(score));
        validation_result := jsonb_set(validation_result, '{status}', to_jsonb(overall_status));
    END;
    
    -- Registrar resultado da validação
    INSERT INTO iam.healthcare_compliance_validations (
        organization_id,
        regulation,
        validator_name,
        status,
        score,
        details
    ) VALUES (
        p_organization_id,
        'pndsb',
        'health_data_protection',
        validation_result->>'status',
        (validation_result->>'score')::INTEGER,
        validation_result
    ) RETURNING id INTO validation_id;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        NULL, -- user_id (será preenchido pelo contexto de execução)
        'compliance'::iam.audit_event_category,
        CASE 
            WHEN validation_result->>'status' = 'passed' THEN 'info'::iam.audit_severity_level
            WHEN validation_result->>'status' = 'warning' THEN 'medium'::iam.audit_severity_level
            ELSE 'high'::iam.audit_severity_level
        END,
        'VALIDATE_PNDSB_HEALTH_DATA_PROTECTION',
        'healthcare_compliance',
        validation_id::TEXT,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        validation_result->>'status',
        NULL, -- response_time
        jsonb_build_object(
            'score', validation_result->>'score',
            'regulation', 'pndsb'
        ),
        NULL, -- request_payload
        validation_result, -- response_payload
        ARRAY['compliance', 'healthcare', 'pndsb'], -- compliance_tags
        ARRAY['PNDSB Sec. 3', 'PNDSB Sec. 5', 'PNDSB Sec. 8'], -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN validation_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
