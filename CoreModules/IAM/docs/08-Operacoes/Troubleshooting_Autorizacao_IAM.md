# Procedimentos de Troubleshooting de Autorização IAM

## Introdução

Este documento fornece procedimentos detalhados para diagnóstico e resolução de problemas relacionados à autorização no módulo IAM da plataforma INNOVABIZ. É destinado às equipes de operações, administradores de IAM e profissionais de suporte técnico responsáveis por manter a operação contínua e segura dos serviços de controle de acesso.

## Matriz de Problemas Comuns

| Sintoma | Possível Causa | Gravidade | Impacto | Tempo Médio de Resolução |
|---------|----------------|-----------|---------|--------------------------|
| Acesso negado incorretamente | Políticas mal configuradas ou cache desatualizado | Alta | Usuários legítimos não conseguem acessar recursos necessários | 15-45 minutos |
| Decisões de autorização lentas | Sobrecarga do motor de políticas | Média | Experiência do usuário degradada | 30-60 minutos |
| Escalação de privilégios | Má configuração de políticas ou vulnerabilidade de segurança | Crítica | Comprometimento potencial da segurança do sistema | 60-120 minutos |
| Problemas em roles e delegações | Erro na propagação de permissões | Alta | Gestores não conseguem delegar permissões corretamente | 30-60 minutos |
| Decisões de autorização inconsistentes | Replicação ou caching incorreto de políticas | Alta | Comportamento imprevisível de acesso | 30-90 minutos |
| Problemas de segregação de função | Conflitos em políticas de SoD | Média | Violações de compliance | 45-90 minutos |

## Procedimentos de Troubleshooting

### 1. Acesso Negado Incorretamente

#### 1.1 Sintomas
- Usuários relatam incapacidade de acessar recursos aos quais deveriam ter permissão
- Erros 403 (Forbidden) em APIs que normalmente funcionam
- Permissões aparecem na interface de administração mas não são efetivas

#### 1.2 Verificações Iniciais
1. **Verificar políticas aplicáveis ao usuário:**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/check-user-policies.sh <user-id> <resource> <action>
   ```

2. **Verificar logs específicos de autorização:**
   ```bash
   kubectl logs -n iam-namespace <rbac-service-pod-name> --tail=100 | grep -i "<user-id>\|<resource>\|denied"
   ```

3. **Verificar cache de políticas:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli --scan --pattern "policy:*" | head -n 10
   ```

4. **Verificar propagação de roles:**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/check-role-propagation.sh <role-id>
   ```

#### 1.3 Diagnóstico Avançado
1. **Analisar políticas em detalhes:**
   ```sql
   SELECT p.policy_id, p.name, p.effect, p.resource_pattern, p.action_pattern, p.conditions 
   FROM iam_schema.policies p 
   JOIN iam_schema.role_policies rp ON p.policy_id = rp.policy_id 
   JOIN iam_schema.user_roles ur ON rp.role_id = ur.role_id 
   WHERE ur.user_id = '<user-id>';
   ```

2. **Verificar hierarquia de permissões:**
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

3. **Verificar ordem de avaliação de políticas:**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/debug-policy-evaluation.sh <user-id> <resource> <action>
   ```

4. **Verificar atributos contextuais (para ABAC):**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/check-context-attributes.sh <request-id>
   ```

#### 1.4 Ações Corretivas

**Nível 1 (Mitigação Rápida):**
1. **Limpar cache de políticas:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli DEL "policy:user:<user-id>*"
   ```

2. **Reiniciar serviço de autorização:**
   ```bash
   kubectl rollout restart deployment -n iam-namespace rbac-service
   ```

