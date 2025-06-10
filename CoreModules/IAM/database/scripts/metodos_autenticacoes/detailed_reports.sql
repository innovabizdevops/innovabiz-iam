-- Relatórios Detalhados para Sistema de Autenticação

-- 1. Relatório de Performance Detalhado
CREATE OR REPLACE FUNCTION report.performance_report(
    p_start_date TIMESTAMP WITH TIME ZONE,
    p_end_date TIMESTAMP WITH TIME ZONE
)
RETURNS TABLE (
    metric_name VARCHAR(100),
    min_value NUMERIC,
    max_value NUMERIC,
    avg_value NUMERIC,
    percentile_95 NUMERIC,
    percentile_99 NUMERIC,
    total_requests BIGINT,
    failed_requests BIGINT,
    error_rate NUMERIC,
    hour TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    WITH hourly_metrics AS (
        SELECT 
            date_trunc('hour', timestamp) as hour,
            metric_name,
            latency_ms,
            success as is_success
        FROM performance_metrics
        WHERE timestamp >= p_start_date
        AND timestamp <= p_end_date
    )
    SELECT 
        metric_name,
        MIN(latency_ms) as min_value,
        MAX(latency_ms) as max_value,
        AVG(latency_ms) as avg_value,
        PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY latency_ms) as percentile_95,
        PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY latency_ms) as percentile_99,
        COUNT(*) as total_requests,
        COUNT(*) FILTER (WHERE NOT is_success) as failed_requests,
        (COUNT(*) FILTER (WHERE NOT is_success) * 100.0 / COUNT(*)) as error_rate,
        hour
    FROM hourly_metrics
    GROUP BY metric_name, hour
    ORDER BY hour, metric_name;
END;
$$ LANGUAGE plpgsql;

-- 2. Relatório de Segurança Detalhado
CREATE OR REPLACE FUNCTION report.security_report(
    p_start_date TIMESTAMP WITH TIME ZONE,
    p_end_date TIMESTAMP WITH TIME ZONE
)
RETURNS TABLE (
    user_id VARCHAR(100),
    authentication_type VARCHAR(100),
    success_count INTEGER,
    failure_count INTEGER,
    suspicious_count INTEGER,
    average_latency NUMERIC,
    max_latency NUMERIC,
    min_latency NUMERIC,
    country VARCHAR(100),
    device_type VARCHAR(100),
    browser VARCHAR(100)
) AS $$
BEGIN
    RETURN QUERY
    WITH user_metrics AS (
        SELECT 
            user_id,
            authentication_type,
            success,
            suspicious_activity,
            latency_ms,
            country,
            device_type,
            browser
        FROM security_metrics
        WHERE timestamp >= p_start_date
        AND timestamp <= p_end_date
    )
    SELECT 
        user_id,
        authentication_type,
        COUNT(*) FILTER (WHERE success) as success_count,
        COUNT(*) FILTER (WHERE NOT success) as failure_count,
        COUNT(*) FILTER (WHERE suspicious_activity) as suspicious_count,
        AVG(latency_ms) as average_latency,
        MAX(latency_ms) as max_latency,
        MIN(latency_ms) as min_latency,
        country,
        device_type,
        browser
    FROM user_metrics
    GROUP BY user_id, authentication_type, country, device_type, browser
    ORDER BY failure_count DESC, suspicious_count DESC;
END;
$$ LANGUAGE plpgsql;

-- 3. Relatório de Conformidade Detalhado
CREATE OR REPLACE FUNCTION report.compliance_report(
    p_start_date TIMESTAMP WITH TIME ZONE,
    p_end_date TIMESTAMP WITH TIME ZONE
)
RETURNS TABLE (
    regulation VARCHAR(100),
    audit_date TIMESTAMP WITH TIME ZONE,
    compliance_score NUMERIC,
    violations_count INTEGER,
    critical_violations INTEGER,
    high_risk_violations INTEGER,
    medium_risk_violations INTEGER,
    low_risk_violations INTEGER
) AS $$
BEGIN
    RETURN QUERY
    WITH violation_metrics AS (
        SELECT 
            regulation,
            audit_time,
            compliance_score,
            risk_level,
            COUNT(*) as violation_count
        FROM compliance_metrics
        WHERE timestamp >= p_start_date
        AND timestamp <= p_end_date
        AND compliance = false
        GROUP BY regulation, audit_time, compliance_score, risk_level
    )
    SELECT 
        regulation,
        audit_time as audit_date,
        compliance_score,
        SUM(violation_count) as violations_count,
        SUM(CASE WHEN risk_level = 'CRITICAL' THEN violation_count ELSE 0 END) as critical_violations,
        SUM(CASE WHEN risk_level = 'HIGH' THEN violation_count ELSE 0 END) as high_risk_violations,
        SUM(CASE WHEN risk_level = 'MEDIUM' THEN violation_count ELSE 0 END) as medium_risk_violations,
        SUM(CASE WHEN risk_level = 'LOW' THEN violation_count ELSE 0 END) as low_risk_violations
    FROM violation_metrics
    GROUP BY regulation, audit_time, compliance_score
    ORDER BY audit_time DESC, regulation;
