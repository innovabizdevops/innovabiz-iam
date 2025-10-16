# IAM Services - Runbook Operacional

## Visão Geral

Este runbook fornece procedimentos operacionais detalhados para o gerenciamento, monitoramento e troubleshooting dos serviços de Identity and Access Management (IAM) da plataforma INNOVABIZ. O documento abrange desde operações de rotina até procedimentos de recuperação de desastres.

### Escopo do Documento
- **Serviços Cobertos**: Autenticação, Autorização, Gestão de Tokens JWT, Sessões
- **Ambientes**: Desenvolvimento, Staging, Produção
- **Arquitetura**: Multi-tenant, Multi-região, Multi-contexto
- **Compliance**: ISO 27001, PCI DSS 4.0, GDPR/LGPD, NIST CSF

---

## Arquitetura e Componentes

### Componentes Principais
```yaml
IAM Services:
  Authentication Service:
    - JWT Token Generation
    - Multi-factor Authentication
    - Password Policies
    - Account Lockout
  
  Authorization Service:
    - Role-Based Access Control (RBAC)
    - Policy Engine
    - Resource Permissions
    - Privilege Escalation Detection
  
  Session Management:
    - Session Creation/Termination
    - Session Timeout
    - Concurrent Session Control
    - Session Hijacking Detection
  
  Audit Service:
    - Access Logs
    - Security Events
    - Compliance Reports
    - Forensic Analysis
```

### Dependências Críticas
- **PostgreSQL**: Armazenamento de usuários e políticas
- **Redis**: Cache de sessões e tokens
- **Kafka**: Eventos de auditoria
- **Vault**: Gestão de chaves criptográficas
- **LDAP/AD**: Integração com diretórios corporativos

---

## Monitoramento e Alertas

### Métricas Críticas

#### Disponibilidade
```bash
# Verificar status do serviço
curl -f http://iam-service:8080/health

# Prometheus query
up{job="iam-service"} == 0
```

#### Performance
```bash
# Latência de autenticação
histogram_quantile(0.95, rate(iam_authentication_duration_seconds_bucket[5m])) > 1

# Taxa de erro
rate(iam_authentication_failures_total[5m]) / rate(iam_authentication_attempts_total[5m]) > 0.01
```

#### Segurança
```bash
# Tentativas de força bruta
rate(iam_authentication_failures_total[5m]) > 10

# Escalação de privilégios
increase(iam_privilege_escalation_attempts_total[5m]) > 0
```

### Alertas Configurados
- **IAMServiceDown**: Serviço indisponível
- **IAMHighAuthenticationFailures**: Taxa alta de falhas
- **IAMPrivilegeEscalationAttempt**: Tentativa de escalação
- **IAMHighLatency**: Latência elevada
- **IAMAuditLogFailure**: Falha no sistema de auditoria

---

## Procedimentos de Troubleshooting

### 1. Serviço IAM Indisponível

#### Sintomas
- Dashboard mostra status vermelho
- Aplicações não conseguem autenticar usuários
- Erro 503 Service Unavailable

#### Diagnóstico
```bash
# Verificar status do pod/container
kubectl get pods -l app=iam-service -n innovabiz

# Verificar logs do serviço
kubectl logs -f deployment/iam-service -n innovabiz --tail=100

# Verificar conectividade com dependências
kubectl exec -it deployment/iam-service -n innovabiz -- \
  curl -f http://postgresql:5432/health
```

#### Resolução
```bash
# Restart do serviço
kubectl rollout restart deployment/iam-service -n innovabiz

# Verificar se voltou ao normal
kubectl get pods -l app=iam-service -n innovabiz
watch kubectl get pods -l app=iam-service -n innovabiz

# Verificar métricas
curl http://iam-service:8080/metrics | grep iam_up
```

#### Escalação
- **Tempo de resolução**: 15 minutos
- **Escalação**: Se não resolvido, escalar para Nível 2

### 2. Alta Taxa de Falhas de Autenticação

#### Sintomas
- Alerta: IAMHighAuthenticationFailures
- Usuários reportam problemas de login
- Dashboard mostra picos de falhas

