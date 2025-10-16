-- Script de Definição de Roles para o Sistema IAM Open X
-- Versão: 1.0
-- Data: 15/05/2025

-- Criação do esquema para gerenciamento de roles
CREATE SCHEMA IF NOT EXISTS iam_access_control;

-- Função para criação segura de usuários
CREATE OR REPLACE FUNCTION iam_access_control.criar_usuario(
    p_username TEXT, 
    p_nivel_acesso INTEGER, 
    p_dominio TEXT DEFAULT NULL
)
RETURNS VOID AS $$
DECLARE
    v_senha TEXT;
BEGIN
    -- Geração de senha temporária segura
    v_senha := 'tmp_' || gen_salt('bf', 8) || '_' || md5(random()::text);
    
    -- Criação do usuário
    EXECUTE format('CREATE USER %I WITH PASSWORD %L', p_username, v_senha);
    
    -- Atribuição de role baseado no nível de acesso
    CASE p_nivel_acesso
        WHEN 1 THEN 
            EXECUTE format('GRANT role_acesso_basico TO %I', p_username);
        WHEN 2 THEN 
            EXECUTE format('GRANT role_acesso_operacional TO %I', p_username);
        WHEN 3 THEN 
            EXECUTE format('GRANT role_acesso_avancado TO %I', p_username);
        WHEN 4 THEN 
            EXECUTE format('GRANT role_admin_sistema TO %I', p_username);
        ELSE 
            RAISE EXCEPTION 'Nível de acesso inválido: %', p_nivel_acesso;
    END CASE;
    
    -- Atribuição de role por domínio, se especificado
    IF p_dominio IS NOT NULL THEN
        EXECUTE format('GRANT role_%s_access TO %I', lower(p_dominio), p_username);
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Criação de Roles Base
CREATE ROLE role_acesso_basico;
CREATE ROLE role_acesso_operacional;
CREATE ROLE role_acesso_avancado;
CREATE ROLE role_admin_sistema;

-- Roles por Domínio
CREATE ROLE role_saude_access;
CREATE ROLE role_seguros_access;
CREATE ROLE role_governo_access;

-- Definição de Permissões para Roles Base
-- Nível 1: Acesso Básico
GRANT SELECT ON ALL TABLES IN SCHEMA public TO role_acesso_basico;

-- Nível 2: Acesso Operacional
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO role_acesso_operacional;

-- Nível 3: Acesso Avançado
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO role_acesso_avancado;

-- Nível 4: Administrador de Sistema
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO role_admin_sistema;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO role_admin_sistema;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO role_admin_sistema;

-- Função de Auditoria de Acessos
CREATE OR REPLACE FUNCTION iam_access_control.auditar_acesso()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO iam_access_control.log_acessos (
        usuario, 
        acao, 
        tabela, 
        timestamp
    ) VALUES (
        current_user,
        TG_OP,
        TG_TABLE_NAME,
        current_timestamp
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Tabela de Log de Acessos
CREATE TABLE IF NOT EXISTS iam_access_control.log_acessos (
    id SERIAL PRIMARY KEY,
    usuario TEXT,
    acao TEXT,
    tabela TEXT,
    timestamp TIMESTAMP DEFAULT current_timestamp
);

-- Exemplo de aplicação da função de auditoria
-- Esta trigger pode ser aplicada em tabelas específicas conforme necessidade
CREATE OR REPLACE TRIGGER trigger_auditar_acesso
AFTER INSERT OR UPDATE OR DELETE ON public.usuarios
FOR EACH ROW EXECUTE FUNCTION iam_access_control.auditar_acesso();

-- Comentários para documentação
COMMENT ON SCHEMA iam_access_control IS 'Esquema de controle de acesso e auditoria do sistema IAM Open X';
COMMENT ON FUNCTION iam_access_control.criar_usuario IS 'Função segura para criação de usuários com níveis de acesso';
COMMENT ON FUNCTION iam_access_control.auditar_acesso IS 'Função de auditoria para registro de acessos e modificações';

-- Exemplo de uso da função de criação de usuário
-- SELECT iam_access_control.criar_usuario('joao.silva', 2, 'Saude');
-- SELECT iam_access_control.criar_usuario('maria.santos', 3, 'Governo');

-- Fim do script de definição de Roles