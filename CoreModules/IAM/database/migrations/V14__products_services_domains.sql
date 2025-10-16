-- V14__products_services_domains.sql
-- Modelagem detalhada dos domínios de produtos, serviços, categorias, tipos, distribuição, canais e mercado
-- Normas: ISO 9001, ISO 8000, DAMA-DMBOK, BIAN, ITIL, entre outras

CREATE SCHEMA IF NOT EXISTS products;
CREATE SCHEMA IF NOT EXISTS services;
CREATE SCHEMA IF NOT EXISTS distribution;
CREATE SCHEMA IF NOT EXISTS channel;

-- Categorias de Produtos
CREATE TABLE products.product_category (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_id UUID,
    hierarchy_level INT DEFAULT 1,
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
COMMENT ON TABLE products.product_category IS 'Categorias de produtos segundo padrões internacionais';

-- Tipos de Produtos
CREATE TABLE products.product_type (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category_id UUID REFERENCES products.product_category(id),
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
COMMENT ON TABLE products.product_type IS 'Tipos de produtos';

-- Produtos
CREATE TABLE products.product (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    description TEXT,
    product_type_id UUID REFERENCES products.product_type(id),
    category_id UUID REFERENCES products.product_category(id),
    status VARCHAR(30) DEFAULT 'active',
    launch_date DATE,
    discontinue_date DATE,
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
COMMENT ON TABLE products.product IS 'Produtos oferecidos pela organização';

-- Categorias de Serviços
CREATE TABLE services.service_category (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_id UUID,
    hierarchy_level INT DEFAULT 1,
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
COMMENT ON TABLE services.service_category IS 'Categorias de serviços segundo padrões internacionais';

-- Tipos de Serviços
CREATE TABLE services.service_type (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category_id UUID REFERENCES services.service_category(id),
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
COMMENT ON TABLE services.service_type IS 'Tipos de serviços';

-- Serviços
CREATE TABLE services.service (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    description TEXT,
    service_type_id UUID REFERENCES services.service_type(id),
    category_id UUID REFERENCES services.service_category(id),
    status VARCHAR(30) DEFAULT 'active',
    launch_date DATE,
    discontinue_date DATE,
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
COMMENT ON TABLE services.service IS 'Serviços oferecidos pela organização';

-- Tipos de Distribuição
CREATE TABLE distribution.distribution_type (
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
COMMENT ON TABLE distribution.distribution_type IS 'Tipos de distribuição de produtos e serviços';

-- Distribuições
CREATE TABLE distribution.distribution (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    distribution_type_id UUID REFERENCES distribution.distribution_type(id),
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
COMMENT ON TABLE distribution.distribution IS 'Distribuição de produtos e serviços';

-- Tipos de Canais
CREATE TABLE channel.channel_type (
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
COMMENT ON TABLE channel.channel_type IS 'Tipos de canais de distribuição';

-- Canais
CREATE TABLE channel.channel (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    channel_type_id UUID REFERENCES channel.channel_type(id),
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
COMMENT ON TABLE channel.channel IS 'Canais de distribuição de produtos e serviços';

-- Relacionamentos e integrações adicionais podem ser criados conforme necessidade, incluindo tabelas N:N para produtos-serviços, produtos-canais, etc.

-- Triggers, views e funções de auditoria, compliance e validação serão adicionados em arquivos específicos para automação e integração BI.
