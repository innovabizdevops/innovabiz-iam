# Plano de Observabilidade e Monitoramento - IAM INNOVABIZ

## Visão Geral

Este documento define a estratégia de observabilidade e monitoramento para o módulo IAM (Identity and Access Management) da plataforma INNOVABIZ, garantindo visibilidade completa sobre operações, performance, segurança e compliance em ambientes multi-tenant, multi-camada, multi-contexto e multi-dimensional. O plano está alinhado com frameworks TOGAF 10.0, COBIT 2019, ISO/IEC 42001, DMBOK 2.0 e requisitos regulatórios internacionais.

## Objetivos

1. **Detecção Proativa**: Identificar problemas antes que impactem os usuários
2. **Visibilidade Total**: Obter visão consolidada de todos os componentes do IAM
3. **Rastreabilidade Completa**: Permitir acompanhamento de transações de ponta a ponta
4. **Auditoria Forense**: Facilitar investigações de segurança e compliance
5. **Inteligência de Negócio**: Fornecer insights operacionais e estratégicos
6. **Compliance Automatizado**: Verificar continuamente conformidade com regulações
7. **Detecção de Anomalias**: Identificar comportamentos suspeitos ou anômalos

## Pilares da Observabilidade

### 1. Métricas

**Framework Principal**: OpenTelemetry Metrics + Prometheus

#### 1.1 Métricas de Aplicação

| Categoria | Métrica | Tipo | Cardinality | Finalidade | Alerta |
|-----------|---------|------|-------------|------------|--------|
| Autenticação | `iam.auth.attempts` | Counter | {status, method, tenant_id} | Volume de tentativas | >30% falha em 5min |
| Autenticação | `iam.auth.latency` | Histogram | {method, tenant_id, status} | Performance | p95 >500ms |
| Autorização | `iam.authz.requests` | Counter | {resource, action, tenant_id, status} | Volume de verificações | >40% negadas em 5min |
| Autorização | `iam.authz.latency` | Histogram | {resource, tenant_id} | Performance | p95 >100ms |
| Sessões | `iam.sessions.active` | Gauge | {tenant_id, auth_method} | Carga atual | >90% capacidade |
| Sessões | `iam.sessions.duration` | Histogram | {tenant_id, user_type} | Tempo de sessão | - |
| API | `iam.api.requests` | Counter | {path, method, status, tenant_id} | Uso de API | Erro >5% em 5min |
| API | `iam.api.latency` | Histogram | {path, method, tenant_id} | Performance | p95 >1s |
| Recursos | `iam.resource.usage` | Gauge | {resource_type, tenant_id} | Utilização | >85% capacidade |
| GraphQL | `iam.graphql.operations` | Counter | {operation, resolver, tenant_id, status} | Uso de GraphQL | Erro >5% em 5min |
| GraphQL | `iam.graphql.latency` | Histogram | {operation, resolver, tenant_id} | Performance | p95 >1s |
| Cache | `iam.cache.hits` | Counter | {cache_type, tenant_id} | Eficácia de cache | Taxa <60% |
| Rate Limiting | `iam.ratelimit.throttles` | Counter | {endpoint, tenant_id} | Proteção DoS | >20 por minuto |
| Segurança | `iam.security.events` | Counter | {event_type, severity, tenant_id} | Volume de eventos | Críticos >5 em 5min |

#### 1.2 Métricas de Infraestrutura

| Categoria | Métrica | Tipo | Cardinality | Finalidade | Alerta |
|-----------|---------|------|-------------|------------|--------|
| Recursos | `iam.infra.cpu` | Gauge | {service, instance} | Utilização CPU | >80% por 5min |
| Recursos | `iam.infra.memory` | Gauge | {service, instance} | Utilização RAM | >85% por 5min |
| Recursos | `iam.infra.disk` | Gauge | {service, instance} | Utilização disco | >85% |
| Rede | `iam.network.traffic` | Counter | {service, direction} | Tráfego de rede | Saturação >80% |
| Rede | `iam.network.errors` | Counter | {service, error_type} | Falhas de conexão | >1% de conexões |
| Kubernetes | `iam.k8s.pods` | Gauge | {service, status} | Estado de pods | Crashed >0 |
| Kubernetes | `iam.k8s.deployments` | Gauge | {service, status} | Estado de deployments | Não disponível >0 |

