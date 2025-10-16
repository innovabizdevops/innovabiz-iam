# Plano de Testes: MCP-IAM Elevation Hooks

**Documento**: INNOVABIZ-IAM-TEST-MCP-HOOKS-v1.0.0  
**Classificação**: Confidencial-Interno  
**Data**: 06/08/2025  
**Estado**: Aprovado  
**Âmbito**: Multi-Mercado, Multi-Tenant  
**Elaborado por**: Equipa de Qualidade INNOVABIZ

## Índice

1. [Estratégia de Testes](#estratégia-de-testes)
2. [Escopo de Testes](#escopo-de-testes)
3. [Testes Unitários](#testes-unitários)
4. [Testes de Integração](#testes-de-integração)
5. [Testes de Conformidade Regulatória](#testes-de-conformidade-regulatória)
6. [Testes de Desempenho](#testes-de-desempenho)
7. [Testes de Segurança](#testes-de-segurança)
8. [Testes de Observabilidade](#testes-de-observabilidade)
9. [Matriz de Teste Multi-Mercado](#matriz-de-teste-multi-mercado)
10. [Automação de Testes](#automação-de-testes)
11. [Ambiente de Testes](#ambiente-de-testes)
12. [Critérios de Aceitação](#critérios-de-aceitação)
13. [Anexos](#anexos)

## Estratégia de Testes

### Visão Geral

A estratégia de teste para os hooks MCP-IAM segue uma abordagem multi-camada e multi-dimensional, garantindo a validação completa do comportamento em diversos contextos operacionais, regulatórios e de mercado. O enfoque principal é na validação da integração dos hooks com o sistema de elevação de privilégios, a correta aplicação de políticas específicas por mercado e a conformidade com regulamentações locais e internacionais.

### Princípios Orientadores

1. **Teste Baseado em Risco**: Priorização dos testes com base na criticidade das funcionalidades e potencial impacto em caso de falhas
2. **Automação Abrangente**: Automação de todos os testes repetitivos para garantir execução consistente e completa
3. **Shift-Left Testing**: Integração de testes desde as fases iniciais do desenvolvimento
4. **Cobertura Multi-Dimensional**: Validação em todas as dimensões (mercado, tenant, módulo, contexto)
5. **Validação Regulatória**: Verificação explícita de requisitos regulatórios específicos por mercado
6. **Testes de Regressão Contínuos**: Garantia de que novas funcionalidades não afetam comportamentos existentes

### Níveis de Teste

| Nível | Objetivo | Responsável | Ferramentas |
|-------|----------|-------------|------------|
| Unitário | Validar componentes individuais | Desenvolvedores | Go Testing, Testify, Mockery |
| Integração | Validar interoperabilidade entre componentes | Equipa de QA | Postman, GoConvey, Testcontainers |
| Sistema | Validar o comportamento do sistema completo | Equipa de QA | k6, Selenium, Cypress |
| Aceitação | Validar requisitos de negócio | Product Owner | Cucumber, BDD Scenarios |
| Conformidade | Validar requisitos regulatórios | Equipa de Compliance | Ferramentas especializadas por regulamento |
| Segurança | Validar proteções de segurança | Equipa de Segurança | OWASP ZAP, SonarQube, Gosec |
| Desempenho | Validar comportamento sob carga | Equipa de Performance | k6, Grafana, Locust |

## Escopo de Testes

### Componentes em Teste

1. **Hooks MCP-IAM**
   - Docker Hook
   - GitHub Hook
   - Desktop Commander Hook
   - Figma Hook
   - Hook Registry
   
2. **Serviço de Elevação**
   - API de Solicitação
   - Sistema de Aprovação
   - Gestor de Tokens
   - Validação de Uso
   
3. **Integrações**
   - Integrações com API Gateway
   - Integrações com Sistemas de Identidade
   - Integrações com Sistemas de MFA
   - Integrações com Sistemas de Auditoria
   - Integrações com OpenTelemetry
   
4. **Configurações Específicas**
   - Configurações por Mercado
   - Configurações por Tenant
   - Configurações Temporais
   - Configurações por Módulo

### Funcionalidades em Teste

| Funcionalidade | Prioridade | Nível de Risco | Complexidade |
|---------------|------------|----------------|--------------|
| Validação de Escopos | Alta | Alto | Média |
| Requisitos de MFA | Alta | Alto | Média |
| Fluxos de Aprovação | Alta | Alto | Alta |
| Geração de Tokens | Alta | Alto | Média |
| Validação de Uso | Alta | Alto | Alta |
| Geração de Auditoria | Alta | Médio | Alta |
| Configurações Específicas por Mercado | Alta | Alto | Alta |
| Operações em Modo de Emergência | Média | Alto | Alta |
| Observabilidade | Média | Médio | Média |
| Limites de Política | Média | Médio | Baixa |
| Expiração de Tokens | Média | Médio | Baixa |

### Itens Fora do Escopo

- Testes de UI para aplicações clientes
- Testes das implementações internas dos MCP Servers
- Validação de infraestrutura cloud (coberta por testes de infraestrutura separados)
- Testes de recuperação de desastre (cobertos por planos de DR específicos)

## Testes Unitários

### 1. Docker Hook

#### 1.1 Validação de Escopos

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| DH-US-001 | Validar escopo docker:run | Hook inicializado | 1. Chamar ValidateScope com "docker:run" | Retorna detalhes do escopo sem erro |
| DH-US-002 | Validar escopo docker:build | Hook inicializado | 1. Chamar ValidateScope com "docker:build" | Retorna detalhes do escopo sem erro |
| DH-US-003 | Validar escopo docker:admin | Hook inicializado | 1. Chamar ValidateScope com "docker:admin" | Retorna detalhes do escopo sem erro |
| DH-US-004 | Validar escopo inválido | Hook inicializado | 1. Chamar ValidateScope com "docker:invalid" | Retorna erro de escopo inválido |
| DH-US-005 | Validar contexto cancelado | Hook inicializado, contexto cancelado | 1. Chamar ValidateScope com contexto cancelado | Retorna erro de contexto cancelado |

#### 1.2 Requisitos de MFA

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| DH-MFA-001 | Validar MFA para docker:run | Hook inicializado | 1. Chamar RequiresMFA para "docker:run" | Retorna true |
| DH-MFA-002 | Validar MFA para docker:build | Hook inicializado | 1. Chamar RequiresMFA para "docker:build" | Retorna true |
| DH-MFA-003 | Validar MFA para docker:admin | Hook inicializado | 1. Chamar RequiresMFA para "docker:admin" | Retorna true |
| DH-MFA-004 | Validar MFA para escopo inválido | Hook inicializado | 1. Chamar RequiresMFA para "docker:invalid" | Retorna erro |
| DH-MFA-005 | Validar adaptação MFA por mercado | Hook inicializado com config específica para Angola | 1. Chamar RequiresMFA para "docker:run" com mercado "angola" | Retorna true e nível MFA forte |

#### 1.3 Requisitos de Aprovação

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| DH-APP-001 | Validar aprovação para docker:run | Hook inicializado | 1. Chamar RequiresApproval para "docker:run" | Retorna false |
| DH-APP-002 | Validar aprovação para docker:build | Hook inicializado | 1. Chamar RequiresApproval para "docker:build" | Retorna false |
| DH-APP-003 | Validar aprovação para docker:admin | Hook inicializado | 1. Chamar RequiresApproval para "docker:admin" | Retorna true |
| DH-APP-004 | Validar aprovação para escopo inválido | Hook inicializado | 1. Chamar RequiresApproval para "docker:invalid" | Retorna erro |
| DH-APP-005 | Validar adaptação aprovação por mercado | Hook inicializado com config específica para Angola | 1. Chamar RequiresApproval para "docker:run" com mercado "angola" | Retorna true para mercado específico |

#### 1.4 Validação de Uso

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| DH-VU-001 | Validar uso para docker:run | Hook inicializado, metadata válido | 1. Chamar ValidateTokenUse para "docker:run" com metadata de container válido | Validação bem-sucedida |
| DH-VU-002 | Validar uso para docker:build | Hook inicializado, metadata válido | 1. Chamar ValidateTokenUse para "docker:build" com metadata de Dockerfile válido | Validação bem-sucedida |
| DH-VU-003 | Validar uso para docker:admin | Hook inicializado, metadata válido | 1. Chamar ValidateTokenUse para "docker:admin" com metadata de operação admin válida | Validação bem-sucedida |
| DH-VU-004 | Rejeitar metadata inválido | Hook inicializado, metadata inválido | 1. Chamar ValidateTokenUse para "docker:run" com metadata incompleto | Retorna erro de metadata inválido |
| DH-VU-005 | Rejeitar operação proibida | Hook inicializado, metadata para operação proibida | 1. Chamar ValidateTokenUse para "docker:run" com metadata para container privilegiado | Retorna erro de operação proibida |
| DH-VU-006 | Validar adaptação por mercado | Hook inicializado, config específica para Angola | 1. Chamar ValidateTokenUse para "docker:run" com metadata específico para Angola | Validação adaptada para regras de mercado |

#### 1.5 Metadados de Auditoria

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| DH-AM-001 | Gerar metadados para docker:run | Hook inicializado, request válido | 1. Chamar GenerateAuditMetadata para "docker:run" com request válido | Metadados completos para auditoria |
| DH-AM-002 | Gerar metadados para docker:build | Hook inicializado, request válido | 1. Chamar GenerateAuditMetadata para "docker:build" com request válido | Metadados completos para auditoria |
| DH-AM-003 | Gerar metadados para docker:admin | Hook inicializado, request válido | 1. Chamar GenerateAuditMetadata para "docker:admin" com request válido | Metadados completos para auditoria |
| DH-AM-004 | Validar metadados específicos por mercado | Hook inicializado, config específica para Angola | 1. Chamar GenerateAuditMetadata para "docker:run" com mercado "angola" | Metadados com campos específicos para Angola |
| DH-AM-005 | Validar campos obrigatórios | Hook inicializado | 1. Chamar GenerateAuditMetadata para qualquer escopo | Metadados contendo todos os campos obrigatórios |

### 2. GitHub Hook

#### 2.1 Validação de Escopos

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| GH-US-001 | Validar escopo github:repo | Hook inicializado | 1. Chamar ValidateScope com "github:repo" | Retorna detalhes do escopo sem erro |
| GH-US-002 | Validar escopo github:admin | Hook inicializado | 1. Chamar ValidateScope com "github:admin" | Retorna detalhes do escopo sem erro |
| GH-US-003 | Validar escopo github:secrets | Hook inicializado | 1. Chamar ValidateScope com "github:secrets" | Retorna detalhes do escopo sem erro |
| GH-US-004 | Validar escopo inválido | Hook inicializado | 1. Chamar ValidateScope com "github:invalid" | Retorna erro de escopo inválido |
| GH-US-005 | Validar contexto cancelado | Hook inicializado, contexto cancelado | 1. Chamar ValidateScope com contexto cancelado | Retorna erro de contexto cancelado |

#### 2.2 Requisitos de MFA

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| GH-MFA-001 | Validar MFA para github:repo | Hook inicializado | 1. Chamar RequiresMFA para "github:repo" | Retorna false |
| GH-MFA-002 | Validar MFA para github:admin | Hook inicializado | 1. Chamar RequiresMFA para "github:admin" | Retorna true |
| GH-MFA-003 | Validar MFA para github:secrets | Hook inicializado | 1. Chamar RequiresMFA para "github:secrets" | Retorna true |
| GH-MFA-004 | Validar MFA para escopo inválido | Hook inicializado | 1. Chamar RequiresMFA para "github:invalid" | Retorna erro |
| GH-MFA-005 | Validar adaptação MFA por mercado | Hook inicializado com config específica para Brasil | 1. Chamar RequiresMFA para "github:repo" com mercado "brasil" | Retorna true por configuração específica de mercado |

#### 2.3 Requisitos de Aprovação

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| GH-APP-001 | Validar aprovação para github:repo | Hook inicializado | 1. Chamar RequiresApproval para "github:repo" | Retorna false |
| GH-APP-002 | Validar aprovação para github:admin | Hook inicializado | 1. Chamar RequiresApproval para "github:admin" | Retorna true |
| GH-APP-003 | Validar aprovação para github:secrets | Hook inicializado | 1. Chamar RequiresApproval para "github:secrets" | Retorna true |
| GH-APP-004 | Validar aprovação para escopo inválido | Hook inicializado | 1. Chamar RequiresApproval para "github:invalid" | Retorna erro |
| GH-APP-005 | Validar adaptação aprovação por mercado | Hook inicializado com config específica para Brasil | 1. Chamar RequiresApproval para "github:repo" com mercado "brasil" | Retorna true para configuração específica |

#### 2.4 Validação de Uso

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| GH-VU-001 | Validar uso para github:repo | Hook inicializado, metadata válido | 1. Chamar ValidateTokenUse para "github:repo" com metadata de repositório válido | Validação bem-sucedida |
| GH-VU-002 | Validar uso para github:admin | Hook inicializado, metadata válido | 1. Chamar ValidateTokenUse para "github:admin" com metadata de operação admin válida | Validação bem-sucedida |
| GH-VU-003 | Validar uso para github:secrets | Hook inicializado, metadata válido | 1. Chamar ValidateTokenUse para "github:secrets" com metadata de segredo válido | Validação bem-sucedida |
| GH-VU-004 | Rejeitar metadata inválido | Hook inicializado, metadata inválido | 1. Chamar ValidateTokenUse para "github:repo" com metadata incompleto | Retorna erro de metadata inválido |
| GH-VU-005 | Rejeitar operação proibida | Hook inicializado, metadata para operação proibida | 1. Chamar ValidateTokenUse para "github:repo" com metadata para repositório protegido | Retorna erro de operação proibida |
| GH-VU-006 | Validar adaptação por mercado | Hook inicializado, config específica para Brasil | 1. Chamar ValidateTokenUse para "github:repo" com metadata específico para Brasil | Validação adaptada para regras de mercado |
#### 2.5 Metadados de Auditoria

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| GH-AM-001 | Gerar metadados para github:repo | Hook inicializado, request válido | 1. Chamar GenerateAuditMetadata para "github:repo" com request válido | Metadados completos para auditoria |
| GH-AM-002 | Gerar metadados para github:admin | Hook inicializado, request válido | 1. Chamar GenerateAuditMetadata para "github:admin" com request válido | Metadados completos para auditoria |
| GH-AM-003 | Gerar metadados para github:secrets | Hook inicializado, request válido | 1. Chamar GenerateAuditMetadata para "github:secrets" com request válido | Metadados completos para auditoria |
| GH-AM-004 | Validar metadados específicos por mercado | Hook inicializado, config específica para Brasil | 1. Chamar GenerateAuditMetadata para "github:admin" com mercado "brasil" | Metadados com campos específicos para Brasil |
| GH-AM-005 | Validar campos obrigatórios | Hook inicializado | 1. Chamar GenerateAuditMetadata para qualquer escopo | Metadados contendo todos os campos obrigatórios |

### 3. Desktop Commander Hook

#### 3.1 Validação de Escopos

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| DC-US-001 | Validar escopo desktop:file | Hook inicializado | 1. Chamar ValidateScope com "desktop:file" | Retorna detalhes do escopo sem erro |
| DC-US-002 | Validar escopo desktop:process | Hook inicializado | 1. Chamar ValidateScope com "desktop:process" | Retorna detalhes do escopo sem erro |
| DC-US-003 | Validar escopo desktop:admin | Hook inicializado | 1. Chamar ValidateScope com "desktop:admin" | Retorna detalhes do escopo sem erro |
| DC-US-004 | Validar escopo desktop:config | Hook inicializado | 1. Chamar ValidateScope com "desktop:config" | Retorna detalhes do escopo sem erro |
| DC-US-005 | Validar escopo inválido | Hook inicializado | 1. Chamar ValidateScope com "desktop:invalid" | Retorna erro de escopo inválido |

#### 3.2 Requisitos de MFA

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| DC-MFA-001 | Validar MFA para desktop:file | Hook inicializado | 1. Chamar RequiresMFA para "desktop:file" | Retorna false |
| DC-MFA-002 | Validar MFA para desktop:process | Hook inicializado | 1. Chamar RequiresMFA para "desktop:process" | Retorna true |
| DC-MFA-003 | Validar MFA para desktop:admin | Hook inicializado | 1. Chamar RequiresMFA para "desktop:admin" | Retorna true |
| DC-MFA-004 | Validar MFA para desktop:config | Hook inicializado | 1. Chamar RequiresMFA para "desktop:config" | Retorna true |
| DC-MFA-005 | Validar adaptação MFA por mercado | Hook inicializado com config específica para Angola | 1. Chamar RequiresMFA para "desktop:file" com mercado "angola" | Retorna true por configuração específica de mercado |

#### 3.3 Requisitos de Aprovação

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| DC-APP-001 | Validar aprovação para desktop:file | Hook inicializado | 1. Chamar RequiresApproval para "desktop:file" | Retorna false |
| DC-APP-002 | Validar aprovação para desktop:process | Hook inicializado | 1. Chamar RequiresApproval para "desktop:process" | Retorna false |
| DC-APP-003 | Validar aprovação para desktop:admin | Hook inicializado | 1. Chamar RequiresApproval para "desktop:admin" | Retorna true |
| DC-APP-004 | Validar aprovação para desktop:config | Hook inicializado | 1. Chamar RequiresApproval para "desktop:config" | Retorna true |
| DC-APP-005 | Validar adaptação aprovação por mercado | Hook inicializado com config específica para Moçambique | 1. Chamar RequiresApproval para "desktop:file" com mercado "mozambique" | Retorna true para configuração específica |

#### 3.4 Validação de Uso

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| DC-VU-001 | Validar uso para desktop:file | Hook inicializado, metadata válido | 1. Chamar ValidateTokenUse para "desktop:file" com metadata de arquivo válido | Validação bem-sucedida |
| DC-VU-002 | Validar uso para desktop:process | Hook inicializado, metadata válido | 1. Chamar ValidateTokenUse para "desktop:process" com metadata de processo válido | Validação bem-sucedida |
| DC-VU-003 | Validar uso para desktop:admin | Hook inicializado, metadata válido | 1. Chamar ValidateTokenUse para "desktop:admin" com metadata de operação admin válida | Validação bem-sucedida |
| DC-VU-004 | Rejeitar metadata inválido | Hook inicializado, metadata inválido | 1. Chamar ValidateTokenUse para "desktop:file" com metadata incompleto | Retorna erro de metadata inválido |
| DC-VU-005 | Rejeitar operação proibida | Hook inicializado, metadata para operação proibida | 1. Chamar ValidateTokenUse para "desktop:file" com metadata para arquivo protegido | Retorna erro de operação proibida |
| DC-VU-006 | Validar acesso a diretórios sensíveis | Hook inicializado, metadata para diretório sensível | 1. Chamar ValidateTokenUse para "desktop:file" com metadata para diretório sensível | Retorna erro específico para diretório sensível |

#### 3.5 Metadados de Auditoria

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| DC-AM-001 | Gerar metadados para desktop:file | Hook inicializado, request válido | 1. Chamar GenerateAuditMetadata para "desktop:file" com request válido | Metadados completos para auditoria |
| DC-AM-002 | Gerar metadados para desktop:process | Hook inicializado, request válido | 1. Chamar GenerateAuditMetadata para "desktop:process" com request válido | Metadados completos para auditoria |
| DC-AM-003 | Gerar metadados para desktop:admin | Hook inicializado, request válido | 1. Chamar GenerateAuditMetadata para "desktop:admin" com request válido | Metadados completos para auditoria |
| DC-AM-004 | Validar metadados específicos por mercado | Hook inicializado, config específica para Moçambique | 1. Chamar GenerateAuditMetadata para "desktop:file" com mercado "mozambique" | Metadados com campos específicos para Moçambique |
| DC-AM-005 | Validar metadados de comandos executados | Hook inicializado, metadata com comandos | 1. Chamar GenerateAuditMetadata para "desktop:process" com comandos no metadata | Metadados incluindo informações detalhadas dos comandos |

### 4. Figma Hook

#### 4.1 Validação de Escopos

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| FH-US-001 | Validar escopo figma:view | Hook inicializado | 1. Chamar ValidateScope com "figma:view" | Retorna detalhes do escopo sem erro |
| FH-US-002 | Validar escopo figma:edit | Hook inicializado | 1. Chamar ValidateScope com "figma:edit" | Retorna detalhes do escopo sem erro |
| FH-US-003 | Validar escopo figma:admin | Hook inicializado | 1. Chamar ValidateScope com "figma:admin" | Retorna detalhes do escopo sem erro |
| FH-US-004 | Validar escopo figma:comment | Hook inicializado | 1. Chamar ValidateScope com "figma:comment" | Retorna detalhes do escopo sem erro |
| FH-US-005 | Validar escopo figma:export | Hook inicializado | 1. Chamar ValidateScope com "figma:export" | Retorna detalhes do escopo sem erro |
| FH-US-006 | Validar escopo figma:library | Hook inicializado | 1. Chamar ValidateScope com "figma:library" | Retorna detalhes do escopo sem erro |
| FH-US-007 | Validar escopo figma:team | Hook inicializado | 1. Chamar ValidateScope com "figma:team" | Retorna detalhes do escopo sem erro |
| FH-US-008 | Validar escopo inválido | Hook inicializado | 1. Chamar ValidateScope com "figma:invalid" | Retorna erro de escopo inválido |

#### 4.2 Requisitos de MFA

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| FH-MFA-001 | Validar MFA para figma:view | Hook inicializado | 1. Chamar RequiresMFA para "figma:view" | Retorna false |
| FH-MFA-002 | Validar MFA para figma:edit | Hook inicializado | 1. Chamar RequiresMFA para "figma:edit" | Retorna false |
| FH-MFA-003 | Validar MFA para figma:admin | Hook inicializado | 1. Chamar RequiresMFA para "figma:admin" | Retorna true |
| FH-MFA-004 | Validar MFA para figma:comment | Hook inicializado | 1. Chamar RequiresMFA para "figma:comment" | Retorna false |
| FH-MFA-005 | Validar MFA para figma:export | Hook inicializado | 1. Chamar RequiresMFA para "figma:export" | Retorna true |
| FH-MFA-006 | Validar MFA para figma:library | Hook inicializado | 1. Chamar RequiresMFA para "figma:library" | Retorna true |
| FH-MFA-007 | Validar MFA para figma:team | Hook inicializado | 1. Chamar RequiresMFA para "figma:team" | Retorna true |
| FH-MFA-008 | Validar adaptação MFA por mercado | Hook inicializado com config específica para Brasil | 1. Chamar RequiresMFA para "figma:edit" com mercado "brasil" | Retorna true por configuração específica de mercado |

#### 4.3 Requisitos de Aprovação

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| FH-APP-001 | Validar aprovação para figma:view | Hook inicializado | 1. Chamar RequiresApproval para "figma:view" | Retorna false |
| FH-APP-002 | Validar aprovação para figma:edit | Hook inicializado | 1. Chamar RequiresApproval para "figma:edit" | Retorna false |
| FH-APP-003 | Validar aprovação para figma:admin | Hook inicializado | 1. Chamar RequiresApproval para "figma:admin" | Retorna true |
| FH-APP-004 | Validar aprovação para figma:comment | Hook inicializado | 1. Chamar RequiresApproval para "figma:comment" | Retorna false |
| FH-APP-005 | Validar aprovação para figma:export | Hook inicializado | 1. Chamar RequiresApproval para "figma:export" | Retorna false |
| FH-APP-006 | Validar aprovação para figma:library | Hook inicializado | 1. Chamar RequiresApproval para "figma:library" | Retorna true |
| FH-APP-007 | Validar aprovação para figma:team | Hook inicializado | 1. Chamar RequiresApproval para "figma:team" | Retorna true |
| FH-APP-008 | Validar adaptação aprovação por mercado | Hook inicializado com config específica para UE | 1. Chamar RequiresApproval para "figma:export" com mercado "eu" | Retorna true para configuração específica |

#### 4.4 Validação de Uso

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| FH-VU-001 | Validar uso para figma:view | Hook inicializado, metadata válido | 1. Chamar ValidateTokenUse para "figma:view" com metadata válido | Validação bem-sucedida |
| FH-VU-002 | Validar uso para figma:edit | Hook inicializado, metadata válido | 1. Chamar ValidateTokenUse para "figma:edit" com metadata válido | Validação bem-sucedida |
| FH-VU-003 | Validar uso para figma:admin | Hook inicializado, metadata válido | 1. Chamar ValidateTokenUse para "figma:admin" com metadata válido | Validação bem-sucedida |
| FH-VU-004 | Validar uso para figma:comment | Hook inicializado, metadata válido | 1. Chamar ValidateTokenUse para "figma:comment" com metadata válido | Validação bem-sucedida |
| FH-VU-005 | Validar uso para figma:export | Hook inicializado, metadata válido | 1. Chamar ValidateTokenUse para "figma:export" com metadata válido | Validação bem-sucedida |
| FH-VU-006 | Validar uso para figma:library | Hook inicializado, metadata válido | 1. Chamar ValidateTokenUse para "figma:library" com metadata válido | Validação bem-sucedida |
| FH-VU-007 | Validar uso para figma:team | Hook inicializado, metadata válido | 1. Chamar ValidateTokenUse para "figma:team" com metadata válido | Validação bem-sucedida |
| FH-VU-008 | Rejeitar metadata inválido | Hook inicializado, metadata inválido | 1. Chamar ValidateTokenUse para "figma:edit" com metadata incompleto | Retorna erro de metadata inválido |
| FH-VU-009 | Rejeitar projeto protegido | Hook inicializado, metadata para projeto protegido | 1. Chamar ValidateTokenUse para "figma:edit" com metadata para projeto protegido | Retorna erro de operação proibida |
| FH-VU-010 | Validar adaptação por mercado | Hook inicializado, config específica para UE | 1. Chamar ValidateTokenUse para "figma:export" com metadata específico para UE | Validação adaptada para regras de mercado |

#### 4.5 Metadados de Auditoria

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| FH-AM-001 | Gerar metadados para figma:view | Hook inicializado, request válido | 1. Chamar GenerateAuditMetadata para "figma:view" com request válido | Metadados completos para auditoria |
| FH-AM-002 | Gerar metadados para figma:edit | Hook inicializado, request válido | 1. Chamar GenerateAuditMetadata para "figma:edit" com request válido | Metadados completos para auditoria |
| FH-AM-003 | Gerar metadados para figma:admin | Hook inicializado, request válido | 1. Chamar GenerateAuditMetadata para "figma:admin" com request válido | Metadados completos para auditoria |
| FH-AM-004 | Validar metadados específicos por mercado | Hook inicializado, config específica para UE | 1. Chamar GenerateAuditMetadata para "figma:export" com mercado "eu" | Metadados com campos específicos para UE |
| FH-AM-005 | Validar metadados para recursos sensíveis | Hook inicializado, metadata com recurso sensível | 1. Chamar GenerateAuditMetadata para "figma:library" com recurso sensível no metadata | Metadados incluindo detalhes específicos sobre o recurso sensível |

### 5. Hook Registry

#### 5.1 Registro de Hooks

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| HR-R-001 | Registrar Docker Hook | Registry inicializado | 1. Chamar RegisterHook com Docker Hook | Hook registrado com sucesso |
| HR-R-002 | Registrar GitHub Hook | Registry inicializado | 1. Chamar RegisterHook com GitHub Hook | Hook registrado com sucesso |
| HR-R-003 | Registrar Desktop Commander Hook | Registry inicializado | 1. Chamar RegisterHook com Desktop Commander Hook | Hook registrado com sucesso |
| HR-R-004 | Registrar Figma Hook | Registry inicializado | 1. Chamar RegisterHook com Figma Hook | Hook registrado com sucesso |
| HR-R-005 | Registrar hook duplicado | Registry inicializado, hook já registrado | 1. Chamar RegisterHook com hook já registrado | Retorna erro de hook duplicado |

#### 5.2 Recuperação de Hooks

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| HR-G-001 | Recuperar Docker Hook | Registry inicializado com hooks | 1. Chamar GetHook com tipo "docker" | Retorna hook Docker corretamente |
| HR-G-002 | Recuperar GitHub Hook | Registry inicializado com hooks | 1. Chamar GetHook com tipo "github" | Retorna hook GitHub corretamente |
| HR-G-003 | Recuperar Desktop Commander Hook | Registry inicializado com hooks | 1. Chamar GetHook com tipo "desktop-commander" | Retorna hook Desktop Commander corretamente |
| HR-G-004 | Recuperar Figma Hook | Registry inicializado com hooks | 1. Chamar GetHook com tipo "figma" | Retorna hook Figma corretamente |
| HR-G-005 | Recuperar hook não existente | Registry inicializado | 1. Chamar GetHook com tipo inexistente | Retorna erro de hook não encontrado |

#### 5.3 Validação de Escopos

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| HR-V-001 | Validar escopo Docker | Registry inicializado com hooks | 1. Chamar ValidateScope com escopo "docker:run" | Validação bem-sucedida, retorna hook Docker |
| HR-V-002 | Validar escopo GitHub | Registry inicializado com hooks | 1. Chamar ValidateScope com escopo "github:repo" | Validação bem-sucedida, retorna hook GitHub |
| HR-V-003 | Validar escopo Desktop Commander | Registry inicializado com hooks | 1. Chamar ValidateScope com escopo "desktop:file" | Validação bem-sucedida, retorna hook Desktop Commander |
| HR-V-004 | Validar escopo Figma | Registry inicializado com hooks | 1. Chamar ValidateScope com escopo "figma:edit" | Validação bem-sucedida, retorna hook Figma |
| HR-V-005 | Validar escopo inválido | Registry inicializado com hooks | 1. Chamar ValidateScope com escopo inválido | Retorna erro de escopo inválido |
| HR-V-006 | Validar escopo sem prefixo | Registry inicializado com hooks | 1. Chamar ValidateScope com escopo sem prefixo | Retorna erro de formato de escopo inválido |## Testes de Integração

### 1. Integração com Serviço de Elevação

#### 1.1 Solicitação de Elevação

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| IE-S-001 | Solicitar elevação para Docker | Serviço inicializado, Docker hook registrado | 1. Enviar solicitação para escopo docker:run<br>2. Processar solicitação no serviço | Solicitação processada, hook Docker validado corretamente |
| IE-S-002 | Solicitar elevação para GitHub | Serviço inicializado, GitHub hook registrado | 1. Enviar solicitação para escopo github:repo<br>2. Processar solicitação no serviço | Solicitação processada, hook GitHub validado corretamente |
| IE-S-003 | Solicitar elevação para Desktop Commander | Serviço inicializado, Desktop Commander hook registrado | 1. Enviar solicitação para escopo desktop:file<br>2. Processar solicitação no serviço | Solicitação processada, hook Desktop Commander validado corretamente |
| IE-S-004 | Solicitar elevação para Figma | Serviço inicializado, Figma hook registrado | 1. Enviar solicitação para escopo figma:edit<br>2. Processar solicitação no serviço | Solicitação processada, hook Figma validado corretamente |
| IE-S-005 | Solicitar elevação para escopo inválido | Serviço inicializado | 1. Enviar solicitação para escopo inválido<br>2. Processar solicitação no serviço | Erro apropriado retornado, indicando escopo inválido |
| IE-S-006 | Solicitar elevação com informações incompletas | Serviço inicializado | 1. Enviar solicitação com dados incompletos<br>2. Processar solicitação no serviço | Erro apropriado retornado, indicando dados incompletos |
| IE-S-007 | Solicitar elevação em modo de emergência | Serviço inicializado, hooks registrados | 1. Enviar solicitação com flag de emergência<br>2. Processar solicitação no serviço | Solicitação processada em modo de emergência, com validações específicas |

#### 1.2 Fluxo de Aprovação

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| IE-A-001 | Aprovar solicitação Docker | Serviço inicializado, solicitação Docker pendente | 1. Enviar aprovação para solicitação<br>2. Processar aprovação no serviço | Solicitação aprovada, token gerado |
| IE-A-002 | Aprovar solicitação GitHub | Serviço inicializado, solicitação GitHub pendente | 1. Enviar aprovação para solicitação<br>2. Processar aprovação no serviço | Solicitação aprovada, token gerado |
| IE-A-003 | Aprovar solicitação Desktop Commander | Serviço inicializado, solicitação Desktop Commander pendente | 1. Enviar aprovação para solicitação<br>2. Processar aprovação no serviço | Solicitação aprovada, token gerado |
| IE-A-004 | Aprovar solicitação Figma | Serviço inicializado, solicitação Figma pendente | 1. Enviar aprovação para solicitação<br>2. Processar aprovação no serviço | Solicitação aprovada, token gerado |
| IE-A-005 | Rejeitar solicitação | Serviço inicializado, solicitação pendente | 1. Enviar rejeição para solicitação<br>2. Processar rejeição no serviço | Solicitação rejeitada, nenhum token gerado |
| IE-A-006 | Aprovar solicitação com aprovador inválido | Serviço inicializado, solicitação pendente | 1. Enviar aprovação com usuário não autorizado<br>2. Processar aprovação no serviço | Erro apropriado retornado, solicitação permanece pendente |
| IE-A-007 | Aprovar solicitação expirada | Serviço inicializado, solicitação expirada | 1. Enviar aprovação para solicitação expirada<br>2. Processar aprovação no serviço | Erro apropriado retornado, indicando expiração da solicitação |
| IE-A-008 | Fluxo de aprovação multi-nível | Serviço inicializado, solicitação que requer múltiplas aprovações | 1. Enviar primeira aprovação<br>2. Verificar estado da solicitação<br>3. Enviar segunda aprovação<br>4. Processar aprovação final no serviço | Solicitação marcada como parcialmente aprovada após primeira aprovação, aprovada após segunda aprovação |

#### 1.3 Validação de Uso de Token

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| IE-T-001 | Validar uso de token Docker | Serviço inicializado, token Docker válido | 1. Enviar validação para token com metadados Docker<br>2. Processar validação no serviço | Uso validado com sucesso |
| IE-T-002 | Validar uso de token GitHub | Serviço inicializado, token GitHub válido | 1. Enviar validação para token com metadados GitHub<br>2. Processar validação no serviço | Uso validado com sucesso |
| IE-T-003 | Validar uso de token Desktop Commander | Serviço inicializado, token Desktop Commander válido | 1. Enviar validação para token com metadados Desktop Commander<br>2. Processar validação no serviço | Uso validado com sucesso |
| IE-T-004 | Validar uso de token Figma | Serviço inicializado, token Figma válido | 1. Enviar validação para token com metadados Figma<br>2. Processar validação no serviço | Uso validado com sucesso |
| IE-T-005 | Validar uso de token inválido | Serviço inicializado | 1. Enviar validação para token inexistente<br>2. Processar validação no serviço | Erro apropriado retornado, uso rejeitado |
| IE-T-006 | Validar uso de token expirado | Serviço inicializado, token expirado | 1. Enviar validação para token expirado<br>2. Processar validação no serviço | Erro apropriado retornado, uso rejeitado |
| IE-T-007 | Validar uso de token com escopo incorreto | Serviço inicializado, token válido | 1. Enviar validação para token com escopo diferente<br>2. Processar validação no serviço | Erro apropriado retornado, uso rejeitado |
| IE-T-008 | Validar uso de token com metadados inválidos | Serviço inicializado, token válido | 1. Enviar validação para token com metadados inválidos<br>2. Processar validação no serviço | Erro apropriado retornado, uso rejeitado |
| IE-T-009 | Validar uso de token por mercado | Serviço inicializado, token válido para mercado específico | 1. Enviar validação para token com mercado correto<br>2. Processar validação no serviço | Uso validado com sucesso, regras específicas de mercado aplicadas |

### 2. Integração com API Gateway

#### 2.1 Roteamento de Requisições

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| IG-R-001 | Rotear solicitação de elevação | API Gateway configurado, serviço inicializado | 1. Enviar requisição para endpoint de elevação no gateway<br>2. Verificar roteamento para serviço correto | Requisição roteada corretamente, resposta recebida |
| IG-R-002 | Rotear aprovação de elevação | API Gateway configurado, serviço inicializado | 1. Enviar requisição de aprovação para gateway<br>2. Verificar roteamento para serviço correto | Requisição roteada corretamente, resposta recebida |
| IG-R-003 | Rotear validação de token | API Gateway configurado, serviço inicializado | 1. Enviar requisição de validação para gateway<br>2. Verificar roteamento para serviço correto | Requisição roteada corretamente, resposta recebida |
| IG-R-004 | Rotear consulta de solicitações | API Gateway configurado, serviço inicializado | 1. Enviar requisição de consulta para gateway<br>2. Verificar roteamento para serviço correto | Requisição roteada corretamente, resposta recebida |

#### 2.2 Autenticação e Autorização

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| IG-A-001 | Autenticar requisição com JWT válido | API Gateway configurado, serviço inicializado | 1. Enviar requisição com token JWT válido<br>2. Verificar autenticação | Autenticação bem-sucedida, requisição processada |
| IG-A-002 | Rejeitar requisição com JWT inválido | API Gateway configurado, serviço inicializado | 1. Enviar requisição com token JWT inválido<br>2. Verificar autenticação | Autenticação falha, erro 401 retornado |
| IG-A-003 | Rejeitar requisição sem JWT | API Gateway configurado, serviço inicializado | 1. Enviar requisição sem token JWT<br>2. Verificar autenticação | Autenticação falha, erro 401 retornado |
| IG-A-004 | Verificar autorização para solicitação | API Gateway configurado, serviço inicializado | 1. Enviar requisição com usuário sem permissão<br>2. Verificar autorização | Autorização falha, erro 403 retornado |

#### 2.3 Rate Limiting e Controle de Tráfego

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| IG-RL-001 | Aplicar rate limit por usuário | API Gateway configurado com rate limiting | 1. Enviar múltiplas requisições em sequência rápida<br>2. Verificar aplicação de limites | Após limite excedido, erro 429 retornado |
| IG-RL-002 | Aplicar rate limit por IP | API Gateway configurado com rate limiting | 1. Enviar múltiplas requisições do mesmo IP<br>2. Verificar aplicação de limites | Após limite excedido, erro 429 retornado |
| IG-RL-003 | Aplicar rate limit por tenant | API Gateway configurado com rate limiting | 1. Enviar múltiplas requisições para o mesmo tenant<br>2. Verificar aplicação de limites | Após limite excedido, erro 429 retornado |
| IG-RL-004 | Verificar comportamento com tráfego distribuído | API Gateway configurado com rate limiting | 1. Enviar requisições distribuídas entre diferentes usuários/IPs<br>2. Verificar aplicação de limites | Rate limits aplicados individualmente por usuário/IP |

### 3. Integração com Sistema de Identidade

#### 3.1 Autenticação de Usuários

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| II-A-001 | Validar identidade do solicitante | Serviço de identidade disponível | 1. Processar solicitação com token de identidade<br>2. Verificar validação da identidade | Identidade validada corretamente |
| II-A-002 | Rejeitar solicitação com identidade inválida | Serviço de identidade disponível | 1. Processar solicitação com token de identidade inválido<br>2. Verificar validação da identidade | Erro apropriado retornado, solicitação rejeitada |
| II-A-003 | Validar identidade do aprovador | Serviço de identidade disponível | 1. Processar aprovação com token de identidade<br>2. Verificar validação da identidade | Identidade validada corretamente |
| II-A-004 | Verificar atributos e papéis | Serviço de identidade disponível | 1. Processar solicitação com usuário que tem papéis específicos<br>2. Verificar recuperação de atributos e papéis | Atributos e papéis recuperados corretamente |

#### 3.2 Verificação de Autorizações

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| II-AZ-001 | Verificar autorização para solicitar | Serviço de identidade disponível | 1. Processar solicitação com usuário autorizado<br>2. Verificar autorização | Autorização validada corretamente |
| II-AZ-002 | Rejeitar solicitação não autorizada | Serviço de identidade disponível | 1. Processar solicitação com usuário não autorizado<br>2. Verificar autorização | Erro apropriado retornado, solicitação rejeitada |
| II-AZ-003 | Verificar autorização para aprovar | Serviço de identidade disponível | 1. Processar aprovação com aprovador autorizado<br>2. Verificar autorização | Autorização validada corretamente |
| II-AZ-004 | Rejeitar aprovação não autorizada | Serviço de identidade disponível | 1. Processar aprovação com usuário não autorizado<br>2. Verificar autorização | Erro apropriado retornado, aprovação rejeitada |

### 4. Integração com Sistema MFA

#### 4.1 Validação de MFA

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| IM-V-001 | Validar MFA para Docker | Serviço MFA disponível | 1. Processar solicitação Docker que requer MFA<br>2. Verificar exigência e validação de MFA | MFA exigido e validado corretamente |
| IM-V-002 | Validar MFA para GitHub | Serviço MFA disponível | 1. Processar solicitação GitHub que requer MFA<br>2. Verificar exigência e validação de MFA | MFA exigido e validado corretamente |
| IM-V-003 | Validar MFA para Desktop Commander | Serviço MFA disponível | 1. Processar solicitação Desktop Commander que requer MFA<br>2. Verificar exigência e validação de MFA | MFA exigido e validado corretamente |
| IM-V-004 | Validar MFA para Figma | Serviço MFA disponível | 1. Processar solicitação Figma que requer MFA<br>2. Verificar exigência e validação de MFA | MFA exigido e validado corretamente |
| IM-V-005 | Processar solicitação sem MFA quando não requerido | Serviço MFA disponível | 1. Processar solicitação que não requer MFA<br>2. Verificar não exigência de MFA | MFA não exigido, solicitação processada |

#### 4.2 Níveis de MFA

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| IM-N-001 | Validar nível básico de MFA | Serviço MFA disponível | 1. Processar solicitação que requer MFA básico<br>2. Verificar validação com fator único | MFA básico validado corretamente |
| IM-N-002 | Validar nível forte de MFA | Serviço MFA disponível | 1. Processar solicitação que requer MFA forte<br>2. Verificar validação com múltiplos fatores | MFA forte validado corretamente |
| IM-N-003 | Validar step-up MFA | Serviço MFA disponível | 1. Processar solicitação que requer elevação de nível MFA<br>2. Verificar exigência de fator adicional | Step-up MFA exigido e validado corretamente |
| IM-N-004 | Adaptar níveis de MFA por mercado | Serviço MFA disponível, configuração específica de mercado | 1. Processar solicitação para mercado específico<br>2. Verificar aplicação de regras MFA específicas | Regras MFA específicas do mercado aplicadas corretamente |

### 5. Integração com Sistema de Auditoria

#### 5.1 Geração de Eventos de Auditoria

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| IA-G-001 | Gerar evento de solicitação | Sistema de auditoria disponível | 1. Processar solicitação de elevação<br>2. Verificar geração de evento de auditoria | Evento de auditoria gerado com todos os campos requeridos |
| IA-G-002 | Gerar evento de aprovação | Sistema de auditoria disponível | 1. Processar aprovação de solicitação<br>2. Verificar geração de evento de auditoria | Evento de auditoria gerado com todos os campos requeridos |
| IA-G-003 | Gerar evento de rejeição | Sistema de auditoria disponível | 1. Processar rejeição de solicitação<br>2. Verificar geração de evento de auditoria | Evento de auditoria gerado com todos os campos requeridos |
| IA-G-004 | Gerar evento de uso de token | Sistema de auditoria disponível | 1. Processar validação de uso de token<br>2. Verificar geração de evento de auditoria | Evento de auditoria gerado com todos os campos requeridos |
| IA-G-005 | Gerar evento de expiração de token | Sistema de auditoria disponível | 1. Expirar token<br>2. Verificar geração de evento de auditoria | Evento de auditoria gerado com todos os campos requeridos |

#### 5.2 Metadados de Auditoria Específicos

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| IA-M-001 | Verificar metadados Docker | Sistema de auditoria disponível | 1. Processar operação Docker<br>2. Verificar metadados específicos no evento de auditoria | Metadados específicos do Docker presentes no evento |
| IA-M-002 | Verificar metadados GitHub | Sistema de auditoria disponível | 1. Processar operação GitHub<br>2. Verificar metadados específicos no evento de auditoria | Metadados específicos do GitHub presentes no evento |
| IA-M-003 | Verificar metadados Desktop Commander | Sistema de auditoria disponível | 1. Processar operação Desktop Commander<br>2. Verificar metadados específicos no evento de auditoria | Metadados específicos do Desktop Commander presentes no evento |
| IA-M-004 | Verificar metadados Figma | Sistema de auditoria disponível | 1. Processar operação Figma<br>2. Verificar metadados específicos no evento de auditoria | Metadados específicos do Figma presentes no evento |
| IA-M-005 | Verificar metadados específicos por mercado | Sistema de auditoria disponível | 1. Processar operação para mercado específico<br>2. Verificar metadados regulatórios específicos | Metadados regulatórios específicos presentes no evento |## Testes de Conformidade Regulatória

### 1. Angola e Moçambique (SADC/PALOP)

#### 1.1 Conformidade com Banco Nacional de Angola (BNA)

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| CR-ANG-001 | Verificar aprovação dupla para operações críticas | Sistema configurado para mercado Angola | 1. Processar solicitação de operação crítica para mercado Angola<br>2. Verificar exigência de aprovação dupla | Aprovação dupla exigida conforme regulamentação BNA |
| CR-ANG-002 | Verificar retenção de logs para auditoria BNA | Sistema configurado para mercado Angola | 1. Gerar eventos de auditoria para operações financeiras<br>2. Verificar configuração de retenção | Período de retenção configurado conforme exigido pelo BNA (mínimo 7 anos) |
| CR-ANG-003 | Validar campos específicos para operações cambiais | Sistema configurado para mercado Angola | 1. Processar operação que envolve câmbio<br>2. Verificar metadados específicos | Metadados específicos para operações cambiais conforme exigido pelo BNA |
| CR-ANG-004 | Verificar limites específicos para transações | Sistema configurado para mercado Angola | 1. Processar operação acima do limite definido pelo BNA<br>2. Verificar validações adicionais | Validações adicionais aplicadas conforme limites BNA |
| CR-ANG-005 | Validar exigências de MFA para operações financeiras | Sistema configurado para mercado Angola | 1. Processar operação financeira sensível<br>2. Verificar nível de MFA exigido | MFA forte exigido conforme diretrizes BNA |

#### 1.2 Conformidade com Banco de Moçambique

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| CR-MOZ-001 | Verificar aprovações para transações de alto valor | Sistema configurado para mercado Moçambique | 1. Processar solicitação de transação de alto valor<br>2. Verificar fluxo de aprovação | Fluxo de aprovação específico conforme regulamentação do Banco de Moçambique |
| CR-MOZ-002 | Validar campos de auditoria para mobile money | Sistema configurado para mercado Moçambique | 1. Processar operação de mobile money<br>2. Verificar campos de auditoria específicos | Campos específicos para mobile money conforme regulamentações locais |
| CR-MOZ-003 | Verificar limites de transação por agente | Sistema configurado para mercado Moçambique | 1. Processar operação de agente acima do limite<br>2. Verificar validações adicionais | Validações adicionais aplicadas conforme limites do Banco de Moçambique |
| CR-MOZ-004 | Validar exigências para operações transfronteiriças | Sistema configurado para mercado Moçambique | 1. Processar operação transfronteiriça<br>2. Verificar validações e campos específicos | Validações e campos específicos para operações transfronteiriças |
| CR-MOZ-005 | Verificar validações específicas SADC | Sistema configurado para mercado Moçambique | 1. Processar operação com impacto SADC<br>2. Verificar conformidade regional | Requisitos de conformidade SADC aplicados corretamente |

### 2. Brasil

#### 2.1 Conformidade com LGPD

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| CR-BR-001 | Verificar justificativas para acesso a dados pessoais | Sistema configurado para mercado Brasil | 1. Processar solicitação que acessa dados pessoais<br>2. Verificar registro de justificativa | Justificativa registrada conforme exigido pela LGPD |
| CR-BR-002 | Validar aprovação específica para dados sensíveis | Sistema configurado para mercado Brasil | 1. Processar solicitação que acessa dados sensíveis<br>2. Verificar aprovação DPO | Aprovação do DPO registrada conforme LGPD |
| CR-BR-003 | Verificar limitação de escopo para operações | Sistema configurado para mercado Brasil | 1. Processar solicitação com múltiplos escopos<br>2. Verificar aplicação do princípio da minimização | Acesso limitado ao mínimo necessário conforme LGPD |
| CR-BR-004 | Validar retenção de logs para conformidade LGPD | Sistema configurado para mercado Brasil | 1. Gerar eventos de auditoria para acesso a dados pessoais<br>2. Verificar política de retenção | Política de retenção configurada conforme LGPD |
| CR-BR-005 | Verificar metadados específicos para auditoria LGPD | Sistema configurado para mercado Brasil | 1. Processar operação que acessa dados pessoais<br>2. Verificar metadados de auditoria | Metadados específicos LGPD presentes nos logs |

#### 2.2 Conformidade com Banco Central do Brasil (BACEN)

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| CR-BR-006 | Validar aprovações para operações financeiras | Sistema configurado para mercado Brasil | 1. Processar operação financeira crítica<br>2. Verificar fluxo de aprovação | Aprovações conforme requisitos BACEN |
| CR-BR-007 | Verificar logs específicos para transações PIX | Sistema configurado para mercado Brasil | 1. Processar operação relacionada ao PIX<br>2. Verificar campos de auditoria específicos | Campos específicos para PIX conforme regulamentações BACEN |
| CR-BR-008 | Validar controles para arranjos de pagamento | Sistema configurado para mercado Brasil | 1. Processar operação em arranjo de pagamento<br>2. Verificar controles específicos | Controles específicos aplicados conforme regulamentações BACEN |
| CR-BR-009 | Verificar integração com listas de PEP locais | Sistema configurado para mercado Brasil | 1. Processar operação com usuário em lista PEP<br>2. Verificar validações adicionais | Validações adicionais aplicadas para Pessoas Politicamente Expostas |
| CR-BR-010 | Validar requisitos de segregação de funções | Sistema configurado para mercado Brasil | 1. Processar operação que requer segregação<br>2. Verificar validações de aprovação | Segregação de funções aplicada conforme requisitos BACEN |

### 3. Europa (UE)

#### 3.1 Conformidade com GDPR

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| CR-EU-001 | Verificar justificativas para processamento de dados | Sistema configurado para mercado UE | 1. Processar solicitação que acessa dados pessoais<br>2. Verificar registro de base legal | Base legal registrada conforme exigido pelo GDPR |
| CR-EU-002 | Validar minimização de dados em logs | Sistema configurado para mercado UE | 1. Gerar eventos de auditoria para operação com dados pessoais<br>2. Verificar minimização nos logs | Dados pessoais minimizados nos logs conforme GDPR |
| CR-EU-003 | Verificar capacidade de atender direito de acesso | Sistema configurado para mercado UE | 1. Simular solicitação de acesso a dados<br>2. Verificar capacidade de extração | Sistema capaz de extrair todos os dados relacionados ao usuário |
| CR-EU-004 | Validar controles de retenção de dados | Sistema configurado para mercado UE | 1. Verificar políticas de retenção<br>2. Validar mecanismos de exclusão automática | Mecanismos de retenção e exclusão conformes com GDPR |
| CR-EU-005 | Verificar metadados específicos para auditoria GDPR | Sistema configurado para mercado UE | 1. Processar operação que acessa dados pessoais<br>2. Verificar metadados de auditoria | Metadados específicos GDPR presentes nos logs |

#### 3.2 Conformidade com PSD2

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| CR-EU-006 | Validar autenticação forte (SCA) | Sistema configurado para mercado UE | 1. Processar operação de pagamento<br>2. Verificar requisito de SCA | Autenticação forte exigida conforme PSD2 |
| CR-EU-007 | Verificar controles para operações de pagamento | Sistema configurado para mercado UE | 1. Processar operação de pagamento<br>2. Verificar controles específicos | Controles específicos aplicados conforme PSD2 |
| CR-EU-008 | Validar isenções de SCA quando aplicáveis | Sistema configurado para mercado UE | 1. Processar operação com critério de isenção<br>2. Verificar aplicação de isenção | Isenções de SCA aplicadas corretamente quando cabíveis |
| CR-EU-009 | Verificar metadados específicos para auditoria PSD2 | Sistema configurado para mercado UE | 1. Processar operação de pagamento<br>2. Verificar metadados de auditoria | Metadados específicos PSD2 presentes nos logs |
| CR-EU-010 | Validar requisitos de monitoramento de fraude | Sistema configurado para mercado UE | 1. Processar operação de pagamento<br>2. Verificar integração com monitoramento de fraude | Integração com monitoramento de fraude conforme PSD2 |

### 4. EUA e Global

#### 4.1 Conformidade com SOX

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| CR-US-001 | Verificar segregação de funções | Sistema configurado para mercado EUA | 1. Processar operação financeira crítica<br>2. Verificar aplicação de segregação de funções | Segregação de funções aplicada conforme SOX |
| CR-US-002 | Validar trilhas de auditoria completas | Sistema configurado para mercado EUA | 1. Realizar operação financeira<br>2. Verificar completude da trilha de auditoria | Trilha de auditoria completa conforme requisitos SOX |
| CR-US-003 | Verificar aprovações para modificações de sistema | Sistema configurado para mercado EUA | 1. Processar operação de modificação de sistema<br>2. Verificar fluxo de aprovação | Aprovações para mudanças conforme SOX |
| CR-US-004 | Validar controles de acesso conforme SOX | Sistema configurado para mercado EUA | 1. Tentar acesso com diferentes níveis de permissão<br>2. Verificar aplicação de controles | Controles de acesso aplicados conforme SOX |
| CR-US-005 | Verificar evidências de revisão de acessos | Sistema configurado para mercado EUA | 1. Gerar relatórios de revisão de acessos<br>2. Verificar conformidade com SOX | Evidências de revisão conforme requisitos SOX |

#### 4.2 Conformidade com AML/CFT Global

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| CR-GL-001 | Verificar validações contra listas de sanções | Sistema configurado para mercado global | 1. Processar operação com parte em lista de sanções<br>2. Verificar validações | Validações contra listas de sanções aplicadas |
| CR-GL-002 | Validar controles para transações de alto risco | Sistema configurado para mercado global | 1. Processar transação de alto risco<br>2. Verificar controles adicionais | Controles adicionais aplicados conforme requisitos AML |
| CR-GL-003 | Verificar rastreabilidade de transações | Sistema configurado para mercado global | 1. Realizar transação complexa<br>2. Verificar rastreabilidade completa | Rastreabilidade completa conforme requisitos CFT |
| CR-GL-004 | Validar geração de alertas para operações suspeitas | Sistema configurado para mercado global | 1. Simular operação suspeita<br>2. Verificar geração de alerta | Alertas gerados conforme requisitos AML/CFT |
| CR-GL-005 | Verificar níveis de aprovação por valor de transação | Sistema configurado para mercado global | 1. Processar transações de diferentes valores<br>2. Verificar níveis de aprovação | Níveis de aprovação escalonados conforme valor |

### 5. China e BRICS

#### 5.1 Conformidade com Regulamentos de Cibersegurança da China

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| CR-CN-001 | Verificar localização de dados | Sistema configurado para mercado China | 1. Processar operação com dados chineses<br>2. Verificar armazenamento local | Dados armazenados localmente conforme regulamentação chinesa |
| CR-CN-002 | Validar aprovações específicas para transferência de dados | Sistema configurado para mercado China | 1. Processar transferência de dados<br>2. Verificar aprovações específicas | Aprovações conforme regulamentação chinesa |
| CR-CN-003 | Verificar campos de auditoria específicos | Sistema configurado para mercado China | 1. Gerar eventos de auditoria<br>2. Verificar campos específicos | Campos específicos conforme regulamentação chinesa |
| CR-CN-004 | Validar controles para operações com criptomoedas | Sistema configurado para mercado China | 1. Processar operação relacionada a criptomoedas<br>2. Verificar validações específicas | Validações específicas conforme regulamentação chinesa |
| CR-CN-005 | Verificar controles para pagamentos transfronteiriços | Sistema configurado para mercado China | 1. Processar pagamento transfronteiriço<br>2. Verificar validações específicas | Validações específicas conforme regulamentação chinesa |

## Testes de Desempenho

### 1. Testes de Carga

#### 1.1 Processamento de Solicitações

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| PF-CG-001 | Validar throughput de solicitações | Sistema em ambiente de teste | 1. Enviar 100 solicitações por segundo<br>2. Medir tempo de resposta e taxa de sucesso | Tempo de resposta < 200ms, taxa de sucesso > 99% |
| PF-CG-002 | Testar limite de solicitações simultâneas | Sistema em ambiente de teste | 1. Aumentar gradualmente número de solicitações simultâneas<br>2. Identificar ponto de degradação | Sistema suporta pelo menos 500 solicitações simultâneas com degradação controlada |
| PF-CG-003 | Validar desempenho sob carga por mercado | Sistema em ambiente de teste | 1. Enviar solicitações para diferentes mercados simultaneamente<br>2. Medir tempo de resposta por mercado | Tempo de resposta consistente entre mercados |
| PF-CG-004 | Testar recuperação após sobrecarga | Sistema em ambiente de teste | 1. Sobrecarregar sistema<br>2. Reduzir carga<br>3. Medir tempo de recuperação | Sistema se recupera em < 30 segundos após redução de carga |
| PF-CG-005 | Validar throughput com diferentes hooks | Sistema em ambiente de teste | 1. Enviar solicitações para diferentes hooks simultaneamente<br>2. Medir desempenho por tipo de hook | Desempenho consistente entre diferentes tipos de hook |#### 1.2 Aprovação e Validação

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| PF-CG-006 | Validar throughput de aprovações | Sistema em ambiente de teste | 1. Enviar 50 aprovações por segundo<br>2. Medir tempo de resposta e taxa de sucesso | Tempo de resposta < 300ms, taxa de sucesso > 99% |
| PF-CG-007 | Validar throughput de validações de token | Sistema em ambiente de teste | 1. Enviar 200 validações por segundo<br>2. Medir tempo de resposta e taxa de sucesso | Tempo de resposta < 100ms, taxa de sucesso > 99.5% |
| PF-CG-008 | Testar desempenho com múltiplos níveis de aprovação | Sistema em ambiente de teste | 1. Processar solicitações com diferentes níveis de aprovação<br>2. Medir impacto no desempenho | Impacto no desempenho < 20% com níveis adicionais de aprovação |
| PF-CG-009 | Validar desempenho de aprovações em massa | Sistema em ambiente de teste | 1. Enviar aprovação para múltiplas solicitações<br>2. Medir tempo de processamento | Processamento de 100 aprovações em lote em < 5 segundos |
| PF-CG-010 | Testar validação de tokens em modo degradado | Sistema em ambiente de teste, simulação de falha parcial | 1. Processar validações de token durante falha parcial<br>2. Medir tempo de resposta e taxa de sucesso | Tempo de resposta < 200ms, taxa de sucesso > 95% |

### 2. Testes de Latência

#### 2.1 Tempos de Resposta

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| PF-LT-001 | Medir latência de solicitação | Sistema em ambiente de teste | 1. Enviar solicitações individuais<br>2. Medir tempo de resposta | Tempo de resposta P95 < 200ms |
| PF-LT-002 | Medir latência de aprovação | Sistema em ambiente de teste | 1. Enviar aprovações individuais<br>2. Medir tempo de resposta | Tempo de resposta P95 < 250ms |
| PF-LT-003 | Medir latência de validação de token | Sistema em ambiente de teste | 1. Enviar validações individuais<br>2. Medir tempo de resposta | Tempo de resposta P95 < 100ms |
| PF-LT-004 | Verificar impacto de validações complexas na latência | Sistema em ambiente de teste | 1. Enviar validações com metadados complexos<br>2. Medir impacto na latência | Aumento de latência < 50ms para validações complexas |
| PF-LT-005 | Medir latência em ambiente distribuído | Sistema em ambiente distribuído | 1. Enviar solicitações entre diferentes regiões<br>2. Medir latência entre componentes | Tempo de resposta total P95 < 500ms em ambiente distribuído |

#### 2.2 Análise de Componentes

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| PF-AC-001 | Analisar contribuição do API Gateway na latência | Sistema instrumentado | 1. Enviar solicitações através do gateway<br>2. Medir latência em cada componente | API Gateway contribui com < 50ms na latência total |
| PF-AC-002 | Analisar contribuição do serviço de elevação na latência | Sistema instrumentado | 1. Processar solicitações de elevação<br>2. Medir latência no serviço | Serviço de elevação processa em < 100ms |
| PF-AC-003 | Analisar contribuição dos hooks na latência | Sistema instrumentado | 1. Processar validações para diferentes hooks<br>2. Medir latência por hook | Hooks processam em < 50ms cada |
| PF-AC-004 | Analisar contribuição do banco de dados na latência | Sistema instrumentado | 1. Realizar operações com banco de dados<br>2. Medir latência nas operações de BD | Operações de BD consomem < 50ms |
| PF-AC-005 | Analisar contribuição da rede na latência | Sistema instrumentado | 1. Processar operações entre componentes<br>2. Medir latência de rede | Latência de rede contribui com < 30ms entre componentes |

### 3. Testes de Escalabilidade

#### 3.1 Escalabilidade Horizontal

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| PF-ES-001 | Validar escalabilidade horizontal do serviço de elevação | Ambiente Kubernetes | 1. Aumentar carga gradualmente<br>2. Observar escalonamento automático<br>3. Medir desempenho | Serviço escala horizontalmente mantendo tempo de resposta estável |
| PF-ES-002 | Validar escalabilidade horizontal do API Gateway | Ambiente Kubernetes | 1. Aumentar carga no gateway<br>2. Observar escalonamento automático<br>3. Medir desempenho | Gateway escala horizontalmente mantendo tempo de resposta estável |
| PF-ES-003 | Testar limites de escalabilidade | Ambiente Kubernetes | 1. Aumentar carga até limite do sistema<br>2. Identificar gargalos<br>3. Validar comportamento | Sistema degrada graciosamente próximo ao limite, sem falhas catastróficas |
| PF-ES-004 | Verificar distribuição de carga entre instâncias | Ambiente com múltiplas instâncias | 1. Gerar carga distribuída<br>2. Medir utilização por instância | Carga distribuída uniformemente entre instâncias (±10%) |
| PF-ES-005 | Testar recuperação após falha de nó | Ambiente Kubernetes com múltiplos nós | 1. Gerar carga constante<br>2. Simular falha em um nó<br>3. Medir tempo de recuperação | Serviço se recupera em < 30 segundos após falha de nó |

#### 3.2 Escalabilidade de Dados

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| PF-ED-001 | Validar desempenho com grande volume de tokens | Banco de dados com muitos tokens | 1. Executar operações com BD com 1 milhão de tokens<br>2. Medir tempo de resposta | Tempo de resposta mantém-se estável com alto volume de dados |
| PF-ED-002 | Testar desempenho com grande volume de solicitações | Banco de dados com muitas solicitações | 1. Executar operações com BD com 500 mil solicitações<br>2. Medir tempo de resposta | Tempo de resposta mantém-se estável com alto volume de dados |
| PF-ED-003 | Verificar impacto de índices no desempenho | Banco de dados com índices otimizados | 1. Executar consultas complexas<br>2. Comparar desempenho com/sem índices | Consultas otimizadas executam em < 50ms mesmo com alto volume |
| PF-ED-004 | Testar particionamento de dados por tenant | Banco de dados particionado | 1. Executar consultas em múltiplos tenants<br>2. Medir isolamento e desempenho | Desempenho consistente independente do número de tenants |
| PF-ED-005 | Validar particionamento de dados por mercado | Banco de dados particionado | 1. Executar consultas em múltiplos mercados<br>2. Medir isolamento e desempenho | Desempenho consistente independente do número de mercados |

## Testes de Segurança

### 1. Análise de Vulnerabilidades

#### 1.1 Análise de Código

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| SG-AC-001 | Análise estática de código (SAST) | Código-fonte disponível | 1. Executar ferramentas SAST (Gosec, SonarQube)<br>2. Analisar resultados | Zero vulnerabilidades críticas ou altas |
| SG-AC-002 | Verificar secrets hardcoded | Código-fonte disponível | 1. Executar ferramentas de detecção de secrets<br>2. Analisar resultados | Zero secrets hardcoded no código |
| SG-AC-003 | Verificar vulnerabilidades em dependências | Código-fonte e dependências disponíveis | 1. Executar ferramenta de análise de dependências<br>2. Analisar resultados | Zero dependências com vulnerabilidades críticas |
| SG-AC-004 | Validar conformidade com práticas seguras de codificação | Código-fonte disponível | 1. Executar verificações de padrões de código<br>2. Analisar resultados | Código em conformidade com práticas seguras |
| SG-AC-005 | Verificar tratamento de erros e exceções | Código-fonte disponível | 1. Analisar padrões de tratamento de erro<br>2. Verificar logging de exceções | Tratamento adequado de erros sem vazamento de informações sensíveis |

#### 1.2 Teste de Penetração

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| SG-PT-001 | Testar autenticação e autorização | Sistema em ambiente de teste | 1. Tentar acessos não autorizados<br>2. Verificar resposta do sistema | Todos os acessos não autorizados são bloqueados |
| SG-PT-002 | Verificar proteção contra injection | Sistema em ambiente de teste | 1. Tentar ataques de injection (SQL, NoSQL, OS)<br>2. Verificar resposta do sistema | Sistema resiste a tentativas de injection |
| SG-PT-003 | Testar proteção de dados em trânsito | Sistema em ambiente de teste | 1. Analisar tráfego de rede<br>2. Verificar criptografia | Dados em trânsito adequadamente protegidos (TLS 1.2+) |
| SG-PT-004 | Verificar proteção contra ataques de força bruta | Sistema em ambiente de teste | 1. Tentar ataques de força bruta<br>2. Verificar mecanismos de proteção | Sistema implementa rate limiting e proteções contra força bruta |
| SG-PT-005 | Testar proteção contra ataques OWASP Top 10 | Sistema em ambiente de teste | 1. Executar testes para cada categoria OWASP Top 10<br>2. Verificar proteções | Sistema protegido contra vulnerabilidades OWASP Top 10 |

### 2. Testes de Segurança de Dados

#### 2.1 Proteção de Dados

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| SG-PD-001 | Verificar criptografia de dados sensíveis em repouso | Sistema com dados | 1. Analisar armazenamento de dados sensíveis<br>2. Verificar criptografia | Dados sensíveis criptografados em repouso |
| SG-PD-002 | Validar segregação de dados por tenant | Sistema multi-tenant | 1. Acessar dados como diferentes tenants<br>2. Verificar isolamento | Completo isolamento de dados entre tenants |
| SG-PD-003 | Verificar mascaramento de dados sensíveis em logs | Sistema gerando logs | 1. Gerar logs com dados sensíveis<br>2. Verificar mascaramento | Dados sensíveis adequadamente mascarados nos logs |
| SG-PD-004 | Testar gerenciamento de chaves criptográficas | Sistema usando criptografia | 1. Analisar ciclo de vida das chaves<br>2. Verificar rotação e proteção | Chaves adequadamente protegidas e com rotação periódica |
| SG-PD-005 | Validar proteção de backups | Sistema com backups | 1. Analisar segurança dos backups<br>2. Verificar criptografia e acesso | Backups criptografados e com acesso controlado |

#### 2.2 Gestão de Tokens

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| SG-GT-001 | Verificar geração segura de tokens | Sistema gerando tokens | 1. Analisar processo de geração de tokens<br>2. Verificar aleatoriedade e segurança | Tokens gerados com entropia suficiente (≥256 bits) |
| SG-GT-002 | Validar armazenamento seguro de tokens | Sistema armazenando tokens | 1. Analisar armazenamento de tokens<br>2. Verificar proteções | Tokens armazenados com hash ou criptografia |
| SG-GT-003 | Verificar expiração de tokens | Sistema com tokens ativos | 1. Verificar mecanismo de expiração<br>2. Testar uso após expiração | Tokens expirados rejeitados pelo sistema |
| SG-GT-004 | Testar revogação de tokens | Sistema com tokens ativos | 1. Revogar token ativo<br>2. Tentar usar token revogado | Tokens revogados rejeitados pelo sistema |
| SG-GT-005 | Validar limitação de escopo de tokens | Sistema com tokens para diferentes escopos | 1. Tentar usar token para escopo não autorizado<br>2. Verificar validação | Tokens limitados estritamente ao escopo autorizado |

### 3. Testes de Segurança Operacional

#### 3.1 Controle de Acesso

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| SG-CA-001 | Verificar RBAC para API Gateway | Sistema com RBAC configurado | 1. Tentar acessos com diferentes papéis<br>2. Verificar autorização | Acesso controlado estritamente por papel |
| SG-CA-002 | Validar controles de acesso por mercado | Sistema com controles por mercado | 1. Tentar acessos cross-mercado<br>2. Verificar isolamento | Acesso restrito ao mercado autorizado |
| SG-CA-003 | Verificar controles de acesso por tenant | Sistema multi-tenant | 1. Tentar acessos cross-tenant<br>2. Verificar isolamento | Acesso restrito ao tenant autorizado |
| SG-CA-004 | Testar segregação de funções | Sistema com segregação configurada | 1. Tentar operações com conflito de interesse<br>2. Verificar bloqueio | Operações com conflito de interesse bloqueadas |
| SG-CA-005 | Validar auditoria de acessos privilegiados | Sistema com auditoria | 1. Realizar acessos privilegiados<br>2. Verificar registros de auditoria | Todos os acessos privilegiados auditados detalhadamente |## Testes de Observabilidade

### 1. Rastreamento Distribuído

#### 1.1 Rastreamento com OpenTelemetry

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| OB-RT-001 | Verificar propagação de contexto | Sistema instrumentado com OpenTelemetry | 1. Iniciar trace em solicitação<br>2. Verificar propagação através de componentes | Contexto propagado corretamente por toda a cadeia |
| OB-RT-002 | Validar correlação de eventos | Sistema instrumentado com OpenTelemetry | 1. Gerar eventos em diferentes componentes<br>2. Verificar correlação através de trace ID | Eventos correlacionados corretamente pelo trace ID |
| OB-RT-003 | Verificar atributos específicos para hooks | Sistema instrumentado com OpenTelemetry | 1. Processar operações de diferentes hooks<br>2. Verificar atributos nos spans | Atributos específicos do hook presentes nos spans |
| OB-RT-004 | Validar spans para operações críticas | Sistema instrumentado com OpenTelemetry | 1. Executar operações críticas<br>2. Verificar presença de spans dedicados | Spans dedicados para cada operação crítica |
| OB-RT-005 | Verificar instrumentação por mercado | Sistema instrumentado com OpenTelemetry | 1. Processar operações para diferentes mercados<br>2. Verificar atributos específicos por mercado | Atributos de mercado presentes nos spans |

#### 1.2 Integração com Sistemas de APM

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| OB-APM-001 | Validar integração com Jaeger | Sistema configurado com Jaeger | 1. Executar operações rastreadas<br>2. Verificar visualização no Jaeger | Traces visualizados corretamente no Jaeger |
| OB-APM-002 | Validar integração com Zipkin | Sistema configurado com Zipkin | 1. Executar operações rastreadas<br>2. Verificar visualização no Zipkin | Traces visualizados corretamente no Zipkin |
| OB-APM-003 | Verificar agregação de traces | Sistema com APM | 1. Executar múltiplas operações similares<br>2. Verificar agregação no APM | Operações agregadas corretamente para análise |
| OB-APM-004 | Validar detecção de anomalias | Sistema com APM | 1. Simular operação anômala<br>2. Verificar detecção pelo APM | Anomalias detectadas e alertadas corretamente |
| OB-APM-005 | Verificar correlação com métricas | Sistema com APM | 1. Executar operação com impacto em métricas<br>2. Verificar correlação no APM | Traces correlacionados com métricas relevantes |

### 2. Logging Estruturado

#### 2.1 Logging com Zap

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| OB-LG-001 | Verificar estrutura de logs | Sistema configurado com Zap | 1. Gerar logs para diferentes operações<br>2. Verificar formato JSON estruturado | Logs gerados em formato JSON estruturado |
| OB-LG-002 | Validar campos obrigatórios | Sistema configurado com Zap | 1. Gerar logs para operações críticas<br>2. Verificar presença de campos obrigatórios | Todos os campos obrigatórios presentes nos logs |
| OB-LG-003 | Verificar correlação com traces | Sistema com logging e tracing | 1. Executar operação rastreada<br>2. Verificar inclusão de trace ID nos logs | Trace ID incluído em todos os logs da operação |
| OB-LG-004 | Validar campos específicos por mercado | Sistema configurado com Zap | 1. Gerar logs para diferentes mercados<br>2. Verificar campos específicos por mercado | Campos regulatórios específicos por mercado presentes |
| OB-LG-005 | Verificar níveis de log apropriados | Sistema configurado com Zap | 1. Executar operações de diferentes criticidades<br>2. Verificar níveis de log utilizados | Níveis de log apropriados para cada tipo de operação |

#### 2.2 Integração com Sistemas de Log

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| OB-SL-001 | Validar envio para Elasticsearch | Sistema integrado com ELK | 1. Gerar logs<br>2. Verificar indexação no Elasticsearch | Logs indexados corretamente no Elasticsearch |
| OB-SL-002 | Verificar visualização no Kibana | Sistema integrado com ELK | 1. Gerar logs<br>2. Verificar dashboards no Kibana | Logs visualizados corretamente nos dashboards |
| OB-SL-003 | Validar busca e filtragem | Sistema integrado com ELK | 1. Gerar logs diversos<br>2. Executar buscas e filtros | Buscas e filtros funcionando corretamente |
| OB-SL-004 | Verificar alertas baseados em logs | Sistema integrado com ELK | 1. Gerar logs de erro<br>2. Verificar acionamento de alertas | Alertas acionados conforme configuração |
| OB-SL-005 | Validar retenção e arquivamento | Sistema integrado com ELK | 1. Verificar política de retenção<br>2. Verificar arquivamento de logs antigos | Logs retidos e arquivados conforme política |

### 3. Métricas e Alertas

#### 3.1 Métricas com Prometheus

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| OB-MT-001 | Verificar métricas de solicitações | Sistema instrumentado com Prometheus | 1. Gerar solicitações<br>2. Verificar métricas de contador | Contador incrementado corretamente para solicitações |
| OB-MT-002 | Validar métricas de aprovação | Sistema instrumentado com Prometheus | 1. Processar aprovações<br>2. Verificar métricas de contador | Contador incrementado corretamente para aprovações |
| OB-MT-003 | Verificar métricas de validação | Sistema instrumentado com Prometheus | 1. Processar validações<br>2. Verificar métricas de contador | Contador incrementado corretamente para validações |
| OB-MT-004 | Validar métricas de latência | Sistema instrumentado com Prometheus | 1. Executar operações<br>2. Verificar histograma de latência | Histograma registrando latências corretamente |
| OB-MT-005 | Verificar métricas específicas por mercado | Sistema instrumentado com Prometheus | 1. Processar operações por mercado<br>2. Verificar labels específicos | Métricas segmentadas corretamente por mercado |

#### 3.2 Alertas e Dashboards

| ID | Descrição | Pré-condições | Passos | Resultados Esperados |
|----|-----------|---------------|--------|---------------------|
| OB-AD-001 | Validar alertas de latência | Sistema com Prometheus e Alertmanager | 1. Simular latência elevada<br>2. Verificar acionamento de alerta | Alerta de latência acionado conforme limites |
| OB-AD-002 | Verificar alertas de erro | Sistema com Prometheus e Alertmanager | 1. Simular taxa de erro elevada<br>2. Verificar acionamento de alerta | Alerta de erro acionado conforme limites |
| OB-AD-003 | Validar dashboards operacionais | Sistema com Grafana | 1. Gerar operações diversas<br>2. Verificar visualização em dashboards | Dashboards operacionais mostrando dados corretamente |
| OB-AD-004 | Verificar dashboards por mercado | Sistema com Grafana | 1. Gerar operações por mercado<br>2. Verificar dashboards específicos | Dashboards por mercado mostrando dados relevantes |
| OB-AD-005 | Validar dashboards de conformidade | Sistema com Grafana | 1. Gerar operações regulatórias<br>2. Verificar dashboards de conformidade | Dashboards de conformidade mostrando métricas regulatórias |

## Matriz de Teste Multi-Mercado

### 1. Matriz de Teste por Mercado e Hook

| Mercado | Docker Hook | GitHub Hook | Desktop Commander Hook | Figma Hook |
|---------|-------------|------------|----------------------|------------|
| Angola | CR-ANG-001 a CR-ANG-005 | CR-ANG-001 a CR-ANG-005 | CR-ANG-001 a CR-ANG-005 | CR-ANG-001 a CR-ANG-005 |
| Moçambique | CR-MOZ-001 a CR-MOZ-005 | CR-MOZ-001 a CR-MOZ-005 | CR-MOZ-001 a CR-MOZ-005 | CR-MOZ-001 a CR-MOZ-005 |
| Brasil | CR-BR-001 a CR-BR-010 | CR-BR-001 a CR-BR-010 | CR-BR-001 a CR-BR-010 | CR-BR-001 a CR-BR-010 |
| UE | CR-EU-001 a CR-EU-010 | CR-EU-001 a CR-EU-010 | CR-EU-001 a CR-EU-010 | CR-EU-001 a CR-EU-010 |
| EUA | CR-US-001 a CR-US-005 | CR-US-001 a CR-US-005 | CR-US-001 a CR-US-005 | CR-US-001 a CR-US-005 |
| China | CR-CN-001 a CR-CN-005 | CR-CN-001 a CR-CN-005 | CR-CN-001 a CR-CN-005 | CR-CN-001 a CR-CN-005 |
| Global | CR-GL-001 a CR-GL-005 | CR-GL-001 a CR-GL-005 | CR-GL-001 a CR-GL-005 | CR-GL-001 a CR-GL-005 |

### 2. Priorização de Testes por Mercado

| Mercado | Prioridade | Testes Críticos | Considerações Especiais |
|---------|------------|----------------|------------------------|
| Angola | Alta | CR-ANG-001, CR-ANG-002, CR-ANG-003 | Requisitos BNA para operações financeiras |
| Moçambique | Alta | CR-MOZ-001, CR-MOZ-003, CR-MOZ-004 | Requisitos para operações transfronteiriças SADC |
| Brasil | Alta | CR-BR-001, CR-BR-002, CR-BR-006, CR-BR-007 | LGPD e requisitos BACEN para PIX |
| UE | Alta | CR-EU-001, CR-EU-002, CR-EU-006 | GDPR e PSD2 para autenticação forte |
| EUA | Média | CR-US-001, CR-US-002 | SOX e requisitos de auditoria |
| China | Média | CR-CN-001, CR-CN-005 | Requisitos de localização de dados |
| Global | Alta | CR-GL-001, CR-GL-003 | AML/CFT e sanções internacionais |

## Automação de Testes

### 1. Estratégia de Automação

#### 1.1 Cobertura de Automação

| Tipo de Teste | % Alvo de Automação | Ferramenta Principal | Observações |
|---------------|---------------------|---------------------|-------------|
| Unitários | 95% | Go Testing, Testify | Cobertura quase completa com mocks |
| Integração | 80% | Testcontainers, GoConvey | Foco nas integrações críticas |
| API | 90% | Postman, Newman | Automação de todas as APIs públicas |
| Conformidade | 70% | Gherkin, Cucumber | Cenários chave para regulamentações |
| Desempenho | 75% | k6, Gatling | Testes de carga e picos automatizados |
| Segurança | 60% | OWASP ZAP, Gosec | Automação de verificações padrão |
| Observabilidade | 50% | Scripts personalizados | Verificação de logs e métricas |

#### 1.2 Abordagem CI/CD

| Fase | Testes Executados | Critérios de Passagem | Duração Estimada |
|------|-------------------|----------------------|------------------|
| Commit | Unitários, Linting | 100% de passagem | < 5 minutos |
| Pull Request | Unitários, Integração, Segurança SAST | 90% de passagem | < 15 minutos |
| Build | Unitários, Integração, API | 95% de passagem | < 30 minutos |
| Ambiente QA | Todos exceto Desempenho | 90% de passagem | < 2 horas |
| Pré-Produção | Todos incluindo Desempenho | 95% de passagem | < 4 horas |
| Produção (Smoke) | Subset de Integração e API | 100% de passagem | < 10 minutos |

### 2. Framework de Teste Automatizado

#### 2.1 Estrutura de Testes Unitários

```go
func TestDockerHook_ValidateScope(t *testing.T) {
    tests := []struct {
        name     string
        market   string
        tenantID string
        scope    string
        wantErr  bool
    }{
        {
            name:     "valid_docker_run",
            market:   "angola",
            tenantID: "tenant123",
            scope:    "docker:run",
            wantErr:  false,
        },
        {
            name:     "invalid_scope",
            market:   "angola",
            tenantID: "tenant123",
            scope:    "docker:invalid",
            wantErr:  true,
        },
        // Testes específicos por mercado
        {
            name:     "angola_specific_validation",
            market:   "angola",
            tenantID: "tenant123",
            scope:    "docker:run",
            wantErr:  false,
        },
        // Mais casos de teste...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup de mocks e contexto
            hook := NewDockerHook()
            ctx := context.Background()
            
            // Execução
            _, err := hook.ValidateScope(ctx, tt.scope, tt.tenantID, tt.market)
            
            // Verificação
            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

#### 2.2 Estrutura de Testes de Integração

```go
func TestIntegration_ElevationService(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Setup de containers para teste
    postgres, err := testcontainers.GenericContainer(...)
    require.NoError(t, err)
    defer postgres.Terminate(context.Background())
    
    // Setup do serviço
    service := setupElevationService(t, postgres)
    
    // Teste de fluxo completo
    t.Run("complete_elevation_flow", func(t *testing.T) {
        // 1. Solicitar elevação
        req := &ElevationRequest{...}
        resp, err := service.RequestElevation(context.Background(), req)
        require.NoError(t, err)
        
        // 2. Aprovar solicitação
        approvalReq := &ApprovalRequest{...}
        approvalResp, err := service.ApproveRequest(context.Background(), approvalReq)
        require.NoError(t, err)
        
        // 3. Validar uso de token
        validationReq := &ValidationRequest{...}
        validationResp, err := service.ValidateToken(context.Background(), validationReq)
        require.NoError(t, err)
        
        // Verificações
        require.NotEmpty(t, resp.RequestID)
        require.NotEmpty(t, approvalResp.TokenID)
        require.True(t, validationResp.Valid)
    })
}
```

## Ambiente de Testes

### 1. Ambientes Requeridos

| Ambiente | Propósito | Configuração | Dados |
|----------|-----------|--------------|-------|
| Desenvolvimento | Testes de desenvolvedores | Containers locais | Dados sintéticos |
| Integração | Testes de integração automatizados | Kubernetes dev-cluster | Dados sintéticos |
| QA | Testes funcionais e manuais | Kubernetes qa-cluster | Dados sintéticos + subset de produção anonimizados |
| Pré-Produção | Validação final | Espelho de produção | Réplica de produção anonimizada |
| Produção | Smoke tests e monitoramento | Ambiente produtivo | Dados reais |

### 2. Configuração Multi-Mercado

| Mercado | Configuração Específica | Dados de Teste |
|---------|-------------------------|----------------|
| Angola | Regras BNA configuradas | Dados de teste específicos BNA |
| Moçambique | Regras Banco de Moçambique configuradas | Dados de teste específicos Banco de Moçambique |
| Brasil | Regras LGPD e BACEN configuradas | Dados de teste específicos LGPD e BACEN |
| UE | Regras GDPR e PSD2 configuradas | Dados de teste específicos GDPR e PSD2 |
| EUA | Regras SOX configuradas | Dados de teste específicos SOX |
| China | Regras de localização configuradas | Dados de teste específicos regulatórios chineses |
| Global | Regras AML/CFT configuradas | Dados de teste para listas de sanções |

## Critérios de Aceitação

### 1. Critérios Funcionais

| ID | Critério | Verificação |
|----|---------|-------------|
| CA-F-001 | Todos os hooks devem validar corretamente escopos específicos | Testes unitários e integração passando |
| CA-F-002 | MFA deve ser exigido conforme configuração por mercado | Testes de conformidade regulatória passando |
| CA-F-003 | Fluxos de aprovação devem funcionar para todos os níveis | Testes de integração e cenários multi-nível passando |
| CA-F-004 | Validação de uso de token deve implementar todas as regras | Testes de validação de token passando para todos os casos |
| CA-F-005 | Auditoria completa deve ser gerada para todas as operações | Verificação de logs e eventos de auditoria |

### 2. Critérios Não-Funcionais

| ID | Critério | Verificação |
|----|---------|-------------|
| CA-NF-001 | Tempo de resposta para validação < 100ms (P95) | Testes de desempenho passando |
| CA-NF-002 | Sistema deve suportar 200 validações/segundo | Testes de carga passando |
| CA-NF-003 | Cobertura de código > 85% | Relatório de cobertura de testes |
| CA-NF-004 | Zero vulnerabilidades críticas ou altas | Relatórios de segurança |
| CA-NF-005 | Conformidade com requisitos regulatórios por mercado | Testes de conformidade passando |

### 3. Critérios de Observabilidade

| ID | Critério | Verificação |
|----|---------|-------------|
| CA-O-001 | 100% das operações devem ser rastreáveis | Verificação de traces em ferramentas APM |
| CA-O-002 | Todos os logs devem seguir formato estruturado | Verificação de logs em ferramentas de log |
| CA-O-003 | Métricas devem ser exportadas para todos os componentes | Verificação em dashboards Prometheus/Grafana |
| CA-O-004 | Alertas devem ser configurados para condições críticas | Verificação de configurações de alerta |
| CA-O-005 | Dashboards operacionais devem estar disponíveis | Verificação visual dos dashboards |

## Anexos

1. Modelo de Dados de Teste
2. Casos de Teste Detalhados por Hook
3. Configurações Específicas por Mercado
4. Matriz Completa de Requisitos Regulatórios
5. Scripts de Automação de Teste