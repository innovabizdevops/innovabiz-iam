-- Script de Instalação do Módulo IAM para Docker
-- Data: 19/05/2025
-- Descrição: Instala o módulo de IAM no banco de dados PostgreSQL em um contêiner Docker

-- Configuração do ambiente
SET client_encoding = 'UTF8';
SET client_min_messages = warning;

-- Início da transação
BEGIN;

-- Criação do esquema IAM se não existir
CREATE SCHEMA IF NOT EXISTS iam;

-- Comentário no esquema
COMMENT ON SCHEMA iam IS 'Esquema para o módulo de Identity and Access Management (IAM)';

-- Configuração de busca para o esquema IAM
SET search_path TO iam, public;

-- Mensagem de início
DO $$
BEGIN
    RAISE NOTICE 'Iniciando instalação do módulo IAM...';
END
$$;

-- Commit da transação
COMMIT;
