-- INNOVABIZ - IAM Compliance Report
-- Author: Eduardo Jeremias
-- Date: 19/05/2025
-- Version: 1.0
-- Description: Script para gerar relatório de conformidade do IAM

-- Configuração do ambiente
SET search_path TO iam, public;
SET statement_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

-- Criar função para gerar relatório em formato JSON
CREATE OR REPLACE FUNCTION generate_iam_compliance_report()
RETURNS JSONB AS $$
DECLARE
    report JSONB;
    check_result JSONB;
    security_checks JSONB[] := '{}';
    
    -- Configurações de conformidade (podem ser parametrizadas)
    min_password_length INT := 12;
    max_password_age_days INT := 90;
    inactive_user_days INT := 90;
    max_failed_attempts INT := 5;
    mfa_required_for_admins BOOLEAN := true;
    
    -- Contadores
    total_checks INT := 0;
    passed_checks INT := 0;
    failed_checks INT := 0;
    warning_checks INT := 0;
    
    -- Variáveis temporárias
    temp_count BIGINT;
    temp_record RECORD;
    temp_array TEXT[];
    temp_json JSONB;
    
    -- Função auxiliar para adicionar resultado da verificação
    PROCEDURE add_check(
        p_check_name TEXT,
        p_check_description TEXT,
        p_status TEXT,
        p_severity TEXT,
        p_findings TEXT[] DEFAULT NULL,
        p_recommendation TEXT DEFAULT NULL
    ) AS $$
    BEGIN
        total_checks := total_checks + 1;
        
        CASE p_status
            WHEN 'PASS' THEN passed_checks := passed_checks + 1;
            WHEN 'FAIL' THEN failed_checks := failed_checks + 1;
            WHEN 'WARNING' THEN warning_checks := warning_checks + 1;
        END CASE;
        
        security_checks := array_append(
            security_checks,
            jsonb_build_object(
                'check_name', p_check_name,
                'check_description', p_check_description,
                'status', p_status,
                'severity', p_severity,
                'findings', COALESCE(p_findings, '{}'),
                'recommendation', p_recommendation,
                'timestamp', NOW()
            )
        );
    END;
    
