-- =====================================================================
-- INNOVABIZ - Validadores de Conformidade para Setor de Saúde
-- Versão: 1.0.0
-- Data de Criação: 14/05/2025
-- Autor: INNOVABIZ DevOps
-- Descrição: Funções e procedimentos para validação de conformidade IAM
--            específicos para o setor de saúde (HIPAA, GDPR, LGPD)
-- Regiões Suportadas: UE/Portugal, Brasil, EUA
-- =====================================================================

-- Criação de schema para validadores de conformidade
CREATE SCHEMA IF NOT EXISTS compliance_validators;

-- =====================================================================
-- Tabelas de Referência para Requisitos Regulatórios
-- =====================================================================

-- Requisitos de HIPAA para IAM em saúde
CREATE TABLE compliance_validators.hipaa_requirements (
    requirement_id VARCHAR(20) PRIMARY KEY,
    requirement_name VARCHAR(255) NOT NULL,
    requirement_description TEXT NOT NULL,
    validation_query TEXT,
    implementation_details TEXT,
    required_auth_level VARCHAR(50),
    required_auth_methods VARCHAR(50)[],
    irr_threshold VARCHAR(10),
    is_mandatory BOOLEAN NOT NULL DEFAULT TRUE,
    applies_to_roles VARCHAR(50)[],
    relevant_section VARCHAR(50),
    reference_url TEXT
);

-- Requisitos de GDPR para IAM em saúde
CREATE TABLE compliance_validators.gdpr_requirements (
    requirement_id VARCHAR(20) PRIMARY KEY,
    requirement_name VARCHAR(255) NOT NULL,
    requirement_description TEXT NOT NULL,
    validation_query TEXT,
    implementation_details TEXT,
    required_auth_level VARCHAR(50),
    required_auth_methods VARCHAR(50)[],
    irr_threshold VARCHAR(10),
    is_mandatory BOOLEAN NOT NULL DEFAULT TRUE,
    applies_to_data_types VARCHAR(50)[],
    relevant_articles VARCHAR(50)[],
    reference_url TEXT
);

-- Requisitos de LGPD para IAM em saúde
CREATE TABLE compliance_validators.lgpd_requirements (
    requirement_id VARCHAR(20) PRIMARY KEY,
    requirement_name VARCHAR(255) NOT NULL,
    requirement_description TEXT NOT NULL,
    validation_query TEXT,
    implementation_details TEXT,
    required_auth_level VARCHAR(50),
    required_auth_methods VARCHAR(50)[],
    irr_threshold VARCHAR(10),
    is_mandatory BOOLEAN NOT NULL DEFAULT TRUE,
    applies_to_data_types VARCHAR(50)[],
    relevant_articles VARCHAR(50)[],
    reference_url TEXT
);

-- =====================================================================
-- Inserção de dados básicos para Requisitos HIPAA
-- =====================================================================

INSERT INTO compliance_validators.hipaa_requirements (
    requirement_id, requirement_name, requirement_description, 
    validation_query, implementation_details, 
    required_auth_level, required_auth_methods, irr_threshold,
    relevant_section, reference_url
) VALUES
('HIPAA-IAM-01', 'Autenticação Única', 
 'Cada usuário deve ser identificado de forma única no sistema',
 'SELECT COUNT(*) = 0 FROM iam_core.users WHERE tenant_id = $1 AND user_type = ''HUMAN'' AND external_id IS NULL',
 'Verificar se todos os usuários possuem identificação única',
 'INTERMEDIATE', ARRAY['KB-01-02', 'PB-03-01'], 'R3',
 '164.312(a)(2)(i)', 'https://www.hhs.gov/hipaa/for-professionals/security/index.html'),

('HIPAA-IAM-02', 'Controle de Terminação de Sessão', 
 'O sistema deve encerrar sessões eletrônicas após um período predeterminado de inatividade',
 'SELECT configurations::jsonb ? ''session_timeout_seconds'' FROM iam_core.tenants WHERE tenant_id = $1',
 'Verificar configuração de timeout de sessão',
 'INTERMEDIATE', NULL, 'R3',
 '164.312(a)(2)(iii)', 'https://www.hhs.gov/hipaa/for-professionals/security/index.html'),

