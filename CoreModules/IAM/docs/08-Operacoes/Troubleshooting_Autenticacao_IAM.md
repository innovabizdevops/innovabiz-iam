# Procedimentos de Troubleshooting de Autenticação IAM

## Introdução

Este documento oferece um guia detalhado para diagnóstico e resolução de problemas relacionados a funcionalidades de autenticação do módulo IAM da plataforma INNOVABIZ. É destinado aos times de operações, administradores de IAM e profissionais de suporte técnico.

## Matriz de Problemas Comuns

| Sintoma | Possível Causa | Gravidade | Impacto | Tempo Médio de Resolução |
|---------|----------------|-----------|---------|--------------------------|
| Falhas massivas de login | Serviço de autenticação indisponível | Crítica | Usuários não conseguem acessar o sistema | 30-60 minutos |
| MFA não funcionando | Problema com provedor MFA ou falha de sincronização | Alta | Usuários com MFA não conseguem autenticar | 15-45 minutos |
| Expiração prematura de tokens | Configuração incorreta, sincronização NTP, chaves corrompidas | Alta | Desconexões frequentes de usuários | 30-60 minutos |
| Erro em SSO | Problema com provedor de identidade externo ou configuração SAML/OIDC | Alta | Usuários federados não conseguem acessar | 45-90 minutos |
| Falhas intermitentes de autenticação | Sobrecarga de recursos ou problemas de conectividade | Média | Experiência de usuário degradada | 30-60 minutos |
| Autenticação lenta | Problemas de performance em banco de dados ou cache | Média | Atrasos no processo de login | 30-90 minutos |

## Procedimentos de Troubleshooting

### 1. Falhas Massivas de Login

#### 1.1 Sintomas
- Múltiplos relatos de falha de login
- Aumento abrupto em erros 401/403 nas APIs de autenticação
- Alertas de disponibilidade do serviço de autenticação

#### 1.2 Verificações Iniciais
1. **Verificar status do serviço de autenticação:**
   ```bash
   kubectl get pods -n iam-namespace | grep auth-service
   kubectl logs -n iam-namespace <auth-service-pod-name> --tail=100
   ```

2. **Verificar métricas de disponibilidade:**
   - Acessar dashboard Grafana de disponibilidade IAM
   - Verificar falhas de conexão com dependências

3. **Verificar conectividade com o banco de dados:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-service-pod-name> -- pg_isready -h <db-host> -p <db-port>
   ```

4. **Verificar status do Redis (cache de sessão):**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli ping
   ```

#### 1.3 Diagnóstico Avançado
1. **Análise detalhada de logs:**
   ```bash
   kubectl logs -n iam-namespace <auth-service-pod-name> --tail=500 | grep ERROR
   ```

2. **Verificar problemas de configuração:**
   ```bash
   kubectl describe configmap -n iam-namespace auth-service-config
   kubectl get secret -n iam-namespace auth-service-secrets -o yaml
   ```

3. **Analisar métricas de performance:**
   - CPU, memória e latência do serviço de autenticação
   - Tempo de resposta do banco de dados

4. **Verificar alterações recentes:**
   - Revisão de deploys recentes
   - Mudanças de configuração ou políticas

#### 1.4 Ações Corretivas

**Nível 1 (Mitigação Rápida):**
1. **Reiniciar o serviço de autenticação:**
   ```bash
   kubectl rollout restart deployment -n iam-namespace auth-service
   ```

