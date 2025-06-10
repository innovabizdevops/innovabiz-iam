-- INNOVABIZ - IAM ML/Analytics Healthcare Triggers
-- Author: Eduardo Jeremias
-- Date: 08/05/2025
-- Version: 1.0
-- Description: Triggers para automação de processos de ML/analytics em saúde.

-- Configurar caminho de busca
SET search_path TO iam_analytics, iam, public;

-- ===============================================================================
-- TRIGGERS PARA ATUALIZAÇÃO AUTOMÁTICA DE FEATURES
-- ===============================================================================

-- Função para atualizar features de ISO 27001 após uma avaliação
CREATE OR REPLACE FUNCTION tg_update_iso27001_features() RETURNS TRIGGER AS $$
BEGIN
    -- Quando uma avaliação for finalizada, extrair features para ML
    IF NEW.status = 'completed' AND (OLD.status != 'completed' OR OLD.status IS NULL) THEN
        PERFORM extract_iso27001_compliance_features(NEW.organization_id, NEW.id);
        
        -- Detectar anomalias
        PERFORM detect_healthcare_compliance_anomalies(NEW.organization_id);
        
        -- Gerar recomendações automaticamente
        PERFORM generate_healthcare_iso27001_recommendations(NEW.organization_id, NEW.id);
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger para ISO 27001
CREATE TRIGGER trg_update_iso27001_features
AFTER INSERT OR UPDATE OF status, overall_score ON iam.iso_assessments
FOR EACH ROW
EXECUTE FUNCTION tg_update_iso27001_features();

-- Função para atualizar features de HIMSS EMRAM após uma avaliação
CREATE OR REPLACE FUNCTION tg_update_himss_emram_features() RETURNS TRIGGER AS $$
BEGIN
    -- Quando uma avaliação for finalizada ou o estágio mudar, extrair features para ML
    IF (NEW.status = 'completed' AND (OLD.status != 'completed' OR OLD.status IS NULL)) OR
       (NEW.current_stage != OLD.current_stage) THEN
        PERFORM extract_himss_emram_compliance_features(NEW.organization_id, NEW.id);
        
        -- Detectar anomalias
        PERFORM detect_healthcare_compliance_anomalies(NEW.organization_id);
        
        -- Gerar recomendações automaticamente
        PERFORM generate_himss_emram_recommendations(NEW.organization_id, NEW.id);
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger para HIMSS EMRAM
CREATE TRIGGER trg_update_himss_emram_features
AFTER INSERT OR UPDATE OF status, current_stage ON iam.himss_assessments
FOR EACH ROW
EXECUTE FUNCTION tg_update_himss_emram_features();

-- ===============================================================================
-- TRIGGERS PARA ANÁLISE DE MUDANÇAS EM COMPLIANCE
-- ===============================================================================

-- Função para detectar mudanças significativas em resultados de controle ISO 27001
CREATE OR REPLACE FUNCTION tg_analyze_iso_control_changes() RETURNS TRIGGER AS $$
DECLARE
    v_old_score FLOAT;
    v_new_score FLOAT;
    v_delta FLOAT;
    v_threshold FLOAT := 0.3; -- 30% de mudança considera significativa
    v_organization_id UUID;
    v_control_name TEXT;
    v_priority TEXT;
