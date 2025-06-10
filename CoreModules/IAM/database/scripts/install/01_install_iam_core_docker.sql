-- Script de Instalação do Módulo IAM - Core (Docker)
-- Data: 19/05/2025
-- Descrição: Instala as tabelas e funções principais do módulo IAM

-- Configuração do ambiente
SET client_encoding = 'UTF8';
SET client_min_messages = warning;
SET search_path = iam, public;

-- Início da transação
BEGIN;

-- Mensagem de início
DO $$
BEGIN
    RAISE NOTICE 'Instalando módulo IAM Core...';
END
$$;

-- Executar scripts do core na ordem correta
\i /tmp/core/01_schema_iam_core.sql
\i /tmp/core/02_views_iam_core.sql
\i /tmp/core/03_functions_iam_core.sql
\i /tmp/core/04_triggers_iam_core.sql

-- Mensagem de conclusão
DO $$
BEGIN
    RAISE NOTICE 'Módulo IAM Core instalado com sucesso.';
END
$$;

-- Commit da transação
COMMIT;
