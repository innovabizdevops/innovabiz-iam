# Runbook de Resposta a Alertas - IAM Audit Service

**Autor**: INNOVABIZ DevOps  
**Versão**: 1.0.0  
**Data de Criação**: 2025-07-31  
**Última Atualização**: 2025-07-31  
**Classificação**: Restrito  

## Índice

1. [Introdução](#introdução)
2. [Classificação de Alertas](#classificação-de-alertas)
3. [Procedimentos de Resposta](#procedimentos-de-resposta)
   - [Alertas de Disponibilidade](#alertas-de-disponibilidade)
   - [Alertas de Latência](#alertas-de-latência)
   - [Alertas de Taxa de Erro](#alertas-de-taxa-de-erro)
   - [Alertas de Eventos de Auditoria](#alertas-de-eventos-de-auditoria)
   - [Alertas de Conformidade](#alertas-de-conformidade)
   - [Alertas de Política de Retenção](#alertas-de-política-de-retenção)
   - [Alertas de Saúde de Banco de Dados](#alertas-de-saúde-de-banco-de-dados)
   - [Alertas de Incidentes de Segurança](#alertas-de-incidentes-de-segurança)
4. [Escalação](#escalação)
5. [Ferramentas e Dashboards](#ferramentas-e-dashboards)
6. [Procedimentos de Mitigação](#procedimentos-de-mitigação)

## Introdução

Este runbook fornece procedimentos detalhados para responder a alertas relacionados ao IAM Audit Service e seu framework de observabilidade. Os alertas são configurados no Prometheus com integração ao Grafana e sistemas de notificação. Cada procedimento foi projetado para garantir resposta rápida e eficaz a incidentes, minimizando o impacto nos negócios e mantendo a conformidade com regulamentações e políticas de segurança.

## Classificação de Alertas

| Nível | Descrição | Tempo de Resposta | Escalação |
|-------|-----------|-------------------|-----------|
| P1 - Crítico | Impacto severo na produção. Sistema inoperante ou perda de dados de auditoria críticos | Imediata (15 min) | SRE + Gerente + CISO |
| P2 - Alto | Funcionalidade crítica afetada. Impacto significativo na operação | 1 hora | SRE + Gerente |
| P3 - Médio | Funcionalidade parcial afetada. Impacto moderado | 4 horas | SRE |
| P4 - Baixo | Problema menor sem impacto imediato na operação | 24 horas | Desenvolvimento |
| P5 - Informacional | Não é um problema, apenas notificação | N/A | N/A |

## Procedimentos de Resposta

### Alertas de Disponibilidade

#### `IAMAuditServiceDown`

**Nível**: P1 - Crítico

**Descrição**: O serviço de auditoria IAM está indisponível.

**Procedimento**:

1. Verificar status do serviço nos endpoints de health:
   ```bash
   curl -i https://{ambiente}-iam-audit.innovabiz.com/health
   curl -i https://{ambiente}-iam-audit.innovabiz.com/ready
   curl -i https://{ambiente}-iam-audit.innovabiz.com/live
   ```

2. Verificar logs do serviço:
   ```bash
   kubectl logs -n iam-system -l app=iam-audit-service --tail=200
   ```

3. Verificar eventos do Kubernetes:
   ```bash
   kubectl get events -n iam-system --sort-by='.lastTimestamp'
   ```

4. Verificar métricas de recursos:
   ```bash
   kubectl top pods -n iam-system -l app=iam-audit-service
   ```

5. Reiniciar o serviço se necessário:
   ```bash
   kubectl rollout restart deployment/iam-audit-service -n iam-system
   ```

6. Verificar dependências (banco de dados, cache, etc.) usando o endpoint de diagnóstico:
   ```bash
   curl -i https://{ambiente}-iam-audit.innovabiz.com/diagnostic?include_deps=true
   ```

7. Se o problema persistir, escalar conforme matriz de escalação.

### Alertas de Latência

#### `IAMAuditServiceHighLatency`

**Nível**: P2 - Alto

**Descrição**: O serviço está respondendo com latência acima do limiar configurado.

**Procedimento**:

1. Verificar dashboard de latência no Grafana para identificar endpoints afetados.

2. Analisar traces no Jaeger para identificar gargalos:
   ```bash
   # Obter ID do trace com maior latência
   curl -s "http://jaeger:16686/api/traces?service=iam-audit-service&limit=20" | jq '.data[] | select(.spans[].duration > 500000) | .traceID'
   ```

3. Verificar carga no banco de dados:
   ```bash
   kubectl exec -it -n iam-system $(kubectl get pods -n iam-system -l app=postgres-iam-audit -o name | head -n 1) -- psql -U postgres -c "SELECT * FROM pg_stat_activity WHERE state = 'active';"
   ```

4. Verificar uso de recursos no cluster:
   ```bash
   kubectl describe nodes | grep -A 10 "Allocated resources"
   ```

5. Verificar se há muitas requisições simultâneas:
   ```bash
   kubectl exec -it -n iam-system $(kubectl get pods -n iam-system -l app=iam-audit-service -o name | head -n 1) -- netstat -ant | grep ESTABLISHED | wc -l
   ```

6. Se a latência for causada por alto volume, considerar escalamento horizontal:
   ```bash
   kubectl scale deployment/iam-audit-service -n iam-system --replicas=<número_maior>
   ```

### Alertas de Taxa de Erro

#### `IAMAuditServiceErrorRateHigh`

**Nível**: P2 - Alto

**Descrição**: Taxa de erros HTTP 5xx acima do limiar configurado.

**Procedimento**:

1. Identificar os endpoints com erros usando o dashboard Grafana.

2. Analisar logs de erro:
   ```bash
   kubectl logs -n iam-system -l app=iam-audit-service --tail=500 | grep -i error | grep -v "level=debug"
   ```

3. Verificar traces com erros no Jaeger:
   ```bash
   curl -s "http://jaeger:16686/api/traces?service=iam-audit-service&tags=%7B%22error%22%3A%22true%22%7D&limit=20" | jq
   ```

4. Verificar memória e CPU do serviço:
   ```bash
   kubectl top pods -n iam-system -l app=iam-audit-service
   ```

5. Se os erros estiverem relacionados ao banco de dados, verificar conexões:
   ```bash
   kubectl exec -it -n iam-system $(kubectl get pods -n iam-system -l app=postgres-iam-audit -o name | head -n 1) -- psql -U postgres -c "SELECT count(*) FROM pg_stat_activity;"
   ```

6. Se houver erros de timeout, verificar configurações de conexão e considerar aumentar timeouts temporariamente.

7. Considerar rollback para versão estável se os erros começaram após deploy recente:
   ```bash
   kubectl rollout undo deployment/iam-audit-service -n iam-system
   ```

### Alertas de Eventos de Auditoria

#### `IAMAuditEventProcessingFailureHigh`

**Nível**: P1 - Crítico

**Descrição**: Alto número de falhas no processamento de eventos de auditoria.

**Procedimento**:

1. Verificar tipos de eventos com falhas usando o dashboard Grafana.

2. Analisar logs específicos de processamento de eventos:
   ```bash
   kubectl logs -n iam-system -l app=iam-audit-service --tail=500 | grep -i "audit.event.processing"
   ```

3. Verificar filas de eventos não processados (se aplicável):
   ```bash
   kubectl exec -it -n iam-system $(kubectl get pods -n iam-system -l app=kafka -o name | head -n 1) -- kafka-consumer-groups.sh --bootstrap-server localhost:9092 --group iam-audit-processor --describe
   ```

4. Verificar espaço em disco para bancos de dados de auditoria:
   ```bash
   kubectl exec -it -n iam-system $(kubectl get pods -n iam-system -l app=postgres-iam-audit -o name | head -n 1) -- df -h
   ```

5. Verificar se há problema de validação de esquema:
   ```bash
   kubectl logs -n iam-system -l app=iam-audit-service --tail=200 | grep -i "schema"
   ```

6. Se necessário, ativar modo de contingência para armazenar eventos em buffer local:
   ```bash
   kubectl set env deployment/iam-audit-service -n iam-system AUDIT_PROCESSING_CONTINGENCY_MODE=true
   ```

### Alertas de Conformidade

#### `IAMAuditComplianceCheckFailuresHigh`

**Nível**: P2 - Alto

**Descrição**: Alto número de falhas em verificações de conformidade.

**Procedimento**:

1. Identificar quais estruturas de conformidade estão falhando (GDPR, PCI-DSS, SOX, etc.) no dashboard.

2. Verificar logs de verificações de conformidade:
   ```bash
   kubectl logs -n iam-system -l app=iam-audit-service --tail=500 | grep -i "compliance.check"
   ```

3. Verificar se as falhas são para um tenant específico:
   ```bash
   curl -s "https://{ambiente}-iam-audit.innovabiz.com/metrics" | grep "compliance_checks_failed_total" | sort
   ```

4. Verificar configurações de conformidade para o tenant afetado:
   ```bash
   kubectl exec -it -n iam-system $(kubectl get pods -n iam-system -l app=iam-audit-service -o name | head -n 1) -- cat /etc/config/compliance/tenant-{tenant-id}.yaml
   ```

5. Notificar oficial de conformidade se houver violação grave.

### Alertas de Política de Retenção

#### `IAMAuditRetentionPolicyFailure`

**Nível**: P2 - Alto

**Descrição**: Falhas na aplicação da política de retenção de dados de auditoria.

**Procedimento**:

1. Verificar logs de execução da política de retenção:
   ```bash
   kubectl logs -n iam-system -l app=iam-audit-service --tail=500 | grep -i "retention.policy"
   ```

2. Verificar espaço em disco:
   ```bash
   kubectl exec -it -n iam-system $(kubectl get pods -n iam-system -l app=postgres-iam-audit -o name | head -n 1) -- df -h
   ```

3. Verificar locks no banco de dados:
   ```bash
   kubectl exec -it -n iam-system $(kubectl get pods -n iam-system -l app=postgres-iam-audit -o name | head -n 1) -- psql -U postgres -c "SELECT relation::regclass::text, mode, locktype, granted FROM pg_locks JOIN pg_database ON pg_locks.database = pg_database.oid WHERE pg_database.datname = 'iam_audit';"
   ```

4. Verificar tamanho das tabelas de auditoria:
   ```bash
   kubectl exec -it -n iam-system $(kubectl get pods -n iam-system -l app=postgres-iam-audit -o name | head -n 1) -- psql -U postgres -c "SELECT pg_size_pretty(pg_total_relation_size('audit_events')) as size;"
   ```

5. Se necessário, executar manualmente a política de retenção para um tenant específico:
   ```bash
   kubectl exec -it -n iam-system $(kubectl get pods -n iam-system -l app=iam-audit-service -o name | head -n 1) -- curl -X POST "http://localhost:8000/internal/maintenance/retention?tenant={tenant-id}"
   ```

### Alertas de Saúde de Banco de Dados

#### `IAMAuditDatabaseConnectionFailure`

**Nível**: P1 - Crítico

**Descrição**: Falhas de conexão com o banco de dados de auditoria.

**Procedimento**:

1. Verificar status do banco de dados:
   ```bash
   kubectl get pods -n iam-system -l app=postgres-iam-audit
   ```

2. Verificar logs do banco de dados:
   ```bash
   kubectl logs -n iam-system -l app=postgres-iam-audit --tail=200
   ```

3. Verificar se o serviço pode conectar ao banco de dados:
   ```bash
   kubectl exec -it -n iam-system $(kubectl get pods -n iam-system -l app=iam-audit-service -o name | head -n 1) -- pg_isready -h postgres-iam-audit
   ```

4. Verificar conexões atuais ao banco de dados:
   ```bash
   kubectl exec -it -n iam-system $(kubectl get pods -n iam-system -l app=postgres-iam-audit -o name | head -n 1) -- psql -U postgres -c "SELECT count(*) FROM pg_stat_activity;"
   ```

5. Se necessário, reiniciar o banco de dados (último recurso):
   ```bash
   kubectl rollout restart statefulset/postgres-iam-audit -n iam-system
   ```

### Alertas de Incidentes de Segurança

#### `IAMAuditSecurityIncidentDetected`

**Nível**: P1 - Crítico

**Descrição**: Detecção de possível incidente de segurança relacionado a eventos de auditoria.

**Procedimento**:

1. Isolar imediatamente evidências preservando logs:
   ```bash
   kubectl cp -n iam-system $(kubectl get pods -n iam-system -l app=iam-audit-service -o name | head -n 1):/var/log/audit/ ./incident-$(date +%Y%m%d-%H%M%S)/
   ```

2. Identificar origem do alerta nos logs:
   ```bash
   kubectl logs -n iam-system -l app=iam-audit-service --tail=500 | grep -i "security.incident"
   ```

3. Verificar tentativas de acesso não autorizado:
   ```bash
   kubectl exec -it -n iam-system $(kubectl get pods -n iam-system -l app=iam-audit-service -o name | head -n 1) -- cat /var/log/audit/auth.log | grep -i "unauthorized"
   ```

4. Coletar metadados de traces suspeitos:
   ```bash
   curl -s "http://jaeger:16686/api/traces?service=iam-audit-service&tags=%7B%22security.incident%22%3A%22true%22%7D&limit=20" | jq > ./incident-$(date +%Y%m%d-%H%M%S)/traces.json
   ```

5. Notificar imediatamente a equipe de segurança e o CISO.

6. Iniciar procedimentos de resposta a incidentes conforme política de segurança.

## Escalação

### Matriz de Escalação

| Nível | Tempo | Contato Primário | Contato Secundário | Gerência |
|-------|-------|------------------|-------------------|----------|
| L1 | 15 min | DevOps On-call | SRE On-call | N/A |
| L2 | 1 hora | SRE Lead | Dev Lead | Gerente de TI |
| L3 | 4 horas | Arquiteto | Dev Manager | Diretor de TI |
| L4 | 8 horas | CTO | CISO | CEO |

### Procedimento de Escalação

1. Se não for possível resolver o incidente no nível atual dentro do tempo alocado, escale para o próximo nível.
2. Ao escalar, forneça:
   - ID do incidente
   - Descrição do problema
   - Ações já realizadas
   - Logs e evidências relevantes
   - Impacto atual nos negócios

## Ferramentas e Dashboards

- **Grafana**: https://{ambiente}-grafana.innovabiz.com
   - Dashboard Principal: IAM Audit Service Overview
   - Dashboard de Conformidade: IAM Audit Compliance
   - Dashboard de Segurança: IAM Audit Security

- **Prometheus**: https://{ambiente}-prometheus.innovabiz.com
   - Alertas: /alerts
   - Regras: /rules

- **Jaeger**: https://{ambiente}-jaeger.innovabiz.com
   - Serviço: iam-audit-service

- **Elasticsearch**: https://{ambiente}-kibana.innovabiz.com
   - Índice: innovabiz-audit-logs-*

## Procedimentos de Mitigação

### Failover para Sistema Secundário

Em caso de falha completa do sistema primário:

```bash
# 1. Verificar status do sistema secundário
kubectl get pods -n iam-system-dr -l app=iam-audit-service

# 2. Atualizar DNS para apontar para sistema DR
kubectl exec -it -n iam-system $(kubectl get pods -n iam-system -l app=dns-manager -o name | head -n 1) -- update-dns-record iam-audit.innovabiz.com <ip-do-sistema-dr>

# 3. Notificar equipes de suporte sobre a mudança
```

### Ativação de Circuit Breaker

Se sistemas dependentes estiverem causando problemas:

```bash
# Ativar circuit breaker para dependências específicas
kubectl set env deployment/iam-audit-service -n iam-system CIRCUIT_BREAKER_ENABLED=true CIRCUIT_BREAKER_SERVICES=identity-service,policy-service
```

### Modo de Degradação Controlada

Para manter funcionalidades essenciais durante problemas graves:

```bash
# Ativar modo de degradação controlada
kubectl set env deployment/iam-audit-service -n iam-system DEGRADATION_MODE=true ESSENTIAL_FEATURES_ONLY=true
```

---

**Nota**: Este runbook deve ser revisado e atualizado a cada 3 meses ou após qualquer incidente significativo.