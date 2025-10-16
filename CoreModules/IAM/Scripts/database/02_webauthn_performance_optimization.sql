-- ============================================================================
-- INNOVABIZ IAM - WebAuthn Performance Optimization
-- ============================================================================
-- Version: 1.0.0
-- Date: 31/07/2025
-- Author: Equipe de Performance INNOVABIZ
-- Classification: Confidencial - Interno
-- 
-- Description: Otimizações avançadas de performance para WebAuthn/FIDO2
-- Target: Suporte a 1M+ credenciais, 10K+ auth/sec, latência <100ms P95
-- ============================================================================

-- ============================================================================
-- ÍNDICES ESPECIALIZADOS PARA ALTA PERFORMANCE
-- ============================================================================

-- Índice composto para consultas de autenticação mais frequentes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webauthn_credentials_auth_lookup 
ON webauthn_credentials (credential_id, status, tenant_id) 
WHERE status = 'active';

-- Índice para consultas de usuário com ordenação por uso recente
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webauthn_credentials_user_recent 
ON webauthn_credentials (user_id, tenant_id, last_used_at DESC NULLS LAST) 
WHERE status = 'active';

-- Índice parcial para credenciais de alto risco
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webauthn_credentials_high_risk 
ON webauthn_credentials (tenant_id, risk_score DESC, last_used_at DESC) 
WHERE status = 'active' AND risk_score > 0.7;

-- Índice para análise de compliance por AAL
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webauthn_credentials_compliance 
ON webauthn_credentials (tenant_id, compliance_level, authenticator_type, created_at DESC) 
WHERE status = 'active';

-- Índice GIN para busca em metadados JSONB
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webauthn_credentials_metadata_gin 
ON webauthn_credentials USING GIN (metadata);

-- Índice para detecção de anomalias de sign count
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webauthn_credentials_sign_count 
ON webauthn_credentials (sign_count, last_used_at DESC) 
WHERE status = 'active' AND sign_count > 0;

-- ============================================================================
-- ÍNDICES PARA EVENTOS DE AUTENTICAÇÃO
-- ============================================================================

-- Índice composto para análise de segurança
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webauthn_events_security_analysis 
ON webauthn_authentication_events (tenant_id, event_type, result, created_at DESC);

-- Índice para análise de risco por IP
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webauthn_events_ip_risk 
ON webauthn_authentication_events (ip_address, risk_score DESC, created_at DESC) 
WHERE risk_score > 0.5;

-- Índice para auditoria de falhas
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webauthn_events_failures 
ON webauthn_authentication_events (tenant_id, user_id, result, created_at DESC) 
WHERE result = 'failure';

-- Índice GIN para busca em fatores de risco
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webauthn_events_risk_factors_gin 
ON webauthn_authentication_events USING GIN (risk_factors);

-- Índice para compliance e auditoria
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webauthn_events_compliance 
ON webauthn_authentication_events (tenant_id, compliance_level, created_at DESC) 
WHERE compliance_level IS NOT NULL;

-- ============================================================================
-- VIEWS MATERIALIZADAS PARA ANALYTICS
-- ============================================================================

-- View materializada para estatísticas de credenciais por tenant
CREATE MATERIALIZED VIEW mv_webauthn_credentials_stats AS
SELECT 
    tenant_id,
    COUNT(*) as total_credentials,
    COUNT(*) FILTER (WHERE status = 'active') as active_credentials,
    COUNT(*) FILTER (WHERE status = 'suspended') as suspended_credentials,
    COUNT(*) FILTER (WHERE status = 'revoked') as revoked_credentials,
    COUNT(*) FILTER (WHERE authenticator_type = 'platform') as platform_authenticators,
    COUNT(*) FILTER (WHERE authenticator_type = 'cross-platform') as cross_platform_authenticators,
    COUNT(*) FILTER (WHERE compliance_level = 'AAL2') as aal2_credentials,
    COUNT(*) FILTER (WHERE compliance_level = 'AAL3') as aal3_credentials,
    COUNT(*) FILTER (WHERE last_used_at > NOW() - INTERVAL '30 days') as active_last_30_days,
    COUNT(*) FILTER (WHERE last_used_at > NOW() - INTERVAL '7 days') as active_last_7_days,
    COUNT(*) FILTER (WHERE last_used_at > NOW() - INTERVAL '1 day') as active_last_24_hours,
    AVG(risk_score) as avg_risk_score,
    MAX(last_used_at) as last_activity,
    MIN(created_at) as first_credential_date,
    MAX(created_at) as last_credential_date
