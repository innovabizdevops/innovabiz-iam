-- INNOVABIZ - IAM Performance Analysis Report
-- Author: Eduardo Jeremias
-- Date: 19/05/2025
-- Version: 1.0
-- Description: Script para gerar relatório de análise de desempenho do esquema IAM

-- Configuração do ambiente
\set QUIET on
\set ON_ERROR_STOP on
\set VERBOSITY terse
\pset format unaligned
\pset tuples_only on
\pset pager off

-- Nome do arquivo de saída
\o iam_performance_analysis_report.md

-- Cabeçalho do relatório
SELECT '# Relatório de Análise de Desempenho - Esquema IAM
';

SELECT '## Visão Geral
';

SELECT 'Este relatório contém uma análise detalhada do desempenho do esquema IAM, incluindo estatísticas de tabelas, índices, consultas e recomendações de otimização.
';

SELECT '**Data de Geração:** ' || now() || '  
';
SELECT '**Banco de Dados:** ' || current_database() || '  
';
SELECT '**Versão do PostgreSQL:** ' || version() || '  
';

-- Seção de Estatísticas de Tabelas
SELECT '
## Estatísticas de Tabelas

A seguir estão as estatísticas de desempenho das tabelas do esquema IAM:

| Tabela | Linhas | Tamanho | Índices | Tamanho dos Índices | Vazios | Atualização mais Recente |
|--------|--------|---------|---------|---------------------|--------|--------------------------|';

SELECT '| ' || 
       table_name || ' | ' ||
       to_char(reltuples, '999,999,999') || ' | ' ||
       pg_size_pretty(pg_total_relation_size(quote_ident(table_name))) || ' | ' ||
       (SELECT COUNT(*) FROM pg_indexes WHERE tablename = t.table_name AND schemaname = 'iam') || ' | ' ||
       pg_size_pretty(COALESCE(pg_indexes_size(quote_ident(table_name)), 0)) || ' | ' ||
       CASE WHEN n_dead_tup > 0 THEN 'Sim (' || n_dead_tup || ')' ELSE 'Não' END || ' | ' ||
       COALESCE(to_char(last_autovacuum, 'YYYY-MM-DD HH24:MI'), 'Nunca') || ' |'
FROM information_schema.tables t
JOIN pg_stat_user_tables s ON s.relname = t.table_name
WHERE t.table_schema = 'iam' AND t.table_type = 'BASE TABLE'
ORDER BY pg_total_relation_size(quote_ident(table_name)) DESC;

-- Seção de Índices Não Utilizados
SELECT '
## Índices Potencialmente Não Utilizados

Os seguintes índices podem não estar sendo utilizados ou podem ser ineficientes:

| Tabela | Índice | Tamanho | Número de Varrimentos | Número de Leituras | Última Leitura |
|--------|--------|---------|----------------------|-------------------|----------------|';

SELECT '| ' || 
       t.tablename || ' | ' ||
       t.indexname || ' | ' ||
       pg_size_pretty(pg_relation_size(quote_ident(t.indexname))) || ' | ' ||
       s.idx_scan || ' | ' ||
       s.idx_tup_read || ' | ' ||
       COALESCE(to_char(s.last_idx_scan, 'YYYY-MM-DD HH24:MI'), 'Nunca') || ' |'
FROM pg_stat_user_indexes t
JOIN pg_stat_user_tables t2 ON t.schemaname = t2.schemaname AND t.relname = t2.relname
JOIN pg_index i ON i.indexrelid = t.indexrelid
JOIN pg_stat_all_indexes s ON s.indexrelid = t.indexrelid
WHERE t.schemaname = 'iam'
  AND t.idx_scan < 100  -- Índices com menos de 100 varreduras
  AND t.idx_scan::float / (COALESCE(NULLIF(t2.n_tup_ins, 0), 1) + COALESCE(NULLIF(t2.n_tup_upd, 0), 1) - COALESCE(NULLIF(t2.n_tup_hot_upd, 0), 1) + COALESCE(NULLIF(t2.n_tup_del, 0), 1)) < 0.01  -- Baixa taxa de uso
  AND NOT i.indisunique  -- Não incluir índices únicos
  AND NOT EXISTS (  -- Não incluir índices que são chaves estrangeiras
    SELECT 1 
    FROM pg_constraint c 
    WHERE c.conindid = t.indexrelid
  )
