#!/bin/bash
# =============================================================================
# Script de Validação de Políticas OPA para INNOVABIZ IAM
# =============================================================================
# Autor: INNOVABIZ DevSecOps Team
# Data: 2025-08-05
# Versão: 1.0.0
# Descrição: Este script executa validação completa das políticas OPA,
#            incluindo verificação sintática, testes unitários, e simulações
#            com diversos contextos regionais e multi-tenant.
# =============================================================================

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Definição de constantes
POLICIES_DIR="${1:-"../policies"}"
TESTS_DIR="${2:-"../policies/tests"}"
OUTPUT_DIR="./validation-results"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
LOG_FILE="${OUTPUT_DIR}/validation_${TIMESTAMP}.log"
REPORT_FILE="${OUTPUT_DIR}/report_${TIMESTAMP}.json"
CONTEXTS=("angola" "europa" "brasil" "sadc" "cplp" "china" "eua")

# Inicialização
mkdir -p "${OUTPUT_DIR}"
echo -e "${BLUE}[INFO]${NC} Iniciando validação das políticas OPA em ${POLICIES_DIR}"
echo -e "${BLUE}[INFO]${NC} Resultados serão gravados em ${LOG_FILE}"
echo -e "${BLUE}[INFO]${NC} Data de execução: $(date)"
echo -e "${BLUE}[INFO]${NC} Executando como: $(whoami)"

{
  echo "======================================================="
  echo "INNOVABIZ IAM - Validação de Políticas OPA"
  echo "======================================================="
  echo "Data de execução: $(date)"
  echo "Diretório de políticas: ${POLICIES_DIR}"
  echo "Diretório de testes: ${TESTS_DIR}"
  echo "======================================================="
} > "${LOG_FILE}"

# Função para verificação sintática das políticas
validate_syntax() {
  echo -e "${CYAN}[TAREFA]${NC} Verificando sintaxe das políticas Rego..."
  echo -e "\n\n## Verificação Sintática das Políticas" >> "${LOG_FILE}"
  
  local errors=0
  local files=$(find "${POLICIES_DIR}" -name "*.rego" | sort)
  local total=$(echo "${files}" | wc -l)
  local count=0
  
  for file in ${files}; do
    count=$((count + 1))
    echo -e "${BLUE}[PROGRESSO]${NC} Verificando ${count}/${total}: ${file}"
    
    if opa check "${file}" &>> "${LOG_FILE}"; then
      echo -e "${GREEN}[SUCESSO]${NC} ${file} - Sintaxe válida"
    else
      echo -e "${RED}[ERRO]${NC} ${file} - Erro de sintaxe"
      errors=$((errors + 1))
    fi
  done
  
  if [ ${errors} -eq 0 ]; then
    echo -e "${GREEN}[SUCESSO]${NC} Todas as ${total} políticas passaram na verificação sintática."
    return 0
  else
    echo -e "${RED}[ERRO]${NC} ${errors} políticas apresentaram erros sintáticos."
    return 1
  fi
}

# Função para executar testes unitários das políticas
run_unit_tests() {
  echo -e "${CYAN}[TAREFA]${NC} Executando testes unitários das políticas..."
  echo -e "\n\n## Testes Unitários das Políticas" >> "${LOG_FILE}"
  
  local errors=0
  local test_files=$(find "${TESTS_DIR}" -name "*_test.rego" | sort)
  local total=$(echo "${test_files}" | wc -l)
  local count=0
  
  for test_file in ${test_files}; do
    count=$((count + 1))
    local base_name=$(basename "${test_file}")
    echo -e "${BLUE}[PROGRESSO]${NC} Executando teste ${count}/${total}: ${base_name}"
    
    # Determinar diretório da política associada
    local policy_dir=$(dirname "${test_file}")
    
    if opa test "${policy_dir}" -v &>> "${LOG_FILE}"; then
      echo -e "${GREEN}[SUCESSO]${NC} ${base_name} - Todos os testes passaram"
    else
      echo -e "${RED}[ERRO]${NC} ${base_name} - Falhas nos testes"
      errors=$((errors + 1))
    fi
  done
  
  if [ ${errors} -eq 0 ]; then
    echo -e "${GREEN}[SUCESSO]${NC} Todos os ${total} arquivos de teste passaram."
    return 0
  else
    echo -e "${RED}[ERRO]${NC} ${errors} arquivos de teste apresentaram falhas."
    return 1
  fi
}

