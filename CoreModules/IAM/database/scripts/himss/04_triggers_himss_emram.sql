-- INNOVABIZ - IAM HIMSS EMRAM Triggers
-- Author: Eduardo Jeremias
-- Date: 08/05/2025
-- Version: 1.0
-- Description: Triggers para o módulo HIMSS EMRAM, garantindo integridade, auditoria e automações.

-- Configurar caminho de busca
SET search_path TO iam, public;

-- ===============================================================================
-- TRIGGERS PARA CÁLCULO AUTOMÁTICO DE ESTÁGIO EMRAM
-- ===============================================================================

-- Função para calcular e atualizar o estágio EMRAM com base nos resultados de critérios
CREATE OR REPLACE FUNCTION fn_calculate_himss_emram_stage()
RETURNS TRIGGER AS $$
DECLARE
    v_assessment_id UUID;
    v_current_stage INTEGER := 0;
    v_stage_compliance RECORD;
    v_target_stage INTEGER;
BEGIN
    -- Determinar a avaliação afetada
    IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN
        v_assessment_id := NEW.assessment_id;
    ELSIF TG_OP = 'DELETE' THEN
        v_assessment_id := OLD.assessment_id;
    END IF;
    
    -- Obter o estágio alvo da avaliação
    SELECT target_stage INTO v_target_stage
    FROM himss_emram_assessments
    WHERE id = v_assessment_id;
    
    -- Calcular conformidade por estágio
    FOR v_stage_compliance IN (
        SELECT 
            hes.stage_number,
            COUNT(cr.id) AS total_criteria,
            COUNT(cr.id) FILTER (WHERE cr.status = 'compliant') AS compliant_criteria,
            COUNT(cr.id) FILTER (WHERE cr.status = 'partial_compliance') AS partial_criteria,
            COUNT(cr.id) FILTER (WHERE cr.is_mandatory = TRUE) AS mandatory_criteria,
            COUNT(cr.id) FILTER (WHERE cr.is_mandatory = TRUE AND cr.status = 'compliant') AS compliant_mandatory_criteria
        FROM 
            himss_emram_stages hes
            JOIN himss_emram_criteria hec ON hes.id = hec.stage_id
            JOIN himss_emram_criteria_results cr ON hec.id = cr.criteria_id
        WHERE 
            cr.assessment_id = v_assessment_id
        GROUP BY 
            hes.stage_number
        ORDER BY 
            hes.stage_number
    ) LOOP
        -- Verificar se atingiu os requisitos para o estágio
        -- Para cada estágio ser alcançado:
        -- 1. Todos os critérios obrigatórios devem estar conformes
        -- 2. Pelo menos 70% dos critérios totais devem estar conformes ou parcialmente conformes
        
        IF v_stage_compliance.compliant_mandatory_criteria = v_stage_compliance.mandatory_criteria AND
           (v_stage_compliance.compliant_criteria + v_stage_compliance.partial_criteria) >= (v_stage_compliance.total_criteria * 0.7) THEN
            -- Este estágio foi atingido
            v_current_stage := v_stage_compliance.stage_number;
        ELSE
            -- Se um estágio não for atingido, interromper a verificação (modelo cumulativo)
            EXIT;
        END IF;
    END LOOP;
    
    -- Estágio calculado não pode ser maior que o estágio alvo
    IF v_current_stage > v_target_stage THEN
        v_current_stage := v_target_stage;
    END IF;
    
    -- Atualizar o estágio atual na avaliação
    UPDATE himss_emram_assessments
    SET 
        current_stage = v_current_stage,
        updated_at = NOW()
    WHERE 
        id = v_assessment_id;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger para calcular estágio EMRAM
DROP TRIGGER IF EXISTS trg_calculate_himss_emram_stage ON himss_emram_criteria_results;
CREATE TRIGGER trg_calculate_himss_emram_stage
AFTER INSERT OR UPDATE OR DELETE ON himss_emram_criteria_results
FOR EACH STATEMENT EXECUTE FUNCTION fn_calculate_himss_emram_stage();

-- ===============================================================================
-- TRIGGERS PARA GERAÇÃO AUTOMÁTICA DE PLANOS DE AÇÃO
-- ===============================================================================

-- Função para gerar planos de ação automaticamente para critérios não conformes
CREATE OR REPLACE FUNCTION fn_auto_generate_himss_emram_action_plans()
RETURNS TRIGGER AS $$
DECLARE
    v_criteria_data RECORD;
    v_organization_id UUID;
    v_action_plan_id UUID;
    v_priority VARCHAR(50);
    v_due_date DATE;
    v_stage_number INTEGER;
