-- Migration: V12__multi_context_identity_base.sql
-- Descrição: Estrutura base para o microserviço de identidade multi-contexto
-- Autor: INNOVABIZ DevOps
-- Data: 20-08-2025

-- Tabela de identidades base
CREATE TABLE IF NOT EXISTS identities (
    identity_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    primary_key_type VARCHAR(50) NOT NULL,  -- cpf, passport, national_id, etc.
    primary_key_value VARCHAR(100) NOT NULL,
    master_person_id UUID,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    trust_level VARCHAR(20) NOT NULL DEFAULT 'BASIC',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,
    CONSTRAINT chk_status CHECK (status IN ('ACTIVE', 'INACTIVE', 'SUSPENDED', 'BLOCKED', 'PENDING_VERIFICATION')),
    CONSTRAINT chk_trust_level CHECK (trust_level IN ('BASIC', 'VERIFIED', 'ENHANCED', 'CERTIFIED', 'DELEGATED'))
);

-- Índices para pesquisa e otimização
CREATE UNIQUE INDEX IF NOT EXISTS idx_identity_primary_key ON identities(primary_key_type, primary_key_value);
CREATE INDEX IF NOT EXISTS idx_identity_status ON identities(status);
CREATE INDEX IF NOT EXISTS idx_identity_trust_level ON identities(trust_level);

-- Tabela de contextos de identidade
CREATE TABLE IF NOT EXISTS identity_contexts (
    context_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identity_id UUID NOT NULL REFERENCES identities(identity_id) ON DELETE CASCADE,
    context_type VARCHAR(50) NOT NULL, -- financial, health, government, education, ecommerce, etc.
    context_subtype VARCHAR(50), -- para classificações mais específicas dentro de cada contexto
    context_status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    trust_score DECIMAL(5,2) CHECK (trust_score >= 0 AND trust_score <= 100),
    verification_level VARCHAR(20) DEFAULT 'BASIC',
    risk_level VARCHAR(20) DEFAULT 'MEDIUM',
    region_code VARCHAR(10), -- código da região relevante (ISO)
    country_code VARCHAR(2), -- código do país (ISO)
    legal_framework VARCHAR(50), -- framework legal aplicável (GDPR, LGPD, etc.)
    issuer VARCHAR(100), -- entidade emissora do contexto
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_verified_at TIMESTAMP,
    metadata JSONB,
    CONSTRAINT chk_context_status CHECK (context_status IN ('ACTIVE', 'INACTIVE', 'SUSPENDED', 'BLOCKED', 'PENDING_VERIFICATION')),
    CONSTRAINT chk_verification_level CHECK (verification_level IN ('BASIC', 'VERIFIED', 'ENHANCED', 'CERTIFIED', 'DELEGATED')),
    CONSTRAINT chk_risk_level CHECK (risk_level IN ('VERY_LOW', 'LOW', 'MEDIUM', 'HIGH', 'VERY_HIGH')),
    UNIQUE(identity_id, context_type, context_subtype)
);

-- Índices para performance
CREATE INDEX IF NOT EXISTS idx_identity_contexts_identity_id ON identity_contexts(identity_id);
CREATE INDEX IF NOT EXISTS idx_identity_contexts_type ON identity_contexts(context_type, context_subtype);
CREATE INDEX IF NOT EXISTS idx_identity_contexts_status ON identity_contexts(context_status);
CREATE INDEX IF NOT EXISTS idx_identity_contexts_region ON identity_contexts(region_code, country_code);

