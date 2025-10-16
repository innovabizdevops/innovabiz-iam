-- =====================================================================
-- INNOVABIZ - Validadores de Conformidade para Ecossistema Open X
-- Versão: 1.0.0
-- Data de Criação: 15/05/2025
-- Autor: INNOVABIZ DevOps
-- Descrição: Funções e procedimentos para validação de conformidade IAM
--            específicos para Open Insurance, Open Health e Open Government
-- Regiões Suportadas: UE/Portugal, Brasil, Angola, EUA
-- =====================================================================

-- Criação de schema para validadores de conformidade se ainda não existir
CREATE SCHEMA IF NOT EXISTS compliance_validators;

-- =====================================================================
-- Tabelas de Referência para Requisitos Regulatórios - Open Insurance
-- =====================================================================

-- Requisitos de Solvência II para IAM em Open Insurance
CREATE TABLE compliance_validators.solvency_ii_requirements (
    requirement_id VARCHAR(20) PRIMARY KEY,
    requirement_name VARCHAR(255) NOT NULL,
    requirement_description TEXT NOT NULL,
    validation_query TEXT,
    implementation_details TEXT,
    required_auth_level VARCHAR(50),
    required_auth_methods VARCHAR(50)[],
    irr_threshold VARCHAR(10),
    is_mandatory BOOLEAN NOT NULL DEFAULT TRUE,
    applies_to_functions VARCHAR(50)[],
    relevant_articles VARCHAR(50)[],
    reference_url TEXT
);

-- Requisitos para SUSEP (Brasil - Open Insurance)
CREATE TABLE compliance_validators.susep_requirements (
    requirement_id VARCHAR(20) PRIMARY KEY,
    requirement_name VARCHAR(255) NOT NULL,
    requirement_description TEXT NOT NULL,
    validation_query TEXT,
    implementation_details TEXT,
    required_auth_level VARCHAR(50),
    required_auth_methods VARCHAR(50)[],
    irr_threshold VARCHAR(10),
    is_mandatory BOOLEAN NOT NULL DEFAULT TRUE,
    applies_to_phases VARCHAR(50)[],
    relevant_resolutions VARCHAR(50)[],
    reference_url TEXT
);

-- =====================================================================
-- Inserção de dados básicos para Requisitos Solvência II (Open Insurance)
-- =====================================================================

INSERT INTO compliance_validators.solvency_ii_requirements (
    requirement_id, requirement_name, requirement_description, 
    validation_query, implementation_details, 
    required_auth_level, required_auth_methods, irr_threshold,
    relevant_articles, reference_url
) VALUES
('SOLV2-IAM-01', 'Mecanismos de Governança de Dados', 
 'Controles de acesso e sistemas de autenticação para garantir a governança de dados',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies 
  WHERE tenant_id = $1 
  AND policy_rules::jsonb @> ''{"governance_controls": true}''
  AND is_enabled = TRUE',
 'Verificar a existência de políticas com controles de governança de dados',
 'VERY_ADVANCED', ARRAY['PR-01-02'], 'R1',
 ARRAY['Art. 44'], 'https://eur-lex.europa.eu/legal-content/PT/TXT/?uri=CELEX:32009L0138'),

('SOLV2-IAM-02', 'Segregação de Funções de Controle de Riscos', 
 'Controles de acesso para segregar funções de controle de riscos',
 'SELECT COUNT(*) > 0 FROM iam_core.role_permissions 
  WHERE tenant_id = $1 
  AND role_type = ''RISK_CONTROL''
  AND permissions_data::jsonb ? ''segregation_enforced''',
 'Verificar a existência de segregação de funções para controle de riscos',
 'ADVANCED', ARRAY['OI-01-02'], 'R2',
 ARRAY['Art. 41'], 'https://eur-lex.europa.eu/legal-content/PT/TXT/?uri=CELEX:32009L0138'),

('SOLV2-IAM-03', 'Identificação de Usuários para Auditoria', 
 'Rastreabilidade de ações para fins de auditoria e conformidade',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_history 
  WHERE tenant_id = $1 
  AND event_details::jsonb ? ''audit_trail''',
 'Verificar a existência de trilhas de auditoria detalhadas',
 'ADVANCED', ARRAY['OI-01-01', 'OI-01-03'], 'R1',
 ARRAY['Art. 47'], 'https://eur-lex.europa.eu/legal-content/PT/TXT/?uri=CELEX:32009L0138'),

('SOLV2-IAM-04', 'Medidas de Segurança para Dados Sensíveis', 
 'Proteção adicional para dados de clientes e informações confidenciais',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_methods 
  WHERE tenant_id = $1 
  AND method_parameters::jsonb @> ''{"sensitive_data_protection": true}''',
 'Verificar a existência de proteções específicas para dados sensíveis',
 'VERY_ADVANCED', ARRAY['PR-01-01', 'PR-01-03'], 'R1',
 ARRAY['Art. 35, Art. 46'], 'https://eur-lex.europa.eu/legal-content/PT/TXT/?uri=CELEX:32009L0138'),

('SOLV2-IAM-05', 'Consentimento e Gestão de Identidade', 
 'Gerenciamento de consentimento para compartilhamento de dados de seguros',
 'SELECT COUNT(*) > 0 FROM iam_core.consent_records 
  WHERE tenant_id = $1 
  AND consent_data::jsonb ? ''insurance_data_sharing''',
 'Verificar a gestão de consentimento para dados de seguros',
 'ADVANCED', ARRAY['OI-01-01', 'OI-01-04'], 'R2',
 ARRAY['Art. 29, Art. 30'], 'https://eur-lex.europa.eu/legal-content/PT/TXT/?uri=CELEX:32009L0138');

-- =====================================================================
-- Inserção de dados básicos para Requisitos SUSEP (Brasil - Open Insurance)
-- =====================================================================

INSERT INTO compliance_validators.susep_requirements (
    requirement_id, requirement_name, requirement_description, 
    validation_query, implementation_details, 
    required_auth_level, required_auth_methods, irr_threshold,
    applies_to_phases, relevant_resolutions, reference_url
) VALUES
('SUSEP-IAM-01', 'Diretório de Participantes Open Insurance', 
 'Integração com o diretório oficial de participantes do Open Insurance Brasil',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND config_parameters::jsonb ? ''opin_directory''',
 'Verificar integração com o diretório Open Insurance',
 'VERY_ADVANCED', ARRAY['OI-02-01'], 'R1',
 ARRAY['Fase 1', 'Fase 2'], 
 ARRAY['Resolução CNSP Nº 415/2021'], 'https://www.gov.br/susep/'),

('SUSEP-IAM-02', 'Consentimento para Compartilhamento de Dados', 
 'Mecanismos de consentimento para compartilhamento de dados de seguros',
 'SELECT COUNT(*) > 0 FROM iam_core.consent_records 
  WHERE tenant_id = $1 
  AND consent_data::jsonb ? ''opin_consent''
  AND consent_status = ''ACTIVE''',
 'Verificar registros de consentimento para Open Insurance',
 'ADVANCED', ARRAY['OI-02-02'], 'R1',
 ARRAY['Fase 2', 'Fase 3'], 
 ARRAY['Resolução CNSP Nº 415/2021'], 'https://www.gov.br/susep/'),