# Função para simular decisões de autorização em diferentes contextos
simulate_decisions() {
  echo -e "${CYAN}[TAREFA]${NC} Simulando decisões de autorização em diferentes contextos..."
  echo -e "\n\n## Simulações de Decisões por Contexto" >> "${LOG_FILE}"
  
  local errors=0
  local regional_contexts="${TESTS_DIR}/inputs"
  
  for context in "${CONTEXTS[@]}"; do
    echo -e "${BLUE}[PROGRESSO]${NC} Simulando decisões para contexto: ${context}"
    echo -e "\n### Contexto: ${context}" >> "${LOG_FILE}"
    
    local input_files=$(find "${regional_contexts}/${context}" -name "*.json" 2>/dev/null)
    
    if [ -z "${input_files}" ]; then
      echo -e "${YELLOW}[AVISO]${NC} Nenhum arquivo de input encontrado para ${context}"
      echo "Nenhum arquivo de input encontrado." >> "${LOG_FILE}"
      continue
    fi
    
    for input_file in ${input_files}; do
      local test_name=$(basename "${input_file}" .json)
      echo -e "${BLUE}[INFO]${NC} Teste: ${test_name}"
      
      if opa eval --data "${POLICIES_DIR}" --input "${input_file}" "data.innovabiz.iam.authz.allow" &>> "${LOG_FILE}"; then
        echo -e "${GREEN}[SUCESSO]${NC} ${test_name} - Decisão processada com sucesso"
      else
        echo -e "${RED}[ERRO]${NC} ${test_name} - Falha ao processar decisão"
        errors=$((errors + 1))
      fi
    done
  done
  
  if [ ${errors} -eq 0 ]; then
    echo -e "${GREEN}[SUCESSO]${NC} Todas as simulações de decisão foram processadas com sucesso."
    return 0
  else
    echo -e "${RED}[ERRO]${NC} ${errors} simulações apresentaram falhas."
    return 1
  fi
}

# Função para análise de cobertura dos testes
analyze_test_coverage() {
  echo -e "${CYAN}[TAREFA]${NC} Analisando cobertura dos testes..."
  echo -e "\n\n## Análise de Cobertura de Testes" >> "${LOG_FILE}"
  
  if opa test --coverage "${POLICIES_DIR}" &> /tmp/coverage.txt; then
    coverage=$(grep -oP 'Coverage: \K[0-9.]+%' /tmp/coverage.txt)
    echo -e "${GREEN}[SUCESSO]${NC} Análise de cobertura concluída: ${coverage}"
    echo "Cobertura de testes: ${coverage}" >> "${LOG_FILE}"
    cat /tmp/coverage.txt >> "${LOG_FILE}"
  else
    echo -e "${RED}[ERRO]${NC} Falha ao analisar cobertura dos testes"
    echo "Falha ao analisar cobertura dos testes" >> "${LOG_FILE}"
    return 1
  fi
  
  return 0
}

