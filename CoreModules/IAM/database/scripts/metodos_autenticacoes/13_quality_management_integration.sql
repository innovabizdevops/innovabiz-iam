-- =====================================================================
-- INNOVABIZ - Integração com Sistema de Gestão da Qualidade e Conformidade
-- Versão: 1.0.0
-- Data de Criação: 15/05/2025
-- Autor: INNOVABIZ DevOps
-- Descrição: Integração entre Validadores de Conformidade IAM e
--            Sistema de Gestão da Qualidade e Conformidade da plataforma
-- =====================================================================

-- Criação de schema para a integração com gestão da qualidade
CREATE SCHEMA IF NOT EXISTS compliance_quality;

-- =====================================================================
-- Tabelas de Mapeamento e Configuração
-- =====================================================================

-- Tabela de mapeamento entre validadores e padrões de qualidade
CREATE TABLE compliance_quality.quality_standard_mapping (
    mapping_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    validator_id VARCHAR(100) NOT NULL,
    validator_name VARCHAR(255) NOT NULL,
    quality_standard VARCHAR(100) NOT NULL,
    process_area VARCHAR(255) NOT NULL,
    standard_clause VARCHAR(100),
    impact_level VARCHAR(20) NOT NULL, -- CRITICAL, MAJOR, MINOR, OBSERVATION
    mapping_source VARCHAR(50) DEFAULT 'MANUAL', -- MANUAL, AUTOMATED, AI_SUGGESTED
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(validator_id, quality_standard, process_area)
);

-- Tabela de matriz de impacto em processos organizacionais
CREATE TABLE compliance_quality.process_impact_matrix (
    impact_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    process_name VARCHAR(255) NOT NULL,
    process_code VARCHAR(100) NOT NULL,
    process_owner VARCHAR(255),
    criticality_level VARCHAR(20) NOT NULL, -- CRITICAL, HIGH, MEDIUM, LOW
    related_standards JSONB, -- Lista de padrões de qualidade relacionados
    impact_areas JSONB, -- Áreas impactadas pelo processo
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, process_code)
);

-- Tabela de configuração de qualidade por tenant
CREATE TABLE compliance_quality.tenant_quality_config (
    tenant_id UUID PRIMARY KEY,
    applicable_standards JSONB NOT NULL, -- Lista de padrões aplicáveis
    quality_system_name VARCHAR(255) DEFAULT 'Quality Management System',
    certification_status JSONB, -- Status de certificação para cada padrão
    approval_workflow VARCHAR(50) DEFAULT 'STANDARD', -- STANDARD, HEAD_OF_QUALITY, COMMITTEE
    document_template_id UUID,
    audit_cycle JSONB, -- Ciclo de auditoria (frequência, tipo)
    custom_settings JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabela de registro de ações corretivas e preventivas
CREATE TABLE compliance_quality.corrective_action_register (
    action_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    validation_id UUID REFERENCES compliance_integrator.validation_history(validation_id),
    action_type VARCHAR(20) NOT NULL, -- CORRECTIVE, PREVENTIVE, IMPROVEMENT
    action_title VARCHAR(255) NOT NULL,
    action_description TEXT NOT NULL,
    non_conformity_description TEXT,
    quality_standard VARCHAR(100) NOT NULL,
    process_area VARCHAR(255) NOT NULL,
    impact_level VARCHAR(20) NOT NULL, -- CRITICAL, MAJOR, MINOR, OBSERVATION
    root_cause TEXT,
    root_cause_analysis_method VARCHAR(50), -- 5WHY, ISHIKAWA, PARETO, FMEA
    assigned_to VARCHAR(255),
    assigned_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deadline TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) DEFAULT 'CREATED', -- CREATED, ASSIGNED, IN_PROGRESS, IMPLEMENTED, VERIFIED, CLOSED, REJECTED
    effectiveness_verification TEXT,
    effectiveness_status VARCHAR(20), -- EFFECTIVE, PARTIALLY_EFFECTIVE, NOT_EFFECTIVE
    closure_date TIMESTAMP WITH TIME ZONE,
    closure_comments TEXT,
    closure_evidence JSONB,
    related_incident_id UUID,
    related_risk_id UUID,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabela de histórico de ações corretivas
CREATE TABLE compliance_quality.corrective_action_history (
    history_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action_id UUID REFERENCES compliance_quality.corrective_action_register(action_id),
    status_from VARCHAR(20),
    status_to VARCHAR(20) NOT NULL,
    changed_by VARCHAR(255),
    change_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    comments TEXT,
    attachments JSONB
);

-- Tabela de métricas e KPIs de qualidade
CREATE TABLE compliance_quality.quality_metrics (
    metric_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    metric_name VARCHAR(255) NOT NULL,
    metric_code VARCHAR(100) NOT NULL,
    metric_category VARCHAR(100) NOT NULL, -- COMPLIANCE, PROCESS, PERFORMANCE, CUSTOMER
    description TEXT,
    calculation_formula TEXT,
    current_value NUMERIC(10,2),
    target_value NUMERIC(10,2),
    threshold_warning NUMERIC(10,2),
    threshold_critical NUMERIC(10,2),
    unit VARCHAR(50),
    trend VARCHAR(20), -- IMPROVING, STABLE, DETERIORATING
    last_updated TIMESTAMP WITH TIME ZONE,
    measurement_period VARCHAR(50), -- DAILY, WEEKLY, MONTHLY, QUARTERLY
    responsible_person VARCHAR(255),
    data_source VARCHAR(100), -- VALIDATORS, INCIDENTS, RISKS, MANUAL
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, metric_code)
);

-- Tabela de histórico de métricas de qualidade
CREATE TABLE compliance_quality.quality_metrics_history (
    history_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_id UUID REFERENCES compliance_quality.quality_metrics(metric_id),
    value_date TIMESTAMP WITH TIME ZONE NOT NULL,
    metric_value NUMERIC(10,2) NOT NULL,
    calculated_by VARCHAR(255),
    notes TEXT
);

-- =====================================================================
-- Funções de Configuração e Mapeamento
-- =====================================================================

-- Função para registrar mapeamento entre validador e padrão de qualidade
CREATE OR REPLACE FUNCTION compliance_quality.register_quality_standard_mapping(
    validator_id VARCHAR(100),
    validator_name VARCHAR(255),
    quality_standard VARCHAR(100),
    process_area VARCHAR(255),
    standard_clause VARCHAR(100) DEFAULT NULL,
    impact_level VARCHAR(20) DEFAULT 'MAJOR',
    mapping_source VARCHAR(50) DEFAULT 'MANUAL'
) RETURNS UUID AS $$
DECLARE
    mapping_id UUID;