2. **Limpar cache de configuração:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli FLUSHDB
   ```

3. **Ativar modo de contingência** (se disponível):
   ```bash
   kubectl patch deployment -n iam-namespace auth-service -p '{"spec":{"template":{"spec":{"containers":[{"name":"auth-service","env":[{"name":"CONTINGENCY_MODE","value":"true"}]}]}}}}'
   ```

**Nível 2 (Resolução):**
1. **Reverter para versão anterior estável:**
   ```bash
   kubectl rollout undo deployment -n iam-namespace auth-service
   ```

2. **Verificar e corrigir problemas de banco de dados:**
   - Analisar locks ou conexões pendentes
   - Otimizar consultas lentas

3. **Aplicar correções de configuração:**
   - Atualizar parâmetros de conexão
   - Ajustar timeouts e limites de recursos

#### 1.5 Verificação de Resolução
1. **Testar autenticação de usuários de teste em diferentes tenants**
2. **Monitorar taxas de sucesso de autenticação por 15 minutos**
3. **Verificar logs em busca de erros recorrentes**
4. **Confirmar métricas de performance normalizadas**

#### 1.6 Ações Pós-Incidente
1. **Realizar análise de causa raiz (RCA)**
2. **Documentar lições aprendidas**
3. **Implementar medidas preventivas**
4. **Atualizar procedimentos, se necessário**

### 2. Problemas com Autenticação Multifator (MFA)

#### 2.1 Sintomas
- Usuários relatam incapacidade de completar autenticação com MFA
- Códigos MFA sendo rejeitados
- Falha ao gerar novos tokens MFA

#### 2.2 Verificações Iniciais
1. **Verificar logs específicos de MFA:**
   ```bash
   kubectl logs -n iam-namespace <auth-service-pod-name> | grep -i "mfa\|totp\|authenticator"
   ```

2. **Verificar status do provedor MFA:**
   - Se integração externa: verificar status do serviço
   - Se interno: verificar status do módulo de MFA

3. **Verificar sincronização de tempo:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-service-pod-name> -- date
   ```

4. **Realizar tentativa de login com conta de teste com MFA:**
   - Testar em diferentes navegadores/dispositivos
   - Testar métodos alternativos (SMS, e-mail, TOTP)

#### 2.3 Diagnóstico Avançado
1. **Analisar configurações de MFA:**
   ```bash
   kubectl get configmap -n iam-namespace mfa-config -o yaml
   ```

2. **Verificar chaves secretas e certificados:**
   ```bash
   kubectl get secret -n iam-namespace mfa-secrets -o yaml
   ```

3. **Rastrear fluxo completo de autenticação:**
   - Analisar logs de transações específicas
   - Verificar mensagens de erro detalhadas

4. **Verificar problemas no banco de dados:**
   ```sql
   SELECT * FROM iam_schema.mfa_devices WHERE last_error IS NOT NULL ORDER BY updated_at DESC LIMIT 10;
   SELECT COUNT(*) FROM iam_schema.mfa_failed_attempts WHERE attempt_time > NOW() - INTERVAL '1 HOUR' GROUP BY user_id;
   ```

#### 2.4 Ações Corretivas

**Nível 1 (Mitigação Rápida):**
1. **Reiniciar serviço de MFA:**
   ```bash
   kubectl rollout restart deployment -n iam-namespace mfa-service
   ```

2. **Sincronizar relógios do sistema:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-pod-name> -- ntpd -gq
   ```

3. **Limpar cache de dispositivos MFA:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli DEL mfa:devicecache:*
   ```

**Nível 2 (Resolução):**
1. **Revogar e solicitar reinscrição de dispositivos problemáticos:**
   ```sql
   UPDATE iam_schema.mfa_devices SET status = 'REVOKED' WHERE device_id IN (SELECT device_id FROM iam_schema.mfa_failed_attempts GROUP BY device_id HAVING COUNT(*) > 5);
   ```

2. **Atualizar configurações de MFA:**
   ```bash
   kubectl apply -f updated-mfa-config.yaml
   ```

3. **Renovar certificados ou chaves de segurança:**
   ```bash
   kubectl create secret generic mfa-secrets --from-file=./new-mfa-keys/ -n iam-namespace
   ```

#### 2.5 Verificação de Resolução
1. **Testar processo completo de MFA com contas de teste**
2. **Verificar se novos códigos estão sendo aceitos**
3. **Monitorar taxas de sucesso de verificação MFA**
4. **Verificar logs em busca de novos erros**

#### 2.6 Ações Pós-Incidente
1. **Notificar usuários sobre resolução**
2. **Documentar causa e solução**
3. **Avaliar implementação de métodos alternativos de MFA**
4. **Atualizar guias de usuário, se necessário**

### 3. Problemas de Expiração Prematura de Tokens

#### 3.1 Sintomas
- Usuários relatam desconexões frequentes
- Tokens JWT expiram antes do tempo configurado
- Sessões terminando inesperadamente

#### 3.2 Verificações Iniciais
1. **Verificar configurações de expiração de token:**
   ```bash
   kubectl get configmap -n iam-namespace auth-config -o yaml | grep -i "expir\|timeout\|duration"
   ```

2. **Verificar logs de validação de token:**
   ```bash
   kubectl logs -n iam-namespace <auth-service-pod-name> | grep -i "token\|jwt\|expir\|valid"
   ```

