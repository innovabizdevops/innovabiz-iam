-- INNOVABIZ - IAM Core Triggers
-- Author: Eduardo Jeremias
-- Date: 08/05/2025
-- Version: 1.0
-- Description: Triggers para o módulo IAM core, garantindo integridade, auditoria e automações.

-- Configurar caminho de busca
SET search_path TO iam, public;

-- ===============================================================================
-- TRIGGERS PARA AUDITORIA
-- ===============================================================================

-- Função para registrar alterações em auditoria
CREATE OR REPLACE FUNCTION iam.fn_record_audit_log()
RETURNS TRIGGER AS $$
DECLARE
    v_action VARCHAR(20);
    v_old_data JSONB := NULL;
    v_new_data JSONB := NULL;
    v_changed_fields JSONB := '{}'::JSONB;
    
    v_user_id UUID;
    v_ip_address TEXT;
    v_session_id UUID;
    v_application_name TEXT;
    v_change_reason TEXT;

    v_entity_id_text TEXT;
    v_organization_id_for_audit UUID; 

    v_details JSONB;
BEGIN
    -- Determinar a ação realizada e preparar dados JSON
    IF TG_OP = 'INSERT' THEN
        v_action := 'INSERT';
        v_new_data := to_jsonb(NEW);
        v_entity_id_text := (v_new_data->>'id')::TEXT;
    ELSIF TG_OP = 'UPDATE' THEN
        v_action := 'UPDATE';
        v_old_data := to_jsonb(OLD);
        v_new_data := to_jsonb(NEW);
        v_entity_id_text := (v_new_data->>'id')::TEXT;

        IF v_new_data IS DISTINCT FROM v_old_data THEN
            SELECT jsonb_object_agg(key, value)
            INTO v_changed_fields
            FROM jsonb_each(v_new_data) new_data_kv
            LEFT JOIN jsonb_each(v_old_data) old_data_kv ON new_data_kv.key = old_data_kv.key
            WHERE old_data_kv.key IS NULL OR new_data_kv.value IS DISTINCT FROM old_data_kv.value;
        END IF;
    ELSIF TG_OP = 'DELETE' THEN
        v_action := 'DELETE';
        v_old_data := to_jsonb(OLD);
        v_entity_id_text := (v_old_data->>'id')::TEXT;
    ELSE
        RAISE WARNING '[IAM AUDIT] Unhandled TG_OP: % in fn_record_audit_log for table %', TG_OP, TG_TABLE_NAME;
        IF TG_OP = 'DELETE' THEN RETURN OLD; ELSE RETURN NEW; END IF; -- Allow operation
    END IF;

    -- Determinar organization_id para auditoria
    IF TG_TABLE_NAME = 'organizations' THEN
        v_organization_id_for_audit := (CASE WHEN TG_OP = 'DELETE' THEN v_old_data->>'id' ELSE v_new_data->>'id' END)::UUID;
    ELSE
        v_organization_id_for_audit := (CASE WHEN TG_OP = 'DELETE' THEN v_old_data->>'organization_id' ELSE v_new_data->>'organization_id' END)::UUID;
    END IF;

    -- Obter informações contextuais da sessão (aplicação DEVE configurá-las)
    BEGIN
        v_user_id := current_setting('app.current_user_id', true)::UUID;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE '[IAM AUDIT] app.current_user_id not set/invalid for table %. Using NULL user_id. SQLSTATE: %, SQLERRM: %', TG_TABLE_NAME, SQLSTATE, SQLERRM;
        v_user_id := NULL;
    END;
    BEGIN
        v_ip_address := current_setting('app.client_ip', true);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE '[IAM AUDIT] app.client_ip not set/invalid for table %. Using NULL ip_address. SQLSTATE: %, SQLERRM: %', TG_TABLE_NAME, SQLSTATE, SQLERRM;
        v_ip_address := NULL;
    END;
    BEGIN
        v_session_id := current_setting('app.session_id', true)::UUID;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE '[IAM AUDIT] app.session_id not set/invalid for table %. Using NULL session_id. SQLSTATE: %, SQLERRM: %', TG_TABLE_NAME, SQLSTATE, SQLERRM;
        v_session_id := NULL;
    END;
    BEGIN
        v_application_name := current_setting('application_name', true);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE '[IAM AUDIT] application_name not set/invalid for table %. Using NULL application_name. SQLSTATE: %, SQLERRM: %', TG_TABLE_NAME, SQLSTATE, SQLERRM;
        v_application_name := NULL;
    END;
    v_change_reason := current_setting('app.change_reason', false); -- Opcional

    -- Preparar detalhes para o log de auditoria
    v_details := jsonb_build_object(
        'old_data', v_old_data,
        'new_data', v_new_data,
        'changed_fields', v_changed_fields,
        'trigger_table', TG_TABLE_NAME,
        'trigger_operation', TG_OP,
        'change_reason', v_change_reason,
        'application_name_from_setting', v_application_name
    );
    
    -- Chamar a função central de registo de auditoria
    PERFORM iam.register_audit_log(
        p_organization_id => v_organization_id_for_audit, 
        p_user_id => v_user_id,
        p_action => v_action || '_' || TG_TABLE_NAME, 
        p_resource_type => TG_TABLE_NAME,
        p_resource_id => v_entity_id_text, 
        p_status => 'SUCCESS',
        p_details => v_details,
        p_ip_address => v_ip_address,
        p_session_id => v_session_id,
        p_request_id => NULL 
    );

    /* -- Lógica para user_status_changes comentada: Tabela não definida e v_audit_id não é retornado por PERFORM.
    IF TG_TABLE_NAME = 'users' AND TG_OP = 'UPDATE' THEN
        IF OLD.status IS DISTINCT FROM NEW.status THEN
            -- INSERT INTO iam.user_status_changes (user_id, audit_log_id, old_status, new_status, changed_by, change_reason) ...
        END IF;
    END IF;
    */

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;

