-- Script para corrigir a função de auditoria
-- Autor: Eduardo Jeremias
-- Data: 19/05/2025
-- Descrição: Corrige a função fn_record_audit_log para usar a estrutura correta da tabela audit_logs

-- Configurar caminho de busca
SET search_path TO iam, public;

-- Desativar triggers temporariamente para evitar erros durante a atualização
ALTER TABLE iam.users DISABLE TRIGGER trg_users_audit;
ALTER TABLE iam.roles DISABLE TRIGGER trg_roles_audit;
ALTER TABLE iam.permissions DISABLE TRIGGER trg_permissions_audit;
ALTER TABLE iam.organizations DISABLE TRIGGER trg_organizations_audit;
ALTER TABLE iam.security_policies DISABLE TRIGGER trg_security_policies_audit;

-- Atualizar a função fn_record_audit_log
CREATE OR REPLACE FUNCTION fn_record_audit_log()
RETURNS TRIGGER AS $$
DECLARE
    v_action VARCHAR(10);
    v_old_data JSONB := NULL;
    v_new_data JSONB := NULL;
    v_changed_fields JSONB := NULL;
    v_user_id UUID;
    v_ip_address VARCHAR(50);
    v_session_id UUID;
    v_organization_id UUID;
    v_resource_id VARCHAR(255);
    v_details JSONB;
BEGIN
    -- Determinar a ação realizada
    IF TG_OP = 'INSERT' THEN
        v_action := 'CREATE';
        v_new_data := to_jsonb(NEW);
        v_resource_id := NEW.id::VARCHAR;
        
        -- Obter organization_id do registro, se disponível
        IF TG_TABLE_NAME = 'users' OR TG_TABLE_NAME = 'roles' OR TG_TABLE_NAME = 'security_policies' THEN
            v_organization_id := NEW.organization_id;
        ELSIF TG_TABLE_NAME = 'user_roles' THEN
            -- Para user_roles, obter organization_id do usuário
            SELECT organization_id INTO v_organization_id 
            FROM users 
            WHERE id = NEW.user_id;
        END IF;
        
    ELSIF TG_OP = 'UPDATE' THEN
        v_action := 'UPDATE';
        v_old_data := to_jsonb(OLD);
        v_new_data := to_jsonb(NEW);
        v_resource_id := NEW.id::VARCHAR;
        
        -- Obter organization_id do registro, se disponível
        IF TG_TABLE_NAME = 'users' OR TG_TABLE_NAME = 'roles' OR TG_TABLE_NAME = 'security_policies' THEN
            v_organization_id := NEW.organization_id;
        ELSIF TG_TABLE_NAME = 'user_roles' THEN
            -- Para user_roles, obter organization_id do usuário
            SELECT organization_id INTO v_organization_id 
            FROM users 
            WHERE id = NEW.user_id;
        END IF;
        
        -- Calcular campos alterados
        SELECT jsonb_object_agg(key, value) INTO v_changed_fields
        FROM jsonb_each(to_jsonb(NEW))
        WHERE (NOT to_jsonb(OLD) ? key) OR (to_jsonb(OLD)->>key IS DISTINCT FROM to_jsonb(NEW)->>key);
        
    ELSIF TG_OP = 'DELETE' THEN
        v_action := 'DELETE';
        v_old_data := to_jsonb(OLD);
        v_resource_id := OLD.id::VARCHAR;
        
        -- Obter organization_id do registro, se disponível
        IF TG_TABLE_NAME = 'users' OR TG_TABLE_NAME = 'roles' OR TG_TABLE_NAME = 'security_policies' THEN
            v_organization_id := OLD.organization_id;
        ELSIF TG_TABLE_NAME = 'user_roles' THEN
            -- Para user_roles, obter organization_id do usuário
            SELECT organization_id INTO v_organization_id 
            FROM users 
            WHERE id = OLD.user_id;
        END IF;
    END IF;

    -- Obter informações do usuário atual
    BEGIN
        -- Tentar obter o ID do usuário da sessão atual
        v_user_id := current_setting('app.current_user_id', TRUE)::UUID;
    EXCEPTION WHEN OTHERS THEN
        -- Se não estiver definido, usar o usuário admin
        SELECT id INTO v_user_id FROM users WHERE username = 'admin' LIMIT 1;
        
        -- Se não existir usuário admin, definir como NULL
        IF v_user_id IS NULL THEN
            v_user_id := NULL;
        END IF;
    END;
    
    -- Obter endereço IP, se disponível
    BEGIN
        v_ip_address := current_setting('app.client_ip', TRUE);
    EXCEPTION WHEN OTHERS THEN
        v_ip_address := NULL;
    END;
    
    -- Obter ID da sessão, se disponível
    BEGIN
        v_session_id := current_setting('app.session_id', TRUE)::UUID;
    EXCEPTION WHEN OTHERS THEN
        v_session_id := NULL;
    END;
    
    -- Se não conseguimos obter o organization_id, usar um valor padrão
    IF v_organization_id IS NULL THEN
        -- Tentar obter a organização padrão (InnovaBiz)
        SELECT id INTO v_organization_id FROM organizations WHERE name = 'InnovaBiz' LIMIT 1;
        
        -- Se não existir, usar a primeira organização disponível
        IF v_organization_id IS NULL THEN
            SELECT id INTO v_organization_id FROM organizations LIMIT 1;
        END IF;
    END IF;
    
    -- Preparar os detalhes do log
    v_details := jsonb_build_object(
        'action', v_action,
        'table', TG_TABLE_NAME,
        'old_data', v_old_data,
        'new_data', v_new_data,
        'changed_fields', v_changed_fields
    );
    
    -- Inserir o registro de auditoria
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
        timestamp
    ) VALUES (
        v_organization_id,
        v_user_id,
        v_action,
        TG_TABLE_NAME,
        v_resource_id,
        'SUCCESS',
        v_details,
        v_ip_address,
        v_session_id,
        NOW()
    );
    
    -- Retornar o registro apropriado
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Reativar os triggers
ALTER TABLE iam.users ENABLE TRIGGER trg_users_audit;
ALTER TABLE iam.roles ENABLE TRIGGER trg_roles_audit;
ALTER TABLE iam.permissions ENABLE TRIGGER trg_permissions_audit;
ALTER TABLE iam.organizations ENABLE TRIGGER trg_organizations_audit;
ALTER TABLE iam.security_policies ENABLE TRIGGER trg_security_policies_audit;

-- Comentário para documentação
COMMENT ON FUNCTION fn_record_audit_log IS 'Registra alterações em entidades no log de auditoria, adaptado para a estrutura da tabela audit_logs';

-- Verificar se a função foi atualizada corretamente
SELECT proname, prosrc 
FROM pg_proc 
WHERE proname = 'fn_record_audit_log' 
AND pronamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'iam');