BEGIN
    -- Só processar se o status mudou para não conforme ou parcialmente conforme
    IF (TG_OP = 'INSERT' OR OLD.status IS DISTINCT FROM NEW.status) AND 
       NEW.status IN ('non_compliant', 'partial_compliance') THEN
        
        -- Obter detalhes do critério e da avaliação
        SELECT 
            hec.criteria_code,
            hec.name,
            hec.is_mandatory,
            hes.stage_number,
            a.organization_id,
            a.target_stage
        INTO v_criteria_data
        FROM 
            himss_emram_criteria hec
            JOIN himss_emram_stages hes ON hec.stage_id = hes.id
            JOIN himss_emram_assessments a ON NEW.assessment_id = a.id
        WHERE 
            hec.id = NEW.criteria_id;
        
        v_organization_id := v_criteria_data.organization_id;
        v_stage_number := v_criteria_data.stage_number;
        
        -- Definir prioridade com base no status de conformidade, obrigatoriedade e estágio alvo
        IF v_criteria_data.is_mandatory AND v_stage_number <= v_criteria_data.target_stage THEN
            IF NEW.status = 'non_compliant' THEN
                v_priority := 'critical';
                v_due_date := CURRENT_DATE + INTERVAL '30 days';
            ELSE -- partial_compliance
                v_priority := 'high';
                v_due_date := CURRENT_DATE + INTERVAL '45 days';
            END IF;
        ELSIF v_stage_number <= v_criteria_data.target_stage THEN
            IF NEW.status = 'non_compliant' THEN
                v_priority := 'high';
                v_due_date := CURRENT_DATE + INTERVAL '60 days';
            ELSE -- partial_compliance
                v_priority := 'medium';
                v_due_date := CURRENT_DATE + INTERVAL '90 days';
            END IF;
        ELSE -- Estágio acima do alvo
            IF NEW.status = 'non_compliant' THEN
                v_priority := 'medium';
                v_due_date := CURRENT_DATE + INTERVAL '120 days';
            ELSE -- partial_compliance
                v_priority := 'low';
                v_due_date := CURRENT_DATE + INTERVAL '180 days';
            END IF;
        END IF;
        
        -- Verificar se já existe um plano de ação aberto para este resultado de critério
        SELECT id INTO v_action_plan_id
        FROM himss_emram_action_plans
        WHERE 
            criteria_result_id = NEW.id
            AND status IN ('open', 'in_progress');
        
        -- Se não existir, criar novo plano de ação
        IF v_action_plan_id IS NULL THEN
            INSERT INTO himss_emram_action_plans (
                organization_id,
                assessment_id,
                criteria_result_id,
                title,
                description,
                priority,
                status,
                due_date,
                created_by,
                updated_by,
                target_stage
            ) VALUES (
                v_organization_id,
                NEW.assessment_id,
                NEW.id,
                'Implementar ' || v_criteria_data.criteria_code || ' - ' || v_criteria_data.name,
                'Plano de ação gerado automaticamente para implementação do critério EMRAM ' || 
                v_criteria_data.criteria_code || ' - ' || v_criteria_data.name || 
                ' (Estágio ' || v_stage_number || '). Status atual: ' || NEW.status,
                v_priority,
                'open',
                v_due_date,
                NEW.updated_by,
                NEW.updated_by,
                v_stage_number
            );
        END IF;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger para gerar planos de ação automaticamente
DROP TRIGGER IF EXISTS trg_auto_generate_himss_emram_action_plans ON himss_emram_criteria_results;
CREATE TRIGGER trg_auto_generate_himss_emram_action_plans
AFTER INSERT OR UPDATE OF status ON himss_emram_criteria_results
FOR EACH ROW EXECUTE FUNCTION fn_auto_generate_himss_emram_action_plans();

-- ===============================================================================
-- TRIGGERS PARA VALIDAÇÃO E INTEGRIDADE DE DADOS
-- ===============================================================================

-- Função para validar e formalizar resultados de avaliação
CREATE OR REPLACE FUNCTION fn_validate_himss_emram_criteria_result()
RETURNS TRIGGER AS $$
DECLARE
    v_criteria_id UUID;
    v_criteria_code VARCHAR(50);
    v_is_mandatory BOOLEAN;
    v_stage_number INTEGER;