ORDER BY pg_relation_size(t.indexrelid) DESC;

-- Seção de Consultas Lentas
SELECT '
## Consultas Lentas

As seguintes consultas podem estar com problemas de desempenho:

| Média de Tempo (ms) | Chamadas | Tempo Total (s) | Consulta |
|---------------------|----------|-----------------|----------|';

SELECT '| ' ||
       ROUND((total_time / 1000 / total_calls)::numeric, 2) || ' | ' ||
       total_calls || ' | ' ||
       ROUND((total_time / 1000)::numeric, 2) || ' | `' ||
       REPLACE(REPLACE(LEFT(query, 150), '`', ''''), '|', '\|') || '` |'
FROM pg_stat_statements
WHERE query !~* '^(''|BEGIN|COMMIT|SET|SHOW|SELECT pg_catalog\.|SELECT version\(|SELECT current_setting\()'
  AND query !~* '^SELECT.*FROM pg_'
  AND query !~* '^SELECT.*FROM information_schema\.'
  AND query ~* 'FROM iam\.'
ORDER BY (total_time / 1000 / total_calls) DESC
LIMIT 10;

-- Seção de Bloqueios
SELECT '
## Bloqueios Ativos

Os seguintes bloqueios estão atualmente ativos no banco de dados:

| PID | Usuário | Aplicação | Duração | Estado | Consulta |
|-----|---------|-----------|---------|--------|----------|';

SELECT '| ' ||
       pid || ' | ' ||
       usename || ' | ' ||
       COALESCE(application_name, '') || ' | ' ||
       age(now(), query_start) || ' | ' ||
       state || ' | `' ||
       COALESCE(SUBSTRING(REPLACE(REPLACE(query, '\n', ' '), '`', '''') FROM 1 FOR 100), '') || '` |'
FROM pg_stat_activity
WHERE datname = current_database()
  AND pid != pg_backend_pid()
  AND state != 'idle'
  AND query_start < (now() - INTERVAL '5 seconds')
ORDER BY query_start;

-- Seção de Recomendações
SELECT '
## Recomendações de Otimização

### 1. Otimização de Índices
';

-- Recomendações para adicionar índices
SELECT '- **Criar índice** para a coluna `' || 
       a.attname || '` na tabela `' || 
       n.nspname || '.' || 
       t.relname || '` para melhorar consultas que filtram por esta coluna.
'
FROM pg_class t
JOIN pg_namespace n ON n.oid = t.relnamespace
JOIN pg_attribute a ON a.attrelid = t.oid
LEFT JOIN pg_index i ON i.indrelid = t.oid AND a.attnum = ANY(i.indkey)
LEFT JOIN pg_stat_all_tables st ON t.relname = st.relname AND n.nspname = st.schemaname
WHERE n.nspname = 'iam'
  AND t.relkind = 'r'
  AND a.attnum > 0
  AND NOT a.attisdropped
  AND i.indexrelid IS NULL
  AND a.attname IN ('user_id', 'organization_id', 'role_id', 'created_at', 'updated_at', 'status')
  AND st.seq_scan > 1000  -- Tabelas com mais de 1000 varreduras sequenciais
GROUP BY n.nspname, t.relname, a.attname, st.seq_scan
ORDER BY st.seq_scan DESC
LIMIT 5;

-- Recomendações para remover índices não utilizados
SELECT '- **Remover índice** `' || 
       t.indexname || '` da tabela `' || 
       t.tablename || '` pois possui baixa utilização (' || 
       COALESCE(s.idx_scan, 0) || ' varreduras).
'
FROM pg_stat_user_indexes t
JOIN pg_stat_user_tables t2 ON t.schemaname = t2.schemaname AND t.relname = t2.relname
JOIN pg_index i ON i.indexrelid = t.indexrelid
LEFT JOIN pg_stat_all_indexes s ON s.indexrelid = t.indexrelid
WHERE t.schemaname = 'iam'
  AND t.idx_scan < 10  -- Índices com menos de 10 varreduras
  AND NOT i.indisunique  -- Não incluir índices únicos
  AND NOT EXISTS (  -- Não incluir índices que são chaves estrangeiras
    SELECT 1 
    FROM pg_constraint c 
    WHERE c.conindid = t.indexrelid
  )
ORDER BY pg_relation_size(t.indexrelid) DESC
LIMIT 5;

-- Recomendações para VACUUM/ANALYZE
SELECT '
### 2. Manutenção de Tabelas
';

SELECT '- Executar `VACUUM ANALYZE ' || 
       table_schema || '.' || 
       table_name || ';` para atualizar estatísticas e limpar registros mortos. ' ||
       COALESCE('Último VACUUM: ' || last_vacuum || ', ', 'Nunca passou por VACUUM, ') ||
       COALESCE('Último ANALYZE: ' || last_analyze || '.', 'Nunca passou por ANALYZE.') || 
       ' Tamanho: ' || pg_size_pretty(pg_total_relation_size(quote_ident(table_name))) ||
       ', Linhas: ' || reltuples::bigint ||
       ', Registros Mortos: ' || n_dead_tup || '.
'
FROM information_schema.tables t
JOIN pg_stat_user_tables s ON s.relname = t.table_name
JOIN pg_class c ON c.relname = t.table_name
JOIN pg_namespace n ON n.oid = c.relnamespace AND n.nspname = t.table_schema
WHERE t.table_schema = 'iam' 
  AND t.table_type = 'BASE TABLE'
  AND (
    (n_dead_tup > 1000 AND n_dead_tup > 0.1 * reltuples) OR  -- Mais de 1000 registros mortos ou mais de 10% de registros mortos
    last_autovacuum IS NULL OR  -- Nunca passou por VACUUM
    (now() - last_autovacuum) > INTERVAL '7 days'  -- Último VACUUM há mais de 7 dias
  )
ORDER BY n_dead_tup DESC
LIMIT 5;

-- Recomendações de Configuração
SELECT '
### 3. Ajustes de Configuração

Baseado na análise do esquema IAM, considere os seguintes ajustes de configuração do PostgreSQL:
';

-- Configurações de memória
SELECT '- **shared_buffers**: Aumente para 25% da memória total do servidor (atual: ' || 
       (SELECT setting || ' ' || unit FROM pg_settings WHERE name = 'shared_buffers') || ')
';

SELECT '- **work_mem**: Aumente para melhorar ordenações e junções (atual: ' || 
       (SELECT setting || ' ' || unit FROM pg_settings WHERE name = 'work_mem') || ')
';

SELECT '- **maintenance_work_mem**: Aumente para operações de manutenção (atual: ' || 
       (SELECT setting || ' ' || unit FROM pg_settings WHERE name = 'maintenance_work_mem') || ')
';

-- Configurações de VACUUM
SELECT '- **autovacuum_vacuum_scale_factor**: Considere reduzir para ' || 
       (SELECT ROUND(setting::numeric * 0.5, 2) || ' para VACUUMs mais frequentes (atual: ' || setting || ')' 
        FROM pg_settings 
        WHERE name = 'autovacuum_vacuum_scale_factor') || '
';

-- Rodapé do relatório
SELECT '
## Conclusão

Este relatório destaca as principais áreas que podem se beneficiar de otimizações no esquema IAM. Recomenda-se:

1. Revisar e implementar as recomendações de índices
2. Agendar manutenção regular das tabelas com VACUUM e ANALYZE
3. Ajustar as configurações do PostgreSQL conforme sugerido
4. Monitorar o desempenho após as alterações

**Nota:** Sempre teste as alterações em um ambiente de não produção antes de aplicar em produção.
';

-- Resetar configurações
\o
