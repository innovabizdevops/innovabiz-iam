-- INNOVABIZ - IAM Database Monitoring Functions (Part 1)
-- Author: Eduardo Jeremias
-- Date: 08/05/2025
-- Version: 1.0
-- Description: Funções para monitoramento e análise de desempenho do banco de dados (Parte 1).

-- Configurar caminho de busca
SET search_path TO iam_monitoring, iam, public;

-- Função para registrar métricas de consulta
CREATE OR REPLACE FUNCTION register_query_metrics(
    p_query_type VARCHAR,
    p_query_text TEXT,
    p_duration_ms FLOAT,
    p_rows_affected INTEGER,
    p_table_scans INTEGER DEFAULT 0,
    p_index_scans INTEGER DEFAULT 0,
    p_organization_id UUID DEFAULT NULL,
    p_user_id UUID DEFAULT NULL,
    p_database_name VARCHAR DEFAULT current_database(),
    p_execution_plan JSONB DEFAULT NULL,
    p_execution_context JSONB DEFAULT NULL,
    p_tags VARCHAR[] DEFAULT ARRAY[]::VARCHAR[]
) RETURNS UUID AS $$
DECLARE
    v_query_hash TEXT;
    v_normalized_query TEXT;
    v_metrics_id UUID;
    v_slow_query_threshold FLOAT;
BEGIN
    -- Gerar hash e normalizar a query
    v_query_hash := MD5(p_query_text);
    v_normalized_query := regexp_replace(p_query_text, '''[^'']*''', '''?''', 'g');
    v_normalized_query := regexp_replace(v_normalized_query, '\$\d+', '$?', 'g');
    
    -- Inserir métricas da query
    INSERT INTO query_metrics (
        query_type,
        query_text,
        normalized_query_text,
        query_hash,
        duration_ms,
        rows_affected,
        table_scans,
        index_scans,
        organization_id,
        user_id,
        database_name,
        execution_plan,
        execution_context,
        tags
    ) VALUES (
        p_query_type,
        p_query_text,
        v_normalized_query,
        v_query_hash,
        p_duration_ms,
        p_rows_affected,
        p_table_scans,
        p_index_scans,
        p_organization_id,
        p_user_id,
        p_database_name,
        p_execution_plan,
        p_execution_context,
        p_tags
    ) RETURNING id INTO v_metrics_id;
    
    -- Verificar se é uma consulta lenta
    SELECT COALESCE((settings->>'slow_query_threshold_ms')::FLOAT, 1000.0)
    INTO v_slow_query_threshold
    FROM metrics_collection_schedule
    WHERE collector_type = 'query_metrics' AND is_active = TRUE
    LIMIT 1;
    
    IF p_duration_ms > v_slow_query_threshold THEN
        -- Gerar análise básica
        INSERT INTO slow_queries (
            query_metrics_id,
            threshold_ms,
            analysis,
            recommendations
        ) VALUES (
            v_metrics_id,
            v_slow_query_threshold,
            'Query excedeu o limiar de tempo de execução de ' || v_slow_query_threshold || 'ms',
            generate_query_recommendations(p_query_text, p_execution_plan, p_table_scans, p_index_scans, p_rows_affected, p_duration_ms)
        );
        
        -- Verificar se deve gerar alerta
        IF p_duration_ms > (v_slow_query_threshold * 5) THEN
            PERFORM generate_performance_alert(
                'slow_query',
                CASE 
                    WHEN p_duration_ms > (v_slow_query_threshold * 10) THEN 'high'
                    WHEN p_duration_ms > (v_slow_query_threshold * 5) THEN 'medium'
                    ELSE 'low'
                END,
                'query',
                p_query_type,
                'Consulta extremamente lenta detectada: ' || p_duration_ms || 'ms',
                jsonb_build_object(
                    'query_metrics_id', v_metrics_id,
                    'query_type', p_query_type,
                    'duration_ms', p_duration_ms,
                    'threshold_ms', v_slow_query_threshold,
                    'rows_affected', p_rows_affected
                ),
                p_duration_ms,
                v_slow_query_threshold,
                p_organization_id
            );
        END IF;
    END IF;
    
    RETURN v_metrics_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION register_query_metrics IS 'Registra métricas de uma consulta SQL e identifica consultas lentas';

-- Função de geração de recomendações para consultas
CREATE OR REPLACE FUNCTION generate_query_recommendations(
    p_query_text TEXT,
    p_execution_plan JSONB DEFAULT NULL,
    p_table_scans INTEGER DEFAULT 0,
    p_index_scans INTEGER DEFAULT 0,
    p_rows_affected INTEGER DEFAULT 0,
    p_duration_ms FLOAT DEFAULT 0
) RETURNS JSONB AS $$
DECLARE
    v_recommendations JSONB := '[]'::JSONB;