# Função para verificação de compliance regional
check_regional_compliance() {
  echo -e "${CYAN}[TAREFA]${NC} Verificando compliance regional das políticas..."
  echo -e "\n\n## Verificação de Compliance Regional" >> "${LOG_FILE}"
  
  local compliance_dir="${TESTS_DIR}/compliance"
  local errors=0
  
  for context in "${CONTEXTS[@]}"; do
    echo -e "${BLUE}[PROGRESSO]${NC} Verificando compliance para região: ${context}"
    echo -e "\n### Região: ${context}" >> "${LOG_FILE}"
    
    local rules_file="${compliance_dir}/${context}_rules.rego"
    
    if [ ! -f "${rules_file}" ]; then
      echo -e "${YELLOW}[AVISO]${NC} Arquivo de regras de compliance não encontrado para ${context}"
      echo "Arquivo de regras não encontrado para ${context}" >> "${LOG_FILE}"
      continue
    fi
    
    if opa test "${rules_file}" "${POLICIES_DIR}" -v &>> "${LOG_FILE}"; then
      echo -e "${GREEN}[SUCESSO]${NC} ${context} - Políticas em compliance com regulamentações"
    else
      echo -e "${RED}[ERRO]${NC} ${context} - Falhas de compliance detectadas"
      errors=$((errors + 1))
    fi
  done
  
  if [ ${errors} -eq 0 ]; then
    echo -e "${GREEN}[SUCESSO]${NC} Todas as verificações de compliance regional passaram."
    return 0
  else
    echo -e "${RED}[ERRO]${NC} ${errors} regiões apresentaram problemas de compliance."
    return 1
  fi
}

# Função para verificar performance das políticas
benchmark_policies() {
  echo -e "${CYAN}[TAREFA]${NC} Realizando benchmark de performance das políticas..."
  echo -e "\n\n## Benchmark de Performance" >> "${LOG_FILE}"
  
  local benchmark_dir="${TESTS_DIR}/benchmark"
  local benchmarks=$(find "${benchmark_dir}" -name "*.json" 2>/dev/null)
  
  if [ -z "${benchmarks}" ]; then
    echo -e "${YELLOW}[AVISO]${NC} Nenhum arquivo de benchmark encontrado"
    echo "Nenhum arquivo de benchmark encontrado." >> "${LOG_FILE}"
    return 0
  fi
  
  for bench in ${benchmarks}; do
    local bench_name=$(basename "${bench}" .json)
    echo -e "${BLUE}[PROGRESSO]${NC} Executando benchmark: ${bench_name}"
    
    if opa bench --count=1000 --data "${POLICIES_DIR}" --input "${bench}" "data.innovabiz.iam.authz.allow" &>> "${LOG_FILE}"; then
      # Extrair métricas de performance
      tail -10 "${LOG_FILE}" | grep "ns/op"
      echo -e "${GREEN}[SUCESSO]${NC} ${bench_name} - Benchmark concluído"
    else
      echo -e "${RED}[ERRO]${NC} ${bench_name} - Falha ao executar benchmark"
    fi
  done
  
  return 0
}

