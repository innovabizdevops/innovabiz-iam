-- Desativa temporariamente os gatilhos de auditoria
-- para permitir a inserção de dados de teste

-- Desativar gatilhos nas tabelas principais
ALTER TABLE iam.organizations DISABLE TRIGGER ALL;
ALTER TABLE iam.users DISABLE TRIGGER ALL;
ALTER TABLE iam.roles DISABLE TRIGGER ALL;
ALTER TABLE iam.permissions DISABLE TRIGGER ALL;
ALTER TABLE iam.user_roles DISABLE TRIGGER ALL;
ALTER TABLE iam.role_permissions DISABLE TRIGGER ALL;
ALTER TABLE iam.security_policies DISABLE TRIGGER ALL;

-- Mensagem de confirmação
DO $$
BEGIN
    RAISE NOTICE 'Gatilhos de auditoria desativados com sucesso';
END
$$;
