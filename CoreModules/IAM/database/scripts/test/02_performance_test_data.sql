-- INNOVABIZ - IAM Performance Test Data
-- Author: Eduardo Jeremias
-- Date: 19/05/2025
-- Version: 1.0
-- Description: Script para gerar dados de teste de desempenho para o módulo IAM

-- Configuração do ambiente
SET search_path TO iam, public;
SET statement_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

-- Configurações (ajustáveis conforme necessário)
DO $$
DECLARE
    -- Número de organizações a serem criadas
    org_count INTEGER := 5;
    
    -- Número de usuários por organização
    users_per_org INTEGER := 1000;
    
    -- Número de funções por organização
    roles_per_org INTEGER := 20;
    
    -- Número de permissões por função
    permissions_per_role INTEGER := 10;
    
    -- Número de funções por usuário
    roles_per_user INTEGER := 3;
    
    -- Número de sessões por usuário
    sessions_per_user INTEGER := 2;
    
    -- Número de entradas de log de auditoria por organização
    audit_logs_per_org INTEGER := 5000;
    
    -- Contadores
    org_id UUID;
    user_id UUID;
    role_id UUID;
    permission_id UUID;
    session_id UUID;
    audit_id UUID;
    
    -- Variáveis temporárias
    i INTEGER;
    j INTEGER;
    k INTEGER;
    l INTEGER;
    m INTEGER;
    n INTEGER;
    
    -- Função para gerar strings aleatórias
    FUNCTION random_string(length INTEGER) RETURNS TEXT AS $$
    DECLARE
        chars TEXT := 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
        result TEXT := '';
        i INTEGER := 0;
    BEGIN
        FOR i IN 1..length LOOP
            result := result || substr(chars, floor(random() * length(chars) + 1)::INTEGER, 1);
        END LOOP;
        RETURN result;
    END;
    $$ LANGUAGE plpgsql;
    
    -- Função para gerar e-mails aleatórios
    FUNCTION random_email() RETURNS TEXT AS $$
    BEGIN
        RETURN random_string(8) || '@' || random_string(6) || '.' || 
               (ARRAY['com', 'net', 'org', 'io', 'dev'])[floor(random() * 5 + 1)::INTEGER];
    END;
    $$ LANGUAGE plpgsql;
    
    -- Função para gerar IPs aleatórios
    FUNCTION random_ip() RETURNS TEXT AS $$
    BEGIN
        RETURN floor(random() * 255)::TEXT || '.' ||
               floor(random() * 255)::TEXT || '.' ||
               floor(random() * 255)::TEXT || '.' ||
               floor(random() * 255)::TEXT;
    END;
    $$ LANGUAGE plpgsql;
    
    -- Função para gerar user agents aleatórios
    FUNCTION random_user_agent() RETURNS TEXT AS $$
    DECLARE
        browsers TEXT[] := ARRAY[
            'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36',
            'Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0',
            'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.3 Safari/605.1.15',
            'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59',
            'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36',
            'Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1',
            'Mozilla/5.0 (iPad; CPU OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1',
            'Mozilla/5.0 (Linux; Android 10; SM-G981B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.162 Mobile Safari/537.36',
            'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 OPR/77.0.4054.277'
        ];
    BEGIN
        RETURN browsers[floor(random() * array_length(browsers, 1) + 1)::INTEGER];
    END;
    $$ LANGUAGE plpgsql;
    
    -- Função para gerar timestamps aleatórios dentro de um intervalo
    FUNCTION random_timestamp(start_timestamp TIMESTAMP, end_timestamp TIMESTAMP) 
    RETURNS TIMESTAMP AS $$
    BEGIN
        RETURN start_timestamp + random() * (end_timestamp - start_timestamp);
    END;
    $$ LANGUAGE plpgsql;
    
