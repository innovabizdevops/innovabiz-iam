-- ============================================================================
-- INNOVABIZ IAM - WebAuthn Initial Data and Configuration
-- ============================================================================
-- Version: 1.0.0
-- Date: 31/07/2025
-- Author: Equipe de Configuração INNOVABIZ
-- Classification: Confidencial - Interno
-- 
-- Description: Dados iniciais e configurações para WebAuthn/FIDO2
-- Includes: FIDO Metadata, Test Data, Regional Configurations
-- ============================================================================

-- ============================================================================
-- METADADOS DE AUTENTICADORES FIDO CONHECIDOS
-- ============================================================================

-- Inserir metadados de autenticadores populares
INSERT INTO webauthn_authenticator_metadata (
    aaguid, vendor_name, device_name, device_model,
    key_protection, matcher_protection, crypto_strength,
    attachment_hint, is_key_restricted, is_fresh_user_verification_required,
    supported_algorithms, certification_level, fido_certified,
    common_criteria_certified, metadata_raw
) VALUES 
-- YubiKey 5 Series
(
    '2fc0579f-8113-47ea-b116-bb5a8db9202a'::UUID,
    'Yubico',
    'YubiKey 5 Series',
    'YK5',
    ARRAY['hardware', 'secure_element'],
    ARRAY['none'],
    128,
    ARRAY['external', 'wired', 'wireless', 'nfc'],
    true,
    false,
    ARRAY[-7, -35, -36, -257, -258, -259],
    'FIDO2_1',
    true,
    false,
    jsonb_build_object(
        'description', 'YubiKey 5 Series FIDO2 authenticator',
        'icon', 'https://www.yubico.com/wp-content/uploads/2019/03/YubiKey-5-Series-500x500.png',
        'supported_transports', ARRAY['usb', 'nfc'],
        'user_verification_methods', ARRAY['presence_internal'],
        'key_protection_types', ARRAY['hardware', 'secure_element'],
        'matcher_protection_types', ARRAY['on_chip'],
        'crypto_strength', 128,
        'operating_env', 'restricted'
    )
),

-- Touch ID (Apple)
(
    'adce0002-35bc-c60a-648b-0b25f1f05503'::UUID,
    'Apple',
    'Touch ID',
    'TouchID',
    ARRAY['hardware', 'tee'],
    ARRAY['tee'],
    256,
    ARRAY['internal'],
    true,
    true,
    ARRAY[-7, -257],
    'FIDO2_1',
    true,
    false,
    jsonb_build_object(
        'description', 'Apple Touch ID built-in authenticator',
        'icon', 'https://developer.apple.com/design/human-interface-guidelines/technologies/touch-id/images/touch-id-intro.png',
        'supported_transports', ARRAY['internal'],
        'user_verification_methods', ARRAY['fingerprint_internal'],
        'key_protection_types', ARRAY['hardware', 'secure_element'],
        'matcher_protection_types', ARRAY['on_chip'],
        'crypto_strength', 256,
        'operating_env', 'restricted'
    )
),

-- Face ID (Apple)
(
    '389c9753-1e30-4c14-b321-dc447d4b5d94'::UUID,
    'Apple',
    'Face ID',
    'FaceID',
    ARRAY['hardware', 'tee'],
    ARRAY['tee'],
    256,
    ARRAY['internal'],
    true,
    true,
    ARRAY[-7, -257],
    'FIDO2_1',
    true,
    false,
    jsonb_build_object(
        'description', 'Apple Face ID built-in authenticator',
        'icon', 'https://developer.apple.com/design/human-interface-guidelines/technologies/face-id/images/face-id-intro.png',
        'supported_transports', ARRAY['internal'],
        'user_verification_methods', ARRAY['faceprint_internal'],
        'key_protection_types', ARRAY['hardware', 'secure_element'],
        'matcher_protection_types', ARRAY['on_chip'],
        'crypto_strength', 256,
        'operating_env', 'restricted'
    )
),

-- Windows Hello
(
    '08987058-cadc-4b81-b6e1-30de50dcbe96'::UUID,
    'Microsoft',
    'Windows Hello',
    'WHfB',
    ARRAY['hardware', 'tpm'],
    ARRAY['tpm'],
    256,
    ARRAY['internal'],
    true,
    true,
    ARRAY[-7, -257],
    'FIDO2_1',
    true,
    false,
    jsonb_build_object(
        'description', 'Windows Hello built-in authenticator',
        'icon', 'https://docs.microsoft.com/en-us/windows/security/identity-protection/hello-for-business/images/hello-face.png',
        'supported_transports', ARRAY['internal'],
        'user_verification_methods', ARRAY['fingerprint_internal', 'faceprint_internal', 'passcode_internal'],
        'key_protection_types', ARRAY['hardware', 'tpm'],
        'matcher_protection_types', ARRAY['on_chip'],
        'crypto_strength', 256,
        'operating_env', 'restricted'
    )
),

