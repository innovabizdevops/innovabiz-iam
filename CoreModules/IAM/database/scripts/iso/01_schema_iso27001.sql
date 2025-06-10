-- INNOVABIZ - IAM ISO 27001 Schema
-- Author: Eduardo Jeremias
-- Date: 08/05/2025
-- Version: 1.0
-- Description: Esquema para armazenamento e gestão de dados relacionados ao padrão ISO 27001.

-- Configurar caminho de busca
SET search_path TO iam, public;

-- Tabela de Controles ISO 27001
CREATE TABLE IF NOT EXISTS iso27001_controls (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    control_id VARCHAR(50) NOT NULL UNIQUE,
    section VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    healthcare_applicability TEXT,
    implementation_guidance TEXT,
    validation_rules JSONB,
    reference_links JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    category VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_iso27001_controls_control_id ON iso27001_controls(control_id);
CREATE INDEX IF NOT EXISTS idx_iso27001_controls_section ON iso27001_controls(section);
CREATE INDEX IF NOT EXISTS idx_iso27001_controls_category ON iso27001_controls(category);
CREATE INDEX IF NOT EXISTS idx_iso27001_controls_is_active ON iso27001_controls(is_active);

-- Tabela de Avaliações ISO 27001
CREATE TABLE IF NOT EXISTS iso27001_assessments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) NOT NULL DEFAULT 'in_progress',
    scope JSONB NOT NULL,
    version VARCHAR(50) DEFAULT '2013',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    healthcare_specific BOOLEAN DEFAULT FALSE,
    score FLOAT,
    framework_id UUID REFERENCES regulatory_frameworks(id),
    metadata JSONB DEFAULT '{}'::JSONB,
    CONSTRAINT iso27001_assessments_status_valid_values CHECK (status IN ('planned', 'in_progress', 'completed', 'canceled'))
);

CREATE INDEX IF NOT EXISTS idx_iso27001_assessments_organization_id ON iso27001_assessments(organization_id);
CREATE INDEX IF NOT EXISTS idx_iso27001_assessments_status ON iso27001_assessments(status);
CREATE INDEX IF NOT EXISTS idx_iso27001_assessments_version ON iso27001_assessments(version);
CREATE INDEX IF NOT EXISTS idx_iso27001_assessments_healthcare_specific ON iso27001_assessments(healthcare_specific);
CREATE INDEX IF NOT EXISTS idx_iso27001_assessments_start_date ON iso27001_assessments(start_date);
CREATE INDEX IF NOT EXISTS idx_iso27001_assessments_framework_id ON iso27001_assessments(framework_id);

-- Tabela de Resultados de Avaliação de Controles ISO 27001
CREATE TABLE IF NOT EXISTS iso27001_control_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    assessment_id UUID NOT NULL REFERENCES iso27001_assessments(id) ON DELETE CASCADE,
    control_id UUID NOT NULL REFERENCES iso27001_controls(id),
    status VARCHAR(50) NOT NULL,
    score FLOAT,
    evidence TEXT,
    notes TEXT,
    implementation_status VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    issues_found JSONB DEFAULT '[]'::JSONB,
    recommendations JSONB DEFAULT '[]'::JSONB,
    healthcare_specific_findings TEXT,
    CONSTRAINT iso27001_control_results_status_valid_values CHECK (status IN ('compliant', 'partial_compliance', 'non_compliant', 'not_applicable')),
    CONSTRAINT iso27001_control_results_implementation_valid_values CHECK (implementation_status IN ('not_implemented', 'partially_implemented', 'implemented', 'not_applicable'))
);

CREATE INDEX IF NOT EXISTS idx_iso27001_control_results_assessment_id ON iso27001_control_results(assessment_id);
CREATE INDEX IF NOT EXISTS idx_iso27001_control_results_control_id ON iso27001_control_results(control_id);
CREATE INDEX IF NOT EXISTS idx_iso27001_control_results_status ON iso27001_control_results(status);
CREATE INDEX IF NOT EXISTS idx_iso27001_control_results_implementation_status ON iso27001_control_results(implementation_status);

