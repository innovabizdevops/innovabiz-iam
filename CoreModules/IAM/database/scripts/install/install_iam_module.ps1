# Script de instalação do Módulo IAM para INABIZ
# Autor: INNOVABIZ DevOps
# Data: 2025-05-19

# Configurações
$env:PGPASSWORD = "sua_senha_aqui"  # Substitua pela senha do PostgreSQL
$dbName = "innabiz_iam"
$dbUser = "postgres"  # Substitua pelo usuário do PostgreSQL
$dbHost = "localhost"  # Substitua pelo host do PostgreSQL
$scriptsPath = "$PSScriptRoot\..\..\..\..\docs\iam\tecnico\scripts"  # Ajuste conforme necessário

# Função para executar um comando e verificar o resultado
function Invoke-CommandWithCheck {
    param (
        [string]$command,
        [string]$successMessage,
        [string]$errorMessage
    )
    
    Write-Host "Executando: $command" -ForegroundColor Cyan
    $output = Invoke-Expression $command
    
    if ($LASTEXITCODE -eq 0) {
        if ($successMessage) { Write-Host $successMessage -ForegroundColor Green }
        return $true
    } else {
        Write-Host $errorMessage -ForegroundColor Red
        Write-Host "Saída do comando: $output" -ForegroundColor Red
        return $false
    }
}

# 1. Verificar se o PostgreSQL está acessível
Write-Host "`n[1/5] Verificando conexão com o PostgreSQL..." -ForegroundColor Yellow
$pgCheck = & psql -h $dbHost -U $dbUser -lqt

if ($LASTEXITCODE -ne 0) {
    Write-Host "Erro: Não foi possível conectar ao PostgreSQL. Verifique as configurações." -ForegroundColor Red
    exit 1
}

# 2. Criar banco de dados se não existir
Write-Host "`n[2/5] Verificando banco de dados $dbName..." -ForegroundColor Yellow
$dbExists = $pgCheck | Select-String -Pattern "\b$dbName\b"

if (-not $dbExists) {
    Write-Host "Criando banco de dados $dbName..." -ForegroundColor Cyan
    $createDbCmd = "createdb -h $dbHost -U $dbUser -E UTF8 --lc-collate='Portuguese_Brazil.1252' --lc-ctype='Portuguese_Brazil.1252' -T template0 $dbName"
    
    if (-not (Invoke-CommandWithCheck -command $createDbCmd `
        -successMessage "Banco de dados $dbName criado com sucesso." `
        -errorMessage "Erro ao criar o banco de dados.")) {
        exit 1
    }
} else {
    Write-Host "Banco de dados $dbName já existe." -ForegroundColor Green
}

# 3. Executar scripts de schema
Write-Host "`n[3/5] Executando scripts de schema..." -ForegroundColor Yellow

$scripts = @(
    "01_iam_core_schema.sql",
    "02_auth_methods_data.sql",
    "03_auth_policies_config.sql",
    "04_healthcare_compliance.sql",
    "05_openbanking_compliance.sql"
)

foreach ($script in $scripts) {
    $scriptPath = Join-Path -Path $scriptsPath -ChildPath $script
    
    if (-not (Test-Path $scriptPath)) {
        Write-Host "Aviso: Arquivo $script não encontrado em $scriptPath" -ForegroundColor Yellow
        continue
    }
    
    Write-Host "Executando $script..." -ForegroundColor Cyan
    $execCmd = "psql -h $dbHost -U $dbUser -d $dbName -f \"$scriptPath\""
    
    if (-not (Invoke-CommandWithCheck -command $execCmd `
        -successMessage "Script $script executado com sucesso." `
        -errorMessage "Erro ao executar o script $script.")) {
        exit 1
    }
}

# 4. Criar usuários e permissões
Write-Host "`n[4/5] Configurando usuários e permissões..." -ForegroundColor Yellow

$sql = @"
-- Criar role de aplicação
DO \$\$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'iam_app') THEN
        CREATE ROLE iam_app WITH NOLOGIN NOSUPERUSER INHERIT NOCREATEDB NOCREATEROLE NOREPLICATION;
    END IF;
    
    -- Conceder permissões
    GRANT USAGE ON SCHEMA iam_core TO iam_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA iam_core TO iam_app;
    GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA iam_core TO iam_app;
    
    -- Mesmo para compliance_validators
    GRANT USAGE ON SCHEMA compliance_validators TO iam_app;
    GRANT SELECT ON ALL TABLES IN SCHEMA compliance_validators TO iam_app;
    
    -- Criar usuário de aplicação
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'innabiz_iam_user') THEN
        CREATE ROLE innabiz_iam_user WITH LOGIN PASSWORD 'mudar@123' NOSUPERUSER INHERIT NOCREATEDB NOCREATEROLE NOREPLICATION;
    END IF;
    
    -- Atribuir role ao usuário
    GRANT iam_app TO innabiz_iam_user;
END
\$\$;
"@

$tempFile = [System.IO.Path]::GetTempFileName() + ".sql"
$sql | Out-File -FilePath $tempFile -Encoding utf8

if (-not (Invoke-CommandWithCheck -command "psql -h $dbHost -U $dbUser -d $dbName -f \"$tempFile\"" `
    -successMessage "Usuários e permissões configurados com sucesso." `
    -errorMessage "Erro ao configurar usuários e permissões.")) {
    Remove-Item $tempFile -ErrorAction SilentlyContinue
    exit 1
}

Remove-Item $tempFile -ErrorAction SilentlyContinue

# 5. Verificar instalação
Write-Host "`n[5/5] Verificando instalação..." -ForegroundColor Yellow

$checkQuery = @"
SELECT 
    'iam_core' as schema,
    COUNT(*) as tables
FROM information_schema.tables 
WHERE table_schema = 'iam_core'
UNION ALL
SELECT 
    'compliance_validators' as schema,
    COUNT(*) as tables
FROM information_schema.tables 
WHERE table_schema = 'compliance_validators';
"@

$tempFile = [System.IO.Path]::GetTempFileName() + ".sql"
$checkQuery | Out-File -FilePath $tempFile -Encoding utf8

Write-Host "`nResumo da instalação:" -ForegroundColor Green
& psql -h $dbHost -U $dbUser -d $dbName -f $tempFile

Remove-Item $tempFile -ErrorAction SilentlyContinue

# Verificação final
$finalCheck = & psql -h $dbHost -U $dbUser -d $dbName -c "SELECT COUNT(*) as total_metodos FROM iam_core.authentication_methods;" -t

if ($LASTEXITCODE -eq 0) {
    Write-Host "`nInstalação concluída com sucesso!" -ForegroundColor Green
    Write-Host "Total de métodos de autenticação instalados: $($finalCheck.Trim())" -ForegroundColor Green
    
    Write-Host "`nPróximos passos:" -ForegroundColor Cyan
    Write-Host "1. Altere a senha do usuário 'innabiz_iam_user'"
    Write-Host "2. Configure as conexões de aplicação para usar as credenciais fornecidas"
    Write-Host "3. Execute os testes de integração para validar a instalação"
    
    exit 0
} else {
    Write-Host "`nA instalação foi concluída, mas ocorreram avisos." -ForegroundColor Yellow
    Write-Host "Verifique os logs acima para mais detalhes." -ForegroundColor Yellow
    exit 1
}