3. **Aplicar política temporária de emergência (caso crítico):**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/apply-emergency-access.sh <user-id> <resource> <action>
   ```

**Nível 2 (Resolução):**
1. **Corrigir políticas incorretas:**
   ```sql
   UPDATE iam_schema.policies 
   SET effect = 'allow', conditions = '{"updated_by": "admin_emergency", "reason": "fix_incorrect_deny"}'
   WHERE policy_id = '<problema-policy-id>';
   ```

2. **Atribuir role ausente ou necessária:**
   ```sql
   INSERT INTO iam_schema.user_roles (user_id, role_id, granted_by, grant_reason) 
   VALUES ('<user-id>', '<role-id>', 'admin_emergency', 'missing_role_fix');
   ```

3. **Corrigir ordem de prioridade de políticas:**
   ```sql
   UPDATE iam_schema.policies 
   SET priority = 100 
   WHERE policy_id = '<high-priority-policy-id>';
   ```

#### 1.5 Verificação de Resolução
1. **Testar acesso do usuário afetado aos recursos**
2. **Verificar logs para confirmar acesso permitido**
3. **Confirmar que a correção não criou problemas de segurança**
4. **Verificar comportamento em diferentes tenants**

#### 1.6 Ações Pós-Incidente
1. **Documentar a causa raiz e solução**
2. **Revisar padrões de configuração de políticas**
3. **Atualizar procedimentos de auditoria de políticas**
4. **Avaliar necessidade de ferramentas de validação de políticas**

### 2. Decisões de Autorização Lentas

#### 2.1 Sintomas
- Aumento no tempo de resposta para operações que requerem autorização
- Timeout em requisições de API durante verificações de permissão
- Alta latência em checkpoints de autorização

#### 2.2 Verificações Iniciais
1. **Verificar métricas de performance:**
   ```bash
   kubectl exec -it -n iam-namespace <prometheus-pod-name> -- curl 'http://localhost:9090/api/v1/query?query=avg_over_time(rbac_decision_duration_ms[15m])'
   ```

2. **Verificar utilização de recursos:**
   ```bash
   kubectl top pods -n iam-namespace | grep rbac
   ```

3. **Verificar logs por indícios de lentidão:**
   ```bash
   kubectl logs -n iam-namespace <rbac-service-pod-name> --tail=200 | grep -i "slow\|timeout\|duration"
   ```

4. **Verificar número de políticas aplicáveis:**
   ```sql
   SELECT COUNT(*) FROM iam_schema.policies WHERE tenant_id = '<tenant-id>';
   ```

#### 2.3 Diagnóstico Avançado
1. **Analisar queries de banco de dados lentas:**
   ```sql
   SELECT query, calls, total_time, mean_time, max_time 
   FROM pg_stat_statements 
   WHERE query LIKE '%policies%' OR query LIKE '%roles%' 
   ORDER BY mean_time DESC LIMIT 10;
   ```

2. **Verificar crescimento de políticas:**
   ```sql
   SELECT DATE_TRUNC('day', created_at) AS day, COUNT(*) 
   FROM iam_schema.policies 
   GROUP BY day 
   ORDER BY day DESC LIMIT 30;
   ```

3. **Avaliar padrões complexos de políticas:**
   ```sql
   SELECT policy_id, name, LENGTH(conditions::text) as condition_complexity 
   FROM iam_schema.policies 
   ORDER BY condition_complexity DESC LIMIT 20;
   ```

4. **Verificar eficiência de cache:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli INFO stats | grep hit_rate
   ```

#### 2.4 Ações Corretivas

**Nível 1 (Mitigação Rápida):**
1. **Aumentar recursos do serviço de autorização:**
   ```bash
   kubectl patch deployment -n iam-namespace rbac-service -p '{"spec":{"template":{"spec":{"containers":[{"name":"rbac-service","resources":{"limits":{"cpu":"2","memory":"4Gi"},"requests":{"cpu":"1","memory":"2Gi"}}}]}}}}'
   ```

2. **Otimizar configurações de cache:**
   ```bash
   kubectl set env deployment/rbac-service -n iam-namespace POLICY_CACHE_TTL=3600 POLICY_CACHE_SIZE=10000
   ```

3. **Escalar horizontalmente o serviço:**
   ```bash
   kubectl scale deployment -n iam-namespace rbac-service --replicas=5
   ```

**Nível 2 (Resolução):**
1. **Otimizar índices de banco de dados:**
   ```sql
   CREATE INDEX IF NOT EXISTS idx_policies_resource_pattern ON iam_schema.policies (resource_pattern);
   CREATE INDEX IF NOT EXISTS idx_policies_action_pattern ON iam_schema.policies (action_pattern);
   ```

