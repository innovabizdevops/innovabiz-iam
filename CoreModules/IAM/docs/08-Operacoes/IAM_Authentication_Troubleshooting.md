# IAM Authentication Troubleshooting Procedures

## Introduction

This document provides detailed procedures for diagnosing and resolving authentication-related issues in the IAM module of the INNOVABIZ platform. It is intended for operations teams, IAM administrators, and technical support professionals responsible for maintaining continuous availability of authentication services.

## Common Issues Matrix

| Symptom | Possible Cause | Severity | Impact | Average Resolution Time |
|---------|----------------|-----------|---------|--------------------------|
| Massive login failures | Authentication service unavailable | Critical | Users unable to access the system | 30-60 minutes |
| MFA not working | Issue with MFA provider or synchronization failure | High | Users with MFA cannot authenticate | 15-45 minutes |
| Premature token expiration | Incorrect configuration, NTP synchronization, corrupted keys | High | Frequent user disconnections | 30-60 minutes |
| SSO error | Issue with external identity provider or SAML/OIDC configuration | High | Federated users cannot access | 45-90 minutes |
| Intermittent authentication failures | Resource overload or connectivity issues | Medium | Degraded user experience | 30-60 minutes |
| Slow authentication | Database or cache performance issues | Medium | Delays in login process | 30-90 minutes |

## Troubleshooting Procedures

### 1. Massive Login Failures

#### 1.1 Symptoms
- Multiple reports of login failure
- Sharp increase in 401/403 errors in authentication APIs
- Authentication service availability alerts
- Increased response time for login operations

#### 1.2 Initial Checks
1. **Check authentication service status:**
   ```bash
   kubectl get pods -n iam-namespace | grep auth-service
   kubectl logs -n iam-namespace <auth-service-pod-name> --tail=100
   ```

2. **Check availability metrics:**
   - Access IAM availability Grafana dashboard
   - Check for connection failures with dependencies

3. **Check database connectivity:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-service-pod-name> -- pg_isready -h <db-host> -p <db-port>
   ```

4. **Check Redis (session cache) status:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli ping
   ```

#### 1.3 Advanced Diagnosis
1. **Detailed log analysis:**
   ```bash
   kubectl logs -n iam-namespace <auth-service-pod-name> --tail=500 | grep ERROR
   ```

2. **Check configuration issues:**
   ```bash
   kubectl describe configmap -n iam-namespace auth-service-config
   kubectl get secret -n iam-namespace auth-service-secrets -o yaml
   ```

3. **Analyze performance metrics:**
   - CPU, memory, and latency of authentication service
   - Database response time

4. **Check recent changes:**
   - Review recent deployments
   - Configuration or policy changes

#### 1.4 Corrective Actions

**Level 1 (Quick Mitigation):**
1. **Restart authentication service:**
   ```bash
   kubectl rollout restart deployment -n iam-namespace auth-service
   ```

2. **Clear configuration cache:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli FLUSHDB
   ```

3. **Activate contingency mode** (if available):
   ```bash
   kubectl patch deployment -n iam-namespace auth-service -p '{"spec":{"template":{"spec":{"containers":[{"name":"auth-service","env":[{"name":"CONTINGENCY_MODE","value":"true"}]}]}}}}'
   ```

**Level 2 (Resolution):**
1. **Roll back to previous stable version:**
   ```bash
   kubectl rollout undo deployment -n iam-namespace auth-service
   ```

2. **Check and fix database issues:**
   - Analyze locks or pending connections
   - Optimize slow queries

3. **Apply configuration fixes:**
   - Update connection parameters
   - Adjust timeouts and resource limits

#### 1.5 Resolution Verification
1. **Test authentication with test users across different tenants**
2. **Monitor authentication success rates for 15 minutes**
3. **Check logs for recurring errors**
4. **Confirm normalized performance metrics**

#### 1.6 Post-Incident Actions
1. **Perform root cause analysis (RCA)**
2. **Document lessons learned**
3. **Implement preventive measures**
4. **Update procedures if necessary**

### 2. Multi-Factor Authentication (MFA) Issues

#### 2.1 Symptoms
- Users report inability to complete MFA authentication
- MFA codes being rejected
- Failure to generate new MFA tokens

#### 2.2 Initial Checks
1. **Check MFA-specific logs:**
   ```bash
   kubectl logs -n iam-namespace <auth-service-pod-name> | grep -i "mfa\|totp\|authenticator"
   ```

2. **Check MFA provider status:**
   - If external integration: check service status
   - If internal: check MFA module status

3. **Check time synchronization:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-service-pod-name> -- date
   ```

4. **Perform login attempt with MFA test account:**
   - Test across different browsers/devices
   - Test alternative methods (SMS, email, TOTP)

#### 2.3 Advanced Diagnosis
1. **Analyze MFA configurations:**
   ```bash
   kubectl get configmap -n iam-namespace mfa-config -o yaml
   ```

2. **Check secret keys and certificates:**
   ```bash
   kubectl get secret -n iam-namespace mfa-secrets -o yaml
   ```

