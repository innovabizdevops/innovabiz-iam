# WebAuthn Deployment Guide

**Documento:** Guia de Deploy WebAuthn/FIDO2  
**Versão:** 1.0.0  
**Data:** 31/07/2025  
**Autor:** Equipe DevOps INNOVABIZ  
**Classificação:** Confidencial - Operacional  

## Índice

1. [Pré-requisitos](#1-pré-requisitos)
2. [Preparação do Ambiente](#2-preparação-do-ambiente)
3. [Deploy com Helm](#3-deploy-com-helm)
4. [Configuração de Secrets](#4-configuração-de-secrets)
5. [Verificação do Deploy](#5-verificação-do-deploy)
6. [Monitoramento](#6-monitoramento)
7. [Troubleshooting](#7-troubleshooting)
8. [Rollback](#8-rollback)

## 1. Pré-requisitos

### 1.1 Infraestrutura

| Componente | Versão Mínima | Recursos Mínimos |
|------------|---------------|------------------|
| **Kubernetes** | 1.24+ | 3 nodes, 8 CPU, 16GB RAM |
| **Helm** | 3.12+ | - |
| **Ingress Controller** | nginx 1.8+ | - |
| **Cert Manager** | 1.12+ | - |
| **Prometheus** | 2.40+ | - |

### 1.2 Dependências

```bash
# Instalar dependências do Helm
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo add cert-manager https://charts.jetstack.io
helm repo update
```

### 1.3 Namespaces

```bash
# Criar namespaces
kubectl create namespace innovabiz-staging
kubectl create namespace innovabiz-production
kubectl create namespace monitoring
```

## 2. Preparação do Ambiente

### 2.1 Configuração de Secrets

#### Staging Environment

```bash
# Database secrets
kubectl create secret generic webauthn-staging-db \
  --from-literal=username=webauthn_user \
  --from-literal=password=staging_db_password \
  --namespace=innovabiz-staging

# Redis secrets
kubectl create secret generic webauthn-staging-redis \
  --from-literal=password=staging_redis_password \
  --namespace=innovabiz-staging

# JWT secrets
kubectl create secret generic webauthn-staging-jwt \
  --from-literal=private-key="$(cat jwt-private-key.pem)" \
  --from-literal=public-key="$(cat jwt-public-key.pem)" \
  --namespace=innovabiz-staging
```

#### Production Environment

```bash
# Database secrets
kubectl create secret generic webauthn-production-db \
  --from-literal=username=webauthn_user \
  --from-literal=password=production_db_password \
  --namespace=innovabiz-production

# Redis secrets
kubectl create secret generic webauthn-production-redis \
  --from-literal=password=production_redis_password \
  --namespace=innovabiz-production

# JWT secrets
kubectl create secret generic webauthn-production-jwt \
  --from-literal=private-key="$(cat jwt-private-key.pem)" \
  --from-literal=public-key="$(cat jwt-public-key.pem)" \
  --namespace=innovabiz-production
```

### 2.2 Configuração de TLS

```bash
# Certificados Let's Encrypt
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    email: devops@innovabiz.com
    privateKeySecretRef:
      name: letsencrypt-staging
    solvers:
    - http01:
        ingress:
          class: nginx
---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: devops@innovabiz.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
```

## 3. Deploy com Helm

### 3.1 Deploy Staging

```bash
# Deploy para staging
helm upgrade --install webauthn-staging ./helm/webauthn \
  --namespace innovabiz-staging \
  --create-namespace \
  --values helm/values-staging.yaml \
  --set image.tag=staging-latest \
  --set secrets.DATABASE_PASSWORD="$(kubectl get secret webauthn-staging-db -o jsonpath='{.data.password}' | base64 -d)" \
  --set secrets.REDIS_PASSWORD="$(kubectl get secret webauthn-staging-redis -o jsonpath='{.data.password}' | base64 -d)" \
  --set secrets.JWT_PRIVATE_KEY="$(kubectl get secret webauthn-staging-jwt -o jsonpath='{.data.private-key}')" \
  --set secrets.JWT_PUBLIC_KEY="$(kubectl get secret webauthn-staging-jwt -o jsonpath='{.data.public-key}')" \
  --wait \
  --timeout=10m
```

### 3.2 Deploy Production

```bash
# Deploy para production
helm upgrade --install webauthn-production ./helm/webauthn \
  --namespace innovabiz-production \
  --create-namespace \
  --values helm/values-production.yaml \
  --set image.tag=1.0.0 \
  --set secrets.DATABASE_PASSWORD="$(kubectl get secret webauthn-production-db -o jsonpath='{.data.password}' | base64 -d)" \
  --set secrets.REDIS_PASSWORD="$(kubectl get secret webauthn-production-redis -o jsonpath='{.data.password}' | base64 -d)" \
  --set secrets.JWT_PRIVATE_KEY="$(kubectl get secret webauthn-production-jwt -o jsonpath='{.data.private-key}')" \
  --set secrets.JWT_PUBLIC_KEY="$(kubectl get secret webauthn-production-jwt -o jsonpath='{.data.public-key}')" \
  --wait \
  --timeout=15m
```

### 3.3 Blue-Green Deployment (Production)

```bash
# Deploy green environment
helm upgrade --install webauthn-green ./helm/webauthn \
  --namespace innovabiz-production \
  --values helm/values-production.yaml \
  --set image.tag=1.0.1 \
  --set service.name=webauthn-green \
  --set deployment.labels.version=green \
  --wait \
  --timeout=15m

# Smoke tests no ambiente green
kubectl run smoke-test --rm -i --tty \
  --image=curlimages/curl \
  --restart=Never \
  --namespace=innovabiz-production \
  -- curl -f https://green-api.innovabiz.com/health

# Switch traffic para green
kubectl patch service webauthn-service \
  --namespace innovabiz-production \
  --patch '{"spec":{"selector":{"version":"green"}}}'

# Cleanup blue environment
helm uninstall webauthn-blue --namespace innovabiz-production || true
```

## 4. Configuração de Secrets

### 4.1 External Secrets Operator (Recomendado)

```yaml
# external-secret.yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: webauthn-secrets
  namespace: innovabiz-production
spec:
  refreshInterval: 15s
  secretStoreRef:
    name: vault-backend
    kind: SecretStore
  target:
    name: webauthn-secrets
    creationPolicy: Owner
  data:
  - secretKey: DATABASE_PASSWORD
    remoteRef:
      key: secret/webauthn/production
      property: database_password
  - secretKey: REDIS_PASSWORD
    remoteRef:
      key: secret/webauthn/production
      property: redis_password
  - secretKey: JWT_PRIVATE_KEY
    remoteRef:
      key: secret/webauthn/production
      property: jwt_private_key
```

### 4.2 Sealed Secrets (Alternativa)

```bash
# Instalar Sealed Secrets
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.18.0/controller.yaml

# Criar sealed secret
echo -n 'production_password' | kubectl create secret generic webauthn-db-secret \
  --dry-run=client --from-file=password=/dev/stdin -o yaml | \
  kubeseal -o yaml > webauthn-db-sealed-secret.yaml

kubectl apply -f webauthn-db-sealed-secret.yaml
```

## 5. Verificação do Deploy

### 5.1 Health Checks

```bash
# Verificar pods
kubectl get pods -n innovabiz-production -l app.kubernetes.io/name=webauthn

# Verificar logs
kubectl logs -n innovabiz-production -l app.kubernetes.io/name=webauthn --tail=100

# Verificar health endpoint
kubectl port-forward -n innovabiz-production svc/webauthn 8080:80
curl http://localhost:8080/health
```

### 5.2 Smoke Tests

```bash
# Teste de conectividade
curl -f https://api.innovabiz.com/api/v1/webauthn/health

# Teste de registro (mock)
curl -X POST https://api.innovabiz.com/api/v1/webauthn/registration/options \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TEST_TOKEN" \
  -H "X-Tenant-ID: test-tenant" \
  -d '{
    "userId": "test-user",
    "username": "test@innovabiz.com",
    "displayName": "Test User"
  }'
```

### 5.3 Load Testing

```bash
# K6 load test
k6 run --vus 10 --duration 30s tests/performance/load-test.js
```

## 6. Monitoramento

### 6.1 Prometheus Metrics

```bash
# Verificar métricas
kubectl port-forward -n innovabiz-production svc/webauthn 9090:80
curl http://localhost:9090/metrics | grep webauthn
```

### 6.2 Grafana Dashboards

```bash
# Importar dashboard
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: webauthn-dashboard
  namespace: monitoring
  labels:
    grafana_dashboard: "1"
data:
  webauthn.json: |
    {
      "dashboard": {
        "title": "WebAuthn Service",
        "panels": [
          {
            "title": "Request Rate",
            "type": "graph",
            "targets": [
              {
                "expr": "rate(webauthn_requests_total[5m])"
              }
            ]
          }
        ]
      }
    }
EOF
```

### 6.3 Alertas

```bash
# Verificar alertas ativos
kubectl get prometheusrules -n innovabiz-production
```

## 7. Troubleshooting

### 7.1 Problemas Comuns

#### Pod não inicia

```bash
# Verificar eventos
kubectl describe pod -n innovabiz-production -l app.kubernetes.io/name=webauthn

# Verificar recursos
kubectl top pods -n innovabiz-production

# Verificar secrets
kubectl get secrets -n innovabiz-production
```

#### Erro de conexão com banco

```bash
# Testar conectividade
kubectl run db-test --rm -i --tty \
  --image=postgres:15 \
  --restart=Never \
  --namespace=innovabiz-production \
  -- psql postgresql://user:pass@postgres:5432/webauthn
```

#### Certificado SSL inválido

```bash
# Verificar certificados
kubectl get certificates -n innovabiz-production
kubectl describe certificate webauthn-production-tls -n innovabiz-production

# Forçar renovação
kubectl delete certificate webauthn-production-tls -n innovabiz-production
```

### 7.2 Logs de Debug

```bash
# Aumentar log level
kubectl patch deployment webauthn-production -n innovabiz-production \
  --patch '{"spec":{"template":{"spec":{"containers":[{"name":"webauthn","env":[{"name":"LOG_LEVEL","value":"debug"}]}]}}}}'

# Verificar logs detalhados
kubectl logs -n innovabiz-production -l app.kubernetes.io/name=webauthn -f
```

## 8. Rollback

### 8.1 Rollback Helm

```bash
# Listar releases
helm history webauthn-production -n innovabiz-production

# Rollback para versão anterior
helm rollback webauthn-production 1 -n innovabiz-production
```

### 8.2 Rollback Blue-Green

```bash
# Switch de volta para blue
kubectl patch service webauthn-service \
  --namespace innovabiz-production \
  --patch '{"spec":{"selector":{"version":"blue"}}}'

# Verificar saúde
curl -f https://api.innovabiz.com/health
```

### 8.3 Emergency Rollback

```bash
# Rollback de emergência
kubectl set image deployment/webauthn-production \
  webauthn=ghcr.io/innovabiz/iam-webauthn:1.0.0 \
  -n innovabiz-production

# Verificar rollout
kubectl rollout status deployment/webauthn-production -n innovabiz-production
```

## 9. Manutenção

### 9.1 Backup

```bash
# Backup do banco
kubectl exec -n innovabiz-production webauthn-postgresql-0 -- \
  pg_dump -U webauthn_user webauthn_production > backup-$(date +%Y%m%d).sql

# Backup de secrets
kubectl get secrets -n innovabiz-production -o yaml > secrets-backup-$(date +%Y%m%d).yaml
```

### 9.2 Atualizações

```bash
# Atualizar dependências
helm dependency update ./helm/webauthn

# Deploy com nova versão
helm upgrade webauthn-production ./helm/webauthn \
  --namespace innovabiz-production \
  --values helm/values-production.yaml \
  --set image.tag=1.0.1
```

---

**Desenvolvido pela equipe INNOVABIZ**  
**© 2025 INNOVABIZ. Todos os direitos reservados.**