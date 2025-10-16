-- ==========================================================================
-- Nome: V21__iam_mobile_money_integration.sql
-- Descrição: Migração para integração do IAM com Mobile Money
-- Autor: Equipa de Desenvolvimento INNOVABIZ
-- Data: 19/08/2025
-- ==========================================================================

-- Criar esquema para integrações se não existir
CREATE SCHEMA IF NOT EXISTS iam_integrations;

-- ==========================================================================
-- TABELAS DE INTEGRAÇÃO COM MOBILE MONEY
-- ==========================================================================

-- Tabela de mapeamento de identidades com Mobile Money
CREATE TABLE IF NOT EXISTS iam_integrations.mobile_money_identities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    usuario_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    mm_usuario_id VARCHAR(255) NOT NULL,
    mm_tenant_id VARCHAR(255) NOT NULL,
    status_vinculo VARCHAR(30) NOT NULL DEFAULT 'ativo',
    nivel_acesso VARCHAR(30) NOT NULL DEFAULT 'basico',
    detalhes_autorizacao JSONB,
    data_vinculacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_atualizacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadados JSONB,
    CONSTRAINT fk_mm_identities_usuario FOREIGN KEY (usuario_id) REFERENCES iam_core.usuarios(id) ON DELETE CASCADE,
    CONSTRAINT fk_mm_identities_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_management.tenants(id),
    CONSTRAINT uk_mm_identities_usuario_tenant UNIQUE (usuario_id, mm_tenant_id),
    CONSTRAINT ck_mm_identities_status CHECK (status_vinculo IN ('ativo', 'inativo', 'suspenso', 'pendente', 'revogado')),
    CONSTRAINT ck_mm_identities_nivel CHECK (nivel_acesso IN ('basico', 'intermediario', 'avancado', 'administrador', 'personalizado'))
);

-- Tabela para tokens de acesso do Mobile Money
CREATE TABLE IF NOT EXISTS iam_integrations.mobile_money_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    identity_id UUID NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    refresh_token_hash VARCHAR(255),
    escopo VARCHAR(255) NOT NULL,
    data_expiracao TIMESTAMP WITH TIME ZONE NOT NULL,
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ultimo_uso TIMESTAMP WITH TIME ZONE,
    metadados JSONB,
    CONSTRAINT fk_mm_tokens_identity FOREIGN KEY (identity_id) REFERENCES iam_integrations.mobile_money_identities(id) ON DELETE CASCADE
);

CREATE INDEX idx_mm_tokens_hash ON iam_integrations.mobile_money_tokens(token_hash);
CREATE INDEX idx_mm_tokens_expiracao ON iam_integrations.mobile_money_tokens(data_expiracao);

-- Tabela para permissões específicas do Mobile Money
CREATE TABLE IF NOT EXISTS iam_integrations.mobile_money_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    identity_id UUID NOT NULL,
    codigo VARCHAR(100) NOT NULL,
    nome VARCHAR(100) NOT NULL,
    descricao VARCHAR(500),
    data_concessao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_expiracao TIMESTAMP WITH TIME ZONE,
    concedido_por UUID,
    CONSTRAINT fk_mm_permissions_identity FOREIGN KEY (identity_id) REFERENCES iam_integrations.mobile_money_identities(id) ON DELETE CASCADE,
    CONSTRAINT uk_mm_permissions_identity_codigo UNIQUE (identity_id, codigo)
);

-- Tabela para registro de consentimentos para Mobile Money
CREATE TABLE IF NOT EXISTS iam_integrations.mobile_money_consent_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    usuario_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    identity_id UUID NOT NULL,
    tipo_consentimento VARCHAR(50) NOT NULL,
    descricao TEXT NOT NULL,
    ip_origem VARCHAR(50) NOT NULL,
    user_agent VARCHAR(500),
    data_consentimento TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_expiracao TIMESTAMP WITH TIME ZONE,
    hash_documento VARCHAR(255),
    documento_url VARCHAR(500),
    metadados JSONB,
    CONSTRAINT fk_mm_consent_usuario FOREIGN KEY (usuario_id) REFERENCES iam_core.usuarios(id) ON DELETE CASCADE,
    CONSTRAINT fk_mm_consent_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_management.tenants(id),
    CONSTRAINT fk_mm_consent_identity FOREIGN KEY (identity_id) REFERENCES iam_integrations.mobile_money_identities(id) ON DELETE CASCADE,
    CONSTRAINT ck_mm_consent_tipo CHECK (tipo_consentimento IN ('acesso_basico', 'transferencia', 'pagamento', 'consulta_saldo', 'historico_transacoes', 'contatos', 'notificacoes', 'biometria', 'localizacao', 'completo'))
);

