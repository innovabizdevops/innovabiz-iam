-- Script de Aumento de Segurança - IAM Open X
-- Versão: 2.0
-- Data: 15/05/2025
-- Novas funcionalidades:
-- - Proteção contra XSS
-- - Proteção contra CSRF
-- - Rate Limiting Avançado
-- - Monitoramento de Comportamento Anormal
-- - Detecção de Malware

-- 1. Funções de Segurança Avançada

-- 1.1 Proteção contra CSRF
CREATE OR REPLACE FUNCTION iam_access_control.prevent_csrf(
    p_token TEXT,
    p_session_id TEXT,
    p_user_agent TEXT
)
RETURNS JSON AS $$
DECLARE
    v_token_valid BOOLEAN := FALSE;
    v_session_valid BOOLEAN := FALSE;
    v_score INT := 0;
    v_recommendations JSONB;
    v_result JSON;
BEGIN
    -- Verificar validade do token
    SELECT 
        token_valid,
        session_valid
    INTO v_token_valid, v_session_valid
    FROM (
        SELECT 
            EXISTS(
                SELECT 1 
                FROM iam_access_control.csrf_tokens 
                WHERE token = p_token 
                AND expires_at > current_timestamp
            ) AS token_valid,
            EXISTS(
                SELECT 1 
                FROM iam_access_control.sessions 
                WHERE session_id = p_session_id 
                AND expires_at > current_timestamp
            ) AS session_valid
    ) t;

    -- Calcular score baseado nas verificações
    IF NOT v_token_valid THEN
        v_score := v_score + 30;
    END IF;
    
    IF NOT v_session_valid THEN
        v_score := v_score + 30;
    END IF;

    -- Verificar se o user agent mudou recentemente
    IF EXISTS (
        SELECT 1 
        FROM iam_access_control.session_history 
        WHERE session_id = p_session_id 
        AND user_agent != p_user_agent 
        AND event_time > current_timestamp - INTERVAL '15 minutos'
    ) THEN
        v_score := v_score + 20;
    END IF;

    -- Gerar recomendações baseadas no score
    v_recommendations := CASE 
        WHEN v_score >= 80 THEN 
            jsonb_build_object(
                'action', 'block',
                'reason', 'High risk of CSRF attack',
                'score', v_score
            )
        WHEN v_score >= 50 THEN 
            jsonb_build_object(
                'action', 'warn',
                'reason', 'Suspicious CSRF activity',
                'score', v_score
            )
        ELSE 
            jsonb_build_object(
                'action', 'allow',
                'reason', 'Valid CSRF token and session',
                'score', v_score
            )
    END;

    -- Registrar evento de segurança
    INSERT INTO iam_access_control.security_events (
        event_type,
        event_time,
        event_details,
        score_risco
    ) VALUES (
        CASE 
            WHEN v_score >= 80 THEN 'CSRF_ATTACK'
            ELSE 'CSRF_CHECK'
        END,
        current_timestamp,
        jsonb_build_object(
            'token_valid', v_token_valid,
            'session_valid', v_session_valid,
            'user_agent', p_user_agent,
            'recommendations', v_recommendations
        ),
        v_score
    );

    -- Preparar resultado
    v_result := json_build_object(
        'success', v_token_valid AND v_session_valid,
        'score_risco', v_score,
        'recommendations', v_recommendations,
        'details', json_build_object(
            'token_valid', v_token_valid,
            'session_valid', v_session_valid,
            'user_agent', p_user_agent
        )
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.2 Rate Limiting Avançado
CREATE OR REPLACE FUNCTION iam_access_control.advanced_rate_limit(
    p_ip_address TEXT,
    p_user_id TEXT,
    p_action_type TEXT,
    p_rate_limit_config JSONB DEFAULT NULL
)
RETURNS JSON AS $$
DECLARE
    v_request_count INT;
    v_user_count INT;
    v_score INT := 0;
    v_recommendations JSONB;
    v_result JSON;
    v_config JSONB;
    v_is_limited BOOLEAN := FALSE;
BEGIN
    -- Usar configuração padrão se não fornecida
    IF p_rate_limit_config IS NULL THEN
        p_rate_limit_config := jsonb_build_object(
            'window_minutes', 1,
            'max_requests', 100,
            'max_unique_users', 50,
            'score_threshold', 70
        );
    END IF;

    -- Contar requisições recentes
    SELECT COUNT(*) INTO v_request_count
    FROM iam_access_control.rate_limit_logs
    WHERE ip_address = p_ip_address
    AND request_time > current_timestamp - INTERVAL '1 minute' * (p_rate_limit_config->>'window_minutes')::INT;

    -- Contar usuários únicos recentes
    SELECT COUNT(DISTINCT user_id) INTO v_user_count
    FROM iam_access_control.rate_limit_logs
    WHERE ip_address = p_ip_address
    AND request_time > current_timestamp - INTERVAL '1 minute' * (p_rate_limit_config->>'window_minutes')::INT;

    -- Calcular score baseado nas verificações
    IF v_request_count > (p_rate_limit_config->>'max_requests')::INT THEN
        v_score := v_score + 40;
    END IF;

    IF v_user_count > (p_rate_limit_config->>'max_unique_users')::INT THEN
        v_score := v_score + 30;
    END IF;

    -- Verificar padrões de ataque conhecidos
    IF EXISTS (
        SELECT 1 
        FROM iam_access_control.known_attack_patterns 
        WHERE ip_address = p_ip_address 
        AND last_seen > current_timestamp - INTERVAL '1 day'
    ) THEN
        v_score := v_score + 20;
    END IF;

    -- Gerar recomendações baseadas no score
    v_recommendations := CASE 
        WHEN v_score >= (p_rate_limit_config->>'score_threshold')::INT THEN 
            jsonb_build_object(
                'action', 'block',
                'reason', 'High rate limiting score',
                'score', v_score,
                'details', jsonb_build_object(
                    'request_count', v_request_count,
                    'user_count', v_user_count,
                    'config', p_rate_limit_config
                )
            )
        WHEN v_score >= 50 THEN 
            jsonb_build_object(
                'action', 'warn',
                'reason', 'Suspicious rate limiting activity',
                'score', v_score,
                'details', jsonb_build_object(
                    'request_count', v_request_count,
                    'user_count', v_user_count,
                    'config', p_rate_limit_config
                )
            )
        ELSE 
            jsonb_build_object(
                'action', 'allow',
                'reason', 'Normal rate limiting activity',
                'score', v_score,
                'details', jsonb_build_object(
                    'request_count', v_request_count,
                    'user_count', v_user_count,
                    'config', p_rate_limit_config
                )
            )
    END;

    -- Decidir se deve limitar
    v_is_limited := v_score >= (p_rate_limit_config->>'score_threshold')::INT;

    -- Registrar o evento
    INSERT INTO iam_access_control.rate_limit_logs (
        ip_address,
        user_id,
        action_type,
        request_time,
        score,
        is_limited
    ) VALUES (
        p_ip_address,
        p_user_id,
        p_action_type,
        current_timestamp,
        v_score,
        v_is_limited
    );

    -- Preparar resultado
    v_result := json_build_object(
        'success', NOT v_is_limited,
        'score_risco', v_score,
        'recommendations', v_recommendations,
        'details', json_build_object(
            'request_count', v_request_count,
            'user_count', v_user_count,
            'config', p_rate_limit_config,
            'is_limited', v_is_limited
        )
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.3 Monitoramento de Comportamento Anormal
CREATE OR REPLACE FUNCTION iam_access_control.detect_abnormal_behavior(
    p_user_id TEXT,
    p_action_type TEXT,
    p_action_time TIMESTAMP,
    p_action_details JSONB
)
RETURNS JSON AS $$
DECLARE
    v_baseline_time INT;
    v_current_time INT;
    v_time_diff INT;
    v_score INT := 0;
    v_recommendations JSONB;
    v_result JSON;
    v_behavior_record RECORD;
BEGIN
    -- Obter linha de base do comportamento
    SELECT jsonb_agg(jsonb_build_object(
        'action_type', action_type,
        'average_time', AVG(EXTRACT(EPOCH FROM (action_time - LAG(action_time) OVER (PARTITION BY action_type ORDER BY action_time))))::INT,
        'std_dev', STDDEV(EXTRACT(EPOCH FROM (action_time - LAG(action_time) OVER (PARTITION BY action_type ORDER BY action_time))))::INT
    )) INTO v_behavior_record
    FROM iam_access_control.user_behavior_logs
    WHERE user_id = p_user_id
    AND action_time > current_timestamp - INTERVAL '1 month'
    GROUP BY action_type;

    -- Calcular tempo atual
    v_current_time := EXTRACT(EPOCH FROM p_action_time);
    
    -- Verificar padrões de tempo
    FOR behavior IN SELECT * FROM jsonb_array_elements(v_behavior_record) LOOP
        IF behavior->>'action_type' = p_action_type THEN
            v_baseline_time := (behavior->>'average_time')::INT;
            v_time_diff := ABS(v_current_time - v_baseline_time);
            
            -- Calcular score baseado no desvio
            IF v_time_diff > 2 * (behavior->>'std_dev')::INT THEN
                v_score := v_score + 30;
            END IF;
            
            IF v_time_diff > 3 * (behavior->>'std_dev')::INT THEN
                v_score := v_score + 20;
            END IF;
        END IF;
    END LOOP;

    -- Verificar padrões de ação
    IF EXISTS (
        SELECT 1 
        FROM iam_access_control.action_patterns 
        WHERE pattern_type = 'ANOMALOUS_SEQUENCE' 
        AND user_id = p_user_id 
        AND last_detected > current_timestamp - INTERVAL '1 day'
    ) THEN
        v_score := v_score + 20;
    END IF;

    -- Gerar recomendações baseadas no score
    v_recommendations := CASE 
        WHEN v_score >= 80 THEN 
            jsonb_build_object(
                'action', 'investigate',
                'reason', 'Highly abnormal behavior detected',
                'score', v_score,
                'details', jsonb_build_object(
                    'time_diff', v_time_diff,
                    'baseline_time', v_baseline_time,
                    'current_time', v_current_time
                )
            )
        WHEN v_score >= 50 THEN 
            jsonb_build_object(
                'action', 'monitor',
                'reason', 'Suspicious behavior pattern',
                'score', v_score,
                'details', jsonb_build_object(
                    'time_diff', v_time_diff,
                    'baseline_time', v_baseline_time,
                    'current_time', v_current_time
                )
            )
        ELSE 
            jsonb_build_object(
                'action', 'normal',
                'reason', 'Normal behavior pattern',
                'score', v_score,
                'details', jsonb_build_object(
                    'time_diff', v_time_diff,
                    'baseline_time', v_baseline_time,
                    'current_time', v_current_time
                )
            )
    END;

    -- Registrar comportamento
    INSERT INTO iam_access_control.user_behavior_logs (
        user_id,
        action_type,
        action_time,
        action_details,
        score_risco
    ) VALUES (
        p_user_id,
        p_action_type,
        p_action_time,
        p_action_details,
        v_score
    );

    -- Preparar resultado
    v_result := json_build_object(
        'user_id', p_user_id,
        'action_type', p_action_type,
        'score_risco', v_score,
        'recommendations', v_recommendations,
        'details', json_build_object(
            'time_diff', v_time_diff,
            'baseline_time', v_baseline_time,
            'current_time', v_current_time,
            'behavior_patterns', v_behavior_record
        )
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.4 Detecção de Malware
CREATE OR REPLACE FUNCTION iam_access_control.detect_malware(
    p_file_path TEXT,
    p_file_type TEXT,
    p_file_size BIGINT,
    p_file_hash TEXT
)
RETURNS JSON AS $$
DECLARE
    v_is_malware BOOLEAN := FALSE;
    v_score INT := 0;
    v_recommendations JSONB;
    v_result JSON;
BEGIN
    -- Verificar assinaturas conhecidas
    IF EXISTS (
        SELECT 1 
        FROM iam_access_control.known_malware_signatures 
        WHERE signature_type = 'HASH' 
        AND signature_value = p_file_hash
    ) THEN
        v_score := v_score + 50;
        v_is_malware := TRUE;
    END IF;

    -- Verificar extensões suspeitas
    IF p_file_type IN ('exe', 'dll', 'js', 'vbs', 'bat', 'cmd', 'scr') THEN
        v_score := v_score + 20;
    END IF;

    -- Verificar tamanho suspeito
    IF p_file_size > 10000000 THEN -- 10MB
        v_score := v_score + 15;
    END IF;

    -- Verificar padrões de arquivo
    IF EXISTS (
        SELECT 1 
        FROM iam_access_control.file_patterns 
        WHERE pattern_type = 'SUSPICIOUS' 
        AND file_type = p_file_type
    ) THEN
        v_score := v_score + 25;
    END IF;

    -- Gerar recomendações baseadas no score
    v_recommendations := CASE 
        WHEN v_score >= 80 THEN 
            jsonb_build_object(
                'action', 'quarantine',
                'reason', 'High risk malware detected',
                'score', v_score,
                'details', jsonb_build_object(
                    'file_path', p_file_path,
                    'file_type', p_file_type,
                    'file_size', p_file_size,
                    'file_hash', p_file_hash
                )
            )
        WHEN v_score >= 50 THEN 
            jsonb_build_object(
                'action', 'scan',
                'reason', 'Suspicious file characteristics',
                'score', v_score,
                'details', jsonb_build_object(
                    'file_path', p_file_path,
                    'file_type', p_file_type,
                    'file_size', p_file_size,
                    'file_hash', p_file_hash
                )
            )
        ELSE 
            jsonb_build_object(
                'action', 'allow',
                'reason', 'File appears safe',
                'score', v_score,
                'details', jsonb_build_object(
                    'file_path', p_file_path,
                    'file_type', p_file_type,
                    'file_size', p_file_size,
                    'file_hash', p_file_hash
                )
            )
    END;

    -- Registrar evento de segurança
    INSERT INTO iam_access_control.security_events (
        event_type,
        event_time,
        event_details,
        score_risco
    ) VALUES (
        CASE 
            WHEN v_score >= 80 THEN 'MALWARE_DETECTED'
            ELSE 'FILE_CHECKED'
        END,
        current_timestamp,
        jsonb_build_object(
            'file_path', p_file_path,
            'file_type', p_file_type,
            'file_size', p_file_size,
            'file_hash', p_file_hash,
            'recommendations', v_recommendations
        ),
        v_score
    );

    -- Preparar resultado
    v_result := json_build_object(
        'file_path', p_file_path,
        'file_type', p_file_type,
        'score_risco', v_score,
        'is_malware', v_is_malware,
        'recommendations', v_recommendations,
        'details', json_build_object(
            'file_size', p_file_size,
            'file_hash', p_file_hash
        )
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.5 Multi-Factor Authentication (MFA)
CREATE OR REPLACE FUNCTION iam_access_control.enable_mfa(
    p_user_id TEXT,
    p_method_type TEXT, -- 'TOTP', 'SMS', 'EMAIL', 'BIOMETRIC'
    p_method_data JSONB
)
RETURNS JSON AS $$
DECLARE
    v_mfa_id UUID;
    v_secret TEXT;
    v_qr_code TEXT;
    v_result JSON;
BEGIN
    -- Gerar ID único para MFA
    v_mfa_id := gen_random_uuid();

    -- Gerar segredo baseado no método
    CASE 
        WHEN p_method_type = 'TOTP' THEN
            v_secret := encode(gen_random_bytes(20), 'base32');
            v_qr_code := 'otpauth://totp/' || p_method_data->>'issuer' || ':' || p_method_data->>'account' ||
                         '?secret=' || v_secret ||
                         '&issuer=' || p_method_data->>'issuer';
        WHEN p_method_type = 'BIOMETRIC' THEN
            -- Configuração específica para biometria
            v_secret := encode(gen_random_bytes(32), 'base64');
    END CASE;

    -- Registrar método de MFA
    INSERT INTO iam_access_control.mfa_methods (
        mfa_id,
        user_id,
        method_type,
        method_data,
        secret,
        qr_code,
        created_at,
        last_verified_at
    ) VALUES (
        v_mfa_id,
        p_user_id,
        p_method_type,
        p_method_data,
        v_secret,
        v_qr_code,
        current_timestamp,
        NULL
    );

    -- Preparar resultado
    v_result := json_build_object(
        'success', TRUE,
        'mfa_id', v_mfa_id,
        'method_type', p_method_type,
        'qr_code', v_qr_code,
        'secret', v_secret,
        'method_data', p_method_data
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.6 Verificação de MFA
CREATE OR REPLACE FUNCTION iam_access_control.verify_mfa(
    p_user_id TEXT,
    p_mfa_id UUID,
    p_code TEXT,
    p_biometric_data JSONB DEFAULT NULL
)
RETURNS JSON AS $$
DECLARE
    v_method_record RECORD;
    v_is_valid BOOLEAN := FALSE;
    v_score INT := 0;
    v_recommendations JSONB;
    v_result JSON;
BEGIN
    -- Obter detalhes do método MFA
    SELECT * INTO v_method_record
    FROM iam_access_control.mfa_methods
    WHERE mfa_id = p_mfa_id
    AND user_id = p_user_id;

    -- Verificar código baseado no método
    CASE 
        WHEN v_method_record.method_type = 'TOTP' THEN
            -- Verificar código TOTP
            IF totp.verify_code(v_method_record.secret, p_code) THEN
                v_is_valid := TRUE;
                v_score := v_score + 50;
            END IF;
        WHEN v_method_record.method_type = 'BIOMETRIC' THEN
            -- Verificar dados biométricos
            IF biometric.verify_data(v_method_record.secret, p_biometric_data) THEN
                v_is_valid := TRUE;
                v_score := v_score + 70;
            END IF;
    END CASE;

    -- Atualizar último verificado
    IF v_is_valid THEN
        UPDATE iam_access_control.mfa_methods
        SET last_verified_at = current_timestamp
        WHERE mfa_id = p_mfa_id;
    END IF;

    -- Gerar recomendações
    v_recommendations := jsonb_build_object(
        'action', CASE 
            WHEN v_is_valid THEN 'allow'
            ELSE 'deny'
        END,
        'reason', CASE 
            WHEN v_is_valid THEN 'MFA verification successful'
            ELSE 'MFA verification failed'
        END,
        'score', v_score
    );

    -- Preparar resultado
    v_result := json_build_object(
        'success', v_is_valid,
        'mfa_id', p_mfa_id,
        'method_type', v_method_record.method_type,
        'score_risco', v_score,
        'recommendations', v_recommendations,
        'details', json_build_object(
            'method_data', v_method_record.method_data,
            'last_verified_at', v_method_record.last_verified_at
        )
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.7 Single Sign-On (SSO)
CREATE OR REPLACE FUNCTION iam_access_control.initiate_sso(
    p_user_id TEXT,
    p_service_provider TEXT,
    p_request_data JSONB
)
RETURNS JSON AS $$
DECLARE
    v_sso_token UUID;
    v_session_id UUID;
    v_expires_at TIMESTAMP;
    v_result JSON;
BEGIN
    -- Gerar token SSO único
    v_sso_token := gen_random_uuid();
    
    -- Obter ID da sessão atual
    SELECT session_id INTO v_session_id
    FROM iam_access_control.sessions
    WHERE user_id = p_user_id
    AND expires_at > current_timestamp;

    -- Definir tempo de expiração (15 minutos)
    v_expires_at := current_timestamp + INTERVAL '15 minutes';

    -- Registrar SSO request
    INSERT INTO iam_access_control.sso_requests (
        sso_token,
        user_id,
        service_provider,
        session_id,
        request_data,
        created_at,
        expires_at
    ) VALUES (
        v_sso_token,
        p_user_id,
        p_service_provider,
        v_session_id,
        p_request_data,
        current_timestamp,
        v_expires_at
    );

    -- Preparar resultado
    v_result := json_build_object(
        'success', TRUE,
        'sso_token', v_sso_token,
        'expires_at', v_expires_at,
        'service_provider', p_service_provider,
        'request_data', p_request_data
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.8 Processar SSO Response
CREATE OR REPLACE FUNCTION iam_access_control.process_sso_response(
    p_sso_token UUID,
    p_response_data JSONB,
    p_service_provider TEXT
)
RETURNS JSON AS $$
DECLARE
    v_request_record RECORD;
    v_is_valid BOOLEAN := FALSE;
    v_score INT := 0;
    v_recommendations JSONB;
    v_result JSON;
BEGIN
    -- Obter detalhes do SSO request
    SELECT * INTO v_request_record
    FROM iam_access_control.sso_requests
    WHERE sso_token = p_sso_token
    AND service_provider = p_service_provider
    AND expires_at > current_timestamp;

    IF v_request_record IS NOT NULL THEN
        -- Verificar integridade da resposta
        IF sso.verify_response(
            v_request_record.user_id,
            p_response_data,
            v_request_record.request_data
        ) THEN
            v_is_valid := TRUE;
            v_score := v_score + 50;
        END IF;

        -- Atualizar status do SSO request
        UPDATE iam_access_control.sso_requests
        SET 
            response_data = p_response_data,
            processed_at = current_timestamp,
            is_valid = v_is_valid
        WHERE sso_token = p_sso_token;
    END IF;

    -- Gerar recomendações
    v_recommendations := jsonb_build_object(
        'action', CASE 
            WHEN v_is_valid THEN 'allow'
            ELSE 'deny'
        END,
        'reason', CASE 
            WHEN v_is_valid THEN 'SSO response valid'
            ELSE 'SSO response invalid'
        END,
        'score', v_score
    );

    -- Preparar resultado
    v_result := json_build_object(
        'success', v_is_valid,
        'sso_token', p_sso_token,
        'service_provider', p_service_provider,
        'score_risco', v_score,
        'recommendations', v_recommendations,
        'details', json_build_object(
            'request_data', v_request_record.request_data,
            'response_data', p_response_data,
            'processed_at', current_timestamp
        )
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.9 OAuth2/OpenID Connect
CREATE OR REPLACE FUNCTION iam_access_control.oauth_authorize(
    p_client_id TEXT,
    p_redirect_uri TEXT,
    p_scope TEXT,
    p_state TEXT,
    p_response_type TEXT
)
RETURNS JSON AS $$
DECLARE
    v_auth_code UUID;
    v_client_record RECORD;
    v_scope_valid BOOLEAN := FALSE;
    v_result JSON;
BEGIN
    -- Verificar cliente
    SELECT * INTO v_client_record
    FROM iam_access_control.oauth_clients
    WHERE client_id = p_client_id
    AND redirect_uri = p_redirect_uri;

    IF v_client_record IS NOT NULL THEN
        -- Verificar escopos
        IF v_client_record.allowed_scopes @> string_to_array(p_scope, ' ')::TEXT[] THEN
            v_scope_valid := TRUE;
        END IF;

        -- Gerar código de autorização
        IF v_scope_valid THEN
            v_auth_code := gen_random_uuid();

            -- Registrar autorização
            INSERT INTO iam_access_control.oauth_authorizations (
                auth_code,
                client_id,
                redirect_uri,
                scope,
                state,
                created_at,
                expires_at
            ) VALUES (
                v_auth_code,
                p_client_id,
                p_redirect_uri,
                p_scope,
                p_state,
                current_timestamp,
                current_timestamp + INTERVAL '10 minutes'
            );
        END IF;
    END IF;

    -- Preparar resultado
    v_result := json_build_object(
        'success', v_scope_valid,
        'auth_code', CASE 
            WHEN v_scope_valid THEN v_auth_code
            ELSE NULL
        END,
        'redirect_uri', p_redirect_uri,
        'scope', p_scope,
        'state', p_state
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.10 OAuth Token
CREATE OR REPLACE FUNCTION iam_access_control.oauth_token(
    p_auth_code UUID,
    p_client_id TEXT,
    p_client_secret TEXT,
    p_grant_type TEXT
)
RETURNS JSON AS $$
DECLARE
    v_auth_record RECORD;
    v_access_token UUID;
    v_refresh_token UUID;
    v_expires_in INT;
    v_result JSON;
BEGIN
    -- Verificar autorização
    SELECT * INTO v_auth_record
    FROM iam_access_control.oauth_authorizations
    WHERE auth_code = p_auth_code
    AND expires_at > current_timestamp;

    IF v_auth_record IS NOT NULL THEN
        -- Gerar tokens
        v_access_token := gen_random_uuid();
        v_refresh_token := gen_random_uuid();
        v_expires_in := 3600; -- 1 hora

        -- Registrar token de acesso
        INSERT INTO iam_access_control.oauth_tokens (
            access_token,
            refresh_token,
            client_id,
            scope,
            created_at,
            expires_at
        ) VALUES (
            v_access_token,
            v_refresh_token,
            p_client_id,
            v_auth_record.scope,
            current_timestamp,
            current_timestamp + INTERVAL '1 hour'
        );

        -- Remover autorização usada
        DELETE FROM iam_access_control.oauth_authorizations
        WHERE auth_code = p_auth_code;
    END IF;

    -- Preparar resultado
    v_result := json_build_object(
        'success', v_auth_record IS NOT NULL,
        'access_token', v_access_token,
        'refresh_token', v_refresh_token,
        'token_type', 'Bearer',
        'expires_in', v_expires_in,
        'scope', CASE 
            WHEN v_auth_record IS NOT NULL THEN v_auth_record.scope
            ELSE NULL
        END
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.11 Autenticação Biométrica
CREATE OR REPLACE FUNCTION iam_access_control.authenticate_biometric(
    p_user_id TEXT,
    p_bio_type TEXT, -- 'fingerprint', 'face', 'iris', 'voice', 'vein', 'palm', 'gait'
    p_bio_data JSONB,
    p_device_info JSONB
)
RETURNS JSON AS $$
DECLARE
    v_template_record RECORD;
    v_score INT := 0;
    v_match_threshold INT := 70; -- Limite mínimo de correspondência
    v_recommendations JSONB;
    v_result JSON;
BEGIN
    -- Obter template biométrico do usuário
    SELECT * INTO v_template_record
    FROM iam_access_control.biometric_templates
    WHERE user_id = p_user_id
    AND bio_type = p_bio_type;

    IF v_template_record IS NOT NULL THEN
        -- Verificar correspondência biométrica
        CASE 
            WHEN p_bio_type = 'fingerprint' THEN
                IF fingerprint.match_template(
                    v_template_record.template_data,
                    p_bio_data->>'fingerprint_data'
                ) >= v_match_threshold THEN
                    v_score := v_score + 90;
                END IF;
            WHEN p_bio_type = 'face' THEN
                IF face.match_template(
                    v_template_record.template_data,
                    p_bio_data->>'face_data'
                ) >= v_match_threshold THEN
                    v_score := v_score + 90;
                END IF;
            WHEN p_bio_type = 'iris' THEN
                IF iris.match_template(
                    v_template_record.template_data,
                    p_bio_data->>'iris_data'
                ) >= v_match_threshold THEN
                    v_score := v_score + 95;
                END IF;
            WHEN p_bio_type = 'voice' THEN
                IF voice.match_template(
                    v_template_record.template_data,
                    p_bio_data->>'voice_data'
                ) >= v_match_threshold THEN
                    v_score := v_score + 85;
                END IF;
            WHEN p_bio_type = 'vein' THEN
                IF vein.match_template(
                    v_template_record.template_data,
                    p_bio_data->>'vein_data'
                ) >= v_match_threshold THEN
                    v_score := v_score + 95;
                END IF;
            WHEN p_bio_type = 'palm' THEN
                IF palm.match_template(
                    v_template_record.template_data,
                    p_bio_data->>'palm_data'
                ) >= v_match_threshold THEN
                    v_score := v_score + 90;
                END IF;
            WHEN p_bio_type = 'gait' THEN
                IF gait.match_template(
                    v_template_record.template_data,
                    p_bio_data->>'gait_data'
                ) >= v_match_threshold THEN
                    v_score := v_score + 80;
                END IF;
        END CASE;

        -- Ajustar score baseado no dispositivo
        IF p_device_info->>'is_secure' = 'true' THEN
            v_score := v_score + 10;
        END IF;

        -- Gerar recomendações
        v_recommendations := jsonb_build_object(
            'action', CASE 
                WHEN v_score >= 80 THEN 'allow'
                WHEN v_score >= 60 THEN 'review'
                ELSE 'deny'
            END,
            'reason', CASE 
                WHEN v_score >= 80 THEN 'High confidence match'
                WHEN v_score >= 60 THEN 'Medium confidence match'
                ELSE 'Low confidence match'
            END,
            'score', v_score,
            'details', jsonb_build_object(
                'bio_type', p_bio_type,
                'device_info', p_device_info,
                'match_threshold', v_match_threshold
            )
        );

        -- Registrar evento de segurança
        INSERT INTO iam_access_control.security_events (
            event_type,
            event_time,
            event_details,
            score_risco
        ) VALUES (
            CASE 
                WHEN v_score >= 80 THEN 'BIOMETRIC_AUTH_SUCCESS'
                ELSE 'BIOMETRIC_AUTH_ATTEMPT'
            END,
            current_timestamp,
            jsonb_build_object(
                'user_id', p_user_id,
                'bio_type', p_bio_type,
                'device_info', p_device_info,
                'recommendations', v_recommendations
            ),
            v_score
        );
    END IF;

    -- Preparar resultado
    v_result := json_build_object(
        'success', v_score >= 80,
        'user_id', p_user_id,
        'bio_type', p_bio_type,
        'score_risco', v_score,
        'recommendations', v_recommendations,
        'details', json_build_object(
            'device_info', p_device_info,
            'match_threshold', v_match_threshold
        )
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.12 Registrar Template Biométrico
CREATE OR REPLACE FUNCTION iam_access_control.register_biometric_template(
    p_user_id TEXT,
    p_bio_type TEXT,
    p_template_data JSONB,
    p_device_info JSONB
)
RETURNS JSON AS $$
DECLARE
    v_template_id UUID;
    v_result JSON;
BEGIN
    -- Gerar ID único para template
    v_template_id := gen_random_uuid();

    -- Registrar template biométrico
    INSERT INTO iam_access_control.biometric_templates (
        template_id,
        user_id,
        bio_type,
        template_data,
        device_info,
        created_at,
        last_updated_at
    ) VALUES (
        v_template_id,
        p_user_id,
        p_bio_type,
        p_template_data,
        p_device_info,
        current_timestamp,
        current_timestamp
    );

    -- Preparar resultado
    v_result := json_build_object(
        'success', TRUE,
        'template_id', v_template_id,
        'user_id', p_user_id,
        'bio_type', p_bio_type,
        'device_info', p_device_info,
        'created_at', current_timestamp
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.13 Autenticação Contextual
CREATE OR REPLACE FUNCTION iam_access_control.contextual_auth(
    p_user_id TEXT,
    p_context_data JSONB, -- {location, device, time, behavior, risk, environment}
    p_required_level INT -- Nível de autenticação necessário
)
RETURNS JSON AS $$
DECLARE
    v_current_level INT := 0;
    v_score INT := 0;
    v_recommendations JSONB;
    v_result JSON;
BEGIN
    -- Verificar localização
    IF p_context_data->>'location_valid' = 'true' THEN
        v_score := v_score + 20;
        v_current_level := v_current_level + 1;
    END IF;

    -- Verificar dispositivo
    IF p_context_data->>'device_valid' = 'true' THEN
        v_score := v_score + 20;
        v_current_level := v_current_level + 1;
    END IF;

    -- Verificar horário
    IF p_context_data->>'time_valid' = 'true' THEN
        v_score := v_score + 15;
        v_current_level := v_current_level + 1;
    END IF;

    -- Verificar comportamento
    IF p_context_data->>'behavior_valid' = 'true' THEN
        v_score := v_score + 25;
        v_current_level := v_current_level + 1;
    END IF;

    -- Verificar risco
    IF p_context_data->>'risk_level'::INT <= 2 THEN
        v_score := v_score + 20;
    END IF;

    -- Verificar ambiente
    IF p_context_data->>'environment_secure' = 'true' THEN
        v_score := v_score + 15;
    END IF;

    -- Gerar recomendações
    v_recommendations := jsonb_build_object(
        'action', CASE 
            WHEN v_current_level >= p_required_level THEN 'allow'
            WHEN v_score >= 80 THEN 'review'
            ELSE 'deny'
        END,
        'reason', CASE 
            WHEN v_current_level >= p_required_level THEN 'Required authentication level met'
            WHEN v_score >= 80 THEN 'High confidence score'
            ELSE 'Insufficient authentication level'
        END,
        'score', v_score,
        'current_level', v_current_level,
        'required_level', p_required_level,
        'details', jsonb_build_object(
            'location', p_context_data->>'location_valid',
            'device', p_context_data->>'device_valid',
            'time', p_context_data->>'time_valid',
            'behavior', p_context_data->>'behavior_valid',
            'risk', p_context_data->>'risk_level',
            'environment', p_context_data->>'environment_secure'
        )
    );

    -- Registrar evento de segurança
    INSERT INTO iam_access_control.security_events (
        event_type,
        event_time,
        event_details,
        score_risco
    ) VALUES (
        CASE 
            WHEN v_current_level >= p_required_level THEN 'CONTEXT_AUTH_SUCCESS'
            ELSE 'CONTEXT_AUTH_ATTEMPT'
        END,
        current_timestamp,
        jsonb_build_object(
            'user_id', p_user_id,
            'context_data', p_context_data,
            'required_level', p_required_level,
            'recommendations', v_recommendations
        ),
        v_score
    );

    -- Preparar resultado
    v_result := json_build_object(
        'success', v_current_level >= p_required_level,
        'user_id', p_user_id,
        'score_risco', v_score,
        'current_level', v_current_level,
        'required_level', p_required_level,
        'recommendations', v_recommendations,
        'details', json_build_object(
            'context_data', p_context_data,
            'current_level', v_current_level
        )
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.14 Configurar Nível de Autenticação
CREATE OR REPLACE FUNCTION iam_access_control.configure_auth_level(
    p_user_id TEXT,
    p_service_id TEXT,
    p_required_level INT,
    p_context_rules JSONB
)
RETURNS JSON AS $$
DECLARE
    v_config_id UUID;
    v_result JSON;
BEGIN
    -- Gerar ID único para configuração
    v_config_id := gen_random_uuid();

    -- Registrar configuração de nível de autenticação
    INSERT INTO iam_access_control.auth_level_configs (
        config_id,
        user_id,
        service_id,
        required_level,
        context_rules,
        created_at,
        last_updated_at
    ) VALUES (
        v_config_id,
        p_user_id,
        p_service_id,
        p_required_level,
        p_context_rules,
        current_timestamp,
        current_timestamp
    );

    -- Preparar resultado
    v_result := json_build_object(
        'success', TRUE,
        'config_id', v_config_id,
        'user_id', p_user_id,
        'service_id', p_service_id,
        'required_level', p_required_level,
        'context_rules', p_context_rules,
        'created_at', current_timestamp
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.15 Tokenização Avançada
CREATE OR REPLACE FUNCTION iam_access_control.advanced_tokenization(
    p_user_id TEXT,
    p_token_type TEXT, -- 'TOTP', 'HOTP', 'OCRA', 'FIDO', 'YUBIKEY'
    p_token_data JSONB,
    p_device_info JSONB
)
RETURNS JSON AS $$
DECLARE
    v_token_id UUID;
    v_secret TEXT;
    v_qr_code TEXT;
    v_score INT := 0;
    v_recommendations JSONB;
    v_result JSON;
BEGIN
    -- Gerar ID único para token
    v_token_id := gen_random_uuid();

    -- Gerar segredo baseado no tipo de token
    CASE 
        WHEN p_token_type = 'TOTP' THEN
            v_secret := encode(gen_random_bytes(20), 'base32');
            v_qr_code := 'otpauth://totp/' || p_token_data->>'issuer' || ':' || p_token_data->>'account' ||
                         '?secret=' || v_secret ||
                         '&issuer=' || p_token_data->>'issuer';
            v_score := v_score + 70;
        WHEN p_token_type = 'HOTP' THEN
            v_secret := encode(gen_random_bytes(20), 'base32');
            v_score := v_score + 60;
        WHEN p_token_type = 'OCRA' THEN
            v_secret := encode(gen_random_bytes(32), 'base64');
            v_score := v_score + 80;
        WHEN p_token_type = 'FIDO' THEN
            v_secret := encode(gen_random_bytes(64), 'base64');
            v_score := v_score + 90;
        WHEN p_token_type = 'YUBIKEY' THEN
            v_secret := encode(gen_random_bytes(16), 'hex');
            v_score := v_score + 95;
    END CASE;

    -- Ajustar score baseado no dispositivo
    IF p_device_info->>'is_secure' = 'true' THEN
        v_score := v_score + 10;
    END IF;

    -- Registrar token
    INSERT INTO iam_access_control.tokens (
        token_id,
        user_id,
        token_type,
        token_data,
        secret,
        qr_code,
        device_info,
        created_at,
        last_verified_at
    ) VALUES (
        v_token_id,
        p_user_id,
        p_token_type,
        p_token_data,
        v_secret,
        v_qr_code,
        p_device_info,
        current_timestamp,
        NULL
    );

    -- Gerar recomendações
    v_recommendations := jsonb_build_object(
        'action', CASE 
            WHEN v_score >= 80 THEN 'allow'
            WHEN v_score >= 60 THEN 'review'
            ELSE 'deny'
        END,
        'reason', CASE 
            WHEN v_score >= 80 THEN 'High security token'
            WHEN v_score >= 60 THEN 'Medium security token'
            ELSE 'Low security token'
        END,
        'score', v_score,
        'details', jsonb_build_object(
            'token_type', p_token_type,
            'device_info', p_device_info
        )
    );

    -- Registrar evento de segurança
    INSERT INTO iam_access_control.security_events (
        event_type,
        event_time,
        event_details,
        score_risco
    ) VALUES (
        'TOKEN_GENERATION',
        current_timestamp,
        jsonb_build_object(
            'user_id', p_user_id,
            'token_type', p_token_type,
            'device_info', p_device_info,
            'recommendations', v_recommendations
        ),
        v_score
    );

    -- Preparar resultado
    v_result := json_build_object(
        'success', TRUE,
        'token_id', v_token_id,
        'token_type', p_token_type,
        'score_risco', v_score,
        'recommendations', v_recommendations,
        'details', json_build_object(
            'token_data', p_token_data,
            'device_info', p_device_info,
            'qr_code', v_qr_code
        )
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.16 Verificação de Token
CREATE OR REPLACE FUNCTION iam_access_control.verify_token(
    p_token_id UUID,
    p_code TEXT,
    p_device_info JSONB
)
RETURNS JSON AS $$
DECLARE
    v_token_record RECORD;
    v_is_valid BOOLEAN := FALSE;
    v_score INT := 0;
    v_recommendations JSONB;
    v_result JSON;
BEGIN
    -- Obter detalhes do token
    SELECT * INTO v_token_record
    FROM iam_access_control.tokens
    WHERE token_id = p_token_id;

    -- Verificar código baseado no tipo de token
    CASE 
        WHEN v_token_record.token_type = 'TOTP' THEN
            IF totp.verify_code(v_token_record.secret, p_code) THEN
                v_is_valid := TRUE;
                v_score := v_score + 70;
            END IF;
        WHEN v_token_record.token_type = 'HOTP' THEN
            IF hotp.verify_code(v_token_record.secret, p_code) THEN
                v_is_valid := TRUE;
                v_score := v_score + 60;
            END IF;
        WHEN v_token_record.token_type = 'OCRA' THEN
            IF ocra.verify_code(v_token_record.secret, p_code, v_token_record.token_data) THEN
                v_is_valid := TRUE;
                v_score := v_score + 80;
            END IF;
        WHEN v_token_record.token_type = 'FIDO' THEN
            IF fido.verify_code(v_token_record.secret, p_code) THEN
                v_is_valid := TRUE;
                v_score := v_score + 90;
            END IF;
        WHEN v_token_record.token_type = 'YUBIKEY' THEN
            IF yubikey.verify_code(v_token_record.secret, p_code) THEN
                v_is_valid := TRUE;
                v_score := v_score + 95;
            END IF;
    END CASE;

    -- Ajustar score baseado no dispositivo
    IF p_device_info->>'is_secure' = 'true' THEN
        v_score := v_score + 10;
    END IF;

    -- Atualizar último verificado
    IF v_is_valid THEN
        UPDATE iam_access_control.tokens
        SET last_verified_at = current_timestamp
        WHERE token_id = p_token_id;
    END IF;

    -- Gerar recomendações
    v_recommendations := jsonb_build_object(
        'action', CASE 
            WHEN v_is_valid THEN 'allow'
            ELSE 'deny'
        END,
        'reason', CASE 
            WHEN v_is_valid THEN 'Token verification successful'
            ELSE 'Token verification failed'
        END,
        'score', v_score,
        'details', jsonb_build_object(
            'token_type', v_token_record.token_type,
            'device_info', p_device_info
        )
    );

    -- Registrar evento de segurança
    INSERT INTO iam_access_control.security_events (
        event_type,
        event_time,
        event_details,
        score_risco
    ) VALUES (
        CASE 
            WHEN v_is_valid THEN 'TOKEN_VERIFIED'
            ELSE 'TOKEN_VERIFICATION_FAILED'
        END,
        current_timestamp,
        jsonb_build_object(
            'token_id', p_token_id,
            'token_type', v_token_record.token_type,
            'device_info', p_device_info,
            'recommendations', v_recommendations
        ),
        v_score
    );

    -- Preparar resultado
    v_result := json_build_object(
        'success', v_is_valid,
        'token_id', p_token_id,
        'token_type', v_token_record.token_type,
        'score_risco', v_score,
        'recommendations', v_recommendations,
        'details', json_build_object(
            'token_data', v_token_record.token_data,
            'device_info', p_device_info
        )
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.17 Certificação Digital
CREATE OR REPLACE FUNCTION iam_access_control.generate_certificate(
    p_user_id TEXT,
    p_cert_type TEXT, -- 'X509', 'PKI', 'S/MIME', 'OPENPGP'
    p_cert_data JSONB,
    p_validity_days INT
)
RETURNS JSON AS $$
DECLARE
    v_cert_id UUID;
    v_private_key TEXT;
    v_public_key TEXT;
    v_certificate TEXT;
    v_score INT := 0;
    v_recommendations JSONB;
    v_result JSON;
BEGIN
    -- Gerar ID único para certificado
    v_cert_id := gen_random_uuid();

    -- Gerar par de chaves baseado no tipo de certificado
    CASE 
        WHEN p_cert_type = 'X509' THEN
            SELECT * INTO v_private_key, v_public_key
            FROM generate_x509_keypair();
            v_score := v_score + 90;
        WHEN p_cert_type = 'PKI' THEN
            SELECT * INTO v_private_key, v_public_key
            FROM generate_pki_keypair();
            v_score := v_score + 85;
        WHEN p_cert_type = 'S/MIME' THEN
            SELECT * INTO v_private_key, v_public_key
            FROM generate_smime_keypair();
            v_score := v_score + 80;
        WHEN p_cert_type = 'OPENPGP' THEN
            SELECT * INTO v_private_key, v_public_key
            FROM generate_openpgp_keypair();
            v_score := v_score + 85;
    END CASE;

    -- Gerar certificado
    v_certificate := generate_certificate(
        v_public_key,
        p_cert_data,
        p_validity_days
    );

    -- Registrar certificado
    INSERT INTO iam_access_control.certificates (
        cert_id,
        user_id,
        cert_type,
        cert_data,
        private_key,
        public_key,
        certificate,
        created_at,
        expires_at
    ) VALUES (
        v_cert_id,
        p_user_id,
        p_cert_type,
        p_cert_data,
        v_private_key,
        v_public_key,
        v_certificate,
        current_timestamp,
        current_timestamp + INTERVAL '1 day' * p_validity_days
    );

    -- Gerar recomendações
    v_recommendations := jsonb_build_object(
        'action', CASE 
            WHEN v_score >= 85 THEN 'allow'
            ELSE 'review'
        END,
        'reason', CASE 
            WHEN v_score >= 85 THEN 'High security certificate'
            ELSE 'Medium security certificate'
        END,
        'score', v_score,
        'details', jsonb_build_object(
            'cert_type', p_cert_type,
            'validity_days', p_validity_days
        )
    );

    -- Registrar evento de segurança
    INSERT INTO iam_access_control.security_events (
        event_type,
        event_time,
        event_details,
        score_risco
    ) VALUES (
        'CERTIFICATE_GENERATION',
        current_timestamp,
        jsonb_build_object(
            'user_id', p_user_id,
            'cert_type', p_cert_type,
            'validity_days', p_validity_days,
            'recommendations', v_recommendations
        ),
        v_score
    );

    -- Preparar resultado
    v_result := json_build_object(
        'success', TRUE,
        'cert_id', v_cert_id,
        'cert_type', p_cert_type,
        'score_risco', v_score,
        'recommendations', v_recommendations,
        'details', json_build_object(
            'cert_data', p_cert_data,
            'validity_days', p_validity_days,
            'certificate', v_certificate
        )
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.18 Verificar Certificado
CREATE OR REPLACE FUNCTION iam_access_control.verify_certificate(
    p_cert_id UUID,
    p_data TEXT,
    p_signature TEXT
)
RETURNS JSON AS $$
DECLARE
    v_cert_record RECORD;
    v_is_valid BOOLEAN := FALSE;
    v_score INT := 0;
    v_recommendations JSONB;
    v_result JSON;
BEGIN
    -- Obter detalhes do certificado
    SELECT * INTO v_cert_record
    FROM iam_access_control.certificates
    WHERE cert_id = p_cert_id
    AND expires_at > current_timestamp;

    IF v_cert_record IS NOT NULL THEN
        -- Verificar assinatura baseado no tipo de certificado
        CASE 
            WHEN v_cert_record.cert_type = 'X509' THEN
                IF verify_x509_signature(
                    v_cert_record.public_key,
                    p_data,
                    p_signature
                ) THEN
                    v_is_valid := TRUE;
                    v_score := v_score + 90;
                END IF;
            WHEN v_cert_record.cert_type = 'PKI' THEN
                IF verify_pki_signature(
                    v_cert_record.public_key,
                    p_data,
                    p_signature
                ) THEN
                    v_is_valid := TRUE;
                    v_score := v_score + 85;
                END IF;
            WHEN v_cert_record.cert_type = 'S/MIME' THEN
                IF verify_smime_signature(
                    v_cert_record.public_key,
                    p_data,
                    p_signature
                ) THEN
                    v_is_valid := TRUE;
                    v_score := v_score + 80;
                END IF;
            WHEN v_cert_record.cert_type = 'OPENPGP' THEN
                IF verify_openpgp_signature(
                    v_cert_record.public_key,
                    p_data,
                    p_signature
                ) THEN
                    v_is_valid := TRUE;
                    v_score := v_score + 85;
                END IF;
        END CASE;

        -- Atualizar último verificado
        IF v_is_valid THEN
            UPDATE iam_access_control.certificates
            SET last_verified_at = current_timestamp
            WHERE cert_id = p_cert_id;
        END IF;
    END IF;

    -- Gerar recomendações
    v_recommendations := jsonb_build_object(
        'action', CASE 
            WHEN v_is_valid THEN 'allow'
            ELSE 'deny'
        END,
        'reason', CASE 
            WHEN v_is_valid THEN 'Certificate verification successful'
            ELSE 'Certificate verification failed'
        END,
        'score', v_score,
        'details', jsonb_build_object(
            'cert_type', v_cert_record.cert_type,
            'expires_at', v_cert_record.expires_at
        )
    );

    -- Registrar evento de segurança
    INSERT INTO iam_access_control.security_events (
        event_type,
        event_time,
        event_details,
        score_risco
    ) VALUES (
        CASE 
            WHEN v_is_valid THEN 'CERTIFICATE_VERIFIED'
            ELSE 'CERTIFICATE_VERIFICATION_FAILED'
        END,
        current_timestamp,
        jsonb_build_object(
            'cert_id', p_cert_id,
            'cert_type', v_cert_record.cert_type,
            'recommendations', v_recommendations
        ),
        v_score
    );

    -- Preparar resultado
    v_result := json_build_object(
        'success', v_is_valid,
        'cert_id', p_cert_id,
        'cert_type', v_cert_record.cert_type,
        'score_risco', v_score,
        'recommendations', v_recommendations,
        'details', json_build_object(
            'cert_data', v_cert_record.cert_data,
            'expires_at', v_cert_record.expires_at
        )
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.19 SSO Avançado
CREATE OR REPLACE FUNCTION iam_access_control.advanced_sso(
    p_user_id TEXT,
    p_service_provider TEXT,
    p_request_data JSONB,
    p_context_data JSONB,
    p_required_level INT
)
RETURNS JSON AS $$
DECLARE
    v_sso_id UUID;
    v_session_id UUID;
    v_score INT := 0;
    v_recommendations JSONB;
    v_result JSON;
BEGIN
    -- Gerar ID único para SSO
    v_sso_id := gen_random_uuid();

    -- Obter ID da sessão atual
    SELECT session_id INTO v_session_id
    FROM iam_access_control.sessions
    WHERE user_id = p_user_id
    AND expires_at > current_timestamp;

    -- Verificar contexto
    IF p_context_data->>'location_valid' = 'true' THEN
        v_score := v_score + 20;
    END IF;

    IF p_context_data->>'device_valid' = 'true' THEN
        v_score := v_score + 20;
    END IF;

    IF p_context_data->>'time_valid' = 'true' THEN
        v_score := v_score + 15;
    END IF;

    IF p_context_data->>'behavior_valid' = 'true' THEN
        v_score := v_score + 25;
    END IF;

    IF p_context_data->>'risk_level'::INT <= 2 THEN
        v_score := v_score + 20;
    END IF;

    -- Verificar nível de autenticação
    IF v_score >= p_required_level THEN
        -- Registrar SSO request
        INSERT INTO iam_access_control.sso_requests (
            sso_id,
            user_id,
            session_id,
            service_provider,
            request_data,
            context_data,
            score_risco,
            created_at,
            expires_at
        ) VALUES (
            v_sso_id,
            p_user_id,
            v_session_id,
            p_service_provider,
            p_request_data,
            p_context_data,
            v_score,
            current_timestamp,
            current_timestamp + INTERVAL '15 minutes'
        );

        -- Gerar recomendações
        v_recommendations := jsonb_build_object(
            'action', 'allow',
            'reason', 'SSO successful with required authentication level',
            'score', v_score,
            'details', jsonb_build_object(
                'service_provider', p_service_provider,
                'context_data', p_context_data,
                'required_level', p_required_level
            )
        );

        -- Registrar evento de segurança
        INSERT INTO iam_access_control.security_events (
            event_type,
            event_time,
            event_details,
            score_risco
        ) VALUES (
            'SSO_SUCCESS',
            current_timestamp,
            jsonb_build_object(
                'sso_id', v_sso_id,
                'user_id', p_user_id,
                'service_provider', p_service_provider,
                'recommendations', v_recommendations
            ),
            v_score
        );
    ELSE
        -- Gerar recomendações para falha
        v_recommendations := jsonb_build_object(
            'action', 'deny',
            'reason', 'Insufficient authentication level',
            'score', v_score,
            'details', jsonb_build_object(
                'service_provider', p_service_provider,
                'context_data', p_context_data,
                'required_level', p_required_level
            )
        );

        -- Registrar evento de segurança
        INSERT INTO iam_access_control.security_events (
            event_type,
            event_time,
            event_details,
            score_risco
        ) VALUES (
            'SSO_FAILED',
            current_timestamp,
            jsonb_build_object(
                'user_id', p_user_id,
                'service_provider', p_service_provider,
                'recommendations', v_recommendations
            ),
            v_score
        );
    END IF;

    -- Preparar resultado
    v_result := json_build_object(
        'success', v_score >= p_required_level,
        'sso_id', v_sso_id,
        'user_id', p_user_id,
        'service_provider', p_service_provider,
        'score_risco', v_score,
        'recommendations', v_recommendations,
        'details', json_build_object(
            'request_data', p_request_data,
            'context_data', p_context_data,
            'required_level', p_required_level
        )
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.20 Processar SSO Response Avançado
CREATE OR REPLACE FUNCTION iam_access_control.process_advanced_sso_response(
    p_sso_id UUID,
    p_response_data JSONB,
    p_context_data JSONB
)
RETURNS JSON AS $$
DECLARE
    v_request_record RECORD;
    v_is_valid BOOLEAN := FALSE;
    v_score INT := 0;
    v_recommendations JSONB;
    v_result JSON;
BEGIN
    -- Obter detalhes do SSO request
    SELECT * INTO v_request_record
    FROM iam_access_control.sso_requests
    WHERE sso_id = p_sso_id
    AND expires_at > current_timestamp;

    IF v_request_record IS NOT NULL THEN
        -- Verificar integridade da resposta
        IF sso.verify_response(
            v_request_record.user_id,
            p_response_data,
            v_request_record.request_data,
            v_request_record.context_data
        ) THEN
            v_is_valid := TRUE;
            v_score := v_score + 50;
        END IF;

        -- Ajustar score baseado no contexto
        IF p_context_data->>'device_secure' = 'true' THEN
            v_score := v_score + 10;
        END IF;

        -- Atualizar status do SSO request
        UPDATE iam_access_control.sso_requests
        SET 
            response_data = p_response_data,
            processed_at = current_timestamp,
            is_valid = v_is_valid,
            score_risco = v_score
        WHERE sso_id = p_sso_id;
    END IF;

    -- Gerar recomendações
    v_recommendations := jsonb_build_object(
        'action', CASE 
            WHEN v_is_valid THEN 'allow'
            ELSE 'deny'
        END,
        'reason', CASE 
            WHEN v_is_valid THEN 'SSO response valid'
            ELSE 'SSO response invalid'
        END,
        'score', v_score,
        'details', jsonb_build_object(
            'service_provider', v_request_record.service_provider,
            'context_data', p_context_data
        )
    );

    -- Registrar evento de segurança
    INSERT INTO iam_access_control.security_events (
        event_type,
        event_time,
        event_details,
        score_risco
    ) VALUES (
        CASE 
            WHEN v_is_valid THEN 'SSO_RESPONSE_VALID'
            ELSE 'SSO_RESPONSE_INVALID'
        END,
        current_timestamp,
        jsonb_build_object(
            'sso_id', p_sso_id,
            'service_provider', v_request_record.service_provider,
            'recommendations', v_recommendations
        ),
        v_score
    );

    -- Preparar resultado
    v_result := json_build_object(
        'success', v_is_valid,
        'sso_id', p_sso_id,
        'service_provider', v_request_record.service_provider,
        'score_risco', v_score,
        'recommendations', v_recommendations,
        'details', json_build_object(
            'request_data', v_request_record.request_data,
            'response_data', p_response_data,
            'context_data', p_context_data
        )
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.21 Recuperação de Acesso
CREATE OR REPLACE FUNCTION iam_access_control.access_recovery(
    p_recovery_method TEXT, -- 'EMAIL', 'SMS', 'PHONE', 'BACKUP_CODE', 'SECURITY_QUESTIONS', 'BIOMETRIC'
    p_recovery_data JSONB,
    p_context_data JSONB
)
RETURNS JSON AS $$
DECLARE
    v_recovery_id UUID;
    v_score INT := 0;
    v_recommendations JSONB;
    v_result JSON;
BEGIN
    -- Gerar ID único para recuperação
    v_recovery_id := gen_random_uuid();

    -- Verificar método de recuperação
    CASE 
        WHEN p_recovery_method = 'EMAIL' THEN
            IF email.verify_recovery(
                p_recovery_data->>'email',
                p_recovery_data->>'code'
            ) THEN
                v_score := v_score + 60;
            END IF;
        WHEN p_recovery_method = 'SMS' THEN
            IF sms.verify_recovery(
                p_recovery_data->>'phone',
                p_recovery_data->>'code'
            ) THEN
                v_score := v_score + 70;
            END IF;
        WHEN p_recovery_method = 'PHONE' THEN
            IF phone.verify_recovery(
                p_recovery_data->>'phone',
                p_recovery_data->>'code'
            ) THEN
                v_score := v_score + 75;
            END IF;
        WHEN p_recovery_method = 'BACKUP_CODE' THEN
            IF backup_code.verify_code(
                p_recovery_data->>'backup_code'
            ) THEN
                v_score := v_score + 80;
            END IF;
        WHEN p_recovery_method = 'SECURITY_QUESTIONS' THEN
            IF security_questions.verify_answers(
                p_recovery_data->'questions'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_recovery_method = 'BIOMETRIC' THEN
            IF biometric.verify_recovery(
                p_recovery_data->>'bio_type',
                p_recovery_data->>'bio_data'
            ) THEN
                v_score := v_score + 90;
            END IF;
    END CASE;

    -- Ajustar score baseado no contexto
    IF p_context_data->>'device_secure' = 'true' THEN
        v_score := v_score + 10;
    END IF;

    -- Registrar tentativa de recuperação
    INSERT INTO iam_access_control.recovery_attempts (
        recovery_id,
        recovery_method,
        recovery_data,
        context_data,
        score_risco,
        created_at
    ) VALUES (
        v_recovery_id,
        p_recovery_method,
        p_recovery_data,
        p_context_data,
        v_score,
        current_timestamp
    );

    -- Gerar recomendações
    v_recommendations := jsonb_build_object(
        'action', CASE 
            WHEN v_score >= 70 THEN 'allow'
            WHEN v_score >= 50 THEN 'review'
            ELSE 'deny'
        END,
        'reason', CASE 
            WHEN v_score >= 70 THEN 'High confidence recovery'
            WHEN v_score >= 50 THEN 'Medium confidence recovery'
            ELSE 'Low confidence recovery'
        END,
        'score', v_score,
        'details', jsonb_build_object(
            'recovery_method', p_recovery_method,
            'context_data', p_context_data
        )
    );

    -- Registrar evento de segurança
    INSERT INTO iam_access_control.security_events (
        event_type,
        event_time,
        event_details,
        score_risco
    ) VALUES (
        CASE 
            WHEN v_score >= 70 THEN 'RECOVERY_SUCCESS'
            ELSE 'RECOVERY_ATTEMPT'
        END,
        current_timestamp,
        jsonb_build_object(
            'recovery_id', v_recovery_id,
            'recovery_method', p_recovery_method,
            'recommendations', v_recommendations
        ),
        v_score
    );

    -- Preparar resultado
    v_result := json_build_object(
        'success', v_score >= 70,
        'recovery_id', v_recovery_id,
        'recovery_method', p_recovery_method,
        'score_risco', v_score,
        'recommendations', v_recommendations,
        'details', json_build_object(
            'recovery_data', p_recovery_data,
            'context_data', p_context_data
        )
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.22 Gerar Código de Recuperação
CREATE OR REPLACE FUNCTION iam_access_control.generate_recovery_code(
    p_user_id TEXT,
    p_recovery_method TEXT,
    p_recovery_data JSONB
)
RETURNS JSON AS $$
DECLARE
    v_recovery_id UUID;
    v_recovery_code TEXT;
    v_result JSON;
BEGIN
    -- Gerar ID único para código de recuperação
    v_recovery_id := gen_random_uuid();

    -- Gerar código de recuperação baseado no método
    CASE 
        WHEN p_recovery_method = 'BACKUP_CODE' THEN
            v_recovery_code := encode(gen_random_bytes(16), 'hex');
        WHEN p_recovery_method = 'SECURITY_QUESTIONS' THEN
            v_recovery_code := encode(gen_random_bytes(32), 'base64');
        WHEN p_recovery_method = 'BIOMETRIC' THEN
            v_recovery_code := encode(gen_random_bytes(64), 'base64');
    END CASE;

    -- Registrar código de recuperação
    INSERT INTO iam_access_control.recovery_codes (
        recovery_id,
        user_id,
        recovery_method,
        recovery_code,
        recovery_data,
        created_at,
        expires_at
    ) VALUES (
        v_recovery_id,
        p_user_id,
        p_recovery_method,
        v_recovery_code,
        p_recovery_data,
        current_timestamp,
        current_timestamp + INTERVAL '24 hours'
    );

    -- Preparar resultado
    v_result := json_build_object(
        'success', TRUE,
        'recovery_id', v_recovery_id,
        'recovery_method', p_recovery_method,
        'recovery_code', v_recovery_code,
        'expires_at', current_timestamp + INTERVAL '24 hours'
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.23 Autenticação por Cartão
CREATE OR REPLACE FUNCTION iam_access_control.card_authentication(
    p_card_type TEXT, -- 'CREDIT', 'DEBIT', 'ID', 'SECURITY', 'SMART', 'HARDWARE', 'SOFTWARE', 'CLOUD', 'LOCAL', 'OFFLINE', 'ONLINE', 'HYBRID'
    p_card_data JSONB,
    p_context_data JSONB
)
RETURNS JSON AS $$
DECLARE
    v_auth_id UUID;
    v_score INT := 0;
    v_recommendations JSONB;
    v_result JSON;
BEGIN
    -- Gerar ID único para autenticação
    v_auth_id := gen_random_uuid();

    -- Verificar tipo de cartão e aplicar score
    CASE 
        WHEN p_card_type = 'CREDIT' THEN
            IF card.verify_credit(
                p_card_data->>'card_number',
                p_card_data->>'cvv',
                p_card_data->>'expiry_date'
            ) THEN
                v_score := v_score + 70;
            END IF;
        WHEN p_card_type = 'DEBIT' THEN
            IF card.verify_debit(
                p_card_data->>'card_number',
                p_card_data->>'pin',
                p_card_data->>'account_number'
            ) THEN
                v_score := v_score + 75;
            END IF;
        WHEN p_card_type = 'ID' THEN
            IF card.verify_id(
                p_card_data->>'card_number',
                p_card_data->>'personal_number',
                p_card_data->>'expiry_date'
            ) THEN
                v_score := v_score + 80;
            END IF;
        WHEN p_card_type = 'SECURITY' THEN
            IF card.verify_security(
                p_card_data->>'card_number',
                p_card_data->>'security_code',
                p_card_data->>'challenge_response'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_card_type = 'SMART' THEN
            IF card.verify_smart(
                p_card_data->>'card_number',
                p_card_data->>'chip_data',
                p_card_data->>'pin'
            ) THEN
                v_score := v_score + 90;
            END IF;
        WHEN p_card_type = 'HARDWARE' THEN
            IF card.verify_hardware(
                p_card_data->>'card_number',
                p_card_data->>'hardware_id',
                p_card_data->>'challenge_response'
            ) THEN
                v_score := v_score + 95;
            END IF;
        WHEN p_card_type = 'SOFTWARE' THEN
            IF card.verify_software(
                p_card_data->>'card_number',
                p_card_data->>'software_token',
                p_card_data->>'timestamp'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_card_type = 'CLOUD' THEN
            IF card.verify_cloud(
                p_card_data->>'card_number',
                p_card_data->>'cloud_token',
                p_card_data->>'session_id'
            ) THEN
                v_score := v_score + 80;
            END IF;
        WHEN p_card_type = 'LOCAL' THEN
            IF card.verify_local(
                p_card_data->>'card_number',
                p_card_data->>'local_token',
                p_card_data->>'device_id'
            ) THEN
                v_score := v_score + 75;
            END IF;
        WHEN p_card_type = 'OFFLINE' THEN
            IF card.verify_offline(
                p_card_data->>'card_number',
                p_card_data->>'offline_token',
                p_card_data->>'device_id'
            ) THEN
                v_score := v_score + 70;
            END IF;
        WHEN p_card_type = 'ONLINE' THEN
            IF card.verify_online(
                p_card_data->>'card_number',
                p_card_data->>'online_token',
                p_card_data->>'session_id'
            ) THEN
                v_score := v_score + 80;
            END IF;
        WHEN p_card_type = 'HYBRID' THEN
            IF card.verify_hybrid(
                p_card_data->>'card_number',
                p_card_data->>'hybrid_token',
                p_card_data->>'device_id',
                p_card_data->>'session_id'
            ) THEN
                v_score := v_score + 85;
            END IF;
    END CASE;

    -- Ajustar score baseado no contexto
    IF p_context_data->>'device_secure' = 'true' THEN
        v_score := v_score + 10;
    END IF;

    -- Registrar tentativa de autenticação
    INSERT INTO iam_access_control.card_auth_attempts (
        auth_id,
        card_type,
        card_data,
        context_data,
        score_risco,
        created_at
    ) VALUES (
        v_auth_id,
        p_card_type,
        p_card_data,
        p_context_data,
        v_score,
        current_timestamp
    );

    -- Gerar recomendações
    v_recommendations := jsonb_build_object(
        'action', CASE 
            WHEN v_score >= 70 THEN 'allow'
            WHEN v_score >= 50 THEN 'review'
            ELSE 'deny'
        END,
        'reason', CASE 
            WHEN v_score >= 70 THEN 'High confidence authentication'
            WHEN v_score >= 50 THEN 'Medium confidence authentication'
            ELSE 'Low confidence authentication'
        END,
        'score', v_score,
        'details', jsonb_build_object(
            'card_type', p_card_type,
            'context_data', p_context_data
        )
    );

    -- Registrar evento de segurança
    INSERT INTO iam_access_control.security_events (
        event_type,
        event_time,
        event_details,
        score_risco
    ) VALUES (
        CASE 
            WHEN v_score >= 70 THEN 'CARD_AUTH_SUCCESS'
            ELSE 'CARD_AUTH_ATTEMPT'
        END,
        current_timestamp,
        jsonb_build_object(
            'auth_id', v_auth_id,
            'card_type', p_card_type,
            'recommendations', v_recommendations
        ),
        v_score
    );

    -- Preparar resultado
    v_result := json_build_object(
        'success', v_score >= 70,
        'auth_id', v_auth_id,
        'card_type', p_card_type,
        'score_risco', v_score,
        'recommendations', v_recommendations,
        'details', json_build_object(
            'card_data', p_card_data,
            'context_data', p_context_data
        )
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.24 Gerar Token de Cartão
CREATE OR REPLACE FUNCTION iam_access_control.generate_card_token(
    p_card_type TEXT,
    p_card_data JSONB,
    p_expiration_minutes INT DEFAULT 30
)
RETURNS JSON AS $$
DECLARE
    v_token_id UUID;
    v_card_token TEXT;
    v_result JSON;
BEGIN
    -- Gerar ID único para token
    v_token_id := gen_random_uuid();

    -- Gerar token baseado no tipo de cartão
    CASE 
        WHEN p_card_type = 'CREDIT' THEN
            v_card_token := encode(gen_random_bytes(16), 'hex');
        WHEN p_card_type = 'DEBIT' THEN
            v_card_token := encode(gen_random_bytes(24), 'hex');
        WHEN p_card_type = 'ID' THEN
            v_card_token := encode(gen_random_bytes(32), 'base64');
        WHEN p_card_type = 'SECURITY' THEN
            v_card_token := encode(gen_random_bytes(48), 'base64');
        WHEN p_card_type = 'SMART' THEN
            v_card_token := encode(gen_random_bytes(64), 'base64');
        WHEN p_card_type = 'HARDWARE' THEN
            v_card_token := encode(gen_random_bytes(80), 'base64');
        WHEN p_card_type = 'SOFTWARE' THEN
            v_card_token := encode(gen_random_bytes(48), 'hex');
        WHEN p_card_type = 'CLOUD' THEN
            v_card_token := encode(gen_random_bytes(64), 'hex');
        WHEN p_card_type = 'LOCAL' THEN
            v_card_token := encode(gen_random_bytes(32), 'hex');
        WHEN p_card_type = 'OFFLINE' THEN
            v_card_token := encode(gen_random_bytes(24), 'hex');
        WHEN p_card_type = 'ONLINE' THEN
            v_card_token := encode(gen_random_bytes(48), 'hex');
        WHEN p_card_type = 'HYBRID' THEN
            v_card_token := encode(gen_random_bytes(64), 'hex');
    END CASE;

    -- Registrar token
    INSERT INTO iam_access_control.card_tokens (
        token_id,
        card_type,
        card_data,
        card_token,
        created_at,
        expires_at
    ) VALUES (
        v_token_id,
        p_card_type,
        p_card_data,
        v_card_token,
        current_timestamp,
        current_timestamp + INTERVAL '1 minute' * p_expiration_minutes
    );

    -- Preparar resultado
    v_result := json_build_object(
        'success', TRUE,
        'token_id', v_token_id,
        'card_type', p_card_type,
        'card_token', v_card_token,
        'expires_at', current_timestamp + INTERVAL '1 minute' * p_expiration_minutes
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.25 Autenticação por Protocolo
CREATE OR REPLACE FUNCTION iam_access_control.protocol_authentication(
    p_protocol_type TEXT, -- 'OAUTH2', 'OIDC', 'SAML', 'WS-FED', 'KERBEROS', 'LDAP', 'RADIUS', 'TACACS', 'DIAMETER', 'DIAMETER++', 'DIAMETER+++', 'DIAMETER++++', 'DIAMETER+++++', 'DIAMETER++++++', 'DIAMETER+++++++', 'DIAMETER++++++++', 'DIAMETER+++++++++', 'DIAMETER++++++++++', 'DIAMETER+++++++++++', 'DIAMETER++++++++++++'
    p_protocol_data JSONB,
    p_context_data JSONB
)
RETURNS JSON AS $$
DECLARE
    v_auth_id UUID;
    v_score INT := 0;
    v_recommendations JSONB;
    v_result JSON;
BEGIN
    -- Gerar ID único para autenticação
    v_auth_id := gen_random_uuid();

    -- Verificar tipo de protocolo e aplicar score
    CASE 
        WHEN p_protocol_type = 'OAUTH2' THEN
            IF protocol.verify_oauth2(
                p_protocol_data->>'client_id',
                p_protocol_data->>'client_secret',
                p_protocol_data->>'access_token',
                p_protocol_data->>'refresh_token'
            ) THEN
                v_score := v_score + 75;
            END IF;
        WHEN p_protocol_type = 'OIDC' THEN
            IF protocol.verify_oidc(
                p_protocol_data->>'client_id',
                p_protocol_data->>'id_token',
                p_protocol_data->>'access_token',
                p_protocol_data->>'refresh_token'
            ) THEN
                v_score := v_score + 80;
            END IF;
        WHEN p_protocol_type = 'SAML' THEN
            IF protocol.verify_saml(
                p_protocol_data->>'assertion',
                p_protocol_data->>'signature',
                p_protocol_data->>'issuer'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_protocol_type = 'WS-FED' THEN
            IF protocol.verify_ws_fed(
                p_protocol_data->>'wa',
                p_protocol_data->>'wresult',
                p_protocol_data->>'wctx'
            ) THEN
                v_score := v_score + 90;
            END IF;
        WHEN p_protocol_type = 'KERBEROS' THEN
            IF protocol.verify_kerberos(
                p_protocol_data->>'ticket',
                p_protocol_data->>'service_ticket',
                p_protocol_data->>'session_key'
            ) THEN
                v_score := v_score + 95;
            END IF;
        WHEN p_protocol_type = 'LDAP' THEN
            IF protocol.verify_ldap(
                p_protocol_data->>'username',
                p_protocol_data->>'password',
                p_protocol_data->>'base_dn'
            ) THEN
                v_score := v_score + 80;
            END IF;
        WHEN p_protocol_type = 'RADIUS' THEN
            IF protocol.verify_radius(
                p_protocol_data->>'username',
                p_protocol_data->>'password',
                p_protocol_data->>'shared_secret'
            ) THEN
                v_score := v_score + 75;
            END IF;
        WHEN p_protocol_type = 'TACACS' THEN
            IF protocol.verify_tacacs(
                p_protocol_data->>'username',
                p_protocol_data->>'password',
                p_protocol_data->>'shared_secret'
            ) THEN
                v_score := v_score + 70;
            END IF;
        WHEN p_protocol_type = 'DIAMETER' THEN
            IF protocol.verify_diameter(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_protocol_type = 'DIAMETER++' THEN
            IF protocol.verify_diameter_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'additional_data'
            ) THEN
                v_score := v_score + 90;
            END IF;
        WHEN p_protocol_type = 'DIAMETER+++' THEN
            IF protocol.verify_diameter_plus_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'additional_data',
                p_protocol_data->>'security_data'
            ) THEN
                v_score := v_score + 95;
            END IF;
        WHEN p_protocol_type = 'DIAMETER++++' THEN
            IF protocol.verify_diameter_plus_plus_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'additional_data',
                p_protocol_data->>'security_data',
                p_protocol_data->>'audit_data'
            ) THEN
                v_score := v_score + 100;
            END IF;
        WHEN p_protocol_type = 'DIAMETER+++++' THEN
            IF protocol.verify_diameter_plus_plus_plus_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'additional_data',
                p_protocol_data->>'security_data',
                p_protocol_data->>'audit_data',
                p_protocol_data->>'trace_data'
            ) THEN
                v_score := v_score + 105;
            END IF;
        WHEN p_protocol_type = 'DIAMETER++++++' THEN
            IF protocol.verify_diameter_plus_plus_plus_plus_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'additional_data',
                p_protocol_data->>'security_data',
                p_protocol_data->>'audit_data',
                p_protocol_data->>'trace_data',
                p_protocol_data->>'monitoring_data'
            ) THEN
                v_score := v_score + 110;
            END IF;
        WHEN p_protocol_type = 'DIAMETER+++++++' THEN
            IF protocol.verify_diameter_plus_plus_plus_plus_plus_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'additional_data',
                p_protocol_data->>'security_data',
                p_protocol_data->>'audit_data',
                p_protocol_data->>'trace_data',
                p_protocol_data->>'monitoring_data',
                p_protocol_data->>'analytics_data'
            ) THEN
                v_score := v_score + 115;
            END IF;
        WHEN p_protocol_type = 'DIAMETER++++++++' THEN
            IF protocol.verify_diameter_plus_plus_plus_plus_plus_plus_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'additional_data',
                p_protocol_data->>'security_data',
                p_protocol_data->>'audit_data',
                p_protocol_data->>'trace_data',
                p_protocol_data->>'monitoring_data',
                p_protocol_data->>'analytics_data',
                p_protocol_data->>'intelligence_data'
            ) THEN
                v_score := v_score + 120;
            END IF;
        WHEN p_protocol_type = 'DIAMETER+++++++++' THEN
            IF protocol.verify_diameter_plus_plus_plus_plus_plus_plus_plus_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'additional_data',
                p_protocol_data->>'security_data',
                p_protocol_data->>'audit_data',
                p_protocol_data->>'trace_data',
                p_protocol_data->>'monitoring_data',
                p_protocol_data->>'analytics_data',
                p_protocol_data->>'intelligence_data',
                p_protocol_data->>'automation_data'
            ) THEN
                v_score := v_score + 125;
            END IF;
        WHEN p_protocol_type = 'DIAMETER++++++++++' THEN
            IF protocol.verify_diameter_plus_plus_plus_plus_plus_plus_plus_plus_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'additional_data',
                p_protocol_data->>'security_data',
                p_protocol_data->>'audit_data',
                p_protocol_data->>'trace_data',
                p_protocol_data->>'monitoring_data',
                p_protocol_data->>'analytics_data',
                p_protocol_data->>'intelligence_data',
                p_protocol_data->>'automation_data',
                p_protocol_data->>'orchestration_data'
            ) THEN
                v_score := v_score + 130;
            END IF;
        WHEN p_protocol_type = 'DIAMETER+++++++++++' THEN
            IF protocol.verify_diameter_plus_plus_plus_plus_plus_plus_plus_plus_plus_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'additional_data',
                p_protocol_data->>'security_data',
                p_protocol_data->>'audit_data',
                p_protocol_data->>'trace_data',
                p_protocol_data->>'monitoring_data',
                p_protocol_data->>'analytics_data',
                p_protocol_data->>'intelligence_data',
                p_protocol_data->>'automation_data',
                p_protocol_data->>'orchestration_data',
                p_protocol_data->>'integration_data'
            ) THEN
                v_score := v_score + 135;
            END IF;
        WHEN p_protocol_type = 'DIAMETER++++++++++++' THEN
            IF protocol.verify_diameter_plus_plus_plus_plus_plus_plus_plus_plus_plus_plus_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'additional_data',
                p_protocol_data->>'security_data',
                p_protocol_data->>'audit_data',
                p_protocol_data->>'trace_data',
                p_protocol_data->>'monitoring_data',
                p_protocol_data->>'analytics_data',
                p_protocol_data->>'intelligence_data',
                p_protocol_data->>'automation_data',
                p_protocol_data->>'orchestration_data',
                p_protocol_data->>'integration_data',
                p_protocol_data->>'collaboration_data'
            ) THEN
                v_score := v_score + 140;
            END IF;
    END CASE;

    -- Ajustar score baseado no contexto
    IF p_context_data->>'device_secure' = 'true' THEN
        v_score := v_score + 10;
    END IF;

    -- Registrar tentativa de autenticação
    INSERT INTO iam_access_control.protocol_auth_attempts (
        auth_id,
        protocol_type,
        protocol_data,
        context_data,
        score_risco,
        created_at
    ) VALUES (
        v_auth_id,
        p_protocol_type,
        p_protocol_data,
        p_context_data,
        v_score,
        current_timestamp
    );

    -- Gerar recomendações
    v_recommendations := jsonb_build_object(
        'action', CASE 
            WHEN v_score >= 70 THEN 'allow'
            WHEN v_score >= 50 THEN 'review'
            ELSE 'deny'
        END,
        'reason', CASE 
            WHEN v_score >= 70 THEN 'High confidence authentication'
            WHEN v_score >= 50 THEN 'Medium confidence authentication'
            ELSE 'Low confidence authentication'
        END,
        'score', v_score,
        'details', jsonb_build_object(
            'protocol_type', p_protocol_type,
            'context_data', p_context_data
        )
    );

    -- Registrar evento de segurança
    INSERT INTO iam_access_control.security_events (
        event_type,
        event_time,
        event_details,
        score_risco
    ) VALUES (
        CASE 
            WHEN v_score >= 70 THEN 'PROTOCOL_AUTH_SUCCESS'
            ELSE 'PROTOCOL_AUTH_ATTEMPT'
        END,
        current_timestamp,
        jsonb_build_object(
            'auth_id', v_auth_id,
            'protocol_type', p_protocol_type,
            'recommendations', v_recommendations
        ),
        v_score
    );

    -- Preparar resultado
    v_result := json_build_object(
        'success', v_score >= 70,
        'auth_id', v_auth_id,
        'protocol_type', p_protocol_type,
        'score_risco', v_score,
        'recommendations', v_recommendations,
        'details', json_build_object(
            'protocol_data', p_protocol_data,
            'context_data', p_context_data
        )
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.26 Gerar Token de Protocolo
CREATE OR REPLACE FUNCTION iam_access_control.generate_protocol_token(
    p_protocol_type TEXT,
    p_protocol_data JSONB,
    p_required_level INT
)
RETURNS JSON AS $$
DECLARE
    v_auth_id UUID;
    v_score INT := 0;
    v_recommendations JSONB;
    v_result JSON;
BEGIN
    -- Gerar ID único para autenticação
    v_auth_id := gen_random_uuid();

    -- Verificar tipo de protocolo e aplicar score
    CASE 
        WHEN p_protocol_type = 'OAUTH2' THEN
            IF protocol.verify_oauth2(
                p_protocol_data->>'client_id',
                p_protocol_data->>'client_secret',
                p_protocol_data->>'redirect_uri',
                p_protocol_data->>'scope'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_protocol_type = 'OPENID_CONNECT' THEN
            IF protocol.verify_openid_connect(
                p_protocol_data->>'client_id',
                p_protocol_data->>'client_secret',
                p_protocol_data->>'redirect_uri',
                p_protocol_data->>'scope'
            ) THEN
                v_score := v_score + 90;
            END IF;
        WHEN p_protocol_type = 'SAML' THEN
            IF protocol.verify_saml(
                p_protocol_data->>'entity_id',
                p_protocol_data->>'assertion',
                p_protocol_data->>'signature',
                p_protocol_data->>'expiry'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_protocol_type = 'WS_FEDERATION' THEN
            IF protocol.verify_ws_federation(
                p_protocol_data->>'realm',
                p_protocol_data->>'reply',
                p_protocol_data->>'wctx',
                p_protocol_data->>'wresult'
            ) THEN
                v_score := v_score + 80;
            END IF;
        WHEN p_protocol_type = 'KERBEROS' THEN
            IF protocol.verify_kerberos(
                p_protocol_data->>'ticket',
                p_protocol_data->>'session_key',
                p_protocol_data->>'expiry',
                p_protocol_data->>'encryption_type'
            ) THEN
                v_score := v_score + 95;
            END IF;
        WHEN p_protocol_type = 'LDAP' THEN
            IF protocol.verify_ldap(
                p_protocol_data->>'dn',
                p_protocol_data->>'credentials',
                p_protocol_data->>'server',
                p_protocol_data->>'encryption'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_protocol_type = 'RADIUS' THEN
            IF protocol.verify_radius(
                p_protocol_data->>'username',
                p_protocol_data->>'password',
                p_protocol_data->>'server',
                p_protocol_data->>'secret'
            ) THEN
                v_score := v_score + 80;
            END IF;
        WHEN p_protocol_type = 'TACACS_PLUS' THEN
            IF protocol.verify_tacacs_plus(
                p_protocol_data->>'username',
                p_protocol_data->>'password',
                p_protocol_data->>'server',
                p_protocol_data->>'secret'
            ) THEN
                v_score := v_score + 75;
            END IF;
        WHEN p_protocol_type = 'DIAMETER' THEN
            IF protocol.verify_diameter(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'user_name'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_protocol_type = 'DIAMETER++' THEN
            IF protocol.verify_diameter_plus_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'user_name'
            ) THEN
                v_score := v_score + 90;
            END IF;
        WHEN p_protocol_type = 'DIAMETER+++' THEN
            IF protocol.verify_diameter_plus_plus_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'user_name'
            ) THEN
                v_score := v_score + 95;
            END IF;
        WHEN p_protocol_type = 'DIAMETER++++' THEN
            IF protocol.verify_diameter_plus_plus_plus_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'user_name'
            ) THEN
                v_score := v_score + 100;
            END IF;
        WHEN p_protocol_type = 'DIAMETER+++++' THEN
            IF protocol.verify_diameter_plus_plus_plus_plus_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'user_name'
            ) THEN
                v_score := v_score + 105;
            END IF;
        WHEN p_protocol_type = 'DIAMETER++++++' THEN
            IF protocol.verify_diameter_plus_plus_plus_plus_plus_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'user_name'
            ) THEN
                v_score := v_score + 110;
            END IF;
        WHEN p_protocol_type = 'DIAMETER+++++++' THEN
            IF protocol.verify_diameter_plus_plus_plus_plus_plus_plus_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'user_name'
            ) THEN
                v_score := v_score + 115;
            END IF;
        WHEN p_protocol_type = 'DIAMETER++++++++' THEN
            IF protocol.verify_diameter_plus_plus_plus_plus_plus_plus_plus_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'user_name'
            ) THEN
                v_score := v_score + 120;
            END IF;
        WHEN p_protocol_type = 'DIAMETER+++++++++' THEN
            IF protocol.verify_diameter_plus_plus_plus_plus_plus_plus_plus_plus_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'user_name'
            ) THEN
                v_score := v_score + 125;
            END IF;
        WHEN p_protocol_type = 'DIAMETER++++++++++' THEN
            IF protocol.verify_diameter_plus_plus_plus_plus_plus_plus_plus_plus_plus_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'user_name'
            ) THEN
                v_score := v_score + 130;
            END IF;
        WHEN p_protocol_type = 'DIAMETER+++++++++++' THEN
            IF protocol.verify_diameter_plus_plus_plus_plus_plus_plus_plus_plus_plus_plus_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'user_name'
            ) THEN
                v_score := v_score + 135;
            END IF;
        WHEN p_protocol_type = 'DIAMETER++++++++++++' THEN
            IF protocol.verify_diameter_plus_plus_plus_plus_plus_plus_plus_plus_plus_plus_plus_plus(
                p_protocol_data->>'session_id',
                p_protocol_data->>'auth_application_id',
                p_protocol_data->>'auth_request_type',
                p_protocol_data->>'user_name'
            ) THEN
                v_score := v_score + 140;
            END IF;
    END CASE;

    -- Ajustar score baseado no contexto
    IF p_context_data->>'device_secure' = 'true' THEN
        v_score := v_score + 10;
    END IF;

    -- Verificar nível de autenticação
    IF v_score >= p_required_level THEN
        -- Gerar token
        v_token := gen_random_uuid();
        
        -- Registrar tentativa de autenticação
        INSERT INTO iam_access_control.protocol_auth_attempts (
            auth_id,
            protocol_type,
            protocol_data,
            required_level,
            score_risco,
            token_id,
            created_at
        ) VALUES (
            v_auth_id,
            p_protocol_type,
            p_protocol_data,
            p_required_level,
            v_score,
            v_token,
            current_timestamp
        );

        -- Gerar recomendações
        v_recommendations := jsonb_build_object(
            'action', 'allow',
            'reason', 'Protocol authentication successful',
            'score', v_score,
            'details', jsonb_build_object(
                'protocol_type', p_protocol_type,
                'required_level', p_required_level,
                'protocol_data', p_protocol_data
            )
        );

        -- Registrar evento de segurança
        INSERT INTO iam_access_control.security_events (
            event_type,
            event_time,
            event_details,
            score_risco
        ) VALUES (
            'PROTOCOL_AUTH_SUCCESS',
            current_timestamp,
            jsonb_build_object(
                'auth_id', v_auth_id,
                'protocol_type', p_protocol_type,
                'recommendations', v_recommendations
            ),
            v_score
        );

        -- Retornar resultado
        RETURN json_build_object(
            'success', true,
            'auth_id', v_auth_id,
            'token', v_token,
            'score', v_score,
            'recommendations', v_recommendations
        );
    ELSE
        -- Gerar recomendações para falha
        v_recommendations := jsonb_build_object(
            'action', 'block',
            'reason', 'Insufficient protocol authentication score',
            'score', v_score,
            'details', jsonb_build_object(
                'protocol_type', p_protocol_type,
                'required_level', p_required_level,
                'protocol_data', p_protocol_data
            )
        );

        -- Registrar evento de segurança
        INSERT INTO iam_access_control.security_events (
            event_type,
            event_time,
            event_details,
            score_risco
        ) VALUES (
            'PROTOCOL_AUTH_FAILURE',
            current_timestamp,
            jsonb_build_object(
                'auth_id', v_auth_id,
                'protocol_type', p_protocol_type,
                'recommendations', v_recommendations
            ),
            v_score
        );

        -- Retornar resultado
        RETURN json_build_object(
            'success', false,
            'auth_id', v_auth_id,
            'score', v_score,
            'recommendations', v_recommendations
        );
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.33 Autenticação por Posse
CREATE OR REPLACE FUNCTION iam_access_control.possession_authentication(
    p_possession_type TEXT,
    p_possession_data JSONB,
    p_required_level INT
)
RETURNS JSON AS $$
DECLARE
    v_auth_id UUID;
    v_score INT := 0;
    v_recommendations JSONB;
    v_result JSON;
BEGIN
    -- Gerar ID único para autenticação
    v_auth_id := gen_random_uuid();

    -- Verificar tipo de posse e aplicar score
    CASE 
        WHEN p_possession_type = 'APP' THEN
            IF possession.verify_app(
                p_possession_data->>'app_id',
                p_possession_data->>'app_version',
                p_possession_data->>'security_level',
                p_possession_data->>'encryption_status'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_possession_type = 'SMS' THEN
            IF possession.verify_sms(
                p_possession_data->>'phone_number',
                p_possession_data->>'otp',
                p_possession_data->>'expiry',
                p_possession_data->>'security_level'
            ) THEN
                v_score := v_score + 75;
            END IF;
        WHEN p_possession_type = 'EMAIL' THEN
            IF possession.verify_email(
                p_possession_data->>'email',
                p_possession_data->>'otp',
                p_possession_data->>'expiry',
                p_possession_data->>'security_level'
            ) THEN
                v_score := v_score + 75;
            END IF;
        WHEN p_possession_type = 'TOKEN_PHYSICAL' THEN
            IF possession.verify_physical_token(
                p_possession_data->>'token_id',
                p_possession_data->>'token_type',
                p_possession_data->>'security_level',
                p_possession_data->>'encryption_status'
            ) THEN
                v_score := v_score + 90;
            END IF;
        WHEN p_possession_type = 'SMART_CARD' THEN
            IF possession.verify_smart_card(
                p_possession_data->>'card_id',
                p_possession_data->>'card_type',
                p_possession_data->>'security_level',
                p_possession_data->>'encryption_status'
            ) THEN
                v_score := v_score + 95;
            END IF;
        WHEN p_possession_type = 'FIDO2' THEN
            IF possession.verify_fido2(
                p_possession_data->>'device_id',
                p_possession_data->>'device_type',
                p_possession_data->>'security_level',
                p_possession_data->>'encryption_status'
            ) THEN
                v_score := v_score + 95;
            END IF;
        WHEN p_possession_type = 'PUSH' THEN
            IF possession.verify_push(
                p_possession_data->>'device_id',
                p_possession_data->>'app_id',
                p_possession_data->>'security_level',
                p_possession_data->>'encryption_status'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_possession_type = 'CERTIFICATE' THEN
            IF possession.verify_certificate(
                p_possession_data->>'cert_id',
                p_possession_data->>'cert_type',
                p_possession_data->>'security_level',
                p_possession_data->>'encryption_status'
            ) THEN
                v_score := v_score + 90;
            END IF;
        WHEN p_possession_type = 'BLUETOOTH' THEN
            IF possession.verify_bluetooth(
                p_possession_data->>'device_id',
                p_possession_data->>'device_type',
                p_possession_data->>'security_level',
                p_possession_data->>'encryption_status'
            ) THEN
                v_score := v_score + 80;
            END IF;
        WHEN p_possession_type = 'NFC' THEN
            IF possession.verify_nfc(
                p_possession_data->>'device_id',
                p_possession_data->>'device_type',
                p_possession_data->>'security_level',
                p_possession_data->>'encryption_status'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_possession_type = 'QR_CODE' THEN
            IF possession.verify_qr_code(
                p_possession_data->>'code_id',
                p_possession_data->>'code_type',
                p_possession_data->>'security_level',
                p_possession_data->>'encryption_status'
            ) THEN
                v_score := v_score + 80;
            END IF;
        WHEN p_possession_type = 'TOKEN_VIRTUAL' THEN
            IF possession.verify_virtual_token(
                p_possession_data->>'token_id',
                p_possession_data->>'token_type',
                p_possession_data->>'security_level',
                p_possession_data->>'encryption_status'
            ) THEN
                v_score := v_score + 90;
            END IF;
        WHEN p_possession_type = 'SECURE_ELEMENT' THEN
            IF possession.verify_secure_element(
                p_possession_data->>'element_id',
                p_possession_data->>'element_type',
                p_possession_data->>'security_level',
                p_possession_data->>'encryption_status'
            ) THEN
                v_score := v_score + 95;
            END IF;
        WHEN p_possession_type = 'OTP_CARD' THEN
            IF possession.verify_otp_card(
                p_possession_data->>'card_id',
                p_possession_data->>'card_type',
                p_possession_data->>'security_level',
                p_possession_data->>'encryption_status'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_possession_type = 'PROXIMITY' THEN
            IF possession.verify_proximity(
                p_possession_data->>'device_id',
                p_possession_data->>'device_type',
                p_possession_data->>'security_level',
                p_possession_data->>'encryption_status'
            ) THEN
                v_score := v_score + 80;
            END IF;
        WHEN p_possession_type = 'RADIO' THEN
            IF possession.verify_radio(
                p_possession_data->>'device_id',
                p_possession_data->>'device_type',
                p_possession_data->>'security_level',
                p_possession_data->>'encryption_status'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_possession_type = 'SIM' THEN
            IF possession.verify_sim(
                p_possession_data->>'sim_id',
                p_possession_data->>'sim_type',
                p_possession_data->>'security_level',
                p_possession_data->>'encryption_status'
            ) THEN
                v_score := v_score + 80;
            END IF;
    END CASE;

    -- Verificar nível de autenticação
    IF v_score >= p_required_level THEN
        -- Registrar tentativa de autenticação
        INSERT INTO iam_access_control.possession_auth_attempts (
            auth_id,
            possession_type,
            possession_data,
            required_level,
            score_risco,
            created_at
        ) VALUES (
            v_auth_id,
            p_possession_type,
            p_possession_data,
            p_required_level,
            v_score,
            current_timestamp
        );

        -- Gerar recomendações
        v_recommendations := jsonb_build_object(
            'action', 'allow',
            'reason', 'Possession authentication successful',
            'score', v_score,
            'details', jsonb_build_object(
                'possession_type', p_possession_type,
                'required_level', p_required_level,
                'possession_data', p_possession_data
            )
        );

        -- Registrar evento de segurança
        INSERT INTO iam_access_control.security_events (
            event_type,
            event_time,
            event_details,
            score_risco
        ) VALUES (
            'POSSESSION_AUTH_SUCCESS',
            current_timestamp,
            jsonb_build_object(
                'auth_id', v_auth_id,
                'possession_type', p_possession_type,
                'recommendations', v_recommendations
            ),
            v_score
        );

        -- Retornar resultado
        RETURN json_build_object(
            'success', true,
            'auth_id', v_auth_id,
            'score', v_score,
            'recommendations', v_recommendations
        );
    ELSE
        -- Gerar recomendações para falha
        v_recommendations := jsonb_build_object(
            'action', 'block',
            'reason', 'Insufficient possession authentication score',
            'score', v_score,
            'details', jsonb_build_object(
                'possession_type', p_possession_type,
                'required_level', p_required_level,
                'possession_data', p_possession_data
            )
        );

        -- Registrar evento de segurança
        INSERT INTO iam_access_control.security_events (
            event_type,
            event_time,
            event_details,
            score_risco
        ) VALUES (
            'POSSESSION_AUTH_FAILURE',
            current_timestamp,
            jsonb_build_object(
                'auth_id', v_auth_id,
                'possession_type', p_possession_type,
                'recommendations', v_recommendations
            ),
            v_score
        );

        -- Retornar resultado
        RETURN json_build_object(
            'success', false,
            'auth_id', v_auth_id,
            'score', v_score,
            'recommendations', v_recommendations
        );
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
CREATE OR REPLACE FUNCTION iam_access_control.generate_protocol_token(
    p_protocol_type TEXT,
    p_protocol_data JSONB,
    p_expiration_minutes INT DEFAULT 30
)
RETURNS JSON AS $$
DECLARE
    v_token_id UUID;
    v_protocol_token TEXT;
    v_result JSON;
BEGIN
    -- Gerar ID único para token
    v_token_id := gen_random_uuid();

    -- Gerar token baseado no tipo de protocolo
    CASE 
        WHEN p_protocol_type = 'OAUTH2' THEN
            v_protocol_token := encode(gen_random_bytes(32), 'hex');
        WHEN p_protocol_type = 'OIDC' THEN
            v_protocol_token := encode(gen_random_bytes(48), 'hex');
        WHEN p_protocol_type = 'SAML' THEN
            v_protocol_token := encode(gen_random_bytes(64), 'base64');
        WHEN p_protocol_type = 'WS-FED' THEN
            v_protocol_token := encode(gen_random_bytes(80), 'base64');
        WHEN p_protocol_type = 'KERBEROS' THEN
            v_protocol_token := encode(gen_random_bytes(96), 'base64');
        WHEN p_protocol_type = 'LDAP' THEN
            v_protocol_token := encode(gen_random_bytes(48), 'hex');
        WHEN p_protocol_type = 'RADIUS' THEN
            v_protocol_token := encode(gen_random_bytes(32), 'hex');
        WHEN p_protocol_type = 'TACACS' THEN
            v_protocol_token := encode(gen_random_bytes(24), 'hex');
        WHEN p_protocol_type = 'DIAMETER' THEN
            v_protocol_token := encode(gen_random_bytes(64), 'hex');
        WHEN p_protocol_type = 'DIAMETER++' THEN
            v_protocol_token := encode(gen_random_bytes(80), 'hex');
        WHEN p_protocol_type = 'DIAMETER+++' THEN
            v_protocol_token := encode(gen_random_bytes(96), 'hex');
        WHEN p_protocol_type = 'DIAMETER++++' THEN
            v_protocol_token := encode(gen_random_bytes(112), 'hex');
        WHEN p_protocol_type = 'DIAMETER+++++' THEN
            v_protocol_token := encode(gen_random_bytes(128), 'hex');
        WHEN p_protocol_type = 'DIAMETER++++++' THEN
            v_protocol_token := encode(gen_random_bytes(144), 'hex');
        WHEN p_protocol_type = 'DIAMETER+++++++' THEN
            v_protocol_token := encode(gen_random_bytes(160), 'hex');
        WHEN p_protocol_type = 'DIAMETER++++++++' THEN
            v_protocol_token := encode(gen_random_bytes(176), 'hex');
        WHEN p_protocol_type = 'DIAMETER+++++++++' THEN
            v_protocol_token := encode(gen_random_bytes(192), 'hex');
        WHEN p_protocol_type = 'DIAMETER++++++++++' THEN
            v_protocol_token := encode(gen_random_bytes(208), 'hex');
        WHEN p_protocol_type = 'DIAMETER+++++++++++' THEN
            v_protocol_token := encode(gen_random_bytes(224), 'hex');
        WHEN p_protocol_type = 'DIAMETER++++++++++++' THEN
            v_protocol_token := encode(gen_random_bytes(240), 'hex');
    END CASE;

    -- Registrar token
    INSERT INTO iam_access_control.protocol_tokens (
        token_id,
        protocol_type,
        protocol_data,
        protocol_token,
        created_at,
        expires_at
    ) VALUES (
        v_token_id,
        p_protocol_type,
        p_protocol_data,
        v_protocol_token,
        current_timestamp,
        current_timestamp + INTERVAL '1 minute' * p_expiration_minutes
    );

    -- Preparar resultado
    v_result := json_build_object(
        'success', TRUE,
        'token_id', v_token_id,
        'protocol_type', p_protocol_type,
        'protocol_token', v_protocol_token,
        'expires_at', current_timestamp + INTERVAL '1 minute' * p_expiration_minutes
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.31 Autenticação Baseada em Conhecimento
CREATE OR REPLACE FUNCTION iam_access_control.knowledge_authentication(
    p_knowledge_type TEXT, -- 'PASSWORD', 'PIN', 'PATTERN', 'QUESTION', 'TOKEN', 'PASSPHRASE', 'IMAGE', 'SINGLE_USE', 'NO_CONNECTION', 'GESTURE', 'SEQUENCE', 'ACTION', 'EVENT', 'STATE', 'CONDITION', 'POLICY', 'COMPLIANCE', 'AUDIT', 'MONITORING', 'TRACE', 'SECURITY'
    p_knowledge_data JSONB,
    p_required_level INT
)
RETURNS JSON AS $$
DECLARE
    v_auth_id UUID;
    v_score INT := 0;
    v_recommendations JSONB;
    v_result JSON;
BEGIN
    -- Gerar ID único para autenticação
    v_auth_id := gen_random_uuid();

    -- Verificar tipo de conhecimento e aplicar score
    CASE 
        WHEN p_knowledge_type = 'PASSWORD' THEN
            IF knowledge.verify_password(
                p_knowledge_data->>'password',
                p_knowledge_data->>'requirements',
                p_knowledge_data->>'complexity',
                p_knowledge_data->>'history'
            ) THEN
                v_score := v_score + 80;
            END IF;
        WHEN p_knowledge_type = 'PIN' THEN
            IF knowledge.verify_pin(
                p_knowledge_data->>'pin',
                p_knowledge_data->>'length',
                p_knowledge_data->>'type',
                p_knowledge_data->>'history'
            ) THEN
                v_score := v_score + 75;
            END IF;
        WHEN p_knowledge_type = 'PATTERN' THEN
            IF knowledge.verify_pattern(
                p_knowledge_data->>'pattern',
                p_knowledge_data->>'complexity',
                p_knowledge_data->>'type',
                p_knowledge_data->>'history'
            ) THEN
                v_score := v_score + 80;
            END IF;
        WHEN p_knowledge_type = 'QUESTION' THEN
            IF knowledge.verify_question(
                p_knowledge_data->>'question',
                p_knowledge_data->>'answer',
                p_knowledge_data->>'security_level',
                p_knowledge_data->>'history'
            ) THEN
                v_score := v_score + 75;
            END IF;
        WHEN p_knowledge_type = 'TOKEN' THEN
            IF knowledge.verify_token(
                p_knowledge_data->>'token',
                p_knowledge_data->>'type',
                p_knowledge_data->>'expiry',
                p_knowledge_data->>'security_level'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_knowledge_type = 'PASSPHRASE' THEN
            IF knowledge.verify_passphrase(
                p_knowledge_data->>'passphrase',
                p_knowledge_data->>'requirements',
                p_knowledge_data->>'complexity',
                p_knowledge_data->>'history'
            ) THEN
                v_score := v_score + 90;
            END IF;
        WHEN p_knowledge_type = 'IMAGE' THEN
            IF knowledge.verify_image(
                p_knowledge_data->>'image_id',
                p_knowledge_data->>'coordinates',
                p_knowledge_data->>'security_level',
                p_knowledge_data->>'history'
            ) THEN
                v_score := v_score + 80;
            END IF;
        WHEN p_knowledge_type = 'SINGLE_USE' THEN
            IF knowledge.verify_single_use(
                p_knowledge_data->>'token',
                p_knowledge_data->>'expiry',
                p_knowledge_data->>'security_level',
                p_knowledge_data->>'history'
            ) THEN
                v_score := v_score + 90;
            END IF;
        WHEN p_knowledge_type = 'NO_CONNECTION' THEN
            IF knowledge.verify_no_connection(
                p_knowledge_data->>'token',
                p_knowledge_data->>'expiry',
                p_knowledge_data->>'security_level',
                p_knowledge_data->>'history'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_knowledge_type = 'GESTURE' THEN
            IF knowledge.verify_gesture(
                p_knowledge_data->>'gesture',
                p_knowledge_data->>'complexity',
                p_knowledge_data->>'type',
                p_knowledge_data->>'history'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_knowledge_type = 'SEQUENCE' THEN
            IF knowledge.verify_sequence(
                p_knowledge_data->>'sequence',
                p_knowledge_data->>'complexity',
                p_knowledge_data->>'type',
                p_knowledge_data->>'history'
            ) THEN
                v_score := v_score + 90;
            END IF;
        WHEN p_knowledge_type = 'ACTION' THEN
            IF knowledge.verify_action(
                p_knowledge_data->>'action',
                p_knowledge_data->>'sequence',
                p_knowledge_data->>'security_level',
                p_knowledge_data->>'history'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_knowledge_type = 'EVENT' THEN
            IF knowledge.verify_event(
                p_knowledge_data->>'event',
                p_knowledge_data->>'sequence',
                p_knowledge_data->>'security_level',
                p_knowledge_data->>'history'
            ) THEN
                v_score := v_score + 90;
            END IF;
        WHEN p_knowledge_type = 'STATE' THEN
            IF knowledge.verify_state(
                p_knowledge_data->>'state',
                p_knowledge_data->>'sequence',
                p_knowledge_data->>'security_level',
                p_knowledge_data->>'history'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_knowledge_type = 'CONDITION' THEN
            IF knowledge.verify_condition(
                p_knowledge_data->>'condition',
                p_knowledge_data->>'sequence',
                p_knowledge_data->>'security_level',
                p_knowledge_data->>'history'
            ) THEN
                v_score := v_score + 80;
            END IF;
        WHEN p_knowledge_type = 'POLICY' THEN
            IF knowledge.verify_policy(
                p_knowledge_data->>'policy',
                p_knowledge_data->>'sequence',
                p_knowledge_data->>'security_level',
                p_knowledge_data->>'history'
            ) THEN
                v_score := v_score + 95;
            END IF;
        WHEN p_knowledge_type = 'COMPLIANCE' THEN
            IF knowledge.verify_compliance(
                p_knowledge_data->>'compliance',
                p_knowledge_data->>'sequence',
                p_knowledge_data->>'security_level',
                p_knowledge_data->>'history'
            ) THEN
                v_score := v_score + 90;
            END IF;
        WHEN p_knowledge_type = 'AUDIT' THEN
            IF knowledge.verify_audit(
                p_knowledge_data->>'audit',
                p_knowledge_data->>'sequence',
                p_knowledge_data->>'security_level',
                p_knowledge_data->>'history'
            ) THEN
                v_score := v_score + 95;
            END IF;
        WHEN p_knowledge_type = 'MONITORING' THEN
            IF knowledge.verify_monitoring(
                p_knowledge_data->>'monitoring',
                p_knowledge_data->>'sequence',
                p_knowledge_data->>'security_level',
                p_knowledge_data->>'history'
            ) THEN
                v_score := v_score + 90;
            END IF;
        WHEN p_knowledge_type = 'TRACE' THEN
            IF knowledge.verify_trace(
                p_knowledge_data->>'trace',
                p_knowledge_data->>'sequence',
                p_knowledge_data->>'security_level',
                p_knowledge_data->>'history'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_knowledge_type = 'SECURITY' THEN
            IF knowledge.verify_security(
                p_knowledge_data->>'security',
                p_knowledge_data->>'sequence',
                p_knowledge_data->>'security_level',
                p_knowledge_data->>'history'
            ) THEN
                v_score := v_score + 100;
            END IF;
    END CASE;

    -- Verificar nível de autenticação
    IF v_score >= p_required_level THEN
        -- Registrar tentativa de autenticação
        INSERT INTO iam_access_control.knowledge_auth_attempts (
            auth_id,
            knowledge_type,
            knowledge_data,
            required_level,
            score_risco,
            created_at
        ) VALUES (
            v_auth_id,
            p_knowledge_type,
            p_knowledge_data,
            p_required_level,
            v_score,
            current_timestamp
        );

        -- Gerar recomendações
        v_recommendations := jsonb_build_object(
            'action', 'allow',
            'reason', 'Knowledge authentication successful',
            'score', v_score,
            'details', jsonb_build_object(
                'knowledge_type', p_knowledge_type,
                'required_level', p_required_level,
                'knowledge_data', p_knowledge_data
            )
        );

        -- Registrar evento de segurança
        INSERT INTO iam_access_control.security_events (
            event_type,
            event_time,
            event_details,
            score_risco
        ) VALUES (
            'KNOWLEDGE_AUTH_SUCCESS',
            current_timestamp,
            jsonb_build_object(
                'auth_id', v_auth_id,
                'knowledge_type', p_knowledge_type,
                'recommendations', v_recommendations
            ),
            v_score
        );
    ELSE
        -- Gerar recomendações para falha
        v_recommendations := jsonb_build_object(
            'action', 'deny',
            'reason', 'Insufficient authentication level',
            'score', v_score,
            'details', jsonb_build_object(
                'knowledge_type', p_knowledge_type,
                'required_level', p_required_level,
                'knowledge_data', p_knowledge_data
            )
        );

        -- Registrar evento de segurança
        INSERT INTO iam_access_control.security_events (
            event_type,
            event_time,
            event_details,
            score_risco
        ) VALUES (
            'KNOWLEDGE_AUTH_FAILED',
            current_timestamp,
            jsonb_build_object(
                'auth_id', v_auth_id,
                'knowledge_type', p_knowledge_type,
                'recommendations', v_recommendations
            ),
            v_score
        );
    END IF;

    -- Preparar resultado
    v_result := json_build_object(
        'success', v_score >= p_required_level,
        'auth_id', v_auth_id,
        'knowledge_type', p_knowledge_type,
        'required_level', p_required_level,
        'score_risco', v_score,
        'recommendations', v_recommendations,
        'details', json_build_object(
            'knowledge_data', p_knowledge_data
        )
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.32 Configurar Nível de Autenticação por Conhecimento
CREATE OR REPLACE FUNCTION iam_access_control.configure_knowledge_level(
    p_service_id TEXT,
    p_knowledge_type TEXT,
    p_required_level INT,
    p_knowledge_rules JSONB
)
RETURNS JSON AS $$
DECLARE
    v_config_id UUID;
    v_result JSON;
BEGIN
    -- Gerar ID único para configuração
    v_config_id := gen_random_uuid();

    -- Registrar configuração
    INSERT INTO iam_access_control.knowledge_levels (
        config_id,
        service_id,
        knowledge_type,
        required_level,
        knowledge_rules,
        created_at,
        updated_at
    ) VALUES (
        v_config_id,
        p_service_id,
        p_knowledge_type,
        p_required_level,
        p_knowledge_rules,
        current_timestamp,
        current_timestamp
    ) ON CONFLICT (service_id, knowledge_type)
    DO UPDATE SET
        required_level = EXCLUDED.required_level,
        knowledge_rules = EXCLUDED.knowledge_rules,
        updated_at = current_timestamp;

    -- Preparar resultado
    v_result := json_build_object(
        'success', TRUE,
        'config_id', v_config_id,
        'service_id', p_service_id,
        'knowledge_type', p_knowledge_type,
        'required_level', p_required_level,
        'knowledge_rules', p_knowledge_rules
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.31 Autenticação Baseada em Conhecimento
CREATE OR REPLACE FUNCTION iam_access_control.device_authentication(
    p_device_type TEXT, -- 'MOBILE', 'DESKTOP', 'TABLET', 'WEARABLE', 'IOT', 'SERVER', 'NETWORK', 'EMBEDDED', 'VIRTUAL', 'CLOUD', 'HYBRID', 'EDGE', 'QUANTUM', 'AI', 'ROBOTIC', 'VR', 'AR', 'HMD', 'SMARTGLASS', 'SMARTWATCH', 'SMARTBAND', 'SMARTRING', 'SMARTCLOTH', 'SMARTHOME', 'SMARTCITY', 'SMARTCAR', 'SMARTDRONE', 'SMARTSHIP', 'SMARTPLANE', 'SMARTTRAIN', 'SMARTBUS', 'SMARTBIKE', 'SMARTSCOOTER', 'SMARTWHEEL', 'SMARTPROSTHETIC', 'SMARTIMPLANT', 'SMARTORGAN', 'SMARTBIO', 'SMARTNANO', 'SMARTMICRO', 'SMARTMACRO', 'SMARTGEO', 'SMARTASTRO', 'SMARTSPACE', 'SMARTMOON', 'SMARTMARS', 'SMARTJUPITER', 'SMARTSATURN', 'SMARTURANUS', 'SMARTNEPTUNE', 'SMARTPLUTO', 'SMARTCOMET', 'SMARTASTEROID', 'SMARTMETEOR', 'SMARTSTAR', 'SMARTGALAXY', 'SMARTUNIVERSE', 'SMARTMULTIVERSE'
    p_device_id TEXT,
    p_device_info JSONB,
    p_required_level INT
)
RETURNS JSON AS $$
DECLARE
    v_auth_id UUID;
    v_score INT := 0;
    v_recommendations JSONB;
    v_result JSON;
BEGIN
    -- Gerar ID único para autenticação
    v_auth_id := gen_random_uuid();

    -- Verificar tipo de dispositivo e aplicar score
    CASE 
        WHEN p_device_type IN ('MOBILE', 'TABLET') THEN
            IF device.verify_mobile(
                p_device_id,
                p_device_info->>'os_version',
                p_device_info->>'device_model',
                p_device_info->>'security_patch',
                p_device_info->>'encryption_status'
            ) THEN
                v_score := v_score + 80;
            END IF;
        WHEN p_device_type = 'DESKTOP' THEN
            IF device.verify_desktop(
                p_device_id,
                p_device_info->>'os_version',
                p_device_info->>'antivirus_status',
                p_device_info->>'firewall_status',
                p_device_info->>'encryption_status'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_device_type IN ('WEARABLE', 'SMARTWATCH', 'SMARTBAND', 'SMARTRING', 'SMARTGLASS') THEN
            IF device.verify_wearable(
                p_device_id,
                p_device_info->>'device_type',
                p_device_info->>'os_version',
                p_device_info->>'security_status',
                p_device_info->>'encryption_status'
            ) THEN
                v_score := v_score + 75;
            END IF;
        WHEN p_device_type IN ('IOT', 'SMARTHOME', 'SMARTCITY', 'SMARTCAR', 'SMARTDRONE') THEN
            IF device.verify_iot(
                p_device_id,
                p_device_info->>'device_type',
                p_device_info->>'security_protocol',
                p_device_info->>'encryption_status',
                p_device_info->>'authentication_status'
            ) THEN
                v_score := v_score + 70;
            END IF;
        WHEN p_device_type IN ('SERVER', 'NETWORK', 'VIRTUAL', 'CLOUD', 'HYBRID', 'EDGE') THEN
            IF device.verify_server(
                p_device_id,
                p_device_info->>'os_version',
                p_device_info->>'security_patch',
                p_device_info->>'encryption_status',
                p_device_info->>'firewall_status'
            ) THEN
                v_score := v_score + 90;
            END IF;
        WHEN p_device_type IN ('QUANTUM', 'AI', 'ROBOTIC', 'VR', 'AR', 'HMD') THEN
            IF device.verify_advanced(
                p_device_id,
                p_device_info->>'device_type',
                p_device_info->>'security_protocol',
                p_device_info->>'encryption_status',
                p_device_info->>'authentication_status'
            ) THEN
                v_score := v_score + 95;
            END IF;
        WHEN p_device_type IN ('SMARTBIKE', 'SMARTSCOOTER', 'SMARTWHEEL', 'SMARTPROSTHETIC', 'SMARTIMPLANT', 'SMARTORGAN', 'SMARTBIO') THEN
            IF device.verify_bio(
                p_device_id,
                p_device_info->>'device_type',
                p_device_info->>'security_status',
                p_device_info->>'encryption_status',
                p_device_info->>'authentication_status'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_device_type IN ('SMARTNANO', 'SMARTMICRO', 'SMARTMACRO', 'SMARTGEO', 'SMARTASTRO', 'SMARTSPACE') THEN
            IF device.verify_space(
                p_device_id,
                p_device_info->>'device_type',
                p_device_info->>'security_protocol',
                p_device_info->>'encryption_status',
                p_device_info->>'authentication_status'
            ) THEN
                v_score := v_score + 90;
            END IF;
        WHEN p_device_type IN ('SMARTMOON', 'SMARTMARS', 'SMARTJUPITER', 'SMARTSATURN', 'SMARTURANUS', 'SMARTNEPTUNE', 'SMARTPLUTO', 'SMARTCOMET', 'SMARTASTEROID', 'SMARTMETEOR', 'SMARTSTAR', 'SMARTGALAXY', 'SMARTUNIVERSE', 'SMARTMULTIVERSE') THEN
            IF device.verify_interstellar(
                p_device_id,
                p_device_info->>'device_type',
                p_device_info->>'security_protocol',
                p_device_info->>'encryption_status',
                p_device_info->>'authentication_status'
            ) THEN
                v_score := v_score + 100;
            END IF;
    END CASE;

    -- Verificar nível de autenticação
    IF v_score >= p_required_level THEN
        -- Registrar tentativa de autenticação
        INSERT INTO iam_access_control.device_auth_attempts (
            auth_id,
            device_type,
            device_id,
            device_info,
            required_level,
            score_risco,
            created_at
        ) VALUES (
            v_auth_id,
            p_device_type,
            p_device_id,
            p_device_info,
            p_required_level,
            v_score,
            current_timestamp
        );

        -- Gerar recomendações
        v_recommendations := jsonb_build_object(
            'action', 'allow',
            'reason', 'Device authentication successful',
            'score', v_score,
            'details', jsonb_build_object(
                'device_type', p_device_type,
                'required_level', p_required_level,
                'device_info', p_device_info
            )
        );

        -- Registrar evento de segurança
        INSERT INTO iam_access_control.security_events (
            event_type,
            event_time,
            event_details,
            score_risco
        ) VALUES (
            'DEVICE_AUTH_SUCCESS',
            current_timestamp,
            jsonb_build_object(
                'auth_id', v_auth_id,
                'device_type', p_device_type,
                'recommendations', v_recommendations
            ),
            v_score
        );
    ELSE
        -- Gerar recomendações para falha
        v_recommendations := jsonb_build_object(
            'action', 'deny',
            'reason', 'Insufficient authentication level',
            'score', v_score,
            'details', jsonb_build_object(
                'device_type', p_device_type,
                'required_level', p_required_level,
                'device_info', p_device_info
            )
        );

        -- Registrar evento de segurança
        INSERT INTO iam_access_control.security_events (
            event_type,
            event_time,
            event_details,
            score_risco
        ) VALUES (
            'DEVICE_AUTH_FAILED',
            current_timestamp,
            jsonb_build_object(
                'auth_id', v_auth_id,
                'device_type', p_device_type,
                'recommendations', v_recommendations
            ),
            v_score
        );
    END IF;

    -- Preparar resultado
    v_result := json_build_object(
        'success', v_score >= p_required_level,
        'auth_id', v_auth_id,
        'device_type', p_device_type,
        'required_level', p_required_level,
        'score_risco', v_score,
        'recommendations', v_recommendations,
        'details', json_build_object(
            'device_id', p_device_id,
            'device_info', p_device_info
        )
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.30 Configurar Nível de Autenticação por Dispositivo
CREATE OR REPLACE FUNCTION iam_access_control.configure_device_level(
    p_service_id TEXT,
    p_device_type TEXT,
    p_required_level INT,
    p_device_rules JSONB
)
RETURNS JSON AS $$
DECLARE
    v_config_id UUID;
    v_result JSON;
BEGIN
    -- Gerar ID único para configuração
    v_config_id := gen_random_uuid();

    -- Registrar configuração
    INSERT INTO iam_access_control.device_levels (
        config_id,
        service_id,
        device_type,
        required_level,
        device_rules,
        created_at,
        updated_at
    ) VALUES (
        v_config_id,
        p_service_id,
        p_device_type,
        p_required_level,
        p_device_rules,
        current_timestamp,
        current_timestamp
    ) ON CONFLICT (service_id, device_type)
    DO UPDATE SET
        required_level = EXCLUDED.required_level,
        device_rules = EXCLUDED.device_rules,
        updated_at = current_timestamp;

    -- Preparar resultado
    v_result := json_build_object(
        'success', TRUE,
        'config_id', v_config_id,
        'service_id', p_service_id,
        'device_type', p_device_type,
        'required_level', p_required_level,
        'device_rules', p_device_rules
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.27 Autenticação Contextual
CREATE OR REPLACE FUNCTION iam_access_control.contextual_authentication(
    p_context_type TEXT, -- 'LOCATION', 'DEVICE', 'TIME', 'BEHAVIOR', 'RISK', 'ENVIRONMENT', 'NETWORK', 'APPLICATION', 'SERVICE', 'TRANSACTION', 'ACTION', 'EVENT', 'STATE', 'CONDITION', 'POLICY', 'COMPLIANCE', 'AUDIT', 'MONITORING', 'TRACE', 'SECURITY'
    p_context_data JSONB,
    p_required_level INT
)
RETURNS JSON AS $$
DECLARE
    v_auth_id UUID;
    v_score INT := 0;
    v_recommendations JSONB;
    v_result JSON;
BEGIN
    -- Gerar ID único para autenticação
    v_auth_id := gen_random_uuid();

    -- Verificar tipo de contexto e aplicar score
    CASE 
        WHEN p_context_type = 'LOCATION' THEN
            IF context.verify_location(
                p_context_data->>'latitude',
                p_context_data->>'longitude',
                p_context_data->>'allowed_regions'
            ) THEN
                v_score := v_score + 75;
            END IF;
        WHEN p_context_type = 'DEVICE' THEN
            IF context.verify_device(
                p_context_data->>'device_id',
                p_context_data->>'device_type',
                p_context_data->>'device_os',
                p_context_data->>'device_model'
            ) THEN
                v_score := v_score + 80;
            END IF;
        WHEN p_context_type = 'TIME' THEN
            IF context.verify_time(
                p_context_data->>'current_time',
                p_context_data->>'allowed_hours',
                p_context_data->>'allowed_days'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_context_type = 'BEHAVIOR' THEN
            IF context.verify_behavior(
                p_context_data->>'user_id',
                p_context_data->>'action_type',
                p_context_data->>'action_time',
                p_context_data->>'action_frequency'
            ) THEN
                v_score := v_score + 90;
            END IF;
        WHEN p_context_type = 'RISK' THEN
            IF context.verify_risk(
                p_context_data->>'risk_score',
                p_context_data->>'risk_threshold',
                p_context_data->>'risk_factors'
            ) THEN
                v_score := v_score + 95;
            END IF;
        WHEN p_context_type = 'ENVIRONMENT' THEN
            IF context.verify_environment(
                p_context_data->>'environment_type',
                p_context_data->>'environment_status',
                p_context_data->>'environment_security'
            ) THEN
                v_score := v_score + 80;
            END IF;
        WHEN p_context_type = 'NETWORK' THEN
            IF context.verify_network(
                p_context_data->>'ip_address',
                p_context_data->>'network_segment',
                p_context_data->>'network_security'
            ) THEN
                v_score := v_score + 75;
            END IF;
        WHEN p_context_type = 'APPLICATION' THEN
            IF context.verify_application(
                p_context_data->>'app_id',
                p_context_data->>'app_version',
                p_context_data->>'app_security'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_context_type = 'SERVICE' THEN
            IF context.verify_service(
                p_context_data->>'service_id',
                p_context_data->>'service_status',
                p_context_data->>'service_security'
            ) THEN
                v_score := v_score + 80;
            END IF;
        WHEN p_context_type = 'TRANSACTION' THEN
            IF context.verify_transaction(
                p_context_data->>'transaction_id',
                p_context_data->>'transaction_amount',
                p_context_data->>'transaction_risk'
            ) THEN
                v_score := v_score + 90;
            END IF;
        WHEN p_context_type = 'ACTION' THEN
            IF context.verify_action(
                p_context_data->>'action_id',
                p_context_data->>'action_type',
                p_context_data->>'action_risk'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_context_type = 'EVENT' THEN
            IF context.verify_event(
                p_context_data->>'event_id',
                p_context_data->>'event_type',
                p_context_data->>'event_risk'
            ) THEN
                v_score := v_score + 90;
            END IF;
        WHEN p_context_type = 'STATE' THEN
            IF context.verify_state(
                p_context_data->>'state_id',
                p_context_data->>'state_type',
                p_context_data->>'state_risk'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_context_type = 'CONDITION' THEN
            IF context.verify_condition(
                p_context_data->>'condition_id',
                p_context_data->>'condition_type',
                p_context_data->>'condition_risk'
            ) THEN
                v_score := v_score + 80;
            END IF;
        WHEN p_context_type = 'POLICY' THEN
            IF context.verify_policy(
                p_context_data->>'policy_id',
                p_context_data->>'policy_type',
                p_context_data->>'policy_risk'
            ) THEN
                v_score := v_score + 95;
            END IF;
        WHEN p_context_type = 'COMPLIANCE' THEN
            IF context.verify_compliance(
                p_context_data->>'compliance_id',
                p_context_data->>'compliance_type',
                p_context_data->>'compliance_risk'
            ) THEN
                v_score := v_score + 90;
            END IF;
        WHEN p_context_type = 'AUDIT' THEN
            IF context.verify_audit(
                p_context_data->>'audit_id',
                p_context_data->>'audit_type',
                p_context_data->>'audit_risk'
            ) THEN
                v_score := v_score + 95;
            END IF;
        WHEN p_context_type = 'MONITORING' THEN
            IF context.verify_monitoring(
                p_context_data->>'monitor_id',
                p_context_data->>'monitor_type',
                p_context_data->>'monitor_risk'
            ) THEN
                v_score := v_score + 90;
            END IF;
        WHEN p_context_type = 'TRACE' THEN
            IF context.verify_trace(
                p_context_data->>'trace_id',
                p_context_data->>'trace_type',
                p_context_data->>'trace_risk'
            ) THEN
                v_score := v_score + 85;
            END IF;
        WHEN p_context_type = 'SECURITY' THEN
            IF context.verify_security(
                p_context_data->>'security_id',
                p_context_data->>'security_type',
                p_context_data->>'security_risk'
            ) THEN
                v_score := v_score + 100;
            END IF;
    END CASE;

    -- Verificar nível de autenticação
    IF v_score >= p_required_level THEN
        -- Registrar tentativa de autenticação
        INSERT INTO iam_access_control.context_auth_attempts (
            auth_id,
            context_type,
            context_data,
            required_level,
            score_risco,
            created_at
        ) VALUES (
            v_auth_id,
            p_context_type,
            p_context_data,
            p_required_level,
            v_score,
            current_timestamp
        );

        -- Gerar recomendações
        v_recommendations := jsonb_build_object(
            'action', 'allow',
            'reason', 'Contextual authentication successful',
            'score', v_score,
            'details', jsonb_build_object(
                'context_type', p_context_type,
                'required_level', p_required_level,
                'context_data', p_context_data
            )
        );

        -- Registrar evento de segurança
        INSERT INTO iam_access_control.security_events (
            event_type,
            event_time,
            event_details,
            score_risco
        ) VALUES (
            'CONTEXT_AUTH_SUCCESS',
            current_timestamp,
            jsonb_build_object(
                'auth_id', v_auth_id,
                'context_type', p_context_type,
                'recommendations', v_recommendations
            ),
            v_score
        );
    ELSE
        -- Gerar recomendações para falha
        v_recommendations := jsonb_build_object(
            'action', 'deny',
            'reason', 'Insufficient authentication level',
            'score', v_score,
            'details', jsonb_build_object(
                'context_type', p_context_type,
                'required_level', p_required_level,
                'context_data', p_context_data
            )
        );

        -- Registrar evento de segurança
        INSERT INTO iam_access_control.security_events (
            event_type,
            event_time,
            event_details,
            score_risco
        ) VALUES (
            'CONTEXT_AUTH_FAILED',
            current_timestamp,
            jsonb_build_object(
                'auth_id', v_auth_id,
                'context_type', p_context_type,
                'recommendations', v_recommendations
            ),
            v_score
        );
    END IF;

    -- Preparar resultado
    v_result := json_build_object(
        'success', v_score >= p_required_level,
        'auth_id', v_auth_id,
        'context_type', p_context_type,
        'required_level', p_required_level,
        'score_risco', v_score,
        'recommendations', v_recommendations,
        'details', json_build_object(
            'context_data', p_context_data
        )
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.28 Configurar Nível de Autenticação Contextual
CREATE OR REPLACE FUNCTION iam_access_control.configure_context_level(
    p_service_id TEXT,
    p_context_type TEXT,
    p_required_level INT,
    p_context_rules JSONB
)
RETURNS JSON AS $$
DECLARE
    v_config_id UUID;
    v_result JSON;
BEGIN
    -- Gerar ID único para configuração
    v_config_id := gen_random_uuid();

    -- Registrar configuração
    INSERT INTO iam_access_control.context_levels (
        config_id,
        service_id,
        context_type,
        required_level,
        context_rules,
        created_at,
        updated_at
    ) VALUES (
        v_config_id,
        p_service_id,
        p_context_type,
        p_required_level,
        p_context_rules,
        current_timestamp,
        current_timestamp
    ) ON CONFLICT (service_id, context_type)
    DO UPDATE SET
        required_level = EXCLUDED.required_level,
        context_rules = EXCLUDED.context_rules,
        updated_at = current_timestamp;

    -- Preparar resultado
    v_result := json_build_object(
        'success', TRUE,
        'config_id', v_config_id,
        'service_id', p_service_id,
        'context_type', p_context_type,
        'required_level', p_required_level,
        'context_rules', p_context_rules
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.1 Proteção contra XSS
CREATE OR REPLACE FUNCTION iam_access_control.prevent_xss(
    p_input TEXT,
    p_context TEXT -- 'html', 'attribute', 'javascript', 'style', 'url'
)
RETURNS JSON AS $$
DECLARE
    v_score INT := 0;
    v_is_suspicious BOOLEAN;
    v_result JSON;
    v_recommendations JSONB;
BEGIN
    -- Verificar caracteres especiais
    IF p_input ~* '[<>"\'&]' THEN
        v_score := v_score + 20;
    END IF;

    -- Verificar tags HTML
    IF p_input ~* '<[a-z][^>]*>' THEN
        v_score := v_score + 30;
    END IF;

    -- Verificar atributos perigosos
    IF p_input ~* '(on[a-z]+|style|script|src|href)=' THEN
        v_score := v_score + 40;
    END IF;

    -- Verificar payloads conhecidos
    IF EXISTS (
        SELECT 1 
        FROM iam_access_control.known_xss_patterns
        WHERE pattern ~* p_input
        AND active = TRUE
    ) THEN
        v_score := v_score + 50;
    END IF;

    -- Determinar se é suspeito
    v_is_suspicious := v_score >= 50;

    -- Gerar recomendações
    v_recommendations := CASE 
        WHEN v_score >= 90 THEN jsonb_build_array(
            'Bloquear entrada imediatamente',
            'Notificar equipe de segurança',
            'Habilitar sanitização estrita',
            'Revisar logs de entrada',
            'Verificar contexto de uso'
        )
        WHEN v_score >= 70 THEN jsonb_build_array(
            'Sanitizar entrada',
            'Registrar tentativa',
            'Monitorar contexto',
            'Avisar equipe de segurança'
        )
        ELSE jsonb_build_array(
            'Monitorar entrada',
            'Registrar atividade',
            'Analisar contexto'
        )
    END;

    -- Registrar evento
    INSERT INTO iam_access_control.security_events (
        event_type,
        event_details,
        event_time,
        risk_score
    ) VALUES (
        CASE 
            WHEN v_is_suspicious THEN 'XSS_ATTEMPT'
            ELSE 'INPUT_VALIDATION'
        END,
        jsonb_build_object(
            'input', p_input,
            'context', p_context,
            'score', v_score,
            'detalhes', jsonb_build_object(
                'has_html_tags', p_input ~* '<[a-z][^>]*>',
                'has_special_chars', p_input ~* '[<>"\'&]',
                'has_dangerous_attrs', p_input ~* '(on[a-z]+|style|script|src|href)=',
                'known_payload', EXISTS (
                    SELECT 1 
                    FROM iam_access_control.known_xss_patterns
                    WHERE pattern ~* p_input
                    AND active = TRUE
                )
            )
        ),
        current_timestamp,
        v_score
    );

    v_result := json_build_object(
        'success', NOT v_is_suspicious,
        'message', CASE 
            WHEN v_is_suspicious THEN 'Tentativa de XSS detectada'
            ELSE 'Entrada válida'
        END,
        'score', v_score,
        'recommendations', v_recommendations
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.2 Proteção contra CSRF
CREATE OR REPLACE FUNCTION iam_access_control.prevent_csrf(
    p_token TEXT,
    p_session_id TEXT,
    p_request_time TIMESTAMP WITH TIME ZONE
)
RETURNS JSON AS $$
DECLARE
    v_token_valid BOOLEAN;
    v_session_valid BOOLEAN;
    v_score INT := 0;
    v_result JSON;
    v_recommendations JSONB;
BEGIN
    -- Verificar validade do token
    SELECT 
        token_valid,
        session_valid
    INTO 
        v_token_valid,
        v_session_valid
    FROM (
        SELECT 
            p_token = ANY(
                SELECT token 
                FROM iam_access_control.csrf_tokens 
                WHERE session_id = p_session_id 
                AND expires_at > current_timestamp
            ) AS token_valid,
            EXISTS (
                SELECT 1 
                FROM iam_access_control.sessions 
                WHERE session_id = p_session_id 
                AND expires_at > current_timestamp
            ) AS session_valid
    ) AS validation;

    -- Calcular score de risco
    IF NOT v_token_valid THEN
        v_score := v_score + 40;
    END IF;

    IF NOT v_session_valid THEN
        v_score := v_score + 30;
    END IF;

    -- Verificar padrões de ataque
    IF EXISTS (
        SELECT 1 
        FROM iam_access_control.security_events 
        WHERE event_type = 'CSRF_ATTEMPT' 
        AND event_time > current_timestamp - INTERVAL '1 minute' 
        AND event_details->>'session_id' = p_session_id
    ) THEN
        v_score := v_score + 20;
    END IF;

    -- Gerar recomendações
    v_recommendations := CASE 
        WHEN v_score >= 90 THEN jsonb_build_array(
            'Bloquear sessão imediatamente',
            'Notificar equipe de segurança',
            'Regerar tokens',
            'Revisar logs de sessão'
        )
        WHEN v_score >= 70 THEN jsonb_build_array(
            'Regerar token',
            'Monitorar sessão',
            'Avisar equipe de segurança'
        )
        ELSE jsonb_build_array(
            'Monitorar atividade',
            'Registrar tentativa',
            'Analisar padrões'
        )
    END;

    -- Registrar evento
    INSERT INTO iam_access_control.security_events (
        event_type,
        event_details,
        event_time,
        risk_score
    ) VALUES (
        CASE 
            WHEN NOT v_token_valid OR NOT v_session_valid THEN 'CSRF_ATTEMPT'
            ELSE 'SESSION_VALIDATION'
        END,
        jsonb_build_object(
            'token_valid', v_token_valid,
            'session_valid', v_session_valid,
            'score', v_score,
            'detalhes', jsonb_build_object(
                'token_expired', NOT v_token_valid,
                'session_expired', NOT v_session_valid,
                'recent_attempts', EXISTS (
                    SELECT 1 
                    FROM iam_access_control.security_events 
                    WHERE event_type = 'CSRF_ATTEMPT' 
                    AND event_time > current_timestamp - INTERVAL '1 minute' 
                    AND event_details->>'session_id' = p_session_id
                )
            )
        ),
        current_timestamp,
        v_score
    );

    v_result := json_build_object(
        'success', v_token_valid AND v_session_valid,
        'message', CASE 
            WHEN NOT v_token_valid OR NOT v_session_valid THEN 'Validação CSRF falhou'
            ELSE 'Validação CSRF bem-sucedida'
        END,
        'score', v_score,
        'recommendations', v_recommendations
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.3 Rate Limiting Avançado
CREATE OR REPLACE FUNCTION iam_access_control.advanced_rate_limit(
    p_ip_address TEXT,
    p_user_id TEXT,
    p_action_type TEXT,
    p_window_size INTERVAL
)
RETURNS JSON AS $$
DECLARE
    v_request_count INT;
    v_user_count INT;
    v_score INT := 0;
    v_result JSON;
    v_recommendations JSONB;
BEGIN
    -- Verificar número de requisições recentes
    SELECT COUNT(*) INTO v_request_count
    FROM iam_access_control.rate_limit_logs
    WHERE ip_address = p_ip_address
    AND action_type = p_action_type
    AND request_time > current_timestamp - p_window_size;

    -- Verificar número de usuários únicos
    SELECT COUNT(DISTINCT user_id) INTO v_user_count
    FROM iam_access_control.rate_limit_logs
    WHERE ip_address = p_ip_address
    AND action_type = p_action_type
    AND request_time > current_timestamp - p_window_size;

    -- Calcular score de risco
    IF v_request_count > 100 THEN
        v_score := v_score + 30;
    END IF;

    IF v_user_count > 5 THEN
        v_score := v_score + 20;
    END IF;

    -- Verificar padrões de ataque
    IF EXISTS (
        SELECT 1 
        FROM iam_access_control.security_events 
        WHERE event_type = 'RATE_LIMIT_VIOLATION' 
        AND event_time > current_timestamp - INTERVAL '5 minutes' 
        AND event_details->>'ip_address' = p_ip_address
    ) THEN
        v_score := v_score + 25;
    END IF;

    -- Gerar recomendações
    v_recommendations := CASE 
        WHEN v_score >= 90 THEN jsonb_build_array(
            'Bloquear IP temporariamente',
            'Notificar equipe de segurança',
            'Habilitar rate limiting estrito',
            'Revisar logs de acesso'
        )
        WHEN v_score >= 70 THEN jsonb_build_array(
            'Habilitar rate limiting',
            'Monitorar comportamento',
            'Avisar equipe de segurança'
        )
        ELSE jsonb_build_array(
            'Monitorar atividade',
            'Registrar tentativa',
            'Analisar padrões'
        )
    END;

    -- Registrar evento
    INSERT INTO iam_access_control.security_events (
        event_type,
        ip_address,
        event_details,
        event_time,
        risk_score
    ) VALUES (
        CASE 
            WHEN v_score >= 70 THEN 'RATE_LIMIT_VIOLATION'
            ELSE 'RATE_LIMIT_CHECK'
        END,
        p_ip_address,
        jsonb_build_object(
            'action_type', p_action_type,
            'request_count', v_request_count,
            'user_count', v_user_count,
            'score', v_score,
            'detalhes', jsonb_build_object(
                'high_request_count', v_request_count > 100,
                'multiple_users', v_user_count > 5,
                'recent_violations', EXISTS (
                    SELECT 1 
                    FROM iam_access_control.security_events 
                    WHERE event_type = 'RATE_LIMIT_VIOLATION' 
                    AND event_time > current_timestamp - INTERVAL '5 minutes' 
                    AND event_details->>'ip_address' = p_ip_address
                )
            )
        ),
        current_timestamp,
        v_score
    );

    -- Registrar na tabela de logs
    INSERT INTO iam_access_control.rate_limit_logs (
        ip_address,
        user_id,
        action_type,
        request_time
    ) VALUES (
        p_ip_address,
        p_user_id,
        p_action_type,
        current_timestamp
    );

    v_result := json_build_object(
        'success', v_score < 70,
        'message', CASE 
            WHEN v_score >= 70 THEN 'Limite de taxa excedido'
            ELSE 'Taxa de requisições dentro do limite'
        END,
        'score', v_score,
        'recommendations', v_recommendations
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.4 Monitoramento de Comportamento Anormal
CREATE OR REPLACE FUNCTION iam_access_control.detect_abnormal_behavior(
    p_user_id TEXT,
    p_action_type TEXT,
    p_action_time TIMESTAMP WITH TIME ZONE
)
RETURNS JSON AS $$
DECLARE
    v_base_line JSONB;
    v_current_behavior JSONB;
    v_score INT := 0;
    v_result JSON;
    v_recommendations JSONB;
BEGIN
    -- Obter linha de base do comportamento
    SELECT jsonb_agg(jsonb_build_object(
        'action_type', action_type,
        'average_time', AVG(EXTRACT(EPOCH FROM (action_time - LAG(action_time) OVER (PARTITION BY action_type ORDER BY action_time))))::INT,
        'count', COUNT(*)
    )) INTO v_base_line
    FROM iam_access_control.user_behavior_logs
    WHERE user_id = p_user_id
    AND action_time > current_timestamp - INTERVAL '30 dias'
    GROUP BY action_type;

    -- Calcular comportamento atual
    WITH current_actions AS (
        SELECT 
            action_type,
            COUNT(*) as count,
            AVG(EXTRACT(EPOCH FROM (action_time - LAG(action_time) OVER (PARTITION BY action_type ORDER BY action_time))))::INT as avg_time
        FROM iam_access_control.user_behavior_logs
        WHERE user_id = p_user_id
        AND action_time > current_timestamp - INTERVAL '1 hora'
        GROUP BY action_type
    )
    SELECT jsonb_agg(to_jsonb(current_actions)) INTO v_current_behavior
    FROM current_actions;

    -- Calcular score de risco
    IF EXISTS (
        SELECT 1 
        FROM jsonb_array_elements(v_current_behavior) current
        JOIN jsonb_array_elements(v_base_line) baseline
        ON current->>'action_type' = baseline->>'action_type'
        WHERE ABS((current->>'avg_time')::INT - (baseline->>'avg_time')::INT) > ((baseline->>'avg_time')::INT * 0.5)
    ) THEN
        v_score := v_score + 30;
    END IF;

    IF EXISTS (
        SELECT 1 
        FROM jsonb_array_elements(v_current_behavior) current
        JOIN jsonb_array_elements(v_base_line) baseline
        ON current->>'action_type' = baseline->>'action_type'
        WHERE (current->>'count')::INT > ((baseline->>'count')::INT * 2)
    ) THEN
        v_score := v_score + 25;
    END IF;

    -- Gerar recomendações
    v_recommendations := CASE 
        WHEN v_score >= 90 THEN jsonb_build_array(
            'Avisar equipe de segurança',
            'Monitorar comportamento',
            'Revisar logs de atividade',
            'Notificar usuário'
        )
        WHEN v_score >= 70 THEN jsonb_build_array(
            'Monitorar comportamento',
            'Avisar equipe de segurança',
            'Analisar padrões'
        )
        ELSE jsonb_build_array(
            'Monitorar atividade',
            'Registrar comportamento',
            'Analisar padrões'
        )
    END;

    -- Registrar evento
    INSERT INTO iam_access_control.security_events (
        event_type,
        user_id,
        event_details,
        event_time,
        risk_score
    ) VALUES (
        CASE 
            WHEN v_score >= 70 THEN 'ABNORMAL_BEHAVIOR'
            ELSE 'BEHAVIOR_ANALYSIS'
        END,
        p_user_id,
        jsonb_build_object(
            'base_line', v_base_line,
            'current_behavior', v_current_behavior,
            'score', v_score,
            'detalhes', jsonb_build_object(
                'time_deviation', EXISTS (
                    SELECT 1 
                    FROM jsonb_array_elements(v_current_behavior) current
                    JOIN jsonb_array_elements(v_base_line) baseline
                    ON current->>'action_type' = baseline->>'action_type'
                    WHERE ABS((current->>'avg_time')::INT - (baseline->>'avg_time')::INT) > ((baseline->>'avg_time')::INT * 0.5)
                ),
                'count_deviation', EXISTS (
                    SELECT 1 
                    FROM jsonb_array_elements(v_current_behavior) current
                    JOIN jsonb_array_elements(v_base_line) baseline
                    ON current->>'action_type' = baseline->>'action_type'
                    WHERE (current->>'count')::INT > ((baseline->>'count')::INT * 2)
                )
            )
        ),
        current_timestamp,
        v_score
    );

    -- Registrar comportamento
    INSERT INTO iam_access_control.user_behavior_logs (
        user_id,
        action_type,
        action_time
    ) VALUES (
        p_user_id,
        p_action_type,
        p_action_time
    );

    v_result := json_build_object(
        'user_id', p_user_id,
        'action_type', p_action_type,
        'score', v_score,
        'recommendations', v_recommendations
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.5 Detecção de Malware
CREATE OR REPLACE FUNCTION iam_access_control.detect_malware(
    p_file_content BYTEA,
    p_file_type TEXT,
    p_file_name TEXT
)
RETURNS JSON AS $$
DECLARE
    v_score INT := 0;
    v_is_suspicious BOOLEAN;
    v_result JSON;
    v_recommendations JSONB;
BEGIN
    -- Verificar assinaturas conhecidas
    IF EXISTS (
        SELECT 1 
        FROM iam_access_control.known_malware_signatures
        WHERE signature_type = p_file_type
        AND signature ~* encode(p_file_content, 'hex')
        AND active = TRUE
    ) THEN
        v_score := v_score + 50;
    END IF;

    -- Verificar extensões suspeitas
    IF p_file_name ~* '\.(exe|dll|js|vbs|bat|cmd|hta|scr|pif|cpl|sys|com)$' THEN
        v_score := v_score + 30;
    END IF;

    -- Verificar tamanho suspeito
    IF octet_length(p_file_content) < 100 OR octet_length(p_file_content) > 10000000 THEN
        v_score := v_score + 20;
    END IF;

    -- Verificar padrões de malware
    IF EXISTS (
        SELECT 1 
        FROM iam_access_control.malware_patterns
        WHERE pattern_type = p_file_type
        AND pattern ~* encode(p_file_content, 'hex')
        AND active = TRUE
    ) THEN
        v_score := v_score + 40;
    END IF;

    -- Determinar se é suspeito
    v_is_suspicious := v_score >= 70;

    -- Gerar recomendações
    v_recommendations := CASE 
        WHEN v_score >= 90 THEN jsonb_build_array(
            'Bloquear arquivo imediatamente',
            'Notificar equipe de segurança',
            'Isolar arquivo',
            'Revisar logs de upload',
            'Investigar origem'
        )
        WHEN v_score >= 70 THEN jsonb_build_array(
            'Isolar arquivo',
            'Avisar equipe de segurança',
            'Monitorar atividade',
            'Analisar conteúdo'
        )
        ELSE jsonb_build_array(
            'Monitorar arquivo',
            'Registrar atividade',
            'Analisar padrões'
        )
    END;

    -- Registrar evento
    INSERT INTO iam_access_control.security_events (
        event_type,
        event_details,
        event_time,
        risk_score
    ) VALUES (
        CASE 
            WHEN v_is_suspicious THEN 'MALWARE_DETECTION'
            ELSE 'FILE_UPLOAD'
        END,
        jsonb_build_object(
            'file_name', p_file_name,
            'file_type', p_file_type,
            'score', v_score,
            'detalhes', jsonb_build_object(
                'has_malware_signature', EXISTS (
                    SELECT 1 
                    FROM iam_access_control.known_malware_signatures
                    WHERE signature_type = p_file_type
                    AND signature ~* encode(p_file_content, 'hex')
                    AND active = TRUE
                ),
                'has_suspicious_extension', p_file_name ~* '\.(exe|dll|js|vbs|bat|cmd|hta|scr|pif|cpl|sys|com)$',
                'has_suspicious_size', octet_length(p_file_content) < 100 OR octet_length(p_file_content) > 10000000,
                'has_malware_pattern', EXISTS (
                    SELECT 1 
                    FROM iam_access_control.malware_patterns
                    WHERE pattern_type = p_file_type
                    AND pattern ~* encode(p_file_content, 'hex')
                    AND active = TRUE
                )
            )
        ),
        current_timestamp,
        v_score
    );

    v_result := json_build_object(
        'success', NOT v_is_suspicious,
        'message', CASE 
            WHEN v_is_suspicious THEN 'Arquivo suspeito detectado'
            ELSE 'Arquivo válido'
        END,
        'score', v_score,
        'recommendations', v_recommendations
    );

    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.6 Detecção de DDoS
    p_ip_address TEXT,
    p_request_time TIMESTAMP WITH TIME ZONE,
    p_request_type TEXT,
    p_request_path TEXT
)
RETURNS JSON AS $$
DECLARE
    v_request_count INT;
    v_ip_count INT;
    v_path_count INT;
    v_ddos_score INT := 0;
    v_is_ddos BOOLEAN;
    v_result JSON;
    v_recommendations JSONB;
    v_ip_requests INT;
    v_user_agents INT;
    v_unique_ips INT;
    v_total_requests INT;
BEGIN
    -- Verificar número de requisições recentes
    SELECT COUNT(*) INTO v_request_count
    FROM iam_access_control.access_logs
    WHERE request_time > current_timestamp - INTERVAL '1 minute';
    
    -- Verificar número de IPs únicos recentes
    SELECT COUNT(DISTINCT ip_address) INTO v_ip_count
    FROM iam_access_control.access_logs
    WHERE request_time > current_timestamp - INTERVAL '1 minute';
    
    -- Verificar número de requisições para o mesmo endpoint
    SELECT COUNT(*) INTO v_path_count
    FROM iam_access_control.access_logs
    WHERE request_path = p_request_path
    AND request_time > current_timestamp - INTERVAL '1 minute';
    
    -- Verificar número de requisições por IP
    WITH ip_requests AS (
        SELECT 
            COUNT(*) as ip_requests,
            COUNT(DISTINCT user_agent) as user_agents
        FROM iam_access_control.access_logs
        WHERE ip_address = p_ip_address
        AND request_time > current_timestamp - INTERVAL '1 minute'
    )
    SELECT ip_requests, user_agents INTO v_ip_requests, v_user_agents
    FROM ip_requests;
    
    -- Calcular score de DDoS
    v_ddos_score := 0;
    
    -- Requisições muito frequentes
    IF v_request_count > 800 THEN
        v_ddos_score := v_ddos_score + 45;
    END IF;
    
    -- Muitos IPs diferentes
    IF v_ip_count > 40 THEN
        v_ddos_score := v_ddos_score + 25;
    END IF;
    
    -- Ataque concentrado em um endpoint
    IF v_path_count > 400 THEN
        v_ddos_score := v_ddos_score + 35;
    END IF;
    
    -- Verificar padrões de ataque conhecidos
    IF EXISTS (
        SELECT 1
        FROM iam_access_control.known_ddos_patterns
        WHERE pattern_type = p_request_type
        AND pattern_path = p_request_path
        AND active = TRUE
    ) THEN
        v_ddos_score := v_ddos_score + 50;
    END IF;
    
    -- Verificar intervalos entre requisições
    WITH request_times AS (
        SELECT 
            request_time,
            LAG(request_time) OVER (ORDER BY request_time) as prev_request_time
        FROM iam_access_control.access_logs
        WHERE ip_address = p_ip_address
        AND request_time > current_timestamp - INTERVAL '1 minute'
    )
    SELECT 
        CASE 
            WHEN AVG(EXTRACT(EPOCH FROM (request_time - prev_request_time))) < 0.1 THEN 20
            ELSE 0
        END as rapid_fire_score
    INTO v_rapid_fire_score
    FROM request_times
    WHERE prev_request_time IS NOT NULL;
    
    v_ddos_score := v_ddos_score + v_rapid_fire_score;
    
    -- Verificar padrões de user agent
    IF v_user_agents > 5 THEN
        v_ddos_score := v_ddos_score + 30;
    END IF;
    
    -- Verificar padrões de ataque distribuído
    WITH distributed_attack AS (
        SELECT 
            COUNT(DISTINCT ip_address) as unique_ips,
            COUNT(*) as total_requests
        FROM iam_access_control.access_logs
        WHERE request_time > current_timestamp - INTERVAL '5 seconds'
    )
    SELECT unique_ips, total_requests INTO v_unique_ips, v_total_requests
    FROM distributed_attack;
    
    IF v_unique_ips > 10 AND v_total_requests > 100 THEN
        v_ddos_score := v_ddos_score + 40;
    END IF;
    
    -- Determinar se é DDoS
    v_is_ddos := v_ddos_score >= 70;
    
    -- Gerar recomendações detalhadas
    v_recommendations := CASE 
        WHEN v_ddos_score >= 90 THEN jsonb_build_array(
            'Bloquear IP temporariamente por 24 horas',
            'Notificar equipe de segurança imediatamente',
            'Habilitar rate limiting estrito',
            'Monitorar endpoints afetados',
            'Iniciar investigação completa',
            'Revisar logs de acesso',
            'Verificar atividades relacionadas'
        )
        WHEN v_ddos_score >= 70 THEN jsonb_build_array(
            'Bloquear IP temporariamente por 4 horas',
            'Notificar equipe de segurança',
            'Habilitar rate limiting',
            'Monitorar endpoints',
            'Analisar padrão de acesso'
        )
        WHEN v_ddos_score >= 50 THEN jsonb_build_array(
            'Avisar equipe de segurança',
            'Habilitar rate limiting leve',
            'Monitorar comportamento',
            'Analisar histórico',
            'Verificar padrões'
        )
        ELSE jsonb_build_array(
            'Monitoramento básico',
            'Registro de atividade',
            'Análise de padrões'
        )
    END;
    
    -- Registrar evento detalhado
    INSERT INTO iam_access_control.security_events (
        event_type,
        ip_address,
        event_details,
        event_time,
        risk_score
    ) VALUES (
        CASE 
            WHEN v_is_ddos THEN 'DDOS_ATTACK'
            ELSE 'ACCESS_ATTEMPT'
        END,
        p_ip_address,
        jsonb_build_object(
            'request_type', p_request_type,
            'request_path', p_request_path,
            'ddos_score', v_ddos_score,
            'detalhes', jsonb_build_object(
                'request_count', v_request_count,
                'ip_count', v_ip_count,
                'path_count', v_path_count,
                'ip_requests', v_ip_requests,
                'user_agents', v_user_agents,
                'unique_ips', v_unique_ips,
                'total_requests', v_total_requests
            ),
            'motivos', jsonb_build_array(
                CASE WHEN v_request_count > 800 THEN 'MUITAS_REQUISICOES' END,
                CASE WHEN v_ip_count > 40 THEN 'MUITOS_IPS' END,
                CASE WHEN v_path_count > 400 THEN 'CONCENTRACAO_ENDPOINT' END,
                CASE WHEN v_user_agents > 5 THEN 'MUITOS_USER_AGENTS' END,
                CASE WHEN v_unique_ips > 10 AND v_total_requests > 100 THEN 'ATAQUE_DISTRIBUIDO' END
            )::jsonb
        ),
        p_request_time,
        v_ddos_score
    );
    
    -- Retornar resultado detalhado
    v_result := json_build_object(
        'success', NOT v_is_ddos,
        'message', CASE 
            WHEN v_is_ddos THEN 'Ataque DDoS detectado'
            ELSE 'Acesso normal'
        END,
        'ddos_score', v_ddos_score,
        'details', json_build_object(
            'request_count', v_request_count,
            'ip_count', v_ip_count,
            'path_count', v_path_count,
            'ip_requests', v_ip_requests,
            'user_agents', v_user_agents,
            'unique_ips', v_unique_ips,
            'total_requests', v_total_requests
        ),
        'recommendations', v_recommendations
    );
    
    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.2 Proteção contra SQL Injection
CREATE OR REPLACE FUNCTION iam_access_control.prevent_sql_injection(
    p_input TEXT,
    p_input_type TEXT,
    p_ip_address TEXT,
    p_request_time TIMESTAMP WITH TIME ZONE
)
RETURNS JSON AS $$
DECLARE
    v_score INT := 0;
    v_is_suspicious BOOLEAN;
    v_result JSON;
    v_recommendations JSONB;
BEGIN
    -- Verificar caracteres especiais
    IF p_input ~* '[\;\'\"\\]' THEN
        v_score := v_score + 20;
    END IF;
    
    -- Verificar palavras-chave SQL
    IF p_input ~* '\b(SELECT|INSERT|UPDATE|DELETE|DROP|ALTER|CREATE|GRANT|REVOKE|EXEC|UNION|FROM|WHERE|GROUP BY|ORDER BY)\b' THEN
        v_score := v_score + 30;
    END IF;
    
    -- Verificar comentários SQL
    IF p_input ~* '--|/*|*/' THEN
        v_score := v_score + 20;
    END IF;
    
    -- Verificar caracteres de escape
    IF p_input ~* '\\[xX][0-9A-Fa-f]{2}|\\u[0-9A-Fa-f]{4}' THEN
        v_score := v_score + 15;
    END IF;
    
    -- Verificar padrões de injeção
    IF EXISTS (
        SELECT 1
        FROM iam_access_control.known_injection_patterns
        WHERE pattern_type = p_input_type
        AND pattern ~* p_input
        AND active = TRUE
    ) THEN
        v_score := v_score + 40;
    END IF;
    
    -- Verificar padrões de injeção comuns
    IF p_input ~* '\b(OR|AND)\s+1=1\b|\b1=1\b|\b1=1\b|\b1=1\b|\b1=1\b|\b1=1\b' THEN
        v_score := v_score + 35;
    END IF;
    
    -- Verificar uso de funções SQL
    IF p_input ~* '\b(SLEEP|BENCHMARK|WAITFOR|EXECUTE|EXEC|SP_EXECUTESQL|SYS\.|INFORMATION_SCHEMA\.)\b' THEN
        v_score := v_score + 45;
    END IF;
    
    -- Determinar se é suspeito
    v_is_suspicious := v_score >= 50;
    
    -- Gerar recomendações detalhadas
    v_recommendations := CASE 
        WHEN v_score >= 90 THEN jsonb_build_array(
            'Validar entrada estritamente',
            'Usar prepared statements obrigatórios',
            'Notificar equipe de segurança imediatamente',
            'Bloquear IP por 24 horas',
            'Iniciar investigação completa',
            'Revisar logs de acesso',
            'Verificar atividades relacionadas'
        )
        WHEN v_score >= 70 THEN jsonb_build_array(
            'Validar entrada estritamente',
            'Usar prepared statements',
            'Notificar equipe de segurança',
            'Bloquear IP temporariamente',
            'Monitorar endpoint',
            'Analisar padrão de acesso'
        )
        WHEN v_score >= 50 THEN jsonb_build_array(
            'Validar entrada',
            'Usar prepared statements',
            'Monitorar comportamento',
            'Analisar histórico',
            'Verificar padrões'
        )
        ELSE jsonb_build_array(
            'Monitoramento básico',
            'Registro de atividade',
            'Análise de padrões'
        )
    END;
    
    -- Registrar evento detalhado
    INSERT INTO iam_access_control.security_events (
        event_type,
        ip_address,
        event_details,
        event_time,
        risk_score
    ) VALUES (
        CASE 
            WHEN v_is_suspicious THEN 'SQL_INJECTION_ATTEMPT'
            ELSE 'INPUT_VALIDATION'
        END,
        p_ip_address,
        jsonb_build_object(
            'input_type', p_input_type,
            'score_risco', v_score,
            'detalhes', jsonb_build_object(
                'caracteres_especiais', p_input ~* '[\;\'\"\\]',
                'palavras_chave_sql', p_input ~* '\b(SELECT|INSERT|UPDATE|DELETE|DROP|ALTER|CREATE|GRANT|REVOKE|EXEC|UNION|FROM|WHERE|GROUP BY|ORDER BY)\b',
                'comentarios_sql', p_input ~* '--|/*|*/',
                'caracteres_escape', p_input ~* '\\[xX][0-9A-Fa-f]{2}|\\u[0-9A-Fa-f]{4}',
                'padroes_injecao', EXISTS (
                    SELECT 1
                    FROM iam_access_control.known_injection_patterns
                    WHERE pattern_type = p_input_type
                    AND pattern ~* p_input
                    AND active = TRUE
                ),
                'padroes_comuns', p_input ~* '\b(OR|AND)\s+1=1\b|\b1=1\b',
                'funcoes_sql', p_input ~* '\b(SLEEP|BENCHMARK|WAITFOR|EXECUTE|EXEC|SP_EXECUTESQL|SYS\.|INFORMATION_SCHEMA\.)\b'
            )
        ),
        p_request_time,
        v_score
    );
    
    -- Retornar resultado detalhado
    v_result := json_build_object(
        'success', NOT v_is_suspicious,
        'message', CASE 
            WHEN v_is_suspicious THEN 'Tentativa de injeção SQL detectada'
            ELSE 'Entrada válida'
        END,
        'score_risco', v_score,
        'recommendations', v_recommendations,
        'details', json_build_object(
            'caracteres_especiais', p_input ~* '[\;\'\"\\]',
            'palavras_chave_sql', p_input ~* '\b(SELECT|INSERT|UPDATE|DELETE|DROP|ALTER|CREATE|GRANT|REVOKE|EXEC|UNION|FROM|WHERE|GROUP BY|ORDER BY)\b',
            'comentarios_sql', p_input ~* '--|/*|*/',
            'caracteres_escape', p_input ~* '\\[xX][0-9A-Fa-f]{2}|\\u[0-9A-Fa-f]{4}',
            'padroes_injecao', EXISTS (
                SELECT 1
                FROM iam_access_control.known_injection_patterns
                WHERE pattern_type = p_input_type
                AND pattern ~* p_input
                AND active = TRUE
            ),
            'padroes_comuns', p_input ~* '\b(OR|AND)\s+1=1\b|\b1=1\b',
            'funcoes_sql', p_input ~* '\b(SLEEP|BENCHMARK|WAITFOR|EXECUTE|EXEC|SP_EXECUTESQL|SYS\.|INFORMATION_SCHEMA\.)\b'
        )
    );
    
    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.3 Proteção contra Brute Force
CREATE OR REPLACE FUNCTION iam_access_control.prevent_brute_force(
    p_username TEXT,
    p_ip_address TEXT,
    p_attempt_time TIMESTAMP WITH TIME ZONE
)
RETURNS JSON AS $$
DECLARE
    v_attempts INT;
    v_last_attempt TIMESTAMP WITH TIME ZONE;
    v_account_locked BOOLEAN;
    v_lock_expires_at TIMESTAMP WITH TIME ZONE;
    v_ip_attempts INT;
    v_global_attempts INT;
    v_result JSON;
BEGIN
    -- Verificar estado atual da conta
    SELECT 
        account_locked,
        lock_expires_at
    INTO v_account_locked, v_lock_expires_at
    FROM iam_access_control.users
    WHERE username = p_username;
    
    -- Se conta está bloqueada e ainda não expirou
    IF v_account_locked AND v_lock_expires_at > current_timestamp THEN
        RETURN json_build_object(
            'success', FALSE,
            'message', 'Conta bloqueada temporariamente',
            'reason', 'FORÇA_BRUTA',
            'unlock_time', v_lock_expires_at
        );
    END IF;
    
    -- Verificar tentativas recentes por usuário
    SELECT COUNT(*) INTO v_attempts
    FROM iam_access_control.login_attempts
    WHERE username = p_username
    AND attempt_time > current_timestamp - INTERVAL '15 minutes';
    
    -- Verificar tentativas recentes por IP
    SELECT COUNT(*) INTO v_ip_attempts
    FROM iam_access_control.login_attempts
    WHERE ip_address = p_ip_address
    AND attempt_time > current_timestamp - INTERVAL '15 minutes';
    
    -- Verificar tentativas globais recentes
    SELECT COUNT(*) INTO v_global_attempts
    FROM iam_access_control.login_attempts
    WHERE attempt_time > current_timestamp - INTERVAL '15 minutes';
    
    -- Definir limites dinâmicos baseados no risco
    IF v_global_attempts > 100 THEN
        -- Modo de emergência: limites mais rígidos
        IF v_attempts >= 3 OR v_ip_attempts >= 3 THEN
            -- Bloqueio progressivo
            CASE 
                WHEN v_attempts >= 8 THEN
                    -- Bloqueio por 24 horas
                    UPDATE iam_access_control.users
                    SET account_locked = TRUE,
                        lock_expires_at = current_timestamp + INTERVAL '24 hours'
                    WHERE username = p_username;
                    RETURN json_build_object(
                        'success', FALSE,
                        'message', 'Conta bloqueada temporariamente (24h) por força bruta',
                        'reason', 'FORÇA_BRUTA_SEVERA',
                        'unlock_time', current_timestamp + INTERVAL '24 hours'
                    );
                WHEN v_attempts >= 5 THEN
                    -- Bloqueio por 4 horas
                    UPDATE iam_access_control.users
                    SET account_locked = TRUE,
                        lock_expires_at = current_timestamp + INTERVAL '4 hours'
                    WHERE username = p_username;
                    RETURN json_build_object(
                        'success', FALSE,
                        'message', 'Conta bloqueada temporariamente (4h) por força bruta',
                        'reason', 'FORÇA_BRUTA_MODERADA',
                        'unlock_time', current_timestamp + INTERVAL '4 hours'
                    );
                ELSE
                    -- Bloqueio por 1 hora
                    UPDATE iam_access_control.users
                    SET account_locked = TRUE,
                        lock_expires_at = current_timestamp + INTERVAL '1 hour'
                    WHERE username = p_username;
                    RETURN json_build_object(
                        'success', FALSE,
                        'message', 'Conta bloqueada temporariamente (1h) por força bruta',
                        'reason', 'FORÇA_BRUTA_LEVE',
                        'unlock_time', current_timestamp + INTERVAL '1 hour'
                    );
                END;
        END IF;
    ELSE
        -- Modo normal: limites padrão
        IF v_attempts >= 5 OR v_ip_attempts >= 5 THEN
            -- Bloqueio progressivo
            CASE 
                WHEN v_attempts >= 8 THEN
                    -- Bloqueio por 12 horas
                    UPDATE iam_access_control.users
                    SET account_locked = TRUE,
                        lock_expires_at = current_timestamp + INTERVAL '12 hours'
                    WHERE username = p_username;
                    RETURN json_build_object(
                        'success', FALSE,
                        'message', 'Conta bloqueada temporariamente (12h) por força bruta',
                        'reason', 'FORÇA_BRUTA_SEVERA',
                        'unlock_time', current_timestamp + INTERVAL '12 hours'
                    );
                WHEN v_attempts >= 5 THEN
                    -- Bloqueio por 3 horas
                    UPDATE iam_access_control.users
                    SET account_locked = TRUE,
                        lock_expires_at = current_timestamp + INTERVAL '3 hours'
                    WHERE username = p_username;
                    RETURN json_build_object(
                        'success', FALSE,
                        'message', 'Conta bloqueada temporariamente (3h) por força bruta',
                        'reason', 'FORÇA_BRUTA_MODERADA',
                        'unlock_time', current_timestamp + INTERVAL '3 hours'
                    );
                ELSE
                    -- Bloqueio por 1 hora
                    UPDATE iam_access_control.users
                    SET account_locked = TRUE,
                        lock_expires_at = current_timestamp + INTERVAL '1 hour'
                    WHERE username = p_username;
                    RETURN json_build_object(
                        'success', FALSE,
                        'message', 'Conta bloqueada temporariamente (1h) por força bruta',
                        'reason', 'FORÇA_BRUTA_LEVE',
                        'unlock_time', current_timestamp + INTERVAL '1 hour'
                    );
                END;
        END IF;
    END IF;
    
    -- Se número de tentativas excedeu o limite
    IF v_attempts >= 5 THEN
        -- Determinar tempo de bloqueio baseado no número de tentativas
        DECLARE
            v_block_duration INTERVAL := CASE 
                WHEN v_attempts >= 10 THEN INTERVAL '24 horas'
                WHEN v_attempts >= 7 THEN INTERVAL '4 horas'
                ELSE INTERVAL '1 hora'
            END;
        
        -- Bloquear conta
        UPDATE iam_access_control.users
        SET 
            account_locked = TRUE,
            lock_expires_at = current_timestamp + v_block_duration
        WHERE username = p_username;
        
        -- Registrar o bloqueio
        INSERT INTO iam_access_control.security_events (
            event_type,
            username,
            ip_address,
            event_details,
            event_time,
            risk_score
        ) VALUES (
            'ACCOUNT_LOCKED',
            p_username,
            p_ip_address,
            jsonb_build_object(
                'reason', 'Proteção contra força bruta',
                'tentativas_usuario', v_attempts,
                'tentativas_ip', v_ip_attempts,
                'tentativas_globais', v_global_attempts,
                'bloqueio', CASE 
                    WHEN v_attempts >= 10 THEN 'ALTO_RISCO'
                    WHEN v_attempts >= 7 THEN 'MODERADO'
                    ELSE 'BAIXO'
                END
            ),
            current_timestamp,
            CASE 
                WHEN v_attempts >= 10 THEN 90
                WHEN v_attempts >= 7 THEN 70
                ELSE 50
            END
        );
        
        RETURN json_build_object(
            'success', FALSE,
            'message', 'Bloqueio temporário por força bruta',
            'reason', 'FORÇA_BRUTA',
            'unlock_time', current_timestamp + v_block_duration,
            'risk_level', CASE 
                WHEN v_attempts >= 10 THEN 'ALTO'
                WHEN v_attempts >= 7 THEN 'MODERADO'
                ELSE 'BAIXO'
            END
        );
    END IF;
    
    -- Registrar tentativa
    INSERT INTO iam_access_control.login_attempts (
        username,
        ip_address,
        attempt_time,
        success
    ) VALUES (
        p_username,
        p_ip_address,
        p_attempt_time,
        FALSE
    );
    
    -- Limpar tentativas antigas com base no risco atual
    DELETE FROM iam_access_control.login_attempts
    WHERE attempt_time < CASE 
        -- Em modo emergência (global_attempts > 100)
        WHEN (SELECT COUNT(*) FROM iam_access_control.login_attempts 
              WHERE attempt_time > current_timestamp - INTERVAL '15 minutos') > 100 
        THEN current_timestamp - INTERVAL '12 horas'
        -- Modo normal
        ELSE current_timestamp - INTERVAL '24 horas'
        END;
    
    RETURN json_build_object(
        'success', TRUE,
        'message', 'Tentativa registrada',
        'attempts', v_attempts,
        'ip_attempts', v_ip_attempts,
        'global_attempts', v_global_attempts,
        'risk_level', CASE 
            -- Modo emergência (global_attempts > 100)
            WHEN (SELECT COUNT(*) FROM iam_access_control.login_attempts 
                  WHERE attempt_time > current_timestamp - INTERVAL '15 minutos') > 100 
            THEN CASE 
                WHEN v_attempts >= 3 THEN 'ALTO'
                WHEN v_attempts >= 2 THEN 'MODERADO'
                ELSE 'BAIXO'
            END
            -- Modo normal
            ELSE CASE 
                WHEN v_attempts >= 7 THEN 'ALTO'
                WHEN v_attempts >= 5 THEN 'MODERADO'
                WHEN v_attempts >= 3 THEN 'BAIXO'
                ELSE 'NORMAL'
            END
        END
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.2 Detecção de Acesso Suspeito
CREATE OR REPLACE FUNCTION iam_access_control.detect_suspicious_access(
    p_username TEXT,
    p_ip_address TEXT,
    p_country_code TEXT
)
RETURNS JSON AS $$
DECLARE
    v_user_record RECORD;
    v_risk_score INT;
    v_suspicious BOOLEAN;
    v_result JSON;
BEGIN
    -- Obter informações do usuário
    SELECT * INTO v_user_record
    FROM iam_access_control.users
    WHERE username = p_username;
    
    -- Calcular score de risco
    v_risk_score := 0;
    
    -- Verificar histórico de IPs (últimas 24h)
    SELECT COUNT(*) INTO v_risk_score
    FROM iam_access_control.access_logs
    WHERE username = p_username
    AND ip_address = p_ip_address
    AND access_time > current_timestamp - INTERVAL '24 horas'
    
    -- Ajustar score baseado no horário
    IF EXTRACT(HOUR FROM current_timestamp) BETWEEN 0 AND 6 THEN
        v_risk_score := v_risk_score + 10;
    END IF;
    
    -- Verificar países de risco
    IF p_country_code IN ('RU', 'CN', 'IR', 'KP') THEN
        v_risk_score := v_risk_score + 20;
    END IF;
    
    -- Verificar IP em lista negra
    IF EXISTS (
        SELECT 1 FROM iam_access_control.blacklisted_ips
        WHERE ip_address = p_ip_address
    ) THEN
        v_risk_score := v_risk_score + 30;
    END IF;
    
    -- Verificar padrões de acesso suspeitos
    IF EXISTS (
        SELECT 1 FROM iam_access_control.access_patterns
        WHERE ip_address = p_ip_address
        AND pattern_type = 'SUSPECT'
        AND last_detected > current_timestamp - INTERVAL '1 hora'
    ) THEN
        v_risk_score := v_risk_score + 15;
    END IF;
    
    -- Ajustar score baseado em tentativas recentes
    SELECT COUNT(*) INTO v_attempts
    FROM iam_access_control.login_attempts
    WHERE username = p_username
    AND attempt_time > current_timestamp - INTERVAL '15 minutos';
    
    IF v_attempts >= 3 THEN
        v_risk_score := v_risk_score + 25;
    END IF;
    AND access_time > current_timestamp - INTERVAL '30 dias';
    
    -- Verificar se o país está na lista de países de risco
    IF p_country_code IN (
        SELECT country_code
        FROM iam_access_control.risky_countries
        WHERE active = TRUE
    ) THEN
        v_risk_score := v_risk_score + 30;
    END IF;
    
    -- Verificar horário incomum
    IF EXTRACT(HOUR FROM current_timestamp) BETWEEN 0 AND 5 THEN
        v_risk_score := v_risk_score + 15;
    END IF;
    
    -- Verificar se o IP está em lista negra
    IF EXISTS (
        SELECT 1
        FROM iam_access_control.blacklisted_ips
        WHERE ip_address = p_ip_address
        AND active = TRUE
    ) THEN
        v_risk_score := v_risk_score + 50;
    END IF;
    
    -- Verificar se o usuário está em uma região restrita
    IF EXISTS (
        SELECT 1
        FROM iam_access_control.restricted_regions
        WHERE region_code = p_country_code
        AND active = TRUE
    ) THEN
        v_risk_score := v_risk_score + 25;
    END IF;
    
    -- Verificar se o IP é de uma rede conhecida
    IF EXISTS (
        SELECT 1
        FROM iam_access_control.known_networks
        WHERE ip_address = p_ip_address
        AND active = TRUE
    ) THEN
        v_risk_score := v_risk_score - 10;
    END IF;
    
    -- Verificar se o usuário tem histórico de comportamento suspeito
    SELECT COALESCE(SUM(CASE 
        WHEN event_type IN ('SUSPICIOUS_ACCESS', 'ACCOUNT_LOCKED') THEN 20
        WHEN event_type = 'FAILED_LOGIN' THEN 10
        ELSE 0
    END), 0) INTO v_risk_score
    FROM iam_access_control.security_events
    WHERE username = p_username
    AND event_time > current_timestamp - INTERVAL '30 dias';
    
    -- Normalizar score (0-100)
    IF v_risk_score > 100 THEN
        v_risk_score := 100;
    END IF;
    
    -- Determinar se é suspeito
    v_suspicious := v_risk_score >= 70;
    
    -- Registrar evento de detecção
    INSERT INTO iam_access_control.security_events (
        event_type,
        username,
        ip_address,
        event_details,
        event_time,
        risk_score
    ) VALUES (
        CASE 
            WHEN v_suspicious THEN 'SUSPICIOUS_ACCESS'
            ELSE 'ACCESS_ATTEMPT'
        END,
        p_username,
        p_ip_address,
        jsonb_build_object(
            'country_code', p_country_code,
            'risk_score', v_risk_score,
            'reasons', jsonb_build_array(
                CASE WHEN p_country_code IN (
                    SELECT country_code
                    FROM iam_access_control.risky_countries
                    WHERE active = TRUE
                ) THEN 'PAÍS DE RISCO' END,
                CASE WHEN EXTRACT(HOUR FROM current_timestamp) BETWEEN 0 AND 5 THEN 'HORÁRIO INCOMUM' END,
                CASE WHEN EXISTS (
                    SELECT 1
                    FROM iam_access_control.blacklisted_ips
                    WHERE ip_address = p_ip_address
                    AND active = TRUE
                ) THEN 'IP NEGRO' END,
                CASE WHEN EXISTS (
                    SELECT 1
                    FROM iam_access_control.restricted_regions
                    WHERE region_code = p_country_code
                    AND active = TRUE
                ) THEN 'REGIÃO RESTRITA' END
            )::jsonb
        ),
        current_timestamp,
        v_risk_score
    );
    
    -- Retornar resultado
    v_result := json_build_object(
        'username', p_username,
        'ip_address', p_ip_address,
    v_recommendations := CASE 
        WHEN v_risk_score >= 90 THEN jsonb_build_array(
            'Ativar autenticação de dois fatores obrigatória',
            'Bloquear IP temporariamente por 24 horas',
            'Notificar administrador de segurança',
            'Iniciar investigação completa',
            'Monitorar em tempo real',
            'Revisar histórico de acesso',
            'Verificar atividades recentes'
        )
        WHEN v_risk_score >= 70 THEN jsonb_build_array(
            'Requerer autenticação de dois fatores',
            'Bloquear IP temporariamente por 4 horas',
            'Notificar administrador',
            'Monitorar comportamento',
            'Analisar padrões de acesso',
            'Verificar localização'
        )
        WHEN v_risk_score >= 50 THEN jsonb_build_array(
            'Requerer autenticação adicional',
            'Bloquear IP temporariamente por 1 hora',
            'Monitorar acesso',
            'Analisar histórico',
            'Verificar padrões'
        )
        ELSE jsonb_build_array(
            'Monitoramento básico',
            'Registro de atividade',
            'Análise de padrões',
            'Monitorar comportamento'
        )
    END;

    -- Registrar evento detalhado
    INSERT INTO iam_access_control.security_events (
        event_type,
        username,
        ip_address,
        event_details,
        event_time,
        risk_score
    ) VALUES (
        CASE 
            WHEN v_suspicious THEN 'SUSPICIOUS_ACCESS'
            ELSE 'ACCESS_ATTEMPT'
        END,
        p_username,
        p_ip_address,
        jsonb_build_object(
            'score_risco', v_risk_score,
            'motivos', jsonb_build_array(
                CASE WHEN v_risk_score >= 90 THEN 'RISCO_CRITICO' END,
                CASE WHEN v_risk_score >= 70 THEN 'RISCO_ALTO' END,
                CASE WHEN v_risk_score >= 50 THEN 'RISCO_MODERADO' END,
                CASE WHEN v_risk_score < 50 THEN 'RISCO_BAIXO' END
            )::jsonb,
            'detalhes', jsonb_build_object(
                'ip_desconhecido', CASE WHEN NOT EXISTS (
                    SELECT 1 FROM iam_access_control.known_ips
                    WHERE ip_address = p_ip_address
                ) THEN true ELSE false END,
                'mudanca_pais', CASE WHEN v_user_record IS NOT NULL AND v_user_record.country_code IS DISTINCT FROM p_country_code THEN true ELSE false END,
                'horario_incomum', CASE WHEN EXTRACT(HOUR FROM p_access_time) BETWEEN 0 AND 6 THEN true ELSE false END,
                'tentativas_recentes', v_attempts_count,
                'velocidade_acesso', v_access_count,
                'comportamento_anomalo', v_anomaly_count
            )
        ),
        current_timestamp,
        v_risk_score
    );
    
    -- Retornar resultado com mais detalhes
    v_result := json_build_object(
        'username', p_username,
        'ip_address', p_ip_address,
        'score_risco', v_risk_score,
        'suspeito', v_suspicious,
        'detalhes', v_details,
        'motivos', jsonb_build_array(
            CASE WHEN v_details ? 'IP_DESCONHECIDO' THEN 'IP desconhecido' END,
            CASE WHEN v_details ? 'PAIS_ALTERADO' THEN 'Mudança de país' END,
            CASE WHEN v_details ? 'HORARIO_SUSPEITO' THEN 'Horário incomum' END,
            CASE WHEN v_details ? 'TENTATIVAS_RECENTES' THEN 'Muitas tentativas recentes' END,
            CASE WHEN v_details ? 'VELOCIDADE_SUSPEITA' THEN 'Velocidade de acesso suspeita' END,
            CASE WHEN v_details ? 'COMPORTAMENTO_ANOMALO' THEN 'Comportamento anômalo' END
        )::jsonb,
        'recomendacoes', v_recommendations
    );
    
    RETURN v_result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.3 Proteção contra SQL Injection
CREATE OR REPLACE FUNCTION iam_access_control.sanitize_input(
    p_input TEXT
)
RETURNS TEXT AS $$
BEGIN
    -- Remover caracteres especiais potencialmente perigosos
    RETURN REGEXP_REPLACE(
        p_input,
        '[\'";\-\(\)\[\]\{\}\s]+',
        ' ',
        'g'
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1.4 Monitoramento de Acesso
CREATE OR REPLACE FUNCTION iam_access_control.monitor_access_patterns(
    p_username TEXT,
    p_access_time TIMESTAMP WITH TIME ZONE,
    p_resource TEXT
)
RETURNS VOID AS $$
DECLARE
    v_access_pattern RECORD;
    v_anomaly BOOLEAN;
BEGIN
    -- Verificar padrões de acesso
    SELECT 
        COUNT(*) as total_accesses,
        AVG(EXTRACT(EPOCH FROM (access_time - lag(access_time) OVER (ORDER BY access_time))) as avg_interval
    INTO v_access_pattern
    FROM iam_access_control.access_logs
    WHERE username = p_username
    AND access_time > current_timestamp - INTERVAL '24 horas'
    GROUP BY username;
    
    -- Detectar anomalias
    v_anomaly := CASE 
        WHEN v_access_pattern.avg_interval IS NULL THEN FALSE
        WHEN v_access_pattern.avg_interval < 60 THEN TRUE -- Acessos muito frequentes
        ELSE FALSE
    END;
    
    -- Registrar evento de segurança se houver anomalia
    IF v_anomaly THEN
        INSERT INTO iam_access_control.security_events (
            event_type,
            username,
            event_details,
            event_time
        ) VALUES (
            'ANOMALIA_PADRAO_ACESSO',
            p_username,
            jsonb_build_object(
                'recurso', p_resource,
                'intervalo_medio', v_access_pattern.avg_interval,
                'total_acessos', v_access_pattern.total_accesses
            ),
            current_timestamp
        );
    END IF;
    
    -- Registrar acesso
    INSERT INTO iam_access_control.access_logs (
        username,
        access_time,
        resource,
        is_anomalous
    ) VALUES (
        p_username,
        p_access_time,
        p_resource,
        v_anomaly
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 2. Tabelas de Segurança

-- Tabela de Tentativas de Login
CREATE TABLE IF NOT EXISTS iam_access_control.login_attempts (
    id BIGSERIAL PRIMARY KEY,
    username TEXT REFERENCES iam_access_control.users(username),
    ip_address TEXT NOT NULL,
    attempt_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    success BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT unique_attempt UNIQUE (username, ip_address, attempt_time)
);

-- Tabela de IPs Conhecidos
CREATE TABLE IF NOT EXISTS iam_access_control.known_ips (
    id BIGSERIAL PRIMARY KEY,
    ip_address TEXT NOT NULL,
    username TEXT REFERENCES iam_access_control.users(username),
    last_used TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    CONSTRAINT unique_known_ip UNIQUE (ip_address, username)
);

-- Tabela de Eventos de Segurança
CREATE TABLE IF NOT EXISTS iam_access_control.security_events (
    id BIGSERIAL PRIMARY KEY,
    event_type TEXT NOT NULL,
    username TEXT REFERENCES iam_access_control.users(username),
    ip_address TEXT,
    event_details JSONB,
    event_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp
);

-- Tabela de Logs de Acesso
CREATE TABLE IF NOT EXISTS iam_access_control.access_logs (
    id BIGSERIAL PRIMARY KEY,
    username TEXT REFERENCES iam_access_control.users(username),
    access_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
    resource TEXT NOT NULL,
    is_anomalous BOOLEAN NOT NULL DEFAULT FALSE
);

-- 3. Triggers de Segurança

-- Trigger para monitorar mudanças em senhas
CREATE OR REPLACE FUNCTION iam_access_control.monitor_password_changes()
RETURNS TRIGGER AS $$
BEGIN
    -- Verificar se a senha foi alterada
    IF NEW.password_hash IS DISTINCT FROM OLD.password_hash THEN
        -- Registrar evento de segurança
        INSERT INTO iam_access_control.security_events (
            event_type,
            username,
            event_details,
            event_time
        ) VALUES (
            'ALTERACAO_SENHA',
            NEW.username,
            jsonb_build_object(
                'hash_senha_antiga', OLD.password_hash,
                'hash_senha_nova', NEW.password_hash
            ),
            current_timestamp
        );
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Criar trigger
CREATE TRIGGER monitor_password_changes_trigger
AFTER UPDATE ON iam_access_control.users
FOR EACH ROW
EXECUTE FUNCTION iam_access_control.monitor_password_changes();

-- 4. Views de Monitoramento

-- View para monitorar acessos suspeitos
CREATE OR REPLACE VIEW iam_access_control.acessos_suspeitos AS
SELECT 
    s.event_time,
    s.username,
    s.ip_address,
    s.event_details->>'score_risco' as score_risco,
    s.event_details->>'codigo_pais' as codigo_pais,
    s.event_details->>'suspeito' as suspeito
FROM iam_access_control.security_events s
WHERE s.event_type = 'ACESSO_SUSPEITO'
AND s.event_time > current_timestamp - INTERVAL '24 horas';

-- View para monitorar bloqueios de conta
CREATE OR REPLACE VIEW iam_access_control.bloqueios_conta AS
SELECT 
    s.event_time,
    s.username,
    s.event_details->>'razao' as razao,
    s.event_details->>'tentativas' as tentativas
FROM iam_access_control.security_events s
WHERE s.event_type = 'ACCOUNT_LOCKED'
AND s.event_time > current_timestamp - INTERVAL '24 horas';

-- 5. Funções de Auditoria

-- Função para gerar relatório de segurança
CREATE OR REPLACE FUNCTION iam_access_control.gerar_relatorio_seguranca(
    p_data_inicio TIMESTAMP WITH TIME ZONE,
    p_data_fim TIMESTAMP WITH TIME ZONE
)
RETURNS TABLE (
    tipo_evento TEXT,
    username TEXT,
    ip_address TEXT,
    score_risco INT,
    suspeito BOOLEAN,
    contagem_eventos INT,
    primeira_ocorrencia TIMESTAMP WITH TIME ZONE,
    ultima_ocorrencia TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        s.event_type,
        s.username,
        s.ip_address,
        (s.event_details->>'score_risco')::INT as score_risco,
        (s.event_details->>'suspeito')::BOOLEAN as suspeito,
        COUNT(*) as contagem_eventos,
        MIN(s.event_time) as primeira_ocorrencia,
        MAX(s.event_time) as ultima_ocorrencia
    FROM iam_access_control.security_events s
    WHERE s.event_time BETWEEN p_data_inicio AND p_data_fim
    GROUP BY 
        s.event_type,
        s.username,
        s.ip_address,
        (s.event_details->>'score_risco')::INT,
        (s.event_details->>'suspeito')::BOOLEAN
    ORDER BY contagem_eventos DESC;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 6. Políticas de Segurança

-- Política de bloqueio de conta
CREATE POLICY bloquear_contas_bloqueadas
ON iam_access_control.users
FOR ALL
USING (
    NOT account_locked
    OR lock_expires_at < current_timestamp
);

-- Política de acesso a informações sensíveis
CREATE POLICY acesso_dados_sensiveis
ON iam_access_control.security_events
FOR SELECT
TO security_admin
USING (
    event_type IN ('ACESSO_SUSPEITO', 'ACCOUNT_LOCKED')
    AND event_time > current_timestamp - INTERVAL '24 horas'
);