BEGIN
    INSERT INTO compliance_quality.quality_standard_mapping (
        validator_id,
        validator_name,
        quality_standard,
        process_area,
        standard_clause,
        impact_level,
        mapping_source
    ) VALUES (
        validator_id,
        validator_name,
        quality_standard,
        process_area,
        standard_clause,
        impact_level,
        mapping_source
    )
    ON CONFLICT (validator_id, quality_standard, process_area)
    DO UPDATE SET
        validator_name = EXCLUDED.validator_name,
        standard_clause = EXCLUDED.standard_clause,
        impact_level = EXCLUDED.impact_level,
        mapping_source = EXCLUDED.mapping_source,
        is_active = TRUE,
        updated_at = CURRENT_TIMESTAMP
    RETURNING mapping_id INTO mapping_id;
    
    RETURN mapping_id;
END;
$$ LANGUAGE plpgsql;

-- Função para registrar matriz de impacto em processos
CREATE OR REPLACE FUNCTION compliance_quality.register_process_impact(
    tenant_id UUID,
    process_name VARCHAR(255),
    process_code VARCHAR(100),
    process_owner VARCHAR(255) DEFAULT NULL,
    criticality_level VARCHAR(20) DEFAULT 'MEDIUM',
    related_standards JSONB DEFAULT NULL,
    impact_areas JSONB DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    impact_id UUID;
BEGIN
    INSERT INTO compliance_quality.process_impact_matrix (
        tenant_id,
        process_name,
        process_code,
        process_owner,
        criticality_level,
        related_standards,
        impact_areas
    ) VALUES (
        tenant_id,
        process_name,
        process_code,
        process_owner,
        criticality_level,
        COALESCE(related_standards, '[]'::JSONB),
        COALESCE(impact_areas, '[]'::JSONB)
    )
    ON CONFLICT (tenant_id, process_code)
    DO UPDATE SET
        process_name = EXCLUDED.process_name,
        process_owner = EXCLUDED.process_owner,
        criticality_level = EXCLUDED.criticality_level,
        related_standards = EXCLUDED.related_standards,
        impact_areas = EXCLUDED.impact_areas,
        updated_at = CURRENT_TIMESTAMP
    RETURNING impact_id INTO impact_id;
    
    RETURN impact_id;
END;
$$ LANGUAGE plpgsql;

-- Função para configurar qualidade para tenant
CREATE OR REPLACE FUNCTION compliance_quality.configure_tenant_quality(
    tenant_id UUID,
    applicable_standards TEXT[],
    quality_system_name VARCHAR(255) DEFAULT 'Quality Management System',
    approval_workflow VARCHAR(50) DEFAULT 'STANDARD',
    audit_cycle JSONB DEFAULT NULL,
    custom_settings JSONB DEFAULT NULL
) RETURNS BOOLEAN AS $$
BEGIN
    INSERT INTO compliance_quality.tenant_quality_config (
        tenant_id,
        applicable_standards,
        quality_system_name,
        approval_workflow,
        audit_cycle,
        custom_settings
    ) VALUES (
        tenant_id,
        jsonb_build_array(VARIADIC applicable_standards),
        quality_system_name,
        approval_workflow,
        COALESCE(audit_cycle, '{}'::JSONB),
        COALESCE(custom_settings, '{}'::JSONB)
    )
    ON CONFLICT (tenant_id)
    DO UPDATE SET
        applicable_standards = jsonb_build_array(VARIADIC applicable_standards),
        quality_system_name = EXCLUDED.quality_system_name,
        approval_workflow = EXCLUDED.approval_workflow,
        audit_cycle = EXCLUDED.audit_cycle,
        custom_settings = EXCLUDED.custom_settings,
        updated_at = CURRENT_TIMESTAMP;
        
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Funções de Análise de Impacto na Qualidade
-- =====================================================================

-- Função para analisar impacto na qualidade a partir de não conformidade
CREATE OR REPLACE FUNCTION compliance_quality.analyze_quality_impact(
    validation_id UUID,
    tenant_id UUID
) RETURNS TABLE (
    quality_standard VARCHAR(100),
    process_area VARCHAR(255),
    impact_level VARCHAR(20),
    non_conformity_description TEXT
) AS $$
DECLARE
    validation_result RECORD;
    validator_mapping RECORD;
    tenant_config RECORD;
    requirement RECORD;
BEGIN
    -- Obter dados da validação
    SELECT * INTO validation_result
    FROM compliance_integrator.validation_history
    WHERE validation_id = analyze_quality_impact.validation_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Validação % não encontrada', validation_id;
    END IF;
    
    -- Obter configuração do tenant
    SELECT * INTO tenant_config
    FROM compliance_quality.tenant_quality_config
    WHERE tenant_id = analyze_quality_impact.tenant_id;
    
    IF NOT FOUND THEN
        -- Criar configuração padrão se não existir
        PERFORM compliance_quality.configure_tenant_quality(
            analyze_quality_impact.tenant_id,
            ARRAY['ISO 9001:2015']
        );
        
        SELECT * INTO tenant_config
        FROM compliance_quality.tenant_quality_config
        WHERE tenant_id = analyze_quality_impact.tenant_id;
    END IF;
    
    -- Processar cada requisito não conforme na validação
    FOR requirement IN 
        SELECT * FROM jsonb_to_recordset(validation_result.results->'requirements')
        AS x(requirement_id TEXT, requirement_name TEXT, is_compliant BOOLEAN, details TEXT)
        WHERE NOT is_compliant
    LOOP
        -- Buscar mapeamento de qualidade para este validador
        SELECT * INTO validator_mapping
        FROM compliance_quality.quality_standard_mapping
        WHERE validator_id = validation_result.validator_id
        AND is_active = TRUE
        LIMIT 1;
        
        IF FOUND THEN
            -- Retornar informações de impacto na qualidade
            quality_standard := validator_mapping.quality_standard;
            process_area := validator_mapping.process_area;
            impact_level := validator_mapping.impact_level;
            non_conformity_description := 'Não conformidade identificada na validação de ' || 
                                         validation_result.validator_name ||
                                         ': ' || requirement.details;
            RETURN NEXT;
        END IF;
    END LOOP;
    
    RETURN;
END;
$$ LANGUAGE plpgsql;

-- Função para gerar ação corretiva a partir de não conformidade
CREATE OR REPLACE FUNCTION compliance_quality.generate_corrective_action(
    validation_id UUID,
    tenant_id UUID,
    assigned_to VARCHAR(255) DEFAULT NULL,
    deadline TIMESTAMP WITH TIME ZONE DEFAULT NULL
) RETURNS SETOF UUID AS $$
DECLARE
    impact_result RECORD;
    action_id UUID;
    default_deadline TIMESTAMP WITH TIME ZONE;
    tenant_config RECORD;
BEGIN
    -- Definir prazo padrão (30 dias) se não especificado
    IF deadline IS NULL THEN
        default_deadline := CURRENT_TIMESTAMP + INTERVAL '30 days';
    ELSE
        default_deadline := deadline;
    END IF;
    
    -- Obter configuração do tenant
    SELECT * INTO tenant_config
    FROM compliance_quality.tenant_quality_config
    WHERE tenant_id = generate_corrective_action.tenant_id;
    
    -- Analisar impacto na qualidade e gerar ações correspondentes
    FOR impact_result IN 
        SELECT * FROM compliance_quality.analyze_quality_impact(
            validation_id,
            tenant_id
        )
    LOOP
        -- Criar ação corretiva
        INSERT INTO compliance_quality.corrective_action_register (
            tenant_id,
            validation_id,
            action_type,
            action_title,
            action_description,
            non_conformity_description,
            quality_standard,
            process_area,
            impact_level,
            assigned_to,
            deadline
        ) VALUES (
            tenant_id,
            validation_id,
            'CORRECTIVE',
            'Ação corretiva para ' || impact_result.quality_standard || ' - ' || impact_result.process_area,
            'Implementar correções para atender aos requisitos de ' || impact_result.quality_standard || 
            ' na área de processo ' || impact_result.process_area,
            impact_result.non_conformity_description,
            impact_result.quality_standard,
            impact_result.process_area,
            impact_result.impact_level,
            assigned_to,
            default_deadline
        ) RETURNING action_id INTO action_id;
        
        -- Registrar histórico da ação
        INSERT INTO compliance_quality.corrective_action_history (
            action_id,
            status_from,
            status_to,
            changed_by,
            comments
        ) VALUES (
            action_id,
            NULL,
            'CREATED',
            'SYSTEM',
            'Ação corretiva criada automaticamente a partir de validação de conformidade'
        );
        
        -- Atualizar métrica de qualidade relacionada, se existir
        UPDATE compliance_quality.quality_metrics
        SET 
            current_value = current_value - 5.0,
            trend = 'DETERIORATING',
            last_updated = CURRENT_TIMESTAMP
        WHERE 
            tenant_id = generate_corrective_action.tenant_id
            AND metric_category = 'COMPLIANCE'
            AND metric_code = 'COMP_QTY_' || REPLACE(impact_result.quality_standard, ':', '_');
        
        -- Retornar ID da ação criada
        RETURN NEXT action_id;
    END LOOP;
    
    RETURN;
END;
$$ LANGUAGE plpgsql;

-- Função para atualizar status de ação corretiva
CREATE OR REPLACE FUNCTION compliance_quality.update_corrective_action_status(
    action_id UUID,
    new_status VARCHAR(20),
    changed_by VARCHAR(255),
    comments TEXT DEFAULT NULL,
    attachments JSONB DEFAULT NULL
) RETURNS BOOLEAN AS $$
DECLARE
    current_status VARCHAR(20);
    action_record RECORD;
BEGIN
    -- Obter status atual
    SELECT status, tenant_id, quality_standard, impact_level INTO action_record
    FROM compliance_quality.corrective_action_register
    WHERE action_id = update_corrective_action_status.action_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Ação corretiva % não encontrada', action_id;
    END IF;

    current_status := action_record.status;
    
    -- Validar transição de status
    IF current_status = new_status THEN
        RETURN FALSE; -- Nenhuma mudança necessária
    END IF;
    
    -- Atualizar status da ação
    UPDATE compliance_quality.corrective_action_register
    SET 
        status = new_status,
        updated_at = CURRENT_TIMESTAMP,
        closure_date = CASE WHEN new_status IN ('CLOSED', 'VERIFIED') THEN CURRENT_TIMESTAMP ELSE closure_date END,
        closure_comments = CASE WHEN new_status IN ('CLOSED', 'VERIFIED') THEN comments ELSE closure_comments END,
        closure_evidence = CASE WHEN new_status IN ('CLOSED', 'VERIFIED') THEN attachments ELSE closure_evidence END
    WHERE action_id = update_corrective_action_status.action_id;
    
    -- Registrar histórico da mudança
    INSERT INTO compliance_quality.corrective_action_history (
        action_id,
        status_from,
        status_to,
        changed_by,
        comments,
        attachments
    ) VALUES (
        update_corrective_action_status.action_id,
        current_status,
        new_status,
        changed_by,
        comments,
        attachments
    );
    
    -- Se status for VERIFIED ou CLOSED, atualizar métricas de qualidade
    IF new_status IN ('VERIFIED', 'CLOSED') THEN
        -- Atualizar métrica relacionada
        UPDATE compliance_quality.quality_metrics
        SET 
            current_value = current_value + 
                CASE 
                    WHEN action_record.impact_level = 'CRITICAL' THEN 10.0
                    WHEN action_record.impact_level = 'MAJOR' THEN 7.5
                    WHEN action_record.impact_level = 'MINOR' THEN 5.0
                    ELSE 2.5
                END,
            trend = 'IMPROVING',
            last_updated = CURRENT_TIMESTAMP
        WHERE 
            tenant_id = action_record.tenant_id
            AND metric_category = 'COMPLIANCE'
            AND metric_code = 'COMP_QTY_' || REPLACE(action_record.quality_standard, ':', '_');
    END IF;
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- Função para verificar eficácia de ação corretiva
CREATE OR REPLACE FUNCTION compliance_quality.verify_action_effectiveness(
    action_id UUID,
    effectiveness_status VARCHAR(20),
    effectiveness_verification TEXT,
    verified_by VARCHAR(255)
) RETURNS BOOLEAN AS $$
BEGIN
    -- Atualizar ação com resultado da verificação
    UPDATE compliance_quality.corrective_action_register
    SET 
        effectiveness_status = verify_action_effectiveness.effectiveness_status,
        effectiveness_verification = verify_action_effectiveness.effectiveness_verification,
        status = CASE 
            WHEN effectiveness_status = 'EFFECTIVE' THEN 'CLOSED'
            WHEN effectiveness_status = 'PARTIALLY_EFFECTIVE' THEN 'IN_PROGRESS'
            WHEN effectiveness_status = 'NOT_EFFECTIVE' THEN 'CREATED'
            ELSE status
        END,
        updated_at = CURRENT_TIMESTAMP
    WHERE action_id = verify_action_effectiveness.action_id;
    
    -- Registrar no histórico
    INSERT INTO compliance_quality.corrective_action_history (
        action_id,
        status_from,
        status_to,
        changed_by,
        comments
    ) VALUES (
        verify_action_effectiveness.action_id,
        'VERIFIED',
        CASE 
            WHEN effectiveness_status = 'EFFECTIVE' THEN 'CLOSED'
            WHEN effectiveness_status = 'PARTIALLY_EFFECTIVE' THEN 'IN_PROGRESS'
            WHEN effectiveness_status = 'NOT_EFFECTIVE' THEN 'CREATED'
            ELSE 'VERIFIED'
        END,
        verified_by,
        'Verificação de eficácia: ' || effectiveness_status || '. ' || effectiveness_verification
    );
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Funções de Métricas de Qualidade
-- =====================================================================

-- Função para configurar métrica de qualidade
CREATE OR REPLACE FUNCTION compliance_quality.configure_quality_metric(
    tenant_id UUID,
    metric_name VARCHAR(255),
    metric_code VARCHAR(100),
    metric_category VARCHAR(100),
    description TEXT DEFAULT NULL,
    calculation_formula TEXT DEFAULT NULL,
    target_value NUMERIC(10,2) DEFAULT 100.0,
    threshold_warning NUMERIC(10,2) DEFAULT 80.0,
    threshold_critical NUMERIC(10,2) DEFAULT 60.0,
    unit VARCHAR(50) DEFAULT '%',
    measurement_period VARCHAR(50) DEFAULT 'MONTHLY',
    responsible_person VARCHAR(255) DEFAULT NULL,
    data_source VARCHAR(100) DEFAULT 'VALIDATORS'
) RETURNS UUID AS $$
DECLARE
    metric_id UUID;
BEGIN
    INSERT INTO compliance_quality.quality_metrics (
        tenant_id,
        metric_name,
        metric_code,
        metric_category,
        description,
        calculation_formula,
        current_value,
        target_value,
        threshold_warning,
        threshold_critical,
        unit,
        trend,
        measurement_period,
        responsible_person,
        data_source
    ) VALUES (
        tenant_id,
        metric_name,
        metric_code,
        metric_category,
        description,
        calculation_formula,
        target_value, -- Inicialmente definido como valor alvo (ideal)
        target_value,
        threshold_warning,
        threshold_critical,
        unit,
        'STABLE',
        measurement_period,
        responsible_person,
        data_source
    )
    ON CONFLICT (tenant_id, metric_code)
    DO UPDATE SET
        metric_name = EXCLUDED.metric_name,
        metric_category = EXCLUDED.metric_category,
        description = EXCLUDED.description,
        calculation_formula = EXCLUDED.calculation_formula,
        target_value = EXCLUDED.target_value,
        threshold_warning = EXCLUDED.threshold_warning,
        threshold_critical = EXCLUDED.threshold_critical,
        unit = EXCLUDED.unit,
        measurement_period = EXCLUDED.measurement_period,
        responsible_person = EXCLUDED.responsible_person,
        data_source = EXCLUDED.data_source,
        updated_at = CURRENT_TIMESTAMP
    RETURNING metric_id INTO metric_id;
    
    RETURN metric_id;
END;
$$ LANGUAGE plpgsql;

-- Função para atualizar valor de métrica de qualidade
CREATE OR REPLACE FUNCTION compliance_quality.update_quality_metric_value(
    tenant_id UUID,
    metric_code VARCHAR(100),
    metric_value NUMERIC(10,2),
    calculated_by VARCHAR(255) DEFAULT 'SYSTEM',
    notes TEXT DEFAULT NULL
) RETURNS BOOLEAN AS $$
DECLARE
    metric_id UUID;
    previous_value NUMERIC(10,2);
BEGIN
    -- Obter ID da métrica e valor anterior
    SELECT m.metric_id, m.current_value INTO metric_id, previous_value
    FROM compliance_quality.quality_metrics m
    WHERE m.tenant_id = update_quality_metric_value.tenant_id
    AND m.metric_code = update_quality_metric_value.metric_code;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Métrica % não encontrada para o tenant %', metric_code, tenant_id;
    END IF;
    
    -- Atualizar valor atual e tendência
    UPDATE compliance_quality.quality_metrics
    SET 
        current_value = metric_value,
        trend = CASE 
            WHEN metric_value > previous_value THEN 'IMPROVING'
            WHEN metric_value < previous_value THEN 'DETERIORATING'
            ELSE 'STABLE'
        END,
        last_updated = CURRENT_TIMESTAMP
    WHERE metric_id = metric_id;
    
    -- Registrar no histórico
    INSERT INTO compliance_quality.quality_metrics_history (
        metric_id,
        value_date,
        metric_value,
        calculated_by,
        notes
    ) VALUES (
        metric_id,
        CURRENT_TIMESTAMP,
        metric_value,
        calculated_by,
        notes
    );
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- Trigger para Processamento Automático
-- =====================================================================

-- Função de trigger para análise automática de impacto na qualidade após validação
CREATE OR REPLACE FUNCTION compliance_quality.validation_quality_trigger()
RETURNS TRIGGER AS $$
DECLARE
    tenant_config RECORD;
    action_ids UUID[];
BEGIN
    -- Verificar se existe configuração de qualidade para o tenant
    SELECT * INTO tenant_config
    FROM compliance_quality.tenant_quality_config
    WHERE tenant_id = NEW.tenant_id;
    
    IF FOUND THEN
        -- Gerar ações corretivas automaticamente
        SELECT array_agg(action_id) INTO action_ids
        FROM compliance_quality.generate_corrective_action(
            NEW.validation_id,
            NEW.tenant_id
        ) AS action_id;
        
        -- Atualizar validação com referência às ações corretivas geradas
        UPDATE compliance_integrator.validation_history
        SET metadata = COALESCE(metadata, '{}'::JSONB) || 
                       jsonb_build_object('quality_actions', action_ids)
        WHERE validation_id = NEW.validation_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Criação do trigger
DROP TRIGGER IF EXISTS validation_quality_trigger 
ON compliance_integrator.validation_history;

CREATE TRIGGER validation_quality_trigger
AFTER INSERT ON compliance_integrator.validation_history
FOR EACH ROW EXECUTE FUNCTION compliance_quality.validation_quality_trigger();

-- =====================================================================
-- Mapeamentos Padrão para Padrões de Qualidade
-- =====================================================================

-- Inserção de mapeamentos padrão entre validadores e padrões de qualidade
DO $$
BEGIN
    -- Mapeamentos para o setor de saúde
    PERFORM compliance_quality.register_quality_standard_mapping(
        'hipaa_compliance', 'Validador HIPAA', 
        'ISO 13485:2016', 'Controle de Registros Médicos', 
        'Seção 4.2.5', 'CRITICAL', 'AUTOMATED'
    );
    
    PERFORM compliance_quality.register_quality_standard_mapping(
        'gdpr_healthcare', 'Validador GDPR para Saúde', 
        'ISO 13485:2016', 'Controle de Dados do Paciente', 
        'Seção 4.2.3', 'MAJOR', 'AUTOMATED'
    );
    
    PERFORM compliance_quality.register_quality_standard_mapping(
        'lgpd_healthcare', 'Validador LGPD para Saúde', 
        'ISO 13485:2016', 'Confidencialidade de Informações', 
        'Seção 4.2.4', 'MAJOR', 'AUTOMATED'
    );
    
    -- Mapeamentos para o setor financeiro
    PERFORM compliance_quality.register_quality_standard_mapping(
        'psd2_compliance', 'Validador PSD2', 
        'ISO 9001:2015', 'Segurança de Acesso', 
        'Seção 8.5.3', 'CRITICAL', 'AUTOMATED'
    );
    
    PERFORM compliance_quality.register_quality_standard_mapping(
        'openbanking_compliance', 'Validador Open Banking', 
        'ISO 9001:2015', 'Controle de Acesso a Dados', 
        'Seção 7.5.3', 'MAJOR', 'AUTOMATED'
    );
    
    -- Mapeamentos para o setor governamental
    PERFORM compliance_quality.register_quality_standard_mapping(
        'eidas_compliance', 'Validador eIDAS', 
        'ISO 9001:2015', 'Identificação e Rastreabilidade', 
        'Seção 8.5.2', 'MAJOR', 'AUTOMATED'
    );
    
    PERFORM compliance_quality.register_quality_standard_mapping(
        'icpbrasil_compliance', 'Validador ICP-Brasil', 
        'ISO 9001:2015', 'Controle de Informação Documentada', 
        'Seção 7.5', 'MAJOR', 'AUTOMATED'
    );
    
    -- Mapeamentos para AR/VR
    PERFORM compliance_quality.register_quality_standard_mapping(
        'ieee_xr_compliance', 'Validador IEEE XR', 
        'ISO/IEC 25000', 'Qualidade de Software', 
        'Seção 6.1', 'MEDIUM', 'AUTOMATED'
    );
    
    PERFORM compliance_quality.register_quality_standard_mapping(
        'nist_xr_compliance', 'Validador NIST XR', 
        'ISO/IEC 25000', 'Segurança de Software', 
        'Seção 7.3', 'MAJOR', 'AUTOMATED'
    );
END;
$$;

-- Configuração de métricas de qualidade padrão
DO $$
DECLARE
    test_tenant_id UUID := '550e8400-e29b-41d4-a716-446655440000'; -- UUID de exemplo
BEGIN
    -- Métricas de conformidade padrão
    PERFORM compliance_quality.configure_quality_metric(
        test_tenant_id,
        'Conformidade ISO 9001',
        'COMP_QTY_ISO_9001_2015',
        'COMPLIANCE',
        'Nível de conformidade com ISO 9001:2015',
        'Porcentagem de requisitos atendidos',
        95.0, 80.0, 60.0,
        '%', 'MONTHLY', 'quality_manager@example.com', 'VALIDATORS'
    );

    PERFORM compliance_quality.configure_quality_metric(
        test_tenant_id,
        'Conformidade ISO 13485',
        'COMP_QTY_ISO_13485_2016',
        'COMPLIANCE',
        'Nível de conformidade com ISO 13485:2016',
        'Porcentagem de requisitos atendidos',
        95.0, 85.0, 70.0,
        '%', 'MONTHLY', 'healthcare_quality@example.com', 'VALIDATORS'
    );
    
    -- Métricas de processo
    PERFORM compliance_quality.configure_quality_metric(
        test_tenant_id,
        'Ações Corretivas em Tempo',
        'PROC_QTY_CAPA_ON_TIME',
        'PROCESS',
        'Porcentagem de ações corretivas concluídas dentro do prazo',
        'Número de ações concluídas no prazo / Total de ações * 100',
        90.0, 75.0, 50.0,
        '%', 'MONTHLY', 'quality_manager@example.com', 'MANUAL'
    );
    
    -- Métricas de desempenho
    PERFORM compliance_quality.configure_quality_metric(
        test_tenant_id,
        'Eficácia de Ações Corretivas',
        'PERF_QTY_CAPA_EFFECT',
        'PERFORMANCE',
        'Porcentagem de ações corretivas consideradas eficazes',
        'Número de ações eficazes / Total de ações verificadas * 100',
        85.0, 70.0, 50.0,
        '%', 'QUARTERLY', 'quality_manager@example.com', 'MANUAL'
    );
END;
$$;

-- =====================================================================
-- Comentários de Documentação
-- =====================================================================

COMMENT ON SCHEMA compliance_quality IS 'Schema para integração entre validadores de conformidade e sistema de gestão da qualidade';

COMMENT ON TABLE compliance_quality.quality_standard_mapping IS 'Mapeamento entre validadores e padrões de qualidade';
COMMENT ON TABLE compliance_quality.process_impact_matrix IS 'Matriz de impacto em processos organizacionais';
COMMENT ON TABLE compliance_quality.tenant_quality_config IS 'Configuração de qualidade por tenant';
COMMENT ON TABLE compliance_quality.corrective_action_register IS 'Registro de ações corretivas e preventivas';
COMMENT ON TABLE compliance_quality.corrective_action_history IS 'Histórico de ações corretivas';
COMMENT ON TABLE compliance_quality.quality_metrics IS 'Métricas e KPIs de qualidade';
COMMENT ON TABLE compliance_quality.quality_metrics_history IS 'Histórico de métricas de qualidade';

COMMENT ON FUNCTION compliance_quality.register_quality_standard_mapping IS 'Registra mapeamento entre validador e padrão de qualidade';
COMMENT ON FUNCTION compliance_quality.register_process_impact IS 'Registra matriz de impacto em processos';
COMMENT ON FUNCTION compliance_quality.configure_tenant_quality IS 'Configura qualidade para tenant';
COMMENT ON FUNCTION compliance_quality.analyze_quality_impact IS 'Analisa impacto na qualidade a partir de não conformidade';
COMMENT ON FUNCTION compliance_quality.generate_corrective_action IS 'Gera ação corretiva a partir de não conformidade';
COMMENT ON FUNCTION compliance_quality.update_corrective_action_status IS 'Atualiza status de ação corretiva';
COMMENT ON FUNCTION compliance_quality.verify_action_effectiveness IS 'Verifica eficácia de ação corretiva';
COMMENT ON FUNCTION compliance_quality.configure_quality_metric IS 'Configura métrica de qualidade';
COMMENT ON FUNCTION compliance_quality.update_quality_metric_value IS 'Atualiza valor de métrica de qualidade';
COMMENT ON FUNCTION compliance_quality.validation_quality_trigger IS 'Trigger para análise automática de impacto na qualidade após validação';

-- =====================================================================
-- Fim do Script
-- =====================================================================

-- =============================================================================
-- Script de Integração IAM com Sistema de Gestão da Qualidade
-- =============================================================================
-- Autor: Eduardo Jeremias
-- Data: 15/05/2025
-- Versão: 1.0
-- Descrição: Este script implementa a integração entre os validadores de
--            conformidade IAM e o Sistema de Gestão da Qualidade.
-- =============================================================================

-- Verificação de ambiente
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'compliance_validators') THEN
        RAISE EXCEPTION 'Schema compliance_validators não existe. Execute os scripts anteriores primeiro.';
    END IF;
