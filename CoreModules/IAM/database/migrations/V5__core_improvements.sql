-- V5__core_improvements.sql
-- Melhorias estruturais para a base de dados principal (core) da Innovabiz

-- 1. Criação da tabela de integrações externas
CREATE SCHEMA IF NOT EXISTS core;

CREATE TABLE IF NOT EXISTS core.external_integrations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID NOT NULL,
    external_system VARCHAR(100) NOT NULL,
    external_id VARCHAR(255),
    sync_status VARCHAR(50) DEFAULT 'pending',
    last_synced_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. Criação da tabela de auditoria
CREATE TABLE IF NOT EXISTS core.audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID,
    action VARCHAR(50) NOT NULL,
    performed_by UUID,
    performed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    old_value JSONB,
    new_value JSONB,
    source_system VARCHAR(100)
);

-- 3. Adição de campos de controle em tabelas principais
ALTER TABLE IF EXISTS iam.organizations
    ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'active',
    ADD COLUMN IF NOT EXISTS external_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1;

ALTER TABLE IF EXISTS compliance.regulatory_frameworks
    ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'active',
    ADD COLUMN IF NOT EXISTS external_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1;

ALTER TABLE IF EXISTS analytics.feature_store
    ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'active',
    ADD COLUMN IF NOT EXISTS external_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1;

ALTER TABLE IF EXISTS monitoring.query_metrics
    ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'active',
    ADD COLUMN IF NOT EXISTS external_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS source_system VARCHAR(100),
    ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1;

-- 4. Criação de tabela de configurações dinâmicas
CREATE TABLE IF NOT EXISTS core.configurations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    config_key VARCHAR(100) NOT NULL,
    config_value TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 5. Criação de tabela de consentimento e privacidade
CREATE TABLE IF NOT EXISTS core.privacy_consents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    consent_type VARCHAR(100) NOT NULL,
    consent_given BOOLEAN DEFAULT TRUE,
    consent_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    revoked_date TIMESTAMP WITH TIME ZONE
);