EXCEPTION
    WHEN OTHERS THEN
        RAISE WARNING '[IAM AUDIT] CRITICAL ERROR in fn_record_audit_log for table % (TG_OP: %): % (%) - Operation proceeded without audit.', TG_TABLE_NAME, TG_OP, SQLERRM, SQLSTATE;
        IF TG_OP = 'DELETE' THEN RETURN OLD; ELSE RETURN NEW; END IF; -- Permite que a operação original continue apesar da falha na auditoria
END;
$$ LANGUAGE plpgsql SECURITY INVOKER; -- ALTERADO PARA SECURITY INVOKER

-- Triggers de auditoria para tabelas principais

-- Organizações
DROP TRIGGER IF EXISTS trg_organizations_audit ON iam.organizations;
CREATE TRIGGER trg_organizations_audit
AFTER INSERT OR UPDATE OR DELETE ON iam.organizations
FOR EACH ROW EXECUTE FUNCTION iam.fn_record_audit_log();

-- Usuários
DROP TRIGGER IF EXISTS trg_users_audit ON iam.users;
CREATE TRIGGER trg_users_audit
AFTER INSERT OR UPDATE OR DELETE ON iam.users
FOR EACH ROW EXECUTE FUNCTION iam.fn_record_audit_log();

-- Roles
DROP TRIGGER IF EXISTS trg_roles_audit ON iam.roles;
CREATE TRIGGER trg_roles_audit
AFTER INSERT OR UPDATE OR DELETE ON iam.roles
FOR EACH ROW EXECUTE FUNCTION iam.fn_record_audit_log();

-- Permissões
DROP TRIGGER IF EXISTS trg_permissions_audit ON iam.permissions;
CREATE TRIGGER trg_permissions_audit
AFTER INSERT OR UPDATE OR DELETE ON iam.permissions
FOR EACH ROW EXECUTE FUNCTION iam.fn_record_audit_log();

-- User Roles (Associação Usuário-Role)
DROP TRIGGER IF EXISTS trg_user_roles_audit ON iam.user_roles;
CREATE TRIGGER trg_user_roles_audit
AFTER INSERT OR UPDATE OR DELETE ON iam.user_roles
FOR EACH ROW EXECUTE FUNCTION iam.fn_record_audit_log();

-- Sessões
DROP TRIGGER IF EXISTS trg_sessions_audit ON iam.sessions;
CREATE TRIGGER trg_sessions_audit
AFTER INSERT OR UPDATE OR DELETE ON iam.sessions
FOR EACH ROW EXECUTE FUNCTION iam.fn_record_audit_log();

