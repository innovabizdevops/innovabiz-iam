-- Sistema de Notificações para Testes de Autenticação

-- 1. Tabela de Tipos de Notificações
CREATE TABLE IF NOT EXISTS notification_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('INFO', 'WARNING', 'ERROR', 'CRITICAL')),
    category VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Tabela de Notificações
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_type_id INTEGER REFERENCES notification_types(id),
    test_id INTEGER NOT NULL,
    test_case_id INTEGER NOT NULL,
    message TEXT NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('PENDING', 'SENT', 'FAILED', 'DELIVERED')),
    delivery_method VARCHAR(50) NOT NULL,
    delivery_target TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    sent_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT
);

-- 3. Tabela de Métricas de Notificações
CREATE TABLE IF NOT EXISTS notification_metrics (
    id SERIAL PRIMARY KEY,
    notification_id UUID REFERENCES notifications(id),
    delivery_time INTERVAL,
    delivery_attempts INTEGER DEFAULT 1,
    response_time INTERVAL,
    success_rate DECIMAL(5,2),
    retry_count INTEGER DEFAULT 0,
    last_retry_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 4. Função para Registrar Notificação
CREATE OR REPLACE FUNCTION notification.register_notification(
    p_test_id INTEGER,
    p_test_case_id INTEGER,
    p_message TEXT,
    p_severity VARCHAR(20),
    p_delivery_method VARCHAR(50),
    p_delivery_target TEXT,
    p_category VARCHAR(50)
)
RETURNS UUID AS $$
DECLARE
    v_notification_type_id INTEGER;
    v_notification_id UUID;
BEGIN
    -- Obter ou criar tipo de notificação
    SELECT id INTO v_notification_type_id 
    FROM notification_types 
    WHERE name = p_category AND severity = p_severity;
    
    IF NOT FOUND THEN
        INSERT INTO notification_types (name, description, severity, category)
        VALUES (p_category, 'Notification for ' || p_category, p_severity, p_category)
        RETURNING id INTO v_notification_type_id;
    END IF;
    
    -- Registrar notificação
    INSERT INTO notifications (
        notification_type_id,
        test_id,
        test_case_id,
        message,
        status,
        delivery_method,
        delivery_target
    ) VALUES (
        v_notification_type_id,
        p_test_id,
        p_test_case_id,
        p_message,
        'PENDING',
        p_delivery_method,
        p_delivery_target
    ) RETURNING id INTO v_notification_id;
    
    RETURN v_notification_id;
END;
$$ LANGUAGE plpgsql;

-- 5. Função para Enviar Notificação
CREATE OR REPLACE FUNCTION notification.send_notification(
    p_notification_id UUID
)
RETURNS BOOLEAN AS $$
DECLARE
    v_notification RECORD;
    v_delivery_time INTERVAL;
    v_response_time INTERVAL;
    v_success BOOLEAN := false;
BEGIN
    -- Obter informações da notificação
    SELECT * INTO v_notification 
    FROM notifications 
    WHERE id = p_notification_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Notification not found';
    END IF;
    
    -- Simular envio de notificação
    PERFORM pg_sleep(1); -- Simula tempo de envio
    
    -- Atualizar status e métricas
    UPDATE notifications 
    SET 
        status = CASE 
            WHEN random() > 0.1 THEN 'DELIVERED' -- 90% de sucesso
            ELSE 'FAILED'
        END,
        sent_at = CURRENT_TIMESTAMP
    WHERE id = p_notification_id;
    
    -- Registrar métricas
    INSERT INTO notification_metrics (
        notification_id,
        delivery_time,
        delivery_attempts,
        response_time,
        success_rate
    ) VALUES (
        p_notification_id,
        CURRENT_TIMESTAMP - v_notification.created_at,
        1,
        INTERVAL '1 second',
        CASE 
            WHEN random() > 0.1 THEN 100.0
            ELSE 0.0
        END
    );
    
    RETURN true;
END;
$$ LANGUAGE plpgsql;

-- 6. Função para Gerar Relatório de Notificações
CREATE OR REPLACE FUNCTION notification.generate_notification_report(
    p_test_id INTEGER,
    p_start_date TIMESTAMP WITH TIME ZONE,
    p_end_date TIMESTAMP WITH TIME ZONE
)
RETURNS TABLE (
    notification_id UUID,
    test_case_id INTEGER,
    message TEXT,
    status VARCHAR(20),
    delivery_method VARCHAR(50),
    delivery_time INTERVAL,
    success_rate DECIMAL(5,2)
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        n.id as notification_id,
        n.test_case_id,
        n.message,
        n.status,
        n.delivery_method,
        nm.delivery_time,
        nm.success_rate
    FROM notifications n
    JOIN notification_metrics nm ON n.id = nm.notification_id
    WHERE n.test_id = p_test_id
    AND n.created_at BETWEEN p_start_date AND p_end_date
    ORDER BY n.created_at DESC;
END;
$$ LANGUAGE plpgsql;

-- 7. Função para Notificar Falhas Críticas
CREATE OR REPLACE FUNCTION notification.notify_critical_failure(
    p_test_id INTEGER,
    p_test_case_id INTEGER,
    p_message TEXT
)
RETURNS UUID AS $$
DECLARE
    v_notification_id UUID;
BEGIN
    -- Registrar notificação crítica
    v_notification_id := notification.register_notification(
        p_test_id,
        p_test_case_id,
        p_message,
        'CRITICAL',
        'EMAIL',
        'security-team@example.com',
        'Critical Failure'
    );
    
    -- Enviar imediatamente
    PERFORM notification.send_notification(v_notification_id);
    
    RETURN v_notification_id;
END;
$$ LANGUAGE plpgsql;

-- 8. Função para Notificar Sucesso de Teste
CREATE OR REPLACE FUNCTION notification.notify_test_success(
    p_test_id INTEGER,
    p_test_case_id INTEGER
)
RETURNS UUID AS $$
DECLARE
    v_notification_id UUID;
BEGIN
    -- Registrar notificação de sucesso
    v_notification_id := notification.register_notification(
        p_test_id,
        p_test_case_id,
        'Test case executed successfully',
        'INFO',
        'SLACK',
        '#test-results',
        'Test Success'
    );
    
    -- Enviar notificação
    PERFORM notification.send_notification(v_notification_id);
    
    RETURN v_notification_id;
END;
$$ LANGUAGE plpgsql;

-- 9. Função para Notificar Falha de Teste
CREATE OR REPLACE FUNCTION notification.notify_test_failure(
    p_test_id INTEGER,
    p_test_case_id INTEGER,
    p_error_message TEXT
)
RETURNS UUID AS $$
DECLARE
    v_notification_id UUID;
BEGIN
    -- Registrar notificação de falha
    v_notification_id := notification.register_notification(
        p_test_id,
        p_test_case_id,
        'Test case failed: ' || p_error_message,
        'ERROR',
        'EMAIL',
        'dev-team@example.com',
        'Test Failure'
    );
    
    -- Enviar notificação
    PERFORM notification.send_notification(v_notification_id);
    
    RETURN v_notification_id;
END;
$$ LANGUAGE plpgsql;

-- 10. Função para Notificar Conformidade
CREATE OR REPLACE FUNCTION notification.notify_compliance_status(
    p_test_id INTEGER,
    p_test_case_id INTEGER,
    p_status VARCHAR(20)
)
RETURNS UUID AS $$
DECLARE
    v_notification_id UUID;
BEGIN
    -- Registrar notificação de status de conformidade
    v_notification_id := notification.register_notification(
        p_test_id,
        p_test_case_id,
        'Compliance status: ' || p_status,
        CASE 
            WHEN p_status = 'COMPLIANT' THEN 'INFO'
            ELSE 'WARNING'
        END,
        'EMAIL',
        'compliance-team@example.com',
        'Compliance Status'
    );
    
    -- Enviar notificação
    PERFORM notification.send_notification(v_notification_id);
    
    RETURN v_notification_id;
END;
$$ LANGUAGE plpgsql;

-- 11. Função para Notificar Métricas de Performance
CREATE OR REPLACE FUNCTION notification.notify_performance_metrics(
    p_test_id INTEGER,
    p_test_case_id INTEGER,
    p_metrics JSONB
)
RETURNS UUID AS $$
DECLARE
    v_notification_id UUID;
    v_message TEXT;
BEGIN
    -- Formatar mensagem com métricas
    v_message := 'Performance metrics for test case ' || p_test_case_id || ': ' || 
                 'Response time: ' || p_metrics->>'response_time' || 
                 ' | Success rate: ' || p_metrics->>'success_rate';
    
    -- Registrar notificação de métricas
    v_notification_id := notification.register_notification(
        p_test_id,
        p_test_case_id,
        v_message,
        'INFO',
        'SLACK',
        '#performance-metrics',
        'Performance Metrics'
    );
    
    -- Enviar notificação
    PERFORM notification.send_notification(v_notification_id);
    
    RETURN v_notification_id;
END;
$$ LANGUAGE plpgsql;

-- 12. Função para Notificar Integração
CREATE OR REPLACE FUNCTION notification.notify_integration_status(
    p_test_id INTEGER,
    p_test_case_id INTEGER,
    p_status VARCHAR(20)
)
RETURNS UUID AS $$
DECLARE
    v_notification_id UUID;
BEGIN
    -- Registrar notificação de status de integração
    v_notification_id := notification.register_notification(
        p_test_id,
        p_test_case_id,
        'Integration test status: ' || p_status,
        CASE 
            WHEN p_status = 'SUCCESS' THEN 'INFO'
            ELSE 'ERROR'
        END,
        'EMAIL',
        'integration-team@example.com',
        'Integration Status'
    );
    
    -- Enviar notificação
    PERFORM notification.send_notification(v_notification_id);
    
    RETURN v_notification_id;
END;
$$ LANGUAGE plpgsql;

-- 13. Função para Notificar Análise de Segurança
CREATE OR REPLACE FUNCTION notification.notify_security_analysis(
    p_test_id INTEGER,
    p_test_case_id INTEGER,
    p_risk_level VARCHAR(20)
)
RETURNS UUID AS $$
DECLARE
    v_notification_id UUID;
BEGIN
    -- Registrar notificação de análise de segurança
    v_notification_id := notification.register_notification(
        p_test_id,
        p_test_case_id,
        'Security risk level: ' || p_risk_level,
        CASE 
            WHEN p_risk_level IN ('LOW', 'MEDIUM') THEN 'WARNING'
            ELSE 'ERROR'
        END,
        'EMAIL',
        'security-team@example.com',
        'Security Analysis'
    );
    
    -- Enviar notificação
    PERFORM notification.send_notification(v_notification_id);
    
    RETURN v_notification_id;
END;
$$ LANGUAGE plpgsql;

-- 14. Função para Notificar Atualizações de Sistema
CREATE OR REPLACE FUNCTION notification.notify_system_update(
    p_test_id INTEGER,
    p_test_case_id INTEGER,
    p_update_type VARCHAR(50)
)
RETURNS UUID AS $$
DECLARE
    v_notification_id UUID;
BEGIN
    -- Registrar notificação de atualização
    v_notification_id := notification.register_notification(
        p_test_id,
        p_test_case_id,
        'System update: ' || p_update_type,
        'INFO',
        'SLACK',
        '#system-updates',
        'System Update'
    );
    
    -- Enviar notificação
    PERFORM notification.send_notification(v_notification_id);
    
    RETURN v_notification_id;
END;
$$ LANGUAGE plpgsql;

-- 15. Função para Notificar Alertas de Monitoramento
CREATE OR REPLACE FUNCTION notification.notify_monitoring_alert(
    p_test_id INTEGER,
    p_test_case_id INTEGER,
    p_alert_type VARCHAR(50)
)
RETURNS UUID AS $$
DECLARE
    v_notification_id UUID;
BEGIN
    -- Registrar notificação de alerta
    v_notification_id := notification.register_notification(
        p_test_id,
        p_test_case_id,
        'Monitoring alert: ' || p_alert_type,
        'WARNING',
        'EMAIL',
        'ops-team@example.com',
        'Monitoring Alert'
    );
    
    -- Enviar notificação
    PERFORM notification.send_notification(v_notification_id);
    
    RETURN v_notification_id;
END;
$$ LANGUAGE plpgsql;

-- 16. Função para Notificar Status de Implementação
CREATE OR REPLACE FUNCTION notification.notify_implementation_status(
    p_test_id INTEGER,
    p_test_case_id INTEGER,
    p_status VARCHAR(20)
)
RETURNS UUID AS $$
DECLARE
    v_notification_id UUID;
BEGIN
    -- Registrar notificação de status de implementação
    v_notification_id := notification.register_notification(
        p_test_id,
        p_test_case_id,
        'Implementation status: ' || p_status,
        CASE 
            WHEN p_status = 'COMPLETED' THEN 'INFO'
            ELSE 'WARNING'
        END,
        'SLACK',
        '#implementation-status',
        'Implementation Status'
    );
    
    -- Enviar notificação
    PERFORM notification.send_notification(v_notification_id);
    
    RETURN v_notification_id;
END;
$$ LANGUAGE plpgsql;

-- 17. Função para Notificar Status de Qualidade
CREATE OR REPLACE FUNCTION notification.notify_quality_status(
    p_test_id INTEGER,
    p_test_case_id INTEGER,
    p_quality_score DECIMAL(5,2)
)
RETURNS UUID AS $$
DECLARE
    v_notification_id UUID;
BEGIN
    -- Registrar notificação de status de qualidade
    v_notification_id := notification.register_notification(
        p_test_id,
        p_test_case_id,
        'Quality score: ' || p_quality_score,
        CASE 
            WHEN p_quality_score >= 90 THEN 'INFO'
            WHEN p_quality_score >= 70 THEN 'WARNING'
            ELSE 'ERROR'
        END,
        'EMAIL',
        'qa-team@example.com',
        'Quality Status'
    );
    
    -- Enviar notificação
    PERFORM notification.send_notification(v_notification_id);
    
    RETURN v_notification_id;
END;
$$ LANGUAGE plpgsql;

-- 18. Função para Notificar Status de Conformidade com Regulamentações
CREATE OR REPLACE FUNCTION notification.notify_regulatory_compliance(
    p_test_id INTEGER,
    p_test_case_id INTEGER,
    p_regulation VARCHAR(50),
    p_status VARCHAR(20)
)
RETURNS UUID AS $$
DECLARE
    v_notification_id UUID;
BEGIN
    -- Registrar notificação de conformidade com regulamentações
    v_notification_id := notification.register_notification(
        p_test_id,
        p_test_case_id,
        p_regulation || ' compliance status: ' || p_status,
        CASE 
            WHEN p_status = 'COMPLIANT' THEN 'INFO'
            ELSE 'ERROR'
        END,
        'EMAIL',
        'compliance-team@example.com',
        'Regulatory Compliance'
    );
    
    -- Enviar notificação
    PERFORM notification.send_notification(v_notification_id);
    
    RETURN v_notification_id;
END;
$$ LANGUAGE plpgsql;

-- 19. Função para Notificar Status de Segurança do Usuário
CREATE OR REPLACE FUNCTION notification.notify_user_security_status(
    p_test_id INTEGER,
    p_test_case_id INTEGER,
    p_user_id VARCHAR(50),
    p_security_level VARCHAR(20)
)
RETURNS UUID AS $$
DECLARE
    v_notification_id UUID;
BEGIN
    -- Registrar notificação de status de segurança do usuário
    v_notification_id := notification.register_notification(
        p_test_id,
        p_test_case_id,
        'User ' || p_user_id || ' security level: ' || p_security_level,
        CASE 
            WHEN p_security_level = 'HIGH' THEN 'INFO'
            ELSE 'WARNING'
        END,
        'EMAIL',
        'security-team@example.com',
        'User Security Status'
    );
    
    -- Enviar notificação
    PERFORM notification.send_notification(v_notification_id);
    
    RETURN v_notification_id;
END;
$$ LANGUAGE plpgsql;

-- 20. Função para Notificar Status de Segurança do Sistema
CREATE OR REPLACE FUNCTION notification.notify_system_security_status(
    p_test_id INTEGER,
    p_test_case_id INTEGER,
    p_system_id VARCHAR(50),
    p_security_level VARCHAR(20)
)
RETURNS UUID AS $$
DECLARE
    v_notification_id UUID;
BEGIN
    -- Registrar notificação de status de segurança do sistema
    v_notification_id := notification.register_notification(
        p_test_id,
        p_test_case_id,
        'System ' || p_system_id || ' security level: ' || p_security_level,
        CASE 
            WHEN p_security_level = 'HIGH' THEN 'INFO'
            ELSE 'WARNING'
        END,
        'EMAIL',
        'security-team@example.com',
        'System Security Status'
    );
    
    -- Enviar notificação
    PERFORM notification.send_notification(v_notification_id);
    
    RETURN v_notification_id;
END;
$$ LANGUAGE plpgsql;

-- Exemplo de uso
-- SELECT notification.notify_test_success(1, 1);
-- SELECT notification.notify_test_failure(1, 1, 'Test failed due to timeout');
-- SELECT notification.notify_compliance_status(1, 1, 'COMPLIANT');