BEGIN
    -- Verificar se é uma atualização e se houve mudança no score
    IF TG_OP = 'UPDATE' AND NEW.score != OLD.score THEN
        v_old_score := OLD.score;
        v_new_score := NEW.score;
        
        -- Calcular a diferença percentual
        IF v_old_score > 0 THEN
            v_delta := ABS(v_new_score - v_old_score) / v_old_score;
        ELSE
            v_delta := 1.0; -- Se era zero, qualquer mudança é 100%
        END IF;
        
        -- Se a mudança for significativa, registrar
        IF v_delta >= v_threshold THEN
            -- Obter informações adicionais do controle
            SELECT 
                a.organization_id, 
                c.name,
                c.priority
            INTO 
                v_organization_id, 
                v_control_name,
                v_priority
            FROM 
                iam.iso_assessments a
                JOIN iam.iso_controls c ON NEW.control_id = c.id
            WHERE 
                a.id = NEW.assessment_id;
            
            -- Registrar a anomalia
            INSERT INTO ml_anomalies (
                entity_type,
                entity_id,
                anomaly_type,
                detection_model,
                anomaly_score,
                anomaly_threshold,
                features_contribution,
                detection_timestamp,
                status
            ) VALUES (
                'iso_controls',
                NEW.control_id,
                'control_score_change',
                'change_detection',
                v_delta,
                v_threshold,
                jsonb_build_object(
                    'old_score', v_old_score,
                    'new_score', v_new_score,
                    'delta', v_delta,
                    'control_name', v_control_name,
                    'assessment_id', NEW.assessment_id,
                    'organization_id', v_organization_id,
                    'change_direction', CASE WHEN v_new_score > v_old_score THEN 'increase' ELSE 'decrease' END,
                    'priority', v_priority
                ),
                NOW(),
                'open'
            );
            
            -- Para controles críticos relacionados à saúde, gerar alerta imediato
            IF v_priority = 'high' AND v_delta >= 0.4 AND (
                v_control_name ILIKE '%patient%' OR
                v_control_name ILIKE '%health%' OR
                v_control_name ILIKE '%data protection%' OR
                v_control_name ILIKE '%access control%' OR
                v_control_name ILIKE '%sensitive%'
            ) THEN
                INSERT INTO iam.notifications (
                    notification_type,
                    title,
                    message,
                    target_users,
                    priority,
                    entity_type,
                    entity_id,
                    status
                )
                SELECT
                    'compliance_alert',
                    'Alteração Crítica em Controle de Segurança',
                    'Uma mudança significativa de ' || ROUND(v_delta * 100) || '% foi detectada no controle "' || 
                    v_control_name || '". ' || CASE WHEN v_new_score < v_old_score 
                                                THEN 'Esta redução de pontuação pode impactar a proteção de dados de saúde.'
                                                ELSE 'Esta melhoria de pontuação fortalece a proteção de dados de saúde.'
                                             END,
                    ARRAY_AGG(u.id),
                    'high',
                    'iso_controls',
                    NEW.control_id,
                    'pending'
                FROM 
                    iam.users u
                    JOIN iam.roles r ON u.role_id = r.id
                WHERE 
                    u.organization_id = v_organization_id
                    AND r.name IN ('Security Officer', 'Admin', 'Compliance Manager')
                GROUP BY 
                    1, 2, 3, 5, 6, 7, 8;
            END IF;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger para análise de mudanças em controles ISO
CREATE TRIGGER trg_analyze_iso_control_changes
AFTER UPDATE OF score ON iam.iso_control_results
FOR EACH ROW
EXECUTE FUNCTION tg_analyze_iso_control_changes();

-- Função para detectar mudanças significativas em resultados de critérios HIMSS EMRAM
CREATE OR REPLACE FUNCTION tg_analyze_himss_criteria_changes() RETURNS TRIGGER AS $$
DECLARE
    v_old_score FLOAT;
    v_new_score FLOAT;
    v_old_compliant BOOLEAN;
    v_new_compliant BOOLEAN;
    v_organization_id UUID;
    v_criteria_name TEXT;
    v_stage INTEGER;
