-- INNOVABIZ IAM - Migração 003: Funções e procedimentos para auditoria
-- Autor: Eduardo Jeremias
-- Versão: 1.0.0
-- Descrição: Cria funções e procedures para o sistema de auditoria multi-contexto

-- Função para detectar automaticamente frameworks de compliance com base no contexto regional
CREATE OR REPLACE FUNCTION detect_compliance_frameworks(
    p_regional_context VARCHAR(10)
)
RETURNS TEXT[] AS $$
DECLARE
    v_frameworks TEXT[];
BEGIN
    CASE p_regional_context
        WHEN 'BR' THEN
            v_frameworks := ARRAY['LGPD', 'BACEN'];
        WHEN 'US' THEN
            v_frameworks := ARRAY['SOX', 'NIST'];
        WHEN 'EU' THEN
            v_frameworks := ARRAY['GDPR', 'PSD2'];
        WHEN 'AO' THEN
            v_frameworks := ARRAY['BNA'];
        ELSE
            v_frameworks := ARRAY[]::TEXT[];
    END CASE;
    
    -- Sempre adiciona PCI_DSS como framework global para operações financeiras
    v_frameworks := array_append(v_frameworks, 'PCI_DSS');
    
    RETURN v_frameworks;
END;
$$ LANGUAGE plpgsql;

-- Função para gerar a chave de particionamento baseada no tenant, região e data
CREATE OR REPLACE FUNCTION generate_partition_key(
    p_tenant_id VARCHAR(100),
    p_regional_context VARCHAR(10),
    p_timestamp TIMESTAMPTZ DEFAULT NULL
)
RETURNS VARCHAR(255) AS $$
DECLARE
    v_timestamp TIMESTAMPTZ;
    v_year_month VARCHAR(7);
    v_regional_context VARCHAR(10);
BEGIN
    -- Define timestamp padrão se não fornecido
    v_timestamp := COALESCE(p_timestamp, NOW());
    
    -- Formata ano-mês (YYYY-MM)
    v_year_month := TO_CHAR(v_timestamp, 'YYYY-MM');
    
    -- Usa 'GLOBAL' como valor padrão se contexto regional não for fornecido
    v_regional_context := COALESCE(p_regional_context, 'GLOBAL');
    
    -- Retorna chave de particionamento no formato tenant_id:regional_context:YYYY-MM
    RETURN p_tenant_id || ':' || v_regional_context || ':' || v_year_month;
END;
$$ LANGUAGE plpgsql;

-- Procedimento para aplicar políticas de retenção aos eventos de auditoria
CREATE OR REPLACE PROCEDURE apply_retention_policy(
    p_policy_id UUID
)
LANGUAGE plpgsql AS $$
DECLARE
    v_policy RECORD;
    v_cutoff_date TIMESTAMPTZ;
    v_anonymized_count INTEGER := 0;
    v_deleted_count INTEGER := 0;
BEGIN
    -- Obtém detalhes da política
    SELECT * INTO v_policy FROM audit_retention_policies WHERE id = p_policy_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Política de retenção não encontrada: %', p_policy_id;
    END IF;
    
    -- Calcula data de corte baseada nos dias de retenção
    v_cutoff_date := NOW() - (v_policy.retention_days || ' days')::INTERVAL;
    
    -- Se anonimização automática está ativada, anonimiza campos sensíveis
    IF v_policy.automatic_anonymization THEN
        WITH updated_events AS (
            UPDATE audit_events 
            SET 
                anonymized_fields = v_policy.anonymization_fields,
                details = jsonb_strip_nulls(jsonb_set(
                    details,
                    '{anonymized_at}',
                    to_jsonb(NOW())
                ))
            WHERE 
                tenant_id = v_policy.tenant_id
                AND (v_policy.regional_context IS NULL OR regional_context = v_policy.regional_context)
                AND (v_policy.category IS NULL OR category = v_policy.category)
                AND created_at < v_cutoff_date
                AND array_length(anonymized_fields, 1) IS NULL OR NOT anonymized_fields @> v_policy.anonymization_fields
            RETURNING id
        )
        SELECT COUNT(*) INTO v_anonymized_count FROM updated_events;
    END IF;
    
    -- Registra aplicação da política nas estatísticas
    INSERT INTO audit_statistics (
        tenant_id,
        regional_context,
        statistics_type,
        period_start,
        period_end,
        statistics_data
    ) VALUES (
        v_policy.tenant_id,
        v_policy.regional_context,
        'RETENTION_POLICY_APPLICATION',
        v_cutoff_date,
        NOW(),
        jsonb_build_object(
            'policy_id', v_policy.id,
            'anonymized_count', v_anonymized_count,
            'deleted_count', v_deleted_count,
            'retention_days', v_policy.retention_days,
            'compliance_framework', v_policy.compliance_framework
        )
    );
    
    RAISE NOTICE 'Política de retenção aplicada: % eventos anonimizados', v_anonymized_count;
