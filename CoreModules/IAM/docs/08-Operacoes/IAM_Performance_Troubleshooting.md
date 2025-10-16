# IAM Performance Troubleshooting Procedures

## Introduction

This document provides detailed procedures for diagnosing and resolving performance-related issues in the IAM module of the INNOVABIZ platform. It is intended for operations teams, system administrators, and technical support professionals responsible for maintaining the performance and scalability of the IAM system in multi-tenant environments and under high load conditions.

## Common Issues Matrix

| Symptom | Possible Cause | Severity | Impact | Average Resolution Time |
|---------|----------------|-----------|---------|--------------------------|
| Slow authentication | Database or cache issues | High | Degraded login experience | 30-60 minutes |
| High latency in authorization decisions | Complex or inefficient policies | High | Delays in all protected operations | 30-90 minutes |
| Timeout in IAM management operations | Resource overload in admin dashboard | Medium | Administrators cannot manage users/access | 30-60 minutes |
| Limited scalability during peak loads | Inadequate resource configurations | High | Failures during high usage periods | 60-120 minutes |
| Gradual performance degradation | Memory leaks or uncontrolled data growth | Medium | Progressive service deterioration | 60-180 minutes |
| Tenant-specific performance issues | Inadequate multi-tenant configurations | High | Isolated impact on specific tenants | 45-90 minutes |

## Troubleshooting Procedures

### 1. Slow Authentication

#### 1.1 Symptoms
- Increased response time during login operations
- User complaints about login process slowness
- Metrics showing elevated processing time in authentication API
- Timeouts during login process

#### 1.2 Initial Checks
1. **Check performance metrics:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=avg_over_time(auth_request_duration_seconds[15m])'
   ```

2. **Analyze logs for slow operations:**
   ```bash
   kubectl logs -n iam-namespace <auth-service-pod-name> --tail=200 | grep -i "slow\|timeout\|duration\|latency"
   ```

3. **Check resource utilization:**
   ```bash
   kubectl top pods -n iam-namespace | grep auth-service
   ```

4. **Check database connection performance:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-service-pod-name> -- pg_isready -h <db-host> -p <db-port>
   kubectl exec -it -n iam-namespace <auth-service-pod-name> -- redis-cli -h <redis-host> ping
   ```

#### 1.3 Advanced Diagnosis
1. **Analyze slow database queries:**
   ```sql
   SELECT query, calls, total_time, mean_time, max_time 
   FROM pg_stat_statements 
   WHERE query LIKE '%users%' OR query LIKE '%credentials%' OR query LIKE '%authentication%' 
   ORDER BY mean_time DESC LIMIT 20;
   ```

2. **Check database connection statistics:**
   ```sql
   SELECT state, count(*) FROM pg_stat_activity GROUP BY state;
   SELECT datname, numbackends, xact_commit, xact_rollback, blks_read, blks_hit, tup_returned, tup_fetched, tup_inserted, tup_updated, tup_deleted
   FROM pg_stat_database WHERE datname = 'iam_database';
   ```

3. **Check cache efficiency:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli INFO stats | grep -E 'keyspace_hits|keyspace_misses|used_memory'
   ```

4. **Analyze authentication profile:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-service-pod-name> -- /app/scripts/profile-auth-flow.sh
   ```

5. **Check authentication counters by type:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=sum(auth_attempts_total) by (method, status)'
   ```

#### 1.4 Corrective Actions

**Level 1 (Quick Mitigation):**
1. **Restart authentication service:**
   ```bash
   kubectl rollout restart deployment -n iam-namespace auth-service
   ```

2. **Clear session cache:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli FLUSHDB
   ```

3. **Scale service horizontally:**
   ```bash
   kubectl scale deployment -n iam-namespace auth-service --replicas=5
   ```

4. **Increase connection capacity:**
   ```bash
   kubectl set env deployment/auth-service -n iam-namespace DB_POOL_MAX_SIZE=50 DB_POOL_MIN_SIZE=10
   ```

**Level 2 (Resolution):**
1. **Optimize database queries:**
   ```sql
   -- Add indexes for frequently searched fields
   CREATE INDEX IF NOT EXISTS idx_users_username_tenant ON iam_schema.users(username, tenant_id);
   CREATE INDEX IF NOT EXISTS idx_auth_logs_timestamp ON iam_schema.authentication_logs(timestamp, user_id);
   
   -- Analyze and optimize tables
   VACUUM ANALYZE iam_schema.users;
   VACUUM ANALYZE iam_schema.credentials;
   ```

2. **Implement more efficient caching strategy:**
   ```bash
   kubectl apply -f optimized-cache-config.yaml -n iam-namespace
   ```

