-- ==========================================================================
-- Nome: V23__iam_bureau_credito_integration.sql
-- Descrição: Migração para integração do IAM com Bureau de Créditos
-- Autor: Equipa de Desenvolvimento INNOVABIZ
-- Data: 19/08/2025
-- ==========================================================================

-- Assegurar que o esquema para integrações existe
CREATE SCHEMA IF NOT EXISTS iam_integrations;

-- ==========================================================================
-- TABELAS DE INTEGRAÇÃO COM BUREAU DE CRÉDITOS
-- ==========================================================================

-- Tabela de mapeamento de identidades com Bureau de Créditos
CREATE TABLE IF NOT EXISTS iam_integrations.bureau_credito_identities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    usuario_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    bc_usuario_id VARCHAR(255) NOT NULL,
    bc_tenant_id VARCHAR(255) NOT NULL,
    tipo_vinculo VARCHAR(50) NOT NULL,
    status_vinculo VARCHAR(30) NOT NULL DEFAULT 'ativo',
    nivel_acesso VARCHAR(30) NOT NULL DEFAULT 'consulta_basica',
    detalhes_autorizacao JSONB,
    data_vinculacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_atualizacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadados JSONB,
    CONSTRAINT fk_bc_identities_usuario FOREIGN KEY (usuario_id) REFERENCES iam_core.usuarios(id) ON DELETE CASCADE,
    CONSTRAINT fk_bc_identities_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_management.tenants(id),
    CONSTRAINT uk_bc_identities_usuario_tenant UNIQUE (usuario_id, bc_tenant_id),
    CONSTRAINT ck_bc_identities_status CHECK (status_vinculo IN ('ativo', 'inativo', 'suspenso', 'pendente', 'revogado')),
    CONSTRAINT ck_bc_identities_nivel CHECK (nivel_acesso IN ('consulta_basica', 'consulta_detalhada', 'consulta_completa', 'consulta_historico', 'administracao', 'personalizado')),
    CONSTRAINT ck_bc_identities_tipo_vinculo CHECK (tipo_vinculo IN ('pessoa_fisica', 'pessoa_juridica', 'instituicao_financeira', 'comercio', 'administrador', 'regulador', 'auditor'))
);

-- Tabela para tokens de acesso do Bureau de Créditos
CREATE TABLE IF NOT EXISTS iam_integrations.bureau_credito_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    identity_id UUID NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    refresh_token_hash VARCHAR(255),
    escopo VARCHAR(255) NOT NULL,
    data_expiracao TIMESTAMP WITH TIME ZONE NOT NULL,
    data_criacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ultimo_uso TIMESTAMP WITH TIME ZONE,
    origem_requisicao VARCHAR(255),
    metadados JSONB,
    CONSTRAINT fk_bc_tokens_identity FOREIGN KEY (identity_id) REFERENCES iam_integrations.bureau_credito_identities(id) ON DELETE CASCADE
);

CREATE INDEX idx_bc_tokens_hash ON iam_integrations.bureau_credito_tokens(token_hash);
CREATE INDEX idx_bc_tokens_expiracao ON iam_integrations.bureau_credito_tokens(data_expiracao);

-- Tabela para autorizações de consulta
CREATE TABLE IF NOT EXISTS iam_integrations.bureau_credito_autorizacoes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    identity_id UUID NOT NULL,
    tipo_consulta VARCHAR(50) NOT NULL,
    finalidade VARCHAR(100) NOT NULL,
    justificativa TEXT,
    data_autorizacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_validade TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'ativa',
    autorizado_por UUID,
    CONSTRAINT fk_bc_autorizacoes_identity FOREIGN KEY (identity_id) REFERENCES iam_integrations.bureau_credito_identities(id) ON DELETE CASCADE,
    CONSTRAINT ck_bc_autorizacoes_tipo CHECK (tipo_consulta IN ('basica', 'detalhada', 'completa', 'historico', 'score', 'alertas', 'monitoramento')),
    CONSTRAINT ck_bc_autorizacoes_status CHECK (status IN ('ativa', 'expirada', 'revogada', 'suspensa', 'pendente'))
);

