-- V16__esg_domains.sql
-- Modelagem detalhada do domínio ESG (Ambiental, Social, Governança)
-- Padrões: GRI, SASB, TCFD, ISO 14001, ISO 26000, CDP, entre outros

CREATE SCHEMA IF NOT EXISTS esg;

-- Indicadores Ambientais
CREATE TABLE esg.environmental_indicator (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    unit VARCHAR(30),
    category VARCHAR(50), -- ex: emissões, energia, resíduos
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    compliance_status VARCHAR(20),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE esg.environmental_indicator IS 'Indicadores ambientais conforme GRI, SASB, TCFD, ISO 14001';

-- Indicadores Sociais
CREATE TABLE esg.social_indicator (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    unit VARCHAR(30),
    category VARCHAR(50), -- ex: diversidade, inclusão, saúde, segurança
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    compliance_status VARCHAR(20),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE esg.social_indicator IS 'Indicadores sociais conforme GRI, SASB, ISO 26000';

-- Indicadores de Governança
CREATE TABLE esg.governance_indicator (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    unit VARCHAR(30),
    category VARCHAR(50), -- ex: políticas, conselhos, ética
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    compliance_status VARCHAR(20),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE esg.governance_indicator IS 'Indicadores de governança conforme GRI, SASB, ISO 26000';

-- Valores reportados por empresa/ano
CREATE TABLE esg.company_esg_report (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL,
    year INT NOT NULL,
    indicator_type VARCHAR(20) NOT NULL, -- environmental, social, governance
    indicator_id UUID NOT NULL,
    value NUMERIC,
    unit VARCHAR(30),
    source VARCHAR(100),
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    compliance_status VARCHAR(20),
    external_ref VARCHAR(200),
    CONSTRAINT fk_company FOREIGN KEY(company_id) REFERENCES organization.company(id)
);
COMMENT ON TABLE esg.company_esg_report IS 'Valores de indicadores ESG reportados por empresa e ano';

-- Tabelas auxiliares para certificações, ratings, fontes, etc., podem ser adicionadas conforme necessidade.
-- Triggers, views e funções de auditoria/compliance seguem o padrão dos demais domínios.
