-- INNOVABIZ - IAM Database Monitoring Views
-- Author: Eduardo Jeremias
-- Date: 08/05/2025
-- Version: 1.0
-- Description: Views para monitoramento e análise de desempenho do banco de dados.

-- Configurar caminho de busca
SET search_path TO iam_monitoring, iam, public;

-- View para consultas lentas recentes
CREATE OR REPLACE VIEW vw_recent_slow_queries AS
SELECT 
    sq.id,
    sq.query_metrics_id,
    qm.query_type,
    qm.query_text,
    qm.normalized_query_text,
    qm.duration_ms,
    qm.rows_affected,
    qm.table_scans,
    qm.index_scans,
    qm.timestamp,
    qm.organization_id,
    qm.database_name,
    sq.threshold_ms,
    sq.analysis,
    sq.recommendations,
    sq.is_resolved,
    sq.resolution_notes,
    sq.resolved_by,
    sq.resolved_at,
    u.username AS resolved_by_username,
    o.name AS organization_name,
    (qm.duration_ms / sq.threshold_ms) AS threshold_ratio
FROM 
    slow_queries sq
JOIN 
    query_metrics qm ON sq.query_metrics_id = qm.id
LEFT JOIN 
    iam.users u ON sq.resolved_by = u.id
LEFT JOIN 
    iam.organizations o ON qm.organization_id = o.id
WHERE 
    qm.timestamp > NOW() - INTERVAL '7 days'
ORDER BY 
    qm.timestamp DESC;

COMMENT ON VIEW vw_recent_slow_queries IS 'Consultas lentas detectadas nos últimos 7 dias';

-- View para tendências de consultas por tipo
CREATE OR REPLACE VIEW vw_query_type_trends AS
SELECT 
    query_type,
    DATE_TRUNC('hour', timestamp) AS time_bucket,
    COUNT(*) AS query_count,
    AVG(duration_ms) AS avg_duration_ms,
    MAX(duration_ms) AS max_duration_ms,
    MIN(duration_ms) AS min_duration_ms,
    SUM(rows_affected) AS total_rows_affected,
    AVG(rows_affected) AS avg_rows_affected,
    SUM(table_scans) AS total_table_scans,
    SUM(index_scans) AS total_index_scans
FROM 
    query_metrics
WHERE 
    timestamp > NOW() - INTERVAL '24 hours'
GROUP BY 
    query_type, DATE_TRUNC('hour', timestamp)
ORDER BY 
    time_bucket DESC, avg_duration_ms DESC;

COMMENT ON VIEW vw_query_type_trends IS 'Tendências de desempenho de consultas agrupadas por tipo em intervalos de hora';

-- View para estatísticas de índices problemáticos
CREATE OR REPLACE VIEW vw_problematic_indexes AS
SELECT 
    i.id,
    i.timestamp,
    i.schema_name,
    i.table_name,
    i.index_name,
    i.index_size,
    pg_size_pretty(i.index_size) AS index_size_pretty,
    i.index_scans,
    i.tuple_reads,
    i.tuple_fetches,
    i.usage_ratio,
    i.bloat_ratio,
    i.duplicity_ratio,
    i.last_vacuum,
    i.last_analyze,
    NOW() - i.last_vacuum AS vacuum_age,
    NOW() - i.last_analyze AS analyze_age,
    CASE
        WHEN i.usage_ratio < 0.01 AND i.index_scans < 100 THEN 'Unused Index'
        WHEN i.bloat_ratio > 0.5 THEN 'Bloated Index'
        WHEN i.duplicity_ratio > 0.9 THEN 'Duplicated Index'
        WHEN NOW() - i.last_vacuum > INTERVAL '30 days' THEN 'Needs Vacuum'
        WHEN NOW() - i.last_analyze > INTERVAL '30 days' THEN 'Needs Analyze'
        ELSE 'OK'
    END AS issue_type
FROM 
    index_stats i
