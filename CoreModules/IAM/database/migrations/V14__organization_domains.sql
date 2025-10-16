-- V14__organization_domains.sql
-- Modelagem detalhada dos domínios organizacionais da suite InnovaBiz
-- Normas: TOGAF, DMBOK, ISO/IEC 11179, ISO 8000, DAMA-DMBOK, COBIT, ISO 9001, ISO 37301, GDPR, entre outras

CREATE SCHEMA IF NOT EXISTS organization;

-- Tipos de Empresa
CREATE TABLE organization.company_type (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    classification VARCHAR(50), -- Ex: S.A., Ltda, ONG
    country VARCHAR(100),
    continent VARCHAR(50),
    reference VARCHAR(200),
    entity VARCHAR(100), -- órgão regulador
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
COMMENT ON TABLE organization.company_type IS 'Tipos de empresa segundo normas internacionais e legislação nacional';
COMMENT ON COLUMN organization.company_type.name IS 'Nome do tipo de empresa (ex: Sociedade Anônima, Limitada)';
COMMENT ON COLUMN organization.company_type.classification IS 'Classificação conforme legislação e normas (ex: S.A., Ltda)';
COMMENT ON COLUMN organization.company_type.reference IS 'Referência normativa ou legal do tipo de empresa';
COMMENT ON COLUMN organization.company_type.entity IS 'Entidade reguladora responsável pelo tipo de empresa';

-- Empresas
CREATE TABLE organization.company (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    description TEXT,
    company_type_id UUID REFERENCES organization.company_type(id),
    tax_id VARCHAR(50),
    registration_number VARCHAR(50),
    founded_date DATE,
    status VARCHAR(30) DEFAULT 'active',
    country VARCHAR(100),
    continent VARCHAR(50),
    main_language VARCHAR(50),
    accessibility VARCHAR(50),
    data_source VARCHAR(100),
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    responsible VARCHAR(100),
    external_ref VARCHAR(200),
    parent_id UUID,
    hierarchy_level INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID
);
COMMENT ON TABLE organization.company IS 'Empresas cadastradas na plataforma, conforme padrões internacionais';
COMMENT ON COLUMN organization.company.company_type_id IS 'Referência ao tipo de empresa';
COMMENT ON COLUMN organization.company.tax_id IS 'Número de identificação fiscal da empresa';
COMMENT ON COLUMN organization.company.registration_number IS 'Número de registro legal da empresa';

-- Acionistas
CREATE TABLE organization.shareholder (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID REFERENCES organization.company(id),
    name VARCHAR(100) NOT NULL,
    shareholder_type VARCHAR(50), -- Ex: Individual, Corporate
    country VARCHAR(100),
    ownership_percentage NUMERIC(5,2),
    status VARCHAR(30) DEFAULT 'active',
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    responsible VARCHAR(100),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    external_ref VARCHAR(200)
);
COMMENT ON TABLE organization.shareholder IS 'Acionistas das empresas cadastradas';
COMMENT ON COLUMN organization.shareholder.company_id IS 'Referência à empresa da qual é acionista';
COMMENT ON COLUMN organization.shareholder.ownership_percentage IS 'Percentual de participação societária';

-- Clientes (Pessoa Jurídica ou Física)
CREATE TABLE organization.client (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    client_type VARCHAR(50), -- Ex: Individual, Corporate
    company_id UUID REFERENCES organization.company(id),
    industry VARCHAR(100),
    country VARCHAR(100),
    status VARCHAR(30) DEFAULT 'active',
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    responsible VARCHAR(100),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    external_ref VARCHAR(200)
);
COMMENT ON TABLE organization.client IS 'Clientes da empresa (pessoa jurídica ou física)';
COMMENT ON COLUMN organization.client.company_id IS 'Empresa à qual o cliente está relacionado';

-- Estrutura Organizacional
CREATE TABLE organization.org_structure (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID REFERENCES organization.company(id),
    name VARCHAR(100) NOT NULL,
    structure_type VARCHAR(50), -- Ex: Departamento, Diretoria, Gabinete
    parent_id UUID,
    hierarchy_level INT DEFAULT 1,
    status VARCHAR(30) DEFAULT 'active',
    responsible VARCHAR(100),
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    external_ref VARCHAR(200)
);
COMMENT ON TABLE organization.org_structure IS 'Estrutura organizacional da empresa (departamentos, diretorias, etc.)';
COMMENT ON COLUMN organization.org_structure.company_id IS 'Empresa à qual pertence a estrutura organizacional';
COMMENT ON COLUMN organization.org_structure.structure_type IS 'Tipo de estrutura (departamento, diretoria, etc.)';

-- Departamentos
CREATE TABLE organization.department (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_structure_id UUID REFERENCES organization.org_structure(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    responsible VARCHAR(100),
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    external_ref VARCHAR(200)
);
COMMENT ON TABLE organization.department IS 'Departamentos das estruturas organizacionais';
COMMENT ON COLUMN organization.department.org_structure_id IS 'Estrutura organizacional à qual pertence o departamento';

-- Comites, Gabinetes, Direção
CREATE TABLE organization.committee (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_structure_id UUID REFERENCES organization.org_structure(id),
    name VARCHAR(100) NOT NULL,
    committee_type VARCHAR(50), -- Ex: Comitê Executivo, Conselho Fiscal
    status VARCHAR(30) DEFAULT 'active',
    responsible VARCHAR(100),
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    external_ref VARCHAR(200)
);
COMMENT ON TABLE organization.committee IS 'Comitês e conselhos da estrutura organizacional';
COMMENT ON COLUMN organization.committee.org_structure_id IS 'Estrutura organizacional à qual pertence o comitê';
COMMENT ON COLUMN organization.committee.committee_type IS 'Tipo de comitê/conselho';

-- Índices e constraints adicionais podem ser adicionados conforme necessário para garantir performance e integridade referencial.

-- Triggers, views e funções de auditoria, compliance e validação serão adicionados em arquivos específicos para automação e integração BI.
