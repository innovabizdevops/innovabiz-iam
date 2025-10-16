-- V21__automation_teams_agents.sql
-- Implementação de automatismo para geração de equipes e agentes especializados
-- Data: 2025-06-12
-- Autor: Sistema Innovabiz

-- 1. Criar schema para automação se não existir
CREATE SCHEMA IF NOT EXISTS automation;

-- 2. Tabela de configuração de especialidades
CREATE TABLE IF NOT EXISTS automation.specialties (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    framework_category VARCHAR(100),
    domain_scope VARCHAR(100),
    required_skills TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Tabela de configuração de equipes
CREATE TABLE IF NOT EXISTS automation.teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    domain_responsibility VARCHAR(100),
    min_members INTEGER DEFAULT 2,
    max_members INTEGER DEFAULT 8,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 4. Tabela de agentes especializados
CREATE TABLE IF NOT EXISTS automation.specialized_agents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    specialty_id UUID REFERENCES automation.specialties(id),
    agent_type VARCHAR(50) NOT NULL, -- 'human', 'automated', 'hybrid', 'vertical', 'horizontal', 'coordinator', 'manager'
    automation_script TEXT,
    schedule_expression VARCHAR(100), -- cron expression
    notification_endpoint VARCHAR(255),
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    certifications TEXT[],
    expertise_level VARCHAR(50) CHECK (expertise_level IN ('junior', 'pleno', 'senior', 'expert')),
    frameworks TEXT[]
);

-- 5. Tabela para matriz de atributos transversais
CREATE TABLE IF NOT EXISTS automation.transversal_attributes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    category VARCHAR(100) NOT NULL,
    description TEXT,
    weight INTEGER DEFAULT 1,
    domain_applicability TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 6. Tabela para relações entre equipes e agentes
CREATE TABLE IF NOT EXISTS automation.team_agent_relationships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID REFERENCES automation.teams(id),
    agent_id UUID REFERENCES automation.specialized_agents(id),
    role VARCHAR(100) NOT NULL, -- 'member', 'lead', 'coordinator', 'specialist'
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 7. Função para geração automática de equipes por domínio
CREATE OR REPLACE FUNCTION automation.generate_teams_for_domain(domain_name TEXT)
RETURNS VOID AS $$
DECLARE
    domains TEXT[] := ARRAY['compliance', 'governance', 'risk', 'iam', 'products', 'analytics', 'esg', 
                           'hr', 'supply_chain', 'crm', 'cxp', 'insurance', 'marketplace', 'microcredit', 
                           'mobile_money', 'payments', 'erp', 'investment'];
    team_types TEXT[] := ARRAY['Development', 'Governance', 'Compliance', 'Audit', 'Security', 'DataScience',
                             'BizOps', 'DevOps', 'DevSecOps', 'MLOps', 'AIOps', 'DataOps', 'SecOps', 'FinOps'];
    i INTEGER;
BEGIN
    IF domain_name = ANY(domains) OR domain_name LIKE '%_domain' THEN
        FOR i IN 1..array_length(team_types, 1) LOOP
            INSERT INTO automation.teams (name, description, domain_responsibility)
            VALUES (
                domain_name || '_' || team_types[i],
                'Auto-generated team for ' || domain_name || ' ' || team_types[i],
                domain_name
            )
            ON CONFLICT DO NOTHING;
        END LOOP;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- 8. Trigger para geração automática quando novo domínio é criado
CREATE OR REPLACE FUNCTION automation.domain_detection_trigger()
RETURNS TRIGGER AS $$
BEGIN
    -- Detectar criação de novo schema/domínio e gerar equipes automaticamente
    PERFORM automation.generate_teams_for_domain(TG_ARGV[0]);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 9. Função para geração automática de agentes especializados
CREATE OR REPLACE FUNCTION automation.generate_agents_for_specialty(specialty_name TEXT, domain_name TEXT DEFAULT NULL)
RETURNS VOID AS $$
DECLARE
    agent_types TEXT[] := ARRAY['Validator', 'Monitor', 'Auditor', 'Reporter', 'Coordinator', 'Manager', 
                               'Specialist', 'VerticalExpert', 'HorizontalExpert', 'Integrator'];
    expertise_levels TEXT[] := ARRAY['junior', 'pleno', 'senior', 'expert'];
    i INTEGER;
    domain_suffix TEXT;
    expertise TEXT;