-- Tabela para registro de auditorias específicas de integração
CREATE TABLE IF NOT EXISTS iam_integrations.mobile_money_audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    usuario_id UUID,
    tenant_id UUID NOT NULL,
    identity_id UUID,
    tipo_evento VARCHAR(100) NOT NULL,
    descricao TEXT NOT NULL,
    detalhes JSONB,
    resultado VARCHAR(50) NOT NULL,
    ip_origem VARCHAR(50),
    user_agent VARCHAR(500),
    data_ocorrencia TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_mm_audit_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_management.tenants(id),
    CONSTRAINT fk_mm_audit_identity FOREIGN KEY (identity_id) REFERENCES iam_integrations.mobile_money_identities(id) ON DELETE SET NULL,
    CONSTRAINT ck_mm_audit_resultado CHECK (resultado IN ('sucesso', 'falha', 'negado', 'erro', 'aviso'))
);

-- ==========================================================================
-- FUNÇÕES E PROCEDIMENTOS ARMAZENADOS
-- ==========================================================================

-- Função para vincular usuário IAM com Mobile Money
CREATE OR REPLACE FUNCTION iam_integrations.vincular_usuario_mobile_money(
    p_usuario_id UUID,
    p_tenant_id UUID,
    p_mm_usuario_id VARCHAR(255),
    p_mm_tenant_id VARCHAR(255),
    p_nivel_acesso VARCHAR(30),
    p_detalhes_autorizacao JSONB DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    v_identity_id UUID;
    v_evento_id UUID;
BEGIN
    -- Verificar se o usuário já possui vínculo ativo
    SELECT id INTO v_identity_id
    FROM iam_integrations.mobile_money_identities
    WHERE usuario_id = p_usuario_id AND mm_tenant_id = p_mm_tenant_id AND status_vinculo = 'ativo';
    
    IF FOUND THEN
        -- Atualizar vínculo existente
        UPDATE iam_integrations.mobile_money_identities
        SET nivel_acesso = COALESCE(p_nivel_acesso, nivel_acesso),
            detalhes_autorizacao = COALESCE(p_detalhes_autorizacao, detalhes_autorizacao),
            data_atualizacao = NOW()
        WHERE id = v_identity_id;
    ELSE
        -- Criar novo vínculo
        INSERT INTO iam_integrations.mobile_money_identities (
            usuario_id, tenant_id, mm_usuario_id, mm_tenant_id,
            nivel_acesso, detalhes_autorizacao
        ) VALUES (
            p_usuario_id, p_tenant_id, p_mm_usuario_id, p_mm_tenant_id,
            COALESCE(p_nivel_acesso, 'basico'), p_detalhes_autorizacao
        )
        RETURNING id INTO v_identity_id;
    END IF;
    
    -- Registrar evento de auditoria
    SELECT iam_audit.registrar_evento_auditoria(
        p_tenant_id,
        p_usuario_id,
        NULL,
        'VINCULACAO_MOBILE_MONEY',
        'INTEGRACAO',
        'media',
        'Vínculo de usuário com Mobile Money',
        jsonb_build_object(
            'mm_usuario_id', p_mm_usuario_id,
            'mm_tenant_id', p_mm_tenant_id,
            'nivel_acesso', COALESCE(p_nivel_acesso, 'basico')
        ),
        'sucesso',
        NULL,
        NULL,
        'IAM_MOBILE_MONEY',
        'usuario',
        p_usuario_id::TEXT
    ) INTO v_evento_id;
    
    RETURN v_identity_id;
END;
$$ LANGUAGE plpgsql;

-- Função para criar token de acesso Mobile Money
CREATE OR REPLACE FUNCTION iam_integrations.criar_token_mobile_money(
    p_identity_id UUID,
    p_token VARCHAR(255),
    p_refresh_token VARCHAR(255),
    p_escopo VARCHAR(255),
    p_duracao_minutos INTEGER DEFAULT 60
) RETURNS UUID AS $$
DECLARE
    v_token_id UUID;
BEGIN
    INSERT INTO iam_integrations.mobile_money_tokens (
        identity_id,
        token_hash,
        refresh_token_hash,
        escopo,
        data_expiracao
    ) VALUES (
        p_identity_id,
        crypt(p_token, gen_salt('bf')),
        CASE WHEN p_refresh_token IS NOT NULL THEN crypt(p_refresh_token, gen_salt('bf')) ELSE NULL END,
        p_escopo,
        NOW() + (p_duracao_minutos * INTERVAL '1 minute')
    )
    RETURNING id INTO v_token_id;
    
    RETURN v_token_id;
END;
$$ LANGUAGE plpgsql;

-- ==========================================================================
-- VIEWS
-- ==========================================================================

-- View para relatório de integrações ativas com Mobile Money
CREATE OR REPLACE VIEW iam_integrations.vw_mobile_money_integrations AS
SELECT
    mmi.id AS identity_id,
    u.id AS usuario_id,
    u.nome_completo,
    u.email,
    t.nome AS tenant_nome,
    t.slug AS tenant_slug,
    mmi.mm_usuario_id,
    mmi.mm_tenant_id,
    mmi.nivel_acesso,
    mmi.status_vinculo,
    mmi.data_vinculacao,
    mmi.data_atualizacao,
    COUNT(mmt.id) AS tokens_ativos,
    MAX(mmt.data_expiracao) AS data_expiracao_ultimo_token
FROM
    iam_integrations.mobile_money_identities mmi
JOIN
    iam_core.usuarios u ON mmi.usuario_id = u.id
JOIN
    tenant_management.tenants t ON mmi.tenant_id = t.id
LEFT JOIN
    iam_integrations.mobile_money_tokens mmt ON 
        mmi.id = mmt.identity_id AND 
        mmt.data_expiracao > NOW()
GROUP BY
    mmi.id, u.id, u.nome_completo, u.email, t.nome, t.slug,
    mmi.mm_usuario_id, mmi.mm_tenant_id, mmi.nivel_acesso, mmi.status_vinculo,
    mmi.data_vinculacao, mmi.data_atualizacao;

-- ==========================================================================
-- PERMISSÕES
-- ==========================================================================

-- Criar permissões padrão para integração Mobile Money
INSERT INTO iam_core.permissoes (
    tenant_id,
    codigo,
    nome,
    descricao,
    modulo,
    acao,
    recurso,
    interno
) VALUES
    ('00000000-0000-0000-0000-000000000001', 'MOBILE_MONEY_ADMIN', 'Administrar Mobile Money', 'Permissão para administrar integrações com Mobile Money', 'INTEGRACAO', 'todas', 'MOBILE_MONEY', true),
    ('00000000-0000-0000-0000-000000000001', 'MOBILE_MONEY_VINCULAR', 'Vincular Mobile Money', 'Permissão para vincular contas com Mobile Money', 'INTEGRACAO', 'criar', 'MOBILE_MONEY_IDENTITY', true),
    ('00000000-0000-0000-0000-000000000001', 'MOBILE_MONEY_DESVINCULAR', 'Desvincular Mobile Money', 'Permissão para desvincular contas com Mobile Money', 'INTEGRACAO', 'excluir', 'MOBILE_MONEY_IDENTITY', true),
    ('00000000-0000-0000-0000-000000000001', 'MOBILE_MONEY_CONSULTA', 'Consultar Mobile Money', 'Permissão para consultar integrações com Mobile Money', 'INTEGRACAO', 'ler', 'MOBILE_MONEY_IDENTITY', true)
ON CONFLICT (tenant_id, codigo) DO NOTHING;

-- Associar permissões ao perfil de administrador
INSERT INTO iam_core.perfil_permissoes (
    perfil_id,
    permissao_id,
    data_concessao
) 
SELECT 
    '00000000-0000-0000-0000-000000000010', 
    id, 
    NOW()
FROM 
    iam_core.permissoes 
WHERE 
    tenant_id = '00000000-0000-0000-0000-000000000001' AND 
    codigo LIKE 'MOBILE_MONEY_%'
ON CONFLICT (perfil_id, permissao_id) DO NOTHING;

-- ==========================================================================
-- COMENTÁRIOS
-- ==========================================================================

COMMENT ON TABLE iam_integrations.mobile_money_identities IS 'Tabela de mapeamento de identidades entre IAM e Mobile Money';
COMMENT ON TABLE iam_integrations.mobile_money_tokens IS 'Tokens de acesso para integração com Mobile Money';
COMMENT ON TABLE iam_integrations.mobile_money_permissions IS 'Permissões específicas para integração com Mobile Money';
COMMENT ON TABLE iam_integrations.mobile_money_consent_log IS 'Registro de consentimentos para acesso ao Mobile Money';
COMMENT ON TABLE iam_integrations.mobile_money_audit_log IS 'Log de auditoria específico para integração com Mobile Money';

COMMENT ON FUNCTION iam_integrations.vincular_usuario_mobile_money IS 'Função para vincular usuário do IAM com conta do Mobile Money';
COMMENT ON FUNCTION iam_integrations.criar_token_mobile_money IS 'Função para criar token de acesso para Mobile Money';

COMMENT ON VIEW iam_integrations.vw_mobile_money_integrations IS 'Visão consolidada de integrações ativas com Mobile Money';