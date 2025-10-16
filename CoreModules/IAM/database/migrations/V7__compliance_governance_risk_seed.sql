-- V7__compliance_governance_risk_seed.sql
-- Dados de exemplo para validação das estruturas de compliance, governança e risco

-- Compliance: Normas e Frameworks
INSERT INTO compliance.norma (codigo, nome, descricao, tipo_norma, orgao_emissor, data_publicacao)
VALUES ('ISO27001', 'ISO/IEC 27001', 'Norma internacional para segurança da informação', 'Norma', 'ISO', '2013-10-01');

INSERT INTO compliance.framework (nome, descricao, categoria)
VALUES ('COBIT', 'Framework de Governança de TI', 'Governança de TI');

-- Relacionamento entre Norma e Framework
INSERT INTO compliance.norma_framework (norma_id, framework_id)
SELECT n.id, f.id FROM compliance.norma n, compliance.framework f WHERE n.codigo = 'ISO27001' AND f.nome = 'COBIT';

-- Consentimento de Usuário
INSERT INTO compliance.consentimento (usuario_id, tipo_consentimento, dado_referente)
VALUES ('00000000-0000-0000-0000-000000000001', 'Termos de Uso', 'Dados Pessoais');

-- Governança: Tipos de Órgão Social e Órgãos
INSERT INTO governance.tipo_orgao_social (nome, descricao)
VALUES ('Diretoria', 'Órgão executivo principal');

INSERT INTO governance.orgao_social (nome, tipo_orgao_social_id)
SELECT 'Diretoria Executiva', id FROM governance.tipo_orgao_social WHERE nome = 'Diretoria';

-- Papéis e Responsabilidades
INSERT INTO governance.papel (nome, descricao)
VALUES ('CEO', 'Chief Executive Officer');

INSERT INTO governance.responsabilidade (papel_id, entidade, entidade_id)
SELECT p.id, 'orgao_social', o.id FROM governance.papel p, governance.orgao_social o WHERE p.nome = 'CEO' AND o.nome = 'Diretoria Executiva';

-- Risco: Tipos de Risco, Riscos e Planos de Mitigação
INSERT INTO risk.tipo_risco (nome, categoria)
VALUES ('Operacional', 'Processos');

INSERT INTO risk.risco (nome, tipo_risco_id, impacto, probabilidade)
SELECT 'Falha de Processo', id, 'Alto', 'Média' FROM risk.tipo_risco WHERE nome = 'Operacional';

INSERT INTO risk.plano_mitigacao (risco_id, descricao, status)
SELECT r.id, 'Implementar revisão de processos e automação', 'em andamento' FROM risk.risco r WHERE r.nome = 'Falha de Processo';

-- Queries de validação
-- Listar todas as normas e frameworks
SELECT * FROM compliance.norma;
SELECT * FROM compliance.framework;
SELECT * FROM compliance.norma_framework;

-- Listar órgãos sociais e seus tipos
SELECT o.nome AS orgao, t.nome AS tipo FROM governance.orgao_social o JOIN governance.tipo_orgao_social t ON o.tipo_orgao_social_id = t.id;

-- Listar papéis e responsabilidades
SELECT p.nome AS papel, r.entidade, r.entidade_id FROM governance.papel p JOIN governance.responsabilidade r ON p.id = r.papel_id;

-- Listar riscos e planos de mitigação
SELECT r.nome AS risco, t.nome AS tipo, p.descricao AS plano FROM risk.risco r JOIN risk.tipo_risco t ON r.tipo_risco_id = t.id LEFT JOIN risk.plano_mitigacao p ON r.id = p.risco_id;
