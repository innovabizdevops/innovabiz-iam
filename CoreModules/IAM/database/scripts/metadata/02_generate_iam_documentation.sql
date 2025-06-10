-- INNOVABIZ - IAM Schema Documentation Generator
-- Author: Eduardo Jeremias
-- Date: 19/05/2025
-- Version: 1.0
-- Description: Script para gerar documentação detalhada do esquema IAM em formato Markdown

-- Configuração do ambiente
\set QUIET on
\set ON_ERROR_STOP on
\set VERBOSITY terse
\pset format unaligned
\pset tuples_only on
\pset pager off

-- Nome do arquivo de saída
\o iam_schema_documentation.md

-- Cabeçalho do documento Markdown
SELECT '# Documentação do Esquema IAM
';

SELECT '## Visão Geral
';

SELECT 'O esquema IAM (Identity and Access Management) é responsável por gerenciar identidades, autenticação, autorização e auditoria no sistema INNOVABIZ. Este documento descreve todas as tabelas, visões, funções e outros objetos do esquema IAM.
';

-- Seção de Tabelas
SELECT '## Tabelas

A seguir estão as tabelas que compõem o esquema IAM:

| Nome | Descrição |
|------|-----------|';

SELECT '| ' || table_name || ' | ' || COALESCE(obj_description(('iam.' || table_name)::regclass, 'pg_class'), 'Sem descrição') || ' |'
FROM information_schema.tables 
WHERE table_schema = 'iam' 
  AND table_type = 'BASE TABLE'
ORDER BY table_name;

-- Para cada tabela, listar colunas e detalhes
SELECT '
## Detalhes das Tabelas
';

SELECT '### ' || table_name || E'\n\n' ||
       COALESCE(obj_description(('iam.' || table_name)::regclass, 'pg_class'), 'Sem descrição') || E'\n\n' ||
       '| Coluna | Tipo | Nulo | Padrão | Descrição |\n' ||
       '|--------|------|------|--------|-----------|\n' ||
       string_agg(
           '| ' || 
           column_name || ' | ' || 
           data_type || COALESCE('(' || character_maximum_length::text, '') || 
           CASE WHEN numeric_precision IS NOT NULL THEN 
               '(' || numeric_precision::text || 
               COALESCE(',' || numeric_scale::text, '') || ')' 
           ELSE '' END || ' | ' ||
           CASE WHEN is_nullable = 'NO' THEN 'NÃO' ELSE 'SIM' END || ' | ' ||
           COALESCE(column_default, '') || ' | ' ||
           COALESCE((
               SELECT pg_catalog.col_description(('iam.' || table_name)::regclass::oid, ordinal_position)
           ), '') || ' |',
           E'\n'
       )
FROM information_schema.columns
WHERE table_schema = 'iam'
GROUP BY table_name
ORDER BY table_name;

-- Índices
SELECT '
## Índices

A seguir estão os índices definidos no esquema IAM:

| Tabela | Nome do Índice | Colunas | Único | Descrição |
|--------|----------------|---------|-------|-----------|';

SELECT '| ' || 
       table_name || ' | ' || 
       indexname || ' | ' ||
       array_to_string(array(
           SELECT pg_get_indexdef(idx.indexrelid, k + 1, true)
           FROM generate_subscripts(idx.indkey, 1) as k
           ORDER BY k
       ), ', ') || ' | ' ||
       CASE WHEN idx.indexprs IS NULL THEN 'NÃO' ELSE 'SIM' END || ' | ' ||
       COALESCE(obj_description(idx.indexrelid, 'pg_class'), '') || ' |'
FROM pg_indexes i
JOIN pg_stat_all_indexes sai ON sai.indexrelname = i.indexname
JOIN pg_index idx ON sai.indexrelid = idx.indexrelid
WHERE i.schemaname = 'iam'
ORDER BY table_name, indexname;

-- Chaves Estrangeiras
SELECT '
## Relacionamentos (Chaves Estrangeiras)

A seguir estão os relacionamentos entre as tabelas do esquema IAM:

| Nome da Restrição | Tabela de Origem | Colunas de Origem | Tabela de Destino | Colunas de Destino | Ação ao Atualizar | Ação ao Excluir |
|-------------------|------------------|-------------------|-------------------|-------------------|-------------------|-----------------|';