BEGIN
    -- Verificar se houve mudança na conformidade ou score
    IF TG_OP = 'UPDATE' AND (
        NEW.is_compliant != OLD.is_compliant OR 
        NEW.compliance_score != OLD.compliance_score
    ) THEN
        v_old_score := OLD.compliance_score;
        v_new_score := NEW.compliance_score;
        v_old_compliant := OLD.is_compliant;
        v_new_compliant := NEW.is_compliant;
        
        -- Obter informações adicionais do critério
        SELECT 
            a.organization_id, 
            c.name,
            NEW.stage
        INTO 
            v_organization_id, 
            v_criteria_name,
            v_stage
        FROM 
            iam.himss_assessments a
            JOIN iam.himss_criteria c ON NEW.criteria_id = c.id
        WHERE 
            a.id = NEW.assessment_id;
        
        -- Se o critério passou de não-conforme para conforme ou vice-versa
        IF NEW.is_compliant != OLD.is_compliant THEN
            -- Registrar alteração como evento de compliance
            INSERT INTO iam.compliance_events (
                organization_id,
                event_type,
                entity_type,
                entity_id,
                event_details,
                status,
                severity
            ) VALUES (
                v_organization_id,
                CASE WHEN NEW.is_compliant THEN 'criteria_compliant' ELSE 'criteria_noncompliant' END,
                'himss_criteria',
                NEW.criteria_id,
                jsonb_build_object(
                    'criteria_name', v_criteria_name,
                    'stage', v_stage,
                    'old_compliant', v_old_compliant,
                    'new_compliant', v_new_compliant,
                    'old_score', v_old_score,
                    'new_score', v_new_score,
                    'assessment_id', NEW.assessment_id
                ),
                'active',
                CASE WHEN NEW.is_compliant THEN 'info' ELSE 'warning' END
            );
            
            -- Verificar impacto no estágio EMRAM
            IF EXISTS (
                SELECT 1
                FROM iam.himss_criteria_results cr
                WHERE cr.assessment_id = NEW.assessment_id
                  AND cr.stage = NEW.stage
                GROUP BY cr.stage
                HAVING COUNT(*) FILTER (WHERE cr.is_compliant = FALSE) = 0
            ) THEN
                -- Todos os critérios do estágio estão conformes, verificar se deve atualizar o estágio atual
                UPDATE iam.himss_assessments
                SET current_stage = GREATEST(current_stage, NEW.stage)
                WHERE id = NEW.assessment_id
                  AND current_stage < NEW.stage;
            END IF;
        END IF;
        
        -- Se um critério para o próximo estágio ficou conforme, gerar recomendação
        IF NOT v_old_compliant AND v_new_compliant THEN
            -- Obter estágio atual da avaliação
            WITH assessment_stage AS (
                SELECT current_stage
                FROM iam.himss_assessments
                WHERE id = NEW.assessment_id
            )
            -- Se este critério é para o próximo estágio
            IF NEW.stage = (SELECT current_stage + 1 FROM assessment_stage) THEN
                -- Verificar quantos critérios faltam para o próximo estágio
                WITH remaining_criteria AS (
                    SELECT COUNT(*) AS remaining
                    FROM iam.himss_criteria_results
                    WHERE assessment_id = NEW.assessment_id
                      AND stage = NEW.stage
                      AND is_compliant = FALSE
                )
                -- Gerar recomendação se estiver próximo de avançar
                IF (SELECT remaining FROM remaining_criteria) <= 3 THEN
                    PERFORM generate_himss_emram_recommendations(v_organization_id, NEW.assessment_id, 3);
                END IF;
            END IF;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger para análise de mudanças em critérios HIMSS
CREATE TRIGGER trg_analyze_himss_criteria_changes
AFTER UPDATE OF is_compliant, compliance_score ON iam.himss_criteria_results
FOR EACH ROW
EXECUTE FUNCTION tg_analyze_himss_criteria_changes();

-- ===============================================================================
-- TRIGGERS PARA INTEGRAÇÃO COM VALIDADORES DE COMPLIANCE
-- ===============================================================================

-- Função para integrar resultado do validador ISO 27001 com o ML/Analytics
CREATE OR REPLACE FUNCTION tg_process_iso27001_validation_result() RETURNS TRIGGER AS $$
DECLARE
    v_features JSONB;
    v_organization_id UUID;
    v_control_ids TEXT[];
    v_control_id UUID;
    v_target_user_id UUID;
