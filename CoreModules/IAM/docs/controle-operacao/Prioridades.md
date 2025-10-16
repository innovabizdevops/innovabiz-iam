# Prioridades do Módulo IAM - INNOVABIZ

## Visão Geral

Este documento estabelece o sistema de priorização para o módulo IAM (Identity and Access Management) da plataforma INNOVABIZ, definindo critérios, níveis e processos para classificação de solicitações, incidentes, desenvolvimento e outras atividades relacionadas ao gerenciamento de identidades e acessos. A estrutura de prioridades garante alocação adequada de recursos e atenção proporcional à criticidade de cada item.

## Estrutura de Prioridades

### Níveis de Prioridade

| Nível | Nome | Descrição | Tempo de Resposta | Tempo de Resolução | Escalonamento |
|-------|------|-----------|-------------------|---------------------|---------------|
| P1 | Crítica | Impacto severo em operações críticas de negócio ou conformidade regulatória. Afeta múltiplos usuários ou tenants. Risco significativo de segurança ou vazamento de dados. | Imediato (< 15 min) | 4 horas | Automático para CISO e CIO após 1 hora |
| P2 | Alta | Impacto significativo em operações de negócio ou função de segurança importante. Afeta um grupo de usuários ou funções críticas específicas. | < 30 min | 8 horas | Automático para Gestor de Segurança após 2 horas |
| P3 | Média | Impacto moderado em operações normais. Soluções alternativas disponíveis. Afeta funcionalidades não-críticas ou um número limitado de usuários. | < 2 horas | 24 horas | Escalonamento manual se não resolvido em 12 horas |
| P4 | Baixa | Impacto mínimo em operações. Problema cosmético ou de conveniência. Não afeta funcionalidade principal. | < 8 horas | 72 horas | Sem escalonamento automático |
| P5 | Planejada | Melhorias, atualizações planejadas ou tarefas de manutenção sem impacto imediato. | 24 horas | Conforme planejamento | Sem escalonamento automático |

### Fatores de Priorização

| Fator | Descrição | Peso | Exemplos |
|-------|-----------|------|----------|
| Impacto no Negócio | Grau em que a operação normal do negócio é afetada | Alto | Autenticação indisponível (Alto), Relatório atrasado (Baixo) |
| Impacto na Segurança | Potencial de violação de segurança ou exposição de dados | Alto | Bypass de controle de acesso (Alto), Log incompleto (Médio) |
| Escopo | Número de usuários, sistemas ou tenants afetados | Médio | Sistema completo (Alto), Usuário único (Baixo) |
| Compliance | Risco de violação de requisitos regulatórios | Alto | Violação de LGPD (Alto), Documentação incompleta (Baixo) |
| Criticidade da Função | Importância da função afetada para operações | Médio | Autenticação MFA (Alto), Customização de perfil (Baixo) |
| Disponibilidade de Workaround | Existência de solução alternativa temporária | Baixo | Sem alternativa (Alto), Alternativa simples disponível (Baixo) |
| SLA Contratual | Acordos de nível de serviço estabelecidos | Médio | Viola SLA (Alto), Dentro de SLA (Baixo) |
| Reputação | Potencial impacto na reputação da organização | Médio | Visível externamente (Alto), Apenas interno (Baixo) |

## Prioridades por Categoria

### Prioridades para Incidentes de Segurança

| ID | Tipo de Incidente | Prioridade Padrão | Critérios de Elevação | Resposta |
|----|-------------------|-------------------|----------------------|----------|
| SEC-01 | Violação confirmada de dados | P1 | - | Ativação de plano de resposta a incidentes, notificação imediata ao CISO e DPO |
| SEC-02 | Tentativa de acesso não autorizado | P2 | Elevado para P1 se em massa ou bem-sucedido | Bloqueio de conta, análise de logs, revisão de controles |
| SEC-03 | Uso indevido de credenciais | P2 | Elevado para P1 se conta privilegiada | Suspensão de credenciais, investigação imediata |
| SEC-04 | Malware/Código malicioso | P1 | - | Isolamento do sistema, análise forense, contenção |
| SEC-05 | Vulnerabilidade de dia zero | P1 | - | Patch emergencial, controles compensatórios imediatos |
| SEC-06 | Vulnerabilidade descoberta | P2/P3 | Elevado conforme CVSS | Análise de impacto, planejamento de remediação |
| SEC-07 | Phishing direcionado | P2 | Elevado para P1 se massivo | Bloqueio de comunicação, alerta aos usuários |
| SEC-08 | Anomalia de comportamento | P3 | Elevado para P2 se persistente | Investigação, monitoramento adicional |
| SEC-09 | Violação de política de segurança | P3 | Elevado conforme severidade | Notificação, avaliação de treinamento, medidas disciplinares |
| SEC-10 | Perda de dispositivo | P2 | Elevado para P1 se com dados críticos | Apagamento remoto, rotação de credenciais, avaliação de exposição |

