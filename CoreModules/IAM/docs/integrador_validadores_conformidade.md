# Documentação Técnica: Integrador de Validadores de Conformidade IAM

**Versão:** 1.0.0  
**Data:** 14/05/2025  
**Autor:** INNOVABIZ DevOps  
**Status:** ✅ Concluído

## 1. Introdução

Este documento descreve a implementação técnica do Framework Integrador de Validadores de Conformidade para o módulo de Identity and Access Management (IAM) da plataforma INNOVABIZ. Este framework unifica os validadores setoriais (Saúde, Financeiro, Governamental e AR/VR) e possibilita avaliações de conformidade multi-setoriais conforme exigido para soluções empresariais multi-regulatórias.

## 2. Arquitetura e Componentes

### 2.1 Visão Geral

O Framework Integrador de Validadores de Conformidade é composto por:

- **Schema Central**: `compliance_integrator` que contém os componentes de integração
- **Mapeamento de Recursos**: Tabelas de mapeamento de setores e regulações
- **Motor de Validação**: Funções que realizam validações multi-setoriais
- **Gerador de Relatórios**: Componentes para criar relatórios em diversos formatos
- **Sistema de Agendamento**: Funcionalidades para automação de validações periódicas

### 2.2 Estrutura de Dados

A implementação está baseada nas seguintes tabelas:

| Tabela | Descrição |
|--------|-----------|
| `compliance_integrator.sectors` | Mapeamento de setores com seus módulos validadores |
| `compliance_integrator.sector_regulations` | Mapeamento de regulações por setor |
| `compliance_integrator.tenant_validator_config` | Configuração de validação por tenant |
| `compliance_integrator.validation_history` | Histórico de validações executadas |

### 2.3 Principais Funções

| Função | Descrição |
|--------|-----------|
| `validate_sector_compliance` | Valida a conformidade de um setor específico |
| `validate_multi_sector_compliance` | Valida a conformidade de múltiplos setores |
| `calculate_multi_sector_score` | Calcula pontuação de conformidade consolidada |
| `generate_compliance_report_json` | Gera relatório em formato JSON |
| `export_compliance_report_xml` | Exporta relatório em formato XML |
| `export_compliance_report_csv` | Exporta relatório em formato CSV |
| `schedule_compliance_validation` | Configura validações agendadas |
| `run_scheduled_validations` | Executa validações agendadas |

## 3. Setores e Regulações Suportados

### 3.1 Setores Implementados

| ID do Setor | Nome do Setor | Descrição |
|-------------|---------------|-----------|
| HEALTHCARE | Saúde | Setor de saúde e telemedicina |
| FINANCIAL | Financeiro | Setor financeiro e bancário |
| GOVERNMENT | Governamental | Setor governamental e público |
| ARVR | Realidade Aumentada/Virtual | Setor de AR/VR e tecnologias imersivas |
| MULTI | Multi-Setorial | Validação aplicável a múltiplos setores |

### 3.2 Regulações por Setor

#### 3.2.1 Saúde
- HIPAA (EUA)
- GDPR_HEALTH (UE)
- LGPD_HEALTH (Brasil)

#### 3.2.2 Financeiro
- PSD2 (UE)
- OPEN_BANKING_BR (Brasil)
- OPEN_BANKING_UK (Reino Unido)

#### 3.2.3 Governamental
- EIDAS (UE)
- ICP_BRASIL (Brasil)
- ANG_EGOV (Angola)

#### 3.2.4 AR/VR
- IEEE_XR (Global)
- NIST_XR (EUA)
- OPENXR (Global)

## 4. Funcionalidades Principais

### 4.1 Validação Multi-Setorial

A validação multi-setorial permite avaliar a conformidade de um tenant em múltiplos setores simultaneamente, considerando:

- Seleção de setores específicos ou todos os setores disponíveis
- Filtro por regiões específicas (UE, Brasil, EUA, Angola, etc.)
- Integração com validadores setoriais específicos
- Execução paralela de validações

### 4.2 Pontuação de Conformidade

O sistema calcula pontuações de conformidade:

- Por setor individual
- Consolidada para todos os setores selecionados
- Com cálculo de percentual de conformidade
- Com determinação de IRR (Índice de Risco Residual)

A escala de IRR segue o seguinte padrão:

| Percentual | IRR | Interpretação |
|------------|-----|---------------|
| ≥ 95% | R1 | Risco residual muito baixo |
| ≥ 85% | R2 | Risco residual baixo |
| ≥ 70% | R3 | Risco residual moderado |
| < 70% | R4 | Risco residual elevado |

### 4.3 Geração de Relatórios

O framework oferece geração de relatórios em formatos variados:

- **JSON**: Para integração com sistemas e dashboards
- **XML**: Para compatibilidade com sistemas legados
- **CSV**: Para análise em ferramentas de planilha

Os relatórios incluem:
- Detalhes completos das validações
- Pontuações por setor e consolidadas
- Metadados de identificação do tenant e timestamp
- IRR calculado

### 4.4 Validações Agendadas

O sistema suporta agendamento de validações:

