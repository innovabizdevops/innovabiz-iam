-- Script para executar todos os testes de blockchain e criptografia

-- 1. Executar testes básicos
SELECT test.run_test_group('BASIC_CRYPTO');

-- 2. Executar testes de assinatura
SELECT test.run_test_group('SIGNATURE_TESTS');

-- 3. Executar testes de hash
SELECT test.run_test_group('HASH_TESTS');

-- 4. Executar testes de blockchain
SELECT test.run_test_group('BLOCKCHAIN_TESTS');

-- 5. Executar testes de smart contracts
SELECT test.run_test_group('SMART_CONTRACT_TESTS');

-- 6. Executar testes de tokens
SELECT test.run_test_group('TOKEN_TESTS');

-- 7. Executar testes de carteiras
SELECT test.run_test_group('WALLET_TESTS');

-- 8. Executar testes de segurança
SELECT test.run_test_group('SECURITY_TESTS');

-- 9. Executar testes de atualização
SELECT test.run_test_group('UPDATE_TESTS');

-- 10. Executar testes de integração
SELECT test.run_test_group('INTEGRATION_TESTS');

-- 11. Gerar relatório consolidado
SELECT test.generate_consolidated_report('BLOCKCHAIN_CRYPTO');
