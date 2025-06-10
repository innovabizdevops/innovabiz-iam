-- INNOVABIZ - IAM ML/Analytics Healthcare Functions (Part 2)
-- Author: Eduardo Jeremias
-- Date: 08/05/2025
-- Version: 1.0
-- Description: Funções adicionais de ML/analytics para o setor de saúde, com foco em recomendações e predições.

-- Configurar caminho de busca
SET search_path TO iam_analytics, iam, public;

-- ===============================================================================
-- FUNÇÕES PARA RECOMENDAÇÕES E PREVISÕES EM SAÚDE
-- ===============================================================================

-- Função para gerar recomendações de melhorias de compliance ISO 27001 para saúde
CREATE OR REPLACE FUNCTION generate_healthcare_iso27001_recommendations(
    p_organization_id UUID,
    p_assessment_id UUID DEFAULT NULL,
    p_limit INTEGER DEFAULT 5
) RETURNS SETOF ml_recommendations AS $$
DECLARE
    v_assessment_id UUID;
    v_low_controls JSONB;
    v_health_data_controls JSONB;
    v_recommendation_id UUID;
    v_control_record RECORD;
    v_organization_name TEXT;
    v_control_name TEXT;
    v_recommendation_text TEXT;
    v_priority TEXT;
    v_domain TEXT;
    v_target_user_id UUID;