BEGIN
    -- Extrair organização alvo da validação
    v_organization_id := (NEW.validation_details->>'organization_id')::UUID;
    IF v_organization_id IS NULL THEN
        RETURN NEW;
    END IF;
    
    -- Obter um usuário alvo para as recomendações
    SELECT u.id INTO v_target_user_id
    FROM iam.users u
    JOIN iam.roles r ON u.role_id = r.id
    WHERE u.organization_id = v_organization_id
    AND r.name IN ('Security Officer', 'Compliance Manager', 'Admin')
    LIMIT 1;
    
    -- Extrair features da validação
    v_features := jsonb_build_object(
        'validator', NEW.validator_type,
        'timestamp', NEW.validation_timestamp,
        'status', NEW.validation_status,
        'score', NEW.score,
        'issues_count', jsonb_array_length(NEW.issues),
        'health_data_issues', jsonb_array_length(
            jsonb_path_query_array(
                NEW.issues, 
                '$[*] ? (@.category == "health_data" || @.description like_regex ".*saúde.*|.*medical.*|.*patient.*|.*health.*")'
            )
        ),
        'critical_issues', jsonb_array_length(
            jsonb_path_query_array(
                NEW.issues, 
                '$[*] ? (@.severity == "critical")'
            )
        ),
        'high_issues', jsonb_array_length(
            jsonb_path_query_array(
                NEW.issues, 
                '$[*] ? (@.severity == "high")'
            )
        )
    );
    
    -- Armazenar no feature store
    INSERT INTO ml_feature_store (
        feature_set_name,
        entity_type,
        entity_id,
        feature_data,
        timestamp,
        metadata
    ) VALUES (
        'validation_results',
        'organizations',
        v_organization_id,
        v_features,
        NEW.validation_timestamp,
        jsonb_build_object(
            'validation_id', NEW.id,
            'validator_type', NEW.validator_type,
            'entity_type', NEW.entity_type
        )
    );
    
    -- Para issues relacionadas a controles específicos, criar recomendações
    IF jsonb_typeof(NEW.issues) = 'array' THEN
        -- Extrair control_ids
        SELECT 
            array_agg(DISTINCT (issue->>'control_id')::TEXT)
        INTO 
            v_control_ids
        FROM 
            jsonb_array_elements(NEW.issues) AS issue
        WHERE 
            issue->>'control_id' IS NOT NULL;
        
        -- Para cada controle com problemas, criar recomendação
        IF v_control_ids IS NOT NULL AND v_target_user_id IS NOT NULL THEN
            FOREACH v_control_id IN ARRAY v_control_ids
            LOOP
                -- Filtrar issues relacionadas a este controle
                WITH control_issues AS (
                    SELECT 
                        jsonb_agg(issue) AS issues
                    FROM 
                        jsonb_array_elements(NEW.issues) AS issue
                    WHERE 
                        issue->>'control_id' = v_control_id
                )
                -- Criar recomendação
                INSERT INTO ml_recommendations (
                    target_user_id,
                    target_entity_type,
                    target_entity_id,
                    recommendation_type,
                    recommendation_model,
                    recommendation_content,
                    relevance_score,
                    explanation,
                    expires_at,
                    context
                )
                SELECT
                    v_target_user_id,
                    'iso_controls',
                    v_control_id::UUID,
                    'validation_fix',
                    'validator_recommendation',
                    jsonb_build_object(
                        'title', 'Correção de Problemas de Validação ISO 27001',
                        'control_id', v_control_id,
                        'control_name', (
                            SELECT name FROM iam.iso_controls WHERE id = v_control_id::UUID
                        ),
                        'issues', issues,
                        'recommendation_text', 'O validador ISO 27001 para saúde detectou ' || 
                                              jsonb_array_length(issues) || ' problemas relacionados a este controle. ' ||
                                              'Recomendamos revisar e resolver estas questões para garantir a conformidade.'
                    ),
                    0.9,
                    'Recomendação automática baseada em resultados de validação ISO 27001 para saúde',
                    NOW() + INTERVAL '30 days',
                    jsonb_build_object(
                        'validation_id', NEW.id,
                        'organization_id', v_organization_id,
                        'recommendation_source', 'iso27001_validator',
                        'healthcare_specific', TRUE
                    )
                FROM 
                    control_issues;
            END LOOP;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger para processar resultados de validação ISO 27001
