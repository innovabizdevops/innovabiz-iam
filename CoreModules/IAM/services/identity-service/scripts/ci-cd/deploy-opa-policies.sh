#!/bin/bash
# =============================================================================
# Script de Deploy Automatizado de Políticas OPA para INNOVABIZ IAM
# =============================================================================
# Autor: INNOVABIZ DevSecOps Team
# Data: 2025-08-05
# Versão: 1.0.0
# Descrição: Este script automatiza o processo de implantação de políticas OPA
#            para diferentes ambientes (dev, qa, hml, prod), com suporte a
#            configurações multi-tenant e multi-região. Inclui validação,
#            versionamento, rollout seguro e verificação pós-deploy.
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

# Parâmetros de entrada
POLICIES_DIR="${1:-"../policies"}"
ENV="${2:-"dev"}"
NAMESPACE="${3:-"innovabiz-iam"}"
BUNDLE_NAME="${4:-"innovabiz-iam-policies"}"
VERSION="${5:-"$(date +%Y%m%d%H%M%S)"}"

# Diretórios e arquivos
TEMP_DIR="/tmp/opa-bundle-${VERSION}"
BUNDLE_FILE="${TEMP_DIR}/${BUNDLE_NAME}.tar.gz"
MANIFEST_DIR="./k8s-manifests"
LOG_DIR="./deploy-logs"
LOG_FILE="${LOG_DIR}/deploy_${ENV}_${VERSION}.log"

# Configurações de ambiente
case "${ENV}" in
  "dev")
    CONFIG_SUFFIX="dev"
    ROLLOUT_STRATEGY="immediate"
    VALIDATE_ONLY="false"
    ROLLBACK_ON_ERROR="true"
    ;;
  "qa")
    CONFIG_SUFFIX="qa"
    ROLLOUT_STRATEGY="phased"
    VALIDATE_ONLY="false"
    ROLLBACK_ON_ERROR="true"
    ;;
  "hml")
    CONFIG_SUFFIX="hml"
    ROLLOUT_STRATEGY="phased"
    VALIDATE_ONLY="false"
    ROLLBACK_ON_ERROR="true"
    ;;
  "prod")
    CONFIG_SUFFIX="prod"
    ROLLOUT_STRATEGY="canary"
    VALIDATE_ONLY="false"
    ROLLBACK_ON_ERROR="true"
    ;;
  *)
    echo -e "${RED}[ERRO]${NC} Ambiente inválido: ${ENV}"
    echo "Ambientes suportados: dev, qa, hml, prod"
    exit 1
    ;;
esac

# Inicialização
mkdir -p "${TEMP_DIR}" "${MANIFEST_DIR}" "${LOG_DIR}"

echo -e "${BLUE}[INFO]${NC} Iniciando deploy de políticas OPA para ambiente ${ENV}"
echo -e "${BLUE}[INFO]${NC} Namespace: ${NAMESPACE}"
echo -e "${BLUE}[INFO]${NC} Versão do bundle: ${VERSION}"
echo -e "${BLUE}[INFO]${NC} Estratégia de rollout: ${ROLLOUT_STRATEGY}"
echo -e "${BLUE}[INFO]${NC} Logs em: ${LOG_FILE}"

{
  echo "======================================================="
  echo "INNOVABIZ IAM - Deploy de Políticas OPA"
  echo "======================================================="
  echo "Data de execução: $(date)"
  echo "Ambiente: ${ENV}"
  echo "Namespace: ${NAMESPACE}"
  echo "Bundle: ${BUNDLE_NAME}"
  echo "Versão: ${VERSION}"
  echo "======================================================="
} > "${LOG_FILE}"

