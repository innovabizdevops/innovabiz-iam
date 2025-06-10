-- INNOVABIZ - IAM Metrics and Statistics
-- Author: Eduardo Jeremias
-- Date: 19/05/2025
-- Version: 1.0
-- Description: Script para gerar métricas e estatísticas do módulo IAM

-- Configuração do ambiente
SET search_path TO iam, public;
SET statement_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

-- Criar tabela para armazenar métricas históricas
CREATE TABLE IF NOT EXISTS iam_metrics_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metric_name TEXT NOT NULL,
    metric_value NUMERIC NOT NULL,
    metric_details JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT uq_metric_date_name UNIQUE (metric_date, metric_name)
);

-- Criar índices para otimização de consultas
CREATE INDEX IF NOT EXISTS idx_iam_metrics_history_date ON iam_metrics_history(metric_date);
CREATE INDEX IF NOT EXISTS idx_iam_metrics_history_name ON iam_metrics_history(metric_name);

-- Comentários para documentação
COMMENT ON TABLE iam_metrics_history IS 'Armazena métricas históricas do módulo IAM para análise de tendências e monitoramento';
COMMENT ON COLUMN iam_metrics_history.metric_date IS 'Data e hora em que a métrica foi coletada';
COMMENT ON COLUMN iam_metrics_history.metric_name IS 'Nome da métrica (ex: total_users, active_sessions, etc.)';
COMMENT ON COLUMN iam_metrics_history.metric_value IS 'Valor numérico da métrica';
COMMENT ON COLUMN iam_metrics_history.metric_details IS 'Detalhes adicionais em formato JSON';

-- Função para coletar métricas do IAM
CREATE OR REPLACE FUNCTION collect_iam_metrics()
RETURNS JSONB AS $$
DECLARE
    metrics JSONB := '{}';
    temp_count BIGINT;
    temp_json JSONB;
    temp_record RECORD;
    temp_array TEXT[];
    
    -- Variáveis para métricas de usuários
    total_users BIGINT;
    active_users BIGINT;
    inactive_users BIGINT;
    locked_users BIGINT;
    users_by_status JSONB;
    users_by_department JSONB;
    
    -- Variáveis para métricas de sessões
    active_sessions BIGINT;
    avg_session_duration INTERVAL;
    sessions_by_client JSONB;
    
    -- Variáveis para métricas de auditoria
    audit_events_24h BIGINT;
    audit_events_by_type JSONB;
    failed_logins_24h BIGINT;
    
    -- Variáveis para métricas de funções e permissões
    total_roles BIGINT;
    avg_permissions_per_role NUMERIC;
    users_without_roles BIGINT;
    
    -- Variáveis para métricas de segurança
    password_age_days NUMERIC;
    mfa_adoption_rate NUMERIC;
    
    -- Variáveis para métricas de desempenho
    query_response_time_ms NUMERIC;
    
