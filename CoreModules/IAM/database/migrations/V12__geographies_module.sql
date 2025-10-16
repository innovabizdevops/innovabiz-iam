-- Schema para Geografias
CREATE SCHEMA IF NOT EXISTS geographies;

CREATE TABLE IF NOT EXISTS geographies.country (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL,
    official_name VARCHAR(200),
    english_name VARCHAR(200),
    abbreviation VARCHAR(20),
    codigo_iso VARCHAR(10),
    iso_alpha2 CHAR(2),
    iso_alpha3 CHAR(3),
    un_locode VARCHAR(10),
    status VARCHAR(20) DEFAULT 'ativo',
    parent_id UUID REFERENCES geographies.country(id),
    hierarchy_level INT DEFAULT 1,
    data_source VARCHAR(100),
    valid_from DATE,
    valid_to DATE,
    responsible VARCHAR(100),
    latitude NUMERIC(10,8),
    longitude NUMERIC(11,8),
    bounding_box GEOMETRY,
    external_ref VARCHAR(200),
    compliance_status VARCHAR(20),
    main_language VARCHAR(50),
    accessibility VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS geographies.province (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL,
    official_name VARCHAR(200),
    english_name VARCHAR(200),
    abbreviation VARCHAR(20),
    codigo_iso VARCHAR(10),
    iso_alpha2 CHAR(2),
    iso_alpha3 CHAR(3),
    un_locode VARCHAR(10),
    status VARCHAR(20) DEFAULT 'ativo',
    country_id UUID REFERENCES geographies.country(id),
    parent_id UUID REFERENCES geographies.province(id),
    hierarchy_level INT DEFAULT 2,
    data_source VARCHAR(100),
    valid_from DATE,
    valid_to DATE,
    responsible VARCHAR(100),
    latitude NUMERIC(10,8),
    longitude NUMERIC(11,8),
    bounding_box GEOMETRY,
    external_ref VARCHAR(200),
    compliance_status VARCHAR(20),
    main_language VARCHAR(50),
    accessibility VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS geographies.state (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL,
    codigo_uf VARCHAR(10),
    country_id UUID REFERENCES geographies.country(id),
    province_id UUID REFERENCES geographies.province(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS geographies.municipality (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL,
    state_id UUID REFERENCES geographies.state(id),
    province_id UUID REFERENCES geographies.province(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS geographies.district (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL,
    municipality_id UUID REFERENCES geographies.municipality(id),
    state_id UUID REFERENCES geographies.state(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS geographies.county (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL,
    district_id UUID REFERENCES geographies.district(id),
    municipality_id UUID REFERENCES geographies.municipality(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS geographies.parish (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL,
    county_id UUID REFERENCES geographies.county(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS geographies.commune (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL,
    municipality_id UUID REFERENCES geographies.municipality(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS geographies.neighborhood (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL,
    district_id UUID REFERENCES geographies.district(id),
    municipality_id UUID REFERENCES geographies.municipality(id),
    commune_id UUID REFERENCES geographies.commune(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS geographies.city (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL,
    state_id UUID REFERENCES geographies.state(id),
    municipality_id UUID REFERENCES geographies.municipality(id),
    district_id UUID REFERENCES geographies.district(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
