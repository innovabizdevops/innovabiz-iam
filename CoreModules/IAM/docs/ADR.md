# Decisões de Arquitetura (ADR) - MCP-IAM Observability

## ADR 001: Arquitetura Multi-Dimensional para Observabilidade

### Contexto

A plataforma INNOVABIZ opera em múltiplos mercados globais (Angola, Brasil, UE, EUA, China, CPLP, SADC, PALOP, BRICS), cada um com requisitos regulatórios e de compliance específicos. Precisamos projetar uma solução de observabilidade para os hooks MCP-IAM que suporte todas estas dimensões simultaneamente, mantendo conformidade com múltiplos frameworks como TOGAF 10.0, COBIT 2019, DMBOK 2.0, ISO 27001, etc.

### Decisão

Implementaremos uma arquitetura de observabilidade multi-dimensional que segmenta a telemetria (logs, métricas e traces) em quatro dimensões principais:

1. **Dimensão de Mercado**: Segmentação por região geográfica/regulatória (Angola, Brasil, UE, etc.)
2. **Dimensão de Tenant**: Segmentação por tipo de cliente (Individual, Business, Government)
3. **Dimensão de Compliance**: Segmentação por framework regulatório (BNA, LGPD, GDPR, etc.)
4. **Dimensão Técnica**: Segmentação por tipo de telemetria e componente técnico

Esta abordagem será implementada usando:
- **Contextualização de Mercado**: Toda telemetria será enriquecida com metadados de mercado
- **Telemetria Dimensional**: Traces, logs e métricas segmentados por dimensão
- **Armazenamento Segregado**: Dados sensíveis armazenados conforme requisitos locais
- **Exportadores Específicos**: Configuração de exportadores dedicados por mercado

### Consequências

#### Positivas
- Conformidade simultânea com múltiplos regimes regulatórios
- Visibilidade granular por mercado e tipo de tenant
- Capacidade de implementar requisitos específicos por região
- Auditoria segmentada conforme necessidades regulatórias

#### Negativas
- Maior complexidade de implementação e manutenção
- Aumento do volume de dados de telemetria
- Necessidade de configuração específica por mercado

### Status
Aprovado

## ADR 002: Escolha de OpenTelemetry como Framework de Observabilidade

### Contexto

A solução MCP-IAM Observability precisa coletar, processar e exportar dados de telemetria (logs, métricas e traces) de forma consistente e interoperável entre todos os módulos core da plataforma INNOVABIZ, mantendo conformidade com múltiplos frameworks de governança e compliance.

### Decisão

Adotaremos o OpenTelemetry como o framework principal para implementação da observabilidade na plataforma INNOVABIZ:

1. **Instrumentação**: Biblioteca OpenTelemetry Go para todas as instrumentações
2. **Coleta**: OpenTelemetry Collector para coleta, processamento e exportação
3. **Protocolos**: OTLP (OpenTelemetry Protocol) para comunicação padronizada
4. **Exportadores**: Suporte a múltiplos backends (Prometheus, Jaeger, Elasticsearch)
5. **Contexto**: Propagação de contexto W3C TraceContext para rastreabilidade entre serviços

Esta decisão se alinha com os seguintes princípios da INNOVABIZ:
- **Interoperabilidade**: Padrão aberto e agnóstico de fornecedor
- **Transversalidade**: Cobertura completa das 3 dimensões de observabilidade
- **Priorização Estratégica**: Adoção de padrão emergente de mercado
- **Escalabilidade**: Arquitetura desacoplada e distribuída
- **Conformidade**: Flexibilidade para atender requisitos específicos por mercado

### Consequências

#### Positivas
- Instrumentação unificada para logs, métricas e traces
- Independência de fornecedor e portabilidade
- Compatibilidade com múltiplos backends de observabilidade
- Facilidade de integração entre módulos e componentes
- Comunidade ativa e adoção crescente no mercado

#### Negativas
- Maturidade relativa (padrão ainda em evolução)
- Maior curva de aprendizado inicial
- Documentação em desenvolvimento

### Status
Aprovado

## ADR 003: Estratégia de Compliance Multi-Framework

### Contexto

A plataforma INNOVABIZ precisa atender simultaneamente a múltiplos frameworks regulatórios e de compliance em diferentes mercados, incluindo BNA (Angola), LGPD/BACEN (Brasil), GDPR/PSD2 (Europa), SOX/CCPA (EUA) e CSL/PIPL (China), além de frameworks globais como ISO 27001.

### Decisão

Implementaremos uma estratégia de compliance multi-framework para observabilidade com os seguintes componentes:

1. **Metadados de Compliance por Mercado**:
   - Registro de frameworks aplicáveis
   - Requisitos de MFA específicos
   - Períodos de retenção de logs
   - Requisitos de aprovação dual

2. **Validadores Dinâmicos**:
   - Sistema de validação contextual de MFA
   - Verificação de escopo adaptativa por mercado
   - Controles específicos por framework

3. **Auditoria Adaptativa**:
   - Eventos de auditoria específicos por framework
   - Retenção diferenciada por requisito regulatório
   - Formatação e armazenamento conforme requisitos locais

4. **Matriz de Mapeamento de Controles**:
   - Mapeamento de controles técnicos para múltiplos frameworks
   - Implementação unificada atendendo a múltiplas regulações

Esta abordagem permite atender a múltiplos frameworks simultaneamente com implementação técnica unificada.

### Consequências

#### Positivas
- Conformidade simultânea com múltiplos frameworks
- Reuso de controles técnicos entre frameworks
- Adaptabilidade a novos requisitos regulatórios
- Capacidade de validação de compliance automatizada

#### Negativas
- Maior complexidade na lógica de validação
- Necessidade de manutenção da matriz de mapeamento
- Risco de conflitos entre requisitos de diferentes frameworks

### Status
Aprovado

## ADR 004: Integração com API Gateway Krakend

### Contexto

A plataforma INNOVABIZ utiliza o Krakend como API Gateway central para exposição de serviços. É necessário definir como a solução MCP-IAM Observability se integrará com este gateway para garantir rastreabilidade completa, autenticação/autorização consistente e monitoramento de compliance em todos os fluxos de requisição.

### Decisão

Implementaremos uma integração bidirecional entre o adaptador MCP-IAM Observability e o API Gateway Krakend com as seguintes características:

1. **Propagação de Contexto de Rastreabilidade**:
   - Headers W3C TraceContext (traceparent, tracestate)
   - Headers de contexto de mercado (X-Market, X-Tenant-Type)
   - Headers de correlação (X-Correlation-ID)

2. **Middleware de Telemetria**:
   - Plugin Krakend para exposição de métricas Prometheus
   - Instrumentação OpenTelemetry para traces de requisições
   - Logging estruturado em formato compatível com ELK

3. **Integração com IAM**:
   - Verificação de tokens JWT via middleware
   - Validação de escopos com granularidade por mercado
   - Auditoria de acessos com context enrichment

4. **Middleware de Compliance**:
   - Validação de requisitos específicos por mercado
   - Rate limiting contextual baseado em políticas
   - Registro de eventos de compliance

Esta abordagem garantirá observabilidade completa de ponta a ponta, com contexto preservado desde o API Gateway até os serviços internos.

### Consequências

#### Positivas
- Rastreabilidade completa de requisições externas até operações internas
- Enriquecimento de contexto consistente
- Políticas de segurança aplicadas no perímetro
- Monitoramento centralizado no ponto de entrada

#### Negativas
- Overhead de processamento adicional no gateway
- Complexidade de configuração aumentada
- Dependência do ciclo de vida do Krakend

### Status
Aprovado

## ADR 005: Armazenamento e Retenção Segmentada de Telemetria

### Contexto

Diferentes mercados onde a plataforma INNOVABIZ opera possuem requisitos distintos para armazenamento e retenção de dados de telemetria, especialmente para logs de auditoria e eventos de segurança. Por exemplo, BNA (Angola) exige 7 anos de retenção, enquanto LGPD (Brasil) estabelece 5 anos para certos dados.

### Decisão

Implementaremos uma estratégia de armazenamento e retenção segmentada de telemetria com as seguintes características:

1. **Segmentação Física por Mercado**:
   - Instâncias dedicadas de armazenamento por região regulatória
   - Políticas de retenção específicas por mercado
   - Backups segmentados por framework de compliance

2. **Hierarquia de Armazenamento**:
   - Hot storage: 30 dias de dados acessíveis para consulta rápida
   - Warm storage: 90-180 dias para consultas ocasionais
   - Cold storage: Dados históricos para compliance de longo prazo
   
3. **Políticas de Retenção Baseadas em Metadata**:
   - Tempo de retenção determinado por metadados de compliance
   - Retenção diferenciada por tipo de evento e criticidade
   - Rotação e exclusão automatizada conforme políticas

4. **Criptografia e Controle de Acesso**:
   - Criptografia em repouso específica por mercado
   - Controle de acesso granular baseado em RBAC
   - Segregação de responsabilidades para acesso a dados sensíveis

Esta abordagem permitirá conformidade simultânea com requisitos de retenção e armazenamento de múltiplas jurisdições.

