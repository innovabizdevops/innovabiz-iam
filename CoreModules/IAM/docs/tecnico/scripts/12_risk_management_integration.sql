-- =====================================================================
-- INNOVABIZ - Integração com Sistema de Gestão de Riscos Corporativos
-- Versão: 1.0.0
-- Data de Criação: 15/05/2025
-- Autor: INNOVABIZ DevOps
-- Descrição: Integração entre Validadores de Conformidade IAM e
--            Sistema de Gestão de Riscos Corporativos da plataforma
-- =====================================================================

-- Criação de schema para a integração com gestão de riscos
CREATE SCHEMA IF NOT EXISTS compliance_risk;

-- =====================================================================
-- Tabelas de Mapeamento e Configuração
-- =====================================================================

-- Tabela de mapeamento entre validadores e categorias de risco
CREATE TABLE compliance_risk.validator_risk_mapping (
    mapping_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    validator_id VARCHAR(100) NOT NULL,
    validator_name VARCHAR(255) NOT NULL,
    risk_category VARCHAR(100) NOT NULL,
    risk_subcategory VARCHAR(100),
    risk_weight NUMERIC(5,2) DEFAULT 1.0,
    impact_level VARCHAR(20) NOT NULL, -- LOW, MEDIUM, HIGH, CRITICAL
    probability_factor NUMERIC(5,2) DEFAULT 1.0,
    mapping_source VARCHAR(50) DEFAULT 'MANUAL', -- MANUAL, AUTOMATED, AI_SUGGESTED
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(validator_id, risk_category)
);