WHERE 
    (i.usage_ratio < 0.01 AND i.index_scans < 100) OR
    i.bloat_ratio > 0.5 OR
    i.duplicity_ratio > 0.9 OR
    NOW() - i.last_vacuum > INTERVAL '30 days' OR
    NOW() - i.last_analyze > INTERVAL '30 days'
ORDER BY 
    CASE
        WHEN i.bloat_ratio > 0.5 THEN 1
        WHEN i.usage_ratio < 0.01 AND i.index_scans < 100 THEN 2
        WHEN i.duplicity_ratio > 0.9 THEN 3
        ELSE 4
    END,
    i.index_size DESC;

COMMENT ON VIEW vw_problematic_indexes IS 'Índices com problemas de desempenho, como baixo uso, bloat ou duplicidade';

-- View para estatísticas de tabelas problemáticas
CREATE OR REPLACE VIEW vw_problematic_tables AS
SELECT 
    t.id,
    t.timestamp,
    t.schema_name,
    t.table_name,
    t.table_size,
    pg_size_pretty(t.table_size) AS table_size_pretty,
    t.tuple_count,
    t.live_tuples,
    t.dead_tuples,
    t.sequential_scans,
    t.index_scans,
    t.vacuum_count,
    t.autovacuum_count,
    t.analyze_count,
    t.autoanalyze_count,
    t.last_vacuum,
    t.last_autovacuum,
    t.last_analyze,
    t.last_autoanalyze,
    t.bloat_ratio,
    t.fillfactor,
    CASE
        WHEN t.dead_tuples > t.live_tuples * 0.2 THEN 'High Dead Tuple Ratio'
        WHEN t.sequential_scans > t.index_scans * 10 AND t.tuple_count > 1000 THEN 'Excessive Sequential Scans'
        WHEN t.bloat_ratio > 0.3 THEN 'Bloated Table'
        WHEN NOW() - COALESCE(t.last_vacuum, t.last_autovacuum, '1970-01-01'::TIMESTAMP) > INTERVAL '7 days' AND t.dead_tuples > 1000 THEN 'Needs Vacuum'
        WHEN NOW() - COALESCE(t.last_analyze, t.last_autoanalyze, '1970-01-01'::TIMESTAMP) > INTERVAL '7 days' AND t.tuple_updates + t.tuple_inserts > 1000 THEN 'Needs Analyze'
        ELSE 'OK'
    END AS issue_type
FROM 
    table_stats t
WHERE 
    t.dead_tuples > t.live_tuples * 0.2 OR
    (t.sequential_scans > t.index_scans * 10 AND t.tuple_count > 1000) OR
    t.bloat_ratio > 0.3 OR
    (NOW() - COALESCE(t.last_vacuum, t.last_autovacuum, '1970-01-01'::TIMESTAMP) > INTERVAL '7 days' AND t.dead_tuples > 1000) OR
    (NOW() - COALESCE(t.last_analyze, t.last_autoanalyze, '1970-01-01'::TIMESTAMP) > INTERVAL '7 days' AND t.tuple_updates + t.tuple_inserts > 1000)
ORDER BY 
    CASE
        WHEN t.dead_tuples > t.live_tuples * 0.2 THEN 1
        WHEN t.bloat_ratio > 0.3 THEN 2
        WHEN t.sequential_scans > t.index_scans * 10 AND t.tuple_count > 1000 THEN 3
        ELSE 4
    END,
    t.table_size DESC;

COMMENT ON VIEW vw_problematic_tables IS 'Tabelas com problemas de desempenho, como alta proporção de tuplas mortas, scans sequenciais excessivos ou bloat';

-- View para tendências gerais de desempenho do banco de dados
CREATE OR REPLACE VIEW vw_database_performance_trends AS
SELECT 
    DATE_TRUNC('hour', timestamp) AS time_bucket,
    AVG(active_connections) AS avg_active_connections,
    MAX(active_connections) AS max_active_connections,
    AVG(idle_connections) AS avg_idle_connections,
    AVG(cache_hit_ratio) AS avg_cache_hit_ratio,
    SUM(deadlocks) AS total_deadlocks,
    SUM(conflicts) AS total_conflicts,
    AVG(temp_files_size) AS avg_temp_files_size,
    MAX(temp_files_size) AS max_temp_files_size,
    MAX(database_size) AS database_size
