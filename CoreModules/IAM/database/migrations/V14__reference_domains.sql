-- V14__reference_domains.sql
-- Modelagem detalhada dos domínios de referência geográfica: regiões, países, estados, municípios, distritos, bairros, etc.
-- Normas: ISO 3166, UN/LOCODE, IBGE, entre outras

CREATE SCHEMA IF NOT EXISTS reference;

-- Continentes
CREATE TABLE reference.continent (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    iso_code VARCHAR(10),
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    data_source VARCHAR(100),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE reference.continent IS 'Continentes segundo padrões internacionais (ISO 3166, UN/LOCODE)';

-- Países
CREATE TABLE reference.country (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    iso_code VARCHAR(10),
    continent_id UUID REFERENCES reference.continent(id),
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    data_source VARCHAR(100),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE reference.country IS 'Países segundo padrões internacionais (ISO 3166, UN/LOCODE)';

-- Estados/Províncias
CREATE TABLE reference.state (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(10),
    country_id UUID REFERENCES reference.country(id),
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    data_source VARCHAR(100),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE reference.state IS 'Estados ou províncias de cada país';

-- Municípios/Cidades
CREATE TABLE reference.city (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(10),
    state_id UUID REFERENCES reference.state(id),
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    data_source VARCHAR(100),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE reference.city IS 'Municípios ou cidades de cada estado/província';

-- Distritos
CREATE TABLE reference.district (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    city_id UUID REFERENCES reference.city(id),
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    data_source VARCHAR(100),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE reference.district IS 'Distritos de cada município/cidade';

-- Bairros
CREATE TABLE reference.neighborhood (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    district_id UUID REFERENCES reference.district(id),
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    data_source VARCHAR(100),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE reference.neighborhood IS 'Bairros de cada distrito';

-- Outras subdivisões (comuna, freguesia, conselho, etc.) podem ser modeladas de forma semelhante, conforme necessidade regional.

-- Tabelas de relacionamento, triggers, views e funções de auditoria, compliance e validação serão adicionados em arquivos específicos para automação e integração BI.
