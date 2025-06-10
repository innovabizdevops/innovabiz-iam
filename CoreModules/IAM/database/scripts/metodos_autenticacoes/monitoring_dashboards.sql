-- Dashboards de Monitoramento para Sistema de Autenticação

-- 1. Dashboard de Performance Geral
CREATE OR REPLACE FUNCTION dashboard.performance_dashboard()
RETURNS TABLE (
    metric_name VARCHAR(100),
    current_value NUMERIC,
    threshold NUMERIC,
    status VARCHAR(20),
    trend VARCHAR(20),
    last_update TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        'Latency' as metric_name,
        AVG(latency_ms) as current_value,
        100 as threshold,
        CASE 
            WHEN AVG(latency_ms) > 100 THEN 'CRITICAL'
            WHEN AVG(latency_ms) > 50 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        CASE 
            WHEN AVG(latency_ms) > COALESCE(LAG(AVG(latency_ms)) OVER (ORDER BY timestamp), 0) THEN 'UP'
            ELSE 'DOWN'
        END as trend,
        MAX(timestamp) as last_update
    FROM performance_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    GROUP BY metric_name
    
    UNION ALL
    
    SELECT 
        'Throughput' as metric_name,
        AVG(throughput) as current_value,
        1000 as threshold,
        CASE 
            WHEN AVG(throughput) < 1000 THEN 'CRITICAL'
            WHEN AVG(throughput) < 2000 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        CASE 
            WHEN AVG(throughput) < COALESCE(LAG(AVG(throughput)) OVER (ORDER BY timestamp), 0) THEN 'DOWN'
            ELSE 'UP'
        END as trend,
        MAX(timestamp) as last_update
    FROM performance_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    GROUP BY metric_name
    
    UNION ALL
    
    SELECT 
        'Error Rate' as metric_name,
        AVG(error_rate) as current_value,
        1 as threshold,
        CASE 
            WHEN AVG(error_rate) > 1 THEN 'CRITICAL'
            WHEN AVG(error_rate) > 0.5 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        CASE 
            WHEN AVG(error_rate) > COALESCE(LAG(AVG(error_rate)) OVER (ORDER BY timestamp), 0) THEN 'UP'
            ELSE 'DOWN'
        END as trend,
        MAX(timestamp) as last_update
    FROM performance_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    GROUP BY metric_name;
END;
$$ LANGUAGE plpgsql;

-- 2. Dashboard de Segurança
CREATE OR REPLACE FUNCTION dashboard.security_dashboard()
RETURNS TABLE (
    metric_name VARCHAR(100),
    current_value NUMERIC,
    threshold NUMERIC,
    status VARCHAR(20),
    trend VARCHAR(20),
    last_update TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        'Authentication Success Rate' as metric_name,
        AVG(success_rate) as current_value,
        99 as threshold,
        CASE 
            WHEN AVG(success_rate) < 99 THEN 'CRITICAL'
            WHEN AVG(success_rate) < 99.5 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        CASE 
            WHEN AVG(success_rate) < COALESCE(LAG(AVG(success_rate)) OVER (ORDER BY timestamp), 0) THEN 'DOWN'
            ELSE 'UP'
        END as trend,
        MAX(timestamp) as last_update
    FROM security_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    GROUP BY metric_name
    
    UNION ALL
    
    SELECT 
        'Failed Login Attempts' as metric_name,
        COUNT(*) as current_value,
        100 as threshold,
        CASE 
            WHEN COUNT(*) > 100 THEN 'CRITICAL'
            WHEN COUNT(*) > 50 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        CASE 
            WHEN COUNT(*) > COALESCE(LAG(COUNT(*)) OVER (ORDER BY timestamp), 0) THEN 'UP'
            ELSE 'DOWN'
        END as trend,
        MAX(timestamp) as last_update
    FROM security_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    AND success = false
    GROUP BY metric_name
    
    UNION ALL
    
    SELECT 
        'Suspicious Activities' as metric_name,
        COUNT(*) as current_value,
        10 as threshold,
        CASE 
            WHEN COUNT(*) > 10 THEN 'CRITICAL'
            WHEN COUNT(*) > 5 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        CASE 
            WHEN COUNT(*) > COALESCE(LAG(COUNT(*)) OVER (ORDER BY timestamp), 0) THEN 'UP'
            ELSE 'DOWN'
        END as trend,
        MAX(timestamp) as last_update
    FROM security_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    AND suspicious_activity = true
    GROUP BY metric_name;
END;
$$ LANGUAGE plpgsql;

-- 3. Dashboard de Conformidade
CREATE OR REPLACE FUNCTION dashboard.compliance_dashboard()
RETURNS TABLE (
    regulation VARCHAR(100),
    compliance_rate NUMERIC,
    violations_count INTEGER,
    last_audit TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20),
    trend VARCHAR(20)
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        'GDPR' as regulation,
        AVG(compliance_score) as compliance_rate,
        COUNT(*) FILTER (WHERE compliance = false) as violations_count,
        MAX(audit_time) as last_audit,
        CASE 
            WHEN AVG(compliance_score) < 90 THEN 'CRITICAL'
            WHEN AVG(compliance_score) < 95 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        CASE 
            WHEN AVG(compliance_score) < COALESCE(LAG(AVG(compliance_score)) OVER (ORDER BY audit_time), 0) THEN 'DOWN'
            ELSE 'UP'
        END as trend
    FROM compliance_metrics
    WHERE regulation = 'GDPR'
    AND audit_time >= NOW() - INTERVAL '1 month'
    GROUP BY regulation
    
    UNION ALL
    
    SELECT 
        'PCI DSS' as regulation,
        AVG(compliance_score) as compliance_rate,
        COUNT(*) FILTER (WHERE compliance = false) as violations_count,
        MAX(audit_time) as last_audit,
        CASE 
            WHEN AVG(compliance_score) < 90 THEN 'CRITICAL'
            WHEN AVG(compliance_score) < 95 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        CASE 
            WHEN AVG(compliance_score) < COALESCE(LAG(AVG(compliance_score)) OVER (ORDER BY audit_time), 0) THEN 'DOWN'
            ELSE 'UP'
        END as trend
    FROM compliance_metrics
    WHERE regulation = 'PCI DSS'
    AND audit_time >= NOW() - INTERVAL '1 month'
    GROUP BY regulation
    
    UNION ALL
    
    SELECT 
        'HIPAA' as regulation,
        AVG(compliance_score) as compliance_rate,
        COUNT(*) FILTER (WHERE compliance = false) as violations_count,
        MAX(audit_time) as last_audit,
        CASE 
            WHEN AVG(compliance_score) < 90 THEN 'CRITICAL'
            WHEN AVG(compliance_score) < 95 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        CASE 
            WHEN AVG(compliance_score) < COALESCE(LAG(AVG(compliance_score)) OVER (ORDER BY audit_time), 0) THEN 'DOWN'
            ELSE 'UP'
        END as trend
    FROM compliance_metrics
    WHERE regulation = 'HIPAA'
    AND audit_time >= NOW() - INTERVAL '1 month'
    GROUP BY regulation;
END;
$$ LANGUAGE plpgsql;

-- 4. Dashboard de Uso do Sistema
CREATE OR REPLACE FUNCTION dashboard.usage_dashboard()
RETURNS TABLE (
    metric_name VARCHAR(100),
    current_value NUMERIC,
    threshold NUMERIC,
    status VARCHAR(20),
    trend VARCHAR(20),
    last_update TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        'Active Users' as metric_name,
        COUNT(DISTINCT user_id) as current_value,
        10000 as threshold,
        CASE 
            WHEN COUNT(DISTINCT user_id) > 10000 THEN 'CRITICAL'
            WHEN COUNT(DISTINCT user_id) > 5000 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        CASE 
            WHEN COUNT(DISTINCT user_id) > COALESCE(LAG(COUNT(DISTINCT user_id)) OVER (ORDER BY timestamp), 0) THEN 'UP'
            ELSE 'DOWN'
        END as trend,
        MAX(timestamp) as last_update
    FROM usage_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    GROUP BY metric_name
    
    UNION ALL
    
    SELECT 
        'API Requests' as metric_name,
        COUNT(*) as current_value,
        100000 as threshold,
        CASE 
            WHEN COUNT(*) > 100000 THEN 'CRITICAL'
            WHEN COUNT(*) > 50000 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        CASE 
            WHEN COUNT(*) > COALESCE(LAG(COUNT(*)) OVER (ORDER BY timestamp), 0) THEN 'UP'
            ELSE 'DOWN'
        END as trend,
        MAX(timestamp) as last_update
    FROM usage_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    GROUP BY metric_name
    
    UNION ALL
    
    SELECT 
        'Session Duration' as metric_name,
        AVG(session_duration) as current_value,
        30 as threshold,
        CASE 
            WHEN AVG(session_duration) > 30 THEN 'CRITICAL'
            WHEN AVG(session_duration) > 15 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        CASE 
            WHEN AVG(session_duration) > COALESCE(LAG(AVG(session_duration)) OVER (ORDER BY timestamp), 0) THEN 'UP'
            ELSE 'DOWN'
        END as trend,
        MAX(timestamp) as last_update
    FROM usage_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    GROUP BY metric_name;
END;
$$ LANGUAGE plpgsql;

-- 5. Dashboard de Recursos
CREATE OR REPLACE FUNCTION dashboard.resources_dashboard()
RETURNS TABLE (
    resource_type VARCHAR(100),
    usage_percentage NUMERIC,
    threshold NUMERIC,
    status VARCHAR(20),
    trend VARCHAR(20),
    last_update TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        'CPU Usage' as resource_type,
        AVG(cpu_usage) as usage_percentage,
        80 as threshold,
        CASE 
            WHEN AVG(cpu_usage) > 80 THEN 'CRITICAL'
            WHEN AVG(cpu_usage) > 60 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        CASE 
            WHEN AVG(cpu_usage) > COALESCE(LAG(AVG(cpu_usage)) OVER (ORDER BY timestamp), 0) THEN 'UP'
            ELSE 'DOWN'
        END as trend,
        MAX(timestamp) as last_update
    FROM resources_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    GROUP BY resource_type
    
    UNION ALL
    
    SELECT 
        'Memory Usage' as resource_type,
        AVG(memory_usage) as usage_percentage,
        80 as threshold,
        CASE 
            WHEN AVG(memory_usage) > 80 THEN 'CRITICAL'
            WHEN AVG(memory_usage) > 60 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        CASE 
            WHEN AVG(memory_usage) > COALESCE(LAG(AVG(memory_usage)) OVER (ORDER BY timestamp), 0) THEN 'UP'
            ELSE 'DOWN'
        END as trend,
        MAX(timestamp) as last_update
    FROM resources_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    GROUP BY resource_type
    
    UNION ALL
    
    SELECT 
        'Disk Usage' as resource_type,
        AVG(disk_usage) as usage_percentage,
        90 as threshold,
        CASE 
            WHEN AVG(disk_usage) > 90 THEN 'CRITICAL'
            WHEN AVG(disk_usage) > 80 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        CASE 
            WHEN AVG(disk_usage) > COALESCE(LAG(AVG(disk_usage)) OVER (ORDER BY timestamp), 0) THEN 'UP'
            ELSE 'DOWN'
        END as trend,
        MAX(timestamp) as last_update
    FROM resources_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    GROUP BY resource_type;
END;
$$ LANGUAGE plpgsql;

-- 6. Dashboard de Autenticação por Método
CREATE OR REPLACE FUNCTION dashboard.authentication_methods_dashboard()
RETURNS TABLE (
    method VARCHAR(100),
    success_rate NUMERIC,
    failure_rate NUMERIC,
    average_latency NUMERIC,
    last_update TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        method,
        AVG(success_rate) as success_rate,
        AVG(failure_rate) as failure_rate,
        AVG(latency_ms) as average_latency,
        MAX(timestamp) as last_update
    FROM authentication_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    GROUP BY method
    ORDER BY AVG(latency_ms) DESC;
END;
$$ LANGUAGE plpgsql;

-- 7. Dashboard de Geolocalização
CREATE OR REPLACE FUNCTION dashboard.geolocation_dashboard()
RETURNS TABLE (
    country VARCHAR(100),
    authentication_count INTEGER,
    success_rate NUMERIC,
    average_latency NUMERIC,
    last_update TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        country,
        COUNT(*) as authentication_count,
        AVG(success_rate) as success_rate,
        AVG(latency_ms) as average_latency,
        MAX(timestamp) as last_update
    FROM geolocation_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    GROUP BY country
    ORDER BY COUNT(*) DESC;
END;
$$ LANGUAGE plpgsql;

-- 8. Dashboard de Dispositivos
CREATE OR REPLACE FUNCTION dashboard.devices_dashboard()
RETURNS TABLE (
    device_type VARCHAR(100),
    authentication_count INTEGER,
    success_rate NUMERIC,
    average_latency NUMERIC,
    last_update TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        device_type,
        COUNT(*) as authentication_count,
        AVG(success_rate) as success_rate,
        AVG(latency_ms) as average_latency,
        MAX(timestamp) as last_update
    FROM device_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    GROUP BY device_type
    ORDER BY COUNT(*) DESC;
END;
$$ LANGUAGE plpgsql;

-- 9. Dashboard de Browsers
CREATE OR REPLACE FUNCTION dashboard.browsers_dashboard()
RETURNS TABLE (
    browser VARCHAR(100),
    authentication_count INTEGER,
    success_rate NUMERIC,
    average_latency NUMERIC,
    last_update TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        browser,
        COUNT(*) as authentication_count,
        AVG(success_rate) as success_rate,
        AVG(latency_ms) as average_latency,
        MAX(timestamp) as last_update
    FROM browser_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    GROUP BY browser
    ORDER BY COUNT(*) DESC;
END;
$$ LANGUAGE plpgsql;

-- 10. Dashboard de APIs
CREATE OR REPLACE FUNCTION dashboard.apis_dashboard()
RETURNS TABLE (
    api_endpoint VARCHAR(100),
    request_count INTEGER,
    success_rate NUMERIC,
    average_latency NUMERIC,
    last_update TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        api_endpoint,
        COUNT(*) as request_count,
        AVG(success_rate) as success_rate,
        AVG(latency_ms) as average_latency,
        MAX(timestamp) as last_update
    FROM api_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    GROUP BY api_endpoint
    ORDER BY COUNT(*) DESC;
END;
$$ LANGUAGE plpgsql;

-- 11. Dashboard de Cache
CREATE OR REPLACE FUNCTION dashboard.cache_dashboard()
RETURNS TABLE (
    cache_type VARCHAR(100),
    hit_rate NUMERIC,
    miss_rate NUMERIC,
    average_latency NUMERIC,
    last_update TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cache_type,
        AVG(hit_rate) as hit_rate,
        AVG(miss_rate) as miss_rate,
        AVG(latency_ms) as average_latency,
        MAX(timestamp) as last_update
    FROM cache_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    GROUP BY cache_type
    ORDER BY AVG(hit_rate) DESC;
END;
$$ LANGUAGE plpgsql;

-- 12. Dashboard de Banco de Dados
CREATE OR REPLACE FUNCTION dashboard.database_dashboard()
RETURNS TABLE (
    operation_type VARCHAR(100),
    success_rate NUMERIC,
    average_latency NUMERIC,
    query_count INTEGER,
    last_update TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        operation_type,
        AVG(success_rate) as success_rate,
        AVG(latency_ms) as average_latency,
        COUNT(*) as query_count,
        MAX(timestamp) as last_update
    FROM database_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    GROUP BY operation_type
    ORDER BY AVG(latency_ms) DESC;
END;
$$ LANGUAGE plpgsql;

-- 13. Dashboard de Criptografia
CREATE OR REPLACE FUNCTION dashboard.encryption_dashboard()
RETURNS TABLE (
    algorithm VARCHAR(100),
    success_rate NUMERIC,
    average_latency NUMERIC,
    operation_count INTEGER,
    last_update TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        algorithm,
        AVG(success_rate) as success_rate,
        AVG(latency_ms) as average_latency,
        COUNT(*) as operation_count,
        MAX(timestamp) as last_update
    FROM encryption_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    GROUP BY algorithm
    ORDER BY AVG(latency_ms) DESC;
END;
$$ LANGUAGE plpgsql;

-- 14. Dashboard de Autenticação Contínua
CREATE OR REPLACE FUNCTION dashboard.continuous_auth_dashboard()
RETURNS TABLE (
    metric_name VARCHAR(100),
    current_value NUMERIC,
    threshold NUMERIC,
    status VARCHAR(20),
    trend VARCHAR(20),
    last_update TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        'Continuous Auth Success Rate' as metric_name,
        AVG(success_rate) as current_value,
        99 as threshold,
        CASE 
            WHEN AVG(success_rate) < 99 THEN 'CRITICAL'
            WHEN AVG(success_rate) < 99.5 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        CASE 
            WHEN AVG(success_rate) < COALESCE(LAG(AVG(success_rate)) OVER (ORDER BY timestamp), 0) THEN 'DOWN'
            ELSE 'UP'
        END as trend,
        MAX(timestamp) as last_update
    FROM continuous_auth_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    GROUP BY metric_name
    
    UNION ALL
    
    SELECT 
        'Continuous Auth Latency' as metric_name,
        AVG(latency_ms) as current_value,
        50 as threshold,
        CASE 
            WHEN AVG(latency_ms) > 50 THEN 'CRITICAL'
            WHEN AVG(latency_ms) > 30 THEN 'WARNING'
            ELSE 'OK'
        END as status,
        CASE 
            WHEN AVG(latency_ms) > COALESCE(LAG(AVG(latency_ms)) OVER (ORDER BY timestamp), 0) THEN 'UP'
            ELSE 'DOWN'
        END as trend,
        MAX(timestamp) as last_update
    FROM continuous_auth_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    GROUP BY metric_name;
END;
$$ LANGUAGE plpgsql;

-- 15. Dashboard de Integração
CREATE OR REPLACE FUNCTION dashboard.integration_dashboard()
RETURNS TABLE (
    integration_type VARCHAR(100),
    success_rate NUMERIC,
    failure_rate NUMERIC,
    average_latency NUMERIC,
    last_update TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        integration_type,
        AVG(success_rate) as success_rate,
        AVG(failure_rate) as failure_rate,
        AVG(latency_ms) as average_latency,
        MAX(timestamp) as last_update
    FROM integration_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
    GROUP BY integration_type
    ORDER BY AVG(latency_ms) DESC;
END;
$$ LANGUAGE plpgsql;

-- Exemplo de uso
-- SELECT * FROM dashboard.performance_dashboard();
-- SELECT * FROM dashboard.security_dashboard();
-- SELECT * FROM dashboard.compliance_dashboard();
