# IAM Authorization Troubleshooting Procedures

## Introduction

This document provides detailed procedures for diagnosing and resolving authorization-related issues in the IAM module of the INNOVABIZ platform. It is intended for operations teams, IAM administrators, and technical support professionals responsible for maintaining continuous and secure operation of access control services.

## Common Issues Matrix

| Symptom | Possible Cause | Severity | Impact | Average Resolution Time |
|---------|----------------|-----------|---------|--------------------------|
| Incorrectly denied access | Misconfigured policies or outdated cache | High | Legitimate users cannot access needed resources | 15-45 minutes |
| Slow authorization decisions | Policy engine overload | Medium | Degraded user experience | 30-60 minutes |
| Privilege escalation | Policy misconfiguration or security vulnerability | Critical | Potential system security compromise | 60-120 minutes |
| Role and delegation issues | Error in permission propagation | High | Managers cannot properly delegate permissions | 30-60 minutes |
| Inconsistent authorization decisions | Incorrect policy replication or caching | High | Unpredictable access behavior | 30-90 minutes |
| Segregation of duty problems | Conflicts in SoD policies | Medium | Compliance violations | 45-90 minutes |

## Troubleshooting Procedures

### 1. Incorrectly Denied Access

#### 1.1 Symptoms
- Users report inability to access resources they should have permission for
- 403 (Forbidden) errors on APIs that normally work
- Permissions appear in the administration interface but are not effective

#### 1.2 Initial Checks
1. **Check policies applicable to the user:**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/check-user-policies.sh <user-id> <resource> <action>
   ```

2. **Check authorization-specific logs:**
   ```bash
   kubectl logs -n iam-namespace <rbac-service-pod-name> --tail=100 | grep -i "<user-id>\|<resource>\|denied"
   ```

3. **Check policy cache:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli --scan --pattern "policy:*" | head -n 10
   ```

4. **Check role propagation:**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/check-role-propagation.sh <role-id>
   ```

#### 1.3 Advanced Diagnosis
1. **Analyze policies in detail:**
   ```sql
   SELECT p.policy_id, p.name, p.effect, p.resource_pattern, p.action_pattern, p.conditions 
   FROM iam_schema.policies p 
   JOIN iam_schema.role_policies rp ON p.policy_id = rp.policy_id 
   JOIN iam_schema.user_roles ur ON rp.role_id = ur.role_id 
   WHERE ur.user_id = '<user-id>';
   ```

2. **Check permission hierarchy:**
   ```sql
   WITH RECURSIVE role_hierarchy AS (
     SELECT role_id, parent_role_id 
     FROM iam_schema.roles 
     WHERE role_id IN (SELECT role_id FROM iam_schema.user_roles WHERE user_id = '<user-id>')
     UNION
     SELECT r.role_id, r.parent_role_id 
     FROM iam_schema.roles r 
     JOIN role_hierarchy rh ON r.role_id = rh.parent_role_id
   ) SELECT * FROM role_hierarchy;
   ```

3. **Check policy evaluation order:**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/debug-policy-evaluation.sh <user-id> <resource> <action>
   ```

4. **Check contextual attributes (for ABAC):**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/check-context-attributes.sh <request-id>
   ```

#### 1.4 Corrective Actions

**Level 1 (Quick Mitigation):**
1. **Clear policy cache:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli DEL "policy:user:<user-id>*"
   ```

2. **Restart authorization service:**
   ```bash
   kubectl rollout restart deployment -n iam-namespace rbac-service
   ```