FROM 
    database_stats
WHERE 
    timestamp > NOW() - INTERVAL '7 days'
GROUP BY 
    time_bucket
ORDER BY 
    time_bucket;

COMMENT ON VIEW vw_database_performance_trends IS 'Tendências de métricas gerais de desempenho do banco de dados ao longo do tempo';

-- View para alertas de desempenho não resolvidos
CREATE OR REPLACE VIEW vw_unresolved_performance_alerts AS
SELECT 
    pa.id,
    pa.alert_type,
    pa.alert_level,
    pa.object_type,
    pa.object_name,
    pa.timestamp,
    pa.message,
    pa.details,
    pa.metric_value,
    pa.threshold_value,
    pa.is_acknowledged,
    pa.acknowledged_by,
    pa.acknowledged_at,
    pa.resolution_status,
    u1.username AS acknowledged_by_username,
    o.name AS organization_name,
    EXTRACT(EPOCH FROM (NOW() - pa.timestamp))/3600 AS hours_since_alert
FROM 
    performance_alerts pa
LEFT JOIN 
    iam.users u1 ON pa.acknowledged_by = u1.id
LEFT JOIN 
    iam.organizations o ON pa.organization_id = o.id
WHERE 
    pa.resolution_status IN ('pending', 'in_progress')
ORDER BY 
    CASE pa.alert_level
        WHEN 'critical' THEN 1
        WHEN 'high' THEN 2
        WHEN 'medium' THEN 3
        WHEN 'low' THEN 4
        ELSE 5
    END,
    pa.timestamp;

COMMENT ON VIEW vw_unresolved_performance_alerts IS 'Alertas de desempenho não resolvidos, ordenados por nível de severidade';

-- View para resumo de relatórios de desempenho
CREATE OR REPLACE VIEW vw_performance_reports_summary AS
SELECT 
    pr.id,
    pr.report_type,
    pr.title,
    pr.timestamp,
    pr.start_period,
    pr.end_period,
    pr.issues_found,
    pr.generated_by,
    pr.report_format,
    pr.status,
    o.name AS organization_name,
    jsonb_array_length(pr.recommendations) AS recommendation_count,
    (pr.metrics_summary->>'avg_query_duration_ms')::FLOAT AS avg_query_duration_ms,
    (pr.metrics_summary->>'slow_queries_count')::INTEGER AS slow_queries_count,
    (pr.metrics_summary->>'query_count')::INTEGER AS total_query_count
FROM 
    performance_reports pr
LEFT JOIN 
    iam.organizations o ON pr.organization_id = o.id
ORDER BY 
    pr.timestamp DESC;

COMMENT ON VIEW vw_performance_reports_summary IS 'Resumo dos relatórios de desempenho gerados';

-- View para problemas de fragmentação e bloat
CREATE OR REPLACE VIEW vw_fragmentation_issues AS
WITH index_bloat AS (
    SELECT 
        schema_name || '.' || table_name || '.' || index_name AS object_name,
        'index' AS object_type,
        bloat_ratio,
        index_size AS size_bytes,
        pg_size_pretty(index_size) AS size_pretty,
        timestamp
    FROM 
        index_stats
    WHERE 
        bloat_ratio > 0.3
    AND 
        timestamp = (SELECT MAX(timestamp) FROM index_stats)
),
table_bloat AS (
    SELECT 
        schema_name || '.' || table_name AS object_name,
        'table' AS object_type,
        bloat_ratio,
        table_size AS size_bytes,
        pg_size_pretty(table_size) AS size_pretty,
        timestamp
    FROM 
        table_stats
    WHERE 
        bloat_ratio > 0.3
    AND 
        timestamp = (SELECT MAX(timestamp) FROM table_stats)
)
SELECT * FROM index_bloat
UNION ALL
SELECT * FROM table_bloat
ORDER BY bloat_ratio DESC, size_bytes DESC;

