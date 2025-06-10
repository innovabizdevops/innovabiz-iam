-- Script de Teste para Recomendações de Conformidade e Segurança - IAM Open X
-- Versão: 1.1
-- Data: 15/05/2025

-- 1. Teste de Criação de Recomendações

-- Teste 1: Usuário sem domínio
SELECT iam_access_control.provision_user(
    'teste_sem_dominio',
    'teste@exemplo.com',
    'Teste Sem Domínio',
    'Teste',
    ARRAY['role_acesso_basico']
);

-- 2. Teste de Segurança

-- Teste 2.1: Tentativa de força bruta com múltiplas tentativas
DO $$
DECLARE
    v_username TEXT := 'teste_sem_dominio';
    v_ip TEXT := '192.168.1.1';
    v_timestamp TIMESTAMP := current_timestamp;
BEGIN
    -- Simular 10 tentativas em 5 segundos
    FOR i IN 1..10 LOOP
        PERFORM iam_access_control.prevent_brute_force(
            v_username,
            v_ip,
            v_timestamp + (i * INTERVAL '0.5 seconds')
        );
    END LOOP;
END $$;

-- Teste 2.2: Acesso suspeito com múltiplos fatores
SELECT iam_access_control.detect_suspicious_access(
    'teste_sem_dominio',
    '1.2.3.4',
    'BR',
    '10.0.0.1',
    'Mozilla/5.0',
    current_timestamp - INTERVAL '2 hours'
);

-- Teste 2.3: Monitoramento de padrões de acesso
DO $$
DECLARE
    v_username TEXT := 'teste_sem_dominio';
    v_timestamp TIMESTAMP := current_timestamp;
BEGIN
    -- Simular padrão de acesso suspeito
    PERFORM iam_access_control.monitor_access_patterns(
        v_username,
        v_timestamp,
        'dashboard/admin'
    );
    
    PERFORM iam_access_control.monitor_access_patterns(
        v_username,
        v_timestamp + INTERVAL '1 second',
        'dashboard/admin'
    );
    
    PERFORM iam_access_control.monitor_access_patterns(
        v_username,
        v_timestamp + INTERVAL '2 seconds',
        'dashboard/admin'
    );
END $$;

-- Teste 2.4: Detecção de DDoS
DO $$
DECLARE
    v_timestamp TIMESTAMP := current_timestamp;
BEGIN
    -- Simular ataque DDoS
    FOR i IN 1..1000 LOOP
        PERFORM iam_access_control.detect_ddos(
            '1.2.3.' || (i % 10),
            'Mozilla/5.0',
            'dashboard/admin',
            v_timestamp + (i * INTERVAL '0.01 seconds')
        );
    END LOOP;
END $$;

-- Teste 2.5: Proteção contra SQL Injection
SELECT iam_access_control.prevent_sql_injection(
    'SELECT * FROM users WHERE id = 1 OR 1=1; --',
    'SELECT',
    'id'
);

-- Teste 2.6: Verificação de score de risco
SELECT * FROM iam_access_control.get_risk_score(
    'teste_sem_dominio',
    '1.2.3.4',
    'Mozilla/5.0',
    current_timestamp
);

-- 3. Teste de Recomendações de Conformidade

-- Teste 3.1: Geração de recomendações
SELECT iam_access_control.create_compliance_recommendations();

-- Teste 3.2: Resolução de recomendação
SELECT iam_access_control.resolve_recommendation(
    (SELECT id FROM iam_access_control.compliance_recommendations 
     WHERE username = 'teste_sem_dominio' 
     AND recommendation_type = 'DOMAIN_ASSIGNMENT' 
     LIMIT 1),
    'admin',
    'Domínio atribuído com sucesso'
);

-- 4. Teste de Relatórios de Segurança

-- Teste 4.1: Gerar relatório de segurança
SELECT * FROM iam_access_control.gerar_relatorio_seguranca(
    current_timestamp - INTERVAL '24 horas',
    current_timestamp
);

-- 5. Teste de Limpeza

-- Limpar usuário de teste
SELECT iam_access_control.cleanup_inactive_users('1 dia');

-- Limpar recomendações resolvidas
DELETE FROM iam_access_control.compliance_recommendations
WHERE status = 'RESOLVED';

-- Limpar histórico de ações
DELETE FROM iam_access_control.compliance_actions_history
WHERE action_type = 'RECOMMENDATION_RESOLVED';

-- Limpar logs de segurança
DELETE FROM iam_access_control.security_events
WHERE event_type IN ('SUSPICIOUS_ACCESS', 'ACCOUNT_LOCKED');

-- Limpar logs de acesso
DELETE FROM iam_access_control.access_logs
WHERE username = 'teste_sem_dominio';
