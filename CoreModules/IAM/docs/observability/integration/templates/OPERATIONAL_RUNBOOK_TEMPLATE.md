# Template para Runbooks Operacionais - INNOVABIZ

## Visão Geral

Este documento fornece o template padrão INNOVABIZ para a criação de runbooks operacionais de observabilidade. Runbooks são guias de procedimentos padronizados para diagnosticar e resolver problemas identificados pelo sistema de monitoramento e alertas. Estes runbooks seguem os princípios multi-dimensionais da plataforma INNOVABIZ, considerando o contexto de tenant, região e módulo.

## Estrutura do Runbook

Os runbooks operacionais da INNOVABIZ seguem uma estrutura padronizada para facilitar o uso em situações de pressão e garantir uma resolução consistente e eficiente dos problemas:

```markdown
# [ID-RUNBOOK] - [Nome do Problema]

## 🚨 Visão Geral

**Módulo:** [Nome do Módulo]  
**Serviço:** [Nome do Serviço]  
**Severidade:** [Crítica/Alta/Média/Baixa]  
**SLA para Resolução:** [Tempo esperado]  
**Equipe Responsável:** [Nome da Equipe]  

## 🔍 Detecção

**Alerta(s) Relacionado(s):**
- [Nome do Alerta 1]
- [Nome do Alerta 2]

**Dashboards de Monitoramento:**
- [Link para Dashboard Principal]
- [Link para Dashboard Detalhado]

**Sintomas Visíveis:**
- [Sintoma 1 - ex: Aumento de latência nas APIs]
- [Sintoma 2 - ex: Erro nos logs com padrão X]
- [Sintoma 3 - ex: Falhas nos health checks]

## 📈 Impacto

**Impacto no Negócio:**
- [Descrição do impacto para o usuário final]
- [Impacto financeiro, se aplicável]
- [Outros sistemas afetados]

**Escopo Multi-dimensional:**
- **Tenants Afetados:** [Todos/Específicos - quais?]
- **Regiões Afetadas:** [Todas/Específicas - quais?]
- **Ambientes:** [Produção/Homologação/Desenvolvimento]

## 🔎 Diagnóstico

### 1. Verificações Iniciais

```bash
# Comandos para verificações iniciais, ex:
kubectl get pods -n {{namespace}}
kubectl logs -f {{pod-name}} -n {{namespace}} | grep ERROR
curl -v https://{{service-endpoint}}/health
```

### 2. Análise de Logs

- Verificar logs de erro em: [caminho/link para visualizar logs]
- Padrões específicos para procurar:
  ```
  ERROR: Connection refused
  FATAL: Database unavailable
  WARN: High memory usage detected
  ```

### 3. Análise de Métricas

- Verificar dashboard [link] para análise de:
  - [Métrica 1] - Valores normais: X-Y, Valores de alerta: > Z
  - [Métrica 2] - Valores normais: X-Y, Valores de alerta: > Z
  - [Correlação entre métricas A e B]

### 4. Rastreamento de Problemas

- Verificar traces relacionados no Jaeger: [link]
- Filtrar por:
  - Tenant: {{tenant_id}}
  - Região: {{region_id}}
  - Tags de erro: error=true

## 🛠️ Resolução

### Cenário 1: [Causa Raiz Comum 1]

**Diagnóstico Detalhado:**
- [Como identificar especificamente este cenário]
- [Evidências que confirmam esta causa raiz]

**Passos de Resolução:**

1. [Passo detalhado 1]
   ```bash
   # Comando específico se aplicável
   kubectl scale deployment {{deployment-name}} --replicas=3 -n {{namespace}}
   ```

2. [Passo detalhado 2]
   ```bash
   # Comando específico se aplicável
   kubectl apply -f updated-config.yaml -n {{namespace}}
   ```

3. [Passo detalhado 3]
   - [Subpassos ou detalhamentos]
   - [Subpassos ou detalhamentos]

**Verificação de Resolução:**
- [Como confirmar que o problema foi resolvido]
- [Métricas que devem normalizar]
- [Tempo esperado para normalização]

### Cenário 2: [Causa Raiz Comum 2]

[... Repetir estrutura do Cenário 1 ...]

### Cenário 3: [Causa Raiz Comum 3]

[... Repetir estrutura do Cenário 1 ...]

## 🔄 Procedimento de Escalação

**Nível 1 - SRE de Plantão:**
- Nome: [Nome do Contato]
- Contato: [E-mail e/ou telefone]
- Horário: [Disponibilidade]

**Nível 2 - Especialista do Módulo:**
- Nome: [Nome do Contato]
- Contato: [E-mail e/ou telefone]
- Horário: [Disponibilidade]

**Nível 3 - Gerência Técnica:**
- Nome: [Nome do Contato]
- Contato: [E-mail e/ou telefone]
- Horário: [Disponibilidade]

## 🛡️ Prevenção

**Melhorias Identificadas:**
- [Melhoria 1 - ex: Aumentar timeouts de conexão]
- [Melhoria 2 - ex: Adicionar retry policies]
- [Melhoria 3 - ex: Implementar circuit breaker]

**Alertas Preventivos Sugeridos:**
- [Sugestão de novos alertas ou modificações em alertas existentes]

## 📚 Referências

- [Link para documentação técnica relacionada]
- [Link para incidentes passados similares]
- [Link para base de conhecimento]
- [Link para documentação de arquitetura]

## 📝 Histórico de Incidentes

| Data | ID Incidente | Resumo | Tenant | Região | Tempo de Resolução | Notas |
|------|-------------|--------|--------|--------|-------------------|-------|
| YYYY-MM-DD | INC-XXXXX | Breve descrição | ID Tenant | ID Região | XX minutos | Observações importantes |

## 🔄 Histórico de Revisões do Runbook

| Data | Versão | Autor | Mudanças |
|------|--------|-------|----------|
| YYYY-MM-DD | 1.0 | Nome do Autor | Versão inicial |
| YYYY-MM-DD | 1.1 | Nome do Autor | Atualizações baseadas no incidente INC-XXXXX |
```