BEGIN
    -- Obter detalhes do critério
    SELECT 
        hec.id,
        hec.criteria_code,
        hec.is_mandatory,
        hes.stage_number
    INTO 
        v_criteria_id,
        v_criteria_code,
        v_is_mandatory,
        v_stage_number
    FROM 
        himss_emram_criteria hec
        JOIN himss_emram_stages hes ON hec.stage_id = hes.id
    WHERE 
        hec.id = NEW.criteria_id;
    
    -- Garantir que o status está definido
    IF NEW.status IS NULL THEN
        RAISE EXCEPTION 'O status de conformidade não pode ser nulo';
    END IF;
    
    -- Definir automaticamente o status de implementação com base no status
    IF NEW.implementation_status IS NULL THEN
        IF NEW.status = 'compliant' THEN
            NEW.implementation_status := 'implemented';
        ELSIF NEW.status = 'partial_compliance' THEN
            NEW.implementation_status := 'partially_implemented';
        ELSIF NEW.status = 'non_compliant' THEN
            NEW.implementation_status := 'not_implemented';
        ELSIF NEW.status = 'not_applicable' THEN
            NEW.implementation_status := 'not_applicable';
        END IF;
    END IF;
    
    -- Validar coerência entre status e status de implementação
    IF (NEW.status = 'compliant' AND NEW.implementation_status != 'implemented') OR
       (NEW.status = 'not_applicable' AND NEW.implementation_status != 'not_applicable') THEN
        RAISE EXCEPTION 'Status de conformidade e status de implementação incompatíveis';
    END IF;
    
    -- Definir ou ajustar a porcentagem de conformidade com base no status
    IF NEW.status = 'compliant' AND (NEW.compliance_percentage IS NULL OR NEW.compliance_percentage < 90) THEN
        NEW.compliance_percentage := 100;
    ELSIF NEW.status = 'partial_compliance' AND (NEW.compliance_percentage IS NULL OR NEW.compliance_percentage < 30 OR NEW.compliance_percentage > 89) THEN
        NEW.compliance_percentage := 60;
    ELSIF NEW.status = 'non_compliant' AND (NEW.compliance_percentage IS NULL OR NEW.compliance_percentage > 29) THEN
        NEW.compliance_percentage := 0;
    ELSIF NEW.status = 'not_applicable' THEN
        NEW.compliance_percentage := NULL;
    END IF;
    
    -- Validações específicas para critérios obrigatórios
    IF v_is_mandatory AND NEW.status = 'not_applicable' THEN
        RAISE EXCEPTION 'Critério % (%): Critérios obrigatórios não podem ser marcados como não aplicáveis', v_criteria_code, v_stage_number;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger para validar e formalizar resultados de avaliação
DROP TRIGGER IF EXISTS trg_validate_himss_emram_criteria_result ON himss_emram_criteria_results;
CREATE TRIGGER trg_validate_himss_emram_criteria_result
BEFORE INSERT OR UPDATE ON himss_emram_criteria_results
FOR EACH ROW EXECUTE FUNCTION fn_validate_himss_emram_criteria_result();

-- ===============================================================================
-- TRIGGERS PARA CERTIFICAÇÕES EMRAM
-- ===============================================================================

-- Função para gerenciar certificações EMRAM quando avaliações são finalizadas
CREATE OR REPLACE FUNCTION fn_manage_himss_emram_certification()
RETURNS TRIGGER AS $$
DECLARE
    v_certification_id UUID;
    v_organization_id UUID;
    v_healthcare_facility_name VARCHAR(255);
