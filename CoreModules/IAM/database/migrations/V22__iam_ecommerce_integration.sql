-- ==========================================================================
-- Nome: V22__iam_ecommerce_integration.sql
-- Descrição: Migração para integração do IAM com E-Commerce
-- Autor: Equipa de Desenvolvimento INNOVABIZ
-- Data: 19/08/2025
-- ==========================================================================

-- Assegurar que o esquema para integrações existe
CREATE SCHEMA IF NOT EXISTS iam_integrations;

-- ==========================================================================
-- TABELAS DE INTEGRAÇÃO COM E-COMMERCE
-- ==========================================================================

-- Tabela de mapeamento de identidades com E-Commerce
CREATE TABLE IF NOT EXISTS iam_integrations.ecommerce_identities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    usuario_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    ec_usuario_id VARCHAR(255) NOT NULL,
    ec_tenant_id VARCHAR(255) NOT NULL,
    ec_perfil_tipo VARCHAR(50) NOT NULL,
    status_vinculo VARCHAR(30) NOT NULL DEFAULT 'ativo',
    nivel_acesso VARCHAR(30) NOT NULL DEFAULT 'cliente',
    detalhes_perfil JSONB,
    data_vinculacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_atualizacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadados JSONB,
    CONSTRAINT fk_ec_identities_usuario FOREIGN KEY (usuario_id) REFERENCES iam_core.usuarios(id) ON DELETE CASCADE,
    CONSTRAINT fk_ec_identities_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_management.tenants(id),
    CONSTRAINT uk_ec_identities_usuario_tenant UNIQUE (usuario_id, ec_tenant_id),
    CONSTRAINT ck_ec_identities_status CHECK (status_vinculo IN ('ativo', 'inativo', 'suspenso', 'pendente', 'revogado')),
    CONSTRAINT ck_ec_identities_nivel CHECK (nivel_acesso IN ('cliente', 'cliente_vip', 'vendedor', 'admin_loja', 'admin_plataforma')),
    CONSTRAINT ck_ec_identities_tipo CHECK (ec_perfil_tipo IN ('cliente_individual', 'cliente_empresa', 'vendedor_individual', 'vendedor_empresa', 'loja', 'administrador'))
);

-- Tabela para tokens de acesso do E-Commerce
CREATE TABLE IF NOT EXISTS iam_integrations.ecommerce_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    identity_id UUID NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    refresh_token_hash VARCHAR(255),
    escopo VARCHAR(255) NOT NULL,
    data_expiracao TIMESTAMP WITH TIME ZONE NOT NULL,
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ultimo_uso TIMESTAMP WITH TIME ZONE,
    dispositivo_id VARCHAR(255),
    metadados JSONB,
    CONSTRAINT fk_ec_tokens_identity FOREIGN KEY (identity_id) REFERENCES iam_integrations.ecommerce_identities(id) ON DELETE CASCADE
);

CREATE INDEX idx_ec_tokens_hash ON iam_integrations.ecommerce_tokens(token_hash);
CREATE INDEX idx_ec_tokens_expiracao ON iam_integrations.ecommerce_tokens(data_expiracao);

-- Tabela de endereços associados a usuários E-Commerce
CREATE TABLE IF NOT EXISTS iam_integrations.ecommerce_enderecos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    identity_id UUID NOT NULL,
    ec_endereco_id VARCHAR(255),
    tipo VARCHAR(30) NOT NULL,
    nome VARCHAR(100) NOT NULL,
    logradouro VARCHAR(255) NOT NULL,
    numero VARCHAR(30),
    complemento VARCHAR(100),
    bairro VARCHAR(100) NOT NULL,
    cidade VARCHAR(100) NOT NULL,
    estado VARCHAR(50) NOT NULL,
    pais VARCHAR(50) NOT NULL DEFAULT 'Angola',
    cep VARCHAR(20),
    referencia VARCHAR(255),
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    principal BOOLEAN NOT NULL DEFAULT false,
    verificado BOOLEAN NOT NULL DEFAULT false,
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_atualizacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_ec_enderecos_identity FOREIGN KEY (identity_id) REFERENCES iam_integrations.ecommerce_identities(id) ON DELETE CASCADE,
    CONSTRAINT ck_ec_enderecos_tipo CHECK (tipo IN ('residencial', 'comercial', 'entrega', 'cobranca', 'outro'))
);

