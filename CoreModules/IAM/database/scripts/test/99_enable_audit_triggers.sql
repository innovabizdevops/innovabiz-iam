-- Reativa os gatilhos de auditoria após a inserção dos dados de teste

-- Reativar gatilhos nas tabelas principais
ALTER TABLE iam.organizations ENABLE TRIGGER ALL;
ALTER TABLE iam.users ENABLE TRIGGER ALL;
ALTER TABLE iam.roles ENABLE TRIGGER ALL;
ALTER TABLE iam.permissions ENABLE TRIGGER ALL;
ALTER TABLE iam.user_roles ENABLE TRIGGER ALL;
ALTER TABLE iam.role_permissions ENABLE TRIGGER ALL;
ALTER TABLE iam.security_policies ENABLE TRIGGER ALL;

-- Mensagem de confirmação
DO $$
BEGIN
    RAISE NOTICE 'Gatilhos de auditoria reativados com sucesso';
END
$$;