END;
$$ LANGUAGE plpgsql;

-- 4. Relatório de Uso do Sistema
CREATE OR REPLACE FUNCTION report.usage_report(
    p_start_date TIMESTAMP WITH TIME ZONE,
    p_end_date TIMESTAMP WITH TIME ZONE
)
RETURNS TABLE (
    metric_name VARCHAR(100),
    total_count BIGINT,
    unique_users INTEGER,
    peak_hour TIMESTAMP WITH TIME ZONE,
    peak_value BIGINT,
    avg_value NUMERIC,
    hour TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    WITH hourly_metrics AS (
        SELECT 
            date_trunc('hour', timestamp) as hour,
            metric_name,
            COUNT(*) as count,
            COUNT(DISTINCT user_id) as unique_users
        FROM usage_metrics
        WHERE timestamp >= p_start_date
        AND timestamp <= p_end_date
        GROUP BY hour, metric_name
    )
    SELECT 
        metric_name,
        SUM(count) as total_count,
        MAX(unique_users) as unique_users,
        hour as peak_hour,
        count as peak_value,
        AVG(count) as avg_value,
        hour
    FROM hourly_metrics
    GROUP BY metric_name, hour, count
    ORDER BY metric_name, hour;
END;
$$ LANGUAGE plpgsql;

-- 5. Relatório de Recursos do Sistema
CREATE OR REPLACE FUNCTION report.resources_report(
    p_start_date TIMESTAMP WITH TIME ZONE,
    p_end_date TIMESTAMP WITH TIME ZONE
)
RETURNS TABLE (
    resource_type VARCHAR(100),
    min_usage NUMERIC,
    max_usage NUMERIC,
    avg_usage NUMERIC,
    percentile_95 NUMERIC,
    percentile_99 NUMERIC,
    peak_hour TIMESTAMP WITH TIME ZONE,
    peak_value NUMERIC,
    hour TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    WITH hourly_metrics AS (
        SELECT 
            date_trunc('hour', timestamp) as hour,
            resource_type,
            usage_percentage
        FROM resources_metrics
        WHERE timestamp >= p_start_date
        AND timestamp <= p_end_date
    )
    SELECT 
        resource_type,
        MIN(usage_percentage) as min_usage,
        MAX(usage_percentage) as max_usage,
        AVG(usage_percentage) as avg_usage,
        PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY usage_percentage) as percentile_95,
        PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY usage_percentage) as percentile_99,
        hour as peak_hour,
        usage_percentage as peak_value,
        hour
    FROM hourly_metrics
    GROUP BY resource_type, hour, usage_percentage
    ORDER BY resource_type, hour;
END;
$$ LANGUAGE plpgsql;

-- 6. Relatório de Autenticação por Método
CREATE OR REPLACE FUNCTION report.authentication_methods_report(
    p_start_date TIMESTAMP WITH TIME ZONE,
    p_end_date TIMESTAMP WITH TIME ZONE
)
RETURNS TABLE (
    method VARCHAR(100),
    total_auths INTEGER,
    success_rate NUMERIC,
    avg_latency NUMERIC,
    max_latency NUMERIC,
    min_latency NUMERIC,
    peak_hour TIMESTAMP WITH TIME ZONE,
    peak_auths INTEGER
) AS $$
BEGIN
    RETURN QUERY
    WITH hourly_metrics AS (
        SELECT 
            date_trunc('hour', timestamp) as hour,
            method,
            COUNT(*) as auth_count,
            AVG(latency_ms) as avg_latency,
            MAX(latency_ms) as max_latency,
            MIN(latency_ms) as min_latency
        FROM authentication_metrics
        WHERE timestamp >= p_start_date
        AND timestamp <= p_end_date
        GROUP BY hour, method
    )
    SELECT 
        method,
        SUM(auth_count) as total_auths,
        AVG(avg_latency) as avg_latency,
        MAX(max_latency) as max_latency,
        MIN(min_latency) as min_latency,
        hour as peak_hour,
        auth_count as peak_auths
    FROM hourly_metrics
    GROUP BY method, hour, auth_count
    ORDER BY total_auths DESC;