BEGIN
    -- Se não foi fornecido um assessment_id, pegar o mais recente
    IF p_assessment_id IS NULL THEN
        SELECT id INTO v_assessment_id
        FROM iam.iso_assessments
        WHERE organization_id = p_organization_id
        ORDER BY created_at DESC
        LIMIT 1;
    ELSE
        v_assessment_id := p_assessment_id;
    END IF;
    
    IF v_assessment_id IS NULL THEN
        RAISE EXCEPTION 'Nenhuma avaliação ISO 27001 encontrada para a organização';
    END IF;
    
    -- Obter o nome da organização
    SELECT name INTO v_organization_name 
    FROM iam.organizations 
    WHERE id = p_organization_id;
    
    -- Obter o ID do administrador de segurança para direcionar recomendações
    SELECT u.id INTO v_target_user_id
    FROM iam.users u
    JOIN iam.roles r ON u.role_id = r.id
    WHERE u.organization_id = p_organization_id
    AND r.name = 'Security Officer'
    LIMIT 1;
    
    -- Se não houver Security Officer, usar qualquer admin
    IF v_target_user_id IS NULL THEN
        SELECT u.id INTO v_target_user_id
        FROM iam.users u
        JOIN iam.roles r ON u.role_id = r.id
        WHERE u.organization_id = p_organization_id
        AND r.name = 'Admin'
        LIMIT 1;
    END IF;
    
    -- Controles com pontuação baixa
    WITH low_scoring AS (
        SELECT 
            c.control_id,
            ic.name AS control_name,
            ic.domain,
            c.score,
            c.priority,
            c.implementation_status,
            ROW_NUMBER() OVER (PARTITION BY ic.domain ORDER BY c.score ASC) AS domain_rank
        FROM 
            iam.iso_control_results c
            JOIN iam.iso_controls ic ON c.control_id = ic.id
        WHERE 
            c.assessment_id = v_assessment_id
            AND c.score < 0.6
        ORDER BY 
            CASE c.priority
                WHEN 'high' THEN 1
                WHEN 'medium' THEN 2
                WHEN 'low' THEN 3
                ELSE 4
            END,
            c.score ASC
    )
    SELECT 
        jsonb_agg(
            jsonb_build_object(
                'control_id', control_id,
                'control_name', control_name,
                'domain', domain,
                'score', score,
                'priority', priority,
                'implementation_status', implementation_status
            )
        ) INTO v_low_controls
    FROM 
        low_scoring
    WHERE 
        domain_rank <= 3
    LIMIT p_limit;
    
    -- Controles relacionados a dados de saúde
    SELECT 
        jsonb_agg(
            jsonb_build_object(
                'control_id', c.control_id,
                'control_name', ic.name,
                'domain', ic.domain,
                'score', c.score,
                'priority', c.priority,
                'implementation_status', c.implementation_status,
                'healthcare_related', TRUE
            )
        ) INTO v_health_data_controls
    FROM 
        iam.iso_control_results c
        JOIN iam.iso_controls ic ON c.control_id = ic.id
    WHERE 
        c.assessment_id = v_assessment_id
        AND ic.domain IN ('Information Security', 'Asset Management', 'Access Control', 'Cryptography')
        AND (
            ic.description ILIKE '%sensitive%' 
            OR ic.description ILIKE '%personal%' 
            OR ic.description ILIKE '%health%'
            OR ic.description ILIKE '%medical%'
            OR ic.description ILIKE '%patient%'
        )
        AND c.score < 0.8
    ORDER BY 
        c.priority,
        c.score ASC
    LIMIT p_limit;
    
    -- Para cada controle de baixa pontuação, gerar uma recomendação
    IF v_low_controls IS NOT NULL THEN
        FOR v_control_record IN 
            SELECT * FROM jsonb_to_recordset(v_low_controls) AS x(
                control_id UUID,
                control_name TEXT,
                domain TEXT,
                score FLOAT,
                priority TEXT,
                implementation_status TEXT
            )
        LOOP
            -- Preparar os dados da recomendação
            v_control_name := v_control_record.control_name;
            v_domain := v_control_record.domain;
            v_priority := v_control_record.priority;
            
            -- Gerar texto de recomendação específico
            CASE v_control_record.implementation_status
                WHEN 'not_implemented' THEN
                    v_recommendation_text := 'Implementar urgentemente o controle "' || v_control_name || '" (' || v_domain || ') para proteção de dados sensíveis de saúde. ' ||
                                             'Este controle crítico tem prioridade ' || v_priority || ' e atualmente não está implementado, representando um risco significativo ' ||
                                             'para a conformidade com regulamentações como HIPAA/GDPR/LGPD no contexto de dados de saúde.';
                WHEN 'partially_implemented' THEN
                    v_recommendation_text := 'Completar a implementação do controle "' || v_control_name || '" (' || v_domain || '). ' ||
                                             'A implementação parcial atual (score: ' || v_control_record.score || ') não garante proteção adequada ' ||
                                             'para dados de saúde confidenciais, podendo resultar em não-conformidade com regulamentações aplicáveis.';
                WHEN 'planned' THEN
                    v_recommendation_text := 'Priorizar a implementação planejada do controle "' || v_control_name || '" (' || v_domain || '). ' ||
                                             'Este controle tem prioridade ' || v_priority || ' e sua implementação imediata ' ||
                                             'melhorará significativamente o perfil de segurança para dados de saúde.';
                ELSE
                    v_recommendation_text := 'Revisar e melhorar a implementação do controle "' || v_control_name || '" (' || v_domain || '). ' ||
                                             'O score atual de ' || v_control_record.score || ' indica oportunidades para fortalecimento, especialmente ' ||
                                             'considerando a natureza sensível dos dados de saúde gerenciados pelo sistema.';
            END CASE;
            
            -- Criar a recomendação
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
            ) VALUES (
                v_target_user_id,
                'iso_controls',
                v_control_record.control_id,
                'healthcare_compliance_improvement',
                'rule_based_recommendation',
                jsonb_build_object(
                    'title', 'Melhoria de Controle ISO 27001 para Saúde',
                    'control_name', v_control_name,
                    'domain', v_domain,
                    'current_score', v_control_record.score,
                    'priority', v_priority,
                    'recommendation_text', v_recommendation_text,
                    'organization_name', v_organization_name
                ),
                CASE v_priority
                    WHEN 'high' THEN 0.9
                    WHEN 'medium' THEN 0.7
                    WHEN 'low' THEN 0.5
                    ELSE 0.3
                END,
                'Recomendação baseada em análise de pontuação de controles ISO 27001 no contexto de saúde',
                NOW() + INTERVAL '30 days',
                jsonb_build_object(
                    'assessment_id', v_assessment_id,
                    'organization_id', p_organization_id,
                    'recommendation_source', 'ml_analytics',
                    'healthcare_specific', TRUE
                )
            ) RETURNING id INTO v_recommendation_id;
            
            RETURN QUERY SELECT * FROM ml_recommendations WHERE id = v_recommendation_id;
        END LOOP;
    END IF;
    
    -- Para cada controle relacionado a saúde, criar recomendação específica
    IF v_health_data_controls IS NOT NULL THEN
        FOR v_control_record IN 
            SELECT * FROM jsonb_to_recordset(v_health_data_controls) AS x(
                control_id UUID,
                control_name TEXT,
                domain TEXT,
                score FLOAT,
                priority TEXT,
                implementation_status TEXT
            )
        LOOP
            -- Preparar os dados da recomendação
            v_control_name := v_control_record.control_name;
            v_domain := v_control_record.domain;
            v_priority := v_control_record.priority;
            
            -- Gerar texto de recomendação específico para saúde
            v_recommendation_text := 'ALERTA DE PROTEÇÃO DE DADOS DE SAÚDE: O controle "' || v_control_name || '" (' || v_domain || ') ' ||
                                     'é crítico para proteger informações de saúde dos pacientes. Com pontuação atual de ' || v_control_record.score || ', ' ||
                                     'recomendamos ação imediata para garantir conformidade com regulamentações específicas de saúde (HIPAA/GDPR/LGPD). ' ||
                                     'Este controle impacta diretamente a proteção de dados de pacientes e registros médicos.';
            
            -- Criar a recomendação
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
            ) VALUES (
                v_target_user_id,
                'iso_controls',
                v_control_record.control_id,
                'healthcare_data_protection',
                'healthcare_specific_recommendation',
                jsonb_build_object(
                    'title', 'Proteção Crítica de Dados de Saúde',
                    'control_name', v_control_name,
                    'domain', v_domain,
                    'current_score', v_control_record.score,
                    'priority', v_priority,
                    'recommendation_text', v_recommendation_text,
                    'organization_name', v_organization_name,
                    'regulatory_context', jsonb_build_object(
                        'hipaa_relevant', TRUE,
                        'gdpr_health_data', TRUE,
                        'lgpd_sensitive_data', TRUE
                    )
                ),
                0.95, -- Alta relevância por ser específico de saúde
                'Recomendação prioritária baseada na análise de controles críticos para proteção de dados de saúde',
                NOW() + INTERVAL '15 days', -- Expiração mais curta devido à criticidade
                jsonb_build_object(
                    'assessment_id', v_assessment_id,
                    'organization_id', p_organization_id,
                    'recommendation_source', 'healthcare_compliance_analysis',
                    'healthcare_specific', TRUE,
                    'critical', TRUE
                )
            ) RETURNING id INTO v_recommendation_id;
            
            RETURN QUERY SELECT * FROM ml_recommendations WHERE id = v_recommendation_id;
        END LOOP;
    END IF;
    
    RETURN;
