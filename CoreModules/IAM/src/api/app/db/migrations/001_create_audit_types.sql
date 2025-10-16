-- INNOVABIZ IAM - Migração 001: Criação de tipos para auditoria
-- Autor: Eduardo Jeremias
-- Versão: 1.0.0
-- Descrição: Cria os tipos enumerados necessários para o sistema de auditoria multi-contexto

-- Criação de tipos enumerados para categorias de eventos de auditoria
CREATE TYPE audit_event_category AS ENUM (
    'AUTHENTICATION',
    'AUTHORIZATION',
    'USER_MANAGEMENT',
    'SYSTEM',
    'DATA_ACCESS',
    'CONFIGURATION',
    'SECURITY',
    'FINANCIAL',
    'CONSENT',
    'API_ACCESS',
    'RESOURCE_MANAGEMENT',
    'EXTERNAL_INTEGRATION'
);

-- Criação de tipos enumerados para severidade de eventos
CREATE TYPE audit_event_severity AS ENUM (
    'DEBUG',
    'INFO',
    'WARNING',
    'ERROR',
    'CRITICAL'
);

-- Criação de tipos enumerados para frameworks de compliance
CREATE TYPE compliance_framework AS ENUM (
    'LGPD',      -- Lei Geral de Proteção de Dados (Brasil)
    'GDPR',      -- General Data Protection Regulation (Europa)
    'SOX',       -- Sarbanes-Oxley Act (EUA)
    'PCI_DSS',   -- Payment Card Industry Data Security Standard (Global)
    'HIPAA',     -- Health Insurance Portability and Accountability Act (EUA)
    'ISO_27001', -- ISO/IEC 27001 Information Security Management (Global)
    'PSD2',      -- Payment Services Directive 2 (Europa)
    'BACEN',     -- Banco Central do Brasil (Brasil)
    'BNA',       -- Banco Nacional de Angola (Angola)
    'NIST',      -- National Institute of Standards and Technology (EUA)
    'CCPA',      -- California Consumer Privacy Act (EUA/California)
    'PIPEDA'     -- Personal Information Protection and Electronic Documents Act (Canadá)
);

-- Criação de tipos enumerados para status de relatório
CREATE TYPE report_status AS ENUM (
    'PENDING',
    'PROCESSING',
    'COMPLETED',
    'FAILED',
    'EXPORTED'
);

-- Comentários para documentação
COMMENT ON TYPE audit_event_category IS 'Categorias de eventos de auditoria para classificação e filtragem';
COMMENT ON TYPE audit_event_severity IS 'Níveis de severidade para eventos de auditoria';
COMMENT ON TYPE compliance_framework IS 'Frameworks regulatórios e de compliance suportados';
COMMENT ON TYPE report_status IS 'Status possíveis para relatórios de compliance';