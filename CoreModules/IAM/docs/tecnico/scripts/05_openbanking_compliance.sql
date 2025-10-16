-- =====================================================================
-- INNOVABIZ - Validadores de Conformidade para Open Banking
-- Versão: 1.0.0
-- Data de Criação: 14/05/2025
-- Autor: INNOVABIZ DevOps
-- Descrição: Funções e procedimentos para validação de conformidade IAM
--            específicos para Open Banking e PSD2
-- Regiões Suportadas: UE/Portugal, Brasil, Reino Unido
-- =====================================================================

-- Criação de schema para validadores de conformidade se ainda não existir
CREATE SCHEMA IF NOT EXISTS compliance_validators;

-- =====================================================================
-- Tabelas de Referência para Requisitos Regulatórios
-- =====================================================================

-- Requisitos de PSD2 para IAM em Open Banking
CREATE TABLE compliance_validators.psd2_requirements (
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

-- Requisitos para Banco Central do Brasil (Open Banking/Open Finance)
CREATE TABLE compliance_validators.bcb_requirements (
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

-- Requisitos do Open Banking Implementation Entity (OBIE) - Reino Unido
CREATE TABLE compliance_validators.obie_requirements (
    requirement_id VARCHAR(20) PRIMARY KEY,
    requirement_name VARCHAR(255) NOT NULL,
    requirement_description TEXT NOT NULL,
    validation_query TEXT,
    implementation_details TEXT,
    required_auth_level VARCHAR(50),
    required_auth_methods VARCHAR(50)[],
    irr_threshold VARCHAR(10),
    is_mandatory BOOLEAN NOT NULL DEFAULT TRUE,
    applies_to_participants VARCHAR(50)[],
    relevant_standards VARCHAR(50)[],
    reference_url TEXT
);

-- =====================================================================
-- Inserção de dados básicos para Requisitos PSD2
-- =====================================================================

INSERT INTO compliance_validators.psd2_requirements (
    requirement_id, requirement_name, requirement_description, 
    validation_query, implementation_details, 
    required_auth_level, required_auth_methods, irr_threshold,
    relevant_articles, reference_url
) VALUES
('PSD2-IAM-01', 'Strong Customer Authentication (SCA)', 
 'Autenticação com pelo menos dois elementos independentes das categorias conhecimento, posse e inerência',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies 
  WHERE tenant_id = $1 
  AND policy_rules::jsonb @> ''{"mandatory_factor_categories": ["KNOWLEDGE", "POSSESSION"]}''
  OR policy_rules::jsonb @> ''{"mandatory_factor_categories": ["KNOWLEDGE", "INHERENCE"]}''
  OR policy_rules::jsonb @> ''{"mandatory_factor_categories": ["POSSESSION", "INHERENCE"]}''',
 'Verificar se há políticas que exigem pelo menos dois fatores de diferentes categorias',
 'ADVANCED', ARRAY['OB-01-02'], 'R1',
 ARRAY['Art. 97(1)'], 'https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX:32015L2366'),

('PSD2-IAM-02', 'Vinculação Dinâmica para Pagamentos', 
 'Elementos de autenticação devem estar dinamicamente vinculados ao valor da transação e ao beneficiário',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies 
  WHERE tenant_id = $1 
  AND policy_rules::jsonb @> ''{"psd2_compliant": true}''
  AND policy_rules::jsonb #> ''{region_specific_rules, EU, require_dynamic_linking}'' = ''true''::jsonb',
 'Verificar se há políticas com vinculação dinâmica em conformidade com PSD2',
 'VERY_ADVANCED', ARRAY['OB-01-02'], 'R1',
 ARRAY['Art. 97(2)'], 'https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX:32015L2366'),

('PSD2-IAM-03', 'Proteção de Dados de Autenticação', 
 'Medidas para proteger a confidencialidade e integridade dos dados de autenticação do usuário',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND method_id IN (''OB-01-01'', ''OB-01-02'') 
  AND is_enabled = TRUE',
 'Verificar configurações específicas para métodos de Open Banking',
 'ADVANCED', ARRAY['OB-01-01', 'OB-01-02'], 'R2',
 ARRAY['Art. 97(3)'], 'https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX:32015L2366'),

('PSD2-IAM-04', 'Isenções de SCA para Transações de Baixo Risco', 
 'Implementação de isenções permitidas para transações de baixo risco ou valor',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies 
  WHERE tenant_id = $1 
  AND policy_rules::jsonb ? ''exemptions''',
 'Verificar se há políticas com isenções configuradas',
 'ADVANCED', NULL, 'R2',
 ARRAY['RTS Art. 10-18'], 'https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX:32018R0389'),

('PSD2-IAM-05', 'Interface de Acesso Dedicada', 
 'Disponibilização de interface dedicada para provedores terceiros acessarem contas de pagamento',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies 
  WHERE tenant_id = $1 
  AND policy_rules::jsonb #> ''{region_specific_rules, EU, require_dedicated_interface}'' = ''true''::jsonb',
 'Verificar se há políticas com requisito de interface dedicada',
 'VERY_ADVANCED', ARRAY['OB-01-01'], 'R1',
 ARRAY['Art. 98(1)(d)'], 'https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX:32015L2366');

-- =====================================================================
-- Inserção de dados básicos para Requisitos BCB (Brasil)
-- =====================================================================

INSERT INTO compliance_validators.bcb_requirements (
    requirement_id, requirement_name, requirement_description, 
    validation_query, implementation_details, 
    required_auth_level, required_auth_methods, irr_threshold,
    applies_to_phases, relevant_resolutions, reference_url
) VALUES
('BCB-IAM-01', 'Autenticação no Open Finance Brasil', 
 'Implementação de autenticação robusta para participantes do ecossistema Open Finance',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_methods 
  WHERE method_id LIKE ''OB%'' 
  AND security_level IN (''ADVANCED'', ''VERY_ADVANCED'')',
 'Verificar se há métodos avançados específicos para Open Banking',
 'ADVANCED', ARRAY['OB-01-01', 'OB-01-02'], 'R2',
 ARRAY['Fase 1', 'Fase 2', 'Fase 3'], 
 ARRAY['Resolução BCB Nº 32/2020'], 'https://www.bcb.gov.br/estabilidadefinanceira/openfinance'),

('BCB-IAM-02', 'Consentimento Dinâmico', 
 'Mecanismo de consentimento e autenticação específico para compartilhamento de dados',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND config_parameters::jsonb ? ''consent_management''',
 'Verificar configurações com gestão de consentimento',
 'ADVANCED', ARRAY['OB-01-01'], 'R2',
 ARRAY['Fase 2', 'Fase 3'], 
 ARRAY['Resolução BCB Nº 32/2020'], 'https://www.bcb.gov.br/estabilidadefinanceira/openfinance'),

('BCB-IAM-03', 'Diretório de Participantes', 
 'Integração com o diretório oficial de participantes para validação de credenciais',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND config_parameters::jsonb ? ''directory_integration''',
 'Verificar configurações com integração ao diretório de participantes',
 'ADVANCED', NULL, 'R2',
 ARRAY['Fase 1'], 
 ARRAY['Manual Operacional do Diretório'], 'https://openfinancebrasil.org.br'),

('BCB-IAM-04', 'Autenticação com Certificados ICP-Brasil', 
 'Uso de certificados validados pela Infraestrutura de Chaves Públicas Brasileira',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND method_id = ''PB-02-01'' 
  AND config_parameters::jsonb ? ''icp_brasil''',
 'Verificar configurações para certificados ICP-Brasil',
 'VERY_ADVANCED', ARRAY['PB-02-01', 'PB-02-02'], 'R1',
 ARRAY['Fase 1', 'Fase 2'], 
 ARRAY['Resolução BCB Nº 32/2020'], 'https://www.bcb.gov.br/estabilidadefinanceira/openfinance'),

('BCB-IAM-05', 'Registros de Autenticação e Consentimento', 
 'Manutenção de registros auditáveis de autenticação e consentimento',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_history 
  WHERE tenant_id = $1 
  AND event_details::jsonb ? ''consent_id''',
 'Verificar histórico com detalhes de consentimento',
 'ADVANCED', NULL, 'R2',
 ARRAY['Fase 2', 'Fase 3'], 
 ARRAY['Resolução BCB Nº 32/2020'], 'https://www.bcb.gov.br/estabilidadefinanceira/openfinance');

-- =====================================================================
-- Inserção de dados básicos para Requisitos OBIE (Reino Unido)
-- =====================================================================

INSERT INTO compliance_validators.obie_requirements (
    requirement_id, requirement_name, requirement_description, 
    validation_query, implementation_details, 
    required_auth_level, required_auth_methods, irr_threshold,
    relevant_standards, reference_url
) VALUES
('OBIE-IAM-01', 'FAPI Profile e OpenID Connect', 
 'Implementação do perfil FAPI e OpenID Connect para autenticação segura',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND method_id = ''FS-01-02'' 
  AND config_parameters::jsonb ? ''fapi_profile''',
 'Verificar configurações com suporte a FAPI',
 'VERY_ADVANCED', ARRAY['FS-01-02', 'OB-01-01'], 'R1',
 ARRAY['FAPI 1.0', 'OIDC'], 'https://openbanking.org.uk/'),

('OBIE-IAM-02', 'Detached Signature', 
 'Suporte para assinatura separada para autenticação de requisições',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND config_parameters::jsonb ? ''detached_signature''',
 'Verificar configurações com suporte a assinatura separada',
 'ADVANCED', ARRAY['FS-01-02', 'OB-01-01'], 'R2',
 ARRAY['Signing Standards v1.1'], 'https://openbanking.org.uk/'),

('OBIE-IAM-03', 'Customer Experience Guidelines', 
 'Aderência às diretrizes de experiência do cliente para fluxos de autenticação',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies 
  WHERE tenant_id = $1 
  AND policy_rules::jsonb ? ''user_experience''',
 'Verificar políticas com diretrizes de experiência do usuário',
 'ADVANCED', NULL, 'R2',
 ARRAY['CX Guidelines v3.1'], 'https://openbanking.org.uk/'),

('OBIE-IAM-04', 'Gerenciamento de Consentimento', 
 'Implementação de mecanismos para gerenciar consentimentos dos usuários',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND config_parameters::jsonb ? ''consent_dashboard''',
 'Verificar configurações com dashboard de consentimento',
 'ADVANCED', ARRAY['OB-01-01'], 'R2',
 ARRAY['CX Guidelines v3.1'], 'https://openbanking.org.uk/'),

('OBIE-IAM-05', 'Autenticação App-to-App', 
 'Suporte para redirecionamento app-to-app para aplicativos móveis',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND config_parameters::jsonb ? ''app_to_app''',
 'Verificar configurações com suporte app-to-app',
 'ADVANCED', ARRAY['PB-03-01', 'OB-01-01'], 'R2',
 ARRAY['App-to-App Specification'], 'https://openbanking.org.uk/');

-- =====================================================================
-- Funções de Validação para Conformidade em Open Banking
-- =====================================================================

-- Função para validar conformidade PSD2
CREATE OR REPLACE FUNCTION compliance_validators.validate_psd2_compliance(
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
    FOR req IN SELECT * FROM compliance_validators.psd2_requirements LOOP
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

-- Função para validar conformidade BCB
CREATE OR REPLACE FUNCTION compliance_validators.validate_bcb_compliance(
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
    FOR req IN SELECT * FROM compliance_validators.bcb_requirements LOOP
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

-- Função para validar conformidade OBIE
CREATE OR REPLACE FUNCTION compliance_validators.validate_obie_compliance(
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
    FOR req IN SELECT * FROM compliance_validators.obie_requirements LOOP
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

-- Função para gerar relatório consolidado multi-framework para Open Banking
CREATE OR REPLACE FUNCTION compliance_validators.generate_openbanking_compliance_report(
    tenant_id UUID,
    include_psd2 BOOLEAN DEFAULT TRUE,
    include_bcb BOOLEAN DEFAULT TRUE,
    include_obie BOOLEAN DEFAULT TRUE
) RETURNS TABLE (
    framework VARCHAR(10),
    requirement_id VARCHAR(20),
    requirement_name VARCHAR(255),
    is_compliant BOOLEAN,
    details TEXT
) AS $$
BEGIN
    -- Validar PSD2
    IF include_psd2 THEN
        RETURN QUERY 
        SELECT 'PSD2'::VARCHAR(10) as framework, * FROM compliance_validators.validate_psd2_compliance(tenant_id);
    END IF;
    
    -- Validar BCB
    IF include_bcb THEN
        RETURN QUERY 
        SELECT 'BCB'::VARCHAR(10) as framework, * FROM compliance_validators.validate_bcb_compliance(tenant_id);
    END IF;
    
    -- Validar OBIE
    IF include_obie THEN
        RETURN QUERY 
        SELECT 'OBIE'::VARCHAR(10) as framework, * FROM compliance_validators.validate_obie_compliance(tenant_id);
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Função para gerar pontuação de conformidade para Open Banking
CREATE OR REPLACE FUNCTION compliance_validators.calculate_openbanking_compliance_score(
    tenant_id UUID
) RETURNS TABLE (
    framework VARCHAR(10),
    compliance_score NUMERIC(5,2),
    total_requirements INTEGER,
    compliant_requirements INTEGER,
    compliance_percentage NUMERIC(5,2)
) AS $$
DECLARE
    psd2_total INTEGER := 0;
    psd2_compliant INTEGER := 0;
    bcb_total INTEGER := 0;
    bcb_compliant INTEGER := 0;
    obie_total INTEGER := 0;
    obie_compliant INTEGER := 0;
    report_row RECORD;
BEGIN
    -- Executar relatório completo
    FOR report_row IN SELECT * FROM compliance_validators.generate_openbanking_compliance_report(tenant_id) LOOP
        -- Contar requisitos por framework
        CASE report_row.framework
            WHEN 'PSD2' THEN
                psd2_total := psd2_total + 1;
                IF report_row.is_compliant THEN
                    psd2_compliant := psd2_compliant + 1;
                END IF;
            WHEN 'BCB' THEN
                bcb_total := bcb_total + 1;
                IF report_row.is_compliant THEN
                    bcb_compliant := bcb_compliant + 1;
                END IF;
            WHEN 'OBIE' THEN
                obie_total := obie_total + 1;
                IF report_row.is_compliant THEN
                    obie_compliant := obie_compliant + 1;
                END IF;
        END CASE;
    END LOOP;
    
    -- Retornar resultados PSD2
    IF psd2_total > 0 THEN
        framework := 'PSD2';
        compliance_score := 4.0 * (psd2_compliant::NUMERIC / psd2_total);
        total_requirements := psd2_total;
        compliant_requirements := psd2_compliant;
        compliance_percentage := 100.0 * (psd2_compliant::NUMERIC / psd2_total);
        RETURN NEXT;
    END IF;
    
    -- Retornar resultados BCB
    IF bcb_total > 0 THEN
        framework := 'BCB';
        compliance_score := 4.0 * (bcb_compliant::NUMERIC / bcb_total);
        total_requirements := bcb_total;
        compliant_requirements := bcb_compliant;
        compliance_percentage := 100.0 * (bcb_compliant::NUMERIC / bcb_total);
        RETURN NEXT;
    END IF;
    
    -- Retornar resultados OBIE
    IF obie_total > 0 THEN
        framework := 'OBIE';
        compliance_score := 4.0 * (obie_compliant::NUMERIC / obie_total);
        total_requirements := obie_total;
        compliant_requirements := obie_compliant;
        compliance_percentage := 100.0 * (obie_compliant::NUMERIC / obie_total);
        RETURN NEXT;
    END IF;
    
    -- Retornar pontuação geral
    framework := 'OVERALL';
    total_requirements := psd2_total + bcb_total + obie_total;
    compliant_requirements := psd2_compliant + bcb_compliant + obie_compliant;
    
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

-- Função para determinar o IRR (Índice de Risco Residual) baseado na conformidade
CREATE OR REPLACE FUNCTION compliance_validators.calculate_openbanking_irr(
    tenant_id UUID
) RETURNS VARCHAR(10) AS $$
DECLARE
    compliance_perc NUMERIC(5,2);
    irr VARCHAR(10);
BEGIN
    SELECT compliance_percentage INTO compliance_perc 
    FROM compliance_validators.calculate_openbanking_compliance_score(tenant_id) 
    WHERE framework = 'OVERALL';
    
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
COMMENT ON TABLE compliance_validators.psd2_requirements IS 'Requisitos de conformidade PSD2 para IAM em Open Banking';
COMMENT ON TABLE compliance_validators.bcb_requirements IS 'Requisitos de conformidade do Banco Central do Brasil para Open Finance';
COMMENT ON TABLE compliance_validators.obie_requirements IS 'Requisitos do Open Banking Implementation Entity (UK) para Open Banking';
COMMENT ON FUNCTION compliance_validators.validate_psd2_compliance IS 'Valida conformidade PSD2 para um tenant específico';
COMMENT ON FUNCTION compliance_validators.validate_bcb_compliance IS 'Valida conformidade BCB para um tenant específico';
COMMENT ON FUNCTION compliance_validators.validate_obie_compliance IS 'Valida conformidade OBIE para um tenant específico';
COMMENT ON FUNCTION compliance_validators.generate_openbanking_compliance_report IS 'Gera relatório consolidado de conformidade para Open Banking';
COMMENT ON FUNCTION compliance_validators.calculate_openbanking_compliance_score IS 'Calcula pontuação de conformidade para frameworks de Open Banking';
COMMENT ON FUNCTION compliance_validators.calculate_openbanking_irr IS 'Calcula o Índice de Risco Residual baseado na conformidade de Open Banking';

-- =====================================================================
-- Fim do Script
-- =====================================================================