('SUSEP-IAM-03', 'Certificados ICP-Brasil', 
 'Uso de certificados ICP-Brasil para autenticação de APIs',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND factor_type = ''DIGITAL_CERTIFICATE''
  AND config_parameters::jsonb @> ''{"certificate_type": "ICP_BRASIL"}''',
 'Verificar uso de certificados ICP-Brasil',
 'VERY_ADVANCED', ARRAY['OI-02-01', 'OI-02-03'], 'R1',
 ARRAY['Fase 1', 'Fase 2'], 
 ARRAY['Resolução CNSP Nº 415/2021'], 'https://www.gov.br/susep/'),

('SUSEP-IAM-04', 'Proteção de Dados LGPD', 
 'Conformidade com LGPD para dados pessoais de seguros',
 'SELECT COUNT(*) > 0 FROM iam_core.data_protection_settings 
  WHERE tenant_id = $1 
  AND settings_data::jsonb @> ''{"lgpd_compliant": true}''',
 'Verificar conformidade com LGPD para dados de seguros',
 'ADVANCED', ARRAY['PR-02-01'], 'R1',
 ARRAY['Fase 1', 'Fase 2', 'Fase 3'], 
 ARRAY['Resolução CNSP Nº 415/2021', 'Lei Nº 13.709/2018'], 'https://www.gov.br/susep/'),

('SUSEP-IAM-05', 'Rastreabilidade de Operações', 
 'Logs e trilhas de auditoria para operações de Open Insurance',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_history 
  WHERE tenant_id = $1 
  AND activity_type = ''OPIN_API_ACCESS''',
 'Verificar logs de acesso às APIs de Open Insurance',
 'ADVANCED', ARRAY['OI-02-01', 'OI-02-04'], 'R1',
 ARRAY['Fase 2', 'Fase 3'], 
 ARRAY['Resolução CNSP Nº 415/2021'], 'https://www.gov.br/susep/');

-- =====================================================================
-- Funções de Validação - Open Insurance
-- =====================================================================

-- Função para validar conformidade com Solvência II (Open Insurance)
CREATE OR REPLACE FUNCTION compliance_validators.validate_solvency_ii_compliance(
    tenant_id UUID
) RETURNS TABLE (
    requirement_id VARCHAR(20),
    requirement_name VARCHAR(255),
    is_compliant BOOLEAN,
    details TEXT
) AS $$
DECLARE
    req RECORD;
    query_result BOOLEAN;
    exec_query TEXT;
BEGIN
    FOR req IN SELECT * FROM compliance_validators.solvency_ii_requirements LOOP
        -- Preparar consulta de validação
        exec_query := 'SELECT (' || req.validation_query || ') AS result';
        
        -- Executar consulta de validação
        EXECUTE exec_query USING tenant_id INTO query_result;
        
        -- Retornar resultado
        requirement_id := req.requirement_id;
        requirement_name := req.requirement_name;
        is_compliant := query_result;
        
        IF query_result THEN
            details := 'Conformidade verificada para ' || req.requirement_name;
        ELSE
            details := 'Não-conformidade detectada para ' || req.requirement_name || '. ' || 
                      'Recomendação: ' || req.implementation_details;
        END IF;
        
        RETURN NEXT;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Função para validar conformidade com SUSEP (Brasil - Open Insurance)
CREATE OR REPLACE FUNCTION compliance_validators.validate_susep_compliance(
    tenant_id UUID
) RETURNS TABLE (
    requirement_id VARCHAR(20),
    requirement_name VARCHAR(255),
    is_compliant BOOLEAN,
    details TEXT
) AS $$
DECLARE
    req RECORD;
    query_result BOOLEAN;
    exec_query TEXT;
BEGIN
    FOR req IN SELECT * FROM compliance_validators.susep_requirements LOOP
        -- Preparar consulta de validação
        exec_query := 'SELECT (' || req.validation_query || ') AS result';
        
        -- Executar consulta de validação
        EXECUTE exec_query USING tenant_id INTO query_result;
        
        -- Retornar resultado
        requirement_id := req.requirement_id;
        requirement_name := req.requirement_name;
        is_compliant := query_result;
        
        IF query_result THEN
            details := 'Conformidade verificada para ' || req.requirement_name;
        ELSE
            details := 'Não-conformidade detectada para ' || req.requirement_name || '. ' || 
                      'Recomendação: ' || req.implementation_details;
        END IF;
        
        RETURN NEXT;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Tabelas de Referência para Requisitos Regulatórios - Open Health
-- =====================================================================

-- Requisitos de HIPAA e GDPR para IAM em Open Health
CREATE TABLE compliance_validators.open_health_requirements (
    requirement_id VARCHAR(20) PRIMARY KEY,
    requirement_name VARCHAR(255) NOT NULL,
    requirement_description TEXT NOT NULL,
    validation_query TEXT,
    implementation_details TEXT,
    required_auth_level VARCHAR(50),
    required_auth_methods VARCHAR(50)[],
    irr_threshold VARCHAR(10),
    is_mandatory BOOLEAN NOT NULL DEFAULT TRUE,
    applies_to_operations VARCHAR(50)[],
    relevant_regulations VARCHAR(50)[],
    reference_url TEXT
);

-- Requisitos para ANS (Brasil - Open Health)
CREATE TABLE compliance_validators.ans_requirements (
    requirement_id VARCHAR(20) PRIMARY KEY,
    requirement_name VARCHAR(255) NOT NULL,
    requirement_description TEXT NOT NULL,
    validation_query TEXT,
    implementation_details TEXT,
    required_auth_level VARCHAR(50),
    required_auth_methods VARCHAR(50)[],
    irr_threshold VARCHAR(10),
    is_mandatory BOOLEAN NOT NULL DEFAULT TRUE,
    applies_to_phases VARCHAR(50)[],
    relevant_resolutions VARCHAR(50)[],
    reference_url TEXT
);

-- =====================================================================
-- Tabelas de Referência para Requisitos Regulatórios - Open Government
-- =====================================================================

-- Requisitos para eIDAS e Open Government (UE/Portugal)
CREATE TABLE compliance_validators.eidas_gov_requirements (
    requirement_id VARCHAR(20) PRIMARY KEY,
    requirement_name VARCHAR(255) NOT NULL,
    requirement_description TEXT NOT NULL,
    validation_query TEXT,
    implementation_details TEXT,
    required_auth_level VARCHAR(50),
    required_auth_methods VARCHAR(50)[],
    irr_threshold VARCHAR(10),
    is_mandatory BOOLEAN NOT NULL DEFAULT TRUE,
    applies_to_services VARCHAR(50)[],
    relevant_articles VARCHAR(50)[],
    reference_url TEXT
);

-- Requisitos para Governo Digital (Brasil - Open Government)
CREATE TABLE compliance_validators.gov_br_requirements (
    requirement_id VARCHAR(20) PRIMARY KEY,
    requirement_name VARCHAR(255) NOT NULL,
    requirement_description TEXT NOT NULL,
    validation_query TEXT,
    implementation_details TEXT,
    required_auth_level VARCHAR(50),
    required_auth_methods VARCHAR(50)[],
    irr_threshold VARCHAR(10),
    is_mandatory BOOLEAN NOT NULL DEFAULT TRUE,
    applies_to_phases VARCHAR(50)[],
    relevant_regulations VARCHAR(50)[],
    reference_url TEXT
);

-- =====================================================================
-- Inserção de dados básicos para Requisitos Open Health
-- =====================================================================

INSERT INTO compliance_validators.open_health_requirements (
    requirement_id, requirement_name, requirement_description, 
    validation_query, implementation_details, 
    required_auth_level, required_auth_methods, irr_threshold,
    relevant_regulations, reference_url
) VALUES
('HEALTH-IAM-01', 'Autenticação para Acesso de Dados de Saúde', 
 'Autenticação forte para acesso a informações de saúde protegidas',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies 
  WHERE tenant_id = $1 
  AND policy_rules::jsonb @> ''{"health_data_access": true}''
  AND policy_rules::jsonb @> ''{"min_factors_required": 2}''',
 'Verificar políticas de autenticação forte para dados de saúde',
 'VERY_ADVANCED', ARRAY['OH-01-01', 'OH-01-02'], 'R1',
 ARRAY['HIPAA §164.312', 'GDPR Art. 9'], 'https://www.hhs.gov/hipaa/'),