- Periodicidade configurável (diária, semanal, mensal, trimestral)
- Configuração por tenant
- Notificações por email (implementação simplificada)
- Histórico de validações executadas

## 5. Integração com Validadores Setoriais

### 5.1 Mecânica de Integração

O Framework Integrador conecta-se aos validadores setoriais através de:

1. **Mapeamento de Funções**: Cada regulação setorial tem uma função de validação mapeada
2. **Execução Dinâmica**: Chamadas dinâmicas às funções de validação através de SQL dinâmico
3. **Agregação de Resultados**: Consolidação dos resultados setoriais em uma visualização unificada

### 5.2 Fluxo de Validação

1. Cliente (tenant) solicita validação
2. Determinação dos setores a serem validados
3. Identificação das regulações aplicáveis por setor
4. Execução das funções de validação correspondentes
5. Agregação dos resultados em um relatório único
6. Cálculo de pontuações e determinação de IRR
7. Geração do relatório no formato solicitado

## 6. Instruções de Uso

### 6.1 Configuração Inicial

Para configurar um tenant para validação multi-setor:

```sql
-- Configurar setores ativos para um tenant
INSERT INTO compliance_integrator.tenant_validator_config (
    tenant_id, active_sectors
)
VALUES (
    'tenant-uuid-aqui', 
    ARRAY['HEALTHCARE', 'FINANCIAL']
);
```

### 6.2 Execução de Validação Ad-hoc

```sql
-- Validar conformidade para setores específicos
SELECT * FROM compliance_integrator.validate_multi_sector_compliance(
    'tenant-uuid-aqui',
    ARRAY['HEALTHCARE', 'FINANCIAL'],
    ARRAY['UE', 'Brasil']
);

-- Calcular pontuação de conformidade
SELECT * FROM compliance_integrator.calculate_multi_sector_score(
    'tenant-uuid-aqui',
    ARRAY['HEALTHCARE', 'FINANCIAL'],
    ARRAY['UE', 'Brasil']
);

-- Gerar relatório em JSON
SELECT compliance_integrator.generate_compliance_report_json(
    'tenant-uuid-aqui',
    ARRAY['HEALTHCARE', 'FINANCIAL'],
    ARRAY['UE', 'Brasil']
);
```

### 6.3 Configuração de Agendamento

```sql
-- Configurar validação agendada mensal
SELECT compliance_integrator.schedule_compliance_validation(
    'tenant-uuid-aqui',
    ARRAY['HEALTHCARE', 'FINANCIAL', 'GOVERNMENT'],
    ARRAY['UE', 'Brasil', 'Angola'],
    'MONTHLY',
    ARRAY['alerta@exemplo.com.br'],
    'CSV'
);
```

### 6.4 Execução de Validações Agendadas

```sql
-- Executar validações agendadas programadas para hoje
CALL compliance_integrator.run_scheduled_validations();
```

## 7. Considerações de Segurança

- **Execução Dinâmica Segura**: As funções que executam SQL dinâmico implementam validações de segurança
- **Isolamento por Tenant**: Todas as consultas são isoladas por tenant para evitar vazamento de dados
- **Registro de Auditoria**: Todas as validações são registradas na tabela de histórico
- **Permissões Granulares**: Recomenda-se configurar permissões específicas para acesso às funções do integrador

## 8. Limitações Conhecidas

- A implementação atual da exportação XML é simplificada e deve ser expandida em ambiente de produção
- Não há integração automática com sistemas de gestão de incidentes para IRRs elevados
- A notificação por email requer implementação adicional específica para o ambiente

## 9. Próximos Passos

- 🚀 Desenvolvimento de dashboard visual para relatórios de conformidade
- 🚀 Integração com sistema de gestão de incidentes para IRRs elevados (R3 e R4)
- ⚙ Implementação de notificações por webhook para sistemas externos
- ⚙ Expansão da exportação XML para conformidade com padrões XML específicos
- ⚙ Adição de novos setores e regulações conforme expansão da plataforma

## 10. Alinhamento com Frameworks

Este Framework Integrador de Validadores de Conformidade foi desenvolvido seguindo os princípios e recomendações de:

- TOGAF: Arquitetura modular e orientada a serviços
- DMBOK: Governança de dados com rastreabilidade e auditoria
- COBIT: Controles e medição de conformidade
- NIST Cybersecurity Framework: Identificação, proteção e detecção de riscos
- BIAN: Alinhamento com capacidades de negócio para o setor financeiro
- ISO/IEC 27001: Gestão de segurança da informação

## 11. Referências

- Documentação dos Validadores Setoriais:
  - [Validadores de Conformidade para Saúde](validadores_conformidade_saude.md)
  - [Validadores de Conformidade para Finanças](validadores_conformidade_financas.md)
  - [Validadores de Conformidade para Governo](validadores_conformidade_governo.md)
  - [Validadores de Conformidade para AR/VR](validadores_conformidade_ar_vr.md)
- Frameworks de Referência:
  - [TOGAF](https://www.opengroup.org/togaf)
  - [BIAN](https://bian.org/)
  - [DMBOK](https://www.dama.org/cpages/body-of-knowledge)
  - [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)
