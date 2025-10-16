# Requisitos Técnicos: Hooks MCP para Elevação de Privilégios

**Documento**: INNOVABIZ-IAM-REQ-MCP-HOOKS-v1.0.0  
**Classificação**: Confidencial  
**Data**: 06/08/2025  
**Estado**: Aprovado  
**Âmbito**: Global (Angola, Moçambique, Brasil, PALOP, SADC, BRICS, UE, EUA)

## Índice

1. [Introdução](#introdução)
2. [Visão Geral](#visão-geral)
3. [Requisitos Funcionais](#requisitos-funcionais)
4. [Requisitos Não-Funcionais](#requisitos-não-funcionais)
5. [Requisitos Regulatórios](#requisitos-regulatórios)
6. [Integração com Outros Módulos](#integração-com-outros-módulos)
7. [Arquitectura Técnica](#arquitectura-técnica)
8. [Casos de Uso](#casos-de-uso)
9. [Glossário](#glossário)

## Introdução

Este documento especifica os requisitos técnicos detalhados para a implementação do sistema de hooks MCP (Model Context Protocol) para o serviço de elevação de privilégios do módulo IAM da plataforma INNOVABIZ. Este sistema visa fornecer um mecanismo uniforme e conforme para autorizar, auditar e governar operações privilegiadas em serviços MCP integrados, incluindo Docker, GitHub, Desktop Commander e Figma.

### Objectivo

Estabelecer um framework de integração para serviços MCP que garanta:
- Conformidade regulatória específica por mercado
- Aplicação consistente de políticas de segurança
- Auditoria granular e completa de operações privilegiadas
- Flexibilidade para diferentes contextos operacionais
- Suporte multi-tenant e multi-mercado

### Âmbito

O sistema abrange os seguintes componentes:
- Interface padrão para hooks MCP
- Implementações específicas para Docker, GitHub, Desktop Commander e Figma
- Registo centralizado para gestão dinâmica de hooks
- Integração com serviço de elevação de privilégios
- Interfaces para sistemas de auditoria e observabilidade

## Visão Geral

O sistema de hooks MCP atua como intermediário entre o serviço central de elevação de privilégios e os diferentes serviços MCP integrados à plataforma INNOVABIZ. Cada hook implementa uma interface comum e fornece validação específica para seu serviço correspondente, garantindo que operações privilegiadas sejam devidamente autorizadas, validadas e auditadas.

### Principais Funcionalidades

- **Validação de Escopos**: Verificar se o escopo solicitado é válido para o serviço e mercado específico
- **Determinação de MFA**: Definir requisitos de MFA baseados na sensibilidade da operação
- **Avaliação de Aprovação**: Determinar se aprovação é necessária baseada no contexto
- **Validação de Solicitações**: Verificar integralidade e conformidade da solicitação
- **Determinação de Aprovadores**: Identificar aprovadores adequados para cada solicitação
- **Validação de Uso**: Verificar uso apropriado de tokens de elevação
- **Aplicação de Limites**: Aplicar limites de política baseados em mercado e tenant
- **Geração de Auditoria**: Produzir metadados padronizados para auditoria

## Requisitos Funcionais

### RF1: Interface MCPHook

O sistema deve fornecer uma interface padrão `MCPHook` que todos os hooks MCP devem implementar:

```go
type MCPHook interface {
    HookType() MCPHookType
    ValidateScope(ctx context.Context, scope string, tenantID string, market string) (*ScopeDetails, error)
    GetRequiredMFA(ctx context.Context, scope string, tenantID string, market string) (MFALevel, error)
    GetRequireApproval(ctx context.Context, scope string, tenantID string, market string) (bool, error)
    ValidateRequest(ctx context.Context, request *ElevationRequest) error
    GetApprovers(ctx context.Context, request *ElevationRequest) ([]string, error)
    ValidateElevationUse(ctx context.Context, tokenID string, scope string, metadata map[string]interface{}) error
    GetPolicyLimits(ctx context.Context, tenantID string, market string) (*PolicyLimits, error)
    GetAuditMetadata(ctx context.Context, tokenID string, scope string) (map[string]interface{}, error)
}
```

### RF2: Hooks Específicos por Serviço

O sistema deve fornecer implementações específicas da interface MCPHook para:

#### RF2.1: Docker
- Validação de operações Docker (run, exec, build, pull, push, network, volume)
- Aplicação de políticas específicas para imagens, comandos e redes sensíveis
- Mapeamento de sensibilidade de comandos Docker

#### RF2.2: GitHub
- Validação de operações GitHub (read, write, delete, PR, issues, security)
- Proteção de branches e repositórios sensíveis
- Mapeamento de operações GitHub para níveis de aprovação

#### RF2.3: Desktop Commander
- Validação de operações em arquivos e comandos locais
- Proteção de diretórios e arquivos sensíveis
- Validação de comandos específicos do sistema

#### RF2.4: Figma
- Validação de operações em designs e bibliotecas
- Proteção de projetos e bibliotecas sensíveis
- Mapeamento de operações de design para níveis de aprovação

### RF3: Registo Centralizado de Hooks

O sistema deve fornecer um registo centralizado (`HookRegistry`) para:
- Registrar hooks MCP disponíveis
- Recuperar hooks por tipo
- Validar escopos para todos os hooks registrados

### RF4: Configuração Específica por Mercado

O sistema deve suportar configurações específicas por mercado para cada hook:
- Níveis de MFA padrão
- Requisitos de aprovação
- Limites de duração de tokens
- Operações e recursos sensíveis
- Operações e recursos proibidos

### RF5: Modo de Emergência

O sistema deve suportar um modo de emergência com:
- Validação específica para solicitações emergenciais
- Restrições para operações permitidas em emergência
- Requisitos de documentação posterior
- Limites de duração específicos para emergências

### RF6: Auditoria e Telemetria

O sistema deve gerar metadados de auditoria padronizados para:
- Solicitações de elevação
- Uso de tokens de elevação
- Aprovações e rejeições
- Operações em modo de emergência

## Requisitos Não-Funcionais

### RNF1: Desempenho

- Tempo de resposta máximo para validação de escopo: 200ms
- Tempo de resposta máximo para validação completa de solicitação: 500ms
- Capacidade para processar no mínimo 100 solicitações por segundo

### RNF2: Disponibilidade

- Disponibilidade mínima de 99.95% para o serviço de hooks
- Degradação graciosa em caso de falha de componentes
- Tempo máximo de recuperação de 10 segundos

### RNF3: Segurança

- Todas as operações devem ser autenticadas e autorizadas
- Logging de todas as operações de elevação
- Criptografia em trânsito e em repouso para todos os dados sensíveis
- Validação de entrada para todos os parâmetros

### RNF4: Escalabilidade

- Suporte para expansão horizontal de componentes
- Capacidade para adicionar novos hooks MCP sem modificação do core
- Suporte para no mínimo 10 hooks MCP simultâneos

### RNF5: Manutenibilidade

- Código modular e bem documentado
- Testes unitários com cobertura mínima de 85%
- Testes de integração automatizados
- Documentação técnica completa

### RNF6: Observabilidade

- Integração com OpenTelemetry para rastreamento
- Logging estruturado com Zap
- Métricas Prometheus para monitoramento
- Alertas para eventos críticos e anômalos

## Requisitos Regulatórios

### RR1: Conformidade com GDPR (UE/EEE)

- Manter registro completo de todas as atividades de processamento
- Garantir que todas as operações tenham base legal clara
- Suportar limitação de acesso baseada em necessidade de conhecimento
- Implementar políticas de retenção de dados para logs e auditorias

### RR2: Conformidade com LGPD (Brasil)

- Logging específico para acesso a dados pessoais
- Suporte para aprovações específicas por DPO
- Registros detalhados de justificativas para operações em dados pessoais
- Mecanismos de relatório para autoridades reguladoras

### RR3: Conformidade com Regulações de Angola e Moçambique

- Aprovação dupla para operações sensíveis
- Políticas específicas para sistemas financeiros em mercados SADC/PALOP
- Requisitos MFA reforçados para operações críticas
- Registros de auditoria específicos para reguladores locais

### RR4: Conformidade com PCI-DSS

- Proteção reforçada para operações que envolvam dados de pagamento
- Segregação de funções para operações em sistemas de processamento de cartões
- Logs de auditoria específicos para operações em dados de pagamento
- Controles de acesso granulares baseados em necessidade de função

## Integração com Outros Módulos

### IM1: Serviço de Elevação IAM

- Integração com o serviço central de elevação de privilégios
- Utilização da API de tokens de elevação
- Integração com serviço de aprovação de solicitações

### IM2: Sistema de Identidades

- Integração com o sistema de identidades para validação de usuários
- Recuperação de informações de roles e grupos
- Verificação de estado de usuário (ativo, bloqueado, etc.)

### IM3: Serviço de MFA

- Integração com o serviço de autenticação multi-fator
- Validação de status de MFA para solicitações
- Suporte para diferentes níveis de MFA (básico, forte)

### IM4: Sistema de Auditoria

- Integração com sistema central de auditoria
- Envio de eventos de auditoria estruturados
- Suporte para consultas de histórico de elevações

### IM5: Observabilidade

- Integração com sistema de telemetria
- Rastreamento distribuído via OpenTelemetry
- Exportação de métricas para Prometheus
- Envio de logs estruturados para agregador central

### IM6: Payment Gateway

- Integração para validação de operações em serviços de pagamento
- Controles específicos para operações financeiras
- Regras específicas para mercados regulados financeiramente

### IM7: Risk Management

- Integração para avaliação de risco de solicitações
- Aplicação dinâmica de políticas baseadas em nível de risco
- Alertas para padrões suspeitos de elevação

## Arquitectura Técnica

### Componentes Principais

1. **Interface MCPHook**: Interface comum para todos os hooks MCP
2. **HookRegistry**: Registro central para hooks MCP
3. **Implementações Específicas**: Docker, GitHub, Desktop Commander, Figma
4. **Serviço de Configuração**: Gestão de configurações específicas por mercado
5. **Cliente de Telemetria**: Integração com OpenTelemetry e logging

### Fluxo de Dados

1. Solicitação de elevação recebida pelo serviço central
2. Identificação do hook MCP apropriado baseado no escopo
3. Validação do escopo e requisitos MFA
4. Determinação de necessidade de aprovação
5. Validação completa da solicitação
6. Geração de token de elevação (se aprovada)
7. Validação de uso do token durante operações
8. Geração de metadados de auditoria

## Casos de Uso

### CU1: Elevação para Operação Docker Sensível

**Ator**: Desenvolvedor  
**Descrição**: Um desenvolvedor precisa executar um container Docker com acesso à rede host.  

**Fluxo Principal**:
1. Desenvolvedor solicita elevação para escopo `docker:run`
2. Sistema identifica operação como sensível devido aos parâmetros
3. Sistema determina necessidade de MFA Básico e aprovação
4. Desenvolvedor completa MFA e fornece justificativa
5. Sistema encaminha para aprovador apropriado
6. Aprovador revisa e aprova solicitação
7. Token de elevação é gerado com duração limitada
8. Desenvolvedor utiliza token para executar container
9. Sistema valida uso do token e gera metadados de auditoria

### CU2: Elevação Emergencial para Acesso GitHub

**Ator**: Operador de Sistema  
**Descrição**: Um operador precisa acesso emergencial para merge em branch protegido para resolver incidente.

**Fluxo Principal**:
1. Operador solicita elevação emergencial para escopo `github:write`
2. Sistema determina que operação é permitida em modo emergencial
3. Operador completa MFA Forte conforme requerido
4. Sistema gera token com duração reduzida (1 hora)
5. Sistema registra solicitação emergencial para revisão posterior
6. Operador utiliza token para realizar merge
7. Sistema valida uso e gera metadados de auditoria detalhados
8. Operador fornece justificação posterior detalhada

## Glossário

- **MCP**: Model Context Protocol, protocolo para interação entre modelos de IA e sistemas externos
- **Hook**: Componente de integração que intercepta e processa solicitações específicas
- **Elevação**: Processo de obtenção de privilégios temporários adicionais
- **Token**: Credencial temporária que concede acesso elevado
- **Escopo**: Definição específica de privilégios solicitados
- **MFA**: Autenticação Multi-Fator, método de segurança que requer múltiplas formas de verificação
- **Tenant**: Organização ou entidade isolada dentro do sistema multi-tenant
- **Mercado**: Região geográfica com requisitos regulatórios específicos