-- INNOVABIZ - IAM Schema Migration Manager
-- Author: Eduardo Jeremias
-- Date: 19/05/2025
-- Version: 1.0
-- Description: Script para gerenciar migrações de esquema do IAM

-- Configuração do ambiente
SET search_path TO iam, public;
SET statement_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

-- Tabela para rastrear migrações
CREATE TABLE IF NOT EXISTS schema_migrations (
    id SERIAL PRIMARY KEY,
    version VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    script_name VARCHAR(255) NOT NULL,
    installed_by VARCHAR(100) NOT NULL,
    installed_on TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    execution_time_ms BIGINT,
    success BOOLEAN NOT NULL,
    checksum VARCHAR(64),
    metadata JSONB,
    CONSTRAINT uq_schema_migrations_version UNIQUE (version)
);

-- Comentários para documentação
COMMENT ON TABLE schema_migrations IS 'Registra todas as migrações de esquema aplicadas ao banco de dados';
COMMENT ON COLUMN schema_migrations.version IS 'Identificador único da versão da migração (formato: YYYY.MM.DD.HHMM)';
COMMENT ON COLUMN schema_migrations.description IS 'Descrição da migração';
COMMENT ON COLUMN schema_migrations.script_name IS 'Nome do script de migração';
COMMENT ON COLUMN schema_migrations.installed_by IS 'Usuário que aplicou a migração';
COMMENT ON COLUMN schema_migrations.installed_on IS 'Data e hora em que a migração foi aplicada';
COMMENT ON COLUMN schema_migrations.execution_time_ms IS 'Tempo de execução da migração em milissegundos';
COMMENT ON COLUMN schema_migrations.success IS 'Indica se a migração foi aplicada com sucesso';
COMMENT ON COLUMN schema_migrations.checksum IS 'Hash SHA-256 do script de migração para verificação de integridade';
COMMENT ON COLUMN schema_migrations.metadata IS 'Metadados adicionais sobre a migração';

-- Índices para otimização
CREATE INDEX IF NOT EXISTS idx_schema_migrations_version ON schema_migrations(version);
CREATE INDEX IF NOT EXISTS idx_schema_migrations_installed_on ON schema_migrations(installed_on);

-- Função para aplicar uma migração
CREATE OR REPLACE FUNCTION apply_migration(
    p_version VARCHAR(50),
    p_description TEXT,
    p_script_name VARCHAR(255),
    p_sql TEXT
) RETURNS BOOLEAN AS $$
DECLARE
    v_start_time TIMESTAMP WITH TIME ZONE;
    v_end_time TIMESTAMP WITH TIME ZONE;
    v_execution_time_ms BIGINT;
    v_checksum TEXT;
    v_success BOOLEAN := FALSE;
    v_error_message TEXT;
    v_error_detail TEXT;
    v_error_hint TEXT;
    v_error_context TEXT;
    v_current_user TEXT;
    v_db_version TEXT;
