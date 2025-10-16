-- Seeds para contratos e anexos

INSERT INTO contracts.contract (id, numero, descricao, parte_a_id, parte_b_id, data_inicio, data_fim, status)
VALUES
  (uuid_generate_v4(), 'C-2025-001', 'Contrato de prestação de serviços de TI', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', '2025-01-01', '2025-12-31', 'ativo'),
  (uuid_generate_v4(), 'C-2025-002', 'Contrato de fornecimento de equipamentos', '11111111-1111-1111-1111-111111111111', '33333333-3333-3333-3333-333333333333', '2025-03-01', '2025-09-30', 'ativo');

INSERT INTO contracts.attachment (id, contract_id, filename, file_url)
SELECT uuid_generate_v4(), c.id, 'contrato_servicos.pdf', 'https://files.innovabiz.com/contratos/servicos.pdf'
FROM contracts.contract c WHERE c.numero = 'C-2025-001';

INSERT INTO contracts.attachment (id, contract_id, filename, file_url)
SELECT uuid_generate_v4(), c.id, 'contrato_equipamentos.pdf', 'https://files.innovabiz.com/contratos/equipamentos.pdf'
FROM contracts.contract c WHERE c.numero = 'C-2025-002';