3. **Adjust timeout and connection configurations:**
   ```bash
   kubectl set env deployment/auth-service -n iam-namespace \
     HTTP_CLIENT_TIMEOUT=10 \
     DB_STATEMENT_TIMEOUT=5000 \
     CONNECTION_POOL_MAX_SIZE=100
   ```

4. **Implement sharding for large tenants:**
   ```bash
   kubectl apply -f tenant-sharding-config.yaml -n iam-namespace
   ```

#### 1.5 Resolution Verification
1. **Monitor response times for 30 minutes**
2. **Check logs for timeout or latency errors**
3. **Run login tests with different user types**
4. **Verify resource utilization metrics**

#### 1.6 Post-Incident Actions
1. **Implement proactive monitoring of authentication times**
2. **Establish alerts for degraded response times**
3. **Document optimized queries and cache configurations**
4. **Plan periodic review of indexes and schemas**

### 2. High Latency in Authorization Decisions

#### 2.1 Symptoms
- Elevated time for permission checks
- API operations with high latency in authorization components
- Logs showing bottlenecks in policy evaluation
- Timeouts in operations requiring multiple permission checks

#### 2.2 Initial Checks
1. **Check policy evaluation metrics:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=avg_over_time(policy_evaluation_duration_seconds[15m])'
   ```

2. **Analyze logs for slow decisions:**
   ```bash
   kubectl logs -n iam-namespace <rbac-service-pod-name> --tail=200 | grep -i "slow\|timeout\|duration\|decision"
   ```

3. **Check active policy count:**
   ```sql
   SELECT tenant_id, COUNT(*) FROM iam_schema.policies GROUP BY tenant_id ORDER BY count DESC;
   ```

4. **Analyze resource utilization:**
   ```bash
   kubectl top pods -n iam-namespace | grep rbac
   ```

#### 2.3 Advanced Diagnosis
1. **Analyze complex policies:**
   ```sql
   SELECT policy_id, tenant_id, effect, LENGTH(resource_pattern) + LENGTH(action_pattern) + LENGTH(COALESCE(conditions::text, '')) as complexity
   FROM iam_schema.policies
   ORDER BY complexity DESC LIMIT 20;
   ```

2. **Check long evaluation paths:**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/analyze-policy-paths.sh --threshold=20
   ```

3. **Analyze authorization decision distribution:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=sum(authorization_decisions_total) by (effect, tenant)'
   ```

4. **Evaluate decision cache efficiency:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli --scan --pattern "decision:*" | wc -l
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli INFO | grep -i hit_rate
   ```

#### 2.4 Corrective Actions

**Level 1 (Quick Mitigation):**
1. **Increase RBAC service resources:**
   ```bash
   kubectl patch deployment -n iam-namespace rbac-service -p '{"spec":{"template":{"spec":{"containers":[{"name":"rbac-service","resources":{"limits":{"cpu":"4","memory":"8Gi"},"requests":{"cpu":"2","memory":"4Gi"}}}]}}}}'
   ```

2. **Adjust decision cache settings:**
   ```bash
   kubectl set env deployment/rbac-service -n iam-namespace \
     DECISION_CACHE_TTL=600 \
     DECISION_CACHE_SIZE=50000
   ```

3. **Scale RBAC service horizontally:**
   ```bash
   kubectl scale deployment -n iam-namespace rbac-service --replicas=8
   ```

**Level 2 (Resolution):**
1. **Optimize policy indexes:**
   ```sql
   CREATE INDEX IF NOT EXISTS idx_policies_combined ON iam_schema.policies 
   USING GIN (to_tsvector('english', resource_pattern || ' ' || action_pattern));
   ANALYZE iam_schema.policies;
   ```

2. **Implement evaluation optimizations:**
   ```bash
   kubectl apply -f policy-evaluation-optimizations.yaml -n iam-namespace
   ```

3. **Restructure complex policies:**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/optimize-policies.sh --tenant-id=<tenant-id>
   ```

4. **Implement parallel policy evaluation:**
   ```bash
   kubectl set env deployment/rbac-service -n iam-namespace \
     PARALLEL_POLICY_EVALUATION=true \
     PARALLEL_THREADS=8
   ```

#### 2.5 Resolution Verification
1. **Monitor decision times for 30 minutes**
2. **Test authorization in high complexity scenarios**
3. **Check impact on CPU and memory utilization**
4. **Analyze response time distribution (p50, p95, p99)**

#### 2.6 Post-Incident Actions
1. **Establish periodic policy reviews**
2. **Implement monitoring for complex policies**
3. **Develop optimization guides for administrators**
4. **Plan refactoring of problematic policies**

### 3. Scalability Issues During Peak Loads

#### 3.1 Symptoms
- Failures during intensive usage periods (start of day, special events)
- Timeout or overload errors during authentication peaks
- Performance degradation when multiple tenants access simultaneously
- High resource utilization alerts

#### 3.2 Initial Checks
1. **Check resource utilization history:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=max_over_time(container_cpu_usage_seconds_total{namespace="iam-namespace"}[1d])'
   ```