('HEALTH-IAM-02', 'Consentimento Específico para Compartilhamento', 
 'Consentimento explícito para compartilhamento de dados de saúde',
 'SELECT COUNT(*) > 0 FROM iam_core.consent_records 
  WHERE tenant_id = $1 
  AND consent_data::jsonb ? ''health_data_sharing''
  AND consent_scope::jsonb ? ''explicit''',
 'Verificar registros de consentimento específico para dados de saúde',
 'ADVANCED', ARRAY['OH-01-03'], 'R1',
 ARRAY['GDPR Art. 9(2)(a)', 'HIPAA §164.508'], 'https://gdpr-info.eu/'),

('HEALTH-IAM-03', 'Trilhas de Auditoria de Acesso', 
 'Mecanismos detalhados de registro de acesso a dados de saúde',
 'SELECT COUNT(*) > 0 FROM iam_core.access_logs 
  WHERE tenant_id = $1 
  AND resource_type = ''HEALTH_DATA''
  AND log_details::jsonb ? ''complete_audit_trail''',
 'Verificar trilhas de auditoria para dados de saúde',
 'ADVANCED', ARRAY['OH-01-01', 'OH-01-04'], 'R1',
 ARRAY['HIPAA §164.312(b)', 'GDPR Art. 30'], 'https://www.hhs.gov/hipaa/'),

('HEALTH-IAM-04', 'Gestão de Identidade para Profissionais de Saúde', 
 'Verificação da identidade e credenciais de profissionais de saúde',
 'SELECT COUNT(*) > 0 FROM iam_core.identity_verification 
  WHERE tenant_id = $1 
  AND verification_type = ''HEALTHCARE_PROFESSIONAL''
  AND verification_parameters::jsonb ? ''credentials_check''',
 'Verificar processos de validação para credenciais de profissionais de saúde',
 'VERY_ADVANCED', ARRAY['OH-01-05'], 'R1',
 ARRAY['HIPAA §164.308', 'GDPR Art. 32'], 'https://www.hhs.gov/hipaa/'),

('HEALTH-IAM-05', 'Revogação de Acesso Emergencial', 
 'Capacidade de revogar acessos em situações de emergência',
 'SELECT COUNT(*) > 0 FROM iam_core.emergency_access_controls 
  WHERE tenant_id = $1 
  AND control_type = ''EMERGENCY_REVOCATION''
  AND is_active = TRUE',
 'Verificar mecanismos de revogação emergencial de acesso',
 'ADVANCED', ARRAY['OH-01-01', 'OH-01-06'], 'R1',
 ARRAY['HIPAA §164.312(a)(2)(ii)', 'GDPR Art. 32'], 'https://www.hhs.gov/hipaa/');

-- =====================================================================
-- Inserção de dados básicos para Requisitos ANS (Brasil - Open Health)
-- =====================================================================

INSERT INTO compliance_validators.ans_requirements (
    requirement_id, requirement_name, requirement_description, 
    validation_query, implementation_details, 
    required_auth_level, required_auth_methods, irr_threshold,
    applies_to_phases, relevant_resolutions, reference_url
) VALUES
('ANS-IAM-01', 'Diretório de Participantes Open Health', 
 'Integração com o diretório de participantes do Open Health Brasil',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND config_parameters::jsonb ? ''openhealth_directory''',
 'Verificar integração com o diretório Open Health',
 'VERY_ADVANCED', ARRAY['OH-02-01'], 'R1',
 ARRAY['Fase 1'], 
 ARRAY['RN Nº 501/2022'], 'https://www.gov.br/ans/'),

('ANS-IAM-02', 'Consentimento para Compartilhamento de Dados de Saúde', 
 'Mecanismos de consentimento específicos para compartilhamento de dados de saúde',
 'SELECT COUNT(*) > 0 FROM iam_core.consent_records 
  WHERE tenant_id = $1 
  AND consent_data::jsonb ? ''openhealth_consent''
  AND consent_status = ''ACTIVE''',
 'Verificar registros de consentimento para Open Health',
 'ADVANCED', ARRAY['OH-02-02'], 'R1',
 ARRAY['Fase 1', 'Fase 2'], 
 ARRAY['RN Nº 501/2022', 'Lei Nº 13.709/2018'], 'https://www.gov.br/ans/'),

('ANS-IAM-03', 'Certificados ICP-Brasil para Dados de Saúde', 
 'Uso de certificados ICP-Brasil para autenticação em APIs de saúde',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND factor_type = ''DIGITAL_CERTIFICATE''
  AND config_parameters::jsonb @> ''{"certificate_type": "ICP_BRASIL", "health_data": true}''',
 'Verificar uso de certificados ICP-Brasil para dados de saúde',
 'VERY_ADVANCED', ARRAY['OH-02-01', 'OH-02-03'], 'R1',
 ARRAY['Fase 1'], 
 ARRAY['RN Nº 501/2022'], 'https://www.gov.br/ans/'),

('ANS-IAM-04', 'Proteção de Dados Sensíveis de Saúde', 
 'Proteções especiais para dados sensíveis conforme LGPD e regulações da ANS',
 'SELECT COUNT(*) > 0 FROM iam_core.data_protection_settings 
  WHERE tenant_id = $1 
  AND settings_data::jsonb @> ''{"sensitive_health_data_protection": true}''',
 'Verificar proteções específicas para dados sensíveis de saúde',
 'VERY_ADVANCED', ARRAY['OH-02-04', 'PR-02-01'], 'R1',
 ARRAY['Fase 1', 'Fase 2'], 
 ARRAY['RN Nº 501/2022', 'Lei Nº 13.709/2018'], 'https://www.gov.br/ans/'),

('ANS-IAM-05', 'Controle de Acesso Baseado em Papéis para Profissionais', 
 'Controle de acesso granular para diferentes papéis no ecossistema de saúde',
 'SELECT COUNT(*) > 0 FROM iam_core.role_definitions 
  WHERE tenant_id = $1 
  AND role_context = ''HEALTHCARE''
  AND role_definitions::jsonb ? ''granular_health_access''',
 'Verificar controles de acesso baseados em papéis para profissionais de saúde',
 'ADVANCED', ARRAY['OH-02-05'], 'R1',
 ARRAY['Fase 2'], 
 ARRAY['RN Nº 501/2022'], 'https://www.gov.br/ans/');

