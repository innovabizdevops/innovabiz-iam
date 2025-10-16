# Dashboard de Qualidade e Conformidade IAM

## Visão Geral

Este documento descreve o Dashboard de Qualidade e Conformidade para monitoramento dos validadores de conformidade do IAM integrados com o Sistema de Gestão da Qualidade da plataforma INNOVABIZ. O dashboard permite acompanhar métricas, não-conformidades e ações corretivas em tempo real, fornecendo uma visão completa da conformidade com os padrões de qualidade.

## Arquitetura do Dashboard

O dashboard foi projetado seguindo uma arquitetura baseada em camadas:

1. **Camada de Dados**: Visões SQL e funções para agregar e transformar dados do sistema de qualidade
2. **Camada de Lógica**: Funções e procedimentos para cálculos de KPIs e análises
3. **Camada de Apresentação**: Interface visual com gráficos, tabelas e indicadores

Esta arquitetura permite uma separação clara entre os dados, a lógica de negócio e a apresentação, facilitando manutenção e extensibilidade.

## Componentes Principais

### 1. Visões da Base de Dados

O dashboard utiliza as seguintes visões da base de dados para fornecer dados agregados:

| Visão | Descrição |
|-------|-----------|
| `dashboard_quality_metrics` | Resumo de métricas de qualidade com status (OK, ATENÇÃO, CRÍTICO) |
| `dashboard_non_conformities_summary` | Resumo de não-conformidades por padrão, impacto e status |
| `dashboard_corrective_actions_summary` | Resumo de ações corretivas por padrão e status |
| `dashboard_non_conformity_trends` | Tendências de não-conformidades nos últimos 12 meses |
| `dashboard_corrective_action_effectiveness` | Eficácia das ações corretivas implementadas |

### 2. Funções de API

O dashboard utiliza as seguintes funções para obter dados para os diferentes componentes visuais:

| Função | Descrição |
|--------|-----------|
| `get_dashboard_kpis` | Retorna os indicadores-chave de desempenho (KPIs) principais |
| `get_non_conformities_by_standard` | Retorna distribuição de não-conformidades por padrão |
| `get_non_conformities_by_impact` | Retorna distribuição de não-conformidades por nível de impacto |
| `get_recent_non_conformities` | Retorna as não-conformidades mais recentes |

## Painéis e Visualizações

O dashboard contém os seguintes painéis principais:

### 1. Painel de KPIs