### Prioridades para Gestão de Identidades

| ID | Tipo de Solicitação | Prioridade Padrão | Critérios de Elevação | SLA |
|----|---------------------|-------------------|----------------------|-----|
| IDM-01 | Bloqueio de contas por suspeita de comprometimento | P1 | - | 15 minutos |
| IDM-02 | Criação de conta para novo colaborador | P3 | P2 se posição executiva ou crítica | 24 horas |
| IDM-03 | Reativação de conta bloqueada | P3 | P2 se usuário crítico | 8 horas |
| IDM-04 | Desativação por desligamento | P2 | P1 se desligamento por violação de segurança | 4 horas |
| IDM-05 | Solicitação de reset de senha | P3 | P2 se múltiplos usuários afetados | 4 horas |
| IDM-06 | Bloqueio por tentativas repetidas de login | P3 | P2 se em massa | 8 horas |
| IDM-07 | Problemas com MFA | P3 | P2 se usuário crítico ou múltiplos usuários | 4 horas |
| IDM-08 | Sincronização com diretório externo | P3 | P2 se falha completa de sincronização | 8 horas |
| IDM-09 | Atualização em massa de identidades | P4 | P3 se afeta atributos críticos | 24 horas |
| IDM-10 | Importação de usuários | P4 | P3 se cronograma crítico de projeto | 24 horas |

### Prioridades para Controle de Acesso

| ID | Tipo de Solicitação | Prioridade Padrão | Critérios de Elevação | SLA |
|----|---------------------|-------------------|----------------------|-----|
| ACC-01 | Falha em política de acesso crítica | P1 | - | 2 horas |
| ACC-02 | Solicitação de acesso privilegiado | P3 | P2 se para sistema crítico | 8 horas |
| ACC-03 | Revisão de acesso emergencial | P2 | P1 se relacionado a incidente | 4 horas |
| ACC-04 | Revogação de acesso | P3 | P1 se por comprometimento, P2 se por desligamento | 8 horas |
| ACC-05 | Erro em mapeamento de federação | P2 | P1 se afeta autenticação de todos usuários federados | 4 horas |
| ACC-06 | Atualização de política de acesso | P3 | P2 se política de segurança crítica | 12 horas |
| ACC-07 | Criação/atualização de papel (role) | P4 | P3 se papel crítico | 24 horas |
| ACC-08 | Auditoria de permissões | P4 | P3 se exigido por regulação | 48 horas |
| ACC-09 | Conflito de segregação de funções | P3 | P2 se em função financeira ou crítica | 12 horas |
| ACC-10 | Acesso temporário para contingência | P2 | P1 se situação de desastre declarada | 4 horas |

### Prioridades para Governança e Compliance

| ID | Tipo de Solicitação | Prioridade Padrão | Critérios de Elevação | SLA |
|----|---------------------|-------------------|----------------------|-----|
| GOV-01 | Falha em validador de compliance crítico | P2 | P1 se deadline regulatório próximo | 8 horas |
| GOV-02 | Geração de relatório regulatório | P3 | P2 se próximo de deadline legal | 24 horas |
| GOV-03 | Configuração de novo validador | P4 | P3 se relacionado a nova regulação | 48 horas |
| GOV-04 | Atualização de política por mudança regulatória | P3 | P2 se prazo legal próximo | 24 horas |
| GOV-05 | Requisição de direito de titular de dados | P3 | P2 se reclamação formal ou prazo legal curto | 24 horas |
| GOV-06 | Configuração de tenant para compliance regional | P3 | P2 se operação em nova região iminente | 48 horas |
| GOV-07 | Discrepância em relatório de compliance | P3 | P2 se em auditoria ativa | 24 horas |
| GOV-08 | Análise de risco para nova funcionalidade | P4 | P3 se funcionalidade crítica | 72 horas |
| GOV-09 | Configuração de retenção de dados | P3 | P2 se risco legal identificado | 48 horas |
| GOV-10 | Avaliação de impacto de privacidade | P4 | P3 se para sistema crítico | 72 horas |

### Prioridades para Desenvolvimento e Melhorias