3. **Trace complete authentication flow:**
   - Analyze specific transaction logs
   - Check detailed error messages

4. **Check database issues:**
   ```sql
   SELECT * FROM iam_schema.mfa_devices WHERE last_error IS NOT NULL ORDER BY updated_at DESC LIMIT 10;
   SELECT COUNT(*) FROM iam_schema.mfa_failed_attempts WHERE attempt_time > NOW() - INTERVAL '1 HOUR' GROUP BY user_id;
   ```

#### 2.4 Corrective Actions

**Level 1 (Quick Mitigation):**
1. **Restart MFA service:**
   ```bash
   kubectl rollout restart deployment -n iam-namespace mfa-service
   ```

2. **Synchronize system clocks:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-pod-name> -- ntpd -gq
   ```

3. **Clear MFA device cache:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli DEL mfa:devicecache:*
   ```

**Level 2 (Resolution):**
1. **Revoke and request re-enrollment of problematic devices:**
   ```sql
   UPDATE iam_schema.mfa_devices SET status = 'REVOKED' WHERE device_id IN (SELECT device_id FROM iam_schema.mfa_failed_attempts GROUP BY device_id HAVING COUNT(*) > 5);
   ```

2. **Update MFA configurations:**
   ```bash
   kubectl apply -f updated-mfa-config.yaml
   ```

3. **Renew certificates or security keys:**
   ```bash
   kubectl create secret generic mfa-secrets --from-file=./new-mfa-keys/ -n iam-namespace
   ```

#### 2.5 Resolution Verification
1. **Test complete MFA process with test accounts**
2. **Verify new codes are being accepted**
3. **Monitor MFA verification success rates**
4. **Check logs for new errors**

#### 2.6 Post-Incident Actions
1. **Notify users about resolution**
2. **Document cause and solution**
3. **Evaluate implementation of alternative MFA methods**
4. **Update user guides if necessary**

### 3. Premature Token Expiration Issues

#### 3.1 Symptoms
- Users report frequent disconnections
- JWT tokens expire before configured time
- Sessions ending unexpectedly

#### 3.2 Initial Checks
1. **Check token expiration configurations:**
   ```bash
   kubectl get configmap -n iam-namespace auth-config -o yaml | grep -i "expir\|timeout\|duration"
   ```

2. **Check token validation logs:**
   ```bash
   kubectl logs -n iam-namespace <auth-service-pod-name> | grep -i "token\|jwt\|expir\|valid"
   ```

3. **Check clock synchronization:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-pod-name> -- date
   kubectl exec -it -n iam-namespace <api-gateway-pod> -- date
   ```

4. **Check JWT signing key status:**
   ```bash
   kubectl get secret -n iam-namespace jwt-keys -o yaml
   ```

#### 3.3 Advanced Diagnosis
1. **Decode and analyze problematic tokens:**
   - Use tools like jwt.io to analyze tokens
   - Check specific claims (exp, iat, nbf)

2. **Analyze session storage:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli --scan --pattern "session:*" | head -n 10
   ```

3. **Check behavior across different tenants:**
   - Test across multiple tenants
   - Check tenant-specific configurations

4. **Analyze signing key integrity:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-pod-name> -- openssl rsa -check -in /path/to/private.key
   ```

#### 3.4 Corrective Actions

**Level 1 (Quick Mitigation):**
1. **Temporarily increase expiration time:**
   ```bash
   kubectl patch configmap -n iam-namespace auth-config --type merge -p '{"data":{"JWT_EXPIRATION_SECONDS":"7200"}}'
   ```

2. **Restart authentication service:**
   ```bash
   kubectl rollout restart deployment -n iam-namespace auth-service
   ```

3. **Synchronize system clocks:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-pod-name> -- ntpd -gq
   ```

**Level 2 (Resolution):**
1. **Rotate JWT signing keys:**
   ```bash
   # Generate new keys
   openssl genrsa -out private_key.pem 4096
   openssl rsa -in private_key.pem -pubout -out public_key.pem
   
   # Update secrets
   kubectl create secret generic jwt-keys --from-file=./private_key.pem --from-file=./public_key.pem -n iam-namespace --dry-run=client -o yaml | kubectl apply -f -
   ```

2. **Adjust token configurations:**
   ```bash
   kubectl apply -f updated-token-config.yaml
   ```

3. **Implement gradual token rotation:**
   ```bash
   kubectl patch deployment -n iam-namespace auth-service -p '{"spec":{"template":{"spec":{"containers":[{"name":"auth-service","env":[{"name":"GRADUAL_TOKEN_ROTATION","value":"true"}]}]}}}}'
   ```

#### 3.5 Resolution Verification
1. **Monitor active session durations**
2. **Check logs for token validation errors**
3. **Test token renewal with test accounts**
4. **Verify behavior across various devices/browsers**

#### 3.6 Post-Incident Actions
1. **Implement improved monitoring for token expiration**
2. **Review key rotation strategy**
3. **Document expected expiration behavior**
4. **Update operational procedures**

## SSO (Single Sign-On) Related Issues

