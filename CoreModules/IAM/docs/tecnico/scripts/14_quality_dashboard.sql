-- =============================================================================
-- Script de Dashboard de Qualidade e Conformidade
-- =============================================================================
-- Autor: Eduardo Jeremias
-- Data: 15/05/2025
-- Versão: 1.0
-- Descrição: Este script implementa visões e funções para suportar o
--            dashboard de qualidade e conformidade, integrando-se com
--            os validadores de IAM e o Sistema de Gestão da Qualidade.
-- =============================================================================

-- Verificação de ambiente
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'quality_management') THEN
        RAISE EXCEPTION 'Schema quality_management não existe. Execute os scripts anteriores primeiro.';
    END IF;
END
$$;

-- =============================================================================
-- 1. Visões para o Dashboard de Qualidade
-- =============================================================================

-- Visão para resumo de métricas de qualidade
CREATE OR REPLACE VIEW quality_management.dashboard_quality_metrics AS
SELECT
    m.metric_name,
    m.metric_type,
    m.standard_id,
    COALESCE(s.standard_name, 'Geral') AS standard_name,
    m.current_value,
    m.target_value,
    m.unit,
    CASE 
        WHEN m.metric_type = 'PERCENT' AND m.current_value >= m.target_value THEN 'OK'
        WHEN m.metric_type = 'PERCENT' AND m.current_value >= (m.target_value * 0.8) THEN 'ATENÇÃO'
        WHEN m.metric_type = 'PERCENT' AND m.current_value < (m.target_value * 0.8) THEN 'CRÍTICO'
        WHEN m.metric_type = 'COUNT' AND m.current_value <= m.target_value THEN 'OK'
        WHEN m.metric_type = 'COUNT' AND m.current_value <= (m.target_value * 1.5) THEN 'ATENÇÃO'
        WHEN m.metric_type = 'COUNT' AND m.current_value > (m.target_value * 1.5) THEN 'CRÍTICO'
        WHEN m.metric_type = 'TIME' AND m.current_value <= m.target_value THEN 'OK'
        WHEN m.metric_type = 'TIME' AND m.current_value <= (m.target_value * 1.5) THEN 'ATENÇÃO'
        WHEN m.metric_type = 'TIME' AND m.current_value > (m.target_value * 1.5) THEN 'CRÍTICO'
        ELSE 'INDETERMINADO'
    END AS status,
    m.last_calculated,
    m.tenant_id
FROM quality_management.quality_metrics m
LEFT JOIN quality_management.quality_standards s ON m.standard_id = s.standard_id;

COMMENT ON VIEW quality_management.dashboard_quality_metrics IS 'Visão para exibição de métricas de qualidade no dashboard';

-- Visão para resumo de não-conformidades
CREATE OR REPLACE VIEW quality_management.dashboard_non_conformities_summary AS
SELECT
    nc.standard_id,
    nc.standard_name,
    nc.impact_level,
    nc.status,
    COUNT(*) AS count,
    MIN(nc.created_at) AS earliest_date,
    MAX(nc.created_at) AS latest_date,
    nc.tenant_id
FROM quality_management.non_conformity nc
GROUP BY 
    nc.standard_id,
    nc.standard_name,
    nc.impact_level,
    nc.status,
    nc.tenant_id;

COMMENT ON VIEW quality_management.dashboard_non_conformities_summary IS 'Visão para resumo de não-conformidades agrupadas por padrão, impacto e status';

-- Visão para o status das ações corretivas
CREATE OR REPLACE VIEW quality_management.dashboard_corrective_actions_summary AS
SELECT
    nc.standard_id,
    nc.standard_name,
    ca.status,
    COUNT(*) AS count,
    COUNT(CASE WHEN ca.due_date < CURRENT_TIMESTAMP AND ca.status != 'COMPLETED' THEN 1 END) AS overdue_count,
    MIN(ca.due_date) AS earliest_due_date,
    MAX(ca.due_date) AS latest_due_date,
    ca.tenant_id
FROM quality_management.corrective_action ca
JOIN quality_management.non_conformity nc ON ca.non_conformity_id = nc.non_conformity_id
GROUP BY 
    nc.standard_id,
    nc.standard_name,
    ca.status,
    ca.tenant_id;

COMMENT ON VIEW quality_management.dashboard_corrective_actions_summary IS 'Visão para resumo de ações corretivas agrupadas por padrão e status';