-- Tabela para registro de consentimentos Bureau de Créditos
CREATE TABLE IF NOT EXISTS iam_integrations.bureau_credito_consent_log (
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
    CONSTRAINT fk_bc_consent_usuario FOREIGN KEY (usuario_id) REFERENCES iam_core.usuarios(id) ON DELETE CASCADE,
    CONSTRAINT fk_bc_consent_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_management.tenants(id),
    CONSTRAINT fk_bc_consent_identity FOREIGN KEY (identity_id) REFERENCES iam_integrations.bureau_credito_identities(id) ON DELETE CASCADE,
    CONSTRAINT ck_bc_consent_tipo CHECK (tipo_consentimento IN ('consulta_basica', 'consulta_detalhada', 'consulta_completa', 'historico_credito', 'monitoramento_continuo', 'compartilhamento_dados', 'uso_marketing', 'score_credito', 'biometria', 'verificacao_identidade'))
);

-- Tabela para registro de auditorias específicas de consultas ao Bureau
CREATE TABLE IF NOT EXISTS iam_integrations.bureau_credito_consulta_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    usuario_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    identity_id UUID,
    autorizacao_id UUID,
    documento_consultado VARCHAR(50) NOT NULL,
    tipo_documento VARCHAR(20) NOT NULL,
    tipo_consulta VARCHAR(50) NOT NULL,
    finalidade VARCHAR(100) NOT NULL,
    resultado VARCHAR(50) NOT NULL,
    detalhes JSONB,
    ip_origem VARCHAR(50),
    user_agent VARCHAR(500),
    data_consulta TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    duracao_ms INTEGER,
    hash_resposta VARCHAR(255),
    CONSTRAINT fk_bc_consulta_log_usuario FOREIGN KEY (usuario_id) REFERENCES iam_core.usuarios(id) ON DELETE CASCADE,
    CONSTRAINT fk_bc_consulta_log_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_management.tenants(id),
    CONSTRAINT fk_bc_consulta_log_identity FOREIGN KEY (identity_id) REFERENCES iam_integrations.bureau_credito_identities(id) ON DELETE SET NULL,
    CONSTRAINT fk_bc_consulta_log_autorizacao FOREIGN KEY (autorizacao_id) REFERENCES iam_integrations.bureau_credito_autorizacoes(id) ON DELETE SET NULL,
    CONSTRAINT ck_bc_consulta_tipo_doc CHECK (tipo_documento IN ('cpf', 'cnpj', 'bi', 'passaporte', 'nif', 'nuit', 'nic', 'outro')),
    CONSTRAINT ck_bc_consulta_tipo CHECK (tipo_consulta IN ('basica', 'detalhada', 'completa', 'historico', 'score', 'alertas', 'monitoramento')),
    CONSTRAINT ck_bc_consulta_resultado CHECK (resultado IN ('sucesso', 'sem_informacao', 'erro', 'acesso_negado', 'dados_incompletos', 'timeout'))
);

-- Tabela para histórico de alterações na integração
CREATE TABLE IF NOT EXISTS iam_integrations.bureau_credito_change_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    identity_id UUID,
    usuario_id UUID,
    tenant_id UUID NOT NULL,
    tipo_alteracao VARCHAR(50) NOT NULL,
    entidade VARCHAR(50) NOT NULL,
    id_entidade UUID NOT NULL,
    dados_anteriores JSONB,
    dados_novos JSONB,
    data_alteracao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    alterado_por UUID,
    ip_origem VARCHAR(50),
    CONSTRAINT fk_bc_change_log_tenant FOREIGN KEY (tenant_id) REFERENCES tenant_management.tenants(id),
    CONSTRAINT fk_bc_change_log_identity FOREIGN KEY (identity_id) REFERENCES iam_integrations.bureau_credito_identities(id) ON DELETE SET NULL,
    CONSTRAINT ck_bc_change_log_tipo CHECK (tipo_alteracao IN ('criacao', 'atualizacao', 'exclusao', 'ativacao', 'desativacao', 'vinculacao', 'desvinculacao', 'acesso_concedido', 'acesso_revogado')),
    CONSTRAINT ck_bc_change_log_entidade CHECK (entidade IN ('identity', 'token', 'autorizacao', 'consentimento'))
);

