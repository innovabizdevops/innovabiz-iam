# Guia Operacional do Módulo IAM

## Introdução

Este guia operacional fornece informações detalhadas para administração, monitoramento, manutenção e resolução de problemas do módulo de Gerenciamento de Identidade e Acesso (IAM) da plataforma INNOVABIZ. Ele é destinado a administradores de sistema, equipes de operação e suporte técnico responsáveis pela operação contínua do sistema IAM.

## Visão Geral Operacional

O módulo IAM requer processos de operação bem definidos para garantir disponibilidade, segurança e desempenho contínuos. Este guia define as práticas operacionais recomendadas, procedimentos de manutenção de rotina e estratégias de resposta a incidentes.

### Papéis e Responsabilidades

| Papel | Responsabilidades |
|-------|------------------|
| **Administrador IAM** | Gerenciamento de configuração, políticas de segurança, auditorias e operações do dia-a-dia |
| **Operador de Sistema** | Monitoramento, manutenção de rotina, atualizações e backups |
| **Analista de Segurança** | Análise de logs, investigação de incidentes, verificações de segurança |
| **Administrador de Banco de Dados** | Otimização, manutenção e backup do banco de dados |
| **Suporte Nível 1** | Resolução de problemas básicos do usuário, escalonamento de questões complexas |
| **Suporte Nível 2** | Resolução de problemas técnicos, investigação de root cause |
| **Suporte Nível 3** | Resolução de problemas complexos, contato com equipe de desenvolvimento |

## Procedimentos Operacionais Padrão (POPs)

### 1. Monitoramento do Sistema

#### 1.1 Monitoramento de Disponibilidade

| Componente | Métrica | Limite | Frequência | Ação se Alerta |
|------------|---------|--------|------------|---------------|
| APIs | Uptime | < 99.9% | Contínuo | Verificar logs, reiniciar serviço se necessário |
| Banco de Dados | Uptime | < 99.99% | Contínuo | Verificar replicação, iniciar failover se necessário |
| Serviços OAuth | Tempo de Resposta | > 500ms | Contínuo | Verificar carga, escalar horizontalmente |
| Cache Redis | Hit Rate | < 80% | 5 min | Verificar evicções, ajustar políticas de cache |

#### 1.2 Monitoramento de Performance

| Métrica | Descrição | Limite | Frequência | Ação se Alerta |
|---------|-----------|--------|------------|---------------|
| CPU | Utilização de CPU | > 70% por 15 min | 1 min | Escalar horizontalmente, investigar processos |
| Memória | Uso de memória | > 80% por 10 min | 1 min | Verificar vazamentos, reiniciar ou escalar |
| Latência de API | Tempo médio resposta | > 200ms p95 | 1 min | Verificar gargalos, otimizar consultas |
| Conexões BD | Conexões ativas | > 80% do pool | 1 min | Aumentar pool, verificar fechamento de conexões |
| Fila Tarefas | Tamanho da fila | > 1000 itens | 1 min | Aumentar workers, verificar processamento |

#### 1.3 Dashboards de Monitoramento

Os seguintes dashboards devem ser mantidos e monitorados:

1. **Dashboard Operacional**: Visão geral em tempo real da saúde do sistema
2. **Dashboard de Segurança**: Alertas e eventos de segurança
3. **Dashboard de Performance**: Métricas detalhadas de desempenho
4. **Dashboard de Auditoria**: Atividades administrativas e eventos críticos
5. **Dashboard de Capacidade**: Tendências de uso de recursos e crescimento

### 2. Gerenciamento de Backup e Recuperação

#### 2.1 Estratégia de Backup

| Tipo | Escopo | Frequência | Retenção | Verificação |
|------|--------|------------|----------|------------|
| Completo | Banco de dados, configurações | Semanal | 6 meses | Restauração mensal para validação |
| Incremental | Banco de dados | Diário | 30 dias | Validação de checksums |
| Contínuo | WAL logs | Tempo real | 7 dias | Replay teste semanal |
| Configuração | Configurações, políticas, código | Após mudanças | 1 ano | Validação após backup |

#### 2.2 Procedimento de Restauração

Passos para restauração completa do sistema:

1. **Preparação**:
   - Identificar ponto de recuperação necessário
   - Preparar ambiente para restauração
   - Notificar stakeholders relevantes

