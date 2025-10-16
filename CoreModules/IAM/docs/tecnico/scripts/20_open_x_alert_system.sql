-- =============================================
-- Sistema de Alertas Inteligentes para Open X
-- INNOVABIZ - IAM - Open X Alert System
-- Versão: 1.0.0
-- Data: 15/05/2025
-- Autor: INNOVABIZ DevOps Team
-- =============================================

-- Este script implementa um sistema de alertas inteligentes para o ecossistema Open X,
-- permitindo detecção proativa de não-conformidades críticas, notificações
-- e integração com diversos canais de comunicação.

-- =============================================
-- Schemas necessários
-- =============================================

-- Verifica e cria os schemas necessários
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'compliance_alert_system') THEN
        CREATE SCHEMA compliance_alert_system;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'compliance_validators') THEN
        RAISE EXCEPTION 'O schema compliance_validators não existe. Execute primeiro os scripts de validadores.';
    END IF;
END $$;

-- =============================================
-- Enumerações e Tipos
-- =============================================

-- Tipo de enum para severidade de alertas
CREATE TYPE compliance_alert_system.alert_severity AS ENUM (
    'INFORMATIVO',
    'BAIXO',
    'MEDIO',
    'ALTO',
    'CRITICO'
);

-- Tipo de enum para status de alertas
CREATE TYPE compliance_alert_system.alert_status AS ENUM (
    'NOVO',
    'RECONHECIDO',
    'EM_ANALISE',
    'EM_MITIGACAO',
    'RESOLVIDO',
    'FECHADO',
    'FALSO_POSITIVO',
    'DUPLICADO',
    'ADIADO'
);

-- Tipo de enum para canais de notificação
CREATE TYPE compliance_alert_system.notification_channel AS ENUM (
    'EMAIL',
    'SMS',
    'PUSH',
    'WEBHOOK',
    'SLACK',
    'TEAMS',
    'TICKET_SYSTEM',
    'API'
);

-- Tipo composto para configuração de regras de alerta
CREATE TYPE compliance_alert_system.alert_rule_config AS (
    condition_type VARCHAR(50),
    threshold_value NUMERIC,
    comparison_operator VARCHAR(10),
    time_window_minutes INTEGER,
    consecutive_occurrences INTEGER,
    cooldown_minutes INTEGER,
    requires_acknowledgment BOOLEAN
);

-- Tipo composto para destinatário de notificação
CREATE TYPE compliance_alert_system.notification_recipient AS (
    recipient_type VARCHAR(20),
    recipient_id VARCHAR(100),
    recipient_name VARCHAR(255),
    channel compliance_alert_system.notification_channel,
    notification_template_id VARCHAR(50)
);

-- =============================================
-- Tabelas para o Sistema de Alertas
-- =============================================

-- Tabela para regras de alertas
CREATE TABLE compliance_alert_system.alert_rules (
    rule_id VARCHAR(50) PRIMARY KEY,
    rule_name VARCHAR(255) NOT NULL,
    description TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    open_x_domain VARCHAR(50),
    framework VARCHAR(50),
    requirement_ids VARCHAR[] DEFAULT NULL,
    irr_thresholds VARCHAR[] DEFAULT ARRAY['R3', 'R4'],
    alert_severity compliance_alert_system.alert_severity NOT NULL,
    rule_config compliance_alert_system.alert_rule_config NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    tenant_id UUID,
    notification_groups VARCHAR[] DEFAULT NULL,
    auto_remediation_action_id VARCHAR(50) DEFAULT NULL,
    custom_query TEXT DEFAULT NULL,
    priority INTEGER DEFAULT 5,
    tags VARCHAR[] DEFAULT NULL
);

-- Tabela para alertas gerados
CREATE TABLE compliance_alert_system.alerts (
    alert_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id VARCHAR(50) REFERENCES compliance_alert_system.alert_rules(rule_id),
    tenant_id UUID NOT NULL,
    open_x_domain VARCHAR(50) NOT NULL,
    framework VARCHAR(50),
    requirement_ids VARCHAR[] NOT NULL,
    alert_title VARCHAR(255) NOT NULL,
    alert_message TEXT NOT NULL,
    details JSONB,
    alert_severity compliance_alert_system.alert_severity NOT NULL,
    status compliance_alert_system.alert_status DEFAULT 'NOVO',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP WITH TIME ZONE,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    acknowledged_by VARCHAR(100),
    expected_resolution_date TIMESTAMP WITH TIME ZONE,
    resolution_notes TEXT,
    economic_impact NUMERIC,
    related_alerts UUID[],
    related_incidents VARCHAR[]
);

-- Tabela para histórico de alertas
CREATE TABLE compliance_alert_system.alert_history (
    history_id SERIAL PRIMARY KEY,
    alert_id UUID REFERENCES compliance_alert_system.alerts(alert_id),
    field_changed VARCHAR(50) NOT NULL,
    old_value TEXT,
    new_value TEXT,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    changed_by VARCHAR(100),
    notes TEXT
);

-- Tabela para ações de remediação automáticas
CREATE TABLE compliance_alert_system.auto_remediation_actions (
    action_id VARCHAR(50) PRIMARY KEY,
    action_name VARCHAR(255) NOT NULL,
    description TEXT,
    action_type VARCHAR(50) NOT NULL,
    action_parameters JSONB,
    script_content TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    approval_required BOOLEAN DEFAULT FALSE,
    approval_role VARCHAR(50)
);

