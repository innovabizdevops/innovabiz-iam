# IAM Backup and Recovery Procedures

## Introduction

This document describes the detailed procedures for backup and recovery of the IAM module of the INNOVABIZ platform. These procedures are essential to ensure operational continuity, protection of critical data, and compliance with regulatory requirements.

## Backup Component Matrix

| Component | Criticality | Backup Frequency | Retention | Method |
|------------|-------------|----------------------|----------|---------|
| IAM Database | Critical | Daily (full), 15 min (incremental) | 7 days (incremental), 1 year (full) | Native PostgreSQL + WAL shipping |
| IAM Configurations | High | After each change | 1 year | Git + CI/CD System |
| Cryptographic keys | Critical | Weekly + after changes | 5 years | HSM + secure vault |
| Audit logs | High | Daily | 7 years | Log shipping + immutable storage |
| Policies and roles | High | Daily + after changes | 1 year | Export + versioning |
| Tenant metadata | High | Daily | 1 year | Structured export |

## Backup Strategy

### 1. Database Backup

#### 1.1 Full Backup

**Frequency**: Daily, during maintenance window (02:00 - 04:00)

**Procedure**:
```bash
# Full PostgreSQL backup with roles, configurations, and data
pg_dump -h <db-host> -U <backup-user> -d iam_database -F custom -Z 9 -f /backup/iam/full/iam_$(date +%Y%m%d).dump

# Verify backup integrity
pg_restore -l /backup/iam/full/iam_$(date +%Y%m%d).dump > /dev/null

# Apply retention policies
find /backup/iam/full/ -name "iam_*.dump" -mtime +365 -delete
```

**Storage destinations**:
1. Primary backup storage (on-premise)
2. Secondary storage (alternate region)
3. Immutable long-term storage (for monthly backups)

#### 1.2 Incremental Backup

**Frequency**: Every 15 minutes

**Procedure**:
```bash
# Continuous WAL shipping configuration
cat << EOF >> postgresql.conf
wal_level = replica
archive_mode = on
archive_command = 'test ! -f /backup/iam/wal/%f && cp %p /backup/iam/wal/%f'
EOF

# WAL management script
/opt/innovabiz/iam/scripts/wal-management.sh --retention=7
```

**Replication setup**:
```bash
# Standby server setup with physical replication
pg_basebackup -h <primary-db-host> -D <standby-data-dir> -U <replication-user> -P -R
```

### 2. Configuration Backup

**Frequency**: Continuous (via version control system) + weekly (snapshot)

**Procedure**:
```bash
# Export current configurations
kubectl get configmap -n iam-namespace -o yaml > /backup/iam/config/configmaps_$(date +%Y%m%d).yaml
kubectl get secret -n iam-namespace -o yaml > /backup/iam/config/secrets_$(date +%Y%m%d).yaml

# Remove sensitive data for storage (secrets are stored in separate vault)
/opt/innovabiz/iam/scripts/sanitize-configs.sh /backup/iam/config/secrets_$(date +%Y%m%d).yaml
```

**Versioning**:
- All configurations are maintained in Git repository with complete history
- Changes go through approval process and CI/CD
- Release tags are created for each production version

### 3. Cryptographic Keys Backup

**Frequency**: Weekly and after any change or rotation

**Procedure**:
```bash
# Export public keys and certificates
/opt/innovabiz/iam/scripts/export-certificates.sh -o /backup/iam/keys/certs_$(date +%Y%m%d).tar.gz

# Backup sensitive cryptographic material (via HSM)
/opt/innovabiz/iam/scripts/backup-hsm.sh -d /backup/iam/keys/hsm_$(date +%Y%m%d).secure

# Verify key backup
/opt/innovabiz/iam/scripts/verify-keys-backup.sh /backup/iam/keys/hsm_$(date +%Y%m%d).secure
```

**Security considerations**:
- Private cryptographic material never leaves HSM in clear text
- HSM backups are encrypted with multi-custodian keys (M-of-N)
- Storage in secure physical vault, in addition to digital backup

### 4. Audit Logs Backup

**Frequency**: Continuous (near real-time) and daily consolidation