BEGIN
    -- Cabeçalho do relatório
    report := jsonb_build_object(
        'report_id', gen_random_uuid(),
        'report_type', 'iam_compliance',
        'generated_at', NOW(),
        'scope', 'iam_schema',
        'standards', jsonb_build_array('ISO27001', 'GDPR', 'LGPD', 'NIST', 'CIS')
    );
    
    -- 1. Verificação de Políticas de Senha
    -- 1.1. Verificar se há usuários sem senha
    EXECUTE 'SELECT COUNT(*) FROM users WHERE password_hash IS NULL OR password_hash = '''' OR password_hash = ''inactive''' INTO temp_count;
    
    IF temp_count > 0 THEN
        EXECUTE 'SELECT array_agg(username) FROM users WHERE password_hash IS NULL OR password_hash = '''' OR password_hash = ''inactive''' INTO temp_array;
        add_check(
            'users_without_password',
            'Verificar usuários sem senha definida',
            'FAIL',
            'HIGH',
            temp_array,
            'Todos os usuários devem ter uma senha forte definida. Considere definir uma senha ou desativar contas não utilizadas.'
        );
    ELSE
        add_check(
            'users_without_password',
            'Verificar usuários sem senha definida',
            'PASS',
            'HIGH'
        );
    END IF;
    
    -- 1.2. Verificar complexidade de senha
    -- Esta é uma verificação conceitual, pois não podemos descriptografar hashes
    add_check(
        'password_complexity',
        'Verificar se as senhas atendem aos requisitos de complexidade',
        'INFO',
        'MEDIUM',
        NULL,
        'Implemente verificações regulares de força de senha e aplique políticas de expiração.'
    );
    
    -- 2. Verificação de Contas de Usuário
    -- 2.1. Verificar contas inativas
    EXECUTE 'SELECT COUNT(*) FROM users WHERE status = ''inactive''' INTO temp_count;
    
    IF temp_count > 0 THEN
        EXECUTE 'SELECT array_agg(username) FROM users WHERE status = ''inactive''' INTO temp_array;
        add_check(
            'inactive_users',
            'Identificar contas de usuário inativas',
            'WARNING',
            'LOW',
            temp_array,
            'Considere arquivar ou excluir contas de usuário inativas que não são mais necessárias.'
        );
    ELSE
        add_check(
            'inactive_users',
            'Identificar contas de usuário inativas',
            'PASS',
            'LOW'
        );
    END IF;
    
    -- 2.2. Verificar contas de administrador
    EXECUTE 'SELECT COUNT(*) FROM user_roles ur JOIN roles r ON ur.role_id = r.id WHERE r.name = ''Administrator''' INTO temp_count;
    
    IF temp_count = 0 THEN
        add_check(
            'admin_accounts',
            'Verificar contas de administrador',
            'FAIL',
            'HIGH',
            NULL,
            'É necessário ter pelo menos uma conta de administrador ativa no sistema.'
        );
    ELSIF temp_count > 3 THEN
        add_check(
            'admin_accounts',
            'Verificar contas de administrador',
            'WARNING',
            'MEDIUM',
            ARRAY['Múltiplas contas de administrador (' || temp_count::text || ')'],
            'Considere reduzir o número de contas de administrador para o mínimo necessário.'
        );
    ELSE
        add_check(
            'admin_accounts',
            'Verificar contas de administrador',
            'PASS',
            'HIGH'
        );
    END IF;
    
    -- 3. Verificação de Funções e Permissões
    -- 3.1. Verificar funções sem permissões atribuídas
    EXECUTE 'SELECT COUNT(*) FROM roles WHERE jsonb_array_length(permissions) = 0' INTO temp_count;
    
    IF temp_count > 0 THEN
        EXECUTE 'SELECT array_agg(name) FROM roles WHERE jsonb_array_length(permissions) = 0' INTO temp_array;
        add_check(
            'roles_without_permissions',
            'Identificar funções sem permissões atribuídas',
            'WARNING',
            'MEDIUM',
            temp_array,
            'Todas as funções devem ter permissões apropriadas atribuídas. Considere atribuir permissões ou remover funções não utilizadas.'
        );
    ELSE
        add_check(
            'roles_without_permissions',
            'Identificar funções sem permissões atribuídas',
            'PASS',
            'MEDIUM'
        );
    END IF;
    
    -- 3.2. Verificar permissões excessivas
    -- Esta é uma verificação conceitual baseada em padrões conhecidos
    add_check(
        'excessive_permissions',
        'Verificar permissões excessivas em funções',
        'INFO',
        'MEDIUM',
        NULL,
        'Revise regularmente as permissões atribuídas a cada função para garantir o princípio do menor privilégio.'
    );
    
    -- 4. Verificação de Auditoria
    -- 4.1. Verificar se o log de auditoria está ativado
    EXECUTE 'SELECT COUNT(*) FROM pg_trigger WHERE tgname = ''trg_audit_log''' INTO temp_count;
    
    IF temp_count = 0 THEN
        add_check(
            'audit_logging_enabled',
            'Verificar se o log de auditoria está ativado',
            'FAIL',
            'HIGH',
            NULL,
            'Ative o log de auditoria para rastrear alterações críticas no sistema.'
        );
    ELSE
        add_check(
            'audit_logging_enabled',
            'Verificar se o log de auditoria está ativado',
            'PASS',
            'HIGH'
        );
    END IF;
    
    -- 4.2. Verificar retenção de logs
    EXECUTE 'SELECT COUNT(*) FROM audit_logs WHERE timestamp < (NOW() - INTERVAL ''365 days'');' INTO temp_count;
    
    IF temp_count > 0 THEN
        add_check(
            'audit_log_retention',
            'Verificar retenção de logs de auditoria',
            'WARNING',
            'MEDIUM',
            ARRAY[temp_count::text || ' registros de auditoria com mais de 1 ano'],
            'Considere arquivar logs antigos para otimizar o desempenho do banco de dados.'
        );
    ELSE
        add_check(
            'audit_log_retention',
            'Verificar retenção de logs de auditoria',
            'PASS',
            'MEDIUM'
        );
    END IF;
    
    -- 5. Verificação de Sessões
    -- 5.1. Verificar sessões expiradas
    EXECUTE 'SELECT COUNT(*) FROM sessions WHERE expires_at < NOW() AND is_active = true' INTO temp_count;
    
    IF temp_count > 0 THEN
        EXECUTE 'SELECT array_agg(id::text) FROM sessions WHERE expires_at < NOW() AND is_active = true' INTO temp_array;
        add_check(
            'expired_sessions',
            'Identificar sessões expiradas ainda ativas',
            'FAIL',
            'MEDIUM',
            temp_array,
            'Sessões expiradas devem ser encerradas automaticamente. Considere implementar um processo de limpeza de sessões.'
        );
    ELSE
        add_check(
            'expired_sessions',
            'Identificar sessões expiradas ainda ativas',
            'PASS',
            'MEDIUM'
        );
    END IF;
    
    -- 5.2. Verificar sessões de longa duração
    EXECUTE 'SELECT COUNT(*) FROM sessions WHERE created_at < (NOW() - INTERVAL ''30 days'') AND is_active = true' INTO temp_count;
    
    IF temp_count > 0 THEN
        add_check(
            'long_running_sessions',
            'Identificar sessões de longa duração',
            'WARNING',
            'LOW',
            ARRAY[temp_count::text || ' sessões ativas com mais de 30 dias'],
            'Considere implementar um tempo máximo de sessão para sessões muito longas.'
        );
    ELSE
        add_check(
            'long_running_sessions',
            'Identificar sessões de longa duração',
            'PASS',
            'LOW'
        );
    END IF;
    
    -- 6. Verificação de Configurações de Segurança
    -- 6.1. Verificar se o SSL está habilitado
    EXECUTE 'SELECT setting FROM pg_settings WHERE name = ''ssl''' INTO temp_record;
    
    IF temp_record.setting = 'off' THEN
        add_check(
            'ssl_enabled',
            'Verificar se o SSL está habilitado',
            'FAIL',
            'HIGH',
            NULL,
            'Habilite o SSL para criptografar as conexões com o banco de dados.'
        );
    ELSE
        add_check(
            'ssl_enabled',
            'Verificar se o SSL está habilitado',
            'PASS',
            'HIGH'
        );
    END IF;
    
    -- 6.2. Verificar se a criptografia de dados em repouso está habilitada
    -- Esta é uma verificação conceitual, pois requer acesso ao sistema de arquivos
    add_check(
        'data_at_rest_encryption',
        'Verificar se a criptografia de dados em repouso está habilitada',
        'INFO',
        'HIGH',
        NULL,
        'Considere habilitar a criptografia de dados em repouso para proteger dados sensíveis.'
    );
    
    -- 7. Verificação de Atualizações e Patches
    -- 7.1. Verificar versão do PostgreSQL
    EXECUTE 'SELECT version()' INTO temp_record;
    
    add_check(
        'postgresql_version',
        'Verificar versão do PostgreSQL',
        'INFO',
        'MEDIUM',
        ARRAY[temp_record.version],
        'Mantenha o PostgreSQL atualizado com as últimas correções de segurança.'
    );
    
    -- 8. Verificação de Backup e Recuperação
    -- 8.1. Verificar se há um procedimento de backup
    add_check(
        'backup_procedure',
        'Verificar procedimentos de backup e recuperação',
        'INFO',
        'HIGH',
        NULL,
        'Certifique-se de que existam procedimentos regulares de backup e que a recuperação tenha sido testada recentemente.'
    );
    
    -- Resumo do relatório
    report := report || jsonb_build_object(
        'summary', jsonb_build_object(
            'total_checks', total_checks,
            'passed_checks', passed_checks,
            'failed_checks', failed_checks,
            'warning_checks', warning_checks,
            'compliance_score', ROUND((passed_checks::float / NULLIF(total_checks, 0)) * 100, 2),
            'generated_at', NOW()
        ),
        'checks', security_checks
    );
    
    -- Retornar o relatório completo
    RETURN report;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para gerar relatório em formato de tabela (mais legível)