2. **Restauração**:
   - Restaurar banco de dados do backup mais recente
   - Aplicar logs de transação até ponto desejado
   - Restaurar configurações e segredos
   - Verificar integridade dos dados restaurados

3. **Validação**:
   - Executar verificações de integridade
   - Testar funcionalidades críticas
   - Verificar conexões com sistemas externos
   - Confirmar políticas e configurações

4. **Ativação**:
   - Redirecionar tráfego para o sistema restaurado
   - Monitorar estabilidade por pelo menos 1 hora
   - Notificar usuários sobre retorno de operação

#### 2.3 Teste de Recuperação de Desastres

Testes de recuperação devem ser realizados regularmente:

- **Teste de Recuperação Parcial**: Mensal
- **Teste de Recuperação Completa**: Trimestral
- **Exercício de DR Total**: Semestral, incluindo failover de região

### 3. Gerenciamento de Usuários e Tenants

#### 3.1 Provisionamento de Tenant

Passos para adicionar um novo tenant:

1. Validar requisitos do tenant (tamanho, SLAs, requisitos de compliance)
2. Criar registro de tenant no sistema
3. Configurar particionamento e políticas de segurança
4. Definir papéis e permissões iniciais
5. Configurar administradores iniciais
6. Estabelecer limites e quotas de recursos
7. Verificar isolamento de dados
8. Executar validação de segurança
9. Ativar monitoramento específico do tenant

#### 3.2 Gerenciamento de Contas Privilegiadas

Procedimentos para contas administrativas:

1. **Criação**:
   - Verificação dupla de identidade pelo administrador e aprovador
   - Criação com privilégios mínimos necessários
   - Habilitação de MFA obrigatório

2. **Revisão**:
   - Auditoria mensal de todas as contas privilegiadas
   - Verificação de acessos e atividades
   - Validação da necessidade contínua de acesso

3. **Rotação**:
   - Rotação trimestral de senhas
   - Rotação imediata após saída de funcionário
   - Revogação em caso de inatividade (30 dias)

### 4. Gerenciamento de Patches e Atualizações

#### 4.1 Classificação de Patches

| Tipo | Descrição | SLA de Aplicação | Processo |
|------|-----------|-----------------|----------|
| Crítico | Vulnerabilidades críticas, zero-day | 24 horas | Emergencial, fora da janela se necessário |
| Segurança | Correções de segurança importantes | 7 dias | Janela regular, teste acelerado |
| Funcional | Correções de bugs, melhorias | 30 dias | Ciclo completo de teste |
| Menor | Melhorias menores, otimizações | Próxima janela | Ciclo normal de lançamento |

#### 4.2 Processo de Aplicação de Patches

1. **Pré-Implementação**:
   - Avaliar impacto do patch
   - Executar testes em ambiente de desenvolvimento/QA
   - Criar plano de rollback
   - Obter aprovações necessárias
   - Notificar stakeholders

2. **Implementação**:
   - Executar backup pré-patch
   - Aplicar em ambiente de homologação e validar
   - Aplicar em janela de manutenção aprovada
   - Verificar logs durante a aplicação
   - Monitorar performance pós-aplicação

3. **Pós-Implementação**:
   - Validar funcionamento do sistema
   - Confirmar resolução do problema original
   - Documentar mudanças e resultados
   - Atualizar documentação se necessário

#### 4.3 Janelas de Manutenção

| Ambiente | Janela Padrão | Duração | Frequência | Observações |
|----------|--------------|---------|------------|-------------|
| Produção | Domingo, 01:00-05:00 | 4 horas | Mensal | Com comunicação prévia de 7 dias |
| Homologação | Quarta, 22:00-02:00 | 4 horas | Quinzenal | Com comunicação prévia de 3 dias |
| Outros | Conforme necessário | Variável | Conforme necessário | Com comunicação prévia de 24 horas |

### 5. Procedimentos de Troubleshooting

#### 5.1 Problemas de Autenticação

| Sintoma | Possíveis Causas | Ações Iniciais | Escalonamento |
|---------|-----------------|---------------|--------------|
| Falhas em massa de login | Serviço de autenticação indisponível | Verificar status do serviço, logs de erros, reiniciar se necessário | Nível 2 se não resolver em 10 minutos |
| MFA não funciona | Problema com provedor MFA ou sincronização | Verificar conectividade com provedor MFA, logs de erro | Nível 2 se não resolver em 15 minutos |
| Expiração prematura de tokens | Configuração incorreta, sincronização NTP, chaves corrompidas | Verificar configuração, sincronização de tempo, estado das chaves | Nível 2 se não identificar causa |

