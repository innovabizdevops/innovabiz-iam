-- INNOVABIZ IAM - Migração 002: Criação de tabelas para auditoria
-- Autor: Eduardo Jeremias
-- Versão: 1.0.0
-- Descrição: Cria as tabelas necessárias para o sistema de auditoria multi-contexto

-- Tabela principal para eventos de auditoria
CREATE TABLE audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category audit_event_category NOT NULL,
    action VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),
    resource_name VARCHAR(255),
    severity audit_event_severity NOT NULL DEFAULT 'INFO',
    success BOOLEAN NOT NULL DEFAULT TRUE,
    error_message TEXT,
    details JSONB,
    tags TEXT[] DEFAULT '{}',
    tenant_id VARCHAR(100) NOT NULL,
    regional_context VARCHAR(10),
    country_code VARCHAR(2),
    language VARCHAR(10),
    user_id VARCHAR(100),
    user_name VARCHAR(255),
    correlation_id VARCHAR(100) NOT NULL,
    source_ip VARCHAR(45),
    http_details JSONB,
    compliance_frameworks TEXT[] DEFAULT '{}',
    masked_fields TEXT[] DEFAULT '{}',
    anonymized_fields TEXT[] DEFAULT '{}',
    partition_key VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tabela para políticas de retenção de auditoria
CREATE TABLE audit_retention_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(100) NOT NULL,
    regional_context VARCHAR(10),
    retention_days INTEGER NOT NULL,
    compliance_framework compliance_framework NOT NULL,
    category audit_event_category,
    description TEXT NOT NULL,
    automatic_anonymization BOOLEAN NOT NULL DEFAULT FALSE,
    anonymization_fields TEXT[] DEFAULT '{}',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tabela para relatórios de compliance
CREATE TABLE audit_compliance_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(100) NOT NULL,
    regional_context VARCHAR(10),
    compliance_framework compliance_framework NOT NULL,
    report_name VARCHAR(255) NOT NULL,
    report_description TEXT,
    status report_status NOT NULL DEFAULT 'PENDING',
    event_count INTEGER DEFAULT 0,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    report_data JSONB,
    report_url VARCHAR(255),
    created_by VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tabela para estatísticas de auditoria
CREATE TABLE audit_statistics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(100) NOT NULL,
    regional_context VARCHAR(10),
    statistics_type VARCHAR(100) NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    statistics_data JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Índices para otimização de consultas
CREATE INDEX idx_audit_events_tenant_id ON audit_events (tenant_id);
CREATE INDEX idx_audit_events_regional_context ON audit_events (regional_context);
CREATE INDEX idx_audit_events_category ON audit_events (category);
CREATE INDEX idx_audit_events_created_at ON audit_events (created_at);
CREATE INDEX idx_audit_events_correlation_id ON audit_events (correlation_id);
CREATE INDEX idx_audit_events_partition_key ON audit_events (partition_key);
CREATE INDEX idx_audit_events_user_id ON audit_events (user_id);
CREATE INDEX idx_audit_events_success ON audit_events (success);
CREATE INDEX idx_audit_events_tags ON audit_events USING GIN (tags);
CREATE INDEX idx_audit_events_details ON audit_events USING GIN (details);

-- Índices para políticas de retenção
CREATE INDEX idx_audit_retention_policies_tenant_id ON audit_retention_policies (tenant_id);
CREATE INDEX idx_audit_retention_policies_compliance ON audit_retention_policies (compliance_framework);
CREATE INDEX idx_audit_retention_policies_active ON audit_retention_policies (active);

-- Índices para relatórios
CREATE INDEX idx_audit_compliance_reports_tenant_id ON audit_compliance_reports (tenant_id);
CREATE INDEX idx_audit_compliance_reports_status ON audit_compliance_reports (status);
CREATE INDEX idx_audit_compliance_reports_created_at ON audit_compliance_reports (created_at);

-- Índices para estatísticas
CREATE INDEX idx_audit_statistics_tenant_id ON audit_statistics (tenant_id);
CREATE INDEX idx_audit_statistics_period ON audit_statistics (period_start, period_end);

-- Comentários para documentação
COMMENT ON TABLE audit_events IS 'Armazena todos os eventos de auditoria com suporte a multi-tenant e multi-regional';
COMMENT ON TABLE audit_retention_policies IS 'Define políticas de retenção para eventos de auditoria por tenant/região';
COMMENT ON TABLE audit_compliance_reports IS 'Armazena relatórios de compliance gerados pelo sistema';
COMMENT ON TABLE audit_statistics IS 'Estatísticas agregadas de eventos de auditoria';

-- Comentários para colunas principais
COMMENT ON COLUMN audit_events.partition_key IS 'Chave de particionamento lógico no formato tenant_id:regional_context:YYYY-MM';
COMMENT ON COLUMN audit_events.masked_fields IS 'Campos que foram mascarados por políticas de privacidade';
COMMENT ON COLUMN audit_events.anonymized_fields IS 'Campos que foram anonimizados por políticas de retenção';
COMMENT ON COLUMN audit_events.compliance_frameworks IS 'Frameworks de compliance aplicáveis a este evento';
COMMENT ON COLUMN audit_events.http_details IS 'Detalhes da requisição HTTP relacionada ao evento';

-- Triggers para atualização automática de timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER trigger_update_audit_events_timestamp 
BEFORE UPDATE ON audit_events
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_audit_retention_policies_timestamp 
BEFORE UPDATE ON audit_retention_policies
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_audit_compliance_reports_timestamp 
BEFORE UPDATE ON audit_compliance_reports
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_audit_statistics_timestamp 
BEFORE UPDATE ON audit_statistics
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();