-- Migrations para criação das tabelas de histórico de pontuação de confiança
-- Compatível com PostgreSQL 12+
-- Este esquema suporta multi-tenant, multi-contexto e multi-dimensional

-- Tabela principal de histórico de pontuação
CREATE TABLE IF NOT EXISTS trust_score_history (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    tenant_id VARCHAR(36) NOT NULL,
    context_id VARCHAR(36) NOT NULL, -- Contexto da avaliação (aplicação, serviço, etc)
    region_code VARCHAR(10), -- Código da região (ex: AO, BR, PT)
    overall_score INTEGER NOT NULL CHECK (overall_score >= 0 AND overall_score <= 100),
    confidence_level DECIMAL(5,4) NOT NULL CHECK (confidence_level >= 0 AND confidence_level <= 1),
    evaluation_time_ms INTEGER NOT NULL, -- Tempo de processamento em ms
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB,
    
    -- Índices para consultas otimizadas
    CONSTRAINT uk_trust_score_history UNIQUE (user_id, tenant_id, context_id, created_at)
);

-- Índices para consultas frequentes
CREATE INDEX idx_trust_score_history_user_tenant ON trust_score_history(user_id, tenant_id);
CREATE INDEX idx_trust_score_history_context ON trust_score_history(context_id);
CREATE INDEX idx_trust_score_history_region ON trust_score_history(region_code);
CREATE INDEX idx_trust_score_history_created_at ON trust_score_history(created_at DESC);
CREATE INDEX idx_trust_score_history_score ON trust_score_history(overall_score);

