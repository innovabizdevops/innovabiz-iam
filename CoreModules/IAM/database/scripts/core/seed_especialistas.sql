-- Popula a tabela de especialistas humanos com exemplos práticos
INSERT INTO equipe_especialistas (nome, funcao, area_atuacao, certificacoes, contato, disponibilidade, observacoes)
VALUES
('Ana Silva', 'Compliance Officer', 'Governança', 'ISO 27001, LGPD', 'ana@empresa.com', TRUE, 'Responsável LGPD'),
('Carlos Souza', 'Engenheiro DevOps', 'Tecnologia', 'AWS, Azure, DevSecOps', 'carlos@empresa.com', TRUE, 'DevSecOps pipelines'),
('João Lima', 'Especialista Setorial', 'Saúde', 'HIPAA, ISO 27799', 'joao@empresa.com', FALSE, 'Consultor externo');

-- Exemplo de atribuição de responsável humano a um validador digital
UPDATE compliance_validators
SET owner = 'Ana Silva'
WHERE code = 'GDPRValidator';

UPDATE compliance_validators
SET owner = 'Carlos Souza'
WHERE code = 'ISO27001ControlChecker';