END;
$$ LANGUAGE plpgsql;

-- 7. Relatório de Geolocalização
CREATE OR REPLACE FUNCTION report.geolocation_report(
    p_start_date TIMESTAMP WITH TIME ZONE,
    p_end_date TIMESTAMP WITH TIME ZONE
)
RETURNS TABLE (
    country VARCHAR(100),
    total_auths INTEGER,
    success_rate NUMERIC,
    avg_latency NUMERIC,
    peak_hour TIMESTAMP WITH TIME ZONE,
    peak_auths INTEGER,
    continent VARCHAR(100),
    region VARCHAR(100)
) AS $$
BEGIN
    RETURN QUERY
    WITH hourly_metrics AS (
        SELECT 
            date_trunc('hour', timestamp) as hour,
            country,
            continent,
            region,
            COUNT(*) as auth_count,
            AVG(latency_ms) as avg_latency
        FROM geolocation_metrics
        WHERE timestamp >= p_start_date
        AND timestamp <= p_end_date
        GROUP BY hour, country, continent, region
    )
    SELECT 
        country,
        SUM(auth_count) as total_auths,
        AVG(avg_latency) as avg_latency,
        hour as peak_hour,
        auth_count as peak_auths,
        continent,
        region
    FROM hourly_metrics
    GROUP BY country, continent, region, hour, auth_count
    ORDER BY total_auths DESC;
END;
$$ LANGUAGE plpgsql;

-- 8. Relatório de Dispositivos
CREATE OR REPLACE FUNCTION report.devices_report(
    p_start_date TIMESTAMP WITH TIME ZONE,
    p_end_date TIMESTAMP WITH TIME ZONE
)
RETURNS TABLE (
    device_type VARCHAR(100),
    total_auths INTEGER,
    success_rate NUMERIC,
    avg_latency NUMERIC,
    peak_hour TIMESTAMP WITH TIME ZONE,
    peak_auths INTEGER,
    os_version VARCHAR(100),
    browser VARCHAR(100)
) AS $$
BEGIN
    RETURN QUERY
    WITH hourly_metrics AS (
        SELECT 
            date_trunc('hour', timestamp) as hour,
            device_type,
            os_version,
            browser,
            COUNT(*) as auth_count,
            AVG(latency_ms) as avg_latency
        FROM device_metrics
        WHERE timestamp >= p_start_date
        AND timestamp <= p_end_date
        GROUP BY hour, device_type, os_version, browser
    )
    SELECT 
        device_type,
        SUM(auth_count) as total_auths,
        AVG(avg_latency) as avg_latency,
        hour as peak_hour,
        auth_count as peak_auths,
        os_version,
        browser
    FROM hourly_metrics
    GROUP BY device_type, os_version, browser, hour, auth_count
    ORDER BY total_auths DESC;
END;
$$ LANGUAGE plpgsql;

-- 9. Relatório de APIs
CREATE OR REPLACE FUNCTION report.apis_report(
    p_start_date TIMESTAMP WITH TIME ZONE,
    p_end_date TIMESTAMP WITH TIME ZONE
)
RETURNS TABLE (
    api_endpoint VARCHAR(100),
    total_requests INTEGER,
    success_rate NUMERIC,
    avg_latency NUMERIC,
    peak_hour TIMESTAMP WITH TIME ZONE,
    peak_requests INTEGER,
    method VARCHAR(10),
    status_code INTEGER
) AS $$
BEGIN
    RETURN QUERY
    WITH hourly_metrics AS (
        SELECT 
            date_trunc('hour', timestamp) as hour,
            api_endpoint,
            method,
            status_code,
            COUNT(*) as request_count,
            AVG(latency_ms) as avg_latency
        FROM api_metrics
        WHERE timestamp >= p_start_date
        AND timestamp <= p_end_date
        GROUP BY hour, api_endpoint, method, status_code
    )
    SELECT 
        api_endpoint,
        SUM(request_count) as total_requests,
        AVG(avg_latency) as avg_latency,
        hour as peak_hour,
        request_count as peak_requests,
        method,
        status_code
    FROM hourly_metrics
    GROUP BY api_endpoint, method, status_code, hour, request_count
    ORDER BY total_requests DESC;
