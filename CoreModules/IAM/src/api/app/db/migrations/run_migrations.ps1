# INNOVABIZ IAM - Script de Execução de Migrações para Auditoria Multi-Contexto
# Autor: Eduardo Jeremias
# Versão: 1.0.0
# Descrição: Script PowerShell para execução das migrações de banco de dados do sistema de auditoria

# Definição de parâmetros
param (
    [string]$Host = "localhost",
    [string]$Port = "5432",
    [string]$Database = "innovabiz_iam",
    [string]$Username = "postgres",
    [string]$Password,
    [switch]$SkipPrompt = $false,
    [switch]$DropTables = $false
)

# Função para exibir banner
function Show-Banner {
    Write-Host "=============================================================" -ForegroundColor Cyan
    Write-Host "            INNOVABIZ IAM - MIGRAÇÕES DE AUDITORIA           " -ForegroundColor Cyan
    Write-Host "             Multi-Tenant | Multi-Regional | Multi-Contexto   " -ForegroundColor Cyan
    Write-Host "=============================================================" -ForegroundColor Cyan
    Write-Host ""
}

# Função para verificar se o psql está instalado
function Test-PostgresInstalled {
    try {
        $psqlVersion = Invoke-Expression "psql --version"
        Write-Host "✅ PostgreSQL CLI encontrado: $psqlVersion" -ForegroundColor Green
        return $true
    } catch {
        Write-Host "❌ PostgreSQL CLI (psql) não encontrado no PATH!" -ForegroundColor Red
        Write-Host "Por favor, instale o PostgreSQL e adicione o diretório bin ao PATH." -ForegroundColor Yellow
        return $false
    }
}

