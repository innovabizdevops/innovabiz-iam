-- INNOVABIZ - IAM ISO 27001 Functions (Parte 1)
-- Author: Eduardo Jeremias
-- Date: 08/05/2025
-- Version: 1.0
-- Description: Funções para gerenciamento de compliance ISO 27001 com foco em saúde.

-- Configurar caminho de busca
SET search_path TO iam, public;

-- Função para criar uma nova avaliação ISO 27001
CREATE OR REPLACE FUNCTION create_iso27001_assessment(
    p_organization_id UUID,
    p_name VARCHAR(255),
    p_description TEXT DEFAULT NULL,
    p_start_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    p_scope JSONB DEFAULT '{}'::JSONB,
    p_version VARCHAR(50) DEFAULT '2013',
    p_created_by UUID DEFAULT NULL,
    p_healthcare_specific BOOLEAN DEFAULT FALSE,
    p_framework_id UUID DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    v_assessment_id UUID;
    v_control_ids UUID[];
BEGIN
    -- Inserir o registro da avaliação
    INSERT INTO iso27001_assessments (
        organization_id,
        name,
        description,
        start_date,
        status,
        scope,
        version,
        created_by,
        updated_by,
        healthcare_specific,
        framework_id
    ) VALUES (
        p_organization_id,
        p_name,
        p_description,
        p_start_date,
        'in_progress',
        p_scope,
        p_version,
        p_created_by,
        p_created_by,
        p_healthcare_specific,
        p_framework_id
    ) RETURNING id INTO v_assessment_id;
    
    -- Obter todos os IDs de controles ativos
    SELECT ARRAY_AGG(id) INTO v_control_ids
    FROM iso27001_controls
    WHERE is_active = TRUE;
    
    -- Se for uma avaliação específica de saúde, incluir apenas controles com healthcare_applicability não nulo
    IF p_healthcare_specific THEN
        SELECT ARRAY_AGG(id) INTO v_control_ids
        FROM iso27001_controls
        WHERE is_active = TRUE
        AND healthcare_applicability IS NOT NULL;
    END IF;

    -- Criar resultados de controle em branco para cada controle ativo
    IF v_control_ids IS NOT NULL THEN
        INSERT INTO iso27001_control_results (
            assessment_id,
            control_id,
            status,
            implementation_status,
            created_by
        )
        SELECT 
            v_assessment_id,
            id,
            'non_compliant',
            'not_implemented',
            p_created_by
        FROM 
            UNNEST(v_control_ids) AS id;
    END IF;
    
    RETURN v_assessment_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION create_iso27001_assessment IS 'Cria uma nova avaliação ISO 27001 e inicializa resultados de controle em branco';

-- Função para atualizar o resultado de um controle ISO 27001
CREATE OR REPLACE FUNCTION update_iso27001_control_result(
    p_result_id UUID,
    p_status VARCHAR(50),
    p_implementation_status VARCHAR(50) DEFAULT NULL,
    p_score FLOAT DEFAULT NULL,
    p_evidence TEXT DEFAULT NULL,
    p_notes TEXT DEFAULT NULL,
    p_updated_by UUID DEFAULT NULL,
    p_issues_found JSONB DEFAULT NULL,
    p_recommendations JSONB DEFAULT NULL,
    p_healthcare_specific_findings TEXT DEFAULT NULL
) RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar se o status é válido
    IF p_status NOT IN ('compliant', 'partial_compliance', 'non_compliant', 'not_applicable') THEN
        RAISE EXCEPTION 'Status inválido: %. Os valores permitidos são: compliant, partial_compliance, non_compliant, not_applicable', p_status;
    END IF;
    
    -- Verificar se o status de implementação é válido
    IF p_implementation_status IS NOT NULL AND p_implementation_status NOT IN ('not_implemented', 'partially_implemented', 'implemented', 'not_applicable') THEN
        RAISE EXCEPTION 'Status de implementação inválido: %. Os valores permitidos são: not_implemented, partially_implemented, implemented, not_applicable', p_implementation_status;
    END IF;
    
    -- Atualizar o resultado do controle
    UPDATE iso27001_control_results
    SET 
        status = p_status,
        implementation_status = COALESCE(p_implementation_status, implementation_status),
        score = COALESCE(p_score, score),
        evidence = COALESCE(p_evidence, evidence),
        notes = COALESCE(p_notes, notes),
        updated_by = COALESCE(p_updated_by, updated_by),
        issues_found = COALESCE(p_issues_found, issues_found),
        recommendations = COALESCE(p_recommendations, recommendations),
        healthcare_specific_findings = COALESCE(p_healthcare_specific_findings, healthcare_specific_findings),
        updated_at = NOW()
    WHERE id = p_result_id;
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_iso27001_control_result IS 'Atualiza o resultado de avaliação de um controle ISO 27001';

-- Função para finalizar uma avaliação ISO 27001 e calcular pontuação
CREATE OR REPLACE FUNCTION finalize_iso27001_assessment(
    p_assessment_id UUID,
    p_updated_by UUID DEFAULT NULL
) RETURNS FLOAT AS $$
DECLARE
    v_total_controls INT;
    v_applicable_controls INT;
    v_compliant_controls INT;
    v_partial_compliance_controls INT;
    v_partial_weight FLOAT := 0.5; -- Peso para controles com conformidade parcial
    v_score FLOAT;
BEGIN
    -- Contar os controles
    SELECT 
        COUNT(*),
        COUNT(*) FILTER (WHERE status != 'not_applicable'),
        COUNT(*) FILTER (WHERE status = 'compliant'),
        COUNT(*) FILTER (WHERE status = 'partial_compliance')
    INTO 
        v_total_controls,
        v_applicable_controls,
        v_compliant_controls,
        v_partial_compliance_controls
    FROM 
        iso27001_control_results
    WHERE 
        assessment_id = p_assessment_id;
    
    -- Calcular a pontuação
    IF v_applicable_controls > 0 THEN
        v_score := (v_compliant_controls + (v_partial_compliance_controls * v_partial_weight)) / v_applicable_controls * 100;
    ELSE
        v_score := NULL;
    END IF;
    
    -- Atualizar a avaliação
    UPDATE iso27001_assessments
    SET 
        status = 'completed',
        end_date = NOW(),
        updated_by = COALESCE(p_updated_by, updated_by),
        updated_at = NOW(),
        score = v_score
    WHERE 
        id = p_assessment_id;
    
    RETURN v_score;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION finalize_iso27001_assessment IS 'Finaliza uma avaliação ISO 27001 e calcula a pontuação com base nos resultados dos controles';

-- Função para criar um plano de ação ISO 27001
CREATE OR REPLACE FUNCTION create_iso27001_action_plan(
    p_organization_id UUID,
    p_title VARCHAR(255),
    p_priority VARCHAR(50),
    p_description TEXT DEFAULT NULL,
    p_assessment_id UUID DEFAULT NULL,
    p_control_result_id UUID DEFAULT NULL,
    p_due_date DATE DEFAULT NULL,
    p_assigned_to UUID DEFAULT NULL,
    p_created_by UUID DEFAULT NULL,
    p_estimated_effort VARCHAR(100) DEFAULT NULL,
    p_healthcare_related BOOLEAN DEFAULT FALSE
) RETURNS UUID AS $$
DECLARE
    v_action_plan_id UUID;
BEGIN
    -- Verificar se a prioridade é válida
    IF p_priority NOT IN ('critical', 'high', 'medium', 'low') THEN
        RAISE EXCEPTION 'Prioridade inválida: %. Os valores permitidos são: critical, high, medium, low', p_priority;
    END IF;
    
    -- Inserir o plano de ação
    INSERT INTO iso27001_action_plans (
        organization_id,
        assessment_id,
        control_result_id,
        title,
        description,
        priority,
        status,
        due_date,
        assigned_to,
        created_by,
        updated_by,
        estimated_effort,
        healthcare_related
    ) VALUES (
        p_organization_id,
        p_assessment_id,
        p_control_result_id,
        p_title,
        p_description,
        p_priority,
        'open',
        p_due_date,
        p_assigned_to,
        p_created_by,
        p_created_by,
        p_estimated_effort,
        p_healthcare_related
    ) RETURNING id INTO v_action_plan_id;
    
    RETURN v_action_plan_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION create_iso27001_action_plan IS 'Cria um novo plano de ação para endereçar não conformidades com ISO 27001';

-- Função para atualizar o status de um plano de ação ISO 27001
CREATE OR REPLACE FUNCTION update_iso27001_action_plan_status(
    p_action_plan_id UUID,
    p_status VARCHAR(50),
    p_updated_by UUID,
    p_completion_notes TEXT DEFAULT NULL
) RETURNS BOOLEAN AS $$
BEGIN
    -- Verificar se o status é válido
    IF p_status NOT IN ('open', 'in_progress', 'completed', 'deferred', 'canceled') THEN
        RAISE EXCEPTION 'Status inválido: %. Os valores permitidos são: open, in_progress, completed, deferred, canceled', p_status;
    END IF;
    
    -- Atualizar o status do plano de ação
    UPDATE iso27001_action_plans
    SET 
        status = p_status,
        updated_by = p_updated_by,
        updated_at = NOW(),
        completed_at = CASE WHEN p_status = 'completed' THEN NOW() ELSE completed_at END,
        completed_by = CASE WHEN p_status = 'completed' THEN p_updated_by ELSE completed_by END,
        completion_notes = CASE WHEN p_status = 'completed' AND p_completion_notes IS NOT NULL THEN p_completion_notes ELSE completion_notes END
    WHERE 
        id = p_action_plan_id;
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_iso27001_action_plan_status IS 'Atualiza o status de um plano de ação ISO 27001';

-- Função para registrar um documento ISO 27001
CREATE OR REPLACE FUNCTION register_iso27001_document(
    p_organization_id UUID,
    p_title VARCHAR(255),
    p_document_type VARCHAR(100),
    p_status VARCHAR(50),
    p_version VARCHAR(50),
    p_description TEXT DEFAULT NULL,
    p_content_url VARCHAR(255) DEFAULT NULL,
    p_storage_path VARCHAR(255) DEFAULT NULL,
    p_file_type VARCHAR(50) DEFAULT NULL,
    p_file_size BIGINT DEFAULT NULL,
    p_created_by UUID DEFAULT NULL,
    p_related_controls JSONB DEFAULT NULL,
    p_next_review_date TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    p_healthcare_specific BOOLEAN DEFAULT FALSE
) RETURNS UUID AS $$
DECLARE
    v_document_id UUID;
BEGIN
    -- Verificar se o status é válido
    IF p_status NOT IN ('draft', 'under_review', 'approved', 'published', 'retired') THEN
        RAISE EXCEPTION 'Status inválido: %. Os valores permitidos são: draft, under_review, approved, published, retired', p_status;
    END IF;
    
    -- Inserir o documento
    INSERT INTO iso27001_documents (
        organization_id,
        title,
        document_type,
        description,
        version,
        status,
        content_url,
        storage_path,
        file_type,
        file_size,
        created_by,
        updated_by,
        related_controls,
        next_review_date,
        healthcare_specific
    ) VALUES (
        p_organization_id,
        p_title,
        p_document_type,
        p_description,
        p_version,
        p_status,
        p_content_url,
        p_storage_path,
        p_file_type,
        p_file_size,
        p_created_by,
        p_created_by,
        COALESCE(p_related_controls, '[]'::JSONB),
        p_next_review_date,
        p_healthcare_specific
    ) RETURNING id INTO v_document_id;
    
    RETURN v_document_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION register_iso27001_document IS 'Registra um novo documento ISO 27001 no sistema';

-- Função para aprovar um documento ISO 27001
CREATE OR REPLACE FUNCTION approve_iso27001_document(
    p_document_id UUID,
    p_approved_by UUID,
    p_next_review_date TIMESTAMP WITH TIME ZONE DEFAULT NULL
) RETURNS BOOLEAN AS $$
BEGIN
    -- Atualizar o documento para aprovado
    UPDATE iso27001_documents
    SET 
        status = 'approved',
        approved_by = p_approved_by,
        approved_at = NOW(),
        updated_by = p_approved_by,
        updated_at = NOW(),
        next_review_date = COALESCE(p_next_review_date, next_review_date)
    WHERE 
        id = p_document_id
        AND status = 'under_review';
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION approve_iso27001_document IS 'Aprova um documento ISO 27001 que está em revisão';

-- Função para publicar um documento ISO 27001
CREATE OR REPLACE FUNCTION publish_iso27001_document(
    p_document_id UUID,
    p_updated_by UUID,
    p_next_review_date TIMESTAMP WITH TIME ZONE DEFAULT NULL
) RETURNS BOOLEAN AS $$
BEGIN
    -- Atualizar o documento para publicado
    UPDATE iso27001_documents
    SET 
        status = 'published',
        updated_by = p_updated_by,
        updated_at = NOW(),
        last_review_date = NOW(),
        next_review_date = COALESCE(p_next_review_date, next_review_date)
    WHERE 
        id = p_document_id
        AND status = 'approved';
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION publish_iso27001_document IS 'Publica um documento ISO 27001 aprovado';