-- Tabela para métodos de pagamento associados a usuários E-Commerce
CREATE TABLE IF NOT EXISTS iam_integrations.ecommerce_metodos_pagamento (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    identity_id UUID NOT NULL,
    ec_metodo_id VARCHAR(255),
    tipo VARCHAR(50) NOT NULL,
    titulo VARCHAR(100) NOT NULL,
    token_pagamento VARCHAR(255),
    ultimos_digitos VARCHAR(4),
    bandeira VARCHAR(50),
    data_expiracao DATE,
    principal BOOLEAN NOT NULL DEFAULT false,
    verificado BOOLEAN NOT NULL DEFAULT false,
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_atualizacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadados JSONB,
    CONSTRAINT fk_ec_metodos_pagamento_identity FOREIGN KEY (identity_id) REFERENCES iam_integrations.ecommerce_identities(id) ON DELETE CASCADE,
    CONSTRAINT ck_ec_metodos_pagamento_tipo CHECK (tipo IN ('cartao_credito', 'cartao_debito', 'boleto', 'mobile_money', 'transferencia', 'pix', 'multicaixa', 'dinheiro', 'outro'))
);

-- Tabela para preferências de usuários E-Commerce
CREATE TABLE IF NOT EXISTS iam_integrations.ecommerce_preferencias (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    identity_id UUID NOT NULL,
    notificacao_pedido BOOLEAN DEFAULT true,
    notificacao_promocao BOOLEAN DEFAULT true,
    notificacao_carrinho BOOLEAN DEFAULT true,
    notificacao_entrega BOOLEAN DEFAULT true,
    marketing_email BOOLEAN DEFAULT true,
    marketing_sms BOOLEAN DEFAULT false,
    marketing_whatsapp BOOLEAN DEFAULT false,
    moeda_preferida VARCHAR(3) DEFAULT 'AOA',
    idioma_preferido VARCHAR(5) DEFAULT 'pt-AO',
    configuracoes JSONB,
    data_atualizacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_ec_preferencias_identity FOREIGN KEY (identity_id) REFERENCES iam_integrations.ecommerce_identities(id) ON DELETE CASCADE
);

-- Tabela para registro de consentimentos E-Commerce
CREATE TABLE IF NOT EXISTS iam_integrations.ecommerce_consent_log (
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
    CONSTRAINT fk_ec_consent_usuario FOREIGN KEY (usuario_id) REFERENCES iam_core.usuarios(id) ON DELETE CASCADE,
    CONSTRAINT fk_ec_consent_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_management.tenants(id),
    CONSTRAINT fk_ec_consent_identity FOREIGN KEY (identity_id) REFERENCES iam_integrations.ecommerce_identities(id) ON DELETE CASCADE,
    CONSTRAINT ck_ec_consent_tipo CHECK (tipo_consentimento IN ('termos_uso', 'privacidade', 'marketing', 'cookies', 'localizacao', 'dados_pessoais', 'perfil_comportamental', 'pagamento', 'biometria', 'completo'))
);

-- ==========================================================================
-- FUNÇÕES E PROCEDIMENTOS ARMAZENADOS
-- ==========================================================================

