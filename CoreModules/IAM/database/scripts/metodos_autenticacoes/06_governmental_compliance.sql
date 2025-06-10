-- =====================================================================
-- INNOVABIZ - Validadores de Conformidade para Setor Governamental
-- Versão: 1.0.0
-- Data de Criação: 14/05/2025
-- Autor: INNOVABIZ DevOps
-- Descrição: Funções e procedimentos para validação de conformidade IAM
--            específicos para o setor governamental e serviços públicos
-- Regiões Suportadas: UE/Portugal, Brasil, Angola
-- =====================================================================

-- Criação de schema para validadores de conformidade se ainda não existir
CREATE SCHEMA IF NOT EXISTS compliance_validators;

-- =====================================================================
-- Tabelas de Referência para Requisitos Regulatórios
-- =====================================================================

-- Requisitos para IAM em serviços governamentais na UE (inclui eIDAS)
CREATE TABLE compliance_validators.eidas_requirements (
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
    relevant_articles VARCHAR(50)[],
    reference_url TEXT
);

-- Requisitos para IAM em serviços governamentais no Brasil (inclui ICP-Brasil)
CREATE TABLE compliance_validators.icp_brasil_requirements (
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
    relevant_standards VARCHAR(50)[],
    reference_url TEXT
);

-- Requisitos para IAM em serviços governamentais em Angola
CREATE TABLE compliance_validators.angola_gov_requirements (
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
    relevant_legislation VARCHAR(50)[],
    reference_url TEXT
);

-- =====================================================================
-- Inserção de dados básicos para Requisitos eIDAS (UE)
-- =====================================================================

INSERT INTO compliance_validators.eidas_requirements (
    requirement_id, requirement_name, requirement_description, 
    validation_query, implementation_details, 
    required_auth_level, required_auth_methods, irr_threshold,
    relevant_articles, reference_url
) VALUES
('EIDAS-IAM-01', 'Níveis de Garantia de Identidade', 
 'Implementação dos níveis de garantia de identidade (Baixo, Substancial, Elevado) conforme eIDAS',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_methods 
  WHERE security_level IN (''BASIC'', ''ADVANCED'', ''VERY_ADVANCED'') 
  AND method_id LIKE ''GP-%''',
 'Verificar se há métodos específicos para diferentes níveis de garantia',
 'ADVANCED', ARRAY['GP-01-01', 'GP-01-02'], 'R2',
 ARRAY['Art. 8'], 'https://eur-lex.europa.eu/legal-content/PT/TXT/?uri=CELEX:32014R0910'),

('EIDAS-IAM-02', 'Assinaturas Qualificadas', 
 'Suporte para assinaturas eletrônicas qualificadas conforme definido no regulamento',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_methods 
  WHERE method_id = ''GP-01-02'' 
  AND security_level = ''VERY_ADVANCED''',
 'Verificar suporte para assinaturas digitais governamentais de alto nível',
 'VERY_ADVANCED', ARRAY['GP-01-02', 'PB-02-02'], 'R1',
 ARRAY['Art. 25-34'], 'https://eur-lex.europa.eu/legal-content/PT/TXT/?uri=CELEX:32014R0910'),

('EIDAS-IAM-03', 'Reconhecimento Transfronteiriço', 
 'Capacidade de reconhecer e aceitar identificação eletrônica de outros estados-membros',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND config_parameters::jsonb ? ''cross_border_recognition''',
 'Verificar configurações de reconhecimento transfronteiriço',
 'ADVANCED', ARRAY['GP-01-01'], 'R2',
 ARRAY['Art. 6'], 'https://eur-lex.europa.eu/legal-content/PT/TXT/?uri=CELEX:32014R0910'),

('EIDAS-IAM-04', 'Autenticação Multifatorial Adaptativa', 
 'Implementação de autenticação multifatorial adaptativa baseada no nível de garantia requerido',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies 
  WHERE tenant_id = $1 
  AND policy_type = ''ADAPTIVE'' 
  AND applies_to_security_profiles::text LIKE ''%GOVERNMENT%''',
 'Verificar políticas adaptativas para perfis governamentais',
 'ADVANCED', NULL, 'R2',
 ARRAY['Art. 8(2)'], 'https://eur-lex.europa.eu/legal-content/PT/TXT/?uri=CELEX:32014R0910'),

