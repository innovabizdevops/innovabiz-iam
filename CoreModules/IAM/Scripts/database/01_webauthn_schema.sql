-- ============================================================================
-- INNOVABIZ IAM - WebAuthn/FIDO2 Database Schema
-- ============================================================================
-- Version: 1.0.0
-- Date: 31/07/2025
-- Author: Equipe de Segurança INNOVABIZ
-- Classification: Confidencial - Interno
-- 
-- Description: Schema principal para suporte a credenciais WebAuthn/FIDO2
-- Standards: W3C WebAuthn Level 3, FIDO2 CTAP2.1, NIST SP 800-63B
-- Compliance: PCI DSS 4.0, GDPR/LGPD, PSD2, ISO 27001
-- ============================================================================

-- Habilitar extensões necessárias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- ============================================================================
-- ENUMS E TIPOS PERSONALIZADOS
-- ============================================================================

-- Tipos de autenticadores
CREATE TYPE authenticator_type AS ENUM (
    'platform',           -- Autenticadores de plataforma (TouchID, FaceID, Windows Hello)
    'cross-platform'       -- Autenticadores roaming (YubiKey, chaves de segurança)
);

-- Status de credenciais
CREATE TYPE credential_status AS ENUM (
    'active',              -- Credencial ativa e utilizável
    'suspended',           -- Temporariamente suspensa
    'revoked',             -- Permanentemente revogada
    'expired'              -- Expirada (se aplicável)
);

-- Formatos de attestation
CREATE TYPE attestation_format AS ENUM (
    'packed',              -- Formato packed (padrão)
    'tpm',                 -- TPM-based attestation
    'android-key',         -- Android Key Attestation
    'android-safetynet',   -- Android SafetyNet
    'fido-u2f',           -- Legacy FIDO U2F
    'apple',              -- Apple Anonymous Attestation
    'none'                -- Sem attestation
);

-- Níveis de garantia de autenticação (NIST AAL)
CREATE TYPE authentication_assurance_level AS ENUM (
    'AAL1',               -- Single-factor authentication
    'AAL2',               -- Multi-factor authentication
    'AAL3'                -- Hardware-based multi-factor authentication
);

-- Tipos de eventos de autenticação
CREATE TYPE webauthn_event_type AS ENUM (
    'registration',        -- Registro de nova credencial
    'authentication',      -- Autenticação bem-sucedida
    'authentication_failed', -- Tentativa de autenticação falhada
    'credential_suspended', -- Credencial suspensa
    'credential_revoked',   -- Credencial revogada
    'sign_count_anomaly'   -- Anomalia no contador de assinatura
);

-- Resultado de eventos
CREATE TYPE event_result AS ENUM (
    'success',
    'failure',
    'warning'
);

-- ============================================================================
-- TABELA PRINCIPAL DE CREDENCIAIS WEBAUTHN
-- ============================================================================

CREATE TABLE webauthn_credentials (
    -- Identificadores únicos
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    credential_id TEXT NOT NULL UNIQUE, -- Base64URL encoded credential ID
    
    -- Relacionamentos
    user_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    
    -- Dados criptográficos
    public_key BYTEA NOT NULL,           -- Chave pública do autenticador
    sign_count BIGINT NOT NULL DEFAULT 0, -- Contador de assinatura (anti-clonagem)
    
    -- Metadados do autenticador
    aaguid UUID,                         -- Authenticator Attestation GUID
    attestation_format attestation_format NOT NULL DEFAULT 'none',
    attestation_data JSONB,              -- Dados completos de attestation
    
    -- Características de segurança
    user_verified BOOLEAN NOT NULL DEFAULT false,     -- Suporte a verificação de usuário
    backup_eligible BOOLEAN NOT NULL DEFAULT false,   -- Elegível para backup
    backup_state BOOLEAN NOT NULL DEFAULT false,      -- Estado atual de backup
    
    -- Configuração de transporte
    transports TEXT[] DEFAULT '{}',      -- Métodos de transporte suportados
    
    -- Classificação e configuração
    authenticator_type authenticator_type NOT NULL,
    device_type TEXT,                    -- Tipo específico do dispositivo
    friendly_name TEXT,                  -- Nome amigável definido pelo usuário
    
    -- Gestão de ciclo de vida
    status credential_status NOT NULL DEFAULT 'active',
    suspension_reason TEXT,              -- Motivo da suspensão/revogação
    
    -- Metadados de compliance
    compliance_level authentication_assurance_level NOT NULL DEFAULT 'AAL2',
    risk_score DECIMAL(3,2) DEFAULT 0.00, -- Score de risco (0.00-1.00)
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE, -- Expiração opcional
    
    -- Auditoria
    created_by UUID,
    updated_by UUID,
    
    -- Metadados adicionais
    metadata JSONB DEFAULT '{}',
    
    -- Constraints
    CONSTRAINT chk_sign_count_positive CHECK (sign_count >= 0),
    CONSTRAINT chk_risk_score_range CHECK (risk_score >= 0.00 AND risk_score <= 1.00),
    CONSTRAINT chk_friendly_name_length CHECK (char_length(friendly_name) <= 100),
    CONSTRAINT chk_suspension_reason CHECK (
        (status IN ('suspended', 'revoked') AND suspension_reason IS NOT NULL) OR
        (status NOT IN ('suspended', 'revoked'))
    )
);

