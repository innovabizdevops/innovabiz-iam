# Situações do Módulo IAM - INNOVABIZ

## Visão Geral

Este documento descreve as possíveis situações operacionais do módulo IAM (Identity and Access Management), incluindo condições especiais de funcionamento, cenários de uso e circunstâncias que requerem atenção operacional.

## Situações de Autenticação

### Situação: Tentativas Excessivas de Login

| Atributo | Descrição |
|----------|-----------|
| **Identificador** | `AUTH_EXCESSIVE_ATTEMPTS` |
| **Descrição** | Usuário realizou múltiplas tentativas de login sem sucesso |
| **Gatilhos** | Mais de 5 tentativas de login malsucedidas em 10 minutos |
| **Ações Automáticas** | 1. Bloqueio temporário da conta por 30 minutos<br>2. Notificação ao administrador<br>3. Registro no log de segurança |
| **Ações Manuais** | 1. Verificação de possível ataque de força bruta<br>2. Contato com usuário para verificação de identidade |
| **Resolução** | Desbloqueio manual por administrador ou automático após período de espera |
| **Métricas** | Frequência de ocorrência, distribuição geográfica, padrões de horário |

### Situação: Autenticação de Localização Incomum

| Atributo | Descrição |
|----------|-----------|
| **Identificador** | `AUTH_UNUSUAL_LOCATION` |
| **Descrição** | Login realizado de localização geográfica atípica para o usuário |
| **Gatilhos** | Login de país ou região diferente do padrão habitual do usuário |
| **Ações Automáticas** | 1. Solicitação de verificação adicional (2FA)<br>2. Notificação ao usuário<br>3. Limitação de acesso a operações sensíveis |
| **Ações Manuais** | 1. Verificação de viagem do usuário<br>2. Confirmação de identidade por canal alternativo |
| **Resolução** | Confirmação pelo usuário ou adição da nova localização à lista de confiança |
| **Métricas** | Taxa de falsos positivos, tempo de resolução, frequência por usuário |

### Situação: Falha em Provedor de Identidade Externo

| Atributo | Descrição |
|----------|-----------|
| **Identificador** | `AUTH_IDP_FAILURE` |
| **Descrição** | Provedor de identidade externo (SAML, OAuth, OIDC) está indisponível ou respondendo com erros |
| **Gatilhos** | Timeout, erro de conexão ou resposta de erro do IdP |
| **Ações Automáticas** | 1. Fallback para método de autenticação alternativo<br>2. Alertas de monitoramento<br>3. Registro detalhado do erro |
| **Ações Manuais** | 1. Contato com suporte do provedor externo<br>2. Ativação temporária de caminhos de autenticação alternativos |
| **Resolução** | Restauração do serviço do IdP ou migração para provedor alternativo |
| **Métricas** | Tempo de indisponibilidade, número de usuários afetados, taxa de sucesso do fallback |

## Situações de Autorização

### Situação: Escalação de Privilégios Detectada

| Atributo | Descrição |
|----------|-----------|
| **Identificador** | `AUTH_PRIVILEGE_ESCALATION` |
| **Descrição** | Detecção de aumento anormal de privilégios de uma conta |
| **Gatilhos** | Adição de permissões administrativas ou sensíveis sem processo aprovado |
| **Ações Automáticas** | 1. Bloqueio das novas permissões<br>2. Alerta de segurança crítico<br>3. Registro para auditoria forense |
| **Ações Manuais** | 1. Investigação imediata<br>2. Revisão de logs de atividade<br>3. Contenção de possível comprometimento |
| **Resolução** | Remoção de acessos não autorizados e remediação da vulnerabilidade |
| **Métricas** | Tempo de detecção, tempo de resposta, extensão do acesso não autorizado |

### Situação: Conflito de Segregação de Funções

| Atributo | Descrição |
|----------|-----------|
| **Identificador** | `AUTH_SOD_CONFLICT` |
| **Descrição** | Usuário possui combinação de papéis que viola princípios de segregação de funções |
| **Gatilhos** | Atribuição de novo papel que gera conflito com papéis existentes |
| **Ações Automáticas** | 1. Bloqueio da atribuição conflitante<br>2. Notificação ao gestor de compliance<br>3. Documentação do conflito |
| **Ações Manuais** | 1. Revisão das políticas de SoD<br>2. Avaliação de exceção com justificativa |
| **Resolução** | Remoção de um dos papéis conflitantes ou aprovação de exceção documentada |
| **Métricas** | Número de conflitos, tempo de resolução, recorrência por departamento |