#### 5.2 Problemas de Autorização

| Sintoma | Possíveis Causas | Ações Iniciais | Escalonamento |
|---------|-----------------|---------------|--------------|
| Acesso negado incorretamente | Política incorreta, cache desatualizado | Verificar políticas, limpar cache, verificar propagação de permissões | Nível 2 se não resolver em 15 minutos |
| Lentidão em decisões | Sobrecarga do motor de políticas, consultas lentas | Verificar métricas do motor de políticas, logs de performance | Nível 2 se persistir mais de 30 minutos |
| Vazamento de privilégios | Bug de segurança, configuração incorreta | Isolar problema, restringir acesso temporariamente | Nível 3 e equipe de segurança imediatamente |

## Monitoramento e Alertas

### 1. Configuração de Alertas

| Categoria | Alerta | Condição | Severidade | Notificação |
|-----------|--------|----------|------------|-------------|
| **Disponibilidade** | Serviço Inativo | Endpoint não responde por 2 min | Crítica | SMS, Email, Ticket |
| **Disponibilidade** | Degradação de Serviço | Resposta > 500ms por 5 min | Alta | Email, Ticket |
| **Segurança** | Tentativas de Login | >10 falhas para mesma conta em 5 min | Alta | Email, Dashboard |
| **Segurança** | Acesso Administrativo | Qualquer acesso a console admin | Média | Log, Dashboard |
| **Performance** | CPU Alta | >85% por 10 min | Alta | Email, Ticket |
| **Performance** | Memória Alta | >90% por 5 min | Alta | Email, Ticket |
| **Banco de Dados** | Replicação Atrasada | >30 seg de lag | Alta | Email, Ticket |
| **Banco de Dados** | Consultas Lentas | Consultas >1s | Média | Log, Dashboard |
| **Aplicação** | Erros de Aplicação | Taxa de erro >1% por 5 min | Alta | Email, Ticket |
| **Capacidade** | Disco Quase Cheio | >85% uso de disco | Alta | Email, Ticket |

### 2. Resposta a Alertas

| Severidade | Tempo Resposta | Tempo Resolução | Processo |
|------------|---------------|-----------------|----------|
| Crítica | 15 min | 2 horas | 1. Ack do alerta<br>2. Mitigação imediata<br>3. Comunicação a stakeholders<br>4. Resolução<br>5. RCA |
| Alta | 30 min | 8 horas | 1. Ack do alerta<br>2. Investigação<br>3. Resolução<br>4. Documentação |
| Média | 2 horas | 24 horas | 1. Ack do alerta<br>2. Investigação planejada<br>3. Resolução conforme prioridade |
| Baixa | 8 horas | Próximo ciclo | Resolução no próximo ciclo de manutenção |

## Gestão de Mudanças

### 1. Processo de Mudança

| Tipo de Mudança | Descrição | Aprovação Requerida | Janela | Notificação |
|-----------------|-----------|---------------------|--------|-------------|
| Emergencial | Correção crítica de segurança ou bug | CISO ou CTO | Imediata | Após a mudança |
| Significativa | Atualização de versão, nova funcionalidade | CAB, Product Owner | Janela padrão | 7 dias antes |
| Menor | Mudança de configuração de baixo impacto | Team Lead | Janela padrão | 3 dias antes |
| Rotineira | Ajustes pré-aprovados | Não requerida | A qualquer momento | Não requerida |

### 2. Documentação da Mudança

Cada mudança deve ser documentada com:

1. **Descrição**:
   - O que está sendo mudado
   - Por que a mudança é necessária
   - Impacto esperado

2. **Plano**:
   - Passos detalhados para implementação
   - Estimativa de tempo para cada passo
   - Pontos de verificação e critérios de sucesso

3. **Rollback**:
   - Plano detalhado de reversão
   - Pontos de decisão para ativação do rollback
   - Procedimento de validação pós-rollback

4. **Aprovações**:
   - Aprovadores requeridos e obtidos
   - Verificação de conformidade
   - Validação técnica

## Manutenção de Rotina

### 1. Tarefas Diárias