BEGIN
    -- Análise básica baseada nas características da consulta
    IF p_table_scans > 0 AND p_index_scans = 0 THEN
        v_recommendations := v_recommendations || jsonb_build_object(
            'type', 'index_suggestion',
            'description', 'A consulta está realizando varredura sequencial de tabela sem usar índices.',
            'recommendation', 'Considere adicionar índices apropriados para melhorar o desempenho.'
        );
    END IF;
    
    -- Verificar se é uma consulta de seleção com muitas linhas
    IF p_query_text ~* '^SELECT' AND p_rows_affected > 10000 THEN
        v_recommendations := v_recommendations || jsonb_build_object(
            'type', 'optimization',
            'description', 'A consulta está retornando um grande número de linhas (' || p_rows_affected || ').',
            'recommendation', 'Considere adicionar limites, filtros adicionais ou implementar paginação.'
        );
    END IF;
    
    -- Verificar por subconsultas ou junções múltiplas
    IF p_query_text ~* 'FROM.*SELECT' THEN
        v_recommendations := v_recommendations || jsonb_build_object(
            'type', 'query_structure',
            'description', 'A consulta contém subconsultas que podem afetar o desempenho.',
            'recommendation', 'Considere reescrever usando CTEs (WITH clause) ou junções para melhor desempenho.'
        );
    END IF;
    
    -- Verificar por muitas junções
    IF (SELECT (LENGTH(p_query_text) - LENGTH(REPLACE(UPPER(p_query_text), 'JOIN', ''))) / 4) > 5 THEN
        v_recommendations := v_recommendations || jsonb_build_object(
            'type', 'query_structure',
            'description', 'A consulta contém muitas junções.',
            'recommendation', 'Considere simplificar a consulta ou dividi-la em várias etapas.'
        );
    END IF;
    
    -- Análise de plano de execução, se disponível
    IF p_execution_plan IS NOT NULL THEN
        -- Detectar varredura sequencial em tabelas grandes
        IF p_execution_plan ? 'Plan' AND p_execution_plan->'Plan'->>'Node Type' = 'Seq Scan' THEN
            v_recommendations := v_recommendations || jsonb_build_object(
                'type', 'execution_plan',
                'description', 'O plano de execução mostra uma varredura sequencial (Seq Scan).',
                'recommendation', 'Adicione índices apropriados para as colunas usadas nas cláusulas WHERE, JOIN ou ORDER BY.'
            );
        END IF;
        
        -- Verificar por hash joins
        IF p_execution_plan ? 'Plan' AND p_execution_plan::TEXT ~* 'Hash Join' THEN
            v_recommendations := v_recommendations || jsonb_build_object(
                'type', 'execution_plan',
                'description', 'O plano mostra Hash Joins, que podem consumir muita memória para conjuntos de dados grandes.',
                'recommendation', 'Verifique se as estatísticas da tabela estão atualizadas. Considere ajustar work_mem se necessário.'
            );
        END IF;
    END IF;
    
    -- Se não tiver recomendações específicas, fornecer uma recomendação genérica
    IF jsonb_array_length(v_recommendations) = 0 THEN
        v_recommendations := jsonb_build_array(jsonb_build_object(
            'type', 'general',
            'description', 'Consulta lenta detectada, mas não foi possível identificar causas específicas.',
            'recommendation', 'Analise as cláusulas WHERE, JOIN e ORDER BY. Verifique se as estatísticas da tabela estão atualizadas.'
        ));
    END IF;
    
    RETURN v_recommendations;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION generate_query_recommendations IS 'Gera recomendações para otimização de consultas baseadas em características e plano de execução';

-- Função para gerar alertas de desempenho
CREATE OR REPLACE FUNCTION generate_performance_alert(
    p_alert_type VARCHAR,
    p_alert_level VARCHAR,
    p_object_type VARCHAR,
    p_object_name VARCHAR,
    p_message TEXT,
    p_details JSONB DEFAULT NULL,
    p_metric_value FLOAT DEFAULT NULL,
    p_threshold_value FLOAT DEFAULT NULL,
    p_organization_id UUID DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    v_alert_id UUID;
BEGIN
    -- Verificar se já existe um alerta similar não resolvido
    SELECT id INTO v_alert_id
    FROM performance_alerts
    WHERE alert_type = p_alert_type
      AND object_type = p_object_type
      AND object_name = p_object_name
      AND resolution_status IN ('pending', 'in_progress')
      AND timestamp > NOW() - INTERVAL '24 hours'
    LIMIT 1;
    
    -- Se existir, atualizar o alerta
    IF v_alert_id IS NOT NULL THEN
        UPDATE performance_alerts
        SET alert_level = GREATEST(alert_level, p_alert_level),
            timestamp = NOW(),
            details = p_details,
            metric_value = p_metric_value,
            threshold_value = p_threshold_value,
            message = p_message
        WHERE id = v_alert_id;
        
        RETURN v_alert_id;
    END IF;
    
    -- Senão, criar um novo alerta
    INSERT INTO performance_alerts (
        alert_type,
        alert_level,
        object_type,
        object_name,
        message,
        details,
        metric_value,
        threshold_value,
        organization_id
    ) VALUES (
        p_alert_type,
        p_alert_level,
        p_object_type,
        p_object_name,
        p_message,
        p_details,
        p_metric_value,
        p_threshold_value,
        p_organization_id
    ) RETURNING id INTO v_alert_id;
    
    RETURN v_alert_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION generate_performance_alert IS 'Gera ou atualiza um alerta de desempenho';

-- Função para reconhecer um alerta
CREATE OR REPLACE FUNCTION acknowledge_performance_alert(
    p_alert_id UUID,
    p_user_id UUID,
    p_notes TEXT DEFAULT NULL
) RETURNS BOOLEAN AS $$
BEGIN
    UPDATE performance_alerts
    SET is_acknowledged = TRUE,
        acknowledged_by = p_user_id,
        acknowledged_at = NOW(),
        details = CASE 
            WHEN details IS NULL THEN jsonb_build_object('acknowledgment_notes', p_notes)
            ELSE details || jsonb_build_object('acknowledgment_notes', p_notes)
        END
    WHERE id = p_alert_id;
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION acknowledge_performance_alert IS 'Marca um alerta de desempenho como reconhecido';
