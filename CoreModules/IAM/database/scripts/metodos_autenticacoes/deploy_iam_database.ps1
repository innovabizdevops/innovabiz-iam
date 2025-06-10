# Script de implantação do banco de dados IAM Open X
# Autor: INNOVABIZ DevOps
# Data: $(Get-Date -Format "yyyy-MM-dd")

# Configurações
$env:PGPASSWORD = "sua_senha_aqui"  # Substitua pela senha do PostgreSQL
$dbName = "iam_open_x"
$dbUser = "postgres"  # Substitua pelo usuário do PostgreSQL
$dbHost = "localhost"  # Substitua pelo host do PostgreSQL
$scriptsPath = "C:\Users\HP\Dropbox\InnovaBiz\docs\iam\tecnico\scripts"

# Função para executar um script SQL
function Execute-SqlScript {
    param (
        [string]$scriptName,
        [string]$dbName
    )
    
    $scriptPath = Join-Path -Path $scriptsPath -ChildPath $scriptName
    
    Write-Host "Executando script: $scriptName" -ForegroundColor Cyan
    
    try {
        $result = & psql -h $dbHost -U $dbUser -d $dbName -f $scriptPath
        Write-Host "Script $scriptName executado com sucesso" -ForegroundColor Green
        return $true
    }
    catch {
        Write-Host "Erro ao executar o script $scriptName : $_" -ForegroundColor Red
        return $false
    }
}

# 1. Criar banco de dados
Write-Host "Criando banco de dados $dbName..." -ForegroundColor Yellow
& createdb -h $dbHost -U $dbUser $dbName

if ($LASTEXITCODE -ne 0) {
    Write-Host "Banco de dados já existe ou ocorreu um erro. Continuando..." -ForegroundColor Yellow
}

# 2. Executar scripts na ordem correta
$scripts = @(
    "01_iam_core_schema.sql",
    "02_auth_methods_data.sql",
    "03_auth_policies_config.sql",
    "04_healthcare_compliance.sql",
    "05_openbanking_compliance.sql"
)

$allScriptsSuccessful = $true

foreach ($script in $scripts) {
    if (-not (Test-Path (Join-Path -Path $scriptsPath -ChildPath $script))) {
        Write-Host "Aviso: Arquivo $script não encontrado. Pulando..." -ForegroundColor Yellow
        continue
    }
    
    if (-not (Execute-SqlScript -scriptName $script -dbName $dbName)) {
        $allScriptsSuccessful = $false
        Write-Host "Erro crítico ao executar $script. Abortando..." -ForegroundColor Red
        break
    }
}

# 3. Verificar se todos os scripts foram executados com sucesso
if ($allScriptsSuccessful) {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "  Banco de dados IAM implantado com sucesso!" -ForegroundColor Green
    Write-Host "  Data/Hora: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')" -ForegroundColor Green
    Write-Host "  Banco de Dados: $dbName" -ForegroundColor Green
    Write-Host "  Host: $dbHost" -ForegroundColor Green
    Write-Host "  Usuário: $dbUser" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    
    # Exibir resumo das tabelas criadas
    Write-Host "`nTabelas criadas no schema iam_core:" -ForegroundColor Cyan
    & psql -h $dbHost -U $dbUser -d $dbName -c "\dt iam_core.*"
    
    Write-Host "`nTabelas criadas no schema compliance_validators:" -ForegroundColor Cyan
    & psql -h $dbHost -U $dbUser -d $dbName -c "\dt compliance_validators.*"
    
    # Exibir contagem de métodos de autenticação cadastrados
    Write-Host "`nResumo de Métodos de Autenticação:" -ForegroundColor Cyan
    & psql -h $dbHost -U $dbUser -d $dbName -c "
        SELECT 
            category_id as Categoria,
            COUNT(*) as Quantidade,
            STRING_AGG(method_name, ', ' ORDER BY method_name) as Metodos_Exemplo
        FROM iam_core.authentication_methods 
        GROUP BY category_id 
        ORDER BY Quantidade DESC;
    "
    
    Write-Host "`nImplantação concluída com sucesso!" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Red
    Write-Host "  Ocorreram erros durante a implantação" -ForegroundColor Red
    Write-Host "  Verifique os logs acima para mais detalhes" -ForegroundColor Red
    Write-Host "========================================" -ForegroundColor Red
    exit 1
}