3. **Verificar sincronização de relógios:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-pod-name> -- date
   kubectl exec -it -n iam-namespace <api-gateway-pod> -- date
   ```

4. **Verificar status de chaves de assinatura JWT:**
   ```bash
   kubectl get secret -n iam-namespace jwt-keys -o yaml
   ```

#### 3.3 Diagnóstico Avançado
1. **Decodificar e analisar tokens problemáticos:**
   - Usar ferramentas como jwt.io para analisar tokens
   - Verificar claims específicos (exp, iat, nbf)

2. **Analisar armazenamento de sessão:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli --scan --pattern "session:*" | head -n 10
   ```

3. **Verificar comportamento em diferentes tenants:**
   - Testar em múltiplos tenants
   - Verificar configurações específicas por tenant

4. **Analisar integridade das chaves de assinatura:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-pod-name> -- openssl rsa -check -in /path/to/private.key
   ```

#### 3.4 Ações Corretivas

**Nível 1 (Mitigação Rápida):**
1. **Aumentar temporariamente tempo de expiração:**
   ```bash
   kubectl patch configmap -n iam-namespace auth-config --type merge -p '{"data":{"JWT_EXPIRATION_SECONDS":"7200"}}'
   ```

2. **Reiniciar serviço de autenticação:**
   ```bash
   kubectl rollout restart deployment -n iam-namespace auth-service
   ```

3. **Sincronizar relógios do sistema:**
   ```bash
   kubectl exec -it -n iam-namespace <auth-pod-name> -- ntpd -gq
   ```

**Nível 2 (Resolução):**
1. **Rotacionar chaves de assinatura JWT:**
   ```bash
   # Gerar novas chaves
   openssl genrsa -out private_key.pem 4096
   openssl rsa -in private_key.pem -pubout -out public_key.pem
   
   # Atualizar secrets
   kubectl create secret generic jwt-keys --from-file=./private_key.pem --from-file=./public_key.pem -n iam-namespace --dry-run=client -o yaml | kubectl apply -f -
   ```

2. **Ajustar configurações de token:**
   ```bash
   kubectl apply -f updated-token-config.yaml
   ```

3. **Implementar rotação gradual de tokens:**
   ```bash
   kubectl patch deployment -n iam-namespace auth-service -p '{"spec":{"template":{"spec":{"containers":[{"name":"auth-service","env":[{"name":"GRADUAL_TOKEN_ROTATION","value":"true"}]}]}}}}'
   ```

#### 3.5 Verificação de Resolução
1. **Monitorar duração de sessões ativas**
2. **Verificar logs em busca de erros de validação de token**
3. **Testar renovação de tokens com contas de teste**
4. **Verificar comportamento em vários dispositivos/navegadores**

#### 3.6 Ações Pós-Incidente
1. **Implementar monitoramento melhorado para expiração de tokens**
2. **Revisar estratégia de rotação de chaves**
3. **Documentar comportamento esperado de expiração**
4. **Atualizar procedimentos operacionais**

## Problemas Relacionados a SSO (Single Sign-On)

### 4.1 Sintomas
- Falhas na autenticação via provedores externos (Google, Microsoft, etc.)
- Erros durante redirecionamento para/de provedores de identidade
- Metadados SAML/OIDC incorretos ou expirados

### 4.2 Verificações Iniciais
1. **Verificar logs específicos de SSO:**
   ```bash
   kubectl logs -n iam-namespace <auth-service-pod-name> | grep -i "sso\|saml\|oidc\|oauth"
   ```

2. **Verificar status dos endpoints de SSO:**
   ```bash
   curl -I https://<auth-domain>/sso/health
   ```

3. **Verificar metadados SAML/OIDC:**
   ```bash
   curl https://<auth-domain>/sso/metadata
   ```

4. **Verificar certificados de SSO:**
   ```bash
   kubectl get secret -n iam-namespace sso-certificates -o yaml
   ```

### 4.3 Diagnóstico Avançado
1. **Analisar tráfego de rede completo:**
   - Capturar e analisar fluxo completo de autenticação
   - Verificar cabeçalhos e parâmetros de solicitação/resposta

2. **Verificar configurações de provedores:**
   ```bash
   kubectl get configmap -n iam-namespace sso-providers-config -o yaml
   ```

3. **Testar com ferramentas específicas:**
   ```bash
   saml2aws validate --provider=<provider>
   oidc-cli validate --issuer=<issuer-url>
   ```

4. **Verificar status do provedor externo:**
   - Verificar status pages dos provedores
   - Testar outros aplicativos com mesmo provedor

### 4.4 Ações Corretivas

**Nível 1 (Mitigação Rápida):**
1. **Reiniciar serviço de SSO:**
   ```bash
   kubectl rollout restart deployment -n iam-namespace sso-service
   ```

2. **Refrescar metadados de federação:**
   ```bash
   kubectl exec -it -n iam-namespace <sso-pod-name> -- /app/scripts/refresh-metadata.sh
   ```

3. **Limpar cache de metadados:**
   ```bash
   kubectl exec -it -n iam-namespace <redis-pod-name> -- redis-cli DEL sso:metadata:*
   ```

**Nível 2 (Resolução):**
1. **Renovar certificados de SSO:**
   ```bash
   # Gerar novos certificados
   openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes
   
   # Atualizar secrets
   kubectl create secret generic sso-certificates --from-file=./cert.pem --from-file=./key.pem -n iam-namespace --dry-run=client -o yaml | kubectl apply -f -
   ```

2. **Atualizar configurações de provedores:**
   ```bash
   kubectl apply -f updated-sso-config.yaml
   ```

3. **Reconfigurar mapeamentos de atributos:**
   ```bash
   kubectl apply -f updated-attribute-mapping.yaml
   ```

### 4.5 Verificação de Resolução
1. **Testar fluxo de SSO com contas de teste**
2. **Verificar mapeamento correto de atributos**
3. **Testar autenticação com diferentes provedores**
4. **Verificar logs em busca de erros de SSO**

### 4.6 Ações Pós-Incidente
1. **Documentar configurações corretas**
2. **Implementar monitoramento específico para SSO**
3. **Revisar e atualizar documentação de integração**
4. **Configurar alertas para expiração de certificados**

## Recursos Adicionais

### Ferramentas de Diagnóstico

1. **Scripts de Diagnóstico:**
   - `/opt/innovabiz/iam/scripts/auth-diagnostics.sh`
   - `/opt/innovabiz/iam/scripts/token-validator.sh`
   - `/opt/innovabiz/iam/scripts/mfa-tester.sh`

2. **Dashboards de Monitoramento:**
   - Grafana IAM: `https://grafana.innovabiz.com/d/iam-overview`
   - Prometheus: `https://prometheus.innovabiz.com/graph`