2. **Consolidar políticas redundantes:**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/consolidate-policies.sh <tenant-id>
   ```

3. **Implementar particionamento de políticas:**
   ```sql
   ALTER TABLE iam_schema.policies PARTITION BY LIST (tenant_id);
   ```

#### 2.5 Verificação de Resolução
1. **Monitorar tempos de resposta de autorização por 30 minutos**
2. **Verificar utilização de CPU e memória sob carga**
3. **Testar autorização em diferentes cenários**
4. **Confirmar eficácia da solução em diferentes tenants**

#### 2.6 Ações Pós-Incidente
1. **Implementar monitoramento específico para tempo de decisão**
2. **Estabelecer limites de alerta para decisões lentas**
3. **Planejar revisões periódicas de políticas**
4. **Documentar otimizações realizadas**

### 3. Problemas de Escalação de Privilégios

#### 3.1 Sintomas
- Usuários com acesso a recursos ou ações além de suas permissões esperadas
- Alertas de segurança sobre acessos anômalos
- Padrões suspeitos de uso de permissões

#### 3.2 Verificações Iniciais
1. **Verificar logs de atribuição de roles:**
   ```bash
   kubectl logs -n iam-namespace <rbac-service-pod-name> --tail=1000 | grep -i "role\|grant\|assign"
   ```

2. **Verificar políticas recentemente modificadas:**
   ```sql
   SELECT policy_id, name, created_at, updated_at, created_by, updated_by 
   FROM iam_schema.policies 
   WHERE updated_at > NOW() - INTERVAL '7 DAYS'
   ORDER BY updated_at DESC;
   ```

3. **Verificar operações administrativas recentes:**
   ```sql
   SELECT * FROM iam_schema.admin_audit_log 
   WHERE action_type IN ('CREATE_POLICY', 'UPDATE_POLICY', 'GRANT_ROLE') 
   AND action_time > NOW() - INTERVAL '7 DAYS'
   ORDER BY action_time DESC;
   ```

4. **Verificar alterações em roles administrativas:**
   ```sql
   SELECT ur.user_id, u.username, r.role_name, ur.granted_at, ur.granted_by
   FROM iam_schema.user_roles ur
   JOIN iam_schema.users u ON ur.user_id = u.user_id
   JOIN iam_schema.roles r ON ur.role_id = r.role_id
   WHERE r.role_name LIKE '%admin%' AND ur.granted_at > NOW() - INTERVAL '30 DAYS'
   ORDER BY ur.granted_at DESC;
   ```

#### 3.3 Diagnóstico Avançado
1. **Analisar caminhos de escalação potenciais:**
   ```bash
   kubectl exec -it -n iam-namespace <security-pod-name> -- /app/scripts/analyze-privilege-paths.sh <suspicious-user-id>
   ```

2. **Verificar conflitos em políticas de SoD:**
   ```sql
   SELECT * FROM iam_schema.sod_conflicts 
   WHERE detection_time > NOW() - INTERVAL '7 DAYS';
   ```

3. **Revisão completa de permissões para usuário suspeito:**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/full-permission-report.sh <suspicious-user-id>
   ```

4. **Verificar tentativas de ataques conhecidos:**
   ```bash
   kubectl logs -n iam-namespace <security-monitoring-pod-name> --tail=5000 | grep -i "attack\|exploit\|privilege\|escalation"
   ```

#### 3.4 Ações Corretivas

**Nível 1 (Contenção Imediata):**
1. **Suspender usuário suspeito:**
   ```sql
   UPDATE iam_schema.users 
   SET status = 'SUSPENDED', 
       suspension_reason = 'security_investigation', 
       suspended_at = NOW(), 
       suspended_by = 'security_response_team'
   WHERE user_id = '<suspicious-user-id>';
   ```

2. **Remover permissões críticas do usuário:**
   ```sql
   DELETE FROM iam_schema.user_roles 
   WHERE user_id = '<suspicious-user-id>' 
   AND role_id IN (SELECT role_id FROM iam_schema.roles WHERE is_administrative = true);
   ```

3. **Verificar e revogar tokens ativos:**
   ```sql
   UPDATE iam_schema.access_tokens 
   SET revoked = true, 
       revocation_reason = 'security_investigation' 
   WHERE user_id = '<suspicious-user-id>' AND revoked = false;
   ```

