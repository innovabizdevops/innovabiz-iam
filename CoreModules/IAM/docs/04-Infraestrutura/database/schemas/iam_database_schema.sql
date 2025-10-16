/*
 * Esquema de Banco de Dados do Módulo IAM - INNOVABIZ
 *
 * Este script cria a estrutura de banco de dados para o módulo IAM
 * com suporte aos 70 métodos de autenticação, adaptações regionais,
 * multi-tenancy e demais requisitos do framework.
 *
 * @autor: INNOVABIZ
 * @copyright: 2025 INNOVABIZ
 * @versão: 1.0.0
 * @data: 2025-05-10
 */

-- Criação do esquema para o módulo IAM
CREATE SCHEMA IF NOT EXISTS iam;

COMMENT ON SCHEMA iam IS 'Esquema para o módulo de Identity and Access Management (IAM)';

-- Tabela de tenants (multi-tenancy)
CREATE TABLE iam.tenants (
    tenant_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code VARCHAR(50) UNIQUE NOT NULL,
    nome VARCHAR(200) NOT NULL,
    descricao TEXT,
    dominio VARCHAR(255),
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_modificacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'ativo',
    regiao VARCHAR(10) NOT NULL DEFAULT 'EU', -- EU, BR, AO, US
    configuracoes JSONB DEFAULT '{}'::jsonb,
    plano VARCHAR(50) DEFAULT 'standard',
    limites JSONB DEFAULT '{}'::jsonb,
    metadados JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT ck_tenant_status CHECK (status IN ('ativo', 'inativo', 'bloqueado', 'trial'))
);

COMMENT ON TABLE iam.tenants IS 'Registro de tenants (clientes) da plataforma INNOVABIZ';

-- Tabela de aplicações por tenant
CREATE TABLE iam.aplicacoes (
    aplicacao_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(tenant_id),
    nome VARCHAR(200) NOT NULL,
    descricao TEXT,
    app_tipo VARCHAR(50) NOT NULL DEFAULT 'web', -- web, mobile, desktop, api
    cliente_id VARCHAR(100) UNIQUE NOT NULL,
    cliente_secret TEXT NOT NULL,
    redirect_uris TEXT[],
    allowed_origins TEXT[],
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_modificacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'ativo',
    configuracoes JSONB DEFAULT '{}'::jsonb,
    metadados JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT ck_aplicacao_status CHECK (status IN ('ativo', 'inativo', 'bloqueado', 'desenvolvimento'))
);

CREATE INDEX idx_aplicacoes_tenant ON iam.aplicacoes(tenant_id);
COMMENT ON TABLE iam.aplicacoes IS 'Registro de aplicações que utilizam o sistema de autenticação';

-- Tabela de usuários
CREATE TABLE iam.usuarios (
    usuario_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(tenant_id),
    nome_usuario VARCHAR(100) NOT NULL,
    email VARCHAR(255),
    telefone VARCHAR(50),
    senha_hash TEXT,
    nome_completo VARCHAR(200),
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_modificacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_ultimo_login TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'ativo',
    dados_verificados BOOLEAN DEFAULT false,
    email_verificado BOOLEAN DEFAULT false,
    telefone_verificado BOOLEAN DEFAULT false,
    mfa_obrigatorio BOOLEAN DEFAULT false,
    tentativas_falhas INTEGER DEFAULT 0,
    data_bloqueio TIMESTAMP WITH TIME ZONE,
    data_ultima_senha TIMESTAMP WITH TIME ZONE,
    preferencias JSONB DEFAULT '{}'::jsonb,
    dados_perfil JSONB DEFAULT '{}'::jsonb,
    metadados JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT ck_usuario_status CHECK (status IN ('ativo', 'inativo', 'bloqueado', 'pendente', 'excluido')),
    CONSTRAINT uq_usuario_tenant_username UNIQUE (tenant_id, nome_usuario),
    CONSTRAINT uq_usuario_tenant_email UNIQUE (tenant_id, email)
);

CREATE INDEX idx_usuarios_tenant ON iam.usuarios(tenant_id);
CREATE INDEX idx_usuarios_email ON iam.usuarios(email) WHERE email IS NOT NULL;
CREATE INDEX idx_usuarios_telefone ON iam.usuarios(telefone) WHERE telefone IS NOT NULL;

COMMENT ON TABLE iam.usuarios IS 'Registro de usuários da plataforma';

