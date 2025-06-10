-- =====================================================================
-- INNOVABIZ - Validadores de Conformidade para AR/VR
-- Versão: 1.0.0
-- Data de Criação: 14/05/2025
-- Autor: INNOVABIZ DevOps
-- Descrição: Funções e procedimentos para validação de conformidade IAM
--            específicos para Realidade Aumentada e Realidade Virtual
-- Regiões Suportadas: Global
-- =====================================================================

-- Criação de schema para validadores de conformidade se ainda não existir
CREATE SCHEMA IF NOT EXISTS compliance_validators;

-- =====================================================================
-- Tabelas de Referência para Requisitos Regulatórios e Padrões AR/VR
-- =====================================================================

-- Requisitos de IEEE XR Standards para AR/VR
CREATE TABLE compliance_validators.ieee_xr_requirements (
    requirement_id VARCHAR(20) PRIMARY KEY,
    requirement_name VARCHAR(255) NOT NULL,
    requirement_description TEXT NOT NULL,
    validation_query TEXT,
    implementation_details TEXT,
    required_auth_level VARCHAR(50),
    required_auth_methods VARCHAR(50)[],
    irr_threshold VARCHAR(10),
    is_mandatory BOOLEAN NOT NULL DEFAULT TRUE,
    applies_to_environments VARCHAR(50)[],
    relevant_standards VARCHAR(50)[],
    reference_url TEXT
);

-- Requisitos de Segurança para AR/VR do NIST
CREATE TABLE compliance_validators.nist_xr_requirements (
    requirement_id VARCHAR(20) PRIMARY KEY,
    requirement_name VARCHAR(255) NOT NULL,
    requirement_description TEXT NOT NULL,
    validation_query TEXT,
    implementation_details TEXT,
    required_auth_level VARCHAR(50),
    required_auth_methods VARCHAR(50)[],
    irr_threshold VARCHAR(10),
    is_mandatory BOOLEAN NOT NULL DEFAULT TRUE,
    applies_to_categories VARCHAR(50)[],
    relevant_publications VARCHAR(50)[],
    reference_url TEXT
);

-- Requisitos do OpenXR Standard para Autenticação
CREATE TABLE compliance_validators.openxr_requirements (
    requirement_id VARCHAR(20) PRIMARY KEY,
    requirement_name VARCHAR(255) NOT NULL,
    requirement_description TEXT NOT NULL,
    validation_query TEXT,
    implementation_details TEXT,
    required_auth_level VARCHAR(50),
    required_auth_methods VARCHAR(50)[],
    irr_threshold VARCHAR(10),
    is_mandatory BOOLEAN NOT NULL DEFAULT TRUE,
    applies_to_platforms VARCHAR(50)[],
    relevant_guidelines VARCHAR(50)[],
    reference_url TEXT
);

-- =====================================================================
-- Inserção de dados básicos para Requisitos IEEE XR
-- =====================================================================

INSERT INTO compliance_validators.ieee_xr_requirements (
    requirement_id, requirement_name, requirement_description, 
    validation_query, implementation_details, 
    required_auth_level, required_auth_methods, irr_threshold,
    relevant_standards, reference_url
) VALUES
('IEEE-XR-01', 'Autenticação Espacial', 
 'Implementação de autenticação baseada em gestos espaciais conforme IEEE 2888',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_methods 
  WHERE method_id = ''AR-01-01'' 
  AND implementation_status IN (''IMPLEMENTED'', ''IN_PROGRESS'')',
 'Verificar suporte para autenticação por gesto espacial',
 'ADVANCED', ARRAY['AR-01-01'], 'R2',
 ARRAY['IEEE 2888'], 'https://standards.ieee.org/project/2888.html'),

('IEEE-XR-02', 'Autenticação Baseada em Olhar', 
 'Implementação de autenticação baseada em padrões de olhar conforme IEEE 2888',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_methods 
  WHERE method_id = ''AR-01-02'' 
  AND implementation_status IN (''IMPLEMENTED'', ''IN_PROGRESS'')',
 'Verificar suporte para autenticação por padrão de olhar',
 'ADVANCED', ARRAY['AR-01-02'], 'R2',
 ARRAY['IEEE 2888'], 'https://standards.ieee.org/project/2888.html'),