# Função para gerar relatório de validação
generate_validation_report() {
  echo -e "${CYAN}[TAREFA]${NC} Gerando relatório de validação..."
  
  # Extrair métricas e resultados do log
  syntax_result=$(grep -c "\[SUCESSO\] Todas as .* políticas passaram na verificação sintática" "${LOG_FILE}" || echo "0")
  tests_result=$(grep -c "\[SUCESSO\] Todos os .* arquivos de teste passaram" "${LOG_FILE}" || echo "0")
  simulation_result=$(grep -c "\[SUCESSO\] Todas as simulações de decisão foram processadas com sucesso" "${LOG_FILE}" || echo "0")
  coverage=$(grep -oP 'Cobertura de testes: \K[0-9.]+%' "${LOG_FILE}" || echo "N/A")
  compliance_result=$(grep -c "\[SUCESSO\] Todas as verificações de compliance regional passaram" "${LOG_FILE}" || echo "0")
  
  # Determinar status geral
  if [[ "${syntax_result}" == "1" && "${tests_result}" == "1" && "${simulation_result}" == "1" && "${compliance_result}" == "1" ]]; then
    status="SUCESSO"
  else
    status="FALHA"
  fi
  
  # Criar relatório JSON
  cat > "${REPORT_FILE}" << EOF
{
  "timestamp": "$(date -Iseconds)",
  "status": "${status}",
  "details": {
    "syntax_validation": {
      "status": ${syntax_result},
      "description": "Validação sintática das políticas Rego"
    },
    "unit_tests": {
      "status": ${tests_result},
      "description": "Testes unitários das políticas"
    },
    "decision_simulation": {
      "status": ${simulation_result},
      "description": "Simulação de decisões de autorização"
    },
    "test_coverage": {
      "value": "${coverage}",
      "description": "Cobertura dos testes"
    },
    "compliance_verification": {
      "status": ${compliance_result},
      "description": "Verificação de compliance regional"
    }
  },
  "log_file": "${LOG_FILE}",
  "execution_environment": {
    "opa_version": "$(opa version | grep Version | cut -d ' ' -f 2)",
    "platform": "$(uname -s)",
    "hostname": "$(hostname)",
    "user": "$(whoami)"
  }
}
EOF

  echo -e "${GREEN}[SUCESSO]${NC} Relatório de validação gerado em ${REPORT_FILE}"
}

# Função para instalar OPA se não estiver disponível
ensure_opa_installed() {
  if ! command -v opa &> /dev/null; then
    echo -e "${YELLOW}[AVISO]${NC} OPA não encontrado. Tentando instalar..."
    
    # Verificar sistema operacional
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
      curl -L -o opa https://openpolicyagent.org/downloads/latest/opa_linux_amd64
      chmod 755 opa
      sudo mv opa /usr/local/bin/opa
    elif [[ "$OSTYPE" == "darwin"* ]]; then
      brew install opa
    elif [[ "$OSTYPE" == "msys"* || "$OSTYPE" == "win32" ]]; then
      echo -e "${YELLOW}[AVISO]${NC} No Windows, é recomendável instalar o OPA manualmente."
      echo -e "${YELLOW}[AVISO]${NC} Por favor, baixe de https://openpolicyagent.org/downloads/ e adicione ao PATH."
      exit 1
    else
      echo -e "${RED}[ERRO]${NC} Sistema operacional não suportado para instalação automática."
      exit 1
    fi
    
    if command -v opa &> /dev/null; then
      echo -e "${GREEN}[SUCESSO]${NC} OPA instalado com sucesso: $(opa version)"
    else
      echo -e "${RED}[ERRO]${NC} Falha ao instalar OPA. Por favor, instale manualmente."
      exit 1
    fi
  fi
}

# Execução principal
main() {
  ensure_opa_installed
  
  echo -e "${PURPLE}=========================================================${NC}"
  echo -e "${PURPLE}  INNOVABIZ IAM - Validação de Políticas OPA   ${NC}"
  echo -e "${PURPLE}=========================================================${NC}"
  
  local start_time=$(date +%s)
  local success=true
  
  # Executar validações
  validate_syntax || success=false
  run_unit_tests || success=false
  simulate_decisions || success=false
  analyze_test_coverage
  check_regional_compliance || success=false
  benchmark_policies
  
  # Gerar relatório
  generate_validation_report
  
  local end_time=$(date +%s)
  local duration=$((end_time - start_time))
  
  echo -e "${PURPLE}=========================================================${NC}"
  echo -e "${BLUE}[INFO]${NC} Duração total: ${duration} segundos"
  
  if [ "${success}" = true ]; then
    echo -e "${GREEN}[SUCESSO]${NC} Todas as validações foram concluídas com sucesso."
    echo -e "${PURPLE}=========================================================${NC}"
    exit 0
  else
    echo -e "${RED}[ERRO]${NC} Algumas validações falharam. Verifique o log para detalhes."
    echo -e "${PURPLE}=========================================================${NC}"
    exit 1
  fi
}

# Executar script
main