-- INNOVABIZ - IAM HIMSS EMRAM Schema
-- Author: Eduardo Jeremias
-- Date: 08/05/2025
-- Version: 1.0
-- Description: Esquema para armazenamento e gestão de dados relacionados ao modelo HIMSS EMRAM (Electronic Medical Record Adoption Model).

-- Configurar caminho de busca
SET search_path TO iam, public;

-- Tabela de Estágios HIMSS EMRAM
CREATE TABLE IF NOT EXISTS himss_emram_stages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stage_number INTEGER NOT NULL UNIQUE CHECK (stage_number BETWEEN 0 AND 7),
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    cumulative BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_himss_emram_stages_stage_number ON himss_emram_stages(stage_number);

-- Tabela de Critérios HIMSS EMRAM
CREATE TABLE IF NOT EXISTS himss_emram_criteria (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stage_id UUID NOT NULL REFERENCES himss_emram_stages(id),
    criteria_code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(100),
    is_mandatory BOOLEAN DEFAULT TRUE NOT NULL,
    validation_rules JSONB,
    implementation_guidance TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    CONSTRAINT himss_emram_criteria_stage_code_unique UNIQUE (stage_id, criteria_code)
);

CREATE INDEX IF NOT EXISTS idx_himss_emram_criteria_stage_id ON himss_emram_criteria(stage_id);
CREATE INDEX IF NOT EXISTS idx_himss_emram_criteria_category ON himss_emram_criteria(category);
CREATE INDEX IF NOT EXISTS idx_himss_emram_criteria_is_active ON himss_emram_criteria(is_active);
CREATE INDEX IF NOT EXISTS idx_himss_emram_criteria_is_mandatory ON himss_emram_criteria(is_mandatory);

-- Tabela de Avaliações HIMSS EMRAM
CREATE TABLE IF NOT EXISTS himss_emram_assessments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    healthcare_facility_name VARCHAR(255) NOT NULL,
    facility_type VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) NOT NULL DEFAULT 'in_progress',
    target_stage INTEGER NOT NULL CHECK (target_stage BETWEEN 0 AND 7),
    current_stage INTEGER CHECK (current_stage BETWEEN 0 AND 7),
    scope JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    primary_contact_id UUID REFERENCES users(id),
    previous_assessment_id UUID REFERENCES himss_emram_assessments(id),
    metadata JSONB DEFAULT '{}'::JSONB,
    CONSTRAINT himss_emram_assessments_status_valid_values CHECK (status IN ('planned', 'in_progress', 'completed', 'certified', 'expired', 'canceled'))
);

CREATE INDEX IF NOT EXISTS idx_himss_emram_assessments_organization_id ON himss_emram_assessments(organization_id);
CREATE INDEX IF NOT EXISTS idx_himss_emram_assessments_status ON himss_emram_assessments(status);
CREATE INDEX IF NOT EXISTS idx_himss_emram_assessments_target_stage ON himss_emram_assessments(target_stage);
CREATE INDEX IF NOT EXISTS idx_himss_emram_assessments_current_stage ON himss_emram_assessments(current_stage);
CREATE INDEX IF NOT EXISTS idx_himss_emram_assessments_start_date ON himss_emram_assessments(start_date);
CREATE INDEX IF NOT EXISTS idx_himss_emram_assessments_facility_type ON himss_emram_assessments(facility_type);

-- Tabela de Resultados de Critérios HIMSS EMRAM
CREATE TABLE IF NOT EXISTS himss_emram_criteria_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    assessment_id UUID NOT NULL REFERENCES himss_emram_assessments(id) ON DELETE CASCADE,
    criteria_id UUID NOT NULL REFERENCES himss_emram_criteria(id),
    status VARCHAR(50) NOT NULL,
    compliance_percentage FLOAT,
    evidence TEXT,
    notes TEXT,
    implementation_status VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    issues_found JSONB DEFAULT '[]'::JSONB,
    recommendations JSONB DEFAULT '[]'::JSONB,
    validation_data JSONB,
    CONSTRAINT himss_emram_criteria_results_status_valid_values CHECK (status IN ('compliant', 'partial_compliance', 'non_compliant', 'not_applicable')),
    CONSTRAINT himss_emram_criteria_results_implementation_valid_values CHECK (implementation_status IN ('not_implemented', 'partially_implemented', 'implemented', 'not_applicable'))
);

