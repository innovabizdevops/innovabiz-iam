-- INNOVABIZ - IAM Metadata Enhancement
-- Author: Eduardo Jeremias
-- Date: 19/05/2025
-- Version: 1.0
-- Description: Script para adicionar metadados, comentários e otimizações às tabelas do IAM

-- Configurar caminho de busca
SET search_path TO iam, public;

-- =============================================
-- 1. Comentários nas Tabelas (se ainda não existirem)
-- =============================================

-- Tabela de Organizações
COMMENT ON TABLE organizations IS 'Armazena informações sobre as organizações que utilizam a plataforma INNOVABIZ, incluindo configurações específicas e metadados.';

-- Tabela de Usuários
COMMENT ON TABLE users IS 'Contém as informações dos usuários do sistema, incluindo credenciais, status e preferências.';

-- Tabela de Funções
COMMENT ON TABLE roles IS 'Define os papéis que podem ser atribuídos aos usuários, agrupando conjuntos de permissões.';

-- Tabela de Permissões
COMMENT ON TABLE permissions IS 'Lista todas as permissões disponíveis no sistema, que podem ser associadas a funções.';

-- Tabela de Atribuição de Funções a Usuários
COMMENT ON TABLE user_roles IS 'Relaciona usuários a funções, permitindo herdar permissões.';

-- Tabela de Sessões
COMMENT ON TABLE sessions IS 'Registra as sessões ativas dos usuários, incluindo tokens e informações de autenticação.';

-- Tabela de Logs de Auditoria
COMMENT ON TABLE audit_logs IS 'Registra todas as ações significativas realizadas no sistema para fins de auditoria e conformidade.';

-- Tabela de Políticas de Segurança
COMMENT ON TABLE security_policies IS 'Armazena as políticas de segurança configuráveis para cada organização.';

-- Tabela de Frameworks Regulatórios
COMMENT ON TABLE regulatory_frameworks IS 'Lista os frameworks regulatórios suportados pelo sistema (ex: GDPR, LGPD, HIPAA).';

-- Tabela de Validadores de Conformidade
COMMENT ON TABLE compliance_validators IS 'Armazena os validadores de conformidade disponíveis para verificação automática de requisitos regulatórios.';

-- =============================================
-- 2. Comentários nas Colunas
-- =============================================

-- Tabela organizations
COMMENT ON COLUMN organizations.id IS 'Identificador único da organização (UUIDv4).';
COMMENT ON COLUMN organizations.name IS 'Nome completo da organização.';
COMMENT ON COLUMN organizations.code IS 'Código único da organização (usado para referência).';
COMMENT ON COLUMN organizations.industry IS 'Setor de atuação da organização.';
COMMENT ON COLUMN organizations.sector IS 'Segmento específico dentro do setor.';
COMMENT ON COLUMN organizations.country_code IS 'Código do país da organização (ISO 3166-1 alpha-2).';
COMMENT ON COLUMN organizations.region_code IS 'Código da região/estado da organização.';
COMMENT ON COLUMN organizations.created_at IS 'Data e hora de criação do registro.';
COMMENT ON COLUMN organizations.updated_at IS 'Data e hora da última atualização do registro.';
COMMENT ON COLUMN organizations.is_active IS 'Indica se a organização está ativa no sistema.';
COMMENT ON COLUMN organizations.settings IS 'Configurações específicas da organização em formato JSON.';
COMMENT ON COLUMN organizations.compliance_settings IS 'Configurações de conformidade específicas da organização.';
COMMENT ON COLUMN organizations.metadata IS 'Metadados adicionais da organização.';

-- Tabela users
COMMENT ON COLUMN users.id IS 'Identificador único do usuário (UUIDv4).';
COMMENT ON COLUMN users.organization_id IS 'Referência à organização à qual o usuário pertence.';
COMMENT ON COLUMN users.username IS 'Nome de usuário único para login.';
COMMENT ON COLUMN users.email IS 'Endereço de e-mail do usuário (deve ser único).';
COMMENT ON COLUMN users.full_name IS 'Nome completo do usuário.';
COMMENT ON COLUMN users.password_hash IS 'Hash da senha do usuário (usando criptografia forte).';
COMMENT ON COLUMN users.status IS 'Status atual do usuário (active, inactive, suspended, locked).';
COMMENT ON COLUMN users.created_at IS 'Data e hora de criação do registro.';
COMMENT ON COLUMN users.updated_at IS 'Data e hora da última atualização do registro.';
COMMENT ON COLUMN users.last_login IS 'Data e hora do último login bem-sucedido.';
COMMENT ON COLUMN users.preferences IS 'Preferências do usuário em formato JSON.';
COMMENT ON COLUMN users.metadata IS 'Metadados adicionais do usuário.';

-- Adicione comentários para as demais tabelas seguindo o mesmo padrão...

-- =============================================
-- 3. Índices Adicionais para Melhor Desempenho
-- =============================================

