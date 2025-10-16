-- ============================================================================
-- Script:      16_economic_dashboard.sql
-- Autor:       Eduardo Jeremias
-- Projeto:     INNOVABIZ - Suíte de Sistema de Governança Inteligente Empresarial
-- Data:        15/05/2025
-- Descrição:   Dashboard para visualização e análise de impactos econômicos
--              relacionados à conformidade IAM e modelagem econômica
-- ============================================================================

-- Garantir que o schema economic_planning exista
CREATE SCHEMA IF NOT EXISTS economic_planning;

-- Comentário no script
COMMENT ON SCHEMA economic_planning IS 'Schema para dashboard econômico de conformidade IAM';

-- Configuração de permissões
GRANT USAGE ON SCHEMA economic_planning TO economic_analyst_role, compliance_manager_role, risk_analyst_role, dashboard_viewer_role;

-- ============================================================================
-- Views de Resumo e Agregação de Impacto Econômico
-- ============================================================================

-- View para resumo de impacto econômico agregado
CREATE OR REPLACE VIEW economic_planning.vw_compliance_economic_impact_summary AS
SELECT
    EXTRACT(YEAR FROM mci.integration_timestamp) AS ano,
    EXTRACT(MONTH FROM mci.integration_timestamp) AS mes,
    vh.validator_id,
    vh.validator_name,
    vh.jurisdiction,
    vh.impact_level,
    COUNT(mci.validation_id) AS total_validations,
    SUM((mci.economic_impact->'impacts'->>'direct_cost')::NUMERIC) AS total_direct_cost,
    SUM((mci.economic_impact->'impacts'->>'indirect_cost')::NUMERIC) AS total_indirect_cost,
    SUM((mci.economic_impact->'impacts'->>'regulatory_penalty')::NUMERIC) AS total_regulatory_penalty,
    SUM((mci.economic_impact->'impacts'->>'total_impact')::NUMERIC) AS total_economic_impact,
    mci.economic_impact->>'currency' AS currency,
    vh.tenant_id
FROM
    economic_planning.model_compliance_integrations mci
JOIN
    iam_validators.validation_history vh ON mci.validation_id = vh.validation_id AND mci.tenant_id = vh.tenant_id
WHERE
    vh.is_compliant = FALSE
GROUP BY
    EXTRACT(YEAR FROM mci.integration_timestamp),
    EXTRACT(MONTH FROM mci.integration_timestamp),
    vh.validator_id,
    vh.validator_name,
    vh.jurisdiction,
    vh.impact_level,
    mci.economic_impact->>'currency',
    vh.tenant_id;

COMMENT ON VIEW economic_planning.vw_compliance_economic_impact_summary IS 'Resumo agregado de impactos econômicos por validador, período e região';

-- View para impacto econômico por validador
CREATE OR REPLACE VIEW economic_planning.vw_economic_impact_by_validator AS
SELECT
    vh.validator_id,
    vh.validator_name,
    vh.validator_type,
    cv.regulatory_framework,
    COUNT(mci.validation_id) AS total_validations,
    COUNT(CASE WHEN vh.is_compliant = FALSE THEN 1 END) AS failed_validations,
    ROUND(COUNT(CASE WHEN vh.is_compliant = FALSE THEN 1 END)::NUMERIC / 
          NULLIF(COUNT(mci.validation_id), 0)::NUMERIC * 100, 2) AS failure_rate,
    SUM((mci.economic_impact->'impacts'->>'total_impact')::NUMERIC) AS total_economic_impact,
    MAX((mci.economic_impact->'impacts'->>'total_impact')::NUMERIC) AS max_impact,
    mci.economic_impact->>'currency' AS currency,
    vh.tenant_id
FROM
    economic_planning.model_compliance_integrations mci
JOIN
    iam_validators.validation_history vh ON mci.validation_id = vh.validation_id AND mci.tenant_id = vh.tenant_id
JOIN
    iam_validators.compliance_validators cv ON vh.validator_id = cv.validator_id AND vh.tenant_id = cv.tenant_id
GROUP BY
    vh.validator_id,
    vh.validator_name,
    vh.validator_type,
    cv.regulatory_framework,
    mci.economic_impact->>'currency',
    vh.tenant_id;