-- Visão para tendências de não-conformidades (últimos 12 meses)
CREATE OR REPLACE VIEW quality_management.dashboard_non_conformity_trends AS
WITH months AS (
    SELECT generate_series(
        date_trunc('month', CURRENT_DATE - INTERVAL '11 months'),
        date_trunc('month', CURRENT_DATE),
        '1 month'::interval
    ) AS month_start
),
standards AS (
    SELECT DISTINCT standard_id, standard_name, tenant_id
    FROM quality_management.non_conformity
),
base_data AS (
    SELECT 
        m.month_start,
        s.standard_id,
        s.standard_name,
        s.tenant_id
    FROM months m
    CROSS JOIN standards s
),
monthly_counts AS (
    SELECT 
        date_trunc('month', created_at) AS month_start,
        standard_id,
        standard_name,
        COUNT(*) AS non_conformity_count,
        tenant_id
    FROM quality_management.non_conformity
    WHERE created_at >= CURRENT_DATE - INTERVAL '11 months'
    GROUP BY 
        date_trunc('month', created_at),
        standard_id,
        standard_name,
        tenant_id
)
SELECT 
    bd.month_start,
    bd.standard_id,
    bd.standard_name,
    COALESCE(mc.non_conformity_count, 0) AS non_conformity_count,
    bd.tenant_id
FROM base_data bd
LEFT JOIN monthly_counts mc ON 
    bd.month_start = mc.month_start AND 
    bd.standard_id = mc.standard_id AND
    bd.tenant_id = mc.tenant_id
ORDER BY 
    bd.standard_id, 
    bd.month_start;

COMMENT ON VIEW quality_management.dashboard_non_conformity_trends IS 'Visão para tendências de não-conformidades nos últimos 12 meses';

-- Visão para eficácia das ações corretivas
CREATE OR REPLACE VIEW quality_management.dashboard_corrective_action_effectiveness AS
SELECT
    nc.standard_id,
    nc.standard_name,
    ca.effectiveness_status,
    COUNT(*) AS count,
    AVG(EXTRACT(EPOCH FROM (ca.completed_date - ca.created_at))/86400)::NUMERIC(10,2) AS avg_days_to_complete,
    ca.tenant_id
FROM quality_management.corrective_action ca
JOIN quality_management.non_conformity nc ON ca.non_conformity_id = nc.non_conformity_id
WHERE ca.status = 'COMPLETED'
AND ca.effectiveness_status IS NOT NULL
GROUP BY 
    nc.standard_id,
    nc.standard_name,
    ca.effectiveness_status,
    ca.tenant_id;

COMMENT ON VIEW quality_management.dashboard_corrective_action_effectiveness IS 'Visão para análise de eficácia das ações corretivas';

-- =============================================================================
-- 2. Funções para Dados do Dashboard
-- =============================================================================

