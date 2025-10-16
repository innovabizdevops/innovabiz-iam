# Índice de Templates de Integração para Observabilidade INNOVABIZ

## Visão Geral

Este documento serve como índice central para todos os templates de integração com o Framework de Observabilidade INNOVABIZ. Estes templates seguem as melhores práticas, padrões e requisitos estabelecidos no Framework de Integração, garantindo consistência, qualidade e conformidade em toda a plataforma.

## Organização dos Templates

Os templates foram organizados em arquivos separados para facilitar a consulta e implementação por equipes de diferentes especialidades:

1. **Instrumentação de Código**
   - [Template para Node.js (Express)](./templates/NODEJS_EXPRESS_TEMPLATE.md)
   - [Template para Python (FastAPI)](./templates/PYTHON_FASTAPI_TEMPLATE.md)
   - [Template para Java (Spring)](./templates/JAVA_SPRING_TEMPLATE.md)
   - [Template para Go](./templates/GO_TEMPLATE.md)

2. **Configuração de Infraestrutura**
   - [Template para OpenTelemetry Collector](./templates/OTEL_COLLECTOR_TEMPLATE.md)
   - [Template para Kubernetes - Infraestrutura Geral](./templates/KUBERNETES_TEMPLATE.md)
   - [Template para Kubernetes - Grafana](./templates/KUBERNETES_YAML/GRAFANA_TEMPLATE.md)
   - [Template para Kubernetes - Loki](./templates/KUBERNETES_YAML/LOKI_TEMPLATE.md)
   - [Template para Kubernetes - Tempo](./templates/KUBERNETES_YAML/TEMPO_TEMPLATE.md)
   - [Template para Kubernetes - AlertManager](./templates/KUBERNETES_YAML/ALERTMANAGER_TEMPLATE.md)

3. **Visualização e Monitoramento**
   - [Template para Dashboards Grafana](./templates/GRAFANA_DASHBOARD_TEMPLATE.md)
   - [Template para Alertas Prometheus](./templates/PROMETHEUS_ALERTS_TEMPLATE.md)
   - [Índice de Dashboards de Observabilidade](./templates/dashboards/DASHBOARDS_INDEX.md)
   - [Dashboard para AlertManager](./templates/dashboards/ALERTMANAGER_DASHBOARD.json)
   - [Documentação do Dashboard AlertManager](./templates/dashboards/ALERTMANAGER_DASHBOARD_DOCUMENTATION.md)
   - [Dashboard para Kubernetes Cluster Overview](./templates/dashboards/KUBERNETES_CLUSTER_DASHBOARD.json) *Atualizado*
   - [Documentação do Dashboard Kubernetes](./templates/dashboards/KUBERNETES_CLUSTER_DOCUMENTATION.md) *Atualizado*
   - [Dashboard para PostgreSQL](./templates/dashboards/POSTGRESQL_DASHBOARD.json) *Implementado*
   - [Documentação do Dashboard PostgreSQL](./templates/dashboards/POSTGRESQL_DASHBOARD_DOCUMENTATION.md) *Implementado*
   - [Dashboard para Redis](./templates/dashboards/REDIS_DASHBOARD.json) *Implementado*
   - [Documentação do Dashboard Redis](./templates/dashboards/REDIS_DASHBOARD_DOCUMENTATION.md) *Implementado*
   - [Dashboard para Kafka](./templates/dashboards/KAFKA_DASHBOARD.json) *Implementado*
   - [Documentação do Dashboard Kafka](./templates/dashboards/KAFKA_DASHBOARD_DOCUMENTATION.md) *Implementado*

4. **Documentação e Operação**
   - [Template para Documentação de Observabilidade](./templates/OBSERVABILITY_DOCS_TEMPLATE.md)
   - [Template para Runbooks Operacionais](./templates/OPERATIONAL_RUNBOOK_TEMPLATE.md)
   - [Template para Infraestrutura como Código (IaC)](./templates/INFRASTRUCTURE_AS_CODE_TEMPLATE.md)
   - [Regras de Alerta Padronizadas](./templates/rules/STANDARD_ALERT_RULES.md) *Novo*
   - [Runbook Operacional AlertManager](./templates/runbooks/ALERTMANAGER_OPERATIONAL_RUNBOOK.md) *Novo*

## Como Usar os Templates

1. **Escolha o template adequado** para sua necessidade específica (instrumentação, configuração, visualização ou documentação)
2. **Copie o template base** para seu projeto ou módulo
3. **Personalize os parâmetros** conforme necessário para seu contexto específico
4. **Siga as melhores práticas** documentadas em cada template
5. **Valide a implementação** usando a checklist fornecida

## Templates de Observabilidade Completa

A plataforma INNOVABIZ fornece agora um conjunto completo de templates para implementação de observabilidade end-to-end para AlertManager:

1. **Documentação Operacional**:
   - [Runbook Operacional AlertManager](./templates/runbooks/ALERTMANAGER_OPERATIONAL_RUNBOOK.md)

2. **Configuração de Alertas**:
   - [Regras de Alerta Padronizadas](./templates/rules/STANDARD_ALERT_RULES.md)

3. **Visualização e Monitoramento**:
   - [Dashboard para AlertManager](./templates/dashboards/ALERTMANAGER_DASHBOARD.json)
   - [Documentação do Dashboard AlertManager](./templates/dashboards/ALERTMANAGER_DASHBOARD_DOCUMENTATION.md)

4. **Índice Organizado**:
   - [Índice de Dashboards de Observabilidade](./templates/dashboards/DASHBOARDS_INDEX.md)

## Requisitos Obrigatórios

Todos os templates aderem aos seguintes requisitos obrigatórios da plataforma INNOVABIZ:

- **Multi-contexto completo**: tenant, região, módulo e ambiente
- **Segurança**: TLS 1.3, mTLS quando aplicável, autenticação e autorização
- **Conformidade**: PCI DSS, GDPR/LGPD, ISO 27001, NIST CSF
- **Performance**: otimização para baixa latência e overhead mínimo
- **Padronização**: nomenclatura, tags e atributos consistentes
- **Extensibilidade**: facilidade para adição de novos contextos e métricas

## Processo de Contribuição e Atualização

Para sugerir melhorias ou atualizações aos templates:

1. Revise o template existente
2. Prepare sua proposta seguindo as diretrizes de contribuição
3. Submeta para revisão pela equipe de Arquitetura
4. Após aprovação, os templates serão atualizados

## Suporte e Recursos Adicionais

- **Documentação Expandida**: Consulte a wiki interna para casos de uso avançados
- **Comunidade**: Participe do canal #observability no Slack corporativo
- **Treinamentos**: Verifique o calendário de workshops e capacitações
- **FAQ**: [Perguntas Frequentes sobre Observabilidade](./FAQ_OBSERVABILITY.md)

---

**Proprietário**: Equipe de Plataforma INNOVABIZ  
**Contato**: platform-team@innovabiz.com  
**Última Atualização**: Julho 2025  
**Status**: 🚀 Em evolução contínua