# Procedimentos de Troubleshooting de Performance IAM

## Introdução

Este documento fornece procedimentos detalhados para diagnóstico e resolução de problemas relacionados à performance no módulo IAM da plataforma INNOVABIZ. É destinado às equipes de operações, administradores de sistemas e profissionais de suporte técnico responsáveis por manter o desempenho e a escalabilidade do sistema IAM em ambientes multi-tenant e com alta carga.

## Matriz de Problemas Comuns

| Sintoma | Possível Causa | Gravidade | Impacto | Tempo Médio de Resolução |
|---------|----------------|-----------|---------|--------------------------|
| Autenticação lenta | Problemas de banco de dados ou cache | Alta | Experiência de login degradada | 30-60 minutos |
| Latência alta em decisões de autorização | Políticas complexas ou ineficientes | Alta | Atrasos em todas as operações protegidas | 30-90 minutos |
| Timeout em operações de gestão IAM | Sobrecarga de recursos no painel administrativo | Média | Administradores não conseguem gerenciar usuários/acessos | 30-60 minutos |
| Escalabilidade limitada em picos de carga | Configurações inadequadas de recursos | Alta | Falhas durante períodos de alta utilização | 60-120 minutos |
| Degradação gradual de performance | Memory leaks ou crescimento não controlado de dados | Média | Deterioração progressiva do serviço | 60-180 minutos |
| Problemas de performance específicos por tenant | Configurações multi-tenant inadequadas | Alta | Impacto isolado em tenants específicos | 45-90 minutos |

## Procedimentos de Troubleshooting

### 1. Autenticação Lenta

#### 1.1 Sintomas
- Aumento no tempo de resposta durante operações de login
- Reclamações de usuários sobre lentidão no processo de autenticação
- Métricas mostrando tempo de processamento elevado na API de autenticação
- Timeouts durante o processo de login

#### 1.2 Verificações Iniciais
1. **Verificar métricas de performance:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=avg_over_time(auth_request_duration_seconds[15m])'
   ```

2. **Analisar logs para operações lentas:**
   ```bash
   kubectl logs -n iam-namespace <auth-service-pod-name> --tail=200 | grep -i "slow\|timeout\|duration\|latency"
   ```

3. **Verificar utilização de recursos:**
   ```bash
   kubectl top pods -n iam-namespace | grep auth-service
   ```

4. **Verificar performance da conexão com banco de dados:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-service-pod-name> -- pg_isready -h <db-host> -p <db-port>
   kubectl exec -it -n iam-namespace <auth-service-pod-name> -- redis-cli -h <redis-host> ping
   ```

#### 1.3 Diagnóstico Avançado
1. **Analisar queries lentas no banco de dados:**
   ```sql
   SELECT query, calls, total_time, mean_time, max_time 
   FROM pg_stat_statements 
   WHERE query LIKE '%users%' OR query LIKE '%credentials%' OR query LIKE '%authentication%' 
   ORDER BY mean_time DESC LIMIT 20;
   ```

2. **Verificar estatísticas de conexão com o banco:**
   ```sql
   SELECT state, count(*) FROM pg_stat_activity GROUP BY state;
   SELECT datname, numbackends, xact_commit, xact_rollback, blks_read, blks_hit, tup_returned, tup_fetched, tup_inserted, tup_updated, tup_deleted
   FROM pg_stat_database WHERE datname = 'iam_database';
   ```

3. **Verificar eficiência do cache:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli INFO stats | grep -E 'keyspace_hits|keyspace_misses|used_memory'
   ```

4. **Analisar perfil de autenticação:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-service-pod-name> -- /app/scripts/profile-auth-flow.sh
   ```

5. **Verificar contadores de autenticação por tipo:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=sum(auth_attempts_total) by (method, status)'
   ```

#### 1.4 Ações Corretivas

**Nível 1 (Mitigação Rápida):**
1. **Reiniciar serviço de autenticação:**
   ```bash
   kubectl rollout restart deployment -n iam-namespace auth-service
   ```

2. **Limpar cache de sessão:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli FLUSHDB
   ```

3. **Escalar horizontalmente o serviço:**
   ```bash
   kubectl scale deployment -n iam-namespace auth-service --replicas=5
   ```

4. **Aumentar capacidade de conexões:**
   ```bash
   kubectl set env deployment/auth-service -n iam-namespace DB_POOL_MAX_SIZE=50 DB_POOL_MIN_SIZE=10
   ```