**Procedure**:
```bash
# Fluentd configuration for continuous collection
cat << EOF > fluentd-iam-audit.conf
<match iam.audit.**>
  @type copy
  <store>
    @type elasticsearch
    host audit-es-cluster
    port 9200
    index_name iam-audit-#{Time.now.strftime('%Y.%m.%d')}
    flush_interval 10s
  </store>
  <store>
    @type s3
    aws_key_id #{ENV['AWS_ACCESS_KEY_ID']}
    aws_sec_key #{ENV['AWS_SECRET_ACCESS_KEY']}
    s3_bucket audit-logs-#{ENV['ENVIRONMENT']}
    path iam/audit-logs/%Y/%m/%d/
    buffer_path /var/log/td-agent/s3
    time_slice_format %Y%m%d%H
    time_slice_wait 10m
  </store>
</match>
EOF

# Daily log integrity verification
/opt/innovabiz/iam/scripts/verify-audit-logs.sh --date=$(date -d "yesterday" +%Y-%m-%d)
```

**Immutability and Integrity**:
- Logs stored with WORM (Write Once Read Many) for compliance
- Digital signature of log files for integrity verification
- Storage in multiple geographic locations

## Recovery Procedures

### 1. Complete Database Recovery

**Use cases**:
- Complete loss of database
- Severe data corruption
- Migration to new infrastructure

**Procedure**:

1. **Target environment preparation**:
   ```bash
   # Check available space
   df -h /var/lib/postgresql/data/
   
   # Stop services accessing the database
   kubectl scale deployment -n iam-namespace --replicas=0 auth-service rbac-service user-management-service
   
   # Prepare empty PostgreSQL server
   initdb -D /var/lib/postgresql/data/
   ```

2. **Full backup restoration**:
   ```bash
   # Restore most recent backup
   LATEST_BACKUP=$(ls -t /backup/iam/full/iam_*.dump | head -1)
   pg_restore -h <db-host> -U <admin-user> -d postgres -c -C $LATEST_BACKUP
   
   # Apply WALs for point-in-time recovery (if needed)
   cp /backup/iam/recovery.conf /var/lib/postgresql/data/
   systemctl restart postgresql
   ```

3. **Post-restore verification**:
   ```bash
   # Check database integrity
   psql -h <db-host> -U <admin-user> -d iam_database -c "SELECT count(*) FROM iam_schema.users;"
   psql -h <db-host> -U <admin-user> -d iam_database -c "SELECT count(*) FROM iam_schema.tenants;"
   
   # Check replication (if applicable)
   psql -h <db-host> -U <admin-user> -d iam_database -c "SELECT * FROM pg_stat_replication;"
   ```

4. **Service restoration**:
   ```bash
   # Start services with reconciliation mode
   kubectl set env deployment -n iam-namespace auth-service RECOVERY_MODE=true
   kubectl scale deployment -n iam-namespace --replicas=1 auth-service
   
   # Monitor logs for errors
   kubectl logs -n iam-namespace -f deployment/auth-service
   
   # Start other services after validation
   kubectl scale deployment -n iam-namespace --replicas=3 rbac-service user-management-service
   ```

### 2. Point-in-Time Recovery

**Use cases**:
- Operational error (e.g., accidental data deletion)
- Partial data corruption
- Recovery after security incident

**Procedure**:

1. **Identify recovery point**:
   ```bash
   # Determine timestamp for recovery
   RECOVERY_TIME="2025-05-08 14:30:00"
   
   # Identify full backup prior to recovery point
   BACKUP_FILE=$(ls -t /backup/iam/full/iam_*.dump | grep -A1 "`date -d "$RECOVERY_TIME" +%Y%m%d`" | tail -1)
   ```

2. **Prepare recovery environment**:
   ```bash
   # Create dedicated server for recovery
   kubectl apply -f recovery-database-instance.yaml
   
   # Configure recovery.conf for PITR
   cat << EOF > recovery.conf
   restore_command = 'cp /backup/iam/wal/%f %p'
   recovery_target_time = '$RECOVERY_TIME'
   recovery_target_action = 'promote'
   EOF
   ```

3. **Execute PITR restore**:
   ```bash
   # Restore base backup
   pg_restore -h <recovery-db-host> -U <admin-user> -d postgres -c -C $BACKUP_FILE
   
   # Copy recovery.conf and restart to apply WALs
   kubectl cp recovery.conf iam-recovery-pod:/var/lib/postgresql/data/
   kubectl exec iam-recovery-pod -- systemctl restart postgresql
   
   # Monitor progress
   kubectl exec iam-recovery-pod -- tail -f /var/log/postgresql/postgresql.log
   ```

