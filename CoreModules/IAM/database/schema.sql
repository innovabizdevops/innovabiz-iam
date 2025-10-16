-- ========================================
-- MÓDULO CORE IAM - DATABASE SCHEMA
-- PostgreSQL 17+ com extensões avançadas
-- ========================================

-- Habilitar extensões necessárias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gin";
CREATE EXTENSION IF NOT EXISTS "postgres_fdw";

-- ========================================
-- SCHEMA: iam_core
-- ========================================
CREATE SCHEMA IF NOT EXISTS iam_core;
SET search_path TO iam_core, public;

-- ========================================
-- TABELAS DE DOMÍNIO
-- ========================================

-- Tabela de Organizações (Multi-tenant)
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'enterprise', 'government', 'education', 'healthcare'
    industry VARCHAR(100),
    country_code CHAR(2) NOT NULL,
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',    status VARCHAR(50) NOT NULL DEFAULT 'active',
    settings JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_by UUID
);

-- Tabela de Identidades (Users/Services/Devices)
CREATE TABLE identities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    identity_type VARCHAR(50) NOT NULL, -- 'user', 'service', 'device', 'bot'
    external_id VARCHAR(255),
    username VARCHAR(255),
    email VARCHAR(255),
    phone VARCHAR(50),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    
    -- Informações pessoais (criptografadas)
    personal_info JSONB DEFAULT '{}', -- encrypted
    
    -- Configurações de segurança
    security_settings JSONB DEFAULT '{}',
    risk_score INTEGER DEFAULT 0,
    trust_level INTEGER DEFAULT 0,
    
    -- Metadados
    attributes JSONB DEFAULT '{}',
    tags TEXT[],