#### 1.3 Métricas de Negócio

| Categoria | Métrica | Tipo | Cardinality | Finalidade | Alerta |
|-----------|---------|------|-------------|------------|--------|
| Usuários | `iam.users.active` | Gauge | {tenant_id, user_type} | Usuários ativos | Queda >30% em 1h |
| Usuários | `iam.users.onboarding` | Counter | {tenant_id, acquisition_channel} | Novos usuários | Queda >50% da média |
| MFA | `iam.mfa.adoption` | Gauge | {tenant_id, method} | Adoção de MFA | <70% dos usuários |
| Compliance | `iam.compliance.score` | Gauge | {tenant_id, regulation} | Nível compliance | Score <85% |
| Compliance | `iam.compliance.violations` | Counter | {tenant_id, severity, regulation} | Violações | Crítica >0 |
| Experiência | `iam.ux.abandonment` | Counter | {tenant_id, auth_stage} | Abandono login | >15% abandono |
| Experiência | `iam.ux.recovery` | Counter | {tenant_id, recovery_type} | Recuperação conta | Aumento >50% |

### 2. Logs

**Framework Principal**: OpenTelemetry Logs + Elasticsearch/Loki + Graylog

#### 2.1 Estratégia de Logging

| Nível | Uso | Retenção | Exemplo |
|-------|-----|----------|---------|
| ERROR | Condições de falha | 2 anos | Falha de autenticação por credencial inválida |
| WARN | Situações potencialmente problemáticas | 1 ano | Tentativas repetidas de login |
| INFO | Atividades significativas | 6 meses | Login bem-sucedido, alteração de permissão |
| DEBUG | Informações detalhadas para troubleshooting | 15 dias | Parâmetros e fluxos de execução |
| TRACE | Informações granulares (apenas em ambiente dev/teste) | 3 dias | Passos detalhados de execução |

#### 2.2 Estrutura de Logs (JSON)

```json
{
  "timestamp": "2025-08-06T15:04:05.123Z",
  "level": "INFO",
  "service": "identity-service",
  "instance": "identity-svc-pod-1234",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "tenant_id": "acme-corp",
  "user_id": "u-12345",
  "session_id": "sess-6789",
  "operation": "UserAuthentication",
  "method": "password",
  "status": "success",
  "request_id": "req-abcd1234",
  "duration_ms": 45,
  "client_ip": "192.168.1.1",
  "user_agent": "Mozilla/5.0...",
  "resource_type": "user",
  "resource_id": "u-12345",
  "action": "authenticate",
  "details": {
    "factors_used": ["password", "otp"],
    "location": "BR-SP",
    "device_id": "dev-5678"
  },
  "security_event": true,
  "compliance": {
    "pci_dss": ["10.2.1"],
    "gdpr": ["Art. 32"],
    "lgpd": ["Art. 46"]
  }
}
```

#### 2.3 Padrões de Logging por Componente

| Componente | Eventos Críticos | Campos Específicos | Frequência |
|------------|------------------|-------------------|------------|
| AuthService | Login, logout, alteração de credencial | auth_method, mfa_status | Alto |
| PermissionService | Verificação, alteração de permissões | permission_code, resource_id | Alto |
| TenantService | Criação/modificação de tenant | tenant_type, parent_tenant | Baixo |
| UserService | CRUD de usuário, alteração de status | user_type, status_reason | Médio |
| TokenService | Emissão, validação, revogação | token_type, expire_time | Alto |
| GraphQL Resolvers | Execução de operações | resolver_name, operation_name | Alto |
| SecurityEventService | Eventos de segurança | event_category, severity | Médio |

### 3. Traces

**Framework Principal**: OpenTelemetry Tracing + Jaeger/Tempo

#### 3.1 Especificação de Spans