## Exemplo de Runbook Preenchido

Abaixo está um exemplo de runbook preenchido para um cenário comum:

```markdown
# RB-IAM-001 - Latência Elevada na API de Autenticação

## 🚨 Visão Geral

**Módulo:** IAM  
**Serviço:** Authentication API  
**Severidade:** Alta  
**SLA para Resolução:** 30 minutos  
**Equipe Responsável:** Equipe de Identidade e Acesso  

## 🔍 Detecção

**Alerta(s) Relacionado(s):**
- HighLatencyAuthAPI
- HighErrorRateAuthAPI
- HighDatabaseLatency

**Dashboards de Monitoramento:**
- [IAM Operational Dashboard](https://grafana.innovabiz.com/d/iam-operational)
- [IAM API Performance](https://grafana.innovabiz.com/d/iam-api-performance)

**Sintomas Visíveis:**
- Latência p95 das APIs de autenticação > 1000ms
- Aumento de timeouts em serviços dependentes
- Aumento de erros 503 nas respostas da API
- Logs indicando lentidão nas consultas do banco de dados

## 📈 Impacto

**Impacto no Negócio:**
- Atraso nos processos de login de usuários
- Possíveis falhas em serviços dependentes que usam autenticação
- Degradação na experiência do usuário
- Potencial perda de transações financeiras se a autenticação falhar

**Escopo Multi-dimensional:**
- **Tenants Afetados:** Potencialmente todos, verificar métricas por tenant
- **Regiões Afetadas:** Normalmente localizado em uma região específica
- **Ambientes:** Produção

## 🔎 Diagnóstico

### 1. Verificações Iniciais

```bash
# Verificar status dos pods
kubectl get pods -n iam

# Verificar logs de erro recentes
kubectl logs -l app=auth-api -n iam --tail=100 | grep ERROR

# Verificar métricas de utilização de recursos
kubectl top pods -l app=auth-api -n iam

