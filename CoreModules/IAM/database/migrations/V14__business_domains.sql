-- V14__business_domains.sql
-- Modelagem detalhada dos domínios de negócio da suite InnovaBiz
-- Normas: TOGAF, DMBOK, ISO/IEC 11179, ISO 8000, DAMA-DMBOK, COBIT, ISO 9001, ISO 37301, entre outras

CREATE SCHEMA IF NOT EXISTS business;

-- Tipos de Modelos de Negócio
CREATE TABLE business.business_model_type (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    reference VARCHAR(200),
    entity VARCHAR(100),
    valid_from DATE,
    valid_to DATE,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    accessibility VARCHAR(50),
    data_source VARCHAR(100),
    parent_id UUID,
    hierarchy_level INT DEFAULT 1,
    external_ref VARCHAR(200)
);
COMMENT ON TABLE business.business_model_type IS 'Tipos de modelos de negócio conforme frameworks e benchmarks internacionais';

-- Modelos de Negócio
CREATE TABLE business.business_model (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    business_model_type_id UUID REFERENCES business.business_model_type(id),
    status VARCHAR(30) DEFAULT 'active',
    reference VARCHAR(200),
    entity VARCHAR(100),
    valid_from DATE,
    valid_to DATE,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    accessibility VARCHAR(50),
    data_source VARCHAR(100),
    parent_id UUID,
    hierarchy_level INT DEFAULT 1,
    external_ref VARCHAR(200)
);
COMMENT ON TABLE business.business_model IS 'Modelos de negócio aplicados na organização';

-- Modelos de Mercado
CREATE TABLE business.market_model (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    reference VARCHAR(200),
    entity VARCHAR(100),
    valid_from DATE,
    valid_to DATE,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    accessibility VARCHAR(50),
    data_source VARCHAR(100),
    parent_id UUID,
    hierarchy_level INT DEFAULT 1,
    external_ref VARCHAR(200)
);
COMMENT ON TABLE business.market_model IS 'Modelos de mercado segundo benchmarks e padrões internacionais';

-- Modelos Operacionais
CREATE TABLE business.operational_model (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    reference VARCHAR(200),
    entity VARCHAR(100),
    valid_from DATE,
    valid_to DATE,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    accessibility VARCHAR(50),
    data_source VARCHAR(100),
    parent_id UUID,
    hierarchy_level INT DEFAULT 1,
    external_ref VARCHAR(200)
);
COMMENT ON TABLE business.operational_model IS 'Modelos operacionais conforme frameworks e melhores práticas';

-- Planos Estratégicos
CREATE TABLE business.strategic_plan (
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
COMMENT ON TABLE business.strategic_plan IS 'Planos estratégicos organizacionais';

-- Planos de Negócio
CREATE TABLE business.business_plan (
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
COMMENT ON TABLE business.business_plan IS 'Planos de negócio segundo padrões internacionais';

-- Planos de Contingência
CREATE TABLE business.contingency_plan (
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
COMMENT ON TABLE business.contingency_plan IS 'Planos de contingência para gestão de riscos e continuidade';

-- Tipos de Estratégias
CREATE TABLE business.strategy_type (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    reference VARCHAR(200),
    entity VARCHAR(100),
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
COMMENT ON TABLE business.strategy_type IS 'Tipos de estratégias organizacionais';

-- Estratégias
CREATE TABLE business.strategy (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    description TEXT,
    strategy_type_id UUID REFERENCES business.strategy_type(id),
    status VARCHAR(30) DEFAULT 'active',
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
COMMENT ON TABLE business.strategy IS 'Estratégias de negócio e organizacionais';

-- Segmentos de Mercado
CREATE TABLE business.market_segment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    market_model_id UUID REFERENCES business.market_model(id),
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
COMMENT ON TABLE business.market_segment IS 'Segmentos de mercado conforme benchmarks e padrões internacionais';

-- Índices e constraints adicionais podem ser adicionados conforme necessário para garantir performance e integridade referencial.

-- Triggers, views e funções de auditoria, compliance e validação serão adicionados em arquivos específicos para automação e integração BI.
