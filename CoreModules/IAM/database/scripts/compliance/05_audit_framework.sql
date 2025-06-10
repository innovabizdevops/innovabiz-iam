-- INNOVABIZ - IAM Audit Framework
-- Author: Eduardo Jeremias
-- Date: 09/05/2025
-- Version: 1.0
-- Description: Framework abrangente de auditoria para IAM com suporte a requisitos internacionais

-- Configuração do esquema e extensões
SET search_path TO iam, public;

-- Criação de tipos enumerados para categorização padronizada
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'audit_severity_level') THEN
        CREATE TYPE iam.audit_severity_level AS ENUM (
            'critical', 
            'high', 
            'medium', 
            'low', 
            'info'
        );
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'audit_event_category') THEN
        CREATE TYPE iam.audit_event_category AS ENUM (
            'authentication', 
            'authorization', 
            'identity_management', 
            'access_management',
            'federation',
            'policy_change',
            'system_configuration',
            'compliance',
            'security',
            'data_access',
            'health_check'
        );
    END IF;
END$$;

-- Tabela principal de logs de auditoria detalhados
CREATE TABLE IF NOT EXISTS iam.detailed_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES iam.organizations(id),
    user_id UUID REFERENCES iam.users(id),
    event_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    event_category iam.audit_event_category NOT NULL,
    severity_level iam.audit_severity_level NOT NULL DEFAULT 'info',
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),
    source_ip VARCHAR(45),
    user_agent TEXT,
    request_id UUID,
    session_id UUID,
    status VARCHAR(50) NOT NULL,
    response_time INTEGER, -- Em milissegundos
    details JSONB NOT NULL DEFAULT '{}'::JSONB,
    request_payload JSONB,
    response_payload JSONB,
    compliance_tags VARCHAR[] DEFAULT ARRAY[]::VARCHAR[],
    regulatory_references VARCHAR[] DEFAULT ARRAY[]::VARCHAR[],
    geo_location JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Índices para pesquisa e performance
CREATE INDEX IF NOT EXISTS idx_detailed_audit_logs_org_id ON iam.detailed_audit_logs(organization_id);
CREATE INDEX IF NOT EXISTS idx_detailed_audit_logs_user_id ON iam.detailed_audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_detailed_audit_logs_event_time ON iam.detailed_audit_logs(event_time);
CREATE INDEX IF NOT EXISTS idx_detailed_audit_logs_event_category ON iam.detailed_audit_logs(event_category);
CREATE INDEX IF NOT EXISTS idx_detailed_audit_logs_severity ON iam.detailed_audit_logs(severity_level);
CREATE INDEX IF NOT EXISTS idx_detailed_audit_logs_action ON iam.detailed_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_detailed_audit_logs_status ON iam.detailed_audit_logs(status);
CREATE INDEX IF NOT EXISTS idx_detailed_audit_logs_compliance ON iam.detailed_audit_logs USING GIN(compliance_tags);

-- Visão para relatórios de compliance GDPR
CREATE OR REPLACE VIEW iam.gdpr_audit_view AS
SELECT 
    dal.id,
    dal.organization_id,
    o.name AS organization_name,
    dal.user_id,
    u.full_name AS user_name,
    u.email AS user_email,
    dal.event_time,
    dal.event_category,
    dal.severity_level,
    dal.action,
    dal.resource_type,
    dal.resource_id,
    dal.source_ip,
    dal.status,
    dal.details,
    dal.compliance_tags,
    dal.regulatory_references
FROM 
    iam.detailed_audit_logs dal
LEFT JOIN 
    iam.organizations o ON dal.organization_id = o.id
LEFT JOIN 
    iam.users u ON dal.user_id = u.id
WHERE 
    'gdpr' = ANY(dal.compliance_tags)
ORDER BY 
    dal.event_time DESC;

