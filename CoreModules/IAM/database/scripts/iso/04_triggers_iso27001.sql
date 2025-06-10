-- INNOVABIZ - IAM ISO 27001 Triggers
-- Author: Eduardo Jeremias
-- Date: 08/05/2025
-- Version: 1.0
-- Description: Triggers para o módulo ISO 27001, garantindo integridade, auditoria e automações.

-- Configurar caminho de busca
SET search_path TO iam, public;

-- ===============================================================================
-- TRIGGERS PARA SINCRONIZAÇÃO DE AVALIAÇÕES E RESULTADOS
-- ===============================================================================

-- Função para sincronizar a pontuação geral da avaliação quando os resultados de controle são atualizados
CREATE OR REPLACE FUNCTION fn_sync_iso27001_assessment_score()
RETURNS TRIGGER AS $$
DECLARE
    v_assessment_id UUID;
    v_total_controls INT;
    v_applicable_controls INT;
    v_compliant_controls INT;
    v_partial_compliance_controls INT;
    v_partial_weight FLOAT := 0.5; -- Peso para controles com conformidade parcial
    v_score FLOAT;
BEGIN
    -- Determinar a avaliação afetada
    IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN
        v_assessment_id := NEW.assessment_id;
    ELSIF TG_OP = 'DELETE' THEN
        v_assessment_id := OLD.assessment_id;
    END IF;
    
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
        assessment_id = v_assessment_id;
    
    -- Calcular a pontuação
    IF v_applicable_controls > 0 THEN
        v_score := (v_compliant_controls + (v_partial_compliance_controls * v_partial_weight)) / v_applicable_controls * 100;
    ELSE
        v_score := NULL;
    END IF;
    
    -- Atualizar a pontuação na avaliação
    UPDATE iso27001_assessments
    SET 
        score = v_score,
        updated_at = NOW()
    WHERE 
        id = v_assessment_id;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger para sincronizar a pontuação da avaliação
DROP TRIGGER IF EXISTS trg_sync_iso27001_assessment_score ON iso27001_control_results;
CREATE TRIGGER trg_sync_iso27001_assessment_score
AFTER INSERT OR UPDATE OR DELETE ON iso27001_control_results
FOR EACH STATEMENT EXECUTE FUNCTION fn_sync_iso27001_assessment_score();

-- ===============================================================================
-- TRIGGERS PARA GERAÇÃO AUTOMÁTICA DE PLANOS DE AÇÃO
-- ===============================================================================

-- Função para gerar planos de ação automaticamente para controles não conformes
CREATE OR REPLACE FUNCTION fn_auto_generate_iso27001_action_plans()
RETURNS TRIGGER AS $$
DECLARE
    v_control_data RECORD;
    v_organization_id UUID;
    v_action_plan_id UUID;
    v_healthcare_related BOOLEAN;
    v_priority VARCHAR(50);
    v_due_date DATE;