-- Políticas de segurança
DROP TRIGGER IF EXISTS trg_security_policies_audit ON iam.security_policies;
CREATE TRIGGER trg_security_policies_audit
AFTER INSERT OR UPDATE OR DELETE ON iam.security_policies
FOR EACH ROW EXECUTE FUNCTION iam.fn_record_audit_log();

-- Regulatory Frameworks
DROP TRIGGER IF EXISTS trg_regulatory_frameworks_audit ON iam.regulatory_frameworks;
CREATE TRIGGER trg_regulatory_frameworks_audit
AFTER INSERT OR UPDATE OR DELETE ON iam.regulatory_frameworks
FOR EACH ROW EXECUTE FUNCTION iam.fn_record_audit_log();

-- Validadores de compliance
DROP TRIGGER IF EXISTS trg_compliance_validators_audit ON iam.compliance_validators;
CREATE TRIGGER trg_compliance_validators_audit
AFTER INSERT OR UPDATE OR DELETE ON iam.compliance_validators
FOR EACH ROW EXECUTE FUNCTION iam.fn_record_audit_log();

-- Auditoria para iam.role_permissions (anteriormente comentada, agora ativa)
DROP TRIGGER IF EXISTS trg_role_permissions_audit ON iam.role_permissions;
CREATE TRIGGER trg_role_permissions_audit
AFTER INSERT OR UPDATE OR DELETE ON iam.role_permissions
FOR EACH ROW EXECUTE FUNCTION iam.fn_record_audit_log();

/* -- Triggers para outras tabelas não definidas ou cuja auditoria será revista:
DROP TRIGGER IF EXISTS trg_password_policies_audit ON iam.password_policies; -- Tabela não definida
CREATE TRIGGER trg_password_policies_audit
AFTER INSERT OR UPDATE OR DELETE ON iam.password_policies
FOR EACH ROW EXECUTE FUNCTION iam.fn_record_audit_log();

DROP TRIGGER IF EXISTS trg_compliance_validation_results_audit ON iam.compliance_validation_results; -- Tabela não definida
CREATE TRIGGER trg_compliance_validation_results_audit
AFTER INSERT OR UPDATE OR DELETE ON iam.compliance_validation_results
FOR EACH ROW EXECUTE FUNCTION iam.fn_record_audit_log();
*/

-- ===============================================================================
-- TRIGGERS PARA MANUTENÇÃO DE UPDATED_AT
-- ===============================================================================

-- Função para atualizar o timestamp updated_at
CREATE OR REPLACE FUNCTION iam.fn_update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Aplicar a função para todas as tabelas que possuem o campo updated_at

DROP TRIGGER IF EXISTS trg_organizations_timestamp ON iam.organizations;
CREATE TRIGGER trg_organizations_timestamp
BEFORE UPDATE ON iam.organizations
FOR EACH ROW EXECUTE FUNCTION iam.fn_update_timestamp();

DROP TRIGGER IF EXISTS trg_users_timestamp ON iam.users;
CREATE TRIGGER trg_users_timestamp
BEFORE UPDATE ON iam.users
FOR EACH ROW EXECUTE FUNCTION iam.fn_update_timestamp();

DROP TRIGGER IF EXISTS trg_roles_timestamp ON iam.roles;
CREATE TRIGGER trg_roles_timestamp
BEFORE UPDATE ON iam.roles
FOR EACH ROW EXECUTE FUNCTION iam.fn_update_timestamp();

DROP TRIGGER IF EXISTS trg_permissions_timestamp ON iam.permissions;
CREATE TRIGGER trg_permissions_timestamp
BEFORE UPDATE ON iam.permissions
FOR EACH ROW EXECUTE FUNCTION iam.fn_update_timestamp();

DROP TRIGGER IF EXISTS trg_user_roles_timestamp ON iam.user_roles;
CREATE TRIGGER trg_user_roles_timestamp
BEFORE UPDATE ON iam.user_roles
FOR EACH ROW EXECUTE FUNCTION iam.fn_update_timestamp();