# Verificar status do health check
curl -v https://auth.api.innovabiz.com/health
```

### 2. Análise de Logs

- Verificar logs de erro no Kibana: [Dashboard IAM Logs](https://kibana.innovabiz.com/app/dashboards#/view/iam-errors)
- Padrões específicos para procurar:
  ```
  ERROR: Connection timeout to database
  WARN: Slow query detected (execution time > 1000ms)
  ERROR: Rate limit exceeded on Redis cache
  ```

### 3. Análise de Métricas

- Verificar dashboard [IAM Database Performance](https://grafana.innovabiz.com/d/iam-database) para análise de:
  - Tempo de resposta de consultas: Normal: 10-50ms, Alerta: > 200ms
  - Utilização de conexões: Normal: 30-60%, Alerta: > 80%
  - Cache hit rate: Normal: > 85%, Alerta: < 70%

### 4. Rastreamento de Problemas

- Verificar traces lentos no Jaeger: [IAM Traces](https://jaeger.innovabiz.com/search?service=auth-api)
- Filtrar por:
  - Duração > 500ms
  - Tags: error=true, db.type=postgresql

## 🛠️ Resolução

### Cenário 1: Sobrecarga de Conexões no Banco de Dados

**Diagnóstico Detalhado:**
- Alta contagem de conexões abertas no PostgreSQL (> 80% do máximo)
- Queries em estado "idle in transaction" ou acumuladas
- Logs indicando timeouts de conexão

**Passos de Resolução:**

1. Verificar e limpar conexões ociosas
   ```bash
   kubectl exec -it iam-db-0 -n iam -- psql -U postgres -c "SELECT count(*) FROM pg_stat_activity WHERE state = 'idle in transaction';"
   kubectl exec -it iam-db-0 -n iam -- psql -U postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state = 'idle in transaction' AND (now() - state_change) > '5 minutes'::interval;"
   ```

2. Reiniciar pods de API de forma escalonada
   ```bash
   # Reiniciar pods de forma escalonada (um por vez)
   kubectl rollout restart deployment/auth-api -n iam
   ```

3. Aumentar temporariamente o pool de conexões se necessário
   ```bash
   # Editar o ConfigMap com os parâmetros do pool de conexões
   kubectl edit configmap auth-api-config -n iam
   # Modificar MAX_POOL_SIZE para um valor 20% maior que o atual
   ```

**Verificação de Resolução:**
- Latência das APIs deve normalizar em 2-5 minutos
- Contagem de conexões no banco deve reduzir para < 60% do máximo
- Erros 503 devem parar de ocorrer

### Cenário 2: Cache Redis Degradado

**Diagnóstico Detalhado:**
- Cache hit rate < 70%
- Latência elevada em operações de cache (> 50ms)
- Erros de timeout em operações Redis

**Passos de Resolução:**

1. Verificar status do cluster Redis
   ```bash
   kubectl exec -it redis-master-0 -n iam -- redis-cli info | grep connected_clients
   kubectl exec -it redis-master-0 -n iam -- redis-cli info | grep used_memory_human
   ```

2. Limpar caches excessivamente grandes se necessário
   ```bash
   kubectl exec -it redis-master-0 -n iam -- redis-cli --scan --pattern "auth:token:*" | xargs redis-cli del
   ```

3. Escalar horizontalmente o Redis se necessário
   ```bash
   kubectl scale statefulset redis-replica --replicas=3 -n iam
   ```

**Verificação de Resolução:**
- Tempo de resposta das operações de cache deve reduzir para < 10ms
- Cache hit rate deve aumentar para > 85%
- Latência da API deve normalizar em 1-3 minutos

### Cenário 3: Problemas em Serviços Externos Integrados

**Diagnóstico Detalhado:**
- Traces mostram gargalo em chamadas para serviço externo (ex: validação de MFA)
- Timeouts em chamadas HTTP externas
- Circuit breaker próximo de abrir

**Passos de Resolução:**

1. Verificar status dos serviços externos
   ```bash
   curl -v https://mfa-provider.example.com/health
   ```

2. Ativar modo de contingência se configurado
   ```bash
   kubectl patch configmap auth-api-config -n iam -p '{"data":{"CONTINGENCY_MODE":"true"}}'
   kubectl rollout restart deployment/auth-api -n iam
   ```

3. Notificar equipe responsável pelo serviço externo

**Verificação de Resolução:**
- Latência da API deve reduzir, mesmo que funcionalidades reduzidas
- Circuit breaker deve permanecer fechado
- Monitorar taxa de erros para garantir que permanece em níveis aceitáveis

## 🔄 Procedimento de Escalação

**Nível 1 - SRE de Plantão:**
- Nome: Equipe de Plantão SRE
- Contato: sre-oncall@innovabiz.com, +XX 9XXXX-XXXX
- Horário: 24x7

**Nível 2 - Especialista IAM:**
- Nome: Equipe de Especialistas IAM
- Contato: iam-team@innovabiz.com, +XX 9XXXX-XXXX
- Horário: Seg-Sex 8h-18h

**Nível 3 - Gerência Técnica:**
- Nome: Gerente de Infraestrutura
- Contato: infra-manager@innovabiz.com, +XX 9XXXX-XXXX
- Horário: Seg-Sex 9h-19h

## 🛡️ Prevenção

**Melhorias Identificadas:**
- Implementar cache em dois níveis (memória local + Redis)
- Otimizar queries frequentes do banco de dados
- Implementar pool de conexões com monitoramento e auto-healing
- Aumentar timeout para serviços externos críticos

**Alertas Preventivos Sugeridos:**
- Alerta para crescimento anormal no pool de conexões (> 70% por 10min)
- Alerta para queda no cache hit rate (< 80% por 15min)

## 📚 Referências

- [Documentação da Arquitetura IAM](https://wiki.innovabiz.com/iam/architecture)
- [Guia de Operação do PostgreSQL](https://wiki.innovabiz.com/databases/postgresql-ops)
- [Runbook de Recuperação de Redis](https://wiki.innovabiz.com/iam/redis-recovery)
- [Dashboard de Incidentes Históricos](https://incidents.innovabiz.com/history/iam)

## 📝 Histórico de Incidentes

| Data | ID Incidente | Resumo | Tenant | Região | Tempo de Resolução | Notas |
|------|-------------|--------|--------|--------|-------------------|-------|
| 2024-06-15 | INC-45892 | Latência alta por excesso de conexões | all | br | 28 minutos | Detectado padrão de não fechamento de conexões no código |
| 2024-05-03 | INC-42156 | Falha no Redis impactou autenticação | default | us | 45 minutos | Recuperação manual necessária após falha de rede |

## 🔄 Histórico de Revisões do Runbook

| Data | Versão | Autor | Mudanças |
|------|--------|-------|----------|
| 2024-04-10 | 1.0 | Ana Silva | Versão inicial |
| 2024-06-20 | 1.1 | Carlos Oliveira | Adicionado cenário de problema Redis |
| 2025-01-15 | 1.2 | Maria Santos | Atualizado para nova versão do Kubernetes e revisão geral |
```

