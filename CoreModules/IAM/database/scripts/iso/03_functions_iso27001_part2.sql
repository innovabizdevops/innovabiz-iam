-- INNOVABIZ - IAM ISO 27001 Functions (Parte 2)
-- Author: Eduardo Jeremias
-- Date: 08/05/2025
-- Version: 1.0
-- Description: Funções para gerenciamento de compliance ISO 27001 com foco em saúde (continuação).

-- Configurar caminho de busca
SET search_path TO iam, public;

-- Função para adicionar ou atualizar um mapeamento entre ISO 27001 e outro framework
CREATE OR REPLACE FUNCTION upsert_iso27001_framework_mapping(
    p_iso_control_id UUID,
    p_framework_id UUID,
    p_framework_control_id VARCHAR(100),
    p_framework_control_name VARCHAR(255) DEFAULT NULL,
    p_mapping_type VARCHAR(50),
    p_mapping_strength VARCHAR(50) DEFAULT NULL,
    p_notes TEXT DEFAULT NULL,
    p_created_by UUID DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    v_existing_mapping_id UUID;
    v_mapping_id UUID;
BEGIN
    -- Verificar se o tipo de mapeamento é válido
    IF p_mapping_type NOT IN ('one_to_one', 'one_to_many', 'partial', 'implied') THEN
        RAISE EXCEPTION 'Tipo de mapeamento inválido: %. Os valores permitidos são: one_to_one, one_to_many, partial, implied', p_mapping_type;
    END IF;
    
    -- Verificar se a força do mapeamento é válida
    IF p_mapping_strength IS NOT NULL AND p_mapping_strength NOT IN ('strong', 'moderate', 'weak') THEN
        RAISE EXCEPTION 'Força de mapeamento inválida: %. Os valores permitidos são: strong, moderate, weak', p_mapping_strength;
    END IF;

    -- Verificar se já existe um mapeamento para esta combinação
    SELECT id INTO v_existing_mapping_id
    FROM iso27001_framework_mapping
    WHERE iso_control_id = p_iso_control_id
      AND framework_id = p_framework_id
      AND framework_control_id = p_framework_control_id;
    
    -- Se já existe, atualizar
    IF v_existing_mapping_id IS NOT NULL THEN
        UPDATE iso27001_framework_mapping
        SET framework_control_name = COALESCE(p_framework_control_name, framework_control_name),
            mapping_type = p_mapping_type,
            mapping_strength = COALESCE(p_mapping_strength, mapping_strength),
            notes = COALESCE(p_notes, notes),
            updated_at = NOW()
        WHERE id = v_existing_mapping_id;
        
        v_mapping_id := v_existing_mapping_id;
    ELSE
        -- Caso contrário, inserir novo
        INSERT INTO iso27001_framework_mapping (
            iso_control_id,
            framework_id,
            framework_control_id,
            framework_control_name,
            mapping_type,
            mapping_strength,
            notes,
            created_by
        ) VALUES (
            p_iso_control_id,
            p_framework_id,
            p_framework_control_id,
            p_framework_control_name,
            p_mapping_type,
            p_mapping_strength,
            p_notes,
            p_created_by
        ) RETURNING id INTO v_mapping_id;
    END IF;
    
    RETURN v_mapping_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION upsert_iso27001_framework_mapping IS 'Adiciona ou atualiza um mapeamento entre controle ISO 27001 e outro framework regulatório';

-- Função para gerar recomendações automáticas com base no resultado de um controle
CREATE OR REPLACE FUNCTION generate_iso27001_recommendations(
    p_control_result_id UUID,
    p_healthcare_specific BOOLEAN DEFAULT FALSE
) RETURNS JSONB AS $$
DECLARE
    v_control_id UUID;
    v_control_code VARCHAR(50);
    v_control_name VARCHAR(255);
    v_status VARCHAR(50);
    v_implementation_status VARCHAR(50);
    v_healthcare_applicability TEXT;
    v_implementation_guidance TEXT;
    v_issues_found JSONB;
    v_recommendations JSONB := '[]'::JSONB;
BEGIN
    -- Obter informações do resultado do controle e controle
    SELECT 
        cr.control_id,
        c.control_id,
        c.name,
        cr.status,
        cr.implementation_status,
        c.healthcare_applicability,
        c.implementation_guidance,
        cr.issues_found
    INTO 
        v_control_id,
        v_control_code,
        v_control_name,
        v_status,
        v_implementation_status,
        v_healthcare_applicability,
        v_implementation_guidance,
        v_issues_found
    FROM 
        iso27001_control_results cr
        JOIN iso27001_controls c ON cr.control_id = c.id
    WHERE 
        cr.id = p_control_result_id;
    
    -- Base de recomendações comuns
    IF v_status = 'non_compliant' OR v_status = 'partial_compliance' THEN
        -- Recomendações gerais com base no status de implementação
        CASE v_implementation_status
            WHEN 'not_implemented' THEN
                v_recommendations := jsonb_build_array(
                    jsonb_build_object(
                        'recommendation_type', 'implementation',
                        'text', 'Desenvolva e implemente uma política ou procedimento formal para atender ao controle ' || v_control_code || ' - ' || v_control_name,
                        'priority', 'high'
                    ),
                    jsonb_build_object(
                        'recommendation_type', 'training',
                        'text', 'Treine a equipe sobre os requisitos e a importância deste controle',
                        'priority', 'medium'
                    ),
                    jsonb_build_object(
                        'recommendation_type', 'documentation',
                        'text', 'Documente a abordagem para implementar este controle dentro da organização',
                        'priority', 'medium'
                    )
                );
                
            WHEN 'partially_implemented' THEN
                v_recommendations := jsonb_build_array(
                    jsonb_build_object(
                        'recommendation_type', 'enhancement',
                        'text', 'Fortaleça as medidas existentes para ' || v_control_code || ' - ' || v_control_name || ' para garantir conformidade total',
                        'priority', 'medium'
                    ),
                    jsonb_build_object(
                        'recommendation_type', 'monitoring',
                        'text', 'Implemente monitoramento e medição contínuos para verificar a eficácia',
                        'priority', 'medium'
                    ),
                    jsonb_build_object(
                        'recommendation_type', 'review',
                        'text', 'Revise os processos atuais para identificar e corrigir lacunas',
                        'priority', 'medium'
                    )
                );
                
            ELSE 
                v_recommendations := jsonb_build_array(
                    jsonb_build_object(
                        'recommendation_type', 'review',
                        'text', 'Revise a implementação do controle ' || v_control_code || ' - ' || v_control_name || ' para identificar áreas de melhoria',
                        'priority', 'medium'
                    )
                );
        END CASE;
        
        -- Adicionar recomendações específicas de saúde, se aplicável
        IF p_healthcare_specific AND v_healthcare_applicability IS NOT NULL THEN
            v_recommendations := v_recommendations || jsonb_build_array(
                jsonb_build_object(
                    'recommendation_type', 'healthcare_specific',
                    'text', 'Implementar garantias adicionais específicas para dados de saúde: ' || 
                            COALESCE(v_implementation_guidance, 'Desenvolva procedimentos específicos para proteger dados sensíveis de saúde e garantir conformidade com regulamentações do setor'),
                    'priority', 'high'
                )
            );
        END IF;
        
        -- Se houver problemas específicos encontrados, adicionar recomendações baseadas neles
        IF v_issues_found IS NOT NULL AND jsonb_array_length(v_issues_found) > 0 THEN
            FOR i IN 0..jsonb_array_length(v_issues_found)-1 LOOP
                v_recommendations := v_recommendations || jsonb_build_array(
                    jsonb_build_object(
                        'recommendation_type', 'issue_specific',
                        'text', 'Resolva a questão: ' || jsonb_extract_path_text(v_issues_found, i::text, 'description') || 
                               ' através de ações corretivas específicas',
                        'priority', 'high',
                        'related_issue', jsonb_extract_path(v_issues_found, i::text)
                    )
                );
            END LOOP;
        END IF;
    END IF;
    
    RETURN v_recommendations;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION generate_iso27001_recommendations IS 'Gera recomendações automáticas com base no resultado da avaliação de um controle ISO 27001';

-- Função para gerar um relatório de resumo de compliance ISO 27001
CREATE OR REPLACE FUNCTION generate_iso27001_compliance_report(
    p_assessment_id UUID,
    p_format VARCHAR DEFAULT 'JSON'
) RETURNS JSONB AS $$
DECLARE
    v_assessment record;
    v_results record;
    v_category_compliance_stats JSONB;
    v_action_plan_stats JSONB;
    v_recommendations JSONB;
    v_report JSONB;
BEGIN
    -- Obter informações da avaliação
    SELECT 
        a.*,
        o.name AS organization_name,
        u.full_name AS created_by_name,
        rf.name AS framework_name
    INTO v_assessment
    FROM 
        iso27001_assessments a
        JOIN organizations o ON a.organization_id = o.id
        LEFT JOIN users u ON a.created_by = u.id
        LEFT JOIN regulatory_frameworks rf ON a.framework_id = rf.id
    WHERE 
        a.id = p_assessment_id;
    
    -- Obter estatísticas dos resultados
    SELECT 
        COUNT(*) AS total_controls,
        SUM(CASE WHEN status = 'compliant' THEN 1 ELSE 0 END) AS compliant_controls,
        SUM(CASE WHEN status = 'partial_compliance' THEN 1 ELSE 0 END) AS partial_controls,
        SUM(CASE WHEN status = 'non_compliant' THEN 1 ELSE 0 END) AS non_compliant_controls,
        SUM(CASE WHEN status = 'not_applicable' THEN 1 ELSE 0 END) AS not_applicable_controls,
        ROUND(
            (SUM(CASE WHEN status = 'compliant' THEN 1 ELSE 0 END)::FLOAT / 
            NULLIF(COUNT(*) - SUM(CASE WHEN status = 'not_applicable' THEN 1 ELSE 0 END), 0)) * 100, 
            2
        ) AS compliance_percentage
    INTO v_results
    FROM 
        iso27001_control_results
    WHERE 
        assessment_id = p_assessment_id;
    
    -- Obter estatísticas por categoria
    WITH category_stats AS (
        SELECT 
            c.category,
            COUNT(*) AS total,
            SUM(CASE WHEN cr.status = 'compliant' THEN 1 ELSE 0 END) AS compliant,
            SUM(CASE WHEN cr.status = 'partial_compliance' THEN 1 ELSE 0 END) AS partial,
            SUM(CASE WHEN cr.status = 'non_compliant' THEN 1 ELSE 0 END) AS non_compliant,
            ROUND(
                (SUM(CASE WHEN cr.status = 'compliant' THEN 1 ELSE 0 END)::FLOAT / 
                NULLIF(COUNT(*) - SUM(CASE WHEN cr.status = 'not_applicable' THEN 1 ELSE 0 END), 0)) * 100, 
                2
            ) AS percentage
        FROM 
            iso27001_control_results cr
            JOIN iso27001_controls c ON cr.control_id = c.id
        WHERE 
            cr.assessment_id = p_assessment_id
        GROUP BY 
            c.category
    )
    SELECT 
        jsonb_agg(
            jsonb_build_object(
                'category', category,
                'total', total,
                'compliant', compliant,
                'partial', partial,
                'non_compliant', non_compliant,
                'compliance_percentage', percentage
            )
        ) INTO v_category_compliance_stats
    FROM 
        category_stats;
    
    -- Obter estatísticas de planos de ação
    SELECT 
        jsonb_build_object(
            'total', COUNT(*),
            'open', SUM(CASE WHEN status = 'open' THEN 1 ELSE 0 END),
            'in_progress', SUM(CASE WHEN status = 'in_progress' THEN 1 ELSE 0 END),
            'completed', SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END),
            'critical_priority', SUM(CASE WHEN priority = 'critical' THEN 1 ELSE 0 END),
            'high_priority', SUM(CASE WHEN priority = 'high' THEN 1 ELSE 0 END),
            'medium_priority', SUM(CASE WHEN priority = 'medium' THEN 1 ELSE 0 END),
            'low_priority', SUM(CASE WHEN priority = 'low' THEN 1 ELSE 0 END),
            'healthcare_related', SUM(CASE WHEN healthcare_related THEN 1 ELSE 0 END),
            'overdue', SUM(CASE WHEN due_date < CURRENT_DATE AND status IN ('open', 'in_progress') THEN 1 ELSE 0 END)
        ) INTO v_action_plan_stats
    FROM 
        iso27001_action_plans
    WHERE 
        assessment_id = p_assessment_id;
    
    -- Obter as principais recomendações
    WITH top_recommendations AS (
        SELECT 
            c.control_id,
            c.name,
            c.category,
            cr.status,
            cr.score,
            cr.recommendations
        FROM 
            iso27001_control_results cr
            JOIN iso27001_controls c ON cr.control_id = c.id
        WHERE 
            cr.assessment_id = p_assessment_id
            AND cr.status IN ('non_compliant', 'partial_compliance')
            AND jsonb_array_length(cr.recommendations) > 0
        ORDER BY 
            cr.score ASC NULLS FIRST,
            CASE cr.status
                WHEN 'non_compliant' THEN 0
                WHEN 'partial_compliance' THEN 1
                ELSE 2
            END
        LIMIT 10
    )
    SELECT 
        jsonb_agg(
            jsonb_build_object(
                'control_id', control_id,
                'control_name', name,
                'category', category,
                'status', status,
                'recommendations', recommendations
            )
        ) INTO v_recommendations
    FROM 
        top_recommendations;
    
    -- Construir o relatório completo
    v_report := jsonb_build_object(
        'report_id', gen_random_uuid(),
        'report_type', 'ISO 27001 Compliance Report',
        'format', p_format,
        'generated_at', NOW(),
        'assessment', jsonb_build_object(
            'id', v_assessment.id,
            'name', v_assessment.name,
            'organization_id', v_assessment.organization_id,
            'organization_name', v_assessment.organization_name,
            'start_date', v_assessment.start_date,
            'end_date', v_assessment.end_date,
            'status', v_assessment.status,
            'version', v_assessment.version,
            'score', v_assessment.score,
            'healthcare_specific', v_assessment.healthcare_specific,
            'framework_name', v_assessment.framework_name,
            'created_by', v_assessment.created_by_name
        ),
        'compliance_summary', jsonb_build_object(
            'total_controls', v_results.total_controls,
            'compliant_controls', v_results.compliant_controls,
            'partial_compliance_controls', v_results.partial_controls,
            'non_compliant_controls', v_results.non_compliant_controls,
            'not_applicable_controls', v_results.not_applicable_controls,
            'compliance_percentage', v_results.compliance_percentage,
            'compliance_level', CASE 
                WHEN v_results.compliance_percentage >= 90 THEN 'High'
                WHEN v_results.compliance_percentage >= 75 THEN 'Medium'
                WHEN v_results.compliance_percentage >= 50 THEN 'Low'
                ELSE 'Critical'
            END
        ),
        'category_compliance', v_category_compliance_stats,
        'action_plans', v_action_plan_stats,
        'top_recommendations', COALESCE(v_recommendations, '[]'::JSONB)
    );
    
    RETURN v_report;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION generate_iso27001_compliance_report IS 'Gera um relatório detalhado de compliance com ISO 27001 para uma avaliação específica';

