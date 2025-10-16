-- Seeds para processos e etapas

INSERT INTO processes.process (id, nome, descricao, status, responsavel_id)
VALUES
  (uuid_generate_v4(), 'Onboarding de Cliente', 'Processo de integração de novos clientes', 'ativo', '11111111-1111-1111-1111-111111111111'),
  (uuid_generate_v4(), 'Implantação de Solução', 'Processo de implantação de solução tecnológica', 'ativo', '22222222-2222-2222-2222-222222222222');

INSERT INTO processes.step (id, process_id, nome, ordem, descricao, responsavel_id, status, inicio_previsto, fim_previsto)
SELECT uuid_generate_v4(), p.id, 'Recepção de Documentos', 1, 'Coleta inicial de documentação', '11111111-1111-1111-1111-111111111111', 'pendente', '2025-07-01', '2025-07-02'
FROM processes.process p WHERE p.nome = 'Onboarding de Cliente';

INSERT INTO processes.step (id, process_id, nome, ordem, descricao, responsavel_id, status, inicio_previsto, fim_previsto)
SELECT uuid_generate_v4(), p.id, 'Configuração de Sistema', 2, 'Configuração dos sistemas para o cliente', '22222222-2222-2222-2222-222222222222', 'pendente', '2025-07-03', '2025-07-05'
FROM processes.process p WHERE p.nome = 'Onboarding de Cliente';

INSERT INTO processes.step (id, process_id, nome, ordem, descricao, responsavel_id, status, inicio_previsto, fim_previsto)
SELECT uuid_generate_v4(), p.id, 'Planejamento', 1, 'Planejamento da implantação', '33333333-3333-3333-3333-333333333333', 'pendente', '2025-08-01', '2025-08-02'
FROM processes.process p WHERE p.nome = 'Implantação de Solução';

INSERT INTO processes.step (id, process_id, nome, ordem, descricao, responsavel_id, status, inicio_previsto, fim_previsto)
SELECT uuid_generate_v4(), p.id, 'Execução Técnica', 2, 'Execução da implantação técnica', '22222222-2222-2222-2222-222222222222', 'pendente', '2025-08-03', '2025-08-10'
FROM processes.process p WHERE p.nome = 'Implantação de Solução';
