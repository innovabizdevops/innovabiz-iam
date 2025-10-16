# Interface de Visualização do Dashboard Open X - Documentação Técnica

**Versão:** 1.0.0  
**Data:** 15/05/2025  
**Autor:** INNOVABIZ DevOps  
**Status:** ✅ Concluído

## 1. Visão Geral

A Interface de Visualização do Dashboard Open X é uma solução completa de frontend para visualização, interação e análise das métricas e indicadores de conformidade do ecossistema Open X na plataforma INNOVABIZ. Esta interface fornece componentes visuais interativos, como gráficos, tabelas, cards e filtros, permitindo que usuários de diferentes perfis acessem as informações relevantes para suas funções.

A interface foi projetada seguindo princípios de design responsivo, acessibilidade e experiência do usuário, garantindo uma navegação intuitiva e eficiente pelos diferentes dashboards e visualizações disponíveis.

## 2. Arquitetura da Interface

### 2.1 Componentes Principais

A interface é composta pelos seguintes componentes principais:

1. **Dashboards**: Conjuntos específicos de visualizações organizadas para atender a diferentes necessidades:
   - Dashboard Principal Open X
   - Dashboard por Domínio Open X
   - Dashboard de Impacto Econômico Open X
   - Dashboard por Jurisdição Open X

2. **Componentes Visuais**:
   - **Cards**: Exibem métricas principais com indicadores de status
   - **Gráficos**: Visualizações interativas (barras, linhas, pizza) para análise de dados
   - **Tabelas**: Dados detalhados com recursos de ordenação, filtragem e paginação
   - **Filtros**: Controles para personalizar a visualização dos dados

3. **Sistema de Layout**:
   - Grid responsivo baseado em coordenadas x,y
   - Componentes redimensionáveis e configuráveis
   - Suporte a diferentes tamanhos de tela

### 2.2 Modelo de Dados

A interface utiliza as seguintes estruturas de dados:

1. **Dashboards**: Definição e metadados dos dashboards disponíveis
2. **Componentes**: Configuração dos elementos visuais de cada dashboard
3. **Preferências de Usuário**: Configurações personalizadas salvas por usuário
4. **Histórico de Visualização**: Registro de interações para análise de uso

### 2.3 Integração com Backend

A interface se integra com:

- **Validadores Open X**: Consome dados das views e funções definidas no script `18_open_x_dashboard.sql`
- **Sistema de Modelagem Econômica**: Utiliza os cálculos de impacto econômico
- **Sistema IAM**: Gerencia permissões e acesso baseado em papéis

## 3. Dashboards Disponíveis

### 3.1 Dashboard Principal Open X

O Dashboard Principal fornece uma visão consolidada de todo o ecossistema Open X:

**Filtros**:
- Seleção de Tenant
- Período de análise

**Métricas principais**:
- Conformidade geral (percentual)
- Domínios em conformidade (contagem)
- Impacto econômico total (valor monetário)

**Visualizações**:
- Gráfico de barras de conformidade por domínio
- Gráfico de pizza de impacto econômico por domínio
- Tabela de não-conformidades críticas (nível de risco R3 e R4)
- Gráfico de linha de tendência de conformidade ao longo do tempo

### 3.2 Dashboard por Domínio Open X

Permite análise detalhada de um domínio específico (Open Insurance, Open Health ou Open Government):

**Filtros**:
- Seleção de Tenant
- Seleção de Domínio Open X

**Métricas principais**:
- Conformidade do domínio selecionado (percentual)

**Visualizações**:
- Gráfico de barras de conformidade por framework
- Tabela detalhada de requisitos, exibindo status de conformidade e nível de risco

### 3.3 Dashboard de Impacto Econômico Open X

Foca na análise financeira das não-conformidades:

**Filtros**:
- Seleção de Tenant
- Seleção múltipla de Domínios Open X

**Métricas principais**:
- Impacto econômico total (valor monetário)
- Impacto médio por não-conformidade (valor monetário)

**Visualizações**:
- Gráfico de pizza da distribuição do impacto econômico por domínio
- Gráfico de barras de simulação de ROI para diferentes níveis de melhoria
- Tabela das não-conformidades com maior impacto econômico

### 3.4 Dashboard por Jurisdição Open X

Permite análise da conformidade segmentada por jurisdição:

**Filtros**:
- Seleção de Tenant
- Seleção de Jurisdição (Portugal/UE, Brasil, Angola, EUA)

**Métricas principais**:
- Conformidade geral na jurisdição selecionada (percentual)

**Visualizações**:
- Gráfico de barras de conformidade por domínio na jurisdição
- Tabela de frameworks regulatórios aplicáveis à jurisdição

## 4. Componentes Visuais

### 4.1 Cards

Os cards exibem métricas-chave com as seguintes características:

- **Design visual**: Ícones intuitivos e códigos de cores para status
- **Thresholds configuráveis**: Limites para aviso (warning) e crítico
- **Indicador de tendência**: Comparação com período anterior (quando disponível)
- **Formatação personalizada**: Configuração de exibição de valores (percentual, monetário, etc.)

### 4.2 Gráficos

A interface suporta os seguintes tipos de gráficos interativos:

- **Barras**: Comparação entre categorias
- **Linhas**: Análise de tendências ao longo do tempo
- **Pizza/Donut**: Distribuição proporcional
- **Mapa de calor**: Visualização de matrizes de risco

Todos os gráficos possuem:
- Tooltips detalhados
- Zoom e pan
- Exportação em diferentes formatos (PNG, PDF, CSV)
- Paletas de cores personalizáveis

### 4.3 Tabelas

As tabelas de dados incluem:

- **Paginação**: Navegação eficiente em grandes conjuntos de dados
- **Ordenação**: Por qualquer coluna, ascendente ou descendente
- **Filtragem**: Busca e filtros por coluna
- **Renderizadores personalizados**: Formatação visual para tipos específicos de dados
- **Exportação**: Para CSV, Excel ou PDF

### 4.4 Filtros

Os filtros disponíveis incluem:

- **Dropdown**: Seleção única ou múltipla
- **Date Range**: Seleção de período
- **Sliders**: Seleção de intervalos numéricos
- **Cascata**: Filtros interdependentes

## 5. Funções de Suporte

### 5.1 get_open_x_domain_validations

Obtém as validações específicas para um determinado domínio Open X:

```sql
SELECT * FROM compliance_validators.get_open_x_domain_validations(
    '00000000-0000-0000-0000-000000000000',  -- tenant_id
    'OPEN_INSURANCE'  -- domain
);
```

### 5.2 get_open_x_domain_requirements

Recupera os requisitos específicos para um determinado domínio Open X:

```sql
SELECT * FROM compliance_validators.get_open_x_domain_requirements(
    'OPEN_HEALTH'  -- domain
);
```

### 5.3 get_open_x_roi_simulations

Simula diferentes níveis de melhoria e calcula o ROI estimado:

```sql
SELECT * FROM compliance_validators.get_open_x_roi_simulations(
    '00000000-0000-0000-0000-000000000000',  -- tenant_id
    ARRAY['OPEN_INSURANCE', 'OPEN_HEALTH'],  -- domains (opcional)
    24  -- horizonte em meses
);
```

## 6. Casos de Uso

### 6.1 Análise de Dashboard Executivo

**Cenário**: Um executivo precisa de uma visão geral rápida da conformidade Open X.

**Passos**:
1. Acessar o Dashboard Principal Open X
2. Visualizar os cards de métricas principais
3. Identificar domínios críticos através do gráfico de conformidade por domínio
4. Verificar o impacto econômico total atual

**Resultado**: Compreensão imediata do status de conformidade e riscos associados.

### 6.2 Análise Detalhada por Analista de Conformidade

**Cenário**: Um analista de conformidade precisa investigar não-conformidades em um domínio específico.

**Passos**:
1. Acessar o Dashboard por Domínio Open X
2. Selecionar o domínio de interesse (ex. Open Insurance)
3. Analisar a conformidade por framework
4. Consultar a tabela detalhada de requisitos
5. Filtrar para visualizar apenas requisitos não conformes

**Resultado**: Identificação de requisitos específicos que precisam de ação corretiva.

### 6.3 Planejamento de Investimentos pelo Analista Econômico

**Cenário**: Um analista econômico precisa justificar investimentos em melhorias de conformidade.

**Passos**:
1. Acessar o Dashboard de Impacto Econômico Open X
2. Visualizar a distribuição do impacto econômico por domínio
3. Identificar não-conformidades com maior impacto financeiro
4. Analisar o gráfico de simulação de ROI para diferentes níveis de melhoria

