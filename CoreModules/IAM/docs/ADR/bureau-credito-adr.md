# ADR: Integração MCP-IAM com Bureau de Crédito (Central de Risco)

## Status

Aprovado

## Data

2025-02-18

## Contexto

A plataforma INNOVABIZ necessita de um módulo Bureau de Crédito (Central de Risco) robusto e altamente integrado com o sistema MCP-IAM Observability para atender aos requisitos multi-mercado, multi-tenant, multi-contexto e multi-dimensional. Este módulo é crucial para processos de concessão de crédito, análise de risco, prevenção de fraudes e atendimento às regulamentações específicas dos diversos mercados-alvo (Angola, Brasil, União Europeia, EUA, China, CPLP, SADC, PALOP e BRICS).

## Decisão

Implementaremos o módulo Bureau de Crédito com as seguintes características arquiteturais:

1. **Arquitetura Multi-Dimensional e Multi-Mercado**:
   - Design adaptável às regulações específicas de cada mercado (BNA, BACEN, GDPR, FCRA)
   - Parametrização dinâmica de regras por mercado e por tipo de consulta
   - Suporte a múltiplos tenants com isolamento completo de dados

2. **Integração Profunda com MCP-IAM Observability**:
   - Traçabilidade completa de todas as consultas via OpenTelemetry
   - Registro detalhado de eventos de auditoria, segurança e compliance
   - Métricas granulares para monitoramento de performance e uso

3. **Modelo de Dados Flexível**:
   - Tipos adaptáveis de consulta (Completa, Score, Básica, Restrições, Histórico, Relacionamento)
   - Suporte a diferentes finalidades de consulta conforme exigências regulatórias
   - Estrutura extensível para registros de crédito e restrições

4. **Mecanismo de Compliance e Autorização Avançado**:
   - Regras de compliance configuráveis por mercado e tipo de consulta
   - Sistema de controle de acesso baseado em escopos específicos
   - Validação de MFA adaptativa conforme criticidade da consulta

5. **Gestão de Consentimento**:
   - Verificação obrigatória de consentimento explícito por mercado
   - Prazos de validade de consentimento configuráveis por finalidade
   - Auditoria completa do ciclo de vida do consentimento

6. **Ciclo de Vida Robusto**:
   - Worker para reset diário de contadores de consulta
   - Inicialização e encerramento graciosos
   - Gestão de limites e quotas por entidade

7. **Observabilidade Abrangente**:
   - Rastreamento de spans para cada etapa do processamento
   - Histogramas de tempo de resposta por tipo de consulta
   - Contadores para consultas, restrições e notificações

8. **Notificações e Alertas Regulatórios**:
   - Sistema de notificação obrigatória conforme requisitos específicos (BNA, LGPD, GDPR, FCRA)
   - Relatório aos reguladores quando aplicável
   - Alertas para violações de compliance

## Consequências

### Positivas

1. **Conformidade Regulatória**: Atendimento aos requisitos específicos de cada mercado, facilitando o processo de obtenção e manutenção de licenças operacionais.

2. **Auditabilidade Completa**: Rastreamento detalhado de todas as operações, permitindo verificação e investigação facilitada em caso de disputas ou auditorias.

3. **Segurança Reforçada**: Controles granulares de autorização e autenticação, com MFA adaptável à criticidade da operação.

4. **Escalabilidade**: Arquitetura que permite adição de novos mercados, novas regras de compliance e novos tipos de consulta com impacto mínimo no código existente.

5. **Integração Total**: Conexão perfeita com demais módulos da plataforma, como Payment Gateway, Risk Management, IAM e Mobile Money.

6. **Observabilidade Avançada**: Monitoramento em tempo real das operações, com métricas detalhadas e alertas automáticos.

7. **Experiência do Usuário**: Tempos de resposta otimizados e resultados precisos adaptados às necessidades específicas do usuário.

### Negativas

1. **Complexidade**: A arquitetura multi-dimensional e as regras específicas por mercado aumentam a complexidade do sistema.

2. **Overhead de Processamento**: O extenso registro de eventos de auditoria, métricas e spans adiciona algum overhead de processamento.

3. **Requisitos de Armazenamento**: O armazenamento de dados históricos de auditoria e logs exige capacidade considerável, especialmente para consultas de longo prazo.