# Função para executar um script SQL
function Invoke-SqlScript {
    param (
        [string]$ScriptPath,
        [string]$ScriptName
    )
    
    Write-Host "⚙️ Executando $ScriptName..." -NoNewline
    
    # Preparar comando PSQL
    $psqlCmd = "psql -h $Host -p $Port -d $Database -U $Username -f `"$ScriptPath`" -v ON_ERROR_STOP=1"
    
    # Se a senha for fornecida, define a variável de ambiente PGPASSWORD
    if ($Password) {
        $env:PGPASSWORD = $Password
    }
    
    try {
        # Executa o comando e captura a saída
        $output = Invoke-Expression $psqlCmd 2>&1
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host " ✅ Sucesso!" -ForegroundColor Green
            return $true
        } else {
            Write-Host " ❌ Falha!" -ForegroundColor Red
            Write-Host "   Erro: $output" -ForegroundColor Red
            return $false
        }
    } catch {
        Write-Host " ❌ Falha!" -ForegroundColor Red
        Write-Host "   Erro: $_" -ForegroundColor Red
        return $false
    } finally {
        # Limpa a variável de ambiente
        if ($Password) {
            $env:PGPASSWORD = ""
        }
    }
}

# Função para confirmar a execução
function Confirm-Execution {
    if ($SkipPrompt) {
        return $true
    }
    
    Write-Host "⚠️ ATENÇÃO: Você está prestes a executar migrações de banco de dados para o sistema de auditoria." -ForegroundColor Yellow
    Write-Host "   Database: $Database@$Host" -ForegroundColor Yellow
    Write-Host "   Usuário: $Username" -ForegroundColor Yellow
    
    if ($DropTables) {
        Write-Host "   🔥 AVISO: A opção -DropTables foi especificada! Todas as tabelas de auditoria existentes serão EXCLUÍDAS!" -ForegroundColor Red
    }
    
    $confirmation = Read-Host "Deseja continuar? (S/N)"
    return $confirmation -eq "S" -or $confirmation -eq "s"
}

# Função para criar script de rollback
function Create-RollbackScript {
    $rollbackPath = Join-Path $PSScriptRoot "rollback_migrations.sql"
    
    $rollbackContent = @"
-- Script de rollback gerado automaticamente
-- Data: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

-- Remover tabelas na ordem inversa de dependência
DROP TABLE IF EXISTS audit_statistics;
DROP TABLE IF EXISTS audit_compliance_reports;
DROP TABLE IF EXISTS audit_retention_policies;
DROP TABLE IF EXISTS audit_events;

-- Remover funções e procedures
DROP FUNCTION IF EXISTS generate_audit_statistics;
DROP FUNCTION IF EXISTS mask_sensitive_fields;
DROP PROCEDURE IF EXISTS apply_retention_policy;
DROP FUNCTION IF EXISTS generate_partition_key;
DROP FUNCTION IF EXISTS detect_compliance_frameworks;
DROP FUNCTION IF EXISTS update_updated_at_column;

-- Remover tipos enumerados
DROP TYPE IF EXISTS report_status;
DROP TYPE IF EXISTS compliance_framework;
DROP TYPE IF EXISTS audit_event_severity;
DROP TYPE IF EXISTS audit_event_category;

-- Finalização
SELECT 'Rollback concluído com sucesso!' as status;
"@

    Set-Content -Path $rollbackPath -Value $rollbackContent
    Write-Host "✅ Script de rollback gerado em $rollbackPath" -ForegroundColor Green
}

# Função principal
function Start-Migrations {
    Show-Banner
    
    # Verifica se o PostgreSQL está instalado
    if (-not (Test-PostgresInstalled)) {
        return
    }
    
    # Solicita senha se não fornecida
    if (-not $Password) {
        $Password = Read-Host "Digite a senha do PostgreSQL para o usuário $Username" -AsSecureString
        $BSTR = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($Password)
        $Password = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($BSTR)
    }
    
    # Confirma a execução
    if (-not (Confirm-Execution)) {
        Write-Host "❌ Operação cancelada pelo usuário." -ForegroundColor Yellow
        return
    }
    
    # Executa drop tables se especificado
    if ($DropTables) {
        Write-Host "🔄 Executando rollback das tabelas existentes..." -ForegroundColor Yellow
        
        # Cria e executa um script de rollback
        Create-RollbackScript
        $rollbackPath = Join-Path $PSScriptRoot "rollback_migrations.sql"
        Invoke-SqlScript -ScriptPath $rollbackPath -ScriptName "Script de rollback"
    }
    
    # Lista de scripts na ordem correta de execução
    $scripts = @(
        @{Path = Join-Path $PSScriptRoot "001_create_audit_types.sql"; Name = "Tipos de Auditoria"},
        @{Path = Join-Path $PSScriptRoot "002_create_audit_tables.sql"; Name = "Tabelas de Auditoria"},
        @{Path = Join-Path $PSScriptRoot "003_create_audit_functions.sql"; Name = "Funções de Auditoria"},
        @{Path = Join-Path $PSScriptRoot "004_insert_initial_data.sql"; Name = "Dados Iniciais"}
    )
    
    # Contador para acompanhar o progresso
    $totalScripts = $scripts.Count
    $successCount = 0
    $failureCount = 0
    
    # Executa cada script na ordem
    foreach ($script in $scripts) {
        $success = Invoke-SqlScript -ScriptPath $script.Path -ScriptName $script.Name
        
        if ($success) {
            $successCount++
        } else {
            $failureCount++
            Write-Host "⚠️ Falha ao executar o script $($script.Name). Interrompendo execução." -ForegroundColor Red
            break
        }
    }
    
    # Exibe resultado final
    Write-Host ""
    Write-Host "=============================================================" -ForegroundColor Cyan
    Write-Host " RESULTADO DAS MIGRAÇÕES DE AUDITORIA MULTI-CONTEXTO" -ForegroundColor Cyan
    Write-Host "=============================================================" -ForegroundColor Cyan
    Write-Host " Total de scripts: $totalScripts" -ForegroundColor Cyan
    Write-Host " Scripts executados com sucesso: $successCount" -ForegroundColor Green
    Write-Host " Scripts com falha: $failureCount" -ForegroundColor Red
    
    if ($failureCount -eq 0) {
        Write-Host ""
        Write-Host "✅ SISTEMA DE AUDITORIA MULTI-CONTEXTO INSTALADO COM SUCESSO!" -ForegroundColor Green
        Write-Host "   - Suporte a Multi-Tenant" -ForegroundColor Cyan
        Write-Host "   - Contextos Regionais: BR, US, EU, AO" -ForegroundColor Cyan
        Write-Host "   - Frameworks de Compliance: LGPD, GDPR, SOX, PCI DSS, BACEN, BNA, PSD2" -ForegroundColor Cyan
        Write-Host "   - Políticas de Retenção e Anonimização Configuradas" -ForegroundColor Cyan
    } else {
        Write-Host ""
        Write-Host "❌ FALHA NA INSTALAÇÃO DO SISTEMA DE AUDITORIA" -ForegroundColor Red
        Write-Host "   Verifique os erros acima e tente novamente." -ForegroundColor Red
    }
}

# Inicia o processo
Start-Migrations