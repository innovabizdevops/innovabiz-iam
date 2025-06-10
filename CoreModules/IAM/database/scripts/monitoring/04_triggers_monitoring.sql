-- INNOVABIZ - IAM Monitoring Triggers
-- Author: Eduardo Jeremias
-- Date: 08/05/2025
-- Version: 1.0
-- Description: Triggers para o módulo de monitoramento de banco de dados.

-- Configurar caminho de busca
SET search_path TO iam_monitoring, iam, public;

-- ===============================================================================
-- TRIGGERS PARA LIMPEZA AUTOMÁTICA DE DADOS ANTIGOS
-- ===============================================================================

-- Função para limpar automaticamente dados antigos de métricas de consultas
CREATE OR REPLACE FUNCTION fn_auto_cleanup_query_metrics()
RETURNS TRIGGER AS $$
DECLARE
    v_retention_days INTEGER;
BEGIN
    -- Obter configuração de retenção
    SELECT retention_days INTO v_retention_days
    FROM metrics_collection_schedule
    WHERE collector_type = 'query_metrics'
    LIMIT 1;
    
    -- Se não houver configuração específica, usar valor padrão (90 dias)
    IF v_retention_days IS NULL THEN
        v_retention_days := 90;
    END IF;
    
    -- Remover dados antigos
    DELETE FROM query_metrics
    WHERE timestamp < NOW() - (v_retention_days || ' days')::INTERVAL;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger para limpar dados antigos após inserção de novas métricas
DROP TRIGGER IF EXISTS trg_auto_cleanup_query_metrics ON query_metrics;
CREATE TRIGGER trg_auto_cleanup_query_metrics
AFTER INSERT ON query_metrics
FOR EACH STATEMENT EXECUTE FUNCTION fn_auto_cleanup_query_metrics();

-- ===============================================================================
-- TRIGGERS PARA DETECÇÃO AUTOMÁTICA DE CONSULTAS LENTAS
-- ===============================================================================

-- Função para registrar automaticamente consultas lentas
CREATE OR REPLACE FUNCTION fn_detect_slow_queries()
RETURNS TRIGGER AS $$
DECLARE
    v_slow_threshold FLOAT;
    v_recommendations JSONB := '[]'::JSONB;
    v_query_pattern VARCHAR;
    v_contains_full_scan BOOLEAN;
    v_table_scan_percentage FLOAT;
BEGIN
    -- Obter limiar configurado para consultas lentas (em ms)
    SELECT COALESCE(CAST(value AS FLOAT), 1000) INTO v_slow_threshold
    FROM monitoring_settings
    WHERE name = 'slow_query_threshold_ms'
    LIMIT 1;
    
    -- Se não houver configuração específica, usar valor padrão (1000ms)
    IF v_slow_threshold IS NULL THEN
        v_slow_threshold := 1000.0;
    END IF;
    
    -- Verificar se a consulta está acima do limiar de lentidão
    IF NEW.duration_ms > v_slow_threshold THEN
        -- Analisar a consulta para gerar recomendações automáticas
        
        -- Verificar se contém varredura de tabela completa
        v_contains_full_scan := NEW.table_scans > 0;
        
        -- Calcular percentual de varreduras de tabela vs. índice
        IF (NEW.table_scans + NEW.index_scans) > 0 THEN
            v_table_scan_percentage := (NEW.table_scans::FLOAT / (NEW.table_scans + NEW.index_scans)) * 100;
        ELSE
            v_table_scan_percentage := 0;
        END IF;
        
        -- Gerar recomendações básicas
        IF v_contains_full_scan AND v_table_scan_percentage > 50 THEN
            v_recommendations := v_recommendations || jsonb_build_object(
                'type', 'index',
                'recommendation', 'Considere adicionar índices apropriados para reduzir varreduras completas de tabela',
                'priority', 'high'
            );
        END IF;
        
        -- Verificar por padrões específicos na consulta
        v_query_pattern := NEW.normalized_query_text;
        
        -- Verificar junções sem índices
        IF v_query_pattern ILIKE '%JOIN%' AND v_table_scan_percentage > 30 THEN
            v_recommendations := v_recommendations || jsonb_build_object(
                'type', 'join_optimization',
                'recommendation', 'Otimize as junções adicionando índices nas colunas de junção',
                'priority', 'medium'
            );
        END IF;
        
        -- Verificar agregações pesadas
        IF v_query_pattern ILIKE '%GROUP BY%' AND NEW.duration_ms > v_slow_threshold * 2 THEN
            v_recommendations := v_recommendations || jsonb_build_object(
                'type', 'aggregation',
                'recommendation', 'Otimize operações de agregação ou considere índices compostos incluindo as colunas de agrupamento',
                'priority', 'medium'
            );
        END IF;
        
        -- Verificar consultas com subqueries
        IF v_query_pattern ILIKE '%IN (SELECT%' OR v_query_pattern ILIKE '%EXISTS (SELECT%' THEN
            v_recommendations := v_recommendations || jsonb_build_object(
                'type', 'subquery',
                'recommendation', 'Considere reformular subconsultas ou utilizar junções quando possível',
                'priority', 'medium'
            );
        END IF;
        
        -- Verificar uso de funções em cláusulas WHERE
        IF v_query_pattern ~* 'WHERE\s+[^=]*\([^)]*\)\s*=' THEN
            v_recommendations := v_recommendations || jsonb_build_object(
                'type', 'function_filter',
                'recommendation', 'Evite usar funções em colunas filtradas na cláusula WHERE, pois isso impede o uso de índices',
                'priority', 'high'
            );
        END IF;
        
        -- Se ainda não tiver recomendações, adicionar uma genérica
        IF jsonb_array_length(v_recommendations) = 0 THEN
            v_recommendations := v_recommendations || jsonb_build_object(
                'type', 'general',
                'recommendation', 'Revise a estrutura da consulta e considere otimizações como índices adicionais, particionamento ou reescrita da consulta',
                'priority', 'medium'
            );
        END IF;
        
        -- Registrar a consulta lenta com recomendações
        INSERT INTO slow_queries (
            query_metrics_id,
            recommendations,
            detection_timestamp,
            resolution_status
        ) VALUES (
            NEW.id,
            v_recommendations,
            NOW(),
            'open'
        );
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger para detectar e registrar consultas lentas
DROP TRIGGER IF EXISTS trg_detect_slow_queries ON query_metrics;
CREATE TRIGGER trg_detect_slow_queries
AFTER INSERT ON query_metrics
FOR EACH ROW EXECUTE FUNCTION fn_detect_slow_queries();