**Nível 2 (Resolução):**
1. **Otimizar queries de banco de dados:**
   ```sql
   -- Adicionar índices para campos frequentemente pesquisados
   CREATE INDEX IF NOT EXISTS idx_users_username_tenant ON iam_schema.users(username, tenant_id);
   CREATE INDEX IF NOT EXISTS idx_auth_logs_timestamp ON iam_schema.authentication_logs(timestamp, user_id);
   
   -- Analisar e otimizar tabelas
   VACUUM ANALYZE iam_schema.users;
   VACUUM ANALYZE iam_schema.credentials;
   ```

2. **Implementar estratégia de cache mais eficiente:**
   ```bash
   kubectl apply -f optimized-cache-config.yaml -n iam-namespace
   ```

3. **Ajustar configurações de timeouts e conexões:**
   ```bash
   kubectl set env deployment/auth-service -n iam-namespace \
     HTTP_CLIENT_TIMEOUT=10 \
     DB_STATEMENT_TIMEOUT=5000 \
     CONNECTION_POOL_MAX_SIZE=100
   ```

4. **Implementar fragmentação (sharding) para tenants grandes:**
   ```bash
   kubectl apply -f tenant-sharding-config.yaml -n iam-namespace
   ```

#### 1.5 Verificação de Resolução
1. **Monitorar tempos de resposta por 30 minutos**
2. **Verificar logs por erros de timeout ou latência**
3. **Executar testes de login com diferentes tipos de usuários**
4. **Verificar métricas de utilização de recursos**

#### 1.6 Ações Pós-Incidente
1. **Implementar monitoramento proativo de tempos de autenticação**
2. **Estabelecer alertas para tempos de resposta degradados**
3. **Documentar queries otimizadas e configurações de cache**
4. **Planejar revisão periódica de índices e esquemas**

### 2. Latência Alta em Decisões de Autorização

#### 2.1 Sintomas
- Tempo elevado para verificações de permissões
- Operações de API com alta latência nos componentes de autorização
- Logs mostrando gargalos na avaliação de políticas
- Timeouts em operações que requerem múltiplas verificações de permissão

#### 2.2 Verificações Iniciais
1. **Verificar métricas de avaliação de políticas:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=avg_over_time(policy_evaluation_duration_seconds[15m])'
   ```

2. **Analisar logs de decisões lentas:**
   ```bash
   kubectl logs -n iam-namespace <rbac-service-pod-name> --tail=200 | grep -i "slow\|timeout\|duration\|decision"
   ```

3. **Verificar contagem de políticas ativas:**
   ```sql
   SELECT tenant_id, COUNT(*) FROM iam_schema.policies GROUP BY tenant_id ORDER BY count DESC;
   ```

4. **Analisar utilização de recursos:**
   ```bash
   kubectl top pods -n iam-namespace | grep rbac
   ```

#### 2.3 Diagnóstico Avançado
1. **Analisar políticas complexas:**
   ```sql
   SELECT policy_id, tenant_id, effect, LENGTH(resource_pattern) + LENGTH(action_pattern) + LENGTH(COALESCE(conditions::text, '')) as complexity
   FROM iam_schema.policies
   ORDER BY complexity DESC LIMIT 20;
   ```

2. **Verificar caminhos de avaliação longos:**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/analyze-policy-paths.sh --threshold=20
   ```

3. **Analisar distribuição de decisões de autorização:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=sum(authorization_decisions_total) by (effect, tenant)'
   ```

4. **Avaliar eficiência do cache de decisões:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli --scan --pattern "decision:*" | wc -l
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli INFO | grep -i hit_rate
   ```

#### 2.4 Ações Corretivas

**Nível 1 (Mitigação Rápida):**
1. **Aumentar recursos do serviço de RBAC:**
   ```bash
   kubectl patch deployment -n iam-namespace rbac-service -p '{"spec":{"template":{"spec":{"containers":[{"name":"rbac-service","resources":{"limits":{"cpu":"4","memory":"8Gi"},"requests":{"cpu":"2","memory":"4Gi"}}}]}}}}'
   ```

2. **Ajustar configurações de cache de decisões:**
   ```bash
   kubectl set env deployment/rbac-service -n iam-namespace \
     DECISION_CACHE_TTL=600 \
     DECISION_CACHE_SIZE=50000
   ```

3. **Escalar horizontalmente o serviço de RBAC:**
   ```bash
   kubectl scale deployment -n iam-namespace rbac-service --replicas=8
   ```

**Nível 2 (Resolução):**
1. **Otimizar índices de políticas:**
   ```sql
   CREATE INDEX IF NOT EXISTS idx_policies_combined ON iam_schema.policies 
   USING GIN (to_tsvector('english', resource_pattern || ' ' || action_pattern));
   ANALYZE iam_schema.policies;
   ```

