-- V14__compliance_domains.sql
-- Modelagem detalhada dos domínios de compliance, normas, frameworks, regulamentos, legislações e suas relações
-- Normas: ISO 37301, ISO 27001, GDPR, ITIL, COBIT, DAMA-DMBOK, entre outras

CREATE SCHEMA IF NOT EXISTS compliance;

-- Tipos de Normas, Frameworks, Regulamentos
CREATE TABLE compliance.compliance_type (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    type VARCHAR(50), -- Ex: Norma, Framework, Regulamento, Lei, Decreto
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    data_source VARCHAR(100),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE compliance.compliance_type IS 'Tipos de compliance: normas, frameworks, regulamentos, leis, decretos, etc.';

-- Normas, Frameworks, Regulamentos, Legislações
CREATE TABLE compliance.compliance_reference (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    description TEXT,
    compliance_type_id UUID REFERENCES compliance.compliance_type(id),
    country VARCHAR(100),
    continent VARCHAR(50),
    start_date DATE,
    end_date DATE,
    state VARCHAR(50),
    situation VARCHAR(50),
    reference VARCHAR(200),
    entity VARCHAR(100),
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    data_source VARCHAR(100),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE compliance.compliance_reference IS 'Normas, frameworks, regulamentos, legislações e decretos aplicáveis';

-- Relação entre Normas/Frameworks e Empresas
CREATE TABLE compliance.company_compliance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID REFERENCES organization.company(id),
    compliance_reference_id UUID REFERENCES compliance.compliance_reference(id),
    status VARCHAR(30) DEFAULT 'active',
    valid_from DATE,
    valid_to DATE,
    responsible VARCHAR(100),
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    data_source VARCHAR(100),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE compliance.company_compliance IS 'Relação entre empresas e normas/frameworks/regulamentos aplicáveis';

-- Relação entre Normas/Frameworks e Processos
CREATE TABLE compliance.process_compliance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    process_id UUID,
    compliance_reference_id UUID REFERENCES compliance.compliance_reference(id),
    status VARCHAR(30) DEFAULT 'active',
    valid_from DATE,
    valid_to DATE,
    responsible VARCHAR(100),
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    data_source VARCHAR(100),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE compliance.process_compliance IS 'Relação entre processos e normas/frameworks/regulamentos aplicáveis';
-- O campo process_id deve ser FK para a tabela de processos, a ser criada no domínio de processos.

-- Relação entre Normas/Frameworks e Produtos/Serviços
CREATE TABLE compliance.product_service_compliance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID,
    service_id UUID,
    compliance_reference_id UUID REFERENCES compliance.compliance_reference(id),
    status VARCHAR(30) DEFAULT 'active',
    valid_from DATE,
    valid_to DATE,
    responsible VARCHAR(100),
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    data_source VARCHAR(100),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE compliance.product_service_compliance IS 'Relação entre produtos/serviços e normas/frameworks/regulamentos aplicáveis';
-- Os campos product_id e service_id devem ser FK para as tabelas de produtos e serviços, respectivamente.

-- Índices e constraints adicionais podem ser adicionados conforme necessário para garantir performance e integridade referencial.

-- Triggers, views e funções de auditoria, compliance e validação serão adicionados em arquivos específicos para automação e integração BI.