END
$$;

-- =============================================================================
-- 1. Criação do Schema e Tabelas Base
-- =============================================================================

-- Schema principal para o sistema de gestão da qualidade
CREATE SCHEMA IF NOT EXISTS quality_management;

COMMENT ON SCHEMA quality_management IS 'Schema para o Sistema de Gestão da Qualidade integrado com validadores IAM';

-- Tabela de mapeamento entre padrões de qualidade e validadores
CREATE TABLE quality_management.standard_validator_mapping (
    mapping_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    standard_id VARCHAR(50) NOT NULL,
    standard_code VARCHAR(50) NOT NULL,
    standard_name VARCHAR(255) NOT NULL,
    validator_id VARCHAR(50) NOT NULL,
    validator_name VARCHAR(255) NOT NULL,
    requirement_id VARCHAR(50) NOT NULL,
    requirement_name VARCHAR(255) NOT NULL,
    relationship_description TEXT,
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE quality_management.standard_validator_mapping IS 'Mapeamento entre padrões de qualidade e validadores de conformidade';

-- Tabela de padrões de qualidade
CREATE TABLE quality_management.quality_standards (
    standard_id VARCHAR(50) PRIMARY KEY,
    standard_code VARCHAR(50) NOT NULL,
    standard_name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    version VARCHAR(50),
    issuing_body VARCHAR(255),
    publication_date DATE,
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE quality_management.quality_standards IS 'Catálogo de padrões de qualidade suportados';

-- Tabela de requisitos de padrões de qualidade
CREATE TABLE quality_management.standard_requirements (
    requirement_id VARCHAR(50) NOT NULL,
    standard_id VARCHAR(50) NOT NULL REFERENCES quality_management.quality_standards(standard_id),
    requirement_code VARCHAR(50) NOT NULL,
    requirement_name VARCHAR(255) NOT NULL,
    description TEXT,
    verification_method TEXT,
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (requirement_id, standard_id)
);

COMMENT ON TABLE quality_management.standard_requirements IS 'Requisitos específicos de cada padrão de qualidade';

-- Tabela de não-conformidades
CREATE TABLE quality_management.non_conformity (
    non_conformity_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    standard_id VARCHAR(50) NOT NULL,
    standard_name VARCHAR(255) NOT NULL,
    validator_id VARCHAR(50) NOT NULL,
    validator_name VARCHAR(255) NOT NULL,
    requirement_id VARCHAR(50) NOT NULL,
    requirement_name VARCHAR(255) NOT NULL,
    validation_id UUID NOT NULL,
    details JSONB NOT NULL,
    impact_level VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN',
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE quality_management.non_conformity IS 'Registro de não-conformidades de qualidade';

-- Tabela de ações corretivas
CREATE TABLE quality_management.corrective_action (
    action_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    non_conformity_id UUID NOT NULL REFERENCES quality_management.non_conformity(non_conformity_id),
    action_type VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    assigned_to VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    due_date TIMESTAMP WITH TIME ZONE,
    completed_date TIMESTAMP WITH TIME ZONE,
    effectiveness_evaluation TEXT,
    effectiveness_status VARCHAR(20),
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE quality_management.corrective_action IS 'Ações corretivas para não-conformidades';

-- Tabela de métricas de qualidade
CREATE TABLE quality_management.quality_metrics (
    metric_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_name VARCHAR(100) NOT NULL,
    metric_type VARCHAR(50) NOT NULL,
    standard_id VARCHAR(50),
    current_value NUMERIC NOT NULL,
    target_value NUMERIC NOT NULL,
    unit VARCHAR(50) NOT NULL,
    calculation_period VARCHAR(50) NOT NULL,
    last_calculated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE quality_management.quality_metrics IS 'Métricas de acompanhamento da qualidade';

-- Tabela de templates de ações
CREATE TABLE quality_management.action_templates (
    template_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    standard_id VARCHAR(50) NOT NULL,
    action_type VARCHAR(50) NOT NULL,
    template_name VARCHAR(255) NOT NULL,
    template_content TEXT NOT NULL,
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE quality_management.action_templates IS 'Templates para geração automática de ações corretivas';

-- Tabela de histórico de qualidade
CREATE TABLE quality_management.quality_history (
    history_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    change_type VARCHAR(50) NOT NULL,
    previous_state JSONB,
    current_state JSONB,
    changed_by VARCHAR(255),
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE quality_management.quality_history IS 'Histórico de alterações para auditoria';

-- Índices para otimização de desempenho
CREATE INDEX idx_non_conformity_tenant ON quality_management.non_conformity(tenant_id);
CREATE INDEX idx_non_conformity_status ON quality_management.non_conformity(status);
CREATE INDEX idx_corrective_action_tenant ON quality_management.corrective_action(tenant_id);
CREATE INDEX idx_corrective_action_status ON quality_management.corrective_action(status);
CREATE INDEX idx_mapping_validator ON quality_management.standard_validator_mapping(validator_id, requirement_id);
CREATE INDEX idx_mapping_standard ON quality_management.standard_validator_mapping(standard_id, tenant_id);

-- =============================================================================
-- 2. Funções para Gerenciamento de Padrões e Mapeamentos
-- =============================================================================

-- Função para registrar um padrão de qualidade
CREATE OR REPLACE FUNCTION quality_management.register_quality_standard(
    p_standard_id VARCHAR(50),
    p_standard_code VARCHAR(50),
    p_standard_name VARCHAR(255),
    p_description TEXT,
    p_category VARCHAR(100),
    p_version VARCHAR(50),
    p_issuing_body VARCHAR(255),
    p_publication_date DATE,
    p_tenant_id UUID
) RETURNS VARCHAR(50) AS $$
BEGIN
    INSERT INTO quality_management.quality_standards (
        standard_id, standard_code, standard_name,
        description, category, version,
        issuing_body, publication_date, tenant_id
    ) VALUES (
        p_standard_id, p_standard_code, p_standard_name,
        p_description, p_category, p_version,
        p_issuing_body, p_publication_date, p_tenant_id
    )
    ON CONFLICT (standard_id) DO UPDATE
    SET standard_code = p_standard_code,
        standard_name = p_standard_name,
        description = p_description,
        category = p_category,
        version = p_version,
        issuing_body = p_issuing_body,
        publication_date = p_publication_date,
        updated_at = CURRENT_TIMESTAMP;
    
    -- Registrar auditoria
    INSERT INTO quality_management.quality_history (
        entity_type, entity_id, change_type, 
        current_state, tenant_id
    ) VALUES (
        'QUALITY_STANDARD', p_standard_id::UUID, 
        'UPSERT',
        jsonb_build_object(
            'standard_id', p_standard_id,
            'standard_name', p_standard_name,
            'version', p_version
        ),
        p_tenant_id
    );
    
    RETURN p_standard_id;
END;
$$ LANGUAGE plpgsql;

-- Função para registrar um requisito de padrão de qualidade
CREATE OR REPLACE FUNCTION quality_management.register_standard_requirement(
    p_requirement_id VARCHAR(50),
    p_standard_id VARCHAR(50),
    p_requirement_code VARCHAR(50),
    p_requirement_name VARCHAR(255),
    p_description TEXT,
    p_verification_method TEXT,
    p_tenant_id UUID
) RETURNS VARCHAR(50) AS $$
BEGIN
    INSERT INTO quality_management.standard_requirements (
        requirement_id, standard_id, requirement_code,
        requirement_name, description, verification_method,
        tenant_id
    ) VALUES (
        p_requirement_id, p_standard_id, p_requirement_code,
        p_requirement_name, p_description, p_verification_method,
        p_tenant_id
    )
    ON CONFLICT (requirement_id, standard_id) DO UPDATE
    SET requirement_code = p_requirement_code,
        requirement_name = p_requirement_name,
        description = p_description,
        verification_method = p_verification_method,
        updated_at = CURRENT_TIMESTAMP;
    
    RETURN p_requirement_id;
END;
$$ LANGUAGE plpgsql;

-- Função para registrar mapeamento entre padrão de qualidade e validador
CREATE OR REPLACE FUNCTION quality_management.register_standard_mapping(
    p_standard_id VARCHAR(50),
    p_standard_code VARCHAR(50),
    p_standard_name VARCHAR(255),
    p_validator_id VARCHAR(50),
    p_validator_name VARCHAR(255),
    p_requirement_id VARCHAR(50),
    p_requirement_name VARCHAR(255),
    p_relationship_description TEXT,
    p_tenant_id UUID
) RETURNS UUID AS $$
DECLARE
    v_mapping_id UUID;
BEGIN
    INSERT INTO quality_management.standard_validator_mapping (
        standard_id, standard_code, standard_name,
        validator_id, validator_name,
        requirement_id, requirement_name,
        relationship_description, tenant_id
    ) VALUES (
        p_standard_id, p_standard_code, p_standard_name,
        p_validator_id, p_validator_name,
        p_requirement_id, p_requirement_name,
        p_relationship_description, p_tenant_id
    )
    RETURNING mapping_id INTO v_mapping_id;
    
    -- Registrar auditoria
    INSERT INTO quality_management.quality_history (
        entity_type, entity_id, change_type, 
        current_state, tenant_id
    ) VALUES (
        'STANDARD_MAPPING', v_mapping_id, 
        'CREATE',
        jsonb_build_object(
            'standard_id', p_standard_id,
            'validator_id', p_validator_id,
            'requirement_id', p_requirement_id
        ),
        p_tenant_id
    );
    
    RETURN v_mapping_id;
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- 3. Funções para Gestão de Não-Conformidades
-- =============================================================================

-- Função auxiliar para atualizar ou inserir métrica
CREATE OR REPLACE FUNCTION quality_management.update_or_insert_metric(
    p_metric_name VARCHAR(100),
    p_metric_type VARCHAR(50),
    p_standard_id VARCHAR(50),
    p_current_value NUMERIC,
    p_target_value NUMERIC,
    p_unit VARCHAR(50),
    p_calculation_period VARCHAR(50),
    p_tenant_id UUID
) RETURNS UUID AS $$
DECLARE
    v_metric_id UUID;
BEGIN
    -- Verificar se a métrica já existe
    SELECT metric_id INTO v_metric_id
    FROM quality_management.quality_metrics
    WHERE metric_name = p_metric_name
    AND (p_standard_id IS NULL OR standard_id = p_standard_id)
    AND tenant_id = p_tenant_id
    LIMIT 1;
    
    -- Se existir, atualizar
    IF v_metric_id IS NOT NULL THEN
        UPDATE quality_management.quality_metrics
        SET current_value = p_current_value,
            target_value = p_target_value,
            last_calculated = CURRENT_TIMESTAMP,
            updated_at = CURRENT_TIMESTAMP
        WHERE metric_id = v_metric_id;
    -- Se não existir, inserir
    ELSE
        INSERT INTO quality_management.quality_metrics (
            metric_name, metric_type, standard_id,
            current_value, target_value, unit,
            calculation_period, tenant_id
        ) VALUES (
            p_metric_name, p_metric_type, p_standard_id,
            p_current_value, p_target_value, p_unit,
            p_calculation_period, p_tenant_id
        )
        RETURNING metric_id INTO v_metric_id;
    END IF;
    
    RETURN v_metric_id;
END;
$$ LANGUAGE plpgsql;

-- Função para criar não-conformidade com base em resultado de validação
CREATE OR REPLACE FUNCTION quality_management.create_non_conformity(
    p_validation_id UUID,
    p_validator_id VARCHAR(50),
    p_validator_name VARCHAR(255),
    p_requirement_id VARCHAR(50),
    p_requirement_name VARCHAR(255),
    p_details JSONB,
    p_tenant_id UUID
) RETURNS UUID AS $$
DECLARE
    v_non_conformity_id UUID;
    v_mapping RECORD;
    v_impact_level VARCHAR(20);
BEGIN
    -- Encontrar o mapeamento para o padrão de qualidade
    SELECT * INTO v_mapping
    FROM quality_management.standard_validator_mapping
    WHERE validator_id = p_validator_id
    AND requirement_id = p_requirement_id
    AND tenant_id = p_tenant_id
    LIMIT 1;
    
    -- Se não encontrar mapeamento, usar valores padrão
    IF v_mapping IS NULL THEN
        v_mapping.standard_id := 'DEFAULT';
        v_mapping.standard_name := 'Padrão de Conformidade IAM';
        v_mapping.standard_code := 'IAM-STD';
    END IF;
    
    -- Determinar nível de impacto baseado nos detalhes
    IF p_details->>'criticality' = 'HIGH' THEN
        v_impact_level := 'MAJOR';
    ELSIF p_details->>'criticality' = 'MEDIUM' THEN
        v_impact_level := 'MODERATE';
    ELSE
        v_impact_level := 'MINOR';
    END IF;
    
    -- Inserir a não-conformidade
    INSERT INTO quality_management.non_conformity (
        standard_id, standard_name,
        validator_id, validator_name,
        requirement_id, requirement_name,
        validation_id, details,
        impact_level, status,
        tenant_id
    ) VALUES (
        v_mapping.standard_id, v_mapping.standard_name,
        p_validator_id, p_validator_name,
        p_requirement_id, p_requirement_name,
        p_validation_id, p_details,
        v_impact_level, 'OPEN',
        p_tenant_id
    )
    RETURNING non_conformity_id INTO v_non_conformity_id;
    
    -- Registrar auditoria
    INSERT INTO quality_management.quality_history (
        entity_type, entity_id, change_type, 
        current_state, tenant_id
    ) VALUES (
        'NON_CONFORMITY', v_non_conformity_id, 
        'CREATE',
        jsonb_build_object(
            'validator_id', p_validator_id,
            'requirement_id', p_requirement_id,
            'impact_level', v_impact_level,
            'details', p_details
        ),
        p_tenant_id
    );
    
    -- Criar ação corretiva automaticamente
    PERFORM quality_management.create_corrective_action(
        v_non_conformity_id,
        'CORRECTIVE',
        'Corrigir não-conformidade de ' || p_requirement_name,
        p_tenant_id
    );
    
    -- Atualizar métricas de qualidade
    PERFORM quality_management.update_quality_metrics(p_tenant_id);
    
    RETURN v_non_conformity_id;
END;
$$ LANGUAGE plpgsql;

-- Função para criar ação corretiva
CREATE OR REPLACE FUNCTION quality_management.create_corrective_action(
    p_non_conformity_id UUID,
    p_action_type VARCHAR(50),
    p_description TEXT,
    p_tenant_id UUID,
    p_assigned_to VARCHAR(255) DEFAULT NULL,
    p_due_date TIMESTAMP WITH TIME ZONE DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    v_action_id UUID;
    v_non_conformity RECORD;
    v_template TEXT;
BEGIN
    -- Obter detalhes da não-conformidade
    SELECT * INTO v_non_conformity
    FROM quality_management.non_conformity
    WHERE non_conformity_id = p_non_conformity_id
    AND tenant_id = p_tenant_id;
    
    -- Criar descrição detalhada baseada em templates
    SELECT template_content INTO v_template
    FROM quality_management.action_templates
    WHERE standard_id = v_non_conformity.standard_id
    AND action_type = p_action_type
    AND tenant_id = p_tenant_id
    LIMIT 1;
    
    -- Se não encontrar template, usar descrição padrão
    IF v_template IS NULL THEN
        v_template := p_description;
    ELSE
        -- Substituir placeholders no template
        v_template := REPLACE(v_template, '{requirement_name}', v_non_conformity.requirement_name);
        v_template := REPLACE(v_template, '{details}', v_non_conformity.details->>'message');
    END IF;
    
    -- Definir data de vencimento baseada no impacto, se não for fornecida
    IF p_due_date IS NULL THEN
        IF v_non_conformity.impact_level = 'MAJOR' THEN
            p_due_date := CURRENT_TIMESTAMP + INTERVAL '3 days';
        ELSIF v_non_conformity.impact_level = 'MODERATE' THEN
            p_due_date := CURRENT_TIMESTAMP + INTERVAL '7 days';
        ELSE
            p_due_date := CURRENT_TIMESTAMP + INTERVAL '14 days';
        END IF;
    END IF;
    
    -- Inserir a ação corretiva
    INSERT INTO quality_management.corrective_action (
        non_conformity_id, action_type, description,
        assigned_to, status, due_date,
        tenant_id
    ) VALUES (
        p_non_conformity_id, p_action_type, v_template,
        p_assigned_to, 'PENDING', p_due_date,
        p_tenant_id
    )
    RETURNING action_id INTO v_action_id;
    
    -- Registrar auditoria
    INSERT INTO quality_management.quality_history (
        entity_type, entity_id, change_type, 
        current_state, tenant_id
    ) VALUES (
        'CORRECTIVE_ACTION', v_action_id, 
        'CREATE',
        jsonb_build_object(
            'non_conformity_id', p_non_conformity_id,
            'action_type', p_action_type,
            'status', 'PENDING',
            'due_date', p_due_date
        ),
        p_tenant_id
    );
    
    RETURN v_action_id;
END;
$$ LANGUAGE plpgsql;

-- Função para atualizar métricas de qualidade
CREATE OR REPLACE FUNCTION quality_management.update_quality_metrics(
    p_tenant_id UUID
) RETURNS VOID AS $$
DECLARE
    v_total_non_conformities INTEGER;
    v_open_non_conformities INTEGER;
    v_compliance_rate NUMERIC;
    v_avg_resolution_days NUMERIC;
    v_standard RECORD;
BEGIN
    -- Calcular total de não-conformidades
    SELECT COUNT(*) INTO v_total_non_conformities
    FROM quality_management.non_conformity
    WHERE tenant_id = p_tenant_id;
    
    -- Calcular não-conformidades abertas
    SELECT COUNT(*) INTO v_open_non_conformities
    FROM quality_management.non_conformity
    WHERE tenant_id = p_tenant_id
    AND status = 'OPEN';
    
    -- Calcular taxa de conformidade
    IF v_total_non_conformities > 0 THEN
        v_compliance_rate := 100 - (v_open_non_conformities::NUMERIC / v_total_non_conformities::NUMERIC * 100);
    ELSE
        v_compliance_rate := 100;
    END IF;
    
    -- Calcular tempo médio de resolução
    SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (completed_date - created_at))/86400), 0) INTO v_avg_resolution_days
    FROM quality_management.corrective_action
    WHERE tenant_id = p_tenant_id
    AND status = 'COMPLETED'
    AND completed_date IS NOT NULL;
    
    -- Atualizar ou inserir métricas globais
    -- 1. Taxa de conformidade
    PERFORM quality_management.update_or_insert_metric(
        'COMPLIANCE_RATE', 'PERCENT', NULL,
        v_compliance_rate, 100, '%',
        'MONTHLY', p_tenant_id
    );
    
    -- 2. Não-conformidades abertas
    PERFORM quality_management.update_or_insert_metric(
        'OPEN_NON_CONFORMITIES', 'COUNT', NULL,
        v_open_non_conformities, 0, 'unidades',
        'DAILY', p_tenant_id
    );
    
    -- 3. Tempo médio de resolução
    PERFORM quality_management.update_or_insert_metric(
        'AVG_RESOLUTION_TIME', 'TIME', NULL,
        v_avg_resolution_days, 10, 'dias',
        'MONTHLY', p_tenant_id
    );
    
    -- Atualizar métricas por padrão de qualidade
    FOR v_standard IN 
        SELECT DISTINCT standard_id, standard_name 
        FROM quality_management.non_conformity
        WHERE tenant_id = p_tenant_id
    LOOP
        -- Calcular total de não-conformidades para este padrão
        SELECT COUNT(*) INTO v_total_non_conformities
        FROM quality_management.non_conformity
        WHERE tenant_id = p_tenant_id
        AND standard_id = v_standard.standard_id;
        
        -- Calcular não-conformidades abertas para este padrão
        SELECT COUNT(*) INTO v_open_non_conformities
        FROM quality_management.non_conformity
        WHERE tenant_id = p_tenant_id
        AND standard_id = v_standard.standard_id
        AND status = 'OPEN';
        
        -- Calcular taxa de conformidade para este padrão
        IF v_total_non_conformities > 0 THEN
            v_compliance_rate := 100 - (v_open_non_conformities::NUMERIC / v_total_non_conformities::NUMERIC * 100);
        ELSE
            v_compliance_rate := 100;
        END IF;
        
        -- Atualizar métrica de conformidade para este padrão
        PERFORM quality_management.update_or_insert_metric(
            v_standard.standard_id || '_COMPLIANCE_RATE', 'PERCENT', v_standard.standard_id,
            v_compliance_rate, 100, '%',
            'MONTHLY', p_tenant_id
        );
    END LOOP;
END;
$$ LANGUAGE plpgsql;
