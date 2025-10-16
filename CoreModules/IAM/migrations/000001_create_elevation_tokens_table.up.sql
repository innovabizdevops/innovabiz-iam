-- Migration: Criação das tabelas para o serviço de elevação de privilégios
-- Autor: Eduardo Jeremias
-- Projeto: INNOVABIZ IAM
-- Descrição: Este script cria as tabelas necessárias para o repositório
-- de tokens de elevação, incluindo tabelas de auditoria e logs conforme
-- as regulamentações e compliance dos mercados alvo (SADC, PALOP, BRICS).

-- Habilita extensões necessárias para garantir compliance e segurança
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Tabela principal de tokens de elevação
CREATE TABLE IF NOT EXISTS elevation_tokens (
    -- Informações principais do token
    id UUID PRIMARY KEY,
    user_id VARCHAR(128) NOT NULL,
    tenant_id VARCHAR(128) NOT NULL,
    market VARCHAR(64) NOT NULL, -- Mercado: angola, mozambique, brasil, etc.
    scopes TEXT[] NOT NULL,
    status VARCHAR(32) NOT NULL CHECK (status IN ('pending_approval', 'active', 'expired', 'revoked', 'denied')),
    justification TEXT NOT NULL,
    
    -- Campos temporais
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Campos de aprovação
    approved_by VARCHAR(128),
    approved_at TIMESTAMP WITH TIME ZONE,
    
    -- Campos de negação
    denied_by VARCHAR(128),
    denied_at TIMESTAMP WITH TIME ZONE,
    deny_reason TEXT,
    
    -- Campos de revogação
    revoked_by VARCHAR(128),
    revoked_at TIMESTAMP WITH TIME ZONE,
    revoke_reason TEXT,
    
    -- Campo de emergência (auto-aprovação)
    emergency BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Campos de utilização
    usage_count INTEGER NOT NULL DEFAULT 0,
    last_used_at TIMESTAMP WITH TIME ZONE,
    
    -- Índices para performance e auditoria
    CONSTRAINT elevation_tokens_status_check CHECK (status IN ('pending_approval', 'active', 'expired', 'revoked', 'denied')),
    CONSTRAINT elevation_tokens_approved_check CHECK ((status != 'active' AND approved_by IS NULL AND approved_at IS NULL) OR 
                                              (status = 'active' AND approved_by IS NOT NULL AND approved_at IS NOT NULL) OR
                                              (emergency = TRUE))
);

-- Comentários para documentação das colunas
COMMENT ON TABLE elevation_tokens IS 'Tokens de elevação de privilégios para o sistema MCP-IAM';
COMMENT ON COLUMN elevation_tokens.id IS 'Identificador único do token de elevação';
COMMENT ON COLUMN elevation_tokens.user_id IS 'ID do usuário solicitante da elevação';
COMMENT ON COLUMN elevation_tokens.tenant_id IS 'ID do tenant (empresa/organização)';
COMMENT ON COLUMN elevation_tokens.market IS 'Mercado de atuação (angola, mozambique, brasil, etc.)';
COMMENT ON COLUMN elevation_tokens.scopes IS 'Escopos de permissões solicitados';
COMMENT ON COLUMN elevation_tokens.status IS 'Status do token: pendente, ativo, expirado, revogado ou negado';
COMMENT ON COLUMN elevation_tokens.justification IS 'Justificativa para a solicitação de elevação';
COMMENT ON COLUMN elevation_tokens.created_at IS 'Data/hora da criação do token';
COMMENT ON COLUMN elevation_tokens.expires_at IS 'Data/hora de expiração do token';
COMMENT ON COLUMN elevation_tokens.approved_by IS 'ID do usuário que aprovou o token';
COMMENT ON COLUMN elevation_tokens.approved_at IS 'Data/hora da aprovação do token';
COMMENT ON COLUMN elevation_tokens.denied_by IS 'ID do usuário que negou o token';
COMMENT ON COLUMN elevation_tokens.denied_at IS 'Data/hora da negação do token';
COMMENT ON COLUMN elevation_tokens.deny_reason IS 'Motivo da negação do token';
COMMENT ON COLUMN elevation_tokens.revoked_by IS 'ID do usuário que revogou o token';
COMMENT ON COLUMN elevation_tokens.revoked_at IS 'Data/hora da revogação do token';
COMMENT ON COLUMN elevation_tokens.revoke_reason IS 'Motivo da revogação do token';
COMMENT ON COLUMN elevation_tokens.emergency IS 'Indica se o token foi criado em modo emergência (auto-aprovação)';
COMMENT ON COLUMN elevation_tokens.usage_count IS 'Contador de utilizações do token';
COMMENT ON COLUMN elevation_tokens.last_used_at IS 'Data/hora da última utilização do token';

-- Histórico de alterações dos tokens para auditoria completa
CREATE TABLE IF NOT EXISTS elevation_token_history (
    id BIGSERIAL PRIMARY KEY,
    token_id UUID NOT NULL REFERENCES elevation_tokens(id),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    action VARCHAR(32) NOT NULL, -- created, approved, denied, revoked, expired, used
    actor_id VARCHAR(128) NOT NULL,
    prev_status VARCHAR(32),
    new_status VARCHAR(32) NOT NULL,
    reason TEXT,
    metadata JSONB,
    CONSTRAINT elevation_token_history_action_check CHECK (
        action IN ('created', 'approved', 'denied', 'revoked', 'expired', 'used', 'updated')
    )
);