END;
$$;

-- Função para mascarar campos sensíveis em um evento de auditoria
CREATE OR REPLACE FUNCTION mask_sensitive_fields(
    p_event_id UUID,
    p_field_names TEXT[]
)
RETURNS VOID AS $$
DECLARE
    v_event RECORD;
    v_field TEXT;
    v_masked_fields TEXT[];
    v_details JSONB;
    v_path TEXT[];
    v_key TEXT;
BEGIN
    -- Obtém o evento atual
    SELECT * INTO v_event FROM audit_events WHERE id = p_event_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Evento de auditoria não encontrado: %', p_event_id;
    END IF;
    
    -- Inicializa campos mascarados com valores existentes
    v_masked_fields := COALESCE(v_event.masked_fields, ARRAY[]::TEXT[]);
    v_details := COALESCE(v_event.details, '{}'::JSONB);
    
    -- Para cada campo a ser mascarado
    FOREACH v_field IN ARRAY p_field_names
    LOOP
        -- Verifica se o campo é um caminho JSON (contém '.')
        IF v_field LIKE '%.%' THEN
            v_path := string_to_array(v_field, '.');
            v_key := v_path[1];
            
            -- Mascara o campo nos detalhes JSONB
            IF v_details ? v_key THEN
                v_details := jsonb_set(v_details, v_path, '"[MASKED]"'::JSONB);
            END IF;
        ELSE
            -- Mascara campos diretos como source_ip
            EXECUTE format('
                UPDATE audit_events 
                SET %I = ''[MASKED]'' 
                WHERE id = $1',
                v_field
            ) USING p_event_id;
        END IF;
        
        -- Adiciona à lista de campos mascarados se não estiver lá
        IF NOT v_field = ANY(v_masked_fields) THEN
            v_masked_fields := array_append(v_masked_fields, v_field);
        END IF;
    END LOOP;
    
    -- Atualiza o registro com os novos campos mascarados e detalhes
    UPDATE audit_events
    SET 
        masked_fields = v_masked_fields,
        details = v_details
    WHERE id = p_event_id;
END;
$$ LANGUAGE plpgsql;

-- Função para gerar estatísticas de eventos de auditoria
CREATE OR REPLACE FUNCTION generate_audit_statistics(
    p_tenant_id VARCHAR(100),
    p_regional_context VARCHAR(10) DEFAULT NULL,
    p_start_date TIMESTAMPTZ,
    p_end_date TIMESTAMPTZ,
    p_group_by TEXT[] DEFAULT ARRAY['category', 'success']
)
RETURNS UUID AS $$
DECLARE
    v_statistics_id UUID;
    v_result JSONB := '{}'::JSONB;
    v_group_field TEXT;
    v_query TEXT;
    v_count INTEGER;