#### Diagnóstico
```bash
# Analisar padrões de falha
kubectl logs deployment/iam-service -n innovabiz | \
  grep "authentication_failure" | tail -50

# Verificar origem das tentativas
kubectl logs deployment/iam-service -n innovabiz | \
  grep "authentication_failure" | \
  awk '{print $8}' | sort | uniq -c | sort -nr

# Verificar se é ataque coordenado
kubectl logs deployment/iam-service -n innovabiz | \
  grep "authentication_failure" | \
  grep -E "$(date '+%Y-%m-%d %H:%M')" | wc -l
```

#### Resolução
```bash
# Se for ataque de força bruta, implementar rate limiting
kubectl patch configmap iam-config -n innovabiz --patch '
data:
  rate_limit_enabled: "true"
  rate_limit_requests_per_minute: "10"
'

# Restart para aplicar configuração
kubectl rollout restart deployment/iam-service -n innovabiz

# Verificar se as falhas diminuíram
watch 'curl -s http://prometheus:9090/api/v1/query?query=rate(iam_authentication_failures_total[5m]) | jq .data.result[0].value[1]'
```

#### Escalação
- **Tempo de resolução**: 30 minutos
- **Escalação**: Se suspeita de ataque sofisticado, escalar para Segurança

### 3. Tentativa de Escalação de Privilégios

#### Sintomas
- Alerta crítico: IAMPrivilegeEscalationAttempt
- Logs de auditoria mostram tentativas suspeitas
- Usuário tenta acessar recursos não autorizados

#### Diagnóstico
```bash
# Identificar usuário e recurso
kubectl logs deployment/iam-service -n innovabiz | \
  grep "privilege_escalation" | tail -10

# Verificar histórico do usuário
kubectl logs deployment/iam-service -n innovabiz | \
  grep "user_id:SUSPICIOUS_USER_ID" | tail -20

# Verificar sessões ativas do usuário
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://iam-service:8080/api/v1/sessions?user_id=SUSPICIOUS_USER_ID
```

#### Resolução Imediata
```bash
# Suspender usuário imediatamente
curl -X POST -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"action": "suspend", "reason": "privilege_escalation_attempt"}' \
  http://iam-service:8080/api/v1/users/SUSPICIOUS_USER_ID/actions

# Invalidar todas as sessões do usuário
curl -X DELETE -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://iam-service:8080/api/v1/sessions?user_id=SUSPICIOUS_USER_ID

# Notificar equipe de segurança
curl -X POST -H "Content-Type: application/json" \
  -d '{
    "alert": "privilege_escalation_detected",
    "user_id": "SUSPICIOUS_USER_ID",
    "timestamp": "'$(date -Iseconds)'",
    "severity": "critical"
  }' \
  http://security-webhook:8080/alerts
```

#### Escalação
- **Tempo de resolução**: Imediato (suspensão)
- **Escalação**: Sempre escalar para Segurança

### 4. Alta Latência em Operações IAM

#### Sintomas
- Alerta: IAMHighLatency
- Usuários reportam lentidão no login
- Dashboard mostra P95 > 1 segundo

#### Diagnóstico
```bash
# Verificar performance do banco de dados
kubectl exec -it postgresql-0 -n innovabiz -- \
  psql -U iam_user -d iam_db -c "
    SELECT query, mean_time, calls 
    FROM pg_stat_statements 
    WHERE query LIKE '%iam%' 
    ORDER BY mean_time DESC LIMIT 10;"

# Verificar cache Redis
kubectl exec -it redis-0 -n innovabiz -- \
  redis-cli info stats | grep keyspace

# Verificar recursos do pod
kubectl top pod -l app=iam-service -n innovabiz
```

#### Resolução
```bash
# Se problema for no banco, otimizar queries
kubectl exec -it postgresql-0 -n innovabiz -- \
  psql -U iam_user -d iam_db -c "REINDEX DATABASE iam_db;"

# Se problema for cache, limpar cache inválido
kubectl exec -it redis-0 -n innovabiz -- \
  redis-cli FLUSHDB

# Se problema for recursos, escalar horizontalmente
kubectl scale deployment iam-service --replicas=5 -n innovabiz

# Verificar se latência melhorou
watch 'curl -s http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95,rate(iam_operation_duration_seconds_bucket[5m])) | jq .data.result[0].value[1]'
```

