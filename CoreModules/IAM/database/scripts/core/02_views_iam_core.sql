-- INNOVABIZ - IAM Core Views
-- Author: Eduardo Jeremias
-- Date: 08/05/2025
-- Version: 1.0
-- Description: Views para o módulo de IAM core, fornecendo visualizações para relatórios e dashboard.

-- Configurar caminho de busca
SET search_path TO iam, public;

-- View de Permissões de Usuário
CREATE OR REPLACE VIEW iam.vw_user_permissions AS
SELECT
    u.id AS user_id,
    u.username,
    u.email,
    u.organization_id,
    o.name AS organization_name,
    r.id AS role_id,
    r.name AS role_name,
    p.id AS permission_id,
    p.code AS permission_code,
    p.name AS permission_name,
    p.resource,
    p.action,
    rp.assigned_at AS permission_assigned_at -- Adicionado para informação extra, se útil
FROM
    iam.users u
JOIN
    iam.organizations o ON u.organization_id = o.id
JOIN
    iam.user_roles ur ON u.id = ur.user_id
JOIN
    iam.roles r ON ur.role_id = r.id AND r.organization_id = u.organization_id -- Garante que a role é da mesma organização do usuário
JOIN
    iam.role_permissions rp ON r.id = rp.role_id AND rp.organization_id = r.organization_id -- Junta com a nova tabela de atribuição de permissões à role
JOIN
    iam.permissions p ON rp.permission_id = p.id -- Junta com a tabela de permissões para obter detalhes da permissão
WHERE
    u.status = 'active'
    AND ur.is_active = TRUE
    AND (ur.expires_at IS NULL OR ur.expires_at > NOW());

COMMENT ON VIEW vw_user_permissions IS 'Visão que mostra todas as permissões de cada usuário ativo no sistema';

-- View de Status de Usuários por Organização
CREATE OR REPLACE VIEW vw_user_status_by_organization AS
SELECT 
    o.id AS organization_id,
    o.name AS organization_name,
    o.industry,
    o.sector,
    o.country_code,
    o.region_code,
    COUNT(u.id) AS total_users,
    SUM(CASE WHEN u.status = 'active' THEN 1 ELSE 0 END) AS active_users,
    SUM(CASE WHEN u.status = 'inactive' THEN 1 ELSE 0 END) AS inactive_users,
    SUM(CASE WHEN u.status = 'suspended' THEN 1 ELSE 0 END) AS suspended_users,
    SUM(CASE WHEN u.status = 'locked' THEN 1 ELSE 0 END) AS locked_users,
    ROUND((SUM(CASE WHEN u.status = 'active' THEN 1 ELSE 0 END)::NUMERIC / 
           NULLIF(COUNT(u.id), 0)::NUMERIC) * 100, 2) AS active_percentage
FROM 
    organizations o
LEFT JOIN 
    users u ON o.id = u.organization_id
GROUP BY 
    o.id, o.name, o.industry, o.sector, o.country_code, o.region_code;

COMMENT ON VIEW vw_user_status_by_organization IS 'Visão que mostra métricas de status de usuários por organização';

-- View de Atividade de Sessões
CREATE OR REPLACE VIEW vw_session_activity AS
SELECT 
    u.organization_id,
    o.name AS organization_name,
    u.id AS user_id,
    u.username,
    u.email,
    s.id AS session_id,
    s.ip_address,
    s.user_agent,
    s.created_at AS session_start,
    s.last_activity,
    s.expires_at,
    EXTRACT(EPOCH FROM (s.last_activity - s.created_at))/3600 AS session_duration_hours,
    CASE 
        WHEN s.expires_at > NOW() AND s.is_active = true THEN 'active' 
        ELSE 'inactive' 
    END AS session_status
FROM 
    sessions s
JOIN 
    users u ON s.user_id = u.id
JOIN 
    organizations o ON u.organization_id = o.id
WHERE 
    s.created_at > NOW() - INTERVAL '30 days';

COMMENT ON VIEW vw_session_activity IS 'Visão que mostra a atividade de sessões dos últimos 30 dias';

