-- V18__hr_domains.sql
-- Modelagem detalhada do domínio Recursos Humanos (RH)
-- Padrões: ISO 30414, ISO 9001, People Analytics, GDPR

CREATE SCHEMA IF NOT EXISTS hr;

-- Colaboradores
CREATE TABLE hr.employee (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    document_id VARCHAR(50),
    birth_date DATE,
    hire_date DATE,
    termination_date DATE,
    department_id UUID REFERENCES organization.department(id),
    position_id UUID REFERENCES hr.position(id),
    status VARCHAR(30) DEFAULT 'active',
    compliance_status VARCHAR(20),
    privacy_level VARCHAR(50),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    external_ref VARCHAR(200)
);
COMMENT ON TABLE hr.employee IS 'Colaboradores da organização conforme ISO 30414 e People Analytics';

-- Cargos
CREATE TABLE hr.position (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    external_ref VARCHAR(200)
);
COMMENT ON TABLE hr.position IS 'Cargos/funções na organização';

-- Competências
CREATE TABLE hr.competency (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    external_ref VARCHAR(200)
);
COMMENT ON TABLE hr.competency IS 'Competências e habilidades dos colaboradores';

-- Matrizes de competências por cargo
CREATE TABLE hr.position_competency (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    position_id UUID REFERENCES hr.position(id),
    competency_id UUID REFERENCES hr.competency(id),
    required_level VARCHAR(50),
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    external_ref VARCHAR(200)
);
COMMENT ON TABLE hr.position_competency IS 'Matriz de competências exigidas por cargo';

-- Indicadores de RH (ex: turnover, absenteísmo, engajamento)
CREATE TABLE hr.hr_indicator (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    unit VARCHAR(30),
    category VARCHAR(50),
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    compliance_status VARCHAR(20),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE hr.hr_indicator IS 'Indicadores de RH para People Analytics e relatórios';

-- Valores reportados por colaborador/ano
CREATE TABLE hr.employee_hr_report (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID REFERENCES hr.employee(id),
    year INT NOT NULL,
    indicator_id UUID REFERENCES hr.hr_indicator(id),
    value NUMERIC,
    unit VARCHAR(30),
    source VARCHAR(100),
    status VARCHAR(30) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    compliance_status VARCHAR(20),
    external_ref VARCHAR(200)
);
COMMENT ON TABLE hr.employee_hr_report IS 'Valores de indicadores de RH reportados por colaborador e ano';

-- Triggers, views e funções de auditoria/compliance seguem o padrão dos demais domínios.
