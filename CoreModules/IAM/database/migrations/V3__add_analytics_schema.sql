-- V3__add_analytics_schema.sql
-- Criação do schema e tabelas principais do domínio analytics

CREATE SCHEMA IF NOT EXISTS analytics;

-- Exemplo de tabela de analytics
CREATE TABLE IF NOT EXISTS analytics.feature_store (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    feature_name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Outras tabelas e objetos do domínio analytics devem ser adicionados aqui