BEGIN
    -- Só processar se o status mudou para não conforme ou parcialmente conforme
    IF (TG_OP = 'INSERT' OR OLD.status IS DISTINCT FROM NEW.status) AND 
       NEW.status IN ('non_compliant', 'partial_compliance') THEN
        
        -- Obter detalhes do controle e da avaliação
        SELECT 
            c.control_id,
            c.name,
            c.healthcare_applicability,
            a.organization_id,
            a.healthcare_specific
        INTO v_control_data
        FROM 
            iso27001_controls c,
            iso27001_assessments a
        WHERE 
            c.id = NEW.control_id
            AND a.id = NEW.assessment_id;
        
        v_organization_id := v_control_data.organization_id;
        
        -- Determinar se é relacionado à saúde
        v_healthcare_related := v_control_data.healthcare_applicability IS NOT NULL AND v_control_data.healthcare_specific;
        
        -- Definir prioridade com base no status de conformidade e importância do controle
        IF NEW.status = 'non_compliant' THEN
            IF v_healthcare_related THEN
                v_priority := 'critical';
                v_due_date := CURRENT_DATE + INTERVAL '30 days';
            ELSE
                v_priority := 'high';
                v_due_date := CURRENT_DATE + INTERVAL '60 days';
            END IF;
        ELSE -- partial_compliance
            IF v_healthcare_related THEN
                v_priority := 'high';
                v_due_date := CURRENT_DATE + INTERVAL '60 days';
            ELSE
                v_priority := 'medium';
                v_due_date := CURRENT_DATE + INTERVAL '90 days';
            END IF;
        END IF;
        
        -- Verificar se já existe um plano de ação aberto para este resultado de controle
        SELECT id INTO v_action_plan_id
        FROM iso27001_action_plans
        WHERE 
            control_result_id = NEW.id
            AND status IN ('open', 'in_progress');
        
        -- Se não existir, criar novo plano de ação
        IF v_action_plan_id IS NULL THEN
            INSERT INTO iso27001_action_plans (
                organization_id,
                assessment_id,
                control_result_id,
                title,
                description,
                priority,
                status,
                due_date,
                created_by,
                updated_by,
                healthcare_related
            ) VALUES (
                v_organization_id,
                NEW.assessment_id,
                NEW.id,
                'Implementar ' || v_control_data.control_id || ' - ' || v_control_data.name,
                'Plano de ação gerado automaticamente para implementação do controle ' || 
                v_control_data.control_id || ' - ' || v_control_data.name || 
                '. Status atual: ' || NEW.status,
                v_priority,
                'open',
                v_due_date,
                NEW.updated_by,
                NEW.updated_by,
                v_healthcare_related
            );
        END IF;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger para gerar planos de ação automaticamente
DROP TRIGGER IF EXISTS trg_auto_generate_iso27001_action_plans ON iso27001_control_results;
CREATE TRIGGER trg_auto_generate_iso27001_action_plans
AFTER INSERT OR UPDATE OF status ON iso27001_control_results
FOR EACH ROW EXECUTE FUNCTION fn_auto_generate_iso27001_action_plans();

-- ===============================================================================
-- TRIGGERS PARA VALIDAÇÃO E INTEGRIDADE DE DADOS
-- ===============================================================================

-- Função para validar e formalizar resultados de avaliação
CREATE OR REPLACE FUNCTION fn_validate_iso27001_control_result()
RETURNS TRIGGER AS $$
DECLARE
    v_control_id UUID;
    v_assessment_healthcare_specific BOOLEAN;
    v_control_healthcare_applicability TEXT;
BEGIN
    -- Obter detalhes do controle
    SELECT 
        c.id,
        c.healthcare_applicability,
        a.healthcare_specific
    INTO 
        v_control_id,
        v_control_healthcare_applicability,
        v_assessment_healthcare_specific
    FROM 
        iso27001_controls c,
        iso27001_assessments a
    WHERE 
        c.id = NEW.control_id
        AND a.id = NEW.assessment_id;
    
    -- Garantir que ambos o status e o status de implementação estão definidos
    IF NEW.status IS NULL THEN
        RAISE EXCEPTION 'O status de conformidade não pode ser nulo';
    END IF;
    
    IF NEW.implementation_status IS NULL THEN
        -- Definir automaticamente com base no status
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
    
    -- Validar score com base no status
    IF NEW.status = 'compliant' AND (NEW.score IS NULL OR NEW.score < 90) THEN
        NEW.score := 100;
    ELSIF NEW.status = 'partial_compliance' AND (NEW.score IS NULL OR NEW.score < 50 OR NEW.score > 89) THEN
        NEW.score := 75;
    ELSIF NEW.status = 'non_compliant' AND (NEW.score IS NULL OR NEW.score > 49) THEN
        NEW.score := 0;
    END IF;
    
    -- Se for uma avaliação específica de saúde, garantir que o campo healthcare_specific_findings esteja preenchido
    IF v_assessment_healthcare_specific AND v_control_healthcare_applicability IS NOT NULL AND 
       NEW.status IN ('non_compliant', 'partial_compliance') AND NEW.healthcare_specific_findings IS NULL THEN
        -- Adicionar automaticamente uma observação padrão
        NEW.healthcare_specific_findings := 'Este controle possui aplicabilidade específica para saúde. Uma avaliação detalhada é necessária.';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger para validar e formalizar resultados de avaliação
