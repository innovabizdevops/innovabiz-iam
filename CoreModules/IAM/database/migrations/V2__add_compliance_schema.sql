-- V2__add_compliance_schema.sql
-- Criação do schema e tabelas principais do domínio compliance

CREATE SCHEMA IF NOT EXISTS compliance;

-- Exemplo de tabela de compliance
CREATE TABLE IF NOT EXISTS compliance.regulatory_frameworks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Outras tabelas e objetos do domínio compliance devem ser adicionados aqui
