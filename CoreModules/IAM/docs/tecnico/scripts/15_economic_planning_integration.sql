-- ============================================================================
-- Script:      15_economic_planning_integration.sql
-- Autor:       Eduardo Jeremias
-- Projeto:     INNOVABIZ - Suíte de Sistema de Governança Inteligente Empresarial
-- Data:        15/05/2025
-- Descrição:   Integração entre Validadores de Conformidade IAM e
--              Sistema de Gestão de Planejamento e Modelagem Econômica
-- ============================================================================

-- Criação do schema para a integração econômica se não existir
CREATE SCHEMA IF NOT EXISTS economic_planning;

-- Comentário no schema
COMMENT ON SCHEMA economic_planning IS 'Schema para integração entre validadores de conformidade IAM e sistema de planejamento econômico';

-- Configuração de permissões
GRANT USAGE ON SCHEMA economic_planning TO economic_analyst_role, compliance_manager_role;

-- ============================================================================
-- Tabelas de Integração
-- ============================================================================

-- Mapeamento entre requisitos de conformidade e variáveis de modelos econômicos
CREATE TABLE IF NOT EXISTS economic_planning.compliance_model_mappings (
    mapping_id SERIAL PRIMARY KEY,
    validator_id VARCHAR(100) NOT NULL,
    validator_name VARCHAR(255) NOT NULL,
    model_variable_id VARCHAR(100) NOT NULL,
    model_variable_name VARCHAR(255) NOT NULL,
    model_id VARCHAR(100) NOT NULL,
    impact_direction VARCHAR(20) CHECK (impact_direction IN ('POSITIVE', 'NEGATIVE', 'NEUTRAL')),
    impact_weight NUMERIC(10,4),
    mapping_justification TEXT,
    compliance_category VARCHAR(100),
    regulatory_reference JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    tenant_id VARCHAR(100) NOT NULL
);

-- Índices para a tabela de mapeamentos
CREATE INDEX IF NOT EXISTS idx_compliance_model_mappings_validator ON economic_planning.compliance_model_mappings(validator_id, tenant_id);
CREATE INDEX IF NOT EXISTS idx_compliance_model_mappings_model ON economic_planning.compliance_model_mappings(model_id, tenant_id);

-- Comentário na tabela
COMMENT ON TABLE economic_planning.compliance_model_mappings IS 'Mapeamento entre validadores de conformidade IAM e variáveis em modelos econômicos';

-- Fatores de impacto econômico para não conformidades
CREATE TABLE IF NOT EXISTS economic_planning.compliance_impact_factors (
    factor_id SERIAL PRIMARY KEY,
    validator_id VARCHAR(100) NOT NULL,
    jurisdiction VARCHAR(50) NOT NULL,
    business_sector VARCHAR(100) NOT NULL,
    direct_cost_factor NUMERIC(10,4),
    indirect_cost_factor NUMERIC(10,4),
    reputational_impact_factor NUMERIC(10,4),
    operational_impact_factor NUMERIC(10,4),
    regulatory_penalty_base NUMERIC(15,2),
    regulatory_penalty_currency VARCHAR(3),
    probability_adjustment NUMERIC(5,4),
    impact_calculation_formula TEXT,
    scenario_parameters JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    tenant_id VARCHAR(100) NOT NULL
);

-- Índices para a tabela de fatores de impacto
CREATE INDEX IF NOT EXISTS idx_compliance_impact_factors_validator ON economic_planning.compliance_impact_factors(validator_id, tenant_id);
CREATE INDEX IF NOT EXISTS idx_compliance_impact_factors_jurisdiction ON economic_planning.compliance_impact_factors(jurisdiction, tenant_id);

-- Comentário na tabela
COMMENT ON TABLE economic_planning.compliance_impact_factors IS 'Fatores de impacto econômico para diferentes tipos de não conformidades IAM';

-- Cenários econômicos para simulação
CREATE TABLE IF NOT EXISTS economic_planning.economic_scenarios (
    scenario_id VARCHAR(100) PRIMARY KEY,
    scenario_name VARCHAR(255) NOT NULL,
    scenario_description TEXT,
    base_parameters JSONB NOT NULL,
    compliance_parameters JSONB,
    is_baseline BOOLEAN DEFAULT FALSE,
    scenario_type VARCHAR(50) NOT NULL,
    time_horizon INTEGER NOT NULL,
    confidence_level NUMERIC(5,4),
    created_by VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    tenant_id VARCHAR(100) NOT NULL
);