END;
$$ LANGUAGE plpgsql;

-- 10. Relatório de Cache
CREATE OR REPLACE FUNCTION report.cache_report(
    p_start_date TIMESTAMP WITH TIME ZONE,
    p_end_date TIMESTAMP WITH TIME ZONE
)
RETURNS TABLE (
    cache_type VARCHAR(100),
    total_operations INTEGER,
    hit_rate NUMERIC,
    miss_rate NUMERIC,
    avg_latency NUMERIC,
    peak_hour TIMESTAMP WITH TIME ZONE,
    peak_operations INTEGER,
    memory_usage BIGINT
) AS $$
BEGIN
    RETURN QUERY
    WITH hourly_metrics AS (
        SELECT 
            date_trunc('hour', timestamp) as hour,
            cache_type,
            COUNT(*) as operation_count,
            AVG(hit_rate) as avg_hit_rate,
            AVG(miss_rate) as avg_miss_rate,
            AVG(latency_ms) as avg_latency,
            AVG(memory_usage) as avg_memory_usage
        FROM cache_metrics
        WHERE timestamp >= p_start_date
        AND timestamp <= p_end_date
        GROUP BY hour, cache_type
    )
    SELECT 
        cache_type,
        SUM(operation_count) as total_operations,
        AVG(avg_hit_rate) as hit_rate,
        AVG(avg_miss_rate) as miss_rate,
        AVG(avg_latency) as avg_latency,
        hour as peak_hour,
        operation_count as peak_operations,
        AVG(avg_memory_usage) as memory_usage
    FROM hourly_metrics
    GROUP BY cache_type, hour, operation_count
    ORDER BY total_operations DESC;
END;
$$ LANGUAGE plpgsql;

-- 11. Relatório de Banco de Dados
CREATE OR REPLACE FUNCTION report.database_report(
    p_start_date TIMESTAMP WITH TIME ZONE,
    p_end_date TIMESTAMP WITH TIME ZONE
)
RETURNS TABLE (
    operation_type VARCHAR(100),
    total_queries INTEGER,
    success_rate NUMERIC,
    avg_latency NUMERIC,
    peak_hour TIMESTAMP WITH TIME ZONE,
    peak_queries INTEGER,
    memory_usage BIGINT,
    disk_usage BIGINT
) AS $$
BEGIN
    RETURN QUERY
    WITH hourly_metrics AS (
        SELECT 
            date_trunc('hour', timestamp) as hour,
            operation_type,
            COUNT(*) as query_count,
            AVG(success_rate) as avg_success_rate,
            AVG(latency_ms) as avg_latency,
            AVG(memory_usage) as avg_memory_usage,
            AVG(disk_usage) as avg_disk_usage
        FROM database_metrics
        WHERE timestamp >= p_start_date
        AND timestamp <= p_end_date
        GROUP BY hour, operation_type
    )
    SELECT 
        operation_type,
        SUM(query_count) as total_queries,
        AVG(avg_success_rate) as success_rate,
        AVG(avg_latency) as avg_latency,
        hour as peak_hour,
        query_count as peak_queries,
        AVG(avg_memory_usage) as memory_usage,
        AVG(avg_disk_usage) as disk_usage
    FROM hourly_metrics
    GROUP BY operation_type, hour, query_count
    ORDER BY total_queries DESC;
END;
$$ LANGUAGE plpgsql;