-- ===============================================================================
-- TRIGGERS PARA ATUALIZAÇÃO AUTOMÁTICA DE ESTATÍSTICAS DE TABELAS E ÍNDICES
-- ===============================================================================

-- Função para agendar atualização de estatísticas após muitas alterações
CREATE OR REPLACE FUNCTION fn_schedule_stats_update()
RETURNS TRIGGER AS $$
DECLARE
    v_table_name TEXT;
    v_schema_name TEXT;
    v_last_updated TIMESTAMP;
    v_changes_since_last_update BIGINT;
    v_threshold BIGINT := 10000; -- Limiar padrão de alterações antes de atualizar estatísticas
BEGIN
    -- Extrair nome do schema e tabela da consulta
    IF NEW.query_type IN ('INSERT', 'UPDATE', 'DELETE') AND NEW.normalized_query_text IS NOT NULL THEN
        -- Tentativa simplificada de extrair o nome da tabela da consulta normalizada
        -- (Em um ambiente real, seria necessário um parser SQL adequado)
        IF NEW.query_type = 'INSERT' THEN
            v_table_name := substring(NEW.normalized_query_text FROM 'INTO\s+([^\s\(]+)');
        ELSIF NEW.query_type IN ('UPDATE', 'DELETE') THEN
            v_table_name := substring(NEW.normalized_query_text FROM '(UPDATE|DELETE\s+FROM)\s+([^\s\(]+)');
        END IF;
        
        -- Separar schema e nome da tabela se tiver um ponto
        IF v_table_name LIKE '%.%' THEN
            v_schema_name := split_part(v_table_name, '.', 1);
            v_table_name := split_part(v_table_name, '.', 2);
        ELSE
            v_schema_name := 'public'; -- Assumir schema público por padrão
        END IF;
        
        -- Verificar quando as estatísticas foram atualizadas pela última vez
        SELECT last_stats_reset INTO v_last_updated
        FROM pg_stat_user_tables
        WHERE schemaname = v_schema_name AND relname = v_table_name;
        
        -- Contar alterações desde a última atualização
        SELECT COUNT(*) INTO v_changes_since_last_update
        FROM query_metrics
        WHERE 
            timestamp > COALESCE(v_last_updated, NOW() - INTERVAL '7 days')
            AND query_type IN ('INSERT', 'UPDATE', 'DELETE')
            AND normalized_query_text ILIKE '%' || v_table_name || '%';
        
        -- Se exceder o limiar, agendar atualização de estatísticas
        IF v_changes_since_last_update > v_threshold THEN
            INSERT INTO scheduled_maintenance_tasks (
                task_type,
                target_object,
                priority,
                scheduled_time,
                task_details,
                status
            ) VALUES (
                'ANALYZE',
                v_schema_name || '.' || v_table_name,
                'medium',
                NOW() + INTERVAL '10 minutes',
                jsonb_build_object(
                    'reason', 'High modification volume detected',
                    'changes_count', v_changes_since_last_update,
                    'last_stats_update', v_last_updated
                ),
                'scheduled'
            )
            ON CONFLICT (task_type, target_object, status) 
            WHERE status IN ('scheduled', 'in_progress')
            DO NOTHING;
        END IF;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger para agendar atualizações de estatísticas