2. **Implementar otimizações de avaliação:**
   ```bash
   kubectl apply -f policy-evaluation-optimizations.yaml -n iam-namespace
   ```

3. **Reestruturar políticas complexas:**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/optimize-policies.sh --tenant-id=<tenant-id>
   ```

4. **Implementar avaliação paralela de políticas:**
   ```bash
   kubectl set env deployment/rbac-service -n iam-namespace \
     PARALLEL_POLICY_EVALUATION=true \
     PARALLEL_THREADS=8
   ```

#### 2.5 Verificação de Resolução
1. **Monitorar tempos de decisão por 30 minutos**
2. **Testar autorização em cenários de alta complexidade**
3. **Verificar impacto na utilização de CPU e memória**
4. **Analisar distribuição de tempos de resposta (p50, p95, p99)**

#### 2.6 Ações Pós-Incidente
1. **Estabelecer revisões periódicas de políticas**
2. **Implementar monitoramento de políticas complexas**
3. **Desenvolver guias de otimização para administradores**
4. **Planejar refatoração de políticas problemáticas**

### 3. Problemas de Escalabilidade em Picos de Carga

#### 3.1 Sintomas
- Falhas durante períodos de uso intenso (início do dia, eventos especiais)
- Erros de timeout ou sobrecarga durante picos de autenticação
- Degradação de performance quando múltiplos tenants acessam simultaneamente
- Alertas de utilização elevada de recursos

#### 3.2 Verificações Iniciais
1. **Verificar histórico de utilização de recursos:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=max_over_time(container_cpu_usage_seconds_total{namespace="iam-namespace"}[1d])'
   ```

2. **Analisar métricas de requisições:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=sum(rate(http_requests_total{namespace="iam-namespace"}[5m])) by (service)'
   ```

3. **Verificar configurações de escalonamento automático:**
   ```bash
   kubectl get hpa -n iam-namespace
   ```

4. **Analisar estatísticas de conexões:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-service-pod-name> -- netstat -ant | grep ESTABLISHED | wc -l
   ```

#### 3.3 Diagnóstico Avançado
1. **Analisar padrões de uso por horário:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query_range?query=sum(rate(http_requests_total{namespace="iam-namespace"}[5m]))&start=2025-05-08T00:00:00Z&end=2025-05-09T00:00:00Z&step=1h'
   ```

2. **Verificar distribuição de carga por tenant:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=sum(rate(http_requests_total{namespace="iam-namespace"}[5m])) by (tenant_id)'
   ```

3. **Verificar gargalos de conexão com dependências:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-service-pod-name> -- /app/scripts/connection-statistics.sh
   ```

4. **Analisar saturação de recursos:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_irate{namespace="iam-namespace"}) / sum(kube_pod_container_resource_limits_cpu_cores{namespace="iam-namespace"})'
   ```

#### 3.4 Ações Corretivas

**Nível 1 (Mitigação Rápida):**
1. **Aumentar limites de recursos:**
   ```bash
   kubectl set resources deployment -n iam-namespace auth-service --limits=cpu=8,memory=16Gi --requests=cpu=4,memory=8Gi
   kubectl set resources deployment -n iam-namespace rbac-service --limits=cpu=8,memory=16Gi --requests=cpu=4,memory=8Gi
   ```

2. **Escalar serviços horizontalmente:**
   ```bash
   kubectl scale deployment -n iam-namespace auth-service --replicas=10
   kubectl scale deployment -n iam-namespace rbac-service --replicas=10
   kubectl scale deployment -n iam-namespace token-service --replicas=6
   ```

3. **Otimizar configurações de conexão:**
   ```bash
   kubectl set env deployment -n iam-namespace auth-service \
     CONNECTION_POOL_MAX_SIZE=200 \
     CONNECTION_POOL_IDLE_TIMEOUT=600
   ```

**Nível 2 (Resolução):**
1. **Implementar escalonamento automático:**
   ```yaml
   apiVersion: autoscaling/v2
   kind: HorizontalPodAutoscaler
   metadata:
     name: auth-service-hpa
     namespace: iam-namespace
   spec:
     scaleTargetRef:
       apiVersion: apps/v1
       kind: Deployment
       name: auth-service
     minReplicas: 5
     maxReplicas: 20
     metrics:
     - type: Resource
       resource:
         name: cpu
         target:
           type: Utilization
           averageUtilization: 70
     - type: Resource
       resource:
         name: memory
         target:
           type: Utilization
           averageUtilization: 75
     behavior:
       scaleUp:
         stabilizationWindowSeconds: 60
       scaleDown:
         stabilizationWindowSeconds: 300
   ```

