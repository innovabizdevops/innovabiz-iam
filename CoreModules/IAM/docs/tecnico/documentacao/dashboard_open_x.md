# Dashboard Open X - Documentação Técnica

**Versão:** 1.0.0  
**Data:** 15/05/2025  
**Autor:** INNOVABIZ DevOps  
**Status:** ✅ Concluído

## 1. Visão Geral

O Dashboard Open X é uma solução integrada para visualização, análise e monitoramento de conformidade e impactos econômicos relacionados ao ecossistema Open X na plataforma INNOVABIZ. Este dashboard oferece uma visão consolidada dos diferentes domínios do Open X (Open Insurance, Open Health e Open Government), permitindo uma análise comparativa e a identificação de tendências, riscos e oportunidades de melhoria.

O dashboard integra-se aos validadores de conformidade específicos para o ecossistema Open X e ao módulo de modelagem econômica, fornecendo métricas detalhadas sobre o estado atual de conformidade e os impactos financeiros associados.

## 2. Arquitetura do Dashboard

### 2.1 Componentes Principais

O Dashboard Open X é composto pelos seguintes componentes principais:

1. **Views de Resumo e Agregação**:
   - `vw_open_x_compliance_summary`: Visão geral da conformidade por domínio e framework
   - `vw_open_x_non_compliance_details`: Detalhes das não-conformidades identificadas
   - `vw_open_x_economic_impact`: Impacto econômico por domínio e framework
   - `vw_open_x_domain_comparison`: Análise comparativa entre domínios
   - `vw_open_x_jurisdiction_compliance`: Conformidade por jurisdição

2. **Funções para KPIs e Análises**:
   - `get_open_x_compliance_metrics`: Métricas de conformidade por domínio
   - `compare_open_x_frameworks`: Comparação entre frameworks regulatórios
   - `get_open_x_economic_impact_by_jurisdiction`: Impacto econômico por jurisdição
   - `simulate_open_x_improvements`: Simulação de melhorias na conformidade

### 2.2 Integração com Outros Módulos

O Dashboard Open X se integra com:

- **Validadores de Conformidade Open X**: Utiliza os resultados das validações de conformidade para todos os domínios Open X.
- **Sistema de Modelagem Econômica**: Calcula os impactos econômicos associados às não-conformidades.
- **Dashboards Existentes**: Complementa os dashboards existentes, fornecendo uma visão específica para o ecossistema Open X.

## 3. Visualizações Disponíveis

### 3.1 Resumo de Conformidade Open X

Esta visualização apresenta um panorama geral da conformidade para todo o ecossistema Open X, incluindo:

- **Pontuação de conformidade** por domínio e framework
- **Número total de requisitos** e quantos estão em conformidade
- **Percentual de conformidade**
- **Nível de risco** associado (R1, R2, R3 ou R4)
- **Descrição do risco** (BAIXO, MODERADO, ALTO, CRÍTICO)

### 3.2 Detalhes de Não-Conformidades

Apresenta informações detalhadas sobre cada não-conformidade identificada:

- **Domínio e framework** afetados
- **ID e nome do requisito** não conforme
- **Detalhes da não-conformidade**
- **Nível de risco residual** (IRR)

### 3.3 Impacto Econômico por Domínio

Esta visualização mostra o impacto econômico das não-conformidades:

- **Total de requisitos** e quantos estão não conformes
- **Percentual de não-conformidade**
- **Impacto monetário total**
- **Impacto médio por não-conformidade**

### 3.4 Comparação entre Domínios Open X

Permite comparar os diferentes domínios do ecossistema Open X:

- **Requisitos conformes e não conformes** por domínio
- **Percentual de conformidade** por domínio
- **Nível de risco** associado a cada domínio

### 3.5 Conformidade por Jurisdição

Apresenta a conformidade segregada por jurisdição:

- **Métricas de conformidade** por jurisdição e domínio
- **Frameworks regulatórios** aplicáveis a cada jurisdição
- **Níveis de risco** por jurisdição

## 4. KPIs e Análises Disponíveis

### 4.1 Métricas de Conformidade por Domínio

Função: `get_open_x_compliance_metrics`

Fornece métricas detalhadas de conformidade para cada domínio Open X:

- Total de requisitos e requisitos conformes
- Percentual de conformidade
- Nível e descrição de risco
- Impacto econômico estimado

### 4.2 Comparação entre Frameworks Regulatórios

Função: `compare_open_x_frameworks`

Permite comparar diferentes frameworks regulatórios dentro de um mesmo domínio:

- Métricas de conformidade por framework
- Impacto econômico estimado por framework
- Classificação de risco por framework

### 4.3 Impacto Econômico por Jurisdição

Função: `get_open_x_economic_impact_by_jurisdiction`

Analisa o impacto econômico das não-conformidades por jurisdição:

- Número de requisitos não conformes por jurisdição
- Impacto econômico estimado por jurisdição
- Impacto médio por requisito não conforme
- Nível de risco baseado no impacto econômico

### 4.4 Simulação de Melhorias

Função: `simulate_open_x_improvements`

Permite simular o impacto de melhorias na conformidade:

- Percentual atual vs. simulado de conformidade
- Impacto econômico atual vs. simulado
- Economia potencial
- ROI estimado para as melhorias

## 5. Casos de Uso

### 5.1 Análise Estratégica de Conformidade

**Cenário**: Análise comparativa da conformidade entre diferentes domínios Open X para priorização de investimentos.

