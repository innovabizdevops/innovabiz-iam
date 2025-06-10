# Script de Instalação do Módulo IAM
# Data: 19/05/2025
# Descrição: Instala o módulo IAM no banco de dados

# Configurações
$dbName = "innovabiz_iam"
$dbUser = "postgres"
$dbPassword = "postgres"
$dbHost = "localhost"
$dbPort = "5433"

# Diretório base
$scriptsDir = Split-Path -Parent $MyInvocation.MyCommand.Path

# Função para executar scripts SQL
function Invoke-SqlScript {
    param (
        [string]$scriptPath
    )
    Write-Host "Executando script: $scriptPath" -ForegroundColor Cyan
    
    # Verifica se o arquivo existe
    if (-not (Test-Path $scriptPath)) {
        Write-Host "Arquivo não encontrado: $scriptPath" -ForegroundColor Red
        return $false
    }
    
    # Executa o script via psql
    $command = "psql -h $dbHost -p $dbPort -U $dbUser -d $dbName -f `"$scriptPath`""
    
    try {
        $env:PGPASSWORD = $dbPassword
        $output = Invoke-Expression $command 2>&1 | Out-String
        
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Erro ao executar script: $scriptPath" -ForegroundColor Red
            Write-Host "Saída: $output" -ForegroundColor Red
            return $false
        }
        
        Write-Host "Script executado com sucesso: $scriptPath" -ForegroundColor Green
        Write-Host $output -ForegroundColor Gray
        return $true
    }
    catch {
        Write-Host "Exceção ao executar script: $_" -ForegroundColor Red
        return $false
    }
    finally {
        $env:PGPASSWORD = ""
    }
}

# Verifica conectividade com o PostgreSQL
try {
    Write-Host "Verificando conexão com o PostgreSQL..." -ForegroundColor Yellow
    $env:PGPASSWORD = $dbPassword
    Invoke-Expression "psql -h $dbHost -p $dbPort -U $dbUser -c `"SELECT 1;`" $dbName"
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Não foi possível conectar ao PostgreSQL. Verifique se o servidor está em execução e as credenciais estão corretas." -ForegroundColor Red
        exit 1
    }
}
catch {
    Write-Host "Erro ao verificar conexão: $_" -ForegroundColor Red
    exit 1
}
finally {
    $env:PGPASSWORD = ""
}

# Lista de scripts em ordem de execução
$scripts = @(
    "install/00_install_extensions.sql", # Adicionado para instalar extensões primeiro
    "install/00_install_iam_module.sql",
    "install/01_install_iam_core.sql",
    "install/02_install_iam_auth.sql",
    "install/03_install_iam_authz.sql"
)

# Executar scripts
$success = $true
foreach ($script in $scripts) {
    $scriptPath = Join-Path -Path $scriptsDir -ChildPath $script
    $result = Invoke-SqlScript -scriptPath $scriptPath
    
    if (-not $result) {
        $success = $false
        Write-Host "Falha na execução do script: $script" -ForegroundColor Red
        break
    }
}

if ($success) {
    Write-Host "Módulo IAM instalado com sucesso!" -ForegroundColor Green
    
    # Verificar instalação
    Write-Host "Verificando instalação..." -ForegroundColor Yellow
    $env:PGPASSWORD = $dbPassword
    $checkInstall = Invoke-Expression "psql -h $dbHost -p $dbPort -U $dbUser -d $dbName -c `"\dt iam.*`""
    Write-Host $checkInstall -ForegroundColor Cyan
    
    Write-Host "`nInstalação concluída com sucesso!" -ForegroundColor Green
}
else {
    Write-Host "Ocorreram erros durante a instalação do módulo IAM." -ForegroundColor Red
    exit 1
}
