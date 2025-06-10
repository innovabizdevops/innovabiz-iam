-- Script de Teste para Validação das Políticas de Acesso IAM Open X
-- Versão: 1.0
-- Data: 15/05/2025

-- 1. Criação de Usuários de Teste
-- Nível 1: Acesso Básico
SELECT iam_access_control.criar_usuario('teste_basico', 1);

-- Nível 2: Acesso Operacional
SELECT iam_access_control.criar_usuario('teste_operacional', 2, 'Saude');

-- Nível 3: Acesso Avançado
SELECT iam_access_control.criar_usuario('teste_avancado', 3, 'Seguros');

-- Nível 4: Administrador de Sistema
SELECT iam_access_control.criar_usuario('teste_admin', 4);

-- 2. Teste de Permissões
-- Teste 1: Usuário Básico tentando inserir dados
SET ROLE role_acesso_basico;
INSERT INTO public.usuarios (nome, email) VALUES ('Teste', 'teste@exemplo.com');
-- Deve falhar com erro de permissão

-- Teste 2: Usuário Operacional inserindo dados
SET ROLE role_acesso_operacional;
INSERT INTO public.usuarios (nome, email) VALUES ('Teste Operacional', 'operacional@exemplo.com');
-- Deve ter sucesso

-- Teste 3: Usuário Avançado alterando sequência
SET ROLE role_acesso_avancado;
SELECT setval('public.usuarios_id_seq', 100);
-- Deve ter sucesso

-- 3. Teste de Auditoria
-- Verificar logs de acesso
SELECT * FROM iam_access_control.log_acessos 
WHERE usuario IN ('teste_basico', 'teste_operacional', 'teste_avancado', 'teste_admin')
ORDER BY timestamp DESC;

-- 4. Teste de Segregação de Funções
-- Usuário de Saúde tentando acessar dados de Seguros
SET ROLE role_saude_access;
SELECT * FROM public.usuarios WHERE dominio = 'Seguros';
-- Deve falhar com erro de permissão

-- 5. Teste de MFA (Multi-Fator)
-- Verificar se usuários de nível 3 e 4 precisam de MFA
SELECT rolname, rolconfig->>'mfa_required' as mfa_required
FROM pg_roles 
WHERE rolname IN ('role_acesso_avancado', 'role_admin_sistema');
-- Deve retornar true para ambos

-- 6. Teste de Expiração de Senha
-- Verificar se a senha expira após 90 dias
SELECT rolname, rolconfig->>'password_expiration' as password_expiration
FROM pg_roles 
WHERE rolname IN ('role_acesso_basico', 'role_acesso_operacional');
-- Deve retornar 90 dias

-- 7. Teste de Histórico de Senhas
-- Verificar se as senhas não podem ser reutilizadas
SELECT rolname, rolconfig->>'password_history_size' as password_history_size
FROM pg_roles 
WHERE rolname IN ('role_acesso_basico', 'role_acesso_operacional');
-- Deve retornar 10

-- 8. Teste de Limpeza (após testes)
-- Resetar para o superusuário
RESET ROLE;

-- Limpar usuários de teste
DROP USER IF EXISTS teste_basico;
DROP USER IF EXISTS teste_operacional;
DROP USER IF EXISTS teste_avancado;
DROP USER IF EXISTS teste_admin;

-- Limpar logs de teste
DELETE FROM iam_access_control.log_acessos 
WHERE usuario IN ('teste_basico', 'teste_operacional', 'teste_avancado', 'teste_admin');

-- Fim dos testes de validação de políticas de acesso