-- Função para validar regras de compliance de um controle ISO 27001
CREATE OR REPLACE FUNCTION validate_iso27001_control_rules(
    p_control_id UUID,
    p_validation_data JSONB
) RETURNS TABLE(
    rule_key TEXT,
    expected JSONB,
    actual JSONB,
    is_compliant BOOLEAN,
    message TEXT
) AS $$
DECLARE
    v_validation_rules JSONB;
    v_rule JSONB;
    v_rule_key TEXT;
    v_expected JSONB;
    v_actual JSONB;
    v_is_compliant BOOLEAN;
    v_message TEXT;
BEGIN
    -- Obter as regras de validação do controle
    SELECT validation_rules INTO v_validation_rules
    FROM iso27001_controls
    WHERE id = p_control_id;
    
    -- Se não houver regras, retornar vazio
    IF v_validation_rules IS NULL OR jsonb_array_length(v_validation_rules) = 0 THEN
        RETURN;
    END IF;
    
    -- Validar cada regra
    FOR i IN 0..jsonb_array_length(v_validation_rules)-1 LOOP
        v_rule := jsonb_extract_path(v_validation_rules, i::text);
        v_rule_key := jsonb_extract_path_text(v_rule, 'key');
        v_expected := jsonb_extract_path(v_rule, 'expected');
        
        -- Extrair o valor atual da estrutura de dados fornecida
        v_actual := NULL;
        BEGIN
            v_actual := jsonb_extract_path(p_validation_data, REPLACE(v_rule_key, '.', ','));
        EXCEPTION WHEN OTHERS THEN
            -- Tentar um caminho alternativo
            BEGIN
                v_actual := p_validation_data #> ARRAY[v_rule_key];
            EXCEPTION WHEN OTHERS THEN
                v_actual := NULL;
            END;
        END;
        
        -- Verificar a conformidade
        IF v_actual IS NULL THEN
            v_is_compliant := FALSE;
            v_message := 'Dado para validação não encontrado';
        ELSIF jsonb_typeof(v_expected) = 'boolean' THEN
            -- Validação booleana
            v_is_compliant := (v_expected = v_actual);
            v_message := CASE WHEN v_is_compliant THEN 'Validação bem-sucedida' ELSE 'Valor booleano não corresponde ao esperado' END;
        ELSIF jsonb_typeof(v_expected) = 'array' THEN
            -- Validação de array (contém)
            IF jsonb_typeof(v_actual) = 'array' THEN
                v_is_compliant := TRUE;
                FOR j IN 0..jsonb_array_length(v_expected)-1 LOOP
                    IF NOT v_actual @> jsonb_extract_path(v_expected, j::text) THEN
                        v_is_compliant := FALSE;
                        EXIT;
                    END IF;
                END LOOP;
                v_message := CASE WHEN v_is_compliant THEN 'Validação bem-sucedida' ELSE 'Array não contém todos os valores esperados' END;
            ELSE
                v_is_compliant := FALSE;
                v_message := 'Esperado um array, mas recebido outro tipo de dado';
            END IF;
        ELSIF jsonb_typeof(v_expected) = 'string' AND jsonb_typeof(v_actual) = 'string' THEN
            -- Validação de string
            v_is_compliant := (v_expected = v_actual);
            v_message := CASE WHEN v_is_compliant THEN 'Validação bem-sucedida' ELSE 'String não corresponde ao valor esperado' END;
        ELSIF jsonb_typeof(v_expected) = 'number' AND jsonb_typeof(v_actual) = 'number' THEN
            -- Validação numérica
            v_is_compliant := (v_expected = v_actual);
            v_message := CASE WHEN v_is_compliant THEN 'Validação bem-sucedida' ELSE 'Valor numérico não corresponde ao esperado' END;
        ELSE
            -- Outros tipos
            v_is_compliant := (v_expected = v_actual);
            v_message := CASE WHEN v_is_compliant THEN 'Validação bem-sucedida' ELSE 'Valor não corresponde ao esperado' END;
        END IF;
        
        -- Retornar o resultado da validação
        rule_key := v_rule_key;
        expected := v_expected;
        actual := v_actual;
        is_compliant := v_is_compliant;
        message := v_message;
        
        RETURN NEXT;
    END LOOP;
    
    RETURN;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION validate_iso27001_control_rules IS 'Valida os dados de acordo com as regras definidas para um controle ISO 27001';

