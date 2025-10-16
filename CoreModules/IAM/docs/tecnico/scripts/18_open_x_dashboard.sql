-- ============================================================================
-- Script:      18_open_x_dashboard.sql
-- Autor:       Eduardo Jeremias
-- Projeto:     INNOVABIZ - Suíte de Sistema de Governança Inteligente Empresarial
-- Data:        15/05/2025
-- Descrição:   Dashboard para visualização e análise de conformidade e impactos
--              econômicos do ecossistema Open X
-- ============================================================================

-- Garantir que o schema compliance_validators exista
CREATE SCHEMA IF NOT EXISTS compliance_validators;

-- Comentário no script
COMMENT ON SCHEMA compliance_validators IS 'Schema para validadores de conformidade do ecossistema Open X';

-- Configuração de permissões
GRANT USAGE ON SCHEMA compliance_validators TO economic_analyst_role, compliance_manager_role, risk_analyst_role, dashboard_viewer_role;

-- ============================================================================
-- Views de Resumo e Agregação para o Ecossistema Open X
-- ============================================================================

-- View para resumo de conformidade Open X (todos os domínios)
CREATE OR REPLACE VIEW compliance_validators.vw_open_x_compliance_summary AS
WITH compliance_scores AS (
    SELECT 
        open_x_domain,
        framework,
        compliance_score,
        total_requirements,
        compliant_requirements,
        compliance_percentage
    FROM 
        compliance_validators.calculate_open_x_compliance_score(NULL)
)
SELECT
    cs.open_x_domain,
    cs.framework,
    cs.compliance_score,
    cs.total_requirements,
    cs.compliant_requirements,
    cs.compliance_percentage,
    CASE 
        WHEN cs.compliance_percentage >= 95 THEN 'R1'
        WHEN cs.compliance_percentage >= 85 THEN 'R2'
        WHEN cs.compliance_percentage >= 70 THEN 'R3'
        ELSE 'R4'
    END AS risk_level,
    CASE 
        WHEN cs.compliance_percentage >= 95 THEN 'BAIXO'
        WHEN cs.compliance_percentage >= 85 THEN 'MODERADO'
        WHEN cs.compliance_percentage >= 70 THEN 'ALTO'
        ELSE 'CRÍTICO'
    END AS risk_description,
    vh.tenant_id
FROM 
    compliance_scores cs
CROSS JOIN
    (SELECT DISTINCT tenant_id FROM iam_validators.validation_history LIMIT 1) vh;

COMMENT ON VIEW compliance_validators.vw_open_x_compliance_summary IS 'Resumo da conformidade para todo o ecossistema Open X por domínio e framework';

-- View para detalhe de não-conformidades por domínio Open X
CREATE OR REPLACE VIEW compliance_validators.vw_open_x_non_compliance_details AS
SELECT
    r.open_x_domain,
    r.framework,
    r.requirement_id,
    r.requirement_name,
    r.is_compliant,
    r.details,
    r.irr_threshold,
    vh.tenant_id
FROM
    compliance_validators.generate_open_x_compliance_report(NULL) r
CROSS JOIN
    (SELECT DISTINCT tenant_id FROM iam_validators.validation_history LIMIT 1) vh
WHERE
    r.is_compliant = FALSE;

COMMENT ON VIEW compliance_validators.vw_open_x_non_compliance_details IS 'Detalhes de não-conformidades para todo o ecossistema Open X';

-- View para impacto econômico por domínio Open X
CREATE OR REPLACE VIEW compliance_validators.vw_open_x_economic_impact AS
WITH economic_impacts AS (
    SELECT
        ocr.open_x_domain,
        ocr.framework,
        ocr.requirement_id,
        ocr.requirement_name,
        ocr.is_compliant,
        ocr.irr_threshold,
        CASE
            WHEN ocr.is_compliant = FALSE THEN
                economic_planning.calculate_compliance_economic_impact(
                    ocr.requirement_id,
                    NULL,
                    'GLOBAL',
                    'FINANCIAL_SERVICES',
                    ocr.irr_threshold,
                    vh.tenant_id
                )->>'monetary_impact'
            ELSE '0'
        END::NUMERIC AS monetary_impact,
        vh.tenant_id
    FROM
        compliance_validators.generate_open_x_compliance_report(NULL) ocr
    CROSS JOIN
        (SELECT DISTINCT tenant_id FROM iam_validators.validation_history LIMIT 1) vh
)
SELECT
    open_x_domain,
    framework,
    COUNT(*) AS total_requirements,
    COUNT(*) FILTER (WHERE is_compliant = FALSE) AS non_compliant_requirements,
    ROUND(COUNT(*) FILTER (WHERE is_compliant = FALSE)::NUMERIC / COUNT(*)::NUMERIC * 100, 2) AS non_compliance_percentage,
    SUM(monetary_impact) AS total_monetary_impact,
    CASE
        WHEN COUNT(*) FILTER (WHERE is_compliant = FALSE) > 0 THEN
            ROUND(SUM(monetary_impact) / COUNT(*) FILTER (WHERE is_compliant = FALSE), 2)
        ELSE 0
    END AS avg_impact_per_non_compliance,
    'EUR' AS currency,
    tenant_id
FROM
    economic_impacts
GROUP BY
    open_x_domain,
    framework,
    tenant_id;

COMMENT ON VIEW compliance_validators.vw_open_x_economic_impact IS 'Impacto econômico por domínio e framework no ecossistema Open X';

-- View para análise comparativa entre domínios Open X
CREATE OR REPLACE VIEW compliance_validators.vw_open_x_domain_comparison AS
SELECT
    open_x_domain,
    COUNT(*) FILTER (WHERE is_compliant = TRUE) AS compliant_requirements,
    COUNT(*) FILTER (WHERE is_compliant = FALSE) AS non_compliant_requirements,
    COUNT(*) AS total_requirements,
    ROUND(COUNT(*) FILTER (WHERE is_compliant = TRUE)::NUMERIC / COUNT(*)::NUMERIC * 100, 2) AS compliance_percentage,
    CASE 
        WHEN COUNT(*) FILTER (WHERE is_compliant = TRUE)::NUMERIC / COUNT(*)::NUMERIC >= 0.95 THEN 'R1'
        WHEN COUNT(*) FILTER (WHERE is_compliant = TRUE)::NUMERIC / COUNT(*)::NUMERIC >= 0.85 THEN 'R2'
        WHEN COUNT(*) FILTER (WHERE is_compliant = TRUE)::NUMERIC / COUNT(*)::NUMERIC >= 0.70 THEN 'R3'
        ELSE 'R4'
    END AS risk_level,
    vh.tenant_id
FROM
    compliance_validators.generate_open_x_compliance_report(NULL) r
CROSS JOIN
    (SELECT DISTINCT tenant_id FROM iam_validators.validation_history LIMIT 1) vh
GROUP BY
    open_x_domain,
    vh.tenant_id;