-- Android Fingerprint
(
    'bada5566-a7aa-401f-bd96-45619a55120d'::UUID,
    'Google',
    'Android Fingerprint',
    'AndroidFP',
    ARRAY['hardware', 'tee'],
    ARRAY['tee'],
    256,
    ARRAY['internal'],
    true,
    true,
    ARRAY[-7, -257],
    'FIDO2_1',
    true,
    false,
    jsonb_build_object(
        'description', 'Android built-in fingerprint authenticator',
        'supported_transports', ARRAY['internal'],
        'user_verification_methods', ARRAY['fingerprint_internal'],
        'key_protection_types', ARRAY['hardware', 'tee'],
        'matcher_protection_types', ARRAY['on_chip'],
        'crypto_strength', 256,
        'operating_env', 'restricted'
    )
),

-- Autenticador de Teste (Desenvolvimento)
(
    '00000000-0000-0000-0000-000000000000'::UUID,
    'INNOVABIZ',
    'Test Authenticator',
    'TestAuth',
    ARRAY['software'],
    ARRAY['software'],
    128,
    ARRAY['internal'],
    false,
    false,
    ARRAY[-7, -257],
    'NONE',
    false,
    false,
    jsonb_build_object(
        'description', 'Autenticador de teste para desenvolvimento',
        'environment', 'development',
        'supported_transports', ARRAY['internal'],
        'user_verification_methods', ARRAY['none'],
        'key_protection_types', ARRAY['software'],
        'matcher_protection_types', ARRAY['software'],
        'crypto_strength', 128,
        'operating_env', 'unrestricted'
    )
);

-- ============================================================================
-- CONFIGURAÇÕES REGIONAIS E DE COMPLIANCE
-- ============================================================================