-- Índices para a tabela de cenários econômicos
CREATE INDEX IF NOT EXISTS idx_economic_scenarios_type ON economic_planning.economic_scenarios(scenario_type, tenant_id);
CREATE INDEX IF NOT EXISTS idx_economic_scenarios_baseline ON economic_planning.economic_scenarios(is_baseline, tenant_id);

-- Comentário na tabela
COMMENT ON TABLE economic_planning.economic_scenarios IS 'Cenários econômicos para simulação de impactos de conformidade IAM';

-- Custos regulatórios por jurisdição e tipo de violação
CREATE TABLE IF NOT EXISTS economic_planning.regulatory_cost_factors (
    cost_factor_id SERIAL PRIMARY KEY,
    jurisdiction VARCHAR(50) NOT NULL,
    regulatory_framework VARCHAR(100) NOT NULL,
    violation_type VARCHAR(100) NOT NULL,
    base_penalty NUMERIC(15,2),
    currency_code VARCHAR(3) NOT NULL,
    calculation_method VARCHAR(50),
    scaling_factor NUMERIC(10,4),
    aggravating_factors JSONB,
    mitigating_factors JSONB,
    reference_legal_text TEXT,
    effective_from DATE NOT NULL,
    effective_to DATE,
    last_updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    tenant_id VARCHAR(100) NOT NULL
);

-- Índices para a tabela de custos regulatórios
CREATE INDEX IF NOT EXISTS idx_regulatory_cost_factors_jurisdiction ON economic_planning.regulatory_cost_factors(jurisdiction, tenant_id);
CREATE INDEX IF NOT EXISTS idx_regulatory_cost_factors_framework ON economic_planning.regulatory_cost_factors(regulatory_framework, tenant_id);

-- Comentário na tabela
COMMENT ON TABLE economic_planning.regulatory_cost_factors IS 'Fatores de custo regulatório por jurisdição e tipo de violação de conformidade';

-- Resultados de validação de modelos econômicos
CREATE TABLE IF NOT EXISTS economic_planning.model_validation_results (
    validation_id SERIAL PRIMARY KEY,
    model_id VARCHAR(100) NOT NULL,
    model_name VARCHAR(255) NOT NULL,
    validator_id VARCHAR(100) NOT NULL,
    validation_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_compliant BOOLEAN NOT NULL,
    compliance_score NUMERIC(5,2),
    issues_found JSONB,
    validation_details TEXT,
    recommendations TEXT,
    risk_level VARCHAR(20),
    validation_user VARCHAR(100),
    tenant_id VARCHAR(100) NOT NULL
);

-- Índices para a tabela de resultados de validação
CREATE INDEX IF NOT EXISTS idx_model_validation_results_model ON economic_planning.model_validation_results(model_id, tenant_id);
CREATE INDEX IF NOT EXISTS idx_model_validation_results_validator ON economic_planning.model_validation_results(validator_id, tenant_id);

-- Comentário na tabela
COMMENT ON TABLE economic_planning.model_validation_results IS 'Resultados de validação de conformidade de modelos econômicos';

-- Histórico de execuções de cenários incluindo fatores de conformidade
CREATE TABLE IF NOT EXISTS economic_planning.compliance_scenario_runs (
    run_id SERIAL PRIMARY KEY,
    scenario_id VARCHAR(100) NOT NULL REFERENCES economic_planning.economic_scenarios(scenario_id),
    run_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    run_parameters JSONB NOT NULL,
    compliance_settings JSONB,
    run_results JSONB,
    execution_time NUMERIC(10,2),
    status VARCHAR(50) NOT NULL,
    error_details TEXT,
    run_by VARCHAR(100),
    tenant_id VARCHAR(100) NOT NULL
);

-- Índices para a tabela de execuções de cenários
CREATE INDEX IF NOT EXISTS idx_compliance_scenario_runs_scenario ON economic_planning.compliance_scenario_runs(scenario_id, tenant_id);
CREATE INDEX IF NOT EXISTS idx_compliance_scenario_runs_timestamp ON economic_planning.compliance_scenario_runs(run_timestamp, tenant_id);

-- Comentário na tabela
COMMENT ON TABLE economic_planning.compliance_scenario_runs IS 'Histórico de execuções de cenários econômicos com fatores de conformidade';