#### Escalação
- **Tempo de resolução**: 45 minutos
- **Escalação**: Se não melhorar, escalar para DBA

---

## Operações de Rotina

### Rotação de Chaves JWT

#### Frequência: Mensal (ou conforme política)

```bash
# Gerar nova chave privada
openssl genpkey -algorithm RSA -out jwt_private_new.pem -pkcs8 -pass pass:$JWT_KEY_PASSWORD

# Extrair chave pública
openssl rsa -pubout -in jwt_private_new.pem -out jwt_public_new.pem -passin pass:$JWT_KEY_PASSWORD

# Atualizar secret no Kubernetes
kubectl create secret generic jwt-keys-new -n innovabiz \
  --from-file=private=jwt_private_new.pem \
  --from-file=public=jwt_public_new.pem

# Atualizar configuração do serviço
kubectl patch deployment iam-service -n innovabiz --patch '
spec:
  template:
    spec:
      containers:
      - name: iam-service
        env:
        - name: JWT_KEY_VERSION
          value: "new"
'

# Aguardar rollout
kubectl rollout status deployment/iam-service -n innovabiz

# Verificar se tokens estão sendo gerados com nova chave
kubectl logs deployment/iam-service -n innovabiz | grep "jwt_key_rotation"

# Após período de transição, remover chave antiga
kubectl delete secret jwt-keys-old -n innovabiz
```

### Limpeza de Sessões Expiradas

#### Frequência: Diária

```bash
# Conectar ao banco de dados
kubectl exec -it postgresql-0 -n innovabiz -- \
  psql -U iam_user -d iam_db

# Executar limpeza
DELETE FROM user_sessions 
WHERE expires_at < NOW() - INTERVAL '1 day';

# Verificar quantidade removida
SELECT COUNT(*) FROM user_sessions WHERE expires_at < NOW();

# Otimizar tabela
VACUUM ANALYZE user_sessions;
```

### Backup de Configurações

#### Frequência: Semanal

```bash
# Backup de ConfigMaps
kubectl get configmap iam-config -n innovabiz -o yaml > iam-config-backup-$(date +%Y%m%d).yaml

# Backup de Secrets (sem valores sensíveis)
kubectl get secret jwt-keys -n innovabiz -o yaml | \
  sed 's/data:/# data:/' > jwt-keys-backup-$(date +%Y%m%d).yaml

# Backup de políticas RBAC
kubectl get rolebinding,clusterrolebinding -l app=iam-service -o yaml > \
  iam-rbac-backup-$(date +%Y%m%d).yaml

# Upload para storage seguro
aws s3 cp iam-*-backup-$(date +%Y%m%d).yaml s3://innovabiz-backups/iam/
```

---

## Recuperação de Desastres

### Cenário 1: Perda Completa do Serviço IAM

#### Tempo de Recuperação: 30 minutos

```bash
# 1. Restaurar deployment
kubectl apply -f iam-service-deployment.yaml -n innovabiz

# 2. Restaurar configurações
kubectl apply -f iam-config-backup-latest.yaml -n innovabiz

# 3. Restaurar secrets
kubectl apply -f jwt-keys-backup-latest.yaml -n innovabiz

# 4. Verificar dependências
kubectl get pods -l app=postgresql -n innovabiz
kubectl get pods -l app=redis -n innovabiz

# 5. Aguardar inicialização
kubectl wait --for=condition=ready pod -l app=iam-service -n innovabiz --timeout=300s

# 6. Verificar saúde
curl -f http://iam-service:8080/health

# 7. Executar testes de fumaça
./scripts/iam-smoke-tests.sh
```

### Cenário 2: Corrupção de Dados de Usuários

#### Tempo de Recuperação: 2 horas