DROP TRIGGER IF EXISTS trg_schedule_stats_update ON query_metrics;
CREATE TRIGGER trg_schedule_stats_update
AFTER INSERT ON query_metrics
FOR EACH ROW EXECUTE FUNCTION fn_schedule_stats_update();

-- ===============================================================================
-- TRIGGERS PARA GERAÇÃO AUTOMÁTICA DE ALERTAS DE DESEMPENHO
-- ===============================================================================

-- Função para gerar alertas de performance com base em métricas
CREATE OR REPLACE FUNCTION fn_generate_performance_alerts()
RETURNS TRIGGER AS $$
DECLARE
    v_alert_id UUID;
    v_active_connections_threshold INTEGER;
    v_cache_hit_ratio_threshold FLOAT;
    v_dead_tuples_threshold FLOAT;
BEGIN
    -- Obter configurações de limites
    SELECT COALESCE(CAST(value AS INTEGER), 100) INTO v_active_connections_threshold
    FROM monitoring_settings
    WHERE name = 'active_connections_threshold'
    LIMIT 1;
    
    SELECT COALESCE(CAST(value AS FLOAT), 0.8) INTO v_cache_hit_ratio_threshold
    FROM monitoring_settings
    WHERE name = 'cache_hit_ratio_threshold'
    LIMIT 1;
    
    SELECT COALESCE(CAST(value AS FLOAT), 0.2) INTO v_dead_tuples_threshold
    FROM monitoring_settings
    WHERE name = 'dead_tuples_threshold'
    LIMIT 1;
    
    -- Gerar alertas com base nas métricas coletadas
    
    -- Alerta para alto número de conexões ativas
    IF NEW.active_connections > v_active_connections_threshold THEN
        INSERT INTO performance_alerts (
            alert_type,
            severity,
            message,
            source_type,
            source_id,
            details,
            status
        ) VALUES (
            'high_connections',
            CASE 
                WHEN NEW.active_connections > v_active_connections_threshold * 1.5 THEN 'critical'
                WHEN NEW.active_connections > v_active_connections_threshold * 1.2 THEN 'high'
                ELSE 'medium'
            END,
            'Alto número de conexões ativas: ' || NEW.active_connections,
            'database_stats',
            NEW.id,
            jsonb_build_object(
                'active_connections', NEW.active_connections,
                'threshold', v_active_connections_threshold,
                'timestamp', NEW.timestamp,
                'recommended_action', 'Verifique conexões ociosas e considere ajustar pool de conexões ou implementar connection pooling.'
            ),
            'open'
        )
        -- Evitar duplicação de alertas em curto período
        ON CONFLICT (alert_type, status) 
        WHERE status = 'open' AND created_at > (NOW() - INTERVAL '4 hours')
        DO NOTHING
        RETURNING id INTO v_alert_id;
    END IF;
    
    -- Alerta para baixa taxa de acerto de cache
    IF NEW.cache_hit_ratio < v_cache_hit_ratio_threshold THEN
        INSERT INTO performance_alerts (
            alert_type,
            severity,
            message,
            source_type,
            source_id,
            details,
            status
        ) VALUES (
            'low_cache_hit_ratio',
            CASE 
                WHEN NEW.cache_hit_ratio < v_cache_hit_ratio_threshold * 0.5 THEN 'critical'
                WHEN NEW.cache_hit_ratio < v_cache_hit_ratio_threshold * 0.8 THEN 'high'
                ELSE 'medium'
            END,
            'Baixa taxa de acerto de cache: ' || ROUND(NEW.cache_hit_ratio * 100, 2) || '%',
            'database_stats',
            NEW.id,
            jsonb_build_object(
                'cache_hit_ratio', NEW.cache_hit_ratio,
                'threshold', v_cache_hit_ratio_threshold,
                'timestamp', NEW.timestamp,
                'recommended_action', 'Considere aumentar shared_buffers ou verificar se há grandes tabelas sem índices adequados.'
            ),
            'open'
        )
        -- Evitar duplicação de alertas em curto período
        ON CONFLICT (alert_type, status) 
        WHERE status = 'open' AND created_at > (NOW() - INTERVAL '4 hours')
        DO NOTHING
        RETURNING id INTO v_alert_id;
    END IF;
    
    -- Alerta para espaço de arquivos temporários elevado
    IF NEW.temp_files_size > 104857600 THEN -- 100MB
        INSERT INTO performance_alerts (
            alert_type,
            severity,
            message,
            source_type,
            source_id,
            details,
            status
        ) VALUES (
            'high_temp_usage',
            CASE 
                WHEN NEW.temp_files_size > 1073741824 THEN 'critical' -- 1GB
                WHEN NEW.temp_files_size > 524288000 THEN 'high' -- 500MB
                ELSE 'medium'
            END,
            'Alto uso de arquivos temporários: ' || pg_size_pretty(NEW.temp_files_size),
            'database_stats',
            NEW.id,
            jsonb_build_object(
                'temp_files_size', NEW.temp_files_size,
                'temp_files_size_pretty', pg_size_pretty(NEW.temp_files_size),
                'timestamp', NEW.timestamp,
                'recommended_action', 'Identifique consultas que estão usando espaço temporário e otimize-as. Considere aumentar work_mem para reduzir gravações em disco.'
            ),
            'open'
        )
        -- Evitar duplicação de alertas em curto período
        ON CONFLICT (alert_type, status) 
        WHERE status = 'open' AND created_at > (NOW() - INTERVAL '4 hours')
        DO NOTHING
        RETURNING id INTO v_alert_id;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger para gerar alertas de performance