('EIDAS-IAM-05', 'Verificação de Identidade', 
 'Procedimentos para verificação de identidade conforme nível de garantia',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND method_id = ''GP-01-01'' 
  AND config_parameters::jsonb ? ''identity_verification_procedure''',
 'Verificar procedimentos de verificação de identidade',
 'VERY_ADVANCED', ARRAY['GP-01-01'], 'R1',
 ARRAY['Art. 8(3)'], 'https://eur-lex.europa.eu/legal-content/PT/TXT/?uri=CELEX:32014R0910');

-- =====================================================================
-- Inserção de dados básicos para Requisitos ICP-Brasil
-- =====================================================================

INSERT INTO compliance_validators.icp_brasil_requirements (
    requirement_id, requirement_name, requirement_description, 
    validation_query, implementation_details, 
    required_auth_level, required_auth_methods, irr_threshold,
    relevant_standards, reference_url
) VALUES
('ICP-IAM-01', 'Certificados ICP-Brasil', 
 'Suporte para certificados digitais emitidos pela hierarquia ICP-Brasil',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_methods 
  WHERE method_id = ''GP-01-02'' 
  AND implementation_status = ''IMPLEMENTED''',
 'Verificar implementação de suporte a certificados ICP-Brasil',
 'VERY_ADVANCED', ARRAY['GP-01-02', 'PB-02-02'], 'R1',
 ARRAY['DOC-ICP-04'], 'https://www.gov.br/iti/pt-br'),

('ICP-IAM-02', 'Assinatura com Certificados A3/A4', 
 'Suporte para assinaturas com certificados de tipo A3 ou A4 (hardware)',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND method_id = ''PB-02-02'' 
  AND config_parameters::jsonb ? ''icp_brasil_a3_a4''',
 'Verificar configurações para certificados A3/A4',
 'VERY_ADVANCED', ARRAY['PB-02-02'], 'R1',
 ARRAY['DOC-ICP-04'], 'https://www.gov.br/iti/pt-br'),

('ICP-IAM-03', 'Login Único Gov.br', 
 'Integração com o sistema de Login Único do governo federal brasileiro',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND config_parameters::jsonb ? ''govbr_integration''',
 'Verificar configurações de integração com Gov.br',
 'ADVANCED', ARRAY['GP-01-01'], 'R2',
 ARRAY['Decreto 8.936/2016'], 'https://www.gov.br/governodigital/pt-br'),

('ICP-IAM-04', 'Validação de Certificados Revogados', 
 'Mecanismo de verificação de revogação de certificados (CRL/OCSP)',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND method_id = ''PB-02-01'' 
  AND config_parameters::jsonb ? ''revocation_check''',
 'Verificar configurações de checagem de revogação',
 'ADVANCED', NULL, 'R2',
 ARRAY['DOC-ICP-05'], 'https://www.gov.br/iti/pt-br'),

('ICP-IAM-05', 'Níveis de Autenticação Gov.br', 
 'Suporte aos níveis de autenticação do Gov.br (Bronze, Prata, Ouro)',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies 
  WHERE tenant_id = $1 
  AND policy_rules::jsonb ? ''govbr_levels''',
 'Verificar políticas para diferentes níveis Gov.br',
 'ADVANCED', ARRAY['GP-01-01'], 'R2',
 ARRAY['Portaria SGD/ME Nº 2.154/2021'], 'https://www.gov.br/governodigital/pt-br');

-- =====================================================================
-- Inserção de dados básicos para Requisitos Angola
-- =====================================================================

INSERT INTO compliance_validators.angola_gov_requirements (
    requirement_id, requirement_name, requirement_description, 
    validation_query, implementation_details, 
    required_auth_level, required_auth_methods, irr_threshold,
    relevant_legislation, reference_url
) VALUES
('ANG-IAM-01', 'Autenticação para Serviços Públicos Digitais', 
 'Implementação de mecanismos de autenticação para serviços de governo eletrônico',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_methods 
  WHERE method_id LIKE ''GP-%'' 
  AND security_level IN (''ADVANCED'', ''VERY_ADVANCED'')',
 'Verificar métodos avançados para serviços governamentais',
 'ADVANCED', ARRAY['GP-01-01', 'GP-01-02'], 'R2',
 ARRAY['Lei 7/17'], 'https://governo.gov.ao/'),

('ANG-IAM-02', 'Assinatura Eletrônica Qualificada', 
 'Suporte para assinaturas eletrônicas com valor legal em Angola',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND method_id = ''GP-01-02'' 
  AND config_parameters::jsonb ? ''angola_qualified_signature''',
 'Verificar configurações para assinaturas qualificadas angolanas',
 'VERY_ADVANCED', ARRAY['GP-01-02'], 'R1',
 ARRAY['Decreto Presidencial 202/19'], 'https://governo.gov.ao/'),

('ANG-IAM-03', 'Documento de Identidade Nacional', 
 'Integração com o sistema de verificação do Bilhete de Identidade',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND method_id = ''GP-01-01'' 
  AND config_parameters::jsonb ? ''angola_id_verification''',
 'Verificar configurações de verificação de identidade angolana',
 'ADVANCED', ARRAY['GP-01-01'], 'R2',
 ARRAY['Lei 1/17'], 'https://governo.gov.ao/'),

