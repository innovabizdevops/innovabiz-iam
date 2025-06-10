-- V1__create_core_schema.sql
-- Criação do schema e tabelas principais do domínio core

CREATE SCHEMA IF NOT EXISTS iam;

-- Exemplo de tabela principal
CREATE TABLE IF NOT EXISTS iam.organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Outras tabelas e objetos do domínio core devem ser adicionados aqui
