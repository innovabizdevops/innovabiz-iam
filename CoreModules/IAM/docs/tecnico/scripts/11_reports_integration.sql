-- =====================================================================
-- INNOVABIZ - Integração com Sistema de Gestão de Relatórios
-- Versão: 1.0.0
-- Data de Criação: 15/05/2025
-- Autor: INNOVABIZ DevOps
-- Descrição: Integração entre Validadores de Conformidade IAM e
--            Sistema de Gestão de Relatórios da plataforma
-- =====================================================================

-- Criação de schema para a integração com gestão de relatórios
CREATE SCHEMA IF NOT EXISTS compliance_reports;

-- =====================================================================
-- Tabelas de Mapeamento e Configuração
-- =====================================================================

-- Tabela de templates de relatórios
CREATE TABLE compliance_reports.report_templates (
    template_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_name VARCHAR(100) NOT NULL,
    description TEXT,
    template_type VARCHAR(50) NOT NULL,
    template_format VARCHAR(20) NOT NULL, -- PDF, EXCEL, CSV, JSON, HTML
    template_structure JSONB NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100)
);

-- Tabela de configuração de relatórios por tenant
CREATE TABLE compliance_reports.tenant_report_config (
    tenant_id UUID PRIMARY KEY,
    default_template_id UUID REFERENCES compliance_reports.report_templates(template_id),
    default_format VARCHAR(20) DEFAULT 'PDF',
    default_language VARCHAR(10) DEFAULT 'pt-BR',
    report_header JSONB,
    report_footer JSONB,
    logo_url TEXT,
    custom_styles JSONB,
    kafka_topic VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabela de histórico de relatórios gerados
CREATE TABLE compliance_reports.report_history (
    report_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    validation_id UUID REFERENCES compliance_integrator.validation_history(validation_id),
    template_id UUID REFERENCES compliance_reports.report_templates(template_id),
    report_format VARCHAR(20) NOT NULL,
    report_language VARCHAR(10) NOT NULL,
    generation_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, PROCESSING, COMPLETED, FAILED
    report_url TEXT,
    kafka_message_id VARCHAR(100),
    generation_details JSONB
);

-- =====================================================================
-- Funções Básicas de Integração
-- =====================================================================

-- Função para registrar template de relatório
CREATE OR REPLACE FUNCTION compliance_reports.register_report_template(
    template_name VARCHAR(100),
    description TEXT,
    template_type VARCHAR(50),
    template_format VARCHAR(20),
    template_structure JSONB,
    created_by VARCHAR(100) DEFAULT 'system'
) RETURNS UUID AS $$
DECLARE
    template_id UUID;
BEGIN
    INSERT INTO compliance_reports.report_templates (
        template_name,
        description,
        template_type,
        template_format,
        template_structure,
        created_by
    ) VALUES (
        template_name,
        description,
        template_type,
        template_format,
        template_structure,
        created_by
    ) RETURNING template_id INTO template_id;
    
    RETURN template_id;
END;
$$ LANGUAGE plpgsql;

-- Função para configurar relatórios para tenant
CREATE OR REPLACE FUNCTION compliance_reports.configure_tenant_reports(
    tenant_id UUID,
    default_template_id UUID,
    default_format VARCHAR(20) DEFAULT 'PDF',
    default_language VARCHAR(10) DEFAULT 'pt-BR',
    report_header JSONB DEFAULT NULL,
    report_footer JSONB DEFAULT NULL,
    logo_url TEXT DEFAULT NULL,
    kafka_topic VARCHAR(100) DEFAULT 'compliance-reports'
) RETURNS BOOLEAN AS $$
BEGIN
    INSERT INTO compliance_reports.tenant_report_config (
        tenant_id,
        default_template_id,
        default_format,
        default_language,
        report_header,
        report_footer,
        logo_url,
        kafka_topic
    ) VALUES (
        tenant_id,
        default_template_id,
        default_format,
        default_language,
        COALESCE(report_header, '{}'::JSONB),
        COALESCE(report_footer, '{}'::JSONB),
        logo_url,
        kafka_topic
    )
    ON CONFLICT (tenant_id)
    DO UPDATE SET
        default_template_id = EXCLUDED.default_template_id,
        default_format = EXCLUDED.default_format,
        default_language = EXCLUDED.default_language,
        report_header = EXCLUDED.report_header,
        report_footer = EXCLUDED.report_footer,
        logo_url = EXCLUDED.logo_url,
        kafka_topic = EXCLUDED.kafka_topic,
        updated_at = CURRENT_TIMESTAMP;
        
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Funções de Geração de Relatórios
-- =====================================================================

-- Função para gerar relatório a partir de resultados de validação
CREATE OR REPLACE FUNCTION compliance_reports.generate_report_from_validation(
    validation_id UUID,
    tenant_id UUID,
    template_id UUID DEFAULT NULL,
    report_format VARCHAR(20) DEFAULT NULL,
    report_language VARCHAR(10) DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    report_id UUID;
    tenant_config RECORD;
    validation_data JSONB;
    kafka_topic VARCHAR(100);
    kafka_message JSONB;
    message_id VARCHAR(100);
    used_template_id UUID;
    used_format VARCHAR(20);
    used_language VARCHAR(10);
BEGIN
    -- Obter configuração do tenant
    SELECT * INTO tenant_config
    FROM compliance_reports.tenant_report_config
    WHERE tenant_id = generate_report_from_validation.tenant_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Tenant % não encontrado ou não configurado para relatórios', tenant_id;
    END IF;
    
    -- Definir valores padrão se não especificados
    used_template_id := COALESCE(template_id, tenant_config.default_template_id);
    used_format := COALESCE(report_format, tenant_config.default_format);
    used_language := COALESCE(report_language, tenant_config.default_language);
    kafka_topic := tenant_config.kafka_topic;
    
    -- Obter dados da validação
    SELECT summary INTO validation_data
    FROM compliance_integrator.validation_history
    WHERE validation_id = generate_report_from_validation.validation_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Validação % não encontrada', validation_id;
    END IF;
    
    -- Gerar ID de relatório
    report_id := gen_random_uuid();
    
    -- Construir mensagem Kafka para processamento assíncrono
    kafka_message := jsonb_build_object(
        'report_id', report_id,
        'tenant_id', tenant_id,
        'validation_id', validation_id,
        'template_id', used_template_id,
        'report_format', used_format,
        'report_language', used_language,
        'validation_data', validation_data,
        'report_header', tenant_config.report_header,
        'report_footer', tenant_config.report_footer,
        'logo_url', tenant_config.logo_url,
        'timestamp', CURRENT_TIMESTAMP
    );
    
    -- Em um ambiente real, enviamos para o Kafka. Aqui simulamos o ID da mensagem
    message_id := 'msg-' || substr(md5(report_id::TEXT), 1, 16);
    
    -- Registrar no histórico
    INSERT INTO compliance_reports.report_history (
        report_id,
        tenant_id,
        validation_id,
        template_id,
        report_format,
        report_language,
        status,
        kafka_message_id,
        generation_details
    ) VALUES (
        report_id,
        tenant_id,
        validation_id,
        used_template_id,
        used_format,
        used_language,
        'PENDING',
        message_id,
        jsonb_build_object(
            'kafka_topic', kafka_topic,
            'kafka_message', kafka_message
        )
    );
    
    -- Comentário: Em um sistema real, a mensagem seria enviada para Kafka:
    -- PERFORM pg_notify('compliance_kafka_producer', kafka_message::TEXT);
    
    RETURN report_id;
END;
$$ LANGUAGE plpgsql;

-- Função para atualizar status de relatório (simulando consumidor Kafka)
CREATE OR REPLACE FUNCTION compliance_reports.update_report_status(
    report_id UUID,
    status VARCHAR(20),
    report_url TEXT DEFAULT NULL,
    details JSONB DEFAULT NULL
) RETURNS BOOLEAN AS $$
BEGIN
    UPDATE compliance_reports.report_history
    SET 
        status = update_report_status.status,
        report_url = COALESCE(update_report_status.report_url, report_history.report_url),
        generation_details = COALESCE(report_history.generation_details, '{}'::JSONB) || 
                             COALESCE(update_report_status.details, '{}'::JSONB)
    WHERE report_id = update_report_status.report_id;
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Templates Padrão de Relatórios
-- =====================================================================

-- Criação de templates padrão para cada framework regulatório
DO $$
DECLARE
    healthcare_template_id UUID;
    financial_template_id UUID;
    government_template_id UUID;
    arvr_template_id UUID;
    consolidated_template_id UUID;
BEGIN
    -- Template para setor de saúde
    healthcare_template_id := compliance_reports.register_report_template(
        'Relatório de Conformidade - Setor de Saúde',
        'Template padrão para relatórios de conformidade do setor de saúde (HIPAA, GDPR, LGPD)',
        'HEALTHCARE',
        'PDF',
        jsonb_build_object(
            'sections', jsonb_build_array(
                jsonb_build_object(
                    'title', 'Resumo Executivo',
                    'content_type', 'summary',
                    'data_source', 'validation_results.overall'
                ),
                jsonb_build_object(
                    'title', 'Conformidade HIPAA',
                    'content_type', 'regulations',
                    'data_source', 'validation_results.hipaa'
                ),
                jsonb_build_object(
                    'title', 'Conformidade GDPR para Saúde',
                    'content_type', 'regulations',
                    'data_source', 'validation_results.gdpr'
                ),
                jsonb_build_object(
                    'title', 'Conformidade LGPD para Saúde',
                    'content_type', 'regulations',
                    'data_source', 'validation_results.lgpd'
                ),
                jsonb_build_object(
                    'title', 'Recomendações',
                    'content_type', 'recommendations',
                    'data_source', 'recommendations'
                )
            ),
            'charts', jsonb_build_array(
                jsonb_build_object(
                    'type', 'donut',
                    'title', 'Pontuação de Conformidade',
                    'data_source', 'compliance_score'
                ),
                jsonb_build_object(
                    'type', 'bar',
                    'title', 'Conformidade por Regulação',
                    'data_source', 'compliance_by_regulation'
                )
            ),
            'version', '1.0.0'
        ),
        'system'
    );
    
    -- Template para setor financeiro
    financial_template_id := compliance_reports.register_report_template(
        'Relatório de Conformidade - Setor Financeiro',
        'Template padrão para relatórios de conformidade do setor financeiro (PSD2, Open Banking)',
        'FINANCIAL',
        'PDF',
        jsonb_build_object(
            'sections', jsonb_build_array(
                jsonb_build_object(
                    'title', 'Resumo Executivo',
                    'content_type', 'summary',
                    'data_source', 'validation_results.overall'
                ),
                jsonb_build_object(
                    'title', 'Conformidade PSD2',
                    'content_type', 'regulations',
                    'data_source', 'validation_results.psd2'
                ),
                jsonb_build_object(
                    'title', 'Conformidade Open Banking',
                    'content_type', 'regulations',
                    'data_source', 'validation_results.openbanking'
                ),
                jsonb_build_object(
                    'title', 'Recomendações',
                    'content_type', 'recommendations',
                    'data_source', 'recommendations'
                )
            ),
            'charts', jsonb_build_array(
                jsonb_build_object(
                    'type', 'donut',
                    'title', 'Pontuação de Conformidade',
                    'data_source', 'compliance_score'
                ),
                jsonb_build_object(
                    'type', 'bar',
                    'title', 'Conformidade por Regulação',
                    'data_source', 'compliance_by_regulation'
                )
            ),
            'version', '1.0.0'
        ),
        'system'
    );
    
    -- Template consolidado multi-setor
    consolidated_template_id := compliance_reports.register_report_template(
        'Relatório de Conformidade Consolidado',
        'Template consolidado para relatórios de conformidade multi-setor',
        'CONSOLIDATED',
        'PDF',
        jsonb_build_object(
            'sections', jsonb_build_array(
                jsonb_build_object(
                    'title', 'Resumo Executivo',
                    'content_type', 'summary',
                    'data_source', 'validation_results.overall'
                ),
                jsonb_build_object(
                    'title', 'Conformidade por Setor',
                    'content_type', 'sectors',
                    'data_source', 'validation_results.sectors'
                ),
                jsonb_build_object(
                    'title', 'Conformidade por Regulação',
                    'content_type', 'regulations',
                    'data_source', 'validation_results.regulations'
                ),
                jsonb_build_object(
                    'title', 'Análise de Riscos',
                    'content_type', 'risks',
                    'data_source', 'risks'
                ),
                jsonb_build_object(
                    'title', 'Recomendações',
                    'content_type', 'recommendations',
                    'data_source', 'recommendations'
                )
            ),
            'charts', jsonb_build_array(
                jsonb_build_object(
                    'type', 'donut',
                    'title', 'Pontuação de Conformidade',
                    'data_source', 'compliance_score'
                ),
                jsonb_build_object(
                    'type', 'bar',
                    'title', 'Conformidade por Setor',
                    'data_source', 'compliance_by_sector'
                ),
                jsonb_build_object(
                    'type', 'radar',
                    'title', 'Perfil de Conformidade',
                    'data_source', 'compliance_profile'
                )
            ),
            'version', '1.0.0'
        ),
        'system'
    );
END;
$$;

-- =====================================================================
-- Trigger para Processamento Automático
-- =====================================================================

-- Função de trigger para geração automática de relatórios após validação
CREATE OR REPLACE FUNCTION compliance_reports.validation_report_trigger()
RETURNS TRIGGER AS $$
DECLARE
    tenant_config RECORD;
    report_id UUID;
BEGIN
    -- Verificar se existe configuração de relatório para o tenant
    SELECT * INTO tenant_config
    FROM compliance_reports.tenant_report_config
    WHERE tenant_id = NEW.tenant_id;
    
    IF FOUND THEN
        -- Gerar relatório automaticamente
        report_id := compliance_reports.generate_report_from_validation(
            NEW.validation_id,
            NEW.tenant_id
        );
        
        -- Processar relatório em segundo plano (comentado - em um sistema real seria feito via Kafka)
        -- PERFORM pg_notify('compliance_report_processor', jsonb_build_object('report_id', report_id)::TEXT);
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Criação do trigger
DROP TRIGGER IF EXISTS validation_report_trigger 
ON compliance_integrator.validation_history;

CREATE TRIGGER validation_report_trigger
AFTER INSERT ON compliance_integrator.validation_history
FOR EACH ROW EXECUTE FUNCTION compliance_reports.validation_report_trigger();

-- =====================================================================
-- Funções de Consulta de Relatórios
-- =====================================================================

-- Função para consultar relatórios por tenant
CREATE OR REPLACE FUNCTION compliance_reports.get_tenant_reports(
    tenant_id UUID,
    status VARCHAR(20) DEFAULT NULL,
    start_date TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    end_date TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    limit_count INTEGER DEFAULT 100
) RETURNS TABLE (
    report_id UUID,
    validation_id UUID,
    template_name VARCHAR(100),
    report_format VARCHAR(20),
    report_language VARCHAR(10),
    generation_date TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20),
    report_url TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        rh.report_id,
        rh.validation_id,
        rt.template_name,
        rh.report_format,
        rh.report_language,
        rh.generation_date,
        rh.status,
        rh.report_url
    FROM 
        compliance_reports.report_history rh
    JOIN 
        compliance_reports.report_templates rt ON rh.template_id = rt.template_id
    WHERE 
        rh.tenant_id = get_tenant_reports.tenant_id
        AND (get_tenant_reports.status IS NULL OR rh.status = get_tenant_reports.status)
        AND (get_tenant_reports.start_date IS NULL OR rh.generation_date >= get_tenant_reports.start_date)
        AND (get_tenant_reports.end_date IS NULL OR rh.generation_date <= get_tenant_reports.end_date)
    ORDER BY 
        rh.generation_date DESC
    LIMIT 
        get_tenant_reports.limit_count;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Comentários de Documentação
-- =====================================================================

COMMENT ON SCHEMA compliance_reports IS 'Schema para integração entre validadores de conformidade e sistema de relatórios';

COMMENT ON TABLE compliance_reports.report_templates IS 'Templates para relatórios de conformidade';
COMMENT ON TABLE compliance_reports.tenant_report_config IS 'Configuração de relatórios por tenant';
COMMENT ON TABLE compliance_reports.report_history IS 'Histórico de relatórios gerados';

COMMENT ON FUNCTION compliance_reports.register_report_template IS 'Registra um novo template de relatório';
COMMENT ON FUNCTION compliance_reports.configure_tenant_reports IS 'Configura relatórios para um tenant';
COMMENT ON FUNCTION compliance_reports.generate_report_from_validation IS 'Gera relatório a partir de resultados de validação';
COMMENT ON FUNCTION compliance_reports.update_report_status IS 'Atualiza o status de um relatório';
COMMENT ON FUNCTION compliance_reports.validation_report_trigger IS 'Trigger para geração automática de relatórios após validação';
COMMENT ON FUNCTION compliance_reports.get_tenant_reports IS 'Consulta relatórios por tenant';

-- =====================================================================
-- Fim do Script
-- =====================================================================