COMMENT ON TABLE elevation_token_history IS 'Histórico de alterações dos tokens de elevação para auditoria';
COMMENT ON COLUMN elevation_token_history.id IS 'Identificador único do registro de histórico';
COMMENT ON COLUMN elevation_token_history.token_id IS 'ID do token associado';
COMMENT ON COLUMN elevation_token_history.timestamp IS 'Data/hora do evento';
COMMENT ON COLUMN elevation_token_history.action IS 'Ação realizada: criação, aprovação, negação, revogação, expiração, uso';
COMMENT ON COLUMN elevation_token_history.actor_id IS 'ID do usuário que realizou a ação';
COMMENT ON COLUMN elevation_token_history.prev_status IS 'Status anterior do token';
COMMENT ON COLUMN elevation_token_history.new_status IS 'Novo status do token após a ação';
COMMENT ON COLUMN elevation_token_history.reason IS 'Motivo da ação, quando aplicável';
COMMENT ON COLUMN elevation_token_history.metadata IS 'Metadados adicionais sobre a ação (JSON)';

-- Índices para melhorar performance de consultas
CREATE INDEX elevation_tokens_user_tenant_idx ON elevation_tokens(user_id, tenant_id);
CREATE INDEX elevation_tokens_tenant_market_status_idx ON elevation_tokens(tenant_id, market, status);
CREATE INDEX elevation_tokens_status_idx ON elevation_tokens(status);
CREATE INDEX elevation_tokens_expires_at_idx ON elevation_tokens(expires_at);
CREATE INDEX elevation_tokens_created_at_idx ON elevation_tokens(created_at);

CREATE INDEX elevation_token_history_token_idx ON elevation_token_history(token_id);
CREATE INDEX elevation_token_history_timestamp_idx ON elevation_token_history(timestamp);
CREATE INDEX elevation_token_history_action_idx ON elevation_token_history(action);

-- Visão para facilitar relatórios de auditoria
CREATE OR REPLACE VIEW elevation_audit_view AS
SELECT 
    t.id as token_id,
    t.user_id,
    t.tenant_id,
    t.market,
    t.status,
    t.created_at,
    t.expires_at,
    t.approved_by,
    t.approved_at,
    t.denied_by,
    t.denied_at,
    t.revoked_by,
    t.revoked_at,
    t.emergency,
    t.usage_count,
    t.last_used_at,
    array_to_string(t.scopes, ', ') as scopes,
    t.justification,
    t.deny_reason,
    t.revoke_reason,
    CASE 
        WHEN t.emergency THEN 'Sim'
        ELSE 'Não'
    END as auto_aprovado,
    CASE
        WHEN t.status = 'active' AND t.expires_at > NOW() THEN 'Sim'
        ELSE 'Não'
    END as valido
FROM elevation_tokens t;

COMMENT ON VIEW elevation_audit_view IS 'Visão para relatórios de auditoria de tokens de elevação';

-- Função para marcar tokens expirados automaticamente
-- Esta função é chamada pelo job scheduler
CREATE OR REPLACE FUNCTION update_expired_tokens() RETURNS INTEGER AS $$
DECLARE
    updated_count INTEGER := 0;
    token_rec RECORD;
BEGIN
    FOR token_rec IN 
        SELECT id FROM elevation_tokens
        WHERE status = 'active' AND expires_at < NOW()
    LOOP
        -- Atualiza o status para expirado
        UPDATE elevation_tokens
        SET status = 'expired'
        WHERE id = token_rec.id;
        
        -- Registra no histórico
        INSERT INTO elevation_token_history (
            token_id, action, actor_id, prev_status, new_status, metadata
        ) VALUES (
            token_rec.id, 'expired', 'system', 'active', 'expired',
            jsonb_build_object('automatic', true, 'timestamp', NOW())
        );
        
        updated_count := updated_count + 1;
    END LOOP;
    
    RETURN updated_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_expired_tokens() IS 'Função para marcar automaticamente tokens expirados';

-- Gatilho para registrar automaticamente o uso de tokens no histórico
CREATE OR REPLACE FUNCTION log_token_usage() RETURNS TRIGGER AS $$
BEGIN
    -- Registra uso no histórico
    IF NEW.usage_count > OLD.usage_count THEN
        INSERT INTO elevation_token_history (
            token_id, action, actor_id, prev_status, new_status, metadata
        ) VALUES (
            NEW.id, 'used', 'system', OLD.status, NEW.status,
            jsonb_build_object('usage_count', NEW.usage_count, 'timestamp', NOW())
        );
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER elevation_tokens_usage_trigger
AFTER UPDATE OF usage_count ON elevation_tokens
FOR EACH ROW EXECUTE FUNCTION log_token_usage();

COMMENT ON TRIGGER elevation_tokens_usage_trigger ON elevation_tokens IS 'Gatilho para registrar uso de tokens no histórico';

-- Privilégios específicos para segurança
-- Os privilégios devem ser gerenciados através do sistema IAM central
-- e adaptados conforme necessário para cada ambiente (dev, staging, prod)