DROP TRIGGER IF EXISTS trg_sessions_timestamp ON iam.sessions;
CREATE TRIGGER trg_sessions_timestamp
BEFORE UPDATE ON iam.sessions
FOR EACH ROW EXECUTE FUNCTION iam.fn_update_timestamp();

DROP TRIGGER IF EXISTS trg_security_policies_timestamp ON iam.security_policies;
CREATE TRIGGER trg_security_policies_timestamp
BEFORE UPDATE ON iam.security_policies
FOR EACH ROW EXECUTE FUNCTION iam.fn_update_timestamp();

DROP TRIGGER IF EXISTS trg_regulatory_frameworks_timestamp ON iam.regulatory_frameworks;
CREATE TRIGGER trg_regulatory_frameworks_timestamp
BEFORE UPDATE ON iam.regulatory_frameworks
FOR EACH ROW EXECUTE FUNCTION iam.fn_update_timestamp();

DROP TRIGGER IF EXISTS trg_compliance_validators_timestamp ON iam.compliance_validators;
CREATE TRIGGER trg_compliance_validators_timestamp
BEFORE UPDATE ON iam.compliance_validators
FOR EACH ROW EXECUTE FUNCTION iam.fn_update_timestamp();

-- Trigger de timestamp para iam.role_permissions
DROP TRIGGER IF EXISTS trg_role_permissions_timestamp ON iam.role_permissions;
CREATE TRIGGER trg_role_permissions_timestamp
BEFORE UPDATE ON iam.role_permissions
FOR EACH ROW EXECUTE FUNCTION iam.fn_update_timestamp();

-- Adicionar aqui outros triggers de timestamp para tabelas que possam ser criadas posteriormente e que tenham 'updated_at'
-- Exemplo: 
-- DROP TRIGGER IF EXISTS trg_nova_tabela_timestamp ON iam.nova_tabela;
-- CREATE TRIGGER trg_nova_tabela_timestamp
-- BEFORE UPDATE ON iam.nova_tabela
-- FOR EACH ROW EXECUTE FUNCTION iam.fn_update_timestamp();
-- Fim dos triggers de timestamp. As definições duplicadas ou incorretas que estavam aqui foram removidas.

DROP TRIGGER IF EXISTS trg_regulatory_frameworks_timestamp ON regulatory_frameworks;
CREATE TRIGGER trg_regulatory_frameworks_timestamp
BEFORE UPDATE ON regulatory_frameworks
FOR EACH ROW EXECUTE FUNCTION iam.fn_update_timestamp();

-- ===============================================================================
-- TRIGGERS PARA VALIDAÇÃO DE DADOS
-- ===============================================================================

-- Função para validar emails
CREATE OR REPLACE FUNCTION fn_validate_email()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.email IS NOT NULL AND NEW.email !~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$' THEN
        RAISE EXCEPTION 'Email inválido: %', NEW.email;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger para validar emails de usuários
DROP TRIGGER IF EXISTS trg_validate_user_email ON iam.users;
CREATE TRIGGER trg_validate_user_email
BEFORE INSERT OR UPDATE OF email ON iam.users
FOR EACH ROW EXECUTE FUNCTION iam.fn_validate_email();

-- Função para validar números de telefone
CREATE OR REPLACE FUNCTION fn_validate_phone_number()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.phone_number IS NOT NULL AND NEW.phone_number !~ '^\+?[0-9\s\-\(\)\.]{8,20}$' THEN
        RAISE EXCEPTION 'Número de telefone inválido: %', NEW.phone_number;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger para validar números de telefone de usuários
DROP TRIGGER IF EXISTS trg_validate_user_phone ON iam.users;
CREATE TRIGGER trg_validate_user_phone
BEFORE INSERT OR UPDATE OF phone_number ON iam.users
FOR EACH ROW EXECUTE FUNCTION iam.fn_validate_phone_number();

-- Função para validar CPF/CNPJ brasileiros
CREATE OR REPLACE FUNCTION fn_validate_brazilian_document()
RETURNS TRIGGER AS $$
DECLARE
    v_document VARCHAR;
    v_sum INTEGER := 0;
    v_digit INTEGER;
    v_rest INTEGER;
    v_cpf_valid BOOLEAN;
    v_cnpj_valid BOOLEAN;