## Tipos de Runbooks Requeridos

Cada módulo deve implementar pelo menos os seguintes tipos de runbooks operacionais:

### 1. Runbooks de Disponibilidade
- Indisponibilidade completa do serviço
- Falhas parciais em componentes específicos
- Degradação progressiva de serviço
- Recuperação após desastre

### 2. Runbooks de Performance
- Latência elevada em APIs
- Degradação de tempo de resposta
- Saturação de recursos
- Otimização de performance

### 3. Runbooks de Dados
- Problemas de integridade de dados
- Falhas em bancos de dados
- Corrupção de cache
- Migração de dados de emergência

### 4. Runbooks de Rede
- Falhas de conectividade
- Problemas de DNS
- Latência de rede
- Falhas em balanceadores de carga

### 5. Runbooks de Segurança
- Resposta a vulnerabilidades
- Detecção de atividades suspeitas
- Mitigação de ataques
- Recuperação após violação

## Melhores Práticas

1. **Clareza e Objetividade**
   - Use linguagem clara e direta
   - Inclua comandos exatos para copiar e colar
   - Evite ambiguidades que possam causar erros

2. **Estrutura Multi-dimensional**
   - Considere sempre o contexto de tenant e região
   - Inclua comandos adaptados para diferentes contextos
   - Documente impactos específicos por dimensão

3. **Revisão e Atualização**
   - Revise runbooks após cada incidente relacionado
   - Atualize regularmente (pelo menos a cada 3 meses)
   - Mantenha um histórico de versões para rastreabilidade

4. **Automatização**
   - Automatize procedimentos de diagnóstico quando possível
   - Inclua scripts auxiliares para verificações complexas
   - Mantenha links para ferramentas de automação relevantes

5. **Testabilidade**
   - Garanta que os runbooks sejam testados em ambientes não-produtivos
   - Realize simulações de incidentes periodicamente
   - Documente resultados de testes e refinamentos necessários

## Processo de Revisão e Validação

Os runbooks operacionais devem passar pelo seguinte processo de validação:

1. **Desenvolvimento inicial** - Escrito pelo time responsável pelo módulo
2. **Revisão técnica** - Revisado por SREs e especialistas em observabilidade
3. **Simulação** - Testado em ambiente controlado
4. **Revisão pós-incidente** - Atualizado após uso em incidente real
5. **Aprovação final** - Aprovado por gerente técnico e adicionado à base de conhecimento

## Checklist de Validação

- [ ] Todas as seções do template estão preenchidas
- [ ] Comandos testados e verificados em ambiente controlado
- [ ] Considerações multi-dimensionais (tenant, região) incluídas
- [ ] Links para ferramentas e recursos atualizados
- [ ] Procedimento de escalação claro e atualizado
- [ ] Histórico de incidentes documentado
- [ ] Histórico de revisões mantido
- [ ] Imagens/capturas de tela adicionadas para clareza (onde aplicável)
- [ ] Alertas relacionados documentados corretamente
- [ ] Linguagem clara e concisa em todo o documento

## Recursos Adicionais

- [Portal de Observabilidade INNOVABIZ](https://observability.innovabiz.com)
- [Biblioteca de Runbooks](https://wiki.innovabiz.com/observability/runbooks)
- [Framework de Resposta a Incidentes](https://wiki.innovabiz.com/incidents/framework)
- [Ferramenta de Automação de Runbooks](https://tools.innovabiz.com/runbook-automation)