-- =====================================================================
-- Inserção de dados básicos para Requisitos eIDAS (UE/PT - Open Government)
-- =====================================================================

INSERT INTO compliance_validators.eidas_gov_requirements (
    requirement_id, requirement_name, requirement_description, 
    validation_query, implementation_details, 
    required_auth_level, required_auth_methods, irr_threshold,
    relevant_articles, reference_url
) VALUES
('EIDAS-IAM-01', 'Identificação Eletrônica Notificada', 
 'Suporte para meios de identificação eletrônica notificados conforme eIDAS',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_methods 
  WHERE tenant_id = $1 
  AND method_parameters::jsonb @> ''{"eidas_notified": true}''',
 'Verificar suporte para meios de identificação notificados eIDAS',
 'VERY_ADVANCED', ARRAY['OG-01-01'], 'R1',
 ARRAY['Art. 6, Art. 7'], 'https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=uriserv:OJ.L_.2014.257.01.0073.01.ENG'),

('EIDAS-IAM-02', 'Níveis de Garantia de Autenticação', 
 'Implementação de níveis de garantia baixo, substancial e elevado',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies 
  WHERE tenant_id = $1 
  AND policy_rules::jsonb @> ''{"eidas_assurance_levels": ["LOW", "SUBSTANTIAL", "HIGH"]}''',
 'Verificar suporte para os três níveis de garantia eIDAS',
 'ADVANCED', ARRAY['OG-01-02'], 'R1',
 ARRAY['Art. 8'], 'https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=uriserv:OJ.L_.2014.257.01.0073.01.ENG'),

('EIDAS-IAM-03', 'Interoperabilidade Transfronteiriça', 
 'Capacidade de aceitar identificação de outros Estados-Membros',
 'SELECT COUNT(*) > 0 FROM iam_core.interoperability_settings 
  WHERE tenant_id = $1 
  AND settings_data::jsonb @> ''{"eidas_cross_border": true}''',
 'Verificar suporte para interoperabilidade transfronteiriça',
 'VERY_ADVANCED', ARRAY['OG-01-03'], 'R1',
 ARRAY['Art. 12'], 'https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=uriserv:OJ.L_.2014.257.01.0073.01.ENG'),

('EIDAS-IAM-04', 'Assinaturas e Selos Eletrônicos', 
 'Suporte para assinaturas e selos eletrônicos qualificados',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND factor_type = ''DIGITAL_SIGNATURE''
  AND config_parameters::jsonb @> ''{"qualified_signature": true}''',
 'Verificar suporte para assinaturas eletrônicas qualificadas',
 'VERY_ADVANCED', ARRAY['OG-01-04'], 'R1',
 ARRAY['Art. 25, Art. 35'], 'https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=uriserv:OJ.L_.2014.257.01.0073.01.ENG'),

('EIDAS-IAM-05', 'Autenticação de Sites na Web', 
 'Conformidade com requisitos para certificados qualificados de autenticação',
 'SELECT COUNT(*) > 0 FROM iam_core.tls_configurations 
  WHERE tenant_id = $1 
  AND config_parameters::jsonb @> ''{"eidas_qwac": true}''',
 'Verificar suporte para certificados qualificados de autenticação web',
 'ADVANCED', ARRAY['OG-01-05'], 'R1',
 ARRAY['Art. 45'], 'https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=uriserv:OJ.L_.2014.257.01.0073.01.ENG');

-- =====================================================================
-- Inserção de dados básicos para Requisitos Gov BR (Brasil - Open Government)
-- =====================================================================

INSERT INTO compliance_validators.gov_br_requirements (
    requirement_id, requirement_name, requirement_description, 
    validation_query, implementation_details, 
    required_auth_level, required_auth_methods, irr_threshold,
    applies_to_phases, relevant_regulations, reference_url
) VALUES
('GOVBR-IAM-01', 'Integração com Gov.br', 
 'Integração com o sistema de identidade Gov.br para autenticação',
 'SELECT COUNT(*) > 0 FROM iam_core.external_idp_configurations 
  WHERE tenant_id = $1 
  AND provider_id = ''GOVBR''
  AND is_active = TRUE',
 'Verificar integração ativa com Gov.br',
 'VERY_ADVANCED', ARRAY['OG-02-01'], 'R1',
 ARRAY['Fase 1'], 
 ARRAY['Decreto Nº 10.332/2020'], 'https://www.gov.br/governodigital/'),

('GOVBR-IAM-02', 'Níveis de Autenticação do Gov.br', 
 'Suporte para diferentes níveis de autenticação do Gov.br (bronze, prata, ouro)',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies 
  WHERE tenant_id = $1 
  AND policy_rules::jsonb @> ''{"govbr_levels": ["BRONZE", "PRATA", "OURO"]}''',
 'Verificar suporte para níveis de autenticação do Gov.br',
 'ADVANCED', ARRAY['OG-02-02'], 'R1',
 ARRAY['Fase 1', 'Fase 2'], 
 ARRAY['Decreto Nº 10.332/2020'], 'https://www.gov.br/governodigital/'),

('GOVBR-IAM-03', 'Certificados ICP-Brasil para Serviços Governamentais', 
 'Uso de certificados ICP-Brasil para autenticação em serviços governamentais',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND factor_type = ''DIGITAL_CERTIFICATE''
  AND config_parameters::jsonb @> ''{"certificate_type": "ICP_BRASIL", "gov_service": true}''',
 'Verificar uso de certificados ICP-Brasil para serviços governamentais',
 'VERY_ADVANCED', ARRAY['OG-02-03'], 'R1',
 ARRAY['Fase 1'], 
 ARRAY['MP 2.200-2/2001'], 'https://www.iti.gov.br/'),

('GOVBR-IAM-04', 'Interoperabilidade entre Órgãos', 
 'Capacidade de interoperabilidade entre diferentes órgãos governamentais',
 'SELECT COUNT(*) > 0 FROM iam_core.interoperability_settings 
  WHERE tenant_id = $1 
  AND settings_data::jsonb @> ''{"government_interop": true}''',
 'Verificar configurações de interoperabilidade entre órgãos',
 'ADVANCED', ARRAY['OG-02-04'], 'R1',
 ARRAY['Fase 2'], 
 ARRAY['Decreto Nº 10.332/2020'], 'https://www.gov.br/governodigital/'),