-- ==========================================================================
-- FUNÇÕES E PROCEDIMENTOS ARMAZENADOS
-- ==========================================================================

-- Função para vincular usuário IAM com Bureau de Créditos
CREATE OR REPLACE FUNCTION iam_integrations.vincular_usuario_bureau_credito(
    p_usuario_id UUID,
    p_tenant_id UUID,
    p_bc_usuario_id VARCHAR(255),
    p_bc_tenant_id VARCHAR(255),
    p_tipo_vinculo VARCHAR(50),
    p_nivel_acesso VARCHAR(30),
    p_detalhes_autorizacao JSONB DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    v_identity_id UUID;
    v_evento_id UUID;
BEGIN
    -- Verificar se o usuário já possui vínculo ativo
    SELECT id INTO v_identity_id
    FROM iam_integrations.bureau_credito_identities
    WHERE usuario_id = p_usuario_id AND bc_tenant_id = p_bc_tenant_id AND status_vinculo = 'ativo';
    
    IF FOUND THEN
        -- Atualizar vínculo existente
        UPDATE iam_integrations.bureau_credito_identities
        SET nivel_acesso = COALESCE(p_nivel_acesso, nivel_acesso),
            tipo_vinculo = COALESCE(p_tipo_vinculo, tipo_vinculo),
            detalhes_autorizacao = COALESCE(p_detalhes_autorizacao, detalhes_autorizacao),
            data_atualizacao = NOW()
        WHERE id = v_identity_id;
        
        -- Registrar alteração
        INSERT INTO iam_integrations.bureau_credito_change_log (
            identity_id, usuario_id, tenant_id, tipo_alteracao, entidade, 
            id_entidade, dados_novos
        ) VALUES (
            v_identity_id, p_usuario_id, p_tenant_id, 'atualizacao', 'identity',
            v_identity_id, jsonb_build_object(
                'nivel_acesso', COALESCE(p_nivel_acesso, nivel_acesso),
                'tipo_vinculo', COALESCE(p_tipo_vinculo, tipo_vinculo),
                'detalhes_autorizacao', p_detalhes_autorizacao
            )
        );
    ELSE
        -- Criar novo vínculo
        INSERT INTO iam_integrations.bureau_credito_identities (
            usuario_id, tenant_id, bc_usuario_id, bc_tenant_id,
            tipo_vinculo, nivel_acesso, detalhes_autorizacao
        ) VALUES (
            p_usuario_id, p_tenant_id, p_bc_usuario_id, p_bc_tenant_id,
            p_tipo_vinculo, COALESCE(p_nivel_acesso, 'consulta_basica'), p_detalhes_autorizacao
        )
        RETURNING id INTO v_identity_id;
        
        -- Registrar criação
        INSERT INTO iam_integrations.bureau_credito_change_log (
            identity_id, usuario_id, tenant_id, tipo_alteracao, entidade, 
            id_entidade, dados_novos
        ) VALUES (
            v_identity_id, p_usuario_id, p_tenant_id, 'criacao', 'identity',
            v_identity_id, jsonb_build_object(
                'nivel_acesso', COALESCE(p_nivel_acesso, 'consulta_basica'),
                'tipo_vinculo', p_tipo_vinculo,
                'detalhes_autorizacao', p_detalhes_autorizacao
            )
        );
    END IF;
    
    -- Registrar evento de auditoria
    SELECT iam_audit.registrar_evento_auditoria(
        p_tenant_id,
        p_usuario_id,
        NULL,
        'VINCULACAO_BUREAU_CREDITO',
        'INTEGRACAO',
        'alta',
        'Vínculo de usuário com Bureau de Créditos',
        jsonb_build_object(
            'bc_usuario_id', p_bc_usuario_id,
            'bc_tenant_id', p_bc_tenant_id,
            'tipo_vinculo', p_tipo_vinculo,
            'nivel_acesso', COALESCE(p_nivel_acesso, 'consulta_basica')
        ),
        'sucesso',
        NULL,
        NULL,
        'IAM_BUREAU_CREDITO',
        'usuario',
        p_usuario_id::TEXT
    ) INTO v_evento_id;
    
    RETURN v_identity_id;
END;
$$ LANGUAGE plpgsql;

-- Função para criar autorização de consulta
CREATE OR REPLACE FUNCTION iam_integrations.criar_autorizacao_consulta(
    p_identity_id UUID,
    p_tipo_consulta VARCHAR(50),
    p_finalidade VARCHAR(100),
    p_justificativa TEXT,
    p_duracao_dias INTEGER DEFAULT 30,
    p_autorizado_por UUID DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    v_autorizacao_id UUID;
    v_usuario_id UUID;
    v_tenant_id UUID;
BEGIN
    -- Obter informações do identity
    SELECT usuario_id, tenant_id INTO v_usuario_id, v_tenant_id
    FROM iam_integrations.bureau_credito_identities
    WHERE id = p_identity_id;
    
    -- Criar autorização
    INSERT INTO iam_integrations.bureau_credito_autorizacoes (
        identity_id,
        tipo_consulta,
        finalidade,
        justificativa,
        data_validade,
        autorizado_por
    ) VALUES (
        p_identity_id,
        p_tipo_consulta,
        p_finalidade,
        p_justificativa,
        NOW() + (p_duracao_dias * INTERVAL '1 day'),
        p_autorizado_por
    )
    RETURNING id INTO v_autorizacao_id;
    
    -- Registrar alteração
    INSERT INTO iam_integrations.bureau_credito_change_log (
        identity_id, usuario_id, tenant_id, tipo_alteracao, entidade, 
        id_entidade, dados_novos, alterado_por
    ) VALUES (
        p_identity_id, v_usuario_id, v_tenant_id, 'criacao', 'autorizacao',
        v_autorizacao_id, jsonb_build_object(
            'tipo_consulta', p_tipo_consulta,
            'finalidade', p_finalidade,
            'duracao_dias', p_duracao_dias
        ),
        p_autorizado_por
    );
    
    RETURN v_autorizacao_id;
END;
$$ LANGUAGE plpgsql;

-- Função para registrar consulta ao Bureau de Créditos
CREATE OR REPLACE FUNCTION iam_integrations.registrar_consulta_bureau(
    p_usuario_id UUID,
    p_tenant_id UUID,
    p_identity_id UUID,
    p_autorizacao_id UUID,
    p_documento_consultado VARCHAR(50),
    p_tipo_documento VARCHAR(20),
    p_tipo_consulta VARCHAR(50),
    p_finalidade VARCHAR(100),
    p_resultado VARCHAR(50),
    p_detalhes JSONB,
    p_ip_origem VARCHAR(50),
    p_user_agent VARCHAR(500),
    p_duracao_ms INTEGER,
    p_hash_resposta VARCHAR(255)
) RETURNS UUID AS $$
DECLARE
    v_consulta_id UUID;
    v_evento_id UUID;
BEGIN
    -- Registrar consulta
    INSERT INTO iam_integrations.bureau_credito_consulta_log (
        usuario_id,
        tenant_id,
        identity_id,
        autorizacao_id,
        documento_consultado,
        tipo_documento,
        tipo_consulta,
        finalidade,
        resultado,
        detalhes,
        ip_origem,
        user_agent,
        duracao_ms,
        hash_resposta
    ) VALUES (
        p_usuario_id,
        p_tenant_id,
        p_identity_id,
        p_autorizacao_id,
        p_documento_consultado,
        p_tipo_documento,
        p_tipo_consulta,
        p_finalidade,
        p_resultado,
        p_detalhes,
        p_ip_origem,
        p_user_agent,
        p_duracao_ms,
        p_hash_resposta
    )
    RETURNING id INTO v_consulta_id;
    
    -- Registrar evento de auditoria (alta severidade para consultas de bureau)
    SELECT iam_audit.registrar_evento_auditoria(
        p_tenant_id,
        p_usuario_id,
        NULL,
        'CONSULTA_BUREAU_CREDITO',
        'CONSULTA',
        'alta',
        'Consulta ao Bureau de Créditos - ' || p_tipo_consulta,
        jsonb_build_object(
            'documento_consultado', SUBSTRING(p_documento_consultado FROM 1 FOR 3) || '******',
            'tipo_documento', p_tipo_documento,
            'tipo_consulta', p_tipo_consulta,
            'finalidade', p_finalidade,
            'resultado', p_resultado
        ),
        p_resultado,
        p_ip_origem,
        p_user_agent,
        'BUREAU_CREDITO',
        'consulta',
        v_consulta_id::TEXT
    ) INTO v_evento_id;
    
    RETURN v_consulta_id;
END;
$$ LANGUAGE plpgsql;

-- ==========================================================================
-- VIEWS
-- ==========================================================================

-- View para relatório de integrações ativas com Bureau de Créditos
CREATE OR REPLACE VIEW iam_integrations.vw_bureau_credito_integrations AS
SELECT
    bci.id AS identity_id,
    u.id AS usuario_id,
    u.nome_completo,
    u.email,
    t.nome AS tenant_nome,
    t.slug AS tenant_slug,
    bci.bc_usuario_id,
    bci.bc_tenant_id,
    bci.tipo_vinculo,
    bci.nivel_acesso,
    bci.status_vinculo,
    bci.data_vinculacao,
    bci.data_atualizacao,
    (SELECT COUNT(*) FROM iam_integrations.bureau_credito_autorizacoes bca
     WHERE bca.identity_id = bci.id AND bca.status = 'ativa' AND bca.data_validade > NOW()) AS autorizacoes_ativas,
    (SELECT COUNT(*) FROM iam_integrations.bureau_credito_consulta_log bcl
     WHERE bcl.identity_id = bci.id) AS total_consultas,
    (SELECT COUNT(*) FROM iam_integrations.bureau_credito_consulta_log bcl
     WHERE bcl.identity_id = bci.id AND bcl.data_consulta > NOW() - INTERVAL '30 days') AS consultas_ultimos_30_dias
FROM
    iam_integrations.bureau_credito_identities bci
JOIN
    iam_core.usuarios u ON bci.usuario_id = u.id
JOIN
    tenant_management.tenants t ON bci.tenant_id = t.id
WHERE
    bci.status_vinculo = 'ativo';

-- View para histórico de consultas por usuário
CREATE OR REPLACE VIEW iam_integrations.vw_bureau_credito_consultas AS
SELECT
    bcl.id AS consulta_id,
    u.id AS usuario_id,
    u.nome_completo,
    u.email,
    t.nome AS tenant_nome,
    bcl.documento_consultado,
    bcl.tipo_documento,
    bcl.tipo_consulta,
    bcl.finalidade,
    bcl.resultado,
    bcl.data_consulta,
    bcl.duracao_ms,
    bci.tipo_vinculo,
    bci.nivel_acesso,
    bca.id AS autorizacao_id,
    bca.justificativa AS autorizacao_justificativa,
    bca.data_autorizacao,
    ua.nome_completo AS autorizado_por
FROM
    iam_integrations.bureau_credito_consulta_log bcl
JOIN
    iam_core.usuarios u ON bcl.usuario_id = u.id
JOIN
    tenant_management.tenants t ON bcl.tenant_id = t.id
LEFT JOIN
    iam_integrations.bureau_credito_identities bci ON bcl.identity_id = bci.id
LEFT JOIN
    iam_integrations.bureau_credito_autorizacoes bca ON bcl.autorizacao_id = bca.id
LEFT JOIN
    iam_core.usuarios ua ON bca.autorizado_por = ua.id;

-- ==========================================================================
-- PERMISSÕES
-- ==========================================================================

-- Criar permissões padrão para integração com Bureau de Créditos
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
    ('00000000-0000-0000-0000-000000000001', 'BUREAU_CREDITO_ADMIN', 'Administrar Bureau de Créditos', 'Permissão para administrar integrações com Bureau de Créditos', 'INTEGRACAO', 'todas', 'BUREAU_CREDITO', true),
    ('00000000-0000-0000-0000-000000000001', 'BUREAU_CREDITO_VINCULAR', 'Vincular Bureau de Créditos', 'Permissão para vincular contas com Bureau de Créditos', 'INTEGRACAO', 'criar', 'BUREAU_CREDITO_IDENTITY', true),
    ('00000000-0000-0000-0000-000000000001', 'BUREAU_CREDITO_AUTORIZAR', 'Autorizar Consultas', 'Permissão para autorizar consultas ao Bureau de Créditos', 'INTEGRACAO', 'criar', 'BUREAU_CREDITO_AUTORIZACAO', true),
    ('00000000-0000-0000-0000-000000000001', 'BUREAU_CREDITO_CONSULTAR', 'Realizar Consultas', 'Permissão para realizar consultas ao Bureau de Créditos', 'INTEGRACAO', 'executar', 'BUREAU_CREDITO_CONSULTA', true),
    ('00000000-0000-0000-0000-000000000001', 'BUREAU_CREDITO_AUDITORIA', 'Auditar Bureau de Créditos', 'Permissão para auditar consultas ao Bureau de Créditos', 'INTEGRACAO', 'ler', 'BUREAU_CREDITO_LOG', true)
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
    codigo LIKE 'BUREAU_CREDITO_%'
ON CONFLICT (perfil_id, permissao_id) DO NOTHING;

-- ==========================================================================
-- COMENTÁRIOS
-- ==========================================================================

COMMENT ON TABLE iam_integrations.bureau_credito_identities IS 'Tabela de mapeamento de identidades entre IAM e Bureau de Créditos';
COMMENT ON TABLE iam_integrations.bureau_credito_tokens IS 'Tokens de acesso para integração com Bureau de Créditos';
COMMENT ON TABLE iam_integrations.bureau_credito_autorizacoes IS 'Autorizações para consultas ao Bureau de Créditos';
COMMENT ON TABLE iam_integrations.bureau_credito_consent_log IS 'Registro de consentimentos para acesso ao Bureau de Créditos';
COMMENT ON TABLE iam_integrations.bureau_credito_consulta_log IS 'Log de consultas realizadas ao Bureau de Créditos';
COMMENT ON TABLE iam_integrations.bureau_credito_change_log IS 'Histórico de alterações na integração com Bureau de Créditos';

COMMENT ON FUNCTION iam_integrations.vincular_usuario_bureau_credito IS 'Função para vincular usuário do IAM com o Bureau de Créditos';
COMMENT ON FUNCTION iam_integrations.criar_autorizacao_consulta IS 'Função para criar autorização de consulta ao Bureau de Créditos';
COMMENT ON FUNCTION iam_integrations.registrar_consulta_bureau IS 'Função para registrar consulta realizada ao Bureau de Créditos';

COMMENT ON VIEW iam_integrations.vw_bureau_credito_integrations IS 'Visão consolidada de integrações ativas com Bureau de Créditos';
COMMENT ON VIEW iam_integrations.vw_bureau_credito_consultas IS 'Visão detalhada das consultas realizadas ao Bureau de Créditos';