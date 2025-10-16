-- ==========================================================================
-- Nome: V20__iam_multi_tenant_core_schema.sql
-- Descrição: Script de migração para estrutura base multi-tenant do IAM
-- Autor: Equipa de Desenvolvimento INNOVABIZ
-- Data: 19/08/2025
-- ==========================================================================

-- Habilitar extensões necessárias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "hstore";

-- ==========================================================================
-- ESQUEMA DE TENANTS
-- ==========================================================================

CREATE SCHEMA IF NOT EXISTS tenant_management;

-- Tabela de Tenants (Organizações)
CREATE TABLE IF NOT EXISTS tenant_management.tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(255) NOT NULL,
    nome_comercial VARCHAR(255),
    slug VARCHAR(100) NOT NULL UNIQUE,
    codigo_pais CHAR(2) NOT NULL,
    tipo_tenant VARCHAR(50) NOT NULL,
    segmento VARCHAR(100) NOT NULL,
    industria VARCHAR(100) NOT NULL,
    data_fundacao DATE,
    status VARCHAR(30) NOT NULL DEFAULT 'ativo',
    nivel_servico VARCHAR(30) NOT NULL DEFAULT 'standard',
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_atualizacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_expiracao TIMESTAMP WITH TIME ZONE,
    qtd_max_usuarios INTEGER DEFAULT 10,
    config_autenticacao JSONB NOT NULL DEFAULT '{"password": {"min_length": 10, "require_uppercase": true, "require_lowercase": true, "require_number": true, "require_special": true, "expiry_days": 90, "history_count": 5}, "mfa": {"required": false, "methods": ["email", "sms", "totp"]}, "session": {"timeout_minutes": 30, "max_duration_hours": 12, "max_concurrent_sessions": 5}, "lockout": {"max_attempts": 5, "duration_minutes": 30}}',
    config_autorizacao JSONB NOT NULL DEFAULT '{"role_hierarchy": true, "default_role": "user", "admin_role": "admin", "permission_inheritance": true}',
    metadados JSONB,
    criado_por UUID,
    atualizado_por UUID,
    CONSTRAINT ck_tenant_status CHECK (status IN ('ativo', 'inativo', 'suspenso', 'bloqueado', 'pendente', 'expirado')),
    CONSTRAINT ck_tenant_nivel CHECK (nivel_servico IN ('gratuito', 'standard', 'premium', 'enterprise', 'personalizado')),
    CONSTRAINT ck_tenant_tipo CHECK (tipo_tenant IN ('governo', 'financeiro', 'saude', 'educacao', 'comercio', 'telecomunicacoes', 'transporte', 'energia', 'manufactura', 'servicos', 'outro'))
);

-- Tabela de Domínios de Tenant
CREATE TABLE IF NOT EXISTS tenant_management.tenant_dominios (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    dominio VARCHAR(255) NOT NULL,
    validado BOOLEAN NOT NULL DEFAULT false,
    primario BOOLEAN NOT NULL DEFAULT false,
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_atualizacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_tenant_dominios_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_management.tenants(id) ON DELETE CASCADE,
    CONSTRAINT uk_tenant_dominios_dominio UNIQUE (dominio)
);

-- Tabela de Configuração por Tenant e Região
CREATE TABLE IF NOT EXISTS tenant_management.tenant_configuracao_regional (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    codigo_pais CHAR(2) NOT NULL,
    regulamentos_aplicaveis JSONB NOT NULL,
    politicas_especificas JSONB,
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_atualizacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    criado_por UUID,
    atualizado_por UUID,
    CONSTRAINT fk_tenant_config_regional_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_management.tenants(id) ON DELETE CASCADE,
    CONSTRAINT uk_tenant_config_regional UNIQUE (tenant_id, codigo_pais)
);

-- ==========================================================================
-- ESQUEMA IAM CORE
-- ==========================================================================

CREATE SCHEMA IF NOT EXISTS iam_core;

