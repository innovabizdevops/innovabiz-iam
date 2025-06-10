-- V4__add_monitoring_schema.sql
-- Criação do schema e tabelas principais do domínio monitoring

CREATE SCHEMA IF NOT EXISTS monitoring;

-- Exemplo de tabela de monitoring
CREATE TABLE IF NOT EXISTS monitoring.query_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    metric_name VARCHAR(255) NOT NULL,
    value NUMERIC,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Outras tabelas e objetos do domínio monitoring devem ser adicionados aqui
