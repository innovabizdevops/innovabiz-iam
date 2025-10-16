-- V14__process_domains.sql
-- Modelagem detalhada dos domínios de processos, procedimentos, atividades
-- Normas: BPMN, ITIL, COBIT, ISO 9001, ISO 56000, entre outras

CREATE SCHEMA IF NOT EXISTS processes;

-- Tipos de Processos
CREATE TABLE processes.process_type (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    accessibility VARCHAR(50),
    data_source VARCHAR(100),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE processes.process_type IS 'Tipos de processos organizacionais conforme BPMN, ITIL, COBIT';

-- Processos
CREATE TABLE processes.process (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    description TEXT,
    process_type_id UUID REFERENCES processes.process_type(id),
    status VARCHAR(30) DEFAULT 'active',
    start_date DATE,
    end_date DATE,
    responsible VARCHAR(100),
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    accessibility VARCHAR(50),
    data_source VARCHAR(100),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    external_ref VARCHAR(200)
);
COMMENT ON TABLE processes.process IS 'Processos organizacionais (macroprocessos e subprocessos)';

-- Procedimentos
CREATE TABLE processes.procedure (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    process_id UUID REFERENCES processes.process(id),
    name VARCHAR(150) NOT NULL,
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    responsible VARCHAR(100),
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    accessibility VARCHAR(50),
    data_source VARCHAR(100),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    external_ref VARCHAR(200)
);
COMMENT ON TABLE processes.procedure IS 'Procedimentos vinculados a processos organizacionais';

-- Atividades
CREATE TABLE processes.activity (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    procedure_id UUID REFERENCES processes.procedure(id),
    name VARCHAR(150) NOT NULL,
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    responsible VARCHAR(100),
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    accessibility VARCHAR(50),
    data_source VARCHAR(100),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    external_ref VARCHAR(200)
);
COMMENT ON TABLE processes.activity IS 'Atividades detalhadas dentro dos procedimentos';

-- Tabelas de relacionamento, triggers, views e funções de auditoria, compliance e validação serão adicionados em arquivos específicos para automação e integração BI.
