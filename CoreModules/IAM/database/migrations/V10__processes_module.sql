-- Schema para Processos
CREATE SCHEMA IF NOT EXISTS processes;

-- Tabela principal de processos
CREATE TABLE IF NOT EXISTS processes.process (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(255) NOT NULL,
    descricao TEXT,
    status VARCHAR(50) DEFAULT 'ativo',
    responsavel_id UUID, -- FK para organizations ou users
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabela de etapas do processo
CREATE TABLE IF NOT EXISTS processes.step (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    process_id UUID REFERENCES processes.process(id),
    nome VARCHAR(255) NOT NULL,
    ordem INTEGER NOT NULL,
    descricao TEXT,
    responsavel_id UUID, -- FK para organizations ou users
    status VARCHAR(50) DEFAULT 'pendente',
    inicio_previsto DATE,
    fim_previsto DATE,
    inicio_real DATE,
    fim_real DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