-- Função para vincular usuário IAM com E-Commerce
CREATE OR REPLACE FUNCTION iam_integrations.vincular_usuario_ecommerce(
    p_usuario_id UUID,
    p_tenant_id UUID,
    p_ec_usuario_id VARCHAR(255),
    p_ec_tenant_id VARCHAR(255),
    p_ec_perfil_tipo VARCHAR(50),
    p_nivel_acesso VARCHAR(30),
    p_detalhes_perfil JSONB DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    v_identity_id UUID;
    v_evento_id UUID;
BEGIN
    -- Verificar se o usuário já possui vínculo ativo
    SELECT id INTO v_identity_id
    FROM iam_integrations.ecommerce_identities
    WHERE usuario_id = p_usuario_id AND ec_tenant_id = p_ec_tenant_id AND status_vinculo = 'ativo';
    
    IF FOUND THEN
        -- Atualizar vínculo existente
        UPDATE iam_integrations.ecommerce_identities
        SET nivel_acesso = COALESCE(p_nivel_acesso, nivel_acesso),
            ec_perfil_tipo = COALESCE(p_ec_perfil_tipo, ec_perfil_tipo),
            detalhes_perfil = COALESCE(p_detalhes_perfil, detalhes_perfil),
            data_atualizacao = NOW()
        WHERE id = v_identity_id;
    ELSE
        -- Criar novo vínculo
        INSERT INTO iam_integrations.ecommerce_identities (
            usuario_id, tenant_id, ec_usuario_id, ec_tenant_id,
            ec_perfil_tipo, nivel_acesso, detalhes_perfil
        ) VALUES (
            p_usuario_id, p_tenant_id, p_ec_usuario_id, p_ec_tenant_id,
            p_ec_perfil_tipo, COALESCE(p_nivel_acesso, 'cliente'), p_detalhes_perfil
        )
        RETURNING id INTO v_identity_id;
    END IF;
    
    -- Registrar evento de auditoria
    SELECT iam_audit.registrar_evento_auditoria(
        p_tenant_id,
        p_usuario_id,
        NULL,
        'VINCULACAO_ECOMMERCE',
        'INTEGRACAO',
        'media',
        'Vínculo de usuário com E-Commerce',
        jsonb_build_object(
            'ec_usuario_id', p_ec_usuario_id,
            'ec_tenant_id', p_ec_tenant_id,
            'ec_perfil_tipo', p_ec_perfil_tipo,
            'nivel_acesso', COALESCE(p_nivel_acesso, 'cliente')
        ),
        'sucesso',
        NULL,
        NULL,
        'IAM_ECOMMERCE',
        'usuario',
        p_usuario_id::TEXT
    ) INTO v_evento_id;
    
    RETURN v_identity_id;
END;
$$ LANGUAGE plpgsql;

-- Função para criar token de acesso E-Commerce
CREATE OR REPLACE FUNCTION iam_integrations.criar_token_ecommerce(
    p_identity_id UUID,
    p_token VARCHAR(255),
    p_refresh_token VARCHAR(255),
    p_escopo VARCHAR(255),
    p_dispositivo_id VARCHAR(255) DEFAULT NULL,
    p_duracao_minutos INTEGER DEFAULT 60
) RETURNS UUID AS $$
DECLARE
    v_token_id UUID;
BEGIN
    INSERT INTO iam_integrations.ecommerce_tokens (
        identity_id,
        token_hash,
        refresh_token_hash,
        escopo,
        dispositivo_id,
        data_expiracao
    ) VALUES (
        p_identity_id,
        crypt(p_token, gen_salt('bf')),
        CASE WHEN p_refresh_token IS NOT NULL THEN crypt(p_refresh_token, gen_salt('bf')) ELSE NULL END,
        p_escopo,
        p_dispositivo_id,
        NOW() + (p_duracao_minutos * INTERVAL '1 minute')
    )
    RETURNING id INTO v_token_id;
    
    RETURN v_token_id;
END;
$$ LANGUAGE plpgsql;

-- Função para adicionar endereço para usuário E-Commerce
CREATE OR REPLACE FUNCTION iam_integrations.adicionar_endereco_ecommerce(
    p_identity_id UUID,
    p_tipo VARCHAR(30),
    p_nome VARCHAR(100),
    p_logradouro VARCHAR(255),
    p_numero VARCHAR(30),
    p_complemento VARCHAR(100),
    p_bairro VARCHAR(100),
    p_cidade VARCHAR(100),
    p_estado VARCHAR(50),
    p_pais VARCHAR(50),
    p_cep VARCHAR(20),
    p_principal BOOLEAN DEFAULT false
) RETURNS UUID AS $$
DECLARE
    v_endereco_id UUID;
BEGIN
    -- Se marcado como principal, desmarcar outros endereços
    IF p_principal THEN
        UPDATE iam_integrations.ecommerce_enderecos
        SET principal = false
        WHERE identity_id = p_identity_id;
    END IF;
    
    INSERT INTO iam_integrations.ecommerce_enderecos (
        identity_id, tipo, nome, logradouro, numero,
        complemento, bairro, cidade, estado, pais,
        cep, principal
    ) VALUES (
        p_identity_id, p_tipo, p_nome, p_logradouro, p_numero,
        p_complemento, p_bairro, p_cidade, p_estado, p_pais,
        p_cep, p_principal
    )
    RETURNING id INTO v_endereco_id;
    
    RETURN v_endereco_id;
END;
$$ LANGUAGE plpgsql;

-- ==========================================================================
-- VIEWS
-- ==========================================================================

-- View para relatório de integrações ativas com E-Commerce
CREATE OR REPLACE VIEW iam_integrations.vw_ecommerce_integrations AS
SELECT
    eci.id AS identity_id,
    u.id AS usuario_id,
    u.nome_completo,
    u.email,
    t.nome AS tenant_nome,
    t.slug AS tenant_slug,
    eci.ec_usuario_id,
    eci.ec_tenant_id,
    eci.ec_perfil_tipo,
    eci.nivel_acesso,
    eci.status_vinculo,
    eci.data_vinculacao,
    eci.data_atualizacao,
    COUNT(ect.id) AS tokens_ativos,
    MAX(ect.data_expiracao) AS data_expiracao_ultimo_token,
    COUNT(DISTINCT ece.id) AS qtd_enderecos,
    COUNT(DISTINCT ecmp.id) AS qtd_metodos_pagamento
FROM
    iam_integrations.ecommerce_identities eci
JOIN
    iam_core.usuarios u ON eci.usuario_id = u.id
JOIN
    tenant_management.tenants t ON eci.tenant_id = t.id
LEFT JOIN
    iam_integrations.ecommerce_tokens ect ON 
        eci.id = ect.identity_id AND 
        ect.data_expiracao > NOW()
LEFT JOIN
    iam_integrations.ecommerce_enderecos ece ON
        eci.id = ece.identity_id
LEFT JOIN
    iam_integrations.ecommerce_metodos_pagamento ecmp ON
        eci.id = ecmp.identity_id
GROUP BY
    eci.id, u.id, u.nome_completo, u.email, t.nome, t.slug,
    eci.ec_usuario_id, eci.ec_tenant_id, eci.ec_perfil_tipo, eci.nivel_acesso, eci.status_vinculo,
    eci.data_vinculacao, eci.data_atualizacao;

-- View para informações completas de clientes
CREATE OR REPLACE VIEW iam_integrations.vw_ecommerce_clientes AS
SELECT
    u.id AS usuario_id,
    u.nome_completo,
    u.email,
    u.celular,
    u.documento_principal,
    u.tipo_documento,
    u.data_nascimento,
    u.genero,
    eci.id AS identity_id,
    eci.ec_usuario_id,
    eci.ec_perfil_tipo,
    eci.nivel_acesso,
    eci.detalhes_perfil,
    ecp.notificacao_pedido,
    ecp.notificacao_promocao,
    ecp.marketing_email,
    ecp.marketing_sms,
    ecp.marketing_whatsapp,
    ecp.moeda_preferida,
    ecp.idioma_preferido,
    (SELECT json_agg(row_to_json(e)) FROM (
        SELECT id, tipo, nome, logradouro, numero, complemento, bairro, cidade, estado, pais, cep, principal
        FROM iam_integrations.ecommerce_enderecos
        WHERE identity_id = eci.id
    ) e) AS enderecos,
    (SELECT json_agg(row_to_json(mp)) FROM (
        SELECT id, tipo, titulo, ultimos_digitos, bandeira, data_expiracao, principal
        FROM iam_integrations.ecommerce_metodos_pagamento
        WHERE identity_id = eci.id
    ) mp) AS metodos_pagamento
FROM
    iam_integrations.ecommerce_identities eci
JOIN
    iam_core.usuarios u ON eci.usuario_id = u.id
LEFT JOIN
    iam_integrations.ecommerce_preferencias ecp ON eci.id = ecp.identity_id
WHERE
    eci.status_vinculo = 'ativo';

-- ==========================================================================
-- PERMISSÕES
-- ==========================================================================

-- Criar permissões padrão para integração E-Commerce
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
    ('00000000-0000-0000-0000-000000000001', 'ECOMMERCE_ADMIN', 'Administrar E-Commerce', 'Permissão para administrar integrações com E-Commerce', 'INTEGRACAO', 'todas', 'ECOMMERCE', true),
    ('00000000-0000-0000-0000-000000000001', 'ECOMMERCE_VINCULAR', 'Vincular E-Commerce', 'Permissão para vincular contas com E-Commerce', 'INTEGRACAO', 'criar', 'ECOMMERCE_IDENTITY', true),
    ('00000000-0000-0000-0000-000000000001', 'ECOMMERCE_DESVINCULAR', 'Desvincular E-Commerce', 'Permissão para desvincular contas com E-Commerce', 'INTEGRACAO', 'excluir', 'ECOMMERCE_IDENTITY', true),
    ('00000000-0000-0000-0000-000000000001', 'ECOMMERCE_CONSULTA', 'Consultar E-Commerce', 'Permissão para consultar integrações com E-Commerce', 'INTEGRACAO', 'ler', 'ECOMMERCE_IDENTITY', true),
    ('00000000-0000-0000-0000-000000000001', 'ECOMMERCE_ENDERECO', 'Gerenciar Endereços E-Commerce', 'Permissão para gerenciar endereços no E-Commerce', 'INTEGRACAO', 'todas', 'ECOMMERCE_ENDERECO', true),
    ('00000000-0000-0000-0000-000000000001', 'ECOMMERCE_PAGAMENTO', 'Gerenciar Métodos Pagamento', 'Permissão para gerenciar métodos de pagamento no E-Commerce', 'INTEGRACAO', 'todas', 'ECOMMERCE_PAGAMENTO', true)
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
    codigo LIKE 'ECOMMERCE_%'
ON CONFLICT (perfil_id, permissao_id) DO NOTHING;

-- ==========================================================================
-- COMENTÁRIOS
-- ==========================================================================

COMMENT ON TABLE iam_integrations.ecommerce_identities IS 'Tabela de mapeamento de identidades entre IAM e E-Commerce';
COMMENT ON TABLE iam_integrations.ecommerce_tokens IS 'Tokens de acesso para integração com E-Commerce';
COMMENT ON TABLE iam_integrations.ecommerce_enderecos IS 'Endereços associados a usuários do E-Commerce';
COMMENT ON TABLE iam_integrations.ecommerce_metodos_pagamento IS 'Métodos de pagamento associados a usuários do E-Commerce';
COMMENT ON TABLE iam_integrations.ecommerce_preferencias IS 'Preferências de usuários no E-Commerce';
COMMENT ON TABLE iam_integrations.ecommerce_consent_log IS 'Registro de consentimentos para acesso ao E-Commerce';

COMMENT ON FUNCTION iam_integrations.vincular_usuario_ecommerce IS 'Função para vincular usuário do IAM com conta do E-Commerce';
COMMENT ON FUNCTION iam_integrations.criar_token_ecommerce IS 'Função para criar token de acesso para E-Commerce';
COMMENT ON FUNCTION iam_integrations.adicionar_endereco_ecommerce IS 'Função para adicionar endereço para usuário do E-Commerce';

COMMENT ON VIEW iam_integrations.vw_ecommerce_integrations IS 'Visão consolidada de integrações ativas com E-Commerce';
COMMENT ON VIEW iam_integrations.vw_ecommerce_clientes IS 'Visão detalhada de clientes do E-Commerce com todas suas informações';