DROP TRIGGER IF EXISTS trg_validate_iso27001_control_result ON iso27001_control_results;
CREATE TRIGGER trg_validate_iso27001_control_result
BEFORE INSERT OR UPDATE ON iso27001_control_results
FOR EACH ROW EXECUTE FUNCTION fn_validate_iso27001_control_result();

-- ===============================================================================
-- TRIGGERS PARA NOTIFICAÇÕES
-- ===============================================================================

-- Função para criar notificações sobre itens expirados ou próximos do vencimento
CREATE OR REPLACE FUNCTION fn_iso27001_expiration_notifications()
RETURNS TRIGGER AS $$
DECLARE
    v_notification_id UUID;
    v_days_until_expiry INTEGER;
    v_user_ids UUID[];
BEGIN
    -- Processar documentos que exigem revisão
    IF TG_TABLE_NAME = 'iso27001_documents' AND TG_OP = 'UPDATE' THEN
        IF NEW.next_review_date IS NOT NULL AND NEW.status = 'published' THEN
            v_days_until_expiry := EXTRACT(DAY FROM (NEW.next_review_date - CURRENT_DATE));
            
            -- Se faltar 30 dias ou menos para revisão
            IF v_days_until_expiry <= 30 AND v_days_until_expiry >= 0 AND 
               (OLD.next_review_date IS NULL OR OLD.next_review_date <> NEW.next_review_date) THEN
            
                -- Obter usuários responsáveis (criador, atualizador, aprovador)
                SELECT ARRAY[NEW.created_by, NEW.updated_by, NEW.approved_by] INTO v_user_ids;
                
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
                    'document_review',
                    'Documento ISO 27001 requer revisão em breve',
                    'O documento "' || NEW.title || '" requer revisão até ' || NEW.next_review_date,
                    v_user_ids,
                    CASE 
                        WHEN v_days_until_expiry <= 7 THEN 'high'
                        WHEN v_days_until_expiry <= 14 THEN 'medium'
                        ELSE 'low'
                    END,
                    'iso27001_documents',
                    NEW.id,
                    NEW.next_review_date,
                    'pending'
                ) RETURNING id INTO v_notification_id;
            END IF;
            
            -- Se já estiver vencido
            IF v_days_until_expiry < 0 AND NEW.status = 'published' AND 
               (OLD.next_review_date IS NULL OR OLD.next_review_date <> NEW.next_review_date OR OLD.status <> 'published') THEN
                
                -- Obter usuários responsáveis (criador, atualizador, aprovador)
                SELECT ARRAY[NEW.created_by, NEW.updated_by, NEW.approved_by] INTO v_user_ids;
                
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
                    'document_review_overdue',
                    'Documento ISO 27001 com revisão vencida',
                    'A revisão do documento "' || NEW.title || '" está vencida desde ' || NEW.next_review_date,
                    v_user_ids,
                    'high',
                    'iso27001_documents',
                    NEW.id,
                    NEW.next_review_date,
                    'pending'
                ) RETURNING id INTO v_notification_id;
            END IF;
        END IF;
    END IF;
    
    -- Processar planos de ação
    IF TG_TABLE_NAME = 'iso27001_action_plans' AND (TG_OP = 'INSERT' OR TG_OP = 'UPDATE') THEN
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
                    'action_plan_due',
                    'Plano de ação ISO 27001 próximo ao vencimento',
                    'O plano de ação "' || NEW.title || '" vence em ' || v_days_until_expiry || ' dias',
                    ARRAY[NEW.assigned_to, NEW.created_by],
                    CASE 
                        WHEN v_days_until_expiry <= 2 THEN 'high'
                        WHEN v_days_until_expiry <= 5 THEN 'medium'
                        ELSE 'low'
                    END,
                    'iso27001_action_plans',
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
                    'action_plan_overdue',
                    'Plano de ação ISO 27001 vencido',
                    'O plano de ação "' || NEW.title || '" está vencido desde ' || NEW.due_date,
                    ARRAY[NEW.assigned_to, NEW.created_by],
                    'high',
                    'iso27001_action_plans',
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