-- Tabela de Usuários
CREATE TABLE IF NOT EXISTS iam_core.usuarios (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    nome_completo VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    nome_usuario VARCHAR(100),
    senha_hash VARCHAR(255),
    documento_principal VARCHAR(50),
    tipo_documento VARCHAR(30),
    celular VARCHAR(30),
    data_nascimento DATE,
    genero VARCHAR(15),
    foto_url VARCHAR(500),
    status VARCHAR(30) NOT NULL DEFAULT 'ativo',
    verificado BOOLEAN NOT NULL DEFAULT false,
    ultimo_login TIMESTAMP WITH TIME ZONE,
    falhas_login_consecutivas INTEGER NOT NULL DEFAULT 0,
    data_bloqueio TIMESTAMP WITH TIME ZONE,
    data_expiracao_senha TIMESTAMP WITH TIME ZONE,
    preferencias JSONB NOT NULL DEFAULT '{"idioma": "pt", "tema": "claro", "notificacoes": {"email": true, "sms": true, "push": true}, "timezone": "UTC"}',
    metadados JSONB,
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_atualizacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    criado_por UUID,
    atualizado_por UUID,
    CONSTRAINT fk_usuarios_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_management.tenants(id),
    CONSTRAINT uk_usuarios_tenant_email UNIQUE (tenant_id, email),
    CONSTRAINT ck_usuarios_status CHECK (status IN ('ativo', 'inativo', 'suspenso', 'bloqueado', 'pendente', 'excluido')),
    CONSTRAINT ck_usuarios_genero CHECK (genero IS NULL OR genero IN ('masculino', 'feminino', 'outro', 'prefiro_nao_informar'))
);

-- Índice para pesquisa rápida por email
CREATE INDEX idx_usuarios_email ON iam_core.usuarios(email);
-- Índice para pesquisa por tenant
CREATE INDEX idx_usuarios_tenant ON iam_core.usuarios(tenant_id);

-- Tabela de Dispositivos
CREATE TABLE IF NOT EXISTS iam_core.dispositivos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    usuario_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    nome VARCHAR(255) NOT NULL,
    identificador VARCHAR(255) NOT NULL,
    tipo_dispositivo VARCHAR(50) NOT NULL,
    sistema_operacional VARCHAR(100),
    navegador VARCHAR(100),
    confiavel BOOLEAN NOT NULL DEFAULT false,
    data_ultimo_acesso TIMESTAMP WITH TIME ZONE,
    metadados JSONB,
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_atualizacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_dispositivos_usuario FOREIGN KEY (usuario_id) REFERENCES iam_core.usuarios(id) ON DELETE CASCADE,
    CONSTRAINT fk_dispositivos_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_management.tenants(id),
    CONSTRAINT uk_dispositivos_identificador UNIQUE (usuario_id, identificador)
);

-- Tabela de Perfis/Funções
CREATE TABLE IF NOT EXISTS iam_core.perfis (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    nome VARCHAR(100) NOT NULL,
    descricao VARCHAR(500),
    interno BOOLEAN NOT NULL DEFAULT false,
    nivel_hierarquia INTEGER NOT NULL DEFAULT 0,
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_atualizacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    criado_por UUID,
    atualizado_por UUID,
    CONSTRAINT fk_perfis_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_management.tenants(id),
    CONSTRAINT uk_perfis_tenant_nome UNIQUE (tenant_id, nome)
);

-- Tabela de Permissões
CREATE TABLE IF NOT EXISTS iam_core.permissoes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    codigo VARCHAR(100) NOT NULL,
    nome VARCHAR(100) NOT NULL,
    descricao VARCHAR(500),
    modulo VARCHAR(100) NOT NULL,
    acao VARCHAR(50) NOT NULL,
    recurso VARCHAR(100) NOT NULL,
    interno BOOLEAN NOT NULL DEFAULT false,
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_atualizacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    criado_por UUID,
    atualizado_por UUID,
    CONSTRAINT fk_permissoes_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_management.tenants(id),
    CONSTRAINT uk_permissoes_tenant_codigo UNIQUE (tenant_id, codigo),
    CONSTRAINT ck_permissoes_acao CHECK (acao IN ('criar', 'ler', 'atualizar', 'excluir', 'executar', 'aprovar', 'rejeitar', 'todas'))
);

-- Tabela de Relação Perfil-Permissão
CREATE TABLE IF NOT EXISTS iam_core.perfil_permissoes (
    perfil_id UUID NOT NULL,
    permissao_id UUID NOT NULL,
    concedido_por UUID,
    data_concessao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (perfil_id, permissao_id),
    CONSTRAINT fk_perfil_permissao_perfil FOREIGN KEY (perfil_id) REFERENCES iam_core.perfis(id) ON DELETE CASCADE,
    CONSTRAINT fk_perfil_permissao_permissao FOREIGN KEY (permissao_id) REFERENCES iam_core.permissoes(id) ON DELETE CASCADE
);