| ID | Tipo de Solicitação | Prioridade Padrão | Critérios de Elevação | Cronograma |
|----|---------------------|-------------------|----------------------|------------|
| DEV-01 | Correção de vulnerabilidade de segurança | P2 | P1 se crítica ou explorada ativamente | Próxima release ou patch emergencial |
| DEV-02 | Implementação de requisito regulatório | P2 | P1 se prazo regulatório curto | Conforme prazo regulatório |
| DEV-03 | Melhoria de performance | P3 | P2 se impacto operacional significativo | Próximo ciclo de release |
| DEV-04 | Nova funcionalidade | P4 | P3 se requisito crítico de negócio | Roadmap padrão |
| DEV-05 | Melhoria de usabilidade | P4 | P3 se afeta muitos usuários | Roadmap padrão |
| DEV-06 | Refatoração técnica | P4 | P3 se dívida técnica crítica | Conforme disponibilidade |
| DEV-07 | Integração com novo sistema | P3 | P2 se sistema crítico ou prazo contratual | Alinhado com cronograma do projeto |
| DEV-08 | Atualização de dependências | P3 | P2 se contém correções de segurança | Ciclo regular de manutenção |
| DEV-09 | Novas APIs | P4 | P3 se requisito crítico de integração | Roadmap padrão |
| DEV-10 | Suporte a novo método de autenticação | P3 | P2 se exigido por política corporativa | Próximo release principal |

## Prioridades Específicas por Setor

### Setor de Saúde

| ID | Cenário | Prioridade | Justificativa |
|----|---------|-----------|---------------|
| HLT-01 | Acesso emergencial a dados de paciente (break-glass) | P1 | Pode afetar atendimento médico de emergência |
| HLT-02 | Validação de credenciais de profissional de saúde | P2 | Necessário para operação segura e conformidade |
| HLT-03 | Integração com sistemas EHR/EMR | P2 | Crítico para continuidade de cuidados |
| HLT-04 | Falha em controle de acesso a dados sensíveis de saúde | P1 | Risco regulatório severo (HIPAA, LGPD, PNDSB) |
| HLT-05 | Gestão de consentimento para compartilhamento | P2 | Requisito legal e ético |
| HLT-06 | Validador de compliance para regulação de saúde | P2 | Mitigação de riscos regulatórios |
| HLT-07 | Acesso para pesquisa a dados anonimizados | P3 | Importante, mas geralmente planejável |
| HLT-08 | Federação com sistema nacional de saúde | P2 | Requisito operacional e regulatório |
| HLT-09 | Segurança para telemedicina | P2 | Crítico para atendimento remoto seguro |
| HLT-10 | Relatórios de auditoria para autoridades de saúde | P2 | Requisito legal |

### Setor Financeiro

| ID | Cenário | Prioridade | Justificativa |
|----|---------|-----------|---------------|
| FIN-01 | Controles para acessos a sistemas de pagamento | P1 | Risco financeiro direto e regulatório |
| FIN-02 | Falha em segregação de funções financeiras | P1 | Risco de fraude e compliance |
| FIN-03 | MFA para transações de alto valor | P1 | Requisito de segurança crítico |
| FIN-04 | Integração com sistemas KYC/AML | P2 | Requisito regulatório |
| FIN-05 | Auditoria de acessos para SOX | P2 | Compliance financeiro |
| FIN-06 | Validadores PCI-DSS | P2 | Requisito para processamento de pagamentos |
| FIN-07 | Federação com parceiros financeiros | P2 | Importante para operação inter-bancária |
| FIN-08 | Gestão de identidade para Open Banking/Finance | P2 | Requisito regulatório emergente |
| FIN-09 | Revisão de acesso privilegiado a sistemas financeiros | P2 | Mitigação de risco de fraude interna |
| FIN-10 | Autenticação para aprovadores financeiros | P2 | Controle crítico para governança |

### Tecnologias AR/VR

| ID | Cenário | Prioridade | Justificativa |
|----|---------|-----------|---------------|
| ARV-01 | Segurança de dados espaciais sensíveis | P2 | Proteção de privacidade em novos contextos |
| ARV-02 | Autenticação em ambientes imersivos | P2 | Fundamental para experiências seguras |
| ARV-03 | Gestão de zonas de privacidade espacial | P2 | Controle de acesso em ambientes compartilhados |
| ARV-04 | Proteção de dados biométricos e perceptuais | P1 | Dados altamente sensíveis com proteção especial |
| ARV-05 | Federação de identidade em plataformas AR/VR | P3 | Interoperabilidade importante, mas não crítica |
| ARV-06 | Monitoramento de comportamento em AR/VR | P3 | Detecção de anomalias em novos contextos |
| ARV-07 | Controles para conteúdo sensível em AR | P2 | Prevenção de exposição inadequada |
| ARV-08 | Validação de dispositivos AR/VR seguros | P3 | Controle de endpoint específico |
| ARV-09 | Gestão de consentimento para dados perceptuais | P2 | Requisito de privacidade emergente |
| ARV-10 | Segurança para âncoras espaciais compartilhadas | P3 | Proteção de recursos compartilhados |

