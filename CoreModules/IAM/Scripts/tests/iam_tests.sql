-- ====================================================================================
-- Test Suite: iam (Identity and Access Management)
-- Description: Testa o módulo de gestão de utilizadores, papéis e permissões.
-- ====================================================================================

BEGIN;

CREATE EXTENSION IF NOT EXISTS pgtap;

SELECT plan(11);

-- 1. Testes de Estrutura
SELECT has_schema('iam', 'O schema iam deve existir');
SELECT has_table('iam', 'users', 'A tabela users deve existir');
SELECT has_function('iam', 'register_user', 'A função register_user deve existir');
SELECT has_function('iam', 'authenticate_user', 'A função authenticate_user deve existir');
SELECT has_function('iam', 'check_permission', 'A função check_permission deve existir');

-- 2. Testes de Registo de Utilizador

-- Teste 6: Registar um utilizador com sucesso
SELECT lives_ok(
    $$ SELECT iam.register_user('testuser', 'test@innovabiz.com', 'StrongPassword123') $$,
    'Deve ser possível registar um novo utilizador.'
);

-- Teste 7: Tentar registar um utilizador com username duplicado
SELECT throws_ok(
    $$ SELECT iam.register_user('testuser', 'another@innovabiz.com', 'password') $$,
    'P0001',
    'Nome de utilizador já existe.',
    'Não deve ser possível registar um utilizador com username duplicado.'
);

-- 3. Testes de Autenticação

-- Teste 8: Autenticar um utilizador com sucesso
SELECT is_not_null(
    (SELECT iam.authenticate_user('testuser', 'StrongPassword123')),
    'A autenticação com credenciais corretas deve retornar um user_id.'
);

-- Teste 9: Falhar a autenticação com senha incorreta
SELECT is_null(
    (SELECT iam.authenticate_user('testuser', 'WrongPassword')),
    'A autenticação com senha incorreta deve retornar NULL.'
);

-- 4. Testes de Controlo de Acesso (RBAC)

-- Setup para o teste de permissões
CREATE TEMP TABLE test_rbac AS
SELECT 
    (SELECT user_id FROM iam.users WHERE username = 'testuser') AS user_id,
    (SELECT role_id FROM iam.roles WHERE role_name = 'merchant_admin') AS role_id,
    (SELECT permission_id FROM iam.permissions WHERE permission_name = 'transactions:read') AS permission_id;

-- Associar a permissão 'transactions:read' ao papel 'merchant_admin'
INSERT INTO iam.role_permissions (role_id, permission_id) VALUES ((SELECT role_id FROM test_rbac), (SELECT permission_id FROM test_rbac));

-- Atribuir o papel 'merchant_admin' ao 'testuser'
INSERT INTO iam.user_roles (user_id, role_id) VALUES ((SELECT user_id FROM test_rbac), (SELECT role_id FROM test_rbac));

-- Teste 10: Verificar uma permissão que o utilizador tem
SELECT ok(
    (SELECT iam.check_permission((SELECT user_id FROM test_rbac), 'transactions:read')),
    'O utilizador deve ter a permissão transactions:read.'
);

-- Teste 11: Verificar uma permissão que o utilizador NÃO tem
SELECT ok(
    NOT (SELECT iam.check_permission((SELECT user_id FROM test_rbac), 'users:delete')),
    'O utilizador NÃO deve ter a permissão users:delete.'
);


SELECT * FROM finish();

ROLLBACK;