-- ============================================================================
-- Funções: Mapeamento e Registro
-- ============================================================================

-- Função para registrar mapeamento entre requisito de conformidade e variável econômica
CREATE OR REPLACE FUNCTION economic_planning.register_compliance_model_mapping(
    p_validator_id VARCHAR(100),
    p_validator_name VARCHAR(255),
    p_model_variable_id VARCHAR(100),
    p_model_variable_name VARCHAR(255),
    p_model_id VARCHAR(100),
    p_impact_direction VARCHAR(20),
    p_impact_weight NUMERIC(10,4),
    p_mapping_justification TEXT,
    p_compliance_category VARCHAR(100),
    p_regulatory_reference JSONB,
    p_tenant_id VARCHAR(100)
) RETURNS INTEGER AS $$
DECLARE
    v_mapping_id INTEGER;
BEGIN
    -- Inserir ou atualizar o mapeamento
    INSERT INTO economic_planning.compliance_model_mappings (
        validator_id, validator_name,
        model_variable_id, model_variable_name,
        model_id, impact_direction, impact_weight,
        mapping_justification, compliance_category,
        regulatory_reference, tenant_id
    ) VALUES (
        p_validator_id, p_validator_name,
        p_model_variable_id, p_model_variable_name,
        p_model_id, p_impact_direction, p_impact_weight,
        p_mapping_justification, p_compliance_category,
        p_regulatory_reference, p_tenant_id
    )
    ON CONFLICT (mapping_id) 
    WHERE tenant_id = p_tenant_id
    DO UPDATE SET
        validator_name = p_validator_name,
        model_variable_name = p_model_variable_name,
        impact_direction = p_impact_direction,
        impact_weight = p_impact_weight,
        mapping_justification = p_mapping_justification,
        compliance_category = p_compliance_category,
        regulatory_reference = p_regulatory_reference,
        updated_at = CURRENT_TIMESTAMP
    RETURNING mapping_id INTO v_mapping_id;

    -- Registrar o evento de mapeamento
    INSERT INTO audit.audit_log (
        event_type, object_type, object_id, 
        action, description, user_id, tenant_id
    ) VALUES (
        'COMPLIANCE_MODEL_MAPPING', 'MODEL_MAPPING', v_mapping_id::TEXT,
        'REGISTER', 'Registro de mapeamento entre validator ' || p_validator_id || 
                    ' e variável de modelo ' || p_model_variable_id,
        current_user, p_tenant_id
    );

    RETURN v_mapping_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

GRANT EXECUTE ON FUNCTION economic_planning.register_compliance_model_mapping TO economic_analyst_role, compliance_manager_role;

-- ============================================================================
-- Funções: Cálculo de Impacto Econômico
-- ============================================================================

-- Função para calcular o impacto econômico de uma não conformidade
CREATE OR REPLACE FUNCTION economic_planning.calculate_compliance_economic_impact(
    p_validator_id VARCHAR(100),
    p_validation_id VARCHAR(100),
    p_jurisdiction VARCHAR(50),
    p_business_sector VARCHAR(100),
    p_severity_level VARCHAR(20),
    p_tenant_id VARCHAR(100)
) RETURNS JSONB AS $$
DECLARE
    v_result JSONB;
    v_direct_cost NUMERIC(15,2);
    v_indirect_cost NUMERIC(15,2);
    v_regulatory_penalty NUMERIC(15,2);
    v_total_impact NUMERIC(15,2);
    v_severity_multiplier NUMERIC(5,2);
BEGIN
    -- Definir multiplicador de severidade
    CASE p_severity_level
        WHEN 'CRITICAL' THEN v_severity_multiplier := 5.0;
        WHEN 'HIGH' THEN v_severity_multiplier := 3.0;
        WHEN 'MEDIUM' THEN v_severity_multiplier := 1.5;
        WHEN 'LOW' THEN v_severity_multiplier := 0.7;
        ELSE v_severity_multiplier := 1.0;
    END CASE;
    
    -- Cálculos simplificados para demonstração
    v_direct_cost := 10000 * v_severity_multiplier;
    v_indirect_cost := v_direct_cost * 0.5;
    v_regulatory_penalty := 5000 * v_severity_multiplier;
    v_total_impact := v_direct_cost + v_indirect_cost + v_regulatory_penalty;
    
    -- Construir o resultado JSON
    v_result := jsonb_build_object(
        'validation_id', p_validation_id,
        'validator_id', p_validator_id,
        'jurisdiction', p_jurisdiction,
        'business_sector', p_business_sector,
        'severity_level', p_severity_level,
        'calculation_timestamp', CURRENT_TIMESTAMP,
        'impacts', jsonb_build_object(
            'direct_cost', round(v_direct_cost, 2),
            'indirect_cost', round(v_indirect_cost, 2),
            'regulatory_penalty', round(v_regulatory_penalty, 2),
            'total_impact', round(v_total_impact, 2)
        ),
        'currency', 'EUR'
    );
    
    RETURN v_result;