CREATE INDEX IF NOT EXISTS idx_himss_emram_criteria_results_assessment_id ON himss_emram_criteria_results(assessment_id);
CREATE INDEX IF NOT EXISTS idx_himss_emram_criteria_results_criteria_id ON himss_emram_criteria_results(criteria_id);
CREATE INDEX IF NOT EXISTS idx_himss_emram_criteria_results_status ON himss_emram_criteria_results(status);
CREATE INDEX IF NOT EXISTS idx_himss_emram_criteria_results_implementation_status ON himss_emram_criteria_results(implementation_status);

-- Tabela de Certificações HIMSS EMRAM
CREATE TABLE IF NOT EXISTS himss_emram_certifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    assessment_id UUID NOT NULL REFERENCES himss_emram_assessments(id),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    healthcare_facility_name VARCHAR(255) NOT NULL,
    certification_date TIMESTAMP WITH TIME ZONE NOT NULL,
    expiration_date TIMESTAMP WITH TIME ZONE NOT NULL,
    stage_achieved INTEGER NOT NULL CHECK (stage_achieved BETWEEN 0 AND 7),
    certificate_number VARCHAR(100) UNIQUE,
    certifying_body VARCHAR(255) NOT NULL,
    certifying_assessor VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    certificate_url VARCHAR(255),
    CONSTRAINT himss_emram_certifications_status_valid_values CHECK (status IN ('active', 'expired', 'revoked', 'suspended'))
);

CREATE INDEX IF NOT EXISTS idx_himss_emram_certifications_assessment_id ON himss_emram_certifications(assessment_id);
CREATE INDEX IF NOT EXISTS idx_himss_emram_certifications_organization_id ON himss_emram_certifications(organization_id);
CREATE INDEX IF NOT EXISTS idx_himss_emram_certifications_stage_achieved ON himss_emram_certifications(stage_achieved);
CREATE INDEX IF NOT EXISTS idx_himss_emram_certifications_status ON himss_emram_certifications(status);
CREATE INDEX IF NOT EXISTS idx_himss_emram_certifications_certification_date ON himss_emram_certifications(certification_date);
CREATE INDEX IF NOT EXISTS idx_himss_emram_certifications_expiration_date ON himss_emram_certifications(expiration_date);

-- Tabela de Planos de Ação HIMSS EMRAM
CREATE TABLE IF NOT EXISTS himss_emram_action_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    assessment_id UUID REFERENCES himss_emram_assessments(id),
    criteria_result_id UUID REFERENCES himss_emram_criteria_results(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    priority VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'open',
    due_date DATE,
    assigned_to UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    completed_at TIMESTAMP WITH TIME ZONE,
    completed_by UUID REFERENCES users(id),
    completion_notes TEXT,
    estimated_effort VARCHAR(100),
    target_stage INTEGER CHECK (target_stage BETWEEN 1 AND 7),
    CONSTRAINT himss_emram_action_plans_priority_valid_values CHECK (priority IN ('critical', 'high', 'medium', 'low')),
    CONSTRAINT himss_emram_action_plans_status_valid_values CHECK (status IN ('open', 'in_progress', 'completed', 'deferred', 'canceled'))
);

CREATE INDEX IF NOT EXISTS idx_himss_emram_action_plans_organization_id ON himss_emram_action_plans(organization_id);
CREATE INDEX IF NOT EXISTS idx_himss_emram_action_plans_assessment_id ON himss_emram_action_plans(assessment_id);
CREATE INDEX IF NOT EXISTS idx_himss_emram_action_plans_criteria_result_id ON himss_emram_action_plans(criteria_result_id);
CREATE INDEX IF NOT EXISTS idx_himss_emram_action_plans_status ON himss_emram_action_plans(status);
CREATE INDEX IF NOT EXISTS idx_himss_emram_action_plans_priority ON himss_emram_action_plans(priority);
CREATE INDEX IF NOT EXISTS idx_himss_emram_action_plans_due_date ON himss_emram_action_plans(due_date);
CREATE INDEX IF NOT EXISTS idx_himss_emram_action_plans_assigned_to ON himss_emram_action_plans(assigned_to);
CREATE INDEX IF NOT EXISTS idx_himss_emram_action_plans_target_stage ON himss_emram_action_plans(target_stage);