('HIPAA-IAM-03', 'Autenticação de Emergência', 
 'Procedimentos para obtenção de acesso necessário às PHI durante uma emergência',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies WHERE tenant_id = $1 AND policy_rules::jsonb ? ''emergency_access''',
 'Verificar se existem políticas de acesso de emergência',
 'ADVANCED', ARRAY['HC-01-01', 'HC-01-03'], 'R2',
 '164.312(a)(2)(ii)', 'https://www.hhs.gov/hipaa/for-professionals/security/index.html'),

('HIPAA-IAM-04', 'Log de Atividades', 
 'Registros de atividade para monitoramento de acesso a PHI',
 'SELECT COUNT(*) > 0 FROM information_schema.tables WHERE table_schema = ''iam_core'' AND table_name = ''authentication_history''',
 'Verificar existência de tabelas de histórico de autenticação',
 'ADVANCED', NULL, 'R2',
 '164.308(a)(1)(ii)(D)', 'https://www.hhs.gov/hipaa/for-professionals/security/index.html'),

('HIPAA-IAM-05', 'Autenticação Multi-Fator', 
 'Implementação de MFA para acesso a sistemas contendo PHI',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies WHERE tenant_id = $1 AND policy_type = ''MFA'' AND is_enabled = TRUE',
 'Verificar se políticas MFA estão habilitadas',
 'ADVANCED', ARRAY['PB-01-01', 'PB-01-03', 'HC-01-01'], 'R2',
 '164.312(a)(1)', 'https://www.hhs.gov/hipaa/for-professionals/security/index.html');

-- =====================================================================
-- Inserção de dados básicos para Requisitos GDPR
-- =====================================================================

INSERT INTO compliance_validators.gdpr_requirements (
    requirement_id, requirement_name, requirement_description, 
    validation_query, implementation_details, 
    required_auth_level, required_auth_methods, irr_threshold,
    relevant_articles, reference_url
) VALUES
('GDPR-IAM-01', 'Autenticação Robusta', 
 'Implementação de autenticação robusta para acesso a dados pessoais sensíveis de saúde',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies WHERE tenant_id = $1 AND policy_rules::jsonb @> ''{"minimum_factor_strength": "ADVANCED"}''',
 'Verificar políticas com força de autenticação avançada',
 'ADVANCED', ARRAY['KB-01-02', 'PB-01-02', 'PB-01-03'], 'R2',
 ARRAY['Art. 32'], 'https://gdpr-info.eu/art-32-gdpr/'),

('GDPR-IAM-02', 'Controle de Acesso a Dados Sensíveis', 
 'Implementação de controles de acesso baseados em função para dados de saúde',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies WHERE tenant_id = $1 AND applies_to_security_profiles::text LIKE ''%HEALTHCARE%''',
 'Verificar políticas específicas para perfis de saúde',
 'ADVANCED', ARRAY['HC-01-01', 'HC-01-02'], 'R2',
 ARRAY['Art. 9, Art. 32'], 'https://gdpr-info.eu/art-9-gdpr/'),

('GDPR-IAM-03', 'Revogação de Acesso', 
 'Procedimentos para revogação imediata de acesso quando não mais necessário',
 'SELECT COUNT(*) > 0 FROM iam_core.sessions WHERE user_id = $1 AND invalidated_at IS NOT NULL',
 'Verificar funcionalidade de invalidação de sessão',
 'INTERMEDIATE', NULL, 'R3',
 ARRAY['Art. 32(1)(b)'], 'https://gdpr-info.eu/art-32-gdpr/'),

('GDPR-IAM-04', 'Auditoria de Acesso a Dados de Saúde', 
 'Capacidade de auditar quem acessou dados pessoais sensíveis de saúde',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_history WHERE user_id = $1 AND event_details::jsonb ? ''data_access''',
 'Verificar registros detalhados de acesso a dados',
 'ADVANCED', NULL, 'R2',
 ARRAY['Art. 30, Art. 32'], 'https://gdpr-info.eu/art-30-gdpr/'),

('GDPR-IAM-05', 'Proteção Específica para Dados de Saúde', 
 'Medidas específicas para proteger dados relativos à saúde como categoria especial',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations WHERE tenant_id = $1 AND method_id = ''HC-01-02''',
 'Verificar configurações específicas para acesso a dados sensíveis de saúde',
 'VERY_ADVANCED', ARRAY['HC-01-01', 'HC-01-02', 'PB-02-02'], 'R1',
 ARRAY['Art. 9(2)(h), Art. 9(3)'], 'https://gdpr-info.eu/art-9-gdpr/');