BEGIN
    -- Apenas processar quando uma avaliação é marcada como concluída
    IF OLD.status != 'completed' AND NEW.status = 'completed' AND NEW.current_stage IS NOT NULL THEN
        
        -- Obter informações da organização e estabelecimento
        v_organization_id := NEW.organization_id;
        v_healthcare_facility_name := NEW.healthcare_facility_name;
        
        -- Verificar se já existe uma certificação ativa para esta avaliação
        SELECT id INTO v_certification_id
        FROM himss_emram_certifications
        WHERE assessment_id = NEW.id AND status = 'active';
        
        -- Se não existir, criar nova certificação
        IF v_certification_id IS NULL THEN
            INSERT INTO himss_emram_certifications (
                assessment_id,
                organization_id,
                healthcare_facility_name,
                certification_date,
                expiration_date,
                stage_achieved,
                certificate_number,
                certifying_body,
                certifying_assessor,
                status,
                created_by
            ) VALUES (
                NEW.id,
                v_organization_id,
                v_healthcare_facility_name,
                NOW(),
                NOW() + INTERVAL '2 years',
                NEW.current_stage,
                'EMRAM-' || TO_CHAR(NOW(), 'YYYYMMDD') || '-' || LEFT(MD5(NEW.id::TEXT), 6),
                'HIMSS Analytics',
                NEW.updated_by::TEXT,
                'active',
                NEW.updated_by
            );
        ELSE
            -- Se existir, atualizar o estágio e a data
            UPDATE himss_emram_certifications
            SET 
                stage_achieved = NEW.current_stage,
                certification_date = NOW(),
                expiration_date = NOW() + INTERVAL '2 years',
                updated_at = NOW()
            WHERE id = v_certification_id;
        END IF;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger para gerenciar certificações
DROP TRIGGER IF EXISTS trg_manage_himss_emram_certification ON himss_emram_assessments;
CREATE TRIGGER trg_manage_himss_emram_certification
AFTER UPDATE OF status ON himss_emram_assessments
FOR EACH ROW EXECUTE FUNCTION fn_manage_himss_emram_certification();

-- ===============================================================================
-- TRIGGERS PARA NOTIFICAÇÕES
-- ===============================================================================

-- Função para criar notificações sobre certificações próximas do vencimento
CREATE OR REPLACE FUNCTION fn_himss_emram_expiration_notifications()
RETURNS TRIGGER AS $$
DECLARE
    v_notification_id UUID;
    v_days_until_expiry INTEGER;
    v_user_ids UUID[];
    v_admin_users UUID[];
