-- INNOVABIZ IAM - Migração 004: Dados iniciais para auditoria
-- Autor: Eduardo Jeremias
-- Versão: 1.0.0
-- Descrição: Insere dados iniciais para o sistema de auditoria multi-contexto

-- Políticas de retenção padrão para contexto Brasil (LGPD)
INSERT INTO audit_retention_policies (
    tenant_id,
    regional_context,
    retention_days,
    compliance_framework,
    category,
    description,
    automatic_anonymization,
    anonymization_fields,
    active
) VALUES 
-- Política geral para eventos de autenticação (Brasil - LGPD)
(
    'default',
    'BR',
    730, -- 2 anos (LGPD)
    'LGPD',
    'AUTHENTICATION',
    'Política de retenção padrão para eventos de autenticação - LGPD',
    TRUE,
    ARRAY['source_ip', 'user_name', 'details.device_info'],
    TRUE
),
-- Política para eventos financeiros (Brasil - BACEN)
(
    'default',
    'BR',
    1825, -- 5 anos (BACEN)
    'BACEN',
    'FINANCIAL',
    'Política de retenção para eventos financeiros - BACEN',
    TRUE,
    ARRAY['source_ip', 'http_details'],
    TRUE
);

-- Políticas de retenção padrão para contexto Estados Unidos (SOX)
INSERT INTO audit_retention_policies (
    tenant_id,
    regional_context,
    retention_days,
    compliance_framework,
    category,
    description,
    automatic_anonymization,
    anonymization_fields,
    active
) VALUES 
-- Política geral para eventos financeiros (EUA - SOX)
(
    'default',
    'US',
    2555, -- 7 anos (SOX)
    'SOX',
    'FINANCIAL',
    'Default retention policy for financial events - SOX',
    FALSE,
    ARRAY[],
    TRUE
),
-- Política para eventos de acesso a dados (EUA - NIST)
(
    'default',
    'US',
    1095, -- 3 anos (NIST)
    'NIST',
    'DATA_ACCESS',
    'Default retention policy for data access events - NIST',
    TRUE,
    ARRAY['source_ip'],
    TRUE
);

-- Políticas de retenção padrão para contexto União Europeia (GDPR)
INSERT INTO audit_retention_policies (
    tenant_id,
    regional_context,
    retention_days,
    compliance_framework,
    category,
    description,
    automatic_anonymization,
    anonymization_fields,
    active
) VALUES 
-- Política geral para eventos de consentimento (UE - GDPR)
(
    'default',
    'EU',
    1095, -- 3 anos (GDPR)
    'GDPR',
    'CONSENT',
    'Default retention policy for consent events - GDPR',
    TRUE,
    ARRAY['source_ip', 'user_name', 'http_details'],
    TRUE
),
-- Política para eventos de acesso a dados (UE - GDPR)
(
    'default',
    'EU',
    730, -- 2 anos (GDPR)
    'GDPR',
    'DATA_ACCESS',
    'Default retention policy for data access events - GDPR',
    TRUE,
    ARRAY['source_ip', 'user_name', 'details'],
    TRUE
),
-- Política para eventos financeiros (UE - PSD2)
(
    'default',
    'EU',
    1825, -- 5 anos (PSD2)
    'PSD2',
    'FINANCIAL',
    'Default retention policy for financial events - PSD2',
    TRUE,
    ARRAY['source_ip'],
    TRUE
);

-- Políticas de retenção padrão para contexto Angola (BNA)
INSERT INTO audit_retention_policies (
    tenant_id,
    regional_context,
    retention_days,
    compliance_framework,
    category,
    description,
    automatic_anonymization,
    anonymization_fields,
    active
) VALUES 
-- Política geral para eventos financeiros (Angola - BNA)
(
    'default',
    'AO',
    3650, -- 10 anos (BNA)
    'BNA',
    'FINANCIAL',
    'Política de retenção para eventos financeiros - BNA',
    FALSE,
    ARRAY[],
    TRUE
),
-- Política para eventos de autenticação (Angola)
(
    'default',
    'AO',
    1095, -- 3 anos
    'BNA',
    'AUTHENTICATION',
    'Política de retenção para eventos de autenticação - Angola',
    TRUE,
    ARRAY['source_ip', 'user_name'],
    TRUE
);

-- Política global para PCI DSS (aplicável a todos os contextos)
INSERT INTO audit_retention_policies (
    tenant_id,
    regional_context,
    retention_days,
    compliance_framework,
    category,
    description,
    automatic_anonymization,
    anonymization_fields,
    active
) VALUES 
(
    'default',
    NULL, -- Aplica-se a todos os contextos regionais
    1095, -- 3 anos (PCI DSS)
    'PCI_DSS',
    'FINANCIAL',
    'Global retention policy for payment card data - PCI DSS',
    TRUE,
    ARRAY['source_ip', 'details.card_data', 'details.payment_info'],
    TRUE
);

-- Estatísticas iniciais vazias para cada contexto regional
INSERT INTO audit_statistics (
    tenant_id,
    regional_context,
    statistics_type,
    period_start,
    period_end,
    statistics_data
) VALUES 
(
    'default',
    'BR',
    'INITIAL_SETUP',
    NOW() - INTERVAL '1 day',
    NOW(),
    '{"setup_completed": true, "policies_created": 2, "compliance_frameworks": ["LGPD", "BACEN", "PCI_DSS"]}'::JSONB
),
(
    'default',
    'US',
    'INITIAL_SETUP',
    NOW() - INTERVAL '1 day',
    NOW(),
    '{"setup_completed": true, "policies_created": 2, "compliance_frameworks": ["SOX", "NIST", "PCI_DSS"]}'::JSONB
),
(
    'default',
    'EU',
    'INITIAL_SETUP',
    NOW() - INTERVAL '1 day',
    NOW(),
    '{"setup_completed": true, "policies_created": 3, "compliance_frameworks": ["GDPR", "PSD2", "PCI_DSS"]}'::JSONB
),
(
    'default',
    'AO',
    'INITIAL_SETUP',
    NOW() - INTERVAL '1 day',
    NOW(),
    '{"setup_completed": true, "policies_created": 2, "compliance_frameworks": ["BNA", "PCI_DSS"]}'::JSONB
);

-- Comentários para documentação
COMMENT ON TABLE audit_retention_policies IS 'Políticas de retenção e anonimização para cada contexto regional e framework de compliance';
COMMENT ON TABLE audit_statistics IS 'Estatísticas de auditoria por contexto regional e período';