BEGIN
    domain_suffix := COALESCE(domain_name, '');
    
    FOR i IN 1..array_length(agent_types, 1) LOOP
        -- Determinar nível de expertise baseado no tipo de agente
        IF agent_types[i] IN ('Coordinator', 'Manager', 'VerticalExpert') THEN
            expertise := 'senior';
        ELSIF agent_types[i] IN ('Specialist', 'Auditor') THEN
            expertise := 'pleno';
        ELSE
            expertise := 'junior';
        END IF;
        
        -- Criar o agente especializado
        INSERT INTO automation.specialized_agents (
            name, 
            specialty_id,
            agent_type,
            automation_script,
            schedule_expression,
            expertise_level,
            frameworks
        )
        SELECT 
            specialty_name || '_' || agent_types[i] || CASE WHEN domain_suffix <> '' THEN '_' || domain_suffix ELSE '' END,
            id,
            CASE 
                WHEN agent_types[i] IN ('Monitor', 'Reporter', 'Validator') THEN 'automated'
                WHEN agent_types[i] IN ('Coordinator', 'Manager') THEN 'vertical'
                WHEN agent_types[i] = 'HorizontalExpert' THEN 'horizontal'
                ELSE 'hybrid'
            END,
            '-- Default script for ' || agent_types[i] || ' in domain ' || COALESCE(domain_suffix, 'global'),
            CASE
                WHEN agent_types[i] = 'Monitor' THEN '*/15 * * * *' -- a cada 15 minutos
                WHEN agent_types[i] = 'Reporter' THEN '0 9 * * 1-5' -- dias úteis às 9h
                WHEN agent_types[i] = 'Validator' THEN '0 */4 * * *' -- a cada 4 horas
                WHEN agent_types[i] = 'Auditor' THEN '0 2 * * *' -- diariamente às 2h
                ELSE NULL
            END,
            expertise,
            ARRAY[framework_category]
        FROM automation.specialties 
        WHERE name = specialty_name
        ON CONFLICT DO NOTHING;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- 10. Função para gerar equipe completa para um domínio
CREATE OR REPLACE FUNCTION automation.create_complete_domain_team(domain_name TEXT)
RETURNS VOID AS $$
DECLARE
    team_id UUID;
    agent_id UUID;
    specialty_record RECORD;
    domain_team_name TEXT;
BEGIN
    -- Primeiro gera as equipes para o domínio
    PERFORM automation.generate_teams_for_domain(domain_name);
    
    -- Para cada especialidade aplicável, gera agentes para o domínio
    FOR specialty_record IN SELECT * FROM automation.specialties LOOP
        PERFORM automation.generate_agents_for_specialty(specialty_record.name, domain_name);
    END LOOP;
    
    -- Configura a relação entre equipes e agentes
    domain_team_name := domain_name || '_Governance';
    
    -- Obter ID da equipe de governança para este domínio
    SELECT id INTO team_id FROM automation.teams WHERE name = domain_team_name;
    
    IF team_id IS NOT NULL THEN
        -- Associar agentes à equipe
        FOR agent_id IN 
            SELECT id FROM automation.specialized_agents 
            WHERE name LIKE '%\_' || domain_name ESCAPE '\' 
            OR name LIKE '%\_' || domain_name || '\_%' ESCAPE '\'
        LOOP
            INSERT INTO automation.team_agent_relationships (team_id, agent_id, role)
            VALUES (
                team_id, 
                agent_id, 
                CASE 
                    WHEN (SELECT agent_type FROM automation.specialized_agents WHERE id = agent_id) = 'vertical' THEN 'lead'
                    WHEN (SELECT agent_type FROM automation.specialized_agents WHERE id = agent_id) = 'horizontal' THEN 'specialist'
                    ELSE 'member'
                END
            )
            ON CONFLICT DO NOTHING;
        END LOOP;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- 11. Dados iniciais para especialidades alinhados com frameworks
