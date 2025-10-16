-- Schema para Contratos
CREATE SCHEMA IF NOT EXISTS contracts;

-- Tabela principal de contratos
CREATE TABLE IF NOT EXISTS contracts.contract (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    numero VARCHAR(50) NOT NULL,
    descricao TEXT,
    parte_a_id UUID NOT NULL, -- FK para organizations ou clientes
    parte_b_id UUID NOT NULL, -- FK para organizations ou fornecedores
    data_inicio DATE NOT NULL,
    data_fim DATE,
    status VARCHAR(50) DEFAULT 'ativo',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabela de anexos de contratos
CREATE TABLE IF NOT EXISTS contracts.attachment (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    contract_id UUID REFERENCES contracts.contract(id),
    filename VARCHAR(255),
    file_url TEXT,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