-- Tabela de Mapeamento ISO 27001 para Outros Frameworks
CREATE TABLE IF NOT EXISTS iso27001_framework_mapping (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    iso_control_id UUID NOT NULL REFERENCES iso27001_controls(id),
    framework_id UUID NOT NULL REFERENCES regulatory_frameworks(id),
    framework_control_id VARCHAR(100) NOT NULL,
    framework_control_name VARCHAR(255),
    mapping_type VARCHAR(50) NOT NULL,
    mapping_strength VARCHAR(50),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    CONSTRAINT iso27001_framework_mapping_type_valid_values CHECK (mapping_type IN ('one_to_one', 'one_to_many', 'partial', 'implied')),
    CONSTRAINT iso27001_framework_mapping_strength_valid_values CHECK (mapping_strength IN ('strong', 'moderate', 'weak'))
);

CREATE INDEX IF NOT EXISTS idx_iso27001_framework_mapping_iso_control_id ON iso27001_framework_mapping(iso_control_id);
CREATE INDEX IF NOT EXISTS idx_iso27001_framework_mapping_framework_id ON iso27001_framework_mapping(framework_id);
CREATE INDEX IF NOT EXISTS idx_iso27001_framework_mapping_framework_control_id ON iso27001_framework_mapping(framework_control_id);
CREATE INDEX IF NOT EXISTS idx_iso27001_framework_mapping_mapping_type ON iso27001_framework_mapping(mapping_type);
CREATE INDEX IF NOT EXISTS idx_iso27001_framework_mapping_mapping_strength ON iso27001_framework_mapping(mapping_strength);

-- Tabela de Planos de Ação ISO 27001
CREATE TABLE IF NOT EXISTS iso27001_action_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    assessment_id UUID REFERENCES iso27001_assessments(id),
    control_result_id UUID REFERENCES iso27001_control_results(id),
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
    healthcare_related BOOLEAN DEFAULT FALSE,
    CONSTRAINT iso27001_action_plans_priority_valid_values CHECK (priority IN ('critical', 'high', 'medium', 'low')),
    CONSTRAINT iso27001_action_plans_status_valid_values CHECK (status IN ('open', 'in_progress', 'completed', 'deferred', 'canceled'))
);

CREATE INDEX IF NOT EXISTS idx_iso27001_action_plans_organization_id ON iso27001_action_plans(organization_id);
CREATE INDEX IF NOT EXISTS idx_iso27001_action_plans_assessment_id ON iso27001_action_plans(assessment_id);
CREATE INDEX IF NOT EXISTS idx_iso27001_action_plans_control_result_id ON iso27001_action_plans(control_result_id);
CREATE INDEX IF NOT EXISTS idx_iso27001_action_plans_status ON iso27001_action_plans(status);
CREATE INDEX IF NOT EXISTS idx_iso27001_action_plans_priority ON iso27001_action_plans(priority);
CREATE INDEX IF NOT EXISTS idx_iso27001_action_plans_due_date ON iso27001_action_plans(due_date);
CREATE INDEX IF NOT EXISTS idx_iso27001_action_plans_assigned_to ON iso27001_action_plans(assigned_to);
CREATE INDEX IF NOT EXISTS idx_iso27001_action_plans_healthcare_related ON iso27001_action_plans(healthcare_related);

-- Tabela de Documentos ISO 27001
CREATE TABLE IF NOT EXISTS iso27001_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    title VARCHAR(255) NOT NULL,
    document_type VARCHAR(100) NOT NULL,
    description TEXT,
    version VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    content_url VARCHAR(255),
    storage_path VARCHAR(255),
    file_type VARCHAR(50),
    file_size BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMP WITH TIME ZONE,
    related_controls JSONB DEFAULT '[]'::JSONB,
    last_review_date TIMESTAMP WITH TIME ZONE,
    next_review_date TIMESTAMP WITH TIME ZONE,
    healthcare_specific BOOLEAN DEFAULT FALSE,
    CONSTRAINT iso27001_documents_status_valid_values CHECK (status IN ('draft', 'under_review', 'approved', 'published', 'retired'))
);

