# INNOVABIZ AlertManager Dashboard - Documentação

## Visão Geral

Este documento descreve o dashboard Grafana para monitoramento do AlertManager na plataforma INNOVABIZ, projetado para fornecer visibilidade completa sobre o funcionamento do sistema de alertas em contexto multi-dimensional (tenant, região, ambiente).

O dashboard foi construído seguindo as melhores práticas de observabilidade, incorporando os princípios multi-contextuais da plataforma INNOVABIZ e garantindo conformidade com os requisitos de PCI DSS 4.0, GDPR/LGPD, ISO 27001 e NIST CSF.

## Estrutura do Dashboard

### 1. Status & Health
Esta seção fornece uma visão instantânea da saúde do AlertManager:
- **AlertManager Status**: Exibe se o serviço está operacional (UP/DOWN)
- **Active Alerts**: Total de alertas ativos sendo gerenciados
- **Active Silences**: Número de silenciamentos ativos
- **Alerts by Severity**: Distribuição de alertas por nível de severidade (critical, high, medium, low)
- **Alerts by Component**: Distribuição de alertas por componente (iam, payment_gateway, etc.)
- **Notification Rate by Integration**: Taxa de envio de notificações por canal de integração (email, slack, pagerduty, etc.)

### 2. Notification Performance & Errors
Esta seção monitora o desempenho do sistema de notificação:
- **Notification Failures by Integration**: Taxa de falhas de notificação por tipo de integração
- **Notification Failure Rate by Tenant**: Taxa percentual de falhas por tenant
- **Notification Latency by Integration**: Latência de envio de notificações (p95/p50) por integração
- **Alert Processing Rate by Tenant**: Throughput de processamento de alertas por tenant

### 3. Cluster Resources & Health
Esta seção monitora os recursos e a saúde do cluster AlertManager:
- **Memory Usage**: Uso de memória por instância
- **CPU Usage**: Utilização de CPU por instância
- **Goroutines**: Número de goroutines por instância (indicador de carga/vazamentos)
- **Open File Descriptors**: Descritores de arquivo abertos por instância
- **Configuration Reloads**: Histórico de recarregamentos de configuração (sucesso/falha)

## Funcionalidades Principais

### Filtragem Multi-dimensional
O dashboard implementa variáveis para filtragem contextual completa:
- **Tenant ID**: Filtragem por tenant específico
- **Region ID**: Filtragem por região geográfica (br, us, eu, ao)
- **Environment**: Filtragem por ambiente (production, staging, development)

### Annotations
O dashboard inclui anotações para eventos importantes:
- **Config Reloads**: Marca pontos de recarregamento de configuração
- **Notification Log Snapshots**: Indica quando snapshots do log de notificação foram criados

### Thresholds e Alertas Visuais
Os painéis implementam thresholds visuais para facilitar a identificação de problemas:
- Verde/Amarelo/Vermelho para diferentes níveis de métricas
- Indicadores coloridos para estados operacionais

## Implementação

### Pré-requisitos
- Prometheus v2.40+ configurado para coletar métricas do AlertManager
- Grafana v9.0+
- AlertManager v0.25+ com exportação de métricas habilitada

### Instruções de Importação
1. Acesse o Grafana através da URL: `https://grafana.<tenant_id>.<region_id>.innovabiz.io`
2. Navegue até Dashboards > Import
3. Cole o conteúdo do arquivo JSON ou faça upload do arquivo
4. Selecione a fonte de dados Prometheus apropriada
5. Clique em "Import"

### Configuração de Alertas
Este dashboard pode ser estendido com alertas do Grafana para monitoramento de segundo nível:

1. **AlertManager Down**
   - Condição: `up{job="alertmanager"} == 0`
   - Severidade: Critical
   - Notificar: Equipe de Plataforma

2. **High Notification Failure Rate**
   - Condição: `sum(rate(alertmanager_notifications_failed_total[5m])) by (integration) / sum(rate(alertmanager_notifications_total[5m])) by (integration) > 0.1`
   - Severidade: High
   - Notificar: Equipe de Plataforma

3. **High AlertManager Memory Usage**
   - Condição: `process_resident_memory_bytes{job="alertmanager"} > 2*10^9` (2GB)
   - Severidade: Medium
   - Notificar: Equipe de Plataforma

## Cenários de Uso

### Monitoramento Operacional
- **Monitoramento diário**: Verificação rápida do status do serviço e alertas ativos
- **Investigação de problemas de notificação**: Análise das taxas de falha e latência
- **Monitoramento de recursos**: Verificação do uso de recursos para planejamento de capacidade

### Auditoria de Segurança e Compliance
- **Rastreamento de alertas de segurança**: Visualização de alertas específicos de segurança
- **Verificação de distribuição por tenant**: Garantir isolamento adequado entre tenants
- **Conformidade com SLAs**: Monitorar métricas contra objetivos de nível de serviço

## Manutenção e Evolução

### Ciclo de Atualização
Este dashboard deve ser revisado trimestralmente como parte do processo de governança de observabilidade. Atualizações devem considerar:

1. Novos recursos do AlertManager
2. Feedback das equipes operacionais
3. Novos requisitos regulatórios
4. Expansão para novos tipos de integração

### Backups e Versionamento
O arquivo JSON deste dashboard deve ser armazenado e versionado no repositório Git junto com o código da plataforma. Cada alteração deve seguir o processo padrão de pull request e aprovação.

## Integração com Plataforma INNOVABIZ

Este dashboard é parte integrante da estratégia de observabilidade da plataforma INNOVABIZ e se integra com:

- **Runbooks**: Vinculado aos runbooks operacionais do AlertManager
- **Regras de Alerta**: Visualiza as regras de alerta padronizadas
- **ITSM**: Complementa o fluxo de ticket para incidentes
- **Documentação**: Referenciado na documentação central de observabilidade

## Suporte e Contatos

- **Proprietário**: Equipe de Plataforma INNOVABIZ
- **Contato Primário**: platform-team@innovabiz.com
- **Documentação Relacionada**: [Runbook Operacional AlertManager](../runbooks/ALERTMANAGER_OPERATIONAL_RUNBOOK.md)

---

**Data de Criação**: 26 de Julho de 2025  
**Versão**: 1.0  
**Autor**: INNOVABIZ Platform Team

*Este documento faz parte da documentação oficial de observabilidade da plataforma INNOVABIZ e está sujeito ao controle de versão e processo de governança documental.*