-- Trigger para notificações de documentos
DROP TRIGGER IF EXISTS trg_iso27001_document_notifications ON iso27001_documents;
CREATE TRIGGER trg_iso27001_document_notifications
AFTER INSERT OR UPDATE ON iso27001_documents
FOR EACH ROW EXECUTE FUNCTION fn_iso27001_expiration_notifications();

-- Trigger para notificações de planos de ação
DROP TRIGGER IF EXISTS trg_iso27001_action_plan_notifications ON iso27001_action_plans;
CREATE TRIGGER trg_iso27001_action_plan_notifications
AFTER INSERT OR UPDATE ON iso27001_action_plans
FOR EACH ROW EXECUTE FUNCTION fn_iso27001_expiration_notifications();

-- ===============================================================================
-- TRIGGERS PARA AUDITORIA
-- ===============================================================================

-- Função para registrar alterações em auditoria
CREATE OR REPLACE FUNCTION fn_record_iso27001_audit_log()
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
    v_entity_type := 'iso27001_' || TG_TABLE_NAME;
    
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
DROP TRIGGER IF EXISTS trg_iso27001_assessments_audit ON iso27001_assessments;
CREATE TRIGGER trg_iso27001_assessments_audit
AFTER INSERT OR UPDATE OR DELETE ON iso27001_assessments
FOR EACH ROW EXECUTE FUNCTION fn_record_iso27001_audit_log();

-- Control Results
DROP TRIGGER IF EXISTS trg_iso27001_control_results_audit ON iso27001_control_results;
CREATE TRIGGER trg_iso27001_control_results_audit
AFTER INSERT OR UPDATE OR DELETE ON iso27001_control_results
FOR EACH ROW EXECUTE FUNCTION fn_record_iso27001_audit_log();

-- Action Plans
DROP TRIGGER IF EXISTS trg_iso27001_action_plans_audit ON iso27001_action_plans;
CREATE TRIGGER trg_iso27001_action_plans_audit
AFTER INSERT OR UPDATE OR DELETE ON iso27001_action_plans
FOR EACH ROW EXECUTE FUNCTION fn_record_iso27001_audit_log();

-- Documents
DROP TRIGGER IF EXISTS trg_iso27001_documents_audit ON iso27001_documents;
CREATE TRIGGER trg_iso27001_documents_audit
AFTER INSERT OR UPDATE OR DELETE ON iso27001_documents
FOR EACH ROW EXECUTE FUNCTION fn_record_iso27001_audit_log();

COMMENT ON FUNCTION fn_sync_iso27001_assessment_score IS 'Sincroniza a pontuação da avaliação ISO 27001 com base nos resultados de controles';
COMMENT ON FUNCTION fn_auto_generate_iso27001_action_plans IS 'Gera planos de ação automaticamente para controles não conformes';
COMMENT ON FUNCTION fn_validate_iso27001_control_result IS 'Valida e formaliza resultados da avaliação de controles ISO 27001';
COMMENT ON FUNCTION fn_iso27001_expiration_notifications IS 'Cria notificações para documentos e planos de ação prestes a expirar';
COMMENT ON FUNCTION fn_record_iso27001_audit_log IS 'Registra alterações nas entidades ISO 27001 no log de auditoria';