BEGIN
    IF NEW.country_code = 'BR' AND NEW.document_number IS NOT NULL THEN
        v_document := REGEXP_REPLACE(NEW.document_number, '[^0-9]', '', 'g');
        
        -- Verificar se é um CPF (11 dígitos) ou CNPJ (14 dígitos)
        IF LENGTH(v_document) = 11 THEN
            -- Validação de CPF
            -- Verificar se todos os dígitos são iguais
            IF v_document ~ '^(\d)\1{10}$' THEN
                v_cpf_valid := FALSE;
            ELSE
                -- Cálculo do primeiro dígito verificador
                v_sum := 0;
                FOR i IN 1..9 LOOP
                    v_sum := v_sum + (SUBSTRING(v_document FROM i FOR 1)::INTEGER * (11 - i));
                END LOOP;
                v_rest := v_sum % 11;
                
                IF v_rest < 2 THEN
                    v_digit := 0;
                ELSE
                    v_digit := 11 - v_rest;
                END IF;
                
                IF v_digit != SUBSTRING(v_document FROM 10 FOR 1)::INTEGER THEN
                    v_cpf_valid := FALSE;
                ELSE
                    -- Cálculo do segundo dígito verificador
                    v_sum := 0;
                    FOR i IN 1..10 LOOP
                        v_sum := v_sum + (SUBSTRING(v_document FROM i FOR 1)::INTEGER * (12 - i));
                    END LOOP;
                    v_rest := v_sum % 11;
                    
                    IF v_rest < 2 THEN
                        v_digit := 0;
                    ELSE
                        v_digit := 11 - v_rest;
                    END IF;
                    
                    v_cpf_valid := (v_digit = SUBSTRING(v_document FROM 11 FOR 1)::INTEGER);
                END IF;
            END IF;
            
            IF NOT v_cpf_valid THEN
                RAISE EXCEPTION 'CPF inválido: %', NEW.document_number;
            END IF;
        ELSIF LENGTH(v_document) = 14 THEN
            -- Validação de CNPJ
            -- Verificar se todos os dígitos são iguais
            IF v_document ~ '^(\d)\1{13}$' THEN
                v_cnpj_valid := FALSE;
            ELSE
                -- Cálculo do primeiro dígito verificador
                v_sum := 0;
                v_sum := v_sum + (SUBSTRING(v_document FROM 1 FOR 1)::INTEGER * 5);
                v_sum := v_sum + (SUBSTRING(v_document FROM 2 FOR 1)::INTEGER * 4);
                v_sum := v_sum + (SUBSTRING(v_document FROM 3 FOR 1)::INTEGER * 3);
                v_sum := v_sum + (SUBSTRING(v_document FROM 4 FOR 1)::INTEGER * 2);
                v_sum := v_sum + (SUBSTRING(v_document FROM 5 FOR 1)::INTEGER * 9);
                v_sum := v_sum + (SUBSTRING(v_document FROM 6 FOR 1)::INTEGER * 8);
                v_sum := v_sum + (SUBSTRING(v_document FROM 7 FOR 1)::INTEGER * 7);
                v_sum := v_sum + (SUBSTRING(v_document FROM 8 FOR 1)::INTEGER * 6);
                v_sum := v_sum + (SUBSTRING(v_document FROM 9 FOR 1)::INTEGER * 5);
                v_sum := v_sum + (SUBSTRING(v_document FROM 10 FOR 1)::INTEGER * 4);
                v_sum := v_sum + (SUBSTRING(v_document FROM 11 FOR 1)::INTEGER * 3);
                v_sum := v_sum + (SUBSTRING(v_document FROM 12 FOR 1)::INTEGER * 2);
                
                v_rest := v_sum % 11;
                IF v_rest < 2 THEN
                    v_digit := 0;
                ELSE
                    v_digit := 11 - v_rest;
                END IF;
                
                IF v_digit != SUBSTRING(v_document FROM 13 FOR 1)::INTEGER THEN
                    v_cnpj_valid := FALSE;
                ELSE
                    -- Cálculo do segundo dígito verificador
                    v_sum := 0;
                    v_sum := v_sum + (SUBSTRING(v_document FROM 1 FOR 1)::INTEGER * 6);
                    v_sum := v_sum + (SUBSTRING(v_document FROM 2 FOR 1)::INTEGER * 5);
                    v_sum := v_sum + (SUBSTRING(v_document FROM 3 FOR 1)::INTEGER * 4);
                    v_sum := v_sum + (SUBSTRING(v_document FROM 4 FOR 1)::INTEGER * 3);
                    v_sum := v_sum + (SUBSTRING(v_document FROM 5 FOR 1)::INTEGER * 2);
                    v_sum := v_sum + (SUBSTRING(v_document FROM 6 FOR 1)::INTEGER * 9);
                    v_sum := v_sum + (SUBSTRING(v_document FROM 7 FOR 1)::INTEGER * 8);
                    v_sum := v_sum + (SUBSTRING(v_document FROM 8 FOR 1)::INTEGER * 7);
                    v_sum := v_sum + (SUBSTRING(v_document FROM 9 FOR 1)::INTEGER * 6);
                    v_sum := v_sum + (SUBSTRING(v_document FROM 10 FOR 1)::INTEGER * 5);
                    v_sum := v_sum + (SUBSTRING(v_document FROM 11 FOR 1)::INTEGER * 4);
                    v_sum := v_sum + (SUBSTRING(v_document FROM 12 FOR 1)::INTEGER * 3);
                    v_sum := v_sum + (SUBSTRING(v_document FROM 13 FOR 1)::INTEGER * 2);
                    
                    v_rest := v_sum % 11;
                    IF v_rest < 2 THEN
                        v_digit := 0;
                    ELSE
                        v_digit := 11 - v_rest;
                    END IF;
                    
                    v_cnpj_valid := (v_digit = SUBSTRING(v_document FROM 14 FOR 1)::INTEGER);
                END IF;
            END IF;
            
            IF NOT v_cnpj_valid THEN
                RAISE EXCEPTION 'CNPJ inválido: %', NEW.document_number;
            END IF;
        ELSE
            RAISE EXCEPTION 'Documento brasileiro deve ser CPF (11 dígitos) ou CNPJ (14 dígitos)';
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger para validar documentos brasileiros
DROP TRIGGER IF EXISTS trg_validate_user_document ON iam.users;
CREATE TRIGGER trg_validate_user_document
BEFORE INSERT OR UPDATE OF document_number, country_code ON iam.users
FOR EACH ROW EXECUTE FUNCTION iam.fn_validate_brazilian_document();