3. **Apply temporary emergency policy (critical cases):**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/apply-emergency-access.sh <user-id> <resource> <action>
   ```

**Level 2 (Resolution):**
1. **Fix incorrect policies:**
   ```sql
   UPDATE iam_schema.policies 
   SET effect = 'allow', conditions = '{"updated_by": "admin_emergency", "reason": "fix_incorrect_deny"}'
   WHERE policy_id = '<problem-policy-id>';
   ```

2. **Assign missing or required role:**
   ```sql
   INSERT INTO iam_schema.user_roles (user_id, role_id, granted_by, grant_reason) 
   VALUES ('<user-id>', '<role-id>', 'admin_emergency', 'missing_role_fix');
   ```

3. **Fix policy priority order:**
   ```sql
   UPDATE iam_schema.policies 
   SET priority = 100 
   WHERE policy_id = '<high-priority-policy-id>';
   ```

#### 1.5 Resolution Verification
1. **Test affected user's access to resources**
2. **Check logs to confirm allowed access**
3. **Confirm the fix didn't create security issues**
4. **Verify behavior across different tenants**

#### 1.6 Post-Incident Actions
1. **Document root cause and solution**
2. **Review policy configuration standards**
3. **Update policy auditing procedures**
4. **Evaluate need for policy validation tools**

### 2. Slow Authorization Decisions

#### 2.1 Symptoms
- Increased response time for operations requiring authorization
- Timeout in API requests during permission checks
- High latency at authorization checkpoints

#### 2.2 Initial Checks
1. **Check performance metrics:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=avg_over_time(rbac_decision_duration_ms[15m])'
   ```

2. **Check resource utilization:**
   ```bash
   kubectl top pods -n iam-namespace | grep rbac
   ```

3. **Check logs for signs of slowness:**
   ```bash
   kubectl logs -n iam-namespace <rbac-service-pod-name> --tail=200 | grep -i "slow\|timeout\|duration"
   ```

4. **Check number of applicable policies:**
   ```sql
   SELECT COUNT(*) FROM iam_schema.policies WHERE tenant_id = '<tenant-id>';
   ```

#### 2.3 Advanced Diagnosis
1. **Analyze slow database queries:**
   ```sql
   SELECT query, calls, total_time, mean_time, max_time 
   FROM pg_stat_statements 
   WHERE query LIKE '%policies%' OR query LIKE '%roles%' 
   ORDER BY mean_time DESC LIMIT 10;
   ```

2. **Check policy growth:**
   ```sql
   SELECT DATE_TRUNC('day', created_at) AS day, COUNT(*) 
   FROM iam_schema.policies 
   GROUP BY day 
   ORDER BY day DESC LIMIT 30;
   ```

3. **Evaluate complex policy patterns:**
   ```sql
   SELECT policy_id, name, LENGTH(conditions::text) as condition_complexity 
   FROM iam_schema.policies 
   ORDER BY condition_complexity DESC LIMIT 20;
   ```

4. **Check cache efficiency:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli INFO stats | grep hit_rate
   ```

#### 2.4 Corrective Actions

**Level 1 (Quick Mitigation):**
1. **Increase authorization service resources:**
   ```bash
   kubectl patch deployment -n iam-namespace rbac-service -p '{"spec":{"template":{"spec":{"containers":[{"name":"rbac-service","resources":{"limits":{"cpu":"2","memory":"4Gi"},"requests":{"cpu":"1","memory":"2Gi"}}}]}}}}'
   ```

2. **Optimize cache configurations:**
   ```bash
   kubectl set env deployment/rbac-service -n iam-namespace POLICY_CACHE_TTL=3600 POLICY_CACHE_SIZE=10000
   ```

3. **Scale service horizontally:**
   ```bash
   kubectl scale deployment -n iam-namespace rbac-service --replicas=5
   ```

**Level 2 (Resolution):**
1. **Optimize database indexes:**
   ```sql
   CREATE INDEX IF NOT EXISTS idx_policies_resource_pattern ON iam_schema.policies (resource_pattern);
   CREATE INDEX IF NOT EXISTS idx_policies_action_pattern ON iam_schema.policies (action_pattern);
   ```

2. **Consolidate redundant policies:**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/consolidate-policies.sh <tenant-id>
   ```

3. **Implement policy partitioning:**
   ```sql
   ALTER TABLE iam_schema.policies PARTITION BY LIST (tenant_id);
   ```

#### 2.5 Resolution Verification
1. **Monitor authorization response times for 30 minutes**
2. **Check CPU and memory utilization under load**
3. **Test authorization in different scenarios**
4. **Confirm solution effectiveness across different tenants**

