-- INNOVABIZ - IAM Core Functions
-- Author: Eduardo Jeremias
-- Date: 08/05/2025
-- Version: 1.0
-- Description: Funções para o módulo de IAM core, implementando lógica de negócio e facilitando operações comuns.

-- Configurar caminho de busca
SET search_path TO iam, public;

-- Função para verificar permissões de um usuário
CREATE OR REPLACE FUNCTION check_user_permission(
    p_user_id UUID,
    p_resource VARCHAR,
    p_action VARCHAR
) RETURNS BOOLEAN AS $$
DECLARE
    has_permission BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1
        FROM vw_user_permissions
        WHERE user_id = p_user_id
          AND resource = p_resource
          AND action = p_action
    ) INTO has_permission;
    
    RETURN has_permission;
END;
$$ LANGUAGE plpgsql SECURITY INVOKER; -- ALTERADO PARA SECURITY INVOKER

COMMENT ON FUNCTION check_user_permission IS 'Verifica se um usuário tem uma permissão específica';

-- Função para registrar um evento de auditoria
CREATE OR REPLACE FUNCTION register_audit_log(
    p_organization_id UUID,
    p_user_id UUID,
    p_action VARCHAR,
    p_resource_type VARCHAR,
    p_resource_id VARCHAR,
    p_status VARCHAR,
    p_details JSONB,
    p_ip_address VARCHAR DEFAULT NULL,
    p_session_id UUID DEFAULT NULL,
    p_request_id VARCHAR DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    v_audit_id UUID;
BEGIN
    INSERT INTO audit_logs (
        organization_id,
        user_id,
        action,
        resource_type,
        resource_id,
        status,
        details,
        ip_address,
        session_id,
        request_id
    ) VALUES (
        p_organization_id,
        p_user_id,
        p_action,
        p_resource_type,
        p_resource_id,
        p_status,
        p_details,
        p_ip_address,
        p_session_id,
        p_request_id
    ) RETURNING id INTO v_audit_id;
    
    RETURN v_audit_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION register_audit_log IS 'Registra um evento de auditoria no sistema';

-- Função para criar um novo usuário com role padrão
CREATE OR REPLACE FUNCTION create_user_with_default_role(
    p_organization_id UUID,
    p_username VARCHAR,
    p_email VARCHAR,
    p_full_name VARCHAR,
    p_password_hash VARCHAR,
    p_default_role_name VARCHAR DEFAULT 'user',
    p_created_by UUID DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    v_user_id UUID;
    v_role_id UUID;
BEGIN
    -- Verificar se o usuário já existe
    IF EXISTS (SELECT 1 FROM users WHERE username = p_username OR email = p_email) THEN
        RAISE EXCEPTION 'Usuário com este username ou email já existe';
    END IF;
    
    -- Criar o usuário
    INSERT INTO users (
        organization_id,
        username,
        email,
        full_name,
        password_hash,
        status
    ) VALUES (
        p_organization_id,
        p_username,
        p_email,
        p_full_name,
        p_password_hash,
        'active'
    ) RETURNING id INTO v_user_id;
    
    -- Buscar a role padrão para a organização
    SELECT id INTO v_role_id
    FROM roles
    WHERE organization_id = p_organization_id AND name = p_default_role_name;
    
    -- Se a role não existir, criar
    IF v_role_id IS NULL THEN
        INSERT INTO roles (
            organization_id,
            name,
            description,
            is_system_role
        ) VALUES (
            p_organization_id,
            p_default_role_name,
            'Role padrão para novos usuários',
            TRUE
        ) RETURNING id INTO v_role_id;
    END IF;
    
    -- Atribuir a role ao usuário
    INSERT INTO user_roles (
        user_id,
        role_id,
        granted_by
    ) VALUES (
        v_user_id,
        v_role_id,
        p_created_by
    );
    
    RETURN v_user_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION create_user_with_default_role IS 'Cria um novo usuário e atribui uma role padrão';

-- Função para atualizar o status de um usuário
CREATE OR REPLACE FUNCTION update_user_status(
    p_user_id UUID,
    p_status VARCHAR,
    p_updated_by UUID,
    p_reason JSONB DEFAULT NULL
) RETURNS BOOLEAN AS $$
DECLARE
    v_old_status VARCHAR;
    v_organization_id UUID;
BEGIN
    -- Verificar se o usuário existe
    SELECT status, organization_id INTO v_old_status, v_organization_id
    FROM users
    WHERE id = p_user_id;
    
    IF v_old_status IS NULL THEN
        RAISE EXCEPTION 'Usuário não encontrado';
    END IF;
    
    -- Verificar se o status é válido
    IF p_status NOT IN ('active', 'inactive', 'suspended', 'locked') THEN
        RAISE EXCEPTION 'Status inválido. Valores permitidos: active, inactive, suspended, locked';
    END IF;
    
    -- Atualizar o status
    UPDATE users
    SET status = p_status
    WHERE id = p_user_id;
    
    -- Registrar auditoria
    PERFORM register_audit_log(
        v_organization_id,
        p_updated_by,
        'update_user_status',
        'user',
        p_user_id::VARCHAR,
        'success',
        jsonb_build_object(
            'old_status', v_old_status,
            'new_status', p_status,
            'reason', p_reason
        )
    );
    
    -- Se estiver desativando o usuário, encerrar sessões ativas
    IF p_status IN ('inactive', 'suspended', 'locked') THEN
        UPDATE sessions
        SET is_active = FALSE
        WHERE user_id = p_user_id AND is_active = TRUE;
    END IF;
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_user_status IS 'Atualiza o status de um usuário e realiza ações associadas';

-- Função para atribuir uma permissão a uma role
CREATE OR REPLACE FUNCTION add_permission_to_role(
    p_role_id UUID,
    p_permission_code VARCHAR,
    p_granted_by UUID
) RETURNS BOOLEAN AS $$
DECLARE
    v_organization_id UUID;
    v_current_permissions JSONB;
BEGIN
    -- Verificar se a role existe
    SELECT organization_id, permissions INTO v_organization_id, v_current_permissions
    FROM roles
    WHERE id = p_role_id;
    
    IF v_organization_id IS NULL THEN
        RAISE EXCEPTION 'Role não encontrada';
    END IF;
    
    -- Verificar se a permissão existe
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE code = p_permission_code) THEN
        RAISE EXCEPTION 'Permissão não encontrada: %', p_permission_code;
    END IF;
    
    -- Verificar se a permissão já está atribuída
    IF v_current_permissions ? p_permission_code THEN
        RETURN FALSE; -- Permissão já atribuída
    END IF;
    
    -- Adicionar a permissão
    UPDATE roles
    SET permissions = permissions || jsonb_build_array(p_permission_code)
    WHERE id = p_role_id;
    
    -- Registrar auditoria
    PERFORM register_audit_log(
        v_organization_id,
        p_granted_by,
        'add_permission_to_role',
        'role',
        p_role_id::VARCHAR,
        'success',
        jsonb_build_object(
            'permission_code', p_permission_code
        )
    );
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION add_permission_to_role IS 'Adiciona uma permissão a uma role';

-- Função para remover uma permissão de uma role
CREATE OR REPLACE FUNCTION remove_permission_from_role(
    p_role_id UUID,
    p_permission_code VARCHAR,
    p_removed_by UUID
) RETURNS BOOLEAN AS $$
DECLARE
    v_organization_id UUID;
    v_current_permissions JSONB;
    v_new_permissions JSONB;
BEGIN
    -- Verificar se a role existe
    SELECT organization_id, permissions INTO v_organization_id, v_current_permissions
    FROM roles
    WHERE id = p_role_id;
    
    IF v_organization_id IS NULL THEN
        RAISE EXCEPTION 'Role não encontrada';
    END IF;
    
    -- Verificar se a permissão está atribuída
    IF NOT (v_current_permissions ? p_permission_code) THEN
        RETURN FALSE; -- Permissão não está atribuída
    END IF;
    
    -- Remover a permissão
    WITH elements AS (
      SELECT jsonb_array_elements_text(v_current_permissions) AS permission
    )
    SELECT jsonb_agg(permission)
    INTO v_new_permissions
    FROM elements
    WHERE permission <> p_permission_code;

    -- Atualizar a role
    UPDATE roles
    SET permissions = COALESCE(v_new_permissions, '[]'::JSONB)
    WHERE id = p_role_id;
    
    -- Registrar auditoria
    PERFORM register_audit_log(
        v_organization_id,
        p_removed_by,
        'remove_permission_from_role',
        'role',
        p_role_id::VARCHAR,
        'success',
        jsonb_build_object(
            'permission_code', p_permission_code
        )
    );
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION remove_permission_from_role IS 'Remove uma permissão de uma role';

-- Função para criar uma nova sessão
CREATE OR REPLACE FUNCTION create_session(
    p_user_id UUID,
    p_token VARCHAR,
    p_ip_address VARCHAR DEFAULT NULL,
    p_user_agent TEXT DEFAULT NULL,
    p_expires_in_hours INT DEFAULT 24
) RETURNS UUID AS $$
DECLARE
    v_session_id UUID;
BEGIN
    -- Verificar se o usuário existe e está ativo
    IF NOT EXISTS (SELECT 1 FROM users WHERE id = p_user_id AND status = 'active') THEN
        RAISE EXCEPTION 'Usuário não encontrado ou inativo';
    END IF;
    
    -- Criar a sessão
    INSERT INTO sessions (
        user_id,
        token,
        ip_address,
        user_agent,
        expires_at
    ) VALUES (
        p_user_id,
        p_token,
        p_ip_address,
        p_user_agent,
        NOW() + (p_expires_in_hours || ' hours')::INTERVAL
    ) RETURNING id INTO v_session_id;
    
    -- Atualizar último login do usuário
    UPDATE users
    SET last_login = NOW()
    WHERE id = p_user_id;
    
    RETURN v_session_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION create_session IS 'Cria uma nova sessão para um usuário';

-- Função para verificar uma sessão
CREATE OR REPLACE FUNCTION verify_session(
    p_token VARCHAR
) RETURNS TABLE(
    session_id UUID,
    user_id UUID,
    username VARCHAR,
    organization_id UUID,
    organization_name VARCHAR,
    is_valid BOOLEAN,
    reason VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        s.id AS session_id,
        u.id AS user_id,
        u.username,
        u.organization_id,
        o.name AS organization_name,
        CASE
            WHEN s.is_active = FALSE THEN FALSE
            WHEN s.expires_at < NOW() THEN FALSE
            WHEN u.status <> 'active' THEN FALSE
            ELSE TRUE
        END AS is_valid,
        CASE
            WHEN s.is_active = FALSE THEN 'Sessão foi encerrada'
            WHEN s.expires_at < NOW() THEN 'Sessão expirada'
            WHEN u.status <> 'active' THEN 'Usuário não está ativo'
            ELSE 'Válida'
        END AS reason
    FROM sessions s
    JOIN users u ON s.user_id = u.id
    JOIN organizations o ON u.organization_id = o.id
    WHERE s.token = p_token;
    
    -- Atualizar timestamp de última atividade se a sessão for válida
    UPDATE sessions
    SET last_activity = NOW()
    WHERE token = p_token
      AND is_active = TRUE
      AND expires_at > NOW();
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION verify_session IS 'Verifica se uma sessão é válida e retorna informações relacionadas';

-- Função para obter estatísticas de cumprimento regulatório por organização
CREATE OR REPLACE FUNCTION get_organization_compliance_summary(
    p_organization_id UUID
) RETURNS TABLE(
    framework_code VARCHAR,
    framework_name VARCHAR,
    sector VARCHAR,
    region VARCHAR,
    validation_count BIGINT,
    compliant_count BIGINT,
    partial_count BIGINT,
    non_compliant_count BIGINT,
    compliance_rate NUMERIC,
    last_validation TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        rf.code AS framework_code,
        rf.name AS framework_name,
        rf.sector,
        rf.region,
        COUNT(vr.id) AS validation_count,
        SUM(CASE WHEN vr.status = 'COMPLIANT' THEN 1 ELSE 0 END) AS compliant_count,
        SUM(CASE WHEN vr.status = 'PARTIAL_COMPLIANCE' THEN 1 ELSE 0 END) AS partial_count,
        SUM(CASE WHEN vr.status = 'NON_COMPLIANT' THEN 1 ELSE 0 END) AS non_compliant_count,
        CASE 
            WHEN COUNT(vr.id) > 0 THEN
                ROUND(
                    (SUM(CASE WHEN vr.status = 'COMPLIANT' THEN 1 ELSE 0 END)::NUMERIC / COUNT(vr.id)) * 100,
                    2
                )
            ELSE 0
        END AS compliance_rate,
        MAX(vr.created_at) AS last_validation
    FROM 
        regulatory_frameworks rf
    LEFT JOIN 
        compliance_validators cv ON rf.id = cv.framework_id
    LEFT JOIN 
        validation_results vr ON cv.code = vr.validator_id AND 
                                 vr.target_data->>'organization_id' = p_organization_id::TEXT
    WHERE 
        rf.is_active = TRUE
    GROUP BY 
        rf.code, rf.name, rf.sector, rf.region
    ORDER BY 
        compliance_rate DESC;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_organization_compliance_summary IS 'Retorna um resumo do status de compliance regulatório para uma organização';

-- Função para limpar sessões expiradas
CREATE OR REPLACE FUNCTION cleanup_expired_sessions() RETURNS INTEGER AS $$
DECLARE
    v_count INTEGER;
BEGIN
    WITH deleted_sessions AS (
        DELETE FROM sessions
        WHERE (expires_at < NOW() OR is_active = FALSE)
          AND last_activity < NOW() - INTERVAL '7 days'
        RETURNING id
    )
    SELECT COUNT(*) INTO v_count FROM deleted_sessions;
    
    RETURN v_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_expired_sessions IS 'Limpa sessões expiradas ou inativas com mais de 7 dias de inatividade';

-- Função para pesquisar usuários por critérios
CREATE OR REPLACE FUNCTION search_users(
    p_organization_id UUID,
    p_search_term VARCHAR DEFAULT NULL,
    p_status VARCHAR DEFAULT NULL,
    p_role_name VARCHAR DEFAULT NULL,
    p_limit INTEGER DEFAULT 100,
    p_offset INTEGER DEFAULT 0
) RETURNS TABLE(
    user_id UUID,
    username VARCHAR,
    email VARCHAR,
    full_name VARCHAR,
    status VARCHAR,
    created_at TIMESTAMP WITH TIME ZONE,
    last_login TIMESTAMP WITH TIME ZONE,
    roles TEXT[],
    total_count BIGINT
) AS $$
BEGIN
    RETURN QUERY
    WITH matched_users AS (
        SELECT 
            u.id,
            u.username,
            u.email,
            u.full_name,
            u.status,
            u.created_at,
            u.last_login,
            array_agg(DISTINCT r.name) AS roles
        FROM 
            users u
        LEFT JOIN 
            user_roles ur ON u.id = ur.user_id AND ur.is_active = TRUE
        LEFT JOIN 
            roles r ON ur.role_id = r.id
        WHERE 
            u.organization_id = p_organization_id
            AND (
                p_search_term IS NULL 
                OR u.username ILIKE '%' || p_search_term || '%'
                OR u.email ILIKE '%' || p_search_term || '%'
                OR u.full_name ILIKE '%' || p_search_term || '%'
            )
            AND (p_status IS NULL OR u.status = p_status)
            AND (p_role_name IS NULL OR r.name = p_role_name)
        GROUP BY 
            u.id, u.username, u.email, u.full_name, u.status, u.created_at, u.last_login
    )
    SELECT 
        mu.id,
        mu.username,
        mu.email,
        mu.full_name,
        mu.status,
        mu.created_at,
        mu.last_login,
        mu.roles,
        COUNT(*) OVER() AS total_count
    FROM 
        matched_users mu
    ORDER BY 
        mu.username
    LIMIT p_limit
    OFFSET p_offset;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION search_users IS 'Pesquisa usuários por vários critérios com suporte a paginação';