![Painel de KPIs](https://exemplo.com/imagens/kpi-panel.png)

Este painel exibe os indicadores-chave de desempenho (KPIs) em formato de cartões, incluindo:

- Taxa de Conformidade (%)
- Não-Conformidades Abertas
- Ações Corretivas Pendentes
- Ações Corretivas Atrasadas
- Tempo Médio de Resolução (dias)

Cada indicador é exibido com um código de cores (verde, amarelo, vermelho) baseado em limiares predefinidos.

### 2. Painel de Distribuição de Não-Conformidades

![Distribuição de Não-Conformidades](https://exemplo.com/imagens/distribution-panel.png)

Este painel contém gráficos de distribuição que mostram:

- Distribuição de não-conformidades por padrão de qualidade (gráfico de pizza)
- Distribuição de não-conformidades por nível de impacto (gráfico de barras)
- Distribuição de não-conformidades por status (gráfico de donut)

### 3. Painel de Tendências

![Tendências de Conformidade](https://exemplo.com/imagens/trends-panel.png)

Este painel exibe gráficos de tendências ao longo do tempo:

- Tendência de não-conformidades por mês (últimos 12 meses)
- Tendência de taxa de conformidade (últimos 12 meses)
- Tempo médio de resolução ao longo do tempo

### 4. Painel de Ações Corretivas

![Painel de Ações Corretivas](https://exemplo.com/imagens/actions-panel.png)

Este painel exibe informações sobre ações corretivas:

- Status de ações corretivas (pendentes, em andamento, concluídas)
- Eficácia das ações corretivas implementadas
- Lista das ações corretivas mais recentes com prazos

### 5. Lista de Não-Conformidades Recentes

Este painel exibe uma tabela com as não-conformidades mais recentes, incluindo:

- ID da não-conformidade
- Padrão de qualidade
- Requisito
- Nível de impacto
- Status
- Data de criação
- Número de ações associadas

## Configuração e Personalização

### Configuração de Perfis de Usuário

O dashboard suporta dois perfis principais de usuário:

1. **Gestor de Qualidade (quality_manager_role)**: Acesso completo a todas as visões e funções, com permissões para atualizar status de ações corretivas.

2. **Visualizador de Qualidade (quality_viewer_role)**: Acesso de leitura a todas as visões e painéis, sem permissões para fazer atualizações.

### Personalização de Visualizações

O dashboard pode ser personalizado das seguintes formas:

- **Filtros por Padrão**: Filtrar visualizações por padrão específico (ex: ISO 9001, ISO 27001, HIPAA)
- **Filtros por Período**: Selecionar período de análise (último mês, último trimestre, último ano)
- **Filtros por Status**: Visualizar apenas não-conformidades ou ações com status específico
- **Filtros por Impacto**: Focar em não-conformidades de alto, médio ou baixo impacto

## Integrações

O dashboard se integra com os seguintes componentes da plataforma INNOVABIZ:

1. **Sistema de Gestão de Qualidade**: Dados de não-conformidades e ações corretivas
2. **IAM Compliance Validators**: Resultados de validação de conformidade IAM
3. **Sistema de Gestão de Riscos**: Visualização de riscos associados a não-conformidades
4. **Sistema de Gestão de Incidentes**: Visualização de incidentes gerados a partir de não-conformidades

## Exemplos de Uso

### Caso de Uso 1: Monitoramento de Conformidade Diária

Um gestor de qualidade utiliza o dashboard diariamente para:
1. Verificar os KPIs gerais para identificar áreas de preocupação
2. Revisar novas não-conformidades e ações corretivas pendentes
3. Acompanhar prazos de ações corretivas para evitar atrasos
4. Atribuir responsáveis para novas ações corretivas

### Caso de Uso 2: Análise Mensal de Tendências

Um gestor de compliance utiliza o dashboard mensalmente para:
1. Analisar tendências de não-conformidades ao longo do tempo
2. Identificar padrões de qualidade com maior número de problemas
3. Avaliar a eficácia das ações corretivas implementadas
4. Preparar relatórios para a alta administração

### Caso de Uso 3: Auditoria de Conformidade

Durante uma auditoria de conformidade, o dashboard é utilizado para:
1. Demonstrar o processo de gestão de não-conformidades
2. Evidenciar a implementação de ações corretivas
3. Apresentar métricas e indicadores de melhoria contínua
4. Comprovar a conformidade com padrões específicos

## Requisitos Técnicos

### Requisitos de Sistema

- PostgreSQL 12 ou superior
- Servidor de aplicação compatível com a pilha tecnológica da plataforma
- Navegador moderno com suporte a JavaScript e SVG

### Dependências

- Script SQL `13_quality_management_integration.sql` já executado
- Schema `quality_management` criado e populado
- Roles `quality_manager_role` e `quality_viewer_role` criadas

## Próximos Passos e Melhorias Futuras

1. **Integração com Business Intelligence**: Conectar o dashboard a ferramentas de BI como PowerBI ou Tableau
2. **Alertas Proativos**: Implementar sistema de alertas automáticos para quando métricas ultrapassarem limiares críticos
3. **Previsões com IA**: Adicionar capacidades preditivas para antecipar possíveis não-conformidades
4. **Exportação de Relatórios**: Adicionar funcionalidade para exportar relatórios em formatos como PDF e Excel
5. **Personalização por Usuário**: Permitir que usuários salvem suas configurações de visualização preferidas

## Referências

- ISO 9001:2015 - Sistema de Gestão da Qualidade
- ISO/IEC 27001:2013 - Sistema de Gestão de Segurança da Informação
- ISO 19011:2018 - Diretrizes para Auditoria de Sistemas de Gestão
- ISO 31000:2018 - Gestão de Riscos
- Documentação da Plataforma INNOVABIZ
