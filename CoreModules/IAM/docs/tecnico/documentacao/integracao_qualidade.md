# Integração IAM Compliance Validators com Sistema de Gestão da Qualidade

## Visão Geral

Este documento descreve a integração entre os validadores de conformidade do IAM (Identity and Access Management) e o Sistema de Gestão da Qualidade da plataforma INNOVABIZ. A integração permite a identificação automática de problemas de qualidade relacionados à conformidade IAM, a geração de ações corretivas, e o monitoramento de processos de melhoria contínua para manter a conformidade com os padrões da indústria.

## Arquitetura da Integração

A integração foi projetada como um módulo que se conecta aos validadores de conformidade existentes e ao Sistema de Gestão da Qualidade, permitindo:

1. **Mapeamento bidirécional**: Associação entre requisitos de conformidade IAM e padrões de qualidade
2. **Automação de não-conformidades**: Conversão automática de falhas de validação em não-conformidades de qualidade
3. **Geração de ações corretivas**: Criação automática de ações corretivas com base em templates predefinidos
4. **Rastreabilidade completa**: Rastreamento da origem da não-conformidade até a resolução
5. **Métricas de qualidade**: Atualização automática de KPIs e métricas de qualidade baseadas em resultados de validação

## Componentes Principais

### 1. Tabelas e Schemas

```sql
-- Schema principal
CREATE SCHEMA IF NOT EXISTS quality_management;

-- Mapeamento entre padrões de qualidade e validadores de conformidade
CREATE TABLE quality_management.standard_validator_mapping (
    mapping_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    standard_id VARCHAR(50) NOT NULL,
    standard_code VARCHAR(50) NOT NULL,
    standard_name VARCHAR(255) NOT NULL,
    validator_id VARCHAR(50) NOT NULL,
    validator_name VARCHAR(255) NOT NULL,
    requirement_id VARCHAR(50) NOT NULL,
    requirement_name VARCHAR(255) NOT NULL,
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Registro de não-conformidades
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

-- Registro de ações corretivas
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

-- Métricas de qualidade
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
```

### 2. Funções Principais

#### 2.1 Mapeamento de Padrões de Qualidade

```sql
-- Função para registrar mapeamento entre padrão de qualidade e validador
CREATE OR REPLACE FUNCTION quality_management.register_standard_mapping(
    p_standard_id VARCHAR(50),
    p_standard_code VARCHAR(50),
    p_standard_name VARCHAR(255),
    p_validator_id VARCHAR(50),
    p_validator_name VARCHAR(255),
    p_requirement_id VARCHAR(50),
    p_requirement_name VARCHAR(255),
    p_tenant_id UUID
) RETURNS UUID AS $$
DECLARE
    v_mapping_id UUID;
BEGIN
    INSERT INTO quality_management.standard_validator_mapping (
        standard_id, standard_code, standard_name,
        validator_id, validator_name,
        requirement_id, requirement_name,
        tenant_id
    ) VALUES (
        p_standard_id, p_standard_code, p_standard_name,
        p_validator_id, p_validator_name,
        p_requirement_id, p_requirement_name,
        p_tenant_id
    )
    RETURNING mapping_id INTO v_mapping_id;
    
    RETURN v_mapping_id;
END;
$$ LANGUAGE plpgsql;
```

#### 2.2 Criação de Não-Conformidades

```sql
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
    
    -- Criar ação corretiva automaticamente
    PERFORM quality_management.create_corrective_action(
        v_non_conformity_id,
        'CORRECTIVE',
        'Corrigir não-conformidade de ' || p_requirement_name,
        p_tenant_id
    );
    
    RETURN v_non_conformity_id;
END;
$$ LANGUAGE plpgsql;
```

#### 2.3 Geração de Ações Corretivas

```sql
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
    
    -- Atualizar métricas de qualidade
    PERFORM quality_management.update_quality_metrics(p_tenant_id);
    
    RETURN v_action_id;
END;
$$ LANGUAGE plpgsql;
```

#### 2.4 Atualização de Métricas de Qualidade