-- Tabela de histórico de senhas
CREATE TABLE iam.historico_senhas (
    historico_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL REFERENCES iam.usuarios(usuario_id),
    senha_hash TEXT NOT NULL,
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    criado_por VARCHAR(100) NOT NULL DEFAULT CURRENT_USER,
    metadados JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_historico_senhas_usuario ON iam.historico_senhas(usuario_id);

COMMENT ON TABLE iam.historico_senhas IS 'Armazena histórico de senhas para impedir reuso';

-- Tabela de métodos de autenticação
CREATE TABLE iam.metodos_autenticacao (
    metodo_id VARCHAR(10) PRIMARY KEY, -- K01, P01, B01, etc.
    codigo_metodo VARCHAR(50) NOT NULL UNIQUE,
    nome_pt VARCHAR(100) NOT NULL,
    nome_en VARCHAR(100) NOT NULL,
    descricao_pt TEXT,
    descricao_en TEXT,
    categoria VARCHAR(50) NOT NULL, -- knowledge, possession, biometric, context, etc.
    fator VARCHAR(20) NOT NULL, -- knowledge, possession, inherence
    complexidade VARCHAR(20) NOT NULL DEFAULT 'medium', -- low, medium, high, very-high
    prioridade INTEGER NOT NULL DEFAULT 50, -- 0-100
    onda_implementacao INTEGER NOT NULL DEFAULT 1, -- 1-7
    nivel_seguranca VARCHAR(20) NOT NULL DEFAULT 'medium', -- low, medium, high, very-high
    pontuacao INTEGER NOT NULL DEFAULT 50, -- 0-100
    status VARCHAR(20) NOT NULL DEFAULT 'planned', -- planned, development, active, disabled, deprecated
    configuracoes JSONB DEFAULT '{}'::jsonb,
    adaptacoes_regionais JSONB DEFAULT '{}'::jsonb,
    metadados JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT ck_metodo_status CHECK (status IN ('planned', 'development', 'active', 'disabled', 'deprecated'))
);

COMMENT ON TABLE iam.metodos_autenticacao IS 'Catálogo de métodos de autenticação disponíveis';

-- Tabela de métodos habilitados por tenant
CREATE TABLE iam.tenant_metodos (
    tenant_metodo_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(tenant_id),
    metodo_id VARCHAR(10) NOT NULL REFERENCES iam.metodos_autenticacao(metodo_id),
    data_habilitacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    habilitado BOOLEAN NOT NULL DEFAULT true,
    aplicacao_default BOOLEAN NOT NULL DEFAULT false,
    configuracoes JSONB DEFAULT '{}'::jsonb,
    metadados JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT uq_tenant_metodo UNIQUE (tenant_id, metodo_id)
);

CREATE INDEX idx_tenant_metodos_tenant ON iam.tenant_metodos(tenant_id);
CREATE INDEX idx_tenant_metodos_metodo ON iam.tenant_metodos(metodo_id);

COMMENT ON TABLE iam.tenant_metodos IS 'Métodos de autenticação habilitados por tenant';

-- Tabela de métodos configurados por usuário
CREATE TABLE iam.usuario_metodos (
    usuario_metodo_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL REFERENCES iam.usuarios(usuario_id),
    metodo_id VARCHAR(10) NOT NULL REFERENCES iam.metodos_autenticacao(metodo_id),
    data_registro TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_modificacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    habilitado BOOLEAN NOT NULL DEFAULT true,
    verificado BOOLEAN NOT NULL DEFAULT false,
    preferencial BOOLEAN NOT NULL DEFAULT false,
    dados_autenticacao JSONB DEFAULT '{}'::jsonb, -- Armazena dados específicos do método (ex: TOTP secret, device tokens)
    nome_dispositivo VARCHAR(200),
    metadados JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT uq_usuario_metodo UNIQUE (usuario_id, metodo_id, nome_dispositivo)
);

CREATE INDEX idx_usuario_metodos_usuario ON iam.usuario_metodos(usuario_id);
CREATE INDEX idx_usuario_metodos_metodo ON iam.usuario_metodos(metodo_id);

COMMENT ON TABLE iam.usuario_metodos IS 'Métodos de autenticação registrados por usuário';

-- Tabela de sessões
CREATE TABLE iam.sessoes (
    sessao_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL REFERENCES iam.usuarios(usuario_id),
    token_refresh TEXT,
    cliente_id VARCHAR(100) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    dispositivo_id VARCHAR(255),
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_expiracao TIMESTAMP WITH TIME ZONE NOT NULL,
    data_ultimo_uso TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ativa BOOLEAN NOT NULL DEFAULT true,
    fatores_autenticados JSONB DEFAULT '[]'::jsonb, -- Lista de métodos usados na autenticação
    nivel_autenticacao VARCHAR(20) NOT NULL DEFAULT 'single_factor', -- single_factor, two_factor, multi_factor
    info_dispositivo JSONB DEFAULT '{}'::jsonb,
    info_localizacao JSONB DEFAULT '{}'::jsonb,
    metadados JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_sessoes_usuario ON iam.sessoes(usuario_id);
CREATE INDEX idx_sessoes_token ON iam.sessoes(token_refresh) WHERE token_refresh IS NOT NULL;
CREATE INDEX idx_sessoes_expiracao ON iam.sessoes(data_expiracao);

COMMENT ON TABLE iam.sessoes IS 'Sessões ativas dos usuários';

-- Tabela de autorização (grupos)
CREATE TABLE iam.grupos (
    grupo_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(tenant_id),
    nome VARCHAR(100) NOT NULL,
    descricao TEXT,
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_modificacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'ativo',
    metadados JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT uq_grupo_tenant_nome UNIQUE (tenant_id, nome),
    CONSTRAINT ck_grupo_status CHECK (status IN ('ativo', 'inativo'))
);

CREATE INDEX idx_grupos_tenant ON iam.grupos(tenant_id);

COMMENT ON TABLE iam.grupos IS 'Grupos para controle de acesso';

-- Tabela de associação usuário-grupo
CREATE TABLE iam.usuario_grupos (
    usuario_grupo_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL REFERENCES iam.usuarios(usuario_id),
    grupo_id UUID NOT NULL REFERENCES iam.grupos(grupo_id),
    data_atribuicao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    atribuido_por VARCHAR(100) NOT NULL DEFAULT CURRENT_USER,
    metadados JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT uq_usuario_grupo UNIQUE (usuario_id, grupo_id)
);

CREATE INDEX idx_usuario_grupos_usuario ON iam.usuario_grupos(usuario_id);
CREATE INDEX idx_usuario_grupos_grupo ON iam.usuario_grupos(grupo_id);

COMMENT ON TABLE iam.usuario_grupos IS 'Associação de usuários a grupos';

-- Tabela de tentativas de autenticação
CREATE TABLE iam.tentativas_autenticacao (
    tentativa_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(tenant_id),
    usuario_id UUID REFERENCES iam.usuarios(usuario_id),
    metodo_id VARCHAR(10) REFERENCES iam.metodos_autenticacao(metodo_id),
    data_tentativa TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sucesso BOOLEAN NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    dispositivo_id VARCHAR(255),
    info_dispositivo JSONB DEFAULT '{}'::jsonb,
    info_localizacao JSONB DEFAULT '{}'::jsonb,
    info_risco JSONB DEFAULT '{}'::jsonb,
    detalhes JSONB DEFAULT '{}'::jsonb,
    codigo_erro VARCHAR(50),
    mensagem_erro TEXT
);

CREATE INDEX idx_tentativas_tenant ON iam.tentativas_autenticacao(tenant_id);
CREATE INDEX idx_tentativas_usuario ON iam.tentativas_autenticacao(usuario_id) WHERE usuario_id IS NOT NULL;
CREATE INDEX idx_tentativas_metodo ON iam.tentativas_autenticacao(metodo_id) WHERE metodo_id IS NOT NULL;
CREATE INDEX idx_tentativas_data ON iam.tentativas_autenticacao(data_tentativa);
CREATE INDEX idx_tentativas_ip ON iam.tentativas_autenticacao(ip_address) WHERE ip_address IS NOT NULL;

COMMENT ON TABLE iam.tentativas_autenticacao IS 'Registro de tentativas de autenticação (bem-sucedidas ou não)';

-- Tabela de desafios de autenticação
CREATE TABLE iam.desafios_autenticacao (
    desafio_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID REFERENCES iam.usuarios(usuario_id),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(tenant_id),
    metodo_id VARCHAR(10) NOT NULL REFERENCES iam.metodos_autenticacao(metodo_id),
    codigo VARCHAR(100) UNIQUE,
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_expiracao TIMESTAMP WITH TIME ZONE NOT NULL,
    data_verificacao TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'pendente', -- pendente, verificado, expirado, cancelado
    tentativas INTEGER NOT NULL DEFAULT 0,
    max_tentativas INTEGER NOT NULL DEFAULT 3,
    ip_criacao VARCHAR(45),
    deviceid_criacao VARCHAR(255),
    info_contexto JSONB DEFAULT '{}'::jsonb,
    dados_desafio JSONB DEFAULT '{}'::jsonb,
    metadados JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT ck_desafio_status CHECK (status IN ('pendente', 'verificado', 'expirado', 'cancelado'))
);

CREATE INDEX idx_desafios_usuario ON iam.desafios_autenticacao(usuario_id) WHERE usuario_id IS NOT NULL;
CREATE INDEX idx_desafios_codigo ON iam.desafios_autenticacao(codigo) WHERE codigo IS NOT NULL;
CREATE INDEX idx_desafios_expiracao ON iam.desafios_autenticacao(data_expiracao);
CREATE INDEX idx_desafios_tenant ON iam.desafios_autenticacao(tenant_id);

COMMENT ON TABLE iam.desafios_autenticacao IS 'Desafios de autenticação (OTPs, magic links, etc.)';

-- Tabela de fluxos de autenticação
CREATE TABLE iam.fluxos_autenticacao (
    fluxo_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(tenant_id),
    nome VARCHAR(100) NOT NULL,
    descricao TEXT,
    passos JSONB NOT NULL DEFAULT '[]'::jsonb,
    adaptativo BOOLEAN NOT NULL DEFAULT false,
    nivel_seguranca VARCHAR(20) NOT NULL DEFAULT 'medium', -- low, medium, high, very-high
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_modificacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    criado_por VARCHAR(100) NOT NULL DEFAULT CURRENT_USER,
    status VARCHAR(20) NOT NULL DEFAULT 'ativo',
    metadados JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT uq_fluxo_tenant_nome UNIQUE (tenant_id, nome),
    CONSTRAINT ck_fluxo_status CHECK (status IN ('ativo', 'inativo', 'rascunho'))
);

CREATE INDEX idx_fluxos_tenant ON iam.fluxos_autenticacao(tenant_id);

COMMENT ON TABLE iam.fluxos_autenticacao IS 'Definições de fluxos de autenticação configuráveis';

-- Tabela de políticas de autenticação
CREATE TABLE iam.politicas_autenticacao (
    politica_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES iam.tenants(tenant_id),
    aplicacao_id UUID REFERENCES iam.aplicacoes(aplicacao_id),
    nome VARCHAR(100) NOT NULL,
    descricao TEXT,
    regras JSONB NOT NULL DEFAULT '{}'::jsonb,
    nivel_risco_minimo VARCHAR(20) NOT NULL DEFAULT 'low', -- low, medium, high
    requer_mfa BOOLEAN NOT NULL DEFAULT false,
    metodos_permitidos VARCHAR(10)[] DEFAULT NULL,
    metodos_proibidos VARCHAR(10)[] DEFAULT NULL,
    fluxo_padrao UUID REFERENCES iam.fluxos_autenticacao(fluxo_id),
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_modificacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    criado_por VARCHAR(100) NOT NULL DEFAULT CURRENT_USER,
    status VARCHAR(20) NOT NULL DEFAULT 'ativo',
    metadados JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT uq_politica_tenant_app_nome UNIQUE (tenant_id, aplicacao_id, nome) DEFERRABLE INITIALLY DEFERRED,
    CONSTRAINT ck_politica_status CHECK (status IN ('ativo', 'inativo', 'rascunho'))
);

CREATE INDEX idx_politicas_tenant ON iam.politicas_autenticacao(tenant_id);
CREATE INDEX idx_politicas_aplicacao ON iam.politicas_autenticacao(aplicacao_id) WHERE aplicacao_id IS NOT NULL;

COMMENT ON TABLE iam.politicas_autenticacao IS 'Políticas de autenticação para tenants e aplicações';

-- Tabela de tokens temporários
CREATE TABLE iam.tokens_temporarios (
    token_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token VARCHAR(255) UNIQUE NOT NULL,
    tenant_id UUID NOT NULL REFERENCES iam.tenants(tenant_id),
    usuario_id UUID REFERENCES iam.usuarios(usuario_id),
    tipo VARCHAR(50) NOT NULL, -- verification, reset_password, access, invite
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_expiracao TIMESTAMP WITH TIME ZONE NOT NULL,
    data_uso TIMESTAMP WITH TIME ZONE,
    usado BOOLEAN NOT NULL DEFAULT false,
    cancelado BOOLEAN NOT NULL DEFAULT false,
    escopo TEXT[] DEFAULT NULL,
    metadados JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT ck_token_tipo CHECK (tipo IN ('verification', 'reset_password', 'access', 'invite'))
);

CREATE INDEX idx_tokens_token ON iam.tokens_temporarios(token);
CREATE INDEX idx_tokens_usuario ON iam.tokens_temporarios(usuario_id) WHERE usuario_id IS NOT NULL;
CREATE INDEX idx_tokens_expiracao ON iam.tokens_temporarios(data_expiracao);
CREATE INDEX idx_tokens_tenant ON iam.tokens_temporarios(tenant_id);

COMMENT ON TABLE iam.tokens_temporarios IS 'Tokens temporários para reset de senha, convites, verificações, etc.';

-- Tabela de dispositivos confiáveis
CREATE TABLE iam.dispositivos_confiaveis (
    dispositivo_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL REFERENCES iam.usuarios(usuario_id),
    nome_dispositivo VARCHAR(200),
    identificador_dispositivo VARCHAR(255) NOT NULL,
    tipo_dispositivo VARCHAR(50) NOT NULL, -- desktop, mobile, tablet, other
    sistema_operacional VARCHAR(100),
    navegador VARCHAR(100),
    data_registro TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_ultimo_uso TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    confiavel BOOLEAN NOT NULL DEFAULT true,
    info_dispositivo JSONB DEFAULT '{}'::jsonb,
    info_localizacao JSONB DEFAULT '{}'::jsonb,
    metadados JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT uq_usuario_dispositivo UNIQUE (usuario_id, identificador_dispositivo)
);

CREATE INDEX idx_dispositivos_usuario ON iam.dispositivos_confiaveis(usuario_id);
CREATE INDEX idx_dispositivos_identificador ON iam.dispositivos_confiaveis(identificador_dispositivo);

COMMENT ON TABLE iam.dispositivos_confiaveis IS 'Dispositivos registrados e confiáveis dos usuários';

-- Tabela de perfis de risco
CREATE TABLE iam.perfis_risco (
    perfil_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL REFERENCES iam.usuarios(usuario_id),
    score_risco INTEGER NOT NULL DEFAULT 0, -- 0-100
    nivel_risco VARCHAR(20) NOT NULL DEFAULT 'medium', -- low, medium, high
    localizacoes_comuns JSONB DEFAULT '[]'::jsonb,
    dispositivos_comuns JSONB DEFAULT '[]'::jsonb,
    padroes_tempo JSONB DEFAULT '{}'::jsonb,
    padroes_comportamento JSONB DEFAULT '{}'::jsonb,
    anomalias_detectadas JSONB DEFAULT '[]'::jsonb,
    data_atualizacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadados JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT ck_perfil_nivel_risco CHECK (nivel_risco IN ('low', 'medium', 'high'))
);

CREATE INDEX idx_perfis_usuario ON iam.perfis_risco(usuario_id);
CREATE INDEX idx_perfis_nivel_risco ON iam.perfis_risco(nivel_risco);

COMMENT ON TABLE iam.perfis_risco IS 'Perfis de risco dos usuários para autenticação adaptativa';

-- Inserção de dados iniciais dos métodos de autenticação
INSERT INTO iam.metodos_autenticacao (
    metodo_id, codigo_metodo, nome_pt, nome_en, categoria, fator, complexidade, 
    prioridade, onda_implementacao, nivel_seguranca, pontuacao, status
) VALUES
    ('K01', 'traditional-password', 'Senha Tradicional', 'Traditional Password', 'knowledge', 'knowledge', 'low', 90, 1, 'low', 60, 'active'),
    ('K02', 'pin', 'PIN', 'PIN', 'knowledge', 'knowledge', 'low', 85, 1, 'low', 50, 'planned'),
    ('K05', 'otp', 'Senha de Uso Único (OTP)', 'One-Time Password (OTP)', 'knowledge', 'possession', 'medium', 85, 1, 'medium', 75, 'development'),
    ('P01', 'totp-hotp', 'TOTP/HOTP', 'TOTP/HOTP', 'possession', 'possession', 'medium', 80, 1, 'medium', 80, 'planned'),
    ('P02', 'fido2-webauthn', 'FIDO2/WebAuthn', 'FIDO2/WebAuthn', 'possession', 'possession', 'high', 90, 1, 'high', 95, 'planned'),
    ('P04', 'push-notification', 'Notificação Push', 'Push Notification', 'possession', 'possession', 'medium', 85, 1, 'medium', 85, 'planned'),
    ('B01', 'fingerprint', 'Reconhecimento de Impressão Digital', 'Fingerprint Recognition', 'biometric', 'inherence', 'high', 95, 1, 'high', 90, 'planned'),
    ('B02', 'facial-recognition', 'Reconhecimento Facial', 'Facial Recognition', 'biometric', 'inherence', 'high', 90, 1, 'high', 88, 'planned'),
    ('A01', 'geolocation', 'Geolocalização', 'Geolocation', 'adaptive', 'context', 'medium', 75, 1, 'medium', 70, 'planned'),
    ('A03', 'device-recognition', 'Reconhecimento de Dispositivo', 'Device Recognition', 'adaptive', 'context', 'medium', 80, 1, 'medium', 75, 'planned');

-- Criação de views úteis
CREATE OR REPLACE VIEW iam.vw_metodos_ativos AS
SELECT m.* 
FROM iam.metodos_autenticacao m
WHERE m.status = 'active';

CREATE OR REPLACE VIEW iam.vw_usuarios_com_mfa AS
SELECT u.usuario_id, u.nome_usuario, u.email, u.tenant_id, COUNT(um.metodo_id) AS num_metodos_mfa
FROM iam.usuarios u
JOIN iam.usuario_metodos um ON u.usuario_id = um.usuario_id
JOIN iam.metodos_autenticacao m ON um.metodo_id = m.metodo_id
WHERE um.habilitado = true AND um.verificado = true
  AND m.fator != 'knowledge'
GROUP BY u.usuario_id, u.nome_usuario, u.email, u.tenant_id
HAVING COUNT(um.metodo_id) > 0;

-- Funções úteis
CREATE OR REPLACE FUNCTION iam.fn_verificar_token(p_token VARCHAR)
RETURNS TABLE (
    token_id UUID,
    tipo VARCHAR,
    valido BOOLEAN,
    usuario_id UUID,
    tenant_id UUID,
    metadata JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        t.token_id,
        t.tipo,
        (NOT t.usado AND NOT t.cancelado AND t.data_expiracao > CURRENT_TIMESTAMP) AS valido,
        t.usuario_id,
        t.tenant_id,
        t.metadados
    FROM iam.tokens_temporarios t
    WHERE t.token = p_token;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Funções para auditoria e logging
CREATE OR REPLACE FUNCTION iam.fn_log_alteracao()
RETURNS TRIGGER AS $$
BEGIN
    NEW.data_modificacao = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers para auditoria e logging
CREATE TRIGGER trg_tenants_audit
BEFORE UPDATE ON iam.tenants
FOR EACH ROW
EXECUTE FUNCTION iam.fn_log_alteracao();

CREATE TRIGGER trg_aplicacoes_audit
BEFORE UPDATE ON iam.aplicacoes
FOR EACH ROW
EXECUTE FUNCTION iam.fn_log_alteracao();

CREATE TRIGGER trg_usuarios_audit
BEFORE UPDATE ON iam.usuarios
FOR EACH ROW
EXECUTE FUNCTION iam.fn_log_alteracao();

-- Índices para melhorar a pesquisa em JSONB
CREATE INDEX idx_usuarios_perfil_gin ON iam.usuarios USING GIN (dados_perfil jsonb_path_ops);
CREATE INDEX idx_tentativas_auth_risk_gin ON iam.tentativas_autenticacao USING GIN (info_risco jsonb_path_ops);
CREATE INDEX idx_fluxos_auth_steps_gin ON iam.fluxos_autenticacao USING GIN (passos jsonb_path_ops);

-- Permissões e segurança
ALTER DEFAULT PRIVILEGES IN SCHEMA iam GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO iam_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA iam GRANT USAGE ON SEQUENCES TO iam_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA iam GRANT EXECUTE ON FUNCTIONS TO iam_app;

COMMENT ON SCHEMA iam IS 'Esquema para gestão de identidade e acesso da plataforma INNOVABIZ';