-- ===============================================================================
-- TRIGGERS PARA ATUALIZAÇÕES AUTOMÁTICAS
-- ===============================================================================

-- Função para sincronizar alterações de permissões de roles
CREATE OR REPLACE FUNCTION fn_sync_user_permissions()
RETURNS TRIGGER AS $$
BEGIN
    -- Atualizar o campo last_permissions_update dos usuários afetados
    -- quando houver alterações nas permissões dos roles
    IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' OR TG_OP = 'DELETE' THEN
        UPDATE users
        SET last_permissions_update = NOW()
        WHERE role_id = CASE 
            WHEN TG_OP = 'DELETE' THEN OLD.role_id
            ELSE NEW.role_id
        END;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger para sincronizar usuários quando as permissões de um role são alteradas
DROP TRIGGER IF EXISTS trg_sync_user_permissions ON role_permissions;
CREATE TRIGGER trg_sync_user_permissions
AFTER INSERT OR UPDATE OR DELETE ON role_permissions
FOR EACH ROW EXECUTE FUNCTION iam.fn_sync_user_permissions();

-- Função para gerenciar tentativas falhas de login
CREATE OR REPLACE FUNCTION fn_manage_failed_login_attempts()
RETURNS TRIGGER AS $$
DECLARE
    v_max_attempts INTEGER;
    v_lockout_duration INTERVAL;
    v_current_attempts INTEGER;