-- Função para obter indicadores principais do dashboard
CREATE OR REPLACE FUNCTION quality_management.get_dashboard_kpis(
    p_tenant_id UUID
) RETURNS TABLE (
    kpi_name VARCHAR(100),
    kpi_value NUMERIC,
    kpi_target NUMERIC,
    kpi_unit VARCHAR(50),
    kpi_status VARCHAR(20)
) AS $$
BEGIN
    RETURN QUERY
    
    -- Taxa de conformidade geral
    SELECT 
        'Taxa de Conformidade'::VARCHAR(100) AS kpi_name,
        COALESCE(
            (SELECT current_value FROM quality_management.quality_metrics 
             WHERE metric_name = 'COMPLIANCE_RATE' 
             AND tenant_id = p_tenant_id
             LIMIT 1), 
            0
        ) AS kpi_value,
        100::NUMERIC AS kpi_target,
        '%'::VARCHAR(50) AS kpi_unit,
        CASE 
            WHEN COALESCE(
                (SELECT current_value FROM quality_management.quality_metrics 
                 WHERE metric_name = 'COMPLIANCE_RATE' 
                 AND tenant_id = p_tenant_id
                 LIMIT 1), 
                0
            ) >= 90 THEN 'OK'
            WHEN COALESCE(
                (SELECT current_value FROM quality_management.quality_metrics 
                 WHERE metric_name = 'COMPLIANCE_RATE' 
                 AND tenant_id = p_tenant_id
                 LIMIT 1), 
                0
            ) >= 75 THEN 'ATENÇÃO'
            ELSE 'CRÍTICO'
        END AS kpi_status
    
    UNION ALL
    
    -- Não-conformidades abertas
    SELECT 
        'Não-Conformidades Abertas'::VARCHAR(100) AS kpi_name,
        COALESCE(
            (SELECT COUNT(*) FROM quality_management.non_conformity 
             WHERE status = 'OPEN' 
             AND tenant_id = p_tenant_id), 
            0
        ) AS kpi_value,
        0::NUMERIC AS kpi_target,
        'unidades'::VARCHAR(50) AS kpi_unit,
        CASE 
            WHEN COALESCE(
                (SELECT COUNT(*) FROM quality_management.non_conformity 
                 WHERE status = 'OPEN' 
                 AND tenant_id = p_tenant_id), 
                0
            ) = 0 THEN 'OK'
            WHEN COALESCE(
                (SELECT COUNT(*) FROM quality_management.non_conformity 
                 WHERE status = 'OPEN' 
                 AND tenant_id = p_tenant_id), 
                0
            ) <= 5 THEN 'ATENÇÃO'
            ELSE 'CRÍTICO'
        END AS kpi_status
    
    UNION ALL
    
    -- Ações corretivas pendentes
    SELECT 
        'Ações Corretivas Pendentes'::VARCHAR(100) AS kpi_name,
        COALESCE(
            (SELECT COUNT(*) FROM quality_management.corrective_action 
             WHERE status = 'PENDING' 
             AND tenant_id = p_tenant_id), 
            0
        ) AS kpi_value,
        0::NUMERIC AS kpi_target,
        'unidades'::VARCHAR(50) AS kpi_unit,
        CASE 
            WHEN COALESCE(
                (SELECT COUNT(*) FROM quality_management.corrective_action 
                 WHERE status = 'PENDING' 
                 AND tenant_id = p_tenant_id), 
                0
            ) = 0 THEN 'OK'
            WHEN COALESCE(
                (SELECT COUNT(*) FROM quality_management.corrective_action 
                 WHERE status = 'PENDING' 
                 AND tenant_id = p_tenant_id), 
                0
            ) <= 5 THEN 'ATENÇÃO'
            ELSE 'CRÍTICO'
        END AS kpi_status
    
    UNION ALL
    
    -- Ações corretivas atrasadas
    SELECT 
        'Ações Corretivas Atrasadas'::VARCHAR(100) AS kpi_name,
        COALESCE(
            (SELECT COUNT(*) FROM quality_management.corrective_action 
             WHERE status != 'COMPLETED' 
             AND due_date < CURRENT_TIMESTAMP
             AND tenant_id = p_tenant_id), 
            0
        ) AS kpi_value,
        0::NUMERIC AS kpi_target,
        'unidades'::VARCHAR(50) AS kpi_unit,
        CASE 
            WHEN COALESCE(
                (SELECT COUNT(*) FROM quality_management.corrective_action 
                 WHERE status != 'COMPLETED' 
                 AND due_date < CURRENT_TIMESTAMP
                 AND tenant_id = p_tenant_id), 
                0
            ) = 0 THEN 'OK'
            WHEN COALESCE(
                (SELECT COUNT(*) FROM quality_management.corrective_action 
                 WHERE status != 'COMPLETED' 
                 AND due_date < CURRENT_TIMESTAMP
                 AND tenant_id = p_tenant_id), 
                0
            ) <= 2 THEN 'ATENÇÃO'
            ELSE 'CRÍTICO'
        END AS kpi_status
    
    UNION ALL
    
    -- Tempo médio de resolução
    SELECT 
        'Tempo Médio de Resolução'::VARCHAR(100) AS kpi_name,
        COALESCE(
            (SELECT current_value FROM quality_management.quality_metrics 
             WHERE metric_name = 'AVG_RESOLUTION_TIME' 
             AND tenant_id = p_tenant_id
             LIMIT 1), 
            0
        ) AS kpi_value,
        10::NUMERIC AS kpi_target,
        'dias'::VARCHAR(50) AS kpi_unit,
        CASE 
            WHEN COALESCE(
                (SELECT current_value FROM quality_management.quality_metrics 
                 WHERE metric_name = 'AVG_RESOLUTION_TIME' 
                 AND tenant_id = p_tenant_id
                 LIMIT 1), 
                0
            ) <= 10 THEN 'OK'
            WHEN COALESCE(
                (SELECT current_value FROM quality_management.quality_metrics 
                 WHERE metric_name = 'AVG_RESOLUTION_TIME' 
                 AND tenant_id = p_tenant_id
                 LIMIT 1), 
                0
            ) <= 15 THEN 'ATENÇÃO'
            ELSE 'CRÍTICO'
        END AS kpi_status;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION quality_management.get_dashboard_kpis IS 'Função para obter os KPIs principais do dashboard';