2. **Analyze request metrics:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=sum(rate(http_requests_total{namespace="iam-namespace"}[5m])) by (service)'
   ```

3. **Check autoscaling configurations:**
   ```bash
   kubectl get hpa -n iam-namespace
   ```

4. **Analyze connection statistics:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-service-pod-name> -- netstat -ant | grep ESTABLISHED | wc -l
   ```

#### 3.3 Advanced Diagnosis
1. **Analyze usage patterns by time:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query_range?query=sum(rate(http_requests_total{namespace="iam-namespace"}[5m]))&start=2025-05-08T00:00:00Z&end=2025-05-09T00:00:00Z&step=1h'
   ```

2. **Check load distribution by tenant:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=sum(rate(http_requests_total{namespace="iam-namespace"}[5m])) by (tenant_id)'
   ```

3. **Check dependency connection bottlenecks:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-service-pod-name> -- /app/scripts/connection-statistics.sh
   ```

4. **Analyze resource saturation:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_irate{namespace="iam-namespace"}) / sum(kube_pod_container_resource_limits_cpu_cores{namespace="iam-namespace"})'
   ```

#### 3.4 Corrective Actions

**Level 1 (Quick Mitigation):**
1. **Increase resource limits:**
   ```bash
   kubectl set resources deployment -n iam-namespace auth-service --limits=cpu=8,memory=16Gi --requests=cpu=4,memory=8Gi
   kubectl set resources deployment -n iam-namespace rbac-service --limits=cpu=8,memory=16Gi --requests=cpu=4,memory=8Gi
   ```

2. **Scale services horizontally:**
   ```bash
   kubectl scale deployment -n iam-namespace auth-service --replicas=10
   kubectl scale deployment -n iam-namespace rbac-service --replicas=10
   kubectl scale deployment -n iam-namespace token-service --replicas=6
   ```

3. **Optimize connection settings:**
   ```bash
   kubectl set env deployment -n iam-namespace auth-service \
     CONNECTION_POOL_MAX_SIZE=200 \
     CONNECTION_POOL_IDLE_TIMEOUT=600
   ```

**Level 2 (Resolution):**
1. **Implement autoscaling:**
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

2. **Implement tenant rate limiting:**
   ```bash
   kubectl apply -f tenant-rate-limiting.yaml -n iam-namespace
   ```

3. **Configure distributed cache strategy:**
   ```bash
   kubectl apply -f distributed-cache-config.yaml -n iam-namespace
   ```

4. **Optimize database configurations for peak hours:**
   ```bash
   kubectl apply -f peak-hours-database-config.yaml -n iam-namespace
   ```

#### 3.5 Resolution Verification
1. **Run load tests simulating peaks**
2. **Monitor autoscaling behavior**
3. **Check request distribution between pods**
4. **Monitor response times during scaling**

#### 3.6 Post-Incident Actions
1. **Establish provisioning patterns based on usage**
2. **Implement capacity prediction based on trends**
3. **Document optimized configurations for peak periods**
4. **Develop procedures for proactive load balancing**

### 4. Multi-Tenant Performance Issues

#### 4.1 Symptoms
- Inconsistent performance between different tenants
- A specific tenant experiences degradation while others work normally
- Alerts for excessive resource usage by one tenant
- Slow IAM operations only for specific tenant

#### 4.2 Initial Checks
1. **Check metrics by tenant:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=sum(rate(http_requests_total{namespace="iam-namespace"}[5m])) by (tenant_id)'
   ```

2. **Analyze data volume by tenant:**
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

3. **Check resources allocated by tenant:**
   ```bash
   kubectl get resourcequota -n iam-namespace
   ```

4. **Analyze cache statistics by tenant:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli --scan --pattern "tenant:*" | sort | uniq -c
   ```

#### 4.3 Advanced Diagnosis
1. **Analyze query isolation by tenant:**
   ```sql
   SELECT tenant_id, query, calls, total_time, mean_time
   FROM pg_stat_statements pss
   JOIN iam_schema.tenant_sessions ts ON pss.userid = ts.session_id
   ORDER BY mean_time DESC LIMIT 20;
   ```

