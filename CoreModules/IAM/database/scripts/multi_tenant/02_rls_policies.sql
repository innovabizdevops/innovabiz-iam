-- INNOVABIZ - IAM Multi-Tenant RLS Policies
-- Author: Eduardo Jeremias
-- Date: 09/05/2025
-- Version: 1.0
-- Description: Implementação de políticas de RLS para isolamento multi-tenant avançado

-- Definição de funções auxiliares para política multi-tenant
CREATE OR REPLACE FUNCTION iam.get_current_tenant_id()
RETURNS UUID AS $$
DECLARE
    tenant_id UUID;
BEGIN
    -- Primeiro tenta obter de variável de sessão
    tenant_id := current_setting('app.current_tenant_id', true)::UUID;
    
    -- Se não encontrar, verifica em current_setting com fallback para formato texto
    IF tenant_id IS NULL THEN
        tenant_id := current_setting('app.current_tenant_id', true);
    END IF;
    
    -- Ainda não encontrou, verifica em contexto JWT se disponível
    IF tenant_id IS NULL AND current_setting('request.jwt.claim.tenant_id', true) IS NOT NULL THEN
        tenant_id := current_setting('request.jwt.claim.tenant_id', true)::UUID;
    END IF;
    
    RETURN tenant_id;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Função para verificar se usuário atual é super admin
CREATE OR REPLACE FUNCTION iam.is_super_admin()
RETURNS BOOLEAN AS $$
DECLARE
    is_admin BOOLEAN;
BEGIN
    -- Verifica se é admin por claim JWT
    is_admin := current_setting('request.jwt.claim.is_super_admin', true)::BOOLEAN;
    RETURN COALESCE(is_admin, FALSE);
EXCEPTION
    WHEN OTHERS THEN
        RETURN FALSE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Aplicação das políticas RLS para tabelas principais

-- Policy para Organizations
ALTER TABLE iam.organizations ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON iam.organizations
    USING (id = iam.get_current_tenant_id() OR iam.is_super_admin());

-- Policy para Users
ALTER TABLE iam.users ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON iam.users
    USING (organization_id = iam.get_current_tenant_id() OR iam.is_super_admin());

-- Policy para Roles
ALTER TABLE iam.roles ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON iam.roles
    USING (organization_id = iam.get_current_tenant_id() OR iam.is_super_admin());

-- Policy para Permissions
ALTER TABLE iam.permissions ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON iam.permissions
    USING (organization_id = iam.get_current_tenant_id() OR iam.is_super_admin());

-- Policy para User_Role_Assignments
ALTER TABLE iam.user_role_assignments ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON iam.user_role_assignments
    USING (
        organization_id = iam.get_current_tenant_id() OR 
        iam.is_super_admin()
    );

-- Policy para Audit Logs
ALTER TABLE iam.audit_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON iam.audit_logs
    USING (
        organization_id = iam.get_current_tenant_id() OR 
        iam.is_super_admin() OR
        (organization_id IS NULL AND iam.is_super_admin())
    );

-- Configuração de funções para ativação automática de tenant context
CREATE OR REPLACE FUNCTION iam.set_tenant_context() 
RETURNS TRIGGER AS $$
BEGIN
    PERFORM set_config('app.current_tenant_id', NEW.organization_id::text, false);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Log de alterações de políticas RLS para auditoria
INSERT INTO iam.audit_logs (
    action, 
    entity_type, 
    entity_id, 
    description,
    metadata
) VALUES (
    'SYSTEM_CONFIG', 
    'RLS_POLICY', 
    NULL, 
    'RLS policies installed or updated for multi-tenant isolation',
    jsonb_build_object(
        'tables', jsonb_build_array(
            'organizations', 
            'users', 
            'roles', 
            'permissions', 
            'user_role_assignments',
            'audit_logs'
        ),
        'policy_type', 'tenant_isolation',
        'version', '1.0'
    )
);