-- Tabela para armazenar pontuações por dimensão
CREATE TABLE IF NOT EXISTS trust_score_dimension_history (
    id BIGSERIAL PRIMARY KEY,
    trust_score_history_id BIGINT NOT NULL REFERENCES trust_score_history(id) ON DELETE CASCADE,
    dimension VARCHAR(32) NOT NULL, -- Enum: IDENTITY, BEHAVIORAL, FINANCIAL, etc.
    score INTEGER NOT NULL CHECK (score >= 0 AND score <= 100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uk_trust_score_dimension_history UNIQUE (trust_score_history_id, dimension)
);

CREATE INDEX idx_trust_score_dimension_history_score_id ON trust_score_dimension_history(trust_score_history_id);
CREATE INDEX idx_trust_score_dimension_type ON trust_score_dimension_history(dimension);

-- Tabela para armazenar fatores que influenciaram a pontuação
CREATE TABLE IF NOT EXISTS trust_score_factors (
    id BIGSERIAL PRIMARY KEY,
    trust_score_history_id BIGINT NOT NULL REFERENCES trust_score_history(id) ON DELETE CASCADE,
    factor_id VARCHAR(64) NOT NULL,
    dimension VARCHAR(32) NOT NULL,
    factor_name VARCHAR(128) NOT NULL,
    factor_description TEXT,
    factor_type VARCHAR(32) NOT NULL, -- Enum: POSITIVE, NEGATIVE, NEUTRAL
    weight DECIMAL(5,4) NOT NULL,
    value DECIMAL(5,4) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trust_score_factors_score_id ON trust_score_factors(trust_score_history_id);
CREATE INDEX idx_trust_score_factors_dimension ON trust_score_factors(dimension);
CREATE INDEX idx_trust_score_factors_type ON trust_score_factors(factor_type);

-- Tabela para armazenar anomalias detectadas
CREATE TABLE IF NOT EXISTS trust_score_anomalies (
    id BIGSERIAL PRIMARY KEY,
    trust_score_history_id BIGINT NOT NULL REFERENCES trust_score_history(id) ON DELETE CASCADE,
    anomaly_id VARCHAR(64) NOT NULL,
    anomaly_type VARCHAR(32) NOT NULL, -- Enum: LOCATION_ANOMALY, DEVICE_ANOMALY, etc.
    description TEXT NOT NULL,
    severity VARCHAR(16) NOT NULL, -- Enum: LOW, MEDIUM, HIGH, CRITICAL
    confidence DECIMAL(5,4) NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    affected_dimensions JSONB, -- Array de dimensões afetadas
    metadata JSONB,
    detected_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_trust_score_anomalies_score_id ON trust_score_anomalies(trust_score_history_id);
CREATE INDEX idx_trust_score_anomalies_type ON trust_score_anomalies(anomaly_type);
CREATE INDEX idx_trust_score_anomalies_severity ON trust_score_anomalies(severity);

-- Tabela de estatísticas de pontuação por tenant/região
CREATE TABLE IF NOT EXISTS trust_score_statistics (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    region_code VARCHAR(10),
    context_id VARCHAR(36),
    dimension VARCHAR(32),
    avg_score DECIMAL(5,2) NOT NULL,
    median_score INTEGER,
    std_deviation DECIMAL(5,2),
    sample_size INTEGER NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    last_updated TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uk_trust_score_statistics UNIQUE (tenant_id, region_code, context_id, dimension, period_start, period_end)
);

CREATE INDEX idx_trust_score_statistics_tenant ON trust_score_statistics(tenant_id, region_code);
CREATE INDEX idx_trust_score_statistics_period ON trust_score_statistics(period_start, period_end);

-- Tabela para armazenar políticas de retenção e agregação
CREATE TABLE IF NOT EXISTS trust_score_retention_policy (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    detail_retention_days INTEGER NOT NULL DEFAULT 90, -- Quantos dias manter dados detalhados
    aggregation_interval VARCHAR(16) NOT NULL DEFAULT 'DAILY', -- HOURLY, DAILY, WEEKLY, MONTHLY
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uk_trust_score_retention_policy UNIQUE (tenant_id)
);

-- Função para atualizar as estatísticas de forma eficiente
CREATE OR REPLACE FUNCTION update_trust_score_statistics()
RETURNS TRIGGER AS $$
BEGIN
    -- Lógica para atualizar estatísticas incrementalmente
    -- Será chamada por triggers nas inserções de histórico
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger para atualizar estatísticas quando novos registros forem adicionados
CREATE TRIGGER trust_score_statistics_update
AFTER INSERT ON trust_score_history
FOR EACH ROW
EXECUTE FUNCTION update_trust_score_statistics();

-- Trigger para manter a consistência entre tabelas relacionadas
CREATE OR REPLACE FUNCTION check_trust_score_consistency()
RETURNS TRIGGER AS $$
BEGIN
    -- Verifica se o registro pai existe antes de inserir
    IF NOT EXISTS (SELECT 1 FROM trust_score_history WHERE id = NEW.trust_score_history_id) THEN
        RAISE EXCEPTION 'Registro pai não encontrado na tabela trust_score_history (ID: %)', NEW.trust_score_history_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER check_dimension_history_consistency
BEFORE INSERT ON trust_score_dimension_history
FOR EACH ROW
EXECUTE FUNCTION check_trust_score_consistency();

CREATE TRIGGER check_factors_consistency
BEFORE INSERT ON trust_score_factors
FOR EACH ROW
EXECUTE FUNCTION check_trust_score_consistency();

CREATE TRIGGER check_anomalies_consistency
BEFORE INSERT ON trust_score_anomalies
FOR EACH ROW
EXECUTE FUNCTION check_trust_score_consistency();

-- Comentários para documentação
COMMENT ON TABLE trust_score_history IS 'Armazena o histórico completo das pontuações de confiança dos usuários';
COMMENT ON TABLE trust_score_dimension_history IS 'Armazena as pontuações por dimensão para cada avaliação de confiança';
COMMENT ON TABLE trust_score_factors IS 'Armazena os fatores que influenciaram cada pontuação de confiança';
COMMENT ON TABLE trust_score_anomalies IS 'Armazena as anomalias detectadas durante as avaliações de confiança';
COMMENT ON TABLE trust_score_statistics IS 'Armazena estatísticas agregadas das pontuações de confiança para análises';
COMMENT ON TABLE trust_score_retention_policy IS 'Define políticas de retenção e agregação de dados por tenant';