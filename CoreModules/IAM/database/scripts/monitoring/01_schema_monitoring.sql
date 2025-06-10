-- INNOVABIZ - IAM Database Monitoring Schema
-- Author: Eduardo Jeremias
-- Date: 08/05/2025
-- Version: 1.0
-- Description: Esquema para monitoramento de desempenho do banco de dados.

-- Criar esquema para monitoramento
CREATE SCHEMA IF NOT EXISTS iam_monitoring;

-- Configurar caminho de busca
SET search_path TO iam_monitoring, iam, public;

-- Criar extensões necessárias se não existirem
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "hstore";

-- Tabela para armazenar métricas de consultas
CREATE TABLE IF NOT EXISTS query_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    query_type VARCHAR(100) NOT NULL,
    query_text TEXT NOT NULL,
    normalized_query_text TEXT,
    query_hash TEXT,
    duration_ms FLOAT NOT NULL,
    rows_affected INTEGER NOT NULL,
    table_scans INTEGER NOT NULL DEFAULT 0,
    index_scans INTEGER NOT NULL DEFAULT 0,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    organization_id UUID,
    user_id UUID,
    database_name VARCHAR(100),
    execution_plan JSONB,
    execution_context JSONB,
    tags VARCHAR[] DEFAULT ARRAY[]::VARCHAR[]
);

CREATE INDEX IF NOT EXISTS idx_query_metrics_query_type ON query_metrics(query_type);
CREATE INDEX IF NOT EXISTS idx_query_metrics_duration ON query_metrics(duration_ms);
CREATE INDEX IF NOT EXISTS idx_query_metrics_timestamp ON query_metrics(timestamp);
CREATE INDEX IF NOT EXISTS idx_query_metrics_organization_id ON query_metrics(organization_id);
CREATE INDEX IF NOT EXISTS idx_query_metrics_query_hash ON query_metrics(query_hash);

-- Particionamento por tempo para métricas de consultas (opcional)
-- Considere implementar particionamento para tabelas de alta volumetria

-- Tabela para armazenar consultas lentas
CREATE TABLE IF NOT EXISTS slow_queries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    query_metrics_id UUID REFERENCES query_metrics(id),
    threshold_ms FLOAT NOT NULL,
    analysis TEXT,
    recommendations JSONB,
    is_resolved BOOLEAN DEFAULT FALSE,
    resolution_notes TEXT,
    resolved_by UUID,
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_slow_queries_query_metrics_id ON slow_queries(query_metrics_id);
CREATE INDEX IF NOT EXISTS idx_slow_queries_is_resolved ON slow_queries(is_resolved);
CREATE INDEX IF NOT EXISTS idx_slow_queries_created_at ON slow_queries(created_at);

-- Tabela para estatísticas gerais de banco de dados
CREATE TABLE IF NOT EXISTS database_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    active_connections INTEGER NOT NULL,
    idle_connections INTEGER NOT NULL,
    cache_hit_ratio FLOAT,
    deadlocks INTEGER,
    conflicts INTEGER,
    temp_files_size BIGINT,
    database_size BIGINT,
    transaction_count BIGINT,
    checkpoint_stats JSONB,
    locks_stats JSONB,
    vacuum_stats JSONB,
    bgwriter_stats JSONB,
    snapshot_duration_ms FLOAT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_database_stats_timestamp ON database_stats(timestamp);

-- Tabela para estatísticas de índices
CREATE TABLE IF NOT EXISTS index_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    schema_name VARCHAR(100) NOT NULL,
    table_name VARCHAR(100) NOT NULL,
    index_name VARCHAR(100) NOT NULL,
    index_size BIGINT NOT NULL,
    index_scans BIGINT NOT NULL,
    tuple_reads BIGINT NOT NULL,
    tuple_fetches BIGINT NOT NULL,
    usage_ratio FLOAT,
    bloat_ratio FLOAT,
    duplicity_ratio FLOAT,
    last_vacuum TIMESTAMP WITH TIME ZONE,
    last_analyze TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_index_stats_timestamp ON index_stats(timestamp);