-- Comentários para documentação
COMMENT ON TABLE webauthn_credentials IS 'Armazena credenciais WebAuthn/FIDO2 registradas pelos usuários';
COMMENT ON COLUMN webauthn_credentials.credential_id IS 'Identificador único da credencial (Base64URL)';
COMMENT ON COLUMN webauthn_credentials.public_key IS 'Chave pública COSE do autenticador';
COMMENT ON COLUMN webauthn_credentials.sign_count IS 'Contador para detecção de clonagem';
COMMENT ON COLUMN webauthn_credentials.aaguid IS 'GUID de attestation do autenticador';
COMMENT ON COLUMN webauthn_credentials.compliance_level IS 'Nível AAL conforme NIST SP 800-63B';

-- ============================================================================
-- TABELA DE EVENTOS DE AUTENTICAÇÃO
-- ============================================================================

CREATE TABLE webauthn_authentication_events (
    -- Identificador único
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Relacionamentos
    credential_id UUID REFERENCES webauthn_credentials(id) ON DELETE SET NULL,
    user_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    
    -- Tipo e resultado do evento
    event_type webauthn_event_type NOT NULL,
    result event_result NOT NULL,
    
    -- Dados da autenticação
    client_data JSONB,               -- clientDataJSON decodificado
    authenticator_data BYTEA,        -- authenticatorData raw
    signature BYTEA,                 -- Assinatura da autenticação
    
    -- Verificações de segurança
    user_verified BOOLEAN,           -- Se o usuário foi verificado
    sign_count BIGINT,              -- Contador no momento da autenticação
    origin_verified BOOLEAN DEFAULT true, -- Se a origem foi verificada
    
    -- Contexto da requisição
    ip_address INET,                 -- Endereço IP do cliente
    user_agent TEXT,                 -- User agent do navegador
    geolocation JSONB,               -- Localização geográfica (se disponível)
    
    -- Avaliação de risco
    risk_score DECIMAL(3,2),         -- Score de risco calculado
    risk_factors JSONB,              -- Fatores que contribuíram para o risco
    
    -- Detalhes de falha (se aplicável)
    error_code TEXT,                 -- Código de erro específico
    error_message TEXT,              -- Mensagem de erro
    error_details JSONB,             -- Detalhes técnicos do erro
    
    -- Metadados de compliance
    compliance_level authentication_assurance_level,
    regulatory_context JSONB,        -- Contexto regulatório aplicável
    
    -- Timestamp
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Metadados adicionais
    metadata JSONB DEFAULT '{}',
    
    -- Constraints
    CONSTRAINT chk_risk_score_range CHECK (risk_score >= 0.00 AND risk_score <= 1.00)
) PARTITION BY RANGE (created_at);

-- Comentários
COMMENT ON TABLE webauthn_authentication_events IS 'Log de todos os eventos relacionados a WebAuthn';
COMMENT ON COLUMN webauthn_authentication_events.client_data IS 'Dados do cliente decodificados do clientDataJSON';
COMMENT ON COLUMN webauthn_authentication_events.risk_score IS 'Score de risco calculado (0.00-1.00)';

