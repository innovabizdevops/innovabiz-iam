-- Schema para Catálogo de Métodos de Autenticação

-- 1. Tabela de Categorias de Métodos
CREATE TABLE IF NOT EXISTS authentication_categories (
    id SERIAL PRIMARY KEY,
    code VARCHAR(10) UNIQUE NOT NULL,
    description TEXT NOT NULL,
    category_type VARCHAR(50) NOT NULL,
    parent_category_id INTEGER REFERENCES authentication_categories(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Tabela de Métodos de Autenticação
CREATE TABLE IF NOT EXISTS authentication_methods (
    id SERIAL PRIMARY KEY,
    method_id VARCHAR(20) UNIQUE NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    security_level VARCHAR(20) NOT NULL,
    irr_level VARCHAR(5) NOT NULL,
    complexity_level VARCHAR(10) NOT NULL,
    maturity_level VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    category_id INTEGER REFERENCES authentication_categories(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Tabela de Casos de Uso
CREATE TABLE IF NOT EXISTS use_cases (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 4. Tabela de Relação Método-Caso de Uso
CREATE TABLE IF NOT EXISTS method_use_case (
    method_id INTEGER REFERENCES authentication_methods(id),
    use_case_id INTEGER REFERENCES use_cases(id),
    PRIMARY KEY (method_id, use_case_id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 5. Tabela de Níveis de Segurança
CREATE TABLE IF NOT EXISTS security_levels (
    id SERIAL PRIMARY KEY,
    level VARCHAR(20) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 6. Tabela de Níveis de Complexidade
CREATE TABLE IF NOT EXISTS complexity_levels (
    id SERIAL PRIMARY KEY,
    level VARCHAR(10) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 7. Tabela de Níveis de Maturidade
CREATE TABLE IF NOT EXISTS maturity_levels (
    id SERIAL PRIMARY KEY,
    level VARCHAR(20) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 8. Tabela de Níveis de IRR
CREATE TABLE IF NOT EXISTS irr_levels (
    id SERIAL PRIMARY KEY,
    level VARCHAR(5) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 9. Tabela de Status
CREATE TABLE IF NOT EXISTS method_status (
    id SERIAL PRIMARY KEY,
    status VARCHAR(20) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 10. Tabela de Setores
CREATE TABLE IF NOT EXISTS sectors (
    id SERIAL PRIMARY KEY,
    code VARCHAR(5) UNIQUE NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 11. Tabela de Relação Método-Setor
CREATE TABLE IF NOT EXISTS method_sector (
    method_id INTEGER REFERENCES authentication_methods(id),
    sector_id INTEGER REFERENCES sectors(id),
    PRIMARY KEY (method_id, sector_id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