-- View de Resumo de Audit Log
CREATE OR REPLACE VIEW vw_audit_log_summary AS
SELECT 
    organization_id,
    DATE_TRUNC('day', timestamp) AS log_date,
    action,
    resource_type,
    status,
    COUNT(*) AS event_count
FROM 
    audit_logs
WHERE 
    timestamp > NOW() - INTERVAL '90 days'
GROUP BY 
    organization_id, DATE_TRUNC('day', timestamp), action, resource_type, status
ORDER BY 
    log_date DESC, event_count DESC;

COMMENT ON VIEW vw_audit_log_summary IS 'Resumo diário de eventos de auditoria dos últimos 90 dias';

-- View de Políticas de Segurança por Organização
CREATE OR REPLACE VIEW vw_security_policies_by_organization AS
SELECT 
    o.id AS organization_id,
    o.name AS organization_name,
    o.industry,
    o.sector,
    o.country_code,
    o.region_code,
    sp.id AS policy_id,
    sp.name AS policy_name,
    sp.policy_type,
    sp.is_active,
    sp.created_at,
    sp.updated_at,
    u1.username AS created_by_user,
    u2.username AS updated_by_user,
    sp.settings
FROM 
    organizations o
JOIN 
    security_policies sp ON o.id = sp.organization_id
LEFT JOIN 
    users u1 ON sp.created_by = u1.id
LEFT JOIN 
    users u2 ON sp.updated_by = u2.id;

COMMENT ON VIEW vw_security_policies_by_organization IS 'Visão de políticas de segurança configuradas por organização';

-- View de Estatísticas de Uso de Roles
CREATE OR REPLACE VIEW vw_role_usage_stats AS
SELECT 
    r.organization_id,
    o.name AS organization_name,
    r.id AS role_id,
    r.name AS role_name,
    r.is_system_role,
    COUNT(DISTINCT ur.user_id) AS assigned_users_count,
    MIN(ur.granted_at) AS first_assignment,
    MAX(ur.granted_at) AS last_assignment,
    COUNT(DISTINCT rp.permission_id) AS permissions_count
FROM 
    iam.roles r
JOIN 
    iam.organizations o ON r.organization_id = o.id
LEFT JOIN 
    iam.user_roles ur ON r.id = ur.role_id AND ur.is_active = TRUE
LEFT JOIN
    iam.role_permissions rp ON r.id = rp.role_id AND rp.organization_id = r.organization_id -- Junta para contar as permissões da role
GROUP BY 
    r.organization_id, o.name, r.id, r.name, r.is_system_role;

COMMENT ON VIEW vw_role_usage_stats IS 'Estatísticas de uso de roles por organização';

-- View de Resumo de Frameworks Regulatórios
CREATE OR REPLACE VIEW vw_regulatory_frameworks_summary AS
SELECT 
    rf.id,
    rf.code,
    rf.name,
    rf.region,
    rf.sector,
    rf.version,
    rf.effective_date,
    rf.is_active,
    COUNT(cv.id) AS validators_count
FROM 
    regulatory_frameworks rf
LEFT JOIN 
    compliance_validators cv ON rf.id = cv.framework_id
GROUP BY 
    rf.id, rf.code, rf.name, rf.region, rf.sector, rf.version, rf.effective_date, rf.is_active
ORDER BY 
    rf.is_active DESC, rf.name ASC;

COMMENT ON VIEW vw_regulatory_frameworks_summary IS 'Resumo dos frameworks regulatórios e contagem de validadores associados';

-- View de Validadores de Compliance
CREATE OR REPLACE VIEW vw_compliance_validators_details AS
SELECT 
    cv.id AS validator_id,
    cv.code AS validator_code,
    cv.name AS validator_name,
    cv.description,
    cv.validator_class,
    cv.version AS validator_version,
    cv.is_active,
    rf.id AS framework_id,
    rf.code AS framework_code,
    rf.name AS framework_name,
    rf.region,
    rf.sector
FROM 
    compliance_validators cv
JOIN 
    regulatory_frameworks rf ON cv.framework_id = rf.id
ORDER BY 
    cv.is_active DESC, rf.sector, rf.name, cv.name;

COMMENT ON VIEW vw_compliance_validators_details IS 'Detalhes dos validadores de compliance configurados no sistema';