('ANG-IAM-04', 'Níveis de Segurança para Serviços Públicos', 
 'Implementação de diferentes níveis de segurança conforme criticidade do serviço',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies 
  WHERE tenant_id = $1 
  AND policy_type = ''CONDITIONAL'' 
  AND policy_rules::jsonb ? ''service_criticality''',
 'Verificar políticas baseadas em criticidade de serviço',
 'ADVANCED', NULL, 'R2',
 ARRAY['PDSI 2019-2022'], 'https://governo.gov.ao/'),

('ANG-IAM-05', 'Proteção de Dados Pessoais', 
 'Conformidade com requisitos de proteção de dados pessoais em Angola',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies 
  WHERE tenant_id = $1 
  AND policy_rules::jsonb ? ''data_protection'' 
  AND applies_to_regions::text LIKE ''%AO%''',
 'Verificar políticas de proteção de dados específicas para Angola',
 'ADVANCED', ARRAY['GP-01-01'], 'R2',
 ARRAY['Lei 22/11'], 'https://governo.gov.ao/');

-- =====================================================================
-- Funções de Validação para Conformidade Governamental
-- =====================================================================

-- Função para validar conformidade eIDAS
CREATE OR REPLACE FUNCTION compliance_validators.validate_eidas_compliance(
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
    FOR req IN SELECT * FROM compliance_validators.eidas_requirements LOOP
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

-- Função para validar conformidade ICP-Brasil
CREATE OR REPLACE FUNCTION compliance_validators.validate_icp_brasil_compliance(
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
    FOR req IN SELECT * FROM compliance_validators.icp_brasil_requirements LOOP
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

-- Função para validar conformidade Angola
CREATE OR REPLACE FUNCTION compliance_validators.validate_angola_gov_compliance(
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
    FOR req IN SELECT * FROM compliance_validators.angola_gov_requirements LOOP
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

-- Função para gerar relatório consolidado multi-framework para Governo
CREATE OR REPLACE FUNCTION compliance_validators.generate_government_compliance_report(
    tenant_id UUID,
    include_eidas BOOLEAN DEFAULT TRUE,
    include_icp_brasil BOOLEAN DEFAULT TRUE,
    include_angola BOOLEAN DEFAULT TRUE
) RETURNS TABLE (
    framework VARCHAR(15),
    requirement_id VARCHAR(20),
    requirement_name VARCHAR(255),
    is_compliant BOOLEAN,
    details TEXT
) AS $$
BEGIN
    -- Validar eIDAS
    IF include_eidas THEN
        RETURN QUERY 
        SELECT 'eIDAS'::VARCHAR(15) as framework, * FROM compliance_validators.validate_eidas_compliance(tenant_id);
    END IF;
    
    -- Validar ICP-Brasil
    IF include_icp_brasil THEN
        RETURN QUERY 
        SELECT 'ICP-Brasil'::VARCHAR(15) as framework, * FROM compliance_validators.validate_icp_brasil_compliance(tenant_id);
    END IF;
    
    -- Validar Angola
    IF include_angola THEN
        RETURN QUERY 
        SELECT 'Angola'::VARCHAR(15) as framework, * FROM compliance_validators.validate_angola_gov_compliance(tenant_id);
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Função para gerar pontuação de conformidade para Governo
CREATE OR REPLACE FUNCTION compliance_validators.calculate_government_compliance_score(
    tenant_id UUID
) RETURNS TABLE (
    framework VARCHAR(15),
    compliance_score NUMERIC(5,2),
    total_requirements INTEGER,
    compliant_requirements INTEGER,
    compliance_percentage NUMERIC(5,2)
) AS $$
DECLARE
    eidas_total INTEGER := 0;
    eidas_compliant INTEGER := 0;
    icp_total INTEGER := 0;
    icp_compliant INTEGER := 0;
    angola_total INTEGER := 0;
    angola_compliant INTEGER := 0;
    report_row RECORD;
BEGIN
    -- Executar relatório completo
    FOR report_row IN SELECT * FROM compliance_validators.generate_government_compliance_report(tenant_id) LOOP
        -- Contar requisitos por framework
        CASE report_row.framework
            WHEN 'eIDAS' THEN
                eidas_total := eidas_total + 1;
                IF report_row.is_compliant THEN
                    eidas_compliant := eidas_compliant + 1;
                END IF;
            WHEN 'ICP-Brasil' THEN
                icp_total := icp_total + 1;
                IF report_row.is_compliant THEN
                    icp_compliant := icp_compliant + 1;
                END IF;
            WHEN 'Angola' THEN
                angola_total := angola_total + 1;
                IF report_row.is_compliant THEN
                    angola_compliant := angola_compliant + 1;
                END IF;
        END CASE;
    END LOOP;
    
    -- Retornar resultados eIDAS
    IF eidas_total > 0 THEN
        framework := 'eIDAS';
        compliance_score := 4.0 * (eidas_compliant::NUMERIC / eidas_total);
        total_requirements := eidas_total;
        compliant_requirements := eidas_compliant;
        compliance_percentage := 100.0 * (eidas_compliant::NUMERIC / eidas_total);
        RETURN NEXT;
    END IF;
    
    -- Retornar resultados ICP-Brasil
    IF icp_total > 0 THEN
        framework := 'ICP-Brasil';
        compliance_score := 4.0 * (icp_compliant::NUMERIC / icp_total);
        total_requirements := icp_total;
        compliant_requirements := icp_compliant;
        compliance_percentage := 100.0 * (icp_compliant::NUMERIC / icp_total);
        RETURN NEXT;
    END IF;
    
    -- Retornar resultados Angola
    IF angola_total > 0 THEN
        framework := 'Angola';
        compliance_score := 4.0 * (angola_compliant::NUMERIC / angola_total);
        total_requirements := angola_total;
        compliant_requirements := angola_compliant;
        compliance_percentage := 100.0 * (angola_compliant::NUMERIC / angola_total);
        RETURN NEXT;
    END IF;
    
    -- Retornar pontuação geral
    framework := 'OVERALL';
    total_requirements := eidas_total + icp_total + angola_total;
    compliant_requirements := eidas_compliant + icp_compliant + angola_compliant;
    
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

-- Função para determinar o IRR (Índice de Risco Residual) para serviços governamentais
CREATE OR REPLACE FUNCTION compliance_validators.calculate_government_irr(
    tenant_id UUID,
    region_code VARCHAR(10)
) RETURNS VARCHAR(10) AS $$
DECLARE
    compliance_perc NUMERIC(5,2);
    framework VARCHAR(15);
    irr VARCHAR(10);
BEGIN
    -- Selecionar framework relevante baseado na região
    CASE region_code
        WHEN 'PT' THEN framework := 'eIDAS';
        WHEN 'BR' THEN framework := 'ICP-Brasil';
        WHEN 'AO' THEN framework := 'Angola';
        ELSE framework := 'OVERALL';
    END CASE;
    
    -- Obter percentual de conformidade para o framework
    IF framework != 'OVERALL' THEN
        SELECT compliance_percentage INTO compliance_perc 
        FROM compliance_validators.calculate_government_compliance_score(tenant_id) 
        WHERE framework = calculate_government_irr.framework;
    ELSE
        SELECT compliance_percentage INTO compliance_perc 
        FROM compliance_validators.calculate_government_compliance_score(tenant_id) 
        WHERE framework = 'OVERALL';
    END IF;
    
    -- Determinar IRR baseado no percentual
    IF compliance_perc >= 95 THEN
        irr := 'R1';
    ELSIF compliance_perc >= 85 THEN
        irr := 'R2';
    ELSIF compliance_perc >= 70 THEN
        irr := 'R3';
    ELSE
        irr := 'R4';
    END IF;
    
    RETURN irr;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Comentários de Documentação
-- =====================================================================

COMMENT ON SCHEMA compliance_validators IS 'Esquema para validadores de conformidade regulatória do INNOVABIZ';
COMMENT ON TABLE compliance_validators.eidas_requirements IS 'Requisitos de conformidade eIDAS para IAM em serviços governamentais da UE';
COMMENT ON TABLE compliance_validators.icp_brasil_requirements IS 'Requisitos de conformidade ICP-Brasil para IAM em serviços governamentais brasileiros';
COMMENT ON TABLE compliance_validators.angola_gov_requirements IS 'Requisitos para IAM em serviços governamentais de Angola';
COMMENT ON FUNCTION compliance_validators.validate_eidas_compliance IS 'Valida conformidade eIDAS para um tenant específico';
COMMENT ON FUNCTION compliance_validators.validate_icp_brasil_compliance IS 'Valida conformidade ICP-Brasil para um tenant específico';
COMMENT ON FUNCTION compliance_validators.validate_angola_gov_compliance IS 'Valida conformidade com requisitos de Angola para um tenant específico';
COMMENT ON FUNCTION compliance_validators.generate_government_compliance_report IS 'Gera relatório consolidado de conformidade para serviços governamentais';
COMMENT ON FUNCTION compliance_validators.calculate_government_compliance_score IS 'Calcula pontuação de conformidade para frameworks governamentais';
COMMENT ON FUNCTION compliance_validators.calculate_government_irr IS 'Calcula o Índice de Risco Residual baseado na conformidade governamental';

-- =====================================================================
-- Fim do Script
-- =====================================================================
