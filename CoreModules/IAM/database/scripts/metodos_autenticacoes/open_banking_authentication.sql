-- Métodos de Open Banking (OB-15-01 a OB-15-06)

-- 1. Redirectionamento Seguro com CIBA
CREATE OR REPLACE FUNCTION openbanking.verify_ciba_redirect(
    p_request_data JSONB,
    p_transaction_id TEXT,
    p_client_id TEXT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar integridade dos dados
    IF NOT EXISTS (
        SELECT 1 FROM jsonb_object_keys(p_request_data)
        WHERE value IN ('transaction_id', 'client_id', 'redirect_uri', 'state')
    ) THEN
        RETURN FALSE;
    END IF;

    -- Verificar validade da transação
    IF NOT EXISTS (
        SELECT 1 FROM transactions 
        WHERE transaction_id = p_transaction_id 
        AND client_id = p_client_id 
        AND status = 'pending'
    ) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- 2. Autenticação Decoupled (App-to-App)
CREATE OR REPLACE FUNCTION openbanking.verify_app_to_app(
    p_app_data JSONB,
    p_app_id TEXT,
    p_session_id TEXT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar integridade do app
    IF NOT EXISTS (
        SELECT 1 FROM jsonb_object_keys(p_app_data)
        WHERE value IN ('app_id', 'session_id', 'permissions', 'timestamp')
    ) THEN
        RETURN FALSE;
    END IF;

    -- Verificar sessão ativa
    IF NOT EXISTS (
        SELECT 1 FROM sessions 
        WHERE app_id = p_app_id 
        AND session_id = p_session_id 
        AND status = 'active'
    ) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- 3. Validação com Certificado eIDAS QTSP
CREATE OR REPLACE FUNCTION openbanking.verify_eidas_qtsp(
    p_certificate_data JSONB,
    p_certificate_id TEXT,
    p_trust_service TEXT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar integridade do certificado
    IF NOT EXISTS (
        SELECT 1 FROM jsonb_object_keys(p_certificate_data)
        WHERE value IN ('certificate_id', 'trust_service', 'valid_until', 'status')
    ) THEN
        RETURN FALSE;
    END IF;

    -- Verificar validade do certificado
    IF NOT EXISTS (
        SELECT 1 FROM certificates 
        WHERE certificate_id = p_certificate_id 
        AND trust_service = p_trust_service 
        AND valid_until > CURRENT_DATE
    ) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- 4. MFA Adaptativo para Transações Financeiras
CREATE OR REPLACE FUNCTION openbanking.verify_adaptive_mfa(
    p_transaction_data JSONB,
    p_amount FLOAT,
    p_risk_score FLOAT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar risco da transação
    IF p_risk_score > 0.7 THEN
        -- Requer MFA forte para transações de alto risco
        IF NOT EXISTS (
            SELECT 1 FROM jsonb_array_elements_text(p_transaction_data->'factors')
            WHERE value::text IN ('biometric', 'hardware_token', 'face_recognition')
        ) THEN
            RETURN FALSE;
        END IF;
    END IF;

    -- Verificar integridade dos dados
    IF NOT EXISTS (
        SELECT 1 FROM jsonb_object_keys(p_transaction_data)
        WHERE value IN ('amount', 'risk_score', 'factors', 'timestamp')
    ) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- 5. Confirmação de Token Binding
CREATE OR REPLACE FUNCTION openbanking.verify_token_binding(
    p_binding_data JSONB,
    p_token_id TEXT,
    p_device_id TEXT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar integridade do binding
    IF NOT EXISTS (
        SELECT 1 FROM jsonb_object_keys(p_binding_data)
        WHERE value IN ('token_id', 'device_id', 'signature', 'timestamp')
    ) THEN
        RETURN FALSE;
    END IF;

    -- Verificar vinculação válida
    IF NOT EXISTS (
        SELECT 1 FROM bindings 
        WHERE token_id = p_token_id 
        AND device_id = p_device_id 
        AND status = 'active'
    ) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- 6. Validação de Padrão de Transação
CREATE OR REPLACE FUNCTION openbanking.verify_transaction_pattern(
    p_pattern_data JSONB,
    p_transaction_id TEXT,
    p_user_id TEXT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar padrão de transação
    IF NOT EXISTS (
        SELECT 1 FROM jsonb_array_elements_text(p_pattern_data->'features')
        WHERE value::text IN ('amount', 'frequency', 'recipient', 'time')
    ) THEN
        RETURN FALSE;
    END IF;

    -- Verificar integridade dos dados
    IF NOT EXISTS (
        SELECT 1 FROM jsonb_object_keys(p_pattern_data)
        WHERE value IN ('transaction_id', 'user_id', 'features', 'risk_score')
    ) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;