-- Tabela de Relação Usuário-Perfil
CREATE TABLE IF NOT EXISTS iam_core.usuario_perfis (
    usuario_id UUID NOT NULL,
    perfil_id UUID NOT NULL,
    concedido_por UUID,
    data_concessao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_expiracao TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (usuario_id, perfil_id),
    CONSTRAINT fk_usuario_perfil_usuario FOREIGN KEY (usuario_id) REFERENCES iam_core.usuarios(id) ON DELETE CASCADE,
    CONSTRAINT fk_usuario_perfil_perfil FOREIGN KEY (perfil_id) REFERENCES iam_core.perfis(id) ON DELETE CASCADE
);

-- Tabela de Métodos de Autenticação MFA
CREATE TABLE IF NOT EXISTS iam_core.metodos_autenticacao (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    usuario_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    tipo_metodo VARCHAR(50) NOT NULL,
    identificador VARCHAR(255) NOT NULL,
    segredo_hash VARCHAR(255),
    verificado BOOLEAN NOT NULL DEFAULT false,
    configuracao JSONB,
    data_ultimo_uso TIMESTAMP WITH TIME ZONE,
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_atualizacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_metodos_autenticacao_usuario FOREIGN KEY (usuario_id) REFERENCES iam_core.usuarios(id) ON DELETE CASCADE,
    CONSTRAINT fk_metodos_autenticacao_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_management.tenants(id),
    CONSTRAINT uk_metodos_autenticacao_tipo_identificador UNIQUE (usuario_id, tipo_metodo, identificador),
    CONSTRAINT ck_metodos_autenticacao_tipo CHECK (tipo_metodo IN ('email', 'sms', 'totp', 'fido2', 'biometria_facial', 'biometria_digital', 'push', 'qrcode', 'hardware_token'))
);

-- Tabela de Sessões
CREATE TABLE IF NOT EXISTS iam_core.sessoes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    usuario_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    refresh_token_hash VARCHAR(255),
    dispositivo_id UUID,
    ip_origem VARCHAR(50) NOT NULL,
    localizacao VARCHAR(255),
    user_agent VARCHAR(500),
    metadados JSONB,
    data_inicio TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_expiracao TIMESTAMP WITH TIME ZONE NOT NULL,
    data_ultimo_acesso TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ativa BOOLEAN NOT NULL DEFAULT true,
    CONSTRAINT fk_sessoes_usuario FOREIGN KEY (usuario_id) REFERENCES iam_core.usuarios(id) ON DELETE CASCADE,
    CONSTRAINT fk_sessoes_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_management.tenants(id),
    CONSTRAINT fk_sessoes_dispositivo FOREIGN KEY (dispositivo_id) REFERENCES iam_core.dispositivos(id) ON DELETE SET NULL
);

-- Índice para busca rápida por sessões ativas
CREATE INDEX idx_sessoes_ativas ON iam_core.sessoes(usuario_id, ativa) WHERE ativa = true;

-- ==========================================================================
-- ESQUEMA DE AUDITORIA
-- ==========================================================================

CREATE SCHEMA IF NOT EXISTS iam_audit;

-- Tabela de Eventos de Auditoria
CREATE TABLE IF NOT EXISTS iam_audit.eventos_auditoria (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    usuario_id UUID,
    sessao_id UUID,
    tipo_evento VARCHAR(100) NOT NULL,
    categoria VARCHAR(100) NOT NULL,
    severidade VARCHAR(20) NOT NULL,
    descricao TEXT NOT NULL,
    detalhes JSONB,
    resultado VARCHAR(50) NOT NULL,
    ip_origem VARCHAR(50),
    user_agent VARCHAR(500),
    modulo VARCHAR(100) NOT NULL,
    recurso_afetado VARCHAR(255),
    id_recurso VARCHAR(255),
    data_ocorrencia TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_eventos_auditoria_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_management.tenants(id),
    CONSTRAINT ck_eventos_auditoria_severidade CHECK (severidade IN ('informativa', 'baixa', 'media', 'alta', 'critica')),
    CONSTRAINT ck_eventos_auditoria_resultado CHECK (resultado IN ('sucesso', 'falha', 'negado', 'erro', 'aviso'))
);