END;
$$ LANGUAGE plpgsql;

GRANT EXECUTE ON FUNCTION economic_planning.calculate_compliance_economic_impact TO economic_analyst_role, compliance_manager_role;

-- ============================================================================
-- Funções: Integração com Validadores IAM
-- ============================================================================

-- Função para integrar resultados de validação IAM com modelos econômicos
CREATE OR REPLACE FUNCTION economic_planning.integrate_compliance_validation_with_model(
    p_validation_id VARCHAR(100),
    p_model_id VARCHAR(100),
    p_tenant_id VARCHAR(100)
) RETURNS JSONB AS $$
DECLARE
    v_validation RECORD;
    v_impact JSONB;
    v_integration_result JSONB;
BEGIN
    -- Obter dados da validação
    SELECT 
        validator_id, 
        validator_name,
        impact_level AS severity_level
    INTO v_validation
    FROM 
        iam_validators.validation_history
    WHERE 
        validation_id = p_validation_id
        AND tenant_id = p_tenant_id;
    
    -- Se não encontrar a validação, retornar erro
    IF v_validation IS NULL THEN
        RETURN jsonb_build_object(
            'status', 'ERROR',
            'message', 'Validação não encontrada',
            'validation_id', p_validation_id
        );
    END IF;
    
    -- Calcular impacto econômico
    v_impact := economic_planning.calculate_compliance_economic_impact(
        v_validation.validator_id,
        p_validation_id,
        'UE', -- Jurisdição padrão
        'FINANCIAL', -- Setor padrão
        v_validation.severity_level,
        p_tenant_id
    );
    
    -- Registrar integração
    INSERT INTO economic_planning.model_compliance_integrations (
        validation_id,
        model_id,
        validator_id,
        economic_impact,
        integration_timestamp,
        tenant_id
    ) VALUES (
        p_validation_id,
        p_model_id,
        v_validation.validator_id,
        v_impact,
        CURRENT_TIMESTAMP,
        p_tenant_id
    );
    
    -- Construir resultado da integração
    v_integration_result := jsonb_build_object(
        'status', 'SUCCESS',
        'validation_id', p_validation_id,
        'model_id', p_model_id,
        'validator_id', v_validation.validator_id,
        'validator_name', v_validation.validator_name,
        'economic_impact', v_impact,
        'integration_timestamp', CURRENT_TIMESTAMP
    );
    
    RETURN v_integration_result;
END;
$$ LANGUAGE plpgsql;

GRANT EXECUTE ON FUNCTION economic_planning.integrate_compliance_validation_with_model TO economic_analyst_role, compliance_manager_role;

-- ============================================================================
-- Funções: Geração de Cenários Econômicos
-- ============================================================================

-- Função para gerar cenários econômicos baseados em histórico de conformidade
CREATE OR REPLACE FUNCTION economic_planning.generate_compliance_economic_scenario(
    p_scenario_name VARCHAR(255),
    p_business_sector VARCHAR(100),
    p_jurisdiction VARCHAR(50),
    p_time_horizon INTEGER,
    p_compliance_level VARCHAR(20),
    p_tenant_id VARCHAR(100)
) RETURNS VARCHAR(100) AS $$
DECLARE
    v_scenario_id VARCHAR(100);
    v_base_parameters JSONB;
    v_compliance_parameters JSONB;