-- Tabela para configurações regionais de WebAuthn
CREATE TABLE IF NOT EXISTS webauthn_regional_config (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    region_code TEXT NOT NULL UNIQUE,
    region_name TEXT NOT NULL,
    
    -- Configurações de compliance
    required_aal authentication_assurance_level NOT NULL DEFAULT 'AAL2',
    allow_platform_authenticators BOOLEAN DEFAULT true,
    allow_cross_platform_authenticators BOOLEAN DEFAULT true,
    require_user_verification BOOLEAN DEFAULT true,
    require_resident_key BOOLEAN DEFAULT false,
    
    -- Configurações de attestation
    attestation_requirement TEXT DEFAULT 'indirect' CHECK (attestation_requirement IN ('none', 'indirect', 'direct', 'enterprise')),
    trusted_attestation_formats TEXT[] DEFAULT ARRAY['packed', 'tpm', 'android-key', 'apple'],
    
    -- Configurações de timeout
    registration_timeout_ms INTEGER DEFAULT 60000,
    authentication_timeout_ms INTEGER DEFAULT 60000,
    
    -- Configurações de segurança
    max_credentials_per_user INTEGER DEFAULT 10,
    credential_expiry_days INTEGER, -- NULL = sem expiração
    require_fresh_registration_days INTEGER DEFAULT 365,
    
    -- Configurações regulatórias
    regulatory_framework TEXT[],
    data_retention_days INTEGER DEFAULT 2555, -- 7 anos
    audit_level TEXT DEFAULT 'standard' CHECK (audit_level IN ('minimal', 'standard', 'enhanced')),
    
    -- Metadados
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

-- Inserir configurações regionais
INSERT INTO webauthn_regional_config (
    region_code, region_name, required_aal, require_user_verification,
    attestation_requirement, regulatory_framework, data_retention_days,
    audit_level, metadata
) VALUES 
-- Brasil
(
    'BR',
    'Brasil',
    'AAL2',
    true,
    'indirect',
    ARRAY['LGPD', 'BACEN', 'PIX'],
    2555, -- 7 anos conforme LGPD
    'enhanced',
    jsonb_build_object(
        'currency', 'BRL',
        'language', 'pt-BR',
        'timezone', 'America/Sao_Paulo',
        'regulatory_authority', 'BACEN',
        'privacy_law', 'LGPD',
        'strong_auth_required', true,
        'biometric_consent_required', true
    )
),

-- Estados Unidos
(
    'US',
    'United States',
    'AAL3',
    true,
    'direct',
    ARRAY['NIST', 'SOX', 'CCPA'],
    2555, -- 7 anos
    'enhanced',
    jsonb_build_object(
        'currency', 'USD',
        'language', 'en-US',
        'timezone', 'America/New_York',
        'regulatory_authority', 'NIST',
        'privacy_law', 'CCPA',
        'strong_auth_required', true,
        'federal_compliance', true
    )
),

-- União Europeia
(
    'EU',
    'European Union',
    'AAL2',
    true,
    'indirect',
    ARRAY['GDPR', 'PSD2', 'eIDAS'],
    2555, -- 7 anos
    'enhanced',
    jsonb_build_object(
        'currency', 'EUR',
        'language', 'en-GB',
        'timezone', 'Europe/London',
        'regulatory_authority', 'EBA',
        'privacy_law', 'GDPR',
        'strong_auth_required', true,
        'psd2_compliance', true,
        'right_to_be_forgotten', true
    )
),

-- Angola
(
    'AO',
    'Angola',
    'AAL2',
    true,
    'indirect',
    ARRAY['BNA', 'Lei_22_2011'],
    2190, -- 6 anos
    'standard',
    jsonb_build_object(
        'currency', 'AOA',
        'language', 'pt-AO',
        'timezone', 'Africa/Luanda',
        'regulatory_authority', 'BNA',
        'privacy_law', 'Lei 22/2011',
        'strong_auth_required', false,
        'developing_market', true
    )
);

-- ============================================================================
-- DADOS DE TESTE PARA DESENVOLVIMENTO
-- ============================================================================

-- Inserir usuários de teste (apenas em ambiente de desenvolvimento)
DO $$
BEGIN
    IF current_setting('app.environment', true) = 'development' THEN
        
        -- Credencial de teste para usuário demo
        INSERT INTO webauthn_credentials (
            credential_id,
            user_id,
            tenant_id,
            public_key,
            sign_count,
            aaguid,
            attestation_format,
            attestation_data,
            user_verified,
            backup_eligible,
            backup_state,
            transports,
            authenticator_type,
            device_type,
            friendly_name,
            status,
            compliance_level,
            risk_score,
            metadata
        ) VALUES (
            'test-credential-demo-user-001',
            '11111111-1111-1111-1111-111111111111'::UUID,
            '22222222-2222-2222-2222-222222222222'::UUID,
            decode('3059301306072a8648ce3d020106082a8648ce3d03010703420004', 'hex'),
            0,
            '00000000-0000-0000-0000-000000000000'::UUID,
            'none',
            jsonb_build_object(
                'fmt', 'none',
                'attStmt', '{}',
                'test_credential', true
            ),
            false,
            false,
            false,
            ARRAY['internal'],
            'platform',
            'test',
            'Credencial de Teste',
            'active',
            'AAL1',
            0.10,
            jsonb_build_object(
                'environment', 'development',
                'test_credential', true,
                'created_by', 'system',
                'purpose', 'development_testing'
            )
        );
        
        -- Eventos de teste
        INSERT INTO webauthn_authentication_events (
            credential_id,
            user_id,
            tenant_id,
            event_type,
            result,
            client_data,
            user_verified,
            sign_count,
            ip_address,
            user_agent,
            risk_score,
            compliance_level,
            metadata
        ) VALUES (
            (SELECT id FROM webauthn_credentials WHERE credential_id = 'test-credential-demo-user-001'),
            '11111111-1111-1111-1111-111111111111'::UUID,
            '22222222-2222-2222-2222-222222222222'::UUID,
            'registration',
            'success',
            jsonb_build_object(
                'type', 'webauthn.create',
                'origin', 'https://dev.innovabiz.com',
                'crossOrigin', false
            ),
            false,
            0,
            '127.0.0.1'::INET,
            'Mozilla/5.0 (Test Browser)',
            0.10,
            'AAL1',
            jsonb_build_object(
                'test_event', true,
                'environment', 'development'
            )
        );
        
        RAISE NOTICE 'Dados de teste inseridos para ambiente de desenvolvimento';
    END IF;
END $$;

-- ============================================================================
-- CONFIGURAÇÕES DE SISTEMA
-- ============================================================================

-- Tabela para configurações globais do sistema WebAuthn
CREATE TABLE IF NOT EXISTS webauthn_system_config (
    key TEXT PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    category TEXT DEFAULT 'general',
    is_sensitive BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Inserir configurações do sistema
INSERT INTO webauthn_system_config (key, value, description, category, is_sensitive) VALUES 
('global.max_credentials_per_user', '10', 'Número máximo de credenciais por usuário', 'limits', false),
('global.challenge_timeout_seconds', '300', 'Timeout para challenges em segundos', 'security', false),
('global.max_authentication_attempts_per_minute', '5', 'Máximo de tentativas de autenticação por minuto', 'rate_limiting', false),
('global.enable_sign_count_verification', 'true', 'Habilitar verificação de sign count', 'security', false),
('global.require_attestation_verification', 'true', 'Exigir verificação de attestation', 'security', false),
('global.supported_algorithms', '[-7, -257, -8, -37, -38, -39]', 'Algoritmos criptográficos suportados', 'crypto', false),
('global.default_user_verification', '"preferred"', 'Requisito padrão de verificação de usuário', 'security', false),
('global.default_attestation_conveyance', '"indirect"', 'Conveyance padrão de attestation', 'security', false),
('monitoring.enable_detailed_logging', 'true', 'Habilitar logging detalhado', 'monitoring', false),
('monitoring.log_client_data', 'false', 'Registrar dados do cliente nos logs', 'monitoring', true),
('monitoring.alert_on_sign_count_anomaly', 'true', 'Alertar sobre anomalias de sign count', 'monitoring', false),
('compliance.data_retention_days', '2555', 'Dias de retenção de dados (7 anos)', 'compliance', false),
('compliance.audit_all_operations', 'true', 'Auditar todas as operações', 'compliance', false),
('compliance.encrypt_sensitive_data', 'true', 'Criptografar dados sensíveis', 'compliance', false);

-- ============================================================================
-- TEMPLATES DE CONFIGURAÇÃO POR TIPO DE ORGANIZAÇÃO
-- ============================================================================

-- Tabela para templates de configuração
CREATE TABLE IF NOT EXISTS webauthn_config_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    template_name TEXT NOT NULL UNIQUE,
    organization_type TEXT NOT NULL,
    security_level TEXT NOT NULL CHECK (security_level IN ('basic', 'standard', 'high', 'critical')),
    
    -- Configurações do template
    config JSONB NOT NULL,
    
    -- Metadados
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Templates para diferentes tipos de organização
INSERT INTO webauthn_config_templates (
    template_name, organization_type, security_level, config, description
) VALUES 
-- Instituições Financeiras
(
    'financial_institution_high_security',
    'financial',
    'critical',
    jsonb_build_object(
        'required_aal', 'AAL3',
        'require_user_verification', true,
        'require_resident_key', true,
        'attestation_requirement', 'direct',
        'max_credentials_per_user', 5,
        'credential_expiry_days', 365,
        'enable_sign_count_verification', true,
        'require_attestation_verification', true,
        'supported_authenticator_types', ARRAY['cross-platform'],
        'trusted_aaguids', ARRAY[
            '2fc0579f-8113-47ea-b116-bb5a8db9202a', -- YubiKey 5
            '08987058-cadc-4b81-b6e1-30de50dcbe96'  -- Windows Hello
        ],
        'risk_assessment', jsonb_build_object(
            'enable_ip_geolocation_check', true,
            'enable_device_fingerprinting', true,
            'max_risk_score', 0.3
        )
    ),
    'Configuração de alta segurança para instituições financeiras'
),

-- E-commerce
(
    'ecommerce_standard_security',
    'ecommerce',
    'standard',
    jsonb_build_object(
        'required_aal', 'AAL2',
        'require_user_verification', true,
        'require_resident_key', false,
        'attestation_requirement', 'indirect',
        'max_credentials_per_user', 10,
        'credential_expiry_days', null,
        'enable_sign_count_verification', true,
        'require_attestation_verification', false,
        'supported_authenticator_types', ARRAY['platform', 'cross-platform'],
        'risk_assessment', jsonb_build_object(
            'enable_ip_geolocation_check', true,
            'enable_device_fingerprinting', false,
            'max_risk_score', 0.7
        )
    ),
    'Configuração padrão para plataformas de e-commerce'
),

-- Aplicações Corporativas
(
    'enterprise_application',
    'enterprise',
    'high',
    jsonb_build_object(
        'required_aal', 'AAL2',
        'require_user_verification', true,
        'require_resident_key', true,
        'attestation_requirement', 'indirect',
        'max_credentials_per_user', 3,
        'credential_expiry_days', 180,
        'enable_sign_count_verification', true,
        'require_attestation_verification', true,
        'supported_authenticator_types', ARRAY['platform', 'cross-platform'],
        'trusted_aaguids', ARRAY[
            'adce0002-35bc-c60a-648b-0b25f1f05503', -- Touch ID
            '389c9753-1e30-4c14-b321-dc447d4b5d94', -- Face ID
            '08987058-cadc-4b81-b6e1-30de50dcbe96'  -- Windows Hello
        ]
    ),
    'Configuração para aplicações corporativas'
);

-- ============================================================================
-- FUNÇÕES PARA APLICAR CONFIGURAÇÕES
-- ============================================================================

-- Função para aplicar template de configuração a um tenant
CREATE OR REPLACE FUNCTION apply_webauthn_config_template(
    p_tenant_id UUID,
    p_template_name TEXT
) RETURNS JSONB AS $$
DECLARE
    template_config JSONB;
    result JSONB;
BEGIN
    -- Buscar configuração do template
    SELECT config INTO template_config
    FROM webauthn_config_templates
    WHERE template_name = p_template_name;
    
    IF template_config IS NULL THEN
        RETURN jsonb_build_object(
            'success', false,
            'error', 'Template not found',
            'template_name', p_template_name
        );
    END IF;
    
    -- Aplicar configurações (implementação específica dependeria da estrutura de configuração por tenant)
    -- Por enquanto, apenas retornar a configuração que seria aplicada
    
    SELECT jsonb_build_object(
        'success', true,
        'tenant_id', p_tenant_id,
        'template_name', p_template_name,
        'applied_config', template_config,
        'applied_at', NOW()
    ) INTO result;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- DADOS DE REFERÊNCIA PARA TESTES DE COMPLIANCE
-- ============================================================================

-- Inserir cenários de teste para compliance
INSERT INTO webauthn_authentication_events (
    user_id, tenant_id, event_type, result, 
    client_data, risk_score, compliance_level, metadata
) 
SELECT 
    '99999999-9999-9999-9999-999999999999'::UUID,
    '88888888-8888-8888-8888-888888888888'::UUID,
    'authentication',
    CASE WHEN random() > 0.1 THEN 'success' ELSE 'failure' END,
    jsonb_build_object(
        'type', 'webauthn.get',
        'origin', 'https://test.innovabiz.com',
        'test_scenario', true
    ),
    random(),
    (ARRAY['AAL1', 'AAL2', 'AAL3'])[floor(random() * 3 + 1)],
    jsonb_build_object(
        'test_data', true,
        'scenario', 'compliance_testing',
        'batch', 'initial_data'
    )
FROM generate_series(1, 100)
WHERE current_setting('app.environment', true) = 'development';

-- ============================================================================
-- GRANTS E PERMISSÕES
-- ============================================================================

-- Permissões para tabelas de configuração
GRANT SELECT ON webauthn_regional_config TO webauthn_app;
GRANT SELECT ON webauthn_system_config TO webauthn_app;
GRANT SELECT ON webauthn_config_templates TO webauthn_app;

-- Permissões para funções
GRANT EXECUTE ON FUNCTION apply_webauthn_config_template(UUID, TEXT) TO webauthn_app;

-- ============================================================================
-- LOG DE INICIALIZAÇÃO
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'INNOVABIZ WebAuthn Initial Data v1.0.0 carregado com sucesso em %', NOW();
    RAISE NOTICE 'Dados inseridos:';
    RAISE NOTICE '- % metadados de autenticadores FIDO', (SELECT COUNT(*) FROM webauthn_authenticator_metadata);
    RAISE NOTICE '- % configurações regionais', (SELECT COUNT(*) FROM webauthn_regional_config);
    RAISE NOTICE '- % configurações de sistema', (SELECT COUNT(*) FROM webauthn_system_config);
    RAISE NOTICE '- % templates de configuração', (SELECT COUNT(*) FROM webauthn_config_templates);
    
    IF current_setting('app.environment', true) = 'development' THEN
        RAISE NOTICE '- Dados de teste inseridos para ambiente de desenvolvimento';
    END IF;
END $$;

-- ============================================================================
-- FIM DO SCRIPT DE DADOS INICIAIS
-- ============================================================================