-- =====================================================================
-- INNOVABIZ - Utilitários Administrativos para Validadores de Conformidade
-- Versão: 1.0.0
-- Data de Criação: 14/05/2025
-- Autor: INNOVABIZ DevOps
-- Descrição: Funções auxiliares para administração, diagnóstico e manutenção
--            dos validadores de conformidade IAM
-- Regiões Suportadas: Global, UE, Brasil, Angola, EUA
-- =====================================================================

-- Criação de schema para utilitários administrativos
CREATE SCHEMA IF NOT EXISTS compliance_admin;

-- =====================================================================
-- Tabelas de Diagnóstico e Monitoramento
-- =====================================================================

-- Tabela para logs de execução e desempenho
CREATE TABLE compliance_admin.validation_performance_logs (
    log_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    validation_type VARCHAR(50) NOT NULL,
    sectors VARCHAR(30)[],
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    duration_ms INTEGER,
    requirements_validated INTEGER,
    error_count INTEGER DEFAULT 0,
    execution_details JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabela para configuração de alertas
CREATE TABLE compliance_admin.alert_configuration (
    alert_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_name VARCHAR(100) NOT NULL,
    description TEXT,
    condition_type VARCHAR(50) NOT NULL, -- IRR_THRESHOLD, VALIDATION_FAILURE, etc.
    condition_parameters JSONB NOT NULL,
    notification_channels VARCHAR(50)[] NOT NULL, -- EMAIL, SMS, WEBHOOK, etc.
    notification_recipients JSONB,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabela de registro de alertas disparados
CREATE TABLE compliance_admin.alert_history (
    alert_event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_id UUID REFERENCES compliance_admin.alert_configuration(alert_id),
    tenant_id UUID,
    triggered_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    alert_details JSONB,
    resolution_status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, ACKNOWLEDGED, RESOLVED
    resolution_notes TEXT,
    resolved_at TIMESTAMP WITH TIME ZONE
);

-- =====================================================================
-- Funções de Diagnóstico
-- =====================================================================

-- Função para verificar a integridade das tabelas de validadores
CREATE OR REPLACE FUNCTION compliance_admin.check_validator_schema_integrity()
RETURNS TABLE (
    schema_name VARCHAR(50),
    table_name VARCHAR(50),
    status VARCHAR(20),
    row_count BIGINT,
    details TEXT
) AS $$
DECLARE
    schemas VARCHAR(50)[] := ARRAY['compliance_validators', 'compliance_integrator', 'compliance_admin'];
    schema_rec RECORD;
    tab_rec RECORD;
    tab_count BIGINT;
BEGIN
    -- Para cada schema
    FOREACH schema_name IN ARRAY schemas LOOP
        -- Verificar se o schema existe
        PERFORM 1 FROM information_schema.schemata WHERE schema_name = check_validator_schema_integrity.schema_name;
        IF NOT FOUND THEN
            status := 'MISSING';
            row_count := 0;
            table_name := '';
            details := 'Schema não encontrado';
            RETURN NEXT;
            CONTINUE;
        END IF;
        
        -- Para cada tabela no schema
        FOR tab_rec IN 
            SELECT table_name AS name 
            FROM information_schema.tables 
            WHERE table_schema = check_validator_schema_integrity.schema_name
              AND table_type = 'BASE TABLE'
        LOOP
            -- Contar registros
            EXECUTE format('SELECT COUNT(*) FROM %I.%I', schema_name, tab_rec.name) INTO tab_count;
            
            table_name := tab_rec.name;
            status := 'OK';
            row_count := tab_count;
            details := 'Tabela íntegra';
            
            -- Verificar se há registros
            IF tab_count = 0 THEN
                details := 'Tabela vazia - possível problema';
                status := 'WARNING';
            END IF;
            
            RETURN NEXT;
        END LOOP;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Função para testar validadores e medir desempenho
CREATE OR REPLACE FUNCTION compliance_admin.test_validator_performance(
    tenant_id UUID,
    sector_id VARCHAR(30) DEFAULT NULL,
    iterations INTEGER DEFAULT 1
) RETURNS TABLE (
    validator_function TEXT,
    avg_duration_ms NUMERIC(10,2),
    min_duration_ms NUMERIC(10,2),
    max_duration_ms NUMERIC(10,2),
    requirements_count INTEGER,
    success_rate NUMERIC(5,2)
) AS $$
DECLARE
    validators TEXT[];
    validator TEXT;
    start_time TIMESTAMP WITH TIME ZONE;
    end_time TIMESTAMP WITH TIME ZONE;
    duration NUMERIC(10,2);
    total_duration NUMERIC(10,2);
    min_dur NUMERIC(10,2) := 999999;
    max_dur NUMERIC(10,2) := 0;
    success_count INTEGER := 0;
    req_count INTEGER := 0;
    i INTEGER;
    exec_result BOOLEAN;
BEGIN
    -- Determinar quais validadores testar
    IF sector_id IS NULL THEN
        SELECT array_agg(validator_function) INTO validators
        FROM compliance_integrator.sector_regulations;
    ELSE
        SELECT array_agg(validator_function) INTO validators
        FROM compliance_integrator.sector_regulations
        WHERE sector_id = test_validator_performance.sector_id;
    END IF;
    
    -- Para cada validador
    FOREACH validator IN ARRAY validators LOOP
        total_duration := 0;
        min_dur := 999999;
        max_dur := 0;
        success_count := 0;
        req_count := 0;
        
        -- Executar múltiplas iterações para média
        FOR i IN 1..iterations LOOP
            -- Registrar início
            INSERT INTO compliance_admin.validation_performance_logs (
                tenant_id, validation_type, start_time
            ) VALUES (
                tenant_id, validator, CURRENT_TIMESTAMP
            ) RETURNING start_time INTO start_time;
            
            -- Executar validação
            BEGIN
                -- Tentativa simplificada - em ambiente real, precisaria adaptar conforme retorno da função
                EXECUTE format('SELECT EXISTS(SELECT 1 FROM %s($1) LIMIT 1)', validator)
                USING tenant_id INTO exec_result;
                
                success_count := success_count + 1;
                
                -- Contar requisitos (simplificado)
                EXECUTE format('SELECT COUNT(*) FROM %s($1)', validator)
                USING tenant_id INTO req_count;
                
                -- Registrar fim
                end_time := CURRENT_TIMESTAMP;
                duration := EXTRACT(EPOCH FROM (end_time - start_time)) * 1000;
                
                -- Atualizar métricas
                total_duration := total_duration + duration;
                min_dur := LEAST(min_dur, duration);
                max_dur := GREATEST(max_dur, duration);
                
                -- Atualizar log
                UPDATE compliance_admin.validation_performance_logs
                SET 
                    end_time = end_time,
                    duration_ms = duration,
                    requirements_validated = req_count,
                    execution_details = jsonb_build_object('success', true)
                WHERE start_time = start_time AND tenant_id = test_validator_performance.tenant_id;
                
            EXCEPTION WHEN OTHERS THEN
                -- Registrar erro
                UPDATE compliance_admin.validation_performance_logs
                SET 
                    end_time = CURRENT_TIMESTAMP,
                    duration_ms = EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - start_time)) * 1000,
                    error_count = 1,
                    execution_details = jsonb_build_object('error', SQLERRM)
                WHERE start_time = start_time AND tenant_id = test_validator_performance.tenant_id;
            END;
        END LOOP;
        
        -- Calcular e retornar métricas
        validator_function := validator;
        avg_duration_ms := CASE WHEN success_count > 0 THEN total_duration / success_count ELSE 0 END;
        min_duration_ms := CASE WHEN min_dur < 999999 THEN min_dur ELSE 0 END;
        max_duration_ms := max_dur;
        requirements_count := req_count;
        success_rate := CASE WHEN iterations > 0 THEN (success_count::NUMERIC / iterations) * 100 ELSE 0 END;
        
        RETURN NEXT;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Funções de Manutenção e Administração