2. **Implementar limitação de taxa por tenant:**
   ```bash
   kubectl apply -f tenant-rate-limiting.yaml -n iam-namespace
   ```

3. **Configurar estratégia de cache distribuído:**
   ```bash
   kubectl apply -f distributed-cache-config.yaml -n iam-namespace
   ```

4. **Otimizar configurações para bancos de dados em horários de pico:**
   ```bash
   kubectl apply -f peak-hours-database-config.yaml -n iam-namespace
   ```

#### 3.5 Verificação de Resolução
1. **Executar testes de carga simulando picos**
2. **Monitorar comportamento de escalonamento automático**
3. **Verificar distribuição de requisições entre pods**
4. **Monitorar tempos de resposta durante escalonamento**

#### 3.6 Ações Pós-Incidente
1. **Estabelecer padrões de provisionamento baseados em uso**
2. **Implementar previsão de capacidade baseada em tendências**
3. **Documentar configurações otimizadas para períodos de pico**
4. **Desenvolver procedimentos para balanceamento proativo de carga**

### 4. Problemas de Performance Multi-Tenant

#### 4.1 Sintomas
- Performance inconsistente entre diferentes tenants
- Um tenant específico experimenta degradação enquanto outros funcionam normalmente
- Alertas de uso excessivo de recursos por um tenant
- Operações de IAM lentas apenas para tenant específico

#### 4.2 Verificações Iniciais
1. **Verificar métricas por tenant:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=sum(rate(http_requests_total{namespace="iam-namespace"}[5m])) by (tenant_id)'
   ```

2. **Analisar volume de dados por tenant:**
   ```sql
   SELECT tenant_id, 
          COUNT(DISTINCT user_id) as users,
          COUNT(DISTINCT role_id) as roles,
          COUNT(DISTINCT policy_id) as policies
   FROM iam_schema.users u
   JOIN iam_schema.tenants t USING (tenant_id)
   LEFT JOIN iam_schema.user_roles ur USING (user_id)
   LEFT JOIN iam_schema.roles r USING (tenant_id)
   LEFT JOIN iam_schema.policies p USING (tenant_id)
   GROUP BY tenant_id
   ORDER BY users DESC;
   ```

3. **Verificar recursos alocados por tenant:**
   ```bash
   kubectl get resourcequota -n iam-namespace
   ```

4. **Analisar estatísticas de cache por tenant:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli --scan --pattern "tenant:*" | sort | uniq -c
   ```

#### 4.3 Diagnóstico Avançado
1. **Analisar isolamento de queries por tenant:**
   ```sql
   SELECT tenant_id, query, calls, total_time, mean_time
   FROM pg_stat_statements pss
   JOIN iam_schema.tenant_sessions ts ON pss.userid = ts.session_id
   ORDER BY mean_time DESC LIMIT 20;
   ```

2. **Verificar problemas de isolamento de recursos:**
   ```bash
   kubectl top pods -n iam-namespace --sort-by=cpu
   kubectl top pods -n iam-namespace --sort-by=memory
   ```

3. **Analisar complexity por tenant:**
   ```sql
   SELECT tenant_id, 
          COUNT(*) as policy_count,
          AVG(LENGTH(resource_pattern)) as avg_resource_length,
          AVG(LENGTH(action_pattern)) as avg_action_length,
          AVG(LENGTH(conditions::text)) as avg_condition_length
   FROM iam_schema.policies
   GROUP BY tenant_id
   ORDER BY policy_count DESC;
   ```

4. **Verificar padrões de acesso por tenant:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=topk(10, sum(rate(http_requests_total{namespace="iam-namespace"}[1h])) by (tenant_id, path))'
   ```

#### 4.4 Ações Corretivas

**Nível 1 (Mitigação Rápida):**
1. **Aplicar limitação de recursos por tenant:**
   ```bash
   kubectl apply -f tenant-resource-limits.yaml -n iam-namespace
   ```

2. **Reiniciar serviços para tenant específico:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-service-pod-name> -- /app/scripts/restart-tenant-services.sh <problematic-tenant-id>
   ```

