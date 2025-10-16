-- ==============================================================================
-- Nome: V12__bureau_credito_integration.sql
-- Descrição: Migração para suporte à integração IAM com Bureau de Créditos
-- Autor: Equipa de Desenvolvimento INNOVABIZ
-- Data: 19/08/2025
-- ==============================================================================

-- Cria tabela para armazenar vínculos entre IAM e Bureau de Créditos
CREATE TABLE IF NOT EXISTS integration_bureau_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    external_tenant_id VARCHAR(255) NOT NULL,
    tipo_vinculo VARCHAR(50) NOT NULL,
    nivel_acesso VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDENTE_APROVACAO',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE,
    status_reason TEXT,
    details JSONB,
    
    CONSTRAINT fk_bureau_identity_usuario 
        FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE,
    CONSTRAINT fk_bureau_identity_tenant 
        FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT ck_bureau_tipo_vinculo 
        CHECK (tipo_vinculo IN ('CONSULTA', 'GESTAO', 'ADMIN', 'AUDITORIA')),
    CONSTRAINT ck_bureau_nivel_acesso 
        CHECK (nivel_acesso IN ('BASICO', 'INTERMEDIARIO', 'COMPLETO')),
    CONSTRAINT ck_bureau_status 
        CHECK (status IN ('ATIVO', 'INATIVO', 'SUSPENSO', 'PENDENTE_APROVACAO'))
);

-- Índices para otimização de consultas frequentes
CREATE INDEX IF NOT EXISTS idx_bureau_identity_usuario ON integration_bureau_identities(usuario_id);
CREATE INDEX IF NOT EXISTS idx_bureau_identity_tenant ON integration_bureau_identities(tenant_id);
CREATE INDEX IF NOT EXISTS idx_bureau_identity_status ON integration_bureau_identities(status);
CREATE INDEX IF NOT EXISTS idx_bureau_identity_external ON integration_bureau_identities(external_id);

-- Cria tabela para armazenar autorizações de consulta ao Bureau
CREATE TABLE IF NOT EXISTS bureau_autorizacoes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identity_id UUID NOT NULL,
    tipo_consulta VARCHAR(50) NOT NULL,
    finalidade VARCHAR(255) NOT NULL,
    justificativa TEXT NOT NULL,
    data_autorizacao TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    data_validade TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ATIVA',
    autorizado_por UUID NOT NULL,
    
    CONSTRAINT fk_autorizacao_identity 
        FOREIGN KEY (identity_id) REFERENCES integration_bureau_identities(id) ON DELETE CASCADE,
    CONSTRAINT fk_autorizacao_autorizado_por 
        FOREIGN KEY (autorizado_por) REFERENCES usuarios(id),
    CONSTRAINT ck_autorizacao_tipo_consulta 
        CHECK (tipo_consulta IN ('SIMPLES', 'DETALHADA', 'ANALISE_CREDITO', 'PREVENCAO_FRAUDE', 
                               'VALIDACAO_IDENTIDADE', 'CONFORMIDADE', 'MONITORAMENTO')),
    CONSTRAINT ck_autorizacao_status 
        CHECK (status IN ('ATIVA', 'EXPIRADA', 'REVOGADA', 'PENDENTE'))
);

-- Índices para otimização de consultas frequentes
CREATE INDEX IF NOT EXISTS idx_bureau_autorizacao_identity ON bureau_autorizacoes(identity_id);
CREATE INDEX IF NOT EXISTS idx_bureau_autorizacao_status ON bureau_autorizacoes(status);
CREATE INDEX IF NOT EXISTS idx_bureau_autorizacao_validade ON bureau_autorizacoes(data_validade);
CREATE INDEX IF NOT EXISTS idx_bureau_autorizacao_tipo ON bureau_autorizacoes(tipo_consulta);

-- Cria tabela para armazenar tokens de acesso ao Bureau
CREATE TABLE IF NOT EXISTS bureau_access_tokens (
    token_id VARCHAR(255) PRIMARY KEY,
    identity_id UUID NOT NULL,
    token_type VARCHAR(50) NOT NULL DEFAULT 'bearer',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    refresh_token VARCHAR(255),
    scopes TEXT[] NOT NULL,
    finalidade VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT fk_bureau_token_identity 
        FOREIGN KEY (identity_id) REFERENCES integration_bureau_identities(id) ON DELETE CASCADE
);

-- Índices para otimização de consultas frequentes
CREATE INDEX IF NOT EXISTS idx_bureau_token_identity ON bureau_access_tokens(identity_id);
CREATE INDEX IF NOT EXISTS idx_bureau_token_expires ON bureau_access_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_bureau_token_revoked ON bureau_access_tokens(revoked);

-- Cria view para facilitar consultas de autorizações válidas
CREATE OR REPLACE VIEW vw_bureau_autorizacoes_validas AS
SELECT a.*
FROM bureau_autorizacoes a
WHERE 
    a.status = 'ATIVA' AND
    a.data_validade > NOW();