| Operação | Spans Obrigatórios | Atributos | Propagação |
|----------|-------------------|-----------|------------|
| Autenticação | `auth.validate`, `auth.mfa`, `auth.create_session` | auth_method, tenant_id | Externa |
| Autorização | `authz.validate_token`, `authz.check_permission` | resource, action, tenant_id | Externa |
| GraphQL Queries | `graphql.parse`, `graphql.validate`, `graphql.execute` | operation_name, tenant_id | Externa |
| GraphQL Mutations | `graphql.parse`, `graphql.validate`, `graphql.execute`, `db.transaction` | operation_name, tenant_id | Externa |
| Federação | `federation.validate`, `federation.resolve` | provider, tenant_id | Externa |
| Database | `db.query`, `db.transaction` | operation, table, tenant_id | Interna |
| External API | `http.client` | method, url, status_code | Externa |

#### 3.2 Sampling Strategy

| Tipo de Operação | Taxa de Amostragem | Critérios |
|------------------|-------------------|-----------|
| Autenticação | 100% | Operações críticas sempre traçadas |
| Autorizações | 20% base, 100% erros | Amostragem de operações normais |
| Consultas Padrão | 10% base, 100% lentas | Foco em problemas de performance |
| Operações críticas | 100% | Identificadas por flag em contexto |
| Troubleshooting | 100% | Ativado por sessão/usuário específico |

#### 3.3 Integração Trace-Log-Metric

- Todos os logs incluem `trace_id` e `span_id` quando disponíveis
- Métricas históricas incluem exemplars com trace_id para amostras significativas
- Traces incluem links para logs e métricas relacionados

### 4. Profiling

**Framework Principal**: Pyroscope/Parca + Continuous Profiling

#### 4.1 Tipos de Perfis

| Tipo | Frequência | Overhead | Finalidade |
|------|------------|----------|------------|
| CPU | Contínuo (1%) | Baixo | Identificar hotspots de processamento |
| Memória | Agendado (15min) | Médio | Detectar vazamentos e uso excessivo |
| Goroutines/Threads | Contínuo (1%) | Baixo | Verificar bloqueios e deadlocks |
| Mutex Contention | Sob demanda | Médio | Resolver gargalos de concorrência |

#### 4.2 Integração com Traces

- Ligação de perfis de CPU com spans de longa duração
- Correlação automática de problemas de memória com operações específicas
- Dashboard unificado de performance

## Instrumentação

### 1. Estratégia de Implementação

#### 1.1 Auto-instrumentação vs. Manual

| Componente | Abordagem | Justificativa |
|------------|-----------|---------------|
| HTTP/gRPC | Auto | Cobertura completa com middleware |
| GraphQL | Híbrida | Auto para framework, manual para lógica de negócio |
| Database | Auto | ORM/driver já instrumentado |
| Lógica de Negócio | Manual | Contexto de negócio requer instrumentação específica |
| Autenticação/Autorização | Manual | Requer atributos específicos de segurança |

#### 1.2 Propagação de Contexto

- Uso de W3C Trace Context para propagação entre serviços HTTP
- Implementação de propagadores personalizados para filas de mensagens
- Injeção de tenant_id em todos os contextos para isolamento multi-tenant

#### 1.3 Bibliotecas e SDK

| Linguagem | Bibliotecas | Componentes |
|-----------|------------|------------|
| Go | OpenTelemetry Go SDK, otelgrpc, otelsql | Serviços core, API Gateway |
| TypeScript/Node.js | @opentelemetry/sdk-node, auto-instrumentations | Admin Portal, Dashboards |
| Python | OpenTelemetry Python, auto-instrumentations | Scripts, ferramentas auxiliares |

## Visualização e Análise

### 1. Dashboards

#### 1.1 Dashboards Operacionais

| Dashboard | Público-alvo | Conteúdo | Atualização |
|-----------|-------------|----------|------------|
| IAM Overview | SRE, DevOps | KPIs gerais, saúde, alertas ativos | Tempo real |
| Authentication Service | Equipe IAM | Métricas detalhadas, logs, traces | Tempo real |
| Authorization Service | Equipe IAM | Métricas detalhadas, logs, traces | Tempo real |
| GraphQL Performance | Desenvolvedores | Operações mais lentas, erros comuns | 5 minutos |
| Infrastructure | SRE, DevOps | CPU, memória, rede, K8s | 1 minuto |
| SLO/SLI Dashboard | SRE, Gerência | Performance vs objetivos | 10 minutos |