BEGIN
    -- Processar certificações
    IF TG_TABLE_NAME = 'himss_emram_certifications' THEN
        IF NEW.expiration_date IS NOT NULL AND NEW.status = 'active' THEN
            v_days_until_expiry := EXTRACT(DAY FROM (NEW.expiration_date - CURRENT_DATE));
            
            -- Obter administradores da organização
            SELECT ARRAY_AGG(u.id) INTO v_admin_users
            FROM users u
            JOIN user_roles ur ON u.id = ur.user_id
            JOIN roles r ON ur.role_id = r.id
            WHERE u.organization_id = NEW.organization_id AND r.name = 'Admin';
            
            -- Se faltar 180 dias (6 meses) ou menos para expiração
            IF v_days_until_expiry <= 180 AND v_days_until_expiry >= 0 AND 
               (TG_OP = 'INSERT' OR OLD.expiration_date IS NULL OR OLD.expiration_date <> NEW.expiration_date) THEN
                
                -- Criar notificação
                INSERT INTO notifications (
                    notification_type,
                    title,
                    message,
                    target_users,
                    priority,
                    entity_type,
                    entity_id,
                    expiration_date,
                    status
                ) VALUES (
                    'emram_certification_expiring',
                    'Certificação EMRAM Estágio ' || NEW.stage_achieved || ' expira em breve',
                    'A certificação EMRAM Estágio ' || NEW.stage_achieved || ' para ' || NEW.healthcare_facility_name || 
                    ' expira em ' || v_days_until_expiry || ' dias (' || NEW.expiration_date || ')',
                    v_admin_users,
                    CASE 
                        WHEN v_days_until_expiry <= 30 THEN 'high'
                        WHEN v_days_until_expiry <= 90 THEN 'medium'
                        ELSE 'low'
                    END,
                    'himss_emram_certifications',
                    NEW.id,
                    NEW.expiration_date,
                    'pending'
                ) RETURNING id INTO v_notification_id;
            END IF;
            
            -- Se já estiver expirado ou prestes a expirar (menos de 30 dias)
            IF (v_days_until_expiry < 0 OR v_days_until_expiry <= 30) AND NEW.status = 'active' AND 
               (TG_OP = 'INSERT' OR OLD.expiration_date IS NULL OR OLD.expiration_date <> NEW.expiration_date OR OLD.status <> 'active') THEN
                
                -- Atualizar automaticamente o status se expirou
                IF v_days_until_expiry < 0 THEN
                    -- Atualizar para expirado
                    UPDATE himss_emram_certifications
                    SET 
                        status = 'expired',
                        updated_at = NOW()
                    WHERE id = NEW.id;
                    
                    -- Criar notificação para informar sobre a expiração
                    INSERT INTO notifications (
                        notification_type,
                        title,
                        message,
                        target_users,
                        priority,
                        entity_type,
                        entity_id,
                        status
                    ) VALUES (
                        'emram_certification_expired',
                        'Certificação EMRAM Estágio ' || NEW.stage_achieved || ' expirou',
                        'A certificação EMRAM Estágio ' || NEW.stage_achieved || ' para ' || NEW.healthcare_facility_name || 
                        ' expirou em ' || NEW.expiration_date,
                        v_admin_users,
                        'high',
                        'himss_emram_certifications',
                        NEW.id,
                        'pending'
                    );
                END IF;
            END IF;
        END IF;
    END IF;
    
    -- Processar planos de ação
    IF TG_TABLE_NAME = 'himss_emram_action_plans' AND (TG_OP = 'INSERT' OR TG_OP = 'UPDATE') THEN
        IF NEW.due_date IS NOT NULL AND NEW.status IN ('open', 'in_progress') THEN
            v_days_until_expiry := EXTRACT(DAY FROM (NEW.due_date - CURRENT_DATE));
            
            -- Se faltar 7 dias ou menos para o vencimento
            IF v_days_until_expiry <= 7 AND v_days_until_expiry >= 0 AND 
               (TG_OP = 'INSERT' OR OLD.due_date IS NULL OR OLD.due_date <> NEW.due_date) THEN
                
                -- Criar notificação
                INSERT INTO notifications (
                    notification_type,
                    title,
                    message,
                    target_users,
                    priority,
                    entity_type,
                    entity_id,
                    expiration_date,
                    status
                ) VALUES (
                    'emram_action_plan_due',
                    'Plano de ação EMRAM Estágio ' || NEW.target_stage || ' próximo ao vencimento',
                    'O plano de ação "' || NEW.title || '" para atingir o Estágio ' || NEW.target_stage || ' vence em ' || v_days_until_expiry || ' dias',
                    ARRAY[NEW.assigned_to, NEW.created_by],
                    CASE 
                        WHEN v_days_until_expiry <= 2 THEN 'high'
                        WHEN v_days_until_expiry <= 5 THEN 'medium'
                        ELSE 'low'
                    END,
                    'himss_emram_action_plans',
                    NEW.id,
                    NEW.due_date,
                    'pending'
                ) RETURNING id INTO v_notification_id;
            END IF;
            
            -- Se já estiver vencido
            IF v_days_until_expiry < 0 AND NEW.status IN ('open', 'in_progress') AND 
               (TG_OP = 'INSERT' OR OLD.due_date IS NULL OR OLD.due_date <> NEW.due_date OR OLD.status NOT IN ('open', 'in_progress')) THEN
                
                -- Criar notificação
                INSERT INTO notifications (
                    notification_type,
                    title,
                    message,
                    target_users,
                    priority,
                    entity_type,
                    entity_id,
                    expiration_date,
                    status
                ) VALUES (
                    'emram_action_plan_overdue',
                    'Plano de ação EMRAM Estágio ' || NEW.target_stage || ' vencido',
                    'O plano de ação "' || NEW.title || '" para atingir o Estágio ' || NEW.target_stage || ' está vencido desde ' || NEW.due_date,
                    ARRAY[NEW.assigned_to, NEW.created_by],
                    'high',
                    'himss_emram_action_plans',
                    NEW.id,
                    NEW.due_date,
                    'pending'
                ) RETURNING id INTO v_notification_id;
            END IF;
        END IF;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger para notificações de certificações
DROP TRIGGER IF EXISTS trg_himss_emram_certification_notifications ON himss_emram_certifications;
CREATE TRIGGER trg_himss_emram_certification_notifications
AFTER INSERT OR UPDATE ON himss_emram_certifications
FOR EACH ROW EXECUTE FUNCTION fn_himss_emram_expiration_notifications();

-- Trigger para notificações de planos de ação
DROP TRIGGER IF EXISTS trg_himss_emram_action_plan_notifications ON himss_emram_action_plans;
CREATE TRIGGER trg_himss_emram_action_plan_notifications
AFTER INSERT OR UPDATE ON himss_emram_action_plans
FOR EACH ROW EXECUTE FUNCTION fn_himss_emram_expiration_notifications();

-- ===============================================================================
-- TRIGGERS PARA AUDITORIA
-- ===============================================================================

