# Script para executar testes dos conectores financeiros
# InnovaBiz DevOps Team, 2025

Write-Host "==== INNOVABIZ - Testes de Conectores Financeiros ====" -ForegroundColor Cyan
Write-Host "Iniciando testes de integração dos conectores financeiros..." -ForegroundColor Cyan

# Definir variáveis de ambiente para testes
$env:NODE_ENV = "test"
$env:TEST_MODE = "integration"
$env:OBSERVABILITY_LEVEL = "debug"

# Executar teste do conector Payment Gateway
Write-Host "Executando testes do Payment Gateway Connector..." -ForegroundColor Yellow
npx jest payment-gateway-connector.test.ts --verbose

# Verificar resultado do primeiro teste
$paymentGatewayResult = $LASTEXITCODE

# Executar teste do conector Mobile Money
Write-Host "Executando testes do Mobile Money Connector..." -ForegroundColor Yellow
npx jest mobile-money-connector.test.ts --verbose

# Verificar resultado do segundo teste
$mobileMoneyResult = $LASTEXITCODE

# Verificar se os testes foram bem-sucedidos
if (($paymentGatewayResult -eq 0) -and ($mobileMoneyResult -eq 0)) {
  Write-Host "✅ Todos os testes de conectores financeiros foram concluídos com sucesso!" -ForegroundColor Green
} else {
  Write-Host "❌ Falha em alguns testes de conectores financeiros. Verifique os logs acima." -ForegroundColor Red
  exit 1
}

Write-Host "==== Testes de Conectores Financeiros Concluídos ====" -ForegroundColor Cyan
