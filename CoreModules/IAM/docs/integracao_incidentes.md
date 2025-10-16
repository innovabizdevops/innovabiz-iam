# Documentação Técnica: Integração de Validadores de Conformidade IAM com Gestão de Incidentes

**Versão:** 1.0.0  
**Data:** 14/05/2025  
**Autor:** INNOVABIZ DevOps  
**Status:** ✅ Concluído

## 1. Introdução

Este documento descreve a implementação técnica da integração entre os Validadores de Conformidade IAM e o Sistema de Gestão de Incidentes na plataforma INNOVABIZ. Esta integração permite a criação automática de incidentes a partir de não conformidades detectadas pelos validadores, garantindo um processo estruturado de resposta e resolução para problemas de conformidade.

## 2. Arquitetura e Componentes

### 2.1 Visão Geral

A integração é composta pelos seguintes elementos arquiteturais:

- **Schema de Integração**: `compliance_incident`, contendo tabelas e funções de integração
- **Motor de Processamento**: Funções e triggers que analisam resultados de validação e determinam a necessidade de incidentes
- **Mapeamento de Severidade**: Sistema de correlação entre IRR (Índice de Risco Residual) e severidade de incidentes
- **Configuração por Tenant**: Personalização das regras de integração por tenant
- **Monitoramento de SLA**: Cálculo e controle de SLAs por nível de incidente

### 2.2 Estrutura de Dados

A implementação utiliza as seguintes tabelas:

| Tabela | Descrição |
|--------|-----------|
| `compliance_incident.irr_severity_mapping` | Mapeamento de níveis IRR para severidades de incidentes |
| `compliance_incident.tenant_integration_config` | Configuração da integração por tenant |
| `compliance_incident.incident_history` | Histórico de incidentes criados a partir de validações |

### 2.3 Principais Funções

| Função | Descrição |
|--------|-----------|
| `create_incident_from_validation` | Cria um incidente a partir de um resultado de validação |
| `process_validation_results` | Analisa resultados de validação e decide sobre criação de incidentes |
| `validation_result_trigger` | Trigger para processamento automático de novos resultados |
| `configure_tenant_integration` | Configura a integração para um tenant específico |
| `get_active_incidents` | Consulta incidentes ativos com status de SLA |

## 3. Mapeamento de Severidade e SLA

### 3.1 Níveis de IRR para Severidade

A plataforma mapeia os níveis de IRR (Índice de Risco Residual) dos validadores para severidades e prioridades de incidentes:

| IRR | Severidade | Prioridade | SLA (horas) | Criação Automática | Descrição |
|-----|------------|------------|-------------|-------------------|-----------|
| R1 | BAIXA | 4 | 168 (7 dias) | Não | Risco residual muito baixo |
| R2 | MÉDIA | 3 | 72 (3 dias) | Não | Risco residual baixo |
| R3 | ALTA | 2 | 24 (1 dia) | Sim | Risco residual moderado |
| R4 | CRÍTICA | 1 | 4 (4 horas) | Sim | Risco residual elevado |

### 3.2 Status de SLA

O sistema calcula e monitora o status de SLA dos incidentes:

- **DENTRO_DO_SLA**: Tempo decorrido menor que 75% do SLA definido
- **EM_RISCO**: Tempo decorrido entre 75% e 100% do SLA definido
- **VIOLADO**: Tempo decorrido maior que o SLA definido
- **FECHADO**: Incidente resolvido

## 4. Fluxo de Integração

### 4.1 Criação de Incidentes

O processo de criação de incidentes segue o seguinte fluxo:

1. Uma validação de conformidade é executada pelo framework integrador
2. O resultado da validação é armazenado em `compliance_integrator.validation_history`
3. O trigger `validation_result_trigger` detecta a nova entrada ou atualização
4. A função `process_validation_results` analisa o resultado:
   - Verifica se já existe um incidente para esta validação
   - Determina o IRR do resultado
   - Consulta a configuração do tenant
   - Decide se é necessário criar um incidente
5. Se necessário, a função `create_incident_from_validation` é executada:
   - Gera um ID de incidente
   - Cria uma entrada em `compliance_incident.incident_history`
   - Formata os detalhes do incidente para inclusão no sistema
6. O incidente é registrado e monitorado até sua resolução

### 4.2 Configuração por Tenant

Cada tenant pode ter configurações específicas para a integração:

- **Habilitação da Integração**: Ativar ou desativar a integração
- **Grupo de Atribuição**: Equipe responsável pelos incidentes
- **Limiar de Criação Automática**: Nível de IRR a partir do qual incidentes são criados automaticamente
- **Configurações Adicionais**: Parâmetros específicos para cada tenant

## 5. Instruções de Uso

### 5.1 Configuração da Integração

Para configurar a integração para um tenant específico:

```sql
SELECT compliance_incident.configure_tenant_integration(
    'tenant-uuid-aqui',             -- ID do tenant
    TRUE,                           -- Habilitar integração
    'IAM_Compliance_Team',          -- Grupo de atribuição
    'R3',                           -- Limiar de criação (R1, R2, R3 ou R4)
    '{"notification_emails": ["compliance@example.com"]}'::JSONB -- Configurações adicionais
);
```

### 5.2 Consulta de Incidentes Ativos

Para consultar incidentes ativos de um tenant específico:

```sql
SELECT * FROM compliance_incident.get_active_incidents('tenant-uuid-aqui');
```

### 5.3 Processamento Manual de Validações

Para processar manualmente uma validação e criar um incidente, se necessário:

```sql
SELECT compliance_incident.process_validation_results('validacao-uuid-aqui');
```

## 6. Considerações de Segurança

- **Isolamento por Tenant**: Todas as operações são isoladas por tenant
- **Acesso Controlado**: O acesso às funções de integração deve ser restrito a usuários autorizados
- **Registro de Auditoria**: Todas as ações são registradas para auditoria
- **Controle de SLA**: Os SLAs são configurados conforme os requisitos de segurança e conformidade

## 7. Alinhamento com Regulamentações

A integração de Validadores de Conformidade com Gestão de Incidentes está alinhada com os seguintes frameworks:

- **ISO/IEC 27001**: Controles de segurança da informação
- **COBIT**: Processos de gerenciamento de incidentes
- **ITIL**: Práticas de gerenciamento de serviços
- **NIST Cybersecurity Framework**: Resposta a incidentes
- **PCI DSS**: Requisitos para tratamento de incidentes de segurança
- **GDPR/LGPD**: Tratamento de incidentes relacionados a dados pessoais

## 8. Próximos Passos

- 🚀 **Integração Expandida**: Implementação de APIs REST para comunicação bidirecional
- 🚀 **Dashboards de SLA**: Desenvolvimento de dashboards visuais para monitoramento de SLA
- ⚙ **Automação de Resoluções**: Implementação de resoluções automáticas para incidentes comuns
- ⚙ **Notificações Aprimoradas**: Expansão do sistema de notificações com suporte a SMS e webhooks
- ⚙ **Machine Learning**: Implementação de detecção de padrões para identificação de causas-raiz

## 9. Referências

- [Documentação do Sistema de Gestão de Incidentes](../../gestao_incidentes/index.md)
- [Documentação dos Validadores de Conformidade IAM](./index_validadores_conformidade.md)
- [ISO/IEC 27035](https://www.iso.org/standard/60803.html) - Gestão de Incidentes de Segurança da Informação
- [ITIL Incident Management](https://www.axelos.com/best-practice-solutions/itil)
- [COBIT 5 DSS03](https://www.isaca.org/resources/cobit) - Gerenciar Problemas