DROP TRIGGER IF EXISTS trg_generate_performance_alerts ON database_stats;
CREATE TRIGGER trg_generate_performance_alerts
AFTER INSERT ON database_stats
FOR EACH ROW EXECUTE FUNCTION fn_generate_performance_alerts();

-- ===============================================================================
-- TRIGGERS PARA AUDITORIA DE MONITORAMENTO
-- ===============================================================================

-- Função para registrar alterações em alertas no log de auditoria
CREATE OR REPLACE FUNCTION fn_audit_alert_changes()
RETURNS TRIGGER AS $$
DECLARE
    v_action VARCHAR(10);
    v_old_data JSONB := NULL;
    v_new_data JSONB := NULL;
    v_changed_fields JSONB := NULL;
    v_user_id UUID;
BEGIN
    -- Determinar a ação realizada
    IF TG_OP = 'UPDATE' THEN
        v_action := 'UPDATE';
        v_old_data := to_jsonb(OLD);
        v_new_data := to_jsonb(NEW);
        
        -- Registrar apenas se o status de resolução foi alterado
        IF OLD.resolution_status IS DISTINCT FROM NEW.resolution_status THEN
            v_changed_fields := jsonb_build_object('resolution_status', jsonb_build_object(
                'old', OLD.resolution_status,
                'new', NEW.resolution_status
            ));
            
            -- Obter usuário que está fazendo a alteração
            v_user_id := current_setting('app.current_user_id', TRUE)::UUID;
            
            -- Registrar no log de auditoria
            INSERT INTO iam.audit_logs (
                entity_type,
                entity_id,
                action,
                old_data,
                new_data,
                changed_fields,
                user_id,
                application_name,
                timestamp
            ) VALUES (
                'performance_alerts',
                NEW.id,
                v_action,
                v_old_data,
                v_new_data,
                v_changed_fields,
                v_user_id,
                current_setting('application_name', TRUE),
                now()
            );
        END IF;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger para auditar alterações em alertas de performance
DROP TRIGGER IF EXISTS trg_audit_alert_changes ON performance_alerts;
CREATE TRIGGER trg_audit_alert_changes
AFTER UPDATE ON performance_alerts
FOR EACH ROW EXECUTE FUNCTION fn_audit_alert_changes();

COMMENT ON FUNCTION fn_auto_cleanup_query_metrics IS 'Limpa automaticamente dados antigos de métricas de consultas';
COMMENT ON FUNCTION fn_detect_slow_queries IS 'Detecta e registra consultas lentas automaticamente';
COMMENT ON FUNCTION fn_schedule_stats_update IS 'Agenda atualizações de estatísticas após grande volume de alterações';
COMMENT ON FUNCTION fn_generate_performance_alerts IS 'Gera alertas automáticos com base em métricas de desempenho';
COMMENT ON FUNCTION fn_audit_alert_changes IS 'Registra alterações em alertas de performance no log de auditoria';