('GOVBR-IAM-05', 'Proteção de Dados LGPD para Dados Governamentais', 
 'Conformidade com LGPD para proteção de dados pessoais em serviços governamentais',
 'SELECT COUNT(*) > 0 FROM iam_core.data_protection_settings 
  WHERE tenant_id = $1 
  AND settings_data::jsonb @> ''{"lgpd_compliant": true, "government_data": true}''',
 'Verificar conformidade com LGPD para dados governamentais',
 'ADVANCED', ARRAY['OG-02-05', 'PR-02-01'], 'R1',
 ARRAY['Fase 1', 'Fase 2'], 
 ARRAY['Lei Nº 13.709/2018'], 'https://www.gov.br/governodigital/');

-- =====================================================================
-- Funções de Validação - Open Health
-- =====================================================================

-- Função para validar conformidade com requisitos Open Health
CREATE OR REPLACE FUNCTION compliance_validators.validate_open_health_compliance(
    tenant_id UUID
) RETURNS TABLE (
    requirement_id VARCHAR(20),
    requirement_name VARCHAR(255),
    is_compliant BOOLEAN,
    details TEXT
) AS $$
DECLARE
    req RECORD;
    query_result BOOLEAN;
    exec_query TEXT;
BEGIN
    FOR req IN SELECT * FROM compliance_validators.open_health_requirements LOOP
        -- Preparar consulta de validação
        exec_query := 'SELECT (' || req.validation_query || ') AS result';
        
        -- Executar consulta de validação
        EXECUTE exec_query USING tenant_id INTO query_result;
        
        -- Retornar resultado
        requirement_id := req.requirement_id;
        requirement_name := req.requirement_name;
        is_compliant := query_result;
        
        IF query_result THEN
            details := 'Conformidade verificada para ' || req.requirement_name;
        ELSE
            details := 'Não-conformidade detectada para ' || req.requirement_name || '. ' || 
                      'Recomendação: ' || req.implementation_details;
        END IF;
        
        RETURN NEXT;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Função para validar conformidade com ANS (Brasil - Open Health)
CREATE OR REPLACE FUNCTION compliance_validators.validate_ans_compliance(
    tenant_id UUID
) RETURNS TABLE (
    requirement_id VARCHAR(20),
    requirement_name VARCHAR(255),
    is_compliant BOOLEAN,
    details TEXT
) AS $$
DECLARE
    req RECORD;
    query_result BOOLEAN;
    exec_query TEXT;
BEGIN
    FOR req IN SELECT * FROM compliance_validators.ans_requirements LOOP
        -- Preparar consulta de validação
        exec_query := 'SELECT (' || req.validation_query || ') AS result';
        
        -- Executar consulta de validação
        EXECUTE exec_query USING tenant_id INTO query_result;
        
        -- Retornar resultado
        requirement_id := req.requirement_id;
        requirement_name := req.requirement_name;
        is_compliant := query_result;
        
        IF query_result THEN
            details := 'Conformidade verificada para ' || req.requirement_name;
        ELSE
            details := 'Não-conformidade detectada para ' || req.requirement_name || '. ' || 
                      'Recomendação: ' || req.implementation_details;
        END IF;
        
        RETURN NEXT;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Funções de Validação - Open Government
-- =====================================================================

-- Função para validar conformidade com eIDAS (UE/PT - Open Government)
CREATE OR REPLACE FUNCTION compliance_validators.validate_eidas_gov_compliance(
    tenant_id UUID
) RETURNS TABLE (
    requirement_id VARCHAR(20),
    requirement_name VARCHAR(255),
    is_compliant BOOLEAN,
    details TEXT
) AS $$
DECLARE
    req RECORD;
    query_result BOOLEAN;
    exec_query TEXT;
BEGIN
    FOR req IN SELECT * FROM compliance_validators.eidas_gov_requirements LOOP
        -- Preparar consulta de validação
        exec_query := 'SELECT (' || req.validation_query || ') AS result';
        
        -- Executar consulta de validação
        EXECUTE exec_query USING tenant_id INTO query_result;
        
        -- Retornar resultado
        requirement_id := req.requirement_id;
        requirement_name := req.requirement_name;
        is_compliant := query_result;
        
        IF query_result THEN
            details := 'Conformidade verificada para ' || req.requirement_name;
        ELSE
            details := 'Não-conformidade detectada para ' || req.requirement_name || '. ' || 
                      'Recomendação: ' || req.implementation_details;
        END IF;
        
        RETURN NEXT;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Função para validar conformidade com Gov BR (Brasil - Open Government)
CREATE OR REPLACE FUNCTION compliance_validators.validate_gov_br_compliance(
    tenant_id UUID
) RETURNS TABLE (
    requirement_id VARCHAR(20),
    requirement_name VARCHAR(255),
    is_compliant BOOLEAN,
    details TEXT
) AS $$
DECLARE
    req RECORD;
    query_result BOOLEAN;
    exec_query TEXT;
BEGIN
    FOR req IN SELECT * FROM compliance_validators.gov_br_requirements LOOP
        -- Preparar consulta de validação
        exec_query := 'SELECT (' || req.validation_query || ') AS result';
        
        -- Executar consulta de validação
        EXECUTE exec_query USING tenant_id INTO query_result;
        
        -- Retornar resultado
        requirement_id := req.requirement_id;
        requirement_name := req.requirement_name;
        is_compliant := query_result;
        
        IF query_result THEN
            details := 'Conformidade verificada para ' || req.requirement_name;
        ELSE
            details := 'Não-conformidade detectada para ' || req.requirement_name || '. ' || 
                      'Recomendação: ' || req.implementation_details;
        END IF;
        
        RETURN NEXT;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Funções de Consolidação e Relatórios - Open X (Todos os Domínios)
-- =====================================================================

-- Gerar relatório consolidado de conformidade para Open Insurance
CREATE OR REPLACE FUNCTION compliance_validators.generate_open_insurance_compliance_report(
    tenant_id UUID
) RETURNS TABLE (
    framework VARCHAR(50),
    requirement_id VARCHAR(20),
    requirement_name VARCHAR(255),
    is_compliant BOOLEAN,
    details TEXT,
    irr_threshold VARCHAR(10)
) AS $$
BEGIN
    -- Resultados Solvência II
    framework := 'SOLVENCY_II';
    FOR requirement_id, requirement_name, is_compliant, details IN 
        SELECT r.requirement_id, r.requirement_name, v.is_compliant, v.details 
        FROM compliance_validators.validate_solvency_ii_compliance(tenant_id) v
        JOIN compliance_validators.solvency_ii_requirements r ON v.requirement_id = r.requirement_id
    LOOP
        irr_threshold := (SELECT t.irr_threshold FROM compliance_validators.solvency_ii_requirements t WHERE t.requirement_id = requirement_id);
        RETURN NEXT;
    END LOOP;
    
    -- Resultados SUSEP (Brasil)
    framework := 'SUSEP_BR';
    FOR requirement_id, requirement_name, is_compliant, details IN 
        SELECT r.requirement_id, r.requirement_name, v.is_compliant, v.details 
        FROM compliance_validators.validate_susep_compliance(tenant_id) v
        JOIN compliance_validators.susep_requirements r ON v.requirement_id = r.requirement_id
    LOOP
        irr_threshold := (SELECT t.irr_threshold FROM compliance_validators.susep_requirements t WHERE t.requirement_id = requirement_id);
        RETURN NEXT;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Gerar relatório consolidado de conformidade para Open Health
CREATE OR REPLACE FUNCTION compliance_validators.generate_open_health_compliance_report(
    tenant_id UUID
) RETURNS TABLE (
    framework VARCHAR(50),
    requirement_id VARCHAR(20),
    requirement_name VARCHAR(255),
    is_compliant BOOLEAN,
    details TEXT,
    irr_threshold VARCHAR(10)
) AS $$
BEGIN
    -- Resultados HIPAA/GDPR (Geral)
    framework := 'OPEN_HEALTH_GENERAL';
    FOR requirement_id, requirement_name, is_compliant, details IN 
        SELECT r.requirement_id, r.requirement_name, v.is_compliant, v.details 
        FROM compliance_validators.validate_open_health_compliance(tenant_id) v
        JOIN compliance_validators.open_health_requirements r ON v.requirement_id = r.requirement_id
    LOOP
        irr_threshold := (SELECT t.irr_threshold FROM compliance_validators.open_health_requirements t WHERE t.requirement_id = requirement_id);
        RETURN NEXT;
    END LOOP;
    
    -- Resultados ANS (Brasil)
    framework := 'ANS_BR';
    FOR requirement_id, requirement_name, is_compliant, details IN 
        SELECT r.requirement_id, r.requirement_name, v.is_compliant, v.details 
        FROM compliance_validators.validate_ans_compliance(tenant_id) v
        JOIN compliance_validators.ans_requirements r ON v.requirement_id = r.requirement_id
    LOOP
        irr_threshold := (SELECT t.irr_threshold FROM compliance_validators.ans_requirements t WHERE t.requirement_id = requirement_id);
        RETURN NEXT;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Gerar relatório consolidado de conformidade para Open Government
CREATE OR REPLACE FUNCTION compliance_validators.generate_open_government_compliance_report(
    tenant_id UUID
) RETURNS TABLE (
    framework VARCHAR(50),
    requirement_id VARCHAR(20),
    requirement_name VARCHAR(255),
    is_compliant BOOLEAN,
    details TEXT,
    irr_threshold VARCHAR(10)
) AS $$
BEGIN
    -- Resultados eIDAS (UE/Portugal)
    framework := 'EIDAS';
    FOR requirement_id, requirement_name, is_compliant, details IN 
        SELECT r.requirement_id, r.requirement_name, v.is_compliant, v.details 
        FROM compliance_validators.validate_eidas_gov_compliance(tenant_id) v
        JOIN compliance_validators.eidas_gov_requirements r ON v.requirement_id = r.requirement_id
    LOOP
        irr_threshold := (SELECT t.irr_threshold FROM compliance_validators.eidas_gov_requirements t WHERE t.requirement_id = requirement_id);
        RETURN NEXT;
    END LOOP;
    
    -- Resultados Gov BR (Brasil)
    framework := 'GOV_BR';
    FOR requirement_id, requirement_name, is_compliant, details IN 
        SELECT r.requirement_id, r.requirement_name, v.is_compliant, v.details 
        FROM compliance_validators.validate_gov_br_compliance(tenant_id) v
        JOIN compliance_validators.gov_br_requirements r ON v.requirement_id = r.requirement_id
    LOOP
        irr_threshold := (SELECT t.irr_threshold FROM compliance_validators.gov_br_requirements t WHERE t.requirement_id = requirement_id);
        RETURN NEXT;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Função de consolidação para todos os frameworks Open X
CREATE OR REPLACE FUNCTION compliance_validators.generate_open_x_compliance_report(
    tenant_id UUID
) RETURNS TABLE (
    open_x_domain VARCHAR(50),
    framework VARCHAR(50),
    requirement_id VARCHAR(20),
    requirement_name VARCHAR(255),
    is_compliant BOOLEAN,
    details TEXT,
    irr_threshold VARCHAR(10)
) AS $$
BEGIN
    -- Resultados Open Insurance
    open_x_domain := 'OPEN_INSURANCE';
    FOR framework, requirement_id, requirement_name, is_compliant, details, irr_threshold IN 
        SELECT * FROM compliance_validators.generate_open_insurance_compliance_report(tenant_id)
    LOOP
        RETURN NEXT;
    END LOOP;
    
    -- Resultados Open Health
    open_x_domain := 'OPEN_HEALTH';
    FOR framework, requirement_id, requirement_name, is_compliant, details, irr_threshold IN 
        SELECT * FROM compliance_validators.generate_open_health_compliance_report(tenant_id)
    LOOP
        RETURN NEXT;
    END LOOP;
    
    -- Resultados Open Government
    open_x_domain := 'OPEN_GOVERNMENT';
    FOR framework, requirement_id, requirement_name, is_compliant, details, irr_threshold IN 
        SELECT * FROM compliance_validators.generate_open_government_compliance_report(tenant_id)
    LOOP
        RETURN NEXT;
    END LOOP;
    
    -- Incluir também Open Banking (já implementado anteriormente)
    open_x_domain := 'OPEN_BANKING';
    FOR framework, compliance_score, total_requirements, compliant_requirements, compliance_percentage IN 
        SELECT framework, compliance_score, total_requirements, compliant_requirements, compliance_percentage 
        FROM compliance_validators.calculate_openbanking_compliance_score(tenant_id)
    LOOP
        -- Para manter consistência com o restante dos relatórios
        requirement_id := 'OPEN_BANKING';
        requirement_name := 'Resumo de Conformidade ' || framework;
        is_compliant := (compliance_percentage >= 85.0);
        details := 'Pontuação: ' || compliance_score || ', Conformidade: ' || compliance_percentage || '%, Requisitos: ' || 
                   compliant_requirements || '/' || total_requirements;
        irr_threshold := CASE 
                            WHEN compliance_percentage >= 95 THEN 'R1'
                            WHEN compliance_percentage >= 85 THEN 'R2'
                            WHEN compliance_percentage >= 70 THEN 'R3'
                            ELSE 'R4'
                         END;
        RETURN NEXT;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Calcular pontuação de conformidade para domínios Open X
CREATE OR REPLACE FUNCTION compliance_validators.calculate_open_x_compliance_score(
    tenant_id UUID
) RETURNS TABLE (
    open_x_domain VARCHAR(50),
    framework VARCHAR(50),
    compliance_score NUMERIC(4,2),
    total_requirements INTEGER,
    compliant_requirements INTEGER,
    compliance_percentage NUMERIC(5,2)
) AS $$
DECLARE
    -- Contadores para Open Insurance
    solv2_total INTEGER := 0;
    solv2_compliant INTEGER := 0;
    susep_total INTEGER := 0;
    susep_compliant INTEGER := 0;
    
    -- Contadores para Open Health
    health_total INTEGER := 0;
    health_compliant INTEGER := 0;
    ans_total INTEGER := 0;
    ans_compliant INTEGER := 0;
    
    -- Contadores para Open Government
    eidas_total INTEGER := 0;
    eidas_compliant INTEGER := 0;
    govbr_total INTEGER := 0;
    govbr_compliant INTEGER := 0;
    
    -- Contadores globais para cada domínio
    openins_total INTEGER := 0;
    openins_compliant INTEGER := 0;
    openhealth_total INTEGER := 0;
    openhealth_compliant INTEGER := 0;
    opengov_total INTEGER := 0;
    opengov_compliant INTEGER := 0;
    
    -- Total global Open X
    openx_total INTEGER := 0;
    openx_compliant INTEGER := 0;
BEGIN
    -- Contagem para Open Insurance (Solvência II)
    SELECT COUNT(*), COUNT(*) FILTER (WHERE is_compliant = TRUE)
    INTO solv2_total, solv2_compliant
    FROM compliance_validators.validate_solvency_ii_compliance(tenant_id);
    
    -- Contagem para Open Insurance (SUSEP)
    SELECT COUNT(*), COUNT(*) FILTER (WHERE is_compliant = TRUE)
    INTO susep_total, susep_compliant
    FROM compliance_validators.validate_susep_compliance(tenant_id);
    
    -- Contagem para Open Health (Geral)
    SELECT COUNT(*), COUNT(*) FILTER (WHERE is_compliant = TRUE)
    INTO health_total, health_compliant
    FROM compliance_validators.validate_open_health_compliance(tenant_id);
    
    -- Contagem para Open Health (ANS)
    SELECT COUNT(*), COUNT(*) FILTER (WHERE is_compliant = TRUE)
    INTO ans_total, ans_compliant
    FROM compliance_validators.validate_ans_compliance(tenant_id);
    
    -- Contagem para Open Government (eIDAS)
    SELECT COUNT(*), COUNT(*) FILTER (WHERE is_compliant = TRUE)
    INTO eidas_total, eidas_compliant
    FROM compliance_validators.validate_eidas_gov_compliance(tenant_id);
    
    -- Contagem para Open Government (Gov BR)
    SELECT COUNT(*), COUNT(*) FILTER (WHERE is_compliant = TRUE)
    INTO govbr_total, govbr_compliant
    FROM compliance_validators.validate_gov_br_compliance(tenant_id);
    
    -- Calcular totais por domínio
    openins_total := solv2_total + susep_total;
    openins_compliant := solv2_compliant + susep_compliant;
    
    openhealth_total := health_total + ans_total;
    openhealth_compliant := health_compliant + ans_compliant;
    
    opengov_total := eidas_total + govbr_total;
    opengov_compliant := eidas_compliant + govbr_compliant;
    
    -- Calcular total global Open X
    openx_total := openins_total + openhealth_total + opengov_total;
    openx_compliant := openins_compliant + openhealth_compliant + opengov_compliant;
    
    -- Retornar resultados para Solvência II
    IF solv2_total > 0 THEN
        open_x_domain := 'OPEN_INSURANCE';
        framework := 'SOLVENCY_II';
        compliance_score := 4.0 * (solv2_compliant::NUMERIC / solv2_total);
        total_requirements := solv2_total;
        compliant_requirements := solv2_compliant;
        compliance_percentage := 100.0 * (solv2_compliant::NUMERIC / solv2_total);
        RETURN NEXT;
    END IF;
    
    -- Retornar resultados para SUSEP
    IF susep_total > 0 THEN
        open_x_domain := 'OPEN_INSURANCE';
        framework := 'SUSEP_BR';
        compliance_score := 4.0 * (susep_compliant::NUMERIC / susep_total);
        total_requirements := susep_total;
        compliant_requirements := susep_compliant;
        compliance_percentage := 100.0 * (susep_compliant::NUMERIC / susep_total);
        RETURN NEXT;
    END IF;
    
    -- Retornar resultados para Open Health (Geral)
    IF health_total > 0 THEN
        open_x_domain := 'OPEN_HEALTH';
        framework := 'HEALTH_GENERAL';
        compliance_score := 4.0 * (health_compliant::NUMERIC / health_total);
        total_requirements := health_total;
        compliant_requirements := health_compliant;
        compliance_percentage := 100.0 * (health_compliant::NUMERIC / health_total);
        RETURN NEXT;
    END IF;
    
    -- Retornar resultados para ANS
    IF ans_total > 0 THEN
        open_x_domain := 'OPEN_HEALTH';
        framework := 'ANS_BR';
        compliance_score := 4.0 * (ans_compliant::NUMERIC / ans_total);
        total_requirements := ans_total;
        compliant_requirements := ans_compliant;
        compliance_percentage := 100.0 * (ans_compliant::NUMERIC / ans_total);
        RETURN NEXT;
    END IF;
    
    -- Retornar resultados para eIDAS
    IF eidas_total > 0 THEN
        open_x_domain := 'OPEN_GOVERNMENT';
        framework := 'EIDAS';
        compliance_score := 4.0 * (eidas_compliant::NUMERIC / eidas_total);
        total_requirements := eidas_total;
        compliant_requirements := eidas_compliant;
        compliance_percentage := 100.0 * (eidas_compliant::NUMERIC / eidas_total);
        RETURN NEXT;
    END IF;
    
    -- Retornar resultados para Gov BR
    IF govbr_total > 0 THEN
        open_x_domain := 'OPEN_GOVERNMENT';
        framework := 'GOV_BR';
        compliance_score := 4.0 * (govbr_compliant::NUMERIC / govbr_total);
        total_requirements := govbr_total;
        compliant_requirements := govbr_compliant;
        compliance_percentage := 100.0 * (govbr_compliant::NUMERIC / govbr_total);
        RETURN NEXT;
    END IF;
    
    -- Retornar resultados para cada domínio Open X
    IF openins_total > 0 THEN
        open_x_domain := 'OPEN_INSURANCE';
        framework := 'OVERALL';
        compliance_score := 4.0 * (openins_compliant::NUMERIC / openins_total);
        total_requirements := openins_total;
        compliant_requirements := openins_compliant;
        compliance_percentage := 100.0 * (openins_compliant::NUMERIC / openins_total);
        RETURN NEXT;
    END IF;
    
    IF openhealth_total > 0 THEN
        open_x_domain := 'OPEN_HEALTH';
        framework := 'OVERALL';
        compliance_score := 4.0 * (openhealth_compliant::NUMERIC / openhealth_total);
        total_requirements := openhealth_total;
        compliant_requirements := openhealth_compliant;
        compliance_percentage := 100.0 * (openhealth_compliant::NUMERIC / openhealth_total);
        RETURN NEXT;
    END IF;
    
    IF opengov_total > 0 THEN
        open_x_domain := 'OPEN_GOVERNMENT';
        framework := 'OVERALL';
        compliance_score := 4.0 * (opengov_compliant::NUMERIC / opengov_total);
        total_requirements := opengov_total;
        compliant_requirements := opengov_compliant;
        compliance_percentage := 100.0 * (opengov_compliant::NUMERIC / opengov_total);
        RETURN NEXT;
    END IF;
    
    -- Retornar pontuação global para todo o ecossistema Open X
    IF openx_total > 0 THEN
        open_x_domain := 'OPEN_X';
        framework := 'OVERALL';
        compliance_score := 4.0 * (openx_compliant::NUMERIC / openx_total);
        total_requirements := openx_total;
        compliant_requirements := openx_compliant;
        compliance_percentage := 100.0 * (openx_compliant::NUMERIC / openx_total);
        RETURN NEXT;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Função para calcular o IRR (Índice de Risco Residual) para um domínio Open X
CREATE OR REPLACE FUNCTION compliance_validators.calculate_open_x_irr(
    tenant_id UUID,
    p_open_x_domain VARCHAR(50) DEFAULT NULL
) RETURNS TABLE (
    open_x_domain VARCHAR(50),
    irr VARCHAR(10),
    risk_level VARCHAR(20),
    compliance_percentage NUMERIC(5,2)
) AS $$
DECLARE
    compliance_perc NUMERIC(5,2);
    domain VARCHAR(50);
BEGIN
    -- Para cada domínio Open X relevante
    FOR domain, compliance_perc IN 
        SELECT s.open_x_domain, s.compliance_percentage 
        FROM compliance_validators.calculate_open_x_compliance_score(tenant_id) s
        WHERE s.framework = 'OVERALL'
        AND (p_open_x_domain IS NULL OR s.open_x_domain = p_open_x_domain)
    LOOP
        open_x_domain := domain;
        
        -- Determinar IRR baseado na % de conformidade
        IF compliance_perc >= 95 THEN
            irr := 'R1';
            risk_level := 'BAIXO';
        ELSIF compliance_perc >= 85 THEN
            irr := 'R2';
            risk_level := 'MODERADO';
        ELSIF compliance_perc >= 70 THEN
            irr := 'R3';
            risk_level := 'ALTO';
        ELSE
            irr := 'R4';
            risk_level := 'CRÍTICO';
        END IF;
        
        compliance_percentage := compliance_perc;
        RETURN NEXT;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Função para integrar as validações de Open X com o módulo econômico
CREATE OR REPLACE FUNCTION compliance_validators.register_open_x_economic_impact(
    tenant_id UUID,
    p_open_x_domain VARCHAR(50),
    p_jurisdiction VARCHAR(50),
    p_business_sector VARCHAR(100)
) RETURNS JSONB AS $$
DECLARE
    compliance_result RECORD;
    impact_result JSONB;
    non_compliant_count INTEGER := 0;
    total_economic_impact NUMERIC(15,2) := 0;
    detailed_impacts JSONB := '[]'::JSONB;
BEGIN
    -- Calcular impacto econômico para cada não-conformidade
    FOR compliance_result IN 
        SELECT r.open_x_domain, r.framework, r.requirement_id, r.requirement_name, r.is_compliant, r.irr_threshold
        FROM compliance_validators.generate_open_x_compliance_report(tenant_id) r
        WHERE (p_open_x_domain IS NULL OR r.open_x_domain = p_open_x_domain)
    LOOP
        -- Se não conforme, calcular impacto econômico
        IF NOT compliance_result.is_compliant THEN
            non_compliant_count := non_compliant_count + 1;
            
            -- Chamar função de impacto econômico existente
            -- Usando o IRR como nível de severidade
            impact_result := economic_planning.calculate_compliance_economic_impact(
                compliance_result.requirement_id,
                NULL, -- validation_id não disponível neste contexto
                p_jurisdiction,
                p_business_sector,
                compliance_result.irr_threshold, -- usando IRR como severidade
                tenant_id
            );
            
            -- Adicionar ao impacto total
            total_economic_impact := total_economic_impact + (impact_result->>'monetary_impact')::NUMERIC;
            
            -- Adicionar aos detalhes de impacto
            detailed_impacts := detailed_impacts || jsonb_build_object(
                'framework', compliance_result.framework,
                'requirement_id', compliance_result.requirement_id,
                'requirement_name', compliance_result.requirement_name,
                'irr', compliance_result.irr_threshold,
                'monetary_impact', (impact_result->>'monetary_impact')::NUMERIC,
                'impact_details', impact_result->>'impact_factors'
            );
        END IF;
    END LOOP;
    
    -- Construir resultado final
    RETURN jsonb_build_object(
        'open_x_domain', p_open_x_domain,
        'jurisdiction', p_jurisdiction, 
        'business_sector', p_business_sector,
        'total_non_compliant', non_compliant_count,
        'total_economic_impact', total_economic_impact,
        'currency', 'EUR',
        'detailed_impacts', detailed_impacts,
        'analysis_timestamp', CURRENT_TIMESTAMP
    );
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Comentários de Documentação
-- =====================================================================

COMMENT ON SCHEMA compliance_validators IS 'Esquema para validadores de conformidade regulatória do INNOVABIZ';

-- Open Insurance
COMMENT ON TABLE compliance_validators.solvency_ii_requirements IS 'Requisitos de conformidade Solvência II para IAM em Open Insurance';
COMMENT ON TABLE compliance_validators.susep_requirements IS 'Requisitos de conformidade SUSEP para Open Insurance no Brasil';
COMMENT ON FUNCTION compliance_validators.validate_solvency_ii_compliance IS 'Valida conformidade com Solvência II para um tenant específico';
COMMENT ON FUNCTION compliance_validators.validate_susep_compliance IS 'Valida conformidade com requisitos SUSEP para um tenant específico';
COMMENT ON FUNCTION compliance_validators.generate_open_insurance_compliance_report IS 'Gera relatório consolidado de conformidade para Open Insurance';

-- Open Health
COMMENT ON TABLE compliance_validators.open_health_requirements IS 'Requisitos de conformidade HIPAA/GDPR para IAM em Open Health';
COMMENT ON TABLE compliance_validators.ans_requirements IS 'Requisitos de conformidade ANS para Open Health no Brasil';
COMMENT ON FUNCTION compliance_validators.validate_open_health_compliance IS 'Valida conformidade com requisitos gerais de Open Health';
COMMENT ON FUNCTION compliance_validators.validate_ans_compliance IS 'Valida conformidade com requisitos ANS para um tenant específico';
COMMENT ON FUNCTION compliance_validators.generate_open_health_compliance_report IS 'Gera relatório consolidado de conformidade para Open Health';

-- Open Government
COMMENT ON TABLE compliance_validators.eidas_gov_requirements IS 'Requisitos de conformidade eIDAS para IAM em Open Government';
COMMENT ON TABLE compliance_validators.gov_br_requirements IS 'Requisitos de conformidade para Governo Digital no Brasil';
COMMENT ON FUNCTION compliance_validators.validate_eidas_gov_compliance IS 'Valida conformidade com eIDAS para um tenant específico';
COMMENT ON FUNCTION compliance_validators.validate_gov_br_compliance IS 'Valida conformidade com Gov.br para um tenant específico';
COMMENT ON FUNCTION compliance_validators.generate_open_government_compliance_report IS 'Gera relatório consolidado de conformidade para Open Government';

-- Open X (Consolidação)
COMMENT ON FUNCTION compliance_validators.generate_open_x_compliance_report IS 'Gera relatório consolidado de conformidade para todo o ecossistema Open X';
COMMENT ON FUNCTION compliance_validators.calculate_open_x_compliance_score IS 'Calcula pontuação de conformidade para todos os domínios Open X';
COMMENT ON FUNCTION compliance_validators.calculate_open_x_irr IS 'Calcula o Índice de Risco Residual para domínios Open X';
COMMENT ON FUNCTION compliance_validators.register_open_x_economic_impact IS 'Registra e calcula o impacto econômico de não-conformidades em Open X';

-- =====================================================================
-- Fim do Script
-- =====================================================================