-- Visão para relatórios de compliance LGPD
CREATE OR REPLACE VIEW iam.lgpd_audit_view AS
SELECT 
    dal.id,
    dal.organization_id,
    o.name AS organization_name,
    dal.user_id,
    u.full_name AS user_name,
    u.email AS user_email,
    dal.event_time,
    dal.event_category,
    dal.severity_level,
    dal.action,
    dal.resource_type,
    dal.resource_id,
    dal.source_ip,
    dal.status,
    dal.details,
    dal.compliance_tags,
    dal.regulatory_references
FROM 
    iam.detailed_audit_logs dal
LEFT JOIN 
    iam.organizations o ON dal.organization_id = o.id
LEFT JOIN 
    iam.users u ON dal.user_id = u.id
WHERE 
    'lgpd' = ANY(dal.compliance_tags)
ORDER BY 
    dal.event_time DESC;

-- Visão para relatórios de compliance HIPAA
CREATE OR REPLACE VIEW iam.hipaa_audit_view AS
SELECT 
    dal.id,
    dal.organization_id,
    o.name AS organization_name,
    dal.user_id,
    u.full_name AS user_name,
    u.email AS user_email,
    dal.event_time,
    dal.event_category,
    dal.severity_level,
    dal.action,
    dal.resource_type,
    dal.resource_id,
    dal.source_ip,
    dal.status,
    dal.details,
    dal.compliance_tags,
    dal.regulatory_references
FROM 
    iam.detailed_audit_logs dal
LEFT JOIN 
    iam.organizations o ON dal.organization_id = o.id
LEFT JOIN 
    iam.users u ON dal.user_id = u.id
WHERE 
    'hipaa' = ANY(dal.compliance_tags)
ORDER BY 
    dal.event_time DESC;