CREATE INDEX IF NOT EXISTS idx_iso27001_documents_organization_id ON iso27001_documents(organization_id);
CREATE INDEX IF NOT EXISTS idx_iso27001_documents_document_type ON iso27001_documents(document_type);
CREATE INDEX IF NOT EXISTS idx_iso27001_documents_status ON iso27001_documents(status);
CREATE INDEX IF NOT EXISTS idx_iso27001_documents_healthcare_specific ON iso27001_documents(healthcare_specific);
CREATE INDEX IF NOT EXISTS idx_iso27001_documents_next_review_date ON iso27001_documents(next_review_date);

-- Função para atualizar o timestamp 'updated_at'
CREATE OR REPLACE FUNCTION update_iso27001_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Aplicar a função aos triggers
CREATE TRIGGER update_iso27001_controls_updated_at
BEFORE UPDATE ON iso27001_controls
FOR EACH ROW EXECUTE FUNCTION update_iso27001_updated_at_column();

CREATE TRIGGER update_iso27001_assessments_updated_at
BEFORE UPDATE ON iso27001_assessments
FOR EACH ROW EXECUTE FUNCTION update_iso27001_updated_at_column();

CREATE TRIGGER update_iso27001_control_results_updated_at
BEFORE UPDATE ON iso27001_control_results
FOR EACH ROW EXECUTE FUNCTION update_iso27001_updated_at_column();

CREATE TRIGGER update_iso27001_framework_mapping_updated_at
BEFORE UPDATE ON iso27001_framework_mapping
FOR EACH ROW EXECUTE FUNCTION update_iso27001_updated_at_column();

CREATE TRIGGER update_iso27001_action_plans_updated_at
BEFORE UPDATE ON iso27001_action_plans
FOR EACH ROW EXECUTE FUNCTION update_iso27001_updated_at_column();

CREATE TRIGGER update_iso27001_documents_updated_at
BEFORE UPDATE ON iso27001_documents
FOR EACH ROW EXECUTE FUNCTION update_iso27001_updated_at_column();