-- =====================================================================
-- Inserção de dados básicos para Requisitos LGPD
-- =====================================================================

INSERT INTO compliance_validators.lgpd_requirements (
    requirement_id, requirement_name, requirement_description, 
    validation_query, implementation_details, 
    required_auth_level, required_auth_methods, irr_threshold,
    relevant_articles, reference_url
) VALUES
('LGPD-IAM-01', 'Autenticação para Dados de Saúde', 
 'Implementação de medidas específicas para dados de saúde como dados sensíveis',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies WHERE tenant_id = $1 AND applies_to_security_profiles::text LIKE ''%HEALTHCARE%'' AND is_enabled = TRUE',
 'Verificar políticas habilitadas para perfis de saúde',
 'ADVANCED', ARRAY['HC-01-01', 'HC-01-02'], 'R2',
 ARRAY['Art. 11, Art. 46'], 'https://www.planalto.gov.br/ccivil_03/_ato2015-2018/2018/lei/l13709.htm'),

('LGPD-IAM-02', 'Registro de Acesso', 
 'Manutenção de registros de acesso a dados pessoais sensíveis',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_history WHERE tenant_id = $1 AND timestamp > now() - interval ''90 days''',
 'Verificar histórico de autenticação recente',
 'INTERMEDIATE', NULL, 'R3',
 ARRAY['Art. 46, Art. 48'], 'https://www.planalto.gov.br/ccivil_03/_ato2015-2018/2018/lei/l13709.htm'),

('LGPD-IAM-03', 'Autenticação Multi-Fator', 
 'Utilização de múltiplos fatores para autenticação em sistemas de saúde',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies WHERE tenant_id = $1 AND policy_type = ''MFA'' AND is_enabled = TRUE',
 'Verificar políticas MFA habilitadas',
 'ADVANCED', ARRAY['KB-01-02', 'PB-01-03', 'HC-01-03'], 'R2',
 ARRAY['Art. 46'], 'https://www.planalto.gov.br/ccivil_03/_ato2015-2018/2018/lei/l13709.htm'),

('LGPD-IAM-04', 'Controle Granular de Acesso', 
 'Controles de acesso específicos para diferentes tipos de dados de saúde',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations WHERE tenant_id = $1 AND config_parameters::jsonb ? ''professional_db_integration''',
 'Verificar integração com bases de dados profissionais',
 'ADVANCED', ARRAY['HC-01-01', 'HC-01-02'], 'R2',
 ARRAY['Art. 46, Art. 47'], 'https://www.planalto.gov.br/ccivil_03/_ato2015-2018/2018/lei/l13709.htm'),

('LGPD-IAM-05', 'Renovação de Autenticação', 
 'Procedimentos para renovação periódica de credenciais de autenticação',
 'SELECT configurations::jsonb->>''password_policy'' ? ''max_age_days'' FROM iam_core.tenants WHERE tenant_id = $1',
 'Verificar política de expiração de senhas',
 'INTERMEDIATE', NULL, 'R3',
 ARRAY['Art. 46'], 'https://www.planalto.gov.br/ccivil_03/_ato2015-2018/2018/lei/l13709.htm');

-- =====================================================================
-- Funções de Validação para Conformidade em Saúde
-- =====================================================================

