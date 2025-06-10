-- INNOVABIZ - IAM Database Monitoring Functions (Part 2)
-- Author: Eduardo Jeremias
-- Date: 08/05/2025
-- Version: 1.0
-- Description: Funções para monitoramento e análise de desempenho do banco de dados (Parte 2).

-- Configurar caminho de busca
SET search_path TO iam_monitoring, iam, public;

-- Função para resolver um alerta de desempenho
CREATE OR REPLACE FUNCTION resolve_performance_alert(
    p_alert_id UUID,
    p_user_id UUID,
    p_resolution_status VARCHAR,
    p_resolution_notes TEXT
) RETURNS BOOLEAN AS $$
BEGIN
    IF p_resolution_status NOT IN ('resolved', 'false_positive', 'deferred') THEN
        RAISE EXCEPTION 'Status de resolução inválido. Valores permitidos: resolved, false_positive, deferred';
    END IF;
    
    UPDATE performance_alerts
    SET resolution_status = p_resolution_status,
        resolved_by = p_user_id,
        resolved_at = NOW(),
        details = CASE 
            WHEN details IS NULL THEN jsonb_build_object('resolution_notes', p_resolution_notes)
            ELSE details || jsonb_build_object('resolution_notes', p_resolution_notes)
        END
    WHERE id = p_alert_id;
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION resolve_performance_alert IS 'Marca um alerta de desempenho como resolvido, falso positivo ou adiado';

