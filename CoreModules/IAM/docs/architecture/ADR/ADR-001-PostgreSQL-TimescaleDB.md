# ADR-001: Adoção de PostgreSQL com extensão TimescaleDB para armazenamento de eventos de auditoria

## Status

Aprovado

## Data

2025-07-31

## Contexto

O IAM Audit Service requer uma solução de armazenamento de dados robusta para eventos de auditoria com as seguintes características:

- Alta performance para registros de séries temporais (time-series data)
- Suporte a consultas complexas para relatórios de compliance
- Escalabilidade para bilhões de registros de eventos
- Capacidade para múltiplos tenants e regiões
- Conformidade com requisitos regulatórios globais (PCI DSS, GDPR, LGPD, etc.)
- Capacidades avançadas de retenção e expurgo de dados
- Integração com o ecossistema de tecnologia existente da INNOVABIZ

## Decisão

Adotar **PostgreSQL com extensão TimescaleDB** como solução primária de armazenamento para eventos de auditoria do IAM Audit Service.

### Justificativa Técnica

- **TimescaleDB**: Extensão especializada para dados de série temporal que oferece:
  - Compressão nativa (10-20x redução no espaço de armazenamento)
  - Particionamento automático por tempo (chunks)
  - Índices otimizados para consultas baseadas em tempo
  - Alta performance para escritas em massa (batch inserts)

- **PostgreSQL**: Base de dados robusta e madura que oferece:
  - Suporte completo a ACID
  - Modelo de dados relacional e JSON/JSONB para flexibilidade
  - Capacidades avançadas de segurança (row-level security, column encryption)
  - Conformidade com standards SQL:2016
  - Extensibilidade via funções, procedimentos e tipos personalizados

### Alinhamento com Requisitos

| Requisito | Como é Atendido |
|-----------|----------------|
| Performance | Chunks otimizados por tempo, compressão nativa, índices hiperespecializados |
| Escalabilidade | Particionamento horizontal, compressão automática de dados históricos |
| Multi-tenancy | Row-level security e particionamento por tenant |
| Compliance | Imutabilidade configurável, auditoria de alterações, backup point-in-time |
| Retenção | Políticas de retenção nativas do TimescaleDB, automation jobs |
| Integração | Conectores nativos com Kafka, Grafana e ecossistema INNOVABIZ |

## Alternativas Consideradas

### 1. Solução NoSQL (MongoDB)

**Prós:**
- Flexibilidade no esquema para eventos heterogêneos
- Escalabilidade horizontal nativa

**Contras:**
- Menor suporte para consultas analíticas complexas
- Capacidades de agregação menos maduras
- Desafios para garantir ACID em cenário multi-regional
- Maior complexidade para atender requisitos de compliance

### 2. Elasticsearch

**Prós:**
- Excelente para busca e análise textual
- Bom suporte para visualizações via Kibana

**Contras:**
- Maior consumo de recursos
- Desafios com consistência de dados
- Custo operacional mais elevado
- Menos adequado para retenção de longo prazo com compliance

### 3. Solução Proprietária Especializada (Splunk)

**Prós:**
- Soluções específicas para auditoria e compliance
- Dashboards e relatórios prontos

**Contras:**
- Alto custo de licenciamento
- Lock-in de fornecedor
- Integração mais complexa com a arquitetura existente
- Menos flexibilidade para customizações

## Consequências

### Positivas

- Aproveitamento da expertise existente da equipe em PostgreSQL
- Redução de custos operacionais e de infraestrutura
- Simplificação da stack tecnológica (PostgreSQL já é usado em outros módulos)
- Melhor desempenho para consultas analíticas e de compliance
- Atendimento aos requisitos regulatórios sem soluções adicionais

### Negativas

- Necessidade de configuração cuidadosa para performance ótima
- Requer monitoramento especializado para TimescaleDB
- Potenciais desafios de escalabilidade extrema (petabytes)

### Mitigação de Riscos

- Implementar estratégia de hypertable partitioning otimizada por tenant e data
- Configurar política de retenção e compressão automatizada
- Estabelecer testes de performance e capacity planning regular
- Monitorar métricas específicas do TimescaleDB no sistema de observabilidade
- Implementar estratégia de backup incremental e recuperação ponto-a-ponto

## Conformidade com Padrões

- **ISO/IEC 27001**: Conformidade com controles de segurança da informação
- **PCI DSS 4.0**: Requisitos 10.2-10.7 para retenção e proteção de logs de auditoria
- **GDPR/LGPD**: Capacidades para atender direitos de titular e períodos de retenção
- **SOX**: Requisitos para trilhas de auditoria imutáveis

## Referências

1. TimescaleDB Documentation - https://docs.timescale.com/
2. PostgreSQL Security Features - https://www.postgresql.org/about/featurematrix/security/
3. Gartner Research: "Critical Capabilities for Operational Database Management Systems" (2025)
4. INNOVABIZ Data Architecture Standards v3.2
5. PCI DSS v4.0 Requirements for Audit Trail