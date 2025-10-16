-- Inserções de dados para o Catálogo de Métodos de Autenticação

-- 1. Inserções de Níveis de Segurança
INSERT INTO security_levels (level, description) VALUES
('Básico', 'Nível de segurança básico - 70-80 pontos'),
('Intermediário', 'Nível de segurança intermediário - 80-90 pontos'),
('Avançado', 'Nível de segurança avançado - 90-100 pontos'),
('Muito Avançado', 'Nível de segurança muito avançado - 100-120 pontos');

-- 2. Inserções de Níveis de Complexidade
INSERT INTO complexity_levels (level, description) VALUES
('Baixa', 'Complexidade baixa'),
('Média', 'Complexidade média'),
('Alta', 'Complexidade alta'),
('Muito Alta', 'Complexidade muito alta');

-- 3. Inserções de Níveis de Maturidade
INSERT INTO maturity_levels (level, description) VALUES
('Emergente', 'Tecnologia emergente'),
('Estabelecida', 'Tecnologia estabelecida'),
('Experimental', 'Tecnologia experimental');

-- 4. Inserções de Níveis de IRR
INSERT INTO irr_levels (level, description) VALUES
('R1', 'Risco 1 - Baixo'),
('R2', 'Risco 2 - Médio'),
('R3', 'Risco 3 - Alto'),
('R4', 'Risco 4 - Muito Alto'),
('R5', 'Risco 5 - Crítico');

-- 5. Inserções de Status
INSERT INTO method_status (status, description) VALUES
('Ativo', 'Método em uso'),
('Em Desenvolvimento', 'Método em desenvolvimento'),
('Desativado', 'Método desativado'),
('Experimental', 'Método em fase experimental');

-- 6. Inserções de Categorias Principais
INSERT INTO authentication_categories (code, description, category_type) VALUES
('KB', 'Métodos Baseados em Conhecimento', 'Principal'),
('PB', 'Métodos Baseados em Posse', 'Principal'),
('AF', 'Métodos Anti-Fraude e Comportamental', 'Principal'),
('BM', 'Métodos Biométricos', 'Principal'),
('DT', 'Dispositivos e Tokens de Segurança', 'Principal'),
('SH', 'Métodos Smart e Híbridos', 'Principal'),
('IA', 'Métodos de IA/ML', 'Principal');

-- 7. Inserções de Setores
INSERT INTO sectors (code, name, description) VALUES
('FN', 'Setor Financeiro', 'Métodos específicos para o setor financeiro'),
('HS', 'Setor Saúde', 'Métodos específicos para o setor saúde'),
('GP', 'Setor Público', 'Métodos específicos para o setor público'),
('ED', 'Setor Educação', 'Métodos específicos para o setor educação'),
('EN', 'Setor Energia', 'Métodos específicos para o setor energia'),
('AG', 'Setor Agronegócio', 'Métodos específicos para o setor agronegócio'),
('CO', 'Setor Construção', 'Métodos específicos para o setor construção'),
('TR', 'Setor Transporte', 'Métodos específicos para o setor transporte'),
('TU', 'Setor Turismo', 'Métodos específicos para o setor turismo'),
('LA', 'Setor Lazer', 'Métodos específicos para o setor lazer'),
('CM', 'Setor Comércio', 'Métodos específicos para o setor comércio');

-- 8. Inserções de Casos de Uso
INSERT INTO use_cases (name, category, description) VALUES
('Mobile', 'Dispositivos Móveis', 'Casos de uso em dispositivos móveis'),
('Enterprise', 'Empresas', 'Casos de uso em empresas'),
('Governo', 'Governo', 'Casos de uso em órgãos governamentais'),
('Bancos', 'Bancos', 'Casos de uso em instituições bancárias'),
('Saúde', 'Saúde', 'Casos de uso em saúde'),
('Educação', 'Educação', 'Casos de uso em educação'),
('Comércio', 'Comércio', 'Casos de uso em comércio'),
('Transporte', 'Transporte', 'Casos de uso em transporte'),
('Turismo', 'Turismo', 'Casos de uso em turismo'),
('Lazer', 'Lazer', 'Casos de uso em lazer');
