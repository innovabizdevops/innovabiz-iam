#!/bin/bash
# =============================================================================
# CI/CD Script para Políticas OPA do RoleService - INNOVABIZ Platform
# =============================================================================
# Versão: 1.0.0
# Data: 2025-08-05
# Autor: INNOVABIZ IAM Team
# 
# Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
# PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III
# =============================================================================

set -e

# Cores para saída no terminal
YELLOW='\033[1;33m'
GREEN='\033[1;32m'
RED='\033[1;31m'
BLUE='\033[1;34m'
RESET='\033[0m'

# Variáveis
TIMESTAMP=$(date +"%Y-%m-%d_%H-%M-%S")
POLICY_DIR="$(dirname "$(readlink -f "$0")")/.."
OPA_BIN=$(which opa || echo "opa")
OUTPUT_DIR="${POLICY_DIR}/dist"
COVERAGE_DIR="${POLICY_DIR}/coverage"
BUNDLE_NAME="innovabiz-iam-role-policies-${TIMESTAMP}.tar.gz"
ENV=${1:-"development"}
DEPLOY_TARGET=${2:-"local"}

# Verificar se OPA está instalado
check_opa() {
    echo -e "${BLUE}======================================${RESET}"
    echo -e "${BLUE}Verificando instalação do OPA...${RESET}"
    echo -e "${BLUE}======================================${RESET}"
    
    if ! command -v "${OPA_BIN}" &> /dev/null; then
        echo -e "${RED}OPA não encontrado. Instalando...${RESET}"
        curl -L -o opa https://openpolicyagent.org/downloads/latest/opa_linux_amd64
        chmod 755 opa
        OPA_BIN="./opa"
    else
        echo -e "${GREEN}OPA encontrado: $(${OPA_BIN} version)${RESET}"
    fi
}

# Criar diretórios necessários
create_dirs() {
    echo -e "${BLUE}======================================${RESET}"
    echo -e "${BLUE}Criando diretórios necessários...${RESET}"
    echo -e "${BLUE}======================================${RESET}"
    
    mkdir -p "${OUTPUT_DIR}"
    mkdir -p "${COVERAGE_DIR}"
}

# Validar políticas
validate_policies() {
    echo -e "${BLUE}======================================${RESET}"
    echo -e "${BLUE}Validando políticas Rego...${RESET}"
    echo -e "${BLUE}======================================${RESET}"
    
    POLICY_FILES=$(find "${POLICY_DIR}" -name "*.rego" -type f | grep -v "_test\.rego$")
    for policy in ${POLICY_FILES}; do
        echo -e "${YELLOW}Validando ${policy}...${RESET}"
        ${OPA_BIN} check "${policy}" && echo -e "${GREEN}✓ OK: ${policy}${RESET}" || { 
            echo -e "${RED}✗ FALHA: ${policy}${RESET}"
            exit 1
        }
    done
}

# Formatar políticas
format_policies() {
    echo -e "${BLUE}======================================${RESET}"
    echo -e "${BLUE}Formatando políticas Rego...${RESET}"
    echo -e "${BLUE}======================================${RESET}"
    
    POLICY_FILES=$(find "${POLICY_DIR}" -name "*.rego" -type f)
    for policy in ${POLICY_FILES}; do
        echo -e "${YELLOW}Formatando ${policy}...${RESET}"
        ${OPA_BIN} fmt -w "${policy}"
    done
}

# Executar testes
run_tests() {
    echo -e "${BLUE}======================================${RESET}"
    echo -e "${BLUE}Executando testes...${RESET}"
    echo -e "${BLUE}======================================${RESET}"
    
    cd "${POLICY_DIR}" && ${OPA_BIN} test . --verbose
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Todos os testes passaram${RESET}"
    else
        echo -e "${RED}✗ Falha nos testes${RESET}"
        exit 1
    fi
}

# Gerar relatório de cobertura
generate_coverage() {
    echo -e "${BLUE}======================================${RESET}"
    echo -e "${BLUE}Gerando relatório de cobertura...${RESET}"
    echo -e "${BLUE}======================================${RESET}"
    
    cd "${POLICY_DIR}" && ${OPA_BIN} test . --coverage --format=json > "${COVERAGE_DIR}/coverage_${TIMESTAMP}.json"
    
    echo -e "${GREEN}✓ Relatório de cobertura gerado em ${COVERAGE_DIR}/coverage_${TIMESTAMP}.json${RESET}"
    
    # Opcional: Converter JSON para formato mais legível
    if command -v jq &> /dev/null; then
        jq . "${COVERAGE_DIR}/coverage_${TIMESTAMP}.json" > "${COVERAGE_DIR}/coverage_${TIMESTAMP}_pretty.json"
    fi
}

# Criar bundle de políticas
create_bundle() {
    echo -e "${BLUE}======================================${RESET}"
    echo -e "${BLUE}Criando bundle de políticas...${RESET}"
    echo -e "${BLUE}======================================${RESET}"
    
    cd "${POLICY_DIR}" && ${OPA_BIN} build -b . -o "${OUTPUT_DIR}/${BUNDLE_NAME}"
    
    echo -e "${GREEN}✓ Bundle criado: ${OUTPUT_DIR}/${BUNDLE_NAME}${RESET}"
}