BEGIN
    -- Verificar se a migração já foi aplicada
    IF EXISTS (SELECT 1 FROM schema_migrations WHERE version = p_version) THEN
        RAISE NOTICE 'A migração % já foi aplicada anteriormente', p_version;
        RETURN TRUE;
    END IF;
    
    -- Obter informações do usuário atual
    SELECT current_user INTO v_current_user;
    
    -- Obter versão do PostgreSQL
    SELECT version() INTO v_db_version;
    
    -- Iniciar transação
    BEGIN
        v_start_time := clock_timestamp();
        
        -- Executar o script SQL da migração
        EXECUTE p_sql;
        
        v_end_time := clock_timestamp();
        v_execution_time_ms := EXTRACT(EPOCH FROM (v_end_time - v_start_time)) * 1000;
        
        -- Calcular checksum do script SQL
        v_checksum := encode(digest(p_sql, 'sha256'), 'hex');
        
        -- Registrar migração bem-sucedida
        INSERT INTO schema_migrations (
            version, 
            description, 
            script_name, 
            installed_by, 
            installed_on, 
            execution_time_ms, 
            success, 
            checksum,
            metadata
        ) VALUES (
            p_version,
            p_description,
            p_script_name,
            v_current_user,
            v_start_time,
            v_execution_time_ms,
            TRUE,
            v_checksum,
            jsonb_build_object(
                'postgres_version', v_db_version,
                'applied_at', v_start_time,
                'environment', current_setting('application_name', true),
                'client_addr', inet_client_addr()::TEXT,
                'client_port', inet_client_port()
            )
        );
        
        v_success := TRUE;
        RAISE NOTICE 'Migração % aplicada com sucesso em % ms', p_version, v_execution_time_ms;
        
    EXCEPTION WHEN OTHERS THEN
        v_end_time := clock_timestamp();
        v_execution_time_ms := EXTRACT(EPOCH FROM (v_end_time - v_start_time)) * 1000;
        
        -- Obter detalhes do erro
        GET STACKED DIAGNOSTICS 
            v_error_message = MESSAGE_TEXT,
            v_error_detail = PG_EXCEPTION_DETAIL,
            v_error_hint = PG_EXCEPTION_HINT,
            v_error_context = PG_EXCEPTION_CONTEXT;
        
        -- Registrar falha na migração
        INSERT INTO schema_migrations (
            version, 
            description, 
            script_name, 
            installed_by, 
            installed_on, 
            execution_time_ms, 
            success, 
            checksum,
            metadata
        ) VALUES (
            p_version,
            p_description,
            p_script_name,
            v_current_user,
            v_start_time,
            v_execution_time_ms,
            FALSE,
            NULL,
            jsonb_build_object(
                'error_message', v_error_message,
                'error_detail', v_error_detail,
                'error_hint', v_error_hint,
                'error_context', v_error_context,
                'postgres_version', v_db_version,
                'environment', current_setting('application_name', true),
                'client_addr', inet_client_addr()::TEXT,
                'client_port', inet_client_port()
            )
        );
        
        RAISE EXCEPTION 'Falha ao aplicar migração %: % (Detalhes: %, Dica: %, Contexto: %)', 
            p_version, v_error_message, v_error_detail, v_error_hint, v_error_context;
    END;
    
    RETURN v_success;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para verificar o status de migração