-- Tabela para configurações de notificação
CREATE TABLE compliance_alert_system.notification_configurations (
    config_id VARCHAR(50) PRIMARY KEY,
    tenant_id UUID,
    notification_group VARCHAR(50) NOT NULL,
    open_x_domain VARCHAR(50),
    alert_severity compliance_alert_system.alert_severity,
    recipients compliance_alert_system.notification_recipient[] NOT NULL,
    template_id VARCHAR(50),
    throttling_enabled BOOLEAN DEFAULT FALSE,
    throttling_window_minutes INTEGER DEFAULT 60,
    throttling_max_notifications INTEGER DEFAULT 10,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabela para modelos de notificação
CREATE TABLE compliance_alert_system.notification_templates (
    template_id VARCHAR(50) PRIMARY KEY,
    template_name VARCHAR(255) NOT NULL,
    subject_template TEXT,
    body_template TEXT NOT NULL,
    channel_type compliance_alert_system.notification_channel,
    format VARCHAR(20) DEFAULT 'TEXT',
    variables JSONB,
    tenant_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabela para notificações enviadas
CREATE TABLE compliance_alert_system.notifications (
    notification_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_id UUID REFERENCES compliance_alert_system.alerts(alert_id),
    config_id VARCHAR(50) REFERENCES compliance_alert_system.notification_configurations(config_id),
    channel compliance_alert_system.notification_channel NOT NULL,
    recipient VARCHAR(255) NOT NULL,
    subject TEXT,
    content TEXT NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'ENVIADO',
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    external_reference_id VARCHAR(100)
);

-- Tabela para métricas de desempenho dos alertas
CREATE TABLE compliance_alert_system.alert_metrics (
    metric_id SERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    open_x_domain VARCHAR(50),
    total_alerts INTEGER DEFAULT 0,
    critical_alerts INTEGER DEFAULT 0,
    high_alerts INTEGER DEFAULT 0,
    medium_alerts INTEGER DEFAULT 0,
    low_alerts INTEGER DEFAULT 0,
    informative_alerts INTEGER DEFAULT 0,
    avg_resolution_time_minutes NUMERIC DEFAULT 0,
    min_resolution_time_minutes NUMERIC,
    max_resolution_time_minutes NUMERIC,
    false_positive_count INTEGER DEFAULT 0,
    most_common_rule_id VARCHAR(50),
    most_common_requirement_id VARCHAR(50),
    escalated_count INTEGER DEFAULT 0,
    sla_compliant_count INTEGER DEFAULT 0,
    sla_violated_count INTEGER DEFAULT 0,
    calculated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- =============================================
-- Funções para Detecção de Alertas
-- =============================================

-- Função para verificar se um valor atende a condição de uma regra
CREATE OR REPLACE FUNCTION compliance_alert_system.check_condition(
    p_value NUMERIC,
    p_operator VARCHAR(10),
    p_threshold NUMERIC
) RETURNS BOOLEAN AS $$
BEGIN
    CASE p_operator
        WHEN '=' THEN RETURN p_value = p_threshold;
        WHEN '!=' THEN RETURN p_value != p_threshold;
        WHEN '<' THEN RETURN p_value < p_threshold;
        WHEN '<=' THEN RETURN p_value <= p_threshold;
        WHEN '>' THEN RETURN p_value > p_threshold;
        WHEN '>=' THEN RETURN p_value >= p_threshold;
        ELSE RAISE EXCEPTION 'Operador de comparação não suportado: %', p_operator;
    END CASE;
END;
$$ LANGUAGE plpgsql;

-- Função para buscar não-conformidades críticas por domínio
CREATE OR REPLACE FUNCTION compliance_alert_system.get_critical_non_compliances(
    p_tenant_id UUID,
    p_open_x_domain VARCHAR DEFAULT NULL,
    p_framework VARCHAR DEFAULT NULL,
    p_irr_thresholds VARCHAR[] DEFAULT ARRAY['R3', 'R4']
) RETURNS TABLE (
    open_x_domain VARCHAR,
    framework VARCHAR,
    requirement_id VARCHAR,
    requirement_name VARCHAR,
    irr_threshold VARCHAR,
    validation_details TEXT,
    economic_impact NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        nc.open_x_domain,
        nc.framework,
        nc.requirement_id,
        nc.requirement_name,
        nc.irr_threshold,
        nc.details AS validation_details,
        COALESCE(ei.impact_amount, economic_planning.calculate_default_impact(nc.open_x_domain, nc.framework, nc.irr_threshold)) AS economic_impact
    FROM 
        compliance_validators.vw_open_x_non_compliance_details nc
    LEFT JOIN 
        economic_planning.economic_impacts ei 
        ON nc.tenant_id = ei.tenant_id AND nc.requirement_id = ei.validator_id
    WHERE 
        nc.tenant_id = p_tenant_id
        AND (p_open_x_domain IS NULL OR nc.open_x_domain = p_open_x_domain)
        AND (p_framework IS NULL OR nc.framework = p_framework)
        AND nc.irr_threshold = ANY(p_irr_thresholds)
    ORDER BY 
        CASE nc.irr_threshold
            WHEN 'R4' THEN 1
            WHEN 'R3' THEN 2
            WHEN 'R2' THEN 3
            ELSE 4
        END,
        economic_impact DESC;
END;
$$ LANGUAGE plpgsql;

-- Função para analisar tendências de conformidade
CREATE OR REPLACE FUNCTION compliance_alert_system.analyze_compliance_trend(
    p_tenant_id UUID,
    p_open_x_domain VARCHAR,
    p_framework VARCHAR DEFAULT NULL,
    p_days_back INTEGER DEFAULT 30
) RETURNS TABLE (
    trend_direction VARCHAR,
    percentage_change NUMERIC,
    days_analyzed INTEGER,
    current_compliance_percentage NUMERIC,
    previous_compliance_percentage NUMERIC
) AS $$
DECLARE
    v_current_compliance NUMERIC;
    v_previous_compliance NUMERIC;
    v_percentage_change NUMERIC;
    v_trend_direction VARCHAR;
BEGIN
    -- Obter conformidade atual
    SELECT 
        AVG(compliance_percentage) INTO v_current_compliance
    FROM 
        compliance_validators.vw_open_x_compliance_summary
    WHERE 
        tenant_id = p_tenant_id
        AND open_x_domain = p_open_x_domain
        AND (p_framework IS NULL OR framework = p_framework);
    
    -- Obter conformidade anterior (do período especificado)
    SELECT 
        AVG(compliance_percentage) INTO v_previous_compliance
    FROM 
        compliance_validators.open_x_validation_history_summary
    WHERE 
        tenant_id = p_tenant_id
        AND open_x_domain = p_open_x_domain
        AND (p_framework IS NULL OR framework = p_framework)
        AND validation_date BETWEEN (CURRENT_DATE - (p_days_back || ' days')::INTERVAL) AND (CURRENT_DATE - '1 day'::INTERVAL);
    
    -- Calcular mudança percentual
    IF v_previous_compliance > 0 THEN
        v_percentage_change := ((v_current_compliance - v_previous_compliance) / v_previous_compliance) * 100;
    ELSE
        v_percentage_change := 0;
    END IF;
    
    -- Determinar direção da tendência
    IF v_percentage_change > 0 THEN
        v_trend_direction := 'MELHORIA';
    ELSIF v_percentage_change < 0 THEN
        v_trend_direction := 'DETERIORACAO';
    ELSE
        v_trend_direction := 'ESTAVEL';
    END IF;
    
    -- Retornar resultados
    trend_direction := v_trend_direction;
    percentage_change := v_percentage_change;
    days_analyzed := p_days_back;
    current_compliance_percentage := v_current_compliance;
    previous_compliance_percentage := v_previous_compliance;
    
    RETURN NEXT;
END;
$$ LANGUAGE plpgsql;

-- Função para detectar ocorrências consecutivas de não-conformidade
CREATE OR REPLACE FUNCTION compliance_alert_system.check_consecutive_non_compliance(
    p_tenant_id UUID,
    p_requirement_id VARCHAR,
    p_consecutive_occurrences INTEGER
) RETURNS BOOLEAN AS $$
DECLARE
    v_consecutive_count INTEGER;
BEGIN
    SELECT 
        COUNT(*) INTO v_consecutive_count
    FROM (
        SELECT 
            validation_date,
            is_compliant
        FROM 
            compliance_validators.open_x_validation_history
        WHERE 
            tenant_id = p_tenant_id
            AND requirement_id = p_requirement_id
        ORDER BY 
            validation_date DESC
        LIMIT p_consecutive_occurrences
    ) AS recent_validations
    WHERE 
        NOT is_compliant;
    
    RETURN v_consecutive_count >= p_consecutive_occurrences;
END;
$$ LANGUAGE plpgsql;

-- Função para verificar alarmes de deterioração de conformidade
CREATE OR REPLACE FUNCTION compliance_alert_system.check_compliance_deterioration(
    p_tenant_id UUID,
    p_open_x_domain VARCHAR,
    p_framework VARCHAR DEFAULT NULL,
    p_threshold_percentage NUMERIC DEFAULT 10.0,
    p_days_back INTEGER DEFAULT 30
) RETURNS BOOLEAN AS $$
DECLARE
    v_trend_result RECORD;
BEGIN
    SELECT * INTO v_trend_result
    FROM compliance_alert_system.analyze_compliance_trend(
        p_tenant_id,
        p_open_x_domain,
        p_framework,
        p_days_back
    );
    
    RETURN v_trend_result.trend_direction = 'DETERIORACAO' 
        AND ABS(v_trend_result.percentage_change) >= p_threshold_percentage;
END;
$$ LANGUAGE plpgsql;

-- Função principal para verificar alertas baseados em regras
CREATE OR REPLACE FUNCTION compliance_alert_system.evaluate_alert_rules(
    p_tenant_id UUID
) RETURNS TABLE (
    rule_id VARCHAR,
    open_x_domain VARCHAR,
    framework VARCHAR,
    requirement_ids VARCHAR[],
    alert_title VARCHAR,
    alert_message TEXT,
    alert_severity compliance_alert_system.alert_severity,
    economic_impact NUMERIC,
    details JSONB
) AS $$
DECLARE
    v_rule RECORD;
    v_rule_config compliance_alert_system.alert_rule_config;
    v_alert_data RECORD;
    v_non_compliance RECORD;
    v_alert_details JSONB;
    v_requirement_ids VARCHAR[] := '{}';
    v_economic_impact NUMERIC := 0;
BEGIN
    -- Iterar sobre todas as regras de alerta ativas
    FOR v_rule IN 
        SELECT * FROM compliance_alert_system.alert_rules 
        WHERE enabled = TRUE AND (tenant_id IS NULL OR tenant_id = p_tenant_id)
    LOOP
        v_rule_config := v_rule.rule_config;
        
        -- Verificar condições baseadas no tipo
        CASE v_rule_config.condition_type
            -- Alerta baseado em não-conformidades críticas
            WHEN 'CRITICAL_NON_COMPLIANCE' THEN
                -- Verificar se existem não-conformidades críticas
                FOR v_non_compliance IN 
                    SELECT * FROM compliance_alert_system.get_critical_non_compliances(
                        p_tenant_id,
                        v_rule.open_x_domain,
                        v_rule.framework,
                        v_rule.irr_thresholds
                    )
                LOOP
                    -- Se requisito específico esperado e não encontrado, pular
                    IF v_rule.requirement_ids IS NOT NULL AND 
                       NOT (v_non_compliance.requirement_id = ANY(v_rule.requirement_ids)) THEN
                        CONTINUE;
                    END IF;
                    
                    -- Verificar se já existe alerta ativo para esta não-conformidade
                    PERFORM 1 
                    FROM compliance_alert_system.alerts
                    WHERE tenant_id = p_tenant_id
                        AND rule_id = v_rule.rule_id
                        AND open_x_domain = v_non_compliance.open_x_domain
                        AND v_non_compliance.requirement_id = ANY(requirement_ids)
                        AND status NOT IN ('RESOLVIDO', 'FECHADO', 'FALSO_POSITIVO');
                        
                    IF FOUND THEN
                        CONTINUE; -- Alerta já existe
                    END IF;
                    
                    -- Adicionar à lista de requisitos
                    v_requirement_ids := array_append(v_requirement_ids, v_non_compliance.requirement_id);
                    v_economic_impact := v_economic_impact + v_non_compliance.economic_impact;
                    
                    -- Construir detalhes do alerta
                    v_alert_details := jsonb_build_object(
                        'requirement_id', v_non_compliance.requirement_id,
                        'requirement_name', v_non_compliance.requirement_name,
                        'irr_threshold', v_non_compliance.irr_threshold,
                        'validation_details', v_non_compliance.validation_details,
                        'economic_impact', v_non_compliance.economic_impact
                    );
                    
                    rule_id := v_rule.rule_id;
                    open_x_domain := v_non_compliance.open_x_domain;
                    framework := v_non_compliance.framework;
                    requirement_ids := ARRAY[v_non_compliance.requirement_id];
                    alert_title := 'Não-conformidade crítica detectada: ' || v_non_compliance.requirement_id;
                    alert_message := 'Requisito "' || v_non_compliance.requirement_name || '" não está em conformidade. Nível de risco: ' || v_non_compliance.irr_threshold;
                    alert_severity := v_rule.alert_severity;
                    economic_impact := v_non_compliance.economic_impact;
                    details := v_alert_details;
                    
                    RETURN NEXT;
                END LOOP;
                
            -- Alerta baseado em deterioração de conformidade
            WHEN 'DETERIORATION_TREND' THEN
                IF compliance_alert_system.check_compliance_deterioration(
                    p_tenant_id,
                    v_rule.open_x_domain,
                    v_rule.framework,
                    v_rule_config.threshold_value,
                    v_rule_config.time_window_minutes / (60 * 24) -- converter minutos para dias
                ) THEN
                    -- Obter detalhes da tendência
                    SELECT * INTO v_alert_data
                    FROM compliance_alert_system.analyze_compliance_trend(
                        p_tenant_id,
                        v_rule.open_x_domain,
                        v_rule.framework,
                        v_rule_config.time_window_minutes / (60 * 24)
                    );
                    
                    -- Verificar se já existe alerta ativo para esta deterioração
                    PERFORM 1 
                    FROM compliance_alert_system.alerts
                    WHERE tenant_id = p_tenant_id
                        AND rule_id = v_rule.rule_id
                        AND open_x_domain = v_rule.open_x_domain
                        AND (v_rule.framework IS NULL OR framework = v_rule.framework)
                        AND status NOT IN ('RESOLVIDO', 'FECHADO', 'FALSO_POSITIVO')
                        AND created_at > (CURRENT_TIMESTAMP - (v_rule_config.cooldown_minutes || ' minutes')::INTERVAL);
                        
                    IF FOUND THEN
                        CONTINUE; -- Alerta já existe ou em período de cooldown
                    END IF;
                    
                    -- Construir detalhes do alerta
                    v_alert_details := jsonb_build_object(
                        'trend_direction', v_alert_data.trend_direction,
                        'percentage_change', v_alert_data.percentage_change,
                        'days_analyzed', v_alert_data.days_analyzed,
                        'current_compliance_percentage', v_alert_data.current_compliance_percentage,
                        'previous_compliance_percentage', v_alert_data.previous_compliance_percentage
                    );
                    
                    rule_id := v_rule.rule_id;
                    open_x_domain := v_rule.open_x_domain;
                    framework := v_rule.framework;
                    requirement_ids := '{}'::VARCHAR[];
                    alert_title := 'Deterioração na conformidade detectada: ' || v_rule.open_x_domain;
                    alert_message := 'A conformidade diminuiu ' || ABS(v_alert_data.percentage_change) || '% nos últimos ' || v_alert_data.days_analyzed || ' dias (de ' || v_alert_data.previous_compliance_percentage || '% para ' || v_alert_data.current_compliance_percentage || '%)';
                    alert_severity := v_rule.alert_severity;
                    economic_impact := NULL;
                    details := v_alert_details;
                    
                    RETURN NEXT;
                END IF;
                
            -- Outros tipos de condição podem ser adicionados conforme necessário
            ELSE
                -- Nada a fazer para tipos desconhecidos
                NULL;
        END CASE;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- =============================================
-- Funções para Processamento de Notificações
-- =============================================

-- Função para formatar uma mensagem de alerta usando um template
CREATE OR REPLACE FUNCTION compliance_alert_system.format_notification(
    p_template_id VARCHAR(50),
    p_alert_data JSONB,
    p_tenant_id UUID DEFAULT NULL
) RETURNS TABLE (
    subject TEXT,
    body TEXT
) AS $$
DECLARE
    v_template RECORD;
    v_subject TEXT;
    v_body TEXT;
    v_variable RECORD;
    v_placeholder VARCHAR;
    v_value TEXT;
BEGIN
    -- Obter o template
    SELECT * INTO v_template
    FROM compliance_alert_system.notification_templates
    WHERE template_id = p_template_id
    AND (tenant_id IS NULL OR tenant_id = p_tenant_id);
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Template de notificação % não encontrado', p_template_id;
    END IF;
    
    -- Inicializar com template padrão
    v_subject := v_template.subject_template;
    v_body := v_template.body_template;
    
    -- Substituir variáveis no template
    FOR v_placeholder, v_value IN SELECT * FROM jsonb_each_text(p_alert_data)
    LOOP
        v_subject := REPLACE(v_subject, '{{' || v_placeholder || '}}', COALESCE(v_value, ''));
        v_body := REPLACE(v_body, '{{' || v_placeholder || '}}', COALESCE(v_value, ''));
    END LOOP;
    
    -- Retornar mensagem formatada
    subject := v_subject;
    body := v_body;
    RETURN NEXT;
END;
$$ LANGUAGE plpgsql;

-- Função para enviar notificações para um alerta específico
CREATE OR REPLACE FUNCTION compliance_alert_system.send_alert_notifications(
    p_alert_id UUID
) RETURNS INTEGER AS $$
DECLARE
    v_alert RECORD;
    v_config RECORD;
    v_recipient RECORD;
    v_notification_data JSONB;
    v_notification_message RECORD;
    v_sent_count INTEGER := 0;
    v_tenant_name VARCHAR;
    v_requirement_names TEXT := '';
BEGIN
    -- Obter informações do alerta
    SELECT a.*, t.tenant_name 
    INTO v_alert
    FROM compliance_alert_system.alerts a
    JOIN iam.tenants t ON a.tenant_id = t.tenant_id
    WHERE a.alert_id = p_alert_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Alerta % não encontrado', p_alert_id;
    END IF;
    
    -- Preparar dados comuns para notificação
    v_notification_data := jsonb_build_object(
        'alert_id', v_alert.alert_id,
        'alert_title', v_alert.alert_title,
        'alert_message', v_alert.alert_message,
        'alert_severity', v_alert.alert_severity,
        'open_x_domain', v_alert.open_x_domain,
        'framework', v_alert.framework,
        'tenant_name', v_alert.tenant_name,
        'created_at', v_alert.created_at,
        'economic_impact', v_alert.economic_impact
    );
    
    -- Adicionar nomes dos requisitos (para melhor legibilidade)
    IF v_alert.requirement_ids IS NOT NULL AND array_length(v_alert.requirement_ids, 1) > 0 THEN
        SELECT string_agg(requirement_name, ', ') INTO v_requirement_names
        FROM (
            SELECT r.requirement_name
            FROM unnest(v_alert.requirement_ids) req_id
            JOIN compliance_validators.get_open_x_domain_requirements(v_alert.open_x_domain) r 
                ON req_id = r.requirement_id
        ) subq;
        
        v_notification_data := v_notification_data || jsonb_build_object('requirement_names', v_requirement_names);
    END IF;
    
    -- Buscar configurações de notificação aplicáveis
    FOR v_config IN 
        SELECT nc.*
        FROM compliance_alert_system.notification_configurations nc
        WHERE nc.enabled = TRUE
        AND (nc.tenant_id IS NULL OR nc.tenant_id = v_alert.tenant_id)
        AND (nc.open_x_domain IS NULL OR nc.open_x_domain = v_alert.open_x_domain)
        AND (nc.alert_severity IS NULL OR nc.alert_severity = v_alert.alert_severity)
        AND EXISTS (
            SELECT 1 FROM compliance_alert_system.alert_rules ar
            WHERE ar.rule_id = v_alert.rule_id
            AND (ar.notification_groups IS NULL OR 
                 nc.notification_group = ANY(ar.notification_groups))
        )
    LOOP
        -- Verificar limitação de frequência (throttling)
        IF v_config.throttling_enabled THEN
            -- Verificar se já enviou muitas notificações no período definido
            IF (
                SELECT COUNT(*)
                FROM compliance_alert_system.notifications
                WHERE config_id = v_config.config_id
                AND sent_at > (CURRENT_TIMESTAMP - (v_config.throttling_window_minutes || ' minutes')::INTERVAL)
            ) >= v_config.throttling_max_notifications THEN
                CONTINUE; -- Pular este grupo de notificação devido ao throttling
            END IF;
        END IF;
        
        -- Processar cada destinatário
        FOREACH v_recipient IN ARRAY v_config.recipients
        LOOP
            -- Formatar a notificação usando o template apropriado
            SELECT * INTO v_notification_message
            FROM compliance_alert_system.format_notification(
                COALESCE(v_recipient.notification_template_id, v_config.template_id),
                v_notification_data,
                v_alert.tenant_id
            );
            
            -- Inserir notificação a ser enviada
            INSERT INTO compliance_alert_system.notifications (
                alert_id,
                config_id,
                channel,
                recipient,
                subject,
                content,
                status
            ) VALUES (
                p_alert_id,
                v_config.config_id,
                v_recipient.channel,
                v_recipient.recipient_id,
                v_notification_message.subject,
                v_notification_message.body,
                'PENDENTE'
            );
            
            v_sent_count := v_sent_count + 1;
        END LOOP;
    END LOOP;
    
    RETURN v_sent_count;
END;
$$ LANGUAGE plpgsql;

-- =============================================
-- Procedimentos para Geração de Alertas
-- =============================================

-- Procedimento para gerar alertas automaticamente
CREATE OR REPLACE PROCEDURE compliance_alert_system.generate_alerts(
    p_tenant_id UUID
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_alert_data RECORD;
    v_alert_id UUID;
    v_notifications_sent INTEGER;
BEGIN
    -- Avaliar regras de alerta
    FOR v_alert_data IN 
        SELECT * FROM compliance_alert_system.evaluate_alert_rules(p_tenant_id)
    LOOP
        -- Criar novo alerta na tabela
        INSERT INTO compliance_alert_system.alerts (
            rule_id,
            tenant_id,
            open_x_domain,
            framework,
            requirement_ids,
            alert_title,
            alert_message,
            alert_severity,
            economic_impact,
            details
        ) VALUES (
            v_alert_data.rule_id,
            p_tenant_id,
            v_alert_data.open_x_domain,
            v_alert_data.framework,
            v_alert_data.requirement_ids,
            v_alert_data.alert_title,
            v_alert_data.alert_message,
            v_alert_data.alert_severity,
            v_alert_data.economic_impact,
            v_alert_data.details
        ) RETURNING alert_id INTO v_alert_id;
        
        -- Enviar notificações para o alerta gerado
        v_notifications_sent := compliance_alert_system.send_alert_notifications(v_alert_id);
        
        -- Registrar em log
        RAISE NOTICE 'Alerta % gerado para o domínio %. % notificações enviadas.',
            v_alert_id, v_alert_data.open_x_domain, v_notifications_sent;
    END LOOP;
END;
$$;

-- Procedimento para atualizar métricas de alerta
CREATE OR REPLACE PROCEDURE compliance_alert_system.update_alert_metrics(
    p_tenant_id UUID,
    p_days_back INTEGER DEFAULT 30
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_period_start TIMESTAMP WITH TIME ZONE;
    v_period_end TIMESTAMP WITH TIME ZONE;
    v_domains VARCHAR[];
    v_domain VARCHAR;
BEGIN
    -- Definir período para análise
    v_period_end := CURRENT_TIMESTAMP;
    v_period_start := v_period_end - (p_days_back || ' days')::INTERVAL;
    
    -- Obter todos os domínios Open X com alertas
    SELECT ARRAY_AGG(DISTINCT open_x_domain) INTO v_domains
    FROM compliance_alert_system.alerts
    WHERE tenant_id = p_tenant_id
    AND created_at BETWEEN v_period_start AND v_period_end;
    
    -- Para cada domínio, calcular métricas
    FOREACH v_domain IN ARRAY v_domains
    LOOP
        -- Excluir métricas existentes para o mesmo período e domínio
        DELETE FROM compliance_alert_system.alert_metrics
        WHERE tenant_id = p_tenant_id
        AND period_start = v_period_start
        AND period_end = v_period_end
        AND open_x_domain = v_domain;
        
        -- Inserir novas métricas
        INSERT INTO compliance_alert_system.alert_metrics (
            tenant_id,
            period_start,
            period_end,
            open_x_domain,
            total_alerts,
            critical_alerts,
            high_alerts,
            medium_alerts,
            low_alerts,
            informative_alerts,
            avg_resolution_time_minutes,
            min_resolution_time_minutes,
            max_resolution_time_minutes,
            false_positive_count,
            most_common_rule_id,
            most_common_requirement_id,
            escalated_count,
            sla_compliant_count,
            sla_violated_count
        )
        SELECT
            p_tenant_id,
            v_period_start,
            v_period_end,
            v_domain,
            COUNT(*),
            COUNT(*) FILTER (WHERE alert_severity = 'CRITICO'),
            COUNT(*) FILTER (WHERE alert_severity = 'ALTO'),
            COUNT(*) FILTER (WHERE alert_severity = 'MEDIO'),
            COUNT(*) FILTER (WHERE alert_severity = 'BAIXO'),
            COUNT(*) FILTER (WHERE alert_severity = 'INFORMATIVO'),
            AVG(EXTRACT(EPOCH FROM (resolved_at - created_at)) / 60) FILTER (WHERE resolved_at IS NOT NULL),
            MIN(EXTRACT(EPOCH FROM (resolved_at - created_at)) / 60) FILTER (WHERE resolved_at IS NOT NULL),
            MAX(EXTRACT(EPOCH FROM (resolved_at - created_at)) / 60) FILTER (WHERE resolved_at IS NOT NULL),
            COUNT(*) FILTER (WHERE status = 'FALSO_POSITIVO'),
            (SELECT rule_id FROM (
                SELECT rule_id, COUNT(*) as rule_count
                FROM compliance_alert_system.alerts
                WHERE tenant_id = p_tenant_id
                AND open_x_domain = v_domain
                AND created_at BETWEEN v_period_start AND v_period_end
                GROUP BY rule_id
                ORDER BY rule_count DESC
                LIMIT 1
            ) AS most_common_rule),
            (SELECT req_id FROM (
                SELECT unnest(requirement_ids) as req_id, COUNT(*) as req_count
                FROM compliance_alert_system.alerts
                WHERE tenant_id = p_tenant_id
                AND open_x_domain = v_domain
                AND created_at BETWEEN v_period_start AND v_period_end
                GROUP BY req_id
                ORDER BY req_count DESC
                LIMIT 1
            ) AS most_common_req),
            0, -- escalated_count (a ser implementado)
            0, -- sla_compliant_count (a ser implementado)
            0  -- sla_violated_count (a ser implementado)
        FROM compliance_alert_system.alerts
        WHERE tenant_id = p_tenant_id
        AND open_x_domain = v_domain
        AND created_at BETWEEN v_period_start AND v_period_end;
    END LOOP;
END;
$$;

-- =============================================
-- Regras de Alerta Predefinidas para Open X
-- =============================================

-- Inserir regras de alerta predefinidas para detecção de não-conformidades críticas
INSERT INTO compliance_alert_system.alert_rules (
    rule_id,
    rule_name,
    description,
    enabled,
    open_x_domain,
    framework,
    irr_thresholds,
    alert_severity,
    rule_config,
    created_by,
    notification_groups,
    priority,
    tags
) VALUES
-- Open Insurance - Regra para não-conformidades críticas em Solvência II
(
    'OI_SOLV2_CRIT_NONCOMP',
    'Não-conformidades Críticas Solvência II',
    'Detecta não-conformidades críticas (R3, R4) em requisitos de Solvência II para Open Insurance',
    TRUE,
    'OPEN_INSURANCE',
    'SOLVENCIA_II',
    ARRAY['R3', 'R4'],
    'CRITICO',
    ROW('CRITICAL_NON_COMPLIANCE', 1, '>=', 0, 1, 1440, TRUE)::compliance_alert_system.alert_rule_config,
    'system',
    ARRAY['compliance_team', 'insurance_team', 'risk_team'],
    1,
    ARRAY['open_insurance', 'solvencia_ii', 'regulatorio', 'critico']
),
-- Open Insurance - Regra para não-conformidades críticas em SUSEP
(
    'OI_SUSEP_CRIT_NONCOMP',
    'Não-conformidades Críticas SUSEP',
    'Detecta não-conformidades críticas (R3, R4) em requisitos SUSEP para Open Insurance',
    TRUE,
    'OPEN_INSURANCE',
    'SUSEP',
    ARRAY['R3', 'R4'],
    'ALTO',
    ROW('CRITICAL_NON_COMPLIANCE', 1, '>=', 0, 1, 1440, TRUE)::compliance_alert_system.alert_rule_config,
    'system',
    ARRAY['compliance_team', 'insurance_team', 'risk_team'],
    2,
    ARRAY['open_insurance', 'susep', 'regulatorio', 'brasil']
),
-- Open Health - Regra para não-conformidades críticas em HIPAA/GDPR
(
    'OH_HIPAA_GDPR_CRIT_NONCOMP',
    'Não-conformidades Críticas HIPAA/GDPR',
    'Detecta não-conformidades críticas (R3, R4) em requisitos HIPAA/GDPR para Open Health',
    TRUE,
    'OPEN_HEALTH',
    'HIPAA_GDPR',
    ARRAY['R3', 'R4'],
    'CRITICO',
    ROW('CRITICAL_NON_COMPLIANCE', 1, '>=', 0, 1, 1440, TRUE)::compliance_alert_system.alert_rule_config,
    'system',
    ARRAY['compliance_team', 'health_team', 'risk_team', 'privacy_team'],
    1,
    ARRAY['open_health', 'hipaa', 'gdpr', 'privacidade', 'critico']
),
-- Open Health - Regra para não-conformidades críticas em ANS
(
    'OH_ANS_CRIT_NONCOMP',
    'Não-conformidades Críticas ANS',
    'Detecta não-conformidades críticas (R3, R4) em requisitos ANS para Open Health',
    TRUE,
    'OPEN_HEALTH',
    'ANS',
    ARRAY['R3', 'R4'],
    'ALTO',
    ROW('CRITICAL_NON_COMPLIANCE', 1, '>=', 0, 1, 1440, TRUE)::compliance_alert_system.alert_rule_config,
    'system',
    ARRAY['compliance_team', 'health_team', 'risk_team'],
    2,
    ARRAY['open_health', 'ans', 'regulatorio', 'brasil']
),
-- Open Government - Regra para não-conformidades críticas em eIDAS
(
    'OG_EIDAS_CRIT_NONCOMP',
    'Não-conformidades Críticas eIDAS',
    'Detecta não-conformidades críticas (R3, R4) em requisitos eIDAS para Open Government',
    TRUE,
    'OPEN_GOVERNMENT',
    'EIDAS',
    ARRAY['R3', 'R4'],
    'CRITICO',
    ROW('CRITICAL_NON_COMPLIANCE', 1, '>=', 0, 1, 1440, TRUE)::compliance_alert_system.alert_rule_config,
    'system',
    ARRAY['compliance_team', 'government_team', 'risk_team'],
    1,
    ARRAY['open_government', 'eidas', 'regulatorio', 'ue']
),
-- Open Government - Regra para não-conformidades críticas em Gov.br
(
    'OG_GOVBR_CRIT_NONCOMP',
    'Não-conformidades Críticas Gov.br',
    'Detecta não-conformidades críticas (R3, R4) em requisitos Gov.br para Open Government',
    TRUE,
    'OPEN_GOVERNMENT',
    'GOV_BR',
    ARRAY['R3', 'R4'],
    'ALTO',
    ROW('CRITICAL_NON_COMPLIANCE', 1, '>=', 0, 1, 1440, TRUE)::compliance_alert_system.alert_rule_config,
    'system',
    ARRAY['compliance_team', 'government_team', 'risk_team'],
    2,
    ARRAY['open_government', 'gov_br', 'regulatorio', 'brasil']
);

-- Inserir regras de alerta para deterioração de conformidade
INSERT INTO compliance_alert_system.alert_rules (
    rule_id,
    rule_name,
    description,
    enabled,
    open_x_domain,
    framework,
    alert_severity,
    rule_config,
    created_by,
    notification_groups,
    priority,
    tags
) VALUES
-- Deterioração de conformidade em Open Insurance
(
    'OI_DETERIORATION',
    'Deterioração de Conformidade Open Insurance',
    'Detecta uma queda significativa na conformidade geral de Open Insurance',
    TRUE,
    'OPEN_INSURANCE',
    NULL,
    'MEDIO',
    ROW('DETERIORATION_TREND', 10.0, '>=', 43200, 1, 10080, FALSE)::compliance_alert_system.alert_rule_config,
    'system',
    ARRAY['compliance_team', 'insurance_team', 'management_team'],
    3,
    ARRAY['open_insurance', 'tendencia', 'deterioracao']
),
-- Deterioração de conformidade em Open Health
(
    'OH_DETERIORATION',
    'Deterioração de Conformidade Open Health',
    'Detecta uma queda significativa na conformidade geral de Open Health',
    TRUE,
    'OPEN_HEALTH',
    NULL,
    'MEDIO',
    ROW('DETERIORATION_TREND', 10.0, '>=', 43200, 1, 10080, FALSE)::compliance_alert_system.alert_rule_config,
    'system',
    ARRAY['compliance_team', 'health_team', 'management_team'],
    3,
    ARRAY['open_health', 'tendencia', 'deterioracao']
),
-- Deterioração de conformidade em Open Government
(
    'OG_DETERIORATION',
    'Deterioração de Conformidade Open Government',
    'Detecta uma queda significativa na conformidade geral de Open Government',
    TRUE,
    'OPEN_GOVERNMENT',
    NULL,
    'MEDIO',
    ROW('DETERIORATION_TREND', 10.0, '>=', 43200, 1, 10080, FALSE)::compliance_alert_system.alert_rule_config,
    'system',
    ARRAY['compliance_team', 'government_team', 'management_team'],
    3,
    ARRAY['open_government', 'tendencia', 'deterioracao']
);

-- Inserir templates de notificação padrão
INSERT INTO compliance_alert_system.notification_templates (
    template_id,
    template_name,
    subject_template,
    body_template,
    channel_type,
    format
) VALUES
-- Template para alertas críticos via email
(
    'CRITICAL_ALERT_EMAIL',
    'Alerta Crítico - Email',
    '[CRÍTICO] Alerta de Conformidade Open X: {{alert_title}}',
    'Prezado(a),

Foi detectado um alerta crítico de conformidade no ambiente de {{tenant_name}}.

Domínio: {{open_x_domain}}
Framework: {{framework}}
Descrição: {{alert_message}}
Impacto Econômico Estimado: R$ {{economic_impact}}

Requisitos afetados: {{requirement_names}}

Este alerta requer atenção imediata conforme procedimentos de conformidade da organização.

Para mais detalhes, acesse o Dashboard de Conformidade Open X.

Atenciosamente,
Sistema de Alertas INNOVABIZ',
    'EMAIL',
    'TEXT'
),
-- Template para alertas de alto risco via email
(
    'HIGH_ALERT_EMAIL',
    'Alerta Alto - Email',
    '[ALTO RISCO] Alerta de Conformidade Open X: {{alert_title}}',
    'Prezado(a),

Foi detectado um alerta de alto risco de conformidade no ambiente de {{tenant_name}}.

Domínio: {{open_x_domain}}
Framework: {{framework}}
Descrição: {{alert_message}}
Impacto Econômico Estimado: R$ {{economic_impact}}

Requisitos afetados: {{requirement_names}}

Este alerta requer atenção conforme procedimentos de conformidade da organização.

Para mais detalhes, acesse o Dashboard de Conformidade Open X.

Atenciosamente,
Sistema de Alertas INNOVABIZ',
    'EMAIL',
    'TEXT'
),
-- Template para mensagens do Slack
(
    'SLACK_ALERT',
    'Alerta - Slack',
    NULL,
    '*[{{alert_severity}}] Alerta de Conformidade Open X*\n\n*{{alert_title}}*\n\n{{alert_message}}\n\n*Domínio:* {{open_x_domain}}\n*Framework:* {{framework}}\n*Impacto Econômico:* R$ {{economic_impact}}\n\n<https://dashboard.innovabiz.com/alerts/{{alert_id}}|Ver Detalhes>',
    'SLACK',
    'MARKDOWN'
);

-- Inserir configurações de notificação padrão
INSERT INTO compliance_alert_system.notification_configurations (
    config_id,
    notification_group,
    alert_severity,
    recipients,
    template_id,
    throttling_enabled,
    throttling_window_minutes,
    throttling_max_notifications
) VALUES
-- Configuração para notificações do time de compliance
(
    'COMPLIANCE_TEAM_CONFIG',
    'compliance_team',
    NULL,
    ARRAY[
        ROW('GROUP', 'compliance_group', 'Time de Compliance', 'EMAIL', 'CRITICAL_ALERT_EMAIL')::compliance_alert_system.notification_recipient,
        ROW('GROUP', 'compliance_slack', 'Canal de Compliance', 'SLACK', 'SLACK_ALERT')::compliance_alert_system.notification_recipient
    ],
    NULL,
    TRUE,
    60,
    10
),
-- Configuração para notificações do time de gestão
(
    'MANAGEMENT_TEAM_CONFIG',
    'management_team',
    'CRITICO',
    ARRAY[
        ROW('GROUP', 'management_group', 'Time de Gestão', 'EMAIL', 'CRITICAL_ALERT_EMAIL')::compliance_alert_system.notification_recipient
    ],
    NULL,
    TRUE,
    1440,
    5
);

-- =============================================
-- Triggers e Jobs Agendados
-- =============================================

-- Trigger para registrar histórico de alterações em alertas
CREATE OR REPLACE FUNCTION compliance_alert_system.log_alert_changes()
RETURNS TRIGGER AS $$
DECLARE
    v_column_name VARCHAR;
    v_old_value TEXT;
    v_new_value TEXT;
BEGIN
    IF (TG_OP = 'UPDATE') THEN
        -- Registrar somente campos que foram alterados
        FOREACH v_column_name IN ARRAY TG_ARGV
        LOOP
            EXECUTE format('SELECT $1.%I::TEXT, $2.%I::TEXT', v_column_name, v_column_name)
            INTO v_old_value, v_new_value
            USING OLD, NEW;
            
            IF v_old_value IS DISTINCT FROM v_new_value THEN
                INSERT INTO compliance_alert_system.alert_history (
                    alert_id,
                    field_changed,
                    old_value,
                    new_value,
                    changed_by
                ) VALUES (
                    NEW.alert_id,
                    v_column_name,
                    v_old_value,
                    v_new_value,
                    current_user
                );
            END IF;
        END LOOP;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Criar o trigger para registrar alterações em alertas
CREATE TRIGGER alert_changes_trigger
AFTER UPDATE ON compliance_alert_system.alerts
FOR EACH ROW
EXECUTE FUNCTION compliance_alert_system.log_alert_changes(
    'status', 'acknowledged_by', 'resolution_notes', 'economic_impact'
);

-- =============================================
-- Configuração de Permissões
-- =============================================

-- Conceder permissões para os roles necessários
DO $$
BEGIN
    -- Verificar se os roles existem antes de conceder permissões
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'compliance_manager_role') THEN
        -- Permissões para gestores de conformidade
        GRANT USAGE ON SCHEMA compliance_alert_system TO compliance_manager_role;
        GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA compliance_alert_system TO compliance_manager_role;
        GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA compliance_alert_system TO compliance_manager_role;
    END IF;
    
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'risk_analyst_role') THEN
        -- Permissões para analistas de risco
        GRANT USAGE ON SCHEMA compliance_alert_system TO risk_analyst_role;
        GRANT SELECT ON ALL TABLES IN SCHEMA compliance_alert_system TO risk_analyst_role;
        GRANT UPDATE ON compliance_alert_system.alerts TO risk_analyst_role;
        GRANT EXECUTE ON FUNCTION compliance_alert_system.evaluate_alert_rules TO risk_analyst_role;
    END IF;
    
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'dashboard_viewer_role') THEN
        -- Permissões para visualizadores do dashboard
        GRANT USAGE ON SCHEMA compliance_alert_system TO dashboard_viewer_role;
        GRANT SELECT ON compliance_alert_system.alerts TO dashboard_viewer_role;
        GRANT SELECT ON compliance_alert_system.alert_metrics TO dashboard_viewer_role;
        GRANT SELECT ON compliance_alert_system.alert_rules TO dashboard_viewer_role;
    END IF;
END $$;

-- =============================================
-- Notificar Conclusão
-- =============================================

DO $$
BEGIN
    RAISE NOTICE 'Sistema de Alertas Inteligentes para Open X configurado com sucesso!';
    RAISE NOTICE 'Componentes configurados:';
    RAISE NOTICE '- Estrutura de dados para alertas e notificações';
    RAISE NOTICE '- Funções para detecção de alertas';
    RAISE NOTICE '- Processamento de notificações';
    RAISE NOTICE '- Regras de alerta predefinidas';
    RAISE NOTICE '- Templates de notificação';
END $$;