BEGIN
    -- Cria ID para as estatísticas
    v_statistics_id := gen_random_uuid();
    
    -- Contagem total de eventos
    SELECT COUNT(*) INTO v_count
    FROM audit_events
    WHERE 
        tenant_id = p_tenant_id
        AND (p_regional_context IS NULL OR regional_context = p_regional_context)
        AND created_at BETWEEN p_start_date AND p_end_date;
    
    v_result := jsonb_set(v_result, '{total_events}', to_jsonb(v_count));
    
    -- Para cada campo de agrupamento, gera estatísticas
    FOREACH v_group_field IN ARRAY p_group_by
    LOOP
        CASE v_group_field
            WHEN 'category' THEN
                WITH category_stats AS (
                    SELECT 
                        category::TEXT,
                        COUNT(*) as count
                    FROM audit_events
                    WHERE 
                        tenant_id = p_tenant_id
                        AND (p_regional_context IS NULL OR regional_context = p_regional_context)
                        AND created_at BETWEEN p_start_date AND p_end_date
                    GROUP BY category
                )
                SELECT jsonb_object_agg(category, count) INTO v_result 
                FROM (
                    SELECT 
                        jsonb_set(v_result, '{categories}', 
                            (SELECT jsonb_object_agg(category, count) FROM category_stats)
                        )
                ) subq;
                
            WHEN 'success' THEN
                WITH success_stats AS (
                    SELECT 
                        success,
                        COUNT(*) as count
                    FROM audit_events
                    WHERE 
                        tenant_id = p_tenant_id
                        AND (p_regional_context IS NULL OR regional_context = p_regional_context)
                        AND created_at BETWEEN p_start_date AND p_end_date
                    GROUP BY success
                )
                SELECT jsonb_set(v_result, '{success_rate}', 
                    (SELECT jsonb_object_agg(success, count) FROM success_stats)
                ) INTO v_result;
                
            WHEN 'severity' THEN
                WITH severity_stats AS (
                    SELECT 
                        severity::TEXT,
                        COUNT(*) as count
                    FROM audit_events
                    WHERE 
                        tenant_id = p_tenant_id
                        AND (p_regional_context IS NULL OR regional_context = p_regional_context)
                        AND created_at BETWEEN p_start_date AND p_end_date
                    GROUP BY severity
                )
                SELECT jsonb_set(v_result, '{severities}', 
                    (SELECT jsonb_object_agg(severity, count) FROM severity_stats)
                ) INTO v_result;
                
            WHEN 'user_id' THEN
                WITH user_stats AS (
                    SELECT 
                        user_id,
                        COUNT(*) as count
                    FROM audit_events
                    WHERE 
                        tenant_id = p_tenant_id
                        AND (p_regional_context IS NULL OR regional_context = p_regional_context)
                        AND created_at BETWEEN p_start_date AND p_end_date
                        AND user_id IS NOT NULL
                    GROUP BY user_id
                    ORDER BY count DESC
                    LIMIT 10
                )
                SELECT jsonb_set(v_result, '{top_users}', 
                    (SELECT jsonb_object_agg(user_id, count) FROM user_stats)
                ) INTO v_result;
        END CASE;
    END LOOP;
    
    -- Insere estatísticas na tabela
    INSERT INTO audit_statistics (
        id,
        tenant_id,
        regional_context,
        statistics_type,
        period_start,
        period_end,
        statistics_data
    ) VALUES (
        v_statistics_id,
        p_tenant_id,
        p_regional_context,
        'PERIODIC_SUMMARY',
        p_start_date,
        p_end_date,
        v_result
    );
    
    RETURN v_statistics_id;
END;
$$ LANGUAGE plpgsql;

-- Comentários para documentação
COMMENT ON FUNCTION detect_compliance_frameworks IS 'Detecta automaticamente os frameworks de compliance aplicáveis com base no contexto regional';
COMMENT ON FUNCTION generate_partition_key IS 'Gera uma chave de particionamento lógico para eventos de auditoria';
COMMENT ON PROCEDURE apply_retention_policy IS 'Aplica políticas de retenção e anonimização a eventos de auditoria';
COMMENT ON FUNCTION mask_sensitive_fields IS 'Mascara campos sensíveis em um evento de auditoria';
COMMENT ON FUNCTION generate_audit_statistics IS 'Gera estatísticas agregadas para eventos de auditoria';