-- Função para registrar alterações em auditoria
CREATE OR REPLACE FUNCTION fn_record_himss_emram_audit_log()
RETURNS TRIGGER AS $$
DECLARE
    v_action VARCHAR(10);
    v_old_data JSONB := NULL;
    v_new_data JSONB := NULL;
    v_changed_fields JSONB := NULL;
    v_user_id UUID;
    v_entity_type VARCHAR;
BEGIN
    -- Determinar a ação realizada
    IF TG_OP = 'INSERT' THEN
        v_action := 'INSERT';
        v_new_data := to_jsonb(NEW);
    ELSIF TG_OP = 'UPDATE' THEN
        v_action := 'UPDATE';
        v_old_data := to_jsonb(OLD);
        v_new_data := to_jsonb(NEW);
        v_changed_fields := jsonb_object_agg(key, value) 
            FROM jsonb_each(v_new_data) 
            WHERE NOT v_old_data ? key OR v_old_data->key <> value;
    ELSIF TG_OP = 'DELETE' THEN
        v_action := 'DELETE';
        v_old_data := to_jsonb(OLD);
    END IF;

    -- Definir tipo de entidade
    v_entity_type := 'himss_emram_' || TG_TABLE_NAME;
    
    -- Obter ID do usuário atual
    v_user_id := current_setting('app.current_user_id', TRUE)::UUID;
    
    -- Inserir o registro de auditoria
    INSERT INTO audit_logs (
        entity_type,
        entity_id,
        action,
        old_data,
        new_data,
        changed_fields,
        user_id,
        application_name,
        timestamp
    ) VALUES (
        v_entity_type,
        CASE 
            WHEN TG_OP = 'DELETE' THEN (v_old_data->>'id')::UUID
            ELSE (v_new_data->>'id')::UUID
        END,
        v_action,
        v_old_data,
        v_new_data,
        v_changed_fields,
        v_user_id,
        current_setting('application_name', TRUE),
        now()
    );
    
    -- Retornar conforme a operação
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Triggers de auditoria para as principais tabelas

-- Assessments
DROP TRIGGER IF EXISTS trg_himss_emram_assessments_audit ON himss_emram_assessments;
CREATE TRIGGER trg_himss_emram_assessments_audit
AFTER INSERT OR UPDATE OR DELETE ON himss_emram_assessments
FOR EACH ROW EXECUTE FUNCTION fn_record_himss_emram_audit_log();

-- Criteria Results
DROP TRIGGER IF EXISTS trg_himss_emram_criteria_results_audit ON himss_emram_criteria_results;
CREATE TRIGGER trg_himss_emram_criteria_results_audit
AFTER INSERT OR UPDATE OR DELETE ON himss_emram_criteria_results
FOR EACH ROW EXECUTE FUNCTION fn_record_himss_emram_audit_log();

-- Certifications
DROP TRIGGER IF EXISTS trg_himss_emram_certifications_audit ON himss_emram_certifications;
CREATE TRIGGER trg_himss_emram_certifications_audit
AFTER INSERT OR UPDATE OR DELETE ON himss_emram_certifications
FOR EACH ROW EXECUTE FUNCTION fn_record_himss_emram_audit_log();

-- Action Plans
DROP TRIGGER IF EXISTS trg_himss_emram_action_plans_audit ON himss_emram_action_plans;
CREATE TRIGGER trg_himss_emram_action_plans_audit
AFTER INSERT OR UPDATE OR DELETE ON himss_emram_action_plans
FOR EACH ROW EXECUTE FUNCTION fn_record_himss_emram_audit_log();

COMMENT ON FUNCTION fn_calculate_himss_emram_stage IS 'Calcula e atualiza o estágio EMRAM com base nos resultados de critérios';
COMMENT ON FUNCTION fn_auto_generate_himss_emram_action_plans IS 'Gera planos de ação automaticamente para critérios não conformes';
COMMENT ON FUNCTION fn_validate_himss_emram_criteria_result IS 'Valida e formaliza resultados da avaliação de critérios EMRAM';
COMMENT ON FUNCTION fn_manage_himss_emram_certification IS 'Gerencia certificações quando avaliações EMRAM são finalizadas';
COMMENT ON FUNCTION fn_himss_emram_expiration_notifications IS 'Cria notificações para certificações e planos de ação prestes a expirar';
COMMENT ON FUNCTION fn_record_himss_emram_audit_log IS 'Registra alterações nas entidades HIMSS EMRAM no log de auditoria';