```bash
# 1. Parar serviço IAM
kubectl scale deployment iam-service --replicas=0 -n innovabiz

# 2. Restaurar backup do banco
kubectl exec -it postgresql-0 -n innovabiz -- \
  pg_restore -U iam_user -d iam_db /backups/iam_db_backup_latest.sql

# 3. Verificar integridade dos dados
kubectl exec -it postgresql-0 -n innovabiz -- \
  psql -U iam_user -d iam_db -c "
    SELECT COUNT(*) FROM users WHERE status = 'active';
    SELECT COUNT(*) FROM roles;
    SELECT COUNT(*) FROM permissions;
  "

# 4. Reiniciar serviço
kubectl scale deployment iam-service --replicas=3 -n innovabiz

# 5. Executar testes completos
./scripts/iam-integration-tests.sh
```

---

## Configurações Recomendadas

### Configuração de Produção

```yaml
# iam-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: iam-config
  namespace: innovabiz
data:
  # Autenticação
  jwt_expiry_minutes: "60"
  refresh_token_expiry_days: "30"
  max_login_attempts: "5"
  account_lockout_minutes: "30"
  
  # Sessões
  session_timeout_minutes: "120"
  max_concurrent_sessions: "3"
  session_cleanup_interval_minutes: "60"
  
  # Segurança
  password_min_length: "12"
  password_require_special_chars: "true"
  mfa_required_for_admin: "true"
  audit_log_retention_days: "365"
  
  # Performance
  connection_pool_size: "50"
  cache_ttl_seconds: "300"
  rate_limit_requests_per_minute: "100"
```

### Limites de Recursos

```yaml
# iam-deployment.yaml
resources:
  requests:
    memory: "512Mi"
    cpu: "250m"
  limits:
    memory: "2Gi"
    cpu: "1000m"

# HPA
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: iam-service-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: iam-service
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

---

## Segurança e Compliance

### Controles de Segurança

#### Criptografia
- **Em Trânsito**: TLS 1.3 para todas as comunicações
- **Em Repouso**: AES-256 para dados sensíveis
- **Chaves**: Rotação automática a cada 30 dias

#### Auditoria
```bash
# Verificar logs de auditoria
kubectl logs deployment/iam-service -n innovabiz | \
  grep "audit" | jq '.timestamp, .user_id, .action, .resource'

# Exportar logs para SIEM
kubectl logs deployment/iam-service -n innovabiz --since=1h | \
  grep "audit" | \
  curl -X POST -H "Content-Type: application/json" \
  -d @- http://siem-collector:8080/logs
```

#### Compliance
- **GDPR**: Direito ao esquecimento implementado
- **PCI DSS**: Dados de cartão não armazenados no IAM
- **SOX**: Controles de acesso auditáveis
- **ISO 27001**: Gestão de riscos de segurança

### Políticas de Acesso

```yaml
# Exemplo de política RBAC
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: iam-operator
  namespace: innovabiz
