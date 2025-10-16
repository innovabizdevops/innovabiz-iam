-- Script para executar todos os testes e gerar relatórios

-- 1. Executar todos os testes
\i run_all_tests.sql

-- 2. Gerar relatório detalhado por categoria
SELECT test.generate_detailed_report('KB-01');
SELECT test.generate_detailed_report('PB-02');
SELECT test.generate_detailed_report('AF-03');
SELECT test.generate_detailed_report('BM-04');
SELECT test.generate_detailed_report('DT-05');
SELECT test.generate_detailed_report('INTEGRATION');
SELECT test.generate_detailed_report('COMPLIANCE');
SELECT test.generate_detailed_report('NOTIFICATIONS');
SELECT test.generate_detailed_report('PERFORMANCE');
SELECT test.generate_detailed_report('MONITORING');
SELECT test.generate_detailed_report('REPORTS');
SELECT test.generate_detailed_report('DISASTER_RECOVERY');
SELECT test.generate_detailed_report('ADVANCED_SECURITY');
SELECT test.generate_detailed_report('USABILITY');
SELECT test.generate_detailed_report('ACCESSIBILITY');
SELECT test.generate_detailed_report('COMPATIBILITY');
SELECT test.generate_detailed_report('BLOCKCHAIN_CRYPTO');

-- 3. Gerar relatório de métricas de performance
SELECT test.generate_performance_metrics();

-- 4. Gerar relatório de segurança
SELECT test.generate_security_report();

-- 5. Gerar relatório de conformidade
SELECT test.generate_compliance_report();

-- 6. Gerar relatório de usabilidade
SELECT test.generate_usability_report();

-- 7. Gerar relatório de acessibilidade
SELECT test.generate_accessibility_report();

-- 8. Gerar relatório de compatibilidade
SELECT test.generate_compatibility_report();

-- 9. Gerar relatório de blockchain e criptografia
SELECT test.generate_blockchain_crypto_report();

-- 10. Gerar relatório consolidado final
SELECT test.generate_final_consolidated_report();