COMMENT ON VIEW economic_planning.vw_economic_impact_by_validator IS 'Impacto econômico agregado por validador e framework regulatório';

-- View para impacto econômico por região
CREATE OR REPLACE VIEW economic_planning.vw_economic_impact_by_region AS
SELECT
    vh.jurisdiction,
    cv.regulatory_framework,
    COUNT(mci.validation_id) AS total_validations,
    COUNT(CASE WHEN vh.is_compliant = FALSE THEN 1 END) AS failed_validations,
    ROUND(COUNT(CASE WHEN vh.is_compliant = FALSE THEN 1 END)::NUMERIC / 
          NULLIF(COUNT(mci.validation_id), 0)::NUMERIC * 100, 2) AS failure_rate,
    SUM((mci.economic_impact->'impacts'->>'total_impact')::NUMERIC) AS total_economic_impact,
    AVG((mci.economic_impact->'impacts'->>'total_impact')::NUMERIC) AS avg_impact,
    mci.economic_impact->>'currency' AS currency,
    vh.tenant_id
FROM
    economic_planning.model_compliance_integrations mci
JOIN
    iam_validators.validation_history vh ON mci.validation_id = vh.validation_id AND mci.tenant_id = vh.tenant_id
JOIN
    iam_validators.compliance_validators cv ON vh.validator_id = cv.validator_id AND vh.tenant_id = cv.tenant_id
GROUP BY
    vh.jurisdiction,
    cv.regulatory_framework,
    mci.economic_impact->>'currency',
    vh.tenant_id;

COMMENT ON VIEW economic_planning.vw_economic_impact_by_region IS 'Impacto econômico agregado por região e framework regulatório';

-- View para tendência de impacto econômico ao longo do tempo
CREATE OR REPLACE VIEW economic_planning.vw_economic_impact_trend AS
SELECT
    DATE_TRUNC('month', mci.integration_timestamp) AS month_date,
    vh.jurisdiction,
    cv.regulatory_framework,
    COUNT(mci.validation_id) AS total_validations,
    COUNT(CASE WHEN vh.is_compliant = FALSE THEN 1 END) AS failed_validations,
    SUM((mci.economic_impact->'impacts'->>'total_impact')::NUMERIC) AS total_economic_impact,
    mci.economic_impact->>'currency' AS currency,
    vh.tenant_id
FROM
    economic_planning.model_compliance_integrations mci
JOIN
    iam_validators.validation_history vh ON mci.validation_id = vh.validation_id AND mci.tenant_id = vh.tenant_id
JOIN
    iam_validators.compliance_validators cv ON vh.validator_id = cv.validator_id AND vh.tenant_id = cv.tenant_id
GROUP BY
    DATE_TRUNC('month', mci.integration_timestamp),
    vh.jurisdiction,
    cv.regulatory_framework,
    mci.economic_impact->>'currency',
    vh.tenant_id
ORDER BY
    month_date,
    vh.jurisdiction,
    cv.regulatory_framework;

COMMENT ON VIEW economic_planning.vw_economic_impact_trend IS 'Tendência de impacto econômico ao longo do tempo por região e framework';

-- View para retorno sobre investimento em conformidade
CREATE OR REPLACE VIEW economic_planning.vw_compliance_roi AS
WITH remediation_costs AS (
    SELECT
        tenant_id,
        SUM(remediation_cost) AS total_remediation_cost
    FROM 
        quality_management.corrective_actions
    WHERE 
        status = 'COMPLETED'
    GROUP BY 
        tenant_id
),
avoided_penalties AS (
    SELECT
        vh.tenant_id,
        SUM((mci.economic_impact->'impacts'->>'regulatory_penalty')::NUMERIC) AS total_avoided_penalties
    FROM
        economic_planning.model_compliance_integrations mci
    JOIN
        iam_validators.validation_history vh ON mci.validation_id = vh.validation_id AND mci.tenant_id = vh.tenant_id
    JOIN
        quality_management.corrective_actions ca ON vh.validation_id = ca.validation_id AND vh.tenant_id = ca.tenant_id
    WHERE
        ca.status = 'COMPLETED'
        AND vh.is_compliant = TRUE
    GROUP BY
        vh.tenant_id
)
SELECT
    rc.tenant_id,
    rc.total_remediation_cost,
    COALESCE(ap.total_avoided_penalties, 0) AS total_avoided_penalties,
    CASE 
        WHEN rc.total_remediation_cost = 0 THEN NULL
        ELSE ROUND((COALESCE(ap.total_avoided_penalties, 0) - rc.total_remediation_cost) / 
                  NULLIF(rc.total_remediation_cost, 0) * 100, 2)
    END AS roi_percentage