BEGIN
    -- Obter configurações de política de segurança
    SELECT 
        COALESCE(value->>'max_login_attempts', '5')::INTEGER,
        COALESCE(value->>'lockout_duration_minutes', '30')::INTEGER || ' minutes'::INTERVAL
    INTO 
        v_max_attempts,
        v_lockout_duration
    FROM security_policies
    WHERE 
        organization_id = NEW.organization_id
        AND name = 'account_lockout_policy'
        AND is_active = TRUE
    ORDER BY 
        updated_at DESC
    LIMIT 1;
    
    -- Se a política não existir, usar valores padrão
    IF v_max_attempts IS NULL THEN
        v_max_attempts := 5;
        v_lockout_duration := '30 minutes'::INTERVAL;
    END IF;
    
    -- Se o login falhou, incrementar tentativas e verificar se deve bloquear
    IF NEW.status = 'failed' THEN
        -- Contar tentativas recentes
        SELECT COUNT(*)
        INTO v_current_attempts
        FROM login_attempts
        WHERE 
            user_id = NEW.user_id
            AND timestamp > NOW() - v_lockout_duration
            AND status = 'failed';
        
        IF v_current_attempts >= v_max_attempts THEN
            -- Bloquear a conta do usuário
            UPDATE users
            SET 
                status = 'locked',
                locked_at = NOW(),
                locked_reason = 'Múltiplas tentativas de login malsucedidas',
                locked_until = NOW() + v_lockout_duration
            WHERE id = NEW.user_id;
            
            -- Registrar o evento no log de auditoria
            INSERT INTO audit_logs (
                entity_type,
                entity_id,
                action,
                old_data,
                new_data,
                user_id,
                application_name,
                timestamp,
                details
            ) VALUES (
                'users',
                NEW.user_id,
                'LOCK',
                NULL,
                NULL,
                NEW.user_id,
                'account_security',
                NOW(),
                jsonb_build_object(
                    'reason', 'Múltiplas tentativas de login malsucedidas',
                    'attempts', v_current_attempts,
                    'lockout_duration', v_lockout_duration
                )
            );
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger para gerenciar tentativas falhas de login
DROP TRIGGER IF EXISTS trg_manage_failed_login_attempts ON login_attempts;
CREATE TRIGGER trg_manage_failed_login_attempts
AFTER INSERT ON login_attempts
FOR EACH ROW EXECUTE FUNCTION iam.fn_manage_failed_login_attempts();

-- Função para finalizar todas as sessões ativas quando o status do usuário muda
CREATE OR REPLACE FUNCTION fn_terminate_sessions_on_user_status_change()
RETURNS TRIGGER AS $$
BEGIN
    -- Se o status do usuário mudar para inativo, bloqueado ou suspenso
    IF OLD.status <> NEW.status AND NEW.status IN ('inactive', 'locked', 'suspended') THEN
        -- Encerrar todas as sessões ativas
        UPDATE iam.user_sessions
        SET 
            status = 'terminated',
            end_time = NOW(),
            termination_reason = 'User status changed to ' || NEW.status
        WHERE 
            user_id = NEW.id
            AND status = 'active';
        
        -- Registrar o evento no log de auditoria
        INSERT INTO audit_logs (
            entity_type,
            entity_id,
            action,
            old_data,
            new_data,
            user_id,
            application_name,
            timestamp,
            details
        ) VALUES (
            'iam.user_sessions',
            NEW.id,
            'TERMINATE',
            NULL,
            NULL,
            NEW.id,
            'session_management',
            NOW(),
            jsonb_build_object(
                'reason', 'User status changed to ' || NEW.status,
                'user_id', NEW.id
            )
        );
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger para encerrar sessões quando o status do usuário muda
DROP TRIGGER IF EXISTS trg_terminate_sessions_on_user_status_change ON iam.users;
CREATE TRIGGER trg_terminate_sessions_on_user_status_change
AFTER UPDATE OF status ON iam.users
FOR EACH ROW EXECUTE FUNCTION iam.fn_terminate_sessions_on_user_status_change();

-- Função para aplicar política de senhas
CREATE OR REPLACE FUNCTION fn_enforce_password_policy()
RETURNS TRIGGER AS $$
DECLARE
    v_policy_id UUID;
    v_organization_id UUID;
    v_policy record;
    v_password_text TEXT;
    v_complexity_valid BOOLEAN := TRUE;
    v_message TEXT;
