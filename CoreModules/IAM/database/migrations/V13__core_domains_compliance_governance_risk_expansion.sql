-- Expansão de todos os domínios principais com governança, compliance, rastreabilidade, interoperabilidade e campos de padronização internacional

-- CONTRATOS
ALTER TABLE contracts
    ADD COLUMN official_number VARCHAR(50),
    ADD COLUMN external_code VARCHAR(50),
    ADD COLUMN english_title VARCHAR(200),
    ADD COLUMN abbreviation VARCHAR(20),
    ADD COLUMN status VARCHAR(20) DEFAULT 'ativo',
    ADD COLUMN parent_id UUID REFERENCES contracts(id),
    ADD COLUMN hierarchy_level INT DEFAULT 1,
    ADD COLUMN data_source VARCHAR(100),
    ADD COLUMN valid_from DATE,
    ADD COLUMN valid_to DATE,
    ADD COLUMN responsible VARCHAR(100),
    ADD COLUMN external_ref VARCHAR(200),
    ADD COLUMN compliance_status VARCHAR(20),
    ADD COLUMN privacy_level VARCHAR(50),
    ADD COLUMN main_language VARCHAR(50),
    ADD COLUMN accessibility VARCHAR(50),
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID;

-- PROCESSOS
ALTER TABLE processes
    ADD COLUMN official_code VARCHAR(50),
    ADD COLUMN external_code VARCHAR(50),
    ADD COLUMN english_name VARCHAR(200),
    ADD COLUMN abbreviation VARCHAR(20),
    ADD COLUMN status VARCHAR(20) DEFAULT 'ativo',
    ADD COLUMN parent_id UUID REFERENCES processes(id),
    ADD COLUMN hierarchy_level INT DEFAULT 1,
    ADD COLUMN data_source VARCHAR(100),
    ADD COLUMN valid_from DATE,
    ADD COLUMN valid_to DATE,
    ADD COLUMN responsible VARCHAR(100),
    ADD COLUMN external_ref VARCHAR(200),
    ADD COLUMN compliance_status VARCHAR(20),
    ADD COLUMN privacy_level VARCHAR(50),
    ADD COLUMN main_language VARCHAR(50),
    ADD COLUMN accessibility VARCHAR(50),
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID;

-- KPIs
ALTER TABLE kpis.indicator
    ADD COLUMN official_code VARCHAR(50),
    ADD COLUMN external_code VARCHAR(50),
    ADD COLUMN english_name VARCHAR(200),
    ADD COLUMN abbreviation VARCHAR(20),
    ADD COLUMN status VARCHAR(20) DEFAULT 'ativo',
    ADD COLUMN parent_id UUID REFERENCES kpis.indicator(id),
    ADD COLUMN hierarchy_level INT DEFAULT 1,
    ADD COLUMN data_source VARCHAR(100),
    ADD COLUMN valid_from DATE,
    ADD COLUMN valid_to DATE,
    ADD COLUMN responsible VARCHAR(100),
    ADD COLUMN external_ref VARCHAR(200),
    ADD COLUMN compliance_status VARCHAR(20),
    ADD COLUMN privacy_level VARCHAR(50),
    ADD COLUMN main_language VARCHAR(50),
    ADD COLUMN accessibility VARCHAR(50),
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID;

-- COMPLIANCE
ALTER TABLE compliance_regulatory_frameworks
    ADD COLUMN official_code VARCHAR(50),
    ADD COLUMN external_code VARCHAR(50),
    ADD COLUMN english_name VARCHAR(200),
    ADD COLUMN abbreviation VARCHAR(20),
    ADD COLUMN status VARCHAR(20) DEFAULT 'ativo',
    ADD COLUMN parent_id UUID REFERENCES compliance_regulatory_frameworks(id),
    ADD COLUMN hierarchy_level INT DEFAULT 1,
    ADD COLUMN data_source VARCHAR(100),
    ADD COLUMN valid_from DATE,
    ADD COLUMN valid_to DATE,
    ADD COLUMN responsible VARCHAR(100),
    ADD COLUMN external_ref VARCHAR(200),
    ADD COLUMN compliance_status VARCHAR(20),
    ADD COLUMN privacy_level VARCHAR(50),
    ADD COLUMN main_language VARCHAR(50),
    ADD COLUMN accessibility VARCHAR(50),
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID;

-- GOVERNANÇA
ALTER TABLE governance_orgao_social
    ADD COLUMN official_code VARCHAR(50),
    ADD COLUMN external_code VARCHAR(50),
    ADD COLUMN english_name VARCHAR(200),
    ADD COLUMN abbreviation VARCHAR(20),
    ADD COLUMN status VARCHAR(20) DEFAULT 'ativo',
    ADD COLUMN parent_id UUID REFERENCES governance_orgao_social(id),
    ADD COLUMN hierarchy_level INT DEFAULT 1,
    ADD COLUMN data_source VARCHAR(100),
    ADD COLUMN valid_from DATE,
    ADD COLUMN valid_to DATE,
    ADD COLUMN responsible VARCHAR(100),
    ADD COLUMN external_ref VARCHAR(200),
    ADD COLUMN compliance_status VARCHAR(20),
    ADD COLUMN privacy_level VARCHAR(50),
    ADD COLUMN main_language VARCHAR(50),
    ADD COLUMN accessibility VARCHAR(50),
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID;

-- RISCO
ALTER TABLE risk_risk
    ADD COLUMN official_code VARCHAR(50),
    ADD COLUMN external_code VARCHAR(50),
    ADD COLUMN english_name VARCHAR(200),
    ADD COLUMN abbreviation VARCHAR(20),
    ADD COLUMN status VARCHAR(20) DEFAULT 'ativo',
    ADD COLUMN parent_id UUID REFERENCES risk_risk(id),
    ADD COLUMN hierarchy_level INT DEFAULT 1,
    ADD COLUMN data_source VARCHAR(100),
    ADD COLUMN valid_from DATE,
    ADD COLUMN valid_to DATE,
    ADD COLUMN responsible VARCHAR(100),
    ADD COLUMN external_ref VARCHAR(200),
    ADD COLUMN compliance_status VARCHAR(20),
    ADD COLUMN privacy_level VARCHAR(50),
    ADD COLUMN main_language VARCHAR(50),
    ADD COLUMN accessibility VARCHAR(50),
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID;
