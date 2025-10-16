-- V17__supply_chain_domains.sql
-- Modelagem detalhada do domínio Supply Chain (Cadeia de Suprimentos)
-- Padrões: ISO 28000, GS1, SCOR, GRI, entre outros

CREATE SCHEMA IF NOT EXISTS supply_chain;

-- Fornecedores
CREATE TABLE supply_chain.supplier (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    tax_id VARCHAR(30),
    country_id UUID REFERENCES reference.country(id),
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    compliance_status VARCHAR(20),
    risk_level VARCHAR(30),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE supply_chain.supplier IS 'Fornecedores da cadeia de suprimentos conforme ISO 28000, GS1, SCOR';

-- Contratos logísticos
CREATE TABLE supply_chain.logistics_contract (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_id UUID REFERENCES supply_chain.supplier(id),
    contract_number VARCHAR(50) NOT NULL,
    start_date DATE,
    end_date DATE,
    value NUMERIC,
    status VARCHAR(30) DEFAULT 'active',
    compliance_status VARCHAR(20),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    external_ref VARCHAR(200)
);
COMMENT ON TABLE supply_chain.logistics_contract IS 'Contratos logísticos com fornecedores';

-- Rastreamento de lotes
CREATE TABLE supply_chain.batch_tracking (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id UUID REFERENCES supply_chain.logistics_contract(id),
    product_id UUID REFERENCES products.product(id),
    batch_code VARCHAR(50) NOT NULL,
    production_date DATE,
    expiry_date DATE,
    status VARCHAR(30) DEFAULT 'active',
    compliance_status VARCHAR(20),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    external_ref VARCHAR(200)
);
COMMENT ON TABLE supply_chain.batch_tracking IS 'Rastreamento de lotes/produtos na cadeia de suprimentos';

-- Compliance e auditoria de fornecedores
CREATE TABLE supply_chain.supplier_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_id UUID REFERENCES supply_chain.supplier(id),
    audit_date DATE NOT NULL,
    result VARCHAR(50),
    auditor VARCHAR(100),
    findings TEXT,
    compliance_status VARCHAR(20),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    external_ref VARCHAR(200)
);
COMMENT ON TABLE supply_chain.supplier_audit IS 'Auditorias e compliance de fornecedores';

-- Triggers, views e funções de auditoria/compliance seguem o padrão dos demais domínios.
