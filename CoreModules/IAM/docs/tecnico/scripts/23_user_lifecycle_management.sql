-- Script de Gestão do Ciclo de Vida de Usuários - IAM Open X
-- Versão: 1.0
-- Data: 15/05/2025

-- 1. Funções de Gestão de Usuários

-- Função para provisionamento de usuário com validação de políticas
CREATE OR REPLACE FUNCTION iam_access_control.provision_user(
    p_username TEXT,
    p_email TEXT,
    p_full_name TEXT,
    p_department TEXT,
    p_role TEXT[],
    p_dominio TEXT DEFAULT NULL
)
RETURNS JSON AS $$
DECLARE
    v_user_record RECORD;
    v_result JSON;
BEGIN
    -- Verificar se o usuário já existe
    SELECT * INTO v_user_record 
    FROM iam_access_control.users 
    WHERE username = p_username;
    
    IF FOUND THEN
        RAISE EXCEPTION 'Usuário já existe: %', p_username;
    END IF;
    
    -- Validar e-mail
    IF p_email !~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$' THEN
        RAISE EXCEPTION 'E-mail inválido: %', p_email;
    END IF;
    
    -- Validar permissões solicitadas
    IF NOT EXISTS (
        SELECT 1 FROM unnest(p_role) r 
        WHERE NOT EXISTS (
            SELECT 1 FROM iam_access_control.valid_roles vr 
            WHERE vr.role_name = r
        )
    ) THEN
        RAISE EXCEPTION 'Permissão inválida solicitada';
    END IF;
    
    -- Criar usuário no banco de dados
    INSERT INTO iam_access_control.users (
        username, 
        email, 
        full_name, 
        department, 
        domain, 
        created_at, 
        last_updated
    ) VALUES (
        p_username, 
        p_email, 
        p_full_name, 
        p_department, 
        p_dominio,
        current_timestamp,
        current_timestamp
    ) RETURNING * INTO v_user_record;
    
    -- Atribuir permissões
    PERFORM iam_access_control.assign_roles(
        p_username,
        p_role,
        current_user
    );
    
    -- Registrar no log de auditoria
    INSERT INTO iam_access_control.audit_log (
        action_type,
        action_by,
        action_date,
        target_user,
        details
    ) VALUES (
        'USER_PROVISIONED',
        current_user,
        current_timestamp,
        p_username,
        jsonb_build_object(
            'email', p_email,
            'roles', p_role,
            'domain', p_dominio
        )
    );
    
    -- Retornar resultado
    v_result := json_build_object(
        'status', 'success',
        'message', 'Usuário provisionado com sucesso',
        'user', json_build_object(
            'username', p_username,
            'email', p_email,
            'full_name', p_full_name,
            'department', p_department,
            'domain', p_dominio,
            'created_at', v_user_record.created_at
        )
    );
    
    RETURN v_result;
EXCEPTION
    WHEN OTHERS THEN
        INSERT INTO iam_access_control.error_log (
            error_type,
            error_message,
            error_date,
            operation,
            details
        ) VALUES (
            SQLSTATE,
            SQLERRM,
            current_timestamp,
            'USER_PROVISIONING',
            jsonb_build_object(
                'username', p_username,
                'email', p_email,
                'roles', p_role
            )
        );
        
        RETURN json_build_object(
            'status', 'error',
            'message', SQLERRM,
            'error_code', SQLSTATE
        );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
