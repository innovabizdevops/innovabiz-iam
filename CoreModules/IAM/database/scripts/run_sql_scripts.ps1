# Script para executar scripts SQL em sequência
# Data: 20/05/2025

# Configurações
$psqlPath = "C:\Program Files\PostgreSQL\17\bin\psql.exe"
$dbName = "innovabiz_iam"
$dbUser = "postgres"
$dbHost = "localhost"
$dbPort = "5432"

# Solicitar senha de forma segura
$dbPassword = Read-Host -AsSecureString -Prompt "Digite a senha do PostgreSQL"
$dbPassword = [Runtime.InteropServices.Marshal]::PtrToStringAuto(
    [Runtime.InteropServices.Marshal]::SecureStringToBSTR($dbPassword)
)

# Função para executar um script SQL
function Execute-SqlScript {
    param (
        [string]$scriptPath
    )
    
    Write-Host "Executando script: $scriptPath" -ForegroundColor Cyan
    
    if (-not (Test-Path $scriptPath)) {
        Write-Host "Arquivo nÆo encontrado: $scriptPath" -ForegroundColor Red
        return $false
    }
    
    $env:PGPASSWORD = $dbPassword
    $arguments = @(
        "-h", $dbHost,
        "-p", $dbPort,
        "-U", $dbUser,
        "-d", $dbName,
        "-f", "`"$scriptPath`""
    )
    
    try {
        $process = Start-Process -FilePath $psqlPath -ArgumentList $arguments -NoNewWindow -Wait -PassThru
        
        if ($process.ExitCode -ne 0) {
            Write-Host "Erro ao executar o script: $scriptPath" -ForegroundColor Red
            return $false
        }
        
        Write-Host "Script executado com sucesso: $scriptPath" -ForegroundColor Green
        return $true
    }
    catch {
        Write-Host "Erro ao executar o script: $_" -ForegroundColor Red
        return $false
    }
    finally {
        $env:PGPASSWORD = ""
    }
}

# Lista de scripts em ordem de execução
$scripts = @(
    "install\00_install_extensions.sql",
    "install\00_install_iam_module.sql",
    "install\01_install_iam_core.sql",
    "install\02_install_iam_auth.sql",
    "install\03_install_iam_authz.sql",
    "core\01_schema_iam_core.sql",
    "core\02_views_iam_core.sql",
    "core\03_functions_iam_core.sql",
    "core\04_triggers_iam_core.sql"
)

# Executar scripts
$success = $true
foreach ($script in $scripts) {
    $scriptPath = Join-Path -Path $PSScriptRoot -ChildPath $script
    $result = Execute-SqlScript -scriptPath $scriptPath
    
    if (-not $result) {
        $success = $false
        Write-Host "Falha na execuçÆo do script: $script" -ForegroundColor Red
        break
    }
}

if ($success) {
    Write-Host "Todos os scripts foram executados com sucesso!" -ForegroundColor Green
    
    # Verificar instalação
    Write-Host "Verificando instalação..." -ForegroundColor Yellow
    $env:PGPASSWORD = $dbPassword
    & $psqlPath -h $dbHost -p $dbPort -U $dbUser -d $dbName -c "\dt iam.*"
    $env:PGPASSWORD = ""
    
    Write-Host "`nInstalação concluída com sucesso!" -ForegroundColor Green
}
else {
    Write-Host "Ocorreram erros durante a instalação do módulo IAM." -ForegroundColor Red
    exit 1
}