-- Particionamento por data (mensal) para tabela de auditoria
CREATE TABLE IF NOT EXISTS iam_audit.eventos_auditoria_particionada (
    LIKE iam_audit.eventos_auditoria INCLUDING ALL
) PARTITION BY RANGE (data_ocorrencia);

-- Criar partição inicial (ajustar conforme necessário)
CREATE TABLE IF NOT EXISTS iam_audit.eventos_auditoria_y2025m08 
    PARTITION OF iam_audit.eventos_auditoria_particionada 
    FOR VALUES FROM ('2025-08-01') TO ('2025-09-01');

CREATE TABLE IF NOT EXISTS iam_audit.eventos_auditoria_y2025m09 
    PARTITION OF iam_audit.eventos_auditoria_particionada 
    FOR VALUES FROM ('2025-09-01') TO ('2025-10-01');

-- Índices para otimizar consultas de auditoria
CREATE INDEX idx_eventos_auditoria_tenant ON iam_audit.eventos_auditoria(tenant_id);
CREATE INDEX idx_eventos_auditoria_usuario ON iam_audit.eventos_auditoria(usuario_id);
CREATE INDEX idx_eventos_auditoria_data ON iam_audit.eventos_auditoria(data_ocorrencia);
CREATE INDEX idx_eventos_auditoria_tipo ON iam_audit.eventos_auditoria(tipo_evento);

-- ==========================================================================
-- ESQUEMA DE RISCO E CONFORMIDADE
-- ==========================================================================

CREATE SCHEMA IF NOT EXISTS iam_risk;

-- Tabela de Perfis de Risco dos Usuários
CREATE TABLE IF NOT EXISTS iam_risk.perfis_risco_usuario (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    usuario_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    nivel_risco VARCHAR(20) NOT NULL DEFAULT 'medio',
    score_risco INTEGER NOT NULL DEFAULT 50,
    fatores_risco JSONB,
    ultima_avaliacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    proxima_avaliacao TIMESTAMP WITH TIME ZONE,
    metadados JSONB,
    CONSTRAINT fk_perfis_risco_usuario_usuario FOREIGN KEY (usuario_id) REFERENCES iam_core.usuarios(id) ON DELETE CASCADE,
    CONSTRAINT fk_perfis_risco_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_management.tenants(id),
    CONSTRAINT ck_perfis_risco_nivel CHECK (nivel_risco IN ('muito_baixo', 'baixo', 'medio', 'alto', 'muito_alto', 'critico')),
    CONSTRAINT ck_perfis_risco_score CHECK (score_risco BETWEEN 0 AND 100)
);

-- Tabela de Eventos de Risco
CREATE TABLE IF NOT EXISTS iam_risk.eventos_risco (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    usuario_id UUID,
    sessao_id UUID,
    tipo_evento VARCHAR(100) NOT NULL,
    severidade VARCHAR(20) NOT NULL,
    descricao TEXT NOT NULL,
    detalhes JSONB,
    metadados_contextuais JSONB,
    ip_origem VARCHAR(50),
    localizacao VARCHAR(255),
    dispositivo_info JSONB,
    data_ocorrencia TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    avaliado BOOLEAN NOT NULL DEFAULT false,
    resultado_avaliacao VARCHAR(50),
    data_avaliacao TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_eventos_risco_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_management.tenants(id),
    CONSTRAINT ck_eventos_risco_severidade CHECK (severidade IN ('informativa', 'baixa', 'media', 'alta', 'critica')),
    CONSTRAINT ck_eventos_risco_resultado CHECK (resultado_avaliacao IS NULL OR resultado_avaliacao IN ('legitimo', 'suspeito', 'fraudulento', 'inconcluso'))
);