INSERT INTO automation.specialties (name, description, framework_category, domain_scope, required_skills)
VALUES 
    ('DataGovernance', 'Data Governance Specialist', 'DMBOK', 'core', ARRAY['SQL', 'Governance', 'Compliance']),
    ('ComplianceAudit', 'Compliance and Audit Specialist', 'ISO27001', 'compliance', ARRAY['Audit', 'Risk', 'Compliance']),
    ('SecurityOps', 'Security Operations Specialist', 'NIST', 'security', ARRAY['Security', 'DevSecOps']),
    ('DataScience', 'Data Science Specialist', 'CRISP-DM', 'analytics', ARRAY['ML', 'Statistics', 'Python']),
    ('DevOpsIntegration', 'DevOps Integration Specialist', 'DevOps', 'infrastructure', ARRAY['CI/CD', 'Automation', 'Scripting']),
    ('AIGovernance', 'AI Governance Specialist', 'OECD-AI', 'ai', ARRAY['Ethics', 'AI', 'Governance']),
    ('KrakendAPIGateway', 'Krakend API Gateway Specialist', 'API-Gateway', 'integration', ARRAY['API', 'Gateway', 'Integration']),
    ('MCPIntegration', 'Model Context Protocol Specialist', 'MCP', 'ai-integration', ARRAY['MCP', 'AI', 'Integration']),
    ('GraphQLIntegration', 'GraphQL Integration Specialist', 'GraphQL', 'api', ARRAY['GraphQL', 'API', 'Query']),
    ('IAMSpecialist', 'Identity & Access Management Specialist', 'IAM', 'security', ARRAY['Identity', 'Access', 'Authentication']),
    ('PostgreSQLAdmin', 'PostgreSQL Database Administrator', 'SQL', 'database', ARRAY['PostgreSQL', 'SQL', 'Administration']),
    ('RedisSpecialist', 'Redis Cache Specialist', 'NoSQL', 'cache', ARRAY['Redis', 'Cache', 'NoSQL']),
    ('KafkaStreaming', 'Kafka Streaming Specialist', 'Streaming', 'events', ARRAY['Kafka', 'Streaming', 'Events']),
    ('SnowflakeSpecialist', 'Snowflake Data Warehouse Specialist', 'DataWarehouse', 'analytics', ARRAY['Snowflake', 'SQL', 'DataOps']),
    ('CosmosDBSpecialist', 'Azure CosmosDB Specialist', 'NoSQL', 'database', ARRAY['CosmosDB', 'Azure', 'NoSQL']),
    ('Neo4jSpecialist', 'Neo4j Graph Database Specialist', 'GraphDB', 'database', ARRAY['Neo4j', 'Graph', 'Cypher']),
    ('StreamlitDashboard', 'Streamlit Dashboard Developer', 'Visualization', 'frontend', ARRAY['Streamlit', 'Python', 'Dashboard']),
    ('SupabaseSpecialist', 'Supabase Backend Specialist', 'BaaS', 'backend', ARRAY['Supabase', 'PostgreSQL', 'API']),
    ('TimescaleDBSpecialist', 'TimescaleDB Time Series Specialist', 'TimeSeries', 'database', ARRAY['TimescaleDB', 'PostgreSQL', 'TimeSeries']),
    ('CorporateGovernance', 'Corporate Governance Specialist', 'OCEG', 'governance', ARRAY['Governance', 'Compliance', 'Risk']),
    ('TOGAFArchitect', 'TOGAF Enterprise Architect', 'TOGAF', 'architecture', ARRAY['TOGAF', 'Architecture', 'Enterprise']),
    ('COBITGovernance', 'COBIT IT Governance Specialist', 'COBIT', 'it-governance', ARRAY['COBIT', 'IT', 'Governance']),
    ('ITILSpecialist', 'ITIL Service Management Specialist', 'ITIL', 'service', ARRAY['ITIL', 'Service', 'Management']),
    ('BABOKAnalyst', 'BABOK Business Analyst', 'BABOK', 'business', ARRAY['BABOK', 'Analysis', 'Business']),
    ('OpenBanking', 'Open Banking Specialist', 'Open-X', 'financial', ARRAY['OpenBanking', 'API', 'Financial']),
    ('MicroCreditSpecialist', 'Microcredit Specialist', 'Financial-Inclusion', 'credit', ARRAY['Microcredit', 'Risk', 'Financial'])
ON CONFLICT DO NOTHING;

-- 12. Dados iniciais para matriz de atributos transversais
INSERT INTO automation.transversal_attributes (name, category, description, weight, domain_applicability)
VALUES
    ('SecurityCompliance', 'Security', 'Security conformity across all domains', 10, ARRAY['*']),
    ('DataPrivacy', 'Privacy', 'GDPR/LGPD compliance for data handling', 10, ARRAY['*']),
    ('AuditTrail', 'Audit', 'Complete audit trail for all operations', 9, ARRAY['*']),
    ('AccessControl', 'Security', 'Role-based access control', 9, ARRAY['*']),
    ('AIEthics', 'Ethics', 'Ethical AI principles application', 8, ARRAY['ai', 'analytics']),
    ('Interoperability', 'Integration', 'System interoperability standards', 7, ARRAY['*']),
    ('Scalability', 'Architecture', 'System scalability design principles', 6, ARRAY['*']),
    ('UserExperience', 'UX', 'Consistent user experience standards', 7, ARRAY['frontend', 'mobile']),
    ('PerformanceOptimization', 'Performance', 'Performance standards and monitoring', 6, ARRAY['*']),
    ('DocumentationStandards', 'Documentation', 'Comprehensive documentation templates', 5, ARRAY['*']),
    ('DevOpsIntegration', 'DevOps', 'CI/CD pipeline integration', 8, ARRAY['development', 'operations']),
    ('MultilingualSupport', 'Localization', 'Support for multiple languages', 4, ARRAY['frontend', 'content'])
ON CONFLICT DO NOTHING;

-- 13. Iniciar geração de equipes para domínios existentes
DO $$
DECLARE
    domains TEXT[] := ARRAY['iam', 'compliance', 'analytics', 'monitoring', 'governance', 'risk', 
                           'processes', 'contracts', 'kpis', 'geographies', 'business',
                           'organization', 'products_services', 'reference', 'esg', 'supply_chain',
                           'hr', 'crm', 'cxp'];
    domain_name TEXT;
BEGIN
    FOREACH domain_name IN ARRAY domains LOOP
        PERFORM automation.create_complete_domain_team(domain_name);
    END LOOP;
END $$;