BEGIN
    -- Gerar ID único para o cenário
    v_scenario_id := 'SC_' || p_business_sector || '_' || 
                    to_char(CURRENT_TIMESTAMP, 'YYYYMMDDHH24MISS') || 
                    '_' || floor(random() * 1000)::TEXT;
    
    -- Definir parâmetros base do cenário
    v_base_parameters := jsonb_build_object(
        'business_sector', p_business_sector,
        'jurisdiction', p_jurisdiction,
        'time_horizon', p_time_horizon,
        'interest_rate', 0.03,
        'inflation_rate', 0.02,
        'growth_rate', 0.025
    );
    
    -- Definir parâmetros de conformidade baseados no nível especificado
    CASE p_compliance_level
        WHEN 'HIGH' THEN
            v_compliance_parameters := jsonb_build_object(
                'compliance_failure_rate', 0.05,
                'severity_distribution', jsonb_build_object(
                    'LOW', 0.70,
                    'MEDIUM', 0.25,
                    'HIGH', 0.04,
                    'CRITICAL', 0.01
                ),
                'regulatory_scrutiny', 'LOW',
                'remediation_efficiency', 0.9
            );
        WHEN 'MEDIUM' THEN
            v_compliance_parameters := jsonb_build_object(
                'compliance_failure_rate', 0.15,
                'severity_distribution', jsonb_build_object(
                    'LOW', 0.40,
                    'MEDIUM', 0.40,
                    'HIGH', 0.15,
                    'CRITICAL', 0.05
                ),
                'regulatory_scrutiny', 'MEDIUM',
                'remediation_efficiency', 0.7
            );
        WHEN 'LOW' THEN
            v_compliance_parameters := jsonb_build_object(
                'compliance_failure_rate', 0.30,
                'severity_distribution', jsonb_build_object(
                    'LOW', 0.20,
                    'MEDIUM', 0.35,
                    'HIGH', 0.30,
                    'CRITICAL', 0.15
                ),
                'regulatory_scrutiny', 'HIGH',
                'remediation_efficiency', 0.5
            );
        ELSE
            v_compliance_parameters := jsonb_build_object(
                'compliance_failure_rate', 0.10,
                'severity_distribution', jsonb_build_object(
                    'LOW', 0.50,
                    'MEDIUM', 0.30,
                    'HIGH', 0.15,
                    'CRITICAL', 0.05
                ),
                'regulatory_scrutiny', 'MEDIUM',
                'remediation_efficiency', 0.7
            );
    END CASE;
    
    -- Inserir o cenário no banco de dados
    INSERT INTO economic_planning.economic_scenarios (
        scenario_id,
        scenario_name,
        scenario_description,
        base_parameters,
        compliance_parameters,
        scenario_type,
        time_horizon,
        created_by,
        tenant_id
    ) VALUES (
        v_scenario_id,
        p_scenario_name,
        'Cenário econômico baseado em conformidade para ' || p_business_sector || 
        ' em ' || p_jurisdiction || ' com nível de conformidade ' || p_compliance_level,
        v_base_parameters,
        v_compliance_parameters,
        'COMPLIANCE_BASED',
        p_time_horizon,
        current_user,
        p_tenant_id
    );
    
    RETURN v_scenario_id;
END;
$$ LANGUAGE plpgsql;

GRANT EXECUTE ON FUNCTION economic_planning.generate_compliance_economic_scenario TO economic_analyst_role, compliance_manager_role;

-- ============================================================================
-- Triggers para Automação da Integração
-- ============================================================================

-- Tabela de configuração de integrações automáticas
CREATE TABLE IF NOT EXISTS economic_planning.automation_config (
    config_id SERIAL PRIMARY KEY,
    validator_pattern VARCHAR(100),
    model_pattern VARCHAR(100),
    auto_integrate BOOLEAN DEFAULT FALSE,
    auto_scenario_generate BOOLEAN DEFAULT FALSE,
    integration_parameters JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    tenant_id VARCHAR(100) NOT NULL
);

-- Trigger para atualizar modelos econômicos quando resultados de validação mudam
CREATE OR REPLACE FUNCTION economic_planning.on_validation_result_update()
RETURNS TRIGGER AS $$
DECLARE
    v_config RECORD;
    v_models RECORD;