COMMENT ON VIEW vw_fragmentation_issues IS 'Problemas de fragmentação (bloat) em tabelas e índices';

-- View para dashboard de desempenho
CREATE OR REPLACE VIEW vw_performance_dashboard AS
WITH recent_stats AS (
    SELECT * FROM database_stats
    WHERE timestamp = (SELECT MAX(timestamp) FROM database_stats)
),
slow_queries_last_24h AS (
    SELECT COUNT(*) AS count FROM slow_queries sq
    JOIN query_metrics qm ON sq.query_metrics_id = qm.id
    WHERE qm.timestamp > NOW() - INTERVAL '24 hours'
),
total_queries_last_24h AS (
    SELECT COUNT(*) AS count FROM query_metrics
    WHERE timestamp > NOW() - INTERVAL '24 hours'
),
avg_duration_last_24h AS (
    SELECT AVG(duration_ms) AS avg_duration FROM query_metrics
    WHERE timestamp > NOW() - INTERVAL '24 hours'
),
unresolved_alerts AS (
    SELECT COUNT(*) AS count,
    SUM(CASE WHEN alert_level = 'critical' THEN 1 ELSE 0 END) AS critical_count,
    SUM(CASE WHEN alert_level = 'high' THEN 1 ELSE 0 END) AS high_count
    FROM performance_alerts
    WHERE resolution_status IN ('pending', 'in_progress')
),
problem_indexes AS (
    SELECT COUNT(*) AS count FROM vw_problematic_indexes
),
problem_tables AS (
    SELECT COUNT(*) AS count FROM vw_problematic_tables
)
SELECT
    rs.active_connections,
    rs.idle_connections,
    rs.cache_hit_ratio,
    rs.deadlocks,
    rs.conflicts,
    pg_size_pretty(rs.temp_files_size) AS temp_files_size,
    pg_size_pretty(rs.database_size) AS database_size,
    sq.count AS slow_queries_count,
    tq.count AS total_queries_count,
    COALESCE(avg_duration.avg_duration, 0) AS avg_query_duration_ms,
    CASE
        WHEN tq.count > 0 THEN ROUND((sq.count::FLOAT / tq.count::FLOAT) * 100, 2)
        ELSE 0
    END AS slow_query_percentage,
    ua.count AS unresolved_alerts_count,
    ua.critical_count AS critical_alerts_count,
    ua.high_count AS high_alerts_count,
    pi.count AS problem_indexes_count,
    pt.count AS problem_tables_count,
    NOW() AS dashboard_timestamp
FROM
    recent_stats rs,
    slow_queries_last_24h sq,
    total_queries_last_24h tq,
    avg_duration_last_24h avg_duration,
    unresolved_alerts ua,
    problem_indexes pi,
    problem_tables pt;

COMMENT ON VIEW vw_performance_dashboard IS 'Dashboard consolidado com métricas atuais de desempenho';

-- View para consultas mais lentas
CREATE OR REPLACE VIEW vw_top_slowest_queries AS
SELECT 
    qm.id,
    qm.query_type,
    qm.query_text,
    qm.normalized_query_text,
    qm.query_hash,
    qm.duration_ms,
    qm.rows_affected,
    qm.table_scans,
    qm.index_scans,
    qm.timestamp,
    qm.organization_id,
    o.name AS organization_name,
    qm.database_name,
    EXISTS(SELECT 1 FROM slow_queries sq WHERE sq.query_metrics_id = qm.id) AS flagged_as_slow,
    ROW_NUMBER() OVER (PARTITION BY qm.query_hash ORDER BY qm.duration_ms DESC) AS rn
FROM 
    query_metrics qm
LEFT JOIN 
    iam.organizations o ON qm.organization_id = o.id
WHERE 
    qm.timestamp > NOW() - INTERVAL '24 hours'
ORDER BY 
    qm.duration_ms DESC
LIMIT 100;

COMMENT ON VIEW vw_top_slowest_queries IS 'As consultas mais lentas das últimas 24 horas';