FROM webauthn_credentials
GROUP BY tenant_id;

-- Índice único para a view materializada
CREATE UNIQUE INDEX ON mv_webauthn_credentials_stats (tenant_id);

-- View materializada para análise de eventos por hora
CREATE MATERIALIZED VIEW mv_webauthn_events_hourly AS
SELECT 
    tenant_id,
    date_trunc('hour', created_at) as hour_bucket,
    event_type,
    result,
    COUNT(*) as event_count,
    COUNT(DISTINCT user_id) as unique_users,
    COUNT(DISTINCT credential_id) as unique_credentials,
    COUNT(DISTINCT ip_address) as unique_ips,
    AVG(risk_score) as avg_risk_score,
    MAX(risk_score) as max_risk_score,
    COUNT(*) FILTER (WHERE risk_score > 0.7) as high_risk_events,
    COUNT(*) FILTER (WHERE user_verified = true) as user_verified_events
FROM webauthn_authentication_events
WHERE created_at >= NOW() - INTERVAL '7 days'
GROUP BY tenant_id, date_trunc('hour', created_at), event_type, result;

-- Índices para a view de eventos
CREATE INDEX ON mv_webauthn_events_hourly (tenant_id, hour_bucket DESC);
CREATE INDEX ON mv_webauthn_events_hourly (hour_bucket DESC, event_type, result);

-- ============================================================================
-- FUNÇÕES OTIMIZADAS PARA CONSULTAS FREQUENTES
-- ============================================================================