2. **Check resource isolation issues:**
   ```bash
   kubectl top pods -n iam-namespace --sort-by=cpu
   kubectl top pods -n iam-namespace --sort-by=memory
   ```

3. **Analyze complexity by tenant:**
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

4. **Check access patterns by tenant:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=topk(10, sum(rate(http_requests_total{namespace="iam-namespace"}[1h])) by (tenant_id, path))'
   ```

#### 4.4 Corrective Actions

**Level 1 (Quick Mitigation):**
1. **Apply tenant resource limits:**
   ```bash
   kubectl apply -f tenant-resource-limits.yaml -n iam-namespace
   ```

2. **Restart services for specific tenant:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-service-pod-name> -- /app/scripts/restart-tenant-services.sh <problematic-tenant-id>
   ```

3. **Clear caches for specific tenant:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli --scan --pattern "tenant:<problematic-tenant-id>:*" | xargs redis-cli DEL
   ```

**Level 2 (Resolution):**
1. **Implement tenant partitioning:**
   ```sql
   -- Partition large tables by tenant_id
   ALTER TABLE iam_schema.users PARTITION BY LIST (tenant_id);
   ALTER TABLE iam_schema.policies PARTITION BY LIST (tenant_id);
   ```

2. **Configure dedicated connection pools by tenant:**
   ```bash
   kubectl apply -f tenant-dedicated-pools.yaml -n iam-namespace
   ```

3. **Implement tenant cache isolation:**
   ```bash
   kubectl apply -f tenant-isolated-cache.yaml -n iam-namespace
   ```

4. **Optimize policies for specific tenant:**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/optimize-tenant-policies.sh <problematic-tenant-id>
   ```

#### 4.5 Resolution Verification
1. **Monitor performance of problematic tenant**
2. **Check impact on other tenants**
3. **Analyze resource utilization metrics**
4. **Test tenant-specific critical operations**

#### 4.6 Post-Incident Actions
1. **Implement tenant-specific alerts**
2. **Establish usage limits by tenant size**
3. **Develop isolation plans for large-scale tenants**
4. **Document multi-tenant configuration best practices**

## Additional Resources

### Diagnostic Tools

1. **Diagnostic Scripts:**
   - `/opt/innovabiz/iam/scripts/performance-analyzer.sh`
   - `/opt/innovabiz/iam/scripts/tenant-resource-usage.sh`
   - `/opt/innovabiz/iam/scripts/database-performance-check.sh`

2. **Monitoring Dashboards:**
   - Grafana IAM Performance: `https://grafana.innovabiz.com/d/iam-performance`
   - Grafana Multi-Tenant: `https://grafana.innovabiz.com/d/iam-tenant-metrics`
   - Prometheus: `https://prometheus.innovabiz.com/graph`

3. **Useful Database Queries:**
   ```sql
   -- Identify slow queries in IAM services
   SELECT substring(query, 1, 100) as query_excerpt, 
          calls, total_time, mean_time, max_time,
          stddev_time, rows
   FROM pg_stat_statements
   WHERE query ILIKE '%iam_schema%'
   ORDER BY mean_time DESC
   LIMIT 20;
   
   -- Check unused indexes
   SELECT s.schemaname,
          s.relname as tablename,
          s.indexrelname as indexname,
          s.idx_scan as index_scans
   FROM pg_stat_user_indexes s
   JOIN pg_index i ON s.indexrelid = i.indexrelid
   WHERE s.schemaname = 'iam_schema' AND s.idx_scan = 0
   ORDER BY s.relname, s.indexrelname;
   
   -- Analyze database locks
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

### References

- [IAM Infrastructure Requirements](../04-Infraestrutura/IAM_Infrastructure_Requirements.md)
- [IAM Technical Architecture](../02-Arquitetura/IAM_Technical_Architecture.md)
- [Multi-Tenant Architecture](../02-Arquitetura/Arquitetura_Multi_Tenant.md)
- [IAM Operational Guide](../08-Operacoes/IAM_Operational_Guide.md)
- [IAM Data Model](../03-Desenvolvimento/IAM_Data_Model.md)

### Escalation Contacts

| Level | Team | Contact | Trigger |
|-------|--------|---------|------------|
| 1 | IAM Support | iam-support@innovabiz.com | Initial issues |
| 2 | IAM Operations | iam-ops@innovabiz.com | After 30 min without L1 resolution |
| 3 | IAM DevOps | iam-devops@innovabiz.com | Complex infrastructure issues |
| 4 | DBA | database-admin@innovabiz.com | Critical database issues |
| 5 | Architecture | architecture@innovabiz.com | Structural performance issues |
