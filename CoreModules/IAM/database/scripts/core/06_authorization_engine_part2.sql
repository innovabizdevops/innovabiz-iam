-- INNOVABIZ - IAM Authorization Engine (Parte 2)
-- Author: Eduardo Jeremias
-- Date: 09/05/2025
-- Version: 1.0
-- Description: Funções e procedimentos para motor de autorização híbrido RBAC/ABAC

-- Configuração do esquema
SET search_path TO iam, public;

-- Função para criar uma permissão detalhada
CREATE OR REPLACE FUNCTION iam.create_permission(
    p_organization_id UUID,
    p_code VARCHAR,
    p_name VARCHAR,
    p_description TEXT,
    p_permission_scope iam.permission_scope,
    p_resource_type VARCHAR,
    p_actions VARCHAR[],
    p_metadata JSONB DEFAULT '{}'::JSONB
) RETURNS UUID AS $$
DECLARE
    permission_id UUID;
BEGIN
    INSERT INTO iam.detailed_permissions (
        organization_id,
        code,
        name,
        description,
        permission_scope,
        resource_type,
        actions,
        metadata
    ) VALUES (
        p_organization_id,
        p_code,
        p_name,
        p_description,
        p_permission_scope,
        p_resource_type,
        p_actions,
        p_metadata
    ) RETURNING id INTO permission_id;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        NULL, -- user_id (será preenchido pelo contexto de execução)
        'authorization'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        'CREATE_PERMISSION',
        'detailed_permissions',
        permission_id::TEXT,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        'success',
        NULL, -- response_time
        jsonb_build_object(
            'permission_code', p_code,
            'permission_name', p_name,
            'resource_type', p_resource_type
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['authorization', 'rbac'], -- compliance_tags
        NULL, -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN permission_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para criar um papel (role)
CREATE OR REPLACE FUNCTION iam.create_role(
    p_organization_id UUID,
    p_code VARCHAR,
    p_name VARCHAR,
    p_description TEXT,
    p_is_system_role BOOLEAN DEFAULT FALSE,
    p_parent_role_id UUID DEFAULT NULL,
    p_metadata JSONB DEFAULT '{}'::JSONB
) RETURNS UUID AS $$
DECLARE
    role_id UUID;
BEGIN
    INSERT INTO iam.detailed_roles (
        organization_id,
        code,
        name,
        description,
        is_system_role,
        parent_role_id,
        metadata
    ) VALUES (
        p_organization_id,
        p_code,
        p_name,
        p_description,
        p_is_system_role,
        p_parent_role_id,
        p_metadata
    ) RETURNING id INTO role_id;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        NULL, -- user_id (será preenchido pelo contexto de execução)
        'authorization'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        'CREATE_ROLE',
        'detailed_roles',
        role_id::TEXT,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        'success',
        NULL, -- response_time
        jsonb_build_object(
            'role_code', p_code,
            'role_name', p_name,
            'is_system_role', p_is_system_role,
            'parent_role_id', p_parent_role_id
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['authorization', 'rbac'], -- compliance_tags
        NULL, -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN role_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para atribuir uma permissão a um papel
CREATE OR REPLACE FUNCTION iam.assign_permission_to_role(
    p_organization_id UUID,
    p_role_id UUID,
    p_permission_id UUID
) RETURNS UUID AS $$
DECLARE
    assignment_id UUID;
BEGIN
    INSERT INTO iam.role_permissions (
        role_id,
        permission_id,
        organization_id
    ) VALUES (
        p_role_id,
        p_permission_id,
        p_organization_id
    ) RETURNING id INTO assignment_id;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        NULL, -- user_id (será preenchido pelo contexto de execução)
        'authorization'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        'ASSIGN_PERMISSION_TO_ROLE',
        'role_permissions',
        assignment_id::TEXT,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        'success',
        NULL, -- response_time
        jsonb_build_object(
            'role_id', p_role_id,
            'permission_id', p_permission_id
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['authorization', 'rbac'], -- compliance_tags
        NULL, -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN assignment_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para atribuir um papel a um usuário
CREATE OR REPLACE FUNCTION iam.assign_role_to_user(
    p_organization_id UUID,
    p_user_id UUID,
    p_role_id UUID,
    p_created_by UUID DEFAULT NULL,
    p_expires_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    p_metadata JSONB DEFAULT '{}'::JSONB
) RETURNS UUID AS $$
DECLARE
    assignment_id UUID;
BEGIN
    INSERT INTO iam.user_roles (
        user_id,
        role_id,
        organization_id,
        expires_at,
        created_by,
        metadata
    ) VALUES (
        p_user_id,
        p_role_id,
        p_organization_id,
        p_expires_at,
        p_created_by,
        p_metadata
    ) RETURNING id INTO assignment_id;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        COALESCE(p_created_by, p_user_id), -- user_id
        'authorization'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        'ASSIGN_ROLE_TO_USER',
        'user_roles',
        assignment_id::TEXT,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        'success',
        NULL, -- response_time
        jsonb_build_object(
            'user_id', p_user_id,
            'role_id', p_role_id,
            'expires_at', p_expires_at
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['authorization', 'rbac'], -- compliance_tags
        NULL, -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN assignment_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para criar uma política ABAC
CREATE OR REPLACE FUNCTION iam.create_attribute_policy(
    p_organization_id UUID,
    p_name VARCHAR,
    p_description TEXT,
    p_effect iam.permission_effect,
    p_priority INTEGER,
    p_resource_type VARCHAR,
    p_resource_pattern VARCHAR,
    p_action_pattern VARCHAR,
    p_condition_expression JSONB,
    p_condition_attributes JSONB,
    p_metadata JSONB DEFAULT '{}'::JSONB
) RETURNS UUID AS $$
DECLARE
    policy_id UUID;
BEGIN
    INSERT INTO iam.attribute_policies (
        organization_id,
        name,
        description,
        effect,
        priority,
        resource_type,
        resource_pattern,
        action_pattern,
        condition_expression,
        condition_attributes,
        metadata
    ) VALUES (
        p_organization_id,
        p_name,
        p_description,
        p_effect,
        p_priority,
        p_resource_type,
        p_resource_pattern,
        p_action_pattern,
        p_condition_expression,
        p_condition_attributes,
        p_metadata
    ) RETURNING id INTO policy_id;
    
    -- Registrar auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        NULL, -- user_id (será preenchido pelo contexto de execução)
        'authorization'::iam.audit_event_category,
        'info'::iam.audit_severity_level,
        'CREATE_ATTRIBUTE_POLICY',
        'attribute_policies',
        policy_id::TEXT,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        'success',
        NULL, -- response_time
        jsonb_build_object(
            'policy_name', p_name,
            'effect', p_effect,
            'resource_type', p_resource_type
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['authorization', 'abac'], -- compliance_tags
        NULL, -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN policy_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para avaliar autorização baseada em RBAC
CREATE OR REPLACE FUNCTION iam.evaluate_rbac_authorization(
    p_user_id UUID,
    p_organization_id UUID,
    p_resource_type VARCHAR,
    p_action VARCHAR
) RETURNS BOOLEAN AS $$
DECLARE
    is_authorized BOOLEAN := FALSE;
BEGIN
    -- Verificar se o usuário tem algum papel com a permissão necessária
    SELECT EXISTS (
        SELECT 1
        FROM iam.user_roles ur
        JOIN iam.role_permissions rp ON ur.role_id = rp.role_id
        JOIN iam.detailed_permissions dp ON rp.permission_id = dp.id
        WHERE ur.user_id = p_user_id
          AND ur.organization_id = p_organization_id
          AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
          AND dp.resource_type = p_resource_type
          AND p_action = ANY(dp.actions)
    ) INTO is_authorized;
    
    -- Registrar decisão em auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        p_user_id,
        'authorization'::iam.audit_event_category,
        CASE 
            WHEN is_authorized THEN 'info'::iam.audit_severity_level
            ELSE 'medium'::iam.audit_severity_level
        END,
        'EVALUATE_RBAC_AUTHORIZATION',
        p_resource_type,
        NULL, -- resource_id
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        CASE 
            WHEN is_authorized THEN 'success'
            ELSE 'denied'
        END,
        NULL, -- response_time
        jsonb_build_object(
            'user_id', p_user_id,
            'resource_type', p_resource_type,
            'action', p_action,
            'authorized', is_authorized
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['authorization', 'rbac'], -- compliance_tags
        NULL, -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN is_authorized;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para avaliar expressão ABAC
CREATE OR REPLACE FUNCTION iam.evaluate_abac_expression(
    p_expression JSONB,
    p_context JSONB
) RETURNS BOOLEAN AS $$
DECLARE
    expression_type TEXT;
    result BOOLEAN;
BEGIN
    -- Obter tipo de expressão
    expression_type := p_expression->>'type';
    
    -- Avaliar com base no tipo
    IF expression_type = 'and' THEN
        -- Avaliar expressão AND (todas as subexpressões devem ser verdadeiras)
        result := TRUE;
        FOR i IN 0..jsonb_array_length(p_expression->'expressions') - 1 LOOP
            IF NOT iam.evaluate_abac_expression(
                jsonb_array_element(p_expression->'expressions', i),
                p_context
            ) THEN
                result := FALSE;
                EXIT;
            END IF;
        END LOOP;
    ELSIF expression_type = 'or' THEN
        -- Avaliar expressão OR (pelo menos uma subexpressão deve ser verdadeira)
        result := FALSE;
        FOR i IN 0..jsonb_array_length(p_expression->'expressions') - 1 LOOP
            IF iam.evaluate_abac_expression(
                jsonb_array_element(p_expression->'expressions', i),
                p_context
            ) THEN
                result := TRUE;
                EXIT;
            END IF;
        END LOOP;
    ELSIF expression_type = 'not' THEN
        -- Avaliar expressão NOT (negação da subexpressão)
        result := NOT iam.evaluate_abac_expression(
            p_expression->'expression',
            p_context
        );
    ELSIF expression_type = 'comparison' THEN
        -- Avaliar comparação
        DECLARE
            operator TEXT;
            left_value JSONB;
            right_value JSONB;
            left_result TEXT;
            right_result TEXT;
        BEGIN
            operator := p_expression->>'operator';
            left_value := p_expression->'left';
            right_value := p_expression->'right';
            
            -- Resolver valores de contexto
            IF left_value->>'type' = 'attribute' THEN
                left_result := p_context->>(left_value->>'name');
            ELSE
                left_result := left_value->>'value';
            END IF;
            
            IF right_value->>'type' = 'attribute' THEN
                right_result := p_context->>(right_value->>'name');
            ELSE
                right_result := right_value->>'value';
            END IF;
            
            -- Realizar comparação
            IF operator = 'eq' THEN
                result := left_result = right_result;
            ELSIF operator = 'ne' THEN
                result := left_result <> right_result;
            ELSIF operator = 'gt' THEN
                result := left_result::numeric > right_result::numeric;
            ELSIF operator = 'ge' THEN
                result := left_result::numeric >= right_result::numeric;
            ELSIF operator = 'lt' THEN
                result := left_result::numeric < right_result::numeric;
            ELSIF operator = 'le' THEN
                result := left_result::numeric <= right_result::numeric;
            ELSIF operator = 'contains' THEN
                result := left_result LIKE '%' || right_result || '%';
            ELSIF operator = 'startswith' THEN
                result := left_result LIKE right_result || '%';
            ELSIF operator = 'endswith' THEN
                result := left_result LIKE '%' || right_result;
            ELSIF operator = 'in' THEN
                -- 'in' operator expects right side to be a comma-separated list
                result := left_result = ANY(string_to_array(right_result, ','));
            ELSE
                -- Operador desconhecido
                result := FALSE;
            END IF;
        END;
    ELSE
        -- Tipo de expressão desconhecido
        result := FALSE;
    END IF;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para avaliar políticas ABAC para um recurso e ação
CREATE OR REPLACE FUNCTION iam.evaluate_abac_policies(
    p_user_id UUID,
    p_organization_id UUID,
    p_resource_type VARCHAR,
    p_resource_id VARCHAR,
    p_action VARCHAR,
    p_context JSONB DEFAULT '{}'::JSONB
) RETURNS iam.permission_effect AS $$
DECLARE
    final_decision iam.permission_effect := 'deny';
    policy RECORD;
    policy_decision BOOLEAN;
    policy_context JSONB;
BEGIN
    -- Preparar contexto para avaliação
    policy_context := jsonb_build_object(
        'user_id', p_user_id,
        'organization_id', p_organization_id,
        'resource_type', p_resource_type,
        'resource_id', p_resource_id,
        'action', p_action,
        'current_time', NOW()
    );
    
    -- Adicionar contexto personalizado
    policy_context := policy_context || p_context;
    
    -- Adicionar atributos do usuário
    SELECT jsonb_build_object(
        'username', u.username,
        'email', u.email,
        'status', u.status,
        'preferences', u.preferences,
        'metadata', u.metadata
    ) INTO policy_context
    FROM iam.users u
    WHERE u.id = p_user_id;
    
    policy_context := policy_context || jsonb_build_object('user', policy_context);
    
    -- Obter políticas aplicáveis em ordem de prioridade
    FOR policy IN (
        SELECT 
            ap.id,
            ap.name,
            ap.effect,
            ap.condition_expression
        FROM iam.attribute_policies ap
        LEFT JOIN iam.policy_set_policies psp ON ap.id = psp.attribute_policy_id
        LEFT JOIN iam.policy_sets ps ON psp.policy_set_id = ps.id
        LEFT JOIN iam.role_policy_sets rps ON ps.id = rps.policy_set_id
        LEFT JOIN iam.user_roles ur ON rps.role_id = ur.role_id
        WHERE ap.organization_id = p_organization_id
          AND ap.is_active = TRUE
          AND (
                -- Match resource type
                ap.resource_type = p_resource_type
                OR ap.resource_type = '*'
              )
          AND (
                -- Match action pattern
                p_action LIKE ap.action_pattern
                OR ap.action_pattern IS NULL
                OR ap.action_pattern = '*'
              )
          AND (
                -- Either direct policy or through role policy set
                ur.user_id = p_user_id
                OR ap.id IN (
                    SELECT cap.id
                    FROM iam.attribute_policies cap
                    WHERE cap.organization_id = p_organization_id
                      AND (
                            cap.resource_type = p_resource_type
                            OR cap.resource_type = '*'
                          )
                )
              )
        ORDER BY 
            ap.priority DESC,
            CASE ap.effect WHEN 'deny' THEN 0 ELSE 1 END -- Deny has precedence over allow
    ) LOOP
        -- Avaliar a expressão da política
        policy_decision := iam.evaluate_abac_expression(
            policy.condition_expression,
            policy_context
        );
        
        -- Se a expressão for verdadeira, aplicar a decisão da política
        IF policy_decision THEN
            final_decision := policy.effect;
            
            -- Registrar decisão em auditoria
            PERFORM iam.log_audit_event(
                p_organization_id,
                p_user_id,
                'authorization'::iam.audit_event_category,
                CASE 
                    WHEN final_decision = 'allow' THEN 'info'::iam.audit_severity_level
                    ELSE 'medium'::iam.audit_severity_level
                END,
                'EVALUATE_ABAC_POLICY',
                p_resource_type,
                p_resource_id,
                NULL, -- source_ip
                NULL, -- user_agent
                NULL, -- request_id
                NULL, -- session_id
                CASE 
                    WHEN final_decision = 'allow' THEN 'success'
                    ELSE 'denied'
                END,
                NULL, -- response_time
                jsonb_build_object(
                    'policy_id', policy.id,
                    'policy_name', policy.name,
                    'effect', policy.effect,
                    'resource_type', p_resource_type,
                    'resource_id', p_resource_id,
                    'action', p_action
                ),
                NULL, -- request_payload
                NULL, -- response_payload
                ARRAY['authorization', 'abac'], -- compliance_tags
                NULL, -- regulatory_references
                NULL  -- geo_location
            );
            
            -- Deny overrides - se encontramos deny, já podemos retornar
            IF final_decision = 'deny' THEN
                RETURN final_decision;
            END IF;
            
            -- Se encontramos um allow, continuamos procurando pois deny tem precedência
        END IF;
    END LOOP;
    
    RETURN final_decision;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função principal de autorização combinando RBAC e ABAC
CREATE OR REPLACE FUNCTION iam.is_authorized(
    p_user_id UUID,
    p_organization_id UUID,
    p_resource_type VARCHAR,
    p_resource_id VARCHAR DEFAULT NULL,
    p_action VARCHAR,
    p_context JSONB DEFAULT '{}'::JSONB,
    p_use_cache BOOLEAN DEFAULT TRUE
) RETURNS BOOLEAN AS $$
DECLARE
    rbac_result BOOLEAN;
    abac_result iam.permission_effect;
    final_decision BOOLEAN;
    cached_decision RECORD;
    cache_key TEXT;
BEGIN
    -- Verificar cache, se habilitado
    IF p_use_cache THEN
        SELECT decision
        INTO cached_decision
        FROM iam.authorization_decisions_cache
        WHERE user_id = p_user_id
          AND organization_id = p_organization_id
          AND resource_type = p_resource_type
          AND COALESCE(resource_id, '') = COALESCE(p_resource_id, '')
          AND action = p_action
          AND expires_at > NOW()
        LIMIT 1;
        
        IF FOUND THEN
            RETURN cached_decision.decision = 'allow';
        END IF;
    END IF;
    
    -- Primeiro, avaliar RBAC para permissões básicas
    rbac_result := iam.evaluate_rbac_authorization(
        p_user_id,
        p_organization_id,
        p_resource_type,
        p_action
    );
    
    -- Se RBAC já autoriza, ainda precisamos verificar ABAC para verificar restrições
    IF rbac_result THEN
        -- Avaliar políticas ABAC
        abac_result := iam.evaluate_abac_policies(
            p_user_id,
            p_organization_id,
            p_resource_type,
            p_resource_id,
            p_action,
            p_context
        );
        
        -- ABAC pode negar mesmo que RBAC autorize
        final_decision := abac_result = 'allow';
    ELSE
        -- Se RBAC já negou, a decisão final é negar
        final_decision := FALSE;
    END IF;
    
    -- Armazenar em cache, se habilitado
    IF p_use_cache THEN
        INSERT INTO iam.authorization_decisions_cache (
            user_id,
            organization_id,
            resource_type,
            resource_id,
            action,
            decision,
            decision_context,
            expires_at
        ) VALUES (
            p_user_id,
            p_organization_id,
            p_resource_type,
            p_resource_id,
            p_action,
            CASE WHEN final_decision THEN 'allow'::iam.permission_effect ELSE 'deny'::iam.permission_effect END,
            p_context,
            NOW() + interval '15 minutes'
        );
    END IF;
    
    -- Registrar decisão final em auditoria
    PERFORM iam.log_audit_event(
        p_organization_id,
        p_user_id,
        'authorization'::iam.audit_event_category,
        CASE 
            WHEN final_decision THEN 'info'::iam.audit_severity_level
            ELSE 'medium'::iam.audit_severity_level
        END,
        'AUTHORIZATION_DECISION',
        p_resource_type,
        p_resource_id,
        NULL, -- source_ip
        NULL, -- user_agent
        NULL, -- request_id
        NULL, -- session_id
        CASE 
            WHEN final_decision THEN 'success'
            ELSE 'denied'
        END,
        NULL, -- response_time
        jsonb_build_object(
            'resource_type', p_resource_type,
            'resource_id', p_resource_id,
            'action', p_action,
            'rbac_result', rbac_result,
            'abac_result', abac_result,
            'final_decision', final_decision
        ),
        NULL, -- request_payload
        NULL, -- response_payload
        ARRAY['authorization', 'rbac', 'abac'], -- compliance_tags
        NULL, -- regulatory_references
        NULL  -- geo_location
    );
    
    RETURN final_decision;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