-- Função otimizada para busca rápida de credencial por ID
CREATE OR REPLACE FUNCTION get_credential_by_id(p_credential_id TEXT)
RETURNS TABLE (
    id UUID,
    user_id UUID,
    tenant_id UUID,
    public_key BYTEA,
    sign_count BIGINT,
    authenticator_type authenticator_type,
    user_verified BOOLEAN,
    backup_eligible BOOLEAN,
    backup_state BOOLEAN,
    transports TEXT[],
    status credential_status,
    compliance_level authentication_assurance_level,
    last_used_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        c.id, c.user_id, c.tenant_id, c.public_key, c.sign_count,
        c.authenticator_type, c.user_verified, c.backup_eligible, c.backup_state,
        c.transports, c.status, c.compliance_level, c.last_used_at
    FROM webauthn_credentials c
    WHERE c.credential_id = p_credential_id 
    AND c.status = 'active';
END;
$$ LANGUAGE plpgsql STABLE;

-- Função para buscar credenciais ativas de um usuário
CREATE OR REPLACE FUNCTION get_user_active_credentials(p_user_id UUID, p_tenant_id UUID)
RETURNS TABLE (
    credential_id TEXT,
    authenticator_type authenticator_type,
    transports TEXT[],
    friendly_name TEXT,
    last_used_at TIMESTAMP WITH TIME ZONE,
    compliance_level authentication_assurance_level
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        c.credential_id, c.authenticator_type, c.transports, 
        c.friendly_name, c.last_used_at, c.compliance_level
    FROM webauthn_credentials c
    WHERE c.user_id = p_user_id 
    AND c.tenant_id = p_tenant_id 
    AND c.status = 'active'
    ORDER BY c.last_used_at DESC NULLS LAST, c.created_at DESC;
END;
$$ LANGUAGE plpgsql STABLE;

-- Função para atualização otimizada de uso de credencial
CREATE OR REPLACE FUNCTION update_credential_usage(
    p_credential_id TEXT,
    p_new_sign_count BIGINT,
    p_ip_address INET DEFAULT NULL,
    p_user_agent TEXT DEFAULT NULL,
    p_risk_score DECIMAL(3,2) DEFAULT NULL
) RETURNS BOOLEAN AS $$
DECLARE
    credential_record RECORD;
    sign_count_anomaly BOOLEAN := false;
BEGIN
    -- Buscar credencial atual com lock
    SELECT id, user_id, tenant_id, sign_count, status
    INTO credential_record
    FROM webauthn_credentials
    WHERE credential_id = p_credential_id
    FOR UPDATE;
    
    -- Verificar se credencial existe e está ativa
    IF NOT FOUND OR credential_record.status != 'active' THEN
        RETURN false;
    END IF;
    
    -- Verificar anomalia de sign count
    IF p_new_sign_count <= credential_record.sign_count AND credential_record.sign_count > 0 THEN
        sign_count_anomaly := true;
        
        -- Suspender credencial
        UPDATE webauthn_credentials 
        SET status = 'suspended',
            suspension_reason = 'SIGN_COUNT_ANOMALY',
            updated_at = NOW()
        WHERE id = credential_record.id;
        
        -- Registrar evento de anomalia
        INSERT INTO webauthn_authentication_events (
            credential_id, user_id, tenant_id, event_type, result,
            sign_count, ip_address, user_agent, risk_score,
            error_code, error_message, metadata
        ) VALUES (
            credential_record.id, credential_record.user_id, credential_record.tenant_id,
            'sign_count_anomaly', 'warning',
            p_new_sign_count, p_ip_address, p_user_agent, COALESCE(p_risk_score, 1.00),
            'SIGN_COUNT_ANOMALY',
            'Sign count anomaly detected - possible credential cloning',
            jsonb_build_object(
                'expected_count', credential_record.sign_count,
                'received_count', p_new_sign_count,
                'anomaly_detected_at', NOW()
            )
        );
        
        RETURN false;
    END IF;
    
    -- Atualizar credencial normalmente
    UPDATE webauthn_credentials 
    SET sign_count = p_new_sign_count,
        last_used_at = NOW(),
        risk_score = COALESCE(p_risk_score, risk_score),
        updated_at = NOW()
    WHERE id = credential_record.id;
    
    RETURN true;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- FUNÇÕES DE ANÁLISE E RELATÓRIOS
-- ============================================================================

-- Função para análise de tendências de autenticação
CREATE OR REPLACE FUNCTION analyze_authentication_trends(
    p_tenant_id UUID,
    p_hours_back INTEGER DEFAULT 24
) RETURNS JSONB AS $$
DECLARE
    result JSONB;
    current_hour_count INTEGER;
    previous_hour_count INTEGER;
    success_rate DECIMAL(5,2);
    unique_users INTEGER;
    high_risk_percentage DECIMAL(5,2);
BEGIN
    -- Estatísticas da última hora
    SELECT COUNT(*) INTO current_hour_count
    FROM webauthn_authentication_events
    WHERE tenant_id = p_tenant_id
    AND created_at >= date_trunc('hour', NOW())
    AND event_type = 'authentication';
    
    -- Estatísticas da hora anterior
    SELECT COUNT(*) INTO previous_hour_count
    FROM webauthn_authentication_events
    WHERE tenant_id = p_tenant_id
    AND created_at >= date_trunc('hour', NOW()) - INTERVAL '1 hour'
    AND created_at < date_trunc('hour', NOW())
    AND event_type = 'authentication';
    
    -- Taxa de sucesso nas últimas 24 horas
    SELECT 
        ROUND(
            (COUNT(*) FILTER (WHERE result = 'success')::DECIMAL / 
             NULLIF(COUNT(*), 0) * 100), 2
        ) INTO success_rate
    FROM webauthn_authentication_events
    WHERE tenant_id = p_tenant_id
    AND created_at >= NOW() - INTERVAL '24 hours'
    AND event_type = 'authentication';
    
    -- Usuários únicos nas últimas 24 horas
    SELECT COUNT(DISTINCT user_id) INTO unique_users
    FROM webauthn_authentication_events
    WHERE tenant_id = p_tenant_id
    AND created_at >= NOW() - INTERVAL '24 hours'
    AND event_type = 'authentication'
    AND result = 'success';
    
    -- Percentual de eventos de alto risco
    SELECT 
        ROUND(
            (COUNT(*) FILTER (WHERE risk_score > 0.7)::DECIMAL / 
             NULLIF(COUNT(*), 0) * 100), 2
        ) INTO high_risk_percentage
    FROM webauthn_authentication_events
    WHERE tenant_id = p_tenant_id
    AND created_at >= NOW() - MAKE_INTERVAL(hours => p_hours_back)
    AND event_type = 'authentication';
    
    -- Construir resultado
    SELECT jsonb_build_object(
        'tenant_id', p_tenant_id,
        'analysis_period_hours', p_hours_back,
        'current_hour_authentications', current_hour_count,
        'previous_hour_authentications', previous_hour_count,
        'hourly_change_percentage', 
            CASE 
                WHEN previous_hour_count > 0 THEN 
                    ROUND(((current_hour_count - previous_hour_count)::DECIMAL / previous_hour_count * 100), 2)
                ELSE NULL 
            END,
        'success_rate_24h', COALESCE(success_rate, 0),
        'unique_users_24h', unique_users,
        'high_risk_percentage', COALESCE(high_risk_percentage, 0),
        'analysis_timestamp', NOW()
    ) INTO result;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql STABLE;

-- Função para detecção de padrões suspeitos
CREATE OR REPLACE FUNCTION detect_suspicious_patterns(
    p_tenant_id UUID,
    p_lookback_hours INTEGER DEFAULT 1
) RETURNS JSONB AS $$
DECLARE
    result JSONB;
    suspicious_ips JSONB;
    credential_anomalies JSONB;
    unusual_locations JSONB;
BEGIN
    -- IPs com muitas falhas
    SELECT jsonb_agg(
        jsonb_build_object(
            'ip_address', ip_address,
            'failure_count', failure_count,
            'unique_users', unique_users
        )
    ) INTO suspicious_ips
    FROM (
        SELECT 
            ip_address,
            COUNT(*) as failure_count,
            COUNT(DISTINCT user_id) as unique_users
        FROM webauthn_authentication_events
        WHERE tenant_id = p_tenant_id
        AND created_at >= NOW() - MAKE_INTERVAL(hours => p_lookback_hours)
        AND result = 'failure'
        AND ip_address IS NOT NULL
        GROUP BY ip_address
        HAVING COUNT(*) >= 10 OR COUNT(DISTINCT user_id) >= 5
        ORDER BY failure_count DESC
        LIMIT 10
    ) suspicious;
    
    -- Credenciais com anomalias de sign count
    SELECT jsonb_agg(
        jsonb_build_object(
            'credential_id', credential_id,
            'user_id', user_id,
            'anomaly_count', anomaly_count,
            'last_anomaly', last_anomaly
        )
    ) INTO credential_anomalies
    FROM (
        SELECT 
            credential_id,
            user_id,
            COUNT(*) as anomaly_count,
            MAX(created_at) as last_anomaly
        FROM webauthn_authentication_events
        WHERE tenant_id = p_tenant_id
        AND created_at >= NOW() - MAKE_INTERVAL(hours => p_lookback_hours)
        AND event_type = 'sign_count_anomaly'
        GROUP BY credential_id, user_id
        ORDER BY anomaly_count DESC
        LIMIT 10
    ) anomalies;
    
    -- Construir resultado
    SELECT jsonb_build_object(
        'tenant_id', p_tenant_id,
        'analysis_period_hours', p_lookback_hours,
        'suspicious_ips', COALESCE(suspicious_ips, '[]'::jsonb),
        'credential_anomalies', COALESCE(credential_anomalies, '[]'::jsonb),
        'analysis_timestamp', NOW()
    ) INTO result;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql STABLE;

-- ============================================================================
-- JOBS DE MANUTENÇÃO AUTOMATIZADA
-- ============================================================================

-- Função para limpeza automática de dados antigos
CREATE OR REPLACE FUNCTION maintenance_cleanup_old_data()
RETURNS JSONB AS $$
DECLARE
    deleted_challenges INTEGER;
    deleted_events INTEGER;
    result JSONB;
BEGIN
    -- Limpar challenges expirados há mais de 1 hora
    DELETE FROM webauthn_challenges 
    WHERE expires_at < NOW() - INTERVAL '1 hour';
    GET DIAGNOSTICS deleted_challenges = ROW_COUNT;
    
    -- Limpar eventos antigos (manter apenas 90 dias)
    DELETE FROM webauthn_authentication_events 
    WHERE created_at < NOW() - INTERVAL '90 days';
    GET DIAGNOSTICS deleted_events = ROW_COUNT;
    
    -- Atualizar estatísticas das tabelas
    ANALYZE webauthn_credentials;
    ANALYZE webauthn_authentication_events;
    ANALYZE webauthn_challenges;
    
    -- Refresh das views materializadas
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_webauthn_credentials_stats;
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_webauthn_events_hourly;
    
    SELECT jsonb_build_object(
        'deleted_challenges', deleted_challenges,
        'deleted_events', deleted_events,
        'maintenance_completed_at', NOW()
    ) INTO result;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- CONFIGURAÇÕES AVANÇADAS DE PERFORMANCE
-- ============================================================================

-- Configurar work_mem para consultas complexas
SET work_mem = '256MB';

-- Configurar shared_buffers para cache
-- (Deve ser configurado no postgresql.conf)
-- shared_buffers = 25% da RAM disponível

-- Configurar effective_cache_size
-- (Deve ser configurado no postgresql.conf)
-- effective_cache_size = 75% da RAM disponível

-- Configurações específicas para as tabelas
ALTER TABLE webauthn_credentials SET (
    fillfactor = 90,  -- Deixar espaço para updates
    autovacuum_vacuum_scale_factor = 0.1,
    autovacuum_analyze_scale_factor = 0.05
);

ALTER TABLE webauthn_authentication_events SET (
    fillfactor = 100, -- Tabela append-only
    autovacuum_vacuum_scale_factor = 0.2,
    autovacuum_analyze_scale_factor = 0.1
);

-- ============================================================================
-- MONITORAMENTO DE PERFORMANCE
-- ============================================================================

-- View para monitorar performance de consultas
CREATE OR REPLACE VIEW v_webauthn_query_performance AS
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    stddev_time,
    rows,
    100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) AS hit_percent
FROM pg_stat_statements 
WHERE query LIKE '%webauthn%'
ORDER BY total_time DESC;

-- View para monitorar uso de índices
CREATE OR REPLACE VIEW v_webauthn_index_usage AS
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_tup_read,
    idx_tup_fetch,
    idx_scan,
    CASE 
        WHEN idx_scan = 0 THEN 'UNUSED'
        WHEN idx_scan < 10 THEN 'LOW_USAGE'
        ELSE 'ACTIVE'
    END as usage_status
FROM pg_stat_user_indexes 
WHERE tablename LIKE 'webauthn%'
ORDER BY idx_scan DESC;

-- ============================================================================
-- GRANTS PARA FUNÇÕES DE PERFORMANCE
-- ============================================================================

-- Permissões para aplicação
GRANT EXECUTE ON FUNCTION get_credential_by_id(TEXT) TO webauthn_app;
GRANT EXECUTE ON FUNCTION get_user_active_credentials(UUID, UUID) TO webauthn_app;
GRANT EXECUTE ON FUNCTION update_credential_usage(TEXT, BIGINT, INET, TEXT, DECIMAL) TO webauthn_app;
GRANT EXECUTE ON FUNCTION analyze_authentication_trends(UUID, INTEGER) TO webauthn_app;
GRANT EXECUTE ON FUNCTION detect_suspicious_patterns(UUID, INTEGER) TO webauthn_app;

-- Permissões para views materializadas
GRANT SELECT ON mv_webauthn_credentials_stats TO webauthn_app;
GRANT SELECT ON mv_webauthn_events_hourly TO webauthn_app;

-- Permissões para monitoramento
GRANT SELECT ON v_webauthn_query_performance TO webauthn_app;
GRANT SELECT ON v_webauthn_index_usage TO webauthn_app;

-- ============================================================================
-- LOG DE OTIMIZAÇÕES APLICADAS
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'INNOVABIZ WebAuthn Performance Optimization v1.0.0 aplicado com sucesso em %', NOW();
    RAISE NOTICE 'Otimizações aplicadas:';
    RAISE NOTICE '- Índices especializados para consultas de alta frequência';
    RAISE NOTICE '- Views materializadas para analytics em tempo real';
    RAISE NOTICE '- Funções otimizadas para operações críticas';
    RAISE NOTICE '- Configurações de autovacuum ajustadas';
    RAISE NOTICE '- Sistema de monitoramento de performance implementado';
    RAISE NOTICE 'Target: 1M+ credenciais, 10K+ auth/sec, <100ms P95 latência';
END $$;

-- ============================================================================
-- FIM DO SCRIPT DE OTIMIZAÇÃO
-- ============================================================================