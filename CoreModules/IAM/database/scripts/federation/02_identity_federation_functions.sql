-- INNOVABIZ - IAM Identity Federation Functions
-- Author: Eduardo Jeremias
-- Date: 09/05/2025
-- Version: 1.0
-- Description: Funções para federação de identidades via SAML e OIDC

-- Configuração do esquema
SET search_path TO iam, public;

-- Função para registrar um provedor de identidade OIDC
CREATE OR REPLACE FUNCTION iam.register_oidc_provider(
    p_organization_id UUID,
    p_name VARCHAR,
    p_description TEXT,
    p_issuer_url VARCHAR,
    p_client_id VARCHAR,
    p_client_secret TEXT,
    p_authorization_endpoint VARCHAR,
    p_token_endpoint VARCHAR,
    p_userinfo_endpoint VARCHAR,
    p_jwks_uri VARCHAR,
    p_end_session_endpoint VARCHAR DEFAULT NULL,
    p_mapping_strategy iam.identity_mapping_strategy DEFAULT 'just_in_time_provisioning',
    p_config_metadata JSONB DEFAULT '{}'::JSONB
) RETURNS UUID AS $$
DECLARE
    provider_id UUID;
BEGIN
    INSERT INTO iam.identity_providers (
        organization_id,
        name,
        description,
        protocol,
        issuer_url,
        client_id,
        client_secret,
        authorization_endpoint,
        token_endpoint,
        userinfo_endpoint,
        jwks_uri,
        end_session_endpoint,
        mapping_strategy,
        status,
        config_metadata
    ) VALUES (
        p_organization_id,
        p_name,
        p_description,
        'oidc',
        p_issuer_url,
        p_client_id,
        p_client_secret,
        p_authorization_endpoint,
        p_token_endpoint,
        p_userinfo_endpoint,
        p_jwks_uri,
        p_end_session_endpoint,
        p_mapping_strategy,
        'inactive',
        p_config_metadata
    ) RETURNING id INTO provider_id;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        NULL, -- user_id (será preenchido pelo contexto de execução)
        'identity_management'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        'REGISTER_OIDC_PROVIDER',
        'identity_providers',
        provider_id::TEXT,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        'success',
        NULL, -- response_time
        jsonb_build_object(
            'provider_name', p_name,
            'issuer_url', p_issuer_url,
            'protocol', 'oidc',
            'mapping_strategy', p_mapping_strategy
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['identity_management', 'federation', 'oidc'], -- compliance_tags
        NULL, -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN provider_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para registrar um provedor de identidade SAML
CREATE OR REPLACE FUNCTION iam.register_saml_provider(
    p_organization_id UUID,
    p_name VARCHAR,
    p_description TEXT,
    p_issuer_url VARCHAR,
    p_metadata_url VARCHAR,
    p_certificate TEXT,
    p_private_key TEXT DEFAULT NULL,
    p_mapping_strategy iam.identity_mapping_strategy DEFAULT 'just_in_time_provisioning',
    p_config_metadata JSONB DEFAULT '{}'::JSONB
) RETURNS UUID AS $$
DECLARE
    provider_id UUID;
BEGIN
    INSERT INTO iam.identity_providers (
        organization_id,
        name,
        description,
        protocol,
        issuer_url,
        metadata_url,
        certificate,
        private_key,
        mapping_strategy,
        status,
        config_metadata
    ) VALUES (
        p_organization_id,
        p_name,
        p_description,
        'saml',
        p_issuer_url,
        p_metadata_url,
        p_certificate,
        p_private_key,
        p_mapping_strategy,
        'inactive',
        p_config_metadata
    ) RETURNING id INTO provider_id;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        NULL, -- user_id (será preenchido pelo contexto de execução)
        'identity_management'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        'REGISTER_SAML_PROVIDER',
        'identity_providers',
        provider_id::TEXT,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        'success',
        NULL, -- response_time
        jsonb_build_object(
            'provider_name', p_name,
            'issuer_url', p_issuer_url,
            'protocol', 'saml',
            'mapping_strategy', p_mapping_strategy
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['identity_management', 'federation', 'saml'], -- compliance_tags
        NULL, -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN provider_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para adicionar mapeamento de atributos ao provedor de identidade
CREATE OR REPLACE FUNCTION iam.add_attribute_mapping(
    p_provider_id UUID,
    p_organization_id UUID,
    p_external_attribute VARCHAR,
    p_internal_attribute VARCHAR,
    p_transformation_expression TEXT DEFAULT NULL,
    p_is_required BOOLEAN DEFAULT FALSE
) RETURNS UUID AS $$
DECLARE
    mapping_id UUID;
BEGIN
    INSERT INTO iam.identity_provider_attribute_mappings (
        provider_id,
        organization_id,
        external_attribute,
        internal_attribute,
        transformation_expression,
        is_required
    ) VALUES (
        p_provider_id,
        p_organization_id,
        p_external_attribute,
        p_internal_attribute,
        p_transformation_expression,
        p_is_required
    ) RETURNING id INTO mapping_id;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        NULL, -- user_id (será preenchido pelo contexto de execução)
        'identity_management'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        'ADD_ATTRIBUTE_MAPPING',
        'identity_provider_attribute_mappings',
        mapping_id::TEXT,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        'success',
        NULL, -- response_time
        jsonb_build_object(
            'provider_id', p_provider_id,
            'external_attribute', p_external_attribute,
            'internal_attribute', p_internal_attribute,
            'is_required', p_is_required
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['identity_management', 'federation'], -- compliance_tags
        NULL, -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN mapping_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para adicionar mapeamento de papel ao provedor de identidade
CREATE OR REPLACE FUNCTION iam.add_role_mapping(
    p_provider_id UUID,
    p_organization_id UUID,
    p_external_role VARCHAR,
    p_internal_role_id UUID,
    p_mapping_condition JSONB DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    mapping_id UUID;
BEGIN
    INSERT INTO iam.identity_provider_role_mappings (
        provider_id,
        organization_id,
        external_role,
        internal_role_id,
        mapping_condition
    ) VALUES (
        p_provider_id,
        p_organization_id,
        p_external_role,
        p_internal_role_id,
        p_mapping_condition
    ) RETURNING id INTO mapping_id;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        NULL, -- user_id (será preenchido pelo contexto de execução)
        'identity_management'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        'ADD_ROLE_MAPPING',
        'identity_provider_role_mappings',
        mapping_id::TEXT,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        'success',
        NULL, -- response_time
        jsonb_build_object(
            'provider_id', p_provider_id,
            'external_role', p_external_role,
            'internal_role_id', p_internal_role_id
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['identity_management', 'federation'], -- compliance_tags
        NULL, -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN mapping_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para adicionar grupo federado
CREATE OR REPLACE FUNCTION iam.add_federated_group(
    p_provider_id UUID,
    p_organization_id UUID,
    p_external_group_id VARCHAR,
    p_external_group_name VARCHAR,
    p_internal_role_id UUID DEFAULT NULL,
    p_auto_role_assignment BOOLEAN DEFAULT FALSE
) RETURNS UUID AS $$
DECLARE
    group_id UUID;
BEGIN
    INSERT INTO iam.federated_groups (
        provider_id,
        organization_id,
        external_group_id,
        external_group_name,
        internal_role_id,
        auto_role_assignment
    ) VALUES (
        p_provider_id,
        p_organization_id,
        p_external_group_id,
        p_external_group_name,
        p_internal_role_id,
        p_auto_role_assignment
    ) RETURNING id INTO group_id;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        NULL, -- user_id (será preenchido pelo contexto de execução)
        'identity_management'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        'ADD_FEDERATED_GROUP',
        'federated_groups',
        group_id::TEXT,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        'success',
        NULL, -- response_time
        jsonb_build_object(
            'provider_id', p_provider_id,
            'external_group_id', p_external_group_id,
            'external_group_name', p_external_group_name,
            'internal_role_id', p_internal_role_id,
            'auto_role_assignment', p_auto_role_assignment
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['identity_management', 'federation'], -- compliance_tags
        NULL, -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN group_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para criar ou atualizar uma identidade federada
CREATE OR REPLACE FUNCTION iam.create_or_update_federated_identity(
    p_provider_id UUID,
    p_organization_id UUID,
    p_external_id VARCHAR,
    p_external_username VARCHAR,
    p_external_email VARCHAR,
    p_external_data JSONB,
    p_existing_user_id UUID DEFAULT NULL
) RETURNS TABLE (
    user_id UUID,
    federated_identity_id UUID,
    is_new_user BOOLEAN
) AS $$
DECLARE
    v_user_id UUID;
    v_federated_identity_id UUID;
    v_is_new_user BOOLEAN := FALSE;
    v_provider_record RECORD;
    v_username VARCHAR;
    v_email VARCHAR;
    v_full_name VARCHAR;
BEGIN
    -- Obter informações do provedor
    SELECT name, mapping_strategy 
    INTO v_provider_record
    FROM iam.identity_providers
    WHERE id = p_provider_id;
    
    -- Verificar se já existe identidade federada
    SELECT fi.id, fi.user_id
    INTO v_federated_identity_id, v_user_id
    FROM iam.federated_identities fi
    WHERE fi.provider_id = p_provider_id
      AND fi.external_id = p_external_id;
    
    IF v_federated_identity_id IS NOT NULL THEN
        -- Atualizar identidade federada existente
        UPDATE iam.federated_identities
        SET external_username = p_external_username,
            external_email = p_external_email,
            last_login = NOW(),
            external_data = p_external_data,
            updated_at = NOW()
        WHERE id = v_federated_identity_id;
    ELSE
        -- Determinar user_id
        IF p_existing_user_id IS NOT NULL THEN
            -- Usar ID de usuário existente fornecido
            v_user_id := p_existing_user_id;
        ELSE
            -- Verificar estratégia de mapeamento
            IF v_provider_record.mapping_strategy = 'pre_provisioned' THEN
                -- Tentar encontrar usuário pré-provisionado pelo email
                SELECT id INTO v_user_id
                FROM iam.users
                WHERE organization_id = p_organization_id
                  AND email = p_external_email
                  AND status = 'active';
                  
                IF v_user_id IS NULL THEN
                    RAISE EXCEPTION 'Usuário pré-provisionado não encontrado para email: %', p_external_email;
                END IF;
            ELSIF v_provider_record.mapping_strategy = 'just_in_time_provisioning' THEN
                -- Tentar encontrar usuário existente pelo email
                SELECT id INTO v_user_id
                FROM iam.users
                WHERE organization_id = p_organization_id
                  AND email = p_external_email;
                  
                IF v_user_id IS NULL THEN
                    -- Criar novo usuário
                    v_username := COALESCE(p_external_username, p_external_email);
                    v_email := p_external_email;
                    v_full_name := COALESCE(p_external_data->>'name', p_external_username, p_external_email);
                    
                    INSERT INTO iam.users (
                        organization_id,
                        username,
                        email,
                        full_name,
                        password_hash, -- Hash genérico, não será usado para login
                        status,
                        metadata
                    ) VALUES (
                        p_organization_id,
                        v_username,
                        v_email,
                        v_full_name,
                        crypt('FEDERATED_IDENTITY_NO_PASSWORD', gen_salt('bf')),
                        'active',
                        jsonb_build_object(
                            'source', 'federated',
                            'provider', v_provider_record.name,
                            'created_at', NOW()
                        )
                    ) RETURNING id INTO v_user_id;
                    
                    v_is_new_user := TRUE;
                END IF;
            ELSE
                RAISE EXCEPTION 'Estratégia de mapeamento não implementada: %', v_provider_record.mapping_strategy;
            END IF;
        END IF;
        
        -- Criar nova identidade federada
        INSERT INTO iam.federated_identities (
            user_id,
            provider_id,
            organization_id,
            external_id,
            external_username,
            external_email,
            last_login,
            external_data
        ) VALUES (
            v_user_id,
            p_provider_id,
            p_organization_id,
            p_external_id,
            p_external_username,
            p_external_email,
            NOW(),
            p_external_data
        ) RETURNING id INTO v_federated_identity_id;
    END IF;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        v_user_id,
        'authentication'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        CASE WHEN v_is_new_user THEN 'CREATE_FEDERATED_USER' ELSE 'UPDATE_FEDERATED_IDENTITY' END,
        'federated_identities',
        v_federated_identity_id::TEXT,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        'success',
        NULL, -- response_time
        jsonb_build_object(
            'provider_id', p_provider_id,
            'external_id', p_external_id,
            'external_email', p_external_email,
            'is_new_user', v_is_new_user
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['authentication', 'federation'], -- compliance_tags
        NULL, -- regulatory_references
        NULL  -- geo_location
    );
    
    -- Retornar informações
    user_id := v_user_id;
    federated_identity_id := v_federated_identity_id;
    is_new_user := v_is_new_user;
    RETURN NEXT;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para atribuir grupos federados a uma identidade federada
CREATE OR REPLACE FUNCTION iam.assign_federated_groups(
    p_federated_identity_id UUID,
    p_organization_id UUID,
    p_external_group_ids VARCHAR[]
) RETURNS INTEGER AS $$
DECLARE
    group_record RECORD;
    groups_count INTEGER := 0;
    v_user_id UUID;
BEGIN
    -- Obter user_id da identidade federada
    SELECT user_id INTO v_user_id
    FROM iam.federated_identities
    WHERE id = p_federated_identity_id;
    
    IF v_user_id IS NULL THEN
        RAISE EXCEPTION 'Identidade federada não encontrada: %', p_federated_identity_id;
    END IF;
    
    -- Limpar associações de grupos existentes para esta identidade
    DELETE FROM iam.federated_user_groups
    WHERE federated_identity_id = p_federated_identity_id;
    
    -- Para cada grupo externo, encontrar o grupo federado correspondente
    FOR group_record IN (
        SELECT fg.id, fg.internal_role_id, fg.auto_role_assignment
        FROM iam.federated_groups fg
        WHERE fg.organization_id = p_organization_id
          AND fg.external_group_id = ANY(p_external_group_ids)
    ) LOOP
        -- Adicionar à tabela de relação
        INSERT INTO iam.federated_user_groups (
            federated_identity_id,
            federated_group_id,
            organization_id
        ) VALUES (
            p_federated_identity_id,
            group_record.id,
            p_organization_id
        );
        
        groups_count := groups_count + 1;
        
        -- Se o grupo tiver atribuição automática de papel, atribuir o papel ao usuário
        IF group_record.auto_role_assignment AND group_record.internal_role_id IS NOT NULL THEN
            -- Verificar se o papel já está atribuído
            IF NOT EXISTS (
                SELECT 1 
                FROM iam.user_roles 
                WHERE user_id = v_user_id 
                  AND role_id = group_record.internal_role_id
            ) THEN
                -- Atribuir papel ao usuário
                PERFORM iam.assign_role_to_user(
                    p_organization_id,
                    v_user_id,
                    group_record.internal_role_id,
                    NULL, -- created_by
                    NULL, -- expires_at
                    jsonb_build_object(
                        'source', 'federated_group',
                        'auto_assigned', TRUE,
                        'assigned_at', NOW()
                    )
                );
            END IF;
        END IF;
    END LOOP;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        v_user_id,
        'identity_management'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        'ASSIGN_FEDERATED_GROUPS',
        'federated_user_groups',
        p_federated_identity_id::TEXT,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        'success',
        NULL, -- response_time
        jsonb_build_object(
            'federated_identity_id', p_federated_identity_id,
            'groups_count', groups_count,
            'external_group_ids', p_external_group_ids
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['identity_management', 'federation'], -- compliance_tags
        NULL, -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN groups_count;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para criar uma sessão federada
CREATE OR REPLACE FUNCTION iam.create_federation_session(
    p_user_id UUID,
    p_provider_id UUID,
    p_organization_id UUID,
    p_external_session_id VARCHAR DEFAULT NULL,
    p_ip_address VARCHAR DEFAULT NULL,
    p_user_agent TEXT DEFAULT NULL,
    p_session_duration_minutes INTEGER DEFAULT 60,
    p_metadata JSONB DEFAULT '{}'::JSONB
) RETURNS TABLE (
    session_id UUID,
    session_token TEXT,
    expires_at TIMESTAMP WITH TIME ZONE
) AS $$
DECLARE
    v_session_id UUID;
    v_session_token TEXT;
    v_expires_at TIMESTAMP WITH TIME ZONE;
BEGIN
    -- Gerar token de sessão
    v_session_token := encode(gen_random_bytes(32), 'hex');
    v_expires_at := NOW() + (p_session_duration_minutes || ' minutes')::INTERVAL;
    
    -- Criar sessão
    INSERT INTO iam.federation_sessions (
        user_id,
        provider_id,
        organization_id,
        session_token,
        external_session_id,
        ip_address,
        user_agent,
        expires_at,
        metadata
    ) VALUES (
        p_user_id,
        p_provider_id,
        p_organization_id,
        v_session_token,
        p_external_session_id,
        p_ip_address,
        p_user_agent,
        v_expires_at,
        p_metadata
    ) RETURNING id INTO v_session_id;
    
    -- Atualizar último login na identidade federada
    UPDATE iam.federated_identities
    SET last_login = NOW()
    WHERE user_id = p_user_id
      AND provider_id = p_provider_id;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        p_user_id,
        'authentication'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        'CREATE_FEDERATION_SESSION',
        'federation_sessions',
        v_session_id::TEXT,
        p_ip_address,
        p_user_agent,
        NULL, -- request_id
        v_session_token, -- session_id
        'success',
        NULL, -- response_time
        jsonb_build_object(
            'provider_id', p_provider_id,
            'session_duration', p_session_duration_minutes
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['authentication', 'federation'], -- compliance_tags
        NULL, -- regulatory_references
        NULL  -- geo_location
    );
    
    -- Retornar detalhes da sessão
    session_id := v_session_id;
    session_token := v_session_token;
    expires_at := v_expires_at;
    RETURN NEXT;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para verificar sessão federada
CREATE OR REPLACE FUNCTION iam.validate_federation_session(
    p_session_token TEXT
) RETURNS TABLE (
    is_valid BOOLEAN,
    user_id UUID,
    organization_id UUID,
    provider_id UUID,
    session_id UUID,
    metadata JSONB
) AS $$
DECLARE
    v_session RECORD;
BEGIN
    -- Buscar sessão
    SELECT 
        fs.id,
        fs.user_id,
        fs.organization_id,
        fs.provider_id,
        fs.revoked,
        fs.expires_at,
        fs.metadata
    INTO v_session
    FROM iam.federation_sessions fs
    WHERE fs.session_token = p_session_token;
    
    -- Verificar se sessão existe e é válida
    IF v_session.id IS NULL THEN
        is_valid := FALSE;
        user_id := NULL;
        organization_id := NULL;
        provider_id := NULL;
        session_id := NULL;
        metadata := NULL;
    ELSIF v_session.revoked OR v_session.expires_at < NOW() THEN
        is_valid := FALSE;
        user_id := v_session.user_id;
        organization_id := v_session.organization_id;
        provider_id := v_session.provider_id;
        session_id := v_session.id;
        metadata := v_session.metadata;
        
        -- Registrar auditoria para sessão inválida
        PERFORM iam.log_audit_event(
            v_session.organization_id,
            v_session.user_id,
            'authentication'::iam.audit_event_category,
            'medium'::iam.audit_severity_level,
            'INVALID_FEDERATION_SESSION',
            'federation_sessions',
            v_session.id::TEXT,
            NULL, -- source_ip
            NULL, -- user_agent
            NULL, -- request_id
            p_session_token, -- session_id
            'denied',
            NULL, -- response_time
            jsonb_build_object(
                'reason', CASE 
                    WHEN v_session.revoked THEN 'session_revoked'
                    ELSE 'session_expired'
                END,
                'expires_at', v_session.expires_at
            ),
            NULL, -- request_payload
            NULL, -- response_payload
            ARRAY['authentication', 'federation'], -- compliance_tags
            NULL, -- regulatory_references
            NULL  -- geo_location
        );
    ELSE
        is_valid := TRUE;
        user_id := v_session.user_id;
        organization_id := v_session.organization_id;
        provider_id := v_session.provider_id;
        session_id := v_session.id;
        metadata := v_session.metadata;
        
        -- Registrar auditoria para sessão válida
        PERFORM iam.log_audit_event(
            v_session.organization_id,
            v_session.user_id,
            'authentication'::iam.audit_event_category,
            'info'::iam.audit_severity_level,
            'VALIDATE_FEDERATION_SESSION',
            'federation_sessions',
            v_session.id::TEXT,
            NULL, -- source_ip
            NULL, -- user_agent
            NULL, -- request_id
            p_session_token, -- session_id
            'success',
            NULL, -- response_time
            jsonb_build_object(
                'expires_at', v_session.expires_at
            ),
            NULL, -- request_payload
            NULL, -- response_payload
            ARRAY['authentication', 'federation'], -- compliance_tags
            NULL, -- regulatory_references
            NULL  -- geo_location
        );
    END IF;
    
    RETURN NEXT;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para obter os papéis de um usuário federado, combinando papéis diretos e de grupos
CREATE OR REPLACE FUNCTION iam.get_federated_user_roles(
    p_user_id UUID,
    p_organization_id UUID
) RETURNS TABLE (
    role_id UUID,
    role_code VARCHAR,
    role_name VARCHAR,
    role_source VARCHAR,
    assigned_via VARCHAR
) AS $$
BEGIN
    -- Papéis atribuídos diretamente ao usuário
    RETURN QUERY
    SELECT 
        dr.id AS role_id,
        dr.code AS role_code,
        dr.name AS role_name,
        'direct'::VARCHAR AS role_source,
        'user_assignment'::VARCHAR AS assigned_via
    FROM iam.user_roles ur
    JOIN iam.detailed_roles dr ON ur.role_id = dr.id
    WHERE ur.user_id = p_user_id
      AND ur.organization_id = p_organization_id
      AND (ur.expires_at IS NULL OR ur.expires_at > NOW());
    
    -- Papéis atribuídos via grupos federados
    RETURN QUERY
    SELECT 
        dr.id AS role_id,
        dr.code AS role_code,
        dr.name AS role_name,
        'federated'::VARCHAR AS role_source,
        'group_' || fg.external_group_name AS assigned_via
    FROM iam.federated_identities fi
    JOIN iam.federated_user_groups fug ON fi.id = fug.federated_identity_id
    JOIN iam.federated_groups fg ON fug.federated_group_id = fg.id
    JOIN iam.detailed_roles dr ON fg.internal_role_id = dr.id
    WHERE fi.user_id = p_user_id
      AND fi.organization_id = p_organization_id
      AND fg.internal_role_id IS NOT NULL;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
