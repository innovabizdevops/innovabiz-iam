-- INNOVABIZ - IAM MFA Authentication Framework (Parte 2)
-- Author: Eduardo Jeremias
-- Date: 09/05/2025
-- Version: 1.0
-- Description: Funções e procedimentos para autenticação multi-fator

-- Configuração do esquema
SET search_path TO iam, public;

-- Função para gerar segredo TOTP
CREATE OR REPLACE FUNCTION iam.generate_totp_secret(
    p_user_id UUID,
    p_organization_id UUID
) RETURNS TEXT AS $$
DECLARE
    secret TEXT;
BEGIN
    -- Gerar um segredo base 32 para TOTP
    secret := encode(gen_random_bytes(20), 'base64');
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        p_user_id,
        'authentication'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        'GENERATE_TOTP_SECRET',
        'user_mfa_methods',
        p_user_id::TEXT,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        'success',
        NULL, -- response_time
        jsonb_build_object('user_id', p_user_id),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['authentication', 'mfa'], -- compliance_tags
        NULL, -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN secret;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para gerar códigos de backup
CREATE OR REPLACE FUNCTION iam.generate_backup_codes(
    p_user_id UUID,
    p_organization_id UUID,
    p_number_of_codes INTEGER DEFAULT 10,
    p_code_length INTEGER DEFAULT 8,
    p_expires_days INTEGER DEFAULT 365
) RETURNS TABLE (
    code TEXT
) AS $$
DECLARE
    new_code TEXT;
    code_hash TEXT;
    expires TIMESTAMP WITH TIME ZONE;