CREATE INDEX IF NOT EXISTS idx_index_stats_schema_table ON index_stats(schema_name, table_name);
CREATE INDEX IF NOT EXISTS idx_index_stats_usage_ratio ON index_stats(usage_ratio);
CREATE INDEX IF NOT EXISTS idx_index_stats_bloat_ratio ON index_stats(bloat_ratio);

-- Tabela para estatísticas de tabelas
CREATE TABLE IF NOT EXISTS table_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    schema_name VARCHAR(100) NOT NULL,
    table_name VARCHAR(100) NOT NULL,
    table_size BIGINT NOT NULL,
    tuple_count BIGINT NOT NULL,
    tuple_inserts BIGINT NOT NULL,
    tuple_updates BIGINT NOT NULL,
    tuple_deletes BIGINT NOT NULL,
    live_tuples BIGINT NOT NULL,
    dead_tuples BIGINT NOT NULL,
    sequential_scans BIGINT NOT NULL,
    sequential_scan_read_tuples BIGINT NOT NULL,
    index_scans BIGINT NOT NULL,
    index_scan_read_tuples BIGINT NOT NULL,
    vacuum_count BIGINT NOT NULL,
    autovacuum_count BIGINT NOT NULL,
    analyze_count BIGINT NOT NULL,
    autoanalyze_count BIGINT NOT NULL,
    last_vacuum TIMESTAMP WITH TIME ZONE,
    last_autovacuum TIMESTAMP WITH TIME ZONE,
    last_analyze TIMESTAMP WITH TIME ZONE,
    last_autoanalyze TIMESTAMP WITH TIME ZONE,
    bloat_ratio FLOAT,
    fillfactor INTEGER
);

CREATE INDEX IF NOT EXISTS idx_table_stats_timestamp ON table_stats(timestamp);
CREATE INDEX IF NOT EXISTS idx_table_stats_schema_table ON table_stats(schema_name, table_name);
CREATE INDEX IF NOT EXISTS idx_table_stats_size ON table_stats(table_size);
CREATE INDEX IF NOT EXISTS idx_table_stats_bloat_ratio ON table_stats(bloat_ratio);

-- Tabela para relatórios de desempenho de banco de dados
CREATE TABLE IF NOT EXISTS performance_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    report_type VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    start_period TIMESTAMP WITH TIME ZONE NOT NULL,
    end_period TIMESTAMP WITH TIME ZONE NOT NULL,
    metrics_summary JSONB NOT NULL,
    issues_found INTEGER NOT NULL DEFAULT 0,
    recommendations JSONB,
    visualizations JSONB,
    generated_by VARCHAR(255),
    report_format VARCHAR(50),
    organization_id UUID,
    distribution_list VARCHAR[],
    is_scheduled BOOLEAN DEFAULT FALSE,
    status VARCHAR(50) DEFAULT 'generated'
);

CREATE INDEX IF NOT EXISTS idx_performance_reports_timestamp ON performance_reports(timestamp);
CREATE INDEX IF NOT EXISTS idx_performance_reports_report_type ON performance_reports(report_type);
CREATE INDEX IF NOT EXISTS idx_performance_reports_organization_id ON performance_reports(organization_id);
CREATE INDEX IF NOT EXISTS idx_performance_reports_status ON performance_reports(status);

-- Tabela para agendamento de coleta de métricas
CREATE TABLE IF NOT EXISTS metrics_collection_schedule (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    collector_type VARCHAR(100) NOT NULL,
    frequency_minutes INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    last_run TIMESTAMP WITH TIME ZONE,
    next_run TIMESTAMP WITH TIME ZONE,
    settings JSONB DEFAULT '{}'::JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID,
    retention_days INTEGER DEFAULT 90
);

CREATE INDEX IF NOT EXISTS idx_metrics_collection_schedule_is_active ON metrics_collection_schedule(is_active);
CREATE INDEX IF NOT EXISTS idx_metrics_collection_schedule_next_run ON metrics_collection_schedule(next_run);