## Situações de Gestão de Identidades

### Situação: Identidades Órfãs Detectadas

| Atributo | Descrição |
|----------|-----------|
| **Identificador** | `IDM_ORPHANED_IDENTITIES` |
| **Descrição** | Contas que permanecem ativas após desligamento ou transferência do usuário |
| **Gatilhos** | Sincronização com HR mostra status inativo mas conta permanece ativa |
| **Ações Automáticas** | 1. Desativação automática após 15 dias<br>2. Revogação de acessos críticos imediatamente<br>3. Inventário de recursos associados |
| **Ações Manuais** | 1. Verificação de transferência de responsabilidades<br>2. Aprovação para extensão em casos especiais |
| **Resolução** | Desativação da conta ou transferência formal para novo responsável |
| **Métricas** | Volume de contas órfãs, tempo médio até detecção, recursos acessíveis |

### Situação: Contas Privilegiadas sem Utilização

| Atributo | Descrição |
|----------|-----------|
| **Identificador** | `IDM_DORMANT_PRIVILEGED` |
| **Descrição** | Contas com privilégios elevados que não são utilizadas por período prolongado |
| **Gatilhos** | Conta admin sem login por mais de 30 dias |
| **Ações Automáticas** | 1. Notificação ao proprietário e gestor<br>2. Agendamento de revogação de privilégios<br>3. Auditoria de necessidade |
| **Ações Manuais** | 1. Confirmação da necessidade contínua dos privilégios<br>2. Documentação da justificativa |
| **Resolução** | Remoção dos privilégios ou confirmação de necessidade com nova data de revisão |
| **Métricas** | Quantidade de contas privilegiadas dormentes, taxa de redução, economia de risco |

## Situações de Multi-tenancy

### Situação: Vazamento de Dados entre Tenants

| Atributo | Descrição |
|----------|-----------|
| **Identificador** | `MT_DATA_LEAKAGE` |
| **Descrição** | Detecção de acesso a dados de um tenant por usuários de outro tenant |
| **Gatilhos** | Logs de acesso mostram operações cross-tenant não autorizadas |
| **Ações Automáticas** | 1. Bloqueio imediato do acesso<br>2. Alerta de segurança de alta prioridade<br>3. Snapshot do estado para investigação |
| **Ações Manuais** | 1. Análise forense do incidente<br>2. Verificação da política RLS<br>3. Notificação aos responsáveis pelos tenants |
| **Resolução** | Correção das políticas de isolamento e avaliação de impacto da exposição |
| **Métricas** | Volume de dados expostos, duração da exposição, impacto regulatório |

### Situação: Falha na Migração entre Tenants

| Atributo | Descrição |
|----------|-----------|
| **Identificador** | `MT_MIGRATION_FAILURE` |
| **Descrição** | Processo de migração de dados entre esquemas de tenants falhou |
| **Gatilhos** | Job de migração termina com erro ou inconsistência de dados |
| **Ações Automáticas** | 1. Rollback automático para estado anterior<br>2. Quarentena dos dados parcialmente migrados<br>3. Alerta operacional |
| **Ações Manuais** | 1. Análise de causa raiz<br>2. Correção dos scripts de migração<br>3. Planejamento de nova tentativa |
| **Resolução** | Migração bem-sucedida ou decisão documentada de não migrar |
| **Métricas** | Taxa de sucesso de migrações, tempo de resolução, impacto em disponibilidade |

## Situações de Compliance

### Situação: Violação de Política Regulatória

| Atributo | Descrição |
|----------|-----------|
| **Identificador** | `COMP_REGULATORY_VIOLATION` |
| **Descrição** | Configuração ou operação que viola requisitos regulatórios (GDPR, LGPD, etc.) |
| **Gatilhos** | Verificação de compliance falha ou denúncia específica |
| **Ações Automáticas** | 1. Limitação de processamento de dados afetados<br>2. Notificação ao DPO<br>3. Registro detalhado para investigação |
| **Ações Manuais** | 1. Análise de impacto regulatório<br>2. Implementação de correções<br>3. Comunicação com autoridades se necessário |
| **Resolução** | Correção da violação e documentação das medidas tomadas |
| **Métricas** | Tempo para resolução, potencial multa evitada, recorrência |