BEGIN
    -- Calcular data de expiração
    expires := NOW() + (p_expires_days || ' days')::INTERVAL;
    
    -- Remover códigos antigos não utilizados
    DELETE FROM iam.user_mfa_backup_codes 
    WHERE user_id = p_user_id 
      AND used = FALSE;
    
    -- Gerar novos códigos
    FOR i IN 1..p_number_of_codes LOOP
        -- Gerar código aleatório
        new_code := '';
        FOR j IN 1..p_code_length LOOP
            new_code := new_code || substring('0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ' FROM floor(random()*36+1)::integer FOR 1);
        END LOOP;
        
        -- Armazenar hash do código
        code_hash := crypt(new_code, gen_salt('bf'));
        
        -- Inserir na tabela
        INSERT INTO iam.user_mfa_backup_codes (
            user_id,
            organization_id,
            code_hash,
            expires_at
        ) VALUES (
            p_user_id,
            p_organization_id,
            code_hash,
            expires
        );
        
        -- Retornar código para o usuário
        code := new_code;
        RETURN NEXT;
    END LOOP;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        p_user_id,
        'authentication'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        'GENERATE_BACKUP_CODES',
        'user_mfa_backup_codes',
        p_user_id::TEXT,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        'success',
        NULL, -- response_time
        jsonb_build_object(
            'user_id', p_user_id,
            'codes_generated', p_number_of_codes,
            'expires_days', p_expires_days
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['authentication', 'mfa'], -- compliance_tags
        NULL, -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para registrar um método MFA para um usuário
CREATE OR REPLACE FUNCTION iam.register_mfa_method(
    p_user_id UUID,
    p_organization_id UUID,
    p_method_type iam.mfa_method_type,
    p_method_name VARCHAR,
    p_secret TEXT DEFAULT NULL,
    p_phone_number VARCHAR DEFAULT NULL,
    p_email VARCHAR DEFAULT NULL,
    p_metadata JSONB DEFAULT '{}'::JSONB
) RETURNS UUID AS $$
DECLARE
    method_id UUID;
    allowed_methods iam.mfa_method_type[];
    is_allowed BOOLEAN := FALSE;
BEGIN
    -- Verificar se o método é permitido para a organização
    SELECT allowed_methods INTO allowed_methods
    FROM iam.mfa_organization_settings
    WHERE organization_id = p_organization_id;
    
    IF allowed_methods IS NULL THEN
        -- Se não há configuração, usar padrão
        allowed_methods := ARRAY['totp', 'email']::iam.mfa_method_type[];
    END IF;
    
    -- Verificar se o método solicitado está entre os permitidos
    FOR i IN 1..array_length(allowed_methods, 1) LOOP
        IF allowed_methods[i] = p_method_type THEN
            is_allowed := TRUE;
            EXIT;
        END IF;
    END LOOP;
    
    IF NOT is_allowed THEN
        RAISE EXCEPTION 'Método MFA não permitido para esta organização: %', p_method_type;
    END IF;
    
    -- Inserir o novo método
    INSERT INTO iam.user_mfa_methods (
        user_id,
        organization_id,
        method_type,
        name,
        secret,
        phone_number,
        email,
        metadata
    ) VALUES (
        p_user_id,
        p_organization_id,
        p_method_type,
        p_method_name,
        p_secret,
        p_phone_number,
        p_email,
        p_metadata
    ) RETURNING id INTO method_id;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        p_user_id,
        'authentication'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        'REGISTER_MFA_METHOD',
        'user_mfa_methods',
        method_id::TEXT,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        'success',
        NULL, -- response_time
        jsonb_build_object(
            'user_id', p_user_id,
            'method_type', p_method_type,
            'method_name', p_method_name
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['authentication', 'mfa'], -- compliance_tags
        NULL, -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN method_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para criar uma sessão MFA
CREATE OR REPLACE FUNCTION iam.create_mfa_session(
    p_user_id UUID,
    p_organization_id UUID,
    p_ip_address VARCHAR DEFAULT NULL,
    p_user_agent TEXT DEFAULT NULL,
    p_session_duration_minutes INTEGER DEFAULT 15,
    p_metadata JSONB DEFAULT '{}'::JSONB
) RETURNS TABLE (
    session_id UUID,
    session_token TEXT,
    challenge_token TEXT,
    expires_at TIMESTAMP WITH TIME ZONE
) AS $$
DECLARE
    v_session_id UUID;
    v_session_token TEXT;
    v_challenge_token TEXT;
    v_expires_at TIMESTAMP WITH TIME ZONE;
BEGIN
    -- Gerar tokens
    v_session_token := encode(gen_random_bytes(32), 'hex');
    v_challenge_token := encode(gen_random_bytes(32), 'hex');
    v_expires_at := NOW() + (p_session_duration_minutes || ' minutes')::INTERVAL;
    
    -- Criar sessão
    INSERT INTO iam.mfa_sessions (
        user_id,
        organization_id,
        session_token,
        challenge_token,
        ip_address,
        user_agent,
        expires_at,
        metadata
    ) VALUES (
        p_user_id,
        p_organization_id,
        v_session_token,
        v_challenge_token,
        p_ip_address,
        p_user_agent,
        v_expires_at,
        p_metadata
    ) RETURNING id INTO v_session_id;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        p_user_id,
        'authentication'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        'CREATE_MFA_SESSION',
        'mfa_sessions',
        v_session_id::TEXT,
        p_ip_address,
        p_user_agent,
        NULL, -- request_id
        NULL, -- session_id
        'success',
        NULL, -- response_time
        jsonb_build_object(
            'user_id', p_user_id,
            'session_duration', p_session_duration_minutes
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['authentication', 'mfa'], -- compliance_tags
        NULL, -- regulatory_references
        NULL  -- geo_location
    );
    
    -- Retornar detalhes da sessão
    session_id := v_session_id;
    session_token := v_session_token;
    challenge_token := v_challenge_token;
    expires_at := v_expires_at;
    RETURN NEXT;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para verificar se um dispositivo é confiável
CREATE OR REPLACE FUNCTION iam.is_device_trusted(
    p_user_id UUID,
    p_organization_id UUID,
    p_device_identifier VARCHAR
) RETURNS BOOLEAN AS $$
DECLARE
    is_trusted BOOLEAN := FALSE;
BEGIN
    SELECT EXISTS (
        SELECT 1 
        FROM iam.trusted_devices
        WHERE user_id = p_user_id
          AND organization_id = p_organization_id
          AND device_identifier = p_device_identifier
          AND revoked = FALSE
          AND expires_at > NOW()
    ) INTO is_trusted;
    
    RETURN is_trusted;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para registrar um dispositivo confiável
CREATE OR REPLACE FUNCTION iam.register_trusted_device(
    p_user_id UUID,
    p_organization_id UUID,
    p_device_identifier VARCHAR,
    p_device_name VARCHAR DEFAULT NULL,
    p_ip_address VARCHAR DEFAULT NULL,
    p_user_agent TEXT DEFAULT NULL,
    p_trust_duration_days INTEGER DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    device_id UUID;
    trust_duration INTEGER;
BEGIN
    -- Se não foi especificado, usar a configuração da organização
    IF p_trust_duration_days IS NULL THEN
        SELECT COALESCE(remember_device_days, 30) INTO trust_duration
        FROM iam.mfa_organization_settings
        WHERE organization_id = p_organization_id;
    ELSE
        trust_duration := p_trust_duration_days;
    END IF;
    
    -- Atualizar se já existe ou inserir novo
    INSERT INTO iam.trusted_devices (
        user_id,
        organization_id,
        device_identifier,
        device_name,
        ip_address,
        user_agent,
        expires_at
    ) VALUES (
        p_user_id,
        p_organization_id,
        p_device_identifier,
        p_device_name,
        p_ip_address,
        p_user_agent,
        NOW() + (trust_duration || ' days')::INTERVAL
    )
    ON CONFLICT (user_id, device_identifier) 
    DO UPDATE SET
        device_name = EXCLUDED.device_name,
        ip_address = EXCLUDED.ip_address,
        user_agent = EXCLUDED.user_agent,
        last_used = NOW(),
        expires_at = NOW() + (trust_duration || ' days')::INTERVAL,
        revoked = FALSE,
        revoked_at = NULL
    RETURNING id INTO device_id;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        p_user_id,
        'authentication'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        'REGISTER_TRUSTED_DEVICE',
        'trusted_devices',
        device_id::TEXT,
        p_ip_address,
        p_user_agent,
        NULL, -- request_id
        NULL, -- session_id
        'success',
        NULL, -- response_time
        jsonb_build_object(
            'user_id', p_user_id,
            'device_name', p_device_name,
            'trust_duration_days', trust_duration
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['authentication', 'mfa'], -- compliance_tags
        NULL, -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN device_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para verificar se o usuário precisa de MFA
CREATE OR REPLACE FUNCTION iam.user_requires_mfa(
    p_user_id UUID,
    p_organization_id UUID,
    p_device_identifier VARCHAR DEFAULT NULL
) RETURNS BOOLEAN AS $$
DECLARE
    org_requires_mfa BOOLEAN;
    has_trusted_device BOOLEAN := FALSE;
    has_active_mfa_methods BOOLEAN;
BEGIN
    -- Verificar configuração da organização
    SELECT required_for_all INTO org_requires_mfa
    FROM iam.mfa_organization_settings
    WHERE organization_id = p_organization_id;
    
    -- Se não há configuração, usar padrão (False)
    IF org_requires_mfa IS NULL THEN
        org_requires_mfa := FALSE;
    END IF;
    
    -- Se a organização não exige MFA, não é necessário
    IF NOT org_requires_mfa THEN
        RETURN FALSE;
    END IF;
    
    -- Verificar se o dispositivo é confiável
    IF p_device_identifier IS NOT NULL THEN
        has_trusted_device := iam.is_device_trusted(p_user_id, p_organization_id, p_device_identifier);
        
        -- Se o dispositivo é confiável, não precisa de MFA
        IF has_trusted_device THEN
            RETURN FALSE;
        END IF;
    END IF;
    
    -- Verificar se o usuário tem métodos MFA ativos
    SELECT EXISTS (
        SELECT 1 
        FROM iam.user_mfa_methods
        WHERE user_id = p_user_id
          AND organization_id = p_organization_id
          AND status = 'enabled'
    ) INTO has_active_mfa_methods;
    
    -- Se o usuário não tem métodos MFA configurados, não pode ser exigido
    IF NOT has_active_mfa_methods THEN
        RETURN FALSE;
    END IF;
    
    -- Se chegou até aqui, MFA é necessário
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