('IEEE-XR-03', 'Multi-fator em Ambientes Imersivos', 
 'Suporte para autenticação multifatorial em contextos de AR/VR',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies 
  WHERE tenant_id = $1 
  AND policy_type = ''MFA'' 
  AND policy_rules::jsonb @> ''{"allowed_methods": ["AR-01-01"]}''',
 'Verificar políticas MFA com métodos AR',
 'ADVANCED', ARRAY['AR-01-01', 'AR-01-02'], 'R2',
 ARRAY['IEEE 2888.1'], 'https://standards.ieee.org/project/2888_1.html'),

('IEEE-XR-04', 'Verificação Contínua em XR', 
 'Implementação de autenticação contínua em ambientes AR/VR',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND method_id IN (''AR-01-01'', ''AR-01-02'') 
  AND config_parameters::jsonb ? ''continuous_authentication''',
 'Verificar configurações para autenticação contínua',
 'ADVANCED', ARRAY['AR-01-01', 'AR-01-02'], 'R2',
 ARRAY['IEEE 2888.3'], 'https://standards.ieee.org/project/2888_3.html'),

('IEEE-XR-05', 'Segurança de Interface Sensorial', 
 'Medidas para proteger entradas sensoriais usadas na autenticação',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND method_id IN (''AR-01-01'', ''AR-01-02'') 
  AND config_parameters::jsonb ? ''sensory_protection''',
 'Verificar proteções para interfaces sensoriais',
 'VERY_ADVANCED', NULL, 'R1',
 ARRAY['IEEE 2888.6'], 'https://standards.ieee.org/project/2888_6.html');

-- =====================================================================
-- Inserção de dados básicos para Requisitos NIST XR
-- =====================================================================

INSERT INTO compliance_validators.nist_xr_requirements (
    requirement_id, requirement_name, requirement_description, 
    validation_query, implementation_details, 
    required_auth_level, required_auth_methods, irr_threshold,
    relevant_publications, reference_url
) VALUES
('NIST-XR-01', 'Autenticação Baseada em Contexto', 
 'Uso de informações contextuais para autenticação em AR/VR',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies 
  WHERE tenant_id = $1 
  AND policy_type = ''ADAPTIVE'' 
  AND policy_rules::jsonb ? ''context_factors''',
 'Verificar políticas adaptativas com fatores contextuais',
 'ADVANCED', NULL, 'R2',
 ARRAY['NIST SP 800-207'], 'https://csrc.nist.gov/publications/detail/sp/800-207/final'),

('NIST-XR-02', 'Identity Proofing em XR', 
 'Procedimentos de prova de identidade adaptados para ambientes XR',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND method_id IN (''AR-01-01'', ''AR-01-02'') 
  AND config_parameters::jsonb ? ''identity_proofing''',
 'Verificar configurações para prova de identidade',
 'VERY_ADVANCED', ARRAY['AR-01-01', 'AR-01-02'], 'R1',
 ARRAY['NIST SP 800-63A'], 'https://pages.nist.gov/800-63-3/sp800-63a.html'),

('NIST-XR-03', 'Autenticação Zero-Trust em XR', 
 'Implementação de princípios de segurança Zero Trust em ambientes XR',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies 
  WHERE tenant_id = $1 
  AND policy_rules::jsonb ? ''zero_trust''',
 'Verificar políticas com princípios zero-trust',
 'VERY_ADVANCED', NULL, 'R1',
 ARRAY['NIST SP 800-207'], 'https://csrc.nist.gov/publications/detail/sp/800-207/final'),

('NIST-XR-04', 'Proteção de Dados Biométricos em XR', 
 'Medidas para proteger dados biométricos coletados em ambientes XR',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND method_id IN (''AR-01-01'', ''AR-01-02'') 
  AND config_parameters::jsonb ? ''biometric_protection''',
 'Verificar proteções para dados biométricos',
 'VERY_ADVANCED', NULL, 'R1',
 ARRAY['NIST SP 800-76'], 'https://csrc.nist.gov/publications/detail/sp/800-76/2/final'),

('NIST-XR-05', 'Mitigação de Ataques em XR', 
 'Implementação de controles para mitigar ataques específicos de AR/VR',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND method_id IN (''AR-01-01'', ''AR-01-02'') 
  AND config_parameters::jsonb ? ''attack_mitigation''',
 'Verificar mitigações para ataques em AR/VR',
 'ADVANCED', NULL, 'R2',
 ARRAY['NIST SP 800-63B'], 'https://pages.nist.gov/800-63-3/sp800-63b.html');

-- =====================================================================
-- Inserção de dados básicos para Requisitos OpenXR
-- =====================================================================