COMMENT ON VIEW compliance_validators.vw_open_x_domain_comparison IS 'Análise comparativa de conformidade entre os diferentes domínios Open X';

-- View para conformidade por jurisdição no ecossistema Open X
CREATE OR REPLACE VIEW compliance_validators.vw_open_x_jurisdiction_compliance AS
WITH jurisdiction_mapping AS (
    -- Mapear jurisdições para os diferentes frameworks regulatórios
    SELECT 'SOLVENCY_II' AS framework, 'EU' AS jurisdiction, 'OPEN_INSURANCE' AS open_x_domain
    UNION SELECT 'SUSEP_BR' AS framework, 'BRASIL' AS jurisdiction, 'OPEN_INSURANCE' AS open_x_domain
    UNION SELECT 'OPEN_HEALTH_GENERAL' AS framework, 'GLOBAL' AS jurisdiction, 'OPEN_HEALTH' AS open_x_domain
    UNION SELECT 'ANS_BR' AS framework, 'BRASIL' AS jurisdiction, 'OPEN_HEALTH' AS open_x_domain
    UNION SELECT 'EIDAS' AS framework, 'EU' AS jurisdiction, 'OPEN_GOVERNMENT' AS open_x_domain
    UNION SELECT 'GOV_BR' AS framework, 'BRASIL' AS jurisdiction, 'OPEN_GOVERNMENT' AS open_x_domain
)
SELECT
    jm.jurisdiction,
    cs.open_x_domain,
    cs.framework,
    cs.compliance_score,
    cs.total_requirements,
    cs.compliant_requirements,
    cs.compliance_percentage,
    CASE 
        WHEN cs.compliance_percentage >= 95 THEN 'R1'
        WHEN cs.compliance_percentage >= 85 THEN 'R2'
        WHEN cs.compliance_percentage >= 70 THEN 'R3'
        ELSE 'R4'
    END AS risk_level,
    vh.tenant_id
FROM 
    compliance_validators.calculate_open_x_compliance_score(NULL) cs
JOIN
    jurisdiction_mapping jm ON cs.framework = jm.framework
CROSS JOIN
    (SELECT DISTINCT tenant_id FROM iam_validators.validation_history LIMIT 1) vh
WHERE
    cs.framework != 'OVERALL';

COMMENT ON VIEW compliance_validators.vw_open_x_jurisdiction_compliance IS 'Análise de conformidade por jurisdição no ecossistema Open X';

-- ============================================================================
-- Funções para KPIs de Conformidade Open X
-- ============================================================================