**Resultado**: Caso de negócio fundamentado para priorização de investimentos.

### 6.4 Análise Regulatória Regional

**Cenário**: Um analista de risco precisa avaliar a conformidade específica para uma jurisdição.

**Passos**:
1. Acessar o Dashboard por Jurisdição Open X
2. Selecionar a jurisdição desejada (ex. Brasil)
3. Analisar a conformidade por domínio na jurisdição
4. Consultar a tabela de frameworks regulatórios específicos da jurisdição

**Resultado**: Compreensão dos riscos regulatórios específicos da região.

## 7. Requisitos Técnicos

### 7.1 Requisitos de Sistema

- **Banco de Dados**: PostgreSQL 13.0 ou superior
- **Navegadores Suportados**: Chrome 90+, Firefox 90+, Safari 14+, Edge 90+
- **Resolução mínima**: 1366x768 (recomendado: 1920x1080)

### 7.2 Dependências

- **Schemas**: 
  - `compliance_dashboard_ui`: Armazena definições de interface
  - `compliance_validators`: Fornece dados de validações
  - `economic_planning`: Fornece cálculos de impacto econômico
  - `iam`: Gerencia autenticação e autorização

- **Scripts Necessários**:
  - `17_open_x_validators.sql`: Validadores Open X
  - `18_open_x_dashboard.sql`: Views e funções do dashboard
  - `19_open_x_dashboard_interface.sql`: Definições da interface

### 7.3 Roles e Permissões

Os seguintes papéis possuem acesso à interface, com diferentes níveis de permissão:

- **economic_analyst_role**: Acesso completo aos dashboards econômicos e simulações
- **compliance_manager_role**: Acesso completo às métricas de conformidade
- **risk_analyst_role**: Acesso às análises de risco e jurisdição
- **dashboard_viewer_role**: Acesso somente leitura às visualizações

## 8. Configuração e Personalização

### 8.1 Preferências de Usuário

Os usuários podem personalizar:

- **Tema**: Claro (padrão) ou escuro
- **Layout**: Reorganizar componentes (arrastar/redimensionar)
- **Filtros padrão**: Definir valores iniciais para filtros
- **Intervalo de atualização**: Configurar frequência de refresh dos dados

### 8.2 Exportação de Dados

Todos os dashboards permitem:

- Exportação de dados em formato CSV/Excel
- Exportação de gráficos como imagens PNG
- Exportação de relatórios completos em PDF
- Agendamento de relatórios por e-mail

## 9. Considerações de Implementação

### 9.1 Responsividade

A interface foi projetada para se adaptar a diferentes tamanhos de tela:

- **Desktop**: Layout completo com todos os componentes
- **Tablet**: Layout ajustado com componentes reorganizados
- **Mobile**: Versão simplificada com foco nas métricas principais

### 9.2 Acessibilidade

A interface segue as diretrizes WCAG 2.1 nível AA:

- Contraste adequado para todos os elementos visuais
- Suporte a navegação por teclado
- Compatibilidade com leitores de tela
- Textos alternativos para elementos gráficos

### 9.3 Internacionalização

Suporte a múltiplos idiomas, com prioridade para:

- Português (padrão)
- Inglês
- Espanhol

### 9.4 Performance

Otimizações para garantir alto desempenho:

- Carregamento assíncrono de componentes
- Paginação para conjuntos grandes de dados
- Cache de consultas frequentes
- Compressão de dados transferidos

## 10. Próximos Passos

- 🚀 **Frontend Web Responsivo**: Implementação em framework JS moderno (React/Angular)
- 🚀 **Módulo de Alertas Inteligentes**: Notificações baseadas em thresholds e tendências
- 🚀 **Integração com BI Externo**: Conectores para Power BI, Tableau, etc.
- ⚙ **Dashboards Personalizáveis**: Editor drag-and-drop para criação de dashboards
- ⚙ **Analytics Avançado**: Previsões baseadas em machine learning

## 11. Referências

- [Material Design Guidelines](https://material.io/design)
- [WCAG 2.1 Accessibility Guidelines](https://www.w3.org/TR/WCAG21/)
- [PostgreSQL JSON Functions](https://www.postgresql.org/docs/current/functions-json.html)
- [Open Finance Brasil](https://openfinancebrasil.org.br/)
- [Open Insurance Brasil](https://openinsurance.com.br/)
