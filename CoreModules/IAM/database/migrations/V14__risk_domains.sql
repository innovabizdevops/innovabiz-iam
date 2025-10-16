-- V14__risk_domains.sql
-- Modelagem detalhada dos domínios de risco: categorias, tipos, níveis, planos, impactos, consequências
-- Normas: ISO 31000, COSO, COBIT, Basel II/III, Solvência II, entre outras

CREATE SCHEMA IF NOT EXISTS risk;

-- Categorias de Risco
CREATE TABLE risk.risk_category (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    accessibility VARCHAR(50),
    data_source VARCHAR(100),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE risk.risk_category IS 'Categorias de risco segundo normas internacionais (ISO 31000, COSO, etc.)';

-- Tipos de Risco
CREATE TABLE risk.risk_type (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category_id UUID REFERENCES risk.risk_category(id),
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    accessibility VARCHAR(50),
    data_source VARCHAR(100),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE risk.risk_type IS 'Tipos de risco vinculados às categorias de risco';

-- Níveis de Risco
CREATE TABLE risk.risk_level (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    description TEXT,
    score_min INT,
    score_max INT,
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    accessibility VARCHAR(50),
    data_source VARCHAR(100),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE risk.risk_level IS 'Níveis de risco (baixo, médio, alto, crítico, etc.)';

-- Planos de Risco
CREATE TABLE risk.risk_plan (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    start_date DATE,
    end_date DATE,
    responsible VARCHAR(100),
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    accessibility VARCHAR(50),
    data_source VARCHAR(100),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    external_ref VARCHAR(200)
);
COMMENT ON TABLE risk.risk_plan IS 'Planos de risco para mitigação e contingência';

-- Riscos
CREATE TABLE risk.risk (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    description TEXT,
    risk_type_id UUID REFERENCES risk.risk_type(id),
    risk_level_id UUID REFERENCES risk.risk_level(id),
    risk_plan_id UUID REFERENCES risk.risk_plan(id),
    status VARCHAR(30) DEFAULT 'active',
    impact VARCHAR(100),
    consequence VARCHAR(100),
    likelihood VARCHAR(50),
    detected_at TIMESTAMP,
    resolved_at TIMESTAMP,
    responsible VARCHAR(100),
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    accessibility VARCHAR(50),
    data_source VARCHAR(100),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    external_ref VARCHAR(200)
);
COMMENT ON TABLE risk.risk IS 'Riscos identificados na organização';

-- Impactos
CREATE TABLE risk.impact (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    data_source VARCHAR(100),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE risk.impact IS 'Impactos potenciais decorrentes dos riscos';

-- Consequências
CREATE TABLE risk.consequence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    data_source VARCHAR(100),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE risk.consequence IS 'Consequências potenciais decorrentes dos riscos';

-- Tabelas de relacionamento, triggers, views e funções de auditoria, compliance e validação serão adicionados em arquivos específicos para automação e integração BI.