CREATE OR REPLACE FUNCTION check_migration_status()
RETURNS TABLE (
    version VARCHAR(50),
    description TEXT,
    installed_on TIMESTAMP WITH TIME ZONE,
    execution_time_ms BIGINT,
    success BOOLEAN,
    installed_by VARCHAR(100),
    script_name VARCHAR(100),
    checksum_match BOOLEAN
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        sm.version,
        sm.description,
        sm.installed_on,
        sm.execution_time_ms,
        sm.success,
        sm.installed_by,
        sm.script_name,
        CASE 
            WHEN sm.checksum IS NULL THEN NULL
            ELSE (SELECT checksum = encode(digest(pg_get_functiondef(oid)::TEXT, 'sha256'), 'hex') 
                  FROM pg_proc 
                  WHERE proname = 'apply_migration' 
                  LIMIT 1)
        END AS checksum_match
    FROM schema_migrations sm
    ORDER BY sm.installed_on DESC;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Função para gerar script de rollback para uma versão específica
CREATE OR REPLACE FUNCTION generate_rollback_script(p_version VARCHAR(50))
RETURNS TEXT AS $$
DECLARE
    v_rollback_sql TEXT := '';
    v_migration_record RECORD;
BEGIN
    -- Obter informações da migração
    SELECT * INTO v_migration_record 
    FROM schema_migrations 
    WHERE version = p_version 
    AND success = TRUE
    LIMIT 1;
    
    IF v_migration_record IS NULL THEN
        RAISE EXCEPTION 'Nenhuma migração bem-sucedida encontrada para a versão %', p_version;
    END IF;
    
    -- Gerar script de rollback baseado no tipo de migração
    -- Esta é uma implementação básica que pode ser estendida conforme necessário
    v_rollback_sql := '-- Script de Rollback para versão ' || p_version || E'\n';
    v_rollback_sql := v_rollback_sql || '-- Descrição: ' || v_migration_record.description || E'\n';
    v_rollback_sql := v_rollback_sql || '-- Aplicado em: ' || v_migration_record.installed_on || E'\n\n';
    
    -- Adicionar instruções de rollback baseadas no nome do script
    -- Esta é uma abordagem simplificada - em um cenário real, você teria scripts de rollback específicos
    IF v_migration_record.script_name LIKE '%add_column%' THEN
        v_rollback_sql := v_rollback_sql || '-- Para reverter uma adição de coluna, execute algo como:\n';
        v_rollback_sql := v_rollback_sql || '-- ALTER TABLE nome_da_tabela DROP COLUMN IF EXISTS nome_da_coluna;' || E'\n';
    ELSIF v_migration_record.script_name LIKE '%create_table%' THEN
        v_rollback_sql := v_rollback_sql || '-- Para reverter a criação de uma tabela, execute algo como:\n';
        v_rollback_sql := v_rollback_sql || '-- DROP TABLE IF EXISTS nome_da_tabela CASCADE;' || E'\n';
    ELSIF v_migration_record.script_name LIKE '%add_foreign_key%' THEN
        v_rollback_sql := v_rollback_sql || '-- Para reverter a adição de uma chave estrangeira, execute algo como:\n';
        v_rollback_sql := v_rollback_sql || '-- ALTER TABLE nome_da_tabela DROP CONSTRAINT IF EXISTS nome_da_constraint;' || E'\n';
    ELSE
        v_rollback_sql := v_rollback_sql || '-- Não foi possível gerar automaticamente um script de rollback para este tipo de migração.\n';
        v_rollback_sql := v_rollback_sql || '-- Consulte a documentação da migração para obter instruções específicas de rollback.' || E'\n';
    END IF;
    
    -- Adicionar aviso sobre backup
    v_rollback_sql := v_rollback_sql || E'\n-- AVISO: Sempre faça backup do banco de dados antes de executar operações de rollback.\n';
    v_rollback_sql := v_rollback_sql || '-- O rollback de migrações pode resultar em perda de dados se não for executado corretamente.\n';
    
    RETURN v_rollback_sql;
EXCEPTION WHEN OTHERS THEN
    RETURN 'Erro ao gerar script de rollback: ' || SQLERRM;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Função para verificar a integridade das migrações
CREATE OR REPLACE FUNCTION verify_migrations_integrity()
RETURNS TABLE (
    check_name TEXT,
    status TEXT,
    details TEXT,
    severity TEXT
) AS $$
BEGIN
    -- Verificar se há migrações com falha
    RETURN QUERY
    SELECT 
        'failed_migrations' AS check_name,
        CASE 
            WHEN COUNT(*) = 0 THEN 'PASS'
            ELSE 'FAIL'
        END AS status,
        COUNT(*)::TEXT || ' migrações com falha encontradas' AS details,
        'HIGH' AS severity
    FROM schema_migrations
    WHERE success = FALSE;
    
    -- Verificar se há migrações duplicadas
    RETURN QUERY
    SELECT 
        'duplicate_versions' AS check_name,
        CASE 
            WHEN COUNT(*) = 0 THEN 'PASS'
            ELSE 'FAIL'
        END AS status,
        COUNT(*)::TEXT || ' versões de migração duplicadas encontradas' AS details,
        'HIGH' AS severity
    FROM (
        SELECT version
        FROM schema_migrations
        GROUP BY version
        HAVING COUNT(*) > 1
    ) AS duplicates;
    
    -- Verificar se há migrações com checksum inválido
    RETURN QUERY
    SELECT 
        'invalid_checksums' AS check_name,
        CASE 
            WHEN COUNT(*) = 0 THEN 'PASS'
            ELSE 'FAIL'
        END AS status,
        COUNT(*)::TEXT || ' migrações com checksum inválido' AS details,
        'MEDIUM' AS severity
    FROM (
        SELECT sm.version
        FROM schema_migrations sm
        WHERE sm.checksum IS NOT NULL
        AND NOT EXISTS (
            SELECT 1 
            FROM pg_proc 
            WHERE proname = 'apply_migration' 
            AND encode(digest(pg_get_functiondef(oid)::TEXT, 'sha256'), 'hex') = sm.checksum
        )
    ) AS invalid_checksums;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Registrar a criação deste script no log de auditoria
INSERT INTO audit_logs (
    id,
    organization_id,
    user_id,
    action,
    resource_type,
    resource_id,
    status,
    details
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111', -- ID da organização padrão
    '22222222-2222-2222-2222-222222222222', -- ID do usuário admin
    'execute',
    'schema',
    'iam',
    'success',
    '{"script": "05_iam_schema_migration_manager.sql", "version": "1.0", "description": "Criação do gerenciador de migrações do esquema IAM"}'
);

-- Mensagem de conclusão
DO $$
BEGIN
    RAISE NOTICE 'Gerenciador de migrações do esquema IAM criado com sucesso!';
    RAISE NOTICE 'Para verificar o status das migrações, execute: SELECT * FROM check_migration_status();';
    RAISE NOTICE 'Para aplicar uma migração, use: SELECT apply_migration(''2023.01.01.0001'', ''Descrição'', ''script_name.sql'', ''SQL_AQUI'');';
END
$$;