-- ============================================================================
-- PARTIÇÕES PARA EVENTOS (PERFORMANCE E RETENÇÃO)
-- ============================================================================

-- Partição para julho 2025
CREATE TABLE webauthn_authentication_events_2025_07 PARTITION OF webauthn_authentication_events
FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');

-- Partição para agosto 2025
CREATE TABLE webauthn_authentication_events_2025_08 PARTITION OF webauthn_authentication_events
FOR VALUES FROM ('2025-08-01') TO ('2025-09-01');

-- Partição para setembro 2025
CREATE TABLE webauthn_authentication_events_2025_09 PARTITION OF webauthn_authentication_events
FOR VALUES FROM ('2025-09-01') TO ('2025-10-01');

-- ============================================================================
-- TABELA DE CHALLENGES TEMPORÁRIOS
-- ============================================================================

CREATE TABLE webauthn_challenges (
    -- Identificador único
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Chave de identificação
    challenge_key TEXT NOT NULL,      -- user_id:tenant_id ou session_id
    challenge_type TEXT NOT NULL,     -- 'registration' ou 'authentication'
    
    -- Dados do challenge
    challenge TEXT NOT NULL,          -- Challenge Base64URL
    
    -- Contexto
    user_id UUID,
    tenant_id UUID,
    session_id TEXT,
    
    -- Metadados da requisição
    origin TEXT NOT NULL,
    user_agent TEXT,
    ip_address INET,
    
    -- Gestão de ciclo de vida
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() + INTERVAL '5 minutes'),
    used_at TIMESTAMP WITH TIME ZONE,
    
    -- Status
    is_used BOOLEAN DEFAULT false,
    
    -- Constraints
    CONSTRAINT chk_challenge_type CHECK (challenge_type IN ('registration', 'authentication')),
    CONSTRAINT chk_not_expired CHECK (expires_at > created_at)
);

-- Comentários
COMMENT ON TABLE webauthn_challenges IS 'Armazena challenges temporários para operações WebAuthn';
COMMENT ON COLUMN webauthn_challenges.challenge_key IS 'Chave para identificar o challenge';

-- ============================================================================
-- TABELA DE METADADOS DE AUTENTICADORES (FIDO METADATA)
-- ============================================================================

CREATE TABLE webauthn_authenticator_metadata (
    -- Identificador único
    aaguid UUID PRIMARY KEY,
    
    -- Informações do fabricante
    vendor_name TEXT,
    device_name TEXT,
    device_model TEXT,
    
    -- Características de segurança
    key_protection TEXT[],            -- Métodos de proteção de chave
    matcher_protection TEXT[],        -- Proteção do matcher biométrico
    crypto_strength INTEGER,          -- Força criptográfica
    attachment_hint TEXT[],           -- Dicas de anexação
    
    -- Capacidades
    is_key_restricted BOOLEAN,        -- Se as chaves são restritas
    is_fresh_user_verification_required BOOLEAN,
    supported_algorithms INTEGER[],   -- Algoritmos suportados (COSE)
    
    -- Certificação
    certification_level TEXT,        -- Nível de certificação FIDO
    certification_date DATE,
    
    -- Metadados de compliance
    fido_certified BOOLEAN DEFAULT false,
    common_criteria_certified BOOLEAN DEFAULT false,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Dados raw do metadata
    metadata_raw JSONB
);

-- Comentários
COMMENT ON TABLE webauthn_authenticator_metadata IS 'Metadados de autenticadores conforme FIDO Metadata Service';

-- ============================================================================
-- ÍNDICES PARA PERFORMANCE
-- ============================================================================

-- Índices principais para credenciais
CREATE INDEX idx_webauthn_credentials_user_tenant ON webauthn_credentials(user_id, tenant_id);
CREATE INDEX idx_webauthn_credentials_credential_id ON webauthn_credentials(credential_id);
CREATE INDEX idx_webauthn_credentials_status_active ON webauthn_credentials(status) WHERE status = 'active';
CREATE INDEX idx_webauthn_credentials_tenant_status ON webauthn_credentials(tenant_id, status);
CREATE INDEX idx_webauthn_credentials_last_used ON webauthn_credentials(last_used_at DESC) WHERE status = 'active';
CREATE INDEX idx_webauthn_credentials_aaguid ON webauthn_credentials(aaguid) WHERE aaguid IS NOT NULL;