BEGIN
    -- Pular validação se a senha não foi alterada
    IF TG_OP = 'UPDATE' AND NEW.password_hash = OLD.password_hash THEN
        RETURN NEW;
    END IF;
    
    -- Obter o ID da organização
    v_organization_id := NEW.organization_id;
    
    -- Obter a política de senha aplicável
    SELECT id INTO v_policy_id
    FROM iam.password_policies
    WHERE organization_id = v_organization_id AND is_active = TRUE
    ORDER BY updated_at DESC
    LIMIT 1;
    
    -- Se não houver política específica da organização, usar política padrão
    IF v_policy_id IS NULL THEN
        SELECT id INTO v_policy_id
        FROM iam.password_policies
        WHERE is_default = TRUE AND is_active = TRUE
        ORDER BY updated_at DESC
        LIMIT 1;
    END IF;
    
    -- Se encontrou uma política, aplicar validações
    IF v_policy_id IS NOT NULL THEN
        SELECT * INTO v_policy
        FROM iam.password_policies
        WHERE id = v_policy_id;
        
        -- Obter a senha em texto plano (em um ambiente real, deve-se cuidar dessa parte)
        v_password_text := current_setting('app.temp_password', TRUE);
        
        -- Se a senha estiver disponível para validação
        IF v_password_text IS NOT NULL THEN
            -- Verificar comprimento mínimo
            IF v_policy.min_length IS NOT NULL AND LENGTH(v_password_text) < v_policy.min_length THEN
                v_complexity_valid := FALSE;
                v_message := 'A senha deve ter pelo menos ' || v_policy.min_length || ' caracteres';
            END IF;
            
            -- Verificar caracteres especiais
            IF v_policy.require_special_char AND v_password_text !~ '[^a-zA-Z0-9]' THEN
                v_complexity_valid := FALSE;
                v_message := 'A senha deve conter pelo menos um caractere especial';
            END IF;
            
            -- Verificar números
            IF v_policy.require_number AND v_password_text !~ '[0-9]' THEN
                v_complexity_valid := FALSE;
                v_message := 'A senha deve conter pelo menos um número';
            END IF;
            
            -- Verificar letras maiúsculas
            IF v_policy.require_uppercase AND v_password_text !~ '[A-Z]' THEN
                v_complexity_valid := FALSE;
                v_message := 'A senha deve conter pelo menos uma letra maiúscula';
            END IF;
            
            -- Verificar letras minúsculas
            IF v_policy.require_lowercase AND v_password_text !~ '[a-z]' THEN
                v_complexity_valid := FALSE;
                v_message := 'A senha deve conter pelo menos uma letra minúscula';
            END IF;
            
            -- Se a senha não atender aos requisitos, gerar erro
            IF NOT v_complexity_valid THEN
                RAISE EXCEPTION 'Política de senha não atendida: %', v_message;
            END IF;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger para aplicar política de senhas
DROP TRIGGER IF EXISTS trg_enforce_password_policy ON iam.users;
CREATE TRIGGER trg_enforce_password_policy
BEFORE INSERT OR UPDATE OF password_hash ON iam.users
FOR EACH ROW EXECUTE FUNCTION iam.fn_enforce_password_policy();

COMMENT ON FUNCTION fn_record_audit_log IS 'Registra alterações em entidades no log de auditoria';
COMMENT ON FUNCTION fn_update_timestamp IS 'Atualiza o campo updated_at com o timestamp atual';
COMMENT ON FUNCTION fn_validate_email IS 'Valida o formato de endereços de email';
COMMENT ON FUNCTION fn_validate_phone_number IS 'Valida o formato de números de telefone';
COMMENT ON FUNCTION fn_validate_brazilian_document IS 'Valida documentos brasileiros (CPF/CNPJ)';
COMMENT ON FUNCTION fn_sync_user_permissions IS 'Sincroniza usuários quando permissões de roles são alteradas';
COMMENT ON FUNCTION fn_manage_failed_login_attempts IS 'Gerencia tentativas falhas de login e aplica bloqueio se necessário';
COMMENT ON FUNCTION fn_terminate_sessions_on_user_status_change IS 'Encerra sessões ativas quando o status do usuário muda';
COMMENT ON FUNCTION fn_enforce_password_policy IS 'Aplica políticas de senha durante criação/alteração de usuários';
