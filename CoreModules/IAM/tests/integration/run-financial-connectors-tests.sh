#!/bin/bash
# Script para executar testes dos conectores financeiros
# InnovaBiz DevOps Team, 2025

echo "==== INNOVABIZ - Testes de Conectores Financeiros ===="
echo "Iniciando testes de integração dos conectores financeiros..."

# Definir variáveis de ambiente para testes
export NODE_ENV=test
export TEST_MODE=integration
export OBSERVABILITY_LEVEL=debug

# Executar teste do conector Payment Gateway
echo "Executando testes do Payment Gateway Connector..."
npx jest payment-gateway-connector.test.ts --verbose

# Executar teste do conector Mobile Money
echo "Executando testes do Mobile Money Connector..."
npx jest mobile-money-connector.test.ts --verbose

# Verificar se os testes foram bem-sucedidos
if [ $? -eq 0 ]; then
  echo "✅ Todos os testes de conectores financeiros foram concluídos com sucesso!"
else
  echo "❌ Falha em alguns testes de conectores financeiros. Verifique os logs acima."
  exit 1
fi

echo "==== Testes de Conectores Financeiros Concluídos ===="