-- Índices para eventos
CREATE INDEX idx_webauthn_events_user_tenant_date ON webauthn_authentication_events(user_id, tenant_id, created_at DESC);
CREATE INDEX idx_webauthn_events_credential_date ON webauthn_authentication_events(credential_id, created_at DESC);
CREATE INDEX idx_webauthn_events_type_result ON webauthn_authentication_events(event_type, result);
CREATE INDEX idx_webauthn_events_ip_date ON webauthn_authentication_events(ip_address, created_at DESC);
CREATE INDEX idx_webauthn_events_risk_score ON webauthn_authentication_events(risk_score DESC) WHERE risk_score > 0.5;

-- Índices para challenges
CREATE INDEX idx_webauthn_challenges_key_type ON webauthn_challenges(challenge_key, challenge_type);
CREATE INDEX idx_webauthn_challenges_expires ON webauthn_challenges(expires_at) WHERE NOT is_used;
CREATE INDEX idx_webauthn_challenges_cleanup ON webauthn_challenges(created_at) WHERE is_used OR expires_at < NOW();

-- ============================================================================
-- ROW LEVEL SECURITY (RLS) PARA MULTI-TENANT
-- ============================================================================

-- Habilitar RLS nas tabelas principais
ALTER TABLE webauthn_credentials ENABLE ROW LEVEL SECURITY;
ALTER TABLE webauthn_authentication_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE webauthn_challenges ENABLE ROW LEVEL SECURITY;

-- Política para credenciais (isolamento por tenant)
CREATE POLICY webauthn_credentials_tenant_isolation ON webauthn_credentials
    FOR ALL
    TO authenticated_users
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- Política para eventos (isolamento por tenant)
CREATE POLICY webauthn_events_tenant_isolation ON webauthn_authentication_events
    FOR ALL
    TO authenticated_users
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- Política para challenges (isolamento por tenant)
CREATE POLICY webauthn_challenges_tenant_isolation ON webauthn_challenges
    FOR ALL
    TO authenticated_users
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID OR tenant_id IS NULL);

-- ============================================================================
-- TRIGGERS PARA AUDITORIA E MANUTENÇÃO
-- ============================================================================

-- Função para atualizar timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger para atualizar updated_at em credenciais
CREATE TRIGGER trigger_webauthn_credentials_updated_at
    BEFORE UPDATE ON webauthn_credentials
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Função para limpeza automática de challenges expirados
CREATE OR REPLACE FUNCTION cleanup_expired_challenges()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM webauthn_challenges 
    WHERE expires_at < NOW() - INTERVAL '1 hour';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    -- Log da limpeza
    INSERT INTO webauthn_authentication_events (
        user_id, tenant_id, event_type, result, error_message, metadata
    ) VALUES (
        '00000000-0000-0000-0000-000000000000'::UUID,
        '00000000-0000-0000-0000-000000000000'::UUID,
        'authentication'::webauthn_event_type,
        'success'::event_result,
        'Cleanup expired challenges',
        jsonb_build_object('deleted_count', deleted_count, 'operation', 'cleanup_challenges')
    );
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- FUNÇÕES DE UTILIDADE
-- ============================================================================

-- Função para obter estatísticas de credenciais por tenant
CREATE OR REPLACE FUNCTION get_webauthn_stats(p_tenant_id UUID)
RETURNS JSONB AS $$
DECLARE
    result JSONB;
