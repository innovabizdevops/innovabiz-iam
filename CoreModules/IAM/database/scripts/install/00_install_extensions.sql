-- Script de Instalação das Extensões Necessárias para o IAM
-- Data: 19/05/2025
-- Descrição: Instala as extensões necessárias para o módulo IAM

-- Configuração do ambiente
SET client_encoding = 'UTF8';
SET client_min_messages = warning;

-- Início da transação
BEGIN;

-- Mensagem de início
DO $$
BEGIN
    RAISE NOTICE 'Instalando extensões necessárias para o módulo IAM...';
END
$$;

-- Instalar extensões necessárias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;
CREATE EXTENSION IF NOT EXISTS "pgcrypto" WITH SCHEMA public;
CREATE EXTENSION IF NOT EXISTS "hstore" WITH SCHEMA public;
CREATE EXTENSION IF NOT EXISTS "ltree" WITH SCHEMA public;
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements" WITH SCHEMA public;

-- Mensagem de conclusão
DO $$
BEGIN
    RAISE NOTICE 'Extensões instaladas com sucesso.';
END
$$;

-- Commit da transação
COMMIT;
