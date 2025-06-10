-- INNOVABIZ - IAM ML/Analytics Healthcare Functions (Part 1)
-- Author: Eduardo Jeremias
-- Date: 08/05/2025
-- Version: 1.0
-- Description: Funções de ML/analytics específicas para o setor de saúde, com foco em compliance.

-- Configurar caminho de busca
SET search_path TO iam_analytics, iam, public;

-- ===============================================================================
-- FUNÇÕES PARA ANÁLISE DE COMPLIANCE EM SAÚDE
-- ===============================================================================

-- Função para extrair features de compliance ISO 27001 para ML
CREATE OR REPLACE FUNCTION extract_iso27001_compliance_features(
    p_organization_id UUID,
    p_assessment_id UUID DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    v_feature_id UUID;
    v_assessment_id UUID;
    v_features JSONB;
    v_feature_vector FLOAT[];
    v_assessment_data JSONB;
    v_control_scores JSONB;
    v_risk_scores JSONB;
    v_compliance_metrics JSONB;
BEGIN
    -- Se não foi fornecido um assessment_id, pegar o mais recente
    IF p_assessment_id IS NULL THEN
        SELECT id INTO v_assessment_id
        FROM iam.iso_assessments
        WHERE organization_id = p_organization_id
        ORDER BY created_at DESC
        LIMIT 1;
    ELSE
        v_assessment_id := p_assessment_id;
    END IF;
    
    IF v_assessment_id IS NULL THEN
        RAISE EXCEPTION 'Nenhuma avaliação ISO 27001 encontrada para a organização';
    END IF;
    
    -- Coletar dados do assessment
    SELECT 
        jsonb_build_object(
            'assessment_id', a.id,
            'assessment_date', a.assessment_date,
            'status', a.status,
            'score', a.overall_score,
            'assessor', a.assessor_id,
            'completion_percentage', a.completion_percentage
        ) INTO v_assessment_data
    FROM iam.iso_assessments a
    WHERE a.id = v_assessment_id;
    
    -- Coletar pontuações por controle
    SELECT 
        jsonb_object_agg(
            c.control_id,
            jsonb_build_object(
                'score', c.score,
                'implementation_status', c.implementation_status,
                'priority', c.priority,
                'has_evidence', (c.evidence IS NOT NULL AND c.evidence != '[]'::jsonb)
            )
        ) INTO v_control_scores
    FROM iam.iso_control_results c
    WHERE c.assessment_id = v_assessment_id;
    
    -- Agregar métricas de risco
    SELECT
        jsonb_build_object(
            'high_risk_controls', COUNT(*) FILTER (WHERE c.priority = 'high' AND c.score < 0.6),
            'medium_risk_controls', COUNT(*) FILTER (WHERE c.priority = 'medium' AND c.score < 0.7),
            'low_risk_controls', COUNT(*) FILTER (WHERE c.priority = 'low' AND c.score < 0.8),
            'critical_gaps', COUNT(*) FILTER (WHERE c.score < 0.4),
            'avg_score_by_domain', jsonb_object_agg(
                ic.domain,
                AVG(c.score)
            )
        ) INTO v_risk_scores
    FROM 
        iam.iso_control_results c
        JOIN iam.iso_controls ic ON c.control_id = ic.id
    WHERE 
        c.assessment_id = v_assessment_id
    GROUP BY 
        1, 2, 3, 4;
        
    -- Calcular métricas de compliance
    SELECT
        jsonb_build_object(
            'total_controls', COUNT(*),
            'implemented_controls', COUNT(*) FILTER (WHERE c.implementation_status = 'implemented'),
            'partially_implemented', COUNT(*) FILTER (WHERE c.implementation_status = 'partially_implemented'),
            'planned', COUNT(*) FILTER (WHERE c.implementation_status = 'planned'),
            'not_implemented', COUNT(*) FILTER (WHERE c.implementation_status = 'not_implemented'),
            'not_applicable', COUNT(*) FILTER (WHERE c.implementation_status = 'not_applicable'),
            'implementation_rate', 
                (COUNT(*) FILTER (WHERE c.implementation_status = 'implemented')::FLOAT / 
                NULLIF(COUNT(*) FILTER (WHERE c.implementation_status != 'not_applicable'), 0))::FLOAT,
            'compliance_score',
                SUM(c.score) / NULLIF(COUNT(*) FILTER (WHERE c.implementation_status != 'not_applicable'), 0)::FLOAT
        ) INTO v_compliance_metrics
    FROM 
        iam.iso_control_results c
    WHERE 
        c.assessment_id = v_assessment_id;
    
    -- Construir o objeto de features
    v_features := jsonb_build_object(
        'assessment', v_assessment_data,
        'control_scores', v_control_scores,
        'risk_metrics', v_risk_scores,
        'compliance_metrics', v_compliance_metrics,
        'timestamp', NOW()
    );
    
    -- Transformar algumas características numéricas em um vetor
    v_feature_vector := ARRAY[
        (v_compliance_metrics->>'implementation_rate')::FLOAT,
        (v_compliance_metrics->>'compliance_score')::FLOAT,
        (v_risk_scores->>'high_risk_controls')::FLOAT,
        (v_risk_scores->>'medium_risk_controls')::FLOAT,
        (v_risk_scores->>'low_risk_controls')::FLOAT,
        (v_risk_scores->>'critical_gaps')::FLOAT,
        (v_assessment_data->>'completion_percentage')::FLOAT,
        (v_assessment_data->>'score')::FLOAT
    ];
    
    -- Inserir no feature store
    INSERT INTO ml_feature_store (
        feature_set_name,
        entity_type,
        entity_id,
        feature_data,
        feature_vector,
        timestamp,
        metadata
    ) VALUES (
        'iso27001_compliance',
        'organizations',
        p_organization_id,
        v_features,
        v_feature_vector,
        NOW(),
        jsonb_build_object(
            'assessment_id', v_assessment_id,
            'feature_type', 'compliance',
            'standard', 'ISO 27001',
            'sector', 'healthcare'
        )
    ) RETURNING id INTO v_feature_id;
    
    RETURN v_feature_id;
END;
$$ LANGUAGE plpgsql;

-- Função para extrair features de compliance HIMSS EMRAM para ML
CREATE OR REPLACE FUNCTION extract_himss_emram_compliance_features(
    p_organization_id UUID,
    p_assessment_id UUID DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    v_feature_id UUID;
    v_assessment_id UUID;
    v_features JSONB;
    v_feature_vector FLOAT[];
    v_assessment_data JSONB;
    v_stage_scores JSONB;
    v_criteria_scores JSONB;
    v_compliance_metrics JSONB;
BEGIN
    -- Se não foi fornecido um assessment_id, pegar o mais recente
    IF p_assessment_id IS NULL THEN
        SELECT id INTO v_assessment_id
        FROM iam.himss_assessments
        WHERE organization_id = p_organization_id
        ORDER BY created_at DESC
        LIMIT 1;
    ELSE
        v_assessment_id := p_assessment_id;
    END IF;
    
    IF v_assessment_id IS NULL THEN
        RAISE EXCEPTION 'Nenhuma avaliação HIMSS EMRAM encontrada para a organização';
    END IF;
    
    -- Coletar dados do assessment
    SELECT 
        jsonb_build_object(
            'assessment_id', a.id,
            'assessment_date', a.assessment_date,
            'status', a.status,
            'current_stage', a.current_stage,
            'target_stage', a.target_stage,
            'assessor', a.assessor_id,
            'completion_percentage', a.completion_percentage
        ) INTO v_assessment_data
    FROM iam.himss_assessments a
    WHERE a.id = v_assessment_id;
    
    -- Coletar pontuações por estágio
    WITH stage_stats AS (
        SELECT 
            cr.stage,
            COUNT(*) AS total_criteria,
            COUNT(*) FILTER (WHERE cr.is_compliant = TRUE) AS compliant_criteria,
            AVG(cr.compliance_score) AS avg_score
        FROM 
            iam.himss_criteria_results cr
        WHERE 
            cr.assessment_id = v_assessment_id
        GROUP BY 
            cr.stage
    )
    SELECT
        jsonb_object_agg(
            s.stage::TEXT,
            jsonb_build_object(
                'total_criteria', s.total_criteria,
                'compliant_criteria', s.compliant_criteria,
                'compliance_rate', (s.compliant_criteria::FLOAT / NULLIF(s.total_criteria, 0)),
                'avg_score', s.avg_score
            )
        ) INTO v_stage_scores
    FROM stage_stats s;
    
    -- Coletar pontuações por critério
    SELECT 
        jsonb_object_agg(
            c.criteria_id,
            jsonb_build_object(
                'is_compliant', c.is_compliant,
                'compliance_score', c.compliance_score,
                'evidence_provided', (c.evidence IS NOT NULL AND c.evidence != '[]'::jsonb),
                'stage', c.stage
            )
        ) INTO v_criteria_scores
    FROM iam.himss_criteria_results c
    WHERE c.assessment_id = v_assessment_id;
    
    -- Calcular métricas de compliance
    WITH stage_compliance AS (
        SELECT
            stage,
            COUNT(*) AS total_criteria,
            COUNT(*) FILTER (WHERE is_compliant = TRUE) AS compliant_criteria
        FROM
            iam.himss_criteria_results
        WHERE
            assessment_id = v_assessment_id
        GROUP BY
            stage
    ),
    stage_completion AS (
        SELECT
            s.stage,
            CASE 
                WHEN s.stage <= (SELECT current_stage FROM iam.himss_assessments WHERE id = v_assessment_id) THEN TRUE
                ELSE sc.compliant_criteria = sc.total_criteria
            END AS is_complete
        FROM
            iam.himss_stages s
            LEFT JOIN stage_compliance sc ON s.stage = sc.stage
    )
    SELECT
        jsonb_build_object(
            'current_stage', (SELECT current_stage FROM iam.himss_assessments WHERE id = v_assessment_id),
            'highest_possible_stage', MAX(s.stage) FILTER (WHERE sc.is_complete),
            'total_criteria', COUNT(cr.*),
            'compliant_criteria', COUNT(*) FILTER (WHERE cr.is_compliant = TRUE),
            'overall_compliance_rate', 
                (COUNT(*) FILTER (WHERE cr.is_compliant = TRUE)::FLOAT / 
                NULLIF(COUNT(*), 0))::FLOAT,
            'gaps_to_next_stage', (
                SELECT COUNT(*) 
                FROM iam.himss_criteria_results 
                WHERE assessment_id = v_assessment_id 
                  AND stage = (SELECT current_stage + 1 FROM iam.himss_assessments WHERE id = v_assessment_id)
                  AND is_compliant = FALSE
            ),
            'completed_stages', jsonb_agg(s.stage) FILTER (WHERE sc.is_complete)
        ) INTO v_compliance_metrics
    FROM 
        iam.himss_stages s
        LEFT JOIN stage_completion sc ON s.stage = sc.stage
        LEFT JOIN iam.himss_criteria_results cr ON cr.assessment_id = v_assessment_id AND cr.stage = s.stage;
    
    -- Construir o objeto de features
    v_features := jsonb_build_object(
        'assessment', v_assessment_data,
        'stage_scores', v_stage_scores,
        'criteria_scores', v_criteria_scores,
        'compliance_metrics', v_compliance_metrics,
        'timestamp', NOW()
    );
    
    -- Transformar algumas características numéricas em um vetor
    v_feature_vector := ARRAY[
        (v_assessment_data->>'current_stage')::FLOAT,
        (v_assessment_data->>'target_stage')::FLOAT,
        (v_assessment_data->>'completion_percentage')::FLOAT,
        (v_compliance_metrics->>'overall_compliance_rate')::FLOAT,
        (v_compliance_metrics->>'gaps_to_next_stage')::FLOAT,
        (v_compliance_metrics->>'highest_possible_stage')::FLOAT
    ];
    
    -- Inserir no feature store
    INSERT INTO ml_feature_store (
        feature_set_name,
        entity_type,
        entity_id,
        feature_data,
        feature_vector,
        timestamp,
        metadata
    ) VALUES (
        'himss_emram_compliance',
        'organizations',
        p_organization_id,
        v_features,
        v_feature_vector,
        NOW(),
        jsonb_build_object(
            'assessment_id', v_assessment_id,
            'feature_type', 'compliance',
            'standard', 'HIMSS EMRAM',
            'sector', 'healthcare'
        )
    ) RETURNING id INTO v_feature_id;
    
    RETURN v_feature_id;
END;
$$ LANGUAGE plpgsql;

-- Função para detectar anomalias em compliance de saúde
CREATE OR REPLACE FUNCTION detect_healthcare_compliance_anomalies(
    p_organization_id UUID,
    p_lookback_days INTEGER DEFAULT 90
) RETURNS SETOF ml_anomalies AS $$
DECLARE
    v_threshold FLOAT := 0.75; -- Limiar para detecção de anomalias
    v_iso_features JSONB;
    v_himss_features JSONB;
    v_iso_metrics JSONB;
    v_himss_metrics JSONB;
    v_historical_iso_avg FLOAT;
    v_historical_himss_avg FLOAT;
    v_current_iso_score FLOAT;
    v_current_himss_score FLOAT;
    v_iso_delta FLOAT;
    v_himss_delta FLOAT;
    v_anomaly_id UUID;
BEGIN
    -- Obter as features mais recentes de ISO 27001
    SELECT feature_data INTO v_iso_features
    FROM ml_feature_store
    WHERE feature_set_name = 'iso27001_compliance'
      AND entity_type = 'organizations'
      AND entity_id = p_organization_id
    ORDER BY timestamp DESC
    LIMIT 1;
    
    -- Obter as features mais recentes de HIMSS EMRAM
    SELECT feature_data INTO v_himss_features
    FROM ml_feature_store
    WHERE feature_set_name = 'himss_emram_compliance'
      AND entity_type = 'organizations'
      AND entity_id = p_organization_id
    ORDER BY timestamp DESC
    LIMIT 1;
    
    -- Verificar se existem dados suficientes
    IF v_iso_features IS NULL AND v_himss_features IS NULL THEN
        RAISE NOTICE 'Sem dados de compliance suficientes para detecção de anomalias';
        RETURN;
    END IF;
    
    -- Para ISO 27001
    IF v_iso_features IS NOT NULL THEN
        v_iso_metrics := v_iso_features->'compliance_metrics';
        v_current_iso_score := (v_iso_metrics->>'compliance_score')::FLOAT;
        
        -- Calcular média histórica
        SELECT AVG((feature_data->'compliance_metrics'->>'compliance_score')::FLOAT) INTO v_historical_iso_avg
        FROM ml_feature_store
        WHERE feature_set_name = 'iso27001_compliance'
          AND entity_type = 'organizations'
          AND entity_id = p_organization_id
          AND timestamp >= NOW() - (p_lookback_days || ' days')::INTERVAL;
        
        -- Calcular delta (mudança percentual)
        IF v_historical_iso_avg IS NOT NULL AND v_historical_iso_avg > 0 THEN
            v_iso_delta := ABS(v_current_iso_score - v_historical_iso_avg) / v_historical_iso_avg;
            
            -- Se delta for significativo, registrar anomalia
            IF v_iso_delta > v_threshold THEN
                INSERT INTO ml_anomalies (
                    entity_type,
                    entity_id,
                    anomaly_type,
                    detection_model,
                    anomaly_score,
                    anomaly_threshold,
                    features_contribution,
                    detection_timestamp,
                    status
                ) VALUES (
                    'organizations',
                    p_organization_id,
                    'iso27001_compliance_change',
                    'statistical_deviation',
                    v_iso_delta,
                    v_threshold,
                    jsonb_build_object(
                        'current_score', v_current_iso_score,
                        'historical_avg', v_historical_iso_avg,
                        'delta', v_iso_delta,
                        'change_direction', CASE WHEN v_current_iso_score < v_historical_iso_avg THEN 'decrease' ELSE 'increase' END,
                        'risk_controls', v_iso_features->'risk_metrics'
                    ),
                    NOW(),
                    'open'
                ) RETURNING id INTO v_anomaly_id;
                
                RETURN QUERY SELECT * FROM ml_anomalies WHERE id = v_anomaly_id;
            END IF;
        END IF;
    END IF;
    
    -- Para HIMSS EMRAM
    IF v_himss_features IS NOT NULL THEN
        v_himss_metrics := v_himss_features->'compliance_metrics';
        v_current_himss_score := (v_himss_metrics->>'overall_compliance_rate')::FLOAT;
        
        -- Calcular média histórica
        SELECT AVG((feature_data->'compliance_metrics'->>'overall_compliance_rate')::FLOAT) INTO v_historical_himss_avg
        FROM ml_feature_store
        WHERE feature_set_name = 'himss_emram_compliance'
          AND entity_type = 'organizations'
          AND entity_id = p_organization_id
          AND timestamp >= NOW() - (p_lookback_days || ' days')::INTERVAL;
        
        -- Calcular delta (mudança percentual)
        IF v_historical_himss_avg IS NOT NULL AND v_historical_himss_avg > 0 THEN
            v_himss_delta := ABS(v_current_himss_score - v_historical_himss_avg) / v_historical_himss_avg;
            
            -- Se delta for significativo, registrar anomalia
            IF v_himss_delta > v_threshold THEN
                INSERT INTO ml_anomalies (
                    entity_type,
                    entity_id,
                    anomaly_type,
                    detection_model,
                    anomaly_score,
                    anomaly_threshold,
                    features_contribution,
                    detection_timestamp,
                    status
                ) VALUES (
                    'organizations',
                    p_organization_id,
                    'himss_emram_compliance_change',
                    'statistical_deviation',
                    v_himss_delta,
                    v_threshold,
                    jsonb_build_object(
                        'current_score', v_current_himss_score,
                        'historical_avg', v_historical_himss_avg,
                        'delta', v_himss_delta,
                        'change_direction', CASE WHEN v_current_himss_score < v_historical_himss_avg THEN 'decrease' ELSE 'increase' END,
                        'current_stage', v_himss_features->'assessment'->'current_stage',
                        'gaps_to_next_stage', v_himss_metrics->'gaps_to_next_stage'
                    ),
                    NOW(),
                    'open'
                ) RETURNING id INTO v_anomaly_id;
                
                RETURN QUERY SELECT * FROM ml_anomalies WHERE id = v_anomaly_id;
            END IF;
        END IF;
    END IF;
    
    RETURN;
END;
$$ LANGUAGE plpgsql;