4. **Validate recovered data**:
   ```bash
   # Connect to recovery instance
   psql -h <recovery-db-host> -U <admin-user> -d iam_database
   
   # Run validation queries
   SELECT count(*) FROM iam_schema.users;
   SELECT max(created_at) FROM iam_schema.users;
   SELECT count(*) FROM iam_schema.tenants;
   ```

5. **Extract recovered data or promote instance**:
   ```bash
   # Extract specific data for selective import
   pg_dump -h <recovery-db-host> -U <admin-user> -d iam_database -t iam_schema.users --data-only -f recovered_users.sql
   
   # Or promote recovery instance to production (complete replacement)
   kubectl apply -f promote-recovery-to-production.yaml
   ```

### 3. Configuration and Key Recovery

**Use cases**:
- Invalid configurations affecting service
- Compromised cryptographic keys
- Environment migration

**Configuration recovery procedure**:

1. **Identify version for restoration**:
   ```bash
   # List available configuration snapshots
   ls -la /backup/iam/config/
   
   # Or identify specific Git commit
   git log --pretty=format:"%h - %an, %ar : %s" --graph -n 20
   ```

2. **Restore configurations**:
   ```bash
   # From snapshot
   kubectl apply -f /backup/iam/config/configmaps_20250501.yaml
   
   # Or from Git (preferred method)
   git checkout tags/v2.5.0 -- config/
   kubectl apply -k ./config/overlays/production/
   ```

3. **Verify configuration application**:
   ```bash
   # Check applied configmaps
   kubectl describe configmap -n iam-namespace auth-service-config
   
   # Restart services to apply new configurations
   kubectl rollout restart deployment -n iam-namespace auth-service rbac-service
   ```

**Cryptographic key recovery procedure**:

1. **Restrict access during recovery**:
   ```bash
   # Put services in maintenance mode
   kubectl patch deployment -n iam-namespace -p '{"spec":{"template":{"spec":{"containers":[{"name":"auth-service","env":[{"name":"MAINTENANCE_MODE","value":"true"}]}]}}}}'
   ```

2. **Recover cryptographic material**:
   ```bash
   # Restore from HSM using M-of-N custody
   /opt/innovabiz/iam/scripts/restore-hsm.sh -i /backup/iam/keys/hsm_20250501.secure
   
   # Restore certificates and public keys
   mkdir -p /tmp/certs-restore
   tar -xzf /backup/iam/keys/certs_20250501.tar.gz -C /tmp/certs-restore
   kubectl create secret generic -n iam-namespace iam-certificates --from-file=/tmp/certs-restore --dry-run=client -o yaml | kubectl apply -f -
   ```

3. **Verify key integrity**:
   ```bash
   # Test signing/verification
   /opt/innovabiz/iam/scripts/test-crypto-keys.sh
   
   # Test JWT functionality
   /opt/innovabiz/iam/scripts/test-jwt-signing.sh
   ```

4. **Return to operational mode**:
   ```bash
   # Remove maintenance mode
   kubectl patch deployment -n iam-namespace -p '{"spec":{"template":{"spec":{"containers":[{"name":"auth-service","env":[{"name":"MAINTENANCE_MODE","value":"false"}]}]}}}}'
   
   # Verify functionality
   curl -v https://iam-api.innovabiz.com/health/crypto
   ```

## Recovery Testing and Validation

### 1. Regular Testing

Regular testing of backup and recovery procedures is mandatory:

| Test Type | Frequency | Scope |
|---------------|------------|--------|
| Full DB recovery test | Bimonthly | Complete restoration in isolated environment |
| PITR test | Quarterly | Recovery to specific point |
| Key recovery test | Semi-annually | HSM + certificates |
| Disaster recovery test | Annual | Complete scenario with region change |

### 2. Standardized Testing Procedure

1. **Preparation**:
   - Create isolated environment for testing
   - Prepare known validation data
   - Document current state for comparison

2. **Execution**:
   - Follow exact production procedures
   - Measure execution time for each step
   - Document issues encountered

3. **Validation**:
   - Verify integrity of restored data
   - Test application functionality
   - Validate integration operations

4. **Documentation**:
   - Record achieved RTO and RPO
   - Document necessary improvements
   - Update runbooks with lessons learned

### 3. Post-Recovery Validation Checklist

- [ ] Record count matches backup moment
- [ ] Referential consistency between tables
- [ ] Users can authenticate successfully
- [ ] Tokens are validated correctly
- [ ] Authorizations are applied as expected
- [ ] Tenants are properly isolated
- [ ] Configurations are correctly applied
- [ ] Audit logs are being generated
- [ ] Monitoring metrics are functional
- [ ] Integrations with external systems are operational

