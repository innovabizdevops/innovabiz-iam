#!/bin/bash

# Executar todos os testes e gerar relatórios
psql -f execute_tests_and_generate_reports.sql

# Verificar status dos testes
echo "\nStatus dos Testes:"
psql -c "SELECT test.get_test_status();"

# Verificar métricas de performance
echo "\nMétricas de Performance:"
psql -c "SELECT test.get_performance_metrics();"

# Verificar métricas de segurança
echo "\nMétricas de Segurança:"
psql -c "SELECT test.get_security_metrics();"

# Verificar métricas de conformidade
echo "\nMétricas de Conformidade:"
psql -c "SELECT test.get_compliance_metrics();"

# Verificar métricas de usabilidade
echo "\nMétricas de Usabilidade:"
psql -c "SELECT test.get_usability_metrics();"

# Verificar métricas de acessibilidade
echo "\nMétricas de Acessibilidade:"
psql -c "SELECT test.get_accessibility_metrics();"

# Verificar métricas de compatibilidade
echo "\nMétricas de Compatibilidade:"
psql -c "SELECT test.get_compatibility_metrics();"

# Verificar métricas de blockchain e criptografia
echo "\nMétricas de Blockchain e Criptografia:"
psql -c "SELECT test.get_blockchain_crypto_metrics();"

# Gerar relatório final
psql -c "SELECT test.generate_final_report();"