# Função para verificar ferramentas necessárias
check_prerequisites() {
  echo -e "${CYAN}[TAREFA]${NC} Verificando pré-requisitos..."
  echo -e "\n## Verificação de Pré-requisitos" >> "${LOG_FILE}"
  
  local missing=false
  
  for cmd in opa kubectl tar gzip jq; do
    if ! command -v "${cmd}" &> /dev/null; then
      echo -e "${RED}[ERRO]${NC} Comando não encontrado: ${cmd}"
      echo "Comando não encontrado: ${cmd}" >> "${LOG_FILE}"
      missing=true
    else
      echo -e "${GREEN}[OK]${NC} ${cmd} encontrado: $(${cmd} --version 2>&1 | head -1)"
      echo "${cmd} encontrado: $(${cmd} --version 2>&1 | head -1)" >> "${LOG_FILE}"
    fi
  done
  
  if ${missing}; then
    echo -e "${RED}[ERRO]${NC} Ferramentas necessárias não encontradas"
    return 1
  fi
  
  # Verificar se está autenticado no cluster
  if ! kubectl get ns &> /dev/null; then
    echo -e "${RED}[ERRO]${NC} Não autenticado no cluster Kubernetes"
    echo "Não autenticado no cluster Kubernetes" >> "${LOG_FILE}"
    return 1
  fi
  
  echo -e "${GREEN}[SUCESSO]${NC} Todos os pré-requisitos verificados"
  return 0
}

# Função para validar políticas antes do deploy
validate_policies() {
  echo -e "${CYAN}[TAREFA]${NC} Validando políticas antes do deploy..."
  echo -e "\n## Validação de Políticas" >> "${LOG_FILE}"
  
  # Executar script de validação se existir
  if [ -f "./validate-policies.sh" ]; then
    echo -e "${BLUE}[INFO]${NC} Executando script de validação..."
    
    if ! bash ./validate-policies.sh "${POLICIES_DIR}" &>> "${LOG_FILE}"; then
      echo -e "${RED}[ERRO]${NC} Validação de políticas falhou"
      echo "Validação falhou, verifique os logs para detalhes" >> "${LOG_FILE}"
      return 1
    fi
  else
    # Validação básica se o script completo não estiver disponível
    echo -e "${BLUE}[INFO]${NC} Script de validação não encontrado, executando verificação básica..."
    
    local files=$(find "${POLICIES_DIR}" -name "*.rego" | sort)
    local errors=0
    
    for file in ${files}; do
      if ! opa check "${file}" &>> "${LOG_FILE}"; then
        echo -e "${RED}[ERRO]${NC} ${file} - Erro de sintaxe"
        errors=$((errors + 1))
      fi
    done
    
    if [ ${errors} -gt 0 ]; then
      echo -e "${RED}[ERRO]${NC} ${errors} políticas apresentaram erros sintáticos"
      return 1
    fi
  fi
  
  echo -e "${GREEN}[SUCESSO]${NC} Validação de políticas concluída com sucesso"
  return 0
}

