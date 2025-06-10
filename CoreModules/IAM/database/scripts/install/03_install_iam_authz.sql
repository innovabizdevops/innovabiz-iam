-- Script de Instalação do Módulo IAM - Autorização
-- Data: 19/05/2025
-- Descrição: Instala as tabelas e funções de autorização do módulo IAM

-- Configuração do ambiente
SET client_encoding = 'UTF8';
SET client_min_messages = warning;
SET search_path = iam, public;

-- Início da transação
BEGIN;

-- Mensagem de início
RAISE NOTICE 'Instalando módulo de Autorização IAM...';

-- Executar scripts de autorização
\i ../core/06_authorization_engine_part1.sql
\i ../core/06_authorization_engine_part2.sql

-- Mensagem de conclusão
RAISE NOTICE 'Módulo de Autorização IAM instalado com sucesso.';

-- Commit da transação
COMMIT;
