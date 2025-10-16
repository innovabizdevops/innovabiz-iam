-- V7__compliance_governance_risk.sql
-- Estrutura detalhada para Compliance, Governança e Risco seguindo normas e melhores práticas globais

CREATE SCHEMA IF NOT EXISTS compliance;
CREATE SCHEMA IF NOT EXISTS governance;
CREATE SCHEMA IF NOT EXISTS risk;

-- Normas e Regulamentos
CREATE TABLE IF NOT EXISTS compliance.norma (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    codigo VARCHAR(50) NOT NULL,
    nome VARCHAR(255) NOT NULL,
    descricao TEXT,
    tipo_norma VARCHAR(50),
    orgao_emissor VARCHAR(100),
    data_publicacao DATE,
    status VARCHAR(50) DEFAULT 'vigente',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Frameworks e Metodologias
CREATE TABLE IF NOT EXISTS compliance.framework (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(255) NOT NULL,
    descricao TEXT,
    categoria VARCHAR(50),
    status VARCHAR(50) DEFAULT 'ativo',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Relacionamento N:N entre Normas e Frameworks
CREATE TABLE IF NOT EXISTS compliance.norma_framework (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    norma_id UUID REFERENCES compliance.norma(id),
    framework_id UUID REFERENCES compliance.framework(id)
);

-- Consentimento de Usuários
CREATE TABLE IF NOT EXISTS compliance.consentimento (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    usuario_id UUID NOT NULL,
    tipo_consentimento VARCHAR(100) NOT NULL,
    dado_referente VARCHAR(100),
    consentido BOOLEAN DEFAULT TRUE,
    data_consentimento TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    revogado_em TIMESTAMP WITH TIME ZONE
);

-- Logs de Auditoria
CREATE TABLE IF NOT EXISTS compliance.audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entidade VARCHAR(100) NOT NULL,
    entidade_id UUID,
    acao VARCHAR(50) NOT NULL,
    usuario_id UUID,
    data_acao TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    valor_anterior JSONB,
    valor_novo JSONB,
    origem VARCHAR(100)
);

-- Tipos de Órgão Social
CREATE TABLE IF NOT EXISTS governance.tipo_orgao_social (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL,
    descricao TEXT
);

-- Órgãos Sociais
CREATE TABLE IF NOT EXISTS governance.orgao_social (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL,
    tipo_orgao_social_id UUID REFERENCES governance.tipo_orgao_social(id)
);

-- Papéis e Responsabilidades
CREATE TABLE IF NOT EXISTS governance.papel (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL,
    descricao TEXT
);

CREATE TABLE IF NOT EXISTS governance.responsabilidade (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    papel_id UUID REFERENCES governance.papel(id),
    entidade VARCHAR(100),
    entidade_id UUID
);

-- Tipos de Risco
CREATE TABLE IF NOT EXISTS risk.tipo_risco (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(100) NOT NULL,
    categoria VARCHAR(100)
);

-- Riscos
CREATE TABLE IF NOT EXISTS risk.risco (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(255) NOT NULL,
    tipo_risco_id UUID REFERENCES risk.tipo_risco(id),
    impacto VARCHAR(100),
    probabilidade VARCHAR(100),
    status VARCHAR(50) DEFAULT 'ativo'
);

-- Planos de Mitigação
CREATE TABLE IF NOT EXISTS risk.plano_mitigacao (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    risco_id UUID REFERENCES risk.risco(id),
    descricao TEXT,
    responsavel_id UUID,
    prazo DATE,
    status VARCHAR(50) DEFAULT 'em andamento'
);
