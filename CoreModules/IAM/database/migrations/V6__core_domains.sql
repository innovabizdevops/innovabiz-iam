-- V6__core_domains.sql
-- Estrutura inicial dos principais domínios de negócio para a base de dados Innovabiz

CREATE SCHEMA IF NOT EXISTS corporate;
CREATE SCHEMA IF NOT EXISTS products;
CREATE SCHEMA IF NOT EXISTS services;

-- Tipos de Empresa
CREATE TABLE IF NOT EXISTS corporate.tipo_empresa (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL,
    descricao TEXT
);

-- Empresas
CREATE TABLE IF NOT EXISTS corporate.empresa (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(255) NOT NULL,
    tipo_empresa_id UUID REFERENCES corporate.tipo_empresa(id),
    cnpj VARCHAR(20),
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Acionistas
CREATE TABLE IF NOT EXISTS corporate.acionista (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(255) NOT NULL,
    tipo_acionista_id UUID,
    empresa_id UUID REFERENCES corporate.empresa(id),
    percentual NUMERIC(5,2),
    status VARCHAR(50) DEFAULT 'active'
);

-- Clientes
CREATE TABLE IF NOT EXISTS corporate.cliente (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(255) NOT NULL,
    tipo_cliente_id UUID,
    empresa_id UUID REFERENCES corporate.empresa(id),
    status VARCHAR(50) DEFAULT 'active'
);

-- Tipos de Módulo
CREATE TABLE IF NOT EXISTS corporate.tipo_modulo (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL
);

-- Módulos
CREATE TABLE IF NOT EXISTS corporate.modulo (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL,
    tipo_modulo_id UUID REFERENCES corporate.tipo_modulo(id)
);

-- Relacionamento Empresa-Módulo
CREATE TABLE IF NOT EXISTS corporate.empresa_modulo (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    empresa_id UUID REFERENCES corporate.empresa(id),
    modulo_id UUID REFERENCES corporate.modulo(id)
);

-- Tipos de Modelo de Negócio
CREATE TABLE IF NOT EXISTS corporate.tipo_modelo_negocio (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL
);

-- Modelos de Negócio
CREATE TABLE IF NOT EXISTS corporate.modelo_negocio (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL,
    tipo_modelo_negocio_id UUID REFERENCES corporate.tipo_modelo_negocio(id)
);

-- Categorias de Produto
CREATE TABLE IF NOT EXISTS products.categoria_produto (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL
);

-- Produtos
CREATE TABLE IF NOT EXISTS products.produto (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL,
    categoria_id UUID REFERENCES products.categoria_produto(id)
);

-- Categorias de Serviço
CREATE TABLE IF NOT EXISTS services.categoria_servico (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL
);

-- Serviços
CREATE TABLE IF NOT EXISTS services.servico (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL,
    categoria_id UUID REFERENCES services.categoria_servico(id)
);

-- Relacionamento Produto-Serviço (para produtos que são também serviços)
CREATE TABLE IF NOT EXISTS products.produto_servico (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    produto_id UUID REFERENCES products.produto(id),
    servico_id UUID REFERENCES services.servico(id)
);

-- Continuação: outros domínios podem ser modelados seguindo este padrão.
-- Para cada domínio, crie tabelas de tipos, entidades e relacionamentos conforme necessidade.