FROM
    remediation_costs rc
LEFT JOIN
    avoided_penalties ap ON rc.tenant_id = ap.tenant_id;

COMMENT ON VIEW economic_planning.vw_compliance_roi IS 'Retorno sobre investimento em ações de conformidade, comparando custos de remediação com penalidades evitadas';

-- ============================================================================
-- Funções para KPIs Econômicos
-- ============================================================================

-- Função para obter o ROI de conformidade para um período específico
CREATE OR REPLACE FUNCTION economic_planning.get_compliance_roi(
    p_tenant_id VARCHAR(100),
    p_start_date DATE,
    p_end_date DATE
) RETURNS TABLE (
    metric_name VARCHAR,
    metric_value NUMERIC,
    percentage NUMERIC,
    currency VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    WITH remediation_costs AS (
        SELECT
            tenant_id,
            SUM(remediation_cost) AS total_remediation_cost
        FROM 
            quality_management.corrective_actions
        WHERE 
            status = 'COMPLETED'
            AND completion_date BETWEEN p_start_date AND p_end_date
            AND tenant_id = p_tenant_id
        GROUP BY 
            tenant_id
    ),
    avoided_penalties AS (
        SELECT
            vh.tenant_id,
            SUM((mci.economic_impact->'impacts'->>'regulatory_penalty')::NUMERIC) AS total_avoided_penalties,
            mci.economic_impact->>'currency' AS currency
        FROM
            economic_planning.model_compliance_integrations mci
        JOIN
            iam_validators.validation_history vh ON mci.validation_id = vh.validation_id AND mci.tenant_id = vh.tenant_id
        JOIN
            quality_management.corrective_actions ca ON vh.validation_id = ca.validation_id AND vh.tenant_id = ca.tenant_id
        WHERE
            ca.status = 'COMPLETED'
            AND ca.completion_date BETWEEN p_start_date AND p_end_date
            AND vh.is_compliant = TRUE
            AND vh.tenant_id = p_tenant_id
        GROUP BY
            vh.tenant_id,
            mci.economic_impact->>'currency'
    )
    SELECT 
        'Custo Total de Remediação' AS metric_name,
        COALESCE(rc.total_remediation_cost, 0) AS metric_value,
        NULL::NUMERIC AS percentage,
        COALESCE(ap.currency, 'EUR') AS currency
    FROM 
        remediation_costs rc
    LEFT JOIN 
        avoided_penalties ap ON rc.tenant_id = ap.tenant_id
    
    UNION ALL
    
    SELECT 
        'Penalidades Evitadas' AS metric_name,
        COALESCE(ap.total_avoided_penalties, 0) AS metric_value,
        NULL::NUMERIC AS percentage,
        COALESCE(ap.currency, 'EUR') AS currency
    FROM 
        remediation_costs rc
    LEFT JOIN 
        avoided_penalties ap ON rc.tenant_id = ap.tenant_id
    
    UNION ALL
    
    SELECT 
        'Retorno Líquido' AS metric_name,
        COALESCE(ap.total_avoided_penalties, 0) - COALESCE(rc.total_remediation_cost, 0) AS metric_value,
        CASE 
            WHEN COALESCE(rc.total_remediation_cost, 0) = 0 THEN NULL
            ELSE ROUND((COALESCE(ap.total_avoided_penalties, 0) - COALESCE(rc.total_remediation_cost, 0)) / 
                     NULLIF(COALESCE(rc.total_remediation_cost, 0), 0) * 100, 2)
        END AS percentage,
        COALESCE(ap.currency, 'EUR') AS currency
    FROM 
        remediation_costs rc
    LEFT JOIN 
        avoided_penalties ap ON rc.tenant_id = ap.tenant_id;
END;
$$ LANGUAGE plpgsql;

-- Função para obter impacto econômico por framework regulatório
CREATE OR REPLACE FUNCTION economic_planning.get_economic_impact_by_framework(
    p_tenant_id VARCHAR(100),
    p_regulatory_framework VARCHAR(100) DEFAULT NULL
) RETURNS TABLE (
    regulatory_framework VARCHAR,
    total_validations BIGINT,
    failed_validations BIGINT,
    failure_rate NUMERIC,
    total_economic_impact NUMERIC,
    avg_impact_per_failure NUMERIC,
    currency VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        cv.regulatory_framework,
        COUNT(mci.validation_id) AS total_validations,
        COUNT(CASE WHEN vh.is_compliant = FALSE THEN 1 END) AS failed_validations,
        ROUND(COUNT(CASE WHEN vh.is_compliant = FALSE THEN 1 END)::NUMERIC / 
              NULLIF(COUNT(mci.validation_id), 0)::NUMERIC * 100, 2) AS failure_rate,
        SUM((mci.economic_impact->'impacts'->>'total_impact')::NUMERIC) AS total_economic_impact,
        CASE 
            WHEN COUNT(CASE WHEN vh.is_compliant = FALSE THEN 1 END) = 0 THEN 0
            ELSE ROUND(SUM((mci.economic_impact->'impacts'->>'total_impact')::NUMERIC) / 
                 COUNT(CASE WHEN vh.is_compliant = FALSE THEN 1 END), 2)
        END AS avg_impact_per_failure,
        mci.economic_impact->>'currency' AS currency
    FROM
        economic_planning.model_compliance_integrations mci
    JOIN
        iam_validators.validation_history vh ON mci.validation_id = vh.validation_id AND mci.tenant_id = vh.tenant_id
    JOIN
        iam_validators.compliance_validators cv ON vh.validator_id = cv.validator_id AND vh.tenant_id = cv.tenant_id
    WHERE
        vh.tenant_id = p_tenant_id
        AND (p_regulatory_framework IS NULL OR cv.regulatory_framework = p_regulatory_framework)
    GROUP BY
        cv.regulatory_framework,
        mci.economic_impact->>'currency'
    ORDER BY
        total_economic_impact DESC;
END;
$$ LANGUAGE plpgsql;

-- Função para obter potenciais penalidades por jurisdição
CREATE OR REPLACE FUNCTION economic_planning.get_potential_penalties_by_jurisdiction(
    p_tenant_id VARCHAR(100),
    p_jurisdiction VARCHAR(50) DEFAULT NULL
) RETURNS TABLE (
    jurisdiction VARCHAR,
    regulatory_framework VARCHAR,
    open_non_compliances BIGINT,
    estimated_penalties NUMERIC,
    max_potential_penalty NUMERIC,
    currency VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        vh.jurisdiction,
        cv.regulatory_framework,
        COUNT(*) AS open_non_compliances,
        SUM((mci.economic_impact->'impacts'->>'regulatory_penalty')::NUMERIC) AS estimated_penalties,
        MAX((mci.economic_impact->'impacts'->>'regulatory_penalty')::NUMERIC) AS max_potential_penalty,
        mci.economic_impact->>'currency' AS currency
    FROM
        economic_planning.model_compliance_integrations mci
    JOIN
        iam_validators.validation_history vh ON mci.validation_id = vh.validation_id AND mci.tenant_id = vh.tenant_id
    JOIN
        iam_validators.compliance_validators cv ON vh.validator_id = cv.validator_id AND vh.tenant_id = cv.tenant_id
    LEFT JOIN
        quality_management.non_conformity nc ON vh.validation_id = nc.validation_id AND vh.tenant_id = nc.tenant_id
    WHERE
        vh.is_compliant = FALSE
        AND (nc.status IS NULL OR nc.status != 'CLOSED')
        AND vh.tenant_id = p_tenant_id
        AND (p_jurisdiction IS NULL OR vh.jurisdiction = p_jurisdiction)
    GROUP BY
        vh.jurisdiction,
        cv.regulatory_framework,
        mci.economic_impact->>'currency'
    ORDER BY
        estimated_penalties DESC;
END;
$$ LANGUAGE plpgsql;

-- Função para obter exposição de risco econômico
CREATE OR REPLACE FUNCTION economic_planning.get_economic_risk_exposure(
    p_tenant_id VARCHAR(100),
    p_confidence_level NUMERIC DEFAULT 0.95
) RETURNS TABLE (
    risk_category VARCHAR,
    base_exposure NUMERIC,
    risk_adjusted_exposure NUMERIC,
    confidence_level NUMERIC,
    currency VARCHAR
) AS $$
DECLARE
    v_z_score NUMERIC;
BEGIN
    -- Cálculo do z-score para o nível de confiança especificado
    -- Para 95% de confiança (padrão), o z-score é aproximadamente 1.96
    IF p_confidence_level = 0.99 THEN
        v_z_score := 2.576;
    ELSIF p_confidence_level = 0.975 THEN
        v_z_score := 2.24;
    ELSIF p_confidence_level = 0.95 THEN
        v_z_score := 1.96;
    ELSIF p_confidence_level = 0.90 THEN
        v_z_score := 1.645;
    ELSE
        v_z_score := 1.96; -- Padrão para 95% de confiança
    END IF;
    
    RETURN QUERY
    WITH risk_exposure AS (
        SELECT
            CASE 
                WHEN vh.impact_level = 'CRITICAL' THEN 'Risco Crítico'
                WHEN vh.impact_level = 'HIGH' THEN 'Risco Alto'
                WHEN vh.impact_level = 'MEDIUM' THEN 'Risco Médio'
                WHEN vh.impact_level = 'LOW' THEN 'Risco Baixo'
                ELSE 'Não Classificado'
            END AS risk_category,
            SUM((mci.economic_impact->'impacts'->>'total_impact')::NUMERIC) AS base_exposure,
            STDDEV((mci.economic_impact->'impacts'->>'total_impact')::NUMERIC) AS std_dev,
            mci.economic_impact->>'currency' AS currency
        FROM
            economic_planning.model_compliance_integrations mci
        JOIN
            iam_validators.validation_history vh ON mci.validation_id = vh.validation_id AND mci.tenant_id = vh.tenant_id
        LEFT JOIN
            quality_management.non_conformity nc ON vh.validation_id = nc.validation_id AND vh.tenant_id = nc.tenant_id
        WHERE
            vh.is_compliant = FALSE
            AND (nc.status IS NULL OR nc.status != 'CLOSED')
            AND vh.tenant_id = p_tenant_id
        GROUP BY
            risk_category,
            mci.economic_impact->>'currency'
    )
    SELECT
        risk_category,
        base_exposure,
        base_exposure + (v_z_score * COALESCE(std_dev, base_exposure * 0.1)) AS risk_adjusted_exposure,
        p_confidence_level AS confidence_level,
        currency
    FROM
        risk_exposure
    ORDER BY
        risk_adjusted_exposure DESC;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Funções para Análise Preditiva
-- ============================================================================

-- Função para prever tendências de impacto econômico
CREATE OR REPLACE FUNCTION economic_planning.predict_economic_impact_trends(
    p_tenant_id VARCHAR(100),
    p_forecast_months INTEGER DEFAULT 6
) RETURNS TABLE (
    forecast_date DATE,
    jurisdiction VARCHAR,
    regulatory_framework VARCHAR,
    predicted_impact NUMERIC,
    prediction_lower_bound NUMERIC,
    prediction_upper_bound NUMERIC,
    confidence_level NUMERIC,
    currency VARCHAR
) AS $$
DECLARE
    v_today DATE := CURRENT_DATE;
    v_history_months INTEGER := 12; -- Histórico utilizado para previsão
BEGIN
    -- Análise de tendência simplificada baseada em média móvel ponderada e tendência linear
    RETURN QUERY
    WITH historical_data AS (
        SELECT
            DATE_TRUNC('month', mci.integration_timestamp)::DATE AS month_date,
            vh.jurisdiction,
            cv.regulatory_framework,
            SUM((mci.economic_impact->'impacts'->>'total_impact')::NUMERIC) AS monthly_impact,
            mci.economic_impact->>'currency' AS currency
        FROM
            economic_planning.model_compliance_integrations mci
        JOIN
            iam_validators.validation_history vh ON mci.validation_id = vh.validation_id AND mci.tenant_id = vh.tenant_id
        JOIN
            iam_validators.compliance_validators cv ON vh.validator_id = cv.validator_id AND vh.tenant_id = cv.tenant_id
        WHERE
            vh.tenant_id = p_tenant_id
            AND mci.integration_timestamp >= (v_today - (v_history_months || ' months')::INTERVAL)
        GROUP BY
            month_date,
            vh.jurisdiction,
            cv.regulatory_framework,
            mci.economic_impact->>'currency'
        ORDER BY
            month_date
    ),
    trend_analysis AS (
        SELECT
            jurisdiction,
            regulatory_framework,
            currency,
            REGR_SLOPE(monthly_impact, EXTRACT(EPOCH FROM month_date - MIN(month_date) OVER (PARTITION BY jurisdiction, regulatory_framework))/86400) AS daily_slope,
            REGR_INTERCEPT(monthly_impact, EXTRACT(EPOCH FROM month_date - MIN(month_date) OVER (PARTITION BY jurisdiction, regulatory_framework))/86400) AS intercept,
            AVG(monthly_impact) AS avg_impact,
            STDDEV(monthly_impact) AS std_impact,
            COUNT(*) AS data_points
        FROM
            historical_data
        GROUP BY
            jurisdiction,
            regulatory_framework,
            currency
    )
    SELECT
        (v_today + (n || ' months')::INTERVAL)::DATE AS forecast_date,
        ta.jurisdiction,
        ta.regulatory_framework,
        -- Previsão combinando tendência linear com média ponderada
        GREATEST(0, ROUND(
            ta.intercept + ta.daily_slope * EXTRACT(EPOCH FROM (v_today + (n || ' months')::INTERVAL) - v_today)/86400
        , 2)) AS predicted_impact,
        -- Limite inferior (intervalo de confiança de 95%)
        GREATEST(0, ROUND(
            (ta.intercept + ta.daily_slope * EXTRACT(EPOCH FROM (v_today + (n || ' months')::INTERVAL) - v_today)/86400) - 
            (1.96 * COALESCE(ta.std_impact, ta.avg_impact * 0.1) / SQRT(GREATEST(ta.data_points, 1)))
        , 2)) AS prediction_lower_bound,
        -- Limite superior (intervalo de confiança de 95%)
        ROUND(
            (ta.intercept + ta.daily_slope * EXTRACT(EPOCH FROM (v_today + (n || ' months')::INTERVAL) - v_today)/86400) + 
            (1.96 * COALESCE(ta.std_impact, ta.avg_impact * 0.1) / SQRT(GREATEST(ta.data_points, 1)))
        , 2) AS prediction_upper_bound,
        0.95 AS confidence_level,
        ta.currency
    FROM
        trend_analysis ta
    CROSS JOIN
        generate_series(1, p_forecast_months) AS n
    ORDER BY
        ta.jurisdiction,
        ta.regulatory_framework,
        forecast_date;
END;
$$ LANGUAGE plpgsql;

-- Função para simular impacto da melhoria de conformidade
CREATE OR REPLACE FUNCTION economic_planning.simulate_compliance_improvement_impact(
    p_tenant_id VARCHAR(100),
    p_improvement_percentage NUMERIC DEFAULT 20.0, -- Percentual de melhoria na taxa de conformidade
    p_simulation_months INTEGER DEFAULT 12 -- Período de simulação
) RETURNS TABLE (
    jurisdiction VARCHAR,
    regulatory_framework VARCHAR,
    current_failure_rate NUMERIC,
    simulated_failure_rate NUMERIC,
    current_monthly_impact NUMERIC,
    simulated_monthly_impact NUMERIC,
    total_savings NUMERIC,
    roi_percentage NUMERIC,
    currency VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    WITH current_metrics AS (
        SELECT
            vh.jurisdiction,
            cv.regulatory_framework,
            COUNT(CASE WHEN vh.is_compliant = FALSE THEN 1 END)::NUMERIC / 
                NULLIF(COUNT(mci.validation_id), 0)::NUMERIC AS failure_rate,
            SUM((mci.economic_impact->'impacts'->>'total_impact')::NUMERIC) / 
                GREATEST(EXTRACT(MONTH FROM AGE(MAX(mci.integration_timestamp), MIN(mci.integration_timestamp))), 1) AS monthly_impact,
            mci.economic_impact->>'currency' AS currency
        FROM
            economic_planning.model_compliance_integrations mci
        JOIN
            iam_validators.validation_history vh ON mci.validation_id = vh.validation_id AND mci.tenant_id = vh.tenant_id
        JOIN
            iam_validators.compliance_validators cv ON vh.validator_id = cv.validator_id AND vh.tenant_id = cv.tenant_id
        WHERE
            vh.tenant_id = p_tenant_id
        GROUP BY
            vh.jurisdiction,
            cv.regulatory_framework,
            mci.economic_impact->>'currency'
    ),
    remediation_costs AS (
        SELECT
            vh.jurisdiction,
            cv.regulatory_framework,
            AVG(ca.remediation_cost) AS avg_remediation_cost
        FROM
            quality_management.corrective_actions ca
        JOIN
            iam_validators.validation_history vh ON ca.validation_id = vh.validation_id AND ca.tenant_id = vh.tenant_id
        JOIN
            iam_validators.compliance_validators cv ON vh.validator_id = cv.validator_id AND vh.tenant_id = cv.tenant_id
        WHERE
            ca.status = 'COMPLETED'
            AND vh.tenant_id = p_tenant_id
        GROUP BY
            vh.jurisdiction,
            cv.regulatory_framework
    )
    SELECT
        cm.jurisdiction,
        cm.regulatory_framework,
        ROUND(cm.failure_rate * 100, 2) AS current_failure_rate,
        ROUND(GREATEST(0, cm.failure_rate * (1 - p_improvement_percentage / 100)) * 100, 2) AS simulated_failure_rate,
        ROUND(cm.monthly_impact, 2) AS current_monthly_impact,
        ROUND(cm.monthly_impact * (1 - p_improvement_percentage / 100), 2) AS simulated_monthly_impact,
        ROUND((cm.monthly_impact - cm.monthly_impact * (1 - p_improvement_percentage / 100)) * p_simulation_months, 2) AS total_savings,
        CASE
            WHEN COALESCE(rc.avg_remediation_cost, 0) = 0 THEN NULL
            ELSE ROUND(
                ((cm.monthly_impact - cm.monthly_impact * (1 - p_improvement_percentage / 100)) * p_simulation_months) / 
                (COALESCE(rc.avg_remediation_cost, 1) * (cm.failure_rate * p_improvement_percentage / 100) * 100) * 100, 
                2)
        END AS roi_percentage,
        cm.currency
    FROM
        current_metrics cm
    LEFT JOIN
        remediation_costs rc ON cm.jurisdiction = rc.jurisdiction AND cm.regulatory_framework = rc.regulatory_framework
    ORDER BY
        total_savings DESC;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Conceder permissões para funções
-- ============================================================================

-- Permissões para funções de KPIs
GRANT EXECUTE ON FUNCTION economic_planning.get_compliance_roi TO economic_analyst_role, compliance_manager_role, dashboard_viewer_role;
GRANT EXECUTE ON FUNCTION economic_planning.get_economic_impact_by_framework TO economic_analyst_role, compliance_manager_role, dashboard_viewer_role;
GRANT EXECUTE ON FUNCTION economic_planning.get_potential_penalties_by_jurisdiction TO economic_analyst_role, compliance_manager_role, dashboard_viewer_role;
GRANT EXECUTE ON FUNCTION economic_planning.get_economic_risk_exposure TO economic_analyst_role, compliance_manager_role, risk_analyst_role, dashboard_viewer_role;

-- Permissões para funções de análise preditiva
GRANT EXECUTE ON FUNCTION economic_planning.predict_economic_impact_trends TO economic_analyst_role, compliance_manager_role, risk_analyst_role;
GRANT EXECUTE ON FUNCTION economic_planning.simulate_compliance_improvement_impact TO economic_analyst_role, compliance_manager_role, risk_analyst_role;

-- ============================================================================
-- Comentários Finais
-- ============================================================================

COMMENT ON SCHEMA economic_planning IS 'Schema para dashboard econômico de conformidade IAM e modelagem econômica. Implementado em 15/05/2025.';