-- Inserir controles ISO 27001 principais relacionados à saúde
INSERT INTO iso27001_controls (control_id, section, name, description, healthcare_applicability, implementation_guidance, validation_rules, reference_links, category)
VALUES
    ('A.5.1.1', 'A.5 Information security policies', 'Policies for information security', 'A set of policies for information security shall be defined, approved by management, published and communicated to employees and relevant external parties.', 'Crítico para proteger dados de saúde sensíveis e garantir compliance com regulamentações regionais como HIPAA, GDPR, LGPD.', 'Desenvolva políticas específicas para dados de saúde que considerem requisitos de privacidade, confidencialidade e integridade.', 
    '[{"key": "security_policies.documented", "expected": true}, {"key": "security_policies.approved", "expected": true}, {"key": "security_policies.communicated", "expected": true}, {"key": "security_policies.healthcare_specific", "expected": true}]', 
    '["https://www.iso.org/standard/54534.html", "https://www.healthit.gov/topic/privacy-security-and-hipaa/security-risk-assessment-tool"]',
    'Policies'),

    ('A.6.1.1', 'A.6 Organization of information security', 'Information security roles and responsibilities', 'All information security responsibilities shall be defined and allocated.', 'Essencial para estabelecer papéis claros na proteção de dados de saúde, particularmente para definir responsabilidades de administradores de sistema, médicos e outros profissionais de saúde.', 'Atribua responsabilidades específicas para manipulação de dados sensíveis de saúde, incluindo a designação de um responsável pela segurança da informação na área de saúde.',
    '[{"key": "security_roles.defined", "expected": true}, {"key": "security_roles.allocated", "expected": true}, {"key": "security_roles.healthcare_specific", "expected": true}, {"key": "security_roles.data_protection_officer", "expected": true}]',
    '["https://www.iso.org/standard/54534.html", "https://www.hhs.gov/hipaa/for-professionals/security/guidance/index.html"]',
    'Organization'),

    ('A.8.2.3', 'A.8 Asset management', 'Handling of assets', 'Procedures for handling assets shall be developed and implemented in accordance with the information classification scheme adopted by the organization.', 'Fundamental para garantir que registros médicos, imagens diagnósticas e outros dados de saúde sejam manuseados de acordo com sua classificação de sensibilidade.', 'Implemente procedimentos específicos para manuseio de ativos de informação de saúde, como registros médicos eletrônicos, imagens radiológicas e resultados de exames.',
    '[{"key": "asset_handling.procedures", "expected": true}, {"key": "asset_handling.classification_based", "expected": true}, {"key": "asset_handling.phi_protection", "expected": true}]',
    '["https://www.iso.org/standard/54534.html", "https://www.hhs.gov/hipaa/for-professionals/security/laws-regulations/index.html"]',
    'Asset Management'),

    ('A.9.2.3', 'A.9 Access control', 'Management of privileged access rights', 'The allocation and use of privileged access rights shall be restricted and controlled.', 'Crítico para controlar quem pode acessar dados de saúde sensíveis, especialmente para sistemas que contêm registros médicos completos.', 'Defina níveis de acesso privilegiado distintos para diferentes funções em ambientes de saúde, como equipe clínica, administradores de sistema e pessoal de suporte.',
    '[{"key": "privileged_access.restricted", "expected": true}, {"key": "privileged_access.controlled", "expected": true}, {"key": "privileged_access.audit_trail", "expected": true}, {"key": "privileged_access.healthcare_role_based", "expected": true}]',
    '["https://www.iso.org/standard/54534.html", "https://www.hhs.gov/hipaa/for-professionals/security/guidance/access-control/index.html"]',
    'Access Control'),

    ('A.12.4.1', 'A.12 Operations security', 'Event logging', 'Event logs recording user activities, exceptions, faults and information security events shall be produced, kept and regularly reviewed.', 'Essencial para rastrear acesso a dados de saúde, identificar incidentes e demonstrar compliance com requisitos de auditoria regulatória.', 'Implemente logs detalhados para acesso, modificação e exportação de dados de saúde, garantindo a rastreabilidade completa de todas as operações.',
    '[{"key": "event_logging.enabled", "expected": true}, {"key": "event_logging.user_activities", "expected": true}, {"key": "event_logging.security_events", "expected": true}, {"key": "event_logging.phi_access_recorded", "expected": true}, {"key": "event_logging.retention_compliant", "expected": true}]',
    '["https://www.iso.org/standard/54534.html", "https://www.hhs.gov/hipaa/for-professionals/security/guidance/audit-controls/index.html"]',
    'Operations Security'),

    ('A.14.2.5', 'A.14 System acquisition, development and maintenance', 'Secure system engineering principles', 'Principles for engineering secure systems shall be established, documented, maintained and applied to any information system implementation efforts.', 'Crucial para sistemas de saúde como EHR/EMR, telemedicina e dispositivos médicos conectados que processam dados de pacientes.', 'Adote princípios de engenharia de segurança específicos para sistemas de saúde, incluindo requisitos de interoperabilidade como HL7 FHIR.',
    '[{"key": "secure_engineering.principles_established", "expected": true}, {"key": "secure_engineering.documented", "expected": true}, {"key": "secure_engineering.applied", "expected": true}, {"key": "secure_engineering.healthcare_specific", "expected": true}]',
    '["https://www.iso.org/standard/54534.html", "https://www.healthit.gov/topic/privacy-security-and-hipaa/security-risk-assessment-tool"]',
    'System Development'),

    ('A.18.1.1', 'A.18 Compliance', 'Identification of applicable legislation and contractual requirements', 'All relevant legislative statutory, regulatory, contractual requirements and the organization''s approach to meet these requirements shall be explicitly identified, documented and kept up to date for each information system and the organization.', 'Fundamental para identificar e cumprir requisitos específicos de saúde como HIPAA, GDPR para saúde, LGPD e regulamentações específicas do setor médico.', 'Mantenha um registro atualizado de todas as leis, regulamentos e padrões aplicáveis a sistemas de saúde nas regiões onde a organização opera.',
    '[{"key": "compliance.requirements_identified", "expected": true}, {"key": "compliance.documented", "expected": true}, {"key": "compliance.updated", "expected": true}, {"key": "compliance.healthcare_specific", "expected": true}, {"key": "compliance.regional_requirements", "expected": true}]',
    '["https://www.iso.org/standard/54534.html", "https://www.hhs.gov/hipaa/for-professionals/compliance-enforcement/index.html"]',
    'Compliance')