3. **Consultas Úteis para o Banco de Dados:**
   ```sql
   -- Verificar tentativas recentes de login por status
   SELECT status, COUNT(*) FROM iam_schema.auth_attempts 
   WHERE attempt_time > NOW() - INTERVAL '1 HOUR' 
   GROUP BY status;
   
   -- Verificar dispositivos MFA com problemas
   SELECT user_id, device_id, COUNT(*) as failure_count
   FROM iam_schema.mfa_failed_attempts 
   WHERE attempt_time > NOW() - INTERVAL '24 HOURS'
   GROUP BY user_id, device_id 
   HAVING COUNT(*) > 3;
   
   -- Verificar tokens revogados recentemente
   SELECT token_id, user_id, revocation_reason, revoked_at
   FROM iam_schema.revoked_tokens
   WHERE revoked_at > NOW() - INTERVAL '24 HOURS'
   ORDER BY revoked_at DESC;
   ```

### Referências

- [Modelo de Segurança IAM](../05-Seguranca/Modelo_Seguranca_IAM.md)
- [Requisitos de Infraestrutura IAM](../04-Infraestrutura/Requisitos_Infraestrutura_IAM.md)
- [Arquitetura Técnica IAM](../02-Arquitetura/Arquitetura_Tecnica_IAM.md)
- [Framework de Compliance IAM](../10-Governanca/Framework_Compliance_IAM.md)
- [Guia Operacional IAM](../08-Operacoes/Guia_Operacional_IAM.md)

### Contatos para Escalação

| Nível | Equipe | Contato | Acionamento |
|-------|--------|---------|------------|
| 1 | Suporte IAM | iam-support@innovabiz.com | Problemas iniciais |
| 2 | Operações IAM | iam-ops@innovabiz.com | Após 30 min sem resolução L1 |
| 3 | Desenvolvimento IAM | iam-dev@innovabiz.com | Após 60 min sem resolução L2 |
| 4 | Arquitetura de Segurança | security-arch@innovabiz.com | Problemas críticos de segurança |