rules:
- apiGroups: [""]
  resources: ["pods", "configmaps", "secrets"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["get", "list", "watch", "patch"]
```

---

## Ciclo de Vida e Manutenção

### Atualizações de Versão

#### Processo Blue-Green
```bash
# 1. Deploy nova versão (green)
kubectl apply -f iam-service-v2-deployment.yaml -n innovabiz

# 2. Aguardar inicialização
kubectl wait --for=condition=ready pod -l app=iam-service,version=v2 -n innovabiz

# 3. Executar testes
./scripts/iam-health-check.sh v2

# 4. Alternar tráfego
kubectl patch service iam-service -n innovabiz --patch '
spec:
  selector:
    version: v2
'

# 5. Monitorar métricas
watch kubectl get pods -l app=iam-service -n innovabiz

# 6. Remover versão antiga (após confirmação)
kubectl delete deployment iam-service-v1 -n innovabiz
```

### Monitoramento de Capacidade

```bash
# Análise de tendências mensais
curl -G http://prometheus:9090/api/v1/query_range \
  --data-urlencode 'query=iam_active_sessions_total' \
  --data-urlencode 'start='$(date -d '30 days ago' +%s) \
  --data-urlencode 'end='$(date +%s) \
  --data-urlencode 'step=3600' | \
  jq '.data.result[0].values' > iam_capacity_trend.json

# Previsão de crescimento
python3 scripts/capacity_forecast.py iam_capacity_trend.json
```

---

## Contatos de Escalação

### Níveis de Escalação

#### Nível 1 - Suporte Técnico (24/7)
- **Tempo de Resposta**: 15 minutos
- **Escopo**: Problemas básicos, reinicializações
- **Contato**: support-l1@innovabiz.com
- **Telefone**: +55 11 9999-0001

#### Nível 2 - Especialista IAM
- **Tempo de Resposta**: 30 minutos
- **Escopo**: Problemas de configuração, performance
- **Contato**: iam-specialist@innovabiz.com
- **Telefone**: +55 11 9999-0002

#### Nível 3 - Arquiteto de Segurança
- **Tempo de Resposta**: 1 hora
- **Escopo**: Incidentes de segurança, compliance
- **Contato**: security-architect@innovabiz.com
- **Telefone**: +55 11 9999-0003

#### Nível 4 - Arquiteto Principal
- **Tempo de Resposta**: 2 horas
- **Escopo**: Decisões arquiteturais críticas
- **Contato**: Eduardo Jeremias (innovabizdevops@gmail.com)
- **Telefone**: +55 11 9999-0004

### Matriz de Escalação

| Tipo de Incidente | Severidade | Nível Inicial | Escalação Automática |
|-------------------|------------|---------------|---------------------|
| Serviço Indisponível | Crítica | Nível 1 | 15 min → Nível 2 |
| Falhas de Autenticação | Alta | Nível 1 | 30 min → Nível 2 |
| Escalação de Privilégios | Crítica | Nível 3 | Imediato |
| Performance Degradada | Média | Nível 1 | 45 min → Nível 2 |
| Falha de Auditoria | Crítica | Nível 3 | Imediato |

---

## Anexos

### Scripts Úteis

#### Health Check Completo
```bash
#!/bin/bash
# iam-health-check.sh

echo "=== IAM Service Health Check ==="

# Verificar pods
echo "1. Checking pods..."
kubectl get pods -l app=iam-service -n innovabiz

# Verificar endpoints
echo "2. Checking endpoints..."
curl -f http://iam-service:8080/health || echo "Health check failed"
curl -f http://iam-service:8080/ready || echo "Readiness check failed"

# Verificar métricas
echo "3. Checking metrics..."
curl -s http://iam-service:8080/metrics | grep iam_up | tail -1

# Verificar dependências
echo "4. Checking dependencies..."
kubectl get pods -l app=postgresql -n innovabiz
kubectl get pods -l app=redis -n innovabiz

echo "=== Health Check Complete ==="
```

#### Teste de Autenticação
```bash
#!/bin/bash
# iam-auth-test.sh

echo "=== IAM Authentication Test ==="

# Teste de login
response=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"test@innovabiz.com","password":"TestPassword123!"}' \
  http://iam-service:8080/api/v1/auth/login)

token=$(echo $response | jq -r '.access_token')

if [ "$token" != "null" ]; then
  echo "✅ Authentication successful"
  
  # Teste de validação de token
  validation=$(curl -s -H "Authorization: Bearer $token" \
    http://iam-service:8080/api/v1/auth/validate)
  
  if [ "$(echo $validation | jq -r '.valid')" = "true" ]; then
    echo "✅ Token validation successful"
  else
    echo "❌ Token validation failed"
  fi
else
  echo "❌ Authentication failed"
fi

echo "=== Authentication Test Complete ==="
```

### Queries Prometheus Úteis

```promql
# Taxa de sucesso de autenticação por tenant
sum(rate(iam_authentication_success_total[5m])) by (tenant_id) / 
sum(rate(iam_authentication_attempts_total[5m])) by (tenant_id)

# Top 10 usuários com mais falhas de autenticação
topk(10, sum(rate(iam_authentication_failures_total[1h])) by (user_id))

# Distribuição de latência por operação
histogram_quantile(0.95, 
  sum(rate(iam_operation_duration_seconds_bucket[5m])) by (le, operation)
)

# Sessões ativas por região
sum(iam_active_sessions_total) by (region_id)

# Taxa de crescimento de usuários
increase(iam_users_total[24h])
```

---

*Runbook criado em: 2025-01-31*  
*Versão: 1.0*  
*Próxima revisão: 2025-02-07*  
*Responsável: Eduardo Jeremias (innovabizdevops@gmail.com)*