-- Função para registrar eventos de auditoria
CREATE OR REPLACE FUNCTION iam.log_audit_event(
    p_organization_id UUID,
    p_user_id UUID,
    p_event_category iam.audit_event_category,
    p_severity_level iam.audit_severity_level,
    p_action VARCHAR,
    p_resource_type VARCHAR,
    p_resource_id VARCHAR,
    p_source_ip VARCHAR,
    p_user_agent TEXT,
    p_request_id UUID,
    p_session_id UUID,
    p_status VARCHAR,
    p_response_time INTEGER,
    p_details JSONB,
    p_request_payload JSONB DEFAULT NULL,
    p_response_payload JSONB DEFAULT NULL,
    p_compliance_tags VARCHAR[] DEFAULT NULL,
    p_regulatory_references VARCHAR[] DEFAULT NULL,
    p_geo_location JSONB DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    audit_id UUID;
BEGIN
    INSERT INTO iam.detailed_audit_logs (
        organization_id,
        user_id,
        event_category,
        severity_level,
        action,
        resource_type,
        resource_id,
        source_ip,
        user_agent,
        request_id,
        session_id,
        status,
        response_time,
        details,
        request_payload,
        response_payload,
        compliance_tags,
        regulatory_references,
        geo_location
    ) VALUES (
        p_organization_id,
        p_user_id,
        p_event_category,
        p_severity_level,
        p_action,
        p_resource_type,
        p_resource_id,
        p_source_ip,
        p_user_agent,
        p_request_id,
        p_session_id,
        p_status,
        p_response_time,
        p_details,
        p_request_payload,
        p_response_payload,
        COALESCE(p_compliance_tags, ARRAY[]::VARCHAR[]),
        COALESCE(p_regulatory_references, ARRAY[]::VARCHAR[]),
        p_geo_location
    ) RETURNING id INTO audit_id;
    
    RETURN audit_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Trigger para log automático em alterações de entidades críticas
CREATE OR REPLACE FUNCTION iam.audit_entity_changes()
RETURNS TRIGGER AS $$
DECLARE
    change_type VARCHAR;
    entity_type VARCHAR;
    entity_id VARCHAR;
    details JSONB;
    user_id UUID;
BEGIN
    -- Determinar tipo de operação
    IF TG_OP = 'INSERT' THEN
        change_type := 'CREATE';
        details := to_jsonb(NEW);
    ELSIF TG_OP = 'UPDATE' THEN
        change_type := 'UPDATE';
        details := jsonb_build_object(
            'old', to_jsonb(OLD),
            'new', to_jsonb(NEW),
            'changed_fields', (
                SELECT jsonb_object_agg(key, value)
                FROM jsonb_each(to_jsonb(NEW))
                WHERE to_jsonb(NEW) -> key <> to_jsonb(OLD) -> key
            )
        );
    ELSIF TG_OP = 'DELETE' THEN
        change_type := 'DELETE';
        details := to_jsonb(OLD);
    END IF;
    
    -- Definir entidade
    entity_type := TG_TABLE_NAME;
    
    -- Obter ID da entidade
    IF TG_OP = 'DELETE' THEN
        entity_id := OLD.id::TEXT;
    ELSE
        entity_id := NEW.id::TEXT;
    END IF;
    
    -- Obter usuário atual
    BEGIN
        user_id := current_setting('app.current_user_id', true)::UUID;
    EXCEPTION WHEN OTHERS THEN
        user_id := NULL;
    END;
    
    -- Determinar organization_id
    DECLARE
        org_id UUID;
    BEGIN
        IF TG_OP = 'DELETE' THEN
            IF OLD.organization_id IS NOT NULL THEN
                org_id := OLD.organization_id;
            END IF;
        ELSE
            IF NEW.organization_id IS NOT NULL THEN
                org_id := NEW.organization_id;
            END IF;
        END IF;
    END;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        org_id,
        user_id,
        'identity_management'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        change_type || '_' || entity_type,
        entity_type,
        entity_id,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        'success',
        NULL, -- response_time
        details,
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['identity_management'], -- compliance_tags
        NULL, -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Aplicação do trigger nas tabelas principais
DROP TRIGGER IF EXISTS audit_users_changes ON iam.users;
CREATE TRIGGER audit_users_changes
AFTER INSERT OR UPDATE OR DELETE ON iam.users
FOR EACH ROW EXECUTE FUNCTION iam.audit_entity_changes();

DROP TRIGGER IF EXISTS audit_roles_changes ON iam.roles;
CREATE TRIGGER audit_roles_changes
AFTER INSERT OR UPDATE OR DELETE ON iam.roles
FOR EACH ROW EXECUTE FUNCTION iam.audit_entity_changes();

DROP TRIGGER IF EXISTS audit_permissions_changes ON iam.permissions;
CREATE TRIGGER audit_permissions_changes
AFTER INSERT OR UPDATE OR DELETE ON iam.permissions
FOR EACH ROW EXECUTE FUNCTION iam.audit_entity_changes();

DROP TRIGGER IF EXISTS audit_user_role_assignments_changes ON iam.user_role_assignments;
CREATE TRIGGER audit_user_role_assignments_changes
AFTER INSERT OR UPDATE OR DELETE ON iam.user_role_assignments
FOR EACH ROW EXECUTE FUNCTION iam.audit_entity_changes();

-- Função para exportar logs de auditoria em formatos específicos para compliance
CREATE OR REPLACE FUNCTION iam.export_audit_logs_for_compliance(
    p_organization_id UUID,
    p_start_date TIMESTAMP WITH TIME ZONE,
    p_end_date TIMESTAMP WITH TIME ZONE,
    p_compliance_standard VARCHAR, -- 'gdpr', 'lgpd', 'hipaa', 'pci-dss', etc.
    p_format VARCHAR DEFAULT 'json' -- 'json', 'csv', 'xml'
) RETURNS TEXT AS $$
DECLARE
    result TEXT;
    query_text TEXT;
BEGIN
    -- Construir consulta base
    query_text := '
        SELECT 
            id,
            organization_id,
            user_id,
            event_time,
            event_category,
            severity_level,
            action,
            resource_type,
            resource_id,
            source_ip,
            status,
            details,
            compliance_tags,
            regulatory_references
        FROM 
            iam.detailed_audit_logs
        WHERE 
            organization_id = ' || quote_literal(p_organization_id) || '
            AND event_time BETWEEN ' || quote_literal(p_start_date) || ' AND ' || quote_literal(p_end_date) || '
            AND ' || quote_literal(p_compliance_standard) || ' = ANY(compliance_tags)
        ORDER BY 
            event_time DESC
    ';
    
    -- Formatar resultado de acordo com o formato solicitado
    IF p_format = 'json' THEN
        EXECUTE 'SELECT array_to_json(array_agg(row_to_json(t)))::TEXT FROM (' || query_text || ') t' INTO result;
    ELSIF p_format = 'csv' THEN
        -- Implementação para CSV omitida por brevidade
        result := 'CSV format generation implemented in application layer';
    ELSIF p_format = 'xml' THEN
        -- Implementação para XML omitida por brevidade
        result := 'XML format generation implemented in application layer';
    ELSE
        RAISE EXCEPTION 'Formato não suportado: %', p_format;
    END IF;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para análise de anomalias em logs de auditoria
CREATE OR REPLACE FUNCTION iam.detect_audit_anomalies(
    p_organization_id UUID,
    p_lookback_period INTERVAL DEFAULT '24 hours'::INTERVAL,
    p_severity_threshold iam.audit_severity_level DEFAULT 'medium'::iam.audit_severity_level
) RETURNS TABLE (
    anomaly_type TEXT,
    description TEXT,
    severity iam.audit_severity_level,
    affected_users UUID[],
    affected_resources TEXT[],
    details JSONB
) AS $$
BEGIN
    -- Detecção de atividade incomum de login
    RETURN QUERY
    WITH login_stats AS (
        SELECT 
            user_id,
            COUNT(*) AS login_count,
            COUNT(DISTINCT source_ip) AS distinct_ips,
            array_agg(DISTINCT source_ip) AS ip_list
        FROM 
            iam.detailed_audit_logs
        WHERE 
            organization_id = p_organization_id
            AND event_time > NOW() - p_lookback_period
            AND event_category = 'authentication'
            AND action = 'LOGIN'
        GROUP BY 
            user_id
    ),
    user_averages AS (
        SELECT 
            AVG(login_count) AS avg_logins,
            percentile_cont(0.95) WITHIN GROUP (ORDER BY login_count) AS p95_logins,
            percentile_cont(0.95) WITHIN GROUP (ORDER BY distinct_ips) AS p95_ips
        FROM 
            login_stats
    )
    SELECT 
        'unusual_login_activity' AS anomaly_type,
        'Atividade de login incomum detectada' AS description,
        CASE 
            WHEN s.login_count > 2 * a.p95_logins THEN 'high'::iam.audit_severity_level
            WHEN s.login_count > a.p95_logins THEN 'medium'::iam.audit_severity_level
            ELSE 'low'::iam.audit_severity_level
        END AS severity,
        ARRAY[s.user_id] AS affected_users,
        ARRAY[]::TEXT[] AS affected_resources,
        jsonb_build_object(
            'login_count', s.login_count,
            'avg_user_logins', a.avg_logins,
            'p95_user_logins', a.p95_logins,
            'distinct_ips', s.distinct_ips,
            'p95_distinct_ips', a.p95_ips,
            'ip_addresses', s.ip_list
        ) AS details
    FROM 
        login_stats s,
        user_averages a
    WHERE 
        (s.login_count > a.p95_logins OR s.distinct_ips > a.p95_ips)
        AND CASE 
            WHEN s.login_count > 2 * a.p95_logins THEN 'high'::iam.audit_severity_level
            WHEN s.login_count > a.p95_logins THEN 'medium'::iam.audit_severity_level
            ELSE 'low'::iam.audit_severity_level
        END >= p_severity_threshold;
    
    -- Outras detecções de anomalias podem ser adicionadas aqui
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
