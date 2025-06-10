-- =====================================================================
-- INNOVABIZ - Integrador de Validadores de Conformidade
-- Versão: 1.0.0
-- Data de Criação: 14/05/2025
-- Autor: INNOVABIZ DevOps
-- Descrição: Framework integrador para validadores de conformidade
--            setoriais do módulo IAM
-- Regiões Suportadas: Global, UE, Brasil, Angola, EUA
-- =====================================================================

-- Criação de schema para o integrador de validadores se ainda não existir
CREATE SCHEMA IF NOT EXISTS compliance_integrator;

-- =====================================================================
-- Tabelas de Mapeamento e Configuração para Integração Multi-Setor
-- =====================================================================

-- Tabela de mapeamento de setores
CREATE TABLE compliance_integrator.sectors (
    sector_id VARCHAR(30) PRIMARY KEY,
    sector_name VARCHAR(100) NOT NULL,
    description TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    priority INTEGER NOT NULL,
    validator_modules VARCHAR(100)[] NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabela de mapeamento de regulações por setor
CREATE TABLE compliance_integrator.sector_regulations (
    sector_id VARCHAR(30) REFERENCES compliance_integrator.sectors(sector_id),
    regulation_id VARCHAR(30),
    regulation_name VARCHAR(100) NOT NULL,
    description TEXT,
    region VARCHAR(50),
    validator_function TEXT NOT NULL,
    irr_threshold VARCHAR(10),
    PRIMARY KEY (sector_id, regulation_id)
);

-- Tabela de configuração de validação por tenant
CREATE TABLE compliance_integrator.tenant_validator_config (
    tenant_id UUID PRIMARY KEY,
    active_sectors VARCHAR(30)[],
    default_report_format VARCHAR(20) DEFAULT 'JSON',
    report_schedule JSONB,
    enabled_validators JSONB,
    notification_settings JSONB,
    reporting_threshold INTEGER DEFAULT 70,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabela de histórico de validações
CREATE TABLE compliance_integrator.validation_history (
    validation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    validation_type VARCHAR(50) NOT NULL,
    sectors VARCHAR(30)[],
    regulations JSONB,
    score NUMERIC(5,2),
    irr VARCHAR(10),
    validation_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    report_url TEXT,
    triggered_by VARCHAR(100),
    summary JSONB
);

-- =====================================================================
-- Inserção de dados básicos para setores
-- =====================================================================

INSERT INTO compliance_integrator.sectors (
    sector_id, sector_name, description, priority, validator_modules
) VALUES
('HEALTHCARE', 'Saúde', 'Setor de saúde e telemedicina', 1, 
 ARRAY['healthcare_compliance']),
('FINANCIAL', 'Financeiro', 'Setor financeiro e bancário', 2, 
 ARRAY['openbanking_compliance']),
('GOVERNMENT', 'Governamental', 'Setor governamental e público', 3, 
 ARRAY['governmental_compliance']),
('ARVR', 'Realidade Aumentada/Virtual', 'Setor de AR/VR e tecnologias imersivas', 4, 
 ARRAY['ar_vr_compliance']),
('MULTI', 'Multi-Setorial', 'Validação aplicável a múltiplos setores', 5, 
 ARRAY['healthcare_compliance', 'openbanking_compliance', 'governmental_compliance', 'ar_vr_compliance']);

-- =====================================================================
-- Inserção de dados básicos para regulações por setor
-- =====================================================================

-- Saúde
INSERT INTO compliance_integrator.sector_regulations (
    sector_id, regulation_id, regulation_name, description, 
    region, validator_function, irr_threshold
) VALUES
('HEALTHCARE', 'HIPAA', 'Health Insurance Portability and Accountability Act', 
 'Regulação americana para proteção de dados de saúde', 
 'EUA', 'compliance_validators.validate_hipaa_compliance', 'R2'),
('HEALTHCARE', 'GDPR_HEALTH', 'General Data Protection Regulation (Healthcare)', 
 'Requisitos de GDPR específicos para o setor de saúde', 
 'UE', 'compliance_validators.validate_gdpr_healthcare_compliance', 'R2'),
('HEALTHCARE', 'LGPD_HEALTH', 'Lei Geral de Proteção de Dados (Saúde)', 
 'Requisitos de LGPD específicos para o setor de saúde', 
 'Brasil', 'compliance_validators.validate_lgpd_healthcare_compliance', 'R2');

-- Financeiro
INSERT INTO compliance_integrator.sector_regulations (
    sector_id, regulation_id, regulation_name, description, 
    region, validator_function, irr_threshold
) VALUES
('FINANCIAL', 'PSD2', 'Payment Services Directive 2', 
 'Diretiva europeia para serviços de pagamento', 
 'UE', 'compliance_validators.validate_psd2_compliance', 'R1'),
('FINANCIAL', 'OPEN_BANKING_BR', 'Open Banking Brasil', 
 'Regulação brasileira para Open Banking', 
 'Brasil', 'compliance_validators.validate_bcb_compliance', 'R1'),
('FINANCIAL', 'OPEN_BANKING_UK', 'Open Banking UK', 
 'Regulação britânica para Open Banking', 
 'Reino Unido', 'compliance_validators.validate_obie_compliance', 'R1');

-- Governamental
INSERT INTO compliance_integrator.sector_regulations (
    sector_id, regulation_id, regulation_name, description, 
    region, validator_function, irr_threshold
) VALUES
('GOVERNMENT', 'EIDAS', 'Electronic Identification, Authentication and Trust Services', 
 'Regulação europeia para identidade digital', 
 'UE', 'compliance_validators.validate_eidas_compliance', 'R1'),
('GOVERNMENT', 'ICP_BRASIL', 'Infraestrutura de Chaves Públicas Brasileira', 
 'Padrão brasileiro para certificados digitais', 
 'Brasil', 'compliance_validators.validate_icp_brasil_compliance', 'R1'),
('GOVERNMENT', 'ANG_EGOV', 'Padrões de E-Governo de Angola', 
 'Regulações de governo eletrônico de Angola', 
 'Angola', 'compliance_validators.validate_angola_egov_compliance', 'R2');

-- AR/VR
INSERT INTO compliance_integrator.sector_regulations (
    sector_id, regulation_id, regulation_name, description, 
    region, validator_function, irr_threshold
) VALUES
('ARVR', 'IEEE_XR', 'IEEE XR Standards', 
 'Padrões IEEE para realidade estendida', 
 'Global', 'compliance_validators.validate_ieee_xr_compliance', 'R2'),
('ARVR', 'NIST_XR', 'NIST XR Standards', 
 'Diretrizes NIST para realidade estendida', 
 'EUA', 'compliance_validators.validate_nist_xr_compliance', 'R2'),
('ARVR', 'OPENXR', 'OpenXR Standards', 
 'Padrões OpenXR para realidade estendida', 
 'Global', 'compliance_validators.validate_openxr_compliance', 'R2');

-- =====================================================================
-- Funções para Validação Integrada Multi-Setor
-- =====================================================================

-- Função para validar conformidade para um setor específico
CREATE OR REPLACE FUNCTION compliance_integrator.validate_sector_compliance(
    tenant_id UUID,
    sector_id VARCHAR(30),
    regions VARCHAR(50)[] DEFAULT NULL
) RETURNS TABLE (
    regulation_id VARCHAR(30),
    regulation_name VARCHAR(100),
    region VARCHAR(50),
    is_compliant BOOLEAN,
    details TEXT
) AS $$
DECLARE
    reg RECORD;
    query_result RECORD;
    exec_query TEXT;
BEGIN
    -- Filtrar regulações por região, se especificado
    FOR reg IN 
        SELECT * FROM compliance_integrator.sector_regulations
        WHERE sector_regulations.sector_id = validate_sector_compliance.sector_id
        AND (regions IS NULL OR region = ANY(regions))
    LOOP
        -- Prepara e executa a consulta de validação dinâmica
        exec_query := 'SELECT * FROM ' || reg.validator_function || '($1)';
        
        FOR query_result IN EXECUTE exec_query USING tenant_id LOOP
            regulation_id := reg.regulation_id;
            regulation_name := reg.regulation_name;
            region := reg.region;
            is_compliant := query_result.is_compliant;
            details := query_result.details;
            RETURN NEXT;
        END LOOP;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Função para validar conformidade multi-setor
CREATE OR REPLACE FUNCTION compliance_integrator.validate_multi_sector_compliance(
    tenant_id UUID,
    sectors VARCHAR(30)[] DEFAULT NULL,
    regions VARCHAR(50)[] DEFAULT NULL
) RETURNS TABLE (
    sector_id VARCHAR(30),
    sector_name VARCHAR(100),
    regulation_id VARCHAR(30),
    regulation_name VARCHAR(100),
    region VARCHAR(50),
    is_compliant BOOLEAN,
    details TEXT
) AS $$
DECLARE
    sect RECORD;
    query_result RECORD;
BEGIN
    -- Define setores padrão se não especificados
    IF sectors IS NULL THEN
        SELECT array_agg(s.sector_id) INTO sectors 
        FROM compliance_integrator.sectors s
        WHERE s.enabled = TRUE
        AND s.sector_id != 'MULTI';
    END IF;
    
    -- Para cada setor, executa validação
    FOR sect IN 
        SELECT * FROM compliance_integrator.sectors
        WHERE sector_id = ANY(sectors)
    LOOP
        FOR query_result IN 
            SELECT * FROM compliance_integrator.validate_sector_compliance(tenant_id, sect.sector_id, regions)
        LOOP
            sector_id := sect.sector_id;
            sector_name := sect.sector_name;
            regulation_id := query_result.regulation_id;
            regulation_name := query_result.regulation_name;
            region := query_result.region;
            is_compliant := query_result.is_compliant;
            details := query_result.details;
            RETURN NEXT;
        END LOOP;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Função para gerar pontuação de conformidade multi-setor
CREATE OR REPLACE FUNCTION compliance_integrator.calculate_multi_sector_score(
    tenant_id UUID,
    sectors VARCHAR(30)[] DEFAULT NULL,
    regions VARCHAR(50)[] DEFAULT NULL
) RETURNS TABLE (
    sector_id VARCHAR(30),
    sector_name VARCHAR(100),
    compliance_score NUMERIC(5,2),
    total_requirements INTEGER,
    compliant_requirements INTEGER,
    compliance_percentage NUMERIC(5,2),
    irr VARCHAR(10)
) AS $$
DECLARE
    sect RECORD;
    total_count INTEGER := 0;
    compliant_count INTEGER := 0;
    score NUMERIC(5,2) := 0;
    percentage NUMERIC(5,2) := 0;
    result_irr VARCHAR(10);
    validation_result RECORD;
BEGIN
    -- Define setores padrão se não especificados
    IF sectors IS NULL THEN
        SELECT array_agg(s.sector_id) INTO sectors 
        FROM compliance_integrator.sectors s
        WHERE s.enabled = TRUE
        AND s.sector_id != 'MULTI';
    END IF;
    
    -- Para cada setor, calcula pontuação
    FOR sect IN 
        SELECT * FROM compliance_integrator.sectors
        WHERE sector_id = ANY(sectors)
    LOOP
        total_count := 0;
        compliant_count := 0;
        
        -- Calcular totais para este setor
        FOR validation_result IN 
            SELECT * FROM compliance_integrator.validate_sector_compliance(tenant_id, sect.sector_id, regions)
        LOOP
            total_count := total_count + 1;
            IF validation_result.is_compliant THEN
                compliant_count := compliant_count + 1;
            END IF;
        END LOOP;
        
        -- Calcular pontuação e percentual
        IF total_count > 0 THEN
            score := 4.0 * (compliant_count::NUMERIC / total_count);
            percentage := 100.0 * (compliant_count::NUMERIC / total_count);
        ELSE
            score := 0;
            percentage := 0;
        END IF;
        
        -- Determinar IRR
        IF percentage >= 95 THEN
            result_irr := 'R1';
        ELSIF percentage >= 85 THEN
            result_irr := 'R2';
        ELSIF percentage >= 70 THEN
            result_irr := 'R3';
        ELSE
            result_irr := 'R4';
        END IF;
        
        -- Retornar resultado
        sector_id := sect.sector_id;
        sector_name := sect.sector_name;
        compliance_score := score;
        total_requirements := total_count;
        compliant_requirements := compliant_count;
        compliance_percentage := percentage;
        irr := result_irr;
        RETURN NEXT;
    END LOOP;
    
    -- Calcular pontuação geral combinada
    total_count := 0;
    compliant_count := 0;
    
    -- Somar totais de todos os setores
    FOR validation_result IN 
        SELECT * FROM compliance_integrator.validate_multi_sector_compliance(tenant_id, sectors, regions)
    LOOP
        total_count := total_count + 1;
        IF validation_result.is_compliant THEN
            compliant_count := compliant_count + 1;
        END IF;
    END LOOP;
    
    -- Calcular pontuação e percentual geral
    IF total_count > 0 THEN
        score := 4.0 * (compliant_count::NUMERIC / total_count);
        percentage := 100.0 * (compliant_count::NUMERIC / total_count);
    ELSE
        score := 0;
        percentage := 0;
    END IF;
    
    -- Determinar IRR geral
    IF percentage >= 95 THEN
        result_irr := 'R1';
    ELSIF percentage >= 85 THEN
        result_irr := 'R2';
    ELSIF percentage >= 70 THEN
        result_irr := 'R3';
    ELSE
        result_irr := 'R4';
    END IF;
    
    -- Retornar resultado geral
    sector_id := 'OVERALL';
    sector_name := 'Todos os Setores';
    compliance_score := score;
    total_requirements := total_count;
    compliant_requirements := compliant_count;
    compliance_percentage := percentage;
    irr := result_irr;
    RETURN NEXT;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Funções para Geração e Exportação de Relatórios
-- =====================================================================

-- Função para gerar relatório de conformidade em formato JSON
CREATE OR REPLACE FUNCTION compliance_integrator.generate_compliance_report_json(
    tenant_id UUID,
    sectors VARCHAR(30)[] DEFAULT NULL,
    regions VARCHAR(50)[] DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    result JSONB;
    validation_results JSONB;
    score_results JSONB;
    timestamp TIMESTAMP WITH TIME ZONE := CURRENT_TIMESTAMP;
    report_id UUID := gen_random_uuid();
BEGIN
    -- Obter resultados de validação
    SELECT jsonb_agg(
        jsonb_build_object(
            'sector_id', sector_id,
            'sector_name', sector_name,
            'regulation_id', regulation_id,
            'regulation_name', regulation_name,
            'region', region,
            'is_compliant', is_compliant,
            'details', details
        )
    ) INTO validation_results
    FROM compliance_integrator.validate_multi_sector_compliance(tenant_id, sectors, regions);
    
    -- Obter resultados de pontuação
    SELECT jsonb_agg(
        jsonb_build_object(
            'sector_id', sector_id,
            'sector_name', sector_name,
            'compliance_score', compliance_score,
            'total_requirements', total_requirements,
            'compliant_requirements', compliant_requirements,
            'compliance_percentage', compliance_percentage,
            'irr', irr
        )
    ) INTO score_results
    FROM compliance_integrator.calculate_multi_sector_score(tenant_id, sectors, regions);
    
    -- Construir relatório
    result := jsonb_build_object(
        'report_id', report_id,
        'tenant_id', tenant_id,
        'timestamp', timestamp,
        'report_type', 'Compliance Validation',
        'sectors', COALESCE(sectors, ARRAY[]::VARCHAR[]),
        'regions', COALESCE(regions, ARRAY[]::VARCHAR[]),
        'validation_results', COALESCE(validation_results, '[]'::JSONB),
        'score_results', COALESCE(score_results, '[]'::JSONB)
    );
    
    -- Registrar histórico
    INSERT INTO compliance_integrator.validation_history (
        validation_id, tenant_id, validation_type, sectors, 
        regulations, score, irr, validation_date, summary
    )
    SELECT 
        report_id, 
        tenant_id, 
        'MULTI_SECTOR',
        sectors,
        validation_results,
        (SELECT compliance_score FROM compliance_integrator.calculate_multi_sector_score(tenant_id, sectors, regions) WHERE sector_id = 'OVERALL'),
        (SELECT irr FROM compliance_integrator.calculate_multi_sector_score(tenant_id, sectors, regions) WHERE sector_id = 'OVERALL'),
        timestamp,
        result;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- Função para exportar relatório em formato XML
CREATE OR REPLACE FUNCTION compliance_integrator.export_compliance_report_xml(
    tenant_id UUID,
    sectors VARCHAR(30)[] DEFAULT NULL,
    regions VARCHAR(50)[] DEFAULT NULL
) RETURNS TEXT AS $$
DECLARE
    json_report JSONB;
    xml_result TEXT;
BEGIN
    -- Obter relatório em JSON
    SELECT compliance_integrator.generate_compliance_report_json(tenant_id, sectors, regions) INTO json_report;
    
    -- Converter para XML (simplificado - em produção usar xmlelement e xmlagg para estrutura adequada)
    xml_result := '<?xml version="1.0" encoding="UTF-8"?>'
               || '<ComplianceReport>'
               || '<ReportID>' || json_report->>'report_id' || '</ReportID>'
               || '<TenantID>' || json_report->>'tenant_id' || '</TenantID>'
               || '<Timestamp>' || json_report->>'timestamp' || '</Timestamp>'
               || '<ReportType>' || json_report->>'report_type' || '</ReportType>';
               
    -- Em produção, implementar conversão completa de JSON para XML
    
    xml_result := xml_result || '</ComplianceReport>';
    
    RETURN xml_result;
END;
$$ LANGUAGE plpgsql;

-- Função para exportar relatório em formato CSV
CREATE OR REPLACE FUNCTION compliance_integrator.export_compliance_report_csv(
    tenant_id UUID,
    sectors VARCHAR(30)[] DEFAULT NULL,
    regions VARCHAR(50)[] DEFAULT NULL
) RETURNS TEXT AS $$
DECLARE
    csv_result TEXT := '';
    report_row RECORD;
BEGIN
    -- Cabeçalhos para resultados de validação
    csv_result := csv_result || 'Setor,Nome do Setor,Regulação,Nome da Regulação,Região,Conformidade,Detalhes' || E'\n';
    
    -- Dados de validação
    FOR report_row IN
        SELECT * FROM compliance_integrator.validate_multi_sector_compliance(tenant_id, sectors, regions)
    LOOP
        csv_result := csv_result || 
            report_row.sector_id || ',' ||
            report_row.sector_name || ',' ||
            report_row.regulation_id || ',' ||
            report_row.regulation_name || ',' ||
            report_row.region || ',' ||
            CASE WHEN report_row.is_compliant THEN 'Sim' ELSE 'Não' END || ',' ||
            '"' || REPLACE(report_row.details, '"', '""') || '"' || E'\n';
    END LOOP;
    
    -- Linha em branco
    csv_result := csv_result || E'\n';
    
    -- Cabeçalhos para pontuação
    csv_result := csv_result || 'Setor,Nome do Setor,Pontuação,Total de Requisitos,Requisitos Conformes,Percentual,IRR' || E'\n';
    
    -- Dados de pontuação
    FOR report_row IN
        SELECT * FROM compliance_integrator.calculate_multi_sector_score(tenant_id, sectors, regions)
    LOOP
        csv_result := csv_result || 
            report_row.sector_id || ',' ||
            report_row.sector_name || ',' ||
            report_row.compliance_score || ',' ||
            report_row.total_requirements || ',' ||
            report_row.compliant_requirements || ',' ||
            report_row.compliance_percentage || '%,' ||
            report_row.irr || E'\n';
    END LOOP;
    
    RETURN csv_result;
END;
$$ LANGUAGE plpgsql;

-- Função para gerenciar validações agendadas
CREATE OR REPLACE FUNCTION compliance_integrator.schedule_compliance_validation(
    tenant_id UUID,
    sectors VARCHAR(30)[] DEFAULT NULL,
    regions VARCHAR(50)[] DEFAULT NULL,
    schedule_type VARCHAR(20) DEFAULT 'DAILY',  -- DAILY, WEEKLY, MONTHLY, QUARTERLY
    notification_emails TEXT[] DEFAULT NULL,
    export_format VARCHAR(10) DEFAULT 'JSON'    -- JSON, XML, CSV
) RETURNS UUID AS $$
DECLARE
    config_id UUID := gen_random_uuid();
    schedule_config JSONB;
BEGIN
    -- Configurar agenda
    schedule_config := jsonb_build_object(
        'schedule_id', config_id,
        'schedule_type', schedule_type,
        'sectors', COALESCE(sectors, ARRAY[]::VARCHAR[]),
        'regions', COALESCE(regions, ARRAY[]::VARCHAR[]),
        'notification_emails', COALESCE(notification_emails, ARRAY[]::TEXT[]),
        'export_format', export_format,
        'is_active', TRUE,
        'last_run', NULL,
        'next_run', CASE 
            WHEN schedule_type = 'DAILY' THEN (CURRENT_DATE + INTERVAL '1 day')::TEXT
            WHEN schedule_type = 'WEEKLY' THEN (CURRENT_DATE + INTERVAL '1 week')::TEXT
            WHEN schedule_type = 'MONTHLY' THEN (CURRENT_DATE + INTERVAL '1 month')::TEXT
            WHEN schedule_type = 'QUARTERLY' THEN (CURRENT_DATE + INTERVAL '3 months')::TEXT
            ELSE NULL
        END
    );
    
    -- Inserir ou atualizar configuração
    INSERT INTO compliance_integrator.tenant_validator_config (
        tenant_id, active_sectors, report_schedule, enabled_validators, notification_settings
    )
    VALUES (
        tenant_id, 
        COALESCE(sectors, ARRAY[]::VARCHAR[]),
        schedule_config,
        jsonb_build_object('all_enabled', TRUE),
        jsonb_build_object('emails', COALESCE(notification_emails, ARRAY[]::TEXT[]))
    )
    ON CONFLICT (tenant_id) 
    DO UPDATE SET
        active_sectors = COALESCE(sectors, compliance_integrator.tenant_validator_config.active_sectors),
        report_schedule = schedule_config,
        notification_settings = jsonb_build_object('emails', COALESCE(notification_emails, ARRAY[]::TEXT[])),
        updated_at = CURRENT_TIMESTAMP;
        
    RETURN config_id;
END;
$$ LANGUAGE plpgsql;

-- Procedimento para executar validações agendadas
CREATE OR REPLACE PROCEDURE compliance_integrator.run_scheduled_validations()
LANGUAGE plpgsql
AS $$
DECLARE
    config RECORD;
    report_result JSONB;
    next_run TIMESTAMP WITH TIME ZONE;
BEGIN
    -- Buscar todas as configurações agendadas para hoje
    FOR config IN
        SELECT 
            tenant_id, 
            report_schedule->>'schedule_id' AS schedule_id,
            report_schedule->>'schedule_type' AS schedule_type,
            report_schedule->'sectors' AS sectors,
            report_schedule->'regions' AS regions,
            report_schedule->>'export_format' AS export_format
        FROM compliance_integrator.tenant_validator_config
        WHERE 
            report_schedule->>'is_active' = 'true'
            AND (report_schedule->>'next_run')::DATE <= CURRENT_DATE
    LOOP
        -- Executar validação
        SELECT compliance_integrator.generate_compliance_report_json(
            config.tenant_id,
            array_remove(ARRAY(SELECT jsonb_array_elements_text(config.sectors)), ''),
            array_remove(ARRAY(SELECT jsonb_array_elements_text(config.regions)), '')
        ) INTO report_result;
        
        -- Calcular próxima execução
        SELECT 
            CASE 
                WHEN config.schedule_type = 'DAILY' THEN CURRENT_DATE + INTERVAL '1 day'
                WHEN config.schedule_type = 'WEEKLY' THEN CURRENT_DATE + INTERVAL '1 week'
                WHEN config.schedule_type = 'MONTHLY' THEN CURRENT_DATE + INTERVAL '1 month'
                WHEN config.schedule_type = 'QUARTERLY' THEN CURRENT_DATE + INTERVAL '3 months'
                ELSE NULL
            END INTO next_run;
        
        -- Atualizar configuração
        UPDATE compliance_integrator.tenant_validator_config
        SET 
            report_schedule = jsonb_set(
                jsonb_set(
                    report_schedule,
                    '{last_run}',
                    to_jsonb(CURRENT_TIMESTAMP)
                ),
                '{next_run}',
                to_jsonb(next_run)
            ),
            updated_at = CURRENT_TIMESTAMP
        WHERE tenant_id = config.tenant_id;
        
        -- Aqui adicionaríamos lógica para enviar notificações por email
        -- Esta é uma implementação simplificada
    END LOOP;
END;
$$;

-- =====================================================================
-- Comentários de Documentação
-- =====================================================================

COMMENT ON SCHEMA compliance_integrator IS 'Esquema para o integrador de validadores de conformidade do INNOVABIZ';
COMMENT ON TABLE compliance_integrator.sectors IS 'Mapeamento de setores para validação de conformidade';
COMMENT ON TABLE compliance_integrator.sector_regulations IS 'Mapeamento de regulações por setor';
COMMENT ON TABLE compliance_integrator.tenant_validator_config IS 'Configuração de validação por tenant';
COMMENT ON TABLE compliance_integrator.validation_history IS 'Histórico de validações de conformidade';

COMMENT ON FUNCTION compliance_integrator.validate_sector_compliance IS 'Valida conformidade para um setor específico';
COMMENT ON FUNCTION compliance_integrator.validate_multi_sector_compliance IS 'Valida conformidade para múltiplos setores';
COMMENT ON FUNCTION compliance_integrator.calculate_multi_sector_score IS 'Calcula pontuação de conformidade multi-setor';
COMMENT ON FUNCTION compliance_integrator.generate_compliance_report_json IS 'Gera relatório de conformidade em formato JSON';
COMMENT ON FUNCTION compliance_integrator.export_compliance_report_xml IS 'Exporta relatório de conformidade em formato XML';
COMMENT ON FUNCTION compliance_integrator.export_compliance_report_csv IS 'Exporta relatório de conformidade em formato CSV';
COMMENT ON FUNCTION compliance_integrator.schedule_compliance_validation IS 'Configura validações de conformidade agendadas';
COMMENT ON PROCEDURE compliance_integrator.run_scheduled_validations IS 'Executa validações de conformidade agendadas';

-- =====================================================================
-- Fim do Script
-- =====================================================================