**Passos**:
1. Visualizar o resumo de conformidade Open X
2. Comparar os domínios utilizando `vw_open_x_domain_comparison`
3. Avaliar o impacto econômico com `vw_open_x_economic_impact`
4. Utilizar `simulate_open_x_improvements` para diferentes cenários

**Resultado**: Identificação dos domínios com maior risco e maior potencial de retorno sobre investimento em melhorias.

### 5.2 Conformidade por Região

**Cenário**: Avaliar a conformidade e riscos específicos para diferentes jurisdições.

**Passos**:
1. Utilizar `vw_open_x_jurisdiction_compliance` para visualizar o panorama por jurisdição
2. Executar `get_open_x_economic_impact_by_jurisdiction` para analisar impactos econômicos regionais
3. Comparar frameworks regulatórios específicos de cada jurisdição

**Resultado**: Identificação de riscos específicos por região e estratégias de mitigação adequadas a cada contexto regulatório.

### 5.3 Simulação de ROI em Melhorias

**Cenário**: Justificar investimentos em melhorias de conformidade com base no retorno financeiro.

**Passos**:
1. Identificar domínios e frameworks com maior impacto econômico
2. Utilizar `simulate_open_x_improvements` com diferentes percentuais de melhoria
3. Comparar custo estimado de remediação com economia potencial

**Resultado**: Caso de negócio fundamentado para priorização de investimentos em conformidade com melhor relação custo-benefício.

## 6. Requisitos Técnicos

### 6.1 Requisitos de Banco de Dados

- PostgreSQL 13.0 ou superior
- Schema `compliance_validators` criado e configurado
- Schema `economic_planning` configurado para cálculos de impacto econômico
- Validadores Open X implementados (script `17_open_x_validators.sql`)

### 6.2 Dependências

- Validadores de conformidade Open X
- Sistema de modelagem econômica
- Funções de cálculo de impacto econômico

### 6.3 Roles e Permissões

O dashboard concede permissões específicas para os seguintes perfis:

- **economic_analyst_role**: Acesso completo às análises econômicas e simulações
- **compliance_manager_role**: Acesso às métricas de conformidade e relatórios
- **risk_analyst_role**: Acesso às análises de risco
- **dashboard_viewer_role**: Acesso somente leitura às visualizações principais

## 7. Exemplos de Uso

### 7.1 Análise de Conformidade por Domínio

```sql
-- Obter métricas de conformidade para todos os domínios Open X
SELECT * FROM compliance_validators.get_open_x_compliance_metrics(
    '00000000-0000-0000-0000-000000000000'  -- tenant_id
);

-- Obter métricas específicas para Open Insurance
SELECT * FROM compliance_validators.get_open_x_compliance_metrics(
    '00000000-0000-0000-0000-000000000000',  -- tenant_id
    'OPEN_INSURANCE'
);
```

### 7.2 Análise de Impacto Econômico por Jurisdição

```sql
-- Obter impacto econômico para todas as jurisdições
SELECT * FROM compliance_validators.get_open_x_economic_impact_by_jurisdiction(
    '00000000-0000-0000-0000-000000000000'  -- tenant_id
);

-- Obter impacto econômico específico para o Brasil
SELECT * FROM compliance_validators.get_open_x_economic_impact_by_jurisdiction(
    '00000000-0000-0000-0000-000000000000',  -- tenant_id
    'BRASIL'
);
```

### 7.3 Simulação de Melhorias

```sql
-- Simular 30% de melhoria em todos os domínios com horizonte de 24 meses
SELECT * FROM compliance_validators.simulate_open_x_improvements(
    '00000000-0000-0000-0000-000000000000',  -- tenant_id
    NULL,  -- todos os domínios
    30,    -- percentual de melhoria
    24     -- meses de horizonte
);
```

## 8. Considerações de Implementação

### 8.1 Configurações Específicas por Região

O dashboard suporta configurações específicas para as regiões prioritárias:

- **Portugal/UE**: Foco em frameworks como Solvência II e eIDAS
- **Brasil**: Foco em frameworks como SUSEP, ANS e Gov.br
- **Angola**: Adaptações para o contexto regulatório angolano
- **EUA**: Suporte para HIPAA e outros regulamentos aplicáveis

### 8.2 Adaptações para Múltiplos Tenants

O dashboard está projetado para ambientes multi-tenant, permitindo:

- Análises isoladas por tenant
- Comparações entre tenants para benchmarking (mediante permissões apropriadas)
- Configurações específicas de thresholds por tenant

## 9. Próximos Passos

- 🚀 **Desenvolvimento de Interface Gráfica**: Implementação das visualizações em frontend com gráficos interativos
- 🚀 **Integração com Sistema de Alertas**: Notificações automáticas para riscos críticos identificados
- 🚀 **Expansão para Novos Domínios**: Incorporação de novos domínios Open X à medida que surgirem
- ⚙ **Refinamento dos Modelos de Impacto Econômico**: Calibração dos parâmetros com dados reais
- ⚙ **Personalização de KPIs**: Permitir configurações específicas de KPIs por tenant e domínio

## 10. Referências

- [Open Finance Brasil](https://openfinancebrasil.org.br/)
- [Open Insurance Brasil](https://openinsurance.com.br/)
- [Solvência II](https://eur-lex.europa.eu/legal-content/PT/TXT/?uri=CELEX:32009L0138)
- [eIDAS](https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=uriserv:OJ.L_.2014.257.01.0073.01.ENG)
- [HIPAA](https://www.hhs.gov/hipaa/)
- [SUSEP](https://www.gov.br/susep/)
- [ANS](https://www.gov.br/ans/)