```sql
-- Função para atualizar métricas de qualidade
CREATE OR REPLACE FUNCTION quality_management.update_quality_metrics(
    p_tenant_id UUID
) RETURNS VOID AS $$
DECLARE
    v_total_non_conformities INTEGER;
    v_open_non_conformities INTEGER;
    v_compliance_rate NUMERIC;
    v_avg_resolution_days NUMERIC;
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
    
    -- Atualizar ou inserir métricas
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
END;
$$ LANGUAGE plpgsql;
```

### 3. Triggers e Automatização

```sql
-- Trigger para criar não-conformidades quando validação falhar
CREATE OR REPLACE FUNCTION quality_management.validation_result_trigger()
RETURNS TRIGGER AS $$
BEGIN
    -- Se a validação falhou, criar não-conformidade
    IF NEW.is_compliant = FALSE THEN
        PERFORM quality_management.create_non_conformity(
            NEW.validation_id,
            NEW.validator_id,
            NEW.validator_name,
            NEW.requirement_id,
            NEW.requirement_name,
            jsonb_build_object(
                'message', NEW.details,
                'criticality', NEW.criticality,
                'validation_date', NEW.validation_date
            ),
            NEW.tenant_id
        );
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Anexar o trigger à tabela de histórico de validação
CREATE TRIGGER after_validation_result
AFTER INSERT ON compliance_validators.validation_history
FOR EACH ROW
EXECUTE FUNCTION quality_management.validation_result_trigger();
```

## Configuração e Uso

### Mapeamento de Padrões de Qualidade

Para configurar o mapeamento entre validadores de conformidade e padrões de qualidade:

```sql
-- Exemplo: Mapear requisito HIPAA para ISO 9001:2015
SELECT quality_management.register_standard_mapping(
    'ISO9001', 'ISO 9001:2015', 'Sistema de Gestão da Qualidade',
    'HIPAA_VALIDATOR', 'Validador de Conformidade HIPAA',
    'HIPAA-01', 'Proteção de Dados de Saúde',
    '12345678-1234-1234-1234-123456789012'  -- tenant_id
);

-- Exemplo: Mapear requisito PSD2 para ISO 27001
SELECT quality_management.register_standard_mapping(
    'ISO27001', 'ISO/IEC 27001:2013', 'Sistema de Gestão de Segurança da Informação',
    'PSD2_VALIDATOR', 'Validador de Conformidade PSD2',
    'PSD2-SCA', 'Autenticação Forte de Cliente',
    '12345678-1234-1234-1234-123456789012'  -- tenant_id
);
```

### Integração com Testes de Validação

A integração com os validadores de conformidade é automática através do trigger. Quando um teste de validação é executado e falha, uma não-conformidade é criada automaticamente:

```sql
-- Executar validação HIPAA
SELECT * FROM compliance_validators.validate_hipaa_compliance('12345678-1234-1234-1234-123456789012');

-- Executar validação PSD2
SELECT * FROM compliance_validators.validate_psd2_compliance('12345678-1234-1234-1234-123456789012');
```

### Verificação de Não-Conformidades

Para verificar as não-conformidades geradas:

```sql
-- Listar todas as não-conformidades abertas
SELECT * FROM quality_management.non_conformity
WHERE status = 'OPEN'
AND tenant_id = '12345678-1234-1234-1234-123456789012';

-- Listar ações corretivas pendentes
SELECT nc.requirement_name, ca.description, ca.status, ca.due_date
FROM quality_management.corrective_action ca
JOIN quality_management.non_conformity nc ON ca.non_conformity_id = nc.non_conformity_id
WHERE ca.status = 'PENDING'
AND ca.tenant_id = '12345678-1234-1234-1234-123456789012';
```

### Métricas de Qualidade

Para visualizar as métricas de qualidade:

```sql
-- Visualizar todas as métricas de qualidade
SELECT metric_name, current_value, target_value, unit,
       CASE WHEN current_value >= target_value THEN 'META ATINGIDA' ELSE 'META NÃO ATINGIDA' END AS status
FROM quality_management.quality_metrics
WHERE tenant_id = '12345678-1234-1234-1234-123456789012';
```

## Integração com Outros Sistemas

### Integração com Sistema de Gestão de Riscos

A integração de qualidade se conecta com o Sistema de Gestão de Riscos para:

1. Converter não-conformidades críticas em riscos de conformidade
2. Associar riscos identificados a controles e padrões de qualidade
3. Utilizar a análise de impacto de riscos para priorizar ações corretivas

