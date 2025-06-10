-- Script para executar todos os casos de teste do IAM Open X

-- 1. Executar testes de conhecimento
SELECT test.run_all_tests('KB-01');

-- 2. Executar testes de posse
SELECT test.run_all_tests('PB-02');

-- 3. Executar testes anti-fraude
SELECT test.run_all_tests('AF-03');

-- 4. Executar testes biométricos
SELECT test.run_all_tests('BM-04');

-- 5. Executar testes de dispositivos/tokens
SELECT test.run_all_tests('DT-05');

-- 6. Executar testes de integração
SELECT test.run_all_tests('INTEGRATION');

-- 7. Executar testes de conformidade
SELECT test.run_all_tests('COMPLIANCE');

-- 8. Executar testes de notificações
SELECT test.run_all_tests('NOTIFICATIONS');

-- 9. Executar testes de performance
SELECT test.run_all_tests('PERFORMANCE');

-- 10. Executar testes de monitoramento
SELECT test.run_all_tests('MONITORING');

-- 11. Executar testes de relatórios
SELECT test.run_all_tests('REPORTS');

-- 12. Executar testes de recuperação de desastres
SELECT test.run_all_tests('DISASTER_RECOVERY');

-- 13. Executar testes de segurança avançada
SELECT test.run_all_tests('ADVANCED_SECURITY');

-- 14. Executar testes de usabilidade
SELECT test.run_all_tests('USABILITY');

-- 15. Executar testes de acessibilidade
SELECT test.run_all_tests('ACCESSIBILITY');

-- 16. Executar testes de compatibilidade (partes 1 e 2)
SELECT test.run_all_tests('COMPATIBILITY');

-- 17. Executar testes de blockchain e criptografia
SELECT test.run_all_tests('BLOCKCHAIN_CRYPTO');

-- Gerar relatório consolidado de todos os testes
SELECT test.generate_consolidated_report();