### 4.1 Symptoms
- Authentication failures via external providers (Google, Microsoft, etc.)
- Errors during redirection to/from identity providers
- Incorrect or expired SAML/OIDC metadata

### 4.2 Initial Checks
1. **Check SSO-specific logs:**
   ```bash
   kubectl logs -n iam-namespace <auth-service-pod-name> | grep -i "sso\|saml\|oidc\|oauth"
   ```

2. **Check SSO endpoint status:**
   ```bash
   curl -I https://<auth-domain>/sso/health
   ```

3. **Check SAML/OIDC metadata:**
   ```bash
   curl https://<auth-domain>/sso/metadata
   ```

4. **Check SSO certificates:**
   ```bash
   kubectl get secret -n iam-namespace sso-certificates -o yaml
   ```

### 4.3 Advanced Diagnosis
1. **Analyze complete network traffic:**
   - Capture and analyze complete authentication flow
   - Check request/response headers and parameters

2. **Check provider configurations:**
   ```bash
   kubectl get configmap -n iam-namespace sso-providers-config -o yaml
   ```

3. **Test with specific tools:**
   ```bash
   saml2aws validate --provider=<provider>
   oidc-cli validate --issuer=<issuer-url>
   ```

4. **Check external provider status:**
   - Check provider status pages
   - Test other applications with same provider

### 4.4 Corrective Actions

**Level 1 (Quick Mitigation):**
1. **Restart SSO service:**
   ```bash
   kubectl rollout restart deployment -n iam-namespace sso-service
   ```

2. **Refresh federation metadata:**
   ```bash
   kubectl exec -it -n iam-namespace <sso-pod-name> -- /app/scripts/refresh-metadata.sh
   ```

3. **Clear metadata cache:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli DEL sso:metadata:*
   ```

**Level 2 (Resolution):**
1. **Renew SSO certificates:**
   ```bash
   # Generate new certificates
   openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes
   
   # Update secrets
   kubectl create secret generic sso-certificates --from-file=./cert.pem --from-file=./key.pem -n iam-namespace --dry-run=client -o yaml | kubectl apply -f -
   ```

2. **Update provider configurations:**
   ```bash
   kubectl apply -f updated-sso-config.yaml
   ```

3. **Reconfigure attribute mappings:**
   ```bash
   kubectl apply -f updated-attribute-mapping.yaml
   ```

### 4.5 Resolution Verification
1. **Test SSO flow with test accounts**
2. **Verify correct attribute mapping**
3. **Test authentication with different providers**
4. **Check logs for SSO errors**

### 4.6 Post-Incident Actions
1. **Document correct configurations**
2. **Implement SSO-specific monitoring**
3. **Review and update integration documentation**
4. **Configure alerts for certificate expiration**

## Additional Resources

### Diagnostic Tools

1. **Diagnostic Scripts:**
   - `/opt/innovabiz/iam/scripts/auth-diagnostics.sh`
   - `/opt/innovabiz/iam/scripts/token-validator.sh`
   - `/opt/innovabiz/iam/scripts/mfa-tester.sh`

2. **Monitoring Dashboards:**
   - Grafana IAM: `https://grafana.innovabiz.com/d/iam-overview`
   - Prometheus: `https://prometheus.innovabiz.com/graph`

3. **Useful Database Queries:**
   ```sql
   -- Check recent login attempts by status
   SELECT status, COUNT(*) FROM iam_schema.auth_attempts 
   WHERE attempt_time > NOW() - INTERVAL '1 HOUR' 
   GROUP BY status;
   
   -- Check MFA devices with issues
   SELECT user_id, device_id, COUNT(*) as failure_count
   FROM iam_schema.mfa_failed_attempts 
   WHERE attempt_time > NOW() - INTERVAL '24 HOURS'
   GROUP BY user_id, device_id 
   HAVING COUNT(*) > 3;
   
   -- Check recently revoked tokens
   SELECT token_id, user_id, revocation_reason, revoked_at
   FROM iam_schema.revoked_tokens
   WHERE revoked_at > NOW() - INTERVAL '24 HOURS'
   ORDER BY revoked_at DESC;
   ```

### References

- [IAM Security Model](../05-Seguranca/IAM_Security_Model.md)
- [IAM Infrastructure Requirements](../04-Infraestrutura/IAM_Infrastructure_Requirements.md)
- [IAM Technical Architecture](../02-Arquitetura/IAM_Technical_Architecture.md)
- [IAM Compliance Framework](../10-Governanca/IAM_Compliance_Framework_EN.md)
- [IAM Operational Guide](../08-Operacoes/IAM_Operational_Guide.md)

### Escalation Contacts

| Level | Team | Contact | Trigger |
|-------|--------|---------|------------|
| 1 | IAM Support | iam-support@innovabiz.com | Initial issues |
| 2 | IAM Operations | iam-ops@innovabiz.com | After 30 min without L1 resolution |
| 3 | IAM Development | iam-dev@innovabiz.com | After 60 min without L2 resolution |
| 4 | Security Architecture | security-arch@innovabiz.com | Critical security issues |
