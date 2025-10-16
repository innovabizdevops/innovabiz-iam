# Guia de Implementação da Stack de Observabilidade INNOVABIZ

![INNOVABIZ Logo](../../../assets/innovabiz-logo.png)

**Versão:** 1.0.0  
**Data de Atualização:** 31/07/2025  
**Classificação:** Oficial  
**Autor:** Equipe INNOVABIZ DevOps  
**Aprovado por:** Eduardo Jeremias  
**E-mail:** innovabizdevops@gmail.com

## Índice

1. [Visão Geral](#1-visão-geral)
2. [Pré-requisitos](#2-pré-requisitos)
3. [Arquitetura de Referência](#3-arquitetura-de-referência)
4. [Implantação por Etapas](#4-implantação-por-etapas)
5. [Configuração Multi-Dimensional](#5-configuração-multi-dimensional)
6. [Integração com IAM Service](#6-integração-com-iam-service)
7. [Validação e Testes](#7-validação-e-testes)
8. [Migração de Dados](#8-migração-de-dados)
9. [Troubleshooting Comum](#9-troubleshooting-comum)
10. [Referências](#10-referências)

## 1. Visão Geral

Este guia fornece instruções detalhadas para a implementação da Stack de Observabilidade da plataforma INNOVABIZ, garantindo alinhamento com os princípios de design multi-tenant, multi-regional e multi-dimensional, bem como conformidade com as regulações aplicáveis (LGPD, GDPR, PCI DSS, SOC 2, ISO 27001).

A stack de observabilidade é composta pelos seguintes componentes principais:
- **OpenTelemetry Collector**: Coleta e normalização de telemetria
- **Prometheus**: Coleta e armazenamento de métricas
- **Loki**: Agregação e indexação de logs
- **Elasticsearch**: Armazenamento e análise avançada de logs
- **Jaeger**: Rastreamento distribuído
- **Grafana**: Visualização de métricas e logs
- **Kibana**: Visualização e análise avançada de logs
- **Fluentd**: Coleta e processamento de logs
- **AlertManager**: Gerenciamento de alertas
- **Observability Portal**: Interface unificada de observabilidade

## 2. Pré-requisitos

### 2.1 Infraestrutura Kubernetes

- Cluster Kubernetes v1.26+ com pelo menos 3 nós worker
- Capacidade de recursos:
  - CPU: Mínimo de 16 vCPUs disponíveis
  - Memória: Mínimo de 64GB disponíveis
  - Armazenamento: Mínimo de 500GB disponíveis (SSD recomendado)
- **Classes de Armazenamento:**
  - StorageClass com suporte a RWO (ReadWriteOnce)
  - StorageClass com suporte a RWX (ReadWriteMany)
- **Namespace dedicado:** `innovabiz-observability`

### 2.2 Requisitos de Rede

- Ingress Controller configurado (NGINX ou similar)
- Certificado TLS válido para os domínios:
  - `observability-portal.{AMBIENTE}.innovabiz.com`
  - `grafana.{AMBIENTE}.innovabiz.com`
  - `prometheus.{AMBIENTE}.innovabiz.com`
  - `alertmanager.{AMBIENTE}.innovabiz.com`
  - `kibana.{AMBIENTE}.innovabiz.com`
  - `jaeger.{AMBIENTE}.innovabiz.com`
  - `elasticsearch.{AMBIENTE}.innovabiz.com`
- Network Policies habilitadas no cluster

### 2.3 Ferramentas de Cliente

- `kubectl` v1.26+
- Helm v3.10+
- `openssl` para geração de certificados
- `jq` para manipulação de JSON

### 2.4 Integração com IAM

- IAM Service operacional com suporte OIDC
- Grupo e papéis de acesso configurados para observabilidade
- Certificados para autenticação mTLS

## 3. Arquitetura de Referência

### 3.1 Diagrama de Arquitetura

```
+------------------------------------------------+
|                Ingress / API Gateway            |
+------------------------------------------------+
           |                |               |
           v                v               v
+----------------+  +-----------------+  +----------------+
| Observability  |  |   Monitoring    |  |    Logging     |
|    Portal      |  |     Stack       |  |     Stack      |
|                |  |                 |  |                |
| - React UI     |  | - Prometheus    |  | - Elasticsearch|
| - Node.js API  |  | - Grafana       |  | - Kibana       |
| - Federation   |  | - AlertManager  |  | - Fluentd      |
|   Service      |  | - OTel Collector|  | - Loki         |
+----------------+  +-----------------+  +----------------+
           |                |               |
           v                v               v
+------------------------------------------------+
|               INNOVABIZ IAM Service             |
|          (Authentication & Authorization)        |
+------------------------------------------------+
           |                |               |
           v                v               v
+----------------+  +-----------------+  +----------------+
|   Payment      |  |      IAM        |  |    Other       |
|   Gateway      |  |    Services     |  |   Modules      |
+----------------+  +-----------------+  +----------------+
```

### 3.2 Comunicação Entre Componentes

- **Ingress → Serviços Frontais**: HTTPS (TLS 1.3)
- **Entre Componentes**: mTLS
- **Para Serviços INNOVABIZ**: mTLS + JWT
- **Para Armazenamento**: TLS + Autenticação

## 4. Implantação por Etapas

### 4.1 Criação de Namespace e RBAC

```bash
kubectl create namespace innovabiz-observability

# Aplicar RBAC para observabilidade
kubectl apply -f rbac/observability-rbac.yaml
```

### 4.2 Configuração de Segredos e ConfigMaps

```bash
# Gerar certificados mTLS
./scripts/generate-mtls-certs.sh

# Criar segredos para certificados
kubectl create secret tls innovabiz-observability-tls \
  --cert=certs/server.crt \
  --key=certs/server.key \
  --namespace=innovabiz-observability

# Criar secrets para credenciais
kubectl create secret generic observability-credentials \
  --from-literal=elasticsearch-password=$(openssl rand -base64 32) \
  --from-literal=grafana-admin-password=$(openssl rand -base64 32) \
  --from-literal=kibana-encryption-key=$(openssl rand -base64 32) \
  --from-literal=prometheus-auth-token=$(openssl rand -base64 32) \
  --from-literal=loki-auth-token=$(openssl rand -base64 32) \
  --namespace=innovabiz-observability

# Aplicar ConfigMaps
kubectl apply -f configmaps/ --namespace=innovabiz-observability
```

### 4.3 Instalação do Stack de Monitoramento

```bash
# Instalação do Prometheus Operator
helm upgrade --install prometheus-operator prometheus-community/kube-prometheus-stack \
  --namespace innovabiz-observability \
  --values helm-values/prometheus-values.yaml

# Instalação do OpenTelemetry Collector
helm upgrade --install opentelemetry-collector open-telemetry/opentelemetry-collector \
  --namespace innovabiz-observability \
  --values helm-values/otel-collector-values.yaml

# Instalação do Jaeger
helm upgrade --install jaeger jaegertracing/jaeger \
  --namespace innovabiz-observability \
  --values helm-values/jaeger-values.yaml
```

### 4.4 Instalação do Stack de Logging

```bash
# Instalação do Elasticsearch e Kibana
helm upgrade --install elasticsearch elastic/elasticsearch \
  --namespace innovabiz-observability \
  --values helm-values/elasticsearch-values.yaml

helm upgrade --install kibana elastic/kibana \
  --namespace innovabiz-observability \
  --values helm-values/kibana-values.yaml

# Instalação do Fluentd
helm upgrade --install fluentd fluent/fluentd \
  --namespace innovabiz-observability \
  --values helm-values/fluentd-values.yaml

# Instalação do Loki
helm upgrade --install loki grafana/loki-stack \
  --namespace innovabiz-observability \
  --values helm-values/loki-values.yaml
```

### 4.5 Instalação do Observability Portal

```bash
# Criação de ConfigMap para o portal
kubectl apply -f configmaps/observability-portal-config.yaml \
  --namespace=innovabiz-observability

# Instalação dos componentes do portal
kubectl apply -f manifests/observability-portal/ \
  --namespace=innovabiz-observability
```

### 4.6 Configuração do Ingress

```bash
# Aplicar regras de ingress
kubectl apply -f ingress/observability-ingress.yaml \
  --namespace=innovabiz-observability
```

## 5. Configuração Multi-Dimensional

### 5.1 Configuração Multi-Tenant

O suporte multi-tenant é implementado em vários níveis:

- **Nível de Armazenamento**:
  - Índices separados por tenant no Elasticsearch
  - Namespaces Loki por tenant
  - Labels de tenant em todas as métricas Prometheus

- **Nível de Interface**:
  - Filtros de tenant no Observability Portal
  - Organizações separadas no Grafana
  - Espaços separados no Kibana

Implementação:

```yaml
# Exemplo para Fluentd (extract from configmap)
<filter **>
  @type record_transformer
  <record>
    tenant_id ${record["kubernetes"]["namespace_labels"]["innovabiz.com/tenant"] || "default"}
    region_id ${record["kubernetes"]["namespace_labels"]["innovabiz.com/region"] || "default"}
    environment ${record["kubernetes"]["namespace_labels"]["innovabiz.com/environment"] || "default"}
    module_id ${record["kubernetes"]["namespace_labels"]["innovabiz.com/module"] || "default"}
    component_id ${record["kubernetes"]["namespace_labels"]["innovabiz.com/component"] || "default"}
  </record>
</filter>
```

### 5.2 Configuração Multi-Regional

A infraestrutura suporta quatro regiões principais:
- Brasil (BR)
- Estados Unidos (US)
- União Europeia (EU)
- Angola (AO)

Cada região tem:
- Subsistema de armazenamento local
- Pipeline de agregação regional
- Federação de alertas
- Contexto regional na telemetria

Implementação:

```yaml
# Exemplo para OpenTelemetry Collector (extract from configmap)
processors:
  resource:
    attributes:
    - key: innovabiz.region
      value: "${INNOVABIZ_REGION}"
      action: upsert
    - key: innovabiz.environment
      value: "${INNOVABIZ_ENVIRONMENT}"
      action: upsert
```

### 5.3 Configuração Multi-Ambiente

Suporte para ambientes:
- Desenvolvimento
- Qualidade
- Homologação
- Produção
- Sandbox

Cada ambiente possui:
- Políticas de retenção específicas
- SLOs diferenciados
- Alertas apropriados ao contexto

### 5.4 Propagação de Contexto

A propagação do contexto multi-dimensional é feita através:

- **W3C Trace Context**: Para tracing distribuído
- **Campos padronizados**: Tenant ID, Region ID, Environment, Module ID, Component ID
- **Headers HTTP**: Para requisições entre serviços
- **Baggage OpenTelemetry**: Para contexto adicional

## 6. Integração com IAM Service

### 6.1 Autenticação via OIDC

```yaml
# Exemplo de configuração OIDC para Grafana (extract from configmap)
auth.generic_oauth:
  enabled: true
  name: INNOVABIZ SSO
  client_id: grafana-observability
  client_secret: ${OIDC_CLIENT_SECRET}
  auth_url: https://iam.innovabiz.com/oauth2/authorize
  token_url: https://iam.innovabiz.com/oauth2/token
  api_url: https://iam.innovabiz.com/userinfo
  scopes: openid profile email
  role_attribute_path: contains(groups[*], 'observability-admin') && 'Admin' || contains(groups[*], 'observability-editor') && 'Editor' || 'Viewer'
```

### 6.2 Autorização RBAC

O modelo RBAC implementa os seguintes papéis:

- **observability-admin**: Acesso completo à stack
- **observability-editor**: Pode criar/editar dashboards e configurar alertas
- **observability-viewer**: Pode visualizar dashboards e métricas
- **tenant-observability-admin**: Administrador limitado ao seu tenant
- **regional-observability-admin**: Administrador limitado à sua região

### 6.3 Auditoria de Acessos

- Logs de auditoria enviados ao IAM Audit Service
- Rastreamento de todas as operações administrativas
- Retenção alinhada com requisitos regulatórios
- Alerta para operações sensíveis

## 7. Validação e Testes

### 7.1 Verificação de Componentes

```bash
# Verificar status dos pods
kubectl get pods -n innovabiz-observability

# Verificar status dos services
kubectl get services -n innovabiz-observability

# Verificar status dos ingress
kubectl get ingress -n innovabiz-observability
```

### 7.2 Testes Funcionais

- **Geração de Telemetria de Teste**:
  ```bash
  ./scripts/generate-test-telemetry.sh
  ```

- **Validação de Dashboards Padrão**:
  - Acessar Grafana e verificar dashboards predefinidos
  - Acessar Kibana e verificar visualizações predefinidas
  - Acessar Portal de Observabilidade e validar visão unificada

- **Teste de Alertas**:
  ```bash
  ./scripts/test-alerting-pipeline.sh
  ```

### 7.3 Testes de Segurança

- **Verificação de TLS**:
  ```bash
  ./scripts/verify-tls-endpoints.sh
  ```

- **Verificação de RBAC**:
  ```bash
  ./scripts/test-rbac-permissions.sh
  ```

- **Verificação de Network Policies**:
  ```bash
  ./scripts/verify-network-policies.sh
  ```

## 8. Migração de Dados

### 8.1 Migração de Configurações Existentes

- Dashboards Grafana
- Alertas e regras
- Configurações de visualização
- Templates e exportações

### 8.2 Migração de Dados Históricos

Procedimento para migração de dados históricos:

1. Snapshot dos dados atuais
2. Transformação para formato compatível
3. Carga incremental por período
4. Validação de integridade e consistência

## 9. Troubleshooting Comum

### 9.1 Problemas de Ingress

- **Sintoma**: URLs do Observability Portal não respondem
- **Solução**: Verificar configuração do Ingress e certificados TLS
  ```bash
  kubectl describe ingress -n innovabiz-observability
  kubectl get certificate -n innovabiz-observability
  ```

### 9.2 Problemas de Autenticação

- **Sintoma**: Falha na autenticação com IAM
- **Solução**: Verificar configuração OIDC e logs
  ```bash
  kubectl logs deployment/observability-portal-backend -n innovabiz-observability
  ```

### 9.3 Problemas de Persistência

- **Sintoma**: Perda de dados após reinicialização
- **Solução**: Verificar PVCs e StorageClasses
  ```bash
  kubectl get pvc -n innovabiz-observability
  kubectl describe pvc -n innovabiz-observability
  ```

## 10. Referências

- [Documentação do OpenTelemetry](https://opentelemetry.io/docs/)
- [Documentação do Prometheus](https://prometheus.io/docs/introduction/overview/)
- [Documentação do Elasticsearch](https://www.elastic.co/guide/index.html)
- [Documentação do Loki](https://grafana.com/docs/loki/latest/)
- [Documentação do Grafana](https://grafana.com/docs/grafana/latest/)
- [Documentação do Kubernetes](https://kubernetes.io/docs/home/)
- [Padrões de Observabilidade CNCF](https://github.com/cncf/tag-observability)

## Informações Adicionais

Este guia está alinhado com:
- ISO/IEC 27001:2013
- PCI DSS v4.0
- LGPD/GDPR
- NIST Cybersecurity Framework
- SOC 2 Type 2

---

© 2025 INNOVABIZ. Todos os direitos reservados.