#### 1.2 Dashboards de Negócio

| Dashboard | Público-alvo | Conteúdo | Atualização |
|-----------|-------------|----------|------------|
| User Acquisition | Produto, Marketing | Onboarding, conversão, abandono | Diária |
| Security Posture | Segurança, Compliance | Eventos, violações, tendências | Hora |
| MFA Adoption | Produto, Segurança | Taxas de adoção por tenant e método | Diária |
| Compliance Status | Legal, Compliance | Status por regulação, tendências | Diária |
| Executive Overview | Executivos | KPIs consolidados | Diária |

### 2. Alerting

#### 2.1 Estratégia de Alertas

| Nível | Tempo de Resposta | Notificação | Exemplo |
|-------|-------------------|------------|---------|
| P1 - Crítico | 15 minutos, 24/7 | Chamada, SMS, Email | Falha total de autenticação |
| P2 - Alto | 1 hora, 24/7 | SMS, Email, Slack | Degradação significativa de serviço |
| P3 - Médio | Horário comercial | Email, Slack | Alerta de tendência negativa |
| P4 - Baixo | Próximo dia útil | Ticket, Slack | Recomendação de melhoria |

#### 2.2 Regras de Alerta

| Nome | Condição | Severidade | Ações | Silenciamento |
|------|----------|-----------|-------|--------------|
| HighAuthFailureRate | iam.auth.attempts{status="failed"} / iam.auth.attempts > 0.3 por 5min | P2 | Notificar Security, IAM | Durante manutenções |
| SlowAuthPerformance | iam.auth.latency{quantile="0.95"} > 500ms por 10min | P3 | Notificar IAM | Durante implantações |
| CriticalSecurityEvent | iam.security.events{severity="critical"} > 0 | P1 | Notificar Security, IAM | Nunca |
| ApiErrorRateHigh | iam.api.requests{status=~"5.."} / iam.api.requests > 0.05 por 5min | P2 | Notificar IAM, SRE | Durante implantações |
| ComplianceScoreLow | iam.compliance.score < 85 | P3 | Notificar Compliance | Durante auditorias |
| ResourceSaturation | iam.infra.cpu > 85% por 15min | P3 | Notificar SRE | Durante scale events |

#### 2.3 Redução de Ruído e Alert Fatigue

- Implementação de correlação de alertas para reduzir notificações duplicadas
- Ajuste automático de limiares baseado em padrões históricos e ML
- Rotação de oncall com horários de descanso garantidos
- Revisão mensal da eficácia dos alertas e ajustes necessários

### 3. Integração com ITSM

- Geração automática de incidentes no ServiceNow para alertas P1/P2
- Vinculação de incidentes com métricas, logs e traces relevantes
- Workflow automatizado para alertas recorrentes
- Relatórios pós-incidente com links para dados de observabilidade

## Análise Avançada

### 1. Machine Learning para Observabilidade

#### 1.1 Detecção de Anomalias

| Caso de Uso | Algoritmos | Dados de Entrada | Ação |
|-------------|-----------|----------------|------|
| Padrões anômalos de autenticação | Isolation Forest, LSTM | Métricas de auth, logs | Alerta de segurança |
| Degradação de performance | Forecast, ARIMA | Métricas de latência | Alerta proativo |
| Falhas de infraestrutura | Random Forest | Métricas de sistema | Previsão de problemas |
| Comportamento anômalo de usuário | k-means, Isolation Forest | Logs de atividade | Notificação de segurança |

#### 1.2 Análise de Causa Raiz

- Análise automatizada de dependências para correlacionar falhas
- Sugestões de resolução baseadas em incidentes anteriores
- Timeline de eventos para facilitar investigação

#### 1.3 Previsão de Capacidade

- Projeção de crescimento de usuários e recursos necessários
- Recomendação de ajustes de escala com base em padrões de uso
- Otimização de custos de infraestrutura

### 2. Observabilidade Contínua

#### 2.1 Testes em Produção