4. **Desafios de Manutenção**: A atualização constante para acompanhar mudanças regulatórias em múltiplos mercados exige esforço contínuo.

## Alternativas Consideradas

1. **Solução por Mercado**: Implementar módulos separados para cada mercado. Rejeitada por causar duplicação de código e dificultar a manutenção.

2. **Serviço Externo**: Utilizar apenas APIs externas de bureaus de crédito existentes. Rejeitada por limitações de personalização, custos elevados e desafios de integração.

3. **Simplificação do Modelo**: Reduzir tipos de consulta e regras de compliance. Rejeitada por não atender aos requisitos regulatórios diversos.

4. **Processamento Assíncrono**: Realizar processamento assíncrono de consultas complexas. Considerada válida para futuras extensões, mas não implementada na versão inicial por requisitos de resposta em tempo real.

## Compliance e Frameworks Aplicáveis

- **Angola**: BNA (Banco Nacional de Angola), Lei de Proteção de Dados
- **Brasil**: BACEN, SCR (Sistema de Informações de Crédito), LGPD
- **União Europeia**: GDPR, PSD2
- **EUA**: FCRA (Fair Credit Reporting Act), GLBA (Gramm-Leach-Bliley Act)
- **Global**: ISO 27001, ISO 27701, PCI DSS, TOGAF

## Métricas e Observabilidade

### Principais Métricas

1. **Volume**:
   - `bureau_credito_consultas_total` (por tipo e mercado)
   - `bureau_credito_consultas_sucesso` (por tipo e mercado)
   - `bureau_credito_limite_excedido` (por entidade)

2. **Performance**:
   - `bureau_credito_tempo_processamento` (histograma por tipo)
   - `bureau_credito_registros_retornados` (por tipo de consulta)

3. **Compliance**:
   - `bureau_credito_notificacoes` (por regulador)
   - `bureau_credito_registros_regulador` (por regulador)

### Spans OpenTelemetry

1. `bureau_credito_consulta` - Span principal da consulta
2. `verificar_autenticacao` - Verificação de autenticação
3. `verificar_autorizacao` - Verificação de autorização
4. `verificar_regras_acesso` - Verificação de regras de acesso
5. `verificar_limite_consultas` - Verificação de limites
6. `verificar_consentimento` - Verificação de consentimento
7. `verificar_compliance` - Verificação de compliance
8. `processar_consulta` - Processamento da consulta
9. `processar_notificacoes` - Processamento de notificações

## Integração com Outros Módulos

1. **IAM**: Autenticação, autorização e gestão de identidades
2. **Risk Management**: Fornecimento de dados para análise de risco
3. **Payment Gateway**: Validação de clientes e análise de risco de transações
4. **Mobile Money**: Avaliação de crédito para concessão de limites
5. **E-Commerce/Marketplace**: Verificação de clientes e prevenção de fraudes
6. **Microcrédito**: Avaliação de crédito para empréstimos

## Próximos Passos

1. Implementar adaptadores específicos para fontes de dados externas
2. Desenvolver dashboards de monitoramento específicos para o Bureau de Crédito
3. Criar testes automatizados para simulação de cenários regulatórios específicos
4. Desenvolver API REST para exposição do serviço via API Gateway Krakend
5. Implementar cache distribuído para otimização de consultas frequentes
6. Desenvolver sistema de pontuação proprietário adaptado aos mercados-alvo

## Referências

1. Regulamentações BNA para Centrais de Risco em Angola
2. BACEN - Resolução CMN sobre SCR (Sistema de Informações de Crédito)
3. GDPR - Artigos específicos sobre processamento de dados financeiros
4. FCRA - Fair Credit Reporting Act (EUA)
5. ISO 27001:2022 - Requisitos de segurança da informação
6. OpenTelemetry Specification v1.0
7. Guia de Implementação MCP-IAM Observability Integration

## Aprovação

- **Autor**: Equipe de Desenvolvimento INNOVABIZ
- **Revisor**: Comitê Técnico INNOVABIZ
- **Aprovador**: Eduardo Jeremias - CTO INNOVABIZ
- **Data de Aprovação**: 2025-02-18