-- Tabela de Benchmarks HIMSS EMRAM
CREATE TABLE IF NOT EXISTS himss_emram_benchmarks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    year INTEGER NOT NULL,
    country VARCHAR(100) NOT NULL,
    region VARCHAR(100),
    facility_type VARCHAR(100) NOT NULL,
    stage0_percentage FLOAT,
    stage1_percentage FLOAT,
    stage2_percentage FLOAT,
    stage3_percentage FLOAT,
    stage4_percentage FLOAT,
    stage5_percentage FLOAT,
    stage6_percentage FLOAT,
    stage7_percentage FLOAT,
    avg_stage FLOAT,
    median_stage INTEGER,
    sample_size INTEGER,
    source VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    notes TEXT,
    CONSTRAINT himss_emram_benchmarks_unique_key UNIQUE (year, country, region, facility_type)
);

CREATE INDEX IF NOT EXISTS idx_himss_emram_benchmarks_year ON himss_emram_benchmarks(year);
CREATE INDEX IF NOT EXISTS idx_himss_emram_benchmarks_country ON himss_emram_benchmarks(country);
CREATE INDEX IF NOT EXISTS idx_himss_emram_benchmarks_region ON himss_emram_benchmarks(region);
CREATE INDEX IF NOT EXISTS idx_himss_emram_benchmarks_facility_type ON himss_emram_benchmarks(facility_type);

-- Função para atualizar o timestamp 'updated_at'
CREATE OR REPLACE FUNCTION update_himss_emram_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Aplicar a função aos triggers
CREATE TRIGGER update_himss_emram_stages_updated_at
BEFORE UPDATE ON himss_emram_stages
FOR EACH ROW EXECUTE FUNCTION update_himss_emram_updated_at_column();

CREATE TRIGGER update_himss_emram_criteria_updated_at
BEFORE UPDATE ON himss_emram_criteria
FOR EACH ROW EXECUTE FUNCTION update_himss_emram_updated_at_column();

CREATE TRIGGER update_himss_emram_assessments_updated_at
BEFORE UPDATE ON himss_emram_assessments
FOR EACH ROW EXECUTE FUNCTION update_himss_emram_updated_at_column();

CREATE TRIGGER update_himss_emram_criteria_results_updated_at
BEFORE UPDATE ON himss_emram_criteria_results
FOR EACH ROW EXECUTE FUNCTION update_himss_emram_updated_at_column();

CREATE TRIGGER update_himss_emram_certifications_updated_at
BEFORE UPDATE ON himss_emram_certifications
FOR EACH ROW EXECUTE FUNCTION update_himss_emram_updated_at_column();

CREATE TRIGGER update_himss_emram_action_plans_updated_at
BEFORE UPDATE ON himss_emram_action_plans
FOR EACH ROW EXECUTE FUNCTION update_himss_emram_updated_at_column();

CREATE TRIGGER update_himss_emram_benchmarks_updated_at
BEFORE UPDATE ON himss_emram_benchmarks
FOR EACH ROW EXECUTE FUNCTION update_himss_emram_updated_at_column();

-- Inserir estágios HIMSS EMRAM
INSERT INTO himss_emram_stages (stage_number, name, description, cumulative)
VALUES
    (0, 'Estágio 0', 'Organização não utiliza as três aplicações médicas departamentais principais: laboratório, farmácia e radiologia.', TRUE),
    (1, 'Estágio 1', 'Todas as três aplicações médicas departamentais principais estão instaladas.', TRUE),
    (2, 'Estágio 2', 'Sistema CDR (Clinical Data Repository), controle de terminologia clínica, visualizador clínico e regras básicas de suporte à decisão clínica implementados.', TRUE),
    (3, 'Estágio 3', 'Documentação clínica de enfermagem e médica, suporte à decisão clínica para verificação de erros de prescrição e PACS disponível fora do departamento de radiologia.', TRUE),
    (4, 'Estágio 4', 'CPOE (Computerized Physician Order Entry) implementado para medicações e suporte à decisão clínica baseado em protocolos.', TRUE),
    (5, 'Estágio 5', 'Sistema de administração de medicamentos com código de barras implementado e sistema de documentação de medicamentos fechado com a farmácia.', TRUE),
    (6, 'Estágio 6', 'Documentação médica completa com modelos estruturados, suporte avançado à decisão clínica e capacidade de compartilhamento de dados entre unidades.', TRUE),
    (7, 'Estágio 7', 'Registro médico eletrônico completo, análise de dados clínicos para melhoria da qualidade e compartilhamento de informações por meio de troca de registros eletrônicos.', TRUE)