INSERT INTO compliance_validators.openxr_requirements (
    requirement_id, requirement_name, requirement_description, 
    validation_query, implementation_details, 
    required_auth_level, required_auth_methods, irr_threshold,
    relevant_guidelines, reference_url
) VALUES
('OXR-01', 'Integração com OpenXR Runtime', 
 'Suporte para autenticação integrada com runtimes OpenXR',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND method_id IN (''AR-01-01'', ''AR-01-02'') 
  AND config_parameters::jsonb ? ''openxr_integration''',
 'Verificar integração com OpenXR',
 'ADVANCED', ARRAY['AR-01-01', 'AR-01-02'], 'R2',
 ARRAY['OpenXR 1.0'], 'https://www.khronos.org/openxr/'),

('OXR-02', 'Suporte para Dispositivos Cross-Platform', 
 'Autenticação consistente entre diferentes plataformas e dispositivos XR',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND method_id IN (''AR-01-01'', ''AR-01-02'') 
  AND config_parameters::jsonb ? ''cross_platform''',
 'Verificar suporte cross-platform',
 'ADVANCED', ARRAY['AR-01-01', 'AR-01-02'], 'R2',
 ARRAY['OpenXR 1.0'], 'https://www.khronos.org/openxr/'),

('OXR-03', 'Segurança de Input em OpenXR', 
 'Proteções para entradas de gestos e olhar em conformidade com OpenXR',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND method_id IN (''AR-01-01'', ''AR-01-02'') 
  AND config_parameters::jsonb ? ''input_security''',
 'Verificar segurança de entrada',
 'ADVANCED', NULL, 'R2',
 ARRAY['OpenXR Input Subsystem'], 'https://www.khronos.org/openxr/'),

('OXR-04', 'Extensões de Segurança OpenXR', 
 'Uso de extensões de segurança OpenXR para autenticação',
 'SELECT COUNT(*) > 0 FROM iam_core.factor_configurations 
  WHERE tenant_id = $1 
  AND method_id IN (''AR-01-01'', ''AR-01-02'') 
  AND config_parameters::jsonb ? ''openxr_extensions''',
 'Verificar uso de extensões de segurança',
 'ADVANCED', NULL, 'R2',
 ARRAY['OpenXR Extensions'], 'https://www.khronos.org/openxr/'),