-- Tabela de Políticas de Conformidade
CREATE TABLE IF NOT EXISTS iam_risk.politicas_conformidade (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    nome VARCHAR(255) NOT NULL,
    descricao TEXT,
    tipo_regulamento VARCHAR(100) NOT NULL,
    jurisdicao VARCHAR(100) NOT NULL,
    regras JSONB NOT NULL,
    ativa BOOLEAN NOT NULL DEFAULT true,
    versao VARCHAR(20) NOT NULL DEFAULT '1.0',
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_atualizacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_desativacao TIMESTAMP WITH TIME ZONE,
    criado_por UUID,
    atualizado_por UUID,
    CONSTRAINT fk_politicas_conformidade_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_management.tenants(id),
    CONSTRAINT uk_politicas_conformidade_nome_tenant UNIQUE (tenant_id, nome, versao),
    CONSTRAINT ck_politicas_conformidade_tipo CHECK (tipo_regulamento IN ('gdpr', 'lgpd', 'pci_dss', 'hipaa', 'bna', 'popia', 'ccpa', 'iso27001', 'sox', 'personalizado'))
);

-- ==========================================================================
-- FUNÇÕES E TRIGGERS
-- ==========================================================================

-- Função para atualizar o timestamp de data_atualizacao automaticamente
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.data_atualizacao = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger para tenant
CREATE TRIGGER update_tenant_timestamp
BEFORE UPDATE ON tenant_management.tenants
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- Trigger para usuario
CREATE TRIGGER update_usuario_timestamp
BEFORE UPDATE ON iam_core.usuarios
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- Função para registrar eventos de auditoria
CREATE OR REPLACE FUNCTION registrar_evento_auditoria(
    p_tenant_id UUID,
    p_usuario_id UUID,
    p_sessao_id UUID,
    p_tipo_evento VARCHAR(100),
    p_categoria VARCHAR(100),
    p_severidade VARCHAR(20),
    p_descricao TEXT,
    p_detalhes JSONB,
    p_resultado VARCHAR(50),
    p_ip_origem VARCHAR(50),
    p_user_agent VARCHAR(500),
    p_modulo VARCHAR(100),
    p_recurso_afetado VARCHAR(255),
    p_id_recurso VARCHAR(255)
) RETURNS UUID AS $$
DECLARE
    evento_id UUID;
BEGIN
    INSERT INTO iam_audit.eventos_auditoria(
        tenant_id, usuario_id, sessao_id, tipo_evento, categoria,
        severidade, descricao, detalhes, resultado, ip_origem,
        user_agent, modulo, recurso_afetado, id_recurso
    ) VALUES (
        p_tenant_id, p_usuario_id, p_sessao_id, p_tipo_evento, p_categoria,
        p_severidade, p_descricao, p_detalhes, p_resultado, p_ip_origem,
        p_user_agent, p_modulo, p_recurso_afetado, p_id_recurso
    )
    RETURNING id INTO evento_id;
    
    RETURN evento_id;
END;
$$ LANGUAGE plpgsql;

-- ==========================================================================
-- DADOS INICIAIS
-- ==========================================================================

-- Inserir tenant padrão para sistema
INSERT INTO tenant_management.tenants (
    id, nome, nome_comercial, slug, codigo_pais, tipo_tenant, segmento, industria, status
) VALUES (
    '00000000-0000-0000-0000-000000000001',
    'INNOVABIZ Sistema',
    'INNOVABIZ',
    'innovabiz-system',
    'AO',
    'financeiro',
    'Financeiro',
    'Tecnologia Financeira',
    'ativo'
)
ON CONFLICT (id) DO NOTHING;

-- Inserir perfis padrão para tenant do sistema
INSERT INTO iam_core.perfis (
    id, tenant_id, nome, descricao, interno, nivel_hierarquia
) VALUES 
    ('00000000-0000-0000-0000-000000000010', '00000000-0000-0000-0000-000000000001', 'admin', 'Administrador do Sistema', true, 100),
    ('00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001', 'user', 'Usuário Padrão', true, 10),
    ('00000000-0000-0000-0000-000000000012', '00000000-0000-0000-0000-000000000001', 'guest', 'Usuário Convidado', true, 1)
ON CONFLICT (tenant_id, nome) DO NOTHING;

-- Comentário final
COMMENT ON SCHEMA tenant_management IS 'Esquema para gerenciamento de tenants (multi-tenant)';
COMMENT ON SCHEMA iam_core IS 'Esquema principal do IAM com tabelas de usuários, autenticação e autorização';
COMMENT ON SCHEMA iam_audit IS 'Esquema para registro de auditoria das operações do IAM';
COMMENT ON SCHEMA iam_risk IS 'Esquema para gestão de risco e conformidade';