-- =====================================================================
-- INNOVABIZ - Integração com Sistema de Gestão de Incidentes
-- Versão: 1.0.0
-- Data de Criação: 14/05/2025
-- Autor: INNOVABIZ DevOps
-- Descrição: Integração entre Validadores de Conformidade IAM e
--            Sistema de Gestão de Incidentes da plataforma
-- =====================================================================

-- Criação de schema para a integração com gestão de incidentes
CREATE SCHEMA IF NOT EXISTS compliance_incident;

-- =====================================================================
-- Tabelas de Mapeamento e Configuração
-- =====================================================================

-- Tabela de configuração de mapeamento de níveis IRR para severidade de incidentes
CREATE TABLE compliance_incident.irr_severity_mapping (
    irr_level VARCHAR(10) PRIMARY KEY,
    incident_severity VARCHAR(20) NOT NULL,
    incident_priority INTEGER NOT NULL,
    auto_create_incident BOOLEAN DEFAULT FALSE,
    sla_hours INTEGER,
    description TEXT
);

-- Tabela de configuração de integração por tenant
CREATE TABLE compliance_incident.tenant_integration_config (
    tenant_id UUID PRIMARY KEY,
    is_integration_enabled BOOLEAN DEFAULT TRUE,
    incident_assignment_group VARCHAR(100),
    auto_create_threshold VARCHAR(10) DEFAULT 'R3', -- IRR a partir do qual incidentes são criados
    integration_settings JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabela de histórico de incidentes criados
CREATE TABLE compliance_incident.incident_history (
    incident_id VARCHAR(50) PRIMARY KEY,
    tenant_id UUID NOT NULL,
    validation_id UUID REFERENCES compliance_integrator.validation_history(validation_id),
    irr_level VARCHAR(10) NOT NULL,
    incident_severity VARCHAR(20) NOT NULL,
    incident_priority INTEGER NOT NULL,
    incident_status VARCHAR(30) DEFAULT 'OPEN',
    creation_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    resolution_date TIMESTAMP WITH TIME ZONE,
    incident_details JSONB,
    compliance_data JSONB
);

-- =====================================================================
-- Dados Iniciais
-- =====================================================================

-- Inserção de dados de mapeamento IRR para severidade
INSERT INTO compliance_incident.irr_severity_mapping 
(irr_level, incident_severity, incident_priority, auto_create_incident, sla_hours, description) 
VALUES
('R1', 'BAIXA', 4, FALSE, 168, 'Risco residual muito baixo - não requer incidente'),
('R2', 'MÉDIA', 3, FALSE, 72, 'Risco residual baixo - incidente opcional'),
('R3', 'ALTA', 2, TRUE, 24, 'Risco residual moderado - requer incidente'),
('R4', 'CRÍTICA', 1, TRUE, 4, 'Risco residual elevado - incidente prioritário');

-- =====================================================================
-- Funções de Integração
-- =====================================================================

-- Função para criar incidente a partir de resultados de validação
CREATE OR REPLACE FUNCTION compliance_incident.create_incident_from_validation(
    validation_id UUID,
    tenant_id UUID
) RETURNS VARCHAR(50) AS $$
DECLARE
    incident_id VARCHAR(50);
    validation_data JSONB;
    irr_level VARCHAR(10);
    incident_severity VARCHAR(20);
    incident_priority INTEGER;
    incident_title VARCHAR(200);
    incident_details JSONB;
    integration_config JSONB;
    assignment_group VARCHAR(100);
BEGIN
    -- Obter configuração de integração do tenant
    SELECT 
        incident_assignment_group,
        integration_settings
    INTO
        assignment_group,
        integration_config
    FROM compliance_incident.tenant_integration_config
    WHERE tenant_id = create_incident_from_validation.tenant_id;
    
    IF NOT FOUND THEN
        assignment_group := 'Compliance_Team';
        integration_config := '{}'::JSONB;
    END IF;
    
    -- Obter dados da validação
    SELECT
        summary,
        irr
    INTO
        validation_data,
        irr_level
    FROM compliance_integrator.validation_history
    WHERE validation_id = create_incident_from_validation.validation_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Validação % não encontrada', validation_id;
    END IF;
    
    -- Obter mapeamento de severidade
    SELECT
        incident_severity,
        incident_priority
    INTO
        incident_severity,
        incident_priority
    FROM compliance_incident.irr_severity_mapping
    WHERE irr_level = compliance_incident.irr_severity_mapping.irr_level;
    
    -- Gerar detalhes do incidente
    incident_title := 'Não-conformidade IAM: IRR ' || irr_level;
    incident_details := jsonb_build_object(
        'description', 'Incidente de não-conformidade IAM detectado pelos validadores automatizados',
        'affected_sectors', validation_data->'sectors',
        'validation_results', validation_data->'validation_results',
        'score_results', validation_data->'score_results',
        'irr_level', irr_level,
        'incident_source', 'IAM_Compliance_Validator',
        'assignment_group', assignment_group,
        'business_service', 'IAM'
    );
    
    -- Gerar ID de incidente (em um sistema real, seria retornado pelo sistema de incidentes)
    incident_id := 'INC' || to_char(CURRENT_TIMESTAMP, 'YYYYMMDD') || '-' || 
                   substr(md5(validation_id::TEXT), 1, 6);
    
    -- Registrar o incidente no histórico
    INSERT INTO compliance_incident.incident_history (
        incident_id, 
        tenant_id, 
        validation_id, 
        irr_level, 
        incident_severity, 
        incident_priority, 
        incident_status, 
        incident_details,
        compliance_data
    ) VALUES (
        incident_id,
        tenant_id,
        validation_id,
        irr_level,
        incident_severity,
        incident_priority,
        'OPEN',
        incident_details,
        validation_data
    );
    
    RETURN incident_id;
END;
$$ LANGUAGE plpgsql;

-- Função para processar resultados de validação e criar incidentes quando necessário
CREATE OR REPLACE FUNCTION compliance_incident.process_validation_results(
    validation_id UUID
) RETURNS VARCHAR(50) AS $$
DECLARE
    tenant_id UUID;
    irr_level VARCHAR(10);
    auto_create_threshold VARCHAR(10);
    should_create_incident BOOLEAN := FALSE;
    incident_id VARCHAR(50) := NULL;
BEGIN
    -- Obter dados da validação
    SELECT
        vh.tenant_id,
        vh.irr
    INTO
        tenant_id,
        irr_level
    FROM compliance_integrator.validation_history vh
    WHERE vh.validation_id = process_validation_results.validation_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Validação % não encontrada', validation_id;
    END IF;
    
    -- Verificar se já existe um incidente para esta validação
    SELECT i.incident_id INTO incident_id
    FROM compliance_incident.incident_history i
    WHERE i.validation_id = process_validation_results.validation_id;
    
    IF FOUND THEN
        RETURN incident_id; -- Incidente já existe
    END IF;
    
    -- Verificar configuração do tenant
    SELECT 
        tic.auto_create_threshold,
        tic.is_integration_enabled
    INTO
        auto_create_threshold,
        should_create_incident
    FROM compliance_incident.tenant_integration_config tic
    WHERE tic.tenant_id = tenant_id;
    
    IF NOT FOUND THEN
        -- Default: R3 e R4 geram incidentes automaticamente
        auto_create_threshold := 'R3';
        should_create_incident := TRUE;
    END IF;
    
    -- Verificar se deve criar incidente baseado no IRR
    IF should_create_incident THEN
        IF (irr_level = 'R4') OR 
           (irr_level = 'R3' AND auto_create_threshold IN ('R3', 'R2', 'R1')) OR
           (irr_level = 'R2' AND auto_create_threshold IN ('R2', 'R1')) OR
           (irr_level = 'R1' AND auto_create_threshold = 'R1') THEN
            
            -- Criar incidente
            incident_id := compliance_incident.create_incident_from_validation(
                validation_id,
                tenant_id
            );
        END IF;
    END IF;
    
    RETURN incident_id;
END;
$$ LANGUAGE plpgsql;

-- Trigger para processar automaticamente novos resultados de validação
CREATE OR REPLACE FUNCTION compliance_incident.validation_result_trigger()
RETURNS TRIGGER AS $$
DECLARE
    incident_id VARCHAR(50);
BEGIN
    -- Apenas processar validações concluídas com IRR
    IF NEW.irr IS NOT NULL THEN
        -- Tenta criar um incidente se necessário
        incident_id := compliance_incident.process_validation_results(NEW.validation_id);
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Criação do trigger
DROP TRIGGER IF EXISTS validation_result_incident_trigger 
ON compliance_integrator.validation_history;

CREATE TRIGGER validation_result_incident_trigger
AFTER INSERT OR UPDATE OF irr ON compliance_integrator.validation_history
FOR EACH ROW EXECUTE FUNCTION compliance_incident.validation_result_trigger();

-- =====================================================================
-- Funções de Administração e Consulta
-- =====================================================================

-- Função para configurar integração por tenant
CREATE OR REPLACE FUNCTION compliance_incident.configure_tenant_integration(
    tenant_id UUID,
    is_enabled BOOLEAN DEFAULT TRUE,
    assignment_group VARCHAR(100) DEFAULT 'Compliance_Team',
    auto_create_threshold VARCHAR(10) DEFAULT 'R3',
    settings JSONB DEFAULT NULL
) RETURNS BOOLEAN AS $$
BEGIN
    INSERT INTO compliance_incident.tenant_integration_config (
        tenant_id,
        is_integration_enabled,
        incident_assignment_group,
        auto_create_threshold,
        integration_settings
    ) VALUES (
        tenant_id,
        is_enabled,
        assignment_group,
        auto_create_threshold,
        COALESCE(settings, '{}'::JSONB)
    )
    ON CONFLICT (tenant_id)
    DO UPDATE SET
        is_integration_enabled = EXCLUDED.is_integration_enabled,
        incident_assignment_group = EXCLUDED.incident_assignment_group,
        auto_create_threshold = EXCLUDED.auto_create_threshold,
        integration_settings = EXCLUDED.integration_settings,
        updated_at = CURRENT_TIMESTAMP;
        
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- Função para consultar incidentes ativos por tenant
CREATE OR REPLACE FUNCTION compliance_incident.get_active_incidents(
    tenant_id UUID
) RETURNS TABLE (
    incident_id VARCHAR(50),
    irr_level VARCHAR(10),
    incident_severity VARCHAR(20),
    incident_priority INTEGER,
    incident_status VARCHAR(30),
    creation_date TIMESTAMP WITH TIME ZONE,
    age_hours NUMERIC,
    sla_hours INTEGER,
    sla_status VARCHAR(20)
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        ih.incident_id,
        ih.irr_level,
        ih.incident_severity,
        ih.incident_priority,
        ih.incident_status,
        ih.creation_date,
        EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - ih.creation_date))/3600 AS age_hours,
        ism.sla_hours,
        CASE
            WHEN ih.incident_status = 'RESOLVED' THEN 'FECHADO'
            WHEN EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - ih.creation_date))/3600 > ism.sla_hours THEN 'VIOLADO'
            WHEN EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - ih.creation_date))/3600 > (ism.sla_hours * 0.75) THEN 'EM_RISCO'
            ELSE 'DENTRO_DO_SLA'
        END AS sla_status
    FROM 
        compliance_incident.incident_history ih
    JOIN 
        compliance_incident.irr_severity_mapping ism ON ih.irr_level = ism.irr_level
    WHERE 
        ih.tenant_id = get_active_incidents.tenant_id AND
        ih.incident_status != 'RESOLVED'
    ORDER BY 
        ih.incident_priority, ih.creation_date;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Comentários de Documentação
-- =====================================================================

COMMENT ON SCHEMA compliance_incident IS 'Esquema para integração entre validadores de conformidade e gestão de incidentes';

COMMENT ON TABLE compliance_incident.irr_severity_mapping IS 'Mapeamento de níveis IRR para severidade de incidentes';
COMMENT ON TABLE compliance_incident.tenant_integration_config IS 'Configuração da integração por tenant';
COMMENT ON TABLE compliance_incident.incident_history IS 'Histórico de incidentes criados a partir de validações';

COMMENT ON FUNCTION compliance_incident.create_incident_from_validation IS 'Cria um incidente a partir de um resultado de validação';
COMMENT ON FUNCTION compliance_incident.process_validation_results IS 'Processa resultados de validação para decidir sobre criação de incidentes';
COMMENT ON FUNCTION compliance_incident.validation_result_trigger IS 'Trigger para processamento automático de resultados de validação';
COMMENT ON FUNCTION compliance_incident.configure_tenant_integration IS 'Configura integração por tenant';
COMMENT ON FUNCTION compliance_incident.get_active_incidents IS 'Lista incidentes ativos por tenant';

-- =====================================================================
-- Fim do Script
-- =====================================================================