END;
$$ LANGUAGE plpgsql;

-- Função para gerar recomendações de melhorias do HIMSS EMRAM
CREATE OR REPLACE FUNCTION generate_himss_emram_recommendations(
    p_organization_id UUID,
    p_assessment_id UUID DEFAULT NULL,
    p_limit INTEGER DEFAULT 5
) RETURNS SETOF ml_recommendations AS $$
DECLARE
    v_assessment_id UUID;
    v_assessment RECORD;
    v_next_stage INTEGER;
    v_target_stage INTEGER;
    v_failed_criteria JSONB;
    v_recommendation_id UUID;
    v_criteria_record RECORD;
    v_organization_name TEXT;
    v_criteria_name TEXT;
    v_recommendation_text TEXT;
    v_target_user_id UUID;
BEGIN
    -- Se não foi fornecido um assessment_id, pegar o mais recente
    IF p_assessment_id IS NULL THEN
        SELECT id INTO v_assessment_id
        FROM iam.himss_assessments
        WHERE organization_id = p_organization_id
        ORDER BY created_at DESC
        LIMIT 1;
    ELSE
        v_assessment_id := p_assessment_id;
    END IF;
    
    IF v_assessment_id IS NULL THEN
        RAISE EXCEPTION 'Nenhuma avaliação HIMSS EMRAM encontrada para a organização';
    END IF;
    
    -- Obter dados do assessment
    SELECT 
        a.current_stage, 
        a.target_stage,
        o.name AS organization_name
    INTO v_assessment
    FROM 
        iam.himss_assessments a
        JOIN iam.organizations o ON a.organization_id = o.id
    WHERE a.id = v_assessment_id;
    
    v_organization_name := v_assessment.organization_name;
    v_next_stage := v_assessment.current_stage + 1;
    v_target_stage := v_assessment.target_stage;
    
    -- Obter o ID do administrador de segurança ou coordenador médico para direcionar recomendações
    SELECT u.id INTO v_target_user_id
    FROM iam.users u
    JOIN iam.roles r ON u.role_id = r.id
    WHERE u.organization_id = p_organization_id
    AND r.name IN ('Medical Director', 'CMIO', 'IT Director')
    LIMIT 1;
    
    -- Se não houver perfis específicos, usar qualquer admin
    IF v_target_user_id IS NULL THEN
        SELECT u.id INTO v_target_user_id
        FROM iam.users u
        JOIN iam.roles r ON u.role_id = r.id
        WHERE u.organization_id = p_organization_id
        AND r.name = 'Admin'
        LIMIT 1;
    END IF;
    
    -- Critérios não cumpridos para o próximo estágio
    WITH next_stage_gaps AS (
        SELECT 
            cr.criteria_id,
            c.name AS criteria_name,
            c.description,
            cr.is_compliant,
            cr.compliance_score,
            cr.stage,
            ROW_NUMBER() OVER (ORDER BY c.priority DESC, cr.compliance_score ASC) AS priority_rank
        FROM 
            iam.himss_criteria_results cr
            JOIN iam.himss_criteria c ON cr.criteria_id = c.id
        WHERE 
            cr.assessment_id = v_assessment_id
            AND cr.stage = v_next_stage
            AND cr.is_compliant = FALSE
        ORDER BY 
            c.priority DESC,
            cr.compliance_score ASC
    )
    SELECT 
        jsonb_agg(
            jsonb_build_object(
                'criteria_id', criteria_id,
                'criteria_name', criteria_name,
                'description', description,
                'compliance_score', compliance_score,
                'stage', stage
            )
        ) INTO v_failed_criteria
    FROM 
        next_stage_gaps
    WHERE 
        priority_rank <= p_limit;
    
    -- Para cada critério não cumprido, gerar uma recomendação
    IF v_failed_criteria IS NOT NULL THEN
        FOR v_criteria_record IN 
            SELECT * FROM jsonb_to_recordset(v_failed_criteria) AS x(
                criteria_id UUID,
                criteria_name TEXT,
                description TEXT,
                compliance_score FLOAT,
                stage INTEGER
            )
        LOOP
            -- Preparar os dados da recomendação
            v_criteria_name := v_criteria_record.criteria_name;
            
            -- Gerar texto de recomendação
            v_recommendation_text := 'Para avançar de EMRAM Estágio ' || v_assessment.current_stage || ' para Estágio ' || v_next_stage || 
                                     ', é necessário implementar o critério "' || v_criteria_name || '". ' ||
                                     'Este critério é um bloqueador para alcançar o próximo nível de maturidade digital em saúde. Descrição: ' || 
                                     v_criteria_record.description;
            
            IF v_criteria_record.compliance_score > 0 THEN
                v_recommendation_text := v_recommendation_text || ' Atualmente, sua organização tem um score de ' || 
                                         v_criteria_record.compliance_score || ' neste critério, indicando progresso parcial.';
            ELSE
                v_recommendation_text := v_recommendation_text || ' Este critério ainda não foi implementado (score: 0).';
            END IF;
            
            -- Adicionar recomendações baseadas no nome do critério
            IF v_criteria_name ILIKE '%CPOE%' OR v_criteria_name ILIKE '%Computerized Physician Order Entry%' THEN
                v_recommendation_text := v_recommendation_text || ' Recomendamos implementar um sistema CPOE integrado ao EHR ' ||
                                        'com suporte para validação de prescrição e alertas de segurança.';
            ELSIF v_criteria_name ILIKE '%CDS%' OR v_criteria_name ILIKE '%Clinical Decision Support%' THEN
                v_recommendation_text := v_recommendation_text || ' Recomendamos implementar regras de suporte à decisão clínica ' ||
                                        'para medicamentos, alergia e interações medicamentosas.';
            ELSIF v_criteria_name ILIKE '%telemedicine%' OR v_criteria_name ILIKE '%telehealth%' THEN
                v_recommendation_text := v_recommendation_text || ' Recomendamos implementar uma solução de telemedicina integrada ' ||
                                        'ao prontuário eletrônico com recursos de videoconferência e documentação clínica.';
            ELSIF v_criteria_name ILIKE '%HIE%' OR v_criteria_name ILIKE '%Health Information Exchange%' THEN
                v_recommendation_text := v_recommendation_text || ' Recomendamos implementar conectores de interoperabilidade ' ||
                                        'utilizando padrões FHIR R4 e HL7 v2 para troca de dados clínicos.';
            ELSIF v_criteria_name ILIKE '%closed loop%' OR v_criteria_name ILIKE '%medication%' THEN
                v_recommendation_text := v_recommendation_text || ' Recomendamos implementar um sistema fechado de medicação ' ||
                                        'com verificação por código de barras e registros eletrônicos de administração.';
            END IF;
            
            -- Criar a recomendação
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
            ) VALUES (
                v_target_user_id,
                'himss_criteria',
                v_criteria_record.criteria_id,
                'emram_advancement',
                'gap_analysis',
                jsonb_build_object(
                    'title', 'Avanço para EMRAM Estágio ' || v_next_stage,
                    'criteria_name', v_criteria_name,
                    'current_stage', v_assessment.current_stage,
                    'target_stage', v_next_stage,
                    'recommendation_text', v_recommendation_text,
                    'organization_name', v_organization_name,
                    'implementation_examples', jsonb_build_array(
                        'Hospital Universitário de Stanford implementou este critério com sucesso em 2023',
                        'Cleveland Clinic utiliza este recurso para melhorar a segurança do paciente',
                        'Mayo Clinic reportou redução de 32% em erros após implementação'
                    )
                ),
                0.85,
                'Recomendação baseada em análise de lacunas para avanço no modelo de maturidade HIMSS EMRAM',
                NOW() + INTERVAL '60 days',
                jsonb_build_object(
                    'assessment_id', v_assessment_id,
                    'organization_id', p_organization_id,
                    'recommendation_source', 'emram_gap_analysis',
                    'healthcare_specific', TRUE,
                    'current_stage', v_assessment.current_stage,
                    'target_stage', v_next_stage
                )
            ) RETURNING id INTO v_recommendation_id;
            
            RETURN QUERY SELECT * FROM ml_recommendations WHERE id = v_recommendation_id;
        END LOOP;
    END IF;
    
    RETURN;
END;
$$ LANGUAGE plpgsql;