-- Função para obter distribuição de não-conformidades por padrão
CREATE OR REPLACE FUNCTION quality_management.get_non_conformities_by_standard(
    p_tenant_id UUID,
    p_status VARCHAR(20) DEFAULT NULL
) RETURNS TABLE (
    standard_id VARCHAR(50),
    standard_name VARCHAR(255),
    non_conformity_count INTEGER,
    percentage NUMERIC(5,2)
) AS $$
DECLARE
    v_total INTEGER;
BEGIN
    -- Calcular total de não-conformidades
    SELECT COUNT(*) INTO v_total
    FROM quality_management.non_conformity
    WHERE tenant_id = p_tenant_id
    AND (p_status IS NULL OR status = p_status);
    
    -- Se não houver não-conformidades, retornar vazio
    IF v_total = 0 THEN
        RETURN;
    END IF;
    
    RETURN QUERY
    SELECT 
        nc.standard_id,
        nc.standard_name,
        COUNT(*)::INTEGER AS non_conformity_count,
        (COUNT(*)::NUMERIC / v_total * 100)::NUMERIC(5,2) AS percentage
    FROM quality_management.non_conformity nc
    WHERE nc.tenant_id = p_tenant_id
    AND (p_status IS NULL OR nc.status = p_status)
    GROUP BY nc.standard_id, nc.standard_name
    ORDER BY COUNT(*) DESC;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION quality_management.get_non_conformities_by_standard IS 'Função para obter distribuição de não-conformidades por padrão';

-- Função para obter distribuição de não-conformidades por impacto
CREATE OR REPLACE FUNCTION quality_management.get_non_conformities_by_impact(
    p_tenant_id UUID,
    p_status VARCHAR(20) DEFAULT NULL
) RETURNS TABLE (
    impact_level VARCHAR(20),
    impact_label VARCHAR(50),
    non_conformity_count INTEGER,
    percentage NUMERIC(5,2)
) AS $$
DECLARE
    v_total INTEGER;
BEGIN
    -- Calcular total de não-conformidades
    SELECT COUNT(*) INTO v_total
    FROM quality_management.non_conformity
    WHERE tenant_id = p_tenant_id
    AND (p_status IS NULL OR status = p_status);
    
    -- Se não houver não-conformidades, retornar vazio
    IF v_total = 0 THEN
        RETURN;
    END IF;
    
    RETURN QUERY
    SELECT 
        nc.impact_level,
        CASE
            WHEN nc.impact_level = 'MAJOR' THEN 'Crítico'
            WHEN nc.impact_level = 'MODERATE' THEN 'Moderado'
            WHEN nc.impact_level = 'MINOR' THEN 'Menor'
            ELSE 'Desconhecido'
        END AS impact_label,
        COUNT(*)::INTEGER AS non_conformity_count,
        (COUNT(*)::NUMERIC / v_total * 100)::NUMERIC(5,2) AS percentage
    FROM quality_management.non_conformity nc
    WHERE nc.tenant_id = p_tenant_id
    AND (p_status IS NULL OR nc.status = p_status)
    GROUP BY nc.impact_level
    ORDER BY 
        CASE
            WHEN nc.impact_level = 'MAJOR' THEN 1
            WHEN nc.impact_level = 'MODERATE' THEN 2
            WHEN nc.impact_level = 'MINOR' THEN 3
            ELSE 4
        END;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION quality_management.get_non_conformities_by_impact IS 'Função para obter distribuição de não-conformidades por nível de impacto';

