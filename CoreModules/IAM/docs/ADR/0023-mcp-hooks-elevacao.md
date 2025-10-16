# ADR-0023: Implementação de Hooks MCP para Elevação de Privilégios

**Data**: 06/08/2025  
**Status**: Aprovado  
**Autores**: Equipa de Arquitetura INNOVABIZ  
**Mercados Alvo**: Global, com ênfase em Angola, Moçambique, Brasil, Europa, CPLP, SADC, PALOP, BRICS  
**Classificação**: Confidencial  
**Versão**: 1.0.0

## Contexto

A plataforma INNOVABIZ necessita de um mecanismo robusto e conforme para gestão de elevação de privilégios para operações MCP (Model Context Protocol), que permita integração segura com serviços externos como Docker, GitHub, Desktop Commander e Figma. Este mecanismo deve respeitar os requisitos regulatórios específicos de cada mercado, especialmente os mercados-alvo prioritários (Angola, Moçambique, Brasil), bem como oferecer segurança multi-camada através de MFA, aprovações dinâmicas e auditoria completa.

## Problema

O sistema atual não possui um mecanismo uniforme e conforme para autorizar e auditar operações de elevação de privilégios em serviços MCP, resultando em:

1. Inconsistência na aplicação de políticas de segurança entre diferentes serviços MCP
2. Dificuldade em garantir conformidade regulatória específica por mercado 
3. Falta de auditoria granular das operações privilegiadas
4. Ausência de mecanismos de aprovação flexíveis baseados em sensibilidade da operação
5. Implementação inconsistente de requisitos de MFA para operações sensíveis
6. Gestão inadequada de operações emergenciais com princípio de privilégio mínimo

## Decisão

Implementar um sistema extensível de hooks MCP para integração com o serviço de elevação de privilégios do módulo IAM da plataforma INNOVABIZ, com as seguintes características:

1. **Interface Genérica `MCPHook`**: Definir uma interface padrão que todos os hooks MCP devem implementar, incluindo métodos para:
   - Validação de escopos
   - Determinação de requisitos MFA
   - Avaliação de necessidade de aprovação
   - Validação de solicitações de elevação
   - Determinação de aprovadores baseado em contexto
   - Validação do uso do token de elevação
   - Aplicação de limites de política
   - Geração de metadados de auditoria

2. **Hooks Específicos por Serviço**:
   - Docker (`DockerHook`): Para operações de containers e imagens
   - GitHub (`GitHubHook`): Para operações em repositórios e código
   - Desktop Commander (`DesktopCommanderHook`): Para operações em arquivos e comandos locais
   - Figma (`FigmaHook`): Para operações em designs e bibliotecas

3. **Registo Centralizado de Hooks**: Implementar um registo que permita descoberta dinâmica e configuração centralizada dos hooks disponíveis.

4. **Multi-mercado e Multi-regulação**:
   - Configurações específicas por mercado (Angola, Moçambique, Brasil, etc.)
   - Adaptação a requisitos regulatórios específicos (GDPR, LGPD, regulações PALOP/SADC)
   - Níveis de aprovação e MFA ajustáveis por mercado e sensibilidade

5. **Integração com Observabilidade**:
   - OpenTelemetry para rastreamento distribuído
   - Zap para logging estruturado
   - Metadados de auditoria padronizados

## Considerações

### Conformidade Regulatória
- **Angola e Moçambique (SADC/PALOP)**: Implementação de requisitos específicos de aprovação dupla para operações sensíveis e conformidade com regulamentações locais de proteção de dados.
- **Brasil (LGPD)**: Implementação de requisitos de logging específicos para acesso a dados pessoais e mecanismos de aprovação específicos para DPO.
- **UE/EEE (GDPR)**: Garantia de conformidade com diretrizes de proteção de dados, incluindo rastreabilidade completa e justificativas explícitas.
- **Mercados BRICS**: Adaptação a requisitos específicos de China, Índia, Rússia e África do Sul.

### Segurança Multi-camada
- MFA adaptativo baseado em sensibilidade da operação (Nenhum, Básico, Forte)
- Aprovação multi-nível baseada em contexto
- Validação granular de escopo e justificativa
- Proteção específica para recursos e operações críticos por serviço MCP

### Escalabilidade e Extensibilidade
- Design modular permitindo adição de novos hooks MCP
- Configuração flexível permitindo ajustes sem alteração de código
- Interface padronizada facilitando integração com novos serviços

### Observabilidade Total
- Telemetria completa de todas as operações
- Logging estruturado para análise de eventos
- Metadados de auditoria padronizados para conformidade

## Consequências

### Positivas
- Implementação uniforme de políticas de segurança para operações privilegiadas
- Conformidade regulatória específica por mercado
- Auditoria granular e completa para todas operações privilegiadas
- Gestão flexível e contextual de aprovações
- Implementação consistente de MFA baseado em risco
- Gestão adequada de operações emergenciais

### Negativas
- Aumento da complexidade do sistema de autorização
- Necessidade de manter configurações específicas por mercado
- Potencial impacto em performance devido a validações adicionais

### Mitigações
- Implementação de caching de políticas e decisões frequentes
- Design para performance com validações em paralelo quando possível
- Instrumentação detalhada para identificar e resolver gargalos

## Métricas de Avaliação
- Taxa de sucesso/falha de solicitações de elevação por tipo de hook
- Tempo médio de processamento de solicitações
- Cobertura de conformidade regulatória por mercado
- Número de incidentes de segurança relacionados a operações privilegiadas
- Taxa de uso de modo emergencial vs normal

## Alternativas Consideradas

1. **Sistema de Permissões Fixas**: Rejeitado por não oferecer flexibilidade necessária para diferentes mercados e regulações.
2. **Implementação Específica por Serviço**: Rejeitado por criar duplicação de código e inconsistência na aplicação de políticas.
3. **Delegação Total para Serviços Externos**: Rejeitado por não garantir conformidade central e auditoria uniforme.
4. **Sistema de Permissões Baseado Apenas em Papéis**: Rejeitado por não oferecer granularidade necessária para operações específicas.

## Referências
- ISO/IEC 27001:2022 (Segurança da Informação)
- NIST SP 800-53 Rev. 5 (Controles de Segurança)
- RFC 6819 (Ameaças OAuth 2.0)
- Especificação OpenTelemetry
- GDPR (Regulamento Geral sobre a Proteção de Dados)
- LGPD (Lei Geral de Proteção de Dados Pessoais)
- Regulamentos de Proteção de Dados da Angola e Moçambique
- Padrões PCI-DSS para serviços financeiros
- OWASP ASVS v4.0.3 (Application Security Verification Standard)

## Aprovações

| Nome                | Papel                | Assinatura | Data       |
|--------------------|---------------------|-----------|------------|
| [Diretor de Tecnologia] | Aprovador        | [Assinatura] | 06/08/2025 |
| [Diretor de Segurança] | Aprovador        | [Assinatura] | 06/08/2025 |
| [Lead de Arquitetura]  | Autor/Revisor    | [Assinatura] | 06/08/2025 |
| [Compliance Officer]   | Revisor          | [Assinatura] | 06/08/2025 |