-- Tabela para alertas de desempenho
CREATE TABLE IF NOT EXISTS performance_alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    alert_type VARCHAR(100) NOT NULL,
    alert_level VARCHAR(50) NOT NULL,
    object_type VARCHAR(100) NOT NULL,
    object_name VARCHAR(255) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    message TEXT NOT NULL,
    details JSONB,
    metric_value FLOAT,
    threshold_value FLOAT,
    is_acknowledged BOOLEAN DEFAULT FALSE,
    acknowledged_by UUID,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    resolution_status VARCHAR(50) DEFAULT 'pending',
    resolution_notes TEXT,
    resolved_by UUID,
    resolved_at TIMESTAMP WITH TIME ZONE,
    notification_sent BOOLEAN DEFAULT FALSE,
    organization_id UUID
);

CREATE INDEX IF NOT EXISTS idx_performance_alerts_timestamp ON performance_alerts(timestamp);
CREATE INDEX IF NOT EXISTS idx_performance_alerts_alert_type ON performance_alerts(alert_type);
CREATE INDEX IF NOT EXISTS idx_performance_alerts_alert_level ON performance_alerts(alert_level);
CREATE INDEX IF NOT EXISTS idx_performance_alerts_is_acknowledged ON performance_alerts(is_acknowledged);
CREATE INDEX IF NOT EXISTS idx_performance_alerts_resolution_status ON performance_alerts(resolution_status);
CREATE INDEX IF NOT EXISTS idx_performance_alerts_organization_id ON performance_alerts(organization_id);

-- Função para atualizar o timestamp 'updated_at'
CREATE OR REPLACE FUNCTION update_monitoring_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Aplicar a função aos triggers
CREATE TRIGGER update_metrics_collection_schedule_updated_at
BEFORE UPDATE ON metrics_collection_schedule
FOR EACH ROW EXECUTE FUNCTION update_monitoring_updated_at_column();

COMMENT ON SCHEMA iam_monitoring IS 'Esquema para monitoramento de desempenho do banco de dados do IAM';
COMMENT ON TABLE query_metrics IS 'Armazena métricas de desempenho de consultas SQL executadas';
COMMENT ON TABLE slow_queries IS 'Registra consultas que excedem o limiar de desempenho';
COMMENT ON TABLE database_stats IS 'Estatísticas gerais do banco de dados coletadas periodicamente';
COMMENT ON TABLE index_stats IS 'Estatísticas de desempenho de índices';
COMMENT ON TABLE table_stats IS 'Estatísticas de desempenho de tabelas';
COMMENT ON TABLE performance_reports IS 'Relatórios de desempenho gerados';
COMMENT ON TABLE metrics_collection_schedule IS 'Configuração de agendamento para coleta de métricas';
COMMENT ON TABLE performance_alerts IS 'Alertas de problemas de desempenho detectados';

-- Inserir configurações iniciais de coleta de métricas
INSERT INTO metrics_collection_schedule 
    (collector_type, frequency_minutes, is_active, next_run, settings, retention_days)
VALUES
    ('query_metrics', 5, TRUE, NOW() + INTERVAL '5 minutes', 
     '{"slow_query_threshold_ms": 1000, "collect_plans": true, "sample_rate": 0.1}'::JSONB, 90),
    ('database_stats', 15, TRUE, NOW() + INTERVAL '15 minutes', 
     '{"include_detailed_stats": true}'::JSONB, 180),
    ('index_stats', 60, TRUE, NOW() + INTERVAL '1 hour', 
     '{"include_bloat_analysis": true}'::JSONB, 365),
    ('table_stats', 60, TRUE, NOW() + INTERVAL '1 hour', 
     '{"include_bloat_analysis": true}'::JSONB, 365),
    ('performance_report', 1440, TRUE, NOW() + INTERVAL '1 day', 
     '{"report_types": ["daily_summary", "slow_queries", "index_recommendations", "vacuum_recommendations"]}'::JSONB, 730)
ON CONFLICT DO NOTHING;