SELECT '| ' ||
       tc.constraint_name || ' | ' ||
       tc.table_name || ' | ' ||
       kcu.column_name || ' | ' ||
       ccu.table_name || ' | ' ||
       ccu.column_name || ' | ' ||
       rc.update_rule || ' | ' ||
       rc.delete_rule || ' |'
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
  ON tc.constraint_name = kcu.constraint_name
  AND tc.table_schema = kcu.table_schema
JOIN information_schema.referential_constraints rc
  ON tc.constraint_name = rc.constraint_name
JOIN information_schema.constraint_column_usage ccu
  ON rc.unique_constraint_name = ccu.constraint_name
  AND rc.unique_constraint_schema = ccu.constraint_schema
WHERE tc.constraint_type = 'FOREIGN KEY'
  AND tc.table_schema = 'iam'
ORDER BY tc.table_name, kcu.column_name;

-- Funções
SELECT '
## Funções

A seguir estão as funções definidas no esquema IAM:

| Nome | Tipo de Retorno | Descrição |
|------|----------------|-----------|';

SELECT '| ' || 
       p.proname || ' | ' || 
       pg_catalog.format_type(p.prorettype, NULL) || ' | ' ||
       COALESCE(pg_catalog.obj_description(p.oid, 'pg_proc'), '') || ' |'
FROM pg_catalog.pg_proc p
LEFT JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace
WHERE n.nspname = 'iam'
  AND p.prokind = 'f'  -- Apenas funções (não procedimentos)
ORDER BY p.proname;

-- Gatilhos (Triggers)
SELECT '
## Gatilhos (Triggers)

A seguir estão os gatilhos definidos no esquema IAM:

| Tabela | Nome do Gatilho | Função do Gatilho | Momento | Eventos | Condição | Descrição |
|--------|-----------------|-------------------|---------|---------|----------|-----------|';

SELECT '| ' || 
       event_object_table || ' | ' ||
       trigger_name || ' | ' ||
       action_statement || ' | ' ||
       action_timing || ' | ' ||
       event_manipulation || ' | ' ||
       COALESCE(action_condition, '') || ' | ' ||
       COALESCE(action_statement, '') || ' |'
FROM information_schema.triggers
WHERE trigger_schema = 'iam'
ORDER BY event_object_table, trigger_name;

-- Visões (Views)
SELECT '
## Visões (Views)

A seguir estão as visões definidas no esquema IAM:

| Nome | Definição | Descrição |
|------|-----------|-----------|';

SELECT '| ' || 
       table_name || ' | ' ||
       REPLACE(view_definition, '|', '\|') || ' | ' ||
       COALESCE(obj_description(('iam.' || table_name)::regclass, 'pg_class'), '') || ' |'
FROM information_schema.views
WHERE table_schema = 'iam'
ORDER BY table_name;

-- Configurações de segurança
SELECT '
## Configurações de Segurança

### Permissões de Esquema

```sql
';

SELECT 'GRANT ' || privilege_type || ' ON SCHEMA iam TO ' || grantee || ';'
FROM information_schema.role_usage_grants
WHERE object_schema = 'iam'
  AND object_type = 'SCHEMA'
  AND grantee != 'PUBLIC'
ORDER BY grantee, privilege_type;

SELECT '```

### Permissões de Tabela

```sql';

SELECT DISTINCT 'GRANT ' || privilege_type || ' ON TABLE iam.' || table_name || ' TO ' || grantee || ';'
FROM information_schema.role_table_grants
WHERE table_schema = 'iam'
  AND grantee NOT IN ('postgres', 'PUBLIC')
ORDER BY grantee, table_name, privilege_type;

SELECT '```

### Permissões de Sequências

```sql';

SELECT 'GRANT ' || privilege_type || ' ON SEQUENCE iam.' || sequence_name || ' TO ' || grantee || ';'
FROM information_schema.role_usage_grants
WHERE object_schema = 'iam'
  AND object_type = 'SEQUENCE'
  AND grantee != 'PUBLIC'
ORDER BY grantee, sequence_name, privilege_type;

SELECT '```
';

-- Rodapé do documento
SELECT '## Metadados do Documento

- **Gerado em**: ' || now() || '
- **Banco de Dados**: ' || current_database() || '
- **Esquema**: iam
- **Versão do PostgreSQL**: ' || version() || '

---
*Documentação gerada automaticamente pelo script `02_generate_iam_documentation.sql`*';

-- Resetar configurações
\o