-- Tabela de atributos contextuais
CREATE TABLE IF NOT EXISTS context_attributes (
    attribute_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    context_id UUID NOT NULL REFERENCES identity_contexts(context_id) ON DELETE CASCADE,
    attribute_key VARCHAR(100) NOT NULL,
    attribute_value TEXT,
    encrypted_value BYTEA, -- valor criptografado para atributos sensíveis
    sensitivity_level VARCHAR(20) NOT NULL DEFAULT 'NORMAL',
    verification_status VARCHAR(20) NOT NULL DEFAULT 'UNVERIFIED',
    verification_source VARCHAR(100),
    verification_timestamp TIMESTAMP,
    expiration_date TIMESTAMP,
    is_required BOOLEAN DEFAULT FALSE,
    is_mutable BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    verified_by UUID, -- referência ao usuário/sistema que verificou
    metadata JSONB,
    CONSTRAINT chk_sensitivity_level CHECK (sensitivity_level IN ('PUBLIC', 'NORMAL', 'SENSITIVE', 'HIGHLY_SENSITIVE', 'RESTRICTED')),
    CONSTRAINT chk_verification_status CHECK (verification_status IN ('UNVERIFIED', 'VERIFIED', 'REJECTED', 'EXPIRED', 'PENDING')),
    UNIQUE(context_id, attribute_key)
);

-- Índices para atributos
CREATE INDEX IF NOT EXISTS idx_context_attributes_context_id ON context_attributes(context_id);
CREATE INDEX IF NOT EXISTS idx_context_attributes_key ON context_attributes(attribute_key);
CREATE INDEX IF NOT EXISTS idx_context_attributes_sensitivity ON context_attributes(sensitivity_level);
CREATE INDEX IF NOT EXISTS idx_context_attributes_verification ON context_attributes(verification_status);

-- Tabela de mapeamento entre atributos e contextos
CREATE TABLE IF NOT EXISTS attribute_context_mappings (
    mapping_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_context_id UUID NOT NULL REFERENCES identity_contexts(context_id) ON DELETE CASCADE,
    target_context_id UUID NOT NULL REFERENCES identity_contexts(context_id) ON DELETE CASCADE,
    source_attribute_key VARCHAR(100) NOT NULL,
    target_attribute_key VARCHAR(100) NOT NULL,
    mapping_type VARCHAR(50) NOT NULL DEFAULT 'DIRECT', -- DIRECT, TRANSFORM, CONDITIONAL
    transformation_rule TEXT, -- regra ou função de transformação, quando aplicável
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    CONSTRAINT chk_mapping_type CHECK (mapping_type IN ('DIRECT', 'TRANSFORM', 'CONDITIONAL', 'DERIVED')),
    CONSTRAINT chk_different_contexts CHECK (source_context_id <> target_context_id),
    UNIQUE(source_context_id, target_context_id, source_attribute_key, target_attribute_key)
);

-- Índices para mapeamentos
CREATE INDEX IF NOT EXISTS idx_attribute_mappings_source ON attribute_context_mappings(source_context_id, source_attribute_key);
CREATE INDEX IF NOT EXISTS idx_attribute_mappings_target ON attribute_context_mappings(target_context_id, target_attribute_key);

-- Comentários para documentação do esquema
COMMENT ON TABLE identities IS 'Tabela principal de identidades base no sistema multi-contexto';
COMMENT ON TABLE identity_contexts IS 'Contextos específicos para cada identidade base (financeiro, saúde, etc.)';
COMMENT ON TABLE context_attributes IS 'Atributos específicos para cada contexto de identidade';
COMMENT ON TABLE attribute_context_mappings IS 'Mapeamentos entre atributos em diferentes contextos';

-- Função para atualização automática do timestamp
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = CURRENT_TIMESTAMP;
   RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers para atualização automática de timestamps
CREATE TRIGGER update_identities_timestamp
BEFORE UPDATE ON identities
FOR EACH ROW EXECUTE PROCEDURE update_timestamp();

CREATE TRIGGER update_identity_contexts_timestamp
BEFORE UPDATE ON identity_contexts
FOR EACH ROW EXECUTE PROCEDURE update_timestamp();

CREATE TRIGGER update_context_attributes_timestamp
BEFORE UPDATE ON context_attributes
FOR EACH ROW EXECUTE PROCEDURE update_timestamp();

CREATE TRIGGER update_attribute_context_mappings_timestamp
BEFORE UPDATE ON attribute_context_mappings
FOR EACH ROW EXECUTE PROCEDURE update_timestamp();