ON CONFLICT (control_id) DO UPDATE
    SET section = EXCLUDED.section,
        name = EXCLUDED.name,
        description = EXCLUDED.description,
        healthcare_applicability = EXCLUDED.healthcare_applicability,
        implementation_guidance = EXCLUDED.implementation_guidance,
        validation_rules = EXCLUDED.validation_rules,
        reference_links = EXCLUDED.reference_links,
        updated_at = NOW(),
        category = EXCLUDED.category;

-- Inserir mapeamentos para outros frameworks
INSERT INTO iso27001_framework_mapping (iso_control_id, framework_id, framework_control_id, framework_control_name, mapping_type, mapping_strength, notes)
VALUES
    -- Mapeamento para HIPAA
    ((SELECT id FROM iso27001_controls WHERE control_id = 'A.5.1.1'), 
     (SELECT id FROM regulatory_frameworks WHERE code = 'HIPAA'),
     '164.308(a)(1)(i)', 'Security Management Process', 'one_to_one', 'strong', 'HIPAA exige políticas de segurança para proteger dados de saúde'),
    
    ((SELECT id FROM iso27001_controls WHERE control_id = 'A.9.2.3'), 
     (SELECT id FROM regulatory_frameworks WHERE code = 'HIPAA'),
     '164.308(a)(3)(ii)(B)', 'Access Authorization', 'one_to_one', 'strong', 'HIPAA exige controle de acesso privilegiado a dados de saúde'),
    
    ((SELECT id FROM iso27001_controls WHERE control_id = 'A.12.4.1'), 
     (SELECT id FROM regulatory_frameworks WHERE code = 'HIPAA'),
     '164.308(a)(1)(ii)(D)', 'Information System Activity Review', 'one_to_one', 'strong', 'HIPAA exige revisão de atividades de sistemas de informação'),
    
    -- Mapeamento para GDPR
    ((SELECT id FROM iso27001_controls WHERE control_id = 'A.18.1.1'), 
     (SELECT id FROM regulatory_frameworks WHERE code = 'GDPR'),
     'Art. 5', 'Principles relating to processing of personal data', 'one_to_many', 'strong', 'GDPR define princípios para processamento de dados pessoais, incluindo dados de saúde'),
    
    ((SELECT id FROM iso27001_controls WHERE control_id = 'A.8.2.3'), 
     (SELECT id FROM regulatory_frameworks WHERE code = 'GDPR'),
     'Art. 32', 'Security of processing', 'one_to_one', 'moderate', 'GDPR exige segurança no processamento de dados pessoais'),
    
    -- Mapeamento para LGPD
    ((SELECT id FROM iso27001_controls WHERE control_id = 'A.18.1.1'), 
     (SELECT id FROM regulatory_frameworks WHERE code = 'LGPD'),
     'Art. 6', 'Princípios', 'one_to_many', 'strong', 'LGPD define princípios para tratamento de dados pessoais, incluindo dados de saúde'),
    
    ((SELECT id FROM iso27001_controls WHERE control_id = 'A.9.2.3'), 
     (SELECT id FROM regulatory_frameworks WHERE code = 'LGPD'),
     'Art. 46', 'Segurança e sigilo de dados', 'partial', 'moderate', 'LGPD exige medidas de segurança para proteção de dados')
ON CONFLICT DO NOTHING;

COMMENT ON TABLE iso27001_controls IS 'Controles definidos pelo padrão ISO/IEC 27001 para segurança da informação';
COMMENT ON TABLE iso27001_assessments IS 'Avaliações de conformidade com ISO 27001 realizadas pela organização';
COMMENT ON TABLE iso27001_control_results IS 'Resultados da avaliação de cada controle ISO 27001';
COMMENT ON TABLE iso27001_framework_mapping IS 'Mapeamento entre controles ISO 27001 e outros frameworks regulatórios';
COMMENT ON TABLE iso27001_action_plans IS 'Planos de ação para endereçar não conformidades com ISO 27001';
COMMENT ON TABLE iso27001_documents IS 'Documentos relacionados à implementação e manutenção do SGSI conforme ISO 27001';