## Governance and Compliance

### 1. Responsibilities

| Role | Responsibilities |
|--------|-------------------|
| DBA | Execution and monitoring of database backups |
| DevOps | Configuration and infrastructure backup |
| Security | Key and cryptographic material management |
| IAM Admin | Functional validation after recovery |
| Compliance | Process compliance verification |

### 2. Documentation and Auditing

- All backup and recovery procedures must be documented in audit logs
- Recovery tests must generate formal reports for auditing
- Deviations from backup policy must go through formal exception process
- Backup/recovery records must be kept according to the retention policy

### 3. Regulatory Requirements

| Region | Regulation | Specific Requirement |
|--------|-------------|----------------------|
| EU/Portugal | GDPR | 72h recovery guarantee, data integrity demonstration |
| Brazil | LGPD | Audit logs for 6 months, integrity validation |
| Angola | PNDSB | Local + remote backup, 48h recovery for critical data |
| USA | HIPAA | Documented recovery tests, encryption in transit and at rest |

## Reporting and Monitoring

### 1. Daily Reports

- Success/failure status of all scheduled backups
- Backup size and growth trends
- Execution time for each backup operation
- Alerts about failures or delays in backup processes

### 2. Continuous Monitoring

- Verification of recent backup integrity
- Validation of storage media accessibility
- Monitoring of replication and WAL archiving
- Alerts about failures in continuous archiving chain

### 3. Backup/Recovery Dashboard

- Visualization of current backup strategy state
- Current vs. target RPO/RTO indicators
- History of restore operations and tests
- Compliance status with backup policy

## References

- [IAM Infrastructure Requirements](../04-Infraestrutura/IAM_Infrastructure_Requirements.md)
- [IAM Technical Architecture](../02-Arquitetura/IAM_Technical_Architecture.md)
- [IAM Operational Guide](../08-Operacoes/IAM_Operational_Guide.md)
- [IAM Compliance Framework](../10-Governanca/IAM_Compliance_Framework_EN.md)

## Appendices

### A. Recovery Plan Template

```markdown
# IAM Recovery Plan - Incident #[ID]

## Incident Details
- **Detection Date/Time**: [YYYY-MM-DD HH:MM]
- **Incident Type**: [Data Corruption / System Failure / Operational Error]
- **Impact**: [Critical / High / Medium / Low]
- **Affected Components**: [DB / Configurations / Keys / Integrations]

## Recovery Objective
- **Recovery Point Objective (RPO)**: [Date/Time]
- **Recovery Time Objective (RTO)**: [Duration]
- **Desired End State**: [Description]

## Recovery Team
- **Coordinator**: [Name] - [Contact]
- **DBA**: [Name] - [Contact]
- **DevOps**: [Name] - [Contact]
- **Security**: [Name] - [Contact]
- **Functional Validator**: [Name] - [Contact]

## Execution Plan
1. **Preparation (T+0h)**
   - [ ] Establish crisis room
   - [ ] Notify stakeholders
   - [ ] Identify required backups
   
2. **Restoration (T+1h)**
   - [ ] Execute procedure X
   - [ ] Apply configurations Y
   
3. **Validation (T+3h)**
   - [ ] Execute validation checklist
   - [ ] Confirm data integrity
   
4. **Activation (T+4h)**
   - [ ] Redirect traffic
   - [ ] Monitor stability
   - [ ] Communicate completion

## Contingency Plan
- **If procedure X fails**: [Alternative action]
- **If validation fails**: [Rollback procedure]
- **If time exceeds RTO**: [Simplified approach]

## Approvals
- [ ] Recovery Coordinator
- [ ] CISO or representative
- [ ] Operations Manager
```

### B. Scenario-based Recovery Matrix

| Scenario | Primary Procedure | Alternative | Estimated Time |
|---------|------------------------|-------------|----------------|
| User data corruption | PITR Recovery | Selective restoration | 2-4 hours |
| Complete database failure | Full restoration + WAL | Standby promotion | 1-2 hours |
| Key compromise | HSM restoration + rotation | New key generation | 2-3 hours |
| Regional disaster recovery | DR region activation | Rebuild in new region | 4-8 hours |
| Post-attack recovery | Restoration to safe point + patches | Clean rebuild | 8-12 hours |