-- Função para agendar a coleta periódica de métricas
CREATE OR REPLACE FUNCTION schedule_metrics_collection(
    p_collector_type VARCHAR,
    p_frequency_minutes INTEGER,
    p_settings JSONB DEFAULT NULL,
    p_retention_days INTEGER DEFAULT NULL,
    p_created_by UUID DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    v_schedule_id UUID;
BEGIN
    -- Verificar se já existe um agendamento para este tipo de coletor
    SELECT id INTO v_schedule_id
    FROM metrics_collection_schedule
    WHERE collector_type = p_collector_type;
    
    -- Se existir, atualizar
    IF v_schedule_id IS NOT NULL THEN
        UPDATE metrics_collection_schedule
        SET frequency_minutes = p_frequency_minutes,
            is_active = TRUE,
            next_run = NOW() + (p_frequency_minutes || ' minutes')::INTERVAL,
            settings = COALESCE(p_settings, settings),
            retention_days = COALESCE(p_retention_days, retention_days),
            updated_at = NOW()
        WHERE id = v_schedule_id;
        
        RETURN v_schedule_id;
    END IF;
    
    -- Senão, criar novo
    INSERT INTO metrics_collection_schedule (
        collector_type,
        frequency_minutes,
        is_active,
        next_run,
        settings,
        created_by,
        retention_days
    ) VALUES (
        p_collector_type,
        p_frequency_minutes,
        TRUE,
        NOW() + (p_frequency_minutes || ' minutes')::INTERVAL,
        COALESCE(p_settings, '{}'::JSONB),
        p_created_by,
        COALESCE(p_retention_days, 90)
    ) RETURNING id INTO v_schedule_id;
    
    RETURN v_schedule_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION schedule_metrics_collection IS 'Agenda ou atualiza a coleta periódica de métricas';

-- Função para registrar estatísticas gerais do banco de dados
CREATE OR REPLACE FUNCTION collect_database_statistics() RETURNS UUID AS $$
DECLARE
    v_stats_id UUID;
    v_start_time TIMESTAMP;
    v_active_connections INTEGER;
    v_idle_connections INTEGER;
    v_cache_hit_ratio FLOAT;
    v_deadlocks INTEGER;
    v_conflicts INTEGER;
    v_temp_files_size BIGINT;
    v_database_size BIGINT;
    v_transaction_count BIGINT;
    v_checkpoint_stats JSONB;
    v_locks_stats JSONB;
    v_vacuum_stats JSONB;
    v_bgwriter_stats JSONB;
BEGIN
    v_start_time := clock_timestamp();
    
    -- Coletar número de conexões
    SELECT 
        COUNT(*) FILTER (WHERE state = 'active') AS active,
        COUNT(*) FILTER (WHERE state = 'idle') AS idle
    INTO v_active_connections, v_idle_connections
    FROM pg_stat_activity;
    
    -- Coletar taxa de acerto do cache
    SELECT 
        sum(heap_blks_hit) / NULLIF(sum(heap_blks_hit) + sum(heap_blks_read), 0) AS ratio
    INTO v_cache_hit_ratio
    FROM pg_statio_user_tables;
    
    -- Coletar informações de deadlocks e conflitos
    SELECT 
        deadlocks,
        conflicts
    INTO v_deadlocks, v_conflicts
    FROM pg_stat_database
    WHERE datname = current_database();
    
    -- Coletar tamanho dos arquivos temporários
    SELECT 
        COALESCE(temp_bytes, 0) AS temp_size
    INTO v_temp_files_size
    FROM pg_stat_database
    WHERE datname = current_database();
    
    -- Coletar tamanho do banco de dados
    SELECT 
        pg_database_size(current_database()) INTO v_database_size;
    
    -- Coletar contagem de transações
    SELECT 
        xact_commit + xact_rollback
    INTO v_transaction_count
    FROM pg_stat_database
    WHERE datname = current_database();
    
    -- Coletar estatísticas de checkpoint
    SELECT jsonb_build_object(
        'checkpoints_timed', checkpoints_timed,
        'checkpoints_req', checkpoints_req,
        'checkpoint_write_time', checkpoint_write_time,
        'checkpoint_sync_time', checkpoint_sync_time,
        'buffers_checkpoint', buffers_checkpoint,
        'buffers_clean', buffers_clean,
        'buffers_backend', buffers_backend
    ) INTO v_checkpoint_stats
    FROM pg_stat_bgwriter;
    
    -- Coletar estatísticas de locks
    WITH lock_data AS (
        SELECT 
            mode,
            COUNT(*) AS count
        FROM pg_locks
        GROUP BY mode
    )
    SELECT 
        jsonb_object_agg(mode, count) INTO v_locks_stats
    FROM lock_data;
    
    -- Coletar estatísticas de vacuum
    WITH vacuum_data AS (
        SELECT 
            schemaname || '.' || relname AS table_name,
            n_live_tup,
            n_dead_tup,
            last_vacuum,
            last_autovacuum,
            vacuum_count,
            autovacuum_count
        FROM pg_stat_user_tables
        ORDER BY n_dead_tup DESC
        LIMIT 10
    )
    SELECT 
        jsonb_agg(to_jsonb(vacuum_data)) INTO v_vacuum_stats
    FROM vacuum_data;
    
    -- Coletar estatísticas do bgwriter
    SELECT jsonb_build_object(
        'buffers_backend_fsync', buffers_backend_fsync,
        'buffers_alloc', buffers_alloc,
        'maxwritten_clean', maxwritten_clean
    ) INTO v_bgwriter_stats
    FROM pg_stat_bgwriter;
    
    -- Inserir as estatísticas coletadas
    INSERT INTO database_stats (
        active_connections,
        idle_connections,
        cache_hit_ratio,
        deadlocks,
        conflicts,
        temp_files_size,
        database_size,
        transaction_count,
        checkpoint_stats,
        locks_stats,
        vacuum_stats,
        bgwriter_stats,
        snapshot_duration_ms
    ) VALUES (
        v_active_connections,
        v_idle_connections,
        v_cache_hit_ratio,
        v_deadlocks,
        v_conflicts,
        v_temp_files_size,
        v_database_size,
        v_transaction_count,
        v_checkpoint_stats,
        v_locks_stats,
        v_vacuum_stats,
        v_bgwriter_stats,
        EXTRACT(EPOCH FROM (clock_timestamp() - v_start_time)) * 1000
    ) RETURNING id INTO v_stats_id;
    
    -- Remover registros antigos com base na configuração de retenção
    DELETE FROM database_stats
    WHERE timestamp < NOW() - (
        SELECT (retention_days || ' days')::INTERVAL
        FROM metrics_collection_schedule
        WHERE collector_type = 'database_stats'
        LIMIT 1
    );
    
    RETURN v_stats_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION collect_database_statistics IS 'Coleta e registra estatísticas gerais do banco de dados';

-- Função para limpar dados antigos com base nas configurações de retenção
CREATE OR REPLACE FUNCTION cleanup_monitoring_data() RETURNS TABLE(
    collection_type VARCHAR,
    deleted_count BIGINT
) AS $$
DECLARE
    v_retention INTERVAL;
    v_deleted BIGINT;
    v_collection_type VARCHAR;
BEGIN
    -- Iterar sobre todos os agendamentos ativos
    FOR v_collection_type, v_retention IN
        SELECT collector_type, (retention_days || ' days')::INTERVAL
        FROM metrics_collection_schedule
        WHERE is_active = TRUE
    LOOP
        CASE v_collection_type
            WHEN 'query_metrics' THEN
                DELETE FROM slow_queries sq
                USING query_metrics qm
                WHERE sq.query_metrics_id = qm.id
                AND qm.timestamp < NOW() - v_retention;
                
                DELETE FROM query_metrics
                WHERE timestamp < NOW() - v_retention
                RETURNING 1 INTO v_deleted;
                
                RETURN QUERY SELECT 'query_metrics'::VARCHAR, COALESCE(v_deleted, 0);
                
            WHEN 'database_stats' THEN
                DELETE FROM database_stats
                WHERE timestamp < NOW() - v_retention
                RETURNING 1 INTO v_deleted;
                
                RETURN QUERY SELECT 'database_stats'::VARCHAR, COALESCE(v_deleted, 0);
                
            WHEN 'index_stats' THEN
                DELETE FROM index_stats
                WHERE timestamp < NOW() - v_retention
                RETURNING 1 INTO v_deleted;
                
                RETURN QUERY SELECT 'index_stats'::VARCHAR, COALESCE(v_deleted, 0);
                
            WHEN 'table_stats' THEN
                DELETE FROM table_stats
                WHERE timestamp < NOW() - v_retention
                RETURNING 1 INTO v_deleted;
                
                RETURN QUERY SELECT 'table_stats'::VARCHAR, COALESCE(v_deleted, 0);
                
            WHEN 'performance_report' THEN
                DELETE FROM performance_reports
                WHERE timestamp < NOW() - v_retention
                RETURNING 1 INTO v_deleted;
                
                RETURN QUERY SELECT 'performance_reports'::VARCHAR, COALESCE(v_deleted, 0);
                
            ELSE
                CONTINUE;
        END CASE;
    END LOOP;
    
    -- Limpar alertas resolvidos antigos
    DELETE FROM performance_alerts
    WHERE resolution_status IN ('resolved', 'false_positive')
    AND resolved_at < NOW() - INTERVAL '90 days'
    RETURNING 1 INTO v_deleted;
    
    RETURN QUERY SELECT 'performance_alerts'::VARCHAR, COALESCE(v_deleted, 0);
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_monitoring_data IS 'Limpa dados antigos do monitoramento com base nas configurações de retenção';

-- Função para gerar relatório de desempenho
CREATE OR REPLACE FUNCTION generate_performance_report(
    p_report_type VARCHAR,
    p_title VARCHAR,
    p_description TEXT DEFAULT NULL,
    p_start_period TIMESTAMP WITH TIME ZONE DEFAULT NOW() - INTERVAL '1 day',
    p_end_period TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    p_organization_id UUID DEFAULT NULL,
    p_generated_by VARCHAR DEFAULT NULL,
    p_report_format VARCHAR DEFAULT 'JSON'
) RETURNS UUID AS $$
DECLARE
    v_report_id UUID;
    v_metrics_summary JSONB := '{}'::JSONB;
    v_issues_found INTEGER := 0;
    v_recommendations JSONB := '[]'::JSONB;
    v_visualizations JSONB := '{}'::JSONB;
    
    -- Variáveis para estatísticas temporárias
    v_query_count BIGINT;
    v_avg_duration FLOAT;
    v_slow_queries BIGINT;
    v_active_connections FLOAT;
    v_cache_hit_ratio FLOAT;
BEGIN
    -- Calcular métricas básicas para o período especificado
    -- 1. Estatísticas de consultas
    SELECT 
        COUNT(*) AS query_count,
        AVG(duration_ms) AS avg_duration,
        COUNT(*) FILTER (WHERE duration_ms > 1000) AS slow_queries
    INTO v_query_count, v_avg_duration, v_slow_queries
    FROM query_metrics
    WHERE timestamp BETWEEN p_start_period AND p_end_period
    AND (p_organization_id IS NULL OR organization_id = p_organization_id);
    
    -- 2. Estatísticas gerais do banco de dados
    SELECT 
        AVG(active_connections) AS active_connections,
        AVG(cache_hit_ratio) AS cache_hit_ratio
    INTO v_active_connections, v_cache_hit_ratio
    FROM database_stats
    WHERE timestamp BETWEEN p_start_period AND p_end_period;
    
    -- Construir o sumário de métricas
    v_metrics_summary := jsonb_build_object(
        'period_start', p_start_period,
        'period_end', p_end_period,
        'query_count', v_query_count,
        'avg_query_duration_ms', v_avg_duration,
        'slow_queries_count', v_slow_queries,
        'avg_active_connections', v_active_connections,
        'avg_cache_hit_ratio', v_cache_hit_ratio
    );
    
    -- Coletar problemas e recomendações com base no tipo de relatório
    CASE p_report_type
        WHEN 'slow_queries' THEN
            -- Coletar consultas lentas
            WITH slow_query_data AS (
                SELECT 
                    qm.query_type,
                    qm.normalized_query_text,
                    qm.duration_ms,
                    qm.table_scans,
                    qm.index_scans,
                    ROW_NUMBER() OVER (PARTITION BY qm.query_hash ORDER BY qm.duration_ms DESC) AS rn
                FROM query_metrics qm
                WHERE qm.timestamp BETWEEN p_start_period AND p_end_period
                AND qm.duration_ms > 1000
                AND (p_organization_id IS NULL OR qm.organization_id = p_organization_id)
            )
            SELECT 
                COUNT(*) INTO v_issues_found
            FROM slow_query_data
            WHERE rn = 1;
            
            -- Gerar recomendações para as consultas lentas mais significativas
            WITH slow_queries_with_recommendations AS (
                SELECT 
                    sq.id,
                    qm.query_type,
                    qm.normalized_query_text,
                    qm.duration_ms,
                    qm.table_scans,
                    qm.index_scans,
                    sq.recommendations
                FROM slow_queries sq
                JOIN query_metrics qm ON sq.query_metrics_id = qm.id
                WHERE qm.timestamp BETWEEN p_start_period AND p_end_period
                AND (p_organization_id IS NULL OR qm.organization_id = p_organization_id)
                ORDER BY qm.duration_ms DESC
                LIMIT 10
            )
            SELECT 
                jsonb_agg(
                    jsonb_build_object(
                        'query_type', query_type,
                        'query_text', normalized_query_text,
                        'duration_ms', duration_ms,
                        'recommendations', recommendations
                    )
                ) INTO v_recommendations
            FROM slow_queries_with_recommendations;
            
            -- Visualizações para consultas lentas
            WITH query_type_distribution AS (
                SELECT 
                    query_type,
                    COUNT(*) AS count,
                    AVG(duration_ms) AS avg_duration
                FROM query_metrics
                WHERE timestamp BETWEEN p_start_period AND p_end_period
                AND duration_ms > 1000
                AND (p_organization_id IS NULL OR organization_id = p_organization_id)
                GROUP BY query_type
                ORDER BY count DESC
            )
            SELECT 
                jsonb_build_object(
                    'slow_queries_by_type', (
                        SELECT jsonb_agg(to_jsonb(query_type_distribution))
                        FROM query_type_distribution
                    ),
                    'slow_queries_timeline', (
                        SELECT jsonb_agg(
                            jsonb_build_object(
                                'time_bucket', time_bucket,
                                'count', count
                            )
                        )
                        FROM (
                            SELECT 
                                DATE_TRUNC('hour', timestamp) AS time_bucket,
                                COUNT(*) AS count
                            FROM query_metrics
                            WHERE timestamp BETWEEN p_start_period AND p_end_period
                            AND duration_ms > 1000
                            AND (p_organization_id IS NULL OR organization_id = p_organization_id)
                            GROUP BY DATE_TRUNC('hour', timestamp)
                            ORDER BY time_bucket
                        ) AS timeline_data
                    )
                ) INTO v_visualizations;
                
        WHEN 'index_recommendations' THEN
            -- Coletar recomendações de índices
            WITH problematic_indexes AS (
                SELECT *
                FROM vw_problematic_indexes
                WHERE timestamp BETWEEN p_start_period AND p_end_period
            )
            SELECT 
                COUNT(*) INTO v_issues_found
            FROM problematic_indexes;
            
            -- Gerar recomendações
            WITH index_recommendations AS (
                SELECT 
                    schema_name,
                    table_name,
                    index_name,
                    index_size_pretty,
                    usage_ratio,
                    bloat_ratio,
                    issue_type,
                    CASE 
                        WHEN issue_type = 'Unused Index' THEN 
                            'Consider dropping unused index: ' || schema_name || '.' || index_name
                        WHEN issue_type = 'Bloated Index' THEN 
                            'Consider rebuilding bloated index: REINDEX INDEX ' || schema_name || '.' || index_name
                        WHEN issue_type IN ('Needs Vacuum', 'Needs Analyze') THEN 
                            'Run maintenance: ' || 
                            CASE 
                                WHEN issue_type = 'Needs Vacuum' THEN 'VACUUM ANALYZE ' 
                                ELSE 'ANALYZE ' 
                            END || 
                            schema_name || '.' || table_name
                        ELSE 'Review index usage and structure'
                    END AS recommendation
                FROM vw_problematic_indexes
                ORDER BY 
                    CASE issue_type
                        WHEN 'Bloated Index' THEN 1
                        WHEN 'Unused Index' THEN 2
                        ELSE 3
                    END,
                    index_size DESC
                LIMIT 20
            )
            SELECT 
                jsonb_agg(
                    jsonb_build_object(
                        'schema_table', schema_name || '.' || table_name,
                        'index_name', index_name,
                        'issue_type', issue_type,
                        'index_size', index_size_pretty,
                        'recommendation', recommendation
                    )
                ) INTO v_recommendations
            FROM index_recommendations;
            
            -- Visualizações para recomendações de índices
            SELECT 
                jsonb_build_object(
                    'index_issues_by_type', (
                        SELECT jsonb_agg(
                            jsonb_build_object(
                                'issue_type', issue_type,
                                'count', count
                            )
                        )
                        FROM (
                            SELECT 
                                issue_type,
                                COUNT(*) AS count
                            FROM vw_problematic_indexes
                            GROUP BY issue_type
                            ORDER BY count DESC
                        ) AS issues_by_type
                    ),
                    'index_size_distribution', (
                        SELECT jsonb_agg(
                            jsonb_build_object(
                                'size_range', size_range,
                                'count', count
                            )
                        )
                        FROM (
                            SELECT 
                                CASE 
                                    WHEN index_size < 1048576 THEN '<1MB'
                                    WHEN index_size < 10485760 THEN '1-10MB'
                                    WHEN index_size < 104857600 THEN '10-100MB'
                                    WHEN index_size < 1073741824 THEN '100MB-1GB'
                                    ELSE '>1GB'
                                END AS size_range,
                                COUNT(*) AS count
                            FROM index_stats
                            WHERE timestamp = (SELECT MAX(timestamp) FROM index_stats)
                            GROUP BY size_range
                            ORDER BY 
                                CASE size_range
                                    WHEN '<1MB' THEN 1
                                    WHEN '1-10MB' THEN 2
                                    WHEN '10-100MB' THEN 3
                                    WHEN '100MB-1GB' THEN 4
                                    ELSE 5
                                END
                        ) AS size_distribution
                    )
                ) INTO v_visualizations;
            
        ELSE -- Relatório geral
            -- Problemas gerais
            SELECT 
                (SELECT COUNT(*) FROM vw_problematic_indexes) +
                (SELECT COUNT(*) FROM vw_problematic_tables) +
                (SELECT COUNT(*) FROM slow_queries sq
                 JOIN query_metrics qm ON sq.query_metrics_id = qm.id
                 WHERE qm.timestamp BETWEEN p_start_period AND p_end_period) INTO v_issues_found;
                
            -- Recomendações gerais
            WITH general_recommendations AS (
                -- Top 3 consultas lentas
                (SELECT 
                    'slow_query' AS issue_type,
                    'Query Type: ' || qm.query_type || ' - Duration: ' || ROUND(qm.duration_ms::NUMERIC, 2) || 'ms' AS issue_description,
                    COALESCE(
                        (SELECT recommendation FROM jsonb_array_elements(sq.recommendations) AS r(recommendation) LIMIT 1),
                        'Optimize query performance'
                    )::TEXT AS recommendation,
                    qm.duration_ms AS priority_value
                FROM slow_queries sq
                JOIN query_metrics qm ON sq.query_metrics_id = qm.id
                WHERE qm.timestamp BETWEEN p_start_period AND p_end_period
                ORDER BY qm.duration_ms DESC
                LIMIT 3)
                
                UNION ALL
                
                -- Top 3 problemas de índice
                (SELECT 
                    'index_issue' AS issue_type,
                    index_name || ' on ' || schema_name || '.' || table_name || ' - ' || issue_type AS issue_description,
                    CASE 
                        WHEN issue_type = 'Unused Index' THEN 
                            'Consider dropping unused index: ' || schema_name || '.' || index_name
                        WHEN issue_type = 'Bloated Index' THEN 
                            'Consider rebuilding bloated index: REINDEX INDEX ' || schema_name || '.' || index_name
                        ELSE 'Review index usage and structure'
                    END AS recommendation,
                    index_size AS priority_value
                FROM vw_problematic_indexes
                ORDER BY 
                    CASE issue_type
                        WHEN 'Bloated Index' THEN 1
                        WHEN 'Unused Index' THEN 2
                        ELSE 3
                    END,
                    index_size DESC
                LIMIT 3)
                
                UNION ALL
                
                -- Top 3 problemas de tabela
                (SELECT 
                    'table_issue' AS issue_type,
                    table_name || ' - ' || issue_type AS issue_description,
                    CASE 
                        WHEN issue_type = 'High Dead Tuple Ratio' THEN 
                            'Run VACUUM ANALYZE on ' || schema_name || '.' || table_name
                        WHEN issue_type = 'Bloated Table' THEN 
                            'Consider rebuilding table or running VACUUM FULL (with caution)'
                        ELSE 'Review table structure and maintenance'
                    END AS recommendation,
                    table_size AS priority_value
                FROM vw_problematic_tables
                ORDER BY 
                    CASE issue_type
                        WHEN 'High Dead Tuple Ratio' THEN 1
                        WHEN 'Bloated Table' THEN 2
                        ELSE 3
                    END,
                    table_size DESC
                LIMIT 3)
            )
            SELECT 
                jsonb_agg(
                    jsonb_build_object(
                        'issue_type', issue_type,
                        'description', issue_description,
                        'recommendation', recommendation
                    )
                ) INTO v_recommendations
            FROM general_recommendations
            ORDER BY priority_value DESC;
            
            -- Visualizações gerais
            SELECT 
                jsonb_build_object(
                    'database_metrics_timeline', (
                        SELECT jsonb_agg(
                            jsonb_build_object(
                                'time_bucket', time_bucket,
                                'active_connections', avg_active_connections,
                                'cache_hit_ratio', avg_cache_hit_ratio
                            )
                        )
                        FROM (
                            SELECT 
                                DATE_TRUNC('hour', timestamp) AS time_bucket,
                                AVG(active_connections) AS avg_active_connections,
                                AVG(cache_hit_ratio) AS avg_cache_hit_ratio
                            FROM database_stats
                            WHERE timestamp BETWEEN p_start_period AND p_end_period
                            GROUP BY DATE_TRUNC('hour', timestamp)
                            ORDER BY time_bucket
                        ) AS db_metrics
                    ),
                    'query_performance_timeline', (
                        SELECT jsonb_agg(
                            jsonb_build_object(
                                'time_bucket', time_bucket,
                                'avg_duration', avg_duration,
                                'query_count', query_count
                            )
                        )
                        FROM (
                            SELECT 
                                DATE_TRUNC('hour', timestamp) AS time_bucket,
                                AVG(duration_ms) AS avg_duration,
                                COUNT(*) AS query_count
                            FROM query_metrics
                            WHERE timestamp BETWEEN p_start_period AND p_end_period
                            AND (p_organization_id IS NULL OR organization_id = p_organization_id)
                            GROUP BY DATE_TRUNC('hour', timestamp)
                            ORDER BY time_bucket
                        ) AS query_metrics
                    ),
                    'issues_by_type', (
                        SELECT jsonb_build_object(
                            'slow_queries', (SELECT COUNT(*) FROM slow_queries sq
                                            JOIN query_metrics qm ON sq.query_metrics_id = qm.id
                                            WHERE qm.timestamp BETWEEN p_start_period AND p_end_period),
                            'index_issues', (SELECT COUNT(*) FROM vw_problematic_indexes),
                            'table_issues', (SELECT COUNT(*) FROM vw_problematic_tables)
                        )
                    )
                ) INTO v_visualizations;
    END CASE;
    
    -- Criar o relatório
    INSERT INTO performance_reports (
        report_type,
        title,
        description,
        start_period,
        end_period,
        metrics_summary,
        issues_found,
        recommendations,
        visualizations,
        generated_by,
        report_format,
        organization_id
    ) VALUES (
        p_report_type,
        p_title,
        p_description,
        p_start_period,
        p_end_period,
        v_metrics_summary,
        v_issues_found,
        v_recommendations,
        v_visualizations,
        p_generated_by,
        p_report_format,
        p_organization_id
    ) RETURNING id INTO v_report_id;
    
    RETURN v_report_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION generate_performance_report IS 'Gera um relatório detalhado de desempenho do banco de dados';