-- =====================================================================

-- Função para limpar histórico de validações antigas
CREATE OR REPLACE FUNCTION compliance_admin.purge_validation_history(
    older_than_days INTEGER DEFAULT 90,
    tenant_id UUID DEFAULT NULL
) RETURNS INTEGER AS $$
DECLARE
    deletion_cutoff TIMESTAMP WITH TIME ZONE := CURRENT_TIMESTAMP - make_interval(days => older_than_days);
    deleted_count INTEGER;
BEGIN
    IF tenant_id IS NULL THEN
        DELETE FROM compliance_integrator.validation_history
        WHERE validation_date < deletion_cutoff
        RETURNING COUNT(*) INTO deleted_count;
    ELSE
        DELETE FROM compliance_integrator.validation_history
        WHERE validation_date < deletion_cutoff
        AND tenant_id = purge_validation_history.tenant_id
        RETURNING COUNT(*) INTO deleted_count;
    END IF;
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Função para atualizar detecção de regulações por região
CREATE OR REPLACE FUNCTION compliance_admin.update_regulation_region_mapping(
    regulation_id VARCHAR(30),
    region VARCHAR(50),
    active BOOLEAN DEFAULT TRUE
) RETURNS BOOLEAN AS $$
DECLARE
    sector VARCHAR(30);