-- Função para validar conformidade HIPAA
CREATE OR REPLACE FUNCTION compliance_validators.validate_hipaa_compliance(
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
    FOR req IN SELECT * FROM compliance_validators.hipaa_requirements LOOP
        -- Preparar consulta de validação
        exec_query := 'SELECT (' || req.validation_query || ') AS result';
        
        -- Executar consulta de validação
        BEGIN
            EXECUTE exec_query USING tenant_id INTO query_result;
            
            -- Retornar resultado
            requirement_id := req.requirement_id;
            requirement_name := req.requirement_name;
            is_compliant := query_result;
            details := CASE 
                WHEN query_result THEN 'Compliant: ' || req.implementation_details
                ELSE 'Non-compliant: ' || req.implementation_details
            END;
            RETURN NEXT;
        EXCEPTION WHEN OTHERS THEN
            requirement_id := req.requirement_id;
            requirement_name := req.requirement_name;
            is_compliant := FALSE;
            details := 'Error validating: ' || SQLERRM;
            RETURN NEXT;
        END;
    END LOOP;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para validar conformidade GDPR
CREATE OR REPLACE FUNCTION compliance_validators.validate_gdpr_compliance(
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
    FOR req IN SELECT * FROM compliance_validators.gdpr_requirements LOOP
        -- Preparar consulta de validação
        exec_query := 'SELECT (' || req.validation_query || ') AS result';
        
        -- Executar consulta de validação
        BEGIN
            EXECUTE exec_query USING tenant_id INTO query_result;
            
            -- Retornar resultado
            requirement_id := req.requirement_id;
            requirement_name := req.requirement_name;
            is_compliant := query_result;
            details := CASE 
                WHEN query_result THEN 'Compliant: ' || req.implementation_details
                ELSE 'Non-compliant: ' || req.implementation_details
            END;
            RETURN NEXT;
        EXCEPTION WHEN OTHERS THEN
            requirement_id := req.requirement_id;
            requirement_name := req.requirement_name;
            is_compliant := FALSE;
            details := 'Error validating: ' || SQLERRM;
            RETURN NEXT;
        END;
    END LOOP;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para validar conformidade LGPD
CREATE OR REPLACE FUNCTION compliance_validators.validate_lgpd_compliance(
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
    FOR req IN SELECT * FROM compliance_validators.lgpd_requirements LOOP
        -- Preparar consulta de validação
        exec_query := 'SELECT (' || req.validation_query || ') AS result';
        
        -- Executar consulta de validação
        BEGIN
            EXECUTE exec_query USING tenant_id INTO query_result;
            
            -- Retornar resultado
            requirement_id := req.requirement_id;
            requirement_name := req.requirement_name;
            is_compliant := query_result;
            details := CASE 
                WHEN query_result THEN 'Compliant: ' || req.implementation_details
                ELSE 'Non-compliant: ' || req.implementation_details
            END;
            RETURN NEXT;
        EXCEPTION WHEN OTHERS THEN
            requirement_id := req.requirement_id;
            requirement_name := req.requirement_name;
            is_compliant := FALSE;
            details := 'Error validating: ' || SQLERRM;
            RETURN NEXT;
        END;
    END LOOP;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para gerar relatório consolidado multi-framework
CREATE OR REPLACE FUNCTION compliance_validators.generate_healthcare_compliance_report(
    tenant_id UUID,
    include_hipaa BOOLEAN DEFAULT TRUE,
    include_gdpr BOOLEAN DEFAULT TRUE,
    include_lgpd BOOLEAN DEFAULT TRUE
) RETURNS TABLE (
    framework VARCHAR(10),
    requirement_id VARCHAR(20),
    requirement_name VARCHAR(255),
    is_compliant BOOLEAN,
    details TEXT
) AS $$
BEGIN
    -- Validar HIPAA
    IF include_hipaa THEN
        RETURN QUERY 
        SELECT 'HIPAA'::VARCHAR(10) as framework, * FROM compliance_validators.validate_hipaa_compliance(tenant_id);
    END IF;
    
    -- Validar GDPR
    IF include_gdpr THEN
        RETURN QUERY 
        SELECT 'GDPR'::VARCHAR(10) as framework, * FROM compliance_validators.validate_gdpr_compliance(tenant_id);
    END IF;
    
    -- Validar LGPD
    IF include_lgpd THEN
        RETURN QUERY 
        SELECT 'LGPD'::VARCHAR(10) as framework, * FROM compliance_validators.validate_lgpd_compliance(tenant_id);
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Função para gerar pontuação de conformidade
CREATE OR REPLACE FUNCTION compliance_validators.calculate_healthcare_compliance_score(
    tenant_id UUID
) RETURNS TABLE (
    framework VARCHAR(10),
    compliance_score NUMERIC(5,2),
    total_requirements INTEGER,
    compliant_requirements INTEGER,
    compliance_percentage NUMERIC(5,2)
) AS $$
DECLARE
    hipaa_total INTEGER := 0;
    hipaa_compliant INTEGER := 0;
    gdpr_total INTEGER := 0;
    gdpr_compliant INTEGER := 0;
    lgpd_total INTEGER := 0;
    lgpd_compliant INTEGER := 0;
    report_row RECORD;
BEGIN
    -- Executar relatório completo
    FOR report_row IN SELECT * FROM compliance_validators.generate_healthcare_compliance_report(tenant_id) LOOP
        -- Contar requisitos por framework
        CASE report_row.framework
            WHEN 'HIPAA' THEN
                hipaa_total := hipaa_total + 1;
                IF report_row.is_compliant THEN
                    hipaa_compliant := hipaa_compliant + 1;
                END IF;
            WHEN 'GDPR' THEN
                gdpr_total := gdpr_total + 1;
                IF report_row.is_compliant THEN
                    gdpr_compliant := gdpr_compliant + 1;
                END IF;
            WHEN 'LGPD' THEN
                lgpd_total := lgpd_total + 1;
                IF report_row.is_compliant THEN
                    lgpd_compliant := lgpd_compliant + 1;
                END IF;
        END CASE;
    END LOOP;
    
    -- Retornar resultados HIPAA
    IF hipaa_total > 0 THEN
        framework := 'HIPAA';
        compliance_score := 4.0 * (hipaa_compliant::NUMERIC / hipaa_total);
        total_requirements := hipaa_total;
        compliant_requirements := hipaa_compliant;
        compliance_percentage := 100.0 * (hipaa_compliant::NUMERIC / hipaa_total);
        RETURN NEXT;
    END IF;
    
    -- Retornar resultados GDPR
    IF gdpr_total > 0 THEN
        framework := 'GDPR';
        compliance_score := 4.0 * (gdpr_compliant::NUMERIC / gdpr_total);
        total_requirements := gdpr_total;
        compliant_requirements := gdpr_compliant;
        compliance_percentage := 100.0 * (gdpr_compliant::NUMERIC / gdpr_total);
        RETURN NEXT;
    END IF;
    
    -- Retornar resultados LGPD
    IF lgpd_total > 0 THEN
        framework := 'LGPD';
        compliance_score := 4.0 * (lgpd_compliant::NUMERIC / lgpd_total);
        total_requirements := lgpd_total;
        compliant_requirements := lgpd_compliant;
        compliance_percentage := 100.0 * (lgpd_compliant::NUMERIC / lgpd_total);
        RETURN NEXT;
    END IF;
    
    -- Retornar pontuação geral
    framework := 'OVERALL';
    total_requirements := hipaa_total + gdpr_total + lgpd_total;
    compliant_requirements := hipaa_compliant + gdpr_compliant + lgpd_compliant;
    
    IF total_requirements > 0 THEN
        compliance_score := 4.0 * (compliant_requirements::NUMERIC / total_requirements);
        compliance_percentage := 100.0 * (compliant_requirements::NUMERIC / total_requirements);
    ELSE
        compliance_score := 0;
        compliance_percentage := 0;
    END IF;
    
    RETURN NEXT;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Comentários de Documentação
-- =====================================================================

COMMENT ON SCHEMA compliance_validators IS 'Esquema para validadores de conformidade regulatória do INNOVABIZ';
COMMENT ON TABLE compliance_validators.hipaa_requirements IS 'Requisitos de conformidade HIPAA para IAM em saúde';
COMMENT ON TABLE compliance_validators.gdpr_requirements IS 'Requisitos de conformidade GDPR para IAM em saúde';
COMMENT ON TABLE compliance_validators.lgpd_requirements IS 'Requisitos de conformidade LGPD para IAM em saúde';
COMMENT ON FUNCTION compliance_validators.validate_hipaa_compliance IS 'Valida conformidade HIPAA para um tenant específico';
COMMENT ON FUNCTION compliance_validators.validate_gdpr_compliance IS 'Valida conformidade GDPR para um tenant específico';
COMMENT ON FUNCTION compliance_validators.validate_lgpd_compliance IS 'Valida conformidade LGPD para um tenant específico';
COMMENT ON FUNCTION compliance_validators.generate_healthcare_compliance_report IS 'Gera relatório consolidado de conformidade para saúde';
COMMENT ON FUNCTION compliance_validators.calculate_healthcare_compliance_score IS 'Calcula pontuação de conformidade para frameworks de saúde';

-- =====================================================================
-- Fim do Script
-- =====================================================================
