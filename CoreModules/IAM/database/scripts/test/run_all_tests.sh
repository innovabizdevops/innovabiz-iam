#!/bin/bash
# Script para executar todos os testes do módulo IAM
# Autor: Eduardo Jeremias
# Data: 19/05/2025

# Configurações
DB_NAME="innovabiz_iam"  # Nome do banco de dados
DB_USER="postgres"       # Usuário do banco de dados
DB_HOST="localhost"      # Host do banco de dados
DB_PORT="5432"           # Porta do PostgreSQL
LOG_DIR="./test_logs"     # Diretório para armazenar logs

# Criar diretório de logs se não existir
mkdir -p "$LOG_DIR"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
LOG_FILE="$LOG_DIR/test_run_$TIMESTAMP.log"

# Função para registrar mensagens no log
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Função para executar um script SQL
exec_sql() {
    local script_name=$1
    local log_file="$LOG_DIR/${script_name%.*}_$TIMESTAMP.log"
    
    log "Executando $script_name..."
    
    # Executar o script SQL e capturar saída
    if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
         -f "$script_name" -v ON_ERROR_STOP=1 > "$log_file" 2>&1; then
        log "ERRO ao executar $script_name. Verifique o arquivo de log: $log_file"
        return 1
    else
        log "$script_name executado com sucesso."
        return 0
    fi
}

# Iniciar execução
log "Iniciando execução dos testes do módulo IAM"
log "Logs detalhados disponíveis em: $LOG_DIR/"
log "Arquivo de log principal: $LOG_FILE"

# 1. Desativar triggers de auditoria
if ! exec_sql "00_disable_audit_triggers.sql"; then
    log "Falha ao desativar triggers de auditoria. Abortando..."
    exit 1
fi

# 2. Executar script de dados de teste básicos
if ! exec_sql "01_create_test_data.sql"; then
    log "Falha ao criar dados de teste básicos. Abortando..."
    exit 1
fi

# 3. Executar script de dados de desempenho (opcional, comentado por padrão)
# if ! exec_sql "02_performance_test_data.sql"; then
#     log "Aviso: Falha ao gerar dados de desempenho. Continuando..."
# fi

# 4. Reativar triggers de auditoria
if ! exec_sql "99_enable_audit_triggers.sql"; then
    log "AVISO: Falha ao reativar triggers de auditoria. Verifique manualmente."
    exit 1
fi

# 5. Executar validações
log "Executando validações pós-teste..."
VALIDATION_LOG="$LOG_DIR/validation_$TIMESTAMP.log"

cat > /tmp/validate_tests.sql << EOF
\set ON_ERROR_STOP on

-- Verificar contagem de registros
SELECT 'Organizações: ' || COUNT(*) FROM iam.organizations;
SELECT 'Usuários: ' || COUNT(*) FROM iam.users;
SELECT 'Funções: ' || COUNT(*) FROM iam.roles;
SELECT 'Permissões: ' || (SELECT COUNT(*) FROM (
    SELECT jsonb_array_elements(permissions) FROM iam.roles
) AS permissions);
SELECT 'Logs de Auditoria: ' || COUNT(*) FROM iam.audit_logs;

-- Verificar integridade de chaves estrangeiras
SELECT 'Verificando usuários sem organização...';
SELECT COUNT(*) FROM iam.users WHERE organization_id NOT IN (SELECT id FROM iam.organizations);

SELECT 'Verificando funções sem organização...';
SELECT COUNT(*) FROM iam.roles WHERE organization_id NOT IN (SELECT id FROM iam.organizations);

SELECT 'Verificando usuários inativos...';
SELECT status, COUNT(*) FROM iam.users GROUP BY status;

-- Verificar índices
SELECT 'Verificando índices...';
SELECT
    t.tablename AS table_name,
    c2.relname AS index_name,
    pg_size_pretty(pg_relation_size(c2.oid)) AS index_size,
    am.amname AS index_type,
    i.indisunique AS is_unique,
    i.indisprimary AS is_primary,
    pg_get_indexdef(i.indexrelid) AS index_definition
FROM
    pg_index i
    JOIN pg_class c ON i.indrelid = c.oid
    JOIN pg_class c2 ON i.indexrelid = c2.oid
    JOIN pg_am am ON c2.relam = am.oid
    JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE
    n.nspname = 'iam'
ORDER BY
    pg_relation_size(i.indexrelid) DESC;

-- Verificar estatísticas de tabelas
SELECT 'Estatísticas de tabelas:';
SELECT
    t.tablename AS table_name,
    c.reltuples AS row_estimate,
    pg_size_pretty(pg_total_relation_size(quote_ident(t.tablename))) AS total_size,
    pg_size_pretty(pg_relation_size(quote_ident(t.tablename))) AS table_size,
    pg_size_pretty(pg_indexes_size(quote_ident(t.tablename))) AS indexes_size,
    t.n_live_tup AS live_tuples,
    t.n_dead_tup AS dead_tuples,
    t.last_vacuum,
    t.last_autovacuum,
    t.last_analyze,
    t.last_autoanalyze
FROM
    pg_stat_user_tables t
    JOIN pg_class c ON c.relname = t.relname
WHERE
    t.schemaname = 'iam'
ORDER BY
    pg_total_relation_size(quote_ident(t.tablename)) DESC;
EOF

if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
     -f /tmp/validate_tests.sql > "$VALIDATION_LOG" 2>&1; then
    log "AVISO: Algumas validações falharam. Verifique o arquivo: $VALIDATION_LOG"
else
    log "Validações concluídas com sucesso. Verifique o arquivo: $VALIDATION_LOG"
fi

# Finalização
log "Testes do módulo IAM concluídos em $(date '+%Y-%m-%d %H:%M:%S')"
log "Arquivo de log principal: $LOG_FILE"
log "Logs detalhados disponíveis em: $LOG_DIR/"

exit 0
