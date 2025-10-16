# ADR-002: Framework de Observabilidade com Prometheus e Grafana para IAM Audit Service

## Status

Aprovado

## Data

2025-07-31

## Contexto

O IAM Audit Service necessita de um framework completo de observabilidade para garantir monitoramento em tempo real, detecção proativa de problemas e visibilidade operacional completa, atendendo aos seguintes requisitos:

- Capacidades abrangentes de coleta e exposição de métricas
- Suporte a contextos múltiplos (tenant, região, ambiente)
- Compatibilidade com padrões arquiteturais da plataforma INNOVABIZ
- Alertas configuráveis por severidade e canais
- Visualizações detalhadas para diferentes stakeholders
- Integração nativa com FastAPI e ecossistema Python
- Requisitos rigorosos de conformidade (PCI DSS, ISO 27001, GDPR, LGPD)
- Suporte a cenários multi-regionais e alta disponibilidade

## Decisão

Implementar um framework completo de observabilidade baseado em **Prometheus e Grafana** com customizações específicas para o IAM Audit Service, incluindo:

1. **Coleção de Métricas**:
   - Prometheus como coletor central de métricas
   - Exporters customizados para métricas de auditoria
   - Instrumentação via middleware FastAPI e decoradores Python
   - OpenMetrics como formato padronizado

2. **Visualização**:
   - Grafana como plataforma de visualização
   - Dashboards especializados para métricas de auditoria, compliance e operações
   - Variáveis de template para filtros multi-contexto (ambiente, região, tenant)
   - Painéis específicos para gestão de retenção de dados

3. **Alertas e Notificações**:
   - Alertmanager para gestão de alertas
   - Integração com canais múltiplos (Slack, Email, SMS)
   - Alertas configuráveis por severidade, SLA e contexto
   - Redução de ruído via deduplicação e inibição inteligente

### Justificativa Técnica

- **Prometheus**: 
  - Modelo de pull para coleta de métricas
  - Linguagem de consulta PromQL poderosa para expressões complexas
  - Modelo de dados dimensional (multidimensional time series)
  - Escalabilidade horizontal via federação e Thanos
  - Integração nativa com Kubernetes e ecossistema cloud-native

- **Grafana**:
  - Visualizações avançadas e dashboards interativos
  - Suporte nativo para PromQL
  - Variáveis de template para filtragem dinâmica
  - Alerting unificado e notificações
  - APIs para automação e gestão programática

- **Customizações INNOVABIZ**:
  - Biblioteca de instrumentação específica para IAM Audit Service
  - Endpoints padronizados (/metrics, /health, /diagnostic)
  - Métricas específicas para eventos de auditoria, retenção e compliance
  - Decoradores para instrumentação automática de funções críticas
  - Middleware para captura de métricas HTTP e erros

## Alternativas Consideradas

### 1. ELK Stack (Elasticsearch, Logstash, Kibana)

**Prós:**
- Excelente para logs e análise textual
- Capacidades avançadas de busca full-text

**Contras:**
- Maior consumo de recursos
- Menos eficiente para métricas numéricas
- Maior complexidade operacional
- Custo mais elevado para escala

### 2. Datadog (SaaS)

**Prós:**
- Solução integrada de observabilidade
- Rápida implementação

**Contras:**
- Custos elevados baseados em volume
- Dependência de provedor externo
- Potenciais desafios com regulamentações regionais
- Menor flexibilidade para customizações específicas

### 3. OpenTelemetry + Backend Customizado

**Prós:**
- Framework unificado para métricas, logs e traces
- Padrão emergente da indústria

**Contras:**
- Maior esforço de implementação inicial
- Necessidade de backend de armazenamento adicional
- Ecossistema ainda em maturação
- Curva de aprendizado mais acentuada

## Consequências

### Positivas

- Alinhamento com o padrão de observabilidade da plataforma INNOVABIZ
- Capacidade de detecção proativa de problemas
- Visibilidade multi-dimensional para métricas de auditoria
- Dashboards especializados para diferentes stakeholders
- Monitoramento eficiente de KPIs de compliance e segurança
- Custo reduzido por reutilização de infraestrutura existente

### Negativas

- Necessidade de gestão de infraestrutura Prometheus
- Potencial para sobrecarga de métricas sem governança adequada
- Requisitos de armazenamento para retenção de longo prazo

### Mitigação de Riscos

- Implementar políticas de cardinality management para evitar explosão de séries
- Configurar retention policies adequadas para diferentes tipos de métricas
- Automatizar a gestão de dashboards via Grafana provisioning
- Implementar documentação detalhada sobre métricas disponíveis
- Centralizar alertas para evitar fadiga de alertas

## Conformidade com Padrões

- **ISO 20000**: Gestão de serviços de TI
- **ISO/IEC 42010**: Arquitetura de sistemas e software
- **ITIL v4**: Observabilidade como parte da gestão de operações
- **SRE Principles**: Monitoramento baseado em service levels
- **Gartner MAIO Framework**: Monitoring, Artificial Intelligence, IT automation and Observability

## Implementação

A implementação inclui:

1. Classe `ObservabilityIntegration` centralizada para FastAPI
2. Middleware para instrumentação automática de HTTP
3. Decoradores para métricas de eventos de auditoria e compliance
4. Endpoints `/metrics`, `/health` e `/diagnostic` padronizados
5. Dashboard Grafana com painéis para retenção, compliance, métricas HTTP e monitoramento de erros
6. Configuração de alertas Prometheus com múltiplos níveis de severidade
7. Documentação completa para operação e extensão

## Referências

1. INNOVABIZ Platform Observability Standards v2.5
2. Prometheus Best Practices - https://prometheus.io/docs/practices/
3. Grafana Dashboard Design Principles - https://grafana.com/docs/grafana/latest/best-practices/
4. Google SRE Book: Monitoring Distributed Systems - https://sre.google/sre-book/monitoring-distributed-systems/
5. PCI DSS 4.0 Requirements for Monitoring & Detection (Req. 11)