-- Adicionar permissões específicas para Bureau de Créditos no sistema IAM
INSERT INTO permissions (code, name, description, resource_type, action_type, module, created_at)
VALUES
    ('bureau_credito:read', 'Consultar Bureau', 'Permite consultar vínculos do Bureau de Créditos', 'bureau_credito', 'read', 'bureau_credito', NOW()),
    ('bureau_credito:create_vinculo', 'Criar Vínculo Bureau', 'Permite criar vínculos com Bureau de Créditos', 'bureau_credito', 'create', 'bureau_credito', NOW()),
    ('bureau_credito:list', 'Listar Vínculos Bureau', 'Permite listar vínculos do Bureau de Créditos', 'bureau_credito', 'read', 'bureau_credito', NOW()),
    ('bureau_credito:list_tenant', 'Listar Vínculos Bureau por Tenant', 'Permite listar vínculos do Bureau por tenant', 'bureau_credito', 'read', 'bureau_credito', NOW()),
    ('bureau_credito:read_autorizacoes', 'Consultar Autorizações Bureau', 'Permite consultar autorizações de Bureau', 'bureau_credito', 'read', 'bureau_credito', NOW()),
    ('bureau_credito:verify', 'Verificar Vínculo Bureau', 'Permite verificar existência de vínculos', 'bureau_credito', 'read', 'bureau_credito', NOW()),
    ('bureau_credito:verify_autorizacao', 'Verificar Autorização Bureau', 'Permite verificar validade de autorizações', 'bureau_credito', 'read', 'bureau_credito', NOW()),
    ('bureau_credito:create_autorizacao', 'Criar Autorização Bureau', 'Permite criar autorizações de consulta', 'bureau_credito', 'create', 'bureau_credito', NOW()),
    ('bureau_credito:generate_token', 'Gerar Token Bureau', 'Permite gerar tokens de acesso ao Bureau', 'bureau_credito', 'create', 'bureau_credito', NOW()),
    ('bureau_credito:revoke_vinculo', 'Revogar Vínculo Bureau', 'Permite revogar vínculos com Bureau', 'bureau_credito', 'delete', 'bureau_credito', NOW());

-- Criar função para atualização automática de status de autorizações expiradas
CREATE OR REPLACE FUNCTION fn_update_expired_bureau_autorizacoes()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE bureau_autorizacoes
    SET status = 'EXPIRADA'
    WHERE status = 'ATIVA' AND data_validade < NOW();
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Criar trigger para execução diária da função
DROP TRIGGER IF EXISTS trig_update_expired_bureau_autorizacoes ON bureau_autorizacoes;
CREATE TRIGGER trig_update_expired_bureau_autorizacoes
AFTER INSERT OR UPDATE ON bureau_autorizacoes
EXECUTE FUNCTION fn_update_expired_bureau_autorizacoes();

-- Adicionar registro de evento de auditoria específico para Bureau de Créditos
INSERT INTO audit_event_types (code, name, description, severity, module, created_at)
VALUES
    ('BUREAU_CREATE_VINCULO', 'Criação de Vínculo Bureau', 'Criação de vínculo com Bureau de Créditos', 'ALTA', 'bureau_credito', NOW()),
    ('BUREAU_CREATE_AUTORIZACAO', 'Criação de Autorização Bureau', 'Criação de autorização de consulta ao Bureau', 'ALTA', 'bureau_credito', NOW()),
    ('BUREAU_GENERATE_TOKEN', 'Geração de Token Bureau', 'Geração de token de acesso ao Bureau', 'ALTA', 'bureau_credito', NOW()),
    ('BUREAU_REVOKE_VINCULO', 'Revogação de Vínculo Bureau', 'Revogação de vínculo com Bureau', 'ALTA', 'bureau_credito', NOW()),
    ('BUREAU_IDENTITY_ACCESS', 'Acesso a Identidade Bureau', 'Consulta de vínculo com Bureau', 'MEDIA', 'bureau_credito', NOW()),
    ('BUREAU_IDENTITIES_BY_USER_ACCESS', 'Listagem de Vínculos Bureau por Usuário', 'Listagem de vínculos por usuário', 'MEDIA', 'bureau_credito', NOW()),
    ('BUREAU_IDENTITIES_BY_TENANT_ACCESS', 'Listagem de Vínculos Bureau por Tenant', 'Listagem de vínculos por tenant', 'MEDIA', 'bureau_credito', NOW()),
    ('BUREAU_AUTORIZACOES_ACCESS', 'Acesso a Autorizações Bureau', 'Consulta de autorizações de Bureau', 'MEDIA', 'bureau_credito', NOW());

-- Criar função para revogação automática de todos os tokens quando um vínculo é revogado
CREATE OR REPLACE FUNCTION fn_revoke_tokens_on_identity_status_change()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'INATIVO' OR NEW.status = 'SUSPENSO' THEN
        UPDATE bureau_access_tokens
        SET revoked = TRUE, revoked_at = NOW()
        WHERE identity_id = NEW.id AND revoked = FALSE;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Criar trigger para revogar tokens quando status da identidade muda
DROP TRIGGER IF EXISTS trig_revoke_tokens_on_identity_status_change ON integration_bureau_identities;
CREATE TRIGGER trig_revoke_tokens_on_identity_status_change
AFTER UPDATE OF status ON integration_bureau_identities
FOR EACH ROW
WHEN (OLD.status <> NEW.status)
EXECUTE FUNCTION fn_revoke_tokens_on_identity_status_change();