BEGIN
    SELECT jsonb_build_object(
        'total_credentials', COUNT(*),
        'active_credentials', COUNT(*) FILTER (WHERE status = 'active'),
        'platform_authenticators', COUNT(*) FILTER (WHERE authenticator_type = 'platform'),
        'cross_platform_authenticators', COUNT(*) FILTER (WHERE authenticator_type = 'cross-platform'),
        'aal2_credentials', COUNT(*) FILTER (WHERE compliance_level = 'AAL2'),
        'aal3_credentials', COUNT(*) FILTER (WHERE compliance_level = 'AAL3'),
        'last_30_days_usage', COUNT(*) FILTER (WHERE last_used_at > NOW() - INTERVAL '30 days')
    ) INTO result
    FROM webauthn_credentials
    WHERE tenant_id = p_tenant_id;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- Função para verificar saúde do sistema WebAuthn
CREATE OR REPLACE FUNCTION webauthn_health_check()
RETURNS JSONB AS $$
DECLARE
    result JSONB;
    expired_challenges INTEGER;
    recent_failures INTEGER;
BEGIN
    -- Contar challenges expirados
    SELECT COUNT(*) INTO expired_challenges
    FROM webauthn_challenges
    WHERE expires_at < NOW() AND NOT is_used;
    
    -- Contar falhas recentes
    SELECT COUNT(*) INTO recent_failures
    FROM webauthn_authentication_events
    WHERE created_at > NOW() - INTERVAL '1 hour'
    AND result = 'failure';
    
    SELECT jsonb_build_object(
        'status', CASE 
            WHEN expired_challenges > 1000 OR recent_failures > 100 THEN 'unhealthy'
            WHEN expired_challenges > 500 OR recent_failures > 50 THEN 'degraded'
            ELSE 'healthy'
        END,
        'expired_challenges', expired_challenges,
        'recent_failures', recent_failures,
        'timestamp', NOW()
    ) INTO result;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- GRANTS E PERMISSÕES
-- ============================================================================

-- Criar role para aplicação WebAuthn
CREATE ROLE webauthn_app;

-- Permissões para tabelas principais
GRANT SELECT, INSERT, UPDATE, DELETE ON webauthn_credentials TO webauthn_app;
GRANT SELECT, INSERT ON webauthn_authentication_events TO webauthn_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON webauthn_challenges TO webauthn_app;
GRANT SELECT ON webauthn_authenticator_metadata TO webauthn_app;

-- Permissões para sequências
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO webauthn_app;

-- Permissões para funções
GRANT EXECUTE ON FUNCTION get_webauthn_stats(UUID) TO webauthn_app;
GRANT EXECUTE ON FUNCTION webauthn_health_check() TO webauthn_app;
GRANT EXECUTE ON FUNCTION cleanup_expired_challenges() TO webauthn_app;

-- ============================================================================
-- CONFIGURAÇÕES DE PERFORMANCE
-- ============================================================================

-- Configurar autovacuum para tabelas de alta rotatividade
ALTER TABLE webauthn_challenges SET (
    autovacuum_vacuum_scale_factor = 0.1,
    autovacuum_analyze_scale_factor = 0.05
);

ALTER TABLE webauthn_authentication_events SET (
    autovacuum_vacuum_scale_factor = 0.2,
    autovacuum_analyze_scale_factor = 0.1
);

-- ============================================================================
-- COMENTÁRIOS FINAIS E METADADOS
-- ============================================================================

-- Inserir metadados do schema
INSERT INTO webauthn_authenticator_metadata (
    aaguid, vendor_name, device_name, metadata_raw
) VALUES (
    '00000000-0000-0000-0000-000000000000'::UUID,
    'INNOVABIZ',
    'Test Authenticator',
    '{"description": "Autenticador de teste para desenvolvimento", "version": "1.0.0"}'::JSONB
) ON CONFLICT (aaguid) DO NOTHING;

-- Log de criação do schema
DO $$
BEGIN
    RAISE NOTICE 'INNOVABIZ WebAuthn/FIDO2 Schema v1.0.0 criado com sucesso em %', NOW();
    RAISE NOTICE 'Compliance: W3C WebAuthn Level 3, FIDO2 CTAP2.1, NIST SP 800-63B';
    RAISE NOTICE 'Segurança: PCI DSS 4.0, GDPR/LGPD, Multi-tenant RLS habilitado';
    RAISE NOTICE 'Performance: Índices otimizados, particionamento por data, autovacuum configurado';
END $$;

-- ============================================================================
-- FIM DO SCRIPT
-- ============================================================================