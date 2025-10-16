-- Schema para Indicadores/KPIs
CREATE SCHEMA IF NOT EXISTS kpis;

-- Tabela principal de indicadores
CREATE TABLE IF NOT EXISTS kpis.indicator (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(255) NOT NULL,
    descricao TEXT,
    unidade VARCHAR(50),
    tipo VARCHAR(50), -- ex: desempenho, financeiro, operacional
    periodicidade VARCHAR(50), -- ex: mensal, trimestral
    responsavel_id UUID, -- FK para organizations ou users
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabela de metas para indicadores
CREATE TABLE IF NOT EXISTS kpis.target (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    indicator_id UUID REFERENCES kpis.indicator(id),
    valor NUMERIC,
    data_inicio DATE,
    data_fim DATE,
    status VARCHAR(50) DEFAULT 'ativo',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabela de resultados mensurados
CREATE TABLE IF NOT EXISTS kpis.result (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    indicator_id UUID REFERENCES kpis.indicator(id),
    valor NUMERIC,
    data DATE,
    observacao TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