| Técnica | Implementação | Benefício |
|---------|--------------|-----------|
| Canary Deployments | Implantação gradual com monitoramento intensivo | Detecção precoce de problemas |
| Synthetic Monitoring | Verificações de login/auth a cada 1 minuto | Validação contínua de funcionalidade |
| Chaos Engineering | Falhas controladas em componentes IAM | Verificação de resiliência |

#### 2.2 SLOs e Error Budgets

| SLO | SLI | Meta | Budget |
|-----|-----|------|--------|
| Disponibilidade de Autenticação | % de requests 2xx/3xx | 99.95% | 21.9 minutos/mês |
| Latência de Autenticação | % requests < 500ms | 99.5% | 3.6 horas/mês |
| Disponibilidade de API | % de requests 2xx/3xx | 99.9% | 43.8 minutos/mês |
| Latência de API | % requests < 1s | 99.5% | 3.6 horas/mês |
| Taxa de erro de autorização | % decisões corretas | 99.999% | 26 segundos/mês |

## Gestão e Governança

### 1. Gestão de Custo e Volume

- Retenção diferenciada por tipo e criticidade de dados
- Amostragem inteligente para traces e logs de baixa prioridade
- Compressão e indexação otimizadas
- Revisão mensal de custos e volumes por equipe

### 2. Controle de Acesso e Segurança

- RBAC para acesso a dashboards e alertas
- Isolamento multi-tenant para visualização de dados
- Mascaramento de informações sensíveis (PII) em logs e traces
- Auditoria de acesso a ferramentas de observabilidade

### 3. Capacitação e Cultura

- Treinamento regular em observabilidade para equipes
- Documentação e wikis atualizadas sobre ferramentas e práticas
- Revisão pós-incidente com foco em melhoria contínua
- Gamificação da instrumentação e qualidade de observabilidade

## Plano de Implementação

### Fase 1: Fundação (M1-M2)

1. Configuração da infraestrutura de coleta (OpenTelemetry Collector)
2. Instrumentação básica de métricas em serviços críticos
3. Implementação de logging estruturado padronizado
4. Dashboards operacionais essenciais

### Fase 2: Expansão (M3-M4)

1. Instrumentação completa de traces em todos os serviços
2. Implementação de alertas para métricas críticas
3. Integração com ITSM para gestão de incidentes
4. Dashboards de negócio iniciais

### Fase 3: Otimização (M5-M6)

1. Implementação de análise de anomalias com ML
2. Profiling contínuo em serviços críticos
3. SLOs/SLIs formalmente definidos e monitorados
4. Testes de caos e resiliência

### Fase 4: Maturidade (M7-M12)

1. Previsão e planejamento baseados em dados históricos
2. Automação avançada de remediação
3. Otimização contínua baseada em métricas de negócio
4. Expansão para observabilidade centrada no usuário

## Métricas de Sucesso

| Métrica | Baseline | Meta 6 meses | Meta 12 meses |
|---------|----------|--------------|--------------|
| MTTR (Mean Time to Resolve) | 4 horas | 2 horas | 45 minutos |
| MTTD (Mean Time to Detect) | 30 minutos | 10 minutos | 2 minutos |
| % Incidentes detectados proativamente | 40% | 70% | 90% |
| % Serviços com instrumentação completa | 30% | 80% | 100% |
| Satisfação com dashboards (pesquisa) | N/A | 7/10 | 9/10 |
| Redução de falsos positivos | N/A | 30% | 70% |

## Referências

1. OpenTelemetry Documentation - https://opentelemetry.io/docs/
2. Google SRE Handbook - https://sre.google/sre-book/
3. Prometheus Best Practices - https://prometheus.io/docs/practices/
4. Observability Engineering (Charity Majors, et al.)
5. ISO 27001:2022 - Requisitos de monitoramento
6. COBIT 2019 - Práticas de gestão DSS01, DSS03
7. NIST Cybersecurity Framework - Categoria DE.AE (Anomalies and Events)

---

*Este documento está em conformidade com os padrões de documentação técnica da INNOVABIZ e deve ser revisado e atualizado regularmente conforme a evolução do sistema.*

*Última atualização: 06/08/2025*