-- Tabela de configuração de análise de risco por tenant
CREATE TABLE compliance_risk.tenant_risk_config (
    tenant_id UUID PRIMARY KEY,
    risk_appetite VARCHAR(20) NOT NULL DEFAULT 'MODERATE', -- LOW, MODERATE, HIGH
    compliance_threshold NUMERIC(5,2) DEFAULT 80.0, -- porcentagem mínima aceitável
    risk_evaluation_frequency VARCHAR(20) DEFAULT 'MONTHLY', -- DAILY, WEEKLY, MONTHLY, QUARTERLY
    auto_remediation_enabled BOOLEAN DEFAULT FALSE,
    notification_level VARCHAR(20) DEFAULT 'ALL', -- CRITICAL_ONLY, HIGH_AND_ABOVE, ALL
    escalation_policy JSONB,
    custom_weights JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabela de registro de riscos identificados
CREATE TABLE compliance_risk.compliance_risk_register (
    risk_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    validation_id UUID REFERENCES compliance_integrator.validation_history(validation_id),
    risk_category VARCHAR(100) NOT NULL,
    risk_subcategory VARCHAR(100),
    risk_title VARCHAR(255) NOT NULL,
    risk_description TEXT NOT NULL,
    impact_level VARCHAR(20) NOT NULL, -- LOW, MEDIUM, HIGH, CRITICAL
    probability_level VARCHAR(20) NOT NULL, -- LOW, MEDIUM, HIGH, VERY_HIGH
    risk_score NUMERIC(7,2) NOT NULL, -- Pontuação calculada de risco
    inherent_risk_score NUMERIC(7,2) NOT NULL, -- Pontuação antes de controles
    residual_risk_score NUMERIC(7,2) NOT NULL, -- Pontuação após controles
    risk_status VARCHAR(20) DEFAULT 'IDENTIFIED', -- IDENTIFIED, ASSESSED, TREATED, MONITORED, CLOSED
    risk_owner VARCHAR(100),
    detection_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    assessment_date TIMESTAMP WITH TIME ZONE,
    treatment_deadline TIMESTAMP WITH TIME ZONE,
    last_review_date TIMESTAMP WITH TIME ZONE,
    next_review_date TIMESTAMP WITH TIME ZONE,
    non_compliance_details JSONB,
    control_measures JSONB,
    mitigation_actions JSONB,
    related_incidents JSONB, -- Referências a incidentes relacionados
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabela de histórico de tratamento de riscos
CREATE TABLE compliance_risk.risk_treatment_history (
    treatment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    risk_id UUID REFERENCES compliance_risk.compliance_risk_register(risk_id),
    treatment_type VARCHAR(50) NOT NULL, -- ACCEPT, MITIGATE, TRANSFER, AVOID
    treatment_description TEXT NOT NULL,
    treatment_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    performed_by VARCHAR(100),
    previous_risk_score NUMERIC(7,2),
    updated_risk_score NUMERIC(7,2),
    effectiveness_rating VARCHAR(20), -- NOT_EFFECTIVE, PARTIALLY_EFFECTIVE, EFFECTIVE, HIGHLY_EFFECTIVE
    supporting_documents JSONB,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================================
-- Funções de Configuração e Mapeamento
-- =====================================================================

-- Função para registrar mapeamento entre validador e categoria de risco
CREATE OR REPLACE FUNCTION compliance_risk.register_validator_risk_mapping(
    validator_id VARCHAR(100),
    validator_name VARCHAR(255),
    risk_category VARCHAR(100),
    risk_subcategory VARCHAR(100) DEFAULT NULL,
    risk_weight NUMERIC(5,2) DEFAULT 1.0,
    impact_level VARCHAR(20) DEFAULT 'MEDIUM',
    probability_factor NUMERIC(5,2) DEFAULT 1.0,
    mapping_source VARCHAR(50) DEFAULT 'MANUAL'
) RETURNS UUID AS $$
DECLARE
    mapping_id UUID;
BEGIN
    INSERT INTO compliance_risk.validator_risk_mapping (
        validator_id,
        validator_name,
        risk_category,
        risk_subcategory,
        risk_weight,
        impact_level,
        probability_factor,
        mapping_source
    ) VALUES (
        validator_id,
        validator_name,
        risk_category,
        risk_subcategory,
        risk_weight,
        impact_level,
        probability_factor,
        mapping_source
    )
    ON CONFLICT (validator_id, risk_category)
    DO UPDATE SET
        validator_name = EXCLUDED.validator_name,
        risk_subcategory = EXCLUDED.risk_subcategory,
        risk_weight = EXCLUDED.risk_weight,
        impact_level = EXCLUDED.impact_level,
        probability_factor = EXCLUDED.probability_factor,
        mapping_source = EXCLUDED.mapping_source,
        is_active = TRUE,
        updated_at = CURRENT_TIMESTAMP
    RETURNING mapping_id INTO mapping_id;
    
    RETURN mapping_id;
END;
$$ LANGUAGE plpgsql;

-- Função para configurar análise de risco para tenant
CREATE OR REPLACE FUNCTION compliance_risk.configure_tenant_risk_analysis(
    tenant_id UUID,
    risk_appetite VARCHAR(20) DEFAULT 'MODERATE',
    compliance_threshold NUMERIC(5,2) DEFAULT 80.0,
    risk_evaluation_frequency VARCHAR(20) DEFAULT 'MONTHLY',
    auto_remediation_enabled BOOLEAN DEFAULT FALSE,
    notification_level VARCHAR(20) DEFAULT 'ALL',
    escalation_policy JSONB DEFAULT NULL,
    custom_weights JSONB DEFAULT NULL
) RETURNS BOOLEAN AS $$
BEGIN
    INSERT INTO compliance_risk.tenant_risk_config (
        tenant_id,
        risk_appetite,
        compliance_threshold,
        risk_evaluation_frequency,
        auto_remediation_enabled,
        notification_level,
        escalation_policy,
        custom_weights
    ) VALUES (
        tenant_id,
        risk_appetite,
        compliance_threshold,
        risk_evaluation_frequency,
        auto_remediation_enabled,
        notification_level,
        COALESCE(escalation_policy, '{}'::JSONB),
        COALESCE(custom_weights, '{}'::JSONB)
    )
    ON CONFLICT (tenant_id)
    DO UPDATE SET
        risk_appetite = EXCLUDED.risk_appetite,
        compliance_threshold = EXCLUDED.compliance_threshold,
        risk_evaluation_frequency = EXCLUDED.risk_evaluation_frequency,
        auto_remediation_enabled = EXCLUDED.auto_remediation_enabled,
        notification_level = EXCLUDED.notification_level,
        escalation_policy = EXCLUDED.escalation_policy,
        custom_weights = EXCLUDED.custom_weights,
        updated_at = CURRENT_TIMESTAMP;
        
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Funções de Análise de Risco e Integração
-- =====================================================================

-- Função para calcular nível de probabilidade
CREATE OR REPLACE FUNCTION compliance_risk.calculate_probability_level(
    compliance_score NUMERIC,
    validator_weight NUMERIC DEFAULT 1.0
) RETURNS VARCHAR(20) AS $$
DECLARE
    adjusted_score NUMERIC;
BEGIN
    adjusted_score := compliance_score * validator_weight;
    
    RETURN CASE 
        WHEN adjusted_score >= 90 THEN 'LOW'
        WHEN adjusted_score >= 75 THEN 'MEDIUM'
        WHEN adjusted_score >= 50 THEN 'HIGH'
        ELSE 'VERY_HIGH'
    END;
END;
$$ LANGUAGE plpgsql;

-- Função para calcular score de risco
CREATE OR REPLACE FUNCTION compliance_risk.calculate_risk_score(
    impact_level VARCHAR(20),
    probability_level VARCHAR(20)
) RETURNS NUMERIC(7,2) AS $$
DECLARE
    impact_value NUMERIC;
    probability_value NUMERIC;
BEGIN
    -- Valores de impacto
    impact_value := CASE impact_level
        WHEN 'LOW' THEN 1.0
        WHEN 'MEDIUM' THEN 2.0
        WHEN 'HIGH' THEN 3.0
        WHEN 'CRITICAL' THEN 4.0
        ELSE 1.0
    END;
    
    -- Valores de probabilidade
    probability_value := CASE probability_level
        WHEN 'LOW' THEN 1.0
        WHEN 'MEDIUM' THEN 2.0
        WHEN 'HIGH' THEN 3.0
        WHEN 'VERY_HIGH' THEN 4.0
        ELSE 1.0
    END;
    
    -- Fórmula de cálculo: Impact * Probability * 6.25 (para escala de 100)
    RETURN impact_value * probability_value * 6.25;
END;
$$ LANGUAGE plpgsql;

-- Função principal para análise de risco a partir de resultado de validação
CREATE OR REPLACE FUNCTION compliance_risk.analyze_compliance_risks(
    validation_id UUID,
    tenant_id UUID
) RETURNS SETOF UUID AS $$
DECLARE
    validation_result RECORD;
    validator_mapping RECORD;
    tenant_config RECORD;
    requirement RECORD;
    risk_id UUID;
    probability_level VARCHAR(20);
    risk_score NUMERIC(7,2);
    residual_score NUMERIC(7,2);
BEGIN
    -- Obter dados da validação
    SELECT * INTO validation_result
    FROM compliance_integrator.validation_history
    WHERE validation_id = analyze_compliance_risks.validation_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Validação % não encontrada', validation_id;
    END IF;
    
    -- Obter configuração do tenant
    SELECT * INTO tenant_config
    FROM compliance_risk.tenant_risk_config
    WHERE tenant_id = analyze_compliance_risks.tenant_id;
    
    IF NOT FOUND THEN
        -- Criar configuração padrão se não existir
        PERFORM compliance_risk.configure_tenant_risk_analysis(
            analyze_compliance_risks.tenant_id
        );
        
        SELECT * INTO tenant_config
        FROM compliance_risk.tenant_risk_config
        WHERE tenant_id = analyze_compliance_risks.tenant_id;
    END IF;
    
    -- Processar cada requisito não conforme na validação
    FOR requirement IN 
        SELECT * FROM jsonb_to_recordset(validation_result.results->'requirements')
        AS x(requirement_id TEXT, requirement_name TEXT, is_compliant BOOLEAN, details TEXT)
        WHERE NOT is_compliant
    LOOP
        -- Buscar mapeamento de risco para este validador
        SELECT * INTO validator_mapping
        FROM compliance_risk.validator_risk_mapping
        WHERE validator_id = validation_result.validator_id
        AND is_active = TRUE
        LIMIT 1;
        
        IF FOUND THEN
            -- Calcular probabilidade e score de risco
            probability_level := compliance_risk.calculate_probability_level(
                (validation_result.compliance_score).value,
                validator_mapping.probability_factor
            );
            
            risk_score := compliance_risk.calculate_risk_score(
                validator_mapping.impact_level,
                probability_level
            );
            
            -- Calcular risco residual (após controles existentes)
            -- Para simplificar, estamos considerando uma redução de 30%
            residual_score := risk_score * 0.7;
            
            -- Registrar risco identificado
            INSERT INTO compliance_risk.compliance_risk_register (
                tenant_id,
                validation_id,
                risk_category,
                risk_subcategory,
                risk_title,
                risk_description,
                impact_level,
                probability_level,
                risk_score,
                inherent_risk_score,
                residual_risk_score,
                non_compliance_details,
                detection_date
            ) VALUES (
                analyze_compliance_risks.tenant_id,
                analyze_compliance_risks.validation_id,
                validator_mapping.risk_category,
                validator_mapping.risk_subcategory,
                'Risco de não conformidade: ' || requirement.requirement_name,
                'Não conformidade identificada na validação de ' || validation_result.validator_name ||
                ': ' || requirement.details,
                validator_mapping.impact_level,
                probability_level,
                risk_score,
                risk_score,
                residual_score,
                jsonb_build_object(
                    'requirement_id', requirement.requirement_id,
                    'requirement_name', requirement.requirement_name,
                    'details', requirement.details,
                    'validator_id', validation_result.validator_id,
                    'validator_name', validation_result.validator_name,
                    'validation_date', validation_result.validation_date
                ),
                CURRENT_TIMESTAMP
            ) RETURNING risk_id INTO risk_id;
            
            -- Retornar ID do risco criado
            RETURN NEXT risk_id;
        END IF;
    END LOOP;
    
    RETURN;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Trigger para Processamento Automático
-- =====================================================================

-- Função de trigger para análise automática de riscos após validação
CREATE OR REPLACE FUNCTION compliance_risk.validation_risk_trigger()
RETURNS TRIGGER AS $$
DECLARE
    tenant_config RECORD;
    risk_ids UUID[];
BEGIN
    -- Verificar se existe configuração de análise de risco para o tenant
    SELECT * INTO tenant_config
    FROM compliance_risk.tenant_risk_config
    WHERE tenant_id = NEW.tenant_id;
    
    IF FOUND THEN
        -- Executar análise de risco automaticamente
        SELECT array_agg(risk_id) INTO risk_ids
        FROM compliance_risk.analyze_compliance_risks(
            NEW.validation_id,
            NEW.tenant_id
        ) AS risk_id;
        
        -- Atualizar validação com referência aos riscos identificados
        UPDATE compliance_integrator.validation_history
        SET metadata = COALESCE(metadata, '{}'::JSONB) || 
                       jsonb_build_object('identified_risks', risk_ids)
        WHERE validation_id = NEW.validation_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Criação do trigger
DROP TRIGGER IF EXISTS validation_risk_trigger 
ON compliance_integrator.validation_history;

CREATE TRIGGER validation_risk_trigger
AFTER INSERT ON compliance_integrator.validation_history
FOR EACH ROW EXECUTE FUNCTION compliance_risk.validation_risk_trigger();

-- =====================================================================
-- Funções de Consulta e Relatório
-- =====================================================================

-- Função para consultar riscos por tenant
CREATE OR REPLACE FUNCTION compliance_risk.get_tenant_risks(
    tenant_id UUID,
    risk_status VARCHAR(20) DEFAULT NULL,
    risk_category VARCHAR(100) DEFAULT NULL,
    min_risk_score NUMERIC DEFAULT NULL,
    max_risk_score NUMERIC DEFAULT NULL,
    start_date TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    end_date TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    limit_count INTEGER DEFAULT 100
) RETURNS TABLE (
    risk_id UUID,
    risk_title VARCHAR(255),
    risk_category VARCHAR(100),
    risk_subcategory VARCHAR(100),
    impact_level VARCHAR(20),
    probability_level VARCHAR(20),
    risk_score NUMERIC(7,2),
    residual_risk_score NUMERIC(7,2),
    risk_status VARCHAR(20),
    detection_date TIMESTAMP WITH TIME ZONE,
    treatment_deadline TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        r.risk_id,
        r.risk_title,
        r.risk_category,
        r.risk_subcategory,
        r.impact_level,
        r.probability_level,
        r.risk_score,
        r.residual_risk_score,
        r.risk_status,
        r.detection_date,
        r.treatment_deadline
    FROM 
        compliance_risk.compliance_risk_register r
    WHERE 
        r.tenant_id = get_tenant_risks.tenant_id
        AND (get_tenant_risks.risk_status IS NULL OR r.risk_status = get_tenant_risks.risk_status)
        AND (get_tenant_risks.risk_category IS NULL OR r.risk_category = get_tenant_risks.risk_category)
        AND (get_tenant_risks.min_risk_score IS NULL OR r.risk_score >= get_tenant_risks.min_risk_score)
        AND (get_tenant_risks.max_risk_score IS NULL OR r.risk_score <= get_tenant_risks.max_risk_score)
        AND (get_tenant_risks.start_date IS NULL OR r.detection_date >= get_tenant_risks.start_date)
        AND (get_tenant_risks.end_date IS NULL OR r.detection_date <= get_tenant_risks.end_date)
    ORDER BY 
        r.risk_score DESC,
        r.detection_date DESC
    LIMIT 
        get_tenant_risks.limit_count;
END;
$$ LANGUAGE plpgsql;

-- Função para resumo de risco por categoria
CREATE OR REPLACE FUNCTION compliance_risk.get_risk_summary_by_category(
    tenant_id UUID
) RETURNS TABLE (
    risk_category VARCHAR(100),
    risk_count INTEGER,
    avg_risk_score NUMERIC(7,2),
    max_risk_score NUMERIC(7,2),
    critical_count INTEGER,
    high_count INTEGER,
    medium_count INTEGER,
    low_count INTEGER,
    treated_count INTEGER,
    open_count INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        r.risk_category,
        COUNT(r.risk_id)::INTEGER AS risk_count,
        AVG(r.risk_score)::NUMERIC(7,2) AS avg_risk_score,
        MAX(r.risk_score)::NUMERIC(7,2) AS max_risk_score,
        COUNT(CASE WHEN r.impact_level = 'CRITICAL' THEN 1 END)::INTEGER AS critical_count,
        COUNT(CASE WHEN r.impact_level = 'HIGH' THEN 1 END)::INTEGER AS high_count,
        COUNT(CASE WHEN r.impact_level = 'MEDIUM' THEN 1 END)::INTEGER AS medium_count,
        COUNT(CASE WHEN r.impact_level = 'LOW' THEN 1 END)::INTEGER AS low_count,
        COUNT(CASE WHEN r.risk_status IN ('TREATED', 'CLOSED') THEN 1 END)::INTEGER AS treated_count,
        COUNT(CASE WHEN r.risk_status NOT IN ('TREATED', 'CLOSED') THEN 1 END)::INTEGER AS open_count
    FROM 
        compliance_risk.compliance_risk_register r
    WHERE 
        r.tenant_id = get_risk_summary_by_category.tenant_id
    GROUP BY 
        r.risk_category
    ORDER BY 
        max_risk_score DESC,
        avg_risk_score DESC;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Mapeamentos Padrão para Setores
-- =====================================================================

-- Inserção de mapeamentos padrão entre validadores e categorias de risco
DO $$
BEGIN
    -- Mapeamentos para o setor de saúde
    PERFORM compliance_risk.register_validator_risk_mapping(
        'hipaa_compliance', 'Validador HIPAA', 
        'DATA_PRIVACY', 'HEALTH_INFORMATION_PROTECTION', 
        1.2, 'HIGH', 1.1, 'AUTOMATED'
    );
    
    PERFORM compliance_risk.register_validator_risk_mapping(
        'gdpr_healthcare', 'Validador GDPR para Saúde', 
        'REGULATORY_COMPLIANCE', 'GDPR', 
        1.3, 'HIGH', 1.2, 'AUTOMATED'
    );
    
    PERFORM compliance_risk.register_validator_risk_mapping(
        'lgpd_healthcare', 'Validador LGPD para Saúde', 
        'REGULATORY_COMPLIANCE', 'LGPD', 
        1.1, 'MEDIUM', 1.0, 'AUTOMATED'
    );
    
    -- Mapeamentos para o setor financeiro
    PERFORM compliance_risk.register_validator_risk_mapping(
        'psd2_compliance', 'Validador PSD2', 
        'FINANCIAL_SERVICES_REGULATION', 'PAYMENT_SERVICES', 
        1.4, 'HIGH', 1.2, 'AUTOMATED'
    );
    
    PERFORM compliance_risk.register_validator_risk_mapping(
        'openbanking_compliance', 'Validador Open Banking', 
        'FINANCIAL_SERVICES_REGULATION', 'OPEN_BANKING', 
        1.3, 'HIGH', 1.1, 'AUTOMATED'
    );
    
    -- Mapeamentos para o setor governamental
    PERFORM compliance_risk.register_validator_risk_mapping(
        'eidas_compliance', 'Validador eIDAS', 
        'DIGITAL_IDENTITY', 'ELECTRONIC_IDENTIFICATION', 
        1.2, 'MEDIUM', 1.0, 'AUTOMATED'
    );
    
    PERFORM compliance_risk.register_validator_risk_mapping(
        'icpbrasil_compliance', 'Validador ICP-Brasil', 
        'DIGITAL_IDENTITY', 'DIGITAL_SIGNATURE', 
        1.2, 'MEDIUM', 1.0, 'AUTOMATED'
    );
    
    -- Mapeamentos para AR/VR
    PERFORM compliance_risk.register_validator_risk_mapping(
        'ieee_xr_compliance', 'Validador IEEE XR', 
        'EMERGING_TECHNOLOGY', 'XR_STANDARDS', 
        1.0, 'MEDIUM', 0.9, 'AUTOMATED'
    );
    
    PERFORM compliance_risk.register_validator_risk_mapping(
        'nist_xr_compliance', 'Validador NIST XR', 
        'CYBERSECURITY', 'XR_SECURITY', 
        1.1, 'HIGH', 1.0, 'AUTOMATED'
    );
END;
$$;

-- =====================================================================
-- Comentários de Documentação
-- =====================================================================

COMMENT ON SCHEMA compliance_risk IS 'Schema para integração entre validadores de conformidade e sistema de gestão de riscos';

COMMENT ON TABLE compliance_risk.validator_risk_mapping IS 'Mapeamento entre validadores e categorias de risco';
COMMENT ON TABLE compliance_risk.tenant_risk_config IS 'Configuração de análise de risco por tenant';
COMMENT ON TABLE compliance_risk.compliance_risk_register IS 'Registro de riscos de conformidade identificados';
COMMENT ON TABLE compliance_risk.risk_treatment_history IS 'Histórico de tratamentos aplicados aos riscos';

COMMENT ON FUNCTION compliance_risk.register_validator_risk_mapping IS 'Registra mapeamento entre validador e categoria de risco';
COMMENT ON FUNCTION compliance_risk.configure_tenant_risk_analysis IS 'Configura análise de risco para tenant';
COMMENT ON FUNCTION compliance_risk.calculate_probability_level IS 'Calcula nível de probabilidade baseado em score de conformidade';
COMMENT ON FUNCTION compliance_risk.calculate_risk_score IS 'Calcula score de risco baseado em impacto e probabilidade';
COMMENT ON FUNCTION compliance_risk.analyze_compliance_risks IS 'Analisa riscos de não conformidade a partir de resultado de validação';
COMMENT ON FUNCTION compliance_risk.validation_risk_trigger IS 'Trigger para análise automática de riscos após validação';
COMMENT ON FUNCTION compliance_risk.get_tenant_risks IS 'Consulta riscos por tenant com filtros';
COMMENT ON FUNCTION compliance_risk.get_risk_summary_by_category IS 'Obtém resumo de riscos agrupados por categoria';

-- =====================================================================
-- Fim do Script
-- =====================================================================
