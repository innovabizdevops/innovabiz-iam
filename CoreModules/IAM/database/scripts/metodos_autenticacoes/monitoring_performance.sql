-- Sistema de Monitoramento de Performance e Métricas

-- Configuração inicial do monitoramento
CREATE OR REPLACE FUNCTION monitoring.setup_monitoring()
RETURNS VOID AS $$
BEGIN
    -- Criar tabela de métricas
    CREATE TABLE IF NOT EXISTS auth_metrics (
        metric_id SERIAL PRIMARY KEY,
        function_name TEXT NOT NULL,
        category TEXT NOT NULL,
        execution_time_ms FLOAT NOT NULL,
        memory_usage_kb FLOAT,
        cpu_usage_percent FLOAT,
        success_rate FLOAT,
        error_rate FLOAT,
        timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        environment TEXT,
        instance_id TEXT
    );
    
    -- Criar tabela de alertas
    CREATE TABLE IF NOT EXISTS auth_alerts (
        alert_id SERIAL PRIMARY KEY,
        metric_id INTEGER REFERENCES auth_metrics(metric_id),
        alert_type TEXT NOT NULL,
        severity TEXT NOT NULL,
        description TEXT,
        threshold_value FLOAT,
        current_value FLOAT,
        notified BOOLEAN DEFAULT false,
        timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
END;
$$ LANGUAGE plpgsql;

-- Função para registrar métricas
CREATE OR REPLACE FUNCTION monitoring.register_metric(
    p_function_name TEXT,
    p_category TEXT,
    p_execution_time_ms FLOAT,
    p_memory_usage_kb FLOAT,
    p_cpu_usage_percent FLOAT,
    p_success_rate FLOAT,
    p_error_rate FLOAT,
    p_environment TEXT,
    p_instance_id TEXT
)
RETURNS INTEGER AS $$
DECLARE
    v_metric_id INTEGER;
BEGIN
    INSERT INTO auth_metrics (
        function_name,
        category,
        execution_time_ms,
        memory_usage_kb,
        cpu_usage_percent,
        success_rate,
        error_rate,
        environment,
        instance_id
    ) VALUES (
        p_function_name,
        p_category,
        p_execution_time_ms,
        p_memory_usage_kb,
        p_cpu_usage_percent,
        p_success_rate,
        p_error_rate,
        p_environment,
        p_instance_id
    ) RETURNING metric_id INTO v_metric_id;
    
    RETURN v_metric_id;
END;
$$ LANGUAGE plpgsql;

-- Função para gerar alertas
CREATE OR REPLACE FUNCTION monitoring.generate_alert(
    p_metric_id INTEGER,
    p_alert_type TEXT,
    p_severity TEXT,
    p_description TEXT,
    p_threshold_value FLOAT,
    p_current_value FLOAT
)
RETURNS INTEGER AS $$
DECLARE
    v_alert_id INTEGER;
BEGIN
    INSERT INTO auth_alerts (
        metric_id,
        alert_type,
        severity,
        description,
        threshold_value,
        current_value
    ) VALUES (
        p_metric_id,
        p_alert_type,
        p_severity,
        p_description,
        p_threshold_value,
        p_current_value
    ) RETURNING alert_id INTO v_alert_id;
    
    RETURN v_alert_id;
END;
$$ LANGUAGE plpgsql;

-- Função para obter métricas por categoria
CREATE OR REPLACE FUNCTION monitoring.get_metrics_by_category(
    p_category TEXT,
    p_time_window INTERVAL DEFAULT INTERVAL '24 hours'
)
RETURNS TABLE (
    function_name TEXT,
    avg_execution_time_ms FLOAT,
    avg_memory_usage_kb FLOAT,
    avg_cpu_usage_percent FLOAT,
    success_rate FLOAT,
    error_rate FLOAT,
    total_executions INTEGER,
    timestamp TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        function_name,
        AVG(execution_time_ms) as avg_execution_time_ms,
        AVG(memory_usage_kb) as avg_memory_usage_kb,
        AVG(cpu_usage_percent) as avg_cpu_usage_percent,
        AVG(success_rate) as success_rate,
        AVG(error_rate) as error_rate,
        COUNT(*) as total_executions,
        timestamp
    FROM auth_metrics
    WHERE category = p_category
    AND timestamp >= NOW() - p_time_window
    GROUP BY function_name, timestamp
    ORDER BY timestamp DESC;
END;
$$ LANGUAGE plpgsql;

-- Função para gerar relatório de performance
CREATE OR REPLACE FUNCTION monitoring.generate_performance_report(
    p_time_window INTERVAL DEFAULT INTERVAL '24 hours'
)
RETURNS TABLE (
    category TEXT,
    avg_execution_time_ms FLOAT,
    avg_memory_usage_kb FLOAT,
    avg_cpu_usage_percent FLOAT,
    success_rate FLOAT,
    error_rate FLOAT,
    total_executions INTEGER,
    alerts_count INTEGER,
    critical_alerts_count INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        m.category,
        AVG(m.execution_time_ms) as avg_execution_time_ms,
        AVG(m.memory_usage_kb) as avg_memory_usage_kb,
        AVG(m.cpu_usage_percent) as avg_cpu_usage_percent,
        AVG(m.success_rate) as success_rate,
        AVG(m.error_rate) as error_rate,
        COUNT(DISTINCT m.metric_id) as total_executions,
        COUNT(DISTINCT a.alert_id) as alerts_count,
        COUNT(CASE WHEN a.severity = 'CRITICAL' THEN 1 END) as critical_alerts_count
    FROM auth_metrics m
    LEFT JOIN auth_alerts a ON m.metric_id = a.metric_id
    WHERE m.timestamp >= NOW() - p_time_window
    GROUP BY m.category
    ORDER BY m.category;
END;
$$ LANGUAGE plpgsql;

-- Função para limpar métricas antigas
CREATE OR REPLACE FUNCTION monitoring.cleanup_old_metrics(
    p_retention_days INTEGER DEFAULT 30
)
RETURNS INTEGER AS $$
DECLARE
    v_deleted_count INTEGER;
BEGIN
    -- Limpar métricas antigas
    DELETE FROM auth_metrics
    WHERE timestamp < NOW() - INTERVAL '1 day' * p_retention_days
    RETURNING COUNT(*) INTO v_deleted_count;
    
    -- Limpar alertas órfãos
    DELETE FROM auth_alerts
    WHERE metric_id NOT IN (SELECT metric_id FROM auth_metrics);
    
    RETURN v_deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Exemplo de uso:
-- SELECT monitoring.setup_monitoring();
-- SELECT monitoring.register_metric(
--     'auth.verify_password',
--     'KB-01',
--     150.5,
--     256.0,
--     10.2,
--     0.99,
--     0.01,
--     'production',
--     'instance-01'
-- );
-- SELECT monitoring.generate_performance_report();
-- SELECT monitoring.cleanup_old_metrics();
