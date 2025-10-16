# ADR-003: Estratégia de Observabilidade Total para o Módulo IAM

**Status:** Aprovado  
**Data:** 2025-08-06  
**Autor:** Equipa INNOVABIZ DevSecOps  
**Stakeholders:** Arquitetura, Desenvolvimento, SRE, Segurança, Compliance, Operações  

## Contexto

O módulo IAM é crítico para a segurança e operação de toda a plataforma INNOVABIZ, gerenciando autenticação, autorização e identidades para todos os componentes. Falhas ou degradações neste módulo podem comprometer toda a plataforma, além de potencialmente criar riscos de segurança e conformidade. Uma estratégia abrangente de observabilidade é essencial para garantir operação confiável, detecção precoce de problemas, e cumprimento de requisitos regulatórios em todos os mercados atendidos.

## Opções Consideradas

1. **Observabilidade Básica**
   * Logging centralizado
   * Métricas principais de sistema (CPU, memória)
   * Monitoramento básico de disponibilidade

2. **Observabilidade Intermediária**
   * Logging estruturado centralizado
   * Métricas de negócio e técnicas
   * Tracing para fluxos principais
   * Alertas para condições críticas

3. **Observabilidade Total Avançada**
   * Logging contextual e correlacionado com trace IDs
   * Métricas granulares técnicas, de negócio e regulatórias
   * Tracing distribuído para 100% das operações
   * Análise comportamental com AI/ML
   * Alertas preditivos e adaptativos
   * Visualizações personalizadas por perfil de utilizador
   * Observabilidade como código

4. **Solução de Terceiros**
   * Adoção de plataforma comercial completa de APM
   * Delegação da implementação para produto de mercado
   * Dependência de vendor específico

## Decisão

**Adotar uma estratégia de Observabilidade Total Avançada baseada em OpenTelemetry como padrão para todo o módulo IAM.**

Esta decisão se baseia nos seguintes fatores:

1. **Criticidade do Módulo**: Como componente central de segurança, o IAM requer o mais alto nível de observabilidade possível.

2. **Requisitos Regulatórios**: Exigências de auditoria detalhada e monitoramento de segurança em todos os mercados-alvo (Angola/BNA, EU/GDPR, EUA/SOX).

3. **Arquitetura Distribuída**: A natureza multi-regional e multi-tenant do sistema requer capacidade de correlacionar eventos em diferentes serviços e localizações.

4. **Detecção Proativa**: Necessidade de identificar anomalias de segurança e performance antes que impactem usuários.

5. **Open Standards**: Alinhamento com a estratégia tecnológica de adoção de padrões abertos e evitar vendor lock-in.

## Implementação Técnica

### Pilares da Implementação

1. **Logging**
   * Logs estruturados em JSON com atributos padronizados
   * Níveis de log consistentes (ERROR, WARN, INFO, DEBUG, TRACE)
   * Correlação com trace IDs e span IDs
   * Enriquecimento contextual (tenant ID, user ID, region)
   * Redação automática de dados sensíveis (PII, credenciais)
   * Armazenamento em Loki com retenção configurável por criticidade

2. **Métricas**
   * Coletadas via OpenTelemetry e armazenadas em Prometheus
   * Métricas técnicas: Latência, throughput, taxa de erro, utilização de recursos
   * Métricas de negócio: Autenticações bem-sucedidas/falhas, operações por tenant
   * Métricas regulatórias: Eventos de segurança, alterações de permissão
   * Dimensões consistentes: service, instance, tenant_id, region
   * Histogramas para distribuição de latência com buckets padronizados
   * Retenção longa em Thanos para análise histórica

3. **Tracing**
   * Instrumentação via OpenTelemetry para 100% dos fluxos
   * Propagação de contexto W3C TraceContext
   * Sampling adaptativo baseado em importância da operação
   * Atributos padronizados em todos os spans
   * Spans específicos para operações críticas de segurança
   * Storage em Jaeger com UI customizada para fluxos IAM

4. **Dashboards & Visualização**
   * Dashboards Grafana por perfil de utilizador:
     * Operacional: Visão de saúde do serviço
     * Segurança: Eventos e anomalias de segurança
     * Compliance: Métricas regulatórias e auditoria
     * Desenvolvimento: Performance e debugging
   * Exploração correlacionada entre logs, métricas e traces

5. **Alertas Inteligentes**
   * Regras de alerta hierárquicas
   * Alertas preditivos baseados em tendências anômalas
   * Roteamento de alertas por severidade e contexto
   * Redução de ruído via correlação de alertas

### Componentes Técnicos