## Prioridades por Região

### União Europeia (GDPR)

| ID | Cenário | Prioridade | Justificativa |
|----|---------|-----------|---------------|
| EU-01 | Violação de dados com notificação obrigatória | P1 | Prazo legal de 72 horas |
| EU-02 | Requisição de direito de portabilidade | P3 | Prazo legal de 30 dias |
| EU-03 | Requisição de direito ao esquecimento | P3 | Prazo legal de 30 dias, elevado se formal |
| EU-04 | Falha na gestão de consentimento | P2 | Base crítica para processamento legal |
| EU-05 | Validação de mecanismos de transferência internacional | P2 | Após invalidação de Privacy Shield |
| EU-06 | Relatório para autoridade de proteção | P2 | Compliance regulatório formal |
| EU-07 | Configuração de minimização de dados | P3 | Princípio fundamental, mas prazo flexível |
| EU-08 | Avaliação de impacto de privacidade | P3 | Requisito para novos processamentos |
| EU-09 | Revisão de acessos a dados pessoais | P3 | Controle regular recomendado |
| EU-10 | Implementação de privacy by design | P3 | Para novos recursos ou sistemas |

### Brasil (LGPD)

| ID | Cenário | Prioridade | Justificativa |
|----|---------|-----------|---------------|
| BR-01 | Incidente com dados pessoais | P1 | Notificação à ANPD em prazo razoável |
| BR-02 | Requisição de titular de dados | P3 | Prazo legal de 15 dias |
| BR-03 | Revisão de base legal para processamento | P2 | Fundamental para processamento legal |
| BR-04 | Configuração de relatório para ANPD | P2 | Compliance regulatório formal |
| BR-05 | Adequação de contratos com operadores | P3 | Requisito legal, prazo flexível |
| BR-06 | Implementação de RIPD | P3 | Para tratamentos de alto risco |
| BR-07 | Revisão de mecanismos de consentimento | P2 | Base fundamental para compliance |
| BR-08 | Configuração de relatório de impacto | P3 | Documentação preventiva |
| BR-09 | Implementação de controles para dados sensíveis | P2 | Proteção especial exigida |
| BR-10 | Gestão de encarregado (DPO) | P3 | Requisito organizacional |

### Estados Unidos (Múltiplas regulações)

| ID | Cenário | Prioridade | Justificativa |
|----|---------|-----------|---------------|
| US-01 | Violação de PHI (HIPAA) | P1 | Requisitos rígidos de notificação |
| US-02 | Controles de segurança PCI DSS | P2 | Requisito para processamento de pagamentos |
| US-03 | Requisição de consumidor (CCPA/CPRA) | P3 | Prazo legal de resposta |
| US-04 | Revisão SOX para controles de acesso financeiros | P2 | Compliance financeiro mandatório |
| US-05 | Adequação a lei estadual específica | P3 | Variação por estado, complexidade |
| US-06 | BAA para parceiros HIPAA | P3 | Requisito contratual para saúde |
| US-07 | Avaliação de risco para HIPAA | P3 | Requisito periódico |
| US-08 | Configuração FedRAMP | P2 | Para serviços ao governo federal |
| US-09 | Validação FERPA para dados educacionais | P3 | Específico para setor educacional |
| US-10 | Controles para GLBA (financeiro) | P2 | Específico para setor financeiro |

### Angola (PNDSB)

| ID | Cenário | Prioridade | Justificativa |
|----|---------|-----------|---------------|
| AO-01 | Incidente com dados de saúde | P1 | Requisito para proteção de dados sensíveis |
| AO-02 | Integração com sistema nacional de saúde | P2 | Requisito operacional para saúde |
| AO-03 | Configuração de controles específicos regionais | P3 | Adaptação a requisitos locais |
| AO-04 | Revisão de transferência internacional de dados | P2 | Controle para fluxo transfronteiriço |
| AO-05 | Configuração de relatórios regulatórios | P3 | Conformidade regional |
| AO-06 | Validadores específicos para PNDSB | P3 | Validação de compliance |
| AO-07 | Adaptação de consentimento para contexto local | P3 | Adequação cultural e legal |
| AO-08 | Controles para dados de pesquisa em saúde | P3 | Requisitos específicos para pesquisa |
| AO-09 | Implementação de controles de acesso regionalizados | P3 | Adaptação a necessidades locais |
| AO-10 | Gestão de identidade para profissionais locais | P2 | Integração com sistemas regionais |