-- Função para obter métricas de conformidade por domínio Open X
CREATE OR REPLACE FUNCTION compliance_validators.get_open_x_compliance_metrics(
    p_tenant_id UUID,
    p_open_x_domain VARCHAR(50) DEFAULT NULL
) RETURNS TABLE (
    open_x_domain VARCHAR,
    total_requirements BIGINT,
    compliant_requirements BIGINT,
    compliance_percentage NUMERIC,
    risk_level VARCHAR,
    risk_description VARCHAR,
    estimated_economic_impact NUMERIC,
    currency VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    WITH compliance_data AS (
        SELECT
            open_x_domain,
            COUNT(*) AS total_requirements,
            COUNT(*) FILTER (WHERE is_compliant = TRUE) AS compliant_requirements,
            ROUND(COUNT(*) FILTER (WHERE is_compliant = TRUE)::NUMERIC / COUNT(*)::NUMERIC * 100, 2) AS compliance_percentage,
            CASE 
                WHEN COUNT(*) FILTER (WHERE is_compliant = TRUE)::NUMERIC / COUNT(*)::NUMERIC >= 0.95 THEN 'R1'
                WHEN COUNT(*) FILTER (WHERE is_compliant = TRUE)::NUMERIC / COUNT(*)::NUMERIC >= 0.85 THEN 'R2'
                WHEN COUNT(*) FILTER (WHERE is_compliant = TRUE)::NUMERIC / COUNT(*)::NUMERIC >= 0.70 THEN 'R3'
                ELSE 'R4'
            END AS risk_level
        FROM
            compliance_validators.generate_open_x_compliance_report(p_tenant_id) r
        WHERE
            p_open_x_domain IS NULL OR open_x_domain = p_open_x_domain
        GROUP BY
            open_x_domain
    ),
    economic_data AS (
        SELECT
            rei.open_x_domain,
            (rei->>'total_economic_impact')::NUMERIC AS estimated_economic_impact,
            rei->>'currency' AS currency
        FROM (
            SELECT
                p_open_x_domain AS domain_filter,
                jsonb_array_elements(
                    CASE 
                        WHEN p_open_x_domain IS NULL THEN 
                            jsonb_build_array(
                                compliance_validators.register_open_x_economic_impact(p_tenant_id, 'OPEN_INSURANCE', 'GLOBAL', 'FINANCIAL_SERVICES'),
                                compliance_validators.register_open_x_economic_impact(p_tenant_id, 'OPEN_HEALTH', 'GLOBAL', 'HEALTHCARE'),
                                compliance_validators.register_open_x_economic_impact(p_tenant_id, 'OPEN_GOVERNMENT', 'GLOBAL', 'GOVERNMENT')
                            )
                        ELSE
                            jsonb_build_array(
                                compliance_validators.register_open_x_economic_impact(p_tenant_id, p_open_x_domain, 'GLOBAL', 
                                    CASE 
                                        WHEN p_open_x_domain = 'OPEN_INSURANCE' THEN 'FINANCIAL_SERVICES'
                                        WHEN p_open_x_domain = 'OPEN_HEALTH' THEN 'HEALTHCARE'
                                        WHEN p_open_x_domain = 'OPEN_GOVERNMENT' THEN 'GOVERNMENT'
                                        ELSE 'FINANCIAL_SERVICES'
                                    END
                                )
                            )
                    END
                ) AS rei
        ) impacts
    )
    SELECT
        cd.open_x_domain,
        cd.total_requirements,
        cd.compliant_requirements,
        cd.compliance_percentage,
        cd.risk_level,
        CASE 
            WHEN cd.risk_level = 'R1' THEN 'BAIXO'
            WHEN cd.risk_level = 'R2' THEN 'MODERADO'
            WHEN cd.risk_level = 'R3' THEN 'ALTO'
            ELSE 'CRÍTICO'
        END AS risk_description,
        COALESCE(ed.estimated_economic_impact, 0) AS estimated_economic_impact,
        COALESCE(ed.currency, 'EUR') AS currency
    FROM
        compliance_data cd
    LEFT JOIN
        economic_data ed ON cd.open_x_domain = ed.open_x_domain;
END;
$$ LANGUAGE plpgsql;

-- Função para comparar frameworks regulatórios por domínio Open X
CREATE OR REPLACE FUNCTION compliance_validators.compare_open_x_frameworks(
    p_tenant_id UUID,
    p_open_x_domain VARCHAR(50) DEFAULT NULL
) RETURNS TABLE (
    open_x_domain VARCHAR,
    framework VARCHAR,
    total_requirements BIGINT,
    compliant_requirements BIGINT,
    compliance_percentage NUMERIC,
    risk_level VARCHAR,
    estimated_economic_impact NUMERIC,
    currency VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    WITH framework_data AS (
        SELECT
            cs.open_x_domain,
            cs.framework,
            cs.total_requirements,
            cs.compliant_requirements,
            cs.compliance_percentage,
            CASE 
                WHEN cs.compliance_percentage >= 95 THEN 'R1'
                WHEN cs.compliance_percentage >= 85 THEN 'R2'
                WHEN cs.compliance_percentage >= 70 THEN 'R3'
                ELSE 'R4'
            END AS risk_level
        FROM 
            compliance_validators.calculate_open_x_compliance_score(p_tenant_id) cs
        WHERE
            (p_open_x_domain IS NULL OR cs.open_x_domain = p_open_x_domain)
            AND cs.framework != 'OVERALL'
    ),
    economic_data AS (
        SELECT
            r.open_x_domain,
            r.framework,
            SUM(
                CASE
                    WHEN r.is_compliant = FALSE THEN
                        (economic_planning.calculate_compliance_economic_impact(
                            r.requirement_id,
                            NULL,
                            'GLOBAL',
                            CASE 
                                WHEN r.open_x_domain = 'OPEN_INSURANCE' THEN 'FINANCIAL_SERVICES'
                                WHEN r.open_x_domain = 'OPEN_HEALTH' THEN 'HEALTHCARE'
                                WHEN r.open_x_domain = 'OPEN_GOVERNMENT' THEN 'GOVERNMENT'
                                ELSE 'FINANCIAL_SERVICES'
                            END,
                            r.irr_threshold,
                            p_tenant_id
                        )->>'monetary_impact')::NUMERIC
                    ELSE 0
                END
            ) AS estimated_economic_impact
        FROM
            compliance_validators.generate_open_x_compliance_report(p_tenant_id) r
        WHERE
            p_open_x_domain IS NULL OR r.open_x_domain = p_open_x_domain
        GROUP BY
            r.open_x_domain,
            r.framework
    )
    SELECT
        fd.open_x_domain,
        fd.framework,
        fd.total_requirements,
        fd.compliant_requirements,
        fd.compliance_percentage,
        fd.risk_level,
        COALESCE(ed.estimated_economic_impact, 0) AS estimated_economic_impact,
        'EUR' AS currency
    FROM
        framework_data fd
    LEFT JOIN
        economic_data ed ON fd.open_x_domain = ed.open_x_domain AND fd.framework = ed.framework
    ORDER BY
        fd.open_x_domain,
        fd.compliance_percentage DESC;
END;
$$ LANGUAGE plpgsql;

-- Função para analisar impacto econômico por jurisdição para o Open X
CREATE OR REPLACE FUNCTION compliance_validators.get_open_x_economic_impact_by_jurisdiction(
    p_tenant_id UUID,
    p_jurisdiction VARCHAR(50) DEFAULT NULL
) RETURNS TABLE (
    jurisdiction VARCHAR,
    open_x_domain VARCHAR,
    non_compliant_requirements BIGINT,
    estimated_economic_impact NUMERIC,
    avg_impact_per_requirement NUMERIC,
    risk_level VARCHAR,
    currency VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    WITH jurisdiction_mapping AS (
        -- Mapear jurisdições para os diferentes frameworks regulatórios
        SELECT 'SOLVENCY_II' AS framework, 'EU' AS jurisdiction, 'OPEN_INSURANCE' AS open_x_domain
        UNION SELECT 'SUSEP_BR' AS framework, 'BRASIL' AS jurisdiction, 'OPEN_INSURANCE' AS open_x_domain
        UNION SELECT 'OPEN_HEALTH_GENERAL' AS framework, 'GLOBAL' AS jurisdiction, 'OPEN_HEALTH' AS open_x_domain
        UNION SELECT 'ANS_BR' AS framework, 'BRASIL' AS jurisdiction, 'OPEN_HEALTH' AS open_x_domain
        UNION SELECT 'EIDAS' AS framework, 'EU' AS jurisdiction, 'OPEN_GOVERNMENT' AS open_x_domain
        UNION SELECT 'GOV_BR' AS framework, 'BRASIL' AS jurisdiction, 'OPEN_GOVERNMENT' AS open_x_domain
    ),
    non_compliance_data AS (
        SELECT
            jm.jurisdiction,
            r.open_x_domain,
            r.framework,
            COUNT(*) FILTER (WHERE r.is_compliant = FALSE) AS non_compliant_requirements,
            SUM(
                CASE
                    WHEN r.is_compliant = FALSE THEN
                        (economic_planning.calculate_compliance_economic_impact(
                            r.requirement_id,
                            NULL,
                            jm.jurisdiction,
                            CASE 
                                WHEN r.open_x_domain = 'OPEN_INSURANCE' THEN 'FINANCIAL_SERVICES'
                                WHEN r.open_x_domain = 'OPEN_HEALTH' THEN 'HEALTHCARE'
                                WHEN r.open_x_domain = 'OPEN_GOVERNMENT' THEN 'GOVERNMENT'
                                ELSE 'FINANCIAL_SERVICES'
                            END,
                            r.irr_threshold,
                            p_tenant_id
                        )->>'monetary_impact')::NUMERIC
                    ELSE 0
                END
            ) AS estimated_economic_impact
        FROM
            compliance_validators.generate_open_x_compliance_report(p_tenant_id) r
        JOIN
            jurisdiction_mapping jm ON r.framework = jm.framework
        WHERE
            p_jurisdiction IS NULL OR jm.jurisdiction = p_jurisdiction
        GROUP BY
            jm.jurisdiction,
            r.open_x_domain,
            r.framework
    ),
    jurisdiction_summary AS (
        SELECT
            jurisdiction,
            open_x_domain,
            SUM(non_compliant_requirements) AS non_compliant_requirements,
            SUM(estimated_economic_impact) AS estimated_economic_impact
        FROM
            non_compliance_data
        GROUP BY
            jurisdiction,
            open_x_domain
    )
    SELECT
        js.jurisdiction,
        js.open_x_domain,
        js.non_compliant_requirements,
        js.estimated_economic_impact,
        CASE
            WHEN js.non_compliant_requirements > 0 THEN
                ROUND(js.estimated_economic_impact / js.non_compliant_requirements, 2)
            ELSE 0
        END AS avg_impact_per_requirement,
        CASE 
            WHEN js.estimated_economic_impact >= 1000000 THEN 'R4'
            WHEN js.estimated_economic_impact >= 500000 THEN 'R3'
            WHEN js.estimated_economic_impact >= 100000 THEN 'R2'
            ELSE 'R1'
        END AS risk_level,
        'EUR' AS currency
    FROM
        jurisdiction_summary js
    ORDER BY
        js.estimated_economic_impact DESC;
END;
$$ LANGUAGE plpgsql;

-- Função para simular melhorias na conformidade do ecossistema Open X
CREATE OR REPLACE FUNCTION compliance_validators.simulate_open_x_improvements(
    p_tenant_id UUID,
    p_open_x_domain VARCHAR(50) DEFAULT NULL,
    p_improvement_percentage NUMERIC DEFAULT 20,
    p_months_horizon INTEGER DEFAULT 12
) RETURNS TABLE (
    open_x_domain VARCHAR,
    framework VARCHAR,
    current_compliance_percentage NUMERIC,
    simulated_compliance_percentage NUMERIC,
    current_economic_impact NUMERIC,
    simulated_economic_impact NUMERIC,
    cost_savings NUMERIC,
    estimated_roi_percentage NUMERIC,
    currency VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    WITH current_metrics AS (
        SELECT
            cs.open_x_domain,
            cs.framework,
            cs.compliance_percentage AS current_compliance_percentage,
            cs.compliance_percentage + (100 - cs.compliance_percentage) * (p_improvement_percentage / 100) AS simulated_compliance_percentage,
            SUM(
                CASE
                    WHEN r.is_compliant = FALSE THEN
                        (economic_planning.calculate_compliance_economic_impact(
                            r.requirement_id,
                            NULL,
                            'GLOBAL',
                            CASE 
                                WHEN r.open_x_domain = 'OPEN_INSURANCE' THEN 'FINANCIAL_SERVICES'
                                WHEN r.open_x_domain = 'OPEN_HEALTH' THEN 'HEALTHCARE'
                                WHEN r.open_x_domain = 'OPEN_GOVERNMENT' THEN 'GOVERNMENT'
                                ELSE 'FINANCIAL_SERVICES'
                            END,
                            r.irr_threshold,
                            p_tenant_id
                        )->>'monetary_impact')::NUMERIC
                    ELSE 0
                END
            ) / 12 AS monthly_impact  -- Mensalizando o impacto para cálculos de ROI
        FROM
            compliance_validators.calculate_open_x_compliance_score(p_tenant_id) cs
        LEFT JOIN
            compliance_validators.generate_open_x_compliance_report(p_tenant_id) r 
                ON cs.open_x_domain = r.open_x_domain AND cs.framework = r.framework
        WHERE
            (p_open_x_domain IS NULL OR cs.open_x_domain = p_open_x_domain)
            AND cs.framework != 'OVERALL'
        GROUP BY
            cs.open_x_domain,
            cs.framework,
            cs.compliance_percentage
    ),
    remediation_costs AS (
        SELECT
            iam_core.get_setting('open_x_remediation_cost_' || 
                LOWER(REPLACE(cm.open_x_domain, 'OPEN_', '')), '10000')::NUMERIC AS avg_remediation_cost,
            cm.open_x_domain,
            cm.framework
        FROM
            current_metrics cm
    )
    SELECT
        cm.open_x_domain,
        cm.framework,
        ROUND(cm.current_compliance_percentage, 2) AS current_compliance_percentage,
        ROUND(cm.simulated_compliance_percentage, 2) AS simulated_compliance_percentage,
        ROUND(cm.monthly_impact * p_months_horizon, 2) AS current_economic_impact,
        ROUND(cm.monthly_impact * (1 - p_improvement_percentage / 100) * p_months_horizon, 2) AS simulated_economic_impact,
        ROUND((cm.monthly_impact - cm.monthly_impact * (1 - p_improvement_percentage / 100)) * p_months_horizon, 2) AS cost_savings,
        CASE
            WHEN rc.avg_remediation_cost = 0 THEN NULL
            ELSE ROUND(
                ((cm.monthly_impact - cm.monthly_impact * (1 - p_improvement_percentage / 100)) * p_months_horizon) / 
                (rc.avg_remediation_cost * (100 - cm.current_compliance_percentage) * (p_improvement_percentage / 100) / 100) * 100,
                2)
        END AS estimated_roi_percentage,
        'EUR' AS currency
    FROM
        current_metrics cm
    JOIN
        remediation_costs rc ON cm.open_x_domain = rc.open_x_domain AND cm.framework = rc.framework
    ORDER BY
        cost_savings DESC;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Conceder permissões para views e funções
-- ============================================================================

-- Permissões para views
GRANT SELECT ON compliance_validators.vw_open_x_compliance_summary TO economic_analyst_role, compliance_manager_role, risk_analyst_role, dashboard_viewer_role;
GRANT SELECT ON compliance_validators.vw_open_x_non_compliance_details TO economic_analyst_role, compliance_manager_role, risk_analyst_role, dashboard_viewer_role;
GRANT SELECT ON compliance_validators.vw_open_x_economic_impact TO economic_analyst_role, compliance_manager_role, risk_analyst_role, dashboard_viewer_role;
GRANT SELECT ON compliance_validators.vw_open_x_domain_comparison TO economic_analyst_role, compliance_manager_role, risk_analyst_role, dashboard_viewer_role;
GRANT SELECT ON compliance_validators.vw_open_x_jurisdiction_compliance TO economic_analyst_role, compliance_manager_role, risk_analyst_role, dashboard_viewer_role;

-- Permissões para funções
GRANT EXECUTE ON FUNCTION compliance_validators.get_open_x_compliance_metrics TO economic_analyst_role, compliance_manager_role, risk_analyst_role, dashboard_viewer_role;
GRANT EXECUTE ON FUNCTION compliance_validators.compare_open_x_frameworks TO economic_analyst_role, compliance_manager_role, risk_analyst_role, dashboard_viewer_role;
GRANT EXECUTE ON FUNCTION compliance_validators.get_open_x_economic_impact_by_jurisdiction TO economic_analyst_role, compliance_manager_role, risk_analyst_role, dashboard_viewer_role;
GRANT EXECUTE ON FUNCTION compliance_validators.simulate_open_x_improvements TO economic_analyst_role, compliance_manager_role, risk_analyst_role;

-- ============================================================================
-- Comentários Finais
-- ============================================================================

COMMENT ON SCHEMA compliance_validators IS 'Schema para dashboard de conformidade do ecossistema Open X. Implementado em 15/05/2025.';

-- ============================================================================
-- Fim do Script
-- ============================================================================

-- =============================================
-- Dashboard Open X
-- INNOVABIZ - IAM - Open X Dashboard
-- Versão: 1.0.0
-- Data: 15/05/2025
-- Autor: INNOVABIZ DevOps Team
-- =============================================

-- Este script implementa o dashboard para o ecossistema Open X,
-- integrando os validadores específicos de Open Insurance, Open Health e Open Government
-- e o sistema de modelagem econômica para análise de impactos financeiros.

-- =============================================
-- Schemas necessários
-- =============================================

-- Verifica e cria os schemas necessários
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'compliance_validators') THEN
        CREATE SCHEMA compliance_validators;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'economic_planning') THEN
        CREATE SCHEMA economic_planning;
    END IF;
END $$;

-- =============================================
-- Views para Resumo e Agregação
-- =============================================

-- View para resumo de conformidade Open X por domínio e framework
CREATE OR REPLACE VIEW compliance_validators.vw_open_x_compliance_summary AS
SELECT
    tenant_id,
    open_x_domain,
    framework,
    COUNT(*) AS total_requirements,
    SUM(CASE WHEN is_compliant THEN 1 ELSE 0 END) AS compliant_requirements,
    ROUND((SUM(CASE WHEN is_compliant THEN 1 ELSE 0 END)::NUMERIC / COUNT(*)::NUMERIC) * 100, 2) AS compliance_percentage,
    CASE
        WHEN (SUM(CASE WHEN is_compliant THEN 1 ELSE 0 END)::NUMERIC / COUNT(*)::NUMERIC) >= 0.90 THEN 'R1'
        WHEN (SUM(CASE WHEN is_compliant THEN 1 ELSE 0 END)::NUMERIC / COUNT(*)::NUMERIC) >= 0.75 THEN 'R2'
        WHEN (SUM(CASE WHEN is_compliant THEN 1 ELSE 0 END)::NUMERIC / COUNT(*)::NUMERIC) >= 0.50 THEN 'R3'
        ELSE 'R4'
    END AS risk_level,
    CASE
        WHEN (SUM(CASE WHEN is_compliant THEN 1 ELSE 0 END)::NUMERIC / COUNT(*)::NUMERIC) >= 0.90 THEN 'BAIXO'
        WHEN (SUM(CASE WHEN is_compliant THEN 1 ELSE 0 END)::NUMERIC / COUNT(*)::NUMERIC) >= 0.75 THEN 'MODERADO'
        WHEN (SUM(CASE WHEN is_compliant THEN 1 ELSE 0 END)::NUMERIC / COUNT(*)::NUMERIC) >= 0.50 THEN 'ALTO'
        ELSE 'CRÍTICO'
    END AS risk_description
FROM (
    -- Open Insurance
    SELECT 
        tenant_id, 
        'OPEN_INSURANCE' AS open_x_domain,
        CASE 
            WHEN framework_id = 'SUSEP' THEN 'SUSEP'
            ELSE 'SOLVENCIA_II'
        END AS framework,
        requirement_id,
        is_compliant
    FROM compliance_validators.open_insurance_validations
    
    UNION ALL
    
    -- Open Health
    SELECT 
        tenant_id, 
        'OPEN_HEALTH' AS open_x_domain,
        CASE 
            WHEN framework_id = 'ANS' THEN 'ANS'
            ELSE 'HIPAA_GDPR'
        END AS framework,
        requirement_id,
        is_compliant
    FROM compliance_validators.open_health_validations
    
    UNION ALL
    
    -- Open Government
    SELECT 
        tenant_id, 
        'OPEN_GOVERNMENT' AS open_x_domain,
        CASE 
            WHEN framework_id = 'GOV_BR' THEN 'GOV_BR'
            ELSE 'EIDAS'
        END AS framework,
        requirement_id,
        is_compliant
    FROM compliance_validators.open_government_validations
) AS combined_validations
GROUP BY tenant_id, open_x_domain, framework;

-- View para detalhes de não-conformidades
CREATE OR REPLACE VIEW compliance_validators.vw_open_x_non_compliance_details AS
SELECT
    nc.tenant_id,
    nc.open_x_domain,
    nc.framework,
    nc.requirement_id,
    nc.requirement_name,
    nc.details,
    nc.irr_threshold
FROM (
    -- Open Insurance não-conformidades
    SELECT 
        v.tenant_id, 
        'OPEN_INSURANCE' AS open_x_domain,
        CASE 
            WHEN v.framework_id = 'SUSEP' THEN 'SUSEP'
            ELSE 'SOLVENCIA_II'
        END AS framework,
        v.requirement_id,
        r.requirement_name,
        v.validation_details AS details,
        r.irr_threshold
    FROM compliance_validators.open_insurance_validations v
    JOIN compliance_validators.open_insurance_requirements r 
        ON v.requirement_id = r.requirement_id
    WHERE NOT v.is_compliant
    
    UNION ALL
    
    -- Open Health não-conformidades
    SELECT 
        v.tenant_id, 
        'OPEN_HEALTH' AS open_x_domain,
        CASE 
            WHEN v.framework_id = 'ANS' THEN 'ANS'
            ELSE 'HIPAA_GDPR'
        END AS framework,
        v.requirement_id,
        r.requirement_name,
        v.validation_details AS details,
        r.irr_threshold
    FROM compliance_validators.open_health_validations v
    JOIN compliance_validators.open_health_requirements r 
        ON v.requirement_id = r.requirement_id
    WHERE NOT v.is_compliant
    
    UNION ALL
    
    -- Open Government não-conformidades
    SELECT 
        v.tenant_id, 
        'OPEN_GOVERNMENT' AS open_x_domain,
        CASE 
            WHEN v.framework_id = 'GOV_BR' THEN 'GOV_BR'
            ELSE 'EIDAS'
        END AS framework,
        v.requirement_id,
        r.requirement_name,
        v.validation_details AS details,
        r.irr_threshold
    FROM compliance_validators.open_government_validations v
    JOIN compliance_validators.open_government_requirements r 
        ON v.requirement_id = r.requirement_id
    WHERE NOT v.is_compliant
) AS nc;

-- View para impacto econômico por domínio e framework
CREATE OR REPLACE VIEW compliance_validators.vw_open_x_economic_impact AS
WITH non_compliance_count AS (
    SELECT
        tenant_id,
        open_x_domain,
        framework,
        COUNT(*) AS non_compliant_count
    FROM compliance_validators.vw_open_x_non_compliance_details
    GROUP BY tenant_id, open_x_domain, framework
),
total_requirements AS (
    SELECT
        tenant_id,
        open_x_domain,
        framework,
        total_requirements
    FROM compliance_validators.vw_open_x_compliance_summary
),
economic_impact AS (
    SELECT
        nc.tenant_id,
        nc.open_x_domain,
        nc.framework,
        SUM(CASE
            WHEN ei.impact_amount IS NOT NULL THEN ei.impact_amount
            ELSE economic_planning.calculate_default_impact(nc.open_x_domain, nc.framework, nc.irr_threshold)
        END) AS total_impact
    FROM compliance_validators.vw_open_x_non_compliance_details nc
    LEFT JOIN economic_planning.economic_impacts ei 
        ON nc.tenant_id = ei.tenant_id 
        AND nc.requirement_id = ei.validator_id
    GROUP BY nc.tenant_id, nc.open_x_domain, nc.framework
)
SELECT
    tr.tenant_id,
    tr.open_x_domain,
    tr.framework,
    tr.total_requirements,
    COALESCE(ncc.non_compliant_count, 0) AS non_compliant_count,
    CASE 
        WHEN tr.total_requirements > 0 THEN 
            ROUND((COALESCE(ncc.non_compliant_count, 0)::NUMERIC / tr.total_requirements::NUMERIC) * 100, 2)
        ELSE 0
    END AS non_compliance_percentage,
    COALESCE(ei.total_impact, 0) AS total_economic_impact,
    CASE 
        WHEN COALESCE(ncc.non_compliant_count, 0) > 0 THEN 
            ROUND(COALESCE(ei.total_impact, 0) / COALESCE(ncc.non_compliant_count, 1), 2)
        ELSE 0
    END AS avg_impact_per_non_compliance
FROM total_requirements tr
LEFT JOIN non_compliance_count ncc 
    ON tr.tenant_id = ncc.tenant_id 
    AND tr.open_x_domain = ncc.open_x_domain 
    AND tr.framework = ncc.framework
LEFT JOIN economic_impact ei 
    ON tr.tenant_id = ei.tenant_id 
    AND tr.open_x_domain = ei.open_x_domain 
    AND tr.framework = ei.framework;

-- View para comparação entre domínios Open X
CREATE OR REPLACE VIEW compliance_validators.vw_open_x_domain_comparison AS
SELECT
    tenant_id,
    open_x_domain,
    SUM(total_requirements) AS total_requirements,
    SUM(compliant_requirements) AS compliant_requirements,
    SUM(total_requirements - compliant_requirements) AS non_compliant_requirements,
    ROUND((SUM(compliant_requirements)::NUMERIC / SUM(total_requirements)::NUMERIC) * 100, 2) AS compliance_percentage,
    CASE
        WHEN (SUM(compliant_requirements)::NUMERIC / SUM(total_requirements)::NUMERIC) >= 0.90 THEN 'R1'
        WHEN (SUM(compliant_requirements)::NUMERIC / SUM(total_requirements)::NUMERIC) >= 0.75 THEN 'R2'
        WHEN (SUM(compliant_requirements)::NUMERIC / SUM(total_requirements)::NUMERIC) >= 0.50 THEN 'R3'
        ELSE 'R4'
    END AS risk_level
FROM compliance_validators.vw_open_x_compliance_summary
GROUP BY tenant_id, open_x_domain;

-- View para conformidade por jurisdição
CREATE OR REPLACE VIEW compliance_validators.vw_open_x_jurisdiction_compliance AS
WITH jurisdiction_mapping AS (
    SELECT 
        tenant_id,
        open_x_domain,
        framework,
        CASE 
            WHEN framework IN ('SOLVENCIA_II', 'EIDAS') THEN 'PORTUGAL_UE'
            WHEN framework IN ('SUSEP', 'ANS', 'GOV_BR') THEN 'BRASIL'
            WHEN framework = 'HIPAA_GDPR' THEN 
                CASE 
                    WHEN jurisdiction_context = 'US' THEN 'EUA'
                    ELSE 'PORTUGAL_UE'
                END
            ELSE 'GLOBAL'
        END AS jurisdiction
    FROM (
        -- Open Insurance
        SELECT 
            v.tenant_id, 
            'OPEN_INSURANCE' AS open_x_domain,
            CASE 
                WHEN v.framework_id = 'SUSEP' THEN 'SUSEP'
                ELSE 'SOLVENCIA_II'
            END AS framework,
            v.jurisdiction_context
        FROM compliance_validators.open_insurance_validations v
        
        UNION ALL
        
        -- Open Health
        SELECT 
            v.tenant_id, 
            'OPEN_HEALTH' AS open_x_domain,
            CASE 
                WHEN v.framework_id = 'ANS' THEN 'ANS'
                ELSE 'HIPAA_GDPR'
            END AS framework,
            v.jurisdiction_context
        FROM compliance_validators.open_health_validations v
        
        UNION ALL
        
        -- Open Government
        SELECT 
            v.tenant_id, 
            'OPEN_GOVERNMENT' AS open_x_domain,
            CASE 
                WHEN v.framework_id = 'GOV_BR' THEN 'GOV_BR'
                ELSE 'EIDAS'
            END AS framework,
            v.jurisdiction_context
        FROM compliance_validators.open_government_validations v
    ) AS all_validations
)
SELECT
    jm.tenant_id,
    jm.jurisdiction,
    jm.open_x_domain,
    jm.framework,
    cs.total_requirements,
    cs.compliant_requirements,
    cs.compliance_percentage,
    cs.risk_level,
    cs.risk_description
FROM jurisdiction_mapping jm
JOIN compliance_validators.vw_open_x_compliance_summary cs
    ON jm.tenant_id = cs.tenant_id
    AND jm.open_x_domain = cs.open_x_domain
    AND jm.framework = cs.framework;

-- =============================================
-- Funções para KPIs e Análises
-- =============================================

-- Função para obter métricas de conformidade por domínio
CREATE OR REPLACE FUNCTION compliance_validators.get_open_x_compliance_metrics(
    p_tenant_id UUID,
    p_domain VARCHAR DEFAULT NULL
)
RETURNS TABLE (
    domain VARCHAR,
    total_requirements BIGINT,
    compliant_requirements BIGINT,
    compliance_percentage NUMERIC,
    risk_level VARCHAR,
    risk_description VARCHAR,
    estimated_economic_impact NUMERIC
)
AS $$
BEGIN
    RETURN QUERY
    WITH domain_compliance AS (
        SELECT 
            s.open_x_domain AS domain,
            SUM(s.total_requirements) AS total_requirements,
            SUM(s.compliant_requirements) AS compliant_requirements,
            ROUND((SUM(s.compliant_requirements)::NUMERIC / SUM(s.total_requirements)::NUMERIC) * 100, 2) AS compliance_percentage,
            CASE
                WHEN (SUM(s.compliant_requirements)::NUMERIC / SUM(s.total_requirements)::NUMERIC) >= 0.90 THEN 'R1'
                WHEN (SUM(s.compliant_requirements)::NUMERIC / SUM(s.total_requirements)::NUMERIC) >= 0.75 THEN 'R2'
                WHEN (SUM(s.compliant_requirements)::NUMERIC / SUM(s.total_requirements)::NUMERIC) >= 0.50 THEN 'R3'
                ELSE 'R4'
            END AS risk_level,
            CASE
                WHEN (SUM(s.compliant_requirements)::NUMERIC / SUM(s.total_requirements)::NUMERIC) >= 0.90 THEN 'BAIXO'
                WHEN (SUM(s.compliant_requirements)::NUMERIC / SUM(s.total_requirements)::NUMERIC) >= 0.75 THEN 'MODERADO'
                WHEN (SUM(s.compliant_requirements)::NUMERIC / SUM(s.total_requirements)::NUMERIC) >= 0.50 THEN 'ALTO'
                ELSE 'CRÍTICO'
            END AS risk_description
        FROM compliance_validators.vw_open_x_compliance_summary s
        WHERE s.tenant_id = p_tenant_id
        AND (p_domain IS NULL OR s.open_x_domain = p_domain)
        GROUP BY s.open_x_domain
    ),
    economic_impact AS (
        SELECT 
            e.open_x_domain AS domain,
            SUM(e.total_economic_impact) AS total_impact
        FROM compliance_validators.vw_open_x_economic_impact e
        WHERE e.tenant_id = p_tenant_id
        AND (p_domain IS NULL OR e.open_x_domain = p_domain)
        GROUP BY e.open_x_domain
    )
    SELECT
        dc.domain,
        dc.total_requirements,
        dc.compliant_requirements,
        dc.compliance_percentage,
        dc.risk_level,
        dc.risk_description,
        COALESCE(ei.total_impact, 0) AS estimated_economic_impact
    FROM domain_compliance dc
    LEFT JOIN economic_impact ei ON dc.domain = ei.domain;
END;
$$ LANGUAGE plpgsql;

-- Função para comparar frameworks regulatórios
CREATE OR REPLACE FUNCTION compliance_validators.compare_open_x_frameworks(
    p_tenant_id UUID,
    p_domain VARCHAR
)
RETURNS TABLE (
    framework VARCHAR,
    total_requirements BIGINT,
    compliant_requirements BIGINT,
    compliance_percentage NUMERIC,
    risk_level VARCHAR,
    risk_description VARCHAR,
    economic_impact NUMERIC,
    avg_impact_per_non_compliance NUMERIC
)
AS $$
BEGIN
    RETURN QUERY
    SELECT
        s.framework,
        s.total_requirements,
        s.compliant_requirements,
        s.compliance_percentage,
        s.risk_level,
        s.risk_description,
        COALESCE(e.total_economic_impact, 0) AS economic_impact,
        COALESCE(e.avg_impact_per_non_compliance, 0) AS avg_impact_per_non_compliance
    FROM compliance_validators.vw_open_x_compliance_summary s
    LEFT JOIN compliance_validators.vw_open_x_economic_impact e
        ON s.tenant_id = e.tenant_id
        AND s.open_x_domain = e.open_x_domain
        AND s.framework = e.framework
    WHERE s.tenant_id = p_tenant_id
    AND s.open_x_domain = p_domain
    ORDER BY s.compliance_percentage DESC;
END;
$$ LANGUAGE plpgsql;

-- Função para obter impacto econômico por jurisdição
CREATE OR REPLACE FUNCTION compliance_validators.get_open_x_economic_impact_by_jurisdiction(
    p_tenant_id UUID,
    p_jurisdiction VARCHAR DEFAULT NULL
)
RETURNS TABLE (
    jurisdiction VARCHAR,
    domain VARCHAR,
    framework VARCHAR,
    non_compliant_count BIGINT,
    total_economic_impact NUMERIC,
    avg_impact_per_non_compliance NUMERIC,
    risk_level VARCHAR
)
AS $$
BEGIN
    RETURN QUERY
    WITH jurisdiction_impact AS (
        SELECT
            j.jurisdiction,
            j.open_x_domain AS domain,
            j.framework,
            (j.total_requirements - j.compliant_requirements) AS non_compliant_count,
            e.total_economic_impact,
            e.avg_impact_per_non_compliance,
            j.risk_level
        FROM compliance_validators.vw_open_x_jurisdiction_compliance j
        LEFT JOIN compliance_validators.vw_open_x_economic_impact e
            ON j.tenant_id = e.tenant_id
            AND j.open_x_domain = e.open_x_domain
            AND j.framework = e.framework
        WHERE j.tenant_id = p_tenant_id
        AND (p_jurisdiction IS NULL OR j.jurisdiction = p_jurisdiction)
    )
    SELECT
        ji.jurisdiction,
        ji.domain,
        ji.framework,
        ji.non_compliant_count,
        COALESCE(ji.total_economic_impact, 0) AS total_economic_impact,
        COALESCE(ji.avg_impact_per_non_compliance, 0) AS avg_impact_per_non_compliance,
        ji.risk_level
    FROM jurisdiction_impact ji
    ORDER BY ji.jurisdiction, ji.total_economic_impact DESC;
END;
$$ LANGUAGE plpgsql;

-- Função para simular impacto de melhorias
CREATE OR REPLACE FUNCTION compliance_validators.simulate_open_x_improvements(
    p_tenant_id UUID,
    p_domain VARCHAR DEFAULT NULL,
    p_improvement_percentage INTEGER DEFAULT 30,
    p_months_horizon INTEGER DEFAULT 12
)
RETURNS TABLE (
    domain VARCHAR,
    framework VARCHAR,
    current_compliance_percentage NUMERIC,
    simulated_compliance_percentage NUMERIC,
    current_economic_impact NUMERIC,
    simulated_economic_impact NUMERIC,
    potential_savings NUMERIC,
    estimated_roi NUMERIC
)
AS $$
DECLARE
    v_max_improvement NUMERIC := 0.95; -- Máximo de 95% de conformidade é considerado realista
    v_improvement_factor NUMERIC;
BEGIN
    -- Calcular fator de melhoria (entre 0 e 1)
    v_improvement_factor := p_improvement_percentage / 100.0;
    
    RETURN QUERY
    WITH current_state AS (
        SELECT
            s.open_x_domain AS domain,
            s.framework,
            s.compliance_percentage AS current_compliance,
            e.total_economic_impact AS current_impact,
            -- Estimativa de custo para implementar as melhorias (30% do impacto atual)
            e.total_economic_impact * 0.3 AS estimated_remediation_cost
        FROM compliance_validators.vw_open_x_compliance_summary s
        LEFT JOIN compliance_validators.vw_open_x_economic_impact e
            ON s.tenant_id = e.tenant_id
            AND s.open_x_domain = e.open_x_domain
            AND s.framework = e.framework
        WHERE s.tenant_id = p_tenant_id
        AND (p_domain IS NULL OR s.open_x_domain = p_domain)
    ),
    simulated_state AS (
        SELECT
            cs.domain,
            cs.framework,
            cs.current_compliance,
            -- Calcular conformidade simulada (limitada a v_max_improvement)
            LEAST(cs.current_compliance + ((100 - cs.current_compliance) * v_improvement_factor), v_max_improvement * 100) AS simulated_compliance,
            cs.current_impact,
            -- Calcular impacto econômico simulado após melhorias
            cs.current_impact * (1 - v_improvement_factor) AS simulated_impact,
            cs.estimated_remediation_cost
        FROM current_state cs
    )
    SELECT
        ss.domain,
        ss.framework,
        ss.current_compliance,
        ss.simulated_compliance,
        ss.current_impact,
        ss.simulated_impact,
        (ss.current_impact - ss.simulated_impact) AS potential_savings,
        CASE 
            WHEN ss.estimated_remediation_cost > 0 THEN
                -- ROI calculado para o horizonte de meses especificado
                (((ss.current_impact - ss.simulated_impact) * (p_months_horizon / 12.0)) / ss.estimated_remediation_cost) * 100
            ELSE 0
        END AS estimated_roi
    FROM simulated_state ss
    ORDER BY ss.domain, estimated_roi DESC;
END;
$$ LANGUAGE plpgsql;

-- Função auxiliar para calcular impacto econômico padrão
CREATE OR REPLACE FUNCTION economic_planning.calculate_default_impact(
    p_domain VARCHAR,
    p_framework VARCHAR,
    p_irr_threshold VARCHAR
)
RETURNS NUMERIC AS $$
DECLARE
    v_base_impact NUMERIC;
    v_multiplier NUMERIC;
BEGIN
    -- Definir impacto base de acordo com o domínio
    CASE p_domain
        WHEN 'OPEN_INSURANCE' THEN v_base_impact := 75000;
        WHEN 'OPEN_HEALTH' THEN v_base_impact := 50000;
        WHEN 'OPEN_GOVERNMENT' THEN v_base_impact := 100000;
        ELSE v_base_impact := 25000;
    END CASE;
    
    -- Aplicar multiplicador de acordo com o framework
    CASE p_framework
        WHEN 'SOLVENCIA_II' THEN v_multiplier := 1.5;
        WHEN 'SUSEP' THEN v_multiplier := 1.2;
        WHEN 'HIPAA_GDPR' THEN v_multiplier := 1.8;
        WHEN 'ANS' THEN v_multiplier := 1.1;
        WHEN 'EIDAS' THEN v_multiplier := 1.4;
        WHEN 'GOV_BR' THEN v_multiplier := 1.0;
        ELSE v_multiplier := 1.0;
    END CASE;
    
    -- Aplicar ajuste baseado no threshold de IRR (Inherent Risk Rating)
    CASE p_irr_threshold
        WHEN 'R1' THEN v_multiplier := v_multiplier * 0.5;
        WHEN 'R2' THEN v_multiplier := v_multiplier * 1.0;
        WHEN 'R3' THEN v_multiplier := v_multiplier * 2.0;
        WHEN 'R4' THEN v_multiplier := v_multiplier * 4.0;
        ELSE v_multiplier := v_multiplier * 1.0;
    END CASE;
    
    RETURN v_base_impact * v_multiplier;
END;
$$ LANGUAGE plpgsql;

-- =============================================
-- Configuração de Permissões
-- =============================================

-- Conceder permissões para os roles necessários
DO $$
BEGIN
    -- Verificar se os roles existem antes de conceder permissões
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'economic_analyst_role') THEN
        -- Permissões para analistas econômicos
        GRANT USAGE ON SCHEMA compliance_validators TO economic_analyst_role;
        GRANT USAGE ON SCHEMA economic_planning TO economic_analyst_role;
        
        GRANT SELECT ON ALL TABLES IN SCHEMA compliance_validators TO economic_analyst_role;
        GRANT SELECT ON ALL TABLES IN SCHEMA economic_planning TO economic_analyst_role;
        
        GRANT EXECUTE ON FUNCTION compliance_validators.get_open_x_compliance_metrics TO economic_analyst_role;
        GRANT EXECUTE ON FUNCTION compliance_validators.compare_open_x_frameworks TO economic_analyst_role;
        GRANT EXECUTE ON FUNCTION compliance_validators.get_open_x_economic_impact_by_jurisdiction TO economic_analyst_role;
        GRANT EXECUTE ON FUNCTION compliance_validators.simulate_open_x_improvements TO economic_analyst_role;
        GRANT EXECUTE ON FUNCTION economic_planning.calculate_default_impact TO economic_analyst_role;
    END IF;
    
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'compliance_manager_role') THEN
        -- Permissões para gestores de conformidade
        GRANT USAGE ON SCHEMA compliance_validators TO compliance_manager_role;
        
        GRANT SELECT ON ALL TABLES IN SCHEMA compliance_validators TO compliance_manager_role;
        
        GRANT EXECUTE ON FUNCTION compliance_validators.get_open_x_compliance_metrics TO compliance_manager_role;
        GRANT EXECUTE ON FUNCTION compliance_validators.compare_open_x_frameworks TO compliance_manager_role;
    END IF;
    
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'risk_analyst_role') THEN
        -- Permissões para analistas de risco
        GRANT USAGE ON SCHEMA compliance_validators TO risk_analyst_role;
        
        GRANT SELECT ON ALL TABLES IN SCHEMA compliance_validators TO risk_analyst_role;
        
        GRANT EXECUTE ON FUNCTION compliance_validators.get_open_x_compliance_metrics TO risk_analyst_role;
        GRANT EXECUTE ON FUNCTION compliance_validators.get_open_x_economic_impact_by_jurisdiction TO risk_analyst_role;
    END IF;
    
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'dashboard_viewer_role') THEN
        -- Permissões para visualizadores do dashboard
        GRANT USAGE ON SCHEMA compliance_validators TO dashboard_viewer_role;
        
        GRANT SELECT ON compliance_validators.vw_open_x_compliance_summary TO dashboard_viewer_role;
        GRANT SELECT ON compliance_validators.vw_open_x_non_compliance_details TO dashboard_viewer_role;
        GRANT SELECT ON compliance_validators.vw_open_x_economic_impact TO dashboard_viewer_role;
        GRANT SELECT ON compliance_validators.vw_open_x_domain_comparison TO dashboard_viewer_role;
        GRANT SELECT ON compliance_validators.vw_open_x_jurisdiction_compliance TO dashboard_viewer_role;
    END IF;
END $$;

-- =============================================
-- Notificar Conclusão
-- =============================================

DO $$
BEGIN
    RAISE NOTICE 'Dashboard Open X configurado com sucesso!';
    RAISE NOTICE 'Visualizações criadas:';
    RAISE NOTICE '- vw_open_x_compliance_summary';
    RAISE NOTICE '- vw_open_x_non_compliance_details';
    RAISE NOTICE '- vw_open_x_economic_impact';
    RAISE NOTICE '- vw_open_x_domain_comparison';
    RAISE NOTICE '- vw_open_x_jurisdiction_compliance';
    RAISE NOTICE 'Funções criadas:';
    RAISE NOTICE '- get_open_x_compliance_metrics()';
    RAISE NOTICE '- compare_open_x_frameworks()';
    RAISE NOTICE '- get_open_x_economic_impact_by_jurisdiction()';
    RAISE NOTICE '- simulate_open_x_improvements()';
    RAISE NOTICE '- calculate_default_impact()';
END $$;
