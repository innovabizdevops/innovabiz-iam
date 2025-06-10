-- INNOVABIZ - IAM MFA Authentication - Integração AR/VR
-- Author: Eduardo Jeremias
-- Date: 09/05/2025
-- Version: 1.0
-- Description: Integração do sistema de autenticação multi-fator com tecnologias AR/VR

-- Configuração do esquema
SET search_path TO iam, public;

-- Tipos enumerados para métodos AR/VR específicos
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ar_auth_gesture_type') THEN
        CREATE TYPE iam.ar_auth_gesture_type AS ENUM (
            'hand_movement',
            'head_movement',
            'controller_pattern',
            'combined_spatial',
            'object_interaction',
            'virtual_keyboard'
        );
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ar_auth_gaze_type') THEN
        CREATE TYPE iam.ar_auth_gaze_type AS ENUM (
            'sequential_targets',
            'timed_pattern',
            'pursuit_tracking',
            'recognition_based',
            'dwell_time_sequence'
        );
    END IF;
END$$;

-- Tabela para padrões de gestos espaciais 3D
CREATE TABLE IF NOT EXISTS iam.user_ar_gesture_auth (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam.users(id),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    mfa_method_id UUID REFERENCES iam.user_mfa_methods(id),
    gesture_name VARCHAR(100) NOT NULL,
    gesture_type iam.ar_auth_gesture_type NOT NULL,
    gesture_data BYTEA NOT NULL, -- Dados criptografados do padrão
    gesture_hash VARCHAR(255) NOT NULL, -- Hash para verificação
    complexity_score INTEGER NOT NULL, -- Pontuação de 1-100 para complexidade
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used TIMESTAMP WITH TIME ZONE,
    status iam.mfa_status DEFAULT 'enabled',
    metadata JSONB DEFAULT '{}'::JSONB,
    UNIQUE(user_id, gesture_name)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_user_ar_gesture_auth_user_id ON iam.user_ar_gesture_auth(user_id);
CREATE INDEX IF NOT EXISTS idx_user_ar_gesture_auth_organization_id ON iam.user_ar_gesture_auth(organization_id);
CREATE INDEX IF NOT EXISTS idx_user_ar_gesture_auth_mfa_method_id ON iam.user_ar_gesture_auth(mfa_method_id);
CREATE INDEX IF NOT EXISTS idx_user_ar_gesture_auth_status ON iam.user_ar_gesture_auth(status);

-- Tabela para padrões de olhar (gaze)
CREATE TABLE IF NOT EXISTS iam.user_ar_gaze_auth (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam.users(id),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    mfa_method_id UUID REFERENCES iam.user_mfa_methods(id),
    pattern_name VARCHAR(100) NOT NULL,
    gaze_type iam.ar_auth_gaze_type NOT NULL,
    pattern_data BYTEA NOT NULL, -- Dados criptografados do padrão
    pattern_hash VARCHAR(255) NOT NULL, -- Hash para verificação
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used TIMESTAMP WITH TIME ZONE,
    status iam.mfa_status DEFAULT 'enabled',
    metadata JSONB DEFAULT '{}'::JSONB,
    UNIQUE(user_id, pattern_name)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_user_ar_gaze_auth_user_id ON iam.user_ar_gaze_auth(user_id);
CREATE INDEX IF NOT EXISTS idx_user_ar_gaze_auth_organization_id ON iam.user_ar_gaze_auth(organization_id);
CREATE INDEX IF NOT EXISTS idx_user_ar_gaze_auth_mfa_method_id ON iam.user_ar_gaze_auth(mfa_method_id);
CREATE INDEX IF NOT EXISTS idx_user_ar_gaze_auth_status ON iam.user_ar_gaze_auth(status);

-- Tabela para senhas espaciais
CREATE TABLE IF NOT EXISTS iam.user_ar_spatial_password (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam.users(id),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    mfa_method_id UUID REFERENCES iam.user_mfa_methods(id),
    password_name VARCHAR(100) NOT NULL,
    password_data BYTEA NOT NULL, -- Dados criptografados da senha espacial
    password_hash VARCHAR(255) NOT NULL, -- Hash para verificação
    dimension_count INTEGER NOT NULL, -- Número de dimensões (3D, 4D, etc)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used TIMESTAMP WITH TIME ZONE,
    status iam.mfa_status DEFAULT 'enabled',
    complexity_score INTEGER NOT NULL, -- Pontuação de 1-100 para complexidade
    metadata JSONB DEFAULT '{}'::JSONB,
    UNIQUE(user_id, password_name)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_user_ar_spatial_password_user_id ON iam.user_ar_spatial_password(user_id);
CREATE INDEX IF NOT EXISTS idx_user_ar_spatial_password_organization_id ON iam.user_ar_spatial_password(organization_id);
CREATE INDEX IF NOT EXISTS idx_user_ar_spatial_password_mfa_method_id ON iam.user_ar_spatial_password(mfa_method_id);
CREATE INDEX IF NOT EXISTS idx_user_ar_spatial_password_status ON iam.user_ar_spatial_password(status);

-- Tabela para sessões de autenticação contínua contextual
CREATE TABLE IF NOT EXISTS iam.ar_continuous_auth_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES iam.users(id),
    organization_id UUID NOT NULL REFERENCES iam.organizations(id),
    session_id UUID NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    confidence_score FLOAT NOT NULL, -- 0.0 a 1.0
    last_verification TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    revoked_reason VARCHAR(100),
    metadata JSONB DEFAULT '{}'::JSONB
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_ar_continuous_auth_sessions_user_id ON iam.ar_continuous_auth_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_ar_continuous_auth_sessions_organization_id ON iam.ar_continuous_auth_sessions(organization_id);
CREATE INDEX IF NOT EXISTS idx_ar_continuous_auth_sessions_session_id ON iam.ar_continuous_auth_sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_ar_continuous_auth_sessions_confidence ON iam.ar_continuous_auth_sessions(confidence_score);
CREATE INDEX IF NOT EXISTS idx_ar_continuous_auth_sessions_expires ON iam.ar_continuous_auth_sessions(expires_at);

-- Função para registrar padrão de gesto espacial 3D
CREATE OR REPLACE FUNCTION iam.register_ar_gesture(
    p_user_id UUID,
    p_organization_id UUID,
    p_gesture_name VARCHAR,
    p_gesture_type iam.ar_auth_gesture_type,
    p_gesture_data BYTEA,
    p_complexity_score INTEGER DEFAULT 50,
    p_metadata JSONB DEFAULT '{}'::JSONB
) RETURNS UUID AS $$
DECLARE
    gesture_id UUID;
    mfa_method_id UUID;
    gesture_hash VARCHAR;
BEGIN
    -- Gerar hash para o padrão de gesto
    gesture_hash := encode(digest(p_gesture_data, 'sha256'), 'hex');
    
    -- Verificar ou criar método MFA correspondente
    SELECT id INTO mfa_method_id 
    FROM iam.user_mfa_methods
    WHERE user_id = p_user_id
      AND organization_id = p_organization_id
      AND method_type = 'ar_spatial_gesture';
      
    IF mfa_method_id IS NULL THEN
        -- Criar novo método MFA
        mfa_method_id := iam.register_mfa_method(
            p_user_id,
            p_organization_id,
            'ar_spatial_gesture',
            p_gesture_name,
            NULL, -- secret
            NULL, -- phone
            NULL, -- email
            jsonb_build_object('gesture_type', p_gesture_type)
        );
        
        -- Atualizar status para ativo
        UPDATE iam.user_mfa_methods
        SET status = 'enabled'
        WHERE id = mfa_method_id;
    END IF;
    
    -- Inserir padrão de gesto
    INSERT INTO iam.user_ar_gesture_auth (
        user_id,
        organization_id,
        mfa_method_id,
        gesture_name,
        gesture_type,
        gesture_data,
        gesture_hash,
        complexity_score,
        metadata
    ) VALUES (
        p_user_id,
        p_organization_id,
        mfa_method_id,
        p_gesture_name,
        p_gesture_type,
        p_gesture_data,
        gesture_hash,
        p_complexity_score,
        p_metadata
    ) RETURNING id INTO gesture_id;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        p_user_id,
        'authentication'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        'REGISTER_AR_GESTURE',
        'user_ar_gesture_auth',
        gesture_id::TEXT,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        'success',
        NULL, -- response_time
        jsonb_build_object(
            'gesture_name', p_gesture_name,
            'gesture_type', p_gesture_type,
            'complexity_score', p_complexity_score
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['authentication', 'mfa', 'ar_vr'], -- compliance_tags
        ARRAY['IEEE2888', 'NIST.SP.800-63-3'], -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN gesture_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para registrar padrão de olhar (gaze)
CREATE OR REPLACE FUNCTION iam.register_ar_gaze_pattern(
    p_user_id UUID,
    p_organization_id UUID,
    p_pattern_name VARCHAR,
    p_gaze_type iam.ar_auth_gaze_type,
    p_pattern_data BYTEA,
    p_metadata JSONB DEFAULT '{}'::JSONB
) RETURNS UUID AS $$
DECLARE
    pattern_id UUID;
    mfa_method_id UUID;
    pattern_hash VARCHAR;
BEGIN
    -- Gerar hash para o padrão de olhar
    pattern_hash := encode(digest(p_pattern_data, 'sha256'), 'hex');
    
    -- Verificar ou criar método MFA correspondente
    SELECT id INTO mfa_method_id 
    FROM iam.user_mfa_methods
    WHERE user_id = p_user_id
      AND organization_id = p_organization_id
      AND method_type = 'ar_gaze_pattern';
      
    IF mfa_method_id IS NULL THEN
        -- Criar novo método MFA
        mfa_method_id := iam.register_mfa_method(
            p_user_id,
            p_organization_id,
            'ar_gaze_pattern',
            p_pattern_name,
            NULL, -- secret
            NULL, -- phone
            NULL, -- email
            jsonb_build_object('gaze_type', p_gaze_type)
        );
        
        -- Atualizar status para ativo
        UPDATE iam.user_mfa_methods
        SET status = 'enabled'
        WHERE id = mfa_method_id;
    END IF;
    
    -- Inserir padrão de olhar
    INSERT INTO iam.user_ar_gaze_auth (
        user_id,
        organization_id,
        mfa_method_id,
        pattern_name,
        gaze_type,
        pattern_data,
        pattern_hash,
        metadata
    ) VALUES (
        p_user_id,
        p_organization_id,
        mfa_method_id,
        p_pattern_name,
        p_gaze_type,
        p_pattern_data,
        pattern_hash,
        p_metadata
    ) RETURNING id INTO pattern_id;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        p_user_id,
        'authentication'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        'REGISTER_AR_GAZE_PATTERN',
        'user_ar_gaze_auth',
        pattern_id::TEXT,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        'success',
        NULL, -- response_time
        jsonb_build_object(
            'pattern_name', p_pattern_name,
            'gaze_type', p_gaze_type
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['authentication', 'mfa', 'ar_vr'], -- compliance_tags
        ARRAY['IEEE2888', 'NIST.SP.800-63-3'], -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN pattern_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para iniciar sessão de autenticação contínua contextual
CREATE OR REPLACE FUNCTION iam.start_ar_continuous_auth(
    p_user_id UUID,
    p_organization_id UUID,
    p_session_id UUID,
    p_device_id VARCHAR,
    p_initial_confidence FLOAT DEFAULT 1.0,
    p_session_duration_hours INTEGER DEFAULT 4,
    p_metadata JSONB DEFAULT '{}'::JSONB
) RETURNS UUID AS $$
DECLARE
    auth_session_id UUID;
BEGIN
    -- Validar confiança inicial
    IF p_initial_confidence < 0.0 OR p_initial_confidence > 1.0 THEN
        RAISE EXCEPTION 'Valor de confiança deve estar entre 0.0 e 1.0';
    END IF;
    
    -- Criar sessão de autenticação contínua
    INSERT INTO iam.ar_continuous_auth_sessions (
        user_id,
        organization_id,
        session_id,
        device_id,
        confidence_score,
        last_verification,
        expires_at,
        metadata
    ) VALUES (
        p_user_id,
        p_organization_id,
        p_session_id,
        p_device_id,
        p_initial_confidence,
        NOW(),
        NOW() + (p_session_duration_hours || ' hours')::INTERVAL,
        p_metadata
    ) RETURNING id INTO auth_session_id;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        p_user_id,
        'authentication'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        'START_AR_CONTINUOUS_AUTH',
        'ar_continuous_auth_sessions',
        auth_session_id::TEXT,
        NULL, -- source_ip
        NULL, -- user_agent
        p_session_id, -- request_id
        p_session_id, -- session_id
        'success',
        NULL, -- response_time
        jsonb_build_object(
            'initial_confidence', p_initial_confidence,
            'session_duration_hours', p_session_duration_hours,
            'device_id', p_device_id
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['authentication', 'mfa', 'ar_vr', 'continuous_auth'], -- compliance_tags
        ARRAY['IEEE2888', 'NIST.SP.800-63-3'], -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN auth_session_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para atualizar pontuação de confiança da autenticação contínua
CREATE OR REPLACE FUNCTION iam.update_ar_auth_confidence(
    p_session_id UUID,
    p_confidence_update FLOAT,
    p_reason VARCHAR DEFAULT NULL
) RETURNS FLOAT AS $$
DECLARE
    current_confidence FLOAT;
    new_confidence FLOAT;
    v_user_id UUID;
    v_organization_id UUID;
BEGIN
    -- Obter detalhes da sessão atual
    SELECT 
        confidence_score, 
        user_id, 
        organization_id 
    INTO 
        current_confidence, 
        v_user_id, 
        v_organization_id
    FROM iam.ar_continuous_auth_sessions
    WHERE session_id = p_session_id
      AND revoked = FALSE
      AND expires_at > NOW();
      
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Sessão de autenticação contínua não encontrada ou expirada';
    END IF;
    
    -- Calcular nova pontuação de confiança
    new_confidence := GREATEST(0.0, LEAST(1.0, current_confidence + p_confidence_update));
    
    -- Atualizar sessão
    UPDATE iam.ar_continuous_auth_sessions
    SET 
        confidence_score = new_confidence,
        last_verification = NOW(),
        metadata = jsonb_set(
            metadata, 
            '{confidence_history}', 
            COALESCE(metadata->'confidence_history', '[]'::jsonb) || 
            jsonb_build_object(
                'timestamp', NOW(),
                'previous', current_confidence,
                'new', new_confidence,
                'change', p_confidence_update,
                'reason', p_reason
            )
        )
    WHERE session_id = p_session_id;
    
    -- Registrar auditoria se houver alteração significativa
    IF ABS(current_confidence - new_confidence) >= 0.1 THEN
        PERFORM iam.log_audit_event(
            v_organization_id,
            v_user_id,
            'authentication'::iam.audit_event_category,
            CASE 
                WHEN new_confidence < 0.3 THEN 'high'::iam.audit_severity_level
                WHEN new_confidence < 0.6 THEN 'medium'::iam.audit_severity_level
                ELSE 'info'::iam.audit_severity_level
            END,
            'UPDATE_AR_AUTH_CONFIDENCE',
            'ar_continuous_auth_sessions',
            p_session_id::TEXT,
            NULL, -- source_ip
            NULL, -- user_agent
            p_session_id, -- request_id
            p_session_id, -- session_id
            'success',
            NULL, -- response_time
            jsonb_build_object(
                'previous_confidence', current_confidence,
                'new_confidence', new_confidence,
                'change', p_confidence_update,
                'reason', p_reason
            ),
            NULL, -- request_payload
            NULL, -- response_payload
            ARRAY['authentication', 'mfa', 'ar_vr', 'continuous_auth'], -- compliance_tags
            NULL, -- regulatory_references
            NULL  -- geo_location
        );
    END IF;
    
    RETURN new_confidence;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
