-- Migration de reversão: Remove as tabelas do serviço de elevação de privilégios
-- Autor: Eduardo Jeremias
-- Projeto: INNOVABIZ IAM
-- Descrição: Este script remove as tabelas, índices, visões e funções criadas
-- para o repositório de tokens de elevação.

-- Remove gatilho
DROP TRIGGER IF EXISTS elevation_tokens_usage_trigger ON elevation_tokens;

-- Remove funções
DROP FUNCTION IF EXISTS log_token_usage();
DROP FUNCTION IF EXISTS update_expired_tokens();

-- Remove visão de auditoria
DROP VIEW IF EXISTS elevation_audit_view;

-- Remove tabela de histórico
DROP TABLE IF EXISTS elevation_token_history;

-- Remove tabela principal
DROP TABLE IF EXISTS elevation_tokens;

-- Note: As extensões uuid-ossp e pgcrypto não são removidas pois podem
-- estar sendo utilizadas por outras partes do sistema.