| Tarefa | Descrição | Responsável | Verificação |
|--------|-----------|-------------|------------|
| Verificação de Logs | Revisar logs de erros e alertas | Operador | Confirmar ausência de erros não tratados |
| Monitoramento de Performance | Revisar dashboards de performance | Operador | Verificar tendências anormais |
| Verificação de Backups | Confirmar sucesso de backups | Administrador BD | Verificar logs de backup |
| Verificação de Segurança | Revisar eventos de segurança | Analista de Segurança | Verificar tentativas suspeitas |

### 2. Tarefas Semanais

| Tarefa | Descrição | Responsável | Verificação |
|--------|-----------|-------------|------------|
| Análise de Tendências | Revisar métricas de longo prazo | Administrador IAM | Identificar padrões e anomalias |
| Teste de Restauração | Restaurar amostra de backup | Administrador BD | Verificar integridade dos dados |
| Revisão de Capacidade | Análise de uso de recursos | Operador | Planejar necessidades futuras |
| Limpeza de Dados | Purgar dados temporários | Administrador IAM | Verificar espaço recuperado |

### 3. Tarefas Mensais

| Tarefa | Descrição | Responsável | Verificação |
|--------|-----------|-------------|------------|
| Auditoria de Acessos | Revisar contas privilegiadas | Administrador IAM | Confirmar necessidade de acesso |
| Revisão de Configuração | Verificar configurações críticas | Administrador IAM | Comparar com baseline |
| Otimização de BD | Analisar e otimizar consultas | Administrador BD | Verificar melhoria de performance |
| Validação de Compliance | Verificar controles regulatórios | Analista de Compliance | Documentar resultados |

### 4. Tarefas Trimestrais

| Tarefa | Descrição | Responsável | Verificação |
|--------|-----------|-------------|------------|
| Teste DR | Exercício completo de recuperação | Equipe DR | Documentar RTO/RPO alcançados |
| Revisão de Arquitetura | Avaliar adequação da arquitetura | Arquiteto | Recomendar melhorias |
| Revisão de Segurança | Análise abrangente de segurança | Analista de Segurança | Documentar resultados |
| Planejamento de Capacidade | Projetar necessidades futuras | Administrador IAM | Atualizar plano de capacidade |

## Gestão de Incidentes

### 1. Classificação de Incidentes

| Severidade | Descrição | Exemplo | Tempo de Resposta |
|------------|-----------|---------|-------------------|
| SEV1 | Crítico - Sistema inoperante | Serviço de autenticação indisponível | 15 minutos |
| SEV2 | Alto - Funcionalidade principal degradada | Lentidão severa em autenticações | 30 minutos |
| SEV3 | Médio - Funcionalidade secundária afetada | Falha em relatórios ou funcionalidade não crítica | 2 horas |
| SEV4 | Baixo - Problema menor, workaround disponível | Erro em interface administrativa | 8 horas |

### 2. Processo de Resposta a Incidentes

1. **Detecção e Registro**:
   - Identificar e registrar o incidente
   - Classificar severidade inicial
   - Notificar equipe apropriada

2. **Resposta Inicial**:
   - Confirmar e refinar classificação
   - Iniciar investigação
   - Implementar contenção inicial
   - Notificar stakeholders conforme necessário

3. **Investigação e Diagnóstico**:
   - Determinar causa raiz
   - Avaliar escopo e impacto
   - Documentar achados

4. **Resolução e Recuperação**:
   - Implementar solução
   - Testar eficácia
   - Restaurar serviço normal
   - Verificar integridade dos dados

5. **Pós-Incidente**:
   - Conduzir análise pós-incidente
   - Documentar lições aprendidas
   - Implementar melhorias preventivas
   - Atualizar procedimentos se necessário

### 3. Comunicação de Incidentes

| Severidade | Quem Notificar | Frequência de Atualizações | Método |
|------------|---------------|---------------------------|--------|
| SEV1 | Todos stakeholders, Management | A cada 30 minutos | Email, SMS, Status Page |
| SEV2 | Stakeholders afetados, Management | A cada 2 horas | Email, Status Page |
| SEV3 | Stakeholders afetados | Uma vez ao dia | Email, Status Page |
| SEV4 | Apenas equipe interna | Ao resolver | Email, Ticket |

## Requisitos de Documentação Operacional

### 1. Documentação Requerida

