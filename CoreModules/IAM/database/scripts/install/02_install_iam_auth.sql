-- Script de Instalação do Módulo IAM - Autenticação
-- Data: 19/05/2025
-- Descrição: Instala as tabelas e funções de autenticação do módulo IAM

-- Configuração do ambiente
SET client_encoding = 'UTF8';
SET client_min_messages = warning;
SET search_path = iam, public;

-- Início da transação
BEGIN;

-- Mensagem de início
RAISE NOTICE 'Instalando módulo de Autenticação IAM...';

-- Executar scripts de autenticação MFA
\i ../core/05_mfa_authentication_part1.sql
\i ../core/05_mfa_authentication_part2.sql
\i ../core/05_mfa_authentication_ar_integration.sql

-- Mensagem de conclusão
RAISE NOTICE 'Módulo de Autenticação IAM instalado com sucesso.';

-- Commit da transação
COMMIT;