BEGIN
    -- Verificar se existem configurações para integração automática
    FOR v_config IN (
        SELECT * FROM economic_planning.automation_config
        WHERE validator_pattern LIKE '%' || NEW.validator_id || '%'
        AND auto_integrate = TRUE
        AND tenant_id = NEW.tenant_id
    ) LOOP
        -- Encontrar modelos compatíveis para integração
        FOR v_models IN (
            SELECT model_id FROM economic_planning.compliance_model_mappings
            WHERE validator_id = NEW.validator_id
            AND model_id LIKE v_config.model_pattern
            AND tenant_id = NEW.tenant_id
        ) LOOP
            -- Executar integração
            PERFORM economic_planning.integrate_compliance_validation_with_model(
                NEW.validation_id,
                v_models.model_id,
                NEW.tenant_id
            );
        END LOOP;
    END LOOP;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Criar o trigger na tabela de histórico de validação
DROP TRIGGER IF EXISTS trig_validation_result_update ON iam_validators.validation_history;

CREATE TRIGGER trig_validation_result_update
AFTER INSERT OR UPDATE ON iam_validators.validation_history
FOR EACH ROW
EXECUTE FUNCTION economic_planning.on_validation_result_update();

-- ============================================================================
-- Dados Iniciais e Exemplos
-- ============================================================================

-- Inserir exemplos de mapeamentos para métodos de autenticação importantes
INSERT INTO economic_planning.compliance_model_mappings (
    validator_id, validator_name,
    model_variable_id, model_variable_name,
    model_id, impact_direction, impact_weight,
    compliance_category, regulatory_reference, tenant_id
) VALUES
('AUTH_MFA_VALIDATOR', 'Validador de Autenticação Multifator',
 'FRAUD_RISK', 'Risco de Fraude',
 'FINANCIAL_RISK_MODEL_001', 'NEGATIVE', 1.2,
 'AUTHENTICATION', '{"framework": "PSD2", "article": "97"}', '1'),
 
('AUTH_STRONG_VALIDATOR', 'Validador de Autenticação Forte',
 'OPERATIONAL_RISK', 'Risco Operacional',
 'FINANCIAL_RISK_MODEL_001', 'NEGATIVE', 0.8,
 'AUTHENTICATION', '{"framework": "GDPR", "article": "32"}', '1'),
 
('OPEN_BANKING_AUTH_VALIDATOR', 'Validador de Autorização Open Banking',
 'REGULATORY_RISK', 'Risco Regulatório',
 'OPEN_BANKING_MODEL_001', 'NEGATIVE', 1.5,
 'OPEN_BANKING', '{"framework": "OPEN_BANKING", "standard": "FAPI"}', '1');

-- Inserir exemplos de fatores de impacto para UE/Portugal e Brasil
INSERT INTO economic_planning.compliance_impact_factors (
    validator_id, jurisdiction, business_sector,
    direct_cost_factor, indirect_cost_factor,
    reputational_impact_factor, operational_impact_factor,
    regulatory_penalty_base, regulatory_penalty_currency,
    tenant_id
) VALUES
('AUTH_MFA_VALIDATOR', 'UE', 'FINANCIAL',
 1.2, 0.8, 1.5, 1.0,
 50000, 'EUR', '1'),
 
('AUTH_MFA_VALIDATOR', 'BR', 'FINANCIAL',
 1.0, 0.6, 1.2, 0.8,
 200000, 'BRL', '1'),
 
('OPEN_BANKING_AUTH_VALIDATOR', 'UE', 'FINANCIAL',
 1.5, 1.0, 2.0, 1.2,
 100000, 'EUR', '1'),
 
('OPEN_BANKING_AUTH_VALIDATOR', 'BR', 'FINANCIAL',
 1.2, 0.8, 1.5, 1.0,
 500000, 'BRL', '1');

-- Inserir cenários econômicos de exemplo
SELECT economic_planning.generate_compliance_economic_scenario(
    'Cenário Base UE - Alta Conformidade',
    'FINANCIAL',
    'UE',
    36, -- 3 anos
    'HIGH',
    '1'
);

SELECT economic_planning.generate_compliance_economic_scenario(
    'Cenário Brasil - Média Conformidade',
    'FINANCIAL',
    'BR',
    36, -- 3 anos
    'MEDIUM',
    '1'
);

SELECT economic_planning.generate_compliance_economic_scenario(
    'Cenário Stress UE - Baixa Conformidade',
    'FINANCIAL',
    'UE',
    36, -- 3 anos
    'LOW',
    '1'
);

-- ============================================================================
-- Comentários Finais
-- ============================================================================

COMMENT ON SCHEMA economic_planning IS 'Schema para integração entre validadores de conformidade IAM e sistema de planejamento econômico. Implementado em 15/05/2025.';