BEGIN
    RAISE NOTICE 'Iniciando geração de dados de teste de desempenho...';
    
    -- Desativar triggers de auditoria temporariamente
    RAISE NOTICE 'Desativando triggers de auditoria...';
    PERFORM disable_audit_triggers();
    
    -- Iniciar transação para melhor desempenho
    RAISE NOTICE 'Iniciando transação...';
    BEGIN
        -- Inserir organizações
        RAISE NOTICE 'Inserindo % organizações...', org_count;
        FOR i IN 1..org_count LOOP
            org_id := gen_random_uuid();
            
            INSERT INTO organizations (
                id, name, code, industry, sector, country_code, region_code, 
                is_active, settings, compliance_settings, metadata
            ) VALUES (
                org_id,
                'Organization ' || i,
                'ORG' || lpad(i::TEXT, 3, '0'),
                (ARRAY['Technology', 'Finance', 'Healthcare', 'Education', 'Retail', 'Manufacturing'])[floor(random() * 6 + 1)::INTEGER],
                (ARRAY['SaaS', 'FinTech', 'HealthTech', 'EdTech', 'E-commerce', 'Industrial'])[floor(random() * 6 + 1)::INTEGER],
                (ARRAY['PT', 'BR', 'US', 'GB', 'ES', 'FR'])[floor(random() * 6 + 1)::INTEGER],
                (ARRAY['LIS', 'SP', 'NY', 'LDN', 'MAD', 'PAR'])[floor(random() * 6 + 1)::INTEGER],
                true,
                jsonb_build_object(
                    'theme', (ARRAY['light', 'dark', 'system'])[floor(random() * 3 + 1)::INTEGER],
                    'locale', (ARRAY['en-US', 'pt-PT', 'pt-BR', 'es-ES', 'fr-FR'])[floor(random() * 5 + 1)::INTEGER],
                    'timezone', (ARRAY['UTC', 'Europe/Lisbon', 'America/Sao_Paulo', 'America/New_York', 'Europe/London'])[floor(random() * 5 + 1)::INTEGER]
                ),
                jsonb_build_object(
                    'gdpr_compliant', (ARRAY[true, false])[floor(random() * 2 + 1)::INTEGER],
                    'data_retention_days', (ARRAY[30, 60, 90, 365])[floor(random() * 4 + 1)::INTEGER],
                    'audit_log_retention_days', (ARRAY[90, 180, 365, 730])[floor(random() * 4 + 1)::INTEGER]
                ),
                jsonb_build_object(
                    'created_by', 'system',
                    'purpose', 'Performance testing',
                    'test_data', true
                )
            ) RETURNING id INTO org_id;
            
            -- Inserir funções para esta organização
            RAISE NOTICE '  Inserindo % funções para a organização %...', roles_per_org, i;
            FOR j IN 1..roles_per_org LOOP
                role_id := gen_random_uuid();
                
                -- Gerar permissões aleatórias
                DECLARE
                    permissions_json JSONB := '[]'::JSONB;
                    k INTEGER;
                    resource_types TEXT[] := ARRAY['user', 'role', 'permission', 'organization', 'audit_log', 'session'];
                    actions TEXT[] := ARRAY['create', 'read', 'update', 'delete', 'list', 'manage'];
                    resources TEXT[] := ARRAY['users', 'roles', 'permissions', 'organizations', 'audit_logs', 'sessions'];
                BEGIN
                    FOR k IN 1..permissions_per_role LOOP
                        permissions_json := permissions_json || jsonb_build_object(
                            'id', gen_random_uuid(),
                            'name', 'Permission ' || k || ' for Role ' || j,
                            'code', resources[(j + k) % array_length(resources, 1) + 1] || ':' || 
                                    actions[(j * k) % array_length(actions, 1) + 1],
                            'description', 'Auto-generated permission for testing',
                            'resource', resources[(j + k) % array_length(resources, 1) + 1],
                            'action', actions[(j * k) % array_length(actions, 1) + 1],
                            'is_active', true
                        );
                    END LOOP;
                    
                    INSERT INTO roles (
                        id, organization_id, name, description, is_system_role, permissions
                    ) VALUES (
                        role_id,
                        org_id,
                        'Role ' || j,
                        'Auto-generated role for performance testing',
                        j = 1, -- Primeira função é uma função de sistema
                        permissions_json
                    );
                END;
            END LOOP;
            
            -- Inserir usuários para esta organização
            RAISE NOTICE '  Inserindo % usuários para a organização %...', users_per_org, i;
            FOR j IN 1..users_per_org LOOP
                user_id := gen_random_uuid();
                
                INSERT INTO users (
                    id, organization_id, username, email, full_name, 
                    password_hash, status, last_login, preferences, metadata
                ) VALUES (
                    user_id,
                    org_id,
                    'user' || j || '_org' || i,
                    'user' || j || '_org' || i || '@example.com',
                    'User ' || j || ' from Org ' || i,
                    -- Senha: Password123! (hash bcrypt)
                    '$2a$10$N9qo8uLOickgx2ZMRZoMy.MQDqShCs6UdHFNC4VC8Uj1GZASjs3l6',
                    (ARRAY['active', 'inactive', 'locked'])[CASE WHEN j % 100 = 0 THEN 3 WHEN j % 10 = 0 THEN 2 ELSE 1 END],
                    CASE WHEN j % 10 != 0 THEN NOW() - (random() * 30)::INTEGER * INTERVAL '1 day' ELSE NULL END,
                    jsonb_build_object(
                        'theme', (ARRAY['light', 'dark', 'system'])[floor(random() * 3 + 1)::INTEGER],
                        'notifications', jsonb_build_object(
                            'email', j % 2 = 0,
                            'push', j % 3 = 0,
                            'in_app', true
                        )
                    ),
                    jsonb_build_object(
                        'department', (ARRAY['Engineering', 'Marketing', 'Sales', 'Support', 'HR', 'Finance'])[(j % 6) + 1],
                        'job_title', 'Employee ' || j,
                        'phone', '+351 9' || lpad(floor(random() * 90000000 + 10000000)::TEXT, 8, '0'),
                        'created_by', 'system',
                        'test_data', true
                    )
                );
                
                -- Atribuir funções a este usuário
                FOR k IN 1..roles_per_user LOOP
                    INSERT INTO user_roles (
                        id, user_id, role_id, granted_by, granted_at, 
                        expires_at, is_active, metadata
                    ) VALUES (
                        gen_random_uuid(),
                        user_id,
                        (SELECT id FROM roles WHERE organization_id = org_id ORDER BY random() LIMIT 1),
                        '00000000-0000-0000-0000-000000000000', -- ID do sistema
                        NOW() - (random() * 365)::INTEGER * INTERVAL '1 day',
                        CASE 
                            WHEN random() > 0.9 THEN NOW() + (random() * 365)::INTEGER * INTERVAL '1 day'
                            ELSE NULL 
                        END,
                        random() > 0.1, -- 90% de chance de estar ativo
                        jsonb_build_object('assigned_by', 'system', 'purpose', 'performance_testing')
                    );
                END LOOP;
                
                -- Criar sessões para este usuário
                FOR k IN 1..sessions_per_user LOOP
                    IF j % 10 != 0 THEN -- 90% dos usuários têm sessões ativas
                        session_id := gen_random_uuid();
                        
                        INSERT INTO sessions (
                            id, user_id, token, ip_address, user_agent,
                            created_at, expires_at, last_activity, metadata, is_active
                        ) VALUES (
                            session_id,
                            user_id,
                            'test_token_' || session_id,
                            random_ip(),
                            random_user_agent(),
                            NOW() - (random() * 7)::INTEGER * INTERVAL '1 day',
                            NOW() + (random() * 7)::INTEGER * INTERVAL '1 day',
                            NOW() - (random() * 24)::INTEGER * INTERVAL '1 hour',
                            jsonb_build_object(
                                'device_id', 'device_' || (j * k) % 1000,
                                'os', (ARRAY['Windows', 'macOS', 'Linux', 'iOS', 'Android'])[floor(random() * 5 + 1)::INTEGER],
                                'browser', (ARRAY['Chrome', 'Firefox', 'Safari', 'Edge', 'Opera'])[floor(random() * 5 + 1)::INTEGER],
                                'location', (ARRAY['Lisbon', 'Porto', 'São Paulo', 'New York', 'London', 'Madrid'])[floor(random() * 6 + 1)::INTEGER]
                            ),
                            true
                        );
                    END IF;
                END LOOP;
                
                -- Inserir logs de auditoria para este usuário
                FOR k IN 1..(audit_logs_per_org / users_per_org) LOOP
                    audit_id := gen_random_uuid();
                    
                    INSERT INTO audit_logs (
                        id, organization_id, user_id, action, resource_type,
                        resource_id, timestamp, ip_address, status, details, session_id
                    ) VALUES (
                        audit_id,
                        org_id,
                        user_id,
                        (ARRAY['create', 'read', 'update', 'delete'])[floor(random() * 4 + 1)::INTEGER],
                        (ARRAY['user', 'role', 'permission', 'organization', 'session'])[floor(random() * 5 + 1)::INTEGER],
                        gen_random_uuid(),
                        NOW() - (random() * 30)::INTEGER * INTERVAL '1 day',
                        random_ip(),
                        (ARRAY['success', 'failure', 'error'])[floor(random() * 3 + 1)::INTEGER],
                        jsonb_build_object(
                            'old_values', '{}',
                            'new_values', jsonb_build_object('field' || floor(random() * 10 + 1)::TEXT, 'value' || floor(random() * 100 + 1)::TEXT),
                            'changed_fields', ARRAY['field' || floor(random() * 10 + 1)::TEXT],
                            'user_agent', random_user_agent(),
                            'source', 'performance_test'
                        ),
                        CASE WHEN random() > 0.3 THEN (SELECT id FROM sessions WHERE user_id = user_id ORDER BY random() LIMIT 1) ELSE NULL END
                    );
                END LOOP;
                
                -- Commit a cada 100 usuários para evitar transações muito longas
                IF j % 100 = 0 THEN
                    COMMIT;
                    BEGIN
                        RAISE NOTICE '  Processados % usuários...', j;
                    EXCEPTION WHEN OTHERS THEN
                        RAISE WARNING 'Erro ao registrar progresso: %', SQLERRM;
                    END;
                END IF;
            END LOOP;
            
            RAISE NOTICE 'Organização % concluída com sucesso!', i;
        END LOOP;
        
        -- Commit final
        COMMIT;
        RAISE NOTICE 'Dados de teste gerados com sucesso!';
        
    EXCEPTION WHEN OTHERS THEN
        -- Em caso de erro, fazer rollback e relatar
        ROLLBACK;
        RAISE EXCEPTION 'Erro ao gerar dados de teste: %', SQLERRM;
    END;
    
    -- Reativar triggers de auditoria
    RAISE NOTICE 'Reativando triggers de auditoria...';
    PERFORM enable_audit_triggers();
    
    -- Atualizar estatísticas
    RAISE NOTICE 'Atualizando estatísticas do banco de dados...';
    ANALYZE;
    
    -- Mensagem final
    RAISE NOTICE 'Processo de geração de dados de teste concluído com sucesso!';
    
END $$;