**Nível 2 (Resolução):**
1. **Corrigir políticas com configuração inadequada:**
   ```sql
   UPDATE iam_schema.policies 
   SET effect = 'deny', 
       updated_by = 'security_response_team', 
       updated_at = NOW()
   WHERE policy_id = '<vulnerable-policy-id>';
   ```

2. **Implementar políticas de segurança mais restritivas:**
   ```bash
   kubectl apply -f stricter-security-policies.yaml -n iam-namespace
   ```

3. **Reforçar verificações de SoD:**
   ```sql
   INSERT INTO iam_schema.sod_policy (name, description, conflicting_roles, detection_action)
   VALUES ('restrict_sensitive_data_access', 'Prevent data modification and approval by same user', 
           ARRAY['data_modifier', 'approval_officer'], 'prevent_and_alert');
   ```

#### 3.5 Verificação de Resolução
1. **Executar varredura de segurança completa**
2. **Verificar logs por novos padrões suspeitos**
3. **Validar se permissões excessivas foram removidas**
4. **Testar novamente caminhos potenciais de escalação**

#### 3.6 Ações Pós-Incidente
1. **Realizar análise forense completa**
2. **Documentar vulnerabilidade e correção**
3. **Implementar verificações adicionais de segurança**
4. **Revisar e atualizar políticas de separação de funções**
5. **Conduzir auditoria completa de permissões**

### 4. Problemas em Roles e Delegações

#### 4.1 Sintomas
- Administradores não conseguem delegar permissões
- Roles atribuídas não são propagadas para usuários
- Inconsistências entre permissões exibidas e aplicadas

#### 4.2 Verificações Iniciais
1. **Verificar logs de delegação:**
   ```bash
   kubectl logs -n iam-namespace <rbac-service-pod-name> --tail=500 | grep -i "delegat\|assign\|grant"
   ```

2. **Verificar configurações de limites de delegação:**
   ```bash
   kubectl get configmap -n iam-namespace rbac-config -o yaml | grep -A 10 delegation
   ```

3. **Verificar permissões de delegação do usuário:**
   ```sql
   SELECT p.policy_id, p.name, p.resource_pattern, p.action_pattern
   FROM iam_schema.policies p 
   JOIN iam_schema.role_policies rp ON p.policy_id = rp.policy_id 
   JOIN iam_schema.user_roles ur ON rp.role_id = ur.role_id 
   WHERE ur.user_id = '<delegator-user-id>'
   AND (p.resource_pattern LIKE '%:role:%' AND p.action_pattern LIKE '%:grant%');
   ```

4. **Verificar eventos recentes de delegação:**
   ```sql
   SELECT * FROM iam_schema.delegation_history
   WHERE delegator_id = '<delegator-user-id>'
   ORDER BY delegation_time DESC LIMIT 20;
   ```

#### 4.3 Diagnóstico Avançado
1. **Analisar hierarquia completa de roles:**
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

2. **Verificar consistência de delegações:**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/validate-delegations.sh <tenant-id>
   ```

3. **Verificar limitações de delegação transversal:**
   ```sql
   SELECT dg.* 
   FROM iam_schema.delegation_graph dg
   JOIN iam_schema.roles r1 ON dg.source_role_id = r1.role_id
   JOIN iam_schema.roles r2 ON dg.target_role_id = r2.role_id
   WHERE r1.organizational_unit_id != r2.organizational_unit_id;
   ```

4. **Verificar propagação de roles secundárias:**
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

#### 4.4 Ações Corretivas

**Nível 1 (Mitigação Rápida):**
1. **Reiniciar serviço de RBAC:**
   ```bash
   kubectl rollout restart deployment -n iam-namespace rbac-service
   ```

2. **Limpar cache de roles:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli DEL "roles:user:<problem-user-id>"
   ```

3. **Forçar recálculo de delegações:**
   ```bash
   kubectl exec -it -n iam-namespace <rbac-service-pod-name> -- /app/scripts/recalculate-delegations.sh <tenant-id>
   ```

**Nível 2 (Resolução):**
1. **Corrigir configurações de delegação:**
   ```bash
   kubectl apply -f corrected-delegation-config.yaml -n iam-namespace
   ```