#### 2.6 Post-Incident Actions
1. **Implement specific monitoring for decision time**
2. **Establish alert thresholds for slow decisions**
3. **Plan periodic policy reviews**
4. **Document optimizations made**

### 3. Privilege Escalation Issues

#### 3.1 Symptoms
- Users with access to resources or actions beyond their expected permissions
- Security alerts about anomalous accesses
- Suspicious patterns of permission usage

#### 3.2 Initial Checks
1. **Check role assignment logs:**
   ```bash
   kubectl logs -n iam-namespace <rbac-service-pod-name> --tail=1000 | grep -i "role\|grant\|assign"
   ```

2. **Check recently modified policies:**
   ```sql
   SELECT policy_id, name, created_at, updated_at, created_by, updated_by 
   FROM iam_schema.policies 
   WHERE updated_at > NOW() - INTERVAL '7 DAYS'
   ORDER BY updated_at DESC;
   ```

3. **Check recent administrative operations:**
   ```sql
   SELECT * FROM iam_schema.admin_audit_log 
   WHERE action_type IN ('CREATE_POLICY', 'UPDATE_POLICY', 'GRANT_ROLE') 
   AND action_time > NOW() - INTERVAL '7 DAYS'
   ORDER BY action_time DESC;
   ```

4. **Check changes to administrative roles:**
   ```sql
   SELECT ur.user_id, u.username, r.role_name, ur.granted_at, ur.granted_by
   FROM iam_schema.user_roles ur
   JOIN iam_schema.users u ON ur.user_id = u.user_id
   JOIN iam_schema.roles r ON ur.role_id = r.role_id
   WHERE r.role_name LIKE '%admin%' AND ur.granted_at > NOW() - INTERVAL '30 DAYS'
   ORDER BY ur.granted_at DESC;
   ```

#### 3.3 Advanced Diagnosis
1. **Analyze potential escalation paths:**
   ```bash
   kubectl exec -it -n iam-namespace <security-pod-name> -- /app/scripts/analyze-privilege-paths.sh <suspicious-user-id>
   ```

2. **Check SoD policy conflicts:**
   ```sql
   SELECT * FROM iam_schema.sod_conflicts 
   WHERE detection_time > NOW() - INTERVAL '7 DAYS';
   ```