# Função para criar bundle de políticas
create_policy_bundle() {
  echo -e "${CYAN}[TAREFA]${NC} Criando bundle de políticas..."
  echo -e "\n## Criação do Bundle de Políticas" >> "${LOG_FILE}"
  
  # Preparar manifest do bundle
  local manifest_file="${TEMP_DIR}/.manifest"
  
  cat > "${manifest_file}" << EOF
{
  "revision": "${VERSION}",
  "roots": ["innovabiz", "system"],
  "metadata": {
    "environment": "${ENV}",
    "created_at": "$(date -Iseconds)",
    "created_by": "$(whoami)",
    "bundle_name": "${BUNDLE_NAME}",
    "description": "INNOVABIZ IAM Policies Bundle"
  }
}
EOF

  # Copiar políticas para diretório temporário
  echo -e "${BLUE}[INFO]${NC} Copiando políticas para o diretório temporário..."
  cp -r "${POLICIES_DIR}"/* "${TEMP_DIR}/"
  
  # Verificar se há configurações específicas de ambiente e aplicá-las
  local env_config="${POLICIES_DIR}/config/config.${CONFIG_SUFFIX}.json"
  if [ -f "${env_config}" ]; then
    echo -e "${BLUE}[INFO]${NC} Aplicando configuração específica para ambiente ${ENV}..."
    cp "${env_config}" "${TEMP_DIR}/config.json"
  else
    echo -e "${YELLOW}[AVISO]${NC} Configuração específica para ${ENV} não encontrada, usando padrão"
    if [ -f "${POLICIES_DIR}/config/config.default.json" ]; then
      cp "${POLICIES_DIR}/config/config.default.json" "${TEMP_DIR}/config.json"
    fi
  fi
  
  # Criar bundle tar.gz
  echo -e "${BLUE}[INFO]${NC} Criando arquivo bundle ${BUNDLE_FILE}..."
  tar -czf "${BUNDLE_FILE}" -C "${TEMP_DIR}" . &>> "${LOG_FILE}"
  
  if [ -f "${BUNDLE_FILE}" ]; then
    local size=$(du -h "${BUNDLE_FILE}" | cut -f1)
    echo -e "${GREEN}[SUCESSO]${NC} Bundle criado com sucesso (${size})"
    echo "Bundle criado: ${BUNDLE_FILE} (${size})" >> "${LOG_FILE}"
    return 0
  else
    echo -e "${RED}[ERRO]${NC} Falha ao criar bundle"
    return 1
  fi
}

# Função para gerar ConfigMap do bundle
create_bundle_configmap() {
  echo -e "${CYAN}[TAREFA]${NC} Gerando ConfigMap do bundle..."
  echo -e "\n## Geração do ConfigMap" >> "${LOG_FILE}"
  
  local configmap_name="${BUNDLE_NAME}-${VERSION}"
  local configmap_file="${MANIFEST_DIR}/${configmap_name}.yaml"
  
  # Codificar bundle em base64
  local bundle_b64=$(base64 -w 0 "${BUNDLE_FILE}")
  
  # Criar ConfigMap
  cat > "${configmap_file}" << EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: ${configmap_name}
  namespace: ${NAMESPACE}
  labels:
    app: opa
    innovabiz.com/component: iam
    innovabiz.com/bundle-version: "${VERSION}"
    innovabiz.com/environment: "${ENV}"
data:
  bundle.tar.gz.b64: "${bundle_b64}"
EOF

  if [ -f "${configmap_file}" ]; then
    echo -e "${GREEN}[SUCESSO]${NC} ConfigMap gerado em ${configmap_file}"
    echo "ConfigMap gerado: ${configmap_file}" >> "${LOG_FILE}"
    return 0
  else
    echo -e "${RED}[ERRO]${NC} Falha ao gerar ConfigMap"
    return 1
  fi
}

# Função para atualizar ConfigMap no Kubernetes
apply_bundle_configmap() {
  echo -e "${CYAN}[TAREFA]${NC} Aplicando ConfigMap no Kubernetes..."
  echo -e "\n## Aplicação do ConfigMap" >> "${LOG_FILE}"
  
  local configmap_name="${BUNDLE_NAME}-${VERSION}"
  local configmap_file="${MANIFEST_DIR}/${configmap_name}.yaml"
  
  if [ "${VALIDATE_ONLY}" = "true" ]; then
    echo -e "${YELLOW}[AVISO]${NC} Modo de validação ativado, não aplicando recursos"
    kubectl apply --dry-run=client -f "${configmap_file}" &>> "${LOG_FILE}"
    echo -e "${GREEN}[SUCESSO]${NC} Validação do ConfigMap concluída"
    return 0
  fi
  
  # Verificar se namespace existe
  if ! kubectl get namespace "${NAMESPACE}" &> /dev/null; then
    echo -e "${BLUE}[INFO]${NC} Criando namespace ${NAMESPACE}..."
    kubectl create namespace "${NAMESPACE}" &>> "${LOG_FILE}"
  fi
  
  # Aplicar ConfigMap
  if kubectl apply -f "${configmap_file}" &>> "${LOG_FILE}"; then
    echo -e "${GREEN}[SUCESSO]${NC} ConfigMap aplicado com sucesso"
    return 0
  else
    echo -e "${RED}[ERRO]${NC} Falha ao aplicar ConfigMap"
    return 1
  fi
}

# Função para atualizar configuração do OPA
update_opa_configuration() {
  echo -e "${CYAN}[TAREFA]${NC} Atualizando configuração do OPA..."
  echo -e "\n## Atualização da Configuração do OPA" >> "${LOG_FILE}"
  
  local configmap_name="${BUNDLE_NAME}-${VERSION}"
  local opa_config_file="${MANIFEST_DIR}/opa-config-${ENV}-${VERSION}.yaml"
  
  # Criar/atualizar ConfigMap de configuração do OPA
  cat > "${opa_config_file}" << EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: opa-config
  namespace: ${NAMESPACE}
  labels:
    innovabiz.com/component: iam
    innovabiz.com/environment: "${ENV}"
data:
  config.yaml: |
    services:
      innovabiz-iam-policy:
        url: http://innovabiz-iam-policy-api.${NAMESPACE}.svc.cluster.local:8181
        credentials:
          bearer:
            token_path: /var/run/secrets/tokens/opa-token
        response_header_timeout_seconds: 5
    bundles:
      innovabiz-iam:
        resource: configmap://${NAMESPACE}/${configmap_name}
        persist: true
        polling:
          min_delay_seconds: 10
          max_delay_seconds: 60
    decision_logs:
      console: true
      service: innovabiz-iam-policy
      reporting:
        min_delay_seconds: 5
        max_delay_seconds: 30
    status:
      service: innovabiz-iam-policy
    plugins:
      envoy_ext_authz_grpc:
        addr: :9191
        path: innovabiz/iam/authz/allow
EOF

  if [ "${VALIDATE_ONLY}" = "true" ]; then
    kubectl apply --dry-run=client -f "${opa_config_file}" &>> "${LOG_FILE}"
    echo -e "${GREEN}[SUCESSO]${NC} Validação da configuração do OPA concluída"
    return 0
  fi
  
  # Aplicar configuração
  if kubectl apply -f "${opa_config_file}" &>> "${LOG_FILE}"; then
    echo -e "${GREEN}[SUCESSO]${NC} Configuração do OPA atualizada com sucesso"
    return 0
  else
    echo -e "${RED}[ERRO]${NC} Falha ao atualizar configuração do OPA"
    return 1
  fi
}

# Função para realizar rollout do OPA
rollout_opa() {
  echo -e "${CYAN}[TAREFA]${NC} Realizando rollout do OPA..."
  echo -e "\n## Rollout do OPA" >> "${LOG_FILE}"
  
  # Identificar deployments do OPA
  local opa_deployments=$(kubectl get deployments -n "${NAMESPACE}" -l app=opa -o name)
  
  if [ -z "${opa_deployments}" ]; then
    echo -e "${YELLOW}[AVISO]${NC} Nenhum deployment do OPA encontrado no namespace ${NAMESPACE}"
    echo "Nenhum deployment do OPA encontrado" >> "${LOG_FILE}"
    return 0
  fi
  
  if [ "${ROLLOUT_STRATEGY}" = "canary" ]; then
    echo -e "${BLUE}[INFO]${NC} Realizando rollout canário..."
    
    # Em ambiente de produção, fazemos rollout gradual
    for deployment in ${opa_deployments}; do
      local deploy_name=$(echo "${deployment}" | cut -d'/' -f2)
      
      echo -e "${BLUE}[INFO]${NC} Rollout canário para ${deploy_name}..."
      echo "Iniciando rollout canário para ${deploy_name}" >> "${LOG_FILE}"
      
      # Adicionar anotação para trigger do rollout
      kubectl patch ${deployment} -n "${NAMESPACE}" -p "{\"spec\":{\"template\":{\"metadata\":{\"annotations\":{\"innovabiz.com/config-version\":\"${VERSION}\"}}}}}" &>> "${LOG_FILE}"
      
      # Verificar se o deployment está em rollout
      kubectl rollout status ${deployment} -n "${NAMESPACE}" --timeout=5m &>> "${LOG_FILE}"
      
      if [ $? -ne 0 ]; then
        echo -e "${RED}[ERRO]${NC} Timeout no rollout do ${deploy_name}"
        
        if [ "${ROLLBACK_ON_ERROR}" = "true" ]; then
          echo -e "${YELLOW}[AVISO]${NC} Realizando rollback do ${deploy_name}..."
          kubectl rollout undo ${deployment} -n "${NAMESPACE}" &>> "${LOG_FILE}"
        fi
        
        return 1
      fi
      
      # Pausa entre deployments para verificar estabilidade
      echo -e "${BLUE}[INFO]${NC} Aguardando 30 segundos para monitorar estabilidade..."
      sleep 30
      
      # Verificar status do deployment após rollout
      if ! kubectl get ${deployment} -n "${NAMESPACE}" -o jsonpath='{.status.conditions[?(@.type=="Available")].status}' | grep -q "True"; then
        echo -e "${RED}[ERRO]${NC} Deployment ${deploy_name} não está disponível após rollout"
        
        if [ "${ROLLBACK_ON_ERROR}" = "true" ]; then
          echo -e "${YELLOW}[AVISO]${NC} Realizando rollback do ${deploy_name}..."
          kubectl rollout undo ${deployment} -n "${NAMESPACE}" &>> "${LOG_FILE}"
        fi
        
        return 1
      fi
    done
  else
    # Para ambientes não-prod, fazemos rollout imediato
    for deployment in ${opa_deployments}; do
      local deploy_name=$(echo "${deployment}" | cut -d'/' -f2)
      
      echo -e "${BLUE}[INFO]${NC} Realizando restart do ${deploy_name}..."
      kubectl rollout restart ${deployment} -n "${NAMESPACE}" &>> "${LOG_FILE}"
      kubectl rollout status ${deployment} -n "${NAMESPACE}" --timeout=2m &>> "${LOG_FILE}"
    done
  fi
  
  echo -e "${GREEN}[SUCESSO]${NC} Rollout concluído com sucesso"
  return 0
}

# Função para verificar saúde do OPA após deploy
verify_opa_health() {
  echo -e "${CYAN}[TAREFA]${NC} Verificando saúde do OPA após deploy..."
  echo -e "\n## Verificação de Saúde" >> "${LOG_FILE}"
  
  # Verificar pods do OPA
  local opa_pods=$(kubectl get pods -n "${NAMESPACE}" -l app=opa -o name)
  
  if [ -z "${opa_pods}" ]; then
    echo -e "${RED}[ERRO]${NC} Nenhum pod do OPA encontrado"
    return 1
  fi
  
  # Verificar status dos pods
  local unhealthy_pods=0
  
  for pod in ${opa_pods}; do
    local pod_name=$(echo "${pod}" | cut -d'/' -f2)
    local status=$(kubectl get ${pod} -n "${NAMESPACE}" -o jsonpath='{.status.phase}')
    local ready=$(kubectl get ${pod} -n "${NAMESPACE}" -o jsonpath='{.status.containerStatuses[0].ready}')
    
    echo -e "${BLUE}[INFO]${NC} Pod ${pod_name}: Status=${status}, Ready=${ready}"
    echo "Pod ${pod_name}: Status=${status}, Ready=${ready}" >> "${LOG_FILE}"
    
    if [ "${status}" != "Running" ] || [ "${ready}" != "true" ]; then
      echo -e "${RED}[ERRO]${NC} Pod ${pod_name} não está saudável"
      unhealthy_pods=$((unhealthy_pods + 1))
      
      # Obter logs do pod
      kubectl logs ${pod} -n "${NAMESPACE}" --tail=50 &>> "${LOG_FILE}"
    fi
  done
  
  if [ ${unhealthy_pods} -gt 0 ]; then
    echo -e "${RED}[ERRO]${NC} ${unhealthy_pods} pods não estão saudáveis"
    return 1
  fi
  
  # Testar API de health do OPA
  local health_ok=false
  local opa_svc=$(kubectl get svc -n "${NAMESPACE}" -l app=opa -o name | head -1)
  
  if [ -n "${opa_svc}" ]; then
    local svc_name=$(echo "${opa_svc}" | cut -d'/' -f2)
    
    # Criar pod temporário para testar o serviço
    echo -e "${BLUE}[INFO]${NC} Testando API de health do OPA via ${svc_name}..."
    
    kubectl run opa-healthcheck --rm --restart=Never -n "${NAMESPACE}" --image=curlimages/curl -i --timeout=30 -- \
      curl -s -o /dev/null -w "%{http_code}" http://${svc_name}:8181/health > /tmp/opa_health_result 2>> "${LOG_FILE}"
    
    local http_code=$(cat /tmp/opa_health_result)
    
    if [ "${http_code}" = "200" ]; then
      echo -e "${GREEN}[SUCESSO]${NC} API de health do OPA está respondendo corretamente"
      health_ok=true
    else
      echo -e "${RED}[ERRO]${NC} API de health do OPA retornou código ${http_code}"
    fi
  else
    echo -e "${YELLOW}[AVISO]${NC} Nenhum serviço do OPA encontrado, não é possível testar a API de health"
  fi
  
  if [ "${health_ok}" = "true" ] && [ ${unhealthy_pods} -eq 0 ]; then
    echo -e "${GREEN}[SUCESSO]${NC} OPA está saudável após o deploy"
    return 0
  else
    echo -e "${RED}[ERRO]${NC} OPA apresenta problemas após o deploy"
    return 1
  fi
}

# Função para registrar a versão do deploy
register_deployment() {
  echo -e "${CYAN}[TAREFA]${NC} Registrando deploy..."
  echo -e "\n## Registro do Deploy" >> "${LOG_FILE}"
  
  local registry_file="./deploy-registry.json"
  local registry_data=()
  
  # Carregar registro existente se houver
  if [ -f "${registry_file}" ]; then
    registry_data=$(cat "${registry_file}")
  else
    registry_data='[]'
  fi
  
  # Adicionar nova entrada
  local new_entry=$(cat << EOF
{
  "version": "${VERSION}",
  "environment": "${ENV}",
  "timestamp": "$(date -Iseconds)",
  "bundle": "${BUNDLE_NAME}",
  "user": "$(whoami)",
  "status": "success",
  "log_file": "${LOG_FILE}"
}
EOF
)
  
  # Atualizar registro
  local updated_registry=$(echo "${registry_data}" | jq ". + [${new_entry}]")
  echo "${updated_registry}" > "${registry_file}"
  
  echo -e "${GREEN}[SUCESSO]${NC} Deploy registrado em ${registry_file}"
  return 0
}

# Função para limpar arquivos temporários
cleanup() {
  echo -e "${CYAN}[TAREFA]${NC} Limpando arquivos temporários..."
  
  # Manter bundle por enquanto para possível rollback
  # rm -rf "${TEMP_DIR}"
  
  echo -e "${GREEN}[SUCESSO]${NC} Limpeza concluída"
}

# Execução principal
main() {
  echo -e "${PURPLE}=========================================================${NC}"
  echo -e "${PURPLE}  INNOVABIZ IAM - Deploy de Políticas OPA (${ENV})  ${NC}"
  echo -e "${PURPLE}=========================================================${NC}"
  
  local start_time=$(date +%s)
  local success=true
  
  # Executar etapas do deploy
  if ! check_prerequisites; then
    echo -e "${RED}[ERRO CRÍTICO]${NC} Pré-requisitos não atendidos, abortando deploy"
    exit 1
  fi
  
  if ! validate_policies; then
    echo -e "${RED}[ERRO CRÍTICO]${NC} Validação de políticas falhou, abortando deploy"
    exit 1
  fi
  
  # Criar e aplicar bundle
  create_policy_bundle || success=false
  
  if [ "${success}" = "true" ]; then
    create_bundle_configmap || success=false
  fi
  
  if [ "${success}" = "true" ]; then
    apply_bundle_configmap || success=false
  fi
  
  if [ "${success}" = "true" ]; then
    update_opa_configuration || success=false
  fi
  
  if [ "${success}" = "true" ]; then
    rollout_opa || success=false
  fi
  
  if [ "${success}" = "true" ]; then
    verify_opa_health || success=false
  fi
  
  # Finalização
  if [ "${success}" = "true" ]; then
    register_deployment
    cleanup
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    echo -e "${PURPLE}=========================================================${NC}"
    echo -e "${BLUE}[INFO]${NC} Duração total: ${duration} segundos"
    echo -e "${GREEN}[SUCESSO]${NC} Deploy de políticas OPA concluído com sucesso"
    echo -e "${BLUE}[INFO]${NC} Versão: ${VERSION}"
    echo -e "${PURPLE}=========================================================${NC}"
    
    # Resumo do deploy
    cat << EOF >> "${LOG_FILE}"

=====================================================
RESUMO DO DEPLOY
=====================================================
Status: SUCESSO
Versão: ${VERSION}
Ambiente: ${ENV}
Bundle: ${BUNDLE_NAME}
Duração: ${duration} segundos
Concluído em: $(date)
=====================================================
EOF
    
    exit 0
  else
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    echo -e "${PURPLE}=========================================================${NC}"
    echo -e "${BLUE}[INFO]${NC} Duração total: ${duration} segundos"
    echo -e "${RED}[ERRO]${NC} Deploy de políticas OPA falhou"
    echo -e "${PURPLE}=========================================================${NC}"
    
    # Resumo do deploy
    cat << EOF >> "${LOG_FILE}"

=====================================================
RESUMO DO DEPLOY
=====================================================
Status: FALHA
Versão: ${VERSION}
Ambiente: ${ENV}
Bundle: ${BUNDLE_NAME}
Duração: ${duration} segundos
Concluído em: $(date)
=====================================================
EOF
    
    exit 1
  fi
}

# Executar script
main