CREATE OR REPLACE FUNCTION generate_iam_compliance_report_table()
RETURNS TABLE (
    check_name TEXT,
    check_description TEXT,
    status TEXT,
    severity TEXT,
    findings TEXT,
    recommendation TEXT
) AS $$
DECLARE
    report_json JSONB;
    check_item JSONB;
BEGIN
    -- Gerar o relatório em JSON
    SELECT * FROM generate_iam_compliance_report() INTO report_json;
    
    -- Retornar cada verificação como uma linha da tabela
    FOR check_item IN SELECT * FROM jsonb_array_elements(report_json->'checks')
    LOOP
        check_name := check_item->>'check_name';
        check_description := check_item->>'check_description';
        status := check_item->>'status';
        severity := check_item->>'severity';
        
        -- Converter array de findings para string
        IF jsonb_array_length(check_item->'findings') > 0 THEN
            findings := array_to_string(ARRAY(SELECT jsonb_array_elements_text(check_item->'findings')), '; ');
        ELSE
            findings := NULL;
        END IF;
        
        recommendation := check_item->>'recommendation';
        
        RETURN NEXT;
    END LOOP;
    
    -- Adicionar linha de resumo
    check_name := 'SUMMARY';
    check_description := 'Resumo da verificação de conformidade';
    status := CASE 
        WHEN (report_json->'summary'->>'failed_checks')::int > 0 THEN 'FAIL' 
        WHEN (report_json->'summary'->>'warning_checks')::int > 0 THEN 'WARNING' 
        ELSE 'PASS' 
    END;
    severity := 'INFO';
    findings := 'Total: ' || (report_json->'summary'->>'total_checks') || 
               ', Aprovados: ' || (report_json->'summary'->>'passed_checks') ||
               ', Falhas: ' || (report_json->'summary'->>'failed_checks') ||
               ', Alertas: ' || (report_json->'summary'->>'warning_checks');
    recommendation := 'Pontuação de conformidade: ' || (report_json->'summary'->>'compliance_score') || '%';
    
    RETURN NEXT;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Exemplo de como usar as funções:
-- SELECT * FROM generate_iam_compliance_report(); -- Retorna JSON
-- SELECT * FROM generate_iam_compliance_report_table(); -- Retorna tabela formatada

-- Criar uma visualização para o relatório de conformidade
CREATE OR REPLACE VIEW vw_iam_compliance_report AS
SELECT 
    check_name,
    check_description,
    status,
    severity,
    findings,
    recommendation
FROM generate_iam_compliance_report_table();

-- Comentário da visão
COMMENT ON VIEW vw_iam_compliance_report IS 'Visão que exibe o relatório de conformidade do IAM em formato tabular';

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
    '{"script": "03_generate_iam_compliance_report.sql", "version": "1.0", "description": "Criação de funções para geração de relatório de conformidade do IAM"}'
);

-- Mensagem de conclusão
DO $$
BEGIN
    RAISE NOTICE 'Script de geração de relatório de conformidade do IAM criado com sucesso!';
    RAISE NOTICE 'Para visualizar o relatório, execute: SELECT * FROM vw_iam_compliance_report;';
END
$$;