3. **Limpar caches para tenant específico:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli --scan --pattern "tenant:<problematic-tenant-id>:*" | xargs redis-cli DEL
   ```

**Nível 2 (Resolução):**
1. **Implementar particionamento por tenant:**
   ```sql
   -- Particionar tabelas grandes por tenant_id
   ALTER TABLE iam_schema.users PARTITION BY LIST (tenant_id);
   ALTER TABLE iam_schema.policies PARTITION BY LIST (tenant_id);
   ```

2. **Configurar pools de conexão dedicados por tenant:**
   ```bash
   kubectl apply -f tenant-dedicated-pools.yaml -n iam-namespace
   ```

3. **Implementar isolamento de cache por tenant:**
   ```bash
   kubectl apply -f tenant-isolated-cache.yaml -n iam-namespace
   ```

4. **Otimizar políticas para tenant específico:**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/optimize-tenant-policies.sh <problematic-tenant-id>
   ```

#### 4.5 Verificação de Resolução
1. **Monitorar performance do tenant problemático**
2. **Verificar impacto em outros tenants**
3. **Analisar métricas de utilização de recursos**
4. **Testar operações críticas específicas do tenant**

#### 4.6 Ações Pós-Incidente
1. **Implementar alertas específicos por tenant**
2. **Estabelecer limites de uso por tamanho de tenant**
3. **Desenvolver planos de isolamento para tenants de grande escala**
4. **Documentar práticas recomendadas de configuração multi-tenant**

## Recursos Adicionais

### Ferramentas de Diagnóstico

1. **Scripts de Diagnóstico:**
   - `/opt/innovabiz/iam/scripts/performance-analyzer.sh`
   - `/opt/innovabiz/iam/scripts/tenant-resource-usage.sh`
   - `/opt/innovabiz/iam/scripts/database-performance-check.sh`

2. **Dashboards de Monitoramento:**
   - Grafana IAM Performance: `https://grafana.innovabiz.com/d/iam-performance`
   - Grafana Multi-Tenant: `https://grafana.innovabiz.com/d/iam-tenant-metrics`
   - Prometheus: `https://prometheus.innovabiz.com/graph`

3. **Consultas Úteis para o Banco de Dados:**
   ```sql
   -- Identificar queries lentas nos serviços IAM
   SELECT substring(query, 1, 100) as query_excerpt, 
          calls, total_time, mean_time, max_time,
          stddev_time, rows
   FROM pg_stat_statements
   WHERE query ILIKE '%iam_schema%'
   ORDER BY mean_time DESC
   LIMIT 20;
   
   -- Verificar índices não utilizados
   SELECT s.schemaname,
          s.relname as tablename,
          s.indexrelname as indexname,
          s.idx_scan as index_scans
   FROM pg_stat_user_indexes s
   JOIN pg_index i ON s.indexrelid = i.indexrelid
   WHERE s.schemaname = 'iam_schema' AND s.idx_scan = 0
   ORDER BY s.relname, s.indexrelname;
   
   -- Analisar bloqueios no banco de dados
   SELECT blocked_locks.pid as blocked_pid,
          blocking_locks.pid as blocking_pid,
          blocked_activity.usename as blocked_user,
          blocking_activity.usename as blocking_user,
          blocked_activity.query as blocked_statement,
          blocking_activity.query as blocking_statement
   FROM pg_catalog.pg_locks blocked_locks
   JOIN pg_catalog.pg_locks blocking_locks 
        ON blocking_locks.locktype = blocked_locks.locktype
        AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
        AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
        AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
        AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
        AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
        AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
        AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
        AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
        AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
        AND blocking_locks.pid != blocked_locks.pid
   JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
   JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
   WHERE NOT blocked_locks.granted;
   ```

### Referências

- [Requisitos de Infraestrutura IAM](../04-Infraestrutura/Requisitos_Infraestrutura_IAM.md)
- [Arquitetura Técnica IAM](../02-Arquitetura/Arquitetura_Tecnica_IAM.md)
- [Arquitetura Multi-Tenant](../02-Arquitetura/Arquitetura_Multi_Tenant.md)
- [Guia Operacional IAM](../08-Operacoes/Guia_Operacional_IAM.md)
- [Modelo de Dados IAM](../03-Desenvolvimento/Modelo_Dados_IAM.md)

### Contatos para Escalação

| Nível | Equipe | Contato | Acionamento |
|-------|--------|---------|------------|
| 1 | Suporte IAM | iam-support@innovabiz.com | Problemas iniciais |
| 2 | Operações IAM | iam-ops@innovabiz.com | Após 30 min sem resolução L1 |
| 3 | DevOps IAM | iam-devops@innovabiz.com | Problemas complexos de infraestrutura |
| 4 | DBA | database-admin@innovabiz.com | Problemas críticos de banco de dados |
| 5 | Arquitetura | architecture@innovabiz.com | Problemas estruturais de performance |