* **OpenTelemetry Collector**: Para coleta e processamento de telemetria
* **Prometheus + Thanos**: Para armazenamento e consulta de métricas
* **Loki**: Para armazenamento e consulta de logs
* **Jaeger**: Para armazenamento e visualização de traces
* **Grafana**: Para dashboards e alertas
* **AlertManager**: Para gestão e roteamento de alertas
* **Vector**: Para processamento e enriquecimento de logs

## Consequências

### Positivas

* Visibilidade completa sobre o comportamento do sistema
* Capacidade aprimorada de diagnosticar problemas
* Redução no MTTR (Mean Time To Resolve) para incidentes
* Evidência automática para auditorias de compliance
* Capacidade de detecção proativa de anomalias de segurança
* Base para implementação futura de AIOps

### Negativas

* Overhead de performance pela instrumentação (estimado em 3-5%)
* Complexidade adicional no código e infraestrutura
* Volume maior de dados para armazenamento e processamento
* Curva de aprendizado para equipes operacionais

### Mitigações

* Sampling inteligente para reduzir overhead em operações de alto volume
* Automação de instrumentação via bibliotecas e frameworks
* Estratégias de compressão e retenção para gerenciamento de custos
* Treinamento dedicado para equipes operacionais

## Métricas-Chave de IAM a Serem Monitoradas

### Métricas de Performance

* `iam.request.duration_ms` - Latência de requisições (p50, p95, p99)
* `iam.request.rate` - Número de requisições por segundo
* `iam.error.rate` - Taxa de erros por tipo e endpoint
* `iam.db.query.duration_ms` - Latência de queries de banco de dados
* `iam.cache.hit_ratio` - Taxa de acertos no cache

### Métricas de Segurança

* `iam.login.success_rate` - Taxa de logins bem-sucedidos
* `iam.login.failure_rate` - Taxa de falhas de login por razão
* `iam.suspicious_activity.count` - Contagem de atividades suspeitas
* `iam.permission.change.count` - Alterações em permissões
* `iam.token.validation.count` - Validações de token
* `iam.brute_force.attempt.count` - Tentativas de força bruta detectadas

### Métricas de Negócio

* `iam.active_users.count` - Usuários ativos por tenant
* `iam.active_sessions.count` - Sessões ativas
* `iam.new_users.count` - Novos usuários registrados
* `iam.group_operations.count` - Operações em grupos por tipo

### Métricas Regulatórias

* `iam.pii.access.count` - Acessos a dados PII
* `iam.admin.action.count` - Ações administrativas
* `iam.gdpr.request.count` - Requisições relacionadas a GDPR
* `iam.sensitive_operation.count` - Operações em dados sensíveis

## Logs e Eventos Críticos

| Categoria | Eventos | Nível | Retenção |
|-----------|---------|-------|----------|
| Segurança | Login falhou, Permissão negada, Token invalidado | ERROR | 2 anos |
| Auditoria | Mudança de permissão, Criação/exclusão de usuário | INFO | 2 anos |
| Performance | Latência alta, Timeout, Falha em dependência | WARN | 6 meses |
| Operacional | Inicialização de serviço, Configuração carregada | INFO | 3 meses |
| Debug | Detalhes de requisição/resposta, Estado interno | DEBUG | 7 dias |

## Conformidade e Governança

Esta estratégia de observabilidade está em conformidade com:

* **ISO/IEC 27001:2022** - Monitoramento e logging para segurança da informação (A.12.4)
* **PCI DSS v4.0** - Requisitos de logging e monitoramento (Req. 10)
* **SOX** - Trilhas de auditoria para controles financeiros
* **GDPR/LGPD** - Registro de operações em dados pessoais (Art. 30)
* **BNA Instrução 7/2021** - Requisitos de monitoramento para serviços financeiros
* **NIST 800-53** - Controles de auditoria e accountability (AU)

## Verificação

O sucesso desta decisão será medido através de:

* Redução de 30% no MTTR para incidentes relacionados ao IAM
* Detecção proativa de 90% dos incidentes antes do impacto ao usuário
* Zero incidentes não detectados em auditorias de segurança
* Satisfação de 100% dos requisitos regulatórios em auditorias

## Referências

* [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
* [Google SRE Book - Monitoring Distributed Systems](https://sre.google/sre-book/monitoring-distributed-systems/)
* [NIST SP 800-92 - Guide to Computer Security Log Management](https://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication800-92.pdf)
* [Observability Engineering - Charity Majors, Liz Fong-Jones](https://www.oreilly.com/library/view/observability-engineering/9781492076438/)