# Script de Instalação do Módulo IAM - Passo a Passo
# Data: 19/05/2025
# Descrição: Instala o módulo IAM no banco de dados PostgreSQL em um contêiner Docker

# Configurações
$dbName = "innovabiz_iam"
$dbUser = "postgres"
$dbPassword = "postgres"
$dbHost = "localhost"
$dbPort = "5432"

# Função para executar um script SQL e verificar erros
function Execute-SqlScript {
    param (
        [string]$scriptPath,
        [string]$scriptName
    )
    
    Write-Host "Executando $scriptName..." -ForegroundColor Cyan
    
    # Verifica se o arquivo existe
    if (-not (Test-Path $scriptPath)) {
        Write-Host "Arquivo não encontrado: $scriptPath" -ForegroundColor Red
        return $false
    }
    
    # Executa o script via docker exec
    $command = "docker exec -i integration-postgres-1 psql -U $dbUser -d $dbName -f $scriptPath"
    
    try {
        $output = Invoke-Expression $command 2>&1 | Out-String
        
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Erro ao executar $scriptName" -ForegroundColor Red
            Write-Host "Saída: $output" -ForegroundColor Red
            return $false
        }
        
        Write-Host "$scriptName executado com sucesso" -ForegroundColor Green
        Write-Host $output -ForegroundColor Gray
        return $true
    }
    catch {
        Write-Host ("Exceção ao executar {0}: {1}" -f $scriptName, $_) -ForegroundColor Red
        return $false
    }
}

# Verificar se o contêiner PostgreSQL está em execução
$containerRunning = docker ps --filter "name=integration-postgres-1" --format '{{.Names}}'
if (-not $containerRunning) {
    Write-Host "Erro: O contêiner PostgreSQL não está em execução." -ForegroundColor Red
    exit 1
}

# Lista de scripts em ordem de execução
$scripts = @(
    @{Path="/tmp/install/00_install_extensions.sql"; Name="Instalação de extensões"},
    @{Path="/core/01_schema_iam_core.sql"; Name="Esquema IAM Core"},
    @{Path="/core/02_views_iam_core.sql"; Name="Views IAM Core"},
    @{Path="/core/03_functions_iam_core.sql"; Name="Funções IAM Core"},
    @{Path="/core/04_triggers_iam_core.sql"; Name="Triggers IAM Core"}
)

# Copiar scripts para o contêiner
Write-Host "Copiando scripts para o contêiner..." -ForegroundColor Yellow
docker cp backend/database/scripts/iam/install integration-postgres-1:/tmp/
docker cp backend/database/scripts/iam/core integration-postgres-1:/

# Executar scripts
$success = $true
foreach ($script in $scripts) {
    $result = Execute-SqlScript -scriptPath $script.Path -scriptName $script.Name
    
    if (-not $result) {
        $success = $false
        Write-Host "Falha na execução do script: $($script.Name)" -ForegroundColor Red
        break
    }
}

if ($success) {
    Write-Host "Módulo IAM instalado com sucesso!" -ForegroundColor Green
    
    # Verificar instalação
    Write-Host "Verificando instalação..." -ForegroundColor Yellow
    docker exec integration-postgres-1 psql -U $dbUser -d $dbName -c "\dt iam.*"
    
    Write-Host "`nInstalação concluída com sucesso!" -ForegroundColor Green
}
else {
    Write-Host "Ocorreram erros durante a instalação do módulo IAM." -ForegroundColor Red
    exit 1
}