ON CONFLICT (stage_number) DO UPDATE
    SET name = EXCLUDED.name,
        description = EXCLUDED.description,
        cumulative = EXCLUDED.cumulative,
        updated_at = NOW();

-- Inserir alguns critérios para os estágios
INSERT INTO himss_emram_criteria (stage_id, criteria_code, name, description, category, is_mandatory, validation_rules, implementation_guidance)
VALUES
    -- Estágio 1
    (
        (SELECT id FROM himss_emram_stages WHERE stage_number = 1),
        'S1.C1',
        'Sistema de Laboratório',
        'Um sistema de informação laboratorial (LIS) capaz de processar pedidos e resultados de exames laboratoriais deve estar operacional.',
        'Sistemas Departamentais',
        TRUE,
        '[{"key": "laboratory_system.operational", "expected": true}, {"key": "laboratory_system.order_processing", "expected": true}, {"key": "laboratory_system.results_processing", "expected": true}]',
        'Implementar um LIS que permita registro, processamento e consulta de resultados laboratoriais. O sistema deve permitir integração com analisadores laboratoriais e com o sistema central.'
    ),
    (
        (SELECT id FROM himss_emram_stages WHERE stage_number = 1),
        'S1.C2',
        'Sistema de Farmácia',
        'Um sistema de informação farmacêutica capaz de gerenciar medicamentos, incluindo pedido, dispensação e inventário deve estar operacional.',
        'Sistemas Departamentais',
        TRUE,
        '[{"key": "pharmacy_system.operational", "expected": true}, {"key": "pharmacy_system.medication_management", "expected": true}, {"key": "pharmacy_system.inventory_control", "expected": true}]',
        'Implementar um sistema farmacêutico para gerenciar todos os aspectos de medicamentos, incluindo a capacidade de verificar interações medicamentosas básicas.'
    ),
    (
        (SELECT id FROM himss_emram_stages WHERE stage_number = 1),
        'S1.C3',
        'Sistema de Radiologia',
        'Um sistema de informação radiológica (RIS) capaz de gerenciar exames radiológicos, incluindo agendamento e rastreamento deve estar operacional.',
        'Sistemas Departamentais',
        TRUE,
        '[{"key": "radiology_system.operational", "expected": true}, {"key": "radiology_system.exam_management", "expected": true}, {"key": "radiology_system.scheduling", "expected": true}]',
        'Implementar um RIS que permita agendamento, registro e rastreamento de exames radiológicos.'
    ),

    -- Estágio 2
    (
        (SELECT id FROM himss_emram_stages WHERE stage_number = 2),
        'S2.C1',
        'Repositório de Dados Clínicos (CDR)',
        'Um CDR que centraliza os dados clínicos de diferentes sistemas deve estar implementado e operacional.',
        'Infraestrutura de Dados',
        TRUE,
        '[{"key": "cdr.operational", "expected": true}, {"key": "cdr.data_integration", "expected": true}, {"key": "cdr.clinical_data_storage", "expected": true}]',
        'Implementar um CDR que armazene dados estruturados dos sistemas departamentais e proporcione um registro longitudinal dos pacientes.'
    ),
    (
        (SELECT id FROM himss_emram_stages WHERE stage_number = 2),
        'S2.C2',
        'Controle de Terminologia Clínica',
        'Um sistema de terminologia clínica controlada que padroniza termos médicos e códigos deve estar implementado.',
        'Dados Clínicos',
        TRUE,
        '[{"key": "terminology.standardized", "expected": true}, {"key": "terminology.codes_implemented", "expected": true}, {"key": "terminology.mapping_capabilities", "expected": true}]',
        'Implementar padrões de terminologia como SNOMED CT, LOINC, CID-10 ou similares para padronizar a codificação dos dados clínicos.'
    ),
    (
        (SELECT id FROM himss_emram_stages WHERE stage_number = 2),
        'S2.C3',
        'Visualizador Clínico',
        'Um visualizador que permite acesso aos dados clínicos do paciente provenientes do CDR deve estar disponível.',
        'Interface Clínica',
        TRUE,
        '[{"key": "clinical_viewer.operational", "expected": true}, {"key": "clinical_viewer.access_to_cdr", "expected": true}, {"key": "clinical_viewer.user_friendly", "expected": true}]',
        'Implementar uma interface que permita visualizar dados clínicos integrados do paciente, incluindo resultados de laboratório, medicamentos e informações demográficas.'
    ),
    
    -- Estágio 3
    (
        (SELECT id FROM himss_emram_stages WHERE stage_number = 3),
        'S3.C1',
        'Documentação de Enfermagem',
        'Documentação eletrônica de enfermagem, incluindo avaliações, planos de cuidados e notas de evolução deve estar implementada.',
        'Documentação Clínica',
        TRUE,
        '[{"key": "nursing_documentation.electronic", "expected": true}, {"key": "nursing_documentation.assessments", "expected": true}, {"key": "nursing_documentation.care_plans", "expected": true}]',
        'Implementar um sistema de documentação de enfermagem estruturado que permita registro padronizado de avaliações, planos de cuidados e evoluções.'
    ),
    (
        (SELECT id FROM himss_emram_stages WHERE stage_number = 3),
        'S3.C2',
        'Documentação Médica',
        'Documentação eletrônica médica, incluindo históricos, evolução e prescrições, deve estar implementada em algumas áreas clínicas.',
        'Documentação Clínica',
        TRUE,
        '[{"key": "physician_documentation.electronic", "expected": true}, {"key": "physician_documentation.progress_notes", "expected": true}, {"key": "physician_documentation.history_physical", "expected": true}]',
        'Implementar capacidades de documentação médica eletrônica, com modelos estruturados para facilitar o registro de dados clínicos relevantes.'
    ),
    (
        (SELECT id FROM himss_emram_stages WHERE stage_number = 3),
        'S3.C3',
        'Suporte à Decisão para Prescrição',
        'Suporte à decisão clínica básico para verificação de erros em prescrições de medicamentos deve estar implementado.',
        'Suporte à Decisão',
        TRUE,
        '[{"key": "medication_decision_support.drug_interactions", "expected": true}, {"key": "medication_decision_support.allergy_checking", "expected": true}, {"key": "medication_decision_support.dosage_checking", "expected": true}]',
        'Implementar verificações de interações medicamentosas, alergias e doses como parte do processo de prescrição eletrônica.'
    ),
    
    -- Estágio 4
    (
        (SELECT id FROM himss_emram_stages WHERE stage_number = 4),
        'S4.C1',
        'CPOE para Medicamentos',
        'Entrada computadorizada de pedidos médicos (CPOE) para medicamentos implementada em pelo menos uma unidade de internação.',
        'Prescrição Eletrônica',
        TRUE,
        '[{"key": "cpoe_medications.implemented", "expected": true}, {"key": "cpoe_medications.structured_entry", "expected": true}, {"key": "cpoe_medications.adoption_rate", "expected": true}]',
        'Implementar CPOE para medicamentos em pelo menos uma unidade de internação, com validação automática e regras de prescrição segura.'
    ),
    (
        (SELECT id FROM himss_emram_stages WHERE stage_number = 4),
        'S4.C2',
        'Suporte à Decisão Baseado em Protocolos',
        'Suporte à decisão clínica baseado em protocolos clínicos e diretrizes deve estar disponível para condições clínicas selecionadas.',
        'Suporte à Decisão',
        TRUE,
        '[{"key": "protocol_decision_support.clinical_protocols", "expected": true}, {"key": "protocol_decision_support.evidence_based", "expected": true}, {"key": "protocol_decision_support.actionable_alerts", "expected": true}]',
        'Implementar regras baseadas em protocolos clínicos que possam guiar o atendimento ao paciente e fornecer alertas acionáveis.'
    ),
    
    -- Estágio 5
    (
        (SELECT id FROM himss_emram_stages WHERE stage_number = 5),
        'S5.C1',
        'Administração de Medicamentos com Código de Barras',
        'Sistema de administração de medicamentos com código de barras (BCMA) implementado com os cinco certos de medicação.',
        'Segurança de Medicamentos',
        TRUE,
        '[{"key": "bcma.implemented", "expected": true}, {"key": "bcma.five_rights", "expected": true}, {"key": "bcma.closed_loop", "expected": true}, {"key": "bcma.adoption_rate", "expected": true}]',
        'Implementar um sistema BCMA que valide os cinco certos: paciente certo, medicamento certo, dose certa, via certa, hora certa.'
    ),
    (
        (SELECT id FROM himss_emram_stages WHERE stage_number = 5),
        'S5.C2',
        'Sistema Fechado de Medicação',
        'Ciclo fechado de medicação entre prescrição, farmácia e administração, garantindo rastreabilidade completa.',
        'Segurança de Medicamentos',
        TRUE,
        '[{"key": "closed_loop_medication.implemented", "expected": true}, {"key": "closed_loop_medication.full_traceability", "expected": true}, {"key": "closed_loop_medication.pharmacy_integration", "expected": true}]',
        'Implementar um sistema que integre completamente o processo de medicação desde a prescrição até a administração, com validação em cada etapa.'
    ),
    
    -- Estágio 6
    (
        (SELECT id FROM himss_emram_stages WHERE stage_number = 6),
        'S6.C1',
        'Documentação Médica Estruturada',
        'Documentação médica completa e estruturada implementada em todas as áreas clínicas e especialidades.',
        'Documentação Clínica',
        TRUE,
        '[{"key": "structured_documentation.all_areas", "expected": true}, {"key": "structured_documentation.templates", "expected": true}, {"key": "structured_documentation.discrete_data", "expected": true}, {"key": "structured_documentation.adoption_rate", "expected": true}]',
        'Implementar documentação estruturada em todas as áreas clínicas, com dados discretos que permitam análise e relatórios.'
    ),
    (
        (SELECT id FROM himss_emram_stages WHERE stage_number = 6),
        'S6.C2',
        'Suporte à Decisão Clínica Avançado',
        'Suporte à decisão clínica avançado, incluindo alertas baseados em condição, variances de protocolos e guias de prática clínica.',
        'Suporte à Decisão',
        TRUE,
        '[{"key": "advanced_cdss.condition_specific", "expected": true}, {"key": "advanced_cdss.variance_detection", "expected": true}, {"key": "advanced_cdss.clinical_pathways", "expected": true}, {"key": "advanced_cdss.usage_metrics", "expected": true}]',
        'Implementar regras avançadas de suporte à decisão que incorporem múltiplas fontes de dados para gerar alertas e recomendações significativos.'
    ),
    
    -- Estágio 7
    (
        (SELECT id FROM himss_emram_stages WHERE stage_number = 7),
        'S7.C1',
        'Análise de Dados Clínicos',
        'Capacidade de analisar dados clínicos para melhorar a qualidade, segurança e eficiência do atendimento ao paciente.',
        'Análise e BI',
        TRUE,
        '[{"key": "clinical_analytics.data_warehouse", "expected": true}, {"key": "clinical_analytics.quality_dashboards", "expected": true}, {"key": "clinical_analytics.predictive_models", "expected": true}, {"key": "clinical_analytics.operational_use", "expected": true}]',
        'Implementar um data warehouse clínico e ferramentas analíticas que permitam análise retrospectiva e preditiva dos dados de pacientes.'
    ),
    (
        (SELECT id FROM himss_emram_stages WHERE stage_number = 7),
        'S7.C2',
        'Interoperabilidade e Compartilhamento de Dados',
        'Compartilhamento de dados clínicos com outras organizações e participação em iniciativas de saúde populacional.',
        'Interoperabilidade',
        TRUE,
        '[{"key": "interoperability.health_information_exchange", "expected": true}, {"key": "interoperability.standard_formats", "expected": true}, {"key": "interoperability.population_health", "expected": true}, {"key": "interoperability.patient_access", "expected": true}]',
        'Implementar integração com sistemas de troca de informações em saúde e garantir conformidade com padrões como HL7 FHIR, DICOM, etc.'
    )
ON CONFLICT DO NOTHING;

COMMENT ON TABLE himss_emram_stages IS 'Estágios definidos pelo modelo HIMSS EMRAM para adoção de registros médicos eletrônicos';
COMMENT ON TABLE himss_emram_criteria IS 'Critérios para cada estágio do modelo HIMSS EMRAM';
COMMENT ON TABLE himss_emram_assessments IS 'Avaliações HIMSS EMRAM realizadas para organizações de saúde';
COMMENT ON TABLE himss_emram_criteria_results IS 'Resultados da avaliação de cada critério HIMSS EMRAM';
COMMENT ON TABLE himss_emram_certifications IS 'Certificações HIMSS EMRAM alcançadas por organizações de saúde';
COMMENT ON TABLE himss_emram_action_plans IS 'Planos de ação para endereçar não conformidades com HIMSS EMRAM';
COMMENT ON TABLE himss_emram_benchmarks IS 'Dados de benchmark HIMSS EMRAM por região, país e tipo de estabelecimento';
