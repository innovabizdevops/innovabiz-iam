-- Seeds para Indicadores/KPIs, Metas e Resultados

INSERT INTO kpis.indicator (id, nome, descricao, unidade, tipo, periodicidade, responsavel_id)
VALUES
  (uuid_generate_v4(), 'NPS', 'Net Promoter Score', 'pontos', 'desempenho', 'mensal', '11111111-1111-1111-1111-111111111111'),
  (uuid_generate_v4(), 'Receita Mensal', 'Receita bruta mensal', 'BRL', 'financeiro', 'mensal', '22222222-2222-2222-2222-222222222222');

INSERT INTO kpis.target (id, indicator_id, valor, data_inicio, data_fim, status)
SELECT uuid_generate_v4(), i.id, 70, '2025-01-01', '2025-12-31', 'ativo'
FROM kpis.indicator i WHERE i.nome = 'NPS';

INSERT INTO kpis.target (id, indicator_id, valor, data_inicio, data_fim, status)
SELECT uuid_generate_v4(), i.id, 100000, '2025-01-01', '2025-12-31', 'ativo'
FROM kpis.indicator i WHERE i.nome = 'Receita Mensal';

INSERT INTO kpis.result (id, indicator_id, valor, data, observacao)
SELECT uuid_generate_v4(), i.id, 68, '2025-02-01', 'NPS abaixo da meta'
FROM kpis.indicator i WHERE i.nome = 'NPS';

INSERT INTO kpis.result (id, indicator_id, valor, data, observacao)
SELECT uuid_generate_v4(), i.id, 105000, '2025-02-01', 'Receita acima da meta'
FROM kpis.indicator i WHERE i.nome = 'Receita Mensal';