```sql
-- Função para correlacionar não-conformidade com risco
CREATE OR REPLACE FUNCTION quality_management.correlate_with_risk(
    p_non_conformity_id UUID,
    p_tenant_id UUID
) RETURNS UUID AS $$
DECLARE
    v_risk_id UUID;
    v_non_conformity RECORD;
BEGIN
    -- Obter detalhes da não-conformidade
    SELECT * INTO v_non_conformity
    FROM quality_management.non_conformity
    WHERE non_conformity_id = p_non_conformity_id
    AND tenant_id = p_tenant_id;
    
    -- Verificar se já existe risco associado
    SELECT risk_id INTO v_risk_id
    FROM risk_management.compliance_risk
    WHERE source_id = p_non_conformity_id::TEXT
    AND source_type = 'QUALITY_NON_CONFORMITY'
    AND tenant_id = p_tenant_id
    LIMIT 1;
    
    -- Se não existir, criar novo risco
    IF v_risk_id IS NULL THEN
        SELECT risk_management.register_compliance_risk(
            'Risco de conformidade: ' || v_non_conformity.requirement_name,
            'Não conformidade com padrão de qualidade ' || v_non_conformity.standard_name,
            CASE
                WHEN v_non_conformity.impact_level = 'MAJOR' THEN 'HIGH'
                WHEN v_non_conformity.impact_level = 'MODERATE' THEN 'MEDIUM'
                ELSE 'LOW'
            END,
            'QUALITY_NON_CONFORMITY',
            p_non_conformity_id::TEXT,
            p_tenant_id
        ) INTO v_risk_id;
    END IF;
    
    RETURN v_risk_id;
END;
$$ LANGUAGE plpgsql;
```

### Integração com Sistema de Gestão de Incidentes

A integração de qualidade também se conecta com o Sistema de Gestão de Incidentes para:

1. Converter não-conformidades críticas em incidentes de conformidade
2. Rastrear incidentes recorrentes relacionados a falhas de qualidade
3. Utilizar o processo de gestão de incidentes para resolver não-conformidades críticas

```sql
-- Função para converter não-conformidade em incidente
CREATE OR REPLACE FUNCTION quality_management.create_incident_from_non_conformity(
    p_non_conformity_id UUID,
    p_tenant_id UUID
) RETURNS UUID AS $$
DECLARE
    v_incident_id UUID;
    v_non_conformity RECORD;
BEGIN
    -- Obter detalhes da não-conformidade
    SELECT * INTO v_non_conformity
    FROM quality_management.non_conformity
    WHERE non_conformity_id = p_non_conformity_id
    AND tenant_id = p_tenant_id;
    
    -- Criar incidente apenas para não-conformidades críticas
    IF v_non_conformity.impact_level = 'MAJOR' THEN
        SELECT incident_management.register_compliance_incident(
            'Incidente de qualidade: ' || v_non_conformity.requirement_name,
            'Falha crítica em requisito de conformidade ' || v_non_conformity.requirement_name,
            'QUALITY',
            CASE
                WHEN v_non_conformity.impact_level = 'MAJOR' THEN 'HIGH'
                WHEN v_non_conformity.impact_level = 'MODERATE' THEN 'MEDIUM'
                ELSE 'LOW'
            END,
            v_non_conformity.details,
            p_tenant_id
        ) INTO v_incident_id;
    END IF;
    
    RETURN v_incident_id;
END;
$$ LANGUAGE plpgsql;
```

## Exemplos de Uso

### Exemplo 1: Validação de Conformidade e Geração de Ações Corretivas

O seguinte workflow demonstra o ciclo completo de validação e ação corretiva:

1. Executar validação de conformidade
2. A validação identifica não-conformidades
3. O sistema gera automaticamente não-conformidades no módulo de qualidade
4. Ações corretivas são criadas e atribuídas automaticamente
5. As métricas de qualidade são atualizadas
6. Riscos associados são registrados para não-conformidades críticas