BEGIN
    -- Encontrar o setor da regulação
    SELECT sector_id INTO sector
    FROM compliance_integrator.sector_regulations
    WHERE regulation_id = update_regulation_region_mapping.regulation_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Regulação % não encontrada', regulation_id;
    END IF;
    
    -- Atualizar região da regulação
    UPDATE compliance_integrator.sector_regulations
    SET region = update_regulation_region_mapping.region
    WHERE regulation_id = update_regulation_region_mapping.regulation_id;
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- Função para configurar alerta de não-conformidade
CREATE OR REPLACE FUNCTION compliance_admin.configure_compliance_alert(
    alert_name VARCHAR(100),
    condition_type VARCHAR(50),
    irr_threshold VARCHAR(10) DEFAULT 'R3', -- relevante para IRR_THRESHOLD
    sector_ids VARCHAR(30)[] DEFAULT NULL,  -- relevante para SECTOR_COMPLIANCE
    notification_emails TEXT[] DEFAULT NULL,
    is_active BOOLEAN DEFAULT TRUE
) RETURNS UUID AS $$
DECLARE
    alert_id UUID;
    condition_params JSONB;
BEGIN
    -- Construir parâmetros da condição
    CASE condition_type
        WHEN 'IRR_THRESHOLD' THEN
            condition_params := jsonb_build_object(
                'irr_level', irr_threshold
            );
        WHEN 'SECTOR_COMPLIANCE' THEN
            condition_params := jsonb_build_object(
                'sectors', sector_ids,
                'threshold_percentage', 75
            );
        ELSE
            condition_params := jsonb_build_object(
                'custom_condition', TRUE
            );
    END CASE;
    
    -- Inserir configuração de alerta
    INSERT INTO compliance_admin.alert_configuration (
        alert_name,
        description,
        condition_type,
        condition_parameters,
        notification_channels,
        notification_recipients,
        is_active
    ) VALUES (
        alert_name,
        'Alerta de conformidade: ' || alert_name,
        condition_type,
        condition_params,
        ARRAY['EMAIL'],
        jsonb_build_object('emails', notification_emails),
        is_active
    ) RETURNING alert_id INTO alert_id;
    
    RETURN alert_id;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Funções de Visualização e Relatórios
-- =====================================================================

-- Função para gerar dashboard de status de conformidade
CREATE OR REPLACE FUNCTION compliance_admin.generate_compliance_dashboard(
    tenant_id UUID
) RETURNS JSONB AS $$
DECLARE
    result JSONB;
    sector_results JSONB;
    historical_trend JSONB;
    recent_validations JSONB;
    active_alerts JSONB;
