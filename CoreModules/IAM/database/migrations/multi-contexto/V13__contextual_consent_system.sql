-- Migration: V13__contextual_consent_system.sql
-- Descrição: Sistema de consentimento contextual para o microserviço de identidade multi-contexto
-- Autor: INNOVABIZ DevOps
-- Data: 20-08-2025

-- Tabela de definição de finalidades de consentimento (consent purposes)
CREATE TABLE IF NOT EXISTS consent_purposes (
    purpose_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purpose_code VARCHAR(100) NOT NULL,
    purpose_name VARCHAR(200) NOT NULL,
    purpose_description TEXT NOT NULL,
    context_type VARCHAR(50) NOT NULL, -- financial, health, government, etc.
    legal_basis VARCHAR(50) NOT NULL,
    risk_level VARCHAR(20) NOT NULL DEFAULT 'MEDIUM',
    requires_explicit_consent BOOLEAN DEFAULT TRUE,
    retention_period INTERVAL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,
    CONSTRAINT chk_legal_basis CHECK (legal_basis IN ('CONSENT', 'CONTRACT', 'LEGAL_OBLIGATION', 'VITAL_INTEREST', 'PUBLIC_TASK', 'LEGITIMATE_INTEREST')),
    CONSTRAINT chk_purpose_risk_level CHECK (risk_level IN ('VERY_LOW', 'LOW', 'MEDIUM', 'HIGH', 'VERY_HIGH')),
    UNIQUE(purpose_code, context_type)
);

-- Tabela de consentimentos contextuais
CREATE TABLE IF NOT EXISTS context_consents (
    consent_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identity_id UUID NOT NULL REFERENCES identities(identity_id) ON DELETE CASCADE,
    context_id UUID NOT NULL REFERENCES identity_contexts(context_id) ON DELETE CASCADE,
    purpose_id UUID NOT NULL REFERENCES consent_purposes(purpose_id),
    third_party_id UUID, -- NULL se consentimento é para uso interno
    third_party_name VARCHAR(200), -- Nome da terceira parte quando relevante
    granted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    last_confirmed_at TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    consent_proof TEXT, -- Evidência do consentimento (hash, token, etc)
    consent_method VARCHAR(50) NOT NULL, -- Método pelo qual o consentimento foi obtido
    ip_address VARCHAR(50), -- Endereço IP de onde o consentimento foi dado
    user_agent TEXT, -- User agent do dispositivo utilizado
    geolocation VARCHAR(100), -- Localização geográfica
    revocation_reason TEXT,
    revoked_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,
    CONSTRAINT chk_consent_status CHECK (status IN ('ACTIVE', 'REVOKED', 'EXPIRED', 'PENDING')),
    CONSTRAINT chk_consent_method CHECK (consent_method IN ('WEB_FORM', 'MOBILE_APP', 'API', 'VERBAL', 'WRITTEN', 'SMS', 'EMAIL', 'BIOMETRIC')),
    UNIQUE(identity_id, context_id, purpose_id, third_party_id)
);

-- Tabela de escopos de consentimento - quais atributos específicos foram permitidos
CREATE TABLE IF NOT EXISTS consent_scopes (
    scope_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    consent_id UUID NOT NULL REFERENCES context_consents(consent_id) ON DELETE CASCADE,
    attribute_key VARCHAR(100) NOT NULL,
    access_level VARCHAR(20) NOT NULL DEFAULT 'READ',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_access_level CHECK (access_level IN ('READ', 'WRITE', 'READ_WRITE', 'DELEGATE')),
    UNIQUE(consent_id, attribute_key)
);

-- Tabela de histórico de consentimento
CREATE TABLE IF NOT EXISTS consent_history (
    history_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    consent_id UUID NOT NULL REFERENCES context_consents(consent_id) ON DELETE CASCADE,
    action_type VARCHAR(50) NOT NULL,
    previous_status VARCHAR(20),
    new_status VARCHAR(20),
    action_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    action_by UUID,
    action_reason TEXT,
    action_source VARCHAR(100),
    ip_address VARCHAR(50),
    user_agent TEXT,
    CONSTRAINT chk_action_type CHECK (action_type IN ('GRANT', 'REVOKE', 'UPDATE', 'EXPIRE', 'RENEW', 'CONFIRM'))
);

