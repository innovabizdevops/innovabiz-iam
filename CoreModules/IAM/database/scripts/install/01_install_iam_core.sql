-- Script de Instalação do Módulo IAM - Core
-- Data: 19/05/2025
-- Descrição: Instala as tabelas e funções principais do módulo IAM

-- Configuração do ambiente
SET client_encoding = 'UTF8';
SET client_min_messages = warning;
SET search_path = iam, public;

-- Início da transação
BEGIN;

-- Mensagem de início
RAISE NOTICE 'Instalando módulo IAM Core...';

-- Executar scripts do core na ordem correta
\i ../core/01_schema_iam_core.sql
\i ../core/02_views_iam_core.sql
\i ../core/03_functions_iam_core.sql
\i ../core/04_triggers_iam_core.sql

-- Mensagem de conclusão
RAISE NOTICE 'Módulo IAM Core instalado com sucesso.';

-- Commit da transação
COMMIT;