BEGIN
    -- Obter resultados por setor
    SELECT jsonb_agg(
        jsonb_build_object(
            'sector_id', sector_id,
            'sector_name', sector_name,
            'compliance_score', compliance_score,
            'compliance_percentage', compliance_percentage,
            'irr', irr,
            'status', CASE 
                WHEN irr = 'R1' THEN 'EXCELENTE'
                WHEN irr = 'R2' THEN 'BOM'
                WHEN irr = 'R3' THEN 'RAZOÁVEL'
                WHEN irr = 'R4' THEN 'CRÍTICO'
                ELSE 'DESCONHECIDO'
            END
        )
    )
    FROM compliance_integrator.calculate_multi_sector_score(tenant_id)
    INTO sector_results;
    
    -- Obter tendência histórica (últimas 5 validações)
    SELECT jsonb_agg(
        jsonb_build_object(
            'date', validation_date,
            'score', score,
            'irr', irr
        )
        ORDER BY validation_date DESC
    )
    FROM compliance_integrator.validation_history
    WHERE tenant_id = generate_compliance_dashboard.tenant_id
    LIMIT 5
    INTO historical_trend;
    
    -- Obter validações recentes
    SELECT jsonb_agg(
        jsonb_build_object(
            'validation_id', validation_id,
            'validation_type', validation_type,
            'validation_date', validation_date,
            'score', score,
            'irr', irr
        )
        ORDER BY validation_date DESC
    )
    FROM compliance_integrator.validation_history
    WHERE tenant_id = generate_compliance_dashboard.tenant_id
    LIMIT 10
    INTO recent_validations;
    
    -- Obter alertas ativos
    SELECT jsonb_agg(
        jsonb_build_object(
            'alert_event_id', alert_event_id,
            'alert_name', ac.alert_name,
            'triggered_at', triggered_at,
            'status', resolution_status
        )
        ORDER BY triggered_at DESC
    )
    FROM compliance_admin.alert_history ah
    JOIN compliance_admin.alert_configuration ac ON ah.alert_id = ac.alert_id
    WHERE tenant_id = generate_compliance_dashboard.tenant_id
    AND resolution_status != 'RESOLVED'
    LIMIT 5
    INTO active_alerts;
    
    -- Construir dashboard completo
    result := jsonb_build_object(
        'tenant_id', tenant_id,
        'generated_at', CURRENT_TIMESTAMP,
        'overall_compliance', (
            SELECT jsonb_build_object(
                'score', compliance_score,
                'percentage', compliance_percentage,
                'irr', irr,
                'status', CASE 
                    WHEN irr = 'R1' THEN 'EXCELENTE'
                    WHEN irr = 'R2' THEN 'BOM'
                    WHEN irr = 'R3' THEN 'RAZOÁVEL'
                    WHEN irr = 'R4' THEN 'CRÍTICO'
                    ELSE 'DESCONHECIDO'
                END
            )
            FROM compliance_integrator.calculate_multi_sector_score(tenant_id)
            WHERE sector_id = 'OVERALL'
        ),
        'sector_results', COALESCE(sector_results, '[]'::JSONB),
        'historical_trend', COALESCE(historical_trend, '[]'::JSONB),
        'recent_validations', COALESCE(recent_validations, '[]'::JSONB),
        'active_alerts', COALESCE(active_alerts, '[]'::JSONB)
    );
    
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Comentários de Documentação
-- =====================================================================

COMMENT ON SCHEMA compliance_admin IS 'Esquema para utilitários administrativos dos validadores de conformidade';
COMMENT ON TABLE compliance_admin.validation_performance_logs IS 'Logs de desempenho de validações de conformidade';
COMMENT ON TABLE compliance_admin.alert_configuration IS 'Configuração de alertas de conformidade';
COMMENT ON TABLE compliance_admin.alert_history IS 'Histórico de alertas de conformidade disparados';

COMMENT ON FUNCTION compliance_admin.check_validator_schema_integrity IS 'Verifica a integridade do schema de validadores';
COMMENT ON FUNCTION compliance_admin.test_validator_performance IS 'Testa o desempenho dos validadores de conformidade';
COMMENT ON FUNCTION compliance_admin.purge_validation_history IS 'Limpa histórico de validações antigas';
COMMENT ON FUNCTION compliance_admin.update_regulation_region_mapping IS 'Atualiza mapeamento de regiões para regulações';
COMMENT ON FUNCTION compliance_admin.configure_compliance_alert IS 'Configura alerta de conformidade';
COMMENT ON FUNCTION compliance_admin.generate_compliance_dashboard IS 'Gera dashboard de status de conformidade';

-- =====================================================================
-- Fim do Script
-- =====================================================================