3. **Complete permission review for suspicious user:**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/full-permission-report.sh <suspicious-user-id>
   ```

4. **Check for known attack attempts:**
   ```bash
   kubectl logs -n iam-namespace <security-monitoring-pod-name> --tail=5000 | grep -i "attack\|exploit\|privilege\|escalation"
   ```

#### 3.4 Corrective Actions

**Level 1 (Immediate Containment):**
1. **Suspend suspicious user:**
   ```sql
   UPDATE iam_schema.users 
   SET status = 'SUSPENDED', 
       suspension_reason = 'security_investigation', 
       suspended_at = NOW(), 
       suspended_by = 'security_response_team'
   WHERE user_id = '<suspicious-user-id>';
   ```

2. **Remove critical permissions from user:**
   ```sql
   DELETE FROM iam_schema.user_roles 
   WHERE user_id = '<suspicious-user-id>' 
   AND role_id IN (SELECT role_id FROM iam_schema.roles WHERE is_administrative = true);
   ```

3. **Check and revoke active tokens:**
   ```sql
   UPDATE iam_schema.access_tokens 
   SET revoked = true, 
       revocation_reason = 'security_investigation' 
   WHERE user_id = '<suspicious-user-id>' AND revoked = false;
   ```

**Level 2 (Resolution):**
1. **Fix policies with inadequate configuration:**
   ```sql
   UPDATE iam_schema.policies 
   SET effect = 'deny', 
       updated_by = 'security_response_team', 
       updated_at = NOW()
   WHERE policy_id = '<vulnerable-policy-id>';
   ```

2. **Implement stricter security policies:**
   ```bash
   kubectl apply -f stricter-security-policies.yaml -n iam-namespace
   ```

3. **Strengthen SoD checks:**
   ```sql
   INSERT INTO iam_schema.sod_policy (name, description, conflicting_roles, detection_action)
   VALUES ('restrict_sensitive_data_access', 'Prevent data modification and approval by same user', 
           ARRAY['data_modifier', 'approval_officer'], 'prevent_and_alert');
   ```

#### 3.5 Resolution Verification
1. **Run complete security scan**
2. **Check logs for new suspicious patterns**
3. **Validate excessive permissions have been removed**
4. **Test potential escalation paths again**

#### 3.6 Post-Incident Actions
1. **Perform complete forensic analysis**
2. **Document vulnerability and fix**
3. **Implement additional security checks**
4. **Review and update separation of duties policies**
5. **Conduct complete permission audit**

### 4. Role and Delegation Issues

#### 4.1 Symptoms
- Administrators cannot delegate permissions
- Assigned roles are not propagated to users
- Inconsistencies between displayed and applied permissions

#### 4.2 Initial Checks
1. **Check delegation logs:**
   ```bash
   kubectl logs -n iam-namespace <rbac-service-pod-name> --tail=500 | grep -i "delegat\|assign\|grant"
   ```

2. **Check delegation limit configurations:**
   ```bash
   kubectl get configmap -n iam-namespace rbac-config -o yaml | grep -A 10 delegation
   ```

3. **Check user's delegation permissions:**
   ```sql
   SELECT p.policy_id, p.name, p.resource_pattern, p.action_pattern
   FROM iam_schema.policies p 
   JOIN iam_schema.role_policies rp ON p.policy_id = rp.policy_id 
   JOIN iam_schema.user_roles ur ON rp.role_id = ur.role_id 
   WHERE ur.user_id = '<delegator-user-id>'
   AND (p.resource_pattern LIKE '%:role:%' AND p.action_pattern LIKE '%:grant%');
   ```

4. **Check recent delegation events:**
   ```sql
   SELECT * FROM iam_schema.delegation_history
   WHERE delegator_id = '<delegator-user-id>'
   ORDER BY delegation_time DESC LIMIT 20;
   ```

#### 4.3 Advanced Diagnosis
1. **Analyze complete role hierarchy:**
   ```sql
   WITH RECURSIVE role_tree AS (
     SELECT role_id, role_name, parent_role_id, 1 AS level
     FROM iam_schema.roles
     WHERE parent_role_id IS NULL
     UNION ALL
     SELECT r.role_id, r.role_name, r.parent_role_id, rt.level + 1
     FROM iam_schema.roles r
     JOIN role_tree rt ON r.parent_role_id = rt.role_id
   )
   SELECT * FROM role_tree 
   ORDER BY level, role_name;
   ```

2. **Check delegation consistency:**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/validate-delegations.sh <tenant-id>
   ```

3. **Check cross-delegation limitations:**
   ```sql
   SELECT dg.* 
   FROM iam_schema.delegation_graph dg
   JOIN iam_schema.roles r1 ON dg.source_role_id = r1.role_id
   JOIN iam_schema.roles r2 ON dg.target_role_id = r2.role_id
   WHERE r1.organizational_unit_id != r2.organizational_unit_id;
   ```

4. **Check secondary role propagation:**
   ```sql
   WITH user_assigned_roles AS (
     SELECT user_id, role_id FROM iam_schema.user_roles
   ),
   role_implies AS (
     SELECT role_id, implied_role_id FROM iam_schema.role_implications
   )
   SELECT ur.user_id, r.role_name, ri.implied_role_id, r2.role_name AS implied_role_name
   FROM user_assigned_roles ur
   JOIN iam_schema.roles r ON ur.role_id = r.role_id
   JOIN role_implies ri ON r.role_id = ri.role_id
   JOIN iam_schema.roles r2 ON ri.implied_role_id = r2.role_id
   WHERE ur.user_id = '<problem-user-id>';
   ```

#### 4.4 Corrective Actions

**Level 1 (Quick Mitigation):**
1. **Restart RBAC service:**
   ```bash
   kubectl rollout restart deployment -n iam-namespace rbac-service
   ```

2. **Clear role cache:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli DEL "roles:user:<problem-user-id>"
   ```

3. **Force delegation recalculation:**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/recalculate-delegations.sh <tenant-id>
   ```