-- Função para obter as não-conformidades mais recentes
CREATE OR REPLACE FUNCTION quality_management.get_recent_non_conformities(
    p_tenant_id UUID,
    p_limit INTEGER DEFAULT 10
) RETURNS TABLE (
    non_conformity_id UUID,
    standard_name VARCHAR(255),
    requirement_name VARCHAR(255),
    impact_level VARCHAR(20),
    status VARCHAR(20),
    creation_date TIMESTAMP WITH TIME ZONE,
    action_count INTEGER,
    pending_action_count INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        nc.non_conformity_id,
        nc.standard_name,
        nc.requirement_name,
        nc.impact_level,
        nc.status,
        nc.created_at AS creation_date,
        COUNT(ca.action_id)::INTEGER AS action_count,
        COUNT(CASE WHEN ca.status != 'COMPLETED' THEN 1 END)::INTEGER AS pending_action_count
    FROM quality_management.non_conformity nc
    LEFT JOIN quality_management.corrective_action ca ON nc.non_conformity_id = ca.non_conformity_id
    WHERE nc.tenant_id = p_tenant_id
    GROUP BY 
        nc.non_conformity_id,
        nc.standard_name,
        nc.requirement_name,
        nc.impact_level,
        nc.status,
        nc.created_at
    ORDER BY nc.created_at DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION quality_management.get_recent_non_conformities IS 'Função para obter as não-conformidades mais recentes';

-- =============================================================================
-- 3. Definição de Permissões
-- =============================================================================

-- Garantir acesso às visões e funções para os grupos de usuários
DO $$
BEGIN
    -- Permissões para visões do dashboard
    EXECUTE 'GRANT SELECT ON quality_management.dashboard_quality_metrics TO quality_manager_role';
    EXECUTE 'GRANT SELECT ON quality_management.dashboard_non_conformities_summary TO quality_manager_role';
    EXECUTE 'GRANT SELECT ON quality_management.dashboard_corrective_actions_summary TO quality_manager_role';
    EXECUTE 'GRANT SELECT ON quality_management.dashboard_non_conformity_trends TO quality_manager_role';
    EXECUTE 'GRANT SELECT ON quality_management.dashboard_corrective_action_effectiveness TO quality_manager_role';
    
    -- Permissões para funções do dashboard
    EXECUTE 'GRANT EXECUTE ON FUNCTION quality_management.get_dashboard_kpis(UUID) TO quality_manager_role';
    EXECUTE 'GRANT EXECUTE ON FUNCTION quality_management.get_non_conformities_by_standard(UUID, VARCHAR) TO quality_manager_role';
    EXECUTE 'GRANT EXECUTE ON FUNCTION quality_management.get_non_conformities_by_impact(UUID, VARCHAR) TO quality_manager_role';
    EXECUTE 'GRANT EXECUTE ON FUNCTION quality_management.get_recent_non_conformities(UUID, INTEGER) TO quality_manager_role';
    
    -- Permissões para usuário somente leitura
    EXECUTE 'GRANT SELECT ON quality_management.dashboard_quality_metrics TO quality_viewer_role';
    EXECUTE 'GRANT SELECT ON quality_management.dashboard_non_conformities_summary TO quality_viewer_role';
    EXECUTE 'GRANT SELECT ON quality_management.dashboard_corrective_actions_summary TO quality_viewer_role';
    EXECUTE 'GRANT SELECT ON quality_management.dashboard_non_conformity_trends TO quality_viewer_role';
    EXECUTE 'GRANT SELECT ON quality_management.dashboard_corrective_action_effectiveness TO quality_viewer_role';
    
    EXECUTE 'GRANT EXECUTE ON FUNCTION quality_management.get_dashboard_kpis(UUID) TO quality_viewer_role';
    EXECUTE 'GRANT EXECUTE ON FUNCTION quality_management.get_non_conformities_by_standard(UUID, VARCHAR) TO quality_viewer_role';
    EXECUTE 'GRANT EXECUTE ON FUNCTION quality_management.get_non_conformities_by_impact(UUID, VARCHAR) TO quality_viewer_role';
    EXECUTE 'GRANT EXECUTE ON FUNCTION quality_management.get_recent_non_conformities(UUID, INTEGER) TO quality_viewer_role';
EXCEPTION
    WHEN undefined_object THEN
        RAISE NOTICE 'Algumas roles podem não existir. Crie-as antes de conceder permissões.';
END;
$$;

-- =============================================================================
-- 4. Comentários Finais
-- =============================================================================

COMMENT ON SCHEMA quality_management IS 'Schema para o Sistema de Gestão da Qualidade integrado com validadores IAM';

-- =============================================================================
-- Fim do Script
-- =============================================================================