-- Função para copiar uma avaliação ISO 27001 para criar uma nova versão
CREATE OR REPLACE FUNCTION copy_iso27001_assessment(
    p_assessment_id UUID,
    p_new_name VARCHAR(255),
    p_created_by UUID
) RETURNS UUID AS $$
DECLARE
    v_original_assessment record;
    v_new_assessment_id UUID;
    v_control_result record;
BEGIN
    -- Obter dados da avaliação original
    SELECT * INTO v_original_assessment
    FROM iso27001_assessments
    WHERE id = p_assessment_id;
    
    -- Criar nova avaliação
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
        v_original_assessment.organization_id,
        p_new_name,
        v_original_assessment.description || ' (Cópia de avaliação anterior: ' || v_original_assessment.name || ')',
        NOW(),
        'in_progress',
        v_original_assessment.scope,
        v_original_assessment.version,
        p_created_by,
        p_created_by,
        v_original_assessment.healthcare_specific,
        v_original_assessment.framework_id
    ) RETURNING id INTO v_new_assessment_id;
    
    -- Copiar resultados de controle
    FOR v_control_result IN
        SELECT * FROM iso27001_control_results
        WHERE assessment_id = p_assessment_id
    LOOP
        INSERT INTO iso27001_control_results (
            assessment_id,
            control_id,
            status,
            implementation_status,
            created_by,
            notes
        ) VALUES (
            v_new_assessment_id,
            v_control_result.control_id,
            'non_compliant', -- Reiniciar o status
            v_control_result.implementation_status, -- Manter o status de implementação
            p_created_by,
            'Baseado em avaliação anterior. Status anterior: ' || v_control_result.status
        );
    END LOOP;
    
    RETURN v_new_assessment_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION copy_iso27001_assessment IS 'Cria uma nova avaliação ISO 27001 baseada em uma avaliação existente';