('OXR-05', 'Conformidade com OpenXR Security Guidelines', 
 'Implementação em conformidade com diretrizes de segurança OpenXR',
 'SELECT COUNT(*) > 0 FROM iam_core.authentication_policies 
  WHERE tenant_id = $1 
  AND policy_rules::jsonb ? ''openxr_security_guidelines''',
 'Verificar conformidade com diretrizes',
 'ADVANCED', NULL, 'R2',
 ARRAY['OpenXR Security Model'], 'https://www.khronos.org/openxr/');

-- =====================================================================
-- Funções de Validação para Conformidade AR/VR
-- =====================================================================

-- Função para validar conformidade IEEE XR
CREATE OR REPLACE FUNCTION compliance_validators.validate_ieee_xr_compliance(
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
    FOR req IN SELECT * FROM compliance_validators.ieee_xr_requirements LOOP
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

-- Função para validar conformidade NIST XR
CREATE OR REPLACE FUNCTION compliance_validators.validate_nist_xr_compliance(
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
    FOR req IN SELECT * FROM compliance_validators.nist_xr_requirements LOOP
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

-- Função para validar conformidade OpenXR
CREATE OR REPLACE FUNCTION compliance_validators.validate_openxr_compliance(
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
    FOR req IN SELECT * FROM compliance_validators.openxr_requirements LOOP
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

-- Função para gerar relatório consolidado multi-framework para AR/VR
CREATE OR REPLACE FUNCTION compliance_validators.generate_arvr_compliance_report(
    tenant_id UUID,
    include_ieee BOOLEAN DEFAULT TRUE,
    include_nist BOOLEAN DEFAULT TRUE,
    include_openxr BOOLEAN DEFAULT TRUE
) RETURNS TABLE (
    framework VARCHAR(15),
    requirement_id VARCHAR(20),
    requirement_name VARCHAR(255),
    is_compliant BOOLEAN,
    details TEXT
) AS $$
BEGIN
    -- Validar IEEE XR
    IF include_ieee THEN
        RETURN QUERY 
        SELECT 'IEEE XR'::VARCHAR(15) as framework, * FROM compliance_validators.validate_ieee_xr_compliance(tenant_id);
    END IF;
    
    -- Validar NIST XR
    IF include_nist THEN
        RETURN QUERY 
        SELECT 'NIST XR'::VARCHAR(15) as framework, * FROM compliance_validators.validate_nist_xr_compliance(tenant_id);
    END IF;
    
    -- Validar OpenXR
    IF include_openxr THEN
        RETURN QUERY 
        SELECT 'OpenXR'::VARCHAR(15) as framework, * FROM compliance_validators.validate_openxr_compliance(tenant_id);
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Função para gerar pontuação de conformidade para AR/VR
CREATE OR REPLACE FUNCTION compliance_validators.calculate_arvr_compliance_score(
    tenant_id UUID
) RETURNS TABLE (
    framework VARCHAR(15),
    compliance_score NUMERIC(5,2),
    total_requirements INTEGER,
    compliant_requirements INTEGER,
    compliance_percentage NUMERIC(5,2)
) AS $$
DECLARE
    ieee_total INTEGER := 0;
    ieee_compliant INTEGER := 0;
    nist_total INTEGER := 0;
    nist_compliant INTEGER := 0;
    openxr_total INTEGER := 0;
    openxr_compliant INTEGER := 0;
    report_row RECORD;
BEGIN
    -- Executar relatório completo
    FOR report_row IN SELECT * FROM compliance_validators.generate_arvr_compliance_report(tenant_id) LOOP
        -- Contar requisitos por framework
        CASE report_row.framework
            WHEN 'IEEE XR' THEN
                ieee_total := ieee_total + 1;
                IF report_row.is_compliant THEN
                    ieee_compliant := ieee_compliant + 1;
                END IF;
            WHEN 'NIST XR' THEN
                nist_total := nist_total + 1;
                IF report_row.is_compliant THEN
                    nist_compliant := nist_compliant + 1;
                END IF;
            WHEN 'OpenXR' THEN
                openxr_total := openxr_total + 1;
                IF report_row.is_compliant THEN
                    openxr_compliant := openxr_compliant + 1;
                END IF;
        END CASE;
    END LOOP;
    
    -- Retornar resultados IEEE XR
    IF ieee_total > 0 THEN
        framework := 'IEEE XR';
        compliance_score := 4.0 * (ieee_compliant::NUMERIC / ieee_total);
        total_requirements := ieee_total;
        compliant_requirements := ieee_compliant;
        compliance_percentage := 100.0 * (ieee_compliant::NUMERIC / ieee_total);
        RETURN NEXT;
    END IF;
    
    -- Retornar resultados NIST XR
    IF nist_total > 0 THEN
        framework := 'NIST XR';
        compliance_score := 4.0 * (nist_compliant::NUMERIC / nist_total);
        total_requirements := nist_total;
        compliant_requirements := nist_compliant;
        compliance_percentage := 100.0 * (nist_compliant::NUMERIC / nist_total);
        RETURN NEXT;
    END IF;
    
    -- Retornar resultados OpenXR
    IF openxr_total > 0 THEN
        framework := 'OpenXR';
        compliance_score := 4.0 * (openxr_compliant::NUMERIC / openxr_total);
        total_requirements := openxr_total;
        compliant_requirements := openxr_compliant;
        compliance_percentage := 100.0 * (openxr_compliant::NUMERIC / openxr_total);
        RETURN NEXT;
    END IF;
    
    -- Retornar pontuação geral
    framework := 'OVERALL';
    total_requirements := ieee_total + nist_total + openxr_total;
    compliant_requirements := ieee_compliant + nist_compliant + openxr_compliant;
    
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

-- Função para determinar o IRR (Índice de Risco Residual) para AR/VR
CREATE OR REPLACE FUNCTION compliance_validators.calculate_arvr_irr(
    tenant_id UUID
) RETURNS VARCHAR(10) AS $$
DECLARE
    compliance_perc NUMERIC(5,2);
    irr VARCHAR(10);
BEGIN
    SELECT compliance_percentage INTO compliance_perc 
    FROM compliance_validators.calculate_arvr_compliance_score(tenant_id) 
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
COMMENT ON TABLE compliance_validators.ieee_xr_requirements IS 'Requisitos de conformidade IEEE para AR/VR';
COMMENT ON TABLE compliance_validators.nist_xr_requirements IS 'Requisitos de conformidade NIST para AR/VR';
COMMENT ON TABLE compliance_validators.openxr_requirements IS 'Requisitos de conformidade OpenXR para AR/VR';
COMMENT ON FUNCTION compliance_validators.validate_ieee_xr_compliance IS 'Valida conformidade IEEE XR para um tenant específico';
COMMENT ON FUNCTION compliance_validators.validate_nist_xr_compliance IS 'Valida conformidade NIST XR para um tenant específico';
COMMENT ON FUNCTION compliance_validators.validate_openxr_compliance IS 'Valida conformidade OpenXR para um tenant específico';
COMMENT ON FUNCTION compliance_validators.generate_arvr_compliance_report IS 'Gera relatório consolidado de conformidade para AR/VR';
COMMENT ON FUNCTION compliance_validators.calculate_arvr_compliance_score IS 'Calcula pontuação de conformidade para frameworks de AR/VR';
COMMENT ON FUNCTION compliance_validators.calculate_arvr_irr IS 'Calcula o Índice de Risco Residual baseado na conformidade de AR/VR';

-- =====================================================================
-- Fim do Script
-- =====================================================================