BEGIN
    -- 1. Métricas de Usuários
    -- Total de usuários
    SELECT COUNT(*) INTO total_users FROM users;
    metrics := jsonb_insert(metrics, '{user_metrics,total_users}', to_jsonb(total_users));
    
    -- Usuários por status
    SELECT 
        COUNT(*) FILTER (WHERE status = 'active') AS active,
        COUNT(*) FILTER (WHERE status = 'inactive') AS inactive,
        COUNT(*) FILTER (WHERE status = 'locked') AS locked
    INTO active_users, inactive_users, locked_users;
    
    metrics := jsonb_insert(metrics, '{user_metrics,active_users}', to_jsonb(active_users));
    metrics := jsonb_insert(metrics, '{user_metrics,inactive_users}', to_jsonb(inactive_users));
    metrics := jsonb_insert(metrics, '{user_metrics,locked_users}', to_jsonb(locked_users));
    
    -- Distribuição de usuários por departamento (se disponível nos metadados)
    SELECT 
        jsonb_object_agg(
            COALESCE(department, 'not_specified'), 
            user_count
        ) INTO users_by_department
    FROM (
        SELECT 
            metadata->>'department' AS department,
            COUNT(*) AS user_count
        FROM users
        WHERE status = 'active'
        GROUP BY metadata->>'department'
        ORDER BY user_count DESC
    ) dept_counts;
    
    metrics := jsonb_insert(metrics, '{user_metrics,users_by_department}', COALESCE(users_by_department, '{}'::jsonb));
    
    -- 2. Métricas de Sessões
    -- Sessões ativas
    SELECT COUNT(*) INTO active_sessions FROM sessions WHERE is_active = true;
    metrics := jsonb_insert(metrics, '{session_metrics,active_sessions}', to_jsonb(active_sessions));
    
    -- Duração média das sessões ativas
    SELECT AVG(NOW() - created_at) INTO avg_session_duration 
    FROM sessions 
    WHERE is_active = true 
    AND created_at > NOW() - INTERVAL '24 hours';
    
    metrics := jsonb_insert(metrics, '{session_metrics,avg_session_duration_seconds}', 
        to_jsonb(COALESCE(EXTRACT(EPOCH FROM avg_session_duration)::NUMERIC, 0)));
    
    -- Sessões por cliente/plataforma
    SELECT 
        jsonb_object_agg(
            COALESCE(user_agent, 'unknown'), 
            session_count
        ) INTO sessions_by_client
    FROM (
        SELECT 
            user_agent,
            COUNT(*) AS session_count
        FROM sessions
        WHERE is_active = true
        GROUP BY user_agent
        ORDER BY session_count DESC
        LIMIT 10
    ) client_counts;
    
    metrics := jsonb_insert(metrics, '{session_metrics,sessions_by_client}', COALESCE(sessions_by_client, '{}'::jsonb));
    
    -- 3. Métricas de Auditoria
    -- Eventos de auditoria nas últimas 24 horas
    SELECT COUNT(*) INTO audit_events_24h 
    FROM audit_logs 
    WHERE timestamp > NOW() - INTERVAL '24 hours';
    
    metrics := jsonb_insert(metrics, '{audit_metrics,audit_events_24h}', to_jsonb(audit_events_24h));
    
    -- Eventos de auditoria por tipo
    SELECT 
        jsonb_object_agg(
            action, 
            action_count
        ) INTO audit_events_by_type
    FROM (
        SELECT 
            action,
            COUNT(*) AS action_count
        FROM audit_logs
        WHERE timestamp > NOW() - INTERVAL '24 hours'
        GROUP BY action
        ORDER BY action_count DESC
        LIMIT 20
    ) action_counts;
    
    metrics := jsonb_insert(metrics, '{audit_metrics,audit_events_by_type}', COALESCE(audit_events_by_type, '{}'::jsonb));
    
    -- Tentativas de login malsucedidas
    SELECT COUNT(*) INTO failed_logins_24h
    FROM audit_logs 
    WHERE action = 'login_failed' 
    AND timestamp > NOW() - INTERVAL '24 hours';
    
    metrics := jsonb_insert(metrics, '{audit_metrics,failed_logins_24h}', to_jsonb(failed_logins_24h));
    
    -- 4. Métricas de Funções e Permissões
    -- Total de funções
    SELECT COUNT(*) INTO total_roles FROM roles;
    metrics := jsonb_insert(metrics, '{role_metrics,total_roles}', to_jsonb(total_roles));
    
    -- Média de permissões por função
    SELECT AVG(jsonb_array_length(permissions)) INTO avg_permissions_per_role FROM roles;
    metrics := jsonb_insert(metrics, '{role_metrics,avg_permissions_per_role}', 
        to_jsonb(COALESCE(ROUND(avg_permissions_per_role::NUMERIC, 2), 0)));
    
    -- Usuários sem funções atribuídas
    SELECT COUNT(*) INTO users_without_roles
    FROM users u
    LEFT JOIN user_roles ur ON u.id = ur.user_id
    WHERE ur.id IS NULL
    AND u.status = 'active';
    
    metrics := jsonb_insert(metrics, '{role_metrics,users_without_roles}', to_jsonb(users_without_roles));
    
    -- 5. Métricas de Segurança
    -- Idade média das senhas (em dias)
    SELECT AVG(EXTRACT(EPOCH FROM (NOW() - created_at)) / 86400) INTO password_age_days
    FROM users
    WHERE status = 'active';
    
    metrics := jsonb_insert(metrics, '{security_metrics,avg_password_age_days}', 
        to_jsonb(COALESCE(ROUND(password_age_days::NUMERIC, 2), 0)));
    
    -- Taxa de adoção de MFA (exemplo simplificado)
    SELECT 
        ROUND(COUNT(*) FILTER (WHERE metadata->'security'->>'mfa_enabled' = 'true') * 100.0 / 
              NULLIF(COUNT(*), 0), 2) INTO mfa_adoption_rate
    FROM users
    WHERE status = 'active';
    
    metrics := jsonb_insert(metrics, '{security_metrics,mfa_adoption_rate}', 
        to_jsonb(COALESCE(mfa_adoption_rate, 0)));
    
    -- 6. Métricas de Desempenho
    -- Tempo médio de resposta de consultas (exemplo simplificado)
    SELECT AVG(EXTRACT(MILLISECONDS FROM (NOW() - timestamp))) INTO query_response_time_ms
    FROM audit_logs
    WHERE action = 'query_executed'
    AND timestamp > NOW() - INTERVAL '1 hour';
    
    metrics := jsonb_insert(metrics, '{performance_metrics,avg_query_response_time_ms}', 
        to_jsonb(COALESCE(ROUND(query_response_time_ms::NUMERIC, 2), 0)));
    
    -- 7. Armazenar métricas históricas
    -- Inserir métricas agregadas na tabela de histórico
    INSERT INTO iam_metrics_history (metric_name, metric_value, metric_details)
    SELECT 
        'total_users', 
        total_users, 
        jsonb_build_object('active', active_users, 'inactive', inactive_users, 'locked', locked_users)
    
    UNION ALL
    
    SELECT 
        'active_sessions', 
        active_sessions, 
        sessions_by_client
    
    UNION ALL
    
    SELECT 
        'audit_events_24h', 
        audit_events_24h, 
        jsonb_build_object('by_type', audit_events_by_type, 'failed_logins', failed_logins_24h)
    
    UNION ALL
    
    SELECT 
        'role_metrics', 
        total_roles, 
        jsonb_build_object(
            'avg_permissions', ROUND(COALESCE(avg_permissions_per_role, 0)::NUMERIC, 2),
            'users_without_roles', users_without_roles
        )
    
    UNION ALL
    
    SELECT 
        'security_metrics', 
        ROUND(COALESCE(mfa_adoption_rate, 0)::NUMERIC, 2), 
        jsonb_build_object(
            'avg_password_age_days', ROUND(COALESCE(password_age_days, 0)::NUMERIC, 2),
            'mfa_adoption_rate', ROUND(COALESCE(mfa_adoption_rate, 0)::NUMERIC, 2)
        );
    
    -- Retornar todas as métricas
    RETURN jsonb_build_object(
        'timestamp', NOW(),
        'metrics', metrics
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para obter métricas de tendência ao longo do tempo
CREATE OR REPLACE FUNCTION get_metric_trend(
    p_metric_name TEXT,
    p_days INT DEFAULT 30,
    p_interval TEXT DEFAULT 'day'
)
RETURNS TABLE (
    metric_date TIMESTAMP WITH TIME ZONE,
    metric_value NUMERIC,
    metric_details JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        date_trunc(p_interval, metric_date) AS metric_date,
        AVG(metric_value) AS metric_value,
        -- Para métricas com detalhes, podemos agregar de forma simples
        -- Em um cenário real, pode ser necessário um tratamento mais sofisticado
        jsonb_agg(metric_details) AS metric_details
    FROM iam_metrics_history
    WHERE metric_name = p_metric_name
    AND metric_date >= (NOW() - (p_days || ' days')::INTERVAL)
    GROUP BY date_trunc(p_interval, metric_date)
    ORDER BY metric_date;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Criar visualização para dashboard
CREATE OR REPLACE VIEW vw_iam_dashboard AS
WITH 
user_metrics AS (
    SELECT 
        'user_metrics' AS metric_category,
        jsonb_build_object(
            'total_users', (SELECT COUNT(*) FROM users),
            'active_users', (SELECT COUNT(*) FROM users WHERE status = 'active'),
            'inactive_users', (SELECT COUNT(*) FROM users WHERE status = 'inactive'),
            'locked_users', (SELECT COUNT(*) FROM users WHERE status = 'locked'),
            'users_added_last_30d', (SELECT COUNT(*) FROM users WHERE created_at >= NOW() - INTERVAL '30 days')
        ) AS metrics
),
session_metrics AS (
    SELECT 
        'session_metrics' AS metric_category,
        jsonb_build_object(
            'active_sessions', (SELECT COUNT(*) FROM sessions WHERE is_active = true),
            'sessions_last_24h', (SELECT COUNT(*) FROM sessions WHERE created_at >= NOW() - INTERVAL '24 hours'),
            'avg_session_duration_min', (
                SELECT COALESCE(EXTRACT(EPOCH FROM AVG(NOW() - created_at)) / 60, 0)
                FROM sessions 
                WHERE is_active = true
            ),
            'unique_ips_last_24h', (SELECT COUNT(DISTINCT ip_address) FROM sessions WHERE created_at >= NOW() - INTERVAL '24 hours')
        ) AS metrics
),
audit_metrics AS (
    SELECT 
        'audit_metrics' AS metric_category,
        jsonb_build_object(
            'audit_events_24h', (SELECT COUNT(*) FROM audit_logs WHERE timestamp >= NOW() - INTERVAL '24 hours'),
            'failed_logins_24h', (SELECT COUNT(*) FROM audit_logs WHERE action = 'login_failed' AND timestamp >= NOW() - INTERVAL '24 hours'),
            'password_changes_24h', (SELECT COUNT(*) FROM audit_logs WHERE action = 'password_change' AND timestamp >= NOW() - INTERVAL '24 hours'),
            'permission_changes_24h', (SELECT COUNT(*) FROM audit_logs WHERE action LIKE '%permission%' AND timestamp >= NOW() - INTERVAL '24 hours')
        ) AS metrics
),
role_metrics AS (
    SELECT 
        'role_metrics' AS metric_category,
        jsonb_build_object(
            'total_roles', (SELECT COUNT(*) FROM roles),
            'avg_permissions_per_role', (SELECT COALESCE(AVG(jsonb_array_length(permissions)), 0) FROM roles),
            'users_without_roles', (
                SELECT COUNT(*) 
                FROM users u
                LEFT JOIN user_roles ur ON u.id = ur.user_id
                WHERE ur.id IS NULL
                AND u.status = 'active'
            ),
            'most_used_roles', (
                SELECT jsonb_agg(jsonb_build_object('role_name', r.name, 'user_count', role_counts.user_count))
                FROM (
                    SELECT 
                        ur.role_id,
                        COUNT(*) AS user_count
                    FROM user_roles ur
                    JOIN users u ON ur.user_id = u.id
                    WHERE u.status = 'active'
                    GROUP BY ur.role_id
                    ORDER BY user_count DESC
                    LIMIT 5
                ) role_counts
                JOIN roles r ON role_counts.role_id = r.id
            )
        ) AS metrics
)
SELECT * FROM user_metrics
UNION ALL SELECT * FROM session_metrics
UNION ALL SELECT * FROM audit_metrics
UNION ALL SELECT * FROM role_metrics;

-- Comentários para documentação
COMMENT ON FUNCTION collect_iam_metrics() IS 'Coleta e retorna métricas abrangentes do módulo IAM, incluindo usuários, sessões, auditorias, funções e segurança';
COMMENT ON FUNCTION get_metric_trend(TEXT, INT, TEXT) IS 'Obtém tendências históricas para uma métrica específica ao longo do tempo';
COMMENT ON VIEW vw_iam_dashboard IS 'Visão agregada das principais métricas do IAM para uso em dashboards';

-- Criar função para agendamento de coleta de métricas
CREATE OR REPLACE FUNCTION schedule_iam_metrics_collection()
RETURNS VOID AS $$
BEGIN
    -- Esta função seria chamada por um agendador externo (como pg_cron)
    -- Exemplo para pg_cron: SELECT cron.schedule('0 * * * *', 'SELECT collect_iam_metrics()');
    -- Esta é apenas uma função de espaço reservado
    RAISE NOTICE 'Esta função deve ser configurada com um agendador externo como pg_cron';
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Registrar a criação deste script no log de auditoria
INSERT INTO audit_logs (
    id,
    organization_id,
    user_id,
    action,
    resource_type,
    resource_id,
    status,
    details
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111', -- ID da organização padrão
    '22222222-2222-2222-2222-222222222222', -- ID do usuário admin
    'execute',
    'schema',
    'iam',
    'success',
    '{"script": "04_generate_iam_metrics.sql", "version": "1.0", "description": "Criação de funções para coleta e análise de métricas do IAM"}'
);

-- Mensagem de conclusão
DO $$
BEGIN
    RAISE NOTICE 'Script de geração de métricas do IAM criado com sucesso!';
    RAISE NOTICE 'Para visualizar o dashboard, execute: SELECT * FROM vw_iam_dashboard;';
    RAISE NOTICE 'Para coletar métricas manualmente, execute: SELECT * FROM collect_iam_metrics();';
END
$$;