# Verificar vulnerabilidades nas políticas
check_vulnerabilities() {
    echo -e "${BLUE}======================================${RESET}"
    echo -e "${BLUE}Verificando vulnerabilidades...${RESET}"
    echo -e "${BLUE}======================================${RESET}"
    
    echo -e "${YELLOW}Procurando por regras que permitam acesso irrestrito...${RESET}"
    POLICY_FILES=$(find "${POLICY_DIR}" -name "*.rego" -type f)
    
    # Verificar regras de permissão irrestrita
    if grep -r "allow { true }" "${POLICY_DIR}" --include="*.rego"; then
        echo -e "${RED}✗ ALERTA: Encontradas regras de permissão irrestrita!${RESET}"
        exit 1
    else
        echo -e "${GREEN}✓ Nenhuma regra de permissão irrestrita encontrada${RESET}"
    fi
    
    # Verificar regras de negação explícita
    for policy in ${POLICY_FILES}; do
        if grep -q "allow {" "${policy}" && ! grep -q "allow = false" "${policy}" && ! grep -q "deny {" "${policy}"; then
            echo -e "${YELLOW}⚠ ATENÇÃO: ${policy} pode não ter regras de negação explícitas${RESET}"
        fi
    done
}

# Implantar políticas
deploy_policies() {
    echo -e "${BLUE}======================================${RESET}"
    echo -e "${BLUE}Implantando políticas para ${ENV}...${RESET}"
    echo -e "${BLUE}======================================${RESET}"
    
    case ${DEPLOY_TARGET} in
        "local")
            echo -e "${YELLOW}Implantando para OPA local...${RESET}"
            # Copiar bundle para diretório local do OPA
            DEST_DIR="/etc/opa/bundles"
            if [ -d "${DEST_DIR}" ]; then
                cp "${OUTPUT_DIR}/${BUNDLE_NAME}" "${DEST_DIR}/"
                echo -e "${GREEN}✓ Bundle implantado para OPA local${RESET}"
            else
                echo -e "${RED}✗ Diretório de destino não encontrado${RESET}"
                exit 1
            fi
            ;;
            
        "kubernetes")
            echo -e "${YELLOW}Implantando para Kubernetes...${RESET}"
            # Atualizar ConfigMap no Kubernetes
            kubectl create configmap iam-role-policies --from-file="${OUTPUT_DIR}/${BUNDLE_NAME}" \
                -n innovabiz-iam --dry-run=client -o yaml | kubectl apply -f -
            
            # Recarregar pods OPA
            kubectl rollout restart deployment opa-server -n innovabiz-iam
            echo -e "${GREEN}✓ Bundle implantado para Kubernetes${RESET}"
            ;;
            
        "cloud")
            echo -e "${YELLOW}Implantando para servidor de configuração na nuvem...${RESET}"
            # Exemplo: Fazer upload para bucket S3 ou similar
            if command -v aws &> /dev/null; then
                aws s3 cp "${OUTPUT_DIR}/${BUNDLE_NAME}" "s3://innovabiz-opa-bundles/${ENV}/iam/role/"
                echo -e "${GREEN}✓ Bundle implantado para a nuvem${RESET}"
            else
                echo -e "${RED}✗ AWS CLI não encontrado${RESET}"
                exit 1
            fi
            ;;
            
        *)
            echo -e "${RED}✗ Destino de implantação desconhecido: ${DEPLOY_TARGET}${RESET}"
            echo -e "${YELLOW}Destinos válidos: local, kubernetes, cloud${RESET}"
            exit 1
            ;;
    esac
}

# Verificar conformidade
check_compliance() {
    echo -e "${BLUE}======================================${RESET}"
    echo -e "${BLUE}Verificando conformidade...${RESET}"
    echo -e "${BLUE}======================================${RESET}"
    
    echo -e "${YELLOW}Verificando conformidade com ISO/IEC 27001:2022...${RESET}"
    grep -q "ISO/IEC 27001" "${POLICY_DIR}" -r --include="*.rego" && \
        echo -e "${GREEN}✓ Referência à ISO/IEC 27001 encontrada${RESET}" || \
        echo -e "${RED}✗ Referência à ISO/IEC 27001 não encontrada${RESET}"
    
    echo -e "${YELLOW}Verificando mecanismos de auditoria...${RESET}"
    grep -q "audit" "${POLICY_DIR}" -r --include="*.rego" && \
        echo -e "${GREEN}✓ Referências a auditoria encontradas${RESET}" || \
        echo -e "${RED}✗ Referências a auditoria não encontradas${RESET}"
    
    echo -e "${YELLOW}Verificando mecanismos de multitenancy...${RESET}"
    grep -q "tenant" "${POLICY_DIR}" -r --include="*.rego" && \
        echo -e "${GREEN}✓ Referências a tenant encontradas${RESET}" || \
        echo -e "${RED}✗ Referências a tenant não encontradas${RESET}"
}

# Função principal
main() {
    echo -e "${BLUE}======================================${RESET}"
    echo -e "${BLUE}Iniciando CI/CD para políticas OPA${RESET}"
    echo -e "${BLUE}Ambiente: ${ENV}${RESET}"
    echo -e "${BLUE}Destino: ${DEPLOY_TARGET}${RESET}"
    echo -e "${BLUE}======================================${RESET}"
    
    check_opa
    create_dirs
    format_policies
    validate_policies
    check_vulnerabilities
    run_tests
    check_compliance
    generate_coverage
    create_bundle
    deploy_policies
    
    echo -e "${BLUE}======================================${RESET}"
    echo -e "${GREEN}✓ CI/CD concluído com sucesso!${RESET}"
    echo -e "${BLUE}======================================${RESET}"
}

# Executar função principal
main