-- 12. Relatório de Criptografia
CREATE OR REPLACE FUNCTION report.encryption_report(
    p_start_date TIMESTAMP WITH TIME ZONE,
    p_end_date TIMESTAMP WITH TIME ZONE
)
RETURNS TABLE (
    algorithm VARCHAR(100),
    total_operations INTEGER,
    success_rate NUMERIC,
    avg_latency NUMERIC,
    peak_hour TIMESTAMP WITH TIME ZONE,
    peak_operations INTEGER,
    memory_usage BIGINT,
    cpu_usage NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    WITH hourly_metrics AS (
        SELECT 
            date_trunc('hour', timestamp) as hour,
            algorithm,
            COUNT(*) as operation_count,
            AVG(success_rate) as avg_success_rate,
            AVG(latency_ms) as avg_latency,
            AVG(memory_usage) as avg_memory_usage,
            AVG(cpu_usage) as avg_cpu_usage
        FROM encryption_metrics
        WHERE timestamp >= p_start_date
        AND timestamp <= p_end_date
        GROUP BY hour, algorithm
    )
    SELECT 
        algorithm,
        SUM(operation_count) as total_operations,
        AVG(avg_success_rate) as success_rate,
        AVG(avg_latency) as avg_latency,
        hour as peak_hour,
        operation_count as peak_operations,
        AVG(avg_memory_usage) as memory_usage,
        AVG(avg_cpu_usage) as cpu_usage
    FROM hourly_metrics
    GROUP BY algorithm, hour, operation_count
    ORDER BY total_operations DESC;
END;
$$ LANGUAGE plpgsql;

-- 13. Relatório de Autenticação Contínua
CREATE OR REPLACE FUNCTION report.continuous_auth_report(
    p_start_date TIMESTAMP WITH TIME ZONE,
    p_end_date TIMESTAMP WITH TIME ZONE
)
RETURNS TABLE (
    user_id VARCHAR(100),
    total_verifications INTEGER,
    success_rate NUMERIC,
    avg_latency NUMERIC,
    peak_hour TIMESTAMP WITH TIME ZONE,
    peak_verifications INTEGER,
    device_type VARCHAR(100),
    browser VARCHAR(100)
) AS $$
BEGIN
    RETURN QUERY
    WITH hourly_metrics AS (
        SELECT 
            date_trunc('hour', timestamp) as hour,
            user_id,
            device_type,
            browser,
            COUNT(*) as verification_count,
            AVG(latency_ms) as avg_latency
        FROM continuous_auth_metrics
        WHERE timestamp >= p_start_date
        AND timestamp <= p_end_date
        GROUP BY hour, user_id, device_type, browser
    )
    SELECT 
        user_id,
        SUM(verification_count) as total_verifications,
        AVG(avg_latency) as avg_latency,
        hour as peak_hour,
        verification_count as peak_verifications,
        device_type,
        browser
    FROM hourly_metrics
    GROUP BY user_id, device_type, browser, hour, verification_count
    ORDER BY total_verifications DESC;
END;
$$ LANGUAGE plpgsql;

-- 14. Relatório de Integração
CREATE OR REPLACE FUNCTION report.integration_report(
    p_start_date TIMESTAMP WITH TIME ZONE,
    p_end_date TIMESTAMP WITH TIME ZONE
)
RETURNS TABLE (
    integration_type VARCHAR(100),
    total_requests INTEGER,
    success_rate NUMERIC,
    avg_latency NUMERIC,
    peak_hour TIMESTAMP WITH TIME ZONE,
    peak_requests INTEGER,
    error_rate NUMERIC,
    response_size BIGINT
) AS $$
BEGIN
    RETURN QUERY
    WITH hourly_metrics AS (
        SELECT 
            date_trunc('hour', timestamp) as hour,
            integration_type,
            COUNT(*) as request_count,
            AVG(success_rate) as avg_success_rate,
            AVG(latency_ms) as avg_latency,
            AVG(response_size) as avg_response_size
        FROM integration_metrics
        WHERE timestamp >= p_start_date
        AND timestamp <= p_end_date
        GROUP BY hour, integration_type
    )
    SELECT 
        integration_type,
        SUM(request_count) as total_requests,
        AVG(avg_success_rate) as success_rate,
        AVG(avg_latency) as avg_latency,
        hour as peak_hour,
        request_count as peak_requests,
        (COUNT(*) FILTER (WHERE NOT success) * 100.0 / COUNT(*)) as error_rate,
        AVG(avg_response_size) as response_size
    FROM hourly_metrics
    GROUP BY integration_type, hour, request_count
    ORDER BY total_requests DESC;
END;
$$ LANGUAGE plpgsql;

-- Exemplo de uso
-- SELECT * FROM report.performance_report('2025-01-01', '2025-01-31');
-- SELECT * FROM report.security_report('2025-01-01', '2025-01-31');
-- SELECT * FROM report.compliance_report('2025-01-01', '2025-01-31');