-- Índices para a tabela audit_logs (otimização para consultas comuns)
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp_org ON audit_logs(organization_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_action ON audit_logs(user_id, action) WHERE user_id IS NOT NULL;

-- Índices para a tabela user_roles (otimização para verificações de permissão)
CREATE INDEX IF NOT EXISTS idx_user_roles_user_org ON user_roles(user_id, organization_id) WHERE is_active = true;

-- Índices para a tabela sessions (otimização para limpeza de sessões expiradas)
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at_org ON sessions(organization_id, expires_at) WHERE is_active = true;

-- =============================================
-- 4. Restrições Adicionais
-- =============================================

-- Garantir que o e-mail do usuário seja válido
ALTER TABLE users ADD CONSTRAINT users_email_check 
    CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');

-- Garantir que o código da organização seja em maiúsculas e sem espaços
ALTER TABLE organizations ADD CONSTRAINT organizations_code_check 
    CHECK (code = upper(trim(both from code)));

-- Garantir que o nome de usuário não contenha caracteres especiais
ALTER TABLE users ADD CONSTRAINT users_username_check 
    CHECK (username ~* '^[a-z0-9_]+$');

-- =============================================
-- 5. Políticas de Retenção de Dados
-- =============================================

-- Criar partições para a tabela audit_logs (se ainda não existirem)
-- Nota: A implementação de partições pode variar dependendo da versão do PostgreSQL

-- Exemplo de função para limpar logs antigos (mais de 1 ano)
CREATE OR REPLACE FUNCTION cleanup_old_audit_logs()
RETURNS void AS $$
BEGIN
    DELETE FROM audit_logs 
    WHERE timestamp < (NOW() - INTERVAL '1 year');
    
    RAISE NOTICE 'Logs de auditoria antigos foram limpos com sucesso';
EXCEPTION
    WHEN OTHERS THEN
        RAISE WARNING 'Erro ao limpar logs de auditoria: %', SQLERRM;
END;
$$ LANGUAGE plpgsql;

-- Agendar a limpeza mensal (requer pg_cron ou agendador externo)
-- Exemplo para pg_cron:
-- SELECT cron.schedule('0 0 1 * *', 'SELECT cleanup_old_audit_logs()');

-- =============================================
-- 6. Visualizações Úteis
-- =============================================

-- Visão para listar usuários com suas funções
CREATE OR REPLACE VIEW vw_user_roles AS
SELECT 
    u.id AS user_id,
    u.username,
    u.email,
    u.full_name,
    u.status AS user_status,
    r.id AS role_id,
    r.name AS role_name,
    r.description AS role_description,
    ur.granted_at,
    ur.expires_at,
    ur.is_active AS role_assignment_active
FROM 
    users u
    JOIN user_roles ur ON u.id = ur.user_id
    JOIN roles r ON ur.role_id = r.id;

COMMENT ON VIEW vw_user_roles IS 'Visão que lista todos os usuários com suas respectivas funções atribuídas.';

-- Visão para auditoria de segurança
CREATE OR REPLACE VIEW vw_security_audit AS
SELECT 
    al.id,
    al.timestamp,
    o.name AS organization_name,
    u.username,
    u.email,
    al.action,
    al.resource_type,
    al.resource_id,
    al.status,
    al.ip_address,
    al.details
FROM 
    audit_logs al
    LEFT JOIN organizations o ON al.organization_id = o.id
    LEFT JOIN users u ON al.user_id = u.id
ORDER BY 
    al.timestamp DESC;

COMMENT ON VIEW vw_security_audit IS 'Visão para auditoria de segurança, mostrando ações realizadas no sistema.';

-- =============================================
-- 7. Funções Úteis
-- =============================================

-- Função para verificar se um usuário tem uma permissão específica
CREATE OR REPLACE FUNCTION has_permission(
    p_user_id UUID,
    p_permission_code VARCHAR(255)
) 
RETURNS BOOLEAN AS $$
DECLARE
    v_has_permission BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1 
        FROM user_roles ur
        JOIN roles r ON ur.role_id = r.id
        JOIN jsonb_array_elements(r.permissions) AS p(permission) 
        WHERE ur.user_id = p_user_id
        AND ur.is_active = true
        AND p.permission->>'code' = p_permission_code
        AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
    ) INTO v_has_permission;
    
    RETURN v_has_permission;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION has_permission IS 'Verifica se um usuário tem uma permissão específica com base em suas funções ativas.';

-- =============================================
-- 8. Permissões
-- =============================================

-- Garantir que o proprietário do esquema seja o usuário correto
ALTER SCHEMA iam OWNER TO innovabiz_admin;

-- Conceder permissões mínimas necessárias
GRANT USAGE ON SCHEMA iam TO innovabiz_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA iam TO innovabiz_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA iam TO innovabiz_app;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA iam TO innovabiz_app;

-- Conceder permissões de leitura para relatórios
GRANT SELECT ON ALL TABLES IN SCHEMA iam TO innovabiz_readonly;
GRANT SELECT ON ALL SEQUENCES IN SCHEMA iam TO innovabiz_readonly;

-- =============================================
-- 9. Finalização
-- =============================================

-- Registrar a execução deste script na tabela de auditoria
INSERT INTO audit_logs (
    id,
    organization_id,
    user_id,
    action,
    resource_type,
    resource_id,
    status,
    details
) VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111', -- ID da organização padrão
    '22222222-2222-2222-2222-222222222222', -- ID do usuário admin
    'execute',
    'schema',
    'iam',
    'success',
    '{"script": "01_add_metadata_to_iam_tables.sql", "version": "1.0", "description": "Adição de metadados, índices e otimizações ao esquema IAM"}'
);

-- Mensagem de conclusão
DO $$
BEGIN
    RAISE NOTICE 'Metadados, índices e otimizações adicionados com sucesso ao esquema IAM.';
END
$$;