-- Tabela de regras de propagação de consentimento entre contextos
CREATE TABLE IF NOT EXISTS consent_propagation_rules (
    rule_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_context_type VARCHAR(50) NOT NULL,
    target_context_type VARCHAR(50) NOT NULL,
    purpose_code VARCHAR(100) NOT NULL,
    propagation_type VARCHAR(20) NOT NULL DEFAULT 'SUGGEST',
    conditions JSONB, -- Condições para a propagação acontecer
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    is_active BOOLEAN DEFAULT TRUE,
    CONSTRAINT chk_propagation_type CHECK (propagation_type IN ('AUTOMATIC', 'SUGGEST', 'REQUIRE')),
    CONSTRAINT chk_different_contexts_propagation CHECK (source_context_type <> target_context_type),
    UNIQUE(source_context_type, target_context_type, purpose_code)
);

-- Tabela de uso de dados baseado em consentimento
CREATE TABLE IF NOT EXISTS consent_data_access (
    access_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    consent_id UUID NOT NULL REFERENCES context_consents(consent_id),
    accessed_by VARCHAR(100) NOT NULL, -- sistema ou usuário que acessou
    access_purpose VARCHAR(200) NOT NULL,
    attributes_accessed JSONB NOT NULL, -- quais atributos foram acessados
    access_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    access_result VARCHAR(50) NOT NULL, -- GRANTED, DENIED, etc
    request_id VARCHAR(100), -- ID de rastreamento da solicitação
    ip_address VARCHAR(50),
    CONSTRAINT chk_access_result CHECK (access_result IN ('GRANTED', 'DENIED', 'PARTIAL', 'ERROR'))
);

-- Tabela de políticas de consentimento por contexto
CREATE TABLE IF NOT EXISTS context_consent_policies (
    policy_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    context_type VARCHAR(50) NOT NULL,
    region_code VARCHAR(10),
    country_code VARCHAR(2),
    legal_framework VARCHAR(50),
    min_age INT CHECK (min_age > 0),
    requires_explicit_confirmation BOOLEAN DEFAULT TRUE,
    max_consent_duration INTERVAL,
    requires_periodic_renewal BOOLEAN DEFAULT FALSE,
    renewal_period INTERVAL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    UNIQUE(context_type, region_code, country_code, legal_framework)
);

-- Índices para otimização de consultas
CREATE INDEX IF NOT EXISTS idx_consent_purposes_context ON consent_purposes(context_type);
CREATE INDEX IF NOT EXISTS idx_context_consents_identity ON context_consents(identity_id);
CREATE INDEX IF NOT EXISTS idx_context_consents_context ON context_consents(context_id);
CREATE INDEX IF NOT EXISTS idx_context_consents_purpose ON context_consents(purpose_id);
CREATE INDEX IF NOT EXISTS idx_context_consents_status ON context_consents(status);
CREATE INDEX IF NOT EXISTS idx_consent_scopes_consent ON consent_scopes(consent_id);
CREATE INDEX IF NOT EXISTS idx_consent_history_consent ON consent_history(consent_id);
CREATE INDEX IF NOT EXISTS idx_consent_history_timestamp ON consent_history(action_timestamp);
CREATE INDEX IF NOT EXISTS idx_consent_data_access_consent ON consent_data_access(consent_id);
CREATE INDEX IF NOT EXISTS idx_consent_data_access_timestamp ON consent_data_access(access_timestamp);
CREATE INDEX IF NOT EXISTS idx_context_consent_policies_context ON context_consent_policies(context_type);

-- Comentários para documentação do esquema
COMMENT ON TABLE consent_purposes IS 'Finalidades definidas para consentimento em cada contexto';
COMMENT ON TABLE context_consents IS 'Consentimentos dados por identidades em contextos específicos';
COMMENT ON TABLE consent_scopes IS 'Escopos específicos (atributos) incluídos em cada consentimento';
COMMENT ON TABLE consent_history IS 'Histórico de ações sobre consentimentos (concessão, revogação, etc.)';
COMMENT ON TABLE consent_propagation_rules IS 'Regras para propagação de consentimentos entre diferentes contextos';
COMMENT ON TABLE consent_data_access IS 'Registro de acessos a dados baseados em consentimentos';
COMMENT ON TABLE context_consent_policies IS 'Políticas de consentimento específicas por contexto e região';

-- Triggers para atualização automática de timestamps
CREATE TRIGGER update_consent_purposes_timestamp
BEFORE UPDATE ON consent_purposes
FOR EACH ROW EXECUTE PROCEDURE update_timestamp();

CREATE TRIGGER update_context_consents_timestamp
BEFORE UPDATE ON context_consents
FOR EACH ROW EXECUTE PROCEDURE update_timestamp();

CREATE TRIGGER update_consent_propagation_rules_timestamp
BEFORE UPDATE ON consent_propagation_rules
FOR EACH ROW EXECUTE PROCEDURE update_timestamp();

CREATE TRIGGER update_context_consent_policies_timestamp
BEFORE UPDATE ON context_consent_policies
FOR EACH ROW EXECUTE PROCEDURE update_timestamp();