### Situação: Expiração Iminente de Certificação

| Atributo | Descrição |
|----------|-----------|
| **Identificador** | `COMP_CERT_EXPIRATION` |
| **Descrição** | Certificação de compliance está próxima do vencimento |
| **Gatilhos** | Menos de 60 dias para expiração de certificação SOC2, ISO27001, etc. |
| **Ações Automáticas** | 1. Alertas escalonados por proximidade da data<br>2. Geração de relatório de status de preparação<br>3. Registro em dashboard executivo |
| **Ações Manuais** | 1. Agendamento de auditoria<br>2. Verificação de prontidão<br>3. Comunicação com certificadora |
| **Resolução** | Renovação bem-sucedida da certificação |
| **Métricas** | Tempo de antecedência, desvios identificados, custo de remediação |

## Situações de Federação

### Situação: Certificado de Federação Expirado

| Atributo | Descrição |
|----------|-----------|
| **Identificador** | `FED_CERT_EXPIRED` |
| **Descrição** | Certificado usado na federação SAML expirou ou está inválido |
| **Gatilhos** | Erro de validação de certificado em tentativa de autenticação |
| **Ações Automáticas** | 1. Fallback para métodos alternativos de autenticação<br>2. Alerta para equipe de operações IAM<br>3. Tentativa de renovação automática se configurado |
| **Ações Manuais** | 1. Geração de novo certificado<br>2. Atualização nos metadados SAML<br>3. Comunicação com parceiros federados |
| **Resolução** | Implementação de certificado válido e restauração da federação |
| **Métricas** | Tempo de indisponibilidade, usuários impactados, prevenção futura |

### Situação: Desalinhamento de Atributos Federados

| Atributo | Descrição |
|----------|-----------|
| **Identificador** | `FED_ATTRIBUTE_MISMATCH` |
| **Descrição** | Provedor de identidade envia atributos em formato ou valores inesperados |
| **Gatilhos** | Erros de mapeamento de atributos após atualização do IdP |
| **Ações Automáticas** | 1. Uso de valores padrão quando possível<br>2. Registro detalhado da discrepância<br>3. Limitação de acesso a recursos críticos |
| **Ações Manuais** | 1. Ajuste dos mapeamentos de atributos<br>2. Contato com administradores do IdP<br>3. Testes de integração |
| **Resolução** | Harmonização dos atributos entre sistemas |
| **Métricas** | Impacto em acessos, tempo de detecção, eficácia da remediação |

## Matriz de Escalonamento

| Situação | Nível 1 (15 min) | Nível 2 (1 hora) | Nível 3 (4 horas) |
|----------|-----------------|------------------|-------------------|
| `AUTH_EXCESSIVE_ATTEMPTS` | Analista de Segurança | Gestor IAM | CISO |
| `AUTH_UNUSUAL_LOCATION` | Suporte IAM | Analista de Segurança | Gestor IAM |
| `AUTH_IDP_FAILURE` | Operações IAM | Arquiteto IAM | Gestor TI |
| `AUTH_PRIVILEGE_ESCALATION` | Analista de Segurança | CISO | CEO |
| `AUTH_SOD_CONFLICT` | Analista de Compliance | Gestor de Compliance | CFO |
| `IDM_ORPHANED_IDENTITIES` | Administrador IAM | Gestor IAM | Gestor de RH |
| `IDM_DORMANT_PRIVILEGED` | Administrador IAM | Gestor IAM | Gestor de Segurança |
| `MT_DATA_LEAKAGE` | Analista de Segurança | CISO | DPO |
| `MT_MIGRATION_FAILURE` | Administrador BD | Arquiteto de Dados | CTO |
| `COMP_REGULATORY_VIOLATION` | Analista de Compliance | DPO | Jurídico |
| `COMP_CERT_EXPIRATION` | Gestor de Compliance | CFO | CEO |
| `FED_CERT_EXPIRED` | Administrador IAM | Arquiteto IAM | Gestor TI |
| `FED_ATTRIBUTE_MISMATCH` | Administrador IAM | Arquiteto IAM | Gestor Integração |