### Consequências

#### Positivas
- Conformidade com requisitos de retenção por mercado
- Otimização de custos com armazenamento hierárquico
- Segregação adequada de dados sensíveis
- Capacidade de auditoria específica por mercado

#### Negativas
- Maior complexidade de infraestrutura
- Custos elevados de armazenamento para longos períodos
- Desafios de consulta em dados distribuídos

### Status
Aprovado

## ADR 006: Monitoramento e Alertas Multi-Dimensionais

### Contexto

A natureza multi-dimensional da plataforma INNOVABIZ (multi-mercado, multi-tenant, multi-contexto) torna necessário um sistema de monitoramento e alertas capaz de detectar anomalias e violações de compliance em todas essas dimensões simultaneamente, com diferentes limiares e políticas por mercado e tenant.

### Decisão

Implementaremos uma estratégia de monitoramento e alertas multi-dimensionais com os seguintes componentes:

1. **Definição Contextual de Limiares**:
   - Limiares de alerta específicos por mercado
   - Sensibilidade ajustada por tipo de tenant
   - Regras específicas por framework de compliance

2. **Categorização de Alertas**:
   - Alertas de segurança (tentativas de fraude, anomalias de autenticação)
   - Alertas de compliance (violações de regras, requisitos não atendidos)
   - Alertas operacionais (performance, disponibilidade, erros)
   - Alertas de negócio (volumes transacionais, tendências)

3. **Roteamento e Escalação Contextual**:
   - Roteamento de alertas baseado em mercado e criticidade
   - Escalação adaptativa conforme severidade e SLAs
   - Notificações diferenciadas por tipo de alerta

4. **Dashboards Segmentados**:
   - Visualizações específicas por mercado
   - Painéis dedicados para compliance por framework
   - Visões operacionais vs. visões de negócio

Esta abordagem garantirá detecção precoce e resposta adequada a eventos em todas as dimensões da plataforma.

### Consequências

#### Positivas
- Detecção precisa de anomalias específicas por contexto
- Redução de ruído com alertas contextualizados
- Resposta adequada conforme criticidade e mercado
- Visibilidade adaptada às necessidades de diferentes stakeholders

#### Negativas
- Maior complexidade na configuração e manutenção de alertas
- Risco de sobrecarga de alertas se mal configurados
- Necessidade de revisão contínua dos limiares

### Status
Aprovado

## ADR 007: Adaptação Dinâmica de Hooks por Mercado

### Contexto

Os hooks MCP-IAM precisam adaptar seu comportamento conforme o mercado, tipo de tenant e contexto de operação para garantir conformidade com requisitos regulatórios específicos e fornecer telemetria adequada para cada cenário.

### Decisão

Implementaremos um sistema de adaptação dinâmica de hooks com os seguintes componentes:

1. **Detecção de Contexto**:
   - Identificação automática de mercado e tenant
   - Carregamento de configurações específicas
   - Mapeamento para frameworks de compliance aplicáveis

2. **Adaptação de Comportamento**:
   - Validações MFA com rigor ajustado por mercado
   - Verificações de escopo com requisitos específicos
   - Auditoria adaptada ao contexto regulatório

3. **Configuração Declarativa**:
   - Regras de adaptação definidas em arquivos de configuração
   - Metadados de compliance carregados dinamicamente
   - Mapeamento de operações para requisitos regulatórios

4. **Injeção de Dependências Contextual**:
   - Carregamento de implementações específicas por mercado
   - Factory methods para criação de serviços contextuais
   - Estratégias de fallback para cenários não cobertos

Esta abordagem permitirá que os hooks MCP-IAM se comportem adequadamente em qualquer contexto de execução.

### Consequências

#### Positivas
- Conformidade automática com requisitos específicos
- Flexibilidade para adaptação a novos mercados
- Manutenção simplificada de regras de negócio
- Desacoplamento entre lógica core e requisitos específicos

#### Negativas
- Maior complexidade de código e configuração
- Overhead de processamento para detecção de contexto
- Risco de comportamento inconsistente se mal configurado

### Status
Aprovado

## ADR 008: Integração Total entre Módulos Core

### Contexto

A plataforma INNOVABIZ possui múltiplos módulos core (IAM, Payment Gateway, Risk Management, Mobile Money, etc.) que precisam compartilhar contexto de observabilidade para fornecer visibilidade completa de ponta a ponta em operações que atravessam múltiplos módulos.

### Decisão

Implementaremos uma estratégia de Integração Total de Observabilidade entre módulos core com os seguintes componentes:

1. **Propagação de Contexto Distribuído**:
   - Uso de W3C TraceContext para propagação entre serviços
   - Headers customizados para contexto de mercado e tenant
   - Baggage items para metadados de compliance

2. **Interface de Observabilidade Comum**:
   - Contrato `ObservabilityConsumer` implementado por todos os módulos
   - Métodos padronizados para logging, métricas e traces
   - Enriquecimento consistente com metadados de compliance

3. **Tracing Distribuído**:
   - Spans aninhados com relacionamento pai-filho entre módulos
   - Atributos padronizados para correlação
   - Visualização unificada de fluxos cross-módulo

4. **Agregação de Telemetria**:
   - Correlação de eventos entre múltiplos módulos
   - Métricas agregadas por fluxo de negócio completo
   - Dashboards integrados mostrando fluxos completos

Esta abordagem permitirá visibilidade completa de operações complexas que atravessam múltiplos módulos da plataforma.

### Consequências

#### Positivas
- Rastreabilidade completa de ponta a ponta
- Visibilidade unificada de operações cross-módulo
- Diagnóstico simplificado de problemas complexos
- Métricas de negócio em nível de fluxo completo

#### Negativas
- Forte acoplamento entre módulos para propagação de contexto
- Overhead de performance para correlação
- Desafios de versionamento de interfaces compartilhadas

### Status
Aprovado

## ADR 009: Estratégia de Teste e Validação Contínua

### Contexto

A solução MCP-IAM Observability necessita garantir conformidade contínua com múltiplos frameworks regulatórios, mesmo com mudanças frequentes no código, configurações e requisitos por mercado. É necessária uma estratégia abrangente de teste e validação que verifique todos os aspectos da solução.

### Decisão

Implementaremos uma estratégia de teste e validação contínua com os seguintes componentes:

1. **Testes Unitários Contextuais**:
   - Testes específicos por mercado e framework
   - Validação de comportamento adaptativo
   - Mocks para serviços externos e dependências

2. **Testes de Integração Multi-Dimensional**:
   - Testes segmentados por mercado
   - Validação de fluxos completos entre módulos
   - Verificação de propagação de contexto

3. **Validação Automatizada de Compliance**:
   - Verificação automatizada de requisitos regulatórios
   - Testes específicos para cada framework
   - Validação de níveis de auditoria e retenção

4. **Simulação de Cenários de Produção**:
   - Injeção de falhas e cenários anômalos
   - Verificação de alertas e resposta a incidentes
   - Testes de carga com volumetria realista por mercado

Esta abordagem garantirá que a solução mantenha sua conformidade e qualidade em todos os cenários de operação.

### Consequências

#### Positivas
- Detecção precoce de problemas de compliance
- Confiança na adaptação a novos mercados
- Validação contínua de requisitos regulatórios
- Documentação viva dos requisitos implementados

#### Negativas
- Maior complexidade na manutenção dos testes
- Tempo de execução elevado para suíte completa
- Necessidade de ambientes de teste específicos por mercado

### Status
Aprovado

## ADR 010: Arquitetura Evolutiva para Novos Mercados e Requisitos

### Contexto

A plataforma INNOVABIZ está em constante expansão para novos mercados e precisa se adaptar rapidamente a novos requisitos regulatórios e frameworks de compliance. É necessária uma arquitetura que permita evolução sem grandes refatorações.

### Decisão

Implementaremos uma arquitetura evolutiva para a solução MCP-IAM Observability com os seguintes princípios:

1. **Design Modular por Mercado**:
   - Módulos plugáveis para cada mercado
   - Configurações isoladas por framework regulatório
   - Desacoplamento entre lógica core e regras específicas

2. **Versionamento de Adaptadores**:
   - Versionamento semântico de adaptadores por mercado
   - Compatibilidade retroativa garantida
   - Migração gradual para novas versões

3. **Feature Flags Contextuais**:
   - Ativação gradual de recursos por mercado/tenant
   - Testes A/B para novos comportamentos
   - Rollback segmentado em caso de problemas

4. **Documentação como Código**:
   - ADRs para todas as decisões significativas
   - Geração automática de documentação de compliance
   - Rastreabilidade entre requisitos e implementação

Esta abordagem permitirá que a plataforma evolua para atender novos mercados e requisitos de forma controlada e segura.

### Consequências

#### Positivas
- Adaptação rápida a novos mercados e regulações
- Risco reduzido em atualizações e evoluções
- Manutenção simplificada de múltiplas versões
- Documentação sempre atualizada com o código

#### Negativas
- Maior complexidade inicial de design
- Overhead de performance com abstrações adicionais
- Necessidade de disciplina no versionamento e documentação

### Status
Aprovado