2. **Corrigir propagação de roles:**
   ```sql
   -- Adicionar implicação de role ausente
   INSERT INTO iam_schema.role_implications (role_id, implied_role_id, grant_type)
   VALUES ('<parent-role-id>', '<implied-role-id>', 'SYSTEM_DEFINED');
   ```

3. **Corrigir limite de delegação:**
   ```sql
   UPDATE iam_schema.delegation_policies
   SET max_depth = 3, allow_cross_department = false
   WHERE policy_id = '<restrictive-policy-id>';
   ```

#### 4.5 Verificação de Resolução
1. **Testar processo de delegação**
2. **Verificar propagação de permissões**
3. **Validar hierarquia de delegação**
4. **Testar cenários de delegação em vários níveis**

#### 4.6 Ações Pós-Incidente
1. **Documentar problemas e soluções de delegação**
2. **Revisar políticas de delegação**
3. **Melhorar mecanismos de validação**
4. **Atualizar documentação de administração**

## Recursos Adicionais

### Ferramentas de Diagnóstico

1. **Scripts de Diagnóstico:**
   - `/opt/innovabiz/iam/scripts/policy-analyzer.sh`
   - `/opt/innovabiz/iam/scripts/permission-checker.sh`
   - `/opt/innovabiz/iam/scripts/delegation-validator.sh`

2. **Dashboards de Monitoramento:**
   - Grafana RBAC: `https://grafana.innovabiz.com/d/rbac-overview`
   - Prometheus: `https://prometheus.innovabiz.com/graph?g0.expr=rbac_decision_duration_seconds`

3. **Consultas Úteis para o Banco de Dados:**
   ```sql
   -- Verificar conflitos potenciais de políticas (allow e deny na mesma ação/recurso)
   SELECT r.resource_pattern, a.action_pattern, 
          COUNT(CASE WHEN p.effect = 'allow' THEN 1 END) as allow_count,
          COUNT(CASE WHEN p.effect = 'deny' THEN 1 END) as deny_count
   FROM iam_schema.policies p,
        LATERAL (SELECT UNNEST(STRING_TO_ARRAY(p.resource_pattern, ',')) as resource_pattern) r,
        LATERAL (SELECT UNNEST(STRING_TO_ARRAY(p.action_pattern, ',')) as action_pattern) a
   GROUP BY r.resource_pattern, a.action_pattern
   HAVING COUNT(CASE WHEN p.effect = 'allow' THEN 1 END) > 0 
      AND COUNT(CASE WHEN p.effect = 'deny' THEN 1 END) > 0;
   
   -- Encontrar políticas que afetam um recurso específico
   SELECT p.*
   FROM iam_schema.policies p
   WHERE resource_pattern LIKE '%<resource-pattern>%'
   OR resource_pattern = '*';
   
   -- Verificar histórico de alterações em políticas críticas
   SELECT * FROM iam_schema.policy_change_history
   WHERE policy_id IN (
     SELECT policy_id FROM iam_schema.policies
     WHERE resource_pattern LIKE '%:admin:%' OR resource_pattern LIKE '%:critical:%'
   )
   ORDER BY change_time DESC;
   ```

### Referências

- [Modelo de Segurança IAM](../05-Seguranca/Modelo_Seguranca_IAM.md)
- [Arquitetura Técnica IAM](../02-Arquitetura/Arquitetura_Tecnica_IAM.md)
- [Framework de Compliance IAM](../10-Governanca/Framework_Compliance_IAM.md)
- [Guia Operacional IAM](../08-Operacoes/Guia_Operacional_IAM.md)
- [Documentação de APIs IAM](../03-Desenvolvimento/Documentacao_API_IAM.md)

### Contatos para Escalação

| Nível | Equipe | Contato | Acionamento |
|-------|--------|---------|------------|
| 1 | Suporte IAM | iam-support@innovabiz.com | Problemas iniciais |
| 2 | Operações IAM | iam-ops@innovabiz.com | Após 30 min sem resolução L1 |
| 3 | Desenvolvimento IAM | iam-dev@innovabiz.com | Após 60 min sem resolução L2 |
| 4 | Segurança de Dados | data-security@innovabiz.com | Problemas críticos de autorização |
| 5 | CISO | ciso@innovabiz.com | Incidentes de segurança confirmados |