```sql
-- Passo 1: Executar validação
SELECT * FROM compliance_validators.validate_hipaa_compliance('12345678-1234-1234-1234-123456789012');

-- Passo 2: Verificar não-conformidades geradas
SELECT * FROM quality_management.non_conformity
WHERE tenant_id = '12345678-1234-1234-1234-123456789012'
ORDER BY created_at DESC
LIMIT 10;

-- Passo 3: Verificar ações corretivas geradas
SELECT nc.requirement_name, ca.description, ca.status
FROM quality_management.corrective_action ca
JOIN quality_management.non_conformity nc ON ca.non_conformity_id = nc.non_conformity_id
WHERE ca.tenant_id = '12345678-1234-1234-1234-123456789012'
ORDER BY ca.created_at DESC
LIMIT 10;

-- Passo 4: Verificar métricas de qualidade
SELECT metric_name, current_value, target_value, unit
FROM quality_management.quality_metrics
WHERE tenant_id = '12345678-1234-1234-1234-123456789012';

-- Passo 5: Verificar riscos associados
SELECT r.risk_name, r.risk_level, r.status
FROM risk_management.compliance_risk r
WHERE r.source_type = 'QUALITY_NON_CONFORMITY'
AND r.tenant_id = '12345678-1234-1234-1234-123456789012';
```

### Exemplo 2: Resolução de Não-Conformidade e Avaliação de Eficácia

Este workflow demonstra o processo de resolução de não-conformidades:

1. Atualizar status da ação corretiva para "Em progresso"
2. Registrar a conclusão da ação corretiva
3. Avaliar a eficácia da ação corretiva
4. Fechar a não-conformidade se a ação for eficaz
5. Verificar melhoria nas métricas de qualidade

```sql
-- Passo 1: Atualizar status da ação corretiva
UPDATE quality_management.corrective_action
SET status = 'IN_PROGRESS', 
    assigned_to = 'Maria Silva',
    updated_at = CURRENT_TIMESTAMP
WHERE action_id = 'UUID-da-acao'
AND tenant_id = '12345678-1234-1234-1234-123456789012';

-- Passo 2: Registrar conclusão da ação corretiva
UPDATE quality_management.corrective_action
SET status = 'COMPLETED', 
    completed_date = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE action_id = 'UUID-da-acao'
AND tenant_id = '12345678-1234-1234-1234-123456789012';

-- Passo 3: Avaliar eficácia da ação
UPDATE quality_management.corrective_action
SET effectiveness_evaluation = 'A ação corretiva implementou controles adequados e nova validação confirmou conformidade.',
    effectiveness_status = 'EFFECTIVE',
    updated_at = CURRENT_TIMESTAMP
WHERE action_id = 'UUID-da-acao'
AND tenant_id = '12345678-1234-1234-1234-123456789012';

-- Passo 4: Fechar a não-conformidade
UPDATE quality_management.non_conformity
SET status = 'CLOSED',
    updated_at = CURRENT_TIMESTAMP
WHERE non_conformity_id = (
    SELECT non_conformity_id 
    FROM quality_management.corrective_action
    WHERE action_id = 'UUID-da-acao'
)
AND tenant_id = '12345678-1234-1234-1234-123456789012';

-- Passo 5: Atualizar métricas de qualidade
SELECT quality_management.update_quality_metrics('12345678-1234-1234-1234-123456789012');
```

## Ciclo de Melhoria Contínua

O sistema implementa o ciclo PDCA (Plan-Do-Check-Act) para melhoria contínua:

1. **Plan (Planejar)**: Mapeamento de validadores para padrões de qualidade
2. **Do (Fazer)**: Execução de validações de conformidade
3. **Check (Verificar)**: Identificação automática de não-conformidades
4. **Act (Agir)**: Implementação de ações corretivas e preventivas

Este ciclo é integrado ao sistema de gestão da qualidade e permite:

- Identificação proativa de problemas de conformidade
- Rastreamento de tendências e padrões de não-conformidades
- Melhoria contínua dos processos de segurança e gestão de identidade
- Documentação automática para auditorias e certificações

## Referências

- ISO 9001:2015 - Sistema de Gestão da Qualidade
- ISO/IEC 27001:2013 - Sistema de Gestão de Segurança da Informação
- ISO 19011:2018 - Diretrizes para Auditoria de Sistemas de Gestão
- ISO 31000:2018 - Gestão de Riscos
- Normas setoriais: ISO 13485, HIPAA, PSD2, eIDAS