**Level 2 (Resolution):**
1. **Fix delegation configurations:**
   ```bash
   kubectl apply -f corrected-delegation-config.yaml -n iam-namespace
   ```

2. **Fix role propagation:**
   ```sql
   -- Add missing role implication
   INSERT INTO iam_schema.role_implications (role_id, implied_role_id, grant_type)
   VALUES ('<parent-role-id>', '<implied-role-id>', 'SYSTEM_DEFINED');
   ```

3. **Fix delegation limit:**
   ```sql
   UPDATE iam_schema.delegation_policies
   SET max_depth = 3, allow_cross_department = false
   WHERE policy_id = '<restrictive-policy-id>';
   ```

#### 4.5 Resolution Verification
1. **Test delegation process**
2. **Verify permission propagation**
3. **Validate delegation hierarchy**
4. **Test multi-level delegation scenarios**

#### 4.6 Post-Incident Actions
1. **Document delegation issues and solutions**
2. **Review delegation policies**
3. **Improve validation mechanisms**
4. **Update administration documentation**

## Additional Resources

### Diagnostic Tools

1. **Diagnostic Scripts:**
   - `/opt/innovabiz/iam/scripts/policy-analyzer.sh`
   - `/opt/innovabiz/iam/scripts/permission-checker.sh`
   - `/opt/innovabiz/iam/scripts/delegation-validator.sh`

2. **Monitoring Dashboards:**
   - Grafana RBAC: `https://grafana.innovabiz.com/d/rbac-overview`
   - Prometheus: `https://prometheus.innovabiz.com/graph?g0.expr=rbac_decision_duration_seconds`

3. **Useful Database Queries:**
   ```sql
   -- Check potential policy conflicts (allow and deny on same action/resource)
   SELECT r.resource_pattern, a.action_pattern, 
          COUNT(CASE WHEN p.effect = 'allow' THEN 1 END) as allow_count,
          COUNT(CASE WHEN p.effect = 'deny' THEN 1 END) as deny_count
   FROM iam_schema.policies p,
        LATERAL (SELECT UNNEST(STRING_TO_ARRAY(p.resource_pattern, ',')) as resource_pattern) r,
        LATERAL (SELECT UNNEST(STRING_TO_ARRAY(p.action_pattern, ',')) as action_pattern) a
   GROUP BY r.resource_pattern, a.action_pattern
   HAVING COUNT(CASE WHEN p.effect = 'allow' THEN 1 END) > 0 
      AND COUNT(CASE WHEN p.effect = 'deny' THEN 1 END) > 0;
   
   -- Find policies affecting a specific resource
   SELECT p.*
   FROM iam_schema.policies p
   WHERE resource_pattern LIKE '%<resource-pattern>%'
   OR resource_pattern = '*';
   
   -- Check history of changes to critical policies
   SELECT * FROM iam_schema.policy_change_history
   WHERE policy_id IN (
     SELECT policy_id FROM iam_schema.policies
     WHERE resource_pattern LIKE '%:admin:%' OR resource_pattern LIKE '%:critical:%'
   )
   ORDER BY change_time DESC;
   ```

### References

- [IAM Security Model](../05-Seguranca/IAM_Security_Model.md)
- [IAM Technical Architecture](../02-Arquitetura/IAM_Technical_Architecture.md)
- [IAM Compliance Framework](../10-Governanca/IAM_Compliance_Framework_EN.md)
- [IAM Operational Guide](../08-Operacoes/IAM_Operational_Guide.md)
- [IAM API Documentation](../03-Desenvolvimento/API_Documentation.md)

### Escalation Contacts

| Level | Team | Contact | Trigger |
|-------|--------|---------|------------|
| 1 | IAM Support | iam-support@innovabiz.com | Initial issues |
| 2 | IAM Operations | iam-ops@innovabiz.com | After 30 min without L1 resolution |
| 3 | IAM Development | iam-dev@innovabiz.com | After 60 min without L2 resolution |
| 4 | Data Security | data-security@innovabiz.com | Critical authorization issues |
| 5 | CISO | ciso@innovabiz.com | Confirmed security incidents |