| Documento | Descrição | Frequência de Atualização | Responsável |
|-----------|-----------|----------------------------|------------|
| Runbook | Procedimentos passo-a-passo para operações comuns | Após cada mudança significativa | Administrador IAM |
| Matriz RACI | Responsabilidades para cada processo operacional | Trimestral | Gerente de Operações |
| Catálogo de Serviços | Serviços fornecidos e SLAs | Semestral | Product Owner |
| Diagrama de Arquitetura | Representação visual dos componentes | Após mudanças | Arquiteto |
| Registro de Ativos | Inventário de todos os componentes | Mensal | Administrador IAM |
| Mapa de Dependências | Dependências entre serviços | Trimestral | Arquiteto |

### 2. Matriz RACI para Operações IAM

| Atividade | Administrador IAM | Operador | Suporte L1 | Suporte L2 | Segurança | DBA | Gerente |
|-----------|-------------------|----------|------------|------------|-----------|-----|---------|
| Monitoramento Diário | A | R | I | C | I | C | I |
| Gerenciamento de Acesso | R | C | I | C | A | I | I |
| Backups | A | R | - | C | I | R | I |
| Patching | A | R | I | C | A | C | I |
| Troubleshooting | A | C | R | R | C | C | I |
| Resposta a Incidentes | A | R | C | R | R | C | A |
| Mudanças | A | R | I | C | A | C | A |

*R - Responsável, A - Aprovador, C - Consultado, I - Informado*

## Ferramentas e Recursos Operacionais

### 1. Ferramentas de Monitoramento

| Ferramenta | Propósito | URL | Acesso |
|------------|-----------|-----|--------|
| Prometheus | Coleta de métricas | https://{tenant-id}.metrics.innovabiz.com | Operadores |
| Grafana | Visualização de métricas | https://{tenant-id}.dashboard.innovabiz.com | Operadores, Management |
| Loki | Agregação de logs | https://{tenant-id}.logs.innovabiz.com | Operadores, Suporte L2 |
| Jaeger | Tracing distribuído | https://{tenant-id}.tracing.innovabiz.com | Desenvolvedores, Suporte L2 |
| Alertmanager | Gestão de alertas | https://{tenant-id}.alerts.innovabiz.com | Operadores |

### 2. Requisitos de Log

| Componente | Nível de Log | Retenção | Objetivos |
|------------|--------------|----------|-----------|
| API REST | INFO em prod, DEBUG em outros | 90 dias | Troubleshooting, Auditoria |
| Serviço Auth | INFO em prod, DEBUG em outros | 1 ano | Segurança, Compliance |
| Banco de Dados | ERROR, WARN, INFO crítico | 1 ano | Performance, Segurança |
| Aplicação | INFO em prod, DEBUG em outros | 90 dias | Troubleshooting |
| Auditoria | Todos eventos | 7 anos | Compliance, Investigação |

## Referências

- [Matriz de Controles de Segurança](../05-Seguranca/Controles_Seguranca_IAM.md)
- [Requisitos de Infraestrutura](../04-Infraestrutura/Requisitos_Infraestrutura_IAM.md)
- [Arquitetura Técnica](../02-Arquitetura/Arquitetura_Tecnica_IAM.md)
- [Framework de Compliance](../10-Governanca/Framework_Compliance_IAM.md)
- [Procedimentos de Troubleshooting](Procedimentos_Troubleshooting_IAM.md)

## Apêndices

### A. Lista de Verificação Operacional Diária

- [ ] Verificar alertas das últimas 24 horas
- [ ] Revisar logs de erro para padrões incomuns
- [ ] Verificar métricas de performance vs. baseline
- [ ] Confirmar sucesso de backups noturnos
- [ ] Verificar status de replicação do banco de dados
- [ ] Revisar eventos de segurança suspeitos
- [ ] Verificar capacidade de armazenamento
- [ ] Confirmar que todos os serviços estão operacionais

### B. Procedimento de Escalonamento

1. **Nível 1 (0-30 minutos)**
   - Equipe de operações e suporte L1
   - Diagnóstico inicial e resolução de problemas básicos
   - Escalonar para L2 se não for resolvido em 30 minutos

2. **Nível 2 (30-60 minutos)**
   - Administrador IAM e equipe técnica especializada
   - Troubleshooting avançado
   - Escalonar para L3 se não for resolvido em 60 minutos

3. **Nível 3 (60+ minutos)**
   - Engenheiros de desenvolvimento
   - Gerente técnico e Product Owner
   - Consideração de soluções de emergência

4. **Nível 4 (Incidente Grave - 120+ minutos)**
   - CTO/CISO
   - Equipe de liderança
   - Comunicação executiva e gerenciamento de crise