CREATE TRIGGER trg_process_iso27001_validation_result
AFTER INSERT ON iam.validation_results
FOR EACH ROW
WHEN (NEW.validator_type = 'ISO27001HealthcareValidator')
EXECUTE FUNCTION tg_process_iso27001_validation_result();

-- Função para integrar resultado do validador HIMSS EMRAM com o ML/Analytics
CREATE OR REPLACE FUNCTION tg_process_himss_emram_validation_result() RETURNS TRIGGER AS $$
DECLARE
    v_features JSONB;
    v_organization_id UUID;
    v_criteria_ids TEXT[];
    v_criteria_id UUID;
    v_target_user_id UUID;
BEGIN
    -- Extrair organização alvo da validação
    v_organization_id := (NEW.validation_details->>'organization_id')::UUID;
    IF v_organization_id IS NULL THEN
        RETURN NEW;
    END IF;
    
    -- Obter um usuário alvo para as recomendações
    SELECT u.id INTO v_target_user_id
    FROM iam.users u
    JOIN iam.roles r ON u.role_id = r.id
    WHERE u.organization_id = v_organization_id
    AND r.name IN ('Medical Director', 'CMIO', 'IT Director', 'Admin')
    LIMIT 1;
    
    -- Extrair features da validação
    v_features := jsonb_build_object(
        'validator', NEW.validator_type,
        'timestamp', NEW.validation_timestamp,
        'status', NEW.validation_status,
        'score', NEW.score,
        'issues_count', jsonb_array_length(NEW.issues),
        'stage_issues', jsonb_build_object(
            'stage_1', jsonb_array_length(
                jsonb_path_query_array(
                    NEW.issues, 
                    '$[*] ? (@.stage == 1)'
                )
            ),
            'stage_2', jsonb_array_length(
                jsonb_path_query_array(
                    NEW.issues, 
                    '$[*] ? (@.stage == 2)'
                )
            ),
            'stage_3', jsonb_array_length(
                jsonb_path_query_array(
                    NEW.issues, 
                    '$[*] ? (@.stage == 3)'
                )
            ),
            'stage_4', jsonb_array_length(
                jsonb_path_query_array(
                    NEW.issues, 
                    '$[*] ? (@.stage == 4)'
                )
            ),
            'stage_5', jsonb_array_length(
                jsonb_path_query_array(
                    NEW.issues, 
                    '$[*] ? (@.stage == 5)'
                )
            ),
            'stage_6', jsonb_array_length(
                jsonb_path_query_array(
                    NEW.issues, 
                    '$[*] ? (@.stage == 6)'
                )
            ),
            'stage_7', jsonb_array_length(
                jsonb_path_query_array(
                    NEW.issues, 
                    '$[*] ? (@.stage == 7)'
                )
            )
        ),
        'critical_issues', jsonb_array_length(
            jsonb_path_query_array(
                NEW.issues, 
                '$[*] ? (@.severity == "critical")'
            )
        )
    );
    
    -- Armazenar no feature store
    INSERT INTO ml_feature_store (
        feature_set_name,
        entity_type,
        entity_id,
        feature_data,
        timestamp,
        metadata
    ) VALUES (
        'validation_results',
        'organizations',
        v_organization_id,
        v_features,
        NEW.validation_timestamp,
        jsonb_build_object(
            'validation_id', NEW.id,
            'validator_type', NEW.validator_type,
            'entity_type', NEW.entity_type
        )
    );
    
    -- Para issues relacionadas a critérios específicos, criar recomendações
    IF jsonb_typeof(NEW.issues) = 'array' THEN
        -- Extrair criteria_ids
        SELECT 
            array_agg(DISTINCT (issue->>'criteria_id')::TEXT)
        INTO 
            v_criteria_ids
        FROM 
            jsonb_array_elements(NEW.issues) AS issue
        WHERE 
            issue->>'criteria_id' IS NOT NULL;
        
        -- Para cada critério com problemas, criar recomendação
        IF v_criteria_ids IS NOT NULL AND v_target_user_id IS NOT NULL THEN
            FOREACH v_criteria_id IN ARRAY v_criteria_ids
            LOOP
                -- Filtrar issues relacionadas a este critério
                WITH criteria_issues AS (
                    SELECT 
                        jsonb_agg(issue) AS issues
                    FROM 
                        jsonb_array_elements(NEW.issues) AS issue
                    WHERE 
                        issue->>'criteria_id' = v_criteria_id
                )
                -- Criar recomendação
                INSERT INTO ml_recommendations (
                    target_user_id,
                    target_entity_type,
                    target_entity_id,
                    recommendation_type,
                    recommendation_model,
                    recommendation_content,
                    relevance_score,
                    explanation,
                    expires_at,
                    context
                )
                SELECT
                    v_target_user_id,
                    'himss_criteria',
                    v_criteria_id::UUID,
                    'validation_fix',
                    'validator_recommendation',
                    jsonb_build_object(
                        'title', 'Correção de Problemas de Validação HIMSS EMRAM',
                        'criteria_id', v_criteria_id,
                        'criteria_name', (
                            SELECT name FROM iam.himss_criteria WHERE id = v_criteria_id::UUID
                        ),
                        'stage', (
                            SELECT stage FROM iam.himss_criteria WHERE id = v_criteria_id::UUID
                        ),
                        'issues', issues,
                        'recommendation_text', 'O validador HIMSS EMRAM detectou ' || 
                                              jsonb_array_length(issues) || ' problemas relacionados a este critério. ' ||
                                              'Recomendamos revisar e resolver estas questões para avançar no modelo de maturidade.'
                    ),
                    0.85,
                    'Recomendação automática baseada em resultados de validação HIMSS EMRAM',
                    NOW() + INTERVAL '45 days',
                    jsonb_build_object(
                        'validation_id', NEW.id,
                        'organization_id', v_organization_id,
                        'recommendation_source', 'himss_emram_validator',
                        'healthcare_specific', TRUE
                    )
                FROM 
                    criteria_issues;
            END LOOP;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger para processar resultados de validação HIMSS EMRAM
CREATE TRIGGER trg_process_himss_emram_validation_result
AFTER INSERT ON iam.validation_results
FOR EACH ROW
WHEN (NEW.validator_type = 'HIMSSEMRAMValidator')
EXECUTE FUNCTION tg_process_himss_emram_validation_result();

-- Comentários nas funções de trigger
COMMENT ON FUNCTION tg_update_iso27001_features IS 'Trigger para extrair features de ML após finalização de avaliação ISO 27001';
COMMENT ON FUNCTION tg_update_himss_emram_features IS 'Trigger para extrair features de ML após finalização de avaliação HIMSS EMRAM';
COMMENT ON FUNCTION tg_analyze_iso_control_changes IS 'Trigger para detectar e analisar mudanças significativas em controles ISO 27001';
COMMENT ON FUNCTION tg_analyze_himss_criteria_changes IS 'Trigger para detectar e analisar mudanças em critérios HIMSS EMRAM';
COMMENT ON FUNCTION tg_process_iso27001_validation_result IS 'Trigger para processar resultados de validação ISO 27001 para saúde';
COMMENT ON FUNCTION tg_process_himss_emram_validation_result IS 'Trigger para processar resultados de validação HIMSS EMRAM';
